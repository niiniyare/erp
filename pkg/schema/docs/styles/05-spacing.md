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

