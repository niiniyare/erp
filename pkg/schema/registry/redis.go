package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStorage implements Storage interface using Redis
type RedisStorage struct {
	client redis.Cmdable
	prefix string
	ttl    time.Duration
}

// RedisConfig configures Redis storage
type RedisConfig struct {
	Client redis.Cmdable // Redis client interface
	Prefix string        // Key prefix for schemas (default: "schema:")
	TTL    time.Duration // Optional TTL for cached schemas
}

// NewRedisStorage creates a new Redis storage backend
func NewRedisStorage(config RedisConfig) (*RedisStorage, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	if config.Prefix == "" {
		config.Prefix = "schema:"
	}

	return &RedisStorage{
		client: config.Client,
		prefix: config.Prefix,
		ttl:    config.TTL,
	}, nil
}

// Get retrieves a schema by ID from Redis
func (rs *RedisStorage) Get(ctx context.Context, id string) ([]byte, error) {
	if err := validateSchemaID(id); err != nil {
		return nil, err
	}

	key := rs.getKey(id)

	data, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("schema %s not found", id)
		}
		return nil, fmt.Errorf("failed to get schema %s from Redis: %w", id, err)
	}

	return data, nil
}

// Set stores a schema by ID to Redis
func (rs *RedisStorage) Set(ctx context.Context, id string, data []byte) error {
	if err := validateSchemaID(id); err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("cannot store empty schema data for %s", id)
	}

	key := rs.getKey(id)

	var err error
	if rs.ttl > 0 {
		err = rs.client.Set(ctx, key, data, rs.ttl).Err()
	} else {
		err = rs.client.Set(ctx, key, data, 0).Err()
	}

	if err != nil {
		return fmt.Errorf("failed to set schema %s in Redis: %w", id, err)
	}

	return nil
}

// Delete removes a schema by ID from Redis
func (rs *RedisStorage) Delete(ctx context.Context, id string) error {
	if err := validateSchemaID(id); err != nil {
		return err
	}

	key := rs.getKey(id)

	deleted, err := rs.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete schema %s from Redis: %w", id, err)
	}

	if deleted == 0 {
		return fmt.Errorf("schema %s not found", id)
	}

	return nil
}

// List returns all schema IDs from Redis
func (rs *RedisStorage) List(ctx context.Context) ([]string, error) {
	pattern := rs.prefix + "*"

	var allKeys []string
	iter := rs.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		allKeys = append(allKeys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan Redis keys: %w", err)
	}

	// Convert keys back to schema IDs
	var schemaIDs []string
	for _, key := range allKeys {
		if strings.HasPrefix(key, rs.prefix) {
			schemaID := strings.TrimPrefix(key, rs.prefix)
			schemaIDs = append(schemaIDs, schemaID)
		}
	}

	return schemaIDs, nil
}

// Exists checks if a schema exists in Redis
func (rs *RedisStorage) Exists(ctx context.Context, id string) (bool, error) {
	if err := validateSchemaID(id); err != nil {
		return false, err
	}

	key := rs.getKey(id)

	exists, err := rs.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if schema %s exists in Redis: %w", id, err)
	}

	return exists > 0, nil
}

// getKey converts schema ID to Redis key
func (rs *RedisStorage) getKey(id string) string {
	return rs.prefix + id
}

// GetPrefix returns the key prefix used by this storage
func (rs *RedisStorage) GetPrefix() string {
	return rs.prefix
}

// GetTTL returns the TTL used by this storage
func (rs *RedisStorage) GetTTL() time.Duration {
	return rs.ttl
}

