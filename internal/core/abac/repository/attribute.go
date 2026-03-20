package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	db "awo/db/sqlc"
	"awo/internal/core/abac/models"
	"awo/internal/platform/cache"
	"awo/internal/shared/convert"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"awo/internal/shared/types"
)

// attributeRepository implements AttributeRepository using SQLC and Store
type attributeRepository struct {
	store   db.Store
	cache   cache.Service
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewAttributeRepository creates a new attribute repository implementation
func NewAttributeRepository(
	store db.Store,
	cache cache.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *attributeRepository {
	return &attributeRepository{
		store:   store,
		cache:   cache,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

func (r attributeRepository) BuildAttributeContext(ctx context.Context, req *BuildAttributeContextRequest) (*models.AttributeContext, error) {
	return nil, nil
}

// CreateAttributeDefinition creates a new attribute definition
func (r *attributeRepository) CreateAttributeDefinition(ctx context.Context, req *CreateAttributeDefinitionRequest) (*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.CreateAttributeDefinition",
		tracing.WithAttributes(
			attribute.String("name", req.Name),
			attribute.String("data_type", string(req.DataType)),
			attribute.String("category", string(req.Category)),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Creating attribute definition",
		logger.Fields{
			"name":      req.Name,
			"data_type": req.DataType,
			"category":  req.Category,
		})

	allowedValuesJSON, err := json.Marshal(req.AllowedValues)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allowed_values: %w", err)
	}

	validationRulesJSON, err := json.Marshal(req.ValidationRules)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal validation_rules: %w", err)
	}

	params := db.CreateAttributeDefinitionParams{
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
	}

	attrDef, err := r.store.CreateAttributeDefinition(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_create_definition_failed", "Total failed attribute definition creations", "error_type").Inc(metrics.Fields{"error_type": "conversion_error"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_CREATE_FAILED", "Failed to create attribute definition").WithDetail("error", err.Error())
	}

	r.metrics.Counter("attribute_repository_create_definition_success", "Total successful attribute definition creations").Inc(nil)

	result, err := r.convertSQLCAttributeDefinitionToModel(attrDef)
	if err != nil {
		return nil, err
	}

	r.logger.InfoContext(ctx, "Attribute definition created successfully",
		logger.Fields{
			"definition_id": result.ID,
			"name":          result.Name,
		})

	return result, nil
}

// GetAttributeDefinitionByID retrieves an attribute definition by ID
func (r *attributeRepository) GetAttributeDefinitionByID(ctx context.Context, id uuid.UUID) (*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeDefinitionByID",
		tracing.WithAttributes(
			attribute.String("definition_id", id.String()),
		))
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("attr_def:%s", id.String())
	var attrDef *models.AttributeDefinition
	if err := r.cache.Get(ctx, cacheKey, &attrDef); err == nil {
		r.metrics.Counter("attribute_repository_definition_cache_hit", "Total attribute definition cache hits").Inc(nil)
		return attrDef, nil
	}

	// Cache miss - get from database
	sqlcAttrDef, err := r.store.GetAttributeDefinition(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.Counter("attribute_repository_definition_not_found", "Total attribute definition not found errors").Inc(nil)
			return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_NOT_FOUND", "Attribute definition not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_get_definition_failed", "Total failed attribute definition retrievals").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_GET_FAILED", "Failed to get attribute definition").WithDetail("error", err.Error())
	}

	// Convert to domain model
	attrDef, err = r.convertSQLCAttributeDefinitionToModel(sqlcAttrDef)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, attrDef, 30*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache attribute definition", logger.Fields{"error": err.Error()})
	}

	r.metrics.Counter("attribute_repository_get_definition_success", "Total successful attribute definition retrievals").Inc(nil)
	return attrDef, nil
}

