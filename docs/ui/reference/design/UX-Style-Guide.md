# ERP UI Style System - UX Developer Guide

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Comprehensive visual style guide and implementation specifications for UX developers
**SCOPE**: Typography, colors, spacing, shadows, components, and styling patterns
**TARGET AUDIENCE**: UX developers, frontend engineers, and design implementers
<!-- LLM-CONTEXT-END -->

## 🎨 Visual Style System Overview

This guide provides **precise implementation specifications** for the ERP UI design system, translating visual designs into **production-ready CSS/TailwindCSS code** for the **Templ + HTMX + Alpine.js + Flowbite** stack.

### **Design Principles**
- **Information Density**: Maximize data visibility without overwhelming users
- **Professional Aesthetics**: Clean, modern business application styling
- **Accessibility First**: High contrast, readable typography, clear interactive states
- **Consistency**: Predictable patterns across all interface elements

---

## 🎯 Color System

<!-- LLM-SECTION-COLORS-START -->
### **Primary Color Palette**

#### **Brand Colors**
```css
/* Primary Brand Colors */
--color-primary: #10B981;     /* Green/Teal - Primary actions, success states */
--color-primary-50: #ECFDF5;  /* Light background tints */
--color-primary-100: #D1FAE5; /* Subtle backgrounds */
--color-primary-500: #10B981; /* Standard primary */
--color-primary-600: #059669; /* Hover states */
--color-primary-700: #047857; /* Active states */

/* Secondary Colors */
--color-blue: #3B82F6;        /* Blue - Information, links */
--color-blue-50: #EFF6FF;
--color-blue-500: #3B82F6;
--color-blue-600: #2563EB;
```

#### **Status Color System**
```css
/* Success States */
--color-success: #10B981;     /* Completed, delivered, positive */
--color-success-bg: #ECFDF5;  /* Success background */
--color-success-text: #065F46; /* Success text */

/* Warning States */
--color-warning: #F59E0B;     /* In progress, pending */
--color-warning-bg: #FFFBEB;  /* Warning background */
--color-warning-text: #92400E; /* Warning text */

/* Error States */
--color-error: #EF4444;       /* Failed, declined, negative */
--color-error-bg: #FEF2F2;    /* Error background */
--color-error-text: #991B1B;  /* Error text */

/* Information States */
--color-info: #3B82F6;        /* Information, processing */
--color-info-bg: #EFF6FF;     /* Info background */
--color-info-text: #1E40AF;   /* Info text */
```

#### **Neutral Gray Palette**
```css
/* Professional Grays */
--color-gray-50: #F9FAFB;     /* Page backgrounds */
--color-gray-100: #F3F4F6;    /* Card backgrounds */
--color-gray-200: #E5E7EB;    /* Borders, dividers */
--color-gray-300: #D1D5DB;    /* Input borders */
--color-gray-400: #9CA3AF;    /* Placeholder text */
--color-gray-500: #6B7280;    /* Secondary text */
--color-gray-600: #4B5563;    /* Primary text */
--color-gray-700: #374151;    /* Headings */
--color-gray-800: #1F2937;    /* High emphasis text */
--color-gray-900: #111827;    /* Maximum contrast */
```

### **TailwindCSS Implementation**
```css
/* Extend your tailwind.config.js */
module.exports = {
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#ECFDF5',
          500: '#10B981',
          600: '#059669',
          700: '#047857'
        },
        success: '#10B981',
        warning: '#F59E0B',
        error: '#EF4444',
        info: '#3B82F6'
      }
    }
  }
}
```
<!-- LLM-SECTION-COLORS-END -->

---

## 📝 Typography System

<!-- LLM-SECTION-TYPOGRAPHY-START -->
### **Font Hierarchy**

#### **Font Family**
```css
/* Primary Font Stack */
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, 
             "Helvetica Neue", Arial, sans-serif;

/* Numbers and Data */
font-feature-settings: 'tnum'; /* Tabular numbers for data alignment */
```

