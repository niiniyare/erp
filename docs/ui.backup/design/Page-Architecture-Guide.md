# ERP Page Architecture Guide - Layout Patterns & Information Architecture

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Comprehensive page layout and information architecture guide for ERP interface development
**SCOPE**: Layout patterns, page structures, navigation patterns, and responsive architectures
**TARGET AUDIENCE**: UX developers, frontend architects, and product designers
<!-- LLM-CONTEXT-END -->

## 🏗️ Architecture Overview

This guide defines the **foundational page structures** and **layout patterns** for the ERP interface, providing **reusable templates** and **architectural guidance** for building consistent, scalable page layouts.

### **Core Architectural Principles**
- **Three-Zone Layout**: Navigation + Content + Context zones
- **Information Hierarchy**: Strategic → Tactical → Operational → Contextual
- **Progressive Disclosure**: Overview → Detail → Action flows
- **Content-First Design**: Layout serves information, not decoration

---

## 🎯 Master Layout Template

<!-- LLM-SECTION-MASTER-LAYOUT-START -->
### **Three-Zone Architecture**

```
┌─────────────────────────────────────────────────────┐
│                    Header Zone                      │
│  [Logo] [Search] [User] [Notifications] [Profile]  │
├─────────────┬─────────────────────────┬─────────────┤
│   Nav Zone  │      Content Zone       │Context Zone │
│             │                         │             │
│ [Dashboard] │    [Page Header]        │ [Filters]   │
│ [Sales]     │    [Metrics Cards]      │ [Activity]  │
│ [Orders]    │    [Data Tables]        │ [Settings]  │
│ [Contacts]  │    [Charts/Graphs]      │ [Help]      │
│ [Tasks]     │    [Action Buttons]     │             │
│ [Settings]  │                         │             │
│             │                         │             │
│             │                         │             │
└─────────────┴─────────────────────────┴─────────────┘
```

#### **Zone Responsibilities**

**Header Zone (Global)**
- Company branding and navigation
- Global search functionality
- User account and notifications
- System-wide actions

**Navigation Zone (Persistent)**
- Primary application navigation
- Contextual sub-navigation
- Breadcrumb trails
- Quick access shortcuts

**Content Zone (Dynamic)**
- Primary page content
- Data visualization
- Forms and interactions
- Main workflow elements

**Context Zone (Conditional)**
- Page-specific filters
- Activity feeds
- Help and documentation
- Secondary actions

### **Layout Implementation**

#### **CSS Grid Master Layout**
```css
.app-layout {
  display: grid;
  grid-template-areas: 
    "header header header"
    "nav content context";
  grid-template-columns: 16rem 1fr 20rem;
  grid-template-rows: 4rem 1fr;
  min-height: 100vh;
}

.app-header {
  grid-area: header;
  background: white;
  border-bottom: 1px solid #E5E7EB;
  display: flex;
  align-items: center;
  padding: 0 1.5rem;
  z-index: 40;
}

.app-nav {
  grid-area: nav;
  background: #F9FAFB;
  border-right: 1px solid #E5E7EB;
  overflow-y: auto;
  padding: 1.5rem 1rem;
}

.app-content {
  grid-area: content;
  padding: 2rem;
  overflow-y: auto;
  background: #F9FAFB;
}

.app-context {
  grid-area: context;
  background: white;
  border-left: 1px solid #E5E7EB;
  padding: 1.5rem;
  overflow-y: auto;
}
```

#### **Responsive Adaptations**
```css
/* Tablet: Hide context zone, expandable nav */
@media (max-width: 1024px) {
  .app-layout {
    grid-template-areas: 
      "header header"
      "nav content";
    grid-template-columns: 16rem 1fr;
  }
  
  .app-context {
    display: none;
  }
}

/* Mobile: Overlay nav, full-width content */
@media (max-width: 768px) {
  .app-layout {
    grid-template-areas: 
      "header"
      "content";
    grid-template-columns: 1fr;
  }
  
  .app-nav {
    position: fixed;
    left: -100%;
    width: 16rem;
    height: 100vh;
    z-index: 50;
    transition: left 0.3s ease;
  }
  
  .app-nav.open {
    left: 0;
  }
}
```

