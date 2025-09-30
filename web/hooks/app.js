// ERP Admin Configuration
window.APP_CONFIG = {
    // API Configuration
    api: {
        baseURL: 'http://localhost:8080',
        timeout: 30000,
        headers: {
            'Content-Type': 'application/json'
        }
    },
    
    // App Information
    app: {
        name: 'Awo ERP Admin',
        version: '1.0.0',
        description: 'Enterprise Resource Planning Administration Panel'
    },
    
    // Theme Configuration
    theme: 'cxd', // cxd, antd, dark
    
    // Feature Flags
    features: {
        tenantManagement: true,
        userManagement: true,
        financeModule: true,
        analytics: true,
        auditLogs: true,
        featureFlags: true,
        systemSettings: true
    },
    
    // Menu Configuration
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
                { label: 'All Tenants', url: '/tenants/list' },
                { label: 'Create Tenant', url: '/tenants/create' },
                { label: 'Tenant Analytics', url: '/tenants/analytics' }
            ]
        },
        {
            label: 'User Management',
            icon: 'fa fa-users',
            url: '/users',
            children: [
                { label: 'All Users', url: '/users/list' },
                { label: 'User Roles', url: '/users/roles' },
                { label: 'Permissions', url: '/users/permissions' }
            ]
        },
        {
            label: 'Finance',
            icon: 'fa fa-chart-line',
            url: '/finance',
            children: [
                { label: 'Accounts', url: '/finance/accounts' },
                { label: 'Transactions', url: '/finance/transactions' },
                { label: 'Reports', url: '/finance/reports' }
            ]
        },
        {
            label: 'System',
            icon: 'fa fa-cog',
            url: '/system',
            children: [
                { label: 'Feature Flags', url: '/system/feature-flags' },
                { label: 'Audit Logs', url: '/system/audit' },
                { label: 'Settings', url: '/system/settings' }
            ]
        }
    ],
    
    // Default tenant for admin operations
    defaultTenant: null,
    
    // Pagination defaults
    pagination: {
        pageSize: 20,
        showSizeChanger: true,
        showQuickJumper: true
    }
};

// API Helper Functions
window.API = {
    // Get headers with tenant context
    getHeaders: function(tenantId = null) {
        const headers = { ...window.APP_CONFIG.api.headers };
        if (tenantId) {
            headers['X-Tenant-ID'] = tenantId;
        }
        return headers;
    },
    
    // Build full API URL
    url: function(path) {
        return window.APP_CONFIG.api.baseURL + (path.startsWith('/') ? path : '/' + path);
    },
    
    // Handle API responses
    handleResponse: function(response) {
        if (response.ok) {
            return response.json();
        }
        throw new Error(`API Error: ${response.status} ${response.statusText}`);
    },
    
    // Standard fetch with config
    fetch: function(url, options = {}) {
        return fetch(window.API.url(url), {
            timeout: window.APP_CONFIG.api.timeout,
            ...options,
            headers: {
                ...window.API.getHeaders(),
                ...options.headers
            }
        }).then(window.API.handleResponse);
    }
};