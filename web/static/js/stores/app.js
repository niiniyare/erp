/**
 * Alpine.js App Store
 * Global application state management
 * Enhanced version with improved error handling and performance
 */
document.addEventListener('alpine:init', () => {
    Alpine.store('app', {
        // UI State
        sidebarOpen: false,
        mobileMenuOpen: false,
        theme: 'light',
        loading: false,
        
        // Application State
        currentTenant: null,
        currentPage: '',
        breadcrumbs: [],
        
        // Modals and Overlays
        activeModal: null,
        modalData: null,
        overlayVisible: false,
        
        // Notifications
        showNotifications: false,
        
        // Internal state (prefixed with _ to indicate private)
        _pendingRequests: 0,
        _previousOverflow: '',
        _resizeHandler: null,
        _keydownHandler: null,
        _htmxBeforeHandler: null,
        _htmxAfterHandler: null,
        
        // Computed Properties
        get isLoading() {
            // Backward compatible: returns true if loading OR if there are pending requests
            return this.loading || this._pendingRequests > 0;
        },
        
        // Storage Helper Methods (internal)
        _persistState(key, value) {
            try {
                localStorage.setItem(key, typeof value === 'string' ? value : JSON.stringify(value));
                return true;
            } catch (e) {
                console.error(`Failed to persist ${key}:`, e);
                return false;
            }
        },
        
        _loadState(key, defaultValue = null) {
            try {
                const saved = localStorage.getItem(key);
                if (!saved) return defaultValue;
                
                // Try to parse as JSON, fall back to raw string
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
        
        // Public Methods (all original methods preserved)
        toggleSidebar() {
            this.sidebarOpen = !this.sidebarOpen;
        },
        
        closeSidebar() {
            this.sidebarOpen = false;
        },
        
        toggleMobileMenu() {
            this.mobileMenuOpen = !this.mobileMenuOpen;
        },
        
        closeMobileMenu() {
            this.mobileMenuOpen = false;
        },
        
        toggleTheme() {
            this.theme = this.theme === 'light' ? 'dark' : 'light';
            this.persistTheme();
        },
        
        persistTheme() {
            this._persistState('theme', this.theme);
            // Keep both classList and attribute for maximum compatibility
            document.documentElement.classList.toggle('dark', this.theme === 'dark');
            document.documentElement.setAttribute('data-theme', this.theme);
        },
        
        loadTheme() {
            const saved = this._loadState('theme');
            if (saved && (saved === 'light' || saved === 'dark')) {
                this.theme = saved;
            } else {
                // Detect system preference
                this.theme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
            }
            this.persistTheme();
        },
        
        setLoading(state) {
            this.loading = Boolean(state);
        },
        
        setCurrentPage(page) {
            this.currentPage = page || '';
        },
        
        setBreadcrumbs(crumbs) {
            if (!Array.isArray(crumbs)) {
                this.breadcrumbs = [];
                return;
            }
            
            // Filter out invalid breadcrumb items
            this.breadcrumbs = crumbs.filter(crumb => {
                return crumb && (
                    typeof crumb === 'string' || 
                    (typeof crumb === 'object' && (crumb.label || crumb.title || crumb.name))
                );
            });
        },
        
        openModal(modalId, data = null) {
            if (!modalId) {
                console.warn('openModal called without modalId');
                return;
            }
            
            this.activeModal = modalId;
            this.modalData = data;
            this.overlayVisible = true;
            
            // Store previous overflow state
            this._previousOverflow = document.body.style.overflow || '';
            document.body.style.overflow = 'hidden';
        },
        
        closeModal() {
            this.activeModal = null;
            this.modalData = null;
            this.overlayVisible = false;
            
            // Restore previous overflow state
            document.body.style.overflow = this._previousOverflow;
            this._previousOverflow = '';
        },
        
        toggleNotifications() {
            this.showNotifications = !this.showNotifications;
        },
        
        closeNotifications() {
            this.showNotifications = false;
        },
        
        setTenant(tenant) {
            // Validate tenant data
            if (!tenant || typeof tenant !== 'object') {
                console.warn('Invalid tenant data provided to setTenant');
                this.currentTenant = null;
                localStorage.removeItem('currentTenant');
                return;
            }
            
            this.currentTenant = tenant;
            this._persistState('currentTenant', tenant);
        },
        
        loadTenant() {
            const saved = this._loadState('currentTenant');
            if (saved && typeof saved === 'object') {
                this.currentTenant = saved;
            } else {
                this.currentTenant = null;
            }
        },
        
        // Cleanup method for removing event listeners
        _cleanup() {
            if (this._resizeHandler) {
                window.removeEventListener('resize', this._resizeHandler);
            }
            if (this._keydownHandler) {
                document.removeEventListener('keydown', this._keydownHandler);
            }
            if (this._htmxBeforeHandler) {
                document.removeEventListener('htmx:beforeRequest', this._htmxBeforeHandler);
            }
            if (this._htmxAfterHandler) {
                document.removeEventListener('htmx:afterRequest', this._htmxAfterHandler);
            }
        },
        
        // Initialize app state
        init() {
            // Load persisted state
            this.loadTheme();
            this.loadTenant();
            
            // Setup resize handler with debouncing for performance
            let resizeTimeout;
            this._resizeHandler = () => {
                clearTimeout(resizeTimeout);
                resizeTimeout = setTimeout(() => {
                    if (window.innerWidth >= 768) {
                        this.closeMobileMenu();
                    }
                }, 150);
            };
            window.addEventListener('resize', this._resizeHandler);
            
            // Setup keyboard handler
            this._keydownHandler = (e) => {
                if (e.key === 'Escape') {
                    if (this.activeModal) {
                        this.closeModal();
                    } else if (this.showNotifications) {
                        this.closeNotifications();
                    } else if (this.mobileMenuOpen) {
                        this.closeMobileMenu();
                    }
                }
            };
            document.addEventListener('keydown', this._keydownHandler);
            
            // Handle HTMX loading states with request counting
            this._htmxBeforeHandler = () => {
                this._pendingRequests++;
                // Keep backward compatibility by also setting loading
                this.setLoading(true);
            };
            
            this._htmxAfterHandler = () => {
                this._pendingRequests = Math.max(0, this._pendingRequests - 1);
                // Only set loading to false if no pending requests
                if (this._pendingRequests === 0) {
                    this.setLoading(false);
                }
            };
            
            document.addEventListener('htmx:beforeRequest', this._htmxBeforeHandler);
            document.addEventListener('htmx:afterRequest', this._htmxAfterHandler);
            
            // Also handle HTMX errors to prevent stuck loading state
            document.addEventListener('htmx:responseError', this._htmxAfterHandler);
            document.addEventListener('htmx:sendError', this._htmxAfterHandler);
            
            // Optional: Cleanup on page unload (useful for SPA scenarios)
            window.addEventListener('beforeunload', () => {
                this._cleanup();
            });
        }
    });
});
