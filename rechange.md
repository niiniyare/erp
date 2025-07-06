# Complete Directory Structure Guide with Code Interconnections

## 📁 Root Level Directories

### `/cmd/` - Application Entry Points
**Purpose**: Contains the main applications that can be built from this project.

**Files & Structure**:
```
cmd/
├── server/
│   └── main.go              # HTTP/gRPC server
├── worker/
│   └── main.go              # Temporal worker
└── migrate/
    └── main.go              # Database migration tool
```

**Example: `cmd/server/main.go`**
```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/niiniyare/erp/internal/platform/config"
    "github.com/niiniyare/erp/internal/platform/database"
    "github.com/niiniyare/erp/internal/api/handlers"
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/platform/cache"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize database
    db, err := database.NewConnection(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Initialize cache
    redisClient := cache.NewRedisClient(cfg.Redis)
    
    // Initialize repositories
    tenantRepo := tenant.NewRepository(db)
    
    // Initialize services
    tenantService := tenant.NewService(tenantRepo, redisClient)
    
    // Initialize API handlers
    apiHandlers := handlers.New(tenantService)
    
    // Start server
    log.Printf("Server starting on port %s", cfg.Server.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, apiHandlers.Router()))
}
```

**Example: `cmd/worker/main.go`**
```go
package main

import (
    "log"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    
    "github.com/niiniyare/erp/internal/workflows/tenant"
    "github.com/niiniyare/erp/internal/platform/config"
)

func main() {
    cfg := config.Load()
    
    // Create Temporal client
    c, err := client.Dial(client.Options{
        HostPort: cfg.Temporal.HostPort,
    })
    if err != nil {
        log.Fatalln("Unable to create client", err)
    }
    defer c.Close()
    
    // Create worker
    w := worker.New(c, "erp-task-queue", worker.Options{})
    
    // Register workflows and activities
    tenant.RegisterWorkflows(w)
    tenant.RegisterActivities(w)
    
    // Start worker
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Unable to start worker", err)
    }
}
```

**Interconnection**: These main files are the orchestrators that wire together all the components from `internal/` packages.

---

## 📁 `/internal/` - Private Application Code

### `/internal/core/` - Business Domain Logic

**Purpose**: Contains the core business domains. Each subdirectory represents a bounded context in Domain-Driven Design.

#### `/internal/core/tenant/`

**Files**:
```
tenant/
├── model.go              # Domain entities and value objects
├── repository.go         # Data access interface
├── service.go           # Business logic
├── workflow.go          # Temporal workflows
└── errors.go            # Domain-specific errors
```

**Example: `model.go`**
```go
package tenant

import (
    "time"
    "github.com/google/uuid"
)

// Tenant represents a tenant in the system
type Tenant struct {
    ID        uuid.UUID          `json:"id"`
    Name      string             `json:"name"`
    Subdomain string             `json:"subdomain"`
    PlanType  PlanType           `json:"plan_type"`
    Status    Status             `json:"status"`
    Settings  map[string]interface{} `json:"settings"`
    CreatedAt time.Time          `json:"created_at"`
    UpdatedAt time.Time          `json:"updated_at"`
}

// PlanType represents subscription plans
type PlanType string

const (
    PlanTypeBasic      PlanType = "basic"
    PlanTypeProfessional PlanType = "professional"
    PlanTypeEnterprise PlanType = "enterprise"
)

// Status represents tenant status
type Status string

const (
    StatusActive    Status = "active"
    StatusSuspended Status = "suspended"
    StatusTrialI    Status = "trial"
)

// CreateTenantRequest represents tenant creation request
type CreateTenantRequest struct {
    Name      string             `json:"name" validate:"required,min=2,max=100"`
    Subdomain string             `json:"subdomain" validate:"required,min=3,max=50,alphanum"`
    PlanType  PlanType           `json:"plan_type"`
    Settings  map[string]interface{} `json:"settings,omitempty"`
}

// UpdateTenantRequest represents tenant update request  
type UpdateTenantRequest struct {
    Name     *string            `json:"name,omitempty"`
    PlanType *PlanType          `json:"plan_type,omitempty"`
    Status   *Status            `json:"status,omitempty"`
    Settings map[string]interface{} `json:"settings,omitempty"`
}
```

