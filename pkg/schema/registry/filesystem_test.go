package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewFilesystemStorage(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid directory",
			path:    tempDir,
			wantErr: false,
		},
		{
			name:    "new directory creation",
			path:    filepath.Join(tempDir, "new_dir"),
			wantErr: false,
		},
		{
			name:    "invalid path - file exists",
			path:    createTempFile(t, tempDir),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewFilesystemStorage(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("NewFilesystemStorage() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewFilesystemStorage() error = %v", err)
				return
			}

			if storage == nil {
				t.Error("NewFilesystemStorage() returned nil storage")
				return
			}

			if storage.GetBasePath() != tt.path {
				t.Errorf("NewFilesystemStorage() basePath = %v, want %v", storage.GetBasePath(), tt.path)
			}

			// Verify directory was created
			if _, err := os.Stat(tt.path); os.IsNotExist(err) {
				t.Errorf("NewFilesystemStorage() failed to create directory %s", tt.path)
			}
		})
	}
}

func TestFilesystemStorage_SetAndGet(t *testing.T) {
	storage, tempDir := setupFilesystemTest(t)
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

			// Verify file structure
			expectedPath := storage.getFilePath(tt.schemaID)
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Expected file %s was not created", expectedPath)
			}
		})
	}

	// Verify directory structure for nested schemas
	nestedPath := filepath.Join(tempDir, "forms", "user", "registration.json")
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Nested directory structure was not created correctly")
	}
}

func TestFilesystemStorage_Delete(t *testing.T) {
	storage, _ := setupFilesystemTest(t)
	ctx := context.Background()

	// First, store a schema
	testData := []byte(`{"id": "test.delete", "title": "Test Delete"}`)
	err := storage.Set(ctx, "test.delete", testData)
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
}

func TestFilesystemStorage_List(t *testing.T) {
	storage, _ := setupFilesystemTest(t)
	ctx := context.Background()

	// Store multiple schemas
	testSchemas := map[string][]byte{
		"user.profile":       []byte(`{"id": "user.profile"}`),
		"user.settings":      []byte(`{"id": "user.settings"}`),
		"forms.registration": []byte(`{"id": "forms.registration"}`),
		"admin.dashboard":    []byte(`{"id": "admin.dashboard"}`),
	}

	for id, data := range testSchemas {
		err := storage.Set(ctx, id, data)
		if err != nil {
			t.Fatalf("Failed to set schema %s: %v", id, err)
		}
	}

	// List all schemas
	schemaIDs, err := storage.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
		return
	}

	// Verify all schemas are listed
	if len(schemaIDs) != len(testSchemas) {
		t.Errorf("List() returned %d schemas, want %d", len(schemaIDs), len(testSchemas))
	}

	// Check each schema ID is present
	schemaIDMap := make(map[string]bool)
	for _, id := range schemaIDs {
		schemaIDMap[id] = true
	}

	for expectedID := range testSchemas {
		if !schemaIDMap[expectedID] {
			t.Errorf("List() missing schema ID: %s", expectedID)
		}
	}
}

func TestFilesystemStorage_Exists(t *testing.T) {
	storage, _ := setupFilesystemTest(t)
	ctx := context.Background()

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
}

