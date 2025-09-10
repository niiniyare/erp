package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"goa.design/goa/v3/pkg"

	"github.com/niiniyare/erp/internal/api/gen/organization"
	"github.com/niiniyare/erp/internal/core/entity"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// EntityHandler handles entity-related HTTP requests
type EntityHandler struct {
	service entity.Service
	tracing tracing.TracingService
	metrics *metrics.MetricsService
}

// NewEntityHandler creates a new entity handler
func NewEntityHandler(service entity.Service, tracing tracing.TracingService, metrics *metrics.MetricsService) *EntityHandler {
	return &EntityHandler{
		service: service,
		tracing: tracing,
		metrics: metrics,
	}
}

// CreateEntity handles entity creation requests
func (h *EntityHandler) CreateEntity(c *gin.Context) {
	// Extract tracing context
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	// Start HTTP span
	ctx, span := h.tracing.StartSpan(ctx, "http.create_entity",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities",
	})
	defer timer.Stop()

	// Log request
	logger.InfoContext(ctx, "Processing create entity request",
		logger.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": c.Request.UserAgent(),
			"client_ip":  c.ClientIP(),
		})

	var req entity.CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/entities",
			"error_type": "validation_error",
		})

		// Log error
		logger.WarnContext(ctx, "Invalid request payload",
			logger.Fields{
				"error":        err.Error(),
				"content_type": c.GetHeader("Content-Type"),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add request attributes to span
	span.SetAttributes(
		attribute.String("entity.name", req.Name),
		attribute.String("entity.code", req.Code),
		attribute.String("entity.type", string(req.Type)),
	)

	// Create entity
	newEntity, err := h.service.CreateEntity(ctx, req)
	if err != nil {
		// Record error in span
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		// Increment error counter
		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/entities",
			"error_type": "business_error",
		})

		// Handle specific business errors
		switch {
		case errors.Is(err, sharedErrors.ErrEntityNameExists):
			logger.WarnContext(ctx, "Entity creation failed - name exists",
				logger.Fields{
					"entity_name": req.Name,
					"entity_type": string(req.Type),
				})
			c.JSON(http.StatusConflict, gin.H{"error": "Entity name already exists"})
		case errors.Is(err, sharedErrors.ErrEntityCodeExists):
			logger.WarnContext(ctx, "Entity creation failed - code exists",
				logger.Fields{
					"entity_code": req.Code,
					"entity_type": string(req.Type),
				})
			c.JSON(http.StatusConflict, gin.H{"error": "Entity code already exists"})
		case errors.Is(err, sharedErrors.ErrInvalidParentEntity):
			logger.WarnContext(ctx, "Entity creation failed - invalid parent",
				logger.Fields{
					"entity_name": req.Name,
					"parent_id":   req.ParentID,
				})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent entity"})
		case errors.Is(err, sharedErrors.ErrCircularReference):
			logger.WarnContext(ctx, "Entity creation failed - circular reference",
				logger.Fields{
					"entity_name": req.Name,
					"parent_id":   req.ParentID,
				})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Circular reference detected"})
		case errors.Is(err, sharedErrors.ErrInvalidEntityType):
			logger.WarnContext(ctx, "Entity creation failed - invalid type",
				logger.Fields{
					"entity_name": req.Name,
					"entity_type": string(req.Type),
				})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity type"})
		default:
			logger.ErrorContext(ctx, "Entity creation failed",
				logger.Fields{
					"error":       err.Error(),
					"entity_name": req.Name,
					"entity_type": string(req.Type),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Success metrics
	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities",
		"status":   "success",
	})

	// Log success
	logger.InfoContext(ctx, "Entity created successfully",
		logger.Fields{
			"entity_id":   newEntity.ID.String(),
			"entity_name": newEntity.Name,
			"entity_type": string(newEntity.Type),
			"status_code": http.StatusCreated,
		})

	c.JSON(http.StatusCreated, newEntity)
}

// GetEntity handles entity retrieval requests
func (h *EntityHandler) GetEntity(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.get_entity",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		))
	defer span.End()

	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities/:identifier",
	})
	defer timer.Stop()

	identifier := c.Param("identifier")
	span.SetAttributes(attribute.String("entity.identifier", identifier))

	logger.InfoContext(ctx, "Processing get entity request",
		logger.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"identifier": identifier,
		})

	entity, err := h.service.GetEntity(ctx, identifier)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/entities/:identifier",
			"error_type": "business_error",
		})

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found",
				logger.Fields{"identifier": identifier})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		default:
			logger.ErrorContext(ctx, "Failed to get entity",
				logger.Fields{
					"identifier": identifier,
					"error":      err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities/:identifier",
		"status":   "success",
	})

	logger.InfoContext(ctx, "Entity retrieved successfully",
		logger.Fields{
			"entity_id":   entity.ID.String(),
			"entity_name": entity.Name,
			"identifier":  identifier,
		})

	c.JSON(http.StatusOK, entity)
}

