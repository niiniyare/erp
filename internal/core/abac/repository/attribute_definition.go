package repository

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/convert"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// attributeDefinitionRepository implements AttributeDefinitionRepository interface
type attributeDefinitionRepository struct {
	store   db.Store
	cache   cache.Service
	tracing tracing.Service
	metrics metrics.MetricsProvider
	logger  logger.Logger
}

// NewAttributeDefinitionRepository creates a new attribute definition repository
func NewAttributeDefinitionRepository(
	store db.Store,
	cache cache.Service,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
	logger logger.Logger,
) AttributeDefinitionRepository {
	return &attributeDefinitionRepository{
		store:   store,
		cache:   cache,
		tracing: tracing,
		metrics: metrics,
		logger:  logger,
	}
}

// CreateAttributeDefinition creates a new ABAC attribute definition
func (r *attributeDefinitionRepository) CreateAttributeDefinition(ctx context.Context, req *CreateAttributeDefinitionRequest) (*models.AttributeDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "attributeDefinitionRepository.CreateAttributeDefinition",
		tracing.WithAttributes(
			attribute.String("attribute.name", req.Name),
			attribute.String("attribute.data_type", string(req.DataType)),
			attribute.String("attribute.category", string(req.Category)),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Creating ABAC attribute definition",
		logger.Fields{
			"attribute_name": req.Name,
			"data_type":      req.DataType,
			"category":       req.Category,
			"is_required":    req.IsRequired,
			"is_sensitive":   req.IsSensitive,
		})

	// Validate request
	if err := req.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Attribute definition validation failed")
		r.recordAttributeDefMetrics(ctx, "create", "validation_error", 0)
		return nil, err
	}

	// Convert to SQLC params
	params, err := r.toAttributeDefCreateParams(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert attribute definition params")
		r.recordAttributeDefMetrics(ctx, "create", "conversion_error", 0)
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_CONVERSION_FAILED", "Failed to convert attribute definition parameters").
			WithDetail("error", err.Error())
	}

	startTime := time.Now()

	// Create attribute definition using SQLC (tenant_id handled by current_tenant_id())
	sqlcAttrDef, err := r.store.CreateAttributeDefinition(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create attribute definition in database")
		r.recordAttributeDefMetrics(ctx, "create", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_CREATE_FAILED", "Failed to create attribute definition").
			WithDetail("error", err.Error())
	}

	// Convert back to domain model
	attrDef, err := r.fromSQLCAttributeDefinition(*sqlcAttrDef)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert attribute definition from database")
		return nil, err
	}

	// Invalidate relevant caches
	if err := r.invalidateAttributeDefCaches(ctx, attrDef); err != nil {
		r.logger.WarnContext(ctx, "Failed to invalidate attribute definition caches",
			logger.Fields{"attribute_id": attrDef.ID, "error": err.Error()})
	}

	// Record metrics
	r.recordAttributeDefMetrics(ctx, "create", "success", time.Since(startTime))

	// Log successful creation
	attrDef.LogCreation(ctx, r.logger)

	r.logger.InfoContext(ctx, "ABAC attribute definition created successfully",
		logger.Fields{
			"attribute_id":   attrDef.ID,
			"attribute_name": attrDef.Name,
			"tenant_id":      attrDef.TenantID,
		})

	return attrDef, nil
}

// GetAttributeDefinitionByID retrieves an attribute definition by ID with caching
func (r *attributeDefinitionRepository) GetAttributeDefinitionByID(ctx context.Context, id uuid.UUID) (*models.AttributeDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "attributeDefinitionRepository.GetAttributeDefinitionByID",
		tracing.WithAttributes(attribute.String("attribute.id", id.String())))
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("attribute_def:%s", id)
	var cachedAttrDef models.AttributeDefinition
	if err := r.cache.Get(ctx, cacheKey, &cachedAttrDef); err == nil {
		r.recordAttributeDefMetrics(ctx, "get", "cache_hit", 0)
		return &cachedAttrDef, nil
	}

	startTime := time.Now()

	// Get from database (tenant filtering handled by SQLC query using current_tenant_id())
	sqlcAttrDef, err := r.store.GetAttributeDefinition(ctx, id)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			r.recordAttributeDefMetrics(ctx, "get", "not_found", time.Since(startTime))
			return nil, errors.ErrAttributeNotFound
		}
		span.SetStatus(codes.Error, "Failed to get attribute definition from database")
		r.recordAttributeDefMetrics(ctx, "get", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_GET_FAILED", "Failed to retrieve attribute definition").
			WithDetail("attribute_id", id.String()).
			WithDetail("error", err.Error())
	}

	// Convert to domain model
	attrDef, err := r.fromSQLCAttributeDefinition(*sqlcAttrDef)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Cache the result
	cacheTTL := 15 * time.Minute
	if err := r.cache.Set(ctx, cacheKey, attrDef, cacheTTL); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache attribute definition",
			logger.Fields{"attribute_id": id, "error": err.Error()})
	}

	r.recordAttributeDefMetrics(ctx, "get", "success", time.Since(startTime))
	return attrDef, nil
}

