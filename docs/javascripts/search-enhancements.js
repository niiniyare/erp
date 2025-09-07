// Search Enhancements for AWO ERP Documentation
// Enhances the default MkDocs search with API-specific features

document.addEventListener('DOMContentLoaded', function() {
    // API endpoint search enhancement
    const searchInput = document.querySelector('[data-md-component="search-query"]');
    const searchResults = document.querySelector('[data-md-component="search-result"]');
    
    if (!searchInput || !searchResults) return;
    
    // API endpoint patterns and metadata
    const apiEndpoints = [
        { path: '/abac/authorize', method: 'POST', category: 'Authorization', description: 'Simple authorization check' },
        { path: '/abac/evaluate', method: 'POST', category: 'Authorization', description: ' policy evaluation' },
        { path: '/abac/evaluate-bulk', method: 'POST', category: 'Authorization', description: 'Bulk authorization requests' },
        { path: '/api/v1/auth/login', method: 'POST', category: 'Authentication', description: 'User login and JWT token generation' },
        { path: '/api/v1/auth/refresh', method: 'POST', category: 'Authentication', description: 'Refresh JWT access token' },
        { path: '/api/v1/users', method: 'GET', category: 'Users', description: 'List users with pagination' },
        { path: '/api/v1/users', method: 'POST', category: 'Users', description: 'Create new user account' },
        { path: '/api/v1/users/{id}', method: 'GET', category: 'Users', description: 'Get user details by ID' },
        { path: '/api/v1/users/{id}', method: 'PUT', category: 'Users', description: 'Update user information' },
        { path: '/api/v1/users/profile', method: 'GET', category: 'Users', description: 'Get current user profile' },
        { path: '/api/v1/organizations', method: 'GET', category: 'Organizations', description: 'List organizations' },
        { path: '/api/v1/organizations', method: 'POST', category: 'Organizations', description: 'Create organization' },
        { path: '/api/v1/tenants', method: 'GET', category: 'Tenants', description: 'List tenants' },
        { path: '/api/v1/tenants', method: 'POST', category: 'Tenants', description: 'Create tenant' },
        { path: '/api/v1/access-requests', method: 'GET', category: 'Access Requests', description: 'List access requests' },
        { path: '/api/v1/access-requests', method: 'POST', category: 'Access Requests', description: 'Submit access request' },
        { path: '/api/v1/feature-flags', method: 'GET', category: 'Feature Flags', description: 'List feature flags' },
        { path: '/api/v1/analytics/users/{id}/behavior', method: 'GET', category: 'Analytics', description: 'User behavior analysis' }
    ];
    
    // Create API search suggestions
    function createAPISuggestions(query) {
        const suggestions = [];
        const lowerQuery = query.toLowerCase();
        
        // Search for matching endpoints
        apiEndpoints.forEach(endpoint => {
            const pathMatch = endpoint.path.toLowerCase().includes(lowerQuery);
            const categoryMatch = endpoint.category.toLowerCase().includes(lowerQuery);
            const descriptionMatch = endpoint.description.toLowerCase().includes(lowerQuery);
            const methodMatch = endpoint.method.toLowerCase().includes(lowerQuery);
            
            if (pathMatch || categoryMatch || descriptionMatch || methodMatch) {
                suggestions.push({
                    type: 'api-endpoint',
                    title: `${endpoint.method} ${endpoint.path}`,
                    description: endpoint.description,
                    category: endpoint.category,
                    url: '/reference/api/swagger-ui/#' + endpoint.path.replace(/[{}]/g, '').replace(/\//g, '_'),
                    score: pathMatch ? 100 : (categoryMatch ? 80 : (methodMatch ? 70 : 50))
                });
            }
        });
        
        return suggestions.sort((a, b) => b.score - a.score);
    }
    
    // Enhanced search result rendering
    function renderAPIResult(suggestion) {
        const methodColor = {
            'GET': '#28a745',
            'POST': '#007bff', 
            'PUT': '#ffc107',
            'DELETE': '#dc3545',
            'PATCH': '#6f42c1'
        };
        
        return `
            <div class="api-search-result" style="border-left: 4px solid ${methodColor[suggestion.title.split(' ')[0]] || '#6c757d'}; padding: 12px; margin: 8px 0; background: #f8f9fa; border-radius: 4px;">
                <div style="display: flex; align-items: center; gap: 8px;">
                    <span style="background: ${methodColor[suggestion.title.split(' ')[0]] || '#6c757d'}; color: white; padding: 2px 6px; border-radius: 3px; font-size: 11px; font-weight: bold;">
                        ${suggestion.title.split(' ')[0]}
                    </span>
                    <code style="background: #e9ecef; padding: 2px 6px; border-radius: 3px; font-size: 13px;">
                        ${suggestion.title.split(' ').slice(1).join(' ')}
                    </code>
                    <span style="background: #007bff; color: white; padding: 2px 6px; border-radius: 3px; font-size: 10px;">
                        ${suggestion.category}
                    </span>
                </div>
                <div style="margin-top: 6px; color: #6c757d; font-size: 13px;">
                    ${suggestion.description}
                </div>
                <div style="margin-top: 6px;">
                    <a href="/reference/api/swagger-ui/" style="color: #007bff; text-decoration: none; font-size: 12px;">
                        📄 View API Documentation →
                    </a>
                </div>
            </div>
        `;
    }
    
    // Search enhancement functionality
    let debounceTimer;
    let originalSearch = null;
    
    function enhanceSearch() {
        clearTimeout(debounceTimer);
        
        debounceTimer = setTimeout(() => {
            const query = searchInput.value.trim();
            
            if (query.length < 2) return;
            
            // Get API suggestions
            const apiSuggestions = createAPISuggestions(query);
            
            if (apiSuggestions.length > 0) {
                // Create API results section
                const apiResultsHTML = `
                    <div class="api-search-section">
                        <h3 style="color: #007bff; border-bottom: 2px solid #007bff; padding-bottom: 4px; margin: 16px 0 12px 0;">
                            🚀 API Endpoints (${apiSuggestions.length})
                        </h3>
                        ${apiSuggestions.slice(0, 5).map(renderAPIResult).join('')}
                        ${apiSuggestions.length > 5 ? `
                            <div style="text-align: center; margin: 12px 0;">
                                <a href="/reference/api/swagger-ui/" style="color: #007bff;">
                                    View all ${apiSuggestions.length} API endpoints in Documentation →
                                </a>
                            </div>
                        ` : ''}
                    </div>
                `;
                
                // Inject API results at the top
                const existingResults = searchResults.innerHTML;
                searchResults.innerHTML = apiResultsHTML + existingResults;
            }
        }, 300);
    }
    
    // Hook into search input
    searchInput.addEventListener('input', enhanceSearch);
    
    // Add keyboard shortcuts for API search
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + K to focus search with API hint
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            searchInput.focus();
            searchInput.placeholder = 'Search docs or try: "POST /auth/login", "ABAC", "user management"...';
            
            setTimeout(() => {
                searchInput.placeholder = 'Search';
            }, 3000);
        }
        
        // ESC to clear search
        if (e.key === 'Escape' && document.activeElement === searchInput) {
            searchInput.value = '';
            searchResults.innerHTML = '';
        }
    });
    
    // Add quick API category filters (mobile-friendly)
    const searchContainer = searchInput.closest('[data-md-component="search"]');
    if (searchContainer) {
        const quickFilters = document.createElement('div');
        quickFilters.className = 'api-quick-filters';
        quickFilters.style.cssText = `
            display: none;
            gap: 6px;
            margin: 8px 0;
            flex-wrap: wrap;
            max-height: 60px;
            overflow-x: auto;
            overflow-y: hidden;
            scrollbar-width: thin;
        `;
        
        // Show filters only when search is focused or has content
        const toggleFilters = (show) => {
            quickFilters.style.display = show ? 'flex' : 'none';
        };
        
        searchInput.addEventListener('focus', () => toggleFilters(true));
        searchInput.addEventListener('blur', () => {
            // Delay hiding to allow button clicks
            setTimeout(() => {
                if (!searchInput.value.trim()) {
                    toggleFilters(false);
                }
            }, 200);
        });
        searchInput.addEventListener('input', () => {
            toggleFilters(searchInput.value.length >= 1);
        });
        
        const categories = [...new Set(apiEndpoints.map(ep => ep.category))];
        categories.forEach((category, index) => {
            const filterBtn = document.createElement('button');
            filterBtn.textContent = category;
            filterBtn.className = 'api-filter-btn';
            filterBtn.style.cssText = `
                background: #f8f9fa;
                border: 1px solid #dee2e6;
                border-radius: 16px;
                padding: 6px 12px;
                font-size: 11px;
                cursor: pointer;
                transition: all 0.2s;
                white-space: nowrap;
                flex-shrink: 0;
                color: #495057;
                min-width: max-content;
                touch-action: manipulation;
            `;
            
            // Add touch-friendly interactions
            filterBtn.addEventListener('touchstart', (e) => {
                e.preventDefault();
                filterBtn.style.background = '#007bff';
                filterBtn.style.color = 'white';
                filterBtn.style.transform = 'scale(0.95)';
            });
            
            filterBtn.addEventListener('touchend', (e) => {
                e.preventDefault();
                setTimeout(() => {
                    filterBtn.style.background = '#f8f9fa';
                    filterBtn.style.color = '#495057';
                    filterBtn.style.transform = 'scale(1)';
                }, 150);
                
                searchInput.value = category.toLowerCase();
                enhanceSearch();
                searchInput.focus();
            });
            
            // Desktop interactions
            filterBtn.addEventListener('click', (e) => {
                e.preventDefault();
                searchInput.value = category.toLowerCase();
                enhanceSearch();
                searchInput.focus();
            });
            
            filterBtn.addEventListener('mouseenter', () => {
                if (!('ontouchstart' in window)) {
                    filterBtn.style.background = '#007bff';
                    filterBtn.style.color = 'white';
                    filterBtn.style.transform = 'translateY(-1px)';
                }
            });
            
            filterBtn.addEventListener('mouseleave', () => {
                if (!('ontouchstart' in window)) {
                    filterBtn.style.background = '#f8f9fa';
                    filterBtn.style.color = '#495057';
                    filterBtn.style.transform = 'translateY(0)';
                }
            });
            
            quickFilters.appendChild(filterBtn);
        });
        
        // Add "All APIs" button as first option
        const allBtn = document.createElement('button');
        allBtn.textContent = '🚀 All APIs';
        allBtn.className = 'api-filter-btn';
        allBtn.style.cssText = `
            background: #e3f2fd;
            border: 1px solid #2196f3;
            border-radius: 16px;
            padding: 6px 12px;
            font-size: 11px;
            cursor: pointer;
            transition: all 0.2s;
            white-space: nowrap;
            flex-shrink: 0;
            color: #1976d2;
            font-weight: 600;
            min-width: max-content;
            touch-action: manipulation;
        `;
        
        allBtn.addEventListener('click', (e) => {
            e.preventDefault();
            window.location.href = '/reference/api/swagger-ui/';
        });
        
        quickFilters.insertBefore(allBtn, quickFilters.firstChild);
        searchContainer.appendChild(quickFilters);
    }
    
    // Analytics tracking for API searches
    function trackAPISearch(query, resultCount) {
        // Track API documentation usage
        if (typeof gtag !== 'undefined') {
            gtag('event', 'api_search', {
                'search_term': query,
                'result_count': resultCount,
                'search_type': 'api_endpoint'
            });
        }
        
        // Console log for debugging
        console.log('API Search:', {
            query: query,
            resultCount: resultCount,
            timestamp: new Date().toISOString()
        });
    }
    
    // Track search usage
    let lastSearchQuery = '';
    searchInput.addEventListener('input', function() {
        const query = this.value.trim();
        if (query !== lastSearchQuery && query.length >= 2) {
            const apiResults = createAPISuggestions(query);
            trackAPISearch(query, apiResults.length);
            lastSearchQuery = query;
        }
    });
});