// GetEntityByID handles entity retrieval by ID
func (h *EntityHandler) GetEntityByID(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.get_entity_by_id",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		logger.WarnContext(ctx, "Invalid entity ID",
			logger.Fields{
				"id_param": idStr,
				"error":    err.Error(),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	span.SetAttributes(attribute.String("entity.id", id.String()))

	entity, err := h.service.GetEntityByID(ctx, id)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found",
				logger.Fields{"entity_id": id.String()})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		default:
			logger.ErrorContext(ctx, "Failed to get entity",
				logger.Fields{
					"entity_id": id.String(),
					"error":     err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, entity)
}

// UpdateEntity handles entity update requests
func (h *EntityHandler) UpdateEntity(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.update_entity",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities/:id",
	})
	defer timer.Stop()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		logger.WarnContext(ctx, "Invalid entity ID",
			logger.Fields{
				"id_param": idStr,
				"error":    err.Error(),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	span.SetAttributes(attribute.String("entity.id", id.String()))

	var req entity.UpdateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/entities/:id",
			"error_type": "validation_error",
		})

		logger.WarnContext(ctx, "Invalid request payload",
			logger.Fields{
				"error":        err.Error(),
				"content_type": c.GetHeader("Content-Type"),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.InfoContext(ctx, "Processing update entity request",
		logger.Fields{
			"entity_id": id.String(),
			"method":    c.Request.Method,
		})

	updatedEntity, err := h.service.UpdateEntity(ctx, id, req)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/entities/:id",
			"error_type": "business_error",
		})

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found for update",
				logger.Fields{"entity_id": id.String()})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		case errors.Is(err, sharedErrors.ErrEntityNameExists):
			logger.WarnContext(ctx, "Entity update failed - name exists",
				logger.Fields{
					"entity_id": id.String(),
					"name":      req.Name,
				})
			c.JSON(http.StatusConflict, gin.H{"error": "Entity name already exists"})
		case errors.Is(err, sharedErrors.ErrEntityCodeExists):
			logger.WarnContext(ctx, "Entity update failed - code exists",
				logger.Fields{
					"entity_id": id.String(),
					"code":      req.Code,
				})
			c.JSON(http.StatusConflict, gin.H{"error": "Entity code already exists"})
		case errors.Is(err, sharedErrors.ErrInvalidParentEntity):
			logger.WarnContext(ctx, "Entity update failed - invalid parent",
				logger.Fields{
					"entity_id": id.String(),
					"parent_id": req.ParentID,
				})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent entity"})
		case errors.Is(err, sharedErrors.ErrCircularReference):
			logger.WarnContext(ctx, "Entity update failed - circular reference",
				logger.Fields{
					"entity_id": id.String(),
					"parent_id": req.ParentID,
				})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Circular reference detected"})
		case errors.Is(err, sharedErrors.ErrInvalidEntityType):
			logger.WarnContext(ctx, "Entity update failed - invalid type",
				logger.Fields{
					"entity_id": id.String(),
					"type":      req.Type,
				})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity type"})
		default:
			logger.ErrorContext(ctx, "Entity update failed",
				logger.Fields{
					"entity_id": id.String(),
					"error":     err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities/:id",
		"status":   "success",
	})

	logger.InfoContext(ctx, "Entity updated successfully",
		logger.Fields{
			"entity_id":   updatedEntity.ID.String(),
			"entity_name": updatedEntity.Name,
		})

	c.JSON(http.StatusOK, updatedEntity)
}

