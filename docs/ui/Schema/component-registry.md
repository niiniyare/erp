# Component Registry System

**Part of**: Schema-Driven UI Framework Technical White Paper  
**Version**: 2.0  
**Focus**: Component Registration, Resolution, and Rendering

---

## Overview

The Component Registry serves as the central nervous system of the framework, providing the mapping layer between JSON schemas and their corresponding Templ component implementations. It enables runtime component resolution, type-safe rendering, and extensible component architecture.

## Core Registry Architecture

### Registry Interface

```go
type ComponentRegistry interface {
    // Component registration
    Register(schemaType string, factory ComponentFactory) error
    RegisterBatch(registrations map[string]ComponentFactory) error
    
    // Component resolution
    Resolve(schemaType string) (ComponentFactory, error)
    ResolveWithFallback(schemaType string) ComponentFactory
    
    // Component creation and rendering
    CreateComponent(schema *Schema) (TemplComponent, error)
    RenderComponent(ctx context.Context, schema *Schema) (templ.Component, error)
    
    // Registry management
    List() []string
    Exists(schemaType string) bool
    Remove(schemaType string) error
    Clear()
    
    // Introspection
    GetSupportedTypes() []string
    GetComponentInfo(schemaType string) (*ComponentInfo, error)
}
```

### Component Factory Interface

```go
type ComponentFactory interface {
    // Component creation
    CreateComponent(schema *Schema) (TemplComponent, error)
    
    // Schema validation
    ValidateSchema(schema *Schema) error
    GetSchemaConstraints() *SchemaConstraints
    
    // Component metadata
    GetComponentInfo() *ComponentInfo
    GetSupportedSchemaTypes() []string
    GetDefaultProps() ComponentProps
    
    // Lifecycle hooks
    OnRegister(registry ComponentRegistry) error
    OnUnregister(registry ComponentRegistry) error
}
```

### Templ Component Interface

```go
type TemplComponent interface {
    // Core rendering
    Render(ctx context.Context, props ComponentProps) templ.Component
    
    // Props handling
    SetProps(props ComponentProps) error
    GetProps() ComponentProps
    ValidateProps(props ComponentProps) error
    GetDefaultProps() ComponentProps
    
    // Component metadata
    GetType() string
    GetVersion() string
    GetDependencies() []string
    
    // Lifecycle management
    OnMount(ctx context.Context) error
    OnUpdate(ctx context.Context, oldProps ComponentProps) error
    OnUnmount(ctx context.Context) error
    
    // State management
    GetState() ComponentState
    SetState(state ComponentState) error
    
    // Event handling
    HandleEvent(ctx context.Context, event *ComponentEvent) error
}
```

## Registry Implementation

### Thread-Safe Registry

```go
type DefaultComponentRegistry struct {
    components   map[string]ComponentFactory
    fallback     ComponentFactory
    mu           sync.RWMutex
    logger       *Logger
    validator    *SchemaValidator
    eventBus     *EventBus
}

func NewComponentRegistry(options ...RegistryOption) *DefaultComponentRegistry {
    registry := &DefaultComponentRegistry{
        components: make(map[string]ComponentFactory),
        logger:     getDefaultLogger(),
        validator:  getDefaultValidator(),
        eventBus:   getDefaultEventBus(),
    }
    
    for _, option := range options {
        option(registry)
    }
    
    return registry
}

func (r *DefaultComponentRegistry) Register(
    schemaType string,
    factory ComponentFactory,
) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if schemaType == "" {
        return errors.New("schema type cannot be empty")
    }
    
    if factory == nil {
        return errors.New("factory cannot be nil")
    }
    
    // Check if already registered
    if _, exists := r.components[schemaType]; exists {
        return fmt.Errorf("component type %s already registered", schemaType)
    }
    
    // Validate factory
    if err := r.validateFactory(factory); err != nil {
        return fmt.Errorf("factory validation failed: %w", err)
    }
    
    // Register component
    r.components[schemaType] = factory
    
    // Call lifecycle hook
    if err := factory.OnRegister(r); err != nil {
        // Rollback registration
        delete(r.components, schemaType)
        return fmt.Errorf("factory registration hook failed: %w", err)
    }
    
    // Emit registration event
    r.eventBus.Emit(&ComponentRegisteredEvent{
        SchemaType: schemaType,
        Factory:    factory,
    })
    
    r.logger.Info("Component registered",
        "schema_type", schemaType,
        "factory", factory.GetComponentInfo().Name)
    
    return nil
}

func (r *DefaultComponentRegistry) Resolve(
    schemaType string,
) (ComponentFactory, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    factory, exists := r.components[schemaType]
    if !exists {
        return nil, fmt.Errorf("component type %s not registered", schemaType)
    }
    
    return factory, nil
}

func (r *DefaultComponentRegistry) ResolveWithFallback(
    schemaType string,
) ComponentFactory {
    if factory, err := r.Resolve(schemaType); err == nil {
        return factory
    }
    
    if r.fallback != nil {
        r.logger.Warn("Using fallback component",
            "requested_type", schemaType)
        return r.fallback
    }
    
    // Return error component as last resort
    return &ErrorComponentFactory{
        ErrorMessage: fmt.Sprintf("Unknown component type: %s", schemaType),
    }
}
```

