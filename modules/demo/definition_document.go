package demo

import "awo.so/awo/def"

// DocumentDefinition — demo document entity demonstrating rich text, file
// upload, datetime, switch, and select (badge) widgets.
var DocumentDefinition = def.SystemDefinition{
	Name:        "document",
	Module:      "demo",
	Label:       "Document",
	LabelPlural: "Documents",
	Description: "Demo document entity showcasing rich text, file upload, datetime, switch, and badge widgets.",
	Icon:        "document",
	Fields: []def.FieldDef{
		{
			Name:        "title",
			Type:        def.FieldTypeData,
			Label:       "Title",
			Required:    true,
			Searchable:  true,
			MaxLen:      255,
			Placeholder: "Document title",
			Icon:        "document",
		},
		{
			Name:    "doc_type",
			Type:    def.FieldTypeSelect,
			Label:   "Type",
			Options: []string{"contract", "proposal", "report", "invoice"},
			Default: func() any { return "proposal" },
			Icon:    "tag",
		},
		{
			Name:  "body",
			Type:  def.FieldTypeLongText,
			Label: "Body",
		},
		{
			Name:   "attachment",
			Type:   def.FieldTypeData,
			Label:  "Attachment URL",
			MaxLen: 500,
		},
		{
			Name:  "created_at",
			Type:  def.FieldTypeDateTime,
			Label: "Created At",
			Icon:  "calendar",
		},
		{
			Name:        "author",
			Type:        def.FieldTypeData,
			Label:       "Author",
			MaxLen:      255,
			Searchable:  true,
			Placeholder: "Author name",
			Icon:        "user",
		},
		{
			Name:    "published",
			Type:    def.FieldTypeBool,
			Label:   "Published",
			Default: func() any { return false },
		},
		{
			Name:    "review_status",
			Type:    def.FieldTypeSelect,
			Label:   "Review Status",
			Options: []string{"draft", "review", "approved"},
			Default: func() any { return "draft" },
			Icon:    "tag",
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"demo.document.create"},
		Read:   []string{"demo.document.read"},
		Write:  []string{"demo.document.update"},
		Delete: []string{"demo.document.delete"},
	},
}

func init() {
	def.Register(&DocumentDefinition)
}
