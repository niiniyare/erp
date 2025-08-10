package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/api/middleware"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// FeatureFlagWorkflowHandler handles feature flag workflow API endpoints
type FeatureFlagWorkflowHandler struct {
	workflowService featureflag.WorkflowService
}

// NewFeatureFlagWorkflowHandler creates a new workflow handler
func NewFeatureFlagWorkflowHandler(workflowService featureflag.WorkflowService) *FeatureFlagWorkflowHandler {
	return &FeatureFlagWorkflowHandler{
		workflowService: workflowService,
	}
}

// RequestFeatureFlagChange handles POST /api/v1/feature-flags/workflows/change
func (h *FeatureFlagWorkflowHandler) RequestFeatureFlagChange(c *gin.Context) {
	log := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWorkflowHandler",
		"method":  "RequestFeatureFlagChange",
	})

	var req featureflag.FeatureFlagChangeRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Invalid request body", logger.Fields{"error": err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.FlagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flag_name is required"})
		return
	}

	if req.ChangeType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "change_type is required"})
		return
	}

	if len(req.Justification) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "justification must be at least 10 characters"})
		return
	}

	result, err := h.workflowService.RequestFeatureFlagChange(c.Request.Context(), &req)
	if err != nil {
		log.Error("Failed to request feature flag change", logger.Fields{"error": err})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate workflow", "details": err.Error()})
		return
	}

	log.Info("Feature flag change workflow requested", logger.Fields{
		"workflow_id": result.WorkflowID,
		"flag_name":   req.FlagName,
		"change_type": req.ChangeType,
	})

	c.JSON(http.StatusAccepted, result)
}

// RequestBulkFeatureFlagChange handles POST /api/v1/feature-flags/workflows/bulk-change
func (h *FeatureFlagWorkflowHandler) RequestBulkFeatureFlagChange(c *gin.Context) {
	log := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWorkflowHandler",
		"method":  "RequestBulkFeatureFlagChange",
	})

	var req featureflag.BulkFeatureFlagChangeRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Invalid request body", logger.Fields{"error": err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate required fields
	if len(req.Changes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "changes array cannot be empty"})
		return
	}

	if len(req.Justification) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "justification must be at least 10 characters"})
		return
	}

	// Validate each change
	for i, change := range req.Changes {
		if change.FlagName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "flag_name is required",
				"index": i,
			})
			return
		}
		if change.ChangeType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "change_type is required",
				"index": i,
			})
			return
		}
	}

	result, err := h.workflowService.RequestBulkFeatureFlagChange(c.Request.Context(), &req)
	if err != nil {
		log.Error("Failed to request bulk feature flag change", logger.Fields{"error": err})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate bulk workflow", "details": err.Error()})
		return
	}

	log.Info("Bulk feature flag change workflow requested", logger.Fields{
		"workflow_id":  result.WorkflowID,
		"change_count": len(req.Changes),
	})

	c.JSON(http.StatusAccepted, result)
}

