package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// UsersListScreen builds the IAM user management listing page.
// Wires to /api/v1/users with permission-gated create button and
// role+status filter bar.
func UsersListScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Users",
		Body: []ast.Node{
			blocks.DataTableBlock(sess, blocks.DataTableConfig{
				APIURL:           "/api/v1/users",
				Title:            "Users",
				PrimaryKey:       "id",
				PageSize:         20,
				AllowCreate:      true,
				CreateURL:        "/users/new",
				CreatePermission: "iam.users.create",
				AllowExport:      true,
				Filter: blocks.FilterBarBlock(sess, blocks.FilterBarConfig{
					ShowSearch:        true,
					SearchPlaceholder: "Search users…",
					ShowStatus:        true,
					StatusOptions: []ast.SelectOption{
						{Label: "Active", Value: "ACTIVE"},
						{Label: "Inactive", Value: "INACTIVE"},
						{Label: "Suspended", Value: "SUSPENDED"},
					},
					ShowTypeFilter: true,
					TypeFieldName:  "role",
					TypeLabel:      "Role",
					TypeOptions: []ast.SelectOption{
						{Label: "Admin", Value: "admin"},
						{Label: "Manager", Value: "manager"},
						{Label: "Staff", Value: "staff"},
					},
				}),
				Columns: []blocks.ColumnDef{
					{Name: "username", Label: "Username", Sortable: true},
					{Name: "email", Label: "Email"},
					{
						Name:  "role",
						Label: "Role",
						Type:  "mapping",
						Map: map[string]string{
							"admin":   `<span class="badge badge-primary">Admin</span>`,
							"manager": `<span class="badge badge-info">Manager</span>`,
							"staff":   `<span class="badge badge-default">Staff</span>`,
							"*":       `<span class="badge badge-default">${value}</span>`,
						},
					},
					blocks.StatusBadgeColumnDef(blocks.StatusBadgeConfig{
						FieldName: "status",
						Label:     "Status",
						Mappings: []blocks.StatusMapping{
							{Value: "ACTIVE", Label: "Active", Color: "success"},
							{Value: "INACTIVE", Label: "Inactive", Color: "warning"},
							{Value: "SUSPENDED", Label: "Suspended", Color: "danger"},
						},
					}),
					{Name: "created_at", Label: "Created", Type: "date", Sortable: true},
				},
				RowActions: []ast.ActionNode{
					{
						Label:      "Edit",
						ActionType: "link",
						Target:     "/users/${id}",
						Icon:       "fa fa-edit",
					},
					{
						Label:       "Delete",
						ActionType:  "ajax",
						Level:       "danger",
						Icon:        "fa fa-trash",
						ConfirmText: "Permanently delete this user?",
						API:         &ast.APISpec{Method: "delete", URL: "/api/v1/users/${id}"},
					},
				},
			}),
		},
	}
}
