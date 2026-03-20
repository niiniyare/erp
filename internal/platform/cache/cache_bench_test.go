package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"awo/internal/platform/config"
)

// Benchmark fixtures
var (
	benchCache    Service
	benchCtx      context.Context
	benchOnce     sync.Once
	benchTenantID uuid.UUID
)

func setupBench() {
	benchOnce.Do(func() {
		cfg := &RedisConfig{
			RedisConfig: &config.RedisConfig{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			},
			PoolSize:              50,
			MinIdleConns:          10,
			EnableCompression:     true,
			CompressionLevel:      6,
			KeyPrefix:             "bench",
			EnableCircuitBreaker:  false,
			RequireTenantContext:  true,
			AllowGlobalOperations: false,
			EnableMemoryCache:     true,
			MemoryCacheMaxSize:    10000,
			EnableTracing:         false,
			EnableMetrics:         false,
			EnableLogging:         false,
		}

		var err error
		benchCache, err = NewRedisClient(cfg)
		if err != nil {
			panic(err)
		}

		benchTenantID = uuid.New()
		benchCtx = WithTenantID(context.Background(), benchTenantID)
	})
}

// BenchmarkCacheGet benchmarks simple Get operations
func BenchmarkCacheGet(b *testing.B) {
	setupBench()

	// Setup test data
	key := "bench:get"
	value := "benchmark-value"
	benchCache.Set(benchCtx, key, value, time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var result string
		for pb.Next() {
			benchCache.Get(benchCtx, key, &result)
		}
	})
}

// BenchmarkCacheSet benchmarks simple Set operations
func BenchmarkCacheSet(b *testing.B) {
	setupBench()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench:set:%d", i)
			benchCache.Set(benchCtx, key, "value", time.Hour)
			i++
		}
	})
}

// BenchmarkCacheGetSet benchmarks combined Get and Set operations
func BenchmarkCacheGetSet(b *testing.B) {
	setupBench()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench:getset:%d", i)

			// Set
			benchCache.Set(benchCtx, key, "value", time.Hour)

			// Get
			var result string
			benchCache.Get(benchCtx, key, &result)

			i++
		}
	})
}

// BenchmarkCacheDelete benchmarks Delete operations
func BenchmarkCacheDelete(b *testing.B) {
	setupBench()

	// Pre-populate keys
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:delete:%d", i)
		benchCache.Set(benchCtx, key, "value", time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:delete:%d", i)
		benchCache.Delete(benchCtx, key)
	}
}

// BenchmarkCacheMGet benchmarks bulk Get operations
func BenchmarkCacheMGet(b *testing.B) {
	setupBench()

	tests := []struct {
		name     string
		keyCount int
	}{
		{"MGet-10", 10},
		{"MGet-50", 50},
		{"MGet-100", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Setup test data
			keys := make([]string, tt.keyCount)
			for i := 0; i < tt.keyCount; i++ {
				key := fmt.Sprintf("bench:mget:%s:%d", tt.name, i)
				keys[i] = key
				benchCache.Set(benchCtx, key, fmt.Sprintf("value-%d", i), time.Hour)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchCache.MGet(benchCtx, keys)
			}
		})
	}
}

// BenchmarkCacheMSet benchmarks bulk Set operations
func BenchmarkCacheMSet(b *testing.B) {
	setupBench()

	tests := []struct {
		name     string
		keyCount int
	}{
		{"MSet-10", 10},
		{"MSet-50", 50},
		{"MSet-100", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Prepare data
			pairs := make(map[string]interface{}, tt.keyCount)
			for i := 0; i < tt.keyCount; i++ {
				key := fmt.Sprintf("bench:mset:%s:%d", tt.name, i)
				pairs[key] = fmt.Sprintf("value-%d", i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchCache.MSet(benchCtx, pairs, time.Hour)
			}
		})
	}
}

