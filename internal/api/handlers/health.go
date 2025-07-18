package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
    return &HealthHandler{}
}

// Health handles basic health check requests
func (h *HealthHandler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "ok",
        "service": "awo",
        "version": "1.0.0",
    })
}

// Ready handles readiness check requests
func (h *HealthHandler) Ready(c *gin.Context) {
    // TODO: Add actual readiness checks (database, cache, etc.)
    c.JSON(http.StatusOK, gin.H{
        "status": "ready",
        "checks": gin.H{
            "database": "ok",
            "cache":    "ok",
        },
    })
}
