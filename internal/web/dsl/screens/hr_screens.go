package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// EmployeeScreen renders the employee detail form.
func EmployeeScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Employee",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/hr/employees/:id",
			SendOn: "${params.id}",
		},
		Body: []ast.Node{
			ast.SectionNode{
				Title: "Personal Information",
				Body: []ast.Node{
					ast.InputTextNode{Name: "first_name", Label: "First Name", Required: true},
					ast.InputTextNode{Name: "last_name", Label: "Last Name", Required: true},
					ast.InputTextNode{Name: "national_id", Label: "National ID"},
					ast.InputDateNode{Name: "date_of_birth", Label: "Date of Birth"},
					ast.SelectNode{
						Name:    "gender",
						Label:   "Gender",
						Options: []ast.SelectOption{{Label: "Male", Value: "M"}, {Label: "Female", Value: "F"}, {Label: "Other", Value: "O"}},
					},
				},
			},
			ast.SectionNode{
				Title: "Employment Details",
				Body: []ast.Node{
					ast.InputTextNode{Name: "employee_number", Label: "Employee #"},
					ast.SelectNode{
						Name:     "department_id",
						Label:    "Department",
						Required: true,
						Source:   &ast.APISpec{Method: "get", URL: "/api/v1/hr/departments/options"},
					},
					ast.SelectNode{
						Name:   "position_id",
						Label:  "Position",
						Source: &ast.APISpec{Method: "get", URL: "/api/v1/hr/positions/options"},
					},
					ast.InputDateNode{Name: "start_date", Label: "Start Date", Required: true},
					ast.InputDateNode{Name: "contract_end_date", Label: "Contract End"},
					ast.SelectNode{
						Name:    "employment_type",
						Label:   "Employment Type",
						Options: []ast.SelectOption{
							{Label: "Full-time", Value: "full_time"},
							{Label: "Part-time", Value: "part_time"},
							{Label: "Contract", Value: "contract"},
							{Label: "Intern", Value: "intern"},
						},
					},
					ast.SelectNode{
						Name:    "status",
						Label:   "Status",
						Options: []ast.SelectOption{
							{Label: "Active", Value: "active"},
							{Label: "On Leave", Value: "on_leave"},
							{Label: "Suspended", Value: "suspended"},
							{Label: "Terminated", Value: "terminated"},
						},
					},
				},
			},
			blocks.AttachmentsBlock(sess),
		},
	}
}

// LeaveRequestScreen renders the leave request form.
func LeaveRequestScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "New Leave Request",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/hr/leave-requests/:id",
			SendOn: "${params.id}",
		},
		Body: []ast.Node{
			ast.SectionNode{
				Title: "Leave Details",
				Body: []ast.Node{
					blocks.PartyBlock(sess, blocks.DefaultEmployeeConfig()),
					ast.SelectNode{
						Name:     "leave_type_id",
						Label:    "Leave Type",
						Required: true,
						Source:   &ast.APISpec{Method: "get", URL: "/api/v1/hr/leave-types/options"},
					},
					ast.InputDateRangeNode{Name: "date_range", Label: "Date Range", Required: true},
					ast.InputTextNode{Name: "reason", Label: "Reason", Required: true},
					ast.SelectNode{
						Name:    "status",
						Label:   "Status",
						Options: []ast.SelectOption{
							{Label: "Pending", Value: "pending"},
							{Label: "Approved", Value: "approved"},
							{Label: "Rejected", Value: "rejected"},
							{Label: "Cancelled", Value: "cancelled"},
						},
					},
				},
			},
			ast.SectionNode{
				Title: "Approval Chain",
				Body: []ast.Node{
					ast.TimelineNode{
						API: &ast.APISpec{Method: "get", URL: "/api/v1/hr/leave-requests/:id/approvals"},
					},
				},
			},
			blocks.AttachmentsBlock(sess),
		},
	}
}

// PayrollSummaryScreen renders the payroll overview dashboard.
func PayrollSummaryScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Payroll Summary",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/hr/payroll/summary",
		},
		Body: []ast.Node{
			blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
				{Label: "Headcount", ValueKey: "headcount", Format: "number"},
				{Label: "Total Payroll", ValueKey: "total_payroll", Format: "currency", Trend: blocks.TrendDown},
				{Label: "Pending Approvals", ValueKey: "pending_approvals", Format: "number"},
				{Label: "On Leave Today", ValueKey: "on_leave_today", Format: "number"},
			}),
			ast.GridNode{
				Columns: []ast.GridColumn{
					{MD: 8, Body: []ast.Node{
						blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
							Title:        "Payroll by Department",
							APIURL:       "/api/v1/hr/payroll/by-department",
							PeriodPicker: true,
						}),
					}},
					{MD: 4, Body: []ast.Node{
						blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{
							Title:    "Recent Approvals",
							Resource: "hr/leave-requests",
							Limit:    10,
						}),
					}},
				},
			},
			ast.CRUDNode{
				API: ast.APISpec{Method: "get", URL: "/api/v1/hr/payroll/runs"},
				Columns: []ast.TableColumn{
					{Name: "period", Label: "Period", Sortable: true},
					{Name: "employees", Label: "Employees", Type: "number"},
					{Name: "gross_pay", Label: "Gross Pay", Type: "number"},
					{Name: "deductions", Label: "Deductions", Type: "number"},
					{Name: "net_pay", Label: "Net Pay", Type: "number"},
					{Name: "status", Label: "Status"},
				},
				PageSize: 12,
			},
		},
	}
}
