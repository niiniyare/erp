package atoms

import "fmt"

// ============================================================================
// ACCESSIBILITY PROPERTIES
// ============================================================================

// AccessibilityProps contains comprehensive ARIA and accessibility-related attributes.
// These properties improve screen reader support, keyboard navigation, and overall
// accessibility compliance with WCAG 2.1+ and WAI-ARIA 1.2 standards.
//
// Usage:
//
//	props := AccessibilityProps{
//	    AriaLabel: "Close dialog",
//	    Role: "button",
//	    TabIndex: intPtr(0),
//	}
//	attrs := props.GetAriaAttributes()
//
// Note: Pointer types (*bool, *int, *float64) are used for tri-state values where
// the absence of a value has different semantics than false/0.
type AccessibilityProps struct {
	// ========================================================================
	// LABELING & DESCRIPTION
	// ========================================================================

	// AriaLabel provides an accessible name for the element.
	// Use when visible label is not present. Overrides element's text content.
	// Example: "Close button", "Search"
	AriaLabel string `json:"ariaLabel,omitempty"`

	// AriaLabelledBy references the ID(s) of elements that label this element.
	// Space-separated list of IDs. Overrides AriaLabel and element's text.
	// Example: "header-id footer-id"
	AriaLabelledBy string `json:"ariaLabelledBy,omitempty"`

	// AriaDescribedBy references the ID(s) of elements that describe this element.
	// Provides additional context beyond the label. Space-separated list.
	// Example: "error-message-id helper-text-id"
	AriaDescribedBy string `json:"ariaDescribedBy,omitempty"`

	// AriaDetails references the ID of an element providing detailed description.
	// More verbose than AriaDescribedBy, typically for extended descriptions.
	// Example: "detailed-instructions-id"
	AriaDetails string `json:"ariaDetails,omitempty"`

	// AriaPlaceholder provides an accessible placeholder hint.
	// Should not replace a proper label. Use sparingly.
	// Example: "Enter your email address"
	AriaPlaceholder string `json:"ariaPlaceholder,omitempty"`

	// AriaRoleDescription provides a human-readable role description.
	// Overrides the default role announcement. Use sparingly.
	// Example: "slide" for a carousel item with role="group"
	AriaRoleDescription string `json:"ariaRoleDescription,omitempty"`

	// ========================================================================
	// ROLE & SEMANTIC PROPERTIES
	// ========================================================================

	// Role overrides the implicit ARIA role of the element.
	// Common values: button, dialog, alert, navigation, main, complementary, etc.
	// Reference: https://www.w3.org/TR/wai-aria-1.2/#role_definitions
	Role string `json:"role,omitempty"`

	// ========================================================================
	// STATE PROPERTIES
	// ========================================================================

	// AriaRequired marks a field as required for form submission.
	// Only set to true when field is actually required.
	AriaRequired bool `json:"ariaRequired,omitempty"`

	// AriaInvalid marks a field as having invalid or incorrectly formatted value.
	// Set to true when validation fails. Should be accompanied by AriaErrorMessage.
	AriaInvalid bool `json:"ariaInvalid,omitempty"`

	// AriaDisabled marks an element as disabled (non-interactive).
	// Different from native disabled attribute - element remains in tab order.
	AriaDisabled bool `json:"ariaDisabled,omitempty"`

	// AriaReadOnly marks a field as read-only (visible but not editable).
	// User can read and copy but not modify the value.
	AriaReadOnly bool `json:"ariaReadOnly,omitempty"`

	// AriaHidden hides element from assistive technologies.
	// Use when element is decorative or redundant. Does not affect visual display.
	// WARNING: Hidden elements are completely ignored by screen readers.
	AriaHidden bool `json:"ariaHidden,omitempty"`

	// AriaBusy indicates element is being modified and assistive tech should wait.
	// Set to true during async operations, false when complete.
	AriaBusy bool `json:"ariaBusy,omitempty"`

	// AriaModal indicates element is a modal dialog.
	// When true, assistive tech should confine navigation to the modal.
	AriaModal bool `json:"ariaModal,omitempty"`

	// AriaMultiline indicates a textbox accepts multiple lines of input.
	// Only applicable to role="textbox".
	AriaMultiline bool `json:"ariaMultiline,omitempty"`

	// AriaMultiSelectable indicates user can select multiple items.
	// Applicable to grid, listbox, tablist, tree roles.
	AriaMultiSelectable bool `json:"ariaMultiSelectable,omitempty"`

	// ========================================================================
	// TRI-STATE PROPERTIES (pointer = absent/true/false all have meaning)
	// ========================================================================

	// AriaExpanded indicates if a grouping element is expanded or collapsed.
	// nil = not expandable, false = collapsed, true = expanded
	// Applicable to: button, combobox, navigation, etc.
	AriaExpanded *bool `json:"ariaExpanded,omitempty"`

	// AriaPressed indicates the pressed state of a toggle button.
	// nil = not a toggle button, false = not pressed, true = pressed
	// Only applicable to button role.
	AriaPressed *bool `json:"ariaPressed,omitempty"`

	// AriaSelected indicates selection state of an item.
	// nil = not selectable, false = not selected, true = selected
	// Applicable to: gridcell, option, row, tab, etc.
	AriaSelected *bool `json:"ariaSelected,omitempty"`

	// ========================================================================
	// ENUMERATED STATE PROPERTIES
	// ========================================================================

	// AriaChecked indicates the checked state of checkboxes, radio buttons, etc.
	// Valid values: "true", "false", "mixed" (indeterminate state)
	// Empty string = not checkable
	AriaChecked string `json:"ariaChecked,omitempty"`

	// AriaCurrent indicates the current item within a container or set.
	// Valid values: "page", "step", "location", "date", "time", "true", "false"
	// Use "page" for current page in navigation, "step" for multi-step process, etc.
	AriaCurrent string `json:"ariaCurrent,omitempty"`

	// AriaAutocomplete indicates autocomplete behavior of an input.
	// Valid values: "inline", "list", "both", "none"
	// "inline" = text completion, "list" = dropdown, "both" = combined
	AriaAutocomplete string `json:"ariaAutocomplete,omitempty"`

	// AriaHasPopup indicates element can trigger a popup.
	// Valid values: "false", "true", "menu", "listbox", "tree", "grid", "dialog"
	// Specific values provide more context than generic "true".
	AriaHasPopup string `json:"ariaHasPopup,omitempty"`

	// AriaOrientation indicates the element's orientation.
	// Valid values: "horizontal", "vertical", "undefined"
	// Applicable to: scrollbar, select, separator, slider, tablist, toolbar, etc.
	AriaOrientation string `json:"ariaOrientation,omitempty"`

	// AriaSort indicates the sorting direction of a column or row.
	// Valid values: "ascending", "descending", "none", "other"
	// Applicable to columnheader and rowheader roles.
	AriaSort string `json:"ariaSort,omitempty"`

	// ========================================================================
	// RELATIONSHIP PROPERTIES
	// ========================================================================

	// AriaControls references the ID(s) of elements controlled by this element.
	// Space-separated list. Example: a button controlling a disclosure panel.
	AriaControls string `json:"ariaControls,omitempty"`

	// AriaOwns references the ID(s) of elements owned by this element.
	// Establishes parent-child relationship when DOM structure doesn't reflect it.
	// Space-separated list.
	AriaOwns string `json:"ariaOwns,omitempty"`

	// AriaActiveDescendant references the ID of the currently active descendant.
	// Used in composite widgets when focus remains on container but selection changes.
	// Example: listbox where focus stays on container but highlighted option changes.
	AriaActiveDescendant string `json:"ariaActiveDescendant,omitempty"`

	// AriaFlowTo references the ID(s) of next element(s) in alternative reading order.
	// Space-separated list. Overrides default DOM reading order.
	AriaFlowTo string `json:"ariaFlowTo,omitempty"`

	// AriaErrorMessage references the ID of the element containing error message.
	// Should only be set when AriaInvalid is true.
	// Example: "email-error"
	AriaErrorMessage string `json:"ariaErrorMessage,omitempty"`

	// ========================================================================
	// LIVE REGION PROPERTIES
	// ========================================================================

	// AriaLive indicates that updates to the region should be announced.
	// Valid values: "off", "polite", "assertive"
	// "polite" = announce when user is idle, "assertive" = announce immediately
	AriaLive string `json:"ariaLive,omitempty"`

	// AriaAtomic indicates whether assistive tech should present entire region or just changes.
	// true = announce entire region on change, false = announce only changed nodes
	// Only relevant when AriaLive is set.
	AriaAtomic bool `json:"ariaAtomic,omitempty"`

	// AriaRelevant indicates what changes should trigger announcements in live region.
	// Valid values: "additions", "removals", "text", "all" (space-separated combinations)
	// Default: "additions text"
	// Only relevant when AriaLive is set.
	AriaRelevant string `json:"ariaRelevant,omitempty"`

	// ========================================================================
	// VALUE PROPERTIES (for range widgets)
	// ========================================================================

	// AriaValueNow indicates the current numeric value.
	// Required for slider, spinbutton, progressbar, scrollbar, etc.
	// Should be between AriaValueMin and AriaValueMax.
	AriaValueNow *float64 `json:"ariaValueNow,omitempty"`

	// AriaValueMin indicates the minimum allowed value.
	// Used with AriaValueNow for range widgets.
	AriaValueMin *float64 `json:"ariaValueMin,omitempty"`

	// AriaValueMax indicates the maximum allowed value.
	// Used with AriaValueNow for range widgets.
	AriaValueMax *float64 `json:"ariaValueMax,omitempty"`

	// AriaValueText provides human-readable text alternative for AriaValueNow.
	// Use when numeric value doesn't convey full meaning.
	// Example: "Medium" for value 50 on a volume slider
	AriaValueText string `json:"ariaValueText,omitempty"`

	// ========================================================================
	// SET SIZE & POSITION PROPERTIES
	// ========================================================================

	// AriaLevel indicates the hierarchical level of an element.
	// 1-based integer. Used in headings, treeitems, rows, etc.
	// Example: h1 = 1, h2 = 2, nested tree item = 3
	AriaLevel *int `json:"ariaLevel,omitempty"`

	// AriaSetSize indicates the total number of items in a set.
	// Used with AriaPosInSet to convey "item X of Y".
	// Example: 10 for a list of 10 items
	AriaSetSize *int `json:"ariaSetSize,omitempty"`

	// AriaPosInSet indicates the element's position within a set.
	// 1-based integer. Used with AriaSetSize.
	// Example: 3 for "item 3 of 10"
	AriaPosInSet *int `json:"ariaPosInSet,omitempty"`

	// ========================================================================
	// KEYBOARD & FOCUS PROPERTIES
	// ========================================================================

	// TabIndex is defined in BaseProps

	// AriaKeyShortcuts indicates keyboard shortcuts that activate the element.
	// Human-readable string describing the shortcut.
	// Example: "Alt+Shift+T", "Control+S"
	AriaKeyShortcuts string `json:"ariaKeyShortcuts,omitempty"`
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// GetAriaAttributes returns a map of all aria-* and role attributes for HTML rendering.
// Returns only non-empty/non-nil values to keep HTML output clean.
// The returned map can be directly applied to HTML elements as attributes.
//
// Example:
//
//	attrs := props.GetAriaAttributes()
//	// attrs = map[string]string{"aria-label": "Close", "role": "button"}
func (a AccessibilityProps) GetAriaAttributes() map[string]string { //nolint:revive // ARIA attributes require comprehensive handling
	attrs := make(map[string]string)

	// Labeling & Description
	if a.AriaLabel != "" {
		attrs["aria-label"] = a.AriaLabel
	}
	if a.AriaLabelledBy != "" {
		attrs["aria-labelledby"] = a.AriaLabelledBy
	}
	if a.AriaDescribedBy != "" {
		attrs["aria-describedby"] = a.AriaDescribedBy
	}
	if a.AriaDetails != "" {
		attrs["aria-details"] = a.AriaDetails
	}
	if a.AriaPlaceholder != "" {
		attrs["aria-placeholder"] = a.AriaPlaceholder
	}
	if a.AriaRoleDescription != "" {
		attrs["aria-roledescription"] = a.AriaRoleDescription
	}

	// Role
	if a.Role != "" {
		attrs["role"] = a.Role
	}

	// Boolean State Properties (only output when true)
	if a.AriaRequired {
		attrs["aria-required"] = "true"
	}
	if a.AriaInvalid {
		attrs["aria-invalid"] = "true"
	}
	if a.AriaDisabled {
		attrs["aria-disabled"] = "true"
	}
	if a.AriaReadOnly {
		attrs["aria-readonly"] = "true"
	}
	if a.AriaHidden {
		attrs["aria-hidden"] = "true"
	}
	if a.AriaBusy {
		attrs["aria-busy"] = "true"
	}
	if a.AriaModal {
		attrs["aria-modal"] = "true"
	}
	if a.AriaMultiline {
		attrs["aria-multiline"] = "true"
	}
	if a.AriaMultiSelectable {
		attrs["aria-multiselectable"] = "true"
	}
	if a.AriaAtomic {
		attrs["aria-atomic"] = "true"
	}

	// Tri-state Properties (pointer types)
	if a.AriaExpanded != nil {
		attrs["aria-expanded"] = boolToString(*a.AriaExpanded)
	}
	if a.AriaPressed != nil {
		attrs["aria-pressed"] = boolToString(*a.AriaPressed)
	}
	if a.AriaSelected != nil {
		attrs["aria-selected"] = boolToString(*a.AriaSelected)
	}

	// Enumerated State Properties
	if a.AriaChecked != "" {
		attrs["aria-checked"] = a.AriaChecked
	}
	if a.AriaCurrent != "" {
		attrs["aria-current"] = a.AriaCurrent
	}
	if a.AriaAutocomplete != "" {
		attrs["aria-autocomplete"] = a.AriaAutocomplete
	}
	if a.AriaHasPopup != "" {
		attrs["aria-haspopup"] = a.AriaHasPopup
	}
	if a.AriaOrientation != "" {
		attrs["aria-orientation"] = a.AriaOrientation
	}
	if a.AriaSort != "" {
		attrs["aria-sort"] = a.AriaSort
	}

	// Relationship Properties
	if a.AriaControls != "" {
		attrs["aria-controls"] = a.AriaControls
	}
	if a.AriaOwns != "" {
		attrs["aria-owns"] = a.AriaOwns
	}
	if a.AriaActiveDescendant != "" {
		attrs["aria-activedescendant"] = a.AriaActiveDescendant
	}
	if a.AriaFlowTo != "" {
		attrs["aria-flowto"] = a.AriaFlowTo
	}
	if a.AriaErrorMessage != "" {
		attrs["aria-errormessage"] = a.AriaErrorMessage
	}

	// Live Region Properties
	if a.AriaLive != "" {
		attrs["aria-live"] = a.AriaLive
	}
	if a.AriaRelevant != "" {
		attrs["aria-relevant"] = a.AriaRelevant
	}

	// Value Properties
	if a.AriaValueNow != nil {
		attrs["aria-valuenow"] = floatToString(*a.AriaValueNow)
	}
	if a.AriaValueMin != nil {
		attrs["aria-valuemin"] = floatToString(*a.AriaValueMin)
	}
	if a.AriaValueMax != nil {
		attrs["aria-valuemax"] = floatToString(*a.AriaValueMax)
	}
	if a.AriaValueText != "" {
		attrs["aria-valuetext"] = a.AriaValueText
	}

	// Set Size & Position Properties
	if a.AriaLevel != nil {
		attrs["aria-level"] = intToString(*a.AriaLevel)
	}
	if a.AriaSetSize != nil {
		attrs["aria-setsize"] = intToString(*a.AriaSetSize)
	}
	if a.AriaPosInSet != nil {
		attrs["aria-posinset"] = intToString(*a.AriaPosInSet)
	}

	// TabIndex is handled by BaseProps
	if a.AriaKeyShortcuts != "" {
		attrs["aria-keyshortcuts"] = a.AriaKeyShortcuts
	}

	return attrs
}

// HasAriaLabel checks if any form of accessible label is present.
// Returns true if AriaLabel or AriaLabelledBy is set.
// Useful for validation to ensure interactive elements have accessible names.
//
// Example:
//
//	if !props.HasAriaLabel() {
//	    return errors.New("interactive element requires accessible label")
//	}
func (a AccessibilityProps) HasAriaLabel() bool {
	return a.AriaLabel != "" || a.AriaLabelledBy != ""
}

// IsInteractive checks if the element is marked as interactive and not disabled.
// Returns true if element has focusable role or tabindex and is not disabled.
// Useful for determining if keyboard/mouse event handlers should be attached.
//
// Example:
//
//	if props.IsInteractive() {
//	    attachClickHandler()
//	}
func (a AccessibilityProps) IsInteractive() bool {
	// Disabled elements are not interactive
	if a.AriaDisabled {
		return false
	}

	// TabIndex is handled by BaseProps

	// Interactive roles
	interactiveRoles := map[string]bool{
		"button": true, "link": true, "textbox": true, "searchbox": true,
		"combobox": true, "listbox": true, "option": true, "radio": true,
		"checkbox": true, "switch": true, "slider": true, "spinbutton": true,
		"menu": true, "menuitem": true, "menuitemcheckbox": true, "menuitemradio": true,
		"tab": true, "treeitem": true, "gridcell": true,
	}

	return interactiveRoles[a.Role]
}

// HasErrorState checks if the element is in an error state.
// Returns true if AriaInvalid is true, indicating validation failure.
// Should typically be accompanied by AriaErrorMessage.
//
// Example:
//
//	if props.HasErrorState() && props.AriaErrorMessage == "" {
//	    log.Warn("Invalid field missing error message")
//	}
func (a AccessibilityProps) HasErrorState() bool {
	return a.AriaInvalid
}

// IsExpandable checks if the element supports expand/collapse behavior.
// Returns true if AriaExpanded is set (pointer is non-nil).
//
// Example:
//
//	if props.IsExpandable() {
//	    renderExpandIcon()
//	}
func (a AccessibilityProps) IsExpandable() bool {
	return a.AriaExpanded != nil
}

// IsExpanded checks if the element is currently expanded.
// Returns false if not expandable or if collapsed.
// Safe to call even when AriaExpanded is nil.
//
// Example:
//
//	if props.IsExpanded() {
//	    renderExpandedContent()
//	}
func (a AccessibilityProps) IsExpanded() bool {
	return a.AriaExpanded != nil && *a.AriaExpanded
}

// IsLiveRegion checks if the element is a live region.
// Returns true if AriaLive is set to "polite" or "assertive".
// Live regions announce dynamic content changes to screen readers.
//
// Example:
//
//	if props.IsLiveRegion() {
//	    setupMutationObserver()
//	}
func (a AccessibilityProps) IsLiveRegion() bool {
	return a.AriaLive == "polite" || a.AriaLive == "assertive"
}

// HasValueRange checks if the element has numeric value properties.
// Returns true if any of AriaValueNow, AriaValueMin, or AriaValueMax is set.
// Useful for determining if element is a range widget (slider, progressbar, etc.).
//
// Example:
//
//	if props.HasValueRange() {
//	    renderValueIndicator()
//	}
func (a AccessibilityProps) HasValueRange() bool {
	return a.AriaValueNow != nil || a.AriaValueMin != nil || a.AriaValueMax != nil
}

// GetValuePercentage calculates the current value as a percentage of the range.
// Returns 0-100 if all value properties are set, -1 if any are missing.
// Useful for visual progress indicators and sliders.
//
// Example:
//
//	if pct := props.GetValuePercentage(); pct >= 0 {
//	    progressBar.SetWidth(fmt.Sprintf("%d%%", int(pct)))
//	}
func (a AccessibilityProps) GetValuePercentage() float64 {
	if a.AriaValueNow == nil || a.AriaValueMin == nil || a.AriaValueMax == nil {
		return -1
	}

	valueRange := *a.AriaValueMax - *a.AriaValueMin
	if valueRange == 0 {
		return 0
	}

	percentage := ((*a.AriaValueNow - *a.AriaValueMin) / valueRange) * 100

	// Clamp to 0-100
	if percentage < 0 {
		return 0
	}
	if percentage > 100 {
		return 100
	}

	return percentage
}

// SetExpanded is a convenience method to set the expanded state.
// Handles pointer allocation automatically.
//
// Example:
//
//	props.SetExpanded(true)  // Expands the element
//	props.SetExpanded(false) // Collapses the element
func (a *AccessibilityProps) SetExpanded(expanded bool) {
	a.AriaExpanded = &expanded
}

// SetPressed is a convenience method to set the pressed state.
// Handles pointer allocation automatically.
//
// Example:
//
//	props.SetPressed(true)  // Marks button as pressed
func (a *AccessibilityProps) SetPressed(pressed bool) {
	a.AriaPressed = &pressed
}

// SetSelected is a convenience method to set the selected state.
// Handles pointer allocation automatically.
//
// Example:
//
//	props.SetSelected(true)  // Marks option as selected
func (a *AccessibilityProps) SetSelected(selected bool) {
	a.AriaSelected = &selected
}

// SetTabIndex is handled by BaseProps

// SetValueRange is a convenience method to set all value properties at once.
// Handles pointer allocation automatically.
//
// Example:
//
//	props.SetValueRange(50, 0, 100) // Current value 50, range 0-100
func (a *AccessibilityProps) SetValueRange(now, min, max float64) {
	a.AriaValueNow = &now
	a.AriaValueMin = &min
	a.AriaValueMax = &max
}

// SetSetPosition is a convenience method to set both position and set size.
// Handles pointer allocation automatically.
//
// Example:
//
//	props.SetSetPosition(3, 10) // Item 3 of 10
func (a *AccessibilityProps) SetSetPosition(position, setSize int) {
	a.AriaPosInSet = &position
	a.AriaSetSize = &setSize
}

// Validate performs basic validation of accessibility properties.
// Returns an error if properties are inconsistent or invalid.
// Does not validate against all ARIA rules, but catches common mistakes.
//
// Example:
//
//	if err := props.Validate(); err != nil {
//	    return fmt.Errorf("invalid accessibility props: %w", err)
//	}
func (a AccessibilityProps) Validate() error { //nolint:revive // ARIA validation requires comprehensive checks
	// Invalid state should have error message
	if a.AriaInvalid && a.AriaErrorMessage == "" {
		return fmt.Errorf("aria-invalid is true but aria-errormessage is not set")
	}

	// Value range validation
	if a.AriaValueNow != nil && a.AriaValueMin != nil && a.AriaValueMax != nil {
		if *a.AriaValueMin > *a.AriaValueMax {
			return fmt.Errorf("aria-valuemin (%f) cannot be greater than aria-valuemax (%f)",
				*a.AriaValueMin, *a.AriaValueMax)
		}
		if *a.AriaValueNow < *a.AriaValueMin || *a.AriaValueNow > *a.AriaValueMax {
			return fmt.Errorf("aria-valuenow (%f) must be between aria-valuemin (%f) and aria-valuemax (%f)",
				*a.AriaValueNow, *a.AriaValueMin, *a.AriaValueMax)
		}
	}

	// Set position validation
	if a.AriaPosInSet != nil && a.AriaSetSize != nil {
		if *a.AriaPosInSet < 1 {
			return fmt.Errorf("aria-posinset must be >= 1, got %d", *a.AriaPosInSet)
		}
		if *a.AriaSetSize < 1 {
			return fmt.Errorf("aria-setsize must be >= 1, got %d", *a.AriaSetSize)
		}
		if *a.AriaPosInSet > *a.AriaSetSize {
			return fmt.Errorf("aria-posinset (%d) cannot be greater than aria-setsize (%d)",
				*a.AriaPosInSet, *a.AriaSetSize)
		}
	}

	// Level validation
	if a.AriaLevel != nil && *a.AriaLevel < 1 {
		return fmt.Errorf("aria-level must be >= 1, got %d", *a.AriaLevel)
	}

	// Validate enumerated values
	if a.AriaChecked != "" && a.AriaChecked != "true" && a.AriaChecked != "false" && a.AriaChecked != "mixed" {
		return fmt.Errorf("aria-checked must be 'true', 'false', or 'mixed', got '%s'", a.AriaChecked)
	}

	if a.AriaCurrent != "" {
		validCurrent := map[string]bool{
			"page": true, "step": true, "location": true, "date": true, "time": true, "true": true, "false": true,
		}
		if !validCurrent[a.AriaCurrent] {
			return fmt.Errorf("aria-current has invalid value '%s'", a.AriaCurrent)
		}
	}

	if a.AriaAutocomplete != "" {
		validAutocomplete := map[string]bool{"inline": true, "list": true, "both": true, "none": true}
		if !validAutocomplete[a.AriaAutocomplete] {
			return fmt.Errorf("aria-autocomplete has invalid value '%s'", a.AriaAutocomplete)
		}
	}

	if a.AriaLive != "" {
		validLive := map[string]bool{"off": true, "polite": true, "assertive": true}
		if !validLive[a.AriaLive] {
			return fmt.Errorf("aria-live has invalid value '%s'", a.AriaLive)
		}
	}

	return nil
}

// ============================================================================
// INTERNAL HELPER FUNCTIONS
// ============================================================================

// boolToString converts a boolean to "true" or "false" string.
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// intToString converts an integer to string.
func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}

