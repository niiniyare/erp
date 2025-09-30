/**
 * Form Validation AlpineJS Hooks (Improved)
 * 
 * Extracted from internal/ui/components/forms/*.templ
 * Provides reusable form validation and interaction patterns
 * 
 * @version 2.0.0
 */

/**
 * @typedef {Object} ValidationRule
 * @property {boolean} [required] - Field is required
 * @property {string} [requiredMessage] - Custom required message
 * @property {number} [minLength] - Minimum length
 * @property {string} [minLengthMessage] - Custom min length message
 * @property {number} [maxLength] - Maximum length
 * @property {string} [maxLengthMessage] - Custom max length message
 * @property {string|RegExp} [pattern] - Validation pattern
 * @property {string} [patternMessage] - Custom pattern message
 * @property {boolean} [email] - Email validation
 * @property {string} [emailMessage] - Custom email message
 * @property {Function} [custom] - Custom validation function
 * @property {Function} [async] - Async validation function
 */

/**
 * @typedef {Object} FormValidationConfig
 * @property {Object} [initialData={}] - Initial form data
 * @property {Object.<string, ValidationRule>} [rules={}] - Validation rules
 * @property {boolean} [realTimeValidation=true] - Enable real-time validation
 * @property {boolean} [validateOnBlur=true] - Validate on blur
 * @property {boolean} [validateOnInput=false] - Validate on input
 * @property {number} [debounce=300] - Debounce delay in ms
 * @property {string} [csrfTokenSelector] - CSRF token meta selector
 * @property {string} [toastStore='toast'] - Alpine toast store name
 */

