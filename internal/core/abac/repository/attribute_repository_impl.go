package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// attributeRepository implements AttributeRepository using SQLC and Store
type attributeRepository struct {
	store   sqlc.Store
	cache   cache.Service
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewAttributeRepository creates a new attribute repository implementation
func NewAttributeRepository(
	store sqlc.Store,
	cache cache.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) AttributeRepository {
	return &attributeRepository{
		store:   store,
		cache:   cache,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// CreateAttributeDefinition creates a new attribute definition
func (r *attributeRepository) CreateAttributeDefinition(ctx context.Context, req *CreateAttributeDefinitionRequest) (*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.CreateAttributeDefinition",
		tracing.WithAttributes(
			tracing.StringAttribute("name", req.Name),
			tracing.StringAttribute("data_type", string(req.DataType)),
			tracing.StringAttribute("category", string(req.Category)),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Creating attribute definition",
		logger.Fields{
			"name":      req.Name,
			"data_type": req.DataType,
			"category":  req.Category,
		})

	params := sqlc.CreateAttributeDefinitionParams{
		Name:               req.Name,
		DisplayName:        req.DisplayName,
		Description:        req.Description,
		DataType:           string(req.DataType),
		Category:           string(req.Category),
		IsRequired:         req.IsRequired,
		IsSensitive:        req.IsSensitive,
		DefaultValue:       req.DefaultValue,
		AllowedValues:      req.AllowedValues,
		ValidationRules:    req.ValidationRules,
		EncryptionRequired: req.EncryptionRequired,
		IsActive:           req.IsActive,
	}

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	var attrDef sqlc.AttributeDefinition
	// Use tenant-aware transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore sqlc.Store) error {
		var err error
		attrDef, err = txStore.CreateAttributeDefinition(ctx, params)
		return err
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "create_definition_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_CREATE_FAILED", "Failed to create attribute definition").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_create_definition")

	result := r.convertSQLCAttributeDefinitionToModel(attrDef)

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
			tracing.StringAttribute("definition_id", id.String()),
		))
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Try cache first
	cacheKey := fmt.Sprintf("attr_def:%s", id.String())
	var attrDef *models.AttributeDefinition
	if err := r.cache.Get(ctx, cacheKey, &attrDef); err == nil {
		r.metrics.IncrementCounter("attribute_repository_definition_cache_hit")
		return attrDef, nil
	}

	// Cache miss - get from database
	sqlcAttrDef, err := r.store.GetAttributeDefinition(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("attribute_repository_definition_not_found")
			return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_NOT_FOUND", "Attribute definition not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "get_definition_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_GET_FAILED", "Failed to get attribute definition").WithErr(err)
	}

	// Convert to domain model
	attrDef = r.convertSQLCAttributeDefinitionToModel(sqlcAttrDef)

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, attrDef, 30*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache attribute definition", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementSuccessCount("attribute_repository_get_definition")
	return attrDef, nil
}

// GetAttributeDefinitionByName retrieves an attribute definition by name
func (r *attributeRepository) GetAttributeDefinitionByName(ctx context.Context, name string) (*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeDefinitionByName",
		tracing.WithAttributes(
			tracing.StringAttribute("name", name),
		))
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Try cache first
	cacheKey := fmt.Sprintf("attr_def_name:%s", name)
	var attrDef *models.AttributeDefinition
	if err := r.cache.Get(ctx, cacheKey, &attrDef); err == nil {
		r.metrics.IncrementCounter("attribute_repository_definition_name_cache_hit")
		return attrDef, nil
	}

	// Cache miss - get from database
	sqlcAttrDef, err := r.store.GetAttributeDefinitionByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("attribute_repository_definition_not_found")
			return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_NOT_FOUND", "Attribute definition not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "get_definition_by_name_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_GET_FAILED", "Failed to get attribute definition by name").WithErr(err)
	}

	// Convert to domain model
	attrDef = r.convertSQLCAttributeDefinitionToModel(sqlcAttrDef)

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, attrDef, 30*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache attribute definition by name", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementSuccessCount("attribute_repository_get_definition_by_name")
	return attrDef, nil
}

