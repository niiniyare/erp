// Package components - EditorControlSchema for rich text and code editors
// Based on JSON schema: EditorControlSchema.json  
package components

import (
	"encoding/json"
	"fmt"
)

// EditorControlSize represents the size options for editor controls
type EditorControlSize string

const (
	EditorControlSizeSM  EditorControlSize = "sm"  // Small editor
	EditorControlSizeMD  EditorControlSize = "md"  // Medium editor (default)
	EditorControlSizeLG  EditorControlSize = "lg"  // Large editor
	EditorControlSizeXL  EditorControlSize = "xl"  // Extra large editor
	EditorControlSizeXXL EditorControlSize = "xxl" // Extra extra large editor
)

// EditorControlType represents the type of editor to use
type EditorControlType string

const (
	EditorTypeMonaco   EditorControlType = "monaco"   // Monaco Editor (VS Code)
	EditorTypeTinyMCE  EditorControlType = "tinymce"  // TinyMCE Rich Text Editor
	EditorTypeCodeMirror EditorControlType = "codemirror" // CodeMirror Editor
	EditorTypeQuill    EditorControlType = "quill"    // Quill Rich Text Editor
	EditorTypeAce      EditorControlType = "ace"      // Ace Editor
	EditorTypeDefault  EditorControlType = "default"  // Default system editor
)

// EditorLanguage represents supported programming languages for syntax highlighting
type EditorLanguage string

const (
	// Programming Languages
	LanguageBat          EditorLanguage = "bat"          // Batch files
	LanguageC            EditorLanguage = "c"            // C programming
	LanguageCoffeescript EditorLanguage = "coffeescript" // CoffeeScript
	LanguageCPP          EditorLanguage = "cpp"          // C++ programming
	LanguageCSharp       EditorLanguage = "csharp"       // C# programming
	LanguageDockerfile   EditorLanguage = "dockerfile"   // Docker files
	LanguageFSharp       EditorLanguage = "fsharp"       // F# programming
	LanguageGo           EditorLanguage = "go"           // Go programming
	LanguageJava         EditorLanguage = "java"         // Java programming
	LanguageJavaScript   EditorLanguage = "javascript"   // JavaScript
	LanguageLua          EditorLanguage = "lua"          // Lua scripting
	LanguageObjectiveC   EditorLanguage = "objective-c"  // Objective-C
	LanguagePHP          EditorLanguage = "php"          // PHP programming
	LanguagePython       EditorLanguage = "python"       // Python programming
	LanguageR            EditorLanguage = "r"            // R statistical
	LanguagePowerShell   EditorLanguage = "powershell"   // PowerShell scripting
	
	// Web Technologies
	LanguageCSS        EditorLanguage = "css"        // CSS stylesheets
	LanguageHTML       EditorLanguage = "html"       // HTML markup
	LanguageLess       EditorLanguage = "less"       // Less CSS preprocessor
	LanguageHandlebars EditorLanguage = "handlebars" // Handlebars templates
	LanguagePug        EditorLanguage = "pug"        // Pug templates
	LanguageRazor      EditorLanguage = "razor"      // Razor templates
	
	// Data Formats
	LanguageJSON     EditorLanguage = "json"     // JSON data
	LanguageXML      EditorLanguage = "xml"      // XML markup
	LanguageYAML     EditorLanguage = "yaml"     // YAML data
	LanguageINI      EditorLanguage = "ini"      // INI configuration
	LanguageMarkdown EditorLanguage = "markdown" // Markdown text
	
	// Other
	LanguagePlaintext EditorLanguage = "plaintext" // Plain text
	LanguageSQL       EditorLanguage = "sql"       // SQL queries
	LanguageShell     EditorLanguage = "shell"     // Shell scripts
)

// EditorVendor represents different editor implementations
type EditorVendor string

const (
	VendorMonaco    EditorVendor = "monaco"     // Monaco Editor
	VendorTinyMCE   EditorVendor = "tinymce"    // TinyMCE
	VendorCodeMirror EditorVendor = "codemirror" // CodeMirror
	VendorQuill     EditorVendor = "quill"      // Quill
	VendorAce       EditorVendor = "ace"        // Ace Editor
)

