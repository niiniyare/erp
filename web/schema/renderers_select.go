package schema

import (
	"context"
	"fmt"
	"html"
	"strings"
)

// SelectRenderer handles select dropdown fields
type SelectRenderer struct {
	*BaseRenderer
}

func NewSelectRenderer(base *BaseRenderer) *SelectRenderer {
	return &SelectRenderer{BaseRenderer: base}
}

func (sr *SelectRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := sr.buildInputAttributes(field, value)
	delete(attrs, "value") // Value is handled in options
	delete(attrs, "placeholder") // Placeholder is handled as first option

	// Build options
	var options []string

	// Add placeholder as first option if specified
	if field.Placeholder != "" {
		options = append(options, fmt.Sprintf(`<option value="" disabled selected>%s</option>`, html.EscapeString(field.Placeholder)))
	}

	// Add options from field
	selectedValue := ""
	if value != nil {
		selectedValue = fmt.Sprintf("%v", value)
	}

	for _, option := range field.Options {
		selected := ""
		if option.Value == selectedValue || (selectedValue == "" && option.Selected) {
			selected = " selected"
		}

		disabled := ""
		if option.Disabled {
			disabled = " disabled"
		}

		options = append(options, fmt.Sprintf(`<option value="%s"%s%s>%s</option>`,
			html.EscapeString(option.Value),
			selected,
			disabled,
			html.EscapeString(option.Label)))
	}

	// Handle dynamic data source
	if field.DataSource != nil && len(field.Options) == 0 {
		attrs["data-source"] = "dynamic"
		if field.DataSource.URL != "" {
			attrs["data-url"] = field.DataSource.URL
		}
		if field.DataSource.SearchField != "" {
			attrs["data-search-field"] = field.DataSource.SearchField
		}
		if field.DataSource.ValueField != "" {
			attrs["data-value-field"] = field.DataSource.ValueField
		}
		if field.DataSource.LabelField != "" {
			attrs["data-label-field"] = field.DataSource.LabelField
		}
	}

	inputHTML := fmt.Sprintf(`<select %s>%s</select>`, sr.attributesToString(attrs), strings.Join(options, ""))
	return sr.RenderContainer(field, inputHTML, errors), nil
}

func (sr *SelectRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldSelect
}

func (sr *SelectRenderer) GetRequiredAssets() []string {
	return []string{"css/select.css"}
}

// MultiSelectRenderer handles multi-select fields
type MultiSelectRenderer struct {
	*BaseRenderer
}

func NewMultiSelectRenderer(base *BaseRenderer) *MultiSelectRenderer {
	return &MultiSelectRenderer{BaseRenderer: base}
}

func (msr *MultiSelectRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := msr.buildInputAttributes(field, value)
	attrs["multiple"] = "multiple"
	delete(attrs, "value") // Value is handled in options

	// Parse selected values
	var selectedValues []string
	if value != nil {
		switch v := value.(type) {
		case []string:
			selectedValues = v
		case []any:
			for _, val := range v {
				selectedValues = append(selectedValues, fmt.Sprintf("%v", val))
			}
		case string:
			// Handle comma-separated values
			if v != "" {
				selectedValues = strings.Split(v, ",")
			}
		}
	}

	// Build options
	var options []string
	for _, option := range field.Options {
		selected := ""
		for _, selectedVal := range selectedValues {
			if option.Value == selectedVal {
				selected = " selected"
				break
			}
		}

		disabled := ""
		if option.Disabled {
			disabled = " disabled"
		}

		options = append(options, fmt.Sprintf(`<option value="%s"%s%s>%s</option>`,
			html.EscapeString(option.Value),
			selected,
			disabled,
			html.EscapeString(option.Label)))
	}

	inputHTML := fmt.Sprintf(`<select %s>%s</select>`, msr.attributesToString(attrs), strings.Join(options, ""))
	return msr.RenderContainer(field, inputHTML, errors), nil
}

func (msr *MultiSelectRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldMultiSelect
}

func (msr *MultiSelectRenderer) GetRequiredAssets() []string {
	return []string{"css/select.css", "css/multi-select.css"}
}

// RadioRenderer handles radio button groups
type RadioRenderer struct {
	*BaseRenderer
}

func NewRadioRenderer(base *BaseRenderer) *RadioRenderer {
	return &RadioRenderer{BaseRenderer: base}
}

func (rr *RadioRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	selectedValue := ""
	if value != nil {
		selectedValue = fmt.Sprintf("%v", value)
	}

	// Get layout configuration
	layout := "vertical"
	if field.Config != nil {
		if l, exists := field.Config["layout"]; exists {
			layout = fmt.Sprintf("%v", l)
		}
	}

	var radioButtons []string
	for i, option := range field.Options {
		radioID := fmt.Sprintf("%s_%d", field.Name, i)
		checked := ""
		if option.Value == selectedValue || (selectedValue == "" && option.Selected) {
			checked = " checked"
		}

		disabled := ""
		if option.Disabled || field.Disabled {
			disabled = " disabled"
		}

		radioButtons = append(radioButtons, fmt.Sprintf(`
			<div class="radio-option">
				<input type="radio" id="%s" name="%s" value="%s"%s%s>
				<label for="%s">%s</label>
			</div>`,
			radioID,
			field.Name,
			html.EscapeString(option.Value),
			checked,
			disabled,
			radioID,
			html.EscapeString(option.Label)))
	}

	inputHTML := fmt.Sprintf(`<div class="radio-group radio-group--%s" role="radiogroup" aria-labelledby="%s-label">%s</div>`,
		layout,
		field.Name,
		strings.Join(radioButtons, ""))

	return rr.RenderContainer(field, inputHTML, errors), nil
}

