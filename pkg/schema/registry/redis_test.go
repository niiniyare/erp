package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

func TestNewRedisStorage(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	tests := []struct {
		name    string
		config  RedisConfig
		wantErr bool
	}{
		{
			name: "valid config with client",
			config: RedisConfig{
				Client: client,
				Prefix: "test:",
				TTL:    time.Hour,
			},
			wantErr: false,
		},
		{
			name: "valid config with default prefix",
			config: RedisConfig{
				Client: client,
			},
			wantErr: false,
		},
		{
			name: "invalid config without client",
			config: RedisConfig{
				Prefix: "test:",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewRedisStorage(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewRedisStorage() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewRedisStorage() error = %v", err)
				return
			}

			if storage == nil {
				t.Error("NewRedisStorage() returned nil storage")
				return
			}

			expectedPrefix := tt.config.Prefix
			if expectedPrefix == "" {
				expectedPrefix = "schema:"
			}

			if storage.GetPrefix() != expectedPrefix {
				t.Errorf("NewRedisStorage() prefix = %v, want %v", storage.GetPrefix(), expectedPrefix)
			}

			if storage.GetTTL() != tt.config.TTL {
				t.Errorf("NewRedisStorage() TTL = %v, want %v", storage.GetTTL(), tt.config.TTL)
			}
		})
	}
}

func TestRedisStorage_SetAndGet(t *testing.T) {
	client, mock := redismock.NewClientMock()
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "test:",
	})
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()
	testSchema := map[string]any{
		"id":    "user.profile",
		"title": "User Profile",
		"fields": []map[string]any{
			{
				"name": "username",
				"type": "text",
			},
		},
	}

	schemaData, err := json.Marshal(testSchema)
	if err != nil {
		t.Fatalf("Failed to marshal test schema: %v", err)
	}

	tests := []struct {
		name     string
		schemaID string
		data     []byte
		wantErr  bool
	}{
		{
			name:     "valid schema",
			schemaID: "user.profile",
			data:     schemaData,
			wantErr:  false,
		},
		{
			name:     "nested schema ID",
			schemaID: "forms.user.registration",
			data:     schemaData,
			wantErr:  false,
		},
		{
			name:     "empty schema ID",
			schemaID: "",
			data:     schemaData,
			wantErr:  true,
		},
		{
			name:     "invalid schema ID with path traversal",
			schemaID: "../malicious",
			data:     schemaData,
			wantErr:  true,
		},
		{
			name:     "empty data",
			schemaID: "empty.schema",
			data:     []byte{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				// Mock SET command
				mock.ExpectSet("test:"+tt.schemaID, tt.data, 0).SetVal("OK")
				// Mock GET command
				mock.ExpectGet("test:" + tt.schemaID).SetVal(string(tt.data))
			}

			// Test Set
			err := storage.Set(ctx, tt.schemaID, tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("Set() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Set() error = %v", err)
				return
			}

			// Test Get
			retrievedData, err := storage.Get(ctx, tt.schemaID)
			if err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}

			if string(retrievedData) != string(tt.data) {
				t.Errorf("Get() data mismatch, got %s, want %s", string(retrievedData), string(tt.data))
			}

			// Verify all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Redis mock expectations not met: %v", err)
			}
		})
	}
}

