# Finance Module — Phase 17: Financial System Consistency Layer

**Date:** 2026-05-11
**Scope:** `internal/core/finance/**` — COA snapshot model, deterministic reporting, ledger replay, cross-system consistency scanning
**Status:** IMPLEMENTATION COMPLETE — pending `go test` execution

---

## Executive Summary

Phase 17 binds the entire financial system into one consistent, reproducible truth model. Any financial report or account balance is now fully deterministic from three canonical inputs: (1) COA snapshot, (2) posted transaction ledger, and (3) posting rules. The system detects silent data corruption, ledger mutation, and COA drift before they reach financial statements.

---

## Phase A — Financial Consistency Domain Model

### New Production Code

**`internal/core/finance/domain/consistency.go`**

Defines the cross-system invariant contract:

| Type | Purpose |
|------|---------|
| `COASnapshot` | Immutable point-in-time COA structure. Reports bind to this, never the live COA. |
| `COASnapshotEntry` | Single frozen account node. `IsPostingEligible()` drives eligibility checks. |
| `COASnapshotRepository` | Persistence interface: Create, GetByID, GetForPeriod, GetAtTime, List. |
| `FinancialReportHash` | Deterministic SHA-256 fingerprint of report output + ledger + snapshot. |
| `ConsistencyViolation` | Single cross-system invariant failure with kind + severity + context. |
| `SystemConsistencyReport` | Full scan output: COA ↔ tx ↔ audit ↔ reports. |
| `ReplayedBalance` / `ReplayReport` | Ledger-replay output types for drift detection. |
| `TrialBalanceReport` / `TrialBalanceLine` | Deterministic, snapshot-bound trial balance. |

**8 ConsistencyViolationKind constants:**
`ORPHAN_POSTING`, `ACCOUNT_INELIGIBLE_AT_POSTING`, `MISSING_AUDIT_ENTRY`, `BALANCE_DRIFT`, `SNAPSHOT_MISMATCH`, `REPORT_HASH_MISMATCH`, `COA_MUTATION_WITHOUT_SNAPSHOT`, `UNREPLAYABLE_LEDGER`

---

## Phase B — COA Snapshot Service

### New Production Code

**`internal/core/finance/service/coa_snapshot.go`** — `COASnapshotService`

- `CreateSnapshot(ctx, req)` — loads all accounts, sorts canonically by AccountCode, computes SHA-256 over canonical JSON of each entry, persists immutable snapshot.
- `ValidateSnapshotHash(ctx, snapshotID)` — recomputes hash from stored entries; returns `(false, nil)` on mismatch (tamper signal).
- `GetByID`, `GetForPeriod`, `GetAtTime` — snapshot retrieval.
- Nil-safe receiver.

**Hash computation:** SHA-256 of concatenated `json.Marshal(entry)` for each entry in AccountCode-ascending order. Identical COA → identical hash on every call.

---

## Phase C — Ledger Replay Reconstructor

### New Production Code

**`internal/core/finance/service/replay_reconstructor.go`** — `ReplayReconstructor`

- `ReplayTenant(ctx, asOfDate)` — loads all POSTED transactions, aggregates debit/credit per account, compares against stored `CurrentBalance`, returns `ReplayReport` with `DriftAccounts` and `ReplayClean` flag.
- `ReplayAccount(ctx, accountID, asOfDate)` — targeted per-account replay.
- Drift = stored balance ≠ replayed net (debit − credit).
- Nil-safe receiver.

**What drift signals:** Direct-SQL balance updates, migration defects, bug-induced resets, and any write that bypasses the transaction ledger.

---

## Phase D — Deterministic Reporting Engine

### New Production Code

**`internal/core/finance/service/deterministic_report.go`** — `DeterministicReportingEngine`

- `GenerateTrialBalance(ctx, req)` — binds to a COA snapshot, aggregates POSTED ledger entries up to AsOfDate, produces sorted `TrialBalanceReport` with SHA-256 hash.
- `VerifyReportHash(ctx, stored, generatedBy)` — regenerates the report from the same inputs and compares against a stored `FinancialReportHash`; returns `(false, nil)` when ledger has been mutated or report is non-deterministic.
- Report hash = SHA-256 of (snapshotHash + ledgerHash + sorted line JSON).
- Nil-safe receiver.

**Invariant:** `GenerateTrialBalance(same snapshot, same ledger, same asOfDate)` → identical `Hash` always.

---

## Phase E — Cross-System Consistency Scanner

### New Production Code

**`internal/core/finance/service/consistency_scanner.go`** — `FinancialConsistencyScanner`

- `ScanSystem(ctx, asOfDate)` — full cross-system scan returning `SystemConsistencyReport`.

**Checks performed:**

| Check | Violation Kind | Severity |
|-------|---------------|---------|
| Entry references account absent from COA | `ORPHAN_POSTING` | CRITICAL |
| Account ineligible at posting time (via snapshot lookup) | `ACCOUNT_INELIGIBLE_AT_POSTING` | CRITICAL |
| Closed/archived account with non-zero balance appears in ledger | `BALANCE_DRIFT` | CRITICAL |
| No snapshot exists for tenant with accounts | `COA_MUTATION_WITHOUT_SNAPSHOT` | HIGH |

`Healthy = false` when any CRITICAL or HIGH violation present.

---

## Phase F — Tests

### New Test Code (4 files, 45 tests)

| File | Tests | Coverage |
|------|-------|---------|
| `coa_snapshot_test.go` | FIN-SNAP-001–012 | Create, hash, determinism, tamper detection, GetAtTime, error paths |
| `replay_reconstructor_test.go` | FIN-REPLAY-001–010 | Empty ledger, drift detection, per-account replay, nil-safe, error propagation |
| `deterministic_report_test.go` | FIN-DREPORT-001–011 | Balanced report, hash determinism, sort order, totals, verify reproducibility, ledger mutation detection |
| `consistency_scanner_test.go` | FIN-SCAN-001–012 | Orphan posting, balance drift, no snapshot, healthy/unhealthy, nil-safe, entry counts |
| `p17_stubs_test.go` | — | `p17StubSnapshotRepo` (in-memory), `p17AccountRepo` (wraps p16) |

---

## Invariant Coverage Summary

| Invariant | Enforced by |
|-----------|------------|
| Every posted entry resolves to valid COA account | `FinancialConsistencyScanner.ScanSystem` → `ORPHAN_POSTING` |
| Every report derived from time-bound COA snapshot | `DeterministicReportingEngine` (SnapshotID required) |
| No report may depend on runtime-only derived state | Snapshot binding + hash verification |
| Audit chain sufficient to reconstruct any report | `FinancialReportHash` persistence (caller responsibility) |
| COA mutations time-stamped and leave snapshot trail | `COASnapshotService.CreateSnapshot` + `COA_MUTATION_WITHOUT_SNAPSHOT` check |
| Balance from ledger replay equals stored balance | `ReplayReconstructor.ReplayTenant` → `DriftAccounts` |

---

## Run Tests

```bash
go test ./internal/core/finance/service/... \
  -run "TestSnapshot|TestReplay|TestDReport|TestScan" \
  -v -cover
```

Full suite (all phases):

```bash
go test ./internal/core/finance/... -cover -v
```
