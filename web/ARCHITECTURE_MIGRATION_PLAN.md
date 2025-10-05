# ERP UI Architecture Migration Plan

## Executive Summary
This document outlines the phased migration from our current TypeScript-heavy client-side architecture to a server-first architecture using Templ + HTMX + Alpine.js + Flowbite as specified in our architecture documentation.

## Current State Assessment

### Architecture Violations
1. **Technology Stack**: TypeScript hooks instead of HTMX-first approach
2. **State Management**: Client-heavy instead of server-first
3. **Component Organization**: Flat structure instead of atomic design
4. **Progressive Enhancement**: JavaScript-dependent instead of server-fallback
5. **Performance**: Heavy client-side logic instead of minimal JavaScript

## Migration Strategy

### Phase 1: Foundation Restructuring (Week 1-2)
**Objective**: Establish proper file structure and build foundations

#### 1.1 Atomic Design File Structure

```web/
├── components/
│   ├── foundations/                    # Design system core
│   │   ├── tokens/
│   │   │   ├── colors.css             # Color palette, semantic colors
│   │   │   ├── typography.css         # Font families, sizes, weights, line heights
│   │   │   ├── spacing.css            # Margin, padding scale
│   │   │   ├── shadows.css            # Box shadow definitions
│   │   │   ├── borders.css            # Border radius, widths
│   │   │   └── breakpoints.css        # Responsive breakpoints
│   │   ├── base/
│   │   │   ├── reset.css              # CSS reset/normalize
│   │   │   ├── globals.css            # Global styles, body, html
│   │   │   └── variables.css          # CSS custom properties
│   │   └── utilities/
│   │       ├── layout.css             # Flexbox, grid utilities
│   │       ├── visibility.css         # Display, opacity utilities
│   │       └── alpine-utils.js        # Alpine.js helper directives
│   │
│   ├── atoms/                          # Basic building blocks
│   │   ├── buttons/
│   │   │   ├── base.templ             # Primary button
│   │   │   ├── secondary.templ        # Secondary variant
│   │   │   ├── ghost.templ            # Ghost/text button
│   │   │   ├── icon.templ             # Icon-only button
│   │   │   ├── button-group.templ     # Button group container
│   │   │   └── loading.templ          # Button loading state
│   │   ├── inputs/
│   │   │   ├── text.templ             # Text input
│   │   │   ├── textarea.templ         # Multi-line text
│   │   │   ├── select.templ           # Dropdown select
│   │   │   ├── checkbox.templ         # Checkbox
│   │   │   ├── radio.templ            # Radio button
│   │   │   ├── toggle.templ           # Toggle switch
│   │   │   ├── date.templ             # Date picker input
│   │   │   ├── file.templ             # File upload
│   │   │   └── label.templ            # Form label
│   │   ├── typography/
│   │   │   ├── heading.templ          # H1-H6 components
│   │   │   ├── text.templ             # Paragraph, span variants
│   │   │   ├── link.templ             # Anchor tag
│   │   │   ├── code.templ             # Inline code
│   │   │   └── blockquote.templ       # Quote block
│   │   ├── icons/
│   │   │   ├── icon.templ             # Base icon wrapper
│   │   │   ├── svg/                   # SVG icon assets
│   │   │   │   ├── user.svg
│   │   │   │   ├── dashboard.svg
│   │   │   │   ├── settings.svg
│   │   │   │   └── ...
│   │   │   └── spinner.templ          # Loading spinner
│   │   ├── images/
│   │   │   ├── avatar.templ           # User avatar
│   │   │   └── logo.templ             # Brand logo
│   │   └── divider/
│   │       ├── horizontal.templ       # Horizontal divider
│   │       └── vertical.templ         # Vertical divider
│   │
│   ├── molecules/                      # Simple component combinations
│   │   ├── forms/
│   │   │   ├── field.templ            # Label + input + error
│   │   │   ├── field-group.templ      # Multiple related fields
│   │   │   ├── search.templ           # Search input with icon
│   │   │   ├── select-with-search.templ # Searchable select
│   │   │   └── date-range.templ       # Start/end date picker
│   │   ├── cards/
│   │   │   ├── base.templ             # Basic card container
│   │   │   ├── with-header.templ      # Card with header/footer
│   │   │   ├── stat.templ             # Stat card (value + label)
│   │   │   └── list-item.templ        # List item card
│   │   ├── navigation/
│   │   │   ├── breadcrumbs.templ      # Breadcrumb navigation
│   │   │   ├── pagination.templ       # Page navigation
│   │   │   ├── tabs.templ             # Tab navigation
│   │   │   └── steps.templ            # Step indicator
│   │   ├── feedback/
│   │   │   ├── alert.templ            # Info/success/warning/error alerts
│   │   │   ├── toast.templ            # Toast notification
│   │   │   ├── badge.templ            # Status badge
│   │   │   ├── tooltip.templ          # Tooltip overlay
│   │   │   └── empty-state.templ      # No data state
│   │   ├── menu/
│   │   │   ├── dropdown.templ         # Dropdown menu
│   │   │   └── context-menu.templ     # Right-click menu
│   │   └── media/
│   │       ├── avatar-with-text.templ # Avatar + name + subtitle
│   │       └── thumbnail.templ        # Image thumbnail with overlay
│   │
│   ├── organisms/                      # Complex composite components
│   │   ├── headers/
│   │   │   ├── site-header.templ      # Main site header
│   │   │   ├── page-header.templ      # Page title + actions
│   │   │   └── mobile-header.templ    # Mobile-specific header
│   │   ├── navigation/
│   │   │   ├── sidebar.templ          # Main sidebar navigation
│   │   │   ├── sidebar-item.templ     # Sidebar menu item
│   │   │   ├── mobile-menu.templ      # Mobile hamburger menu
│   │   │   └── user-menu.templ        # User profile dropdown
│   │   ├── data-tables/
│   │   │   ├── table.templ            # Base table structure
│   │   │   ├── table-header.templ     # Sortable header row
│   │   │   ├── table-row.templ        # Data row
│   │   │   ├── table-cell.templ       # Individual cell
│   │   │   ├── table-filters.templ    # Filter controls
│   │   │   ├── table-actions.templ    # Bulk actions toolbar
│   │   │   └── table-pagination.templ # Table-specific pagination
│   │   ├── modals/
│   │   │   ├── dialog.templ           # Basic dialog modal
│   │   │   ├── confirm.templ          # Confirmation dialog
│   │   │   ├── drawer.templ           # Side drawer/panel
│   │   │   └── modal-overlay.templ    # Backdrop overlay
│   │   ├── forms/
│   │   │   ├── login-form.templ       # Authentication form
│   │   │   ├── filter-panel.templ     # Advanced filter form
│   │   │   └── multi-step-form.templ  # Wizard-style form
│   │   └── widgets/
│   │       ├── stat-grid.templ        # Dashboard stat cards
│   │       ├── chart-card.templ       # Chart with header
│   │       └── activity-feed.templ    # Activity timeline
│   │
│   └── features/                       # Feature-specific compositions
│       ├── user-management/
│       │   ├── user-table.templ       # Users data table
│       │   ├── user-form.templ        # Create/edit user
│       │   ├── user-profile.templ     # User profile display
│       │   ├── role-selector.templ    # Role assignment
│       │   └── permissions-grid.templ # Permission checkboxes
│       ├── dashboard/
│       │   ├── summary-cards.templ    # KPI summary
│       │   ├── recent-activity.templ  # Activity widget
│       │   └── quick-actions.templ    # Action shortcuts
│       ├── data-management/
│       │   ├── crud-table.templ       # Generic CRUD table
│       │   ├── import-wizard.templ    # Data import flow
│       │   └── export-dialog.templ    # Export options
│       ├── notifications/
│       │   ├── notification-center.templ # Notification dropdown
│       │   ├── notification-item.templ   # Single notification
│       │   └── notification-settings.templ # Preferences
│       └── tenant/
│           ├── tenant-selector.templ  # Switch tenant dropdown
│           └── tenant-settings.templ  # Tenant configuration
│
├── layouts/                            # Page layout templates
│   ├── base.templ                      # HTML shell (head, body)
│   ├── auth.templ                      # Authentication layout
│   ├── app.templ                       # Main app layout (with sidebar)
│   ├── minimal.templ                   # Clean layout (no nav)
│   └── print.templ                     # Print-optimized layout
│
├── partials/                           # Reusable page fragments
│   ├── head.templ                      # Meta tags, CSS links
│   ├── scripts.templ                   # JS scripts, analytics
│   ├── nav.templ                       # Main navigation partial
│   └── footer.templ                    # Site footer
│
├── pages/                              # Page-level templates
│   ├── auth/
│   │   ├── login.templ                # Login page
│   │   ├── register.templ             # Registration page
│   │   ├── forgot-password.templ      # Password reset
│   │   └── verify-email.templ         # Email verification
│   ├── dashboard/
│   │   ├── index.templ                # Main dashboard
│   │   └── analytics.templ            # Analytics dashboard
│   ├── settings/
│   │   ├── profile.templ              # User profile settings
│   │   ├── account.templ              # Account settings
│   │   ├── security.templ             # Security settings
│   │   └── preferences.templ          # User preferences
│   ├── admin/
│   │   ├── users.templ                # User management page
│   │   ├── roles.templ                # Role management
│   │   ├── tenants.templ              # Tenant management
│   │   └── system.templ               # System settings
│   └── errors/
│       ├── 404.templ                  # Not found
│       ├── 500.templ                  # Server error
│       └── 403.templ                  # Forbidden
│
├── domain/                             # Business domain modules
│   ├── finance/                        # Finance module
│   │   ├── accounts/                  # Chart of accounts service
│   │   │   ├── list.templ             # Account list view
│   │   │   ├── form.templ             # Create/edit account
│   │   │   ├── detail.templ           # Account detail view
│   │   │   └── tree.templ             # Hierarchical account tree
│   │   ├── ledger/                    # General ledger service
│   │   │   ├── entries.templ          # Journal entries
│   │   │   ├── entry-form.templ       # Entry creation
│   │   │   └── trial-balance.templ    # Trial balance report
│   │   ├── reports/                   # Financial reporting
│   │   │   ├── balance-sheet.templ    # Balance sheet
│   │   │   ├── income-statement.templ # P&L statement
│   │   │   ├── cash-flow.templ        # Cash flow statement
│   │   │   └── report-filters.templ   # Report date/options
│   │   └── currency/                  # Currency management
│   │       ├── list.templ             # Currency list
│   │       ├── exchange-rates.templ   # Exchange rate management
│   │       └── conversion.templ       # Currency converter
│   │
│   ├── hr/                            # Human Resources module
│   │   ├── employees/
│   │   │   ├── directory.templ        # Employee directory
│   │   │   ├── profile.templ          # Employee profile
│   │   │   └── onboarding.templ       # Onboarding workflow
│   │   ├── attendance/
│   │   │   ├── timesheet.templ        # Time tracking
│   │   │   ├── calendar.templ         # Attendance calendar
│   │   │   └── requests.templ         # Leave requests
│   │   └── payroll/
│   │       ├── overview.templ         # Payroll dashboard
│   │       ├── pay-run.templ          # Process payroll
│   │       └── payslips.templ         # Payslip generation
│   │
│   ├── crm/                           # Customer Relationship Management
│   │   ├── contacts/
│   │   │   ├── list.templ             # Contact list
│   │   │   ├── detail.templ           # Contact detail
│   │   │   └── form.templ             # Contact form
│   │   ├── leads/
│   │   │   ├── pipeline.templ         # Sales pipeline
│   │   │   ├── kanban.templ           # Lead kanban board
│   │   │   └── conversion.templ       # Lead conversion
│   │   └── opportunities/
│   │       ├── list.templ             # Opportunity list
│   │       └── forecast.templ         # Sales forecast
│   │
│   ├── inventory/                     # Inventory Management
│   │   ├── products/
│   │   │   ├── catalog.templ          # Product catalog
│   │   │   ├── detail.templ           # Product detail
│   │   │   └── variants.templ         # Product variants
│   │   ├── warehouses/
│   │   │   ├── list.templ             # Warehouse list
│   │   │   └── stock-levels.templ     # Stock by warehouse
│   │   └── movements/
│   │       ├── transactions.templ     # Stock movements
│   │       └── adjustments.templ      # Inventory adjustments
│   │
│   └── projects/                      # Project Management
│       ├── list.templ                 # Project list
│       ├── kanban.templ               # Project board
│       ├── gantt.templ                # Gantt chart view
│       ├── tasks/
│       │   ├── list.templ             # Task list
│       │   └── detail.templ           # Task detail
│       └── time-tracking/
│           └── tracker.templ          # Time tracker widget
│
└── static/                            # Static assets
    ├── css/
    │   ├── foundations.css            # Compiled foundation styles
    │   ├── components.css             # Component-specific styles
    │   ├── themes.css                 # Theme variations (light/dark)
    │   └── print.css                  # Print styles
    │
    ├── js/
    │   ├── stores/                    # Alpine.js stores (minimal)
    │   │   ├── ui.js                  # UI state (modals, dropdowns)
    │   │   ├── notifications.js       # Toast notifications
    │   │   └── theme.js               # Theme switching
    │   ├── utils/
    │   │   ├── htmx-extensions.js     # Custom HTMX extensions
    │   │   ├── alpine-directives.js   # Custom Alpine directives
    │   │   └── form-validation.js     # Client-side validation
    │   └── app.js                     # Main app initialization
    │
    ├── images/
    │   ├── logo.svg                   # Brand logo
    │   ├── logo-dark.svg              # Dark mode logo
    │   ├── favicon.ico                # Site favicon
    │   └── illustrations/             # Empty states, errors
    │       ├── empty.svg
    │       ├── 404.svg
    │       └── error.svg
    │
    ├── fonts/                         # Custom web fonts
    │   └── inter/                     # Font family folder
    │       ├── inter-regular.woff2
    │       ├── inter-medium.woff2
    │       └── inter-bold.woff2
    │
    └── vendor/                        # Third-party libraries
        ├── htmx.min.js                # HTMX library
        ├── alpine.min.js              # Alpine.js
        └── chart.min.js               # Charting library (if needed)
```

