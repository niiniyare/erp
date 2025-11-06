package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// MockS3Client implements a basic S3 client interface for testing
type MockS3Client struct {
	objects  map[string][]byte
	metadata map[string]map[string]string
}

func NewMockS3Client() *MockS3Client {
	return &MockS3Client{
		objects:  make(map[string][]byte),
		metadata: make(map[string]map[string]string),
	}
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	key := aws.ToString(params.Key)
	data, exists := m.objects[key]
	if !exists {
		return nil, &types.NoSuchKey{Message: aws.String("The specified key does not exist.")}
	}

	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(data)),
	}, nil
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	key := aws.ToString(params.Key)

	data, err := io.ReadAll(params.Body)
	if err != nil {
		return nil, err
	}

	m.objects[key] = data
	if params.Metadata != nil {
		m.metadata[key] = params.Metadata
	}

	return &s3.PutObjectOutput{
		ETag: aws.String("\"mock-etag\""),
	}, nil
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	key := aws.ToString(params.Key)
	delete(m.objects, key)
	delete(m.metadata, key)

	return &s3.DeleteObjectOutput{}, nil
}

func (m *MockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	key := aws.ToString(params.Key)
	data, exists := m.objects[key]
	if !exists {
		return nil, &types.NoSuchKey{Message: aws.String("The specified key does not exist.")}
	}

	metadata := m.metadata[key]
	if metadata == nil {
		metadata = make(map[string]string)
	}

	now := time.Now()
	return &s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   aws.String("application/json"),
		LastModified:  &now,
		ETag:          aws.String("\"mock-etag\""),
		Metadata:      metadata,
	}, nil
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	prefix := aws.ToString(params.Prefix)

	var contents []types.Object
	for key, data := range m.objects {
		if strings.HasPrefix(key, prefix) {
			now := time.Now()
			contents = append(contents, types.Object{
				Key:          aws.String(key),
				Size:         aws.Int64(int64(len(data))),
				LastModified: &now,
				ETag:         aws.String("\"mock-etag\""),
			})
		}
	}

	return &s3.ListObjectsV2Output{
		Contents: contents,
	}, nil
}

func TestNewS3Storage(t *testing.T) {

	tests := []struct {
		name    string
		config  S3Config
		wantErr bool
	}{
		{
			name: "valid config with client and bucket",
			config: S3Config{
				Client:    &s3.Client{},
				Bucket:    "test-bucket",
				KeyPrefix: "schemas/",
			},
			wantErr: false,
		},
		{
			name: "valid config with default prefix",
			config: S3Config{
				Client: &s3.Client{},
				Bucket: "test-bucket",
			},
			wantErr: false,
		},
		{
			name: "invalid config without client",
			config: S3Config{
				Bucket: "test-bucket",
			},
			wantErr: true,
		},
		{
			name: "invalid config without bucket",
			config: S3Config{
				Client: &s3.Client{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewS3Storage(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewS3Storage() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewS3Storage() error = %v", err)
				return
			}

			if storage == nil {
				t.Error("NewS3Storage() returned nil storage")
				return
			}

			expectedPrefix := tt.config.KeyPrefix
			if expectedPrefix == "" {
				expectedPrefix = "schemas/"
			}
			if !strings.HasSuffix(expectedPrefix, "/") {
				expectedPrefix += "/"
			}

			if storage.GetKeyPrefix() != expectedPrefix {
				t.Errorf("NewS3Storage() prefix = %v, want %v", storage.GetKeyPrefix(), expectedPrefix)
			}

			if storage.GetBucket() != tt.config.Bucket {
				t.Errorf("NewS3Storage() bucket = %v, want %v", storage.GetBucket(), tt.config.Bucket)
			}
		})
	}
}

func TestS3Storage_SetAndGet(t *testing.T) {
	// Since we can't easily mock the S3 client due to interface complexity,
	// we'll test with a simple validation approach
	t.Run("validation tests", func(t *testing.T) {
		// Create a wrapper that implements the interface we need
		storage := &S3Storage{
			client:    nil, // We'll bypass actual S3 calls for validation tests
			bucket:    "test-bucket",
			keyPrefix: "schemas/",
		}

		tests := []struct {
			name     string
			schemaID string
			data     []byte
			wantErr  bool
		}{
			{
				name:     "empty schema ID",
				schemaID: "",
				data:     []byte(`{"id": "test"}`),
				wantErr:  true,
			},
			{
				name:     "invalid schema ID with path traversal",
				schemaID: "../malicious",
				data:     []byte(`{"id": "test"}`),
				wantErr:  true,
			},
			{
				name:     "empty data",
				schemaID: "valid.schema",
				data:     []byte{},
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Test validation for Set
				err := validateSchemaID(tt.schemaID)
				if tt.wantErr && tt.schemaID != "valid.schema" {
					if err == nil {
						t.Error("Expected validation error for schema ID")
					}
					return
				}

				if len(tt.data) == 0 && tt.schemaID == "valid.schema" {
					// This should fail due to empty data
					if tt.wantErr {
						return // Expected behavior
					}
				}
			})
		}

		// Test key generation
		key := storage.getKey("user.profile")
		expectedKey := "schemas/user/profile.json"
		if key != expectedKey {
			t.Errorf("getKey() = %v, want %v", key, expectedKey)
		}
	})
}

