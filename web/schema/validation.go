package schema

import (
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Validator provides comprehensive validation for schemas and data
type Validator struct {
	customValidators map[string]CustomValidatorFunc
	asyncValidators  map[string]AsyncValidatorFunc
	errorHandler     *ErrorHandler
}

// CustomValidatorFunc defines custom validation function signature
type CustomValidatorFunc func(ctx context.Context, value interface{}, params map[string]interface{}) error

// AsyncValidatorFunc defines async validation function signature
type AsyncValidatorFunc func(ctx context.Context, value interface{}, params map[string]interface{}) error

// ValidationResult holds validation results
type ValidationResult struct {
	Valid       bool                   `json:"valid"`
	Errors      []SchemaError          `json:"errors,omitempty"`
	FieldErrors map[string][]string    `json:"fieldErrors,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		customValidators: make(map[string]CustomValidatorFunc),
		asyncValidators:  make(map[string]AsyncValidatorFunc),
		errorHandler:     NewErrorHandler(nil), // Inject logger as needed
	}
}

// RegisterCustomValidator registers a custom validation function
func (v *Validator) RegisterCustomValidator(name string, fn CustomValidatorFunc) {
	v.customValidators[name] = fn
}

// RegisterAsyncValidator registers an async validation function
func (v *Validator) RegisterAsyncValidator(name string, fn AsyncValidatorFunc) {
	v.asyncValidators[name] = fn
}

// ValidateSchema validates the schema structure itself
func (v *Validator) ValidateSchema(ctx context.Context, schema *Schema) *ValidationResult {
	collector := NewErrorCollector()
	
	// Core schema validation
	v.validateSchemaCore(schema, collector)
	
	// Validate fields
	v.validateSchemaFields(schema, collector)
	
	// Validate actions
	v.validateSchemaActions(schema, collector)
	
	// Validate layout
	v.validateSchemaLayout(schema, collector)
	
	// Validate enterprise features
	v.validateSchemaEnterpriseFeatures(schema, collector)
	
	// Validate framework integration
	v.validateSchemaFrameworkIntegration(schema, collector)
	
	return &ValidationResult{
		Valid:       !collector.HasErrors(),
		Errors:      collector.Errors().Errors(),
		FieldErrors: collector.Errors().ErrorsByField(),
	}
}

// ValidateData validates form data against schema
func (v *Validator) ValidateData(ctx context.Context, schema *Schema, data map[string]interface{}) *ValidationResult {
	collector := NewErrorCollector()
	
	// Validate each field's data
	for _, field := range schema.Fields {
		v.validateFieldData(ctx, &field, data, collector)
	}
	
	// Validate cross-field rules
	if schema.Validation != nil {
		v.validateCrossFieldRules(ctx, schema.Validation.CrossFieldRules, data, collector)
	}
	
	// Validate custom rules
	if schema.Validation != nil {
		v.validateCustomRules(ctx, schema.Validation.CustomRules, data, collector)
	}
	
	return &ValidationResult{
		Valid:       !collector.HasErrors(),
		Errors:      collector.Errors().Errors(),
		FieldErrors: collector.Errors().ErrorsByField(),
	}
}

// Schema structure validation methods
func (v *Validator) validateSchemaCore(schema *Schema, collector *ErrorCollector) {
	if schema.ID == "" {
		collector.AddError(ErrInvalidSchemaID)
	} else if !v.isValidID(schema.ID) {
		collector.AddError(ErrInvalidSchemaID.WithDetail("pattern", "must be alphanumeric with hyphens"))
	}
	
	if schema.Type == "" {
		collector.AddError(ErrInvalidSchemaType)
	}
	
	if schema.Title == "" {
		collector.AddError(ErrInvalidSchemaTitle)
	} else if utf8.RuneCountInString(schema.Title) > 200 {
		collector.AddError(ErrInvalidSchemaTitle.WithDetail("maxLength", 200))
	}
	
	if schema.Version != "" && !v.isValidSemVer(schema.Version) {
		collector.AddValidationError("version", "invalid_semver", "version must be valid semantic version")
	}
}

func (v *Validator) validateSchemaFields(schema *Schema, collector *ErrorCollector) {
	fieldNames := make(map[string]bool)
	
	for i, field := range schema.Fields {
		fieldPath := fmt.Sprintf("fields[%d]", i)
		
		// Validate field structure
		v.validateField(&field, fieldPath, collector)
		
		// Check for duplicate field names
		if field.Name != "" {
			if fieldNames[field.Name] {
				collector.AddFieldError(fieldPath+".name", ErrFieldDuplicate)
			}
			fieldNames[field.Name] = true
		}
	}
}

func (v *Validator) validateField(field *Field, path string, collector *ErrorCollector) {
	if field.Name == "" {
		collector.AddFieldError(path+".name", ErrInvalidFieldName)
	} else if !v.isValidFieldName(field.Name) {
		collector.AddFieldError(path+".name", ErrInvalidFieldName.WithDetail("pattern", "must be valid identifier"))
	}
	
	if field.Type == "" {
		collector.AddFieldError(path+".type", ErrInvalidFieldType)
	} else if !v.isValidFieldType(field.Type) {
		collector.AddFieldError(path+".type", ErrInvalidFieldType.WithDetail("allowed", v.getAllowedFieldTypes()))
	}
	
	if field.Label == "" {
		collector.AddFieldError(path+".label", ErrInvalidFieldLabel)
	}
	
	// Validate field validation rules
	if field.Validation != nil {
		v.validateFieldValidationRules(field.Validation, path+".validation", collector)
	}
	
	// Validate options for select-type fields
	if v.fieldRequiresOptions(field.Type) && len(field.Options) == 0 && field.DataSource == nil {
		collector.AddFieldError(path, NewValidationError("options_required", "field type requires options or data source"))
	}
	
	// Validate data source
	if field.DataSource != nil {
		v.validateDataSource(field.DataSource, path+".dataSource", collector)
	}
}

func (v *Validator) validateSchemaActions(schema *Schema, collector *ErrorCollector) {
	actionIDs := make(map[string]bool)
	
	for i, action := range schema.Actions {
		actionPath := fmt.Sprintf("actions[%d]", i)
		
		if action.ID == "" {
			collector.AddFieldError(actionPath+".id", NewValidationError("action_id_required", "action ID is required"))
		} else if actionIDs[action.ID] {
			collector.AddFieldError(actionPath+".id", NewValidationError("action_id_duplicate", "action ID must be unique"))
		} else {
			actionIDs[action.ID] = true
		}
		
		if action.Text == "" {
			collector.AddFieldError(actionPath+".text", NewValidationError("action_text_required", "action text is required"))
		}
		
		if action.Type == "" {
			collector.AddFieldError(actionPath+".type", NewValidationError("action_type_required", "action type is required"))
		}
	}
}

func (v *Validator) validateSchemaLayout(schema *Schema, collector *ErrorCollector) {
	if schema.Layout == nil {
		return
	}
	
	layout := schema.Layout
	layoutPath := "layout"
	
	if layout.Columns < 1 || layout.Columns > 24 {
		collector.AddFieldError(layoutPath+".columns", NewValidationError("invalid_columns", "columns must be between 1 and 24"))
	}
	
	// Validate sections
	sectionIDs := make(map[string]bool)
	for i, section := range layout.Sections {
		sectionPath := fmt.Sprintf("%s.sections[%d]", layoutPath, i)
		
		if section.ID == "" {
			collector.AddFieldError(sectionPath+".id", NewValidationError("section_id_required", "section ID is required"))
		} else if sectionIDs[section.ID] {
			collector.AddFieldError(sectionPath+".id", NewValidationError("section_id_duplicate", "section ID must be unique"))
		} else {
			sectionIDs[section.ID] = true
		}
		
		if len(section.Fields) == 0 {
			collector.AddFieldError(sectionPath+".fields", NewValidationError("section_fields_required", "section must have at least one field"))
		}
	}
}

func (v *Validator) validateSchemaEnterpriseFeatures(schema *Schema, collector *ErrorCollector) {
	// Validate tenant configuration
	if schema.Tenant != nil && schema.Tenant.Enabled {
		if schema.Tenant.Field == "" {
			collector.AddFieldError("tenant.field", NewValidationError("tenant_field_required", "tenant field is required when tenant is enabled"))
		}
		
		if schema.Tenant.Isolation == "" {
			collector.AddFieldError("tenant.isolation", NewValidationError("tenant_isolation_required", "tenant isolation is required"))
		} else if !v.isValidTenantIsolation(schema.Tenant.Isolation) {
			collector.AddFieldError("tenant.isolation", NewValidationError("invalid_tenant_isolation", "invalid tenant isolation type"))
		}
	}
	
	// Validate workflow configuration
	if schema.Workflow != nil && schema.Workflow.Enabled {
		if schema.Workflow.ID == "" {
			collector.AddFieldError("workflow.id", NewValidationError("workflow_id_required", "workflow ID is required"))
		}
	}
	
	// Validate security configuration
	if schema.Security != nil {
		if schema.Security.CSRF != nil && schema.Security.CSRF.Enabled {
			if schema.Security.CSRF.TokenField == "" {
				collector.AddFieldError("security.csrf.tokenField", NewValidationError("csrf_token_field_required", "CSRF token field is required"))
			}
		}
	}
}

func (v *Validator) validateSchemaFrameworkIntegration(schema *Schema, collector *ErrorCollector) {
	// Validate HTMX configuration
	if schema.HTMX != nil && schema.HTMX.Enabled {
		if schema.HTMX.Target != "" && !v.isValidCSSSelector(schema.HTMX.Target) {
			collector.AddFieldError("htmx.target", NewValidationError("invalid_css_selector", "invalid CSS selector"))
		}
		
		if schema.HTMX.Get != "" && !v.isValidURL(schema.HTMX.Get) {
			collector.AddFieldError("htmx.get", NewValidationError("invalid_url", "invalid URL"))
		}
		
		if schema.HTMX.Post != "" && !v.isValidURL(schema.HTMX.Post) {
			collector.AddFieldError("htmx.post", NewValidationError("invalid_url", "invalid URL"))
		}
	}
	
	// Validate Alpine.js configuration
	if schema.Alpine != nil && schema.Alpine.Enabled {
		if schema.Alpine.XData != "" && !v.isValidJSObject(schema.Alpine.XData) {
			collector.AddFieldError("alpine.xData", NewValidationError("invalid_js_object", "invalid JavaScript object"))
		}
	}
}

// Field data validation methods
func (v *Validator) validateFieldData(ctx context.Context, field *Field, data map[string]interface{}, collector *ErrorCollector) {
	value, exists := data[field.Name]
	
	// Check if field is required
	isRequired := field.Required
	if field.Conditional != nil {
		// TODO: Evaluate conditional requirements
	}
	
	// Required field validation
	if isRequired && (!exists || v.isEmpty(value)) {
		collector.AddFieldError(field.Name, ErrRequiredField)
		return
	}
	
	// Skip validation if field is empty and not required
	if !exists || v.isEmpty(value) {
		return
	}
	
	// Type-specific validation
	if err := v.validateFieldValueByType(field.Type, value); err != nil {
		collector.AddFieldError(field.Name, err)
		return
	}
	
	// Validation rules
	if field.Validation != nil {
		v.validateFieldValueByRules(ctx, field, value, collector)
	}
}

func (v *Validator) validateFieldValueByType(fieldType FieldType, value interface{}) SchemaError {
	switch fieldType {
	case FieldEmail:
		return v.validateEmail(value)
	case FieldURL:
		return v.validateURL(value)
	case FieldNumber, FieldCurrency:
		return v.validateNumber(value)
	case FieldPhone:
		return v.validatePhone(value)
	case FieldDate, FieldDateTime:
		return v.validateDate(value)
	case FieldJSON:
		return v.validateJSON(value)
	default:
		return nil
	}
}

func (v *Validator) validateFieldValueByRules(ctx context.Context, field *Field, value interface{}, collector *ErrorCollector) {
	rules := field.Validation
	
	// Length validation
	if rules.MinLength != nil || rules.MaxLength != nil {
		if err := v.validateLength(value, rules.MinLength, rules.MaxLength); err != nil {
			collector.AddFieldError(field.Name, err)
		}
	}
	
	// Range validation
	if rules.Min != nil || rules.Max != nil {
		if err := v.validateRange(value, rules.Min, rules.Max); err != nil {
			collector.AddFieldError(field.Name, err)
		}
	}
	
	// Pattern validation
	if rules.Pattern != "" {
		if err := v.validatePattern(value, rules.Pattern); err != nil {
			collector.AddFieldError(field.Name, err)
		}
	}
	
	// Custom validations
	for _, customRule := range rules.Custom {
		if validator, exists := v.customValidators[customRule]; exists {
			if err := validator(ctx, value, nil); err != nil {
				collector.AddFieldError(field.Name, WrapError(err, "custom_validation", "custom validation failed"))
			}
		}
	}
	
	// Async validation (would be handled separately in practice)
	if rules.Async != nil {
		// TODO: Implement async validation
	}
}

func (v *Validator) validateCrossFieldRules(ctx context.Context, rules []CrossFieldRule, data map[string]interface{}, collector *ErrorCollector) {
	for _, rule := range rules {
		// TODO: Implement cross-field validation using condition package
		// This would evaluate the conditions and formulas against the data
	}
}

func (v *Validator) validateCustomRules(ctx context.Context, rules []CustomRule, data map[string]interface{}, collector *ErrorCollector) {
	for _, rule := range rules {
		if validator, exists := v.customValidators[rule.Name]; exists {
			if err := validator(ctx, data, nil); err != nil {
				collector.AddError(WrapError(err, rule.ID, rule.Message))
			}
		}
	}
}

// Specific validation helper methods
func (v *Validator) validateEmail(value interface{}) SchemaError {
	str, ok := value.(string)
	if !ok {
		return NewValidationError("invalid_type", "value must be a string")
	}
	
	if _, err := mail.ParseAddress(str); err != nil {
		return NewValidationError("invalid_email", "invalid email format")
	}
	
	return nil
}

func (v *Validator) validateURL(value interface{}) SchemaError {
	str, ok := value.(string)
	if !ok {
		return NewValidationError("invalid_type", "value must be a string")
	}
	
	if _, err := url.Parse(str); err != nil {
		return NewValidationError("invalid_url", "invalid URL format")
	}
	
	return nil
}

func (v *Validator) validateNumber(value interface{}) SchemaError {
	switch value.(type) {
	case int, int32, int64, float32, float64:
		return nil
	case string:
		str := value.(string)
		if _, err := strconv.ParseFloat(str, 64); err != nil {
			return NewValidationError("invalid_number", "invalid number format")
		}
		return nil
	default:
		return NewValidationError("invalid_type", "value must be a number")
	}
}

func (v *Validator) validatePhone(value interface{}) SchemaError {
	str, ok := value.(string)
	if !ok {
		return NewValidationError("invalid_type", "value must be a string")
	}
	
	// Basic phone validation - can be enhanced with proper phone number library
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	cleanPhone := strings.ReplaceAll(str, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")
	
	if !phoneRegex.MatchString(cleanPhone) {
		return NewValidationError("invalid_phone", "invalid phone number format")
	}
	
	return nil
}

func (v *Validator) validateDate(value interface{}) SchemaError {
	str, ok := value.(string)
	if !ok {
		return NewValidationError("invalid_type", "value must be a string")
	}
	
	// Try multiple date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"01-02-2006",
	}
	
	for _, format := range formats {
		if _, err := time.Parse(format, str); err == nil {
			return nil
		}
	}
	
	return NewValidationError("invalid_date", "invalid date format")
}

func (v *Validator) validateJSON(value interface{}) SchemaError {
	str, ok := value.(string)
	if !ok {
		return NewValidationError("invalid_type", "value must be a string")
	}
	
	if !json.Valid([]byte(str)) {
		return NewValidationError("invalid_json", "invalid JSON format")
	}
	
	return nil
}

func (v *Validator) validateLength(value interface{}, minLength, maxLength *int) SchemaError {
	str := fmt.Sprintf("%v", value)
	length := utf8.RuneCountInString(str)
	
	if minLength != nil && length < *minLength {
		return NewValidationError("min_length", fmt.Sprintf("minimum length is %d", *minLength))
	}
	
	if maxLength != nil && length > *maxLength {
		return NewValidationError("max_length", fmt.Sprintf("maximum length is %d", *maxLength))
	}
	
	return nil
}

func (v *Validator) validateRange(value interface{}, min, max *float64) SchemaError {
	var numValue float64
	var err error
	
	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int32:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float32:
		numValue = float64(v)
	case float64:
		numValue = v
	case string:
		numValue, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return NewValidationError("invalid_number", "value must be a number")
		}
	default:
		return NewValidationError("invalid_type", "value must be a number")
	}
	
	if min != nil && numValue < *min {
		return NewValidationError("min_value", fmt.Sprintf("minimum value is %f", *min))
	}
	
	if max != nil && numValue > *max {
		return NewValidationError("max_value", fmt.Sprintf("maximum value is %f", *max))
	}
	
	return nil
}

func (v *Validator) validatePattern(value interface{}, pattern string) SchemaError {
	str := fmt.Sprintf("%v", value)
	
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return NewValidationError("invalid_pattern", "invalid regular expression pattern")
	}
	
	if !matched {
		return NewValidationError("pattern_mismatch", "value does not match required pattern")
	}
	
	return nil
}

// Helper validation methods
func (v *Validator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	
	str, ok := value.(string)
	if ok {
		return strings.TrimSpace(str) == ""
	}
	
	return false
}

func (v *Validator) isValidID(id string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`, id)
	return matched && len(id) <= 100
}

