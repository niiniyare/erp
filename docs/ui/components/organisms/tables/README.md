# Advanced Table Components

**FILE PURPOSE**: Complete data table system with advanced features  
**SCOPE**: All table variants, data management, and interactive features  
**TARGET AUDIENCE**: Developers implementing data-intensive interfaces

## 📋 Table System Overview

Our advanced table system provides comprehensive data display and management capabilities for enterprise workflows. Built with schema-driven architecture, it supports sorting, filtering, pagination, selection, and real-time updates.

### Schema References
- **Primary Schema**: `TableSchema.json`, `TableSchema2.json`
- **Related Schemas**: `TableColumn.json`, `TableColumnObject.json`, `TableColumnWithType.json`
- **CRUD Integration**: `CRUDSchema.json`, `CRUD2Schema.json`

### Table Component Types

| Component | Purpose | Features | Schema |
|-----------|---------|----------|--------|
| **[Basic Table](table.md)** | Simple data display | Columns, rows, styling | `TableSchema.json` |
| **[Advanced Table](advanced-table.md)** | Feature-rich data management | Sorting, filtering, pagination | `TableSchema2.json` |
| **[CRUD Table](crud-table.md)** | Full data operations | Create, read, update, delete | `CRUDSchema.json` |
| **[Data Grid](data-grid.md)** | Enterprise data grid | Virtual scrolling, cell editing | `CRUD2Schema.json` |

## 🏗️ Architecture Overview

### Component Hierarchy
```
TableContainer
├── TableHeader
│   ├── TableTitle
│   ├── TableSearch
│   ├── TableFilters
│   └── TableActions
├── TableContent
│   ├── TableHead
│   │   └── TableColumn[]
│   └── TableBody
│       └── TableRow[]
│           └── TableCell[]
├── TableFooter
│   ├── TablePagination
│   ├── TableInfo
│   └── TablePerPage
└── TableModals
    ├── FilterModal
    ├── ColumnModal
    └── ExportModal
```

### Props Structure
```go
type TableProps struct {
    // Data
    Data        []map[string]interface{} `json:"data"`
    Columns     []ColumnConfig           `json:"columns"`
    
    // Features
    Sortable    bool                     `json:"sortable"`
    Filterable  bool                     `json:"filterable"`
    Searchable  bool                     `json:"searchable"`
    Selectable  bool                     `json:"selectable"`
    Paginated   bool                     `json:"paginated"`
    
    // Pagination
    Page        int                      `json:"page"`
    PerPage     int                      `json:"perPage"`
    Total       int                      `json:"total"`
    
    // Actions
    Actions     []TableAction            `json:"actions"`
    BulkActions []BulkAction            `json:"bulkActions"`
    
    // Styling
    Striped     bool                     `json:"striped"`
    Hoverable   bool                     `json:"hoverable"`
    Bordered    bool                     `json:"bordered"`
    Compact     bool                     `json:"compact"`
    Responsive  bool                     `json:"responsive"`
    
    // Events
    OnSort      string                   `json:"onSort"`
    OnFilter    string                   `json:"onFilter"`
    OnSelect    string                   `json:"onSelect"`
    OnAction    string                   `json:"onAction"`
}
```

## 📊 Basic Table Implementation

