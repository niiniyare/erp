package registry

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage implements Storage interface using AWS S3
type S3Storage struct {
	client    *s3.Client
	bucket    string
	keyPrefix string
}

// S3Config configures S3 storage
type S3Config struct {
	Client    *s3.Client // AWS S3 client
	Bucket    string     // S3 bucket name
	KeyPrefix string     // Optional key prefix for schemas (default: "schemas/")
}

// NewS3Storage creates a new S3 storage backend
func NewS3Storage(config S3Config) (*S3Storage, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("S3 client is required")
	}

	if config.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}

	if config.KeyPrefix == "" {
		config.KeyPrefix = "schemas/"
	}

	// Ensure prefix ends with /
	if !strings.HasSuffix(config.KeyPrefix, "/") {
		config.KeyPrefix += "/"
	}

	return &S3Storage{
		client:    config.Client,
		bucket:    config.Bucket,
		keyPrefix: config.KeyPrefix,
	}, nil
}

// Get retrieves a schema by ID from S3
func (s3s *S3Storage) Get(ctx context.Context, id string) ([]byte, error) {
	if err := validateSchemaID(id); err != nil {
		return nil, err
	}

	key := s3s.getKey(id)

	result, err := s3s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("schema %s not found", id)
		}
		return nil, fmt.Errorf("failed to get schema %s from S3: %w", id, err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema %s from S3: %w", id, err)
	}

	return data, nil
}

// Set stores a schema by ID to S3
func (s3s *S3Storage) Set(ctx context.Context, id string, data []byte) error {
	if err := validateSchemaID(id); err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("cannot store empty schema data for %s", id)
	}

	key := s3s.getKey(id)

	_, err := s3s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
		Metadata: map[string]string{
			"schema-id": id,
			"type":      "schema",
		},
	})

	if err != nil {
		return fmt.Errorf("failed to store schema %s in S3: %w", id, err)
	}

	return nil
}

// Delete removes a schema by ID from S3
func (s3s *S3Storage) Delete(ctx context.Context, id string) error {
	if err := validateSchemaID(id); err != nil {
		return err
	}

	key := s3s.getKey(id)

	// Check if object exists first
	_, err := s3s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return fmt.Errorf("schema %s not found", id)
		}
		return fmt.Errorf("failed to check schema %s in S3: %w", id, err)
	}

	// Delete the object
	_, err = s3s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete schema %s from S3: %w", id, err)
	}

	return nil
}

// List returns all schema IDs from S3
func (s3s *S3Storage) List(ctx context.Context) ([]string, error) {
	var schemaIDs []string

	paginator := s3.NewListObjectsV2Paginator(s3s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3s.bucket),
		Prefix: aws.String(s3s.keyPrefix),
	})

	for paginator.HasMorePages() {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects in S3: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.Key != nil {
				// Convert S3 key back to schema ID
				key := *obj.Key
				if strings.HasPrefix(key, s3s.keyPrefix) && strings.HasSuffix(key, ".json") {
					schemaID := strings.TrimPrefix(key, s3s.keyPrefix)
					schemaID = strings.TrimSuffix(schemaID, ".json")

					// Convert path separators back to dots
					schemaID = strings.ReplaceAll(schemaID, "/", ".")

					schemaIDs = append(schemaIDs, schemaID)
				}
			}
		}
	}

	return schemaIDs, nil
}

// Exists checks if a schema exists in S3
func (s3s *S3Storage) Exists(ctx context.Context, id string) (bool, error) {
	if err := validateSchemaID(id); err != nil {
		return false, err
	}

	key := s3s.getKey(id)

	_, err := s3s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if schema %s exists in S3: %w", id, err)
	}

	return true, nil
}

// getKey converts schema ID to S3 key
// Schema IDs like "user.profile" become "schemas/user/profile.json"
func (s3s *S3Storage) getKey(id string) string {
	// Replace dots with path separators for nested structure
	pathParts := strings.Split(id, ".")
	path := strings.Join(pathParts, "/")
	return s3s.keyPrefix + path + ".json"
}

// GetBucket returns the S3 bucket name used by this storage
func (s3s *S3Storage) GetBucket() string {
	return s3s.bucket
}

// GetKeyPrefix returns the key prefix used by this storage
func (s3s *S3Storage) GetKeyPrefix() string {
	return s3s.keyPrefix
}

// Stats returns S3 storage statistics
func (s3s *S3Storage) Stats(ctx context.Context) (*S3Stats, error) {
	stats := &S3Stats{
		Type:   "s3",
		Bucket: s3s.bucket,
		Prefix: s3s.keyPrefix,
	}

	// Count objects and calculate total size
	paginator := s3.NewListObjectsV2Paginator(s3s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3s.bucket),
		Prefix: aws.String(s3s.keyPrefix),
	})

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate S3 stats: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.Key != nil && strings.HasSuffix(*obj.Key, ".json") {
				stats.SchemaCount++
				if obj.Size != nil {
					stats.TotalSize += *obj.Size
				}
			}
		}
	}

	return stats, nil
}

// S3Stats represents S3 storage statistics
type S3Stats struct {
	Type        string `json:"type"`
	Bucket      string `json:"bucket"`
	Prefix      string `json:"key_prefix"`
	SchemaCount int    `json:"schema_count"`
	TotalSize   int64  `json:"total_size_bytes"`
}

