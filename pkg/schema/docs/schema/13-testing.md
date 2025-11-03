# Testing

## Unit Tests

### Test Schema Parsing

```go
func TestParseSchema(t *testing.T) {
    jsonData := []byte(`{
        "id": "test-form",
        "type": "form",
        "title": "Test"
    }`)
    
    parser := parse.NewParser()
    schema, err := parser.Parse(jsonData)
    
    assert.NoError(t, err)
    assert.Equal(t, "test-form", schema.ID)
}
```

### Test Validation

```go
func TestValidation(t *testing.T) {
    schema := &schema.Schema{
        Fields: []schema.Field{
            {
                Name:     "email",
                Type:     "email",
                Required: true,
            },
        },
    }
    
    validator := validate.NewValidator(nil)
    
    // Test empty - should fail
    errors := validator.ValidateData(ctx, schema, map[string]any{})
    assert.Len(t, errors, 1)
    
    // Test valid
    errors = validator.ValidateData(ctx, schema, map[string]any{
        "email": "test@example.com",
    })
    assert.Len(t, errors, 0)
}
```

### Test Storage

```go
func TestFilesystemStorage(t *testing.T) {
    storage := registry.NewFilesystemStorage("./test_schemas")
    defer os.RemoveAll("./test_schemas")
    
    // Test Set
    data := []byte(`{"id":"test","type":"form","title":"Test"}`)
    err := storage.Set(ctx, "test", data)
    assert.NoError(t, err)
    
    // Test Get
    retrieved, err := storage.Get(ctx, "test")
    assert.NoError(t, err)
    assert.Equal(t, data, retrieved)
    
    // Test Delete
    err = storage.Delete(ctx, "test")
    assert.NoError(t, err)
    
    // Test Exists
    exists, err := storage.Exists(ctx, "test")
    assert.False(t, exists)
}
```

## Integration Tests

```go
func TestFormHandler(t *testing.T) {
    app := fiber.New()
    
    // Setup
    registry := registry.NewFilesystemRegistry("./schemas")
    app.Get("/form/:id", func(c *fiber.Ctx) error {
        schema, _ := registry.Get(c.Context(), c.Params("id"))
        return views.FormPage(schema, nil).Render(
            c.Context(),
            c.Response().BodyWriter(),
        )
    })
    
    // Test
    req := httptest.NewRequest("GET", "/form/contact-form", nil)
    resp, err := app.Test(req)
    
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Mock Storage

```go
type MockStorage struct {
    data map[string][]byte
}

func (m *MockStorage) Get(ctx context.Context, id string) ([]byte, error) {
    if data, ok := m.data[id]; ok {
        return data, nil
    }
    return nil, fmt.Errorf("not found")
}

func (m *MockStorage) Set(ctx context.Context, id string, data []byte) error {
    m.data[id] = data
    return nil
}
```

[← Back](12-renderer.md) | [Home](README.md)