func (v *Validator) isValidFieldName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, name)
	return matched && len(name) <= 100
}

func (v *Validator) isValidFieldType(fieldType FieldType) bool {
	validTypes := map[FieldType]bool{
		FieldText: true, FieldEmail: true, FieldPassword: true, FieldNumber: true,
		FieldHidden: true, FieldDate: true, FieldTime: true, FieldDateTime: true,
		FieldDateRange: true, FieldTextarea: true, FieldRichText: true, FieldCode: true,
		FieldJSON: true, FieldSelect: true, FieldMultiSelect: true, FieldRadio: true,
		FieldCheckbox: true, FieldTreeSelect: true, FieldCascader: true, FieldTransfer: true,
		FieldSwitch: true, FieldSlider: true, FieldRating: true, FieldColor: true,
		FieldFile: true, FieldImage: true, FieldSignature: true, FieldPhone: true,
		FieldURL: true, FieldCurrency: true, FieldTags: true, FieldLocation: true,
		FieldRelation: true, FieldAutoComplete: true, FieldDisplay: true, FieldDivider: true,
		FieldHTML: true,
	}
	return validTypes[fieldType]
}

func (v *Validator) getAllowedFieldTypes() []string {
	return []string{
		string(FieldText), string(FieldEmail), string(FieldPassword), string(FieldNumber),
		string(FieldHidden), string(FieldDate), string(FieldTime), string(FieldDateTime),
		string(FieldDateRange), string(FieldTextarea), string(FieldRichText), string(FieldCode),
		string(FieldJSON), string(FieldSelect), string(FieldMultiSelect), string(FieldRadio),
		string(FieldCheckbox), string(FieldTreeSelect), string(FieldCascader), string(FieldTransfer),
		string(FieldSwitch), string(FieldSlider), string(FieldRating), string(FieldColor),
		string(FieldFile), string(FieldImage), string(FieldSignature), string(FieldPhone),
		string(FieldURL), string(FieldCurrency), string(FieldTags), string(FieldLocation),
		string(FieldRelation), string(FieldAutoComplete), string(FieldDisplay), string(FieldDivider),
		string(FieldHTML),
	}
}

