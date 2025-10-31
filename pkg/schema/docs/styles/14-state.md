## 14. 📊 State Management

### 14.1 State Management Philosophy

State should live at the appropriate level:

1. **Local state** — Component-specific (useState, useReducer)
2. **Lifted state** — Shared between siblings (props)
3. **Context state** — Subtree-wide (Context API)
4. **Global state** — App-wide (Redux, Zustand, etc.)

### 14.2 Local State Patterns

```typescript
// Simple state
function Counter() {
  const [count, setCount] = React.useState(0);
  
  return (
    <div>
      <p>Count: {count}</p>
      <Button onClick={() => setCount(count + 1)}>Increment</Button>
    </div>
  );
}

// Complex state with reducer
interface State {
  items: Item[];
  filter: string;
  sortBy: string;
}

type Action = 
  | { type: 'ADD_ITEM'; payload: Item }
  | { type: 'REMOVE_ITEM'; payload: string }
  | { type: 'SET_FILTER'; payload: string }
  | { type: 'SET_SORT'; payload: string };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'ADD_ITEM':
      return { ...state, items: [...state.items, action.payload] };
    case 'REMOVE_ITEM':
      return { 
        ...state, 
        items: state.items.filter(item => item.id !== action.payload) 
      };
    case 'SET_FILTER':
      return { ...state, filter: action.payload };
    case 'SET_SORT':
      return { ...state, sortBy: action.payload };
    default:
      return state;
  }
}

function ItemList() {
  const [state, dispatch] = React.useReducer(reducer, {
    items: [],
    filter: '',
    sortBy: 'name'
  });
  
  // Use state and dispatch
}
```

### 14.3 Context State

```typescript
// Theme context
interface ThemeContextValue {
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;
}

const ThemeContext = React.createContext<ThemeContextValue | null>(null);

function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = React.useState<'light' | 'dark'>('light');
  
  React.useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }, [theme]);
  
  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

function useTheme() {
  const context = React.useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
}

// Usage
function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  
  return (
    <Button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>
      {theme === 'light' ? <Icon name="moon" /> : <Icon name="sun" />}
    </Button>
  );
}
```

### 14.4 Form State

```typescript
// Form state management
interface FormState {
  values: Record<string, any>;
  errors: Record<string, string>;
  touched: Record<string, boolean>;
  isSubmitting: boolean;
}

function useForm<T extends Record<string, any>>(
  initialValues: T,
  validate: (values: T) => Record<string, string>
) {
  const [values, setValues] = React.useState<T>(initialValues);
  const [errors, setErrors] = React.useState<Record<string, string>>({});
  const [touched, setTouched] = React.useState<Record<string, boolean>>({});
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  
  const handleChange = (name: string, value: any) => {
    setValues(prev => ({ ...prev, [name]: value }));
  };
  
  const handleBlur = (name: string) => {
    setTouched(prev => ({ ...prev, [name]: true }));
    const fieldErrors = validate(values);
    setErrors(fieldErrors);
  };
  
  const handleSubmit = async (
    onSubmit: (values: T) => Promise<void>
  ) => {
    setIsSubmitting(true);
    const validationErrors = validate(values);
    
    if (Object.keys(validationErrors).length === 0) {
      try {
        await onSubmit(values);
      } catch (error) {
        console.error(error);
      }
    } else {
      setErrors(validationErrors);
    }
    
    setIsSubmitting(false);
  };
  
  return {
    values,
    errors,
    touched,
    isSubmitting,
    handleChange,
    handleBlur,
    handleSubmit,
  };
}

// Usage
function LoginForm() {
  const form = useForm(
    { email: '', password: '' },
    (values) => {
      const errors: Record<string, string> = {};
      if (!values.email) errors.email = 'Email is required';
      if (!values.password) errors.password = 'Password is required';
      return errors;
    }
  );
  
  return (
    <form onSubmit={(e) => {
      e.preventDefault();
      form.handleSubmit(async (values) => {
        await login(values);
      });
    }}>
      <FormField>
        <FormLabel>Email</FormLabel>
        <Input
          value={form.values.email}
          onChange={(e) => form.handleChange('email', e.target.value)}
          onBlur={() => form.handleBlur('email')}
        />
        {form.touched.email && form.errors.email && (
          <FormMessage>{form.errors.email}</FormMessage>
        )}
      </FormField>
      
      <Button type="submit" disabled={form.isSubmitting}>
        {form.isSubmitting ? 'Logging in...' : 'Login'}
      </Button>
    </form>
  );
}
```

### 14.5 Async State

```typescript
// Async state hook
interface AsyncState<T> {
  data: T | null;
  error: Error | null;
  loading: boolean;
}

function useAsync<T>(
  asyncFunction: () => Promise<T>,
  dependencies: any[] = []
) {
  const [state, setState] = React.useState<AsyncState<T>>({
    data: null,
    error: null,
    loading: true,
  });
  
  React.useEffect(() => {
    let cancelled = false;
    
    setState({ data: null, error: null, loading: true });
    
    asyncFunction()
      .then(data => {
        if (!cancelled) {
          setState({ data, error: null, loading: false });
        }
      })
      .catch(error => {
        if (!cancelled) {
          setState({ data: null, error, loading: false });
        }
      });
    
    return () => {
      cancelled = true;
    };
  }, dependencies);
  
  return state;
}

// Usage
function UserProfile({ userId }: { userId: string }) {
  const { data: user, loading, error } = useAsync(
    () => fetchUser(userId),
    [userId]
  );
  
  if (loading) return <Skeleton />;
  if (error) return <ErrorMessage error={error} />;
  if (!user) return <EmptyState />;
  
  return <div>{user.name}</div>;
}
```

---

