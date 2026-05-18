package accounts

import (
	"awo.so/internal/web/amis"
	"awo.so/internal/web/registry"
	"awo.so/internal/web/ui"
)

func init() {
	registry.Register("/finance/accounts", Schema)
}

var mockAccounts = amis.A{
	amis.M{"id": "acc-001", "code": "1000", "name": "Cash and Bank", "account_type": "ASSET", "currency": "KES", "balance": 2500000, "is_active": true},
	amis.M{"id": "acc-002", "code": "1100", "name": "Accounts Receivable", "account_type": "ASSET", "currency": "KES", "balance": 850000, "is_active": true},
	amis.M{"id": "acc-003", "code": "1200", "name": "Inventory", "account_type": "ASSET", "currency": "KES", "balance": 1200000, "is_active": true},
	amis.M{"id": "acc-004", "code": "2000", "name": "Accounts Payable", "account_type": "LIABILITY", "currency": "KES", "balance": 320000, "is_active": true},
	amis.M{"id": "acc-005", "code": "2100", "name": "Bank Loan", "account_type": "LIABILITY", "currency": "KES", "balance": 5000000, "is_active": true},
	amis.M{"id": "acc-006", "code": "3000", "name": "Owner Equity", "account_type": "EQUITY", "currency": "KES", "balance": 1500000, "is_active": true},
	amis.M{"id": "acc-007", "code": "4000", "name": "Sales Revenue", "account_type": "REVENUE", "currency": "KES", "balance": 3800000, "is_active": true},
	amis.M{"id": "acc-008", "code": "5000", "name": "Cost of Goods Sold", "account_type": "EXPENSE", "currency": "KES", "balance": 2100000, "is_active": true},
}

func Schema(_ ui.UISessionContext) ui.Schema {
	return amis.Page("").
		Toolbar(
			amis.M{
				"type": "button", "label": "New Account", "icon": "fa fa-plus",
				"level": "primary", "actionType": "dialog",
				"dialog": amis.M{
					"title": "New Account",
					"body": amis.Form("post:/api/v1/finance/accounts").Fields(
						amis.Required(amis.TextField("code", "Account Code")),
						amis.Required(amis.TextField("name", "Account Name")),
						amis.Required(amis.SelectField("account_type", "Type",
							amis.SelectOpt("Asset", "ASSET"),
							amis.SelectOpt("Liability", "LIABILITY"),
							amis.SelectOpt("Equity", "EQUITY"),
							amis.SelectOpt("Revenue", "REVENUE"),
							amis.SelectOpt("Expense", "EXPENSE"),
						)),
						amis.SelectField("currency", "Currency",
							amis.SelectOpt("KES", "KES"),
							amis.SelectOpt("USD", "USD"),
							amis.SelectOpt("EUR", "EUR"),
						),
						amis.Optional(amis.TextAreaField("description", "Description")),
					).Build(),
				},
			},
			amis.M{"type": "button", "label": "Export CSV", "icon": "fa fa-download"},
			amis.M{"type": "button", "label": "Refresh", "icon": "fa fa-rotate", "actionType": "reload", "target": "accounts-table"},
		).
		Body([]any{
			amis.Breadcrumb(
				amis.BC("Home", "#dashboard"),
				amis.BC("Finance"),
				amis.BC("Chart of Accounts"),
			),
			amis.CRUD("get:/api/v1/finance/accounts").
				StaticData(mockAccounts).
				Filter(amis.Form("").Fields(
					amis.TextField("name", "Account Name"),
					amis.SelectField("account_type", "Type",
						amis.SelectOpt("All", ""),
						amis.SelectOpt("Asset", "ASSET"),
						amis.SelectOpt("Liability", "LIABILITY"),
						amis.SelectOpt("Equity", "EQUITY"),
						amis.SelectOpt("Revenue", "REVENUE"),
						amis.SelectOpt("Expense", "EXPENSE"),
					),
				).Build()).
				Columns(
					amis.Column("code", "Code").Width(80).Sortable(),
					amis.Column("name", "Account Name").Sortable(),
					amis.Column("account_type", "Type").Map(amis.M{
						"ASSET":     "<span class='label label-info'>Asset</span>",
						"LIABILITY": "<span class='label label-warning'>Liability</span>",
						"EQUITY":    "<span class='label label-primary'>Equity</span>",
						"REVENUE":   "<span class='label label-success'>Revenue</span>",
						"EXPENSE":   "<span class='label label-danger'>Expense</span>",
					}),
					amis.Column("currency", "CCY").Width(60).Align("center"),
					amis.Column("balance", "Balance").Type("number").Align("right").Sortable(),
					amis.Column("is_active", "Active").Tpl(
						"${is_active ? '<i class=\"fa fa-check text-success\"></i>' : '<i class=\"fa fa-times text-muted\"></i>'}",
					).Align("center").Width(70),
					amis.Column("", "Actions").Buttons(
						amis.EditBtn("put:/api/v1/finance/accounts/${id}",
							amis.Required(amis.TextField("code", "Code")),
							amis.Required(amis.TextField("name", "Name")),
							amis.SwitchField("is_active", "Active"),
						),
						amis.DeleteBtn("delete:/api/v1/finance/accounts/${id}"),
					),
				).
				DefaultSort("code", "asc").
				PerPage(20).
				Build(),
		}).Build()
}
