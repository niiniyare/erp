package bridge

import (
	"context"
	"fmt"
	"sync"

	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
	"github.com/niiniyare/erp/web/components/atoms"
)

// ============================================================================
// STEP 1: ADD VALIDATION CHAIN
// ============================================================================

// Validator defines a validation step in the pipeline.
type Validator interface {
	Validate(ctx context.Context, component any) error
	SetNext(Validator)
}

// BaseValidator provides chain implementation.
type BaseValidator struct {
	next Validator
}

func (v *BaseValidator) SetNext(next Validator) {
	v.next = next
}

// Validate provides default implementation that calls next validator
func (v *BaseValidator) Validate(ctx context.Context, component any) error {
	return v.callNext(ctx, component)
}

func (v *BaseValidator) callNext(ctx context.Context, component any) error {
	if v.next != nil {
		return v.next.Validate(ctx, component)
	}
	return nil
}

// RequiredFieldsValidator checks for required fields.
type RequiredFieldsValidator struct {
	BaseValidator
}

func (v *RequiredFieldsValidator) Validate(ctx context.Context, component any) error {
	switch c := component.(type) {
	case atoms.ButtonProps:
		if c.Text == "" && c.Icon.Name == "" {
			return fmt.Errorf("button must have text or icon")
		}
	case atoms.InputProps:
		if c.BaseProps.Name == "" {
			return fmt.Errorf("input must have name attribute")
		}
	case atoms.SelectProps:
		if c.BaseProps.Name == "" {
			return fmt.Errorf("select must have name attribute")
		}
		if len(c.Options) == 0 {
			return fmt.Errorf("select must have at least one option")
		}
	case atoms.CheckboxProps:
		if c.BaseProps.Name == "" {
			return fmt.Errorf("checkbox must have name attribute")
		}
	case atoms.RadioGroupProps:
		if c.BaseProps.Name == "" {
			return fmt.Errorf("radio must have name attribute")
		}
		// if len(c.Options) == 0 {
		// 	return fmt.Errorf("radio must have at least one option")
		// }
	}
	return v.callNext(ctx, component)
}

// AccessibilityValidator checks accessibility requirements.
type AccessibilityValidator struct {
	BaseValidator
}

func (v *AccessibilityValidator) Validate(ctx context.Context, component any) error {
	// Check for aria-label or label on form elements
	switch c := component.(type) {
	case atoms.InputProps:
		if c.AriaLabel == "" && c.Label == "" && c.Placeholder == "" {
			return fmt.Errorf("input should have aria-label, label, or placeholder for accessibility")
		}
	case atoms.ButtonProps:
		if c.Text == "" && c.AriaLabel == "" {
			return fmt.Errorf("icon-only button should have aria-label")
		}
	}
	return v.callNext(ctx, component)
}

// SizeValidator validates size constraints.
type SizeValidator struct {
	BaseValidator
}

func (v *SizeValidator) Validate(ctx context.Context, component any) error {
	// Validate size-related constraints
	switch c := component.(type) {
	case atoms.TextareaProps:
		if c.Rows < 0 || c.Cols < 0 {
			return fmt.Errorf("textarea rows and cols must be non-negative")
		}
		if c.MaxLength < 0 {
			return fmt.Errorf("textarea maxLength must be non-negative")
		}
	case atoms.InputProps:
		if c.MaxLength > 0 && c.MinLength > c.MaxLength {
			return fmt.Errorf("input minLength cannot exceed maxLength")
		}
	}
	return v.callNext(ctx, component)
}

// ValidationChain manages the validation pipeline.
type ValidationChain struct {
	head Validator
	mu   sync.RWMutex
}

func NewValidationChain() *ValidationChain {
	required := &RequiredFieldsValidator{}
	accessibility := &AccessibilityValidator{}
	size := &SizeValidator{}

	required.SetNext(accessibility)
	accessibility.SetNext(size)

	return &ValidationChain{head: required}
}

func (c *ValidationChain) Validate(ctx context.Context, component any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.head != nil {
		return c.head.Validate(ctx, component)
	}
	return nil
}

