package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/user"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// AccessRequestHandler handles access request workflow endpoints
type AccessRequestHandler struct {
	accessRequestService    user.AccessRequestService
	conditionalAccessService user.ConditionalAccessService
	analyticsService        user.UserAnalyticsService
	tracing                 *tracing.TracingService
	metrics                 *metrics.MetricsService
}

// NewAccessRequestHandler creates a new access request handler
func NewAccessRequestHandler(
	accessRequestService user.AccessRequestService,
	conditionalAccessService user.ConditionalAccessService,
	analyticsService user.UserAnalyticsService,
	tracing *tracing.TracingService,
	metrics *metrics.MetricsService,
) *AccessRequestHandler {
	return &AccessRequestHandler{
		accessRequestService:      accessRequestService,
		conditionalAccessService:  conditionalAccessService,
		analyticsService:          analyticsService,
		tracing:                   tracing,
		metrics:                   metrics,
	}
}

// CreateAccessRequest creates a new access request
func (h *AccessRequestHandler) CreateAccessRequest(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.CreateAccessRequest")
	defer span.End()

	var req user.CreateAccessRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get requester ID from context/headers
	requesterIDStr := c.GetHeader("X-User-ID")
	if requesterIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required in X-User-ID header"})
		return
	}

	requesterID, err := uuid.Parse(requesterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	span.SetAttributes(
		attribute.String("requester_id", requesterID.String()),
		attribute.String("request_type", string(req.RequestType)),
		attribute.String("entity_id", req.EntityID.String()),
	)

	// Create access request
	accessRequest, err := h.accessRequestService.CreateAccessRequest(ctx, &req, requesterID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create access request")
		logger.Error("Failed to create access request", logger.Fields{
			"error":        err.Error(),
			"requester_id": requesterID,
			"request_type": req.RequestType,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access request", "details": err.Error()})
		return
	}

	h.metrics.IncrementCounter("access_request_created", map[string]any{"type": string(req.RequestType)})

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Access request created successfully",
		"access_request": accessRequest,
	})
}

// ProcessAccessRequest processes an access request (approve/reject)
func (h *AccessRequestHandler) ProcessAccessRequest(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.ProcessAccessRequest")
	defer span.End()

	requestIDStr := c.Param("id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID format"})
		return
	}

	var req user.AccessRequestApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get approver ID from context/headers
	approverIDStr := c.GetHeader("X-User-ID")
	if approverIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required in X-User-ID header"})
		return
	}

	approverID, err := uuid.Parse(approverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	span.SetAttributes(
		attribute.String("request_id", requestID.String()),
		attribute.String("approver_id", approverID.String()),
		attribute.String("action", req.Action),
	)

	// Process access request
	processedRequest, err := h.accessRequestService.ProcessAccessRequest(ctx, requestID, &req, approverID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to process access request")
		logger.Error("Failed to process access request", logger.Fields{
			"error":      err.Error(),
			"request_id": requestID,
			"action":     req.Action,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process access request", "details": err.Error()})
		return
	}

	h.metrics.IncrementCounter("access_request_processed", map[string]any{"action": req.Action})

	c.JSON(http.StatusOK, gin.H{
		"message":        "Access request processed successfully",
		"access_request": processedRequest,
	})
}

// GetAccessRequest gets an access request by ID
func (h *AccessRequestHandler) GetAccessRequest(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.GetAccessRequest")
	defer span.End()

	requestIDStr := c.Param("id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID format"})
		return
	}

	span.SetAttributes(attribute.String("request_id", requestID.String()))

	// Get access request
	accessRequest, err := h.accessRequestService.GetAccessRequest(ctx, requestID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access request")
		logger.Error("Failed to get access request", logger.Fields{
			"error":      err.Error(),
			"request_id": requestID,
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "Access request not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_request": accessRequest,
	})
}

// ListAccessRequests lists access requests with optional filters
func (h *AccessRequestHandler) ListAccessRequests(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.ListAccessRequests")
	defer span.End()

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.Query("status")
	requestType := c.Query("type")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build list request
	listReq := &user.ListAccessRequestsRequest{
		Limit:  limit,
		Offset: offset,
	}

	if status != "" {
		approvalStatus := user.ApprovalStatus(status)
		listReq.ApprovalStatus = &approvalStatus
	}

	if requestType != "" {
		reqType := user.RequestType(requestType)
		listReq.RequestType = &reqType
	}

	// Get user ID from header for filtering (optional)
	if userIDStr := c.GetHeader("X-User-ID"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			listReq.RequesterID = &userID
		}
	}

	span.SetAttributes(
		attribute.Int("limit", limit),
		attribute.Int("offset", offset),
		attribute.String("status", status),
		attribute.String("type", requestType),
	)

	// List access requests
	accessRequests, err := h.accessRequestService.ListAccessRequests(ctx, listReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list access requests")
		logger.Error("Failed to list access requests", logger.Fields{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list access requests", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_requests": accessRequests,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(accessRequests),
		},
	})
}

