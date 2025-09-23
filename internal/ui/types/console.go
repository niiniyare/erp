package types

import "time"

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

// UserInfo represents current user information
type UserInfo struct {
	ID       string
	Name     string
	Email    string
	Role     string
	TenantID string
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalTenants      int
	ActiveUsers       int
	TotalTransactions int64
	SystemHealth      string
	LastBackup        time.Time
}

// ActivityItem represents a recent activity item
type ActivityItem struct {
	ID          string
	Type        string
	Description string
	UserID      string
	UserName    string
	Timestamp   time.Time
	Severity    string
}

// QuickAction represents a quick action button
type QuickAction struct {
	Title       string
	Description string
	URL         string
	Icon        string
	Permission  string
}

// Notification represents a system notification
type Notification struct {
	ID       string
	Type     string
	Title    string
	Message  string
	Severity string
	Created  time.Time
	Read     bool
}

// Tenant management types

// TenantItem represents a tenant in the list view
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
	Form         TenantForm
	Industries   []SelectOption
	CompanySizes []SelectOption
	Error        string
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

// Pagination represents pagination information
type Pagination struct {
	CurrentPage int
	Limit       int
	Total       int
	HasNext     bool
	HasPrev     bool
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