func (c *ValidationChain) AddValidator(v Validator) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.head == nil {
		c.head = v
		return
	}

	// Add to end of chain
	current := c.head
	for {
		if base, ok := current.(*BaseValidator); ok && base.next != nil {
			current = base.next
		} else {
			current.SetNext(v)
			break
		}
	}
}

// ============================================================================
// STEP 2: ENHANCE BRIDGE WITH VALIDATION
// ============================================================================

// ConvertToTemplWithValidation converts and validates a component
func (b *Bridge) ConvertToTemplWithValidation(ctx context.Context, component schemaui.Component) (TemplComponent, error) {
	// Convert
	templComponent, err := b.ConvertToTempl(ctx, component)
	if err != nil {
		return TemplComponent{}, err
	}

	// Create and use validation chain
	validator := NewValidationChain()
	if err := validator.Validate(ctx, templComponent.Props); err != nil {
		return TemplComponent{}, fmt.Errorf("validation failed: %w", err)
	}

	return templComponent, nil
}

func (b *Bridge) ConvertToSchemaWithValidation(ctx context.Context, component TemplComponent) (schemaui.Component, error) {
	// Validate first
	validator := NewValidationChain()
	if err := validator.Validate(ctx, component.Props); err != nil {
		return schemaui.Component{}, fmt.Errorf("validation failed: %w", err)
	}

	// Convert
	return b.ConvertTemplToSchema(ctx, component)
}

// ============================================================================
// STEP 3: ADD COMPONENT VISITOR FOR TREE OPERATIONS
// ============================================================================

// ComponentVisitor defines operations on component trees.
type ComponentVisitor interface {
	VisitButton(ctx context.Context, props atoms.ButtonProps) error
	VisitInput(ctx context.Context, props atoms.InputProps) error
	VisitTextarea(ctx context.Context, props atoms.TextareaProps) error
	VisitSelect(ctx context.Context, props atoms.SelectProps) error
	VisitCheckbox(ctx context.Context, props atoms.CheckboxProps) error
	VisitRadio(ctx context.Context, props atoms.RadioProps) error
}

// BaseVisitor provides default no-op implementations.
type BaseVisitor struct{}

func (v *BaseVisitor) VisitButton(ctx context.Context, props atoms.ButtonProps) error     { return nil }
func (v *BaseVisitor) VisitInput(ctx context.Context, props atoms.InputProps) error       { return nil }
func (v *BaseVisitor) VisitTextarea(ctx context.Context, props atoms.TextareaProps) error { return nil }
func (v *BaseVisitor) VisitSelect(ctx context.Context, props atoms.SelectProps) error     { return nil }
func (v *BaseVisitor) VisitCheckbox(ctx context.Context, props atoms.CheckboxProps) error { return nil }
func (v *BaseVisitor) VisitRadio(ctx context.Context, props atoms.RadioProps) error       { return nil }

// AcceptVisitor allows a component to accept visitors.
func AcceptVisitor(ctx context.Context, component TemplComponent, visitor ComponentVisitor) error {
	switch component.Type {
	case "Button":
		if props, ok := component.Props.(atoms.ButtonProps); ok {
			return visitor.VisitButton(ctx, props)
		}
	case "Input":
		if props, ok := component.Props.(atoms.InputProps); ok {
			return visitor.VisitInput(ctx, props)
		}
	case "Textarea":
		if props, ok := component.Props.(atoms.TextareaProps); ok {
			return visitor.VisitTextarea(ctx, props)
		}
	case "Select":
		if props, ok := component.Props.(atoms.SelectProps); ok {
			return visitor.VisitSelect(ctx, props)
		}
	case "Checkbox":
		if props, ok := component.Props.(atoms.CheckboxProps); ok {
			return visitor.VisitCheckbox(ctx, props)
		}
	case "Radio":
		if props, ok := component.Props.(atoms.RadioProps); ok {
			return visitor.VisitRadio(ctx, props)
		}
	}
	return fmt.Errorf("unknown component type: %s", component.Type)
}

