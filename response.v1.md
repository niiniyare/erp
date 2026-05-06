# Finance Module — Remediation Status

> Date: 2026-04-29 (updated 2026-04-29)
> Baseline: `docs/reference/modules/financial/review.md`

---

## Summary

All CRITICAL application-layer bugs, schema, query, and infrastructure issues are fixed. All critical paths now have unit tests. Remaining work is non-critical design debt.

---

## Application Layer Fixes

| ID | Issue | Status | File |
|----|-------|--------|------|
| CRIT-01 | Wrong arg count in `NewTransactionService` | **FIXED** | `service/service.go` |
| CRIT-02 | `TransactionEntryService.repo` always nil | **FIXED** | `service/service.go` — wired `deps.TransactionRepo` |
| CRIT-03 | `CreateTransaction` drops all entries | **FIXED** | `service/transaction_service.go` — entries persisted after header |
| CRIT-04 | `CreateTransaction` never sets `TenantID` | **FIXED** | `service/transaction_service.go` — `shared.GetTenantID(ctx)` |
| CRIT-05 | Uniqueness check returns `(nil, nil)` on duplicate | **FIXED** | `service/transaction_service.go` — explicit error return |
| CRIT-06 | Balance update error silently swallowed | **FIXED** | `service/transaction_service.go` — error returned and propagated |
| CRIT-07 | Reversal marks original reversed before post | **FIXED** | `service/transaction_service.go` — correct order: create→post→mark |
| CRIT-08 | Reversal has no wrapping DB transaction | **FIXED** — `domain.TxRunner` interface + atomic path in `ReverseTransaction`; best-effort fallback when `TxRunner` nil |
| CRIT-09 | `ListTransactions` nil pointer dereference | **FIXED** | `service/transaction_service.go` — default before span |
| CRIT-10 | Inline post bypasses approval | **FIXED** | `service/transaction_service.go` — approval switch in `postTransactionInline` |
| BUG-01 | `UpdateEntry` returns stale pre-update state | **FIXED** | `service/transaction_entry_service.go` — re-fetch after update |
| BUG-02 | `DeleteEntry` allows deleting from POSTED txn | **FIXED** | `service/transaction_entry_service.go` — parent status check |
| BUG-05 | `ReverseTransaction` returns draft reversal | **FIXED** | `service/transaction_service.go` — returns posted reversal |
| DESIGN-05 | Missing `ErrTransactionCancelled` sentinel | **FIXED** | `domain/errors.go` |
| MISMATCH-02 | `TransactionTypeReversal` undefined | **FIXED** | `domain/types.go` |
| RISK-05 | `GetNormalBalanceForRootType` unused duplicate | **FIXED** | Removed duplicate in `types.go`; canonical impl in `accounts.go` |
| RISK-02 | Approval hardcoded false | **FIXED** | `service/transaction_service.go` — `RequiresApproval()` wired |

---

## Schema Fixes

| ID | Issue | Status | File |
|----|-------|--------|------|
| SCHEMA-01 | `DECIMAL(15,2)` precision loss | **FIXED** | `000905` — changed to `DECIMAL(19,4)` |
| SCHEMA-04+05 | Balance trigger race + N+1 loop | **FIXED** | `000909` — batched UPDATE + SELECT FOR UPDATE |
| SCHEMA-06 | Reversal history missing FK constraints | **FIXED** | `000921` |
| SCHEMA-07 | Empty string allowed as reversal reason | **FIXED** | `000921` — added CHECK constraint |
| SCHEMA-08 | `REVERSAL` type missing from CHECK | **FIXED** | `000904` |
| SCHEMA-09 | Redundant `UNIQUE(tenant_id, id)` on periods | **FIXED** | `000924` |
| SCHEMA-02+03 | `is_leaf` vs `is_leaf_account` column split | **NOT AN ISSUE** — 000907 adds `is_leaf_account` as regular column via ALTER TABLE |
| SCHEMA-10 | `created_by` nullable on audit tables | **FIXED** — `000901`/`000902` `created_by NOT NULL`; `create_standard_account_groups` updated with `p_created_by` param |
| SCHEMA-11 | `entity_id` nullable on accounts | **NOT AN ISSUE** — deliberate design for global/multi-entity accounts |
| SCHEMA-12 | `project_id` has no FK | **NOT AN ISSUE** — intentionally deferred (migration comment: "add FK when projects table exists") |
| SCHEMA-13 | Missing `admin_role` RLS policy on periods/currencies | **FIXED** — `000924` |
| SCHEMA-14 | Statement view reads denormalized `current_balance` | **FIXED** — `000906` `v_financial_statement_builder` and `v_financial_statement_structure` now join live entry-computed balances; also fixed `DECIMAL(15,2)→DECIMAL(19,4)` on `current_balance`/`ytd_balance` in `000902` |

