# Finance Module Review (Brutal Audit)

> Reviewed: `internal/core/finance/` and `docs/reference/modules/financial/`
> Date: 2026-04-28
> Verdict: **Do not ship. Do not call this "production-ready". Fix the compile errors first.**

---


<!-- toc -->

- [1. Critical Issues (Must Fix)](#1-critical-issues-must-fix)
  - [[CRIT-01] `NewServices` calls `NewTransactionService` with wrong argument count — compile error](#crit-01-newservices-calls-newtransactionservice-with-wrong-argument-count--compile-error)
  - [[CRIT-02] `TransactionEntryService` initialised with `nil` repository — guaranteed runtime panic](#crit-02-transactionentryservice-initialised-with-nil-repository--guaranteed-runtime-panic)
  - [[CRIT-03] `CreateTransaction` saves the header but silently drops all entries](#crit-03-createtransaction-saves-the-header-but-silently-drops-all-entries)
  - [[CRIT-04] `CreateTransaction` never sets `TenantID` on the new transaction](#crit-04-createtransaction-never-sets-tenantid-on-the-new-transaction)
  - [[CRIT-05] `IsTransactionNumberUnique` duplicate-check returns `(nil, nil)` on duplicate — silent data corruption](#crit-05-istransactionnumberunique-duplicate-check-returns-nil-nil-on-duplicate--silent-data-corruption)
  - [[CRIT-06] Account balance update failure after posting is silently swallowed](#crit-06-account-balance-update-failure-after-posting-is-silently-swallowed)
  - [[CRIT-07] `ReverseTransaction` marks original as reversed before the reversal is posted — leaves system in corrupt state on failure](#crit-07-reversetransaction-marks-original-as-reversed-before-the-reversal-is-posted--leaves-system-in-corrupt-state-on-failure)
  - [[CRIT-08] `ReverseTransaction` entire multi-step flow has no wrapping DB transaction](#crit-08-reversetransaction-entire-multi-step-flow-has-no-wrapping-db-transaction)
  - [[CRIT-09] `ListTransactions` dereferences `*req.Limit` before nil check — nil pointer panic](#crit-09-listtransactions-dereferences-reqlimit-before-nil-check--nil-pointer-panic)
  - [[CRIT-10] Context tenant key uses plain `string` type — tenant isolation silently broken](#crit-10-context-tenant-key-uses-plain-string-type--tenant-isolation-silently-broken)
- [2. Mismatches Between Docs and Code](#2-mismatches-between-docs-and-code)
  - [[MISMATCH-01] Docs claim "double-entry enforced at the database layer" — no evidence it exists](#mismatch-01-docs-claim-double-entry-enforced-at-the-database-layer--no-evidence-it-exists)
  - [[MISMATCH-02] Docs list `TransactionTypeReversal` — code has no such type](#mismatch-02-docs-list-transactiontypereversal--code-has-no-such-type)
  - [[MISMATCH-03] Docs describe Period gating as "enforced by trigger" — code skips it when `periodRepo` is nil](#mismatch-03-docs-describe-period-gating-as-enforced-by-trigger--code-skips-it-when-periodrepo-is-nil)
  - [[MISMATCH-04] Docs describe 10 services — top-level `Service` interface exposes 3](#mismatch-04-docs-describe-10-services--top-level-service-interface-exposes-3)
  - [[MISMATCH-05] Docs say reversal requires approval — code hardcodes `ApprovalRequired: false`](#mismatch-05-docs-say-reversal-requires-approval--code-hardcodes-approvalrequired-false)
  - [[MISMATCH-06] Docs describe `JournalEntry` as aggregate root — code calls it `Transaction`](#mismatch-06-docs-describe-journalentry-as-aggregate-root--code-calls-it-transaction)
- [3. Design Problems](#3-design-problems)
  - [[DESIGN-01] `TransactionEntryService` uses `domain.TransactionRepository` — wrong repository type](#design-01-transactionentryservice-uses-domaintransactionrepository--wrong-repository-type)
  - [[DESIGN-02] `UpdateTransaction` accepts full `domain.Transaction` as mutation request](#design-02-updatetransaction-accepts-full-domaintransaction-as-mutation-request)
  - [[DESIGN-03] `UpdateTransaction` allows mutation of `PENDING_APPROVAL`, `APPROVED`, `CANCELLED` transactions](#design-03-updatetransaction-allows-mutation-of-pending_approval-approved-cancelled-transactions)
  - [[DESIGN-04] `financeService.getCurrentTenantID` / `getCurrentEntityID` — defined, never called](#design-04-financeservicegetcurrenttenantid--getcurrententityid--defined-never-called)
  - [[DESIGN-05] `ParseRejectionReason` implemented as a method receiver with unused `self`](#design-05-parserejectionreason-implemented-as-a-method-receiver-with-unused-self)
  - [[DESIGN-06] `service/service.go` package declaration is broken](#design-06-serviceservicego-package-declaration-is-broken)
  - [[DESIGN-07] Two implementations of `PostTransaction` with divergent behaviour](#design-07-two-implementations-of-posttransaction-with-divergent-behaviour)
- [4. Bugs & Edge Cases](#4-bugs--edge-cases)
  - [[BUG-01] `UpdateEntry` returns the pre-update (stale) entry](#bug-01-updateentry-returns-the-pre-update-stale-entry)
  - [[BUG-02] `DeleteEntry` allows deleting entries from posted transactions](#bug-02-deleteentry-allows-deleting-entries-from-posted-transactions)
  - [[BUG-03] `CreateReversalTransaction` (domain) uses `TransactionTypeAdjustment` — inconsistent with service layer](#bug-03-createreversaltransaction-domain-uses-transactiontypeadjustment--inconsistent-with-service-layer)
  - [[BUG-04] `ValidateAmountConsistency` uses currency-agnostic tolerance of `0.01`](#bug-04-validateamountconsistency-uses-currency-agnostic-tolerance-of-001)
  - [[BUG-05] `ReverseTransaction` returns draft-state reversal, not the posted state](#bug-05-reversetransaction-returns-draft-state-reversal-not-the-posted-state)
  - [[BUG-06] `GetTransactionByNumber` passes `nil` entity ID — cross-entity data leak](#bug-06-gettransactionbynumber-passes-nil-entity-id--cross-entity-data-leak)
  - [[BUG-07] `CreateEntryRequest.Validate()` constructs dummy domain object and filters fake errors](#bug-07-createentryrequestvalidate-constructs-dummy-domain-object-and-filters-fake-errors)
- [5. Financial Logic Risks](#5-financial-logic-risks)
  - [[RISK-01] No atomicity between transaction post and entry creation](#risk-01-no-atomicity-between-transaction-post-and-entry-creation)
  - [[RISK-02] Approval hardcoded to `false` for all new transactions](#risk-02-approval-hardcoded-to-false-for-all-new-transactions)
  - [[RISK-03] Account balance materialisation can silently diverge from the ledger](#risk-03-account-balance-materialisation-can-silently-diverge-from-the-ledger)
  - [[RISK-04] No period enforcement when `periodRepo` is nil (which it always is in `NewServices`)](#risk-04-no-period-enforcement-when-periodrepo-is-nil-which-it-always-is-in-newservices)
  - [[RISK-05] `GetNormalBalanceForRootType` is commented out — normal balance direction is never validated](#risk-05-getnormalbalanceforroottype-is-commented-out--normal-balance-direction-is-never-validated)
  - [[RISK-06] Reversal bypasses approval and posts immediately with current exchange rate](#risk-06-reversal-bypasses-approval-and-posts-immediately-with-current-exchange-rate)
- [6. Code Quality Issues](#6-code-quality-issues)
  - [Naming](#naming)
  - [Structure](#structure)
  - [Readability](#readability)
  - [Duplication](#duplication)
- [7. Missing Tests](#7-missing-tests)
  - [What is not tested (visible test files: `temporal_integration_test.go` only)](#what-is-not-tested-visible-test-files-temporal_integration_testgo-only)
  - [What must be tested](#what-must-be-tested)
- [Finance Module — Brutal Code Review](#finance-module--brutal-code-review)
- [PART 1: Application Layer Review](#part-1-application-layer-review)
  - [CRITICAL BUGS](#critical-bugs)
  - [DOMAIN / DESIGN ISSUES](#domain--design-issues)
  - [DOC vs CODE MISMATCHES](#doc-vs-code-mismatches)
  - [FINANCIAL CORRECTNESS RISKS](#financial-correctness-risks)
  - [MISSING TESTS](#missing-tests)
- [PART 2: Database & Infrastructure Audit](#part-2-database--infrastructure-audit)
  - [SCHEMA — CRITICAL](#schema--critical)
  - [QUERY LAYER — CRITICAL](#query-layer--critical)
  - [INFRASTRUCTURE — CRITICAL](#infrastructure--critical)
  - [OBSERVABILITY GAPS](#observability-gaps)
  - [SUMMARY TABLE](#summary-table)
- [8. Final Verdict](#8-final-verdict)

<!-- tocstop -->

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


## Finance Module — Brutal Code Review

---

## PART 1: Application Layer Review

### CRITICAL BUGS

**CRIT-01 — `service.go`: `NewTransactionService` called with wrong arg count** *(FIXED by user)*
`transactionEntryService` initialized BEFORE being passed to `NewTransactionService`. Original call passed wrong number of args — compile error. Fixed during session.

**CRIT-02 — `service.go`: `TransactionEntryService.repo` is always nil**
```go
transactionEntryService := NewTransactionEntryService(
    // deps.TransactionEntryRepo,  ← COMMENTED OUT
    nil,
    ...
)
```
Every method on `transactionEntryService` that calls `s.repo.*` panics at runtime. This is not a theoretical bug — it is a guaranteed panic on first use.

**CRIT-03 — `transaction_entry_service.go`: `repo` field has wrong type**
```go
type transactionEntryService struct {
    repo domain.TransactionRepository  // ← WRONG: should be TransactionEntryRepository
```
Even if the nil is fixed, assigning a `TransactionEntryRepo` to a `TransactionRepository` field would be a type error. The struct was copy-pasted from `transactionService` and never corrected.

**CRIT-04 — `transaction_entry_service.go:UpdateEntry`: returns stale pre-update value**
```go
existingEntry, _ := s.repo.GetEntryByID(ctx, req.EntryID)
// ... update ...
return existingEntry, nil  // ← returns PRE-update state
```
Callers receive the old entry. No re-fetch after update. Silent data corruption at API boundary.

**CRIT-05 — `transaction_entry_service.go:DeleteEntry`: deletes entries from POSTED transactions**
Only checks `entry.Reconciled` before deleting. No check on parent transaction status. Deleting a journal entry from a POSTED transaction corrupts the general ledger with no recovery path.

**CRIT-06 — `transaction_service.go:updateAccountBalances`: error silently swallowed**
Balance update failure returns no error to caller. Posting "succeeds" while account balances remain wrong. The ledger is now inconsistent with no log entry, no metric, no alert.

**CRIT-07 — `transaction_service.go:ReverseTransaction`: marks original reversed BEFORE posting reversal**
```go
original.IsReversed = true  // marked here
s.repo.Update(ctx, original)
// then posts reversal — can fail
s.PostTransaction(ctx, ...)
```
If post fails, original is permanently flagged as reversed with no reversal transaction. Unrecoverable state.

**CRIT-08 — `transaction_service.go:ReverseTransaction`: no DB transaction wrapping**
Original update, reversal create, reversal post — three separate DB writes with no enclosing transaction. Any failure leaves partial state. A network blip between steps = corrupted books.

**CRIT-09 — `transaction_service.go:ListTransactions`: nil pointer dereference**
```go
span.SetAttributes(attribute.Int("limit", *req.Limit))  // dereference before nil check
if req.Limit != nil { ... }
```
Request with no `Limit` set panics in the span attribute call before the nil guard.

**CRIT-10 — `transaction_service.go:postTransactionInline`: bypasses approval check**
The inline post path allows DRAFT→POSTED even when `ApprovalRequired=true`. The pipeline path has the guard; inline path does not. The approval workflow is bypassable depending on which code path executes.

---

### DOMAIN / DESIGN ISSUES

**DESIGN-01 — `transaction_entry.go:ValidateAmountConsistency`: hardcoded 0.01 tolerance**
KWD/IQD/OMR use 3 decimal places. A valid 0.001 difference in those currencies triggers a false validation error. Tolerance must be currency-aware.

**DESIGN-02 — `transaction_entry.go:CreateEntryRequest.Validate`: dummy UUID hack**
Creates a throw-away object just to reuse `ValidateBusinessRules`. This is a validation framework problem, not a workaround worth keeping.

**DESIGN-03 — `transaction_entry.go:ValidateBusinessRules`: FIXME left in production**
```go
// FIXME: Check AllowManualEntries on the account
```
The check is skipped entirely. Accounts with `allow_manual_entries=false` can receive manual journal entries. Business rule violation, not a cosmetic todo.

**DESIGN-04 — `transaction_entry.go`: deprecated field still in active use**
`CostCenter *string` marked deprecated in favor of `CostCenterID *uuid.UUID`, but `GetDimensionalAnalysis()` still reads `CostCenter`. Reports built on this produce wrong dimensional data without error.

**DESIGN-05 — `domain/errors.go`: missing `ErrTransactionCancelled`**
No sentinel for cancelled transaction state. Code that tries to edit a CANCELLED transaction has no clean error to return — falls through to generic error or nil which the caller can't distinguish.

---

### DOC vs CODE MISMATCHES

| Area | Doc Says | Code Does |
|------|----------|-----------|
| Reversal | Creates and posts reversal atomically | Creates reversal in DRAFT, requires separate post call |
| Approval | Required when `ApprovalRequired=true` before posting | Bypassed in inline post path |
| Entry validation | `AllowManualEntries` enforced | Skipped (FIXME) |
| Amount precision | 4 decimal places supported | Hardcoded 0.01 tolerance ignores currency |
| `UpdateEntry` | Returns updated entry | Returns pre-update stale entry |

---

### FINANCIAL CORRECTNESS RISKS

- No idempotency key on `PostTransaction` — double-posting possible under retry
- `ReverseTransaction` not atomic — partial reversal leaves books in undefined state
- Account balance update errors silently dropped — ledger can drift from reality indefinitely
- No period-close guard in service layer — postings to closed periods not blocked at application level
- `GetDimensionalAnalysis()` uses deprecated string cost center — cost center reporting unreliable

---

### MISSING TESTS

- `ReverseTransaction` with post failure mid-way
- `postTransactionInline` with `ApprovalRequired=true` (approval bypass)
- `DeleteEntry` on entry belonging to POSTED transaction
- `UpdateEntry` return value correctness (stale vs fresh)
- `ListTransactions` with nil `req.Limit` (nil deref)
- `updateAccountBalances` error path — verify it propagates
- Currency-specific amount tolerance in `ValidateAmountConsistency`
- `ValidateBusinessRules` with `AllowManualEntries=false`

---

## PART 2: Database & Infrastructure Audit

*(Sources: `db/migration/0009xx_*.sql`, `db/queries/finance_*.sql`, `internal/shared/{logger,metrics,tracing}`, `internal/platform/cache`)*

---

### SCHEMA — CRITICAL

**SCHEMA-01 — `000905`: Journal entry amounts use `DECIMAL(15,2)` — precision loss for 3+ decimal currencies**
```sql
debit_amount    DECIMAL(15,2)
credit_amount   DECIMAL(15,2)
original_amount DECIMAL(15,2)
```
KWD, IQD, OMR, JOD, BHD all use 3 decimal places. Every posting in these currencies silently rounds to 2 dp. Financial statements are wrong by design for these markets. Transaction header uses `DECIMAL(15,4)` — already inconsistent with its own entries. Minimum safe precision: `DECIMAL(19,4)`.

**SCHEMA-02 — `000902` vs queries: `is_leaf` generated column vs `is_leaf_account` in queries**
```sql
-- 000902 schema defines:
is_leaf BOOLEAN GENERATED ALWAYS AS (NOT has_children) STORED

-- CreateAccount query (finance_accounts.sql) inserts:
is_leaf_account  ← regular column, cannot insert into generated column
```
Generated columns cannot be the target of INSERT. Either the query always fails at runtime, or a missing migration renames/replaces the column. The trigger in 000909 also references `is_leaf_account` — consistent with the queries but inconsistent with 000902.

**SCHEMA-03 — `000909`: `update_account_hierarchy_flags` trigger references nonexistent column**
```sql
UPDATE finance_accounts SET is_leaf_account = ...
```
If `is_leaf` is the actual column name (per 000902), this UPDATE fails with column not found on every hierarchy modification. Inserting a child account = parent update fails = child insert rolls back. Hierarchy writes are broken.

**SCHEMA-04 — `000909`: Balance update trigger has concurrent posting race condition**
```sql
UPDATE finance_accounts
SET current_balance = current_balance + debit - credit
WHERE id = account_id
-- No SELECT FOR UPDATE, no advisory lock
```
Two concurrent postings to the same account both read `current_balance=1000`, both write `1000+delta`. One delta is lost silently. Balance is wrong, no error raised, no detection.

**SCHEMA-05 — `000909`: N+1 UPDATE loop in balance trigger**
```sql
FOR entry IN SELECT ... FROM finance_transaction_entries ... LOOP
    UPDATE finance_accounts SET ... WHERE id = entry.account_id;
END LOOP;
```
One UPDATE per journal entry line. A 20-line journal fires 20 sequential UPDATEs inside a trigger. Should be a single `UPDATE ... FROM (SELECT account_id, SUM(...) GROUP BY account_id)`.

**SCHEMA-06 — `000921`: `finance_reversal_history` has no FK constraints**
```sql
original_transaction_id UUID NOT NULL  -- no REFERENCES finance_transactions(id)
reversal_transaction_id UUID NOT NULL  -- no REFERENCES finance_transactions(id)
reversed_by UUID                       -- no REFERENCES users(id)
```
Soft-deleted transactions leave orphaned reversal history silently. Reversal history is unreliable for audit.

**SCHEMA-07 — `000921`: Empty string allowed as reversal reason**
```sql
reason TEXT NOT NULL DEFAULT ''
```
`NOT NULL DEFAULT ''` means reason is always satisfiable with an empty string. Audit trails with blank reversal reasons are meaningless. Fix: `CHECK (length(trim(reason)) > 0)`.

**SCHEMA-08 — `000904`: `transaction_type` CHECK excludes 'REVERSAL'**
```sql
CHECK (transaction_type IN ('MANUAL','SYSTEM','IMPORTED','RECURRING','ADJUSTMENT','CLOSING'))
```
`ReverseTransaction` creates a new transaction — presumably typed 'REVERSAL' — which fails this CHECK. Reversals either fail at insert or use wrong type like 'ADJUSTMENT', making reversal transactions indistinguishable from adjustments in reports.

**SCHEMA-09 — `000924`: `UNIQUE(tenant_id, id)` on accounting periods is redundant**
`id` is already the primary key (globally unique UUID). This unique constraint adds index overhead with zero correctness benefit.

**SCHEMA-10 — `000901`/`000902`: `created_by` nullable on audit-critical tables**
Both `finance_account_groups` and `finance_accounts` allow `created_by IS NULL`. Financial audit trail cannot identify who created an account. Unacceptable for a GL module.

**SCHEMA-11 — `000902`: `entity_id` nullable — accounts can exist without entity**
In multi-entity tenants, entity-filtered reports silently exclude these orphan accounts. Entity-level P&L is wrong.

**SCHEMA-12 — `000905`: `project_id UUID` has no FK reference**
Dangling UUID. Projects can be deleted while entries still reference them — dimensional reporting breaks silently.

**SCHEMA-13 — `000911`: `finance_accounting_periods` and `finance_currencies` missing `admin_role` RLS policy**
RLS enabled but only `readonly_role` SELECT policy exists. `admin_role` cannot INSERT, UPDATE, or DELETE periods or currencies. Period close/reopen operations fail for admin users.

**SCHEMA-14 — `000906`: Financial statement view reads denormalized `current_balance`**
`v_financial_statement_builder` reads `current_balance` from `finance_accounts`. If the balance update trigger failed (race from SCHEMA-04, or error), statements present stale/wrong balances silently. Should compute from `finance_transaction_entries` directly.

---

### QUERY LAYER — CRITICAL

**QUERY-01 — `finance_accounts.sql`: `entity_id` filter is dead code in 12+ queries**
```sql
-- Pattern in ListAccountsWithGroups, SearchAccountsWithGroupInfo, GetChartOfAccountsComplete,
-- GetAccountsByStatement, GetTrialBalanceData, GetAccountBalancesList,
-- GetLeafAccountsWithGroups, GetCashFlowAccountsList, GetAccountGroupSummary,
-- GetAccountsWithRecentActivity, GetStaleAccountBalances, GetAccountActivitySummary:
AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()  -- ← BUG: always true, entity_id never checked
)
-- Should be:
    OR entity_id = sqlc.narg('entity_id')::uuid
```
Copy-paste error replicated across a dozen queries. Entity filter does nothing. Multi-entity tenants get all entities' data regardless of filter. Entity-scoped P&L, trial balance, and cash flow reports are all broken.

**QUERY-02 — `finance_transaction_entries.sql:DeleteTransactionEntry`: hard DELETE**
```sql
DELETE FROM finance_transaction_entries WHERE id = ... AND tenant_id = ...
```
Every other delete uses soft delete (`deleted_at = NOW()`). This hard-deletes journal entry lines with no status check on the parent transaction. A POSTED transaction's entries can be permanently erased. No audit trail. Irreversible ledger corruption.

**QUERY-03 — `finance_transaction_entries.sql:DeleteTransactionEntries`: hard DELETE of all entries**
Same as QUERY-02 but deletes ALL entries for a transaction in one statement. Combined with no status guard, this can erase an entire posted journal with one call.

**QUERY-04 — `finance_accounts.sql:ValidateAccountHierarchy`: cycle detection is wrong**
```sql
WHERE parent_account_id = sqlc.narg('parent_account_id')
  AND id = sqlc.narg('parent_account_id')  -- self-reference only
```
Catches only direct self-reference. Does not detect ancestor cycles (A→B→C→A). A circular hierarchy is silently accepted, then the materialized path trigger loops or produces corrupt paths.

**QUERY-05 — `finance_accounts.sql:GetAccountSubtree`: LIKE-based subtree is unreliable**
```sql
h.full_path LIKE '%' || account_code || '%'
```
Account code "1000" matches "10001", "21000X", etc. Subtree queries return unrelated accounts. Hierarchy-based rolled-up balances are wrong.

**QUERY-06 — `finance_transactions.sql:PostTransaction`: DB allows DRAFT→POSTED**
```sql
AND transaction_status IN ('APPROVED', 'DRAFT')
```
DB-level enforcement allows posting from DRAFT, bypassing approval workflow. Root cause of application-layer CRIT-10 — the query itself undermines the approval model.

**QUERY-07 — `finance_transactions.sql:CreateTransactionWithDefaults`: hardcodes USD**
```sql
VALUES (..., 'USD', ...)  -- currency_code hardcoded
```
Multi-currency tenants silently get USD as default. No warning at compile or runtime.

**QUERY-08 — `finance_transactions.sql:GetTransactionActivity`: uses `created_at` not `transaction_date`**
```sql
WHERE created_at >= sqlc.arg('date_from') AND created_at <= sqlc.arg('date_to')
```
Filters by row creation time, not economic transaction date. Backdated entries appear in the wrong period's activity report.

**QUERY-09 — `finance_transaction_entries.sql:UpdateTransactionEntry`: no parent status check**
No check that parent transaction is DRAFT. POSTED transactions' entries can be silently modified after posting. Immutability of posted journals is a core accounting principle — this query violates it.

**QUERY-10 — `finance_transaction_entries.sql:MarkEntriesReconciled`: no POSTED check**
Entries from DRAFT or CANCELLED transactions can be marked reconciled. Reconciliation of unposted entries corrupts the reconciliation report.

**QUERY-11 — `finance_transaction_entries.sql:GetEntryTaxSummary`: semantically wrong taxable amount**
```sql
SUM(te.debit_amount + te.credit_amount) AS taxable_amount
```
Numerically happens to work (only one side non-zero per constraint) but is misleading and will break if the constraint is ever relaxed. Should be `SUM(GREATEST(debit_amount, credit_amount))`.

**QUERY-12 — Dead `entity_id` filter is systemic — affects ALL reporting query files**
The `OR tenant_id = current_tenant_id()` copy-paste bug extends beyond `finance_accounts.sql`. Same dead filter appears in:
- `finance_reporting_views.sql`: `GetFinancialStatementBuilder`, `GetBalanceSheetData`, `GetIncomeStatementData`, `GetFinancialStatementStructure`, `GetGroupBalanceSummary`, `GetTransactionSummary`, `GetUnreconciledTransactions`, `GetTransactionsByAccount`, `GetAccountUtilizationStats`, `GetTopAccountsByBalance`, `GetAccountsRequiringAttention`
- `finance_accounts_views.sql`: `GetAccountHierarchyComplete`, `CountAccountsWithGroups`, `GetAccountsByFinancialStatement`, `GetFinancialStatementData`, `GetAccountActivity`, `GetActiveAccounts`, `GetInactiveAccounts`, `GetHighActivityAccounts`

Every financial report in the system — balance sheet, income statement, trial balance, cash flow — is broken for multi-entity tenants. Entity isolation in reporting is completely non-functional. Total affected query count: **30+**.

**QUERY-13 — `finance_exchange_rates.sql:ListExchangeRates`: OFFSET hardcoded to 0**
```sql
LIMIT  $5
OFFSET 0   -- ← caller's offset parameter silently ignored
```
Pagination is broken. ListExchangeRates always starts from the first row regardless of offset passed by caller.

**QUERY-14 — `finance_approval_workflow.sql:GetPendingWorkflowsByUser`: unbounded fetch, wrong design**
Query returns ALL pending/in-progress workflows for the entire tenant — no LIMIT, no user filter in SQL. Comment acknowledges: "caller filters by assigned user in application layer". Under load this fetches every pending workflow for every user's dashboard load. N workflows × M users = M full-table reads per page view.

**QUERY-15 — Mixed positional (`$N`) and named (`sqlc.arg`) params across query files**
`finance_approval_workflow.sql`, `finance_account_balances.sql`, `finance_account_validation_rules.sql` use raw `$1/$2...` positional params. All other finance files use `sqlc.arg()`/`sqlc.narg()`. Positional params generate unnamed fields in SQLC output — no way to tell what `$4` is without counting. Inconsistency also breaks uniform SQLC config and linting.

**QUERY-16 — `finance_reporting_views.sql:GetTransactionsByAccount`: LIKE on account codes**
```sql
AND ts.account_codes LIKE '%' || sqlc.arg('account_code') || '%'
```
Prefix collision: account "1000" matches transactions involving "10001", "21000", "31000X". Transaction lookup by account returns wrong transactions. Same structural bug as `GetAccountSubtree`.

**QUERY-17 — `finance_accounts_views.sql:GetAccountHierarchyComplete`: hierarchy silently truncated at depth 10**
```sql
WHERE ah.hierarchy_level < 10  -- Prevent infinite recursion
```
Legitimate hierarchies deeper than 10 levels are silently dropped — no error, no warning, partial results. Caller cannot detect truncation. This makes rolled-up balances wrong for deep chart-of-account structures.

**QUERY-18 — `finance_account_balances.sql:DeleteAccountBalance`: hard DELETE of period balance snapshots**
Period-end balance snapshots are audit evidence. Hard DELETE erases them permanently. These should be append-only or at minimum soft-deleted.

**QUERY-19 — `finance_account_validation_rules.sql:DeleteAccountValidationRule`: hard DELETE**
Validation rules are tenant configuration — hard DELETE erases the history of which rules were active. When audit asks "why was this posting allowed?", the answer may be gone.

**QUERY-20 — `finance_reversal_history.sql`: TOCTOU race on double-reversal guard**
Application calls `IsReversalTransaction` then `InsertReversalHistory` as two separate queries with no enclosing transaction or advisory lock. Concurrent reversal attempts both pass the `IsReversalTransaction` check, both proceed to insert. The `UNIQUE(tenant_id, original_transaction_id)` constraint catches the second insert and throws an unhandled error — correct outcome but wrong mechanism. A true atomic guard requires `INSERT ... ON CONFLICT DO NOTHING` plus a return value check, inside a single transaction.

---

### INFRASTRUCTURE — CRITICAL

**INFRA-01 — `metrics/metrics.go`: `meter` field commented out — nil interface panics**
```go
type OTelProvider struct {
    // meter metric.Meter  ← commented out
}
```
First call to any metric method invokes a nil interface. Server crashes. No metrics emitted, entire service down.

**INFRA-02 — `metrics/metrics.go`: non-deterministic Prometheus label keys**
```go
for key := range labels {  // map iteration — random order
    labelKeys = append(labelKeys, key)
}
```
Prometheus requires consistent label key order across registrations. Random order causes "inconsistent label cardinality" panics on second call. Auto-registered metrics are unreliable.

**INFRA-03 — `tracing/tracing.go`: financial traces dropped when `SamplingRate < 1.0`**
`TraceIDRatioBased` sampler. Postings, reversals, and balance updates may have no trace at all. For financial operations, sampling must be 1.0 or use a custom sampler that always samples finance-tagged spans.

**INFRA-04 — `tracing/tracing.go`: `Insecure: true` in `DefaultConfig()`**
TLS disabled by default for OTLP exporter. Environments using `DefaultConfig()` without override send traces unencrypted. Compliance violation for financial data.

**INFRA-05 — `logger/logger.go`: `DefaultConfig()` sets `Development: true`**
Production deployments that don't explicitly override config get console format and debug-level logging. Wrong format for log aggregators, excessive noise, potential log injection via console format.

**INFRA-06 — `logger/logger.go`: `Fatal` accessible from business logic**
`Fatal` calls `os.Exit(1)`. If finance service validation calls `logger.Fatal`, the entire process exits — no graceful shutdown, no connection drain, no metric flush. Fatal must not be callable from domain/service code.

**INFRA-07 — `cache/cache.go`: `globalMemoryCache` has no tenant prefix**
Plain key-value store shared across all tenants. Two tenants using the same cache key (e.g., `"accounts:list"`) share data. Tenant isolation is broken for memory cache fallback.

**INFRA-08 — `cache/cache.go`: `MemoryCacheMaxSize` declared but not enforced**
```go
const MemoryCacheMaxSize = 1000
// Set() does: c.items[key] = entry  ← no size check
```
Cache grows without bound. Under load, OOM. The constant is decoration.

**INFRA-09 — `cache/cache.go`: circuit breaker opens with no fallback**
Threshold: 5 errors / 30s timeout. When Redis is unhealthy, circuit opens and cache calls return errors. Finance service has no fallback — cache errors propagate as operation failures rather than triggering DB reads.

**INFRA-10 — `cache/cache.go`: `getTenantInfo()` cache key diverges from RLS context**
Falls back to slug or subdomain if UUID absent. Cache hit keyed by slug returns data without RLS `app.tenant_id` context being set. Cached tenant data may bypass RLS boundaries.

---

### OBSERVABILITY GAPS

- Finance service never calls `tracing.DBAttributes()` — DB spans have no table/operation context
- No `span.SetStatus(codes.Error, ...)` on error paths — traces show success even on failure
- No financial-specific metric names or histogram buckets — all use auto-generated help text
- `updateAccountBalances` error silently dropped — no metric, no trace event, no log
- Partial reversal state has no alerting — corrupted books produce no observable signal

---

### SUMMARY TABLE

| Severity | Count | Primary Areas |
|----------|-------|---------------|
| CRITICAL (app layer) | 10 | nil panic, stale return, ledger corruption, atomicity, nil deref |
| CRITICAL (schema) | 14 | precision loss, broken trigger, race condition, missing FK, RLS gap |
| CRITICAL (queries) | 20 | dead entity filter (30+ queries), hard delete, TOCTOU race, pagination broken, unbounded fetch |
| CRITICAL (infra) | 10 | nil meter, random labels, tenant data leak, unbounded cache |
| Design/Financial | 5 | tolerance, approval bypass, stale balance, FK-less dimensions |

**Fix order:**
1. CRIT-02 — nil repo panic (service won't start safely)
2. QUERY-02/03 — hard DELETE of journal entries (irreversible data loss)
3. QUERY-01/12 — dead entity_id filter across 30+ queries (every financial report broken for multi-entity)
4. SCHEMA-01 — DECIMAL(15,2) precision (silent rounding on every posting)
5. INFRA-01 — nil meter crash (metrics bring down service)
6. SCHEMA-04 — balance trigger race (concurrent posting corrupts balances)
7. CRIT-07/08 + QUERY-20 — reversal atomicity (corrupted books on failure)
8. QUERY-09 — UpdateTransactionEntry on POSTED transactions (immutability violation)
9. SCHEMA-02/03 — `is_leaf` column naming split (hierarchy writes broken)
10. SCHEMA-06 + QUERY-18/19 — missing FKs and hard deletes destroying audit trail


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