#### 1.2 Alpine.js Store Architecture
Replace complex TypeScript interfaces with simple Alpine stores:

```javascript
// web/static/js/stores/app.js
Alpine.store('app', {
    loading: false,
    theme: 'light',
    language: 'en',
    
    setLoading(loading) {
        this.loading = loading;
    },
    
    setTheme(theme) {
        this.theme = theme;
        document.documentElement.setAttribute('data-theme', theme);
    }
});

// web/static/js/stores/user.js
Alpine.store('user', {
    current: null,
    permissions: [],
    
    hasPermission(permission) {
        return this.permissions.includes(permission);
    },
    
    update(userData) {
        this.current = { ...this.current, ...userData };
    }
});
```

### Phase 2: Server-First Component Migration (Week 3-4)
**Objective**: Convert complex TypeScript hooks to simple Alpine.js + HTMX

#### 2.1 Component Conversion Pattern
Transform complex TypeScript hooks into server-first components:

**Before (TypeScript Hook)**:
```typescript
interface DataTableStore {
    // 20+ properties and methods
    updateStatus(status: UserStatus): Promise<void>; // Client API calls
}
```

**After (Server-First + Alpine.js)**:
```go
// server-side templ component
templ DataTable(props DataTableProps) {
    <div x-data="simpleDataTable()" 
         hx-get="/api/data-table/refresh"
         hx-trigger="refresh-table from:body">
        <!-- Server-rendered content -->
    </div>
}
```

