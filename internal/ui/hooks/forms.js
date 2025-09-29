/**
 * Form Validation AlpineJS Hooks
 * 
 * Extracted from internal/ui/components/forms/*.templ
 * Provides reusable Form validation and interaction patterns
 */

/**
 * Form Validation Hook
 * Usage: x-data="useFormValidation(config)"
 */
function useFormValidation(config = {}) {
    return {
        // State
        formData: config.initialData || {},
        validationErrors: {},
        validationRules: config.rules || {},
        
        // Status tracking
        isValid: false,
        isDirty: false,
        isSubmitting: false,
        submitAttempted: false,
        
        // Real-time validation
        realTimeValidation: config.realTimeValidation !== false,
        validateOnBlur: config.validateOnBlur !== false,
        validateOnInput: config.validateOnInput || false,
        
        // Debounce settings
        validationDebounce: config.debounce || 300,
        validationTimeouts: {},
        
        // Initialization
        init() {
            this.validateForm();
            
            // Watch for form changes
            this.$watch('formData', () => {
                this.markDirty();
                if (this.realTimeValidation) {
                    this.debouncedValidation();
                }
            }, { deep: true });
        },
        
        // Core validation methods
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
        
        validateField(field, value, rules) {
            const errors = [];
            
            // Required validation
            if (rules.required && this.isEmpty(value)) {
                errors.push(rules.requiredMessage || `${this.getFieldLabel(field)} is required`);
            }
            
            // Skip other validations if field is empty and not required
            if (this.isEmpty(value) && !rules.required) {
                return errors;
            }
            
            // Min length validation
            if (rules.minLength && String(value).length < rules.minLength) {
                errors.push(rules.minLengthMessage || `${this.getFieldLabel(field)} must be at least ${rules.minLength} characters`);
            }
            
            // Max length validation
            if (rules.maxLength && String(value).length > rules.maxLength) {
                errors.push(rules.maxLengthMessage || `${this.getFieldLabel(field)} cannot exceed ${rules.maxLength} characters`);
            }
            
            // Pattern validation
            if (rules.pattern && !new RegExp(rules.pattern).test(String(value))) {
                errors.push(rules.patternMessage || `${this.getFieldLabel(field)} format is invalid`);
            }
            
            // Email validation
            if (rules.email && !this.isValidEmail(String(value))) {
                errors.push(rules.emailMessage || `${this.getFieldLabel(field)} must be a valid email address`);
            }
            
            // Custom validation functions
            if (rules.custom && typeof rules.custom === 'function') {
                const customError = rules.custom(value, this.formData);
                if (customError) {
                    errors.push(customError);
                }
            }
            
            // Async validation (if configured)
            if (rules.async && typeof rules.async === 'function') {
                this.performAsyncValidation(field, value, rules.async);
            }
            
            return errors;
        },
        
        // Async validation with server-side checks
        async performAsyncValidation(field, value, asyncValidator) {
            try {
                const result = await asyncValidator(value, this.formData);
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
                console.error('Async validation failed:', error);
            }
        },
        
        // Debounced validation for real-time feedback
        debouncedValidation() {
            clearTimeout(this.validationTimeout);
            this.validationTimeout = setTimeout(() => {
                this.validateForm();
            }, this.validationDebounce);
        },
        
        // Form submission
        async submitForm(endpoint, options = {}) {
            this.submitAttempted = true;
            
            if (this.isSubmitting) return;
            
            this.isSubmitting = true;
            
            try {
                // Final validation
                if (!this.validateForm()) {
                    this.focusFirstError();
                    return { success: false, errors: this.validationErrors };
                }
                
                // Prepare form data
                const formData = this.prepareFormData();
                
                // Submit to server
                const response = await this.makeRequest(endpoint, {
                    method: options.method || 'POST',
                    data: formData
                });
                
                if (response.success) {
                    this.onSubmitSuccess(response);
                } else {
                    this.onSubmitError(response);
                }
                
                return response;
                
            } catch (error) {
                console.error('Form submission failed:', error);
                const errorResponse = { success: false, error: error.message };
                this.onSubmitError(errorResponse);
                return errorResponse;
                
            } finally {
                this.isSubmitting = false;
            }
        },
        
        // Form data preparation
        prepareFormData() {
            const formData = new FormData();
            
            // Add all form fields
            Object.entries(this.formData).forEach(([key, value]) => {
                if (value instanceof FileList) {
                    Array.from(value).forEach(file => formData.append(key, file));
                } else if (Array.isArray(value)) {
                    value.forEach(v => formData.append(`${key}[]`, v));
                } else if (value !== null && value !== undefined) {
                    formData.append(key, value);
                }
            });
            
            // Add CSRF token
            const csrfToken = document.querySelector('meta[name="csrf-token"]')?.content;
            if (csrfToken) {
                formData.append('_token', csrfToken);
            }
            
            return formData;
        },
        
        // Event handlers
        onSubmitSuccess(response) {
            this.isDirty = false;
            this.validationErrors = {};
            
            // Show success message
            if (response.message) {
                Alpine.store('toast')?.success?.(response.message);
            }
            
            // Emit success event
            this.$dispatch('form-submit-success', { response, formData: this.formData });
        },
        
        onSubmitError(response) {
            if (response.errors) {
                this.validationErrors = response.errors;
                this.focusFirstError();
            }
            
            // Show error message
            const message = response.message || 'Form submission failed. Please check the errors and try again.';
            Alpine.store('toast')?.error?.(message);
            
            // Emit error event
            this.$dispatch('form-submit-error', { response, formData: this.formData });
        },
        
        // Field interaction handlers
        handleFieldInput(field, value) {
            this.setNestedValue(this.formData, field, value);
            
            if (this.validateOnInput) {
                this.validateSingleField(field);
            }
        },
        
        handleFieldBlur(field) {
            if (this.validateOnBlur) {
                this.validateSingleField(field);
            }
        },
        
        validateSingleField(field) {
            const rules = this.validationRules[field];
            if (!rules) return;
            
            const value = this.getNestedValue(this.formData, field);
            const errors = this.validateField(field, value, rules);
            
            if (errors.length > 0) {
                this.validationErrors[field] = errors;
            } else {
                delete this.validationErrors[field];
            }
        },
        
        // Utility methods
        isEmpty(value) {
            return value === null || value === undefined || value === '' || 
                   (Array.isArray(value) && value.length === 0);
        },
        
        isValidEmail(email) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            return emailRegex.test(email);
        },
        
        getFieldLabel(field) {
            // Convert field name to human-readable label
            return field.split('_').map(word => 
                word.charAt(0).toUpperCase() + word.slice(1)
            ).join(' ');
        },
        
        getNestedValue(obj, path) {
            return path.split('.').reduce((current, key) => current?.[key], obj);
        },
        
        setNestedValue(obj, path, value) {
            const keys = path.split('.');
            const lastKey = keys.pop();
            const target = keys.reduce((current, key) => {
                if (!(key in current)) {
                    current[key] = {};
                }
                return current[key];
            }, obj);
            target[lastKey] = value;
        },
        
        focusFirstError() {
            const firstErrorField = Object.keys(this.validationErrors)[0];
            if (firstErrorField) {
                const element = document.querySelector(`[name="${firstErrorField}"], [data-field="${firstErrorField}"]`);
                if (element) {
                    element.focus();
                    element.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }
            }
        },
        
        markDirty() {
            this.isDirty = true;
        },
        
        reset() {
            this.formData = config.initialData || {};
            this.validationErrors = {};
            this.isDirty = false;
            this.submitAttempted = false;
        },
        
        async makeRequest(url, options) {
            const response = await fetch(url, {
                method: options.method || 'POST',
                body: options.data,
                headers: {
                    'X-Requested-With': 'XMLHttpRequest'
                }
            });
            
            return await response.json();
        },
        
        // Computed properties
        get hasErrors() {
            return Object.keys(this.validationErrors).length > 0;
        },
        
        get canSubmit() {
            return this.isValid && !this.isSubmitting;
        },
        
        get fieldErrors() {
            return (field) => this.validationErrors[field] || [];
        },
        
        get hasFieldError() {
            return (field) => !!(this.validationErrors[field] && this.validationErrors[field].length > 0);
        }
    };
}

