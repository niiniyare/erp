package users

import (
	"awo.so/internal/web/amis"
	dslscreens "awo.so/internal/web/dsl/screens"
	"awo.so/internal/web/registry"
	"awo.so/internal/web/ui"
)

func init() {
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/users",
		Module:      "iam",
		Title:       "Users",
		Description: "User management — list, create, edit, deactivate",
		Fn:          Schema,
		ASTFn: func(sess ui.UISessionContext) any {
			return dslscreens.UsersListScreen(sess)
		},
	})
}

var mockUsers = amis.A{
	amis.M{"id": "usr-001", "username": "alice.wanjiku", "email": "alice@awoerp.com", "role": "admin", "status": "ACTIVE", "created_at": "2026-01-10"},
	amis.M{"id": "usr-002", "username": "brian.otieno", "email": "brian@awoerp.com", "role": "manager", "status": "ACTIVE", "created_at": "2026-01-15"},
	amis.M{"id": "usr-003", "username": "carol.maina", "email": "carol@awoerp.com", "role": "staff", "status": "ACTIVE", "created_at": "2026-02-01"},
	amis.M{"id": "usr-004", "username": "david.kamau", "email": "david@awoerp.com", "role": "staff", "status": "INACTIVE", "created_at": "2026-02-10"},
	amis.M{"id": "usr-005", "username": "eve.njeri", "email": "eve@awoerp.com", "role": "manager", "status": "ACTIVE", "created_at": "2026-03-05"},
	amis.M{"id": "usr-006", "username": "frank.ouma", "email": "frank@awoerp.com", "role": "staff", "status": "SUSPENDED", "created_at": "2026-03-20"},
}

func Schema(_ ui.UISessionContext) ui.Schema {
	return amis.Page("").
		Toolbar(
			amis.M{
				"type": "button", "label": "New User", "icon": "fa fa-plus",
				"level": "primary", "actionType": "dialog",
				"dialog": amis.M{
					"title": "New User",
					"body": amis.Form("post:/api/v1/users").Fields(
						amis.Required(amis.TextField("username", "Username")),
						amis.Required(amis.TextField("email", "Email")),
						amis.Required(amis.TextField("password", "Password")),
						amis.SelectField("role", "Role",
							amis.SelectOpt("Admin", "admin"),
							amis.SelectOpt("Manager", "manager"),
							amis.SelectOpt("Staff", "staff"),
						),
					).Build(),
				},
			},
			amis.M{"type": "button", "label": "Export CSV", "icon": "fa fa-download"},
		).
		Body([]any{
			amis.Breadcrumb(
				amis.BC("Home", "#dashboard"),
				amis.BC("Users"),
			),
			amis.CRUD("get:/api/v1/users").
				StaticData(mockUsers).
				Filter(amis.Form("").Fields(
					amis.TextField("username", "Username"),
					amis.SelectField("role", "Role",
						amis.SelectOpt("All", ""),
						amis.SelectOpt("Admin", "admin"),
						amis.SelectOpt("Manager", "manager"),
						amis.SelectOpt("Staff", "staff"),
					),
					amis.SelectField("status", "Status",
						amis.SelectOpt("All", ""),
						amis.SelectOpt("Active", "ACTIVE"),
						amis.SelectOpt("Inactive", "INACTIVE"),
						amis.SelectOpt("Suspended", "SUSPENDED"),
					),
				).Build()).
				Columns(
					amis.Column("username", "Username").Sortable(),
					amis.Column("email", "Email"),
					amis.Column("role", "Role").Map(amis.M{
						"admin":   "<span class='label label-primary'>Admin</span>",
						"manager": "<span class='label label-info'>Manager</span>",
						"staff":   "<span class='label label-default'>Staff</span>",
					}),
					amis.Column("status", "Status").Map(amis.M{
						"ACTIVE":    "<span class='label label-success'>Active</span>",
						"INACTIVE":  "<span class='label label-warning'>Inactive</span>",
						"SUSPENDED": "<span class='label label-danger'>Suspended</span>",
					}),
					amis.Column("created_at", "Created").Type("date").Sortable(),
					amis.Column("", "Actions").Buttons(
						amis.EditBtn("put:/api/v1/users/${id}",
							amis.Required(amis.TextField("username", "Username")),
							amis.Required(amis.TextField("email", "Email")),
							amis.SelectField("role", "Role",
								amis.SelectOpt("Admin", "admin"),
								amis.SelectOpt("Manager", "manager"),
								amis.SelectOpt("Staff", "staff"),
							),
						),
						amis.DeleteBtn("delete:/api/v1/users/${id}"),
					),
				).
				DefaultSort("created_at", "desc").
				PerPage(20).
				Build(),
		}).Build()
}
