/**
 * @fileoverview Enhanced Tree View Configuration Script for Flowbite
 * Manages hierarchical tree structures with Alpine.js and htmx integration
 * @version 2.0.0
 */

/**
 * Constants
 */
const TREE_CONSTANTS = {
    EVENTS: {
        SELECT: 'tree:select',
        EXPAND: 'tree:expand',
        COLLAPSE: 'tree:collapse',
        CHECK: 'tree:check',
        LOAD: 'tree:load',
        DROP: 'tree:drop',
        ERROR: 'tree:error'
    },
    DEFAULT_ICONS: {
        FILE: 'fa fa-file',
        FOLDER: 'fa fa-folder',
        FOLDER_OPEN: 'fa fa-folder-open',
        EXPAND: 'fa fa-chevron-down',
        COLLAPSE: 'fa fa-chevron-right'
    },
    LIMITS: {
        MAX_SEARCH_RESULTS: 100,
        SEARCH_DEBOUNCE_MS: 300
    }
};

/**
 * @typedef {Object} TreeNode
 * @property {string} id - Unique identifier
 * @property {string} label - Display label
 * @property {string} [icon] - Icon class
 * @property {string} [url] - Navigation URL
 * @property {boolean} [expanded] - Expanded state
 * @property {boolean} [selected] - Selected state
 * @property {boolean} [disabled] - Disabled state
 * @property {boolean} [loading] - Loading state
 * @property {TreeNode[]} [children] - Child nodes
 * @property {string} [badge] - Badge text
 * @property {string} [badgeColor] - Badge color (primary, success, danger, warning, info)
 * @property {Object} [htmx] - htmx configuration for lazy loading
 * @property {string} [htmx.get] - htmx GET URL for children
 * @property {string} [htmx.post] - htmx POST URL
 * @property {string} [htmx.target] - htmx target selector
 * @property {string} [htmx.swap] - htmx swap strategy
 * @property {string} [htmx.trigger] - htmx trigger event
 * @property {Object} [checkbox] - Checkbox configuration
 * @property {boolean} [checkbox.enabled] - Enable checkbox
 * @property {boolean} [checkbox.checked] - Checked state
 * @property {boolean} [checkbox.indeterminate] - Indeterminate state
 * @property {Array<ContextMenuItem>} [contextMenu] - Context menu items
 * @property {Object} [meta] - Additional metadata
 * @property {number} [level] - Tree level (auto-calculated)
 * @property {string} [parentId] - Parent node ID
 */

/**
 * @typedef {Object} ContextMenuItem
 * @property {string} id - Menu item ID
 * @property {string} label - Display label
 * @property {string} [icon] - Icon class
 * @property {Function} action - Click handler
 * @property {boolean} [disabled] - Disabled state
 * @property {string} [separator] - Separator type
 */

/**
 * @typedef {Object} TreeViewConfig
 * @property {TreeNode[]} nodes - Tree nodes
 * @property {boolean} [multiSelect] - Enable multi-selection
 * @property {boolean} [showCheckboxes] - Show checkboxes
 * @property {boolean} [showIcons] - Show icons
 * @property {boolean} [showLines] - Show connecting lines
 * @property {boolean} [showBadges] - Show badges
 * @property {boolean} [draggable] - Enable drag and drop
 * @property {boolean} [searchable] - Enable search
 * @property {number} [maxSearchResults] - Maximum search results
 * @property {string} [defaultIcon] - Default icon for nodes
 * @property {string} [folderIcon] - Icon for folder/parent nodes
 * @property {string} [folderOpenIcon] - Icon for open folder nodes
 * @property {string} [expandedIcon] - Icon for expanded nodes
 * @property {string} [collapsedIcon] - Icon for collapsed nodes
 * @property {boolean} [lazyLoad] - Enable lazy loading
 * @property {boolean} [keyboard] - Enable keyboard navigation
 * @property {Function} [onSelect] - Selection callback
 * @property {Function} [onExpand] - Expand callback
 * @property {Function} [onCollapse] - Collapse callback
 * @property {Function} [onCheck] - Check callback
 * @property {Function} [loadChildren] - Load children callback
 * @property {Function} [onDrop] - Drop callback
 * @property {Function} [validateDrop] - Validate drop operation
 * @property {Object} [theme] - Theme configuration
 */

/**
 * Utility Functions
 */
class TreeUtils {
    /**
     * Deep merge objects
     */
    static deepMerge(target, source) {
        const result = { ...target };
        for (const key in source) {
            if (source[key] && typeof source[key] === 'object' && !Array.isArray(source[key])) {
                result[key] = this.deepMerge(target[key] || {}, source[key]);
            } else {
                result[key] = source[key];
            }
        }
        return result;
    }