// GetAttributeDefinitionByName retrieves an attribute definition by name
func (r *attributeRepository) GetAttributeDefinitionByName(ctx context.Context, name string) (*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeDefinitionByName",
		tracing.WithAttributes(
			tracing.StringAttribute("name", name),
		))
	defer span.End()

	sqlcAttrDef, err := r.store.GetAttributeDefinitionByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("attribute_repository_definition_not_found")
			return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_NOT_FOUND", "Attribute definition not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "get_definition_by_name_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_GET_FAILED", "Failed to get attribute definition by name").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_get_definition_by_name")
	return r.convertSQLCAttributeDefinitionToModel(attrDef), nil
}

// UpdateAttributeDefinition updates an existing attribute definition
func (r *attributeRepository) UpdateAttributeDefinition(ctx context.Context, req *UpdateAttributeDefinitionRequest) (*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.UpdateAttributeDefinition",
		tracing.WithAttributes(
			tracing.StringAttribute("definition_id", req.ID.String()),
		))
	defer span.End()

	params := sqlc.UpdateAttributeDefinitionParams{
		ID:                 req.ID,
		Name:               req.Name,
		DisplayName:        req.DisplayName,
		Description:        req.Description,
		DataType:           req.DataType,
		Category:           req.Category,
		IsRequired:         req.IsRequired,
		IsSensitive:        req.IsSensitive,
		DefaultValue:       req.DefaultValue,
		AllowedValues:      req.AllowedValues,
		ValidationRules:    req.ValidationRules,
		EncryptionRequired: req.EncryptionRequired,
		IsActive:           req.IsActive,
	}

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	var attrDef sqlc.AttributeDefinition
	// Use tenant-aware transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore sqlc.Store) error {
		var err error
		attrDef, err = txStore.UpdateAttributeDefinition(ctx, params)
		return err
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "update_definition_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_UPDATE_FAILED", "Failed to update attribute definition").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_update_definition")
	return r.convertSQLCAttributeDefinitionToModel(attrDef), nil
}

// DeleteAttributeDefinition deletes an attribute definition
func (r *attributeRepository) DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.DeleteAttributeDefinition",
		tracing.WithAttributes(
			tracing.StringAttribute("definition_id", id.String()),
		))
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Use tenant-aware transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore sqlc.Store) error {
		return txStore.DeleteAttributeDefinition(ctx, id)
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "delete_definition_failed")
		return errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_DELETE_FAILED", "Failed to delete attribute definition").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_delete_definition")

	r.logger.InfoContext(ctx, "Attribute definition deleted successfully",
		logger.Fields{"definition_id": id})

	return nil
}

// ListAttributeDefinitions lists attribute definitions with filtering and pagination
func (r *attributeRepository) ListAttributeDefinitions(ctx context.Context, req *ListAttributeDefinitionsRequest) ([]*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.ListAttributeDefinitions")
	defer span.End()

	params := sqlc.ListAttributeDefinitionsParams{
		Column1: req.Search,
		Column2: req.Category,
		Column3: req.IsActive,
		Column4: int32(req.Limit),
		Column5: int32(req.Offset),
	}

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Try cache first for list
	cacheKey := fmt.Sprintf("attr_defs:search:%s:cat:%s:active:%v:limit:%d:offset:%d",
		*req.Search, *req.Category, *req.IsActive, req.Limit, req.Offset)
	var attrDefs []sqlc.ListAttributeDefinitionsRow
	if err := r.cache.Get(ctx, cacheKey, &attrDefs); err == nil {
		r.metrics.IncrementCounter("attribute_repository_list_cache_hit")
		// Convert cached results
		result := make([]*models.AttributeDefinition, len(attrDefs))
		for i, attrDef := range attrDefs {
			result[i] = r.convertSQLCAttributeDefinitionToModel(sqlc.AttributeDefinition{
				ID:                 attrDef.ID,
				TenantID:           attrDef.TenantID,
				Name:               attrDef.Name,
				DisplayName:        attrDef.DisplayName,
				Description:        attrDef.Description,
				DataType:           attrDef.DataType,
				Category:           attrDef.Category,
				IsRequired:         attrDef.IsRequired,
				IsSensitive:        attrDef.IsSensitive,
				DefaultValue:       attrDef.DefaultValue,
				AllowedValues:      attrDef.AllowedValues,
				ValidationRules:    attrDef.ValidationRules,
				EncryptionRequired: attrDef.EncryptionRequired,
				IsActive:           attrDef.IsActive,
				CreatedAt:          attrDef.CreatedAt,
			})
		}
		return result, nil
	}

	// Cache miss - get from database
	attrDefs, err := r.store.ListAttributeDefinitions(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "list_definitions_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITIONS_LIST_FAILED", "Failed to list attribute definitions").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_list_definitions")

	result := make([]*models.AttributeDefinition, len(attrDefs))
	for i, attrDef := range attrDefs {
		result[i] = r.convertSQLCAttributeDefinitionToModel(sqlc.AttributeDefinition{
			ID:                 attrDef.ID,
			TenantID:           attrDef.TenantID,
			Name:               attrDef.Name,
			DisplayName:        attrDef.DisplayName,
			Description:        attrDef.Description,
			DataType:           attrDef.DataType,
			Category:           attrDef.Category,
			IsRequired:         attrDef.IsRequired,
			IsSensitive:        attrDef.IsSensitive,
			DefaultValue:       attrDef.DefaultValue,
			AllowedValues:      attrDef.AllowedValues,
			ValidationRules:    attrDef.ValidationRules,
			EncryptionRequired: attrDef.EncryptionRequired,
			IsActive:           attrDef.IsActive,
			CreatedAt:          attrDef.CreatedAt,
		})
	}

	// Cache the results
	if err := r.cache.Set(ctx, cacheKey, attrDefs, 15*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache attribute definitions list", logger.Fields{"error": err.Error()})
	}

	return result, nil
}

