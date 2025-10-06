/**
 * Alpine.js User Store
 * User authentication and profile state management
 * Enhanced version with improved security, performance, and error handling
 */
document.addEventListener('alpine:init', () => {
    Alpine.store('user', {
        // User State
        isAuthenticated: false,
        profile: null,
        permissions: [],
        roles: [],
        preferences: {},
        
        // Authentication State
        loginAttempts: 0,
        maxLoginAttempts: 3,
        lockoutTime: null,
        
        // Session State
        sessionExpiry: null,
        lastActivity: Date.now(),
        
        // Internal state (private)
        _activityTimer: null,
        _sessionCheckInterval: null,
        _activityHandlers: [],
        _htmxErrorHandler: null,
        _lastActivityUpdate: 0,
        _activityThrottle: 5000, // Update activity max once per 5 seconds
        
        // Storage Helper Methods (internal)
        _persistState(key, value) {
            try {
                localStorage.setItem(key, typeof value === 'string' ? value : JSON.stringify(value));
                return true;
            } catch (e) {
                console.error(`Failed to persist ${key}:`, e);
                // Could be quota exceeded or private browsing mode
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
        
        // Validation helpers
        _isValidUserData(userData) {
            return userData && 
                   typeof userData === 'object' && 
                   userData.profile && 
                   typeof userData.profile === 'object';
        },
        
        _sanitizeArray(value) {
            return Array.isArray(value) ? value.filter(item => 
                item !== null && item !== undefined
            ) : [];
        },
        
        // Public Methods (all original methods preserved)
        setUser(userData) {
            if (!userData || typeof userData !== 'object') {
                console.warn('Invalid user data provided to setUser');
                return false;
            }
            
            this.profile = userData;
            this.isAuthenticated = true;
            this.loginAttempts = 0;
            this.lockoutTime = null;
            this.updateLastActivity();
            this.persistUserData();
            return true;
        },
        
        setPermissions(permissions) {
            this.permissions = this._sanitizeArray(permissions);
            this.persistUserData();
        },
        
        setRoles(roles) {
            this.roles = this._sanitizeArray(roles);
            this.persistUserData();
        },
        
        hasPermission(permission) {
            if (!permission) return false;
            return this.permissions.includes(permission);
        },
        
        hasRole(role) {
            if (!role) return false;
            return this.roles.includes(role);
        },
        
        hasAnyRole(roles) {
            if (!Array.isArray(roles) || roles.length === 0) return false;
            return roles.some(role => this.hasRole(role));
        },
        
        hasAllRoles(roles) {
            if (!Array.isArray(roles) || roles.length === 0) return false;
            return roles.every(role => this.hasRole(role));
        },
        
        updateProfile(profileData) {
            if (!this.profile) {
                console.warn('Cannot update profile: user not authenticated');
                return false;
            }
            
            if (!profileData || typeof profileData !== 'object') {
                console.warn('Invalid profile data provided');
                return false;
            }
            
            this.profile = { ...this.profile, ...profileData };
            this.persistUserData();
            return true;
        },
        
        updatePreferences(prefs) {
            if (!prefs || typeof prefs !== 'object') {
                console.warn('Invalid preferences provided');
                return false;
            }
            
            this.preferences = { ...this.preferences, ...prefs };
            this.persistUserData();
            return true;
        },
        
        getPreference(key, defaultValue = null) {
            if (!key) return defaultValue;
            return this.preferences.hasOwnProperty(key) ? this.preferences[key] : defaultValue;
        },
        
        incrementLoginAttempts() {
            this.loginAttempts++;
            
            if (this.loginAttempts >= this.maxLoginAttempts) {
                this.lockoutTime = Date.now() + (15 * 60 * 1000); // 15 minutes
                // Persist lockout state
                this._persistState('lockoutTime', this.lockoutTime);
                this._persistState('loginAttempts', this.loginAttempts);
            }
        },
        
        isLockedOut() {
            if (!this.lockoutTime) return false;
            
            const lockedOut = Date.now() < this.lockoutTime;
            
            // Clear lockout if expired
            if (!lockedOut && this.lockoutTime) {
                this.lockoutTime = null;
                this.loginAttempts = 0;
                this._removeState('lockoutTime');
                this._removeState('loginAttempts');
            }
            
            return lockedOut;
        },
        
        getRemainingLockoutTime() {
            if (!this.isLockedOut()) return 0;
            return Math.ceil((this.lockoutTime - Date.now()) / 1000);
        },
        
        updateLastActivity() {
            const now = Date.now();
            
            // Throttle updates to avoid excessive writes
            if (now - this._lastActivityUpdate < this._activityThrottle) {
                return;
            }
            
            this.lastActivity = now;
            this._lastActivityUpdate = now;
            this._persistState('userLastActivity', this.lastActivity.toString());
        },
        
        checkSessionExpiry() {
            if (!this.isAuthenticated) return false;
            
            if (this.sessionExpiry && Date.now() > this.sessionExpiry) {
                console.info('Session expired');
                this.logout();
                return true;
            }
            return false;
        },
        
        extendSession(expiryTime) {
            if (!expiryTime || typeof expiryTime !== 'number') {
                console.warn('Invalid expiry time provided');
                return false;
            }
            
            this.sessionExpiry = expiryTime;
            this._persistState('sessionExpiry', expiryTime.toString());
            return true;
        },
        
        logout() {
            this.isAuthenticated = false;
            this.profile = null;
            this.permissions = [];
            this.roles = [];
            this.preferences = {};
            this.sessionExpiry = null;
            this.lastActivity = Date.now();
            this.clearUserData();
            
            // Emit custom event for logout
            window.dispatchEvent(new CustomEvent('user:logout'));
        },
        
        persistUserData() {
            if (!this.isAuthenticated || !this.profile) {
                return false;
            }
            
            const userData = {
                profile: this.profile,
                permissions: this.permissions,
                roles: this.roles,
                preferences: this.preferences,
                lastActivity: this.lastActivity
            };
            
            return this._persistState('userData', userData);
        },
        
        loadUserData() {
            const saved = this._loadState('userData');
            
            if (saved && this._isValidUserData(saved)) {
                this.profile = saved.profile;
                this.permissions = this._sanitizeArray(saved.permissions);
                this.roles = this._sanitizeArray(saved.roles);
                this.preferences = saved.preferences || {};
                this.lastActivity = saved.lastActivity || Date.now();
                this.isAuthenticated = true;
            }
            
            // Load session expiry
            const sessionExpiry = this._loadState('sessionExpiry');
            if (sessionExpiry) {
                this.sessionExpiry = parseInt(sessionExpiry);
            }
            
            // Load lockout state
            const lockoutTime = this._loadState('lockoutTime');
            if (lockoutTime) {
                this.lockoutTime = parseInt(lockoutTime);
                this.loginAttempts = parseInt(this._loadState('loginAttempts', 0));
            }
        },
        
        clearUserData() {
            this._removeState('userData');
            this._removeState('sessionExpiry');
            this._removeState('userLastActivity');
            this._removeState('lockoutTime');
            this._removeState('loginAttempts');
        },
        
        // Activity tracking with throttling
        trackActivity() {
            if (this.isAuthenticated) {
                this.updateLastActivity();
            }
        },
        
        // Cleanup method
        _cleanup() {
            // Clear activity timer
            if (this._activityTimer) {
                clearTimeout(this._activityTimer);
                this._activityTimer = null;
            }
            
            // Clear session check interval
            if (this._sessionCheckInterval) {
                clearInterval(this._sessionCheckInterval);
                this._sessionCheckInterval = null;
            }
            
            // Remove activity event listeners
            this._activityHandlers.forEach(({ event, handler }) => {
                document.removeEventListener(event, handler);
            });
            this._activityHandlers = [];
            
            // Remove HTMX error handler
            if (this._htmxErrorHandler) {
                document.removeEventListener('htmx:responseError', this._htmxErrorHandler);
                this._htmxErrorHandler = null;
            }
        },
        
        // Initialize user store
        init() {
            // Load persisted data
            this.loadUserData();
            this.checkSessionExpiry();
            
            // Setup activity tracking with proper throttling
            const activityEvents = ['click', 'keypress', 'scroll', 'mousemove'];
            
            activityEvents.forEach(event => {
                const handler = () => {
                    // Clear existing timer
                    if (this._activityTimer) {
                        clearTimeout(this._activityTimer);
                    }
                    
                    // Set new timer with 1 second delay
                    this._activityTimer = setTimeout(() => {
                        this.trackActivity();
                    }, 1000);
                };
                
                document.addEventListener(event, handler, { passive: true });
                this._activityHandlers.push({ event, handler });
            });
            
            // Check session expiry periodically
            this._sessionCheckInterval = setInterval(() => {
                this.checkSessionExpiry();
            }, 60000); // Check every minute
            
            // Handle HTMX authentication errors
            this._htmxErrorHandler = (event) => {
                if (event.detail.xhr && event.detail.xhr.status === 401) {
                    this.logout();
                    
                    // Allow custom handling before redirect
                    const shouldRedirect = window.dispatchEvent(
                        new CustomEvent('user:unauthorized', { 
                            cancelable: true,
                            detail: { event }
                        })
                    );
                    
                    // Only redirect if not prevented
                    if (shouldRedirect) {
                        window.location.href = '/auth/login';
                    }
                }
            };
            document.addEventListener('htmx:responseError', this._htmxErrorHandler);
            
            // Cleanup on page unload
            window.addEventListener('beforeunload', () => {
                this._cleanup();
            });
            
            // Emit initialization event
            window.dispatchEvent(new CustomEvent('user:initialized', {
                detail: { isAuthenticated: this.isAuthenticated }
            }));
        }
    });
});
