/**
 * @fileoverview Navbar Configuration Script for Flowbite
 * Manages navbar state, items, and interactions with Alpine.js and htmx
 * @version 1.0.0
 */

/**
 * @typedef {Object} NavbarBrand
 * @property {string} logo - Logo image URL
 * @property {string} logoAlt - Logo alt text
 * @property {string} text - Brand text
 * @property {string} url - Brand link URL
 */

/**
 * @typedef {Object} NavbarAction
 * @property {string} type - Action type (button, link, dropdown)
 * @property {string} label - Action label
 * @property {string} [icon] - Icon class
 * @property {string} [url] - Action URL
 * @property {Function} [onClick] - Click handler
 * @property {Object} [htmx] - htmx attributes
 * @property {string} [htmx.get] - htmx GET URL
 * @property {string} [htmx.post] - htmx POST URL
 * @property {string} [htmx.target] - htmx target selector
 * @property {string} [htmx.swap] - htmx swap strategy
 */

/**
 * @typedef {Object} NavbarItem
 * @property {string} id - Unique identifier
 * @property {string} label - Display label
 * @property {string} [url] - Navigation URL
 * @property {string} [icon] - Icon class
 * @property {boolean} [active] - Active state
 * @property {boolean} [disabled] - Disabled state
 * @property {string} [badge] - Badge text
 * @property {string} [badgeColor] - Badge color (info, success, warning, danger)
 * @property {NavbarItem[]} [children] - Dropdown items
 * @property {string} [permission] - Required permission
 * @property {Object} [htmx] - htmx configuration
 * @property {Object} [meta] - Additional metadata
 */

/**
 * @typedef {Object} NavbarSearchConfig
 * @property {boolean} enabled - Enable search
 * @property {string} placeholder - Search placeholder
 * @property {string} [endpoint] - Search API endpoint
 * @property {number} [minChars] - Minimum characters to trigger search
 * @property {number} [debounce] - Debounce delay in ms
 * @property {Object} [htmx] - htmx search configuration
 */

/**
 * @typedef {Object} NavbarConfig
 * @property {NavbarBrand} brand - Brand configuration
 * @property {NavbarItem[]} items - Navigation items
 * @property {NavbarAction[]} [actions] - Action buttons/dropdowns
 * @property {NavbarSearchConfig} [search] - Search configuration
 * @property {boolean} [sticky] - Sticky navbar
 * @property {string} [position] - Position (top, bottom)
 * @property {boolean} [transparent] - Transparent background
 * @property {string} [theme] - Theme variant
 * @property {Object} [responsive] - Responsive settings
 * @property {string} [responsive.breakpoint] - Mobile breakpoint
 */

/**
 * Navbar Manager Class
 * Manages navbar state and behavior
 */
class NavbarManager {
    /**
     * @param {NavbarConfig} config - Navbar configuration
     * @param {Object} [dependencies] - External dependencies
     * @param {Object} [dependencies.configManager] - Global config manager
     * @param {Object} [dependencies.apiManager] - API manager
     */
    constructor(config, dependencies = {}) {
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
     * @param {NavbarConfig} config - User configuration
     * @returns {NavbarConfig} Merged configuration
     */
    _mergeDefaults(config) {
        const defaults = {
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
    _deepMerge(target, source) {
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
     * @param {NavbarItem} item - Navbar item
     * @returns {boolean} Visibility state
     */
    isItemVisible(item) {
        if (!item.permission) return true;
        if (this.configManager) {
            return this.configManager.get(`features.${item.permission}`) === true;
        }
        return true;
    }

    /**
     * Filter items by visibility
     * @returns {NavbarItem[]} Visible items
     */
    getVisibleItems() {
        return this.config.items.filter(item => !item.disabled && this.isItemVisible(item));
    }

    /**
     * Set active item by URL or ID
     * @param {string} identifier - URL or ID
     */
    setActive(identifier) {
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
     * @returns {NavbarConfig} Configuration
     */
    getConfig() {
        return this.config;
    }

    /**
     * Update configuration
     * @param {Partial<NavbarConfig>} updates - Configuration updates
     */
    updateConfig(updates) {
        this.config = this._deepMerge(this.config, updates);
    }
}

/**
 * Create Alpine.js Navbar Store
 * @param {NavbarManager} navbarManager - Navbar manager instance
 * @returns {Function} Alpine store factory
 */
function createNavbarStore(navbarManager) {
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
        init() {
            // Set active item based on current URL
            navbarManager.setActive(window.location.pathname);
            
            // Close dropdowns on outside click
            document.addEventListener('click', (e) => {
                if (!e.target.closest('[data-dropdown-toggle]')) {
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
        toggleMobileMenu() {
            this.mobileMenuOpen = !this.mobileMenuOpen;
        },

        /**
         * Close mobile menu
         */
        closeMobileMenu() {
            this.mobileMenuOpen = false;
        },

        /**
         * Toggle dropdown
         * @param {string} dropdownId - Dropdown identifier
         */
        toggleDropdown(dropdownId) {
            this.activeDropdown = this.activeDropdown === dropdownId ? null : dropdownId;
        },

        /**
         * Check if dropdown is open
         * @param {string} dropdownId - Dropdown identifier
         * @returns {boolean} Open state
         */
        isDropdownOpen(dropdownId) {
            return this.activeDropdown === dropdownId;
        },

        /**
         * Get navbar configuration
         * @returns {NavbarConfig} Configuration
         */
        getConfig() {
            return navbarManager.getConfig();
        },

        /**
         * Get visible items
         * @returns {NavbarItem[]} Visible items
         */
        getItems() {
            return navbarManager.getVisibleItems();
        },

        /**
         * Navigate to URL
         * @param {string} url - Target URL
         * @param {Event} [event] - Click event
         */
        navigate(url, event) {
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
         * @param {string} query - Search query
         */
        handleSearch(query) {
            this.searchQuery = query;
            
            const config = navbarManager.getConfig().search;
            if (!config.enabled || query.length < config.minChars) {
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
            }, config.debounce);
        },

        /**
         * Perform search
         * @param {string} query - Search query
         */
        async performSearch(query) {
            const config = navbarManager.getConfig().search;
            
            if (!config.endpoint) return;

            this.searchLoading = true;

            try {
                if (navbarManager.apiManager) {
                    this.searchResults = await navbarManager.apiManager.get(
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
        clearSearch() {
            this.searchQuery = '';
            this.searchResults = [];
        },

        /**
         * Check if item is active
         * @param {NavbarItem} item - Navbar item
         * @returns {boolean} Active state
         */
        isActive(item) {
            return item.active === true;
        },

        /**
         * Check if item has children
         * @param {NavbarItem} item - Navbar item
         * @returns {boolean} Has children
         */
        hasChildren(item) {
            return Array.isArray(item.children) && item.children.length > 0;
        },

        /**
         * Get htmx attributes for item
         * @param {NavbarItem} item - Navbar item
         * @returns {Object} htmx attributes
         */
        getHtmxAttrs(item) {
            if (!item.htmx) return {};
            
            const attrs = {};
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
 * @param {NavbarConfig} config - Navbar configuration
 * @param {Object} [dependencies] - External dependencies
 * @returns {Object} Navbar instance
 */
function initializeNavbar(config, dependencies = {}) {
    const manager = new NavbarManager(config, dependencies);
    const store = createNavbarStore(manager);

    return {
        manager,
        store
    };
}

// Export to window
window.ERPAdmin = window.ERPAdmin || {};
window.ERPAdmin.Navbar = {
    NavbarManager,
    createNavbarStore,
    initializeNavbar
};
