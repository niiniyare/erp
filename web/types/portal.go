package types

import (
	"time"
)

// PortalDashboardData represents data for the portal dashboard
type PortalDashboardData struct {
	Title          string               `json:"title"`
	ClientInfo     PortalClientInfo     `json:"client_info"`
	AccountSummary PortalAccountSummary `json:"account_summary"`
	RecentActivity []ActivityItem       `json:"recent_activity"`
	QuickActions   []QuickAction        `json:"quick_actions"`
	Notifications  []Notification       `json:"notifications"`
	CSRFToken      string               `json:"csrf_token"`
}

// PortalClientInfo represents client information in the portal
type PortalClientInfo struct {
	ID              string    `json:"id"`
	CompanyName     string    `json:"company_name"`
	ContactName     string    `json:"contact_name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Status          string    `json:"status"`
	AccountType     string    `json:"account_type"`
	MemberSince     time.Time `json:"member_since"`
	BillingAddress  Address   `json:"billing_address"`
	ShippingAddress Address   `json:"shipping_address"`
}

// PortalAccountSummary represents account summary information
type PortalAccountSummary struct {
	CurrentBalance  float64            `json:"current_balance"`
	OpenInvoices    int                `json:"open_invoices"`
	ActiveOrders    int                `json:"active_orders"`
	OpenTickets     int                `json:"open_tickets"`
	NextDueDate     time.Time          `json:"next_due_date"`
	NextDueAmount   float64            `json:"next_due_amount"`
	CreditLimit     float64            `json:"credit_limit"`
	AvailableCredit float64            `json:"available_credit"`
	LastPayment     *PaymentInfo       `json:"last_payment,omitempty"`
	YearToDate      *YearToDateSummary `json:"year_to_date,omitempty"`
}

// PaymentInfo represents payment information
type PaymentInfo struct {
	Amount float64   `json:"amount"`
	Date   time.Time `json:"date"`
	Method string    `json:"method"`
}

// YearToDateSummary represents year-to-date summary
type YearToDateSummary struct {
	TotalPurchases    float64 `json:"total_purchases"`
	TotalPayments     float64 `json:"total_payments"`
	OrderCount        int     `json:"order_count"`
	AverageOrderValue float64 `json:"average_order_value"`
}

// PortalInvoice represents an invoice in the portal context
type PortalInvoice struct {
	ID             string            `json:"id"`
	Number         string            `json:"number"`
	Date           time.Time         `json:"date"`
	DueDate        time.Time         `json:"due_date"`
	Amount         float64           `json:"amount"`
	Status         string            `json:"status"`
	Description    string            `json:"description"`
	Currency       string            `json:"currency"`
	TaxAmount      float64           `json:"tax_amount"`
	PaidDate       *time.Time        `json:"paid_date,omitempty"`
	LineItems      []InvoiceLineItem `json:"line_items,omitempty"`
	BillingAddress Address           `json:"billing_address,omitempty"`
}

// InvoiceLineItem represents a line item in an invoice
type InvoiceLineItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
}

// PortalInvoiceFilters represents filtering options for invoice lists
type PortalInvoiceFilters struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status,omitempty"`
	Year     int    `json:"year,omitempty"`
}

// PortalInvoicesPageData represents data for the portal invoices page
type PortalInvoicesPageData struct {
	Title      string                `json:"title"`
	Invoices   []PortalInvoice       `json:"invoices"`
	Pagination *PaginationMeta       `json:"pagination"`
	Filters    *PortalInvoiceFilters `json:"filters"`
	CSRFToken  string                `json:"csrf_token"`
}

// PortalInvoiceDetailPageData represents data for invoice detail page
type PortalInvoiceDetailPageData struct {
	Title     string         `json:"title"`
	Invoice   *PortalInvoice `json:"invoice"`
	CSRFToken string         `json:"csrf_token"`
}

// PaymentResult represents the result of a payment operation
type PaymentResult struct {
	PaymentID string  `json:"payment_id"`
	NewStatus string  `json:"new_status"`
	Amount    float64 `json:"amount"`
	Method    string  `json:"method"`
}

// PortalOrder represents an order in the portal context
type PortalOrder struct {
	ID           string     `json:"id"`
	Number       string     `json:"number"`
	Date         time.Time  `json:"date"`
	Status       string     `json:"status"`
	Amount       float64    `json:"amount"`
	ItemCount    int        `json:"item_count"`
	ShipDate     *time.Time `json:"ship_date,omitempty"`
	DeliveryDate *time.Time `json:"delivery_date,omitempty"`
	TrackingCode string     `json:"tracking_code,omitempty"`
	Description  string     `json:"description"`
	Currency     string     `json:"currency"`
}

