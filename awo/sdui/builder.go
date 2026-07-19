package sdui

import (
	"encoding/json"
	"fmt"
	"strings"

	"awo.so/awo/compiler"
	"awo.so/awo/def"
)

// Builder generates amis page schemas from compiled EntitySchemas.
type Builder struct {
	apiBase string // e.g. "/api/v1/entities"
}

// NewBuilder returns a Builder that generates API paths under apiBase.
func NewBuilder(apiBase string) *Builder {
	return &Builder{apiBase: strings.TrimRight(apiBase, "/")}
}

func (b *Builder) entityURL(entityName string) string {
	return b.apiBase + "/" + entityName
}

// ListPage generates a CRUD list page schema for the entity.
func (b *Builder) ListPage(es *compiler.EntitySchema) ([]byte, error) {
	name := es.LocalName
	label := es.Label
	url := b.entityURL(name)

	var columns []Column
	for _, f := range es.Fields {
		if f.Sensitive || f.Hidden {
			continue
		}
		columns = append(columns, Column{
			Name:  f.Name,
			Label: fieldLabel(f),
			Type:  columnType(f.Type),
		})
	}
	// Always include id and timestamps.
	columns = append([]Column{{Name: "id", Label: "ID", Type: "text"}}, columns...)
	columns = append(columns, Column{Name: "created_at", Label: "Created", Type: "datetime"})

	crud := CRUD{
		Type:         "crud",
		API:          url,
		Columns:      columns,
		Pagination:   true,
		PageSize:     20,
		SyncLocation: true,
	}

	page := Page{
		Type:  "page",
		Title: label,
		Body:  crud,
	}
	return json.MarshalIndent(page, "", "  ")
}

// CreatePage generates a form page for record creation.
func (b *Builder) CreatePage(es *compiler.EntitySchema) ([]byte, error) {
	name := es.LocalName
	label := es.Label
	url := b.entityURL(name)

	form := Form{
		Type:       "form",
		API:        "POST " + url,
		Body:       b.formControls(es, false),
		SubmitText: "Create " + label,
	}
	page := Page{
		Type:  "page",
		Title: "New " + label,
		Body:  form,
	}
	return json.MarshalIndent(page, "", "  ")
}

// EditPage generates a form page for record editing.
func (b *Builder) EditPage(es *compiler.EntitySchema) ([]byte, error) {
	name := es.LocalName
	label := es.Label
	url := b.entityURL(name) + "/${id}"

	form := Form{
		Type:       "form",
		API:        "PATCH " + url,
		Body:       b.formControls(es, true),
		SubmitText: "Save " + label,
	}
	page := Page{
		Type:  "page",
		Title: "Edit " + label,
		Body:  form,
	}
	return json.MarshalIndent(page, "", "  ")
}

// DetailPage generates a static detail view page.
func (b *Builder) DetailPage(es *compiler.EntitySchema) ([]byte, error) {
	name := es.LocalName
	label := es.Label
	url := b.entityURL(name) + "/${id}"

	// Detail view uses a form in read-only mode.
	var controls []FormControl
	for _, f := range es.Fields {
		if f.Sensitive || f.Hidden {
			continue
		}
		ctrl := b.fieldToControl(f, true)
		ctrl.ReadOnly = true
		ctrl.Disabled = true
		controls = append(controls, ctrl)
	}

	form := Form{
		Type: "form",
		API:  "GET " + url,
		Body: controls,
	}
	page := Page{
		Type:  "page",
		Title: label + " Detail",
		Body:  form,
	}
	return json.MarshalIndent(page, "", "  ")
}

// formControls builds the form body for create (isEdit=false) or edit (isEdit=true).
func (b *Builder) formControls(es *compiler.EntitySchema, isEdit bool) []FormControl {
	var controls []FormControl
	for _, f := range es.Fields {
		if f.Sensitive || f.Hidden {
			continue
		}
		// NamingSeries: auto-assigned — show read-only.
		if f.Type == def.FieldTypeNamingSeries {
			ctrl := FormControl{
				Type:     "input-text",
				Name:     f.Name,
				Label:    fieldLabel(f),
				Disabled: true,
			}
			controls = append(controls, ctrl)
			continue
		}
		ctrl := b.fieldToControl(f, isEdit)
		controls = append(controls, ctrl)
	}
	return controls
}

