package generator

import (
	"time"

	"github.com/google/uuid"
)

// TestUser represents a user entity for testing UI generation
type TestUser struct {
	ID          uuid.UUID  `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	Email       string     `json:"email" db:"email,unique,not null" ui:"component=email;label=Email Address;required=true;placeholder=Enter your email" validate:"required,email" form:"section=basic"`
	Username    string     `json:"username" db:"username,unique,not null" ui:"component=text;label=Username;required=true;placeholder=Choose a username" validate:"required,minlen=3,maxlen=20" form:"section=basic"`
	FirstName   string     `json:"first_name" db:"first_name,not null" ui:"component=text;label=First Name;required=true;placeholder=Enter first name" validate:"required,minlen=2" form:"section=basic"`
	LastName    string     `json:"last_name" db:"last_name,not null" ui:"component=text;label=Last Name;required=true;placeholder=Enter last name" validate:"required,minlen=2" form:"section=basic"`
	Password    string     `json:"-" db:"password_hash,not null" ui:"component=password;label=Password;required=true;placeholder=Enter password" validate:"required,minlen=8" form:"section=security"`
	Phone       *string    `json:"phone" db:"phone" ui:"component=tel;label=Phone Number;placeholder=Enter phone number" validate:"omitempty" form:"section=contact"`
	Avatar      *string    `json:"avatar" db:"avatar" ui:"component=image;label=Profile Picture" form:"section=profile"`
	Status      string     `json:"status" db:"status,not null" ui:"component=select;label=Status;options=active,inactive,pending;badge=true" validate:"required,oneof=active inactive pending" form:"section=basic"`
	IsAdmin     bool       `json:"is_admin" db:"is_admin,not null" ui:"component=checkbox;label=Administrator Access" form:"section=permissions"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at" ui:"component=datetime;label=Last Login;readonly=true;format=relative" table:"sortable=true"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at,not null" ui:"component=datetime;label=Created;readonly=true;format=relative" table:"sortable=true"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;label=Updated;readonly=true;format=relative" table:"sortable=true"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at" ui:"component=hidden"`

	// Relationships
	Roles      []TestRole      `json:"roles,omitempty" ui:"relation=roles;component=multi-select" form:"section=permissions"`
	Profile    *TestProfile    `json:"profile,omitempty" ui:"relation=profile" form:"section=profile"`
	Department *TestDepartment `json:"department,omitempty" ui:"relation=department;component=select" form:"section=organization"`
	Manager    *TestUser       `json:"manager,omitempty" ui:"relation=manager;component=select" form:"section=organization"`
}

// TestRole represents a user role for testing
type TestRole struct {
	ID          uuid.UUID `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	Name        string    `json:"name" db:"name,not null" ui:"component=text;label=Role Name;required=true" validate:"required,minlen=2"`
	DisplayName string    `json:"display_name" db:"display_name,not null" ui:"component=text;label=Display Name;required=true" validate:"required"`
	Description *string   `json:"description" db:"description" ui:"component=textarea;label=Description;rows=3"`
	IsActive    bool      `json:"is_active" db:"is_active,not null" ui:"component=toggle;label=Active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at,not null" ui:"component=datetime;readonly=true"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;readonly=true"`
}

// TestProfile represents extended user profile information
type TestProfile struct {
	ID          uuid.UUID `json:"id" db:"id,primarykey" ui:"component=hidden"`
	UserID      uuid.UUID `json:"user_id" db:"user_id,not null,unique" ui:"component=hidden"`
	Bio         *string   `json:"bio" db:"bio" ui:"component=rich-text;label=Biography;help=Tell us about yourself"`
	Website     *string   `json:"website" db:"website" ui:"component=url;label=Website" validate:"omitempty,url"`
	Location    *string   `json:"location" db:"location" ui:"component=text;label=Location;placeholder=City, Country"`
	Timezone    string    `json:"timezone" db:"timezone,not null" ui:"component=select;label=Timezone;options=UTC,America/New_York,America/Los_Angeles,Europe/London"`
	Language    string    `json:"language" db:"language,not null" ui:"component=select;label=Language;options=en,es,fr,de"`
	DateFormat  string    `json:"date_format" db:"date_format,not null" ui:"component=select;label=Date Format;options=MM/dd/yyyy,dd/MM/yyyy,yyyy-MM-dd"`
	Preferences *string   `json:"preferences" db:"preferences" ui:"component=hidden"` // JSON field
	CreatedAt   time.Time `json:"created_at" db:"created_at,not null" ui:"component=hidden"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at,not null" ui:"component=hidden"`
}