#### **Templ Master Layout Component**
```go
templ MasterLayout(props MasterLayoutProps) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>{ props.Title } - ERP System</title>
        <link href="/static/css/app.css" rel="stylesheet">
    </head>
    <body>
        <div class="app-layout">
            <!-- Header Zone -->
            <header class="app-header">
                @AppHeader(props.HeaderProps)
            </header>
            
            <!-- Navigation Zone -->
            <nav class="app-nav">
                @AppNavigation(props.NavProps)
            </nav>
            
            <!-- Content Zone -->
            <main class="app-content">
                { children... }
            </main>
            
            <!-- Context Zone -->
            if props.ShowContext {
                <aside class="app-context">
                    @AppContext(props.ContextProps)
                </aside>
            }
        </div>
        
        <script src="/static/js/app.js"></script>
    </body>
    </html>
}
```
<!-- LLM-SECTION-MASTER-LAYOUT-END -->

---

## 📊 Page Type Architectures

<!-- LLM-SECTION-PAGE-TYPES-START -->
### **1. Dashboard Page Architecture**

Based on `dashboard-main-overview.webp`

#### **Information Architecture**
```
Dashboard Page Structure:
├── Page Header
│   ├── Welcome Message ("Hi, Erin Aminoff")
│   ├── Context Description ("Here's your business summary...")
│   └── Quick Actions ([Customize] [Add new])
├── Metrics Section (Top Priority)
│   ├── Total Sales Card
│   ├── Total Product Card
│   └── Total Customer Card
├── Analytics Section (Charts & Graphs)
│   ├── Sales Performance Chart (Left 2/3)
│   └── Order Fulfillment Tracking (Right 1/3)
├── Data Tables Section
│   └── Top Selling Products Table
└── Activity Feed Section
    └── Recent Activity Stream
```

#### **Dashboard Layout Implementation**
```css
.dashboard-layout {
  display: grid;
  grid-template-areas:
    "header header header"
    "metrics metrics metrics"
    "charts charts activity"
    "tables tables activity";
  grid-template-columns: 1fr 1fr 20rem;
  gap: 2rem;
}

.dashboard-header {
  grid-area: header;
}

.dashboard-metrics {
  grid-area: metrics;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
}

.dashboard-charts {
  grid-area: charts;
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 1.5rem;
}

.dashboard-tables {
  grid-area: tables;
}

.dashboard-activity {
  grid-area: activity;
}
```

#### **Templ Dashboard Component**
```go
templ DashboardPage(props DashboardProps) {
    @MasterLayout(MasterLayoutProps{
        Title: "Dashboard",
        ShowContext: true,
    }) {
        <div class="dashboard-layout">
            <!-- Page Header -->
            <section class="dashboard-header">
                <h1 class="text-2xl font-bold text-gray-900 mb-2">
                    Hi, { props.UserName }
                </h1>
                <p class="text-gray-600 mb-4">
                    Here's your business summary for this month
                </p>
                <div class="flex space-x-3">
                    <button class="btn-secondary">
                        @CustomizeIcon()
                        Customize
                    </button>
                    <button class="btn-primary">
                        @PlusIcon()
                        Add new
                    </button>
                </div>
            </section>
            
            <!-- Metrics Cards -->
            <section class="dashboard-metrics">
                @MetricCard(props.SalesMetric)
                @MetricCard(props.ProductMetric)
                @MetricCard(props.CustomerMetric)
            </section>
            
            <!-- Charts Section -->
            <section class="dashboard-charts">
                <div class="bg-white rounded-lg shadow-sm border p-6">
                    <h3 class="text-lg font-semibold mb-4">Sales Performance</h3>
                    @SalesChart(props.SalesData)
                </div>
                <div class="bg-white rounded-lg shadow-sm border p-6">
                    <h3 class="text-lg font-semibold mb-4">Order & Fulfillment Tracking</h3>
                    @OrderChart(props.OrderData)
                </div>
            </section>
            
            <!-- Data Tables -->
            <section class="dashboard-tables">
                <div class="bg-white rounded-lg shadow-sm border">
                    <div class="p-6 border-b border-gray-200">
                        <h3 class="text-lg font-semibold">Top Selling</h3>
                    </div>
                    @ProductTable(props.ProductData)
                </div>
            </section>
        </div>
    }
}
```

