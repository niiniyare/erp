## 6. 🧩 Component Architecture

### 6.1 Component Philosophy

Components are the building blocks of the UI. Well-designed components are:

1. **Single Responsibility** — Do one thing well
2. **Composable** — Combine into complex interfaces
3. **Predictable** — Consistent behavior
4. **Accessible** — Work for everyone
5. **Documented** — Clear usage examples

### 6.2 Component Hierarchy

```
┌───────────────────────────────────────┐
│         Atomic Components             │
│  (Button, Input, Icon, Badge)         │
└─────────────┬─────────────────────────┘
              │
              ▼
┌───────────────────────────────────────┐
│      Molecular Components             │
│  (FormField, SearchBox, MenuItem)     │
└─────────────┬─────────────────────────┘
              │
              ▼
┌───────────────────────────────────────┐
│      Organism Components              │
│  (Form, Table, Navigation, Modal)     │
└─────────────┬─────────────────────────┘
              │
              ▼
┌───────────────────────────────────────┐
│         Template Components           │
│  (PageLayout, DashboardLayout)        │
└───────────────────────────────────────┘
```

### 6.3 Component Anatomy

Every component should have these elements:

```typescript
interface Component {
  // Visual
  variants: Variant[];        // Different visual styles
  sizes: Size[];              // Size options
  states: State[];            // Interactive states
  
  // Behavior
  props: ComponentProps;      // Configuration
  events: EventHandler[];     // User interactions
  validation: Validator[];    // Input validation
  
  // Accessibility
  ariaLabel: string;
  ariaDescribedBy?: string;
  role?: string;
  tabIndex?: number;
  
  // Documentation
  description: string;
  examples: Example[];
  guidelines: Guideline[];
}
```

### 6.4 Variant System

Components should have semantic variants:

```css
/* Button variants */
.button-primary {
  background: var(--primary);
  color: var(--primary-foreground);
}

.button-secondary {
  background: var(--secondary);
  color: var(--secondary-foreground);
}

.button-destructive {
  background: var(--destructive);
  color: var(--destructive-foreground);
}

.button-outline {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--foreground);
}

.button-ghost {
  background: transparent;
  color: var(--foreground);
}

.button-link {
  background: transparent;
  color: var(--primary);
  text-decoration: underline;
}
```

### 6.5 Size System

Consistent sizing across components:

```css
/* Size scale */
.component-xs  { /* Extra small */
  height: var(--size-6);   /* 24px */
  padding: var(--space-1) var(--space-2);
  font-size: var(--font-size-xs);
}

.component-sm  { /* Small */
  height: var(--size-8);   /* 32px */
  padding: var(--space-1-5) var(--space-3);
  font-size: var(--font-size-sm);
}

.component-md  { /* Medium (default) */
  height: var(--size-10);  /* 40px */
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-base);
}

.component-lg  { /* Large */
  height: var(--size-12);  /* 48px */
  padding: var(--space-3) var(--space-6);
  font-size: var(--font-size-lg);
}

.component-xl  { /* Extra large */
  height: var(--size-14);  /* 56px */
  padding: var(--space-4) var(--space-8);
  font-size: var(--font-size-xl);
}
```

### 6.6 State System

Visual feedback for interaction states:

```css
/* Default state */
.component-default {
  background: var(--background);
  color: var(--foreground);
  border: 1px solid var(--border);
}

/* Hover state */
.component:hover {
  background: var(--background-emphasis);
  border-color: var(--border-strong);
}

/* Focus state */
.component:focus-visible {
  outline: 2px solid var(--ring);
  outline-offset: 2px;
}

/* Active state */
.component:active {
  background: var(--interactive-active);
}

/* Disabled state */
.component:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

/* Error state */
.component-error {
  border-color: var(--destructive);
}

/* Success state */
.component-success {
  border-color: var(--success);
}
```

### 6.7 Composition Patterns

#### **Slot-based Composition**

```typescript
// Button with icon
<Button>
  <Icon slot="icon-left" name="check" />
  <span>Save Changes</span>
  <Icon slot="icon-right" name="arrow-right" />
</Button>

// Card composition
<Card>
  <CardHeader slot="header">
    <CardTitle>User Profile</CardTitle>
    <CardDescription>Manage your account settings</CardDescription>
  </CardHeader>
  <CardContent slot="content">
    <!-- Main content -->
  </CardContent>
  <CardFooter slot="footer">
    <Button>Save</Button>
  </CardFooter>
</Card>
```

#### **Compound Components**

```typescript
// Form field composition
<FormField>
  <FormLabel>Email Address</FormLabel>
  <FormControl>
    <Input type="email" />
  </FormControl>
  <FormDescription>We'll never share your email.</FormDescription>
  <FormMessage>Please enter a valid email</FormMessage>
</FormField>

// Select composition
<Select>
  <SelectTrigger>
    <SelectValue placeholder="Choose an option" />
  </SelectTrigger>
  <SelectContent>
    <SelectItem value="1">Option 1</SelectItem>
    <SelectItem value="2">Option 2</SelectItem>
  </SelectContent>
</Select>
```

### 6.8 Component API Design

Consistent prop naming across components:

| Prop Category | Props | Example |
|---------------|-------|---------|
| **Visual** | `variant`, `size`, `color` | `variant="primary"` |
| **State** | `disabled`, `loading`, `error` | `disabled={true}` |
| **Content** | `label`, `placeholder`, `value` | `placeholder="Enter text"` |
| **Behavior** | `onClick`, `onChange`, `onSubmit` | `onClick={handleClick}` |
| **Accessibility** | `aria-label`, `aria-describedby`, `role` | `aria-label="Close"` |
| **Styling** | `className`, `style` | `className="custom"` |

### 6.9 Default Values

All components should have sensible defaults:

```typescript
// Good: Sensible defaults
<Button>             // variant="primary", size="md"
<Input />            // type="text", size="md"
<Card />             // variant="default", padding="md"

// User only specifies what's different
<Button variant="outline" size="sm">
<Input type="email" error={true}>
<Card padding="lg">
```

### 6.10 Component Best Practices

**Do:**
- ✅ Follow single responsibility principle
- ✅ Use semantic HTML elements
- ✅ Provide keyboard navigation
- ✅ Include ARIA attributes
- ✅ Support all states (hover, focus, disabled)
- ✅ Use design tokens for styling
- ✅ Document props and usage

**Don't:**
- ❌ Create overly complex components
- ❌ Hardcode colors or spacing
- ❌ Forget focus indicators
- ❌ Ignore disabled states
- ❌ Use divs for everything
- ❌ Skip documentation
- ❌ Create components for one-time use

---

