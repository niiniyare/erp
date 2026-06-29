package ast

import "awo.so/internal/web/ui"

// ─── FormNode ─────────────────────────────────────────────────────────────────

// FormNode renders an AMIS form.
// Maps to AMIS type "form".
//
// Required: at least one node in Body.
// Use API to enable server-side submit. Leave nil for client-only forms.
type FormNode struct {
	// API is the submit endpoint. nil = no auto-submit.
	API *APISpec
	// InitAPI fetches initial form data on mount.
	InitAPI *APISpec
	Body    []Node
	Actions []ActionNode
	// Mode: "normal" (default) | "horizontal" | "inline"
	Mode string
	// Title is an optional form heading.
	Title string
	// WrapWithPanel wraps the form in a Panel card when true.
	WrapWithPanel bool
	// ReloadOn is a comma-separated list of AMIS events that trigger re-init.
	ReloadOn string
}

func (f FormNode) NodeType() string { return "form" }

func (f FormNode) Validate() error {
	if len(f.Body) == 0 {
		return ErrRequiredField("form", "Body")
	}
	if f.API != nil {
		if err := f.API.Validate("form"); err != nil {
			return err
		}
	}
	if f.InitAPI != nil {
		if err := f.InitAPI.Validate("form.initApi"); err != nil {
			return err
		}
	}
	return nil
}

func (f FormNode) Compile() ui.M {
	m := ui.M{
		"type": "form",
		"body": compileNodes(f.Body),
	}
	if f.API != nil {
		m["api"] = f.API.Compile()
	}
	if f.InitAPI != nil {
		m["initApi"] = f.InitAPI.Compile()
	}
	mode := f.Mode
	if mode == "" {
		mode = "normal"
	}
	m["mode"] = mode
	if f.Title != "" {
		m["title"] = f.Title
	}
	if f.WrapWithPanel {
		m["wrapWithPanel"] = true
	}
	if len(f.Actions) > 0 {
		actions := make(ui.A, 0, len(f.Actions))
		for _, a := range f.Actions {
			actions = append(actions, a.Compile())
		}
		m["actions"] = actions
	}
	if f.ReloadOn != "" {
		m["reload"] = f.ReloadOn
	}
	return m
}

func (f FormNode) Children() []Node { return f.Body }

var (
	_ Node          = FormNode{}
	_ ContainerNode = FormNode{}
)

// ─── FilterBarNode ────────────────────────────────────────────────────────────

// FilterBarNode is a specialised form used as the filter header for listing pages.
// Distinct from FormNode — it submits by setting AMIS page-level query params
// rather than calling an API endpoint.
// Maps to AMIS type "form" with submit behaviour "filter".
//
// Required: at least one node in Body.
type FilterBarNode struct {
	Body []Node
	// Collapsible collapses the filter bar under a [Filter] button.
	Collapsible bool
	// ShowCount displays the number of active filters when collapsed.
	ShowCount bool
}

func (f FilterBarNode) NodeType() string { return "form" }

func (f FilterBarNode) Validate() error {
	if len(f.Body) == 0 {
		return ErrRequiredField("filter_bar", "Body")
	}
	return nil
}

func (f FilterBarNode) Compile() ui.M {
	m := ui.M{
		"type":           "form",
		"submitOnChange": true,
		"body":           compileNodes(f.Body),
		"actions":        ui.A{}, // no submit button — change triggers automatically
	}
	if f.Collapsible {
		m["collapsible"] = true
		if f.ShowCount {
			m["showActiveFiltersCount"] = true
		}
	}
	return m
}

func (f FilterBarNode) Children() []Node { return f.Body }

var (
	_ Node          = FilterBarNode{}
	_ ContainerNode = FilterBarNode{}
)

// ─── InputTextNode ────────────────────────────────────────────────────────────

// InputTextNode renders a plain text input field.
// Maps to AMIS type "input-text".
//
// Required: Name.
type InputTextNode struct {
	Name        string
	Label       string
	Required    bool
	Placeholder string
	VisibleOn   string
	DisabledOn  string
	MaxLength   int
	// ClearValue is the value sent when the field is cleared. Default: "".
	ClearValue string
	// Trim strips whitespace on submit when true.
	Trim bool
	// AddOn renders a prefix/suffix label inside the input box.
	AddOn string
}

func (i InputTextNode) NodeType() string { return "input-text" }

func (i InputTextNode) Validate() error {
	if i.Name == "" {
		return ErrRequiredField("input-text", "Name")
	}
	return nil
}