### Simple Data Table
```go
templ BasicTable(props TableProps) {
    <div class="table-container">
        <table class={ buildTableClasses(props) }>
            <thead>
                <tr>
                    if props.Selectable {
                        <th class="select-column">
                            <input 
                                type="checkbox"
                                class="select-all-checkbox"
                                @change="toggleAllRows($event.target.checked)"
                            />
                        </th>
                    }
                    for _, column := range props.Columns {
                        @TableHeader(column, props.Sortable)
                    }
                    if len(props.Actions) > 0 {
                        <th class="actions-column">Actions</th>
                    }
                </tr>
            </thead>
            <tbody>
                for i, row := range props.Data {
                    @TableRow(row, props.Columns, props.Actions, props.Selectable, i)
                }
            </tbody>
        </table>
    </div>
}

templ TableHeader(column ColumnConfig, sortable bool) {
    <th 
        class={ buildHeaderClasses(column, sortable) }
        style={ buildHeaderStyles(column) }
        if sortable && column.Sortable {
            @click={ fmt.Sprintf("sortTable('%s')", column.Key) }
        }>
        
        <div class="header-content">
            <span class="header-text">{ column.Title }</span>
            
            if sortable && column.Sortable {
                <div class="sort-indicators">
                    <span class="sort-asc" data-sort="asc">▲</span>
                    <span class="sort-desc" data-sort="desc">▼</span>
                </div>
            }
            
            if column.Filterable {
                <button class="filter-toggle" @click="toggleFilter('{ column.Key }')">
                    🔽
                </button>
            }
        </div>
        
        if column.Resizable {
            <div class="resize-handle" @mousedown="startResize('{ column.Key }', $event)"></div>
        }
    </th>
}

templ TableRow(row map[string]interface{}, columns []ColumnConfig, actions []TableAction, selectable bool, index int) {
    <tr 
        class="table-row"
        data-row-id={ fmt.Sprintf("%v", row["id"]) }
        :class="{ 'selected': selectedRows.includes('{ fmt.Sprintf("%v", row["id"]) }') }">
        
        if selectable {
            <td class="select-cell">
                <input 
                    type="checkbox"
                    class="row-select"
                    :checked="selectedRows.includes('{ fmt.Sprintf("%v", row["id"]) }')"
                    @change="toggleRowSelection('{ fmt.Sprintf("%v", row["id"]) }', $event.target.checked)"
                />
            </td>
        }
        
        for _, column := range columns {
            @TableCell(row, column)
        }
        
        if len(actions) > 0 {
            <td class="actions-cell">
                @TableActions(actions, row)
            </td>
        }
    </tr>
}

templ TableCell(row map[string]interface{}, column ColumnConfig) {
    <td 
        class={ buildCellClasses(column) }
        style={ buildCellStyles(column) }>
        
        switch column.Type {
        case "text":
            { fmt.Sprintf("%v", row[column.Key]) }
        case "number":
            <span class="number-value">{ formatNumber(row[column.Key]) }</span>
        case "currency":
            <span class="currency-value">{ formatCurrency(row[column.Key]) }</span>
        case "date":
            <span class="date-value">{ formatDate(row[column.Key]) }</span>
        case "status":
            @StatusBadge(StatusProps{
                Status: fmt.Sprintf("%v", row[column.Key]),
                Variant: getStatusVariant(fmt.Sprintf("%v", row[column.Key])),
            })
        case "boolean":
            @BooleanIndicator(row[column.Key] == true)
        case "link":
            <a href={ fmt.Sprintf("%v", row[column.LinkField]) } class="table-link">
                { fmt.Sprintf("%v", row[column.Key]) }
            </a>
        case "image":
            <img src={ fmt.Sprintf("%v", row[column.Key]) } alt="" class="table-image" />
        case "custom":
            @RenderCustomCell(row, column)
        default:
            { fmt.Sprintf("%v", row[column.Key]) }
        }
    </td>
}
```

## 🔧 Advanced Features

### Column Configuration
```go
type ColumnConfig struct {
    // Basic properties
    Key         string      `json:"key"`         // Data field key
    Title       string      `json:"title"`       // Display title
    Type        ColumnType  `json:"type"`        // Data type
    Width       string      `json:"width"`       // Column width
    MinWidth    string      `json:"minWidth"`    // Minimum width
    MaxWidth    string      `json:"maxWidth"`    // Maximum width
    
    // Features
    Sortable    bool        `json:"sortable"`    // Enable sorting
    Filterable  bool        `json:"filterable"`  // Enable filtering
    Resizable   bool        `json:"resizable"`   // Enable resizing
    Searchable  bool        `json:"searchable"`  // Include in search
    
    // Display
    Hidden      bool        `json:"hidden"`      // Hide column
    Fixed       FixedType   `json:"fixed"`       // left, right, none
    Align       AlignType   `json:"align"`       // left, center, right
    
    // Formatting
    Format      string      `json:"format"`      // Display format
    Prefix      string      `json:"prefix"`      // Value prefix
    Suffix      string      `json:"suffix"`      // Value suffix
    
    // Links and actions
    LinkField   string      `json:"linkField"`   // Field for link href
    OnClick     string      `json:"onClick"`     // Click handler
    
    // Custom rendering
    Template    string      `json:"template"`    // Custom template
    Renderer    string      `json:"renderer"`    // Custom renderer function
}

type ColumnType string

const (
    ColumnText     ColumnType = "text"
    ColumnNumber   ColumnType = "number"
    ColumnCurrency ColumnType = "currency"
    ColumnDate     ColumnType = "date"
    ColumnDateTime ColumnType = "datetime"
    ColumnStatus   ColumnType = "status"
    ColumnBoolean  ColumnType = "boolean"
    ColumnLink     ColumnType = "link"
    ColumnImage    ColumnType = "image"
    ColumnCustom   ColumnType = "custom"
)
```

