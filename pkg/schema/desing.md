## 1. 🎯 Design Philosophy

### 1.1 Core Principles

The Schema Engine Design System is built on five foundational principles that guide every design decision:

#### **Accessibility First**

Every component must meet WCAG 2.1 Level AA standards minimum. Accessibility is not an afterthought—it's the foundation.

```
┌─────────────────────────────────────┐
│  Accessibility Requirements         │
├─────────────────────────────────────┤
│  ✓ Keyboard navigation             │
│  ✓ Screen reader support           │
│  ✓ Color contrast (4.5:1 min)      │
│  ✓ Focus indicators                │
│  ✓ ARIA attributes                 │
│  ✓ Semantic HTML                   │
└─────────────────────────────────────┘
```

**Why:** 15% of the world's population has some form of disability. Accessible design is inclusive design.

#### **Component Composability**

Components should be like LEGO blocks—simple, reusable, and composable into complex interfaces.

```
Simple Components → Composite Components → Complex Layouts

Button           → ButtonGroup         → Navigation Bar
Input            → FormField           → Search Form
Card             → CardList            → Dashboard
```

**Why:** Composable systems scale better, reduce code duplication, and maintain consistency.

#### **Design Token Foundation**

All visual properties derive from a centralized token system. Never hardcode values.

```
❌ Bad:  style={{ color: '#3b82f6', padding: '16px' }}
✅ Good: style={{ color: 'var(--primary)', padding: 'var(--space-4)' }}
```

**Why:** Tokens enable theming, consistency, and maintainability at scale.

#### **Progressive Enhancement**

Start with semantic HTML and enhance with CSS and JavaScript. The system should work without JavaScript for critical functionality.

```
Base Layer (HTML)
    ↓
Enhancement Layer (CSS)
    ↓
Interaction Layer (JavaScript)
```

**Why:** Better performance, SEO, and reliability across devices and network conditions.

#### **Developer Experience**

The system should be intuitive, well-documented, and provide clear error messages.

```
Good DX Checklist:
✓ Clear naming conventions
✓ Comprehensive documentation
✓ TypeScript support
✓ Helpful error messages
✓ Auto-completion
✓ Minimal configuration
```

**Why:** Developer productivity directly impacts product velocity and quality.

### 1.2 Design Values

| Value | Description | Impact |
|-------|-------------|--------|
| **Clarity** | Visual hierarchy should be immediately obvious | Reduces cognitive load |
| **Efficiency** | Minimize clicks, reduce friction | Improves task completion |
| **Consistency** | Patterns repeat predictably | Faster learning curve |
| **Flexibility** | Adapt to diverse use cases | Broader applicability |
| **Beauty** | Aesthetically pleasing interfaces | Higher user satisfaction |

### 1.3 Inspiration Sources

While inspired by shadcn/ui, the Schema Engine Design System draws from multiple sources:

- **shadcn/ui** — Composability, accessibility, customization
- **Tailwind CSS** — Utility-first approach, design constraints
- **Material Design** — Motion principles, elevation system
- **Apple Human Interface** — Clarity, deference, depth
- **IBM Carbon** — Enterprise patterns, data visualization
- **Atlassian Design** — Complex workflow patterns
- **Ant Design** — Comprehensive component library

---

## 2. 🎨 Design Tokens

### 2.1 Token Architecture

Design tokens are the atomic units of the design system. They create a single source of truth for all visual properties.

```
┌──────────────────────────────────────────┐
│         Token Hierarchy                   │
├──────────────────────────────────────────┤
│  Global Tokens (Primitive)               │
│    ↓                                      │
│  Semantic Tokens (Purpose)               │
│    ↓                                      │
│  Component Tokens (Specific)             │
└──────────────────────────────────────────┘
```

#### **Global Tokens (Primitive Values)**

These are the raw values—the building blocks.

```json
{
  "color": {
    "gray": {
      "50": "#fafafa",
      "100": "#f5f5f5",
      "200": "#e5e5e5",
      "300": "#d4d4d4",
      "400": "#a3a3a3",
      "500": "#737373",
      "600": "#525252",
      "700": "#404040",
      "800": "#262626",
      "900": "#171717"
    },
    "blue": {
      "50": "#eff6ff",
      "100": "#dbeafe",
      "500": "#3b82f6",
      "900": "#1e3a8a"
    }
  },
  "spacing": {
    "0": "0",
    "1": "0.25rem",
    "2": "0.5rem",
    "4": "1rem",
    "8": "2rem"
  },
  "font-size": {
    "xs": "0.75rem",
    "sm": "0.875rem",
    "base": "1rem",
    "lg": "1.125rem"
  }
}
```

#### **Semantic Tokens (Purpose-Driven)**

These map global tokens to semantic meaning.

```json
{
  "color": {
    "background": {
      "default": "{color.gray.50}",
      "subtle": "{color.gray.100}",
      "emphasis": "{color.gray.200}"
    },
    "text": {
      "default": "{color.gray.900}",
      "subtle": "{color.gray.600}",
      "disabled": "{color.gray.400}"
    },
    "border": {
      "default": "{color.gray.300}",
      "focus": "{color.blue.500}"
    },
    "feedback": {
      "success": "{color.green.500}",
      "error": "{color.red.500}",
      "warning": "{color.yellow.500}",
      "info": "{color.blue.500}"
    }
  }
}
```

#### **Component Tokens (Specific Use Cases)**

These define how components consume semantic tokens.

```json
{
  "button": {
    "primary": {
      "background": "{color.primary}",
      "text": "{color.primary-foreground}",
      "border": "transparent",
      "hover-background": "{color.primary.dark}",
      "padding-x": "{spacing.4}",
      "padding-y": "{spacing.2}",
      "border-radius": "{radius.md}"
    },
    "secondary": {
      "background": "{color.secondary}",
      "text": "{color.secondary-foreground}",
      "border": "{color.border}"
    }
  },
  "input": {
    "background": "{color.background.default}",
    "text": "{color.text.default}",
    "border": "{color.border.default}",
    "focus-border": "{color.border.focus}",
    "placeholder": "{color.text.subtle}",
    "padding-x": "{spacing.3}",
    "padding-y": "{spacing.2}",
    "height": "{size.10}"
  }
}
```

### 2.2 Token Categories

| Category | Purpose | Examples |
|----------|---------|----------|
| **Color** | All color values | Background, text, border colors |
| **Spacing** | Layout and padding | Margins, paddings, gaps |
| **Typography** | Text properties | Font sizes, weights, line heights |
| **Size** | Dimensional values | Widths, heights, icon sizes |
| **Border** | Border properties | Widths, radius, styles |
| **Shadow** | Elevation effects | Box shadows, drop shadows |
| **Duration** | Animation timing | Transition and animation durations |
| **Easing** | Animation curves | Cubic bezier functions |
| **Z-index** | Stacking order | Layer hierarchy values |
| **Breakpoint** | Responsive queries | Screen size thresholds |

### 2.3 Token Format

Tokens should be defined in a platform-agnostic format (JSON) and transformed for different platforms:

```json
{
  "color": {
    "primary": {
      "value": "222.2 47.4% 11.2%",
      "type": "color",
      "format": "hsl",
      "description": "Primary brand color"
    }
  },
  "spacing": {
    "4": {
      "value": "1rem",
      "type": "spacing",
      "pixel-value": "16px",
      "description": "Base spacing unit"
    }
  }
}
```

**Output Formats:**

- **CSS Variables:** `--color-primary: 222.2 47.4% 11.2%;`
- **SCSS Variables:** `$color-primary: hsl(222.2, 47.4%, 11.2%);`
- **JavaScript:** `export const colorPrimary = 'hsl(222.2 47.4% 11.2%)';`
- **TypeScript:** `export const tokens: DesignTokens = { ... };`
- **JSON:** Direct token consumption

### 2.4 Token Naming Convention

Follow a consistent naming pattern for clarity and predictability:

```
Format: [category]-[property]-[variant]-[state]

Examples:
- color-background-default
- color-background-subtle
- color-text-default
- color-text-disabled
- spacing-4
- font-size-base
- border-radius-md
- shadow-lg
- duration-normal
```

**Benefits:**
- Autocomplete-friendly
- Immediately understandable
- Easy to search
- Consistent across platforms

---

## 3. 🌈 Color System

### 3.1 Color Philosophy

Color is one of the most powerful tools in design. Our color system is built on three principles:

1. **Semantic Meaning** — Colors communicate purpose
2. **Accessibility** — All combinations meet contrast requirements
3. **Flexibility** — Easy to customize and extend

### 3.2 Color Architecture

```
┌────────────────────────────────────────────────┐
│            Color System Structure               │
├────────────────────────────────────────────────┤
│                                                 │
│  Base Colors (Primitive Palette)               │
│    ↓                                            │
│  Semantic Colors (Functional Assignment)       │
│    ↓                                            │
│  Component Colors (Context-Specific)           │
│    ↓                                            │
│  State Colors (Interaction States)             │
│                                                 │
└────────────────────────────────────────────────┘
```

### 3.3 Base Color Palette

#### **Neutral Colors (Gray Scale)**

The foundation of the UI. Used for text, backgrounds, and borders.

```css
--gray-50:  hsl(0, 0%, 98%);    /* Lightest */
--gray-100: hsl(0, 0%, 96%);
--gray-200: hsl(0, 0%, 90%);
--gray-300: hsl(0, 0%, 83%);
--gray-400: hsl(0, 0%, 64%);
--gray-500: hsl(0, 0%, 45%);    /* Mid-point */
--gray-600: hsl(0, 0%, 32%);
--gray-700: hsl(0, 0%, 25%);
--gray-800: hsl(0, 0%, 15%);
--gray-900: hsl(0, 0%, 9%);     /* Darkest */
```

**Usage Guidelines:**
- 50-200: Backgrounds and subtle borders
- 300-500: Secondary text and borders
- 600-900: Primary text and emphasis

#### **Brand Colors**

Primary brand identity colors that define your product's personality.

```css
/* Primary - Main brand color */
--primary-50:  hsl(222, 47%, 95%);
--primary-100: hsl(222, 47%, 90%);
--primary-500: hsl(222, 47%, 50%);   /* Core brand */
--primary-900: hsl(222, 47%, 11%);

/* Secondary - Supporting brand color */
--secondary-50:  hsl(210, 40%, 98%);
--secondary-500: hsl(210, 40%, 50%);
--secondary-900: hsl(210, 40%, 10%);

/* Accent - Attention-grabbing color */
--accent-500: hsl(280, 80%, 55%);
```

#### **Semantic Colors**

Colors with functional meaning across the interface.

```css
/* Success - Positive outcomes */
--success-50:  hsl(142, 76%, 95%);
--success-500: hsl(142, 76%, 36%);
--success-900: hsl(142, 76%, 20%);

/* Error - Problems and destructive actions */
--error-50:  hsl(0, 84%, 95%);
--error-500: hsl(0, 84%, 60%);
--error-900: hsl(0, 84%, 35%);

/* Warning - Caution and important notices */
--warning-50:  hsl(38, 92%, 95%);
--warning-500: hsl(38, 92%, 50%);
--warning-900: hsl(38, 92%, 30%);

/* Info - Neutral information */
--info-50:  hsl(199, 89%, 95%);
--info-500: hsl(199, 89%, 48%);
--info-900: hsl(199, 89%, 28%);
```

### 3.4 Semantic Color Application

Map base colors to semantic purposes:

```css
:root {
  /* Backgrounds */
  --background: var(--gray-50);
  --background-subtle: var(--gray-100);
  --background-emphasis: var(--gray-200);
  --background-inverse: var(--gray-900);
  
  /* Text */
  --text-default: var(--gray-900);
  --text-subtle: var(--gray-600);
  --text-disabled: var(--gray-400);
  --text-inverse: var(--gray-50);
  --text-link: var(--primary-500);
  
  /* Borders */
  --border-default: var(--gray-300);
  --border-strong: var(--gray-400);
  --border-subtle: var(--gray-200);
  
  /* Interactive */
  --interactive-default: var(--primary-500);
  --interactive-hover: var(--primary-600);
  --interactive-active: var(--primary-700);
  --interactive-disabled: var(--gray-300);
  
  /* Feedback */
  --feedback-success: var(--success-500);
  --feedback-error: var(--error-500);
  --feedback-warning: var(--warning-500);
  --feedback-info: var(--info-500);
  
  /* Surface */
  --surface-default: var(--gray-50);
  --surface-raised: #ffffff;
  --surface-overlay: rgba(0, 0, 0, 0.5);
}
```

### 3.5 Color Usage Guidelines

#### **Text Hierarchy**

```css
.text-primary {
  color: var(--text-default);
  /* Headings, primary content */
}

.text-secondary {
  color: var(--text-subtle);
  /* Supporting text, descriptions */
}

.text-tertiary {
  color: var(--gray-500);
  /* Metadata, timestamps */
}

.text-disabled {
  color: var(--text-disabled);
  /* Disabled states */
}
```

#### **Background Layers**

```css
.layer-base {
  background: var(--background);
  /* Page background */
}

.layer-raised {
  background: var(--surface-raised);
  /* Cards, panels, elevated content */
}

.layer-overlay {
  background: var(--surface-overlay);
  /* Modal overlays, drawers */
}
```

### 3.6 Contrast Requirements

All color combinations must meet WCAG 2.1 contrast ratios:

| Use Case | Minimum Ratio | Example |
|----------|---------------|---------|
| Normal Text | 4.5:1 | Body text on backgrounds |
| Large Text (18pt+) | 3:1 | Headings, large labels |
| UI Components | 3:1 | Borders, icons, controls |
| Graphical Objects | 3:1 | Charts, graphs, diagrams |

**Testing Tools:**
- WebAIM Contrast Checker
- Stark (Figma plugin)
- Chrome DevTools (Lighthouse)
- axe DevTools

### 3.7 Color Accessibility Matrix

```
                Background Colors
              ┌────┬────┬────┬────┐
              │ 50 │100 │200 │900 │
        ┌─────┼────┼────┼────┼────┤
        │ 900 │ ✓  │ ✓  │ ✓  │ ✗  │
Text    │ 700 │ ✓  │ ✓  │ ✓  │ ✗  │
Colors  │ 500 │ ✓  │ ⚠  │ ✗  │ ✓  │
        │ 300 │ ✗  │ ✗  │ ✗  │ ✓  │
        └─────┴────┴────┴────┴────┘

Legend: ✓ Pass  ⚠ AA Large Only  ✗ Fail
```