### **2. List View Page Architecture**

Based on `contacts-management-list.webp`

#### **Information Architecture**
```
List View Page Structure:
├── Page Header
│   ├── Breadcrumb Navigation
│   ├── Page Title & Description
│   └── View Controls (Grid/List toggles)
├── Tab Navigation
│   ├── Primary Tabs (Opportunities, Accounts, Contacts, Leads)
│   └── Status Indicators (counts)
├── Filter & Search Bar
│   ├── Search Input
│   ├── Filter Dropdown
│   └── Status Filter
├── Content Grid/List
│   ├── Status-Based Grouping
│   │   ├── "New Leads" Section
│   │   ├── "Recently added" Section
│   │   └── "Signed in" Section
│   └── Item Cards/Rows
└── Pagination/Load More
```

#### **List Layout Implementation**
```css
.list-layout {
  display: grid;
  grid-template-areas:
    "header"
    "tabs"
    "filters"
    "content";
  grid-template-rows: auto auto auto 1fr;
  gap: 1.5rem;
}

.list-header {
  grid-area: header;
  display: flex;
  justify-content: between;
  align-items: center;
}

.list-tabs {
  grid-area: tabs;
}

.list-filters {
  grid-area: filters;
  display: flex;
  gap: 1rem;
  align-items: center;
}

.list-content {
  grid-area: content;
}

.contact-groups {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(20rem, 1fr));
  gap: 2rem;
}

.contact-group {
  background: white;
  border-radius: 0.5rem;
  border: 1px solid #E5E7EB;
  overflow: hidden;
}

.contact-group-header {
  padding: 1rem 1.5rem;
  background: #F9FAFB;
  border-bottom: 1px solid #E5E7EB;
  font-weight: 600;
}

.contact-cards {
  padding: 1rem;
  display: grid;
  gap: 0.75rem;
}
```

#### **Templ List Page Component**
```go
templ ContactListPage(props ContactListProps) {
    @MasterLayout(MasterLayoutProps{
        Title: "Contacts",
        ShowContext: true,
    }) {
        <div class="list-layout">
            <!-- Page Header -->
            <section class="list-header">
                <div>
                    @Breadcrumb(props.Breadcrumb)
                    <h1 class="text-2xl font-bold text-gray-900 mt-1">Contacts</h1>
                    <p class="text-gray-600">List of people or organizations for communication</p>
                </div>
                <div class="flex items-center space-x-2">
                    @ViewToggle(props.ViewMode)
                </div>
            </section>
            
            <!-- Tab Navigation -->
            <section class="list-tabs">
                @TabNavigation(props.Tabs)
            </section>
            
            <!-- Filters -->
            <section class="list-filters">
                <div class="flex-1">
                    @SearchInput(props.SearchProps)
                </div>
                @FilterDropdown(props.FilterProps)
                @StatusFilter(props.StatusProps)
            </section>
            
            <!-- Content Grid -->
            <section class="list-content">
                <div class="contact-groups">
                    for _, group := range props.ContactGroups {
                        <div class="contact-group">
                            <div class="contact-group-header">
                                { group.Title } 
                                <span class="text-sm font-normal text-gray-500">
                                    { fmt.Sprintf("%d", len(group.Contacts)) }
                                </span>
                            </div>
                            <div class="contact-cards">
                                for _, contact := range group.Contacts {
                                    @ContactCard(contact)
                                }
                            </div>
                        </div>
                    }
                </div>
            </section>
        </div>
    }
}
```

### **3. Task Management Page Architecture**

