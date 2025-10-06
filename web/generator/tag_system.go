package generator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TagSystem provides comprehensive tag-based customization for UI generation
type TagSystem struct {
	parsers    map[string]TagParser
	validators map[string]TagValidator
	config     TagSystemConfig
}

// TagSystemConfig configures the tag system behavior
type TagSystemConfig struct {
	DefaultComponent      string            `json:"defaultComponent"`
	CustomTagPrefixes     []string          `json:"customTagPrefixes"`
	StrictValidation      bool              `json:"strictValidation"`
	AllowUnknownTags      bool              `json:"allowUnknownTags"`
	ComponentMappings     map[string]string `json:"componentMappings"`
	ValidationRuleMappings map[string]string `json:"validationRuleMappings"`
}

// TagParser defines the interface for parsing specific tag types
type TagParser interface {
	Parse(tagValue string) (map[string]interface{}, error)
	GetSupportedKeys() []string
}

// TagValidator validates tag configurations
type TagValidator interface {
	Validate(tagData map[string]interface{}) []ValidationIssue
}

// ValidationIssue represents a tag validation problem
type ValidationIssue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Level   string `json:"level"` // "error", "warning", "info"
}

// NewTagSystem creates a new tag system with default configuration
func NewTagSystem(config TagSystemConfig) *TagSystem {
	ts := &TagSystem{
		parsers:    make(map[string]TagParser),
		validators: make(map[string]TagValidator),
		config:     config,
	}
	
	// Register default parsers
	ts.registerDefaultParsers()
	
	// Register default validators
	ts.registerDefaultValidators()
	
	return ts
}

// ParseFieldTags parses all tags on a field and returns structured data
func (ts *TagSystem) ParseFieldTags(field reflect.StructField) (*FieldTagData, error) {
	tagData := &FieldTagData{
		FieldName: field.Name,
		FieldType: field.Type.String(),
		Tags:      make(map[string]map[string]interface{}),
		Issues:    []ValidationIssue{},
	}
	
	// Parse each supported tag
	for tagName, parser := range ts.parsers {
		tagValue := field.Tag.Get(tagName)
		if tagValue == "" {
			continue
		}
		
		parsed, err := parser.Parse(tagValue)
		if err != nil {
			tagData.Issues = append(tagData.Issues, ValidationIssue{
				Field:   fmt.Sprintf("%s.%s", field.Name, tagName),
				Message: fmt.Sprintf("Failed to parse %s tag: %v", tagName, err),
				Code:    "PARSE_ERROR",
				Level:   "error",
			})
			continue
		}
		
		tagData.Tags[tagName] = parsed
		
		// Validate parsed tag data
		if validator, exists := ts.validators[tagName]; exists {
			if issues := validator.Validate(parsed); len(issues) > 0 {
				tagData.Issues = append(tagData.Issues, issues...)
			}
		}
	}
	
	return tagData, nil
}

// ApplyTagCustomizations applies tag-based customizations to field info
func (ts *TagSystem) ApplyTagCustomizations(fieldInfo *FieldInfo, tagData *FieldTagData) error {
	// Apply UI tag customizations
	if uiTags, exists := tagData.Tags["ui"]; exists {
		ts.applyUITagCustomizations(fieldInfo, uiTags)
	}
	
	// Apply validation tag customizations
	if validateTags, exists := tagData.Tags["validate"]; exists {
		ts.applyValidationTagCustomizations(fieldInfo, validateTags)
	}
	
	// Apply database tag customizations
	if dbTags, exists := tagData.Tags["db"]; exists {
		ts.applyDatabaseTagCustomizations(fieldInfo, dbTags)
	}
	
	// Apply JSON tag customizations
	if jsonTags, exists := tagData.Tags["json"]; exists {
		ts.applyJSONTagCustomizations(fieldInfo, jsonTags)
	}
	
	// Apply custom tag prefixes
	for _, prefix := range ts.config.CustomTagPrefixes {
		if customTags, exists := tagData.Tags[prefix]; exists {
			ts.applyCustomTagCustomizations(fieldInfo, prefix, customTags)
		}
	}
	
	return nil
}

// FieldTagData holds parsed tag information for a field
type FieldTagData struct {
	FieldName string                            `json:"fieldName"`
	FieldType string                            `json:"fieldType"`
	Tags      map[string]map[string]interface{} `json:"tags"`
	Issues    []ValidationIssue                 `json:"issues"`
}

