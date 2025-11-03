package schema

import (
	"context"
	"fmt"
	"html"
	"strings"
)

// DateRenderer handles date input fields
type DateRenderer struct {
	*BaseRenderer
}

func NewDateRenderer(base *BaseRenderer) *DateRenderer {
	return &DateRenderer{BaseRenderer: base}
}

func (dr *DateRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := dr.buildInputAttributes(field, value)
	attrs["type"] = "date"

	inputHTML := fmt.Sprintf(`<input %s>`, dr.attributesToString(attrs))
	return dr.RenderContainer(field, inputHTML, errors), nil
}

func (dr *DateRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldDate
}

func (dr *DateRenderer) GetRequiredAssets() []string {
	return []string{"css/date.css"}
}

// TimeRenderer handles time input fields
type TimeRenderer struct {
	*BaseRenderer
}

func NewTimeRenderer(base *BaseRenderer) *TimeRenderer {
	return &TimeRenderer{BaseRenderer: base}
}

func (tr *TimeRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := tr.buildInputAttributes(field, value)
	attrs["type"] = "time"

	inputHTML := fmt.Sprintf(`<input %s>`, tr.attributesToString(attrs))
	return tr.RenderContainer(field, inputHTML, errors), nil
}

func (tr *TimeRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldTime
}

func (tr *TimeRenderer) GetRequiredAssets() []string {
	return []string{"css/time.css"}
}

// DateTimeRenderer handles datetime-local input fields
type DateTimeRenderer struct {
	*BaseRenderer
}

func NewDateTimeRenderer(base *BaseRenderer) *DateTimeRenderer {
	return &DateTimeRenderer{BaseRenderer: base}
}

func (dtr *DateTimeRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := dtr.buildInputAttributes(field, value)
	attrs["type"] = "datetime-local"

	inputHTML := fmt.Sprintf(`<input %s>`, dtr.attributesToString(attrs))
	return dtr.RenderContainer(field, inputHTML, errors), nil
}

func (dtr *DateTimeRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldDateTime
}

func (dtr *DateTimeRenderer) GetRequiredAssets() []string {
	return []string{"css/datetime.css"}
}

// DateRangeRenderer handles date range fields
type DateRangeRenderer struct {
	*BaseRenderer
}

func NewDateRangeRenderer(base *BaseRenderer) *DateRangeRenderer {
	return &DateRangeRenderer{BaseRenderer: base}
}

