# Schema Engine Design System

> **Complete styling and theming documentation for the Awo ERP Schema Engine**  
> **Purpose:** Comprehensive design system with accessibility-first approach  
> **Architecture:** Token-based with component composition  
> **Status:** Production ready with enterprise-grade features

---

## 📋 Design System Overview

**What This System Provides:**
- Complete design token architecture (colors, typography, spacing, etc.)
- Component styling patterns and composition rules
- Accessibility-first approach with WCAG 2.1 Level AA compliance
- Multi-theme support with brand customization capabilities
- Performance-optimized CSS architecture

**Key Design Principles:**
- **Accessibility First** - WCAG 2.1 compliance built-in, not retrofitted
- **Component Composability** - LEGO-block approach for interface building
- **Design Token Foundation** - All visual properties derive from centralized tokens
- **Progressive Enhancement** - Semantic HTML enhanced with CSS and JavaScript
- **Developer Experience** - Intuitive, well-documented with clear patterns

---

## 📁 Documentation Structure

| # | File | Purpose | Key Topics |
|---|------|---------|------------|
| 1 | **[Philosophy](docs/styles/01-philosophy.md)** | Design principles | Accessibility-first, composability, token-driven |
| 2 | **[Design Tokens](docs/styles/02-tokens.md)** | Token architecture | Global → semantic → component hierarchy |
| 3 | **[Colors](docs/styles/03-colors.md)** | Color system | Semantic colors, contrast compliance, color theory |
| 4 | **[Typography](docs/styles/04-typography.md)** | Font system | Modular scale, font stacks, responsive typography |
| 5 | **[Spacing](docs/styles/05-spacing.md)** | Layout spacing | Base-8 scale, padding/margin patterns |
| 6 | **[Architecture](docs/styles/06-architecture.md)** | CSS organization | File structure, naming conventions, methodologies |
| 7 | **[Responsive](docs/styles/07-responsive.md)** | Mobile-first design | Breakpoints, container queries, adaptive patterns |
| 8 | **[Accessibility](docs/styles/08-accessibility.md)** | A11y compliance | WCAG guidelines, keyboard nav, screen readers |
| 9 | **[Animation](docs/styles/09-animation.md)** | Motion design | Performance-conscious animations, reduced motion |
| 10 | **[Dark Mode](docs/styles/10-darkmode.md)** | Theme switching | Dark theme implementation, color adjustments |
| 11 | **[Theming](docs/styles/11-theming.md)** | Custom themes | Brand customization, theme generation, validation |
| 12 | **[Patterns](docs/styles/12-patterns.md)** | Common patterns | Layout patterns, interaction patterns, best practices |
| 13 | **[Composition](docs/styles/13-composition.md)** | Component building | Atomic design, variant systems, compound components |
| 14 | **[State](docs/styles/14-state.md)** | UI states | Hover, focus, active, disabled, loading states |
| 15 | **[Icons](docs/styles/15-icons.md)** | Icon system | SVG sprites, sizing, accessibility, icon patterns |
| 16 | **[Data Visualization](docs/styles/16-dataviz.md)** | Charts & graphs | Accessible data viz, color palettes, responsive charts |
| 17 | **[Forms](docs/styles/17-forms.md)** | Form styling | Input patterns, validation states, accessibility |
| 18 | **[Layouts](docs/styles/18-layouts.md)** | Page layouts | Grid systems, flexbox patterns, layout components |
| 19 | **[Performance](docs/styles/19-performance.md)** | CSS optimization | Core Web Vitals, bundle optimization, best practices |
| 20 | **[Implementation](docs/styles/20-implementation.md)** | Development guide | Setup, tooling, contributing, testing strategies |

**Total:** 20 comprehensive documentation files covering every aspect of the design system

---

## 🎯 Quick Reference for Common Tasks

