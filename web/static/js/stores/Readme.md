# Alpine.js Stores - Design Patterns & Usage Guide

## Table of Contents
1. [Overview](#overview)
2. [Core Design Patterns](#core-design-patterns)
3. [Architecture Principles](#architecture-principles)
4. [Store Implementation Guide](#store-implementation-guide)
5. [Usage Examples](#usage-examples)
6. [Best Practices](#best-practices)
7. [Common Pitfalls](#common-pitfalls)
8. [Testing Guidelines](#testing-guidelines)

---

## Overview

This document outlines the design patterns, architectural decisions, and best practices used in our Alpine.js store implementations. All future stores should follow these patterns to maintain consistency, reliability, and performance.

### Store Purpose
- **App Store**: Global UI state and application-level concerns
- **User Store**: Authentication, authorization, and user profile management
- **Notifications Store**: User notifications and messaging system

---

## Core Design Patterns

### 1. **State Segregation Pattern**

Separate state into clear categories with descriptive comments:

```javascript
Alpine.store('example', {
    // Public State (documented, stable API)
    publicProperty: 'value',
    isActive: false,
    
    // Internal State (private, prefixed with _)
    _timers: new Map(),
    _cacheInvalidated: true,
    _eventHandlers: [],
    
    // Configuration (constants, settings)
    maxItems: 50,
    debounceDelay: 300
});
```

**Why?**
- Clear API boundaries
- Prevents accidental usage of internal state
- Makes refactoring safer
- Self-documenting code

### 2. **Private Method Convention**

Prefix all internal methods with underscore:

```javascript
// Public API - stable, documented
methodName() {
    return this._internalHelper();
},

// Private - internal use only, can change
_internalHelper() {
    // Implementation details
}
```

**When to make a method private:**
- Used only internally
- Implementation may change
- Not part of the public contract
- Helper or utility function

### 3. **Centralized Storage Pattern**

Always use helper methods for localStorage operations:

```javascript
// ✅ CORRECT - Centralized with error handling
_persistState(key, value) {
    try {
        localStorage.setItem(key, JSON.stringify(value));
        return true;
    } catch (e) {
        console.error(`Failed to persist ${key}:`, e);
        if (e.name === 'QuotaExceededError') {
            this._handleStorageQuotaExceeded();
        }
        return false;
    }
},

_loadState(key, defaultValue = null) {
    try {
        const saved = localStorage.getItem(key);
        return saved ? JSON.parse(saved) : defaultValue;
    } catch (e) {
        console.error(`Failed to load ${key}:`, e);
        return defaultValue;
    }
},

// ❌ WRONG - Direct localStorage access
someMethod() {
    localStorage.setItem('key', JSON.stringify(value)); // No error handling!
}
```

**Benefits:**
- Consistent error handling
- Single point of maintenance
- Quota exceeded handling
- Easy to add encryption/compression later

### 4. **Memory Management Pattern**

Track and cleanup all side effects:

```javascript
Alpine.store('example', {
    // Track all resources
    _timers: new Map(),
    _intervals: [],
    _eventHandlers: [],
    
    init() {
        // Setup resources
        this._setupEventListeners();
        
        // Cleanup on unload
        window.addEventListener('beforeunload', () => {
            this._cleanup();
        });
    },
    
    _cleanup() {
        // Clear timers
        this._timers.forEach(timer => clearTimeout(timer));
        this._timers.clear();
        
        // Clear intervals
        this._intervals.forEach(id => clearInterval(id));
        this._intervals = [];
        
        // Remove event listeners
        this._eventHandlers.forEach(({ event, handler, target }) => {
            target.removeEventListener(event, handler);
        });
        this._eventHandlers = [];
    }
});
```

**Critical Resources to Track:**
- setTimeout/setInterval
- Event listeners
- HTMX event handlers
- WebSocket connections
- Fetch abort controllers
- Animation frames

### 5. **Debouncing Pattern**

Debounce expensive operations:

```javascript
_persistenceTimer: null,
_persistenceDebounce: 300,

_persistData(options = {}) {
    // Clear existing timer
    if (this._persistenceTimer) {
        clearTimeout(this._persistenceTimer);
    }
    
    const doPersist = () => {
        // Actual persistence logic
        this._persistState('key', this.data);
    };
    
    // Immediate or debounced
    if (options.immediate) {
        doPersist();
    } else {
        this._persistenceTimer = setTimeout(doPersist, this._persistenceDebounce);
    }
}
```

**When to debounce:**
- localStorage writes
- API calls
- DOM updates
- Search queries
- Scroll/resize handlers

### 6. **Validation Pattern**

Validate all inputs:

```javascript
// Validation helpers
_isValidData(data) {
    return data && 
           typeof data === 'object' && 
           data.requiredField &&
           typeof data.requiredField === 'string';
},

_sanitizeArray(value) {
    return Array.isArray(value) 
        ? value.filter(item => item !== null && item !== undefined)
        : [];
},

// Public method with validation
setData(data) {
    if (!this._isValidData(data)) {
        console.warn('Invalid data provided to setData');
        return false;
    }
    
    this.data = data;
    this._persist();
    return true;
}
```

**Validation Checklist:**
- Type checking
- Null/undefined handling
- Range validation
- Required fields
- Data sanitization
- Return success/failure

### 7. **Event-Driven Pattern**

Emit custom events for major state changes:

```javascript
add(item) {
    this.items.push(item);
    
    // Emit custom event
    window.dispatchEvent(new CustomEvent('store:itemAdded', {
        detail: { item, timestamp: Date.now() }
    }));
    
    return item.id;
}
```

**Event Naming Convention:**
- Format: `storeName:action`
- Examples: `user:login`, `notification:dismissed`, `app:themeChanged`

**When to emit events:**
- State changes affecting other components
- User actions completed
- Async operations finished
- Errors occurred
- Lifecycle events (init, cleanup)

### 8. **Caching Pattern**

Cache computed values with smart invalidation:

```javascript
_cachedValue: null,
_cacheInvalidated: true,

_invalidateCache() {
    this._cacheInvalidated = true;
    this._cachedValue = null;
},

get computedValue() {
    if (!this._cacheInvalidated && this._cachedValue !== null) {
        return this._cachedValue;
    }
    
    // Expensive computation
    const result = this._performExpensiveCalculation();
    
    this._cachedValue = result;
    this._cacheInvalidated = false;
    
    return result;
},

// Invalidate when state changes
updateState(newState) {
    this.state = newState;
    this._invalidateCache(); // ← Important!
}
```

---

## Architecture Principles

### 1. **Backward Compatibility First**

All improvements must maintain backward compatibility:

```javascript
// ✅ CORRECT - Add computed property, keep original
loading: false,
_pendingRequests: 0,

get isLoading() {
    return this.loading || this._pendingRequests > 0;
},

setLoading(state) {
    this.loading = Boolean(state); // Original API preserved
}

// ❌ WRONG - Breaking change
// Removed 'loading' property and replaced with computed only
```

### 2. **Progressive Enhancement**

Add features without breaking existing code:

```javascript
// Old usage still works
store.add(item);

// New usage adds options
store.add(item, { 
    skipPersist: true,
    silent: false 
});
```

### 3. **Fail Gracefully**

Never throw errors from store methods:

```javascript
// ✅ CORRECT
someMethod(data) {
    try {
        // Operation
        return true;
    } catch (e) {
        console.error('Operation failed:', e);
        return false;
    }
}

// ❌ WRONG
someMethod(data) {
    if (!data) throw new Error('Data required');
}
```

### 4. **Single Responsibility**

Each store manages one domain:

- **App Store**: UI state, navigation, modals, theme
- **User Store**: Auth, permissions, profile, session
- **Notifications Store**: Messages, alerts, toasts

Don't mix concerns across stores.

### 5. **Explicit Over Implicit**

Be explicit about side effects:

```javascript
// ✅ CORRECT - Clear what happens
updateProfile(data) {
    this.profile = { ...this.profile, ...data };
    this._persistData();
    this._invalidateCache();
    window.dispatchEvent(new CustomEvent('user:profileUpdated'));
    return true;
}

// ❌ WRONG - Hidden side effects
updateProfile(data) {
    this.profile = data; // What else happens?
}
```

---

## Store Implementation Guide

### Template Structure

```javascript
document.addEventListener('alpine:init', () => {
    Alpine.store('storeName', {
        // ==============================================
        // PUBLIC STATE
        // ==============================================
        publicProperty: 'initial value',
        isActive: false,
        items: [],
        
        // ==============================================
        // CONFIGURATION
        // ==============================================
        maxItems: 100,
        debounceDelay: 300,
        
        // ==============================================
        // INTERNAL STATE (Private - prefix with _)
        // ==============================================
        _timers: new Map(),
        _intervals: [],
        _eventHandlers: [],
        _cachedValue: null,
        _cacheInvalidated: true,
        
        // ==============================================
        // COMPUTED PROPERTIES
        // ==============================================
        get computedValue() {
            return this._calculateValue();
        },
        
        // ==============================================
        // STORAGE HELPERS (Private)
        // ==============================================
        _persistState(key, value) {
            try {
                localStorage.setItem(key, JSON.stringify(value));
                return true;
            } catch (e) {
                console.error(`Failed to persist ${key}:`, e);
                return false;
            }
        },
        
        _loadState(key, defaultValue = null) {
            try {
                const saved = localStorage.getItem(key);
                return saved ? JSON.parse(saved) : defaultValue;
            } catch (e) {
                console.error(`Failed to load ${key}:`, e);
                return defaultValue;
            }
        },
        
        _removeState(key) {
            try {
                localStorage.removeItem(key);
            } catch (e) {
                console.error(`Failed to remove ${key}:`, e);
            }
        },
        
        // ==============================================
        // VALIDATION HELPERS (Private)
        // ==============================================
        _isValidData(data) {
            return data && typeof data === 'object';
        },
        
        _sanitizeArray(value) {
            return Array.isArray(value) ? value : [];
        },
        
        // ==============================================
        // PUBLIC METHODS
        // ==============================================
        publicMethod(param) {
            if (!this._isValidData(param)) {
                console.warn('Invalid parameter');
                return false;
            }
            
            // Implementation
            this._invalidateCache();
            this._persist();
            
            window.dispatchEvent(new CustomEvent('store:action'));
            return true;
        },
        
        // ==============================================
        // PRIVATE METHODS
        // ==============================================
        _internalHelper() {
            // Internal logic
        },
        
        _invalidateCache() {
            this._cacheInvalidated = true;
            this._cachedValue = null;
        },
        
        _persist() {
            // Debounced persistence
        },
        
        // ==============================================
        // CLEANUP
        // ==============================================
        _cleanup() {
            // Clear timers
            this._timers.forEach(timer => clearTimeout(timer));
            this._timers.clear();
            
            // Clear intervals
            this._intervals.forEach(id => clearInterval(id));
            this._intervals = [];
            
            // Remove event listeners
            this._eventHandlers.forEach(({ event, handler, target }) => {
                target.removeEventListener(event, handler);
            });
            this._eventHandlers = [];
        },
        
        // ==============================================
        // INITIALIZATION
        // ==============================================
        init() {
            // Load persisted state
            this._loadData();
            
            // Setup event listeners
            this._setupListeners();
            
            // Cleanup on unload
            window.addEventListener('beforeunload', () => {
                this._cleanup();
            });
            
            // Emit initialization event
            window.dispatchEvent(new CustomEvent('store:initialized'));
        }
    });
});
```

---

## Usage Examples

### App Store Usage

```javascript
// Toggle sidebar
Alpine.store('app').toggleSidebar();

// Open modal with data
Alpine.store('app').openModal('userEdit', { userId: 123 });

// Close modal
Alpine.store('app').closeModal();

// Set loading state
Alpine.store('app').setLoading(true);

// Change theme
Alpine.store('app').toggleTheme();

// Listen for events
window.addEventListener('app:modalClosed', (e) => {
    console.log('Modal closed:', e.detail);
});
```

### User Store Usage

```javascript
// Set user on login
Alpine.store('user').setUser({
    id: 1,
    name: 'John Doe',
    email: 'john@example.com'
});

// Set permissions
Alpine.store('user').setPermissions(['read', 'write', 'delete']);

// Check permission
if (Alpine.store('user').hasPermission('write')) {
    // Allow action
}

// Check role
if (Alpine.store('user').hasRole('admin')) {
    // Show admin UI
}

// Check multiple roles
if (Alpine.store('user').hasAnyRole(['admin', 'moderator'])) {
    // Show moderation tools
}

// Update profile
Alpine.store('user').updateProfile({
    avatar: 'new-avatar.jpg',
    theme: 'dark'
});

// Logout
Alpine.store('user').logout();

// Listen for events
window.addEventListener('user:logout', () => {
    // Redirect to login
    window.location.href = '/login';
});
```

### Notifications Store Usage

```javascript
// Add notifications
const id = Alpine.store('notifications').success('Changes saved!');

Alpine.store('notifications').error('Failed to save', {
    title: 'Error',
    persistent: true
});

Alpine.store('notifications').warning('Low disk space', {
    priority: 'high',
    actions: [
        { id: 'cleanup', label: 'Clean Up', dismiss: true }
    ]
});

// With expiration
Alpine.store('notifications').info('Session expires in 5 minutes', {
    expiresAt: Date.now() + (5 * 60 * 1000)
});

// Dismiss notification
Alpine.store('notifications').dismiss(id);

// Mark as read
Alpine.store('notifications').markAsRead(id);

// Mark all as read
Alpine.store('notifications').markAllAsRead();

// Filter notifications
Alpine.store('notifications').setFilter('unread');

// Get filtered items
const unread = Alpine.store('notifications').filteredItems;

// Execute action
Alpine.store('notifications').executeAction(notificationId, actionId);

// Pause/resume auto-dismiss
Alpine.store('notifications').pauseAutoDismiss(id);
Alpine.store('notifications').resumeAutoDismiss(id);

// Listen for events
window.addEventListener('notification:added', (e) => {
    console.log('New notification:', e.detail);
});
```

### Alpine Template Usage

```html
<!-- App Store -->
<div x-data>
    <!-- Sidebar toggle -->
    <button @click="$store.app.toggleSidebar()">
        Menu
    </button>
    
    <!-- Conditional rendering -->
    <div x-show="$store.app.sidebarOpen">
        Sidebar content
    </div>
    
    <!-- Loading state -->
    <div x-show="$store.app.isLoading">
        <div class="spinner"></div>
    </div>
    
    <!-- Theme-aware class -->
    <div :class="$store.app.theme === 'dark' ? 'bg-gray-900' : 'bg-white'">
        Content
    </div>
</div>

<!-- User Store -->
<div x-data>
    <!-- Show if authenticated -->
    <template x-if="$store.user.isAuthenticated">
        <div>
            Welcome, <span x-text="$store.user.profile.name"></span>
        </div>
    </template>
    
    <!-- Permission-based rendering -->
    <template x-if="$store.user.hasPermission('write')">
        <button>Edit</button>
    </template>
    
    <!-- Role-based rendering -->
    <template x-if="$store.user.hasRole('admin')">
        <a href="/admin">Admin Panel</a>
    </template>
</div>

<!-- Notifications Store -->
<div x-data>
    <!-- Badge with count -->
    <button @click="$store.notifications.toggle()">
        Notifications
        <span x-show="$store.notifications.hasUnread"
              x-text="$store.notifications.unreadCount"
              class="badge">
        </span>
    </button>
    
    <!-- Notification list -->
    <div x-show="$store.notifications.isOpen">
        <template x-for="notif in $store.notifications.filteredItems" :key="notif.id">
            <div :class="notif.read ? 'opacity-50' : ''"
                 @click="$store.notifications.markAsRead(notif.id)">
                <p x-text="notif.message"></p>
                <button @click.stop="$store.notifications.dismiss(notif.id)">
                    Dismiss
                </button>
            </div>
        </template>
        
        <!-- Empty state -->
        <div x-show="$store.notifications.isEmpty">
            No notifications
        </div>
    </div>
</div>
```

---

## Best Practices

### 1. **Return Values**

Always return meaningful values:

```javascript
// ✅ CORRECT - Returns boolean success
updateData(data) {
    if (!this._isValidData(data)) {
        return false;
    }
    this.data = data;
    return true;
}

// Usage
if (store.updateData(newData)) {
    // Success handling
} else {
    // Error handling
}
```

### 2. **Error Handling**

Handle errors gracefully, never throw:

```javascript
// ✅ CORRECT
_persistData() {
    try {
        localStorage.setItem('key', JSON.stringify(this.data));
        return true;
    } catch (e) {
        console.error('Persistence failed:', e);
        
        if (e.name === 'QuotaExceededError') {
            this._handleQuotaExceeded();
        }
        
        return false;
    }
}

// ❌ WRONG
_persistData() {
    localStorage.setItem('key', JSON.stringify(this.data));
    // No error handling - will break app if storage fails
}
```

### 3. **Immutability**

Use immutable updates for objects and arrays:

```javascript
// ✅ CORRECT - Immutable
updateProfile(updates) {
    this.profile = { ...this.profile, ...updates };
}

updateItems(newItem) {
    this.items = [...this.items, newItem];
}

// ❌ WRONG - Mutating
updateProfile(updates) {
    Object.assign(this.profile, updates); // May not trigger reactivity
}

updateItems(newItem) {
    this.items.push(newItem); // Direct mutation
}
```

### 4. **Event Listeners**

Always track and cleanup event listeners:

```javascript
// ✅ CORRECT - Tracked and cleaned
_setupListeners() {
    const handler = (e) => this._handleEvent(e);
    document.addEventListener('click', handler);
    
    this._eventHandlers.push({
        event: 'click',
        handler,
        target: document
    });
}

_cleanup() {
    this._eventHandlers.forEach(({ event, handler, target }) => {
        target.removeEventListener(event, handler);
    });
    this._eventHandlers = [];
}

// ❌ WRONG - Memory leak
_setupListeners() {
    document.addEventListener('click', (e) => this._handleEvent(e));
    // No way to remove this listener!
}
```

### 5. **Throttling/Debouncing**

Throttle high-frequency events:

```javascript
// For scroll, mousemove, resize
_lastUpdate: 0,
_throttleDelay: 100,

handleHighFrequency() {
    const now = Date.now();
    if (now - this._lastUpdate < this._throttleDelay) {
        return; // Skip this update
    }
    
    this._lastUpdate = now;
    this._doExpensiveOperation();
}

// For search, API calls, persistence
_debounceTimer: null,
_debounceDelay: 300,

handleDebounced() {
    clearTimeout(this._debounceTimer);
    
    this._debounceTimer = setTimeout(() => {
        this._doExpensiveOperation();
    }, this._debounceDelay);
}
```

### 6. **Documentation**

Document public APIs with JSDoc:

```javascript
/**
 * Updates user profile with new data
 * @param {Object} profileData - Profile fields to update
 * @param {string} [profileData.name] - User's display name
 * @param {string} [profileData.email] - User's email address
 * @param {string} [profileData.avatar] - Avatar URL
 * @returns {boolean} True if update successful, false otherwise
 * @emits user:profileUpdated
 */
updateProfile(profileData) {
    if (!this.profile || !profileData) {
        return false;
    }
    
    this.profile = { ...this.profile, ...profileData };
    this._persist();
    
    window.dispatchEvent(new CustomEvent('user:profileUpdated', {
        detail: { profile: this.profile }
    }));
    
    return true;
}
```

### 7. **Type Safety**

Validate types at runtime:

```javascript
_isValidConfig(config) {
    return config &&
           typeof config === 'object' &&
           typeof config.timeout === 'number' &&
           config.timeout > 0 &&
           Array.isArray(config.items);
}

setConfig(config) {
    if (!this._isValidConfig(config)) {
        console.warn('Invalid config:', config);
        return false;
    }
    
    this.config = config;
    return true;
}
```

---

## Common Pitfalls

### 1. **Direct localStorage Usage**

```javascript
// ❌ WRONG
someMethod() {
    localStorage.setItem('key', JSON.stringify(value));
}

// ✅ CORRECT
someMethod() {
    this._persistState('key', value);
}
```

### 2. **Not Cleaning Up Resources**

```javascript
// ❌ WRONG - Memory leak
init() {
    setInterval(() => {
        this.checkStatus();
    }, 1000);
}

// ✅ CORRECT
init() {
    const intervalId = setInterval(() => {
        this.checkStatus();
    }, 1000);
    
    this._intervals.push(intervalId);
}
```

### 3. **Mutating State Directly**

```javascript
// ❌ WRONG
addItem(item) {
    this.items.push(item); // Direct mutation
}

// ✅ CORRECT
addItem(item) {
    this.items = [...this.items, item]; // Immutable
}
```

### 4. **No Input Validation**

```javascript
// ❌ WRONG
setUser(userData) {
    this.user = userData; // What if userData is null?
}

// ✅ CORRECT
setUser(userData) {
    if (!userData || typeof userData !== 'object') {
        console.warn('Invalid user data');
        return false;
    }
    
    this.user = userData;
    return true;
}
```

### 5. **Forgetting Cache Invalidation**

```javascript
// ❌ WRONG
updateData(newData) {
    this.data = newData;
    // Forgot to invalidate cache!
}

// ✅ CORRECT
updateData(newData) {
    this.data = newData;
    this._invalidateCache();
}
```

### 6. **Storing Functions in localStorage**

```javascript
// ❌ WRONG
_persist() {
    const data = {
        items: this.items,
        callback: this.onComplete // Function won't serialize!
    };
    localStorage.setItem('data', JSON.stringify(data));
}

// ✅ CORRECT
_persist() {
    const data = {
        items: this.items.map(item => {
            const { callback, ...rest } = item;
            return rest; // Remove functions
        })
    };
    localStorage.setItem('data', JSON.stringify(data));
}
```

### 7. **Not Handling Storage Quota**

```javascript
// ❌ WRONG
_persist() {
    localStorage.setItem('key', JSON.stringify(this.data));
}

// ✅ CORRECT
_persist() {
    try {
        localStorage.setItem('key', JSON.stringify(this.data));
    } catch (e) {
        if (e.name === 'QuotaExceededError') {
            this._handleQuotaExceeded();
        }
    }
}
```

---

## Testing Guidelines

### Unit Testing Stores

```javascript
describe('User Store', () => {
    let store;
    
    beforeEach(() => {
        // Clear localStorage
        localStorage.clear();
        
        // Get fresh store instance
        store = Alpine.store('user');
    });
    
    afterEach(() => {
        // Cleanup
        store._cleanup();
    });
    
    it('should set user data', () => {
        const userData = { id: 1, name: 'John' };
        const result = store.setUser(userData);
        
        expect(result).toBe(true);
        expect(store.profile).toEqual(userData);
        expect(store.isAuthenticated).toBe(true);
    });
    
    it('should validate user data', () => {
        const result = store.setUser(null);
        
        expect(result).toBe(false);
        expect(store.profile).toBeNull();
    });
    
    it('should persist to localStorage', () => {
        const userData = { id: 1, name: 'John' };
        store.setUser(userData);
        
        const saved = JSON.parse(localStorage.getItem('userData'));
        expect(saved.profile).toEqual(userData);
    });
    
    it('should emit events', (done) => {
        window.addEventListener('user:login', (e) => {
            expect(e.detail.user).toBeDefined();
            done();
        });
        
        store.setUser({ id: 1, name: 'John' });
    });
});
```

### Integration Testing

```javascript
describe('Store Integration', () => {
    it('should coordinate between stores', () => {
        // Login user
        Alpine.store('user').setUser({ id: 1, name: 'John' });
        
        // Add notification
        const notifId = Alpine.store('notifications').success('Logged in');
        
        // Logout should clear both
        Alpine.store('user').logout();
        
        expect(Alpine.store('user').isAuthenticated).toBe(false);
        expect(Alpine.store('notifications').items).toEqual([]);
    });
});
```

---

## Checklist for New Stores

When creating a new store, ensure:

- [ ] State segregation (public/private/config)
- [ ] Private methods prefixed with `_`
- [ ] Centralized storage helpers
- [ ] Input validation on all setters
- [ ] Immutable state updates
- [ ] Memory management (timers/listeners tracked)
- [ ] Cleanup method implemented
- [ ] Debouncing for expensive operations
- [ ] Cache invalidation where needed
- [ ] Custom events for major actions
- [ ] Error handling (no throws)
- [ ] Return values from methods
- [ ] JSDoc comments on public methods
- [ ] Backward compatibility maintained
- [ ] Event listeners properly cleaned up
- [ ] localStorage quota handling
- [ ] Functions not serialized to storage
- [ ] Unit tests written
- [ ] Integration tests for interactions

---

## Summary

Following these patterns ensures:

✅ **Performance** - Efficient through caching and debouncing  
✅ **Reliability** - Graceful error handling  
✅ **Maintainability** - Clear structure and documentation  
✅ **Scalability** - Memory-efficient resource management  
✅ **Compatibility** - Backward compatible changes  
✅ **Developer Experience** - Predictable, well-documented APIs

All future store implementations should follow these guidelines to maintain code quality and consistency across the application.
