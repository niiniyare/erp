# Schema Runtime

## Overview

The **Runtime** is the execution engine that brings schemas to life in your application. It handles rendering, state management, validation, and user interactions at execution time.

### Purpose

- **Render** schemas into actual UI components
- **Manage** form/component state during user interaction
- **Validate** data in real-time as users input values
- **Handle** events (change, blur, submit, etc.)
- **Execute** conditional logic and dynamic behavior

### Runtime Flow

```
Enriched Schema → Runtime → Live UI → User Interaction → State Updates → Re-validation
```

---

## Core Components

### 1. Renderer
Converts schema definitions into UI components.

### 2. State Manager
Tracks form values, touched fields, and dirty state.

### 3. Runtime Validator
Performs real-time validation as users interact.

### 4. Event Handler
Responds to user actions (click, change, submit).

### 5. Conditional Renderer
Shows/hides fields based on runtime conditions.

---

## Basic Runtime Setup

### Initialize Runtime

```go
package runtime

import (
    "context"
    "sync"
)

// Runtime manages the execution of a schema
type Runtime struct {
    schema    *schema.Schema
    state     *State
    validator *RuntimeValidator
    renderer  *Renderer
    events    *EventBus
    mu        sync.RWMutex
}

func NewRuntime(schema *Schema) *Runtime {
    return &Runtime{
        schema:    schema,
        state:     NewState(schema),
        validator: NewRuntimeValidator(schema),
        renderer:  NewRenderer(schema),
        events:    NewEventBus(),
    }
}

// Initialize prepares the runtime with initial data
func (r *Runtime) Initialize(ctx context.Context, initialData map[string]interface{}) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Set initial values
    if err := r.state.Initialize(initialData); err != nil {
        return err
    }
    
    // Run initial validation
    if err := r.validator.ValidateAll(ctx, r.state.GetAll()); err != nil {
        return err
    }
    
    // Evaluate initial conditional rules
    if err := r.applyConditionalRules(ctx); err != nil {
        return err
    }
    
    return nil
}
```

### Handler Example

```go
func HandleFormRender(c *fiber.Ctx) error {
    // 1. Get enriched schema (from enricher middleware)
    schema := c.Locals("enriched_schema").(*Schema)
    
    // 2. Initialize runtime
    rt := runtime.NewRuntime(schema)
    
    // 3. Load existing data if editing
    existingData := loadFormData(c.Context(), c.Params("id"))
    if err := rt.Initialize(c.Context(), existingData); err != nil {
        return err
    }
    
    // 4. Render the form
    return views.FormPage(schema, rt.GetState()).Render(
        c.Context(),
        c.Response().BodyWriter(),
    )
}
```

---

## State Management

### State Structure

```go
// State holds runtime form state
type State struct {
    values   map[string]interface{} // Current field values
    touched  map[string]bool        // Fields user has interacted with
    dirty    map[string]bool        // Fields that changed from initial
    errors   map[string][]string    // Validation errors per field
    initial  map[string]interface{} // Initial values for dirty checking
    mu       sync.RWMutex
}

func NewState(schema *Schema) *State {
    return &State{
        values:  make(map[string]interface{}),
        touched: make(map[string]bool),
        dirty:   make(map[string]bool),
        errors:  make(map[string][]string),
        initial: make(map[string]interface{}),
    }
}

// Initialize sets initial state
func (s *State) Initialize(data map[string]interface{}) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    for key, value := range data {
        s.values[key] = value
        s.initial[key] = value // Store for dirty checking
    }
    
    return nil
}

// SetValue updates a field value
func (s *State) SetValue(path string, value interface{}) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.values[path] = value
    
    // Mark as dirty if changed from initial
    if s.initial[path] != value {
        s.dirty[path] = true
    } else {
        s.dirty[path] = false
    }
    
    return nil
}

// GetValue retrieves a field value
func (s *State) GetValue(path string) (interface{}, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    value, exists := s.values[path]
    return value, exists
}

// Touch marks a field as touched (user interacted with it)
func (s *State) Touch(path string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.touched[path] = true
}

// IsDirty checks if field changed from initial value
func (s *State) IsDirty(path string) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    return s.dirty[path]
}

// GetErrors returns validation errors for a field
func (s *State) GetErrors(path string) []string {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    return s.errors[path]
}

// SetErrors sets validation errors for a field
func (s *State) SetErrors(path string, errors []string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.errors[path] = errors
}

// IsValid checks if entire form is valid
func (s *State) IsValid() bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    return len(s.errors) == 0
}

// GetAll returns all current values
func (s *State) GetAll() map[string]interface{} {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    // Return a copy to prevent external modification
    result := make(map[string]interface{}, len(s.values))
    for k, v := range s.values {
        result[k] = v
    }
    
    return result
}

// Reset clears all state back to initial
func (s *State) Reset() {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.values = make(map[string]interface{})
    s.touched = make(map[string]bool)
    s.dirty = make(map[string]bool)
    s.errors = make(map[string][]string)
    
    // Restore initial values
    for k, v := range s.initial {
        s.values[k] = v
    }
}
```