func TestFilesystemStorage_Stats(t *testing.T) {
	storage, _ := setupFilesystemTest(t)
	ctx := context.Background()

	// Test empty storage
	stats, err := storage.Stats(ctx)
	if err != nil {
		t.Errorf("Stats() error = %v", err)
		return
	}

	if stats.Type != "filesystem" {
		t.Errorf("Stats() type = %s, want filesystem", stats.Type)
	}
	if stats.SchemaCount != 0 {
		t.Errorf("Stats() schema count = %d, want 0", stats.SchemaCount)
	}
	if stats.TotalSize != 0 {
		t.Errorf("Stats() total size = %d, want 0", stats.TotalSize)
	}

	// Add some schemas
	testSchemas := map[string][]byte{
		"schema.one":   []byte(`{"id": "schema.one", "title": "Schema One"}`),
		"schema.two":   []byte(`{"id": "schema.two", "title": "Schema Two"}`),
		"schema.three": []byte(`{"id": "schema.three", "title": "Schema Three"}`),
	}

	for id, data := range testSchemas {
		err := storage.Set(ctx, id, data)
		if err != nil {
			t.Fatalf("Failed to set schema %s: %v", id, err)
		}
	}

	// Test stats with data
	stats, err = storage.Stats(ctx)
	if err != nil {
		t.Errorf("Stats() error = %v", err)
		return
	}

	if stats.SchemaCount != len(testSchemas) {
		t.Errorf("Stats() schema count = %d, want %d", stats.SchemaCount, len(testSchemas))
	}
	if stats.TotalSize <= 0 {
		t.Errorf("Stats() total size = %d, should be > 0", stats.TotalSize)
	}
}

func TestFilesystemStorage_ContextCancellation(t *testing.T) {
	storage, _ := setupFilesystemTest(t)

	// Test context cancellation for various operations
	testData := []byte(`{"id": "test.context"}`)

	t.Run("cancelled context for Set", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := storage.Set(ctx, "test.cancelled", testData)
		if err == nil {
			t.Error("Set() should return error for cancelled context")
		}
		if err != context.Canceled {
			t.Errorf("Set() error = %v, want %v", err, context.Canceled)
		}
	})

	t.Run("cancelled context for Get", func(t *testing.T) {
		// First set a schema with valid context
		err := storage.Set(context.Background(), "test.context", testData)
		if err != nil {
			t.Fatalf("Failed to set test schema: %v", err)
		}

		// Then try to get with cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = storage.Get(ctx, "test.context")
		if err == nil {
			t.Error("Get() should return error for cancelled context")
		}
		if err != context.Canceled {
			t.Errorf("Get() error = %v, want %v", err, context.Canceled)
		}
	})
}

func TestFilesystemStorage_AtomicWrites(t *testing.T) {
	storage, tempDir := setupFilesystemTest(t)
	ctx := context.Background()

	schemaID := "test.atomic"
	initialData := []byte(`{"id": "test.atomic", "version": 1}`)
	updatedData := []byte(`{"id": "test.atomic", "version": 2}`)

	// Set initial data
	err := storage.Set(ctx, schemaID, initialData)
	if err != nil {
		t.Fatalf("Failed to set initial data: %v", err)
	}

	// Verify no temporary files exist
	files, err := filepath.Glob(filepath.Join(tempDir, "*.tmp"))
	if err != nil {
		t.Fatalf("Failed to check for temp files: %v", err)
	}
	if len(files) > 0 {
		t.Errorf("Found unexpected temporary files: %v", files)
	}

	// Update data
	err = storage.Set(ctx, schemaID, updatedData)
	if err != nil {
		t.Fatalf("Failed to update data: %v", err)
	}

	// Verify updated data is correct
	retrievedData, err := storage.Get(ctx, schemaID)
	if err != nil {
		t.Fatalf("Failed to get updated data: %v", err)
	}

	if string(retrievedData) != string(updatedData) {
		t.Errorf("Retrieved data = %s, want %s", string(retrievedData), string(updatedData))
	}

	// Verify no temporary files remain
	files, err = filepath.Glob(filepath.Join(tempDir, "*.tmp"))
	if err != nil {
		t.Fatalf("Failed to check for temp files: %v", err)
	}
	if len(files) > 0 {
		t.Errorf("Found temporary files after atomic write: %v", files)
	}
}

