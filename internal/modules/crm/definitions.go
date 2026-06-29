// Package crm declares EntityDefinitions for the CRM module.
package crm

import (
	"context"

	"awo.so/framework/def"
)

func init() {
	def.Register(Customer)
	def.Register(Contract)
}

func allowCRM(_ context.Context, viewer def.ViewerContext, _ def.Op, _ def.Record) error {
	if viewer.IsSystem() || viewer.HasRole("crm_manager") || viewer.HasRole("sales_rep") || viewer.HasRole("tenant_admin") {
		return def.ErrAllow
	}
	return def.ErrDeny
}

// Customer is a Person record typed as CUSTOMER, exposed as its own CRM entity.
var Customer = &def.EntityDefinition{
	Name:       "customer",
	Label:      "Customer",
	Table:      "persons",
	Module:     "CRM",
	SoftDelete: true,
	Fields: []*def.FieldDef{
		def.Field("first_name").OfType(def.FieldTypeSmallText).WithLabel("First Name").
			RequiredField().SearchableField().WithMaxLen(100),
		def.Field("last_name").OfType(def.FieldTypeSmallText).WithLabel("Last Name").
			RequiredField().SearchableField().WithMaxLen(100),
		def.Field("email").OfType(def.FieldTypeSmallText).WithLabel("Email").
			UniqueField().WithMaxLen(320),
		def.Field("phone").OfType(def.FieldTypeSmallText).WithLabel("Phone").WithMaxLen(30),
		def.Field("tax_id").OfType(def.FieldTypeSmallText).WithLabel("Tax ID").
			SensitiveField().WithMaxLen(50),
		def.Field("is_active").OfType(def.FieldTypeBool).WithLabel("Active").WithDefault(true),
		// person_type is fixed to CUSTOMER — set via BeforeHook in registration.
		def.Field("person_type").OfType(def.FieldTypeData).
			WithDefault("CUSTOMER").ReadOnlyField().HiddenInList(),
	},
	Hooks: []def.HookDef{
		// Ensure person_type is always CUSTOMER regardless of payload.
		def.BeforeHook(def.OpCreate|def.OpUpdate, func(_ context.Context, m *def.Mutation) error {
			if m.After != nil {
				m.After.Set("person_type", "CUSTOMER")
			}
			return nil
		}),
	},
	Policies: []def.PolicyDef{
		def.Policy(def.OpAll, allowCRM),
	},
}

// Contract is a signed agreement with a counterparty.
var Contract = &def.EntityDefinition{
	Name:       "contract",
	Label:      "Contract",
	Table:      "contracts",
	Module:     "CRM",
	SoftDelete: true,
	Audited:    true,
	Fields: []*def.FieldDef{
		def.Field("number").OfType(def.FieldTypeSmallText).WithLabel("Contract #").
			RequiredField().UniqueField().SearchableField().WithMaxLen(50),
		def.Field("title").OfType(def.FieldTypeSmallText).WithLabel("Title").
			RequiredField().SearchableField().WithMaxLen(255),
		def.Field("status").OfType(def.FieldTypeSelect).WithLabel("Status").
			WithOptions("DRAFT", "ACTIVE", "EXPIRED", "TERMINATED", "PENDING_APPROVAL").
			WithDefault("DRAFT"),
		def.Field("contract_type").OfType(def.FieldTypeSmallText).WithLabel("Type").
			WithMaxLen(50),
		def.Field("counterparty_name").OfType(def.FieldTypeSmallText).WithLabel("Counterparty").
			RequiredField().SearchableField().WithMaxLen(255),
		def.Field("counterparty_email").OfType(def.FieldTypeSmallText).WithLabel("Counterparty Email").
			WithMaxLen(320),
		def.Field("start_date").OfType(def.FieldTypeDate).WithLabel("Start Date"),
		def.Field("end_date").OfType(def.FieldTypeDate).WithLabel("End Date"),
		def.Field("value").OfType(def.FieldTypeCurrency).WithLabel("Value"),
		def.Field("currency_code").OfType(def.FieldTypeSmallText).WithLabel("Currency").
			WithMaxLen(3).WithDefault("USD"),
		def.Field("description").OfType(def.FieldTypeLongText).WithLabel("Description"),
	},
	Policies: []def.PolicyDef{
		def.Policy(def.OpAll, allowCRM),
	},
}