### Component Creation and Rendering

```go
func (r *DefaultComponentRegistry) CreateComponent(
    schema *Schema,
) (TemplComponent, error) {
    factory := r.ResolveWithFallback(schema.Type)
    
    // Validate schema against factory constraints
    if err := factory.ValidateSchema(schema); err != nil {
        return nil, fmt.Errorf("schema validation failed: %w", err)
    }
    
    // Create component
    component, err := factory.CreateComponent(schema)
    if err != nil {
        return nil, fmt.Errorf("component creation failed: %w", err)
    }
    
    // Convert schema to props
    props, err := r.schemaToProps(schema)
    if err != nil {
        return nil, fmt.Errorf("props conversion failed: %w", err)
    }
    
    // Set props
    if err := component.SetProps(props); err != nil {
        return nil, fmt.Errorf("props setting failed: %w", err)
    }
    
    return component, nil
}

func (r *DefaultComponentRegistry) RenderComponent(
    ctx context.Context,
    schema *Schema,
) (templ.Component, error) {
    component, err := r.CreateComponent(schema)
    if err != nil {
        return nil, err
    }
    
    // Call mount lifecycle
    if err := component.OnMount(ctx); err != nil {
        r.logger.Error("Component mount failed",
            "error", err,
            "schema_type", schema.Type)
        // Continue with rendering despite mount failure
    }
    
    return component.Render(ctx, component.GetProps()), nil
}
```

## Component Factory Implementations

### Base Component Factory

```go
type BaseComponentFactory struct {
    componentInfo    *ComponentInfo
    schemaValidator  *SchemaValidator
    propsValidator   *PropsValidator
    defaultProps     ComponentProps
}

func (f *BaseComponentFactory) GetComponentInfo() *ComponentInfo {
    return f.componentInfo
}

func (f *BaseComponentFactory) GetSupportedSchemaTypes() []string {
    return f.componentInfo.SupportedTypes
}

func (f *BaseComponentFactory) GetDefaultProps() ComponentProps {
    return f.defaultProps
}

func (f *BaseComponentFactory) ValidateSchema(schema *Schema) error {
    return f.schemaValidator.Validate(schema)
}

func (f *BaseComponentFactory) OnRegister(registry ComponentRegistry) error {
    // Default implementation - no action needed
    return nil
}

func (f *BaseComponentFactory) OnUnregister(registry ComponentRegistry) error {
    // Default implementation - no action needed
    return nil
}
```

### Button Component Factory

```go
type ButtonComponentFactory struct {
    *BaseComponentFactory
}

func NewButtonComponentFactory() *ButtonComponentFactory {
    return &ButtonComponentFactory{
        BaseComponentFactory: &BaseComponentFactory{
            componentInfo: &ComponentInfo{
                Name:           "Button",
                Description:    "Interactive button component",
                Version:        "1.0.0",
                SupportedTypes: []string{"button", "submit", "reset"},
                Category:       "control",
                Tags:           []string{"interactive", "form", "action"},
            },
            schemaValidator: newButtonSchemaValidator(),
            propsValidator:  newButtonPropsValidator(),
            defaultProps: ComponentProps{
                "type":     "button",
                "variant":  "primary",
                "size":     "medium",
                "disabled": false,
            },
        },
    }
}

func (f *ButtonComponentFactory) CreateComponent(
    schema *Schema,
) (TemplComponent, error) {
    return &ButtonComponent{
        schema:       schema,
        defaultProps: f.GetDefaultProps(),
    }, nil
}

type ButtonComponent struct {
    schema       *Schema
    props        ComponentProps
    defaultProps ComponentProps
    state        ComponentState
}

func (c *ButtonComponent) Render(
    ctx context.Context,
    props ComponentProps,
) templ.Component {
    return ButtonTemplate(ButtonTemplateProps{
        Type:      props.GetString("type", "button"),
        Variant:   props.GetString("variant", "primary"),
        Size:      props.GetString("size", "medium"),
        Disabled:  props.GetBool("disabled", false),
        Text:      props.GetString("text", "Button"),
        ID:        props.GetString("id", ""),
        Class:     props.GetString("class", ""),
        OnClick:   props.GetString("onClick", ""),
        AriaLabel: props.GetString("ariaLabel", ""),
    })
}

templ ButtonTemplate(props ButtonTemplateProps) {
    <button
        type={ props.Type }
        class={ getButtonClasses(props) }
        if props.ID != "" {
            id={ props.ID }
        }
        if props.Disabled {
            disabled
        }
        if props.OnClick != "" {
            onclick={ props.OnClick }
        }
        if props.AriaLabel != "" {
            aria-label={ props.AriaLabel }
        }
    >
        { props.Text }
    </button>
}
```