#### **Typography Scale**
```css
/* Display Level - Dashboard metrics */
.text-display-large {
  font-size: 2.5rem;    /* 40px */
  line-height: 1.1;     /* 44px */
  font-weight: 700;     /* Bold */
  letter-spacing: -0.02em;
}

.text-display-medium {
  font-size: 2rem;      /* 32px */
  line-height: 1.2;     /* 38px */
  font-weight: 600;     /* Semi-bold */
  letter-spacing: -0.01em;
}

/* Heading Level - Section titles */
.text-heading-large {
  font-size: 1.5rem;    /* 24px */
  line-height: 1.3;     /* 32px */
  font-weight: 600;     /* Semi-bold */
}

.text-heading-medium {
  font-size: 1.25rem;   /* 20px */
  line-height: 1.4;     /* 28px */
  font-weight: 600;     /* Semi-bold */
}

.text-heading-small {
  font-size: 1.125rem;  /* 18px */
  line-height: 1.4;     /* 25px */
  font-weight: 600;     /* Semi-bold */
}

/* Body Level - Standard content */
.text-body-large {
  font-size: 1rem;      /* 16px */
  line-height: 1.5;     /* 24px */
  font-weight: 400;     /* Regular */
}

.text-body-medium {
  font-size: 0.875rem;  /* 14px */
  line-height: 1.4;     /* 20px */
  font-weight: 400;     /* Regular */
}

.text-body-small {
  font-size: 0.75rem;   /* 12px */
  line-height: 1.3;     /* 16px */
  font-weight: 400;     /* Regular */
}

/* Caption Level - Metadata */
.text-caption {
  font-size: 0.625rem;  /* 10px */
  line-height: 1.2;     /* 12px */
  font-weight: 500;     /* Medium */
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
```

#### **Text Color Classes**
```css
/* Text Color Hierarchy */
.text-primary { color: var(--color-gray-900); }    /* Primary content */
.text-secondary { color: var(--color-gray-600); }  /* Secondary content */
.text-tertiary { color: var(--color-gray-500); }   /* Tertiary content */
.text-placeholder { color: var(--color-gray-400); } /* Placeholder text */

/* Status Text Colors */
.text-success { color: var(--color-success-text); }
.text-warning { color: var(--color-warning-text); }
.text-error { color: var(--color-error-text); }
.text-info { color: var(--color-info-text); }
```

### **Flowbite Typography Integration**
```html
<!-- Metric Display -->
<h1 class="text-3xl font-bold text-gray-900">$275,825.00</h1>

<!-- Section Headings -->
<h2 class="text-xl font-semibold text-gray-800">Total Sales</h2>

<!-- Body Text -->
<p class="text-sm text-gray-600">vs Last month</p>

<!-- Caption Text -->
<span class="text-xs font-medium text-gray-500 uppercase tracking-wide">Status</span>
```
<!-- LLM-SECTION-TYPOGRAPHY-END -->

---

## 📏 Spacing System

<!-- LLM-SECTION-SPACING-START -->
### **Grid System: 8px Base Unit**

#### **Spacing Scale**
```css
/* Micro Spacing - Component internals */
--space-1: 0.25rem;   /* 4px */
--space-2: 0.5rem;    /* 8px */
--space-3: 0.75rem;   /* 12px */
--space-4: 1rem;      /* 16px */

/* Component Spacing - Between elements */
--space-5: 1.25rem;   /* 20px */
--space-6: 1.5rem;    /* 24px */
--space-8: 2rem;      /* 32px */

/* Section Spacing - Major content blocks */
--space-10: 2.5rem;   /* 40px */
--space-12: 3rem;     /* 48px */
--space-16: 4rem;     /* 64px */

/* Layout Spacing - Main layout zones */
--space-20: 5rem;     /* 80px */
--space-24: 6rem;     /* 96px */
```

#### **Component Spacing Patterns**
```css
/* Card Internal Spacing */
.card-padding {
  padding: 1.5rem;     /* 24px */
}

.card-padding-compact {
  padding: 1rem;       /* 16px */
}

/* Element Spacing */
.element-gap {
  gap: 1rem;           /* 16px - between related elements */
}

.section-gap {
  gap: 2rem;           /* 32px - between sections */
}

/* Form Spacing */
.form-field-gap {
  margin-bottom: 1rem; /* 16px */
}

.form-section-gap {
  margin-bottom: 2rem; /* 32px */
}
```

