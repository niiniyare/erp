// Package sdui generates amis JSON schemas from compiled EntitySchemas.
// amis (Baidu open-source React renderer) interprets these schemas to build
// complete ERP UI without any custom JavaScript.
package sdui

// Page is the top-level amis page component.
type Page struct {
	Type  string `json:"type"` // always "page"
	Title string `json:"title,omitempty"`
	Body  any    `json:"body"`
}

// CRUD is the amis crud component used for list/search views.
type CRUD struct {
	Type         string   `json:"type"` // "crud"
	API          string   `json:"api"`
	Columns      []Column `json:"columns"`
	Filter       *Form    `json:"filter,omitempty"`
	Toolbar      []any    `json:"toolbar,omitempty"`
	PageSize     int      `json:"pageSize,omitempty"`
	Pagination   bool     `json:"pagination"`
	Searchable   bool     `json:"searchable,omitempty"`
	SyncLocation bool     `json:"syncLocation,omitempty"`
}

// Form is the amis form component used for create/edit views.
type Form struct {
	Type       string        `json:"type"` // "form"
	API        string        `json:"api,omitempty"`
	Body       []FormControl `json:"body"`
	SubmitText string        `json:"submitText,omitempty"`
	Reset      bool          `json:"reset,omitempty"`
}

// FormControl is a single field widget inside a Form.
type FormControl struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Required    bool     `json:"required,omitempty"`
	Disabled    bool     `json:"disabled,omitempty"`
	ReadOnly    bool     `json:"readOnly,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Options     []Option `json:"options,omitempty"`
	Source      string   `json:"source,omitempty"` // API URL for dynamic options
	Multiple    bool     `json:"multiple,omitempty"`
	Min         any      `json:"min,omitempty"`
	Max         any      `json:"max,omitempty"`
	Step        any      `json:"step,omitempty"`
	Precision   int      `json:"precision,omitempty"`
	MinRows     int      `json:"minRows,omitempty"`
	MaxRows     int      `json:"maxRows,omitempty"`
	Description string   `json:"description,omitempty"`
}

// Option is a label/value pair for select controls.
type Option struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

// Column is a column descriptor in a CRUD table.
type Column struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Type  string `json:"type,omitempty"` // "text", "date", "datetime", "number", "switch", "tag"
	Width int    `json:"width,omitempty"`
	Fixed string `json:"fixed,omitempty"` // "left" | "right"
}

// Button is an amis action button.
type Button struct {
	Type       string `json:"type"` // "button"
	Label      string `json:"label"`
	ActionType string `json:"actionType"`      // "link", "dialog", "ajax", etc.
	Level      string `json:"level,omitempty"` // "primary", "danger", "default"
	URL        string `json:"url,omitempty"`
	Target     string `json:"target,omitempty"`
}

// NavItem is one sidebar navigation entry.
type NavItem struct {
	Label    string    `json:"label"`
	Icon     string    `json:"icon,omitempty"`
	To       string    `json:"to,omitempty"`
	Children []NavItem `json:"children,omitempty"`
}
