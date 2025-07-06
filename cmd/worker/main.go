package main

import (
    "log"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    
    "github.com/niiniyare/erp/internal/workflows/tenant"
    "github.com/niiniyare/erp/internal/platform/config"
)

func main() {
    cfg := config.Load()
    
    // Create Temporal client
    c, err := client.Dial(client.Options{
        HostPort: cfg.Temporal.HostPort,
    })
    if err != nil {
        log.Fatalln("Unable to create client", err)
    }
    defer c.Close()
    
    // Create worker
    w := worker.New(c, "erp-task-queue", worker.Options{})
    
    // Register workflows and activities
    w.RegisterWorkflow(tenant.TenantOnboardingWorkflow)
    w.RegisterActivity(tenant.CreateTenantActivity)
    w.RegisterActivity(tenant.SetupDefaultOrganizationActivity)
    w.RegisterActivity(tenant.SendWelcomeEmailActivity)
    
    // Start worker
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Unable to start worker", err)
    }
}
