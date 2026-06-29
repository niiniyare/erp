package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// RoleScreen renders the role detail form: name, description, permissions, members.
func RoleScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Role",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/iam/roles/:id",
			SendOn: "${params.id}",
		},
		Body: []ast.Node{
			ast.SectionNode{
				Title: "Role Details",
				Body: []ast.Node{
					ast.InputTextNode{Name: "name", Label: "Role Name", Required: true},
					ast.InputTextNode{Name: "code", Label: "Role Code", Required: true},
					ast.InputTextNode{Name: "description", Label: "Description"},
					ast.SelectNode{
						Name:    "status",
						Label:   "Status",
						Options: []ast.SelectOption{{Label: "Active", Value: "active"}, {Label: "Inactive", Value: "inactive"}},
					},
				},
			},
			ast.SectionNode{
				Title: "Permissions",
				Body: []ast.Node{
					ast.MultiSelectNode{
						Name:       "permissions",
						Label:      "Assigned Permissions",
						Source:     &ast.APISpec{Method: "get", URL: "/api/v1/iam/permissions/options"},
						Searchable: true,
					},
				},
			},
			ast.SectionNode{
				Title: "Members",
				Body: []ast.Node{
					ast.CRUDNode{
						API: ast.APISpec{Method: "get", URL: "/api/v1/iam/roles/${id}/members"},
						Columns: []ast.TableColumn{
							{Name: "display_name", Label: "Name"},
							{Name: "email", Label: "Email"},
							{Name: "assigned_at", Label: "Assigned", Type: "date"},
							{Name: "expires_at", Label: "Expires", Type: "date"},
						},
						RowActions: []ast.ActionNode{
							{
								Label:       "Remove",
								ActionType:  "ajax",
								API:         &ast.APISpec{Method: "delete", URL: "/api/v1/iam/roles/${roleId}/members/${id}"},
								ConfirmText: "Remove this member from the role?",
								Level:       "danger",
							},
						},
						PageSize: 20,
					},
				},
			},
			blocks.AttachmentsBlock(sess),
		},
	}
}

// PolicyScreen renders the Casbin policy list: subject, domain, object, action.
func PolicyScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Authorization Policies",
		Body: []ast.Node{
			ast.CRUDNode{
				API: ast.APISpec{Method: "get", URL: "/api/v1/iam/policies"},
				Filter: ast.FilterBarNode{
					Body: []ast.Node{
						ast.InputTextNode{Name: "subject", Label: "Subject", Placeholder: "Role or user…"},
						ast.InputTextNode{Name: "resource", Label: "Resource"},
						ast.SelectNode{
							Name:    "effect",
							Label:   "Effect",
							Options: []ast.SelectOption{{Label: "Allow", Value: "allow"}, {Label: "Deny", Value: "deny"}},
						},
					},
				},
				Toolbar: []ast.Node{
					ast.ActionNode{
						Label:      "Add Policy",
						ActionType: "dialog",
						Level:      "primary",
						VisibleOn:  `${permissions["iam.policies.create"]}`,
						Dialog: &ast.DialogNode{
							Title: "Add Policy Rule",
							Body: []ast.Node{
								ast.FormNode{
									API: &ast.APISpec{Method: "post", URL: "/api/v1/iam/policies"},
									Body: []ast.Node{
										ast.InputTextNode{Name: "subject", Label: "Subject (role/user)", Required: true},
										ast.InputTextNode{Name: "domain", Label: "Domain (tenant ID)", Required: true},
										ast.InputTextNode{Name: "resource", Label: "Resource", Required: true},
										ast.InputTextNode{Name: "action", Label: "Action", Required: true},
										ast.SelectNode{
											Name:    "effect",
											Label:   "Effect",
											Options: []ast.SelectOption{{Label: "Allow", Value: "allow"}, {Label: "Deny", Value: "deny"}},
										},
									},
								},
							},
						},
					},
				},
				Columns: []ast.TableColumn{
					{Name: "subject", Label: "Subject", Sortable: true},
					{Name: "domain", Label: "Domain"},
					{Name: "resource", Label: "Resource"},
					{Name: "action", Label: "Action"},
					{Name: "effect", Label: "Effect"},
					{Name: "created_at", Label: "Created", Type: "date", Sortable: true},
				},
				RowActions: []ast.ActionNode{
					{
						Label:       "Delete",
						ActionType:  "ajax",
						API:         &ast.APISpec{Method: "delete", URL: "/api/v1/iam/policies/${id}"},
						ConfirmText: "Delete this policy rule?",
						Level:       "danger",
					},
				},
				PageSize: 30,
			},
		},
	}
}

// UserScreen renders the user detail: profile, active roles, session history.
func UserScreen(sess ui.UISessionContext) ast.Node {
	userID := sess.Param("id", "")
	return ast.PageNode{
		Title: "User Profile",
		InitAPI: &ast.APISpec{
			Method: "get",
			URL:    "/api/v1/iam/users/" + userID,
		},
		Body: []ast.Node{
			ast.TabsNode{
				Tabs: []ast.Tab{
					{
						Title: "Profile",
						Hash:  "profile",
						Body: []ast.Node{
							ast.SectionNode{
								Title: "Identity",
								Body: []ast.Node{
									ast.InputTextNode{Name: "display_name", Label: "Display Name", Required: true},
									ast.InputTextNode{Name: "email", Label: "Email", Required: true},
									ast.SelectNode{
										Name:  "status",
										Label: "Status",
										Options: []ast.SelectOption{
											{Label: "Active", Value: "active"},
											{Label: "Suspended", Value: "suspended"},
											{Label: "Archived", Value: "archived"},
										},
									},
								},
							},
						},
					},
					{
						Title: "Roles",
						Hash:  "roles",
						Body: []ast.Node{
							ast.CRUDNode{
								API: ast.APISpec{Method: "get", URL: "/api/v1/iam/users/" + userID + "/roles"},
								Columns: []ast.TableColumn{
									{Name: "role_name", Label: "Role"},
									{Name: "scope", Label: "Scope"},
									{Name: "assigned_at", Label: "Assigned", Type: "date"},
									{Name: "expires_at", Label: "Expires", Type: "date"},
								},
								PageSize: 20,
							},
						},
					},
					{
						Title: "Session History",
						Hash:  "sessions",
						Body: []ast.Node{
							ast.TimelineNode{
								API: &ast.APISpec{Method: "get", URL: "/api/v1/iam/users/" + userID + "/sessions"},
							},
						},
					},
				},
				MountOnEnter:  true,
				UnmountOnExit: false,
			},
		},
	}
}
