package types

import (
	"time"

	"github.com/a-h/templ"
)

// BaseLayoutData represents data for the base layout template
type BaseLayoutData struct {
	Title     string
	User      UserInfo
	CSRFToken string
	Content   templ.Component
}

// LoginPageData represents data for the login page
type LoginPageData struct {
	Title       string
	CSRFToken   string
	Error       string
	RedirectURL string
}

// DashboardData represents data for the dashboard page
type DashboardData struct {
	Title          string
	User           UserInfo
	SystemStats    SystemStats
	RecentActivity []ActivityItem
	QuickActions   []QuickAction
	Notifications  []Notification
	CSRFToken      string
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalTenants      int
	ActiveUsers       int
	TotalTransactions int64
	SystemHealth      string
	LastBackup        time.Time
}

// Tenant management types

// ConsoleTenant represents a tenant in the Console service with enhanced data
type ConsoleTenant struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Status      string      `json:"status"`
	Industry    string      `json:"industry"`
	ContactInfo ContactInfo `json:"contactInfo"`
	Stats       TenantStats `json:"stats"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// ConsoleTenantFilters represents filtering options for tenant list
type ConsoleTenantFilters struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search"`
	Status   string `json:"status"`
	Industry string `json:"industry"`
}

// ConsoleTenantListResult represents the result of tenant listing
type ConsoleTenantListResult struct {
	Tenants    []ConsoleTenant `json:"tenants"`
	TotalCount int             `json:"totalCount"`
	Page       int             `json:"page"`
	PageSize   int             `json:"pageSize"`
	HasNext    bool            `json:"hasNext"`
	HasPrev    bool            `json:"hasPrev"`
}

// ConsoleTenantCreateRequest represents tenant creation request
type ConsoleTenantCreateRequest struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Industry     string `json:"industry"`
	ContactEmail string `json:"contactEmail"`
	ContactPhone string `json:"contactPhone"`
	CompanyName  string `json:"companyName"`
}

