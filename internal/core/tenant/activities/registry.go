package activities

import "go.temporal.io/sdk/worker"

// Register registers all tenant activities with a Temporal worker.
func (a *Activities) Register(w worker.Worker) {
	w.RegisterActivity(a.ProvisionTenantActivity)
	w.RegisterActivity(a.CreateDefaultConfigActivity)
	w.RegisterActivity(a.InitUsageActivity)
	w.RegisterActivity(a.ActivateTenantActivity)
	w.RegisterActivity(a.CleanupTenantActivity)
	w.RegisterActivity(a.SendWelcomeNotificationActivity)
	w.RegisterActivity(a.BulkUpdateStatusActivity)
	w.RegisterActivity(a.BulkSoftDeleteActivity)
}