// ApproveFeatureFlagChange handles POST /api/v1/feature-flags/workflows/:workflow_id/approve
func (h *FeatureFlagWorkflowHandler) ApproveFeatureFlagChange(c *gin.Context) {
	log := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWorkflowHandler",
		"method":  "ApproveFeatureFlagChange",
	})

	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id is required"})
		return
	}

	var req featureflag.ApprovalRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Invalid request body", logger.Fields{"error": err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get approver ID from context (should be set by auth middleware)
	userID, exists := c.Get("authorized_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	approverUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	req.ApproverID = approverUUID

	err := h.workflowService.ApproveFeatureFlagChange(c.Request.Context(), workflowID, &req)
	if err != nil {
		log.Error("Failed to approve feature flag change", logger.Fields{
			"error":       err,
			"workflow_id": workflowID,
			"approver_id": approverUUID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve workflow", "details": err.Error()})
		return
	}

	log.Info("Feature flag change approved", logger.Fields{
		"workflow_id": workflowID,
		"approver_id": approverUUID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":     "Workflow approved successfully",
		"workflow_id": workflowID,
		"approved_by": approverUUID,
		"approved_at": time.Now(),
	})
}

// RejectFeatureFlagChange handles POST /api/v1/feature-flags/workflows/:workflow_id/reject
func (h *FeatureFlagWorkflowHandler) RejectFeatureFlagChange(c *gin.Context) {
	logger := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWorkflowHandler",
		"method":  "RejectFeatureFlagChange",
	})

	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id is required"})
		return
	}

	var req featureflag.RejectionRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body", logger.Fields{"error": err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate rejection comments
	if len(req.Comments) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rejection comments must be at least 10 characters"})
		return
	}

	// Get approver ID from context
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	approverUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	req.ApproverID = approverUUID

	err := h.workflowService.RejectFeatureFlagChange(c.Request.Context(), workflowID, &req)
	if err != nil {
		logger.Error("Failed to reject feature flag change", logger.Fields{
			"error":       err,
			"workflow_id": workflowID,
			"approver_id": approverUUID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject workflow", "details": err.Error()})
		return
	}

	logger.Info("Feature flag change rejected", logger.Fields{
		"workflow_id": workflowID,
		"approver_id": approverUUID,
		"reason":      req.Reason,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":     "Workflow rejected successfully",
		"workflow_id": workflowID,
		"rejected_by": approverUUID,
		"rejected_at": time.Now(),
		"reason":      req.Reason,
	})
}

// GetWorkflowStatus handles GET /api/v1/feature-flags/workflows/:workflow_id/status
func (h *FeatureFlagWorkflowHandler) GetWorkflowStatus(c *gin.Context) {
	logger := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWorkflowHandler",
		"method":  "GetWorkflowStatus",
	})

	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id is required"})
		return
	}

	status, err := h.workflowService.GetWorkflowStatus(c.Request.Context(), workflowID)
	if err != nil {
		logger.Error("Failed to get workflow status", logger.Fields{
			"error":       err,
			"workflow_id": workflowID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ListPendingApprovals handles GET /api/v1/feature-flags/workflows/pending-approvals
func (h *FeatureFlagWorkflowHandler) ListPendingApprovals(c *gin.Context) {
	logger := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWorkflowHandler",
		"method":  "ListPendingApprovals",
	})

	// Parse query parameters
	req := &featureflag.ListPendingApprovalsRequest{
		Limit:  20, // Default limit
		Offset: 0,  // Default offset
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	if flagName := c.Query("flag_name"); flagName != "" {
		req.FlagName = &flagName
	}

	// Get approver ID from context (for filtering approvals assigned to current user)
	if userID, exists := c.Get(middleware.UserIDKey); exists {
		if approverUUID, ok := userID.(uuid.UUID); ok {
			req.ApproverID = &approverUUID
		}
	}

	response, err := h.workflowService.ListPendingApprovals(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to list pending approvals", logger.Fields{"error": err})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending approvals", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelWorkflow handles DELETE /api/v1/feature-flags/workflows/:workflow_id
func (h *FeatureFlagWorkflowHandler) CancelWorkflow(c *gin.Context) {
	logger := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWorkflowHandler",
		"method":  "CancelWorkflow",
	})

	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id is required"})
		return
	}

	// Parse optional reason from request body
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req) // Optional, ignore errors

	if req.Reason == "" {
		req.Reason = "Cancelled by user"
	}

	err := h.workflowService.CancelWorkflow(c.Request.Context(), workflowID, req.Reason)
	if err != nil {
		logger.Error("Failed to cancel workflow", logger.Fields{
			"error":       err,
			"workflow_id": workflowID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel workflow", "details": err.Error()})
		return
	}

	logger.Info("Workflow cancelled", logger.Fields{
		"workflow_id": workflowID,
		"reason":      req.Reason,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":     "Workflow cancelled successfully",
		"workflow_id": workflowID,
		"reason":      req.Reason,
		"cancelled_at": time.Now(),
	})
}

// ScheduleAutoRollback handles POST /api/v1/feature-flags/workflows/auto-rollback
func (h *FeatureFlagWorkflowHandler) ScheduleAutoRollback(c *gin.Context) {
	logger := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWorkflowHandler",
		"method":  "ScheduleAutoRollback",
	})

	var req featureflag.ScheduleAutoRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body", logger.Fields{"error": err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.FlagID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flag_id is required"})
		return
	}

	if req.FlagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flag_name is required"})
		return
	}

	if req.RollbackAt.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rollback_at is required"})
		return
	}

	// Validate rollback time is in the future
	if req.RollbackAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rollback_at must be in the future"})
		return
	}

	result, err := h.workflowService.ScheduleAutoRollback(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to schedule auto-rollback", logger.Fields{"error": err})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule auto-rollback", "details": err.Error()})
		return
	}

	logger.Info("Auto-rollback scheduled", logger.Fields{
		"schedule_id": result.ScheduleID,
		"flag_name":   req.FlagName,
		"rollback_at": req.RollbackAt,
	})

	c.JSON(http.StatusCreated, result)
}

// CancelAutoRollback handles DELETE /api/v1/feature-flags/workflows/auto-rollback/:schedule_id
func (h *FeatureFlagWorkflowHandler) CancelAutoRollback(c *gin.Context) {
	logger := logger.WithFields(logger.Fields{
		"handler": "FeatureFlagWorkflowHandler",
		"method":  "CancelAutoRollback",
	})

	scheduleID := c.Param("schedule_id")
	if scheduleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_id is required"})
		return
	}

	err := h.workflowService.CancelAutoRollback(c.Request.Context(), scheduleID)
	if err != nil {
		logger.Error("Failed to cancel auto-rollback", logger.Fields{
			"error":       err,
			"schedule_id": scheduleID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel auto-rollback", "details": err.Error()})
		return
	}

	logger.Info("Auto-rollback cancelled", logger.Fields{
		"schedule_id": scheduleID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":      "Auto-rollback cancelled successfully",
		"schedule_id":  scheduleID,
		"cancelled_at": time.Now(),
	})
}