### Sorting Implementation
```go
templ SortableTable(props TableProps) {
    <div x-data={ fmt.Sprintf(`{
        sortBy: '%s',
        sortDirection: '%s',
        data: %s,
        sortColumn(column) {
            if (this.sortBy === column) {
                this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortBy = column;
                this.sortDirection = 'asc';
            }
            this.sortData();
        },
        sortData() {
            this.data.sort((a, b) => {
                let aVal = a[this.sortBy];
                let bVal = b[this.sortBy];
                
                // Handle different data types
                if (typeof aVal === 'string') {
                    aVal = aVal.toLowerCase();
                    bVal = bVal.toLowerCase();
                } else if (typeof aVal === 'number') {
                    aVal = parseFloat(aVal);
                    bVal = parseFloat(bVal);
                } else if (aVal instanceof Date) {
                    aVal = aVal.getTime();
                    bVal = bVal.getTime();
                }
                
                if (this.sortDirection === 'asc') {
                    return aVal > bVal ? 1 : -1;
                } else {
                    return aVal < bVal ? 1 : -1;
                }
            });
        }
    }`, props.SortBy, props.SortDirection, toJSON(props.Data)) }>
        
        @BasicTable(props)
    </div>
}
```

### Filtering System
```go
templ FilterableTable(props TableProps) {
    <div x-data={ fmt.Sprintf(`{
        filters: {},
        filteredData: %s,
        applyFilters() {
            this.filteredData = this.data.filter(row => {
                return Object.keys(this.filters).every(key => {
                    const filter = this.filters[key];
                    const value = row[key];
                    
                    if (!filter || filter === '') return true;
                    
                    // Different filter types
                    switch (filter.type) {
                        case 'text':
                            return value.toString().toLowerCase().includes(filter.value.toLowerCase());
                        case 'number':
                            return this.compareNumbers(value, filter.operator, filter.value);
                        case 'date':
                            return this.compareDates(value, filter.operator, filter.value);
                        case 'select':
                            return filter.values.includes(value);
                        default:
                            return value.toString().toLowerCase().includes(filter.toString().toLowerCase());
                    }
                });
            });
        },
        compareNumbers(value, operator, filterValue) {
            const num = parseFloat(value);
            const filter = parseFloat(filterValue);
            switch (operator) {
                case 'eq': return num === filter;
                case 'gt': return num > filter;
                case 'lt': return num < filter;
                case 'gte': return num >= filter;
                case 'lte': return num <= filter;
                default: return true;
            }
        }
    }`, toJSON(props.Data)) }>
        
        // Filter controls
        <div class="table-filters">
            for _, column := range props.Columns {
                if column.Filterable {
                    @ColumnFilter(column)
                }
            }
        </div>
        
        // Table with filtered data
        @BasicTable(props)
    </div>
}

templ ColumnFilter(column ColumnConfig) {
    <div class="filter-control" x-show="showFilters">
        <label class="filter-label">{ column.Title }</label>
        
        switch column.Type {
        case ColumnText:
            <input 
                type="text"
                class="filter-input"
                x-model={ fmt.Sprintf("filters.%s", column.Key) }
                @input="applyFilters()"
                placeholder={ fmt.Sprintf("Filter %s...", column.Title) }
            />
        case ColumnNumber, ColumnCurrency:
            <div class="number-filter">
                <select x-model={ fmt.Sprintf("filters.%s.operator", column.Key) }>
                    <option value="eq">Equals</option>
                    <option value="gt">Greater than</option>
                    <option value="lt">Less than</option>
                    <option value="gte">Greater or equal</option>
                    <option value="lte">Less or equal</option>
                </select>
                <input 
                    type="number"
                    x-model={ fmt.Sprintf("filters.%s.value", column.Key) }
                    @input="applyFilters()"
                />
            </div>
        case ColumnDate:
            <div class="date-filter">
                <input 
                    type="date"
                    x-model={ fmt.Sprintf("filters.%s.from", column.Key) }
                    @change="applyFilters()"
                />
                <span>to</span>
                <input 
                    type="date"
                    x-model={ fmt.Sprintf("filters.%s.to", column.Key) }
                    @change="applyFilters()"
                />
            </div>
        case ColumnStatus:
            <select 
                x-model={ fmt.Sprintf("filters.%s", column.Key) }
                @change="applyFilters()"
                multiple>
                for _, option := range column.Options {
                    <option value={ option.Value }>{ option.Label }</option>
                }
            </select>
        }
    </div>
}
```

