package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// AttributeValidationService defines the interface for attribute validation
type AttributeValidationService interface {
	// Core validation methods
	ValidateAttribute(ctx context.Context, attrValue *models.AttributeValue, definition *models.AttributeDefinition) error
	ValidateAttributeCollection(ctx context.Context, attributes map[string]*models.AttributeValue) ([]ValidationError, error)
	ValidateAttributeContext(ctx context.Context, attrContext *models.AttributeContext) ([]ValidationError, error)

	// Validation rules
	ValidateDataType(ctx context.Context, value interface{}, dataType types.AttributeDataType) error
	ValidateConstraints(ctx context.Context, value interface{}, constraints map[string]interface{}) error
	ValidateEnumValues(ctx context.Context, value interface{}, allowedValues []interface{}) error

	// Utility methods
	NormalizeAttributeValue(ctx context.Context, value interface{}, dataType types.AttributeDataType) (interface{}, error)
	GetValidationRules(ctx context.Context, attributeName string) (*ValidationRules, error)
	ValidateRequired(ctx context.Context, attributes map[string]*models.AttributeValue, requiredAttributes []string) error
}

// ValidationError represents a validation error
type ValidationError struct {
	AttributeName string `json:"attribute_name"`
	ErrorCode     string `json:"error_code"`
	ErrorMessage  string `json:"error_message"`
	Field         string `json:"field,omitempty"`
	Value         string `json:"value,omitempty"`
}

// ValidationRules represents validation rules for an attribute
type ValidationRules struct {
	Required        bool                    `json:"required"`
	DataType        types.AttributeDataType `json:"data_type"`
	MinLength       *int                    `json:"min_length,omitempty"`
	MaxLength       *int                    `json:"max_length,omitempty"`
	MinValue        *float64                `json:"min_value,omitempty"`
	MaxValue        *float64                `json:"max_value,omitempty"`
	Pattern         *string                 `json:"pattern,omitempty"`
	AllowedValues   []interface{}           `json:"allowed_values,omitempty"`
	CustomValidator *string                 `json:"custom_validator,omitempty"`
	Constraints     map[string]interface{}  `json:"constraints,omitempty"`
}

