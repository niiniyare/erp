# 🎨 Schema Engine Design System

**Complete Style Guide for JSON-Driven UI Components**

**Version:** 1.0.0  
**Status:** Production Ready  
**Design Philosophy:** Accessible, Composable, Customizable  
**Last Updated:** 2025-10-29

---

## Overview

A comprehensive design system that powers the Schema Engine UI components. Built on accessibility-first principles with complete theming support, responsive design patterns, and performance optimization.

## Core Principles

### 🎯 **Accessibility First**
- WCAG 2.1 Level AA compliance minimum
- Keyboard navigation and screen reader support
- 4.5:1 color contrast ratios
- Semantic HTML foundation

### 🧩 **Component Composability**
- LEGO-like building blocks
- Simple → Composite → Complex patterns
- Reusable and consistent interfaces

### 🎨 **Design Token Foundation**
- Centralized token system
- Never hardcode visual values
- Enables theming and consistency

### ⚡ **Progressive Enhancement**
- HTML-first foundation
- CSS enhancement layer
- JavaScript interaction layer

## Quick Reference

### Color Palette
```css
/* Primary Colors */
--primary: #3b82f6;
--primary-hover: #2563eb;
--primary-active: #1d4ed8;

/* Semantic Colors */
--success: #10b981;
--warning: #f59e0b;
--error: #ef4444;
--info: #06b6d4;
```

### Typography Scale
```css
/* Font Sizes */
--text-xs: 0.75rem;    /* 12px */
--text-sm: 0.875rem;   /* 14px */
--text-base: 1rem;     /* 16px */
--text-lg: 1.125rem;   /* 18px */
--text-xl: 1.25rem;    /* 20px */
```

### Spacing System
```css
/* Consistent 4px base unit */
--space-1: 0.25rem;  /* 4px */
--space-2: 0.5rem;   /* 8px */
--space-4: 1rem;     /* 16px */
--space-8: 2rem;     /* 32px */
```

## Architecture

The design system follows a **component-first architecture** with three layers:

1. **Foundation Layer** - Tokens, utilities, base styles
2. **Component Layer** - Reusable UI components  
3. **Pattern Layer** - Complex layouts and compositions

```
Foundation (Tokens) → Components → Patterns → Applications
```

## Documentation Structure

### Part I: Foundation & Tokens
- [01. Design Philosophy](./styles/01-philosophy.md) - Core principles and accessibility
- [02. Design Tokens](./styles/02-tokens.md) - Token system and organization
- [03. Color System](./styles/03-colors.md) - Color palette and usage guidelines
- [04. Typography](./styles/04-typography.md) - Font system and hierarchy
- [05. Spacing & Layout](./styles/05-spacing.md) - Spacing system and grid

### Part II: Component Architecture
- [06. Component Architecture](./styles/06-architecture.md) - Component structure and APIs
- [07. Responsive Design](./styles/07-responsive.md) - Mobile-first patterns
- [08. Accessibility](./styles/08-accessibility.md) - A11y implementation guide
- [09. Animation & Motion](./styles/09-animation.md) - Motion design principles
- [10. Dark Mode](./styles/10-darkmode.md) - Dark theme implementation

### Part III: Advanced Patterns
- [11. Theming Strategy](./styles/11-theming.md) - Custom theme creation
- [12. Component Patterns](./styles/12-patterns.md) - Common UI patterns
- [13. Composition Principles](./styles/13-composition.md) - Layout composition
- [14. State Management](./styles/14-state.md) - Visual state patterns
- [15. Icon System](./styles/15-icons.md) - Icon usage and customization

### Part IV: Specialized Components
- [16. Data Visualization](./styles/16-dataviz.md) - Charts and data displays
- [17. Forms & Validation](./styles/17-forms.md) - Form styling and feedback
- [18. Layout Patterns](./styles/18-layouts.md) - Page and section layouts
- [19. Performance Guidelines](./styles/19-performance.md) - CSS optimization
- [20. Implementation Guide](./styles/20-implementation.md) - Integration patterns

## Key Features

| Feature | Status | Description |
|---------|--------|-------------|
| **Design Tokens** | ✅ Complete | 200+ tokens for colors, spacing, typography |
| **Dark Mode** | ✅ Complete | Automatic theme switching |
| **Responsive** | ✅ Complete | Mobile-first breakpoint system |
| **Accessibility** | ✅ Complete | WCAG 2.1 AA compliant |
| **Animation** | ✅ Complete | Performance-optimized transitions |
| **Theming** | ✅ Complete | Custom CSS property system |
| **Component Library** | ✅ Complete | 40+ styled components |
| **Icon System** | ✅ Complete | SVG icon library with customization |

## Browser Support

| Browser | Version | Support |
|---------|---------|---------|
| **Chrome** | 88+ | ✅ Full |
| **Firefox** | 84+ | ✅ Full |
| **Safari** | 14+ | ✅ Full |
| **Edge** | 88+ | ✅ Full |

**Modern CSS Features:**
- CSS Custom Properties (CSS Variables)
- CSS Grid and Flexbox
- Container Queries (progressive enhancement)

## Integration

### Quick Start
```html
<!-- Include design system CSS -->
<link rel="stylesheet" href="/css/schema-design-system.css">

<!-- Theme configuration -->
<script>
  document.documentElement.setAttribute('data-theme', 'light');
</script>
```

### Schema Integration
```json
{
  "type": "form",
  "theme": "enterprise",
  "components": {
    "button": {
      "variant": "primary",
      "size": "md"
    }
  }
}
```

## Performance

- **CSS Bundle Size:** ~45KB (minified + gzipped)
- **Component Styles:** Lazy-loaded per component
- **Critical CSS:** Inlined for above-the-fold content
- **Runtime Impact:** Zero JavaScript for base styling

---

**Get Started:** Begin with [Design Philosophy](./styles/01-philosophy.md) or jump to [Implementation Guide](./styles/20-implementation.md) for immediate integration.