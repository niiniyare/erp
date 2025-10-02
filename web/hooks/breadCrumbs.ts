
interface HtmxConfig {
    get?: string;
    target?: string;
    swap?: string;
    post?: string;
}

interface BreadcrumbItem {
    id: string;
    label: string;
    url?: string;
    icon?: string;
    active?: boolean;
    disabled?: boolean;
    htmx?: HtmxConfig;
    meta?: Record<string, any>;
}

interface BreadcrumbConfig {
    items: BreadcrumbItem[];
    separator?: 'slash' | 'chevron' | 'arrow' | 'dot';
    homeIcon?: string;
    homeUrl?: string;
    homeLabel?: string;
    showHome?: boolean;
    maxItems?: number;
    collapsible?: boolean;
    theme?: 'light' | 'dark';
    aria?: {
        label?: string;
        current?: string;
    };
}

interface BreadcrumbDependencies {
    configManager?: any;
}

class BreadcrumbManager {
    private config: BreadcrumbConfig;
    private configManager: any;
    public history: BreadcrumbItem[][] = [];

    constructor(config: BreadcrumbConfig, dependencies: BreadcrumbDependencies = {}) {
        this.config = this._mergeDefaults(config);
        this.configManager = dependencies.configManager;
    }

    private _mergeDefaults(config: BreadcrumbConfig): BreadcrumbConfig {
        const defaults: BreadcrumbConfig = {
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

    public getSeparator(): string {
        const separators = {
            slash: '/',
            chevron: 'fa fa-chevron-right',
            arrow: 'fa fa-arrow-right',
            dot: '•'
        };
        return separators[this.config.separator!] || separators.chevron;
    }

    public getAllItems(): BreadcrumbItem[] {
        const items: BreadcrumbItem[] = [];
        
        if (this.config.showHome) {
            items.push({
                id: 'home',
                label: this.config.homeLabel!,
                url: this.config.homeUrl,
                icon: this.config.homeIcon,
                active: false
            });
        }

        return [...items, ...this.config.items];
    }

    public getVisibleItems(): BreadcrumbItem[] {
        const allItems = this.getAllItems();
        
        if (!this.config.collapsible || allItems.length <= this.config.maxItems!) {
            return allItems;
        }

        const first = allItems[0];
        const last = allItems[allItems.length - 1];
        const remaining = this.config.maxItems! - 2;
        
        const middle = allItems.slice(1, -1).slice(-remaining);

        return [
            first,
            { id: 'ellipsis', label: '...', disabled: true },
            ...middle,
            last
        ];
    }

    public getCollapsedItems(): BreadcrumbItem[] {
        const allItems = this.getAllItems();
        const visibleItems = this.getVisibleItems();
        
        if (allItems.length <= this.config.maxItems!) {
            return [];
        }

        return allItems.filter(item => 
            !visibleItems.some(visibleItem => visibleItem.id === item.id) && item.id !== 'ellipsis'
        );
    }

    public setFromPath(path: string, options: { labelFormatter?: (segment: string, index: number) => string } = {}): void {
        const segments = path.split('/').filter(Boolean);
        const items: BreadcrumbItem[] = [];
        let currentPath = '';

        segments.forEach((segment, index) => {
            currentPath += `/${segment}`;
            const isLast = index === segments.length - 1;
            
            items.push({
                id: `path-${index}`,
                label: options.labelFormatter 
                    ? options.labelFormatter(segment, index) 
                    : this._formatLabel(segment),
                url: isLast ? undefined : currentPath,
                active: isLast
            });
        });

        this.config.items = items;
    }

    private _formatLabel(segment: string): string {
        return segment
            .split('-')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1))
            .join(' ');
    }

    public addItem(item: BreadcrumbItem): void {
        this.config.items.forEach(i => i.active = false);
        this.config.items.push({ ...item, active: true });
    }

    public removeItem(id: string): void {
        this.config.items = this.config.items.filter(item => item.id !== id);
    }

    public setItems(items: BreadcrumbItem[]): void {
        this.config.items = items;
    }

    public clear(): void {
        this.config.items = [];
    }