func (drr *DateRangeRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	// Parse date range value
	var startDate, endDate string
	if value != nil {
		switch v := value.(type) {
		case map[string]any:
			if start, exists := v["start"]; exists {
				startDate = fmt.Sprintf("%v", start)
			}
			if end, exists := v["end"]; exists {
				endDate = fmt.Sprintf("%v", end)
			}
		case string:
			// Handle comma-separated values
			if v != "" {
				parts := strings.Split(v, ",")
				if len(parts) >= 1 {
					startDate = strings.TrimSpace(parts[0])
				}
				if len(parts) >= 2 {
					endDate = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	startAttrs := map[string]string{
		"type":  "date",
		"id":    field.Name + "-start",
		"name":  field.Name + "[start]",
		"class": "date-range-input date-range-start",
		"value": startDate,
	}

	endAttrs := map[string]string{
		"type":  "date",
		"id":    field.Name + "-end",
		"name":  field.Name + "[end]",
		"class": "date-range-input date-range-end",
		"value": endDate,
	}

	if field.Required {
		startAttrs["required"] = "required"
		endAttrs["required"] = "required"
	}

	if field.Disabled {
		startAttrs["disabled"] = "disabled"
		endAttrs["disabled"] = "disabled"
	}

	inputHTML := fmt.Sprintf(`
		<div class="date-range-container">
			<input %s placeholder="Start date">
			<span class="date-range-separator">to</span>
			<input %s placeholder="End date">
		</div>`, drr.attributesToString(startAttrs), drr.attributesToString(endAttrs))

	return drr.RenderContainer(field, inputHTML, errors), nil
}

func (drr *DateRangeRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldDateRange
}

func (drr *DateRangeRenderer) GetRequiredAssets() []string {
	return []string{"css/date-range.css"}
}

// FileRenderer handles file upload fields
type FileRenderer struct {
	*BaseRenderer
}

func NewFileRenderer(base *BaseRenderer) *FileRenderer {
	return &FileRenderer{BaseRenderer: base}
}

func (fr *FileRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := fr.buildInputAttributes(field, value)
	attrs["type"] = "file"
	delete(attrs, "value") // File inputs don't have values

	// Configuration
	multiple := false
	accept := ""
	maxSize := ""

	if field.Config != nil {
		if m, exists := field.Config["multiple"]; exists {
			multiple = m.(bool)
		}
		if a, exists := field.Config["accept"]; exists {
			accept = fmt.Sprintf("%v", a)
		}
		if size, exists := field.Config["maxSize"]; exists {
			maxSize = fmt.Sprintf("%v", size)
		}
	}

	if multiple {
		attrs["multiple"] = "multiple"
	}
	if accept != "" {
		attrs["accept"] = accept
	}
	if maxSize != "" {
		attrs["data-max-size"] = maxSize
	}

	inputHTML := fmt.Sprintf(`
		<div class="file-upload">
			<input %s onchange="handleFileSelect(this)">
			<label for="%s" class="file-upload-label">
				<span class="file-upload-icon">📁</span>
				<span class="file-upload-text">Choose files or drag and drop</span>
			</label>
			<div class="file-upload-list" id="%s-list"></div>
		</div>`, fr.attributesToString(attrs), field.Name, field.Name)

	return fr.RenderContainer(field, inputHTML, errors), nil
}

func (fr *FileRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldFile
}

func (fr *FileRenderer) GetRequiredAssets() []string {
	return []string{"css/file-upload.css", "js/file-upload.js"}
}

// ImageRenderer handles image upload fields
type ImageRenderer struct {
	*BaseRenderer
}

func NewImageRenderer(base *BaseRenderer) *ImageRenderer {
	return &ImageRenderer{BaseRenderer: base}
}

func (ir *ImageRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := ir.buildInputAttributes(field, value)
	attrs["type"] = "file"
	attrs["accept"] = "image/*"
	delete(attrs, "value") // File inputs don't have values

	// Configuration
	maxSize := "5MB"
	showPreview := true

	if field.Config != nil {
		if size, exists := field.Config["maxSize"]; exists {
			maxSize = fmt.Sprintf("%v", size)
		}
		if preview, exists := field.Config["showPreview"]; exists {
			showPreview = preview.(bool)
		}
	}

	attrs["data-max-size"] = maxSize

	var previewHTML string
	if showPreview {
		previewHTML = fmt.Sprintf(`<div class="image-preview" id="%s-preview"></div>`, field.Name)
	}

	inputHTML := fmt.Sprintf(`
		<div class="image-upload">
			<input %s onchange="handleImageSelect(this)">
			<label for="%s" class="image-upload-label">
				<span class="image-upload-icon">🖼️</span>
				<span class="image-upload-text">Choose image or drag and drop</span>
			</label>
			%s
		</div>`, ir.attributesToString(attrs), field.Name, previewHTML)

	return ir.RenderContainer(field, inputHTML, errors), nil
}

func (ir *ImageRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldImage
}

func (ir *ImageRenderer) GetRequiredAssets() []string {
	return []string{"css/image-upload.css", "js/image-upload.js"}
}

// RichTextRenderer handles rich text editor fields
type RichTextRenderer struct {
	*BaseRenderer
}

func NewRichTextRenderer(base *BaseRenderer) *RichTextRenderer {
	return &RichTextRenderer{BaseRenderer: base}
}

func (rtr *RichTextRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	// Get the value to put inside textarea
	var textValue string
	if value != nil {
		textValue = fmt.Sprintf("%v", value)
	} else if field.Default != nil {
		textValue = fmt.Sprintf("%v", field.Default)
	}

	attrs := map[string]string{
		"id":    field.Name,
		"name":  field.Name,
		"class": "rich-text-editor",
	}

	if field.Required {
		attrs["required"] = "required"
	}

	// Configuration
	toolbar := "basic"
	height := "300"

	if field.Config != nil {
		if t, exists := field.Config["toolbar"]; exists {
			toolbar = fmt.Sprintf("%v", t)
		}
		if h, exists := field.Config["height"]; exists {
			height = fmt.Sprintf("%v", h)
		}
	}

	attrs["data-toolbar"] = toolbar
	attrs["data-height"] = height

	inputHTML := fmt.Sprintf(`
		<div class="rich-text-container">
			<textarea %s>%s</textarea>
		</div>`, rtr.attributesToString(attrs), html.EscapeString(textValue))

	return rtr.RenderContainer(field, inputHTML, errors), nil
}

func (rtr *RichTextRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldRichText
}

func (rtr *RichTextRenderer) GetRequiredAssets() []string {
	return []string{"css/rich-text.css", "js/rich-text-editor.js"}
}

// CodeRenderer handles code editor fields
type CodeRenderer struct {
	*BaseRenderer
}

func NewCodeRenderer(base *BaseRenderer) *CodeRenderer {
	return &CodeRenderer{BaseRenderer: base}
}

func (cr *CodeRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	// Get the value to put inside textarea
	var textValue string
	if value != nil {
		textValue = fmt.Sprintf("%v", value)
	} else if field.Default != nil {
		textValue = fmt.Sprintf("%v", field.Default)
	}

	attrs := map[string]string{
		"id":    field.Name,
		"name":  field.Name,
		"class": "code-editor",
	}

	if field.Required {
		attrs["required"] = "required"
	}

	// Configuration
	language := "javascript"
	theme := "light"
	lineNumbers := true

	if field.Config != nil {
		if lang, exists := field.Config["language"]; exists {
			language = fmt.Sprintf("%v", lang)
		}
		if t, exists := field.Config["theme"]; exists {
			theme = fmt.Sprintf("%v", t)
		}
		if ln, exists := field.Config["lineNumbers"]; exists {
			lineNumbers = ln.(bool)
		}
	}

	attrs["data-language"] = language
	attrs["data-theme"] = theme
	if lineNumbers {
		attrs["data-line-numbers"] = "true"
	}

	inputHTML := fmt.Sprintf(`
		<div class="code-editor-container">
			<textarea %s>%s</textarea>
		</div>`, cr.attributesToString(attrs), html.EscapeString(textValue))

	return cr.RenderContainer(field, inputHTML, errors), nil
}

func (cr *CodeRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldCode
}

func (cr *CodeRenderer) GetRequiredAssets() []string {
	return []string{"css/code-editor.css", "js/code-editor.js"}
}

// JSONRenderer handles JSON editor fields
type JSONRenderer struct {
	*BaseRenderer
}

func NewJSONRenderer(base *BaseRenderer) *JSONRenderer {
	return &JSONRenderer{BaseRenderer: base}
}

func (jr *JSONRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	// Get the value to put inside textarea
	var textValue string
	if value != nil {
		textValue = fmt.Sprintf("%v", value)
	} else if field.Default != nil {
		textValue = fmt.Sprintf("%v", field.Default)
	}

	attrs := map[string]string{
		"id":    field.Name,
		"name":  field.Name,
		"class": "json-editor",
	}

	if field.Required {
		attrs["required"] = "required"
	}

	// Configuration
	validateJSON := true
	formatJSON := true

	if field.Config != nil {
		if validate, exists := field.Config["validate"]; exists {
			validateJSON = validate.(bool)
		}
		if format, exists := field.Config["format"]; exists {
			formatJSON = format.(bool)
		}
	}

	if validateJSON {
		attrs["data-validate"] = "true"
	}
	if formatJSON {
		attrs["data-format"] = "true"
	}

	inputHTML := fmt.Sprintf(`
		<div class="json-editor-container">
			<textarea %s>%s</textarea>
			<div class="json-editor-toolbar">
				<button type="button" onclick="formatJSON('%s')">Format</button>
				<button type="button" onclick="validateJSON('%s')">Validate</button>
			</div>
		</div>`, jr.attributesToString(attrs), html.EscapeString(textValue), field.Name, field.Name)

	return jr.RenderContainer(field, inputHTML, errors), nil
}

func (jr *JSONRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldJSON
}

func (jr *JSONRenderer) GetRequiredAssets() []string {
	return []string{"css/json-editor.css", "js/json-editor.js"}
}

// Placeholder renderers for remaining field types
type (
	SignatureRenderer struct{ *BaseRenderer }
	LocationRenderer  struct{ *BaseRenderer }
	RelationRenderer  struct{ *BaseRenderer }
	DisplayRenderer   struct{ *BaseRenderer }
	DividerRenderer   struct{ *BaseRenderer }
	HTMLRenderer      struct{ *BaseRenderer }
)

func NewSignatureRenderer(base *BaseRenderer) *SignatureRenderer { return &SignatureRenderer{base} }
func NewLocationRenderer(base *BaseRenderer) *LocationRenderer   { return &LocationRenderer{base} }
func NewRelationRenderer(base *BaseRenderer) *RelationRenderer   { return &RelationRenderer{base} }
func NewDisplayRenderer(base *BaseRenderer) *DisplayRenderer     { return &DisplayRenderer{base} }
func NewDividerRenderer(base *BaseRenderer) *DividerRenderer     { return &DividerRenderer{base} }
func NewHTMLRenderer(base *BaseRenderer) *HTMLRenderer           { return &HTMLRenderer{base} }

func (sr *SignatureRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	inputHTML := fmt.Sprintf(`<div class="signature-pad" data-field="%s">Signature field - implementation needed</div>`, field.Name)
	return sr.RenderContainer(field, inputHTML, errors), nil
}

func (sr *SignatureRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldSignature
}

func (sr *SignatureRenderer) GetRequiredAssets() []string {
	return []string{"css/signature.css", "js/signature.js"}
}

func (lr *LocationRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	inputHTML := fmt.Sprintf(`<div class="location-picker" data-field="%s">Location picker - implementation needed</div>`, field.Name)
	return lr.RenderContainer(field, inputHTML, errors), nil
}

func (lr *LocationRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldLocation
}

func (lr *LocationRenderer) GetRequiredAssets() []string {
	return []string{"css/location.css", "js/location.js"}
}

func (rr *RelationRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	inputHTML := fmt.Sprintf(`<div class="relation-picker" data-field="%s">Relation picker - implementation needed</div>`, field.Name)
	return rr.RenderContainer(field, inputHTML, errors), nil
}

func (rr *RelationRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldRelation
}

func (rr *RelationRenderer) GetRequiredAssets() []string {
	return []string{"css/relation.css", "js/relation.js"}
}

func (dr *DisplayRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	displayValue := ""
	if value != nil {
		displayValue = fmt.Sprintf("%v", value)
	}
	return fmt.Sprintf(`<div class="field-display">%s</div>`, html.EscapeString(displayValue)), nil
}

func (dr *DisplayRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldDisplay
}
func (dr *DisplayRenderer) GetRequiredAssets() []string { return []string{"css/display.css"} }

func (dr *DividerRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	return `<hr class="field-divider">`, nil
}

func (dr *DividerRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldDivider
}
func (dr *DividerRenderer) GetRequiredAssets() []string { return []string{"css/divider.css"} }

func (hr *HTMLRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	htmlContent := ""
	if field.Default != nil {
		htmlContent = fmt.Sprintf("%v", field.Default)
	}
	return fmt.Sprintf(`<div class="field-html">%s</div>`, htmlContent), nil
}
func (hr *HTMLRenderer) SupportsFieldType(fieldType FieldType) bool { return fieldType == FieldHTML }
func (hr *HTMLRenderer) GetRequiredAssets() []string                { return []string{"css/html-field.css"} }