// DeleteEntity handles entity deletion requests
func (h *EntityHandler) DeleteEntity(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.delete_entity",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities/:id",
	})
	defer timer.Stop()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		logger.WarnContext(ctx, "Invalid entity ID",
			logger.Fields{
				"id_param": idStr,
				"error":    err.Error(),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	span.SetAttributes(attribute.String("entity.id", id.String()))

	// Check for permanent deletion query parameter
	permanent := c.Query("permanent") == "true"
	span.SetAttributes(attribute.Bool("permanent", permanent))

	logger.InfoContext(ctx, "Processing delete entity request",
		logger.Fields{
			"entity_id": id.String(),
			"permanent": permanent,
		})

	err = h.service.DeleteEntity(ctx, id, permanent)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/entities/:id",
			"error_type": "business_error",
		})

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found for deletion",
				logger.Fields{"entity_id": id.String()})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		case errors.Is(err, sharedErrors.ErrEntityHasChildren):
			logger.WarnContext(ctx, "Entity deletion failed - has children",
				logger.Fields{"entity_id": id.String()})
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete entity with children"})
		default:
			logger.ErrorContext(ctx, "Entity deletion failed",
				logger.Fields{
					"entity_id": id.String(),
					"error":     err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities/:id",
		"status":   "success",
	})

	logger.InfoContext(ctx, "Entity deleted successfully",
		logger.Fields{
			"entity_id": id.String(),
			"permanent": permanent,
		})

	if permanent {
		c.JSON(http.StatusOK, gin.H{"message": "Entity permanently deleted"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Entity deleted"})
	}
}

// RestoreEntity handles entity restoration requests
func (h *EntityHandler) RestoreEntity(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.restore_entity",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		logger.WarnContext(ctx, "Invalid entity ID",
			logger.Fields{
				"id_param": idStr,
				"error":    err.Error(),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	span.SetAttributes(attribute.String("entity.id", id.String()))

	logger.InfoContext(ctx, "Processing restore entity request",
		logger.Fields{"entity_id": id.String()})

	err = h.service.RestoreEntity(ctx, id)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found for restoration",
				logger.Fields{"entity_id": id.String()})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		default:
			logger.ErrorContext(ctx, "Entity restoration failed",
				logger.Fields{
					"entity_id": id.String(),
					"error":     err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	logger.InfoContext(ctx, "Entity restored successfully",
		logger.Fields{"entity_id": id.String()})

	c.JSON(http.StatusOK, gin.H{"message": "Entity restored successfully"})
}

// ListEntities handles entity listing requests
func (h *EntityHandler) ListEntities(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.list_entities",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities",
	})
	defer timer.Stop()

	// Parse query parameters
	req := entity.ListEntitiesRequest{
		Limit:  50, // Default limit
		Offset: 0,  // Default offset
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		entityType := entity.EntityType(typeStr)
		req.Type = &entityType
	}

	if parentIDStr := c.Query("parent_id"); parentIDStr != "" {
		if parentID, err := uuid.Parse(parentIDStr); err == nil {
			req.ParentID = &parentID
		}
	}

	if activeStr := c.Query("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			req.IsActive = &active
		}
	}

	if hiddenStr := c.Query("hidden"); hiddenStr != "" {
		if hidden, err := strconv.ParseBool(hiddenStr); err == nil {
			req.IsHidden = &hidden
		}
	}

	// Add query parameters to span
	span.SetAttributes(
		attribute.Int("limit", req.Limit),
		attribute.Int("offset", req.Offset),
	)

	logger.InfoContext(ctx, "Processing list entities request",
		logger.Fields{
			"limit":  req.Limit,
			"offset": req.Offset,
		})

	entities, err := h.service.ListEntities(ctx, req)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/entities",
			"error_type": "business_error",
		})

		logger.ErrorContext(ctx, "Failed to list entities",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities",
		"status":   "success",
	})

	logger.InfoContext(ctx, "Entities listed successfully",
		logger.Fields{
			"count":  len(entities),
			"limit":  req.Limit,
			"offset": req.Offset,
		})

	c.JSON(http.StatusOK, gin.H{
		"entities": entities,
		"offset":   req.Offset,
		"limit":    req.Limit,
		"count":    len(entities),
	})
}

