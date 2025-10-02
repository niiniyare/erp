/**
 * Modal AlpineJS Hooks (TypeScript)
 * 
 * Extracted from internal/ui/components/layout/modal.templ
 * Provides reusable Modal functionality for all ERP services
 * 
 * @version 2.0.0
 */

/// <reference path="./types.d.ts" />

// Types
interface ModalConfig {
    backdrop?: 'static' | 'dismissible';
    keyboard?: boolean;
    focus?: boolean;
    size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
    position?: 'center' | 'top' | 'bottom' | 'left' | 'right';
}

interface ModalSubmitOptions {
    method?: string;
    headers?: Record<string, string>;
    onSuccess?: (response: any) => void;
    onError?: (response: any) => void;
}

interface RequestOptions {
    method?: string;
    data?: FormData | Record<string, any>;
    headers?: Record<string, string>;
}

interface ModalStackItem {
    id: string;
    isOpen: boolean;
    backdrop: 'static' | 'dismissible';
    keyboard: boolean;
    focus: boolean;
}

interface ModalStore {
    isOpen: boolean;
    currentModal: string;
    modalStack: ModalStackItem[];
    backdrop: 'static' | 'dismissible';
    keyboard: boolean;
    focus: boolean;
    _scrollPosition: number;
    _activeElement: Element | null;
    _focusTrap: ((e: KeyboardEvent) => void) | null;
    _keydownHandler: ((e: KeyboardEvent) => void) | null;
    
    open(modalId: string, config?: ModalConfig): void;
    close(modalId?: string | null): boolean;
    toggle(modalId: string, config?: ModalConfig): void;
    closeAll(): void;
    isModalOpen(modalId: string): boolean;
    submit(modalId: string, formData: FormData | Record<string, any>, endpoint: string, options?: ModalSubmitOptions): Promise<any>;
    _disableBodyScroll(): void;
    _enableBodyScroll(): void;
    _setInitialFocus(modalId: string): void;
    _setupKeyboardHandlers(modalId: string): void;
    _cleanupKeyboardHandlers(): void;
    _setupFocusTrap(modalId: string): void;
    _cleanupFocusTrap(): void;
    _makeRequest(url: string, options: RequestOptions): Promise<any>;
    _showValidationErrors(errors: Record<string, string | string[]>): void;
    _showToast(type: string, message: string): void;
    _formDataToObject(formData: FormData | Record<string, any>): Record<string, any>;
    _dispatchEvent(eventName: string, detail: any): void;
}

interface UseModalConfig extends ModalConfig {
    id: string;
    closable?: boolean;
}

interface ModalComponent {
    id: string;
    size: string;
    position: string;
    backdrop: 'static' | 'dismissible';
    closable: boolean;
    keyboard: boolean;
    localLoading: boolean;
    formData: Record<string, any>;
    validationErrors: Record<string, string[]>;
    _eventListeners: Array<{type: string, handler: EventListener, element: EventTarget}>;
    _destroyed: boolean;
    
    init(): void;
    destroy(): void;
    _addEventListener(type: string, handler: EventListener, target?: EventTarget): void;
    open(): void;
    close(): void;
    toggle(): void;
    submitForm(formElementOrEvent: HTMLFormElement | Event, endpoint: string, options?: ModalSubmitOptions): Promise<any>;
    handleBackdropClick(event: MouseEvent): void;
    reset(): void;
    onOpened(detail: any): void;
    onClosed(detail: any): void;
    onSubmitSuccess(response: any): void;
    onSubmitError(response: any): void;
    getFieldError(field: string): string | null;
    hasFieldError(field: string): boolean;
    clearFieldError(field: string): void;
    clearAllErrors(): void;
    
    readonly isCurrentModal: boolean;
    readonly isOpen: boolean;
    readonly hasErrors: boolean;
    readonly canSubmit: boolean;
}

interface UseFormModalConfig extends UseModalConfig {
    saveOnClose?: boolean;
    confirmOnClose?: boolean;
    saveEndpoint?: string;
    customConfirm?: () => Promise<boolean>;
}

