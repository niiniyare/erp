package modal

import (
	"github.com/niiniyare/erp/web/components/atoms"
)

// ModalSize defines the size variants for modals
type ModalSize string

const (
	ModalSizeSM  ModalSize = "sm"  // Small: 320px
	ModalSizeMD  ModalSize = "md"  // Medium: 512px (default)
	ModalSizeLG  ModalSize = "lg"  // Large: 640px
	ModalSizeXL  ModalSize = "xl"  // Extra Large: 768px
	ModalSize2XL ModalSize = "2xl" // 2X Large: 896px
	ModalSizeFull ModalSize = "full" // Full screen
)

// ModalType defines the modal behavior
type ModalType string

const (
	ModalTypeDefault ModalType = "default" // Standard modal
	ModalTypeForm    ModalType = "form"    // Form modal with validation
	ModalTypeConfirm ModalType = "confirm" // Confirmation dialog
	ModalTypeAlert   ModalType = "alert"   // Alert/info modal
)

// ModalAction represents an action button in the modal
type ModalAction struct {
	Text        string               `json:"text"`
	Icon        string               `json:"icon,omitempty"`
	Variant     atoms.ButtonVariant  `json:"variant"`
	Size        atoms.ButtonSize     `json:"size"`
	OnClick     string               `json:"onclick,omitempty"`
	
	// HTMX attributes
	HxPost      string               `json:"hxPost,omitempty"`
	HxGet       string               `json:"hxGet,omitempty"`
	HxTarget    string               `json:"hxTarget,omitempty"`
	HxSwap      string               `json:"hxSwap,omitempty"`
	HxConfirm   string               `json:"hxConfirm,omitempty"`
	
	// Form attributes
	Type        string               `json:"type,omitempty"`        // submit, button, reset
	Form        string               `json:"form,omitempty"`        // Form ID to submit
	Disabled    bool                 `json:"disabled,omitempty"`
	
	// Styling
	Position    string               `json:"position,omitempty"`    // left, right
	AutoFocus   bool                 `json:"autoFocus,omitempty"`
	
	ID          string               `json:"id,omitempty"`
	Class       string               `json:"class,omitempty"`
}

// ModalHeader defines the modal header configuration
type ModalHeader struct {
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Closable    bool   `json:"closable"`             // Show close button
	Level       int    `json:"level,omitempty"`      // Heading level (1-6)
	
	// Custom content
	CustomContent string `json:"customContent,omitempty"`
}

// ModalBody defines the modal body configuration
type ModalBody struct {
	Content     string `json:"content,omitempty"`
	HTML        string `json:"html,omitempty"`        // Raw HTML content
	Scrollable  bool   `json:"scrollable"`            // Enable scrolling
	Padding     string `json:"padding,omitempty"`     // Custom padding classes
	
	// Form integration
	FormID      string `json:"formId,omitempty"`      // Form element ID
	
	// Loading state
	Loading     bool   `json:"loading,omitempty"`
	LoadingText string `json:"loadingText,omitempty"`
}

// ModalFooter defines the modal footer configuration
type ModalFooter struct {
	Actions     []ModalAction `json:"actions,omitempty"`
	Alignment   string        `json:"alignment,omitempty"`   // left, center, right, between
	Compact     bool          `json:"compact,omitempty"`     // Reduce spacing
	
	// Custom content
	CustomContent string      `json:"customContent,omitempty"`
}

// ModalProps defines properties for the Modal organism
type ModalProps struct {
	// Core configuration
	ID          string     `json:"id"`
	Size        ModalSize  `json:"size"`
	Type        ModalType  `json:"type"`
	Open        bool       `json:"open"`
	
	// Content sections
	Header      *ModalHeader `json:"header,omitempty"`
	Body        ModalBody    `json:"body"`
	Footer      *ModalFooter `json:"footer,omitempty"`
	
	// Behavior
	Closable    bool   `json:"closable"`              // Can be closed by user
	CloseOnBackdrop bool `json:"closeOnBackdrop"`     // Close on backdrop click
	CloseOnEscape   bool `json:"closeOnEscape"`       // Close on ESC key
	FocusTrap       bool `json:"focusTrap"`           // Trap focus within modal
	
	// HTMX integration
	HxGet       string `json:"hxGet,omitempty"`       // Load content dynamically
	HxPost      string `json:"hxPost,omitempty"`      // Submit form via HTMX
	HxTarget    string `json:"hxTarget,omitempty"`    // Target for HTMX response
	HxSwap      string `json:"hxSwap,omitempty"`      // HTMX swap strategy
	HxTrigger   string `json:"hxTrigger,omitempty"`   // HTMX trigger events
	
	// Callbacks
	OnOpen      string `json:"onOpen,omitempty"`      // Alpine.js callback
	OnClose     string `json:"onClose,omitempty"`     // Alpine.js callback
	OnConfirm   string `json:"onConfirm,omitempty"`   // Confirmation callback
	
	// Styling
	Backdrop    string `json:"backdrop,omitempty"`    // Backdrop styling
	Position    string `json:"position,omitempty"`    // center, top, bottom
	Animation   string `json:"animation,omitempty"`   // Enter/leave animation
	ZIndex      string `json:"zIndex,omitempty"`      // Custom z-index
	
	// Accessibility
	AriaLabel   string `json:"ariaLabel,omitempty"`
	AriaDescribedBy string `json:"ariaDescribedBy,omitempty"`
	Role        string `json:"role,omitempty"`        // dialog, alertdialog
	
	// HTML attributes
	Class       string `json:"class,omitempty"`
	DataTestID  string `json:"dataTestId,omitempty"`
}

// ConfirmModalProps defines a simplified interface for confirmation dialogs
type ConfirmModalProps struct {
	Title       string `json:"title"`
	Message     string `json:"message"`
	ConfirmText string `json:"confirmText"`
	CancelText  string `json:"cancelText"`
	OnConfirm   string `json:"onConfirm"`
	OnCancel    string `json:"onCancel,omitempty"`
	Variant     string `json:"variant,omitempty"`     // danger, warning, info
	Icon        string `json:"icon,omitempty"`
}

// FormModalProps defines a simplified interface for form modals
type FormModalProps struct {
	Title       string        `json:"title"`
	FormID      string        `json:"formId"`
	FormAction  string        `json:"formAction"`
	FormMethod  string        `json:"formMethod"`
	SubmitText  string        `json:"submitText"`
	CancelText  string        `json:"cancelText"`
	Size        ModalSize     `json:"size"`
	
	// HTMX form handling
	HxPost      string        `json:"hxPost,omitempty"`
	HxTarget    string        `json:"hxTarget,omitempty"`
	HxSwap      string        `json:"hxSwap,omitempty"`
	
	// Validation
	Validate    bool          `json:"validate"`
	
	// Content
	Content     string        `json:"content,omitempty"`
}