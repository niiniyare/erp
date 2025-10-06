# ERP UI Implementation Guide - Complete UX Developer Handbook

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Comprehensive implementation guide combining styles, architecture, and practical development patterns
**SCOPE**: End-to-end development workflow, best practices, and production-ready code examples
**TARGET AUDIENCE**: UX developers implementing the ERP design system
<!-- LLM-CONTEXT-END -->

## 🎯 Implementation Overview

This guide provides a **complete development workflow** for implementing the ERP UI design system using **Templ + HTMX + Alpine.js + Flowbite**, combining the visual styles and page architectures into **production-ready implementations**.

### **Development Stack Integration**
- **Templ**: Type-safe Go templating for component structure
- **HTMX**: Server-driven interactions and dynamic content
- **Alpine.js**: Client-side reactivity and state management  
- **Flowbite**: Pre-built UI components and styling framework
- **TailwindCSS**: Utility-first styling system

---

## 🚀 Quick Start Implementation

<!-- LLM-SECTION-QUICKSTART-START -->
### **1. Project Setup**

#### **Directory Structure**
```
project/
├── internal/
│   └── ui/
│       ├── components/          # Reusable Templ components
│       │   ├── layout/          # Layout components
│       │   ├── cards/           # Card components
│       │   ├── forms/           # Form components
│       │   └── modals/          # Modal components
│       ├── pages/               # Full page templates
│       │   ├── dashboard.templ
│       │   ├── contacts.templ
│       │   └── tasks.templ
│       └── assets/              # Static assets
│           ├── css/
│           ├── js/
│           └── images/
├── static/                      # Compiled assets
└── cmd/
    └── server/
        └── main.go
```

#### **Base Template Setup**
```go
// internal/ui/components/layout/base.templ
package layout

templ Base(title string, showContext bool) {
    <!DOCTYPE html>
    <html lang="en" class="h-full bg-gray-50">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>{ title } - ERP System</title>
        
        <!-- TailwindCSS + Flowbite -->
        <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
        <link href="https://cdnjs.cloudflare.com/ajax/libs/flowbite/1.8.1/flowbite.min.css" rel="stylesheet">
        
        <!-- Custom styles -->
        <link href="/static/css/app.css" rel="stylesheet">
        
        <!-- HTMX -->
        <script src="https://unpkg.com/htmx.org@1.9.6"></script>
        
        <!-- Alpine.js -->
        <script defer src="https://unpkg.com/alpinejs@3.13.0/dist/cdn.min.js"></script>
    </head>
    <body class="h-full" x-data="appState()">
        <div class="app-layout">
            @Header()
            @Navigation()
            <main class="app-content">
                { children... }
            </main>
            if showContext {
                @ContextPanel()
            }
        </div>
        
        <!-- Flowbite JS -->
        <script src="https://cdnjs.cloudflare.com/ajax/libs/flowbite/1.8.1/flowbite.min.js"></script>
        <script src="/static/js/app.js"></script>
    </body>
    </html>
}
```

### **2. Core CSS Setup**

#### **app.css - Custom Styles**
```css
/* Import Tailwind base */
@import 'tailwindcss/base';
@import 'tailwindcss/components';
@import 'tailwindcss/utilities';

/* Custom CSS variables */
:root {
  --color-primary: #10B981;
  --color-primary-50: #ECFDF5;
  --color-primary-600: #059669;
  --color-primary-700: #047857;
  
  --shadow-card: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
  --shadow-elevated: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

/* Layout Grid */
.app-layout {
  display: grid;
  grid-template-areas: 
    "header header header"
    "nav content context";
  grid-template-columns: 16rem 1fr 20rem;
  grid-template-rows: 4rem 1fr;
  min-height: 100vh;
  background: #F9FAFB;
}

.app-header {
  grid-area: header;
  background: white;
  border-bottom: 1px solid #E5E7EB;
  box-shadow: var(--shadow-card);
  z-index: 40;
}

.app-nav {
  grid-area: nav;
  background: white;
  border-right: 1px solid #E5E7EB;
  overflow-y: auto;
}

.app-content {
  grid-area: content;
  padding: 2rem;
  overflow-y: auto;
}

.app-context {
  grid-area: context;
  background: white;
  border-left: 1px solid #E5E7EB;
  padding: 1.5rem;
  overflow-y: auto;
}

/* Component base styles */
.card {
  background: white;
  border-radius: 0.5rem;
  box-shadow: var(--shadow-card);
  border: 1px solid #E5E7EB;
  padding: 1.5rem;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn-primary:hover {
  background: var(--color-primary-600);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(16, 185, 129, 0.25);
}

/* Responsive adjustments */
@media (max-width: 1024px) {
  .app-layout {
    grid-template-areas: 
      "header header"
      "nav content";
    grid-template-columns: 16rem 1fr;
  }
  .app-context { display: none; }
}

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
  
  .app-nav.open { left: 0; }
}
```