// BenchmarkCacheMDelete benchmarks bulk Delete operations
func BenchmarkCacheMDelete(b *testing.B) {
	setupBench()

	tests := []struct {
		name     string
		keyCount int
	}{
		{"MDelete-10", 10},
		{"MDelete-50", 50},
		{"MDelete-100", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			keys := make([]string, tt.keyCount)
			for i := 0; i < tt.keyCount; i++ {
				keys[i] = fmt.Sprintf("bench:mdelete:%s:%d", tt.name, i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Setup keys before each iteration
				b.StopTimer()
				for _, key := range keys {
					benchCache.Set(benchCtx, key, "value", time.Hour)
				}
				b.StartTimer()

				benchCache.MDelete(benchCtx, keys)
			}
		})
	}
}

// BenchmarkMemoryCacheGet benchmarks in-memory Get operations
func BenchmarkMemoryCacheGet(b *testing.B) {
	setupBench()

	key := "bench:memory:get"
	value := "memory-value"
	benchCache.SetMemory(benchCtx, key, value, time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var result string
		for pb.Next() {
			benchCache.GetMemory(benchCtx, key, &result)
		}
	})
}

// BenchmarkMemoryCacheSet benchmarks in-memory Set operations
func BenchmarkMemoryCacheSet(b *testing.B) {
	setupBench()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench:memory:set:%d", i)
			benchCache.SetMemory(benchCtx, key, "value", time.Hour)
			i++
		}
	})
}

// BenchmarkGlobalMemoryCacheGet benchmarks global memory Get operations
func BenchmarkGlobalMemoryCacheGet(b *testing.B) {
	setupBench()

	key := "bench:global:get"
	value := "global-value"
	benchCache.SetGlobalMemory(key, value, time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var result string
		for pb.Next() {
			benchCache.GetGlobalMemory(key, &result)
		}
	})
}

// BenchmarkCompression benchmarks compression overhead
func BenchmarkCompression(b *testing.B) {
	setupBench()

	tests := []struct {
		name string
		size int
	}{
		{"Small-100B", 100},
		{"Medium-1KB", 1024},
		{"Large-10KB", 10240},
		{"VeryLarge-100KB", 102400},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create data of specified size
			data := make([]byte, tt.size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			key := fmt.Sprintf("bench:compression:%s", tt.name)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchCache.Set(benchCtx, key, data, time.Hour)
			}
		})
	}
}

// BenchmarkTenantIsolation benchmarks tenant context overhead
func BenchmarkTenantIsolation(b *testing.B) {
	setupBench()

	tenants := make([]context.Context, 10)
	for i := 0; i < 10; i++ {
		tenants[i] = WithTenantID(context.Background(), uuid.New())
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ctx := tenants[i%len(tenants)]
			key := fmt.Sprintf("bench:tenant:%d", i)
			benchCache.Set(ctx, key, "value", time.Hour)
			i++
		}
	})
}

// BenchmarkNamespace benchmarks namespace overhead
func BenchmarkNamespace(b *testing.B) {
	setupBench()

	namespaces := []string{"users", "sessions", "products", "orders", "analytics"}
	contexts := make([]context.Context, len(namespaces))

	for i, ns := range namespaces {
		contexts[i] = WithNamespace(benchCtx, ns)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ctx := contexts[i%len(contexts)]
			key := fmt.Sprintf("bench:ns:%d", i)
			benchCache.Set(ctx, key, "value", time.Hour)
			i++
		}
	})
}