### 3.8 Color Modifiers

Use alpha channels and blend modes for subtle variations:

```css
/* Opacity modifiers */
--opacity-0: 0;
--opacity-10: 0.1;
--opacity-20: 0.2;
--opacity-40: 0.4;
--opacity-60: 0.6;
--opacity-80: 0.8;
--opacity-100: 1;

/* Usage */
.overlay {
  background: hsl(var(--gray-900) / var(--opacity-50));
}

.hover-effect {
  background: hsl(var(--primary) / var(--opacity-10));
}
```

### 3.9 Color Context Guidelines

**Do:**
- ✅ Use semantic colors for their intended purpose
- ✅ Test color combinations for accessibility
- ✅ Provide non-color alternatives (icons, labels)
- ✅ Use color consistently across the system
- ✅ Consider colorblind users (8% of men)

**Don't:**
- ❌ Rely solely on color to convey information
- ❌ Use too many colors (limit to 3-5 brand colors)
- ❌ Ignore contrast requirements
- ❌ Use color arbitrarily
- ❌ Assume everyone sees color the same way

---

## 4. 📝 Typography

### 4.1 Typography Philosophy

Typography is the voice of your interface. Good typography:
- Establishes visual hierarchy
- Enhances readability
- Reinforces brand identity
- Guides user attention

### 4.2 Type Scale

A harmonious type scale creates rhythm and hierarchy. We use a modular scale with a 1.250 ratio (Major Third):

```css
:root {
  /* Base size: 16px */
  --font-size-xs:   0.75rem;   /* 12px */
  --font-size-sm:   0.875rem;  /* 14px */
  --font-size-base: 1rem;      /* 16px - Body text */
  --font-size-lg:   1.125rem;  /* 18px */
  --font-size-xl:   1.25rem;   /* 20px */
  --font-size-2xl:  1.5rem;    /* 24px */
  --font-size-3xl:  1.875rem;  /* 30px */
  --font-size-4xl:  2.25rem;   /* 36px */
  --font-size-5xl:  3rem;      /* 48px - Hero headings */
}
```

**Visual Scale:**

```
5xl  ████████████████████  Hero Heading
4xl  ████████████████      Page Title
3xl  ████████████          Section Heading
2xl  ██████████            Subsection
xl   ████████              Large Body
lg   ███████               Emphasized Text
base ██████                Body Text (Base)
sm   █████                 Secondary Text
xs   ████                  Captions
```

### 4.3 Font Families

```css
:root {
  /* Sans-serif - UI and body text */
  --font-family-sans: 
    'Inter', 
    -apple-system, 
    BlinkMacSystemFont, 
    'Segoe UI', 
    'Roboto', 
    'Oxygen', 
    'Ubuntu', 
    'Cantarell', 
    sans-serif;
  
  /* Serif - Headings and editorial */
  --font-family-serif: 
    'Georgia', 
    'Cambria', 
    'Times New Roman', 
    serif;
  
  /* Mono - Code and data */
  --font-family-mono: 
    'JetBrains Mono',
    'Fira Code', 
    'Consolas', 
    'Monaco', 
    'Courier New', 
    monospace;
}
```

**Font Loading Strategy:**

```html
<!-- Preload critical fonts -->
<link rel="preload" 
      href="/fonts/Inter-Regular.woff2" 
      as="font" 
      type="font/woff2" 
      crossorigin>

<!-- Use font-display: swap for better performance -->
<style>
  @font-face {
    font-family: 'Inter';
    src: url('/fonts/Inter-Regular.woff2') format('woff2');
    font-display: swap;
    font-weight: 400;
  }
</style>
```

### 4.4 Font Weights

```css
:root {
  --font-weight-thin:       100;
  --font-weight-extralight: 200;
  --font-weight-light:      300;
  --font-weight-normal:     400;  /* Body text */
  --font-weight-medium:     500;  /* Emphasis */
  --font-weight-semibold:   600;  /* Headings */
  --font-weight-bold:       700;  /* Strong emphasis */
  --font-weight-extrabold:  800;
  --font-weight-black:      900;
}
```

**Usage Guidelines:**

| Weight | Use Case | Example |
|--------|----------|---------|
| 300 | Light emphasis, large text | Hero text |
| 400 | Body text, default | Paragraphs |
| 500 | Medium emphasis | Labels, captions |
| 600 | Headings, subheadings | Section titles |
| 700 | Strong emphasis | Alerts, CTAs |

### 4.5 Line Height

Line height (leading) affects readability and visual density:

```css
:root {
  --line-height-none:    1;      /* Icons, single lines */
  --line-height-tight:   1.25;   /* Headings */
  --line-height-snug:    1.375;  /* Short text blocks */
  --line-height-normal:  1.5;    /* Body text (optimal) */
  --line-height-relaxed: 1.625;  /* Long-form content */
  --line-height-loose:   2;      /* Poetry, special layouts */
}
```

**Readability Guidelines:**
- Body text (40-75 characters): 1.5-1.625
- Headings: 1.2-1.375
- Short lines (<40 chars): 1.375-1.5
- Long lines (>75 chars): 1.625-2

### 4.6 Letter Spacing (Tracking)

```css
:root {
  --letter-spacing-tighter: -0.05em;  /* Large headings */
  --letter-spacing-tight:   -0.025em; /* Headings */
  --letter-spacing-normal:  0;        /* Body text */
  --letter-spacing-wide:    0.025em;  /* Small caps */
  --letter-spacing-wider:   0.05em;   /* All caps */
  --letter-spacing-widest:  0.1em;    /* Spaced headings */
}
```

**Rules of Thumb:**
- Larger text → Tighter tracking
- Smaller text → Wider tracking
- ALL CAPS → Wider tracking
- Light text on dark → Slightly wider

### 4.7 Text Styles

Pre-defined text styles for common use cases:

```css
/* Display - Hero text */
.text-display {
  font-size: var(--font-size-5xl);
  font-weight: var(--font-weight-bold);
  line-height: var(--line-height-tight);
  letter-spacing: var(--letter-spacing-tighter);
}

/* Heading 1 - Page title */
.text-h1 {
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-semibold);
  line-height: var(--line-height-tight);
}

/* Heading 2 - Section title */
.text-h2 {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-semibold);
  line-height: var(--line-height-snug);
}

/* Heading 3 - Subsection */
.text-h3 {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  line-height: var(--line-height-snug);
}

/* Body - Default text */
.text-body {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-normal);
  line-height: var(--line-height-normal);
}

/* Body Small - Secondary text */
.text-body-sm {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-normal);
  line-height: var(--line-height-normal);
}

/* Caption - Metadata */
.text-caption {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-normal);
  line-height: var(--line-height-normal);
  color: var(--text-subtle);
}

/* Code - Monospace */
.text-code {
  font-family: var(--font-family-mono);
  font-size: 0.875em;
  background: var(--background-subtle);
  padding: 0.125rem 0.25rem;
  border-radius: var(--radius-sm);
}
```

### 4.8 Responsive Typography

Typography should adapt to screen size:

```css
/* Mobile-first approach */
.text-responsive {
  font-size: var(--font-size-base);
}

/* Tablet and up */
@media (min-width: 768px) {
  .text-responsive {
    font-size: var(--font-size-lg);
  }
}

/* Desktop and up */
@media (min-width: 1024px) {
  .text-responsive {
    font-size: var(--font-size-xl);
  }
}

/* Fluid typography using clamp */
.text-fluid {
  font-size: clamp(
    1rem,              /* Min: 16px */
    0.875rem + 0.5vw,  /* Preferred: scales with viewport */
    1.5rem             /* Max: 24px */
  );
}
```

### 4.9 Typography Hierarchy

Visual hierarchy through size, weight, and spacing:

```
┌─────────────────────────────────────┐
│  Display (5xl, bold)                │  ← Highest hierarchy
│                                     │
│  H1 (4xl, semibold)                 │
│                                     │
│  H2 (3xl, semibold)                 │
│                                     │
│  H3 (2xl, semibold)                 │
│                                     │
│  Body (base, normal)                │  ← Base hierarchy
│                                     │
│  Body Small (sm, normal)            │
│                                     │
│  Caption (xs, normal, subtle)       │  ← Lowest hierarchy
└─────────────────────────────────────┘
```

### 4.10 Best Practices

**Do:**
- ✅ Use a consistent type scale
- ✅ Limit to 2-3 font families maximum
- ✅ Ensure sufficient contrast (4.5:1 minimum)
- ✅ Use relative units (rem, em) for accessibility
- ✅ Test with actual content, not Lorem Ipsum
- ✅ Optimize line length (50-75 characters)