func TestValidateSchemaID(t *testing.T) {
	tests := []struct {
		name     string
		schemaID string
		wantErr  bool
	}{
		{
			name:     "valid simple ID",
			schemaID: "user",
			wantErr:  false,
		},
		{
			name:     "valid dotted ID",
			schemaID: "user.profile",
			wantErr:  false,
		},
		{
			name:     "valid complex ID",
			schemaID: "forms.user.registration.step1",
			wantErr:  false,
		},
		{
			name:     "empty ID",
			schemaID: "",
			wantErr:  true,
		},
		{
			name:     "ID with path traversal",
			schemaID: "../malicious",
			wantErr:  true,
		},
		{
			name:     "ID with double dots",
			schemaID: "user..profile",
			wantErr:  true,
		},
		{
			name:     "ID with forward slash",
			schemaID: "user/profile",
			wantErr:  true,
		},
		{
			name:     "ID with backslash",
			schemaID: "user\\profile",
			wantErr:  true,
		},
		{
			name:     "ID with invalid character",
			schemaID: "user<profile",
			wantErr:  true,
		},
		{
			name:     "very long ID",
			schemaID: strings.Repeat("a", 250),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchemaID(tt.schemaID)
			if tt.wantErr {
				if err == nil {
					t.Error("validateSchemaID() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validateSchemaID() error = %v", err)
				}
			}
		})
	}
}

func TestFilesystemStorage_ConcurrentAccess(t *testing.T) {
	storage, _ := setupFilesystemTest(t)
	ctx := context.Background()

	// Test concurrent reads and writes
	numGoroutines := 10
	numOperations := 5

	errChan := make(chan error, numGoroutines*numOperations)
	done := make(chan bool, numGoroutines)

	// Start multiple goroutines performing operations
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				schemaID := fmt.Sprintf("worker%d.schema%d", workerID, j)
				testData := []byte(fmt.Sprintf(`{"id": "%s", "worker": %d}`, schemaID, workerID))

				// Set
				if err := storage.Set(ctx, schemaID, testData); err != nil {
					errChan <- err
					return
				}

				// Get
				if _, err := storage.Get(ctx, schemaID); err != nil {
					errChan <- err
					return
				}

				// Exists
				if _, err := storage.Exists(ctx, schemaID); err != nil {
					errChan <- err
					return
				}

				// Brief pause to increase chance of concurrent access
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	close(errChan)
	for err := range errChan {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Verify final state
	schemaIDs, err := storage.List(ctx)
	if err != nil {
		t.Errorf("Failed to list schemas after concurrent operations: %v", err)
	}

	expectedCount := numGoroutines * numOperations
	if len(schemaIDs) != expectedCount {
		t.Errorf("Expected %d schemas after concurrent operations, got %d", expectedCount, len(schemaIDs))
	}
}

// Helper functions

func setupFilesystemTest(t *testing.T) (*FilesystemStorage, string) {
	tempDir := t.TempDir()

	storage, err := NewFilesystemStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create filesystem storage: %v", err)
	}

	return storage, tempDir
}

func createTempFile(t *testing.T, dir string) string {
	file, err := os.CreateTemp(dir, "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()
	return file.Name()
}

func BenchmarkFilesystemStorage_Set(b *testing.B) {
	storage, _ := setupFilesystemBenchmark(b)
	ctx := context.Background()
	testData := []byte(`{"id": "benchmark", "title": "Benchmark Schema", "fields": [{"name": "test", "type": "text"}]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schemaID := fmt.Sprintf("benchmark.schema.%d", i)
		err := storage.Set(ctx, schemaID, testData)
		if err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkFilesystemStorage_Get(b *testing.B) {
	storage, _ := setupFilesystemBenchmark(b)
	ctx := context.Background()
	testData := []byte(`{"id": "benchmark", "title": "Benchmark Schema"}`)

	// Pre-populate with schemas
	for i := 0; i < b.N; i++ {
		schemaID := fmt.Sprintf("benchmark.schema.%d", i)
		storage.Set(ctx, schemaID, testData)
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

func setupFilesystemBenchmark(b *testing.B) (*FilesystemStorage, string) {
	tempDir := b.TempDir()

	storage, err := NewFilesystemStorage(tempDir)
	if err != nil {
		b.Fatalf("Failed to create filesystem storage: %v", err)
	}

	return storage, tempDir
}
