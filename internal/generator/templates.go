package generator

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
	"unicode"
)

//go:embed templates/*
var templateFS embed.FS

// TemplateFunctions returns the function map for template processing
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		// String case conversions
		"camelCase":  ToCamelCase,
		"pascalCase": ToPascalCase,
		"snakeCase":  ToSnakeCase,
		"kebabCase":  ToKebabCase,
		"lowerCase":  strings.ToLower,
		"upperCase":  strings.ToUpper,
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"title":      strings.Title,

		// Pluralization
		"pluralize":   Pluralize,
		"singularize": Singularize,

		// String utilities
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"contains":  strings.Contains,
		"replace":   strings.ReplaceAll,
		"trim":      strings.TrimSpace,
		"split":     strings.Split,
		"join":      strings.Join,

		// ERP-specific helpers
		"rootType":         inferRootType,
		"permissions":      generateDefaultPermissions,
		"accountType":      inferAccountType,
		"statementSection": inferStatementSection,
		"cashFlowCategory": inferCashFlowCategory,

		// Template utilities
		"indent":       indent,
		"comment":      comment,
		"docComment":   docComment,
		"generateTODO": generateTODO,

		// Validation helpers
		"validationTags":         generateValidationTags,
		"generateDBTags":         generateDBTags,
		"generateJSONTags":       generateJSONTags,
		"generateValidationTags": generateValidationTags,
		"dbTags":                 generateDBTags,
		"jsonTags":               generateJSONTags,
	}
}

// String case conversion functions

// ToCamelCase converts string to camelCase
func ToCamelCase(s string) string {
	if s == "" {
		return s
	}

	// Split by underscore, hyphen, or space
	parts := strings.FieldsFunc(s, func(c rune) bool {
		return c == '_' || c == '-' || c == ' '
	})

	if len(parts) == 0 {
		return s
	}

	// First part lowercase, rest title case
	result := strings.ToLower(parts[0])
	for i := 1; i < len(parts); i++ {
		result += strings.Title(strings.ToLower(parts[i]))
	}

	return result
}

// ToPascalCase converts string to PascalCase
func ToPascalCase(s string) string {
	if s == "" {
		return s
	}

	// Split by underscore, hyphen, or space
	parts := strings.FieldsFunc(s, func(c rune) bool {
		return c == '_' || c == '-' || c == ' '
	})

	if len(parts) == 0 {
		return s
	}

	var result strings.Builder
	for _, part := range parts {
		if part != "" {
			result.WriteString(strings.Title(strings.ToLower(part)))
		}
	}

	return result.String()
}

// ToSnakeCase converts string to snake_case
func ToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteByte('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}

	// Handle multiple consecutive underscores
	return strings.ReplaceAll(result.String(), "__", "_")
}

// ToKebabCase converts string to kebab-case
func ToKebabCase(s string) string {
	return strings.ReplaceAll(ToSnakeCase(s), "_", "-")
}

// Pluralization functions

// Pluralize converts singular to plural (basic English rules)
func Pluralize(s string) string {
	if s == "" {
		return s
	}

	lower := strings.ToLower(s)

	// Irregular plurals
	irregulars := map[string]string{
		"person":     "people",
		"child":      "children",
		"tooth":      "teeth",
		"foot":       "feet",
		"mouse":      "mice",
		"goose":      "geese",
		"man":        "men",
		"woman":      "women",
		"leaf":       "leaves",
		"life":       "lives",
		"knife":      "knives",
		"wife":       "wives",
		"half":       "halves",
		"shelf":      "shelves",
		"calf":       "calves",
		"wolf":       "wolves",
		"analysis":   "analyses",
		"basis":      "bases",
		"crisis":     "crises",
		"diagnosis":  "diagnoses",
		"hypothesis": "hypotheses",
		"oasis":      "oases",
		"synopsis":   "synopses",
		"thesis":     "theses",
		"phenomenon": "phenomena",
		"criterion":  "criteria",
		"datum":      "data",
	}

	if plural, ok := irregulars[lower]; ok {
		// Preserve original case pattern
		if isAllUpper(s) {
			return strings.ToUpper(plural)
		}
		if isCapitalized(s) {
			return strings.Title(plural)
		}
		return plural
	}

	// Regular pluralization rules
	switch {
	case strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "sh") ||
		strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z"):
		return s + "es"
	case strings.HasSuffix(lower, "y") && len(s) > 1 && !isVowel(rune(lower[len(lower)-2])):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(lower, "f"):
		return s[:len(s)-1] + "ves"
	case strings.HasSuffix(lower, "fe"):
		return s[:len(s)-2] + "ves"
	case strings.HasSuffix(lower, "o") && len(s) > 1 && !isVowel(rune(lower[len(lower)-2])):
		return s + "es"
	default:
		return s + "s"
	}
}

