package settings

import (
	"awo.so/internal/web/amis"
	"awo.so/internal/web/registry"
)

func init() {
	registry.Register("/settings", Schema)
}

func Schema(ctx amis.Ctx) amis.Schema {
	return amis.Page("Settings").
		Body(
			amis.Tabs().
				Tab("Profile", profileTab()).
				Tab("Security", securityTab()).
				Tab("Notifications", notificationsTab()).
				Build(),
		).Build()
}

func profileTab() amis.Schema {
	return amis.Form("put:/api/v1/users/${user_id}").
		InitAPI("get:/api/v1/users/${user_id}").
		Title("").
		Fields(
			amis.Section("Personal Information"),
			amis.Required(amis.TextField("username", "Username")),
			amis.Required(amis.TextField("email", "Email")),
			amis.Optional(amis.TextField("full_name", "Full Name")),
			amis.Optional(amis.TextField("phone", "Phone")),
			amis.Divider(),
			amis.Section("Preferences"),
			amis.SelectField("timezone", "Timezone",
				amis.SelectOpt("Africa/Nairobi (EAT)", "Africa/Nairobi"),
				amis.SelectOpt("UTC", "UTC"),
				amis.SelectOpt("Europe/London (GMT)", "Europe/London"),
				amis.SelectOpt("America/New_York (ET)", "America/New_York"),
			),
			amis.SelectField("language", "Language",
				amis.SelectOpt("English", "en"),
				amis.SelectOpt("Swahili", "sw"),
			),
		).Build()
}

func securityTab() amis.Schema {
	return amis.Form("post:/api/v1/users/${user_id}/change-password").
		Title("").
		Fields(
			amis.Section("Change Password"),
			amis.Required(amis.TextField("current_password", "Current Password")),
			amis.Required(amis.TextField("new_password", "New Password")),
			amis.Required(amis.TextField("confirm_password", "Confirm Password")),
		).Build()
}

func notificationsTab() amis.Schema {
	return amis.Form("put:/api/v1/users/${user_id}/notifications").
		InitAPI("get:/api/v1/users/${user_id}/notifications").
		Title("").
		Fields(
			amis.Section("Email Notifications"),
			amis.SwitchField("notify_login", "Login alerts"),
			amis.SwitchField("notify_transaction", "New transactions"),
			amis.SwitchField("notify_report", "Weekly reports"),
			amis.Divider(),
			amis.Section("In-App Notifications"),
			amis.SwitchField("notify_mentions", "Mentions"),
			amis.SwitchField("notify_tasks", "Task assignments"),
		).Build()
}
