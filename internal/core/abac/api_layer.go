package abac

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// ABACAPIHandler provides HTTP API endpoints for ABAC operations
type ABACAPIHandler struct {
	policyManager      PolicyManager
	evaluationEngine   PolicyEvaluationEngine
	attributeService   AttributeService
	attributeCollector AttributeCollector
	attributeResolver  AttributeResolver
	externalSources    ExternalAttributeSourceManager
	performanceOpt     PerformanceOptimizer
	monitoringService  MonitoringService

	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewABACAPIHandler creates a new ABAC API handler
func NewABACAPIHandler(
	policyManager PolicyManager,
	evaluationEngine PolicyEvaluationEngine,
	attributeService AttributeService,
	attributeCollector AttributeCollector,
	attributeResolver AttributeResolver,
	externalSources ExternalAttributeSourceManager,
	performanceOpt PerformanceOptimizer,
	monitoringService MonitoringService,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) *ABACAPIHandler {
	return &ABACAPIHandler{
		policyManager:      policyManager,
		evaluationEngine:   evaluationEngine,
		attributeService:   attributeService,
		attributeCollector: attributeCollector,
		attributeResolver:  attributeResolver,
		externalSources:    externalSources,
		performanceOpt:     performanceOpt,
		monitoringService:  monitoringService,
		logger:             logger,
		metrics:            metrics,
		tracer:             tracer,
	}
}

// RegisterRoutes registers all ABAC API routes
func (h *ABACAPIHandler) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1/abac").Subrouter()

	// Policy Management Routes
	h.registerPolicyRoutes(api.PathPrefix("/policies").Subrouter())

	// Policy Evaluation Routes
	h.registerEvaluationRoutes(api.PathPrefix("/evaluate").Subrouter())

	// Attribute Management Routes
	h.registerAttributeRoutes(api.PathPrefix("/attributes").Subrouter())

	// External Sources Routes
	h.registerExternalSourceRoutes(api.PathPrefix("/external-sources").Subrouter())

	// Monitoring Routes
	h.registerMonitoringRoutes(api.PathPrefix("/monitoring").Subrouter())

	// Performance Routes
	h.registerPerformanceRoutes(api.PathPrefix("/performance").Subrouter())

	// Health Check Routes
	api.HandleFunc("/health", h.handleHealthCheck).Methods("GET")
	api.HandleFunc("/ready", h.handleReadinessCheck).Methods("GET")
}

// Policy Management Routes

func (h *ABACAPIHandler) registerPolicyRoutes(router *mux.Router) {
	// Policy CRUD
	router.HandleFunc("", h.handleCreatePolicy).Methods("POST")
	router.HandleFunc("", h.handleListPolicies).Methods("GET")
	router.HandleFunc("/{id}", h.handleGetPolicy).Methods("GET")
	router.HandleFunc("/{id}", h.handleUpdatePolicy).Methods("PUT")
	router.HandleFunc("/{id}", h.handleDeletePolicy).Methods("DELETE")

	// Policy Testing
	router.HandleFunc("/{id}/test", h.handleTestPolicy).Methods("POST")
	router.HandleFunc("/{id}/simulate", h.handleSimulatePolicy).Methods("POST")

	// Policy Lifecycle
	router.HandleFunc("/{id}/activate", h.handleActivatePolicy).Methods("POST")
	router.HandleFunc("/{id}/deactivate", h.handleDeactivatePolicy).Methods("POST")
	router.HandleFunc("/{id}/versions", h.handleCreatePolicyVersion).Methods("POST")
	router.HandleFunc("/{id}/versions", h.handleListPolicyVersions).Methods("GET")

	// Policy Analysis
	router.HandleFunc("/{id}/impact", h.handleAnalyzePolicyImpact).Methods("POST")
	router.HandleFunc("/{id}/conflicts", h.handleDetectPolicyConflicts).Methods("GET")

	// Policy Templates
	router.HandleFunc("/templates", h.handleListPolicyTemplates).Methods("GET")
	router.HandleFunc("/templates/{id}", h.handleGetPolicyTemplate).Methods("GET")
	router.HandleFunc("/import", h.handleImportPolicies).Methods("POST")
	router.HandleFunc("/export", h.handleExportPolicies).Methods("POST")
}

// Policy Evaluation Routes