### State Subscription (Optional)

```go
// StateObserver gets notified of state changes
type StateObserver func(path string, oldValue, newValue interface{})

// Subscribe to state changes
func (s *State) Subscribe(observer StateObserver) func() {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Add observer to list
    // Return unsubscribe function
}
```

---

## Runtime Validation

### Validation Strategies

```go
type ValidationStrategy string

const (
    ValidateEager    ValidationStrategy = "eager"    // Validate on every change
    ValidateLazy     ValidationStrategy = "lazy"     // Validate on blur
    ValidateOnSubmit ValidationStrategy = "onSubmit" // Validate only on submit
)

// RuntimeValidator validates during user interaction
type RuntimeValidator struct {
    schema   *Schema
    strategy ValidationStrategy
    debounce time.Duration
}

func NewRuntimeValidator(schema *Schema) *RuntimeValidator {
    return &RuntimeValidator{
        schema:   schema,
        strategy: ValidateLazy, // Default: validate on blur
        debounce: 300 * time.Millisecond,
    }
}

// ValidateField validates a single field
func (v *RuntimeValidator) ValidateField(
    ctx context.Context,
    fieldName string,
    value interface{},
) []string {
    field := v.schema.FindField(fieldName)
    if field == nil {
        return nil
    }
    
    var errors []string
    
    // Check required
    if field.Required && isEmpty(value) {
        errors = append(errors, fmt.Sprintf("%s is required", field.Label))
    }
    
    // Check type
    if !isValidType(value, field.Type) {
        errors = append(errors, fmt.Sprintf("%s must be a %s", field.Label, field.Type))
    }
    
    // Check min/max length
    if field.MinLength != nil {
        if length := getLength(value); length < *field.MinLength {
            errors = append(errors, fmt.Sprintf("%s must be at least %d characters", field.Label, *field.MinLength))
        }
    }
    
    if field.MaxLength != nil {
        if length := getLength(value); length > *field.MaxLength {
            errors = append(errors, fmt.Sprintf("%s must be at most %d characters", field.Label, *field.MaxLength))
        }
    }
    
    // Check pattern
    if field.Pattern != "" {
        matched, _ := regexp.MatchString(field.Pattern, fmt.Sprintf("%v", value))
        if !matched {
            errors = append(errors, fmt.Sprintf("%s format is invalid", field.Label))
        }
    }
    
    // Custom validations
    for _, rule := range field.Validations {
        if err := v.validateRule(ctx, rule, value); err != nil {
            errors = append(errors, err.Error())
        }
    }
    
    return errors
}

// ValidateAll validates entire form
func (v *RuntimeValidator) ValidateAll(
    ctx context.Context,
    data map[string]interface{},
) map[string][]string {
    allErrors := make(map[string][]string)
    
    for _, field := range v.schema.Fields {
        value := data[field.Name]
        
        if errors := v.ValidateField(ctx, field.Name, value); len(errors) > 0 {
            allErrors[field.Name] = errors
        }
    }
    
    return allErrors
}

// ValidateDebounced validates after a delay (for performance)
func (v *RuntimeValidator) ValidateDebounced(
    ctx context.Context,
    fieldName string,
    value interface{},
) <-chan []string {
    result := make(chan []string, 1)
    
    go func() {
        time.Sleep(v.debounce)
        errors := v.ValidateField(ctx, fieldName, value)
        result <- errors
        close(result)
    }()
    
    return result
}
```

