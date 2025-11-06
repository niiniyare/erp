package registry

import (
	"context"
	"testing"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

// MockStorage for testing
type MockStorage struct {
	data map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string][]byte),
	}
}

func (m *MockStorage) Get(ctx context.Context, id string) ([]byte, error) {
	if data, exists := m.data[id]; exists {
		return data, nil
	}
	return nil, &StorageError{Operation: "get", ID: id, Message: "not found"}
}

func (m *MockStorage) Set(ctx context.Context, id string, data []byte) error {
	m.data[id] = data
	return nil
}

func (m *MockStorage) Delete(ctx context.Context, id string) error {
	if _, exists := m.data[id]; !exists {
		return &StorageError{Operation: "delete", ID: id, Message: "not found"}
	}
	delete(m.data, id)
	return nil
}

func (m *MockStorage) List(ctx context.Context) ([]string, error) {
	ids := make([]string, 0, len(m.data))
	for id := range m.data {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *MockStorage) Exists(ctx context.Context, id string) (bool, error) {
	_, exists := m.data[id]
	return exists, nil
}

// StorageError for testing
type StorageError struct {
	Operation string
	ID        string
	Message   string
}

func (e *StorageError) Error() string {
	return e.Operation + " " + e.ID + ": " + e.Message
}

func TestNewRegistry(t *testing.T) {
	storage := NewMockStorage()

	tests := []struct {
		name   string
		config RegistryConfig
		check  func(*Registry) bool
	}{
		{
			name: "default config",
			config: RegistryConfig{
				Storage: storage,
			},
			check: func(r *Registry) bool {
				return r.storage == storage && r.memoryCache != nil && r.parser != nil
			},
		},
		{
			name: "with cache storage",
			config: RegistryConfig{
				Storage: storage,
				Cache:   NewMockStorage(),
			},
			check: func(r *Registry) bool {
				return r.storage == storage && r.cache != nil
			},
		},
		{
			name: "with custom TTL",
			config: RegistryConfig{
				Storage:   storage,
				MemoryTTL: 10 * time.Minute,
				CacheTTL:  2 * time.Hour,
			},
			check: func(r *Registry) bool {
				return r.storage == storage
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry(tt.config)

			if registry == nil {
				t.Fatal("NewRegistry() returned nil")
			}

			if !tt.check(registry) {
				t.Error("Registry validation failed")
			}
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	storage := NewMockStorage()
	registry := NewRegistry(RegistryConfig{Storage: storage})
	ctx := context.Background()

	// Test schema not found
	_, err := registry.Get(ctx, "non-existent")
	if err == nil {
		t.Error("Get() should return error for non-existent schema")
	}

	// Create test schema
	testSchema := &schema.Schema{
		ID:          "test-form",
		Type:        "form",
		Title:       "Test Form",
		Description: "A test form",
		Fields:      []schema.Field{},
		Actions:     []schema.Action{},
	}

	// Store schema
	err = registry.Set(ctx, "test-form", testSchema)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Get schema
	retrieved, err := registry.Get(ctx, "test-form")
	if err != nil {
		t.Errorf("Get() failed: %v", err)
	}

	if retrieved == nil {
		t.Error("Get() returned nil schema")
	}

	if retrieved.ID != "test-form" {
		t.Errorf("Get() returned wrong schema: got %s, want test-form", retrieved.ID)
	}

	// Test memory cache hit
	retrieved2, err := registry.Get(ctx, "test-form")
	if err != nil {
		t.Errorf("Get() from cache failed: %v", err)
	}

	if retrieved2.ID != "test-form" {
		t.Error("Get() from cache returned wrong schema")
	}
}

func TestRegistry_Set(t *testing.T) {
	storage := NewMockStorage()
	registry := NewRegistry(RegistryConfig{Storage: storage})
	ctx := context.Background()

	testSchema := &schema.Schema{
		ID:          "test-form",
		Type:        "form",
		Title:       "Test Form",
		Description: "A test form",
		Fields:      []schema.Field{},
		Actions:     []schema.Action{},
	}

	// Test successful set
	err := registry.Set(ctx, "test-form", testSchema)
	if err != nil {
		t.Errorf("Set() failed: %v", err)
	}

	// Verify schema was stored
	exists, err := registry.Exists(ctx, "test-form")
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}

	if !exists {
		t.Error("Set() did not store schema")
	}

	// Test invalid schema
	invalidSchema := &schema.Schema{
		// Missing required fields
	}

	err = registry.Set(ctx, "invalid", invalidSchema)
	if err == nil {
		t.Error("Set() should fail for invalid schema")
	}
}

func TestRegistry_Delete(t *testing.T) {
	storage := NewMockStorage()
	registry := NewRegistry(RegistryConfig{Storage: storage})
	ctx := context.Background()

	testSchema := &schema.Schema{
		ID:          "test-form",
		Type:        "form",
		Title:       "Test Form",
		Description: "A test form",
		Fields:      []schema.Field{},
		Actions:     []schema.Action{},
	}

	// Store schema first
	err := registry.Set(ctx, "test-form", testSchema)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Delete schema
	err = registry.Delete(ctx, "test-form")
	if err != nil {
		t.Errorf("Delete() failed: %v", err)
	}

	// Verify schema was deleted
	exists, err := registry.Exists(ctx, "test-form")
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}

	if exists {
		t.Error("Delete() did not remove schema")
	}

	// Test delete non-existent schema
	err = registry.Delete(ctx, "non-existent")
	if err == nil {
		t.Error("Delete() should fail for non-existent schema")
	}
}

func TestRegistry_List(t *testing.T) {
	storage := NewMockStorage()
	registry := NewRegistry(RegistryConfig{Storage: storage})
	ctx := context.Background()

	// Test empty list
	schemas, err := registry.List(ctx, map[string]any{})
	if err != nil {
		t.Errorf("List() failed: %v", err)
	}

	if len(schemas) != 0 {
		t.Errorf("List() should return empty list, got %d items", len(schemas))
	}

	// Add some schemas
	testSchemas := []*schema.Schema{
		{
			ID:          "form1",
			Type:        "form",
			Title:       "Form 1",
			Description: "First form",
			Fields:      []schema.Field{},
			Actions:     []schema.Action{},
		},
		{
			ID:          "form2",
			Type:        "form",
			Title:       "Form 2",
			Description: "Second form",
			Fields:      []schema.Field{},
			Actions:     []schema.Action{},
		},
	}

	for _, s := range testSchemas {
		err := registry.Set(ctx, s.ID, s)
		if err != nil {
			t.Fatalf("Set() failed for %s: %v", s.ID, err)
		}
	}

	// Test list with items
	schemas, err = registry.List(ctx, map[string]any{})
	if err != nil {
		t.Errorf("List() failed: %v", err)
	}

	if len(schemas) != 2 {
		t.Errorf("List() should return 2 items, got %d", len(schemas))
	}

	// Check IDs are present
	found := make(map[string]bool)
	for _, s := range schemas {
		found[s.ID] = true
	}

	if !found["form1"] || !found["form2"] {
		t.Error("List() missing expected IDs")
	}
}

func TestRegistry_Exists(t *testing.T) {
	storage := NewMockStorage()
	registry := NewRegistry(RegistryConfig{Storage: storage})
	ctx := context.Background()

	// Test non-existent schema
	exists, err := registry.Exists(ctx, "non-existent")
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}

	if exists {
		t.Error("Exists() should return false for non-existent schema")
	}

	// Add schema
	testSchema := &schema.Schema{
		ID:          "test-form",
		Type:        "form",
		Title:       "Test Form",
		Description: "A test form",
		Fields:      []schema.Field{},
		Actions:     []schema.Action{},
	}

	err = registry.Set(ctx, "test-form", testSchema)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Test existing schema
	exists, err = registry.Exists(ctx, "test-form")
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}

	if !exists {
		t.Error("Exists() should return true for existing schema")
	}
}

