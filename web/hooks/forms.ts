/**
 * Form Validation AlpineJS Hooks (TypeScript)
 * 
 * Extracted from internal/ui/components/forms/*.templ
 * Provides reusable form validation and interaction patterns
 * 
 * @version 2.0.0
 */

/// <reference path="./types.d.ts" />

// Types
interface ValidationRule {
    required?: boolean;
    requiredMessage?: string;
    minLength?: number;
    minLengthMessage?: string;
    maxLength?: number;
    maxLengthMessage?: string;
    pattern?: string | RegExp;
    patternMessage?: string;
    email?: boolean;
    emailMessage?: string;
    custom?: (value: any, formData: Record<string, any>) => string | null;
    async?: (value: any, formData: Record<string, any>, signal: AbortSignal) => Promise<{ error?: string } | null>;
}

interface FormValidationConfig {
    initialData?: Record<string, any>;
    rules?: Record<string, ValidationRule>;
    realTimeValidation?: boolean;
    validateOnBlur?: boolean;
    validateOnInput?: boolean;
    debounce?: number;
    csrfTokenSelector?: string;
    toastStore?: string;
}

interface RequestOptions {
    method?: string;
    data?: FormData;
    headers?: Record<string, string>;
}

interface SubmissionOptions {
    method?: string;
    headers?: Record<string, string>;
    transformData?: (data: FormData) => FormData;
}

interface FormValidationStore {
    // State
    formData: Record<string, any>;
    validationErrors: Record<string, string[]>;
    validationRules: Record<string, ValidationRule>;
    
    // Status tracking
    isValid: boolean;
    isDirty: boolean;
    isSubmitting: boolean;
    submitAttempted: boolean;
    
    // Configuration
    realTimeValidation: boolean;
    validateOnBlur: boolean;
    validateOnInput: boolean;
    validationDebounce: number;
    csrfTokenSelector: string;
    toastStoreName: string;
    
    // Internal state
    _validationTimeout: any;
    _asyncControllers: Map<string, AbortController>;
    _destroyed: boolean;
    
    // Methods
    init(): void;
    destroy(): void;
    _handleFormDataChange(): void;
    validateForm(): boolean;
    validateField(field: string, value: any, rules: ValidationRule): string[];
    _performAsyncValidation(field: string, value: any, asyncValidator: ValidationRule['async']): Promise<void>;
    _abortAsyncValidation(field: string): void;
    _abortAllAsyncValidations(): void;
    _clearValidationTimeout(): void;
    _debouncedValidation(): void;
    submitForm(endpoint: string, options?: SubmissionOptions): Promise<any>;
    prepareFormData(): FormData;
    _appendFormDataValue(formData: FormData, key: string, value: any): void;
    _onSubmitSuccess(response: any): void;
    _onSubmitError(response: any): void;
    handleFieldInput(field: string, value: any): void;
    handleFieldBlur(field: string): void;
    validateSingleField(field: string): void;
    isEmpty(value: any): boolean;
    isValidEmail(email: string): boolean;
    getFieldLabel(field: string): string;
    getNestedValue(obj: any, path: string): any;
    setNestedValue(obj: any, path: string, value: any): void;
    focusFirstError(): void;
    markDirty(): void;
    reset(newInitialData?: Record<string, any>): void;
    _makeRequest(url: string, options: RequestOptions): Promise<any>;
    getFieldErrors(field: string): string[];
    hasFieldError(field: string): boolean;
    getFirstFieldError(field: string): string | null;
    
    // Computed properties
    readonly hasErrors: boolean;
    readonly canSubmit: boolean;
    readonly fieldErrors: (field: string) => string[];
}

interface PasswordStrengthConfig {
    fieldName?: string;
    minLength?: number;
}

interface PasswordStrengthStore {
    password: string;
    strength: number;
    feedback: string;
    
    init(): void;
    calculateStrength(): void;
    readonly strengthClass: string;
    readonly strengthText: string;
    readonly strengthPercentage: number;
    readonly isStrongEnough: boolean;
}

interface ProgressiveDisclosureConfig {
    sections?: Record<string, boolean>;
    initialSection?: number;
}

interface ProgressiveDisclosureStore {
    sections: Record<string, boolean>;
    currentSection: number;
    
    showSection(sectionKey: string): void;
    hideSection(sectionKey: string): void;
    toggleSection(sectionKey: string): void;
    isSectionVisible(sectionKey: string): boolean;
    showAllSections(): void;
    hideAllSections(): void;
    nextSection(): boolean;
    prevSection(): boolean;
    goToSection(section: number): boolean;
    readonly maxSections: number;
    readonly progress: number;
    readonly isFirstSection: boolean;
    readonly isLastSection: boolean;
}

interface MultiStepFormConfig extends FormValidationConfig {
    stepValidation?: Record<number, Record<string, ValidationRule>>;
    allowStepSkipping?: boolean;
}

