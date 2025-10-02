/**
 * @fileoverview Enhanced Tree View Configuration Script for Flowbite (TypeScript)
 * Manages hierarchical tree structures with Alpine.js and htmx integration
 * @version 2.0.0
 */

// Types
interface TreeConstants {
    EVENTS: {
        SELECT: string;
        EXPAND: string;
        COLLAPSE: string;
        CHECK: string;
        LOAD: string;
        DROP: string;
        ERROR: string;
    };
    DEFAULT_ICONS: {
        FILE: string;
        FOLDER: string;
        FOLDER_OPEN: string;
        EXPAND: string;
        COLLAPSE: string;
    };
    LIMITS: {
        MAX_SEARCH_RESULTS: number;
        SEARCH_DEBOUNCE_MS: number;
    };
}

interface HtmxTreeConfig {
    get?: string;
    post?: string;
    target?: string;
    swap?: string;
    trigger?: string;
}

interface TreeCheckbox {
    enabled: boolean;
    checked?: boolean;
    indeterminate?: boolean;
}

interface ContextMenuItem {
    id: string;
    label: string;
    icon?: string;
    action: (node: TreeNode, item: ContextMenuItem) => void;
    disabled?: boolean;
    separator?: string;
}

interface TreeNode {
    id: string;
    label: string;
    icon?: string;
    url?: string;
    expanded?: boolean;
    selected?: boolean;
    disabled?: boolean;
    loading?: boolean;
    children?: TreeNode[];
    badge?: string;
    badgeColor?: 'primary' | 'success' | 'danger' | 'warning' | 'info';
    htmx?: HtmxTreeConfig;
    checkbox?: TreeCheckbox;
    contextMenu?: ContextMenuItem[];
    meta?: Record<string, any>;
    level?: number;
    parentId?: string;
}

interface TreeTheme {
    indent: number;
    lineColor: string;
}

interface TreeViewConfig {
    nodes: TreeNode[];
    multiSelect?: boolean;
    showCheckboxes?: boolean;
    showIcons?: boolean;
    showLines?: boolean;
    showBadges?: boolean;
    draggable?: boolean;
    searchable?: boolean;
    maxSearchResults?: number;
    defaultIcon?: string;
    folderIcon?: string;
    folderOpenIcon?: string;
    expandedIcon?: string;
    collapsedIcon?: string;
    lazyLoad?: boolean;
    keyboard?: boolean;
    onSelect?: (node: TreeNode, selected: TreeNode[]) => void;
    onExpand?: (node: TreeNode) => void;
    onCollapse?: (node: TreeNode) => void;
    onCheck?: (node: TreeNode, checked: boolean) => void;
    loadChildren?: (node: TreeNode) => Promise<TreeNode[]>;
    onDrop?: (draggedNode: TreeNode, targetNode: TreeNode) => void;
    validateDrop?: (draggedNode: TreeNode, targetNode: TreeNode) => boolean;
    theme?: TreeTheme;
}

interface TreeDependencies {
    apiManager?: any;
}

interface TreeState {
    expanded: string[];
    selected: string[];
    checked: string[];
}

interface TreeViewStore {
    searchQuery: string;
    searchResults: TreeNode[];
    draggedNode: TreeNode | null;
    dropTarget: string | null;
    searchDebounced: ((query: string) => void) | null;
    
