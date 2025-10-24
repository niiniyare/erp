package atoms

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

// UtilTestSuite tests the utility functions in util.go
type UtilTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *UtilTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestUtilTestSuite runs the util test suite
func TestUtilTestSuite(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}

// ============================================================================
// CLASS MANIPULATION TESTS
// ============================================================================

func (suite *UtilTestSuite) TestJoinClasses() {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"empty input", []string{}, ""},
		{"single class", []string{"btn"}, "btn"},
		{"multiple classes", []string{"btn", "btn-primary"}, "btn btn-primary"},
		{"with duplicates", []string{"btn", "btn-primary", "btn"}, "btn btn-primary"},
		{"with whitespace", []string{"btn ", " btn-primary ", "  btn-lg  "}, "btn btn-primary btn-lg"},
		{"with empty strings", []string{"btn", "", "btn-primary", ""}, "btn btn-primary"},
		{"complex string", []string{"btn btn-sm", "btn-primary", "hover:btn-secondary"}, "btn btn-sm btn-primary hover:btn-secondary"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := JoinClasses(tt.input...)
			suite.Equal(tt.expected, result)
		})
	}
}

func (suite *UtilTestSuite) TestCombineClasses() {
	// Test that CombineClasses is an alias for JoinClasses
	result1 := JoinClasses("btn", "btn-primary")
	result2 := CombineClasses("btn", "btn-primary")
	suite.Equal(result1, result2)
}

func (suite *UtilTestSuite) TestConditionalClass() {
	suite.Equal("active", ConditionalClass(true, "active"))
	suite.Equal("", ConditionalClass(false, "active"))
}

func (suite *UtilTestSuite) TestConditionalClasses() {
	suite.Equal("btn-primary", ConditionalClasses(true, "btn-primary", "btn-secondary"))
	suite.Equal("btn-secondary", ConditionalClasses(false, "btn-primary", "btn-secondary"))
}

func (suite *UtilTestSuite) TestAppendClasses() {
	result := AppendClasses("btn", "btn-primary", "btn-lg")
	suite.Equal("btn btn-primary btn-lg", result)
}

func (suite *UtilTestSuite) TestRemoveClasses() {
	result := RemoveClasses("btn btn-primary btn-lg active", "btn-primary", "active")
	suite.Equal("btn btn-lg", result)
}

func (suite *UtilTestSuite) TestHasClass() {
	classList := "btn btn-primary btn-lg"
	suite.True(HasClass(classList, "btn-primary"))
	suite.False(HasClass(classList, "btn-secondary"))
	suite.True(HasClass(classList, "btn"))
}

func (suite *UtilTestSuite) TestToggleClass() {
	// Add class when condition is true
	result1 := ToggleClass("btn", "active", true)
	suite.Contains(result1, "active")

	// Remove class when condition is false
	result2 := ToggleClass("btn active", "active", false)
	suite.NotContains(result2, "active")
}

// ============================================================================
// ACCESSIBILITY TESTS
// ============================================================================

func (suite *UtilTestSuite) TestGenerateAriaDescribedBy() {
	tests := []struct {
		name     string
		ids      []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"single id", []string{"help-1"}, "help-1"},
		{"multiple ids", []string{"help-1", "error-1"}, "help-1 error-1"},
		{"with empty strings", []string{"help-1", "", "error-1", ""}, "help-1 error-1"},
		{"with whitespace", []string{" help-1 ", " error-1 "}, "help-1 error-1"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := GenerateAriaDescribedBy(tt.ids...)
			suite.Equal(tt.expected, result)
		})
	}
}

func (suite *UtilTestSuite) TestGenerateAriaLabelledBy() {
	// Should behave same as GenerateAriaDescribedBy
	result := GenerateAriaLabelledBy("label-1", "label-2")
	suite.Equal("label-1 label-2", result)
}

func (suite *UtilTestSuite) TestSanitizeForID() {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Text", "simple-text"},
		{"With Special!@#$%", "with-special"},
		{"Multi   Spaces", "multi---spaces"},
		{"123-test_ID", "123-test_id"},
		{"", ""},
		{"UPPERCASE", "uppercase"},
	}

	for _, tt := range tests {
		result := SanitizeForID(tt.input)
		suite.Equal(tt.expected, result)
	}
}

// ============================================================================
// ATTRIBUTE MANIPULATION TESTS
// ============================================================================

func (suite *UtilTestSuite) TestMergeAttributes() {
	attr1 := map[string]string{"class": "btn", "id": "button1"}
	attr2 := map[string]string{"class": "btn-primary", "data-test": "true"}
	attr3 := map[string]string{"id": "button2"}

	result := MergeAttributes(attr1, attr2, attr3)

	suite.Equal("btn-primary", result["class"]) // attr2 overrides attr1
	suite.Equal("button2", result["id"])        // attr3 overrides attr1
	suite.Equal("true", result["data-test"])
}