func (v *Validator) fieldRequiresOptions(fieldType FieldType) bool {
	requiresOptions := map[FieldType]bool{
		FieldSelect:      true,
		FieldMultiSelect: true,
		FieldRadio:       true,
		FieldCheckbox:    true,
		FieldTreeSelect:  true,
		FieldCascader:    true,
	}
	return requiresOptions[fieldType]
}

func (v *Validator) isValidSemVer(version string) bool {
	semVerRegex := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	return semVerRegex.MatchString(version)
}

func (v *Validator) isValidTenantIsolation(isolation string) bool {
	validIsolations := map[string]bool{
		"strict": true,
		"shared": true,
		"hybrid": true,
	}
	return validIsolations[isolation]
}

func (v *Validator) isValidCSSSelector(selector string) bool {
	// Basic CSS selector validation - can be enhanced
	return strings.HasPrefix(selector, "#") || strings.HasPrefix(selector, ".") || 
		   regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`).MatchString(selector)
}

func (v *Validator) isValidURL(urlStr string) bool {
	_, err := url.Parse(urlStr)
	return err == nil && (strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") || strings.HasPrefix(urlStr, "/"))
}

func (v *Validator) isValidJSObject(obj string) bool {
	// Basic validation - in production would use a proper JS parser
	return strings.HasPrefix(obj, "{") && strings.HasSuffix(obj, "}")
}

func (v *Validator) validateFieldValidationRules(validation *FieldValidation, path string, collector *ErrorCollector) {
	if validation.MinLength != nil && *validation.MinLength < 0 {
		collector.AddFieldError(path+".minLength", NewValidationError("invalid_min_length", "minimum length cannot be negative"))
	}
	
	if validation.MaxLength != nil && *validation.MaxLength < 0 {
		collector.AddFieldError(path+".maxLength", NewValidationError("invalid_max_length", "maximum length cannot be negative"))
	}
	
	if validation.MinLength != nil && validation.MaxLength != nil && *validation.MinLength > *validation.MaxLength {
		collector.AddFieldError(path, NewValidationError("invalid_length_range", "minimum length cannot be greater than maximum length"))
	}
	
	if validation.Min != nil && validation.Max != nil && *validation.Min > *validation.Max {
		collector.AddFieldError(path, NewValidationError("invalid_value_range", "minimum value cannot be greater than maximum value"))
	}
	
	if validation.Pattern != "" {
		if _, err := regexp.Compile(validation.Pattern); err != nil {
			collector.AddFieldError(path+".pattern", NewValidationError("invalid_regex_pattern", "invalid regular expression pattern"))
		}
	}
}

func (v *Validator) validateDataSource(dataSource *DataSource, path string, collector *ErrorCollector) {
	if dataSource.Type == "" {
		collector.AddFieldError(path+".type", NewValidationError("data_source_type_required", "data source type is required"))
	}
	
	if dataSource.Type == DataSourceAPI {
		if dataSource.URL == "" {
			collector.AddFieldError(path+".url", NewValidationError("api_url_required", "API URL is required for API data source"))
		} else if !v.isValidURL(dataSource.URL) {
			collector.AddFieldError(path+".url", NewValidationError("invalid_api_url", "invalid API URL"))
		}
	}
	
	if dataSource.Type == DataSourceStatic && len(dataSource.Data) == 0 {
		collector.AddFieldError(path+".data", NewValidationError("static_data_required", "static data is required for static data source"))
	}
}