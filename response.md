# Finance Module Review (Brutal Audit)

> Reviewed: `internal/core/finance/` and `docs/reference/modules/financial/`
> Date: 2026-04-28
> Verdict: **Do not ship. Do not call this "production-ready". Fix the compile errors first.**

---

## 1. Critical Issues (Must Fix)

### [CRIT-01] `NewServices` calls `NewTransactionService` with wrong argument count — compile error
- **Location:** `service/service.go:59–65`
- **Problem:** `NewTransactionService` signature requires 7 parameters (`repo`, `accountRepo`, `periodRepo`, `reversalHistoryRepo`, `entryService`, `tracing`, `metrics`). `NewServices` calls it with 5 (`TransactionRepo`, `AccountRepo`, `transactionEntryService`, `Tracing`, `Metrics`). Missing `periodRepo` and `reversalHistoryRepo`.
- **Why it's wrong:** This is a compile error. The binary cannot be built. The module does not run.
- **Suggested fix:** Pass `deps.PeriodRepo` and `nil` (or `deps.ReversalHistoryRepo` once added) as arguments 3 and 4.

---

### [CRIT-02] `TransactionEntryService` initialised with `nil` repository — guaranteed runtime panic
- **Location:** `service/service.go:51–56`
- **Problem:** `NewTransactionEntryService` receives `nil` as its first argument (the repository), with the actual repo call commented out and a TODO left in place.
- **Why it's wrong:** Every call that hits `s.repo.*` — `CreateEntry`, `CreateEntries`, `GetEntryByID`, `GetEntriesByTransaction`, `UpdateEntry`, `DeleteEntry`, `ReconcileEntries` — will nil-pointer panic. The entire entry subsystem is broken at initialisation time.
- **Suggested fix:** Create `domain.TransactionEntryRepository` interface, wire a concrete implementation, and pass it here.

---

### [CRIT-03] `CreateTransaction` saves the header but silently drops all entries
- **Location:** `service/transaction_service.go:146–191`
- **Problem:** `CreateTransactionRequest.Entries` is validated for balance and minimum count, but when building the `domain.Transaction` struct, entries are never persisted. The code stops after `s.repo.Create(ctx, transaction)`. No entry-creation calls follow.
- **Why it's wrong:** You produce transaction headers with zero journal lines. The general ledger is empty. Double-entry bookkeeping is not enforced. This is not a minor oversight — it makes the entire module non-functional from day one.
- **Suggested fix:** After the header is created, call `s.entryService.CreateEntries(ctx, entries)` within a DB transaction. If entries fail, roll back the header.

---

### [CRIT-04] `CreateTransaction` never sets `TenantID` on the new transaction
- **Location:** `service/transaction_service.go:146–157`
- **Problem:** The `domain.Transaction` struct built inside `CreateTransaction` has no `TenantID` field assignment. `TenantID` remains `uuid.Nil`.
- **Why it's wrong:** Every transaction lands in the DB with a null tenant ID. Row-level security is bypassed. Multi-tenant data isolation collapses. Any tenant can read any other tenant's transactions.
- **Suggested fix:** Extract tenant ID from context and assign it: `TenantID: tenantID`.

---

### [CRIT-05] `IsTransactionNumberUnique` duplicate-check returns `(nil, nil)` on duplicate — silent data corruption
- **Location:** `service/transaction_service.go:130–139`
- **Problem:**
  ```go
  if unique, err := s.repo.IsTransactionNumberUnique(ctx, req.EntityID, req.TransactionNumber, nil); err != nil || !unique {
      return nil, err
  }
  ```
  When `!unique && err == nil` (duplicate detected, no DB error), this returns `(nil, nil)`. The caller receives a nil error and nil transaction, interprets it as success, and the duplicate transaction proceeds to be created.
- **Why it's wrong:** Duplicate transaction numbers break reconciliation, external references, and ledger integrity. The guard that's supposed to prevent duplicates silently allows them.
- **Suggested fix:**
  ```go
  if err != nil {
      return nil, fmt.Errorf("uniqueness check failed: %w", err)
  }
  if !unique {
      return nil, domain.ErrTransactionNumberExists
  }
  ```

---

### [CRIT-06] Account balance update failure after posting is silently swallowed
- **Location:** `service/transaction_service.go:693–696`
- **Problem:**
  ```go
  if err := s.updateAccountBalances(ctx, entries); err != nil {
      logger.ErrorContext(ctx, "Failed to update account balances after posting", ...)
  }
  ```
  The error is logged and discarded. The transaction is marked POSTED but account balances are not updated.