```javascript
// Simple Alpine.js behavior
function simpleDataTable() {
    return {
        selectedRows: [],
        
        toggleRow(id) {
            const index = this.selectedRows.indexOf(id);
            if (index > -1) {
                this.selectedRows.splice(index, 1);
            } else {
                this.selectedRows.push(id);
            }
        }
    };
}
```

#### 2.2 HTMX Integration Patterns
Implement server-driven updates:

```go
// Go handler with HTMX response
func (h *Handler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
    // Server-side business logic
    err := h.userService.UpdateStatus(userID, status)
    if err != nil {
        w.Header().Set("HX-Trigger", `{"error": "Status update failed"}`)
        return
    }
    
    // Return updated component
    w.Header().Set("HX-Trigger", `{"statusUpdated": true}`)
    tmpl := components.UserStatusIndicator(user)
    tmpl.Render(r.Context(), w)
}
```

### Phase 3: Progressive Enhancement Implementation (Week 5)
**Objective**: Ensure all components work without JavaScript

#### 3.1 Progressive Enhancement Pattern
```go
// Base functionality without JavaScript
templ UserForm(props UserFormProps) {
    <form method="POST" action="/users">
        @FormField("name", props.Name, "")
        <button type="submit">Save User</button>
    </form>
}

// Enhanced with HTMX
templ EnhancedUserForm(props UserFormProps) {
    <form method="POST" action="/users"
          hx-post="/users" 
          hx-target="#user-list" 
          hx-swap="afterbegin">
        @FormField("name", props.Name, "")
        <button type="submit">Save User</button>
    </form>
}
```

