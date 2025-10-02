/**
 * @fileoverview Navbar Configuration Script for Flowbite (TypeScript)
 * Manages navbar state, items, and interactions with Alpine.js and htmx
 * @version 1.0.0
 */

// Types
interface NavbarBrand {
    logo?: string;
    logoAlt?: string;
    text?: string;
    url?: string;
}

interface HtmxConfig {
    get?: string;
    post?: string;
    target?: string;
    swap?: string;
    trigger?: string;
}

interface NavbarAction {
    type: 'button' | 'link' | 'dropdown';
    label: string;
    icon?: string;
    url?: string;
    onClick?: () => void;
    htmx?: HtmxConfig;
}

interface NavbarItem {
    id: string;
    label: string;
    url?: string;
    icon?: string;
    active?: boolean;
    disabled?: boolean;
    badge?: string;
    badgeColor?: 'info' | 'success' | 'warning' | 'danger';
    children?: NavbarItem[];
    permission?: string;
    htmx?: HtmxConfig;
    meta?: Record<string, any>;
}

interface NavbarSearchConfig {
    enabled: boolean;
    placeholder?: string;
    endpoint?: string;
    minChars?: number;
    debounce?: number;
    htmx?: HtmxConfig;
}

interface NavbarResponsiveConfig {
    breakpoint?: string;
}

interface NavbarConfig {
    brand: NavbarBrand;
    items: NavbarItem[];
    actions?: NavbarAction[];
    search?: NavbarSearchConfig;
    sticky?: boolean;
    position?: 'top' | 'bottom';
    transparent?: boolean;
    theme?: string;
    responsive?: NavbarResponsiveConfig;
}

interface NavbarDependencies {
    configManager?: any;
    apiManager?: any;
}

interface NavbarState {
    mobileMenuOpen: boolean;
    activeDropdown: string | null;
    searchQuery: string;
    searchResults: any[];
}

interface NavbarStore {
    mobileMenuOpen: boolean;
    activeDropdown: string | null;
    searchQuery: string;
    searchResults: any[];
    searchLoading: boolean;
    searchDebounceTimer: any;
    
    init(): void;
    toggleMobileMenu(): void;
    closeMobileMenu(): void;
    toggleDropdown(dropdownId: string): void;
    isDropdownOpen(dropdownId: string): boolean;
    getConfig(): NavbarConfig;
    getItems(): NavbarItem[];
    navigate(url: string, event?: Event): void;
    handleSearch(query: string): void;
    performSearch(query: string): Promise<void>;
    clearSearch(): void;
    isActive(item: NavbarItem): boolean;
    hasChildren(item: NavbarItem): boolean;
    getHtmxAttrs(item: NavbarItem): Record<string, string>;
}

interface NavbarInstance {
    manager: NavbarManager;
    store: () => NavbarStore;
}

// Export to make this a module
export {};

/**
 * Navbar Manager Class
 * Manages navbar state and behavior
 */
class NavbarManager {
    private config: NavbarConfig;
    private configManager: any;
    private apiManager: any;
    private state: NavbarState;

    /**
     * @param config - Navbar configuration
     * @param dependencies - External dependencies
     */
    constructor(config: NavbarConfig, dependencies: NavbarDependencies = {}) {
        this.config = this._mergeDefaults(config);
        this.configManager = dependencies.configManager;
        this.apiManager = dependencies.apiManager;
        this.state = {
            mobileMenuOpen: false,
            activeDropdown: null,
            searchQuery: '',
            searchResults: []
        };
    }

    /**
     * Merge with default configuration
     * @private
     * @param config - User configuration
     * @returns Merged configuration
     */
    private _mergeDefaults(config: NavbarConfig): NavbarConfig {
        const defaults: NavbarConfig = {
            brand: {
                logo: '',
                logoAlt: 'Logo',
                text: '',
                url: '/'
            },
            items: [],
            actions: [],
            search: {
                enabled: false,
                placeholder: 'Search...',
                minChars: 2,
                debounce: 300
            },
            sticky: true,
            position: 'top',
            transparent: false,
            theme: 'light',
            responsive: {
                breakpoint: 'md'
            }
        };

        return this._deepMerge(defaults, config);
    }

    /**
     * Deep merge objects
     * @private
     */
    private _deepMerge(target: any, source: any): any {
        const result = { ...target };
        for (const key in source) {
            if (source[key] && typeof source[key] === 'object' && !Array.isArray(source[key])) {
                result[key] = this._deepMerge(target[key] || {}, source[key]);
            } else {
                result[key] = source[key];
            }
        }
        return result;
    }

    /**
     * Check if item should be visible based on permissions
     * @param item - Navbar item
     * @returns Visibility state
     */
    isItemVisible(item: NavbarItem): boolean {
        if (!item.permission) return true;
        if (this.configManager) {
            return this.configManager.get(`features.${item.permission}`) === true;
        }
        return true;
    }

    /**
     * Filter items by visibility
     * @returns Visible items
     */
    getVisibleItems(): NavbarItem[] {
        return this.config.items.filter(item => !item.disabled && this.isItemVisible(item));
    }

    /**
     * Set active item by URL or ID
     * @param identifier - URL or ID
     */
    setActive(identifier: string): void {
        this.config.items.forEach(item => {
            item.active = item.url === identifier || item.id === identifier;
            if (item.children) {
                item.children.forEach(child => {
                    child.active = child.url === identifier || child.id === identifier;
                });
            }
        });
    }

    /**
     * Get configuration for rendering
     * @returns Configuration
     */
    getConfig(): NavbarConfig {
        return this.config;
    }

