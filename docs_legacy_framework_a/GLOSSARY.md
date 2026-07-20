> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Glossary
---

# Glossary

Technical terms and acronyms used throughout AwoERP documentation.

## A

**ABAC** — Attribute-Based Access Control. Not used in AwoERP; we use RBAC + EntityScope.

**Activity** (Temporal) — A function that performs I/O or side effects within a Temporal workflow. Always idempotent. Executed by workers.

**Audit Log** — Append-only record of who changed what resource, when, and from what state. Stored in `audit_log` table.

## B

**BusinessError** — Typed error struct carrying an HTTP status code and client-facing message. Used to communicate domain errors through the layer stack.

## C

**Casbin** — Authorization library used for RBAC policy enforcement. Model: `(sub, dom, obj, act, eft)`.

**CRUD** — Create, Read, Update, Delete. Standard data operations.

## D

**Dead-Letter Queue (DLQ)** — Storage for events/messages that have exhausted all retry attempts. Inspectable and replayable via `awoctl`.

**Domain Error** — A named error representing a specific business rule violation (e.g., `ErrContractNotEditable`). Defined in the domain layer.

**Domain Event** — An immutable record of something that happened in the system (e.g., `ContractSubmitted`). Implements the `Event` interface with `Topic()` and `GetTenantID()`.

## E

**EntityScope** — Restricts data visibility to a specific entity, a subtree, or all entities. Set per user, enforced in the service layer.

**Expand-Contract** — Zero-downtime migration pattern. First add new column/table (expand), deploy new code, backfill, then remove old column/table (contract).

**Event Bus** — Publish/subscribe infrastructure. In-process in dev/test; Redis Streams in production.

**Event Outbox** — Postgres table that buffers events within a transaction. Ensures at-least-once event delivery. Relay goroutine drains it to the event bus.

## F

**Fiber** — Go HTTP framework used for the API server (`github.com/gofiber/fiber/v2`).

**Feature Flag** — Boolean toggle enabling/disabling a module feature per tenant. Stored in `feature_flags` and `tenant_feature_overrides` tables.

## G

**GUC** — Grand Unified Configuration. PostgreSQL session-level parameter. AwoERP uses `app.tenant_id` GUC for RLS enforcement.

## I

**IAM** — Identity and Access Management. The module handling users, roles, permissions, and sessions.

## M

**MDG** — Module Development Guide. The 23-step guide in Portal 4 describing how to build a new AwoERP module.

**Migration** — SQL file that changes the database schema. Managed by `golang-migrate`. Append-only; never modify applied migrations.

**Module** — A self-contained business domain package (`internal/core/{module}/`). Has its own domain, repo, service, handler, and Wire set.

**Multi-tenancy** — Architecture where multiple tenant organizations share one database instance, isolated by RLS.

## O

**Optimistic Locking** — Concurrency control using a `version` integer. Update only succeeds if `WHERE version = @version`. Returns 409 on mismatch.

**Outbox Pattern** — See *Event Outbox*.

## P

**pgx** — PostgreSQL driver for Go (`github.com/jackc/pgx/v5`). Used for all database access.

**Principal** — The identity of a user in the authorization system. Created from `ResolvedSession` via `sess.ToPrincipal()`.

## R

**RBAC** — Role-Based Access Control. Permissions assigned to roles; roles assigned to users.

**ResolvedSession** — Pre-computed session struct stored in Redis. Contains user ID, tenant ID, roles, entity scope, feature flags, and settings. Never has `Can()` or `HasRole()` methods — those are on `AuthzService`.

**RLS** — Row-Level Security. PostgreSQL feature enforcing tenant isolation at the database level using the `app.tenant_id` GUC.

## S

**Session Token** — UUID stored in Redis. Passed in `Authorization: Bearer {token}` header. Expires after 24h sliding TTL.

**Shopspring Decimal** — Go library for precise monetary arithmetic (`github.com/shopspring/decimal`). Used instead of `float64` for all monetary values.

**Soft Delete** — Marking a record as deleted via `deleted_at` timestamp without removing the row. Standard pattern for all business data.

**SQLC** — SQL compiler that generates type-safe Go from SQL queries. Queries live in `db/queries/`; generated code in `internal/shared/db/sqlc/`.

## T

**Task Queue** (Temporal) — Named queue workers poll for workflow and activity tasks. Format: `awoerp.{module}`.

**Temporal** — Durable workflow orchestration platform. Used for long-running, multi-step processes that must survive restarts.

**Tenant** — An organization using AwoERP. Has its own users, data, configuration, and feature flags.

## U

**UUID** — Universally Unique Identifier. All primary keys in AwoERP are UUIDs (v4, `gen_random_uuid()`). Never SERIAL integers.

## W

**Wire** — Google's compile-time dependency injection tool. Generates `wire_gen.go` from provider functions. Never edit `wire_gen.go` manually.

**Workflow** (Temporal) — Durable, deterministic function that orchestrates activities. Survives process restarts. Must not perform I/O directly.

**Worker** (Temporal) — Process polling a task queue and executing workflows and activities.

## Z

**Zero-UUID** — `00000000-0000-0000-0000-000000000000`. Must never be used as a tenant ID in production — RLS would match all rows without a tenant policy.