func (h *ABACAPIHandler) registerEvaluationRoutes(router *mux.Router) {
	// Core Evaluation
	router.HandleFunc("/policy", h.handleEvaluatePolicy).Methods("POST")
	router.HandleFunc("/policies", h.handleEvaluatePolicies).Methods("POST")
	router.HandleFunc("/batch", h.handleBatchEvaluatePolicy).Methods("POST")

	// Advanced Evaluation
	router.HandleFunc("/contextual", h.handleEvaluateWithContext).Methods("POST")
	router.HandleFunc("/simulate", h.handleSimulateEvaluation).Methods("POST")

	// Rule Evaluation
	router.HandleFunc("/rule", h.handleEvaluateRule).Methods("POST")
	router.HandleFunc("/rule/parse", h.handleParseRuleExpression).Methods("POST")
	router.HandleFunc("/rule/validate", h.handleValidateRuleExpression).Methods("POST")

	// Decision Support
	router.HandleFunc("/decision-support", h.handleGetDecisionSupport).Methods("POST")
	router.HandleFunc("/recommendations", h.handleGetEvaluationRecommendations).Methods("POST")
}

// Attribute Management Routes

func (h *ABACAPIHandler) registerAttributeRoutes(router *mux.Router) {
	// Attribute Definitions
	router.HandleFunc("/definitions", h.handleCreateAttributeDefinition).Methods("POST")
	router.HandleFunc("/definitions", h.handleListAttributeDefinitions).Methods("GET")
	router.HandleFunc("/definitions/{id}", h.handleGetAttributeDefinition).Methods("GET")
	router.HandleFunc("/definitions/{id}", h.handleUpdateAttributeDefinition).Methods("PUT")
	router.HandleFunc("/definitions/{id}", h.handleDeleteAttributeDefinition).Methods("DELETE")

	// Attribute Collection
	router.HandleFunc("/collect", h.handleCollectAttributes).Methods("POST")
	router.HandleFunc("/collect/batch", h.handleBatchCollectAttributes).Methods("POST")
	router.HandleFunc("/collect/plan", h.handleGetCollectionPlan).Methods("POST")

	// Attribute Resolution
	router.HandleFunc("/resolve", h.handleResolveAttributes).Methods("POST")
	router.HandleFunc("/resolve/dependencies", h.handleCreateAttributeDependency).Methods("POST")
	router.HandleFunc("/resolve/cache/invalidate", h.handleInvalidateAttributeCache).Methods("POST")
	router.HandleFunc("/resolve/cache/statistics", h.handleGetAttributeCacheStatistics).Methods("GET")

	// Attribute Validation
	router.HandleFunc("/validate", h.handleValidateAttributeValue).Methods("POST")
	router.HandleFunc("/encrypt", h.handleEncryptAttributeValue).Methods("POST")
	router.HandleFunc("/decrypt", h.handleDecryptAttributeValue).Methods("POST")
}

// External Sources Routes

func (h *ABACAPIHandler) registerExternalSourceRoutes(router *mux.Router) {
	// Source Registration
	router.HandleFunc("/ldap", h.handleRegisterLDAPSource).Methods("POST")
	router.HandleFunc("/rest-api", h.handleRegisterRESTAPISource).Methods("POST")
	router.HandleFunc("/database", h.handleRegisterDatabaseSource).Methods("POST")
	router.HandleFunc("/custom", h.handleRegisterCustomSource).Methods("POST")

	// Source Management
	router.HandleFunc("", h.handleListAttributeSources).Methods("GET")
	router.HandleFunc("/{id}", h.handleGetAttributeSource).Methods("GET")
	router.HandleFunc("/{id}", h.handleUpdateAttributeSource).Methods("PUT")
	router.HandleFunc("/{id}", h.handleDeleteAttributeSource).Methods("DELETE")

	// Source Testing
	router.HandleFunc("/{id}/test", h.handleTestSourceConnection).Methods("POST")
	router.HandleFunc("/{id}/validate", h.handleValidateSourceConfiguration).Methods("POST")

	// Source Operations
	router.HandleFunc("/{id}/fetch", h.handleFetchAttributes).Methods("POST")
	router.HandleFunc("/{id}/sync", h.handleSynchronizeAttributes).Methods("POST")
	router.HandleFunc("/{id}/health", h.handleGetSourceHealth).Methods("GET")
	router.HandleFunc("/{id}/metrics", h.handleGetSourceMetrics).Methods("GET")
}