Based on `task-management-kanban.webp`

#### **Information Architecture**
```
Task Management Page Structure:
├── Page Header
│   ├── Breadcrumb (Team spaces > Tasks)
│   ├── Page Title & Description
│   └── Team Members (Avatars)
├── View Navigation
│   ├── Overview Tab
│   ├── Board Tab (Active)
│   ├── List Tab
│   ├── Table Tab
│   └── Timeline Tab
├── Kanban Board
│   ├── Column: "To do" (4 tasks)
│   ├── Column: "In Progress" (5 tasks)
│   └── Column: "Done" (3 tasks)
└── Task Cards
    ├── Status Indicators
    ├── Priority Labels
    ├── Assignment Info
    └── Metadata (comments, links, dates)
```

#### **Kanban Layout Implementation**
```css
.kanban-layout {
  display: grid;
  grid-template-areas:
    "header"
    "tabs"
    "board";
  grid-template-rows: auto auto 1fr;
  gap: 1.5rem;
  height: 100%;
}

.kanban-header {
  grid-area: header;
  display: flex;
  justify-content: between;
  align-items: center;
}

.kanban-tabs {
  grid-area: tabs;
}

.kanban-board {
  grid-area: board;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
  overflow-x: auto;
  padding-bottom: 1rem;
}

.kanban-column {
  background: #F3F4F6;
  border-radius: 0.75rem;
  padding: 1rem;
  min-height: 60vh;
}

.kanban-column-header {
  display: flex;
  justify-content: between;
  align-items: center;
  margin-bottom: 1rem;
  padding: 0.5rem;
}

.kanban-column-title {
  font-weight: 600;
  font-size: 0.875rem;
  color: #374151;
}

.kanban-column-count {
  background: #E5E7EB;
  color: #6B7280;
  border-radius: 9999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.task-list {
  display: grid;
  gap: 0.75rem;
}
```

#### **Templ Kanban Component**
```go
templ TaskKanbanPage(props TaskKanbanProps) {
    @MasterLayout(MasterLayoutProps{
        Title: "Tasks",
        ShowContext: false,
    }) {
        <div class="kanban-layout" x-data="kanbanBoard()">
            <!-- Page Header -->
            <section class="kanban-header">
                <div>
                    @Breadcrumb(props.Breadcrumb)
                    <h1 class="text-2xl font-bold text-gray-900 mt-1">Tasks</h1>
                    <p class="text-gray-600">Short description will be placed right over here</p>
                </div>
                <div class="flex items-center space-x-2">
                    @AvatarGroup(props.TeamMembers)
                </div>
            </section>
            
            <!-- View Tabs -->
            <section class="kanban-tabs">
                @ViewTabs(props.ViewTabs)
            </section>
            
            <!-- Kanban Board -->
            <section class="kanban-board">
                for _, column := range props.Columns {
                    <div class="kanban-column" data-column={ column.ID }>
                        <div class="kanban-column-header">
                            <h3 class="kanban-column-title">{ column.Title }</h3>
                            <span class="kanban-column-count">{ fmt.Sprintf("%d", len(column.Tasks)) }</span>
                        </div>
                        <div class="task-list" 
                             x-data="{ tasks: [] }"
                             x-droppable="column.id">
                            for _, task := range column.Tasks {
                                @TaskCard(task)
                            }
                        </div>
                        <button class="w-full mt-3 p-2 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-md transition-colors">
                            + Add task
                        </button>
                    </div>
                }
            </section>
        </div>
    }
}
```
<!-- LLM-SECTION-PAGE-TYPES-END -->

---

## 🎪 Modal & Overlay Architectures

<!-- LLM-SECTION-MODALS-START -->
### **Modal Layout Patterns**

Based on `notification-panel-modal.png` and `campaign-creation-form.png`

#### **1. Notification Panel Modal**