    init(): void;
    getVisibleNodes(): TreeNode[];
    getRootNodes(): TreeNode[];
    toggleNode(nodeId: string, event?: Event): Promise<void>;
    selectNode(nodeId: string, event?: MouseEvent | KeyboardEvent): void;
    checkNode(nodeId: string, checked: boolean, event?: Event): void;
    isExpanded(nodeId: string): boolean;
    isSelected(node: TreeNode): boolean;
    isLoading(node: TreeNode): boolean;
    hasChildren(nodeId: string): boolean;
    getNodeIcon(node: TreeNode): string;
    getExpandIcon(nodeId: string): string;
    getIndentStyle(node: TreeNode): { paddingLeft: string };
    getBadgeClass(node: TreeNode): string;
    handleSearch(query: string): void;
    performSearch(query: string): Promise<void>;
    clearSearch(): void;
    expandAll(): void;
    collapseAll(): void;
    getSelectedNodes(): TreeNode[];
    getCheckedNodes(leafOnly?: boolean): TreeNode[];
    setupDragAndDrop(): void;
    onNodeDrop(draggedNode: TreeNode, targetNode: TreeNode): void;
    isDropTarget(nodeId: string): boolean;
    setupKeyboardNavigation(): void;
    handleKeyDown?: (event: KeyboardEvent) => Promise<boolean>;
    navigateDown(currentId: string): void;
    navigateUp(currentId: string): void;
    getHtmxAttrs(node: TreeNode): Record<string, string>;
    showContextMenu(node: TreeNode, event: MouseEvent): void;
    handleContextMenuItem(node: TreeNode, item: ContextMenuItem, event?: Event): void;
    getConfig(): TreeViewConfig;
    navigate(node: TreeNode, event?: Event): void;
    refreshNode(nodeId: string): Promise<void>;
    getNode(nodeId: string): TreeNode | undefined;
    focusNode(nodeId: string): void;
    exportState(): TreeState;
    importState(state: TreeState): Promise<void>;
    
    // Drag and drop handlers
    handleDragStart?: (nodeId: string, event: DragEvent) => void;
    handleDragEnd?: (event: DragEvent) => void;
    handleDragOver?: (nodeId: string, event: DragEvent) => void;
    handleDragLeave?: () => void;
    handleDrop?: (nodeId: string, event: DragEvent) => void;
}

interface TreeViewInstance {
    manager: TreeViewManager;
    store: () => TreeViewStore;
    constants: TreeConstants;
}

// Export to make this a module
export {};

/**
 * Constants
 */