func (i InputTextNode) Compile() ui.M {
	m := ui.M{
		"type": "input-text",
		"name": i.Name,
	}
	setLabel(m, i.Label, i.Name)
	setBool(m, "required", i.Required)
	setStr(m, "placeholder", i.Placeholder)
	setStr(m, "visibleOn", i.VisibleOn)
	setStr(m, "disabledOn", i.DisabledOn)
	if i.MaxLength > 0 {
		m["maxLength"] = i.MaxLength
	}
	setBool(m, "trimContents", i.Trim)
	if i.AddOn != "" {
		m["addOn"] = ui.M{"label": i.AddOn}
	}
	return m
}

var _ Node = InputTextNode{}

// ─── InputNumberNode ─────────────────────────────────────────────────────────

// InputNumberNode renders a numeric input field.
// Maps to AMIS type "input-number".
//
// Required: Name.
type InputNumberNode struct {
	Name       string
	Label      string
	Required   bool
	Min        *float64
	Max        *float64
	Step       float64
	Precision  int // decimal places shown
	VisibleOn  string
	DisabledOn string
	Prefix     string // e.g. "KES"
	Suffix     string // e.g. "%"
}

func (i InputNumberNode) NodeType() string { return "input-number" }

func (i InputNumberNode) Validate() error {
	if i.Name == "" {
		return ErrRequiredField("input-number", "Name")
	}
	return nil
}

func (i InputNumberNode) Compile() ui.M {
	m := ui.M{
		"type": "input-number",
		"name": i.Name,
	}
	setLabel(m, i.Label, i.Name)
	setBool(m, "required", i.Required)
	setStr(m, "visibleOn", i.VisibleOn)
	setStr(m, "disabledOn", i.DisabledOn)
	if i.Min != nil {
		m["min"] = *i.Min
	}
	if i.Max != nil {
		m["max"] = *i.Max
	}
	if i.Step != 0 {
		m["step"] = i.Step
	}
	if i.Precision > 0 {
		m["precision"] = i.Precision
	}
	if i.Prefix != "" {
		m["prefix"] = i.Prefix
	}
	if i.Suffix != "" {
		m["suffix"] = i.Suffix
	}
	return m
}

var _ Node = InputNumberNode{}

// ─── InputDateNode ────────────────────────────────────────────────────────────

// InputDateNode renders a date picker.
// Maps to AMIS type "input-date".
//
// Required: Name.
type InputDateNode struct {
	Name        string
	Label       string
	Required    bool
	Format      string // value format (default: "YYYY-MM-DD")
	InputFormat string // display format (default: "YYYY-MM-DD")
	MinDate     string // e.g. "2020-01-01" or "${start_date}"
	MaxDate     string
	VisibleOn   string
	DisabledOn  string
}

func (i InputDateNode) NodeType() string { return "input-date" }

func (i InputDateNode) Validate() error {
	if i.Name == "" {
		return ErrRequiredField("input-date", "Name")
	}
	return nil
}

func (i InputDateNode) Compile() ui.M {
	m := ui.M{
		"type": "input-date",
		"name": i.Name,
	}
	setLabel(m, i.Label, i.Name)
	setBool(m, "required", i.Required)
	format := i.Format
	if format == "" {
		format = "YYYY-MM-DD"
	}
	m["format"] = format
	inputFmt := i.InputFormat
	if inputFmt == "" {
		inputFmt = "YYYY-MM-DD"
	}
	m["inputFormat"] = inputFmt
	setStr(m, "minDate", i.MinDate)
	setStr(m, "maxDate", i.MaxDate)
	setStr(m, "visibleOn", i.VisibleOn)
	setStr(m, "disabledOn", i.DisabledOn)
	return m
}

var _ Node = InputDateNode{}

// ─── InputDateRangeNode ───────────────────────────────────────────────────────

// InputDateRangeNode renders a date-range picker (start + end date).
// Maps to AMIS type "input-date-range".
//
// Required: Name.
type InputDateRangeNode struct {
	Name       string
	Label      string
	Required   bool
	Format     string // default: "YYYY-MM-DD"
	Delimiter  string // separator between start and end values (default: ",")
	VisibleOn  string
	DisabledOn string
}

func (i InputDateRangeNode) NodeType() string { return "input-date-range" }

func (i InputDateRangeNode) Validate() error {
	if i.Name == "" {
		return ErrRequiredField("input-date-range", "Name")
	}
	return nil
}