```
Notification Modal Structure:
├── Modal Header
│   ├── Title ("Notification")
│   └── Close Button
├── Tab Navigation
│   ├── "View All" (Active)
│   ├── "Mentions"
│   └── "Archive"
├── Notification List
│   ├── User Avatar
│   ├── Notification Text
│   ├── Timestamp
│   ├── Action Buttons (Approve/Deny)
│   └── File Attachments
└── Modal Footer (Optional)
```

#### **Modal CSS Architecture**
```css
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
  animation: fadeIn 0.2s ease-out;
}

.modal-container {
  background: white;
  border-radius: 0.75rem;
  max-width: 32rem;
  width: 90vw;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 20px 25px rgba(0, 0, 0, 0.1);
  animation: slideIn 0.3s ease-out;
}

.modal-header {
  padding: 1.5rem 1.5rem 0;
  border-bottom: 1px solid #E5E7EB;
  display: flex;
  justify-content: between;
  align-items: center;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  max-height: 60vh;
}

.modal-tabs {
  display: flex;
  border-bottom: 1px solid #E5E7EB;
  margin: 0 -1.5rem 1.5rem;
}

.modal-tab {
  flex: 1;
  padding: 0.75rem 1rem;
  text-align: center;
  font-size: 0.875rem;
  font-weight: 500;
  color: #6B7280;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.modal-tab.active {
  color: #3B82F6;
  border-bottom-color: #3B82F6;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideIn {
  from { 
    opacity: 0; 
    transform: scale(0.95) translateY(-10px); 
  }
  to { 
    opacity: 1; 
    transform: scale(1) translateY(0); 
  }
}
```

#### **Templ Modal Component**
```go
templ NotificationModal(props NotificationModalProps) {
    <div class="modal-overlay" 
         x-show="showNotifications" 
         x-transition:enter="transition ease-out duration-200"
         x-transition:enter-start="opacity-0"
         x-transition:enter-end="opacity-100"
         @click.self="showNotifications = false">
        
        <div class="modal-container"
             x-transition:enter="transition ease-out duration-300"
             x-transition:enter-start="opacity-0 scale-95 -translate-y-10"
             x-transition:enter-end="opacity-100 scale-100 translate-y-0">
            
            <!-- Modal Header -->
            <div class="modal-header">
                <h2 class="text-lg font-semibold text-gray-900">Notification</h2>
                <button @click="showNotifications = false" 
                        class="text-gray-400 hover:text-gray-600">
                    @CloseIcon()
                </button>
            </div>
            
            <!-- Modal Body -->
            <div class="modal-body">
                <!-- Tabs -->
                <div class="modal-tabs" x-data="{ activeTab: 'all' }">
                    <button class="modal-tab"
                            :class="{ 'active': activeTab === 'all' }"
                            @click="activeTab = 'all'">
                        View All
                    </button>
                    <button class="modal-tab"
                            :class="{ 'active': activeTab === 'mentions' }"
                            @click="activeTab = 'mentions'">
                        Mentions
                    </button>
                    <button class="modal-tab"
                            :class="{ 'active': activeTab === 'archive' }"
                            @click="activeTab = 'archive'">
                        Archive
                    </button>
                </div>
                
                <!-- Notification List -->
                <div class="space-y-4">
                    for _, notification := range props.Notifications {
                        @NotificationItem(notification)
                    }
                </div>
            </div>
        </div>
    </div>
}
```

#### **2. Form Modal Architecture**

Based on `campaign-creation-form.png`

```
Form Modal Structure:
├── Modal Header
│   └── Title ("Campaign Info")
├── Form Sections
│   ├── Basic Information
│   │   ├── Campaign Name Input
│   │   ├── Brand Input
│   │   └── Channel Selection
│   ├── Description
│   │   └── Textarea Input
│   ├── Audience Settings
│   │   ├── Target Customers
│   │   ├── Email/SMS Fields
│   │   └── Customer Dropdown
│   ├── Time Management
│   │   ├── Schedule Dropdown
│   │   └── Date Range Picker
│   └── Rules Configuration
│       ├── Spend Rules
│       └── Budget Rules
└── Modal Footer
    ├── Cancel Button
    └── Create Campaign Button
```

