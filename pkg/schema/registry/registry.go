package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

// Storage defines the interface for schema storage backends
type Storage interface {
	// Get retrieves a schema by ID
	Get(ctx context.Context, id string) ([]byte, error)

	// Set stores a schema
	Set(ctx context.Context, id string, data []byte) error

	// Delete removes a schema
	Delete(ctx context.Context, id string) error

	// List returns all schema IDs
	List(ctx context.Context) ([]string, error)

	// Exists checks if schema exists
	Exists(ctx context.Context, id string) (bool, error)
}

// Registry manages schema storage and caching
type Registry struct {
	storage     Storage
	cache       Storage // Optional Redis cache
	memoryCache map[string]*cacheEntry
	mu          sync.RWMutex
	parser      *Parser
}

type cacheEntry struct {
	schema    *schema.Schema
	expiresAt time.Time
}

// RegistryConfig configures the registry
type RegistryConfig struct {
	Storage     Storage
	Cache       Storage // Optional Redis cache
	MemoryTTL   time.Duration
	CacheTTL    time.Duration
	EnableCache bool
}

// NewRegistry creates a new schema registry
func NewRegistry(config RegistryConfig) *Registry {
	if config.MemoryTTL == 0 {
		config.MemoryTTL = 5 * time.Minute
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 1 * time.Hour
	}

	return &Registry{
		storage:     config.Storage,
		cache:       config.Cache,
		memoryCache: make(map[string]*cacheEntry),
		parser:      NewParser(),
	}
}

// Get retrieves a schema by ID with caching
func (r *Registry) Get(ctx context.Context, id string) (*schema.Schema, error) {
	// 1. Check memory cache (1ms)
	r.mu.RLock()
	if entry, exists := r.memoryCache[id]; exists && time.Now().Before(entry.expiresAt) {
		r.mu.RUnlock()
		return entry.schema, nil
	}
	r.mu.RUnlock()

	// 2. Check Redis cache (5ms)
	if r.cache != nil {
		if data, err := r.cache.Get(ctx, id); err == nil {
			schema, err := r.parser.Parse(data)
			if err == nil {
				// Cache in memory
				r.cacheInMemory(id, schema)
				return schema, nil
			}
		}
	}

	// 3. Load from primary storage (50ms)
	data, err := r.storage.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %s", id)
	}

	// Parse JSON to schema
	schema, err := r.parser.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	// Cache in Redis
	if r.cache != nil {
		go func() {
			// Cache in background to not block request
			r.cache.Set(context.Background(), id, data)
		}()
	}

	// Cache in memory
	r.cacheInMemory(id, schema)

	return schema, nil
}

// Set stores a schema
func (r *Registry) Set(ctx context.Context, id string, schema *schema.Schema) error {
	// Validate schema first
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Marshal to JSON
	data, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Store in primary storage
	if err := r.storage.Set(ctx, id, data); err != nil {
		return fmt.Errorf("failed to store schema: %w", err)
	}

	// Update caches
	r.cacheInMemory(id, schema)
	if r.cache != nil {
		go func() {
			r.cache.Set(context.Background(), id, data)
		}()
	}

	return nil
}

// Delete removes a schema
func (r *Registry) Delete(ctx context.Context, id string) error {
	// Delete from primary storage
	if err := r.storage.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete schema: %w", err)
	}

	// Remove from caches
	r.mu.Lock()
	delete(r.memoryCache, id)
	r.mu.Unlock()

	if r.cache != nil {
		go func() {
			r.cache.Delete(context.Background(), id)
		}()
	}

	return nil
}

// List returns all schema IDs
func (r *Registry) List(ctx context.Context) ([]string, error) {
	return r.storage.List(ctx)
}

// Exists checks if a schema exists
func (r *Registry) Exists(ctx context.Context, id string) (bool, error) {
	return r.storage.Exists(ctx, id)
}

// cacheInMemory stores schema in memory cache
func (r *Registry) cacheInMemory(id string, schema *schema.Schema) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.memoryCache[id] = &cacheEntry{
		schema:    schema,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
}

// CleanupExpired removes expired entries from memory cache
func (r *Registry) CleanupExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, entry := range r.memoryCache {
		if now.After(entry.expiresAt) {
			delete(r.memoryCache, id)
		}
	}
}

// StartCleanupTask starts a background task to clean expired entries
func (r *Registry) StartCleanupTask() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			r.CleanupExpired()
		}
	}()
}