**Example: `repository.go`**
```go
package tenant

import (
    "context"
    "github.com/google/uuid"
)

// Repository defines the interface for tenant data access
type Repository interface {
    Create(ctx context.Context, tenant *Tenant) error
    GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
    GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
    Update(ctx context.Context, id uuid.UUID, updates UpdateTenantRequest) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, offset, limit int) ([]*Tenant, error)
    Exists(ctx context.Context, subdomain string) (bool, error)
}

// repository implements Repository interface
type repository struct {
    db Database // This would be your database interface
}

// NewRepository creates a new tenant repository
func NewRepository(db Database) Repository {
    return &repository{db: db}
}

// Create implements Repository.Create
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    query := `
        INSERT INTO tenants (id, name, subdomain, plan_type, status, settings, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
    `
    
    return r.db.ExecContext(ctx, query, 
        tenant.ID, tenant.Name, tenant.Subdomain, 
        tenant.PlanType, tenant.Status, tenant.Settings)
}

// GetByID implements Repository.GetByID
func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    query := `
        SELECT id, name, subdomain, plan_type, status, settings, created_at, updated_at
        FROM tenants 
        WHERE id = $1
    `
    
    tenant := &Tenant{}
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &tenant.ID, &tenant.Name, &tenant.Subdomain,
        &tenant.PlanType, &tenant.Status, &tenant.Settings,
        &tenant.CreatedAt, &tenant.UpdatedAt,
    )
    
    if err != nil {
        return nil, err
    }
    
    return tenant, nil
}
```

**Example: `service.go`**
```go
package tenant

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "github.com/niiniyare/erp/internal/platform/cache"
    "github.com/niiniyare/erp/internal/shared/errors"
)

// Service defines tenant business logic interface
type Service interface {
    CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
    GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error)
    GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
    UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) error
    DeactivateTenant(ctx context.Context, id uuid.UUID) error
    ListTenants(ctx context.Context, offset, limit int) ([]*Tenant, error)
}

// service implements Service interface
type service struct {
    repo  Repository
    cache cache.Service
}

// NewService creates a new tenant service
func NewService(repo Repository, cache cache.Service) Service {
    return &service{
        repo:  repo,
        cache: cache,
    }
}

// CreateTenant implements Service.CreateTenant
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    // Validate subdomain uniqueness
    exists, err := s.repo.Exists(ctx, req.Subdomain)
    if err != nil {
        return nil, fmt.Errorf("failed to check subdomain existence: %w", err)
    }
    if exists {
        return nil, errors.ErrSubdomainAlreadyExists
    }
    
    // Create tenant entity
    tenant := &Tenant{
        ID:        uuid.New(),
        Name:      req.Name,
        Subdomain: req.Subdomain,
        PlanType:  req.PlanType,
        Status:    StatusActive,
        Settings:  req.Settings,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    // Save to database
    if err := s.repo.Create(ctx, tenant); err != nil {
        return nil, fmt.Errorf("failed to create tenant: %w", err)
    }
    
    // Cache the tenant
    cacheKey := fmt.Sprintf("tenant:subdomain:%s", tenant.Subdomain)
    s.cache.Set(ctx, cacheKey, tenant, 30*time.Minute)
    
    return tenant, nil
}

// GetTenantBySubdomain implements Service.GetTenantBySubdomain with caching
func (s *service) GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("tenant:subdomain:%s", subdomain)
    if cached := s.cache.Get(ctx, cacheKey); cached != nil {
        if tenant, ok := cached.(*Tenant); ok {
            return tenant, nil
        }
    }
    
    // Get from database
    tenant, err := s.repo.GetBySubdomain(ctx, subdomain)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    s.cache.Set(ctx, cacheKey, tenant, 30*time.Minute)
    
    return tenant, nil
}
```

**Example: `workflow.go`**
```go
package tenant

import (
    "time"
    
    "go.temporal.io/sdk/workflow"
    "go.temporal.io/sdk/activity"
)

// TenantOnboardingWorkflow handles tenant setup process
func TenantOnboardingWorkflow(ctx workflow.Context, req CreateTenantRequest) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // Step 1: Create tenant record
    var tenantID string
    err := workflow.ExecuteActivity(ctx, CreateTenantActivity, req).Get(ctx, &tenantID)
    if err != nil {
        return err
    }
    
    // Step 2: Setup default organization structure
    err = workflow.ExecuteActivity(ctx, SetupDefaultOrganizationActivity, tenantID).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // Step 3: Create admin user
    err = workflow.ExecuteActivity(ctx, CreateAdminUserActivity, tenantID, req).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // Step 4: Send welcome email
    err = workflow.ExecuteActivity(ctx, SendWelcomeEmailActivity, tenantID).Get(ctx, nil)
    if err != nil {
        // Log error but don't fail workflow
        workflow.GetLogger(ctx).Error("Failed to send welcome email", "error", err)
    }
    
    return nil
}

// CreateTenantActivity creates a new tenant
func CreateTenantActivity(ctx context.Context, req CreateTenantRequest) (string, error) {
    // This would interact with your service layer
    // Implementation details here...
    return "", nil
}
```

#### `/internal/core/organization/`

**Files**:
```
organization/
├── model.go              # Organization hierarchy entities
├── repository.go         # Organization data access
├── service.go           # Organization business logic
├── workflow.go          # Org setup workflows
└── tree.go              # Hierarchy operations
```

**Example: `model.go`**
```go
package organization

import (
    "time"
    "github.com/google/uuid"
)

// Organization represents an organizational unit
type Organization struct {
    ID          uuid.UUID              `json:"id"`
    TenantID    uuid.UUID              `json:"tenant_id"`
    ParentID    *uuid.UUID             `json:"parent_id,omitempty"`
    Name        string                 `json:"name"`
    Code        string                 `json:"code,omitempty"`
    Type        OrganizationType       `json:"type"`
    Description string                 `json:"description,omitempty"`
    ManagerID   *uuid.UUID             `json:"manager_id,omitempty"`
    IsActive    bool                   `json:"is_active"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
    
    // Relations (loaded separately)
    Parent   *Organization   `json:"parent,omitempty"`
    Children []*Organization `json:"children,omitempty"`
    Manager  *User          `json:"manager,omitempty"` // From user domain
}

