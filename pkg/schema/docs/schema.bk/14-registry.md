## Section 14: Registry System

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Schema registry architecture
- Storage strategies (memory, file, database)
- Schema versioning
- Loading and caching
- Hot reloading capabilities

### 14.1 Registry Interface

```go
// Registry defines the interface for schema storage and retrieval
type Registry interface {
    // Register adds or updates a schema
    Register(ctx context.Context, schema *Schema) error
    
    // Get retrieves a schema by ID
    Get(ctx context.Context, id string) (*Schema, error)
    
    // List returns all schemas with optional filters
    List(ctx context.Context, filters *Filters) ([]*Schema, error)
    
    // Delete removes a schema
    Delete(ctx context.Context, id string) error
    
    // Exists checks if schema exists
    Exists(ctx context.Context, id string) bool
    
    // GetVersion retrieves specific version of schema
    GetVersion(ctx context.Context, id string, version string) (*Schema, error)
    
    // Search finds schemas by criteria
    Search(ctx context.Context, query string) ([]*Schema, error)
}

// Filters for listing schemas
type Filters struct {
    Type     *Type    `json:"type,omitempty"`
    Category *string  `json:"category,omitempty"`
    Module   *string  `json:"module,omitempty"`
    Tags     []string `json:"tags,omitempty"`
    Limit    int      `json:"limit,omitempty"`
    Offset   int      `json:"offset,omitempty"`
}
```

### 14.2 Memory Registry Implementation

**Best for:** Development, testing, small applications.

```go
// MemoryRegistry stores schemas in memory
type MemoryRegistry struct {
    mu      sync.RWMutex
    schemas map[string]*Schema
    index   map[string][]string // category/module -> schema IDs
}

// NewMemoryRegistry creates a new in-memory registry
func NewMemoryRegistry() *MemoryRegistry {
    return &MemoryRegistry{
        schemas: make(map[string]*Schema),
        index:   make(map[string][]string),
    }
}

// Register adds or updates a schema
func (r *MemoryRegistry) Register(ctx context.Context, schema *Schema) error {
    if err := schema.Validate(); err != nil {
        return fmt.Errorf("schema validation failed: %w", err)
    }
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Store schema
    r.schemas[schema.ID] = schema
    
    // Update indices
    if schema.Category != "" {
        key := "category:" + schema.Category
        r.addToIndex(key, schema.ID)
    }
    if schema.Module != "" {
        key := "module:" + schema.Module
        r.addToIndex(key, schema.ID)
    }
    for _, tag := range schema.Tags {
        key := "tag:" + tag
        r.addToIndex(key, schema.ID)
    }
    
    return nil
}

// Get retrieves a schema by ID
func (r *MemoryRegistry) Get(ctx context.Context, id string) (*Schema, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    schema, exists := r.schemas[id]
    if !exists {
        return nil, fmt.Errorf("schema not found: %s", id)
    }
    
    // Return a copy to prevent external modifications
    return schema.Clone(), nil
}

// List returns all schemas with optional filters
func (r *MemoryRegistry) List(ctx context.Context, filters *Filters) ([]*Schema, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var schemas []*Schema
    
    if filters == nil {
        // Return all schemas
        for _, schema := range r.schemas {
            schemas = append(schemas, schema.Clone())
        }
    } else {
        // Apply filters
        for _, schema := range r.schemas {
            if r.matchesFilters(schema, filters) {
                schemas = append(schemas, schema.Clone())
            }
        }
    }
    
    // Apply pagination
    if filters != nil {
        start := filters.Offset
        end := start + filters.Limit
        if start < len(schemas) {
            if end > len(schemas) {
                end = len(schemas)
            }
            schemas = schemas[start:end]
        } else {
            schemas = nil
        }
    }
    
    return schemas, nil
}

// Delete removes a schema
func (r *MemoryRegistry) Delete(ctx context.Context, id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    schema, exists := r.schemas[id]
    if !exists {
        return fmt.Errorf("schema not found: %s", id)
    }
    
    // Remove from indices
    if schema.Category != "" {
        r.removeFromIndex("category:"+schema.Category, id)
    }
    if schema.Module != "" {
        r.removeFromIndex("module:"+schema.Module, id)
    }
    for _, tag := range schema.Tags {
        r.removeFromIndex("tag:"+tag, id)
    }
    
    // Remove schema
    delete(r.schemas, id)
    
    return nil
}

// Exists checks if schema exists
func (r *MemoryRegistry) Exists(ctx context.Context, id string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    _, exists := r.schemas[id]
    return exists
}

// Helper methods
func (r *MemoryRegistry) addToIndex(key, id string) {
    if r.index[key] == nil {
        r.index[key] = []string{}
    }
    
    // Check if already exists
    for _, existingID := range r.index[key] {
        if existingID == id {
            return
        }
    }
    
    r.index[key] = append(r.index[key], id)
}

func (r *MemoryRegistry) removeFromIndex(key, id string) {
    ids := r.index[key]
    for i, existingID := range ids {
        if existingID == id {
            r.index[key] = append(ids[:i], ids[i+1:]...)
            break
        }
    }
}

func (r *MemoryRegistry) matchesFilters(schema *Schema, filters *Filters) bool {
    if filters.Type != nil && schema.Type != *filters.Type {
        return false
    }
    if filters.Category != nil && schema.Category != *filters.Category {
        return false
    }
    if filters.Module != nil && schema.Module != *filters.Module {
        return false
    }
    if len(filters.Tags) > 0 {
        hasTag := false
        for _, filterTag := range filters.Tags {
            for _, schemaTag := range schema.Tags {
                if schemaTag == filterTag {
                    hasTag = true
                    break
                }
            }
            if hasTag {
                break
            }
        }
        if !hasTag {
            return false
        }
    }
    return true
}
```

