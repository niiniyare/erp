package sidebar

import (
	"github.com/niiniyare/erp/web/components/atoms"
)

// SidebarItem represents a navigation item in the sidebar
type SidebarItem struct {
	// Content
	Text        string `json:"text"`
	Icon        string `json:"icon,omitempty"`
	Badge       string `json:"badge,omitempty"`
	BadgeColor  string `json:"badgeColor,omitempty"`
	Description string `json:"description,omitempty"`

	// Navigation
	Href     string `json:"href,omitempty"`
	Target   string `json:"target,omitempty"`
	Active   bool   `json:"active"`
	Disabled bool   `json:"disabled"`

	// Hierarchy
	Children []SidebarItem `json:"children,omitempty"`
	Expanded bool          `json:"expanded,omitempty"`
	Level    int           `json:"level,omitempty"` // Nesting level

	// Interaction
	OnClick     string `json:"onclick,omitempty"`
	Collapsible bool   `json:"collapsible,omitempty"` // Can be collapsed

	// HTMX attributes
	HxGet    string `json:"hxGet,omitempty"`
	HxPost   string `json:"hxPost,omitempty"`
	HxTarget string `json:"hxTarget,omitempty"`
	HxSwap   string `json:"hxSwap,omitempty"`

	// Styling
	Type      string `json:"type,omitempty"` // "item", "header", "divider"
	Highlight bool   `json:"highlight,omitempty"`

	// HTML attributes
	ID    string `json:"id,omitempty"`
	Class string `json:"class,omitempty"`
}

// SidebarBrand defines the brand/logo area in sidebar
type SidebarBrand struct {
	Text      string `json:"text,omitempty"`
	Logo      string `json:"logo,omitempty"`
	LogoAlt   string `json:"logoAlt,omitempty"`
	LogoSmall string `json:"logoSmall,omitempty"` // Collapsed state logo
	Href      string `json:"href,omitempty"`
	ShowText  bool   `json:"showText"`
	ID        string `json:"id,omitempty"`
}

// SidebarUser represents user information in footer
type SidebarUser struct {
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Role     string `json:"role,omitempty"`
	Status   string `json:"status,omitempty"`
	Initials string `json:"initials,omitempty"`
	OnClick  string `json:"onclick,omitempty"`
	Href     string `json:"href,omitempty"`
}

// SidebarAction represents an action button in footer
type SidebarAction struct {
	Text    string        `json:"text,omitempty"`
	Icon    string        `json:"icon"`
	Variant atoms.Variant `json:"variant"`
	Size    atoms.Size    `json:"size"`
	Tooltip string        `json:"tooltip,omitempty"`
	OnClick string        `json:"onclick,omitempty"`

	// HTMX attributes
	HxPost   string `json:"hxPost,omitempty"`
	HxGet    string `json:"hxGet,omitempty"`
	HxTarget string `json:"hxTarget,omitempty"`

	ID string `json:"id,omitempty"`
}

// SidebarFooter defines the footer content
type SidebarFooter struct {
	// User information
	User *SidebarUser `json:"user,omitempty"`

	// Footer items
	Items []SidebarItem `json:"items,omitempty"`

	// Actions
	Actions []SidebarAction `json:"actions,omitempty"`

	// Content
	Text      string `json:"text,omitempty"`
	Copyright string `json:"copyright,omitempty"`

	// Layout
	Compact bool `json:"compact,omitempty"`
}

// SidebarProps defines properties for the Sidebar organism
type SidebarProps struct {
	// Navigation structure
	Items []SidebarItem `json:"items"`

	// Branding
	Brand *SidebarBrand `json:"brand,omitempty"`

	// Footer content
	Footer *SidebarFooter `json:"footer,omitempty"`

	// Layout configuration
	Collapsible bool   `json:"collapsible"`
	Collapsed   bool   `json:"collapsed"`
	Position    string `json:"position,omitempty"` // "left", "right"
	Overlay     bool   `json:"overlay,omitempty"`  // Mobile overlay mode
	Width       string `json:"width,omitempty"`    // Custom width

	// Styling
	Variant string `json:"variant,omitempty"` // "default", "minimal", "bordered"
	Theme   string `json:"theme,omitempty"`   // "light", "dark", "auto"

	// HTML attributes
	ID         string `json:"id,omitempty"`
	Class      string `json:"class,omitempty"`
	DataTestID string `json:"dataTestId,omitempty"`

	// Accessibility
	AriaLabel string `json:"ariaLabel,omitempty"`
}

// SidebarGroup represents a group of related items
type SidebarGroup struct {
	Title       string        `json:"title"`
	Items       []SidebarItem `json:"items"`
	Collapsible bool          `json:"collapsible,omitempty"`
	Expanded    bool          `json:"expanded,omitempty"`
	ID          string        `json:"id,omitempty"`
}
