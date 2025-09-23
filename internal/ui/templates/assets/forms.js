// Form components Alpine.js functions

// Tenant Form Component
function tenantForm() {
    return {
        formValid: false,
        
        initForm() {
            // Initialize form validation state
            this.validateForm();
            
            // Watch for form changes
            this.$el.addEventListener('input', () => {
                this.validateForm();
            });
            
            // Auto-generate subdomain from tenant name
            const nameField = this.$el.querySelector('[name="name"]');
            const subdomainField = this.$el.querySelector('[name="subdomain"]');
            
            if (nameField && subdomainField && !subdomainField.value) {
                nameField.addEventListener('input', (e) => {
                    if (!subdomainField.value) {
                        const subdomain = this.generateSubdomain(e.target.value);
                        subdomainField.value = subdomain;
                        
                        // Trigger validation on the subdomain field
                        subdomainField.dispatchEvent(new Event('input', { bubbles: true }));
                    }
                });
            }
        },
        
        validateForm() {
            const form = this.$el.closest('form');
            if (!form) return;
            
            const requiredFields = form.querySelectorAll('[required]');
            let allValid = true;
            
            requiredFields.forEach(field => {
                if (!field.value.trim()) {
                    allValid = false;
                }
            });
            
            this.formValid = allValid;
            
            // Update submit button state
            const submitButton = form.querySelector('[type="submit"]');
            if (submitButton) {
                submitButton.disabled = !allValid;
                submitButton.classList.toggle('opacity-50', !allValid);
                submitButton.classList.toggle('cursor-not-allowed', !allValid);
            }
        },
        
        generateSubdomain(name) {
            return name
                .toLowerCase()
                .replace(/[^a-z0-9\s-]/g, '') // Remove special characters
                .replace(/\s+/g, '-') // Replace spaces with hyphens
                .replace(/-+/g, '-') // Replace multiple hyphens with single
                .replace(/^-|-$/g, '') // Remove leading/trailing hyphens
                .substring(0, 50); // Limit length
        }
    }
}

// User Form Component
function userForm() {
    return {
        formValid: false,
        
        initForm() {
            this.validateForm();
            
            // Watch for form changes
            this.$el.addEventListener('input', () => {
                this.validateForm();
            });
            
            // Auto-generate username from email
            const emailField = this.$el.querySelector('[name="email"]');
            const usernameField = this.$el.querySelector('[name="username"]');
            
            if (emailField && usernameField && !usernameField.value) {
                emailField.addEventListener('input', (e) => {
                    if (!usernameField.value) {
                        const username = this.generateUsername(e.target.value);
                        usernameField.value = username;
                        usernameField.dispatchEvent(new Event('input', { bubbles: true }));
                    }
                });
            }
        },
        
        validateForm() {
            const form = this.$el.querySelector('form');
            if (!form) return;
            
            const requiredFields = form.querySelectorAll('[required]');
            let allValid = true;
            
            requiredFields.forEach(field => {
                if (!field.value.trim()) {
                    allValid = false;
                }
            });
            
            // Check password confirmation
            const password = form.querySelector('[name="password"]');
            const passwordConfirm = form.querySelector('[name="password_confirm"]');
            
            if (password && passwordConfirm && password.value !== passwordConfirm.value) {
                allValid = false;
            }
            
            this.formValid = allValid;
        },
        
        generateUsername(email) {
            const username = email.split('@')[0];
            return username
                .toLowerCase()
                .replace(/[^a-z0-9]/g, '')
                .substring(0, 20);
        },
        
        handleSubmit(e) {
            if (!this.formValid) {
                e.preventDefault();
                Alpine.store('app').addNotification('error', 'Please fix form errors before submitting');
            }
        }
    }
}

// Password Strength Component
function passwordStrength() {
    return {
        password: '',
        strength: 'weak',
        strengthPercent: 0,
        hasMinLength: false,
        hasUppercase: false,
        hasLowercase: false,
        hasNumber: false,
        hasSpecial: false,
        
        init() {
            this.$watch('password', () => {
                this.calculateStrength();
            });
            
            // Watch password input
            const passwordField = this.$el.querySelector('[name="password"]');
            if (passwordField) {
                passwordField.addEventListener('input', (e) => {
                    this.password = e.target.value;
                });
            }
        },
        
        calculateStrength() {
            this.hasMinLength = this.password.length >= 8;
            this.hasUppercase = /[A-Z]/.test(this.password);
            this.hasLowercase = /[a-z]/.test(this.password);
            this.hasNumber = /[0-9]/.test(this.password);
            this.hasSpecial = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(this.password);
            
            const score = [
                this.hasMinLength,
                this.hasUppercase,
                this.hasLowercase,
                this.hasNumber,
                this.hasSpecial
            ].filter(Boolean).length;
            
            if (score < 3) {
                this.strength = 'weak';
                this.strengthPercent = 25;
            } else if (score < 5) {
                this.strength = 'medium';
                this.strengthPercent = 60;
            } else {
                this.strength = 'strong';
                this.strengthPercent = 100;
            }
        }
    }
}

// Global form utilities
window.formUtils = {
    // Auto-save form drafts
    enableAutoSave(formId, interval = 30000) {
        const form = document.getElementById(formId);
        if (!form) return;
        
        setInterval(() => {
            const formData = new FormData(form);
            const data = {};
            for (let [key, value] of formData.entries()) {
                data[key] = value;
            }
            
            Alpine.store('forms').saveDraft(formId, data);
        }, interval);
    },
    
    // Restore form from draft
    restoreFromDraft(formId) {
        const draft = Alpine.store('forms').loadDraft(formId);
        if (!draft) return false;
        
        const form = document.getElementById(formId);
        if (!form) return false;
        
        Object.keys(draft).forEach(key => {
            const field = form.querySelector(`[name="${key}"]`);
            if (field) {
                field.value = draft[key];
                field.dispatchEvent(new Event('input', { bubbles: true }));
            }
        });
        
        return true;
    },
    
    // Validate entire form
    validateForm(formId) {
        const form = document.getElementById(formId);
        if (!form) return false;
        
        const requiredFields = form.querySelectorAll('[required]');
        let isValid = true;
        
        requiredFields.forEach(field => {
            if (!field.value.trim()) {
                isValid = false;
                field.classList.add('border-red-500');
            } else {
                field.classList.remove('border-red-500');
            }
        });
        
        return isValid;
    },
    
    // Show form errors
    showFormErrors(errors) {
        Object.keys(errors).forEach(fieldName => {
            const field = document.querySelector(`[name="${fieldName}"]`);
            if (field) {
                const errorDiv = field.parentNode.querySelector('.text-red-500');
                if (errorDiv) {
                    errorDiv.textContent = errors[fieldName];
                    errorDiv.style.display = 'block';
                }
            }
        });
    },
    
    // Clear form errors
    clearFormErrors(formId) {
        const form = document.getElementById(formId);
        if (!form) return;
        
        const errorDivs = form.querySelectorAll('.text-red-500');
        errorDivs.forEach(div => {
            div.textContent = '';
            div.style.display = 'none';
        });
        
        const errorFields = form.querySelectorAll('.border-red-500');
        errorFields.forEach(field => {
            field.classList.remove('border-red-500');
        });
    }
};