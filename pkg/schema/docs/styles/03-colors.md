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

