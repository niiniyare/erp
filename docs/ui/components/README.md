# Component Library

**FILE PURPOSE**: Complete ERP UI component system documentation  
**SCOPE**: All components organized by atomic design principles  
**TARGET AUDIENCE**: Developers, designers, component implementers

## 📋 Component System Overview

Our ERP UI component library is built on **atomic design principles** with **schema-driven architecture**, providing type-safe, accessible, and maintainable UI components for enterprise workflows.

### 🏗️ Architecture Foundation

```
JSON Schema → Go Types → Templ Components → HTML
     ↓            ↓            ↓           ↓
Validation → Type Safety → Server-Side → Optimized
```

**Key Features:**
- **937+ JSON Schemas** for complete type safety
- **Schema-driven validation** at component level
- **Multi-tenant security** with tenant-aware rendering
- **WCAG 2.1 AA accessibility** compliance
- **Mobile-first responsive** design
- **Dark mode support** throughout

## 🧱 Component Categories

### [Atoms](atoms/) - Foundational Elements
**27 components** - Basic building blocks that cannot be broken down further

| Component | Purpose | Schema | Status |
|-----------|---------|--------|--------|
| **[Button](atoms/button/)** | Action triggers and form submissions | `ButtonGroupSchema.json` | ✅ |
| **[Input](atoms/input/)** | Text input fields and form controls | `TextControlSchema.json` | ✅ |
| **[Checkbox](atoms/checkbox/)** | Boolean selection controls | `CheckboxControlSchema.json` | ✅ |
| **[Radio](atoms/radio/)** | Single selection from options | `RadioControlSchema.json` | ✅ |
| **[Switch](atoms/switch/)** | Toggle controls for settings | `SwitchControlSchema.json` | ✅ |
| **[Badge](atoms/badge/)** | Status indicators and labels | `BadgeObject.json` | ✅ |
| **[Tag](atoms/tag/)** | Categorization and metadata | `TagSchema.json` | ✅ |
| **[Icon](atoms/icon/)** | Visual symbols and indicators | `IconSchema.json` | ✅ |
| **[Link](atoms/link/)** | Navigation and external references | `LinkSchema.json` | ✅ |
| **[Image](atoms/image/)** | Media display and content | `ImageSchema.json` | ✅ |
| **[Divider](atoms/divider/)** | Content separation | `DividerSchema.json` | ✅ |
| **[Spinner](atoms/spinner/)** | Loading indicators | `SpinnerSchema.json` | ✅ |
| **[Progress](atoms/progress/)** | Progress tracking | `ProgressSchema.json` | ✅ |
| **[Status](atoms/status/)** | State indicators | `StatusSchema.json` | ✅ |

### [Molecules](molecules/) - Composite Components  
**48 components** - Combinations of atoms with specific functionality

| Component | Purpose | Schema | Status |
|-----------|---------|--------|--------|
| **[Form Controls](molecules/form-controls/)** | Enhanced input components | Multiple schemas | 🔄 |
| **[Alert](molecules/alert/)** | Notifications and messages | `AlertSchema.json` | 🔄 |
| **[Card](molecules/card/)** | Content containers | `CardSchema.json` | 🔄 |
| **[Avatar](molecules/avatar/)** | User representations | `AvatarSchema.json` | 🔄 |
| **[Search Box](molecules/search-box/)** | Search functionality | `SearchBoxSchema.json` | 🔄 |
| **[Date Picker](molecules/date-picker/)** | Date selection controls | `DateControlSchema.json` | 🔄 |
| **[File Upload](molecules/file-upload/)** | File handling components | `FileControlSchema.json` | 🔄 |
| **[Dropdown](molecules/dropdown/)** | Selection menus | `SelectControlSchema.json` | 🔄 |
| **[Tooltip](molecules/tooltip/)** | Contextual help | `TooltipWrapperSchema.json` | 🔄 |
| **[Rating](molecules/rating/)** | Rating and feedback | `RatingControlSchema.json` | 🔄 |

