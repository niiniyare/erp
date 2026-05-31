// Contextual Help System for AWO ERP Documentation
// Provides smart contextual assistance and documentation shortcuts

document.addEventListener('DOMContentLoaded', function() {
    'use strict';
    
    // Help system configuration
    const HELP_CONFIG = {
        enabled: true,
        apiBaseUrl: 'http://localhost:8080',
        shortcuts: {
            'ctrl+?': 'toggleHelp',
            'ctrl+k': 'focusSearch', 
            'ctrl+shift+a': 'showAPIGuide',
            'alt+h': 'showContextualHelp'
        }
    };
    
    // Context detection patterns
    const CONTEXT_PATTERNS = {
        'api': {
            patterns: [/\/api\//, /swagger/, /endpoint/, /authentication/, /authorization/],
            helpType: 'api',
            resources: [
                { title: ' Interactive API Explorer', url: '/reference/api/swagger-ui/' },
                { title: 'Authentication Guide', url: '/reference/api/auth/' },
                { title: 'Complete Integration Tutorial', url: '/reference/api/tutorials/api-integration/' }
            ]
        },
        'abac': {
            patterns: [/abac/, /authorization/, /policy/, /access control/],
            helpType: 'abac',
            resources: [
                { title: 'ABAC Authorization API', url: '/reference/api/abac/' },
                { title: 'Policy Evaluation Guide', url: '/reference/modules/user/abac/policy_evaluation/' },
                { title: 'ABAC Design Overview', url: '/architecture/abac-design-and-roadmap/' }
            ]
        },
        'financial': {
            patterns: [/financial/, /accounting/, /double-entry/, /transaction/],
            helpType: 'financial',
            resources: [
                { title: 'Financial Module Overview', url: '/reference/modules/financial/architecture-guide/' },
                { title: 'Currency Management', url: '/reference/modules/financial/currency-management/' },
                { title: 'Security & Compliance', url: '/reference/modules/financial/security-compliance-guide/' }
            ]
        },
        'users': {
            patterns: [/user/, /profile/, /attributes/, /analytics/],
            helpType: 'users',
            resources: [
                { title: 'Users API Guide', url: '/reference/api/users/' },
                { title: 'Identity & RBAC', url: '/reference/modules/user/01_identity_model_and_rbac/' },
                { title: 'User Analytics', url: '/reference/api/users/#user-analytics' }
            ]
        },
        'organizations': {
            patterns: [/organization/, /department/, /hierarchy/, /structure/],
            helpType: 'organizations', 
            resources: [
                { title: 'Organizations API', url: '/reference/api/organizations/' },
                { title: 'Organizational Hierarchy', url: '/reference/api/organizations/#organization-types-and-structure' },
                { title: 'Organization Analytics', url: '/reference/api/organizations/#analytics-and-reporting' }
            ]
        },
        'development': {
            patterns: [/development/, /contributing/, /architecture/, /testing/],
            helpType: 'development',
            resources: [
                { title: 'Developer Quick Start', url: '/getting-started/01-developer-quick-start/' },
                { title: 'Best Practices', url: '/contributing/01-best-practices/' },
                { title: 'Architecture Overview', url: '/architecture/system-overview/' }
            ]
        }
    };
    
    // Smart context detection
    class ContextualHelper {
        constructor() {
            this.currentContext = null;
            this.helpVisible = false;
            this.setupKeyboardShortcuts();
            this.setupFloatingHelp();
            this.detectPageContext();
        }
        
        detectPageContext() {
            const currentPath = window.location.pathname;
            const pageTitle = document.title.toLowerCase();
            const pageContent = document.body.textContent.toLowerCase();
            
            // Detect context based on URL, title, and content
            for (const [contextName, config] of Object.entries(CONTEXT_PATTERNS)) {
                const matches = config.patterns.some(pattern => 
                    pattern.test(currentPath) || 
                    pattern.test(pageTitle) ||
                    pattern.test(pageContent)
                );
                
                if (matches) {
                    this.currentContext = contextName;
                    this.updateContextualHelp(config);
                    break;
                }
            }
        }
        
        updateContextualHelp(contextConfig) {
            // Add contextual help hints to the page
            this.addContextualHints(contextConfig);
            
            // Update floating help content
            this.updateFloatingHelp(contextConfig);
        }
        
        addContextualHints(contextConfig) {
            // Add contextual hints to relevant sections
            const codeBlocks = document.querySelectorAll('pre code');
            codeBlocks.forEach(block => {
                if (block.textContent.includes('curl') || block.textContent.includes('POST') || block.textContent.includes('/api/')) {
                    this.addAPIHint(block);
                }
            });
            
            // Add quick access buttons for current context
            this.addQuickAccessButtons(contextConfig);
        }
        
        addAPIHint(codeBlock) {
            const hint = document.createElement('div');
            hint.className = 'contextual-hint api-hint';
            hint.innerHTML = `
                <!-- <div style="background: #e3f2fd; border: 1px solid #2196f3; border-radius: 4px; padding: 8px; margin: 8px 0; font-size: 12px;"> -->
                <!--      <strong>Quick Tip:</strong> Test this endpoint in the  -->
                <!--     <a href="/reference/api/swagger-ui/" style="color: #1976d2;">Interactive API Explorer</a> -->
                <!-- </div> -->
            `;
            
            codeBlock.parentElement.insertBefore(hint, codeBlock.nextSibling);
        }
        
        addQuickAccessButtons(contextConfig) {
            // Add floating quick access menu
            const quickAccess = document.createElement('div');
            quickAccess.id = 'contextual-quick-access';
            quickAccess.style.cssText = `
                position: fixed;
                bottom: 20px;
                right: 80px;
                z-index: 1000;
                background: white;
                border: 1px solid #e0e0e0;
                border-radius: 8px;
                box-shadow: 0 4px 12px rgba(0,0,0,0.15);
                padding: 12px;
                max-width: 250px;
                transform: translateX(100%);
                transition: transform 0.3s ease;
            `;
            
            quickAccess.innerHTML = `
                <div style="font-weight: bold; margin-bottom: 8px; color: #1976d2;">
                     Related Resources
                </div>
                ${contextConfig.resources.map(resource => `
                    <div style="margin: 4px 0;">
                        <a href="${resource.url}" style="color: #1976d2; text-decoration: none; font-size: 12px;">
                            ${resource.title}
                        </a>
                    </div>
                `).join('')}
                <div style="margin-top: 8px; padding-top: 8px; border-top: 1px solid #e0e0e0; font-size: 11px; color: #666;">
                    Press <kbd>Ctrl+?</kbd> for more help
                </div>
            `;
            
            document.body.appendChild(quickAccess);
            
            // Show after a delay
            setTimeout(() => {
                quickAccess.style.transform = 'translateX(0)';
            }, 1000);
            
            // Auto-hide after some time
            setTimeout(() => {
                if (quickAccess.parentElement) {
                    quickAccess.style.transform = 'translateX(100%)';
                    setTimeout(() => {
                        if (quickAccess.parentElement) {
                            quickAccess.remove();
                        }
                    }, 300);
                }
            }, 10000);
        }
        
        setupFloatingHelp() {
            const helpButton = document.createElement('button');
            helpButton.id = 'contextual-help-button';
            helpButton.innerHTML = '?';
            helpButton.style.cssText = `
                position: fixed;
                bottom: 20px;
                right: 20px;
                width: 40px;
                height: 40px;
                border-radius: 50%;
                background: #2196f3;
                color: white;
                border: none;
                font-size: 18px;
                font-weight: bold;
                cursor: pointer;
                z-index: 1001;
                box-shadow: 0 2px 8px rgba(33,150,243,0.3);
                transition: all 0.2s ease;
            `;
            
            helpButton.addEventListener('mouseenter', () => {
                helpButton.style.transform = 'scale(1.1)';
                helpButton.style.boxShadow = '0 4px 12px rgba(33,150,243,0.5)';
            });
            
            helpButton.addEventListener('mouseleave', () => {
                helpButton.style.transform = 'scale(1)';
                helpButton.style.boxShadow = '0 2px 8px rgba(33,150,243,0.3)';
            });
            
            helpButton.addEventListener('click', () => {
                this.toggleHelpModal();
            });
            
            document.body.appendChild(helpButton);
        }
        
        updateFloatingHelp(contextConfig) {
            // Update the help button with context-specific information
            const helpButton = document.getElementById('contextual-help-button');
            if (helpButton) {
                helpButton.title = `Get help with ${contextConfig.helpType} (Ctrl+?)`;
            }
        }
        
        setupKeyboardShortcuts() {
            document.addEventListener('keydown', (e) => {
                // Ctrl+? for help
                if (e.ctrlKey && e.key === '/') {
                    e.preventDefault();
                    this.toggleHelpModal();
                }
                
                // Ctrl+K for search (handled by search-enhancements.js but we can add context)
                if (e.ctrlKey && e.key === 'k') {
                    // Add context-specific search hints
                    setTimeout(() => {
                        const searchInput = document.querySelector('[data-md-component="search-query"]');
                        if (searchInput && this.currentContext) {
                            searchInput.placeholder = `Search docs or try "${this.getContextualSearchHint()}"...`;
                        }
                    }, 100);
                }
                
                // Ctrl+Shift+A for API guide
                if (e.ctrlKey && e.shiftKey && e.key === 'A') {
                    e.preventDefault();
                    window.location.href = '/reference/api/swagger-ui/';
                }
            });
        }
        
        getContextualSearchHint() {
            const hints = {
                'api': 'POST /auth/login',
                'abac': 'ABAC policy',
                'financial': 'double entry',
                'users': 'user attributes',
                'organizations': 'org hierarchy',
                'development': 'best practices'
            };
            
            return hints[this.currentContext] || 'API endpoints';
        }
        
        toggleHelpModal() {
            if (this.helpVisible) {
                this.hideHelpModal();
            } else {
                this.showHelpModal();
            }
        }
        
        showHelpModal() {
            this.helpVisible = true;
            
            const modal = document.createElement('div');
            modal.id = 'contextual-help-modal';
            modal.style.cssText = `
                position: fixed;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                background: rgba(0,0,0,0.5);
                z-index: 2000;
                display: flex;
                align-items: center;
                justify-content: center;
            `;
            
            const modalContent = document.createElement('div');
            modalContent.style.cssText = `
                background: white;
                border-radius: 8px;
                padding: 24px;
                max-width: 600px;
                max-height: 80vh;
                overflow-y: auto;
                position: relative;
            `;
            
            modalContent.innerHTML = this.generateHelpContent();
            modal.appendChild(modalContent);
            
            // Close on overlay click
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.hideHelpModal();
                }
            });
            
            // Close on Escape
            document.addEventListener('keydown', (e) => {
                if (e.key === 'Escape' && this.helpVisible) {
                    this.hideHelpModal();
                }
            });
            
            document.body.appendChild(modal);
        }
        
        hideHelpModal() {
            this.helpVisible = false;
            const modal = document.getElementById('contextual-help-modal');
            if (modal) {
                modal.remove();
            }
        }
        
        generateHelpContent() {
            const contextConfig = CONTEXT_PATTERNS[this.currentContext];
            const contextName = this.currentContext || 'general';
            
            return `
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                    <h2 style="margin: 0; color: #1976d2;"> Documentation Help</h2>
                    <button onclick="document.getElementById('contextual-help-modal').remove(); window.contextualHelper.helpVisible = false;" 
                            style="background: none; border: none; font-size: 24px; cursor: pointer;">×</button>
                </div>
                
                <div style="margin-bottom: 16px;">
                    <strong>Current Context:</strong> <span style="color: #2196f3;">${contextName.toUpperCase()}</span>
                </div>
                
                ${contextConfig ? `
                    <div style="margin-bottom: 20px;">
                        <h3 style="color: #1976d2;"> Relevant Resources</h3>
                        ${contextConfig.resources.map(resource => `
                            <div style="margin: 8px 0;">
                                <a href="${resource.url}" style="color: #1976d2; text-decoration: none;">
                                    ${resource.title} →
                                </a>
                            </div>
                        `).join('')}
                    </div>
                ` : ''}
                
                <div style="margin-bottom: 20px;">
                    <h3 style="color: #1976d2;">⌨️ Keyboard Shortcuts</h3>
                    <div style="display: grid; grid-template-columns: auto 1fr; gap: 8px; font-size: 14px;">
                        <kbd style="padding: 2px 6px; background: #f5f5f5; border-radius: 3px;">Ctrl + K</kbd>
                        <span>Focus search</span>
                        <kbd style="padding: 2px 6px; background: #f5f5f5; border-radius: 3px;">Ctrl + ?</kbd>
                        <span>Toggle this help</span>
                        <kbd style="padding: 2px 6px; background: #f5f5f5; border-radius: 3px;">Ctrl + Shift + A</kbd>
                        <span>Open API Explorer</span>
                        <kbd style="padding: 2px 6px; background: #f5f5f5; border-radius: 3px;">Esc</kbd>
                        <span>Close modals/clear search</span>
                    </div>
                </div>
                
                <div style="margin-bottom: 16px;">
                    <h3 style="color: #1976d2;"> Quick Actions</h3>
                    <div style="display: flex; gap: 8px; flex-wrap: wrap;">
                        <a href="/reference/api/swagger-ui/" 
                           style="background: #2196f3; color: white; padding: 8px 12px; border-radius: 4px; text-decoration: none; font-size: 12px;">
                            API Docs
                        </a>
                        <a href="/getting-started/01-developer-quick-start/" 
                           style="background: #4caf50; color: white; padding: 8px 12px; border-radius: 4px; text-decoration: none; font-size: 12px;">
                            Quick Start
                        </a>
                        <a href="/contributing/01-best-practices/" 
                           style="background: #ff9800; color: white; padding: 8px 12px; border-radius: 4px; text-decoration: none; font-size: 12px;">
                            Best Practices
                        </a>
                    </div>
                </div>
                
                <div style="padding-top: 16px; border-top: 1px solid #e0e0e0; font-size: 12px; color: #666;">
                     <strong>Tip:</strong> The help system adapts based on the current page content. 
                    Navigate to different sections to see context-specific assistance.
                </div>
            `;
        }
    }
    
    // Initialize contextual help system
    window.contextualHelper = new ContextualHelper();
    
    // Add custom styles
    const helpStyles = document.createElement('style');
    helpStyles.textContent = `
        .contextual-hint {
            animation: fadeInUp 0.3s ease-out;
        }
        
        @keyframes fadeInUp {
            from {
                opacity: 0;
                transform: translateY(10px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        
        #contextual-help-button:active {
            transform: scale(0.95) !important;
        }
        
        #contextual-help-modal {
            animation: fadeIn 0.2s ease-out;
        }
        
        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }
        
        kbd {
            font-family: ui-monospace, SFMono-Regular, Consolas, 'Liberation Mono', Menlo, monospace;
        }
        
        @media (max-width: 768px) {
            #contextual-quick-access {
                right: 10px;
                bottom: 80px;
                max-width: 200px;
            }
            
            #contextual-help-button {
                right: 10px;
                bottom: 10px;
            }
        }
    `;
    
    document.head.appendChild(helpStyles);
    
    // Track help system usage
    function trackHelpUsage(action, context) {
        if (typeof gtag !== 'undefined') {
            gtag('event', 'contextual_help', {
                'help_action': action,
                'help_context': context || 'general'
            });
        }
        
        console.log('Contextual Help:', action, 'in', context || 'general', 'context');
    }
    
    // Expose tracking function
    window.trackHelpUsage = trackHelpUsage;
});