const TREE_CONSTANTS: TreeConstants = {
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
 * Utility Functions
 */
class TreeUtils {
    /**
     * Deep merge objects
     */
    static deepMerge(target: any, source: any): any {
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
    static sanitizeUrl(url?: string): string | null {
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
    static sanitizeHtml(html: string): string {
        const div = document.createElement('div');
        div.textContent = html;
        return div.innerHTML;
    }

    /**
     * Debounce function
     */
    static debounce(func: (...args: any[]) => void, wait: number): (...args: any[]) => void {
        let timeout: any;
        return function executedFunction(...args: any[]) {
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
    static canCheck(node: TreeNode): boolean {
        return node.checkbox?.enabled === true && !node.disabled;
    }

    /**
     * Emit custom event
     */
    static emitEvent(eventName: string, detail: any): void {
        window.dispatchEvent(new CustomEvent(eventName, { detail }));
    }
}

/**
 * Node Manager - Handles node operations
 */
class NodeManager {
    private flatMap: Map<string, TreeNode> = new Map();

    /**
     * Build flat map of all nodes
     */
    buildFlatMap(nodes: TreeNode[], level: number = 0, parentId: string | null = null): void {
        nodes.forEach(node => {
            node.level = level;
            node.parentId = parentId || undefined;
            this.flatMap.set(node.id, node);
            
            if (node.children?.length && node.children.length > 0) {
                this.buildFlatMap(node.children, level + 1, node.id);
            }
        });
    }

    /**
     * Get node by ID
     */
    getNode(id: string): TreeNode | undefined {
        return this.flatMap.get(id);
    }

    /**
     * Get all nodes
     */
    getAllNodes(): TreeNode[] {
        return Array.from(this.flatMap.values());
    }

    /**
     * Get parent node
     */
    getParent(nodeId: string): TreeNode | null {
        const node = this.getNode(nodeId);
        return node?.parentId ? this.getNode(node.parentId) || null : null;
    }

    /**
     * Get children nodes
     */
    getChildren(nodeId: string): TreeNode[] {
        const node = this.getNode(nodeId);
        return node?.children || [];
    }

    /**
     * Get siblings
     */
    getSiblings(nodeId: string): TreeNode[] {
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
    getPath(nodeId: string): TreeNode[] {
        const path: TreeNode[] = [];
        let currentId: string | undefined = nodeId;
        
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
    hasChildren(nodeId: string): boolean {
        const node = this.getNode(nodeId);
        return Array.isArray(node?.children) && node.children.length > 0;
    }

    /**
     * Check if node is leaf
     */
    isLeaf(nodeId: string): boolean {
        return !this.hasChildren(nodeId);
    }

    /**
     * Check if node is ancestor of another
     */
    isAncestorOf(ancestorId: string, descendantId: string): boolean {
        let currentId: string | undefined = descendantId;
        
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
    clear(): void {
        this.flatMap.clear();
    }

    /**
     * Update nodes
     */
    updateNodes(nodes: TreeNode[]): void {
        this.clear();
        this.buildFlatMap(nodes);
    }
}

/**
 * State Manager - Handles tree state
 */
class StateManager {
    expandedNodes: Set<string> = new Set();
    selectedNodes: string[] = [];
    checkedNodes: Set<string> = new Set();
    focusedNode: string | null = null;

    /**
     * Initialize state from nodes
     */
    initialize(nodeManager: NodeManager): void {
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
    isExpanded(nodeId: string): boolean {
        return this.expandedNodes.has(nodeId);
    }

    /**
     * Set expanded state
     */
    setExpanded(nodeId: string, expanded: boolean): void {
        if (expanded) {
            this.expandedNodes.add(nodeId);
        } else {
            this.expandedNodes.delete(nodeId);
        }
    }

    /**
     * Check if selected
     */
    isSelected(nodeId: string): boolean {
        return this.selectedNodes.includes(nodeId);
    }

    /**
     * Set selected state
     */
    setSelected(nodeId: string, selected: boolean, multiSelect: boolean = false): void {
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
    clearSelection(): void {
        this.selectedNodes = [];
    }

    /**
     * Check if checked
     */
    isChecked(nodeId: string): boolean {
        return this.checkedNodes.has(nodeId);
    }

    /**
     * Set checked state
     */
    setChecked(nodeId: string, checked: boolean): void {
        if (checked) {
            this.checkedNodes.add(nodeId);
        } else {
            this.checkedNodes.delete(nodeId);
        }
    }

    /**
     * Get checked nodes
     */
    getCheckedNodeIds(): string[] {
        return Array.from(this.checkedNodes);
    }

    /**
     * Set focused node
     */
    setFocus(nodeId: string): void {
        this.focusedNode = nodeId;
    }

    /**
     * Get focused node
     */
    getFocus(): string | null {
        return this.focusedNode;
    }

    /**
     * Expand all
     */
    expandAll(nodeManager: NodeManager): void {
        nodeManager.getAllNodes().forEach(node => {
            if (nodeManager.hasChildren(node.id)) {
                this.expandedNodes.add(node.id);
            }
        });
    }

    /**
     * Collapse all
     */
    collapseAll(): void {
        this.expandedNodes.clear();
    }

    /**
     * Reset state
     */
    reset(): void {
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
    private config: TreeViewConfig;
    private apiManager: any;
    public nodeManager: NodeManager;
    public stateManager: StateManager;

    constructor(config: TreeViewConfig, dependencies: TreeDependencies = {}) {
        this.config = this._mergeDefaults(config);
        this.apiManager = dependencies.apiManager;
        
        this.nodeManager = new NodeManager();
        this.stateManager = new StateManager();
        
        this._initialize();
    }

    /**
     * Initialize tree
     */
    private _initialize(): void {
        this.nodeManager.buildFlatMap(this.config.nodes);
        this.stateManager.initialize(this.nodeManager);
    }

    /**
     * Merge with default configuration
     */
    private _mergeDefaults(config: TreeViewConfig): TreeViewConfig {
        const defaults: TreeViewConfig = {
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
            onSelect: undefined,
            onExpand: undefined,
            onCollapse: undefined,
            onCheck: undefined,
            onDrop: undefined,
            loadChildren: undefined,
            validateDrop: undefined,
            theme: {
                indent: 20,
                lineColor: '#d1d5db'
            }
        };

        return TreeUtils.deepMerge(defaults, config);
    }

    // Delegation methods to NodeManager
    getNode(id: string): TreeNode | undefined { return this.nodeManager.getNode(id); }
    getAllNodes(): TreeNode[] { return this.nodeManager.getAllNodes(); }
    getParent(nodeId: string): TreeNode | null { return this.nodeManager.getParent(nodeId); }
    getChildren(nodeId: string): TreeNode[] { return this.nodeManager.getChildren(nodeId); }
    getSiblings(nodeId: string): TreeNode[] { return this.nodeManager.getSiblings(nodeId); }
    getPath(nodeId: string): TreeNode[] { return this.nodeManager.getPath(nodeId); }
    hasChildren(nodeId: string): boolean { return this.nodeManager.hasChildren(nodeId); }
    isLeaf(nodeId: string): boolean { return this.nodeManager.isLeaf(nodeId); }

    // Delegation methods to StateManager
    isExpanded(nodeId: string): boolean { return this.stateManager.isExpanded(nodeId); }
    isSelected(nodeId: string): boolean { return this.stateManager.isSelected(nodeId); }
    isChecked(nodeId: string): boolean { return this.stateManager.isChecked(nodeId); }

    /**
     * Get visible nodes
     */
    getVisibleNodes(nodes: TreeNode[] = this.config.nodes): TreeNode[] {
        const visible: TreeNode[] = [];
        
        nodes.forEach(node => {
            visible.push(node);
            
            if (this.isExpanded(node.id) && node.children?.length && node.children.length > 0) {
                visible.push(...this.getVisibleNodes(node.children));
            }
        });
        
        return visible;
    }

    /**
     * Expand node
     */
    async expand(nodeId: string): Promise<void> {
        const node = this.getNode(nodeId);
        if (!node) return;

        // Lazy load children if needed
        if (this.config.lazyLoad && !node.children && node.htmx?.get) {
            node.loading = true;
            
            try {
                const children = await this._loadChildren(node);
                node.children = children;
                this.nodeManager.buildFlatMap(children, (node.level || 0) + 1, node.id);
                
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
    collapse(nodeId: string): void {
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
    async toggle(nodeId: string): Promise<void> {
        if (this.isExpanded(nodeId)) {
            this.collapse(nodeId);
        } else {
            await this.expand(nodeId);
        }
    }

    /**
     * Load children via API or callback
     */
    private async _loadChildren(node: TreeNode): Promise<TreeNode[]> {
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
    select(nodeId: string, append: boolean = false): void {
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
    deselect(nodeId: string): void {
        const node = this.getNode(nodeId);
        if (!node) return;

        node.selected = false;
        this.stateManager.setSelected(nodeId, false);
    }

    /**
     * Get selected nodes
     */
    getSelectedNodes(): TreeNode[] {
        return this.stateManager.selectedNodes
            .map(id => this.getNode(id))
            .filter((node): node is TreeNode => Boolean(node));
    }

    /**
     * Check/uncheck node
     */
    check(nodeId: string, checked: boolean, cascade: boolean = true): void {
        const node = this.getNode(nodeId);
        if (!node || !TreeUtils.canCheck(node) || !node.checkbox) return;

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
    private _cascadeCheck(children: TreeNode[], checked: boolean): void {
        children.forEach(child => {
            if (TreeUtils.canCheck(child) && child.checkbox) {
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
    private _updateParentCheckState(parentId?: string): void {
        if (!parentId) return;

        const parent = this.getNode(parentId);
        if (!parent || !TreeUtils.canCheck(parent) || !parent.checkbox) return;

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
    getCheckedNodes(leafOnly: boolean = false): TreeNode[] {
        const checked = this.stateManager.getCheckedNodeIds()
            .map(id => this.getNode(id))
            .filter((node): node is TreeNode => Boolean(node));

        if (leafOnly) {
            return checked.filter(node => this.isLeaf(node.id));
        }

        return checked;
    }

    /**
     * Search nodes
     */
    search(query: string): TreeNode[] {
        if (!query) return [];

        const lowerQuery = query.toLowerCase();
        const results = this.getAllNodes().filter(node =>
            node.label.toLowerCase().includes(lowerQuery)
        );

        // Limit results
        return results.slice(0, this.config.maxSearchResults || TREE_CONSTANTS.LIMITS.MAX_SEARCH_RESULTS);
    }

    /**
     * Expand path to node (async to handle lazy loading)
     */
    async expandPathTo(nodeId: string): Promise<void> {
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
    expandAll(): void {
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
    collapseAll(): void {
        this.getAllNodes().forEach(node => {
            node.expanded = false;
        });
        this.stateManager.collapseAll();
    }

    /**
     * Validate drop operation
     */
    validateDrop(draggedNode: TreeNode, targetNode: TreeNode): boolean {
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
    getNextVisibleNode(currentId: string): TreeNode | null {
        const visible = this.getVisibleNodes();
        const index = visible.findIndex(n => n.id === currentId);
        return index >= 0 && index < visible.length - 1 ? visible[index + 1] : null;
    }

    /**
     * Get previous visible node (for keyboard navigation)
     */
    getPreviousVisibleNode(currentId: string): TreeNode | null {
        const visible = this.getVisibleNodes();
        const index = visible.findIndex(n => n.id === currentId);
        return index > 0 ? visible[index - 1] : null;
    }

    /**
     * Get configuration
     */
    getConfig(): TreeViewConfig {
        return this.config;
    }

    /**
     * Update nodes
     */
    updateNodes(nodes: TreeNode[]): void {
        this.config.nodes = nodes;
        this.nodeManager.updateNodes(nodes);
        this.stateManager.reset();
        this.stateManager.initialize(this.nodeManager);
    }
}

/**
 * Create Alpine.js Tree View Store
 */
function createTreeViewStore(treeManager: TreeViewManager): () => TreeViewStore {
    return () => ({
        searchQuery: '',
        searchResults: [],
        draggedNode: null,
        dropTarget: null,
        searchDebounced: null,

        /**
         * Initialize tree view store
         */
        init(): void {
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
                (query: string) => this.performSearch(query),
                TREE_CONSTANTS.LIMITS.SEARCH_DEBOUNCE_MS
            );
        },

        /**
         * Get visible nodes
         */
        getVisibleNodes(): TreeNode[] {
            if (this.searchQuery && this.searchResults.length > 0) {
                return this.searchResults;
            }
            return treeManager.getVisibleNodes();
        },

        /**
         * Get all root nodes
         */
        getRootNodes(): TreeNode[] {
            return treeManager.getConfig().nodes;
        },

        /**
         * Toggle node expansion
         */
        async toggleNode(nodeId: string, event?: Event): Promise<void> {
            if (event) {
                event.stopPropagation();
            }
            await treeManager.toggle(nodeId);
        },

        /**
         * Select node
         */
        selectNode(nodeId: string, event?: MouseEvent | KeyboardEvent): void {
            const append = (event as MouseEvent)?.ctrlKey || (event as MouseEvent)?.metaKey;
            treeManager.select(nodeId, append);
        },

        /**
         * Check/uncheck node
         */
        checkNode(nodeId: string, checked: boolean, event?: Event): void {
            if (event) {
                event.stopPropagation();
            }
            treeManager.check(nodeId, checked);
        },

        /**
         * Check if node is expanded
         */
        isExpanded(nodeId: string): boolean {
            return treeManager.isExpanded(nodeId);
        },

        /**
         * Check if node is selected
         */
        isSelected(node: TreeNode): boolean {
            return node.selected === true;
        },

        /**
         * Check if node is loading
         */
        isLoading(node: TreeNode): boolean {
            return node.loading === true;
        },

        /**
         * Check if node has children
         */
        hasChildren(nodeId: string): boolean {
            return treeManager.hasChildren(nodeId);
        },

        /**
         * Get node icon
         */
        getNodeIcon(node: TreeNode): string {
            if (node.icon) return node.icon;

            const config = treeManager.getConfig();
            if (this.hasChildren(node.id)) {
                return this.isExpanded(node.id) 
                    ? config.folderOpenIcon || TREE_CONSTANTS.DEFAULT_ICONS.FOLDER_OPEN
                    : config.folderIcon || TREE_CONSTANTS.DEFAULT_ICONS.FOLDER;
            }

            return config.defaultIcon || TREE_CONSTANTS.DEFAULT_ICONS.FILE;
        },

        /**
         * Get expand/collapse icon
         */
        getExpandIcon(nodeId: string): string {
            const config = treeManager.getConfig();
            return this.isExpanded(nodeId) 
                ? config.expandedIcon || TREE_CONSTANTS.DEFAULT_ICONS.EXPAND
                : config.collapsedIcon || TREE_CONSTANTS.DEFAULT_ICONS.COLLAPSE;
        },

        /**
         * Get indent style for node
         */
        getIndentStyle(node: TreeNode): { paddingLeft: string } {
            const config = treeManager.getConfig();
            const indent = config.theme?.indent || 20;
            return {
                paddingLeft: `${(node.level || 0) * indent}px`
            };
        },

        /**
         * Get badge color class
         */
        getBadgeClass(node: TreeNode): string {
            if (!node.badge) return '';
            
            const colorMap: Record<string, string> = {
                primary: 'bg-blue-100 text-blue-800',
                success: 'bg-green-100 text-green-800',
                danger: 'bg-red-100 text-red-800',
                warning: 'bg-yellow-100 text-yellow-800',
                info: 'bg-gray-100 text-gray-800'
            };

            return colorMap[node.badgeColor || 'info'] || colorMap.info;
        },

        /**
         * Handle search input
         */
        handleSearch(query: string): void {
            this.searchQuery = query;
            if (this.searchDebounced) {
                this.searchDebounced(query);
            }
        },

        /**
         * Perform search
         */
        async performSearch(query: string): Promise<void> {
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
        clearSearch(): void {
            this.searchQuery = '';
            this.searchResults = [];
        },

        /**
         * Expand all nodes
         */
        expandAll(): void {
            treeManager.expandAll();
        },

        /**
         * Collapse all nodes
         */
        collapseAll(): void {
            treeManager.collapseAll();
        },

        /**
         * Get selected nodes
         */
        getSelectedNodes(): TreeNode[] {
            return treeManager.getSelectedNodes();
        },

        /**
         * Get checked nodes
         */
        getCheckedNodes(leafOnly: boolean = false): TreeNode[] {
            return treeManager.getCheckedNodes(leafOnly);
        },

        /**
         * Setup drag and drop handlers
         */
        setupDragAndDrop(): void {
            // Drag start
            this.handleDragStart = (nodeId: string, event: DragEvent) => {
                this.draggedNode = treeManager.getNode(nodeId) || null;
                event.dataTransfer!.effectAllowed = 'move';
                event.dataTransfer!.setData('text/plain', nodeId);
                (event.target as HTMLElement).classList.add('opacity-50');
            };

            // Drag end
            this.handleDragEnd = (event: DragEvent) => {
                (event.target as HTMLElement).classList.remove('opacity-50');
                this.draggedNode = null;
                this.dropTarget = null;
            };

            // Drag over
            this.handleDragOver = (nodeId: string, event: DragEvent) => {
                event.preventDefault();
                
                const targetNode = treeManager.getNode(nodeId);
                if (!this.draggedNode || !targetNode) return;

                // Validate drop
                if (treeManager.validateDrop(this.draggedNode, targetNode)) {
                    this.dropTarget = nodeId;
                    event.dataTransfer!.dropEffect = 'move';
                } else {
                    event.dataTransfer!.dropEffect = 'none';
                }
            };

            // Drag leave
            this.handleDragLeave = () => {
                this.dropTarget = null;
            };

            // Drop
            this.handleDrop = (nodeId: string, event: DragEvent) => {
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
        onNodeDrop(draggedNode: TreeNode, targetNode: TreeNode): void {
            // Fire custom callback if provided
            if (treeManager.getConfig().onDrop) {
                treeManager.getConfig().onDrop!(draggedNode, targetNode);
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
        isDropTarget(nodeId: string): boolean {
            return this.dropTarget === nodeId;
        },

        /**
         * Setup keyboard navigation
         */
        setupKeyboardNavigation(): void {
            this.handleKeyDown = async (event: KeyboardEvent): Promise<boolean> => {
                const focusedId = treeManager.stateManager.getFocus();
                if (!focusedId) return false;

                const focusedNode = treeManager.getNode(focusedId);
                if (!focusedNode) return false;

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
                        for (const sibling of siblings) {
                            if (treeManager.hasChildren(sibling.id)) {
                                await treeManager.expand(sibling.id);
                            }
                        }
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
        navigateDown(currentId: string): void {
            const next = treeManager.getNextVisibleNode(currentId);
            if (next) {
                treeManager.select(next.id);
            }
        },

        /**
         * Navigate up in tree
         */
        navigateUp(currentId: string): void {
            const prev = treeManager.getPreviousVisibleNode(currentId);
            if (prev) {
                treeManager.select(prev.id);
            }
        },

        /**
         * Get htmx attributes for node
         */
        getHtmxAttrs(node: TreeNode): Record<string, string> {
            if (!node.htmx) return {};
            
            const attrs: Record<string, string> = {};
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
        showContextMenu(node: TreeNode, event: MouseEvent): void {
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
        handleContextMenuItem(node: TreeNode, item: ContextMenuItem, event?: Event): void {
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
        getConfig(): TreeViewConfig {
            return treeManager.getConfig();
        },

        /**
         * Navigate to node URL
         */
        navigate(node: TreeNode, event?: Event): void {
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
        async refreshNode(nodeId: string): Promise<void> {
            const node = treeManager.getNode(nodeId);
            if (!node) return;

            // Clear children and reload
            node.children = undefined;
            node.expanded = false;
            treeManager.stateManager.setExpanded(nodeId, false);

            await treeManager.expand(nodeId);
        },

        /**
         * Get node by ID (utility for templates)
         */
        getNode(nodeId: string): TreeNode | undefined {
            return treeManager.getNode(nodeId);
        },

        /**
         * Focus on node (programmatically)
         */
        focusNode(nodeId: string): void {
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
        exportState(): TreeState {
            return {
                expanded: Array.from(treeManager.stateManager.expandedNodes),
                selected: treeManager.stateManager.selectedNodes,
                checked: Array.from(treeManager.stateManager.checkedNodes)
            };
        },

        /**
         * Import tree state (restore from persistence)
         */
        async importState(state: TreeState): Promise<void> {
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
function initializeTreeView(config: TreeViewConfig, dependencies: TreeDependencies = {}): TreeViewInstance {
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
function createSimpleTreeView(nodes: TreeNode[], options: Partial<TreeViewConfig> = {}): TreeViewInstance {
    return initializeTreeView({
        nodes,
        ...options
    });
}

/**
 * Create a file browser tree view
 */
function createFileBrowserTreeView(nodes: TreeNode[], options: Partial<TreeViewConfig> = {}): TreeViewInstance {
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
function createCheckboxTreeView(nodes: TreeNode[], options: Partial<TreeViewConfig> = {}): TreeViewInstance {
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
function createLazyTreeView(nodes: TreeNode[], loadChildren: (node: TreeNode) => Promise<TreeNode[]>, options: Partial<TreeViewConfig> = {}): TreeViewInstance {
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
(window as any).ERPAdmin = (window as any).ERPAdmin || {};
(window as any).ERPAdmin.TreeView = {
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