// ConsoleTenantCreateResult represents tenant creation result
type ConsoleTenantCreateResult struct {
	TenantID   string `json:"tenantId"`
	WorkflowID string `json:"workflowId"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

// ConsoleTenantsPageData represents data for the console tenants page
type ConsoleTenantsPageData struct {
	Title      string                `json:"title"`
	Tenants    []ConsoleTenant       `json:"tenants"`
	Pagination *PaginationMeta       `json:"pagination"`
	Filters    *ConsoleTenantFilters `json:"filters"`
	CSRFToken  string                `json:"csrfToken"`
}

// ConsoleTenantFormData represents tenant form data
type ConsoleTenantFormData struct {
	Title     string `json:"title"`
	Action    string `json:"action"`
	Method    string `json:"method"`
	CSRFToken string `json:"csrfToken"`
}

// Legacy types for backward compatibility

// TenantItem represents a tenant in the list view (legacy)
type TenantItem struct {
	ID          string
	Name        string
	Email       string
	Subdomain   *string
	Status      string
	Industry    string
	CompanySize string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TenantDetail represents detailed tenant information
type TenantDetail struct {
	ID          string
	Name        string
	Email       string
	Subdomain   *string
	Status      string
	Industry    string
	CompanySize string
	Timezone    string
	Currency    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TenantsPageData represents data for the tenants list page
type TenantsPageData struct {
	Title      string
	CSRFToken  string
	Tenants    []TenantItem
	Pagination Pagination
}

// TenantDetailPageData represents data for the tenant detail page
type TenantDetailPageData struct {
	Title     string
	CSRFToken string
	Tenant    TenantDetail
}

// CreateTenantPageData represents data for the create tenant page
type CreateTenantPageData struct {
	Title        string
	CSRFToken    string
	Tenant       TenantDetail
	Form         TenantForm
	Industries   []SelectOption
	CompanySizes []SelectOption
	Error        string
}

// UserFormData represents data for the user form page
type UserFormData struct {
	Title              string
	CSRFToken          string
	User               UserProfile
	Roles              []SelectOption
	Departments        []SelectOption
	ShowPasswordFields bool
	Error              string
}

// UserProfile represents detailed user information
type UserProfile struct {
	ID         string
	FirstName  string
	LastName   string
	Email      string
	Username   string
	Phone      string
	JobTitle   string
	Role       string
	Department string
}

// TenantForm represents the tenant creation/edit form
type TenantForm struct {
	Name        string
	Email       string
	Subdomain   *string
	Industry    string
	CompanySize string
}

// SelectOption represents an option in a select field
type SelectOption struct {
	Value    string
	Label    string
	Selected bool
}

// Navigation and UI component types

// NavItem represents a navigation menu item
type NavItem struct {
	Label      string
	URL        string
	Icon       string
	Permission string
	Badge      string
	Children   []NavItem
}

// Modal component types

// ModalProps represents the properties for a base modal
type ModalProps struct {
	ID              string
	Title           string
	Subtitle        string
	Size            string // sm, md, lg, xl, 2xl
	ShowCloseButton bool
	Actions         []ModalAction
}

// ModalAction represents an action button in a modal
type ModalAction struct {
	Type        string // button, submit, htmx
	Label       string
	LoadingText string
	OnClick     string
	FormID      string
	URL         string
	Target      string
	Swap        string
}

// ConfirmDialogProps represents properties for confirmation dialogs
type ConfirmDialogProps struct {
	ID          string
	Title       string
	Message     string
	ConfirmText string
	CancelText  string
	ActionURL   string
	Target      string
	Swap        string
}

// FormModalProps represents properties for form modals
type FormModalProps struct {
	ID         string
	Title      string
	Subtitle   string
	Size       string
	FormID     string
	SubmitText string
	ActionURL  string
	Target     string
	Swap       string
}

// InfoModalProps represents properties for informational modals
type InfoModalProps struct {
	ID       string
	Title    string
	Subtitle string
	Size     string
}

// Form component types

// InputProps represents properties for input fields
type InputProps struct {
	ID              string
	Name            string
	Type            string
	Label           string
	Value           string
	Placeholder     string
	Required        bool
	Disabled        bool
	ReadOnly        bool
	HasError        bool
	HelpText        string
	Tooltip         string
	MaxLength       int
	Autocomplete    string
	FormID          string
	ValidateOnInput bool
	ValidationURL   string
}

// SelectProps represents properties for select fields
type SelectProps struct {
	ID          string
	Name        string
	Label       string
	Required    bool
	Disabled    bool
	Multiple    bool
	HasError    bool
	HelpText    string
	Placeholder string
	Searchable  bool
	Options     []SelectOption
}

// TextareaProps represents properties for textarea fields
type TextareaProps struct {
	ID          string
	Name        string
	Label       string
	Value       string
	Placeholder string
	Required    bool
	Disabled    bool
	ReadOnly    bool
	HasError    bool
	HelpText    string
	Rows        int
	MaxLength   int
	AutoResize  bool
	FormID      string
}

// CheckboxProps represents properties for checkbox fields
type CheckboxProps struct {
	ID       string
	Name     string
	Label    string
	Value    string
	Checked  bool
	Required bool
	Disabled bool
	HelpText string
}

// DataTable component types

// DataTableProps represents properties for the data table component
type DataTableProps struct {
	Columns           []DataTableColumn
	Rows              []DataTableRow
	Pagination        Pagination
	Searchable        bool
	SearchPlaceholder string
	SearchURL         string
	Selectable        bool
	Sortable          bool
	SortURL           string
	Filters           []DataTableFilter
	FilterURL         string
	BulkActions       []DataTableBulkAction
	Actions           []DataTableAction
	AddURL            string
	AddButtonText     string
	EmptyMessage      string
}

// DataTableColumn represents a table column
type DataTableColumn struct {
	Key      string
	Label    string
	Sortable bool
	Width    string
}

// DataTableRow represents a table row
type DataTableRow struct {
	ID   string
	Data map[string]string
}

// DataTableRowProps represents properties for individual table rows
type DataTableRowProps struct {
	Row        DataTableRow
	Columns    []DataTableColumn
	Actions    []DataTableAction
	Selectable bool
	Index      int
}

// DataTableFilter represents a filter option
type DataTableFilter struct {
	Label string
	Value string
}

// DataTableBulkAction represents a bulk action
type DataTableBulkAction struct {
	Label  string
	Action string
	Icon   string
}

// DataTableAction represents a row action
type DataTableAction struct {
	Label   string
	Type    string // link, button, modal
	URL     string
	Action  string
	Icon    string
	ModalID string
}
