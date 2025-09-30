/**
 * @fileoverview Breadcrumb Configuration Script for Flowbite
 * Manages breadcrumb navigation with Alpine.js and htmx integration
 * @version 1.0.0
 */

/**
 * @typedef {Object} BreadcrumbItem
 * @property {string} id - Unique identifier
 * @property {string} label - Display label
 * @property {string} [url] - Navigation URL
 * @property {string} [icon] - Icon class
 * @property {boolean} [active] - Active/current page indicator
 * @property {boolean} [disabled] - Disabled state
 * @property {Object} [htmx] - htmx configuration
 * @property {string} [htmx.get] - htmx GET URL
 * @property {string} [htmx.target] - htmx target
 * @property {string} [htmx.swap] - htmx swap strategy
 * @property {Object} [meta] - Additional metadata
 */

/**
 * @typedef {Object} BreadcrumbConfig
 * @property {BreadcrumbItem[]} items - Breadcrumb items
 * @property {string} [separator] - Separator type (slash, chevron, arrow, dot)
 * @property {string} [homeIcon] - Home icon class
 * @property {string} [homeUrl] - Home URL
 * @property {string} [homeLabel] - Home label
 * @property {boolean} [showHome] - Show home item
 * @property {number} [maxItems] - Maximum visible items before collapse
 * @property {boolean} [collapsible] - Enable collapsing for long breadcrumbs
 * @property {string} [theme] - Theme variant (light, dark)
 * @property {Object} [aria] - ARIA labels for accessibility
 */

/**
 * Breadcrumb Manager Class
 * Manages breadcrumb state and navigation
 */
class BreadcrumbManager {
    /**
     * @param {BreadcrumbConfig} config - Breadcrumb configuration
     * @param {Object} [dependencies] - External dependencies
     * @param {Object} [dependencies.configManager] - Global config manager
     */
    constructor(config, dependencies = {}) {
        this.config = this._mergeDefaults(config);
        this.configManager = dependencies.configManager;
        this.history = [];
    }

