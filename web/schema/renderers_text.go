package schema

import (
	"context"
	"fmt"
)

// TextRenderer handles text input fields
type TextRenderer struct {
	*BaseRenderer
}

func NewTextRenderer(base *BaseRenderer) *TextRenderer {
	return &TextRenderer{BaseRenderer: base}
}

func (tr *TextRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := tr.buildInputAttributes(field, value)
	attrs["type"] = "text"

	// Apply input masking if specified
	if field.Mask != nil {
		attrs["data-mask"] = field.Mask.Pattern
		if field.Mask.Placeholder != "" {
			attrs["data-mask-placeholder"] = field.Mask.Placeholder
		}
	}

	inputHTML := fmt.Sprintf(`<input %s>`, tr.attributesToString(attrs))
	return tr.RenderContainer(field, inputHTML, errors), nil
}

func (tr *TextRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldText
}

func (tr *TextRenderer) GetRequiredAssets() []string {
	return []string{"css/input.css"}
}

// EmailRenderer handles email input fields
type EmailRenderer struct {
	*BaseRenderer
}

func NewEmailRenderer(base *BaseRenderer) *EmailRenderer {
	return &EmailRenderer{BaseRenderer: base}
}

func (er *EmailRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := er.buildInputAttributes(field, value)
	attrs["type"] = "email"
	attrs["autocomplete"] = "email"

	inputHTML := fmt.Sprintf(`<input %s>`, er.attributesToString(attrs))
	return er.RenderContainer(field, inputHTML, errors), nil
}

func (er *EmailRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldEmail
}

func (er *EmailRenderer) GetRequiredAssets() []string {
	return []string{"css/input.css"}
}

// PasswordRenderer handles password input fields
type PasswordRenderer struct {
	*BaseRenderer
}

func NewPasswordRenderer(base *BaseRenderer) *PasswordRenderer {
	return &PasswordRenderer{BaseRenderer: base}
}

func (pr *PasswordRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := pr.buildInputAttributes(field, value)
	attrs["type"] = "password"
	attrs["autocomplete"] = "current-password"

	// Add password visibility toggle if enabled
	var inputHTML string
	if field.Config != nil {
		if showToggle, exists := field.Config["showToggle"]; exists && showToggle == true {
			inputHTML = fmt.Sprintf(`
				<div class="password-field">
					<input %s>
					<button type="button" class="password-toggle" aria-label="Toggle password visibility" onclick="togglePasswordVisibility('%s')">
						<span class="password-toggle-icon" aria-hidden="true">👁</span>
					</button>
				</div>`, pr.attributesToString(attrs), field.Name)
		} else {
			inputHTML = fmt.Sprintf(`<input %s>`, pr.attributesToString(attrs))
		}
	} else {
		inputHTML = fmt.Sprintf(`<input %s>`, pr.attributesToString(attrs))
	}

	return pr.RenderContainer(field, inputHTML, errors), nil
}

func (pr *PasswordRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldPassword
}

func (pr *PasswordRenderer) GetRequiredAssets() []string {
	return []string{"css/input.css", "js/password-toggle.js"}
}

// URLRenderer handles URL input fields
type URLRenderer struct {
	*BaseRenderer
}

func NewURLRenderer(base *BaseRenderer) *URLRenderer {
	return &URLRenderer{BaseRenderer: base}
}

func (ur *URLRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := ur.buildInputAttributes(field, value)
	attrs["type"] = "url"
	attrs["autocomplete"] = "url"

	inputHTML := fmt.Sprintf(`<input %s>`, ur.attributesToString(attrs))
	return ur.RenderContainer(field, inputHTML, errors), nil
}

func (ur *URLRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldURL
}

func (ur *URLRenderer) GetRequiredAssets() []string {
	return []string{"css/input.css"}
}

// PhoneRenderer handles phone input fields
type PhoneRenderer struct {
	*BaseRenderer
}

func NewPhoneRenderer(base *BaseRenderer) *PhoneRenderer {
	return &PhoneRenderer{BaseRenderer: base}
}

func (pr *PhoneRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := pr.buildInputAttributes(field, value)
	attrs["type"] = "tel"
	attrs["autocomplete"] = "tel"

	// Add phone number formatting
	if field.Mask == nil {
		attrs["data-mask"] = "(000) 000-0000"
		attrs["data-mask-placeholder"] = "(___) ___-____"
	}

	inputHTML := fmt.Sprintf(`<input %s>`, pr.attributesToString(attrs))
	return pr.RenderContainer(field, inputHTML, errors), nil
}

func (pr *PhoneRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldPhone
}

func (pr *PhoneRenderer) GetRequiredAssets() []string {
	return []string{"css/input.css", "js/phone-mask.js"}
}

// HiddenRenderer handles hidden input fields
type HiddenRenderer struct {
	*BaseRenderer
}

func NewHiddenRenderer(base *BaseRenderer) *HiddenRenderer {
	return &HiddenRenderer{BaseRenderer: base}
}

func (hr *HiddenRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := map[string]string{
		"type": "hidden",
		"id":   field.Name,
		"name": field.Name,
	}

	if value != nil {
		attrs["value"] = fmt.Sprintf("%v", value)
	} else if field.Default != nil {
		attrs["value"] = fmt.Sprintf("%v", field.Default)
	}

	// Hidden fields don't need container
	return fmt.Sprintf(`<input %s>`, hr.attributesToString(attrs)), nil
}

func (hr *HiddenRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldHidden
}

func (hr *HiddenRenderer) GetRequiredAssets() []string {
	return []string{} // No assets needed for hidden fields
}

// TextareaRenderer handles textarea fields
type TextareaRenderer struct {
	*BaseRenderer
}

func NewTextareaRenderer(base *BaseRenderer) *TextareaRenderer {
	return &TextareaRenderer{BaseRenderer: base}
}

func (tr *TextareaRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := tr.buildInputAttributes(field, value)
	delete(attrs, "value") // Value goes inside textarea tag

	// Add textarea-specific attributes
	if field.Config != nil {
		if rows, exists := field.Config["rows"]; exists {
			attrs["rows"] = fmt.Sprintf("%v", rows)
		} else {
			attrs["rows"] = "3"
		}

		if cols, exists := field.Config["cols"]; exists {
			attrs["cols"] = fmt.Sprintf("%v", cols)
		}

		if resize, exists := field.Config["resize"]; exists {
			attrs["style"] = fmt.Sprintf("resize: %v", resize)
		}
	} else {
		attrs["rows"] = "3"
	}

	// Get the value to put inside textarea
	var textValue string
	if value != nil {
		textValue = fmt.Sprintf("%v", value)
	} else if field.Default != nil {
		textValue = fmt.Sprintf("%v", field.Default)
	}

	inputHTML := fmt.Sprintf(`<textarea %s>%s</textarea>`, tr.attributesToString(attrs), textValue)
	return tr.RenderContainer(field, inputHTML, errors), nil
}

func (tr *TextareaRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldTextarea
}

func (tr *TextareaRenderer) GetRequiredAssets() []string {
	return []string{"css/textarea.css"}
}