interface FormModalComponent extends ModalComponent {
    saveOnClose: boolean;
    confirmOnClose: boolean;
    saveEndpoint: string;
    customConfirm: (() => Promise<boolean>) | null;
    isDirty: boolean;
    initialFormData: Record<string, any>;
    
    saveInitialFormData(): void;
    _checkDirty(): void;
    markDirty(): void;
    markClean(): void;
    close(): Promise<boolean>;
    confirmClose(): Promise<boolean>;
    _confirmCloseAfterSaveFailed(): Promise<boolean>;
    autoSave(): Promise<boolean>;
    onFormChange(event: Event): void;
    onFormInput(event: Event): void;
    reset(): void;
    onClosed(detail: any): void;
    onSubmitSuccess(response: any): void;
    
    readonly hasUnsavedChanges: boolean;
}

interface DialogOptions {
    title?: string;
    message?: string;
    confirmText?: string;
    cancelText?: string;
    type?: 'warning' | 'danger' | 'info' | 'success' | 'error';
}

interface AlertOptions {
    title?: string;
    message?: string;
    okText?: string;
    type?: 'success' | 'error' | 'warning' | 'info';
}

interface PromptOptions {
    title?: string;
    message?: string;
    defaultValue?: string;
    placeholder?: string;
    okText?: string;
    cancelText?: string;
    inputType?: string;
}