**Don't:**
- ❌ Use more than 3-4 font weights
- ❌ Center-align long blocks of text
- ❌ Set body text smaller than 16px
- ❌ Use all caps for long text
- ❌ Ignore line height in dense layouts
- ❌ Use pure black (#000) on pure white (#fff)

---

## 5. 📏 Spacing & Layout

### 5.1 Spacing Philosophy

Consistent spacing creates visual rhythm and improves scanability. Our spacing system uses a base-8 scale for mathematical harmony.

**Why Base-8?**
- Easily divisible (8, 4, 2, 1)
- Aligns with device pixel densities
- Creates consistent vertical rhythm
- Simplifies math for developers

### 5.2 Spacing Scale

```css
:root {
  /* Base unit: 4px */
  --space-0:   0;
  --space-0-5: 0.125rem;  /* 2px  */
  --space-1:   0.25rem;   /* 4px  */
  --space-1-5: 0.375rem;  /* 6px  */
  --space-2:   0.5rem;    /* 8px  */
  --space-2-5: 0.625rem;  /* 10px */
  --space-3:   0.75rem;   /* 12px */
  --space-3-5: 0.875rem;  /* 14px */
  --space-4:   1rem;      /* 16px - Base spacing */
  --space-5:   1.25rem;   /* 20px */
  --space-6:   1.5rem;    /* 24px */
  --space-7:   1.75rem;   /* 28px */
  --space-8:   2rem;      /* 32px */
  --space-9:   2.25rem;   /* 36px */
  --space-10:  2.5rem;    /* 40px */
  --space-11:  2.75rem;   /* 44px */
  --space-12:  3rem;      /* 48px */
  --space-14:  3.5rem;    /* 56px */
  --space-16:  4rem;      /* 64px */
  --space-20:  5rem;      /* 80px */
  --space-24:  6rem;      /* 96px */
  --space-32:  8rem;      /* 128px */
}
```

**Visual Spacing Scale:**

```
32  ████████████████████████████████  Section spacing
24  ████████████████████████          Large gaps
20  ████████████████████              Component spacing
16  ████████████████                  Major spacing
12  ████████████                      Card padding
8   ████████                          Element spacing
6   ██████                            Tight spacing
4   ████                              Base unit
2   ██                                Hairline gaps
1   █                                 Micro spacing
```

### 5.3 Spacing Application

#### **Component Internal Spacing**

```css
/* Buttons */
.button {
  padding: var(--space-2) var(--space-4);  /* 8px 16px */
}

.button-sm {
  padding: var(--space-1-5) var(--space-3);  /* 6px 12px */
}

.button-lg {
  padding: var(--space-3) var(--space-6);  /* 12px 24px */
}

/* Cards */
.card {
  padding: var(--space-6);  /* 24px */
}

.card-compact {
  padding: var(--space-4);  /* 16px */
}

/* Form fields */
.input {
  padding: var(--space-2-5) var(--space-3);  /* 10px 12px */
}
```

#### **Layout Spacing**

```css
/* Stack spacing (vertical) */
.stack-xs {
  gap: var(--space-2);  /* 8px */
}

.stack-sm {
  gap: var(--space-4);  /* 16px */
}

.stack-md {
  gap: var(--space-6);  /* 24px */
}

.stack-lg {
  gap: var(--space-8);  /* 32px */
}

/* Grid spacing */
.grid-tight {
  gap: var(--space-4);  /* 16px */
}

.grid-normal {
  gap: var(--space-6);  /* 24px */
}

.grid-loose {
  gap: var(--space-8);  /* 32px */
}

/* Section spacing */
.section {
  padding-top: var(--space-16);     /* 64px */
  padding-bottom: var(--space-16);  /* 64px */
}
```

### 5.4 Container System

```css
/* Container widths */
:root {
  --container-sm:  640px;
  --container-md:  768px;
  --container-lg:  1024px;
  --container-xl:  1280px;
  --container-2xl: 1536px;
}

/* Container with responsive padding */
.container {
  width: 100%;
  max-width: var(--container-xl);
  margin-inline: auto;
  padding-inline: var(--space-4);  /* 16px on mobile */
}

@media (min-width: 768px) {
  .container {
    padding-inline: var(--space-6);  /* 24px on tablet */
  }
}

@media (min-width: 1024px) {
  .container {
    padding-inline: var(--space-8);  /* 32px on desktop */
  }
}
```

### 5.5 Margin & Padding Utilities

```css
/* Margin utilities */
.m-0  { margin: var(--space-0); }
.m-1  { margin: var(--space-1); }
.m-2  { margin: var(--space-2); }
.m-4  { margin: var(--space-4); }
/* ... continue for all spacing values */

/* Directional margins */
.mt-4 { margin-top: var(--space-4); }
.mr-4 { margin-right: var(--space-4); }
.mb-4 { margin-bottom: var(--space-4); }
.ml-4 { margin-left: var(--space-4); }

/* Horizontal & Vertical */
.mx-4 { margin-inline: var(--space-4); }
.my-4 { margin-block: var(--space-4); }

/* Padding utilities */
.p-0  { padding: var(--space-0); }
.p-1  { padding: var(--space-1); }
.p-2  { padding: var(--space-2); }
.p-4  { padding: var(--space-4); }

/* Directional paddings */
.pt-4 { padding-top: var(--space-4); }
.pr-4 { padding-right: var(--space-4); }
.pb-4 { padding-bottom: var(--space-4); }
.pl-4 { padding-left: var(--space-4); }

/* Horizontal & Vertical */
.px-4 { padding-inline: var(--space-4); }
.py-4 { padding-block: var(--space-4); }
```

### 5.6 Gap Utilities (Flexbox & Grid)

```css
.gap-0  { gap: var(--space-0); }
.gap-1  { gap: var(--space-1); }
.gap-2  { gap: var(--space-2); }
.gap-4  { gap: var(--space-4); }
.gap-6  { gap: var(--space-6); }
.gap-8  { gap: var(--space-8); }

/* Row & Column gaps */
.gap-x-4 { column-gap: var(--space-4); }
.gap-y-4 { row-gap: var(--space-4); }
```

### 5.7 Optical Spacing Adjustments

Sometimes mathematical spacing needs visual adjustment:

```css
/* Icons next to text need less space than text alone */
.icon-text-spacing {
  gap: var(--space-2);  /* Instead of space-4 */
}

/* Headings often need negative top margin */
.heading-after-paragraph {
  margin-top: calc(var(--space-8) * -0.25);
}

/* First/last child adjustments */
.stack > *:first-child {
  margin-top: 0;
}

.stack > *:last-child {
  margin-bottom: 0;
}
```

### 5.8 Spacing Decision Tree

```
Need spacing? → Ask these questions:

1. Is it component-internal padding?
   → Use component tokens (button-padding-x)

2. Is it between related items?
   → Use small gaps (space-2 to space-4)

3. Is it between sections?
   → Use large gaps (space-8 to space-16)

4. Is it page-level spacing?
   → Use section spacing (space-16 to space-32)
```

---

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

## 7. 📱 Responsive Design

### 7.1 Responsive Philosophy

Mobile-first responsive design ensures optimal experiences across all devices. Our approach:

1. **Mobile-First** — Design for smallest screens, enhance for larger
2. **Content-First** — Prioritize content over device-specific layouts
3. **Progressive Enhancement** — Add features as screen size increases
4. **Performance-Aware** — Optimize for slower mobile connections

### 7.2 Breakpoint System

```css
:root {
  /* Breakpoint values */
  --breakpoint-sm:  640px;   /* Landscape phones */
  --breakpoint-md:  768px;   /* Tablets */
  --breakpoint-lg:  1024px;  /* Laptops */
  --breakpoint-xl:  1280px;  /* Desktops */
  --breakpoint-2xl: 1536px;  /* Large desktops */
}

/* Media query mixins */
@custom-media --sm  (min-width: 640px);
@custom-media --md  (min-width: 768px);
@custom-media --lg  (min-width: 1024px);
@custom-media --xl  (min-width: 1280px);
@custom-media --2xl (min-width: 1536px);
```

**Breakpoint Usage:**

```css
/* Mobile-first approach */
.component {
  padding: var(--space-4);
  font-size: var(--font-size-sm);
}

/* Tablet and up */
@media (min-width: 768px) {
  .component {
    padding: var(--space-6);
    font-size: var(--font-size-base);
  }
}

/* Desktop and up */
@media (min-width: 1024px) {
  .component {
    padding: var(--space-8);
    font-size: var(--font-size-lg);
  }
}
```

### 7.3 Device Contexts

| Device | Breakpoint | Viewport Width | Design Considerations |
|--------|------------|----------------|----------------------|
| **Mobile (Portrait)** | < 640px | 320px - 640px | Single column, touch targets 44px+ |
| **Mobile (Landscape)** | 640px - 768px | 640px - 768px | Two columns possible, reduced vertical space |
| **Tablet (Portrait)** | 768px - 1024px | 768px - 1024px | Sidebar layouts, multi-column content |
| **Tablet (Landscape)** | 1024px - 1280px | 1024px - 1280px | Full desktop features, side-by-side views |
| **Desktop** | 1280px+ | 1280px+ | Maximum content width, rich interactions |

### 7.4 Responsive Grid System

```css
/* Mobile: Single column */
.grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: 1fr;
}

/* Tablet: 2 columns */
@media (min-width: 768px) {
  .grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-6);
  }
}

/* Desktop: 3-4 columns */
@media (min-width: 1024px) {
  .grid {
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-8);
  }
}

@media (min-width: 1280px) {
  .grid {
    grid-template-columns: repeat(4, 1fr);
  }
}
```

### 7.5 Container Queries

Modern responsive design with container queries:

```css
/* Container-aware components */
.card-container {
  container-type: inline-size;
  container-name: card;
}

.card {
  padding: var(--space-4);
}

/* When container is > 400px */
@container card (min-width: 400px) {
  .card {
    display: grid;
    grid-template-columns: 100px 1fr;
    gap: var(--space-4);
  }
}

/* When container is > 600px */
@container card (min-width: 600px) {
  .card {
    padding: var(--space-6);
  }
}
```

### 7.6 Responsive Typography

```css
/* Fluid typography with clamp */
.heading {
  font-size: clamp(
    1.5rem,                /* Min: 24px (mobile) */
    1rem + 2vw,            /* Preferred: scales with viewport */
    3rem                   /* Max: 48px (desktop) */
  );
}

/* Responsive line length */
.text-content {
  max-width: min(65ch, 90vw);
  margin-inline: auto;
}

/* Responsive line height */
.paragraph {
  line-height: 1.6;
}

@media (min-width: 768px) {
  .paragraph {
    line-height: 1.7;
  }
}
```

### 7.7 Touch Targets

Minimum touch target sizes for mobile:

```css
:root {
  --touch-target-min: 44px;  /* iOS/Android minimum */
  --touch-target-comfortable: 48px;
}

/* Ensure adequate touch targets */
.button,
.link,
.checkbox,
.radio {
  min-height: var(--touch-target-min);
  min-width: var(--touch-target-min);
}

/* Increase spacing on mobile */
.button-group {
  gap: var(--space-2);
}

@media (min-width: 768px) {
  .button-group {
    gap: var(--space-1);
  }
}
```

### 7.8 Responsive Navigation

```css
/* Mobile: Hamburger menu */
.navigation {
  position: fixed;
  top: 0;
  left: -100%;
  width: 80vw;
  height: 100vh;
  transition: left 0.3s;
}

.navigation.open {
  left: 0;
}

/* Desktop: Horizontal menu */
@media (min-width: 1024px) {
  .navigation {
    position: static;
    width: auto;
    height: auto;
    display: flex;
    flex-direction: row;
  }
}
```

### 7.9 Responsive Images

```html
<!-- Responsive image with srcset -->
<img 
  src="image-800w.jpg"
  srcset="
    image-400w.jpg 400w,
    image-800w.jpg 800w,
    image-1200w.jpg 1200w
  "
  sizes="
    (max-width: 640px) 100vw,
    (max-width: 1024px) 50vw,
    800px
  "
  alt="Description"
  loading="lazy"
/>

<!-- Art direction with picture -->
<picture>
  <source media="(min-width: 1024px)" srcset="desktop.jpg" />
  <source media="(min-width: 768px)" srcset="tablet.jpg" />
  <img src="mobile.jpg" alt="Description" />
</picture>
```

### 7.10 Responsive Tables

```css
/* Mobile: Card-based layout */
.table {
  display: block;
}

.table-row {
  display: block;
  padding: var(--space-4);
  border-bottom: 1px solid var(--border);
}

.table-cell {
  display: flex;
  justify-content: space-between;
  padding: var(--space-2) 0;
}

.table-cell::before {
  content: attr(data-label);
  font-weight: var(--font-weight-semibold);
}

/* Desktop: Traditional table */
@media (min-width: 768px) {
  .table {
    display: table;
  }
  
  .table-row {
    display: table-row;
  }
  
  .table-cell {
    display: table-cell;
    padding: var(--space-3);
  }
  
  .table-cell::before {
    content: none;
  }
}
```

### 7.11 Responsive Utilities

```css
/* Display utilities */
.hide-mobile { display: none; }
.show-mobile { display: block; }

@media (min-width: 768px) {
  .hide-mobile { display: block; }
  .show-mobile { display: none; }
  .hide-tablet { display: none; }
}

@media (min-width: 1024px) {
  .hide-desktop { display: none; }
  .hide-tablet { display: block; }
}

/* Responsive spacing */
.responsive-padding {
  padding: var(--space-4);
}

@media (min-width: 768px) {
  .responsive-padding {
    padding: var(--space-6);
  }
}

@media (min-width: 1024px) {
  .responsive-padding {
    padding: var(--space-8);
  }
}
```

### 7.12 Best Practices

**Do:**
- ✅ Design mobile-first, enhance progressively
- ✅ Test on real devices, not just browser resize
- ✅ Use relative units (rem, em, %) for flexibility
- ✅ Optimize images for different screen sizes
- ✅ Ensure touch targets are 44px minimum
- ✅ Test in landscape and portrait orientations

**Don't:**
- ❌ Design desktop-first and scale down
- ❌ Use fixed pixel widths for layouts
- ❌ Assume hover interactions (mobile has no hover)
- ❌ Ignore performance on mobile networks
- ❌ Use small text (< 16px) on mobile
- ❌ Forget to test on actual devices

---

## 8. ♿ Accessibility

### 8.1 Accessibility Philosophy

Accessibility (a11y) is not optional—it's a fundamental requirement. Inclusive design benefits everyone:

- **15%** of the world has some form of disability
- **100%** benefit from accessible design (situational disabilities)
- Legal requirements (ADA, Section 508, WCAG)

**Core Accessibility Principles (POUR):**

1. **Perceivable** — Information must be presentable to users in ways they can perceive
2. **Operable** — Interface components must be operable by all users
3. **Understandable** — Information and operation must be understandable
4. **Robust** — Content must work with current and future technologies

### 8.2 WCAG 2.1 Compliance

Target: **Level AA** (minimum), **Level AAA** (where possible)

| Level | Description | Requirement |
|-------|-------------|-------------|
| **A** | Basic accessibility | Minimum for all content |
| **AA** | Recommended level | Industry standard, legal requirement |
| **AAA** | Enhanced accessibility | Ideal, but not always feasible |

### 8.3 Color Contrast

**WCAG 2.1 Contrast Requirements:**

```css
/* Text contrast ratios */
:root {
  /* Normal text (< 18pt): 4.5:1 minimum (AA) */
  --text-on-light: var(--gray-900);    /* 15:1 ratio */
  --text-on-primary: var(--gray-50);    /* 8:1 ratio */
  
  /* Large text (≥ 18pt): 3:1 minimum (AA) */
  --large-text-on-light: var(--gray-700); /* 7:1 ratio */
  
  /* UI components: 3:1 minimum (AA) */
  --border-on-light: var(--gray-400);   /* 3.5:1 ratio */
}

/* Ensure sufficient contrast */
.text-default {
  color: var(--text-default);
  background: var(--background);
  /* Ratio: 16:1 (exceeds 4.5:1) */
}

.button-primary {
  color: var(--primary-foreground);
  background: var(--primary);
  /* Ratio: 8:1 (exceeds 4.5:1) */
}
```

**Testing Tools:**
- Chrome DevTools Lighthouse
- WebAIM Contrast Checker
- Stark (Figma plugin)
- axe DevTools

### 8.4 Keyboard Navigation

All interactive elements must be keyboard accessible:

```css
/* Visible focus indicators */
:focus-visible {
  outline: 2px solid var(--ring);
  outline-offset: 2px;
  /* Ring color: 4.5:1 contrast minimum */
}

/* Remove default outline but keep focus-visible */
:focus:not(:focus-visible) {
  outline: none;
}

/* Custom focus styles */
.button:focus-visible {
  outline: 2px solid var(--primary);
  outline-offset: 2px;
  box-shadow: 0 0 0 4px hsl(var(--primary) / 0.2);
}
```

**Keyboard Navigation Patterns:**

| Key | Action |
|-----|--------|
| **Tab** | Move focus forward |
| **Shift + Tab** | Move focus backward |
| **Enter/Space** | Activate button/link |
| **Escape** | Close modal/dropdown |
| **Arrow Keys** | Navigate menus/lists |
| **Home/End** | Jump to first/last item |

### 8.5 ARIA Attributes

Use ARIA to enhance semantics when HTML alone isn't sufficient:

```html
<!-- Button with accessible name -->
<button aria-label="Close dialog">
  <svg>...</svg> <!-- Icon only, no text -->
</button>

<!-- Expandable section -->
<button 
  aria-expanded="false" 
  aria-controls="content-1"
>
  Toggle Section
</button>
<div id="content-1" hidden>
  Content...
</div>

<!-- Loading state -->
<button aria-busy="true" aria-live="polite">
  Loading...
</button>

<!-- Invalid input -->
<input 
  type="email" 
  aria-invalid="true" 
  aria-describedby="email-error"
/>
<span id="email-error" role="alert">
  Please enter a valid email
</span>

<!-- Progress indicator -->
<div 
  role="progressbar" 
  aria-valuenow="60" 
  aria-valuemin="0" 
  aria-valuemax="100"
>
  60% complete
</div>
```

### 8.6 Semantic HTML

Use semantic HTML elements for better accessibility:

```html
<!-- ❌ Bad: Non-semantic divs -->
<div class="header">
  <div class="navigation">
    <div class="link">Home</div>
  </div>
</div>

<!-- ✅ Good: Semantic HTML -->
<header>
  <nav aria-label="Main navigation">
    <a href="/">Home</a>
  </nav>
</header>

<!-- Form with semantic structure -->
<form>
  <fieldset>
    <legend>Personal Information</legend>
    <label for="name">Name</label>
    <input id="name" type="text" required />
  </fieldset>
</form>

<!-- Article with proper structure -->
<article>
  <header>
    <h1>Article Title</h1>
    <p>By <span>Author Name</span></p>
  </header>
  <section>
    <h2>Section Title</h2>
    <p>Content...</p>
  </section>
  <footer>
    <time datetime="2025-10-29">October 29, 2025</time>
  </footer>
</article>
```

### 8.7 Screen Reader Support

Optimize for screen readers:

```html
<!-- Skip navigation link -->
<a href="#main-content" class="skip-link">
  Skip to main content
</a>

<nav>...</nav>

<main id="main-content">
  <!-- Main content -->
</main>

<!-- Visually hidden but accessible to screen readers -->
<span class="sr-only">
  Screen reader only text
</span>

<style>
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}
</style>

<!-- Live regions for dynamic content -->
<div aria-live="polite" aria-atomic="true">
  <!-- Announcements for screen readers -->
</div>

<div aria-live="assertive">
  <!-- Urgent announcements -->
</div>
```

### 8.8 Focus Management

Manage focus in dynamic interfaces:

```typescript
// Modal focus trap
class Modal {
  open() {
    // Store current focus
    this.previousFocus = document.activeElement;
    
    // Show modal
    this.element.removeAttribute('hidden');
    
    // Focus first focusable element
    const firstFocusable = this.element.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
    firstFocusable?.focus();
    
    // Trap focus within modal
    this.element.addEventListener('keydown', this.trapFocus);
  }
  
  close() {
    // Hide modal
    this.element.setAttribute('hidden', '');
    
    // Restore focus
    this.previousFocus?.focus();
    
    // Remove focus trap
    this.element.removeEventListener('keydown', this.trapFocus);
  }
  
  trapFocus(e) {
    if (e.key === 'Tab') {
      const focusableElements = this.element.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];
      
      if (e.shiftKey && document.activeElement === firstElement) {
        e.preventDefault();
        lastElement.focus();
      } else if (!e.shiftKey && document.activeElement === lastElement) {
        e.preventDefault();
        firstElement.focus();
      }
    }
  }
}
```

### 8.9 Form Accessibility

```html
<!-- Accessible form field -->
<div class="form-field">
  <label for="email">
    Email Address
    <span aria-label="required">*</span>
  </label>
  <input 
    id="email"
    type="email"
    required
    aria-required="true"
    aria-describedby="email-hint email-error"
  />
  <span id="email-hint" class="hint">
    We'll never share your email
  </span>
  <span id="email-error" class="error" role="alert" aria-live="polite">
    <!-- Error message appears here -->
  </span>
</div>

<!-- Checkbox group -->
<fieldset>
  <legend>Notification Preferences</legend>
  <div>
    <input type="checkbox" id="email-notif" name="notifications" value="email" />
    <label for="email-notif">Email notifications</label>
  </div>
  <div>
    <input type="checkbox" id="sms-notif" name="notifications" value="sms" />
    <label for="sms-notif">SMS notifications</label>
  </div>
</fieldset>

<!-- Radio group -->
<fieldset>
  <legend>Shipping Method</legend>
  <div>
    <input type="radio" id="standard" name="shipping" value="standard" />
    <label for="standard">Standard (5-7 days)</label>
  </div>
  <div>
    <input type="radio" id="express" name="shipping" value="express" />
    <label for="express">Express (2-3 days)</label>
  </div>
</fieldset>
```

### 8.10 Alternative Text

Provide meaningful alt text for images:

```html
<!-- Informative image -->
<img src="chart.png" alt="Bar chart showing 60% increase in sales from Q1 to Q2" />

<!-- Decorative image -->
<img src="decorative-line.png" alt="" />
<!-- Empty alt for decorative images -->

<!-- Functional image (button/link) -->
<a href="/search">
  <img src="search-icon.svg" alt="Search" />
</a>

<!-- Complex image with long description -->
<figure>
  <img src="complex-diagram.png" alt="System architecture diagram" aria-describedby="diagram-desc" />
  <figcaption id="diagram-desc">
    The diagram shows a three-tier architecture with...
  </figcaption>
</figure>
```

### 8.11 Reduced Motion

Respect user motion preferences:

```css
/* Respect prefers-reduced-motion */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}

/* Provide alternative for animations */
.animated-element {
  animation: slide-in 0.3s ease-out;
}

@media (prefers-reduced-motion: reduce) {
  .animated-element {
    animation: none;
    /* Provide instant state change instead */
  }
}
```

### 8.12 Accessibility Testing Checklist

**Automated Testing:**
- [ ] Run Lighthouse accessibility audit
- [ ] Use axe DevTools extension
- [ ] Validate HTML (W3C Validator)
- [ ] Check color contrast ratios

**Manual Testing:**
- [ ] Navigate entire site with keyboard only
- [ ] Test with screen reader (NVDA, JAWS, VoiceOver)
- [ ] Zoom to 200% (text should reflow)
- [ ] Test with browser extensions disabled
- [ ] Check focus indicators are visible
- [ ] Verify form errors are announced

**User Testing:**
- [ ] Test with actual users with disabilities
- [ ] Get feedback on assistive tech experience
- [ ] Iterate based on real-world usage

---

## 9. 🎬 Animation & Motion

### 9.1 Motion Philosophy

Animation serves purpose—it's not decoration. Good motion:

1. **Guides attention** — Draws eye to important changes
2. **Provides feedback** — Confirms user actions
3. **Shows relationships** — Connects cause and effect
4. **Maintains context** — Helps users track state changes
5. **Adds personality** — Makes interface feel alive

### 9.2 Motion Principles

#### **Duration**

Animation speed affects perceived quality:

```css
:root {
  /* Micro-interactions */
  --duration-instant: 50ms;   /* Hover feedback */
  --duration-fast:    100ms;  /* Button press */
  --duration-normal:  200ms;  /* Standard transitions */
  --duration-slow:    300ms;  /* Drawers, modals */
  --duration-slower:  500ms;  /* Page transitions */
}

/* Usage */
.button {
  transition: background-color var(--duration-fast);
}

.modal {
  transition: opacity var(--duration-normal);
}

.drawer {
  transition: transform var(--duration-slow);
}
```

**Duration Guidelines:**

| Duration | Use Case | Example |
|----------|----------|---------|
| 50ms | Instant feedback | Hover state |
| 100ms | Quick interactions | Button press, checkbox |
| 200ms | Standard transitions | Color change, opacity |
| 300ms | Larger elements | Modal, dropdown |
| 500ms+ | Complex animations | Page transitions, loaders |

#### **Easing**

Easing curves create natural motion:

```css
:root {
  /* Standard easings */
  --ease-linear:     linear;
  --ease-in:         cubic-bezier(0.4, 0, 1, 1);
  --ease-out:        cubic-bezier(0, 0, 0.2, 1);
  --ease-in-out:     cubic-bezier(0.4, 0, 0.2, 1);
  
  /* Custom easings */
  --ease-smooth:     cubic-bezier(0.4, 0, 0.2, 1);      /* Material Design */
  --ease-bounce:     cubic-bezier(0.68, -0.55, 0.27, 1.55);
  --ease-elastic:    cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

/* Usage */
.slide-in {
  animation: slide 300ms var(--ease-out);
}

.bounce {
  animation: bounce 500ms var(--ease-bounce);
}
```

**Easing Selection:**

- **Ease-out** — Entering elements (fast start, slow end)
- **Ease-in** — Exiting elements (slow start, fast end)
- **Ease-in-out** — Persistent elements (smooth both ends)

### 9.3 Common Animation Patterns

#### **Fade**

```css
@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes fade-out {
  from {
    opacity: 1;
  }
  to {
    opacity: 0;
  }
}

.fade-in {
  animation: fade-in var(--duration-normal) var(--ease-out);
}
```

#### **Slide**

```css
@keyframes slide-in-bottom {
  from {
    transform: translateY(100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@keyframes slide-out-top {
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(-100%);
    opacity: 0;
  }
}

.slide-in {
  animation: slide-in-bottom var(--duration-slow) var(--ease-out);
}
```

#### **Scale**

```css
@keyframes scale-in {
  from {
    transform: scale(0.95);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

@keyframes scale-out {
  from {
    transform: scale(1);
    opacity: 1;
  }
  to {
    transform: scale(0.95);
    opacity: 0;
  }
}

.scale-in {
  animation: scale-in var(--duration-normal) var(--ease-out);
}
```

#### **Spin**

```css
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.spinner {
  animation: spin 1s linear infinite;
}
```

### 9.4 Transition Properties

```css
/* Single property transition */
.button {
  transition: background-color var(--duration-fast) var(--ease-out);
}

/* Multiple properties */
.card {
  transition: 
    transform var(--duration-normal) var(--ease-out),
    box-shadow var(--duration-normal) var(--ease-out);
}

/* All properties (use sparingly) */
.element {
  transition: all var(--duration-normal) var(--ease-out);
}

/* Delayed transition */
.delayed {
  transition: opacity var(--duration-normal) var(--ease-out) 100ms;
}
```

### 9.5 Loading States

```css
/* Skeleton loader */
@keyframes skeleton {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.skeleton {
  background: linear-gradient(
    90deg,
    var(--gray-200) 0%,
    var(--gray-100) 50%,
    var(--gray-200) 100%
  );
  background-size: 200% 100%;
  animation: skeleton 1.5s ease-in-out infinite;
}

/* Pulse */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.loading-pulse {
  animation: pulse 2s ease-in-out infinite;
}

/* Progress bar */
@keyframes progress {
  from {
    transform: translateX(-100%);
  }
  to {
    transform: translateX(100%);
  }
}

.progress-indeterminate::after {
  content: '';
  position: absolute;
  width: 50%;
  height: 100%;
  background: var(--primary);
  animation: progress 1.5s ease-in-out infinite;
}
```

### 9.6 Micro-interactions

```css
/* Button press */
.button:active {
  transform: scale(0.98);
  transition: transform var(--duration-instant);
}

/* Checkbox check */
@keyframes check {
  0% {
    stroke-dashoffset: 24;
  }
  100% {
    stroke-dashoffset: 0;
  }
}

.checkbox-icon {
  stroke-dasharray: 24;
  stroke-dashoffset: 24;
  transition: stroke-dashoffset var(--duration-normal) var(--ease-out);
}

.checkbox:checked + .checkbox-icon {
  stroke-dashoffset: 0;
}

/* Toggle switch */
.toggle-knob {
  transform: translateX(0);
  transition: transform var(--duration-normal) var(--ease-out);
}

.toggle:checked + .toggle-knob {
  transform: translateX(20px);
}
```

### 9.7 Page Transitions

```css
/* View transition API */
@view-transition {
  navigation: auto;
}

/* Customize transition */
::view-transition-old(root) {
  animation: fade-out var(--duration-normal) var(--ease-in);
}

::view-transition-new(root) {
  animation: fade-in var(--duration-normal) var(--ease-out);
}

/* Fallback for non-supporting browsers */
@supports not (view-transition-name: none) {
  .page-enter {
    animation: fade-in var(--duration-slower) var(--ease-out);
  }
  
  .page-exit {
    animation: fade-out var(--duration-slower) var(--ease-in);
  }
}
```

### 9.8 Scroll Animations

```css
/* Scroll-triggered animations */
.fade-in-on-scroll {
  opacity: 0;
  transform: translateY(20px);
  transition: 
    opacity var(--duration-slow) var(--ease-out),
    transform var(--duration-slow) var(--ease-out);
}

.fade-in-on-scroll.visible {
  opacity: 1;
  transform: translateY(0);
}

/* Intersection Observer implementation */
<script>
const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('visible');
    }
  });
}, {
  threshold: 0.1
});

document.querySelectorAll('.fade-in-on-scroll').forEach(el => {
  observer.observe(el);
});
</script>
```

### 9.9 Performance Optimization

```css
/* Use transform and opacity for best performance */
/* ✅ Good: GPU-accelerated */
.element {
  transform: translateX(100px);
  opacity: 0.5;
}

/* ❌ Bad: Triggers layout/paint */
.element {
  left: 100px;
  margin-left: 20px;
}

/* Force GPU acceleration */
.accelerated {
  will-change: transform, opacity;
  /* Remove after animation completes */
}

/* Contain layout shifts */
.animated-container {
  contain: layout style paint;
}
```

### 9.10 Motion Best Practices

**Do:**
- ✅ Keep animations under 500ms (most cases)
- ✅ Use `transform` and `opacity` for performance
- ✅ Respect `prefers-reduced-motion`
- ✅ Provide animation purpose (feedback, guidance)
- ✅ Test on slower devices
- ✅ Use will-change sparingly

**Don't:**
- ❌ Animate width, height, or position (use transform)
- ❌ Animate too many elements simultaneously
- ❌ Use animations without purpose
- ❌ Ignore motion sensitivity
- ❌ Create distracting animations
- ❌ Forget to remove will-change after animation

---

## 10. 🌓 Dark Mode

### 10.1 Dark Mode Philosophy

Dark mode isn't just inverted colors—it's a carefully designed alternative color scheme that:

1. **Reduces eye strain** in low-light environments
2. **Saves battery** on OLED displays
3. **Provides user choice** and personalization
4. **Maintains accessibility** and readability

### 10.2 Implementation Strategy

```css
/* Strategy 1: CSS Classes */
:root {
  /* Light mode (default) */
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
}

.dark {
  /* Dark mode */
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
}

/* Strategy 2: Data Attribute */
[data-theme="light"] {
  --background: 0 0% 100%;
}

[data-theme="dark"] {
  --background: 222.2 84% 4.9%;
}

/* Strategy 3: Media Query */
@media (prefers-color-scheme: dark) {
  :root {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
  }
}
```

### 10.3 Complete Color Tokens

```css
:root {
  /* Light mode colors */
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  
  --card: 0 0% 100%;
  --card-foreground: 222.2 84% 4.9%;
  
  --popover: 0 0% 100%;
  --popover-foreground: 222.2 84% 4.9%;
  
  --primary: 222.2 47.4% 11.2%;
  --primary-foreground: 210 40% 98%;
  
  --secondary: 210 40% 96.1%;
  --secondary-foreground: 222.2 47.4% 11.2%;
  
  --muted: 210 40% 96.1%;
  --muted-foreground: 215.4 16.3% 46.9%;
  
  --accent: 210 40% 96.1%;
  --accent-foreground: 222.2 47.4% 11.2%;
  
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 210 40% 98%;
  
  --border: 214.3 31.8% 91.4%;
  --input: 214.3 31.8% 91.4%;
  --ring: 222.2 84% 4.9%;
  
  --radius: 0.5rem;
}

.dark {
  /* Dark mode colors */
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
  
  --card: 222.2 84% 4.9%;
  --card-foreground: 210 40% 98%;
  
  --popover: 222.2 84% 4.9%;
  --popover-foreground: 210 40% 98%;
  
  --primary: 210 40% 98%;
  --primary-foreground: 222.2 47.4% 11.2%;
  
  --secondary: 217.2 32.6% 17.5%;
  --secondary-foreground: 210 40% 98%;
  
  --muted: 217.2 32.6% 17.5%;
  --muted-foreground: 215 20.2% 65.1%;
  
  --accent: 217.2 32.6% 17.5%;
  --accent-foreground: 210 40% 98%;
  
  --destructive: 0 62.8% 30.6%;
  --destructive-foreground: 210 40% 98%;
  
  --border: 217.2 32.6% 17.5%;
  --input: 217.2 32.6% 17.5%;
  --ring: 212.7 26.8% 83.9%;
}
```

### 10.4 Dark Mode JavaScript

```typescript
// Theme management
class ThemeManager {
  private theme: 'light' | 'dark' | 'system' = 'system';
  
  constructor() {
    this.init();
  }
  
  init() {
    // Load saved preference
    const saved = localStorage.getItem('theme');
    if (saved) {
      this.theme = saved as 'light' | 'dark' | 'system';
    }
    
    // Apply theme
    this.apply();
    
    // Listen for system changes
    window.matchMedia('(prefers-color-scheme: dark)')
      .addEventListener('change', () => {
        if (this.theme === 'system') {
          this.apply();
        }
      });
  }
  
  setTheme(theme: 'light' | 'dark' | 'system') {
    this.theme = theme;
    localStorage.setItem('theme', theme);
    this.apply();
  }
  
  apply() {
    const root = document.documentElement;
    
    if (this.theme === 'system') {
      const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      root.classList.toggle('dark', isDark);
    } else {
      root.classList.toggle('dark', this.theme === 'dark');
    }
  }
  
  getTheme() {
    return this.theme;
  }
}

// Usage
const themeManager = new ThemeManager();

// Toggle theme
document.getElementById('theme-toggle')?.addEventListener('click', () => {
  const current = themeManager.getTheme();
  const next = current === 'dark' ? 'light' : 'dark';
  themeManager.setTheme(next);
});
```

### 10.5 Dark Mode Considerations

#### **Avoid Pure Black**

```css
/* ❌ Bad: Pure black is harsh */
.dark {
  --background: 0 0% 0%;
}

/* ✅ Good: Slightly lighter for better readability */
.dark {
  --background: 222.2 84% 4.9%;  /* ~#0a0a0f */
}
```

#### **Reduce White Contrast**

```css
/* ❌ Bad: Pure white on dark */
.dark {
  --text: 0 0% 100%;
}

/* ✅ Good: Slightly dimmed white */
.dark {
  --text: 210 40% 98%;  /* ~#f9fafb */
}
```

#### **Adjust Shadows**

```css
/* Light mode shadows */
.card {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

/* Dark mode shadows */
.dark .card {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
  /* Stronger shadows for depth */
}
```

#### **Elevation System**

```css
:root {
  --elevation-1: hsl(var(--background));
  --elevation-2: hsl(var(--muted) / 0.3);
  --elevation-3: hsl(var(--muted) / 0.5);
}

.dark {
  --elevation-1: hsl(222.2 84% 4.9%);    /* Base */
  --elevation-2: hsl(222.2 84% 6%);      /* Card */
  --elevation-3: hsl(222.2 84% 8%);      /* Modal */
}
```

### 10.6 Image Handling

```css
/* Reduce brightness of images in dark mode */
.dark img {
  opacity: 0.9;
}

/* Invert logos/icons */
.dark .logo {
  filter: invert(1) hue-rotate(180deg);
}

/* Use different images for dark mode */
<picture>
  <source srcset="logo-dark.svg" media="(prefers-color-scheme: dark)" />
  <img src="logo-light.svg" alt="Logo" />
</picture>
```

### 10.7 Color Semantic Consistency

Ensure colors maintain meaning across themes:

```css
/* Success color */
:root {
  --success: 142.1 76.2% 36.3%;  /* Green in light */
}

.dark {
  --success: 142.1 70.6% 45.3%;  /* Lighter green in dark */
}

/* Error color */
:root {
  --error: 0 84.2% 60.2%;  /* Red in light */
}

.dark {
  --error: 0 62.8% 50.6%;  /* Adjusted red in dark */
}
```

### 10.8 Testing Dark Mode

**Checklist:**
- [ ] All text meets contrast requirements (4.5:1)
- [ ] Interactive elements are visible
- [ ] Focus indicators are visible
- [ ] Shadows provide adequate depth
- [ ] Images don't look washed out
- [ ] Color meanings are preserved
- [ ] Test in actual dark environments
- [ ] Verify on OLED displays

---

## 11. 🎨 Theming Strategy

### 11.1 Theming Philosophy

Theming enables customization while maintaining consistency. A good theming system:

1. **Supports brand identity** — Easy to apply brand colors
2. **Maintains accessibility** — Enforces contrast requirements
3. **Scales effortlessly** — Works across all components
4. **Enables experimentation** — Quick theme switching
5. **Preserves semantics** — Color meanings remain clear

### 11.2 Theme Architecture

```typescript
interface Theme {
  id: string;
  name: string;
  colors: ColorPalette;
  typography: Typography;
  spacing: SpacingScale;
  radius: RadiusScale;
  shadows: ShadowScale;
}

interface ColorPalette {
  // Base colors
  background: string;
  foreground: string;
  
  // Component colors
  card: string;
  cardForeground: string;
  popover: string;
  popoverForeground: string;
  
  // Brand colors
  primary: string;
  primaryForeground: string;
  secondary: string;
  secondaryForeground: string;
  
  // Semantic colors
  muted: string;
  mutedForeground: string;
  accent: string;
  accentForeground: string;
  destructive: string;
  destructiveForeground: string;
  
  // UI elements
  border: string;
  input: string;
  ring: string;
}
```

### 11.3 Pre-built Themes

```typescript
// Default theme
const defaultTheme: Theme = {
  id: 'default',
  name: 'Default',
  colors: {
    background: '0 0% 100%',
    foreground: '222.2 84% 4.9%',
    primary: '222.2 47.4% 11.2%',
    primaryForeground: '210 40% 98%',
    // ... rest of colors
  }
};

// Slate theme
const slateTheme: Theme = {
  id: 'slate',
  name: 'Slate',
  colors: {
    background: '0 0% 100%',
    foreground: '222.2 84% 4.9%',
    primary: '215.4 16.3% 46.9%',
    primaryForeground: '210 40% 98%',
    // ... rest of colors
  }
};

// Blue theme
const blueTheme: Theme = {
  id: 'blue',
  name: 'Blue',
  colors: {
    background: '0 0% 100%',
    foreground: '222.2 84% 4.9%',
    primary: '221.2 83.2% 53.3%',
    primaryForeground: '210 40% 98%',
    // ... rest of colors
  }
};
```

### 11.4 Theme Application

```typescript
class ThemeSystem {
  private currentTheme: Theme;
  
  applyTheme(theme: Theme) {
    this.currentTheme = theme;
    const root = document.documentElement;
    
    // Apply colors
    Object.entries(theme.colors).forEach(([key, value]) => {
      const cssVar = `--${this.camelToKebab(key)}`;
      root.style.setProperty(cssVar, value);
    });
    
    // Apply typography
    Object.entries(theme.typography).forEach(([key, value]) => {
      const cssVar = `--font-${this.camelToKebab(key)}`;
      root.style.setProperty(cssVar, value);
    });
    
    // Save preference
    localStorage.setItem('theme-id', theme.id);
  }
  
  private camelToKebab(str: string): string {
    return str.replace(/[A-Z]/g, letter => `-${letter.toLowerCase()}`);
  }
}
```

### 11.5 Theme Generator

```typescript
// Generate theme from primary color
function generateTheme(primaryColor: string): Theme {
  // Parse HSL values
  const [h, s, l] = parseHSL(primaryColor);
  
  return {
    id: `custom-${Date.now()}`,
    name: 'Custom Theme',
    colors: {
      background: '0 0% 100%',
      foreground: '222.2 84% 4.9%',
      
      // Generate variations of primary
      primary: `${h} ${s}% ${l}%`,
      primaryForeground: `${h} ${s}% ${l > 50 ? 10 : 98}%`,
      
      // Generate complementary colors
      secondary: `${h} ${s * 0.5}% ${l + 10}%`,
      accent: `${(h + 30) % 360} ${s}% ${l}%`,
      
      // ... generate rest of palette
    }
  };
}
```

### 11.6 Brand Theme Example

```typescript
// Company brand theme
const companyTheme: Theme = {
  id: 'acme-corp',
  name: 'ACME Corp',
  colors: {
    // Brand colors from style guide
    primary: '210 100% 50%',        // Brand Blue
    primaryForeground: '0 0% 100%',
    
    secondary: '150 80% 40%',       // Brand Green
    secondaryForeground: '0 0% 100%',
    
    accent: '30 100% 50%',          // Accent Orange
    accentForeground: '0 0% 100%',
    
    // Neutral colors
    background: '0 0% 100%',
    foreground: '210 15% 20%',
    muted: '210 15% 96%',
    mutedForeground: '210 15% 45%',
    
    // Semantic colors
    destructive: '0 80% 50%',
    destructiveForeground: '0 0% 100%',
    
    // UI elements
    border: '210 15% 90%',
    input: '210 15% 95%',
    ring: '210 100% 50%',
  },
  typography: {
    sans: '"Roboto", sans-serif',
    mono: '"Roboto Mono", monospace',
  },
  radius: {
    sm: '0.25rem',
    md: '0.5rem',
    lg: '0.75rem',
  }
};
```

### 11.7 Theme Validation

```typescript
function validateTheme(theme: Theme): ValidationResult {
  const errors: string[] = [];
  
  // Check contrast ratios
  const bgFgContrast = calculateContrast(
    theme.colors.background,
    theme.colors.foreground
  );
  
  if (bgFgContrast < 4.5) {
    errors.push('Foreground/background contrast too low (< 4.5:1)');
  }
  
  const primaryContrast = calculateContrast(
    theme.colors.primary,
    theme.colors.primaryForeground
  );
  
  if (primaryContrast < 4.5) {
    errors.push('Primary color contrast too low');
  }
  
  // Check all required properties
  const requiredColors = [
    'background', 'foreground', 'primary', 'primaryForeground'
  ];
  
  for (const color of requiredColors) {
    if (!theme.colors[color]) {
      errors.push(`Missing required color: ${color}`);
    }
  }
  
  return {
    valid: errors.length === 0,
    errors
  };
}
```

### 11.8 Theme Switcher UI

```typescript
// Theme picker component
<div class="theme-picker">
  <button 
    data-theme="default"
    class="theme-option"
    aria-label="Default theme"
  >
    <div class="theme-preview">
      <span style="background: hsl(222.2 47.4% 11.2%)"></span>
      <span style="background: hsl(210 40% 96.1%)"></span>
      <span style="background: hsl(0 84.2% 60.2%)"></span>
    </div>
    <span>Default</span>
  </button>
  
  <button 
    data-theme="slate"
    class="theme-option"
    aria-label="Slate theme"
  >
    <div class="theme-preview">
      <span style="background: hsl(215.4 16.3% 46.9%)"></span>
      <span style="background: hsl(210 40% 96.1%)"></span>
      <span style="background: hsl(0 84.2% 60.2%)"></span>
    </div>
    <span>Slate</span>
  </button>
  
  <!-- More themes... -->
</div>

<style>
.theme-picker {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: var(--space-4);
}

.theme-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-4);
  border: 2px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
}

.theme-option[data-active="true"] {
  border-color: var(--primary);
}

.theme-preview {
  display: flex;
  gap: var(--space-1);
}

.theme-preview span {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-sm);
}
</style>
```

---

## 12. 🔄 Component Patterns

### 12.1 Pattern Philosophy

Component patterns are proven solutions to common UI problems. They provide:

1. **Consistency** — Same problems solved the same way
2. **Efficiency** — Don't reinvent the wheel
3. **Usability** — Familiar patterns users understand
4. **Accessibility** — Built-in a11y best practices

### 12.2 Data Display Patterns

#### **Card Pattern**

```typescript
// Basic card structure
<Card>
  <CardHeader>
    <CardTitle>Card Title</CardTitle>
    <CardDescription>Supporting description text</CardDescription>
  </CardHeader>
  <CardContent>
    Main content goes here
  </CardContent>
  <CardFooter>
    <Button>Action</Button>
  </CardFooter>
</Card>

// Card variations
<Card variant="outline">      <!-- Outlined card -->
<Card variant="filled">       <!-- Filled background -->
<Card variant="elevated">     <!-- With shadow -->
<Card interactive>            <!-- Hoverable/clickable -->
```

**CSS Implementation:**

```css
.card {
  border-radius: var(--radius-lg);
  background: hsl(var(--card));
  color: hsl(var(--card-foreground));
  box-shadow: var(--shadow-sm);
}

.card-outline {
  border: 1px solid hsl(var(--border));
}

.card-interactive {
  cursor: pointer;
  transition: box-shadow var(--duration-normal);
}

.card-interactive:hover {
  box-shadow: var(--shadow-md);
}
```

#### **List Pattern**

```typescript
// Simple list
<List>
  <ListItem>Item 1</ListItem>
  <ListItem>Item 2</ListItem>
  <ListItem>Item 3</ListItem>
</List>

// Interactive list
<List>
  <ListItem interactive onClick={handler}>
    <ListItemIcon><Icon name="user" /></ListItemIcon>
    <ListItemText>
      <ListItemTitle>John Doe</ListItemTitle>
      <ListItemDescription>john@example.com</ListItemDescription>
    </ListItemText>
    <ListItemAction><Icon name="chevron-right" /></ListItemAction>
  </ListItem>
</List>
```

#### **Table Pattern**

```typescript
<Table>
  <TableHeader>
    <TableRow>
      <TableHead>Name</TableHead>
      <TableHead>Email</TableHead>
      <TableHead>Role</TableHead>
      <TableHead>Actions</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    <TableRow>
      <TableCell>John Doe</TableCell>
      <TableCell>john@example.com</TableCell>
      <TableCell>Admin</TableCell>
      <TableCell>
        <Button size="sm" variant="ghost">Edit</Button>
      </TableCell>
    </TableRow>
  </TableBody>
</Table>
```

### 12.3 Navigation Patterns

#### **Sidebar Navigation**

```typescript
<Sidebar>
  <SidebarHeader>
    <Logo />
  </SidebarHeader>
  <SidebarContent>
    <SidebarGroup>
      <SidebarGroupLabel>Main</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarItem href="/dashboard" active>
          <Icon name="home" />
          Dashboard
        </SidebarItem>
        <SidebarItem href="/projects">
          <Icon name="folder" />
          Projects
        </SidebarItem>
      </SidebarGroupContent>
    </SidebarGroup>
  </SidebarContent>
  <SidebarFooter>
    <UserMenu />
  </SidebarFooter>
</Sidebar>
```

#### **Tabs Pattern**

```typescript
<Tabs defaultValue="account">
  <TabsList>
    <TabsTrigger value="account">Account</TabsTrigger>
    <TabsTrigger value="password">Password</TabsTrigger>
    <TabsTrigger value="notifications">Notifications</TabsTrigger>
  </TabsList>
  <TabsContent value="account">
    Account settings content
  </TabsContent>
  <TabsContent value="password">
    Password settings content
  </TabsContent>
  <TabsContent value="notifications">
    Notification settings content
  </TabsContent>
</Tabs>
```

#### **Breadcrumb Pattern**

```typescript
<Breadcrumb>
  <BreadcrumbList>
    <BreadcrumbItem>
      <BreadcrumbLink href="/">Home</BreadcrumbLink>
    </BreadcrumbItem>
    <BreadcrumbSeparator />
    <BreadcrumbItem>
      <BreadcrumbLink href="/products">Products</BreadcrumbLink>
    </BreadcrumbItem>
    <BreadcrumbSeparator />
    <BreadcrumbItem>
      <BreadcrumbPage>Current Product</BreadcrumbPage>
    </BreadcrumbItem>
  </BreadcrumbList>
</Breadcrumb>
```

### 12.4 Feedback Patterns

#### **Toast/Notification Pattern**

```typescript
// Toast notification
toast({
  title: "Success!",
  description: "Your changes have been saved.",
  variant: "success",
  duration: 3000,
});

// Different variants
toast({ variant: "success" });    // Green checkmark
toast({ variant: "error" });      // Red error
toast({ variant: "warning" });    // Yellow warning
toast({ variant: "info" });       // Blue info
```

#### **Alert Pattern**

```typescript
<Alert variant="success">
  <AlertIcon><Icon name="check-circle" /></AlertIcon>
  <AlertTitle>Success</AlertTitle>
  <AlertDescription>
    Your profile has been updated successfully.
  </AlertDescription>
</Alert>

<Alert variant="destructive">
  <AlertIcon><Icon name="alert-circle" /></AlertIcon>
  <AlertTitle>Error</AlertTitle>
  <AlertDescription>
    Unable to save changes. Please try again.
  </AlertDescription>
</Alert>
```

#### **Progress Pattern**

```typescript
// Linear progress
<Progress value={60} max={100} />

// Circular progress
<CircularProgress value={60} />

// Indeterminate progress
<Progress indeterminate />

// With label
<Progress value={60} showLabel />
```

### 12.5 Overlay Patterns

#### **Modal/Dialog Pattern**

```typescript
<Dialog>
  <DialogTrigger asChild>
    <Button>Open Dialog</Button>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Confirm Action</DialogTitle>
      <DialogDescription>
        Are you sure you want to proceed?
      </DialogDescription>
    </DialogHeader>
    <DialogBody>
      <!-- Main content -->
    </DialogBody>
    <DialogFooter>
      <Button variant="outline">Cancel</Button>
      <Button>Confirm</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

#### **Dropdown Pattern**

```typescript
<DropdownMenu>
  <DropdownMenuTrigger asChild>
    <Button variant="outline">
      Options
      <Icon name="chevron-down" />
    </Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent>
    <DropdownMenuLabel>My Account</DropdownMenuLabel>
    <DropdownMenuSeparator />
    <DropdownMenuItem>
      <Icon name="user" />
      Profile
    </DropdownMenuItem>
    <DropdownMenuItem>
      <Icon name="settings" />
      Settings
    </DropdownMenuItem>
    <DropdownMenuSeparator />
    <DropdownMenuItem destructive>
      <Icon name="log-out" />
      Logout
    </DropdownMenuItem>
  </DropdownMenuContent>
</DropdownMenu>
```

#### **Popover Pattern**

```typescript
<Popover>
  <PopoverTrigger asChild>
    <Button variant="ghost">
      <Icon name="info" />
    </Button>
  </PopoverTrigger>
  <PopoverContent>
    <PopoverHeader>Information</PopoverHeader>
    <PopoverBody>
      Additional context or help text goes here.
    </PopoverBody>
  </PopoverContent>
</Popover>
```

### 12.6 Input Patterns

#### **Form Field Pattern**

```typescript
<FormField>
  <FormLabel htmlFor="email">
    Email
    <FormRequired>*</FormRequired>
  </FormLabel>
  <FormControl>
    <Input 
      id="email"
      type="email"
      placeholder="you@example.com"
    />
  </FormControl>
  <FormDescription>
    We'll never share your email.
  </FormDescription>
  <FormMessage>
    <!-- Error message appears here -->
  </FormMessage>
</FormField>
```

#### **Search Pattern**

```typescript
<SearchBox>
  <SearchInput 
    placeholder="Search..." 
    value={query}
    onChange={handleSearch}
  />
  <SearchButton>
    <Icon name="search" />
  </SearchButton>
  {query && (
    <SearchClear onClick={clearSearch}>
      <Icon name="x" />
    </SearchClear>
  )}
</SearchBox>

<!-- With results -->
<SearchResults>
  <SearchResultItem>Result 1</SearchResultItem>
  <SearchResultItem>Result 2</SearchResultItem>
</SearchResults>
```

#### **Date Picker Pattern**

```typescript
<DatePicker>
  <DatePickerTrigger>
    <Input 
      value={date ? formatDate(date) : ""}
      placeholder="Select date"
      readOnly
    />
    <Icon name="calendar" />
  </DatePickerTrigger>
  <DatePickerContent>
    <Calendar 
      selected={date}
      onSelect={setDate}
    />
  </DatePickerContent>
</DatePicker>
```

### 12.7 Empty States

```typescript
<EmptyState>
  <EmptyStateIcon>
    <Icon name="inbox" size="xl" />
  </EmptyStateIcon>
  <EmptyStateTitle>No messages yet</EmptyStateTitle>
  <EmptyStateDescription>
    When you receive messages, they'll appear here.
  </EmptyStateDescription>
  <EmptyStateAction>
    <Button>Send your first message</Button>
  </EmptyStateAction>
</EmptyState>
```

### 12.8 Loading States

```typescript
// Skeleton loading
<Card>
  <CardHeader>
    <Skeleton className="h-4 w-[250px]" />
    <Skeleton className="h-3 w-[200px]" />
  </CardHeader>
  <CardContent>
    <Skeleton className="h-[200px]" />
  </CardContent>
</Card>

// Spinner loading
<div className="loading-container">
  <Spinner size="lg" />
  <p>Loading content...</p>
</div>

// Progress loading
<ProgressBar value={progress} />
```

---

## 13. 🧱 Composition Principles

### 13.1 Composition Philosophy

Composition over inheritance. Build complex UIs by combining simple components rather than creating specialized variants.

```
❌ Bad: Create specific variants
<ButtonWithIcon />
<ButtonWithBadge />
<ButtonWithIconAndBadge />

✅ Good: Compose elements
<Button>
  <Icon />
  <span>Text</span>
  <Badge />
</Button>
```

### 13.2 Slot-Based Composition

```typescript
// Component with named slots
interface CardProps {
  children: React.ReactNode;
}

function Card({ children }: CardProps) {
  const slots = {
    header: null,
    content: null,
    footer: null
  };
  
  React.Children.forEach(children, (child) => {
    if (React.isValidElement(child)) {
      if (child.type === CardHeader) slots.header = child;
      if (child.type === CardContent) slots.content = child;
      if (child.type === CardFooter) slots.footer = child;
    }
  });
  
  return (
    <div className="card">
      {slots.header}
      {slots.content}
      {slots.footer}
    </div>
  );
}

// Usage
<Card>
  <CardHeader>Header content</CardHeader>
  <CardContent>Main content</CardContent>
  <CardFooter>Footer actions</CardFooter>
</Card>
```

### 13.3 Render Props Pattern

```typescript
interface DataListProps<T> {
  data: T[];
  renderItem: (item: T, index: number) => React.ReactNode;
  renderEmpty?: () => React.ReactNode;
  renderLoading?: () => React.ReactNode;
  loading?: boolean;
}

function DataList<T>({ 
  data, 
  renderItem, 
  renderEmpty,
  renderLoading,
  loading 
}: DataListProps<T>) {
  if (loading && renderLoading) {
    return <>{renderLoading()}</>;
  }
  
  if (data.length === 0 && renderEmpty) {
    return <>{renderEmpty()}</>;
  }
  
  return (
    <div className="data-list">
      {data.map((item, index) => renderItem(item, index))}
    </div>
  );
}

// Usage
<DataList
  data={users}
  renderItem={(user) => (
    <UserCard key={user.id} user={user} />
  )}
  renderEmpty={() => <EmptyState />}
  renderLoading={() => <Skeleton />}
  loading={isLoading}
/>
```

### 13.4 Compound Components

```typescript
// Tabs compound component
interface TabsContextValue {
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

const TabsContext = React.createContext<TabsContextValue | null>(null);

function Tabs({ defaultValue, children }: TabsProps) {
  const [activeTab, setActiveTab] = React.useState(defaultValue);
  
  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      <div className="tabs">{children}</div>
    </TabsContext.Provider>
  );
}

function TabsList({ children }: { children: React.ReactNode }) {
  return <div className="tabs-list">{children}</div>;
}

function TabsTrigger({ value, children }: TabsTriggerProps) {
  const context = React.useContext(TabsContext);
  const isActive = context?.activeTab === value;
  
  return (
    <button
      className={`tabs-trigger ${isActive ? 'active' : ''}`}
      onClick={() => context?.setActiveTab(value)}
    >
      {children}
    </button>
  );
}

function TabsContent({ value, children }: TabsContentProps) {
  const context = React.useContext(TabsContext);
  if (context?.activeTab !== value) return null;
  
  return <div className="tabs-content">{children}</div>;
}

// Export as compound
Tabs.List = TabsList;
Tabs.Trigger = TabsTrigger;
Tabs.Content = TabsContent;

// Usage
<Tabs defaultValue="tab1">
  <Tabs.List>
    <Tabs.Trigger value="tab1">Tab 1</Tabs.Trigger>
    <Tabs.Trigger value="tab2">Tab 2</Tabs.Trigger>
  </Tabs.List>
  <Tabs.Content value="tab1">Content 1</Tabs.Content>
  <Tabs.Content value="tab2">Content 2</Tabs.Content>
</Tabs>
```

### 13.5 Polymorphic Components

```typescript
// Component that can render as different elements
type PolymorphicProps<E extends React.ElementType> = {
  as?: E;
  children: React.ReactNode;
} & React.ComponentPropsWithoutRef<E>;

function Box<E extends React.ElementType = 'div'>({
  as,
  children,
  ...props
}: PolymorphicProps<E>) {
  const Component = as || 'div';
  return <Component {...props}>{children}</Component>;
}

// Usage
<Box>Renders as div</Box>
<Box as="section">Renders as section</Box>
<Box as="a" href="/">Renders as link</Box>
<Box as={CustomComponent}>Renders as custom component</Box>
```

### 13.6 Layout Composition

```typescript
// Stack layout
<Stack direction="vertical" spacing={4}>
  <Box>Item 1</Box>
  <Box>Item 2</Box>
  <Box>Item 3</Box>
</Stack>

// Grid layout
<Grid cols={3} gap={4}>
  <Box>1</Box>
  <Box>2</Box>
  <Box>3</Box>
</Grid>

// Flex layout
<Flex justify="between" align="center">
  <Box>Left</Box>
  <Box>Right</Box>
</Flex>
```

### 13.7 Controlled vs Uncontrolled

```typescript
// Controlled component
function ControlledInput() {
  const [value, setValue] = React.useState('');
  
  return (
    <Input 
      value={value}
      onChange={(e) => setValue(e.target.value)}
    />
  );
}

// Uncontrolled component
function UncontrolledInput() {
  const inputRef = React.useRef<HTMLInputElement>(null);
  
  const handleSubmit = () => {
    console.log(inputRef.current?.value);
  };
  
  return <Input ref={inputRef} defaultValue="" />;
}

// Hybrid: Support both patterns
interface InputProps {
  value?: string;
  defaultValue?: string;
  onChange?: (value: string) => void;
}

function Input({ value, defaultValue, onChange }: InputProps) {
  const [internalValue, setInternalValue] = React.useState(defaultValue);
  const isControlled = value !== undefined;
  const currentValue = isControlled ? value : internalValue;
  
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    if (!isControlled) {
      setInternalValue(newValue);
    }
    onChange?.(newValue);
  };
  
  return (
    <input 
      value={currentValue}
      onChange={handleChange}
    />
  );
}
```

---

## 14. 📊 State Management

### 14.1 State Management Philosophy

State should live at the appropriate level:

1. **Local state** — Component-specific (useState, useReducer)
2. **Lifted state** — Shared between siblings (props)
3. **Context state** — Subtree-wide (Context API)
4. **Global state** — App-wide (Redux, Zustand, etc.)

### 14.2 Local State Patterns

```typescript
// Simple state
function Counter() {
  const [count, setCount] = React.useState(0);
  
  return (
    <div>
      <p>Count: {count}</p>
      <Button onClick={() => setCount(count + 1)}>Increment</Button>
    </div>
  );
}

