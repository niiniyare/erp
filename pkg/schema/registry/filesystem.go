package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FilesystemStorage implements Storage interface using local filesystem
type FilesystemStorage struct {
	basePath string
}

// NewFilesystemStorage creates a new filesystem storage backend
func NewFilesystemStorage(basePath string) (*FilesystemStorage, error) {
	// Ensure the base directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory %s: %w", basePath, err)
	}

	// Verify the directory is writable
	testFile := filepath.Join(basePath, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return nil, fmt.Errorf("base directory %s is not writable: %w", basePath, err)
	}
	os.Remove(testFile) // Clean up test file

	return &FilesystemStorage{
		basePath: basePath,
	}, nil
}

// Get retrieves a schema by ID from the filesystem
func (fs *FilesystemStorage) Get(ctx context.Context, id string) ([]byte, error) {
	if err := validateSchemaID(id); err != nil {
		return nil, err
	}

	filePath := fs.getFilePath(id)

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("schema %s not found", id)
		}
		return nil, fmt.Errorf("failed to read schema %s: %w", id, err)
	}

	return data, nil
}

// Set stores a schema by ID to the filesystem
func (fs *FilesystemStorage) Set(ctx context.Context, id string, data []byte) error {
	if err := validateSchemaID(id); err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("cannot store empty schema data for %s", id)
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filePath := fs.getFilePath(id)

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for schema %s: %w", id, err)
	}

	// Write to temporary file first for atomic operation
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file for schema %s: %w", id, err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file on failure
		return fmt.Errorf("failed to rename temporary file for schema %s: %w", id, err)
	}

	return nil
}

// Delete removes a schema by ID from the filesystem
func (fs *FilesystemStorage) Delete(ctx context.Context, id string) error {
	if err := validateSchemaID(id); err != nil {
		return err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filePath := fs.getFilePath(id)

	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("schema %s not found", id)
		}
		return fmt.Errorf("failed to delete schema %s: %w", id, err)
	}

	return nil
}

// List returns all schema IDs from the filesystem
func (fs *FilesystemStorage) List(ctx context.Context) ([]string, error) {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var schemaIDs []string

	err := filepath.WalkDir(fs.basePath, func(path string, d os.DirEntry, err error) error {
		// Check context cancellation during walk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		// Skip directories and non-JSON files
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		// Convert file path back to schema ID
		relPath, err := filepath.Rel(fs.basePath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// Remove .json extension and convert path separators to dots
		schemaID := strings.TrimSuffix(relPath, ".json")
		schemaID = strings.ReplaceAll(schemaID, string(filepath.Separator), ".")

		schemaIDs = append(schemaIDs, schemaID)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}

	return schemaIDs, nil
}

// Exists checks if a schema exists in the filesystem
func (fs *FilesystemStorage) Exists(ctx context.Context, id string) (bool, error) {
	if err := validateSchemaID(id); err != nil {
		return false, err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	filePath := fs.getFilePath(id)

	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if schema %s exists: %w", id, err)
	}

	return true, nil
}

// getFilePath converts schema ID to file path
// Schema IDs like "user.profile" become "user/profile.json"
func (fs *FilesystemStorage) getFilePath(id string) string {
	// Replace dots with path separators for nested structure
	pathParts := strings.Split(id, ".")
	pathParts = append(pathParts[:len(pathParts)-1], pathParts[len(pathParts)-1]+".json")
	return filepath.Join(append([]string{fs.basePath}, pathParts...)...)
}

// validateSchemaID validates that the schema ID is safe for filesystem use
func validateSchemaID(id string) error {
	if id == "" {
		return fmt.Errorf("schema ID cannot be empty")
	}

	// Check for invalid characters that could cause path traversal
	if strings.Contains(id, "..") {
		return fmt.Errorf("schema ID cannot contain '..'")
	}

	if strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return fmt.Errorf("schema ID cannot contain path separators")
	}

	// Check for other potentially dangerous characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(id, char) {
			return fmt.Errorf("schema ID cannot contain character: %s", char)
		}
	}

	// Ensure it's not too long (filesystem limits)
	if len(id) > 200 {
		return fmt.Errorf("schema ID too long (max 200 characters)")
	}

	return nil
}

// GetBasePath returns the base path used by this storage
func (fs *FilesystemStorage) GetBasePath() string {
	return fs.basePath
}

// Stats returns storage statistics
func (fs *FilesystemStorage) Stats(ctx context.Context) (*StorageStats, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	stats := &StorageStats{
		Type: "filesystem",
	}

	// Count schemas and calculate total size
	err := filepath.WalkDir(fs.basePath, func(path string, d os.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			stats.SchemaCount++

			if info, err := d.Info(); err == nil {
				stats.TotalSize += info.Size()
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to calculate storage stats: %w", err)
	}

	return stats, nil
}

// StorageStats represents filesystem storage statistics
type StorageStats struct {
	Type        string `json:"type"`
	SchemaCount int    `json:"schema_count"`
	TotalSize   int64  `json:"total_size_bytes"`
}