// GetAttributeDefinitionByName retrieves an attribute definition by name
func (r *attributeRepository) GetAttributeDefinitionByName(ctx context.Context, name string) (*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeDefinitionByName",
		tracing.WithAttributes(
			attribute.String("name", name),
		))
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("attr_def_name:%s", name)
	var attrDef *models.AttributeDefinition
	if err := r.cache.Get(ctx, cacheKey, &attrDef); err == nil {
		r.metrics.Counter("attribute_repository_definition_name_cache_hit", "Total attribute definition name cache hits").Inc(nil)
		return attrDef, nil
	}

	// Cache miss - get from database
	sqlcAttrDef, err := r.store.GetAttributeDefinitionByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.Counter("attribute_repository_definition_not_found", "Total attribute definition not found errors").Inc(nil)
			return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_NOT_FOUND", "Attribute definition not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_get_definition_by_name_failed", "Total failed attribute definition retrievals by name").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_GET_FAILED", "Failed to get attribute definition by name").WithDetail("error", err.Error())
	}

	// Convert to domain model
	attrDef, err = r.convertSQLCAttributeDefinitionToModel(sqlcAttrDef)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, attrDef, 30*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache attribute definition by name", logger.Fields{"error": err.Error()})
	}

	r.metrics.Counter("attribute_repository_get_definition_by_name_success", "Total successful attribute definition retrievals by name").Inc(nil)
	return attrDef, nil
}

// UpdateAttributeDefinition updates an existing attribute definition
func (r *attributeRepository) UpdateAttributeDefinition(ctx context.Context, req *UpdateAttributeDefinitionRequest) (*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.UpdateAttributeDefinition",
		tracing.WithAttributes(
			attribute.String("display_name", *req.DisplayName),
		))
	defer span.End()

	allowedValuesJSON, err := json.Marshal(req.AllowedValues)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allowed_values: %w", err)
	}

	validationRulesJSON, err := json.Marshal(req.ValidationRules)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal validation_rules: %w", err)
	}

	params := db.UpdateAttributeDefinitionParams{
		// ID:                 req.ID,
		// Name:               *req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		// DataType:           nullString(types.StringPtr(string(req.DataType))),
		// Category:           nullString(types.StringPtr(string(req.Category))),
		IsRequired:         req.IsRequired,
		IsSensitive:        req.IsSensitive,
		DefaultValue:       req.DefaultValue,
		AllowedValues:      allowedValuesJSON,
		ValidationRules:    validationRulesJSON,
		EncryptionRequired: req.EncryptionRequired,
		IsActive:           req.IsActive,
	}

	attrDef, err := r.store.UpdateAttributeDefinition(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_update_definition_failed", "Total failed attribute definition updates").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_UPDATE_FAILED", "Failed to update attribute definition").WithDetail("error", err.Error())
	}

	r.metrics.Counter("attribute_repository_update_definition_success", "Total successful attribute definition updates").Inc(nil)
	return r.convertSQLCAttributeDefinitionToModel(attrDef)
}

// DeleteAttributeDefinition deletes an attribute definition
func (r *attributeRepository) DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.DeleteAttributeDefinition",
		tracing.WithAttributes(
			attribute.String("definition_id", id.String()),
		))
	defer span.End()

	err := r.store.DeleteAttributeDefinition(ctx, id)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_delete_definition_failed", "Total failed attribute definition deletions").Inc(nil)
		return errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_DELETE_FAILED", "Failed to delete attribute definition").WithDetail("error", err.Error())
	}

	r.metrics.Counter("attribute_repository_delete_definition_success", "Total successful attribute definition deletions").Inc(nil)

	r.logger.InfoContext(ctx, "Attribute definition deleted successfully",
		logger.Fields{"definition_id": id})

	return nil
}

