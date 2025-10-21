package atoms

import (
	"fmt"
	"strings"
)

// ============================================================================
// CLASS MANIPULATION UTILITIES
// ============================================================================

// JoinClasses joins multiple class strings, removing duplicates and empty strings.
func JoinClasses(classes ...string) string {
	seen := make(map[string]bool)
	var result []string

	for _, classList := range classes {
		// Split by whitespace
		parts := strings.Fields(classList)
		for _, class := range parts {
			trimmed := strings.TrimSpace(class)
			if trimmed != "" && !seen[trimmed] {
				seen[trimmed] = true
				result = append(result, trimmed)
			}
		}
	}

	if len(result) == 0 {
		return ""
	}

	return strings.Join(result, " ")
}

// CombineClasses merges multiple class strings, removing duplicates.
// Alias for JoinClasses for backwards compatibility.
func CombineClasses(classLists ...string) string {
	return JoinClasses(classLists...)
}

// ConditionalClass returns class if condition is true, empty string otherwise.
func ConditionalClass(condition bool, class string) string {
	if condition {
		return class
	}
	return ""
}

// ConditionalClasses returns trueClass if condition is true, falseClass otherwise.
func ConditionalClasses(condition bool, trueClass, falseClass string) string {
	if condition {
		return trueClass
	}
	return falseClass
}

// AppendClasses appends new classes to existing classes.
func AppendClasses(baseClasses string, additionalClasses ...string) string {
	all := append([]string{baseClasses}, additionalClasses...)
	return JoinClasses(all...)
}

// RemoveClasses removes specific classes from a class string.
func RemoveClasses(classList string, classesToRemove ...string) string {
	removeMap := make(map[string]bool)
	for _, class := range classesToRemove {
		removeMap[strings.TrimSpace(class)] = true
	}

	parts := strings.Fields(classList)
	var result []string

	for _, class := range parts {
		if !removeMap[class] {
			result = append(result, class)
		}
	}

	return strings.Join(result, " ")
}

// HasClass checks if a classList contains a specific class.
func HasClass(classList, targetClass string) bool {
	parts := strings.Fields(classList)
	target := strings.TrimSpace(targetClass)

	for _, class := range parts {
		if class == target {
			return true
		}
	}
	return false
}

// ToggleClass adds or removes a class based on condition.
func ToggleClass(classList, class string, condition bool) string {
	if condition {
		return AppendClasses(classList, class)
	}
	return RemoveClasses(classList, class)
}

// ============================================================================
// ACCESSIBILITY UTILITIES
// ============================================================================

// GenerateAriaDescribedBy creates aria-describedby from multiple IDs.
// Filters out empty strings automatically.
func GenerateAriaDescribedBy(ids ...string) string {
	var validIDs []string
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			validIDs = append(validIDs, trimmed)
		}
	}
	return strings.Join(validIDs, " ")
}

// GenerateAriaLabelledBy creates aria-labelledby from multiple IDs.
// Filters out empty strings automatically.
func GenerateAriaLabelledBy(ids ...string) string {
	return GenerateAriaDescribedBy(ids...) // Same implementation
}

// SanitizeForID converts a string to a valid HTML ID.
// Removes special characters and spaces.
func SanitizeForID(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and special chars with hyphens
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		} else if r == ' ' {
			result.WriteRune('-')
		}
	}

	return result.String()
}

// ============================================================================
// ATTRIBUTE MANIPULATION UTILITIES
// ============================================================================

// MergeAttributes merges multiple attribute maps.
// Later maps override earlier ones for duplicate keys.
func MergeAttributes(attrMaps ...map[string]string) map[string]string {
	result := make(map[string]string)

	for _, attrs := range attrMaps {
		for key, value := range attrs {
			result[key] = value
		}
	}

	return result
}

// FilterEmptyAttributes removes empty string values from attributes.
func FilterEmptyAttributes(attrs map[string]string) map[string]string {
	result := make(map[string]string)

	for key, value := range attrs {
		if strings.TrimSpace(value) != "" {
			result[key] = value
		}
	}

	return result
}

// BoolToString converts a boolean to "true" or "false" string.
func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// StringToBool converts a string to boolean.
func StringToBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes" || s == "on"
}

// ============================================================================
// STRING UTILITIES
// ============================================================================