// BenchmarkComplexStruct benchmarks caching complex structures
func BenchmarkComplexStruct(b *testing.B) {
	setupBench()

	type ComplexData struct {
		ID       string                 `json:"id"`
		Name     string                 `json:"name"`
		Email    string                 `json:"email"`
		Age      int                    `json:"age"`
		Active   bool                   `json:"active"`
		Tags     []string               `json:"tags"`
		Metadata map[string]interface{} `json:"metadata"`
		Created  time.Time              `json:"created"`
	}

	data := ComplexData{
		ID:     uuid.New().String(),
		Name:   "Benchmark User",
		Email:  "bench@example.com",
		Age:    30,
		Active: true,
		Tags:   []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
		Metadata: map[string]interface{}{
			"role":        "admin",
			"permissions": []string{"read", "write", "delete"},
			"settings": map[string]interface{}{
				"theme": "dark",
				"lang":  "en",
			},
		},
		Created: time.Now(),
	}

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench:complex:set:%d", i)
			benchCache.Set(benchCtx, key, data, time.Hour)
		}
	})

	b.Run("Get", func(b *testing.B) {
		key := "bench:complex:get"
		benchCache.Set(benchCtx, key, data, time.Hour)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result ComplexData
			benchCache.Get(benchCtx, key, &result)
		}
	})
}

// BenchmarkConcurrentAccess benchmarks concurrent cache access
func BenchmarkConcurrentAccess(b *testing.B) {
	setupBench()

	tests := []struct {
		name       string
		goroutines int
		readRatio  float64 // 0.0 = all writes, 1.0 = all reads
	}{
		{"ReadHeavy-100", 100, 0.9},
		{"Balanced-100", 100, 0.5},
		{"WriteHeavy-100", 100, 0.1},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Pre-populate some keys
			for i := 0; i < 1000; i++ {
				key := fmt.Sprintf("bench:concurrent:%s:%d", tt.name, i)
				benchCache.Set(benchCtx, key, fmt.Sprintf("value-%d", i), time.Hour)
			}

			b.ResetTimer()
			b.SetParallelism(tt.goroutines)
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("bench:concurrent:%s:%d", tt.name, i%1000)

					// Decide whether to read or write based on ratio
					if float64(i%100)/100.0 < tt.readRatio {
						// Read
						var result string
						benchCache.Get(benchCtx, key, &result)
					} else {
						// Write
						benchCache.Set(benchCtx, key, fmt.Sprintf("value-%d", i), time.Hour)
					}
					i++
				}
			})
		})
	}
}

// BenchmarkPatternOperations benchmarks pattern-based operations
func BenchmarkPatternOperations(b *testing.B) {
	setupBench()

	// Setup test data
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("bench:pattern:user:%d", i)
		benchCache.Set(benchCtx, key, fmt.Sprintf("user-%d", i), time.Hour)
	}

	b.Run("Keys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchCache.Keys(benchCtx, "user:*")
		}
	})

	b.Run("DeletePattern", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// Re-populate keys
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("bench:pattern:temp:%d", j)
				benchCache.Set(benchCtx, key, "value", time.Hour)
			}
			b.StartTimer()

			benchCache.DeletePattern(benchCtx, "temp:*")
		}
	})
}

// BenchmarkTTLOperations benchmarks TTL-related operations
func BenchmarkTTLOperations(b *testing.B) {
	setupBench()

	key := "bench:ttl"
	benchCache.Set(benchCtx, key, "value", time.Hour)

	b.Run("TTL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchCache.TTL(benchCtx, key)
		}
	})

	b.Run("Expire", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchCache.Expire(benchCtx, key, time.Hour)
		}
	})

	b.Run("Exists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchCache.Exists(benchCtx, key)
		}
	})
}

// BenchmarkKeyBuilding benchmarks key construction overhead
func BenchmarkKeyBuilding(b *testing.B) {
	setupBench()

	b.Run("SimpleTenantKey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("simple:%d", i)
			benchCache.Set(benchCtx, key, "value", time.Hour)
		}
	})

	b.Run("TenantWithNamespace", func(b *testing.B) {
		ctx := WithNamespace(benchCtx, "users")
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("namespaced:%d", i)
			benchCache.Set(ctx, key, "value", time.Hour)
		}
	})

	b.Run("ComplexKey", func(b *testing.B) {
		ctx := WithNamespace(benchCtx, "analytics")
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("report:%s:%s:%d", "sales", "2024-01", i)
			benchCache.Set(ctx, key, "value", time.Hour)
		}
	})
}
