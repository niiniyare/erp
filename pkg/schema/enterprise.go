package schema

import (
	"context"
	"time"

	"github.com/niiniyare/erp/pkg/condition"
)

// Security defines security and access control configuration
type Security struct {
	CSRF          *CSRF       `json:"csrf,omitempty"`                        // CSRF protection
	RateLimit     *RateLimit  `json:"rateLimit,omitempty"`                   // Rate limiting
	Encryption    *Encryption `json:"encryption,omitempty"`                  // Field encryption
	Sanitization  bool        `json:"sanitization,omitempty"`                // HTML sanitization
	Origins       []string    `json:"origins,omitempty" validate:"dive,url"` // Allowed origins
	ContentType   []string    `json:"contentType,omitempty"`                 // Allowed content types
	MaxFileSize   int64       `json:"maxFileSize,omitempty"`                 // Max upload size in bytes
	AllowedExts   []string    `json:"allowedExts,omitempty"`                 // Allowed file extensions
	XSSProtection bool        `json:"xssProtection,omitempty"`               // XSS protection
	SQLInjection  bool        `json:"sqlInjection,omitempty"`                // SQL injection protection
}

// CSRF defines CSRF protection configuration
type CSRF struct {
	Enabled     bool   `json:"enabled"`                                                     // Enable CSRF protection
	FieldName   string `json:"fieldName,omitempty" example:"_csrf"`                         // Form field name
	HeaderName  string `json:"headerName,omitempty" example:"X-CSRF-Token"`                 // Header name
	CookieName  string `json:"cookieName,omitempty" example:"csrf_token"`                   // Cookie name
	TokenLength int    `json:"tokenLength,omitempty" validate:"min=16,max=64" example:"32"` // Token length
	Expiry      int    `json:"expiry,omitempty"`                                            // Token expiry in seconds
}

// RateLimit defines rate limiting configuration
type RateLimit struct {
	Enabled       bool          `json:"enabled"`                                    // Enable rate limiting
	MaxRequests   int           `json:"maxRequests" validate:"min=1" example:"100"` // Max requests
	Window        time.Duration `json:"window" example:"3600"`                      // Time window (seconds)
	ByIP          bool          `json:"byIp,omitempty"`                             // Rate limit by IP
	ByUser        bool          `json:"byUser,omitempty"`                           // Rate limit by user
	ByTenant      bool          `json:"byTenant,omitempty"`                         // Rate limit by tenant
	Burst         int           `json:"burst,omitempty"`                            // Burst capacity
	BlockDuration time.Duration `json:"blockDuration,omitempty"`                    // Block duration after limit
}

// Encryption defines field encryption configuration
type Encryption struct {
	Enabled     bool     `json:"enabled"`                          // Enable encryption
	Fields      []string `json:"fields" validate:"dive,fieldname"` // Fields to encrypt
	Algorithm   string   `json:"algorithm,omitempty" validate:"oneof=AES-256-GCM ChaCha20-Poly1305" example:"AES-256-GCM"`
	KeyRotation bool     `json:"keyRotation,omitempty"` // Enable key rotation
	KeyVersion  int      `json:"keyVersion,omitempty"`  // Current key version
}

// Tenant defines multi-tenancy configuration
type Tenant struct {
	Enabled    bool     `json:"enabled"`                                                          // Enable tenant isolation
	Field      string   `json:"field,omitempty" validate:"fieldname" example:"tenant_id"`         // Tenant ID field
	Isolation  string   `json:"isolation" validate:"oneof=strict shared hybrid" example:"strict"` // Isolation level
	Inherit    bool     `json:"inherit,omitempty"`                                                // Inherit from parent context
	AllowCross bool     `json:"allowCross,omitempty"`                                             // Allow cross-tenant access
	Whitelist  []string `json:"whitelist,omitempty" validate:"dive,uuid"`                         // Allowed tenant IDs
	Blacklist  []string `json:"blacklist,omitempty" validate:"dive,uuid"`                         // Blocked tenant IDs
	Validation bool     `json:"validation,omitempty"`                                             // Validate tenant access
}

