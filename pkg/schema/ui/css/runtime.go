// Package css - CSS-in-Go runtime system for dynamic CSS generation
// This system allows components to generate CSS at runtime with type safety and optimization
package css

import (
	"fmt"
	"sort"
	"strings"
)

// CSSRule represents a single CSS rule with selectors and declarations
type CSSRule struct {
	// CSS selector (e.g., ".button", "#header", "div.card")
	Selector string
	// CSS declarations as key-value pairs
	Declarations map[string]string
	// Media query for responsive design (optional)
	MediaQuery string
	// Pseudo-classes and pseudo-elements (e.g., ":hover", "::before")
	PseudoSelector string
	// Priority for rule ordering (higher numbers come last)
	Priority int
}

// CSSStylesheet represents a collection of CSS rules that can be rendered
type CSSStylesheet struct {
	// Collection of CSS rules
	Rules []CSSRule
	// Global CSS variables (custom properties)
	Variables map[string]string
	// Keyframe animations
	Keyframes map[string][]CSSKeyframe
	// Font face declarations
	FontFaces []CSSFontFace
	// Import statements
	Imports []string
}

// CSSKeyframe represents a keyframe in a CSS animation
type CSSKeyframe struct {
	// Percentage or keyword (0%, 50%, 100%, from, to)
	Position string
	// CSS declarations for this keyframe
	Declarations map[string]string
}

// CSSFontFace represents a @font-face declaration
type CSSFontFace struct {
	FontFamily  string
	Src         []string
	FontWeight  string
	FontStyle   string
	FontDisplay string
}

// CSSGenerator provides the main interface for generating CSS
type CSSGenerator struct {
	// Current stylesheet being built
	stylesheet *CSSStylesheet
	// Configuration options
	config *CSSConfig
}

// CSSConfig contains configuration options for CSS generation
type CSSConfig struct {
	// Whether to minify the output CSS
	Minify bool
	// Whether to include comments in output
	IncludeComments bool
	// Prefix for generated class names
	ClassPrefix string
	// Whether to auto-prefix CSS properties
	AutoPrefix bool
	// Indentation for pretty-printing (ignored if Minify is true)
	Indent string
	// Whether to sort CSS rules by selector
	SortRules bool
	// Whether to remove duplicate rules
	RemoveDuplicates bool
}

// DefaultCSSConfig returns sensible default configuration
func DefaultCSSConfig() *CSSConfig {
	return &CSSConfig{
		Minify:           false,
		IncludeComments:  true,
		ClassPrefix:      "ui-",
		AutoPrefix:       true,
		Indent:           "  ",
		SortRules:        true,
		RemoveDuplicates: true,
	}
}

// NewCSSGenerator creates a new CSS generator with default configuration
func NewCSSGenerator() *CSSGenerator {
	return &CSSGenerator{
		stylesheet: &CSSStylesheet{
			Rules:     make([]CSSRule, 0),
			Variables: make(map[string]string),
			Keyframes: make(map[string][]CSSKeyframe),
			FontFaces: make([]CSSFontFace, 0),
			Imports:   make([]string, 0),
		},
		config: DefaultCSSConfig(),
	}
}

// NewCSSGeneratorWithConfig creates a CSS generator with custom configuration
func NewCSSGeneratorWithConfig(config *CSSConfig) *CSSGenerator {
	gen := NewCSSGenerator()
	gen.config = config
	return gen
}

// AddRule adds a CSS rule to the stylesheet
func (g *CSSGenerator) AddRule(selector string, declarations map[string]string) *CSSGenerator {
	rule := CSSRule{
		Selector:     g.prefixSelector(selector),
		Declarations: declarations,
		Priority:     0,
	}
	g.stylesheet.Rules = append(g.stylesheet.Rules, rule)
	return g
}

// AddRuleWithMedia adds a CSS rule with a media query
func (g *CSSGenerator) AddRuleWithMedia(selector string, declarations map[string]string, mediaQuery string) *CSSGenerator {
	rule := CSSRule{
		Selector:     g.prefixSelector(selector),
		Declarations: declarations,
		MediaQuery:   mediaQuery,
		Priority:     1, // Media queries get higher priority
	}
	g.stylesheet.Rules = append(g.stylesheet.Rules, rule)
	return g
}

