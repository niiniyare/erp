package schema

import (
	"context"
	"fmt"
	"html"
	"strings"
)

// TreeSelectRenderer handles tree/hierarchical select fields
type TreeSelectRenderer struct {
	*BaseRenderer
}

func NewTreeSelectRenderer(base *BaseRenderer) *TreeSelectRenderer {
	return &TreeSelectRenderer{BaseRenderer: base}
}

func (tsr *TreeSelectRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := tsr.buildInputAttributes(field, value)
	attrs["data-component"] = "tree-select"

	// Configuration
	multiple := false
	checkable := false
	if field.Config != nil {
		if m, exists := field.Config["multiple"]; exists {
			multiple = m.(bool)
		}
		if c, exists := field.Config["checkable"]; exists {
			checkable = c.(bool)
		}
	}

	if multiple {
		attrs["data-multiple"] = "true"
	}
	if checkable {
		attrs["data-checkable"] = "true"
	}

	// Build tree structure from options
	treeData := tsr.buildTreeData(field.Options)
	attrs["data-tree"] = treeData

	inputHTML := fmt.Sprintf(`
		<div class="tree-select" %s>
			<div class="tree-select-selector">
				<span class="tree-select-placeholder">%s</span>
				<span class="tree-select-arrow">▼</span>
			</div>
			<div class="tree-select-dropdown">
				<div class="tree-select-tree"></div>
			</div>
		</div>`, tsr.attributesToString(attrs), html.EscapeString(field.Placeholder))

	return tsr.RenderContainer(field, inputHTML, errors), nil
}

func (tsr *TreeSelectRenderer) buildTreeData(options []Option) string {
	// Simplified tree data building - in production would handle hierarchical data properly
	var items []string
	for _, option := range options {
		items = append(items, fmt.Sprintf(`{"value":"%s","label":"%s","disabled":%t}`,
			option.Value, option.Label, option.Disabled))
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ","))
}

func (tsr *TreeSelectRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldTreeSelect
}

func (tsr *TreeSelectRenderer) GetRequiredAssets() []string {
	return []string{"css/tree-select.css", "js/tree-select.js"}
}

// CascaderRenderer handles cascading select fields
type CascaderRenderer struct {
	*BaseRenderer
}

func NewCascaderRenderer(base *BaseRenderer) *CascaderRenderer {
	return &CascaderRenderer{BaseRenderer: base}
}

func (csr *CascaderRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := csr.buildInputAttributes(field, value)
	attrs["data-component"] = "cascader"

	// Configuration
	expandTrigger := "click"
	if field.Config != nil {
		if trigger, exists := field.Config["expandTrigger"]; exists {
			expandTrigger = fmt.Sprintf("%v", trigger)
		}
	}

	attrs["data-expand-trigger"] = expandTrigger

	inputHTML := fmt.Sprintf(`
		<div class="cascader" %s>
			<div class="cascader-input">
				<span class="cascader-placeholder">%s</span>
				<span class="cascader-arrow">▼</span>
			</div>
			<div class="cascader-menus"></div>
		</div>`, csr.attributesToString(attrs), html.EscapeString(field.Placeholder))

	return csr.RenderContainer(field, inputHTML, errors), nil
}

func (csr *CascaderRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldCascader
}

func (csr *CascaderRenderer) GetRequiredAssets() []string {
	return []string{"css/cascader.css", "js/cascader.js"}
}

// TransferRenderer handles transfer/dual list fields
type TransferRenderer struct {
	*BaseRenderer
}

func NewTransferRenderer(base *BaseRenderer) *TransferRenderer {
	return &TransferRenderer{BaseRenderer: base}
}

