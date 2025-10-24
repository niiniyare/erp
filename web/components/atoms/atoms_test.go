package atoms

import (
	"sync"
	"testing"
)

// ============================================================================
// TYPE TESTS
// ============================================================================

func TestSize_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		size  Size
		valid bool
	}{
		{"valid xs", SizeXS, true},
		{"valid sm", SizeSM, true},
		{"valid md", SizeMD, true},
		{"valid lg", SizeLG, true},
		{"valid xl", SizeXL, true},
		{"invalid", Size("invalid"), false},
		{"empty", Size(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.size.IsValid(); got != tt.valid {
				t.Errorf("Size.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestValidationState_IsError(t *testing.T) {
	tests := []struct {
		name  string
		state ValidationState
		want  bool
	}{
		{"error state", StateError, true},
		{"success state", StateSuccess, false},
		{"default state", StateDefault, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsError(); got != tt.want {
				t.Errorf("ValidationState.IsError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateID_ThreadSafety(t *testing.T) {
	const goroutines = 100
	const iterations = 1000

	ResetIDCounter()

	ids := make(chan string, goroutines*iterations)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				ids <- GenerateID("test")
			}
		}()
	}

	wg.Wait()
	close(ids)

	// Check for duplicates
	seen := make(map[string]bool)
	count := 0
	for id := range ids {
		if seen[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		seen[id] = true
		count++
	}

	expectedCount := goroutines * iterations
	if count != expectedCount {
		t.Errorf("Expected %d unique IDs, got %d", expectedCount, count)
	}
}

func TestEnsureID(t *testing.T) {
	ResetIDCounter()

	tests := []struct {
		name   string
		id     string
		prefix string
	}{
		{"existing id", "custom-id", "test"},
		{"empty id", "", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnsureID(tt.id, tt.prefix)
			if tt.id != "" && result != tt.id {
				t.Errorf("EnsureID() should return existing id, got %v", result)
			}
			if tt.id == "" && result == "" {
				t.Error("EnsureID() should generate id when empty")
			}
		})
	}
}

// ============================================================================
// PROPS TESTS
// ============================================================================

func TestBaseProps_HasID(t *testing.T) {
	tests := []struct {
		name  string
		props BaseProps
		want  bool
	}{
		{"with id", BaseProps{ID: "test-id"}, true},
		{"without id", BaseProps{}, false},
		{"empty id", BaseProps{ID: ""}, false},
		{"whitespace id", BaseProps{ID: "  "}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.props.HasID(); got != tt.want {
				t.Errorf("BaseProps.HasID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseProps_GetClasses(t *testing.T) {
	tests := []struct {
		name              string
		props             BaseProps
		additionalClasses []string
		expectedContains  []string
	}{
		{
			name:              "no classes",
			props:             BaseProps{},
			additionalClasses: nil,
			expectedContains:  nil,
		},
		{
			name:              "base class only",
			props:             BaseProps{Class: "base-class"},
			additionalClasses: nil,
			expectedContains:  []string{"base-class"},
		},
		{
			name:              "additional classes",
			props:             BaseProps{Class: "base"},
			additionalClasses: []string{"additional"},
			expectedContains:  []string{"base", "additional"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.props.GetClasses(tt.additionalClasses...)
			for _, expected := range tt.expectedContains {
				if !HasClass(result, expected) {
					t.Errorf("GetClasses() = %v, should contain %v", result, expected)
				}
			}
		})
	}
}

func TestAccessibilityProps_GetAriaAttributes(t *testing.T) {
	props := AccessibilityProps{
		AriaLabel:    "Test Label",
		AriaRequired: true,
		AriaInvalid:  false,
	}

	attrs := props.GetAriaAttributes()

	if attrs["aria-label"] != "Test Label" {
		t.Errorf("Expected aria-label to be 'Test Label', got %v", attrs["aria-label"])
	}

	if attrs["aria-required"] != "true" {
		t.Errorf("Expected aria-required to be 'true', got %v", attrs["aria-required"])
	}

	if _, exists := attrs["aria-invalid"]; exists {
		t.Error("aria-invalid should not exist when false")
	}
}

func TestValidationProps_GetFeedbackMessage(t *testing.T) {
	tests := []struct {
		name          string
		props         ValidationProps
		expectedMsg   string
		expectedState ValidationState
	}{
		{
			name: "error message",
			props: ValidationProps{
				State:     StateError,
				ErrorText: "Error occurred",
			},
			expectedMsg:   "Error occurred",
			expectedState: StateError,
		},
		{
			name: "success message",
			props: ValidationProps{
				State:       StateSuccess,
				SuccessText: "Success!",
			},
			expectedMsg:   "Success!",
			expectedState: StateSuccess,
		},
		{
			name: "fallback to help text",
			props: ValidationProps{
				State:    StateDefault,
				HelpText: "Helper text",
			},
			expectedMsg:   "Helper text",
			expectedState: StateDefault,
		},
		{
			name:          "no message",
			props:         ValidationProps{State: StateDefault},
			expectedMsg:   "",
			expectedState: StateDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, state := tt.props.GetFeedbackMessage()
			if msg != tt.expectedMsg {
				t.Errorf("GetFeedbackMessage() msg = %v, want %v", msg, tt.expectedMsg)
			}
			if state != tt.expectedState {
				t.Errorf("GetFeedbackMessage() state = %v, want %v", state, tt.expectedState)
			}
		})
	}
}

func TestInteractionProps_IsInteractive(t *testing.T) {
	tests := []struct {
		name  string
		props InteractionProps
		want  bool
	}{
		{"interactive", InteractionProps{}, true},
		{"disabled", InteractionProps{Disabled: true}, false},
		{"readonly", InteractionProps{ReadOnly: true}, false},
		{"both", InteractionProps{Disabled: true, ReadOnly: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.props.IsInteractive(); got != tt.want {
				t.Errorf("InteractionProps.IsInteractive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// VALIDATION TESTS
// ============================================================================

func TestRequiredValidator(t *testing.T) {
	validator := RequiredValidator{Message: "Required field"}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid string", "hello", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"nil value", nil, true},
		{"empty slice", []any{}, true},
		{"non-empty slice", []any{1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequiredValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMinLengthValidator(t *testing.T) {
	validator := MinLengthValidator{MinLength: 5}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid length", "hello world", false},
		{"exact length", "hello", false},
		{"too short", "hi", true},
		{"empty", "", true},
		{"non-string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("MinLengthValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPatternValidator(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		value   string
		wantErr bool
	}{
		{"valid email", EmailRegex, "user@example.com", false},
		{"invalid email", EmailRegex, "not-an-email", true},
		{"valid URL", URLRegex, "https://example.com", false},
		{"invalid URL", URLRegex, "not a url", true},
		{"numeric pattern", NumericRegex, "12345", false},
		{"non-numeric", NumericRegex, "abc123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := PatternValidator{Pattern: tt.pattern}
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("PatternValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailValidator(t *testing.T) {
	validator := EmailValidator{}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid simple", "user@example.com", false},
		{"valid with dots", "user.name@example.com", false},
		{"valid subdomain", "user@mail.example.com", false},
		{"invalid no @", "userexample.com", true},
		{"invalid no domain", "user@", true},
		{"invalid no TLD", "user@example", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRangeValidator(t *testing.T) {
	validator := RangeValidator{Min: 10, Max: 100}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid int", 50, false},
		{"valid float", 50.5, false},
		{"min boundary", 10, false},
		{"max boundary", 100, false},
		{"below min", 5, true},
		{"above max", 150, true},
		{"string number valid", "50", false},
		{"string number invalid", "5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("RangeValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationChain(t *testing.T) {
	chain := ValidationChain{
		Validators: []Validator{
			RequiredValidator{},
			MinLengthValidator{MinLength: 5},
			PatternValidator{Pattern: AlphaRegex},
		},
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "hello", false},
		{"too short", "hi", true},
		{"has numbers", "hello123", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := chain.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidationChain.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// FORM STATE TESTS
// ============================================================================

func TestFormState_RegisterField(t *testing.T) {
	form := NewFormState()
	form.RegisterField("email", RequiredValidator{}, EmailValidator{})

	if _, exists := form.Fields["email"]; !exists {
		t.Error("Field 'email' should be registered")
	}

	if len(form.Fields["email"].Validators) != 2 {
		t.Errorf("Expected 2 validators, got %d", len(form.Fields["email"].Validators))
	}
}

func TestFormState_SetFieldValue(t *testing.T) {
	form := NewFormState()
	form.RegisterField("name")

	form.SetFieldValue("name", "John Doe")

	if form.Fields["name"].Value != "John Doe" {
		t.Errorf("Expected value 'John Doe', got %v", form.Fields["name"].Value)
	}

	if !form.Fields["name"].IsDirty {
		t.Error("Field should be marked as dirty")
	}
}

func TestFormState_ValidateAll(t *testing.T) {
	form := NewFormState()
	form.RegisterField("email", RequiredValidator{}, EmailValidator{})
	form.RegisterField("age", RequiredValidator{}, RangeValidator{Min: 18, Max: 120})

	// Invalid state
	form.SetFieldValue("email", "invalid")
	form.SetFieldValue("age", 15)

	if form.ValidateAll() {
		t.Error("Form should be invalid")
	}

	// Valid state
	form.SetFieldValue("email", "user@example.com")
	form.SetFieldValue("age", 25)

	if !form.ValidateAll() {
		t.Error("Form should be valid")
	}
}

func TestFormState_GetAllErrors(t *testing.T) {
	form := NewFormState()
	form.RegisterField("email", RequiredValidator{}, EmailValidator{})
	form.RegisterField("password", MinLengthValidator{MinLength: 8})

	form.SetFieldValue("email", "invalid")
	form.SetFieldValue("password", "short")

	form.ValidateAll()

	errors := form.GetAllErrors()

	if len(errors) != 2 {
		t.Errorf("Expected 2 fields with errors, got %d", len(errors))
	}

	if _, hasEmailErrors := errors["email"]; !hasEmailErrors {
		t.Error("Email field should have errors")
	}

	if _, hasPasswordErrors := errors["password"]; !hasPasswordErrors {
		t.Error("Password field should have errors")
	}
}

func TestFormState_Reset(t *testing.T) {
	form := NewFormState()
	form.RegisterField("name")

	form.SetFieldValue("name", "Test")
	form.Fields["name"].MarkAsTouched()
	form.SubmitCount = 5

	form.Reset()

	if form.Fields["name"].Value != nil {
		t.Error("Field value should be nil after reset")
	}

	if form.Fields["name"].IsDirty {
		t.Error("Field should not be dirty after reset")
	}

	if form.Fields["name"].IsTouched {
		t.Error("Field should not be touched after reset")
	}

	if form.SubmitCount != 0 {
		t.Errorf("Submit count should be 0, got %d", form.SubmitCount)
	}
}

func TestFormState_JSON(t *testing.T) {
	form := NewFormState()
	form.RegisterField("email")
	form.SetFieldValue("email", "test@example.com")

	// Serialize
	data, err := form.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Deserialize
	newForm := NewFormState()
	newForm.RegisterField("email")
	if err := newForm.FromJSON(data); err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if newForm.Fields["email"].Value != "test@example.com" {
		t.Error("Value should be preserved after JSON round-trip")
	}
}

// ============================================================================
// UTILITY TESTS
// ============================================================================

// func TestJoinClasses(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		classes  []string
// 		expected []string
// 	}{
// 		{
// 			name:     "simple join",
// 			classes:  []string{"class1", "class2"},
// 			expected: []string{"class1", "class2"},
// 		},
// 		{
// 			name:     "remove duplicates",
// 			classes:
