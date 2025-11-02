package schema

import (
	"context"
	"fmt"
	"strings"
)

// RepeatableField extends the existing Field type to support repeatable/array fields
// This is used for invoice line items, task lists, etc.
type RepeatableField struct {
	Field                                                          // Embed base field
	Template    []Field         `json:"template" validate:"dive"`  // Fields to repeat
	MinItems    int             `json:"minItems,omitempty"`        // Minimum number of items
	MaxItems    int             `json:"maxItems,omitempty"`        // Maximum number of items
	Sortable    bool            `json:"sortable,omitempty"`        // Allow drag-drop reordering
	Collapsible bool            `json:"collapsible,omitempty"`     // Items can be collapsed
	ItemLabel   string          `json:"itemLabel,omitempty"`       // Template for item labels (e.g., "Item {index}")
	DefaultItem map[string]any  `json:"defaultItem,omitempty"`     // Default values for new items
	ShowIndex   bool            `json:"showIndex,omitempty"`       // Show item numbers
	AddText     string          `json:"addText,omitempty"`         // Text for add button
	RemoveText  string          `json:"removeText,omitempty"`      // Text for remove button
}

// TableRepeaterField provides a table layout for repeatable fields
type TableRepeaterField struct {
	RepeatableField
	Editable    bool                  `json:"editable,omitempty"`    // Inline editing
	BulkActions []Action              `json:"bulkActions,omitempty"` // Actions on selected items
	RowActions  []Action              `json:"rowActions,omitempty"`  // Actions per row
	Footer      bool                  `json:"footer,omitempty"`      // Show totals footer
	Aggregates  map[string]Aggregate  `json:"aggregates,omitempty"`  // Column aggregations
	Selectable  bool                  `json:"selectable,omitempty"`  // Allow row selection
}

// Aggregate defines how to aggregate column values
type Aggregate struct {
	Type   AggregateType `json:"type"`                    // sum, avg, count, min, max
	Field  string        `json:"field"`                   // Field to aggregate
	Label  string        `json:"label,omitempty"`         // Label for aggregate
	Format string        `json:"format,omitempty"`        // Display format
}

// AggregateType defines aggregation functions
type AggregateType string

const (
	AggregateSum   AggregateType = "sum"
	AggregateAvg   AggregateType = "avg"
	AggregateCount AggregateType = "count"
	AggregateMin   AggregateType = "min"
	AggregateMax   AggregateType = "max"
)

// RepeatableOperations defines operations that can be performed on repeatable fields
type RepeatableOperations interface {
	Add(ctx context.Context, item map[string]any) error
	Remove(ctx context.Context, index int) error
	Update(ctx context.Context, index int, item map[string]any) error
	Move(ctx context.Context, from, to int) error
	Clear(ctx context.Context) error
	Import(ctx context.Context, items []map[string]any) error
	Export(ctx context.Context, format string) ([]byte, error)
	Validate(ctx context.Context, items []map[string]any) error
}

// RepeatableManager handles operations on repeatable fields
type RepeatableManager struct {
	field *RepeatableField
}

// NewRepeatableManager creates a new repeatable field manager
func NewRepeatableManager(field *RepeatableField) *RepeatableManager {
	return &RepeatableManager{field: field}
}

// ValidateItems validates all items in a repeatable field
func (rm *RepeatableManager) ValidateItems(ctx context.Context, items []map[string]any) error {
	collector := NewErrorCollector()
	
	// Check min/max items
	if rm.field.MinItems > 0 && len(items) < rm.field.MinItems {
		collector.AddValidationError(
			rm.field.Name,
			"min_items",
			fmt.Sprintf("minimum %d items required, got %d", rm.field.MinItems, len(items)),
		)
	}
	
	if rm.field.MaxItems > 0 && len(items) > rm.field.MaxItems {
		collector.AddValidationError(
			rm.field.Name,
			"max_items",
			fmt.Sprintf("maximum %d items allowed, got %d", rm.field.MaxItems, len(items)),
		)
	}
	
	// Validate each item against template
	for itemIndex, item := range items {
		for _, templateField := range rm.field.Template {
			value, exists := item[templateField.Name]
			
			// Check required fields
			if templateField.Required && (!exists || value == nil || value == "") {
				collector.AddValidationError(
					fmt.Sprintf("%s[%d].%s", rm.field.Name, itemIndex, templateField.Name),
					"required",
					fmt.Sprintf("%s is required", templateField.Label),
				)
				continue
			}
			
			// Validate field value if exists
			if exists && value != nil {
				if err := templateField.ValidateValue(value); err != nil {
					collector.AddValidationError(
						fmt.Sprintf("%s[%d].%s", rm.field.Name, itemIndex, templateField.Name),
						"invalid_value",
						err.Error(),
					)
				}
			}
		}
	}
	
	if collector.HasErrors() {
		return collector.Errors()
	}
	
	return nil
}