    /**
     * Sanitize URL to prevent XSS
     */
    static sanitizeUrl(url) {
        if (!url) return null;
        
        const trimmed = url.trim().toLowerCase();
        if (trimmed.startsWith('javascript:') || trimmed.startsWith('data:') || trimmed.startsWith('vbscript:')) {
            console.warn('Potentially dangerous URL blocked:', url);
            return null;
        }
        
        return url;
    }

    /**
     * Sanitize HTML to prevent XSS
     */
    static sanitizeHtml(html) {
        const div = document.createElement('div');
        div.textContent = html;
        return div.innerHTML;
    }

    /**
     * Debounce function
     */
    static debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    /**
     * Check if node can be checked
     */
    static canCheck(node) {
        return node.checkbox?.enabled && !node.disabled;
    }

    /**
     * Emit custom event
     */
    static emitEvent(eventName, detail) {
        window.dispatchEvent(new CustomEvent(eventName, { detail }));
    }
}

/**
 * Node Manager - Handles node operations
 */
class NodeManager {
    constructor() {
        this.flatMap = new Map();
    }

    /**
     * Build flat map of all nodes
     */
    buildFlatMap(nodes, level = 0, parentId = null) {
        nodes.forEach(node => {
            node.level = level;
            node.parentId = parentId;
            this.flatMap.set(node.id, node);
            
            if (node.children?.length > 0) {
                this.buildFlatMap(node.children, level + 1, node.id);
            }
        });
    }

    /**
     * Get node by ID
     */
    getNode(id) {
        return this.flatMap.get(id);
    }

    /**
     * Get all nodes
     */
    getAllNodes() {
        return Array.from(this.flatMap.values());
    }

    /**
     * Get parent node
     */
    getParent(nodeId) {
        const node = this.getNode(nodeId);
        return node?.parentId ? this.getNode(node.parentId) : null;
    }

    /**
     * Get children nodes
     */
    getChildren(nodeId) {
        const node = this.getNode(nodeId);
        return node?.children || [];
    }

    /**
     * Get siblings
     */
    getSiblings(nodeId) {
        const node = this.getNode(nodeId);
        if (!node?.parentId) {
            return [];
        }
        
        const parent = this.getParent(nodeId);
        return parent?.children?.filter(n => n.id !== nodeId) || [];
    }

    /**
     * Get path to node (ancestors)
     */
    getPath(nodeId) {
        const path = [];
        let currentId = nodeId;
        
        while (currentId) {
            const node = this.getNode(currentId);
            if (!node) break;
            path.unshift(node);
            currentId = node.parentId;
        }
        
        return path;
    }

    /**
     * Check if node has children
     */
    hasChildren(nodeId) {
        const node = this.getNode(nodeId);
        return Array.isArray(node?.children) && node.children.length > 0;
    }

    /**
     * Check if node is leaf
     */
    isLeaf(nodeId) {
        return !this.hasChildren(nodeId);
    }

    /**
     * Check if node is ancestor of another
     */
    isAncestorOf(ancestorId, descendantId) {
        let currentId = descendantId;
        
        while (currentId) {
            if (currentId === ancestorId) return true;
            const node = this.getNode(currentId);
            currentId = node?.parentId;
        }
        
        return false;
    }

    /**
     * Clear flat map
     */
    clear() {
        this.flatMap.clear();
    }

    /**
     * Update nodes
     */
    updateNodes(nodes) {
        this.clear();
        this.buildFlatMap(nodes);
    }
}

/**
 * State Manager - Handles tree state
 */
class StateManager {
    constructor() {
        this.expandedNodes = new Set();
        this.selectedNodes = [];
        this.checkedNodes = new Set();
        this.focusedNode = null;
    }

    /**
     * Initialize state from nodes
     */
    initialize(nodeManager) {
        nodeManager.getAllNodes().forEach(node => {
            if (node.expanded) {
                this.expandedNodes.add(node.id);
            }
            if (node.selected) {
                this.selectedNodes.push(node.id);
            }
            if (node.checkbox?.checked) {
                this.checkedNodes.add(node.id);
            }
        });
    }

    /**
     * Check if expanded
     */
    isExpanded(nodeId) {
        return this.expandedNodes.has(nodeId);
    }

    /**
     * Set expanded state
     */
    setExpanded(nodeId, expanded) {
        if (expanded) {
            this.expandedNodes.add(nodeId);
        } else {
            this.expandedNodes.delete(nodeId);
        }
    }

    /**
     * Check if selected
     */
    isSelected(nodeId) {
        return this.selectedNodes.includes(nodeId);
    }