// EditorToolbarItem represents toolbar buttons for rich text editors
type EditorToolbarItem string

const (
	ToolbarBold          EditorToolbarItem = "bold"          // Bold formatting
	ToolbarItalic        EditorToolbarItem = "italic"        // Italic formatting
	ToolbarUnderline     EditorToolbarItem = "underline"     // Underline formatting
	ToolbarStrikethrough EditorToolbarItem = "strikethrough" // Strikethrough formatting
	ToolbarHeading       EditorToolbarItem = "heading"       // Heading levels
	ToolbarQuote         EditorToolbarItem = "quote"         // Block quote
	ToolbarCode          EditorToolbarItem = "code"          // Inline code
	ToolbarCodeBlock     EditorToolbarItem = "code-block"    // Code block
	ToolbarUnorderedList EditorToolbarItem = "unordered-list" // Bullet list
	ToolbarOrderedList   EditorToolbarItem = "ordered-list"  // Numbered list
	ToolbarLink          EditorToolbarItem = "link"          // Insert link
	ToolbarImage         EditorToolbarItem = "image"         // Insert image
	ToolbarTable         EditorToolbarItem = "table"         // Insert table
	ToolbarHorizontalRule EditorToolbarItem = "horizontal-rule" // Horizontal line
	ToolbarUndo          EditorToolbarItem = "undo"          // Undo action
	ToolbarRedo          EditorToolbarItem = "redo"          // Redo action
	ToolbarFullscreen    EditorToolbarItem = "fullscreen"    // Fullscreen toggle
)

// EditorOptions represents configuration options for different editor types
type EditorOptions struct {
	// Monaco Editor Options
	Theme                   string            `json:"theme,omitempty"`                   // Editor theme
	WordWrap               string            `json:"wordWrap,omitempty"`               // Word wrapping: "on", "off", "wordWrapColumn", "bounded"
	LineNumbers            string            `json:"lineNumbers,omitempty"`            // Line numbers: "on", "off", "relative", "interval"
	MiniMap                *MinimapOptions   `json:"minimap,omitempty"`                // Minimap configuration
	FontSize               int               `json:"fontSize,omitempty"`               // Font size in pixels
	FontFamily             string            `json:"fontFamily,omitempty"`             // Font family
	TabSize                int               `json:"tabSize,omitempty"`                // Tab size
	InsertSpaces           bool              `json:"insertSpaces,omitempty"`           // Use spaces instead of tabs
	FoldingStrategy        string            `json:"foldingStrategy,omitempty"`        // Code folding strategy
	FormatOnPaste          bool              `json:"formatOnPaste,omitempty"`          // Format on paste
	FormatOnType           bool              `json:"formatOnType,omitempty"`           // Format on type
	AutoIndent             string            `json:"autoIndent,omitempty"`             // Auto indent: "none", "keep", "brackets", "advanced", "full"
	
	// Rich Text Editor Options  
	MenuBar                bool              `json:"menubar,omitempty"`                // Show menu bar
	Toolbar                interface{}       `json:"toolbar,omitempty"`                // Toolbar configuration (string or array)
	Plugins                []string          `json:"plugins,omitempty"`                // Enabled plugins
	StatusBar              bool              `json:"statusbar,omitempty"`              // Show status bar
	Branding               bool              `json:"branding,omitempty"`               // Show editor branding
	ElementPath            bool              `json:"elementpath,omitempty"`            // Show element path
	Resize                 string            `json:"resize,omitempty"`                 // Resize behavior: "true", "false", "both"
	
	// Behavior Options
	AutoSave               bool              `json:"autoSave,omitempty"`               // Enable auto-save
	AutoSaveInterval       int               `json:"autoSaveInterval,omitempty"`       // Auto-save interval in milliseconds
	Placeholder            string            `json:"placeholder,omitempty"`            // Placeholder text
	ReadOnly               bool              `json:"readOnly,omitempty"`               // Read-only mode
	SpellCheck             bool              `json:"spellCheck,omitempty"`             // Spell checking
	
	// Advanced Options
	CustomOptions          map[string]interface{} `json:"customOptions,omitempty"`     // Custom editor-specific options
}

