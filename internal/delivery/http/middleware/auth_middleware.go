package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/dto"
	"github.com/yohagos/go-clean-user-api/internal/domain/usecase"
)

func AuthMiddleware(
	authUseCase usecase.AuthUseCase,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Code:    http.StatusUnauthorized,
				Details: "invalid authorization header format",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Code:    http.StatusUnauthorized,
				Details: "invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		userID, err := authUseCase.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Code:    http.StatusUnauthorized,
				Details: "invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", *userID)
		c.Next()
	}
}

func OptionalAuthMiddleware(authUseCase usecase.AuthUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				userID, err := authUseCase.ValidateToken(c.Request.Context(), parts[1])
				if err == nil {
					c.Set("user_id", userID)
				}
			}
		}
		c.Next()
	}
}
