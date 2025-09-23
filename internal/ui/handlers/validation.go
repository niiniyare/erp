package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/niiniyare/erp/internal/core/iam"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ValidationHandler handles real-time form validation requests
type ValidationHandler struct {
	iamService iam.Service
}

// NewValidationHandler creates a new validation handler
func NewValidationHandler(iamService iam.Service) *ValidationHandler {
	return &ValidationHandler{
		iamService: iamService,
	}
}

// ValidationResponse represents the response from validation endpoints
type ValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// ValidateEmail validates email format and uniqueness
func (h *ValidationHandler) ValidateEmail(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	fieldName := r.FormValue("field")

	if fieldName == "" {
		fieldName = "email"
	}

	// Basic format validation
	if email == "" {
		h.renderValidationError(w, "Email is required")
		return
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		h.renderValidationError(w, "Please enter a valid email address")
		return
	}

	// Check for uniqueness (if this is a new user registration)
	ctx := r.Context()

	// Skip uniqueness check if we're editing an existing user
	userID := r.FormValue("user_id")
	if userID == "" {
		exists, err := h.checkEmailExists(ctx, email)
		if err != nil {
			h.renderValidationError(w, "Unable to validate email")
			return
		}

		if exists {
			h.renderValidationError(w, "This email address is already in use")
			return
		}
	}

	h.renderValidationSuccess(w, "")
}

// ValidateUsername validates username format and uniqueness
func (h *ValidationHandler) ValidateUsername(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))

	// Basic format validation
	if username == "" {
		h.renderValidationError(w, "Username is required")
		return
	}

	if len(username) < 3 {
		h.renderValidationError(w, "Username must be at least 3 characters long")
		return
	}

	if len(username) > 50 {
		h.renderValidationError(w, "Username must be less than 50 characters")
		return
	}

	// Username format validation (alphanumeric, underscore, hyphen)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		h.renderValidationError(w, "Username can only contain letters, numbers, underscores, and hyphens")
		return
	}

	// Check for uniqueness
	ctx := r.Context()
	userID := r.FormValue("user_id")

	if userID == "" {
		exists, err := h.checkUsernameExists(ctx, username)
		if err != nil {
			h.renderValidationError(w, "Unable to validate username")
			return
		}

		if exists {
			h.renderValidationError(w, "This username is already taken")
			return
		}
	}

	h.renderValidationSuccess(w, "")
}

// ValidatePassword validates password strength
func (h *ValidationHandler) ValidatePassword(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")

	if password == "" {
		h.renderValidationError(w, "Password is required")
		return
	}

	// Password strength validation
	if len(password) < 8 {
		h.renderValidationError(w, "Password must be at least 8 characters long")
		return
	}

	if len(password) > 128 {
		h.renderValidationError(w, "Password must be less than 128 characters")
		return
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		h.renderValidationError(w, "Password must contain at least one uppercase letter")
		return
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		h.renderValidationError(w, "Password must contain at least one lowercase letter")
		return
	}

	// Check for at least one number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		h.renderValidationError(w, "Password must contain at least one number")
		return
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		h.renderValidationError(w, "Password must contain at least one special character")
		return
	}

	h.renderValidationSuccess(w, "Strong password")
}

// ValidatePasswordConfirm validates password confirmation
func (h *ValidationHandler) ValidatePasswordConfirm(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	confirmPassword := r.FormValue("password_confirm")

	if confirmPassword == "" {
		h.renderValidationError(w, "Please confirm your password")
		return
	}

	if password != confirmPassword {
		h.renderValidationError(w, "Passwords do not match")
		return
	}

	h.renderValidationSuccess(w, "Passwords match")
}

// ValidateTenantName validates tenant name format and uniqueness
func (h *ValidationHandler) ValidateTenantName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))

	if name == "" {
		h.renderValidationError(w, "Tenant name is required")
		return
	}

	if len(name) < 2 {
		h.renderValidationError(w, "Tenant name must be at least 2 characters long")
		return
	}

	if len(name) > 100 {
		h.renderValidationError(w, "Tenant name must be less than 100 characters")
		return
	}

	// Check for uniqueness
	ctx := r.Context()
	tenantID := r.FormValue("tenant_id")

	if tenantID == "" {
		exists, err := h.checkTenantNameExists(ctx, name)
		if err != nil {
			h.renderValidationError(w, "Unable to validate tenant name")
			return
		}

		if exists {
			h.renderValidationError(w, "This tenant name is already taken")
			return
		}
	}

	h.renderValidationSuccess(w, "")
}

