package abac

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// EvaluationCache provides a high-performance, thread-safe cache for policy evaluation results.
type EvaluationCache struct {
	cache *ristretto.Cache[string, *PolicyEvaluationResult]
}

// NewEvaluationCache creates a new EvaluationCache with default settings.
func NewEvaluationCache() *EvaluationCache {
	// Initialize Ristretto cache.
	// These values can be tuned based on performance testing.
	// MaxCost: 100MB, NumCounters: 10M, BufferItems: 64
	cache, err := ristretto.NewCache[string, *PolicyEvaluationResult](&ristretto.Config[string, *PolicyEvaluationResult]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 27, // maximum cost of cache (128MB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		// If the cache fails to initialize, something is seriously wrong.
		// We panic to prevent the application from running in a broken state.
		panic(fmt.Sprintf("failed to create evaluation cache: %v", err))
	}
	return &EvaluationCache{cache: cache}
}

// Get retrieves a cached PolicyEvaluationResult.
func (ec *EvaluationCache) Get(req *PolicyEvaluationRequest) (*PolicyEvaluationResult, bool) {
	key, err := ec.generateCacheKey(req)
	if err != nil {
		// If we can't generate a key, we can't cache, but this shouldn't stop the evaluation.
		// We log this as a warning.
		logger.Warn("Failed to generate cache key for GET", logger.Fields{"error": err})
		return nil, false
	}

	value, found := ec.cache.Get(key)
	if !found {
		return nil, false
	}

	if value == nil {
		logger.Error("Nil value found in evaluation cache", logger.Fields{"key": key})
		return nil, false
	}

	return value, true
}

// Set stores a PolicyEvaluationResult in the cache.
func (ec *EvaluationCache) Set(req *PolicyEvaluationRequest, result *PolicyEvaluationResult) {
	if !req.Options.EnableCaching {
		return
	}

	key, err := ec.generateCacheKey(req)
	if err != nil {
		logger.Warn("Failed to generate cache key for SET", logger.Fields{"error": err})
		return
	}

	// Use the TTL from the request options, or a default if not provided.
	ttl := req.Options.CacheTTL
	if ttl <= 0 {
		ttl = 5 * time.Minute // Default TTL
	}

	// The "cost" of the item is set to 1 for simplicity.
	// For more advanced use cases, this could be the size of the result object in bytes.
	ec.cache.SetWithTTL(key, result, 1, ttl)
}

// generateCacheKey creates a stable, unique hash for a given policy evaluation request.
func (ec *EvaluationCache) generateCacheKey(req *PolicyEvaluationRequest) (string, error) {
	// We create a struct with only the fields that affect the decision,
	// ensuring the cache key is consistent.
	keyData := struct {
		PolicyID    string
		Subject     SubjectContext
		Resource    ResourceContext
		Action      ActionContext
		Environment EnvironmentContext
	}{
		PolicyID:    req.PolicyID.String(),
		Subject:     req.Subject,
		Resource:    req.Resource,
		Action:      req.Action,
		Environment: req.Environment,
	}

	// Marshal the struct to JSON. JSON provides a stable representation.
	jsonData, err := json.Marshal(keyData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cache key data: %w", err)
	}

	// Hash the JSON data to create a fixed-size, unique key.
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash), nil
}
