## Section 20: Testing Strategies

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Unit testing schemas
- Integration testing
- End-to-end testing
- Test data generation
- Testing best practices

### 20.1 Schema Unit Tests

```go
// schema_test.go

package schema_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/awoerp/internal/schema"
)

func TestNewSchema(t *testing.T) {
    s := schema.NewSchema("test-form", schema.TypeForm, "Test Form")
    
    assert.Equal(t, "test-form", s.ID)
    assert.Equal(t, schema.TypeForm, s.Type)
    assert.Equal(t, "Test Form", s.Title)
    assert.NotNil(t, s.Meta)
    assert.NotNil(t, s.State)
}

func TestSchemaValidation(t *testing.T) {
    tests := []struct {
        name    string
        schema  *schema.Schema
        wantErr bool
    }{
        {
            name: "valid schema",
            schema: &schema.Schema{
                ID:    "valid",
                Type:  schema.TypeForm,
                Title: "Valid Form",
                Fields: []schema.Field{
                    {Name: "email", Type: schema.FieldEmail, Label: "Email"},
                },
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            schema: &schema.Schema{
                Type:  schema.TypeForm,
                Title: "Invalid Form",
            },
            wantErr: true,
        },
        {
            name: "duplicate field names",
            schema: &schema.Schema{
                ID:    "duplicate",
                Type:  schema.TypeForm,
                Title: "Duplicate Fields",
                Fields: []schema.Field{
                    {Name: "email", Type: schema.FieldEmail, Label: "Email 1"},
                    {Name: "email", Type: schema.FieldEmail, Label: "Email 2"},
                },
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.schema.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestCircularDependencies(t *testing.T) {
    s := &schema.Schema{
        ID:    "circular",
        Type:  schema.TypeForm,
        Title: "Circular Dependencies",
        Fields: []schema.Field{
            {
                Name:         "field_a",
                Type:         schema.FieldText,
                Label:        "Field A",
                Dependencies: []string{"field_b"},
            },
            {
                Name:         "field_b",
                Type:         schema.FieldText,
                Label:        "Field B",
                Dependencies: []string{"field_a"},
            },
        },
    }
    
    err := s.DetectCircularDependencies()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circular dependency")
}
```

### 20.2 Validator Tests

```go
// validator_test.go

package validate_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/awoerp/internal/schema"
    "github.com/yourusername/awoerp/internal/validate"
)

func TestValidateField(t *testing.T) {
    validator := validate.NewValidator(nil)
    ctx := context.Background()
    
    tests := []struct {
        name    string
        field   *schema.Field
        value   any
        wantErr bool
    }{
        {
            name: "required field with value",
            field: &schema.Field{
                Name:     "email",
                Type:     schema.FieldEmail,
                Label:    "Email",
                Required: true,
            },
            value:   "test@example.com",
            wantErr: false,
        },
        {
            name: "required field without value",
            field: &schema.Field{
                Name:     "email",
                Type:     schema.FieldEmail,
                Label:    "Email",
                Required: true,
            },
            value:   "",
            wantErr: true,
        },
        {
            name: "min length validation",
            field: &schema.Field{
                Name:  "username",
                Type:  schema.FieldText,
                Label: "Username",
                Validation: &schema.FieldValidation{
                    MinLength: intPtr(3),
                },
            },
            value:   "ab",
            wantErr: true,
        },
        {
            name: "pattern validation",
            field: &schema.Field{
                Name:  "username",
                Type:  schema.FieldText,
                Label: "Username",
                Validation: &schema.FieldValidation{
                    Pattern: "^[a-z]+$",
                },
            },
            value:   "ABC123",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errors := validator.ValidateField(ctx, tt.field, tt.value, nil)
            if tt.wantErr {
                assert.NotEmpty(t, errors)
            } else {
                assert.Empty(t, errors)
            }
        })
    }
}

func intPtr(i int) *int {
    return &i
}
```

### 20.3 Registry Tests

```go
// registry_test.go

package schema_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/awoerp/internal/schema"
)

func TestMemoryRegistry(t *testing.T) {
    registry := schema.NewMemoryRegistry()
    ctx := context.Background()
    
    // Test Register
    s := schema.NewSchema("test-form", schema.TypeForm, "Test Form")
    err := registry.Register(ctx, s)
    assert.NoError(t, err)
    
    // Test Get
    retrieved, err := registry.Get(ctx, "test-form")
    assert.NoError(t, err)
    assert.Equal(t, s.ID, retrieved.ID)
    assert.Equal(t, s.Title, retrieved.Title)
    
    // Test Exists
    exists := registry.Exists(ctx, "test-form")
    assert.True(t, exists)
    
    // Test Delete
    err = registry.Delete(ctx, "test-form")
    assert.NoError(t, err)
    
    exists = registry.Exists(ctx, "test-form")
    assert.False(t, exists)
}

func TestRegistryFilters(t *testing.T) {
    registry := schema.NewMemoryRegistry()
    ctx := context.Background()
    
    // Register multiple schemas
    schemas := []*schema.Schema{
        {ID: "form1", Type: schema.TypeForm, Title: "Form 1", Category: "users"},
        {ID: "form2", Type: schema.TypeForm, Title: "Form 2", Category: "users"},
        {ID: "form3", Type: schema.TypeForm, Title: "Form 3", Category: "products"},
    }
    
    for _, s := range schemas {
        registry.Register(ctx, s)
    }
    
    // Filter by category
    category := "users"
    filters := &schema.Filters{
        Category: &category,
    }
    
    results, err := registry.List(ctx, filters)
    assert.NoError(t, err)
    assert.Len(t, results, 2)
}
```