// Truncate truncates a string to maxLength and adds suffix if truncated.
func Truncate(s string, maxLength int, suffix string) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= len(suffix) {
		return suffix
	}

	return s[:maxLength-len(suffix)] + suffix
}

// Capitalize capitalizes the first letter of a string.
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	// Convert first rune to uppercase
	runes := []rune(s)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] = runes[0] - 32 // Convert to uppercase
	}

	return string(runes)
}

// CamelToKebab converts camelCase to kebab-case.
func CamelToKebab(s string) string {
	var result strings.Builder

	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('-')
			}
			result.WriteRune(r + 32) // Convert to lowercase
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// KebabToCamel converts kebab-case to camelCase.
func KebabToCamel(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) == 0 {
		return s
	}

	var result strings.Builder
	result.WriteString(parts[0])

	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result.WriteString(Capitalize(parts[i]))
		}
	}

	return result.String()
}

// IsEmpty checks if a string is empty or whitespace-only.
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// DefaultString returns defaultValue if s is empty.
func DefaultString(s, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}

// ============================================================================
// SLICE UTILITIES
// ============================================================================

// Contains checks if a slice contains a value.
func Contains[T comparable](slice []T, value T) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// Unique removes duplicate values from a slice while preserving order.
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	var result []T

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Filter filters a slice based on a predicate function.
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T

	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}

	return result
}

// Map transforms a slice using a mapping function.
func Map[T, U any](slice []T, mapper func(T) U) []U {
	result := make([]U, len(slice))

	for i, item := range slice {
		result[i] = mapper(item)
	}

	return result
}

// ============================================================================
// VALIDATION UTILITIES
// ============================================================================

// IsValidEmail performs basic email validation.
func IsValidEmail(email string) bool {
	regex, err := getCompiledRegex(EmailRegex)
	if err != nil {
		return false
	}
	return regex.MatchString(email)
}

// IsValidURL performs basic URL validation.
func IsValidURL(url string) bool {
	regex, err := getCompiledRegex(URLRegex)
	if err != nil {
		return false
	}
	return regex.MatchString(url)
}