### **3. JavaScript Setup**

#### **app.js - Alpine.js Configuration**
```javascript
// Global application state
document.addEventListener('alpine:init', () => {
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
});

// Main app state component
function appState() {
    return {
        mobileNavOpen: false,
        
        init() {
            // Initialize app
            this.loadUserData();
        },
        
        async loadUserData() {
            // Load user data from server
            try {
                const response = await fetch('/api/user');
                const user = await response.json();
                Alpine.store('app').user = user;
            } catch (error) {
                console.error('Failed to load user data:', error);
            }
        },
        
        toggleMobileNav() {
            this.mobileNavOpen = !this.mobileNavOpen;
        }
    };
}

// HTMX configuration
htmx.config.globalViewTransitions = true;
htmx.config.refreshOnHistoryMiss = true;

// Global error handler
document.body.addEventListener('htmx:error', function(evt) {
    console.error('HTMX Error:', evt.detail);
    // Show user-friendly error message
    Alpine.store('app').addNotification({
        type: 'error',
        message: 'Something went wrong. Please try again.',
        timestamp: new Date()
    });
});
```
<!-- LLM-SECTION-QUICKSTART-END -->

---

## 🧩 Component Implementation Patterns

<!-- LLM-SECTION-COMPONENTS-START -->
### **1. Metric Card Component**

#### **Complete Implementation**
```go
// internal/ui/components/cards/metric.templ
package cards

import "fmt"

type MetricCardProps struct {
    ID          string
    Label       string
    Value       string
    Trend       string // "up", "down", "neutral"
    TrendValue  string
    TrendLabel  string
    Icon        string
    IconColor   string
    HxGet       string // HTMX endpoint for detail view
}

templ MetricCard(props MetricCardProps) {
    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-all duration-200 cursor-pointer"
         if props.HxGet != "" {
             hx-get={ props.HxGet }
             hx-target="#modal-container"
             hx-swap="innerHTML"
         }>
        
        <div class="flex items-center justify-between">
            <!-- Icon and Content -->
            <div class="flex items-center">
                <div class={ fmt.Sprintf("flex-shrink-0 w-12 h-12 rounded-lg flex items-center justify-center %s", props.IconColor) }>
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
                        @TrendIcon(props.Trend)
                        <span class="text-xs font-medium ml-1">
                            { props.TrendValue } { props.TrendLabel }
                        </span>
                    </div>
                </div>
            </div>
            
            <!-- Action Menu -->
            <div class="relative" x-data="{ open: false }">
                <button @click="open = !open" 
                        class="text-gray-400 hover:text-gray-600 transition-colors">
                    @DotsVerticalIcon()
                </button>
                <div x-show="open" 
                     @click.away="open = false"
                     x-transition
                     class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg border border-gray-200 z-10">
                    <div class="py-1">
                        <a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">View Details</a>
                        <a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">Export Data</a>
                        <a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">Configure</a>
                    </div>
                </div>
            </div>
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

templ TrendIcon(trend string) {
    switch trend {
    case "up":
        <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M3.293 9.707a1 1 0 010-1.414l6-6a1 1 0 011.414 0l6 6a1 1 0 01-1.414 1.414L10 4.414 4.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd"></path>
        </svg>
    case "down":
        <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M16.707 10.293a1 1 0 010 1.414l-6 6a1 1 0 01-1.414 0l-6-6a1 1 0 111.414-1.414L10 15.586l5.293-5.293a1 1 0 011.414 0z" clip-rule="evenodd"></path>
        </svg>
    default:
        <div class="w-3 h-3 bg-gray-400 rounded-full"></div>
    }
}
```

### **2. Contact Card Component**

