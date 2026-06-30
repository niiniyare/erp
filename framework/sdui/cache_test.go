package sdui_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"

	"awo.so/framework/sdui"
)

func newTestCache(t *testing.T) (sdui.Cache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return rdb, mr
}

func TestCacheGetMiss(t *testing.T) {
	cache, _ := newTestCache(t)
	got, err := sdui.CacheGet(context.Background(), cache, "nonexistent")
	if err != nil {
		t.Fatalf("want nil error on miss, got %v", err)
	}
	if got != nil {
		t.Fatalf("want nil on miss, got %v", got)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache, _ := newTestCache(t)
	ctx := context.Background()
	payload := []byte(`{"type":"page"}`)

	if err := sdui.CacheSet(ctx, cache, "test-key", payload); err != nil {
		t.Fatal(err)
	}
	got, err := sdui.CacheGet(ctx, cache, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Errorf("want %q, got %q", payload, got)
	}
}

func TestInvalidateEntitySchemas(t *testing.T) {
	cache, _ := newTestCache(t)
	ctx := context.Background()
	hash := "abc12345deadbeef"
	tenant := "00000000-0000-0000-0000-000000000001"

	// Seed all four view types.
	for _, view := range []string{"crud", "form_create", "form_edit", "detail"} {
		_ = sdui.CacheSet(ctx, cache, "page:"+hash+":"+view+":"+tenant, []byte(`{}`))
	}

	if err := sdui.InvalidateEntitySchemas(ctx, cache, hash, tenant); err != nil {
		t.Fatal(err)
	}

	// All entries should be gone.
	for _, view := range []string{"crud", "form_create", "form_edit", "detail"} {
		got, _ := sdui.CacheGet(ctx, cache, "page:"+hash+":"+view+":"+tenant)
		if got != nil {
			t.Errorf("key for view %q should be invalidated but still exists", view)
		}
	}
}

func TestCacheExpiry(t *testing.T) {
	cache, mr := newTestCache(t)
	ctx := context.Background()

	if err := sdui.CacheSet(ctx, cache, "expiring-key", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}

	// Fast-forward miniredis TTL past 5 minutes.
	mr.FastForward(sdui.CacheTTL + 1)

	got, err := sdui.CacheGet(ctx, cache, "expiring-key")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Error("key should have expired but is still present")
	}
}