func (tr *TransferRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
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
			if v != "" {
				selectedValues = strings.Split(v, ",")
			}
		}
	}

	// Build available and selected lists
	var availableItems []string
	var selectedItems []string

	for _, option := range field.Options {
		isSelected := false
		for _, selectedVal := range selectedValues {
			if option.Value == selectedVal {
				isSelected = true
				break
			}
		}

		item := fmt.Sprintf(`
			<div class="transfer-item" data-value="%s">
				<span class="transfer-item-content">%s</span>
			</div>`, html.EscapeString(option.Value), html.EscapeString(option.Label))

		if isSelected {
			selectedItems = append(selectedItems, item)
		} else {
			availableItems = append(availableItems, item)
		}
	}

	inputHTML := fmt.Sprintf(`
		<div class="transfer-container" data-field="%s">
			<div class="transfer-panel">
				<div class="transfer-header">Available</div>
				<div class="transfer-body" id="%s-available">%s</div>
			</div>
			<div class="transfer-operations">
				<button type="button" class="transfer-btn" onclick="transferItems('%s', 'toSelected')">→</button>
				<button type="button" class="transfer-btn" onclick="transferItems('%s', 'toAvailable')">←</button>
			</div>
			<div class="transfer-panel">
				<div class="transfer-header">Selected</div>
				<div class="transfer-body" id="%s-selected">%s</div>
			</div>
			<input type="hidden" name="%s" id="%s" value="%s">
		</div>`,
		field.Name,
		field.Name, strings.Join(availableItems, ""),
		field.Name,
		field.Name,
		field.Name, strings.Join(selectedItems, ""),
		field.Name, field.Name, strings.Join(selectedValues, ","))

	return tr.RenderContainer(field, inputHTML, errors), nil
}

func (tr *TransferRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldTransfer
}

func (tr *TransferRenderer) GetRequiredAssets() []string {
	return []string{"css/transfer.css", "js/transfer.js"}
}

// AutoCompleteRenderer handles autocomplete/search fields
type AutoCompleteRenderer struct {
	*BaseRenderer
}

func NewAutoCompleteRenderer(base *BaseRenderer) *AutoCompleteRenderer {
	return &AutoCompleteRenderer{BaseRenderer: base}
}

func (acr *AutoCompleteRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := acr.buildInputAttributes(field, value)
	attrs["type"] = "text"
	attrs["autocomplete"] = "off"
	attrs["role"] = "combobox"
	attrs["aria-expanded"] = "false"
	attrs["aria-autocomplete"] = "list"

	// Configuration
	minLength := 1
	if field.Config != nil {
		if min, exists := field.Config["minLength"]; exists {
			if minInt, ok := min.(int); ok {
				minLength = minInt
			}
		}
	}

	attrs["data-min-length"] = fmt.Sprintf("%d", minLength)

	// Add data source configuration
	if field.DataSource != nil {
		attrs["data-source"] = "dynamic"
		if field.DataSource.URL != "" {
			attrs["data-url"] = field.DataSource.URL
		}
		if field.DataSource.SearchField != "" {
			attrs["data-search-field"] = field.DataSource.SearchField
		}
	} else {
		// Use static options
		attrs["data-source"] = "static"
		var options []string
		for _, option := range field.Options {
			options = append(options, fmt.Sprintf(`{"value":"%s","label":"%s"}`,
				option.Value, option.Label))
		}
		attrs["data-options"] = fmt.Sprintf("[%s]", strings.Join(options, ","))
	}

	inputHTML := fmt.Sprintf(`
		<div class="autocomplete-container">
			<input %s>
			<div class="autocomplete-dropdown" role="listbox" aria-label="Suggestions"></div>
		</div>`, acr.attributesToString(attrs))

	return acr.RenderContainer(field, inputHTML, errors), nil
}

func (acr *AutoCompleteRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldAutoComplete
}

func (acr *AutoCompleteRenderer) GetRequiredAssets() []string {
	return []string{"css/autocomplete.css", "js/autocomplete.js"}
}

// TagsRenderer handles tag input fields
type TagsRenderer struct {
	*BaseRenderer
}

func NewTagsRenderer(base *BaseRenderer) *TagsRenderer {
	return &TagsRenderer{BaseRenderer: base}
}

