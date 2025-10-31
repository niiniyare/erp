## Section 19: Data Sources

### 🎯 Learning Objectives

By the end of this section, you will understand:
- DataSource configuration
- Static vs dynamic options
- API integration
- Database queries
- Caching strategies

### 19.1 DataSource Struct

```go
type DataSource struct {
    Type    string            `json:"type"`    // "static", "api", "query", "function"
    URL     string            `json:"url,omitempty"`
    Method  string            `json:"method,omitempty"`
    Headers map[string]string `json:"headers,omitempty"`
    Query   string            `json:"query,omitempty"` // SQL query
    Function string           `json:"function,omitempty"` // Registered function name
    Cache    bool             `json:"cache,omitempty"`
    CacheTTL int              `json:"cacheTTL,omitempty"` // seconds
    Transform string          `json:"transform,omitempty"` // JS transform function
}
```

### 19.2 DataSource Loader

```go
// datasource/loader.go

package datasource

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/yourusername/awoerp/internal/schema"
)

type Loader struct {
    db       *sql.DB
    http     *http.Client
    cache    *Cache
    registry *FunctionRegistry
}

func NewLoader(db *sql.DB) *Loader {
    return &Loader{
        db: db,
        http: &http.Client{
            Timeout: 10 * time.Second,
        },
        cache:    NewCache(),
        registry: NewFunctionRegistry(),
    }
}

// Load loads options from data source
func (l *Loader) Load(ctx context.Context, ds *schema.DataSource, params map[string]string) ([]schema.Option, error) {
    // Check cache first
    if ds.Cache {
        cacheKey := l.getCacheKey(ds, params)
        if options, ok := l.cache.Get(cacheKey); ok {
            return options, nil
        }
    }
    
    var options []schema.Option
    var err error
    
    // Load based on type
    switch ds.Type {
    case "api":
        options, err = l.loadFromAPI(ctx, ds, params)
    case "query":
        options, err = l.loadFromQuery(ctx, ds, params)
    case "function":
        options, err = l.loadFromFunction(ctx, ds, params)
    default:
        return nil, fmt.Errorf("unknown data source type: %s", ds.Type)
    }
    
    if err != nil {
        return nil, err
    }
    
    // Cache results
    if ds.Cache {
        cacheKey := l.getCacheKey(ds, params)
        ttl := time.Duration(ds.CacheTTL) * time.Second
        l.cache.Set(cacheKey, options, ttl)
    }
    
    return options, nil
}

// loadFromAPI loads options from HTTP API
func (l *Loader) loadFromAPI(ctx context.Context, ds *schema.DataSource, params map[string]string) ([]schema.Option, error) {
    // Replace URL parameters
    url := ds.URL
    for key, value := range params {
        url = strings.Replace(url, "{"+key+"}", value, -1)
    }
    
    // Create request
    req, err := http.NewRequestWithContext(ctx, ds.Method, url, nil)
    if err != nil {
        return nil, err
    }
    
    // Add headers
    for key, value := range ds.Headers {
        req.Header.Set(key, value)
    }
    
    // Send request
    resp, err := l.http.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }
    
    // Parse response
    var result struct {
        Data []struct {
            Value string `json:"value"`
            Label string `json:"label"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    // Convert to options
    options := make([]schema.Option, len(result.Data))
    for i, item := range result.Data {
        options[i] = schema.Option{
            Value: item.Value,
            Label: item.Label,
        }
    }
    
    return options, nil
}

// loadFromQuery loads options from database query
func (l *Loader) loadFromQuery(ctx context.Context, ds *schema.DataSource, params map[string]string) ([]schema.Option, error) {
    // Replace query parameters
    query := ds.Query
    args := []interface{}{}
    argCount := 1
    
    for key, value := range params {
        placeholder := "{" + key + "}"
        if strings.Contains(query, placeholder) {
            query = strings.Replace(query, placeholder, fmt.Sprintf("$%d", argCount), -1)
            args = append(args, value)
            argCount++
        }
    }
    
    // Execute query
    rows, err := l.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    // Parse results
    var options []schema.Option
    for rows.Next() {
        var value, label string
        if err := rows.Scan(&value, &label); err != nil {
            return nil, err
        }
        options = append(options, schema.Option{
            Value: value,
            Label: label,
        })
    }
    
    return options, nil
}

// loadFromFunction loads options from registered function
func (l *Loader) loadFromFunction(ctx context.Context, ds *schema.DataSource, params map[string]string) ([]schema.Option, error) {
    fn := l.registry.Get(ds.Function)
    if fn == nil {
        return nil, fmt.Errorf("function not found: %s", ds.Function)
    }
    
    return fn(ctx, params)
}

// Cache implementation
type Cache struct {
    mu    sync.RWMutex
    items map[string]*cacheItem
}

type cacheItem struct {
    options   []schema.Option
    expiresAt time.Time
}

func NewCache() *Cache {
    c := &Cache{
        items: make(map[string]*cacheItem),
    }
    
    // Start cleanup goroutine
    go c.cleanup()
    
    return c
}

func (c *Cache) Get(key string) ([]schema.Option, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    item, ok := c.items[key]
    if !ok || time.Now().After(item.expiresAt) {
        return nil, false
    }
    
    return item.options, true
}

func (c *Cache) Set(key string, options []schema.Option, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.items[key] = &cacheItem{
        options:   options,
        expiresAt: time.Now().Add(ttl),
    }
}

func (c *Cache) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, item := range c.items {
            if now.After(item.expiresAt) {
                delete(c.items, key)
            }
        }
        c.mu.Unlock()
    }
}

func (l *Loader) getCacheKey(ds *schema.DataSource, params map[string]string) string {
    return fmt.Sprintf("%s:%s:%v", ds.Type, ds.URL, params)
}

// Function Registry
type FunctionRegistry struct {
    mu        sync.RWMutex
    functions map[string]DataSourceFunc
}

type DataSourceFunc func(ctx context.Context, params map[string]string) ([]schema.Option, error)

func NewFunctionRegistry() *FunctionRegistry {
    return &FunctionRegistry{
        functions: make(map[string]DataSourceFunc),
    }
}

func (r *FunctionRegistry) Register(name string, fn DataSourceFunc) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.functions[name] = fn
}

func (r *FunctionRegistry) Get(name string) DataSourceFunc {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.functions[name]
}
```

### 19.3 Usage Examples

**Example 1: API Data Source**
```json
{
  "name": "country",
  "type": "select",
  "label": "Country",
  "dataSource": {
    "type": "api",
    "url": "https://api.example.com/countries",
    "method": "GET",
    "cache": true,
    "cacheTTL": 3600
  }
}
```

**Example 2: Database Query**
```json
{
  "name": "state",
  "type": "select",
  "label": "State",
  "dependencies": ["country"],
  "dataSource": {
    "type": "query",
    "query": "SELECT code AS value, name AS label FROM states WHERE country_code = {country} ORDER BY name",
    "cache": true,
    "cacheTTL": 1800
  }
}
```

**Example 3: Registered Function**
```go
// Register custom function
loader.registry.Register("get_departments", func(ctx context.Context, params map[string]string) ([]schema.Option, error) {
    // Custom logic
    departments := []schema.Option{
        {Value: "eng", Label: "Engineering"},
        {Value: "sales", Label: "Sales"},
        {Value: "hr", Label: "Human Resources"},
    }
    return departments, nil
})
```

```json
{
  "name": "department",
  "type": "select",
  "label": "Department",
  "dataSource": {
    "type": "function",
    "function": "get_departments"
  }
}
```

---