// ============================================================================
// STEP 4: CONCRETE VISITOR IMPLEMENTATIONS
// ============================================================================

// CodeGeneratorVisitor generates Templ code from components.
type CodeGeneratorVisitor struct {
	BaseVisitor
	output []string
}

func NewCodeGeneratorVisitor() *CodeGeneratorVisitor {
	return &CodeGeneratorVisitor{
		output: make([]string, 0),
	}
}

func (v *CodeGeneratorVisitor) VisitButton(ctx context.Context, props atoms.ButtonProps) error {
	code := fmt.Sprintf(`@atoms.Button(atoms.ButtonProps{
		Text: "%s",
		Variant: atoms.%s,
		Size: atoms.%s,
	})`, props.Text, props.Variant, props.Size)
	v.output = append(v.output, code)
	return nil
}

func (v *CodeGeneratorVisitor) VisitInput(ctx context.Context, props atoms.InputProps) error {
	code := fmt.Sprintf(`@atoms.Input(atoms.InputProps{
		Type: "%s",
		Name: "%s",
		Placeholder: "%s",
	})`, props.Type, props.Name, props.Placeholder)
	v.output = append(v.output, code)
	return nil
}

func (v *CodeGeneratorVisitor) GetOutput() string {
	result := "templ GeneratedComponent() {\n"
	for _, line := range v.output {
		result += "\t" + line + "\n"
	}
	result += "}\n"
	return result
}

// MetricsVisitor collects metrics about component usage.
type MetricsVisitor struct {
	BaseVisitor
	componentCounts map[string]int
	totalRequired   int
	totalDisabled   int
	mu              sync.Mutex
}

func NewMetricsVisitor() *MetricsVisitor {
	return &MetricsVisitor{
		componentCounts: make(map[string]int),
	}
}

func (v *MetricsVisitor) VisitButton(ctx context.Context, props atoms.ButtonProps) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.componentCounts["Button"]++
	if props.Disabled {
		v.totalDisabled++
	}
	return nil
}

func (v *MetricsVisitor) VisitInput(ctx context.Context, props atoms.InputProps) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.componentCounts["Input"]++
	if props.Required {
		v.totalRequired++
	}
	if props.Disabled {
		v.totalDisabled++
	}
	return nil
}

func (v *MetricsVisitor) VisitSelect(ctx context.Context, props atoms.SelectProps) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.componentCounts["Select"]++
	if props.Required {
		v.totalRequired++
	}
	if props.Disabled {
		v.totalDisabled++
	}
	return nil
}

func (v *MetricsVisitor) GetMetrics() map[string]any {
	v.mu.Lock()
	defer v.mu.Unlock()
	return map[string]any{
		"component_counts": v.componentCounts,
		"total_required":   v.totalRequired,
		"total_disabled":   v.totalDisabled,
	}
}

// AccessibilityAuditVisitor checks for accessibility issues.
type AccessibilityAuditVisitor struct {
	BaseVisitor
	issues []string
	mu     sync.Mutex
}

func NewAccessibilityAuditVisitor() *AccessibilityAuditVisitor {
	return &AccessibilityAuditVisitor{
		issues: make([]string, 0),
	}
}

func (v *AccessibilityAuditVisitor) VisitButton(ctx context.Context, props atoms.ButtonProps) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if props.Icon.Name != "" && props.Text == "" && props.AccessibilityProps.AriaLabel == "" {
		v.issues = append(v.issues, "Icon-only button missing aria-label")
	}
	return nil
}

func (v *AccessibilityAuditVisitor) VisitInput(ctx context.Context, props atoms.InputProps) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if props.Label == "" && props.AriaLabel == "" {
		v.issues = append(v.issues, fmt.Sprintf("Input '%s' missing label or aria-label", props.Name))
	}
	return nil
}

func (v *AccessibilityAuditVisitor) GetIssues() []string {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.issues
}

// ============================================================================
// STEP 5: ADD MIDDLEWARE PATTERN FOR TRANSFORMATIONS
// ============================================================================

// Middleware can transform components during conversion.
type Middleware func(ctx context.Context, component TemplComponent) (TemplComponent, error)