// registerDefaultParsers sets up built-in tag parsers
func (ts *TagSystem) registerDefaultParsers() {
	ts.parsers["ui"] = &UITagParser{}
	ts.parsers["validate"] = &ValidateTagParser{}
	ts.parsers["db"] = &DatabaseTagParser{}
	ts.parsers["json"] = &JSONTagParser{}
	ts.parsers["form"] = &FormTagParser{}
	ts.parsers["table"] = &TableTagParser{}
	ts.parsers["filter"] = &FilterTagParser{}
}

// registerDefaultValidators sets up built-in tag validators
func (ts *TagSystem) registerDefaultValidators() {
	ts.validators["ui"] = &UITagValidator{}
	ts.validators["validate"] = &ValidateTagValidator{}
	ts.validators["form"] = &FormTagValidator{}
}

// UI Tag Parser
type UITagParser struct{}

func (p *UITagParser) Parse(tagValue string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	if tagValue == "" {
		return result, nil
	}
	
	// Parse key=value pairs separated by semicolons
	pairs := strings.Split(tagValue, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Parse specific types
		switch key {
		case "component":
			result["component"] = value
		case "label":
			result["label"] = value
		case "required":
			result["required"] = value == "true"
		case "readonly":
			result["readonly"] = value == "true"
		case "hidden":
			result["hidden"] = value == "true"
		case "disabled":
			result["disabled"] = value == "true"
		case "placeholder":
			result["placeholder"] = value
		case "help":
			result["help"] = value
		case "tooltip":
			result["tooltip"] = value
		case "icon":
			result["icon"] = value
		case "size":
			result["size"] = value
		case "width":
			result["width"] = value
		case "height":
			result["height"] = value
		case "options":
			result["options"] = parseOptions(value)
		case "min", "max":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				result[key] = num
			}
		case "minLength", "maxLength":
			if num, err := strconv.Atoi(value); err == nil {
				result[key] = num
			}
		case "pattern":
			result["pattern"] = value
		case "format":
			result["format"] = value
		case "validation":
			result["validation"] = value
		case "relation":
			result["relation"] = value
		case "display":
			result["display"] = value
		case "sort":
			result["sort"] = value == "true"
		case "filter":
			result["filter"] = value == "true"
		case "search":
			result["search"] = value == "true"
		case "group":
			result["group"] = value
		case "section":
			result["section"] = value
		case "order":
			if num, err := strconv.Atoi(value); err == nil {
				result["order"] = num
			}
		case "colspan":
			if num, err := strconv.Atoi(value); err == nil {
				result["colspan"] = num
			}
		case "class":
			result["class"] = value
		case "style":
			result["style"] = value
		case "attrs":
			result["attrs"] = parseAttributes(value)
		default:
			// Allow unknown attributes (could be extended with configuration later)
			result[key] = value
		}
	}
	
	return result, nil
}

func (p *UITagParser) GetSupportedKeys() []string {
	return []string{
		"component", "label", "required", "readonly", "hidden", "disabled",
		"placeholder", "help", "tooltip", "icon", "size", "width", "height",
		"options", "min", "max", "minLength", "maxLength", "pattern", "format",
		"validation", "relation", "display", "sort", "filter", "search",
		"group", "section", "order", "colspan", "class", "style", "attrs",
	}
}

// Validation Tag Parser
type ValidateTagParser struct{}

func (p *ValidateTagParser) Parse(tagValue string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	if tagValue == "" {
		return result, nil
	}
	
	// Parse validation rules separated by commas
	rules := strings.Split(tagValue, ",")
	var validationRules []map[string]interface{}
	
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		
		// Parse rule with optional parameters
		parts := strings.SplitN(rule, "=", 2)
		ruleName := strings.TrimSpace(parts[0])
		ruleValue := ""
		if len(parts) > 1 {
			ruleValue = strings.TrimSpace(parts[1])
		}
		
		ruleData := map[string]interface{}{
			"rule": ruleName,
		}
		
		// Parse rule-specific parameters
		switch ruleName {
		case "required":
			ruleData["required"] = true
		case "email":
			ruleData["isEmail"] = true
		case "url":
			ruleData["isUrl"] = true
		case "numeric":
			ruleData["isNumeric"] = true
		case "alpha":
			ruleData["isAlpha"] = true
		case "alphanum":
			ruleData["isAlphaNum"] = true
		case "min":
			if val, err := strconv.ParseFloat(ruleValue, 64); err == nil {
				ruleData["min"] = val
			}
		case "max":
			if val, err := strconv.ParseFloat(ruleValue, 64); err == nil {
				ruleData["max"] = val
			}
		case "len":
			if val, err := strconv.Atoi(ruleValue); err == nil {
				ruleData["length"] = val
			}
		case "minlen":
			if val, err := strconv.Atoi(ruleValue); err == nil {
				ruleData["minLength"] = val
			}
		case "maxlen":
			if val, err := strconv.Atoi(ruleValue); err == nil {
				ruleData["maxLength"] = val
			}
		case "regex":
			ruleData["pattern"] = ruleValue
		case "oneof":
			ruleData["oneOf"] = strings.Split(ruleValue, " ")
		default:
			if ruleValue != "" {
				ruleData["value"] = ruleValue
			}
		}
		
		validationRules = append(validationRules, ruleData)
	}
	
	result["rules"] = validationRules
	return result, nil
}