// MinimapOptions represents Monaco Editor minimap configuration
type MinimapOptions struct {
	Enabled        bool   `json:"enabled"`                  // Enable minimap
	Side           string `json:"side,omitempty"`           // Side: "left", "right"
	ShowSlider     string `json:"showSlider,omitempty"`     // Show slider: "always", "mouseover"
	RenderCharacters bool  `json:"renderCharacters,omitempty"` // Render actual characters
	MaxColumn      int    `json:"maxColumn,omitempty"`      // Maximum column to render
	Scale          int    `json:"scale,omitempty"`          // Scale factor
}

// EditorControlSchema represents a rich text or code editor control
// This component provides advanced text editing capabilities with syntax highlighting,
// formatting tools, and extensible plugin architecture
// Based on AMis Editor control: https://aisuda.bce.baidu.com/amis/zh-CN/components/form/editor
type EditorControlSchema struct {
	BaseComponentProps

	// Component type identifier (required)
	Type string `json:"type"` // Must be "editor"

	// Field name for form submission, supports multi-level paths (e.g., "a.b.c") (required)
	Name string `json:"name"`

	// Display label for the editor control
	Label string `json:"label,omitempty"`

	// Default content value
	Value string `json:"value,omitempty"`

	// Placeholder text shown when editor is empty
	Placeholder string `json:"placeholder,omitempty"`

	// Size of the editor control
	Size EditorControlSize `json:"size,omitempty"`

	// Whether the field is required
	Required bool `json:"required,omitempty"`

	// Whether the control is read-only
	ReadOnly bool `json:"readOnly,omitempty"`

	// Editor Configuration
	// Programming language for syntax highlighting
	Language EditorLanguage `json:"language,omitempty"`
	
	// Editor implementation to use
	EditorType EditorControlType `json:"editorType,omitempty"`
	
	// Editor vendor/implementation
	Vendor EditorVendor `json:"vendor,omitempty"`
	
	// Whether to allow fullscreen mode
	AllowFullScreen bool `json:"allowFullscreen,omitempty"`
	
	// Editor-specific configuration options
	Options *EditorOptions `json:"options,omitempty"`

	// Display Configuration
	// Height of the editor (pixels or CSS value)
	Height any `json:"height,omitempty"` // number or string
	
	// Width of the editor (pixels or CSS value) 
	Width any `json:"width,omitempty"` // number or string

	// Code Editor Features
	// Whether to enable word wrapping
	WordWrap bool `json:"wordWrap,omitempty"`
	
	// Whether to show code folding gutter
	FoldGutter bool `json:"foldGutter,omitempty"`
	
	// Whether to show line numbers
	LineNumbers bool `json:"lineNumbers,omitempty"`
	
	// Whether to enable auto-completion
	AutoComplete bool `json:"autoComplete,omitempty"`
	
	// Whether to show minimap (Monaco Editor)
	MiniMap bool `json:"miniMap,omitempty"`
	
	// Whether line numbers are selectable
	SelectOnLineNumbers bool `json:"selectOnLineNumbers,omitempty"`
	
	// Message shown in read-only mode
	ReadOnlyMessage string `json:"readOnlyMessage,omitempty"`
	
	// Whether to allow scrolling beyond last line
	ScrollBeyondLastLine bool `json:"scrollBeyondLastLine,omitempty"`
	
	// Whether to auto-focus editor on load
	AutoFocus bool `json:"autoFocus,omitempty"`

	// Rich Text Editor Features
	// Toolbar buttons configuration
	Toolbar []EditorToolbarItem `json:"toolbar,omitempty"`
	
	// Whether to show status bar
	StatusBar bool `json:"statusBar,omitempty"`
	
	// Whether to enable spell checking
	SpellChecker bool `json:"spellChecker,omitempty"`
	
	// Whether editor height grows with content
	AutoGrow bool `json:"autoGrow,omitempty"`
	
	// Minimum height when auto-grow is enabled
	AutoGrowMinHeight int `json:"autoGrowMinHeight,omitempty"`
	
	// Maximum height when auto-grow is enabled
	AutoGrowMaxHeight int `json:"autoGrowMaxHeight,omitempty"`

	// Content Validation
	// Maximum character length
	MaxLength int `json:"maxLength,omitempty"`
	
	// Whether to show word/character count
	ShowWordCount bool `json:"showWordCount,omitempty"`
	
	// Allowed HTML tags for rich text (security)
	AllowedTags []string `json:"allowedTags,omitempty"`
	
	// Whether to convert Markdown syntax automatically
	ConvertMarkdown bool `json:"convertMarkdown,omitempty"`

	// Form Integration Properties
	// Whether to submit form when content changes
	SubmitOnChange bool `json:"submitOnChange,omitempty"`
	
	// Whether to validate on every change
	ValidateOnChange bool `json:"validateOnChange,omitempty"`
	
	// Whether to clear value when form item is hidden
	ClearValueOnHidden bool `json:"clearValueOnHidden,omitempty"`
	
	// Remote validation API
	ValidateAPI any `json:"validateApi,omitempty"`

	// Auto-fill Configuration
	AutoFill map[string]string `json:"autoFill,omitempty"`
	InitAutoFill bool `json:"initAutoFill,omitempty"`

	// Advanced Properties
	// Description content supporting HTML
	Description string `json:"description,omitempty"`
	// Input hint shown on focus
	Hint string `json:"hint,omitempty"`
	// Custom validation error messages
	ValidationErrors map[string]string `json:"validationErrors,omitempty"`
	// Validation rules configuration
	Validations any `json:"validations,omitempty"`

	// Static Display Properties
	Static bool `json:"static,omitempty"`
	StaticOn string `json:"staticOn,omitempty"`
	StaticPlaceholder string `json:"staticPlaceholder,omitempty"`
	StaticClassName string `json:"staticClassName,omitempty"`
	StaticLabelClassName string `json:"staticLabelClassName,omitempty"`
	StaticInputClassName string `json:"staticInputClassName,omitempty"`
	StaticSchema any `json:"staticSchema,omitempty"`

	// CSS Classes for styling
	InputClassName string `json:"inputClassName,omitempty"`
	LabelClassName string `json:"labelClassName,omitempty"`
	DescriptionClassName string `json:"descriptionClassName,omitempty"`

	// Callbacks
	// JavaScript callback when editor is mounted
	EditorDidMount string `json:"editorDidMount,omitempty"`

	// Design-time Configuration
	EditorSetting *EditorSetting `json:"editorSetting,omitempty"`
}