// CreateAttributeValue creates a new attribute value
func (r *attributeRepository) CreateAttributeValue(ctx context.Context, req *CreateAttributeValueRequest) (*models.AttributeValue, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.CreateAttributeValue",
		tracing.WithAttributes(
			tracing.StringAttribute("definition_id", req.DefinitionID.String()),
			tracing.StringAttribute("entity_id", req.EntityID.String()),
		))
	defer span.End()

	params := sqlc.CreateAttributeValueParams{
		DefinitionID:   req.DefinitionID,
		EntityID:       req.EntityID,
		Value:          req.Value,
		EncryptedValue: req.EncryptedValue,
		IsEncrypted:    req.IsEncrypted,
		Version:        req.Version,
		EffectiveFrom:  req.EffectiveFrom,
		EffectiveTo:    req.EffectiveTo,
		CreatedBy:      req.CreatedBy,
	}

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	var attrValue sqlc.AttributeValue
	// Use tenant-aware transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore sqlc.Store) error {
		var err error
		attrValue, err = txStore.CreateAttributeValue(ctx, params)
		return err
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "create_value_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_VALUE_CREATE_FAILED", "Failed to create attribute value").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_create_value")
	return r.convertSQLCAttributeValueToModel(attrValue), nil
}

// GetAttributeValuesByEntity retrieves attribute values by entity ID
func (r *attributeRepository) GetAttributeValuesByEntity(ctx context.Context, entityID uuid.UUID, category types.AttributeCategory) ([]*models.AttributeValue, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeValuesByEntity",
		tracing.WithAttributes(
			tracing.StringAttribute("entity_id", entityID.String()),
			tracing.StringAttribute("category", string(category)),
		))
	defer span.End()

	var categoryStr *string
	if category != "" {
		cat := string(category)
		categoryStr = &cat
	}

	params := sqlc.GetAttributeValuesByEntityParams{
		EntityID: entityID,
		Column2:  categoryStr,
	}

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Try cache first
	cacheKey := fmt.Sprintf("entity_attrs:%s:%s", entityID.String(), string(category))
	var rows []sqlc.GetAttributeValuesByEntityRow
	if err := r.cache.Get(ctx, cacheKey, &rows); err == nil {
		r.metrics.IncrementCounter("attribute_repository_entity_values_cache_hit")
		// Convert cached results
		result := make([]*models.AttributeValue, len(rows))
		for i, row := range rows {
			result[i] = r.convertSQLCAttributeValueRowToModel(row)
		}
		return result, nil
	}

	// Cache miss - get from database
	rows, err := r.store.GetAttributeValuesByEntity(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "get_values_by_entity_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_VALUES_GET_FAILED", "Failed to get attribute values by entity").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_get_values_by_entity")

	result := make([]*models.AttributeValue, len(rows))
	for i, row := range rows {
		result[i] = r.convertSQLCAttributeValueRowToModel(row)
	}

	// Cache the results
	if err := r.cache.Set(ctx, cacheKey, rows, 20*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache entity attribute values", logger.Fields{"error": err.Error()})
	}

	return result, nil
}