func (suite *UtilTestSuite) TestFilterEmptyAttributes() {
	attrs := map[string]string{
		"valid":     "value",
		"empty":     "",
		"whitespace": "   ",
		"another":   "valid-value",
	}

	result := FilterEmptyAttributes(attrs)

	suite.Contains(result, "valid")
	suite.Contains(result, "another")
	suite.NotContains(result, "empty")
	suite.NotContains(result, "whitespace")
}

func (suite *UtilTestSuite) TestBoolToString() {
	suite.Equal("true", BoolToString(true))
	suite.Equal("false", BoolToString(false))
}

func (suite *UtilTestSuite) TestStringToBool() {
	trueValues := []string{"true", "TRUE", "True", "1", "yes", "YES", "on", "ON"}
	falseValues := []string{"false", "FALSE", "0", "no", "off", "", "invalid"}

	for _, val := range trueValues {
		suite.True(StringToBool(val), "Expected %s to be true", val)
	}

	for _, val := range falseValues {
		suite.False(StringToBool(val), "Expected %s to be false", val)
	}
}

// ============================================================================
// STRING UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestTruncate() {
	tests := []struct {
		input     string
		maxLength int
		suffix    string
		expected  string
	}{
		{"short", 10, "...", "short"},
		{"this is a long string", 10, "...", "this is..."},
		{"exact", 5, "...", "exact"},
		{"", 5, "...", ""},
		{"test", 2, "...", "..."},
	}

	for _, tt := range tests {
		result := Truncate(tt.input, tt.maxLength, tt.suffix)
		suite.Equal(tt.expected, result)
	}
}

func (suite *UtilTestSuite) TestCapitalize() {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"HELLO", "HELLO"},
		{"Hello", "Hello"},
		{"", ""},
		{"a", "A"},
		{"123", "123"},
	}

	for _, tt := range tests {
		result := Capitalize(tt.input)
		suite.Equal(tt.expected, result)
	}
}

func (suite *UtilTestSuite) TestCamelToKebab() {
	tests := []struct {
		input    string
		expected string
	}{
		{"camelCase", "camel-case"},
		{"PascalCase", "pascal-case"},
		{"simpleword", "simpleword"},
		{"XMLHttpRequest", "x-m-l-http-request"},
		{"", ""},
		{"A", "a"},
	}

	for _, tt := range tests {
		result := CamelToKebab(tt.input)
		suite.Equal(tt.expected, result)
	}
}

func (suite *UtilTestSuite) TestKebabToCamel() {
	tests := []struct {
		input    string
		expected string
	}{
		{"kebab-case", "kebabCase"},
		{"simple", "simple"},
		{"multi-word-string", "multiWordString"},
		{"", ""},
		{"single-", "single"},
		{"-leading", "Leading"},
	}

	for _, tt := range tests {
		result := KebabToCamel(tt.input)
		suite.Equal(tt.expected, result)
	}
}

func (suite *UtilTestSuite) TestIsEmpty() {
	suite.True(IsEmpty(""))
	suite.True(IsEmpty("   "))
	suite.True(IsEmpty("\t\n"))
	suite.False(IsEmpty("text"))
	suite.False(IsEmpty(" text "))
}

func (suite *UtilTestSuite) TestDefaultString() {
	suite.Equal("default", DefaultString("", "default"))
	suite.Equal("default", DefaultString("   ", "default"))
	suite.Equal("value", DefaultString("value", "default"))
}

// ============================================================================
// SLICE UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestContains() {
	stringSlice := []string{"a", "b", "c"}
	suite.True(Contains(stringSlice, "b"))
	suite.False(Contains(stringSlice, "d"))

	intSlice := []int{1, 2, 3}
	suite.True(Contains(intSlice, 2))
	suite.False(Contains(intSlice, 4))
}

func (suite *UtilTestSuite) TestUnique() {
	input := []string{"a", "b", "a", "c", "b"}
	expected := []string{"a", "b", "c"}
	result := Unique(input)
	suite.Equal(expected, result)

	intInput := []int{1, 2, 1, 3, 2}
	intExpected := []int{1, 2, 3}
	intResult := Unique(intInput)
	suite.Equal(intExpected, intResult)
}

func (suite *UtilTestSuite) TestFilter() {
	input := []int{1, 2, 3, 4, 5}
	result := Filter(input, func(x int) bool { return x%2 == 0 })
	expected := []int{2, 4}
	suite.Equal(expected, result)
}

