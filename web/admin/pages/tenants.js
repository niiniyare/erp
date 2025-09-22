// Tenant Management Pages
window.TenantPages = {
    
    // Tenant List Page
    list: {
        type: 'page',
        title: 'Tenant Management',
        subTitle: 'Manage all tenants in the system',
        body: [
            {
                type: 'crud',
                api: {
                    method: 'GET',
                    url: '/api/v1/tenants',
                    headers: {
                        'X-Tenant-ID': '${tenantId}'
                    }
                },
                headerToolbar: [
                    {
                        type: 'button',
                        label: 'Create Tenant',
                        level: 'primary',
                        actionType: 'dialog',
                        dialog: window.TenantPages.createDialog
                    },
                    'reload',
                    {
                        type: 'export-excel',
                        label: 'Export'
                    }
                ],
                footerToolbar: ['pagination'],
                columns: [
                    {
                        name: 'id',
                        label: 'ID',
                        type: 'text',
                        width: 200,
                        copyable: true
                    },
                    {
                        name: 'name',
                        label: 'Name',
                        type: 'text',
                        searchable: true
                    },
                    {
                        name: 'subdomain',
                        label: 'Subdomain',
                        type: 'text',
                        searchable: true
                    },
                    {
                        name: 'status',
                        label: 'Status',
                        type: 'status',
                        map: {
                            'ACTIVE': { status: 'success', text: 'Active' },
                            'SUSPENDED': { status: 'warning', text: 'Suspended' },
                            'PENDING': { status: 'info', text: 'Pending' },
                            'ARCHIVED': { status: 'default', text: 'Archived' }
                        }
                    },
                    {
                        name: 'plan_type',
                        label: 'Plan',
                        type: 'tag',
                        displayMode: 'normal'
                    },
                    {
                        name: 'created_at',
                        label: 'Created',
                        type: 'datetime',
                        format: 'YYYY-MM-DD HH:mm'
                    },
                    {
                        type: 'operation',
                        label: 'Actions',
                        width: 200,
                        buttons: [
                            {
                                type: 'button',
                                icon: 'fa fa-eye',
                                tooltip: 'View Details',
                                level: 'link',
                                actionType: 'dialog',
                                dialog: window.TenantPages.viewDialog
                            },
                            {
                                type: 'button',
                                icon: 'fa fa-edit',
                                tooltip: 'Edit',
                                level: 'link',
                                actionType: 'dialog',
                                dialog: window.TenantPages.editDialog
                            },
                            {
                                type: 'dropdown-button',
                                icon: 'fa fa-ellipsis-v',
                                tooltip: 'More Actions',
                                trigger: 'click',
                                closeOnItemClick: true,
                                buttons: [
                                    {
                                        type: 'button',
                                        label: 'Suspend',
                                        icon: 'fa fa-pause',
                                        visibleOn: 'this.status === "ACTIVE"',
                                        actionType: 'ajax',
                                        api: {
                                            method: 'POST',
                                            url: '/api/v1/tenants/${id}/suspend',
                                            headers: { 'X-Tenant-ID': '${id}' },
                                            data: { reason: 'Admin suspension' }
                                        },
                                        confirmText: 'Are you sure you want to suspend this tenant?'
                                    },
                                    {
                                        type: 'button',
                                        label: 'Reactivate',
                                        icon: 'fa fa-play',
                                        visibleOn: 'this.status === "SUSPENDED"',
                                        actionType: 'ajax',
                                        api: {
                                            method: 'POST',
                                            url: '/api/v1/tenants/${id}/reactivate',
                                            headers: { 'X-Tenant-ID': '${id}' },
                                            data: { reason: 'Admin reactivation' }
                                        },
                                        confirmText: 'Are you sure you want to reactivate this tenant?'
                                    },
                                    {
                                        type: 'divider'
                                    },
                                    {
                                        type: 'button',
                                        label: 'Analytics',
                                        icon: 'fa fa-chart-bar',
                                        actionType: 'link',
                                        link: '/tenants/analytics?id=${id}'
                                    }
                                ]
                            }
                        ]
                    }
                ]
            }
        ]
    },
    
    // Create Tenant Dialog
    createDialog: {
        title: 'Create New Tenant',
        size: 'lg',
        body: {
            type: 'form',
            api: {
                method: 'POST',
                url: '/api/v1/tenants'
            },
            body: [
                {
                    type: 'grid',
                    columns: [
                        {
                            type: 'input-text',
                            name: 'name',
                            label: 'Tenant Name',
                            required: true,
                            placeholder: 'Enter tenant name'
                        },
                        {
                            type: 'input-text',
                            name: 'subdomain',
                            label: 'Subdomain',
                            required: true,
                            placeholder: 'Enter subdomain'
                        }
                    ]
                },
                {
                    type: 'grid',
                    columns: [
                        {
                            type: 'input-email',
                            name: 'email',
                            label: 'Admin Email',
                            required: true,
                            placeholder: 'admin@tenant.com'
                        },
                        {
                            type: 'input-email',
                            name: 'contact_email',
                            label: 'Contact Email',
                            required: true,
                            placeholder: 'contact@tenant.com'
                        }
                    ]
                },
                {
                    type: 'grid',
                    columns: [
                        {
                            type: 'select',
                            name: 'status',
                            label: 'Status',
                            value: 'ACTIVE',
                            options: [
                                { label: 'Active', value: 'ACTIVE' },
                                { label: 'Pending', value: 'PENDING' },
                                { label: 'Suspended', value: 'SUSPENDED' }
                            ]
                        },
                        {
                            type: 'select',
                            name: 'company_size',
                            label: 'Company Size',
                            value: 'Small',
                            options: [
                                { label: 'Startup', value: 'Startup' },
                                { label: 'Small', value: 'Small' },
                                { label: 'Medium', value: 'Medium' },
                                { label: 'Large', value: 'Large' },
                                { label: 'Enterprise', value: 'Enterprise' }
                            ]
                        }
                    ]
                },
                {
                    type: 'grid',
                    columns: [
                        {
                            type: 'input-text',
                            name: 'country_code',
                            label: 'Country Code',
                            value: 'US',
                            placeholder: 'US'
                        },
                        {
                            type: 'input-text',
                            name: 'currency_code',
                            label: 'Currency',
                            value: 'USD',
                            placeholder: 'USD'
                        }
                    ]
                },
                {
                    type: 'input-text',
                    name: 'industry',
                    label: 'Industry',
                    placeholder: 'Technology, Healthcare, etc.'
                },
                {
                    type: 'input-text',
                    name: 'timezone',
                    label: 'Timezone',
                    value: 'UTC',
                    placeholder: 'UTC'
                }
            ]
        }
    },
    
    // View Tenant Dialog
    viewDialog: {
        title: 'Tenant Details',
        size: 'lg',
        body: {
            type: 'service',
            api: '/api/v1/tenants/${id}',
            body: [
                {
                    type: 'panel',
                    title: 'Basic Information',
                    body: [
                        {
                            type: 'grid',
                            columns: [
                                {
                                    type: 'static',
                                    name: 'name',
                                    label: 'Name'
                                },
                                {
                                    type: 'static',
                                    name: 'subdomain',
                                    label: 'Subdomain'
                                }
                            ]
                        },
                        {
                            type: 'grid',
                            columns: [
                                {
                                    type: 'static',
                                    name: 'status',
                                    label: 'Status'
                                },
                                {
                                    type: 'static',
                                    name: 'plan_type',
                                    label: 'Plan Type'
                                }
                            ]
                        }
                    ]
                },
                {
                    type: 'panel',
                    title: 'Timestamps',
                    body: [
                        {
                            type: 'grid',
                            columns: [
                                {
                                    type: 'static-datetime',
                                    name: 'created_at',
                                    label: 'Created At',
                                    format: 'YYYY-MM-DD HH:mm:ss'
                                },
                                {
                                    type: 'static-datetime',
                                    name: 'updated_at',
                                    label: 'Updated At',
                                    format: 'YYYY-MM-DD HH:mm:ss'
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    },
    
    // Edit Tenant Dialog
    editDialog: {
        title: 'Edit Tenant',
        size: 'lg',
        body: {
            type: 'form',
            initApi: '/api/v1/tenants/${id}',
            api: {
                method: 'PUT',
                url: '/api/v1/tenants/${id}',
                headers: { 'X-Tenant-ID': '${id}' }
            },
            body: [
                {
                    type: 'static',
                    name: 'id',
                    label: 'Tenant ID'
                },
                {
                    type: 'grid',
                    columns: [
                        {
                            type: 'input-text',
                            name: 'name',
                            label: 'Name',
                            required: true
                        },
                        {
                            type: 'select',
                            name: 'status',
                            label: 'Status',
                            options: [
                                { label: 'Active', value: 'ACTIVE' },
                                { label: 'Pending', value: 'PENDING' },
                                { label: 'Suspended', value: 'SUSPENDED' },
                                { label: 'Archived', value: 'ARCHIVED' }
                            ]
                        }
                    ]
                },
                {
                    type: 'input-text',
                    name: 'industry',
                    label: 'Industry'
                }
            ]
        }
    }
    
};