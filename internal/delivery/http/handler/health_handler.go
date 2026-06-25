package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type HealthHandler struct {
	db        *sqlx.DB
	startTime time.Time
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(h.startTime).String(),
	}

	if err := h.db.Ping(); err != nil {
		health["status"] = "unhealthy"
		health["database"] = "disconnected"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	health["database"] = "connected"
	c.JSON(http.StatusOK, health)
}
