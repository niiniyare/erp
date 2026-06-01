---
title: Export Pipeline
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Pipeline Overview](01-pipeline-overview.md)"
  - "[Pipeline Design](02-pipeline-design.md)"
---

# Export Pipeline

## Contract Export to CSV

```go
// internal/core/contracts/pipeline/export_pipeline.go
package pipeline

import (
    "bytes"
    "context"
    "encoding/csv"
    "time"

    "awo.so/internal/core/contracts/domain"
    "awo.so/internal/core/contracts/repository"
    "github.com/google/uuid"
)

type ExportPipeline struct {
    repo     repository.ContractRepository
    tenantID uuid.UUID
}

type ExportFilter struct {
    Status   *domain.ContractStatus
    EntityID *uuid.UUID
    FromDate *time.Time
    ToDate   *time.Time
}

func (p *ExportPipeline) ExportCSV(ctx context.Context, filter ExportFilter) ([]byte, error) {
    // Step 1: Query
    contracts, err := p.repo.List(ctx, repository.ListContractsParams{
        TenantID: p.tenantID,
        Status:   filter.Status,
        EntityID: filter.EntityID,
    })
    if err != nil {
        return nil, err
    }

    // Step 2: Format
    var buf bytes.Buffer
    w := csv.NewWriter(&buf)

    // Header
    _ = w.Write([]string{
        "contract_number", "title", "status", "contract_type",
        "total_value", "currency", "start_date", "end_date",
        "created_at",
    })

    for _, c := range contracts {
        _ = w.Write([]string{
            c.ContractNumber,
            c.Title,
            string(c.Status),
            string(c.ContractType),
            c.TotalValue.String(),
            c.Currency,
            c.StartDate.Format("2006-01-02"),
            c.EndDate.Format("2006-01-02"),
            c.CreatedAt.Format(time.RFC3339),
        })
    }

    w.Flush()
    if err := w.Error(); err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}
```

## HTTP Handler: Download Endpoint

```go
// GET /api/v1/contracts/export?status=active&format=csv
func (h *contractHandler) Export(c *fiber.Ctx) error {
    session := middleware.SessionFrom(c)

    format := c.Query("format", "csv")
    if format != "csv" {
        return fiber.NewError(fiber.StatusBadRequest, "only format=csv supported")
    }

    statusStr := c.Query("status")
    var statusFilter *domain.ContractStatus
    if statusStr != "" {
        s := domain.ContractStatus(statusStr)
        statusFilter = &s
    }

    data, err := h.svc.ExportCSV(c.Context(), service.ExportContractsRequest{
        TenantID:  session.TenantID,
        Principal: session.ToPrincipal(),
        Status:    statusFilter,
    })
    if err != nil {
        return h.mapError(err)
    }

    filename := "contracts-" + time.Now().Format("20060102") + ".csv"
    c.Set("Content-Disposition", "attachment; filename="+filename)
    c.Set("Content-Type", "text/csv; charset=utf-8")
    return c.Status(fiber.StatusOK).Send(data)
}
```

## Large Export: Streaming with Temporal

For exports > 50K rows, run as a Temporal workflow to avoid request timeouts:

```go
func ContractLargeExportWorkflow(ctx workflow.Context, input LargeExportInput) (LargeExportResult, error) {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    var result LargeExportResult
    err := workflow.ExecuteActivity(ctx, activities.GenerateLargeExport, input).Get(ctx, &result)
    return result, err
}
```

The handler starts the workflow and returns a `202 Accepted` with a job ID. The client polls for completion.