// GetAttributeDefinitionByName retrieves an attribute definition by name
func (r *attributeDefinitionRepository) GetAttributeDefinitionByName(ctx context.Context, name string) (*models.AttributeDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "attributeDefinitionRepository.GetAttributeDefinitionByName",
		tracing.WithAttributes(attribute.String("attribute.name", name)))
	defer span.End()

	startTime := time.Now()

	// Get from database (tenant filtering handled automatically)
	sqlcAttrDef, err := r.store.GetAttributeDefinitionByName(ctx, name)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			r.recordAttributeDefMetrics(ctx, "get_by_name", "not_found", time.Since(startTime))
			return nil, errors.ErrAttributeNotFound
		}
		span.SetStatus(codes.Error, "Failed to get attribute definition by name")
		r.recordAttributeDefMetrics(ctx, "get_by_name", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_GET_FAILED", "Failed to retrieve attribute definition by name").
			WithDetail("attribute_name", name).
			WithDetail("error", err.Error())
	}

	attrDef, err := r.fromSQLCAttributeDefinition(*sqlcAttrDef)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	r.recordAttributeDefMetrics(ctx, "get_by_name", "success", time.Since(startTime))
	return attrDef, nil
}

// UpdateAttributeDefinition updates an existing attribute definition
func (r *attributeDefinitionRepository) UpdateAttributeDefinition(ctx context.Context, id uuid.UUID, req *UpdateAttributeDefinitionRequest) (*models.AttributeDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "attributeDefinitionRepository.UpdateAttributeDefinition",
		tracing.WithAttributes(attribute.String("attribute.id", id.String())))
	defer span.End()

	r.logger.InfoContext(ctx, "Updating ABAC attribute definition",
		logger.Fields{
			"attribute_id": id,
		})

	startTime := time.Now()

	// Convert to SQLC params
	params, err := r.toAttributeDefUpdateParams(id, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert attribute definition update params")
		r.recordAttributeDefMetrics(ctx, "update", "conversion_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_UPDATE_CONVERSION_FAILED", "Failed to convert attribute definition update parameters").
			WithDetail("error", err.Error())
	}

	// Update attribute definition
	sqlcAttrDef, err := r.store.UpdateAttributeDefinition(ctx, *params)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			r.recordAttributeDefMetrics(ctx, "update", "not_found", time.Since(startTime))
			return nil, errors.ErrAttributeNotFound
		}
		span.SetStatus(codes.Error, "Failed to update attribute definition in database")
		r.recordAttributeDefMetrics(ctx, "update", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_UPDATE_FAILED", "Failed to update attribute definition").
			WithDetail("attribute_id", id.String()).
			WithDetail("error", err.Error())
	}

	// Convert back to domain model
	attrDef, err := r.fromSQLCAttributeDefinition(*sqlcAttrDef)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Invalidate caches
	if err := r.invalidateAttributeDefCaches(ctx, attrDef); err != nil {
		r.logger.WarnContext(ctx, "Failed to invalidate attribute definition caches",
			logger.Fields{"attribute_id": attrDef.ID, "error": err.Error()})
	}

	r.recordAttributeDefMetrics(ctx, "update", "success", time.Since(startTime))

	r.logger.InfoContext(ctx, "ABAC attribute definition updated successfully",
		logger.Fields{
			"attribute_id":   attrDef.ID,
			"attribute_name": attrDef.Name,
		})

	return attrDef, nil
}

