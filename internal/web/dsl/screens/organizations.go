package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// OrganizationsScreen builds the tenant / organization management listing page.
// Wires to /api/v1/tenants with platform-gated create and lifecycle status badges.
func OrganizationsScreen(sess ui.UISessionContext) ast.Node {
	return ast.PageNode{
		Title: "Organizations",
		Body: []ast.Node{
			blocks.DataTableBlock(sess, blocks.DataTableConfig{
				APIURL:           "/api/v1/tenants",
				Title:            "Organizations",
				PrimaryKey:       "id",
				PageSize:         20,
				AllowCreate:      sess.IsPlatform, // platform admins only
				CreateURL:        "/organizations/new",
				CreatePermission: "platform.tenants.create",
				AllowExport:      true,
				Filter: blocks.FilterBarBlock(sess, blocks.FilterBarConfig{
					ShowSearch:        true,
					SearchPlaceholder: "Search organizations…",
					ShowStatus:        true,
					StatusOptions: []ast.SelectOption{
						{Label: "Pending", Value: "PENDING"},
						{Label: "Active", Value: "ACTIVE"},
						{Label: "Suspended", Value: "SUSPENDED"},
						{Label: "Archived", Value: "ARCHIVED"},
					},
					ShowTypeFilter: true,
					TypeFieldName:  "subscription_tier",
					TypeLabel:      "Plan",
					TypeOptions: []ast.SelectOption{
						{Label: "Free", Value: "FREE"},
						{Label: "Starter", Value: "STARTER"},
						{Label: "Business", Value: "BUSINESS"},
						{Label: "Enterprise", Value: "ENTERPRISE"},
					},
				}),
				Columns: []blocks.ColumnDef{
					{Name: "name", Label: "Company", Sortable: true},
					{Name: "industry", Label: "Industry"},
					{
						Name:  "company_size",
						Label: "Size",
						Type:  "mapping",
						Map: map[string]string{
							"SMALL":      "1–10",
							"MEDIUM":     "11–50",
							"LARGE":      "51–200",
							"ENTERPRISE": "200+",
							"*":          "${value}",
						},
					},
					blocks.StatusBadgeColumnDef(blocks.StatusBadgeConfig{
						FieldName: "status",
						Label:     "Status",
						Mappings: []blocks.StatusMapping{
							{Value: "PENDING", Label: "Pending", Color: "warning"},
							{Value: "ACTIVE", Label: "Active", Color: "success"},
							{Value: "SUSPENDED", Label: "Suspended", Color: "danger"},
							{Value: "ARCHIVED", Label: "Archived", Color: "default"},
						},
					}),
					{
						Name:  "subscription_tier",
						Label: "Plan",
						Type:  "mapping",
						Map: map[string]string{
							"FREE":       `<span class="badge badge-default">Free</span>`,
							"STARTER":    `<span class="badge badge-info">Starter</span>`,
							"BUSINESS":   `<span class="badge badge-primary">Business</span>`,
							"ENTERPRISE": `<span class="badge badge-success">Enterprise</span>`,
							"*":          `<span class="badge badge-default">${value}</span>`,
						},
					},
					{Name: "created_at", Label: "Created", Type: "date", Sortable: true},
				},
				RowActions: []ast.ActionNode{
					{
						Label:      "View",
						ActionType: "link",
						Target:     "/organizations/${id}",
						Icon:       "fa fa-eye",
					},
					{
						Label:      "Edit",
						ActionType: "link",
						Target:     "/organizations/${id}/edit",
						Icon:       "fa fa-edit",
					},
				},
			}),
		},
	}
}