// TestDepartment represents an organizational department
type TestDepartment struct {
	ID          uuid.UUID  `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	Name        string     `json:"name" db:"name,not null" ui:"component=text;label=Department Name;required=true" validate:"required"`
	Code        string     `json:"code" db:"code,not null,unique" ui:"component=text;label=Department Code;required=true" validate:"required,alphanum,maxlen=10"`
	ParentID    *uuid.UUID `json:"parent_id" db:"parent_id" ui:"component=select;label=Parent Department;relation=department"`
	ManagerID   *uuid.UUID `json:"manager_id" db:"manager_id" ui:"component=select;label=Department Manager;relation=user"`
	Description *string    `json:"description" db:"description" ui:"component=textarea;label=Description"`
	Budget      *float64   `json:"budget" db:"budget" ui:"component=number;label=Budget;format=currency" validate:"omitempty,min=0"`
	IsActive    bool       `json:"is_active" db:"is_active,not null" ui:"component=toggle;label=Active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at,not null" ui:"component=datetime;readonly=true"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;readonly=true"`

	// Hierarchy relationships
	Parent    *TestDepartment  `json:"parent,omitempty" ui:"relation=department"`
	Children  []TestDepartment `json:"children,omitempty" ui:"relation=departments"`
	Manager   *TestUser        `json:"manager,omitempty" ui:"relation=manager"`
	Employees []TestUser       `json:"employees,omitempty" ui:"relation=employees"`
}