// DeleteAttributeDefinition soft-deletes an attribute definition
func (r *attributeDefinitionRepository) DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "attributeDefinitionRepository.DeleteAttributeDefinition",
		tracing.WithAttributes(attribute.String("attribute.id", id.String())))
	defer span.End()

	r.logger.InfoContext(ctx, "Deleting ABAC attribute definition", logger.Fields{"attribute_id": id})

	startTime := time.Now()

	// Soft delete attribute definition
	err := r.store.SoftDeleteAttributeDefinition(ctx, id)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			r.recordAttributeDefMetrics(ctx, "delete", "not_found", time.Since(startTime))
			return errors.ErrAttributeNotFound
		}
		span.SetStatus(codes.Error, "Failed to delete attribute definition")
		r.recordAttributeDefMetrics(ctx, "delete", "db_error", time.Since(startTime))
		return errors.NewBusinessError("ATTRIBUTE_DEF_DELETE_FAILED", "Failed to delete attribute definition").
			WithDetail("attribute_id", id.String()).
			WithDetail("error", err.Error())
	}

	// Invalidate all related caches
	cachePatterns := []string{
		fmt.Sprintf("attribute_def:%s", id),
		"attribute_defs:*",
	}

	for _, pattern := range cachePatterns {
		if err := r.cache.DeletePattern(ctx, pattern); err != nil {
			r.logger.WarnContext(ctx, "Failed to invalidate cache pattern",
				logger.Fields{"pattern": pattern, "error": err.Error()})
		}
	}

	r.recordAttributeDefMetrics(ctx, "delete", "success", time.Since(startTime))

	r.logger.InfoContext(ctx, "ABAC attribute definition deleted successfully", logger.Fields{"attribute_id": id})
	return nil
}

// ListAttributeDefinitions lists attribute definitions with filtering
func (r *attributeDefinitionRepository) ListAttributeDefinitions(ctx context.Context, req *ListAttributeDefinitionsRequest) ([]*models.AttributeDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "attributeDefinitionRepository.ListAttributeDefinitions")
	defer span.End()

	startTime := time.Now()

	// Build cache key based on filters
	cacheKey := r.buildListCacheKey(req)
	var cachedAttrDefs []*models.AttributeDefinition
	if err := r.cache.Get(ctx, cacheKey, &cachedAttrDefs); err == nil {
		r.recordAttributeDefMetrics(ctx, "list", "cache_hit", 0)
		return cachedAttrDefs, nil
	}

	// Convert to SQLC params
	params, err := r.toListAttributeDefParams(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert params")
		r.recordAttributeDefMetrics(ctx, "list", "param_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_LIST_FAILED", "Failed to convert request parameters").WithCause(err)
	}

	// Get from database
	sqlcAttrDefs, err := r.store.ListAttributeDefinitions(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list attribute definitions from database")
		r.recordAttributeDefMetrics(ctx, "list", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_LIST_FAILED", "Failed to list attribute definitions").
			WithDetail("error", err.Error())
	}

	// Convert to domain models
	attrDefs := make([]*models.AttributeDefinition, len(sqlcAttrDefs))
	for i, sqlcAttrDef := range sqlcAttrDefs {
		attrDef, err := r.fromSQLCAttributeDefinition(*sqlcAttrDef)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		attrDefs[i] = attrDef
	}

	// Cache the results
	cacheTTL := 10 * time.Minute
	if err := r.cache.Set(ctx, cacheKey, attrDefs, cacheTTL); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache attribute definitions list",
			logger.Fields{"cache_key": cacheKey, "error": err.Error()})
	}

	r.recordAttributeDefMetrics(ctx, "list", "success", time.Since(startTime))

	r.logger.InfoContext(ctx, "Listed attribute definitions",
		logger.Fields{
			"count":        len(attrDefs),
			"list_time_ms": time.Since(startTime).Milliseconds(),
		})

	return attrDefs, nil
}