### Pagination Implementation
```go
templ PaginatedTable(props TableProps) {
    <div x-data={ fmt.Sprintf(`{
        currentPage: %d,
        perPage: %d,
        totalItems: %d,
        data: %s,
        get totalPages() {
            return Math.ceil(this.totalItems / this.perPage);
        },
        get paginatedData() {
            const start = (this.currentPage - 1) * this.perPage;
            const end = start + this.perPage;
            return this.data.slice(start, end);
        },
        changePage(page) {
            if (page >= 1 && page <= this.totalPages) {
                this.currentPage = page;
            }
        },
        changePerPage(newPerPage) {
            this.perPage = parseInt(newPerPage);
            this.currentPage = 1; // Reset to first page
        }
    }`, props.Page, props.PerPage, props.Total, toJSON(props.Data)) }>
        
        // Table content
        @BasicTable(props)
        
        // Pagination controls
        <div class="table-pagination">
            <div class="pagination-info">
                <span x-text="`Showing ${(currentPage - 1) * perPage + 1} to ${Math.min(currentPage * perPage, totalItems)} of ${totalItems} entries`"></span>
            </div>
            
            <div class="pagination-controls">
                <button 
                    @click="changePage(1)"
                    :disabled="currentPage === 1"
                    class="pagination-btn">
                    First
                </button>
                
                <button 
                    @click="changePage(currentPage - 1)"
                    :disabled="currentPage === 1"
                    class="pagination-btn">
                    Previous
                </button>
                
                <template x-for="page in Array.from({length: totalPages}, (_, i) => i + 1).slice(Math.max(0, currentPage - 3), currentPage + 2)">
                    <button 
                        @click="changePage(page)"
                        :class="{ 'active': page === currentPage }"
                        class="pagination-btn page-btn"
                        x-text="page">
                    </button>
                </template>
                
                <button 
                    @click="changePage(currentPage + 1)"
                    :disabled="currentPage === totalPages"
                    class="pagination-btn">
                    Next
                </button>
                
                <button 
                    @click="changePage(totalPages)"
                    :disabled="currentPage === totalPages"
                    class="pagination-btn">
                    Last
                </button>
            </div>
            
            <div class="per-page-control">
                <label>Show:</label>
                <select @change="changePerPage($event.target.value)" :value="perPage">
                    <option value="10">10</option>
                    <option value="25">25</option>
                    <option value="50">50</option>
                    <option value="100">100</option>
                </select>
                <span>per page</span>
            </div>
        </div>
    </div>
}
```

## 🎯 Row Selection and Bulk Actions

