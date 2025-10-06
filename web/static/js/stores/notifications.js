/**
 * Alpine.js Notifications Store
 * User notification and messaging state management
 */
document.addEventListener('alpine:init', () => {
    Alpine.store('notifications', {
        // Notification State
        items: [],
        unreadCount: 0,
        maxNotifications: 50,
        
        // UI State
        isOpen: false,
        filter: 'all', // 'all', 'unread', 'read'
        sortBy: 'newest', // 'newest', 'oldest', 'priority'
        
        // Auto-dismiss settings
        autoDismissTimeout: 5000,
        pauseOnHover: true,
        
        // Notification types and their configs
        types: {
            success: { icon: 'check-circle', color: 'green', autoDismiss: true },
            error: { icon: 'x-circle', color: 'red', autoDismiss: false },
            warning: { icon: 'exclamation-triangle', color: 'yellow', autoDismiss: false },
            info: { icon: 'information-circle', color: 'blue', autoDismiss: true }
        },
        
        // Internal state (private)
        _timers: new Map(), // Track auto-dismiss timers
        _cleanupInterval: null,
        _htmxAfterHandler: null,
        _htmxErrorHandler: null,
        _persistenceTimer: null,
        _persistenceDebounce: 300, // Debounce persistence writes
        _cachedFiltered: null,
        _cacheInvalidated: true,
        
        // Storage Helper Methods (internal)
        _persistState(key, value) {
            try {
                const serialized = typeof value === 'string' ? value : JSON.stringify(value);
                localStorage.setItem(key, serialized);
                return true;
            } catch (e) {
                console.error(`Failed to persist ${key}:`, e);
                // Handle quota exceeded gracefully
                if (e.name === 'QuotaExceededError') {
                    this._handleStorageQuotaExceeded();
                }
                return false;
            }
        },
        
        _loadState(key, defaultValue = null) {
            try {
                const saved = localStorage.getItem(key);
                if (!saved) return defaultValue;
                
                try {
                    return JSON.parse(saved);
                } catch {
                    return saved;
                }
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
        
        _handleStorageQuotaExceeded() {
            // Remove oldest read notifications to free up space
            const readNotifications = this.items.filter(n => n.read);
            if (readNotifications.length > 0) {
                const toRemove = Math.ceil(readNotifications.length / 2);
                for (let i = 0; i < toRemove; i++) {
                    const oldest = readNotifications[readNotifications.length - 1 - i];
                    this.dismiss(oldest.id, { skipPersist: true });
                }
                // Try persisting again
                this._persistNotifications({ immediate: true });
            }
        },
        
        // Validation helpers
        _isValidNotification(notification) {
            return notification && 
                   typeof notification === 'object' && 
                   notification.message && 
                   typeof notification.message === 'string' &&
                   notification.message.trim().length > 0;
        },
        
        _sanitizeNotification(notification) {
            // Remove functions before storage (they can't be serialized)
            const sanitized = { ...notification };
            
            if (Array.isArray(sanitized.actions)) {
                sanitized.actions = sanitized.actions.map(action => {
                    const { callback, ...rest } = action;
                    return rest; // Remove callback function
                });
            }
            
            return sanitized;
        },
        
        _invalidateCache() {
            this._cacheInvalidated = true;
            this._cachedFiltered = null;
        },
        
        // Public Methods (all original methods preserved)
        add(notification) {
            if (!this._isValidNotification(notification)) {
                console.warn('Invalid notification data');
                return null;
            }
            
            const id = this._generateId();
            const now = Date.now();
            const type = notification.type || 'info';
            const config = this.types[type] || this.types.info;
            
            const newNotification = {
                id,
                message: notification.message.trim(),
                title: notification.title || '',
                type,
                icon: notification.icon || config.icon,
                color: notification.color || config.color,
                read: false,
                persistent: notification.persistent || false,
                priority: notification.priority || 'normal',
                timestamp: now,
                expiresAt: notification.expiresAt || null,
                actions: notification.actions || [],
                data: notification.data || {},
                autoDismiss: notification.hasOwnProperty('autoDismiss') 
                    ? notification.autoDismiss 
                    : config.autoDismiss
            };
            
            // Add to beginning of array (newest first)
            this.items.unshift(newNotification);
            this.unreadCount++;
            this._invalidateCache();
            
            // Trim to max notifications
            if (this.items.length > this.maxNotifications) {
                const removed = this.items.splice(this.maxNotifications);
                // Adjust unread count and cancel timers for removed items
                removed.forEach(item => {
                    if (!item.read) {
                        this.unreadCount = Math.max(0, this.unreadCount - 1);
                    }
                    this._cancelTimer(item.id);
                });
            }
            
            this._persistNotifications();
            
            // Auto-dismiss if configured
            if (newNotification.autoDismiss && !newNotification.persistent) {
                const timer = setTimeout(() => {
                    this.dismiss(id);
                }, this.autoDismissTimeout);
                
                this._timers.set(id, timer);
            }
            
            // Emit custom event
            window.dispatchEvent(new CustomEvent('notification:added', {
                detail: { id, notification: newNotification }
            }));
            
            return id;
        },
        
        success(message, options = {}) {
            return this.add({ ...options, message, type: 'success' });
        },
        
        error(message, options = {}) {
            return this.add({ ...options, message, type: 'error' });
        },
        
        warning(message, options = {}) {
            return this.add({ ...options, message, type: 'warning' });
        },
        
        info(message, options = {}) {
            return this.add({ ...options, message, type: 'info' });
        },
        
        dismiss(id, options = {}) {
            const index = this.items.findIndex(item => item.id === id);
            if (index === -1) return false;
            
            const notification = this.items[index];
            
            // Cancel auto-dismiss timer
            this._cancelTimer(id);
            
            // Update unread count
            if (!notification.read) {
                this.unreadCount = Math.max(0, this.unreadCount - 1);
            }
            
            // Remove from array
            this.items.splice(index, 1);
            this._invalidateCache();
            
            // Persist unless explicitly skipped
            if (!options.skipPersist) {
                this._persistNotifications();
            }
            
            // Emit custom event
            window.dispatchEvent(new CustomEvent('notification:dismissed', {
                detail: { id, notification }
            }));
            
            return true;
        },
        
        markAsRead(id) {
            const notification = this.items.find(item => item.id === id);
            if (!notification || notification.read) return false;
            
            notification.read = true;
            this.unreadCount = Math.max(0, this.unreadCount - 1);
            this._invalidateCache();
            this._persistNotifications();
            
            // Emit custom event
            window.dispatchEvent(new CustomEvent('notification:read', {
                detail: { id, notification }
            }));
            
            return true;
        },
        
        markAllAsRead() {
            let marked = 0;
            this.items.forEach(item => {
                if (!item.read) {
                    item.read = true;
                    marked++;
                }
            });
            
            if (marked > 0) {
                this.unreadCount = 0;
                this._invalidateCache();
                this._persistNotifications();
                
                window.dispatchEvent(new CustomEvent('notification:allRead', {
                    detail: { count: marked }
                }));
            }
            
            return marked;
        },
        
        clear() {
            // Cancel all timers
            this._timers.forEach((timer, id) => {
                clearTimeout(timer);
            });
            this._timers.clear();
            
            const count = this.items.length;
            this.items = [];
            this.unreadCount = 0;
            this._invalidateCache();
            this._persistNotifications();
            
            window.dispatchEvent(new CustomEvent('notification:cleared', {
                detail: { count }
            }));
        },
        
        clearRead() {
            const beforeCount = this.items.length;
            const readIds = this.items.filter(item => item.read).map(item => item.id);
            
            // Cancel timers for read notifications
            readIds.forEach(id => this._cancelTimer(id));
            
            this.items = this.items.filter(item => !item.read);
            this._invalidateCache();
            
            if (beforeCount !== this.items.length) {
                this._persistNotifications();
                
                window.dispatchEvent(new CustomEvent('notification:readCleared', {
                    detail: { count: readIds.length }
                }));
            }
        },
        
        // Computed getters with caching
        get filteredItems() {
            // Return cached result if still valid
            if (!this._cacheInvalidated && this._cachedFiltered) {
                return this._cachedFiltered;
            }
            
            let filtered = [...this.items];
            
            // Apply filter
            switch (this.filter) {
                case 'unread':
                    filtered = filtered.filter(item => !item.read);
                    break;
                case 'read':
                    filtered = filtered.filter(item => item.read);
                    break;
            }
            
            // Apply sorting
            switch (this.sortBy) {
                case 'oldest':
                    filtered.sort((a, b) => a.timestamp - b.timestamp);
                    break;
                case 'priority':
                    const priorityOrder = { high: 3, normal: 2, low: 1 };
                    filtered.sort((a, b) => {
                        const aPriority = priorityOrder[a.priority] || 2;
                        const bPriority = priorityOrder[b.priority] || 2;
                        if (aPriority === bPriority) {
                            return b.timestamp - a.timestamp;
                        }
                        return bPriority - aPriority;
                    });
                    break;
                default: // 'newest'
                    filtered.sort((a, b) => b.timestamp - a.timestamp);
            }
            
            // Cache the result
            this._cachedFiltered = filtered;
            this._cacheInvalidated = false;
            
            return filtered;
        },
        
        get hasUnread() {
            return this.unreadCount > 0;
        },
        
        get isEmpty() {
            return this.items.length === 0;
        },
        
        // UI Methods
        toggle() {
            this.isOpen = !this.isOpen;
        },
        
        open() {
            this.isOpen = true;
        },
        
        close() {
            this.isOpen = false;
        },
        
        setFilter(filter) {
            if (['all', 'unread', 'read'].includes(filter) && this.filter !== filter) {
                this.filter = filter;
                this._invalidateCache();
            }
        },
        
        setSortBy(sortBy) {
            if (['newest', 'oldest', 'priority'].includes(sortBy) && this.sortBy !== sortBy) {
                this.sortBy = sortBy;
                this._invalidateCache();
            }
        },
        
        // Action handling
        executeAction(notificationId, actionId) {
            const notification = this.items.find(item => item.id === notificationId);
            if (!notification) return false;
            
            const action = notification.actions.find(a => a.id === actionId);
            if (!action) return false;
            
            // Mark as read when action is executed
            this.markAsRead(notificationId);
            
            // Execute action callback if provided
            if (typeof action.callback === 'function') {
                try {
                    action.callback(notification, action);
                } catch (e) {
                    console.error('Error executing notification action:', e);
                    return false;
                }
            }
            
            // Dismiss notification if action specifies
            if (action.dismiss) {
                this.dismiss(notificationId);
            }
            
            window.dispatchEvent(new CustomEvent('notification:actionExecuted', {
                detail: { notificationId, actionId, action }
            }));
            
            return true;
        },
        
        // Timer management
        _cancelTimer(id) {
            const timer = this._timers.get(id);
            if (timer) {
                clearTimeout(timer);
                this._timers.delete(id);
            }
        },
        
        pauseAutoDismiss(id) {
            this._cancelTimer(id);
        },
        
        resumeAutoDismiss(id) {
            const notification = this.items.find(item => item.id === id);
            if (notification && notification.autoDismiss && !notification.persistent) {
                const timer = setTimeout(() => {
                    this.dismiss(id);
                }, this.autoDismissTimeout);
                
                this._timers.set(id, timer);
            }
        },
        
        // Utility methods
        _generateId() {
            return `notif_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
        },
        
        _persistNotifications(options = {}) {
            // Clear existing debounce timer
            if (this._persistenceTimer) {
                clearTimeout(this._persistenceTimer);
            }
            
            const doPersist = () => {
                const data = {
                    items: this.items.map(item => this._sanitizeNotification(item)),
                    unreadCount: this.unreadCount
                };
                this._persistState('notifications', data);
            };
            
            // Persist immediately if requested, otherwise debounce
            if (options.immediate) {
                doPersist();
            } else {
                this._persistenceTimer = setTimeout(doPersist, this._persistenceDebounce);
            }
        },
        
        _loadNotifications() {
            const data = this._loadState('notifications');
            
            if (data && typeof data === 'object') {
                this.items = Array.isArray(data.items) ? data.items : [];
                this.unreadCount = typeof data.unreadCount === 'number' ? data.unreadCount : 0;
                
                // Validate and clean loaded data
                this.items = this.items.filter(item => {
                    return item && 
                           item.id && 
                           item.message && 
                           typeof item.timestamp === 'number';
                });
                
                // Recalculate unread count to ensure accuracy
                const actualUnread = this.items.filter(item => !item.read).length;
                if (actualUnread !== this.unreadCount) {
                    this.unreadCount = actualUnread;
                }
                
                this._cleanupExpired();
                this._invalidateCache();
            }
        },
        
        _cleanupExpired() {
            const now = Date.now();
            const beforeCount = this.items.length;
            
            this.items = this.items.filter(item => {
                const expired = item.expiresAt && now > item.expiresAt;
                
                if (expired) {
                    if (!item.read) {
                        this.unreadCount = Math.max(0, this.unreadCount - 1);
                    }
                    this._cancelTimer(item.id);
                    return false;
                }
                
                return true;
            });
            
            if (this.items.length !== beforeCount) {
                this._invalidateCache();
                this._persistNotifications({ immediate: true });
            }
        },
        
        // Cleanup method
        _cleanup() {
            // Clear all timers
            this._timers.forEach(timer => clearTimeout(timer));
            this._timers.clear();
            
            // Clear intervals
            if (this._cleanupInterval) {
                clearInterval(this._cleanupInterval);
                this._cleanupInterval = null;
            }
            
            // Clear persistence timer
            if (this._persistenceTimer) {
                clearTimeout(this._persistenceTimer);
                this._persistenceTimer = null;
            }
            
            // Remove event listeners
            if (this._htmxAfterHandler) {
                document.removeEventListener('htmx:afterRequest', this._htmxAfterHandler);
                this._htmxAfterHandler = null;
            }
            
            if (this._htmxErrorHandler) {
                document.removeEventListener('htmx:responseError', this._htmxErrorHandler);
                this._htmxErrorHandler = null;
            }
        },
        
        // Server sync methods (placeholder for future implementation)
        syncWithServer() {
            // TODO: Implement server synchronization
            console.info('Server sync not yet implemented');
        },
        
        markAsReadOnServer(id) {
            // TODO: Implement server-side read status update
            console.info('Server-side read status not yet implemented');
        },
        
        // Initialize notifications store
        init() {
            this._loadNotifications();
            
            // Set up periodic cleanup of expired notifications
            this._cleanupInterval = setInterval(() => {
                this._cleanupExpired();
            }, 60000); // Check every minute
            
            // Listen for HTMX events to show server notifications
            this._htmxAfterHandler = (event) => {
                const response = event.detail.xhr;
                if (!response) return;
                
                // Check for notification headers from server
                const notificationHeader = response.getResponseHeader('X-Notification');
                if (notificationHeader) {
                    try {
                        const notification = JSON.parse(notificationHeader);
                        this.add(notification);
                    } catch (e) {
                        console.error('Failed to parse server notification:', e);
                    }
                }
            };
            document.addEventListener('htmx:afterRequest', this._htmxAfterHandler);
            
            // Listen for errors to show error notifications
            this._htmxErrorHandler = (event) => {
                const xhr = event.detail.xhr;
                if (!xhr) return;
                
                const status = xhr.status;
                let message = 'An error occurred';
                
                switch (status) {
                    case 400:
                        message = 'Bad request. Please check your input.';
                        break;
                    case 401:
                        message = 'Authentication required. Please log in.';
                        break;
                    case 403:
                        message = 'You do not have permission to perform this action.';
                        break;
                    case 404:
                        message = 'The requested resource was not found.';
                        break;
                    case 500:
                        message = 'Server error. Please try again later.';
                        break;
                    default:
                        if (status >= 500) {
                            message = `Server error ${status}. Please try again later.`;
                        } else if (status >= 400) {
                            message = `Request failed with error ${status}.`;
                        }
                }
                
                this.error(message, {
                    title: 'Request Failed',
                    persistent: status >= 500,
                    data: { 
                        status, 
                        url: event.detail.requestConfig?.path || event.detail.pathInfo?.requestPath
                    }
                });
            };
            document.addEventListener('htmx:responseError', this._htmxErrorHandler);
            
            // Cleanup on page unload
            window.addEventListener('beforeunload', () => {
                this._cleanup();
            });
            
            // Emit initialization event
            window.dispatchEvent(new CustomEvent('notifications:initialized', {
                detail: { 
                    count: this.items.length, 
                    unread: this.unreadCount 
                }
            }));
        }
    });
});