interface MultiStepFormStore extends FormValidationStore, ProgressiveDisclosureStore {
    stepValidation: Record<number, Record<string, ValidationRule>>;
    allowStepSkipping: boolean;
    completedSteps: Set<number>;
    _stepValidationCache: Map<number, boolean>;
    
    goToStep(step: number, skipValidation?: boolean): Promise<boolean>;
    nextStep(): Promise<boolean>;
    prevStep(): Promise<boolean>;
    validateCurrentStep(): Promise<boolean>;
    validateStep(step: number): Promise<boolean>;
    validateAllSteps(): Promise<boolean>;
    markStepCompleted(step: number): void;
    markStepIncomplete(step: number): void;
    isStepCompleted(step: number): boolean;
    isStepAccessible(step: number): boolean;
    getStepErrors(step: number): Record<string, string[]>;
    hasStepErrors(step: number): boolean;
    submitMultiStepForm(endpoint: string, options?: SubmissionOptions): Promise<any>;
    resetMultiStepForm(newInitialData?: Record<string, any>): void;
    _updateStepValidationRules(): void;
    _shouldValidateStep(step: number): boolean;
    
    readonly canGoToNextStep: boolean;
    readonly canGoToPrevStep: boolean;
    readonly isCurrentStepValid: boolean;
    readonly completionPercentage: number;
    readonly isFormComplete: boolean;
    readonly currentStepNumber: number;
    readonly totalSteps: number;
}

// Export to make this a module
export {};

/**
 * Form Validation Hook
 * 
 * Provides comprehensive form validation with support for sync/async validation,
 * real-time feedback, and customizable validation rules.
 * 
 * @param config - Configuration options
 * @returns Alpine.js component data
 * 
 * @example
 * <div x-data="useFormValidation({
 *   initialData: { email: '', password: '' },
 *   rules: {
 *     email: { required: true, email: true },
 *     password: { required: true, minLength: 8 }
 *   }
 * })">
 */