func (i InputDateRangeNode) Compile() ui.M {
	m := ui.M{
		"type": "input-date-range",
		"name": i.Name,
	}
	setLabel(m, i.Label, i.Name)
	setBool(m, "required", i.Required)
	format := i.Format
	if format == "" {
		format = "YYYY-MM-DD"
	}
	m["format"] = format
	delim := i.Delimiter
	if delim == "" {
		delim = ","
	}
	m["delimiter"] = delim
	setStr(m, "visibleOn", i.VisibleOn)
	setStr(m, "disabledOn", i.DisabledOn)
	return m
}

var _ Node = InputDateRangeNode{}

// ─── SelectNode ───────────────────────────────────────────────────────────────

// SelectNode renders a dropdown selector.
// Maps to AMIS type "select".
//
// Required: Name. Either Source or Options must be set.
type SelectNode struct {
	Name     string
	Label    string
	Required bool
	// Source is a remote API for options.
	Source *APISpec
	// Options is a static option list. Use when values are fixed at compile time.
	Options    []SelectOption
	ValueField string // default: "value"
	LabelField string // default: "label"
	Multiple   bool
	Clearable  bool
	Searchable bool
	VisibleOn  string
	DisabledOn string
	// DefaultValue is the initial selected value.
	DefaultValue any
	// Placeholder text shown when no value is selected.
	Placeholder string
}

// SelectOption is a static option entry.
type SelectOption struct {
	Label string
	Value any
}

func (s SelectNode) NodeType() string { return "select" }

func (s SelectNode) Validate() error {
	if s.Name == "" {
		return ErrRequiredField("select", "Name")
	}
	if s.Source == nil && len(s.Options) == 0 {
		return ErrInvalidField("select", "Source/Options", "at least one must be set")
	}
	if s.Source != nil {
		if err := s.Source.Validate("select"); err != nil {
			return err
		}
	}
	return nil
}

func (s SelectNode) Compile() ui.M {
	m := ui.M{
		"type": "select",
		"name": s.Name,
	}
	setLabel(m, s.Label, s.Name)
	setBool(m, "required", s.Required)
	setBool(m, "multiple", s.Multiple)
	setBool(m, "clearable", s.Clearable)
	setBool(m, "searchable", s.Searchable)
	setStr(m, "visibleOn", s.VisibleOn)
	setStr(m, "disabledOn", s.DisabledOn)
	setStr(m, "placeholder", s.Placeholder)

	vf := s.ValueField
	if vf == "" {
		vf = "value"
	}
	m["valueField"] = vf
	lf := s.LabelField
	if lf == "" {
		lf = "label"
	}
	m["labelField"] = lf

	if s.Source != nil {
		m["source"] = s.Source.Compile()
	}
	if len(s.Options) > 0 {
		opts := make(ui.A, 0, len(s.Options))
		for _, opt := range s.Options {
			opts = append(opts, ui.M{"label": opt.Label, "value": opt.Value})
		}
		m["options"] = opts
	}
	if s.DefaultValue != nil {
		m["value"] = s.DefaultValue
	}
	return m
}

var _ Node = SelectNode{}

// ─── ComboNode ────────────────────────────────────────────────────────────────

// ComboNode renders a repeatable group of fields (used for line item rows, etc.).
// Maps to AMIS type "combo".
//
// Required: Name and at least one node in Items.
type ComboNode struct {
	Name  string
	Label string
	Items []Node // fields within each row
	// Multiple allows adding multiple rows.
	Multiple bool
	// AddButtonLabel overrides the default "+" button text.
	AddButtonLabel string
	// MaxLength caps the number of rows. 0 = unlimited.
	MaxLength  int
	VisibleOn  string
	DisabledOn string
}

func (c ComboNode) NodeType() string { return "combo" }

func (c ComboNode) Validate() error {
	if c.Name == "" {
		return ErrRequiredField("combo", "Name")
	}
	if len(c.Items) == 0 {
		return ErrRequiredField("combo", "Items")
	}
	return nil
}

func (c ComboNode) Compile() ui.M {
	m := ui.M{
		"type":  "combo",
		"name":  c.Name,
		"items": compileNodes(c.Items),
	}
	setLabel(m, c.Label, c.Name)
	setBool(m, "multiple", c.Multiple)
	setStr(m, "visibleOn", c.VisibleOn)
	setStr(m, "disabledOn", c.DisabledOn)
	if c.AddButtonLabel != "" {
		m["addButtonLabel"] = c.AddButtonLabel
	}
	if c.MaxLength > 0 {
		m["maxLength"] = c.MaxLength
	}
	return m
}

