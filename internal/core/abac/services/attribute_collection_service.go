package services

//go:generate go run go.uber.org/mock/mockgen -source=attribute_collection_service.go -destination=mock.go -package=services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// AttributeCollectionService defines the interface for collecting attributes
type AttributeCollectionService interface {
	// Core collection methods
	CollectUserAttributes(ctx context.Context, userID uuid.UUID) (map[string]*models.AttributeValue, error)
	CollectResourceAttributes(ctx context.Context, resourceType string, resourceID *uuid.UUID) (map[string]*models.AttributeValue, error)
	CollectEnvironmentAttributes(ctx context.Context, req *EnvironmentAttributeRequest) (map[string]*models.AttributeValue, error)
	CollectSessionAttributes(ctx context.Context, sessionData map[string]any) (map[string]*models.AttributeValue, error)
	CollectEntityAttributes(ctx context.Context, entityID uuid.UUID) (map[string]*models.AttributeValue, error)
	CollectActionAttributes(ctx context.Context, action string) (map[string]*models.AttributeValue, error)

	//  collection
	CollectAllAttributes(ctx context.Context, req *AttributeCollectionRequest) (*models.AttributeContext, error)

	// Validation and enrichment
	ValidateAttributes(ctx context.Context, attributes map[string]*models.AttributeValue) error
	EnrichAttributeContext(ctx context.Context, context *models.AttributeContext) (*models.AttributeContext, error)
}

// AttributeCollectionRequest represents a request to collect all attributes
type AttributeCollectionRequest struct {
	UserID          uuid.UUID      `json:"user_id" validate:"required"`
	ResourceType    string         `json:"resource_type" validate:"required"`
	ResourceID      *uuid.UUID     `json:"resource_id,omitempty"`
	Action          string         `json:"action" validate:"required"`
	EntityID        *uuid.UUID     `json:"entity_id,omitempty"`
	SessionData     map[string]any `json:"session_data,omitempty"`
	EnvironmentData map[string]any `json:"environment_data,omitempty"`
	RequestID       string         `json:"request_id"`
	CollectExpired  bool           `json:"collect_expired"` // Whether to include expired attributes
}

