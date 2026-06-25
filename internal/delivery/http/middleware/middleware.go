package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/logger"
	"go.uber.org/zap"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logger.Log.Error(
					"panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(stack)),
					zap.String("request id", c.GetString("request_id")),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{
						"error": "internal server error",
						"request_id": c.GetString("request_id"),
					},
				)
			}
		}()
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		allowedOrigins := []string{
			"http://localhost:8080",
			"http://127.0.0.1:8080",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}

		allowed := false
		for _, o := range allowedOrigins {
			if strings.Contains(origin, o) || origin == "*" {
				allowed = true
				break
			}
		}
		if origin != "" && !allowed {
			origin = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Refresh-Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, X-Refresh-Token")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func RateLimiterMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(100, time.Minute)
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !limiter.Allow(clientIP) {
			c.JSON(429, gin.H{
				"error":   "too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}