---

## Query Layer Fixes

| ID | Issue | Status | File |
|----|-------|--------|------|
| QUERY-01/12 | Dead `entity_id` filter across 30+ queries | **FIXED** | `finance_accounts.sql`, `finance_accounts_views.sql`, `finance_reporting_views.sql` |
| QUERY-02 | Hard DELETE `DeleteTransactionEntry` | **FIXED** | `finance_transaction_entries.sql` — soft-delete |
| QUERY-03 | Hard DELETE `DeleteTransactionEntries` | **FIXED** | `finance_transaction_entries.sql` — soft-delete |
| QUERY-06 | `PostTransaction` allows DRAFT→POSTED in DB | **FIXED** | `finance_transactions.sql` — `= 'APPROVED'` only |
| QUERY-09 | `UpdateTransactionEntry` no parent status check | **FIXED** | `finance_transaction_entries.sql` — JOIN with status guard |
| QUERY-13 | `ListExchangeRates` OFFSET hardcoded to 0 | **FIXED** | `finance_exchange_rates.sql` |
| QUERY-18 | Hard DELETE of balance snapshots | **FIXED** — `000907` added `deleted_at`+`deleted_by`; `finance_account_balances.sql` all queries updated |
| QUERY-19 | Hard DELETE of validation rules | **FIXED** | `finance_account_validation_rules.sql` — soft-delete via `is_active=FALSE` |
| QUERY-20 | TOCTOU race on double-reversal guard | **FIXED** | `finance_reversal_history.sql` — `INSERT ... ON CONFLICT DO NOTHING RETURNING inserted` |
| QUERY-04 | Cycle detection only catches self-reference | **FIXED** | `finance_accounts.sql` — recursive ancestor CTE |
| QUERY-05/16 | LIKE-based subtree/account matching prefix collision | **FIXED** | `finance_reporting_views.sql` — anchored prefix patterns |
| QUERY-07 | Hardcoded USD in `CreateTransactionWithDefaults` | **FIXED** | `finance_transactions.sql` — `sqlc.arg('currency_code')` |
| QUERY-08 | `GetTransactionActivity` filters by `created_at` not `transaction_date` | **FIXED** | `finance_transactions.sql` |
| QUERY-10 | `MarkEntriesReconciled` no POSTED check | **FIXED** | `finance_transaction_entries.sql` — JOIN + status guard |
| QUERY-11 | `GetEntryTaxSummary` wrong taxable amount | **FIXED** | `finance_transaction_entries.sql` — `GREATEST(debit, credit)` |
| QUERY-14 | `GetPendingWorkflowsByUser` unbounded fetch | **FIXED** | `finance_approval_workflow.sql` — user filter + LIMIT/OFFSET |
| QUERY-15 | Mixed positional/named params | **NOT AN ISSUE** — each file uses positional `$N` consistently; no intra-query mixing |
| QUERY-17 | Hierarchy truncated at depth 10 silently | **FIXED** | `finance_accounts_views.sql` — depth 20, `truncated` column added |

---

## Infrastructure Fixes

