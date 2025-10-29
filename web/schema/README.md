# Awo ERP Schema System Architecture

**Version:** 1.0.0  
**Last Updated:** October 2025  
**Status:** Active Development  
**Framework Integration:** Fiber + templ + HTMX + Alpine.js  

---

## 🏗️ Architecture Overview

The Awo ERP Schema System implements a **declarative, type-safe, JSON-driven UI architecture** that enables developers to build complex enterprise forms and interfaces through configuration rather than code. This system bridges the gap between traditional backend schemas and modern frontend component libraries.

### Core Philosophy

1. **Schema-First Design**: UI components are generated from JSON schemas rather than hand-coded
2. **Type Safety Throughout**: Compile-time validation from Go structs to templ templates
3. **Declarative Configuration**: Complex UI behaviors defined through structured data
4. **Performance by Design**: Zero JavaScript bundle overhead with server-side rendering
5. **Enterprise Ready**: Built-in multi-tenancy, security, workflow, and validation

### Design Principles

- **Separation of Concerns**: Schema definitions are separate from implementation logic
- **Composability**: Complex forms built from reusable schema components  
- **Extensibility**: Easy to add new field types and UI patterns
- **Maintainability**: JSON schemas can be version controlled and migrated
- **Developer Experience**: IntelliSense and type checking for schema authoring

---

## 📁 Package Structure

```
@web/schema/
├── README.md              # This file - architecture overview
├── schema.go              # Convenience aggregation package
│
├── core/                  # Foundation types and interfaces
│   ├── base.go           # Base types, common interfaces
│   ├── validation.go     # Validation rules and constraints
│   ├── events.go         # Event system definitions
│   ├── context.go        # Context and metadata handling
│   └── errors.go         # Schema-specific error types
│
├── form/                  # Form schema definitions
│   ├── form.go           # Main FormSchema struct
│   ├── fields.go         # Field types and configurations
│   ├── layout.go         # Form layout and responsive design
│   ├── workflow.go       # Workflow and approval schemas
│   ├── validation.go     # Form-specific validation rules
│   └── presets.go        # Pre-built form templates
│
├── ui/                    # UI component schemas
│   ├── components.go     # Component property definitions
│   ├── theme.go          # Theme and styling schemas
│   ├── responsive.go     # Responsive design configurations
│   ├── animation.go      # Animation and transition schemas
│   └── typography.go     # Typography and spacing schemas
│
├── data/                  # Data handling schemas
│   ├── sources.go        # Data source configurations
│   ├── transformers.go   # Data transformation schemas
│   ├── cache.go          # Caching strategy schemas
│   ├── tenant.go         # Multi-tenant data isolation
│   └── async.go          # Async operations and validation
│
├── integration/           # External system integration
│   ├── htmx.go           # HTMX attribute configurations
│   ├── alpine.go         # Alpine.js directive schemas
│   ├── api.go            # REST API integration schemas
│   ├── webhooks.go       # Webhook configuration schemas
│   └── events.go         # External event system integration
│
├── builder/               # Schema construction utilities
│   ├── factory.go        # Schema factory patterns
│   ├── registry.go       # Schema registration and discovery
│   ├── migration.go      # Schema version migration
│   ├── validation.go     # Schema structure validation
│   └── generator.go      # Code generation utilities
│
└── docs/                  # Documentation and examples
    ├── examples/         # Complete schema examples
    ├── patterns/         # Common design patterns
    ├── migration/        # Migration guides
    └── api/              # API documentation
```

---

## 🔄 Data Flow Architecture

```
JSON Schema Definition
       ↓
Go Schema Structs (Type Safety)
       ↓
Schema Validation & Processing
       ↓
templ Template Generation
       ↓
HTMX/Alpine.js Enhancement
       ↓
Server-Side Rendering
       ↓
Interactive HTML Response
```

### Key Decision Points

