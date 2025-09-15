package handlers

// DEPRECATED: This file contains legacy Gin-based WebSocket handlers.
// The system now uses Goa-based feature flag handlers with native HTTP.
// WebSocket functionality should be integrated through standard HTTP handlers.

/*
import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// FeatureFlagWebSocketHandler handles WebSocket connections for real-time feature flag updates (DEPRECATED)
type FeatureFlagWebSocketHandler struct {
	webSocketService featureflag.WebSocketService
}

// NewFeatureFlagWebSocketHandler creates a new WebSocket handler (DEPRECATED)
func NewFeatureFlagWebSocketHandler(webSocketService featureflag.WebSocketService) *FeatureFlagWebSocketHandler {
	return &FeatureFlagWebSocketHandler{
		webSocketService: webSocketService,
	}
}

// All WebSocket handler methods commented out - use native HTTP handlers instead
// HandleConnection, GetConnectionStats, SendTestNotification, BroadcastMessage, etc.
// should be implemented as standard HTTP handlers integrated with Goa server
*/