// GetEntityChildren handles entity children retrieval requests
func (h *EntityHandler) GetEntityChildren(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.get_entity_children",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		logger.WarnContext(ctx, "Invalid entity ID",
			logger.Fields{
				"id_param": idStr,
				"error":    err.Error(),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	span.SetAttributes(attribute.String("entity.id", id.String()))

	logger.InfoContext(ctx, "Processing get entity children request",
		logger.Fields{"entity_id": id.String()})

	children, err := h.service.GetEntityChildren(ctx, id)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found",
				logger.Fields{"entity_id": id.String()})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		default:
			logger.ErrorContext(ctx, "Failed to get entity children",
				logger.Fields{
					"entity_id": id.String(),
					"error":     err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	logger.InfoContext(ctx, "Entity children retrieved successfully",
		logger.Fields{
			"entity_id":      id.String(),
			"children_count": len(children),
		})

	c.JSON(http.StatusOK, gin.H{
		"children": children,
		"count":    len(children),
	})
}

// GetEntityAncestors handles entity ancestors retrieval requests
func (h *EntityHandler) GetEntityAncestors(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.get_entity_ancestors",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		logger.WarnContext(ctx, "Invalid entity ID",
			logger.Fields{
				"id_param": idStr,
				"error":    err.Error(),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	span.SetAttributes(attribute.String("entity.id", id.String()))

	logger.InfoContext(ctx, "Processing get entity ancestors request",
		logger.Fields{"entity_id": id.String()})

	ancestors, err := h.service.GetEntityAncestors(ctx, id)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found",
				logger.Fields{"entity_id": id.String()})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		default:
			logger.ErrorContext(ctx, "Failed to get entity ancestors",
				logger.Fields{
					"entity_id": id.String(),
					"error":     err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	logger.InfoContext(ctx, "Entity ancestors retrieved successfully",
		logger.Fields{
			"entity_id":       id.String(),
			"ancestors_count": len(ancestors),
		})

	c.JSON(http.StatusOK, gin.H{
		"ancestors": ancestors,
		"count":     len(ancestors),
	})
}

// GetEntityWithHierarchy handles entity with hierarchy retrieval requests
func (h *EntityHandler) GetEntityWithHierarchy(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.get_entity_with_hierarchy",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		logger.WarnContext(ctx, "Invalid entity ID",
			logger.Fields{
				"id_param": idStr,
				"error":    err.Error(),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	span.SetAttributes(attribute.String("entity.id", id.String()))

	logger.InfoContext(ctx, "Processing get entity with hierarchy request",
		logger.Fields{"entity_id": id.String()})

	entityWithHierarchy, err := h.service.GetEntityWithHierarchy(ctx, id)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found",
				logger.Fields{"entity_id": id.String()})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		default:
			logger.ErrorContext(ctx, "Failed to get entity with hierarchy",
				logger.Fields{
					"entity_id": id.String(),
					"error":     err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	logger.InfoContext(ctx, "Entity with hierarchy retrieved successfully",
		logger.Fields{
			"entity_id":    id.String(),
			"entity_level": entityWithHierarchy.Level,
		})

	c.JSON(http.StatusOK, entityWithHierarchy)
}

// GetEntityTree handles entity tree retrieval requests
func (h *EntityHandler) GetEntityTree(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.get_entity_tree",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	timer := h.metrics.Timer("http_request_duration", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities/tree",
	})
	defer timer.Stop()

	logger.InfoContext(ctx, "Processing get entity tree request")

	entityTree, err := h.service.GetEntityTree(ctx)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
			"method":     c.Request.Method,
			"endpoint":   "/api/v1/entities/tree",
			"error_type": "business_error",
		})

		logger.ErrorContext(ctx, "Failed to get entity tree",
			logger.Fields{"error": err.Error()})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
		"method":   c.Request.Method,
		"endpoint": "/api/v1/entities/tree",
		"status":   "success",
	})

	logger.InfoContext(ctx, "Entity tree retrieved successfully",
		logger.Fields{"entities_count": len(entityTree)})

	c.JSON(http.StatusOK, gin.H{
		"tree":  entityTree,
		"count": len(entityTree),
	})
}

