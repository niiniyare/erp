package demo

import "awo.so/awo/def"

// CustomerDefinition — demo customer entity demonstrating text, select, date,
// bool, and color field types.
var CustomerDefinition = def.SystemDefinition{
	Name:        "customer",
	Module:      "demo",
	Label:       "Customer",
	LabelPlural: "Customers",
	Description: "Demo customer entity showcasing text, select, date, switch, and color widgets.",
	Icon:        "user",
	Fields: []def.FieldDef{
		{
			Name:        "name",
			Type:        def.FieldTypeData,
			Label:       "Full Name",
			Required:    true,
			Searchable:  true,
			MaxLen:      255,
			Placeholder: "Enter customer full name",
			Icon:        "user",
		},
		{
			Name:        "email",
			Type:        def.FieldTypeData,
			Label:       "Email Address",
			Required:    true,
			Searchable:  true,
			MaxLen:      255,
			Placeholder: "customer@example.com",
			Icon:        "mail",
		},
		{
			Name:        "phone",
			Type:        def.FieldTypeData,
			Label:       "Phone",
			MaxLen:      50,
			Placeholder: "+254 700 000 000",
			Icon:        "phone",
		},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Label:   "Status",
			Options: []string{"active", "inactive", "prospect"},
			Default: func() any { return "prospect" },
			Icon:    "tag",
		},
		{
			Name:  "joined_at",
			Type:  def.FieldTypeDate,
			Label: "Joined Date",
			Icon:  "calendar",
		},
		{
			Name:        "company",
			Type:        def.FieldTypeData,
			Label:       "Company",
			MaxLen:      255,
			Searchable:  true,
			Placeholder: "Company name",
			Icon:        "building",
		},
		{
			Name:  "notes",
			Type:  def.FieldTypeSmallText,
			Label: "Notes",
		},
		{
			Name:    "newsletter",
			Type:    def.FieldTypeBool,
			Label:   "Subscribe to Newsletter",
			Default: func() any { return false },
		},
		{
			Name:   "tag_color",
			Type:   def.FieldTypeData,
			Label:  "Tag Color",
			MaxLen: 20,
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"demo.customer.create"},
		Read:   []string{"demo.customer.read"},
		Write:  []string{"demo.customer.update"},
		Delete: []string{"demo.customer.delete"},
	},
}

func init() {
	def.Register(&CustomerDefinition)
}
