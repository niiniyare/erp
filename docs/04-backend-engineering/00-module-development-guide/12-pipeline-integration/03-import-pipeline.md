---
title: Import Pipeline
portal: 4 — Backend Engineering
section: 00-module-development-guide/12-pipeline-integration
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-pipeline-design.md
    title: Pipeline Design
---

# Import Pipeline

## Full Import Pipeline Implementation

```go
// internal/core/contracts/pipeline/import_pipeline.go
package pipeline

import (
    "bytes"
    "context"
    "encoding/csv"
    "fmt"
    "strconv"
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"

    "awo.so/internal/core/contracts/domain"
    "awo.so/internal/core/contracts/repository"
)

type ImportPipeline struct {
    repo        repository.ContractRepository
    vendorRepo  VendorRepository
    tenantID    uuid.UUID
    entityID    uuid.UUID
    createdBy   uuid.UUID
}

func NewImportPipeline(
    repo repository.ContractRepository,
    vendorRepo VendorRepository,
    tenantID, entityID, createdBy uuid.UUID,
) *ImportPipeline {
    return &ImportPipeline{
        repo:       repo,
        vendorRepo: vendorRepo,
        tenantID:   tenantID,
        entityID:   entityID,
        createdBy:  createdBy,
    }
}

func (p *ImportPipeline) Run(ctx context.Context, csvData []byte) (*ImportResult, error) {
    result := &ImportResult{}

    // Stage 1: Parse CSV
    rawRows, err := p.parseCSV(csvData)
    if err != nil {
        return nil, fmt.Errorf("parse CSV: %w", err)
    }
    result.TotalRows = len(rawRows)

    // Stage 2: Validate
    validRows, rowErrs := p.validateRows(rawRows)
    result.Errors = append(result.Errors, rowErrs...)

    // Stage 3: Enrich
    enrichedRows, enrichErrs := p.enrichRows(ctx, validRows)
    result.Errors = append(result.Errors, enrichErrs...)

    // Stage 4: Persist (batch)
    createdIDs, persistErrs := p.persistRows(ctx, enrichedRows)
    result.Errors = append(result.Errors, persistErrs...)
    result.ContractIDs = createdIDs

    result.Succeeded = len(createdIDs)
    result.Failed = len(result.Errors)

    return result, nil
}

// parseCSV reads the CSV header and rows.
func (p *ImportPipeline) parseCSV(data []byte) ([]RawRow, error) {
    r := csv.NewReader(bytes.NewReader(data))
    r.TrimLeadingSpace = true

    records, err := r.ReadAll()
    if err != nil {
        return nil, err
    }
    if len(records) < 2 {
        return nil, fmt.Errorf("CSV must have at least a header row and one data row")
    }

    headers := records[0]
    colIdx := buildColumnIndex(headers)

    var rows []RawRow
    for i, record := range records[1:] {
        rows = append(rows, RawRow{
            RowNumber:      i + 2, // 1-based, header is row 1
            ContractNumber: getCol(record, colIdx, "contract_number"),
            Title:          getCol(record, colIdx, "title"),
            VendorName:     getCol(record, colIdx, "vendor_name"),
            ContractType:   getCol(record, colIdx, "contract_type"),
            StartDate:      getCol(record, colIdx, "start_date"),
            EndDate:        getCol(record, colIdx, "end_date"),
            TotalValue:     getCol(record, colIdx, "total_value"),
            Currency:       getCol(record, colIdx, "currency"),
        })
    }
    return rows, nil
}

// validateRows checks each row's data types and required fields.
func (p *ImportPipeline) validateRows(rows []RawRow) ([]ValidatedRow, []RowError) {
    var valid []ValidatedRow
    var errs  []RowError

    for _, row := range rows {
        var rowErrs []RowError

        if _, err := domain.NewContractNumber(row.ContractNumber); err != nil {
            rowErrs = append(rowErrs, RowError{row.RowNumber, "contract_number", err.Error()})
        }
        if row.Title == "" {
            rowErrs = append(rowErrs, RowError{row.RowNumber, "title", "required"})
        }

        startDate, err := time.Parse("2006-01-02", row.StartDate)
        if err != nil {
            rowErrs = append(rowErrs, RowError{row.RowNumber, "start_date", "must be YYYY-MM-DD"})
        }
        endDate, err := time.Parse("2006-01-02", row.EndDate)
        if err != nil {
            rowErrs = append(rowErrs, RowError{row.RowNumber, "end_date", "must be YYYY-MM-DD"})
        }

        totalValue, err := decimal.NewFromString(row.TotalValue)
        if err != nil {
            rowErrs = append(rowErrs, RowError{row.RowNumber, "total_value", "must be a number"})
        }

        if len(rowErrs) > 0 {
            errs = append(errs, rowErrs...)
            continue
        }

        _ = startDate; _ = endDate; _ = totalValue
        valid = append(valid, ValidatedRow{RawRow: row})
    }
    return valid, errs
}

// persistRows batch-inserts contracts, collecting per-row errors.
func (p *ImportPipeline) persistRows(ctx context.Context, rows []EnrichedRow) ([]uuid.UUID, []RowError) {
    var ids  []uuid.UUID
    var errs []RowError

    for _, row := range rows {
        startDate, _ := time.Parse("2006-01-02", row.StartDate)
        endDate, _   := time.Parse("2006-01-02", row.EndDate)
        totalValue, _ := decimal.NewFromString(row.TotalValue)

        contract, err := p.repo.Create(ctx, repository.CreateContractParams{
            TenantID:       p.tenantID,
            EntityID:       p.entityID,
            ContractNumber: row.ContractNumber,
            Title:          row.Title,
            VendorID:       row.VendorID,
            ContractType:   domain.ContractType(row.ContractType),
            StartDate:      startDate,
            EndDate:        endDate,
            TotalValue:     totalValue,
            Currency:       row.Currency,
            CreatedBy:      p.createdBy,
        })
        if err != nil {
            if errors.Is(err, domain.ErrContractAlreadyExists) {
                errs = append(errs, RowError{row.RowNumber, "contract_number", "already exists"})
            } else {
                errs = append(errs, RowError{row.RowNumber, "", "internal error: " + err.Error()})
            }
            continue
        }
        ids = append(ids, contract.ID)
    }
    return ids, errs
}
```

## HTTP Handler: Upload Endpoint

```go
// POST /api/v1/contracts/import
func (h *contractHandler) Import(c *fiber.Ctx) error {
    session := middleware.SessionFrom(c)

    file, err := c.FormFile("file")
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "file field required")
    }
    if file.Size > 10*1024*1024 { // 10 MB limit
        return fiber.NewError(fiber.StatusRequestEntityTooLarge, "file exceeds 10 MB limit")
    }

    f, err := file.Open()
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "failed to read file")
    }
    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "failed to read file")
    }

    result, err := h.svc.Import(c.Context(), service.ImportContractsRequest{
        TenantID:  session.TenantID,
        EntityID:  session.EntityID(),
        UserID:    session.UserID,
        Principal: session.ToPrincipal(),
        CSVData:   data,
    })
    if err != nil {
        return h.mapError(err)
    }

    return c.Status(fiber.StatusOK).JSON(result)
}
```