/**
 * Form Validation Hook
 * 
 * Provides comprehensive form validation with support for sync/async validation,
 * real-time feedback, and customizable validation rules.
 * 
 * @param {FormValidationConfig} config - Configuration options
 * @returns {Object} Alpine.js component data
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
function useFormValidation(config = {}) {
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
        init() {
            this.validateForm();
            
            // Watch for form changes with optimized deep watching
            this.$watch('formData', () => {
                this._handleFormDataChange();
            }, { deep: true });
        },
        
        /**
         * Cleanup on component destroy
         */
        destroy() {
            this._destroyed = true;
            this._clearValidationTimeout();
            this._abortAllAsyncValidations();
        },
        
        /**
         * Handle form data changes
         * @private
         */
        _handleFormDataChange() {
            if (this._destroyed) return;
            
            this.markDirty();
            
            if (this.realTimeValidation) {
                this._debouncedValidation();
            }
        },
        
        /**
         * Validate the entire form
         * @returns {boolean} True if form is valid
         */
        validateForm() {
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
         * @param {string} field - Field name
         * @param {*} value - Field value
         * @param {ValidationRule} rules - Validation rules
         * @returns {string[]} Array of error messages
         */
        validateField(field, value, rules) {
            const errors = [];
            
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
         * @param {string} field - Field name
         * @param {*} value - Field value
         * @param {Function} asyncValidator - Async validation function
         * @private
         */
        async _performAsyncValidation(field, value, asyncValidator) {
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
            } catch (error) {
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
         * @param {string} field - Field name
         * @private
         */
        _abortAsyncValidation(field) {
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
        _abortAllAsyncValidations() {
            this._asyncControllers.forEach(controller => controller.abort());
            this._asyncControllers.clear();
        },
        
        /**
         * Clear validation timeout
         * @private
         */
        _clearValidationTimeout() {
            if (this._validationTimeout) {
                clearTimeout(this._validationTimeout);
                this._validationTimeout = null;
            }
        },
        
        /**
         * Debounced validation for real-time feedback
         * @private
         */
        _debouncedValidation() {
            this._clearValidationTimeout();
            this._validationTimeout = setTimeout(() => {
                if (!this._destroyed) {
                    this.validateForm();
                }
            }, this.validationDebounce);
        },
        
        /**
         * Submit the form
         * @param {string} endpoint - Submission endpoint URL
         * @param {Object} [options={}] - Submission options
         * @param {string} [options.method='POST'] - HTTP method
         * @param {Object} [options.headers={}] - Additional headers
         * @param {Function} [options.transformData] - Transform function for form data
         * @returns {Promise<Object>} Response object
         */
        async submitForm(endpoint, options = {}) {
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
                
            } catch (error) {
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
         * @returns {FormData} Prepared form data
         */
        prepareFormData() {
            const formData = new FormData();
            
            // Add all form fields
            Object.entries(this.formData).forEach(([key, value]) => {
                this._appendFormDataValue(formData, key, value);
            });
            
            // Add CSRF token
            const csrfToken = document.querySelector(this.csrfTokenSelector)?.content;
            if (csrfToken) {
                formData.append('_token', csrfToken);
            }
            
            return formData;
        },
        
        /**
         * Append value to FormData with type handling
         * @param {FormData} formData - FormData object
         * @param {string} key - Field key
         * @param {*} value - Field value
         * @private
         */
        _appendFormDataValue(formData, key, value) {
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
         * @param {Object} response - Server response
         * @private
         */
        _onSubmitSuccess(response) {
            this.isDirty = false;
            this.validationErrors = {};
            
            // Show success message
            const toastStore = Alpine.store(this.toastStoreName);
            if (response.message && toastStore?.success) {
                toastStore.success(response.message);
            }
            
            // Emit success event
            this.$dispatch('form-submit-success', { 
                response, 
                formData: { ...this.formData } 
            });
        },
        
        /**
         * Handle form submission error
         * @param {Object} response - Error response
         * @private
         */
        _onSubmitError(response) {
            if (response.errors) {
                this.validationErrors = response.errors;
                this.focusFirstError();
            }
            
            // Show error message
            const message = response.message || 'Form submission failed. Please check the errors and try again.';
            const toastStore = Alpine.store(this.toastStoreName);
            if (toastStore?.error) {
                toastStore.error(message);
            }
            
            // Emit error event
            this.$dispatch('form-submit-error', { 
                response, 
                formData: { ...this.formData } 
            });
        },
        
        /**
         * Handle field input event
         * @param {string} field - Field name
         * @param {*} value - New value
         */
        handleFieldInput(field, value) {
            this.setNestedValue(this.formData, field, value);
            
            if (this.validateOnInput || this.submitAttempted) {
                this.validateSingleField(field);
            }
        },
        
        /**
         * Handle field blur event
         * @param {string} field - Field name
         */
        handleFieldBlur(field) {
            if (this.validateOnBlur || this.submitAttempted) {
                this.validateSingleField(field);
            }
        },
        
        /**
         * Validate a single field by name
         * @param {string} field - Field name
         */
        validateSingleField(field) {
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
         * @param {*} value - Value to check
         * @returns {boolean} True if empty
         */
        isEmpty(value) {
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
         * @param {string} email - Email to validate
         * @returns {boolean} True if valid email
         */
        isValidEmail(email) {
            // More comprehensive email regex
            const emailRegex = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;
            return emailRegex.test(email);
        },
        
        /**
         * Convert field name to human-readable label
         * @param {string} field - Field name
         * @returns {string} Human-readable label
         */
        getFieldLabel(field) {
            // Handle dot notation
            const fieldName = field.split('.').pop();
            
            return fieldName
                .split('_')
                .map(word => word.charAt(0).toUpperCase() + word.slice(1))
                .join(' ');
        },
        
        /**
         * Get nested object value by path
         * @param {Object} obj - Source object
         * @param {string} path - Dot-notation path
         * @returns {*} Value at path
         */
        getNestedValue(obj, path) {
            if (!path || !obj) return undefined;
            return path.split('.').reduce((current, key) => current?.[key], obj);
        },
        
        /**
         * Set nested object value by path
         * @param {Object} obj - Target object
         * @param {string} path - Dot-notation path
         * @param {*} value - Value to set
         */
        setNestedValue(obj, path, value) {
            if (!path || !obj) return;
            
            const keys = path.split('.');
            const lastKey = keys.pop();
            
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
        focusFirstError() {
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
                const element = document.querySelector(selector);
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
        markDirty() {
            this.isDirty = true;
        },
        
        /**
         * Reset form to initial state
         * @param {Object} [newInitialData] - Optional new initial data
         */
        reset(newInitialData) {
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
         * @param {string} url - Request URL
         * @param {Object} options - Request options
         * @returns {Promise<Object>} Response data
         * @private
         */
        async _makeRequest(url, options) {
            const requestOptions = {
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
         * @param {string} field - Field name
         * @returns {string[]} Array of error messages
         */
        getFieldErrors(field) {
            return this.validationErrors[field] || [];
        },
        
        /**
         * Check if a field has errors
         * @param {string} field - Field name
         * @returns {boolean} True if field has errors
         */
        hasFieldError(field) {
            return !!(this.validationErrors[field] && this.validationErrors[field].length > 0);
        },
        
        /**
         * Get the first error for a field
         * @param {string} field - Field name
         * @returns {string|null} First error message or null
         */
        getFirstFieldError(field) {
            const errors = this.getFieldErrors(field);
            return errors.length > 0 ? errors[0] : null;
        },
        
        // Computed properties (getters for backward compatibility)
        get hasErrors() {
            return Object.keys(this.validationErrors).length > 0;
        },
        
        get canSubmit() {
            return this.isValid && !this.isSubmitting;
        },
        
        // Backward compatibility aliases
        get fieldErrors() {
            return (field) => this.getFieldErrors(field);
        },
        
        get hasFieldError() {
            return (field) => this.hasFieldError(field);
        }
    };
}

/**
 * @typedef {Object} PasswordStrengthConfig
 * @property {string} [fieldName='password'] - Name of password field
 * @property {number} [minLength=8] - Minimum password length
 */

/**
 * Password Strength Hook
 * 
 * Calculates password strength and provides feedback
 * 
 * @param {PasswordStrengthConfig} [config={}] - Configuration options
 * @returns {Object} Alpine.js component data
 * 
 * @example
 * <div x-data="usePasswordStrength()">
 */
function usePasswordStrength(config = {}) {
    const minLength = config.minLength || 8;
    
    return {
        password: '',
        strength: 0,
        feedback: '',
        
        init() {
            this.$watch('password', () => {
                this.calculateStrength();
            });
        },
        
        /**
         * Calculate password strength score
         */
        calculateStrength() {
            let score = 0;
            const feedback = [];
            
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
         * @returns {string} CSS class name
         */
        get strengthClass() {
            const classes = ['weak', 'weak', 'fair', 'good', 'strong', 'very-strong'];
            return classes[Math.min(this.strength, classes.length - 1)] || 'weak';
        },
        
        /**
         * Get human-readable strength text
         * @returns {string} Strength description
         */
        get strengthText() {
            const texts = ['Very Weak', 'Weak', 'Fair', 'Good', 'Strong', 'Very Strong'];
            return texts[Math.min(this.strength, texts.length - 1)] || 'Very Weak';
        },
        
        /**
         * Get strength percentage (0-100)
         * @returns {number} Strength percentage
         */
        get strengthPercentage() {
            return Math.min((this.strength / 5) * 100, 100);
        },
        
        /**
         * Check if password meets minimum requirements
         * @returns {boolean} True if password is strong enough
         */
        get isStrongEnough() {
            return this.strength >= 3; // At least "Good"
        }
    };
}

/**
 * @typedef {Object} ProgressiveDisclosureConfig
 * @property {Object.<string, boolean>} [sections={}] - Initial section visibility
 * @property {number} [initialSection=0] - Initial active section index
 */

/**
 * Progressive Disclosure Hook
 * 
 * Manages progressive disclosure of form sections or content
 * 
 * @param {ProgressiveDisclosureConfig} [config={}] - Configuration options
 * @returns {Object} Alpine.js component data
 * 
 * @example
 * <div x-data="useProgressiveDisclosure({ sections: { basic: true, advanced: false } })">
 */
function useProgressiveDisclosure(config = {}) {
    return {
        sections: { ...(config.sections || {}) },
        currentSection: config.initialSection || 0,
        
        /**
         * Show a section
         * @param {string} sectionKey - Section identifier
         */
        showSection(sectionKey) {
            this.sections[sectionKey] = true;
        },
        
        /**
         * Hide a section
         * @param {string} sectionKey - Section identifier
         */
        hideSection(sectionKey) {
            this.sections[sectionKey] = false;
        },
        
        /**
         * Toggle section visibility
         * @param {string} sectionKey - Section identifier
         */
        toggleSection(sectionKey) {
            this.sections[sectionKey] = !this.sections[sectionKey];
        },
        
        /**
         * Check if a section is visible
         * @param {string} sectionKey - Section identifier
         * @returns {boolean} True if section is visible
         */
        isSectionVisible(sectionKey) {
            return !!this.sections[sectionKey];
        },
        
        /**
         * Show all sections
         */
        showAllSections() {
            Object.keys(this.sections).forEach(key => {
                this.sections[key] = true;
            });
        },
        
        /**
         * Hide all sections
         */
        hideAllSections() {
            Object.keys(this.sections).forEach(key => {
                this.sections[key] = false;
            });
        },
        
        /**
         * Move to next section
         * @returns {boolean} True if moved successfully
         */
        nextSection() {
            if (this.currentSection < this.maxSections - 1) {
                this.currentSection++;
                return true;
            }
            return false;
        },
        
        /**
         * Move to previous section
         * @returns {boolean} True if moved successfully
         */
        prevSection() {
            if (this.currentSection > 0) {
                this.currentSection--;
                return true;
            }
            return false;
        },
        
        /**
         * Go to specific section
         * @param {number} section - Section index
         * @returns {boolean} True if moved successfully
         */
        goToSection(section) {
            if (section >= 0 && section < this.maxSections) {
                this.currentSection = section;
                return true;
            }
            return false;
        },
        
        /**
         * Get total number of sections
         * @returns {number} Section count
         */
        get maxSections() {
            return Object.keys(this.sections).length;
        },
        
        /**
         * Get progress percentage
         * @returns {number} Progress (0-100)
         */
        get progress() {
            if (this.maxSections === 0) return 0;
            return ((this.currentSection + 1) / this.maxSections) * 100;
        },
        
        /**
         * Check if on first section
         * @returns {boolean} True if on first section
         */
        get isFirstSection() {
            return this.currentSection === 0;
        },
        
        /**
         * Check if on last section
         * @returns {boolean} True if on last section
         */
        get isLastSection() {
            return this.currentSection === this.maxSections - 1;
        }
    };
}

/**
 * @typedef {Object} MultiStepFormConfig
 * @extends FormValidationConfig
 * @property {Object.<number, Object.<string, ValidationRule>>} [stepValidation={}] - Validation rules per step
 * @property {boolean} [allowStepSkipping=false] - Allow skipping steps without validation
 */

/**
 * Multi-step Form Hook
 * 
 * Combines form validation with step-by-step navigation
 * 
 * @param {MultiStepFormConfig} [config={}] - Configuration options
 * @returns {Object} Alpine.js component data
 * 
 * @example
 * <div x-data="useMultiStepForm({
 *   stepValidation: {
 *     0: { email: { required: true, email: true } },
 *     1: { password: { required: true, minLength: 8 } }
 *   }
 * })">
 */
function useMultiStepForm(config = {}) {
    // Create base components without method conflicts
    const formValidation = useFormValidation(config);
    const progressiveDisclosure = useProgressiveDisclosure(config);
    
    // Extract methods that might conflict
    const { init: formInit, destroy: formDestroy, ...formRest } = formValidation;
    const { 
        init: disclosureInit, 
        nextSection, 
        prevSection, 
        goToSection,
        ...disclosureRest 
    } = progressiveDisclosure;
    
    return {
        // Merge base functionality
        ...formRest,
        ...disclosureRest,
        
        // Multi-step specific configuration
        stepValidation: config.stepValidation || {},
        allowStepSkipping: config.allowStepSkipping || false,
        
        // Multi-step specific state
        completedSteps: new Set(),
        _stepValidationCache: new Map(),
        
        /**
         * Initialize multi-step form
         */
        init() {
            // Call both parent inits
            if (formInit) formInit.call(this);
            if (disclosureInit) disclosureInit.call(this);
            
            // Mark first step as current
            this._updateStepValidationRules();
        },
        
        /**
         * Cleanup on destroy
         */
        destroy() {
            if (formDestroy) formDestroy.call(this);
            this._stepValidationCache.clear();
        },
        
        /**
         * Update validation rules for current step
         * @private
         */
        _updateStepValidationRules() {
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
         * @param {number} step - Target step index
         * @param {boolean} [skipValidation=false] - Skip validation
         * @returns {Promise<boolean>} True if navigation successful
         */
        async goToStep(step, skipValidation = false) {
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
            this.$dispatch('step-changed', { 
                step, 
                previousStep: this.currentSection,
                isCompleted: this.isStepCompleted(step)
            });
            
            return true;
        },
        
        /**
         * Move to next step
         * @returns {Promise<boolean>} True if navigation successful
         */
        async nextStep() {
            if (!this.canGoToNextStep) {
                return false;
            }
            return await this.goToStep(this.currentSection + 1);
        },
        
        /**
         * Move to previous step
         * @returns {Promise<boolean>} True if navigation successful
         */
        async prevStep() {
            if (!this.canGoToPrevStep) {
                return false;
            }
            // No validation when going back
            return await this.goToStep(this.currentSection - 1, true);
        },
        
        /**
         * Validate current step
         * @returns {Promise<boolean>} True if step is valid
         */
        async validateCurrentStep() {
            return this.validateStep(this.currentSection);
        },
        
        /**
         * Validate a specific step
         * @param {number} step - Step index to validate
         * @returns {Promise<boolean>} True if step is valid
         */
        async validateStep(step) {
            const stepRules = this.stepValidation[step];
            if (!stepRules) {
                return true; // No validation rules for this step
            }
            
            // Validate only fields in this step
            const stepErrors = {};
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
         * @returns {Promise<boolean>} True if all steps are valid
         */
        async validateAllSteps() {
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
         * @param {number} step - Step index
         * @returns {boolean} True if step has validation rules
         * @private
         */
        _shouldValidateStep(step) {
            return !!this.stepValidation[step] && 
                   Object.keys(this.stepValidation[step]).length > 0;
        },
        
        /**
         * Mark step as completed
         * @param {number} step - Step index
         */
        markStepCompleted(step) {
            this.completedSteps.add(step);
            
            this.$dispatch('step-completed', { 
                step,
                completedSteps: Array.from(this.completedSteps)
            });
        },
        
        /**
         * Mark step as incomplete
         * @param {number} step - Step index
         */
        markStepIncomplete(step) {
            this.completedSteps.delete(step);
        },
        
        /**
         * Check if step is completed
         * @param {number} step - Step index
         * @returns {boolean} True if step is completed
         */
        isStepCompleted(step) {
            return this.completedSteps.has(step);
        },
        
        /**
         * Check if step is accessible (all previous steps completed or skipping allowed)
         * @param {number} step - Step index
         * @returns {boolean} True if step can be accessed
         */
        isStepAccessible(step) {
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
         * @param {number} step - Step index
         * @returns {Object.<string, string[]>} Errors by field
         */
        getStepErrors(step) {
            const stepRules = this.stepValidation[step];
            if (!stepRules) return {};
            
            const stepFields = Object.keys(stepRules);
            const stepErrors = {};
            
            stepFields.forEach(field => {
                if (this.validationErrors[field]) {
                    stepErrors[field] = this.validationErrors[field];
                }
            });
            
            return stepErrors;
        },
        
        /**
         * Check if step has errors
         * @param {number} step - Step index
         * @returns {boolean} True if step has errors
         */
        hasStepErrors(step) {
            const stepErrors = this.getStepErrors(step);
            return Object.keys(stepErrors).length > 0;
        },
        
        /**
         * Submit multi-step form
         * @param {string} endpoint - Submission endpoint
         * @param {Object} [options={}] - Submission options
         * @returns {Promise<Object>} Response object
         */
        async submitMultiStepForm(endpoint, options = {}) {
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
         * @param {Object} [newInitialData] - Optional new initial data
         */
        resetMultiStepForm(newInitialData) {
            this.reset(newInitialData);
            this.completedSteps.clear();
            this.currentSection = 0;
            this._updateStepValidationRules();
            this._stepValidationCache.clear();
        },
        
        // Computed properties
        
        /**
         * Check if can go to next step
         * @returns {boolean} True if next step is available
         */
        get canGoToNextStep() {
            return this.currentSection < this.maxSections - 1;
        },
        
        /**
         * Check if can go to previous step
         * @returns {boolean} True if previous step is available
         */
        get canGoToPrevStep() {
            return this.currentSection > 0;
        },
        
        /**
         * Check if current step is valid
         * @returns {boolean} True if current step has no errors
         */
        get isCurrentStepValid() {
            return !this.hasStepErrors(this.currentSection);
        },
        
        /**
         * Get completion percentage
         * @returns {number} Percentage (0-100)
         */
        get completionPercentage() {
            if (this.maxSections === 0) return 0;
            return (this.completedSteps.size / this.maxSections) * 100;
        },
        
        /**
         * Check if form is complete (all steps completed)
         * @returns {boolean} True if all steps are completed
         */
        get isFormComplete() {
            return this.completedSteps.size === this.maxSections;
        },
        
        /**
         * Get current step number (1-indexed for display)
         * @returns {number} Current step number
         */
        get currentStepNumber() {
            return this.currentSection + 1;
        },
        
        /**
         * Get total step count
         * @returns {number} Total steps
         */
        get totalSteps() {
            return this.maxSections;
        },
        
        // Override goToSection to use goToStep
        goToSection(section) {
            return this.goToStep(section);
        },
        
        // Backward compatibility aliases
        submitForm: function(endpoint, options = {}) {
            console.warn('useMultiStepForm: Use submitMultiStepForm() instead of submitForm()');
            return this.submitMultiStepForm(endpoint, options);
        },
        
        reset: function(newInitialData) {
            return this.resetMultiStepForm(newInitialData);
        }
    };
}

/**
 * Export hooks to window for global access
 */
if (typeof window !== 'undefined') {
    window.useFormValidation = useFormValidation;
    window.usePasswordStrength = usePasswordStrength;
    window.useProgressiveDisclosure = useProgressiveDisclosure;
    window.useMultiStepForm = useMultiStepForm;
}

/**
 * Export for module systems
 */
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        useFormValidation,
        usePasswordStrength,
        useProgressiveDisclosure,
        useMultiStepForm
    };
}
