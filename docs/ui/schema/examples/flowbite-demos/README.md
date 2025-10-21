# Flowbite Component Examples

**FILE PURPOSE**: Organized examples of all Flowbite components for ERP UI  
**SCOPE**: Complete component library with working examples  
**TARGET AUDIENCE**: Developers implementing Flowbite components

## 📁 Component Categories

### 🧱 [Basic Components](basic/)
Essential UI elements and foundation components
- **[Buttons](basic/buttons.md)** - Primary, secondary, and action buttons
- **[Button Groups](basic/button-group.md)** - Grouped button collections
- **[Badges](basic/badge.md)** - Status indicators and labels
- **[Alerts](basic/alerts.md)** - Success, warning, error notifications
- **[Avatars](basic/avatar.md)** - User profile images and placeholders
- **[Spinners](basic/spinner.md)** - Loading indicators
- **[Progress Bars](basic/progress.md)** - Progress tracking
- **[Ratings](basic/rating.md)** - Star ratings and reviews
- **[Typography](basic/typography.md)** - Text styling and formatting
- **[Indicators](basic/indicators.md)** - Status and notification badges
- **[Keyboard Keys](basic/kbd.md)** - Keyboard key display

### 📝 [Form Components](forms/)
Interactive form elements and input validation
- **[Forms](forms/forms.md)** - Complete form layouts and validation
- **[Date Picker](forms/datepicker.md)** - Date and time selection
- **[Clipboard](forms/clipboard.md)** - Copy-to-clipboard functionality

### 📊 [Layout Components](layout/)
Structural components for page organization
- **[Cards](layout/card.md)** - Content containers and panels
- **[Tables](layout/tables.md)** - Data tables with sorting and filtering
- **[List Groups](layout/list-group.md)** - Vertical lists and menus
- **[Accordions](layout/accordion.md)** - Expandable content sections
- **[Tabs](layout/tabs.md)** - Tabbed content navigation
- **[Steppers](layout/stepper.md)** - Multi-step process indicators
- **[Timeline](layout/timeline.md)** - Chronological event display
- **[Gallery](layout/gallery.md)** - Image and media galleries
- **[Carousel](layout/carousel.md)** - Image and content sliders
- **[Pagination](layout/pagination.md)** - Page navigation controls
- **[Banner](layout/banner.md)** - Promotional and announcement banners
- **[Footer](layout/footer.md)** - Page footer layouts
- **[Jumbotron](layout/jumbotron.md)** - Hero sections and highlights

### 🧭 [Navigation Components](navigation/)
Site navigation and wayfinding elements
- **[Navbar](navigation/navbar.md)** - Top navigation bars
- **[Sidebar](navigation/sidebar.md)** - Side navigation menus
- **[Breadcrumb](navigation/breadcrumb.md)** - Hierarchical navigation
- **[Bottom Navigation](navigation/bottom-navigation.md)** - Mobile bottom navigation
- **[Mega Menu](navigation/mega-menu.md)** - Large dropdown menus

### 🚀 [Advanced Components](advanced/)
Complex interactive components
- **[Modals](advanced/modal.md)** - Dialog boxes and overlays
- **[Dropdowns](advanced/dropdowns.md)** - Dropdown menus and selects
- **[Drawer](advanced/drawer.md)** - Slide-out panels
- **[Popover](advanced/popover.md)** - Context-sensitive information
- **[Tooltips](advanced/tooltips.md)** - Hover help text
- **[Toast](advanced/toast.md)** - Temporary notifications
- **[Speed Dial](advanced/speed-dial.md)** - Floating action buttons
- **[Chat Bubble](advanced/chat-bubble.md)** - Messaging interfaces
- **[Skeleton](advanced/skeleton.md)** - Loading placeholders
- **[Device Mockups](advanced/device-mockups.md)** - Device frame presentations
- **[Video](advanced/video.md)** - Video player components

### 📈 [Data Tables](datatables/)
Advanced data table functionality with DataTables integration
- **[Getting Started](datatables/Getting-Started.md)** - Setup and basic usage
- **[API Reference](datatables/API.md)** - Complete API documentation
- **[Events](datatables/Events.md)** - Event handling and callbacks
- **[Options](datatables/Options.md)** - Configuration options
- **[Columns](datatables/columns.md)** - Column configuration
- **[Data Management](datatables/data.md)** - Data loading and manipulation

## 🚀 Quick Start

### Using Flowbite Components in ERP
```go
// 1. Import component props
import "github.com/niiniyare/erp/pkg/ui/components"

// 2. Create component props
props := components.ButtonProps{
    Text:    "Save Changes",
    Variant: components.ButtonPrimary,
    Size:    components.ButtonMD,
}

// 3. Render with schema factory
component, _ := factory.RenderToTempl(ctx, "FlowbiteButtonSchema", props)

// 4. Use in Templ template
templ MyPage() {
    <div class="p-6">
        @component
    </div>
}
```

### Schema-Driven Development
All Flowbite components are enhanced with:
- **Type Safety**: Go struct validation
- **Schema Validation**: JSON schema compliance
- **HTMX Integration**: Server-side interactivity
- **Multi-Tenant**: Tenant-aware styling

## 📚 Integration Guides

### Primary Integration
- **[Flowbite Integration](../../integration/flowbite.md)** - Complete integration guide
- **[Schema System](../../fundamentals/schema-system.md)** - Type-safe component generation
- **[Styling Approach](../../fundamentals/styling-approach.md)** - CSS architecture

### Technology Integration
- **[HTMX Patterns](../../patterns/htmx-integration.md)** - Server interaction patterns
- **[Templ Integration](../../fundamentals/templ-integration.md)** - Template development
- **[Validation Guide](../../guides/validation-guide.md)** - Form validation

## 🎨 Customization

### Brand Integration
All components support:
- **Theme Colors**: Custom brand colors
- **Typography**: Custom fonts and sizing
- **Spacing**: Consistent spacing system
- **Dark Mode**: Full dark mode support

### ERP-Specific Features
- **Multi-Tenant**: Tenant-aware styling
- **Security**: Input validation and sanitization
- **Performance**: Optimized for large datasets
- **Accessibility**: WCAG 2.1 AA compliance

## 📱 Responsive Design

All components are built with:
- **Mobile-First**: Optimized for mobile devices
- **Breakpoint System**: TailwindCSS responsive utilities
- **Touch-Friendly**: Appropriate touch targets
- **Progressive Enhancement**: Works without JavaScript

## 🧪 Testing

Component examples include:
- **Visual Tests**: Playwright visual regression tests
- **Unit Tests**: Go component unit tests
- **Integration Tests**: Full component integration tests
- **Accessibility Tests**: WCAG compliance validation

---

**Navigation**: [← Back to Integration](../../integration/) | [Quick Start →](../../quick-start/)