### 14.3 File Registry Implementation

**Best for:** Production with version control, GitOps workflows.

```go
// FileRegistry stores schemas as JSON files
type FileRegistry struct {
    baseDir string
    cache   *MemoryRegistry // In-memory cache
    watcher *fsnotify.Watcher // For hot reloading
}

// NewFileRegistry creates a file-based registry
func NewFileRegistry(baseDir string) (*FileRegistry, error) {
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create base directory: %w", err)
    }
    
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, fmt.Errorf("failed to create watcher: %w", err)
    }
    
    registry := &FileRegistry{
        baseDir: baseDir,
        cache:   NewMemoryRegistry(),
        watcher: watcher,
    }
    
    // Load existing schemas
    if err := registry.loadAll(); err != nil {
        return nil, fmt.Errorf("failed to load schemas: %w", err)
    }
    
    // Start watching for changes
    if err := watcher.Add(baseDir); err != nil {
        return nil, fmt.Errorf("failed to watch directory: %w", err)
    }
    
    go registry.watchFiles()
    
    return registry, nil
}

// Register adds or updates a schema
func (r *FileRegistry) Register(ctx context.Context, schema *Schema) error {
    if err := schema.Validate(); err != nil {
        return fmt.Errorf("schema validation failed: %w", err)
    }
    
    // Write to file
    filePath := r.getFilePath(schema.ID)
    data, err := json.MarshalIndent(schema, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal schema: %w", err)
    }
    
    if err := os.WriteFile(filePath, data, 0644); err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }
    
    // Update cache
    return r.cache.Register(ctx, schema)
}

// Get retrieves a schema by ID
func (r *FileRegistry) Get(ctx context.Context, id string) (*Schema, error) {
    // Try cache first
    schema, err := r.cache.Get(ctx, id)
    if err == nil {
        return schema, nil
    }
    
    // Load from file if not in cache
    filePath := r.getFilePath(id)
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("schema not found: %s", id)
    }
    
    schema = &Schema{}
    if err := json.Unmarshal(data, schema); err != nil {
        return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
    }
    
    // Add to cache
    r.cache.Register(ctx, schema)
    
    return schema, nil
}

// List returns all schemas with optional filters
func (r *FileRegistry) List(ctx context.Context, filters *Filters) ([]*Schema, error) {
    return r.cache.List(ctx, filters)
}

// Delete removes a schema
func (r *FileRegistry) Delete(ctx context.Context, id string) error {
    // Delete file
    filePath := r.getFilePath(id)
    if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to delete file: %w", err)
    }
    
    // Remove from cache
    return r.cache.Delete(ctx, id)
}

// Exists checks if schema exists
func (r *FileRegistry) Exists(ctx context.Context, id string) bool {
    return r.cache.Exists(ctx, id)
}

// Helper methods
func (r *FileRegistry) getFilePath(id string) string {
    return filepath.Join(r.baseDir, id+".json")
}

func (r *FileRegistry) loadAll() error {
    entries, err := os.ReadDir(r.baseDir)
    if err != nil {
        return err
    }
    
    ctx := context.Background()
    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
            continue
        }
        
        filePath := filepath.Join(r.baseDir, entry.Name())
        data, err := os.ReadFile(filePath)
        if err != nil {
            continue // Skip files that can't be read
        }
        
        schema := &Schema{}
        if err := json.Unmarshal(data, schema); err != nil {
            continue // Skip invalid schemas
        }
        
        r.cache.Register(ctx, schema)
    }
    
    return nil
}

// watchFiles watches for file changes and reloads schemas
func (r *FileRegistry) watchFiles() {
    ctx := context.Background()
    
    for {
        select {
        case event, ok := <-r.watcher.Events:
            if !ok {
                return
            }
            
            // Reload changed schema
            if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
                data, err := os.ReadFile(event.Name)
                if err != nil {
                    continue
                }
                
                schema := &Schema{}
                if err := json.Unmarshal(data, schema); err != nil {
                    continue
                }
                
                r.cache.Register(ctx, schema)
                log.Printf("Reloaded schema: %s", schema.ID)
            }
            
            // Remove deleted schema
            if event.Op&fsnotify.Remove == fsnotify.Remove {
                id := strings.TrimSuffix(filepath.Base(event.Name), ".json")
                r.cache.Delete(ctx, id)
                log.Printf("Removed schema: %s", id)
            }
            
        case err, ok := <-r.watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Watcher error: %v", err)
        }
    }
}

// Close stops the file watcher
func (r *FileRegistry) Close() error {
    return r.watcher.Close()
}
```

