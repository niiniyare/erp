package css

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Validator provides CSS property validation functionality
type Validator struct {
	loader *SchemaLoader
}

// NewValidator creates a new CSS validator
func NewValidator(schemaDir string) *Validator {
	return &Validator{
		loader: NewSchemaLoader(schemaDir),
	}
}

// ValidateStyles validates a complete Styles object
func (v *Validator) ValidateStyles(styles *Styles) []ValidationError {
	var errors []ValidationError

	properties := styles.getAllProperties()

	for propertyName, value := range properties {
		if value == "" {
			continue
		}

		if err := v.ValidateProperty(propertyName, value); err != nil {
			errors = append(errors, ValidationError{
				Property: propertyName,
				Value:    value,
				Message:  err.Error(),
			})
		}
	}

	// Validate custom properties
	for propertyName, value := range styles.Custom {
		if value == "" {
			continue
		}

		if err := v.ValidateCustomProperty(propertyName, value); err != nil {
			errors = append(errors, ValidationError{
				Property: propertyName,
				Value:    value,
				Message:  err.Error(),
				IsCustom: true,
			})
		}
	}

	return errors
}

// ValidateProperty validates a single CSS property
func (v *Validator) ValidateProperty(propertyName, value string) error {
	// First try schema validation
	if err := v.loader.ValidateValue(propertyName, value); err == nil {
		return nil
	}

	// Fallback to pattern-based validation for common properties
	return v.validateWithPatterns(propertyName, value)
}

// ValidateCustomProperty validates custom CSS properties
func (v *Validator) ValidateCustomProperty(propertyName, value string) error {
	// Basic validation for custom properties
	if propertyName == "" {
		return fmt.Errorf("custom property name cannot be empty")
	}

	if value == "" {
		return fmt.Errorf("custom property value cannot be empty")
	}

	// CSS custom properties should start with --
	if strings.HasPrefix(propertyName, "--") {
		return nil // CSS custom properties are flexible
	}

	// For other custom properties, use pattern validation
	return v.validateWithPatterns(propertyName, value)
}

// validateWithPatterns provides fallback validation using regex patterns
func (v *Validator) validateWithPatterns(propertyName, value string) error {
	switch propertyName {
	case "color", "background-color", "border-color":
		return v.validateColor(value)
	case "width", "height", "max-width", "max-height", "min-width", "min-height":
		return v.validateSize(value)
	case "margin", "padding", "margin-top", "margin-bottom", "margin-left", "margin-right",
		"padding-top", "padding-bottom", "padding-left", "padding-right":
		return v.validateSpacing(value)
	case "font-size", "line-height":
		return v.validateFontSize(value)
	case "font-weight":
		return v.validateFontWeight(value)
	case "display":
		return v.validateDisplay(value)
	case "position":
		return v.validatePosition(value)
	case "flex-direction":
		return v.validateFlexDirection(value)
	case "justify-content":
		return v.validateJustifyContent(value)
	case "align-items", "align-content":
		return v.validateAlignment(value)
	case "border-radius":
		return v.validateBorderRadius(value)
	case "opacity":
		return v.validateOpacity(value)
	case "z-index":
		return v.validateZIndex(value)
	default:
		// For unknown properties, just check if it's not empty
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("property value cannot be empty")
		}
		return nil
	}
}