// GetAttributeDefinitionsByCategory retrieves attribute definitions by category
func (r *attributeDefinitionRepository) GetAttributeDefinitionsByCategory(ctx context.Context, category types.AttributeCategory) ([]*models.AttributeDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "attributeDefinitionRepository.GetAttributeDefinitionsByCategory",
		tracing.WithAttributes(attribute.String("attribute.category", string(category))))
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("attribute_defs:category:%s", category)
	var cachedAttrDefs []*models.AttributeDefinition
	if err := r.cache.Get(ctx, cacheKey, &cachedAttrDefs); err == nil {
		r.recordAttributeDefMetrics(ctx, "get_by_category", "cache_hit", 0)
		return cachedAttrDefs, nil
	}

	startTime := time.Now()

	// Get from database
	sqlcAttrDefs, err := r.store.ListAttributeDefinitionsByCategory(ctx, string(category))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get attribute definitions by category")
		r.recordAttributeDefMetrics(ctx, "get_by_category", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_GET_BY_CATEGORY_FAILED", "Failed to get attribute definitions by category").
			WithDetail("category", string(category)).
			WithDetail("error", err.Error())
	}

	// Convert to domain models
	attrDefs := make([]*models.AttributeDefinition, len(sqlcAttrDefs))
	for i, sqlcAttrDef := range sqlcAttrDefs {
		attrDef, err := r.fromSQLCAttributeDefinition(*sqlcAttrDef)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		attrDefs[i] = attrDef
	}

	// Cache the results
	cacheTTL := 10 * time.Minute
	if err := r.cache.Set(ctx, cacheKey, attrDefs, cacheTTL); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache attribute definitions by category",
			logger.Fields{"cache_key": cacheKey, "error": err.Error()})
	}

	r.recordAttributeDefMetrics(ctx, "get_by_category", "success", time.Since(startTime))
	return attrDefs, nil
}

// GetRequiredAttributeDefinitions retrieves all required attribute definitions
func (r *attributeDefinitionRepository) GetRequiredAttributeDefinitions(ctx context.Context) ([]*models.AttributeDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "attributeDefinitionRepository.GetRequiredAttributeDefinitions")
	defer span.End()

	// Try cache first
	cacheKey := "attribute_defs:required"
	var cachedAttrDefs []*models.AttributeDefinition
	if err := r.cache.Get(ctx, cacheKey, &cachedAttrDefs); err == nil {
		r.recordAttributeDefMetrics(ctx, "get_required", "cache_hit", 0)
		return cachedAttrDefs, nil
	}

	startTime := time.Now()

	// Get from database
	sqlcAttrDefs, err := r.store.GetRequiredAttributeDefinitions(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get required attribute definitions")
		r.recordAttributeDefMetrics(ctx, "get_required", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_GET_REQUIRED_FAILED", "Failed to get required attribute definitions").
			WithDetail("error", err.Error())
	}

	// Convert to domain models
	attrDefs := make([]*models.AttributeDefinition, len(sqlcAttrDefs))
	for i, sqlcAttrDef := range sqlcAttrDefs {
		attrDef, err := r.fromSQLCAttributeDefinition(*sqlcAttrDef)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		attrDefs[i] = attrDef
	}

	// Cache the results
	cacheTTL := 15 * time.Minute // Required attributes change less frequently
	if err := r.cache.Set(ctx, cacheKey, attrDefs, cacheTTL); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache required attribute definitions",
			logger.Fields{"cache_key": cacheKey, "error": err.Error()})
	}

	r.recordAttributeDefMetrics(ctx, "get_required", "success", time.Since(startTime))
	return attrDefs, nil
}

// GetAttributeDefinitionsByIDs retrieves multiple attribute definitions by IDs
func (r *attributeDefinitionRepository) GetAttributeDefinitionsByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.AttributeDefinition, error) {
	ctx, span := r.tracing.StartSpan(ctx, "attributeDefinitionRepository.GetAttributeDefinitionsByIDs",
		tracing.WithAttributes(attribute.Int("attribute.ids_count", len(ids))))
	defer span.End()

	if len(ids) == 0 {
		return []*models.AttributeDefinition{}, nil
	}

	startTime := time.Now()

	// Get from database
	sqlcAttrDefs, err := r.store.GetAttributeDefinitionsByIDs(ctx, ids)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get attribute definitions by IDs")
		r.recordAttributeDefMetrics(ctx, "get_by_ids", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_DEF_GET_BY_IDS_FAILED", "Failed to get attribute definitions by IDs").
			WithDetail("error", err.Error())
	}

	// Convert to domain models
	attrDefs := make([]*models.AttributeDefinition, len(sqlcAttrDefs))
	for i, sqlcAttrDef := range sqlcAttrDefs {
		attrDef, err := r.fromSQLCAttributeDefinition(*sqlcAttrDef)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		attrDefs[i] = attrDef
	}

	r.recordAttributeDefMetrics(ctx, "get_by_ids", "success", time.Since(startTime))
	return attrDefs, nil
}

// ─── HELPER METHODS ─────────────────────────────────────────────