func (p *ValidateTagParser) GetSupportedKeys() []string {
	return []string{
		"required", "email", "url", "numeric", "alpha", "alphanum",
		"min", "max", "len", "minlen", "maxlen", "regex", "oneof",
	}
}

// Database Tag Parser
type DatabaseTagParser struct{}

func (p *DatabaseTagParser) Parse(tagValue string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	if tagValue == "" {
		return result, nil
	}
	
	// Handle different database tag formats
	if tagValue == "-" {
		result["ignored"] = true
		return result, nil
	}
	
	// Parse comma-separated options
	parts := strings.Split(tagValue, ",")
	if len(parts) > 0 {
		result["column"] = strings.TrimSpace(parts[0])
	}
	
	for i := 1; i < len(parts); i++ {
		option := strings.TrimSpace(parts[i])
		switch option {
		case "primarykey", "primary_key":
			result["primaryKey"] = true
		case "unique":
			result["unique"] = true
		case "index":
			result["index"] = true
		case "not null", "notnull":
			result["notNull"] = true
		case "autoincrement", "auto_increment":
			result["autoIncrement"] = true
		default:
			// Store unknown options
			if options, exists := result["options"]; exists {
				result["options"] = append(options.([]string), option)
			} else {
				result["options"] = []string{option}
			}
		}
	}
	
	return result, nil
}

func (p *DatabaseTagParser) GetSupportedKeys() []string {
	return []string{
		"column", "primaryKey", "unique", "index", "notNull", "autoIncrement", "options",
	}
}

// JSON Tag Parser
type JSONTagParser struct{}

func (p *JSONTagParser) Parse(tagValue string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	if tagValue == "" {
		return result, nil
	}
	
	if tagValue == "-" {
		result["ignored"] = true
		return result, nil
	}
	
	// Parse comma-separated options
	parts := strings.Split(tagValue, ",")
	if len(parts) > 0 && parts[0] != "" {
		result["name"] = strings.TrimSpace(parts[0])
	}
	
	for i := 1; i < len(parts); i++ {
		option := strings.TrimSpace(parts[i])
		switch option {
		case "omitempty":
			result["omitEmpty"] = true
		case "string":
			result["string"] = true
		default:
			// Store unknown options
			if options, exists := result["options"]; exists {
				result["options"] = append(options.([]string), option)
			} else {
				result["options"] = []string{option}
			}
		}
	}
	
	return result, nil
}

func (p *JSONTagParser) GetSupportedKeys() []string {
	return []string{"name", "omitEmpty", "string", "options"}
}

// Form Tag Parser
type FormTagParser struct{}

func (p *FormTagParser) Parse(tagValue string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// Parse form-specific configurations
	pairs := strings.Split(tagValue, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "tab":
			result["tab"] = value
		case "section":
			result["section"] = value
		case "inline":
			result["inline"] = value == "true"
		case "cols":
			if num, err := strconv.Atoi(value); err == nil {
				result["cols"] = num
			}
		case "rows":
			if num, err := strconv.Atoi(value); err == nil {
				result["rows"] = num
			}
		case "dependsOn":
			result["dependsOn"] = value
		case "showWhen":
			result["showWhen"] = value
		case "hideWhen":
			result["hideWhen"] = value
		default:
			result[key] = value
		}
	}
	
	return result, nil
}

func (p *FormTagParser) GetSupportedKeys() []string {
	return []string{
		"tab", "section", "inline", "cols", "rows", "dependsOn", "showWhen", "hideWhen",
	}
}

// Table Tag Parser
type TableTagParser struct{}

func (p *TableTagParser) Parse(tagValue string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// Parse table-specific configurations
	pairs := strings.Split(tagValue, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "width":
			result["width"] = value
		case "align":
			result["align"] = value
		case "sortable":
			result["sortable"] = value == "true"
		case "filterable":
			result["filterable"] = value == "true"
		case "searchable":
			result["searchable"] = value == "true"
		case "resizable":
			result["resizable"] = value == "true"
		case "fixed":
			result["fixed"] = value
		case "priority":
			if num, err := strconv.Atoi(value); err == nil {
				result["priority"] = num
			}
		case "breakpoint":
			result["breakpoint"] = value
		default:
			result[key] = value
		}
	}
	
	return result, nil
}

