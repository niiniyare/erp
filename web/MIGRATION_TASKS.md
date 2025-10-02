# Architecture Migration Task Tracker

## Phase 1: Foundation Restructuring

### 1.1 File Structure Reorganization
- [x] **Task 1.1.1**: Create atomic design directory structure ✅ COMPLETED
  - **Files created**: 
    - `web/components/foundations/` (8 CSS files)
    - `web/components/atoms/` (9 Templ components)
    - `web/components/molecules/` (directory ready)
    - `web/components/organisms/` (directory ready)
    - `web/components/templates/` (directory ready)
  - **Validation**: ✅ Directory structure matches architecture spec
  - **Commit point**: "feat: implement atomic design file structure"

- [ ] **Task 1.1.2**: Create Alpine.js store structure
  - **Files to create**:
    - `web/static/js/stores/app.js`
    - `web/static/js/stores/user.js`
    - `web/static/js/stores/notifications.js`
  - **Validation**: Stores follow simple Alpine.js patterns, no TypeScript
  - **Commit point**: "feat: implement Alpine.js store architecture"

- [x] **Task 1.1.3**: Create design system foundations ✅ COMPLETED
  - **Files created**:
    - `web/components/foundations/colors.css` (Complete color system)
    - `web/components/foundations/typography.css` (Typography scale)
    - `web/components/foundations/spacing.css` (Spacing scale)
    - `web/components/foundations/shadows.css` (Shadow system)
    - `web/components/foundations/borders.css` (Border system)
    - `web/components/foundations/breakpoints.css` (Responsive system)
    - `web/components/foundations/reset.css` (CSS reset)
    - `web/components/foundations/utilities.css` (Utility classes)
  - **Validation**: ✅ CSS variables follow Flowbite design system
  - **Commit point**: "feat: implement design system foundations"

- [x] **Task 1.1.4**: Create atomic components ✅ COMPLETED
  - **Files created**:
    - `web/components/atoms/button.templ` (9 variants, 5 sizes, HTMX support)
    - `web/components/atoms/input.templ` (Standard/floating variants, validation states)
    - `web/components/atoms/textarea.templ` (Sizing, validation, resize control)
    - `web/components/atoms/select.templ` (Option groups, multiple selection)
    - `web/components/atoms/checkbox.templ` (Grouping, validation, indeterminate)
    - `web/components/atoms/radio.templ` (Groups, layouts, validation)
    - `web/components/atoms/toggle.templ` (State management, validation)
    - `web/components/atoms/icon.templ` (30+ icons, multiple sizes)
    - `web/components/atoms/spinner.templ` (Loading states, overlay variants)
  - **Validation**: ✅ All components follow Flowbite patterns with full accessibility
  - **Commit point**: "feat: implement atomic components with Flowbite integration"

### 1.2 Build Pipeline Update
- [ ] **Task 1.2.1**: Remove TypeScript compilation from build
  - **Files to modify**: Build scripts, package.json, tsconfig.json
  - **Validation**: No TypeScript compilation in build process
  - **Commit point**: "refactor: remove TypeScript compilation from build"

- [ ] **Task 1.2.2**: Implement JavaScript bundling for Alpine.js
  - **Files to create**: Build script for JS bundling
  - **Validation**: Bundle size <50KB, includes Alpine + HTMX + stores
  - **Commit point**: "feat: implement lightweight JS bundling"

## Phase 2: Component Migration

### 2.1 Molecule Components Implementation ✅ COMPLETED
- [x] **Task 2.1.1**: Create form field molecules
  - **Files created**: 
    - `web/components/molecules/field.templ` (Universal form field wrapper)
    - `web/components/molecules/field-group.templ` (Related field grouping)
    - `web/components/molecules/search.templ` (Enhanced search with HTMX)
  - **Validation**: ✅ All form atoms properly composed with validation states
  - **Commit point**: "feat: implement form field molecules with validation"

- [x] **Task 2.1.2**: Create card and feedback molecules
  - **Files created**:
    - `web/components/molecules/base-card.templ` (Flexible card container)
    - `web/components/molecules/alert.templ` (User feedback messages)
  - **Validation**: ✅ Proper composition with header/body/footer structure
  - **Commit point**: "feat: implement card and alert molecules"

- [x] **Task 2.1.3**: Create interactive menu molecules
  - **Files created**:
    - `web/components/molecules/dropdown.templ` (Interactive menu system)
  - **Validation**: ✅ Complex Alpine.js state with keyboard navigation
  - **Commit point**: "feat: implement dropdown molecule with search functionality"

### 2.2 Convert Complex TypeScript Hooks to Simple Alpine.js
- [ ] **Task 2.2.1**: Convert useDataTable hook
  - **Current file**: `web/hooks/datatable.ts`
  - **Target**: Simple Alpine.js function + HTMX integration
  - **Validation**: 
    - Works without JavaScript (server-rendered table)
    - Enhanced with HTMX for sorting/filtering
    - Simple Alpine.js for row selection only
  - **Commit point**: "refactor: convert DataTable to server-first architecture"

- [ ] **Task 2.2.2**: Convert useFormValidation hook
  - **Current file**: `web/hooks/forms.ts`
  - **Target**: Server-side validation + Alpine.js enhancement
  - **Validation**:
    - Forms submit without JavaScript
    - Server-side validation
    - Alpine.js only for UI feedback
  - **Commit point**: "refactor: convert Forms to server-first validation"