---

## Event Handling

### Event System

```go
// EventType represents runtime events
type EventType string

const (
    EventChange EventType = "change"
    EventBlur   EventType = "blur"
    EventFocus  EventType = "focus"
    EventSubmit EventType = "submit"
)

// Event represents a runtime event
type Event struct {
    Type      EventType
    Field     string
    Value     interface{}
    OldValue  interface{}
    Timestamp time.Time
}

// EventHandler handles runtime events
type EventHandler struct {
    runtime   *Runtime
    handlers  map[EventType][]EventCallback
    mu        sync.RWMutex
}

type EventCallback func(ctx context.Context, event *Event) error

// OnChange handles field value changes
func (h *EventHandler) OnChange(ctx context.Context, event *Event) error {
    // 1. Update state
    if err := h.runtime.state.SetValue(event.Field, event.Value); err != nil {
        return err
    }
    
    // 2. Validate based on strategy
    if h.runtime.validator.strategy == ValidateEager {
        errors := h.runtime.validator.ValidateField(ctx, event.Field, event.Value)
        h.runtime.state.SetErrors(event.Field, errors)
    }
    
    // 3. Re-evaluate conditional rules
    if err := h.runtime.applyConditionalRules(ctx); err != nil {
        return err
    }
    
    // 4. Trigger custom handlers
    h.mu.RLock()
    callbacks := h.handlers[EventChange]
    h.mu.RUnlock()
    
    for _, callback := range callbacks {
        if err := callback(ctx, event); err != nil {
            log.Error().Err(err).Msg("Event callback failed")
        }
    }
    
    return nil
}

// OnBlur handles field blur (focus lost)
func (h *EventHandler) OnBlur(ctx context.Context, event *Event) error {
    // Mark field as touched
    h.runtime.state.Touch(event.Field)
    
    // Validate on blur if lazy strategy
    if h.runtime.validator.strategy == ValidateLazy {
        errors := h.runtime.validator.ValidateField(ctx, event.Field, event.Value)
        h.runtime.state.SetErrors(event.Field, errors)
    }
    
    return nil
}

// OnSubmit handles form submission
func (h *EventHandler) OnSubmit(ctx context.Context) error {
    // 1. Validate all fields
    allData := h.runtime.state.GetAll()
    allErrors := h.runtime.validator.ValidateAll(ctx, allData)
    
    // 2. Update state with all errors
    for field, errors := range allErrors {
        h.runtime.state.SetErrors(field, errors)
    }
    
    // 3. Check if valid
    if !h.runtime.state.IsValid() {
        return fmt.Errorf("form validation failed")
    }
    
    // 4. Trigger submit handlers
    event := &Event{
        Type:      EventSubmit,
        Timestamp: time.Now(),
    }
    
    h.mu.RLock()
    callbacks := h.handlers[EventSubmit]
    h.mu.RUnlock()
    
    for _, callback := range callbacks {
        if err := callback(ctx, event); err != nil {
            return err
        }
    }
    
    return nil
}

// Register custom event handler
func (h *EventHandler) Register(eventType EventType, callback EventCallback) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    h.handlers[eventType] = append(h.handlers[eventType], callback)
}
```

---

## Conditional Rendering

### Runtime Conditionals

```go
// ConditionalRenderer handles dynamic field visibility
type ConditionalRenderer struct {
    runtime *Runtime
}

// ApplyConditionalRules evaluates and applies conditional rules
func (c *ConditionalRenderer) ApplyConditionalRules(ctx context.Context) error {
    data := c.runtime.state.GetAll()
    
    for i := range c.runtime.schema.Fields {
        field := &c.runtime.schema.Fields[i]
        
        // Skip if no runtime info
        if field.Runtime == nil || field.Runtime.Condition == "" {
            continue
        }
        
        // Evaluate condition
        result, err := c.evaluateCondition(field.Runtime.Condition, data)
        if err != nil {
            log.Error().
                Err(err).
                Str("field", field.Name).
                Str("condition", field.Runtime.Condition).
                Msg("Failed to evaluate condition")
            continue
        }
        
        // Apply result
        field.Runtime.Visible = result
    }
    
    return nil
}

// evaluateCondition parses and evaluates a condition expression
func (c *ConditionalRenderer) evaluateCondition(
    condition string,
    data map[string]interface{},
) (bool, error) {
    // Simple expression parser (use a library like govaluate in production)
    // Supports: field == "value", field != "value", field > 10, etc.
    
    // Example: "delivery_method == 'ship'"
    parts := strings.Split(condition, "==")
    if len(parts) == 2 {
        fieldName := strings.TrimSpace(parts[0])
        expectedValue := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
        
        actualValue := fmt.Sprintf("%v", data[fieldName])
        return actualValue == expectedValue, nil
    }
    
    // Add more operators as needed (!=, >, <, >=, <=, contains, etc.)
    
    return false, fmt.Errorf("unsupported condition format: %s", condition)
}
```