// MiddlewareChain manages middleware execution.
type MiddlewareChain struct {
	middlewares []Middleware
	mu          sync.RWMutex
}

func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]Middleware, 0),
	}
}

func (c *MiddlewareChain) Use(m Middleware) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.middlewares = append(c.middlewares, m)
}

func (c *MiddlewareChain) Execute(ctx context.Context, component TemplComponent) (TemplComponent, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := component
	var err error

	for _, middleware := range c.middlewares {
		result, err = middleware(ctx, result)
		if err != nil {
			return TemplComponent{}, err
		}
	}

	return result, nil
}

// Example middleware: Add CSS classes
func AddDefaultClassesMiddleware(ctx context.Context, component TemplComponent) (TemplComponent, error) {
	switch props := component.Props.(type) {
	case atoms.ButtonProps:
		if props.Class == "" {
			props.Class = "btn-default"
		}
		component.Props = props
	case atoms.InputProps:
		if props.Class == "" {
			props.Class = "input-default"
		}
		component.Props = props
	}
	return component, nil
}

// Example middleware: Add security attributes
func SecurityAttributesMiddleware(ctx context.Context, component TemplComponent) (TemplComponent, error) {
	switch props := component.Props.(type) {
	case atoms.InputProps:
		if props.Type == "password" {
			// Ensure autocomplete is off for passwords
			// This would require extending the InputProps struct
		}
		component.Props = props
	}
	return component, nil
}

// ============================================================================
// STEP 6: USAGE EXAMPLES
// ============================================================================

// Example: Converting with validation
func ExampleConvertWithValidation(b *Bridge) {
	ctx := context.Background()

	schemaComponent := schemaui.Component{
		BaseComponent: schemaui.BaseComponent{
			Type: schemaui.ComponentButton,
		},
		// ... other fields
	}

	templComponent, err := b.ConvertToTemplWithValidation(ctx, schemaComponent)
	if err != nil {
		fmt.Printf("Conversion/validation failed: %v\n", err)
		return
	}

	fmt.Printf("Successfully converted and validated: %+v\n", templComponent)
}

// Example: Using visitor for code generation
func ExampleCodeGeneration(components []TemplComponent) {
	ctx := context.Background()
	visitor := NewCodeGeneratorVisitor()

	for _, component := range components {
		if err := AcceptVisitor(ctx, component, visitor); err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
	}

	fmt.Println(visitor.GetOutput())
}

// Example: Collecting metrics
func ExampleMetricsCollection(components []TemplComponent) {
	ctx := context.Background()
	visitor := NewMetricsVisitor()

	for _, component := range components {
		AcceptVisitor(ctx, component, visitor)
	}

	metrics := visitor.GetMetrics()
	fmt.Printf("Metrics: %+v\n", metrics)
}

// Example: Accessibility audit
func ExampleAccessibilityAudit(components []TemplComponent) {
	ctx := context.Background()
	visitor := NewAccessibilityAuditVisitor()

	for _, component := range components {
		AcceptVisitor(ctx, component, visitor)
	}

	issues := visitor.GetIssues()
	if len(issues) > 0 {
		fmt.Println("Accessibility issues found:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}
}

// Example: Using middleware
func ExampleMiddlewareUsage(b *Bridge) {
	ctx := context.Background()

	middleware := NewMiddlewareChain()
	middleware.Use(AddDefaultClassesMiddleware)
	middleware.Use(SecurityAttributesMiddleware)

	component := TemplComponent{
		Type:  "Button",
		Props: atoms.ButtonProps{Text: "Click me"},
	}

	transformed, err := middleware.Execute(ctx, component)
	if err != nil {
		fmt.Printf("Middleware error: %v\n", err)
		return
	}

	fmt.Printf("Transformed: %+v\n", transformed)
}

// ============================================================================
// STEP 7: UPDATE BRIDGE CONSTRUCTOR
// ============================================================================

// Update NewBridge to include validation
func NewBridgeWithValidation() *Bridge {
	bridge := NewBridge()
	// Validation is created on-demand in validation methods
	return bridge
}
