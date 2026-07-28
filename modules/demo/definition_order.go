package demo

import "awo.so/awo/def"

// OrderDefinition — demo order entity demonstrating lookup (FK), date,
// datetime, select, money, switch, and textarea widgets.
var OrderDefinition = def.SystemDefinition{
	Name:        "order",
	Module:      "demo",
	Label:       "Order",
	LabelPlural: "Orders",
	Description: "Demo order entity showcasing FK lookup, datetime, select, money, and switch widgets.",
	Icon:        "document",
	Fields: []def.FieldDef{
		{
			Name:       "customer_id",
			Type:       def.FieldTypeLink,
			Label:      "Customer",
			LinkTarget: "demo_customer",
			Required:   true,
			Icon:       "user",
		},
		{
			Name:       "product_id",
			Type:       def.FieldTypeLink,
			Label:      "Product",
			LinkTarget: "demo_product",
			Icon:       "tag",
		},
		{
			Name:     "order_date",
			Type:     def.FieldTypeDate,
			Label:    "Order Date",
			Required: true,
			Icon:     "calendar",
		},
		{
			Name:  "delivery_at",
			Type:  def.FieldTypeDateTime,
			Label: "Delivery Date & Time",
			Icon:  "calendar",
		},
		{
			Name:       "status",
			Type:       def.FieldTypeSelect,
			Label:      "Status",
			Options:    []string{"draft", "confirmed", "shipped", "delivered", "cancelled"},
			Default:    func() any { return "draft" },
			Searchable: true,
			Icon:       "tag",
		},
		{
			Name:  "total_amount",
			Type:  def.FieldTypeCurrency,
			Label: "Total Amount",
			Icon:  "money",
		},
		{
			Name:  "notes",
			Type:  def.FieldTypeSmallText,
			Label: "Notes",
		},
		{
			Name:    "urgent",
			Type:    def.FieldTypeBool,
			Label:   "Urgent",
			Default: func() any { return false },
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"demo.order.create"},
		Read:   []string{"demo.order.read"},
		Write:  []string{"demo.order.update"},
		Delete: []string{"demo.order.delete"},
	},
}

func init() {
	def.Register(&OrderDefinition)
}