// ListAttributeDefinitions lists attribute definitions with filtering and pagination
func (r *attributeRepository) ListAttributeDefinitions(ctx context.Context, req *ListAttributeDefinitionsRequest) ([]*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.ListAttributeDefinitions")
	defer span.End()

	limit, err := convert.IntToInt32(req.Limit)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid limit value: %w", err)
	}
	offset, err := convert.IntToInt32(req.Offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid offset value: %w", err)
	}

	params := db.ListAttributeDefinitionsParams{
		Limit:  limit,
		Offset: offset,
	}

	if req.Search != "" {
		params.Search = req.Search
	}

	if req.Category != nil {
		s := string(*req.Category)
		params.Category = s
	}

	if req.IsActive != nil {
		params.IsActive = *req.IsActive
	}

	// Try cache first for list
	cacheKey := fmt.Sprintf("attr_defs:search:%s:cat:%s:active:%v:limit:%d:offset:%d",
		req.Search, *req.Category, *req.IsActive, req.Limit, req.Offset)
	var attrDefs []*db.AttributeDefinition
	if err := r.cache.Get(ctx, cacheKey, &attrDefs); err == nil {
		r.metrics.Counter("attribute_repository_list_cache_hit", "Total attribute definition list cache hits").Inc(nil)
		// Convert cached results
		result := make([]*models.AttributeDefinition, len(attrDefs))
		for i, attrDef := range attrDefs {
			result[i], err = r.convertSQLCAttributeDefinitionToModel(attrDef)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	}

	// Cache miss - get from database
	sqlcAttrDefs, err := r.store.ListAttributeDefinitions(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_list_definitions_failed", "Total failed attribute definition list retrievals").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITIONS_LIST_FAILED", "Failed to list attribute definitions").WithDetail("error", err.Error())
	}

	r.metrics.Counter("attribute_repository_list_definitions_success", "Total successful attribute definition list retrievals").Inc(nil)

	result := make([]*models.AttributeDefinition, len(sqlcAttrDefs))
	for i, attrDef := range sqlcAttrDefs {
		result[i], err = r.convertSQLCAttributeDefinitionToModel(attrDef)
		if err != nil {
			return nil, err
		}
	}

	// Cache the results
	if err := r.cache.Set(ctx, cacheKey, sqlcAttrDefs, 15*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache attribute definitions list", logger.Fields{"error": err.Error()})
	}

	return result, nil
}

// CreateAttributeValue creates a new attribute value
func (r *attributeRepository) CreateAttributeValue(ctx context.Context, req *CreateAttributeValueRequest) (*models.AttributeValue, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.CreateAttributeValue",
		tracing.WithAttributes(
			attribute.String("definition_id", req.DefinitionID.String()),
			attribute.String("entity_id", req.EntityID.String()),
		))
	defer span.End()

	var efective_from sql.NullTime

	if !req.EffectiveFrom.IsZero() {
		efective_from = sql.NullTime{
			Time:  req.EffectiveFrom,
			Valid: true,
		}
	}
	var efective_to sql.NullTime

	if !req.EffectiveTo.IsZero() {
		efective_to = sql.NullTime{
			Time:  *req.EffectiveTo,
			Valid: true,
		}
	}

	params := db.CreateAttributeValueParams{
		DefinitionID:   req.DefinitionID,
		EntityID:       req.EntityID,
		Value:          req.Value,
		EncryptedValue: req.EncryptedValue,
		IsEncrypted:    &req.IsEncrypted,
		Version:        &req.Version,
		EffectiveFrom:  efective_from,
		EffectiveTo:    efective_to,
		CreatedBy:      &req.CreatedBy,
	}

	attrValue, err := r.store.CreateAttributeValue(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_create_value_failed", "Total failed attribute value creations").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_VALUE_CREATE_FAILED", "Failed to create attribute value").WithDetail("error", err.Error())
	}

	r.metrics.Counter("attribute_repository_create_value_success", "Total successful attribute value creations").Inc(nil)
	return r.convertSQLCAttributeValueToModel(attrValue)
}

// GetAttributeValuesByEntity retrieves attribute values by entity ID
func (r *attributeRepository) GetAttributeValuesByEntity(ctx context.Context, entityID uuid.UUID, category types.AttributeCategory) ([]*models.AttributeValue, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeValuesByEntity",
		tracing.WithAttributes(
			attribute.String("entity_id", entityID.String()),
			attribute.String("category", string(category)),
		))
	defer span.End()

	var categoryStr *string
	if category != "" {
		s := string(category)
		categoryStr = &s
	}

	params := db.GetAttributeValuesByEntityParams{
		EntityID: entityID,
		Category: *categoryStr,
	}

	// Try cache first
	cacheKey := fmt.Sprintf("entity_attrs:%s:%s", entityID.String(), string(category))
	var rows []*db.GetAttributeValuesByEntityRow
	if err := r.cache.Get(ctx, cacheKey, &rows); err == nil {
		r.metrics.Counter("attribute_repository_entity_values_cache_hit", "Total attribute values by entity cache hits").Inc(nil)
		// Convert cached results
		result := make([]*models.AttributeValue, len(rows))
		for i, row := range rows {
			result[i], err = r.convertSQLCAttributeValueRowToModel(row)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	}

	// Cache miss - get from database
	rows, err := r.store.GetAttributeValuesByEntity(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_get_values_by_entity_failed", "Total failed attribute values by entity retrievals").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_VALUES_GET_FAILED", "Failed to get attribute values by entity").WithDetail("error", err.Error())
	}

	r.metrics.Counter("attribute_repository_get_values_by_entity_success", "Total successful attribute values by entity retrievals").Inc(nil)

	result := make([]*models.AttributeValue, len(rows))
	for i, row := range rows {
		result[i], err = r.convertSQLCAttributeValueRowToModel(row)
		if err != nil {
			return nil, err
		}
	}

	// Cache the results
	if err := r.cache.Set(ctx, cacheKey, rows, 20*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache entity attribute values", logger.Fields{"error": err.Error()})
	}

	return result, nil
}