    /**
     * Set selected state
     */
    setSelected(nodeId, selected, multiSelect = false) {
        if (selected) {
            if (!multiSelect) {
                this.selectedNodes = [nodeId];
            } else if (!this.selectedNodes.includes(nodeId)) {
                this.selectedNodes.push(nodeId);
            }
        } else {
            this.selectedNodes = this.selectedNodes.filter(id => id !== nodeId);
        }
    }

    /**
     * Clear selection
     */
    clearSelection() {
        this.selectedNodes = [];
    }

    /**
     * Check if checked
     */
    isChecked(nodeId) {
        return this.checkedNodes.has(nodeId);
    }

    /**
     * Set checked state
     */
    setChecked(nodeId, checked) {
        if (checked) {
            this.checkedNodes.add(nodeId);
        } else {
            this.checkedNodes.delete(nodeId);
        }
    }

    /**
     * Get checked nodes
     */
    getCheckedNodeIds() {
        return Array.from(this.checkedNodes);
    }

    /**
     * Set focused node
     */
    setFocus(nodeId) {
        this.focusedNode = nodeId;
    }

    /**
     * Get focused node
     */
    getFocus() {
        return this.focusedNode;
    }

    /**
     * Expand all
     */
    expandAll(nodeManager) {
        nodeManager.getAllNodes().forEach(node => {
            if (nodeManager.hasChildren(node.id)) {
                this.expandedNodes.add(node.id);
            }
        });
    }

    /**
     * Collapse all
     */
    collapseAll() {
        this.expandedNodes.clear();
    }

    /**
     * Reset state
     */
    reset() {
        this.expandedNodes.clear();
        this.selectedNodes = [];
        this.checkedNodes.clear();
        this.focusedNode = null;
    }
}

/**
 * Tree View Manager Class
 */
class TreeViewManager {
    constructor(config, dependencies = {}) {
        this.config = this._mergeDefaults(config);
        this.apiManager = dependencies.apiManager;
        
        this.nodeManager = new NodeManager();
        this.stateManager = new StateManager();
        
        this._initialize();
    }

    /**
     * Initialize tree
     */
    _initialize() {
        this.nodeManager.buildFlatMap(this.config.nodes);
        this.stateManager.initialize(this.nodeManager);
    }

    /**
     * Merge with default configuration
     */
    _mergeDefaults(config) {
        const defaults = {
            nodes: [],
            multiSelect: false,
            showCheckboxes: false,
            showIcons: true,
            showLines: true,
            showBadges: true,
            draggable: false,
            searchable: false,
            keyboard: true,
            maxSearchResults: TREE_CONSTANTS.LIMITS.MAX_SEARCH_RESULTS,
            defaultIcon: TREE_CONSTANTS.DEFAULT_ICONS.FILE,
            folderIcon: TREE_CONSTANTS.DEFAULT_ICONS.FOLDER,
            folderOpenIcon: TREE_CONSTANTS.DEFAULT_ICONS.FOLDER_OPEN,
            expandedIcon: TREE_CONSTANTS.DEFAULT_ICONS.EXPAND,
            collapsedIcon: TREE_CONSTANTS.DEFAULT_ICONS.COLLAPSE,
            lazyLoad: false,
            onSelect: null,
            onExpand: null,
            onCollapse: null,
            onCheck: null,
            onDrop: null,
            loadChildren: null,
            validateDrop: null,
            theme: {
                indent: 20,
                lineColor: '#d1d5db'
            }
        };

        return TreeUtils.deepMerge(defaults, config);
    }

    // Delegation methods to NodeManager
    getNode(id) { return this.nodeManager.getNode(id); }
    getAllNodes() { return this.nodeManager.getAllNodes(); }
    getParent(nodeId) { return this.nodeManager.getParent(nodeId); }
    getChildren(nodeId) { return this.nodeManager.getChildren(nodeId); }
    getSiblings(nodeId) { return this.nodeManager.getSiblings(nodeId); }
    getPath(nodeId) { return this.nodeManager.getPath(nodeId); }
    hasChildren(nodeId) { return this.nodeManager.hasChildren(nodeId); }
    isLeaf(nodeId) { return this.nodeManager.isLeaf(nodeId); }

    // Delegation methods to StateManager
    isExpanded(nodeId) { return this.stateManager.isExpanded(nodeId); }
    isSelected(nodeId) { return this.stateManager.isSelected(nodeId); }
    isChecked(nodeId) { return this.stateManager.isChecked(nodeId); }

    /**
     * Get visible nodes
     */
    getVisibleNodes(nodes = this.config.nodes) {
        const visible = [];
        
        nodes.forEach(node => {
            visible.push(node);
            
            if (this.isExpanded(node.id) && node.children?.length > 0) {
                visible.push(...this.getVisibleNodes(node.children));
            }
        });
        
        return visible;
    }