// Add styles for API search enhancements
const apiSearchStyles = document.createElement('style');
apiSearchStyles.textContent = `
    .api-search-result {
        transition: transform 0.2s ease, box-shadow 0.2s ease;
    }
    
    .api-search-result:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0,123,255,0.15);
    }
    
    .api-quick-filters button:hover {
        transform: translateY(-1px);
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    
    @media (max-width: 768px) {
        .api-search-result {
            margin: 6px 0;
            padding: 8px;
        }
        
        .api-quick-filters {
            justify-content: flex-start;
            overflow-x: auto;
            overflow-y: hidden;
            -webkit-overflow-scrolling: touch;
            scrollbar-width: thin;
            scrollbar-color: #ddd transparent;
            padding: 4px 0;
        }
        
        .api-quick-filters::-webkit-scrollbar {
            height: 4px;
        }
        
        .api-quick-filters::-webkit-scrollbar-track {
            background: transparent;
        }
        
        .api-quick-filters::-webkit-scrollbar-thumb {
            background: #ddd;
            border-radius: 2px;
        }
        
        .api-filter-btn {
            font-size: 10px !important;
            padding: 4px 8px !important;
            min-width: max-content !important;
            flex-shrink: 0 !important;
            margin-right: 4px;
        }
        
        /* Improve touch targets for mobile */
        .api-filter-btn {
            min-height: 32px;
            display: flex;
            align-items: center;
            justify-content: center;
        }
    }
    
    @media (max-width: 480px) {
        .api-quick-filters {
            gap: 4px;
            margin: 6px 0;
        }
        
        .api-filter-btn {
            font-size: 9px !important;
            padding: 4px 6px !important;
            min-height: 28px;
        }
    }
`;

document.head.appendChild(apiSearchStyles);