// BatchGet retrieves multiple schemas in a single operation using parallel requests
func (s3s *S3Storage) BatchGet(ctx context.Context, ids []string) (map[string][]byte, error) {
	if len(ids) == 0 {
		return make(map[string][]byte), nil
	}

	// Validate all IDs first
	for _, id := range ids {
		if err := validateSchemaID(id); err != nil {
			return nil, fmt.Errorf("invalid schema ID %s: %w", id, err)
		}
	}

	result := make(map[string][]byte)
	errChan := make(chan error, len(ids))
	dataChan := make(chan struct {
		id   string
		data []byte
	}, len(ids))

	// Launch parallel requests
	for _, id := range ids {
		go func(schemaID string) {
			data, err := s3s.Get(ctx, schemaID)
			if err != nil {
				errChan <- fmt.Errorf("failed to get schema %s: %w", schemaID, err)
				return
			}
			dataChan <- struct {
				id   string
				data []byte
			}{schemaID, data}
		}(id)
	}

	// Collect results
	for i := 0; i < len(ids); i++ {
		select {
		case err := <-errChan:
			return nil, err
		case data := <-dataChan:
			result[data.id] = data.data
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return result, nil
}

// BatchSet stores multiple schemas using parallel requests
func (s3s *S3Storage) BatchSet(ctx context.Context, schemas map[string][]byte) error {
	if len(schemas) == 0 {
		return nil
	}

	// Validate all IDs and data first
	for id, data := range schemas {
		if err := validateSchemaID(id); err != nil {
			return fmt.Errorf("invalid schema ID %s: %w", id, err)
		}
		if len(data) == 0 {
			return fmt.Errorf("cannot store empty schema data for %s", id)
		}
	}

	errChan := make(chan error, len(schemas))
	done := make(chan struct{}, len(schemas))

	// Launch parallel requests
	for id, data := range schemas {
		go func(schemaID string, schemaData []byte) {
			if err := s3s.Set(ctx, schemaID, schemaData); err != nil {
				errChan <- fmt.Errorf("failed to set schema %s: %w", schemaID, err)
				return
			}
			done <- struct{}{}
		}(id, data)
	}

	// Wait for all operations to complete
	for i := 0; i < len(schemas); i++ {
		select {
		case err := <-errChan:
			return err
		case <-done:
			// Success
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// BatchDelete removes multiple schemas using parallel requests
func (s3s *S3Storage) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// Validate all IDs first
	for _, id := range ids {
		if err := validateSchemaID(id); err != nil {
			return fmt.Errorf("invalid schema ID %s: %w", id, err)
		}
	}

	errChan := make(chan error, len(ids))
	done := make(chan struct{}, len(ids))

	// Launch parallel requests
	for _, id := range ids {
		go func(schemaID string) {
			if err := s3s.Delete(ctx, schemaID); err != nil {
				errChan <- fmt.Errorf("failed to delete schema %s: %w", schemaID, err)
				return
			}
			done <- struct{}{}
		}(id)
	}

	// Wait for all operations to complete
	for i := 0; i < len(ids); i++ {
		select {
		case err := <-errChan:
			return err
		case <-done:
			// Success
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// GetObjectInfo returns metadata about a schema object in S3
func (s3s *S3Storage) GetObjectInfo(ctx context.Context, id string) (*S3ObjectInfo, error) {
	if err := validateSchemaID(id); err != nil {
		return nil, err
	}

	key := s3s.getKey(id)

	result, err := s3s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("schema %s not found", id)
		}
		return nil, fmt.Errorf("failed to get object info for schema %s: %w", id, err)
	}

	info := &S3ObjectInfo{
		SchemaID:     id,
		Key:          key,
		ContentType:  aws.ToString(result.ContentType),
		Size:         aws.ToInt64(result.ContentLength),
		LastModified: result.LastModified,
		ETag:         aws.ToString(result.ETag),
		Metadata:     result.Metadata,
	}

	return info, nil
}

// S3ObjectInfo represents metadata about an S3 object
type S3ObjectInfo struct {
	SchemaID     string            `json:"schema_id"`
	Key          string            `json:"key"`
	ContentType  string            `json:"content_type"`
	Size         int64             `json:"size_bytes"`
	LastModified *time.Time        `json:"last_modified"`
	ETag         string            `json:"etag"`
	Metadata     map[string]string `json:"metadata"`
}

// SetWithMetadata stores a schema with custom metadata
func (s3s *S3Storage) SetWithMetadata(ctx context.Context, id string, data []byte, metadata map[string]string) error {
	if err := validateSchemaID(id); err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("cannot store empty schema data for %s", id)
	}

	key := s3s.getKey(id)

	// Merge with default metadata
	finalMetadata := map[string]string{
		"schema-id": id,
		"type":      "schema",
	}
	for k, v := range metadata {
		finalMetadata[k] = v
	}

	_, err := s3s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
		Metadata:    finalMetadata,
	})

	if err != nil {
		return fmt.Errorf("failed to store schema %s with metadata in S3: %w", id, err)
	}

	return nil
}
