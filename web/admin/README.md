# ERP Admin UI

This directory contains the admin interface for the ERP system, built with [Baidu amis](https://baidu.github.io/amis/).

## 🚀 **Quick Start**

The admin UI is embedded directly into the Go server and served at:
```
http://localhost:8080/admin/
```

## 📁 **Directory Structure**

```
web/admin/
├── index.html          # Main HTML entry point
├── config/
│   └── app-config.js   # Application configuration
├── js/
│   └── main.js         # Main application logic
├── pages/
│   └── tenants.js      # Tenant management pages
└── components/
    └── (future components)
```

## 🎨 **Framework: Baidu amis**

This admin interface is built using [Baidu amis](https://baidu.github.io/amis/), a low-code frontend framework that allows us to build complex admin interfaces through JSON configuration.

### Key Features:
- **JSON Schema Driven**: UI is defined through JSON schemas
- **Rich Components**: Built-in CRUD tables, forms, charts, and more
- **API Integration**: Direct integration with REST APIs
- **Responsive Design**: Mobile-friendly interface
- **Theme Support**: Multiple theme options

## 🛠 **Available Features**

### ✅ **Currently Implemented**
- **Dashboard**: System overview with key metrics
- **Tenant Management**: 
  - List all tenants with search and pagination
  - Create new tenants with validation
  - View tenant details
  - Edit tenant information
  - Suspend/Reactivate tenants
  - Tenant analytics (placeholder)

### 🚧 **Planned Features**
- **User Management**: User CRUD and role management
- **Finance Module**: Account and transaction management
- **System Settings**: Feature flags and configuration
- **Audit Logs**: System activity monitoring
- **Analytics Dashboard**: Advanced reporting

## 🔧 **Configuration**

The admin UI configuration is centralized in `config/app-config.js`:

```javascript
window.APP_CONFIG = {
    api: {
        baseURL: 'http://localhost:8080',  // API server URL
        timeout: 30000
    },
    features: {
        tenantManagement: true,
        userManagement: true,
        financeModule: true,
        // ... other feature flags
    }
}
```

## 🔗 **API Integration**

The admin UI integrates directly with the ERP API endpoints:

- **Tenants**: `/api/v1/tenants/*`
- **Users**: `/api/v1/users/*` 
- **Finance**: `/api/v1/finance/*`
- **System**: `/api/v1/system/*`

### Authentication
Most API endpoints require tenant context via the `X-Tenant-ID` header. The admin UI handles this automatically.

## 🎯 **Usage Examples**

### Adding a New Page
1. Create a new file in `pages/` (e.g., `users.js`)
2. Define the page schema using amis JSON format
3. Add the page to the routing in `js/main.js`
4. Update the navigation menu in `config/app-config.js`

### Example Page Schema:
```javascript
window.UserPages = {
    list: {
        type: 'page',
        title: 'User Management',
        body: [
            {
                type: 'crud',
                api: '/api/v1/users',
                columns: [
                    { name: 'name', label: 'Name' },
                    { name: 'email', label: 'Email' },
                    { name: 'status', label: 'Status' }
                ]
            }
        ]
    }
};
```

## 🔒 **Security**

- Admin UI is served over the same HTTPS connection as the API
- All API requests include proper CORS headers
- Sensitive operations require confirmation dialogs
- No sensitive data is stored in browser localStorage

## 🚀 **Development**

The admin UI is automatically embedded into the Go binary during build. No separate build process is required.

To modify the admin UI:
1. Edit files in `web/admin/`
2. Rebuild the Go server: `go build -o server ./cmd/server`
3. Restart the server

## 📱 **Mobile Support**

The admin interface is responsive and works on mobile devices, though it's optimized for desktop use.

## 🎨 **Theming**

The interface uses the amis `cxd` theme by default. You can change themes by modifying the CSS imports in `index.html` and updating the theme configuration in `app-config.js`.

Available themes:
- `cxd` (default)
- `antd` 
- `dark`

## 🔍 **Troubleshooting**

### Admin UI not loading
- Check that the server is running on port 8080
- Verify the admin UI files are properly embedded
- Check browser console for JavaScript errors

### API calls failing
- Ensure the API server is accessible
- Check CORS configuration
- Verify X-Tenant-ID headers are being sent

### Development Issues
- Clear browser cache after rebuilding
- Check Go build output for embed errors
- Verify file paths are correct