func TestRedisStorage_Delete(t *testing.T) {
	client, mock := redismock.NewClientMock()
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "test:",
	})
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()
	testData := []byte(`{"id": "test.delete", "title": "Test Delete"}`)

	// Mock SET for initial setup
	mock.ExpectSet("test:test.delete", testData, 0).SetVal("OK")
	// Mock EXISTS for checking existence
	mock.ExpectExists("test:test.delete").SetVal(1)
	// Mock DEL for deletion (returns 1 for successful deletion)
	mock.ExpectDel("test:test.delete").SetVal(1)
	// Mock EXISTS again for verification
	mock.ExpectExists("test:test.delete").SetVal(0)
	// Mock DEL for non-existent schema (returns 0)
	mock.ExpectDel("test:non.existent").SetVal(0)

	// First, store a schema
	err = storage.Set(ctx, "test.delete", testData)
	if err != nil {
		t.Fatalf("Failed to set test schema: %v", err)
	}

	// Verify it exists
	exists, err := storage.Exists(ctx, "test.delete")
	if err != nil {
		t.Fatalf("Failed to check if schema exists: %v", err)
	}
	if !exists {
		t.Fatal("Schema should exist before deletion")
	}

	// Delete it
	err = storage.Delete(ctx, "test.delete")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify it no longer exists
	exists, err = storage.Exists(ctx, "test.delete")
	if err != nil {
		t.Errorf("Failed to check if schema exists after deletion: %v", err)
	}
	if exists {
		t.Error("Schema should not exist after deletion")
	}

	// Try to delete non-existent schema
	err = storage.Delete(ctx, "non.existent")
	if err == nil {
		t.Error("Delete() should return error for non-existent schema")
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Redis mock expectations not met: %v", err)
	}
}

func TestRedisStorage_List(t *testing.T) {
	client, mock := redismock.NewClientMock()
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "test:",
	})
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()

	// Mock SCAN command to return schema keys
	expectedKeys := []string{
		"test:user.profile",
		"test:user.settings",
		"test:forms.registration",
		"test:admin.dashboard",
	}

	mock.ExpectScan(0, "test:*", 0).SetVal(expectedKeys, 0)

	// List all schemas
	schemaIDs, err := storage.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
		return
	}

	expectedIDs := []string{
		"user.profile",
		"user.settings",
		"forms.registration",
		"admin.dashboard",
	}

	// Verify all schemas are listed
	if len(schemaIDs) != len(expectedIDs) {
		t.Errorf("List() returned %d schemas, want %d", len(schemaIDs), len(expectedIDs))
	}

	// Check each schema ID is present
	schemaIDMap := make(map[string]bool)
	for _, id := range schemaIDs {
		schemaIDMap[id] = true
	}

	for _, expectedID := range expectedIDs {
		if !schemaIDMap[expectedID] {
			t.Errorf("List() missing schema ID: %s", expectedID)
		}
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Redis mock expectations not met: %v", err)
	}
}

func TestRedisStorage_Exists(t *testing.T) {
	client, mock := redismock.NewClientMock()
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "test:",
	})
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()

	// Mock EXISTS commands
	mock.ExpectExists("test:non.existent").SetVal(0)
	mock.ExpectSet("test:test.exists", []byte(`{"id": "test.exists"}`), 0).SetVal("OK")
	mock.ExpectExists("test:test.exists").SetVal(1)

	// Test non-existent schema
	exists, err := storage.Exists(ctx, "non.existent")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() should return false for non-existent schema")
	}

	// Store a schema
	testData := []byte(`{"id": "test.exists"}`)
	err = storage.Set(ctx, "test.exists", testData)
	if err != nil {
		t.Fatalf("Failed to set test schema: %v", err)
	}

	// Test existing schema
	exists, err = storage.Exists(ctx, "test.exists")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() should return true for existing schema")
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Redis mock expectations not met: %v", err)
	}
}

func TestRedisStorage_TTL(t *testing.T) {
	client, mock := redismock.NewClientMock()
	ttl := time.Hour
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "test:",
		TTL:    ttl,
	})
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()
	testData := []byte(`{"id": "test.ttl"}`)

	// Mock SET command with TTL
	mock.ExpectSet("test:test.ttl", testData, ttl).SetVal("OK")

	err = storage.Set(ctx, "test.ttl", testData)
	if err != nil {
		t.Errorf("Set() with TTL error = %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Redis mock expectations not met: %v", err)
	}
}