// Export to make this a module
export {};

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
    (window as any).Alpine.store('modal', {
        // State
        isOpen: false,
        currentModal: '',
        modalStack: [],
        
        // Configuration
        backdrop: 'static' as 'static' | 'dismissible',
        keyboard: true,
        focus: true,
        
        // Internal state
        _scrollPosition: 0,
        _activeElement: null,
        _focusTrap: null,
        _keydownHandler: null,
        
        /**
         * Open a modal
         * @param modalId - Modal identifier
         * @param config - Modal configuration
         */
        open(modalId: string, config: ModalConfig = {}): void {
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
         * @param modalId - Modal to close (null for current)
         * @returns True if modal was closed
         */
        close(modalId: string | null = null): boolean {
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
                if (this._activeElement && (this._activeElement as HTMLElement).focus) {
                    try {
                        (this._activeElement as HTMLElement).focus();
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
         * @param modalId - Modal identifier
         * @param config - Modal configuration
         */
        toggle(modalId: string, config: ModalConfig = {}): void {
            if (this.isOpen && this.currentModal === modalId) {
                this.close(modalId);
            } else {
                this.open(modalId, config);
            }
        },
        
        /**
         * Close all modals
         */
        closeAll(): void {
            this._cleanupKeyboardHandlers();
            this._cleanupFocusTrap();
            
            this.modalStack = [];
            this.isOpen = false;
            this.currentModal = '';
            
            this._enableBodyScroll();
            
            if (this._activeElement && (this._activeElement as HTMLElement).focus) {
                try {
                    (this._activeElement as HTMLElement).focus();
                } catch (e) {
                    // Element might be removed from DOM
                }
            }
            this._activeElement = null;
            
            this._dispatchEvent('modal:all-closed', {});
        },
        
        /**
         * Check if a specific modal is currently open
         * @param modalId - Modal identifier
         * @returns True if modal is open
         */
        isModalOpen(modalId: string): boolean {
            return this.isOpen && this.currentModal === modalId;
        },
        
        /**
         * Submit modal form
         * @param modalId - Modal identifier
         * @param formData - Form data to submit
         * @param endpoint - Submission endpoint
         * @param options - Submission options
         * @returns Response object
         */
        async submit(modalId: string, formData: FormData | Record<string, any>, endpoint: string, options: ModalSubmitOptions = {}): Promise<any> {
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
                
            } catch (error: any) {
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
        _disableBodyScroll(): void {
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
        _enableBodyScroll(): void {
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
         * @param modalId - Modal identifier
         * @private
         */
        _setInitialFocus(modalId: string): void {
            requestAnimationFrame(() => {
                const modal = document.getElementById(modalId);
                if (!modal) return;
                
                // Try to find autofocus element first
                let focusElement = modal.querySelector('[autofocus]') as HTMLElement;
                
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
                    focusElement = focusableElements[0] as HTMLElement;
                }
                
                if (focusElement && focusElement.focus) {
                    focusElement.focus();
                }
            });
        },
        
        /**
         * Setup keyboard event handlers
         * @param modalId - Modal identifier
         * @private
         */
        _setupKeyboardHandlers(modalId: string): void {
            this._cleanupKeyboardHandlers();
            
            this._keydownHandler = (e: KeyboardEvent) => {
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
        _cleanupKeyboardHandlers(): void {
            if (this._keydownHandler) {
                document.removeEventListener('keydown', this._keydownHandler);
                this._keydownHandler = null;
            }
        },
        
        /**
         * Setup focus trap for accessibility
         * @param modalId - Modal identifier
         * @private
         */
        _setupFocusTrap(modalId: string): void {
            const modal = document.getElementById(modalId);
            if (!modal) return;
            
            this._focusTrap = (e: KeyboardEvent) => {
                if (e.key !== 'Tab') return;
                
                const focusableSelector = [
                    'button:not([disabled])',
                    '[href]:not([disabled])',
                    'input:not([disabled]):not([type="hidden"])',
                    'select:not([disabled])',
                    'textarea:not([disabled])',
                    '[tabindex]:not([tabindex="-1"]):not([disabled])'
                ].join(', ');
                
                const focusableElements = Array.from(modal.querySelectorAll(focusableSelector)) as HTMLElement[];
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
        _cleanupFocusTrap(): void {
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
         * @param url - Request URL
         * @param options - Request options
         * @returns Response data
         * @private
         */
        async _makeRequest(url: string, options: RequestOptions): Promise<any> {
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
            const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
            
            const headers: Record<string, string> = {
                'X-Requested-With': 'XMLHttpRequest',
                ...(options.headers || {})
            };
            
            if (csrfToken) {
                headers['X-CSRF-Token'] = csrfToken;
            }
            
            const response = await fetch(url, {
                method: options.method || 'POST',
                body: body as BodyInit,
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
         * @param errors - Validation errors
         * @private
         */
        _showValidationErrors(errors: Record<string, string | string[]>): void {
            // Clear previous errors
            document.querySelectorAll('.validation-error, [data-validation-for]').forEach(el => {
                if (el.hasAttribute('data-validation-for')) {
                    (el as HTMLElement).textContent = '';
                    (el as HTMLElement).style.display = 'none';
                }
            });
            
            // Show new errors
            Object.entries(errors).forEach(([field, messages]) => {
                const errorElement = document.querySelector(`[data-validation-for="${field}"]`) as HTMLElement;
                if (errorElement) {
                    const message = Array.isArray(messages) ? messages[0] : messages;
                    errorElement.textContent = message;
                    errorElement.style.display = 'block';
                    
                    // Also add error class to input
                    const inputElement = document.querySelector(`[name="${field}"]`) as HTMLElement;
                    if (inputElement) {
                        inputElement.classList.add('error', 'is-invalid');
                    }
                }
            });
        },
        
        /**
         * Show toast message
         * @param type - Toast type (success, error, info, warning)
         * @param message - Message to display
         * @private
         */
        _showToast(type: string, message: string): void {
            const toastStore = (window as any).Alpine?.store?.('toast');
            if (toastStore && typeof toastStore[type] === 'function') {
                toastStore[type](message);
            }
        },
        
        /**
         * Convert FormData to plain object
         * @param formData - Form data
         * @returns Plain object
         * @private
         */
        _formDataToObject(formData: FormData | Record<string, any>): Record<string, any> {
            if (!(formData instanceof FormData)) {
                return formData;
            }
            
            const obj: Record<string, any> = {};
            for (const [key, value] of (formData as any).entries()) {
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
         * @param eventName - Event name
         * @param detail - Event detail
         * @private
         */
        _dispatchEvent(eventName: string, detail: any): void {
            window.dispatchEvent(new CustomEvent(eventName, {
                detail,
                bubbles: true,
                cancelable: true
            }));
        }
    } as ModalStore);
});

/**
 * Modal Component Hook
 * 
 * Provides instance-level modal management with form handling
 * and lifecycle hooks.
 * 
 * @param config - Configuration options
 * @returns Alpine.js component data
 * 
 * @example
 * <div x-data="useModal({ id: 'my-modal', size: 'lg' })">
 */
function useModal(config: UseModalConfig = {} as UseModalConfig): ModalComponent {
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
        init(): void {
            if (!this.id) {
                console.error('Modal ID is required');
                return;
            }
            
            // Setup event listeners
            this._addEventListener('modal:opened', (e: any) => {
                if (e.detail.modalId === this.id) {
                    this.onOpened(e.detail);
                }
            });
            
            this._addEventListener('modal:closed', (e: any) => {
                if (e.detail.modalId === this.id) {
                    this.onClosed(e.detail);
                }
            });
            
            // Keyboard event handling for dismissible modals
            if (this.backdrop === 'dismissible' && this.keyboard) {
                const keyHandler: EventListener = (e: Event) => {
                    const keyEvent = e as KeyboardEvent;
                    if (keyEvent.key === 'Escape' && this.isCurrentModal) {
                        this.close();
                    }
                };
                (this as any).$el.addEventListener('keydown', keyHandler);
                this._eventListeners.push({ type: 'keydown', handler: keyHandler, element: (this as any).$el });
            }
        },
        
        /**
         * Cleanup on component destroy
         */
        destroy(): void {
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
         * @param type - Event type
         * @param handler - Event handler
         * @param target - Event target
         * @private
         */
        _addEventListener(type: string, handler: EventListener, target: EventTarget = window): void {
            target.addEventListener(type, handler);
            this._eventListeners.push({ type, handler, element: target });
        },
        
        /**
         * Open this modal
         */
        open(): void {
            if (this._destroyed) {
                console.warn('Cannot open destroyed modal');
                return;
            }
            
            (window as any).Alpine.store('modal').open(this.id, {
                backdrop: this.backdrop,
                keyboard: this.keyboard,
                size: this.size,
                position: this.position
            });
        },
        
        /**
         * Close this modal
         */
        close(): void {
            if (this.closable) {
                (window as any).Alpine.store('modal').close(this.id);
            }
        },
        
        /**
         * Toggle this modal
         */
        toggle(): void {
            (window as any).Alpine.store('modal').toggle(this.id, {
                backdrop: this.backdrop,
                keyboard: this.keyboard,
                size: this.size,
                position: this.position
            });
        },
        
        /**
         * Submit form within modal
         * @param formElementOrEvent - Form element or submit event
         * @param endpoint - Submission endpoint
         * @param options - Submission options
         * @returns Response object
         */
        async submitForm(formElementOrEvent: HTMLFormElement | Event, endpoint: string, options: ModalSubmitOptions = {}): Promise<any> {
            if (this.localLoading) {
                return { success: false, error: 'Form is already submitting' };
            }
            
            // Extract form element
            let formElement = formElementOrEvent;
            if (formElementOrEvent instanceof Event) {
                formElementOrEvent.preventDefault();
                formElement = formElementOrEvent.target as HTMLFormElement;
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
                
                const response = await (window as any).Alpine.store('modal').submit(
                    this.id,
                    formData,
                    endpoint,
                    {
                        ...options,
                        onSuccess: (res: any) => {
                            this.onSubmitSuccess(res);
                            if (options.onSuccess) options.onSuccess(res);
                        },
                        onError: (res: any) => {
                            this.onSubmitError(res);
                            if (options.onError) options.onError(res);
                        }
                    }
                );
                
                if (!response.success && response.errors) {
                    this.validationErrors = response.errors;
                }
                
                return response;
                
            } catch (error: any) {
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
         * @param event - Click event
         */
        handleBackdropClick(event: MouseEvent): void {
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
        reset(): void {
            this.formData = {};
            this.validationErrors = {};
            this.localLoading = false;
        },
        
        /**
         * Lifecycle hook - called when modal opens
         * @param detail - Event detail
         */
        onOpened(detail: any): void {
            // Override in specific modals
        },
        
        /**
         * Lifecycle hook - called when modal closes
         * @param detail - Event detail
         */
        onClosed(detail: any): void {
            this.reset();
        },
        
        /**
         * Lifecycle hook - called on successful form submission
         * @param response - Server response
         */
        onSubmitSuccess(response: any): void {
            // Override in specific modals
        },
        
        /**
         * Lifecycle hook - called on form submission error
         * @param response - Error response
         */
        onSubmitError(response: any): void {
            // Override in specific modals
        },
        
        /**
         * Get validation error for a field
         * @param field - Field name
         * @returns First error message or null
         */
        getFieldError(field: string): string | null {
            const errors = this.validationErrors[field];
            if (!errors) return null;
            return Array.isArray(errors) ? errors[0] : errors;
        },
        
        /**
         * Check if field has error
         * @param field - Field name
         * @returns True if field has error
         */
        hasFieldError(field: string): boolean {
            return !!this.validationErrors[field];
        },
        
        /**
         * Clear validation error for a field
         * @param field - Field name
         */
        clearFieldError(field: string): void {
            if (this.validationErrors[field]) {
                delete this.validationErrors[field];
            }
        },
        
        /**
         * Clear all validation errors
         */
        clearAllErrors(): void {
            this.validationErrors = {};
        },
        
        // Computed properties
        
        /**
         * Check if this is the current modal
         * @returns True if current modal
         */
        get isCurrentModal(): boolean {
            return (window as any).Alpine.store('modal').currentModal === this.id;
        },
        
        /**
         * Check if modal is open
         * @returns True if open
         */
        get isOpen(): boolean {
            return this.isCurrentModal && (window as any).Alpine.store('modal').isOpen;
        },
        
        /**
         * Check if form has validation errors
         * @returns True if has errors
         */
        get hasErrors(): boolean {
            return Object.keys(this.validationErrors).length > 0;
        },
        
        /**
         * Check if form can be submitted
         * @returns True if can submit
         */
        get canSubmit(): boolean {
            return !this.localLoading && !this.hasErrors;
        }
    };
}

/**
 * Form Modal Hook
 * 
 * Specialized modal for forms with dirty tracking and auto-save support
 * 
 * @param config - Configuration options
 * @returns Alpine.js component data
 * 
 * @example
 * <div x-data="useFormModal({ 
 *   id: 'edit-modal', 
 *   confirmOnClose: true,
 *   saveOnClose: false
 * })">
 */
function useFormModal(config: UseFormModalConfig = {} as UseFormModalConfig): FormModalComponent {
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
        init(): void {
            if (modalInit) modalInit.call(this);
            
            // Save initial form data
            this.saveInitialFormData();
            
            // Watch for form changes
            if (typeof (this as any).$watch === 'function') {
                (this as any).$watch('formData', () => {
                    this._checkDirty();
                }, { deep: true });
            }
        },
        
        /**
         * Cleanup
         */
        destroy(): void {
            if (modalDestroy) modalDestroy.call(this);
        },
        
        /**
         * Save initial form data for dirty checking
         */
        saveInitialFormData(): void {
            this.initialFormData = JSON.parse(JSON.stringify(this.formData || {}));
        },
        
        /**
         * Check if form data has changed
         * @private
         */
        _checkDirty(): void {
            this.isDirty = JSON.stringify(this.formData) !== JSON.stringify(this.initialFormData);
        },
        
        /**
         * Mark form as dirty
         */
        markDirty(): void {
            this.isDirty = true;
        },
        
        /**
         * Mark form as clean (pristine)
         */
        markClean(): void {
            this.isDirty = false;
            this.saveInitialFormData();
        },
        
        /**
         * Close modal with dirty check
         * @returns True if closed successfully
         */
        async close(): Promise<boolean> {
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
         * @returns True if user confirms
         */
        async confirmClose(): Promise<boolean> {
            if (this.customConfirm && typeof this.customConfirm === 'function') {
                return await this.customConfirm();
            }
            
            // Default browser confirm
            return window.confirm('You have unsaved changes. Are you sure you want to close?');
        },
        
        /**
         * Confirm closing after auto-save failed
         * @returns True if user wants to proceed
         * @private
         */
        async _confirmCloseAfterSaveFailed(): Promise<boolean> {
            return window.confirm('Auto-save failed. Close without saving?');
        },
        
        /**
         * Auto-save form data
         * @returns True if saved successfully
         */
        async autoSave(): Promise<boolean> {
            if (!this.saveEndpoint) {
                console.warn('No save endpoint configured for auto-save');
                return false;
            }
            
            const formElement = (this as any).$el.querySelector('form') as HTMLFormElement;
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
         * @param event - Input event
         */
        onFormChange(event: Event): void {
            this.markDirty();
            
            // Clear validation error for changed field
            if (event.target && (event.target as HTMLInputElement).name) {
                this.clearFieldError((event.target as HTMLInputElement).name);
            }
        },
        
        /**
         * Handle form input
         * @param event - Input event
         */
        onFormInput(event: Event): void {
            if (event.target && (event.target as HTMLInputElement).name) {
                this.clearFieldError((event.target as HTMLInputElement).name);
            }
        },
        
        /**
         * Reset form modal to initial state
         */
        reset(): void {
            this.formData = JSON.parse(JSON.stringify(this.initialFormData));
            this.validationErrors = {};
            this.localLoading = false;
            this.isDirty = false;
        },
        
        /**
         * Override onClosed to handle dirty state
         * @param detail - Event detail
         */
        onClosed(detail: any): void {
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
         * @param response - Server response
         */
        onSubmitSuccess(response: any): void {
            this.markClean();
            
            // Call parent if exists
            if (modalRest.onSubmitSuccess) {
                modalRest.onSubmitSuccess.call(this, response);
            }
        },
        
        // Computed properties
        
        /**
         * Check if form has unsaved changes
         * @returns True if form is dirty
         */
        get hasUnsavedChanges(): boolean {
            return this.isDirty;
        },
        
        /**
         * Check if can submit (not loading and either dirty or no dirty check)
         * @returns True if can submit
         */
        get canSubmit(): boolean {
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
 * @param config - Additional configuration
 * @returns Modal component data
 */
function createConsoleModal(config: UseModalConfig = {} as UseModalConfig): ModalComponent {
    return useModal({
        ...ModalConfigs.console,
        ...config
    } as UseModalConfig);
}

/**
 * Create a workspace service modal
 * @param config - Additional configuration
 * @returns Modal component data
 */
function createWorkspaceModal(config: UseModalConfig = {} as UseModalConfig): ModalComponent {
    return useModal({
        ...ModalConfigs.workspace,
        ...config
    } as UseModalConfig);
}

/**
 * Create a portal service modal
 * @param config - Additional configuration
 * @returns Modal component data
 */
function createPortalModal(config: UseModalConfig = {} as UseModalConfig): ModalComponent {
    return useModal({
        ...ModalConfigs.portal,
        ...config
    } as UseModalConfig);
}

/**
 * Create a fullscreen modal
 * @param config - Additional configuration
 * @returns Modal component data
 */
function createFullscreenModal(config: UseModalConfig = {} as UseModalConfig): ModalComponent {
    return useModal({
        ...ModalConfigs.fullscreen,
        ...config
    } as UseModalConfig);
}

/**
 * Create a sidebar modal
 * @param config - Additional configuration
 * @returns Modal component data
 */
function createSidebarModal(config: UseModalConfig = {} as UseModalConfig): ModalComponent {
    return useModal({
        ...ModalConfigs.sidebar,
        ...config
    } as UseModalConfig);
}

/**
 * Confirmation Dialog Helper
 * 
 * Shows a modal confirmation dialog
 * 
 * @param options - Confirmation options
 * @returns True if confirmed
 */
async function confirmDialog(options: DialogOptions): Promise<boolean> {
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
            const titleEl = confirmModal.querySelector('[data-confirm-title]') as HTMLElement;
            const messageEl = confirmModal.querySelector('[data-confirm-message]') as HTMLElement;
            const confirmBtn = confirmModal.querySelector('[data-confirm-button]') as HTMLElement;
            const cancelBtn = confirmModal.querySelector('[data-cancel-button]') as HTMLElement;
            
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
                (window as any).Alpine.store('modal').close('confirm-dialog');
                resolve(true);
            };
            
            const handleCancel = () => {
                cleanup();
                (window as any).Alpine.store('modal').close('confirm-dialog');
                resolve(false);
            };
            
            const cleanup = () => {
                if (confirmBtn) confirmBtn.removeEventListener('click', handleConfirm);
                if (cancelBtn) cancelBtn.removeEventListener('click', handleCancel);
                window.removeEventListener('modal:closed', handleModalClosed);
            };
            
            const handleModalClosed = (e: any) => {
                if (e.detail.modalId === 'confirm-dialog') {
                    cleanup();
                    resolve(false);
                }
            };
            
            if (confirmBtn) confirmBtn.addEventListener('click', handleConfirm);
            if (cancelBtn) cancelBtn.addEventListener('click', handleCancel);
            window.addEventListener('modal:closed', handleModalClosed);
            
            (window as any).Alpine.store('modal').open('confirm-dialog', { backdrop: 'static' });
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
 * @param options - Alert options
 * @returns Promise that resolves when closed
 */
async function alertDialog(options: AlertOptions): Promise<void> {
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
            const titleEl = alertModal.querySelector('[data-alert-title]') as HTMLElement;
            const messageEl = alertModal.querySelector('[data-alert-message]') as HTMLElement;
            const okBtn = alertModal.querySelector('[data-ok-button]') as HTMLElement;
            
            if (titleEl) titleEl.textContent = title;
            if (messageEl) messageEl.textContent = message;
            if (okBtn) okBtn.textContent = okText;
            
            // Set type class
            alertModal.className = alertModal.className.replace(/\btype-\w+\b/g, '');
            alertModal.classList.add(`type-${type}`);
            
            // Setup event handlers
            const handleOk = () => {
                cleanup();
                (window as any).Alpine.store('modal').close('alert-dialog');
                resolve();
            };
            
            const cleanup = () => {
                if (okBtn) okBtn.removeEventListener('click', handleOk);
                window.removeEventListener('modal:closed', handleModalClosed);
            };
            
            const handleModalClosed = (e: any) => {
                if (e.detail.modalId === 'alert-dialog') {
                    cleanup();
                    resolve();
                }
            };
            
            if (okBtn) okBtn.addEventListener('click', handleOk);
            window.addEventListener('modal:closed', handleModalClosed);
            
            (window as any).Alpine.store('modal').open('alert-dialog', { backdrop: 'static' });
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
 * @param options - Prompt options
 * @returns User input or null if cancelled
 */
async function promptDialog(options: PromptOptions): Promise<string | null> {
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
            const titleEl = promptModal.querySelector('[data-prompt-title]') as HTMLElement;
            const messageEl = promptModal.querySelector('[data-prompt-message]') as HTMLElement;
            const inputEl = promptModal.querySelector('[data-prompt-input]') as HTMLInputElement;
            const okBtn = promptModal.querySelector('[data-ok-button]') as HTMLElement;
            const cancelBtn = promptModal.querySelector('[data-cancel-button]') as HTMLElement;
            
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
                (window as any).Alpine.store('modal').close('prompt-dialog');
                resolve(value);
            };
            
            const handleCancel = () => {
                cleanup();
                (window as any).Alpine.store('modal').close('prompt-dialog');
                resolve(null);
            };
            
            const handleKeydown = (e: KeyboardEvent) => {
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
            
            const handleModalClosed = (e: any) => {
                if (e.detail.modalId === 'prompt-dialog') {
                    cleanup();
                    resolve(null);
                }
            };
            
            if (okBtn) okBtn.addEventListener('click', handleOk);
            if (cancelBtn) cancelBtn.addEventListener('click', handleCancel);
            if (inputEl) inputEl.addEventListener('keydown', handleKeydown);
            window.addEventListener('modal:closed', handleModalClosed);
            
            (window as any).Alpine.store('modal').open('prompt-dialog', { backdrop: 'static' });
            
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
(window as any).useModal = useModal;
(window as any).useFormModal = useFormModal;
(window as any).ModalConfigs = ModalConfigs;
(window as any).createConsoleModal = createConsoleModal;
(window as any).createWorkspaceModal = createWorkspaceModal;
(window as any).createPortalModal = createPortalModal;
(window as any).createFullscreenModal = createFullscreenModal;
(window as any).createSidebarModal = createSidebarModal;
(window as any).confirmDialog = confirmDialog;
(window as any).alertDialog = alertDialog;
(window as any).promptDialog = promptDialog;