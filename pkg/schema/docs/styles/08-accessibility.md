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