### Selection Implementation
```go
templ SelectableTable(props TableProps) {
    <div x-data={ fmt.Sprintf(`{
        selectedRows: [],
        allSelected: false,
        get selectedCount() {
            return this.selectedRows.length;
        },
        get isIndeterminate() {
            return this.selectedCount > 0 && this.selectedCount < this.data.length;
        },
        toggleAllRows(checked) {
            if (checked) {
                this.selectedRows = this.data.map(row => row.id);
            } else {
                this.selectedRows = [];
            }
            this.allSelected = checked;
        },
        toggleRowSelection(rowId, checked) {
            if (checked) {
                if (!this.selectedRows.includes(rowId)) {
                    this.selectedRows.push(rowId);
                }
            } else {
                this.selectedRows = this.selectedRows.filter(id => id !== rowId);
            }
            this.updateAllSelectedState();
        },
        updateAllSelectedState() {
            this.allSelected = this.selectedRows.length === this.data.length;
        },
        clearSelection() {
            this.selectedRows = [];
            this.allSelected = false;
        }
    }`) }>
        
        // Bulk actions bar
        <div x-show="selectedCount > 0" class="bulk-actions-bar">
            <div class="selection-info">
                <span x-text="`${selectedCount} items selected`"></span>
                <button @click="clearSelection()" class="clear-selection">Clear</button>
            </div>
            
            <div class="bulk-actions">
                for _, action := range props.BulkActions {
                    @BulkActionButton(action)
                }
            </div>
        </div>
        
        // Table with selection
        @BasicTable(props)
    </div>
}

templ BulkActionButton(action BulkAction) {
    <button 
        class={ fmt.Sprintf("bulk-action-btn bulk-action-%s", action.Variant) }
        @click={ fmt.Sprintf("%s(selectedRows)", action.Handler) }
        :disabled="selectedCount === 0">
        
        if action.Icon != "" {
            @Icon(IconProps{Name: action.Icon, Size: "16"})
        }
        <span>{ action.Label }</span>
    </button>
}
```

### Action Definitions
```go
type TableAction struct {
    Label    string      `json:"label"`
    Icon     string      `json:"icon"`
    Variant  string      `json:"variant"`  // primary, secondary, danger
    Handler  string      `json:"handler"`  // JavaScript function
    Confirm  string      `json:"confirm"`  // Confirmation message
    Disabled string      `json:"disabled"` // Disable condition
}

type BulkAction struct {
    Label    string      `json:"label"`
    Icon     string      `json:"icon"`
    Variant  string      `json:"variant"`  // primary, secondary, danger
    Handler  string      `json:"handler"`  // JavaScript function
    Confirm  string      `json:"confirm"`  // Confirmation message
    MinItems int         `json:"minItems"` // Minimum selected items
    MaxItems int         `json:"maxItems"` // Maximum selected items
}

// Example usage
tableProps := TableProps{
    Actions: []TableAction{
        {
            Label:   "Edit",
            Icon:    "edit",
            Variant: "secondary",
            Handler: "editItem",
        },
        {
            Label:   "Delete", 
            Icon:    "trash",
            Variant: "danger",
            Handler: "deleteItem",
            Confirm: "Are you sure you want to delete this item?",
        },
    },
    BulkActions: []BulkAction{
        {
            Label:    "Delete Selected",
            Icon:     "trash",
            Variant:  "danger",
            Handler:  "deleteMultiple",
            Confirm:  "Are you sure you want to delete the selected items?",
            MinItems: 1,
        },
        {
            Label:   "Export Selected",
            Icon:    "download",
            Variant: "secondary", 
            Handler: "exportSelected",
        },
    },
}
```

## 🎨 Styling System