    public navigateToItem(id: string): BreadcrumbItem | null {
        const item = this.getAllItems().find(i => i.id === id);
        if (item && !item.active && !item.disabled) {
            const index = this.config.items.findIndex(i => i.id === id);
            if (index >= 0) {
                this.config.items = this.config.items.slice(0, index + 1);
                this.config.items.forEach((i, idx) => {
                    i.active = idx === index;
                });
            }
        }
        return item || null;
    }

    public getConfig(): BreadcrumbConfig {
        return this.config;
    }

    public updateConfig(updates: Partial<BreadcrumbConfig>): void {
        this.config = this._deepMerge(this.config, updates);
    }
}

interface BreadcrumbStore {
    showCollapsed: boolean;
    init(): void;
    getItems(): BreadcrumbItem[];
    getAllItems(): BreadcrumbItem[];
    getCollapsedItems(): BreadcrumbItem[];
    getSeparator(): string;
    toggleCollapsed(): void;
    hasCollapsedItems(): boolean;
    navigate(itemId: string, event?: Event): void;
    isActive(item: BreadcrumbItem): boolean;
    isDisabled(item: BreadcrumbItem): boolean;
    isEllipsis(item: BreadcrumbItem): boolean;
    getHtmxAttrs(item: BreadcrumbItem): Record<string, string>;
    getConfig(): BreadcrumbConfig;
    updateItems(items: BreadcrumbItem[]): void;
    addItem(item: BreadcrumbItem): void;
}

function createBreadcrumbStore(breadcrumbManager: BreadcrumbManager): () => BreadcrumbStore {
    return () => ({
        showCollapsed: false,

        init() {
            const config = breadcrumbManager.getConfig();
            if (config.items.length === 0) {
                breadcrumbManager.setFromPath(window.location.pathname);
            }

            window.addEventListener('navigate', (e: any) => {
                if (e.detail?.url) {
                    breadcrumbManager.setFromPath(e.detail.url);
                }
            });
        },

        getItems() {
            return breadcrumbManager.getVisibleItems();
        },

        getAllItems() {
            return breadcrumbManager.getAllItems();
        },

        getCollapsedItems() {
            return breadcrumbManager.getCollapsedItems();
        },

        getSeparator() {
            return breadcrumbManager.getSeparator();
        },

        toggleCollapsed() {
            this.showCollapsed = !this.showCollapsed;
        },

        hasCollapsedItems() {
            return this.getCollapsedItems().length > 0;
        },

        navigate(itemId: string, event?: Event) {
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

        isActive(item: BreadcrumbItem) {
            return item.active === true;
        },

        isDisabled(item: BreadcrumbItem) {
            return item.disabled === true;
        },

        isEllipsis(item: BreadcrumbItem) {
            return item.id === 'ellipsis';
        },

        getHtmxAttrs(item: BreadcrumbItem): Record<string, string> {
            if (!item.htmx) return {};
            
            const attrs: Record<string, string> = {};
            if (item.htmx.get) attrs['hx-get'] = item.htmx.get;
            if (item.htmx.post) attrs['hx-post'] = item.htmx.post;
            if (item.htmx.target) attrs['hx-target'] = item.htmx.target;
            if (item.htmx.swap) attrs['hx-swap'] = item.htmx.swap;
            
            return attrs;
        },

        getConfig() {
            return breadcrumbManager.getConfig();
        },

        updateItems(items: BreadcrumbItem[]) {
            breadcrumbManager.setItems(items);
        },

        addItem(item: BreadcrumbItem) {
            breadcrumbManager.addItem(item);
        }
    });
}

interface BreadcrumbInstance {
    manager: BreadcrumbManager;
    store: () => BreadcrumbStore;
}

function initializeBreadcrumb(config: BreadcrumbConfig, dependencies: BreadcrumbDependencies = {}): BreadcrumbInstance {
    const manager = new BreadcrumbManager(config, dependencies);
    const store = createBreadcrumbStore(manager);

    return {
        manager,
        store
    };
}

// Export to make this a module
export {};

(window as any).ERPAdmin = (window as any).ERPAdmin || {};
(window as any).ERPAdmin.Breadcrumb = {
    BreadcrumbManager,
    createBreadcrumbStore,
    initializeBreadcrumb
};
