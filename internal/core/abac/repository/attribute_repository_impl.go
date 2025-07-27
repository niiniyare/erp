package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// attributeRepository implements AttributeRepository using SQLC
type attributeRepository struct {
	db      sqlc.DBTX
	queries *sqlc.Queries
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewAttributeRepository creates a new attribute repository implementation
func NewAttributeRepository(
	db sqlc.DBTX,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) AttributeRepository {
	return &attributeRepository{
		db:      db,
		queries: sqlc.New(db),
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

	attrDef, err := r.queries.CreateAttributeDefinition(ctx, params)
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

	attrDef, err := r.queries.GetAttributeDefinition(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("attribute_repository_definition_not_found")
			return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_NOT_FOUND", "Attribute definition not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "get_definition_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITION_GET_FAILED", "Failed to get attribute definition").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_get_definition")
	return r.convertSQLCAttributeDefinitionToModel(attrDef), nil
}

// GetAttributeDefinitionByName retrieves an attribute definition by name
func (r *attributeRepository) GetAttributeDefinitionByName(ctx context.Context, name string) (*models.AttributeDefinition, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeDefinitionByName",
		tracing.WithAttributes(
			tracing.StringAttribute("name", name),
		))
	defer span.End()

	attrDef, err := r.queries.GetAttributeDefinitionByName(ctx, name)
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

	attrDef, err := r.queries.UpdateAttributeDefinition(ctx, params)
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

	err := r.queries.DeleteAttributeDefinition(ctx, id)
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

	attrDefs, err := r.queries.ListAttributeDefinitions(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("attribute_repository", "list_definitions_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "ATTRIBUTE_DEFINITIONS_LIST_FAILED", "Failed to list attribute definitions").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("attribute_repository_list_definitions")

	result := make([]*models.AttributeDefinition, len(attrDefs))
	for i, attrDef := range attrDefs {
		result[i] = r.convertSQLCAttributeDefinitionToModel(attrDef)
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

	attrValue, err := r.queries.CreateAttributeValue(ctx, params)
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

	rows, err := r.queries.GetAttributeValuesByEntity(ctx, params)
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

	return result, nil
}

// GetAttributeStats gets attribute statistics
func (r *attributeRepository) GetAttributeStats(ctx context.Context) (*AttributeStats, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetAttributeStats")
	defer span.End()

	stats, err := r.queries.GetAttributeStats(ctx)
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