// Color validation patterns
func (v *Validator) validateColor(value string) error {
	value = strings.TrimSpace(value)

	// Named colors
	namedColors := []string{
		"transparent", "currentcolor", "inherit", "initial", "unset",
		"black", "white", "red", "green", "blue", "yellow", "orange", "purple", "pink", "gray", "grey",
	}

	for _, color := range namedColors {
		if strings.EqualFold(value, color) {
			return nil
		}
	}

	// Hex colors (#rgb, #rrggbb, #rgba, #rrggbbaa)
	hexPattern := regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{4}|[0-9a-fA-F]{8})$`)
	if hexPattern.MatchString(value) {
		return nil
	}

	// RGB/RGBA
	rgbPattern := regexp.MustCompile(`^rgba?\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*(?:,\s*(?:0?\.\d+|1|0))?\s*\)$`)
	if rgbPattern.MatchString(value) {
		return nil
	}

	// HSL/HSLA
	hslPattern := regexp.MustCompile(`^hsla?\(\s*\d+\s*,\s*\d+%\s*,\s*\d+%\s*(?:,\s*(?:0?\.\d+|1|0))?\s*\)$`)
	if hslPattern.MatchString(value) {
		return nil
	}

	return fmt.Errorf("invalid color value: %s", value)
}

// Size validation (width, height, etc.)
func (v *Validator) validateSize(value string) error {
	value = strings.TrimSpace(value)

	// Keywords
	keywords := []string{"auto", "inherit", "initial", "unset", "max-content", "min-content", "fit-content"}
	for _, keyword := range keywords {
		if strings.EqualFold(value, keyword) {
			return nil
		}
	}

	// Percentage
	if strings.HasSuffix(value, "%") {
		numStr := strings.TrimSuffix(value, "%")
		if _, err := strconv.ParseFloat(numStr, 64); err != nil {
			return fmt.Errorf("invalid percentage value: %s", value)
		}
		return nil
	}

	// Length units
	units := []string{"px", "em", "rem", "vh", "vw", "vmin", "vmax", "pt", "pc", "in", "cm", "mm", "ex", "ch"}
	for _, unit := range units {
		if strings.HasSuffix(value, unit) {
			numStr := strings.TrimSuffix(value, unit)
			if _, err := strconv.ParseFloat(numStr, 64); err != nil {
				return fmt.Errorf("invalid size value: %s", value)
			}
			return nil
		}
	}

	// Pure number (should be 0 for most properties)
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		if num == 0 {
			return nil
		}
		return fmt.Errorf("non-zero length requires a unit: %s", value)
	}

	return fmt.Errorf("invalid size value: %s", value)
}

// Spacing validation (margin, padding)
func (v *Validator) validateSpacing(value string) error {
	value = strings.TrimSpace(value)

	// Handle multiple values (e.g., "10px 20px")
	parts := strings.Fields(value)
	if len(parts) > 4 {
		return fmt.Errorf("too many spacing values: %s", value)
	}

	for _, part := range parts {
		if err := v.validateSize(part); err != nil {
			return fmt.Errorf("invalid spacing component '%s' in '%s': %v", part, value, err)
		}
	}

	return nil
}

// Font size validation
func (v *Validator) validateFontSize(value string) error {
	value = strings.TrimSpace(value)

	// Keywords
	keywords := []string{
		"xx-small", "x-small", "small", "medium", "large", "x-large", "xx-large",
		"smaller", "larger", "inherit", "initial", "unset",
	}

	for _, keyword := range keywords {
		if strings.EqualFold(value, keyword) {
			return nil
		}
	}

	// Size values
	return v.validateSize(value)
}

// Font weight validation
func (v *Validator) validateFontWeight(value string) error {
	value = strings.TrimSpace(value)

	// Keywords
	keywords := []string{
		"normal", "bold", "bolder", "lighter", "inherit", "initial", "unset",
	}

	for _, keyword := range keywords {
		if strings.EqualFold(value, keyword) {
			return nil
		}
	}

	// Numeric values (100-900)
	if num, err := strconv.Atoi(value); err == nil {
		if num >= 100 && num <= 900 && num%100 == 0 {
			return nil
		}
		return fmt.Errorf("font weight must be between 100-900 in increments of 100: %s", value)
	}

	return fmt.Errorf("invalid font weight: %s", value)
}

// Display validation
func (v *Validator) validateDisplay(value string) error {
	validValues := []string{
		"none", "block", "inline", "inline-block", "flex", "inline-flex",
		"grid", "inline-grid", "table", "table-cell", "table-row",
		"list-item", "run-in", "contents", "inherit", "initial", "unset",
	}

	for _, valid := range validValues {
		if strings.EqualFold(value, valid) {
			return nil
		}
	}

	return fmt.Errorf("invalid display value: %s", value)
}

// Position validation
func (v *Validator) validatePosition(value string) error {
	validValues := []string{
		"static", "relative", "absolute", "fixed", "sticky",
		"inherit", "initial", "unset",
	}

	for _, valid := range validValues {
		if strings.EqualFold(value, valid) {
			return nil
		}
	}

	return fmt.Errorf("invalid position value: %s", value)
}

// Flex direction validation
func (v *Validator) validateFlexDirection(value string) error {
	validValues := []string{
		"row", "row-reverse", "column", "column-reverse",
		"inherit", "initial", "unset",
	}

	for _, valid := range validValues {
		if strings.EqualFold(value, valid) {
			return nil
		}
	}

	return fmt.Errorf("invalid flex-direction value: %s", value)
}

// Justify content validation
func (v *Validator) validateJustifyContent(value string) error {
	validValues := []string{
		"flex-start", "flex-end", "center", "space-between", "space-around",
		"space-evenly", "start", "end", "left", "right",
		"inherit", "initial", "unset",
	}

	for _, valid := range validValues {
		if strings.EqualFold(value, valid) {
			return nil
		}
	}

	return fmt.Errorf("invalid justify-content value: %s", value)
}

// Alignment validation
func (v *Validator) validateAlignment(value string) error {
	validValues := []string{
		"flex-start", "flex-end", "center", "baseline", "stretch",
		"start", "end", "self-start", "self-end",
		"inherit", "initial", "unset",
	}

	for _, valid := range validValues {
		if strings.EqualFold(value, valid) {
			return nil
		}
	}

	return fmt.Errorf("invalid alignment value: %s", value)
}

// Border radius validation
func (v *Validator) validateBorderRadius(value string) error {
	return v.validateSpacing(value) // Same rules as spacing
}

// Opacity validation
func (v *Validator) validateOpacity(value string) error {
	value = strings.TrimSpace(value)

	// Keywords
	if strings.EqualFold(value, "inherit") || strings.EqualFold(value, "initial") || strings.EqualFold(value, "unset") {
		return nil
	}

	// Numeric value (0-1)
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		if num >= 0 && num <= 1 {
			return nil
		}
		return fmt.Errorf("opacity must be between 0 and 1: %s", value)
	}

	return fmt.Errorf("invalid opacity value: %s", value)
}

// Z-index validation
func (v *Validator) validateZIndex(value string) error {
	value = strings.TrimSpace(value)

	// Keywords
	if strings.EqualFold(value, "auto") || strings.EqualFold(value, "inherit") ||
		strings.EqualFold(value, "initial") || strings.EqualFold(value, "unset") {
		return nil
	}

	// Integer value
	if _, err := strconv.Atoi(value); err == nil {
		return nil
	}

	return fmt.Errorf("invalid z-index value: %s", value)
}

// ValidationError represents a CSS validation error
type ValidationError struct {
	Property string `json:"property"`
	Value    string `json:"value"`
	Message  string `json:"message"`
	IsCustom bool   `json:"is_custom"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	if ve.IsCustom {
		return fmt.Sprintf("custom property '%s' with value '%s': %s", ve.Property, ve.Value, ve.Message)
	}
	return fmt.Sprintf("property '%s' with value '%s': %s", ve.Property, ve.Value, ve.Message)
}

// HasErrors checks if there are any validation errors
func HasErrors(errors []ValidationError) bool {
	return len(errors) > 0
}

// FormatErrors formats validation errors into a readable string
func FormatErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return "No validation errors"
	}

	var lines []string
	for _, err := range errors {
		lines = append(lines, fmt.Sprintf("- %s", err.Error()))
	}

	return fmt.Sprintf("CSS Validation Errors:\n%s", strings.Join(lines, "\n"))
}