    /**
     * Expand node
     */
    async expand(nodeId) {
        const node = this.getNode(nodeId);
        if (!node) return;

        // Lazy load children if needed
        if (this.config.lazyLoad && !node.children && node.htmx?.get) {
            node.loading = true;
            
            try {
                const children = await this._loadChildren(node);
                node.children = children;
                this.nodeManager.buildFlatMap(children, node.level + 1, node.id);
                
                TreeUtils.emitEvent(TREE_CONSTANTS.EVENTS.LOAD, { node, children });
            } catch (error) {
                console.error('Failed to load children:', error);
                TreeUtils.emitEvent(TREE_CONSTANTS.EVENTS.ERROR, { node, error });
                return;
            } finally {
                node.loading = false;
            }
        }

        node.expanded = true;
        this.stateManager.setExpanded(nodeId, true);

        if (this.config.onExpand) {
            this.config.onExpand(node);
        }

        TreeUtils.emitEvent(TREE_CONSTANTS.EVENTS.EXPAND, { node });
    }

    /**
     * Collapse node
     */
    collapse(nodeId) {
        const node = this.getNode(nodeId);
        if (!node) return;

        node.expanded = false;
        this.stateManager.setExpanded(nodeId, false);

        if (this.config.onCollapse) {
            this.config.onCollapse(node);
        }

        TreeUtils.emitEvent(TREE_CONSTANTS.EVENTS.COLLAPSE, { node });
    }

    /**
     * Toggle node expansion
     */
    async toggle(nodeId) {
        if (this.isExpanded(nodeId)) {
            this.collapse(nodeId);
        } else {
            await this.expand(nodeId);
        }
    }

    /**
     * Load children via API or callback
     */
    async _loadChildren(node) {
        if (this.config.loadChildren) {
            return this.config.loadChildren(node);
        }

        if (this.apiManager && node.htmx?.get) {
            const url = TreeUtils.sanitizeUrl(node.htmx.get);
            if (!url) throw new Error('Invalid URL');
            return this.apiManager.get(url);
        }

        return [];
    }

    /**
     * Select node
     */
    select(nodeId, append = false) {
        const node = this.getNode(nodeId);
        if (!node || node.disabled) return;

        const multiSelect = this.config.multiSelect && append;

        if (!multiSelect) {
            // Clear previous selection
            this.stateManager.selectedNodes.forEach(id => {
                const n = this.getNode(id);
                if (n) n.selected = false;
            });
            this.stateManager.clearSelection();
        }

        node.selected = true;
        this.stateManager.setSelected(nodeId, true, multiSelect);
        this.stateManager.setFocus(nodeId);

        if (this.config.onSelect) {
            this.config.onSelect(node, this.getSelectedNodes());
        }

        TreeUtils.emitEvent(TREE_CONSTANTS.EVENTS.SELECT, { 
            node, 
            selected: this.getSelectedNodes() 
        });
    }

    /**
     * Deselect node
     */
    deselect(nodeId) {
        const node = this.getNode(nodeId);
        if (!node) return;

        node.selected = false;
        this.stateManager.setSelected(nodeId, false);
    }

    /**
     * Get selected nodes
     */
    getSelectedNodes() {
        return this.stateManager.selectedNodes
            .map(id => this.getNode(id))
            .filter(Boolean);
    }

    /**
     * Check/uncheck node
     */
    check(nodeId, checked, cascade = true) {
        const node = this.getNode(nodeId);
        if (!node || !TreeUtils.canCheck(node)) return;

        node.checkbox.checked = checked;
        node.checkbox.indeterminate = false;
        this.stateManager.setChecked(nodeId, checked);

        // Cascade to children
        if (cascade && node.children) {
            this._cascadeCheck(node.children, checked);
        }

        // Update parent indeterminate state
        this._updateParentCheckState(node.parentId);

        if (this.config.onCheck) {
            this.config.onCheck(node, checked);
        }

        TreeUtils.emitEvent(TREE_CONSTANTS.EVENTS.CHECK, { node, checked });
    }

    /**
     * Cascade check state to children
     */
    _cascadeCheck(children, checked) {
        children.forEach(child => {
            if (TreeUtils.canCheck(child)) {
                child.checkbox.checked = checked;
                child.checkbox.indeterminate = false;
                this.stateManager.setChecked(child.id, checked);

                if (child.children) {
                    this._cascadeCheck(child.children, checked);
                }
            }
        });
    }