// floatToString converts a float64 to string.
// Removes unnecessary decimal zeros for cleaner output.
func floatToString(f float64) string {
	// Check if it's a whole number
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%g", f)
}

// ============================================================================
// CONVENIENCE CONSTRUCTOR FUNCTIONS
// ============================================================================

// NewAccessibilityProps creates a new AccessibilityProps with commonly used defaults.
// This is a convenience constructor for typical use cases.
//
// Example:
//
//	props := NewAccessibilityProps()
func NewAccessibilityProps() *AccessibilityProps {
	return &AccessibilityProps{}
}

// NewButtonProps creates accessibility props for a button element.
// Sets role="button" and makes it keyboard accessible.
//
// Example:
//
//	props := NewButtonProps("Save changes")
func NewButtonProps(label string) *AccessibilityProps {
	return &AccessibilityProps{
		Role:      "button",
		AriaLabel: label,
		// TabIndex is handled by BaseProps
	}
}

// NewDialogProps creates accessibility props for a modal dialog.
// Sets role="dialog", aria-modal="true", and provides labeling.
//
// Example:
//
//	props := NewDialogProps("confirmDialog", "Confirm Action")
func NewDialogProps(labelledBy, label string) *AccessibilityProps {
	return &AccessibilityProps{
		Role:           "dialog",
		AriaModal:      true,
		AriaLabelledBy: labelledBy,
		AriaLabel:      label,
	}
}

// NewLiveRegionProps creates accessibility props for a live region.
// Use for dynamic content that should be announced to screen readers.
//
// Example:
//
//	props := NewLiveRegionProps("polite", true) // Announce entire region politely
func NewLiveRegionProps(politeness string, atomic bool) *AccessibilityProps {
	return &AccessibilityProps{
		AriaLive:   politeness,
		AriaAtomic: atomic,
	}
}

// NewProgressBarProps creates accessibility props for a progress bar.
// Sets up value range and role.
//
// Example:
//
//	props := NewProgressBarProps(50, 0, 100, "Loading...")
func NewProgressBarProps(now, min, max float64, valueText string) *AccessibilityProps {
	return &AccessibilityProps{
		Role:          "progressbar",
		AriaValueNow:  &now,
		AriaValueMin:  &min,
		AriaValueMax:  &max,
		AriaValueText: valueText,
	}
}