### **TailwindCSS Spacing Classes**
```html
<!-- Card Padding -->
<div class="p-6">        <!-- 24px padding -->
<div class="p-4">        <!-- 16px padding -->

<!-- Element Gaps -->
<div class="space-y-4">  <!-- 16px vertical spacing -->
<div class="space-y-6">  <!-- 24px vertical spacing -->
<div class="gap-4">      <!-- 16px gap in flex/grid -->

<!-- Margins -->
<div class="mb-4">       <!-- 16px bottom margin -->
<div class="mb-8">       <!-- 32px bottom margin -->
```
<!-- LLM-SECTION-SPACING-END -->

---

## 🖼️ Shadow & Elevation System

<!-- LLM-SECTION-SHADOWS-START -->
### **Shadow Levels**

#### **Card Shadows**
```css
/* Subtle Card Shadow - Default cards */
.shadow-card {
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 
              0 1px 2px 0 rgba(0, 0, 0, 0.06);
}

/* Elevated Card Shadow - Important cards */
.shadow-card-elevated {
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 
              0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

/* Modal Shadow - Overlays */
.shadow-modal {
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 
              0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

/* Floating Action Shadow - CTAs */
.shadow-floating {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 
              0 4px 6px -2px rgba(0, 0, 0, 0.05);
}
```

#### **Flowbite Shadow Classes**
```html
<!-- Default Card -->
<div class="shadow-sm bg-white rounded-lg">

<!-- Elevated Card -->
<div class="shadow-md bg-white rounded-lg">

<!-- Modal Overlay -->
<div class="shadow-xl bg-white rounded-lg">
```

### **Border Radius System**
```css
/* Border Radius Scale */
--radius-sm: 0.125rem;    /* 2px - small elements */
--radius-default: 0.25rem; /* 4px - standard radius */
--radius-md: 0.375rem;     /* 6px - cards */
--radius-lg: 0.5rem;       /* 8px - large cards */
--radius-xl: 0.75rem;      /* 12px - modals */
--radius-full: 9999px;     /* Circular - avatars, badges */
```
<!-- LLM-SECTION-SHADOWS-END -->

---

## 🎛️ Component Styling Specifications

<!-- LLM-SECTION-COMPONENTS-START -->
### **Metric Cards**

#### **Standard Metric Card**
```css
.metric-card {
  background: white;
  border-radius: 0.5rem;           /* 8px */
  padding: 1.5rem;                 /* 24px */
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border: 1px solid #E5E7EB;       /* Gray-200 */
}

.metric-value {
  font-size: 2rem;                 /* 32px */
  font-weight: 700;                /* Bold */
  color: #111827;                  /* Gray-900 */
  line-height: 1.1;
}

.metric-label {
  font-size: 0.875rem;             /* 14px */
  font-weight: 500;                /* Medium */
  color: #6B7280;                  /* Gray-500 */
  margin-bottom: 0.5rem;           /* 8px */
}

.metric-trend {
  font-size: 0.75rem;              /* 12px */
  font-weight: 500;                /* Medium */
  display: flex;
  align-items: center;
  gap: 0.25rem;                    /* 4px */
}

.metric-trend.positive { color: #059669; }  /* Green-600 */
.metric-trend.negative { color: #DC2626; }  /* Red-600 */
```

#### **Templ Implementation**
```go
templ MetricCard(props MetricCardProps) {
    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div class="flex items-center justify-between">
            <div class="flex items-center">
                <div class={ getIconClasses(props.IconType) }>
                    @Icon(props.IconName)
                </div>
                <div class="ml-3">
                    <p class="text-sm font-medium text-gray-500">
                        { props.Label }
                    </p>
                    <p class="text-2xl font-bold text-gray-900">
                        { props.Value }
                    </p>
                    <div class={ getTrendClasses(props.Trend) }>
                        @TrendIcon(props.Trend)
                        <span class="text-xs font-medium">
                            { props.TrendValue } vs Last month
                        </span>
                    </div>
                </div>
            </div>
            <button class="text-gray-400 hover:text-gray-600">
                @DotsIcon()
            </button>
        </div>
    </div>
}
```

### **Contact Cards**

