# Finance Module Production Hardening Report

## 1. Summary

Audited `internal/core/finance/**` end-to-end — domain, service, tests.
Found and fixed **5 code bugs**, **1 silent data-integrity failure**, **1 missing security check**, **1 metric gap**.
Added **9 invariant/regression/concurrency tests**.

| Risk Category | Bugs Found | Fixed |
|---|---|---|
| Wrong type (double-reversal bypass) | 1 | ✅ |
| Wrong identity + hardcoded reason (SOD bypass in reject) | 2 | ✅ |
| Silent data-integrity failure (reconcile not-found skip) | 2 | ✅ |
| Missing tenant isolation in PostTransaction | 1 | ✅ |
| Dead TODO in domain (misleading, hides missing period enforcement) | 1 | ✅ |
| Missing metric (reconciliation_errors) | 1 | ✅ |
| Tests for above invariants | 0 | ✅ 9 added |

---

## 2. Financial Invariants Enforcement

### Rules enforced and where

| Invariant | Domain | Service | DB |
|---|---|---|---|
| Debits == Credits | ✅ `IsBalanced()` + `Validate()` | ✅ `ValidateTransaction()` blocks post | ✅ DB trigger |
| No post to inactive account | ✅ `ValidateBusinessRules()` | ✅ `ValidateTransaction()` per-entry | — |
| No post to closed period | — (infra dep) | ✅ `periodRepo.GetPeriodForDate()` + `AllowsPosting()` | ✅ DB trigger (hard_closed/locked) |
| Cannot reverse a reversal | ✅ `TransactionTypeReversal` distinct type | ✅ `reversalHistoryRepo.IsReversal()` + IsReversed flag | — |
| Approval before post | ✅ `CanBePosted()` | ✅ status switch in inline path | — |
| Idempotent transaction number | — | ✅ `IsTransactionNumberUnique()` pre-check | ✅ unique constraint |

### Bug Fixed: Wrong reversal type in domain

**File:** `domain/transaction.go`
**Bug:** `CreateReversalTransaction()` set `TransactionType = TransactionTypeAdjustment`.
**Impact:** Code calling the domain method directly (not the service) produced reversals typed as ADJUSTMENT. The double-reversal guard checks for `TransactionTypeReversal` — ADJUSTMENT bypassed it. Reversals of reversals could slip through.
**Fix:** Changed to `TransactionTypeReversal`.
**Test:** `TestINV005_DomainCreateReversalTransaction_TypeMustBeReversal`

---

## 3. Concurrency & Consistency Fixes

### Findings

| Issue | Severity | Status |
|---|---|---|
| `CreateTransaction`: header + entries not in single DB tx | HIGH | Documented; best-effort cleanup exists. TxRunner must be wired. |
| `CreateRecurringTransaction`: no rollback on partial entry failure | HIGH | Documented. TxRunner must be wired for this path. |
| `ReverseTransaction` atomic path: correct when TxRunner injected | OK | Best-effort path has CRITICAL log on step-4 failure |
| Concurrent `PostTransaction`: no panic (validated) | MEDIUM | ✅ Test added |

### What was fixed

Concurrent posting test (`TestCONC001`) verifies 10 goroutines calling `PostTransaction` on same ID produce no panics. DB optimistic locking via `version` field is the correctness backstop.

**Outstanding — requires TxRunner wiring:**
`CreateTransaction` and `CreateRecurringTransaction` have non-atomic header + entry creation. The `TxRunner` interface exists and is wired for `ReverseTransaction`. It must also be wired for creation paths.

---

## 4. Database Safety Improvements

### Present (validated)
- `transaction_number` unique constraint per entity
- DB trigger blocks posting to `hard_closed` / `locked` periods
- Soft deletes on all financial tables (`deleted_at`)
- `account_code` CHECK constraint (8-digit pattern)
- `account_status` CHECK constraint

### Missing — flagged for migration (DBA action required)

1. **No CHECK enforcing `total_debit = total_credit` at DB level.** Service enforces it; a direct SQL INSERT bypasses it.
   **Recommended:** `ADD CONSTRAINT chk_txn_balanced CHECK (total_debit_amount = total_credit_amount) DEFERRABLE INITIALLY DEFERRED` (defer until end of posting tx).

2. **No CHECK preventing both debit > 0 and credit > 0 on same entry row.**
   **Recommended:** `ADD CONSTRAINT chk_entry_single_side CHECK (NOT (debit_amount > 0 AND credit_amount > 0))`.

3. **Optimistic locking in `repo.Update()` must be `WHERE id = $1 AND version = $old_version`.** Verify in repo SQL — not auditable from service layer.

### Tenant isolation bug fixed

**Bug:** `postTransactionInline` loaded a transaction by ID without verifying the record's `TenantID` matched the caller's tenant. Authenticated caller from tenant A could post tenant B's transaction given a known UUID.

**Fix:** Added tenant mismatch check before any status logic. Returns `TENANT_MISMATCH` / HTTP 403 + ERROR log flagged as SECURITY on violation.

**Test:** `TestINV009_PostTransaction_CrossTenantAccess_MustFail`

---

## 5. Observability Enhancements

### Logging — strong
Every mutation logs who (caller identity), what (IDs, numbers), and outcome. Errors include full context. CRITICAL log on reversal step-4 failure.

### Metrics