// Factory function to create a basic code editor
func NewCodeEditor(name string, language EditorLanguage) *EditorControlSchema {
	return &EditorControlSchema{
		Type:                 "editor",
		Name:                 name,
		EditorType:           EditorTypeMonaco,
		Language:             language,
		Size:                 EditorControlSizeMD,
		Height:               400,
		WordWrap:             true,
		LineNumbers:          true,
		AutoComplete:         true,
		MiniMap:              false,
		AllowFullScreen:      true,
		AutoFocus:            false,
		ScrollBeyondLastLine: false,
		Options: &EditorOptions{
			Theme:          "vs-dark",
			TabSize:        2,
			InsertSpaces:   true,
			AutoIndent:     "advanced",
			FormatOnPaste:  true,
			FormatOnType:   true,
		},
	}
}

// Factory function to create a rich text editor
func NewRichTextEditor(name string) *EditorControlSchema {
	return &EditorControlSchema{
		Type:           "editor",
		Name:           name,
		EditorType:     EditorTypeTinyMCE,
		Size:           EditorControlSizeMD,
		Height:         300,
		AllowFullScreen: true,
		StatusBar:      true,
		SpellChecker:   true,
		AutoGrow:       true,
		AutoGrowMinHeight: 200,
		AutoGrowMaxHeight: 600,
		ShowWordCount:  true,
		Toolbar: []EditorToolbarItem{
			ToolbarBold, ToolbarItalic, ToolbarUnderline,
			ToolbarHeading, ToolbarQuote, ToolbarCode,
			ToolbarUnorderedList, ToolbarOrderedList,
			ToolbarLink, ToolbarImage, ToolbarTable,
			ToolbarUndo, ToolbarRedo, ToolbarFullscreen,
		},
		Options: &EditorOptions{
			MenuBar:     true,
			StatusBar:   true,
			Branding:    false,
			ElementPath: false,
			Resize:      "both",
			Plugins: []string{
				"advlist", "autolink", "lists", "link", "image", "charmap",
				"preview", "anchor", "searchreplace", "visualblocks", "code",
				"fullscreen", "insertdatetime", "media", "table", "help", "wordcount",
			},
		},
	}
}