#### **Contact Card Styling**
```css
.contact-card {
  background: white;
  border-radius: 0.5rem;           /* 8px */
  padding: 1rem;                   /* 16px */
  border: 1px solid #E5E7EB;       /* Gray-200 */
  transition: all 0.2s ease;
}

.contact-card:hover {
  border-color: #D1D5DB;           /* Gray-300 */
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
}

.contact-avatar {
  width: 2.5rem;                   /* 40px */
  height: 2.5rem;                  /* 40px */
  border-radius: 50%;
  object-fit: cover;
}

.contact-name {
  font-size: 0.875rem;             /* 14px */
  font-weight: 600;                /* Semi-bold */
  color: #111827;                  /* Gray-900 */
}

.contact-meta {
  font-size: 0.75rem;              /* 12px */
  color: #6B7280;                  /* Gray-500 */
  display: flex;
  align-items: center;
  gap: 0.25rem;                    /* 4px */
}

.contact-status {
  font-size: 0.75rem;              /* 12px */
  font-weight: 500;                /* Medium */
  padding: 0.125rem 0.5rem;        /* 2px 8px */
  border-radius: 9999px;           /* Full */
  background: #F3F4F6;             /* Gray-100 */
  color: #374151;                  /* Gray-700 */
}
```

### **Task Cards**

#### **Kanban Task Card**
```css
.task-card {
  background: white;
  border-radius: 0.375rem;         /* 6px */
  padding: 0.75rem;                /* 12px */
  border: 1px solid #E5E7EB;       /* Gray-200 */
  cursor: pointer;
  transition: all 0.2s ease;
}

.task-card:hover {
  border-color: #3B82F6;           /* Blue-500 */
  transform: translateY(-1px);
}

.task-title {
  font-size: 0.875rem;             /* 14px */
  font-weight: 600;                /* Semi-bold */
  color: #111827;                  /* Gray-900 */
  margin-bottom: 0.5rem;           /* 8px */
}

.task-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.75rem;             /* 12px */
}

.task-priority {
  padding: 0.125rem 0.375rem;      /* 2px 6px */
  border-radius: 0.25rem;          /* 4px */
  font-size: 0.625rem;             /* 10px */
  font-weight: 600;                /* Semi-bold */
  text-transform: uppercase;
}

.task-priority.low { 
  background: #DBEAFE; 
  color: #1E40AF; 
}
.task-priority.medium { 
  background: #FEF3C7; 
  color: #92400E; 
}
.task-priority.high { 
  background: #FEE2E2; 
  color: #991B1B; 
}
```

### **Data Tables**

#### **Table Styling**
```css
.data-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 0.5rem;           /* 8px */
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.table-header {
  background: #F9FAFB;             /* Gray-50 */
  border-bottom: 1px solid #E5E7EB; /* Gray-200 */
}

.table-header th {
  padding: 0.75rem 1rem;           /* 12px 16px */
  text-align: left;
  font-size: 0.75rem;              /* 12px */
  font-weight: 600;                /* Semi-bold */
  color: #374151;                  /* Gray-700 */
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.table-row {
  border-bottom: 1px solid #F3F4F6; /* Gray-100 */
  transition: background-color 0.2s ease;
}

.table-row:hover {
  background: #F9FAFB;             /* Gray-50 */
}

.table-cell {
  padding: 1rem;                   /* 16px */
  font-size: 0.875rem;             /* 14px */
  color: #111827;                  /* Gray-900 */
}
```

### **Modal Overlays**

#### **Modal Styling**
```css
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);   /* Semi-transparent black */
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
}

.modal-content {
  background: white;
  border-radius: 0.75rem;           /* 12px */
  max-width: 32rem;                 /* 512px */
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 25px rgba(0, 0, 0, 0.1);
}

.modal-header {
  padding: 1.5rem 1.5rem 0;        /* 24px 24px 0 */
  border-bottom: 1px solid #E5E7EB; /* Gray-200 */
}

.modal-title {
  font-size: 1.125rem;              /* 18px */
  font-weight: 600;                 /* Semi-bold */
  color: #111827;                   /* Gray-900 */
}

.modal-body {
  padding: 1.5rem;                  /* 24px */
}

.modal-footer {
  padding: 0 1.5rem 1.5rem;        /* 0 24px 24px */
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;                     /* 12px */
}
```
<!-- LLM-SECTION-COMPONENTS-END -->