// AddRuleWithPseudo adds a CSS rule with pseudo-selector
func (g *CSSGenerator) AddRuleWithPseudo(selector string, pseudoSelector string, declarations map[string]string) *CSSGenerator {
	rule := CSSRule{
		Selector:       g.prefixSelector(selector),
		Declarations:   declarations,
		PseudoSelector: pseudoSelector,
		Priority:       2, // Pseudo-selectors get higher priority
	}
	g.stylesheet.Rules = append(g.stylesheet.Rules, rule)
	return g
}

// AddVariable adds a CSS custom property (variable)
func (g *CSSGenerator) AddVariable(name, value string) *CSSGenerator {
	// Ensure variable name starts with --
	if !strings.HasPrefix(name, "--") {
		name = "--" + name
	}
	g.stylesheet.Variables[name] = value
	return g
}

// AddKeyframes adds CSS keyframe animation
func (g *CSSGenerator) AddKeyframes(name string, keyframes []CSSKeyframe) *CSSGenerator {
	g.stylesheet.Keyframes[name] = keyframes
	return g
}

// AddFontFace adds a font face declaration
func (g *CSSGenerator) AddFontFace(fontFace CSSFontFace) *CSSGenerator {
	g.stylesheet.FontFaces = append(g.stylesheet.FontFaces, fontFace)
	return g
}

// AddImport adds an import statement
func (g *CSSGenerator) AddImport(url string) *CSSGenerator {
	g.stylesheet.Imports = append(g.stylesheet.Imports, url)
	return g
}

// Helper method to prefix selectors
func (g *CSSGenerator) prefixSelector(selector string) string {
	if g.config.ClassPrefix == "" {
		return selector
	}

	// Don't prefix if selector already has a prefix or is a tag/ID/pseudo selector
	if strings.HasPrefix(selector, "."+g.config.ClassPrefix) ||
		strings.HasPrefix(selector, "#") ||
		strings.HasPrefix(selector, ":") ||
		!strings.HasPrefix(selector, ".") {
		return selector
	}

	// Add prefix to class selectors
	if strings.HasPrefix(selector, ".") {
		return "." + g.config.ClassPrefix + selector[1:]
	}

	return selector
}

// Generate produces the final CSS string
func (g *CSSGenerator) Generate() string {
	var css strings.Builder

	// Add imports first
	for _, imp := range g.stylesheet.Imports {
		css.WriteString(fmt.Sprintf("@import url('%s');\n", imp))
	}

	if len(g.stylesheet.Imports) > 0 {
		css.WriteString("\n")
	}

	// Add CSS variables in :root
	if len(g.stylesheet.Variables) > 0 {
		css.WriteString(":root {\n")

		// Sort variables for consistent output
		var varNames []string
		for name := range g.stylesheet.Variables {
			varNames = append(varNames, name)
		}
		sort.Strings(varNames)

		for _, name := range varNames {
			value := g.stylesheet.Variables[name]
			if g.config.Minify {
				css.WriteString(fmt.Sprintf("%s:%s;", name, value))
			} else {
				css.WriteString(fmt.Sprintf("%s%s: %s;\n", g.config.Indent, name, value))
			}
		}

		css.WriteString("}\n\n")
	}

	// Add font faces
	for _, fontFace := range g.stylesheet.FontFaces {
		css.WriteString(g.generateFontFace(fontFace))
		css.WriteString("\n")
	}

	// Add keyframes
	for name, keyframes := range g.stylesheet.Keyframes {
		css.WriteString(g.generateKeyframes(name, keyframes))
		css.WriteString("\n")
	}

	// Process and sort rules
	rules := g.processRules()

	// Group rules by media query
	rulesByMedia := make(map[string][]CSSRule)
	var normalRules []CSSRule

	for _, rule := range rules {
		if rule.MediaQuery != "" {
			rulesByMedia[rule.MediaQuery] = append(rulesByMedia[rule.MediaQuery], rule)
		} else {
			normalRules = append(normalRules, rule)
		}
	}

	// Add normal rules first
	for _, rule := range normalRules {
		css.WriteString(g.generateRule(rule))
	}

	// Add media query rules
	for mediaQuery, mediaRules := range rulesByMedia {
		css.WriteString(fmt.Sprintf("@media %s {\n", mediaQuery))
		for _, rule := range mediaRules {
			// Don't include media query in the rule itself when inside @media block
			rule.MediaQuery = ""
			css.WriteString(g.addIndent(g.generateRule(rule)))
		}
		css.WriteString("}\n\n")
	}

	return strings.TrimSpace(css.String())
}