#### 3.2 Validation Strategy
Server-side validation with client enhancement:

```go
func (h *Handler) ValidateUserForm(w http.ResponseWriter, r *http.Request) {
    form := parseUserForm(r)
    errors := h.validator.Validate(form)
    
    if len(errors) > 0 {
        if isHTMXRequest(r) {
            // HTMX: Return form with errors
            w.Header().Set("HX-Reswap", "outerHTML")
            w.WriteHeader(422)
            tmpl := components.UserFormWithErrors(form, errors)
            tmpl.Render(r.Context(), w)
        } else {
            // Standard: Redirect with errors
            http.Redirect(w, r, "/users/new?errors="+encodeErrors(errors), 302)
        }
        return
    }
    
    // Success path...
}
```

### Phase 4: Performance Optimization (Week 6)
**Objective**: Achieve <50KB JavaScript footprint

#### 4.1 Asset Optimization
```bash
# Remove TypeScript compilation
# Keep only essential Alpine.js stores and utilities
# Total JavaScript target: <50KB

static/js/
├── alpine.min.js         # ~15KB
├── htmx.min.js          # ~10KB
├── stores/              # ~10KB total
└── utils/               # ~10KB total
Total: ~45KB
```

#### 4.2 Build Pipeline Update
```yaml
# Updated build process
stages:
  - templ_generate:
      command: "templ generate"
  - css_optimize:
      command: "tailwindcss -i static/css/main.css -o dist/main.css --minify"
  - js_bundle:
      command: "cat static/js/**/*.js | uglifyjs -c -m > dist/app.min.js"
  - go_build:
      command: "go build -ldflags='-s -w' -o bin/server"
```