func (suite *UtilTestSuite) TestMap() {
	input := []int{1, 2, 3}
	result := Map(input, func(x int) string { return string(rune('0' + x)) })
	expected := []string{"1", "2", "3"}
	suite.Equal(expected, result)
}

// ============================================================================
// VALIDATION UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestIsValidEmail() {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@example.org",
	}

	invalidEmails := []string{
		"invalid-email",
		"@example.com",
		"test@",
		"",
	}

	for _, email := range validEmails {
		suite.True(IsValidEmail(email), "Expected %s to be valid", email)
	}

	for _, email := range invalidEmails {
		suite.False(IsValidEmail(email), "Expected %s to be invalid", email)
	}
}

func (suite *UtilTestSuite) TestIsValidURL() {
	validURLs := []string{
		"https://example.com",
		"http://www.example.org",
		"https://sub.domain.com/path",
	}

	invalidURLs := []string{
		"not-a-url",
		"ftp://example.com", // Only http/https supported
		"",
		"example.com", // Missing protocol
	}

	for _, url := range validURLs {
		suite.True(IsValidURL(url), "Expected %s to be valid", url)
	}

	for _, url := range invalidURLs {
		suite.False(IsValidURL(url), "Expected %s to be invalid", url)
	}
}

func (suite *UtilTestSuite) TestIsNumeric() {
	suite.True(IsNumeric("123"))
	suite.True(IsNumeric("0"))
	suite.False(IsNumeric("12.3"))
	suite.False(IsNumeric("abc"))
	suite.False(IsNumeric(""))
	suite.False(IsNumeric("12a"))
}

func (suite *UtilTestSuite) TestIsAlpha() {
	suite.True(IsAlpha("abc"))
	suite.True(IsAlpha("ABC"))
	suite.True(IsAlpha("AbC"))
	suite.False(IsAlpha("abc123"))
	suite.False(IsAlpha(""))
	suite.False(IsAlpha("abc "))
}

func (suite *UtilTestSuite) TestIsAlphanumeric() {
	suite.True(IsAlphanumeric("abc123"))
	suite.True(IsAlphanumeric("ABC"))
	suite.True(IsAlphanumeric("123"))
	suite.False(IsAlphanumeric("abc 123"))
	suite.False(IsAlphanumeric(""))
	suite.False(IsAlphanumeric("abc!"))
}

// ============================================================================
// FORMAT UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestFormatBytes() {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		suite.Equal(tt.expected, result)
	}
}

func (suite *UtilTestSuite) TestFormatPlural() {
	suite.Equal("item", FormatPlural(1, "item", "items"))
	suite.Equal("items", FormatPlural(0, "item", "items"))
	suite.Equal("items", FormatPlural(2, "item", "items"))
	suite.Equal("items", FormatPlural(10, "item", "items"))
}

func (suite *UtilTestSuite) TestFormatList() {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"apple"}, "apple"},
		{[]string{"apple", "banana"}, "apple and banana"},
		{[]string{"apple", "banana", "cherry"}, "apple, banana, and cherry"},
		{[]string{"a", "b", "c", "d"}, "a, b, c, and d"},
	}

	for _, tt := range tests {
		result := FormatList(tt.input)
		suite.Equal(tt.expected, result)
	}
}

// ============================================================================
// COLOR UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestIsValidHexColor() {
	validColors := []string{"#FF0000", "#ff0000", "#F00", "#f00"}
	invalidColors := []string{"FF0000", "#GG0000", "#FF00", "red", ""}

	for _, color := range validColors {
		suite.True(IsValidHexColor(color), "Expected %s to be valid", color)
	}

	for _, color := range invalidColors {
		suite.False(IsValidHexColor(color), "Expected %s to be invalid", color)
	}
}

func (suite *UtilTestSuite) TestNormalizeHexColor() {
	tests := []struct {
		input    string
		expected string
	}{
		{"ff0000", "#FF0000"},
		{"#ff0000", "#FF0000"},
		{"  #f00  ", "#F00"},
		{"F00", "#F00"},
	}

	for _, tt := range tests {
		result := NormalizeHexColor(tt.input)
		suite.Equal(tt.expected, result)
	}
}

// ============================================================================
// DEBUG UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestDebugInfo() {
	debugInfo := NewDebugInfo(ComponentButton, "test-button")
	debugInfo.AddProp("variant", "primary").AddError("test error")

	suite.Equal(ComponentButton, debugInfo.ComponentType)
	suite.Equal("test-button", debugInfo.ComponentID)
	suite.Equal("primary", debugInfo.Props["variant"])
	suite.Contains(debugInfo.Errors, "test error")

	// Test ToMap
	infoMap := debugInfo.ToMap()
	suite.Equal("button", infoMap["componentType"])
	suite.Equal("test-button", infoMap["componentId"])

	// Test ToJSON
	jsonStr := debugInfo.ToJSON()
	suite.Contains(jsonStr, "button")
	suite.Contains(jsonStr, "test-button")
}

