package core

import (
	"sync"
	"time"
)

// HistoryManager manages undo/redo operations for the schema builder
type HistoryManager struct {
	entries    []*HistoryEntry
	currentPos int
	maxSize    int
	mu         sync.RWMutex
}

// HistoryEntry represents a single action in the history
type HistoryEntry struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Data        any       `json:"data"`
	Timestamp   time.Time `json:"timestamp"`
	UserID      string    `json:"userId,omitempty"`
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(maxSize int) *HistoryManager {
	return &HistoryManager{
		entries:    make([]*HistoryEntry, 0, maxSize),
		currentPos: -1,
		maxSize:    maxSize,
	}
}

// Push adds a new entry to the history
func (h *HistoryManager) Push(entry *HistoryEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove any entries after current position (if we're not at the end)
	if h.currentPos < len(h.entries)-1 {
		h.entries = h.entries[:h.currentPos+1]
	}

	// Add new entry
	h.entries = append(h.entries, entry)
	h.currentPos = len(h.entries) - 1

	// Trim if exceeding max size
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[1:]
		h.currentPos = len(h.entries) - 1
	}
}

// Undo moves back in history and returns the entry to undo
func (h *HistoryManager) Undo() *HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.currentPos < 0 {
		return nil
	}

	entry := h.entries[h.currentPos]
	h.currentPos--
	return entry
}

// Redo moves forward in history and returns the entry to redo
func (h *HistoryManager) Redo() *HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.currentPos >= len(h.entries)-1 {
		return nil
	}

	h.currentPos++
	return h.entries[h.currentPos]
}

// CanUndo returns true if there are entries to undo
func (h *HistoryManager) CanUndo() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.currentPos >= 0
}

// CanRedo returns true if there are entries to redo
func (h *HistoryManager) CanRedo() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.currentPos < len(h.entries)-1
}

// GetHistory returns the current history state
func (h *HistoryManager) GetHistory() ([]*HistoryEntry, int) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy of entries to prevent external modification
	entriesCopy := make([]*HistoryEntry, len(h.entries))
	copy(entriesCopy, h.entries)

	return entriesCopy, h.currentPos
}

// Clear removes all history entries
func (h *HistoryManager) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = h.entries[:0]
	h.currentPos = -1
}

// GetUndoDescription returns description of the next undo action
func (h *HistoryManager) GetUndoDescription() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.currentPos >= 0 && h.currentPos < len(h.entries) {
		return h.entries[h.currentPos].Description
	}
	return ""
}

// GetRedoDescription returns description of the next redo action
func (h *HistoryManager) GetRedoDescription() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.currentPos+1 < len(h.entries) {
		return h.entries[h.currentPos+1].Description
	}
	return ""
}
