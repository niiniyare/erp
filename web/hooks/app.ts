// ERP Admin Configuration

interface ApiConfig {
    baseURL: string;
    timeout: number;
    headers: {
        'Content-Type': string;
    };
}

interface AppInfo {
    name: string;
    version: string;
    description: string;
}

interface FeatureFlags {
    tenantManagement: boolean;
    userManagement: boolean;
    financeModule: boolean;
    analytics: boolean;
    auditLogs: boolean;
    featureFlags: boolean;
    systemSettings: boolean;
}

interface MenuItem {
    label: string;
    icon: string;
    url: string;
    children: MenuItem[];
}

interface PaginationConfig {
    pageSize: number;
    showSizeChanger: boolean;
    showQuickJumper: boolean;
}

interface AppConfig {
    api: ApiConfig;
    app: AppInfo;
    theme: string;
    features: FeatureFlags;
    menu: MenuItem[];
    defaultTenant: string | null;
    pagination: PaginationConfig;
}

interface ApiHelper {
    getHeaders(tenantId?: string | null): Record<string, string>;
    url(path: string): string;
    handleResponse(response: Response): Promise<any>;
    fetch(url: string, options?: RequestInit): Promise<any>;
}

// Export to make this a module
export {};

window.APP_CONFIG = {
    api: {
        baseURL: 'http://localhost:8080',
        timeout: 30000,
        headers: {
            'Content-Type': 'application/json'
        }
    },
    
    app: {
        name: 'Awo ERP Admin',
        version: '1.0.0',
        description: 'Enterprise Resource Planning Administration Panel'
    },
    
    theme: 'cxd', // cxd, antd, dark
    
    features: {
        tenantManagement: true,
        userManagement: true,
        financeModule: true,
        analytics: true,
        auditLogs: true,
        featureFlags: true,
        systemSettings: true
    },
    
    menu: [
        {
            label: 'Dashboard',
            icon: 'fa fa-tachometer-alt',
            url: '/dashboard',
            children: []
        },
        {
            label: 'Tenant Management',
            icon: 'fa fa-building',
            url: '/tenants',
            children: [
                { label: 'All Tenants', url: '/tenants/list', children: [] },
                { label: 'Create Tenant', url: '/tenants/create', children: [] },
                { label: 'Tenant Analytics', url: '/tenants/analytics', children: [] }
            ]
        },
        {
            label: 'User Management',
            icon: 'fa fa-users',
            url: '/users',
            children: [
                { label: 'All Users', url: '/users/list', children: [] },
                { label: 'User Roles', url: '/users/roles', children: [] },
                { label: 'Permissions', url: '/users/permissions', children: [] }
            ]
        },
        {
            label: 'Finance',
            icon: 'fa fa-chart-line',
            url: '/finance',
            children: [
                { label: 'Accounts', url: '/finance/accounts', children: [] },
                { label: 'Transactions', url: '/finance/transactions', children: [] },
                { label: 'Reports', url: '/finance/reports', children: [] }
            ]
        },
        {
            label: 'System',
            icon: 'fa fa-cog',
            url: '/system',
            children: [
                { label: 'Feature Flags', url: '/system/feature-flags', children: [] },
                { label: 'Audit Logs', url: '/system/audit', children: [] },
                { label: 'Settings', url: '/system/settings', children: [] }
            ]
        }
    ],
    
    defaultTenant: null,
    
    pagination: {
        pageSize: 20,
        showSizeChanger: true,
        showQuickJumper: true
    }
};

window.API = {
    getHeaders: function(tenantId: string | null = null): Record<string, string> {
        const headers: Record<string, string> = { ...window.APP_CONFIG.api.headers };
        if (tenantId) {
            headers['X-Tenant-ID'] = tenantId;
        }
        return headers;
    },
    
    url: function(path: string): string {
        return window.APP_CONFIG.api.baseURL + (path.startsWith('/') ? path : '/' + path);
    },
    
    handleResponse: function(response: Response): Promise<any> {
        if (response.ok) {
            return response.json();
        }
        throw new Error(`API Error: ${response.status} ${response.statusText}`);
    },
    
    fetch: function(url: string, options: RequestInit = {}): Promise<any> {
        const fetchOptions: RequestInit & { timeout?: number } = {
            ...options,
            headers: {
                ...window.API.getHeaders(),
                ...options.headers,
            },
        };

        // The AbortController and timeout logic is a common pattern for implementing request timeouts.
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), window.APP_CONFIG.api.timeout);
        fetchOptions.signal = controller.signal;

        return fetch(window.API.url(url), fetchOptions)
            .then(response => {
                clearTimeout(timeoutId);
                return window.API.handleResponse(response);
            })
            .catch(error => {
                clearTimeout(timeoutId);
                throw error;
            });
    }
};
