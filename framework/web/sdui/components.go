// Package sdui generates amis JSON page definitions from EntityDefinitions.
package sdui

import (
	"awo.so/framework/internal/core"
	"awo.so/framework/pkg/fieldtype"
)

// Component is a JSON-serialisable amis component node.
type Component = map[string]any

// formField converts an internal Field to an amis form item component.
func formField(f core.Field) Component {
	c := Component{
		"name":  f.Name,
		"label": labelOf(f),
	}

	switch f.Type {
	case fieldtype.Data, fieldtype.SmallText:
		c["type"] = "input-text"
		if f.MaxLen > 0 {
			c["maxLength"] = f.MaxLen
		}
	case fieldtype.LongText:
		c["type"] = "textarea"
	case fieldtype.Int:
		c["type"] = "input-number"
		c["precision"] = 0
	case fieldtype.Float:
		c["type"] = "input-number"
	case fieldtype.Currency:
		c["type"] = "input-number"
		c["precision"] = 4
		c["prefix"] = "KES"
	case fieldtype.Bool:
		c["type"] = "switch"
	case fieldtype.Date:
		c["type"] = "input-date"
		c["format"] = "YYYY-MM-DD"
	case fieldtype.DateTime:
		c["type"] = "input-datetime"
		c["format"] = "YYYY-MM-DD HH:mm:ss"
		c["utc"] = true
	case fieldtype.Time:
		c["type"] = "input-time"
	case fieldtype.UUID:
		c["type"] = "input-text"
		c["disabled"] = true
	case fieldtype.Select:
		c["type"] = "select"
		c["options"] = choiceOptions(f.Choices)
	case fieldtype.MultiSelect:
		c["type"] = "select"
		c["multiple"] = true
		c["options"] = choiceOptions(f.Choices)
	case fieldtype.JSON:
		c["type"] = "json-editor"
	case fieldtype.Link:
		c["type"] = "select"
		c["source"] = "/api/v1/" + f.LinkTarget + "?limit=50"
		c["labelField"] = "name"
		c["valueField"] = "id"
		c["searchable"] = true
	case fieldtype.Attach:
		c["type"] = "input-file"
	case fieldtype.AttachImage:
		c["type"] = "input-image"
	default:
		c["type"] = "input-text"
	}

	if f.Required {
		c["required"] = true
	}
	if f.ReadOnly || f.Immutable {
		c["disabled"] = true
	}
	if f.Placeholder != "" {
		c["placeholder"] = f.Placeholder
	}
	if f.Description != "" {
		c["description"] = f.Description
	}

	return c
}

// tableColumn converts an internal Field to an amis table column descriptor.
func tableColumn(f core.Field) Component {
	col := Component{
		"name":  f.Name,
		"label": labelOf(f),
	}

	switch f.Type {
	case fieldtype.Currency:
		col["type"] = "number"
		col["prefix"] = "KES "
	case fieldtype.Bool:
		col["type"] = "status"
	case fieldtype.Date:
		col["type"] = "date"
		col["format"] = "YYYY-MM-DD"
	case fieldtype.DateTime:
		col["type"] = "datetime"
		col["format"] = "YYYY-MM-DD HH:mm"
	default:
		col["type"] = "text"
	}

	if f.Sensitive {
		col["type"] = "text"
		col["value"] = "***"
	}

	return col
}

// filterItem converts a field to an amis filter bar item.
func filterItem(f core.Field) Component {
	item := Component{"name": f.Name, "label": labelOf(f)}
	switch f.Type {
	case fieldtype.Select, fieldtype.MultiSelect:
		item["type"] = "select"
		item["options"] = choiceOptions(f.Choices)
		item["clearable"] = true
	case fieldtype.Date:
		item["type"] = "input-date-range"
	case fieldtype.Bool:
		item["type"] = "select"
		item["options"] = []Component{
			{"label": "Yes", "value": true},
			{"label": "No", "value": false},
		}
	default:
		item["type"] = "input-text"
	}
	return item
}

func choiceOptions(choices []string) []Component {
	opts := make([]Component, len(choices))
	for i, c := range choices {
		opts[i] = Component{"label": c, "value": c}
	}
	return opts
}

func labelOf(f core.Field) string {
	if f.Label != "" && f.Label != f.Name {
		return f.Label
	}
	return f.Name
}
