package featureflag

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	db "awo.so/db/sqlc"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// WebSocketService handles real-time feature flag updates
type WebSocketService interface {
	// Connection management
	HandleConnection(w http.ResponseWriter, r *http.Request)
	BroadcastToTenant(ctx context.Context, tenantID uuid.UUID, message *WebSocketMessage) error
	BroadcastToAllTenants(ctx context.Context, message *WebSocketMessage) error

	// Feature flag specific notifications
	NotifyFlagChange(ctx context.Context, tenantID uuid.UUID, event *FeatureFlagChangeEvent) error
	NotifyFlagCreated(ctx context.Context, tenantID uuid.UUID, flag *FeatureFlag) error
	NotifyFlagDeleted(ctx context.Context, tenantID uuid.UUID, flagName string) error
	NotifyApprovalRequired(ctx context.Context, tenantID uuid.UUID, event *ApprovalRequiredEvent) error

	// Health and metrics
	GetConnectionStats() *ConnectionStats
	CloseAllConnections()
}

// WebSocketConnection represents a single WebSocket connection
type WebSocketConnection struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	UserID   uuid.UUID
	Conn     *websocket.Conn
	Send     chan *WebSocketMessage
	LastSeen time.Time
	mu       sync.RWMutex
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string    `json:"type"`
	Event     string    `json:"event"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	MessageID uuid.UUID `json:"message_id"`
}

// FeatureFlagChangeEvent represents a feature flag change for real-time updates
type FeatureFlagChangeEvent struct {
	FlagID          uuid.UUID      `json:"flag_id"`
	FlagName        string         `json:"flag_name"`
	ChangeType      string         `json:"change_type"` // enable, disable, update_rollout, created, deleted
	OldValue        any            `json:"old_value,omitempty"`
	NewValue        any            `json:"new_value"`
	ChangedBy       uuid.UUID      `json:"changed_by"`
	AccessRequestID *uuid.UUID     `json:"access_request_id,omitempty"`
	AppliedAt       time.Time      `json:"applied_at"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// ApprovalRequiredEvent represents an approval request event
