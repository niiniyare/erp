package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// FeatureFlagWebSocketHandler handles WebSocket connections for real-time feature flag updates
type FeatureFlagWebSocketHandler struct {
	webSocketService featureflag.WebSocketService
}

// NewFeatureFlagWebSocketHandler creates a new WebSocket handler
func NewFeatureFlagWebSocketHandler(webSocketService featureflag.WebSocketService) *FeatureFlagWebSocketHandler {
	return &FeatureFlagWebSocketHandler{
		webSocketService: webSocketService,
	}
}

// HandleConnection handles WebSocket connection upgrade and management
func (h *FeatureFlagWebSocketHandler) HandleConnection(c *gin.Context) {
	log := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWebSocketHandler",
		"method":  "HandleConnection",
	})

	// Validate authentication
	if _, exists := c.Get("authorized_user_id"); !exists {
		log.Error("WebSocket connection attempt without authentication")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Validate tenant context
	if _, exists := c.Get("authorized_tenant_id"); !exists {
		log.Error("WebSocket connection attempt without tenant context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context required"})
		return
	}

	log.Info("Handling WebSocket connection for feature flag updates")

	// Delegate to the WebSocket service
	h.webSocketService.HandleConnection(c)
}

// GetConnectionStats returns current WebSocket connection statistics
func (h *FeatureFlagWebSocketHandler) GetConnectionStats(c *gin.Context) {
	log := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWebSocketHandler",
		"method":  "GetConnectionStats",
	})

	// Validate admin permissions for stats access
	if !h.isAdmin(c) {
		log.Error("Unauthorized attempt to access connection stats")
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	stats := h.webSocketService.GetConnectionStats()

	log.Info("Retrieved WebSocket connection stats", logger.Fields{
		"total_connections": stats.TotalConnections,
		"tenant_count":      len(stats.ConnectionsByTenant),
	})

	c.JSON(http.StatusOK, stats)
}

// SendTestNotification sends a test notification to verify WebSocket connectivity
func (h *FeatureFlagWebSocketHandler) SendTestNotification(c *gin.Context) {
	log := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWebSocketHandler",
		"method":  "SendTestNotification",
	})

	// Validate admin permissions
	if !h.isAdmin(c) {
		log.Error("Unauthorized attempt to send test notification")
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("authorized_tenant_id")
	if !exists {
		log.Error("Tenant context not found")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context required"})
		return
	}

	tenantUUID, ok := tenantID.(uuid.UUID)
	if !ok {
		log.Error("Invalid tenant ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Create test notification
	testEvent := &featureflag.FeatureFlagChangeEvent{
		FlagName:   "test-flag",
		ChangeType: "test",
		NewValue:   true,
		ChangedBy:  uuid.Nil,
		AppliedAt:  time.Now(),
		Metadata: map[string]interface{}{
			"test":    true,
			"message": "This is a test notification to verify WebSocket connectivity",
		},
	}

	err := h.webSocketService.NotifyFlagChange(c.Request.Context(), tenantUUID, testEvent)
	if err != nil {
		log.Error("Failed to send test notification", logger.Fields{"error": err})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	log.Info("Test notification sent successfully", logger.Fields{
		"tenant_id": tenantUUID,
		"flag_name": testEvent.FlagName,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":   "Test notification sent successfully",
		"tenant_id": tenantUUID,
		"timestamp": testEvent.AppliedAt,
	})
}

// BroadcastMessage broadcasts a custom message to all tenant connections (admin only)
func (h *FeatureFlagWebSocketHandler) BroadcastMessage(c *gin.Context) {
	log := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWebSocketHandler",
		"method":  "BroadcastMessage",
	})

	// Validate admin permissions
	if !h.isAdmin(c) {
		log.Error("Unauthorized attempt to broadcast message")
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Parse request body
	var req struct {
		Type    string      `json:"type" binding:"required"`
		Event   string      `json:"event" binding:"required"`
		Data    interface{} `json:"data"`
		Message string      `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Invalid request body", logger.Fields{"error": err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("authorized_tenant_id")
	if !exists {
		log.Error("Tenant context not found")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context required"})
		return
	}

	tenantUUID, ok := tenantID.(uuid.UUID)
	if !ok {
		log.Error("Invalid tenant ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Create broadcast message
	message := &featureflag.WebSocketMessage{
		Type:     req.Type,
		Event:    req.Event,
		TenantID: tenantUUID,
		Data: map[string]interface{}{
			"message": req.Message,
			"data":    req.Data,
			"from":    "admin",
		},
	}

	err := h.webSocketService.BroadcastToTenant(c.Request.Context(), tenantUUID, message)
	if err != nil {
		log.Error("Failed to broadcast message", logger.Fields{"error": err})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to broadcast message"})
		return
	}

	log.Info("Message broadcast successfully", logger.Fields{
		"tenant_id":    tenantUUID,
		"message_type": req.Type,
		"event":        req.Event,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":    "Message broadcast successfully",
		"tenant_id":  tenantUUID,
		"recipients": "all_tenant_connections",
	})
}

// CloseAllConnections closes all WebSocket connections (emergency use only)
func (h *FeatureFlagWebSocketHandler) CloseAllConnections(c *gin.Context) {
	log := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWebSocketHandler",
		"method":  "CloseAllConnections",
	})

	// Validate admin permissions
	if !h.isAdmin(c) {
		log.Error("Unauthorized attempt to close all connections")
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	log.Warn("Emergency closure of all WebSocket connections requested")

	h.webSocketService.CloseAllConnections()

	log.Info("All WebSocket connections closed")

	c.JSON(http.StatusOK, gin.H{
		"message": "All WebSocket connections closed",
		"status":  "success",
	})
}

// isAdmin checks if the current user has admin permissions
func (h *FeatureFlagWebSocketHandler) isAdmin(c *gin.Context) bool {
	// Check for admin role or permission
	if role, exists := c.Get("user_role"); exists {
		if roleStr, ok := role.(string); ok && roleStr == "admin" {
			return true
		}
	}

	// Check for specific admin permission
	if permissions, exists := c.Get("user_permissions"); exists {
		if permList, ok := permissions.([]string); ok {
			for _, perm := range permList {
				if perm == "admin" || perm == "websocket_admin" || perm == "feature_flag_admin" {
					return true
				}
			}
		}
	}

	return false
}
