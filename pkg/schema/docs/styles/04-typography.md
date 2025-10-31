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