## Implementation Checklist

### Phase 1: Foundation ✅ COMPLETED
- [x] Create atomic design file structure
- [x] Implement design system foundations (8 CSS files)
- [x] Create atomic components (8 Templ components)
- [ ] Implement Alpine.js stores 
- [ ] Set up build pipeline

**Completed Work**:
- **Foundation Layer**: Complete design system with colors, typography, spacing, shadows, borders, breakpoints, reset, and utilities
- **Atomic Components**: Button, Input, Textarea, Select, Checkbox, Radio, Toggle, Icon, Spinner
- **Flowbite Integration**: All components follow Flowbite design patterns
- **Server-First Architecture**: Full Templ + HTMX + Alpine.js integration
- **Accessibility**: ARIA labels, keyboard navigation, screen reader support
- **Validation States**: Error, success, and default states for all form components
- **Dark Theme Support**: Complete dark mode styling

### Phase 2: Component Migration ✅ COMPLETED
- [x] Convert Form components (molecules) - Field, FieldGroup, Search
- [x] Build molecule compositions (cards, alerts, dropdowns)
- [x] Convert DataTable component (organism) with filtering system
- [x] Convert Modal components (organism) with HTMX integration
- [x] Convert Navigation components (organism) - SiteHeader, Sidebar
- [x] Implement feature modules (user-management, dashboard)
- [x] Create layout template system (base, app, auth, minimal)
- [x] Implement HTMX integration patterns throughout