func TestRedisStorage_Stats(t *testing.T) {
	client, mock := redismock.NewClientMock()
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "test:",
		TTL:    time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()

	// Mock SCAN command for counting schemas
	expectedKeys := []string{
		"test:schema1",
		"test:schema2",
		"test:schema3",
	}
	mock.ExpectScan(0, "test:*", 0).SetVal(expectedKeys, 0)

	// Mock INFO command for memory usage
	infoResponse := "# Memory\r\nused_memory:1048576\r\nused_memory_human:1.00M\r\n"
	mock.ExpectInfo("memory").SetVal(infoResponse)

	stats, err := storage.Stats(ctx)
	if err != nil {
		t.Errorf("Stats() error = %v", err)
		return
	}

	if stats.Type != "redis" {
		t.Errorf("Stats() type = %s, want redis", stats.Type)
	}
	if stats.SchemaCount != len(expectedKeys) {
		t.Errorf("Stats() schema count = %d, want %d", stats.SchemaCount, len(expectedKeys))
	}
	if stats.UsedMemory != 1048576 {
		t.Errorf("Stats() used memory = %d, want 1048576", stats.UsedMemory)
	}
	if stats.TTL != time.Hour {
		t.Errorf("Stats() TTL = %v, want %v", stats.TTL, time.Hour)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Redis mock expectations not met: %v", err)
	}
}

func TestRedisStorage_GetWithTTL(t *testing.T) {
	// Skip this test for now due to pipeline mocking complexity
	t.Skip("Skipping pipeline test due to mock limitations")
}

func TestRedisStorage_BatchOperations(t *testing.T) {
	// Skip this test for now due to pipeline mocking complexity
	t.Skip("Skipping batch operations test due to mock limitations")
}

func TestRedisStorage_Ping(t *testing.T) {
	client, mock := redismock.NewClientMock()
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "test:",
	})
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()

	// Mock PING command
	mock.ExpectPing().SetVal("PONG")

	err = storage.Ping(ctx)
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Redis mock expectations not met: %v", err)
	}
}

func TestRedisStorage_FlushAll(t *testing.T) {
	client, mock := redismock.NewClientMock()
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "test:",
	})
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()

	// Mock SCAN and DEL commands
	keys := []string{"test:schema1", "test:schema2", "test:schema3"}
	mock.ExpectScan(0, "test:*", 0).SetVal(keys, 0)
	mock.ExpectDel(keys...).SetVal(int64(len(keys)))

	err = storage.FlushAll(ctx)
	if err != nil {
		t.Errorf("FlushAll() error = %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Redis mock expectations not met: %v", err)
	}
}

func BenchmarkRedisStorage_Set(b *testing.B) {
	client, mock := redismock.NewClientMock()
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "bench:",
	})
	if err != nil {
		b.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()
	testData := []byte(`{"id": "benchmark", "title": "Benchmark Schema", "fields": [{"name": "test", "type": "text"}]}`)

	// Mock all SET operations
	for i := 0; i < b.N; i++ {
		mock.ExpectSet(fmt.Sprintf("bench:benchmark.schema.%d", i), testData, 0).SetVal("OK")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schemaID := fmt.Sprintf("benchmark.schema.%d", i)
		err := storage.Set(ctx, schemaID, testData)
		if err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkRedisStorage_Get(b *testing.B) {
	client, mock := redismock.NewClientMock()
	storage, err := NewRedisStorage(RedisConfig{
		Client: client,
		Prefix: "bench:",
	})
	if err != nil {
		b.Fatalf("Failed to create Redis storage: %v", err)
	}

	ctx := context.Background()
	testData := []byte(`{"id": "benchmark", "title": "Benchmark Schema"}`)

	// Mock all GET operations
	for i := 0; i < b.N; i++ {
		mock.ExpectGet(fmt.Sprintf("bench:benchmark.schema.%d", i)).SetVal(string(testData))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schemaID := fmt.Sprintf("benchmark.schema.%d", i)
		_, err := storage.Get(ctx, schemaID)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}
