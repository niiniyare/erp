## 16. 📊 Data Visualization

### 16.1 Visualization Philosophy

Data visualization makes complex information understandable. Good visualizations:

1. **Tell a story** — Communicate insights clearly
2. **Maintain accuracy** — Don't distort data
3. **Stay accessible** — Work for all users
4. **Guide attention** — Highlight what matters
5. **Respect context** — Match user's mental model

### 16.2 Chart Color Palette

```css
:root {
  /* Categorical colors (qualitative) */
  --chart-1: 221.2 83.2% 53.3%;   /* Blue */
  --chart-2: 142.1 76.2% 36.3%;   /* Green */
  --chart-3: 38.7 92.0% 50.2%;    /* Orange */
  --chart-4: 280 80% 55%;         /* Purple */
  --chart-5: 0 84.2% 60.2%;       /* Red */
  --chart-6: 199 89% 48%;         /* Cyan */
  --chart-7: 45 93% 47%;          /* Yellow */
  --chart-8: 330 81% 60%;         /* Pink */
  
  /* Sequential colors (quantitative) */
  --chart-seq-1: 221.2 83.2% 95%;
  --chart-seq-2: 221.2 83.2% 85%;
  --chart-seq-3: 221.2 83.2% 70%;
  --chart-seq-4: 221.2 83.2% 53.3%;
  --chart-seq-5: 221.2 83.2% 40%;
  
  /* Diverging colors */
  --chart-negative: 0 84.2% 60.2%;    /* Red */
  --chart-neutral: 210 40% 96.1%;     /* Gray */
  --chart-positive: 142.1 76.2% 36.3%; /* Green */
}
```

### 16.3 Chart Components

```typescript
// Bar Chart
<BarChart data={data} width={600} height={400}>
  <XAxis dataKey="name" />
  <YAxis />
  <Tooltip />
  <Legend />
  <Bar dataKey="value" fill="hsl(var(--chart-1))" />
</BarChart>

// Line Chart
<LineChart data={data} width={600} height={400}>
  <XAxis dataKey="date" />
  <YAxis />
  <Tooltip />
  <Legend />
  <Line 
    type="monotone" 
    dataKey="revenue" 
    stroke="hsl(var(--chart-1))" 
    strokeWidth={2}
  />
</LineChart>

// Pie Chart
<PieChart width={400} height={400}>
  <Pie
    data={data}
    dataKey="value"
    nameKey="name"
    cx="50%"
    cy="50%"
    outerRadius={120}
    label
  />
  <Tooltip />
</PieChart>
```

### 16.4 Data Table Patterns

```typescript
// Sortable table
interface Column<T> {
  key: keyof T;
  header: string;
  sortable?: boolean;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
}

function DataTable<T>({ data, columns }: DataTableProps<T>) {
  const [sortKey, setSortKey] = React.useState<keyof T | null>(null);
  const [sortOrder, setSortOrder] = React.useState<'asc' | 'desc'>('asc');
  
  const sortedData = React.useMemo(() => {
    if (!sortKey) return data;
    
    return [...data].sort((a, b) => {
      const aVal = a[sortKey];
      const bVal = b[sortKey];
      
      if (aVal < bVal) return sortOrder === 'asc' ? -1 : 1;
      if (aVal > bVal) return sortOrder === 'asc' ? 1 : -1;
      return 0;
    });
  }, [data, sortKey, sortOrder]);
  
  const handleSort = (key: keyof T) => {
    if (sortKey === key) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortOrder('asc');
    }
  };
  
  return (
    <Table>
      <TableHeader>
        <TableRow>
          {columns.map((column) => (
            <TableHead
              key={String(column.key)}
              onClick={() => column.sortable && handleSort(column.key)}
              className={column.sortable ? 'sortable' : ''}
            >
              {column.header}
              {sortKey === column.key && (
                <Icon name={sortOrder === 'asc' ? 'arrow-up' : 'arrow-down'} />
              )}
            </TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {sortedData.map((row, index) => (
          <TableRow key={index}>
            {columns.map((column) => (
              <TableCell key={String(column.key)}>
                {column.render 
                  ? column.render(row[column.key], row)
                  : String(row[column.key])
                }
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
```

### 16.5 Accessibility in Visualizations

```typescript
// Accessible chart
<BarChart data={data} width={600} height={400}>
  <title>Monthly Revenue Chart</title>
  <desc>
    Bar chart showing monthly revenue from January to December.
    Revenue ranges from $10,000 to $50,000.
  </desc>
  
  {/* Visual content */}
  
  {/* Text alternative */}
  <text className="sr-only">
    January: $25,000, February: $30,000, March: $28,000...
  </text>
</BarChart>

// Provide data table alternative
<details>
  <summary>View data as table</summary>
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Month</TableHead>
        <TableHead>Revenue</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {data.map((item) => (
        <TableRow key={item.month}>
          <TableCell>{item.month}</TableCell>
          <TableCell>${item.revenue}</TableCell>
        </TableRow>
      ))}
    </TableBody>
  </Table>
</details>
```

---

