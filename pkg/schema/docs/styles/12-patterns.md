## 12. 🔄 Component Patterns

### 12.1 Pattern Philosophy

Component patterns are proven solutions to common UI problems. They provide:

1. **Consistency** — Same problems solved the same way
2. **Efficiency** — Don't reinvent the wheel
3. **Usability** — Familiar patterns users understand
4. **Accessibility** — Built-in a11y best practices

### 12.2 Data Display Patterns

#### **Card Pattern**

```typescript
// Basic card structure
<Card>
  <CardHeader>
    <CardTitle>Card Title</CardTitle>
    <CardDescription>Supporting description text</CardDescription>
  </CardHeader>
  <CardContent>
    Main content goes here
  </CardContent>
  <CardFooter>
    <Button>Action</Button>
  </CardFooter>
</Card>

// Card variations
<Card variant="outline">      <!-- Outlined card -->
<Card variant="filled">       <!-- Filled background -->
<Card variant="elevated">     <!-- With shadow -->
<Card interactive>            <!-- Hoverable/clickable -->
```

**CSS Implementation:**

```css
.card {
  border-radius: var(--radius-lg);
  background: hsl(var(--card));
  color: hsl(var(--card-foreground));
  box-shadow: var(--shadow-sm);
}

.card-outline {
  border: 1px solid hsl(var(--border));
}

.card-interactive {
  cursor: pointer;
  transition: box-shadow var(--duration-normal);
}

.card-interactive:hover {
  box-shadow: var(--shadow-md);
}
```

#### **List Pattern**

```typescript
// Simple list
<List>
  <ListItem>Item 1</ListItem>
  <ListItem>Item 2</ListItem>
  <ListItem>Item 3</ListItem>
</List>

// Interactive list
<List>
  <ListItem interactive onClick={handler}>
    <ListItemIcon><Icon name="user" /></ListItemIcon>
    <ListItemText>
      <ListItemTitle>John Doe</ListItemTitle>
      <ListItemDescription>john@example.com</ListItemDescription>
    </ListItemText>
    <ListItemAction><Icon name="chevron-right" /></ListItemAction>
  </ListItem>
</List>
```

#### **Table Pattern**

```typescript
<Table>
  <TableHeader>
    <TableRow>
      <TableHead>Name</TableHead>
      <TableHead>Email</TableHead>
      <TableHead>Role</TableHead>
      <TableHead>Actions</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    <TableRow>
      <TableCell>John Doe</TableCell>
      <TableCell>john@example.com</TableCell>
      <TableCell>Admin</TableCell>
      <TableCell>
        <Button size="sm" variant="ghost">Edit</Button>
      </TableCell>
    </TableRow>
  </TableBody>
</Table>
```

### 12.3 Navigation Patterns

#### **Sidebar Navigation**

```typescript
<Sidebar>
  <SidebarHeader>
    <Logo />
  </SidebarHeader>
  <SidebarContent>
    <SidebarGroup>
      <SidebarGroupLabel>Main</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarItem href="/dashboard" active>
          <Icon name="home" />
          Dashboard
        </SidebarItem>
        <SidebarItem href="/projects">
          <Icon name="folder" />
          Projects
        </SidebarItem>
      </SidebarGroupContent>
    </SidebarGroup>
  </SidebarContent>
  <SidebarFooter>
    <UserMenu />
  </SidebarFooter>
</Sidebar>
```

#### **Tabs Pattern**

```typescript
<Tabs defaultValue="account">
  <TabsList>
    <TabsTrigger value="account">Account</TabsTrigger>
    <TabsTrigger value="password">Password</TabsTrigger>
    <TabsTrigger value="notifications">Notifications</TabsTrigger>
  </TabsList>
  <TabsContent value="account">
    Account settings content
  </TabsContent>
  <TabsContent value="password">
    Password settings content
  </TabsContent>
  <TabsContent value="notifications">
    Notification settings content
  </TabsContent>
</Tabs>
```

#### **Breadcrumb Pattern**

```typescript
<Breadcrumb>
  <BreadcrumbList>
    <BreadcrumbItem>
      <BreadcrumbLink href="/">Home</BreadcrumbLink>
    </BreadcrumbItem>
    <BreadcrumbSeparator />
    <BreadcrumbItem>
      <BreadcrumbLink href="/products">Products</BreadcrumbLink>
    </BreadcrumbItem>
    <BreadcrumbSeparator />
    <BreadcrumbItem>
      <BreadcrumbPage>Current Product</BreadcrumbPage>
    </BreadcrumbItem>
  </BreadcrumbList>
</Breadcrumb>
```

### 12.4 Feedback Patterns

#### **Toast/Notification Pattern**

```typescript
// Toast notification
toast({
  title: "Success!",
  description: "Your changes have been saved.",
  variant: "success",
  duration: 3000,
});

// Different variants
toast({ variant: "success" });    // Green checkmark
toast({ variant: "error" });      // Red error
toast({ variant: "warning" });    // Yellow warning
toast({ variant: "info" });       // Blue info
```

