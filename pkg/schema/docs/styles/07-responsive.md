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