#### **Form Modal Implementation**
```go
templ CampaignFormModal(props CampaignFormProps) {
    <div class="modal-overlay" x-show="showCampaignForm">
        <div class="modal-container max-w-2xl">
            <!-- Header -->
            <div class="modal-header">
                <h2 class="text-lg font-semibold">Campaign Info</h2>
                <button @click="showCampaignForm = false">
                    @CloseIcon()
                </button>
            </div>
            
            <!-- Form Body -->
            <div class="modal-body">
                <form hx-post="/api/campaigns" 
                      hx-target="#campaign-list"
                      hx-swap="beforeend"
                      class="space-y-6">
                    
                    <!-- Basic Info Section -->
                    <section>
                        <h3 class="text-sm font-semibold text-gray-900 mb-4">Basic Information</h3>
                        <div class="grid grid-cols-2 gap-4">
                            <div>
                                <label class="form-label">Campaign name</label>
                                <input type="text" class="form-input" 
                                       name="name" 
                                       placeholder="Taste the Future: Frozen Delights Delivered">
                            </div>
                            <div>
                                <label class="form-label">Brand</label>
                                <input type="text" class="form-input" 
                                       name="brand" 
                                       placeholder="Damory Food Indonesia">
                            </div>
                        </div>
                        <div class="mt-4">
                            <label class="form-label">Channel</label>
                            @ChannelSelector(props.Channels)
                        </div>
                    </section>
                    
                    <!-- Description Section -->
                    <section>
                        <h3 class="text-sm font-semibold text-gray-900 mb-4">Description</h3>
                        <textarea class="form-input h-24" 
                                  name="description"
                                  placeholder="A social media and Google Ads campaign..."></textarea>
                    </section>
                    
                    <!-- Audience Section -->
                    <section>
                        <h3 class="text-sm font-semibold text-gray-900 mb-4">Audience</h3>
                        <div class="grid grid-cols-2 gap-4">
                            <div>
                                <label class="form-label">Target customers</label>
                                <input type="number" class="form-input" 
                                       name="target_customers" 
                                       value="10,000">
                            </div>
                            <div>
                                <label class="form-label">Email only</label>
                                <input type="number" class="form-input" 
                                       name="email_only" 
                                       value="3,890">
                            </div>
                        </div>
                        <!-- More audience fields... -->
                    </section>
                    
                    <!-- Additional sections... -->
                </form>
            </div>
            
            <!-- Footer -->
            <div class="modal-footer">
                <button type="button" class="btn-secondary" 
                        @click="showCampaignForm = false">
                    Cancel
                </button>
                <button type="submit" class="btn-primary">
                    Create Campaign
                </button>
            </div>
        </div>
    </div>
}
```
<!-- LLM-SECTION-MODALS-END -->

---

## 📱 Responsive Architecture Patterns

<!-- LLM-SECTION-RESPONSIVE-ARCH-START -->
### **Breakpoint Strategy**

#### **Responsive Layout Transformations**

**Desktop (1024px+)**
```
┌─────────────────────────────────────────────────────┐
│                    Header                           │
├─────────────┬─────────────────────────┬─────────────┤
│     Nav     │        Content          │   Context   │
│   (256px)   │       (flexible)        │   (320px)   │
│             │                         │             │
└─────────────┴─────────────────────────┴─────────────┘
```

**Tablet (768px - 1023px)**
```
┌─────────────────────────────────────────────────────┐
│                    Header                           │
├─────────────┬─────────────────────────────────────────┤
│     Nav     │              Content                  │
│   (200px)   │            (flexible)                 │
│             │                                       │
└─────────────┴─────────────────────────────────────────┘
```

**Mobile (0 - 767px)**
```
┌─────────────────────────────────────────────────────┐
│                 Header + Menu                       │
├─────────────────────────────────────────────────────┤
│                                                     │
│                   Content                           │
│                 (full width)                        │
│                                                     │
└─────────────────────────────────────────────────────┘

Nav becomes overlay/drawer
Context becomes bottom sheet or hidden
```