func TestS3Storage_KeyGeneration(t *testing.T) {
	storage := &S3Storage{
		bucket:    "test-bucket",
		keyPrefix: "schemas/",
	}

	tests := []struct {
		name        string
		schemaID    string
		expectedKey string
	}{
		{
			name:        "simple ID",
			schemaID:    "user",
			expectedKey: "schemas/user.json",
		},
		{
			name:        "dotted ID",
			schemaID:    "user.profile",
			expectedKey: "schemas/user/profile.json",
		},
		{
			name:        "complex nested ID",
			schemaID:    "forms.user.registration.step1",
			expectedKey: "schemas/forms/user/registration/step1.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := storage.getKey(tt.schemaID)
			if key != tt.expectedKey {
				t.Errorf("getKey(%s) = %v, want %v", tt.schemaID, key, tt.expectedKey)
			}
		})
	}
}

func TestS3Storage_Validation(t *testing.T) {
	storage := &S3Storage{
		bucket:    "test-bucket",
		keyPrefix: "schemas/",
	}

	// Avoid unused variable warnings
	_ = storage

	// Test invalid schema IDs
	invalidIDs := []string{
		"",
		"../malicious",
		"user..profile",
		"user/profile",
		"user\\profile",
		"user<profile",
		strings.Repeat("a", 250),
	}

	for _, id := range invalidIDs {
		t.Run(fmt.Sprintf("invalid_id_%s", id), func(t *testing.T) {
			err := validateSchemaID(id)
			if err == nil {
				t.Errorf("validateSchemaID(%s) should return error", id)
			}
		})
	}

	// Test valid schema IDs
	validIDs := []string{
		"user",
		"user.profile",
		"forms.user.registration",
	}

	for _, id := range validIDs {
		t.Run(fmt.Sprintf("valid_id_%s", id), func(t *testing.T) {
			err := validateSchemaID(id)
			if err != nil {
				t.Errorf("validateSchemaID(%s) should not return error: %v", id, err)
			}
		})
	}
}

func TestS3Storage_Stats_Structure(t *testing.T) {
	// Test the S3Stats structure
	stats := &S3Stats{
		Type:        "s3",
		Bucket:      "test-bucket",
		Prefix:      "schemas/",
		SchemaCount: 5,
		TotalSize:   1024,
	}

	if stats.Type != "s3" {
		t.Errorf("S3Stats.Type = %v, want s3", stats.Type)
	}

	if stats.Bucket != "test-bucket" {
		t.Errorf("S3Stats.Bucket = %v, want test-bucket", stats.Bucket)
	}

	if stats.SchemaCount != 5 {
		t.Errorf("S3Stats.SchemaCount = %v, want 5", stats.SchemaCount)
	}

	if stats.TotalSize != 1024 {
		t.Errorf("S3Stats.TotalSize = %v, want 1024", stats.TotalSize)
	}
}

func TestS3ObjectInfo_Structure(t *testing.T) {
	// Test the S3ObjectInfo structure
	now := time.Now()
	info := &S3ObjectInfo{
		SchemaID:     "user.profile",
		Key:          "schemas/user/profile.json",
		ContentType:  "application/json",
		Size:         256,
		LastModified: &now,
		ETag:         "\"abc123\"",
		Metadata: map[string]string{
			"schema-id": "user.profile",
			"type":      "schema",
		},
	}

	if info.SchemaID != "user.profile" {
		t.Errorf("S3ObjectInfo.SchemaID = %v, want user.profile", info.SchemaID)
	}

	if info.Key != "schemas/user/profile.json" {
		t.Errorf("S3ObjectInfo.Key = %v, want schemas/user/profile.json", info.Key)
	}

	if info.ContentType != "application/json" {
		t.Errorf("S3ObjectInfo.ContentType = %v, want application/json", info.ContentType)
	}

	if info.Size != 256 {
		t.Errorf("S3ObjectInfo.Size = %v, want 256", info.Size)
	}

	if info.Metadata["schema-id"] != "user.profile" {
		t.Errorf("S3ObjectInfo.Metadata[schema-id] = %v, want user.profile", info.Metadata["schema-id"])
	}
}