### 20.4 Handler Integration Tests

```go
// handler_test.go

package handlers_test

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/awoerp/handlers"
    "github.com/yourusername/awoerp/internal/schema"
)

func TestHandleCreate(t *testing.T) {
    // Setup
    registry := schema.NewMemoryRegistry()
    validator := validate.NewValidator(nil)
    enricher := schema.NewEnricher()
    repo := &MockRepository{}
    
    handler := handlers.NewSchemaFormHandler(registry, enricher, validator, repo)
    
    // Register test schema
    s := schema.NewSchema("test-form", schema.TypeForm, "Test Form")
    s.AddField(schema.Field{
        Name:     "email",
        Type:     schema.FieldEmail,
        Label:    "Email",
        Required: true,
    })
    registry.Register(context.Background(), s)
    
    // Test valid submission
    t.Run("valid submission", func(t *testing.T) {
        form := strings.NewReader("email=test@example.com")
        req := httptest.NewRequest(http.MethodPost, "/schema/test-form", form)
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        
        w := httptest.NewRecorder()
        handler.HandleCreate(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        assert.Contains(t, w.Body.String(), "success")
    })
    
    // Test invalid submission
    t.Run("invalid submission", func(t *testing.T) {
        form := strings.NewReader("email=")
        req := httptest.NewRequest(http.MethodPost, "/schema/test-form", form)
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        
        w := httptest.NewRecorder()
        handler.HandleCreate(w, req)
        
        assert.Contains(t, w.Body.String(), "required")
    })
}

// Mock repository for testing
type MockRepository struct {
    data map[string]map[string]any
}

func (r *MockRepository) Create(ctx context.Context, schemaID string, data map[string]any) (string, error) {
    if r.data == nil {
        r.data = make(map[string]map[string]any)
    }
    id := uuid.New().String()
    r.data[id] = data
    return id, nil
}

func (r *MockRepository) Get(ctx context.Context, schemaID, id string) (map[string]any, error) {
    return r.data[id], nil
}

func (r *MockRepository) Update(ctx context.Context, schemaID, id string, data map[string]any) error {
    r.data[id] = data
    return nil
}

func (r *MockRepository) Delete(ctx context.Context, schemaID, id string) error {
    delete(r.data, id)
    return nil
}
```

### 20.5 End-to-End Tests

```go
// e2e_test.go

package e2e_test

import (
    "context"
    "testing"
    "github.com/chromedp/chromedp"
    "github.com/stretchr/testify/assert"
)

func TestUserRegistrationFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }
    
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()
    
    var title string
    err := chromedp.Run(ctx,
        chromedp.Navigate("http://localhost:8080/schema/user-registration/new"),
        chromedp.Title(&title),
        chromedp.WaitVisible(`#email`),
        chromedp.SendKeys(`#email`, "test@example.com"),
        chromedp.SendKeys(`#password`, "SecurePass123!"),
        chromedp.Click(`button[type="submit"]`),
        chromedp.WaitVisible(`.alert-success`),
    )
    
    assert.NoError(t, err)
    assert.Contains(t, title, "User Registration")
}
```

### 20.6 Test Data Generation

```go
// testdata/generator.go

package testdata

import (
    "github.com/yourusername/awoerp/internal/schema"
)

// GenerateTestSchema creates a test schema
func GenerateTestSchema(id string) *schema.Schema {
    s := schema.NewSchema(id, schema.TypeForm, "Test Form")
    
    s.AddField(schema.Field{
        Name:     "first_name",
        Type:     schema.FieldText,
        Label:    "First Name",
        Required: true,
    })
    
    s.AddField(schema.Field{
        Name:     "last_name",
        Type:     schema.FieldText,
        Label:    "Last Name",
        Required: true,
    })
    
    s.AddField(schema.Field{
        Name:     "email",
        Type:     schema.FieldEmail,
        Label:    "Email",
        Required: true,
        Validation: &schema.FieldValidation{
            MaxLength: intPtr(255),
        },
    })
    
    return s
}

// GenerateTestData creates test form data
func GenerateTestData() map[string]any {
    return map[string]any{
        "first_name": "John",
        "last_name":  "Doe",
        "email":      "john.doe@example.com",
    }
}

func intPtr(i int) *int {
    return &i
}
```

### 20.7 Best Practices

1. **Test at Multiple Levels**: Unit → Integration → E2E
2. **Use Table-Driven Tests**: For validation and field tests
3. **Mock External Dependencies**: Use interfaces and mocks
4. **Test Error Paths**: Not just happy paths
5. **Use Test Fixtures**: Reusable test data
6. **Separate Test Concerns**: Validation, rendering, handlers
7. **Run Tests in CI/CD**: Automated testing

---

**Part 4 (Sections 14-20) is now complete!**

We've covered:
- Section 14: Registry System
- Section 15: Validator Implementation
- Section 16: Enricher System
- Section 17: Renderer System (templ)
- Section 18: Handler Patterns
- Section 19: Data Sources
- Section 20: Testing Strategies