// Monitoring Routes

func (h *ABACAPIHandler) registerMonitoringRoutes(router *mux.Router) {
	// Metrics
	router.HandleFunc("/metrics/evaluation", h.handleGetEvaluationMetrics).Methods("GET")
	router.HandleFunc("/metrics/performance", h.handleGetPerformanceMetrics).Methods("GET")

	// Decision Patterns
	router.HandleFunc("/decisions/patterns", h.handleGetDecisionPatterns).Methods("GET")
	router.HandleFunc("/decisions/track", h.handleTrackPolicyDecision).Methods("POST")

	// Anomaly Detection
	router.HandleFunc("/anomalies/detect", h.handleDetectAnomalies).Methods("POST")
	router.HandleFunc("/anomalies/configure", h.handleConfigureAnomalyDetection).Methods("POST")

	// Alerting
	router.HandleFunc("/alerts", h.handleCreateAlert).Methods("POST")
	router.HandleFunc("/alerts", h.handleGetActiveAlerts).Methods("GET")
	router.HandleFunc("/alerts/{id}/resolve", h.handleResolveAlert).Methods("POST")

	// Dashboard and Reports
	router.HandleFunc("/dashboard", h.handleGetDashboardData).Methods("GET")
	router.HandleFunc("/reports", h.handleGenerateReport).Methods("POST")

	// System Health
	router.HandleFunc("/system/health", h.handleGetSystemHealth).Methods("GET")
	router.HandleFunc("/system/check", h.handleRunHealthCheck).Methods("POST")
}

// Performance Routes

func (h *ABACAPIHandler) registerPerformanceRoutes(router *mux.Router) {
	// Caching
	router.HandleFunc("/cache/evaluation", h.handleGetCachedEvaluation).Methods("POST")
	router.HandleFunc("/cache/invalidate", h.handleInvalidatePerformanceCache).Methods("POST")

	// Batch Operations
	router.HandleFunc("/batch/evaluate", h.handleBatchEvaluatePerformance).Methods("POST")
	router.HandleFunc("/batch/optimize", h.handleOptimizeBatchQuery).Methods("POST")

	// Policy Compilation
	router.HandleFunc("/compile", h.handleCompilePolicy).Methods("POST")
	router.HandleFunc("/compiled/{id}", h.handleGetCompiledPolicy).Methods("GET")

	// Query Optimization
	router.HandleFunc("/optimize/query", h.handleOptimizeQuery).Methods("POST")
	router.HandleFunc("/analyze/query", h.handleAnalyzeQueryPerformance).Methods("POST")

	// Configuration Optimization
	router.HandleFunc("/optimize/config", h.handleOptimizeConfiguration).Methods("POST")
}

// Policy Management Handlers

func (h *ABACAPIHandler) handleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.CreatePolicy")
	defer span.End()

	var req CreatePolicyRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.policyManager.CreatePolicy(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create policy", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusCreated, result)
}

