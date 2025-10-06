package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health handles basic health check requests
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"service": "erp",
		"version": "1.0.0",
	})
}

// Ready handles readiness check requests
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// TODO: Add actual readiness checks (database, cache, etc.)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ready",
		"checks": map[string]any{
			"database": "ok",
			"cache":    "ok",
		},
	})
}
