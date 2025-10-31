## 13. 🧱 Composition Principles

### 13.1 Composition Philosophy

Composition over inheritance. Build complex UIs by combining simple components rather than creating specialized variants.

```
❌ Bad: Create specific variants
<ButtonWithIcon />
<ButtonWithBadge />
<ButtonWithIconAndBadge />

✅ Good: Compose elements
<Button>
  <Icon />
  <span>Text</span>
  <Badge />
</Button>
```

### 13.2 Slot-Based Composition

```typescript
// Component with named slots
interface CardProps {
  children: React.ReactNode;
}

function Card({ children }: CardProps) {
  const slots = {
    header: null,
    content: null,
    footer: null
  };
  
  React.Children.forEach(children, (child) => {
    if (React.isValidElement(child)) {
      if (child.type === CardHeader) slots.header = child;
      if (child.type === CardContent) slots.content = child;
      if (child.type === CardFooter) slots.footer = child;
    }
  });
  
  return (
    <div className="card">
      {slots.header}
      {slots.content}
      {slots.footer}
    </div>
  );
}

// Usage
<Card>
  <CardHeader>Header content</CardHeader>
  <CardContent>Main content</CardContent>
  <CardFooter>Footer actions</CardFooter>
</Card>
```

### 13.3 Render Props Pattern

```typescript
interface DataListProps<T> {
  data: T[];
  renderItem: (item: T, index: number) => React.ReactNode;
  renderEmpty?: () => React.ReactNode;
  renderLoading?: () => React.ReactNode;
  loading?: boolean;
}

function DataList<T>({ 
  data, 
  renderItem, 
  renderEmpty,
  renderLoading,
  loading 
}: DataListProps<T>) {
  if (loading && renderLoading) {
    return <>{renderLoading()}</>;
  }
  
  if (data.length === 0 && renderEmpty) {
    return <>{renderEmpty()}</>;
  }
  
  return (
    <div className="data-list">
      {data.map((item, index) => renderItem(item, index))}
    </div>
  );
}

// Usage
<DataList
  data={users}
  renderItem={(user) => (
    <UserCard key={user.id} user={user} />
  )}
  renderEmpty={() => <EmptyState />}
  renderLoading={() => <Skeleton />}
  loading={isLoading}
/>
```

### 13.4 Compound Components

```typescript
// Tabs compound component
interface TabsContextValue {
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

const TabsContext = React.createContext<TabsContextValue | null>(null);

function Tabs({ defaultValue, children }: TabsProps) {
  const [activeTab, setActiveTab] = React.useState(defaultValue);
  
  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      <div className="tabs">{children}</div>
    </TabsContext.Provider>
  );
}

function TabsList({ children }: { children: React.ReactNode }) {
  return <div className="tabs-list">{children}</div>;
}

function TabsTrigger({ value, children }: TabsTriggerProps) {
  const context = React.useContext(TabsContext);
  const isActive = context?.activeTab === value;
  
  return (
    <button
      className={`tabs-trigger ${isActive ? 'active' : ''}`}
      onClick={() => context?.setActiveTab(value)}
    >
      {children}
    </button>
  );
}

function TabsContent({ value, children }: TabsContentProps) {
  const context = React.useContext(TabsContext);
  if (context?.activeTab !== value) return null;
  
  return <div className="tabs-content">{children}</div>;
}

// Export as compound
Tabs.List = TabsList;
Tabs.Trigger = TabsTrigger;
Tabs.Content = TabsContent;

// Usage
<Tabs defaultValue="tab1">
  <Tabs.List>
    <Tabs.Trigger value="tab1">Tab 1</Tabs.Trigger>
    <Tabs.Trigger value="tab2">Tab 2</Tabs.Trigger>
  </Tabs.List>
  <Tabs.Content value="tab1">Content 1</Tabs.Content>
  <Tabs.Content value="tab2">Content 2</Tabs.Content>
</Tabs>
```

### 13.5 Polymorphic Components

```typescript
// Component that can render as different elements
type PolymorphicProps<E extends React.ElementType> = {
  as?: E;
  children: React.ReactNode;
} & React.ComponentPropsWithoutRef<E>;

function Box<E extends React.ElementType = 'div'>({
  as,
  children,
  ...props
}: PolymorphicProps<E>) {
  const Component = as || 'div';
  return <Component {...props}>{children}</Component>;
}

// Usage
<Box>Renders as div</Box>
<Box as="section">Renders as section</Box>
<Box as="a" href="/">Renders as link</Box>
<Box as={CustomComponent}>Renders as custom component</Box>
```

### 13.6 Layout Composition

```typescript
// Stack layout
<Stack direction="vertical" spacing={4}>
  <Box>Item 1</Box>
  <Box>Item 2</Box>
  <Box>Item 3</Box>
</Stack>

// Grid layout
<Grid cols={3} gap={4}>
  <Box>1</Box>
  <Box>2</Box>
  <Box>3</Box>
</Grid>

// Flex layout
<Flex justify="between" align="center">
  <Box>Left</Box>
  <Box>Right</Box>
</Flex>
```

### 13.7 Controlled vs Uncontrolled

```typescript
// Controlled component
function ControlledInput() {
  const [value, setValue] = React.useState('');
  
  return (
    <Input 
      value={value}
      onChange={(e) => setValue(e.target.value)}
    />
  );
}

// Uncontrolled component
function UncontrolledInput() {
  const inputRef = React.useRef<HTMLInputElement>(null);
  
  const handleSubmit = () => {
    console.log(inputRef.current?.value);
  };
  
  return <Input ref={inputRef} defaultValue="" />;
}

// Hybrid: Support both patterns
interface InputProps {
  value?: string;
  defaultValue?: string;
  onChange?: (value: string) => void;
}

function Input({ value, defaultValue, onChange }: InputProps) {
  const [internalValue, setInternalValue] = React.useState(defaultValue);
  const isControlled = value !== undefined;
  const currentValue = isControlled ? value : internalValue;
  
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    if (!isControlled) {
      setInternalValue(newValue);
    }
    onChange?.(newValue);
  };
  
  return (
    <input 
      value={currentValue}
      onChange={handleChange}
    />
  );
}
```

---