// ============================================================================
// PERFORMANCE UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestLazyLoadConfig() {
	defaultConfig := DefaultLazyLoad()
	suite.True(defaultConfig.Enabled)
	suite.Equal("0.1", defaultConfig.Threshold)
	suite.Equal("50px", defaultConfig.RootMargin)

	// Test ToDataAttributes
	attrs := defaultConfig.ToDataAttributes()
	suite.Equal("true", attrs.Get("lazy"))
	suite.Equal("0.1", attrs.Get("threshold"))
	suite.Equal("50px", attrs.Get("root-margin"))
}

// ============================================================================
// CONVERSION UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestPtr() {
	value := "test"
	ptr := Ptr(value)
	suite.Equal(&value, ptr)
	suite.Equal(value, *ptr)
}

func (suite *UtilTestSuite) TestDerefOr() {
	value := "test"
	ptr := &value
	
	suite.Equal("test", DerefOr(ptr, "default"))
	suite.Equal("default", DerefOr((*string)(nil), "default"))
}

// ============================================================================
// ERROR HANDLING UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestMust() {
	// Should return value when no error
	result := Must("success", nil)
	suite.Equal("success", result)

	// Should panic when error exists
	suite.Panics(func() {
		Must("fail", errors.New("test error"))
	})
}

func (suite *UtilTestSuite) TestIgnoreError() {
	result := IgnoreError("value", errors.New("ignored"))
	suite.Equal("value", result)
}

// ============================================================================
// TESTING UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestMockProps() {
	props := MockProps(ComponentButton)
	
	suite.Contains(props, "id")
	suite.Contains(props, "name")
	suite.Contains(props, "dataTestId")
	suite.Equal(SizeMD, props["size"])
	suite.Equal(VariantDefault, props["variant"])
}

func (suite *UtilTestSuite) TestAssertNonEmpty() {
	// Should not panic for non-empty string
	suite.NotPanics(func() {
		AssertNonEmpty("value", "field")
	})

	// Should panic for empty string
	suite.Panics(func() {
		AssertNonEmpty("", "field")
	})
}

func (suite *UtilTestSuite) TestAssertValid() {
	// Should not panic for true condition
	suite.NotPanics(func() {
		AssertValid(true, "should not fail")
	})

	// Should panic for false condition
	suite.Panics(func() {
		AssertValid(false, "should fail")
	})
}

// ============================================================================
// SIZE CLASS UTILITIES TESTS
// ============================================================================

func (suite *UtilTestSuite) TestGetActionSizeClasses() {
	tests := []struct {
		size     Size
		expected []string
	}{
		{SizeXS, []string{"px-2", "py-1", "text-xs"}},
		{SizeSM, []string{"px-3", "py-1.5", "text-sm"}},
		{SizeMD, []string{"px-4", "py-2", "text-sm"}},
		{SizeLG, []string{"px-6", "py-2.5", "text-base"}},
		{SizeXL, []string{"px-8", "py-3", "text-lg"}},
	}

	for _, tt := range tests {
		result := GetActionSizeClasses(tt.size)
		suite.Equal(tt.expected, result)
	}
}

func (suite *UtilTestSuite) TestGetStatusSizeClasses() {
	tests := []struct {
		size     Size
		expected []string
	}{
		{SizeXS, []string{"px-1.5", "py-0.5", "text-xs"}},
		{SizeSM, []string{"px-2", "py-0.5", "text-xs"}},
		{SizeMD, []string{"px-2.5", "py-0.5", "text-xs"}},
		{SizeLG, []string{"px-3", "py-1", "text-sm"}},
		{SizeXL, []string{"px-4", "py-1.5", "text-sm"}},
	}

	for _, tt := range tests {
		result := GetStatusSizeClasses(tt.size)
		suite.Equal(tt.expected, result)
	}
}

// ============================================================================
// BUILDER PATTERN TESTS
// ============================================================================

func (suite *UtilTestSuite) TestApplyOptions() {
	type TestStruct struct {
		Value string
		Count int
	}

	obj := &TestStruct{}
	
	option1 := func(t *TestStruct) { t.Value = "test" }
	option2 := func(t *TestStruct) { t.Count = 42 }
	
	ApplyOptions(obj, option1, option2)
	
	suite.Equal("test", obj.Value)
	suite.Equal(42, obj.Count)
}