#### **Interactive Contact Card**
```go
// internal/ui/components/cards/contact.templ
package cards

type ContactCardProps struct {
    ID          string
    Name        string
    Email       string
    Phone       string
    Company     string
    Avatar      string
    Status      string
    Date        string
    Tags        []string
    HxClick     string
}

templ ContactCard(props ContactCardProps) {
    <div class="bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 hover:shadow-sm transition-all duration-200 cursor-pointer group"
         if props.HxClick != "" {
             hx-get={ props.HxClick }
             hx-target="#contact-detail"
             hx-swap="innerHTML"
             hx-push-url="true"
         }
         x-data="{ hover: false }"
         @mouseenter="hover = true"
         @mouseleave="hover = false">
        
        <div class="flex items-start space-x-3">
            <!-- Avatar with status indicator -->
            <div class="relative flex-shrink-0">
                <img class="w-10 h-10 rounded-full object-cover border-2 border-gray-100" 
                     src={ props.Avatar } 
                     alt={ props.Name }
                     loading="lazy">
                <div class={ getStatusDotClasses(props.Status) }></div>
            </div>
            
            <!-- Content -->
            <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between mb-1">
                    <h4 class="text-sm font-semibold text-gray-900 truncate group-hover:text-primary-600 transition-colors">
                        { props.Name }
                    </h4>
                    <span class="text-xs text-gray-500 flex-shrink-0">
                        { props.Date }
                    </span>
                </div>
                
                <!-- Company -->
                <p class="text-xs text-gray-600 mb-2 truncate">
                    { props.Company }
                </p>
                
                <!-- Contact Info -->
                <div class="space-y-1">
                    <div class="flex items-center text-xs text-gray-500">
                        @PhoneIcon("w-3 h-3")
                        <span class="ml-1 truncate">{ props.Phone }</span>
                    </div>
                    <div class="flex items-center text-xs text-gray-500">
                        @EmailIcon("w-3 h-3")
                        <span class="ml-1 truncate">{ props.Email }</span>
                    </div>
                </div>
                
                <!-- Tags -->
                if len(props.Tags) > 0 {
                    <div class="flex flex-wrap gap-1 mt-2">
                        for _, tag := range props.Tags {
                            <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                                { tag }
                            </span>
                        }
                    </div>
                }
            </div>
            
            <!-- Quick Actions (appear on hover) -->
            <div class="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button class="p-1 text-gray-400 hover:text-primary-600 rounded"
                        hx-post={ fmt.Sprintf("/api/contacts/%s/favorite", props.ID) }
                        hx-swap="none">
                    @HeartIcon("w-4 h-4")
                </button>
                <button class="p-1 text-gray-400 hover:text-primary-600 rounded"
                        onclick={ fmt.Sprintf("window.open('mailto:%s')", props.Email) }>
                    @EmailIcon("w-4 h-4")
                </button>
            </div>
        </div>
    </div>
}

func getStatusDotClasses(status string) string {
    base := "absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-white"
    switch status {
    case "online":
        return base + " bg-green-400"
    case "away":
        return base + " bg-yellow-400"
    case "offline":
        return base + " bg-gray-400"
    default:
        return base + " bg-gray-300"
    }
}
```

### **3. Data Table Component**

#### **Responsive Data Table**
```go
// internal/ui/components/tables/data_table.templ
package tables

type Column struct {
    Key       string
    Label     string
    Sortable  bool
    Width     string
    Align     string
}

type DataTableProps struct {
    ID        string
    Columns   []Column
    Data      []map[string]interface{}
    HxGet     string // For pagination/sorting
    Searchable bool
    Filterable bool
}

templ DataTable(props DataTableProps) {
    <div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden"
         x-data="dataTable()"
         x-init="init({ hxGet: '{ props.HxGet }' })">
        
        <!-- Table Header -->
        if props.Searchable || props.Filterable {
            <div class="p-4 border-b border-gray-200">
                <div class="flex items-center justify-between">
                    if props.Searchable {
                        <div class="relative">
                            <input type="text" 
                                   x-model="searchTerm"
                                   @input.debounced.300ms="search()"
                                   placeholder="Search..."
                                   class="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500">
                            <div class="absolute inset-y-0 left-0 pl-3 flex items-center">
                                @SearchIcon("w-5 h-5 text-gray-400")
                            </div>
                        </div>
                    }
                    
                    if props.Filterable {
                        <div class="flex items-center space-x-2">
                            <button class="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                                @FilterIcon("w-4 h-4 mr-2")
                                Filter
                            </button>
                        </div>
                    }
                </div>
            </div>
        }
        
        <!-- Table -->
        <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                    <tr>
                        for _, col := range props.Columns {
                            <th scope="col" 
                                class={ getHeaderClasses(col) }
                                if col.Sortable {
                                    @click={ fmt.Sprintf("sort('%s')", col.Key) }
                                    :class="{ 'bg-gray-100': sortField === '{ col.Key }' }"
                                }>
                                <div class="flex items-center space-x-1">
                                    <span>{ col.Label }</span>
                                    if col.Sortable {
                                        <div class="flex flex-col">
                                            @SortIcon()
                                        </div>
                                    }
                                </div>
                            </th>
                        }
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                    <template x-for="(row, index) in filteredData" :key="index">
                        <tr class="hover:bg-gray-50 transition-colors cursor-pointer"
                            @click="selectRow(row)">
                            for _, col := range props.Columns {
                                <td class={ getCellClasses(col) }>
                                    <span x-text={ fmt.Sprintf("row.%s || '-'", col.Key) }></span>
                                </td>
                            }
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>
        
        <!-- Loading State -->
        <div x-show="loading" class="p-8 text-center">
            <div class="inline-flex items-center space-x-2">
                @SpinnerIcon("w-5 h-5 animate-spin text-primary-600")
                <span class="text-gray-600">Loading...</span>
            </div>
        </div>
        
        <!-- Empty State -->
        <div x-show="!loading && filteredData.length === 0" class="p-8 text-center">
            <div class="text-gray-500">
                <p class="text-lg font-medium">No results found</p>
                <p class="text-sm">Try adjusting your search or filter criteria</p>
            </div>
        </div>
    </div>
}

func getHeaderClasses(col Column) string {
    base := "px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
    if col.Sortable {
        base += " cursor-pointer hover:bg-gray-100 transition-colors"
    }
    if col.Align == "center" {
        base += " text-center"
    } else if col.Align == "right" {
        base += " text-right"
    }
    return base
}

func getCellClasses(col Column) string {
    base := "px-6 py-4 whitespace-nowrap text-sm text-gray-900"
    if col.Align == "center" {
        base += " text-center"
    } else if col.Align == "right" {
        base += " text-right"
    }
    return base
}
```

