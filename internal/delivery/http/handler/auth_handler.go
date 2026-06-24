package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/dto"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
	"github.com/yohagos/go-clean-user-api/internal/domain/usecase"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email, name and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body entity.RegisterRequest true "Registration details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req entity.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	user, tokens, err := h.authUseCase.Register(c.Request.Context(), &entity.RegisterRequest{
		Email: req.Email,
		Name: req.Name,
		Password: req.Password,
	})
	if err != nil {
		switch err {
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
	c.JSON(http.StatusCreated, gin.H{
		"user":   user,
		"tokens": tokens,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body entity.LoginRequest true "Login credentials"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req entity.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	tokens, err := h.authUseCase.Login(c.Request.Context(), &entity.LoginRequest{
		Email: req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err == entity.ErrUnauthorized {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "internal credentials",
				Code:    http.StatusUnauthorized,
				Details: "Invalid email or password",
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

	c.JSON(http.StatusOK, tokens)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body object true "Refresh token"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid refresh token",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	tokens, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case entity.ErrInvalidToken, entity.ErrExpiredToken, entity.ErrRevokedToken:
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "internal refresh token",
				Code:    http.StatusUnauthorized,
				Details: err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal server error",
				Code:    http.StatusInternalServerError,
				Details: err.Error(),
			})
		}
	}

	c.JSON(http.StatusOK, tokens)
}

// Logout godoc
// @Summary Logout user
// @Description Revoke access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Refresh-Token header string true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	refreshToken := c.GetHeader("X-Refresh-Token")

	if err := h.authUseCase.Logout(c.Request.Context(), accessToken, refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "logout failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}
