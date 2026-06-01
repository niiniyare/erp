---
title: Pipeline Design
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Pipeline Overview](01-pipeline-overview.md)"
  - "[Import Pipeline](03-import-pipeline.md)"
  - "[Export Pipeline](04-export-pipeline.md)"
---

# Pipeline Design

## Pipeline Result Types

```go
// internal/core/contracts/pipeline/types.go
package pipeline

import "github.com/google/uuid"

// RawRow is one row as parsed from CSV — all strings.
type RawRow struct {
    RowNumber       int
    ContractNumber  string
    Title           string
    VendorName      string
    ContractType    string
    StartDate       string
    EndDate         string
    TotalValue      string
    Currency        string
}

// ValidatedRow has passed schema validation.
type ValidatedRow struct {
    RawRow
    // parsed types added by validation
}

// EnrichedRow has vendor ID resolved from VendorName.
type EnrichedRow struct {
    ValidatedRow
    VendorID uuid.UUID
    EntityID uuid.UUID
}

// RowError records a per-row failure.
type RowError struct {
    RowNumber int
    Field     string
    Message   string
}

// ImportResult is the pipeline's final output.
type ImportResult struct {
    TotalRows  int
    Succeeded  int
    Failed     int
    Errors     []RowError
    ContractIDs []uuid.UUID
}
```

## Error Accumulation Pattern

Pipelines accumulate per-row errors rather than failing on first error. This gives users a complete error report in one pass:

```go
type validateRowStage struct{}

func (s *validateRowStage) ProcessBatch(rows []RawRow) ([]ValidatedRow, []RowError) {
    var valid []ValidatedRow
    var errs  []RowError

    for _, row := range rows {
        if row.ContractNumber == "" {
            errs = append(errs, RowError{
                RowNumber: row.RowNumber,
                Field:     "contract_number",
                Message:   "contract_number is required",
            })
            continue
        }
        if _, err := domain.NewContractNumber(row.ContractNumber); err != nil {
            errs = append(errs, RowError{
                RowNumber: row.RowNumber,
                Field:     "contract_number",
                Message:   err.Error(),
            })
            continue
        }
        // ... validate other fields ...

        valid = append(valid, ValidatedRow{RawRow: row})
    }
    return valid, errs
}
```

## Batch Size

Process rows in batches to limit memory usage for large imports:

```go
const defaultBatchSize = 100

func processBatches[T any](rows []T, batchSize int, fn func([]T) error) error {
    for i := 0; i < len(rows); i += batchSize {
        end := i + batchSize
        if end > len(rows) {
            end = len(rows)
        }
        if err := fn(rows[i:end]); err != nil {
            return err
        }
    }
    return nil
}
```

## Concurrency

For large files, use a worker pool for CPU-bound stages (validation, enrichment). Keep persistence stages sequential to avoid connection pool exhaustion:

```go
func (p *importPipeline) enrichBatch(ctx context.Context, rows []ValidatedRow) ([]EnrichedRow, []RowError) {
    type result struct {
        row EnrichedRow
        err *RowError
    }

    results := make([]result, len(rows))
    var wg sync.WaitGroup
    sem := make(chan struct{}, 10) // max 10 concurrent enrichments

    for i, row := range rows {
        wg.Add(1)
        go func(i int, row ValidatedRow) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()

            vendorID, err := p.vendorRepo.FindByName(ctx, row.VendorName)
            if err != nil {
                results[i] = result{err: &RowError{
                    RowNumber: row.RowNumber,
                    Field:     "vendor_name",
                    Message:   "vendor not found: " + row.VendorName,
                }}
                return
            }
            results[i] = result{row: EnrichedRow{ValidatedRow: row, VendorID: vendorID}}
        }(i, row)
    }
    wg.Wait()

    var enriched []EnrichedRow
    var errs    []RowError
    for _, r := range results {
        if r.err != nil {
            errs = append(errs, *r.err)
        } else {
            enriched = append(enriched, r.row)
        }
    }
    return enriched, errs
}
```