type ApprovalRequiredEvent struct {
	AccessRequestID uuid.UUID      `json:"access_request_id"`
	FlagName        string         `json:"flag_name"`
	ChangeType      string         `json:"change_type"`
	RequestedBy     uuid.UUID      `json:"requested_by"`
	Justification   string         `json:"justification"`
	BusinessReason  string         `json:"business_reason,omitempty"`
	RequestedAt     time.Time      `json:"requested_at"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// ConnectionStats represents WebSocket connection statistics
type ConnectionStats struct {
	TotalConnections    int               `json:"total_connections"`
	ConnectionsByTenant map[uuid.UUID]int `json:"connections_by_tenant"`
	ConnectionsByUser   map[uuid.UUID]int `json:"connections_by_user"`
	AverageLatency      time.Duration     `json:"average_latency"`
	MessagesPerSecond   float64           `json:"messages_per_second"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

// webSocketService implements WebSocketService
type webSocketService struct {
	store       db.Store
	tracing     tracing.Service
	metrics     metrics.MetricsProvider
	connections map[uuid.UUID]*WebSocketConnection
	tenantConns map[uuid.UUID][]*WebSocketConnection
	mu          sync.RWMutex
	upgrader    websocket.Upgrader

	// Message tracking for metrics
	messageCount int64
	lastSecond   time.Time
	msgCountMu   sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(
	store db.Store,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) WebSocketService {
	return &webSocketService{
		store:       store,
		tracing:     tracing,
		metrics:     metrics,
		connections: make(map[uuid.UUID]*WebSocketConnection),
		tenantConns: make(map[uuid.UUID][]*WebSocketConnection),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking based on tenant configuration
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		lastSecond: time.Now(),
	}
}

// HandleConnection handles new WebSocket connections
func (s *webSocketService) HandleConnection(w http.ResponseWriter, r *http.Request) {
	ctxWithTrace, span := s.tracing.StartSpan(r.Context(), "webSocketService.HandleConnection")
	defer span.End()

	log := logger.WithFields(logger.Fields{"service": "websocket", "method": "HandleConnection"})

	// Extract tenant and user from context/headers
	tenantID, err := extractTenantIDFromRequest(r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to extract tenant ID")
		log.Error("Failed to extract tenant ID", logger.Fields{"error": err})
		http.Error(w, `{"error": "Tenant ID required"}`, http.StatusBadRequest)
		return
	}

	userID, err := extractUserIDFromRequest(r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to extract user ID")
		log.Error("Failed to extract user ID", logger.Fields{"error": err})
		http.Error(w, `{"error": "User ID required"}`, http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("user.id", userID.String()),
	)

	// Upgrade connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to upgrade connection")
		log.Error("Failed to upgrade connection", logger.Fields{"error": err})
		return
	}

	// Create WebSocket connection
	wsConn := &WebSocketConnection{
		ID:       uuid.New(),
		TenantID: tenantID,
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan *WebSocketMessage, 256),
		LastSeen: time.Now(),
	}

	// Register connection
	s.registerConnection(wsConn)

	log.Info("WebSocket connection established", logger.Fields{
		"connection_id": wsConn.ID,
		"tenant_id":     tenantID,
		"user_id":       userID,
	})

	// Send welcome message
	welcomeMessage := &WebSocketMessage{
		Type:     "system",
		Event:    "connected",
		TenantID: tenantID,
		Data: map[string]any{
			"connection_id": wsConn.ID,
			"message":       "Connected to feature flag real-time updates",
		},
		Timestamp: time.Now(),
		MessageID: uuid.New(),
	}
	wsConn.Send <- welcomeMessage

	// Start connection handlers
	go s.handleConnectionRead(ctxWithTrace, wsConn)
	go s.handleConnectionWrite(ctxWithTrace, wsConn)

	s.metrics.IncrementCounter("websocket_connections_established", map[string]any{
		"tenant_id": tenantID.String(),
	})
}

// registerConnection registers a new WebSocket connection
func (s *webSocketService) registerConnection(conn *WebSocketConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connections[conn.ID] = conn

	if _, exists := s.tenantConns[conn.TenantID]; !exists {
		s.tenantConns[conn.TenantID] = make([]*WebSocketConnection, 0)
	}
	s.tenantConns[conn.TenantID] = append(s.tenantConns[conn.TenantID], conn)
}

// unregisterConnection removes a WebSocket connection
func (s *webSocketService) unregisterConnection(conn *WebSocketConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.connections, conn.ID)

	// Remove from tenant connections
	if tenantConns, exists := s.tenantConns[conn.TenantID]; exists {
		for i, c := range tenantConns {
			if c.ID == conn.ID {
				s.tenantConns[conn.TenantID] = append(tenantConns[:i], tenantConns[i+1:]...)
				break
			}
		}

		// Clean up empty tenant connection lists
		if len(s.tenantConns[conn.TenantID]) == 0 {
			delete(s.tenantConns, conn.TenantID)
		}
	}

	close(conn.Send)
	conn.Conn.Close()

	s.metrics.IncrementCounter("websocket_connections_closed", map[string]any{
		"tenant_id": conn.TenantID.String(),
	})
}

// handleConnectionRead handles incoming messages from WebSocket connection
func (s *webSocketService) handleConnectionRead(ctx context.Context, conn *WebSocketConnection) {
	defer s.unregisterConnection(conn)
	log := logger.WithFields(logger.Fields{"service": "websocket", "method": "handleConnectionRead"})

	conn.Conn.SetReadLimit(512)
	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.Conn.SetPongHandler(func(string) error {
		conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.mu.Lock()
		conn.LastSeen = time.Now()
		conn.mu.Unlock()
		return nil
	})

	for {
		_, message, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("WebSocket read error", logger.Fields{
					"connection_id": conn.ID,
					"error":         err,
				})
			}
			break
		}

		// Handle incoming message (could be heartbeat, subscription, etc.)
		s.handleIncomingMessage(ctx, conn, message)
	}
}

