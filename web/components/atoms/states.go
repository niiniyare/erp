package atoms

// ============================================================================
// INTERACTION PROPERTIES
// ============================================================================

// InteractionProps handles user interaction states.
// Controls how users can interact with the component.
type InteractionProps struct {
	Disabled  bool `json:"disabled,omitempty"`  // Component is disabled
	Required  bool `json:"required,omitempty"`  // Field is required
	ReadOnly  bool `json:"readonly,omitempty"`  // Field is read-only
	AutoFocus bool `json:"autofocus,omitempty"` // Auto-focus on page load
}

// IsInteractive checks if the component accepts user input.
func (i InteractionProps) IsInteractive() bool {
	return !i.Disabled && !i.ReadOnly
}

// ShouldShowRequired checks if required indicator should be shown.
func (i InteractionProps) ShouldShowRequired() bool {
	return i.Required && !i.Disabled
}

// GetInteractionAttrs returns HTML attributes for interaction state.
func (i InteractionProps) GetInteractionAttrs() map[string]string {
	attrs := make(map[string]string)

	if i.Disabled {
		attrs["disabled"] = "disabled"
	}
	if i.Required {
		attrs["required"] = "required"
	}
	if i.ReadOnly {
		attrs["readonly"] = "readonly"
	}
	if i.AutoFocus {
		attrs["autofocus"] = "autofocus"
	}

	return attrs
}
