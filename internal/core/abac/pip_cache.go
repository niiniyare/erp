package abac

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PIPCache provides a thread-safe, expiring cache for PIP attributes.
type PIPCache struct {
	cache map[string]*PIPCacheEntry
	mutex sync.RWMutex
}

// NewPIPCache creates a new PIPCache.
func NewPIPCache() *PIPCache {
	return &PIPCache{
		cache: make(map[string]*PIPCacheEntry),
	}
}

// Get retrieves an item from the cache. It returns the value and true if the item exists and has not expired.
func (pc *PIPCache) Get(category, attributeID string, subjectID uuid.UUID) (any, bool) {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	key := fmt.Sprintf("%s:%s:%s", category, attributeID, subjectID.String())
	entry, exists := pc.cache[key]
	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		// Although we can't delete from the map in a read-locked section,
		// we treat it as a cache miss. A separate cleanup goroutine could handle expired items.
		return nil, false
	}

	return entry.Value, true
}

// Set adds an item to the cache with a default TTL.
func (pc *PIPCache) Set(category, attributeID string, subjectID uuid.UUID, value any) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	key := fmt.Sprintf("%s:%s:%s", category, attributeID, subjectID.String())
	// Using a default TTL of 5 minutes as specified in the original plan.
	// This could be made configurable later.
	pc.cache[key] = &PIPCacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
}

// PIPCacheEntry represents a single entry in the PIP cache.
type PIPCacheEntry struct {
	Value     any
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired.
func (pce *PIPCacheEntry) IsExpired() bool {
	return time.Now().After(pce.ExpiresAt)
}
