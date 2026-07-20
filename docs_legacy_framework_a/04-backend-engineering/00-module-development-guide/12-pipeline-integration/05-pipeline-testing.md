> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Pipeline Testing
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Pipeline Overview](01-pipeline-overview.md)"
  - "[Import Pipeline](03-import-pipeline.md)"
  - "[Export Pipeline](04-export-pipeline.md)"
  - "[Testing Overview](../18-testing/01-testing-overview.md)"
---

# Pipeline Testing

## Import Pipeline Integration Test

```go
// internal/core/contracts/pipeline/import_pipeline_test.go
package pipeline_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "awo.so/internal/core/contracts/pipeline"
    "awo.so/internal/core/contracts/repository"
    "awo.so/internal/testutil"
)

func TestImportPipeline_HappyPath(t *testing.T) {
    store    := testutil.NewTestStore(t)
    repo     := repository.NewContractRepository(store)
    tenantID := testutil.CreateTestTenant(t, store)
    entityID := testutil.CreateTestEntity(t, store, tenantID)

    p := pipeline.NewImportPipeline(repo, &stubVendorRepo{}, tenantID, entityID, testutil.TestUserID)

    csvData := []byte(`contract_number,title,vendor_name,contract_type,start_date,end_date,total_value,currency
CONT-2025-0001,Test Contract A,Acme Corp,service,2025-01-01,2025-12-31,50000,USD
CONT-2025-0002,Test Contract B,Beta Inc,goods,2025-01-01,2025-12-31,25000,USD
`)

    result, err := p.Run(context.Background(), csvData)
    require.NoError(t, err)

    assert.Equal(t, 2, result.TotalRows)
    assert.Equal(t, 2, result.Succeeded)
    assert.Equal(t, 0, result.Failed)
    assert.Empty(t, result.Errors)
    assert.Len(t, result.ContractIDs, 2)
}

func TestImportPipeline_ValidationErrors_AccumulatesAllRowErrors(t *testing.T) {
    // ...
    csvData := []byte(`contract_number,title,vendor_name,contract_type,start_date,end_date,total_value,currency
,Missing Number,Acme Corp,service,2025-01-01,2025-12-31,50000,USD
CONTRACT-BAD-FMT,Has Title,Beta Inc,goods,not-a-date,2025-12-31,25000,USD
CONT-2025-0001,Also Valid,Acme Corp,service,2025-01-01,2025-12-31,100,USD
`)

    result, err := p.Run(context.Background(), csvData)
    require.NoError(t, err) // pipeline errors are in result, not returned as error

    assert.Equal(t, 3, result.TotalRows)
    assert.Equal(t, 1, result.Succeeded)  // row 3 succeeds
    assert.Equal(t, 2, result.Failed)

    // Verify errors reference correct rows
    rowNumbers := make(map[int]bool)
    for _, e := range result.Errors {
        rowNumbers[e.RowNumber] = true
    }
    assert.True(t, rowNumbers[2], "row 2 should have error")
    assert.True(t, rowNumbers[3], "row 3 should have error")
}

func TestImportPipeline_DuplicateContractNumber_IsRowError(t *testing.T) {
    // Pre-create one contract
    // ...

    csvData := []byte(/* rows with one duplicate number */)
    result, err := p.Run(context.Background(), csvData)
    require.NoError(t, err)

    assert.Equal(t, 1, result.Failed)
    assert.Equal(t, "contract_number", result.Errors[0].Field)
    assert.Contains(t, result.Errors[0].Message, "already exists")
}

// Stub for vendor resolution in tests
type stubVendorRepo struct{}
func (s *stubVendorRepo) FindByName(_ context.Context, name string) (uuid.UUID, error) {
    // Return a predictable UUID for testing
    return uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)), nil
}
```

## Export Pipeline Test

```go
func TestExportPipeline_ProducesValidCSV(t *testing.T) {
    store    := testutil.NewTestStore(t)
    repo     := repository.NewContractRepository(store)
    tenantID := testutil.CreateTestTenant(t, store)
    entityID := testutil.CreateTestEntity(t, store, tenantID)

    // Create two test contracts
    testutil.CreateTestContract(t, store, tenantID, entityID)
    testutil.CreateTestContract(t, store, tenantID, entityID)

    p := pipeline.NewExportPipeline(repo, tenantID)
    data, err := p.ExportCSV(context.Background(), pipeline.ExportFilter{})
    require.NoError(t, err)
    require.NotEmpty(t, data)

    // Parse the exported CSV and verify
    r := csv.NewReader(bytes.NewReader(data))
    records, err := r.ReadAll()
    require.NoError(t, err)

    assert.Equal(t, "contract_number", records[0][0], "first column header")
    assert.Len(t, records, 3, "header + 2 data rows")
}
```