// TestProject represents a project for task management testing
type TestProject struct {
	ID          uuid.UUID  `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	Name        string     `json:"name" db:"name,not null" ui:"component=text;label=Project Name;required=true" validate:"required"`
	Description *string    `json:"description" db:"description" ui:"component=rich-text;label=Description"`
	Status      string     `json:"status" db:"status,not null" ui:"component=select;label=Status;options=planning,active,on_hold,completed,cancelled;badge=true" validate:"required"`
	Priority    string     `json:"priority" db:"priority,not null" ui:"component=select;label=Priority;options=low,medium,high,urgent;badge=true" validate:"required"`
	StartDate   *time.Time `json:"start_date" db:"start_date" ui:"component=date;label=Start Date"`
	EndDate     *time.Time `json:"end_date" db:"end_date" ui:"component=date;label=End Date"`
	Budget      *float64   `json:"budget" db:"budget" ui:"component=number;label=Budget;format=currency" validate:"omitempty,min=0"`
	Progress    int        `json:"progress" db:"progress,not null" ui:"component=number;label=Progress (%);min=0;max=100" validate:"min=0,max=100"`
	ManagerID   uuid.UUID  `json:"manager_id" db:"manager_id,not null" ui:"component=select;label=Project Manager;relation=user;required=true" validate:"required"`
	ClientID    *uuid.UUID `json:"client_id" db:"client_id" ui:"component=select;label=Client;relation=client"`
	Tags        *string    `json:"tags" db:"tags" ui:"component=text;label=Tags;help=Comma-separated tags"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at,not null" ui:"component=datetime;readonly=true"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;readonly=true"`

	// Relationships
	Manager *TestUser   `json:"manager,omitempty" ui:"relation=manager"`
	Client  *TestClient `json:"client,omitempty" ui:"relation=client"`
	Tasks   []TestTask  `json:"tasks,omitempty" ui:"relation=tasks"`
	Team    []TestUser  `json:"team,omitempty" ui:"relation=team;component=multi-select"`
}

// TestTask represents a task within a project
type TestTask struct {
	ID             uuid.UUID  `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID       uuid.UUID  `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	ProjectID      uuid.UUID  `json:"project_id" db:"project_id,not null" ui:"component=select;label=Project;relation=project;required=true" validate:"required"`
	Title          string     `json:"title" db:"title,not null" ui:"component=text;label=Task Title;required=true" validate:"required"`
	Description    *string    `json:"description" db:"description" ui:"component=textarea;label=Description;rows=4"`
	Status         string     `json:"status" db:"status,not null" ui:"component=select;label=Status;options=todo,in_progress,review,done;badge=true" validate:"required"`
	Priority       string     `json:"priority" db:"priority,not null" ui:"component=select;label=Priority;options=low,medium,high,urgent;badge=true" validate:"required"`
	AssigneeID     *uuid.UUID `json:"assignee_id" db:"assignee_id" ui:"component=select;label=Assignee;relation=user"`
	EstimatedHours *float64   `json:"estimated_hours" db:"estimated_hours" ui:"component=number;label=Estimated Hours;min=0" validate:"omitempty,min=0"`
	ActualHours    *float64   `json:"actual_hours" db:"actual_hours" ui:"component=number;label=Actual Hours;min=0" validate:"omitempty,min=0"`
	DueDate        *time.Time `json:"due_date" db:"due_date" ui:"component=datetime;label=Due Date"`
	CompletedAt    *time.Time `json:"completed_at" db:"completed_at" ui:"component=datetime;label=Completed;readonly=true"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at,not null" ui:"component=datetime;readonly=true"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;readonly=true"`

	// Relationships
	Project  *TestProject `json:"project,omitempty" ui:"relation=project"`
	Assignee *TestUser    `json:"assignee,omitempty" ui:"relation=assignee"`
}

// TestClient represents a client/customer entity
type TestClient struct {
	ID            uuid.UUID `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID      uuid.UUID `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	Name          string    `json:"name" db:"name,not null" ui:"component=text;label=Client Name;required=true" validate:"required"`
	CompanyName   *string   `json:"company_name" db:"company_name" ui:"component=text;label=Company Name"`
	Email         string    `json:"email" db:"email,not null" ui:"component=email;label=Email;required=true" validate:"required,email"`
	Phone         *string   `json:"phone" db:"phone" ui:"component=tel;label=Phone"`
	Address       *string   `json:"address" db:"address" ui:"component=textarea;label=Address;rows=3"`
	City          *string   `json:"city" db:"city" ui:"component=text;label=City"`
	Country       *string   `json:"country" db:"country" ui:"component=select;label=Country;options=US,CA,UK,DE,FR,AU"`
	PostalCode    *string   `json:"postal_code" db:"postal_code" ui:"component=text;label=Postal Code"`
	Website       *string   `json:"website" db:"website" ui:"component=url;label=Website" validate:"omitempty,url"`
	Industry      *string   `json:"industry" db:"industry" ui:"component=select;label=Industry;options=technology,finance,healthcare,education,retail,manufacturing"`
	AnnualRevenue *float64  `json:"annual_revenue" db:"annual_revenue" ui:"component=number;label=Annual Revenue;format=currency" validate:"omitempty,min=0"`
	EmployeeCount *int      `json:"employee_count" db:"employee_count" ui:"component=number;label=Employee Count" validate:"omitempty,min=1"`
	Status        string    `json:"status" db:"status,not null" ui:"component=select;label=Status;options=prospect,active,inactive,closed;badge=true" validate:"required"`
	Notes         *string   `json:"notes" db:"notes" ui:"component=rich-text;label=Notes"`
	CreatedAt     time.Time `json:"created_at" db:"created_at,not null" ui:"component=datetime;readonly=true"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;readonly=true"`

	// Relationships
	Projects []TestProject `json:"projects,omitempty" ui:"relation=projects"`
	Contacts []TestContact `json:"contacts,omitempty" ui:"relation=contacts"`
}