| Task | Relevant Files | Key Concepts |
|------|----------------|-----------------|
| **Setup design tokens** | 02-tokens.md, 11-theming.md | Token hierarchy, theme configuration |
| **Implement dark mode** | 10-darkmode.md, 03-colors.md | Color adjustments, CSS custom properties |
| **Create accessible components** | 08-accessibility.md, 17-forms.md | WCAG compliance, keyboard navigation |
| **Optimize performance** | 19-performance.md, 09-animation.md | Core Web Vitals, efficient animations |
| **Build responsive layouts** | 07-responsive.md, 18-layouts.md | Mobile-first, container queries |
| **Style forms** | 17-forms.md, 14-state.md | Input patterns, validation feedback |
| **Implement animations** | 09-animation.md, 14-state.md | Performance-conscious motion |
| **Custom theming** | 11-theming.md, 02-tokens.md | Brand customization, theme generation |

---

## 🏗️ Technology Stack

```yaml
CSS Architecture:
  - Methodology: CSS Custom Properties + Component-based
  - Tokens: Hierarchical (Global → Semantic → Component)
  - Responsive: Mobile-first with container queries
  - Accessibility: WCAG 2.1 Level AA compliance

Theming:
  - Multi-theme support (light/dark + custom)
  - Brand color generation from primary palette
  - Automatic contrast validation
  - CSS custom property-based switching

Performance:
  - Core Web Vitals optimization
  - Critical CSS inlining
  - Component-level code splitting
  - Optimized animation techniques

Development:
  - TypeScript definitions for design tokens
  - Automated accessibility testing
  - Visual regression testing
  - Component documentation generation
```

---

## 🎨 Design Token Architecture

### Token Hierarchy

```css
/* Global Tokens (Primitive Values) */
--color-blue-500: #3b82f6;
--spacing-4: 1rem;
--font-size-lg: 1.125rem;

/* Semantic Tokens (Purpose-driven) */
--color-primary: var(--color-blue-500);
--spacing-component-padding: var(--spacing-4);
--font-size-body: var(--font-size-lg);

/* Component Tokens (Specific Use Cases) */
--button-background-primary: var(--color-primary);
--button-padding-horizontal: var(--spacing-component-padding);
--button-font-size: var(--font-size-body);
```

### Core Token Categories

```yaml
Color Tokens:
  - Primary/Secondary brand colors
  - Semantic colors (success, warning, error, info)
  - Neutral grays for text and backgrounds
  - Accessibility-compliant contrast ratios

Spacing Tokens:
  - Base-8 scale (8, 16, 24, 32, 40, 48...)
  - Component-specific spacing values
  - Responsive spacing with fluid scaling

Typography Tokens:
  - Modular scale (1.250 ratio - Major Third)
  - Font families (UI, Editorial, Monospace)
  - Line heights optimized for readability
  - Responsive typography with clamp()
```

---

## 🎭 Theme System

### Pre-built Themes

```javascript
// Default Theme
const defaultTheme = {
  colors: {
    primary: '#3b82f6',
    secondary: '#64748b',
    success: '#22c55e',
    warning: '#f59e0b',
    error: '#ef4444'
  }
}

// Dark Theme
const darkTheme = {
  colors: {
    background: '#0f172a',
    surface: '#1e293b',
    text: '#f1f5f9'
  }
}

// Custom Brand Theme
const customTheme = generateTheme({
  primaryColor: '#7c3aed', // Brand purple
  brandFont: 'Custom Sans'
})
```

---

## 📱 Responsive Design

### Breakpoint System

```css
/* Mobile-first breakpoints */
@media (min-width: 640px)  { /* sm */ }
@media (min-width: 768px)  { /* md */ }
@media (min-width: 1024px) { /* lg */ }
@media (min-width: 1280px) { /* xl */ }
@media (min-width: 1536px) { /* 2xl */ }
```

### Container Query Support

```css
/* Component-aware responsiveness */
.card {
  container-type: inline-size;
}

@container (min-width: 300px) {
  .card-content {
    display: grid;
    grid-template-columns: 1fr 1fr;
  }
}
```

---

## ♿ Accessibility Features

### WCAG 2.1 Compliance

