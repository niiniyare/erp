# Awo DataTable

A **headless, TypeScript-first DataTable** that works with any design system. Built for performance, flexibility, and developer experience.

## ✨ Features

- 🎨 **Headless** - Bring your own UI (works with any design system)
- ⚡ **High Performance** - Virtual scrolling for large datasets
- 🔄 **Server-Side & Client-Side** - Full control over data loading
- 📱 **Responsive** - Works on any device
- 🎯 **TypeScript First** - Full type safety
- 🔌 **Plugin System** - Extensible architecture
- 💾 **State Management** - Built-in state persistence
- 🎭 **Event System** - React to any change
- 📊 **Advanced Features**:
  - Multi-column sorting
  - Complex filtering (14 operators)
  - Global search
  - Row selection (single/multiple)
  - Column visibility
  - Export (CSV, JSON, TXT)
  - Pagination
  - And more!

## 📦 Installation

```bash
npm install @awo/datatable
# or
yarn add @awo/datatable
# or
pnpm add @awo/datatable
```

## 🚀 Quick Start

### Basic Usage

```typescript
import { createDataTable } from '@awo/datatable';

// Define your data structure
interface User {
  id: number;
  name: string;
  email: string;
  role: string;
}

// Create the table
const table = createDataTable<User>({
  columns: [
    { id: 'id', field: 'id', label: 'ID' },
    { id: 'name', field: 'name', label: 'Name' },
    { id: 'email', field: 'email', label: 'Email' },
    { id: 'role', field: 'role', label: 'Role' },
  ],
  data: [
    { id: 1, name: 'John Doe', email: 'john@example.com', role: 'Admin' },
    { id: 2, name: 'Jane Smith', email: 'jane@example.com', role: 'User' },
  ],
});

// Listen to state changes
table.on('state:change', (state) => {
  console.log('State updated:', state);
  renderTable(state.visibleRows);
});

// Use the API
table.search('john');
table.sort('name', 'asc');
table.goToPage(1);
```

## 🎯 Framework Integration

### React Example

```tsx
import { createDataTable, DataTable } from '@awo/datatable';
import { useEffect, useState, useMemo } from 'react';

interface User {
  id: number;
  name: string;
  email: string;
}

function UsersTable() {
  const [state, setState] = useState(null);

  const table = useMemo(() => {
    return createDataTable<User>({
      columns: [
        { id: 'id', field: 'id', label: 'ID' },
        { id: 'name', field: 'name', label: 'Name', sortable: true },
        { id: 'email', field: 'email', label: 'Email', sortable: true },
      ],
      data: [],
      pagination: { enabled: true, pageSize: 10 },
      selection: { enabled: true, multiple: true },
    });
  }, []);

  useEffect(() => {
    // Listen to state changes
    const unsubscribe = table.on('state:change', setState);
    
    // Load data
    fetchUsers().then(users => table.setData(users));

    return () => {
      unsubscribe();
      table.destroy();
    };
  }, [table]);

  if (!state) return <div>Loading...</div>;

  return (
    <div>
      {/* Search */}
      <input
        type="text"
        placeholder="Search..."
        onChange={(e) => table.search(e.target.value)}
      />

      {/* Table */}
      <table className="w-full">
        <thead>
          <tr>
            {state.selection?.enabled && (
              <th>
                <input
                  type="checkbox"
                  checked={state.selectionState.isAllSelected}
                  onChange={(e) => e.target.checked ? table.selectAll() : table.deselectAll()}
                />
              </th>
            )}
            {state.visibleColumns.map((col) => (
              <th
                key={col.id}
                onClick={() => table.sort(col.id)}
                className="cursor-pointer"
              >
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {state.visibleRows.map((row) => (
            <tr key={row.id} className={row.selected ? 'bg-blue-50' : ''}>
              {state.selection?.enabled && (
                <td>
                  <input
                    type="checkbox"
                    checked={row.selected}
                    onChange={() => table.toggleRowSelection(row.id)}
                  />
                </td>
              )}
              {state.visibleColumns.map((col) => (
                <td key={col.id}>{row.data[col.field]}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>

      {/* Pagination */}
      {state.paginationState && (
        <div className="flex gap-2 mt-4">
          <button onClick={() => table.firstPage()}>First</button>
          <button onClick={() => table.previousPage()}>Previous</button>
          <span>
            Page {state.paginationState.pageIndex + 1} of {state.paginationState.totalPages}
          </span>
          <button onClick={() => table.nextPage()}>Next</button>
          <button onClick={() => table.lastPage()}>Last</button>
        </div>
      )}
    </div>
  );
}
```