func (tr *TagsRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	// Parse existing tags
	var tags []string
	if value != nil {
		switch v := value.(type) {
		case []string:
			tags = v
		case []any:
			for _, tag := range v {
				tags = append(tags, fmt.Sprintf("%v", tag))
			}
		case string:
			if v != "" {
				tags = strings.Split(v, ",")
			}
		}
	}

	// Build tag elements
	var tagElements []string
	for _, tag := range tags {
		tagElements = append(tagElements, fmt.Sprintf(`
			<span class="tag">
				<span class="tag-text">%s</span>
				<button type="button" class="tag-remove" onclick="removeTag('%s', '%s')" aria-label="Remove tag">×</button>
			</span>`, html.EscapeString(tag), field.Name, html.EscapeString(tag)))
	}

	// Configuration
	maxTags := 0
	allowCustom := true
	if field.Config != nil {
		if max, exists := field.Config["maxTags"]; exists {
			if maxInt, ok := max.(int); ok {
				maxTags = maxInt
			}
		}
		if custom, exists := field.Config["allowCustom"]; exists {
			allowCustom = custom.(bool)
		}
	}

	attrs := map[string]string{
		"type":        "text",
		"class":       "tags-input",
		"placeholder": field.Placeholder,
		"id":          field.Name + "-input",
	}

	if maxTags > 0 {
		attrs["data-max-tags"] = fmt.Sprintf("%d", maxTags)
	}
	if !allowCustom {
		attrs["data-allow-custom"] = "false"
	}

	// Add predefined options for suggestions
	if len(field.Options) > 0 {
		var suggestions []string
		for _, option := range field.Options {
			suggestions = append(suggestions, fmt.Sprintf(`"%s"`, option.Value))
		}
		attrs["data-suggestions"] = fmt.Sprintf("[%s]", strings.Join(suggestions, ","))
	}

	inputHTML := fmt.Sprintf(`
		<div class="tags-container" data-field="%s">
			<div class="tags-list">%s</div>
			<input %s onkeydown="handleTagInput(event, '%s')" onblur="addTagOnBlur(event, '%s')">
			<input type="hidden" name="%s" id="%s" value="%s">
		</div>`,
		field.Name,
		strings.Join(tagElements, ""),
		tr.attributesToString(attrs),
		field.Name,
		field.Name,
		field.Name,
		field.Name,
		strings.Join(tags, ","))

	return tr.RenderContainer(field, inputHTML, errors), nil
}

func (tr *TagsRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldTags
}

func (tr *TagsRenderer) GetRequiredAssets() []string {
	return []string{"css/tags.css", "js/tags.js"}
}

// ColorRenderer handles color picker fields
type ColorRenderer struct {
	*BaseRenderer
}

func NewColorRenderer(base *BaseRenderer) *ColorRenderer {
	return &ColorRenderer{BaseRenderer: base}
}

func (cr *ColorRenderer) Render(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	attrs := cr.buildInputAttributes(field, value)
	attrs["type"] = "color"

	// Set default value if none provided
	if attrs["value"] == "" {
		attrs["value"] = "#000000"
	}

	// Configuration
	showInput := true
	if field.Config != nil {
		if show, exists := field.Config["showInput"]; exists {
			showInput = show.(bool)
		}
	}

	var inputHTML string
	if showInput {
		textAttrs := map[string]string{
			"type":        "text",
			"id":          field.Name + "-text",
			"class":       "color-input-text",
			"value":       attrs["value"],
			"placeholder": "#000000",
			"pattern":     "^#[0-9A-Fa-f]{6}$",
		}

		inputHTML = fmt.Sprintf(`
			<div class="color-picker-container">
				<input %s onchange="updateColorText('%s', this.value)">
				<input %s onchange="updateColorPicker('%s', this.value)" oninput="updateColorPicker('%s', this.value)">
			</div>`, cr.attributesToString(attrs), field.Name, cr.attributesToString(textAttrs), field.Name, field.Name)
	} else {
		inputHTML = fmt.Sprintf(`<input %s>`, cr.attributesToString(attrs))
	}

	return cr.RenderContainer(field, inputHTML, errors), nil
}

func (cr *ColorRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldColor
}

func (cr *ColorRenderer) GetRequiredAssets() []string {
	return []string{"css/color-picker.css", "js/color-picker.js"}
}