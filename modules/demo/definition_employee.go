package demo

import "awo.so/awo/def"

// EmployeeDefinition — demo employee entity demonstrating text, select, date,
// currency, and JSON (org_unit tree) widgets.
var EmployeeDefinition = def.SystemDefinition{
	Name:        "employee",
	Module:      "demo",
	Label:       "Employee",
	LabelPlural: "Employees",
	Description: "Demo employee entity showcasing text, select, date, salary, and org tree widgets.",
	Icon:        "user",
	Fields: []def.FieldDef{
		{
			Name:        "name",
			Type:        def.FieldTypeData,
			Label:       "Full Name",
			Required:    true,
			Searchable:  true,
			MaxLen:      255,
			Placeholder: "Employee full name",
			Icon:        "user",
		},
		{
			Name:     "employee_id",
			Type:     def.FieldTypeData,
			Label:    "Employee ID",
			Required: true,
			Unique:   true,
			MaxLen:   50,
			Icon:     "search",
		},
		{
			Name:        "email",
			Type:        def.FieldTypeData,
			Label:       "Email",
			Searchable:  true,
			MaxLen:      255,
			Placeholder: "employee@company.com",
			Icon:        "mail",
		},
		{
			Name:    "department",
			Type:    def.FieldTypeSelect,
			Label:   "Department",
			Options: []string{"engineering", "sales", "hr", "finance", "operations"},
			Icon:    "building",
		},
		{
			Name:  "start_date",
			Type:  def.FieldTypeDate,
			Label: "Start Date",
			Icon:  "calendar",
		},
		{
			Name:  "salary",
			Type:  def.FieldTypeCurrency,
			Label: "Monthly Salary",
			Icon:  "money",
		},
		{
			Name:  "consent_signature",
			Type:  def.FieldTypeData,
			Label: "Consent Signature (base64)",
			MaxLen: 10000,
		},
		{
			Name:  "org_unit",
			Type:  def.FieldTypeJSON,
			Label: "Org Unit",
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"demo.employee.create"},
		Read:   []string{"demo.employee.read"},
		Write:  []string{"demo.employee.update"},
		Delete: []string{"demo.employee.delete"},
	},
}

func init() {
	def.Register(&EmployeeDefinition)
}
