package builders

import "awo.so/internal/web/ast"

// EmployeePickerNode returns a Select node for choosing an employee.
func EmployeePickerNode(name, label string, required bool) ast.Node {
	return ast.SelectNode{
		Name:       name,
		Label:      label,
		Required:   required,
		Source:     &ast.APISpec{Method: "get", URL: "/api/v1/hr/employees/options"},
		Searchable: true,
	}
}

// DepartmentPickerNode returns a Select node for choosing a department.
func DepartmentPickerNode(name, label string) ast.Node {
	return ast.SelectNode{
		Name:      name,
		Label:     label,
		Source:    &ast.APISpec{Method: "get", URL: "/api/v1/hr/departments/options"},
		Clearable: true,
	}
}

// JobPositionPickerNode returns a Select node for choosing a job position/role.
func JobPositionPickerNode(name, label string) ast.Node {
	return ast.SelectNode{
		Name:      name,
		Label:     label,
		Source:    &ast.APISpec{Method: "get", URL: "/api/v1/hr/positions/options"},
		Clearable: true,
	}
}