func (c ComboNode) Children() []Node { return c.Items }

var (
	_ Node          = ComboNode{}
	_ ContainerNode = ComboNode{}
)

// ─── MultiSelectNode ──────────────────────────────────────────────────────────

// MultiSelectNode is a convenience wrapper around SelectNode with Multiple=true.
// Provided so DSL code reads clearly: MultiSelectNode vs SelectNode.
//
// Required: Name. Either Source or Options must be set.
type MultiSelectNode struct {
	Name       string
	Label      string
	Required   bool
	Source     *APISpec
	Options    []SelectOption
	ValueField string
	LabelField string
	Searchable bool
	Clearable  bool
	Delimiter  string // value separator in submitted data (default ",")
	VisibleOn  string
	DisabledOn string
}

func (m MultiSelectNode) NodeType() string { return "select" }

func (m MultiSelectNode) Validate() error {
	if m.Name == "" {
		return ErrRequiredField("multi-select", "Name")
	}
	if m.Source == nil && len(m.Options) == 0 {
		return ErrInvalidField("multi-select", "Source/Options", "at least one must be set")
	}
	if m.Source != nil {
		if err := m.Source.Validate("multi-select"); err != nil {
			return err
		}
	}
	return nil
}

func (m MultiSelectNode) Compile() ui.M {
	inner := SelectNode{
		Name:       m.Name,
		Label:      m.Label,
		Required:   m.Required,
		Source:     m.Source,
		Options:    m.Options,
		ValueField: m.ValueField,
		LabelField: m.LabelField,
		Multiple:   true,
		Clearable:  m.Clearable,
		Searchable: m.Searchable,
		VisibleOn:  m.VisibleOn,
		DisabledOn: m.DisabledOn,
	}
	out := inner.Compile()
	if m.Delimiter != "" {
		out["delimiter"] = m.Delimiter
	}
	return out
}

var _ Node = MultiSelectNode{}

// ─── CheckboxNode ─────────────────────────────────────────────────────────────

// CheckboxNode renders a checkbox field.
// Maps to AMIS type "checkbox".
//
// Required: Name.
type CheckboxNode struct {
	Name  string
	Label string
	// Option is the text shown next to the checkbox (different from Label which is
	// the field label in form layout).
	Option     string
	Required   bool
	VisibleOn  string
	DisabledOn string
}

func (c CheckboxNode) NodeType() string { return "checkbox" }

func (c CheckboxNode) Validate() error {
	if c.Name == "" {
		return ErrRequiredField("checkbox", "Name")
	}
	return nil
}

func (c CheckboxNode) Compile() ui.M {
	m := ui.M{
		"type": "checkbox",
		"name": c.Name,
	}
	setLabel(m, c.Label, c.Name)
	setBool(m, "required", c.Required)
	setStr(m, "option", c.Option)
	setStr(m, "visibleOn", c.VisibleOn)
	setStr(m, "disabledOn", c.DisabledOn)
	return m
}

var _ Node = CheckboxNode{}

// ─── ToggleNode ───────────────────────────────────────────────────────────────

// ToggleNode renders a toggle switch (boolean field).
// Maps to AMIS type "switch".
//
// Required: Name.
type ToggleNode struct {
	Name       string
	Label      string
	TrueValue  any // value submitted when on (default: true)
	FalseValue any // value submitted when off (default: false)
	VisibleOn  string
	DisabledOn string
}

func (t ToggleNode) NodeType() string { return "switch" }

func (t ToggleNode) Validate() error {
	if t.Name == "" {
		return ErrRequiredField("switch", "Name")
	}
	return nil
}

func (t ToggleNode) Compile() ui.M {
	m := ui.M{
		"type": "switch",
		"name": t.Name,
	}
	setLabel(m, t.Label, t.Name)
	if t.TrueValue != nil {
		m["trueValue"] = t.TrueValue
	}
	if t.FalseValue != nil {
		m["falseValue"] = t.FalseValue
	}
	setStr(m, "visibleOn", t.VisibleOn)
	setStr(m, "disabledOn", t.DisabledOn)
	return m
}

var _ Node = ToggleNode{}

// ─── DialogNode ───────────────────────────────────────────────────────────────