// Complex state with reducer
interface State {
  items: Item[];
  filter: string;
  sortBy: string;
}

type Action = 
  | { type: 'ADD_ITEM'; payload: Item }
  | { type: 'REMOVE_ITEM'; payload: string }
  | { type: 'SET_FILTER'; payload: string }
  | { type: 'SET_SORT'; payload: string };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'ADD_ITEM':
      return { ...state, items: [...state.items, action.payload] };
    case 'REMOVE_ITEM':
      return { 
        ...state, 
        items: state.items.filter(item => item.id !== action.payload) 
      };
    case 'SET_FILTER':
      return { ...state, filter: action.payload };
    case 'SET_SORT':
      return { ...state, sortBy: action.payload };
    default:
      return state;
  }
}

function ItemList() {
  const [state, dispatch] = React.useReducer(reducer, {
    items: [],
    filter: '',
    sortBy: 'name'
  });
  
  // Use state and dispatch
}
```

### 14.3 Context State

```typescript
// Theme context
interface ThemeContextValue {
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;
}

const ThemeContext = React.createContext<ThemeContextValue | null>(null);

function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = React.useState<'light' | 'dark'>('light');
  
  React.useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }, [theme]);
  
  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

function useTheme() {
  const context = React.useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
}

// Usage
function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  
  return (
    <Button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>
      {theme === 'light' ? <Icon name="moon" /> : <Icon name="sun" />}
    </Button>
  );
}
```

### 14.4 Form State

```typescript
// Form state management
interface FormState {
  values: Record<string, any>;
  errors: Record<string, string>;
  touched: Record<string, boolean>;
  isSubmitting: boolean;
}

