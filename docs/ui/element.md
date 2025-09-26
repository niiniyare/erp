# Templ UI Component Library Documentation

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Core Elements](#core-elements)
4. [Element Reference](#element-reference)
5. [Composing Complex Components](#composing-complex-components)
6. [Dependencies & External Libraries](#dependencies--external-libraries)
7. [Best Practices](#best-practices)
8. [Future Publishing Guidance](#future-publishing-guidance)
9. [References](#references)

---

## 1. Introduction

### Purpose of this Documentation

This documentation serves as a comprehensive reference for foundational UI element components built with **Templ.guide**, enhanced by **Flowbite** styling, and optionally extended with **Alpine.js** interactivity. These atomic elements form the building blocks for more complex, configurable components tailored to specific feature requirements.

### Technologies Used

- **[Templ.guide](https://templ.guide)**: A templating language for Go that compiles to type-safe Go code
- **[Flowbite](https://flowbite.com)**: A UI component library based on Tailwind CSS providing ready-to-use HTML components
- **[Alpine.js](https://alpinejs.dev)**: Optional lightweight JavaScript framework for adding reactivity and interactivity

### How to Use This Documentation

Each element is documented with:
- **Function signature** (when necessary for clarity)
- **Attribute tables** with detailed specifications
- **Conceptual usage walkthroughs** instead of full code implementations
- **Dependency requirements** and setup instructions
- **Composition guidelines** for building complex components

---

## 2. Getting Started

### Installation Requirements

#### Core Dependencies

```bash
# Install Templ CLI
go install github.com/a-h/templ/cmd/templ@latest

# Initialize Go module (if not already done)
go mod init your-project

# Add Templ dependency
go get github.com/a-h/templ
```

#### Project Setup with Templ.guide

1. **Create your project structure:**
   ```
   project/
   ├── components/
   │   ├── elements/     # Foundational elements
   │   └── composed/     # Complex components
   ├── static/
   │   ├── css/
   │   └── js/
   └── main.go
   ```

2. **Generate Go code from templates:**
   ```bash
   templ generate
   ```

3. **Watch for changes during development:**
   ```bash
   templ generate --watch
   ```

#### Flowbite Setup

**Option 1: CDN (Recommended for development)**
```html
<!-- CSS -->
<link href="https://cdn.jsdelivr.net/npm/flowbite@2.5.1/dist/flowbite.min.css" rel="stylesheet" />

<!-- JavaScript -->
<script src="https://cdn.jsdelivr.net/npm/flowbite@2.5.1/dist/flowbite.min.js"></script>
```

**Option 2: NPM Installation**
```bash
npm install flowbite
```

#### Optional Alpine.js Setup

**CDN Integration:**
```html
<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
```

**NPM Installation:**
```bash
npm install alpinejs
```

---

## 3. Core Elements (Templ Functions)

### Element Philosophy

Foundational elements follow these principles:

1. **Atomic Design**: Each element serves a single, well-defined purpose
2. **Composability**: Elements can be combined to create complex components
3. **Accessibility**: Built-in ARIA attributes and semantic HTML
4. **Theming**: Consistent with Flowbite's design system
5. **Progressive Enhancement**: Core functionality works without JavaScript

### Naming Conventions

- **Element names**: Use descriptive names that reflect the UI element (e.g., `Button`, `Input`, `Card`)
- **Attributes**: Follow camelCase for Go parameters, kebab-case for HTML attributes
- **Variants**: Use suffixes to indicate variations (e.g., `ButtonPrimary`, `ButtonSecondary`)

### Attribute Structure and Best Practices

| Attribute Category | Description | Example |
|-------------------|-------------|---------|
| **Core** | Essential functionality attributes | `id`, `class`, `disabled` |
| **Content** | Text and data attributes | `text`, `placeholder`, `value` |
| **Styling** | Visual appearance modifiers | `variant`, `size`, `color` |
| **Behavior** | Interaction and state attributes | `onClick`, `onSubmit`, `loading` |
| **Accessibility** | ARIA and semantic attributes | `ariaLabel`, `role`, `tabIndex` |

---

## 4. Element Reference

### 4.1 Button Element

#### Description
A flexible button component that supports multiple variants, sizes, and states, based on Flowbite's button system.

#### Function Signature
```go
templ Button(props ButtonProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `text` | `string` | `""` | Button text content |
| `variant` | `string` | `"primary"` | Visual style variant (`primary`, `secondary`, `success`, `danger`, `warning`, `info`, `light`, `dark`) |
| `size` | `string` | `"base"` | Button size (`xs`, `sm`, `base`, `lg`, `xl`) |
| `disabled` | `bool` | `false` | Whether the button is disabled |
| `loading` | `bool` | `false` | Shows loading state with spinner |
| `fullWidth` | `bool` | `false` | Makes button full width of container |
| `onClick` | `string` | `""` | JavaScript click handler |
| `type` | `string` | `"button"` | HTML button type (`button`, `submit`, `reset`) |
| `id` | `string` | `""` | HTML id attribute |
| `class` | `string` | `""` | Additional CSS classes |
| `ariaLabel` | `string` | `""` | Accessibility label |

#### Conceptual Usage Walkthrough

**Step 1: Basic Button Creation**
When creating a basic button, you specify the text content and optionally the variant. The button automatically receives Flowbite's styling classes and proper semantic HTML structure.

**Step 2: State Management**
If you enable the `loading` attribute, the button displays a spinner icon and becomes non-interactive. The `disabled` attribute makes the button visually distinct and prevents interaction.

**Step 3: Size and Layout**
The `size` attribute adjusts padding, font size, and overall dimensions. The `fullWidth` attribute makes the button span its container's width, useful for mobile interfaces or form layouts.

**Step 4: Accessibility Integration**
The component automatically includes appropriate ARIA attributes. Custom `ariaLabel` values override the default text-based accessibility label.

#### Dependencies
- **Required**: Flowbite CSS framework
- **Optional**: Alpine.js for enhanced interactivity

#### Notes
- Buttons automatically include focus states and keyboard navigation
- Loading state disables the button to prevent double-submission
- Color variants follow Flowbite's semantic color system

---

### 4.2 Input Element

#### Description
A versatile input field component supporting various input types, validation states, and accessibility features.

#### Function Signature
```go
templ Input(props InputProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `type` | `string` | `"text"` | HTML input type (`text`, `email`, `password`, `number`, `tel`, `url`, `search`) |
| `value` | `string` | `""` | Input field value |
| `placeholder` | `string` | `""` | Placeholder text |
| `name` | `string` | `""` | Form field name |
| `id` | `string` | `""` | HTML id attribute |
| `disabled` | `bool` | `false` | Whether input is disabled |
| `readonly` | `bool` | `false` | Whether input is read-only |
| `required` | `bool` | `false` | Whether input is required |
| `size` | `string` | `"base"` | Input size (`sm`, `base`, `lg`) |
| `validation` | `string` | `"none"` | Validation state (`none`, `success`, `error`) |
| `helperText` | `string` | `""` | Helper or error text below input |
| `label` | `string` | `""` | Associated label text |
| `class` | `string` | `""` | Additional CSS classes |
| `onChange` | `string` | `""` | JavaScript change handler |
| `onFocus` | `string` | `""` | JavaScript focus handler |
| `onBlur` | `string` | `""` | JavaScript blur handler |

#### Conceptual Usage Walkthrough

**Step 1: Basic Input Setup**
Start by defining the input type and basic properties. The component automatically generates appropriate HTML structure with Flowbite styling classes.

**Step 2: Label Association**
When you provide a `label`, the component creates a properly associated label element with the `for` attribute linking to the input's id. This ensures screen readers can correctly identify the input's purpose.

**Step 3: Validation Integration**
The `validation` attribute changes the input's visual state. When set to `error`, the border becomes red and error styling is applied. The `helperText` displays below the input in the appropriate color.

**Step 4: Form Integration**
The `name` attribute enables proper form submission and data collection. Required inputs automatically include HTML5 validation attributes.

#### Dependencies
- **Required**: Flowbite CSS framework
- **Optional**: Alpine.js for real-time validation

---

### 4.3 Card Element

#### Description
A flexible container component for grouping related content with optional headers, footers, and actions.

#### Function Signature
```go
templ Card(props CardProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `title` | `string` | `""` | Card header title |
| `subtitle` | `string` | `""` | Card header subtitle |
| `content` | `templ.Component` | `nil` | Main card content |
| `footer` | `templ.Component` | `nil` | Card footer content |
| `image` | `string` | `""` | Header image URL |
| `imageAlt` | `string` | `""` | Image alt text |
| `bordered` | `bool` | `true` | Whether card has border |
| `shadow` | `string` | `"base"` | Shadow intensity (`none`, `sm`, `base`, `md`, `lg`, `xl`) |
| `padding` | `string` | `"base"` | Internal padding (`none`, `sm`, `base`, `lg`, `xl`) |
| `maxWidth` | `string` | `""` | Maximum width constraint |
| `class` | `string` | `""` | Additional CSS classes |
| `clickable` | `bool` | `false` | Whether entire card is clickable |
| `onClick` | `string` | `""` | JavaScript click handler |

#### Conceptual Usage Walkthrough

**Step 1: Content Structure**
Cards follow a hierarchical content structure: optional image, title/subtitle header, main content area, and optional footer. Each section can be independently controlled.

**Step 2: Visual Styling**
The `shadow` and `bordered` attributes control the card's visual prominence. Higher shadow values create more prominent cards, while removing borders creates a flatter design.

**Step 3: Interactive Behavior**
When `clickable` is enabled, the entire card becomes interactive with hover states and cursor changes. This is useful for navigation cards or selectable items.

**Step 4: Content Composition**
The `content` and `footer` attributes accept other Templ components, enabling complex nested layouts while maintaining the card's structure.

#### Dependencies
- **Required**: Flowbite CSS framework
- **Optional**: Alpine.js for dynamic content updates

---

### 4.4 Alert Element

#### Description
A notification component for displaying important messages with various severity levels and optional actions.

#### Function Signature
```go
templ Alert(props AlertProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `message` | `string` | `""` | Alert message text |
| `variant` | `string` | `"info"` | Alert type (`info`, `success`, `warning`, `error`) |
| `title` | `string` | `""` | Optional alert title |
| `dismissible` | `bool` | `false` | Whether alert can be dismissed |
| `icon` | `bool` | `true` | Whether to show status icon |
| `border` | `bool` | `false` | Whether to show colored border |
| `actions` | `templ.Component` | `nil` | Custom action buttons |
| `class` | `string` | `""` | Additional CSS classes |
| `onDismiss` | `string` | `""` | JavaScript dismiss handler |

#### Conceptual Usage Walkthrough

**Step 1: Message Display**
Alerts automatically apply appropriate color schemes and icons based on the variant. Info alerts use blue styling, success uses green, warning uses yellow, and error uses red.

**Step 2: Dismissible Behavior**
When `dismissible` is enabled, a close button appears. The alert can be hidden via JavaScript or Alpine.js. The dismiss action can trigger custom handlers for cleanup or state management.

**Step 3: Action Integration**
Custom actions can be added through the `actions` component slot. These typically include buttons for related actions like "Retry", "Learn More", or "Settings".

#### Dependencies
- **Required**: Flowbite CSS framework
- **Required**: Flowbite JavaScript (for dismissible functionality)
- **Optional**: Alpine.js for state management

---

### 4.5 Modal Element

#### Description
An overlay component for displaying content above the main interface, with backdrop and focus management.

#### Function Signature
```go
templ Modal(props ModalProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `title` | `string` | `""` | Modal header title |
| `content` | `templ.Component` | `nil` | Modal body content |
| `footer` | `templ.Component` | `nil` | Modal footer content |
| `size` | `string` | `"default"` | Modal size (`xs`, `sm`, `default`, `lg`, `xl`, `2xl`) |
| `position` | `string` | `"center"` | Modal position (`center`, `top`) |
| `backdrop` | `string` | `"default"` | Backdrop behavior (`default`, `static`, `none`) |
| `keyboard` | `bool` | `true` | Whether ESC key closes modal |
| `persistent` | `bool` | `false` | Whether modal prevents closing |
| `id` | `string` | `""` | Modal identifier |
| `class` | `string` | `""` | Additional CSS classes |

#### Conceptual Usage Walkthrough

**Step 1: Content Organization**
Modals structure content into header, body, and footer sections. The header typically contains a title and close button, the body contains the main content, and the footer contains action buttons.

**Step 2: Size and Positioning**
The `size` attribute controls the modal's width and maximum dimensions. The `position` attribute determines whether the modal appears centered or at the top of the viewport.

**Step 3: Interaction Behavior**
Backdrop behavior controls how users can dismiss the modal. Static backdrops prevent clicking outside to close, while the default behavior allows this. Keyboard support enables ESC key dismissal.

**Step 4: Focus Management**
The modal automatically manages focus, trapping keyboard navigation within the modal content and returning focus to the triggering element when closed.

#### Dependencies
- **Required**: Flowbite CSS framework
- **Required**: Flowbite JavaScript (for modal functionality)
- **Optional**: Alpine.js for state management

---

### 4.6 Dropdown Element

#### Description
A toggleable overlay component for displaying a list of actions, links, or other interactive elements.

#### Function Signature
```go
templ Dropdown(props DropdownProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `trigger` | `templ.Component` | `nil` | Element that triggers dropdown |
| `items` | `[]DropdownItem` | `[]` | Dropdown menu items |
| `placement` | `string` | `"bottom"` | Dropdown position (`top`, `bottom`, `left`, `right`) |
| `offset` | `int` | `10` | Distance from trigger element |
| `strategy` | `string` | `"absolute"` | Positioning strategy (`absolute`, `fixed`) |
| `autoClose` | `bool` | `true` | Whether clicking outside closes dropdown |
| `divider` | `bool` | `false` | Whether to show dividers between items |
| `class` | `string` | `""` | Additional CSS classes |

#### Conceptual Usage Walkthrough

**Step 1: Trigger Configuration**
The trigger component (typically a button) activates the dropdown when clicked. The trigger can be any interactive element that accepts click events.

**Step 2: Item Structure**
Dropdown items can be links, buttons, or text elements. Each item can specify its own styling, actions, and behavior. Dividers can separate groups of related items.

**Step 3: Positioning Logic**
The dropdown automatically calculates optimal positioning based on available space. If the preferred placement would cause overflow, the dropdown repositions itself intelligently.

**Step 4: Keyboard Navigation**
Arrow keys navigate between items, Enter activates the focused item, and ESC closes the dropdown. This ensures full keyboard accessibility.

#### Dependencies
- **Required**: Flowbite CSS framework
- **Required**: Flowbite JavaScript (for positioning and interactions)
- **Optional**: Alpine.js for custom state management

---

### 4.7 Navigation Element

#### Description
A responsive navigation component supporting horizontal and vertical layouts with nested menu support.

#### Function Signature
```go
templ Navigation(props NavigationProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `items` | `[]NavItem` | `[]` | Navigation menu items |
| `orientation` | `string` | `"horizontal"` | Layout direction (`horizontal`, `vertical`) |
| `brand` | `templ.Component` | `nil` | Brand/logo component |
| `responsive` | `bool` | `true` | Whether navigation is responsive |
| `sticky` | `bool` | `false` | Whether navigation sticks to top |
| `background` | `string` | `"white"` | Background color variant |
| `border` | `bool` | `true` | Whether to show bottom border |
| `class` | `string` | `""` | Additional CSS classes |

#### Conceptual Usage Walkthrough

**Step 1: Menu Structure**
Navigation items can contain text, links, and nested sub-menus. Each item can be marked as active to indicate the current page or section.

**Step 2: Responsive Behavior**
When responsive mode is enabled, the navigation automatically transforms into a mobile-friendly hamburger menu at smaller screen sizes. The menu can be toggled via JavaScript or Alpine.js.

**Step 3: Brand Integration**
The brand slot typically contains a logo, company name, or home link. It's positioned prominently and maintains visibility across all breakpoints.

#### Dependencies
- **Required**: Flowbite CSS framework
- **Required**: Flowbite JavaScript (for responsive behavior)
- **Optional**: Alpine.js for state management

---

### 4.8 Form Elements

#### 4.8.1 Checkbox Element

#### Function Signature
```go
templ Checkbox(props CheckboxProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `checked` | `bool` | `false` | Whether checkbox is checked |
| `label` | `string` | `""` | Checkbox label text |
| `name` | `string` | `""` | Form field name |
| `value` | `string` | `""` | Checkbox value |
| `disabled` | `bool` | `false` | Whether checkbox is disabled |
| `indeterminate` | `bool` | `false` | Whether checkbox is in indeterminate state |
| `color` | `string` | `"blue"` | Checkbox color theme |
| `size` | `string` | `"base"` | Checkbox size (`sm`, `base`, `lg`) |
| `helperText` | `string` | `""` | Helper text below checkbox |

#### 4.8.2 Radio Element

#### Function Signature
```go
templ Radio(props RadioProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `checked` | `bool` | `false` | Whether radio is selected |
| `label` | `string` | `""` | Radio label text |
| `name` | `string` | `""` | Form field name (group identifier) |
| `value` | `string` | `""` | Radio value |
| `disabled` | `bool` | `false` | Whether radio is disabled |
| `color` | `string` | `"blue"` | Radio color theme |
| `size` | `string` | `"base"` | Radio size (`sm`, `base`, `lg`) |

#### 4.8.3 Select Element

#### Function Signature
```go
templ Select(props SelectProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `options` | `[]SelectOption` | `[]` | Available options |
| `value` | `string` | `""` | Selected value |
| `placeholder` | `string` | `""` | Placeholder text |
| `multiple` | `bool` | `false` | Whether multiple selection is allowed |
| `searchable` | `bool` | `false` | Whether options are searchable |
| `disabled` | `bool` | `false` | Whether select is disabled |
| `size` | `string` | `"base"` | Select size (`sm`, `base`, `lg`) |
| `validation` | `string` | `"none"` | Validation state (`none`, `success`, `error`) |

#### 4.8.4 Textarea Element

#### Function Signature
```go
templ Textarea(props TextareaProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `value` | `string` | `""` | Textarea content |
| `placeholder` | `string` | `""` | Placeholder text |
| `rows` | `int` | `4` | Number of visible rows |
| `resize` | `string` | `"both"` | Resize behavior (`none`, `both`, `horizontal`, `vertical`) |
| `disabled` | `bool` | `false` | Whether textarea is disabled |
| `readonly` | `bool` | `false` | Whether textarea is read-only |
| `maxLength` | `int` | `0` | Maximum character limit |
| `validation` | `string` | `"none"` | Validation state (`none`, `success`, `error`) |

---

### 4.9 Typography Elements

#### 4.9.1 Heading Element

#### Function Signature
```go
templ Heading(props HeadingProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `level` | `int` | `1` | Heading level (1-6) |
| `text` | `string` | `""` | Heading text content |
| `size` | `string` | `"auto"` | Size override (`auto`, `xs`, `sm`, `base`, `lg`, `xl`, `2xl`, `3xl`, `4xl`, `5xl`, `6xl`, `7xl`, `8xl`, `9xl`) |
| `weight` | `string` | `"auto"` | Font weight (`auto`, `thin`, `light`, `normal`, `medium`, `semibold`, `bold`, `extrabold`, `black`) |
| `color` | `string` | `"auto"` | Text color variant |
| `align` | `string` | `"left"` | Text alignment (`left`, `center`, `right`, `justify`) |
| `class` | `string` | `""` | Additional CSS classes |

#### 4.9.2 Paragraph Element

#### Function Signature
```go
templ Paragraph(props ParagraphProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `text` | `string` | `""` | Paragraph text content |
| `size` | `string` | `"base"` | Text size (`xs`, `sm`, `base`, `lg`, `xl`) |
| `weight` | `string` | `"normal"` | Font weight |
| `color` | `string` | `"gray-900"` | Text color |
| `leading` | `string` | `"normal"` | Line height (`tight`, `normal`, `relaxed`, `loose`) |
| `align` | `string` | `"left"` | Text alignment |
| `class` | `string` | `""` | Additional CSS classes |

#### 4.9.3 Link Element

#### Function Signature
```go
templ Link(props LinkProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `href` | `string` | `""` | Link destination URL |
| `text` | `string` | `""` | Link text content |
| `target` | `string` | `""` | Link target (`_blank`, `_self`, etc.) |
| `rel` | `string` | `""` | Link relationship |
| `underline` | `bool` | `true` | Whether link is underlined |
| `color` | `string` | `"blue"` | Link color variant |
| `external` | `bool` | `false` | Whether link is external |
| `class` | `string` | `""` | Additional CSS classes |

---

### 4.10 Layout Elements

#### 4.10.1 Container Element

#### Function Signature
```go
templ Container(props ContainerProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `content` | `templ.Component` | `nil` | Container content |
| `maxWidth` | `string` | `"7xl"` | Maximum width (`sm`, `md`, `lg`, `xl`, `2xl`, `3xl`, `4xl`, `5xl`, `6xl`, `7xl`, `full`) |
| `padding` | `string` | `"base"` | Internal padding (`none`, `sm`, `base`, `lg`, `xl`) |
| `center` | `bool` | `true` | Whether container is centered |
| `class` | `string` | `""` | Additional CSS classes |

#### 4.10.2 Grid Element

#### Function Signature
```go
templ Grid(props GridProps) {}
```

#### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `content` | `templ.Component` | `nil` | Grid content |
| `columns` | `string` | `"1"` | Number of columns (`1`, `2`, `3`, `4`, `5`, `6`, `12`) |
| `gap` | `string` | `"4"` | Grid gap size (`0`, `1`, `2`, `3`, `4`, `5`, `6`, `8`, `10`, `12`) |
| `responsive` | `map[string]string` | `{}` | Responsive column settings |
| `class` | `string` | `""` | Additional CSS classes |

---

## 5. Composing Complex Components

### Composition Philosophy

Complex components are built by combining foundational elements using Templ's composition features. This approach provides:

1. **Reusability**: Foundational elements can be used across multiple complex components
2. **Maintainability**: Changes to foundational elements automatically propagate
3. **Consistency**: Visual and behavioral consistency across the application
4. **Testability**: Each layer can be tested independently

### Composition Patterns

#### 5.1 Container-Content Pattern

Use the Container element as a wrapper for other elements, providing consistent spacing and layout:

**Conceptual Structure:**
```
Container
├── Heading (title)
├── Paragraph (description)
└── Grid
    ├── Card (item 1)
    ├── Card (item 2)
    └── Card (item 3)
```

**Usage Walkthrough:**
1. Start with a Container element to establish maximum width and centering
2. Add a Heading element for the section title
3. Include a Paragraph element for descriptive text
4. Use a Grid element to organize multiple Cards in a responsive layout

#### 5.2 Form Composition Pattern

Combine form elements with validation and layout components:

**Conceptual Structure:**
```
Card
├── Heading (form title)
├── Form Container
│   ├── Input (name field)
│   ├── Input (email field)
│   ├── Textarea (message field)
│   └── Button (submit)
└── Alert (validation messages)
```

**Usage Walkthrough:**
1. Wrap the entire form in a Card for visual grouping
2. Use a Heading element for the form title
3. Group form fields in a logical container
4. Add validation using Alert elements that appear conditionally
5. Include action buttons at the form bottom

#### 5.3 Navigation Layout Pattern

Create complex navigation structures by combining Navigation with Dropdown elements:

**Conceptual Structure:**
```
Navigation
├── Brand Logo
├── Primary Links
│   ├── Link (Home)
│   ├── Dropdown (Products)
│   │   ├── Link (Product A)
│   │   ├── Link (Product B)
│   │   └── Divider
│   └── Link (Contact)
└── User Actions
    └── Dropdown (User Menu)
```

**Usage Walkthrough:**
1. Start with the Navigation element as the container
2. Add brand/logo in the brand slot
3. Create primary navigation links
4. Use Dropdown elements for sub-navigation
5. Include user-specific actions in a separate section

### Attribute Inheritance and Overrides

When composing elements, attributes flow in a predictable manner:

1. **Default Values**: Each element starts with its default attribute values
2. **Component Props**: Values passed to the complex component override defaults
3. **Direct Attributes**: Attributes set directly on elements have highest priority
4. **CSS Classes**: Additional classes are appended, not replaced

### State Management in Composition

For complex components requiring state management:

1. **Alpine.js Integration**: Use `x-data` attributes at the component root
2. **State Sharing**: Pass state down through component props
3. **Event Handling**: Use Alpine.js directives for user interactions
4. **Validation**: Combine form element validation states with Alert components

---

## 6. Dependencies & External Libraries

### 6.1 Flowbite Usage Guidelines

#### CSS Framework Integration

Flowbite provides the foundation for all visual styling. Each element automatically includes appropriate Flowbite classes based on its attributes.

**Color System:**
- Primary: Blue tones for main actions and branding
- Secondary: Gray tones for secondary actions
- Semantic: Green (success), Red (error), Yellow (warning), Blue (info)
- Custom: Defined through Tailwind CSS customization

**Spacing System:**
- Follows Tailwind CSS spacing scale (0, 1, 2, 3, 4, 5, 6, 8, 10, 12, etc.)
- Consistent across padding, margin, and gap properties
- Responsive modifiers for different screen sizes

**Typography Scale:**
- Text sizes: xs, sm, base, lg, xl, 2xl, 3xl, 4xl, 5xl, 6xl, 7xl, 8xl, 9xl
- Font weights: thin, light, normal, medium, semibold, bold, extrabold, black
- Line heights: tight, normal, relaxed, loose

#### JavaScript Components

Some Flowbite components require JavaScript for full functionality:

**Required JavaScript Elements:**
- Modal: Opening, closing, backdrop handling, focus management
- Dropdown: Positioning, click outside detection, keyboard navigation  
- Alert: Dismissible functionality
- Navigation: Mobile menu toggle, responsive behavior

**Integration Example:**
```html
<!-- Include Flowbite JavaScript -->
<script src="https://cdn.jsdelivr.net/npm/flowbite@2.5.1/dist/flowbite.min.js"></script>

<!-- Elements automatically initialize -->
<div data-modal-target="example-modal" data-modal-toggle="example-modal">
  <!-- Modal trigger -->
</div>
```

### 6.2 Alpine.js Optional Enhancements

#### When to Use Alpine.js

Alpine.js enhances components with reactive behavior and state management:

**Recommended Use Cases:**
- Form validation with real-time feedback
- Dynamic content updates without page reloads
- Interactive data filtering and searching
- Complex state management across components
- Custom animations and transitions

**Integration Pattern:**
```html
<!-- Alpine.js state management -->
<div x-data="{ open: false, loading: false }">
  <!-- Templ components with Alpine directives -->
</div>
```

#### State Management Patterns

**Component-Level State:**
- Use `x-data` on the root element of complex components
- Define reactive properties for UI state
- Handle user interactions with `x-on` directives

**Cross-Component Communication:**
- Use Alpine.js stores for global state
- Emit and listen for custom events
- Share state through URL parameters or local storage

### 6.3 Styling Considerations

#### Customization Approach

**Theme Customization:**
1. Override Flowbite CSS variables for color schemes
2. Extend Tailwind CSS configuration for custom spacing/sizing
3. Add custom CSS classes for brand-specific styling
4. Use CSS custom properties for dynamic theming

**Responsive Design:**
- All components include responsive breakpoints by default
- Mobile-first design approach
- Automatic layout adjustments for different screen sizes
- Touch-friendly sizing for interactive elements

#### CSS Architecture

**Class Naming Strategy:**
- Utility-first approach with Tailwind CSS
- Component-specific classes when needed
- BEM methodology for custom components
- Prefix custom classes to avoid conflicts

**CSS Override Hierarchy:**
1. Tailwind CSS base styles
2. Flowbite component styles  
3. Custom utility classes
4. Component-specific styles
5. Inline styles (avoid when possible)

---

## 7. Best Practices

### 7.1 Accessibility Guidelines

#### Semantic HTML Structure

**Heading Hierarchy:**
- Use proper heading levels (h1-h6) to create document outline
- Don't skip heading levels for visual styling
- Include only one h1 per page
- Use the Heading element's `level` attribute for semantic correctness

**Form Accessibility:**
- Associate labels with form controls using proper markup
- Include required field indicators
- Provide clear error messages
- Use fieldsets for grouping related form elements

**Interactive Elements:**
- Ensure all interactive elements are keyboard accessible
- Provide sufficient color contrast (WCAG AA: 4.5:1 for normal text)
- Include focus indicators for keyboard navigation
- Use ARIA attributes when semantic HTML isn't sufficient

#### ARIA Implementation

**Automatic ARIA Attributes:**
- Form elements automatically include `aria-required`, `aria-invalid`
- Buttons include `aria-pressed` for toggle states
- Modal elements include `aria-hidden`, `aria-labelledby`, `aria-describedby`
- Navigation elements include `aria-current` for active states

**Custom ARIA Usage:**
- Use `ariaLabel` attributes for elements without visible text
- Include `aria-describedby` for elements with helper text
- Set `role` attributes for custom interactive elements
- Use `aria-expanded` for collapsible content

#### Screen Reader Support

**Content Structure:**
- Provide meaningful alt text for images
- Use skip links for main content navigation  
- Include landmark regions (`main`, `nav`, `aside`, `footer`)
- Ensure content order matches visual presentation

### 7.2 Performance Considerations

#### Rendering Optimization

**Component Efficiency:**
- Use Templ's `templ.OnceHandle` for scripts and styles that should only render once
- Minimize nested component calls in loops
- Cache expensive computations outside of templates
- Use fragments for partial page updates

**Asset Loading:**
- Load critical CSS inline for above-the-fold content
- Defer non-critical JavaScript to avoid blocking rendering
- Use CDN for external libraries when possible
- Implement lazy loading for images and heavy components

#### Memory Management

**Go Best Practices:**
- Avoid memory leaks in long-running applications
- Use context cancellation for cleanup
- Pool frequently allocated objects when appropriate
- Monitor garbage collection performance

### 7.3 Maintainability and Scalability

#### Code Organization

**File Structure:**
```
components/
├── elements/           # Foundational elements
│   ├── button.templ
│   ├── input.templ
│   └── card.templ
├── composed/          # Complex components
│   ├── forms/
│   ├── navigation/
│   └── layouts/
└── utilities/         # Shared utilities
    ├── types.go       # Shared type definitions
    └── helpers.go     # Helper functions
```

**Component Naming:**
- Use PascalCase for component names
- Include the component type in the name (Button, InputField, NavigationBar)
- Use descriptive names that indicate function
- Avoid abbreviations that might be unclear

#### Versioning Strategy

**API Stability:**
- Maintain backward compatibility for published components
- Use semantic versioning for component library releases
- Document breaking changes clearly
- Provide migration guides for major version updates

**Component Evolution:**
- Add new attributes without breaking existing usage
- Use default values to maintain compatibility
- Deprecate attributes gradually rather than removing immediately
- Provide clear upgrade paths for deprecated features

#### Testing Strategy

**Unit Testing:**
- Test component rendering with various attribute combinations
- Verify HTML output matches expected structure
- Test accessibility attributes are correctly applied
- Validate CSS classes are applied appropriately

**Integration Testing:**
- Test component composition and interaction
- Verify form submission and validation flows
- Test responsive behavior across breakpoints
- Validate JavaScript functionality

**Visual Regression Testing:**
- Capture screenshots of components in different states
- Test across multiple browsers and devices
- Verify color schemes and theming
- Validate animations and transitions

---

## 8. Future Publishing Guidance

### 8.1 Preparing Components for Public Release

#### Documentation Standards

**Component Documentation:**
- Include comprehensive examples for each component
- Document all attributes with types and default values
- Provide migration guides for version updates
- Include accessibility notes and requirements

**Code Quality:**
- Follow Go coding standards and best practices
- Include comprehensive test coverage
- Use consistent naming conventions
- Document public APIs with Go doc comments

#### Package Structure for Publishing

**Go Module Organization:**
```
github.com/your-org/templ-ui/
├── go.mod
├── README.md
├── LICENSE
├── components/
│   ├── button/
│   │   ├── button.templ
│   │   ├── button_test.go
│   │   └── examples_test.go
│   └── card/
│       ├── card.templ
│       └── card_test.go
├── examples/
│   ├── basic/
│   └── advanced/
└── docs/
    ├── getting-started.md
    └── components/
```

**Distribution Considerations:**
- Include pre-compiled Go files in releases
- Provide CDN links for CSS and JavaScript dependencies
- Create installation scripts for common setups
- Include example projects and starter templates

### 8.2 Versioning and Updates

#### Semantic Versioning

**Version Categories:**
- **Major (X.0.0)**: Breaking changes to component APIs
- **Minor (0.X.0)**: New components or backward-compatible features
- **Patch (0.0.X)**: Bug fixes and minor improvements

**Change Documentation:**
- Maintain a detailed CHANGELOG.md
- Include migration instructions for breaking changes
- Document new features and improvements
- List bug fixes and security updates

#### Update Distribution

**Release Process:**
1. Update version numbers in go.mod and package files
2. Generate and test component compilation
3. Update documentation and examples
4. Create release notes and migration guides
5. Publish to Go module registry
6. Update CDN distributions
7. Notify users of available updates

**Backward Compatibility:**
- Maintain API compatibility within major versions
- Provide deprecation warnings before removing features
- Include polyfills or compatibility layers when possible
- Document upgrade paths clearly

---

## 9. References

### Official Documentation Links

#### Templ.guide Resources
- **Main Documentation**: [https://templ.guide/llms.md](https://templ.guide/llms.md)
- **Component Syntax**: Templ components are Go functions that return `templ.Component` interface
- **Template Generation**: Use `templ generate` command to compile templates to Go code
- **HTTP Integration**: Use `templ.Handler()` for HTTP server integration
- **Composition**: Support for component children and parameter passing

#### Flowbite Resources  
- **Component Library**: [https://raw.githubusercontent.com/themesberg/flowbite/refs/heads/main/llms.txt](https://raw.githubusercontent.com/themesberg/flowbite/refs/heads/main/llms.txt)
- **CSS Framework**: Based on Tailwind CSS with ready-to-use components
- **JavaScript Library**: Provides interactive functionality for components
- **Design System**: Comprehensive color, spacing, and typography systems

#### Alpine.js Resources
- **Official Documentation**: [https://alpinejs.dev](https://alpinejs.dev)
- **Getting Started**: Lightweight JavaScript framework for reactive UI
- **Directives**: `x-data`, `x-on`, `x-show`, `x-if`, `x-for` for behavior
- **State Management**: Global stores and component-level state

### Related Technologies

#### Tailwind CSS
- **Utility-First**: CSS framework providing building blocks for custom designs
- **Responsive Design**: Mobile-first breakpoint system
- **Customization**: Extensive theming and configuration options

#### HTMX
- **Hypermedia**: Enables AJAX, CSS Transitions, WebSockets and Server Sent Events
- **Integration**: Works well with Templ for dynamic content updates
- **Progressive Enhancement**: Enhances existing HTML with minimal JavaScript

#### Go Ecosystem
- **Standard Library**: `net/http`, `html/template`, `context` packages
- **Testing**: Built-in testing framework and benchmarking
- **Modules**: Go module system for dependency management

---

### Additional Resources

#### Community and Support
- **GitHub Discussions**: Community support and feature requests
- **Documentation Issues**: Report documentation problems and suggestions
- **Code Examples**: Community-contributed examples and patterns

#### Learning Materials
- **Component Design**: Atomic design principles and component hierarchies
- **Accessibility**: WCAG guidelines and best practices
- **Performance**: Web performance optimization techniques
- **Go Development**: Go programming best practices and patterns

---

*This documentation serves as a comprehensive reference for building UI components with Templ.guide, Flowbite, and Alpine.js. For the most current information, always refer to the official documentation links provided above.*