**Pre-existing:**
- `transactions_posted_total` (post success)
- `transaction_posting_errors` with `error_type` tag (post failure)
- `transaction_posting_duration` histogram
- `transaction_approval_duration` histogram

**Added:**
- `reconciliation_errors` counter with `error_type` tag in `ReconcileEntries` and `UnreconcileEntries`

### Tracing
All service methods have OTel spans. Pipeline and inline paths are traced. No gaps found.

---

## 6. Anti-Corruption Fixes

### Domain vs service reversal type disagreement (fixed)

`domain.CreateReversalTransaction()` produced ADJUSTMENT type; service produced REVERSAL. Now both produce REVERSAL. All guards that check transaction type now work correctly regardless of which path creates the reversal.

### Dead code in domain (fixed)

`transaction.ValidateBusinessRules()` had a dead `TODO` block implying period validation would happen in domain. Removed — replaced with explanatory comment directing readers to the service layer where `PeriodRepository` is available.

---

## 7. Test Hardening

New file: `service/transaction_hardening_test.go`

| Test | Guarantee |
|---|---|
| `TestINV001` | Unbalanced transaction (debits ≠ credits) NEVER posts |
| `TestINV002` | Cannot reverse a reversal |
| `TestINV003` | Posting to hard-closed period fails with PERIOD_CLOSED |
| `TestINV004` | Posting to inactive account fails |
| `TestINV005` | Domain `CreateReversalTransaction` type is REVERSAL not ADJUSTMENT (regression) |
| `TestINV006` | Submitter cannot reject own transaction (SOD) |
| `TestINV007` | `RejectTransaction` records approver identity, not submitter (regression) |
| `TestINV008` | `ReconcileEntries` fails on not-found — does NOT silently skip (regression) |
| `TestINV009` | Cross-tenant posting blocked |
| `TestCONC001` | 10 concurrent `PostTransaction` calls — no panic |

---

## 8. Fragility Eliminated

### Bug: `RejectTransaction` — three compounding bugs

**File:** `service/transaction.go`

| Sub-bug | Description | Fixed |
|---|---|---|
| Wrong rejectedBy | `rejectedBy` was always `transaction.CreatedBy` (submitter) | ✅ Now uses `rejectorID` from context |
| Hardcoded reason | `rejectionReason` was always `RejectionReasonOther` regardless of input | ✅ Caller passes `reason`; interface updated |
| No SOD check | Submitter could reject their own pending transaction | ✅ `rejectorID == CreatedBy → SOD_VIOLATION` |
| No auth check | No error if caller has no identity | ✅ `MISSING_REJECTOR` / HTTP 401 |

Interface changed: `RejectTransaction(ctx, id, reason domain.RejectionReason, notes string)`.

### Bug: `ReconcileEntries` / `UnreconcileEntries` — silent skip on not-found

**File:** `service/transaction_entry.go`

**Bug:** When `GetEntryByID` failed during reconciliation, code logged a warning and `continue`d. Caller received `nil` error, believed all entries were reconciled.
**Impact:** Silent incomplete reconciliation. Financial statements could be wrong.
**Fix:** Not-found now aborts the batch with an error. Already-reconciled entries still idempotently skip.

### Hidden coupling: pipeline path lacks tenant isolation

`postTransactionViaPipeline` does not have the tenant isolation check that was added to `postTransactionInline`. If the pipeline is active, tenant mismatch is not caught at service level. **Pipeline must add a tenant isolation stage** before the GL write.

### Magic value: `"BIWEEKLY"` string literal

`service/transaction.go:1648` — `case "BIWEEKLY"` not backed by a constant. If domain adds `RecurringFrequencyBiweekly` constant later, this case becomes dead code silently. Define constant and use it.

---

## 9. Performance Safeguards

### N+1 query in `ValidateTransaction`

`s.accountRepo.GetByID(ctx, entry.AccountID)` called per entry in a loop. For a 20-entry transaction = 20 sequential DB round-trips before posting.

**Status:** Not fixed here (requires `GetByIDs` batch method on `AccountsRepository`). Flagged for next sprint. The `CreateEntries` bulk path already caches accounts.

### Pagination enforced

| Method | Default | Max |
|---|---|---|
| `ListTransactions` | 50 | 1000 |
| `GetEntriesByAccountID` | 100 | 1000 |
| `SearchTransactions` | 50 | 200 |
| `SearchEntries` | 100 | 500 |

No full-table scans exposed through service layer.

---

## 10. Final Verdict

**CONDITIONALLY SAFE FOR PRODUCTION**

| Condition | Status |
|---|---|
| All bugs in this audit | ✅ FIXED |
| Invariant tests locking regressions | ✅ DONE |
| `TxRunner` wired into `CreateTransaction` for atomicity | ⚠️ REQUIRED before GA |
| `TxRunner` wired into `CreateRecurringTransaction` | ⚠️ REQUIRED before GA |
| DB CHECK constraints (balanced, single-side) | ⚠️ Migration required |
| Tenant isolation in pipeline path | ⚠️ Pipeline stage required |
| N+1 batch account fetch in `ValidateTransaction` | ⚠️ Performance risk at scale |

The module is safe under single-user, single-request conditions with all guards in place. The critical production blocker is `CreateTransaction` non-atomicity — if `repo.Delete` best-effort cleanup fails after an entry error, an orphaned DRAFT header persists in the DB. This is tolerable only if a reconciliation sweep runs and is monitored. Wire `TxRunner` before going to production at scale.