function useForm<T extends Record<string, any>>(
  initialValues: T,
  validate: (values: T) => Record<string, string>
) {
  const [values, setValues] = React.useState<T>(initialValues);
  const [errors, setErrors] = React.useState<Record<string, string>>({});
  const [touched, setTouched] = React.useState<Record<string, boolean>>({});
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  
  const handleChange = (name: string, value: any) => {
    setValues(prev => ({ ...prev, [name]: value }));
  };
  
  const handleBlur = (name: string) => {
    setTouched(prev => ({ ...prev, [name]: true }));
    const fieldErrors = validate(values);
    setErrors(fieldErrors);
  };
  
  const handleSubmit = async (
    onSubmit: (values: T) => Promise<void>
  ) => {
    setIsSubmitting(true);
    const validationErrors = validate(values);
    
    if (Object.keys(validationErrors).length === 0) {
      try {
        await onSubmit(values);
      } catch (error) {
        console.error(error);
      }
    } else {
      setErrors(validationErrors);
    }
    
    setIsSubmitting(false);
  };
  
  return {
    values,
    errors,
    touched,
    isSubmitting,
    handleChange,
    handleBlur,
    handleSubmit,
  };
}

// Usage
function LoginForm() {
  const form = useForm(
    { email: '', password: '' },
    (values) => {
      const errors: Record<string, string> = {};
      if (!values.email) errors.email = 'Email is required';
      if (!values.password) errors.password = 'Password is required';
      return errors;
    }
  );
  
  return (
    <form onSubmit={(e) => {
      e.preventDefault();
      form.handleSubmit(async (values) => {
        await login(values);
      });
    }}>
      <FormField>
        <FormLabel>Email</FormLabel>
        <Input
          value={form.values.email}
          onChange={(e) => form.handleChange('email', e.target.value)}
          onBlur={() => form.handleBlur('email')}
        />
        {form.touched.email && form.errors.email && (
          <FormMessage>{form.errors.email}</FormMessage>
        )}
      </FormField>
      
      <Button type="submit" disabled={form.isSubmitting}>
        {form.isSubmitting ? 'Logging in...' : 'Login'}
      </Button>
    </form>
  );
}
```

### 14.5 Async State

```typescript
// Async state hook
interface AsyncState<T> {
  data: T | null;
  error: Error | null;
  loading: boolean;
}