#### **Alpine.js Table Component**
```javascript
// Add to app.js
function dataTable() {
    return {
        data: [],
        filteredData: [],
        loading: false,
        searchTerm: '',
        sortField: null,
        sortDirection: 'asc',
        
        async init(config) {
            this.config = config;
            await this.loadData();
        },
        
        async loadData() {
            this.loading = true;
            try {
                if (this.config.hxGet) {
                    const response = await fetch(this.config.hxGet);
                    this.data = await response.json();
                } else {
                    // Use static data passed from server
                    this.data = this.initialData || [];
                }
                this.filterAndSort();
            } catch (error) {
                console.error('Failed to load table data:', error);
            } finally {
                this.loading = false;
            }
        },
        
        search() {
            this.filterAndSort();
        },
        
        sort(field) {
            if (this.sortField === field) {
                this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortField = field;
                this.sortDirection = 'asc';
            }
            this.filterAndSort();
        },
        
        filterAndSort() {
            let filtered = [...this.data];
            
            // Apply search filter
            if (this.searchTerm) {
                const term = this.searchTerm.toLowerCase();
                filtered = filtered.filter(row =>
                    Object.values(row).some(value =>
                        String(value).toLowerCase().includes(term)
                    )
                );
            }
            
            // Apply sorting
            if (this.sortField) {
                filtered.sort((a, b) => {
                    const aVal = a[this.sortField];
                    const bVal = b[this.sortField];
                    const modifier = this.sortDirection === 'asc' ? 1 : -1;
                    
                    if (typeof aVal === 'number' && typeof bVal === 'number') {
                        return (aVal - bVal) * modifier;
                    }
                    
                    return String(aVal).localeCompare(String(bVal)) * modifier;
                });
            }
            
            this.filteredData = filtered;
        },
        
        selectRow(row) {
            // Handle row selection
            this.$dispatch('row-selected', { row });
        }
    };
}
```
<!-- LLM-SECTION-COMPONENTS-END -->

---

## 📱 Page Implementation Examples

<!-- LLM-SECTION-PAGES-START -->
### **1. Dashboard Page Implementation**

