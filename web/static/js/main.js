// Awo ERP Main JavaScript

// Theme management
class ThemeManager {
    constructor() {
        this.currentTheme = localStorage.getItem('theme') || 'system';
        this.applyTheme();
        this.setupEventListeners();
    }

    applyTheme() {
        const root = document.documentElement;
        const body = document.body;
        
        if (this.currentTheme === 'system') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            const theme = prefersDark ? 'dark' : 'light';
            root.setAttribute('data-theme', theme);
            
            // Also update Tailwind classes
            if (theme === 'dark') {
                root.classList.add('dark');
                body.classList.remove('bg-gray-50');
                body.classList.add('bg-gray-900');
            } else {
                root.classList.remove('dark');
                body.classList.remove('bg-gray-900');
                body.classList.add('bg-gray-50');
            }
        } else {
            root.setAttribute('data-theme', this.currentTheme);
            
            // Also update Tailwind classes
            if (this.currentTheme === 'dark') {
                root.classList.add('dark');
                body.classList.remove('bg-gray-50');
                body.classList.add('bg-gray-900');
            } else {
                root.classList.remove('dark');
                body.classList.remove('bg-gray-900');
                body.classList.add('bg-gray-50');
            }
        }
    }

    setTheme(theme) {
        this.currentTheme = theme;
        localStorage.setItem('theme', theme);
        this.applyTheme();
    }

    setupEventListeners() {
        // Listen for system theme changes
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
            if (this.currentTheme === 'system') {
                this.applyTheme();
            }
        });
    }
}

// Component loader for dynamic imports
class ComponentLoader {
    constructor() {
        this.loadedComponents = new Set();
    }

    async loadComponent(name) {
        if (this.loadedComponents.has(name)) {
            return;
        }

        try {
            const script = document.createElement('script');
            script.src = `/static/js/${name}.min.js`;
            script.async = true;
            script.onload = () => {
                this.loadedComponents.add(name);
                console.log(`Component ${name} loaded successfully`);
            };
            script.onerror = () => {
                console.error(`Failed to load component: ${name}`);
            };
            document.head.appendChild(script);
        } catch (error) {
            console.error(`Error loading component ${name}:`, error);
        }
    }

    // Auto-load components based on DOM elements
    autoLoadComponents() {
        const componentMap = {
            '[data-component="dropdown"]': 'dropdown',
            '[data-component="dialog"]': 'dialog',
            '[data-component="toast"]': 'toast',
            '[data-component="sidebar"]': 'sidebar',
            '[data-component="tabs"]': 'tabs',
            '[data-component="calendar"]': 'calendar',
            '[data-component="chart"]': 'chart',
            '[data-component="datepicker"]': 'datepicker',
            '[data-component="carousel"]': 'carousel'
        };

        Object.entries(componentMap).forEach(([selector, component]) => {
            if (document.querySelector(selector)) {
                this.loadComponent(component);
            }
        });
    }
}

// Simple notification system
class NotificationManager {
    constructor() {
        this.container = this.createContainer();
    }

    createContainer() {
        const container = document.createElement('div');
        container.id = 'notification-container';
        container.className = 'fixed top-4 right-4 z-50 space-y-2';
        document.body.appendChild(container);
        return container;
    }

    show(message, type = 'info', duration = 5000) {
        const notification = document.createElement('div');
        notification.className = `
            p-4 rounded-md shadow-lg transform transition-all duration-300 translate-x-full
            ${this.getTypeClasses(type)}
        `;
        notification.innerHTML = `
            <div class="flex items-center justify-between">
                <span>${message}</span>
                <button class="ml-4 text-xl leading-none" onclick="this.parentElement.parentElement.remove()">×</button>
            </div>
        `;

        this.container.appendChild(notification);

        // Animate in
        setTimeout(() => {
            notification.classList.remove('translate-x-full');
        }, 10);

        // Auto remove
        if (duration > 0) {
            setTimeout(() => {
                notification.classList.add('translate-x-full');
                setTimeout(() => notification.remove(), 300);
            }, duration);
        }
    }

    getTypeClasses(type) {
        const classes = {
            success: 'bg-green-500 text-white',
            error: 'bg-red-500 text-white',
            warning: 'bg-yellow-500 text-black',
            info: 'bg-blue-500 text-white'
        };
        return classes[type] || classes.info;
    }
}

// Form enhancement utilities
class FormEnhancer {
    constructor() {
        this.enhanceForms();
    }

    enhanceForms() {
        document.querySelectorAll('form').forEach(form => {
            this.addFormValidation(form);
            this.addSubmitHandler(form);
        });
    }

    addFormValidation(form) {
        const inputs = form.querySelectorAll('input, select, textarea');
        inputs.forEach(input => {
            input.addEventListener('blur', () => this.validateField(input));
            input.addEventListener('input', () => this.clearFieldError(input));
        });
    }