function useAsync<T>(
  asyncFunction: () => Promise<T>,
  dependencies: any[] = []
) {
  const [state, setState] = React.useState<AsyncState<T>>({
    data: null,
    error: null,
    loading: true,
  });
  
  React.useEffect(() => {
    let cancelled = false;
    
    setState({ data: null, error: null, loading: true });
    
    asyncFunction()
      .then(data => {
        if (!cancelled) {
          setState({ data, error: null, loading: false });
        }
      })
      .catch(error => {
        if (!cancelled) {
          setState({ data: null, error, loading: false });
        }
      });
    
    return () => {
      cancelled = true;
    };
  }, dependencies);
  
  return state;
}

// Usage
function UserProfile({ userId }: { userId: string }) {
  const { data: user, loading, error } = useAsync(
    () => fetchUser(userId),
    [userId]
  );
  
  if (loading) return <Skeleton />;
  if (error) return <ErrorMessage error={error} />;
  if (!user) return <EmptyState />;
  
  return <div>{user.name}</div>;
}
```

---

## 15. 🎨 Icon System

### 15.1 Icon Philosophy

Icons enhance UX by providing visual cues and saving space. A good icon system:

1. **Consistent style** — Same design language
2. **Clear meaning** — Immediately recognizable
3. **Scalable** — Works at multiple sizes
4. **Accessible** — Proper labels and fallbacks
5. **Performant** — Optimized delivery

### 15.2 Icon Implementation

```typescript
// Icon component
interface IconProps {
  name: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
  'aria-label'?: string;
}