#### **Complete Dashboard Page**
```go
// internal/ui/pages/dashboard.templ
package pages

import (
    "project/internal/ui/components/layout"
    "project/internal/ui/components/cards"
    "project/internal/ui/components/charts"
)

type DashboardPageProps struct {
    User        User
    Metrics     []cards.MetricCardProps
    SalesData   charts.ChartData
    OrderData   charts.ChartData
    TopProducts []ProductData
}

templ DashboardPage(props DashboardPageProps) {
    @layout.Base("Dashboard", true) {
        <div class="dashboard-container" 
             x-data="dashboardState()"
             x-init="init()"
             hx-ext="sse"
             sse-connect="/api/dashboard/stream">
            
            <!-- Page Header -->
            <section class="mb-8">
                <div class="flex items-center justify-between">
                    <div>
                        <h1 class="text-2xl font-bold text-gray-900">
                            Hi, { props.User.Name }
                        </h1>
                        <p class="text-gray-600 mt-1">
                            Here's your business summary for this month
                        </p>
                    </div>
                    <div class="flex items-center space-x-3">
                        <button class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors">
                            @CustomizeIcon("w-4 h-4 mr-2")
                            Customize
                        </button>
                        <button class="inline-flex items-center px-4 py-2 bg-primary-600 text-white rounded-md text-sm font-medium hover:bg-primary-700 transition-colors"
                                hx-get="/api/dashboard/new-item-modal"
                                hx-target="#modal-container">
                            @PlusIcon("w-4 h-4 mr-2")
                            Add new
                        </button>
                    </div>
                </div>
            </section>
            
            <!-- Metrics Grid -->
            <section class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8"
                     hx-trigger="sse:metrics-update"
                     hx-swap="outerHTML">
                for _, metric := range props.Metrics {
                    @cards.MetricCard(metric)
                }
            </section>
            
            <!-- Charts Section -->
            <section class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
                <!-- Sales Performance Chart -->
                <div class="lg:col-span-2">
                    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                        <div class="flex items-center justify-between mb-6">
                            <h3 class="text-lg font-semibold text-gray-900">Sales Performance</h3>
                            <div class="flex items-center space-x-2">
                                <select class="text-sm border-gray-300 rounded-md">
                                    <option>Yearly</option>
                                    <option selected>Monthly</option>
                                    <option>Weekly</option>
                                </select>
                            </div>
                        </div>
                        @charts.LineChart(props.SalesData)
                    </div>
                </div>
                
                <!-- Order Fulfillment Tracking -->
                <div class="lg:col-span-1">
                    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                        <h3 class="text-lg font-semibold text-gray-900 mb-6">Order & Fulfillment Tracking</h3>
                        @charts.BarChart(props.OrderData)
                    </div>
                </div>
            </section>
            
            <!-- Data Tables Section -->
            <section class="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <!-- Top Selling Products -->
                <div class="lg:col-span-2">
                    <div class="bg-white rounded-lg shadow-sm border border-gray-200">
                        <div class="p-6 border-b border-gray-200 flex items-center justify-between">
                            <h3 class="text-lg font-semibold text-gray-900">Top Selling</h3>
                            <button class="text-sm text-primary-600 hover:text-primary-700 font-medium">
                                View all
                            </button>
                        </div>
                        @ProductTable(props.TopProducts)
                    </div>
                </div>
                
                <!-- Activity Feed -->
                <div class="lg:col-span-1">
                    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6"
                         hx-get="/api/dashboard/activity"
                         hx-trigger="load, sse:activity-update"
                         hx-swap="innerHTML">
                        <!-- Activity feed content will be loaded here -->
                    </div>
                </div>
            </section>
        </div>
    }
}

// Alpine.js dashboard state
templ DashboardScript() {
    <script>
    function dashboardState() {
        return {
            refreshInterval: null,
            
            init() {
                // Set up auto-refresh for non-SSE browsers
                if (!window.EventSource) {
                    this.refreshInterval = setInterval(() => {
                        this.refreshMetrics();
                    }, 30000); // 30 seconds
                }
            },
            
            async refreshMetrics() {
                try {
                    const response = await fetch('/api/dashboard/metrics');
                    const data = await response.json();
                    // Update metrics without full page reload
                    this.$dispatch('metrics-updated', data);
                } catch (error) {
                    console.error('Failed to refresh metrics:', error);
                }
            },
            
            destroy() {
                if (this.refreshInterval) {
                    clearInterval(this.refreshInterval);
                }
            }
        };
    }
    </script>
}
```

### **2. Contact List Page Implementation**

