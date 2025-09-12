package handlers

// DEPRECATED: This file contains legacy Gin-based workflow handlers.
// The system now uses comprehensive Goa-based feature flag handlers.
// Workflow functionality is available through the Goa API endpoints.

/*
import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/featureflag/workflow"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// FeatureFlagWorkflowHandler handles feature flag workflow API endpoints (DEPRECATED)
type FeatureFlagWorkflowHandler struct {
	workflowService featureflag.WorkflowService
}

// NewFeatureFlagWorkflowHandler creates a new workflow handler (DEPRECATED)
func NewFeatureFlagWorkflowHandler(workflowService featureflag.WorkflowService) *FeatureFlagWorkflowHandler {
	return &FeatureFlagWorkflowHandler{
		workflowService: workflowService,
	}
}

// All workflow handler methods commented out - use Goa handlers instead
// RequestFeatureFlagChange, RequestBulkFeatureFlagChange, ApproveFeatureFlagChange, etc.
// are all implemented as Goa handlers in the main server
*/