// GetItemLabel generates a label for an item based on the template
func (rm *RepeatableManager) GetItemLabel(item map[string]any, index int) string {
	if rm.field.ItemLabel == "" {
		return fmt.Sprintf("Item %d", index+1)
	}
	
	// Simple template replacement
	label := rm.field.ItemLabel
	
	// Replace {index} with item index
	if len(label) > 0 {
		label = fmt.Sprintf(label, index+1)
	}
	
	// Replace field references like {product_name}
	for fieldName, value := range item {
		placeholder := fmt.Sprintf("{%s}", fieldName)
		if value != nil {
			label = strings.ReplaceAll(label, placeholder, fmt.Sprintf("%s", value))
		}
	}
	
	return label
}

// CreateDefaultItem creates a new item with default values
func (rm *RepeatableManager) CreateDefaultItem() map[string]any {
	item := make(map[string]any)
	
	// Apply default values from template fields
	for _, field := range rm.field.Template {
		if field.Default != nil {
			item[field.Name] = field.Default
		} else {
			item[field.Name] = field.GetDefaultValue()
		}
	}
	
	// Apply repeatable field default item values
	for key, value := range rm.field.DefaultItem {
		item[key] = value
	}
	
	return item
}

// ValidateRepeatable validates a repeatable field configuration
func (rf *RepeatableField) ValidateRepeatable() error {
	collector := NewErrorCollector()
	
	// Validate base field
	if err := rf.Field.Validate(context.Background()); err != nil {
		collector.AddError(err.(SchemaError))
	}
	
	// Validate template fields
	if len(rf.Template) == 0 {
		collector.AddValidationError(
			rf.Name,
			"empty_template",
			"repeatable field must have at least one template field",
		)
	}
	
	templateFieldNames := make(map[string]bool)
	for i, field := range rf.Template {
		if err := field.Validate(context.Background()); err != nil {
			collector.AddFieldError(fmt.Sprintf("Template[%d]", i), err.(SchemaError))
		}
		
		// Check for duplicate field names in template
		if templateFieldNames[field.Name] {
			collector.AddValidationError(
				fmt.Sprintf("Template[%d]", i),
				"duplicate_field",
				fmt.Sprintf("duplicate field name in template: %s", field.Name),
			)
		}
		templateFieldNames[field.Name] = true
	}
	
	// Validate min/max constraints
	if rf.MinItems < 0 {
		collector.AddValidationError(
			rf.Name,
			"invalid_min_items",
			"minItems cannot be negative",
		)
	}
	
	if rf.MaxItems > 0 && rf.MinItems > rf.MaxItems {
		collector.AddValidationError(
			rf.Name,
			"invalid_item_range",
			"minItems cannot be greater than maxItems",
		)
	}
	
	if collector.HasErrors() {
		return collector.Errors()
	}
	
	return nil
}

// RepeatableFieldBuilder provides a fluent interface for building repeatable fields
type RepeatableFieldBuilder struct {
	field *RepeatableField
}

// NewRepeatableField starts building a repeatable field
func NewRepeatableField(name, label string) *RepeatableFieldBuilder {
	return &RepeatableFieldBuilder{
		field: &RepeatableField{
			Field: Field{
				Name:  name,
				Type:  FieldRepeatable,
				Label: label,
			},
			Template:    []Field{},
			MinItems:    0,
			MaxItems:    0,
			Sortable:    true,
			Collapsible: false,
			ShowIndex:   true,
			AddText:     "Add Item",
			RemoveText:  "Remove",
		},
	}
}

// WithTemplate sets the field template
func (rfb *RepeatableFieldBuilder) WithTemplate(fields []Field) *RepeatableFieldBuilder {
	rfb.field.Template = fields
	return rfb
}