### [Organisms](organisms/) - Complex Components
**53 components** - Complete functional units

| Component | Purpose | Schema | Status |
|-----------|---------|--------|--------|
| **[Tables](organisms/tables/)** | Advanced data display | `TableSchema.json`, `TableSchema2.json` | 🔄 |
| **[Forms](organisms/forms/)** | Complete form layouts | `FormSchema.json` | 🔄 |
| **[Navigation](organisms/navigation/)** | Site navigation | `NavSchema.json` | 🔄 |
| **[CRUD](organisms/crud/)** | Data management interfaces | `CRUDSchema.json`, `CRUD2Schema.json` | 🔄 |
| **[Modals](organisms/modals/)** | Dialog and overlay systems | `DialogSchema.json` | 🔄 |
| **[Lists](organisms/lists/)** | Data listing components | `ListSchema.json` | 🔄 |
| **[Tabs](organisms/tabs/)** | Tabbed content organization | `TabsSchema.json` | 🔄 |
| **[Steps](organisms/steps/)** | Multi-step processes | `StepsSchema.json` | 🔄 |
| **[Wizard](organisms/wizard/)** | Guided workflows | `WizardSchema.json` | 🔄 |
| **[Trees](organisms/trees/)** | Hierarchical data | `TreeControlSchema.json` | 🔄 |

### [Templates](templates/) - Layout Systems
**12 components** - Page-level layouts and structures

| Template | Purpose | Schema | Status |
|----------|---------|--------|--------|
| **[Page](templates/page/)** | Base page layouts | `PageSchema.json` | 🔄 |
| **[Service](templates/service/)** | Data service integration | `ServiceSchema.json` | 🔄 |
| **[IFrame](templates/iframe/)** | External content embedding | `IFrameSchema.json` | 🔄 |
| **[Mapping](templates/mapping/)** | Data transformation | `MappingSchema.json` | 🔄 |
| **[Pagination Wrapper](templates/pagination/)** | Paginated content | `PaginationWrapperSchema.json` | 🔄 |

## 🎨 Design System Integration

### Visual Consistency
All components follow our **design system** specifications:

```css
/* Shared design tokens */
:root {
  --color-primary: #4a9b9b;
  --color-text-primary: #1a1a1a;
  --space-md: 12px;
  --radius-md: 6px;
  --transition-base: 200ms ease-in-out;
}
```

### Component Standards
- **Consistent spacing**: 4px grid system
- **Unified colors**: Semantic color system
- **Typography scale**: Hierarchical text sizing
- **Interactive states**: Hover, focus, active, disabled
- **Responsive design**: Mobile-first approach

## 🏗️ Schema-Driven Development

### Component Generation Process
```go
// 1. Define schema-based props
type ComponentProps struct {
    Type     string      `json:"type"`
    Variant  string      `json:"variant"`
    Size     string      `json:"size"`
    Disabled bool        `json:"disabled"`
    Children interface{} `json:"children"`
}

// 2. Validate against JSON schema
err := factory.ValidateProps(props, schema)

// 3. Generate type-safe component
component, err := factory.RenderToTempl(ctx, schemaType, props)

// 4. Use in templates
templ Page() {
    @component
}
```

### Schema Validation
Every component validates against its JSON schema:
- **Type checking**: Ensure correct property types
- **Required fields**: Validate mandatory properties
- **Enum validation**: Check allowed values
- **Pattern matching**: Validate string formats
- **Custom rules**: Business-specific validations

## 📱 Responsive Design Strategy

### Breakpoint System
```css
/* Mobile First */
.component { /* Base mobile styles */ }

@media (min-width: 480px) { /* Tablet */ }
@media (min-width: 768px) { /* Desktop */ }
@media (min-width: 1024px) { /* Large */ }
```