// TestContact represents a contact person within a client organization
type TestContact struct {
	ID        uuid.UUID `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	ClientID  uuid.UUID `json:"client_id" db:"client_id,not null" ui:"component=select;label=Client;relation=client;required=true" validate:"required"`
	FirstName string    `json:"first_name" db:"first_name,not null" ui:"component=text;label=First Name;required=true" validate:"required"`
	LastName  string    `json:"last_name" db:"last_name,not null" ui:"component=text;label=Last Name;required=true" validate:"required"`
	Email     string    `json:"email" db:"email,not null" ui:"component=email;label=Email;required=true" validate:"required,email"`
	Phone     *string   `json:"phone" db:"phone" ui:"component=tel;label=Phone"`
	JobTitle  *string   `json:"job_title" db:"job_title" ui:"component=text;label=Job Title"`
	IsPrimary bool      `json:"is_primary" db:"is_primary,not null" ui:"component=checkbox;label=Primary Contact"`
	CreatedAt time.Time `json:"created_at" db:"created_at,not null" ui:"component=datetime;readonly=true"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;readonly=true"`

	// Relationships
	Client *TestClient `json:"client,omitempty" ui:"relation=client"`
}

// TestInvoice represents a financial invoice for testing
type TestInvoice struct {
	ID            uuid.UUID  `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	InvoiceNumber string     `json:"invoice_number" db:"invoice_number,not null,unique" ui:"component=text;label=Invoice Number;required=true;readonly=true" validate:"required"`
	ClientID      uuid.UUID  `json:"client_id" db:"client_id,not null" ui:"component=select;label=Client;relation=client;required=true" validate:"required"`
	ProjectID     *uuid.UUID `json:"project_id" db:"project_id" ui:"component=select;label=Project;relation=project"`
	IssueDate     time.Time  `json:"issue_date" db:"issue_date,not null" ui:"component=date;label=Issue Date;required=true" validate:"required"`
	DueDate       time.Time  `json:"due_date" db:"due_date,not null" ui:"component=date;label=Due Date;required=true" validate:"required"`
	Status        string     `json:"status" db:"status,not null" ui:"component=select;label=Status;options=draft,sent,paid,overdue,cancelled;badge=true" validate:"required"`
	Subtotal      float64    `json:"subtotal" db:"subtotal,not null" ui:"component=number;label=Subtotal;format=currency;readonly=true" validate:"required,min=0"`
	TaxAmount     float64    `json:"tax_amount" db:"tax_amount,not null" ui:"component=number;label=Tax Amount;format=currency" validate:"min=0"`
	TaxRate       float64    `json:"tax_rate" db:"tax_rate,not null" ui:"component=number;label=Tax Rate (%);min=0;max=100" validate:"min=0,max=100"`
	Total         float64    `json:"total" db:"total,not null" ui:"component=number;label=Total;format=currency;readonly=true" validate:"required,min=0"`
	Currency      string     `json:"currency" db:"currency,not null" ui:"component=select;label=Currency;options=USD,EUR,GBP,CAD,AUD" validate:"required"`
	Notes         *string    `json:"notes" db:"notes" ui:"component=textarea;label=Notes;rows=3"`
	Terms         *string    `json:"terms" db:"terms" ui:"component=textarea;label=Payment Terms;rows=2"`
	PaidAt        *time.Time `json:"paid_at" db:"paid_at" ui:"component=datetime;label=Paid Date;readonly=true"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at,not null" ui:"component=datetime;readonly=true"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;readonly=true"`

	// Relationships
	Client    *TestClient       `json:"client,omitempty" ui:"relation=client"`
	Project   *TestProject      `json:"project,omitempty" ui:"relation=project"`
	LineItems []TestInvoiceItem `json:"line_items,omitempty" ui:"relation=line_items"`
}

// TestInvoiceItem represents a line item in an invoice
type TestInvoiceItem struct {
	ID          uuid.UUID `json:"id" db:"id,primarykey" ui:"component=hidden"`
	InvoiceID   uuid.UUID `json:"invoice_id" db:"invoice_id,not null" ui:"component=hidden"`
	Description string    `json:"description" db:"description,not null" ui:"component=text;label=Description;required=true" validate:"required"`
	Quantity    float64   `json:"quantity" db:"quantity,not null" ui:"component=number;label=Quantity;required=true;min=0" validate:"required,min=0"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price,not null" ui:"component=number;label=Unit Price;format=currency;required=true;min=0" validate:"required,min=0"`
	Total       float64   `json:"total" db:"total,not null" ui:"component=number;label=Total;format=currency;readonly=true" validate:"required,min=0"`
	CreatedAt   time.Time `json:"created_at" db:"created_at,not null" ui:"component=hidden"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at,not null" ui:"component=hidden"`

	// Relationships
	Invoice *TestInvoice `json:"invoice,omitempty" ui:"relation=invoice"`
}