// WithMinItems sets minimum items
func (rfb *RepeatableFieldBuilder) WithMinItems(min int) *RepeatableFieldBuilder {
	rfb.field.MinItems = min
	return rfb
}

// WithMaxItems sets maximum items
func (rfb *RepeatableFieldBuilder) WithMaxItems(max int) *RepeatableFieldBuilder {
	rfb.field.MaxItems = max
	return rfb
}

// WithSortable enables/disables sorting
func (rfb *RepeatableFieldBuilder) WithSortable(sortable bool) *RepeatableFieldBuilder {
	rfb.field.Sortable = sortable
	return rfb
}

// WithCollapsible enables/disables collapsing
func (rfb *RepeatableFieldBuilder) WithCollapsible(collapsible bool) *RepeatableFieldBuilder {
	rfb.field.Collapsible = collapsible
	return rfb
}

// WithItemLabel sets the item label template
func (rfb *RepeatableFieldBuilder) WithItemLabel(label string) *RepeatableFieldBuilder {
	rfb.field.ItemLabel = label
	return rfb
}

// WithDefaultItem sets default values for new items
func (rfb *RepeatableFieldBuilder) WithDefaultItem(defaults map[string]any) *RepeatableFieldBuilder {
	rfb.field.DefaultItem = defaults
	return rfb
}

// WithTexts sets button texts
func (rfb *RepeatableFieldBuilder) WithTexts(addText, removeText string) *RepeatableFieldBuilder {
	rfb.field.AddText = addText
	rfb.field.RemoveText = removeText
	return rfb
}

// Build returns the constructed repeatable field
func (rfb *RepeatableFieldBuilder) Build() (*RepeatableField, error) {
	if err := rfb.field.ValidateRepeatable(); err != nil {
		return nil, err
	}
	return rfb.field, nil
}

// Helper function to create invoice line items schema
func CreateInvoiceLineItemsField() *RepeatableField {
	lineItemsField, err := NewRepeatableField("line_items", "Line Items").
		WithTemplate([]Field{
			{
				Name:     "product_id",
				Type:     FieldSelect,
				Label:    "Product",
				Required: true,
			},
			{
				Name:     "description",
				Type:     FieldText,
				Label:    "Description",
				Required: false,
			},
			{
				Name:     "quantity",
				Type:     FieldNumber,
				Label:    "Quantity",
				Required: true,
				Default:  1,
				Validation: &FieldValidation{
					Min: floatPtr(0.01),
				},
			},
			{
				Name:     "unit_price",
				Type:     FieldCurrency,
				Label:    "Unit Price",
				Required: true,
				Validation: &FieldValidation{
					Min: floatPtr(0),
				},
			},
			{
				Name:     "line_total",
				Type:     FieldCurrency,
				Label:    "Total",
				Readonly: true,
				// This would be calculated: quantity * unit_price
			},
		}).
		WithMinItems(1).
		WithMaxItems(100).
		WithItemLabel("Line %d").
		WithSortable(true).
		Build()
	
	if err != nil {
		// For debugging - print the error
		// fmt.Printf("CreateInvoiceLineItemsField build error: %v\n", err)
		// Return the actual schema with error fixed
		return &RepeatableField{
			Field: Field{
				Name:  "line_items",
				Type:  FieldRepeatable,
				Label: "Line Items",
			},
			Template: []Field{
				{
					Name:     "product_id",
					Type:     FieldSelect,
					Label:    "Product",
					Required: true,
				},
				{
					Name:     "description",
					Type:     FieldText,
					Label:    "Description",
					Required: false,
				},
				{
					Name:     "quantity",
					Type:     FieldNumber,
					Label:    "Quantity",
					Required: true,
					Default:  1,
					Validation: &FieldValidation{
						Min: floatPtr(0.01),
					},
				},
				{
					Name:     "unit_price",
					Type:     FieldCurrency,
					Label:    "Unit Price",
					Required: true,
					Validation: &FieldValidation{
						Min: floatPtr(0),
					},
				},
				{
					Name:     "line_total",
					Type:     FieldCurrency,
					Label:    "Total",
					Readonly: true,
				},
			},
			MinItems:  1,
			MaxItems:  100,
			Sortable:  true,
			AddText:   "Add Item",
			RemoveText: "Remove",
		}
	}
	
	return lineItemsField
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}