    /**
     * Update configuration
     * @param updates - Configuration updates
     */
    updateConfig(updates: Partial<NavbarConfig>): void {
        this.config = this._deepMerge(this.config, updates);
    }
}

/**
 * Create Alpine.js Navbar Store
 * @param navbarManager - Navbar manager instance
 * @returns Alpine store factory
 */
function createNavbarStore(navbarManager: NavbarManager): () => NavbarStore {
    return () => ({
        mobileMenuOpen: false,
        activeDropdown: null,
        searchQuery: '',
        searchResults: [],
        searchLoading: false,
        searchDebounceTimer: null,

        /**
         * Initialize navbar store
         */
        init(): void {
            // Set active item based on current URL
            navbarManager.setActive(window.location.pathname);
            
            // Close dropdowns on outside click
            document.addEventListener('click', (e: MouseEvent) => {
                if (!(e.target as Element).closest('[data-dropdown-toggle]')) {
                    this.activeDropdown = null;
                }
            });

            // Handle URL changes (for SPA navigation)
            window.addEventListener('popstate', () => {
                navbarManager.setActive(window.location.pathname);
            });
        },

        /**
         * Toggle mobile menu
         */
        toggleMobileMenu(): void {
            this.mobileMenuOpen = !this.mobileMenuOpen;
        },

        /**
         * Close mobile menu
         */
        closeMobileMenu(): void {
            this.mobileMenuOpen = false;
        },

        /**
         * Toggle dropdown
         * @param dropdownId - Dropdown identifier
         */
        toggleDropdown(dropdownId: string): void {
            this.activeDropdown = this.activeDropdown === dropdownId ? null : dropdownId;
        },

        /**
         * Check if dropdown is open
         * @param dropdownId - Dropdown identifier
         * @returns Open state
         */
        isDropdownOpen(dropdownId: string): boolean {
            return this.activeDropdown === dropdownId;
        },

        /**
         * Get navbar configuration
         * @returns Configuration
         */
        getConfig(): NavbarConfig {
            return navbarManager.getConfig();
        },

        /**
         * Get visible items
         * @returns Visible items
         */
        getItems(): NavbarItem[] {
            return navbarManager.getVisibleItems();
        },

        /**
         * Navigate to URL
         * @param url - Target URL
         * @param event - Click event
         */
        navigate(url: string, event?: Event): void {
            if (event) {
                event.preventDefault();
            }
            
            navbarManager.setActive(url);
            this.closeMobileMenu();
            
            // Trigger navigation (adjust based on your routing)
            window.history.pushState({}, '', url);
            
            // Dispatch custom event for SPA routers
            window.dispatchEvent(new CustomEvent('navigate', { detail: { url } }));
        },

        /**
         * Handle search input
         * @param query - Search query
         */
        handleSearch(query: string): void {
            this.searchQuery = query;
            
            const config = navbarManager.getConfig().search;
            if (!config || !config.enabled || query.length < (config.minChars || 2)) {
                this.searchResults = [];
                return;
            }

            // Clear existing timer
            if (this.searchDebounceTimer) {
                clearTimeout(this.searchDebounceTimer);
            }

            // Debounce search
            this.searchDebounceTimer = setTimeout(() => {
                this.performSearch(query);
            }, config.debounce || 300);
        },

        /**
         * Perform search
         * @param query - Search query
         */
        async performSearch(query: string): Promise<void> {
            const config = navbarManager.getConfig().search;
            
            if (!config || !config.endpoint) return;

            this.searchLoading = true;

            try {
                if ((navbarManager as any).apiManager) {
                    this.searchResults = await (navbarManager as any).apiManager.get(
                        config.endpoint,
                        { q: query }
                    );
                }
            } catch (error) {
                console.error('Search failed:', error);
                this.searchResults = [];
            } finally {
                this.searchLoading = false;
            }
        },

        /**
         * Clear search
         */
        clearSearch(): void {
            this.searchQuery = '';
            this.searchResults = [];
        },

        /**
         * Check if item is active
         * @param item - Navbar item
         * @returns Active state
         */
        isActive(item: NavbarItem): boolean {
            return item.active === true;
        },

        /**
         * Check if item has children
         * @param item - Navbar item
         * @returns Has children
         */
        hasChildren(item: NavbarItem): boolean {
            return Array.isArray(item.children) && item.children.length > 0;
        },

        /**
         * Get htmx attributes for item
         * @param item - Navbar item
         * @returns htmx attributes
         */
        getHtmxAttrs(item: NavbarItem): Record<string, string> {
            if (!item.htmx) return {};
            
            const attrs: Record<string, string> = {};
            if (item.htmx.get) attrs['hx-get'] = item.htmx.get;
            if (item.htmx.post) attrs['hx-post'] = item.htmx.post;
            if (item.htmx.target) attrs['hx-target'] = item.htmx.target;
            if (item.htmx.swap) attrs['hx-swap'] = item.htmx.swap;
            if (item.htmx.trigger) attrs['hx-trigger'] = item.htmx.trigger;
            
            return attrs;
        }
    });
}

/**
 * Initialize Navbar
 * @param config - Navbar configuration
 * @param dependencies - External dependencies
 * @returns Navbar instance
 */
function initializeNavbar(config: NavbarConfig, dependencies: NavbarDependencies = {}): NavbarInstance {
    const manager = new NavbarManager(config, dependencies);
    const store = createNavbarStore(manager);

    return {
        manager,
        store
    };
}

// Export to window
(window as any).ERPAdmin = (window as any).ERPAdmin || {};
(window as any).ERPAdmin.Navbar = {
    NavbarManager,
    createNavbarStore,
    initializeNavbar
};