func (rr *RadioRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldRadio
}

func (rr *RadioRenderer) GetRequiredAssets() []string {
	return []string{"css/radio.css"}
}

// CheckboxRenderer handles checkbox groups
type CheckboxRenderer struct {
	*BaseRenderer
}

func NewCheckboxRenderer(base *BaseRenderer) *CheckboxRenderer {
	return &CheckboxRenderer{BaseRenderer: base}
}

func (cr *CheckboxRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	// Parse selected values
	var selectedValues []string
	if value != nil {
		switch v := value.(type) {
		case []string:
			selectedValues = v
		case []any:
			for _, val := range v {
				selectedValues = append(selectedValues, fmt.Sprintf("%v", val))
			}
		case string:
			// Handle comma-separated values or single boolean
			if v == "true" || v == "1" {
				selectedValues = append(selectedValues, "true")
			} else if v != "" && v != "false" && v != "0" {
				selectedValues = strings.Split(v, ",")
			}
		case bool:
			if v {
				selectedValues = append(selectedValues, "true")
			}
		}
	}

	// Get layout configuration
	layout := "vertical"
	if field.Config != nil {
		if l, exists := field.Config["layout"]; exists {
			layout = fmt.Sprintf("%v", l)
		}
	}

	// Handle single checkbox (boolean)
	if len(field.Options) == 0 {
		checkboxID := field.Name
		checked := ""
		if len(selectedValues) > 0 {
			checked = " checked"
		}

		disabled := ""
		if field.Disabled {
			disabled = " disabled"
		}

		inputHTML := fmt.Sprintf(`
			<div class="checkbox-single">
				<input type="checkbox" id="%s" name="%s" value="true"%s%s>
				<label for="%s">%s</label>
			</div>`,
			checkboxID,
			field.Name,
			checked,
			disabled,
			checkboxID,
			html.EscapeString(field.Label))

		return cr.RenderContainer(field, inputHTML, errors), nil
	}

	// Handle checkbox group
	var checkboxes []string
	for i, option := range field.Options {
		checkboxID := fmt.Sprintf("%s_%d", field.Name, i)
		checked := ""
		for _, selectedVal := range selectedValues {
			if option.Value == selectedVal {
				checked = " checked"
				break
			}
		}

		disabled := ""
		if option.Disabled || field.Disabled {
			disabled = " disabled"
		}

		checkboxes = append(checkboxes, fmt.Sprintf(`
			<div class="checkbox-option">
				<input type="checkbox" id="%s" name="%s[]" value="%s"%s%s>
				<label for="%s">%s</label>
			</div>`,
			checkboxID,
			field.Name,
			html.EscapeString(option.Value),
			checked,
			disabled,
			checkboxID,
			html.EscapeString(option.Label)))
	}

	inputHTML := fmt.Sprintf(`<div class="checkbox-group checkbox-group--%s" role="group" aria-labelledby="%s-label">%s</div>`,
		layout,
		field.Name,
		strings.Join(checkboxes, ""))

	return cr.RenderContainer(field, inputHTML, errors), nil
}

func (cr *CheckboxRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldCheckbox
}

func (cr *CheckboxRenderer) GetRequiredAssets() []string {
	return []string{"css/checkbox.css"}
}

// SwitchRenderer handles toggle switch fields
type SwitchRenderer struct {
	*BaseRenderer
}

func NewSwitchRenderer(base *BaseRenderer) *SwitchRenderer {
	return &SwitchRenderer{BaseRenderer: base}
}

func (sr *SwitchRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	checked := false
	if value != nil {
		switch v := value.(type) {
		case bool:
			checked = v
		case string:
			checked = v == "true" || v == "1" || v == "on"
		case int:
			checked = v == 1
		}
	}

	// Get switch configuration
	size := "medium"
	onText := "On"
	offText := "Off"

	if field.Config != nil {
		if s, exists := field.Config["size"]; exists {
			size = fmt.Sprintf("%v", s)
		}
		if on, exists := field.Config["onText"]; exists {
			onText = fmt.Sprintf("%v", on)
		}
		if off, exists := field.Config["offText"]; exists {
			offText = fmt.Sprintf("%v", off)
		}
	}

	checkedAttr := ""
	if checked {
		checkedAttr = " checked"
	}

	disabled := ""
	if field.Disabled {
		disabled = " disabled"
	}

	inputHTML := fmt.Sprintf(`
		<div class="switch-container switch--%s">
			<input type="checkbox" id="%s" name="%s" value="true" class="switch-input"%s%s>
			<label for="%s" class="switch-label">
				<span class="switch-track">
					<span class="switch-thumb"></span>
					<span class="switch-text switch-text--on">%s</span>
					<span class="switch-text switch-text--off">%s</span>
				</span>
			</label>
		</div>`,
		size,
		field.Name,
		field.Name,
		checkedAttr,
		disabled,
		field.Name,
		html.EscapeString(onText),
		html.EscapeString(offText))

	return sr.RenderContainer(field, inputHTML, errors), nil
}

func (sr *SwitchRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldSwitch
}

func (sr *SwitchRenderer) GetRequiredAssets() []string {
	return []string{"css/switch.css"}
}