#### **Responsive Implementation**
```css
/* Mobile-first responsive grid */
.responsive-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr;
}

/* Tablet and up */
@media (min-width: 768px) {
  .responsive-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
  }
}

/* Desktop and up */
@media (min-width: 1024px) {
  .responsive-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 2rem;
  }
}

/* Navigation responsive behavior */
.app-nav {
  transform: translateX(-100%);
  transition: transform 0.3s ease;
}

@media (min-width: 1024px) {
  .app-nav {
    transform: translateX(0);
    position: static;
  }
}

.app-nav.mobile-open {
  transform: translateX(0);
}
```

### **Content Adaptation Strategies**

#### **Dashboard Responsive Patterns**
```css
/* Metric cards responsive stack */
.metrics-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr;
}

@media (min-width: 640px) {
  .metrics-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .metrics-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

/* Chart responsive behavior */
.chart-container {
  position: relative;
  height: 20rem;
}

@media (max-width: 767px) {
  .chart-container {
    height: 16rem;
  }
}

/* Table to card transformation */
.data-table {
  display: table;
}

@media (max-width: 767px) {
  .data-table {
    display: block;
  }
  
  .table-row {
    display: block;
    background: white;
    border-radius: 0.5rem;
    padding: 1rem;
    margin-bottom: 0.75rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }
  
  .table-cell {
    display: block;
    padding: 0.25rem 0;
  }
  
  .table-cell:before {
    content: attr(data-label) ": ";
    font-weight: 600;
    color: #6B7280;
  }
}
```

#### **Alpine.js Responsive State Management**
```javascript
Alpine.data('responsiveLayout', () => ({
  isMobile: window.innerWidth < 768,
  isTablet: window.innerWidth >= 768 && window.innerWidth < 1024,
  isDesktop: window.innerWidth >= 1024,
  
  mobileNavOpen: false,
  contextPanelOpen: false,
  
  init() {
    this.updateBreakpoints();
    window.addEventListener('resize', () => this.updateBreakpoints());
  },
  
  updateBreakpoints() {
    this.isMobile = window.innerWidth < 768;
    this.isTablet = window.innerWidth >= 768 && window.innerWidth < 1024;
    this.isDesktop = window.innerWidth >= 1024;
    
    // Auto-close mobile nav on desktop
    if (this.isDesktop) {
      this.mobileNavOpen = false;
    }
  },
  
  toggleMobileNav() {
    this.mobileNavOpen = !this.mobileNavOpen;
  },
  
  showContext() {
    if (this.isMobile) {
      this.contextPanelOpen = true;
    }
  }
}));
```
<!-- LLM-SECTION-RESPONSIVE-ARCH-END -->

---

## 🔧 Implementation Best Practices

<!-- LLM-SECTION-BEST-PRACTICES-START -->
### **Performance Optimization**

#### **Lazy Loading Patterns**
```html
<!-- Lazy load heavy components -->
<div hx-get="/api/dashboard/charts" 
     hx-trigger="intersect once"
     hx-target="#charts-container"
     hx-indicator="#charts-loading">
  <div id="charts-loading" class="skeleton-chart"></div>
  <div id="charts-container"></div>
</div>

<!-- Progressive enhancement -->
<div class="metric-card" 
     hx-get="/api/metrics/details" 
     hx-trigger="click"
     hx-target="#metric-modal"
     hx-swap="innerHTML">
  <!-- Static content works without JS -->
  <h3>Total Sales</h3>
  <p>$275,825</p>
</div>
```

#### **State Management Architecture**
```javascript
// Global application state
Alpine.store('app', {
  user: null,
  notifications: [],
  sidebarOpen: false,
  
  // Layout state
  layout: {
    navCollapsed: false,
    contextVisible: true,
    theme: 'light'
  },
  
  // Data cache
  cache: new Map(),
  
  // Methods
  toggleSidebar() {
    this.layout.navCollapsed = !this.layout.navCollapsed;
  },
  
  addNotification(notification) {
    this.notifications.unshift(notification);
    if (this.notifications.length > 10) {
      this.notifications = this.notifications.slice(0, 10);
    }
  }
});
```