// GetAttributeValue retrieves a single attribute value by definition ID and entity ID
func (r *attributeRepository) GetAttributeValue(ctx context.Context, definitionID uuid.UUID, entityID uuid.UUID) (*models.AttributeValue, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeValue",
		tracing.WithAttributes(
			attribute.String("definition_id", definitionID.String()),
			attribute.String("entity_id", entityID.String()),
		))
	defer span.End()

	// For now, implement a simple version that searches through all categories
	// TODO: Optimize this with a direct database query
	categories := types.AllAttributeCategories()

	for _, category := range categories {
		values, err := r.GetAttributeValuesByEntity(ctx, entityID, category)
		if err != nil {
			continue // Skip errors for categories that might not exist
		}

		// Find the specific value by definition ID
		for _, value := range values {
			if value.DefinitionID == definitionID {
				return value, nil
			}
		}
	}

	// Not found
	return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_VALUE_NOT_FOUND", "Attribute value not found").
		WithDetail("definition_id", definitionID.String()).
		WithDetail("entity_id", entityID.String())
}

// GetAttributeValuesByDefinitions retrieves attribute values for multiple definitions and an entity
func (r *attributeRepository) GetAttributeValuesByDefinitions(ctx context.Context, definitionIDs []uuid.UUID, entityID uuid.UUID) ([]*models.AttributeValue, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeValuesByDefinitions",
		tracing.WithAttributes(
			attribute.Int("definition_count", len(definitionIDs)),
			attribute.String("entity_id", entityID.String()),
		))
	defer span.End()

	if len(definitionIDs) == 0 {
		return []*models.AttributeValue{}, nil
	}

	var result []*models.AttributeValue

	// For now, implement by calling GetAttributeValue for each definition
	// TODO: Optimize with a bulk database query
	for _, definitionID := range definitionIDs {
		value, err := r.GetAttributeValue(ctx, definitionID, entityID)
		if err != nil {
			// Skip values that are not found, but log other errors
			if !errors.IsBusinessErrorCode(err, "ATTRIBUTE_VALUE_NOT_FOUND") {
				r.logger.WarnContext(ctx, "Error getting attribute value", logger.Fields{
					"definition_id": definitionID.String(),
					"entity_id":     entityID.String(),
					"error":         err.Error(),
				})
			}
			continue
		}
		result = append(result, value)
	}

	return result, nil
}

// GetResourceAttributeContext retrieves attribute context for a resource
func (r *attributeRepository) GetResourceAttributeContext(ctx context.Context, resourceType string, resourceID uuid.UUID) (*models.AttributeContext, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetResourceAttributeContext",
		tracing.WithAttributes(
			attribute.String("resource_type", resourceType),
			attribute.String("resource_id", resourceID.String()),
		))
	defer span.End()

	// For now, return a simple mock implementation
	// TODO: Implement proper resource attribute context retrieval
	return &models.AttributeContext{
		UserAttributes:        make(map[string]*models.AttributeValue),
		ResourceAttributes:    make(map[string]*models.AttributeValue),
		EnvironmentAttributes: make(map[string]*models.AttributeValue),
		SessionAttributes:     make(map[string]*models.AttributeValue),
		EntityAttributes:      make(map[string]*models.AttributeValue),
		ActionAttributes:      make(map[string]*models.AttributeValue),
		CollectedAt:           time.Now(),
		TenantID:              uuid.New(), // TODO: Get actual tenant ID from context
	}, nil
}