// PortalOrderFilters represents filtering options for order lists
type PortalOrderFilters struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status,omitempty"`
	Year     int    `json:"year,omitempty"`
}

// PortalOrdersPageData represents data for the portal orders page
type PortalOrdersPageData struct {
	Title      string              `json:"title"`
	Orders     []PortalOrder       `json:"orders"`
	Pagination *PaginationMeta     `json:"pagination"`
	Filters    *PortalOrderFilters `json:"filters"`
	CSRFToken  string              `json:"csrf_token"`
}

// PortalOrderDetailPageData represents data for order detail page
type PortalOrderDetailPageData struct {
	Title     string       `json:"title"`
	Order     *PortalOrder `json:"order"`
	CSRFToken string       `json:"csrf_token"`
}

// PortalTicket represents a support ticket in the portal context
type PortalTicket struct {
	ID          string                   `json:"id"`
	Number      string                   `json:"number"`
	Subject     string                   `json:"subject"`
	Description string                   `json:"description"`
	Status      string                   `json:"status"`
	Priority    string                   `json:"priority"`
	Category    string                   `json:"category"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Messages    []PortalTicketMessage    `json:"messages,omitempty"`
	Attachments []PortalTicketAttachment `json:"attachments,omitempty"`
}

// PortalTicketMessage represents a message in a support ticket
type PortalTicketMessage struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id"`
	Author    string    `json:"author"`
	IsClient  bool      `json:"is_client"`
	CreatedAt time.Time `json:"created_at"`
}

// PortalTicketAttachment represents an attachment in a support ticket
type PortalTicketAttachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url"`
}

// PortalTicketCreateRequest represents data for creating a new ticket
type PortalTicketCreateRequest struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Category    string `json:"category"`
}

// PortalTicketMessageRequest represents data for adding a message to a ticket
type PortalTicketMessageRequest struct {
	Content string `json:"content"`
}

// PortalTicketFilters represents filtering options for ticket lists
type PortalTicketFilters struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status,omitempty"`
}

// PortalTicketsPageData represents data for the portal tickets page
type PortalTicketsPageData struct {
	Title      string               `json:"title"`
	Tickets    []PortalTicket       `json:"tickets"`
	Pagination *PaginationMeta      `json:"pagination"`
	Filters    *PortalTicketFilters `json:"filters"`
	CSRFToken  string               `json:"csrf_token"`
}

// PortalTicketDetailPageData represents data for ticket detail page
type PortalTicketDetailPageData struct {
	Title     string        `json:"title"`
	Ticket    *PortalTicket `json:"ticket"`
	CSRFToken string        `json:"csrf_token"`
}

// PortalTicketFormData represents data for ticket forms
type PortalTicketFormData struct {
	Title      string   `json:"title"`
	Action     string   `json:"action"`
	Method     string   `json:"method"`
	Categories []string `json:"categories"`
	Priorities []string `json:"priorities"`
	CSRFToken  string   `json:"csrf_token"`
}

// PortalProfileUpdateRequest represents data for updating client profile
type PortalProfileUpdateRequest struct {
	ContactName     string  `json:"contact_name"`
	Email           string  `json:"email"`
	Phone           string  `json:"phone"`
	BillingAddress  Address `json:"billing_address"`
	ShippingAddress Address `json:"shipping_address"`
}

// PortalAccountPageData represents data for the account page
type PortalAccountPageData struct {
	Title      string            `json:"title"`
	ClientInfo *PortalClientInfo `json:"client_info"`
	CSRFToken  string            `json:"csrf_token"`
}

// PortalWidget represents data for portal widgets
type PortalWidget struct {
	Balance       Money     `json:"balance"`
	NextDueAmount Money     `json:"next_due_amount"`
	NextDueDate   time.Time `json:"next_due_date"`
	Status        string    `json:"status"`
}

// Money represents a monetary amount with currency
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// UINotification represents a UI notification (from shared types)
type UINotification struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	Title    string    `json:"title"`
	Message  string    `json:"message"`
	Severity string    `json:"severity"`
	Created  time.Time `json:"created"`
	Read     bool      `json:"read"`
}

// SearchFilters represents common search and filtering options
type SearchFilter struct {
	Query string `json:"q,omitempty"`
}

// PhonePattern represents regex pattern for phone validation
const PhonePattern = `^\+?[1-9]\d{1,14}$`