### **Accessibility Architecture**

#### **Semantic HTML Structure**
```html
<!-- Proper landmark structure -->
<header role="banner" class="app-header">
  <nav role="navigation" aria-label="Main navigation">
    <!-- Navigation items -->
  </nav>
</header>

<main role="main" class="app-content">
  <section aria-labelledby="dashboard-heading">
    <h1 id="dashboard-heading">Dashboard</h1>
    <!-- Dashboard content -->
  </section>
</main>

<aside role="complementary" class="app-context">
  <!-- Context content -->
</aside>
```

#### **Screen Reader Optimization**
```html
<!-- Skip links -->
<a href="#main-content" class="sr-only focus:not-sr-only">
  Skip to main content
</a>

<!-- Descriptive headings -->
<h2 id="metrics-heading" class="sr-only">
  Business Metrics Overview
</h2>

<!-- ARIA labels for data -->
<div class="metric-card" 
     aria-labelledby="sales-label" 
     aria-describedby="sales-trend">
  <h3 id="sales-label">Total Sales</h3>
  <p class="metric-value">$275,825</p>
  <p id="sales-trend" class="sr-only">
    Sales increased by 28% compared to last month
  </p>
</div>
```

### **Error Handling Architecture**

#### **Graceful Degradation Patterns**
```html
<!-- Fallback content for failed requests -->
<div hx-get="/api/dashboard" 
     hx-target="#dashboard-content"
     hx-on:htmx:error="handleError">
  
  <div id="dashboard-content">
    <!-- Static fallback content -->
    <div class="error-state">
      <h2>Dashboard Temporarily Unavailable</h2>
      <p>Please refresh the page or try again later.</p>
      <button onclick="location.reload()">Refresh Page</button>
    </div>
  </div>
</div>

<!-- Error state components -->
<div class="error-boundary" 
     x-data="errorHandler()"
     x-on:error.window="handleError($event)">
  <!-- Content that might error -->
</div>
```

#### **Loading State Management**
```css
/* Loading skeletons */
.skeleton {
  background: linear-gradient(90deg, 
    #f0f0f0 25%, 
    #e0e0e0 50%, 
    #f0f0f0 75%);
  background-size: 200% 100%;
  animation: loading 1.5s infinite;
  border-radius: 0.375rem;
}

.skeleton-text {
  height: 1rem;
  margin-bottom: 0.5rem;
}

.skeleton-text:last-child {
  width: 75%;
}

.skeleton-card {
  height: 8rem;
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```
<!-- LLM-SECTION-BEST-PRACTICES-END -->

---

## 🎯 Architecture Quick Reference

### **Layout Templates**
- **Dashboard**: Metrics → Charts → Tables + Activity feed
- **List View**: Header → Tabs → Filters → Grouped cards/table
- **Detail View**: Header → Tabs → Form sections → Actions
- **Kanban**: Header → View tabs → Column layout → Draggable cards

### **Component Hierarchy**
1. **Page Level**: Full page layouts with navigation
2. **Section Level**: Content areas within pages
3. **Component Level**: Reusable UI components
4. **Element Level**: Basic HTML elements with styling

### **Responsive Strategy**
- **Mobile**: Single column, overlay navigation, simplified interactions
- **Tablet**: Two-column grids, collapsed navigation, touch-optimized
- **Desktop**: Full three-zone layout, hover interactions, dense information

### **State Management**
- **Alpine.js**: Component-level reactive state
- **HTMX**: Server-driven state and navigation
- **Local Storage**: User preferences and layout settings
- **URL State**: Current page and filter states

This architecture guide provides the **foundational patterns** for building **consistent, scalable, and maintainable** ERP interface layouts that serve **business productivity** while maintaining **excellent user experience** across all device types.

---

*Use these patterns as **starting templates** and adapt them based on specific business requirements while maintaining the core architectural principles for **optimal user experience** and **development efficiency**.*