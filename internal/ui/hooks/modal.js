/**
 * Modal AlpineJS Hooks
 * 
 * Extracted from internal/ui/components/layout/modal.templ
 * Provides reusable Modal functionality for all ERP services
 */

/**
 * Modal Store - Global state management for modals
 * Usage: $store.modal.open('modal-id')
 */
document.addEventListener('alpine:init', () => {
    Alpine.store('modal', {
        // State
        isOpen: false,
        currentModal: '',
        modalStack: [], // For nested modals
        
        // Configuration
        backdrop: 'static', // 'static' or 'dismissible'
        keyboard: true,     // ESC key support
        focus: true,        // Focus management
        
        // Methods
        open(modalId, config = {}) {
            this.modalStack.push({
                id: this.currentModal,
                isOpen: this.isOpen
            });
            
            this.currentModal = modalId;
            this.isOpen = true;
            this.backdrop = config.backdrop || 'static';
            this.keyboard = config.keyboard !== false;
            this.focus = config.focus !== false;
            
            // Focus management
            if (this.focus) {
                this.setInitialFocus(modalId);
            }
            
            // Prevent body scroll
            document.body.classList.add('modal-open');
            
            // Emit event
            window.dispatchEvent(new CustomEvent('modal:opened', {
                detail: { modalId, config }
            }));
        },
        
        close(modalId = null) {
            if (modalId && modalId !== this.currentModal) {
                return; // Not the current modal
            }
            
            const previousModal = this.modalStack.pop();
            
            if (previousModal && previousModal.id) {
                // Restore previous modal
                this.currentModal = previousModal.id;
                this.isOpen = previousModal.isOpen;
            } else {
                // No previous modal
                this.isOpen = false;
                this.currentModal = '';
                
                // Restore body scroll
                document.body.classList.remove('modal-open');
            }
            
            // Emit event
            window.dispatchEvent(new CustomEvent('modal:closed', {
                detail: { modalId: modalId || this.currentModal }
            }));
        },
        
        toggle(modalId, config = {}) {
            if (this.isOpen && this.currentModal === modalId) {
                this.close(modalId);
            } else {
                this.open(modalId, config);
            }
        },
        
        closeAll() {
            this.modalStack = [];
            this.isOpen = false;
            this.currentModal = '';
            document.body.classList.remove('modal-open');
            
            window.dispatchEvent(new CustomEvent('modal:all-closed'));
        },
        
        // Form submission helper
        async submit(modalId, formData, endpoint, options = {}) {
            try {
                const response = await this.makeRequest(endpoint, {
                    method: options.method || 'POST',
                    data: formData
                });
                
                if (response.success) {
                    this.close(modalId);
                    
                    // Show success message
                    if (response.message) {
                        Alpine.store('toast')?.success?.(response.message);
                    }
                    
                    // Emit success event
                    window.dispatchEvent(new CustomEvent('modal:submit-success', {
                        detail: { modalId, response, formData }
                    }));
                    
                    return response;
                } else {
                    // Handle validation errors
                    this.showValidationErrors(response.errors || {});
                    return response;
                }
                
            } catch (error) {
                console.error('Modal form submission failed:', error);
                Alpine.store('toast')?.error?.('Form submission failed. Please try again.');
                return { success: false, error: error.message };
            }
        },
        
        // Focus management
        setInitialFocus(modalId) {
            requestAnimationFrame(() => {
                const modal = document.getElementById(modalId);
                if (!modal) return;
                
                // Find first focusable element
                const focusableElements = modal.querySelectorAll(
                    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
                );
                
                if (focusableElements.length > 0) {
                    focusableElements[0].focus();
                }
            });
        },
        
        // Utility methods
        async makeRequest(url, options) {
            const formData = new FormData();
            
            if (options.data) {
                Object.entries(options.data).forEach(([key, value]) => {
                    if (value instanceof FileList) {
                        Array.from(value).forEach(file => formData.append(key, file));
                    } else if (Array.isArray(value)) {
                        value.forEach(v => formData.append(`${key}[]`, v));
                    } else {
                        formData.append(key, value);
                    }
                });
            }
            
            const response = await fetch(url, {
                method: options.method || 'POST',
                body: formData,
                headers: {
                    'X-Requested-With': 'XMLHttpRequest',
                    'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.content
                }
            });
            
            return await response.json();
        },
        
        showValidationErrors(errors) {
            // Clear previous errors
            document.querySelectorAll('.validation-error').forEach(el => {
                el.textContent = '';
                el.style.display = 'none';
            });
            
            // Show new errors
            Object.entries(errors).forEach(([field, messages]) => {
                const errorElement = document.querySelector(`[data-validation-for="${field}"]`);
                if (errorElement) {
                    errorElement.textContent = Array.isArray(messages) ? messages[0] : messages;
                    errorElement.style.display = 'block';
                }
            });
        }
    });
});