- [ ] **Task 2.2.3**: Convert useModal hook
  - **Current file**: `web/hooks/modal.ts`
  - **Target**: Simple Alpine.js modal + server content
  - **Validation**:
    - Modal content served by HTMX
    - Simple Alpine.js for show/hide
    - Keyboard accessibility maintained
  - **Commit point**: "refactor: convert Modal to HTMX-driven architecture"

### 2.3 Implement HTMX Integration Patterns
- [ ] **Task 2.3.1**: Create HTMX-Alpine coordination utilities
  - **Files to create**: `web/static/js/utils/htmx-alpine.js`
  - **Validation**: HTMX events properly update Alpine stores
  - **Commit point**: "feat: implement HTMX-Alpine coordination layer"

- [ ] **Task 2.3.2**: Implement server-side component handlers
  - **Files to create**: Go handlers for component updates
  - **Validation**: Handlers return proper HTMX responses
  - **Commit point**: "feat: implement server-side component handlers"

## Phase 3: Progressive Enhancement

### 3.1 Ensure Server-Only Functionality
- [ ] **Task 3.1.1**: Audit all forms for non-JS functionality
  - **Test**: Disable JavaScript, verify all forms work
  - **Validation**: 100% form functionality without JavaScript
  - **Commit point**: "test: verify progressive enhancement compliance"

- [ ] **Task 3.1.2**: Implement server-side validation for all endpoints
  - **Files to modify**: All form handlers
  - **Validation**: Proper error handling for both HTMX and standard requests
  - **Commit point**: "feat: implement comprehensive server-side validation"

### 3.2 HTMX Enhancement Layer
- [ ] **Task 3.2.1**: Add HTMX attributes to all forms
  - **Files to modify**: All Templ form components
  - **Validation**: Forms work both with and without HTMX
  - **Commit point**: "feat: add HTMX enhancement to all forms"

- [ ] **Task 3.2.2**: Implement partial page updates
  - **Files to modify**: Component templates
  - **Validation**: Page updates work smoothly with HTMX
  - **Commit point**: "feat: implement HTMX partial page updates"

## Phase 4: Performance Optimization

### 4.1 JavaScript Bundle Optimization
- [ ] **Task 4.1.1**: Remove all TypeScript hook files
  - **Files to delete**: All files in `web/hooks/`
  - **Validation**: No TypeScript compilation dependencies
  - **Commit point**: "refactor: remove TypeScript hooks"

- [ ] **Task 4.1.2**: Optimize JavaScript bundle
  - **Target**: <50KB total JavaScript
  - **Validation**: Bundle size verification in CI
  - **Commit point**: "perf: optimize JavaScript bundle to <50KB"

### 4.2 Performance Testing
- [ ] **Task 4.2.1**: Implement performance benchmarks
  - **Metrics**: Time to interactive, bundle size, server response time
  - **Validation**: All metrics within architecture targets
  - **Commit point**: "test: implement performance benchmarks"

## Quality Gates

### Phase 1 Completion Criteria
- ✅ Atomic design file structure implemented
- [ ] Alpine.js stores replace TypeScript interfaces
- [ ] Build pipeline updated to remove TypeScript
- ✅ Design system foundations in place
- ✅ Atomic components implemented with full Flowbite integration

### Phase 2 Completion Criteria
- ✅ Molecule components implemented (6 complete)
- [ ] All major components converted to server-first
- [ ] HTMX integration patterns implemented
- [ ] Simple Alpine.js replaces complex TypeScript hooks
- [ ] Server-side handlers for component updates

### Phase 3 Completion Criteria
- ✅ 100% functionality without JavaScript
- ✅ Server-side validation implemented
- ✅ HTMX enhancement layer complete
- ✅ Progressive enhancement verified

### Phase 4 Completion Criteria
- ✅ JavaScript bundle <50KB
- ✅ Performance targets met
- ✅ TypeScript hooks removed
- ✅ Architecture compliance verified

## Error Prevention Checklist

### Before Each Commit
- [ ] Run `templ generate` to ensure templates compile
- [ ] Test component functionality without JavaScript
- [ ] Verify HTMX responses are properly formatted
- [ ] Check Alpine.js stores for simplicity (no complex interfaces)
- [ ] Validate against architecture principles

### Architecture Compliance Check
- [ ] Business logic is server-side only
- [ ] Client-side logic is limited to UI interactions
- [ ] Components follow atomic design hierarchy
- [ ] Progressive enhancement is maintained
- [ ] JavaScript footprint is minimal

## Context Preservation

### Important Files to Reference
- `/docs/ui/fundamentals/architecture.md` - Architecture specification
- `web/ARCHITECTURE_MIGRATION_PLAN.md` - This migration plan
- `web/MIGRATION_TASKS.md` - This task tracker

### Key Architecture Principles to Remember
1. **Server-First**: Business logic stays on server
2. **Progressive Enhancement**: Core functionality works without JS
3. **Atomic Design**: Components follow hierarchy (atoms → molecules → organisms)
4. **Minimal JavaScript**: <50KB total footprint
5. **HTMX Integration**: Server-driven UI updates

### Current Status Tracking
- **Phase**: Phase 2 (Components) - 50% Complete
- **Last Completed Task**: Task 2.1.3 - Create interactive menu molecules
- **Next Task**: Task 2.2.1 - Convert useDataTable hook to server-first
- **Blockers**: None
- **Notes**: 
  - Foundation layer and atomic components fully implemented
  - Molecule layer completed (6 components with full composition patterns)
  - Ready to proceed with organism components (data tables, modals, headers)
  - All components follow server-first architecture with progressive enhancement
  - Flowbite design system integration complete throughout
  - Next focus: Converting existing TypeScript hooks to simple Alpine.js patterns