**Completed Components**:

**Molecules (6 components)**:
- **field.templ**: Universal form field wrapper with validation states
- **field-group.templ**: Related field grouping with collapsible functionality
- **search.templ**: Enhanced search with HTMX integration and debouncing
- **base-card.templ**: Flexible card container with header/body/footer
- **alert.templ**: User feedback messages with actions and auto-dismiss
- **dropdown.templ**: Interactive menu system with search and keyboard navigation

**Organisms (5 components)**:
- **table/**: Complete data table system with sorting, pagination, bulk actions, HTMX filtering
- **filter-panel/**: Advanced filtering system with presets, date ranges, and real-time updates
- **siteheader/**: Navigation header with user menu, notifications, and responsive design
- **sidebar/**: Collapsible navigation with role-based menu items and hierarchical structure
- **modal/**: Complete modal system with HTMX content loading, form handling, and Alpine.js state

**Feature Modules**:
- **user-management/**: 5 components (user-table, user-form, user-profile, role-selector, permissions-grid)
- **dashboard/**: 3 components (summary-cards, recent-activity, quick-actions)

**Layout System**:
- **base.templ**: Core HTML structure with HTMX/Alpine.js integration
- **app.templ**: Main application layout with sidebar and header
- **auth.templ**: Authentication pages layout with social providers
- **minimal.templ**: Clean layout for simple pages and error states

### Phase 3: Progressive Enhancement ✅ COMPLETED  
- [x] Ensure all forms work without JS (server-first architecture)
- [x] Implement server-side validation patterns
- [x] Add HTMX enhancements throughout all components
- [x] Test graceful degradation (all components work without JavaScript)
- [x] Resolve template function naming conflicts with Go types
- [x] Complete organism layer with modal and filter-panel systems
- [x] Finalize feature module implementation for user-management and dashboard

### Phase 4: Optimization ⏳ READY FOR IMPLEMENTATION
- [ ] Remove TypeScript compilation from build pipeline
- [ ] Optimize JavaScript bundle (<50KB target)
- [ ] Implement Alpine.js stores for state management
- [ ] Performance testing and monitoring

## Success Criteria

### Technical Metrics
- JavaScript bundle size: <50KB
- Time to interactive: <2s
- Server response time: <200ms
- Progressive enhancement: 100% functional without JS

### Architecture Compliance
- ✅ Server-first architecture
- ✅ Progressive enhancement
- ✅ Atomic design structure
- ✅ HTMX integration
- ✅ Minimal JavaScript footprint

## Risk Mitigation

### Development Velocity
- Implement changes incrementally
- Maintain parallel old/new components during transition
- Feature flags for gradual rollout

### Quality Assurance
- Automated testing at each phase
- Manual testing of progressive enhancement
- Performance monitoring

### Team Coordination
- Daily architecture alignment checks
- Code review focusing on architecture compliance
- Documentation updates with each phase

## Timeline

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| 1 | Week 1-2 | File structure, Alpine stores, build pipeline |
| 2 | Week 3-4 | Component migration, HTMX integration |
| 3 | Week 5 | Progressive enhancement, validation |
| 4 | Week 6 | Performance optimization, testing |

Total Duration: 6 weeks
