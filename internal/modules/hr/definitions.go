// Package hr declares EntityDefinitions for the Human Resources module.
package hr

import (
	"context"

	"awo.so/framework/definition"
)

func init() {
	definition.Register(Person)
	definition.Register(Employee)
}

func allowHR(_ context.Context, viewer definition.ViewerContext, _ definition.Op, _ definition.Record) error {
	if viewer.IsSystem() || viewer.HasRole("hr_manager") || viewer.HasRole("tenant_admin") {
		return definition.ErrAllow
	}
	return definition.ErrDeny
}

// Person is the unified party record (employee, customer, vendor, contact).
var Person = &definition.EntityDefinition{
	Name:       "person",
	Label:      "Person",
	Table:      "persons",
	Module:     "HR",
	SoftDelete: true,
	Audited:    true,
	Fields: []*definition.FieldDef{
		definition.Field("first_name").OfType(definition.FieldTypeSmallText).WithLabel("First Name").
			RequiredField().SearchableField().WithMaxLen(100),
		definition.Field("last_name").OfType(definition.FieldTypeSmallText).WithLabel("Last Name").
			RequiredField().SearchableField().WithMaxLen(100),
		definition.Field("middle_name").OfType(definition.FieldTypeSmallText).WithLabel("Middle Name").
			WithMaxLen(100),
		definition.Field("email").OfType(definition.FieldTypeSmallText).WithLabel("Email").
			WithMaxLen(320),
		definition.Field("phone").OfType(definition.FieldTypeSmallText).WithLabel("Phone").
			WithMaxLen(30),
		definition.Field("person_type").OfType(definition.FieldTypeSelect).WithLabel("Type").
			RequiredField().
			WithOptions("INDIVIDUAL", "EMPLOYEE", "CONTACT", "CUSTOMER", "VENDOR", "CONTRACTOR").
			WithDefault("INDIVIDUAL"),
		definition.Field("birth_date").OfType(definition.FieldTypeDate).WithLabel("Date of Birth").
			SensitiveField(),
		definition.Field("national_id").OfType(definition.FieldTypeSmallText).WithLabel("National ID").
			SensitiveField().WithMaxLen(50),
		definition.Field("tax_id").OfType(definition.FieldTypeSmallText).WithLabel("Tax ID").
			SensitiveField().WithMaxLen(50),
		definition.Field("is_active").OfType(definition.FieldTypeBool).WithLabel("Active").WithDefault(true),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowHR),
	},
}

// Employee extends Person with employment-specific fields.
var Employee = &definition.EntityDefinition{
	Name:       "employee",
	Label:      "Employee",
	Table:      "employees",
	Module:     "HR",
	SoftDelete: true,
	Audited:    true,
	Fields: []*definition.FieldDef{
		definition.Field("person_id").OfType(definition.FieldTypeLink).WithLabel("Person").
			RequiredField().ImmutableField().LinksTo("person"),
		definition.Field("employee_number").OfType(definition.FieldTypeSmallText).WithLabel("Employee #").
			RequiredField().UniqueField().SearchableField().WithMaxLen(30),
		definition.Field("position_title").OfType(definition.FieldTypeSmallText).WithLabel("Position").
			WithMaxLen(150),
		definition.Field("employment_status").OfType(definition.FieldTypeSelect).WithLabel("Status").
			RequiredField().
			WithOptions("ACTIVE", "INACTIVE", "TERMINATED", "ON_LEAVE", "SUSPENDED").
			WithDefault("ACTIVE"),
		definition.Field("hire_date").OfType(definition.FieldTypeDate).WithLabel("Hire Date").RequiredField(),
		definition.Field("termination_date").OfType(definition.FieldTypeDate).WithLabel("Termination Date"),
		definition.Field("manager_id").OfType(definition.FieldTypeLink).WithLabel("Manager").
			LinksTo("employee"),
		definition.Field("security_level").OfType(definition.FieldTypeInt).WithLabel("Security Level").
			WithDefault(1),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowHR),
	},
}