// RevokeAccessRequest revokes an access request
func (h *AccessRequestHandler) RevokeAccessRequest(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.RevokeAccessRequest")
	defer span.End()

	requestIDStr := c.Param("id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID format"})
		return
	}

	span.SetAttributes(attribute.String("request_id", requestID.String()))

	// Revoke access request
	revokedRequest, err := h.accessRequestService.RevokeAccessRequest(ctx, requestID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to revoke access request")
		logger.Error("Failed to revoke access request", logger.Fields{
			"error":      err.Error(),
			"request_id": requestID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke access request", "details": err.Error()})
		return
	}

	h.metrics.IncrementCounter("access_request_revoked", map[string]any{})

	c.JSON(http.StatusOK, gin.H{
		"message":        "Access request revoked successfully",
		"access_request": revokedRequest,
	})
}

// GetAccessRequestStats gets access request statistics
func (h *AccessRequestHandler) GetAccessRequestStats(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.GetAccessRequestStats")
	defer span.End()

	// Parse date parameters
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	var fromDate, toDate *time.Time

	if fromDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			fromDate = &parsed
		}
	}

	if toDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", toDateStr); err == nil {
			toDate = &parsed
		}
	}

	// Get statistics
	stats, err := h.accessRequestService.GetAccessRequestStats(ctx, fromDate, toDate)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access request statistics")
		logger.Error("Failed to get access request statistics", logger.Fields{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// EvaluateConditionalAccess evaluates conditional access for a user
func (h *AccessRequestHandler) EvaluateConditionalAccess(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.EvaluateConditionalAccess")
	defer span.End()

	var accessContext user.AccessContext
	if err := c.ShouldBindJSON(&accessContext); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Fill in IP address and user agent from request
	accessContext.IPAddress = c.ClientIP()
	accessContext.UserAgent = c.GetHeader("User-Agent")

	span.SetAttributes(
		attribute.String("user_id", accessContext.UserID.String()),
		attribute.String("ip_address", accessContext.IPAddress),
		attribute.String("requested_resource", accessContext.RequestedResource),
	)

	// Evaluate conditional access
	result, err := h.conditionalAccessService.EvaluateAccess(ctx, &accessContext)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to evaluate conditional access")
		logger.Error("Failed to evaluate conditional access", logger.Fields{
			"error":   err.Error(),
			"user_id": accessContext.UserID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to evaluate conditional access", "details": err.Error()})
		return
	}

	h.metrics.IncrementCounter("conditional_access_evaluation", map[string]any{"decision": string(result.Decision)})

	c.JSON(http.StatusOK, gin.H{
		"evaluation_result": result,
	})
}

// GetUserBehaviorAnalytics gets user behavior analytics
func (h *AccessRequestHandler) GetUserBehaviorAnalytics(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.GetUserBehaviorAnalytics")
	defer span.End()

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Parse analysis window parameter
	analysisWindowStr := c.DefaultQuery("analysis_window", "720h") // Default 30 days
	analysisWindow, err := time.ParseDuration(analysisWindowStr)
	if err != nil {
		analysisWindow = 30 * 24 * time.Hour // Default to 30 days
	}

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("analysis_window", analysisWindow.String()),
	)

	// Analyze user behavior
	behaviorPattern, err := h.analyticsService.AnalyzeUserBehavior(ctx, userID, analysisWindow)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to analyze user behavior")
		logger.Error("Failed to analyze user behavior", logger.Fields{
			"error":   err.Error(),
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze user behavior", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"behavior_pattern": behaviorPattern,
	})
}

// GetUserRiskAssessment gets user risk assessment
func (h *AccessRequestHandler) GetUserRiskAssessment(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.GetUserRiskAssessment")
	defer span.End()

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	span.SetAttributes(attribute.String("user_id", userID.String()))

	// Assess user risk
	riskAssessment, err := h.analyticsService.AssessUserRisk(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to assess user risk")
		logger.Error("Failed to assess user risk", logger.Fields{
			"error":   err.Error(),
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assess user risk", "details": err.Error()})
		return
	}

	h.metrics.ObserveHistogram("user_risk_score", float64(riskAssessment.OverallRiskScore), map[string]any{
		"risk_level": riskAssessment.RiskLevel,
	})

	c.JSON(http.StatusOK, gin.H{
		"risk_assessment": riskAssessment,
	})
}

// CreateConditionalAccessRule creates a new conditional access rule
func (h *AccessRequestHandler) CreateConditionalAccessRule(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.CreateConditionalAccessRule")
	defer span.End()

	var rule user.ConditionalAccessRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Generate ID and timestamps
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// Get creator ID from header
	if creatorIDStr := c.GetHeader("X-User-ID"); creatorIDStr != "" {
		if creatorID, err := uuid.Parse(creatorIDStr); err == nil {
			rule.CreatedBy = creatorID
		}
	}

	span.SetAttributes(
		attribute.String("rule_id", rule.ID.String()),
		attribute.String("rule_type", string(rule.RuleType)),
		attribute.String("entity_id", rule.EntityID.String()),
	)

	// Create rule
	err := h.conditionalAccessService.CreateRule(ctx, &rule)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create conditional access rule")
		logger.Error("Failed to create conditional access rule", logger.Fields{
			"error":   err.Error(),
			"rule_id": rule.ID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule", "details": err.Error()})
		return
	}

	h.metrics.IncrementCounter("conditional_access_rule_created", map[string]any{"type": string(rule.RuleType)})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Conditional access rule created successfully",
		"rule":    rule,
	})
}

// GetUserPersonalizedInsights gets personalized insights for a user
func (h *AccessRequestHandler) GetUserPersonalizedInsights(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.GetUserPersonalizedInsights")
	defer span.End()

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	span.SetAttributes(attribute.String("user_id", userID.String()))

	// Get personalized insights
	insights, err := h.analyticsService.GetPersonalizedInsights(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get personalized insights")
		logger.Error("Failed to get personalized insights", logger.Fields{
			"error":   err.Error(),
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get insights", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"insights": insights,
	})
}

// DetectUserAnomalies detects anomalies for a user activity
func (h *AccessRequestHandler) DetectUserAnomalies(c *gin.Context) {
	ctx, span := h.tracing.StartSpan(c.Request.Context(), "AccessRequestHandler.DetectUserAnomalies")
	defer span.End()

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var activity user.UserActivity
	if err := c.ShouldBindJSON(&activity); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Set user ID and fill context
	activity.UserID = userID
	activity.IPAddress = c.ClientIP()
	activity.UserAgent = c.GetHeader("User-Agent")
	activity.Timestamp = time.Now()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("activity_type", activity.ActivityType),
		attribute.String("resource_accessed", activity.ResourceAccessed),
	)

	// Detect anomalies
	anomalies, err := h.analyticsService.DetectAnomalies(ctx, userID, &activity)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to detect anomalies")
		logger.Error("Failed to detect anomalies", logger.Fields{
			"error":   err.Error(),
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to detect anomalies", "details": err.Error()})
		return
	}

	h.metrics.IncrementCounter("user_anomaly_detection", map[string]any{
		"anomaly_count": strconv.Itoa(len(anomalies)),
	})

	c.JSON(http.StatusOK, gin.H{
		"anomalies": anomalies,
		"count":     len(anomalies),
	})
}