func TestRegistry_MemoryCache(t *testing.T) {
	storage := NewMockStorage()
	registry := NewRegistry(RegistryConfig{Storage: storage})
	ctx := context.Background()

	testSchema := &schema.Schema{
		ID:          "test-form",
		Type:        "form",
		Title:       "Test Form",
		Description: "A test form",
		Fields:      []schema.Field{},
		Actions:     []schema.Action{},
	}

	// Store schema
	err := registry.Set(ctx, "test-form", testSchema)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Get schema (should cache it)
	_, err = registry.Get(ctx, "test-form")
	if err != nil {
		t.Errorf("Get() failed: %v", err)
	}

	// Verify it's in memory cache
	registry.mu.RLock()
	entry, exists := registry.memoryCache["test-form"]
	registry.mu.RUnlock()

	if !exists {
		t.Error("Schema not found in memory cache")
	}

	if entry.schema.ID != "test-form" {
		t.Error("Wrong schema in memory cache")
	}

	// Test cache expiry
	registry.mu.Lock()
	registry.memoryCache["test-form"].expiresAt = time.Now().Add(-1 * time.Hour)
	registry.mu.Unlock()

	// Get should still work (reload from storage)
	_, err = registry.Get(ctx, "test-form")
	if err != nil {
		t.Errorf("Get() failed after cache expiry: %v", err)
	}
}