// processRules handles deduplication and sorting
func (g *CSSGenerator) processRules() []CSSRule {
	rules := g.stylesheet.Rules

	// Remove duplicates if configured
	if g.config.RemoveDuplicates {
		rules = g.removeDuplicateRules(rules)
	}

	// Sort rules if configured
	if g.config.SortRules {
		sort.Slice(rules, func(i, j int) bool {
			// First sort by priority
			if rules[i].Priority != rules[j].Priority {
				return rules[i].Priority < rules[j].Priority
			}
			// Then by selector
			return rules[i].Selector < rules[j].Selector
		})
	}

	return rules
}

// removeDuplicateRules removes duplicate CSS rules
func (g *CSSGenerator) removeDuplicateRules(rules []CSSRule) []CSSRule {
	seen := make(map[string]bool)
	var unique []CSSRule

	for _, rule := range rules {
		key := g.getRuleKey(rule)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, rule)
		}
	}

	return unique
}

// getRuleKey creates a unique key for a CSS rule
func (g *CSSGenerator) getRuleKey(rule CSSRule) string {
	return fmt.Sprintf("%s|%s|%s", rule.Selector, rule.PseudoSelector, rule.MediaQuery)
}

// generateRule generates CSS for a single rule
func (g *CSSGenerator) generateRule(rule CSSRule) string {
	var css strings.Builder

	// Build selector
	selector := rule.Selector
	if rule.PseudoSelector != "" {
		selector += rule.PseudoSelector
	}

	if g.config.Minify {
		css.WriteString(fmt.Sprintf("%s{", selector))
	} else {
		css.WriteString(fmt.Sprintf("%s {\n", selector))
	}

	// Sort declarations for consistent output
	var properties []string
	for prop := range rule.Declarations {
		properties = append(properties, prop)
	}
	sort.Strings(properties)

	// Add declarations
	for _, prop := range properties {
		value := rule.Declarations[prop]
		// Auto-prefix if configured
		if g.config.AutoPrefix {
			value = g.autoPrefixValue(prop, value)
		}

		if g.config.Minify {
			css.WriteString(fmt.Sprintf("%s:%s;", prop, value))
		} else {
			css.WriteString(fmt.Sprintf("%s%s: %s;\n", g.config.Indent, prop, value))
		}
	}

	if g.config.Minify {
		css.WriteString("}")
	} else {
		css.WriteString("}\n\n")
	}

	return css.String()
}

// generateFontFace generates CSS for @font-face
func (g *CSSGenerator) generateFontFace(fontFace CSSFontFace) string {
	var css strings.Builder

	if g.config.Minify {
		css.WriteString("@font-face{")
	} else {
		css.WriteString("@font-face {\n")
	}

	// Font family (required)
	if g.config.Minify {
		css.WriteString(fmt.Sprintf("font-family:'%s';", fontFace.FontFamily))
	} else {
		css.WriteString(fmt.Sprintf("%sfont-family: '%s';\n", g.config.Indent, fontFace.FontFamily))
	}

	// Src (required)
	if len(fontFace.Src) > 0 {
		src := strings.Join(fontFace.Src, ", ")
		if g.config.Minify {
			css.WriteString(fmt.Sprintf("src:%s;", src))
		} else {
			css.WriteString(fmt.Sprintf("%ssrc: %s;\n", g.config.Indent, src))
		}
	}

	// Optional properties
	if fontFace.FontWeight != "" {
		if g.config.Minify {
			css.WriteString(fmt.Sprintf("font-weight:%s;", fontFace.FontWeight))
		} else {
			css.WriteString(fmt.Sprintf("%sfont-weight: %s;\n", g.config.Indent, fontFace.FontWeight))
		}
	}

	if fontFace.FontStyle != "" {
		if g.config.Minify {
			css.WriteString(fmt.Sprintf("font-style:%s;", fontFace.FontStyle))
		} else {
			css.WriteString(fmt.Sprintf("%sfont-style: %s;\n", g.config.Indent, fontFace.FontStyle))
		}
	}

	if fontFace.FontDisplay != "" {
		if g.config.Minify {
			css.WriteString(fmt.Sprintf("font-display:%s;", fontFace.FontDisplay))
		} else {
			css.WriteString(fmt.Sprintf("%sfont-display: %s;\n", g.config.Indent, fontFace.FontDisplay))
		}
	}

	if g.config.Minify {
		css.WriteString("}")
	} else {
		css.WriteString("}")
	}

	return css.String()
}