    /**
     * Update parent check state with proper indeterminate handling
     */
    _updateParentCheckState(parentId) {
        if (!parentId) return;

        const parent = this.getNode(parentId);
        if (!parent || !TreeUtils.canCheck(parent)) return;

        const children = parent.children || [];
        const checkableChildren = children.filter(TreeUtils.canCheck);
        
        if (checkableChildren.length === 0) return;

        // Count checked and indeterminate children
        const checkedCount = checkableChildren.filter(c => c.checkbox?.checked).length;
        const indeterminateCount = checkableChildren.filter(c => c.checkbox?.indeterminate).length;
        
        const hasChecked = checkedCount > 0 || indeterminateCount > 0;
        const allChecked = checkedCount === checkableChildren.length && indeterminateCount === 0;

        parent.checkbox.checked = allChecked;
        parent.checkbox.indeterminate = hasChecked && !allChecked;

        this.stateManager.setChecked(parent.id, allChecked);

        // Continue up the tree
        this._updateParentCheckState(parent.parentId);
    }

    /**
     * Get checked nodes
     */
    getCheckedNodes(leafOnly = false) {
        const checked = this.stateManager.getCheckedNodeIds()
            .map(id => this.getNode(id))
            .filter(Boolean);

        if (leafOnly) {
            return checked.filter(node => this.isLeaf(node.id));
        }

        return checked;
    }

    /**
     * Search nodes
     */
    search(query) {
        if (!query) return [];

        const lowerQuery = query.toLowerCase();
        const results = this.getAllNodes().filter(node =>
            node.label.toLowerCase().includes(lowerQuery)
        );

        // Limit results
        return results.slice(0, this.config.maxSearchResults);
    }

    /**
     * Expand path to node (async to handle lazy loading)
     */
    async expandPathTo(nodeId) {
        const path = this.getPath(nodeId);
        
        for (const node of path) {
            if (node.id !== nodeId && this.hasChildren(node.id)) {
                await this.expand(node.id);
            }
        }
    }

    /**
     * Expand all nodes
     */
    expandAll() {
        this.getAllNodes().forEach(node => {
            if (this.hasChildren(node.id)) {
                node.expanded = true;
            }
        });
        this.stateManager.expandAll(this.nodeManager);
    }

    /**
     * Collapse all nodes
     */
    collapseAll() {
        this.getAllNodes().forEach(node => {
            node.expanded = false;
        });
        this.stateManager.collapseAll();
    }

    /**
     * Validate drop operation
     */
    validateDrop(draggedNode, targetNode) {
        // Prevent dropping onto self
        if (draggedNode.id === targetNode.id) return false;

        // Prevent dropping onto descendants
        if (this.nodeManager.isAncestorOf(draggedNode.id, targetNode.id)) return false;

        // Custom validation
        if (this.config.validateDrop) {
            return this.config.validateDrop(draggedNode, targetNode);
        }

        return true;
    }

    /**
     * Get next visible node (for keyboard navigation)
     */
    getNextVisibleNode(currentId) {
        const visible = this.getVisibleNodes();
        const index = visible.findIndex(n => n.id === currentId);
        return index >= 0 && index < visible.length - 1 ? visible[index + 1] : null;
    }

    /**
     * Get previous visible node (for keyboard navigation)
     */
    getPreviousVisibleNode(currentId) {
        const visible = this.getVisibleNodes();
        const index = visible.findIndex(n => n.id === currentId);
        return index > 0 ? visible[index - 1] : null;
    }

    /**
     * Get configuration
     */
    getConfig() {
        return this.config;
    }

    /**
     * Update nodes
     */
    updateNodes(nodes) {
        this.config.nodes = nodes;
        this.nodeManager.updateNodes(nodes);
        this.stateManager.reset();
        this.stateManager.initialize(this.nodeManager);
    }
}

/**
 * Create Alpine.js Tree View Store
 */