// handleConnectionWrite handles outgoing messages to WebSocket connection
func (s *webSocketService) handleConnectionWrite(ctx context.Context, conn *WebSocketConnection) {
	ticker := time.NewTicker(54 * time.Second) // Send ping every 54 seconds
	defer func() {
		ticker.Stop()
		conn.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-conn.Send:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Send the message
			jsonData, _ := json.Marshal(message)
			w.Write(jsonData)

			// Add queued messages to the same write
			n := len(conn.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				nextMessage := <-conn.Send
				nextJSON, _ := json.Marshal(nextMessage)
				w.Write(nextJSON)
			}

			if err := w.Close(); err != nil {
				return
			}

			s.trackMessage()

		case <-ticker.C:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleIncomingMessage processes messages from clients
func (s *webSocketService) handleIncomingMessage(ctx context.Context, conn *WebSocketConnection, message []byte) {
	log := logger.WithFields(logger.Fields{"service": "websocket", "method": "handleIncomingMessage"})
	var msg map[string]any
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Error("Invalid WebSocket message", logger.Fields{
			"connection_id": conn.ID,
			"error":         err,
		})
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "heartbeat":
		// Update last seen time
		conn.mu.Lock()
		conn.LastSeen = time.Now()
		conn.mu.Unlock()

		// Send heartbeat response
		response := &WebSocketMessage{
			Type:      "system",
			Event:     "heartbeat",
			TenantID:  conn.TenantID,
			Data:      map[string]any{"status": "ok"},
			Timestamp: time.Now(),
			MessageID: uuid.New(),
		}
		select {
		case conn.Send <- response:
		default:
			// Connection is blocked, close it
			s.unregisterConnection(conn)
		}

	case "subscribe":
		// Handle subscription to specific flag updates
		if flagName, ok := msg["flag_name"].(string); ok {
			log.Info("Client subscribed to flag", logger.Fields{
				"connection_id": conn.ID,
				"flag_name":     flagName,
			})
		}

	default:
		log.Warn("Unknown message type", logger.Fields{
			"connection_id": conn.ID,
			"type":          msgType,
		})
	}
}

// BroadcastToTenant sends a message to all connections for a specific tenant
func (s *webSocketService) BroadcastToTenant(ctx context.Context, tenantID uuid.UUID, message *WebSocketMessage) error {
	ctx, span := s.tracing.StartSpan(ctx, "webSocketService.BroadcastToTenant")
	defer span.End()
	log := logger.WithFields(logger.Fields{"service": "websocket", "method": "BroadcastToTenant"})

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	s.mu.RLock()
	connections := make([]*WebSocketConnection, len(s.tenantConns[tenantID]))
	copy(connections, s.tenantConns[tenantID])
	s.mu.RUnlock()

	if len(connections) == 0 {
		log.Info("No WebSocket connections found for tenant", logger.Fields{"tenant_id": tenantID})
		return nil
	}

	// Add message ID and timestamp if not set
	if message.MessageID == uuid.Nil {
		message.MessageID = uuid.New()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Send to all connections
	sent := 0
	for _, conn := range connections {
		select {
		case conn.Send <- message:
			sent++
		default:
			// Connection is blocked, remove it
			log.Warn("Removing blocked WebSocket connection", logger.Fields{
				"connection_id": conn.ID,
				"tenant_id":     tenantID,
			})
			go s.unregisterConnection(conn)
		}
	}

	s.metrics.IncrementCounter("websocket_messages_broadcast", map[string]any{
		"tenant_id":     tenantID.String(),
		"message_type":  message.Type,
		"message_event": message.Event,
		"connections":   sent,
	})

	log.Info("Message broadcast to tenant", logger.Fields{
		"tenant_id":     tenantID,
		"message_type":  message.Type,
		"message_event": message.Event,
		"sent_to":       sent,
		"total_conns":   len(connections),
	})

	return nil
}

// BroadcastToAllTenants sends a message to all connections
func (s *webSocketService) BroadcastToAllTenants(ctx context.Context, message *WebSocketMessage) error {
	ctx, span := s.tracing.StartSpan(ctx, "webSocketService.BroadcastToAllTenants")
	defer span.End()

	s.mu.RLock()
	allConnections := make([]*WebSocketConnection, 0, len(s.connections))
	for _, conn := range s.connections {
		allConnections = append(allConnections, conn)
	}
	s.mu.RUnlock()

	if len(allConnections) == 0 {
		return nil
	}

	// Add message ID and timestamp if not set
	if message.MessageID == uuid.Nil {
		message.MessageID = uuid.New()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Send to all connections
	sent := 0
	for _, conn := range allConnections {
		// Update message tenant ID for each connection
		message.TenantID = conn.TenantID

		select {
		case conn.Send <- message:
			sent++
		default:
			go s.unregisterConnection(conn)
		}
	}

	s.metrics.IncrementCounter("websocket_messages_broadcast_all", map[string]any{
		"message_type":  message.Type,
		"message_event": message.Event,
		"connections":   sent,
	})

	return nil
}

// NotifyFlagChange sends a feature flag change notification
func (s *webSocketService) NotifyFlagChange(ctx context.Context, tenantID uuid.UUID, event *FeatureFlagChangeEvent) error {
	message := &WebSocketMessage{
		Type:      "feature_flag",
		Event:     "flag_changed",
		TenantID:  tenantID,
		Data:      event,
		Timestamp: time.Now(),
		MessageID: uuid.New(),
	}

	return s.BroadcastToTenant(ctx, tenantID, message)
}

// NotifyFlagCreated sends a feature flag creation notification
func (s *webSocketService) NotifyFlagCreated(ctx context.Context, tenantID uuid.UUID, flag *FeatureFlag) error {
	event := &FeatureFlagChangeEvent{
		FlagID:     flag.ID,
		FlagName:   flag.Name,
		ChangeType: "created",
		NewValue:   flag.DefaultValue,
		ChangedBy:  uuid.Nil, // Could be set by the caller
		AppliedAt:  time.Now(),
	}

	return s.NotifyFlagChange(ctx, tenantID, event)
}

// NotifyFlagDeleted sends a feature flag deletion notification
func (s *webSocketService) NotifyFlagDeleted(ctx context.Context, tenantID uuid.UUID, flagName string) error {
	event := &FeatureFlagChangeEvent{
		FlagName:   flagName,
		ChangeType: "deleted",
		ChangedBy:  uuid.Nil,
		AppliedAt:  time.Now(),
	}

	return s.NotifyFlagChange(ctx, tenantID, event)
}

// NotifyApprovalRequired sends an approval required notification
func (s *webSocketService) NotifyApprovalRequired(ctx context.Context, tenantID uuid.UUID, event *ApprovalRequiredEvent) error {
	message := &WebSocketMessage{
		Type:      "access_request",
		Event:     "approval_required",
		TenantID:  tenantID,
		Data:      event,
		Timestamp: time.Now(),
		MessageID: uuid.New(),
	}

	return s.BroadcastToTenant(ctx, tenantID, message)
}

// GetConnectionStats returns current connection statistics
func (s *webSocketService) GetConnectionStats() *ConnectionStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &ConnectionStats{
		TotalConnections:    len(s.connections),
		ConnectionsByTenant: make(map[uuid.UUID]int),
		ConnectionsByUser:   make(map[uuid.UUID]int),
		UpdatedAt:           time.Now(),
	}

	for _, conn := range s.connections {
		stats.ConnectionsByTenant[conn.TenantID]++
		stats.ConnectionsByUser[conn.UserID]++
	}

	// Calculate messages per second
	s.msgCountMu.RLock()
	if time.Since(s.lastSecond) >= time.Second {
		stats.MessagesPerSecond = float64(s.messageCount)
		s.messageCount = 0
		s.lastSecond = time.Now()
	}
	s.msgCountMu.RUnlock()

	return stats
}

// CloseAllConnections closes all WebSocket connections
func (s *webSocketService) CloseAllConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, conn := range s.connections {
		conn.Conn.Close()
		close(conn.Send)
	}

	s.connections = make(map[uuid.UUID]*WebSocketConnection)
	s.tenantConns = make(map[uuid.UUID][]*WebSocketConnection)
}

// trackMessage increments message counter for metrics
func (s *webSocketService) trackMessage() {
	s.msgCountMu.Lock()
	s.messageCount++
	s.msgCountMu.Unlock()
}

// Helper functions for extracting context information
func extractTenantIDFromRequest(r *http.Request) (uuid.UUID, error) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		// Try to get from query parameter
		tenantIDStr = r.URL.Query().Get("tenant_id")
	}

	if tenantIDStr == "" {
		return uuid.Nil, fmt.Errorf("tenant ID not found in headers or query parameters")
	}

	return uuid.Parse(tenantIDStr)
}

func extractUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		// Try to get from query parameter
		userIDStr = r.URL.Query().Get("user_id")
	}

	if userIDStr == "" {
		return uuid.Nil, fmt.Errorf("user ID not found in headers or query parameters")
	}

	return uuid.Parse(userIDStr)
}
