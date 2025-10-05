package siteheader

import (
	"github.com/niiniyare/erp/web/components/atoms"
	"github.com/niiniyare/erp/web/components/molecules"
)

// HeaderBrand defines the brand/logo area
type HeaderBrand struct {
	Text     string `json:"text,omitempty"`
	Logo     string `json:"logo,omitempty"`
	LogoAlt  string `json:"logoAlt,omitempty"`
	LogoDark string `json:"logoDark,omitempty"` // Dark mode logo
	Href     string `json:"href,omitempty"`
	Target   string `json:"target,omitempty"`
	ShowText bool   `json:"showText"`
	ID       string `json:"id,omitempty"`
}

// NavItem represents a navigation item
type NavItem struct {
	Text       string    `json:"text"`
	Href       string    `json:"href,omitempty"`
	Icon       string    `json:"icon,omitempty"`
	Active     bool      `json:"active"`
	Disabled   bool      `json:"disabled"`
	Badge      string    `json:"badge,omitempty"`
	BadgeColor string    `json:"badgeColor,omitempty"`
	Children   []NavItem `json:"children,omitempty"` // For dropdown menus
	External   bool      `json:"external,omitempty"`
	OnClick    string    `json:"onclick,omitempty"`

	// HTMX attributes
	HxGet    string `json:"hxGet,omitempty"`
	HxPost   string `json:"hxPost,omitempty"`
	HxTarget string `json:"hxTarget,omitempty"`

	ID string `json:"id,omitempty"`
}

// HeaderSearch defines search configuration
type HeaderSearch struct {
	Enabled      bool                   `json:"enabled"`
	Placeholder  string                 `json:"placeholder,omitempty"`
	SearchProps  *molecules.SearchProps `json:"searchProps,omitempty"`
	Shortcut     string                 `json:"shortcut,omitempty"` // Keyboard shortcut (e.g., "⌘K")
	MobileHidden bool                   `json:"mobileHidden,omitempty"`
}

// HeaderUser represents user information
type HeaderUser struct {
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Status   string `json:"status,omitempty"` // "online", "away", "busy", "offline"
	Role     string `json:"role,omitempty"`
	Initials string `json:"initials,omitempty"`
}

// HeaderMenuItem represents a user menu item
type HeaderMenuItem struct {
	Text    string `json:"text"`
	Href    string `json:"href,omitempty"`
	Icon    string `json:"icon,omitempty"`
	Divider bool   `json:"divider,omitempty"`
	OnClick string `json:"onclick,omitempty"`

	// HTMX attributes
	HxPost    string `json:"hxPost,omitempty"`
	HxGet     string `json:"hxGet,omitempty"`
	HxConfirm string `json:"hxConfirm,omitempty"`

	ID string `json:"id,omitempty"`
}

// HeaderUserMenu defines user menu configuration
type HeaderUserMenu struct {
	User       HeaderUser       `json:"user"`
	MenuItems  []HeaderMenuItem `json:"menuItems,omitempty"`
	ShowAvatar bool             `json:"showAvatar"`
	ShowName   bool             `json:"showName"`
	ShowStatus bool             `json:"showStatus"`
	Position   string           `json:"position,omitempty"` // "left", "right"
}

// HeaderAction defines action buttons in the header
type HeaderAction struct {
	Text       string              `json:"text,omitempty"`
	Icon       string              `json:"icon"`
	Variant    atoms.ButtonVariant `json:"variant"`
	Size       atoms.ButtonSize    `json:"size"`
	Badge      string              `json:"badge,omitempty"`
	BadgeColor string              `json:"badgeColor,omitempty"`
	Tooltip    string              `json:"tooltip,omitempty"`
	OnClick    string              `json:"onclick,omitempty"`

	// HTMX attributes
	HxPost   string `json:"hxPost,omitempty"`
	HxGet    string `json:"hxGet,omitempty"`
	HxTarget string `json:"hxTarget,omitempty"`

	// Visibility
	MobileHidden  bool `json:"mobileHidden,omitempty"`
	DesktopHidden bool `json:"desktopHidden,omitempty"`

	ID string `json:"id,omitempty"`
}

// HeaderNotifications defines notification bell configuration
type HeaderNotifications struct {
	Enabled  bool   `json:"enabled"`
	Count    int    `json:"count,omitempty"`
	MaxCount int    `json:"maxCount,omitempty"` // Max count to display (e.g., 99+)
	Href     string `json:"href,omitempty"`
	OnClick  string `json:"onclick,omitempty"`

	// HTMX for dynamic updates
	HxGet     string `json:"hxGet,omitempty"`
	HxTarget  string `json:"hxTarget,omitempty"`
	HxTrigger string `json:"hxTrigger,omitempty"`
}

// SiteHeaderProps defines properties for the SiteHeader organism
type SiteHeaderProps struct {
	// Brand configuration
	Brand HeaderBrand `json:"brand"`

	// Navigation
	Navigation []NavItem `json:"navigation,omitempty"`

	// Search configuration
	Search *HeaderSearch `json:"search,omitempty"`

	// User menu
	UserMenu *HeaderUserMenu `json:"userMenu,omitempty"`

	// Quick actions
	Actions []HeaderAction `json:"actions,omitempty"`

	// Notifications
	Notifications *HeaderNotifications `json:"notifications,omitempty"`

	// Layout configuration
	Sticky      bool `json:"sticky"`
	Transparent bool `json:"transparent"`
	Border      bool `json:"border"`
	Shadow      bool `json:"shadow"`

	// HTML attributes
	ID         string `json:"id,omitempty"`
	Class      string `json:"class,omitempty"`
	DataTestID string `json:"dataTestId,omitempty"`
}