#### **Alert Pattern**

```typescript
<Alert variant="success">
  <AlertIcon><Icon name="check-circle" /></AlertIcon>
  <AlertTitle>Success</AlertTitle>
  <AlertDescription>
    Your profile has been updated successfully.
  </AlertDescription>
</Alert>

<Alert variant="destructive">
  <AlertIcon><Icon name="alert-circle" /></AlertIcon>
  <AlertTitle>Error</AlertTitle>
  <AlertDescription>
    Unable to save changes. Please try again.
  </AlertDescription>
</Alert>
```

#### **Progress Pattern**

```typescript
// Linear progress
<Progress value={60} max={100} />

// Circular progress
<CircularProgress value={60} />

// Indeterminate progress
<Progress indeterminate />

// With label
<Progress value={60} showLabel />
```

### 12.5 Overlay Patterns

#### **Modal/Dialog Pattern**

```typescript
<Dialog>
  <DialogTrigger asChild>
    <Button>Open Dialog</Button>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Confirm Action</DialogTitle>
      <DialogDescription>
        Are you sure you want to proceed?
      </DialogDescription>
    </DialogHeader>
    <DialogBody>
      <!-- Main content -->
    </DialogBody>
    <DialogFooter>
      <Button variant="outline">Cancel</Button>
      <Button>Confirm</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

#### **Dropdown Pattern**

```typescript
<DropdownMenu>
  <DropdownMenuTrigger asChild>
    <Button variant="outline">
      Options
      <Icon name="chevron-down" />
    </Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent>
    <DropdownMenuLabel>My Account</DropdownMenuLabel>
    <DropdownMenuSeparator />
    <DropdownMenuItem>
      <Icon name="user" />
      Profile
    </DropdownMenuItem>
    <DropdownMenuItem>
      <Icon name="settings" />
      Settings
    </DropdownMenuItem>
    <DropdownMenuSeparator />
    <DropdownMenuItem destructive>
      <Icon name="log-out" />
      Logout
    </DropdownMenuItem>
  </DropdownMenuContent>
</DropdownMenu>
```

#### **Popover Pattern**

```typescript
<Popover>
  <PopoverTrigger asChild>
    <Button variant="ghost">
      <Icon name="info" />
    </Button>
  </PopoverTrigger>
  <PopoverContent>
    <PopoverHeader>Information</PopoverHeader>
    <PopoverBody>
      Additional context or help text goes here.
    </PopoverBody>
  </PopoverContent>
</Popover>
```

### 12.6 Input Patterns

#### **Form Field Pattern**

```typescript
<FormField>
  <FormLabel htmlFor="email">
    Email
    <FormRequired>*</FormRequired>
  </FormLabel>
  <FormControl>
    <Input 
      id="email"
      type="email"
      placeholder="you@example.com"
    />
  </FormControl>
  <FormDescription>
    We'll never share your email.
  </FormDescription>
  <FormMessage>
    <!-- Error message appears here -->
  </FormMessage>
</FormField>
```

#### **Search Pattern**

```typescript
<SearchBox>
  <SearchInput 
    placeholder="Search..." 
    value={query}
    onChange={handleSearch}
  />
  <SearchButton>
    <Icon name="search" />
  </SearchButton>
  {query && (
    <SearchClear onClick={clearSearch}>
      <Icon name="x" />
    </SearchClear>
  )}
</SearchBox>

<!-- With results -->
<SearchResults>
  <SearchResultItem>Result 1</SearchResultItem>
  <SearchResultItem>Result 2</SearchResultItem>
</SearchResults>
```

#### **Date Picker Pattern**

```typescript
<DatePicker>
  <DatePickerTrigger>
    <Input 
      value={date ? formatDate(date) : ""}
      placeholder="Select date"
      readOnly
    />
    <Icon name="calendar" />
  </DatePickerTrigger>
  <DatePickerContent>
    <Calendar 
      selected={date}
      onSelect={setDate}
    />
  </DatePickerContent>
</DatePicker>
```

### 12.7 Empty States

```typescript
<EmptyState>
  <EmptyStateIcon>
    <Icon name="inbox" size="xl" />
  </EmptyStateIcon>
  <EmptyStateTitle>No messages yet</EmptyStateTitle>
  <EmptyStateDescription>
    When you receive messages, they'll appear here.
  </EmptyStateDescription>
  <EmptyStateAction>
    <Button>Send your first message</Button>
  </EmptyStateAction>
</EmptyState>
```

### 12.8 Loading States

```typescript
// Skeleton loading
<Card>
  <CardHeader>
    <Skeleton className="h-4 w-[250px]" />
    <Skeleton className="h-3 w-[200px]" />
  </CardHeader>
  <CardContent>
    <Skeleton className="h-[200px]" />
  </CardContent>
</Card>

// Spinner loading
<div className="loading-container">
  <Spinner size="lg" />
  <p>Loading content...</p>
</div>

// Progress loading
<ProgressBar value={progress} />
```

---