function useFormValidation(config: FormValidationConfig = {}): FormValidationStore {
    return {
        // State
        formData: { ...(config.initialData || {}) },
        validationErrors: {},
        validationRules: config.rules || {},
        
        // Status tracking
        isValid: false,
        isDirty: false,
        isSubmitting: false,
        submitAttempted: false,
        
        // Configuration
        realTimeValidation: config.realTimeValidation !== false,
        validateOnBlur: config.validateOnBlur !== false,
        validateOnInput: config.validateOnInput || false,
        validationDebounce: config.debounce || 300,
        csrfTokenSelector: config.csrfTokenSelector || 'meta[name="csrf-token"]',
        toastStoreName: config.toastStore || 'toast',
        
        // Internal state
        _validationTimeout: null,
        _asyncControllers: new Map(),
        _destroyed: false,
        
        /**
         * Initialize the form validation component
         */
        init(): void {
            this.validateForm();
            
            // Watch for form changes with optimized deep watching
            if (typeof (this as any).$watch === 'function') {
                (this as any).$watch('formData', () => {
                    this._handleFormDataChange();
                }, { deep: true });
            }
        },
        
        /**
         * Cleanup on component destroy
         */
        destroy(): void {
            this._destroyed = true;
            this._clearValidationTimeout();
            this._abortAllAsyncValidations();
        },
        
        /**
         * Handle form data changes
         * @private
         */
        _handleFormDataChange(): void {
            if (this._destroyed) return;
            
            this.markDirty();
            
            if (this.realTimeValidation) {
                this._debouncedValidation();
            }
        },
        
        /**
         * Validate the entire form
         * @returns True if form is valid
         */
        validateForm(): boolean {
            this.validationErrors = {};
            let isFormValid = true;
            
            Object.entries(this.validationRules).forEach(([field, rules]) => {
                const value = this.getNestedValue(this.formData, field);
                const fieldErrors = this.validateField(field, value, rules);
                
                if (fieldErrors.length > 0) {
                    this.validationErrors[field] = fieldErrors;
                    isFormValid = false;
                }
            });
            
            this.isValid = isFormValid;
            return isFormValid;
        },
        
        /**
         * Validate a single field
         * @param field - Field name
         * @param value - Field value
         * @param rules - Validation rules
         * @returns Array of error messages
         */
        validateField(field: string, value: any, rules: ValidationRule): string[] {
            const errors: string[] = [];
            
            // Required validation
            if (rules.required && this.isEmpty(value)) {
                errors.push(rules.requiredMessage || `${this.getFieldLabel(field)} is required`);
                return errors; // Early return for required field
            }
            
            // Skip other validations if field is empty and not required
            if (this.isEmpty(value)) {
                return errors;
            }
            
            const stringValue = String(value);
            
            // Min length validation
            if (rules.minLength !== undefined && stringValue.length < rules.minLength) {
                errors.push(
                    rules.minLengthMessage || 
                    `${this.getFieldLabel(field)} must be at least ${rules.minLength} characters`
                );
            }
            
            // Max length validation
            if (rules.maxLength !== undefined && stringValue.length > rules.maxLength) {
                errors.push(
                    rules.maxLengthMessage || 
                    `${this.getFieldLabel(field)} cannot exceed ${rules.maxLength} characters`
                );
            }
            
            // Pattern validation
            if (rules.pattern) {
                const pattern = typeof rules.pattern === 'string' ? new RegExp(rules.pattern) : rules.pattern;
                if (!pattern.test(stringValue)) {
                    errors.push(rules.patternMessage || `${this.getFieldLabel(field)} format is invalid`);
                }
            }
            
            // Email validation
            if (rules.email && !this.isValidEmail(stringValue)) {
                errors.push(rules.emailMessage || `${this.getFieldLabel(field)} must be a valid email address`);
            }
            
            // Custom synchronous validation
            if (rules.custom && typeof rules.custom === 'function') {
                try {
                    const customError = rules.custom(value, this.formData);
                    if (customError) {
                        errors.push(customError);
                    }
                } catch (error) {
                    console.error(`Custom validation error for field "${field}":`, error);
                    errors.push('Validation error occurred');
                }
            }
            
            // Async validation (non-blocking)
            if (rules.async && typeof rules.async === 'function') {
                this._performAsyncValidation(field, value, rules.async);
            }
            
            return errors;
        },
        
        /**
         * Perform async validation with request cancellation
         * @param field - Field name
         * @param value - Field value
         * @param asyncValidator - Async validation function
         * @private
         */
        async _performAsyncValidation(field: string, value: any, asyncValidator: ValidationRule['async']): Promise<void> {
            if (!asyncValidator) return;
            
            // Cancel previous async validation for this field
            this._abortAsyncValidation(field);
            
            // Create new abort controller
            const controller = new AbortController();
            this._asyncControllers.set(field, controller);
            
            try {
                const result = await asyncValidator(value, this.formData, controller.signal);
                
                // Check if this validation was aborted
                if (controller.signal.aborted) return;
                
                // Update validation errors based on result
                if (result && result.error) {
                    this.validationErrors[field] = [result.error];
                    this.isValid = false;
                } else {
                    // Remove async errors if validation passes
                    if (this.validationErrors[field]) {
                        delete this.validationErrors[field];
                    }
                    this.validateForm(); // Recheck overall form validity
                }
            } catch (error: any) {
                if (error.name === 'AbortError') {
                    // Validation was cancelled, ignore
                    return;
                }
                console.error(`Async validation failed for field "${field}":`, error);
            } finally {
                this._asyncControllers.delete(field);
            }
        },
        
        /**
         * Abort async validation for a specific field
         * @param field - Field name
         * @private
         */
        _abortAsyncValidation(field: string): void {
            const controller = this._asyncControllers.get(field);
            if (controller) {
                controller.abort();
                this._asyncControllers.delete(field);
            }
        },
        
        /**
         * Abort all pending async validations
         * @private
         */
        _abortAllAsyncValidations(): void {
            this._asyncControllers.forEach(controller => controller.abort());
            this._asyncControllers.clear();
        },
        
        /**
         * Clear validation timeout
         * @private
         */
        _clearValidationTimeout(): void {
            if (this._validationTimeout) {
                clearTimeout(this._validationTimeout);
                this._validationTimeout = null;
            }
        },
        
        /**
         * Debounced validation for real-time feedback
         * @private
         */
        _debouncedValidation(): void {
            this._clearValidationTimeout();
            this._validationTimeout = setTimeout(() => {
                if (!this._destroyed) {
                    this.validateForm();
                }
            }, this.validationDebounce);
        },
        
        /**
         * Submit the form
         * @param endpoint - Submission endpoint URL
         * @param options - Submission options
         * @returns Response object
         */
        async submitForm(endpoint: string, options: SubmissionOptions = {}): Promise<any> {
            this.submitAttempted = true;
            
            if (this.isSubmitting) {
                return { success: false, error: 'Form is already submitting' };
            }
            
            this.isSubmitting = true;
            
            try {
                // Final validation
                if (!this.validateForm()) {
                    this.focusFirstError();
                    return { success: false, errors: this.validationErrors };
                }
                
                // Prepare form data
                let formData = this.prepareFormData();
                
                // Apply transformation if provided
                if (options.transformData && typeof options.transformData === 'function') {
                    formData = options.transformData(formData);
                }
                
                // Submit to server
                const response = await this._makeRequest(endpoint, {
                    method: options.method || 'POST',
                    data: formData,
                    headers: options.headers || {}
                });
                
                if (response.success) {
                    this._onSubmitSuccess(response);
                } else {
                    this._onSubmitError(response);
                }
                
                return response;
                
            } catch (error: any) {
                console.error('Form submission failed:', error);
                const errorResponse = { 
                    success: false, 
                    error: error.message || 'Network error occurred'
                };
                this._onSubmitError(errorResponse);
                return errorResponse;
                
            } finally {
                this.isSubmitting = false;
            }
        },
        
        /**
         * Prepare form data for submission
         * @returns Prepared form data
         */
        prepareFormData(): FormData {
            const formData = new FormData();
            
            // Add all form fields
            Object.entries(this.formData).forEach(([key, value]) => {
                this._appendFormDataValue(formData, key, value);
            });
            
            // Add CSRF token
            const csrfToken = document.querySelector(this.csrfTokenSelector)?.getAttribute('content');
            if (csrfToken) {
                formData.append('_token', csrfToken);
            }
            
            return formData;
        },
        
        /**
         * Append value to FormData with type handling
         * @param formData - FormData object
         * @param key - Field key
         * @param value - Field value
         * @private
         */
        _appendFormDataValue(formData: FormData, key: string, value: any): void {
            if (value === null || value === undefined) {
                return;
            }
            
            if (value instanceof FileList || value instanceof File) {
                const files = value instanceof FileList ? Array.from(value) : [value];
                files.forEach(file => formData.append(key, file));
            } else if (Array.isArray(value)) {
                value.forEach(v => formData.append(`${key}[]`, v));
            } else if (typeof value === 'object' && !(value instanceof Date)) {
                // Serialize objects as JSON
                formData.append(key, JSON.stringify(value));
            } else {
                formData.append(key, value);
            }
        },
        
        /**
         * Handle successful form submission
         * @param response - Server response
         * @private
         */
        _onSubmitSuccess(response: any): void {
            this.isDirty = false;
            this.validationErrors = {};
            
            // Show success message
            const toastStore = (window as any).Alpine?.store?.(this.toastStoreName);
            if (response.message && toastStore?.success) {
                toastStore.success(response.message);
            }
            
            // Emit success event
            if (typeof (this as any).$dispatch === 'function') {
                (this as any).$dispatch('form-submit-success', { 
                    response, 
                    formData: { ...this.formData } 
                });
            }
        },
        
        /**
         * Handle form submission error
         * @param response - Error response
         * @private
         */
        _onSubmitError(response: any): void {
            if (response.errors) {
                this.validationErrors = response.errors;
                this.focusFirstError();
            }
            
            // Show error message
            const message = response.message || 'Form submission failed. Please check the errors and try again.';
            const toastStore = (window as any).Alpine?.store?.(this.toastStoreName);
            if (toastStore?.error) {
                toastStore.error(message);
            }
            
            // Emit error event
            if (typeof (this as any).$dispatch === 'function') {
                (this as any).$dispatch('form-submit-error', { 
                    response, 
                    formData: { ...this.formData } 
                });
            }
        },
        
        /**
         * Handle field input event
         * @param field - Field name
         * @param value - New value
         */
        handleFieldInput(field: string, value: any): void {
            this.setNestedValue(this.formData, field, value);
            
            if (this.validateOnInput || this.submitAttempted) {
                this.validateSingleField(field);
            }
        },
        
        /**
         * Handle field blur event
         * @param field - Field name
         */
        handleFieldBlur(field: string): void {
            if (this.validateOnBlur || this.submitAttempted) {
                this.validateSingleField(field);
            }
        },
        
        /**
         * Validate a single field by name
         * @param field - Field name
         */
        validateSingleField(field: string): void {
            const rules = this.validationRules[field];
            if (!rules) {
                console.warn(`No validation rules found for field: ${field}`);
                return;
            }
            
            const value = this.getNestedValue(this.formData, field);
            const errors = this.validateField(field, value, rules);
            
            if (errors.length > 0) {
                this.validationErrors[field] = errors;
                this.isValid = false;
            } else {
                delete this.validationErrors[field];
                // Revalidate to update isValid status
                this.isValid = Object.keys(this.validationErrors).length === 0;
            }
        },
        
        /**
         * Check if a value is empty
         * @param value - Value to check
         * @returns True if empty
         */
        isEmpty(value: any): boolean {
            if (value === null || value === undefined || value === '') {
                return true;
            }
            
            if (Array.isArray(value)) {
                return value.length === 0;
            }
            
            if (typeof value === 'object') {
                return Object.keys(value).length === 0;
            }
            
            return false;
        },
        
        /**
         * Validate email format
         * @param email - Email to validate
         * @returns True if valid email
         */
        isValidEmail(email: string): boolean {
            // More comprehensive email regex
            const emailRegex = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;
            return emailRegex.test(email);
        },
        
        /**
         * Convert field name to human-readable label
         * @param field - Field name
         * @returns Human-readable label
         */
        getFieldLabel(field: string): string {
            // Handle dot notation
            const fieldName = field.split('.').pop() || field;
            
            return fieldName
                .split('_')
                .map(word => word.charAt(0).toUpperCase() + word.slice(1))
                .join(' ');
        },
        
        /**
         * Get nested object value by path
         * @param obj - Source object
         * @param path - Dot-notation path
         * @returns Value at path
         */
        getNestedValue(obj: any, path: string): any {
            if (!path || !obj) return undefined;
            return path.split('.').reduce((current, key) => current?.[key], obj);
        },
        
        /**
         * Set nested object value by path
         * @param obj - Target object
         * @param path - Dot-notation path
         * @param value - Value to set
         */
        setNestedValue(obj: any, path: string, value: any): void {
            if (!path || !obj) return;
            
            const keys = path.split('.');
            const lastKey = keys.pop();
            
            if (!lastKey) return;
            
            const target = keys.reduce((current, key) => {
                if (!(key in current) || typeof current[key] !== 'object') {
                    current[key] = {};
                }
                return current[key];
            }, obj);
            
            target[lastKey] = value;
        },
        
        /**
         * Focus the first field with an error
         */
        focusFirstError(): void {
            const firstErrorField = Object.keys(this.validationErrors)[0];
            if (!firstErrorField) return;
            
            // Try multiple selectors
            const selectors = [
                `[name="${firstErrorField}"]`,
                `[data-field="${firstErrorField}"]`,
                `#${firstErrorField}`,
                `[id$="-${firstErrorField}"]`
            ];
            
            for (const selector of selectors) {
                const element = document.querySelector(selector) as HTMLElement;
                if (element) {
                    element.focus();
                    element.scrollIntoView({ 
                        behavior: 'smooth', 
                        block: 'center',
                        inline: 'nearest'
                    });
                    return;
                }
            }
            
            console.warn(`Could not find element for field: ${firstErrorField}`);
        },
        
        /**
         * Mark form as dirty (modified)
         */
        markDirty(): void {
            this.isDirty = true;
        },
        
        /**
         * Reset form to initial state
         * @param newInitialData - Optional new initial data
         */
        reset(newInitialData?: Record<string, any>): void {
            const initialData = newInitialData || config.initialData || {};
            this.formData = { ...initialData };
            this.validationErrors = {};
            this.isDirty = false;
            this.submitAttempted = false;
            this.isValid = false;
            this._abortAllAsyncValidations();
        },
        
        /**
         * Make HTTP request
         * @param url - Request URL
         * @param options - Request options
         * @returns Response data
         * @private
         */
        async _makeRequest(url: string, options: RequestOptions): Promise<any> {
            const requestOptions: RequestInit = {
                method: options.method || 'POST',
                body: options.data,
                headers: {
                    'X-Requested-With': 'XMLHttpRequest',
                    ...(options.headers || {})
                }
            };
            
            const response = await fetch(url, requestOptions);
            
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
         * Get all validation errors for a field
         * @param field - Field name
         * @returns Array of error messages
         */
        getFieldErrors(field: string): string[] {
            return this.validationErrors[field] || [];
        },
        
        /**
         * Check if a field has errors
         * @param field - Field name
         * @returns True if field has errors
         */
        hasFieldError(field: string): boolean {
            return !!(this.validationErrors[field] && this.validationErrors[field].length > 0);
        },
        
        /**
         * Get the first error for a field
         * @param field - Field name
         * @returns First error message or null
         */
        getFirstFieldError(field: string): string | null {
            const errors = this.getFieldErrors(field);
            return errors.length > 0 ? errors[0] : null;
        },
        
        // Computed properties (getters for backward compatibility)
        get hasErrors(): boolean {
            return Object.keys(this.validationErrors).length > 0;
        },
        
        get canSubmit(): boolean {
            return this.isValid && !this.isSubmitting;
        },
        
        // Backward compatibility aliases
        get fieldErrors() {
            return (field: string) => this.getFieldErrors(field);
        }
    };
}

/**
 * Password Strength Hook
 * 
 * Calculates password strength and provides feedback
 * 
 * @param config - Configuration options
 * @returns Alpine.js component data
 * 
 * @example
 * <div x-data="usePasswordStrength()">
 */
function usePasswordStrength(config: PasswordStrengthConfig = {}): PasswordStrengthStore {
    const minLength = config.minLength || 8;
    
    return {
        password: '',
        strength: 0,
        feedback: '',
        
        init(): void {
            if (typeof (this as any).$watch === 'function') {
                (this as any).$watch('password', () => {
                    this.calculateStrength();
                });
            }
        },
        
        /**
         * Calculate password strength score
         */
        calculateStrength(): void {
            let score = 0;
            const feedback: string[] = [];
            
            if (this.password.length === 0) {
                this.strength = 0;
                this.feedback = '';
                return;
            }
            
            // Length check
            if (this.password.length >= minLength) {
                score += 1;
            } else {
                feedback.push(`Use at least ${minLength} characters`);
            }
            
            // Uppercase check
            if (/[A-Z]/.test(this.password)) {
                score += 1;
            } else {
                feedback.push('Add uppercase letters');
            }
            
            // Lowercase check
            if (/[a-z]/.test(this.password)) {
                score += 1;
            } else {
                feedback.push('Add lowercase letters');
            }
            
            // Number check
            if (/[0-9]/.test(this.password)) {
                score += 1;
            } else {
                feedback.push('Add numbers');
            }
            
            // Special character check
            if (/[^A-Za-z0-9]/.test(this.password)) {
                score += 1;
            } else {
                feedback.push('Add special characters');
            }
            
            // Bonus for very long passwords
            if (this.password.length >= 12) {
                score = Math.min(score + 1, 5);
            }
            
            this.strength = score;
            this.feedback = feedback.join(', ');
        },
        
        /**
         * Get CSS class for strength indicator
         * @returns CSS class name
         */
        get strengthClass(): string {
            const classes = ['weak', 'weak', 'fair', 'good', 'strong', 'very-strong'];
            return classes[Math.min(this.strength, classes.length - 1)] || 'weak';
        },
        
        /**
         * Get human-readable strength text
         * @returns Strength description
         */
        get strengthText(): string {
            const texts = ['Very Weak', 'Weak', 'Fair', 'Good', 'Strong', 'Very Strong'];
            return texts[Math.min(this.strength, texts.length - 1)] || 'Very Weak';
        },
        
        /**
         * Get strength percentage (0-100)
         * @returns Strength percentage
         */
        get strengthPercentage(): number {
            return Math.min((this.strength / 5) * 100, 100);
        },
        
        /**
         * Check if password meets minimum requirements
         * @returns True if password is strong enough
         */
        get isStrongEnough(): boolean {
            return this.strength >= 3; // At least "Good"
        }
    };
}

/**
 * Progressive Disclosure Hook
 * 
 * Manages progressive disclosure of form sections or content
 * 
 * @param config - Configuration options
 * @returns Alpine.js component data
 * 
 * @example
 * <div x-data="useProgressiveDisclosure({ sections: { basic: true, advanced: false } })">
 */
function useProgressiveDisclosure(config: ProgressiveDisclosureConfig = {}): ProgressiveDisclosureStore {
    return {
        sections: { ...(config.sections || {}) },
        currentSection: config.initialSection || 0,
        
        /**
         * Show a section
         * @param sectionKey - Section identifier
         */
        showSection(sectionKey: string): void {
            this.sections[sectionKey] = true;
        },
        
        /**
         * Hide a section
         * @param sectionKey - Section identifier
         */
        hideSection(sectionKey: string): void {
            this.sections[sectionKey] = false;
        },
        
        /**
         * Toggle section visibility
         * @param sectionKey - Section identifier
         */
        toggleSection(sectionKey: string): void {
            this.sections[sectionKey] = !this.sections[sectionKey];
        },
        
        /**
         * Check if a section is visible
         * @param sectionKey - Section identifier
         * @returns True if section is visible
         */
        isSectionVisible(sectionKey: string): boolean {
            return !!this.sections[sectionKey];
        },
        
        /**
         * Show all sections
         */
        showAllSections(): void {
            Object.keys(this.sections).forEach(key => {
                this.sections[key] = true;
            });
        },
        
        /**
         * Hide all sections
         */
        hideAllSections(): void {
            Object.keys(this.sections).forEach(key => {
                this.sections[key] = false;
            });
        },
        
        /**
         * Move to next section
         * @returns True if moved successfully
         */
        nextSection(): boolean {
            if (this.currentSection < this.maxSections - 1) {
                this.currentSection++;
                return true;
            }
            return false;
        },
        
        /**
         * Move to previous section
         * @returns True if moved successfully
         */
        prevSection(): boolean {
            if (this.currentSection > 0) {
                this.currentSection--;
                return true;
            }
            return false;
        },
        
        /**
         * Go to specific section
         * @param section - Section index
         * @returns True if moved successfully
         */
        goToSection(section: number): boolean {
            if (section >= 0 && section < this.maxSections) {
                this.currentSection = section;
                return true;
            }
            return false;
        },
        
        /**
         * Get total number of sections
         * @returns Section count
         */
        get maxSections(): number {
            return Object.keys(this.sections).length;
        },
        
        /**
         * Get progress percentage
         * @returns Progress (0-100)
         */
        get progress(): number {
            if (this.maxSections === 0) return 0;
            return ((this.currentSection + 1) / this.maxSections) * 100;
        },
        
        /**
         * Check if on first section
         * @returns True if on first section
         */
        get isFirstSection(): boolean {
            return this.currentSection === 0;
        },
        
        /**
         * Check if on last section
         * @returns True if on last section
         */
        get isLastSection(): boolean {
            return this.currentSection === this.maxSections - 1;
        }
    };
}

/**
 * Multi-step Form Hook
 * 
 * Combines form validation with step-by-step navigation
 * 
 * @param config - Configuration options
 * @returns Alpine.js component data
 * 
 * @example
 * <div x-data="useMultiStepForm({
 *   stepValidation: {
 *     0: { email: { required: true, email: true } },
 *     1: { password: { required: true, minLength: 8 } }
 *   }
 * })">
 */
function useMultiStepForm(config: MultiStepFormConfig = {}): MultiStepFormStore {
    // Create base components without method conflicts
    const formValidation = useFormValidation(config);
    const progressiveDisclosure = useProgressiveDisclosure(config as any);
    
    // Extract methods that might conflict
    const { init: formInit, destroy: formDestroy, ...formRest } = formValidation;
    const { 
        nextSection, 
        prevSection, 
        goToSection,
        ...disclosureRest 
    } = progressiveDisclosure;
    
    return {
        // Merge base functionality
        ...formRest,
        ...disclosureRest,
        
        // Add missing methods from progressive disclosure
        nextSection: nextSection,
        prevSection: prevSection,
        
        // Multi-step specific configuration
        stepValidation: config.stepValidation || {},
        allowStepSkipping: config.allowStepSkipping || false,
        
        // Multi-step specific state
        completedSteps: new Set<number>(),
        _stepValidationCache: new Map<number, boolean>(),
        
        /**
         * Initialize multi-step form
         */
        init(): void {
            // Call both parent inits
            if (formInit) formInit.call(this);
            // Initialize progressive disclosure if available
            if ((progressiveDisclosure as any).init) {
                (progressiveDisclosure as any).init.call(this);
            }
            
            // Mark first step as current
            this._updateStepValidationRules();
        },
        
        /**
         * Cleanup on destroy
         */
        destroy(): void {
            if (formDestroy) formDestroy.call(this);
            this._stepValidationCache.clear();
        },
        
        /**
         * Update validation rules for current step
         * @private
         */
        _updateStepValidationRules(): void {
            const stepRules = this.stepValidation[this.currentSection];
            if (stepRules) {
                // Merge step rules with global rules
                this.validationRules = {
                    ...config.rules,
                    ...stepRules
                };
            }
        },
        
        /**
         * Navigate to a specific step
         * @param step - Target step index
         * @param skipValidation - Skip validation
         * @returns True if navigation successful
         */
        async goToStep(step: number, skipValidation: boolean = false): Promise<boolean> {
            // Validate bounds
            if (step < 0 || step >= this.maxSections) {
                console.warn(`Invalid step index: ${step}`);
                return false;
            }
            
            // Skip validation if explicitly allowed or going backwards
            const shouldValidate = !skipValidation && 
                                  !this.allowStepSkipping && 
                                  step > this.currentSection;
            
            if (shouldValidate && this._shouldValidateStep(this.currentSection)) {
                const isValid = await this.validateCurrentStep();
                if (!isValid) {
                    this.focusFirstError();
                    return false;
                }
            }
            
            // Mark current step as completed if moving forward
            if (step > this.currentSection) {
                this.markStepCompleted(this.currentSection);
            }
            
            // Navigate to new step
            this.currentSection = step;
            this._updateStepValidationRules();
            
            // Emit step change event
            if (typeof (this as any).$dispatch === 'function') {
                (this as any).$dispatch('step-changed', { 
                    step, 
                    previousStep: this.currentSection,
                    isCompleted: this.isStepCompleted(step)
                });
            }
            
            return true;
        },
        
        /**
         * Move to next step
         * @returns True if navigation successful
         */
        async nextStep(): Promise<boolean> {
            if (!this.canGoToNextStep) {
                return false;
            }
            return await this.goToStep(this.currentSection + 1);
        },
        
        /**
         * Move to previous step
         * @returns True if navigation successful
         */
        async prevStep(): Promise<boolean> {
            if (!this.canGoToPrevStep) {
                return false;
            }
            // No validation when going back
            return await this.goToStep(this.currentSection - 1, true);
        },
        
        /**
         * Validate current step
         * @returns True if step is valid
         */
        async validateCurrentStep(): Promise<boolean> {
            return this.validateStep(this.currentSection);
        },
        
        /**
         * Validate a specific step
         * @param step - Step index to validate
         * @returns True if step is valid
         */
        async validateStep(step: number): Promise<boolean> {
            const stepRules = this.stepValidation[step];
            if (!stepRules) {
                return true; // No validation rules for this step
            }
            
            // Validate only fields in this step
            const stepErrors: Record<string, string[]> = {};
            let isStepValid = true;
            
            for (const [field, rules] of Object.entries(stepRules)) {
                const value = this.getNestedValue(this.formData, field);
                const fieldErrors = this.validateField(field, value, rules);
                
                if (fieldErrors.length > 0) {
                    stepErrors[field] = fieldErrors;
                    isStepValid = false;
                }
            }
            
            // Update validation errors
            // Clear old step errors, add new ones
            Object.keys(stepRules).forEach(field => {
                delete this.validationErrors[field];
            });
            
            Object.assign(this.validationErrors, stepErrors);
            
            // Update overall validity
            this.isValid = Object.keys(this.validationErrors).length === 0;
            
            return isStepValid;
        },
        
        /**
         * Validate all steps
         * @returns True if all steps are valid
         */
        async validateAllSteps(): Promise<boolean> {
            let allValid = true;
            
            for (let step = 0; step < this.maxSections; step++) {
                const isValid = await this.validateStep(step);
                if (!isValid) {
                    allValid = false;
                }
            }
            
            return allValid;
        },
        
        /**
         * Check if step should be validated
         * @param step - Step index
         * @returns True if step has validation rules
         * @private
         */
        _shouldValidateStep(step: number): boolean {
            return !!this.stepValidation[step] && 
                   Object.keys(this.stepValidation[step]).length > 0;
        },
        
        /**
         * Mark step as completed
         * @param step - Step index
         */
        markStepCompleted(step: number): void {
            this.completedSteps.add(step);
            
            if (typeof (this as any).$dispatch === 'function') {
                (this as any).$dispatch('step-completed', { 
                    step,
                    completedSteps: Array.from(this.completedSteps)
                });
            }
        },
        
        /**
         * Mark step as incomplete
         * @param step - Step index
         */
        markStepIncomplete(step: number): void {
            this.completedSteps.delete(step);
        },
        
        /**
         * Check if step is completed
         * @param step - Step index
         * @returns True if step is completed
         */
        isStepCompleted(step: number): boolean {
            return this.completedSteps.has(step);
        },
        
        /**
         * Check if step is accessible (all previous steps completed or skipping allowed)
         * @param step - Step index
         * @returns True if step can be accessed
         */
        isStepAccessible(step: number): boolean {
            if (this.allowStepSkipping) {
                return true;
            }
            
            // Check if all previous steps are completed
            for (let i = 0; i < step; i++) {
                if (!this.isStepCompleted(i)) {
                    return false;
                }
            }
            
            return true;
        },
        
        /**
         * Get step validation errors
         * @param step - Step index
         * @returns Errors by field
         */
        getStepErrors(step: number): Record<string, string[]> {
            const stepRules = this.stepValidation[step];
            if (!stepRules) return {};
            
            const stepFields = Object.keys(stepRules);
            const stepErrors: Record<string, string[]> = {};
            
            stepFields.forEach(field => {
                if (this.validationErrors[field]) {
                    stepErrors[field] = this.validationErrors[field];
                }
            });
            
            return stepErrors;
        },
        
        /**
         * Check if step has errors
         * @param step - Step index
         * @returns True if step has errors
         */
        hasStepErrors(step: number): boolean {
            const stepErrors = this.getStepErrors(step);
            return Object.keys(stepErrors).length > 0;
        },
        
        /**
         * Submit multi-step form
         * @param endpoint - Submission endpoint
         * @param options - Submission options
         * @returns Response object
         */
        async submitMultiStepForm(endpoint: string, options: SubmissionOptions = {}): Promise<any> {
            // Validate all steps before submission
            const allValid = await this.validateAllSteps();
            
            if (!allValid) {
                // Navigate to first step with errors
                for (let step = 0; step < this.maxSections; step++) {
                    if (this.hasStepErrors(step)) {
                        await this.goToStep(step, true);
                        this.focusFirstError();
                        break;
                    }
                }
                
                return { 
                    success: false, 
                    errors: this.validationErrors,
                    message: 'Please complete all steps correctly'
                };
            }
            
            // Use parent submit method
            return await this.submitForm(endpoint, options);
        },
        
        /**
         * Reset multi-step form
         * @param newInitialData - Optional new initial data
         */
        resetMultiStepForm(newInitialData?: Record<string, any>): void {
            this.reset(newInitialData);
            this.completedSteps.clear();
            this.currentSection = 0;
            this._updateStepValidationRules();
            this._stepValidationCache.clear();
        },
        
        // Computed properties
        
        /**
         * Check if can go to next step
         * @returns True if next step is available
         */
        get canGoToNextStep(): boolean {
            return this.currentSection < this.maxSections - 1;
        },
        
        /**
         * Check if can go to previous step
         * @returns True if previous step is available
         */
        get canGoToPrevStep(): boolean {
            return this.currentSection > 0;
        },
        
        /**
         * Check if current step is valid
         * @returns True if current step has no errors
         */
        get isCurrentStepValid(): boolean {
            return !this.hasStepErrors(this.currentSection);
        },
        
        /**
         * Get completion percentage
         * @returns Percentage (0-100)
         */
        get completionPercentage(): number {
            if (this.maxSections === 0) return 0;
            return (this.completedSteps.size / this.maxSections) * 100;
        },
        
        /**
         * Check if form is complete (all steps completed)
         * @returns True if all steps are completed
         */
        get isFormComplete(): boolean {
            return this.completedSteps.size === this.maxSections;
        },
        
        /**
         * Get current step number (1-indexed for display)
         * @returns Current step number
         */
        get currentStepNumber(): number {
            return this.currentSection + 1;
        },
        
        /**
         * Get total step count
         * @returns Total steps
         */
        get totalSteps(): number {
            return this.maxSections;
        },
        
        // Override goToSection to use goToStep
        goToSection(section: number): boolean {
            this.goToStep(section);
            return true;
        }
    };
}

/**
 * Export hooks to window for global access
 */
(window as any).useFormValidation = useFormValidation;
(window as any).usePasswordStrength = usePasswordStrength;
(window as any).useProgressiveDisclosure = useProgressiveDisclosure;
(window as any).useMultiStepForm = useMultiStepForm;