- **Level AA minimum**, Level AAA where possible
- Color contrast ratios: 4.5:1 (normal text), 3:1 (large text)
- Keyboard navigation support for all interactive elements
- Screen reader optimization with semantic HTML and ARIA
- Focus management for dynamic content
- Reduced motion support with `prefers-reduced-motion`

### Testing Strategy

```yaml
Automated Testing:
  - Lighthouse accessibility audits
  - axe DevTools integration
  - Contrast ratio validation
  - Keyboard navigation testing

Manual Testing:
  - Screen reader testing (NVDA, JAWS, VoiceOver)
  - Keyboard-only navigation
  - High contrast mode testing
  - Zoom testing up to 200%
```

---

## ⚡ Performance Optimization

### Core Web Vitals Targets

```yaml
Performance Metrics:
  - LCP (Largest Contentful Paint): < 2.5s
  - FID (First Input Delay): < 100ms
  - CLS (Cumulative Layout Shift): < 0.1

Optimization Techniques:
  - Critical CSS inlining
  - Component-level code splitting
  - Efficient animation with transform/opacity
  - Image optimization with responsive images
  - Bundle size optimization through tree-shaking
```

---

## 🧪 Implementation Guide

### Setup Instructions

1. **Install design tokens:**
   ```bash
   npm install @awo/design-tokens
   ```

2. **Import base styles:**
   ```css
   @import '@awo/design-tokens/css/tokens.css';
   @import '@awo/design-tokens/css/base.css';
   ```

3. **Configure theme:**
   ```javascript
   import { ThemeProvider } from '@awo/design-system'
   import { defaultTheme } from '@awo/design-tokens'
   
   function App() {
     return (
       <ThemeProvider theme={defaultTheme}>
         {/* Your app */}
       </ThemeProvider>
     )
   }
   ```

### Development Workflow

```yaml
Component Development:
  1. Start with semantic HTML structure
  2. Apply design tokens (no hardcoded values)
  3. Implement responsive behavior
  4. Add accessibility features
  5. Test across themes and devices
  6. Document usage examples

Testing Checklist:
  - Visual regression tests
  - Accessibility audits
  - Performance benchmarks
  - Cross-browser compatibility
  - Theme switching functionality
```

---

## 📊 Component Architecture

### Atomic Design Structure

```
Atoms → Molecules → Organisms → Templates → Pages
  ↓         ↓           ↓          ↓         ↓
Button    Card      Header    PageLayout  HomePage
Input   SearchBox  Sidebar   FormLayout  UserProfile
Icon     NavItem   DataTable  Dashboard   Settings
```

### Variant System

```typescript
interface ButtonProps {
  variant: 'primary' | 'secondary' | 'destructive' | 'outline' | 'ghost'
  size: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  state: 'default' | 'hover' | 'focus' | 'active' | 'disabled'
}
```

---

## 🔗 External Resources

- **Design Token Specification:** https://design-tokens.github.io/community-group/
- **WCAG 2.1 Guidelines:** https://www.w3.org/WAI/WCAG21/quickref/
- **CSS Container Queries:** https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Container_Queries
- **Core Web Vitals:** https://web.dev/vitals/

---

## 📈 Documentation Reading Order

### For Developers (Implementation Focus)
1. **Start:** 06-architecture.md (CSS organization)
2. **Tokens:** 02-tokens.md → 03-colors.md → 04-typography.md
3. **Implementation:** 20-implementation.md → 13-composition.md
4. **Optimization:** 19-performance.md → 08-accessibility.md

### For Designers (Design Focus)
1. **Philosophy:** 01-philosophy.md → 12-patterns.md
2. **Visual System:** 03-colors.md → 04-typography.md → 05-spacing.md
3. **Theming:** 11-theming.md → 10-darkmode.md
4. **Responsive:** 07-responsive.md → 18-layouts.md

### For Complete Understanding
Follow the numerical order (01 → 20) for comprehensive coverage of the entire design system.

---

**Version:** 1.0.0  
**Status:** Production Ready  
**Last Updated:** 2025-11-02  
**Compliance:** WCAG 2.1 Level AA  
**Performance:** Core Web Vitals Optimized  