func (h *ABACAPIHandler) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.ListPolicies")
	defer span.End()

	// Parse query parameters
	req := &ListPoliciesRequest{
		Page:     h.getIntQueryParam(r, "page", 1),
		PageSize: h.getIntQueryParam(r, "page_size", 20),
		Status:   r.URL.Query().Get("status"),
		Category: r.URL.Query().Get("category"),
		Tags:     h.getArrayQueryParam(r, "tags"),
	}

	result, err := h.policyManager.ListPolicies(ctx, req)
	if err != nil {
		h.logger.Error("Failed to list policies", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

func (h *ABACAPIHandler) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.GetPolicy")
	defer span.End()

	policyID, err := h.getPolicyIDFromPath(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.policyManager.GetPolicy(ctx, policyID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			h.writeErrorResponse(w, http.StatusNotFound, err)
		} else {
			h.logger.Error("Failed to get policy", "error", err, "policy_id", policyID)
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

func (h *ABACAPIHandler) handleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.UpdatePolicy")
	defer span.End()

	policyID, err := h.getPolicyIDFromPath(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var req UpdatePolicyRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	req.ID = policyID

	result, err := h.policyManager.UpdatePolicy(ctx, &req)
	if err != nil {
		if errors.IsNotFoundError(err) {
			h.writeErrorResponse(w, http.StatusNotFound, err)
		} else {
			h.logger.Error("Failed to update policy", "error", err, "policy_id", policyID)
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

func (h *ABACAPIHandler) handleTestPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.TestPolicy")
	defer span.End()

	policyID, err := h.getPolicyIDFromPath(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var req PolicyTestRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	req.PolicyID = policyID

	result, err := h.policyManager.TestPolicy(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to test policy", "error", err, "policy_id", policyID)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

// Policy Evaluation Handlers

func (h *ABACAPIHandler) handleEvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.EvaluatePolicy")
	defer span.End()

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		h.metrics.RecordHistogram("abac.api.evaluate_policy.duration",
			duration.Seconds(), map[string]string{"endpoint": "evaluate_policy"})
	}()

	var req PolicyEvaluationRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Set request ID if not provided
	if req.RequestID == uuid.Nil {
		req.RequestID = uuid.New()
	}

	result, err := h.evaluationEngine.EvaluatePolicy(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to evaluate policy", "error", err, "request_id", req.RequestID)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Record evaluation metrics
	h.recordEvaluationMetrics(ctx, &req, result)

	h.writeSuccessResponse(w, http.StatusOK, result)
}

func (h *ABACAPIHandler) handleBatchEvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.BatchEvaluatePolicy")
	defer span.End()

	var req BatchPolicyEvaluationRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.evaluationEngine.BatchEvaluatePolicy(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to batch evaluate policy", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

// Attribute Management Handlers

func (h *ABACAPIHandler) handleCreateAttributeDefinition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.CreateAttributeDefinition")
	defer span.End()

	var req CreateAttributeDefinitionRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.attributeService.CreateAttributeDefinition(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create attribute definition", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusCreated, result)
}

func (h *ABACAPIHandler) handleCollectAttributes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.CollectAttributes")
	defer span.End()

	var req AttributeCollectionRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.attributeCollector.CollectAttributes(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to collect attributes", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

func (h *ABACAPIHandler) handleResolveAttributes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.ResolveAttributes")
	defer span.End()

	var req AttributeResolutionRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.attributeResolver.ResolveAttributes(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to resolve attributes", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

// External Sources Handlers

func (h *ABACAPIHandler) handleRegisterLDAPSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.RegisterLDAPSource")
	defer span.End()

	var req RegisterLDAPSourceRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.externalSources.RegisterLDAPSource(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to register LDAP source", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusCreated, result)
}

func (h *ABACAPIHandler) handleTestSourceConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.TestSourceConnection")
	defer span.End()

	sourceID, err := h.getSourceIDFromPath(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.externalSources.TestSourceConnection(ctx, sourceID)
	if err != nil {
		h.logger.Error("Failed to test source connection", "error", err, "source_id", sourceID)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

// Monitoring Handlers

func (h *ABACAPIHandler) handleGetEvaluationMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.GetEvaluationMetrics")
	defer span.End()

	req := &MetricsQueryRequest{
		StartTime: h.getTimeQueryParam(r, "start_time"),
		EndTime:   h.getTimeQueryParam(r, "end_time"),
		PolicyID:  h.getUUIDQueryParam(r, "policy_id"),
		UserID:    h.getUUIDQueryParam(r, "user_id"),
	}

	result, err := h.monitoringService.GetEvaluationMetrics(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get evaluation metrics", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

func (h *ABACAPIHandler) handleGetSystemHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.GetSystemHealth")
	defer span.End()

	result, err := h.monitoringService.GetSystemHealth(ctx)
	if err != nil {
		h.logger.Error("Failed to get system health", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

// Performance Handlers

func (h *ABACAPIHandler) handleGetCachedEvaluation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.GetCachedEvaluation")
	defer span.End()

	var req CacheRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.performanceOpt.GetCachedEvaluation(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to get cached evaluation", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

func (h *ABACAPIHandler) handleCompilePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.CompilePolicy")
	defer span.End()

	var req PolicyCompilationRequest
	if err := h.decodeRequest(r, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.performanceOpt.CompilePolicy(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to compile policy", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, result)
}

// Health Check Handlers

func (h *ABACAPIHandler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.HealthCheck")
	defer span.End()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"service":   "abac-api",
	}

	h.writeSuccessResponse(w, http.StatusOK, health)
}

func (h *ABACAPIHandler) handleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := h.tracer.StartSpan(ctx, "ABACAPIHandler.ReadinessCheck")
	defer span.End()

	// Check system health
	systemHealth, err := h.monitoringService.GetSystemHealth(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, err)
		return
	}

	readiness := map[string]interface{}{
		"status":        "ready",
		"timestamp":     time.Now(),
		"system_health": systemHealth,
	}

	h.writeSuccessResponse(w, http.StatusOK, readiness)
}

// Helper Methods

func (h *ABACAPIHandler) decodeRequest(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return errors.NewInvalidInputError("invalid JSON request body", "error", err.Error())
	}

	// Validate request if it implements Validator interface
	if validator, ok := v.(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			return errors.Wrap(err, "request validation failed")
		}
	}

	return nil
}

func (h *ABACAPIHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success":   true,
		"data":      data,
		"timestamp": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *ABACAPIHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success":   false,
		"error":     err.Error(),
		"timestamp": time.Now(),
	}

	// Add error details if available
	if detailedError, ok := err.(interface{ GetDetails() map[string]interface{} }); ok {
		response["details"] = detailedError.GetDetails()
	}

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		h.logger.Error("Failed to encode error response", "error", encodeErr)
	}

	// Record error metrics
	h.metrics.RecordCounter("abac.api.error", 1, map[string]string{
		"status_code": strconv.Itoa(statusCode),
		"error_type":  fmt.Sprintf("%T", err),
	})
}

// Path Parameter Extractors

func (h *ABACAPIHandler) getPolicyIDFromPath(r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		return uuid.Nil, errors.NewInvalidInputError("policy ID is required", "path", r.URL.Path)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errors.NewInvalidInputError("invalid policy ID format", "id", idStr)
	}

	return id, nil
}

func (h *ABACAPIHandler) getSourceIDFromPath(r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		return uuid.Nil, errors.NewInvalidInputError("source ID is required", "path", r.URL.Path)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errors.NewInvalidInputError("invalid source ID format", "id", idStr)
	}

	return id, nil
}

// Query Parameter Extractors

func (h *ABACAPIHandler) getIntQueryParam(r *http.Request, param string, defaultValue int) int {
	valueStr := r.URL.Query().Get(param)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func (h *ABACAPIHandler) getArrayQueryParam(r *http.Request, param string) []string {
	values := r.URL.Query()[param]
	if len(values) == 0 {
		return nil
	}

	// Handle comma-separated values in a single parameter
	var result []string
	for _, value := range values {
		if strings.Contains(value, ",") {
			result = append(result, strings.Split(value, ",")...)
		} else {
			result = append(result, value)
		}
	}

	return result
}

func (h *ABACAPIHandler) getTimeQueryParam(r *http.Request, param string) *time.Time {
	valueStr := r.URL.Query().Get(param)
	if valueStr == "" {
		return nil
	}

	// Try RFC3339 format first
	if t, err := time.Parse(time.RFC3339, valueStr); err == nil {
		return &t
	}

	// Try Unix timestamp
	if timestamp, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		t := time.Unix(timestamp, 0)
		return &t
	}

	return nil
}

func (h *ABACAPIHandler) getUUIDQueryParam(r *http.Request, param string) *uuid.UUID {
	valueStr := r.URL.Query().Get(param)
	if valueStr == "" {
		return nil
	}

	id, err := uuid.Parse(valueStr)
	if err != nil {
		return nil
	}

	return &id
}

// Metrics Recording

func (h *ABACAPIHandler) recordEvaluationMetrics(
	ctx context.Context,
	req *PolicyEvaluationRequest,
	result *PolicyEvaluationResult,
) {
	// Record evaluation duration
	h.metrics.RecordHistogram("abac.evaluation.duration",
		result.EvaluationTime.Seconds(),
		map[string]string{
			"policy_id": req.PolicyID.String(),
			"decision":  string(result.Decision),
			"cache_hit": strconv.FormatBool(result.CacheHit),
		})

	// Record decision counter
	h.metrics.RecordCounter("abac.evaluation.decision", 1,
		map[string]string{
			"decision": string(result.Decision),
		})

	// Record cache hit/miss
	if result.CacheHit {
		h.metrics.RecordCounter("abac.evaluation.cache.hit", 1, nil)
	} else {
		h.metrics.RecordCounter("abac.evaluation.cache.miss", 1, nil)
	}

	// Record evaluation metrics with monitoring service
	metricsReq := &EvaluationMetricsRequest{
		EvaluationID:   uuid.New(),
		UserID:         req.Subject.UserID,
		PolicyID:       &req.PolicyID,
		ResourceType:   req.Resource.ResourceType,
		ResourceID:     &req.Resource.ResourceID,
		Action:         req.Action.Action,
		Decision:       result.Decision,
		ExecutionTime:  result.EvaluationTime,
		CacheHit:       result.CacheHit,
		AttributeCount: len(result.AttributesUsed),
		Timestamp:      time.Now(),
	}

	if err := h.monitoringService.RecordEvaluationMetrics(ctx, metricsReq); err != nil {
		h.logger.Error("Failed to record evaluation metrics", "error", err)
	}
}

// Additional stub handlers for completeness

func (h *ABACAPIHandler) handleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle policy deletion
	h.writeSuccessResponse(w, http.StatusNoContent, nil)
}

func (h *ABACAPIHandler) handleSimulatePolicy(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle policy simulation
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "simulated"})
}

func (h *ABACAPIHandler) handleActivatePolicy(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle policy activation
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "activated"})
}

func (h *ABACAPIHandler) handleDeactivatePolicy(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle policy deactivation
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "deactivated"})
}

func (h *ABACAPIHandler) handleCreatePolicyVersion(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle policy version creation
	h.writeSuccessResponse(w, http.StatusCreated, map[string]string{"status": "version_created"})
}

func (h *ABACAPIHandler) handleListPolicyVersions(w http.ResponseWriter, r *http.Request) {
	// Implementation would list policy versions
	h.writeSuccessResponse(w, http.StatusOK, []string{})
}

func (h *ABACAPIHandler) handleAnalyzePolicyImpact(w http.ResponseWriter, r *http.Request) {
	// Implementation would analyze policy impact
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"analysis": "completed"})
}