    /**
     * Merge with default configuration
     * @private
     * @param {BreadcrumbConfig} config - User configuration
     * @returns {BreadcrumbConfig} Merged configuration
     */
    _mergeDefaults(config) {
        const defaults = {
            items: [],
            separator: 'chevron',
            homeIcon: 'fa fa-home',
            homeUrl: '/',
            homeLabel: 'Home',
            showHome: true,
            maxItems: 5,
            collapsible: true,
            theme: 'light',
            aria: {
                label: 'Breadcrumb',
                current: 'Current page'
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
     * Get separator icon/symbol
     * @returns {string} Separator HTML or icon class
     */
    getSeparator() {
        const separators = {
            slash: '/',
            chevron: 'fa fa-chevron-right',
            arrow: 'fa fa-arrow-right',
            dot: '•'
        };
        return separators[this.config.separator] || separators.chevron;
    }

    /**
     * Get all items including home
     * @returns {BreadcrumbItem[]} All breadcrumb items
     */
    getAllItems() {
        const items = [];
        
        if (this.config.showHome) {
            items.push({
                id: 'home',
                label: this.config.homeLabel,
                url: this.config.homeUrl,
                icon: this.config.homeIcon,
                active: false
            });
        }

        return [...items, ...this.config.items];
    }

    /**
     * Get visible items (with collapse logic)
     * @returns {BreadcrumbItem[]} Visible items
     */
    getVisibleItems() {
        const allItems = this.getAllItems();
        
        if (!this.config.collapsible || allItems.length <= this.config.maxItems) {
            return allItems;
        }

        // Keep first, last, and some middle items
        const first = allItems[0];
        const last = allItems[allItems.length - 1];
        const remaining = this.config.maxItems - 2;
        
        const middle = allItems.slice(1, -1).slice(-remaining);

        return [
            first,
            { id: 'ellipsis', label: '...', disabled: true },
            ...middle,
            last
        ];
    }

    /**
     * Get collapsed items
     * @returns {BreadcrumbItem[]} Collapsed items
     */
    getCollapsedItems() {
        const allItems = this.getAllItems();
        const visibleItems = this.getVisibleItems();
        
        if (allItems.length <= this.config.maxItems) {
            return [];
        }

        return allItems.filter(item => 
            !visibleItems.includes(item) && item.id !== 'ellipsis'
        );
    }

    /**
     * Set items from URL path
     * @param {string} path - URL path
     * @param {Object} [options] - Options
     * @param {Function} [options.labelFormatter] - Function to format labels from path segments
     */
    setFromPath(path, options = {}) {
        const segments = path.split('/').filter(Boolean);
        const items = [];
        let currentPath = '';

        segments.forEach((segment, index) => {
            currentPath += `/${segment}`;
            const isLast = index === segments.length - 1;
            
            items.push({
                id: `path-${index}`,
                label: options.labelFormatter 
                    ? options.labelFormatter(segment, index) 
                    : this._formatLabel(segment),
                url: isLast ? null : currentPath,
                active: isLast
            });
        });

        this.config.items = items;
    }

    /**
     * Format label from path segment
     * @private
     * @param {string} segment - Path segment
     * @returns {string} Formatted label
     */
    _formatLabel(segment) {
        return segment
            .split('-')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1))
            .join(' ');
    }

    /**
     * Add breadcrumb item
     * @param {BreadcrumbItem} item - Item to add
     */
    addItem(item) {
        // Set previous items as inactive
        this.config.items.forEach(i => i.active = false);
        
        // Add new item as active
        this.config.items.push({ ...item, active: true });
    }

    /**
     * Remove item by ID
     * @param {string} id - Item ID
     */
    removeItem(id) {
        this.config.items = this.config.items.filter(item => item.id !== id);
    }

    /**
     * Set items
     * @param {BreadcrumbItem[]} items - New items
     */
    setItems(items) {
        this.config.items = items;
    }

    /**
     * Clear all items
     */
    clear() {
        this.config.items = [];
    }

    /**
     * Navigate to item
     * @param {string} id - Item ID
     * @returns {BreadcrumbItem|null} Item navigated to
     */
    navigateToItem(id) {
        const item = this.getAllItems().find(i => i.id === id);
        if (item && !item.active && !item.disabled) {
            // Remove items after this one
            const index = this.config.items.findIndex(i => i.id === id);
            if (index >= 0) {
                this.config.items = this.config.items.slice(0, index + 1);
                this.config.items.forEach((i, idx) => {
                    i.active = idx === index;
                });
            }
        }
        return item;
    }

    /**
     * Get configuration
     * @returns {BreadcrumbConfig} Configuration
     */
    getConfig() {
        return this.config;
    }

    /**
     * Update configuration
     * @param {Partial<BreadcrumbConfig>} updates - Updates
     */
    updateConfig(updates) {
        this.config = this._deepMerge(this.config, updates);
    }
}

/**
 * Create Alpine.js Breadcrumb Store
 * @param {BreadcrumbManager} breadcrumbManager - Breadcrumb manager instance
 * @returns {Function} Alpine store factory
 */
function createBreadcrumbStore(breadcrumbManager) {
    return () => ({
        showCollapsed: false,

        /**
         * Initialize breadcrumb store
         */
        init() {
            // Auto-generate from current path if no items provided
            const config = breadcrumbManager.getConfig();
            if (config.items.length === 0) {
                breadcrumbManager.setFromPath(window.location.pathname);
            }

            // Listen for navigation events
            window.addEventListener('navigate', (e) => {
                if (e.detail?.url) {
                    breadcrumbManager.setFromPath(e.detail.url);
                }
            });
        },

        /**
         * Get visible items
         * @returns {BreadcrumbItem[]} Visible items
         */
        getItems() {
            return breadcrumbManager.getVisibleItems();
        },

        /**
         * Get all items
         * @returns {BreadcrumbItem[]} All items
         */
        getAllItems() {
            return breadcrumbManager.getAllItems();
        },

        /**
         * Get collapsed items
         * @returns {BreadcrumbItem[]} Collapsed items
         */
        getCollapsedItems() {
            return breadcrumbManager.getCollapsedItems();
        },

        /**
         * Get separator
         * @returns {string} Separator
         */
        getSeparator() {
            return breadcrumbManager.getSeparator();
        },

        /**
         * Toggle collapsed items visibility
         */
        toggleCollapsed() {
            this.showCollapsed = !this.showCollapsed;
        },

        /**
         * Check if has collapsed items
         * @returns {boolean} Has collapsed items
         */
        hasCollapsedItems() {
            return this.getCollapsedItems().length > 0;
        },

        /**
         * Navigate to breadcrumb item
         * @param {string} itemId - Item ID
         * @param {Event} [event] - Click event
         */
        navigate(itemId, event) {
            if (event) {
                event.preventDefault();
            }

            const item = breadcrumbManager.navigateToItem(itemId);
            
            if (item && item.url) {
                window.history.pushState({}, '', item.url);
                window.dispatchEvent(new CustomEvent('navigate', { 
                    detail: { url: item.url } 
                }));
            }
        },

        /**
         * Check if item is active
         * @param {BreadcrumbItem} item - Item
         * @returns {boolean} Active state
         */
        isActive(item) {
            return item.active === true;
        },

        /**
         * Check if item is disabled
         * @param {BreadcrumbItem} item - Item
         * @returns {boolean} Disabled state
         */
        isDisabled(item) {
            return item.disabled === true;
        },

        /**
         * Check if item is ellipsis
         * @param {BreadcrumbItem} item - Item
         * @returns {boolean} Is ellipsis
         */
        isEllipsis(item) {
            return item.id === 'ellipsis';
        },

        /**
         * Get htmx attributes for item
         * @param {BreadcrumbItem} item - Item
         * @returns {Object} htmx attributes
         */
        getHtmxAttrs(item) {
            if (!item.htmx) return {};
            
            const attrs = {};
            if (item.htmx.get) attrs['hx-get'] = item.htmx.get;
            if (item.htmx.post) attrs['hx-post'] = item.htmx.post;
            if (item.htmx.target) attrs['hx-target'] = item.htmx.target;
            if (item.htmx.swap) attrs['hx-swap'] = item.htmx.swap;
            
            return attrs;
        },

        /**
         * Get configuration
         * @returns {BreadcrumbConfig} Configuration
         */
        getConfig() {
            return breadcrumbManager.getConfig();
        },

        /**
         * Update items
         * @param {BreadcrumbItem[]} items - New items
         */
        updateItems(items) {
            breadcrumbManager.setItems(items);
        },

        /**
         * Add item
         * @param {BreadcrumbItem} item - Item to add
         */
        addItem(item) {
            breadcrumbManager.addItem(item);
        }
    });
}

/**
 * Initialize Breadcrumb
 * @param {BreadcrumbConfig} config - Breadcrumb configuration
 * @param {Object} [dependencies] - External dependencies
 * @returns {Object} Breadcrumb instance
 */
function initializeBreadcrumb(config, dependencies = {}) {
    const manager = new BreadcrumbManager(config, dependencies);
    const store = createBreadcrumbStore(manager);

    return {
        manager,
        store
    };
}

// Export to window
window.ERPAdmin = window.ERPAdmin || {};
window.ERPAdmin.Breadcrumb = {
    BreadcrumbManager,
    createBreadcrumbStore,
    initializeBreadcrumb
};