// TestEvent represents a calendar event for testing calendar patterns
type TestEvent struct {
	ID             uuid.UUID `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID       uuid.UUID `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	Title          string    `json:"title" db:"title,not null" ui:"component=text;label=Event Title;required=true" validate:"required"`
	Description    *string   `json:"description" db:"description" ui:"component=rich-text;label=Description"`
	StartTime      time.Time `json:"start_time" db:"start_time,not null" ui:"component=datetime;label=Start Time;required=true" validate:"required"`
	EndTime        time.Time `json:"end_time" db:"end_time,not null" ui:"component=datetime;label=End Time;required=true" validate:"required"`
	AllDay         bool      `json:"all_day" db:"all_day,not null" ui:"component=checkbox;label=All Day Event"`
	Location       *string   `json:"location" db:"location" ui:"component=text;label=Location"`
	Type           string    `json:"type" db:"type,not null" ui:"component=select;label=Event Type;options=meeting,appointment,deadline,reminder;badge=true" validate:"required"`
	Status         string    `json:"status" db:"status,not null" ui:"component=select;label=Status;options=scheduled,in_progress,completed,cancelled;badge=true" validate:"required"`
	OrganizerID    uuid.UUID `json:"organizer_id" db:"organizer_id,not null" ui:"component=select;label=Organizer;relation=user;required=true" validate:"required"`
	IsRecurring    bool      `json:"is_recurring" db:"is_recurring,not null" ui:"component=checkbox;label=Recurring Event"`
	RecurrenceRule *string   `json:"recurrence_rule" db:"recurrence_rule" ui:"component=text;label=Recurrence Rule;help=RRULE format"`
	CreatedAt      time.Time `json:"created_at" db:"created_at,not null" ui:"component=datetime;readonly=true"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;readonly=true"`

	// Relationships
	Organizer *TestUser  `json:"organizer,omitempty" ui:"relation=organizer"`
	Attendees []TestUser `json:"attendees,omitempty" ui:"relation=attendees;component=multi-select"`
}

// TestNotification represents a notification for testing notification patterns
type TestNotification struct {
	ID        uuid.UUID  `json:"id" db:"id,primarykey" ui:"component=hidden"`
	TenantID  uuid.UUID  `json:"tenant_id" db:"tenant_id,not null" ui:"component=hidden"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id,not null" ui:"component=select;label=User;relation=user;required=true" validate:"required"`
	Type      string     `json:"type" db:"type,not null" ui:"component=select;label=Type;options=info,success,warning,error;badge=true" validate:"required"`
	Title     string     `json:"title" db:"title,not null" ui:"component=text;label=Title;required=true" validate:"required"`
	Message   string     `json:"message" db:"message,not null" ui:"component=textarea;label=Message;required=true;rows=3" validate:"required"`
	IsRead    bool       `json:"is_read" db:"is_read,not null" ui:"component=checkbox;label=Read"`
	ReadAt    *time.Time `json:"read_at" db:"read_at" ui:"component=datetime;label=Read At;readonly=true"`
	ActionURL *string    `json:"action_url" db:"action_url" ui:"component=url;label=Action URL" validate:"omitempty,url"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at" ui:"component=datetime;label=Expires At"`
	Priority  string     `json:"priority" db:"priority,not null" ui:"component=select;label=Priority;options=low,normal,high,urgent;badge=true" validate:"required"`
	CreatedAt time.Time  `json:"created_at" db:"created_at,not null" ui:"component=datetime;readonly=true"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at,not null" ui:"component=datetime;readonly=true"`

	// Relationships
	User *TestUser `json:"user,omitempty" ui:"relation=user"`
}