function createTreeViewStore(treeManager) {
    return () => ({
        searchQuery: '',
        searchResults: [],
        draggedNode: null,
        dropTarget: null,
        searchDebounced: null,

        /**
         * Initialize tree view store
         */
        init() {
            // Setup drag and drop if enabled
            if (treeManager.getConfig().draggable) {
                this.setupDragAndDrop();
            }

            // Setup keyboard navigation if enabled
            if (treeManager.getConfig().keyboard) {
                this.setupKeyboardNavigation();
            }

            // Setup debounced search
            this.searchDebounced = TreeUtils.debounce(
                (query) => this.performSearch(query),
                TREE_CONSTANTS.LIMITS.SEARCH_DEBOUNCE_MS
            );
        },

        /**
         * Get visible nodes
         */
        getVisibleNodes() {
            if (this.searchQuery && this.searchResults.length > 0) {
                return this.searchResults;
            }
            return treeManager.getVisibleNodes();
        },

        /**
         * Get all root nodes
         */
        getRootNodes() {
            return treeManager.getConfig().nodes;
        },

        /**
         * Toggle node expansion
         */
        async toggleNode(nodeId, event) {
            if (event) {
                event.stopPropagation();
            }
            await treeManager.toggle(nodeId);
        },

        /**
         * Select node
         */
        selectNode(nodeId, event) {
            const append = event?.ctrlKey || event?.metaKey;
            treeManager.select(nodeId, append);
        },

        /**
         * Check/uncheck node
         */
        checkNode(nodeId, checked, event) {
            if (event) {
                event.stopPropagation();
            }
            treeManager.check(nodeId, checked);
        },

        /**
         * Check if node is expanded
         */
        isExpanded(nodeId) {
            return treeManager.isExpanded(nodeId);
        },

        /**
         * Check if node is selected
         */
        isSelected(node) {
            return node.selected === true;
        },

        /**
         * Check if node is loading
         */
        isLoading(node) {
            return node.loading === true;
        },

        /**
         * Check if node has children
         */
        hasChildren(nodeId) {
            return treeManager.hasChildren(nodeId);
        },

        /**
         * Get node icon
         */
        getNodeIcon(node) {
            if (node.icon) return node.icon;

            const config = treeManager.getConfig();
            if (this.hasChildren(node.id)) {
                return this.isExpanded(node.id) 
                    ? config.folderOpenIcon 
                    : config.folderIcon;
            }

            return config.defaultIcon;
        },

        /**
         * Get expand/collapse icon
         */
        getExpandIcon(nodeId) {
            const config = treeManager.getConfig();
            return this.isExpanded(nodeId) 
                ? config.expandedIcon 
                : config.collapsedIcon;
        },

        /**
         * Get indent style for node
         */
        getIndentStyle(node) {
            const config = treeManager.getConfig();
            return {
                paddingLeft: `${node.level * config.theme.indent}px`
            };
        },

        /**
         * Get badge color class
         */
        getBadgeClass(node) {
            if (!node.badge) return '';
            
            const colorMap = {
                primary: 'bg-blue-100 text-blue-800',
                success: 'bg-green-100 text-green-800',
                danger: 'bg-red-100 text-red-800',
                warning: 'bg-yellow-100 text-yellow-800',
                info: 'bg-gray-100 text-gray-800'
            };

            return colorMap[node.badgeColor] || colorMap.info;
        },

        /**
         * Handle search input
         */
        handleSearch(query) {
            this.searchQuery = query;
            this.searchDebounced(query);
        },

        /**
         * Perform search
         */
        async performSearch(query) {
            if (!query) {
                this.searchResults = [];
                return;
            }

            const results = treeManager.search(query);
            this.searchResults = results;

            // Expand path to all results
            for (const node of results) {
                await treeManager.expandPathTo(node.id);
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
         * Expand all nodes
         */
        expandAll() {
            treeManager.expandAll();
        },

        /**
         * Collapse all nodes
         */
        collapseAll() {
            treeManager.collapseAll();
        },

        /**
         * Get selected nodes
         */
        getSelectedNodes() {
            return treeManager.getSelectedNodes();
        },

        /**
         * Get checked nodes
         */
        getCheckedNodes(leafOnly = false) {
            return treeManager.getCheckedNodes(leafOnly);
        },

        /**
         * Setup drag and drop handlers
         */
        setupDragAndDrop() {
            // Drag start
            this.handleDragStart = (nodeId, event) => {
                this.draggedNode = treeManager.getNode(nodeId);
                event.dataTransfer.effectAllowed = 'move';
                event.dataTransfer.setData('text/plain', nodeId);
                event.target.classList.add('opacity-50');
            };

            // Drag end
            this.handleDragEnd = (event) => {
                event.target.classList.remove('opacity-50');
                this.draggedNode = null;
                this.dropTarget = null;
            };

            // Drag over
            this.handleDragOver = (nodeId, event) => {
                event.preventDefault();
                
                const targetNode = treeManager.getNode(nodeId);
                if (!this.draggedNode || !targetNode) return;

                // Validate drop
                if (treeManager.validateDrop(this.draggedNode, targetNode)) {
                    this.dropTarget = nodeId;
                    event.dataTransfer.dropEffect = 'move';
                } else {
                    event.dataTransfer.dropEffect = 'none';
                }
            };

            // Drag leave
            this.handleDragLeave = () => {
                this.dropTarget = null;
            };

            // Drop
            this.handleDrop = (nodeId, event) => {
                event.preventDefault();
                event.stopPropagation();
                
                const targetNode = treeManager.getNode(nodeId);
                
                if (this.draggedNode && targetNode) {
                    if (treeManager.validateDrop(this.draggedNode, targetNode)) {
                        this.onNodeDrop(this.draggedNode, targetNode);
                    }
                }

                this.draggedNode = null;
                this.dropTarget = null;
            };
        },

        /**
         * Handle node drop
         */
        onNodeDrop(draggedNode, targetNode) {
            // Fire custom callback if provided
            if (treeManager.getConfig().onDrop) {
                treeManager.getConfig().onDrop(draggedNode, targetNode);
            }

            // Emit event
            TreeUtils.emitEvent(TREE_CONSTANTS.EVENTS.DROP, { 
                dragged: draggedNode, 
                target: targetNode 
            });
        },

        /**
         * Check if node is drop target
         */
        isDropTarget(nodeId) {
            return this.dropTarget === nodeId;
        },

        /**
         * Setup keyboard navigation
         */
        setupKeyboardNavigation() {
            this.handleKeyDown = async (event) => {
                const focusedId = treeManager.stateManager.getFocus();
                if (!focusedId) return;

                const focusedNode = treeManager.getNode(focusedId);
                if (!focusedNode) return;

                let handled = true;

                switch (event.key) {
                    case 'ArrowDown':
                        event.preventDefault();
                        this.navigateDown(focusedId);
                        break;

                    case 'ArrowUp':
                        event.preventDefault();
                        this.navigateUp(focusedId);
                        break;

                    case 'ArrowRight':
                        event.preventDefault();
                        if (treeManager.hasChildren(focusedId)) {
                            if (!treeManager.isExpanded(focusedId)) {
                                await treeManager.expand(focusedId);
                            } else {
                                // Move to first child
                                const children = treeManager.getChildren(focusedId);
                                if (children.length > 0) {
                                    treeManager.select(children[0].id);
                                }
                            }
                        }
                        break;

                    case 'ArrowLeft':
                        event.preventDefault();
                        if (treeManager.isExpanded(focusedId)) {
                            treeManager.collapse(focusedId);
                        } else {
                            // Move to parent
                            const parent = treeManager.getParent(focusedId);
                            if (parent) {
                                treeManager.select(parent.id);
                            }
                        }
                        break;

                    case 'Enter':
                    case ' ':
                        event.preventDefault();
                        if (focusedNode.checkbox?.enabled) {
                            treeManager.check(focusedId, !focusedNode.checkbox.checked);
                        } else {
                            await treeManager.toggle(focusedId);
                        }
                        break;

                    case 'Home':
                        event.preventDefault();
                        const roots = treeManager.getConfig().nodes;
                        if (roots.length > 0) {
                            treeManager.select(roots[0].id);
                        }
                        break;

                    case 'End':
                        event.preventDefault();
                        const visible = treeManager.getVisibleNodes();
                        if (visible.length > 0) {
                            treeManager.select(visible[visible.length - 1].id);
                        }
                        break;

                    case '*':
                        event.preventDefault();
                        // Expand all siblings
                        const siblings = treeManager.getSiblings(focusedId);
                        siblings.forEach(async (sibling) => {
                            if (treeManager.hasChildren(sibling.id)) {
                                await treeManager.expand(sibling.id);
                            }
                        });
                        break;

                    default:
                        handled = false;
                }

                return handled;
            };
        },

        /**
         * Navigate down in tree
         */
        navigateDown(currentId) {
            const next = treeManager.getNextVisibleNode(currentId);
            if (next) {
                treeManager.select(next.id);
            }
        },

        /**
         * Navigate up in tree
         */
        navigateUp(currentId) {
            const prev = treeManager.getPreviousVisibleNode(currentId);
            if (prev) {
                treeManager.select(prev.id);
            }
        },

        /**
         * Get htmx attributes for node
         */
        getHtmxAttrs(node) {
            if (!node.htmx) return {};
            
            const attrs = {};
            if (node.htmx.get) attrs['hx-get'] = node.htmx.get;
            if (node.htmx.post) attrs['hx-post'] = node.htmx.post;
            if (node.htmx.target) attrs['hx-target'] = node.htmx.target;
            if (node.htmx.swap) attrs['hx-swap'] = node.htmx.swap;
            if (node.htmx.trigger) attrs['hx-trigger'] = node.htmx.trigger;
            
            return attrs;
        },

        /**
         * Show context menu
         */
        showContextMenu(node, event) {
            if (!node.contextMenu || node.contextMenu.length === 0) return;
            
            event.preventDefault();
            event.stopPropagation();

            // Emit event for context menu handling
            TreeUtils.emitEvent('tree:contextmenu', {
                node,
                x: event.clientX,
                y: event.clientY,
                items: node.contextMenu
            });
        },

        /**
         * Handle context menu item click
         */
        handleContextMenuItem(node, item, event) {
            if (event) {
                event.preventDefault();
                event.stopPropagation();
            }

            if (item.disabled) return;

            if (typeof item.action === 'function') {
                item.action(node, item);
            }
        },

        /**
         * Get configuration
         */
        getConfig() {
            return treeManager.getConfig();
        },

        /**
         * Navigate to node URL
         */
        navigate(node, event) {
            if (event) {
                event.preventDefault();
            }

            if (node.disabled) return;

            const url = TreeUtils.sanitizeUrl(node.url);
            if (!url) {
                console.warn('Invalid or missing URL for node:', node.id);
                return;
            }

            try {
                window.history.pushState({}, '', url);
                TreeUtils.emitEvent('tree:navigate', { url, node });
            } catch (error) {
                console.error('Navigation failed:', error);
                TreeUtils.emitEvent(TREE_CONSTANTS.EVENTS.ERROR, { 
                    node, 
                    error,
                    context: 'navigation' 
                });
            }
        },

        /**
         * Refresh node (reload children)
         */
        async refreshNode(nodeId) {
            const node = treeManager.getNode(nodeId);
            if (!node) return;

            // Clear children and reload
            node.children = null;
            node.expanded = false;
            treeManager.stateManager.setExpanded(nodeId, false);

            await treeManager.expand(nodeId);
        },

        /**
         * Get node by ID (utility for templates)
         */
        getNode(nodeId) {
            return treeManager.getNode(nodeId);
        },

        /**
         * Focus on node (programmatically)
         */
        focusNode(nodeId) {
            const node = treeManager.getNode(nodeId);
            if (!node || node.disabled) return;

            treeManager.select(nodeId);
            
            // Scroll into view if needed
            const element = document.querySelector(`[data-node-id="${nodeId}"]`);
            if (element) {
                element.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            }
        },

        /**
         * Export tree state (for persistence)
         */
        exportState() {
            return {
                expanded: Array.from(treeManager.stateManager.expandedNodes),
                selected: treeManager.stateManager.selectedNodes,
                checked: Array.from(treeManager.stateManager.checkedNodes)
            };
        },

        /**
         * Import tree state (restore from persistence)
         */
        async importState(state) {
            if (!state) return;

            // Restore expanded state
            if (state.expanded) {
                for (const nodeId of state.expanded) {
                    await treeManager.expand(nodeId);
                }
            }

            // Restore selected state
            if (state.selected && state.selected.length > 0) {
                state.selected.forEach(nodeId => {
                    const node = treeManager.getNode(nodeId);
                    if (node) {
                        node.selected = true;
                        treeManager.stateManager.setSelected(nodeId, true, true);
                    }
                });
            }

            // Restore checked state
            if (state.checked) {
                state.checked.forEach(nodeId => {
                    const node = treeManager.getNode(nodeId);
                    if (node && TreeUtils.canCheck(node)) {
                        treeManager.check(nodeId, true, false);
                    }
                });
            }
        }
    });
}

/**
 * Initialize Tree View
 */
function initializeTreeView(config, dependencies = {}) {
    const manager = new TreeViewManager(config, dependencies);
    const store = createTreeViewStore(manager);

    return {
        manager,
        store,
        constants: TREE_CONSTANTS
    };
}

/**
 * Create a simple tree view with minimal configuration
 */
function createSimpleTreeView(nodes, options = {}) {
    return initializeTreeView({
        nodes,
        ...options
    });
}

/**
 * Create a file browser tree view
 */
function createFileBrowserTreeView(nodes, options = {}) {
    return initializeTreeView({
        nodes,
        showIcons: true,
        showLines: true,
        draggable: true,
        searchable: true,
        keyboard: true,
        defaultIcon: 'fa fa-file',
        folderIcon: 'fa fa-folder',
        folderOpenIcon: 'fa fa-folder-open',
        ...options
    });
}

/**
 * Create a checkbox tree view
 */
function createCheckboxTreeView(nodes, options = {}) {
    return initializeTreeView({
        nodes,
        showCheckboxes: true,
        showIcons: true,
        multiSelect: true,
        keyboard: true,
        ...options
    });
}

/**
 * Create a lazy-loading tree view
 */
function createLazyTreeView(nodes, loadChildren, options = {}) {
    return initializeTreeView({
        nodes,
        lazyLoad: true,
        loadChildren,
        showIcons: true,
        keyboard: true,
        ...options
    });
}

// Export to window
window.ERPAdmin = window.ERPAdmin || {};
window.ERPAdmin.TreeView = {
    // Core classes
    TreeViewManager,
    NodeManager,
    StateManager,
    TreeUtils,
    
    // Factory functions
    createTreeViewStore,
    initializeTreeView,
    
    // Convenience factories
    createSimpleTreeView,
    createFileBrowserTreeView,
    createCheckboxTreeView,
    createLazyTreeView,
    
    // Constants
    CONSTANTS: TREE_CONSTANTS
};

// Export as ES module if supported
if (typeof module !== 'undefined' && module.exports) {
    module.exports = window.ERPAdmin.TreeView;
}