// Singularize converts plural to singular (basic English rules)
func Singularize(s string) string {
	if s == "" {
		return s
	}

	lower := strings.ToLower(s)

	// Irregular singulars (reverse of plurals)
	irregulars := map[string]string{
		"people":     "person",
		"children":   "child",
		"teeth":      "tooth",
		"feet":       "foot",
		"mice":       "mouse",
		"geese":      "goose",
		"men":        "man",
		"women":      "woman",
		"leaves":     "leaf",
		"lives":      "life",
		"knives":     "knife",
		"wives":      "wife",
		"halves":     "half",
		"shelves":    "shelf",
		"calves":     "calf",
		"wolves":     "wolf",
		"analyses":   "analysis",
		"bases":      "basis",
		"crises":     "crisis",
		"diagnoses":  "diagnosis",
		"hypotheses": "hypothesis",
		"oases":      "oasis",
		"synopses":   "synopsis",
		"theses":     "thesis",
		"phenomena":  "phenomenon",
		"criteria":   "criterion",
		"data":       "datum",
	}

	if singular, ok := irregulars[lower]; ok {
		// Preserve original case pattern
		if isAllUpper(s) {
			return strings.ToUpper(singular)
		}
		if isCapitalized(s) {
			return strings.Title(singular)
		}
		return singular
	}

	// Regular singularization rules
	switch {
	case strings.HasSuffix(lower, "ies") && len(s) > 3:
		return s[:len(s)-3] + "y"
	case strings.HasSuffix(lower, "ves") && len(s) > 3:
		return s[:len(s)-3] + "f"
	case strings.HasSuffix(lower, "ses") && len(s) > 3:
		return s[:len(s)-2]
	case strings.HasSuffix(lower, "es") && len(s) > 2:
		// Check if it's a word that actually ends in 'es'
		without_es := s[:len(s)-2]
		lower_without_es := strings.ToLower(without_es)
		if strings.HasSuffix(lower_without_es, "s") || strings.HasSuffix(lower_without_es, "sh") ||
			strings.HasSuffix(lower_without_es, "ch") || strings.HasSuffix(lower_without_es, "x") ||
			strings.HasSuffix(lower_without_es, "z") {
			return without_es
		}
		return s[:len(s)-1] // Just remove 's'
	case strings.HasSuffix(lower, "s") && len(s) > 1:
		return s[:len(s)-1]
	default:
		return s
	}
}

// ERP-specific helper functions

// inferAccountType infers account type from module name
func inferAccountType(moduleName string) string {
	lowerName := strings.ToLower(moduleName)

	switch {
	case strings.Contains(lowerName, "cash") || strings.Contains(lowerName, "bank"):
		return "AccountTypeCash"
	case strings.Contains(lowerName, "receivable") || strings.Contains(lowerName, "customer"):
		return "AccountTypeAccountsReceivable"
	case strings.Contains(lowerName, "inventory") || strings.Contains(lowerName, "stock"):
		return "AccountTypeInventory"
	case strings.Contains(lowerName, "payable") || strings.Contains(lowerName, "vendor"):
		return "AccountTypeAccountsPayable"
	case strings.Contains(lowerName, "revenue") || strings.Contains(lowerName, "sales"):
		return "AccountTypeRevenue"
	case strings.Contains(lowerName, "expense") || strings.Contains(lowerName, "cost"):
		return "AccountTypeExpense"
	default:
		return "AccountTypeCurrentAsset"
	}
}

// inferStatementSection infers financial statement section
func inferStatementSection(moduleName string) string {
	lowerName := strings.ToLower(moduleName)

	switch {
	case strings.Contains(lowerName, "asset") || strings.Contains(lowerName, "inventory") ||
		strings.Contains(lowerName, "receivable") || strings.Contains(lowerName, "cash"):
		return "StatementSectionAssets"
	case strings.Contains(lowerName, "liability") || strings.Contains(lowerName, "payable"):
		return "StatementSectionLiabilities"
	case strings.Contains(lowerName, "equity") || strings.Contains(lowerName, "capital"):
		return "StatementSectionEquity"
	case strings.Contains(lowerName, "revenue") || strings.Contains(lowerName, "income"):
		return "StatementSectionRevenue"
	case strings.Contains(lowerName, "expense") || strings.Contains(lowerName, "cost"):
		return "StatementSectionExpenses"
	default:
		return "StatementSectionAssets"
	}
}

