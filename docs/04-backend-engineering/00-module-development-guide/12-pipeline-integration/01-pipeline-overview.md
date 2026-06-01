---
title: Pipeline Integration Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/12-pipeline-integration
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-pipeline-design.md
    title: Pipeline Design
  - path: ../13-event-driven/01-event-driven-overview.md
    title: Event-Driven Overview
---

# Pipeline Integration Overview

Pipelines are ordered, multi-stage data processing chains. Modules use them for:

- **Import pipelines**: validate → transform → persist batch data (CSV uploads, API ingestion)
- **Export pipelines**: query → format → deliver (report generation, data sync)
- **Processing pipelines**: enrich → validate → route (document processing, approval routing)

## Architecture

```
Input Source
    │
    ▼
Stage 1: Parse / Deserialize
    │
    ▼
Stage 2: Validate (per-row)
    │
    ▼
Stage 3: Enrich (lookup foreign keys, resolve defaults)
    │
    ▼
Stage 4: Persist (batch insert)
    │
    ▼
Stage 5: Post-process (notifications, audit, events)
    │
    ▼
Result: ImportResult{Succeeded, Failed []RowError}
```

## Pipeline Interface

```go
// internal/platform/pipeline/pipeline.go
package pipeline

import "context"

// Stage is a single processing step.
// It receives items and emits processed items (or errors).
type Stage[In, Out any] interface {
    Process(ctx context.Context, in In) (Out, error)
}

// Pipeline chains stages together.
type Pipeline[In, Out any] interface {
    Run(ctx context.Context, input In) (Out, error)
}
```

## When Pipeline vs. Direct Service Call

| Scenario | Use |
|----------|-----|
| Single contract created by user | Service.Create() directly |
| 100-row CSV import | Import pipeline |
| Nightly report generation | Export pipeline + Temporal |
| Real-time webhook ingestion | Event pipeline |
| Document OCR → structured data | Processing pipeline |

## Contract Import Pipeline

The contracts module provides a CSV import pipeline for bulk contract creation:

```
CSV bytes
    │
    ▼
ParseCSVStage       → []RawRow
    │
    ▼
ValidateRowStage    → []ValidatedRow  (accumulates RowErrors)
    │
    ▼
EnrichRowStage      → []EnrichedRow   (resolve vendor IDs, entity IDs)
    │
    ▼
PersistRowStage     → []CreatedContract (batch insert)
    │
    ▼
NotifyStage         → ImportResult
```