    validateField(field) {
        const isValid = field.checkValidity();
        const errorElement = field.parentElement.querySelector('.field-error');

        if (!isValid) {
            this.showFieldError(field, field.validationMessage);
        } else if (errorElement) {
            errorElement.remove();
        }

        return isValid;
    }

    showFieldError(field, message) {
        this.clearFieldError(field);
        
        const errorElement = document.createElement('div');
        errorElement.className = 'field-error text-red-500 text-sm mt-1';
        errorElement.textContent = message;
        
        field.parentElement.appendChild(errorElement);
        field.classList.add('border-red-500');
    }

    clearFieldError(field) {
        const errorElement = field.parentElement.querySelector('.field-error');
        if (errorElement) {
            errorElement.remove();
        }
        field.classList.remove('border-red-500');
    }

    addSubmitHandler(form) {
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            
            // Validate all fields
            const inputs = form.querySelectorAll('input, select, textarea');
            let isFormValid = true;
            
            inputs.forEach(input => {
                if (!this.validateField(input)) {
                    isFormValid = false;
                }
            });

            if (isFormValid) {
                console.log('Form is valid, submitting...');
                // Here you would typically submit the form via AJAX
                window.notifications.show('Form submitted successfully!', 'success');
            } else {
                window.notifications.show('Please fix the errors in the form', 'error');
            }
        });
    }
}

// Component showcase functionality
class ComponentShowcase {
    constructor() {
        this.initializeInteractions();
        this.setupDemoHandlers();
    }

    initializeInteractions() {
        // Add click handlers for buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('button:not([disabled])')) {
                this.handleButtonClick(e.target);
            }
        });

        // Add change handlers for form elements
        document.addEventListener('change', (e) => {
            if (e.target.matches('input, select, textarea')) {
                this.handleFormElementChange(e.target);
            }
        });

        // Add hover effects for interactive elements
        document.addEventListener('mouseover', (e) => {
            if (e.target.matches('.interactive')) {
                e.target.classList.add('shadow-lg', 'transform', '-translate-y-1');
            }
        });

        document.addEventListener('mouseout', (e) => {
            if (e.target.matches('.interactive')) {
                e.target.classList.remove('shadow-lg', 'transform', '-translate-y-1');
            }
        });
    }

    handleButtonClick(button) {
        const buttonText = button.textContent.trim();
        const variant = this.getButtonVariant(button);
        
        // Create ripple effect
        this.createRippleEffect(button);
        
        // Show notification based on button type
        const messages = {
            'Default': 'Default button clicked!',
            'Destructive': 'Destructive action performed!',
            'Outline': 'Outline button activated!',
            'Secondary': 'Secondary action triggered!',
            'Ghost': 'Ghost button pressed!',
            'Link': 'Link button followed!',
            'Action': 'Card action executed!',
            'Edit': 'Edit mode activated!',
            'Delete': 'Item deletion requested!',
            'Previous': 'Previous page requested!',
            'Next': 'Next page requested!',
            'Submit': 'Form submission attempted!'
        };

        const message = messages[buttonText] || `${buttonText} button clicked!`;
        const type = variant === 'destructive' ? 'warning' : 'info';
        
        window.notifications.show(message, type, 3000);
    }

    getButtonVariant(button) {
        if (button.classList.contains('bg-red-600')) return 'destructive';
        if (button.classList.contains('border')) return 'outline';
        if (button.classList.contains('bg-gray-100')) return 'secondary';
        if (button.classList.contains('hover:bg-gray-100')) return 'ghost';
        if (button.classList.contains('text-blue-600')) return 'link';
        return 'default';
    }

    createRippleEffect(element) {
        const ripple = document.createElement('span');
        const rect = element.getBoundingClientRect();
        const size = Math.max(rect.width, rect.height);
        const x = rect.width / 2 - size / 2;
        const y = rect.height / 2 - size / 2;
        
        ripple.style.cssText = `
            position: absolute;
            border-radius: 50%;
            transform: scale(0);
            animation: ripple 600ms linear;
            background-color: rgba(255, 255, 255, 0.6);
            width: ${size}px;
            height: ${size}px;
            left: ${x}px;
            top: ${y}px;
            pointer-events: none;
        `;
        
        // Add CSS for ripple animation if not exists
        if (!document.querySelector('#ripple-style')) {
            const style = document.createElement('style');
            style.id = 'ripple-style';
            style.textContent = `
                @keyframes ripple {
                    to {
                        transform: scale(4);
                        opacity: 0;
                    }
                }
            `;
            document.head.appendChild(style);
        }
        
        element.style.position = 'relative';
        element.style.overflow = 'hidden';
        element.appendChild(ripple);
        
        setTimeout(() => ripple.remove(), 600);
    }

    handleFormElementChange(element) {
        const type = element.type;
        const value = element.value;
        
        if (type === 'checkbox') {
            const status = element.checked ? 'checked' : 'unchecked';
            window.notifications.show(`Checkbox ${status}`, 'info', 2000);
        } else if (type === 'radio') {
            window.notifications.show(`Radio option selected: ${element.nextElementSibling.textContent}`, 'info', 2000);
        } else if (element.tagName === 'SELECT') {
            window.notifications.show(`Option selected: ${value}`, 'info', 2000);
        } else if (value.length > 0) {
            window.notifications.show(`Input updated: ${value.substring(0, 20)}${value.length > 20 ? '...' : ''}`, 'info', 2000);
        }
    }

    setupDemoHandlers() {
        // Setup progress bar demo
        this.setupProgressDemo();
        
        // Setup tab switching demo
        this.setupTabDemo();
        
        // Setup theme switcher
        this.setupThemeSwitcher();
        
        // Setup copy to clipboard functionality
        this.setupCopyButtons();
    }

    setupProgressDemo() {
        const progressBars = document.querySelectorAll('.progress-fill, [class*="bg-blue-600"][style*="width"]');
        
        progressBars.forEach((bar, index) => {
            setInterval(() => {
                const currentWidth = parseInt(bar.style.width) || 0;
                const newWidth = (currentWidth + 10) % 110;
                bar.style.width = `${newWidth}%`;
                
                // Update percentage text if exists
                const percentText = bar.closest('.space-y-6')?.querySelector('.flex span:last-child');
                if (percentText) {
                    percentText.textContent = `${Math.min(newWidth, 100)}%`;
                }
            }, 2000 + (index * 500));
        });
    }

    setupTabDemo() {
        document.querySelectorAll('[aria-label="Tabs"] button').forEach(tab => {
            tab.addEventListener('click', () => {
                // Remove active state from all tabs
                tab.parentElement.querySelectorAll('button').forEach(t => {
                    t.classList.remove('border-blue-500', 'text-blue-600');
                    t.classList.add('border-transparent', 'text-gray-500');
                });
                
                // Add active state to clicked tab
                tab.classList.add('border-blue-500', 'text-blue-600');
                tab.classList.remove('border-transparent', 'text-gray-500');
                
                window.notifications.show(`Switched to ${tab.textContent} tab`, 'info', 2000);
            });
        });
    }

    setupThemeSwitcher() {
        // Add theme switcher button to the page
        const themeSwitcher = document.createElement('button');
        themeSwitcher.innerHTML = `
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"/>
            </svg>
        `;
        themeSwitcher.className = 'fixed bottom-4 right-4 bg-blue-600 text-white p-3 rounded-full shadow-lg hover:bg-blue-700 transition-colors z-50';
        themeSwitcher.title = 'Toggle theme';
        
        themeSwitcher.addEventListener('click', () => {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            window.themeManager.setTheme(newTheme);
            window.notifications.show(`Switched to ${newTheme} theme`, 'info', 2000);
        });
        
        document.body.appendChild(themeSwitcher);
    }

    setupCopyButtons() {
        document.querySelectorAll('.component-demo').forEach(demo => {
            const copyButton = document.createElement('button');
            copyButton.innerHTML = `
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                </svg>
            `;
            copyButton.className = 'absolute top-2 right-2 p-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors';
            copyButton.title = 'Copy HTML';
            
            copyButton.addEventListener('click', () => {
                const html = demo.innerHTML;
                navigator.clipboard.writeText(html).then(() => {
                    window.notifications.show('HTML copied to clipboard!', 'success', 2000);
                }).catch(() => {
                    window.notifications.show('Failed to copy HTML', 'error', 2000);
                });
            });
            
            demo.appendChild(copyButton);
        });
    }
}