### Form Component Factory

```go
type FormComponentFactory struct {
    *BaseComponentFactory
    fieldRegistry *FieldRegistry
}

func NewFormComponentFactory(fieldRegistry *FieldRegistry) *FormComponentFactory {
    return &FormComponentFactory{
        BaseComponentFactory: &BaseComponentFactory{
            componentInfo: &ComponentInfo{
                Name:           "Form",
                Description:    "Form container with validation",
                Version:        "1.0.0",
                SupportedTypes: []string{"form"},
                Category:       "container",
                Dependencies:   []string{"field-registry", "validator"},
            },
        },
        fieldRegistry: fieldRegistry,
    }
}

func (f *FormComponentFactory) CreateComponent(
    schema *Schema,
) (TemplComponent, error) {
    // Validate form schema
    if err := f.validateFormSchema(schema); err != nil {
        return nil, err
    }
    
    // Create field components
    fields, err := f.createFieldComponents(schema.Fields)
    if err != nil {
        return nil, fmt.Errorf("field creation failed: %w", err)
    }
    
    // Create action components
    actions, err := f.createActionComponents(schema.Actions)
    if err != nil {
        return nil, fmt.Errorf("action creation failed: %w", err)
    }
    
    return &FormComponent{
        schema:  schema,
        fields:  fields,
        actions: actions,
        validator: newFormValidator(schema.ValidationRules),
    }, nil
}
```

## Advanced Registry Features

### Component Composition

```go
type CompositeComponentFactory struct {
    *BaseComponentFactory
    childFactories map[string]ComponentFactory
    registry       ComponentRegistry
}

func (f *CompositeComponentFactory) CreateComponent(
    schema *Schema,
) (TemplComponent, error) {
    component := &CompositeComponent{
        schema:    schema,
        children:  make([]TemplComponent, 0),
        registry:  f.registry,
    }
    
    // Create child components
    for _, childSchema := range schema.Children {
        childComponent, err := f.registry.CreateComponent(childSchema)
        if err != nil {
            return nil, fmt.Errorf("child component creation failed: %w", err)
        }
        component.children = append(component.children, childComponent)
    }
    
    return component, nil
}

type CompositeComponent struct {
    schema   *Schema
    children []TemplComponent
    registry ComponentRegistry
}

func (c *CompositeComponent) Render(
    ctx context.Context,
    props ComponentProps,
) templ.Component {
    return CompositeTemplate(c.children, props)
}

templ CompositeTemplate(children []TemplComponent, props ComponentProps) {
    <div class={ getCompositeClasses(props) }>
        for _, child := range children {
            @child.Render(ctx, props)
        }
    </div>
}
```

### Higher-Order Components

```go
type HOCFactory struct {
    baseFactory ComponentFactory
    enhancers   []ComponentEnhancer
}

type ComponentEnhancer interface {
    Enhance(component TemplComponent) TemplComponent
}

func (f *HOCFactory) CreateComponent(
    schema *Schema,
) (TemplComponent, error) {
    // Create base component
    component, err := f.baseFactory.CreateComponent(schema)
    if err != nil {
        return nil, err
    }
    
    // Apply enhancers
    for _, enhancer := range f.enhancers {
        component = enhancer.Enhance(component)
    }
    
    return component, nil
}

// Error boundary enhancer
type ErrorBoundaryEnhancer struct {
    fallbackComponent TemplComponent
    errorReporter     ErrorReporter
}

func (e *ErrorBoundaryEnhancer) Enhance(
    component TemplComponent,
) TemplComponent {
    return &ErrorBoundaryComponent{
        wrapped:           component,
        fallback:          e.fallbackComponent,
        errorReporter:     e.errorReporter,
    }
}

type ErrorBoundaryComponent struct {
    wrapped       TemplComponent
    fallback      TemplComponent
    errorReporter ErrorReporter
}

func (c *ErrorBoundaryComponent) Render(
    ctx context.Context,
    props ComponentProps,
) templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        defer func() {
            if r := recover(); r != nil {
                c.errorReporter.Report(ctx, fmt.Errorf("component panic: %v", r))
                c.fallback.Render(ctx, props).Render(ctx, w)
            }
        }()
        
        return c.wrapped.Render(ctx, props).Render(ctx, w)
    })
}
```