// generateKeyframes generates CSS for @keyframes
func (g *CSSGenerator) generateKeyframes(name string, keyframes []CSSKeyframe) string {
	var css strings.Builder

	if g.config.Minify {
		css.WriteString(fmt.Sprintf("@keyframes %s{", name))
	} else {
		css.WriteString(fmt.Sprintf("@keyframes %s {\n", name))
	}

	for _, keyframe := range keyframes {
		if g.config.Minify {
			css.WriteString(fmt.Sprintf("%s{", keyframe.Position))
		} else {
			css.WriteString(fmt.Sprintf("%s%s {\n", g.config.Indent, keyframe.Position))
		}

		// Sort declarations
		var properties []string
		for prop := range keyframe.Declarations {
			properties = append(properties, prop)
		}
		sort.Strings(properties)

		for _, prop := range properties {
			value := keyframe.Declarations[prop]
			if g.config.Minify {
				css.WriteString(fmt.Sprintf("%s:%s;", prop, value))
			} else {
				css.WriteString(fmt.Sprintf("%s%s%s: %s;\n", g.config.Indent, g.config.Indent, prop, value))
			}
		}

		if g.config.Minify {
			css.WriteString("}")
		} else {
			css.WriteString(fmt.Sprintf("%s}\n", g.config.Indent))
		}
	}

	if g.config.Minify {
		css.WriteString("}")
	} else {
		css.WriteString("}")
	}

	return css.String()
}

// autoPrefixValue adds vendor prefixes if needed
func (g *CSSGenerator) autoPrefixValue(property, value string) string {
	// Simple auto-prefixing for common properties
	// In a real implementation, you'd want a more comprehensive solution
	prefixMap := map[string][]string{
		"transform":       {"-webkit-", "-moz-", "-ms-", ""},
		"transition":      {"-webkit-", "-moz-", "-o-", ""},
		"box-shadow":      {"-webkit-", "-moz-", ""},
		"border-radius":   {"-webkit-", "-moz-", ""},
		"user-select":     {"-webkit-", "-moz-", "-ms-", ""},
		"appearance":      {"-webkit-", "-moz-", ""},
		"backdrop-filter": {"-webkit-", ""},
		"clip-path":       {"-webkit-", ""},
	}

	if prefixes, exists := prefixMap[property]; exists {
		// For now, just return the standard property
		// In a full implementation, you'd generate multiple rules
		_ = prefixes
		return value
	}

	return value
}

// addIndent adds indentation to CSS content
func (g *CSSGenerator) addIndent(content string) string {
	if g.config.Minify {
		return content
	}

	lines := strings.Split(content, "\n")
	var indented []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			indented = append(indented, g.config.Indent+line)
		} else {
			indented = append(indented, line)
		}
	}

	return strings.Join(indented, "\n")
}

// Reset clears the current stylesheet
func (g *CSSGenerator) Reset() *CSSGenerator {
	g.stylesheet = &CSSStylesheet{
		Rules:     make([]CSSRule, 0),
		Variables: make(map[string]string),
		Keyframes: make(map[string][]CSSKeyframe),
		FontFaces: make([]CSSFontFace, 0),
		Imports:   make([]string, 0),
	}
	return g
}

// GetRuleCount returns the number of CSS rules
func (g *CSSGenerator) GetRuleCount() int {
	return len(g.stylesheet.Rules)
}

// GetVariableCount returns the number of CSS variables
func (g *CSSGenerator) GetVariableCount() int {
	return len(g.stylesheet.Variables)
}

// Clone creates a copy of the generator
func (g *CSSGenerator) Clone() *CSSGenerator {
	clone := NewCSSGeneratorWithConfig(g.config)

	// Copy rules
	clone.stylesheet.Rules = make([]CSSRule, len(g.stylesheet.Rules))
	copy(clone.stylesheet.Rules, g.stylesheet.Rules)

	// Copy variables
	for name, value := range g.stylesheet.Variables {
		clone.stylesheet.Variables[name] = value
	}

	// Copy keyframes
	for name, keyframes := range g.stylesheet.Keyframes {
		clone.stylesheet.Keyframes[name] = make([]CSSKeyframe, len(keyframes))
		copy(clone.stylesheet.Keyframes[name], keyframes)
	}

	// Copy font faces
	clone.stylesheet.FontFaces = make([]CSSFontFace, len(g.stylesheet.FontFaces))
	copy(clone.stylesheet.FontFaces, g.stylesheet.FontFaces)

	// Copy imports
	clone.stylesheet.Imports = make([]string, len(g.stylesheet.Imports))
	copy(clone.stylesheet.Imports, g.stylesheet.Imports)

	return clone
}
