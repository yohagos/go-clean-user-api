package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/dto"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
	"github.com/yohagos/go-clean-user-api/internal/domain/usecase"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
}

func NewUserHandler(userUseCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a user (admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateUserRequest true "User creation request"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	user, err := h.userUseCase.CreateUser(c.Request.Context(), req.Email, req.Name)
	if err != nil {
		switch err {
		case entity.ErrInvalidEmail, entity.ErrInvalidName:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   err.Error(),
				Code:    http.StatusBadRequest,
				Details: "Please provide valid email and name",
			})
		case entity.ErrEmailExists:
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   err.Error(),
				Code:    http.StatusConflict,
				Details: "User with this email already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal server error",
				Code:    http.StatusInternalServerError,
				Details: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.ToUserResponse(user))
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a single user by their UUID
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid user id",
			Code:    http.StatusBadRequest,
			Details: "User ID must be a valid UUID",
		})
		return
	}

	user, err := h.userUseCase.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrUserNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   err.Error(),
				Code:    http.StatusNotFound,
				Details: "User not found with given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal server error",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Get paginated list of users
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of users to return (default: 20, max: 100)"
// @Param offset query int false "Number of users to skip (default: 0)"
// @Success 200 {object} dto.UsersListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "2ß"))
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		offset = 0
	}

	users, err := h.userUseCase.GetAllUsers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal server error",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	response := dto.UsersListResponse{
		Users:      dto.ToUserResponseList(users),
		TotalCount: len(users),
		Limit:      limit,
		Offset:     offset,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user email and/or name
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "User update request"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid user id",
			Code:    http.StatusBadRequest,
			Details: "User ID must be a valid UUID",
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	user, err := h.userUseCase.UpdateUser(c.Request.Context(), id, req.Email, req.Name)
	if err != nil {
		switch err {
		case entity.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   err.Error(),
				Code:    http.StatusNotFound,
				Details: "User not found with given ID",
			})
		case entity.ErrEmailExists:
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   err.Error(),
				Code:    http.StatusConflict,
				Details: "Another user already has this email",
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal server error",
				Code:    http.StatusInternalServerError,
				Details: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid user id",
			Code:    http.StatusBadRequest,
			Details: "User ID must be a valid UUID",
		})
		return
	}

	if err := h.userUseCase.DeleteUser(c.Request.Context(), id); err != nil {
		if err == entity.ErrUserNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   err.Error(),
				Code:    http.StatusNotFound,
				Details: "User not found with the given id",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal server error",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