// OrganizationType defines types of organizational units
type OrganizationType string

const (
    OrgTypeDepartment OrganizationType = "department"
    OrgTypeRegion     OrganizationType = "region"
    OrgTypeDivision   OrganizationType = "division"
    OrgTypeTeam       OrganizationType = "team"
    OrgTypeOffice     OrganizationType = "office"
)

// OrganizationUser represents user assignment to organization
type OrganizationUser struct {
    ID             uuid.UUID `json:"id"`
    TenantID       uuid.UUID `json:"tenant_id"`
    OrganizationID uuid.UUID `json:"organization_id"`
    UserID         uuid.UUID `json:"user_id"`
    Role           string    `json:"role"`
    IsPrimary      bool      `json:"is_primary"`
    JoinedAt       time.Time `json:"joined_at"`
}
```

**Example: `tree.go` (Hierarchy Operations)**
```go
package organization

import (
    "context"
    "github.com/google/uuid"
)

// TreeService handles organization hierarchy operations
type TreeService interface {
    GetHierarchy(ctx context.Context, tenantID uuid.UUID) (*OrganizationTree, error)
    GetAncestors(ctx context.Context, orgID uuid.UUID) ([]*Organization, error)
    GetDescendants(ctx context.Context, orgID uuid.UUID) ([]*Organization, error)
    MoveOrganization(ctx context.Context, orgID, newParentID uuid.UUID) error
}

// OrganizationTree represents the complete org hierarchy
type OrganizationTree struct {
    Root     *OrganizationNode `json:"root"`
    TenantID uuid.UUID         `json:"tenant_id"`
}

// OrganizationNode represents a node in the org tree
type OrganizationNode struct {
    Organization *Organization       `json:"organization"`
    Children     []*OrganizationNode `json:"children"`
    Level        int                 `json:"level"`
}

// treeService implements TreeService
type treeService struct {
    repo Repository
}

// NewTreeService creates a new tree service
func NewTreeService(repo Repository) TreeService {
    return &treeService{repo: repo}
}