// GetNextSequence handles sequence number retrieval requests
func (h *EntityHandler) GetNextSequence(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.get_next_sequence",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	var req entity.EntitySequenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		logger.WarnContext(ctx, "Invalid request payload",
			logger.Fields{
				"error":        err.Error(),
				"content_type": c.GetHeader("Content-Type"),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(
		attribute.String("entity.id", req.EntityID.String()),
		attribute.String("sequence.key", req.Key),
		attribute.Int("sequence.fiscal_year", req.FiscalYear),
	)

	logger.InfoContext(ctx, "Processing get next sequence request",
		logger.Fields{
			"entity_id":   req.EntityID.String(),
			"key":         req.Key,
			"fiscal_year": req.FiscalYear,
		})

	sequence, err := h.service.GetNextSequence(ctx, req)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found",
				logger.Fields{"entity_id": req.EntityID.String()})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		default:
			logger.ErrorContext(ctx, "Failed to get next sequence",
				logger.Fields{
					"entity_id":   req.EntityID.String(),
					"key":         req.Key,
					"fiscal_year": req.FiscalYear,
					"error":       err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	logger.InfoContext(ctx, "Next sequence retrieved successfully",
		logger.Fields{
			"entity_id":   req.EntityID.String(),
			"key":         req.Key,
			"fiscal_year": req.FiscalYear,
			"sequence":    sequence,
		})

	c.JSON(http.StatusOK, gin.H{
		"sequence":    sequence,
		"entity_id":   req.EntityID.String(),
		"key":         req.Key,
		"fiscal_year": req.FiscalYear,
	})
}

// ResetSequence handles sequence number reset requests
func (h *EntityHandler) ResetSequence(c *gin.Context) {
	ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

	ctx, span := h.tracing.StartSpan(ctx, "http.reset_sequence",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	var req entity.EntitySequenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		logger.WarnContext(ctx, "Invalid request payload",
			logger.Fields{
				"error":        err.Error(),
				"content_type": c.GetHeader("Content-Type"),
			})

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(
		attribute.String("entity.id", req.EntityID.String()),
		attribute.String("sequence.key", req.Key),
		attribute.Int("sequence.fiscal_year", req.FiscalYear),
	)

	logger.InfoContext(ctx, "Processing reset sequence request",
		logger.Fields{
			"entity_id":   req.EntityID.String(),
			"key":         req.Key,
			"fiscal_year": req.FiscalYear,
		})

	err := h.service.ResetSequence(ctx, req)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())

		switch {
		case errors.Is(err, sharedErrors.ErrEntityNotFound):
			logger.WarnContext(ctx, "Entity not found",
				logger.Fields{"entity_id": req.EntityID.String()})
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		default:
			logger.ErrorContext(ctx, "Failed to reset sequence",
				logger.Fields{
					"entity_id":   req.EntityID.String(),
					"key":         req.Key,
					"fiscal_year": req.FiscalYear,
					"error":       err.Error(),
				})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	logger.InfoContext(ctx, "Sequence reset successfully",
		logger.Fields{
			"entity_id":   req.EntityID.String(),
			"key":         req.Key,
			"fiscal_year": req.FiscalYear,
		})

	c.JSON(http.StatusOK, gin.H{
		"message":     "Sequence reset successfully",
		"entity_id":   req.EntityID.String(),
		"key":         req.Key,
		"fiscal_year": req.FiscalYear,
	})
}

// OrganizationGoaHandler implements the GOA organization service following the data flow pattern
type OrganizationGoaHandler struct {
	entityService entity.Service
	tracing       tracing.TracingService
	metrics       *metrics.MetricsService
}

// NewOrganizationGoaHandler creates a new GOA organization handler following Clean Architecture pattern
func NewOrganizationGoaHandler(entitySvc entity.Service, tracing tracing.TracingService, metrics *metrics.MetricsService) organization.Service {
	return &OrganizationGoaHandler{
		entityService: entitySvc,
		tracing:       tracing,
		metrics:       metrics,
	}
}

// Create creates a new organization following the data flow pattern
func (h *OrganizationGoaHandler) Create(ctx context.Context, p *organization.CreateOrganizationPayload) (*organization.Organization, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.create",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.name", p.Name),
			attribute.String("entity.type", string(p.EntityType)),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_create_duration", metrics.Fields{
		"operation": "create",
	})
	defer timer.Stop()

	// Convert GOA payload to entity domain request
	req := entity.CreateEntityRequest{
		Name:          p.Name,
		Code:          generateEntityCode(p.Name), // Generate code from name
		Type:          entity.EntityType(p.EntityType),
		IsActive:      true,
		IsHidden:      false,
		AccrualMethod: true, // Default to accrual accounting
		FYStartMonth:  1,    // Default fiscal year starts in January
		Address:       map[string]any{},
		Settings:      map[string]any{},
		Metadata:      map[string]any{},
	}

	// Debug log to check values
	logger.Info("Creating organization with values", logger.Fields{
		"name":           req.Name,
		"code":           req.Code,
		"type":           string(req.Type),
		"fy_start_month": req.FYStartMonth,
		"accrual_method": req.AccrualMethod,
	})

	// Set parent ID if provided
	if p.ParentID != nil {
		if parentUUID, err := uuid.Parse(*p.ParentID); err == nil {
			req.ParentID = &parentUUID
		}
	}

	// Map organization settings if provided
	if p.Settings != nil {
		// Extract fiscal year start from settings if available
		if p.Settings.FiscalYearStart != "" {
			// Parse fiscal year start (format: "MM-DD")
			if len(p.Settings.FiscalYearStart) >= 2 {
				if month := parseFiscalYearMonth(p.Settings.FiscalYearStart); month > 0 {
					req.FYStartMonth = month
				}
			}
		}

		// Add organization settings to entity settings
		if p.Settings.Timezone != "" {
			req.Settings["timezone"] = p.Settings.Timezone
		}
		if p.Settings.Currency != "" {
			req.Settings["currency"] = p.Settings.Currency
		}
		if p.Settings.BusinessHours != nil {
			req.Settings["business_hours"] = p.Settings.BusinessHours
		}
		if p.Settings.Integrations != nil {
			req.Settings["integrations"] = p.Settings.Integrations
		}
		if p.Settings.Preferences != nil {
			req.Settings["preferences"] = p.Settings.Preferences
		}
	}

	// Add additional fields if provided
	if p.Settings != nil || p.Description != nil {
		req.Metadata = make(map[string]any)
		if p.Description != nil {
			req.Metadata["description"] = *p.Description
		}
		if p.LegalName != nil {
			req.Metadata["legal_name"] = *p.LegalName
		}
		if p.LegalEntityType != nil {
			req.Metadata["legal_entity_type"] = *p.LegalEntityType
		}
		if p.Website != nil {
			req.Metadata["website"] = *p.Website
		}
		if p.Industry != nil {
			req.Metadata["industry"] = *p.Industry
		}
	}

	// Create entity through domain service
	entityResult, err := h.entityService.CreateEntity(ctx, req)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("organization_create_errors_total", metrics.Fields{
			"entity_type": string(p.EntityType),
		})
		return nil, "", mapEntityError(err)
	}

	// Convert entity result to GOA organization
	org := &organization.Organization{
		ID:         entityResult.ID.String(),
		Name:       entityResult.Name,
		EntityType: organization.EntityType(entityResult.Type),
		Status:     organization.OrganizationStatus("ACTIVE"),
		CreatedAt:  entityResult.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  entityResult.UpdatedAt.Format(time.RFC3339),
	}

	// Add optional fields from metadata
	if entityResult.Metadata != nil {
		if desc, ok := entityResult.Metadata["description"].(string); ok {
			org.Description = &desc
		}
		if legalName, ok := entityResult.Metadata["legal_name"].(string); ok {
			org.LegalName = &legalName
		}
		if legalType, ok := entityResult.Metadata["legal_entity_type"].(string); ok {
			org.LegalEntityType = &legalType
		}
		if website, ok := entityResult.Metadata["website"].(string); ok {
			org.Website = &website
		}
		if industry, ok := entityResult.Metadata["industry"].(string); ok {
			org.Industry = &industry
		}
	}

	// Set parent ID if exists
	if entityResult.ParentID != nil {
		parentIDStr := entityResult.ParentID.String()
		org.ParentID = &parentIDStr
	}

	h.metrics.IncrementCounter("organization_create_total", metrics.Fields{
		"entity_type": string(p.EntityType),
	})
	span.SetAttributes(attribute.String("result.organization_id", org.ID))

	return org, "default", nil
}