// ValidateSubdomain validates subdomain format and uniqueness
func (h *ValidationHandler) ValidateSubdomain(w http.ResponseWriter, r *http.Request) {
	subdomain := strings.TrimSpace(strings.ToLower(r.FormValue("subdomain")))

	if subdomain == "" {
		h.renderValidationError(w, "Subdomain is required")
		return
	}

	if len(subdomain) < 3 {
		h.renderValidationError(w, "Subdomain must be at least 3 characters long")
		return
	}

	if len(subdomain) > 50 {
		h.renderValidationError(w, "Subdomain must be less than 50 characters")
		return
	}

	// Subdomain format validation
	subdomainRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !subdomainRegex.MatchString(subdomain) {
		h.renderValidationError(w, "Subdomain can only contain lowercase letters, numbers, and hyphens")
		return
	}

	// Cannot start or end with hyphen
	if strings.HasPrefix(subdomain, "-") || strings.HasSuffix(subdomain, "-") {
		h.renderValidationError(w, "Subdomain cannot start or end with a hyphen")
		return
	}

	// Reserved subdomains
	reserved := []string{"www", "api", "admin", "app", "mail", "ftp", "support", "help", "docs", "blog"}
	for _, res := range reserved {
		if subdomain == res {
			h.renderValidationError(w, "This subdomain is reserved")
			return
		}
	}

	// Check for uniqueness
	ctx := r.Context()
	tenantID := r.FormValue("tenant_id")

	if tenantID == "" {
		exists, err := h.checkSubdomainExists(ctx, subdomain)
		if err != nil {
			h.renderValidationError(w, "Unable to validate subdomain")
			return
		}

		if exists {
			h.renderValidationError(w, "This subdomain is already taken")
			return
		}
	}

	h.renderValidationSuccess(w, "Subdomain is available")
}

// ValidateRequired validates that a required field is not empty
func (h *ValidationHandler) ValidateRequired(w http.ResponseWriter, r *http.Request) {
	fieldName := r.FormValue("field")
	value := strings.TrimSpace(r.FormValue("value"))

	if value == "" {
		fieldLabel := h.getFieldLabel(fieldName)
		h.renderValidationError(w, fieldLabel+" is required")
		return
	}

	h.renderValidationSuccess(w, "")
}

// ValidateLength validates field length constraints
func (h *ValidationHandler) ValidateLength(w http.ResponseWriter, r *http.Request) {
	fieldName := r.FormValue("field")
	value := r.FormValue("value")
	minLength := r.FormValue("min")
	maxLength := r.FormValue("max")

	fieldLabel := h.getFieldLabel(fieldName)

	if minLength != "" {
		var min int
		if err := json.Unmarshal([]byte(minLength), &min); err == nil {
			if len(value) < min {
				h.renderValidationError(w, fieldLabel+" must be at least "+minLength+" characters long")
				return
			}
		}
	}

	if maxLength != "" {
		var max int
		if err := json.Unmarshal([]byte(maxLength), &max); err == nil {
			if len(value) > max {
				h.renderValidationError(w, fieldLabel+" must be less than "+maxLength+" characters")
				return
			}
		}
	}

	h.renderValidationSuccess(w, "")
}

// Helper methods

func (h *ValidationHandler) renderValidationError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="text-red-500 text-sm" role="alert" aria-live="polite">` + message + `</div>`))
}

func (h *ValidationHandler) renderValidationSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if message != "" {
		w.Write([]byte(`<div class="text-green-500 text-sm">` + message + `</div>`))
	} else {
		w.Write([]byte(""))
	}
}

func (h *ValidationHandler) checkEmailExists(ctx context.Context, email string) (bool, error) {
	// Implementation would check against the user repository
	// For now, return false (email is available)
	return false, nil
}

func (h *ValidationHandler) checkUsernameExists(ctx context.Context, username string) (bool, error) {
	// Implementation would check against the user repository
	// For now, return false (username is available)
	return false, nil
}

func (h *ValidationHandler) checkTenantNameExists(ctx context.Context, name string) (bool, error) {
	// Implementation would check against the tenant repository
	// For now, return false (tenant name is available)
	return false, nil
}

func (h *ValidationHandler) checkSubdomainExists(ctx context.Context, subdomain string) (bool, error) {
	// Implementation would check against the tenant repository
	// For now, return false (subdomain is available)
	return false, nil
}

func (h *ValidationHandler) getFieldLabel(fieldName string) string {
	labels := map[string]string{
		"name":             "Name",
		"email":            "Email",
		"username":         "Username",
		"password":         "Password",
		"password_confirm": "Password confirmation",
		"subdomain":        "Subdomain",
		"company_name":     "Company name",
		"industry":         "Industry",
		"company_size":     "Company size",
		"phone":            "Phone number",
		"address":          "Address",
		"description":      "Description",
	}

	if label, exists := labels[fieldName]; exists {
		return label
	}

	// Convert snake_case to Title Case using cases.Title
	parts := strings.Split(fieldName, "_")
	for i, part := range parts {
		parts[i] = cases.Title(language.English).String(part)
	}
	return strings.Join(parts, " ")
}

// RouteValidation sets up validation routes
func (h *ValidationHandler) RouteValidation(mux *http.ServeMux) {
	mux.HandleFunc("/validate/email", h.ValidateEmail)
	mux.HandleFunc("/validate/username", h.ValidateUsername)
	mux.HandleFunc("/validate/password", h.ValidatePassword)
	mux.HandleFunc("/validate/password-confirm", h.ValidatePasswordConfirm)
	mux.HandleFunc("/validate/tenant-name", h.ValidateTenantName)
	mux.HandleFunc("/validate/subdomain", h.ValidateSubdomain)
	mux.HandleFunc("/validate/required", h.ValidateRequired)
	mux.HandleFunc("/validate/length", h.ValidateLength)
}