#### **Complete Contact Management Page**
```go
// internal/ui/pages/contacts.templ
package pages

type ContactListPageProps struct {
    Breadcrumb      []BreadcrumbItem
    Tabs           []TabItem
    ContactGroups  []ContactGroup
    Filters        FilterProps
    ViewMode       string
}

type ContactGroup struct {
    Title     string
    Count     int
    Contacts  []cards.ContactCardProps
}

templ ContactListPage(props ContactListPageProps) {
    @layout.Base("Contacts", true) {
        <div class="contacts-container" 
             x-data="contactsState()"
             x-init="init()">
            
            <!-- Page Header -->
            <section class="mb-6">
                <div class="flex items-center justify-between">
                    <div>
                        @Breadcrumb(props.Breadcrumb)
                        <h1 class="text-2xl font-bold text-gray-900 mt-2">Contacts</h1>
                        <p class="text-gray-600">List of people or organizations for communication</p>
                    </div>
                    <div class="flex items-center space-x-2">
                        <!-- View Toggle -->
                        <div class="flex bg-gray-100 rounded-lg p-1">
                            <button class="flex items-center px-3 py-1.5 text-sm font-medium rounded-md transition-colors"
                                    :class="viewMode === 'grid' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'"
                                    @click="setViewMode('grid')">
                                @GridIcon("w-4 h-4")
                                <span class="ml-1 hidden sm:inline">Grid</span>
                            </button>
                            <button class="flex items-center px-3 py-1.5 text-sm font-medium rounded-md transition-colors"
                                    :class="viewMode === 'list' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'"
                                    @click="setViewMode('list')">
                                @ListIcon("w-4 h-4")
                                <span class="ml-1 hidden sm:inline">List</span>
                            </button>
                        </div>
                    </div>
                </div>
            </section>
            
            <!-- Tab Navigation -->
            <section class="mb-6">
                <div class="border-b border-gray-200">
                    <nav class="-mb-px flex space-x-8">
                        for _, tab := range props.Tabs {
                            <button class="py-2 px-1 border-b-2 font-medium text-sm transition-colors"
                                    :class="activeTab === '{ tab.ID }' ? 'border-primary-500 text-primary-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
                                    @click="setActiveTab('{ tab.ID }')"
                                    hx-get={ fmt.Sprintf("/api/contacts?tab=%s", tab.ID) }
                                    hx-target="#contacts-content"
                                    hx-swap="innerHTML">
                                { tab.Label }
                                if tab.Count > 0 {
                                    <span class="ml-2 bg-gray-100 text-gray-900 py-0.5 px-2.5 rounded-full text-xs font-medium">
                                        { fmt.Sprintf("%d", tab.Count) }
                                    </span>
                                }
                            </button>
                        }
                    </nav>
                </div>
            </section>
            
            <!-- Filters and Search -->
            <section class="mb-6">
                <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-3 sm:space-y-0">
                    <!-- Search -->
                    <div class="relative flex-1 max-w-md">
                        <input type="text" 
                               x-model="searchTerm"
                               @input.debounced.300ms="search()"
                               placeholder="Search contacts..."
                               class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500">
                        <div class="absolute inset-y-0 left-0 pl-3 flex items-center">
                            @SearchIcon("w-5 h-5 text-gray-400")
                        </div>
                    </div>
                    
                    <!-- Filters -->
                    <div class="flex items-center space-x-3">
                        <div class="relative" x-data="{ open: false }">
                            <button @click="open = !open"
                                    class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                                @FilterIcon("w-4 h-4 mr-2")
                                Filter
                                @ChevronDownIcon("w-4 h-4 ml-2")
                            </button>
                            <!-- Filter dropdown content -->
                        </div>
                        
                        <div class="relative" x-data="{ open: false }">
                            <select class="border border-gray-300 rounded-lg text-sm">
                                <option>All status</option>
                                <option>Active</option>
                                <option>Inactive</option>
                            </select>
                        </div>
                    </div>
                </div>
            </section>
            
            <!-- Contacts Content -->
            <section id="contacts-content">
                @ContactsGrid(props.ContactGroups, props.ViewMode)
            </section>
        </div>
    }
}

templ ContactsGrid(groups []ContactGroup, viewMode string) {
    <div class="contacts-grid"
         :class="viewMode === 'list' ? 'space-y-4' : 'grid grid-cols-1 lg:grid-cols-3 gap-6'">
        for _, group := range groups {
            <div class="contact-group bg-white rounded-lg border border-gray-200 overflow-hidden">
                <div class="px-6 py-4 bg-gray-50 border-b border-gray-200">
                    <div class="flex items-center justify-between">
                        <h3 class="text-sm font-semibold text-gray-900">{ group.Title }</h3>
                        <span class="text-xs text-gray-500 bg-gray-200 px-2 py-1 rounded-full">
                            { fmt.Sprintf("%d", group.Count) }
                        </span>
                    </div>
                </div>
                <div class="p-4 space-y-3">
                    for _, contact := range group.Contacts {
                        @cards.ContactCard(contact)
                    }
                </div>
            </div>
        }
    </div>
}
```

