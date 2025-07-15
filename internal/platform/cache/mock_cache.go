package cache

import (
	"context"
	"time"
)

// MockCache is a mock implementation of the Cache interface
type MockCache struct {
	store map[string][]byte
}

// NewMockCache creates a new mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		store: make(map[string][]byte),
	}
}

// Get retrieves an item from the cache
func (c *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	_, ok := c.store[key]
	if !ok {
		return ErrCacheMiss
	}
	return nil
}

// Set adds an item to the cache
func (c *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.store[key] = []byte{}
	return nil
}

// Delete removes an item from the cache
func (c *MockCache) Delete(ctx context.Context, key string) error {
	delete(c.store, key)
	return nil
}

// Flush flushes the cache
func (c *MockCache) Flush(ctx context.Context) error {
	c.store = make(map[string][]byte)
	return nil
}