// GetUserAttributeContext retrieves attribute context for a user
func (r *attributeRepository) GetUserAttributeContext(ctx context.Context, userID uuid.UUID) (*models.AttributeContext, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetUserAttributeContext",
		tracing.WithAttributes(
			attribute.String("user_id", userID.String()),
		))
	defer span.End()

	// For now, return a simple mock implementation
	// TODO: Implement proper user attribute context retrieval
	return &models.AttributeContext{
		UserAttributes:        make(map[string]*models.AttributeValue),
		ResourceAttributes:    make(map[string]*models.AttributeValue),
		EnvironmentAttributes: make(map[string]*models.AttributeValue),
		SessionAttributes:     make(map[string]*models.AttributeValue),
		EntityAttributes:      make(map[string]*models.AttributeValue),
		ActionAttributes:      make(map[string]*models.AttributeValue),
		CollectedAt:           time.Now(),
		TenantID:              uuid.New(), // TODO: Get actual tenant ID from context
	}, nil
}

// StoreAttributeValues stores multiple attribute values in bulk
func (r *attributeRepository) StoreAttributeValues(ctx context.Context, values []*StoreAttributeValueRequest) ([]*models.AttributeValue, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.StoreAttributeValues",
		tracing.WithAttributes(
			attribute.Int("value_count", len(values)),
		))
	defer span.End()

	if len(values) == 0 {
		return []*models.AttributeValue{}, nil
	}

	var result []*models.AttributeValue

	// For now, implement by calling CreateAttributeValue for each value
	// TODO: Implement proper bulk insert
	for _, value := range values {
		// Convert value.Value from any to string for CreateAttributeValueRequest
		valueStr := ""
		if value.Value != nil {
			valueStr = fmt.Sprintf("%v", value.Value)
		}

		// Use current time for effective dates if not provided
		effectiveFrom := time.Now()

		created, err := r.CreateAttributeValue(ctx, &CreateAttributeValueRequest{
			DefinitionID:  value.DefinitionID,
			EntityID:      value.EntityID,
			Value:         valueStr,
			IsEncrypted:   false,
			Version:       1,
			EffectiveFrom: effectiveFrom,
			EffectiveTo:   value.ExpiresAt,
			CreatedBy:     uuid.New(), // TODO: Get actual user ID from context
		})
		if err != nil {
			r.logger.WarnContext(ctx, "Error storing attribute value", logger.Fields{
				"definition_id": value.DefinitionID.String(),
				"entity_id":     value.EntityID.String(),
				"error":         err.Error(),
			})
			continue
		}
		result = append(result, created)
	}

	return result, nil
}

// UpdateAttributeValue updates an existing attribute value
func (r *attributeRepository) UpdateAttributeValue(ctx context.Context, definitionID, entityID uuid.UUID, value any) (*models.AttributeValue, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.UpdateAttributeValue",
		tracing.WithAttributes(
			attribute.String("definition_id", definitionID.String()),
			attribute.String("entity_id", entityID.String()),
		))
	defer span.End()

	// Convert value to string
	valueStr := ""
	if value != nil {
		valueStr = fmt.Sprintf("%v", value)
	}

	// For now, implement as delete and create
	// TODO: Implement proper update
	err := r.DeleteAttributeValue(ctx, definitionID, entityID)
	if err != nil {
		// Log error but continue with creation
		r.logger.WarnContext(ctx, "Error deleting existing attribute value", logger.Fields{
			"definition_id": definitionID.String(),
			"entity_id":     entityID.String(),
			"error":         err.Error(),
		})
	}

	// Create new value
	return r.CreateAttributeValue(ctx, &CreateAttributeValueRequest{
		DefinitionID:  definitionID,
		EntityID:      entityID,
		Value:         valueStr,
		IsEncrypted:   false,
		Version:       1,
		EffectiveFrom: time.Now(),
		CreatedBy:     uuid.New(), // TODO: Get actual user ID from context
	})
}

// GetAttributeStats gets attribute statistics
func (r *attributeRepository) GetAttributeStats(ctx context.Context) (*AttributeStats, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeStats")
	defer span.End()

	stats, err := r.store.GetAttributeStats(ctx)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_get_stats_failed", "Total failed attribute stats retrievals").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_STATS_GET_FAILED", "Failed to get attribute statistics").WithDetail("error", err.Error())
	}

	r.metrics.Counter("attribute_repository_get_stats_success", "Total successful attribute stats retrievals").Inc(nil)

	return &AttributeStats{
		TotalDefinitions:       int(stats.TotalDefinitions),
		TotalValues:            int(stats.TotalValues),
		EntitiesWithAttributes: int(stats.EntitiesWithAttributes),
		UserAttributes:         stats.UserAttributes,
		ResourceAttributes:     stats.ResourceAttributes,
		EnvironmentAttributes:  stats.EnvironmentAttributes,
	}, nil
}