#### **Alpine.js Contacts State**
```javascript
// Add to app.js
function contactsState() {
    return {
        viewMode: 'grid',
        activeTab: 'contacts',
        searchTerm: '',
        loading: false,
        
        init() {
            // Initialize contacts page
            this.loadContacts();
        },
        
        setViewMode(mode) {
            this.viewMode = mode;
            // Persist preference
            localStorage.setItem('contacts-view-mode', mode);
        },
        
        setActiveTab(tab) {
            this.activeTab = tab;
        },
        
        async search() {
            if (this.searchTerm.length < 2 && this.searchTerm.length > 0) {
                return; // Don't search for single characters
            }
            
            this.loading = true;
            try {
                const params = new URLSearchParams({
                    tab: this.activeTab,
                    search: this.searchTerm
                });
                
                const response = await fetch(`/api/contacts?${params}`);
                const html = await response.text();
                
                document.getElementById('contacts-content').innerHTML = html;
            } catch (error) {
                console.error('Search failed:', error);
            } finally {
                this.loading = false;
            }
        },
        
        async loadContacts() {
            // Load initial contacts data
            const savedViewMode = localStorage.getItem('contacts-view-mode');
            if (savedViewMode) {
                this.viewMode = savedViewMode;
            }
        }
    };
}
```
<!-- LLM-SECTION-PAGES-END -->

---

## 🚀 Performance & Production Optimization

<!-- LLM-SECTION-PERFORMANCE-START -->
### **1. Lazy Loading Implementation**

#### **Progressive Component Loading**
```go
// Lazy load heavy components
templ LazyDashboardCharts() {
    <div class="chart-container"
         hx-get="/api/dashboard/charts"
         hx-trigger="intersect once"
         hx-target="this"
         hx-swap="innerHTML"
         hx-indicator="#chart-loading">
        
        <!-- Loading skeleton -->
        <div id="chart-loading" class="space-y-4">
            <div class="skeleton h-6 w-32 mb-4"></div>
            <div class="skeleton h-64 w-full"></div>
        </div>
    </div>
}

// Progressive enhancement for forms
templ ContactForm(props ContactFormProps) {
    <form class="contact-form"
          action="/api/contacts"
          method="POST"
          hx-post="/api/contacts"
          hx-target="#contact-list"
          hx-swap="beforeend"
          hx-on::after-request="resetForm()">
        
        <!-- Form fields -->
        @FormFields(props)
        
        <!-- Submit button with loading state -->
        <button type="submit" 
                class="btn-primary"
                hx-disable-elt="this"
                hx-indicator="#form-loading">
            <span class="htmx-indicator" id="form-loading">
                @SpinnerIcon("w-4 h-4 animate-spin mr-2")
                Saving...
            </span>
            <span class="htmx-request-none">
                Create Contact
            </span>
        </button>
    </form>
}
```

### **2. Caching Strategies**

#### **HTMX Response Caching**
```javascript
// Add to app.js
// Configure HTMX caching
htmx.config.getCacheBusterParam = false;
htmx.config.refreshOnHistoryMiss = true;

// Custom caching for frequently accessed endpoints
const cache = new Map();
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

document.body.addEventListener('htmx:configRequest', function(evt) {
    const url = evt.detail.path;
    const method = evt.detail.verb;
    
    // Only cache GET requests
    if (method === 'GET') {
        const cacheKey = url;
        const cached = cache.get(cacheKey);
        
        if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
            // Return cached response
            evt.detail.xhr.response = cached.response;
            evt.preventDefault();
            return;
        }
    }
});

document.body.addEventListener('htmx:afterRequest', function(evt) {
    const url = evt.detail.xhr.responseURL;
    const method = evt.detail.requestConfig.verb;
    
    // Cache successful GET responses
    if (method === 'GET' && evt.detail.xhr.status === 200) {
        cache.set(url, {
            response: evt.detail.xhr.response,
            timestamp: Date.now()
        });
    }
});
```

### **3. Bundle Optimization**

#### **CSS Optimization**
```css
/* Critical CSS - inline in head */
.app-layout {
  display: grid;
  grid-template-areas: "header header header" "nav content context";
  grid-template-columns: 16rem 1fr 20rem;
  grid-template-rows: 4rem 1fr;
  min-height: 100vh;
}

.app-header, .app-nav, .app-content, .app-context {
  /* Critical layout styles */
}

/* Non-critical CSS - load asynchronously */
.card-hover-effects,
.advanced-animations,
.detailed-components {
  /* Load after initial render */
}
```