### Vue 3 Example

```vue
<template>
  <div v-if="state">
    <!-- Search -->
    <input
      v-model="searchQuery"
      @input="handleSearch"
      type="text"
      placeholder="Search..."
    />

    <!-- Table -->
    <table>
      <thead>
        <tr>
          <th v-for="col in state.visibleColumns" :key="col.id" @click="table.sort(col.id)">
            {{ col.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in state.visibleRows" :key="row.id">
          <td v-for="col in state.visibleColumns" :key="col.id">
            {{ row.data[col.field] }}
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Pagination -->
    <div v-if="state.paginationState">
      <button @click="table.previousPage()">Previous</button>
      <span>Page {{ state.paginationState.pageIndex + 1 }}</span>
      <button @click="table.nextPage()">Next</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { createDataTable } from '@awo/datatable';

interface User {
  id: number;
  name: string;
  email: string;
}

const table = createDataTable<User>({
  columns: [
    { id: 'name', field: 'name', label: 'Name' },
    { id: 'email', field: 'email', label: 'Email' },
  ],
  data: [],
});

const state = ref(null);
const searchQuery = ref('');

const handleSearch = () => {
  table.search(searchQuery.value);
};

onMounted(() => {
  const unsubscribe = table.on('state:change', (newState) => {
    state.value = newState;
  });

  // Load data
  fetchUsers().then(users => table.setData(users));

  onUnmounted(() => {
    unsubscribe();
    table.destroy();
  });
});
</script>
```

### Vanilla JavaScript

```javascript
import { createDataTable } from '@awo/datatable';

const table = createDataTable({
  columns: [
    { id: 'id', field: 'id', label: 'ID' },
    { id: 'name', field: 'name', label: 'Name' },
    { id: 'email', field: 'email', label: 'Email' },
  ],
  data: users,
});

// Listen to changes
table.on('state:change', (state) => {
  const tbody = document.querySelector('#data-table tbody');
  tbody.innerHTML = '';

  state.visibleRows.forEach((row) => {
    const tr = document.createElement('tr');
    state.visibleColumns.forEach((col) => {
      const td = document.createElement('td');
      td.textContent = row.data[col.field];
      tr.appendChild(td);
    });
    tbody.appendChild(tr);
  });
});

// Setup search
document.querySelector('#search').addEventListener('input', (e) => {
  table.search(e.target.value);
});
```

## 🎨 Design System Integration

### Tailwind CSS

```tsx
// Reusable Table Component
function TailwindTable({ table, state }) {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            {state.visibleColumns.map((col) => (
              <th
                key={col.id}
                className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                onClick={() => table.sort(col.id)}
              >
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {state.visibleRows.map((row) => (
            <tr
              key={row.id}
              className="hover:bg-gray-50 transition-colors"
            >
              {state.visibleColumns.map((col) => (
                <td key={col.id} className="px-6 py-4 whitespace-nowrap">
                  {row.data[col.field]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### shadcn/ui

```tsx
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