- **Why it's wrong:** The GL shows a posted transaction but account balances are stale/wrong. Trial balance reports and financial statements are incorrect. This is a ledger integrity failure that will silently corrupt financial data.
- **Suggested fix:** Either propagate the error (requiring the caller to retry or compensate) or ensure this runs inside the same DB transaction as the post.

---

### [CRIT-07] `ReverseTransaction` marks original as reversed before the reversal is posted — leaves system in corrupt state on failure
- **Location:** `service/transaction_service.go:826–833`
- **Problem:**
  ```go
  if err := s.repo.Reverse(ctx, id, reversalTransaction.ID, reason); err != nil { ... }
  _, err = s.PostTransaction(ctx, reversalTransaction.ID, nil)
  if err != nil {
      return nil, fmt.Errorf("failed to post reversal transaction: %w", err)
  }
  ```
  The original transaction is flagged as reversed BEFORE the reversal transaction is posted. If `PostTransaction` fails (period closed, validation error, period repo nil, etc.), the original is permanently marked reversed with no valid reversal posted. The entry is gone from the ledger with no audit trail.
- **Why it's wrong:** You cannot undo a reversal flag without direct DB intervention. The original transaction becomes permanently inaccessible. This is an unrecoverable financial data corruption.
- **Suggested fix:** Mark original as reversed only AFTER the reversal is successfully posted, inside a single DB transaction.

---

### [CRIT-08] `ReverseTransaction` entire multi-step flow has no wrapping DB transaction
- **Location:** `service/transaction_service.go:813–833`
- **Problem:** The reversal flow: create reversal header → create entries one by one → mark original reversed → post reversal. Any failure partway through leaves partially created data with no rollback. If the 3rd entry creation fails, 2 reversal entries exist with no header completion.
- **Why it's wrong:** Partial reversals corrupt the ledger. Each individual step can succeed while the overall operation fails.
- **Suggested fix:** Wrap the entire reversal operation in a DB transaction (via `TxRunner`).

---

### [CRIT-09] `ListTransactions` dereferences `*req.Limit` before nil check — nil pointer panic
- **Location:** `service/transaction_service.go:477–488`
- **Problem:**
  ```go
  ctx, span := s.tracing.StartSpan(ctx, "...",
      tracing.WithAttributes(
          attribute.Int("limit", *req.Limit),   // ← deref before nil check
          attribute.Int("offset", *req.Offset),
      ))
  ...
  if *req.Limit <= 0 { // nil check comes here, 10 lines too late
  ```
- **Why it's wrong:** Any caller that passes a nil `Limit` panics before the nil guard runs.
- **Suggested fix:** Check and default `req.Limit` and `req.Offset` before the span.

---

### [CRIT-10] Context tenant key uses plain `string` type — tenant isolation silently broken
- **Location:** `service.go:81, 91`
- **Problem:**
  ```go
  if tenantID, ok := ctx.Value("tenant_id").(uuid.UUID); ok {
  ```
  Go context values keyed by plain strings are invisible to static analysis and collide trivially. Any middleware using a typed context key for `"tenant_id"` will cause this lookup to silently return `uuid.Nil`.
- **Why it's wrong:** The helper returns `uuid.Nil` silently, logs a warning, and proceeds. All downstream operations run with no tenant context. These helpers are never called in the visible service code anyway (see Design Issues), but the pattern propagates a dangerous anti-pattern.
- **Suggested fix:** Use an unexported type key:
  ```go
  type contextKey string
  const tenantIDKey contextKey = "tenant_id"
  ```

---

## 2. Mismatches Between Docs and Code

### [MISMATCH-01] Docs claim "double-entry enforced at the database layer" — no evidence it exists
- **Docs say:** `spec.md §1`: "The GL enforces double-entry integrity at the constraint level, not just the application layer. An unbalanced entry cannot exist in the database."
- **Code does:** Balance check is application-only (`IsBalanced()` in domain). No DB trigger, constraint, or CHECK exists in the visible codebase. `financial.sql` would need to prove otherwise.
- **Impact:** Bypass the application layer (direct DB insert, broken migration, test seed), and unbalanced entries land in the ledger with no DB-level guard.

---

