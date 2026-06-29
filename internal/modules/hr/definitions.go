// Package hr declares EntityDefinitions for the Human Resources module.
package hr

import (
	"context"

	"awo.so/framework/def"
)

func init() {
	def.Register(Person)
	def.Register(Employee)
}

func allowHR(_ context.Context, viewer def.ViewerContext, _ def.Op, _ def.Record) error {
	if viewer.IsSystem() || viewer.HasRole("hr_manager") || viewer.HasRole("tenant_admin") {
		return def.ErrAllow
	}
	return def.ErrDeny
}

// Person is the unified party record (employee, customer, vendor, contact).
var Person = &def.EntityDefinition{
	Name:       "person",
	Label:      "Person",
	Table:      "persons",
	Module:     "HR",
	SoftDelete: true,
	Audited:    true,
	Fields: []*def.FieldDef{
		def.Field("first_name").OfType(def.FieldTypeSmallText).WithLabel("First Name").
			RequiredField().SearchableField().WithMaxLen(100),
		def.Field("last_name").OfType(def.FieldTypeSmallText).WithLabel("Last Name").
			RequiredField().SearchableField().WithMaxLen(100),
		def.Field("middle_name").OfType(def.FieldTypeSmallText).WithLabel("Middle Name").
			WithMaxLen(100),
		def.Field("email").OfType(def.FieldTypeSmallText).WithLabel("Email").
			WithMaxLen(320),
		def.Field("phone").OfType(def.FieldTypeSmallText).WithLabel("Phone").
			WithMaxLen(30),
		def.Field("person_type").OfType(def.FieldTypeSelect).WithLabel("Type").
			RequiredField().
			WithOptions("INDIVIDUAL", "EMPLOYEE", "CONTACT", "CUSTOMER", "VENDOR", "CONTRACTOR").
			WithDefault("INDIVIDUAL"),
		def.Field("birth_date").OfType(def.FieldTypeDate).WithLabel("Date of Birth").
			SensitiveField(),
		def.Field("national_id").OfType(def.FieldTypeSmallText).WithLabel("National ID").
			SensitiveField().WithMaxLen(50),
		def.Field("tax_id").OfType(def.FieldTypeSmallText).WithLabel("Tax ID").
			SensitiveField().WithMaxLen(50),
		def.Field("is_active").OfType(def.FieldTypeBool).WithLabel("Active").WithDefault(true),
	},
	Policies: []def.PolicyDef{
		def.Policy(def.OpAll, allowHR),
	},
}

// Employee extends Person with employment-specific fields.
var Employee = &def.EntityDefinition{
	Name:       "employee",
	Label:      "Employee",
	Table:      "employees",
	Module:     "HR",
	SoftDelete: true,
	Audited:    true,
	Fields: []*def.FieldDef{
		def.Field("person_id").OfType(def.FieldTypeLink).WithLabel("Person").
			RequiredField().ImmutableField().LinksTo("person"),
		def.Field("employee_number").OfType(def.FieldTypeSmallText).WithLabel("Employee #").
			RequiredField().UniqueField().SearchableField().WithMaxLen(30),
		def.Field("position_title").OfType(def.FieldTypeSmallText).WithLabel("Position").
			WithMaxLen(150),
		def.Field("employment_status").OfType(def.FieldTypeSelect).WithLabel("Status").
			RequiredField().
			WithOptions("ACTIVE", "INACTIVE", "TERMINATED", "ON_LEAVE", "SUSPENDED").
			WithDefault("ACTIVE"),
		def.Field("hire_date").OfType(def.FieldTypeDate).WithLabel("Hire Date").RequiredField(),
		def.Field("termination_date").OfType(def.FieldTypeDate).WithLabel("Termination Date"),
		def.Field("manager_id").OfType(def.FieldTypeLink).WithLabel("Manager").
			LinksTo("employee"),
		def.Field("security_level").OfType(def.FieldTypeInt).WithLabel("Security Level").
			WithDefault(1),
	},
	Policies: []def.PolicyDef{
		def.Policy(def.OpAll, allowHR),
	},
}