// Get retrieves an organization by ID following the data flow pattern
func (h *OrganizationGoaHandler) Get(ctx context.Context, p *organization.GetPayload) (*organization.Organization, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.get",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_get_duration", metrics.Fields{
		"operation": "get",
	})
	defer timer.Stop()

	// Parse organization ID
	orgUUID, err := uuid.Parse(p.ID)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("organization_get_errors_total", metrics.Fields{
			"error_type": "invalid_id",
		})
		return nil, "", organization.MakeBadRequest(err)
	}

	// Get entity through domain service
	entityResult, err := h.entityService.GetEntityByID(ctx, orgUUID)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("organization_get_errors_total", metrics.Fields{
			"error_type": "not_found",
		})
		return nil, "", mapEntityError(err)
	}

	// Convert entity to GOA organization
	org := &organization.Organization{
		ID:         entityResult.ID.String(),
		Name:       entityResult.Name,
		EntityType: organization.EntityType(entityResult.Type),
		Status:     organization.OrganizationStatus("ACTIVE"),
		CreatedAt:  entityResult.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  entityResult.UpdatedAt.Format(time.RFC3339),
	}

	// Add optional fields from metadata
	if entityResult.Metadata != nil {
		if desc, ok := entityResult.Metadata["description"].(string); ok {
			org.Description = &desc
		}
		if legalName, ok := entityResult.Metadata["legal_name"].(string); ok {
			org.LegalName = &legalName
		}
		if legalType, ok := entityResult.Metadata["legal_entity_type"].(string); ok {
			org.LegalEntityType = &legalType
		}
		if website, ok := entityResult.Metadata["website"].(string); ok {
			org.Website = &website
		}
		if industry, ok := entityResult.Metadata["industry"].(string); ok {
			org.Industry = &industry
		}
	}

	// Set parent ID if exists
	if entityResult.ParentID != nil {
		parentIDStr := entityResult.ParentID.String()
		org.ParentID = &parentIDStr
	}

	h.metrics.IncrementCounter("organization_get_total", metrics.Fields{})
	return org, "default", nil
}