#### **JavaScript Code Splitting**
```javascript
// Core app functionality - load immediately
const coreApp = {
    init() {
        // Essential functionality
    }
};

// Advanced features - load on demand
async function loadAdvancedFeatures() {
    const { advancedCharts } = await import('./advanced-charts.js');
    const { reportingTools } = await import('./reporting.js');
    
    return { advancedCharts, reportingTools };
}

// Load advanced features when needed
document.addEventListener('DOMContentLoaded', async () => {
    coreApp.init();
    
    // Load advanced features on user interaction
    document.addEventListener('click', async (e) => {
        if (e.target.matches('[data-advanced-feature]')) {
            const features = await loadAdvancedFeatures();
            // Initialize advanced features
        }
    }, { once: true });
});
```

### **4. Error Boundaries & Fallbacks**

#### **Graceful Error Handling**
```go
// Error boundary component
templ ErrorBoundary(fallback templ.Component) {
    <div class="error-boundary" 
         hx-on:htmx:error="handleComponentError"
         hx-on:htmx:timeout="handleTimeout">
        { children... }
        
        <!-- Fallback content -->
        <div class="error-fallback hidden">
            @fallback
        </div>
    </div>
}

// Network error fallback
templ NetworkErrorFallback() {
    <div class="flex flex-col items-center justify-center p-8 text-center">
        @ExclamationIcon("w-12 h-12 text-gray-400 mb-4")
        <h3 class="text-lg font-medium text-gray-900 mb-2">Connection Error</h3>
        <p class="text-gray-600 mb-4">Unable to load content. Please check your connection.</p>
        <button onclick="location.reload()" 
                class="btn-primary">
            Try Again
        </button>
    </div>
}
```

#### **Error Handling JavaScript**
```javascript
// Global error handler
function handleComponentError(evt) {
    console.error('Component error:', evt.detail);
    
    const errorBoundary = evt.target.closest('.error-boundary');
    if (errorBoundary) {
        const fallback = errorBoundary.querySelector('.error-fallback');
        const content = errorBoundary.querySelector(':not(.error-fallback)');
        
        if (fallback && content) {
            content.style.display = 'none';
            fallback.style.display = 'block';
        }
    }
    
    // Report error to monitoring service
    if (window.errorReporting) {
        window.errorReporting.report(evt.detail);
    }
}

// Retry mechanism
function retryFailedRequest(element, retries = 3) {
    return new Promise((resolve, reject) => {
        const attempt = (remaining) => {
            htmx.trigger(element, 'retry');
            
            setTimeout(() => {
                if (element.classList.contains('htmx-request')) {
                    if (remaining > 0) {
                        attempt(remaining - 1);
                    } else {
                        reject(new Error('Max retries exceeded'));
                    }
                } else {
                    resolve();
                }
            }, 1000 * (4 - remaining)); // Exponential backoff
        };
        
        attempt(retries);
    });
}
```
<!-- LLM-SECTION-PERFORMANCE-END -->

---

## 🎯 Quick Implementation Checklist

### **Essential Components** ✅
- [ ] **Master Layout** - Three-zone responsive layout
- [ ] **Navigation** - Sidebar with collapsible sections  
- [ ] **Metric Cards** - Dashboard KPI components
- [ ] **Contact Cards** - CRM contact display
- [ ] **Data Tables** - Sortable, filterable tables
- [ ] **Modal System** - Overlay dialogs and forms

### **Page Templates** ✅
- [ ] **Dashboard Page** - Executive overview with metrics
- [ ] **Contact List** - CRM contact management
- [ ] **Task Kanban** - Project management board
- [ ] **Detail Views** - Item-specific information
- [ ] **Form Pages** - Data entry interfaces

### **Interactive Features** ✅
- [ ] **Real-time Updates** - HTMX + SSE integration
- [ ] **Search & Filter** - Dynamic content filtering
- [ ] **Drag & Drop** - Kanban task management
- [ ] **Progressive Loading** - Lazy loading patterns
- [ ] **Error Handling** - Graceful failure recovery

### **Production Ready** ✅
- [ ] **Performance** - Caching, lazy loading, code splitting
- [ ] **Accessibility** - WCAG compliance, screen readers
- [ ] **Responsive** - Mobile, tablet, desktop optimization
- [ ] **Error Boundaries** - Fallback components
- [ ] **Monitoring** - Error reporting and analytics

This implementation guide provides **production-ready code examples** and **best practices** for building the complete ERP interface using modern web technologies while maintaining **high performance** and **excellent user experience**.

---

*Follow this guide step-by-step to implement a **scalable, maintainable, and user-friendly** ERP interface that serves **business productivity** needs while providing **exceptional developer experience**.*