---

## Complete Handler Example

### API Endpoints

```go
// POST /api/forms/:id/field/:field
func HandleFieldUpdate(c *fiber.Ctx) error {
    // 1. Get schema and initialize runtime
    schema := c.Locals("enriched_schema").(*Schema)
    rt := getOrCreateRuntime(c, schema)
    
    // 2. Parse new value
    var payload struct {
        Value interface{} `json:"value"`
    }
    if err := c.BodyParser(&payload); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
    }
    
    fieldName := c.Params("field")
    
    // 3. Create change event
    oldValue, _ := rt.state.GetValue(fieldName)
    event := &runtime.Event{
        Type:      runtime.EventChange,
        Field:     fieldName,
        Value:     payload.Value,
        OldValue:  oldValue,
        Timestamp: time.Now(),
    }
    
    // 4. Handle the event
    if err := rt.events.OnChange(c.Context(), event); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    // 5. Return updated state
    return c.JSON(fiber.Map{
        "value":  payload.Value,
        "errors": rt.state.GetErrors(fieldName),
        "valid":  len(rt.state.GetErrors(fieldName)) == 0,
    })
}

// POST /api/forms/:id/submit
func HandleFormSubmit(c *fiber.Ctx) error {
    schema := c.Locals("enriched_schema").(*Schema)
    rt := getOrCreateRuntime(c, schema)
    
    // Parse form data
    var formData map[string]interface{}
    if err := c.BodyParser(&formData); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid form data"})
    }
    
    // Update runtime state with all values
    for field, value := range formData {
        rt.state.SetValue(field, value)
    }
    
    // Handle submit
    if err := rt.events.OnSubmit(c.Context()); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error":  "Validation failed",
            "errors": rt.state.errors,
        })
    }
    
    // Save to database
    savedData, err := saveFormData(c.Context(), c.Params("id"), formData)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to save"})
    }
    
    return c.JSON(fiber.Map{
        "success": true,
        "data":    savedData,
    })
}

// Helper to get or create runtime for session
func getOrCreateRuntime(c *fiber.Ctx, schema *Schema) *runtime.Runtime {
    formID := c.Params("id")
    sessionID := c.Cookies("session_id")
    
    key := fmt.Sprintf("runtime:%s:%s", sessionID, formID)
    
    // Try to get from cache/session
    if cached := getFromCache(key); cached != nil {
        return cached.(*runtime.Runtime)
    }
    
    // Create new runtime
    rt := runtime.NewRuntime(schema)
    
    // Load existing data if editing
    if formID != "new" {
        existingData := loadFormData(c.Context(), formID)
        rt.Initialize(c.Context(), existingData)
    }
    
    // Cache for this session
    setInCache(key, rt, 30*time.Minute)
    
    return rt
}
```

---

## Performance Optimization

### Debouncing Validation

```go
// Debounce validation for performance
type DebouncedValidator struct {
    validator *RuntimeValidator
    timers    map[string]*time.Timer
    mu        sync.Mutex
}

func (d *DebouncedValidator) ValidateField(
    ctx context.Context,
    fieldName string,
    value interface{},
    callback func(errors []string),
) {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    // Cancel existing timer
    if timer, exists := d.timers[fieldName]; exists {
        timer.Stop()
    }
    
    // Create new timer
    d.timers[fieldName] = time.AfterFunc(300*time.Millisecond, func() {
        errors := d.validator.ValidateField(ctx, fieldName, value)
        callback(errors)
    })
}
```

