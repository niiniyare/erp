package sdui

import (
	"encoding/json"
	"fmt"

	"awo.so/framework/definition"
	"awo.so/framework/internal/core"
	"awo.so/framework/pkg/fieldtype"
)

// BuildPage generates a standard amis CRUD page definition for an EntityDefinition.
// If the definition has a custom PageBuilder set, that is called instead.
//
// The returned JSON is suitable for caching in Redis and serving directly to amis clients.
func BuildPage(def *definition.EntityDefinition) (json.RawMessage, error) {
	if def.PageBuilder != nil {
		return def.PageBuilder(def)
	}
	return buildDefaultCRUDPage(def)
}

// buildDefaultCRUDPage generates a standard amis CRUD block:
// - Filter bar (filterable fields)
// - Table (all non-sensitive, non-hidden fields)
// - Create / Edit modal form
// - Row actions: edit, delete (+ submit/cancel if submittable)
func buildDefaultCRUDPage(def *definition.EntityDefinition) (json.RawMessage, error) {
	columns := make([]Component, 0, len(def.Fields))
	filterItems := make([]Component, 0)
	formItems := make([]Component, 0, len(def.Fields))

	for _, f := range def.Fields {
		if f.Hidden {
			continue
		}
		if !f.Sensitive {
			columns = append(columns, tableColumn(f))
		}
		if isFilterable(f) {
			filterItems = append(filterItems, filterItem(f))
		}
		if !f.ReadOnly {
			formItems = append(formItems, formField(f))
		}
	}

	// Row actions
	rowActions := []Component{
		{"type": "button", "label": "Edit", "actionType": "dialog", "dialog": editDialog(def.Name, formItems)},
		{
			"type": "button", "label": "Delete", "level": "danger",
			"actionType":  "ajax",
			"confirmText": fmt.Sprintf("Delete this %s?", def.Label),
			"api": Component{
				"method": "delete",
				"url":    fmt.Sprintf("/api/v1/%s/${id}", def.Name),
			},
		},
	}
	if def.IsSubmittable {
		rowActions = append(
			rowActions,
			Component{
				"type": "button", "label": "Submit", "level": "success",
				"actionType": "ajax",
				"api":        Component{"method": "post", "url": fmt.Sprintf("/api/v1/%s/${id}/submit", def.Name)},
			},
			Component{
				"type": "button", "label": "Cancel",
				"actionType": "ajax",
				"api":        Component{"method": "post", "url": fmt.Sprintf("/api/v1/%s/${id}/cancel", def.Name)},
			},
		)
	}

	columns = append(columns, Component{
		"type":    "operation",
		"label":   "Actions",
		"buttons": rowActions,
	})

	// Toolbar: create button
	toolbar := []Component{
		{
			"type":       "button",
			"label":      "New " + def.Label,
			"level":      "primary",
			"actionType": "dialog",
			"dialog":     createDialog(def.Name, formItems),
		},
	}

	// CRUD block
	crud := Component{
		"type":         "crud",
		"api":          fmt.Sprintf("/api/v1/%s", def.Name),
		"syncLocation": false,
		"columns":      columns,
		"toolbar":      toolbar,
		"filter": Component{
			"title":      "Filter",
			"body":       filterItems,
			"submitText": "Search",
		},
	}

	page := Component{
		"type":  "page",
		"title": def.Label,
		"body":  []Component{crud},
	}

	b, err := json.Marshal(page)
	if err != nil {
		return nil, fmt.Errorf("sdui: marshal page for %s: %w", def.Name, err)
	}
	return b, nil
}

func createDialog(entityName string, formItems []Component) Component {
	return Component{
		"title": "Create",
		"body": Component{
			"type": "form",
			"api":  fmt.Sprintf("/api/v1/%s", entityName),
			"body": formItems,
		},
	}
}

func editDialog(entityName string, formItems []Component) Component {
	return Component{
		"title": "Edit",
		"body": Component{
			"type":    "form",
			"method":  "put",
			"api":     fmt.Sprintf("/api/v1/%s/${id}", entityName),
			"initApi": fmt.Sprintf("/api/v1/%s/${id}", entityName),
			"body":    formItems,
		},
	}
}

// isFilterable returns true for field types that make sense in a filter bar.
func isFilterable(f core.Field) bool {
	switch f.Type {
	case fieldtype.Select, fieldtype.MultiSelect, fieldtype.Bool, fieldtype.Date,
		fieldtype.DateTime, fieldtype.Data, fieldtype.Link:
		return true
	}
	return false
}