1. **Why Go Structs Over Pure JSON?**
   - Compile-time validation prevents runtime errors
   - IDE support with autocompletion and refactoring
   - Strong typing prevents configuration mistakes
   - Easy integration with existing Go codebase

2. **Why templ Over Alternative Templating?**
   - Type-safe template compilation
   - Component composition patterns
   - Performance optimizations out of the box
   - Seamless Go integration

3. **Why HTMX + Alpine.js Over SPA Frameworks?**
   - Server-side rendering for better performance
   - Reduced JavaScript bundle size (near zero)
   - Simplified state management
   - Better SEO and accessibility

---

## 🏛️ Component Type Hierarchy

### Foundation Layer (core/)
```
BaseSchema ──┐
             ├── ValidationRule
             ├── EventHandler  
             ├── MetadataContainer
             └── ContextProvider
```

### Form Layer (form/)
```
FormSchema ──┬── FieldSchema ──┬── InputField
             │                 ├── SelectField
             │                 ├── DateField
             │                 └── CustomField
             │
             ├── LayoutSchema ──┬── GridLayout
             │                  ├── FlexLayout
             │                  └── TabLayout
             │
             └── WorkflowSchema ─┬── ApprovalChain
                                 ├── NotificationRule
                                 └── StateTransition
```

### UI Layer (ui/)
```
ComponentSchema ──┬── ButtonSchema
                  ├── InputSchema
                  ├── CardSchema
                  └── NavigationSchema

ThemeSchema ──┬── ColorPalette
              ├── Typography
              └── Spacing
```

### Data Layer (data/)
```
DataSourceSchema ──┬── APISource
                   ├── StaticSource
                   └── DatabaseSource

TransformSchema ──┬── FilterTransform
                  ├── MapTransform
                  └── ValidationTransform
```

### Integration Layer (integration/)
```
HTMXSchema ──┬── RequestConfig
             ├── ResponseConfig
             └── TriggerConfig

AlpineSchema ──┬── StateConfig
               ├── DirectiveConfig
               └── EventConfig
```

---

## 🔗 Dependency Mapping

### Type Dependencies

**Core Dependencies** (No external dependencies):
- `BaseSchema`: Foundation interface
- `ValidationRule`: Standalone validation logic
- `EventHandler`: Event system foundation

**Form Dependencies**:
- `FormSchema` → `core.BaseSchema`, `ui.ComponentSchema`
- `FieldSchema` → `core.ValidationRule`, `data.DataSourceSchema`
- `WorkflowSchema` → `core.EventHandler`, `integration.HTMXSchema`

**UI Dependencies**:
- `ComponentSchema` → `core.BaseSchema`, `ui.ThemeSchema`
- `ThemeSchema` → `core.MetadataContainer`
- `ResponsiveSchema` → `ui.ComponentSchema`

**Data Dependencies**:
- `DataSourceSchema` → `core.BaseSchema`, `integration.APISchema`
- `TransformSchema` → `data.DataSourceSchema`
- `TenantSchema` → `core.ContextProvider`

**Integration Dependencies**:
- `HTMXSchema` → `core.EventHandler`, `data.DataSourceSchema`
- `AlpineSchema` → `ui.ComponentSchema`, `core.ValidationRule`

### Import Hierarchy
```
Level 1: core/        (No internal dependencies)
Level 2: ui/, data/   (Depends on core)
Level 3: form/        (Depends on core, ui, data)
Level 4: integration/ (Depends on all above)
Level 5: builder/     (Depends on all above)
```

---

## ✅ Development Roadmap

### Phase A: Foundation Setup
- [ ] **A1**: Create core package structure and base types
- [ ] **A2**: Implement validation system with rule engine
- [ ] **A3**: Design event system architecture
- [ ] **A4**: Create context and metadata handling
- [ ] **A5**: Implement error handling system

### Phase B: Core Schema Types
- [ ] **B1**: Define FormSchema with basic field support
- [ ] **B2**: Implement FieldSchema with 15+ field types
- [ ] **B3**: Create layout system (grid, flex, tabs)
- [ ] **B4**: Design workflow and approval schemas
- [ ] **B5**: Build validation rule combinations