// DialogNode wraps content in an AMIS modal dialog.
// Maps to AMIS type "dialog".
//
// Required: Title and at least one node in Body.
type DialogNode struct {
	Title   string
	Body    []Node
	Actions []ActionNode
	// Size: "xs" | "sm" | "md" (default) | "lg" | "xl" | "full"
	Size string
	// CloseOnEsc allows closing the dialog with the Escape key.
	CloseOnEsc bool
}

func (d DialogNode) NodeType() string { return "dialog" }

func (d DialogNode) Validate() error {
	if d.Title == "" {
		return ErrRequiredField("dialog", "Title")
	}
	if len(d.Body) == 0 {
		return ErrRequiredField("dialog", "Body")
	}
	return nil
}

func (d DialogNode) Compile() ui.M {
	m := ui.M{
		"type":  "dialog",
		"title": d.Title,
		"body":  compileNodes(d.Body),
	}
	size := d.Size
	if size == "" {
		size = "md"
	}
	m["size"] = size
	if len(d.Actions) > 0 {
		actions := make(ui.A, 0, len(d.Actions))
		for _, a := range d.Actions {
			actions = append(actions, a.Compile())
		}
		m["actions"] = actions
	}
	if d.CloseOnEsc {
		m["closeOnEsc"] = true
	}
	return m
}

func (d DialogNode) Children() []Node { return d.Body }

var (
	_ Node          = DialogNode{}
	_ ContainerNode = DialogNode{}
)

// ─── DrawerNode ───────────────────────────────────────────────────────────────

// DrawerNode wraps content in an AMIS side-drawer panel.
// Maps to AMIS type "drawer".
//
// Required: Title and at least one node in Body.
type DrawerNode struct {
	Title   string
	Body    []Node
	Actions []ActionNode
	// Position: "right" (default) | "left" | "top" | "bottom"
	Position string
	// Width is a CSS width value (e.g. "500px", "40%"). Default: "500px".
	Width string
}

func (d DrawerNode) NodeType() string { return "drawer" }

func (d DrawerNode) Validate() error {
	if d.Title == "" {
		return ErrRequiredField("drawer", "Title")
	}
	if len(d.Body) == 0 {
		return ErrRequiredField("drawer", "Body")
	}
	return nil
}

func (d DrawerNode) Compile() ui.M {
	pos := d.Position
	if pos == "" {
		pos = "right"
	}
	w := d.Width
	if w == "" {
		w = "500px"
	}
	m := ui.M{
		"type":     "drawer",
		"title":    d.Title,
		"body":     compileNodes(d.Body),
		"position": pos,
		"width":    w,
	}
	if len(d.Actions) > 0 {
		actions := make(ui.A, 0, len(d.Actions))
		for _, a := range d.Actions {
			actions = append(actions, a.Compile())
		}
		m["actions"] = actions
	}
	return m
}

func (d DrawerNode) Children() []Node { return d.Body }

var (
	_ Node          = DrawerNode{}
	_ ContainerNode = DrawerNode{}
)

// ─── WizardNode ───────────────────────────────────────────────────────────────

// WizardNode renders a multi-step form wizard.
// Maps to AMIS type "wizard".
//
// Required: at least two Steps and API.
// Each step renders its own Body fields; the wizard advances on "Next".
// The final step's submit calls API.
type WizardNode struct {
	// API is the submit endpoint called on final step completion. Required.
	API APISpec
	// Steps are the ordered wizard panels. Required: at least 2.
	Steps []WizardStep
	// Mode: "horizontal" (default) | "vertical"
	Mode string
	// StartStep is the initial step index (0-based). Default: 0.
	StartStep int
}

// WizardStep is one panel in a WizardNode.
type WizardStep struct {
	// Title is the step label in the wizard progress bar. Required.
	Title string
	// SubTitle is optional descriptive text under the title.
	SubTitle string
	// Body is the form fields rendered in this step.
	Body []Node
	// API is an optional per-step submit endpoint (overrides wizard-level API for this step only).
	API *APISpec
	// VisibleOn is a boolean AMIS expression; hides the step when false.
	VisibleOn string
}

func (w WizardNode) NodeType() string { return "wizard" }

func (w WizardNode) Validate() error {
	if err := w.API.Validate("wizard"); err != nil {
		return err
	}
	if len(w.Steps) < 2 {
		return ErrInvalidField("wizard", "Steps", "wizard must have at least 2 steps")
	}
	for i, step := range w.Steps {
		if step.Title == "" {
			return ErrInvalidField("wizard", "Steps", formatColumnErr(i, "Title must not be empty"))
		}
	}
	return nil
}