// List retrieves organizations with pagination following the data flow pattern
func (h *OrganizationGoaHandler) List(ctx context.Context, p *organization.ListPayload) (*organization.ListResult, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.list",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_list_duration", metrics.Fields{
		"operation": "list",
	})
	defer timer.Stop()

	// TODO: Integrate with existing entity listing logic using h.entityService.ListEntities
	result := &organization.ListResult{
		Data:       []*organization.Organization{},
		Pagination: &organization.PaginationMeta{CurrentPage: 1, PageSize: 20, TotalItems: 0, TotalPages: 0, HasNext: false, HasPrev: false},
	}

	h.metrics.IncrementCounter("organization_list_total", metrics.Fields{})
	return result, nil
}

// Update updates an existing organization following the data flow pattern
func (h *OrganizationGoaHandler) Update(ctx context.Context, p *organization.UpdateOrganizationPayload) (*organization.Organization, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.update",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_update_duration", metrics.Fields{
		"operation": "update",
	})
	defer timer.Stop()

	// TODO: Integrate with existing entity update logic using h.entityService.UpdateEntity
	org := &organization.Organization{
		ID:         p.ID,
		Name:       *p.Name,
		EntityType: *p.EntityType,
		Status:     *p.Status,
		CreatedAt:  "2024-01-01T00:00:00Z",
		UpdatedAt:  "2024-01-01T00:00:00Z",
	}

	h.metrics.IncrementCounter("organization_update_total", metrics.Fields{})
	return org, "default", nil
}