func (p *TableTagParser) GetSupportedKeys() []string {
	return []string{
		"width", "align", "sortable", "filterable", "searchable", "resizable",
		"fixed", "priority", "breakpoint",
	}
}

// Filter Tag Parser
type FilterTagParser struct{}

func (p *FilterTagParser) Parse(tagValue string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// Parse filter-specific configurations
	pairs := strings.Split(tagValue, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "type":
			result["type"] = value
		case "operator":
			result["operator"] = value
		case "multiple":
			result["multiple"] = value == "true"
		case "preset":
			result["preset"] = value == "true"
		case "presetValue":
			result["presetValue"] = value
		case "range":
			result["range"] = value == "true"
		case "format":
			result["format"] = value
		default:
			result[key] = value
		}
	}
	
	return result, nil
}

func (p *FilterTagParser) GetSupportedKeys() []string {
	return []string{
		"type", "operator", "multiple", "preset", "presetValue", "range", "format",
	}
}

// Tag Validators

// UI Tag Validator
type UITagValidator struct{}

func (v *UITagValidator) Validate(tagData map[string]interface{}) []ValidationIssue {
	var issues []ValidationIssue
	
	// Validate component type
	if component, exists := tagData["component"]; exists {
		if componentStr, ok := component.(string); ok {
			if !isValidComponent(componentStr) {
				issues = append(issues, ValidationIssue{
					Field:   "component",
					Message: fmt.Sprintf("Unknown component type: %s", componentStr),
					Code:    "INVALID_COMPONENT",
					Level:   "warning",
				})
			}
		}
	}
	
	// Validate numeric ranges
	if min, minExists := tagData["min"]; minExists {
		if max, maxExists := tagData["max"]; maxExists {
			if minVal, minOk := min.(float64); minOk {
				if maxVal, maxOk := max.(float64); maxOk {
					if minVal > maxVal {
						issues = append(issues, ValidationIssue{
							Field:   "min/max",
							Message: "Minimum value cannot be greater than maximum value",
							Code:    "INVALID_RANGE",
							Level:   "error",
						})
					}
				}
			}
		}
	}
	
	// Validate length constraints
	if minLen, minExists := tagData["minLength"]; minExists {
		if maxLen, maxExists := tagData["maxLength"]; maxExists {
			if minVal, minOk := minLen.(int); minOk {
				if maxVal, maxOk := maxLen.(int); maxOk {
					if minVal > maxVal {
						issues = append(issues, ValidationIssue{
							Field:   "minLength/maxLength",
							Message: "Minimum length cannot be greater than maximum length",
							Code:    "INVALID_LENGTH_RANGE",
							Level:   "error",
						})
					}
				}
			}
		}
	}
	
	return issues
}

// Validate Tag Validator
type ValidateTagValidator struct{}

func (v *ValidateTagValidator) Validate(tagData map[string]interface{}) []ValidationIssue {
	var issues []ValidationIssue
	
	if rules, exists := tagData["rules"]; exists {
		if rulesList, ok := rules.([]map[string]interface{}); ok {
			for _, rule := range rulesList {
				if ruleName, exists := rule["rule"]; exists {
					if ruleNameStr, ok := ruleName.(string); ok {
						if !isValidValidationRule(ruleNameStr) {
							issues = append(issues, ValidationIssue{
								Field:   "rule",
								Message: fmt.Sprintf("Unknown validation rule: %s", ruleNameStr),
								Code:    "INVALID_VALIDATION_RULE",
								Level:   "warning",
							})
						}
					}
				}
			}
		}
	}
	
	return issues
}

// Form Tag Validator
type FormTagValidator struct{}

func (v *FormTagValidator) Validate(tagData map[string]interface{}) []ValidationIssue {
	var issues []ValidationIssue
	
	// Validate conditional dependencies
	if dependsOn, exists := tagData["dependsOn"]; exists {
		if dependsOnStr, ok := dependsOn.(string); ok {
			if dependsOnStr == "" {
				issues = append(issues, ValidationIssue{
					Field:   "dependsOn",
					Message: "dependsOn cannot be empty",
					Code:    "EMPTY_DEPENDENCY",
					Level:   "error",
				})
			}
		}
	}
	
	return issues
}

// Helper methods for applying customizations