// Initialize everything when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Initialize core systems
    window.themeManager = new ThemeManager();
    window.componentLoader = new ComponentLoader();
    window.notifications = new NotificationManager();
    window.formEnhancer = new FormEnhancer();
    window.componentShowcase = new ComponentShowcase();

    // Auto-load components
    window.componentLoader.autoLoadComponents();

    // Add smooth scrolling to anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth'
                });
            }
        });
    });

    // Add fade-in animation to main content
    const mainContent = document.querySelector('main');
    if (mainContent) {
        mainContent.classList.add('fade-in');
    }

    // Add staggered animations to component sections
    document.querySelectorAll('.component-section').forEach((section, index) => {
        section.style.animationDelay = `${index * 100}ms`;
        section.classList.add('fade-in');
    });

    console.log('Awo ERP UI initialized successfully');
});

// Utility functions
window.AwERP = {
    // Theme utilities
    setTheme: (theme) => window.themeManager.setTheme(theme),
    
    // Notification utilities
    notify: (message, type, duration) => window.notifications.show(message, type, duration),
    
    // Component loading
    loadComponent: (name) => window.componentLoader.loadComponent(name),
    
    // Utility functions
    formatCurrency: (amount, currency = 'USD') => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency
        }).format(amount);
    },
    
    formatDate: (date, options = {}) => {
        return new Intl.DateTimeFormat('en-US', options).format(new Date(date));
    },
    
    debounce: (func, wait) => {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
};