// Workflow defines workflow and approval configuration
type Workflow struct {
	Enabled       bool                 `json:"enabled"`                                 // Enable workflow
	ID            string               `json:"id,omitempty" validate:"uuid"`            // Workflow ID
	Stage         string               `json:"stage,omitempty"`                         // Current stage
	Status        string               `json:"status,omitempty"`                        // Current status
	Actions       []WorkflowAction     `json:"actions,omitempty" validate:"dive"`       // Available actions
	Transitions   []WorkflowTransition `json:"transitions,omitempty" validate:"dive"`   // Stage transitions
	Notifications []Notification       `json:"notifications,omitempty" validate:"dive"` // Notifications
	Approvals     *ApprovalConfig      `json:"approvals,omitempty"`                     // Approval configuration
	History       bool                 `json:"history,omitempty"`                       // Track workflow history
	Timeout       *WorkflowTimeout     `json:"timeout,omitempty"`                       // Stage timeout
	Audit         bool                 `json:"audit,omitempty"`                         // Enable audit logging
	Versioning    bool                 `json:"versioning,omitempty"`                    // Enable versioning
}

// WorkflowAction defines an action that can be taken in the workflow
type WorkflowAction struct {
	ID          string                    `json:"id" validate:"required"`
	Label       string                    `json:"label" validate:"required"`
	Type        string                    `json:"type" validate:"oneof=approve reject submit return delegate escalate" example:"approve"`
	ToStage     string                    `json:"toStage,omitempty"`     // Next stage
	ToStatus    string                    `json:"toStatus,omitempty"`    // Next status
	Permissions []string                  `json:"permissions,omitempty"` // Required permissions
	RequireNote bool                      `json:"requireNote,omitempty"` // Require comment
	Icon        string                    `json:"icon,omitempty" validate:"icon_name"`
	Variant     string                    `json:"variant,omitempty" validate:"oneof=primary secondary outline destructive"`
	Condition   *condition.ConditionGroup `json:"condition,omitempty"` // Action visibility conditions
}