func (w WizardNode) Compile() ui.M {
	steps := make(ui.A, 0, len(w.Steps))
	for _, step := range w.Steps {
		s := ui.M{"title": step.Title}
		if step.SubTitle != "" {
			s["subTitle"] = step.SubTitle
		}
		if len(step.Body) > 0 {
			s["body"] = compileNodes(step.Body)
		}
		if step.API != nil {
			s["api"] = step.API.Compile()
		}
		if step.VisibleOn != "" {
			s["visibleOn"] = step.VisibleOn
		}
		steps = append(steps, s)
	}
	mode := w.Mode
	if mode == "" {
		mode = "horizontal"
	}
	m := ui.M{
		"type":  "wizard",
		"api":   w.API.Compile(),
		"steps": steps,
		"mode":  mode,
	}
	if w.StartStep > 0 {
		m["startStep"] = w.StartStep
	}
	return m
}

func (w WizardNode) Children() []Node {
	var all []Node
	for _, step := range w.Steps {
		all = append(all, step.Body...)
	}
	return all
}

var (
	_ Node          = WizardNode{}
	_ ContainerNode = WizardNode{}
)

// ─── PickerNode ───────────────────────────────────────────────────────────────

// PickerNode renders a pop-up entity selector (cross-entity reference picker).
// Maps to AMIS type "picker".
//
// Used for cross-entity selectors: "Select Supplier", "Select Customer", etc.
// Opens a CRUD list in a dialog; the selected row value is written to Name.
//
// Required: Name and Source.
type PickerNode struct {
	// Name is the field key that receives the selected value. Required.
	Name string
	// Label is the field label shown in the form.
	Label string
	// Source is the API URL that powers the picker's CRUD list. Required.
	Source string
	// Columns are the columns shown in the picker dialog.
	Columns []TableColumn
	// ValueField is the field from the selected row used as the submitted value (default: "id").
	ValueField string
	// LabelField is the field shown as the selected item label (default: "name").
	LabelField string
	// Multiple allows selecting multiple items.
	Multiple bool
	// Required marks the field as mandatory.
	Required bool
	// VisibleOn is a boolean AMIS expression controlling visibility.
	VisibleOn string
	// DisabledOn is a boolean AMIS expression controlling disabled state.
	DisabledOn string
	// Embed renders the CRUD list inline instead of in a dialog when true.
	Embed bool
}

func (p PickerNode) NodeType() string { return "picker" }

func (p PickerNode) Validate() error {
	if p.Name == "" {
		return ErrRequiredField("picker", "Name")
	}
	if p.Source == "" {
		return ErrRequiredField("picker", "Source")
	}
	return nil
}

func (p PickerNode) Compile() ui.M {
	m := ui.M{
		"type":   "picker",
		"name":   p.Name,
		"source": p.Source,
	}
	setLabel(m, p.Label, p.Name)
	vf := p.ValueField
	if vf == "" {
		vf = "id"
	}
	m["valueField"] = vf
	lf := p.LabelField
	if lf == "" {
		lf = "name"
	}
	m["labelField"] = lf
	setBool(m, "multiple", p.Multiple)
	setBool(m, "required", p.Required)
	setBool(m, "embed", p.Embed)
	setStr(m, "visibleOn", p.VisibleOn)
	setStr(m, "disabledOn", p.DisabledOn)
	if len(p.Columns) > 0 {
		cols := make(ui.A, 0, len(p.Columns))
		for _, col := range p.Columns {
			cols = append(cols, col.compile())
		}
		m["pickerSchema"] = ui.M{
			"type":    "crud",
			"columns": cols,
		}
	}
	return m
}

var _ Node = PickerNode{}

// ─── field helper utilities ───────────────────────────────────────────────────

// setLabel sets m["label"] if label is non-empty, otherwise falls back to fallback.
func setLabel(m ui.M, label, fallback string) {
	if label != "" {
		m["label"] = label
	} else if fallback != "" {
		m["label"] = fallback
	}
}

// setStr sets m[key] = val only when val is non-empty.
func setStr(m ui.M, key, val string) {
	if val != "" {
		m[key] = val
	}
}

// setBool sets m[key] = true only when val is true.
func setBool(m ui.M, key string, val bool) {
	if val {
		m[key] = true
	}
}
