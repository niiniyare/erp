package amis

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

// CRUDBuilder builds an AMIS `crud` component — the primary list/table surface.
type CRUDBuilder struct{ base }

// CRUD creates a new CRUD builder. api is the list endpoint, e.g.
// "get:/api/v1/finance/invoices".
func CRUD(api string) *CRUDBuilder {
	b := &CRUDBuilder{base{s: M{
		"type":         "crud",
		"api":          api,
		"syncLocation": true, // Decision: filter state must persist in URL
	}}}
	return b
}

// Columns sets the table columns.
func (b *CRUDBuilder) Columns(cols ...*ColumnBuilder) *CRUDBuilder {
	arr := make(A, len(cols))
	for i, c := range cols {
		arr[i] = c.Build()
	}
	b.s["columns"] = arr
	return b
}

// Filter sets the filter form schema. Pass a FormBuilder or raw M.
func (b *CRUDBuilder) Filter(f any) *CRUDBuilder {
	switch v := f.(type) {
	case *FormBuilder:
		b.set("filter", v.Build())
	default:
		b.set("filter", v)
	}
	return b
}

// Toolbar sets top toolbar items (e.g. a create button).
func (b *CRUDBuilder) Toolbar(items ...any) *CRUDBuilder {
	b.set("headerToolbar", items)
	return b
}

// BulkActions sets actions available when rows are selected.
func (b *CRUDBuilder) BulkActions(actions ...any) *CRUDBuilder {
	b.set("bulkActions", actions)
	return b
}

// EmptyText sets the message shown when no records exist.
func (b *CRUDBuilder) EmptyText(msg string) *CRUDBuilder {
	b.set("placeholder", msg)
	return b
}

// DefaultSort sets the default sort column and direction.
func (b *CRUDBuilder) DefaultSort(field, order string) *CRUDBuilder {
	b.set("orderBy", field)
	b.set("orderDir", order)
	return b
}

// PerPage sets the default page size.
func (b *CRUDBuilder) PerPage(n int) *CRUDBuilder { b.set("perPage", n); return b }

// StaticData replaces the API call with inline mock rows.
// The crud will render immediately with no HTTP requests — useful for demos and tests.
func (b *CRUDBuilder) StaticData(items A) *CRUDBuilder {
	delete(b.s, "api")
	b.s["source"] = "${items}"
	b.s["data"] = M{"items": items, "count": len(items)}
	return b
}

// MarshalJSON implements json.Marshaler.
func (b *CRUDBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *CRUDBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Column
// ---------------------------------------------------------------------------

// ColumnBuilder builds a single table column.
type ColumnBuilder struct{ base }

// Column creates a new column. name is the data field key; label is the header.
func Column(name, label string) *ColumnBuilder {
	b := &ColumnBuilder{base{s: M{
		"name":  name,
		"label": label,
	}}}
	return b
}

// Type sets the column renderer type: "text" | "date" | "datetime" | "number" |
// "currency" | "status" | "mapping" | "tag" | "image" | "link" | "tpl".
// Defaults to "text" if not set.
func (b *ColumnBuilder) Type(t string) *ColumnBuilder { b.set("type", t); return b }

// Width sets the column width in pixels.
func (b *ColumnBuilder) Width(px int) *ColumnBuilder { b.set("width", px); return b }

// Align sets text alignment: "left" | "center" | "right".
// Convention: numbers right, text left, status center.
func (b *ColumnBuilder) Align(a string) *ColumnBuilder { b.set("align", a); return b }

// Sortable enables column sorting.
func (b *ColumnBuilder) Sortable() *ColumnBuilder { b.set("sortable", true); return b }

// Fixed pins the column: "left" | "right".
func (b *ColumnBuilder) Fixed(side string) *ColumnBuilder { b.set("fixed", side); return b }

// Tpl sets a template for the cell value. Use ${name} expressions.
func (b *ColumnBuilder) Tpl(tpl string) *ColumnBuilder {
	b.set("type", "tpl")
	b.set("tpl", tpl)
	return b
}

// Map sets a value mapping (e.g. status codes to labels).
func (b *ColumnBuilder) Map(mapping M) *ColumnBuilder {
	b.set("type", "mapping")
	b.set("map", mapping)
	return b
}

// Buttons turns this column into an operation column with action buttons.
func (b *ColumnBuilder) Buttons(btns ...M) *ColumnBuilder {
	b.set("type", "operation")
	b.set("label", "Actions")
	b.set("buttons", btns)
	return b
}

// VisibleOn sets a conditional visibility expression.
func (b *ColumnBuilder) VisibleOn(expr string) *ColumnBuilder {
	b.set("visibleOn", expr)
	return b
}

// MarshalJSON implements json.Marshaler.
func (b *ColumnBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *ColumnBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Action buttons (shorthand helpers for common row actions)
// ---------------------------------------------------------------------------

// EditBtn returns a standard row edit button that opens a dialog.
// api is the PUT endpoint; fields are the form fields to show in the dialog.
func EditBtn(api string, fields ...M) M {
	return M{
		"type":       "button",
		"label":      "Edit",
		"icon":       "fa fa-pencil",
		"actionType": "dialog",
		"dialog": M{
			"title": "Edit",
			"body": M{
				"type": "form",
				"api":  api,
				"body": fields,
			},
		},
	}
}

// ViewBtn returns a standard row view button that opens a drawer.
func ViewBtn(schemaOrBody any) M {
	return M{
		"type":       "button",
		"label":      "View",
		"icon":       "fa fa-eye",
		"actionType": "drawer",
		"drawer": M{
			"title": "Detail",
			"size":  "lg",
			"body":  schemaOrBody,
		},
	}
}

// DeleteBtn returns a standard row delete button with confirmation.
// api is the DELETE endpoint, e.g. "delete:/api/v1/invoices/${id}".
func DeleteBtn(api string) M {
	return M{
		"type":        "button",
		"label":       "Delete",
		"icon":        "fa fa-trash",
		"level":       "danger",
		"actionType":  "ajax",
		"api":         api,
		"confirmText": "Are you sure you want to delete this record?",
	}
}

// CreateBtn returns a toolbar create button that opens a dialog with a form.
func CreateBtn(label, api string, fields ...M) M {
	return M{
		"type":       "button",
		"label":      label,
		"icon":       "fa fa-plus",
		"level":      "primary",
		"actionType": "dialog",
		"dialog": M{
			"title": label,
			"body": M{
				"type": "form",
				"api":  api,
				"body": fields,
			},
		},
	}
}