func (r *attributeRepository) CleanupExpiredAttributes(ctx context.Context) error {
	return nil
}

func (r *attributeRepository) DeleteAttributeValue(ctx context.Context, definitionID, entityID uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.DeleteAttributeValue",
		tracing.WithAttributes(
			attribute.String("definition_id", definitionID.String()),
			attribute.String("entity_id", entityID.String()),
		))
	defer span.End()

	err := r.store.DeleteAttributeValue(ctx, db.DeleteAttributeValueParams{
		DefinitionID: definitionID,
		EntityID:     entityID,
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("attribute_repository_delete_value_failed", "Total failed attribute value deletions").Inc(nil)
		return errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_VALUE_DELETE_FAILED", "Failed to delete attribute value").WithDetail("error", err.Error())
	}

	r.metrics.Counter("attribute_repository_delete_value_success", "Total successful attribute value deletions").Inc(nil)

	r.logger.InfoContext(ctx, "Attribute value deleted successfully",
		logger.Fields{"definition_id": definitionID, "entity_id": entityID})

	return nil
}

// Helper methods to convert SQLC types to domain models

func (r *attributeRepository) convertSQLCAttributeDefinitionToModel(attrDef *db.AttributeDefinition) (*models.AttributeDefinition, error) {
	var allowedValues []string
	if err := json.Unmarshal(attrDef.AllowedValues, &allowedValues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed_values: %w", err)
	}

	var validationRules map[string]any
	if err := json.Unmarshal(attrDef.ValidationRules, &validationRules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal validation_rules: %w", err)
	}

	return &models.AttributeDefinition{
		ID:                 attrDef.ID,
		TenantID:           attrDef.TenantID,
		Name:               attrDef.Name,
		DisplayName:        attrDef.DisplayName,
		Description:        attrDef.Description,
		DataType:           types.AttributeDataType(attrDef.DataType),
		Category:           types.AttributeCategory(attrDef.Category),
		IsRequired:         *attrDef.IsRequired,
		IsSensitive:        *attrDef.IsSensitive,
		DefaultValue:       attrDef.DefaultValue,
		AllowedValues:      allowedValues,
		ValidationRules:    validationRules,
		EncryptionRequired: *attrDef.EncryptionRequired,
		IsActive:           *attrDef.IsActive,
		CreatedAt:          attrDef.CreatedAt.Time,
	}, nil
}

func (r *attributeRepository) convertSQLCAttributeValueToModel(attrValue *db.AttributeValue) (*models.AttributeValue, error) {
	return &models.AttributeValue{
		DefinitionID: attrValue.DefinitionID,
		Name:         "", // Name is not in the db.AttributeValue, should be fetched from definition
		Value:        attrValue.Value,
		DataType:     "", // DataType is not in the db.AttributeValue
		Category:     "", // Category is not in the db.AttributeValue
		Source:       "", // Source is not in the db.AttributeValue
		Confidence:   0,  // Confidence is not in the db.AttributeValue
		Timestamp:    attrValue.CreatedAt.Time,
		ExpiresAt:    &attrValue.EffectiveTo.Time,
		IsSensitive:  false, // IsSensitive is not in the db.AttributeValue
	}, nil
}

func (r *attributeRepository) convertSQLCAttributeValueRowToModel(row *db.GetAttributeValuesByEntityRow) (*models.AttributeValue, error) {
	return &models.AttributeValue{
		DefinitionID: row.DefinitionID,
		Name:         row.AttributeName,
		Value:        row.Value,
		DataType:     types.AttributeDataType(row.DataType),
		Category:     types.AttributeCategory(row.Category),
		Source:       "", // Source is not in the db.GetAttributeValuesByEntityRow
		Confidence:   0,  // Confidence is not in the db.GetAttributeValuesByEntityRow
		Timestamp:    row.CreatedAt.Time,
		ExpiresAt:    &row.EffectiveTo.Time,
		IsSensitive:  false, // IsSensitive is not in the db.GetAttributeValuesByEntityRow
	}, nil
}