// WorkflowTransition defines a stage transition
type WorkflowTransition struct {
	From       string                    `json:"from" validate:"required"`   // From stage
	To         string                    `json:"to" validate:"required"`     // To stage
	Action     string                    `json:"action" validate:"required"` // Triggering action
	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`       // Transition conditions
	Validator  string                    `json:"validator,omitempty"`        // Custom validator
	Auto       bool                      `json:"auto,omitempty"`             // Automatic transition
	Delay      time.Duration             `json:"delay,omitempty"`            // Transition delay
}

// Notification defines a workflow notification
type Notification struct {
	Event    string   `json:"event" validate:"required" example:"approval_required"` // Trigger event
	Template string   `json:"template" validate:"required"`                          // Notification template
	To       []string `json:"to" validate:"required,dive"`                           // Recipient roles/users
	CC       []string `json:"cc,omitempty" validate:"dive"`                          // CC recipients
	BCC      []string `json:"bcc,omitempty" validate:"dive"`                         // BCC recipients
	Method   string   `json:"method" validate:"oneof=email sms push in-app webhook" example:"email"`
	Priority string   `json:"priority,omitempty" validate:"oneof=low normal high urgent"`
}

// ApprovalConfig defines approval requirements
type ApprovalConfig struct {
	Required        bool             `json:"required"`                      // Approval required
	MinApprovals    int              `json:"minApprovals" validate:"min=1"` // Min approvals needed
	Approvers       []ApproverConfig `json:"approvers" validate:"dive"`     // Approver configuration
	Sequential      bool             `json:"sequential,omitempty"`          // Sequential approval
	AllowDelegation bool             `json:"allowDelegation,omitempty"`     // Allow delegation
	Timeout         *time.Duration   `json:"timeout,omitempty"`             // Approval timeout
	EscalationTo    string           `json:"escalationTo,omitempty"`        // Escalation target
	RemindAfter     *time.Duration   `json:"remindAfter,omitempty"`         // Send reminder after
	AutoApprove     bool             `json:"autoApprove,omitempty"`         // Auto-approve on timeout
}

// ApproverConfig defines who can approve
type ApproverConfig struct {
	Type  string `json:"type" validate:"oneof=role user group dynamic manager" example:"role"`
	Value string `json:"value" validate:"required"` // Role/User/Group ID
	Order int    `json:"order,omitempty"`           // Approval order
	Level int    `json:"level,omitempty"`           // Approval level
}

// WorkflowTimeout defines timeout for workflow stages
type WorkflowTimeout struct {
	Duration time.Duration `json:"duration" validate:"min=1"` // Timeout duration
	Action   string        `json:"action" validate:"oneof=escalate auto-approve auto-reject notify"`
	NotifyAt []int         `json:"notifyAt,omitempty"` // Notify at % of timeout (e.g., [50, 75, 90])
}

// Validation defines cross-field validation rules
type Validation struct {
	Rules      []ValidationRule `json:"rules" validate:"dive"`                                                     // Validation rules
	Mode       string           `json:"mode,omitempty" validate:"oneof=onChange onBlur onSubmit" example:"onBlur"` // When to validate
	Async      bool             `json:"async,omitempty"`                                                           // Async validation
	DebounceMS int              `json:"debounceMs,omitempty"`                                                      // Debounce delay
}

// ValidationRule defines a cross-field validation rule
type ValidationRule struct {
	ID        string                    `json:"id" validate:"required"`
	Type      string                    `json:"type" validate:"required" example:"compare"` // Rule type
	Fields    []string                  `json:"fields" validate:"required,dive,fieldname"`  // Fields involved
	Condition *condition.ConditionGroup `json:"condition,omitempty"`                        // When to apply
	Message   string                    `json:"message" validate:"required"`                // Error message
	Validator string                    `json:"validator,omitempty"`                        // Custom validator function
	Severity  string                    `json:"severity,omitempty" validate:"oneof=error warning info" example:"error"`
	Priority  int                       `json:"priority,omitempty"` // Validation priority
}

// Events defines lifecycle event handlers
type Events struct {
	// Form lifecycle
	OnMount   string `json:"onMount,omitempty" validate:"js_function"`   // Form mounted
	OnUnmount string `json:"onUnmount,omitempty" validate:"js_function"` // Form unmounted

	// Submission lifecycle
	OnSubmit        string `json:"onSubmit,omitempty" validate:"js_function"`        // Submit triggered
	BeforeSubmit    string `json:"beforeSubmit,omitempty" validate:"js_function"`    // Before submit
	AfterSubmit     string `json:"afterSubmit,omitempty" validate:"js_function"`     // After submit
	OnSubmitSuccess string `json:"onSubmitSuccess,omitempty" validate:"js_function"` // Submit success
	OnSubmitError   string `json:"onSubmitError,omitempty" validate:"js_function"`   // Submit error

	// Form state
	OnReset    string `json:"onReset,omitempty" validate:"js_function"`    // Form reset
	OnValidate string `json:"onValidate,omitempty" validate:"js_function"` // Validation
	OnChange   string `json:"onChange,omitempty" validate:"js_function"`   // Any field changed
	OnDirty    string `json:"onDirty,omitempty" validate:"js_function"`    // Form becomes dirty
	OnPristine string `json:"onPristine,omitempty" validate:"js_function"` // Form becomes pristine

	// Data events
	OnLoad  string `json:"onLoad,omitempty" validate:"js_function"`  // Data loaded
	OnError string `json:"onError,omitempty" validate:"js_function"` // Error occurred

	// Field events
	OnFieldChange   string `json:"onFieldChange,omitempty" validate:"js_function"`   // Individual field changed
	OnFieldFocus    string `json:"onFieldFocus,omitempty" validate:"js_function"`    // Field focused
	OnFieldBlur     string `json:"onFieldBlur,omitempty" validate:"js_function"`     // Field blurred
	OnFieldValidate string `json:"onFieldValidate,omitempty" validate:"js_function"` // Field validated
}

// I18n defines internationalization configuration
type I18n struct {
	Enabled          bool              `json:"enabled"` // Enable i18n
	DefaultLocale    string            `json:"defaultLocale" validate:"locale" example:"en-US"`
	SupportedLocales []string          `json:"supportedLocales" validate:"dive,locale"`
	Translations     map[string]string `json:"translations,omitempty"` // Translation keys
	DateFormat       string            `json:"dateFormat,omitempty" example:"MM/DD/YYYY"`
	TimeFormat       string            `json:"timeFormat,omitempty" example:"HH:mm:ss"`
	NumberFormat     string            `json:"numberFormat,omitempty" example:"1,000.00"`
	CurrencyFormat   string            `json:"currencyFormat,omitempty" example:"$1,000.00"`
	Currency         string            `json:"currency,omitempty" validate:"iso4217" example:"USD"`
	Direction        string            `json:"direction,omitempty" validate:"oneof=ltr rtl" example:"ltr"`
	FallbackLocale   string            `json:"fallbackLocale,omitempty" validate:"locale"`
	LoadPath         string            `json:"loadPath,omitempty"` // Path to translation files
}

// HTMX defines HTMX integration for the entire form
type HTMX struct {
	Enabled   bool              `json:"enabled"` // Enable HTMX
	Get       string            `json:"get,omitempty" validate:"url"`
	Post      string            `json:"post,omitempty" validate:"url"`
	Put       string            `json:"put,omitempty" validate:"url"`
	Delete    string            `json:"delete,omitempty" validate:"url"`
	Patch     string            `json:"patch,omitempty" validate:"url"`
	Trigger   string            `json:"trigger,omitempty"`                        // Trigger event
	Target    string            `json:"target,omitempty" validate:"css_selector"` // Update target
	Swap      string            `json:"swap,omitempty" validate:"oneof=innerHTML outerHTML beforebegin afterbegin beforeend afterend delete none"`
	Select    string            `json:"select,omitempty" validate:"css_selector"`    // Response selector
	Indicator string            `json:"indicator,omitempty" validate:"css_selector"` // Loading indicator
	PushURL   string            `json:"pushUrl,omitempty" validate:"url"`            // Push to history
	Headers   map[string]string `json:"headers,omitempty"`                           // Request headers
	Vals      string            `json:"vals,omitempty"`                              // Additional values
	Confirm   string            `json:"confirm,omitempty"`                           // Confirmation prompt
	Boost     bool              `json:"boost,omitempty"`                             // Boost links
	Sync      string            `json:"sync,omitempty"`                              // Sync specification
	Validate  bool              `json:"validate,omitempty"`                          // Validate before request
	Timeout   int               `json:"timeout,omitempty"`                           // Request timeout in ms
	Retry     int               `json:"retry,omitempty"`                             // Number of retries
}

// Alpine defines Alpine.js integration for the entire form
type Alpine struct {
	Enabled     bool   `json:"enabled"`                                     // Enable Alpine.js
	XData       string `json:"xData,omitempty" validate:"js_object"`        // Component data
	XInit       string `json:"xInit,omitempty" validate:"js_function"`      // Initialization
	XShow       string `json:"xShow,omitempty" validate:"js_expression"`    // Show/hide
	XIf         string `json:"xIf,omitempty" validate:"js_expression"`      // Conditional
	XModel      string `json:"xModel,omitempty" validate:"js_variable"`     // Two-way binding
	XBind       string `json:"xBind,omitempty" validate:"js_object"`        // Attribute bindings
	XOn         string `json:"xOn,omitempty" validate:"js_object"`          // Event handlers
	XText       string `json:"xText,omitempty" validate:"js_expression"`    // Text content
	XHTML       string `json:"xHtml,omitempty" validate:"js_expression"`    // HTML content
	XRef        string `json:"xRef,omitempty" validate:"js_variable"`       // Reference
	XCloak      bool   `json:"xCloak,omitempty"`                            // Cloak until ready
	XTransition string `json:"xTransition,omitempty"`                       // Transition effect
	XTeleport   string `json:"xTeleport,omitempty" validate:"css_selector"` // Teleport target
	XFor        string `json:"xFor,omitempty" validate:"js_expression"`     // Loop directive
	XEffect     string `json:"xEffect,omitempty" validate:"js_function"`    // Side effects
}

// Meta contains metadata and tracking information
type Meta struct {
	CreatedAt     time.Time        `json:"createdAt,omitempty"`                                        // Creation timestamp
	UpdatedAt     time.Time        `json:"updatedAt,omitempty"`                                        // Last update timestamp
	CreatedBy     string           `json:"createdBy,omitempty" validate:"uuid"`                        // Creator user ID
	UpdatedBy     string           `json:"updatedBy,omitempty" validate:"uuid"`                        // Last updater user ID
	Deprecated    bool             `json:"deprecated,omitempty"`                                       // Schema is deprecated
	Experimental  bool             `json:"experimental,omitempty"`                                     // Experimental feature
	Changelog     []ChangelogEntry `json:"changelog,omitempty" validate:"dive"`                        // Version history
	CustomData    map[string]any   `json:"customData,omitempty"`                                       // Custom metadata
	Author        string           `json:"author,omitempty"`                                           // Schema author
	License       string           `json:"license,omitempty"`                                          // License information
	Repository    string           `json:"repository,omitempty" validate:"url"`                        // Source repository
	Documentation string           `json:"documentation,omitempty" validate:"url"`                     // Documentation URL
	Status        string           `json:"status,omitempty" validate:"oneof=draft published archived"` // Schema status
	Theme         *ThemeConfig     `json:"theme,omitempty"`                                            // Theme configuration
}

// ChangelogEntry represents a version change
type ChangelogEntry struct {
	Version     string    `json:"version" validate:"semver"`
	Date        time.Time `json:"date"`
	Author      string    `json:"author,omitempty"`
	Description string    `json:"description"`
	Breaking    bool      `json:"breaking,omitempty"`  // Breaking change
	Changes     []string  `json:"changes,omitempty"`   // List of changes
	Migration   string    `json:"migration,omitempty"` // Migration guide
}

// ThemeConfig stores theme configuration for a schema
type ThemeConfig struct {
	ID        string          `json:"id,omitempty"`        // Theme ID to use
	Overrides *ThemeOverrides `json:"overrides,omitempty"` // Theme overrides
}

// Helper methods for enterprise features

// IsCSRFEnabled checks if CSRF protection is enabled
func (s *Security) IsCSRFEnabled() bool {
	return s.CSRF != nil && s.CSRF.Enabled
}

// IsRateLimitEnabled checks if rate limiting is enabled
func (s *Security) IsRateLimitEnabled() bool {
	return s.RateLimit != nil && s.RateLimit.Enabled
}

// IsEncryptionEnabled checks if encryption is enabled
func (s *Security) IsEncryptionEnabled() bool {
	return s.Encryption != nil && s.Encryption.Enabled
}

// ShouldEncryptField checks if a specific field should be encrypted
func (s *Security) ShouldEncryptField(fieldName string) bool {
	if !s.IsEncryptionEnabled() {
		return false
	}
	for _, field := range s.Encryption.Fields {
		if field == fieldName {
			return true
		}
	}
	return false
}

// IsTenantEnabled checks if multi-tenancy is enabled
func (t *Tenant) IsTenantEnabled() bool {
	return t != nil && t.Enabled
}

// IsWorkflowEnabled checks if workflow is enabled
func (w *Workflow) IsWorkflowEnabled() bool {
	return w != nil && w.Enabled
}

// GetCurrentStage returns the current workflow stage
func (w *Workflow) GetCurrentStage() string {
	if w == nil {
		return ""
	}
	return w.Stage
}

// GetCurrentStatus returns the current workflow status
func (w *Workflow) GetCurrentStatus() string {
	if w == nil {
		return ""
	}
	return w.Status
}

// GetAvailableActions returns actions available in current stage
func (w *Workflow) GetAvailableActions(ctx context.Context, data map[string]any, evaluator *condition.Evaluator) []WorkflowAction {
	if w == nil || !w.Enabled {
		return []WorkflowAction{}
	}

	available := make([]WorkflowAction, 0)
	for _, action := range w.Actions {
		// Check if action has conditions
		if action.Condition != nil && evaluator != nil {
			evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
			result, err := evaluator.Evaluate(ctx, action.Condition, evalCtx)
			if err != nil || !result {
				continue
			}
		}
		available = append(available, action)
	}

	return available
}

// IsI18nEnabled checks if internationalization is enabled
func (i *I18n) IsI18nEnabled() bool {
	return i != nil && i.Enabled
}

// GetLocale returns the default locale or fallback
func (i *I18n) GetLocale() string {
	if i == nil {
		return "en-US"
	}
	if i.DefaultLocale != "" {
		return i.DefaultLocale
	}
	return "en-US"
}

// IsHTMXEnabled checks if HTMX is enabled
func (h *HTMX) IsHTMXEnabled() bool {
	return h != nil && h.Enabled
}

// IsAlpineEnabled checks if Alpine.js is enabled
func (a *Alpine) IsAlpineEnabled() bool {
	return a != nil && a.Enabled
}

// GetHTTPMethod returns the appropriate HTTP method for HTMX
func (h *HTMX) GetHTTPMethod() string {
	if h == nil {
		return "GET"
	}
	if h.Post != "" {
		return "POST"
	}
	if h.Put != "" {
		return "PUT"
	}
	if h.Patch != "" {
		return "PATCH"
	}
	if h.Delete != "" {
		return "DELETE"
	}
	return "GET"
}

// GetURL returns the appropriate URL for HTMX based on method
func (h *HTMX) GetURL() string {
	if h == nil {
		return ""
	}
	switch h.GetHTTPMethod() {
	case "POST":
		return h.Post
	case "PUT":
		return h.Put
	case "PATCH":
		return h.Patch
	case "DELETE":
		return h.Delete
	default:
		return h.Get
	}
}

// AddChangelog adds a new changelog entry
func (m *Meta) AddChangelog(entry ChangelogEntry) {
	if m.Changelog == nil {
		m.Changelog = []ChangelogEntry{}
	}
	m.Changelog = append(m.Changelog, entry)
}

// GetLatestVersion returns the latest version from changelog
func (m *Meta) GetLatestVersion() string {
	if m == nil || len(m.Changelog) == 0 {
		return ""
	}
	return m.Changelog[len(m.Changelog)-1].Version
}

// IsDeprecated checks if the schema is deprecated
func (m *Meta) IsDeprecated() bool {
	return m != nil && m.Deprecated
}

// IsExperimental checks if the schema is experimental
func (m *Meta) IsExperimental() bool {
	return m != nil && m.Experimental
}