### Base Table Styles
```css
.table-container {
    overflow-x: auto;
    border: 1px solid var(--color-border-light);
    border-radius: var(--radius-lg);
    background: var(--color-bg-surface);
}

.table {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--font-size-sm);
    
    th {
        background: var(--color-bg-secondary);
        font-weight: var(--font-weight-semibold);
        color: var(--color-text-secondary);
        padding: var(--space-md);
        text-align: left;
        border-bottom: 1px solid var(--color-border-light);
        user-select: none;
        white-space: nowrap;
        
        &.sortable {
            cursor: pointer;
            
            &:hover {
                background: var(--color-bg-hover);
            }
        }
        
        .header-content {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: var(--space-sm);
        }
        
        .sort-indicators {
            display: flex;
            flex-direction: column;
            opacity: 0.3;
            
            .sort-asc,
            .sort-desc {
                font-size: 10px;
                line-height: 1;
            }
            
            &.active {
                opacity: 1;
                color: var(--color-primary);
            }
        }
    }
    
    td {
        padding: var(--space-md);
        border-bottom: 1px solid var(--color-border-light);
        color: var(--color-text-primary);
        vertical-align: middle;
        
        &.number-cell {
            text-align: right;
            font-family: var(--font-family-mono);
        }
        
        &.currency-cell {
            text-align: right;
            font-family: var(--font-family-mono);
            color: var(--color-text-success);
        }
        
        &.date-cell {
            font-family: var(--font-family-mono);
            color: var(--color-text-secondary);
        }
    }
    
    tbody tr {
        transition: var(--transition-quick);
        
        &:hover {
            background: var(--color-bg-hover);
        }
        
        &.selected {
            background: var(--color-primary-light);
            border-left: 3px solid var(--color-primary);
        }
    }
    
    // Table variants
    &.striped tbody tr:nth-child(even) {
        background: var(--color-bg-zebra);
    }
    
    &.bordered {
        th, td {
            border-right: 1px solid var(--color-border-light);
            
            &:last-child {
                border-right: none;
            }
        }
    }
    
    &.compact {
        th, td {
            padding: var(--space-sm) var(--space-md);
        }
    }
}
```

### Responsive Table Design
```css
@media (max-width: 767px) {
    .table-container {
        /* Switch to card layout on mobile */
        .table {
            display: block;
            
            thead {
                display: none;
            }
            
            tbody,
            tr,
            td {
                display: block;
            }
            
            tr {
                border: 1px solid var(--color-border-light);
                border-radius: var(--radius-md);
                margin-bottom: var(--space-md);
                padding: var(--space-md);
                background: var(--color-bg-surface);
            }
            
            td {
                border: none;
                padding: var(--space-sm) 0;
                
                &::before {
                    content: attr(data-label);
                    font-weight: var(--font-weight-semibold);
                    color: var(--color-text-secondary);
                    display: block;
                    margin-bottom: var(--space-xs);
                }
            }
        }
    }
    
    /* Stack pagination controls */
    .table-pagination {
        flex-direction: column;
        gap: var(--space-md);
        
        .pagination-controls {
            justify-content: center;
        }
    }
    
    /* Hide complex features on mobile */
    .bulk-actions-bar,
    .table-filters {
        display: none;
    }
}
```

## 📱 Mobile Optimizations

### Touch-Friendly Interactions
```css
@media (hover: none) {
    .table {
        th.sortable {
            /* Remove hover effects on touch */
            &:hover {
                background: var(--color-bg-secondary);
            }
        }
        
        /* Larger touch targets */
        .pagination-btn,
        .action-btn {
            min-height: 44px;
            min-width: 44px;
        }
        
        /* Simplified actions */
        .table-actions {
            .action-btn {
                padding: var(--space-md);
                
                .action-text {
                    display: none; /* Show icons only */
                }
            }
        }
    }
}
```

### Virtual Scrolling for Large Datasets
```go
templ VirtualTable(props TableProps) {
    <div x-data={ fmt.Sprintf(`{
        itemHeight: 40,
        containerHeight: 400,
        scrollTop: 0,
        data: %s,
        get visibleData() {
            const startIndex = Math.floor(this.scrollTop / this.itemHeight);
            const endIndex = Math.min(
                startIndex + Math.ceil(this.containerHeight / this.itemHeight) + 5,
                this.data.length
            );
            return this.data.slice(startIndex, endIndex).map((item, index) => ({
                ...item,
                originalIndex: startIndex + index
            }));
        },
        get totalHeight() {
            return this.data.length * this.itemHeight;
        },
        get offsetY() {
            return Math.floor(this.scrollTop / this.itemHeight) * this.itemHeight;
        }
    }`, toJSON(props.Data)) }>
        
        <div class="virtual-table-container" style="height: 400px; overflow: auto;">
            <div 
                class="virtual-spacer"
                :style="`height: ${totalHeight}px; position: relative;`">
                
                <div 
                    class="visible-rows"
                    :style="`transform: translateY(${offsetY}px);`">
                    
                    <template x-for="row in visibleData" :key="row.originalIndex">
                        @VirtualTableRow()
                    </template>
                </div>
            </div>
        </div>
    </div>
}
```

## 🧪 Testing Strategy

