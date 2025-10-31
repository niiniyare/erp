## 18. 🏗️ Layout Patterns

### 18.1 Common Layouts

#### **App Shell**

```typescript
<div className="app-shell">
  <Header />
  <div className="app-body">
    <Sidebar />
    <main className="app-main">
      <Outlet />
    </main>
  </div>
</div>

<style>
.app-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.app-body {
  display: flex;
  flex: 1;
}

.app-main {
  flex: 1;
  padding: var(--space-6);
  overflow-y: auto;
}
</style>
```

#### **Dashboard Layout**

```typescript
<div className="dashboard">
  <div className="dashboard-header">
    <h1>Dashboard</h1>
    <div className="dashboard-actions">
      <Button>Export</Button>
      <Button variant="primary">New Report</Button>
    </div>
  </div>
  
  <div className="dashboard-stats">
    <StatCard title="Revenue" value="$45,000" />
    <StatCard title="Users" value="1,234" />
    <StatCard title="Growth" value="+12%" />
  </div>
  
  <div className="dashboard-content">
    <Card>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
      </CardHeader>
      <CardContent>
        {/* Content */}
      </CardContent>
    </Card>
  </div>
</div>
```

#### **Split View**

```css
.split-view {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: var(--space-6);
  height: 100%;
}

.split-view-sidebar {
  border-right: 1px solid var(--border);
  padding: var(--space-6);
  overflow-y: auto;
}

.split-view-main {
  padding: var(--space-6);
  overflow-y: auto;
}

/* Responsive */
@media (max-width: 768px) {
  .split-view {
    grid-template-columns: 1fr;
  }
  
  .split-view-sidebar {
    border-right: none;
    border-bottom: 1px solid var(--border);
  }
}
```

### 18.2 Grid Systems

```css
/* 12-column grid */
.grid-12 {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: var(--space-6);
}

.col-span-1  { grid-column: span 1; }
.col-span-2  { grid-column: span 2; }
.col-span-3  { grid-column: span 3; }
.col-span-4  { grid-column: span 4; }
.col-span-6  { grid-column: span 6; }
.col-span-8  { grid-column: span 8; }
.col-span-12 { grid-column: span 12; }

/* Responsive columns */
@media (max-width: 768px) {
  .col-span-4,
  .col-span-6,
  .col-span-8 {
    grid-column: span 12;
  }
}
```

### 18.3 Sticky Elements

```css
/* Sticky header */
.sticky-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: hsl(var(--background));
  border-bottom: 1px solid hsl(var(--border));
}

/* Sticky sidebar */
.sticky-sidebar {
  position: sticky;
  top: var(--space-6);
  max-height: calc(100vh - var(--space-12));
  overflow-y: auto;
}
```

---