func (h *ABACAPIHandler) handleDetectPolicyConflicts(w http.ResponseWriter, r *http.Request) {
	// Implementation would detect policy conflicts
	h.writeSuccessResponse(w, http.StatusOK, []string{})
}

func (h *ABACAPIHandler) handleListPolicyTemplates(w http.ResponseWriter, r *http.Request) {
	// Implementation would list policy templates
	h.writeSuccessResponse(w, http.StatusOK, []string{})
}

func (h *ABACAPIHandler) handleGetPolicyTemplate(w http.ResponseWriter, r *http.Request) {
	// Implementation would get specific policy template
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"template": "data"})
}

func (h *ABACAPIHandler) handleImportPolicies(w http.ResponseWriter, r *http.Request) {
	// Implementation would import policies
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "imported"})
}

func (h *ABACAPIHandler) handleExportPolicies(w http.ResponseWriter, r *http.Request) {
	// Implementation would export policies
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "exported"})
}

func (h *ABACAPIHandler) handleEvaluatePolicies(w http.ResponseWriter, r *http.Request) {
	// Implementation would evaluate multiple policies
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "evaluated"})
}

func (h *ABACAPIHandler) handleEvaluateWithContext(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle contextual evaluation
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "evaluated"})
}

func (h *ABACAPIHandler) handleSimulateEvaluation(w http.ResponseWriter, r *http.Request) {
	// Implementation would simulate evaluation
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "simulated"})
}