func TestRegistry_CleanupExpired(t *testing.T) {
	storage := NewMockStorage()
	registry := NewRegistry(RegistryConfig{Storage: storage})

	// Add expired entry
	registry.mu.Lock()
	registry.memoryCache["expired"] = &cacheEntry{
		schema:    &schema.Schema{ID: "expired"},
		expiresAt: time.Now().Add(-1 * time.Hour),
	}
	registry.memoryCache["valid"] = &cacheEntry{
		schema:    &schema.Schema{ID: "valid"},
		expiresAt: time.Now().Add(1 * time.Hour),
	}
	registry.mu.Unlock()

	// Run cleanup
	registry.CleanupExpired()

	// Check results
	registry.mu.RLock()
	_, expiredExists := registry.memoryCache["expired"]
	_, validExists := registry.memoryCache["valid"]
	registry.mu.RUnlock()

	if expiredExists {
		t.Error("Expired entry should be removed")
	}

	if !validExists {
		t.Error("Valid entry should remain")
	}
}

func TestRegistry_WithCache(t *testing.T) {
	storage := NewMockStorage()
	cache := NewMockStorage()
	registry := NewRegistry(RegistryConfig{
		Storage: storage,
		Cache:   cache,
	})
	ctx := context.Background()

	testSchema := &schema.Schema{
		ID:          "test-form",
		Type:        "form",
		Title:       "Test Form",
		Description: "A test form",
		Fields:      []schema.Field{},
		Actions:     []schema.Action{},
	}

	// Store schema
	err := registry.Set(ctx, "test-form", testSchema)
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Clear memory cache to force cache lookup
	registry.mu.Lock()
	delete(registry.memoryCache, "test-form")
	registry.mu.Unlock()

	// Get should hit Redis cache
	retrieved, err := registry.Get(ctx, "test-form")
	if err != nil {
		t.Errorf("Get() from cache failed: %v", err)
	}

	if retrieved.ID != "test-form" {
		t.Error("Get() from cache returned wrong schema")
	}
}

// Benchmark tests
func BenchmarkRegistry_Get(b *testing.B) {
	storage := NewMockStorage()
	registry := NewRegistry(RegistryConfig{Storage: storage})
	ctx := context.Background()

	testSchema := &schema.Schema{
		ID:          "benchmark-form",
		Type:        "form",
		Title:       "Benchmark Form",
		Description: "A benchmark form",
		Fields:      []schema.Field{},
		Actions:     []schema.Action{},
	}

	registry.Set(ctx, "benchmark-form", testSchema)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.Get(ctx, "benchmark-form")
	}
}

func BenchmarkRegistry_Set(b *testing.B) {
	storage := NewMockStorage()
	registry := NewRegistry(RegistryConfig{Storage: storage})
	ctx := context.Background()

	testSchema := &schema.Schema{
		ID:          "benchmark-form",
		Type:        "form",
		Title:       "Benchmark Form",
		Description: "A benchmark form",
		Fields:      []schema.Field{},
		Actions:     []schema.Action{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.Set(ctx, "benchmark-form", testSchema)
	}
}