// IsNumeric checks if a string contains only numeric characters.
func IsNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// IsAlpha checks if a string contains only alphabetic characters.
func IsAlpha(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

// IsAlphanumeric checks if a string contains only alphanumeric characters.
func IsAlphanumeric(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

// ============================================================================
// FORMAT UTILITIES
// ============================================================================

// FormatBytes formats bytes to human-readable string (KB, MB, GB).
func FormatBytes(bytes int64) string {
	const unit = 1024

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// FormatPlural returns singular or plural form based on count.
func FormatPlural(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

// FormatList formats a slice as a comma-separated list with "and" before last item.
func FormatList(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

// ============================================================================
// COLOR UTILITIES
// ============================================================================

// IsValidHexColor checks if a string is a valid hex color.
func IsValidHexColor(color string) bool {
	regex, err := getCompiledRegex(HexColorRegex)
	if err != nil {
		return false
	}
	return regex.MatchString(color)
}

// NormalizeHexColor ensures hex color has # prefix and is uppercase.
func NormalizeHexColor(color string) string {
	color = strings.TrimSpace(color)

	if !strings.HasPrefix(color, "#") {
		color = "#" + color
	}

	return strings.ToUpper(color)
}

// ============================================================================
// DEBUG UTILITIES
// ============================================================================

// DebugInfo provides component debugging information.
type DebugInfo struct {
	ComponentType   ComponentType
	ComponentID     string
	Props           map[string]interface{}
	ComputedClasses string
	ValidationState ValidationState
	Errors          []string
}

// NewDebugInfo creates debug information for a component.
func NewDebugInfo(componentType ComponentType, id string) *DebugInfo {
	return &DebugInfo{
		ComponentType: componentType,
		ComponentID:   id,
		Props:         make(map[string]interface{}),
		Errors:        []string{},
	}
}

// AddProp adds a property to debug info.
func (di *DebugInfo) AddProp(key string, value interface{}) *DebugInfo {
	di.Props[key] = value
	return di
}

// AddError adds an error to debug info.
func (di *DebugInfo) AddError(err string) *DebugInfo {
	di.Errors = append(di.Errors, err)
	return di
}

// ToMap converts debug info to a map for JSON serialization.
func (di *DebugInfo) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"componentType":   di.ComponentType.String(),
		"componentId":     di.ComponentID,
		"props":           di.Props,
		"computedClasses": di.ComputedClasses,
		"validationState": di.ValidationState.String(),
		"errors":          di.Errors,
	}
}

// ToJSON converts debug info to JSON string.
func (di *DebugInfo) ToJSON() string {
	return ToJSONString(di.ToMap())
}

// ============================================================================
// PERFORMANCE UTILITIES
// ============================================================================

// LazyLoadConfig configures lazy loading behavior.
type LazyLoadConfig struct {
	Enabled    bool
	Threshold  string // Intersection observer threshold
	RootMargin string // Intersection observer root margin
}

// DefaultLazyLoad returns sensible lazy load defaults.
func DefaultLazyLoad() LazyLoadConfig {
	return LazyLoadConfig{
		Enabled:    true,
		Threshold:  "0.1",
		RootMargin: "50px",
	}
}

// ToDataAttributes converts lazy load config to data attributes.
func (l LazyLoadConfig) ToDataAttributes() DataAttributes {
	attrs := NewDataAttributes()
	attrs.Set("lazy", BoolToString(l.Enabled))
	attrs.Set("threshold", l.Threshold)
	attrs.Set("root-margin", l.RootMargin)
	return attrs
}

// ============================================================================
// BUILDER PATTERN HELPERS
// ============================================================================

// ComponentBuilder is a generic interface for component builders.
type ComponentBuilder[T any] interface {
	Build() T
}

// WithOption is a functional option pattern for configuring components.
type WithOption[T any] func(*T)

// ApplyOptions applies functional options to a value.
func ApplyOptions[T any](target *T, options ...WithOption[T]) {
	for _, option := range options {
		if option != nil {
			option(target)
		}
	}
}

// ============================================================================
// CONVERSION UTILITIES
// ============================================================================

// Ptr returns a pointer to a value (useful for optional fields).
func Ptr[T any](v T) *T {
	return &v
}

// DerefOr dereferences a pointer or returns default value if nil.
func DerefOr[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// ============================================================================
// ERROR HANDLING UTILITIES
// ============================================================================

// Must panics if error is not nil (use for initialization only).
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// IgnoreError returns value and ignores error (use sparingly).
func IgnoreError[T any](value T, _ error) T {
	return value
}

// ============================================================================
// TESTING UTILITIES
// ============================================================================

// MockProps creates mock component props for testing.
func MockProps(componentType ComponentType) map[string]interface{} {
	return map[string]interface{}{
		"id":          GenerateID(string(componentType)),
		"name":        fmt.Sprintf("test-%s", componentType),
		"dataTestId":  fmt.Sprintf("%s-test", componentType),
		"disabled":    false,
		"required":    false,
		"size":        SizeMD,
		"variant":     VariantDefault,
		"colorScheme": ColorDefault,
	}
}

// AssertNonEmpty panics if string is empty (for testing).
func AssertNonEmpty(s, fieldName string) {
	if IsEmpty(s) {
		panic(fmt.Sprintf("%s cannot be empty", fieldName))
	}
}

// AssertValid panics if condition is false (for testing).
func AssertValid(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

// GetActionSizeClasses returns CSS classes for action component sizing.
func GetActionSizeClasses(size Size) []string {
	switch size {
	case SizeXS:
		return []string{"px-2", "py-1", "text-xs"}
	case SizeSM:
		return []string{"px-3", "py-1.5", "text-sm"}
	case SizeMD:
		return []string{"px-4", "py-2", "text-sm"}
	case SizeLG:
		return []string{"px-6", "py-2.5", "text-base"}
	case SizeXL:
		return []string{"px-8", "py-3", "text-lg"}
	default:
		return []string{"px-4", "py-2", "text-sm"}
	}
}

// GetStatusSizeClasses returns CSS classes for status component sizing.
func GetStatusSizeClasses(size Size) []string {
	switch size {
	case SizeXS:
		return []string{"px-1.5", "py-0.5", "text-xs"}
	case SizeSM:
		return []string{"px-2", "py-0.5", "text-xs"}
	case SizeMD:
		return []string{"px-2.5", "py-0.5", "text-xs"}
	case SizeLG:
		return []string{"px-3", "py-1", "text-sm"}
	case SizeXL:
		return []string{"px-4", "py-1.5", "text-sm"}
	default:
		return []string{"px-2.5", "py-0.5", "text-xs"}
	}
}