func (ts *TagSystem) applyUITagCustomizations(fieldInfo *FieldInfo, uiTags map[string]interface{}) {
	if component, exists := uiTags["component"]; exists {
		if componentStr, ok := component.(string); ok {
			fieldInfo.Component = componentStr
		}
	}
	
	if label, exists := uiTags["label"]; exists {
		if labelStr, ok := label.(string); ok {
			fieldInfo.Label = labelStr
		}
	}
	
	if required, exists := uiTags["required"]; exists {
		if requiredBool, ok := required.(bool); ok {
			fieldInfo.Required = requiredBool
		}
	}
	
	if placeholder, exists := uiTags["placeholder"]; exists {
		if placeholderStr, ok := placeholder.(string); ok {
			fieldInfo.Placeholder = placeholderStr
		}
	}
	
	if hidden, exists := uiTags["hidden"]; exists {
		if hiddenBool, ok := hidden.(bool); ok {
			fieldInfo.Hidden = hiddenBool
		}
	}
	
	if options, exists := uiTags["options"]; exists {
		if optionsList, ok := options.([]string); ok {
			fieldInfo.Options = optionsList
		}
	}
	
	// Apply additional UI customizations...
}

func (ts *TagSystem) applyValidationTagCustomizations(fieldInfo *FieldInfo, validateTags map[string]interface{}) {
	// Apply validation rules to fieldInfo
	// This would convert parsed validation rules to the fieldInfo validation format
}

func (ts *TagSystem) applyDatabaseTagCustomizations(fieldInfo *FieldInfo, dbTags map[string]interface{}) {
	// Apply database-related customizations
	// This might affect how relationships are detected or database constraints are handled
}

func (ts *TagSystem) applyJSONTagCustomizations(fieldInfo *FieldInfo, jsonTags map[string]interface{}) {
	// Apply JSON serialization customizations
	// This might affect field naming in generated schemas
}

func (ts *TagSystem) applyCustomTagCustomizations(fieldInfo *FieldInfo, prefix string, customTags map[string]interface{}) {
	// Apply custom tag prefix customizations
	// This allows for domain-specific tag extensions
}

// Utility functions

func parseOptions(value string) []string {
	if value == "" {
		return nil
	}
	
	options := strings.Split(value, ",")
	for i, option := range options {
		options[i] = strings.TrimSpace(option)
	}
	
	return options
}

func parseAttributes(value string) map[string]string {
	attrs := make(map[string]string)
	if value == "" {
		return attrs
	}
	
	// Parse key:value pairs separated by commas
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			attrs[key] = val
		}
	}
	
	return attrs
}

func isValidComponent(component string) bool {
	validComponents := []string{
		"text", "textarea", "select", "multi-select", "checkbox", "radio",
		"toggle", "email", "password", "url", "tel", "number", "date",
		"datetime", "time", "file", "image", "rich-text", "hidden",
	}
	
	for _, valid := range validComponents {
		if component == valid {
			return true
		}
	}
	
	return false
}

func isValidValidationRule(rule string) bool {
	validRules := []string{
		"required", "email", "url", "numeric", "alpha", "alphanum",
		"min", "max", "len", "minlen", "maxlen", "regex", "oneof",
	}
	
	for _, valid := range validRules {
		if rule == valid {
			return true
		}
	}
	
	return false
}

// RegisterCustomParser allows registering custom tag parsers
func (ts *TagSystem) RegisterCustomParser(tagName string, parser TagParser) {
	ts.parsers[tagName] = parser
}

// RegisterCustomValidator allows registering custom tag validators
func (ts *TagSystem) RegisterCustomValidator(tagName string, validator TagValidator) {
	ts.validators[tagName] = validator
}

// GetSupportedTags returns a list of all supported tag names
func (ts *TagSystem) GetSupportedTags() []string {
	var tags []string
	for tagName := range ts.parsers {
		tags = append(tags, tagName)
	}
	return tags
}

// TODO: Future enhancements for tag system:
// 1. Dynamic tag registration system
// 2. Tag inheritance and composition
// 3. Conditional tag processing based on context
// 4. Tag preprocessing and transformation pipelines
// 5. Integration with external validation libraries
// 6. Tag-based code generation templates
// 7. Tag documentation and auto-completion support
// 8. Performance optimization for large-scale tag processing

// NOTE: Design considerations:
// 1. The tag system is extensible through custom parsers and validators
// 2. Tag parsing supports multiple formats (key=value, comma-separated, etc.)
// 3. Validation provides different severity levels (error, warning, info)
// 4. The system can be configured for strict or lenient tag validation
// 5. Custom tag prefixes allow domain-specific extensions