// Validate validates CreateAttributeDefinitionRequest
func (req *CreateAttributeDefinitionRequest) Validate() error {
	if req.Name == "" {
		return errors.NewBusinessError("ATTRIBUTE_NAME_REQUIRED", "Attribute name is required")
	}
	if !req.DataType.IsValid() {
		return errors.NewBusinessError("INVALID_ATTRIBUTE_DATA_TYPE", "Invalid attribute data type")
	}
	if !req.Category.IsValid() {
		return errors.NewBusinessError("INVALID_ATTRIBUTE_CATEGORY", "Invalid attribute category")
	}
	// Validate enum values if data type is ENUM
	if req.DataType == types.AttributeDataTypeEnum && len(req.AllowedValues) == 0 {
		return errors.NewBusinessError("ENUM_VALUES_REQUIRED", "Enum data type requires allowed values")
	}
	return nil
}

// toAttributeDefCreateParams converts CreateAttributeDefinitionRequest to SQLC params
func (r *attributeDefinitionRepository) toAttributeDefCreateParams(req *CreateAttributeDefinitionRequest) (*db.CreateAttributeDefinitionParams, error) {
	// Handle JSON marshaling for complex types
	allowedValuesJSON, err := json.Marshal(req.AllowedValues)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allowed_values: %w", err)
	}

	validationRulesJSON, err := json.Marshal(req.ValidationRules)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal validation_rules: %w", err)
	}

	return &db.CreateAttributeDefinitionParams{
		Name:               req.Name,
		DisplayName:        req.DisplayName,
		Description:        req.Description,
		DataType:           string(req.DataType),
		Category:           string(req.Category),
		IsRequired:         &req.IsRequired,
		IsSensitive:        &req.IsSensitive,
		DefaultValue:       req.DefaultValue,
		AllowedValues:      allowedValuesJSON,
		ValidationRules:    validationRulesJSON,
		EncryptionRequired: &req.EncryptionRequired,
		IsActive:           &req.IsActive,
		// tenant_id is set automatically by current_tenant_id() in SQLC query
	}, nil
}

// toAttributeDefUpdateParams converts UpdateAttributeDefinitionRequest to SQLC params
func (r *attributeDefinitionRepository) toAttributeDefUpdateParams(id uuid.UUID, req *UpdateAttributeDefinitionRequest) (*db.UpdateAttributeDefinitionParams, error) {
	params := db.UpdateAttributeDefinitionParams{
		ID: id,
	}

	if req.DisplayName != nil {
		params.DisplayName = req.DisplayName
	}
	if req.Description != nil {
		params.Description = req.Description
	}
	if req.IsRequired != nil {
		params.IsRequired = req.IsRequired
	}
	if req.IsSensitive != nil {
		params.IsSensitive = req.IsSensitive
	}
	if req.DefaultValue != nil {
		params.DefaultValue = req.DefaultValue
	}
	if req.AllowedValues != nil {
		allowedValuesJSON, err := json.Marshal(req.AllowedValues)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal allowed_values: %w", err)
		}
		params.AllowedValues = allowedValuesJSON
	}
	if req.ValidationRules != nil {
		validationRulesJSON, err := json.Marshal(req.ValidationRules)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal validation_rules: %w", err)
		}
		params.ValidationRules = validationRulesJSON
	}
	if req.EncryptionRequired != nil {
		params.EncryptionRequired = req.EncryptionRequired
	}
	if req.IsActive != nil {
		params.IsActive = req.IsActive
	}

	return &params, nil
}

// toListAttributeDefParams converts ListAttributeDefinitionsRequest to SQLC params
func (r *attributeDefinitionRepository) toListAttributeDefParams(req *ListAttributeDefinitionsRequest) (*db.ListAttributeDefinitionsParams, error) {
	limit, err := convert.IntToInt32(req.Limit)
	if err != nil {
		return nil, fmt.Errorf("invalid limit value: %w", err)
	}
	offset, err := convert.IntToInt32(req.Offset)
	if err != nil {
		return nil, fmt.Errorf("invalid offset value: %w", err)
	}

	params := &db.ListAttributeDefinitionsParams{
		Limit:  limit,
		Offset: offset,
	}

	if req.Category != nil {
		params.Category = string(*req.Category)
	}
	if req.IsActive != nil {
		params.IsActive = *req.IsActive
	}

	return params, nil
}

