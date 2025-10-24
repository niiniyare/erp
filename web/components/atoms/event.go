package atoms

// ============================================================================
// ALPINE.JS EVENT HANDLERS
// ============================================================================

// EventProps contains event handling properties for components.
// Alias for AlpinEventHandlers for backward compatibility.
type EventProps = AlpinEventHandlers

// AlpinEventHandlers contains Alpine.js event handling directives.
// These will be rendered as x-on:* attributes.
type AlpinEventHandlers struct {
	OnChange     string `json:"onChange,omitempty"`     // x-on:change
	OnInput      string `json:"onInput,omitempty"`      // x-on:input
	OnFocus      string `json:"onFocus,omitempty"`      // x-on:focus
	OnBlur       string `json:"onBlur,omitempty"`       // x-on:blur
	OnClick      string `json:"onClick,omitempty"`      // x-on:click
	OnKeyDown    string `json:"onKeyDown,omitempty"`    // x-on:keydown
	OnKeyUp      string `json:"onKeyUp,omitempty"`      // x-on:keyup
	OnMouseEnter string `json:"onMouseEnter,omitempty"` // x-on:mouseenter
	OnMouseLeave string `json:"onMouseLeave,omitempty"` // x-on:mouseleave
	OnSubmit     string `json:"onSubmit,omitempty"`     // x-on:submit
	OnLoad       string `json:"onLoad,omitempty"`       // x-on:load
}

// GetEventAttributes returns a map of Alpine.js event attributes.
// Automatically prefixes with x-on: for Alpine.js.
func (a AlpinEventHandlers) GetEventAttributes() map[string]string {
	attrs := make(map[string]string)
	
	// Use helper function to reduce complexity
	addEventAttribute := func(attr, value string) {
		if value != "" {
			attrs["x-on:"+attr] = value
		}
	}

	addEventAttribute("change", a.OnChange)
	addEventAttribute("input", a.OnInput)
	addEventAttribute("focus", a.OnFocus)
	addEventAttribute("blur", a.OnBlur)
	addEventAttribute("click", a.OnClick)
	addEventAttribute("keydown", a.OnKeyDown)
	addEventAttribute("keyup", a.OnKeyUp)
	addEventAttribute("mouseenter", a.OnMouseEnter)
	addEventAttribute("mouseleave", a.OnMouseLeave)
	addEventAttribute("submit", a.OnSubmit)
	addEventAttribute("load", a.OnLoad)

	return attrs
}

// HasEventHandlers checks if any event handlers are defined.
func (a AlpinEventHandlers) HasEventHandlers() bool {
	return a.OnChange != "" || a.OnInput != "" || a.OnFocus != "" ||
		a.OnBlur != "" || a.OnClick != "" || a.OnKeyDown != "" ||
		a.OnKeyUp != "" || a.OnMouseEnter != "" || a.OnMouseLeave != "" ||
		a.OnSubmit != "" || a.OnLoad != ""
}