// ValidationContext provides context for validation operations
type ValidationContext struct {
	TenantID       uuid.UUID              `json:"tenant_id"`
	UserID         uuid.UUID              `json:"user_id"`
	RequestID      string                 `json:"request_id"`
	ValidationTime time.Time              `json:"validation_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// attributeValidationService implements AttributeValidationService
type attributeValidationService struct {
	attrDefRepo repository.AttributeDefinitionRepository
	tracing     tracing.TracingService
	metrics     metrics.Provider
	logger      logger.Logger
}

// NewAttributeValidationService creates a new attribute validation service
func NewAttributeValidationService(
	attrDefRepo repository.AttributeDefinitionRepository,
	tracing tracing.TracingService,
	metrics metrics.Provider,
	logger logger.Logger,
) AttributeValidationService {
	return &attributeValidationService{
		attrDefRepo: attrDefRepo,
		tracing:     tracing,
		metrics:     metrics,
		logger:      logger,
	}
}

// ValidateAttribute validates a single attribute value against its definition
func (s *attributeValidationService) ValidateAttribute(ctx context.Context, attrValue *models.AttributeValue, definition *models.AttributeDefinition) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeValidationService.ValidateAttribute",
		tracing.WithAttributes(
			attribute.String("attribute.name", attrValue.Name),
			attribute.String("attribute.category", string(attrValue.Category)),
		))
	defer span.End()

	startTime := time.Now()
	defer func() {
		s.recordValidationMetrics(ctx, "validate_attribute", "success", time.Since(startTime))
	}()

	// Check if attribute is required
	if definition.Required && (attrValue.Value == nil || isEmptyValue(attrValue.Value)) {
		err := errors.NewBusinessError("ATTRIBUTE_REQUIRED", "Required attribute is missing or empty").
			WithDetail("attribute_name", attrValue.Name).
			WithDetail("category", string(attrValue.Category))
		span.RecordError(err)
		span.SetStatus(codes.Error, "Required attribute validation failed")
		return err
	}

	// Skip validation for nil values on non-required attributes
	if attrValue.Value == nil {
		return nil
	}

	// Validate data type
	if err := s.ValidateDataType(ctx, attrValue.Value, definition.DataType); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Data type validation failed")
		return err
	}

	// Validate constraints
	if definition.Constraints != nil {
		if err := s.ValidateConstraints(ctx, attrValue.Value, definition.Constraints); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Constraint validation failed")
			return err
		}
	}

	// Validate allowed values (enum validation)
	if len(definition.AllowedValues) > 0 {
		if err := s.ValidateEnumValues(ctx, attrValue.Value, definition.AllowedValues); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Enum validation failed")
			return err
		}
	}

	// Validate expiration
	if attrValue.ExpiresAt != nil && time.Now().After(*attrValue.ExpiresAt) {
		err := errors.NewBusinessError("ATTRIBUTE_EXPIRED", "Attribute value has expired").
			WithDetail("attribute_name", attrValue.Name).
			WithDetail("expires_at", attrValue.ExpiresAt.Format(time.RFC3339))
		span.RecordError(err)
		span.SetStatus(codes.Error, "Attribute expiration validation failed")
		return err
	}

	return nil
}

// ValidateAttributeCollection validates a collection of attributes
func (s *attributeValidationService) ValidateAttributeCollection(ctx context.Context, attributes map[string]*models.AttributeValue) ([]ValidationError, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeValidationService.ValidateAttributeCollection",
		tracing.WithAttributes(attribute.Int("attributes.count", len(attributes))))
	defer span.End()

	startTime := time.Now()
	defer func() {
		s.recordValidationMetrics(ctx, "validate_collection", "success", time.Since(startTime))
	}()

	var validationErrors []ValidationError

	// Get all attribute definitions for the attributes we're validating
	attributeNames := make([]string, 0, len(attributes))
	for name := range attributes {
		attributeNames = append(attributeNames, name)
	}

	definitions, err := s.attrDefRepo.GetByNames(ctx, attributeNames)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get attribute definitions")
		return nil, errors.NewBusinessError("VALIDATION_DEFINITION_FETCH_FAILED", "Failed to fetch attribute definitions").
			WithDetail("attribute_names", strings.Join(attributeNames, ", "))
	}

	// Create a map for quick definition lookup
	definitionMap := make(map[string]*models.AttributeDefinition)
	for _, def := range definitions {
		definitionMap[def.Name] = def
	}

	// Validate each attribute
	for name, attrValue := range attributes {
		definition, exists := definitionMap[name]
		if !exists {
			validationErrors = append(validationErrors, ValidationError{
				AttributeName: name,
				ErrorCode:     "ATTRIBUTE_DEFINITION_NOT_FOUND",
				ErrorMessage:  fmt.Sprintf("No definition found for attribute '%s'", name),
			})
			continue
		}

		if err := s.ValidateAttribute(ctx, attrValue, definition); err != nil {
			if businessErr, ok := err.(*errors.BusinessError); ok {
				validationErrors = append(validationErrors, ValidationError{
					AttributeName: name,
					ErrorCode:     businessErr.Code,
					ErrorMessage:  businessErr.Message,
					Value:         fmt.Sprintf("%v", attrValue.Value),
				})
			} else {
				validationErrors = append(validationErrors, ValidationError{
					AttributeName: name,
					ErrorCode:     "VALIDATION_ERROR",
					ErrorMessage:  err.Error(),
					Value:         fmt.Sprintf("%v", attrValue.Value),
				})
			}
		}
	}

	if len(validationErrors) > 0 {
		s.logger.WarnContext(ctx, "Attribute collection validation failed",
			logger.Fields{"validation_errors": validationErrors})
	}

	return validationErrors, nil
}

// ValidateAttributeContext validates an entire attribute context
func (s *attributeValidationService) ValidateAttributeContext(ctx context.Context, attrContext *models.AttributeContext) ([]ValidationError, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeValidationService.ValidateAttributeContext")
	defer span.End()

	startTime := time.Now()
	defer func() {
		s.recordValidationMetrics(ctx, "validate_context", "success", time.Since(startTime))
	}()

	var allValidationErrors []ValidationError

	// Validate user attributes
	if attrContext.UserAttributes != nil {
		if errors, err := s.ValidateAttributeCollection(ctx, attrContext.UserAttributes); err != nil {
			return nil, err
		} else {
			allValidationErrors = append(allValidationErrors, errors...)
		}
	}

	// Validate resource attributes
	if attrContext.ResourceAttributes != nil {
		if errors, err := s.ValidateAttributeCollection(ctx, attrContext.ResourceAttributes); err != nil {
			return nil, err
		} else {
			allValidationErrors = append(allValidationErrors, errors...)
		}
	}

	// Validate environment attributes
	if attrContext.EnvironmentAttributes != nil {
		if errors, err := s.ValidateAttributeCollection(ctx, attrContext.EnvironmentAttributes); err != nil {
			return nil, err
		} else {
			allValidationErrors = append(allValidationErrors, errors...)
		}
	}

	// Validate action attributes
	if attrContext.ActionAttributes != nil {
		if errors, err := s.ValidateAttributeCollection(ctx, attrContext.ActionAttributes); err != nil {
			return nil, err
		} else {
			allValidationErrors = append(allValidationErrors, errors...)
		}
	}

	// Validate entity attributes
	if attrContext.EntityAttributes != nil {
		if errors, err := s.ValidateAttributeCollection(ctx, attrContext.EntityAttributes); err != nil {
			return nil, err
		} else {
			allValidationErrors = append(allValidationErrors, errors...)
		}
	}

	// Validate session attributes
	if attrContext.SessionAttributes != nil {
		if errors, err := s.ValidateAttributeCollection(ctx, attrContext.SessionAttributes); err != nil {
			return nil, err
		} else {
			allValidationErrors = append(allValidationErrors, errors...)
		}
	}

	return allValidationErrors, nil
}

// ValidateDataType validates that a value matches the expected data type
func (s *attributeValidationService) ValidateDataType(ctx context.Context, value interface{}, dataType types.AttributeDataType) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeValidationService.ValidateDataType",
		tracing.WithAttributes(attribute.String("data_type", string(dataType))))
	defer span.End()

	switch dataType {
	case types.AttributeDataTypeString:
		if _, ok := value.(string); !ok {
			return errors.NewBusinessError("INVALID_DATA_TYPE", "Value must be a string").
				WithDetail("expected_type", "string").
				WithDetail("actual_value", fmt.Sprintf("%v", value))
		}

	case types.AttributeDataTypeNumber:
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Valid number types
		default:
			return errors.NewBusinessError("INVALID_DATA_TYPE", "Value must be a number").
				WithDetail("expected_type", "number").
				WithDetail("actual_value", fmt.Sprintf("%v", value))
		}

	case types.AttributeDataTypeBoolean:
		if _, ok := value.(bool); !ok {
			return errors.NewBusinessError("INVALID_DATA_TYPE", "Value must be a boolean").
				WithDetail("expected_type", "boolean").
				WithDetail("actual_value", fmt.Sprintf("%v", value))
		}

	case types.AttributeDataTypeDate:
		switch v := value.(type) {
		case time.Time:
			// Already a time.Time
		case string:
			// Try to parse as RFC3339 date
			if _, err := time.Parse(time.RFC3339, v); err != nil {
				return errors.NewBusinessError("INVALID_DATE_FORMAT", "Date string must be in RFC3339 format").
					WithDetail("expected_format", "RFC3339").
					WithDetail("actual_value", v)
			}
		default:
			return errors.NewBusinessError("INVALID_DATA_TYPE", "Value must be a date (time.Time or RFC3339 string)").
				WithDetail("expected_type", "date").
				WithDetail("actual_value", fmt.Sprintf("%v", value))
		}

	case types.AttributeDataTypeJSON:
		// Try to marshal to JSON to ensure it's valid
		if _, err := json.Marshal(value); err != nil {
			return errors.NewBusinessError("INVALID_JSON", "Value must be valid JSON").
				WithDetail("json_error", err.Error()).
				WithDetail("actual_value", fmt.Sprintf("%v", value))
		}

	case types.AttributeDataTypeArray:
		// Check if it's a slice or array
		switch value.(type) {
		case []interface{}, []string, []int, []float64, []bool:
			// Valid array types
		default:
			return errors.NewBusinessError("INVALID_DATA_TYPE", "Value must be an array").
				WithDetail("expected_type", "array").
				WithDetail("actual_value", fmt.Sprintf("%v", value))
		}

	case types.AttributeDataTypeEnum:
		// Enum values are typically strings, but could be other types
		// The actual enum validation happens in ValidateEnumValues
		break

	default:
		return errors.NewBusinessError("UNSUPPORTED_DATA_TYPE", "Unsupported data type").
			WithDetail("data_type", string(dataType))
	}

	return nil
}

// ValidateConstraints validates value against constraints
func (s *attributeValidationService) ValidateConstraints(ctx context.Context, value interface{}, constraints map[string]interface{}) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeValidationService.ValidateConstraints")
	defer span.End()

	for constraintName, constraintValue := range constraints {
		switch constraintName {
		case "min_length":
			if err := s.validateMinLength(value, constraintValue); err != nil {
				return err
			}
		case "max_length":
			if err := s.validateMaxLength(value, constraintValue); err != nil {
				return err
			}
		case "min_value":
			if err := s.validateMinValue(value, constraintValue); err != nil {
				return err
			}
		case "max_value":
			if err := s.validateMaxValue(value, constraintValue); err != nil {
				return err
			}
		case "pattern":
			if err := s.validatePattern(value, constraintValue); err != nil {
				return err
			}
		case "custom":
			if err := s.validateCustomConstraint(ctx, value, constraintValue); err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidateEnumValues validates that a value is within allowed enum values
func (s *attributeValidationService) ValidateEnumValues(ctx context.Context, value interface{}, allowedValues []interface{}) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeValidationService.ValidateEnumValues")
	defer span.End()

	for _, allowedValue := range allowedValues {
		if compareValues(value, allowedValue) {
			return nil
		}
	}

	return errors.NewBusinessError("INVALID_ENUM_VALUE", "Value is not in the list of allowed values").
		WithDetail("actual_value", fmt.Sprintf("%v", value)).
		WithDetail("allowed_values", fmt.Sprintf("%v", allowedValues))
}

// NormalizeAttributeValue normalizes an attribute value according to its data type
func (s *attributeValidationService) NormalizeAttributeValue(ctx context.Context, value interface{}, dataType types.AttributeDataType) (interface{}, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeValidationService.NormalizeAttributeValue")
	defer span.End()

	switch dataType {
	case types.AttributeDataTypeString:
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str), nil
		}
		return fmt.Sprintf("%v", value), nil

	case types.AttributeDataTypeNumber:
		switch v := value.(type) {
		case string:
			if num, err := strconv.ParseFloat(v, 64); err == nil {
				return num, nil
			}
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return v, nil
		}

	case types.AttributeDataTypeBoolean:
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return strings.ToLower(v) == "true", nil
		case int, int8, int16, int32, int64:
			return v != 0, nil
		}

	case types.AttributeDataTypeDate:
		switch v := value.(type) {
		case time.Time:
			return v, nil
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t, nil
			}
		}

	default:
		return value, nil
	}

	return value, nil
}

// GetValidationRules retrieves validation rules for an attribute
func (s *attributeValidationService) GetValidationRules(ctx context.Context, attributeName string) (*ValidationRules, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeValidationService.GetValidationRules",
		tracing.WithAttributes(attribute.String("attribute.name", attributeName)))
	defer span.End()

	definition, err := s.attrDefRepo.GetByName(ctx, attributeName)
	if err != nil {
		return nil, err
	}

	rules := &ValidationRules{
		Required:      definition.Required,
		DataType:      definition.DataType,
		AllowedValues: definition.AllowedValues,
		Constraints:   definition.Constraints,
	}

	// Extract specific constraints
	if definition.Constraints != nil {
		if minLen, ok := definition.Constraints["min_length"].(int); ok {
			rules.MinLength = &minLen
		}
		if maxLen, ok := definition.Constraints["max_length"].(int); ok {
			rules.MaxLength = &maxLen
		}
		if minVal, ok := definition.Constraints["min_value"].(float64); ok {
			rules.MinValue = &minVal
		}
		if maxVal, ok := definition.Constraints["max_value"].(float64); ok {
			rules.MaxValue = &maxVal
		}
		if pattern, ok := definition.Constraints["pattern"].(string); ok {
			rules.Pattern = &pattern
		}
		if validator, ok := definition.Constraints["custom_validator"].(string); ok {
			rules.CustomValidator = &validator
		}
	}

	return rules, nil
}

// ValidateRequired validates that all required attributes are present
func (s *attributeValidationService) ValidateRequired(ctx context.Context, attributes map[string]*models.AttributeValue, requiredAttributes []string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeValidationService.ValidateRequired")
	defer span.End()

	var missingAttributes []string

	for _, required := range requiredAttributes {
		if attr, exists := attributes[required]; !exists || attr.Value == nil || isEmptyValue(attr.Value) {
			missingAttributes = append(missingAttributes, required)
		}
	}

	if len(missingAttributes) > 0 {
		return errors.NewBusinessError("REQUIRED_ATTRIBUTES_MISSING", "Required attributes are missing").
			WithDetail("missing_attributes", strings.Join(missingAttributes, ", "))
	}

	return nil
}

// Helper methods for constraint validation

func (s *attributeValidationService) validateMinLength(value interface{}, constraint interface{}) error {
	minLength, ok := constraint.(int)
	if !ok {
		return errors.NewBusinessError("INVALID_CONSTRAINT", "min_length constraint must be an integer")
	}

	str, ok := value.(string)
	if !ok {
		return nil // Skip length validation for non-strings
	}

	if len(str) < minLength {
		return errors.NewBusinessError("MIN_LENGTH_VIOLATION", "Value is shorter than minimum length").
			WithDetail("actual_length", len(str)).
			WithDetail("min_length", minLength)
	}

	return nil
}

func (s *attributeValidationService) validateMaxLength(value interface{}, constraint interface{}) error {
	maxLength, ok := constraint.(int)
	if !ok {
		return errors.NewBusinessError("INVALID_CONSTRAINT", "max_length constraint must be an integer")
	}

	str, ok := value.(string)
	if !ok {
		return nil // Skip length validation for non-strings
	}

	if len(str) > maxLength {
		return errors.NewBusinessError("MAX_LENGTH_VIOLATION", "Value is longer than maximum length").
			WithDetail("actual_length", len(str)).
			WithDetail("max_length", maxLength)
	}

	return nil
}

func (s *attributeValidationService) validateMinValue(value interface{}, constraint interface{}) error {
	minValue, ok := constraint.(float64)
	if !ok {
		return errors.NewBusinessError("INVALID_CONSTRAINT", "min_value constraint must be a number")
	}

	var numValue float64
	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float32:
		numValue = float64(v)
	case float64:
		numValue = v
	default:
		return nil // Skip numeric validation for non-numbers
	}

	if numValue < minValue {
		return errors.NewBusinessError("MIN_VALUE_VIOLATION", "Value is less than minimum value").
			WithDetail("actual_value", numValue).
			WithDetail("min_value", minValue)
	}

	return nil
}

func (s *attributeValidationService) validateMaxValue(value interface{}, constraint interface{}) error {
	maxValue, ok := constraint.(float64)
	if !ok {
		return errors.NewBusinessError("INVALID_CONSTRAINT", "max_value constraint must be a number")
	}

	var numValue float64
	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float32:
		numValue = float64(v)
	case float64:
		numValue = v
	default:
		return nil // Skip numeric validation for non-numbers
	}

	if numValue > maxValue {
		return errors.NewBusinessError("MAX_VALUE_VIOLATION", "Value is greater than maximum value").
			WithDetail("actual_value", numValue).
			WithDetail("max_value", maxValue)
	}

	return nil
}

func (s *attributeValidationService) validatePattern(value interface{}, constraint interface{}) error {
	pattern, ok := constraint.(string)
	if !ok {
		return errors.NewBusinessError("INVALID_CONSTRAINT", "pattern constraint must be a string")
	}

	str, ok := value.(string)
	if !ok {
		return nil // Skip pattern validation for non-strings
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return errors.NewBusinessError("INVALID_PATTERN", "Invalid regular expression pattern").
			WithDetail("pattern", pattern).
			WithDetail("error", err.Error())
	}

	if !regex.MatchString(str) {
		return errors.NewBusinessError("PATTERN_VIOLATION", "Value does not match required pattern").
			WithDetail("actual_value", str).
			WithDetail("pattern", pattern)
	}

	return nil
}

func (s *attributeValidationService) validateCustomConstraint(ctx context.Context, value interface{}, constraint interface{}) error {
	// TODO: Implement custom constraint validation
	// This could involve calling external validation services or applying custom business rules
	s.logger.WarnContext(ctx, "Custom constraint validation not implemented",
		logger.Fields{"constraint": constraint, "value": value})
	return nil
}

// Helper functions

func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

func compareValues(a, b interface{}) bool {
	// Handle different numeric types
	switch va := a.(type) {
	case int:
		switch vb := b.(type) {
		case int:
			return va == vb
		case float64:
			return float64(va) == vb
		}
	case float64:
		switch vb := b.(type) {
		case int:
			return va == float64(vb)
		case float64:
			return va == vb
		}
	case string:
		if vb, ok := b.(string); ok {
			return va == vb
		}
	case bool:
		if vb, ok := b.(bool); ok {
			return va == vb
		}
	}

	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// recordValidationMetrics records validation operation metrics
func (s *attributeValidationService) recordValidationMetrics(ctx context.Context, operation, status string, duration time.Duration) {
	// Validation operation counter
	counter := s.metrics.Counter(
		"abac_attribute_validation_operations_total",
		"Total number of attribute validation operations",
		"operation", "status",
	)

	counter.Inc(ctx, metrics.Fields{
		"operation": operation,
		"status":    status,
	})

	// Validation operation duration histogram
	histogram := s.metrics.Histogram(
		"abac_attribute_validation_operation_duration_seconds",
		"Duration of attribute validation operations",
		metrics.StandardHTTPDurationBuckets(),
		"operation", "status",
	)

	histogram.Observe(ctx, duration.Seconds(), metrics.Fields{
		"operation": operation,
		"status":    status,
	})
}
