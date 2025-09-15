package handlers

// DEPRECATED: This file contains legacy Gin-based access request handlers.
// The system now uses Goa-based access request handlers.
// This file is kept for reference but all functionality has been replaced.

/*
import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/analytics"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// AccessRequestHandler handles access request workflow endpoints (DEPRECATED)
type AccessRequestHandler struct {
	accessRequestService     request.AccessRequestService
	conditionalAccessService conditional.ConditionalAccessService
	analyticsService         analytics.UserAnalyticsService
	tracing                  tracing.TracingService
	metrics                  *metrics.MetricsService
}

// NewAccessRequestHandler creates a new access request handler (DEPRECATED)
func NewAccessRequestHandler(
	accessRequestService request.AccessRequestService,
	conditionalAccessService conditional.ConditionalAccessService,
	analyticsService analytics.UserAnalyticsService,
	tracing tracing.TracingService,
	metrics *metrics.MetricsService,
) *AccessRequestHandler {
	return &AccessRequestHandler{
		accessRequestService:     accessRequestService,
		conditionalAccessService: conditionalAccessService,
		analyticsService:         analyticsService,
		tracing:                  tracing,
		metrics:                  metrics,
	}
}

// All methods commented out - use Goa handlers instead
// CreateAccessRequest, GetAccessRequest, ProcessAccessRequest, etc.
// are all implemented as Goa handlers in the main server
*/