### [MISMATCH-02] Docs list `TransactionTypeReversal` — code has no such type
- **Docs say:** The domain model references reversal as a distinct transaction type.
- **Code does:** `types.go` defines 11 types. No `REVERSAL` type. `CreateReversalTransaction()` in domain uses `TransactionTypeAdjustment`. `ReverseTransaction()` in service preserves the original transaction's type. Two paths, two inconsistent types, neither is `REVERSAL`.
- **Impact:** Reversal transactions cannot be reliably identified by type. Reporting and audit queries are unreliable.

---

### [MISMATCH-03] Docs describe Period gating as "enforced by trigger" — code skips it when `periodRepo` is nil
- **Docs say:** `spec.md §5`: "All posting goes through a period gate... enforced by a trigger."
- **Code does:** `postTransactionInline` at line 660: `if s.periodRepo != nil { ... period check ... }`. If `periodRepo` is nil (which it is in `NewServices` since it's not passed), the period gate is silently skipped. There's no DB trigger visible.
- **Impact:** Transactions can be posted into closed accounting periods. Period-end close is meaningless.

---

### [MISMATCH-04] Docs describe 10 services — top-level `Service` interface exposes 3
- **Docs say:** `Services` struct contains Account, Transaction, TransactionEntry, Period, ExchangeRate, Currency, CostCenter, Budget, Tax, Reconciliation.
- **Code does:** The top-level `Service` interface (`service.go:21–26`) only exposes `Account()`, `Transaction()`, `TransactionEntry()`. Period, ExchangeRate, Currency, CostCenter, Budget, Tax, Reconciliation are unreachable through the public interface.
- **Impact:** 7 of 10 advertised services are architecturally orphaned from the public API.

---

### [MISMATCH-05] Docs say reversal requires approval — code hardcodes `ApprovalRequired: false`
- **Docs say:** Approval workflow applies to significant financial operations including reversals.
- **Code does:** `ReverseTransaction` at line 809: `ApprovalRequired: false, ApprovalStatus: ApprovalStatusNotRequired`. Reversals bypass all approval controls unconditionally.
- **Impact:** Anyone with reversal permission can instantly reverse any posted transaction with no oversight. This is a financial control failure.

---

### [MISMATCH-06] Docs describe `JournalEntry` as aggregate root — code calls it `Transaction`
- **Docs say:** `spec.md §2`: Primary aggregate is `JournalEntry` with sub-entity `JournalLine`.
- **Code does:** Domain uses `Transaction` and `TransactionEntry`. No `JournalEntry` or `JournalLine` types exist.
- **Impact:** Documentation and code describe different models. Onboarding developers will be confused; specs cannot be traced to implementation.

---

## 3. Design Problems

### [DESIGN-01] `TransactionEntryService` uses `domain.TransactionRepository` — wrong repository type
- **Explanation:** `transactionEntryService.repo` is typed `domain.TransactionRepository`, not a `TransactionEntryRepository`. The comment in `service/service.go:33` even admits it: `// TODO: Create separate entry repository`. Entries and transactions share one repository. This means entry operations cannot be independently tested, scaled, or mocked. DDD aggregate boundary is violated.
- **Better approach:** Define `domain.TransactionEntryRepository` interface. Implement it separately. Inject it into the entry service.

---

### [DESIGN-02] `UpdateTransaction` accepts full `domain.Transaction` as mutation request
- **Explanation:** `UpdateTransaction(ctx, id, req domain.Transaction)` allows callers to overwrite `TenantID`, `CreatedBy`, `TransactionNumber`, `TransactionStatus`, and other fields that are either immutable or managed by system logic.
- **Better approach:** Define `UpdateTransactionRequest` DTO with only the mutable fields a caller is allowed to change. Apply the patch explicitly in the service.

---

### [DESIGN-03] `UpdateTransaction` allows mutation of `PENDING_APPROVAL`, `APPROVED`, `CANCELLED` transactions
- **Explanation:** The check at line 320 only blocks `POSTED`. `CanBeEdited()` in domain correctly limits to `DRAFT` and `REJECTED`. The service ignores its own domain logic.
- **Better approach:** Call `CanBeEdited()` in the service and return an appropriate error.

---

### [DESIGN-04] `financeService.getCurrentTenantID` / `getCurrentEntityID` — defined, never called
- **Explanation:** Two helper methods on the top-level `financeService` struct extract tenant/entity from context. None of the sub-services (`accountService`, `transactionService`, etc.) call them. They are dead code. The services operate without tenant context.
- **Better approach:** Either inject tenant context at construction time or pass it through service method signatures. Dead helpers give false confidence.

---

### [DESIGN-05] `ParseRejectionReason` implemented as a method receiver with unused `self`
- **Explanation:** `types.go:220`:
  ```go
  func (r RejectionReason) ParseRejectionReason(s string) (RejectionReason, error) {
  ```
  The receiver `r` is never used. You call this as `someRejection.ParseRejectionReason("OTHER")`, which is nonsensical. It should be `ParseRejectionReason(s string)` as a package-level function.
- **Better approach:** Remove the receiver. Make it a standalone function consistent with `ParseTransactionType` and `ParseTransactionStatus`.

---

### [DESIGN-06] `service/service.go` package declaration is broken
- **Explanation:** Line 1 is a one-liner comment that embeds `package finance` as plain text inside the comment, then line 3 declares `package service`. The package doc is a copy-pasted mess that includes a wrong package name and runs all doc text onto a single line without formatting.
- **Better approach:** Fix the package doc to be properly formatted multi-line godoc. Remove the embedded `package finance` text.

---

### [DESIGN-07] Two implementations of `PostTransaction` with divergent behaviour
- **Explanation:** `postTransactionViaPipeline` and `postTransactionInline` have different logic paths. The inline path checks period (conditionally), validates entries, and updates account balances. The pipeline path delegates all that to stages but then does a second repo fetch. Which path is active depends on how the service was constructed (`NewTransactionService` vs `NewTransactionServiceWithPipeline`). `NewServices` uses `NewTransactionService` (no pipeline). The two paths are not tested equivalently.
- **Better approach:** Pick one path. Delete the other. Having two divergent production code paths in a financial system is an audit nightmare.

---

## 4. Bugs & Edge Cases

### [BUG-01] `UpdateEntry` returns the pre-update (stale) entry
- **Scenario:** `UpdateEntry` reads `existingEntry` before the update, calls `repo.UpdateEntry`, then returns `existingEntry` (the old value).
- **Impact:** Callers receive the old entry data. Any downstream system that reads the "updated" entry will act on stale data.
- **Fix:** After `repo.UpdateEntry`, call `repo.GetEntryByID` to fetch and return the fresh record.

---

### [BUG-02] `DeleteEntry` allows deleting entries from posted transactions
- **Scenario:** Call `DeleteEntry` on an entry whose parent transaction is `POSTED`.
- **Impact:** The check at line 439 only blocks deletion of reconciled entries. There's no check that the parent transaction is posted. Deleting an entry from a posted transaction corrupts the GL — the transaction no longer balances.
- **Fix:** Fetch the parent transaction, check `TransactionStatus == POSTED`, return an error.

---

### [BUG-03] `CreateReversalTransaction` (domain) uses `TransactionTypeAdjustment` — inconsistent with service layer
- **Scenario:** `domain/transaction.go:559` uses `TransactionTypeAdjustment`. `service/transaction_service.go:802` uses `transaction.TransactionType` (original type). Two different callers produce different reversal types.
- **Impact:** The `IsSystemGenerated()` check will return different answers for reversals created through these two paths. Reporting is inconsistent.
- **Fix:** Add `TransactionTypeReversal`. Use it consistently in both paths.

---

### [BUG-04] `ValidateAmountConsistency` uses currency-agnostic tolerance of `0.01`
- **Scenario:** An entry in JPY (no decimal places) converts 100 JPY at rate 0.0091 → 0.91 USD. Tolerance is 0.01. Works. But for a 10,000,000 USD transaction, a 0.01 rounding error is still "acceptable" even though the absolute financial impact matters.
- **Impact:** Silent currency conversion errors pass validation. Functional currencies with different precision scales are handled incorrectly.
- **Fix:** Make tolerance currency-aware. Use the minor unit of the target currency (2 decimal places for USD, 0 for JPY, etc.).

---

### [BUG-05] `ReverseTransaction` returns draft-state reversal, not the posted state
- **Scenario:** After calling `PostTransaction`, the updated transaction is returned from that call but discarded: `_, err = s.PostTransaction(...)`. The function then returns `reversalTransaction` which is the draft-state struct built at line 798.
- **Impact:** Callers see a DRAFT transaction after a successful reversal. Status in API response is wrong.
- **Fix:** Capture the return value of `PostTransaction` and return it.

---

### [BUG-06] `GetTransactionByNumber` passes `nil` entity ID — cross-entity data leak
- **Scenario:** `s.repo.GetByNumber(ctx, nil, number)` — nil entity ID.
- **Impact:** If the repository doesn't enforce entity scoping when entity ID is nil, a transaction number search crosses entity boundaries within the same tenant. Separate legal entities' data leaks.
- **Fix:** Extract entity ID from context or require it as a parameter.

---

### [BUG-07] `CreateEntryRequest.Validate()` constructs dummy domain object and filters fake errors
- **Scenario:** Line 322–348 creates a `TransactionEntry` with `uuid.New()` for `TenantID` and `TransactionID` (values that always pass validation), runs the full `Validate()`, then manually filters out `tenant_id`, `transaction_id`, and `entry_number` errors.
- **Impact:** If validation logic for those fields changes, the filter silently drops real errors. This is fragile by design.
- **Fix:** Extract reusable validation helpers into private functions. Call them directly instead of constructing fake objects.

---

## 5. Financial Logic Risks

### [RISK-01] No atomicity between transaction post and entry creation
- **Explanation:** `CreateTransaction` saves header, then entries are created separately (when they exist at all — see CRIT-03). There is no database transaction wrapping both. A crash between header save and entry creation leaves an orphaned header.
- **Potential damage:** Orphaned transaction headers with no journal lines pollute the ledger. Balance reports are wrong. Auditors will flag unpaired entries.

---

### [RISK-02] Approval hardcoded to `false` for all new transactions
- **Explanation:** `CreateTransaction` at line 155: `ApprovalRequired: false`. The comment says "Will be determined by business rules" but no such rules exist.
- **Potential damage:** Every transaction bypasses the approval workflow unconditionally. No segregation of duties. High-value transactions post without any authorisation. SOX/IFRS compliance requirements are unmet.

---

### [RISK-03] Account balance materialisation can silently diverge from the ledger
- **Explanation:** `updateAccountBalances` failure is swallowed (CRIT-06). Additionally, it is not called in the pipeline path at all — the pipeline delegates to stages but balance update is not a named stage in the visible pipeline code.
- **Potential damage:** Trial balance does not match sum of journal entries. Financial statements are wrong. Audits will find unexplained discrepancies.

---

### [RISK-04] No period enforcement when `periodRepo` is nil (which it always is in `NewServices`)
- **Explanation:** `NewServices` does not pass `PeriodRepo` to `NewTransactionService`. It's missing from the call. Period gate skipped.
- **Potential damage:** Postings land in closed periods. Year-end close is meaningless. Prior-period financial statements are retroactively modified. This violates every accounting standard.

---

### [RISK-05] `GetNormalBalanceForRootType` is commented out — normal balance direction is never validated
- **Explanation:** `types.go:230–243` has the function commented out with no replacement. The comment in `rootTypeSet` documentation explicitly references this function but it doesn't exist.
- **Potential damage:** Debit/credit direction is never validated against account type. Assets can be credited to zero without warning. Revenue can be debited without warning. Financial statements silently invert.

---

### [RISK-06] Reversal bypasses approval and posts immediately with current exchange rate
- **Explanation:** `ReverseTransaction` calls `PostTransaction` directly. It also does not preserve the original transaction's exchange rate state at the time of reversal — it copies `transaction.ExchangeRate` but this may differ from the rate at original posting if the rate was updated since.
- **Potential damage:** FX reversals do not use the original rate, creating fictitious FX gains/losses in the P&L. Combined with no approval requirement, this is an uncontrolled financial event.

---

## 6. Code Quality Issues

### Naming
- `Accounts` (plural) is used as the name of a single account entity. Every reference reads `*domain.Accounts`, `account *domain.Accounts`. A single account is not "Accounts". Rename to `Account`.
- `TransactionEntry.EntryNumber` vs `TransactionEntry.ID` — both exist. `EntryNumber` is an `int32` sequential index. It duplicates the positional meaning already implied by slice order and creates ordering ambiguity when entries are returned in different orders.
- Error codes are inconsistent: `"not_posted"` (lowercase) in `ReverseTransaction:748` vs `"INVALID_STATUS"`, `"BUSINESS_RULE_ERROR"` (SCREAMING_SNAKE) everywhere else.
- `ErrCannotReverseReversal` — the variable name is fine but the domain sentinel is never returned directly from the service; instead the service constructs a `NewBusinessError` wrapping `domain.ErrCannotReverseReversal.Error()`. Mixing sentinel errors and string extraction from sentinels is inconsistent.

### Structure
- `service/service.go:33`: `TransactionEntryRepo domain.TransactionRepository // TODO: Create separate entry repository` — this TODO should be a tracked issue, not dead commented code in a production file.
- `domain/transaction.go:598–604`: 7 lines of TODO/NOTE comments at the bottom of the file describing features that don't exist. These belong in a backlog, not in source code.
- `domain/transaction_entry.go:540–546`: Same problem — 7 lines of TODO/NOTE at end of file.
- `domain/types.go:230–243`: `GetNormalBalanceForRootType` is commented out but still takes up 14 lines. Either implement it or delete it.

### Readability
- `service/service.go:1` — the package comment is a single unbroken line with `//` in the middle, "package finance" embedded in plain text, and no structure. It is worse than no comment at all.
- `postTransactionViaPipeline` logs "Transaction posted successfully via pipeline" but `postTransactionInline` logs "Transaction posted successfully" — no way to distinguish which path fired in production logs without reading source.

### Duplication
- Balance check logic is duplicated in `domain.Transaction.Validate()`, `domain.CreateTransactionRequest.Validate()`, and `postTransactionInline`. Three implementations of "debits must equal credits" that can diverge.
- Timer start/stop/histogram pattern is copy-pasted identically in every service method (20+ times). Extract a helper.
- `ReverseTransaction` in domain (`CreateReversalTransaction`) and `ReverseTransaction` in service both independently build a reversal object and swap debit/credit. Two reversal implementations, neither called through the other.

---

## 7. Missing Tests

### What is not tested (visible test files: `temporal_integration_test.go` only)
- **Zero unit tests** for any domain entity: `Transaction.Validate()`, `TransactionEntry.Validate()`, `IsBalanced()`, `CanBePosted()`, `CanBeReversed()`, `CreateReversalTransaction()`.
- **Zero unit tests** for service layer: `CreateTransaction`, `PostTransaction`, `ReverseTransaction`, `ApproveTransaction`.
- **Zero integration tests** for the ledger: no test verifies that posting a transaction updates account balances correctly.
- **Zero tests** for the period gate: posting into a closed period is untested.
- **Zero tests** for multi-currency: conversion consistency, FX gain/loss, tolerance edge cases.
- **Zero tests** for the approval workflow state machine.

### What must be tested
- Double-entry balance enforcement: every `Validate()` path, including entries that sum to zero with opposing signs.
- `CreateTransaction` → entries are persisted (currently would fail because CRIT-03 means entries are never saved).
- `ReverseTransaction` atomicity: simulate failure between steps and assert no partial state.
- `PostTransaction` with a nil `periodRepo` vs a closed period — confirm period gate fires.
- `ListTransactions` with nil `Limit` and nil `Offset` — confirm no panic.
- Account balance update failure after posting — confirm error propagation.
- Duplicate transaction number — confirm `ErrTransactionNumberExists`, not silent `(nil, nil)`.
- `UpdateEntry` returns post-update state, not pre-update state.
- `DeleteEntry` on a posted transaction — confirm rejection.
- `GetTransactionByNumber` with nil entity — confirm tenant/entity scoping.
- All `ApprovalStatus` / `TransactionStatus` state machine transitions.

---

## 8. Final Verdict

This module is **not production-ready**. It cannot even be compiled in its current state (CRIT-01). If you somehow fixed the compile error and ran it:

- Transactions would be created with no entries (CRIT-03), no tenant ID (CRIT-04), and broken duplicate detection (CRIT-05).
- Every entry operation would panic immediately because the entry repository is nil (CRIT-02).
- Posting would update the ledger but silently discard balance update failures (CRIT-06).
- Reversals leave the system in an unrecoverable corrupt state if posting fails mid-operation (CRIT-07, CRIT-08).
- The approval workflow is unconditionally bypassed for all transactions (RISK-02) including reversals (MISMATCH-05).
- Period enforcement is entirely absent (RISK-04) because `periodRepo` is never injected.
- Account normal balance direction is never validated because `GetNormalBalanceForRootType` is commented out (RISK-05).

The documentation describes a well-designed double-entry accounting system with period gating, approval workflows, and DB-level integrity constraints. The code delivers none of these guarantees. The architecture is recognisable but the implementation is incomplete, the critical paths are untested, and the financial data integrity controls are missing or broken.

Fix the compile errors. Write the entries. Add the DB transaction. Then come back.