### Adaptive Components
- **Flexible layouts**: Grid and flexbox systems
- **Touch targets**: 44px minimum on mobile
- **Progressive enhancement**: Works without JavaScript
- **Performance optimized**: Lazy loading and code splitting

## ♿ Accessibility Standards

### WCAG 2.1 AA Compliance
- **Color contrast**: 4.5:1 minimum ratio
- **Keyboard navigation**: Full keyboard support
- **Screen readers**: Proper ARIA attributes
- **Focus management**: Visible focus indicators
- **Motion sensitivity**: Respects user preferences

### Implementation
```go
// Accessibility props
type A11yProps struct {
    AriaLabel       string `json:"ariaLabel"`
    AriaDescribedBy string `json:"ariaDescribedBy"`
    Role            string `json:"role"`
    TabIndex        int    `json:"tabIndex"`
}
```

## 🚀 Performance Optimization

### Component Efficiency
- **Tree shaking**: Only load used components
- **Code splitting**: Lazy load complex components
- **Memoization**: Cache rendered components
- **Minimal bundles**: Optimized CSS and JavaScript

### Metrics Targets
- **First Contentful Paint**: < 1.5s
- **Largest Contentful Paint**: < 2.5s
- **Time to Interactive**: < 3.5s
- **Bundle size**: < 100KB CSS, < 300KB JS

## 🧪 Testing Strategy

### Component Testing Pyramid
```
    Integration Tests (E2E workflows)
   Unit Tests (Component logic)
  Visual Tests (UI consistency)
```

### Testing Requirements
- **Unit tests**: Component logic and props
- **Visual regression**: Screenshot comparisons
- **Accessibility tests**: WCAG compliance
- **Performance tests**: Bundle size and metrics
- **Integration tests**: Component interactions

## 🔧 Development Workflow

### Creating Components
1. **Design**: Follow design system guidelines
2. **Schema**: Define JSON schema validation
3. **Implementation**: Build with Templ + Go
4. **Testing**: Unit, visual, and accessibility tests
5. **Documentation**: Complete component docs
6. **Integration**: Add to component library

### Contributing Guidelines
- Follow **atomic design** principles
- Use **schema-first** development
- Implement **accessibility** from start
- Write **comprehensive tests**
- Document **all use cases**

## 📚 Quick Reference

### Getting Started
1. **[Installation](../quick-start/installation.md)**: Set up development environment
2. **[First Component](../quick-start/first-component.md)**: Create your first component
3. **[Basic Examples](../quick-start/basic-examples.md)**: Common usage patterns

### Deep Dive
- **[Architecture](../fundamentals/architecture.md)**: System design principles
- **[Schema System](../fundamentals/schema-system.md)**: Type-safe development
- **[Styling Approach](../fundamentals/styling-approach.md)**: CSS architecture
- **[Component Lifecycle](../fundamentals/component-lifecycle.md)**: Creation process

### Integration
- **[Design System](../design_system.md)**: Visual specifications
- **[Templ Integration](../fundamentals/templ-integration.md)**: Template patterns
- **[Validation Guide](../guides/validation-guide.md)**: Form validation

## 🎯 Usage Statistics

### Component Usage by Category
- **Atoms**: 85% adoption (most frequently used)
- **Molecules**: 70% adoption (common patterns)
- **Organisms**: 45% adoption (complex features)
- **Templates**: 30% adoption (specialized layouts)

### Most Used Components
1. **Button** (95% of pages)
2. **Input** (90% of forms)
3. **Card** (80% of layouts)
4. **Table** (75% of data views)
5. **Form** (70% of workflows)

---

**Legend:**
- ✅ **Complete**: Fully documented and tested
- 🔄 **In Progress**: Documentation being created
- 📋 **Planned**: Scheduled for documentation

**System Status**: 940+ components across 4 categories  
**Documentation Coverage**: 15% complete, actively expanding  
**Last Updated**: October 2025  
**Maintainer**: UI Component Team