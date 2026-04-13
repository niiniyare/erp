package amis

// ---------------------------------------------------------------------------
// Form
// ---------------------------------------------------------------------------

// FormBuilder builds an AMIS `form` component.
type FormBuilder struct{ base }

// Form creates a new form builder. api is the submit endpoint, e.g.
// "post:/api/v1/finance/invoices".
func Form(api string) *FormBuilder {
	s := M{
		"type":          "form",
		"wrapWithPanel": true,
	}
	if api != "" {
		s["api"] = api
	}
	return &FormBuilder{base{s: s}}
}

// Fields sets the form fields.
func (b *FormBuilder) Fields(fields ...M) *FormBuilder {
	b.s["body"] = fields
	return b
}

// Title sets the form title (shown in panel header).
func (b *FormBuilder) Title(t string) *FormBuilder { b.set("title", t); return b }

// Mode sets the form layout mode: "normal" | "horizontal" | "inline".
// Default is "normal" (labels above fields, single column — see Decision 7).
func (b *FormBuilder) Mode(mode string) *FormBuilder { b.set("mode", mode); return b }

// Redirect sets the URL to navigate after successful submit.
func (b *FormBuilder) Redirect(url string) *FormBuilder { b.set("redirect", url); return b }

// ReloadOn sets a target component to reload after submit.
func (b *FormBuilder) ReloadOn(target string) *FormBuilder { b.set("reload", target); return b }

// InitAPI loads initial values from an API when the form mounts.
func (b *FormBuilder) InitAPI(api string) *FormBuilder { b.set("initApi", api); return b }

// Debug shows the form data panel (dev only).
func (b *FormBuilder) Debug() *FormBuilder { b.set("debug", true); return b }

// MarshalJSON implements json.Marshaler.
func (b *FormBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *FormBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Wizard
// ---------------------------------------------------------------------------

// WizardBuilder builds an AMIS `wizard` multi-step form.
type WizardBuilder struct{ base }

// Wizard creates a multi-step wizard. api is the final submit endpoint.
func Wizard(api string) *WizardBuilder {
	b := &WizardBuilder{base{s: M{
		"type": "wizard",
		"api":  api,
		"steps": A{},
	}}}
	return b
}

// Step adds a wizard step. title is the step label; fields are the fields shown.
func (b *WizardBuilder) Step(title string, fields ...M) *WizardBuilder {
	steps := b.s["steps"].(A)
	b.s["steps"] = append(steps, M{
		"title": title,
		"body":  fields,
	})
	return b
}

// ReviewStep adds a final review step (Decision: wizards must end with review).
// It renders a descriptions-style summary of all collected data.
func (b *WizardBuilder) ReviewStep(items ...M) *WizardBuilder {
	steps := b.s["steps"].(A)
	b.s["steps"] = append(steps, M{
		"title": "Review",
		"body": M{
			"type":  "descriptions",
			"items": items,
		},
	})
	return b
}

// MarshalJSON implements json.Marshaler.
func (b *WizardBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *WizardBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Field helpers — return M directly so they compose without extra types
// ---------------------------------------------------------------------------

// TextField creates a text input field.
func TextField(name, label string) M {
	return M{"type": "input-text", "name": name, "label": label}
}

// NumberField creates a number input field.
func NumberField(name, label string) M {
	return M{"type": "input-number", "name": name, "label": label}
}

// TextAreaField creates a multi-line text input.
func TextAreaField(name, label string) M {
	return M{"type": "textarea", "name": name, "label": label}
}

// DateField creates a date picker.
func DateField(name, label string) M {
	return M{"type": "input-date", "name": name, "label": label}
}

// DateTimeField creates a datetime picker.
func DateTimeField(name, label string) M {
	return M{"type": "input-datetime", "name": name, "label": label}
}

// DateRangeField creates a date range picker.
func DateRangeField(name, label string) M {
	return M{"type": "input-date-range", "name": name, "label": label}
}

// SelectField creates a dropdown select. options: [{label, value}].
func SelectField(name, label string, options ...M) M {
	opts := make(A, len(options))
	for i, o := range options {
		opts[i] = o
	}
	return M{"type": "select", "name": name, "label": label, "options": opts}
}

// SelectOpt is a convenience to build a select option.
func SelectOpt(label string, value any) M {
	return M{"label": label, "value": value}
}

// SelectAPIField creates a select whose options are loaded from an API.
func SelectAPIField(name, label, api string) M {
	return M{
		"type":          "select",
		"name":          name,
		"label":         label,
		"source":        api,
		"labelField":    "label",
		"valueField":    "value",
	}
}

// SwitchField creates a boolean toggle switch.
func SwitchField(name, label string) M {
	return M{"type": "switch", "name": name, "label": label}
}

// CheckboxField creates a single checkbox.
func CheckboxField(name, label string) M {
	return M{"type": "checkbox", "name": name, "label": label}
}

// RadioField creates a radio group. options: [{label, value}].
func RadioField(name, label string, options ...M) M {
	opts := make(A, len(options))
	for i, o := range options {
		opts[i] = o
	}
	return M{"type": "radios", "name": name, "label": label, "options": opts}
}

// FileField creates a file upload field.
func FileField(name, label, uploadAPI string) M {
	return M{"type": "input-file", "name": name, "label": label, "receiver": uploadAPI}
}

// ImageField creates an image upload field.
func ImageField(name, label, uploadAPI string) M {
	return M{"type": "input-image", "name": name, "label": label, "receiver": uploadAPI}
}

// HiddenField creates a hidden field that carries data without display.
func HiddenField(name string) M {
	return M{"type": "hidden", "name": name}
}

// ---------------------------------------------------------------------------
// Field modifiers — chain onto any field M
// ---------------------------------------------------------------------------

// Required marks a field as required.
func Required(field M) M {
	field["required"] = true
	return field
}

// Optional marks a field explicitly as optional (shows "(optional)" label).
func Optional(field M) M {
	field["labelRemark"] = M{"type": "remark", "content": "optional"}
	return field
}

// Placeholder sets the input placeholder.
func Placeholder(field M, text string) M {
	field["placeholder"] = text
	return field
}

// Default sets the default value.
func Default(field M, val any) M {
	field["value"] = val
	return field
}

// VisibleOn adds a conditional visibility expression (AMIS expression syntax).
func VisibleOn(field M, expr string) M {
	field["visibleOn"] = expr
	field["clearValueOnHidden"] = true
	return field
}

// DisabledOn adds a conditional disabled expression.
func DisabledOn(field M, expr string) M {
	field["disabledOn"] = expr
	return field
}

// Desc adds a description below the field.
func Desc(field M, text string) M {
	field["description"] = text
	return field
}

// Validate adds a validation rule. rule: "isEmail" | "isUrl" | "isInt" |
// "minimum:N" | "maximum:N" | "minLength:N" | "maxLength:N" | "regex:pattern".
func Validate(field M, rules ...string) M {
	if len(rules) > 0 {
		field["validations"] = rules[0] // AMIS accepts a comma-joined string
	}
	return field
}

// ---------------------------------------------------------------------------
// Divider / Section headers inside forms
// ---------------------------------------------------------------------------

// Divider returns a horizontal rule to separate form sections.
func Divider() M {
	return M{"type": "divider"}
}

// Section returns a static section header inside a form.
func Section(title string) M {
	return M{"type": "static", "label": title, "value": "", "className": "font-semibold text-lg mt-4"}
}