function ShadcnTable({ table, state }) {
  return (
    <div className="space-y-4">
      <Input
        placeholder="Search..."
        onChange={(e) => table.search(e.target.value)}
      />

      <Table>
        <TableHeader>
          <TableRow>
            {state.visibleColumns.map((col) => (
              <TableHead key={col.id} onClick={() => table.sort(col.id)}>
                {col.label}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {state.visibleRows.map((row) => (
            <TableRow key={row.id}>
              {state.visibleColumns.map((col) => (
                <TableCell key={col.id}>
                  {row.data[col.field]}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>

      <div className="flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          {state.paginationState.totalRows} total rows
        </div>
        <div className="flex gap-2">
          <Button onClick={() => table.previousPage()} size="sm">
            Previous
          </Button>
          <Button onClick={() => table.nextPage()} size="sm">
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
```

## 🔧 Advanced Features

### Server-Side Data

```typescript
const table = createDataTable<User>({
  columns: [...],
  dataMode: 'server',
  serverSide: {
    url: '/api/users',
    method: 'GET',
    headers: {
      'Authorization': 'Bearer token',
    },
    transformer: (response) => ({
      data: response.data.items,
      total: response.data.total,
    }),
  },
  pagination: { enabled: true, pageSize: 20 },
});

// Load data
await table.reload();
```

### Advanced Filtering

```typescript
// Add specific filter
table.addFilter({
  columnId: 'age',
  operator: 'greaterThan',
  value: 18,
});

// Between filter
table.addFilter({
  columnId: 'salary',
  operator: 'between',
  value: [50000, 100000],
});

// Multiple filters
table.filter([
  { columnId: 'role', operator: 'in', value: ['Admin', 'Manager'] },
  { columnId: 'active', operator: 'equals', value: true },
]);

// Custom filter function
const table = createDataTable({
  // ...
  filtering: {
    enabled: true,
    customFilters: {
      age: (row, filter) => {
        const age = row.data.age;
        return age >= 18 && age <= 65;
      },
    },
  },
});
```

### Custom Sorting

```typescript
const table = createDataTable({
  columns: [
    {
      id: 'date',
      field: 'createdAt',
      label: 'Created',
      comparator: (a, b) => {
        return new Date(a).getTime() - new Date(b).getTime();
      },
    },
  ],
  customComparators: {
    priority: (a, b) => {
      const priorityOrder = { high: 3, medium: 2, low: 1 };
      return priorityOrder[a] - priorityOrder[b];
    },
  },
});
```

### Export Data

```typescript
import { createExportPlugin } from '@awo/datatable';

const exportPlugin = createExportPlugin();
table.use(exportPlugin);

// Export as CSV
exportPlugin.export({
  format: 'csv',
  filename: 'users-export',
  includeHeaders: true,
  selectedOnly: false,
});

// Export selected rows as JSON
exportPlugin.export({
  format: 'json',
  filename: 'selected-users',
  selectedOnly: true,
});
```

### Virtual Scrolling

```typescript
import { createVirtualScrollPlugin } from '@awo/datatable';

const virtualScroll = createVirtualScrollPlugin({
  rowHeight: 48,
  containerHeight: 600,
  overscan: 5,
});

table.use(virtualScroll);

// Attach to container
virtualScroll.attachToContainer(document.querySelector('#table-container'));

// Scroll to specific row
virtualScroll.scrollToRow(100);
```

## 🎯 API Reference

### Core Methods

```typescript
// Data Management
table.setData(data: T[]): void
table.getData(): Row<T>[]
table.addRow(data: T): void
table.updateRow(id: RowId, data: Partial<T>): void
table.deleteRow(id: RowId): void
table.clearData(): void

// Sorting
table.sort(columnId: string, direction?: 'asc' | 'desc' | null): void
table.multiSort(sorts: SortState[]): void
table.clearSort(): void

// Filtering
table.filter(filters: Filter[]): void
table.addFilter(filter: Filter): void
table.removeFilter(columnId: string): void
table.clearFilters(): void
table.search(query: string): void

// Pagination
table.goToPage(page: number): void
table.nextPage(): void
table.previousPage(): void
table.setPageSize(size: number): void

// Selection
table.selectRow(id: RowId): void
table.deselectRow(id: RowId): void
table.toggleRowSelection(id: RowId): void
table.selectAll(): void
table.deselectAll(): void
table.getSelectedRows(): Row<T>[]

// Columns
table.showColumn(columnId: string): void
table.hideColumn(columnId: string): void
table.toggleColumn(columnId: string): void

// State
table.getState(): DataTableState<T>
table.setState(state: Partial<DataTableState<T>>): void
table.resetState(): void

// Server-side
table.reload(): Promise<void>

// Events
table.on(event, callback): () => void
table.off(event, callback): void

// Lifecycle
table.destroy(): void
```

### Events

```typescript
table.on('data:change', (rows) => {});
table.on('sort:change', (sortState) => {});
table.on('filter:change', (filterState) => {});
table.on('page:change', (paginationState) => {});
table.on('selection:change', (selectionState) => {});
table.on('row:select', (row, selected) => {});
table.on('row:click', (row, event) => {});
table.on('search:change', (query) => {});
table.on('state:change', (state) => {});
```

## 📊 Performance

- ✅ Handles **100,000+ rows** with virtual scrolling
- ✅ Optimized sorting and filtering algorithms
- ✅ Efficient state updates (no unnecessary re-renders)
- ✅ Debounced search (configurable)
- ✅ Request cancellation for server-side loading
- ✅ Memory-efficient (only renders visible rows)

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

## 📄 License

MIT © Awo ERP

## 🙏 Credits

Built with ❤️ by the Awo ERP team.
