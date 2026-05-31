# ERP Design System & Components Guide

## Table of Contents
1. [Overview & Philosophy](#overview--philosophy)
2. [Design Tokens](#design-tokens)
3. [Core Components](#core-components)
4. [Composite Patterns](#composite-patterns)
5. [Page Templates](#page-templates)
6. [Adjustability & Variants](#adjustability--variants)
7. [Accessibility Guidelines](#accessibility-guidelines)

---

## Overview & Philosophy

### Goal
Create a **unified, flexible design system** that works across all ERP pages while remaining adjustable for specific use cases. The system supports complex business workflows, data-heavy interfaces, and collaborative features.

### Key Principles
- **Consistency First** — All pages share the same tokens and base components
- **Flexibility** — Components have variants and can be combined for different layouts
- **Scalability** — Easy to add new pages without redesigning the foundation
- **Accessibility** — WCAG AA compliant with proper contrast and keyboard navigation
- **Performance** — Minimal visual complexity, fast to load and interact with
- **Collaboration** — Built-in support for activity feeds, comments, and file uploads
- **Data Density** — Support both comfortable and compact viewing modes

### Design Approach
Build from atoms (basic components) → molecules (combined components) → organisms (complete sections) → templates (full pages).

---

## Design Tokens

### Color Palette

**Primary Colors**
- **Primary Blue**: #0F6FFF (main actions, links, highlights)
- **Primary Light**: #E7F1FF (backgrounds for primary elements, hover states)
- **Primary Teal**: #0B7E6B (secondary primary for CTAs, active states)

**Semantic Colors**
- **Success**: #10B981 (positive actions, completed status)
- **Success Light**: #D1F4E5 (success badge background)
- **Danger**: #EF4444 (destructive actions, error states)
- **Danger Light**: #FEE2E2 (danger badge background)
- **Warning**: #F59E0B (caution, alerts, pending status)
- **Warning Light**: #FEF3C7 (warning badge background)
- **Orange**: #FF8C42 (accent for specific statuses like "Opened", "In Research")

**Neutral Colors**
- **Text Primary**: #0F1724 (body text, headings)
- **Text Muted**: #667085 (secondary text, labels, hints)
- **Background (Page)**: #F6F7F8 (main background)
- **Surface (Cards)**: #FFFFFF (card backgrounds, modals)
- **Surface Hover**: #F9FAFB (hover state for rows, items)
- **Border**: #E6E9EE (dividers, input borders)
- **Badge Background**: #F3F4F6 (muted badge backgrounds)
- **Dark Background**: #2D3748 or #1F2937 (dark mode alternative)

### Typography Scale

**Heading Hierarchy**
- **H1**: 28px / 36px line-height (page titles)
- **H2**: 20px / 28px line-height (section headings, modal titles)
- **H3**: 16px / 24px line-height (subsection headings, card titles, form sections)

**Body Text**
- **Body**: 14px or 16px / 24px line-height (regular content, table text)
- **Body Small**: 14px / 20px line-height (descriptions, metadata)
- **Caption**: 12px / 16px line-height (labels, helper text, timestamps)

**Font Weight**
- Regular: 400 (body text)
- Medium: 500 (button labels, input labels)
- Semibold: 600 (card titles, emphasis, active navigation)
- Bold: 700 (headings, strong emphasis)

### Spacing Scale (in pixels)

Follow this scale consistently across all components:
- **4px** — tight spacing (between small elements)
- **8px** — small spacing (padding in compact components)
- **12px** — medium spacing (padding around content)
- **16px** — standard spacing (gap between sections)
- **24px** — large spacing (card padding, section gaps)
- **32px** — extra large spacing (major section separation)
- **48px** — page-level spacing (top-level padding)

### Border Radius

- **Small (4–6px)**: Input fields, buttons, small components, badges
- **Medium (8–10px)**: Cards, modals, medium components
- **Large (12px)**: Drawers, full-page containers

### Elevation & Shadows

**Card Shadow** (used on cards, chips, badges)
- `0 2px 6px rgba(15, 23, 36, 0.06)`

**Drawer/Modal Shadow** (used on overlays, popovers, drawers)
- `0 10px 30px rgba(15, 23, 36, 0.12)`

**Row Hover Shadow** (subtle elevation on hover)
- `0 1px 3px rgba(15, 23, 36, 0.04)`

---

## Core Components

### 1. Buttons

**Purpose**: Trigger actions or navigate to new states.

**Variants**

- **Primary Button**: Solid blue background (#0F6FFF), white text
  - Use for main actions (Create, Save, Submit)
  - Highest visual priority
  - Hover: 90% opacity or #0D5FD9

- **Secondary Button**: Light border (#E6E9EE), dark text
  - Use for supporting actions (Cancel, Edit, View, Export)
  - Medium visual priority

- **Ghost Button**: No background, dark text, underline on hover
  - Use for tertiary actions or inline actions
  - Lowest visual priority

- **Danger Button**: Solid red background (#EF4444), white text
  - Use for destructive actions (Delete, Remove)
  - High contrast to indicate risk

- **Success Button**: Solid green background (#10B981 or #0B7E6B), white text
  - Use for confirmatory actions (Approve, Create, Start)

**Sizes**

- **Small**: 6px vertical / 12px horizontal padding, 12px font size
  - Use in compact tables, inline actions

- **Medium** (default): 8px vertical / 16px horizontal padding, 14px font size
  - Use in toolbars, form actions, most UI

- **Large**: 12px vertical / 24px horizontal padding, 16px font size
  - Use for prominent CTAs (main form submissions)

**States**

- Default (as specified above)
- Hover: 90% opacity or lighter shade
- Active/Pressed: 80% opacity or darker shade
- Disabled: 50% opacity, cursor not-allowed, no hover effect
- Loading: Show spinner icon, disable interactions

**Spacing Rules**

- Space buttons 8px apart horizontally
- When in a group (e.g., "Cancel" and "Save"), place Primary on the right
- Minimum touch target: 44px × 44px

---

### 2. Input Fields

**Purpose**: Capture user data in forms and filters.

**Field Types**

- **Text Input**: Single-line text entry
- **Email Input**: Email validation enabled
- **Number Input**: Numeric entry only
- **Textarea**: Multi-line text entry (minimum 100px height)
- **Select Dropdown**: Choose from predefined options
- **Multi-select**: Select multiple options (shows as chips/tags)
- **Date Picker**: Single date selection with calendar
- **Date Range Picker**: Start and end dates
- **Checkbox**: Binary choice (can be grouped)
- **Radio Button**: Single choice from a list
- **Toggle**: On/off switch

**Sizing**

- **Small**: 24px height, 12px font, 8px padding
  - Compact tables, inline filters

- **Medium** (default): 32px height, 14px font, 12px padding
  - Forms, tables

- **Large**: 40px height, 16px font, 16px padding
  - Important form fields, accessibility

**States**

- **Default**: Border (#E6E9EE), placeholder text in muted color
- **Focused**: Border changes to primary blue (#0F6FFF), 2px blue outline
- **Filled**: Shows entered value in text color (#0F1724)
- **Error**: Border changes to danger red (#EF4444), error message below in small red text
- **Disabled**: 50% opacity, cursor not-allowed, no interactions

**Labels & Helpers**

- Always include a label above the input (Caption, 12px, medium weight)
- Optional/Required indicator: Add "(optional)" or "*" after label
- Helper text below field: use Caption size, muted text color
- Error message: same size as helper text, danger color

**Spacing Rules**

- Gap between label and input: 8px
- Gap between input and helper text: 4px
- Gap between form fields: 16px

---

### 3. Cards

**Purpose**: Group related content in a contained, scannable format.

**Structure**

- Padding: 16–24px on all sides
- Border Radius: 8–10px
- Background: White (#FFFFFF)
- Shadow: Card shadow (0 2px 6px rgba(15, 23, 36, 0.06))
- Border: Optional 1px border (#E6E9EE) for subtle definition

**Card Types**

- **Standard Card**: Title + content
- **KPI Card**: Metric name, large value, trend indicator (up/down arrow), percentage change
  - Example: "$275,825" with "↑ 28% vs Last month" in green
- **Lead Card**: Avatar + name + metadata + action buttons + progress indicator
- **Action Card**: Interactive card with hover state (scale 1.02x, shadow increase) leading to detail view
- **Empty State Card**: Centered icon + message + CTA button
- **Status Card**: Icon + count + label + trend indicator
  - Used for status summaries (e.g., "On hold: 8", "Completed: 301")

**Content Rules**

- Heading (H3): 16px, semibold, top of card
- Body text (Body): 14–16px, regular, dark text
- Metadata (Caption): 12px, muted text
- Icons: 16px or 24px, aligned left or right
- Action buttons: Placed bottom-right or inline with metadata
- Badges/tags: Use appropriate semantic colors

**Spacing Inside Card**

- Title to content: 12px gap
- Between content sections: 16px gap
- Icon to text: 8px gap
- Metadata to action buttons: 12px gap

---

### 4. Tables

**Purpose**: Display structured data in rows and columns.

**Structure**

- Header row: Semibold text, light gray background (#F6F7F8), 12px caption font, 12px padding
- Data rows: Regular text, 14px body font, 12px padding
- Row height: 44px (default), 32px (compact mode)
- Border-bottom on each row: 1px (#E6E9EE)

**Features to Include**

- **Column Headers**: Semibold, sortable (click to sort ascending/descending), icon shows sort direction
- **Sticky Header**: Header remains visible when scrolling
- **Row Hover**: Light gray background (#F9FAFB) on hover, subtle shadow
- **Row Selection**: Checkbox in first column, enables bulk actions
- **Bulk Action Bar**: Sticky bar at top showing "X selected" + actions (Edit, Delete, Export)
- **Empty State**: Centered message + icon + "Create first item" CTA
- **Loading State**: Skeleton rows (gray placeholder bars) while loading
- **Status Indicators**: Colored badges, icons, or text to show row status
- **Pagination Footer**: "Showing 1–50 of 150 items" with Previous/Next buttons

**Actions Column**

- Use icon-only buttons (edit, delete, more)
- Icons: 16px size, muted text color on default, primary color on hover
- "More" menu (3-dot icon): Opens dropdown with additional actions
- Minimum spacing between action buttons: 8px
- Only show 2–3 primary actions; use "More" menu for additional options

**Progress Indicator in Tables**

- Show as percentage badge or progress bar within row
- Color-coded: warning (yellow) for in-progress, success (green) for near-completion

**Responsive Behavior**

- Desktop: Full table with all columns visible
- Tablet: Hide less important columns, combine into "Details" action
- Mobile: Convert to stacked card layout (one row per card)

**Pagination/Scroll**

- Show row count: "Showing 1–50 of 150 items"
- Include pagination buttons (Previous, Next) or infinite scroll
- Optional: Allow user to select rows per page (25, 50, 100)

---

### 5. Dropdowns & Select

**Purpose**: Allow users to select from a list of options.

**Trigger Button**

- Style: Secondary button, ghost button, or input field with down arrow
- Label: Shows selected option or placeholder
- Icon: Chevron down (16px), rotates 180° when open

**Dropdown Menu**

- Width: Match trigger button or be wider (at least 200px)
- Max height: ~300px (with scroll if longer)
- Position: Below trigger, align left by default
- Background: White (#FFFFFF)
- Border: 1px (#E6E9EE)
- Shadow: Drawer shadow (0 10px 30px...)
- Border radius: 8–10px

**Menu Items**

- Padding: 12px horizontal × 8px vertical
- Font: Body text (14px)
- Background on hover: Light gray (#F9FAFB)
- Background on selected: Primary light (#E7F1FF)
- Icon support: Show checkmark (✓) for selected items
- Disabled items: 50% opacity, no hover effect

**Search in Dropdown** (if many options)

- Include search input at top of menu
- Auto-filters options as user types
- Keyboard support: Arrow keys to navigate, Enter to select, Esc to close

**Multi-select Dropdown**

- Shows selected items as pills/chips before dropdown
- Pill styling: Light background (#F3F4F6), dark text, removable (X button)
- "Selected X items" text shown if space is limited
- Hover on truncated pills shows tooltip with full list

---

### 6. Modals & Drawers

**Purpose**: Present focused content or forms that require user attention.

**Modal (Full-screen overlay)**

- Center on screen
- Width: 90% on tablet/mobile, 500–600px on desktop
- Max height: 90vh with scroll if overflow
- Background: White (#FFFFFF)
- Border radius: 8–10px
- Shadow: Drawer shadow
- Scrim (backdrop): Semi-transparent black (50% opacity)

**Drawer (Side panel)**

- Position: Right side of screen (or left if needed)
- Width: 66% on desktop (can be 40% or 100% if needed)
- Height: Full viewport
- Background: White (#FFFFFF)
- Shadow: Drawer shadow, cast on left edge
- No border radius (full-height)
- Smooth slide-in animation (250–300ms)

**Header Section**

- Title (H2): 20px, bold, dark text
- Close button (X icon): 20px, positioned top-right, cursor pointer
- Border-bottom: 1px (#E6E9EE)
- Padding: 24px

**Content Section**

- Padding: 24px on all sides
- Scrollable if content exceeds available space
- Sections can be collapsible (show/hide with toggle)
- Form sections separated by dividers

**Footer/Actions Section**

- Border-top: 1px (#E6E9EE)
- Padding: 24px
- Buttons: Sticky to bottom, right-aligned
- Primary CTA (Save/Submit) on right, Secondary (Cancel) to its left
- Spacing between buttons: 12px

**Interactions**

- Close on ESC key (keyboard)
- Close on scrim click (for modal, if appropriate)
- Prevent body scroll when modal/drawer is open
- Focus trap: Tab only cycles through controls in modal/drawer

---

### 7. Filter Chips & Tags / Status Badges

**Purpose**: Display applied filters, categories, metadata, and status in a compact format.

**Filter Chip (Applied filter)**

- Background: Primary light (#E7F1FF)
- Text: Primary blue (#0F6FFF)
- Padding: 6px horizontal × 4px vertical
- Border radius: 6px
- Font: Caption (12px, medium weight)
- Remove button (X): 12px icon, appears on hover or inline

**Status Badge**

- **Active**: Success color (#10B981 or #0B7E6B)
- **In Progress**: Primary blue (#0F6FFF)
- **On Hold**: Warning color (#F59E0B)
- **Completed**: Success color (#10B981)
- **Cancelled/Inactive**: Danger color (#EF4444)
- **Draft**: Neutral (#667085)

Badge Styling:
- Padding: 6px horizontal × 3px vertical
- Border radius: 4–6px (slightly rounded) or 20px (pill shape)
- Font: Caption (12px, medium weight)
- With Icon: Icon (12–16px) + text, icon on left or can be icon-only

**Color-coded Status System**

- **Orange/Amber (#FF8C42)**: "Opened", "Replied", "In Research"
- **Green (#10B981)**: "Booked", "Completed", "Active"
- **Red (#EF4444)**: "Ignored", "Cancelled", "Inactive"
- **Blue (#0F6FFF)**: "In Progress"

**Icon + Status Combination**

- Use icons to enhance status understanding (mail icon for "Opened", check for "Completed", etc.)
- Icons should align with semantic meaning

---

### 8. Navigation

**Sidebar Navigation**

- Width: 240–280px
- Background: White (#FFFFFF)
- Border-right: 1px (#E6E9EE)
- Logo/branding at top: 20px, bold, primary color, 16–24px padding
- Fixed or collapsible (hamburger menu on mobile)

**Navigation Items**

- Height: 40px
- Padding: 12px horizontal, 8px vertical
- Icon (optional): 18–20px left-aligned
- Label: Body text (14–16px)
- Spacing between items: 4px
- Can be nested/hierarchical (parent → children)

States:
- **Default**: Text color #0F1724, no background
- **Hover**: Background #F9FAFB, text stays dark
- **Active**: Background #E7F1FF, text color #0F6FFF, bold label
- **Disabled**: 50% opacity, no hover

**Top Navigation Bar**

- Height: 64px
- Background: White (#FFFFFF)
- Border-bottom: 1px (#E6E9EE)
- Left: Logo, breadcrumb, or section title
- Center: Search bar (optional) or page title
- Right: Notifications icon, user menu, settings
- Padding: 16px horizontal

**Breadcrumb**

- Format: "Home > Sales Center > Contacts"
- Separator: "/" or ">" icon
- Last item: Bold or primary color (current page)
- Previous items: Muted text, clickable/interactive
- Font: Caption (12px)

**Tab Navigation**

- Horizontal list of tabs
- Height: 44–48px
- Text: Body font (14–16px, medium weight)
- Border-bottom on active tab: 2–3px primary color
- Padding: 12px horizontal per tab
- Spacing between tabs: 16–24px
- Can be scrollable horizontally on mobile

**Notifications Badge**

- Small red circle with white number (e.g., "3")
- Positioned top-right of icon
- Size: 18–24px
- Shows unread count

---

### 9. Activity Feed & Timeline

**Purpose**: Show historical actions, comments, and events in chronological order.

**Activity Item Structure**

- Avatar (24–32px)
- User name + action (bold) + entity name (linked)
- Timestamp (caption text, muted)
- Optional: Icon indicating action type (comment, edit, upload)
- Optional: Attached files or content preview

**Activity Types**

- **Status change**: "John changed status of 'Project X' from 'To Do' to 'In Progress'"
- **Comment**: Comment text displayed below user info
- **File upload**: File icon + filename + file size
- **Mention**: "Jane mentioned you in Project X"
- **Assignment**: "John assigned you to Task Y"

**Spacing & Layout**

- Padding: 12px horizontal, 8px vertical per item
- Gap between items: 8–12px
- Avatar to text: 12px gap
- List is scrollable; most recent at top

**States & Interactions**

- Hover on item: slight background change (#F9FAFB)
- Click on linked entity: navigate to detail view
- File download: icon with download action

---

### 10. Data Visualization

**Purpose**: Display metrics, trends, and performance data.

**KPI Cards in Dashboard**

- Large metric value (H1 size, 28–32px)
- Metric name (Caption, muted)
- Trend indicator: up/down arrow + percentage
  - Up trend: Green (#10B981)
  - Down trend: Red (#EF4444)
  - Example: "↑ 28% vs Last month"

**Charts & Graphs**

- Use recharts or Chart.js for consistency
- Color palette: Primary blue, teal, green, orange, red
- Tooltips on hover showing exact values
- Legend below or to side of chart
- Responsive: stack vertically on mobile

**Progress Bars**

- Full width: 100%, height: 8px
- Color: Primary blue (#0F6FFF) or semantic color
- Background: Light gray (#F3F4F6)
- Border radius: 4px
- Optional: Percentage text beside bar

**Bar Charts for Status**

- Stacked bars showing different statuses (e.g., "Order Processing", "Shipped/In Transit", "Delivered")
- Each status has its own color
- Labels on bars or legend
- Example: Order fulfillment tracking with color-coded segments

---

## Composite Patterns

### Pattern 1: Data Display with Filters

**Structure**

1. Heading + CTA button (top-right)
2. Controls bar:
   - Date range picker or filter chip (left)
   - Search input (center-left)
   - View toggle buttons (table/grid/board, center-right)
   - Export button (right)
   - Bulk action bar (if rows selected)
3. Data section:
   - Table or grid view with status indicators
   - Empty state if no data
4. Pagination footer

**Spacing**

- Heading to controls: 24px
- Controls to data: 16px
- Between control elements: 12px

**Interactions**

- Click filter → dropdown opens, shows filter options with checkboxes
- Select filters → table updates in real-time
- Click row → drawer or detail page opens
- Select checkbox → bulk action bar appears (sticky at top)
- Click "Delete" in bulk bar → confirm modal appears
- Export → CSV or PDF download

**Status Row Example**

- Date column with date range and sorting
- Status column with colored badges (Opened, Replied, Ignored, Booked)
- Company column with icon + text
- Tags/Category column with filter chips
- Last interaction timestamp

---

### Pattern 2: Form in Drawer

**Structure**

1. Header: "Create/Edit [Entity]"
2. Sections (collapsible):
   - Section title (H3, semibold)
   - Form fields
   - Divider between sections
3. Footer: "Cancel" and "Submit" buttons (sticky)

**Form Fields**

- Group related fields in sections (e.g., "Campaign Info", "Audience", "Time Manage", "Create Rules")
- Label above each field (not placeholder)
- Helper text below field if needed
- Error message replaces helper text when validation fails
- Multi-select fields show selected items as chips/tags

**Validation**

- Validate on blur (field loses focus)
- Show inline error message in red
- Disable submit button if form has errors
- Show summary of errors at top of form if multiple errors
- Required fields marked with * or "(required)"

**Multi-column Layout**

- 2 columns for wider forms (e.g., "Brand" and "Channel" side-by-side)
- Stack to 1 column on tablet/mobile

---

### Pattern 3: KPI Dashboard

**Structure**

1. Row of KPI cards (typically 3–4 in desktop view)
2. Each KPI shows:
   - Metric icon (optional)
   - Metric name (Caption, muted)
   - Large value (H1 or very large font, 28–32px)
   - Trend indicator (up/down arrow + percentage, semantic color)

**Example KPI Card**

- Icon: Dollar sign or chart icon
- Name: "Total Sales"
- Value: "$275,825.00"
- Trend: "↑ 28% vs Last month" (in green)

**Responsive**

- Desktop (3–4 columns): All KPIs in one row
- Tablet (2 columns): 2 KPIs per row
- Mobile (1 column): 1 KPI per row, full width

**Interactions**

- Hover on KPI → slight shadow increase or background color change
- Click on KPI → navigate to detail view or drill-down report

---

### Pattern 4: Card Grid (Lead/Contact View)

**Structure**

- Grid layout: 3 columns on desktop, 2 on tablet, 1 on mobile
- Each card shows:
  - Avatar (40–48px) or initials in colored circle
  - Title/Name (H3, semibold)
  - Contact info: email icon + email, phone icon + phone, location
  - Metadata (2–3 lines of Caption text)
  - Badge or progress indicator (status, priority)
  - Action buttons (edit, message, call, more menu) at bottom

**Spacing**

- Gap between cards: 16px
- Card padding: 16px internal
- Title to content: 8px
- Content sections: 8–12px gap
- Metadata to actions: 12px

**Interactions**

- Hover on card → shadow deepens, scale slightly (1.02x), background lightens
- Click card → opens detail view or drawer with full information
- Click action button → performs inline action (call, message, edit)
- "Show more →" link → expands card or navigates to full details

---

### Pattern 5: Task/Project Board View

**Structure**

- Kanban-style columns: "To Do", "In Progress", "Done", etc.
- Each column contains cards with tasks
- Cards show:
  - Status badge or color-coded label
  - Task title (semibold)
  - Brief description
  - Assignee avatars (stacked)
  - Priority badge (color-coded: Low, Medium, High)
  - Due date
  - Comment count + attachment count
  - Metadata (creation date, tags)

**Card Styling**

- Padding: 12px
- Border radius: 8px
- Shadow: Card shadow
- Border: 1px subtle border
- Drag-and-drop enabled between columns

**Interactions**

- Hover on card → slight elevation increase
- Click card → opens drawer or modal with full details
- Drag card → moves between columns, updates status
- Click assignee avatar → shows team member info or can reassign

---

## Page Templates

### Template 1: Dashboard

**Page Layout**

1. Top section: Personalized greeting ("Hi, [Name]") + "Customize" button + "Add new" CTA
2. KPI row: 3–4 KPI cards showing key metrics
3. Chart row: 1–2 data visualizations (line chart, bar chart)
4. Secondary metrics: Order/Fulfillment tracking, Activity feed
5. Data section: Recent items table, top sellers, or trending data
6. Footer: "View all" link to detailed report page

**Example Dashboard**

- "Hi, Erin Amioff - Here's your business summary for this month"
- KPIs: Total Sales, Total Product, Total Customer
- Charts: Sales Performance (line chart), Order & Fulfillment Tracking (stacked bar)
- Tables: Top Selling products with status
- Activity: Recent changes, status updates

**Responsive**

- Desktop: All sections visible side-by-side
- Tablet: Stack sections vertically, reduce card grid to 2 columns
- Mobile: Further stacking, single column, minimize charts

---

### Template 2: List / Table View

**Page Layout**

1. Header: Page title + "New [Entity]" button (right-aligned)
2. Tab navigation: Opportunities, Accounts, Contacts, Leads (with counts)
3. Controls: Date range picker, Search, Filter, Sort by, View toggle
4. Status summary row: Status cards (e.g., "On hold: 8", "Completed: 301") with trends
5. Data: Table or grid with sorting, filtering, selection
6. Footer: Pagination or "Load more" button

**Example: Leads Page**

- Title: "Leads"
- Subtitle: "Who have shown interest in a product or service"
- Status cards: On hold (↓ 1.2%), Rejected (↓ 0.3%), Completed (↑ 2.9%), Cancelled (↓ 0.1%)
- Controls: Date filter, Filter button, Sort by dropdown, Search
- Card grid or list view with contact info, tags, progress indicators
- Each lead card: Avatar + name + email + phone + location + category + progress + action buttons

---

### Template 3: Detail View / Modal

**Page Layout**

1. Header: Entity name + back button + more menu (3-dot icon)
2. Tabs or sections:
   - Overview (main content, key fields)
   - Related data (linked entities)
   - Activity/Timeline (comments, changes)
   - Attachments (files, documents)
3. Main content area: Scrollable, displays all entity details
4. Side panel (optional): Summary info, quick actions, tags

**Example: Order Details**

- Header: "Order #AT456BB" with close button
- Sections: Items, Created at, Delivery Services, Payment method, Status, Customer name, Email, Phone
- Timeline: Order processing status (Order Processed → Payment Confirmed → Order Placed)
- Payment breakdown: Subtotal, Shipping fee, Total

---

### Template 4: Form/Editor

**Page Layout**

Option A: Full-page form
1. Heading + breadcrumb
2. Multi-section form (see Pattern 2 in Composite Patterns)
3. Sticky footer with "Cancel" and "Submit" buttons

Option B: Drawer form (recommended for inline edits)
1. Header in drawer (see Pattern 2 in Composite Patterns)
2. Sections with fields
3. Sticky footer in drawer with action buttons

**Example: Campaign Creation Form**

- Drawer title: "Create Campaign" or "Edit Campaign"
- Sections:
  - Campaign Info (name, brand, channel multi-select)
  - Description (textarea)
  - Audience (target customers, email only, SMS only, customer type dropdown)
  - Time Manage (check frequency, run length date range)
  - Create Rules (spend, increase budget) with add button
- Footer: Cancel button, Create Campaign button

---

### Template 5: Notification Center

**Page Layout**

1. Header: "Notification" title + close button (X)
2. Tabs: "View All", "Mentions", "Archive"
3. Activity list:
   - User avatar + action description + timestamp
   - Optional: Attached file preview
   - Optional: Action buttons (Deny, Approve)
4. Scrollable, most recent at top

**Example Notifications**

- "Jessie Joee comment in Facebook Campaign [In Progress]" - 12 minutes ago
- "Teo Le added file to WhatsApp Ads Campaign [Active]" - 44 minutes ago
  - File: "LezatikaFoods_MarketingAssets_Sept2024.zip" (3.1 MB)
- "Sarah requested access to Instagram Ads for Lezatos" - 56 minutes ago
  - Action buttons: Deny, Approve

---

## Adjustability & Variants

### Component Variants

**By Size**

- Small: Compact layouts, tables, inline actions
- Medium: Default, most UI
- Large: Prominent CTAs, accessibility-focused

**By Density**

- Comfortable: Spacious, good for data entry (default)
  - Row height: 44px
  - Padding: 16px
  - Gap: 16px

- Compact: Dense, good for data review and scanning
  - Row height: 32px
  - Padding: 12px
  - Gap: 12px

**By State**

- Default
- Hover
- Active/Selected
- Disabled
- Loading
- Error
- Success

### Theme Customization

**Brand Color Override**

If your brand has different primary/secondary colors:
1. Replace the primary blue (#0F6FFF) throughout (buttons, links, active states)
2. Replace the teal accent (#0B7E6B) for secondary CTAs
3. Adjust the semantic colors (success, warning, danger) if needed
4. Ensure contrast meets WCAG AA (4.5:1 for text)
5. Test all interactive states (hover, active, disabled)

**Dark Mode (Optional)**

If you want a dark theme:
1. Invert background colors:
   - Page background: #1F2937 or #2D3748
   - Card background: #111827 or #1A202C
   - Surface hover: #374151
2. Adjust text colors for readability:
   - Primary text: #F3F4F6 or #E5E7EB
   - Muted text: #D1D5DB or #9CA3AF
3. Reduce shadow intensity or increase contrast
4. Use lighter primary colors for better contrast against dark backgrounds
5. Maintain semantic color meanings but adjust brightness

### Responsive Adjustments

**Mobile Considerations**

- Increase touch targets to 44px minimum
- Stack multi-column layouts vertically
- Convert tables to card layouts
- Hide less critical information (show "more" button)
- Full-width inputs and buttons (margin: 0, width: 100%)
- Collapsible navigation (hamburger menu)
- Single-column grid layouts
- Simplified charts (show only key metrics)

**Tablet Considerations**

- 2-column layouts instead of 4
- Larger padding for easier touch interaction (16px minimum)
- Simplified navigation (drawer or hamburger menu)
- Adjust font sizes up slightly (14–16px for body text)
- Mid-size icons (18–20px)

**Desktop Considerations**

- Multi-column layouts (3–4 columns)
- Consistent padding (16–24px)
- Full sidebar navigation
- All features visible and accessible
- Optimized for keyboard navigation

---

## Accessibility Guidelines

### Color Contrast

- Text on background: Minimum 4.5:1 ratio (WCAG AA)
- Large text (18px+): Minimum 3:1 ratio
- Use semantic colors (red for danger, green for success) + text/icon to convey meaning
- Don't rely on color alone; use icons, patterns, or text labels
- Test contrast using tools like WebAIM or Contrast Checker

**Contrast Examples**

- Primary blue (#0F6FFF) on white (#FFFFFF): 5.1:1 ✓
- Text primary (#0F1724) on page background (#F6F7F8): 10.1:1 ✓
- Text muted (#667085) on white (#FFFFFF): 4.6:1 ✓

### Keyboard Navigation

- All interactive elements must be keyboard accessible (buttons, links, inputs, dropdowns)
- Tab order: Logical, left-to-right, top-to-bottom
- Focus indicator: 2–3px blue outline (#0F6FFF) with 2px offset, never remove it
- Escape key: Closes modals, drawers, dropdowns, menus
- Enter key: Activates buttons, submits forms, opens dropdowns
- Arrow keys: Navigate tabs, dropdowns, table rows, menu items
- Space key: Toggle checkboxes, activate buttons
- Skip to content link: Available on page top for keyboard users

### Focus States

- Visible focus indicator on all buttons, links, inputs, form controls
- Focus indicator color: Primary blue with high contrast
- Never remove focus indicator (even in hover state)
- Focus should be immediately visible and obvious
- Example: `outline: 3px solid #0F6FFF; outline-offset: 2px;`

### ARIA Labels & Semantics

- Form inputs: Associated with `<label>` element via `for` attribute
- Buttons: Clear text label or `aria-label` if icon-only
- Icons: `aria-label` for functional icons, `aria-hidden="true"` if decorative
- Modals: `aria-modal="true"`, focus trap on open
- Dropdowns: `aria-haspopup="listbox"`, `aria-expanded="true/false"`
- Tables: `<thead>`, `<tbody>`, `<th>` for headers, `role="columnheader"`
- Links: Descriptive text (not "click here")
- Error messages: Associated with input via `aria-describedby`
- Loading states: `aria-busy="true"` on loading elements
- Live regions: `aria-live="polite"` for status updates

### Text & Readability

- Use semantic HTML: `<h1>`, `<h2>`, `<h3>`, `<ul>`, `<ol>`, `<p>`, `<strong>`, `<em>`
- Line height: Minimum 1.5 (1.5× font size)
- Line length: Maximum 80 characters for readability (optimal: 50–75 chars)
- Font size: Minimum 14px for body text (16px preferred)
- Avoid all caps (harder to read than sentence case)
- Use bold for emphasis, not color alone
- Justify text left, not fully justified
- List items should be clear and scannable

### Motion & Animation

- Animations: 200–300ms duration (not too fast or slow)
- Easing: Use ease-out for entrance, ease-in for exit
- Avoid flashing (more than 3× per second)
- Provide "reduce motion" option for users who prefer reduced animation
- Respect `prefers-reduced-motion` media query
- Use motion to guide attention, not distract
- Auto-playing animations should have pause controls

**Example CSS for Reduced Motion**

```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation: none !important;
    transition: none !important;
  }
}
```

### Form Accessibility

- Label all form fields (visible label above, not placeholder)
- Use `<fieldset>` and `<legend>` for grouped fields
- Required fields marked with * and `aria-required="true"`
- Error messages linked to input via `aria-describedby`
- Error state visually distinct (red border, error icon, error text)
- Help text below input linked via `aria-describedby`
- Input types: email, number, tel, date, etc. for mobile keyboards
- Validate on blur (not on keystroke)

### Images & Icons

- All images should have `alt` text describing content
- Decorative images: `alt=""` or `aria-hidden="true"`
- Icons: Use `aria-label` if standalone, or part of labeled button
- Charts: Provide data table or text summary for screen readers

### Touch & Motor Considerations

- Minimum touch target: 44px × 44px (WCAG 2.5.5)
- Space between targets: 8px minimum
- Clickable area should be obvious and easy to hit
- Hover and click states should be the same size
- Use large buttons for critical actions
- Avoid precise dragging if possible; offer alternative input

---

## Implementation Checklist

### Before Building Each Page

- [ ] Define which template pattern applies
- [ ] List all components needed
- [ ] Check color contrast for all text (use WebAIM Contrast Checker)
- [ ] Plan responsive behavior (desktop/tablet/mobile breakpoints)
- [ ] Identify keyboard interactions needed (Tab, Enter, Escape, Arrow keys)
- [ ] Plan loading and empty states (spinners, skeletons, messaging)
- [ ] Identify error scenarios and error messages
- [ ] Plan focus management (where does focus go after action)
- [ ] Identify which elements need ARIA labels

### Quality Assurance - Design

- [ ] All text meets contrast requirements (4.5:1 minimum)
- [ ] Color is not the only way to convey information (also use icons, text, patterns)
- [ ] Responsive design works on mobile (375px), tablet (768px), desktop (1440px+)
- [ ] Touch targets are minimum 44px × 44px
- [ ] Loading states provided (spinners, skeleton loaders, progress indicators)
- [ ] Empty states are helpful (icon, message, CTA button)
- [ ] Error states are clear (red border, error message, icon)
- [ ] Hover and focus states are visible
- [ ] Modal/drawer appears with smooth animation
- [ ] Interactions feel responsive (no lag or delay)

### Quality Assurance - Accessibility

- [ ] Keyboard navigation works throughout (Tab, Enter, Escape, Arrow keys)
- [ ] Focus indicators visible on all interactive elements
- [ ] Focus order is logical (left-to-right, top-to-bottom)
- [ ] All form fields have associated labels
- [ ] Error messages linked to inputs via aria-describedby
- [ ] Modals are properly trapped (Tab cycles within modal)
- [ ] All text meets contrast requirements (test with WebAIM)
- [ ] Screen reader testing (NVDA, JAWS, or browser reader)
- [ ] No keyboard traps (user can always Tab away)
- [ ] Page titles are descriptive
- [ ] Links have descriptive text (not "click here")

### Quality Assurance - Functionality

- [ ] Forms validate and show clear error messages
- [ ] Links and buttons are clickable and functional
- [ ] No broken links or missing components
- [ ] Sorting, filtering, and searching work correctly
- [ ] Pagination works or infinite scroll functions
- [ ] File uploads work with size/type validation
- [ ] Images load and render correctly
- [ ] Charts and data visualizations load without errors
- [ ] Modals/drawers open and close properly
- [ ] Animations are smooth (60fps)

---

## Design Patterns Summary

### Common ERP Patterns

**Pattern: Data + Filters + Actions**
- Top controls for filtering, searching, sorting
- Main data display (table or grid)
- Bulk actions on row selection
- Pagination or infinite scroll

**Pattern: Master-Detail**
- List/table view on left or main area
- Detail panel/drawer on right or modal
- Quick edit capabilities
- Activity/timeline section

**Pattern: Dashboard**
- KPI cards showing key metrics
- Charts and visualizations
- Recent activity feed
- Quick action CTAs

**Pattern: Multi-step Form**
- Form sections or steps
- Progress indicator
- Previous/Next buttons
- Review step before submission

**Pattern: Workflow/Approval**
- Status cards showing stages
- Activity feed showing approvals
- Approval/rejection buttons
- Comments section for context

---

## File Organization

If building in Figma, use this structure:

- **00 - Tokens**: Color palette, typography scale, spacing, shadows, border radius
- **01 - Components**: Base components (atoms) organized by category
  - Buttons
  - Inputs
  - Cards
  - Tables
  - Navigation
  - Modals/Drawers
  - Badges/Chips
  - etc.
- **02 - Variants**: All component states and size variants
- **03 - Patterns**: Composite patterns (Data + Filters, Form in Drawer, KPI Dashboard, etc.)
- **04 - Templates**: Full page templates (Dashboard, List, Detail, Form, Notifications)
- **05 - Screens**: Actual screen designs with real data
- **06 - Prototype**: Interactive flows and transitions
- **07 - Handoff**: Spec sheets, documentation, CSS exports, tokens file

### Figma Library Setup

- Publish components for team use
- Version control for component updates
- Document component usage and dos/don'ts
- Share token specs (colors, typography, spacing)
- Include accessibility guidelines in comments
- Link to this documentation in library description

---

## Development Guidelines

### HTML/CSS Structure

Use semantic HTML tags and CSS utility classes (Tailwind, Bootstrap, or custom):

```html
<button class="btn btn--primary btn--medium">Save</button>
<button class="btn btn--secondary" disabled>Cancel</button>

<div class="card">
  <h3 class="card__title">KPI Title</h3>
  <p class="card__value">$275,825</p>
  <p class="card__trend trend--positive">↑ 28% vs Last month</p>
</div>

<div class="table__wrapper">
  <table class="table table--striped">
    <thead>
      <tr>
        <th>Column Name</th>
        <th>Column Name</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>Data</td>
        <td>Data</td>
      </tr>
    </tbody>
  </table>
</div>
```

### CSS Variables (Tokens)

```css
:root {
  --color-primary: #0F6FFF;
  --color-primary-light: #E7F1FF;
  --color-teal: #0B7E6B;
  --color-success: #10B981;
  --color-danger: #EF4444;
  --color-warning: #F59E0B;
  --color-text-primary: #0F1724;
  --color-text-muted: #667085;
  --color-background: #F6F7F8;
  --color-surface: #FFFFFF;
  --color-border: #E6E9EE;
  
  --font-size-h1: 28px;
  --font-size-h2: 20px;
  --font-size-h3: 16px;
  --font-size-body: 16px;
  --font-size-caption: 12px;
  
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 12px;
  --spacing-lg: 16px;
  --spacing-xl: 24px;
  --spacing-2xl: 32px;
  --spacing-3xl: 48px;
  
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 10px;
  
  --shadow-card: 0 2px 6px rgba(15, 23, 36, 0.06);
  --shadow-drawer: 0 10px 30px rgba(15, 23, 36, 0.12);
}
```

### Component Implementation

Each component should support:
- All documented variants (size, state, type)
- Accessibility attributes (aria-labels, roles, etc.)
- Responsive behavior
- Dark mode (if applicable)
- Animation and transitions

---

## Next Steps

1. **Validate** — Share this guide with your team for feedback and input
2. **Build** — Create components in Figma, HTML/CSS, or your preferred tool
3. **Document** — Add this guide + component specs to your design tool
4. **Develop** — Build components in your codebase (React, Vue, Web Components, etc.)
5. **Test** — Review for accessibility (keyboard nav, screen readers, contrast), responsiveness, and consistency
6. **Handoff** — Export tokens, CSS variables, component specs, and accessibility guidelines for developers
7. **Implement** — Integrate components into pages and workflows
8. **Iterate** — Update guide and components as new features or patterns are added
9. **Maintain** — Keep design system and code in sync as product evolves

### Rollout Strategy

- Phase 1: Core components (Buttons, Inputs, Cards, Navigation)
- Phase 2: Complex components (Tables, Modals, Drawers)
- Phase 3: Patterns and page templates
- Phase 4: Specialized components (Charts, Forms, Workflows)
- Phase 5: Advanced features (Dark mode, Responsive variants)

### Team Communication

- Document all decisions and reasoning
- Share updates and changes with team
- Gather feedback on designs and implementations
- Regular design system reviews
- Keep library and documentation up-to-date

---

**Version History**

- v1.0 (Sept 2025) — Initial design system created
- v2.0 (Current) — with real UI patterns, activity feeds, notifications, improved accessibility, and implementation guidelines

**Last Updated**: October 2025

