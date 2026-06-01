---
title: Pipeline Patterns Quick Reference
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Pipeline Overview](01-pipeline-overview.md)"
  - "[Pipeline Design](02-pipeline-design.md)"
  - "[Import Pipeline](03-import-pipeline.md)"
  - "[Export Pipeline](04-export-pipeline.md)"
---

# Pipeline Patterns Quick Reference

## Import Pipeline Structure

```
HTTP Upload (multipart/form-data)
  └── Parse stage      CSV rows → RawRow
        └── Validate stage   RawRow → ValidatedRow (accumulate errors)
              └── Enrich stage    ValidatedRow → EnrichedRow (DB lookups)
                    └── Persist stage   EnrichedRow → saved record
                          └── Result    ImportResult{succeeded, failed, errors}
```

## Stage Function Signature

```go
type Stage[In, Out any] func(ctx context.Context, in []In) ([]Out, []RowError)
```

Each stage returns successful items and per-row errors. Errors accumulate — don't abort on first failure.

## Import Handler

```go
func (h *ContractHandler) Import(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)

    file, err := c.FormFile("file")
    if err != nil {
        return fiber.NewError(400, "file is required")
    }
    if file.Size > 10*1024*1024 {
        return fiber.NewError(413, "file must be under 10 MB")
    }

    f, err := file.Open()
    if err != nil {
        return fiber.NewError(500, "could not open file")
    }
    defer f.Close()

    result, err := h.svc.Import(c.UserContext(), sess, f)
    if err != nil {
        return mapError(err)
    }

    return c.JSON(result)
}
```

## Import Result Shape

```go
type ImportResult struct {
    TotalRows   int         `json:"total_rows"`
    Succeeded   int         `json:"succeeded"`
    Failed      int         `json:"failed"`
    Errors      []RowError  `json:"errors,omitempty"`
    ContractIDs []uuid.UUID `json:"contract_ids"`
}

type RowError struct {
    RowNumber int    `json:"row_number"`
    Field     string `json:"field"`
    Message   string `json:"message"`
}
```

## CSV Parsing

```go
func parseCSV(r io.Reader) ([]RawRow, error) {
    reader := csv.NewReader(r)
    reader.TrimLeadingSpace = true

    headers, err := reader.Read()
    if err != nil {
        return nil, fmt.Errorf("read headers: %w", err)
    }
    idx := headerIndex(headers)

    var rows []RawRow
    lineNum := 2   // 1-indexed, header is row 1
    for {
        record, err := reader.Read()
        if errors.Is(err, io.EOF) {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("row %d: %w", lineNum, err)
        }
        rows = append(rows, RawRow{
            LineNumber:     lineNum,
            ContractNumber: record[idx["contract_number"]],
            Title:          record[idx["title"]],
            TotalValue:     record[idx["total_value"]],
        })
        lineNum++
    }
    return rows, nil
}
```

## Validate Stage

```go
func validateRows(rows []RawRow) ([]ValidatedRow, []RowError) {
    var valid []ValidatedRow
    var errs []RowError

    for _, row := range rows {
        var rowErrs []RowError

        if row.ContractNumber == "" {
            rowErrs = append(rowErrs, RowError{row.LineNumber, "contract_number", "required"})
        }
        if row.Title == "" {
            rowErrs = append(rowErrs, RowError{row.LineNumber, "title", "required"})
        }

        val, err := decimal.NewFromString(row.TotalValue)
        if err != nil {
            rowErrs = append(rowErrs, RowError{row.LineNumber, "total_value", "must be a number"})
        }

        if len(rowErrs) > 0 {
            errs = append(errs, rowErrs...)
            continue   // skip this row — don't enrich invalid rows
        }

        valid = append(valid, ValidatedRow{
            LineNumber:     row.LineNumber,
            ContractNumber: row.ContractNumber,
            Title:          row.Title,
            TotalValue:     val,
        })
    }
    return valid, errs
}
```

## Export Handler

```go
func (h *ContractHandler) Export(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)
    status := c.Query("status")

    data, err := h.svc.Export(c.UserContext(), sess, service.ExportParams{
        Status: parseOptionalStatus(status),
    })
    if err != nil {
        return mapError(err)
    }

    filename := fmt.Sprintf("contracts-%s.csv", time.Now().Format("20060102"))
    c.Set("Content-Type", "text/csv")
    c.Set("Content-Disposition", "attachment; filename="+filename)
    return c.Send(data)
}
```

## Large Export via Temporal

If record count may exceed memory limits:

```go
func (h *ContractHandler) ExportLarge(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)

    // Start Temporal workflow — returns immediately
    jobID, err := h.svc.StartExportJob(c.UserContext(), sess, service.ExportParams{...})
    if err != nil {
        return mapError(err)
    }

    // Return job ID — client polls for completion
    return c.Status(202).JSON(fiber.Map{
        "job_id":     jobID,
        "status_url": "/api/v1/contracts/export-jobs/" + jobID,
    })
}
```

## Concurrency Limit

Use semaphore to bound concurrent DB writes in persist stage:

```go
sem := make(chan struct{}, 10)   // max 10 concurrent
var wg sync.WaitGroup
var mu sync.Mutex
var succeeded int
var errs []RowError

for _, row := range enriched {
    sem <- struct{}{}
    wg.Add(1)
    go func(r EnrichedRow) {
        defer wg.Done()
        defer func() { <-sem }()

        _, err := repo.Create(ctx, tenantID, toCreateParams(r))
        mu.Lock()
        if err != nil {
            errs = append(errs, RowError{r.LineNumber, "persist", err.Error()})
        } else {
            succeeded++
        }
        mu.Unlock()
    }(row)
}
wg.Wait()
```