func (b *Builder) fieldToControl(f def.FieldDef, isEdit bool) FormControl {
	ctrl := FormControl{
		Name:        f.Name,
		Label:       fieldLabel(f),
		Required:    f.Required,
		Description: f.Description,
	}

	if isEdit && f.Immutable {
		ctrl.Disabled = true
	}

	switch f.Type {
	case def.FieldTypeData, def.FieldTypeSmallText:
		ctrl.Type = "input-text"
		if f.MaxLen > 0 {
			ctrl.Placeholder = fmt.Sprintf("Max %d characters", f.MaxLen)
		}

	case def.FieldTypeLongText:
		ctrl.Type = "textarea"
		ctrl.MinRows = 3
		ctrl.MaxRows = 10

	case def.FieldTypeInt:
		ctrl.Type = "input-number"
		ctrl.Step = 1
		if f.Min != nil {
			ctrl.Min = int64(*f.Min)
		}
		if f.Max != nil {
			ctrl.Max = int64(*f.Max)
		}

	case def.FieldTypeFloat:
		ctrl.Type = "input-number"
		if f.Min != nil {
			ctrl.Min = *f.Min
		}
		if f.Max != nil {
			ctrl.Max = *f.Max
		}

	case def.FieldTypeCurrency:
		ctrl.Type = "input-number"
		ctrl.Precision = 4
		if f.Min != nil {
			ctrl.Min = *f.Min
		}
		if f.Max != nil {
			ctrl.Max = *f.Max
		}

	case def.FieldTypeBool:
		ctrl.Type = "switch"

	case def.FieldTypeDate:
		ctrl.Type = "input-date"

	case def.FieldTypeDateTime:
		ctrl.Type = "input-datetime"

	case def.FieldTypeTime:
		ctrl.Type = "input-time"

	case def.FieldTypeSelect:
		ctrl.Type = "select"
		for _, opt := range f.Options {
			ctrl.Options = append(ctrl.Options, Option{Label: opt, Value: opt})
		}

	case def.FieldTypeMultiSelect:
		ctrl.Type = "select"
		ctrl.Multiple = true
		for _, opt := range f.Options {
			ctrl.Options = append(ctrl.Options, Option{Label: opt, Value: opt})
		}

	case def.FieldTypeLink:
		ctrl.Type = "select"
		if f.LinkTarget != "" {
			ctrl.Source = b.apiBase + "/" + f.LinkTarget + "?_select=id,name"
		}

	case def.FieldTypeLinkList:
		ctrl.Type = "select"
		ctrl.Multiple = true
		if f.LinkTarget != "" {
			ctrl.Source = b.apiBase + "/" + f.LinkTarget + "?_select=id,name"
		}

	case def.FieldTypeJSON:
		ctrl.Type = "json-editor"

	case def.FieldTypeNamingSeries:
		ctrl.Type = "input-text"
		ctrl.Disabled = true

	default:
		ctrl.Type = "input-text"
	}

	return ctrl
}

// fieldLabel returns the display label for a field: FieldDef.Label if set,
// otherwise a title-cased version of the field name.
func fieldLabel(f def.FieldDef) string {
	if f.Label != "" {
		return f.Label
	}
	return toTitle(f.Name)
}

// toTitle converts snake_case to Title Case.
func toTitle(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// columnType maps a FieldType to an amis column type string.
func columnType(ft def.FieldType) string {
	switch ft {
	case def.FieldTypeBool:
		return "switch"
	case def.FieldTypeDate:
		return "date"
	case def.FieldTypeDateTime:
		return "datetime"
	case def.FieldTypeInt, def.FieldTypeFloat, def.FieldTypeCurrency:
		return "number"
	default:
		return "text"
	}
}