// EnvironmentAttributeRequest represents environment attribute collection request
type EnvironmentAttributeRequest struct {
	IPAddress   string         `json:"ip_address,omitempty"`
	UserAgent   string         `json:"user_agent,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
	Location    map[string]any `json:"location,omitempty"`
	DeviceInfo  map[string]any `json:"device_info,omitempty"`
	NetworkInfo map[string]any `json:"network_info,omitempty"`
}

// attributeCollectionService implements AttributeCollectionService
type attributeCollectionService struct {
	attrDefRepo  repository.AttributeDefinitionRepository
	attrRepo     repository.AttributeRepository
	identityRepo identity.Repository
	cache        cache.Service
	tracing      tracing.TracingService
	metrics      metrics.MetricsProvider
	logger       logger.Logger
}

// NewAttributeCollectionService creates a new attribute collection service
func NewAttributeCollectionService(
	attrDefRepo repository.AttributeDefinitionRepository,
	attrRepo repository.AttributeRepository,
	identityRepo identity.Repository,
	cache cache.Service,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
	logger logger.Logger,
) AttributeCollectionService {
	return &attributeCollectionService{
		attrDefRepo:  attrDefRepo,
		attrRepo:     attrRepo,
		identityRepo: identityRepo,
		cache:        cache,
		tracing:      tracing,
		metrics:      metrics,
		logger:       logger,
	}
}

// CollectUserAttributes collects user-related attributes
func (s *attributeCollectionService) CollectUserAttributes(ctx context.Context, userID uuid.UUID) (map[string]*models.AttributeValue, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCollectionService.CollectUserAttributes",
		tracing.WithAttributes(attribute.String("user.id", userID.String())))
	defer span.End()

	s.logger.InfoContext(ctx, "Collecting user attributes", logger.Fields{"user_id": userID})
	startTime := time.Now()

	// Try cache first
	cacheKey := fmt.Sprintf("user_attributes:%s", userID)
	var cachedAttributes map[string]*models.AttributeValue
	if err := s.cache.Get(ctx, cacheKey, &cachedAttributes); err == nil {
		s.recordCollectionMetrics(ctx, "user", "cache_hit", len(cachedAttributes), time.Since(startTime))
		return cachedAttributes, nil
	}

	// Get user attribute definitions
	userAttrDefs, err := s.attrDefRepo.GetAttributeDefinitionsByCategory(ctx, types.AttributeCategoryUser)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get user attribute definitions")
		s.recordCollectionMetrics(ctx, "user", "error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("USER_ATTR_DEFS_FAILED", "Failed to get user attribute definitions").
			WithDetail("user_id", userID.String()).
			WithDetail("error", err.Error())
	}

	// Get user from identity service
	user, err := s.identityRepo.GetUserByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get user from identity service")
		s.recordCollectionMetrics(ctx, "user", "error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("USER_FETCH_FAILED", "Failed to fetch user").
			WithDetail("user_id", userID.String()).
			WithDetail("error", err.Error())
	}

	attributes := make(map[string]*models.AttributeValue)

	// Collect basic user attributes
	if err := s.collectBasicUserAttributes(ctx, user, userAttrDefs, attributes); err != nil {
		span.RecordError(err)
		s.recordCollectionMetrics(ctx, "user", "error", 0, time.Since(startTime))
		return nil, err
	}

	// Collect person attributes if available
	if user.PersonID != nil {
		if err := s.collectPersonAttributes(ctx, *user.PersonID, userAttrDefs, attributes); err != nil {
			s.logger.WarnContext(ctx, "Failed to collect person attributes",
				logger.Fields{"user_id": userID, "person_id": *user.PersonID, "error": err.Error()})
		}
	}

	// Collect employee attributes if available
	if user.EmployeeID != nil {
		if err := s.collectEmployeeAttributes(ctx, *user.EmployeeID, userAttrDefs, attributes); err != nil {
			s.logger.WarnContext(ctx, "Failed to collect employee attributes",
				logger.Fields{"user_id": userID, "employee_id": *user.EmployeeID, "error": err.Error()})
		}
	}

	// Cache the results
	cacheTTL := 10 * time.Minute
	if err := s.cache.Set(ctx, cacheKey, attributes, cacheTTL); err != nil {
		s.logger.WarnContext(ctx, "Failed to cache user attributes",
			logger.Fields{"user_id": userID, "error": err.Error()})
	}

	s.recordCollectionMetrics(ctx, "user", "success", len(attributes), time.Since(startTime))
	s.logger.InfoContext(ctx, "User attributes collected successfully",
		logger.Fields{
			"user_id":         userID,
			"attribute_count": len(attributes),
			"collection_time": time.Since(startTime).Milliseconds(),
		})

	return attributes, nil
}

// CollectResourceAttributes collects resource-related attributes
func (s *attributeCollectionService) CollectResourceAttributes(ctx context.Context, resourceType string, resourceID *uuid.UUID) (map[string]*models.AttributeValue, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCollectionService.CollectResourceAttributes",
		tracing.WithAttributes(
			attribute.String("resource.type", resourceType),
			attribute.String("resource.id", func() string {
				if resourceID != nil {
					return resourceID.String()
				}
				return "nil"
			}()),
		))
	defer span.End()

	s.logger.InfoContext(ctx, "Collecting resource attributes",
		logger.Fields{"resource_type": resourceType, "resource_id": resourceID})
	startTime := time.Now()

	// Try cache first
	var cacheKey string
	if resourceID != nil {
		cacheKey = fmt.Sprintf("resource_attributes:%s:%s", resourceType, *resourceID)
	} else {
		cacheKey = fmt.Sprintf("resource_attributes:%s:global", resourceType)
	}

	var cachedAttributes map[string]*models.AttributeValue
	if err := s.cache.Get(ctx, cacheKey, &cachedAttributes); err == nil {
		s.recordCollectionMetrics(ctx, "resource", "cache_hit", len(cachedAttributes), time.Since(startTime))
		return cachedAttributes, nil
	}

	// Get resource attribute definitions
	resourceAttrDefs, err := s.attrDefRepo.GetAttributeDefinitionsByCategory(ctx, types.AttributeCategoryResource)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get resource attribute definitions")
		s.recordCollectionMetrics(ctx, "resource", "error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("RESOURCE_ATTR_DEFS_FAILED", "Failed to get resource attribute definitions").
			WithDetail("resource_type", resourceType).
			WithDetail("error", err.Error())
	}

	attributes := make(map[string]*models.AttributeValue)

	// Collect basic resource attributes
	s.collectBasicResourceAttributes(ctx, resourceType, resourceID, resourceAttrDefs, attributes)

	// If specific resource ID provided, collect instance-specific attributes
	if resourceID != nil {
		if err := s.collectSpecificResourceAttributes(ctx, resourceType, *resourceID, resourceAttrDefs, attributes); err != nil {
			s.logger.WarnContext(ctx, "Failed to collect specific resource attributes",
				logger.Fields{"resource_type": resourceType, "resource_id": *resourceID, "error": err.Error()})
		}
	}

	// Cache the results
	cacheTTL := 15 * time.Minute // Resource attributes are more static
	if err := s.cache.Set(ctx, cacheKey, attributes, cacheTTL); err != nil {
		s.logger.WarnContext(ctx, "Failed to cache resource attributes",
			logger.Fields{"cache_key": cacheKey, "error": err.Error()})
	}

	s.recordCollectionMetrics(ctx, "resource", "success", len(attributes), time.Since(startTime))
	s.logger.InfoContext(ctx, "Resource attributes collected successfully",
		logger.Fields{
			"resource_type":   resourceType,
			"resource_id":     resourceID,
			"attribute_count": len(attributes),
			"collection_time": time.Since(startTime).Milliseconds(),
		})

	return attributes, nil
}

// CollectEnvironmentAttributes collects environment-related attributes
func (s *attributeCollectionService) CollectEnvironmentAttributes(ctx context.Context, req *EnvironmentAttributeRequest) (map[string]*models.AttributeValue, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCollectionService.CollectEnvironmentAttributes")
	defer span.End()

	s.logger.InfoContext(ctx, "Collecting environment attributes", logger.Fields{"request": req})
	startTime := time.Now()

	// Get environment attribute definitions
	envAttrDefs, err := s.attrDefRepo.GetAttributeDefinitionsByCategory(ctx, types.AttributeCategoryEnvironment)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get environment attribute definitions")
		s.recordCollectionMetrics(ctx, "environment", "error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("ENV_ATTR_DEFS_FAILED", "Failed to get environment attribute definitions").
			WithDetail("error", err.Error())
	}

	attributes := make(map[string]*models.AttributeValue)

	// Collect basic environment attributes
	now := time.Now()
	if req.Timestamp.IsZero() {
		req.Timestamp = now
	}

	// Time-based attributes
	s.addAttributeIfDefined(attributes, envAttrDefs, "timestamp", req.Timestamp, types.AttributeDataTypeDate, types.AttributeSourceEnvironment)
	s.addAttributeIfDefined(attributes, envAttrDefs, "hour_of_day", req.Timestamp.Hour(), types.AttributeDataTypeNumber, types.AttributeSourceEnvironment)
	s.addAttributeIfDefined(attributes, envAttrDefs, "day_of_week", req.Timestamp.Weekday().String(), types.AttributeDataTypeString, types.AttributeSourceEnvironment)
	s.addAttributeIfDefined(attributes, envAttrDefs, "is_weekend", req.Timestamp.Weekday() == time.Saturday || req.Timestamp.Weekday() == time.Sunday, types.AttributeDataTypeBoolean, types.AttributeSourceEnvironment)

	// Network attributes
	if req.IPAddress != "" {
		s.addAttributeIfDefined(attributes, envAttrDefs, "ip_address", req.IPAddress, types.AttributeDataTypeString, types.AttributeSourceEnvironment)
		s.addAttributeIfDefined(attributes, envAttrDefs, "is_internal_ip", s.isInternalIP(req.IPAddress), types.AttributeDataTypeBoolean, types.AttributeSourceComputed)
	}

	if req.UserAgent != "" {
		s.addAttributeIfDefined(attributes, envAttrDefs, "user_agent", req.UserAgent, types.AttributeDataTypeString, types.AttributeSourceEnvironment)
	}

	// Location attributes
	if req.Location != nil {
		for key, value := range req.Location {
			attrName := fmt.Sprintf("location_%s", key)
			s.addAttributeIfDefined(attributes, envAttrDefs, attrName, value, types.AttributeDataTypeString, types.AttributeSourceEnvironment)
		}
	}

	// Device attributes
	if req.DeviceInfo != nil {
		for key, value := range req.DeviceInfo {
			attrName := fmt.Sprintf("device_%s", key)
			s.addAttributeIfDefined(attributes, envAttrDefs, attrName, value, types.AttributeDataTypeString, types.AttributeSourceEnvironment)
		}
	}

	// Network info attributes
	if req.NetworkInfo != nil {
		for key, value := range req.NetworkInfo {
			attrName := fmt.Sprintf("network_%s", key)
			s.addAttributeIfDefined(attributes, envAttrDefs, attrName, value, types.AttributeDataTypeString, types.AttributeSourceEnvironment)
		}
	}

	s.recordCollectionMetrics(ctx, "environment", "success", len(attributes), time.Since(startTime))
	s.logger.InfoContext(ctx, "Environment attributes collected successfully",
		logger.Fields{
			"attribute_count": len(attributes),
			"collection_time": time.Since(startTime).Milliseconds(),
		})

	return attributes, nil
}

// CollectSessionAttributes collects session-related attributes
func (s *attributeCollectionService) CollectSessionAttributes(ctx context.Context, sessionData map[string]any) (map[string]*models.AttributeValue, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCollectionService.CollectSessionAttributes")
	defer span.End()

	s.logger.InfoContext(ctx, "Collecting session attributes", logger.Fields{"session_data_keys": len(sessionData)})
	startTime := time.Now()

	// Get session attribute definitions
	sessionAttrDefs, err := s.attrDefRepo.GetAttributeDefinitionsByCategory(ctx, types.AttributeCategorySession)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get session attribute definitions")
		s.recordCollectionMetrics(ctx, "session", "error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("SESSION_ATTR_DEFS_FAILED", "Failed to get session attribute definitions").
			WithDetail("error", err.Error())
	}

	attributes := make(map[string]*models.AttributeValue)

	// Process session data
	for key, value := range sessionData {
		// Normalize key name
		normalizedKey := s.normalizeAttributeName(key)

		// Determine data type
		dataType := s.inferDataType(value)

		s.addAttributeIfDefined(attributes, sessionAttrDefs, normalizedKey, value, dataType, types.AttributeSourceSession)
	}

	s.recordCollectionMetrics(ctx, "session", "success", len(attributes), time.Since(startTime))
	s.logger.InfoContext(ctx, "Session attributes collected successfully",
		logger.Fields{
			"attribute_count": len(attributes),
			"collection_time": time.Since(startTime).Milliseconds(),
		})

	return attributes, nil
}

// CollectEntityAttributes collects entity-related attributes
func (s *attributeCollectionService) CollectEntityAttributes(ctx context.Context, entityID uuid.UUID) (map[string]*models.AttributeValue, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCollectionService.CollectEntityAttributes",
		tracing.WithAttributes(attribute.String("entity.id", entityID.String())))
	defer span.End()

	s.logger.InfoContext(ctx, "Collecting entity attributes", logger.Fields{"entity_id": entityID})
	startTime := time.Now()

	// Try cache first
	cacheKey := fmt.Sprintf("entity_attributes:%s", entityID)
	var cachedAttributes map[string]*models.AttributeValue
	if err := s.cache.Get(ctx, cacheKey, &cachedAttributes); err == nil {
		s.recordCollectionMetrics(ctx, "entity", "cache_hit", len(cachedAttributes), time.Since(startTime))
		return cachedAttributes, nil
	}

	// Get entity attribute definitions
	entityAttrDefs, err := s.attrDefRepo.GetAttributeDefinitionsByCategory(ctx, types.AttributeCategoryEntity)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get entity attribute definitions")
		s.recordCollectionMetrics(ctx, "entity", "error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("ENTITY_ATTR_DEFS_FAILED", "Failed to get entity attribute definitions").
			WithDetail("entity_id", entityID.String()).
			WithDetail("error", err.Error())
	}

	attributes := make(map[string]*models.AttributeValue)

	// TODO: Implement entity attribute collection based on your entity service
	// For now, we'll add basic entity attributes
	s.addAttributeIfDefined(attributes, entityAttrDefs, "entity_id", entityID, types.AttributeDataTypeString, types.AttributeSourceIdentity)

	// Cache the results
	cacheTTL := 20 * time.Minute // Entity attributes change less frequently
	if err := s.cache.Set(ctx, cacheKey, attributes, cacheTTL); err != nil {
		s.logger.WarnContext(ctx, "Failed to cache entity attributes",
			logger.Fields{"entity_id": entityID, "error": err.Error()})
	}

	s.recordCollectionMetrics(ctx, "entity", "success", len(attributes), time.Since(startTime))
	s.logger.InfoContext(ctx, "Entity attributes collected successfully",
		logger.Fields{
			"entity_id":       entityID,
			"attribute_count": len(attributes),
			"collection_time": time.Since(startTime).Milliseconds(),
		})

	return attributes, nil
}

// CollectActionAttributes collects action-related attributes
func (s *attributeCollectionService) CollectActionAttributes(ctx context.Context, action string) (map[string]*models.AttributeValue, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCollectionService.CollectActionAttributes",
		tracing.WithAttributes(attribute.String("action", action)))
	defer span.End()

	s.logger.InfoContext(ctx, "Collecting action attributes", logger.Fields{"action": action})
	startTime := time.Now()

	// Try cache first
	cacheKey := fmt.Sprintf("action_attributes:%s", action)
	var cachedAttributes map[string]*models.AttributeValue
	if err := s.cache.Get(ctx, cacheKey, &cachedAttributes); err == nil {
		s.recordCollectionMetrics(ctx, "action", "cache_hit", len(cachedAttributes), time.Since(startTime))
		return cachedAttributes, nil
	}

	// Get action attribute definitions
	actionAttrDefs, err := s.attrDefRepo.GetAttributeDefinitionsByCategory(ctx, types.AttributeCategoryAction)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get action attribute definitions")
		s.recordCollectionMetrics(ctx, "action", "error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("ACTION_ATTR_DEFS_FAILED", "Failed to get action attribute definitions").
			WithDetail("action", action).
			WithDetail("error", err.Error())
	}

	attributes := make(map[string]*models.AttributeValue)

	// Collect basic action attributes
	s.addAttributeIfDefined(attributes, actionAttrDefs, "action", action, types.AttributeDataTypeString, types.AttributeSourceSession)
	s.addAttributeIfDefined(attributes, actionAttrDefs, "action_type", s.categorizeAction(action), types.AttributeDataTypeString, types.AttributeSourceComputed)
	s.addAttributeIfDefined(attributes, actionAttrDefs, "is_read_action", s.isReadAction(action), types.AttributeDataTypeBoolean, types.AttributeSourceComputed)
	s.addAttributeIfDefined(attributes, actionAttrDefs, "is_write_action", s.isWriteAction(action), types.AttributeDataTypeBoolean, types.AttributeSourceComputed)
	s.addAttributeIfDefined(attributes, actionAttrDefs, "is_delete_action", s.isDeleteAction(action), types.AttributeDataTypeBoolean, types.AttributeSourceComputed)

	// Cache the results - actions are static
	cacheTTL := 30 * time.Minute
	if err := s.cache.Set(ctx, cacheKey, attributes, cacheTTL); err != nil {
		s.logger.WarnContext(ctx, "Failed to cache action attributes",
			logger.Fields{"action": action, "error": err.Error()})
	}

	s.recordCollectionMetrics(ctx, "action", "success", len(attributes), time.Since(startTime))
	s.logger.InfoContext(ctx, "Action attributes collected successfully",
		logger.Fields{
			"action":          action,
			"attribute_count": len(attributes),
			"collection_time": time.Since(startTime).Milliseconds(),
		})

	return attributes, nil
}

// CollectAllAttributes collects all attributes for a evaluation context
func (s *attributeCollectionService) CollectAllAttributes(ctx context.Context, req *AttributeCollectionRequest) (*models.AttributeContext, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCollectionService.CollectAllAttributes",
		tracing.WithAttributes(
			attribute.String("user.id", req.UserID.String()),
			attribute.String("resource.type", req.ResourceType),
			attribute.String("action", req.Action),
		))
	defer span.End()

	s.logger.InfoContext(ctx, "Collecting all attributes for evaluation context",
		logger.Fields{
			"user_id":       req.UserID,
			"resource_type": req.ResourceType,
			"resource_id":   req.ResourceID,
			"action":        req.Action,
			"entity_id":     req.EntityID,
			"request_id":    req.RequestID,
		})

	startTime := time.Now()
	tenantID := req.UserID // TODO: Extract proper tenant ID from context

	// Create attribute context
	attrContext := models.NewAttributeContext(tenantID)

	// Collect user attributes
	userAttrs, err := s.CollectUserAttributes(ctx, req.UserID)
	if err != nil {
		span.RecordError(err)
		s.recordCollectionMetrics(ctx, "all", "user_error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("USER_ATTR_COLLECTION_FAILED", "Failed to collect user attributes").
			WithDetail("user_id", req.UserID.String()).
			WithDetail("error", err.Error())
	}
	attrContext.UserAttributes = userAttrs

	// Collect resource attributes
	resourceAttrs, err := s.CollectResourceAttributes(ctx, req.ResourceType, req.ResourceID)
	if err != nil {
		span.RecordError(err)
		s.recordCollectionMetrics(ctx, "all", "resource_error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("RESOURCE_ATTR_COLLECTION_FAILED", "Failed to collect resource attributes").
			WithDetail("resource_type", req.ResourceType).
			WithDetail("error", err.Error())
	}
	attrContext.ResourceAttributes = resourceAttrs

	// Collect environment attributes
	envReq := &EnvironmentAttributeRequest{
		Timestamp: time.Now(),
	}
	if req.EnvironmentData != nil {
		// Extract known environment fields
		if ipAddr, ok := req.EnvironmentData["ip_address"].(string); ok {
			envReq.IPAddress = ipAddr
		}
		if userAgent, ok := req.EnvironmentData["user_agent"].(string); ok {
			envReq.UserAgent = userAgent
		}
		if location, ok := req.EnvironmentData["location"].(map[string]any); ok {
			envReq.Location = location
		}
		if deviceInfo, ok := req.EnvironmentData["device_info"].(map[string]any); ok {
			envReq.DeviceInfo = deviceInfo
		}
	}

	envAttrs, err := s.CollectEnvironmentAttributes(ctx, envReq)
	if err != nil {
		span.RecordError(err)
		s.recordCollectionMetrics(ctx, "all", "environment_error", 0, time.Since(startTime))
		return nil, errors.NewBusinessError("ENV_ATTR_COLLECTION_FAILED", "Failed to collect environment attributes").
			WithDetail("error", err.Error())
	}
	attrContext.EnvironmentAttributes = envAttrs

	// Collect session attributes
	if req.SessionData != nil {
		sessionAttrs, err := s.CollectSessionAttributes(ctx, req.SessionData)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to collect session attributes",
				logger.Fields{"error": err.Error()})
		} else {
			attrContext.SessionAttributes = sessionAttrs
		}
	}

	// Collect entity attributes if provided
	if req.EntityID != nil {
		entityAttrs, err := s.CollectEntityAttributes(ctx, *req.EntityID)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to collect entity attributes",
				logger.Fields{"entity_id": *req.EntityID, "error": err.Error()})
		} else {
			attrContext.EntityAttributes = entityAttrs
		}
	}

	// Collect action attributes
	actionAttrs, err := s.CollectActionAttributes(ctx, req.Action)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to collect action attributes",
			logger.Fields{"action": req.Action, "error": err.Error()})
	} else {
		attrContext.ActionAttributes = actionAttrs
	}

	// Validate collected attributes
	allAttrs := attrContext.GetAllAttributes()
	if err := s.ValidateAttributes(ctx, allAttrs); err != nil {
		span.RecordError(err)
		s.recordCollectionMetrics(ctx, "all", "validation_error", len(allAttrs), time.Since(startTime))
		return nil, errors.NewBusinessError("ATTR_VALIDATION_FAILED", "Attribute validation failed").
			WithDetail("error", err.Error())
	}

	// Remove expired attributes if not requested
	if !req.CollectExpired {
		s.removeExpiredAttributes(attrContext)
	}

	// Log collection results
	attrContext.LogCollection(ctx, s.logger)

	s.recordCollectionMetrics(ctx, "all", "success", len(allAttrs), time.Since(startTime))
	s.logger.InfoContext(ctx, "All attributes collected successfully",
		logger.Fields{
			"user_id":             req.UserID,
			"resource_type":       req.ResourceType,
			"action":              req.Action,
			"total_attributes":    len(allAttrs),
			"user_attributes":     len(attrContext.UserAttributes),
			"resource_attributes": len(attrContext.ResourceAttributes),
			"env_attributes":      len(attrContext.EnvironmentAttributes),
			"session_attributes":  len(attrContext.SessionAttributes),
			"entity_attributes":   len(attrContext.EntityAttributes),
			"action_attributes":   len(attrContext.ActionAttributes),
			"collection_time":     time.Since(startTime).Milliseconds(),
		})

	return attrContext, nil
}

// ValidateAttributes validates collected attributes against their definitions
func (s *attributeCollectionService) ValidateAttributes(ctx context.Context, attributes map[string]*models.AttributeValue) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCollectionService.ValidateAttributes")
	defer span.End()

	for name, attrValue := range attributes {
		// Get attribute definition
		attrDef, err := s.attrDefRepo.GetAttributeDefinitionByName(ctx, name)
		if err != nil {
			// If attribute definition not found, skip validation but log warning
			if errors.IsBusinessErrorCode(err, "ATTRIBUTE_NOT_FOUND") {
				s.logger.WarnContext(ctx, "Attribute definition not found for collected attribute",
					logger.Fields{"attribute_name": name})
				continue
			}
			return err
		}

		// Validate attribute value against definition
		if err := s.validateAttributeValue(attrValue, attrDef); err != nil {
			return errors.NewBusinessError("ATTRIBUTE_VALIDATION_FAILED", "Attribute validation failed").
				WithDetail("attribute_name", name).
				WithDetail("error", err.Error())
		}
	}

	return nil
}

// EnrichAttributeContext enriches attribute context with computed attributes
func (s *attributeCollectionService) EnrichAttributeContext(ctx context.Context, attrContext *models.AttributeContext) (*models.AttributeContext, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCollectionService.EnrichAttributeContext")
	defer span.End()

	s.logger.InfoContext(ctx, "Enriching attribute context")

	// Add computed attributes based on existing attributes
	s.addComputedAttributes(attrContext)

	// Add derived attributes
	s.addDerivedAttributes(attrContext)

	return attrContext, nil
}

// ─── HELPER METHODS ─────────────────────────────────────────────

// collectBasicUserAttributes collects basic user attributes
func (s *attributeCollectionService) collectBasicUserAttributes(
	ctx context.Context,
	user *identity.User,
	attrDefs []*models.AttributeDefinition,
	attributes map[string]*models.AttributeValue,
) error {
	// Basic user attributes
	s.addAttributeIfDefined(attributes, attrDefs, "user_id", user.ID, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "username", user.Username, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "email", user.Email, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "user_type", user.UserType, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "account_status", user.AccountStatus, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "is_active", user.IsActive, types.AttributeDataTypeBoolean, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "mfa_enabled", user.MfaEnabled, types.AttributeDataTypeBoolean, types.AttributeSourceIdentity)

	// Add user-specific attributes
	if user.UserAttributes != nil {
		for key, value := range user.UserAttributes {
			normalizedKey := s.normalizeAttributeName(fmt.Sprintf("user_%s", key))
			dataType := s.inferDataType(value)
			s.addAttributeIfDefined(attributes, attrDefs, normalizedKey, value, dataType, types.AttributeSourceIdentity)
		}
	}

	return nil
}

// collectPersonAttributes collects person-related attributes
func (s *attributeCollectionService) collectPersonAttributes(
	ctx context.Context,
	personID uuid.UUID,
	attrDefs []*models.AttributeDefinition,
	attributes map[string]*models.AttributeValue,
) error {
	person, err := s.identityRepo.GetPersonByID(ctx, personID)
	if err != nil {
		return err
	}

	// Basic person attributes
	s.addAttributeIfDefined(attributes, attrDefs, "person_id", person.ID, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "person_type", person.PersonType, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "first_name", person.FirstName, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "last_name", person.LastName, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "full_name", person.GetFullName(), types.AttributeDataTypeString, types.AttributeSourceComputed)

	if person.Email != "" {
		s.addAttributeIfDefined(attributes, attrDefs, "person_email", person.Email, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	}

	// Add person security attributes
	if person.SecurityAttributes != nil {
		for key, value := range person.SecurityAttributes {
			normalizedKey := s.normalizeAttributeName(fmt.Sprintf("person_%s", key))
			dataType := s.inferDataType(value)
			s.addAttributeIfDefined(attributes, attrDefs, normalizedKey, value, dataType, types.AttributeSourceIdentity)
		}
	}

	return nil
}

// collectEmployeeAttributes collects employee-related attributes
func (s *attributeCollectionService) collectEmployeeAttributes(
	ctx context.Context,
	employeeID uuid.UUID,
	attrDefs []*models.AttributeDefinition,
	attributes map[string]*models.AttributeValue,
) error {
	employee, err := s.identityRepo.GetEmployeeByID(ctx, employeeID)
	if err != nil {
		return err
	}

	// Basic employee attributes
	s.addAttributeIfDefined(attributes, attrDefs, "employee_id", employee.ID, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "employee_number", employee.EmployeeNumber, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "employment_status", employee.Status, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	s.addAttributeIfDefined(attributes, attrDefs, "security_level", employee.SecurityLevel, types.AttributeDataTypeNumber, types.AttributeSourceIdentity)

	if employee.PositionTitle != nil {
		s.addAttributeIfDefined(attributes, attrDefs, "position_title", *employee.PositionTitle, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	}

	if employee.DepartmentID != nil {
		s.addAttributeIfDefined(attributes, attrDefs, "department_id", *employee.DepartmentID, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	}

	if employee.ManagerID != nil {
		s.addAttributeIfDefined(attributes, attrDefs, "manager_id", *employee.ManagerID, types.AttributeDataTypeString, types.AttributeSourceIdentity)
	}

	// Add employee access attributes
	if employee.AccessAttributes != nil {
		for key, value := range employee.AccessAttributes {
			normalizedKey := s.normalizeAttributeName(fmt.Sprintf("employee_%s", key))
			dataType := s.inferDataType(value)
			s.addAttributeIfDefined(attributes, attrDefs, normalizedKey, value, dataType, types.AttributeSourceIdentity)
		}
	}

	return nil
}

// collectBasicResourceAttributes collects basic resource attributes
func (s *attributeCollectionService) collectBasicResourceAttributes(
	ctx context.Context,
	resourceType string,
	resourceID *uuid.UUID,
	attrDefs []*models.AttributeDefinition,
	attributes map[string]*models.AttributeValue,
) {
	s.addAttributeIfDefined(attributes, attrDefs, "resource_type", resourceType, types.AttributeDataTypeString, types.AttributeSourceResource)

	if resourceID != nil {
		s.addAttributeIfDefined(attributes, attrDefs, "resource_id", *resourceID, types.AttributeDataTypeString, types.AttributeSourceResource)
	}

	// Add resource type categorization
	s.addAttributeIfDefined(attributes, attrDefs, "resource_category", s.categorizeResourceType(resourceType), types.AttributeDataTypeString, types.AttributeSourceComputed)
}

// collectSpecificResourceAttributes collects attributes for a specific resource instance
func (s *attributeCollectionService) collectSpecificResourceAttributes(
	ctx context.Context,
	resourceType string,
	resourceID uuid.UUID,
	attrDefs []*models.AttributeDefinition,
	attributes map[string]*models.AttributeValue,
) error {
	// TODO: Implement based on your resource service
	// This would typically involve calling a resource service to get resource details
	s.logger.InfoContext(ctx, "Collecting specific resource attributes",
		logger.Fields{"resource_type": resourceType, "resource_id": resourceID})

	return nil
}

// addAttributeIfDefined adds an attribute to the collection if it's defined in the attribute definitions
func (s *attributeCollectionService) addAttributeIfDefined(
	attributes map[string]*models.AttributeValue,
	attrDefs []*models.AttributeDefinition,
	name string,
	value any,
	dataType types.AttributeDataType,
	source types.AttributeSource,
) {
	// Check if attribute is defined
	var attrDef *models.AttributeDefinition
	for _, def := range attrDefs {
		if def.Name == name {
			attrDef = def
			break
		}
	}

	if attrDef == nil {
		// Attribute not defined, skip
		return
	}

	// Create attribute value
	attrValue := models.NewAttributeValue(
		attrDef.ID,
		name,
		value,
		dataType,
		attrDef.Category,
		source,
	)

	// Set sensitivity flag
	attrValue.IsSensitive = attrDef.IsSensitive

	attributes[name] = attrValue
}

// validateAttributeValue validates an attribute value against its definition
func (s *attributeCollectionService) validateAttributeValue(attrValue *models.AttributeValue, attrDef *models.AttributeDefinition) error {
	// Check data type consistency
	if attrValue.DataType != attrDef.DataType {
		return fmt.Errorf("data type mismatch: expected %s, got %s", attrDef.DataType, attrValue.DataType)
	}

	// Check allowed values for enum types
	if attrDef.DataType == types.AttributeDataTypeEnum {
		if len(attrDef.AllowedValues) > 0 {
			strValue := fmt.Sprintf("%v", attrValue.Value)
			for _, allowed := range attrDef.AllowedValues {
				if allowed == strValue {
					return nil // Valid enum value
				}
			}
			return fmt.Errorf("invalid enum value: %v, allowed values: %v", strValue, attrDef.AllowedValues)
		}
	}

	// TODO: Add more validation rules based on ValidationRules field

	return nil
}

// normalizeAttributeName normalizes attribute names to follow naming conventions
func (s *attributeCollectionService) normalizeAttributeName(name string) string {
	// Convert to lowercase and replace spaces/special chars with underscores
	normalized := strings.ToLower(name)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}

// inferDataType infers the data type from the value
func (s *attributeCollectionService) inferDataType(value any) types.AttributeDataType {
	switch value.(type) {
	case bool:
		return types.AttributeDataTypeBoolean
	case int, int32, int64, float32, float64:
		return types.AttributeDataTypeNumber
	case time.Time:
		return types.AttributeDataTypeDate
	case []any:
		return types.AttributeDataTypeArray
	case map[string]any:
		return types.AttributeDataTypeJSON
	default:
		return types.AttributeDataTypeString
	}
}

// isInternalIP checks if an IP address is internal/private
func (s *attributeCollectionService) isInternalIP(ip string) bool {
	// Simple check for common private IP ranges
	return strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "172.16.") ||
		ip == "127.0.0.1" ||
		ip == "localhost"
}

// categorizeAction categorizes actions into types
func (s *attributeCollectionService) categorizeAction(action string) string {
	action = strings.ToLower(action)
	switch {
	case strings.Contains(action, "read") || strings.Contains(action, "get") || strings.Contains(action, "view") || strings.Contains(action, "list"):
		return "read"
	case strings.Contains(action, "write") || strings.Contains(action, "create") || strings.Contains(action, "update") || strings.Contains(action, "post") || strings.Contains(action, "put"):
		return "write"
	case strings.Contains(action, "delete") || strings.Contains(action, "remove"):
		return "delete"
	case strings.Contains(action, "execute") || strings.Contains(action, "run"):
		return "execute"
	default:
		return "other"
	}
}

// isReadAction checks if action is a read operation
func (s *attributeCollectionService) isReadAction(action string) bool {
	return s.categorizeAction(action) == "read"
}

// isWriteAction checks if action is a write operation
func (s *attributeCollectionService) isWriteAction(action string) bool {
	return s.categorizeAction(action) == "write"
}

// isDeleteAction checks if action is a delete operation
func (s *attributeCollectionService) isDeleteAction(action string) bool {
	return s.categorizeAction(action) == "delete"
}

// categorizeResourceType categorizes resource types
func (s *attributeCollectionService) categorizeResourceType(resourceType string) string {
	resourceType = strings.ToLower(resourceType)
	switch {
	case strings.Contains(resourceType, "user") || strings.Contains(resourceType, "person") || strings.Contains(resourceType, "employee"):
		return "identity"
	case strings.Contains(resourceType, "document") || strings.Contains(resourceType, "file"):
		return "document"
	case strings.Contains(resourceType, "system") || strings.Contains(resourceType, "service"):
		return "system"
	case strings.Contains(resourceType, "data"):
		return "data"
	default:
		return "general"
	}
}

// addComputedAttributes adds computed attributes based on existing attributes
func (s *attributeCollectionService) addComputedAttributes(attrContext *models.AttributeContext) {
	// Add computed attributes based on existing data

	// Compute user risk score based on various factors
	if userRiskScore := s.computeUserRiskScore(attrContext); userRiskScore >= 0 {
		attrContext.UserAttributes["computed_risk_score"] = &models.AttributeValue{
			Name:      "computed_risk_score",
			Value:     userRiskScore,
			DataType:  types.AttributeDataTypeNumber,
			Category:  types.AttributeCategoryUser,
			Source:    types.AttributeSourceComputed,
			Timestamp: time.Now(),
		}
	}

	// Compute access pattern based on time and location
	if accessPattern := s.computeAccessPattern(attrContext); accessPattern != "" {
		attrContext.EnvironmentAttributes["computed_access_pattern"] = &models.AttributeValue{
			Name:      "computed_access_pattern",
			Value:     accessPattern,
			DataType:  types.AttributeDataTypeString,
			Category:  types.AttributeCategoryEnvironment,
			Source:    types.AttributeSourceComputed,
			Timestamp: time.Now(),
		}
	}
}

// addDerivedAttributes adds derived attributes
func (s *attributeCollectionService) addDerivedAttributes(attrContext *models.AttributeContext) {
	// Add derived attributes based on business logic

	// TODO: Implement based on your business requirements
}

// computeUserRiskScore computes a risk score for the user
func (s *attributeCollectionService) computeUserRiskScore(attrContext *models.AttributeContext) float64 {
	score := 0.0

	// TODO: Implement risk scoring algorithm based on your requirements
	// Example factors:
	// - Failed login attempts
	// - Account age
	// - Previous access patterns
	// - Security level

	return score
}

// computeAccessPattern computes access pattern based on environment
func (s *attributeCollectionService) computeAccessPattern(attrContext *models.AttributeContext) string {
	// TODO: Implement access pattern analysis
	// Example patterns:
	// - "normal_hours" vs "after_hours"
	// - "usual_location" vs "unusual_location"
	// - "typical_device" vs "new_device"

	return "normal"
}

// removeExpiredAttributes removes expired attributes from context
func (s *attributeCollectionService) removeExpiredAttributes(attrContext *models.AttributeContext) {
	s.removeExpiredFromMap(attrContext.UserAttributes)
	s.removeExpiredFromMap(attrContext.ResourceAttributes)
	s.removeExpiredFromMap(attrContext.EnvironmentAttributes)
	s.removeExpiredFromMap(attrContext.SessionAttributes)
	s.removeExpiredFromMap(attrContext.EntityAttributes)
	s.removeExpiredFromMap(attrContext.ActionAttributes)
}

// removeExpiredFromMap removes expired attributes from a map
func (s *attributeCollectionService) removeExpiredFromMap(attributes map[string]*models.AttributeValue) {
	for name, attr := range attributes {
		if attr.IsExpired() {
			delete(attributes, name)
		}
	}
}

// recordCollectionMetrics records attribute collection metrics
func (s *attributeCollectionService) recordCollectionMetrics(ctx context.Context, category, status string, count int, duration time.Duration) {
	// Collection counter
	counter := s.metrics.Counter(
		"abac_attribute_collection_total",
		"Total number of attribute collections",
		"category", "status",
	)

	counter.Inc(metrics.Fields{
		"category": category,
		"status":   status,
	})

	// Collection duration histogram
	histogram := s.metrics.Histogram(
		"abac_attribute_collection_duration_seconds",
		"Time spent collecting attributes",
		metrics.StandardHTTPDurationBuckets(),
		"category", "status",
	)

	histogram.Observe(duration.Seconds(), metrics.Fields{
		"category": category,
		"status":   status,
	})

	// Attribute count histogram
	if count > 0 {
		countHist := s.metrics.Histogram(
			"abac_collected_attributes_count",
			"Number of attributes collected per operation",
			[]float64{0, 1, 5, 10, 20, 50, 100, 200},
			"category", "status",
		)

		countHist.Observe(float64(count), metrics.Fields{
			"category": category,
			"status":   status,
		})
	}
}