// Hierarchy retrieves organization hierarchy following the data flow pattern
func (h *OrganizationGoaHandler) Hierarchy(ctx context.Context, p *organization.HierarchyPayload) (*organization.OrganizationHierarchy, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.hierarchy",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.id", p.ID),
			attribute.Int("hierarchy.depth", int(p.Depth)),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_hierarchy_duration", metrics.Fields{
		"operation": "hierarchy",
	})
	defer timer.Stop()

	// TODO: Integrate with existing entity hierarchy logic using h.entityService.GetEntityWithHierarchy
	result := &organization.OrganizationHierarchy{
		Root:      &organization.OrganizationNode{ID: p.ID, Name: "Root Org", EntityType: "COMPANY", Status: "ACTIVE", Children: []string{}, Level: 0},
		Children:  []*organization.OrganizationNode{},
		Ancestors: []*organization.OrganizationNode{},
		Depth:     p.Depth,
	}

	h.metrics.IncrementCounter("organization_hierarchy_total", metrics.Fields{})
	return result, nil
}

// Archive archives an organization following the data flow pattern
func (h *OrganizationGoaHandler) Archive(ctx context.Context, p *organization.ArchivePayload) error {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.archive",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_archive_duration", metrics.Fields{
		"operation": "archive",
	})
	defer timer.Stop()

	logger.Info("Organization archive called", logger.Fields{
		"id": p.ID,
	})

	// TODO: Integrate with existing entity archival logic using h.entityService.DeleteEntity
	h.metrics.IncrementCounter("organization_archive_total", metrics.Fields{})
	return nil
}

// Helper functions for organization GOA handler

// generateEntityCode creates a standardized entity code from the name
func generateEntityCode(name string) string {
	// Convert to uppercase, replace spaces with underscores, and limit length
	code := strings.ReplaceAll(strings.ToUpper(name), " ", "_")
	if len(code) > 20 {
		code = code[:20]
	}
	return code
}

// parseFiscalYearMonth parses fiscal year start string (MM-DD format) to month number
func parseFiscalYearMonth(fyStart string) int {
	// Extract month from "MM-DD" format
	if len(fyStart) >= 2 {
		if month, err := strconv.Atoi(fyStart[:2]); err == nil {
			if month >= 1 && month <= 12 {
				return month
			}
		}
	}
	return 1 // Default to January if parsing fails
}

// mapEntityError converts domain errors to GOA service errors
func mapEntityError(err error) error {
	switch {
	case errors.Is(err, sharedErrors.ErrEntityNotFound):
		return organization.MakeNotFound(err)
	case errors.Is(err, sharedErrors.ErrEntityNameExists):
		return organization.MakeConflict(err)
	case errors.Is(err, sharedErrors.ErrEntityCodeExists):
		return organization.MakeConflict(err)
	case errors.Is(err, sharedErrors.ErrInvalidInput):
		return organization.MakeBadRequest(err)
	case errors.Is(err, sharedErrors.ErrInvalidEntityType):
		return organization.MakeBadRequest(err)
	case sharedErrors.IsValidationError(err):
		return organization.MakeBadRequest(err)
	default:
		return goa.NewServiceError(err, "internal_error", false, false, false)
	}
}