### 14.4 Database Registry Implementation

**Best for:** Production, multi-instance deployments, schema versioning.

```go
// DBRegistry stores schemas in PostgreSQL
type DBRegistry struct {
    db    *sql.DB
    cache *MemoryRegistry
}

// NewDBRegistry creates a database-backed registry
func NewDBRegistry(db *sql.DB) *DBRegistry {
    return &DBRegistry{
        db:    db,
        cache: NewMemoryRegistry(),
    }
}

// Register adds or updates a schema
func (r *DBRegistry) Register(ctx context.Context, schema *Schema) error {
    if err := schema.Validate(); err != nil {
        return fmt.Errorf("schema validation failed: %w", err)
    }
    
    // Serialize schema
    data, err := json.Marshal(schema)
    if err != nil {
        return fmt.Errorf("failed to marshal schema: %w", err)
    }
    
    // Insert or update
    query := `
        INSERT INTO schemas (id, type, version, title, data, category, module, tags, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
        ON CONFLICT (id) DO UPDATE SET
            type = EXCLUDED.type,
            version = EXCLUDED.version,
            title = EXCLUDED.title,
            data = EXCLUDED.data,
            category = EXCLUDED.category,
            module = EXCLUDED.module,
            tags = EXCLUDED.tags,
            updated_at = NOW()
    `
    
    _, err = r.db.ExecContext(ctx, query,
        schema.ID,
        schema.Type,
        schema.Version,
        schema.Title,
        data,
        schema.Category,
        schema.Module,
        pq.Array(schema.Tags),
    )
    
    if err != nil {
        return fmt.Errorf("failed to save schema: %w", err)
    }
    
    // Update cache
    return r.cache.Register(ctx, schema)
}

// Get retrieves a schema by ID
func (r *DBRegistry) Get(ctx context.Context, id string) (*Schema, error) {
    // Try cache first
    schema, err := r.cache.Get(ctx, id)
    if err == nil {
        return schema, nil
    }
    
    // Load from database
    query := `SELECT data FROM schemas WHERE id = $1`
    
    var data []byte
    err = r.db.QueryRowContext(ctx, query, id).Scan(&data)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("schema not found: %s", id)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to query schema: %w", err)
    }
    
    schema = &Schema{}
    if err := json.Unmarshal(data, schema); err != nil {
        return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
    }
    
    // Add to cache
    r.cache.Register(ctx, schema)
    
    return schema, nil
}

// GetVersion retrieves specific version of schema
func (r *DBRegistry) GetVersion(ctx context.Context, id string, version string) (*Schema, error) {
    query := `
        SELECT data 
        FROM schema_versions 
        WHERE schema_id = $1 AND version = $2
    `
    
    var data []byte
    err := r.db.QueryRowContext(ctx, query, id, version).Scan(&data)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("schema version not found: %s@%s", id, version)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to query schema version: %w", err)
    }
    
    schema := &Schema{}
    if err := json.Unmarshal(data, schema); err != nil {
        return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
    }
    
    return schema, nil
}

// List returns all schemas with optional filters
func (r *DBRegistry) List(ctx context.Context, filters *Filters) ([]*Schema, error) {
    query := `SELECT data FROM schemas WHERE 1=1`
    args := []interface{}{}
    argCount := 1
    
    if filters != nil {
        if filters.Type != nil {
            query += fmt.Sprintf(" AND type = $%d", argCount)
            args = append(args, *filters.Type)
            argCount++
        }
        if filters.Category != nil {
            query += fmt.Sprintf(" AND category = $%d", argCount)
            args = append(args, *filters.Category)
            argCount++
        }
        if filters.Module != nil {
            query += fmt.Sprintf(" AND module = $%d", argCount)
            args = append(args, *filters.Module)
            argCount++
        }
        if len(filters.Tags) > 0 {
            query += fmt.Sprintf(" AND tags && $%d", argCount)
            args = append(args, pq.Array(filters.Tags))
            argCount++
        }
        
        if filters.Limit > 0 {
            query += fmt.Sprintf(" LIMIT $%d", argCount)
            args = append(args, filters.Limit)
            argCount++
        }
        if filters.Offset > 0 {
            query += fmt.Sprintf(" OFFSET $%d", argCount)
            args = append(args, filters.Offset)
            argCount++
        }
    }
    
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query schemas: %w", err)
    }
    defer rows.Close()
    
    var schemas []*Schema
    for rows.Next() {
        var data []byte
        if err := rows.Scan(&data); err != nil {
            continue
        }
        
        schema := &Schema{}
        if err := json.Unmarshal(data, schema); err != nil {
            continue
        }
        
        schemas = append(schemas, schema)
    }
    
    return schemas, nil
}

// Delete removes a schema
func (r *DBRegistry) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM schemas WHERE id = $1`
    
    _, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete schema: %w", err)
    }
    
    // Remove from cache
    return r.cache.Delete(ctx, id)
}