---

## 🔘 Interactive States

<!-- LLM-SECTION-INTERACTIVE-START -->
### **Button Styles**

#### **Primary Button**
```css
.btn-primary {
  background: #10B981;             /* Primary color */
  color: white;
  border: none;
  border-radius: 0.375rem;         /* 6px */
  padding: 0.5rem 1rem;            /* 8px 16px */
  font-size: 0.875rem;             /* 14px */
  font-weight: 500;                /* Medium */
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary:hover {
  background: #059669;             /* Primary-600 */
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(16, 185, 129, 0.25);
}

.btn-primary:active {
  background: #047857;             /* Primary-700 */
  transform: translateY(0);
}

.btn-primary:disabled {
  background: #D1D5DB;             /* Gray-300 */
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}
```

#### **Secondary Button**
```css
.btn-secondary {
  background: white;
  color: #374151;                  /* Gray-700 */
  border: 1px solid #D1D5DB;       /* Gray-300 */
  border-radius: 0.375rem;         /* 6px */
  padding: 0.5rem 1rem;            /* 8px 16px */
  font-size: 0.875rem;             /* 14px */
  font-weight: 500;                /* Medium */
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary:hover {
  background: #F9FAFB;             /* Gray-50 */
  border-color: #9CA3AF;           /* Gray-400 */
}
```

### **Form Input Styles**

#### **Standard Input**
```css
.form-input {
  width: 100%;
  padding: 0.5rem 0.75rem;         /* 8px 12px */
  border: 1px solid #D1D5DB;       /* Gray-300 */
  border-radius: 0.375rem;         /* 6px */
  font-size: 0.875rem;             /* 14px */
  transition: all 0.2s ease;
  background: white;
}

.form-input:focus {
  outline: none;
  border-color: #3B82F6;           /* Blue-500 */
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input::placeholder {
  color: #9CA3AF;                  /* Gray-400 */
}

.form-input:disabled {
  background: #F3F4F6;             /* Gray-100 */
  border-color: #E5E7EB;           /* Gray-200 */
  color: #9CA3AF;                  /* Gray-400 */
  cursor: not-allowed;
}
```

### **Navigation States**

#### **Sidebar Navigation**
```css
.nav-item {
  display: flex;
  align-items: center;
  padding: 0.5rem 0.75rem;         /* 8px 12px */
  border-radius: 0.375rem;         /* 6px */
  color: #6B7280;                  /* Gray-500 */
  text-decoration: none;
  transition: all 0.2s ease;
  margin-bottom: 0.125rem;         /* 2px */
}

.nav-item:hover {
  background: #F3F4F6;             /* Gray-100 */
  color: #374151;                  /* Gray-700 */
}

.nav-item.active {
  background: #EFF6FF;             /* Blue-50 */
  color: #1D4ED8;                  /* Blue-700 */
  font-weight: 600;                /* Semi-bold */
}

.nav-icon {
  width: 1rem;                     /* 16px */
  height: 1rem;                    /* 16px */
  margin-right: 0.75rem;           /* 12px */
}
```
<!-- LLM-SECTION-INTERACTIVE-END -->

---

## 📱 Responsive Design Patterns

<!-- LLM-SECTION-RESPONSIVE-START -->
### **Breakpoint System**

```css
/* Mobile First Approach */
/* xs: 0px - 639px (mobile) */
/* sm: 640px - 767px (large mobile) */
/* md: 768px - 1023px (tablet) */
/* lg: 1024px - 1279px (desktop) */
/* xl: 1280px+ (large desktop) */
```

### **Layout Adaptations**

#### **Dashboard Grid**
```css
/* Desktop: 3-column grid */
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;                     /* 24px */
}

/* Tablet: 2-column grid */
@media (max-width: 1023px) {
  .dashboard-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;                     /* 16px */
  }
}

/* Mobile: 1-column stack */
@media (max-width: 767px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
    gap: 0.75rem;                  /* 12px */
  }
}
```

#### **Contact Cards**
```css
/* Desktop: 3-column grid */
.contact-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;                       /* 16px */
}

/* Tablet: 2-column grid */
@media (max-width: 1023px) {
  .contact-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* Mobile: 1-column list */
@media (max-width: 767px) {
  .contact-grid {
    grid-template-columns: 1fr;
  }
  
  .contact-card {
    padding: 0.75rem;              /* 12px */
  }
}
```