function Icon({ name, size = 'md', className, 'aria-label': ariaLabel }: IconProps) {
  const sizeMap = {
    xs: 12,
    sm: 16,
    md: 20,
    lg: 24,
    xl: 32,
  };
  
  return (
    <svg
      className={`icon icon-${size} ${className}`}
      width={sizeMap[size]}
      height={sizeMap[size]}
      aria-label={ariaLabel}
      aria-hidden={!ariaLabel}
    >
      <use href={`/icons.svg#${name}`} />
    </svg>
  );
}

// Usage
<Icon name="check" size="sm" aria-label="Success" />
<Icon name="arrow-right" />
```

### 15.3 Icon Sizes

```css
:root {
  --icon-xs: 12px;
  --icon-sm: 16px;
  --icon-md: 20px;
  --icon-lg: 24px;
  --icon-xl: 32px;
  --icon-2xl: 40px;
}

.icon {
  display: inline-block;
  vertical-align: middle;
  fill: currentColor;
}

.icon-xs { width: var(--icon-xs); height: var(--icon-xs); }
.icon-sm { width: var(--icon-sm); height: var(--icon-sm); }
.icon-md { width: var(--icon-md); height: var(--icon-md); }
.icon-lg { width: var(--icon-lg); height: var(--icon-lg); }
.icon-xl { width: var(--icon-xl); height: var(--icon-xl); }
```

### 15.4 Icon Library Structure

```
icons/
├── actions/
│   ├── add.svg
│   ├── edit.svg
│   ├── delete.svg
│   └── search.svg
├── navigation/
│   ├── arrow-left.svg
│   ├── arrow-right.svg
│   ├── chevron-down.svg
│   └── menu.svg
├── status/
│   ├── check.svg
│   ├── error.svg
│   ├── warning.svg
│   └── info.svg
├── social/
│   ├── twitter.svg
│   ├── github.svg
│   └── linkedin.svg
└── misc/
    ├── user.svg
    ├── settings.svg
    └── help.svg
```

### 15.5 Icon Sprite Generation

```javascript
// Build script to generate sprite
const fs = require('fs');
const path = require('path');
const SVGO = require('svgo');

async function generateSprite() {
  const iconsDir = path.join(__dirname, 'icons');
  const outputPath = path.join(__dirname, 'public/icons.svg');
  
  const svgo = new SVGO({
    plugins: [
      { removeViewBox: false },
      { removeDimensions: true },
    ],
  });
  
  let sprite = '<svg xmlns="http://www.w3.org/2000/svg" style="display:none">';
  
  // Process all SVG files
  const files = getAllSvgFiles(iconsDir);
  
  for (const file of files) {
    const content = fs.readFileSync(file, 'utf-8');
    const optimized = await svgo.optimize(content);
    const id = path.basename(file, '.svg');
    
    // Wrap in symbol
    sprite += `<symbol id="${id}" viewBox="0 0 24 24">${optimized.data}</symbol>`;
  }
  
  sprite += '</svg>';
  
  fs.writeFileSync(outputPath, sprite);
  console.log(`Generated sprite with ${files.length} icons`);
}
```

### 15.6 Animated Icons

```css
/* Spinning icon */
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.icon-spin {
  animation: spin 1s linear infinite;
}

/* Pulsing icon */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.icon-pulse {
  animation: pulse 2s ease-in-out infinite;
}

/* Usage */
<Icon name="loader" className="icon-spin" />
<Icon name="notification" className="icon-pulse" />
```

---

## 16. 📊 Data Visualization

### 16.1 Visualization Philosophy

Data visualization makes complex information understandable. Good visualizations:

1. **Tell a story** — Communicate insights clearly
2. **Maintain accuracy** — Don't distort data
3. **Stay accessible** — Work for all users
4. **Guide attention** — Highlight what matters
5. **Respect context** — Match user's mental model

### 16.2 Chart Color Palette

```css
:root {
  /* Categorical colors (qualitative) */
  --chart-1: 221.2 83.2% 53.3%;   /* Blue */
  --chart-2: 142.1 76.2% 36.3%;   /* Green */
  --chart-3: 38.7 92.0% 50.2%;    /* Orange */
  --chart-4: 280 80% 55%;         /* Purple */
  --chart-5: 0 84.2% 60.2%;       /* Red */
  --chart-6: 199 89% 48%;         /* Cyan */
  --chart-7: 45 93% 47%;          /* Yellow */
  --chart-8: 330 81% 60%;         /* Pink */
  
  /* Sequential colors (quantitative) */
  --chart-seq-1: 221.2 83.2% 95%;
  --chart-seq-2: 221.2 83.2% 85%;
  --chart-seq-3: 221.2 83.2% 70%;
  --chart-seq-4: 221.2 83.2% 53.3%;
  --chart-seq-5: 221.2 83.2% 40%;
  
  /* Diverging colors */
  --chart-negative: 0 84.2% 60.2%;    /* Red */
  --chart-neutral: 210 40% 96.1%;     /* Gray */
  --chart-positive: 142.1 76.2% 36.3%; /* Green */
}
```

### 16.3 Chart Components

```typescript
// Bar Chart
<BarChart data={data} width={600} height={400}>
  <XAxis dataKey="name" />
  <YAxis />
  <Tooltip />
  <Legend />
  <Bar dataKey="value" fill="hsl(var(--chart-1))" />
</BarChart>

// Line Chart
<LineChart data={data} width={600} height={400}>
  <XAxis dataKey="date" />
  <YAxis />
  <Tooltip />
  <Legend />
  <Line 
    type="monotone" 
    dataKey="revenue" 
    stroke="hsl(var(--chart-1))" 
    strokeWidth={2}
  />
</LineChart>

// Pie Chart
<PieChart width={400} height={400}>
  <Pie
    data={data}
    dataKey="value"
    nameKey="name"
    cx="50%"
    cy="50%"
    outerRadius={120}
    label
  />
  <Tooltip />
</PieChart>
```

### 16.4 Data Table Patterns

```typescript
// Sortable table
interface Column<T> {
  key: keyof T;
  header: string;
  sortable?: boolean;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
}