// Exists checks if schema exists
func (r *DBRegistry) Exists(ctx context.Context, id string) bool {
    if r.cache.Exists(ctx, id) {
        return true
    }
    
    query := `SELECT 1 FROM schemas WHERE id = $1`
    
    var exists int
    err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
    return err == nil
}

// Search finds schemas by criteria
func (r *DBRegistry) Search(ctx context.Context, query string) ([]*Schema, error) {
    sqlQuery := `
        SELECT data 
        FROM schemas 
        WHERE 
            title ILIKE $1 OR
            data::text ILIKE $1
        LIMIT 50
    `
    
    searchPattern := "%" + query + "%"
    
    rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern)
    if err != nil {
        return nil, fmt.Errorf("failed to search schemas: %w", err)
    }
    defer rows.Close()
    
    var schemas []*Schema
    for rows.Next() {
        var data []byte
        if err := rows.Scan(&data); err != nil {
            continue
        }
        
        schema := &Schema{}
        if err := json.Unmarshal(data, schema); err != nil {
            continue
        }
        
        schemas = append(schemas, schema)
    }
    
    return schemas, nil
}
```

**Database schema:**
```sql
CREATE TABLE schemas (
    id VARCHAR(100) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    version VARCHAR(20),
    title VARCHAR(200) NOT NULL,
    data JSONB NOT NULL,
    category VARCHAR(50),
    module VARCHAR(50),
    tags TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indices for fast lookup
CREATE INDEX idx_schemas_type ON schemas(type);
CREATE INDEX idx_schemas_category ON schemas(category);
CREATE INDEX idx_schemas_module ON schemas(module);
CREATE INDEX idx_schemas_tags ON schemas USING GIN(tags);

-- Full-text search index
CREATE INDEX idx_schemas_search ON schemas USING GIN(to_tsvector('english', title || ' ' || data::text));

-- Version history table
CREATE TABLE schema_versions (
    id SERIAL PRIMARY KEY,
    schema_id VARCHAR(100) NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,
    version VARCHAR(20) NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(100),
    UNIQUE(schema_id, version)
);

CREATE INDEX idx_schema_versions_schema_id ON schema_versions(schema_id);
```

### 14.5 Registry Usage Examples

#### Example 1: Register Schema

```go
func main() {
    // Create registry
    registry := NewMemoryRegistry()
    ctx := context.Background()
    
    // Create schema
    schema := schema.NewSchema("user-form", schema.TypeForm, "User Registration")
    schema.AddField(schema.Field{
        Name:     "email",
        Type:     schema.FieldEmail,
        Label:    "Email",
        Required: true,
    })
    
    // Register
    if err := registry.Register(ctx, schema); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Schema registered successfully")
}
```

#### Example 2: Load and Use Schema

```go
func HandleForm(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Load schema
    schema, err := registry.Get(ctx, "user-form")
    if err != nil {
        http.Error(w, "Schema not found", 404)
        return
    }
    
    // Enrich with user context
    user := GetUserFromContext(ctx)
    enriched, err := enricher.Enrich(ctx, schema, user)
    if err != nil {
        http.Error(w, "Enrichment failed", 500)
        return
    }
    
    // Render form
    views.FormRenderer(enriched, nil).Render(ctx, w)
}
```

#### Example 3: List Schemas by Category

```go
func ListUserSchemas(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    category := "users"
    filters := &schema.Filters{
        Category: &category,
        Limit:    20,
    }
    
    schemas, err := registry.List(ctx, filters)
    if err != nil {
        http.Error(w, "Failed to list schemas", 500)
        return
    }
    
    views.SchemaList(schemas).Render(ctx, w)
}
```

#### Example 4: Hot Reload with File Registry

```go
func main() {
    // Create file registry (with hot reload)
    registry, err := NewFileRegistry("./schemas")
    if err != nil {
        log.Fatal(err)
    }
    defer registry.Close()
    
    // Start server
    http.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
        // This will always use the latest schema from disk
        schema, _ := registry.Get(r.Context(), "user-form")
        views.FormRenderer(schema, nil).Render(r.Context(), w)
    })
    
    log.Println("Server started with hot reload enabled")
    http.ListenAndServe(":8080", nil)
}
```

---

