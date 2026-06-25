// Package crm declares EntityDefinitions for the CRM module.
package crm

import (
	"context"

	"awo.so/framework/definition"
)

func init() {
	definition.Register(Customer)
	definition.Register(Contract)
}

func allowCRM(_ context.Context, viewer definition.ViewerContext, _ definition.Op, _ definition.Record) error {
	if viewer.IsSystem() || viewer.HasRole("crm_manager") || viewer.HasRole("sales_rep") || viewer.HasRole("tenant_admin") {
		return definition.ErrAllow
	}
	return definition.ErrDeny
}

// Customer is a Person record typed as CUSTOMER, exposed as its own CRM entity.
var Customer = &definition.EntityDefinition{
	Name:       "customer",
	Label:      "Customer",
	Table:      "persons",
	Module:     "CRM",
	SoftDelete: true,
	Fields: []*definition.FieldDef{
		definition.Field("first_name").OfType(definition.FieldTypeSmallText).WithLabel("First Name").
			RequiredField().SearchableField().WithMaxLen(100),
		definition.Field("last_name").OfType(definition.FieldTypeSmallText).WithLabel("Last Name").
			RequiredField().SearchableField().WithMaxLen(100),
		definition.Field("email").OfType(definition.FieldTypeSmallText).WithLabel("Email").
			UniqueField().WithMaxLen(320),
		definition.Field("phone").OfType(definition.FieldTypeSmallText).WithLabel("Phone").WithMaxLen(30),
		definition.Field("tax_id").OfType(definition.FieldTypeSmallText).WithLabel("Tax ID").
			SensitiveField().WithMaxLen(50),
		definition.Field("is_active").OfType(definition.FieldTypeBool).WithLabel("Active").WithDefault(true),
		// person_type is fixed to CUSTOMER — set via BeforeHook in registration.
		definition.Field("person_type").OfType(definition.FieldTypeData).
			WithDefault("CUSTOMER").ReadOnlyField().HiddenInList(),
	},
	Hooks: []definition.HookDef{
		// Ensure person_type is always CUSTOMER regardless of payload.
		definition.BeforeHook(definition.OpCreate|definition.OpUpdate, func(_ context.Context, m *definition.Mutation) error {
			if m.After != nil {
				m.After.Set("person_type", "CUSTOMER")
			}
			return nil
		}),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowCRM),
	},
}

// Contract is a signed agreement with a counterparty.
var Contract = &definition.EntityDefinition{
	Name:       "contract",
	Label:      "Contract",
	Table:      "contracts",
	Module:     "CRM",
	SoftDelete: true,
	Audited:    true,
	Fields: []*definition.FieldDef{
		definition.Field("number").OfType(definition.FieldTypeSmallText).WithLabel("Contract #").
			RequiredField().UniqueField().SearchableField().WithMaxLen(50),
		definition.Field("title").OfType(definition.FieldTypeSmallText).WithLabel("Title").
			RequiredField().SearchableField().WithMaxLen(255),
		definition.Field("status").OfType(definition.FieldTypeSelect).WithLabel("Status").
			WithOptions("DRAFT", "ACTIVE", "EXPIRED", "TERMINATED", "PENDING_APPROVAL").
			WithDefault("DRAFT"),
		definition.Field("contract_type").OfType(definition.FieldTypeSmallText).WithLabel("Type").
			WithMaxLen(50),
		definition.Field("counterparty_name").OfType(definition.FieldTypeSmallText).WithLabel("Counterparty").
			RequiredField().SearchableField().WithMaxLen(255),
		definition.Field("counterparty_email").OfType(definition.FieldTypeSmallText).WithLabel("Counterparty Email").
			WithMaxLen(320),
		definition.Field("start_date").OfType(definition.FieldTypeDate).WithLabel("Start Date"),
		definition.Field("end_date").OfType(definition.FieldTypeDate).WithLabel("End Date"),
		definition.Field("value").OfType(definition.FieldTypeCurrency).WithLabel("Value"),
		definition.Field("currency_code").OfType(definition.FieldTypeSmallText).WithLabel("Currency").
			WithMaxLen(3).WithDefault("USD"),
		definition.Field("description").OfType(definition.FieldTypeLongText).WithLabel("Description"),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowCRM),
	},
}