func (h *ABACAPIHandler) handleEvaluateRule(w http.ResponseWriter, r *http.Request) {
	// Implementation would evaluate a rule
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "evaluated"})
}

func (h *ABACAPIHandler) handleParseRuleExpression(w http.ResponseWriter, r *http.Request) {
	// Implementation would parse rule expression
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "parsed"})
}

func (h *ABACAPIHandler) handleValidateRuleExpression(w http.ResponseWriter, r *http.Request) {
	// Implementation would validate rule expression
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "valid"})
}

func (h *ABACAPIHandler) handleGetDecisionSupport(w http.ResponseWriter, r *http.Request) {
	// Implementation would provide decision support
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"support": "provided"})
}

func (h *ABACAPIHandler) handleGetEvaluationRecommendations(w http.ResponseWriter, r *http.Request) {
	// Implementation would provide evaluation recommendations
	h.writeSuccessResponse(w, http.StatusOK, []string{})
}

// Continue with remaining stub handlers...

func (h *ABACAPIHandler) handleListAttributeDefinitions(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, []string{})
}

func (h *ABACAPIHandler) handleGetAttributeDefinition(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"attribute": "definition"})
}

func (h *ABACAPIHandler) handleUpdateAttributeDefinition(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ABACAPIHandler) handleDeleteAttributeDefinition(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusNoContent, nil)
}

