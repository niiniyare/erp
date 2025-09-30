/**
 * Modal AlpineJS Hooks (
 
 * Extracted from internal/ui/components/layout/modal.templ
 * Provides reusable Modal functionality for all ERP services
 * 
 * @version 2.0.0
 */

/**
 * @typedef {Object} ModalConfig
 * @property {string} [backdrop='static'] - Backdrop behavior: 'static' or 'dismissible'
 * @property {boolean} [keyboard=true] - Enable ESC key to close
 * @property {boolean} [focus=true] - Auto-focus first element
 * @property {string} [size='md'] - Modal size: 'sm', 'md', 'lg', 'xl'
 * @property {string} [position='center'] - Modal position
 */

/**
 * @typedef {Object} ModalSubmitOptions
 * @property {string} [method='POST'] - HTTP method
 * @property {Object} [headers={}] - Additional headers
 * @property {Function} [onSuccess] - Success callback
 * @property {Function} [onError] - Error callback
 */

/**
 * Modal Store - Global state management for modals
 * 
 * Manages modal lifecycle, focus, and nested modals with proper cleanup
 * and accessibility support.
 * 
 * @example
 * Alpine.store('modal').open('my-modal', { backdrop: 'dismissible' });
 */
document.addEventListener('alpine:init', () => {
    Alpine.store('modal', {
        // State
        isOpen: false,
        currentModal: '',
        modalStack: [],
        
        // Configuration
        backdrop: 'static',
        keyboard: true,
        focus: true,
        
        // Internal state
        _scrollPosition: 0,
        _activeElement: null,
        _focusTrap: null,
        _keydownHandler: null,
        
        /**
         * Open a modal
         * @param {string} modalId - Modal identifier
         * @param {ModalConfig} [config={}] - Modal configuration
         */
        open(modalId, config = {}) {
            if (!modalId) {
                console.error('Modal ID is required');
                return;
            }
            
            // Check if modal exists
            const modalElement = document.getElementById(modalId);
            if (!modalElement) {
                console.error(`Modal not found: ${modalId}`);
                return;
            }
            
            // Save current modal to stack if one is open
            if (this.isOpen && this.currentModal) {
                this.modalStack.push({
                    id: this.currentModal,
                    isOpen: this.isOpen,
                    backdrop: this.backdrop,
                    keyboard: this.keyboard,
                    focus: this.focus
                });
            }
            
            // Store currently focused element
            this._activeElement = document.activeElement;
            
            // Update state
            this.currentModal = modalId;
            this.isOpen = true;
            this.backdrop = config.backdrop || 'static';
            this.keyboard = config.keyboard !== false;
            this.focus = config.focus !== false;
            
            // Prevent body scroll
            this._disableBodyScroll();
            
            // Focus management
            if (this.focus) {
                this._setInitialFocus(modalId);
            }
            
            // Setup keyboard handlers
            if (this.keyboard) {
                this._setupKeyboardHandlers(modalId);
            }
            
            // Setup focus trap
            this._setupFocusTrap(modalId);
            
            // Emit event
            this._dispatchEvent('modal:opened', { modalId, config });
        },
        
        /**
         * Close a modal
         * @param {string|null} [modalId=null] - Modal to close (null for current)
         * @returns {boolean} True if modal was closed
         */
        close(modalId = null) {
            if (modalId && modalId !== this.currentModal) {
                return false; // Not the current modal
            }
            
            const closingModalId = modalId || this.currentModal;
            
            // Cleanup
            this._cleanupKeyboardHandlers();
            this._cleanupFocusTrap();
            
            // Restore previous modal or close completely
            const previousModal = this.modalStack.pop();
            
            if (previousModal && previousModal.id) {
                // Restore previous modal state
                this.currentModal = previousModal.id;
                this.isOpen = previousModal.isOpen;
                this.backdrop = previousModal.backdrop;
                this.keyboard = previousModal.keyboard;
                this.focus = previousModal.focus;
                
                // Re-setup handlers for previous modal
                if (this.keyboard) {
                    this._setupKeyboardHandlers(this.currentModal);
                }
                if (this.focus) {
                    this._setupFocusTrap(this.currentModal);
                }
            } else {
                // No previous modal - fully close
                this.isOpen = false;
                this.currentModal = '';
                
                // Restore body scroll
                this._enableBodyScroll();
                
                // Restore focus
                if (this._activeElement && this._activeElement.focus) {
                    try {
                        this._activeElement.focus();
                    } catch (e) {
                        // Element might be removed from DOM
                    }
                }
                this._activeElement = null;
            }
            
            // Emit event
            this._dispatchEvent('modal:closed', { modalId: closingModalId });
            
            return true;
        },
        
        /**
         * Toggle modal open/close state
         * @param {string} modalId - Modal identifier
         * @param {ModalConfig} [config={}] - Modal configuration
         */
        toggle(modalId, config = {}) {
            if (this.isOpen && this.currentModal === modalId) {
                this.close(modalId);
            } else {
                this.open(modalId, config);
            }
        },
        
        /**
         * Close all modals
         */
        closeAll() {
            this._cleanupKeyboardHandlers();
            this._cleanupFocusTrap();
            
            this.modalStack = [];
            this.isOpen = false;
            this.currentModal = '';
            
            this._enableBodyScroll();
            
            if (this._activeElement && this._activeElement.focus) {
                try {
                    this._activeElement.focus();
                } catch (e) {
                    // Element might be removed from DOM
                }
            }
            this._activeElement = null;
            
            this._dispatchEvent('modal:all-closed', {});
        },
        
        /**
         * Check if a specific modal is currently open
         * @param {string} modalId - Modal identifier
         * @returns {boolean} True if modal is open
         */
        isModalOpen(modalId) {
            return this.isOpen && this.currentModal === modalId;
        },
        
        /**
         * Submit modal form
         * @param {string} modalId - Modal identifier
         * @param {FormData|Object} formData - Form data to submit
         * @param {string} endpoint - Submission endpoint
         * @param {ModalSubmitOptions} [options={}] - Submission options
         * @returns {Promise<Object>} Response object
         */
        async submit(modalId, formData, endpoint, options = {}) {
            if (!endpoint) {
                console.error('Endpoint is required for modal submission');
                return { success: false, error: 'No endpoint provided' };
            }
            
            try {
                const response = await this._makeRequest(endpoint, {
                    method: options.method || 'POST',
                    data: formData,
                    headers: options.headers || {}
                });
                
                if (response.success) {
                    // Call success callback
                    if (options.onSuccess && typeof options.onSuccess === 'function') {
                        options.onSuccess(response);
                    }
                    
                    // Close modal
                    this.close(modalId);
                    
                    // Show success message
                    if (response.message) {
                        this._showToast('success', response.message);
                    }
                    
                    // Emit success event
                    this._dispatchEvent('modal:submit-success', {
                        modalId,
                        response,
                        formData: this._formDataToObject(formData)
                    });
                    
                    return response;
                } else {
                    // Call error callback
                    if (options.onError && typeof options.onError === 'function') {
                        options.onError(response);
                    }
                    
                    // Handle validation errors
                    if (response.errors) {
                        this._showValidationErrors(response.errors);
                    }
                    
                    // Show error message
                    if (response.message) {
                        this._showToast('error', response.message);
                    }
                    
                    return response;
                }
                
            } catch (error) {
                console.error('Modal form submission failed:', error);
                
                const errorResponse = {
                    success: false,
                    error: error.message || 'Network error occurred'
                };
                
                // Call error callback
                if (options.onError && typeof options.onError === 'function') {
                    options.onError(errorResponse);
                }
                
                this._showToast('error', 'Form submission failed. Please try again.');
                
                return errorResponse;
            }
        },
        
        /**
         * Disable body scroll
         * @private
         */
        _disableBodyScroll() {
            // Save current scroll position
            this._scrollPosition = window.pageYOffset || document.documentElement.scrollTop;
            
            // Add modal-open class
            document.body.classList.add('modal-open');
            document.body.style.top = `-${this._scrollPosition}px`;
            document.body.style.position = 'fixed';
            document.body.style.width = '100%';
        },
        
        /**
         * Enable body scroll
         * @private
         */
        _enableBodyScroll() {
            document.body.classList.remove('modal-open');
            document.body.style.position = '';
            document.body.style.top = '';
            document.body.style.width = '';
            
            // Restore scroll position
            window.scrollTo(0, this._scrollPosition);
            this._scrollPosition = 0;
        },
        
        /**
         * Set initial focus on modal open
         * @param {string} modalId - Modal identifier
         * @private
         */
        _setInitialFocus(modalId) {
            requestAnimationFrame(() => {
                const modal = document.getElementById(modalId);
                if (!modal) return;
                
                // Try to find autofocus element first
                let focusElement = modal.querySelector('[autofocus]');
                
                // If no autofocus, find first focusable element
                if (!focusElement) {
                    const focusableSelector = [
                        'button:not([disabled])',
                        '[href]:not([disabled])',
                        'input:not([disabled]):not([type="hidden"])',
                        'select:not([disabled])',
                        'textarea:not([disabled])',
                        '[tabindex]:not([tabindex="-1"]):not([disabled])'
                    ].join(', ');
                    
                    const focusableElements = modal.querySelectorAll(focusableSelector);
                    focusElement = focusableElements[0];
                }
                
                if (focusElement && focusElement.focus) {
                    focusElement.focus();
                }
            });
        },
        
        /**
         * Setup keyboard event handlers
         * @param {string} modalId - Modal identifier
         * @private
         */
        _setupKeyboardHandlers(modalId) {
            this._cleanupKeyboardHandlers();
            
            this._keydownHandler = (e) => {
                if (e.key === 'Escape' && this.keyboard && this.currentModal === modalId) {
                    e.preventDefault();
                    this.close(modalId);
                }
            };
            
            document.addEventListener('keydown', this._keydownHandler);
        },
        
        /**
         * Cleanup keyboard event handlers
         * @private
         */
        _cleanupKeyboardHandlers() {
            if (this._keydownHandler) {
                document.removeEventListener('keydown', this._keydownHandler);
                this._keydownHandler = null;
            }
        },
        
        /**
         * Setup focus trap for accessibility
         * @param {string} modalId - Modal identifier
         * @private
         */
        _setupFocusTrap(modalId) {
            const modal = document.getElementById(modalId);
            if (!modal) return;
            
            this._focusTrap = (e) => {
                if (e.key !== 'Tab') return;
                
                const focusableSelector = [
                    'button:not([disabled])',
                    '[href]:not([disabled])',
                    'input:not([disabled]):not([type="hidden"])',
                    'select:not([disabled])',
                    'textarea:not([disabled])',
                    '[tabindex]:not([tabindex="-1"]):not([disabled])'
                ].join(', ');
                
                const focusableElements = Array.from(modal.querySelectorAll(focusableSelector));
                const firstElement = focusableElements[0];
                const lastElement = focusableElements[focusableElements.length - 1];
                
                if (e.shiftKey) {
                    // Shift + Tab
                    if (document.activeElement === firstElement) {
                        e.preventDefault();
                        lastElement.focus();
                    }
                } else {
                    // Tab
                    if (document.activeElement === lastElement) {
                        e.preventDefault();
                        firstElement.focus();
                    }
                }
            };
            
            modal.addEventListener('keydown', this._focusTrap);
        },
        
        /**
         * Cleanup focus trap
         * @private
         */
        _cleanupFocusTrap() {
            if (this._focusTrap && this.currentModal) {
                const modal = document.getElementById(this.currentModal);
                if (modal) {
                    modal.removeEventListener('keydown', this._focusTrap);
                }
                this._focusTrap = null;
            }
        },
        
        /**
         * Make HTTP request
         * @param {string} url - Request URL
         * @param {Object} options - Request options
         * @returns {Promise<Object>} Response data
         * @private
         */
        async _makeRequest(url, options) {
            let body = options.data;
            
            // Convert object to FormData if needed
            if (!(body instanceof FormData) && typeof body === 'object') {
                const formData = new FormData();
                
                Object.entries(body).forEach(([key, value]) => {
                    if (value === null || value === undefined) {
                        return;
                    }
                    
                    if (value instanceof FileList || value instanceof File) {
                        const files = value instanceof FileList ? Array.from(value) : [value];
                        files.forEach(file => formData.append(key, file));
                    } else if (Array.isArray(value)) {
                        value.forEach(v => formData.append(`${key}[]`, v));
                    } else if (typeof value === 'object' && !(value instanceof Date)) {
                        formData.append(key, JSON.stringify(value));
                    } else {
                        formData.append(key, value);
                    }
                });
                
                body = formData;
            }
            
            // Get CSRF token
            const csrfToken = document.querySelector('meta[name="csrf-token"]')?.content;
            
            const headers = {
                'X-Requested-With': 'XMLHttpRequest',
                ...(options.headers || {})
            };
            
            if (csrfToken) {
                headers['X-CSRF-Token'] = csrfToken;
            }
            
            const response = await fetch(url, {
                method: options.method || 'POST',
                body,
                headers
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }
            
            return { success: true, data: await response.text() };
        },
        
        /**
         * Show validation errors
         * @param {Object.<string, string|string[]>} errors - Validation errors
         * @private
         */
        _showValidationErrors(errors) {
            // Clear previous errors
            document.querySelectorAll('.validation-error, [data-validation-for]').forEach(el => {
                if (el.hasAttribute('data-validation-for')) {
                    el.textContent = '';
                    el.style.display = 'none';
                }
            });
            
            // Show new errors
            Object.entries(errors).forEach(([field, messages]) => {
                const errorElement = document.querySelector(`[data-validation-for="${field}"]`);
                if (errorElement) {
                    const message = Array.isArray(messages) ? messages[0] : messages;
                    errorElement.textContent = message;
                    errorElement.style.display = 'block';
                    
                    // Also add error class to input
                    const inputElement = document.querySelector(`[name="${field}"]`);
                    if (inputElement) {
                        inputElement.classList.add('error', 'is-invalid');
                    }
                }
            });
        },
        
        /**
         * Show toast message
         * @param {string} type - Toast type (success, error, info, warning)
         * @param {string} message - Message to display
         * @private
         */
        _showToast(type, message) {
            const toastStore = Alpine.store('toast');
            if (toastStore && typeof toastStore[type] === 'function') {
                toastStore[type](message);
            }
        },
        
        /**
         * Convert FormData to plain object
         * @param {FormData|Object} formData - Form data
         * @returns {Object} Plain object
         * @private
         */
        _formDataToObject(formData) {
            if (!(formData instanceof FormData)) {
                return formData;
            }
            
            const obj = {};
            for (const [key, value] of formData.entries()) {
                if (obj[key]) {
                    // Handle multiple values
                    if (Array.isArray(obj[key])) {
                        obj[key].push(value);
                    } else {
                        obj[key] = [obj[key], value];
                    }
                } else {
                    obj[key] = value;
                }
            }
            return obj;
        },
        
        /**
         * Dispatch custom event
         * @param {string} eventName - Event name
         * @param {Object} detail - Event detail
         * @private
         */
        _dispatchEvent(eventName, detail) {
            window.dispatchEvent(new CustomEvent(eventName, {
                detail,
                bubbles: true,
                cancelable: true
            }));
        }
    });
});

/**
 * @typedef {Object} UseModalConfig
 * @extends ModalConfig
 * @property {string} id - Modal identifier
 * @property {string} [size='md'] - Modal size
 * @property {string} [position='center'] - Modal position
 * @property {boolean} [closable=true] - Can be closed by user
 */

/**
 * Modal Component Hook
 * 
 * Provides instance-level modal management with form handling
 * and lifecycle hooks.
 * 
 * @param {UseModalConfig} [config={}] - Configuration options
 * @returns {Object} Alpine.js component data
 * 
 * @example
 * <div x-data="useModal({ id: 'my-modal', size: 'lg' })">
 */
function useModal(config = {}) {
    if (!config.id) {
        console.warn('Modal ID is recommended for useModal');
    }
    
    return {
        // Configuration
        id: config.id || '',
        size: config.size || 'md',
        position: config.position || 'center',
        backdrop: config.backdrop || 'static',
        closable: config.closable !== false,
        keyboard: config.keyboard !== false,
        
        // State
        localLoading: false,
        formData: {},
        validationErrors: {},
        
        // Internal state
        _eventListeners: [],
        _destroyed: false,
        
        /**
         * Initialize modal component
         */
        init() {
            if (!this.id) {
                console.error('Modal ID is required');
                return;
            }
            
            // Setup event listeners
            this._addEventListener('modal:opened', (e) => {
                if (e.detail.modalId === this.id) {
                    this.onOpened(e.detail);
                }
            });
            
            this._addEventListener('modal:closed', (e) => {
                if (e.detail.modalId === this.id) {
                    this.onClosed(e.detail);
                }
            });
            
            // Keyboard event handling for dismissible modals
            if (this.backdrop === 'dismissible' && this.keyboard) {
                const keyHandler = (e) => {
                    if (e.key === 'Escape' && this.isCurrentModal) {
                        this.close();
                    }
                };
                this.$el.addEventListener('keydown', keyHandler);
                this._eventListeners.push({ type: 'keydown', handler: keyHandler, element: this.$el });
            }
        },
        
        /**
         * Cleanup on component destroy
         */
        destroy() {
            this._destroyed = true;
            
            // Remove all event listeners
            this._eventListeners.forEach(({ type, handler, element }) => {
                element.removeEventListener(type, handler);
            });
            this._eventListeners = [];
            
            // Close modal if it's open
            if (this.isOpen) {
                this.close();
            }
        },
        
        /**
         * Add event listener with cleanup tracking
         * @param {string} type - Event type
         * @param {Function} handler - Event handler
         * @param {EventTarget} [target=window] - Event target
         * @private
         */
        _addEventListener(type, handler, target = window) {
            target.addEventListener(type, handler);
            this._eventListeners.push({ type, handler, element: target });
        },
        
        /**
         * Open this modal
         */
        open() {
            if (this._destroyed) {
                console.warn('Cannot open destroyed modal');
                return;
            }
            
            Alpine.store('modal').open(this.id, {
                backdrop: this.backdrop,
                keyboard: this.keyboard,
                size: this.size,
                position: this.position
            });
        },
        
        /**
         * Close this modal
         */
        close() {
            if (this.closable) {
                Alpine.store('modal').close(this.id);
            }
        },
        
        /**
         * Toggle this modal
         */
        toggle() {
            Alpine.store('modal').toggle(this.id, {
                backdrop: this.backdrop,
                keyboard: this.keyboard,
                size: this.size,
                position: this.position
            });
        },
        
        /**
         * Submit form within modal
         * @param {HTMLFormElement|Event} formElementOrEvent - Form element or submit event
         * @param {string} endpoint - Submission endpoint
         * @param {ModalSubmitOptions} [options={}] - Submission options
         * @returns {Promise<Object>} Response object
         */
        async submitForm(formElementOrEvent, endpoint, options = {}) {
            if (this.localLoading) {
                return { success: false, error: 'Form is already submitting' };
            }
            
            // Extract form element
            let formElement = formElementOrEvent;
            if (formElementOrEvent instanceof Event) {
                formElementOrEvent.preventDefault();
                formElement = formElementOrEvent.target;
            }
            
            if (!(formElement instanceof HTMLFormElement)) {
                console.error('Invalid form element provided');
                return { success: false, error: 'Invalid form element' };
            }
            
            this.localLoading = true;
            this.validationErrors = {};
            
            try {
                const formData = new FormData(formElement);
                
                // Add any additional data from this.formData
                if (this.formData && typeof this.formData === 'object') {
                    Object.entries(this.formData).forEach(([key, value]) => {
                        if (value !== null && value !== undefined) {
                            formData.append(key, value);
                        }
                    });
                }
                
                const response = await Alpine.store('modal').submit(
                    this.id,
                    formData,
                    endpoint,
                    {
                        ...options,
                        onSuccess: (res) => {
                            this.onSubmitSuccess(res);
                            if (options.onSuccess) options.onSuccess(res);
                        },
                        onError: (res) => {
                            this.onSubmitError(res);
                            if (options.onError) options.onError(res);
                        }
                    }
                );
                
                if (!response.success && response.errors) {
                    this.validationErrors = response.errors;
                }
                
                return response;
                
            } catch (error) {
                console.error('Form submission error:', error);
                return {
                    success: false,
                    error: error.message || 'Form submission failed'
                };
            } finally {
                this.localLoading = false;
            }
        },
        
        /**
         * Handle backdrop click
         * @param {MouseEvent} event - Click event
         */
        handleBackdropClick(event) {
            // Only close if clicking directly on backdrop (not bubbled from content)
            if (event.target === event.currentTarget && 
                this.backdrop === 'dismissible' && 
                this.closable) {
                this.close();
            }
        },
        
        /**
         * Reset modal state
         */
        reset() {
            this.formData = {};
            this.validationErrors = {};
            this.localLoading = false;
        },
        
        /**
         * Lifecycle hook - called when modal opens
         * @param {Object} detail - Event detail
         */
        onOpened(detail) {
            // Override in specific modals
        },
        
        /**
         * Lifecycle hook - called when modal closes
         * @param {Object} detail - Event detail
         */
        onClosed(detail) {
            this.reset();
        },
        
        /**
         * Lifecycle hook - called on successful form submission
         * @param {Object} response - Server response
         */
        onSubmitSuccess(response) {
            // Override in specific modals
        },
        
        /**
         * Lifecycle hook - called on form submission error
         * @param {Object} response - Error response
         */
        onSubmitError(response) {
            // Override in specific modals
        },
        
        /**
         * Get validation error for a field
         * @param {string} field - Field name
         * @returns {string|null} First error message or null
         */
        getFieldError(field) {
            const errors = this.validationErrors[field];
            if (!errors) return null;
            return Array.isArray(errors) ? errors[0] : errors;
        },
        
        /**
         * Check if field has error
         * @param {string} field - Field name
         * @returns {boolean} True if field has error
         */
        hasFieldError(field) {
            return !!this.validationErrors[field];
        },
        
        /**
         * Clear validation error for a field
         * @param {string} field - Field name
         */
        clearFieldError(field) {
            if (this.validationErrors[field]) {
                delete this.validationErrors[field];
            }
        },
        
        /**
         * Clear all validation errors
         */
        clearAllErrors() {
            this.validationErrors = {};
        },
        
        // Computed properties
        
        /**
         * Check if this is the current modal
         * @returns {boolean} True if current modal
         */
        get isCurrentModal() {
            return Alpine.store('modal').currentModal === this.id;
        },
        
        /**
         * Check if modal is open
         * @returns {boolean} True if open
         */
        get isOpen() {
            return this.isCurrentModal && Alpine.store('modal').isOpen;
        },
        
        /**
         * Check if form has validation errors
         * @returns {boolean} True if has errors
         */
        get hasErrors() {
            return Object.keys(this.validationErrors).length > 0;
        },
        
        /**
         * Check if form can be submitted
         * @returns {boolean} True if can submit
         */
        get canSubmit() {
            return !this.localLoading && !this.hasErrors;
        }
    };
}

/**
 * @typedef {Object} UseFormModalConfig
 * @extends UseModalConfig
 * @property {boolean} [saveOnClose=false] - Auto-save on close
 * @property {boolean} [confirmOnClose=true] - Confirm before closing with unsaved changes
 * @property {string} [saveEndpoint] - Endpoint for auto-save
 * @property {Function} [customConfirm] - Custom confirmation function
 */

/**
 * Form Modal Hook
 * 
 * Specialized modal for forms with dirty tracking and auto-save support
 * 
 * @param {UseFormModalConfig} [config={}] - Configuration options
 * @returns {Object} Alpine.js component data
 * 
 * @example
 * <div x-data="useFormModal({ 
 *   id: 'edit-modal', 
 *   confirmOnClose: true,
 *   saveOnClose: false
 * })">
 */
function useFormModal(config = {}) {
    const modal = useModal(config);
    
    // Extract init and destroy for proper override
    const { init: modalInit, destroy: modalDestroy, close: modalClose, ...modalRest } = modal;
    
    return {
        ...modalRest,
        
        // Form-specific configuration
        saveOnClose: config.saveOnClose || false,
        confirmOnClose: config.confirmOnClose !== false,
        saveEndpoint: config.saveEndpoint || '',
        customConfirm: config.customConfirm || null,
        
        // Form-specific state
        isDirty: false,
        initialFormData: {},
        
        /**
         * Initialize form modal
         */
        init() {
            if (modalInit) modalInit.call(this);
            
            // Save initial form data
            this.saveInitialFormData();
            
            // Watch for form changes
            this.$watch('formData', () => {
                this._checkDirty();
            }, { deep: true });
        },
        
        /**
         * Cleanup
         */
        destroy() {
            if (modalDestroy) modalDestroy.call(this);
        },
        
        /**
         * Save initial form data for dirty checking
         */
        saveInitialFormData() {
            this.initialFormData = JSON.parse(JSON.stringify(this.formData || {}));
        },
        
        /**
         * Check if form data has changed
         * @private
         */
        _checkDirty() {
            this.isDirty = JSON.stringify(this.formData) !== JSON.stringify(this.initialFormData);
        },
        
        /**
         * Mark form as dirty
         */
        markDirty() {
            this.isDirty = true;
        },
        
        /**
         * Mark form as clean (pristine)
         */
        markClean() {
            this.isDirty = false;
            this.saveInitialFormData();
        },
        
        /**
         * Close modal with dirty check
         * @returns {Promise<boolean>} True if closed successfully
         */
        async close() {
            if (this.isDirty && this.confirmOnClose) {
                const confirmed = await this.confirmClose();
                if (!confirmed) {
                    return false;
                }
            }
            
            if (this.isDirty && this.saveOnClose) {
                const saved = await this.autoSave();
                if (!saved) {
                    // Auto-save failed, ask user what to do
                    const proceedAnyway = await this._confirmCloseAfterSaveFailed();
                    if (!proceedAnyway) {
                        return false;
                    }
                }
            }
            
            modalClose.call(this);
            return true;
        },
        
        /**
         * Confirm closing modal with unsaved changes
         * @returns {Promise<boolean>} True if user confirms
         */
        async confirmClose() {
            if (this.customConfirm && typeof this.customConfirm === 'function') {
                return await this.customConfirm();
            }
            
            // Default browser confirm
            return window.confirm('You have unsaved changes. Are you sure you want to close?');
        },
        
        /**
         * Confirm closing after auto-save failed
         * @returns {Promise<boolean>} True if user wants to proceed
         * @private
         */
        async _confirmCloseAfterSaveFailed() {
            return window.confirm('Auto-save failed. Close without saving?');
        },
        
        /**
         * Auto-save form data
         * @returns {Promise<boolean>} True if saved successfully
         */
        async autoSave() {
            if (!this.saveEndpoint) {
                console.warn('No save endpoint configured for auto-save');
                return false;
            }
            
            const formElement = this.$el.querySelector('form');
            if (!formElement) {
                console.warn('No form element found for auto-save');
                return false;
            }
            
            try {
                const response = await this.submitForm(
                    formElement,
                    this.saveEndpoint,
                    {
                        method: 'POST',
                        onSuccess: () => {
                            this.markClean();
                        }
                    }
                );
                
                return response.success;
            } catch (error) {
                console.error('Auto-save failed:', error);
                return false;
            }
        },
        
        /**
         * Handle form input change
         * @param {Event} event - Input event
         */
        onFormChange(event) {
            this.markDirty();
            
            // Clear validation error for changed field
            if (event.target && event.target.name) {
                this.clearFieldError(event.target.name);
            }
        },
        
        /**
         * Handle form input
         * @param {Event} event - Input event
         */
        onFormInput(event) {
            if (event.target && event.target.name) {
                this.clearFieldError(event.target.name);
            }
        },
        
        /**
         * Reset form modal to initial state
         */
        reset() {
            this.formData = JSON.parse(JSON.stringify(this.initialFormData));
            this.validationErrors = {};
            this.localLoading = false;
            this.isDirty = false;
        },
        
        /**
         * Override onClosed to handle dirty state
         * @param {Object} detail - Event detail
         */
        onClosed(detail) {
            // Call parent onClosed if it exists
            if (modalRest.onClosed) {
                modalRest.onClosed.call(this, detail);
            }
            
            // Reset dirty state
            this.isDirty = false;
            this.saveInitialFormData();
        },
        
        /**
         * Override onSubmitSuccess to mark clean
         * @param {Object} response - Server response
         */
        onSubmitSuccess(response) {
            this.markClean();
            
            // Call parent if exists
            if (modalRest.onSubmitSuccess) {
                modalRest.onSubmitSuccess.call(this, response);
            }
        },
        
        // Computed properties
        
        /**
         * Check if form has unsaved changes
         * @returns {boolean} True if form is dirty
         */
        get hasUnsavedChanges() {
            return this.isDirty;
        },
        
        /**
         * Check if can submit (not loading and either dirty or no dirty check)
         * @returns {boolean} True if can submit
         */
        get canSubmit() {
            return !this.localLoading && !this.hasErrors;
        }
    };
}

/**
 * Service-specific Modal configurations
 */
const ModalConfigs = {
    /**
     * Console service modal configuration
     */
    console: {
        size: 'lg',
        position: 'center',
        backdrop: 'static',
        closable: true,
        keyboard: true
    },
    
    /**
     * Workspace service modal configuration
     */
    workspace: {
        size: 'md',
        position: 'center',
        backdrop: 'dismissible',
        closable: true,
        keyboard: true
    },
    
    /**
     * Portal service modal configuration
     */
    portal: {
        size: 'sm',
        position: 'center',
        backdrop: 'static',
        closable: true,
        keyboard: true
    },
    
    /**
     * Full-screen modal configuration
     */
    fullscreen: {
        size: 'full',
        position: 'center',
        backdrop: 'static',
        closable: true,
        keyboard: true
    },
    
    /**
     * Sidebar modal configuration
     */
    sidebar: {
        size: 'md',
        position: 'right',
        backdrop: 'dismissible',
        closable: true,
        keyboard: true
    }
};

/**
 * Create a console service modal
 * @param {UseModalConfig} [config={}] - Additional configuration
 * @returns {Object} Modal component data
 */
function createConsoleModal(config = {}) {
    return useModal({
        ...ModalConfigs.console,
        ...config
    });
}

/**
 * Create a workspace service modal
 * @param {UseModalConfig} [config={}] - Additional configuration
 * @returns {Object} Modal component data
 */
function createWorkspaceModal(config = {}) {
    return useModal({
        ...ModalConfigs.workspace,
        ...config
    });
}

/**
 * Create a portal service modal
 * @param {UseModalConfig} [config={}] - Additional configuration
 * @returns {Object} Modal component data
 */
function createPortalModal(config = {}) {
    return useModal({
        ...ModalConfigs.portal,
        ...config
    });
}

/**
 * Create a fullscreen modal
 * @param {UseModalConfig} [config={}] - Additional configuration
 * @returns {Object} Modal component data
 */
function createFullscreenModal(config = {}) {
    return useModal({
        ...ModalConfigs.fullscreen,
        ...config
    });
}

/**
 * Create a sidebar modal
 * @param {UseModalConfig} [config={}] - Additional configuration
 * @returns {Object} Modal component data
 */
function createSidebarModal(config = {}) {
    return useModal({
        ...ModalConfigs.sidebar,
        ...config
    });
}

/**
 * Confirmation Dialog Helper
 * 
 * Shows a modal confirmation dialog
 * 
 * @param {Object} options - Confirmation options
 * @param {string} options.title - Dialog title
 * @param {string} options.message - Dialog message
 * @param {string} [options.confirmText='Confirm'] - Confirm button text
 * @param {string} [options.cancelText='Cancel'] - Cancel button text
 * @param {string} [options.type='warning'] - Dialog type (warning, danger, info)
 * @returns {Promise<boolean>} True if confirmed
 */
async function confirmDialog(options) {
    return new Promise((resolve) => {
        const {
            title = 'Confirm',
            message = 'Are you sure?',
            confirmText = 'Confirm',
            cancelText = 'Cancel',
            type = 'warning'
        } = options;
        
        // Try to use a custom confirm modal if available
        const confirmModal = document.getElementById('confirm-dialog');
        
        if (confirmModal) {
            // Use custom modal
            const titleEl = confirmModal.querySelector('[data-confirm-title]');
            const messageEl = confirmModal.querySelector('[data-confirm-message]');
            const confirmBtn = confirmModal.querySelector('[data-confirm-button]');
            const cancelBtn = confirmModal.querySelector('[data-cancel-button]');
            
            if (titleEl) titleEl.textContent = title;
            if (messageEl) messageEl.textContent = message;
            if (confirmBtn) confirmBtn.textContent = confirmText;
            if (cancelBtn) cancelBtn.textContent = cancelText;
            
            // Set type class
            confirmModal.className = confirmModal.className.replace(/\btype-\w+\b/g, '');
            confirmModal.classList.add(`type-${type}`);
            
            // Setup event handlers
            const handleConfirm = () => {
                cleanup();
                Alpine.store('modal').close('confirm-dialog');
                resolve(true);
            };
            
            const handleCancel = () => {
                cleanup();
                Alpine.store('modal').close('confirm-dialog');
                resolve(false);
            };
            
            const cleanup = () => {
                if (confirmBtn) confirmBtn.removeEventListener('click', handleConfirm);
                if (cancelBtn) cancelBtn.removeEventListener('click', handleCancel);
                window.removeEventListener('modal:closed', handleModalClosed);
            };
            
            const handleModalClosed = (e) => {
                if (e.detail.modalId === 'confirm-dialog') {
                    cleanup();
                    resolve(false);
                }
            };
            
            if (confirmBtn) confirmBtn.addEventListener('click', handleConfirm);
            if (cancelBtn) cancelBtn.addEventListener('click', handleCancel);
            window.addEventListener('modal:closed', handleModalClosed);
            
            Alpine.store('modal').open('confirm-dialog', { backdrop: 'static' });
        } else {
            // Fallback to browser confirm
            resolve(window.confirm(`${title}\n\n${message}`));
        }
    });
}

/**
 * Alert Dialog Helper
 * 
 * Shows a modal alert dialog
 * 
 * @param {Object} options - Alert options
 * @param {string} options.title - Dialog title
 * @param {string} options.message - Dialog message
 * @param {string} [options.okText='OK'] - OK button text
 * @param {string} [options.type='info'] - Dialog type (success, error, warning, info)
 * @returns {Promise<void>}
 */
async function alertDialog(options) {
    return new Promise((resolve) => {
        const {
            title = 'Alert',
            message = '',
            okText = 'OK',
            type = 'info'
        } = options;
        
        // Try to use a custom alert modal if available
        const alertModal = document.getElementById('alert-dialog');
        
        if (alertModal) {
            // Use custom modal
            const titleEl = alertModal.querySelector('[data-alert-title]');
            const messageEl = alertModal.querySelector('[data-alert-message]');
            const okBtn = alertModal.querySelector('[data-ok-button]');
            
            if (titleEl) titleEl.textContent = title;
            if (messageEl) messageEl.textContent = message;
            if (okBtn) okBtn.textContent = okText;
            
            // Set type class
            alertModal.className = alertModal.className.replace(/\btype-\w+\b/g, '');
            alertModal.classList.add(`type-${type}`);
            
            // Setup event handlers
            const handleOk = () => {
                cleanup();
                Alpine.store('modal').close('alert-dialog');
                resolve();
            };
            
            const cleanup = () => {
                if (okBtn) okBtn.removeEventListener('click', handleOk);
                window.removeEventListener('modal:closed', handleModalClosed);
            };
            
            const handleModalClosed = (e) => {
                if (e.detail.modalId === 'alert-dialog') {
                    cleanup();
                    resolve();
                }
            };
            
            if (okBtn) okBtn.addEventListener('click', handleOk);
            window.addEventListener('modal:closed', handleModalClosed);
            
            Alpine.store('modal').open('alert-dialog', { backdrop: 'static' });
        } else {
            // Fallback to browser alert
            window.alert(`${title}\n\n${message}`);
            resolve();
        }
    });
}

/**
 * Prompt Dialog Helper
 * 
 * Shows a modal prompt dialog for user input
 * 
 * @param {Object} options - Prompt options
 * @param {string} options.title - Dialog title
 * @param {string} options.message - Dialog message
 * @param {string} [options.defaultValue=''] - Default input value
 * @param {string} [options.placeholder=''] - Input placeholder
 * @param {string} [options.okText='OK'] - OK button text
 * @param {string} [options.cancelText='Cancel'] - Cancel button text
 * @param {string} [options.inputType='text'] - Input type
 * @returns {Promise<string|null>} User input or null if cancelled
 */
async function promptDialog(options) {
    return new Promise((resolve) => {
        const {
            title = 'Input',
            message = '',
            defaultValue = '',
            placeholder = '',
            okText = 'OK',
            cancelText = 'Cancel',
            inputType = 'text'
        } = options;
        
        // Try to use a custom prompt modal if available
        const promptModal = document.getElementById('prompt-dialog');
        
        if (promptModal) {
            // Use custom modal
            const titleEl = promptModal.querySelector('[data-prompt-title]');
            const messageEl = promptModal.querySelector('[data-prompt-message]');
            const inputEl = promptModal.querySelector('[data-prompt-input]');
            const okBtn = promptModal.querySelector('[data-ok-button]');
            const cancelBtn = promptModal.querySelector('[data-cancel-button]');
            
            if (titleEl) titleEl.textContent = title;
            if (messageEl) messageEl.textContent = message;
            if (okBtn) okBtn.textContent = okText;
            if (cancelBtn) cancelBtn.textContent = cancelText;
            
            if (inputEl) {
                inputEl.value = defaultValue;
                inputEl.placeholder = placeholder;
                inputEl.type = inputType;
            }
            
            // Setup event handlers
            const handleOk = () => {
                const value = inputEl ? inputEl.value : null;
                cleanup();
                Alpine.store('modal').close('prompt-dialog');
                resolve(value);
            };
            
            const handleCancel = () => {
                cleanup();
                Alpine.store('modal').close('prompt-dialog');
                resolve(null);
            };
            
            const handleKeydown = (e) => {
                if (e.key === 'Enter') {
                    handleOk();
                }
            };
            
            const cleanup = () => {
                if (okBtn) okBtn.removeEventListener('click', handleOk);
                if (cancelBtn) cancelBtn.removeEventListener('click', handleCancel);
                if (inputEl) inputEl.removeEventListener('keydown', handleKeydown);
                window.removeEventListener('modal:closed', handleModalClosed);
            };
            
            const handleModalClosed = (e) => {
                if (e.detail.modalId === 'prompt-dialog') {
                    cleanup();
                    resolve(null);
                }
            };
            
            if (okBtn) okBtn.addEventListener('click', handleOk);
            if (cancelBtn) cancelBtn.addEventListener('click', handleCancel);
            if (inputEl) inputEl.addEventListener('keydown', handleKeydown);
            window.addEventListener('modal:closed', handleModalClosed);
            
            Alpine.store('modal').open('prompt-dialog', { backdrop: 'static' });
            
            // Focus input after modal opens
            setTimeout(() => {
                if (inputEl) inputEl.focus();
            }, 100);
        } else {
            // Fallback to browser prompt
            const result = window.prompt(`${title}\n\n${message}`, defaultValue);
            resolve(result);
        }
    });
}

/**
 * Export hooks and utilities to window
 */
if (typeof window !== 'undefined') {
    window.useModal = useModal;
    window.useFormModal = useFormModal;
    window.ModalConfigs = ModalConfigs;
    window.createConsoleModal = createConsoleModal;
    window.createWorkspaceModal = createWorkspaceModal;
    window.createPortalModal = createPortalModal;
    window.createFullscreenModal = createFullscreenModal;
    window.createSidebarModal = createSidebarModal;
    window.confirmDialog = confirmDialog;
    window.alertDialog = alertDialog;
    window.promptDialog = promptDialog;
}

/**
 * Export for module systems
 */
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        useModal,
        useFormModal,
        ModalConfigs,
        createConsoleModal,
        createWorkspaceModal,
        createPortalModal,
        createFullscreenModal,
        createSidebarModal,
        confirmDialog,
        alertDialog,
        promptDialog
    };
}
