// Main Application JavaScript
(function() {
    'use strict';

    console.log('Main.js loading...');

    // Wait for all dependencies to load
    function initializeApp() {
        console.log('Initializing app...');

        // Check if required dependencies are loaded
        if (typeof amis === 'undefined') {
            console.error('Amis library not loaded');
            showError('Failed to load Amis library');
            return;
        }

        if (typeof window.APP_CONFIG === 'undefined') {
            console.error('App config not loaded');
            showError('Failed to load application configuration');
            return;
        }

        if (typeof window.TenantPages === 'undefined') {
            console.error('Tenant pages not loaded');
            showError('Failed to load tenant pages configuration');
            return;
        }

        console.log('All dependencies loaded successfully');

        // Initialize the admin application
        setupAdminApp();
    }

    function setupAdminApp() {
        console.log('Setting up admin app...');

        try {
            const appElement = document.getElementById('app');
            if (!appElement) {
                throw new Error('App element not found');
            }

            // Create the main application schema
            const appSchema = createAppSchema();

            // Render the application
            const amisScoped = amis.embed(appElement, appSchema, {
                theme: window.APP_CONFIG.theme || 'cxd',
                locale: 'en'
            });

            console.log('Admin app initialized successfully');

        } catch (error) {
            console.error('Error initializing app:', error);
            showError('Failed to initialize admin interface: ' + error.message);
        }
    }

    function createAppSchema() {
        return {
            type: "app",
            brandName: window.APP_CONFIG.app.name,
            logo: "/assets/logo.png",
            header: {
                type: "wrapper",
                className: "w-full",
                body: [
                    {
                        type: "flex",
                        justify: "space-between",
                        alignItems: "center",
                        items: [
                            {
                                type: "tpl",
                                tpl: `<h1 style="margin:0;color:white;"><i class="fas fa-cogs"></i> ${window.APP_CONFIG.app.name}</h1>`
                            },
                            {
                                type: "dropdown-button",
                                label: "Admin",
                                level: "link",
                                className: "text-white",
                                buttons: [
                                    {
                                        type: "button",
                                        label: "Profile",
                                        actionType: "url",
                                        url: "/profile"
                                    },
                                    {
                                        type: "button",
                                        label: "Logout",
                                        actionType: "url",
                                        url: "/logout"
                                    }
                                ]
                            }
                        ]
                    }
                ]
            },
            asideResizor: true,
            aside: [
                {
                    type: "nav",
                    stacked: true,
                    links: window.APP_CONFIG.menu.map(item => ({
                        label: item.label,
                        to: item.url,
                        icon: item.icon,
                        children: item.children && item.children.length > 0 ? item.children.map(child => ({
                            label: child.label,
                            to: child.url
                        })) : undefined
                    }))
                }
            ],
            pages: [
                {
                    url: "/dashboard",
                    schema: createDashboardPage()
                },
                {
                    url: "/tenants",
                    redirect: "/tenants/list"
                },
                {
                    url: "/tenants/list",
                    schema: window.TenantPages.list
                },
                {
                    url: "/tenants/create",
                    schema: {
                        type: "page",
                        title: "Create Tenant",
                        body: window.TenantPages.createDialog.body
                    }
                },
                {
                    url: "/tenants/analytics",
                    schema: createTenantAnalyticsPage()
                },
                {
                    url: "/users",
                    redirect: "/users/list"
                },
                {
                    url: "/users/list",
                    schema: createUsersPage()
                },
                {
                    url: "/users/roles",
                    schema: createUserRolesPage()
                },
                {
                    url: "/users/permissions",
                    schema: createPermissionsPage()
                },
                {
                    url: "/finance",
                    redirect: "/finance/accounts"
                },
                {
                    url: "/finance/accounts",
                    schema: createFinanceAccountsPage()
                },
                {
                    url: "/finance/transactions",
                    schema: createTransactionsPage()
                },
                {
                    url: "/finance/reports",
                    schema: createFinanceReportsPage()
                },
                {
                    url: "/system",
                    redirect: "/system/feature-flags"
                },
                {
                    url: "/system/feature-flags",
                    schema: createFeatureFlagsPage()
                },
                {
                    url: "/system/audit",
                    schema: createAuditLogsPage()
                },
                {
                    url: "/system/settings",
                    schema: createSettingsPage()
                }
            ]
        };
    }

    function createDashboardPage() {
        return {
            type: "page",
            title: "Dashboard",
            body: [
                {
                    type: "grid",
                    columns: [
                        {
                            type: "card",
                            className: "mb-4",
                            header: {
                                title: "Total Tenants",
                                subTitle: "Active tenant accounts"
                            },
                            body: {
                                type: "tpl",
                                tpl: '<div class="text-center"><h2 class="text-primary">42</h2></div>'
                            }
                        },
                        {
                            type: "card",
                            className: "mb-4",
                            header: {
                                title: "Active Users",
                                subTitle: "Currently logged in"
                            },
                            body: {
                                type: "tpl",
                                tpl: '<div class="text-center"><h2 class="text-success">1,247</h2></div>'
                            }
                        },
                        {
                            type: "card",
                            className: "mb-4",
                            header: {
                                title: "System Uptime",
                                subTitle: "Current system status"
                            },
                            body: {
                                type: "tpl",
                                tpl: '<div class="text-center"><h2 class="text-info">98.5%</h2></div>'
                            }
                        },
                        {
                            type: "card",
                            className: "mb-4",
                            header: {
                                title: "Storage Used",
                                subTitle: "Total disk usage"
                            },
                            body: {
                                type: "tpl",
                                tpl: '<div class="text-center"><h2 class="text-warning">2.3GB</h2></div>'
                            }
                        }
                    ]
                },
                {
                    type: "card",
                    header: {
                        title: "Recent Activity"
                    },
                    body: {
                        type: "table",
                        data: {
                            items: [
                                {
                                    time: "10:30 AM",
                                    action: "New tenant created",
                                    user: "admin@awo.com",
                                    status: "success"
                                },
                                {
                                    time: "09:45 AM",
                                    action: "User login",
                                    user: "john.doe@company.com",
                                    status: "success"
                                },
                                {
                                    time: "09:15 AM",
                                    action: "System backup",
                                    user: "System",
                                    status: "complete"
                                }
                            ]
                        },
                        columns: [
                            { name: "time", label: "Time" },
                            { name: "action", label: "Action" },
                            { name: "user", label: "User" },
                            {
                                name: "status",
                                label: "Status",
                                type: "mapping",
                                mapping: {
                                    success: "<span class='label label-success'>Success</span>",
                                    complete: "<span class='label label-info'>Complete</span>",
                                    error: "<span class='label label-danger'>Error</span>"
                                }
                            }
                        ]
                    }
                }
            ]
        };
    }

    function createTenantAnalyticsPage() {
        return {
            type: "page",
            title: "Tenant Analytics",
            body: [
                {
                    type: "tpl",
                    tpl: "<p>Tenant analytics dashboard coming soon...</p>"
                }
            ]
        };
    }

    function createUsersPage() {
        return {
            type: "page",
            title: "User Management",
            body: [
                {
                    type: "tpl",
                    tpl: "<p>User management interface coming soon...</p>"
                }
            ]
        };
    }

    function createUserRolesPage() {
        return {
            type: "page",
            title: "User Roles",
            body: [
                {
                    type: "tpl",
                    tpl: "<p>User roles management coming soon...</p>"
                }
            ]
        };
    }

    function createPermissionsPage() {
        return {
            type: "page",
            title: "Permissions",
            body: [
                {
                    type: "tpl",
                    tpl: "<p>Permissions management coming soon...</p>"
                }
            ]
        };
    }

    function createFinanceAccountsPage() {
        return {
            type: "page",
            title: "Finance Accounts",
            body: [
                {
                    type: "tpl",
                    tpl: "<p>Finance accounts management coming soon...</p>"
                }
            ]
        };
    }

    function createTransactionsPage() {
        return {
            type: "page",
            title: "Transactions",
            body: [
                {
                    type: "tpl",
                    tpl: "<p>Transaction management coming soon...</p>"
                }
            ]
        };
    }

    function createFinanceReportsPage() {
        return {
            type: "page",
            title: "Finance Reports",
            body: [
                {
                    type: "tpl",
                    tpl: "<p>Finance reports coming soon...</p>"
                }
            ]
        };
    }

    function createFeatureFlagsPage() {
        return {
            type: "page",
            title: "Feature Flags",
            body: [
                {
                    type: "tpl",
                    tpl: "<p>Feature flags management coming soon...</p>"
                }
            ]
        };
    }

    function createAuditLogsPage() {
        return {
            type: "page",
            title: "Audit Logs",
            body: [
                {
                    type: "log",
                    height: 400,
                    source: {
                        method: "get",
                        url: "/api/v1/audit/logs"
                    }
                }
            ]
        };
    }

    function createSettingsPage() {
        return {
            type: "page",
            title: "System Settings",
            body: [
                {
                    type: "form",
                    body: [
                        {
                            type: "input-text",
                            name: "systemName",
                            label: "System Name",
                            value: "Awo ERP System"
                        },
                        {
                            type: "select",
                            name: "language",
                            label: "Default Language",
                            options: [
                                { label: "English", value: "en" },
                                { label: "Spanish", value: "es" },
                                { label: "French", value: "fr" }
                            ],
                            value: "en"
                        },
                        {
                            type: "input-number",
                            name: "sessionTimeout",
                            label: "Session Timeout (minutes)",
                            value: 30
                        }
                    ]
                }
            ]
        };
    }

    function showError(message) {
        const appElement = document.getElementById('app');
        if (appElement) {
            appElement.innerHTML = `
                <div style="padding: 2rem; text-align: center;">
                    <div style="background: #fed7d7; color: #742a2a; padding: 1rem; border-radius: 4px; border-left: 4px solid #f56565;">
                        <strong>Error:</strong> ${message}
                    </div>
                    <p style="margin-top: 1rem; color: #666;">
                        Please check the browser console for more details.
                    </p>
                </div>
            `;
        }
    }

    // Initialize when DOM is ready and all scripts are loaded
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', function() {
            setTimeout(initializeApp, 100);
        });
    } else {
        setTimeout(initializeApp, 100);
    }

    console.log('Main.js loaded successfully');

})();