// GetAttributeStats gets attribute statistics
func (r *attributeRepository) GetAttributeStats(ctx context.Context) (*AttributeStats, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeStats")
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	stats, err := r.store.GetAttributeStats(ctx)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "get_stats_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_STATS_GET_FAILED", "Failed to get attribute statistics").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_get_stats")

	return &AttributeStats{
		TotalDefinitions:       int(stats.TotalDefinitions),
		TotalValues:            int(stats.TotalValues),
		EntitiesWithAttributes: int(stats.EntitiesWithAttributes),
		UserAttributes:         int(stats.UserAttributes.Int64),
		ResourceAttributes:     int(stats.ResourceAttributes.Int64),
		EnvironmentAttributes:  int(stats.EnvironmentAttributes.Int64),
	}, nil
}

// Helper methods to convert SQLC types to domain models

func (r *attributeRepository) convertSQLCAttributeDefinitionToModel(attrDef sqlc.AttributeDefinition) *models.AttributeDefinition {
	var displayName *string
	if attrDef.DisplayName.Valid {
		displayName = &attrDef.DisplayName.String
	}

	var description *string
	if attrDef.Description.Valid {
		description = &attrDef.Description.String
	}

	var defaultValue *string
	if attrDef.DefaultValue.Valid {
		defaultValue = &attrDef.DefaultValue.String
	}

	return &models.AttributeDefinition{
		ID:                 attrDef.ID,
		TenantID:           attrDef.TenantID,
		Name:               attrDef.Name,
		DisplayName:        displayName,
		Description:        description,
		DataType:           types.AttributeDataType(attrDef.DataType),
		Category:           types.AttributeCategory(attrDef.Category),
		IsRequired:         attrDef.IsRequired,
		IsSensitive:        attrDef.IsSensitive,
		DefaultValue:       defaultValue,
		AllowedValues:      attrDef.AllowedValues,
		ValidationRules:    attrDef.ValidationRules,
		EncryptionRequired: attrDef.EncryptionRequired,
		IsActive:           attrDef.IsActive,
		CreatedAt:          attrDef.CreatedAt,
	}
}

func (r *attributeRepository) convertSQLCAttributeValueToModel(attrValue sqlc.AttributeValue) *models.AttributeValue {
	var createdBy *uuid.UUID
	if attrValue.CreatedBy.Valid {
		createdBy = &attrValue.CreatedBy.UUID
	}

	var updatedBy *uuid.UUID
	if attrValue.UpdatedBy.Valid {
		updatedBy = &attrValue.UpdatedBy.UUID
	}

	return &models.AttributeValue{
		ID:             attrValue.ID,
		TenantID:       attrValue.TenantID,
		DefinitionID:   attrValue.DefinitionID,
		EntityID:       attrValue.EntityID,
		Value:          attrValue.Value,
		EncryptedValue: attrValue.EncryptedValue,
		IsEncrypted:    attrValue.IsEncrypted,
		Version:        attrValue.Version,
		EffectiveFrom:  attrValue.EffectiveFrom,
		EffectiveTo:    attrValue.EffectiveTo,
		CreatedAt:      attrValue.CreatedAt,
		CreatedBy:      createdBy,
		UpdatedAt:      attrValue.UpdatedAt,
		UpdatedBy:      updatedBy,
	}
}

func (r *attributeRepository) convertSQLCAttributeValueRowToModel(row sqlc.GetAttributeValuesByEntityRow) *models.AttributeValue {
	var createdBy *uuid.UUID
	if row.CreatedBy.Valid {
		createdBy = &row.CreatedBy.UUID
	}

	var updatedBy *uuid.UUID
	if row.UpdatedBy.Valid {
		updatedBy = &row.UpdatedBy.UUID
	}

	// Create attribute value with embedded attribute definition info
	attrValue := &models.AttributeValue{
		ID:             row.ID,
		TenantID:       row.TenantID,
		DefinitionID:   row.DefinitionID,
		EntityID:       row.EntityID,
		Value:          row.Value,
		EncryptedValue: row.EncryptedValue,
		IsEncrypted:    row.IsEncrypted,
		Version:        row.Version,
		EffectiveFrom:  row.EffectiveFrom,
		EffectiveTo:    row.EffectiveTo,
		CreatedAt:      row.CreatedAt,
		CreatedBy:      createdBy,
		UpdatedAt:      row.UpdatedAt,
		UpdatedBy:      updatedBy,
	}

	// Add metadata from joined attribute definition
	attrValue.AttributeName = row.AttributeName
	attrValue.DataType = types.AttributeDataType(row.DataType)
	attrValue.Category = types.AttributeCategory(row.Category)

	return attrValue
}