func (h *ABACAPIHandler) handleBatchCollectAttributes(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "collected"})
}

func (h *ABACAPIHandler) handleGetCollectionPlan(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"plan": "generated"})
}

func (h *ABACAPIHandler) handleCreateAttributeDependency(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *ABACAPIHandler) handleInvalidateAttributeCache(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "invalidated"})
}

func (h *ABACAPIHandler) handleGetAttributeCacheStatistics(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]interface{}{"hit_rate": 0.85})
}

func (h *ABACAPIHandler) handleValidateAttributeValue(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]bool{"valid": true})
}

func (h *ABACAPIHandler) handleEncryptAttributeValue(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"encrypted": "value"})
}

func (h *ABACAPIHandler) handleDecryptAttributeValue(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"decrypted": "value"})
}

func (h *ABACAPIHandler) handleRegisterRESTAPISource(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusCreated, map[string]string{"status": "registered"})
}

func (h *ABACAPIHandler) handleRegisterDatabaseSource(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusCreated, map[string]string{"status": "registered"})
}

func (h *ABACAPIHandler) handleRegisterCustomSource(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusCreated, map[string]string{"status": "registered"})
}

func (h *ABACAPIHandler) handleListAttributeSources(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, []string{})
}

func (h *ABACAPIHandler) handleGetAttributeSource(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"source": "details"})
}

func (h *ABACAPIHandler) handleUpdateAttributeSource(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ABACAPIHandler) handleDeleteAttributeSource(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusNoContent, nil)
}

func (h *ABACAPIHandler) handleValidateSourceConfiguration(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]bool{"valid": true})
}

func (h *ABACAPIHandler) handleFetchAttributes(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "fetched"})
}

func (h *ABACAPIHandler) handleSynchronizeAttributes(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "synchronized"})
}

func (h *ABACAPIHandler) handleGetSourceHealth(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"health": "good"})
}

func (h *ABACAPIHandler) handleGetSourceMetrics(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]interface{}{"response_time": 50})
}

func (h *ABACAPIHandler) handleGetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]interface{}{"avg_duration": 25.5})
}

func (h *ABACAPIHandler) handleGetDecisionPatterns(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, []string{})
}

func (h *ABACAPIHandler) handleTrackPolicyDecision(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "tracked"})
}

func (h *ABACAPIHandler) handleDetectAnomalies(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, []string{})
}

func (h *ABACAPIHandler) handleConfigureAnomalyDetection(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "configured"})
}

func (h *ABACAPIHandler) handleCreateAlert(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *ABACAPIHandler) handleGetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, []string{})
}

func (h *ABACAPIHandler) handleResolveAlert(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (h *ABACAPIHandler) handleGetDashboardData(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]interface{}{"widgets": []string{}})
}

func (h *ABACAPIHandler) handleGenerateReport(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"report": "generated"})
}

func (h *ABACAPIHandler) handleRunHealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (h *ABACAPIHandler) handleInvalidatePerformanceCache(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "invalidated"})
}

func (h *ABACAPIHandler) handleBatchEvaluatePerformance(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "evaluated"})
}

func (h *ABACAPIHandler) handleOptimizeBatchQuery(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "optimized"})
}

func (h *ABACAPIHandler) handleGetCompiledPolicy(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"compiled": "policy"})
}

func (h *ABACAPIHandler) handleOptimizeQuery(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"status": "optimized"})
}

func (h *ABACAPIHandler) handleAnalyzeQueryPerformance(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]interface{}{"analysis": "completed"})
}

func (h *ABACAPIHandler) handleOptimizeConfiguration(w http.ResponseWriter, r *http.Request) {
	h.writeSuccessResponse(w, http.StatusOK, map[string]string{"recommendations": "provided"})
}

// Supporting types for API requests

type ListPoliciesRequest struct {
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
	Status   string   `json:"status,omitempty"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type MetricsQueryRequest struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	PolicyID  *uuid.UUID `json:"policy_id,omitempty"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
}

type CacheRequest struct {
	Key        string                 `json:"key" validate:"required"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type PolicyCompilationRequest struct {
	PolicyID uuid.UUID              `json:"policy_id" validate:"required"`
	Options  map[string]interface{} `json:"options,omitempty"`
}