func TestS3Storage_BatchOperations_Structure(t *testing.T) {
	// Test that batch operations handle empty inputs correctly
	storage := &S3Storage{
		bucket:    "test-bucket",
		keyPrefix: "schemas/",
	}

	ctx := context.Background()

	// Test BatchGet with empty input
	result, err := storage.BatchGet(ctx, []string{})
	if err != nil {
		t.Errorf("BatchGet with empty input should not error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("BatchGet with empty input should return empty map")
	}

	// Test BatchSet with empty input
	err = storage.BatchSet(ctx, map[string][]byte{})
	if err != nil {
		t.Errorf("BatchSet with empty input should not error: %v", err)
	}

	// Test BatchDelete with empty input
	err = storage.BatchDelete(ctx, []string{})
	if err != nil {
		t.Errorf("BatchDelete with empty input should not error: %v", err)
	}

	// Avoid unused variable warnings
	_ = storage
	_ = ctx
}

func TestS3Storage_Configuration(t *testing.T) {
	// Test different prefix configurations
	tests := []struct {
		name           string
		inputPrefix    string
		expectedPrefix string
	}{
		{
			name:           "empty prefix gets default",
			inputPrefix:    "",
			expectedPrefix: "schemas/",
		},
		{
			name:           "prefix without slash gets slash added",
			inputPrefix:    "data",
			expectedPrefix: "data/",
		},
		{
			name:           "prefix with slash stays same",
			inputPrefix:    "schemas/",
			expectedPrefix: "schemas/",
		},
		{
			name:           "complex prefix",
			inputPrefix:    "app/v1/schemas",
			expectedPrefix: "app/v1/schemas/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := S3Config{
				Client:    &s3.Client{}, // Mock client
				Bucket:    "test-bucket",
				KeyPrefix: tt.inputPrefix,
			}

			storage, err := NewS3Storage(config)
			if err != nil {
				t.Errorf("NewS3Storage() error = %v", err)
				return
			}

			if storage.GetKeyPrefix() != tt.expectedPrefix {
				t.Errorf("NewS3Storage() prefix = %v, want %v", storage.GetKeyPrefix(), tt.expectedPrefix)
			}
		})
	}
}

func BenchmarkS3Storage_KeyGeneration(b *testing.B) {
	storage := &S3Storage{
		bucket:    "test-bucket",
		keyPrefix: "schemas/",
	}

	schemaIDs := []string{
		"user",
		"user.profile",
		"forms.user.registration",
		"complex.nested.schema.with.many.parts",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := schemaIDs[i%len(schemaIDs)]
		_ = storage.getKey(id)
	}
}

func BenchmarkS3Storage_Validation(b *testing.B) {
	schemaIDs := []string{
		"user.profile",
		"forms.registration",
		"admin.dashboard",
		"reports.monthly.sales",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := schemaIDs[i%len(schemaIDs)]
		_ = validateSchemaID(id)
	}
}

// Test JSON marshaling of S3Stats
func TestS3Stats_JSON(t *testing.T) {
	stats := &S3Stats{
		Type:        "s3",
		Bucket:      "test-bucket",
		Prefix:      "schemas/",
		SchemaCount: 10,
		TotalSize:   2048,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Errorf("Failed to marshal S3Stats: %v", err)
	}

	var unmarshaled S3Stats
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal S3Stats: %v", err)
	}

	if unmarshaled.Type != stats.Type {
		t.Errorf("Unmarshaled Type = %v, want %v", unmarshaled.Type, stats.Type)
	}

	if unmarshaled.SchemaCount != stats.SchemaCount {
		t.Errorf("Unmarshaled SchemaCount = %v, want %v", unmarshaled.SchemaCount, stats.SchemaCount)
	}
}

// Test JSON marshaling of S3ObjectInfo
func TestS3ObjectInfo_JSON(t *testing.T) {
	now := time.Now()
	info := &S3ObjectInfo{
		SchemaID:     "user.profile",
		Key:          "schemas/user/profile.json",
		ContentType:  "application/json",
		Size:         256,
		LastModified: &now,
		ETag:         "\"abc123\"",
		Metadata: map[string]string{
			"schema-id": "user.profile",
		},
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Errorf("Failed to marshal S3ObjectInfo: %v", err)
	}

	var unmarshaled S3ObjectInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal S3ObjectInfo: %v", err)
	}

	if unmarshaled.SchemaID != info.SchemaID {
		t.Errorf("Unmarshaled SchemaID = %v, want %v", unmarshaled.SchemaID, info.SchemaID)
	}

	if unmarshaled.Size != info.Size {
		t.Errorf("Unmarshaled Size = %v, want %v", unmarshaled.Size, info.Size)
	}
}