/**
 * Modal Component Hook
 * Usage: x-data="useModal(config)"
 */
function useModal(config = {}) {
    return {
        // State
        localLoading: false,
        formData: {},
        validationErrors: {},
        
        // Configuration
        id: config.id || '',
        size: config.size || 'md',
        position: config.position || 'center',
        backdrop: config.backdrop || 'static',
        closable: config.closable !== false,
        
        // Lifecycle
        init() {
            // Listen for modal events
            window.addEventListener('modal:opened', (e) => {
                if (e.detail.modalId === this.id) {
                    this.onOpened(e.detail);
                }
            });
            
            window.addEventListener('modal:closed', (e) => {
                if (e.detail.modalId === this.id) {
                    this.onClosed(e.detail);
                }
            });
            
            // Keyboard event handling
            if (this.backdrop !== 'static') {
                this.$el.addEventListener('keydown', (e) => {
                    if (e.key === 'Escape' && this.isCurrentModal) {
                        this.close();
                    }
                });
            }
        },
        
        // Computed properties
        get isCurrentModal() {
            return Alpine.store('modal').currentModal === this.id;
        },
        
        get isOpen() {
            return this.isCurrentModal && Alpine.store('modal').isOpen;
        },
        
        // Methods
        open() {
            Alpine.store('modal').open(this.id, {
                backdrop: this.backdrop,
                size: this.size,
                position: this.position
            });
        },
        
        close() {
            if (this.closable) {
                Alpine.store('modal').close(this.id);
            }
        },
        
        toggle() {
            Alpine.store('modal').toggle(this.id, {
                backdrop: this.backdrop,
                size: this.size,
                position: this.position
            });
        },
        
        // Form handling
        async submitForm(formElement, endpoint, options = {}) {
            if (this.localLoading) return;
            
            this.localLoading = true;
            this.validationErrors = {};
            
            try {
                const formData = new FormData(formElement);
                
                // Add any additional data
                if (this.formData) {
                    Object.entries(this.formData).forEach(([key, value]) => {
                        formData.append(key, value);
                    });
                }
                
                const response = await Alpine.store('modal').submit(
                    this.id,
                    formData,
                    endpoint,
                    options
                );
                
                if (!response.success && response.errors) {
                    this.validationErrors = response.errors;
                }
                
                return response;
                
            } finally {
                this.localLoading = false;
            }
        },
        
        // Event handlers
        onOpened(detail) {
            // Override in specific modals
        },
        
        onClosed(detail) {
            // Reset form state
            this.formData = {};
            this.validationErrors = {};
            this.localLoading = false;
        },
        
        // Backdrop click handling
        handleBackdropClick(event) {
            if (event.target === event.currentTarget && this.backdrop === 'dismissible') {
                this.close();
            }
        }
    };
}

/**
 * Service-specific Modal configurations
 */
const ModalConfigs = {
    console: {
        size: 'lg',
        position: 'center',
        backdrop: 'static',
        closable: true
    },
    
    workspace: {
        size: 'md',
        position: 'center',
        backdrop: 'dismissible',
        closable: true
    },
    
    portal: {
        size: 'sm',
        position: 'center',
        backdrop: 'static',
        closable: true
    }
};

/**
 * Convenience factories for service-specific Modals
 */
window.createConsoleModal = (config = {}) => {
    return useModal({
        ...ModalConfigs.console,
        ...config
    });
};

window.createWorkspaceModal = (config = {}) => {
    return useModal({
        ...ModalConfigs.workspace,
        ...config
    });
};

window.createPortalModal = (config = {}) => {
    return useModal({
        ...ModalConfigs.portal,
        ...config
    });
};

/**
 * Form Modal Hook - Specialized for form modals
 * Usage: x-data="useFormModal(config)"
 */
function useFormModal(config = {}) {
    const modal = useModal(config);
    
    return {
        ...modal,
        
        // Form-specific state
        isDirty: false,
        saveOnClose: config.saveOnClose || false,
        confirmOnClose: config.confirmOnClose || true,
        
        // Enhanced form handling
        markDirty() {
            this.isDirty = true;
        },
        
        async close() {
            if (this.isDirty && this.confirmOnClose) {
                const confirmed = await this.confirmClose();
                if (!confirmed) return;
            }
            
            if (this.isDirty && this.saveOnClose) {
                await this.autoSave();
            }
            
            modal.close.call(this);
        },
        
        async confirmClose() {
            return window.confirm('You have unsaved changes. Are you sure you want to close?');
        },
        
        async autoSave() {
            if (this.saveEndpoint) {
                const formElement = this.$el.querySelector('form');
                if (formElement) {
                    await this.submitForm(formElement, this.saveEndpoint, { method: 'POST' });
                }
            }
        },
        
        onFormChange() {
            this.markDirty();
        }
    };
}

// Export for use in components
window.useModal = useModal;
window.useFormModal = useFormModal;
window.ModalConfigs = ModalConfigs;