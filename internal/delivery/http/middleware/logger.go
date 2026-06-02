package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yohagos/go-clean-user-api/internal/logger"
	"go.uber.org/zap"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Log.Info(
			"HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("response_id", c.GetString("request_id")),
			zap.Int("body_size", c.Writer.Size()),
		)
	}
}
