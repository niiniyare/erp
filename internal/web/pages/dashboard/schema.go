package dashboard

import (
	"awo.so/internal/web/amis"
	"awo.so/internal/web/registry"
	"awo.so/internal/web/ui"
)

func init() {
	registry.RegisterPage(registry.PageRegistration{
		Route:       "/dashboard",
		Module:      "dashboard",
		Title:       "Dashboard",
		Description: "Main dashboard with KPI cards, charts, and recent activity",
		Fn:          Schema,
	})
}

func Schema(_ ui.UISessionContext) ui.Schema {
	return amis.Page("").
		Body([]any{
			// ── Header row ──
			amis.M{
				"type":       "flex",
				"justify":    "space-between",
				"alignItems": "center",
				"className":  "mb-4",
				"items": amis.A{
					amis.Breadcrumb(amis.BC("Home", "#dashboard")),
					amis.M{
						"type":      "tpl",
						"tpl":       "<span class='text-muted'>Last updated: April 13, 2026</span>",
						"className": "text-sm",
					},
				},
			},
			// ── KPI cards ──
			amis.Grid(
				statCard("Total Users", "24", "fa-users", "bg-primary"),
				statCard("Active Sessions", "8", "fa-circle-dot", "bg-success"),
				statCard("Pending Tasks", "13", "fa-clock", "bg-warning"),
				statCard("Errors (24h)", "2", "fa-triangle-exclamation", "bg-danger"),
			).Build(),
			amis.HDivider(),
			// ── Charts ──
			amis.Grid(
				amis.Col(8, amis.Panel("User Signups — Weekly").Body(signupsChart()).Build()),
				amis.Col(4, amis.Panel("Users by Role").Body(rolesChart()).Build()),
			).Build(),
			amis.HDivider(),
			// ── Recent activity ──
			amis.Panel("Recent Activity").Body(activityTable()).Build(),
		}).Build()
}

func statCard(title, value, icon, colorClass string) amis.M {
	return amis.M{
		"body": amis.M{
			"type":      "tpl",
			"className": colorClass + " text-white rounded p-4",
			"tpl": "<div class='d-flex align-items-center gap-3'>" +
				"<i class='fa " + icon + " fa-2x opacity-75'></i>" +
				"<div><div style='font-size:2rem;font-weight:700;line-height:1'>" + value + "</div>" +
				"<div class='opacity-75 mt-1'>" + title + "</div></div></div>",
		},
	}
}

func signupsChart() amis.M {
	return amis.Chart("").Config(amis.M{
		"xAxis":   amis.M{"type": "category", "data": amis.A{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}},
		"yAxis":   amis.M{"type": "value"},
		"tooltip": amis.M{"trigger": "axis"},
		"series": amis.A{amis.M{
			"data":      amis.A{4, 7, 3, 9, 6, 12, 8},
			"type":      "line",
			"smooth":    true,
			"areaStyle": amis.M{},
		}},
	}).Height(220).Build()
}

func rolesChart() amis.M {
	return amis.Chart("").Config(amis.M{
		"tooltip": amis.M{"trigger": "item"},
		"legend":  amis.M{"bottom": 0},
		"series": amis.A{amis.M{
			"type":   "pie",
			"radius": "60%",
			"data": amis.A{
				amis.M{"name": "Admin", "value": 3},
				amis.M{"name": "Manager", "value": 7},
				amis.M{"name": "Staff", "value": 14},
			},
		}},
	}).Height(220).Build()
}

func activityTable() amis.M {
	return amis.CRUD("").
		StaticData(amis.A{
			amis.M{"user": "alice.wanjiku", "action": "Created invoice INV-2026-043", "timestamp": "2026-04-13 14:32"},
			amis.M{"user": "brian.otieno", "action": "Approved payment PMT-2026-011", "timestamp": "2026-04-13 13:15"},
			amis.M{"user": "carol.maina", "action": "Added account 5100 — Office Rent", "timestamp": "2026-04-13 11:48"},
			amis.M{"user": "alice.wanjiku", "action": "Voided payment PMT-2026-012", "timestamp": "2026-04-13 10:02"},
			amis.M{"user": "david.kamau", "action": "Updated tenant Baobab Supplies Co", "timestamp": "2026-04-12 17:55"},
		}).
		Columns(
			amis.Column("user", "User").Width(160),
			amis.Column("action", "Action"),
			amis.Column("timestamp", "Time").Width(160).Align("right"),
		).
		Toolbar().
		PerPage(5).
		Build()
}
