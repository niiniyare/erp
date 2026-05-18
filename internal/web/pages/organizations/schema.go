package organizations

import (
	"awo.so/internal/web/amis"
	"awo.so/internal/web/registry"
	"awo.so/internal/web/ui"
)

func init() {
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/organizations",
		Module:      "tenant",
		Title:       "Organizations",
		Description: "Tenant / organization management",
		Fn:          Schema,
	})
}

var mockTenants = amis.A{
	amis.M{"id": "ten-001", "name": "Savanna Tech Ltd", "slug": "savanna-tech", "industry": "Technology", "company_size": "MEDIUM", "status": "ACTIVE", "subscription_tier": "BUSINESS", "created_at": "2025-11-01"},
	amis.M{"id": "ten-002", "name": "Baobab Supplies Co", "slug": "baobab-supplies", "industry": "Retail", "company_size": "SMALL", "status": "ACTIVE", "subscription_tier": "STARTER", "created_at": "2025-12-15"},
	amis.M{"id": "ten-003", "name": "Kilimanjaro Finance", "slug": "kilimanjaro-finance", "industry": "Financial Services", "company_size": "LARGE", "status": "ACTIVE", "subscription_tier": "ENTERPRISE", "created_at": "2026-01-08"},
	amis.M{"id": "ten-004", "name": "Serengeti Farms", "slug": "serengeti-farms", "industry": "Agriculture", "company_size": "SMALL", "status": "PENDING", "subscription_tier": "FREE", "created_at": "2026-03-01"},
	amis.M{"id": "ten-005", "name": "Nairobi Digital Media", "slug": "nairobi-digital", "industry": "Media", "company_size": "MEDIUM", "status": "SUSPENDED", "subscription_tier": "STARTER", "created_at": "2026-01-20"},
}

func Schema(_ ui.UISessionContext) ui.Schema {
	return amis.Page("").
		Toolbar(
			amis.M{
				"type": "button", "label": "New Tenant", "icon": "fa fa-plus",
				"level": "primary", "actionType": "dialog",
				"dialog": amis.M{
					"title": "New Tenant",
					"body": amis.Form("post:/api/v1/tenants").Fields(
						amis.Required(amis.TextField("name", "Company Name")),
						amis.Required(amis.TextField("email", "Admin Email")),
						amis.TextField("industry", "Industry"),
						amis.SelectField("company_size", "Size",
							amis.SelectOpt("1–10", "SMALL"),
							amis.SelectOpt("11–50", "MEDIUM"),
							amis.SelectOpt("51–200", "LARGE"),
							amis.SelectOpt("200+", "ENTERPRISE"),
						),
					).Build(),
				},
			},
			amis.M{"type": "button", "label": "Export CSV", "icon": "fa fa-download"},
		).
		Body([]any{
			amis.Breadcrumb(
				amis.BC("Home", "#dashboard"),
				amis.BC("Tenants"),
			),
			amis.CRUD("get:/api/v1/tenants").
				StaticData(mockTenants).
				Filter(amis.Form("").Fields(
					amis.TextField("name", "Company Name"),
					amis.SelectField("status", "Status",
						amis.SelectOpt("All", ""),
						amis.SelectOpt("Pending", "PENDING"),
						amis.SelectOpt("Active", "ACTIVE"),
						amis.SelectOpt("Suspended", "SUSPENDED"),
					),
					amis.SelectField("subscription_tier", "Plan",
						amis.SelectOpt("All", ""),
						amis.SelectOpt("Free", "FREE"),
						amis.SelectOpt("Starter", "STARTER"),
						amis.SelectOpt("Business", "BUSINESS"),
						amis.SelectOpt("Enterprise", "ENTERPRISE"),
					),
				).Build()).
				Columns(
					amis.Column("name", "Company").Sortable(),
					amis.Column("industry", "Industry"),
					amis.Column("company_size", "Size").Map(amis.M{
						"SMALL":      "1–10",
						"MEDIUM":     "11–50",
						"LARGE":      "51–200",
						"ENTERPRISE": "200+",
					}).Align("center"),
					amis.Column("status", "Status").Map(amis.M{
						"PENDING":   "<span class='label label-warning'>Pending</span>",
						"ACTIVE":    "<span class='label label-success'>Active</span>",
						"SUSPENDED": "<span class='label label-danger'>Suspended</span>",
						"ARCHIVED":  "<span class='label label-default'>Archived</span>",
					}),
					amis.Column("subscription_tier", "Plan").Map(amis.M{
						"FREE":       "<span class='label label-default'>Free</span>",
						"STARTER":    "<span class='label label-info'>Starter</span>",
						"BUSINESS":   "<span class='label label-primary'>Business</span>",
						"ENTERPRISE": "<span class='label label-success'>Enterprise</span>",
					}),
					amis.Column("created_at", "Created").Type("date").Sortable(),
					amis.Column("", "Actions").Buttons(
						amis.ViewBtn(tenantDetail()),
						amis.EditBtn("put:/api/v1/tenants/${id}",
							amis.Required(amis.TextField("name", "Company Name")),
							amis.TextField("industry", "Industry"),
							amis.SelectField("status", "Status",
								amis.SelectOpt("Active", "ACTIVE"),
								amis.SelectOpt("Suspended", "SUSPENDED"),
							),
						),
					),
				).
				DefaultSort("created_at", "desc").
				PerPage(20).
				Build(),
		}).Build()
}

func tenantDetail() amis.Schema {
	return amis.Descriptions("").
		Item("Company", "name").
		Item("Slug", "slug").
		Item("Industry", "industry").
		Item("Size", "company_size").
		Item("Status", "status").
		Item("Plan", "subscription_tier").
		Item("Created", "created_at").
		Build()
}
