package core

import (
	"context"
	"time"
)

// BaseSchema provides foundation interface for all schema types
type BaseSchema interface {
	GetID() string
	GetType() SchemaType
	Validate() error
	GetMetadata() *Metadata
}

// SchemaType defines the type of schema
type SchemaType string

const (
	SchemaTypeForm      SchemaType = "form"
	SchemaTypeField     SchemaType = "field"
	SchemaTypeComponent SchemaType = "component"
	SchemaTypeLayout    SchemaType = "layout"
	SchemaTypeWorkflow  SchemaType = "workflow"
	SchemaTypeTheme     SchemaType = "theme"
	SchemaTypeData      SchemaType = "data"
)

// Metadata contains common metadata for all schemas
type Metadata struct {
	ID           string                 `json:"id"`
	Type         SchemaType             `json:"type"`
	Version      string                 `json:"version,omitempty"`
	CreatedAt    time.Time              `json:"createdAt,omitempty"`
	UpdatedAt    time.Time              `json:"updatedAt,omitempty"`
	CreatedBy    string                 `json:"createdBy,omitempty"`
	UpdatedBy    string                 `json:"updatedBy,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Category     string                 `json:"category,omitempty"`
	Module       string                 `json:"module,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Icon         string                 `json:"icon,omitempty"`
	Color        string                 `json:"color,omitempty"`
	Order        int                    `json:"order,omitempty"`
	Deprecated   bool                   `json:"deprecated"`
	Experimental bool                   `json:"experimental"`
	CustomData   map[string]any `json:"customData,omitempty"`
}

// SchemaReference represents a reference to another schema
type SchemaReference struct {
	ID      string     `json:"id"`
	Type    SchemaType `json:"type"`
	Version string     `json:"version,omitempty"`
}

// ConditionalSchema provides conditional logic support
type ConditionalSchema struct {
	Show    *ConditionGroup `json:"show,omitempty"`
	Hide    *ConditionGroup `json:"hide,omitempty"`
	Enable  *ConditionGroup `json:"enable,omitempty"`
	Disable *ConditionGroup `json:"disable,omitempty"`
	Require *ConditionGroup `json:"require,omitempty"`
}

// ConditionGroup represents a group of conditions (placeholder for condition package integration)
type ConditionGroup struct {
	ID          string                 `json:"id,omitempty"`
	Conjunction string                 `json:"conjunction"` // "and", "or"
	Conditions  []Condition            `json:"conditions"`
	Groups      []ConditionGroup       `json:"groups,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// Condition represents a single condition
type Condition struct {
	Field    string        `json:"field"`
	Operator string        `json:"operator"`
	Value    any   `json:"value"`
	Values   []any `json:"values,omitempty"`
}

// PermissionSchema defines access control
type PermissionSchema struct {
	View   *ConditionGroup `json:"view,omitempty"`
	Create *ConditionGroup `json:"create,omitempty"`
	Edit   *ConditionGroup `json:"edit,omitempty"`
	Delete *ConditionGroup `json:"delete,omitempty"`
	Roles  []string        `json:"roles,omitempty"`
	Groups []string        `json:"groups,omitempty"`
}

// TenantSchema defines multi-tenant configuration
type TenantSchema struct {
	Enabled    bool     `json:"enabled"`
	Field      string   `json:"field,omitempty"`     // Field that stores tenant ID
	Isolation  string   `json:"isolation"`           // strict, shared, hybrid
	Inherit    bool     `json:"inherit"`             // Inherit from parent
	AllowCross bool     `json:"allowCross"`          // Allow cross-tenant access
	Whitelist  []string `json:"whitelist,omitempty"` // Allowed tenants
	Blacklist  []string `json:"blacklist,omitempty"` // Blocked tenants
}

// I18nSchema defines internationalization support
type I18nSchema struct {
	Enabled          bool              `json:"enabled"`
	DefaultLocale    string            `json:"defaultLocale"`
	SupportedLocales []string          `json:"supportedLocales"`
	Translations     map[string]string `json:"translations,omitempty"`
	DateFormat       string            `json:"dateFormat,omitempty"`
	TimeFormat       string            `json:"timeFormat,omitempty"`
	NumberFormat     string            `json:"numberFormat,omitempty"`
	CurrencyFormat   string            `json:"currencyFormat,omitempty"`
}

// EvaluationContext provides context for condition evaluation
type EvaluationContext struct {
	Data        map[string]any `json:"data"`
	User        *UserContext           `json:"user,omitempty"`
	Tenant      *TenantContext         `json:"tenant,omitempty"`
	Request     *RequestContext        `json:"request,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// UserContext provides user information for evaluation
type UserContext struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Groups      []string `json:"groups"`
	Permissions []string `json:"permissions"`
}

// TenantContext provides tenant information for evaluation
type TenantContext struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Domain string `json:"domain"`
}

// RequestContext provides request information for evaluation
type RequestContext struct {
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	UserAgent string            `json:"userAgent"`
	IP        string            `json:"ip"`
}

// Evaluator interface for condition evaluation
type Evaluator interface {
	Evaluate(ctx context.Context, condition *ConditionGroup, evalCtx *EvaluationContext) (bool, error)
}

// SchemaRegistry interface for schema management
type SchemaRegistry interface {
	Register(schema BaseSchema) error
	Get(id string, schemaType SchemaType) (BaseSchema, error)
	List(schemaType SchemaType) ([]BaseSchema, error)
	Update(schema BaseSchema) error
	Delete(id string, schemaType SchemaType) error
}

// Renderer interface for schema rendering
type Renderer interface {
	Render(ctx context.Context, schema BaseSchema, data any) (string, error)
	CanRender(schemaType SchemaType) bool
}

// Transformer interface for data transformation
type Transformer interface {
	Transform(ctx context.Context, input any, config map[string]any) (any, error)
}

// Validator interface for schema validation
type Validator interface {
	Validate(ctx context.Context, schema BaseSchema) error
	ValidateData(ctx context.Context, schema BaseSchema, data any) error
}

// Helper functions for schema creation
func NewMetadata(id string, schemaType SchemaType) *Metadata {
	return &Metadata{
		ID:        id,
		Type:      schemaType,
		Version:   "1.0.0",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewEvaluationContext(data map[string]any) *EvaluationContext {
	return &EvaluationContext{
		Data:      data,
		Timestamp: time.Now(),
	}
}

func (m *Metadata) GetID() string {
	return m.ID
}

func (m *Metadata) GetType() SchemaType {
	return m.Type
}

func (m *Metadata) UpdateTimestamp() {
	m.UpdatedAt = time.Now()
}
