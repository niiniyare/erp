// Search Analytics for AWO ERP Documentation
// Tracks search usage patterns and API documentation effectiveness

(function() {
    'use strict';
    
    // Analytics configuration
    const ANALYTICS_CONFIG = {
        enabled: true,
        endpoint: '/api/v1/analytics/search', // Optional: send to backend
        batchSize: 10,
        flushInterval: 30000, // 30 seconds
        trackingId: 'awo-erp-docs'
    };
    
    // Analytics data store
    let analyticsQueue = [];
    let sessionId = generateSessionId();
    let searchSession = null;
    
    function generateSessionId() {
        return 'search_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }
    
    function createSearchEvent(type, data) {
        return {
            type: type,
            sessionId: sessionId,
            timestamp: new Date().toISOString(),
            userAgent: navigator.userAgent,
            url: window.location.href,
            ...data
        };
    }
    
    // Track search patterns
    class SearchAnalytics {
        constructor() {
            this.currentSearch = null;
            this.searchStartTime = null;
            this.resultInteractions = [];
            this.popularQueries = new Map();
            this.apiSearchPatterns = new Map();
        }
        
        trackSearchStart(query) {
            this.currentSearch = query.trim().toLowerCase();
            this.searchStartTime = Date.now();
            this.resultInteractions = [];
            
            // Track popular queries
            const count = this.popularQueries.get(this.currentSearch) || 0;
            this.popularQueries.set(this.currentSearch, count + 1);
            
            // Track API-specific searches
            if (this.isAPISearch(query)) {
                const apiPattern = this.extractAPIPattern(query);
                if (apiPattern) {
                    const apiCount = this.apiSearchPatterns.get(apiPattern) || 0;
                    this.apiSearchPatterns.set(apiPattern, apiCount + 1);
                }
            }
            
            const event = createSearchEvent('search_start', {
                query: this.currentSearch,
                queryLength: query.length,
                isAPISearch: this.isAPISearch(query),
                searchType: this.categorizeSearch(query)
            });
            
            this.queueEvent(event);
        }
        
        trackSearchResults(query, resultCount, apiResultCount = 0) {
            if (!this.currentSearch || this.currentSearch !== query.trim().toLowerCase()) {
                return;
            }
            
            const event = createSearchEvent('search_results', {
                query: this.currentSearch,
                resultCount: resultCount,
                apiResultCount: apiResultCount,
                searchTime: Date.now() - this.searchStartTime,
                hasResults: resultCount > 0
            });
            
            this.queueEvent(event);
        }
        
        trackResultClick(query, resultUrl, resultTitle, position, isAPIResult = false) {
            if (!this.currentSearch) return;
            
            const interaction = {
                url: resultUrl,
                title: resultTitle,
                position: position,
                timestamp: Date.now(),
                isAPIResult: isAPIResult
            };
            
            this.resultInteractions.push(interaction);
            
            const event = createSearchEvent('result_click', {
                query: this.currentSearch,
                resultUrl: resultUrl,
                resultTitle: resultTitle,
                position: position,
                isAPIResult: isAPIResult,
                timeToClick: Date.now() - this.searchStartTime
            });
            
            this.queueEvent(event);
        }
        
        trackSearchComplete() {
            if (!this.currentSearch) return;
            
            const event = createSearchEvent('search_complete', {
                query: this.currentSearch,
                totalTime: Date.now() - this.searchStartTime,
                interactions: this.resultInteractions.length,
                clickedResults: this.resultInteractions.map(i => ({
                    url: i.url,
                    position: i.position,
                    isAPIResult: i.isAPIResult
                }))
            });
            
            this.queueEvent(event);
            this.currentSearch = null;
        }
        
        isAPISearch(query) {
            const apiTerms = [
                'api', 'endpoint', 'rest', 'post', 'get', 'put', 'delete', 'patch',
                'auth', 'login', 'token', 'jwt', 'abac', 'authorize', 'permission',
                'user', 'organization', 'tenant', 'entity', 'request', 'response',
                '/api/', 'curl', 'json', 'http', 'swagger'
            ];
            
            const lowerQuery = query.toLowerCase();
            return apiTerms.some(term => lowerQuery.includes(term));
        }
        
        extractAPIPattern(query) {
            const endpointPattern = /\/(api|abac)\/[^\s]*/i;
            const match = query.match(endpointPattern);
            if (match) return match[0];
            
            const httpMethodPattern = /(GET|POST|PUT|DELETE|PATCH)\s+/i;
            const methodMatch = query.match(httpMethodPattern);
            if (methodMatch) return methodMatch[1].toUpperCase();
            
            return null;
        }
        
        categorizeSearch(query) {
            const lowerQuery = query.toLowerCase();
            
            if (this.isAPISearch(query)) return 'api';
            if (lowerQuery.includes('tutorial') || lowerQuery.includes('guide') || lowerQuery.includes('how')) return 'tutorial';
            if (lowerQuery.includes('error') || lowerQuery.includes('troubleshoot') || lowerQuery.includes('fix')) return 'troubleshooting';
            if (lowerQuery.includes('config') || lowerQuery.includes('setup') || lowerQuery.includes('install')) return 'configuration';
            if (lowerQuery.includes('example') || lowerQuery.includes('code') || lowerQuery.includes('sample')) return 'examples';
            
            return 'general';
        }
        
        queueEvent(event) {
            analyticsQueue.push(event);
            
            if (analyticsQueue.length >= ANALYTICS_CONFIG.batchSize) {
                this.flushEvents();
            }
        }
        
        flushEvents() {
            if (analyticsQueue.length === 0) return;
            
            const events = analyticsQueue.slice();
            analyticsQueue = [];
            
            // Send to console for debugging
            console.group('📊 Search Analytics Batch');
            events.forEach(event => {
                console.log(`${event.type}:`, event);
            });
            console.groupEnd();
            
            // Send to Google Analytics if available
            if (typeof gtag !== 'undefined') {
                events.forEach(event => {
                    gtag('event', event.type, {
                        search_term: event.query,
                        result_count: event.resultCount,
                        search_type: event.searchType,
                        custom_parameter: JSON.stringify(event)
                    });
                });
            }
            
            // Send to backend analytics endpoint if configured
            if (ANALYTICS_CONFIG.endpoint && ANALYTICS_CONFIG.enabled) {
                fetch(ANALYTICS_CONFIG.endpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        events: events,
                        metadata: {
                            userAgent: navigator.userAgent,
                            timestamp: new Date().toISOString(),
                            sessionId: sessionId
                        }
                    })
                }).catch(err => {
                    console.warn('Failed to send search analytics:', err);
                });
            }
        }
        
        generateReport() {
            const report = {
                sessionId: sessionId,
                timestamp: new Date().toISOString(),
                popularQueries: Array.from(this.popularQueries.entries())
                    .sort(([,a], [,b]) => b - a)
                    .slice(0, 10),
                apiSearchPatterns: Array.from(this.apiSearchPatterns.entries())
                    .sort(([,a], [,b]) => b - a),
                totalSearches: Array.from(this.popularQueries.values()).reduce((a, b) => a + b, 0),
                apiSearchPercentage: this.calculateAPISearchPercentage(),
                queuedEvents: analyticsQueue.length
            };
            
            return report;
        }
        
        calculateAPISearchPercentage() {
            const totalSearches = Array.from(this.popularQueries.values()).reduce((a, b) => a + b, 0);
            const apiSearches = Array.from(this.apiSearchPatterns.values()).reduce((a, b) => a + b, 0);
            
            return totalSearches > 0 ? (apiSearches / totalSearches * 100).toFixed(1) : 0;
        }
    }
    
    // Initialize analytics
    const searchAnalytics = new SearchAnalytics();
    
    // Hook into search functionality
    document.addEventListener('DOMContentLoaded', function() {
        const searchInput = document.querySelector('[data-md-component="search-query"]');
        if (!searchInput) return;
        
        let searchTimeout;
        let lastQuery = '';
        
        // Track search input
        searchInput.addEventListener('input', function() {
            const query = this.value.trim();
            
            clearTimeout(searchTimeout);
            
            if (query.length >= 2 && query !== lastQuery) {
                searchTimeout = setTimeout(() => {
                    searchAnalytics.trackSearchStart(query);
                    lastQuery = query;
                }, 500);
            }
        });
        
        // Track result clicks
        document.addEventListener('click', function(e) {
            const link = e.target.closest('a');
            if (!link || !lastQuery) return;
            
            const resultContainer = link.closest('[data-md-component="search-result"]');
            if (!resultContainer) return;
            
            const isAPIResult = link.closest('.api-search-result') !== null;
            const position = Array.from(resultContainer.children).indexOf(link.closest('li, div')) + 1;
            
            searchAnalytics.trackResultClick(
                lastQuery,
                link.href,
                link.textContent.trim(),
                position,
                isAPIResult
            );
        });
        
        // Track search completion (when search is closed or cleared)
        document.addEventListener('keydown', function(e) {
            if (e.key === 'Escape' || (e.ctrlKey && e.key === 'k')) {
                if (lastQuery) {
                    searchAnalytics.trackSearchComplete();
                }
            }
        });
        
        // Track page visibility changes
        document.addEventListener('visibilitychange', function() {
            if (document.visibilityState === 'hidden') {
                searchAnalytics.flushEvents();
            }
        });
        
        // Flush events periodically
        setInterval(() => {
            searchAnalytics.flushEvents();
        }, ANALYTICS_CONFIG.flushInterval);
        
        // Flush events before page unload
        window.addEventListener('beforeunload', function() {
            searchAnalytics.flushEvents();
        });
        
        // Add analytics dashboard access (for debugging)
        if (window.location.search.includes('debug=analytics')) {
            const dashboardBtn = document.createElement('button');
            dashboardBtn.textContent = '📊 Analytics Report';
            dashboardBtn.style.cssText = `
                position: fixed;
                bottom: 20px;
                right: 20px;
                z-index: 1000;
                background: #007bff;
                color: white;
                border: none;
                border-radius: 4px;
                padding: 8px 12px;
                cursor: pointer;
                font-size: 12px;
            `;
            
            dashboardBtn.addEventListener('click', function() {
                const report = searchAnalytics.generateReport();
                console.group('📈 Search Analytics Report');
                console.log('Session ID:', report.sessionId);
                console.log('Total Searches:', report.totalSearches);
                console.log('API Search Percentage:', report.apiSearchPercentage + '%');
                console.table(report.popularQueries);
                console.log('API Patterns:', report.apiSearchPatterns);
                console.groupEnd();
                
                alert(`Search Analytics Report:\n\n` +
                      `Total Searches: ${report.totalSearches}\n` +
                      `API Searches: ${report.apiSearchPercentage}%\n` +
                      `Popular Queries: ${report.popularQueries.slice(0, 3).map(([q]) => q).join(', ')}\n\n` +
                      `Full report in console.`);
            });
            
            document.body.appendChild(dashboardBtn);
        }
    });
    
    // Expose analytics for external access
    window.AWOSearchAnalytics = {
        getReport: () => searchAnalytics.generateReport(),
        flush: () => searchAnalytics.flushEvents(),
        isEnabled: () => ANALYTICS_CONFIG.enabled
    };
    
})();