### Unit Tests
```go
func TestTableComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    TableProps
        expected []string
    }{
        {
            name: "basic table",
            props: TableProps{
                Columns: []ColumnConfig{
                    {Key: "name", Title: "Name", Type: ColumnText},
                    {Key: "email", Title: "Email", Type: ColumnText},
                },
                Data: []map[string]interface{}{
                    {"name": "John", "email": "john@example.com"},
                },
            },
            expected: []string{"table", "Name", "Email", "John"},
        },
        {
            name: "sortable table",
            props: TableProps{
                Sortable: true,
                Columns: []ColumnConfig{
                    {Key: "name", Title: "Name", Sortable: true},
                },
            },
            expected: []string{"sortable", "sort-indicators"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := renderTable(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, result, expected)
            }
        })
    }
}
```

### Performance Tests
```javascript
describe('Table Performance', () => {
    test('handles large datasets efficiently', async () => {
        const largeDataset = Array.from({length: 10000}, (_, i) => ({
            id: i,
            name: `Item ${i}`,
            value: Math.random() * 1000
        }));
        
        const startTime = performance.now();
        render(<Table data={largeDataset} columns={columns} />);
        const endTime = performance.now();
        
        expect(endTime - startTime).toBeLessThan(100); // Should render in under 100ms
    });
    
    test('virtual scrolling maintains performance', async () => {
        // Test virtual scrolling with 100k items
        // Should maintain 60fps scrolling
    });
});
```

## 📚 Usage Examples

### Basic Data Table
```go
templ ProductTable() {
    @Table(TableProps{
        Columns: []ColumnConfig{
            {Key: "sku", Title: "SKU", Type: ColumnText, Width: "120px"},
            {Key: "name", Title: "Product Name", Type: ColumnText, Sortable: true},
            {Key: "price", Title: "Price", Type: ColumnCurrency, Align: AlignRight},
            {Key: "stock", Title: "Stock", Type: ColumnNumber, Align: AlignRight},
            {Key: "status", Title: "Status", Type: ColumnStatus},
        },
        Data: productData,
        Sortable: true,
        Filterable: true,
        Paginated: true,
        PerPage: 25,
        Actions: []TableAction{
            {Label: "Edit", Icon: "edit", Handler: "editProduct"},
            {Label: "Delete", Icon: "trash", Variant: "danger", Handler: "deleteProduct"},
        },
    })
}
```

### Advanced CRUD Table
```go
templ UserManagementTable() {
    @CRUDTable(CRUDTableProps{
        Title: "User Management",
        API: "/api/users",
        Columns: []ColumnConfig{
            {Key: "avatar", Title: "", Type: ColumnImage, Width: "50px"},
            {Key: "name", Title: "Name", Type: ColumnText, Sortable: true, Searchable: true},
            {Key: "email", Title: "Email", Type: ColumnText, Sortable: true, Searchable: true},
            {Key: "role", Title: "Role", Type: ColumnStatus, Filterable: true},
            {Key: "lastLogin", Title: "Last Login", Type: ColumnDateTime, Sortable: true},
            {Key: "active", Title: "Active", Type: ColumnBoolean},
        },
        Features: CRUDFeatures{
            Create: true,
            Read: true,
            Update: true,
            Delete: true,
            Search: true,
            Filter: true,
            Export: true,
            Import: true,
        },
        Permissions: map[string]bool{
            "create": hasPermission("users.create"),
            "update": hasPermission("users.update"),
            "delete": hasPermission("users.delete"),
        },
    })
}
```

## 🔗 Related Components

- **[CRUD Operations](crud-table.md)**: Full data management
- **[Data Grid](data-grid.md)**: Enterprise grid features
- **[Search and Filter](../../molecules/search-filter/)**: Enhanced filtering
- **[Pagination](../../molecules/pagination/)**: Advanced pagination
- **[Export Tools](../../molecules/export/)**: Data export functionality

---

**Component Status**: 🔄 In Development  
**Schema Reference**: `TableSchema.json`, `TableSchema2.json`, `CRUDSchema.json`  
**Features**: Sorting, filtering, pagination, selection, responsive design  
**Performance**: Supports 10k+ rows with virtual scrolling