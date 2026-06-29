package repository

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/platform/cache"
)

// atomicCache is a minimal cache.Service fake whose GetAndDelete is atomic
// (mutex-protected fetch-and-delete). Set stores JSON-encoded values to match
// the real Redis client's serialisation contract.
//
// All methods not used by GetPendingMFA / StorePendingMFA / DeletePendingMFA
// panic with "not implemented" so test failures surface clearly.
type atomicCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newAtomicCache() *atomicCache {
	return &atomicCache{data: make(map[string]string)}
}

func (c *atomicCache) Get(_ context.Context, key string, dest any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[key]
	if !ok {
		return cache.ErrCacheMiss
	}
	return json.Unmarshal([]byte(v), dest)
}

func (c *atomicCache) GetAndDelete(_ context.Context, key string, dest any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[key]
	if !ok {
		return cache.ErrCacheMiss
	}
	delete(c.data, key)
	return json.Unmarshal([]byte(v), dest)
}

func (c *atomicCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = string(b)
	return nil
}

func (c *atomicCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

// Unused interface methods — panic on call so test errors surface clearly.
func (c *atomicCache) Flush(_ context.Context) error { panic("not impl") }

func (c *atomicCache) MGet(_ context.Context, _ []string) ([]cache.Result, error) {
	panic("not impl")
}

func (c *atomicCache) MSet(_ context.Context, _ map[string]any, _ time.Duration) error {
	panic("not impl")
}
func (c *atomicCache) MDelete(_ context.Context, _ []string) error            { panic("not impl") }
func (c *atomicCache) DeletePattern(_ context.Context, _ string) error        { panic("not impl") }
func (c *atomicCache) Keys(_ context.Context, _ string) ([]string, error)     { panic("not impl") }
func (c *atomicCache) Exists(_ context.Context, _ string) (bool, error)       { panic("not impl") }
func (c *atomicCache) TTL(_ context.Context, _ string) (time.Duration, error) { panic("not impl") }
func (c *atomicCache) Expire(_ context.Context, _ string, _ time.Duration) error {
	panic("not impl")
}
func (c *atomicCache) GetMemory(_ context.Context, _ string, _ any) error { panic("not impl") }
func (c *atomicCache) SetMemory(_ context.Context, _ string, _ any, _ time.Duration) error {
	panic("not impl")
}
func (c *atomicCache) DeleteMemory(_ context.Context, _ string) error         { panic("not impl") }
func (c *atomicCache) GetGlobalMemory(_ string, _ any) error                  { panic("not impl") }
func (c *atomicCache) SetGlobalMemory(_ string, _ any, _ time.Duration) error { panic("not impl") }
func (c *atomicCache) DeleteGlobalMemory(_ string) error                      { panic("not impl") }
func (c *atomicCache) Ping(_ context.Context) error                           { panic("not impl") }
func (c *atomicCache) Stats() cache.CacheStats                                { panic("not impl") }
func (c *atomicCache) Reset()                                                 {}
func (c *atomicCache) Close() error                                           { return nil }

// ---------------------------------------------------------------------------

// TestGetPendingMFA_AtomicConsumption verifies that two concurrent calls to
// GetPendingMFA for the same token result in exactly one success and one
// "not found" error. This guards against duplicate-session creation from
// concurrent CompleteMFALogin requests.
func TestGetPendingMFA_AtomicConsumption(t *testing.T) {
	fakeCache := newAtomicCache()
	repo := &sessionRepo{cache: fakeCache}

	ctx := context.Background()
	pendingToken := "test-pending-token"
	userID := uuid.New()

	// Store a pending MFA entry (same path as StorePendingMFA).
	if err := repo.StorePendingMFA(ctx, pendingToken, userID); err != nil {
		t.Fatalf("StorePendingMFA: %v", err)
	}

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		successes []uuid.UUID
		failures  []error
	)

	// Fire two concurrent GetPendingMFA calls for the same token.
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := repo.GetPendingMFA(ctx, pendingToken)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failures = append(failures, err)
			} else {
				successes = append(successes, id)
			}
		}()
	}

	wg.Wait()

	if len(successes) != 1 {
		t.Errorf("want 1 success, got %d", len(successes))
	}
	if len(failures) != 1 {
		t.Errorf("want 1 failure, got %d", len(failures))
	}
	if len(successes) == 1 && successes[0] != userID {
		t.Errorf("success returned wrong userID: got %v, want %v", successes[0], userID)
	}
}

// TestGetPendingMFA_NotFound verifies that a missing or already-consumed
// pending token returns an error.
func TestGetPendingMFA_NotFound(t *testing.T) {
	repo := &sessionRepo{cache: newAtomicCache()}
	_, err := repo.GetPendingMFA(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for missing pending token, got nil")
	}
}