### **Typography Scaling**
```css
/* Desktop Typography */
.metric-value {
  font-size: 2rem;                 /* 32px */
}

/* Mobile Typography */
@media (max-width: 767px) {
  .metric-value {
    font-size: 1.5rem;              /* 24px */
  }
  
  .section-title {
    font-size: 1.125rem;            /* 18px */
  }
}
```
<!-- LLM-SECTION-RESPONSIVE-END -->

---

## ⚡ Animation & Transitions

<!-- LLM-SECTION-ANIMATIONS-START -->
### **Micro-interactions**

#### **Standard Transitions**
```css
/* Default transition for interactive elements */
.transition-default {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Smooth entrance animations */
.fade-in {
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from { 
    opacity: 0; 
    transform: translateY(10px); 
  }
  to { 
    opacity: 1; 
    transform: translateY(0); 
  }
}

/* Card hover effects */
.card-hover {
  transition: all 0.2s ease;
}

.card-hover:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}
```

#### **Loading States**
```css
/* Skeleton loading animation */
.skeleton {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: loading 1.5s infinite;
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* Pulse animation for loading elements */
.pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
```
<!-- LLM-SECTION-ANIMATIONS-END -->

---

## 🛠️ Implementation Templates

<!-- LLM-SECTION-IMPLEMENTATION-START -->
### **Templ Component Templates**

#### **Complete Metric Card Component**
```go
type MetricCardProps struct {
    ID          string
    Label       string
    Value       string
    Trend       string // "up", "down", "neutral"
    TrendValue  string
    Icon        string
    IconColor   string
    OnClick     string
}

templ MetricCard(props MetricCardProps) {
    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow duration-200">
        <div class="flex items-center justify-between">
            <!-- Icon and Content -->
            <div class="flex items-center">
                <div class={ fmt.Sprintf("flex-shrink-0 w-10 h-10 rounded-lg flex items-center justify-center %s", props.IconColor) }>
                    @Icon(props.Icon, "w-6 h-6 text-white")
                </div>
                <div class="ml-4">
                    <p class="text-sm font-medium text-gray-500 mb-1">
                        { props.Label }
                    </p>
                    <p class="text-2xl font-bold text-gray-900">
                        { props.Value }
                    </p>
                    <div class={ getTrendClasses(props.Trend) }>
                        if props.Trend == "up" {
                            @TrendUpIcon()
                        } else if props.Trend == "down" {
                            @TrendDownIcon()
                        }
                        <span class="text-xs font-medium ml-1">
                            { props.TrendValue } vs Last month
                        </span>
                    </div>
                </div>
            </div>
            <!-- Action Menu -->
            <button class="text-gray-400 hover:text-gray-600 transition-colors">
                @DotsVerticalIcon()
            </button>
        </div>
    </div>
}

func getTrendClasses(trend string) string {
    base := "flex items-center text-xs font-medium"
    switch trend {
    case "up":
        return base + " text-green-600"
    case "down":
        return base + " text-red-600"
    default:
        return base + " text-gray-500"
    }
}
```

#### **Contact Card Component**
```go
type ContactCardProps struct {
    ID          string
    Name        string
    Email       string
    Phone       string
    Company     string
    Avatar      string
    Status      string
    Date        string
    OnClick     string
}

templ ContactCard(props ContactCardProps) {
    <div class="bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 hover:shadow-sm transition-all duration-200 cursor-pointer"
         onclick={ props.OnClick }>
        <div class="flex items-start space-x-3">
            <!-- Avatar -->
            <img class="w-10 h-10 rounded-full object-cover" 
                 src={ props.Avatar } 
                 alt={ props.Name }>
            
            <!-- Content -->
            <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between">
                    <p class="text-sm font-semibold text-gray-900 truncate">
                        { props.Name }
                    </p>
                    <span class="text-xs text-gray-500">
                        { props.Date }
                    </span>
                </div>
                
                <!-- Contact Info -->
                <div class="mt-1 space-y-1">
                    <div class="flex items-center text-xs text-gray-500">
                        @PhoneIcon()
                        <span class="ml-1">{ props.Phone }</span>
                    </div>
                    <div class="flex items-center text-xs text-gray-500">
                        @EmailIcon()
                        <span class="ml-1 truncate">{ props.Email }</span>
                    </div>
                    <div class="flex items-center text-xs text-gray-500">
                        @BuildingIcon()
                        <span class="ml-1">{ props.Company }</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
}
```