### Memoization

```go
// Cache validation results for unchanged values
type CachedValidator struct {
    validator *RuntimeValidator
    cache     map[string]cachedResult
    mu        sync.RWMutex
}

type cachedResult struct {
    value  interface{}
    errors []string
}

func (c *CachedValidator) ValidateField(
    ctx context.Context,
    fieldName string,
    value interface{},
) []string {
    // Check cache
    c.mu.RLock()
    cached, exists := c.cache[fieldName]
    c.mu.RUnlock()
    
    if exists && cached.value == value {
        return cached.errors // Return cached result
    }
    
    // Validate
    errors := c.validator.ValidateField(ctx, fieldName, value)
    
    // Update cache
    c.mu.Lock()
    c.cache[fieldName] = cachedResult{
        value:  value,
        errors: errors,
    }
    c.mu.Unlock()
    
    return errors
}
```

---

## Error Handling

```go
// RuntimeError represents runtime-specific errors
type RuntimeError struct {
    Type    string
    Field   string
    Message string
    Err     error
}

func (e *RuntimeError) Error() string {
    return fmt.Sprintf("runtime error [%s] on field '%s': %s", e.Type, e.Field, e.Message)
}

// Error types
const (
    ErrTypeValidation   = "validation"
    ErrTypeState        = "state"
    ErrTypeConditional  = "conditional"
    ErrTypeEvent        = "event"
)

// HandleError provides user-friendly error messages
func (r *Runtime) HandleError(err error) map[string]interface{} {
    if runtimeErr, ok := err.(*RuntimeError); ok {
        return map[string]interface{}{
            "error": map[string]interface{}{
                "type":    runtimeErr.Type,
                "field":   runtimeErr.Field,
                "message": runtimeErr.Message,
            },
        }
    }
    
    return map[string]interface{}{
        "error": map[string]interface{}{
            "type":    "unknown",
            "message": "An unexpected error occurred",
        },
    }
}
```

---

## Best Practices

### ✅ DO:

- **Validate on appropriate events** based on UX needs (blur for better UX than on-every-keystroke)
- **Debounce expensive validations** (API calls, complex regex)
- **Cache runtime instances** per session to maintain state across requests
- **Use optimistic UI updates** for better perceived performance
- **Clear state** when user navigates away or completes action
- **Log runtime errors** for debugging and monitoring

### ❌ DON'T:

- **Keep runtime state on client only** - always sync critical state to server
- **Block UI** during validation - use async validation
- **Validate on every keystroke** for complex rules - debounce instead
- **Store sensitive data** in client-side runtime state
- **Forget to clean up** - clear old runtime instances from cache

---

## Frontend Integration Example

### React Hook

```typescript
// useFormRuntime.ts
function useFormRuntime(schemaId: string) {
  const [state, setState] = useState({});
  const [errors, setErrors] = useState({});
  
  const handleChange = async (field: string, value: any) => {
    // Optimistic update
    setState(prev => ({ ...prev, [field]: value }));
    
    // Call backend
    const response = await fetch(`/api/forms/${schemaId}/field/${field}`, {
      method: 'POST',
      body: JSON.stringify({ value }),
    });
    
    const result = await response.json();
    
    // Update with validation results
    setErrors(prev => ({ ...prev, [field]: result.errors }));
  };
  
  const handleSubmit = async () => {
    const response = await fetch(`/api/forms/${schemaId}/submit`, {
      method: 'POST',
      body: JSON.stringify(state),
    });
    
    if (!response.ok) {
      const result = await response.json();
      setErrors(result.errors);
      return false;
    }
    
    return true;
  };
  
  return { state, errors, handleChange, handleSubmit };
}
```

---

## Summary

The runtime brings schemas to life by:
- **Rendering** schemas into interactive UI
- **Managing** state throughout user interaction
- **Validating** data in real-time with configurable strategies
- **Handling** events and executing business logic
- **Applying** conditional rules dynamically

The runtime sits between your enriched schemas and the actual user interface, providing a robust execution layer that handles all the complexity of user interaction, validation, and state management.

---

[← Back: Extended Components](./14-extended_components.md) | [Next: Complete Example →](16-example.md)