// Stats returns Redis storage statistics
func (rs *RedisStorage) Stats(ctx context.Context) (*RedisStats, error) {
	stats := &RedisStats{
		Type: "redis",
	}

	// Get schema count
	pattern := rs.prefix + "*"
	iter := rs.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		stats.SchemaCount++
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to calculate schema count: %w", err)
	}

	// Get memory usage if available (Redis INFO command)
	if infoCmd := rs.client.Info(ctx, "memory"); infoCmd.Err() == nil {
		info := infoCmd.Val()
		if lines := strings.Split(info, "\r\n"); len(lines) > 0 {
			for _, line := range lines {
				if strings.HasPrefix(line, "used_memory:") {
					if memStr := strings.TrimPrefix(line, "used_memory:"); memStr != "" {
						var mem int64
						if _, err := fmt.Sscanf(memStr, "%d", &mem); err == nil {
							stats.UsedMemory = mem
						}
					}
					break
				}
			}
		}
	}

	// Get TTL info
	stats.TTL = rs.ttl

	return stats, nil
}

// RedisStats represents Redis storage statistics
type RedisStats struct {
	Type        string        `json:"type"`
	SchemaCount int           `json:"schema_count"`
	UsedMemory  int64         `json:"used_memory_bytes"`
	TTL         time.Duration `json:"ttl_seconds"`
}

// Ping checks Redis connectivity
func (rs *RedisStorage) Ping(ctx context.Context) error {
	return rs.client.Ping(ctx).Err()
}

// FlushAll removes all schemas from Redis (use with caution)
func (rs *RedisStorage) FlushAll(ctx context.Context) error {
	pattern := rs.prefix + "*"

	var keys []string
	iter := rs.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan Redis keys for flush: %w", err)
	}

	if len(keys) > 0 {
		if err := rs.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete Redis keys: %w", err)
		}
	}

	return nil
}

// GetWithTTL retrieves a schema and its remaining TTL
func (rs *RedisStorage) GetWithTTL(ctx context.Context, id string) ([]byte, time.Duration, error) {
	if err := validateSchemaID(id); err != nil {
		return nil, 0, err
	}

	key := rs.getKey(id)

	// Use pipeline for atomic operations
	pipe := rs.client.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, 0, fmt.Errorf("failed to get schema %s with TTL from Redis: %w", id, err)
	}

	data, err := getCmd.Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, 0, fmt.Errorf("schema %s not found", id)
		}
		return nil, 0, fmt.Errorf("failed to get schema %s from Redis: %w", id, err)
	}

	ttl, err := ttlCmd.Result()
	if err != nil {
		// If TTL command fails, return data with unknown TTL
		return data, -1, nil
	}

	return data, ttl, nil
}

// SetWithCustomTTL stores a schema with a custom TTL
func (rs *RedisStorage) SetWithCustomTTL(ctx context.Context, id string, data []byte, ttl time.Duration) error {
	if err := validateSchemaID(id); err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("cannot store empty schema data for %s", id)
	}

	key := rs.getKey(id)

	err := rs.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set schema %s in Redis with TTL %v: %w", id, ttl, err)
	}

	return nil
}

// BatchGet retrieves multiple schemas in a single operation
func (rs *RedisStorage) BatchGet(ctx context.Context, ids []string) (map[string][]byte, error) {
	if len(ids) == 0 {
		return make(map[string][]byte), nil
	}

	// Validate all IDs first
	for _, id := range ids {
		if err := validateSchemaID(id); err != nil {
			return nil, fmt.Errorf("invalid schema ID %s: %w", id, err)
		}
	}

	// Convert IDs to keys
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = rs.getKey(id)
	}

	// Use MGET for batch retrieval
	values, err := rs.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to batch get schemas from Redis: %w", err)
	}

	// Build result map
	result := make(map[string][]byte)
	for i, value := range values {
		if value != nil {
			if str, ok := value.(string); ok {
				result[ids[i]] = []byte(str)
			}
		}
	}

	return result, nil
}

// BatchSet stores multiple schemas in a single operation
func (rs *RedisStorage) BatchSet(ctx context.Context, schemas map[string][]byte) error {
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

	// Use pipeline for batch operations
	pipe := rs.client.Pipeline()

	for id, data := range schemas {
		key := rs.getKey(id)
		if rs.ttl > 0 {
			pipe.Set(ctx, key, data, rs.ttl)
		} else {
			pipe.Set(ctx, key, data, 0)
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to batch set schemas in Redis: %w", err)
	}

	return nil
}