### **Alpine.js Integration Patterns**

#### **Interactive Table Component**
```javascript
// Table with sorting and filtering
Alpine.data('dataTable', () => ({
    data: [],
    filteredData: [],
    sortField: null,
    sortDirection: 'asc',
    searchTerm: '',
    
    init() {
        this.loadData();
        this.$watch('searchTerm', () => this.filterData());
    },
    
    async loadData() {
        // HTMX integration for data loading
        const response = await fetch('/api/data');
        this.data = await response.json();
        this.filterData();
    },
    
    sort(field) {
        if (this.sortField === field) {
            this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
        } else {
            this.sortField = field;
            this.sortDirection = 'asc';
        }
        this.filterData();
    },
    
    filterData() {
        let filtered = this.data;
        
        // Apply search filter
        if (this.searchTerm) {
            filtered = filtered.filter(item => 
                Object.values(item).some(value => 
                    String(value).toLowerCase().includes(this.searchTerm.toLowerCase())
                )
            );
        }
        
        // Apply sorting
        if (this.sortField) {
            filtered.sort((a, b) => {
                const aVal = a[this.sortField];
                const bVal = b[this.sortField];
                const modifier = this.sortDirection === 'asc' ? 1 : -1;
                return aVal > bVal ? modifier : -modifier;
            });
        }
        
        this.filteredData = filtered;
    }
}));
```

### **HTMX Integration Examples**

#### **Dynamic Content Loading**
```html
<!-- Auto-updating dashboard metrics -->
<div hx-get="/api/dashboard/metrics" 
     hx-trigger="load, every 30s" 
     hx-target="#metrics-container"
     hx-swap="innerHTML">
    <div id="metrics-container">
        <!-- Metrics will be loaded here -->
    </div>
</div>

<!-- Infinite scroll contact list -->
<div hx-get="/api/contacts" 
     hx-trigger="intersect once" 
     hx-target="#contact-list" 
     hx-swap="beforeend">
    <div id="contact-list">
        <!-- Contact cards will be appended here -->
    </div>
</div>

<!-- Real-time search -->
<input type="text" 
       hx-get="/api/search" 
       hx-trigger="keyup changed delay:300ms" 
       hx-target="#search-results"
       placeholder="Search contacts...">
```
<!-- LLM-SECTION-IMPLEMENTATION-END -->

---

## 🎯 Quick Reference

### **Most Used Classes**
```html
<!-- Cards -->
<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">

<!-- Buttons -->
<button class="bg-primary-500 text-white px-4 py-2 rounded-md hover:bg-primary-600">

<!-- Text -->
<h1 class="text-2xl font-bold text-gray-900">
<p class="text-sm text-gray-600">

<!-- Spacing -->
<div class="space-y-4">        <!-- Vertical spacing -->
<div class="space-x-4">        <!-- Horizontal spacing -->
<div class="p-6">              <!-- Padding -->
<div class="mb-4">             <!-- Margin bottom -->

<!-- Grid -->
<div class="grid grid-cols-3 gap-6">      <!-- 3-column grid -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">  <!-- Responsive -->
```

### **Color Usage Guide**
- **Green/Teal (#10B981)**: Success states, positive metrics, primary actions
- **Blue (#3B82F6)**: Information, links, secondary actions  
- **Red (#EF4444)**: Errors, negative metrics, destructive actions
- **Orange (#F59E0B)**: Warnings, pending states, moderate priority
- **Gray**: Text hierarchy, borders, backgrounds

### **Spacing Guidelines**
- **4px/8px**: Micro spacing within components
- **16px/24px**: Standard spacing between elements
- **32px/48px**: Section spacing
- **64px+**: Layout-level spacing

This style guide provides everything needed to implement the ERP UI design system with **pixel-perfect accuracy** and **consistent user experience** across all components and pages.