// Factory function to create a JSON editor with validation
func NewJSONEditor(name string) *EditorControlSchema {
	editor := NewCodeEditor(name, LanguageJSON)
	editor.Options.FormatOnPaste = true
	editor.Options.FormatOnType = true
	editor.ConvertMarkdown = false
	editor.MaxLength = 0 // No limit for JSON
	return editor
}

// Factory function to create a Markdown editor
func NewMarkdownEditor(name string) *EditorControlSchema {
	return &EditorControlSchema{
		Type:            "editor",
		Name:            name,
		EditorType:      EditorTypeMonaco,
		Language:        LanguageMarkdown,
		Size:            EditorControlSizeLG,
		Height:          500,
		WordWrap:        true,
		LineNumbers:     true,
		AllowFullScreen: true,
		ConvertMarkdown: true,
		ShowWordCount:   true,
		Toolbar: []EditorToolbarItem{
			ToolbarBold, ToolbarItalic, ToolbarHeading,
			ToolbarQuote, ToolbarCode, ToolbarCodeBlock,
			ToolbarUnorderedList, ToolbarOrderedList,
			ToolbarLink, ToolbarImage, ToolbarTable,
			ToolbarFullscreen,
		},
		Options: &EditorOptions{
			Theme:         "vs",
			TabSize:       2,
			InsertSpaces:  true,
			WordWrap:      "on",
			AutoIndent:    "advanced",
			FormatOnPaste: false, // Don't format markdown
		},
	}
}

// Validation function for EditorControlSchema
func (e *EditorControlSchema) Validate() error {
	if e.Type != "editor" {
		return fmt.Errorf("invalid editor control type: %s, must be 'editor'", e.Type)
	}
	if e.Name == "" {
		return fmt.Errorf("editor control name is required")
	}

	// Validate enum values
	if e.Size != "" &&
		e.Size != EditorControlSizeSM &&
		e.Size != EditorControlSizeMD &&
		e.Size != EditorControlSizeLG &&
		e.Size != EditorControlSizeXL &&
		e.Size != EditorControlSizeXXL {
		return fmt.Errorf("invalid size: %s", e.Size)
	}

	if e.EditorType != "" &&
		e.EditorType != EditorTypeMonaco &&
		e.EditorType != EditorTypeTinyMCE &&
		e.EditorType != EditorTypeCodeMirror &&
		e.EditorType != EditorTypeQuill &&
		e.EditorType != EditorTypeAce &&
		e.EditorType != EditorTypeDefault {
		return fmt.Errorf("invalid editorType: %s", e.EditorType)
	}

	if e.Vendor != "" &&
		e.Vendor != VendorMonaco &&
		e.Vendor != VendorTinyMCE &&
		e.Vendor != VendorCodeMirror &&
		e.Vendor != VendorQuill &&
		e.Vendor != VendorAce {
		return fmt.Errorf("invalid vendor: %s", e.Vendor)
	}

	// Validate height constraints for auto-grow
	if e.AutoGrow && e.AutoGrowMaxHeight > 0 && e.AutoGrowMinHeight > 0 {
		if e.AutoGrowMinHeight >= e.AutoGrowMaxHeight {
			return fmt.Errorf("autoGrowMinHeight (%d) must be less than autoGrowMaxHeight (%d)", 
				e.AutoGrowMinHeight, e.AutoGrowMaxHeight)
		}
	}

	return nil
}

// ToJSON converts EditorControlSchema to JSON string
func (e *EditorControlSchema) ToJSON() (string, error) {
	if err := e.Validate(); err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal EditorControlSchema to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON creates EditorControlSchema from JSON string
func EditorControlFromJSON(jsonData string) (*EditorControlSchema, error) {
	var config EditorControlSchema
	err := json.Unmarshal([]byte(jsonData), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to EditorControlSchema: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &config, nil
}