/**
 * Password Strength Hook
 * Usage: x-data="usePasswordStrength()"
 */
function usePasswordStrength() {
    return {
        password: '',
        strength: 0,
        feedback: '',
        
        init() {
            this.$watch('password', () => {
                this.calculateStrength();
            });
        },
        
        calculateStrength() {
            let score = 0;
            const feedback = [];
            
            if (this.password.length === 0) {
                this.strength = 0;
                this.feedback = '';
                return;
            }
            
            // Length check
            if (this.password.length >= 8) {
                score += 1;
            } else {
                feedback.push('Use at least 8 characters');
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
            
            this.strength = score;
            this.feedback = feedback.join(', ');
        },
        
        get strengthClass() {
            const classes = ['weak', 'fair', 'good', 'strong', 'very-strong'];
            return classes[this.strength] || 'weak';
        },
        
        get strengthText() {
            const texts = ['Very Weak', 'Weak', 'Fair', 'Good', 'Strong'];
            return texts[this.strength] || 'Very Weak';
        }
    };
}

/**
 * Progressive Disclosure Hook
 * Usage: x-data="useProgressiveDisclosure()"
 */
function useProgressiveDisclosure(config = {}) {
    return {
        sections: config.sections || {},
        currentSection: config.initialSection || 0,
        
        showSection(sectionKey) {
            this.sections[sectionKey] = true;
        },
        
        hideSection(sectionKey) {
            this.sections[sectionKey] = false;
        },
        
        toggleSection(sectionKey) {
            this.sections[sectionKey] = !this.sections[sectionKey];
        },
        
        nextSection() {
            this.currentSection = Math.min(this.currentSection + 1, this.maxSections - 1);
        },
        
        prevSection() {
            this.currentSection = Math.max(this.currentSection - 1, 0);
        },
        
        goToSection(section) {
            this.currentSection = section;
        },
        
        get maxSections() {
            return Object.keys(this.sections).length;
        },
        
        get progress() {
            return ((this.currentSection + 1) / this.maxSections) * 100;
        }
    };
}

/**
 * Multi-step Form Hook
 * Usage: x-data="useMultiStepForm(config)"
 */
function useMultiStepForm(config = {}) {
    const formValidation = useFormValidation(config);
    const progressiveDisclosure = useProgressiveDisclosure(config);
    
    return {
        ...formValidation,
        ...progressiveDisclosure,
        
        // Multi-step specific state
        completedSteps: new Set(),
        stepValidation: config.stepValidation || {},
        
        // Step navigation
        async goToStep(step) {
            // Validate current step before moving
            if (this.shouldValidateStep(this.currentSection)) {
                const isValid = await this.validateCurrentStep();
                if (!isValid) {
                    return false;
                }
            }
            
            this.markStepCompleted(this.currentSection);
            this.goToSection(step);
            return true;
        },
        
        async nextStep() {
            const success = await this.goToStep(this.currentSection + 1);
            return success;
        },
        
        async prevStep() {
            this.goToSection(this.currentSection - 1);
            return true;
        },
        
        validateCurrentStep() {
            const stepRules = this.stepValidation[this.currentSection];
            if (!stepRules) return true;
            
            // Validate only fields in current step
            const stepErrors = {};
            let isStepValid = true;
            
            Object.entries(stepRules).forEach(([field, rules]) => {
                const value = this.getNestedValue(this.formData, field);
                const fieldErrors = this.validateField(field, value, rules);
                
                if (fieldErrors.length > 0) {
                    stepErrors[field] = fieldErrors;
                    isStepValid = false;
                }
            });
            
            // Update validation errors with step-specific errors
            this.validationErrors = { ...this.validationErrors, ...stepErrors };
            
            return isStepValid;
        },
        
        shouldValidateStep(step) {
            return !!this.stepValidation[step];
        },
        
        markStepCompleted(step) {
            this.completedSteps.add(step);
        },
        
        isStepCompleted(step) {
            return this.completedSteps.has(step);
        },
        
        get canGoToNextStep() {
            return this.currentSection < this.maxSections - 1;
        },
        
        get canGoToPrevStep() {
            return this.currentSection > 0;
        }
    };
}

// Export for use in components
window.useFormValidation = useFormValidation;
window.usePasswordStrength = usePasswordStrength;
window.useProgressiveDisclosure = useProgressiveDisclosure;
window.useMultiStepForm = useMultiStepForm;