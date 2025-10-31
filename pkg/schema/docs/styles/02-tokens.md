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

