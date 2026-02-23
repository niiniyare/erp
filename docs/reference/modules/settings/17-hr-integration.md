# Chapter 17: HR Module Integration

[← Finance Integration](./16-finance-integration.md) | [Next: Inventory Integration →](./18-inventory-integration.md)

---

## HR Configuration Keys

HR relies on configuration for payroll calculations, overtime rules, and benefits eligibility. Entity-level overrides allow different rules for manufacturing floors vs. office workers within the same organization.

---

## Loading Payroll Configuration

```go
type HRSettingsIntegration struct {
    settingsService service.SettingsService
    payrollService  PayrollService
    eventSubscriber EventSubscriber
}

func (h *HRSettingsIntegration) GetPayrollConfiguration(
    ctx context.Context, tenantID tenant.ID, entityID *entity.ID,
) (*PayrollConfiguration, error) {

    configKeys := []service.ConfigurationKey{
        {Module: "hr", Key: "default_pay_frequency"},
        {Module: "hr", Key: "overtime_threshold"},
        {Module: "hr", Key: "overtime_multiplier"},
        {Module: "hr", Key: "pay_period_start_day"},
        {Module: "hr", Key: "tax_jurisdiction"},
        {Module: "hr", Key: "benefits_waiting_period"},
        {Module: "hr", Key: "vacation_accrual_method"},
        {Module: "hr", Key: "sick_leave_annual_hours"},
    }

    resolved, err := h.settingsService.ResolveConfigurations(ctx, service.ResolutionRequest{
        TenantID:       tenantID,
        EntityID:       entityID,
        Configurations: configKeys,
    })
    if err != nil {
        return nil, err
    }

    config := &PayrollConfiguration{TenantID: tenantID, EntityID: entityID}

    for _, cfg := range resolved.Configurations {
        switch cfg.ConfigKey {
        case "default_pay_frequency":
            config.PayFrequency = PayFrequency(cfg.Value.(string))
        case "overtime_threshold":
            config.OvertimeThreshold = int(cfg.Value.(float64))
        case "overtime_multiplier":
            config.OvertimeMultiplier = decimal.NewFromFloat(cfg.Value.(float64))
        case "pay_period_start_day":
            config.PayPeriodStartDay = cfg.Value.(string)
        case "tax_jurisdiction":
            config.TaxJurisdiction = cfg.Value.(string)
        case "benefits_waiting_period":
            config.BenefitsWaitingPeriod = int(cfg.Value.(float64))
        case "vacation_accrual_method":
            config.VacationAccrualMethod = cfg.Value.(string)
        case "sick_leave_annual_hours":
            config.SickLeaveAnnualHours = int(cfg.Value.(float64))
        }
    }

    return config, nil
}
```

---

## Overtime Calculation

```go
func (s *HRService) CalculateOvertimePay(
    ctx context.Context, employee *Employee, hoursWorked float64,
) (*OvertimePay, error) {

    configs, err := s.configService.ListEffectiveConfigurations(
        ctx, &employee.EntityID, "hr",
    )
    if err != nil {
        return nil, fmt.Errorf("failed to get HR configuration: %w", err)
    }

    overtimeThreshold := 40.0
    overtimeMultiplier := 1.5

    for _, config := range configs {
        switch config.Key.String() {
        case "overtime_threshold":
            if val, err := config.Value.AsDecimal(); err == nil {
                overtimeThreshold, _ = val.Float64()
            }
        case "overtime_multiplier":
            if val, err := config.Value.AsDecimal(); err == nil {
                overtimeMultiplier, _ = val.Float64()
            }
        }
    }

    if hoursWorked <= overtimeThreshold {
        return &OvertimePay{RegularHours: hoursWorked, OvertimeHours: 0}, nil
    }

    return &OvertimePay{
        RegularHours:       overtimeThreshold,
        OvertimeHours:      hoursWorked - overtimeThreshold,
        OvertimeMultiplier: overtimeMultiplier,
    }, nil
}
```

---

## Reacting to Configuration Changes

HR services subscribe to configuration change events to trigger payroll recalculations:

```go
func (h *HRSettingsIntegration) StartConfigurationListener(ctx context.Context) error {
    return h.eventSubscriber.Subscribe(ctx, "settings.configuration.changed",
        func(event ConfigurationChangedEvent) error {
            if event.Module != "hr" {
                return nil
            }

            payrollAffecting := []string{
                "default_pay_frequency",
                "overtime_threshold",
                "overtime_multiplier",
                "tax_jurisdiction",
            }

            for _, key := range payrollAffecting {
                if event.ConfigKey == key {
                    return h.handlePayrollConfigurationChange(ctx, event)
                }
            }
            return nil
        })
}

func (h *HRSettingsIntegration) handlePayrollConfigurationChange(
    ctx context.Context, event ConfigurationChangedEvent,
) error {
    employees, err := h.payrollService.GetActiveEmployees(
        ctx, event.TenantID, event.EntityID,
    )
    if err != nil {
        return err
    }

    for _, employee := range employees {
        err := h.payrollService.ScheduleRecalculation(ctx, PayrollRecalculationRequest{
            EmployeeID:    employee.ID,
            TenantID:      event.TenantID,
            EntityID:      event.EntityID,
            ConfigChange:  event.ConfigKey,
            EffectiveDate: time.Now(),
            Reason:        "Configuration change",
        })
        if err != nil {
            log.Printf("Failed to schedule recalculation for %s: %v", employee.ID, err)
            // Continue processing other employees
        }
    }
    return nil
}
```

---

## Common HR Configuration Scenarios

| Scenario | Config Key | Level |
|---|---|---|
| Manufacturing floor works 45h weeks | `hr.overtime_threshold = 45` | Entity (factory) |
| Switch to weekly payroll | `hr.default_pay_frequency = "weekly"` | Tenant |
| New state tax rules | `hr.tax_jurisdiction = "US-CA"` | Entity (CA location) |
| Extend benefits waiting period | `hr.benefits_waiting_period = 90` | Tenant |
| Increase sick leave for hourly workers | `hr.sick_leave_annual_hours = 56` | Entity |

---

[← Finance Integration](./16-finance-integration.md) | [Next: Inventory Integration →](./18-inventory-integration.md)
