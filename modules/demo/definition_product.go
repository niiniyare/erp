package demo

import "awo.so/awo/def"

// ProductDefinition — demo product entity demonstrating money, number,
// multi-select, rich text, and file upload widgets.
var ProductDefinition = def.SystemDefinition{
	Name:        "product",
	Module:      "demo",
	Label:       "Product",
	LabelPlural: "Products",
	Description: "Demo product entity showcasing currency, number, multi-select, rich text, and file upload widgets.",
	Icon:        "tag",
	Fields: []def.FieldDef{
		{
			Name:        "name",
			Type:        def.FieldTypeData,
			Label:       "Product Name",
			Required:    true,
			Searchable:  true,
			MaxLen:      255,
			Placeholder: "Enter product name",
			Icon:        "tag",
		},
		{
			Name:     "sku",
			Type:     def.FieldTypeData,
			Label:    "SKU",
			Required: true,
			Unique:   true,
			MaxLen:   100,
			Icon:     "search",
		},
		{
			Name:     "price",
			Type:     def.FieldTypeCurrency,
			Label:    "Price",
			Required: true,
			Icon:     "money",
		},
		{
			Name:    "stock_qty",
			Type:    def.FieldTypeInt,
			Label:   "Stock Quantity",
			Default: func() any { return int64(0) },
			Icon:    "building",
		},
		{
			Name:    "categories",
			Type:    def.FieldTypeMultiSelect,
			Label:   "Categories",
			Options: []string{"electronics", "clothing", "food", "furniture", "software", "services", "other"},
		},
		{
			Name:  "description",
			Type:  def.FieldTypeLongText,
			Label: "Description",
		},
		{
			Name:   "image",
			Type:   def.FieldTypeData,
			Label:  "Image URL",
			MaxLen: 500,
		},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Label:   "Status",
			Options: []string{"active", "discontinued"},
			Default: func() any { return "active" },
			Icon:    "tag",
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"demo.product.create"},
		Read:   []string{"demo.product.read"},
		Write:  []string{"demo.product.update"},
		Delete: []string{"demo.product.delete"},
	},
}

func init() {
	def.Register(&ProductDefinition)
}