// inferCashFlowCategory infers cash flow category
func inferCashFlowCategory(moduleName string) string {
	lowerName := strings.ToLower(moduleName)

	switch {
	case strings.Contains(lowerName, "sales") || strings.Contains(lowerName, "receivable") ||
		strings.Contains(lowerName, "payable") || strings.Contains(lowerName, "inventory"):
		return "CashFlowCategoryOperating"
	case strings.Contains(lowerName, "asset") || strings.Contains(lowerName, "equipment") ||
		strings.Contains(lowerName, "property"):
		return "CashFlowCategoryInvesting"
	case strings.Contains(lowerName, "loan") || strings.Contains(lowerName, "equity") ||
		strings.Contains(lowerName, "capital"):
		return "CashFlowCategoryFinancing"
	default:
		return "CashFlowCategoryOperating"
	}
}

// Template utility functions

// indent adds indentation to each line
func indent(spaces int, text string) string {
	indentation := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = indentation + line
		}
	}
	return strings.Join(lines, "\n")
}

// comment formats a single-line comment
func comment(text string) string {
	return "// " + text
}

// docComment formats a documentation comment
func docComment(entityName, description string) string {
	return fmt.Sprintf("// %s %s", entityName, description)
}

// generateTODO generates a TODO comment
func generateTODO(description string) string {
	return fmt.Sprintf("// TODO: %s", description)
}

// generateValidationTags generates struct validation tags
func generateValidationTags(fieldName, fieldType string) string {
	lowerField := strings.ToLower(fieldName)

	switch {
	case strings.Contains(lowerField, "email"):
		return `validate:"required,email"`
	case strings.Contains(lowerField, "url"):
		return `validate:"required,url"`
	case strings.Contains(lowerField, "phone"):
		return `validate:"required,e164"`
	case strings.Contains(lowerField, "id") && fieldType == "uuid.UUID":
		return `validate:"required,uuid"`
	case strings.Contains(lowerField, "code"):
		return `validate:"required,min=1,max=50"`
	case strings.Contains(lowerField, "name"):
		return `validate:"required,min=1,max=200"`
	case strings.Contains(lowerField, "description"):
		return `validate:"omitempty,max=1000"`
	case fieldType == "string":
		return `validate:"required,min=1"`
	case fieldType == "decimal.Decimal":
		return `validate:"required,numeric"`
	case fieldType == "bool":
		return "" // No validation needed for bool
	default:
		return `validate:"required"`
	}
}

// generateDBTags generates database struct tags
func generateDBTags(fieldName string) string {
	dbColumn := ToSnakeCase(fieldName)

	switch {
	case strings.HasSuffix(strings.ToLower(fieldName), "id"):
		return fmt.Sprintf(`db:"%s"`, dbColumn)
	case strings.Contains(strings.ToLower(fieldName), "created") ||
		strings.Contains(strings.ToLower(fieldName), "updated") ||
		strings.Contains(strings.ToLower(fieldName), "deleted"):
		return fmt.Sprintf(`db:"%s"`, dbColumn)
	default:
		return fmt.Sprintf(`db:"%s"`, dbColumn)
	}
}

// generateJSONTags generates JSON struct tags
func generateJSONTags(fieldName string) string {
	jsonField := ToCamelCase(fieldName)

	if strings.HasSuffix(strings.ToLower(fieldName), "id") ||
		strings.Contains(strings.ToLower(fieldName), "optional") {
		return fmt.Sprintf(`json:"%s,omitempty"`, jsonField)
	}

	return fmt.Sprintf(`json:"%s"`, jsonField)
}

// Helper functions for string case checking

func isVowel(r rune) bool {
	vowels := "aeiouAEIOU"
	return strings.ContainsRune(vowels, r)
}

func isAllUpper(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func isCapitalized(s string) bool {
	if len(s) == 0 {
		return false
	}
	return unicode.IsUpper(rune(s[0]))
}