// fromSQLCAttributeDefinition converts SQLC attribute definition to domain model
func (r *attributeDefinitionRepository) fromSQLCAttributeDefinition(sqlcAttrDef db.AttributeDefinition) (*models.AttributeDefinition, error) {
	var allowedValues []string
	if len(sqlcAttrDef.AllowedValues) > 0 {
		if err := json.Unmarshal(sqlcAttrDef.AllowedValues, &allowedValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed_values: %w", err)
		}
	}

	var validationRules map[string]any
	if len(sqlcAttrDef.ValidationRules) > 0 {
		if err := json.Unmarshal(sqlcAttrDef.ValidationRules, &validationRules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal validation_rules: %w", err)
		}
	}

	return &models.AttributeDefinition{
		ID:                 sqlcAttrDef.ID,
		TenantID:           sqlcAttrDef.TenantID,
		Name:               sqlcAttrDef.Name,
		DisplayName:        sqlcAttrDef.DisplayName,
		Description:        sqlcAttrDef.Description,
		DataType:           types.AttributeDataType(sqlcAttrDef.DataType),
		Category:           types.AttributeCategory(sqlcAttrDef.Category),
		IsRequired:         *sqlcAttrDef.IsRequired,
		IsSensitive:        *sqlcAttrDef.IsSensitive,
		DefaultValue:       sqlcAttrDef.DefaultValue,
		AllowedValues:      allowedValues,
		ValidationRules:    validationRules,
		EncryptionRequired: *sqlcAttrDef.EncryptionRequired,
		IsActive:           *sqlcAttrDef.IsActive,
		CreatedAt:          sqlcAttrDef.CreatedAt.Time,
	}, nil
}

// buildListCacheKey builds a cache key for list operations
func (r *attributeDefinitionRepository) buildListCacheKey(req *ListAttributeDefinitionsRequest) string {
	key := "attribute_defs:list"

	if req.Category != nil {
		key += fmt.Sprintf(":category:%s", *req.Category)
	}
	if req.DataType != nil {
		key += fmt.Sprintf(":datatype:%s", *req.DataType)
	}
	if req.IsRequired != nil {
		key += fmt.Sprintf(":required:%t", *req.IsRequired)
	}
	if req.IsSensitive != nil {
		key += fmt.Sprintf(":sensitive:%t", *req.IsSensitive)
	}
	if req.IsActive != nil {
		key += fmt.Sprintf(":active:%t", *req.IsActive)
	}
	key += fmt.Sprintf(":limit:%d:offset:%d", req.Limit, req.Offset)

	return key
}

// invalidateAttributeDefCaches invalidates all caches related to an attribute definition
func (r *attributeDefinitionRepository) invalidateAttributeDefCaches(ctx context.Context, attrDef *models.AttributeDefinition) error {
	cachePatterns := []string{
		fmt.Sprintf("attribute_def:%s", attrDef.ID),
		"attribute_defs:*",
		fmt.Sprintf("attribute_defs:category:%s", attrDef.Category),
		"attribute_defs:required",
	}

	for _, pattern := range cachePatterns {
		if err := r.cache.DeletePattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to delete cache pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// recordAttributeDefMetrics records attribute definition operation metrics
func (r *attributeDefinitionRepository) recordAttributeDefMetrics(ctx context.Context, operation, status string, duration time.Duration) {
	counter := r.metrics.Counter(
		"abac_attribute_definition_repository_operations_total",
		"Total number of attribute definition repository operations",
		"operation", "status",
	)

	counter.Inc(metrics.Fields{
		"operation": operation,
		"status":    status,
	})

	if duration > 0 {
		histogram := r.metrics.Histogram(
			"abac_attribute_definition_repository_operation_duration_seconds",
			"Duration of attribute definition repository operations",
			metrics.StandardHTTPDurationBuckets(),
			"operation", "status",
		)

		histogram.Observe(duration.Seconds(), metrics.Fields{
			"operation": operation,
			"status":    status,
		})
	}
}

// derefString dereferences a *string, returning "" if nil
func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// ptrBool returns a pointer to a bool
func ptrBool(b bool) *bool {
	return &b
}

// nullString converts a *string to sql.NullString
func nullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{}
}

// nullBool converts a *bool to sql.NullBool
func nullBool(b *bool) sql.NullBool {
	if b != nil {
		return sql.NullBool{Bool: *b, Valid: true}
	}
	return sql.NullBool{}
}