### Lazy Loading Components

```go
type LazyComponentFactory struct {
    loaderFunc    func() (ComponentFactory, error)
    loadOnce      sync.Once
    loadedFactory ComponentFactory
    loadError     error
}

func NewLazyComponentFactory(
    loader func() (ComponentFactory, error),
) *LazyComponentFactory {
    return &LazyComponentFactory{
        loaderFunc: loader,
    }
}

func (f *LazyComponentFactory) CreateComponent(
    schema *Schema,
) (TemplComponent, error) {
    factory, err := f.getFactory()
    if err != nil {
        return nil, fmt.Errorf("lazy loading failed: %w", err)
    }
    
    return factory.CreateComponent(schema)
}

func (f *LazyComponentFactory) getFactory() (ComponentFactory, error) {
    f.loadOnce.Do(func() {
        f.loadedFactory, f.loadError = f.loaderFunc()
    })
    
    return f.loadedFactory, f.loadError
}
```

### Plugin System

```go
type PluginRegistry struct {
    plugins     map[string]Plugin
    hooks       map[string][]PluginHook
    mu          sync.RWMutex
}

type Plugin interface {
    GetInfo() *PluginInfo
    Initialize(registry ComponentRegistry) error
    RegisterComponents(registry ComponentRegistry) error
    Shutdown() error
}

type PluginHook interface {
    Execute(ctx context.Context, data interface{}) error
}

func (r *PluginRegistry) LoadPlugin(plugin Plugin) error {
    info := plugin.GetInfo()
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Check dependencies
    if err := r.checkDependencies(info.Dependencies); err != nil {
        return fmt.Errorf("dependency check failed: %w", err)
    }
    
    // Initialize plugin
    if err := plugin.Initialize(r.componentRegistry); err != nil {
        return fmt.Errorf("plugin initialization failed: %w", err)
    }
    
    // Register plugin components
    if err := plugin.RegisterComponents(r.componentRegistry); err != nil {
        return fmt.Errorf("component registration failed: %w", err)
    }
    
    r.plugins[info.Name] = plugin
    
    return nil
}
```

## Performance Optimizations

### Component Caching

```go
type CachedComponentFactory struct {
    baseFactory ComponentFactory
    cache       *ComponentCache
    hasher      ComponentHasher
}

func (f *CachedComponentFactory) CreateComponent(
    schema *Schema,
) (TemplComponent, error) {
    // Generate cache key
    key, err := f.hasher.Hash(schema)
    if err != nil {
        return f.baseFactory.CreateComponent(schema)
    }
    
    // Check cache
    if component, found := f.cache.Get(key); found {
        return component, nil
    }
    
    // Create component
    component, err := f.baseFactory.CreateComponent(schema)
    if err != nil {
        return nil, err
    }
    
    // Cache component
    f.cache.Set(key, component)
    
    return component, nil
}
```

### Component Pooling

```go
type ComponentPool struct {
    factory ComponentFactory
    pool    sync.Pool
}

func NewComponentPool(factory ComponentFactory) *ComponentPool {
    return &ComponentPool{
        factory: factory,
        pool: sync.Pool{
            New: func() interface{} {
                component, _ := factory.CreateComponent(&Schema{})
                return component
            },
        },
    }
}

func (p *ComponentPool) Get(schema *Schema) (TemplComponent, error) {
    component := p.pool.Get().(TemplComponent)
    
    // Reset and configure component
    props, err := schemaToProps(schema)
    if err != nil {
        return nil, err
    }
    
    if err := component.SetProps(props); err != nil {
        return nil, err
    }
    
    return component, nil
}

func (p *ComponentPool) Put(component TemplComponent) {
    // Reset component state
    component.SetProps(component.GetDefaultProps())
    p.pool.Put(component)
}
```

---

This component registry system provides a robust, extensible foundation for mapping schemas to components while maintaining type safety, performance, and modularity.