function DataTable<T>({ data, columns }: DataTableProps<T>) {
  const [sortKey, setSortKey] = React.useState<keyof T | null>(null);
  const [sortOrder, setSortOrder] = React.useState<'asc' | 'desc'>('asc');
  
  const sortedData = React.useMemo(() => {
    if (!sortKey) return data;
    
    return [...data].sort((a, b) => {
      const aVal = a[sortKey];
      const bVal = b[sortKey];
      
      if (aVal < bVal) return sortOrder === 'asc' ? -1 : 1;
      if (aVal > bVal) return sortOrder === 'asc' ? 1 : -1;
      return 0;
    });
  }, [data, sortKey, sortOrder]);
  
  const handleSort = (key: keyof T) => {
    if (sortKey === key) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortOrder('asc');
    }
  };
  
  return (
    <Table>
      <TableHeader>
        <TableRow>
          {columns.map((column) => (
            <TableHead
              key={String(column.key)}
              onClick={() => column.sortable && handleSort(column.key)}
              className={column.sortable ? 'sortable' : ''}
            >
              {column.header}
              {sortKey === column.key && (
                <Icon name={sortOrder === 'asc' ? 'arrow-up' : 'arrow-down'} />
              )}
            </TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {sortedData.map((row, index) => (
          <TableRow key={index}>
            {columns.map((column) => (
              <TableCell key={String(column.key)}>
                {column.render 
                  ? column.render(row[column.key], row)
                  : String(row[column.key])
                }
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
```

### 16.5 Accessibility in Visualizations

```typescript
// Accessible chart
<BarChart data={data} width={600} height={400}>
  <title>Monthly Revenue Chart</title>
  <desc>
    Bar chart showing monthly revenue from January to December.
    Revenue ranges from $10,000 to $50,000.
  </desc>
  
  {/* Visual content */}
  
  {/* Text alternative */}
  <text className="sr-only">
    January: $25,000, February: $30,000, March: $28,000...
  </text>
</BarChart>

// Provide data table alternative
<details>
  <summary>View data as table</summary>
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Month</TableHead>
        <TableHead>Revenue</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {data.map((item) => (
        <TableRow key={item.month}>
          <TableCell>{item.month}</TableCell>
          <TableCell>${item.revenue}</TableCell>
        </TableRow>
      ))}
    </TableBody>
  </Table>
</details>
```

---

## 17. 📝 Forms & Validation

### 17.1 Form Design Principles

1. **Clear labels** — Every field needs a label
2. **Smart defaults** — Pre-fill when possible
3. **Inline validation** — Provide immediate feedback
4. **Error prevention** — Guide users to success
5. **Progressive disclosure** — Show complexity gradually

### 17.2 Form Field Components

```typescript
// Text input
<FormField>
  <FormLabel htmlFor="name">Name</FormLabel>
  <FormControl>
    <Input 
      id="name"
      placeholder="John Doe"
      required
    />
  </FormControl>
  <FormDescription>Your full legal name</FormDescription>
  <FormMessage />
</FormField>

// Select
<FormField>
  <FormLabel>Country</FormLabel>
  <FormControl>
    <Select>
      <SelectTrigger>
        <SelectValue placeholder="Select country" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="us">United States</SelectItem>
        <SelectItem value="uk">United Kingdom</SelectItem>
        <SelectItem value="ca">Canada</SelectItem>
      </SelectContent>
    </Select>
  </FormControl>
</FormField>

// Checkbox
<FormField>
  <div className="flex items-center gap-2">
    <FormControl>
      <Checkbox id="terms" />
    </FormControl>
    <FormLabel htmlFor="terms">
      I agree to the terms and conditions
    </FormLabel>
  </div>
</FormField>

// Radio group
<FormField>
  <FormLabel>Notification preference</FormLabel>
  <FormControl>
    <RadioGroup defaultValue="email">
      <div className="flex items-center gap-2">
        <RadioGroupItem value="email" id="email" />
        <Label htmlFor="email">Email</Label>
      </div>
      <div className="flex items-center gap-2">
        <RadioGroupItem value="sms" id="sms" />
        <Label htmlFor="sms">SMS</Label>
      </div>
    </RadioGroup>
  </FormControl>
</FormField>
```

### 17.3 Validation Patterns

```typescript
// Validation rules
interface ValidationRule {
  required?: boolean | string;
  minLength?: { value: number; message: string };
  maxLength?: { value: number; message: string };
  pattern?: { value: RegExp; message: string };
  validate?: (value: any) => boolean | string;
}

// Common validators
const validators = {
  required: (message = 'This field is required') => ({
    required: message,
  }),
  
  email: () => ({
    pattern: {
      value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
      message: 'Invalid email address',
    },
  }),
  
  minLength: (length: number) => ({
    minLength: {
      value: length,
      message: `Must be at least ${length} characters`,
    },
  }),
  
  maxLength: (length: number) => ({
    maxLength: {
      value: length,
      message: `Must be no more than ${length} characters`,
    },
  }),
  
  password: () => ({
    minLength: { value: 8, message: 'Password must be at least 8 characters' },
    pattern: {
      value: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
      message: 'Password must contain uppercase, lowercase, and number',
    },
  }),
};

// Usage
<FormField
  name="email"
  rules={{
    ...validators.required(),
    ...validators.email(),
  }}
>
  <Input type="email" />
</FormField>
```

### 17.4 Form Submission

```typescript
function ContactForm() {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [submitError, setSubmitError] = React.useState<string | null>(null);
  const [submitSuccess, setSubmitSuccess] = React.useState(false);
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setSubmitError(null);
    
    const formData = new FormData(e.target as HTMLFormElement);
    const data = Object.fromEntries(formData);
    
    try {
      await submitContactForm(data);
      setSubmitSuccess(true);
      (e.target as HTMLFormElement).reset();
    } catch (error) {
      setSubmitError('Failed to submit form. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {submitSuccess && (
        <Alert variant="success">
          <AlertTitle>Success!</AlertTitle>
          <AlertDescription>
            Your message has been sent.
          </AlertDescription>
        </Alert>
      )}
      
      {submitError && (
        <Alert variant="destructive">
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{submitError}</AlertDescription>
        </Alert>
      )}
      
      <FormField>
        <FormLabel>Email</FormLabel>
        <Input name="email" type="email" required />
      </FormField>
      
      <FormField>
        <FormLabel>Message</FormLabel>
        <Textarea name="message" required />
      </FormField>
      
      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? (
          <>
            <Icon name="loader" className="icon-spin" />
            Sending...
          </>
        ) : (
          'Send Message'
        )}
      </Button>
    </form>
  );
}
```

### 17.5 Multi-Step Forms

```typescript
function MultiStepForm() {
  const [step, setStep] = React.useState(1);
  const [formData, setFormData] = React.useState({});
  
  const updateFormData = (data: Record<string, any>) => {
    setFormData(prev => ({ ...prev, ...data }));
  };
  
  const nextStep = () => setStep(prev => prev + 1);
  const prevStep = () => setStep(prev => prev - 1);
  
  return (
    <div>
      <StepIndicator currentStep={step} totalSteps={3} />
      
      {step === 1 && (
        <Step1 data={formData} onNext={(data) => {
          updateFormData(data);
          nextStep();
        }} />
      )}
      
      {step === 2 && (
        <Step2 
          data={formData} 
          onNext={(data) => {
            updateFormData(data);
            nextStep();
          }}
          onBack={prevStep}
        />
      )}
      
      {step === 3 && (
        <Step3 
          data={formData}
          onSubmit={async (data) => {
            updateFormData(data);
            await submitForm({ ...formData, ...data });
          }}
          onBack={prevStep}
        />
      )}
    </div>
  );
}
```

---

## 18. 🏗️ Layout Patterns

### 18.1 Common Layouts

#### **App Shell**

```typescript
<div className="app-shell">
  <Header />
  <div className="app-body">
    <Sidebar />
    <main className="app-main">
      <Outlet />
    </main>
  </div>
</div>

<style>
.app-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.app-body {
  display: flex;
  flex: 1;
}

.app-main {
  flex: 1;
  padding: var(--space-6);
  overflow-y: auto;
}
</style>
```

#### **Dashboard Layout**

```typescript
<div className="dashboard">
  <div className="dashboard-header">
    <h1>Dashboard</h1>
    <div className="dashboard-actions">
      <Button>Export</Button>
      <Button variant="primary">New Report</Button>
    </div>
  </div>
  
  <div className="dashboard-stats">
    <StatCard title="Revenue" value="$45,000" />
    <StatCard title="Users" value="1,234" />
    <StatCard title="Growth" value="+12%" />
  </div>
  
  <div className="dashboard-content">
    <Card>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
      </CardHeader>
      <CardContent>
        {/* Content */}
      </CardContent>
    </Card>
  </div>
</div>
```

#### **Split View**

```css
.split-view {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: var(--space-6);
  height: 100%;
}

.split-view-sidebar {
  border-right: 1px solid var(--border);
  padding: var(--space-6);
  overflow-y: auto;
}

.split-view-main {
  padding: var(--space-6);
  overflow-y: auto;
}

/* Responsive */
@media (max-width: 768px) {
  .split-view {
    grid-template-columns: 1fr;
  }
  
  .split-view-sidebar {
    border-right: none;
    border-bottom: 1px solid var(--border);
  }
}
```

### 18.2 Grid Systems

```css
/* 12-column grid */
.grid-12 {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: var(--space-6);
}

.col-span-1  { grid-column: span 1; }
.col-span-2  { grid-column: span 2; }
.col-span-3  { grid-column: span 3; }
.col-span-4  { grid-column: span 4; }
.col-span-6  { grid-column: span 6; }
.col-span-8  { grid-column: span 8; }
.col-span-12 { grid-column: span 12; }

/* Responsive columns */
@media (max-width: 768px) {
  .col-span-4,
  .col-span-6,
  .col-span-8 {
    grid-column: span 12;
  }
}
```

### 18.3 Sticky Elements

```css
/* Sticky header */
.sticky-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: hsl(var(--background));
  border-bottom: 1px solid hsl(var(--border));
}

/* Sticky sidebar */
.sticky-sidebar {
  position: sticky;
  top: var(--space-6);
  max-height: calc(100vh - var(--space-12));
  overflow-y: auto;
}
```

---

## 19. ⚡ Performance Guidelines

### 19.1 Performance Philosophy

Performance is a feature. Fast interfaces feel responsive and professional.

**Core Web Vitals:**
- **LCP** (Largest Contentful Paint): < 2.5s
- **FID** (First Input Delay): < 100ms
- **CLS** (Cumulative Layout Shift): < 0.1

### 19.2 CSS Performance

```css
/* ✅ Good: Use transform and opacity */
.element {
  transform: translateX(100px);
  opacity: 0.5;
  transition: transform 200ms, opacity 200ms;
}

/* ❌ Bad: Avoid layout-triggering properties */
.element {
  left: 100px;
  margin-left: 20px;
  width: 500px;
}

/* Use will-change sparingly */
.animating-element {
  will-change: transform;
}

.animating-element.animation-complete {
  will-change: auto;
}

/* Contain layout calculations */
.card {
  contain: layout style paint;
}
```

### 19.3 Image Optimization

```html
<!-- Responsive images -->
<img
  src="image-800w.jpg"
  srcset="
    image-400w.jpg 400w,
    image-800w.jpg 800w,
    image-1200w.jpg 1200w
  "
  sizes="(max-width: 640px) 100vw, 800px"
  alt="Description"
  loading="lazy"
  decoding="async"
/>

<!-- Modern formats with fallback -->
<picture>
  <source type="image/avif" srcset="image.avif" />
  <source type="image/webp" srcset="image.webp" />
  <img src="image.jpg" alt="Description" />
</picture>
```

### 19.4 Code Splitting

```typescript
// Lazy load components
const Dashboard = React.lazy(() => import('./Dashboard'));
const Settings = React.lazy(() => import('./Settings'));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </Suspense>
  );
}

// Dynamic imports
async function loadModule() {
  const module = await import('./heavy-module');
  module.init();
}
```

### 19.5 Memoization

```typescript
// Memo component
const ExpensiveComponent = React.memo(({ data }) => {
  return <div>{/* Render data */}</div>;
});

// useMemo for expensive calculations
function DataList({ items }) {
  const sortedItems = React.useMemo(() => {
    return items.sort((a, b) => a.name.localeCompare(b.name));
  }, [items]);
  
  return <>{/* Render sortedItems */}</>;
}

// useCallback for stable references
function Parent() {
  const handleClick = React.useCallback(() => {
    console.log('Clicked');
  }, []);
  
  return <Child onClick={handleClick} />;
}
```

### 19.6 Bundle Size Optimization

```javascript
// Import only what you need
// ❌ Bad
import _ from 'lodash';

// ✅ Good
import debounce from 'lodash/debounce';

// Tree-shakeable imports
// ✅ Good
import { Button } from '@/components/ui/button';

// Analyze bundle
// package.json
{
  "scripts": {
    "analyze": "ANALYZE=true next build"
  }
}
```

---

## 20. 🚀 Implementation Guide

### 20.1 Setup & Installation

```bash
# Install dependencies
npm install

# Development
npm run dev

# Build for production
npm run build

# Type checking
npm run type-check

# Linting
npm run lint

# Testing
npm run test
```

### 20.2 Project Structure

```
schema-engine/
├── src/
│   ├── components/
│   │   ├── ui/              # Base components
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   └── card.tsx
│   │   ├── forms/           # Form components
│   │   ├── layouts/         # Layout components
│   │   └── patterns/        # Composite patterns
│   ├── styles/
│   │   ├── tokens/          # Design tokens
│   │   ├── globals.css      # Global styles
│   │   └── themes/          # Theme variants
│   ├── hooks/               # Custom hooks
│   ├── utils/               # Utilities
│   └── types/               # TypeScript types
├── public/
│   ├── fonts/
│   └── icons/
├── docs/                    # Documentation
└── examples/                # Usage examples
```

### 20.3 Component Template

```typescript
import * as React from 'react';
import { cn } from '@/lib/utils';

/**
 * Button component props
 */
interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  /**
   * Visual variant
   * @default "default"
   */
  variant?: 'default' | 'primary' | 'destructive' | 'outline' | 'ghost';
  
  /**
   * Size variant
   * @default "md"
   */
  size?: 'sm' | 'md' | 'lg';
  
  /**
   * Loading state
   * @default false
   */
  loading?: boolean;
}

/**
 * Button component for user actions
 * 
 * @example
 * <Button variant="primary" onClick={handleClick}>
 *   Click me
 * </Button>
 */
const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'md', loading, children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          'button',
          `button-${variant}`,
          `button-${size}`,
          loading && 'button-loading',
          className
        )}
        disabled={loading || props.disabled}
        {...props}
      >
        {loading && <Spinner />}
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';

export { Button, type ButtonProps };
```

### 20.4 Testing Strategy

```typescript
// Component test
import { render, screen, fireEvent } from '@testing-library/react';
import { Button } from './button';

describe('Button', () => {
  it('renders children', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });
  
  it('handles click events', () => {
    const handleClick = jest.fn();
    render(<Button onClick={handleClick}>Click me</Button>);
    
    fireEvent.click(screen.getByText('Click me'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });
  
  it('disables when loading', () => {
    render(<Button loading>Click me</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
  });
  
  it('applies variant classes', () => {
    render(<Button variant="primary">Click me</Button>);
    expect(screen.getByRole('button')).toHaveClass('button-primary');
  });
});
```

### 20.5 Documentation Template

```markdown
# Component Name

Brief description of the component and its purpose.

## Usage

\`\`\`tsx
import { Component } from '@/components/ui/component';

function Example() {
  return <Component>Content</Component>;
}
\`\`\`

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| variant | string | "default" | Visual variant |
| size | string | "md" | Size variant |

## Variants

### Default
\`\`\`tsx
<Component variant="default">Default variant</Component>
\`\`\`

### Primary
\`\`\`tsx
<Component variant="primary">Primary variant</Component>
\`\`\`

## Accessibility

- Keyboard navigable
- Screen reader compatible
- ARIA attributes included

## Examples

### Basic Example
\`\`\`tsx
<Component>Basic usage</Component>
\`\`\`

### Advanced Example
\`\`\`tsx
<Component variant="primary" size="lg">
  Advanced usage
</Component>
\`\`\`
```

### 20.6 Contributing Guidelines

```markdown
# Contributing to Schema Engine

## Getting Started

1. Fork the repository
2. Clone your fork
3. Create a feature branch
4. Make your changes
5. Write/update tests
6. Submit a pull request

## Component Checklist

- [ ] Follows design system tokens
- [ ] Includes TypeScript types
- [ ] Has proper accessibility attributes
- [ ] Includes tests (>80% coverage)
- [ ] Documented with examples
- [ ] Responsive design
- [ ] Dark mode support
- [ ] No console errors/warnings

## Code Style

- Use TypeScript
- Follow ESLint rules
- Use Prettier for formatting
- Write meaningful commit messages
```

---

## 📚 Conclusion

This comprehensive design system documentation provides everything needed to build consistent, accessible, and performant user interfaces with the Schema Engine Design System.

### Key Takeaways

1. **Design Tokens** — Foundation for consistency and theming
2. **Accessibility First** — WCAG 2.1 AA compliance minimum
3. **Component Composition** — Build complex UIs from simple parts
4. **Responsive Design** — Mobile-first, progressive enhancement
5. **Performance** — Optimize for Core Web Vitals
6. **Documentation** — Clear examples and guidelines

### Next Steps

1. Review the component specifications
2. Implement base components with design tokens
3. Build composite patterns from base components
4. Test accessibility and performance
5. Document usage examples
6. Iterate based on user feedback

### Resources

- **Specification:** Schema Engine v1.0 Complete Specification
- **Component Library:** Implementation examples
- **Design Tokens:** Token definitions and exports
- **Accessibility:** WCAG guidelines and testing tools
- **Performance:** Optimization best practices

---

**Document Information:**
- **Version:** 1.0.0 (Complete)
- **Last Updated:** 2025-10-29
- **Status:** Production Ready
- **Sections:** 1-20 (Complete)

**Author:** Schema Engine Team  
**License:** MIT  
**Contact:** design-system@schema-engine.dev

---