// GetHierarchy builds complete organization hierarchy
func (ts *treeService) GetHierarchy(ctx context.Context, tenantID uuid.UUID) (*OrganizationTree, error) {
    // Get all organizations for tenant
    orgs, err := ts.repo.GetByTenant(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    
    // Build tree structure
    tree := &OrganizationTree{TenantID: tenantID}
    nodeMap := make(map[uuid.UUID]*OrganizationNode)
    
    // Create nodes
    for _, org := range orgs {
        node := &OrganizationNode{
            Organization: org,
            Children:     []*OrganizationNode{},
        }
        nodeMap[org.ID] = node
    }
    
    // Build parent-child relationships
    for _, org := range orgs {
        node := nodeMap[org.ID]
        if org.ParentID != nil {
            parent := nodeMap[*org.ParentID]
            parent.Children = append(parent.Children, node)
            node.Level = parent.Level + 1
        } else {
            // Root organization
            tree.Root = node
            node.Level = 0
        }
    }
    
    return tree, nil
}
```

---

### `/internal/platform/` - Infrastructure & Platform Concerns

**Purpose**: Contains infrastructure code that supports the business domains but isn't part of the core business logic.

#### `/internal/platform/config/`

**Files**:
```
config/
├── config.go             # Configuration structure and loading
├── validation.go         # Configuration validation
└── defaults.go          # Default values
```

**Example: `config.go`**
```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

// Config represents application configuration
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    Temporal TemporalConfig `yaml:"temporal"`
    Auth     AuthConfig     `yaml:"auth"`
    Features FeatureConfig  `yaml:"features"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
    Port         string        `yaml:"port"`
    ReadTimeout  time.Duration `yaml:"read_timeout"`
    WriteTimeout time.Duration `yaml:"write_timeout"`
    GRPCPort     string        `yaml:"grpc_port"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
    Host            string `yaml:"host"`
    Port            int    `yaml:"port"`
    User            string `yaml:"user"`
    Password        string `yaml:"password"`
    Database        string `yaml:"database"`
    SSLMode         string `yaml:"ssl_mode"`
    MaxOpenConns    int    `yaml:"max_open_conns"`
    MaxIdleConns    int    `yaml:"max_idle_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// Load loads configuration from environment variables and files
func Load() *Config {
    config := &Config{}
    
    // Load from environment variables with defaults
    config.Server.Port = getEnv("SERVER_PORT", "8080")
    config.Server.GRPCPort = getEnv("GRPC_PORT", "9090")
    config.Server.ReadTimeout = getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second)
    config.Server.WriteTimeout = getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second)
    
    config.Database.Host = getEnv("DB_HOST", "localhost")
    config.Database.Port = getIntEnv("DB_PORT", 5432)
    config.Database.User = getEnv("DB_USER", "postgres")
    config.Database.Password = getEnv("DB_PASSWORD", "")
    config.Database.Database = getEnv("DB_NAME", "erp_system")
    config.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")
    config.Database.MaxOpenConns = getIntEnv("DB_MAX_OPEN_CONNS", 25)
    config.Database.MaxIdleConns = getIntEnv("DB_MAX_IDLE_CONNS", 5)
    config.Database.ConnMaxLifetime = getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute)
    
    // Validate configuration
    if err := config.Validate(); err != nil {
        panic(fmt.Sprintf("Invalid configuration: %v", err))
    }
    
    return config
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
```

#### `/internal/platform/database/`

**Files**:
```
database/
├── connection.go         # Database connection management
├── migrations.go         # Migration runner
├── transaction.go        # Transaction utilities
└── health.go            # Database health checks
```

**Example: `connection.go`**
```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
    "github.com/niiniyare/erp/internal/platform/config"
)

// Database interface abstracts database operations
type Database interface {
    ExecContext(ctx context.Context, query string, args ...interface{}) error
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
    BeginTx(ctx context.Context) (Transaction, error)
    Close() error
    Health(ctx context.Context) error
}

// Transaction interface for database transactions
type Transaction interface {
    ExecContext(ctx context.Context, query string, args ...interface{}) error
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
    Commit() error
    Rollback() error
}

// database implements Database interface
type database struct {
    db *sql.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg config.DatabaseConfig) (Database, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
    )
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    
    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return &database{db: db}, nil
}

// ExecContext executes a query with context
func (d *database) ExecContext(ctx context.Context, query string, args ...interface{}) error {
    _, err := d.db.ExecContext(ctx, query, args...)
    return err
}

// QueryContext executes a query that returns rows
func (d *database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    return d.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row
func (d *database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
    return d.db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a new transaction
func (d *database) BeginTx(ctx context.Context) (Transaction, error) {
    tx, err := d.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    return &transaction{tx: tx}, nil
}
```

#### `/internal/platform/middleware/`

**Files**:
```
middleware/
├── tenant.go             # Tenant context middleware
├── auth.go              # Authentication middleware
├── logging.go           # Request logging
├── cors.go              # CORS handling
└── ratelimit.go         # Rate limiting
```

**Example: `tenant.go`**
```go
package middleware

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/shared/errors"
)

// TenantContextKey is the key for tenant context
type TenantContextKey string

const (
    TenantIDKey TenantContextKey = "tenant_id"
    TenantKey   TenantContextKey = "tenant"
)

// TenantMiddleware extracts tenant from subdomain and sets context
func TenantMiddleware(tenantService tenant.Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract subdomain from request
        subdomain := extractSubdomain(c.Request.Host)
        if subdomain == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subdomain"})
            c.Abort()
            return
        }
        
        // Get tenant by subdomain
        tenant, err := tenantService.GetTenantBySubdomain(c.Request.Context(), subdomain)
        if err != nil {
            if errors.Is(err, errors.ErrTenantNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            }
            c.Abort()
            return
        }
        
        // Set tenant context for RLS
        ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenant.ID.String())
        ctx = context.WithValue(ctx, TenantKey, tenant)
        c.Request = c.Request.WithContext(ctx)
        
        // Set database session variable for RLS
        if err := setTenantRLS(ctx, tenant.ID.String()); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set tenant context"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// extractSubdomain extracts subdomain from host
func extractSubdomain(host string) string {
    parts := strings.Split(host, ".")
    if len(parts) >= 3 {
        return parts[0]
    }
    return ""
}

// setTenantRLS sets the PostgreSQL session variable for RLS
func setTenantRLS(ctx context.Context, tenantID string) error {
    // This would execute: SET app.tenant_id = 'tenant-uuid'
    // Implementation depends on your database connection
    return nil
}

// GetTenantFromContext retrieves tenant from context
func GetTenantFromContext(ctx context.Context) (*tenant.Tenant, bool) {
    tenant, ok := ctx.Value(TenantKey).(*tenant.Tenant)
    return tenant, ok
}

// GetTenantIDFromContext retrieves tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (string, bool) {
    tenantID, ok := ctx.Value(TenantIDKey).(string)
    return tenantID, ok
}
```

---

### `/internal/api/` - API Layer

**Purpose**: Contains API-related code including Goa design files, generated code, and custom handlers.

#### `/internal/api/design/`

**Files**:
```
design/
├── api.go               # Main API definition
├── tenant.go            # Tenant API design
├── organization.go      # Organization API design
├── auth.go             # Authentication API design
└── types.go            # Common type definitions
```

**Example: `api.go`**
```go
package design

import (
    . "goa.design/goa/v3/dsl"
)

// API defines the global properties of the API
var _ = API("erp-system", func() {
    Title("Multi-Tenant ERP System API")
    Description("A scalable multi-tenant ERP system with REST and gRPC interfaces")
    Version("1.0")
    
    Server("erp-server", func() {
        Description("Main ERP server")
        Host("localhost", func() {
            URI("http://localhost:8080")
        })
    })
    
    // Global security definition
    BasicAuthSecurity("basic_auth", func() {
        Description("Basic authentication")
    })
    
    JWTSecurity("jwt", func() {
        Description("JWT token authentication")
        Scope("tenant:read", "Read tenant data")
        Scope("tenant:write", "Write tenant data")
        Scope("admin", "Administrative access")
    })
})

// Common result types
var TenantResult = ResultType("application/vnd.tenant", func() {
    Description("Tenant result type")
    
    Attributes(func() {
        Attribute("id", String, "Tenant unique identifier")
        Attribute("name", String, "Tenant name")
        Attribute("subdomain", String, "Tenant subdomain")
        Attribute("plan_type", String, "Subscription plan type")
        Attribute("status", String, "Tenant status")
        Attribute("created_at", String, "Creation timestamp")
        Attribute("updated_at", String, "Update timestamp")
        
        Required("id", "name", "subdomain", "plan_type", "status")
    })
    
    View("default", func() {
        Attribute("id")
        Attribute("name")
        Attribute("subdomain")
        Attribute("plan_type")
        Attribute("status")
        Attribute("created_at")
        Attribute("updated_at")
    })
    
    View("minimal", func() {
        Attribute("id")
        Attribute("name")
        Attribute("subdomain")
    })
})
```

**Example: `tenant.go`**
```go
package design

import (
    . "goa.design/goa/v3/dsl"
)

// TenantService defines the tenant management service
var _ = Service("tenant", func() {
    Description("Tenant management service")
    
    // Create tenant endpoint
    Method("create", func() {
        Description("Create a new tenant")
        
        Payload(func() {
            Attribute("name", String, "Tenant name")
            Attribute("subdomain", String, "Unique subdomain")
            Attribute("plan_type", String, "Subscription plan")
            Attribute("settings", MapOf(String, Any), "Custom settings")
            
            Required("name", "subdomain")
        })
        
        Result(TenantResult)
        
        Error("bad_request", String, "Invalid request")
        Error("conflict", String, "Subdomain already exists")
        Error("internal_error", String, "Internal server error")
        
        HTTP(func() {
            POST("/tenants")
            Response(StatusCreated)
            Response("bad_request", StatusBadRequest)
            Response("conflict", StatusConflict)
            Response("internal_error", StatusInternalServerError)
        })
        
        GRPC(func() {
            Response(CodeOK)
            Response("bad_request", CodeInvalidArgument)
            Response("conflict", CodeAlreadyExists)
            Response("internal_error", CodeInternal)
        })
    })
    
    // Get tenant endpoint
    Method("get", func() {
        Description("Get tenant by ID")
        
        Payload(func() {
            Attribute("id", String, "Tenant ID")
            Required("id")
        })
        
        Result(TenantResult)
        
        Error("not_found", String, "Tenant not found")
        Error("internal_error", String, "Internal server error")
        
        Security(JWTSecurity)
        
        HTTP(func() {
            GET("/tenants/{id}")
            Response(StatusOK)
            Response("not_found", StatusNotFound)
            Response("internal_error", StatusInternalServerError)
        })
        
        GRPC(func() {
            Response(CodeOK)
            Response("not_found", CodeNotFound)
            Response("internal_error", CodeInternal)
        })
    })
    
    // List tenants endpoint
    Method("list", func() {
        Description("List tenants with pagination")
        
        Payload(func() {
            Attribute("offset", Int, "Pagination offset")
            Attribute("limit", Int, "Pagination limit")
        })
        
        Result(CollectionOf(TenantResult, func() {
            View("minimal")
        }))
        
        Security(JWTSecurity, "admin")
        
        HTTP(func() {
            GET("/tenants")
            Param("offset")
            Param("limit")
            Response(StatusOK)
        })
        
        GRPC(func() {
            Response(CodeOK)
        })
    })
})
```

#### `/internal/api/handlers/`

**Files**:
```
handlers/
├── tenant.go            # Tenant HTTP/gRPC handlers
├── organization.go      # Organization handlers  
├── auth.go             # Authentication handlers
├── health.go           # Health check handlers
└── router.go           # Router setup
```

**Example: `tenant.go`**
```go
package handlers

import (
    "context"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/api/gen/tenant" as genTenant
    "github.com/niiniyare/erp/internal/shared/errors"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
    service tenant.Service
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(service tenant.Service) *TenantHandler {
    return &TenantHandler{service: service}
}

// CreateTenant handles tenant creation requests
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    var req tenant.CreateTenantRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Validate request
    if err := validateCreateTenantRequest(req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Create tenant
    newTenant, err := h.service.CreateTenant(c.Request.Context(), req)
    if err != nil {
        switch {
        case errors.Is(err, errors.ErrSubdomainAlreadyExists):
            c.JSON(http.StatusConflict, gin.H{"error": "Subdomain already exists"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        }
        return
    }
    
    c.JSON(http.StatusCreated, newTenant)
}

// GetTenant handles tenant retrieval requests
func (h *TenantHandler) GetTenant(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
        return
    }
    
    tenant, err := h.service.GetTenant(c.Request.Context(), id)
    if err != nil {
        switch {
        case errors.Is(err, errors.ErrTenantNotFound):
            c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        }
        return
    }
    
    c.JSON(http.StatusOK, tenant)
}

// Goa-generated service implementation
type tenantService struct {
    handler *TenantHandler
}

// NewTenantService creates a Goa service implementation
func NewTenantService(handler *TenantHandler) genTenant.Service {
    return &tenantService{handler: handler}
}

// Create implements the generated service interface
func (s *tenantService) Create(ctx context.Context, p *genTenant.CreatePayload) (*genTenant.TenantResult, error) {
    req := tenant.CreateTenantRequest{
        Name:      p.Name,
        Subdomain: p.Subdomain,
        PlanType:  tenant.PlanType(p.PlanType),
        Settings:  p.Settings,
    }
    
    newTenant, err := s.handler.service.CreateTenant(ctx, req)
    if err != nil {
        // Convert to Goa errors
        switch {
        case errors.Is(err, errors.ErrSubdomainAlreadyExists):
            return nil, genTenant.MakeConflict(err)
        default:
            return nil, genTenant.MakeInternalError(err)
        }
    }
    
    // Convert to Goa result type
    result := &genTenant.TenantResult{
        ID:        newTenant.ID.String(),
        Name:      newTenant.Name,
        Subdomain: newTenant.Subdomain,
        PlanType:  string(newTenant.PlanType),
        Status:    string(newTenant.Status),
        CreatedAt: newTenant.CreatedAt.Format(time.RFC3339),
        UpdatedAt: newTenant.UpdatedAt.Format(time.RFC3339),
    }
    
    return result, nil
}
```

---

### `/internal/workflows/` - Temporal Workflows

**Purpose**: Contains Temporal workflow definitions organized by domain.

**Files**:
```
workflows/
├── tenant/
│   ├── onboarding.go    # Tenant onboarding workflow
│   └── lifecycle.go     # Tenant lifecycle workflows
├── organization/
│   ├── setup.go         # Organization setup workflow
│   └── restructure.go   # Org restructuring workflow
└── shared/
    ├── activities.go    # Shared activities
    └── utils.go         # Workflow utilities
```

---

### `/internal/shared/` - Shared Utilities

**Purpose**: Contains code shared across multiple domains.

**Files**:
```
shared/
├── errors/
│   ├── errors.go        # Common error definitions
│   └── codes.go         # Error codes
├── types/
│   ├── pagination.go   # Pagination types
│   └── common.go       # Common types
├── utils/
│   ├── validation.go   # Validation utilities
│   ├── crypto.go       # Cryptographic utilities
│   └── strings.go      # String utilities
└── constants/
    └── constants.go    # Application constants
```

---

## 🔗 How Components Interconnect

### **Dependency Flow**
```
cmd/server/main.go
    ↓ imports & initializes
internal/platform/config
internal/platform/database
internal/core/tenant (Repository, Service)
internal/api/handlers
    ↓ handlers use
internal/core/tenant/service.go
    ↓ service uses  
internal/core/tenant/repository.go
    ↓ repository uses
internal/platform/database
```

### **Request Flow Example**
1. **HTTP Request** → `cmd/server/main.go` (router)
2. **Middleware** → `internal/platform/middleware/tenant.go` (extract tenant)
3. **Handler** → `internal/api/handlers/tenant.go` (process request)
4. **Service** → `internal/core/tenant/service.go` (business logic)
5. **Repository** → `internal/core/tenant/repository.go` (data access)
6. **Database** → `internal/platform/database/connection.go`

### **Cross-Domain Communication**
```go
// Organization service needs user information
package organization

import (
    "github.com/niiniyare/erp/internal/core/user" // Import user domain
)

type Service struct {
    repo        Repository
    userService user.Service  // Dependency injection
}

func (s *Service) AssignManager(ctx context.Context, orgID, userID uuid.UUID) error {
    // Validate user exists and belongs to tenant
    user, err := s.userService.GetUser(ctx, userID)
    if err != nil {
        return err
    }
    
    // Update organization
    return s.repo.UpdateManager(ctx, orgID, userID)
}
```

This structure provides clear separation of concerns while maintaining flexibility for future microservices extraction. Each domain is self-contained but can communicate through well-defined interfaces.

follow this project structure and re implement this project and ask me as you need for clearification  