| ID | Issue | Status | File |
|----|-------|--------|------|
| INFRA-01 | OTel `meter` nil → nil interface panic | **FIXED** | `internal/shared/metrics/metrics.go` — `otel.GetMeterProvider().Meter(name)` |
| INFRA-02 | Non-deterministic Prometheus label order | **FIXED** | `internal/shared/metrics/metrics.go` — `sort.Strings(keys)` |
| INFRA-03 | Financial traces dropped at `SamplingRate < 1.0` | **NOT AN ISSUE** — `DefaultConfig()` already sets `SamplingRate: 1.0` |
| INFRA-04 | `Insecure: true` default on OTLP exporter | **FIXED** | `internal/shared/tracing/tracing.go` — default `false` |
| INFRA-05 | `DefaultConfig()` sets `Development: true` | **FIXED** | `internal/shared/logger/logger.go` — default `false`, format `"json"` |
| INFRA-06 | `Fatal` calls `os.Exit(1)` from business logic | **FIXED** | `internal/shared/logger/zerolog.go` — logs as ERROR with `fatal=true` field, no exit |
| INFRA-07 | `globalMemoryCache` no tenant prefix | **NOT AN ISSUE** — `buildMemoryKey` already prefixes with `idType-tenantID:key` |
| INFRA-08 | `MemoryCacheMaxSize` not enforced | **NOT AN ISSUE** — `SetMemory` already checks size limit with eviction |

---

## Tests Added

| Suite | File | Coverage |
|-------|------|----------|
| `TestTransactionServiceSuite` | `service/transaction_service_test.go` | CreateTransaction, PostTransaction (balanced/unbalanced/closed-period/inactive-account), ReverseTransaction (mirror entries, cannot-reverse-reversal, entry-failure partial state), ApproveTransaction (SOD violation, different approver), ListTransactions (nil limit, zero limit clamp), GetTransactionWithEntries |
| `TestEntryServiceSuite` | `service/transaction_entry_service_test.go` | UpdateEntry returns post-update state (BUG-01), DeleteEntry blocked on POSTED/REVERSED txn (BUG-02), DeleteEntry allowed on DRAFT, reconciled entry blocked, GetEntriesByTransactionID |
| `TestTransactionSuite` | `domain/transaction_test.go` | IsBalanced, CalculateTotals, CanBePosted, CanBeReversed, CanBeEdited, Validate (15+ cases), IsSystemGenerated, GetNetAmount, RejectionReason validity |
| `TestAccountServiceSuite` | `service/account_service_test.go` | Account CRUD, hierarchy, activation |
| `TestPeriodServiceSuite` | `service/period_service_test.go` | Period lifecycle |
| `TestBudgetServiceSuite` | `service/budget_service_test.go` | Budget operations |
| Domain suites | `domain/*_test.go` | AccountingPeriod, TransactionEntry, TransactionStatus, RootType, ValidationResult, BudgetLineItem |

All suites pass (`go test ./internal/core/finance/... -v`).

---

## Remaining Work (Non-Critical Design Debt)

| ID | Issue | Priority |
|----|-------|----------|
| ~~BUG-04~~ | ~~`ValidateAmountConsistency` uses currency-agnostic `0.01` tolerance~~ | **FIXED** — `CurrencyMinorUnits`/`CurrencyConversionTolerance` in `domain/currency.go`; `ValidateAmountConsistency` uses `0.5×10^(-dp)` per currency |
| ~~BUG-06~~ | ~~`GetTransactionByNumber` passes `nil` entity ID — cross-entity data leak~~ | **FIXED** — `shared.WithEntityID`/`GetEntityIDPtr` added to `shared/context.go`; service now passes `shared.GetEntityIDPtr(ctx)` |
| BUG-07 | `CreateEntryRequest.Validate` constructs dummy UUID object to reuse validation — fragile | Low |
| MISMATCH-04 | 7 of 10 services unreachable via public `Service` interface | Medium |
| DESIGN-02/03 | `UpdateTransaction` accepts full struct; allows mutation of non-DRAFT transactions | Medium |
| MISMATCH-06 | Naming: docs say `JournalEntry`/`JournalLine`, code uses `Transaction`/`TransactionEntry` | Low (cosmetic, codebase-wide rename) |