### Phase C: UI Component Integration
- [ ] **C1**: Map existing templ components to schemas
- [ ] **C2**: Create theme and styling system
- [ ] **C3**: Implement responsive design schemas
- [ ] **C4**: Add animation and transition support
- [ ] **C5**: Build component composition patterns

### Phase D: Data Layer
- [ ] **D1**: Implement data source abstraction
- [ ] **D2**: Create transformation pipeline
- [ ] **D3**: Add caching strategies
- [ ] **D4**: Build multi-tenant data isolation
- [ ] **D5**: Implement async validation system

### Phase E: Framework Integration
- [ ] **E1**: HTMX attribute mapping and generation
- [ ] **E2**: Alpine.js directive integration
- [ ] **E3**: API endpoint schema definitions
- [ ] **E4**: Webhook configuration system
- [ ] **E5**: Event system integration

### Phase F: Developer Experience
- [ ] **F1**: Schema factory and builder patterns
- [ ] **F2**: Registry system for schema discovery
- [ ] **F3**: Migration tools for schema evolution
- [ ] **F4**: Validation tools for schema correctness
- [ ] **F5**: Code generation utilities

### Phase G: templ Integration
- [ ] **G1**: Create schema-to-templ renderer
- [ ] **G2**: Implement component factory system
- [ ] **G3**: Build form generation pipeline
- [ ] **G4**: Add runtime schema processing
- [ ] **G5**: Optimize rendering performance

### Phase H: JSON-to-UI Pipeline
- [ ] **H1**: JSON schema parser and validator
- [ ] **H2**: Runtime schema loading system
- [ ] **H3**: Dynamic component instantiation
- [ ] **H4**: Live schema editing capabilities
- [ ] **H5**: Visual schema builder foundation

### Phase I: Production Features
- [ ] **I1**: Security validation and sanitization
- [ ] **I2**: Performance monitoring and optimization
- [ ] **I3**: Error handling and recovery
- [ ] **I4**: Logging and debugging tools
- [ ] **I5**: Documentation and examples

### Phase J: Enterprise Features
- [ ] **J1**: Advanced workflow engine integration
- [ ] **J2**: Multi-tenant schema isolation
- [ ] **J3**: Role-based schema access control
- [ ] **J4**: Audit logging for schema changes
- [ ] **J5**: Enterprise deployment patterns

---

## 🎯 Success Metrics

### Technical Goals
- [ ] **Zero JavaScript Bundle**: Achieve <5KB total JS (Alpine.js only)
- [ ] **Type Safety**: 100% compile-time validation of schemas
- [ ] **Performance**: <50ms form rendering time
- [ ] **Coverage**: Support for 25+ field types and 15+ layouts

### Developer Experience Goals  
- [ ] **Learning Curve**: New developer productive in <2 hours
- [ ] **Documentation**: 100% API coverage with examples
- [ ] **Tooling**: VS Code extension with schema IntelliSense
- [ ] **Migration**: Zero-downtime schema updates

### Business Goals
- [ ] **Productivity**: 5x faster form development vs hand-coding
- [ ] **Consistency**: Standardized UI patterns across modules
- [ ] **Maintainability**: Non-technical users can edit form configs
- [ ] **Flexibility**: Support for custom business rule validation

---

## 🤝 Contributing

This schema system follows strict architectural principles:

1. **No Breaking Changes**: New features must be backward compatible
2. **Type Safety First**: All new types must include validation
3. **Documentation Required**: Every public type needs examples
4. **Performance Conscious**: No performance regressions allowed
5. **Test Coverage**: 85%+ coverage for all new code

Start with Phase A tasks and work sequentially through the roadmap. Each phase builds on the previous phase's foundations.

---

**Next Steps**: Begin with `✅ Phase A: Foundation Setup` and check off tasks as you complete them. The architecture is designed to support incremental development while maintaining system integrity.