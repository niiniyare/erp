# Awo Framework — Third-Order Governance Review (2027–2045)

All previous recommendations assumed implemented. These are problems that emerge only after years of production, at scale, with an ecosystem of contributors and modules.

---

## 1. Metadata Versioning

### The Unversioned Schema Problem

`EntityDefinition` has no version field. At 50,000 tenants and hundreds of contributors, schema evolution is the most dangerous daily activity. Currently there is no mechanism to answer: "which version of `contact` definition produced this DB record?"

**Module-level schema versioning** must be first-class:

```go
type EntityDefinition struct {
    Name          string
    SchemaVersion int          // increments with breaking field changes
    Module        ModuleRef    // {Name: "crm", Version: "2.1.0"}
    // ...
}
```

`SchemaVersion` is stamped into every record at write time (`schema_version` column on system entities, top-level key in JSONB for custom entities). When the compiler reads a record, it knows which field schema to apply. This enables **in-place schema migration** without a DB migration: old records get schema version 1 semantics; new records get schema version 2 semantics. The store layer applies the correct field mapping based on the stamped version.

Without this: a module ships `contact` v2 with a renamed field. Every existing record silently fails validation. You discover this in production, not at compile time.

### Multi-Version Runtime

During rolling deployments, two compiled schemas execute simultaneously. They must agree on what they will write to the DB, but need not agree on what they expect to read.

**Write compatibility rule**: V2 may only write values that V1 can read without error. If V2 adds a nullable column with default, V1 reads NULL (acceptable). If V2 adds NOT NULL without default, V1 cannot insert (startup failure for V1 via migration verification check).

**Read compatibility rule**: unknown fields are silently ignored. The compiled SELECT lists only fields the compiling version knows about. Extra DB columns are invisible. Required — the framework must enforce this at the SQL layer by always using explicit column lists, never `SELECT *`.

The migration fingerprint check (startup Phase 5) must become **bidirectional**: new version verifies it can read data written by old version, AND old version (simulated via compatibility test) can read data written by new version.

### Backward/Forward Compatibility Matrix

Every change to `EntityDefinition` must be classified:

| Change | Backward compatible | Forward compatible | Notes |
|---|---|---|---|
| Add optional field | Yes | Yes | Old code ignores it |
| Add required field | No | Yes (if DB nullable+default) | Requires migration first |
| Remove field | No | No | Never permitted post-v1.0 |
| Rename field | No | No | Never permitted — add new, deprecate old |
| Change field Kind | Breaking | Breaking | Requires new field name |
| Add edge | Yes | Yes | |
| Remove edge | No | No | Requires deprecation window |
| Add Op to policy | Yes | Yes | More permissive |
| Remove Op from policy | No | No | Less permissive — active users lose access |

The framework CLI must enforce this matrix as a pre-commit check: `awo compat-check --from=v1.2 --to=HEAD`.

### Schema Negotiation

API clients built against entity schema v1.0 must continue to work when the server advances to v1.3. Three mechanisms:

**1. Additive field tolerance**: responses include all fields the server knows. Clients that don't know about new fields ignore them. Standard JSON tolerance.

**2. Request schema pinning**: `Accept-Schema: contact/1.0` header. Server responds with only fields that existed in schema v1.0. This enables clients to pin to a known schema for years. The server maintains field availability maps per schema version:

```go
type FieldAvailability struct {
    IntroducedIn  int  // SchemaVersion when this field was added
    DeprecatedIn  int  // 0 = not deprecated
    RemovedIn     int  // 0 = not removed (should never happen for published fields)
}
```

**3. Schema discovery endpoint**: `GET /api/{entity}/schema[?version=1.0]` returns the field list for that schema version. Clients can introspect before constructing requests.

### Module Compatibility Matrix

Module A v2.0 requires `awo/def` at API surface X. Module B v1.0 requires `awo/def` at API surface Y (subset of X). Both compile into the same binary.

Go modules handle this correctly if `awo/def` is backward compatible (old code compiles against new library). The framework governance rule: `def/` is ALWAYS backward compatible within a major version. No method removal. No signature changes. Only additions.

The compatibility matrix that must be maintained:

```
awo/def v1.0 → supports modules compiled against def v1.0
awo/def v1.1 → supports modules compiled against def v1.0, v1.1
awo/def v1.2 → supports modules compiled against def v1.0, v1.1, v1.2
```

This is Go's standard backward compatibility guarantee, applied strictly. The framework must have CI tests that compile against the previous 3 minor versions' public APIs.

### Deprecation Strategy

For any exported symbol in `def/`:

```
1. Mark deprecated in release M: // Deprecated: use EventDef instead. Will be removed in v1.M+8
2. Emit compiler warning via go vet check (not build error)
3. Provide migration codegen: `awo migrate --from=WorkflowTriggers --to=EventDef`
4. Remove in release M+8 (2 LTS cycles = minimum 3 years)
5. Alias in awo/compat/ for additional 2 LTS cycles
6. Final removal from awo/compat/ in M+16
```

Total deprecation window: minimum 6 years before deletion. For regulated industries where upgrade cycles are 3-4 years, this is the minimum viable deprecation period.

---

## 2. Plugin & Module Ecosystem

### Compilation-Unit Model — Commit to It Explicitly

The framework must make an explicit, documented architectural decision before v1.0: **all modules compile into the binary**. No runtime plugin loading. No Go `plugin` package.

Rationale:
- Go's `plugin` package requires identical build flags, compiler version, and dependency versions between the host and plugin. In practice, this is unmaintainable at ecosystem scale.
- For regulated industries: runtime-loaded code cannot be audited as part of the binary. Binary attestation (reproducible builds, signed binaries) requires a single compilation unit.
- Security: a runtime-loaded plugin has full process access. There is no sandboxing mechanism in Go that prevents a malicious plugin from reading arbitrary memory.

This decision eliminates the "marketplace installs modules without rebuild" use case. Accept this tradeoff. Document it.

**Consequence**: module updates require rebuild and redeploy. This is the correct behavior for production ERP systems. A bank does not hot-patch production binaries.

### Module Manifest Format

Every module declares its identity and requirements:

```go
var Module = def.ModuleManifest{
    Name:         "crm",
    Version:      "2.1.0",
    Description:  "Customer relationship management",
    Author:       "Acme Corp",
    License:      "MIT",
    RequiresDef:  ">=1.2.0, <2.0.0",
    RequiresDB:   ">=15.0",
    Conflicts:    []string{"legacy-crm"},
    Entities:     []string{"crm_contact", "crm_opportunity", "crm_activity"},
    Provides:     []string{"crm.contact_api"},
    Requires:     []string{"platform.tenant"},
}
```

The `awo` CLI reads all `Module` declarations at build time and:
1. Verifies no module conflicts
2. Verifies all required capabilities are provided
3. Verifies version constraint satisfaction
4. Produces a module dependency graph
5. Fails the build if the graph has cycles or unresolved dependencies

**Checked at build time, not at runtime.** No runtime module resolution.

### Capability Tokens

Instead of entity-name-level dependencies, use capability tokens:

```go
Provides: []string{"crm.contact_crud", "crm.contact_search"}
Requires: []string{"platform.tenant", "iam.session"}
```

This allows Module B to declare it requires `crm.contact_crud` without knowing which module provides it. If CRM module is replaced by a third party that also provides `crm.contact_crud`, Module B continues to work.

Capability tokens are the module dependency contract, not entity names.

### Hook Sandboxing Within Compilation-Unit Model

True isolation is impossible in Go without separate processes. Viable mitigations:

**Panic containment**: every hook invocation wrapped in `defer recover()`. Hook panics become errors, not process crashes.

**Deadline enforcement**: every hook execution has a context with deadline. The runner cancels the context at deadline. Hooks that respect `ctx.Done()` stop. Hooks that don't: runner continues, marks hook as "timed out," logs it.

**Consequence**: a third-party hook that blocks indefinitely leaks a goroutine. The framework must expose `awo_goroutine_leaks_total{hook_name}` metric and alert when it exceeds threshold. Operational response: disable the offending hook and rebuild without that module.

### Digital Signatures

The `awo` CLI must support:

```bash
awo module verify      # verify all modules' source hashes against signed manifests
awo module sign        # maintainer: sign module manifest with private key
awo module trust       # add maintainer public key to trust store
```

Each module has an `awo.sig` file containing the Ed25519 signature of the module source hash. The `awo` CLI verifies before `go build`. Not optional for regulated industries. Supply-chain attacks via Go module proxy are a real threat.

---

## 3. Runtime Isolation

### Tenant-Level Resource Quotas

50,000 tenants in one process. A single tenant's expensive query can exhaust the DB connection pool and degrade all others.

**Per-tenant semaphore**: each tenant has configurable max-concurrent-requests quota. Default: 100. Exceeded: 429. Implementation: `sync.Semaphore` per tenant, stored in `sync.Map`. Created on-demand, GC'd when idle.

**Per-tenant query timeout**: separate from request timeout. Query exceeding per-tenant DB timeout is cancelled at pgx level (context cancellation → PostgreSQL cancels query). Tenant gets 504. Others unaffected.

**Goroutine leak detection**: framework tracks active goroutines per tenant at hook boundaries. Unbounded growth → alert + optional circuit-break for that tenant.

### Circuit Breaker Architecture

Four levels:

**Per dependency, global**: if PostgreSQL error rate exceeds 50% over 30s, open circuit for all tenants. Return 503. Protects DB from thundering-herd retry storms.

**Per dependency, per tenant**: if specific tenant's query error rate exceeds 80% over 60s, open circuit for that tenant only.

**Per entity, per Op**: if `contact.Create` error rate exceeds threshold, open circuit for that specific operation.

**Per third-party hook**: error rate > 30% or p99 latency > 5x normal → mark degraded. After 3 degraded windows → disable hook and alert.

State transitions: closed → open → half-open → closed. Per-process state (no shared state between instances — each instance makes its own circuit-breaking decision independently).

### Context Propagation Guarantees

Every operation crossing a logical boundary must carry:
1. Request ID (correlation)
2. Tenant ID (isolation)
3. Viewer context (authorization)
4. Deadline (resource control)
5. Trace context (W3C traceparent)

Typed context keys (not string keys):

```go
type ctxKey struct{ name string }

// Required values — panic if not set (programming error, not runtime error)
func RequestIDFromCtx(ctx context.Context) string {
    v, ok := ctx.Value(ctxKeyRequestID).(string)
    if !ok || v == "" {
        panic("request_id missing from context — framework bug, not user error")
    }
    return v
}
```

Missing context values are programming errors. Panics here are correct — surface during development, not production.

---

## 4. Distributed Architecture

### Region Tagging on EntityDefinition

```go
type EntityDefinition struct {
    // ...
    DataLocality def.DataLocality  // LocalRegion | GlobalPrimary | GeoRestricted
}
```

- `LocalRegion`: reads and writes always go to regional DB pool. Records never leave region.
- `GlobalPrimary`: reads can go to any region (replicated). Writes always go to primary region.
- `GeoRestricted`: additional constraints applied at pg driver level.

The pg driver maintains N regional connection pools. Tenant configuration maps tenant → primary region. Runtime selects pool based on entity data locality + tenant's primary region.

Not optional for GDPR, HIPAA, or Kenya's Data Protection Act. Must be in the architecture before v1.0.

### Cross-Region Consistency Model — Explicit

The framework must declare its consistency guarantees per operation type. The consistency model on EntityDefinition:

```go
type DataLocality struct {
    WriteRegion   RegionPolicy  // PrimaryOnly | NearestRegion | AnyRegion
    ReadRegion    RegionPolicy  // PrimaryOnly | NearestRegion | AnyRegion
    Consistency   Consistency   // Linearizable | Eventual | BoundedStaleness
    MaxStaleMs    int           // for BoundedStaleness: maximum allowed replica lag
}
```

**System entity writes (financial, inventory, IAM)**: single-region write, async replication. Consistency: eventual for reads from non-primary regions. Writes: linearizable within primary region.

### Event Ordering in Multi-Region

Monotonic sequence per partition: each entity (partition key = entity ID + tenant ID) has a monotonically increasing sequence number. Outbox processor assigns via PostgreSQL sequence. Consumer buffers out-of-order events and applies in sequence order.

`EventDef.Ordered bool`: if true, outbox guarantees ordered delivery for events on same entity ID. If false (default): at-least-once, unordered.

### Active-Active — Only for Non-Financial Entities

The framework must enforce this at compile time:

```
// Compiler validates: IsolationLevel == Serializable → WriteRegion == PrimaryOnly
// If violation: compilation failure with explicit error
```

Active-active deployments are fundamentally incompatible with ACID consistency for financial data. Eliminate the class of bugs where a developer declares an active-active financial entity expecting the framework to handle conflicts.

---

## 5. Transaction Model

### The Ambient Transaction Problem

When a `BeforeSave` hook calls `store.Create()` for a related entity, it must join the outer transaction. The hook must receive the exact context carrying the `pgx.Tx`. If any intermediate code calls `context.WithValue(ctx, ...)` and loses the `pgx.Tx`, the hook creates a new transaction on a pool connection — outside the outer transaction. Silent correctness bug.

**Fix**: TxStore passed explicitly to hooks, not retrieved from context:

```go
// Hook that needs store access declares it explicitly
type BeforeSaveHookWithStore interface {
    BeforeSave(ctx context.Context, rec def.MutableRecord, store driver.EntityStore) error
}

// Hook that doesn't need store access uses simpler signature
type BeforeSaveHook interface {
    BeforeSave(ctx context.Context, rec def.MutableRecord) error
}
```

The runner detects which interface the hook implements and calls appropriately. The `store` passed to `BeforeSaveHookWithStore` is always the transactional store. No ambient context magic.

### Savepoint Semantics

```go
type TxStore interface {
    driver.EntityStore
    Savepoint(ctx context.Context, name string) error
    RollbackToSavepoint(ctx context.Context, name string) error
    ReleaseSavepoint(ctx context.Context, name string) error
}
```

Third-party hooks can use savepoints. The framework's runner does not use savepoints implicitly.

### Idempotency Keys — Mandatory for Financial Entities

Every mutation on entities with `IsolationLevel: Serializable` or entities in the `finance` module must support idempotency keys.

**Request side**: client sends `Idempotency-Key: <uuid>`. Framework computes: `idem_key = SHA256(tenant_id + entity + op + client_uuid)`.

**Server side**:
1. Check Redis: `GET idem:{idem_key}` → if found, return cached response immediately.
2. If not found: execute operation.
3. Write idempotency key to `idempotency_keys` DB table inside same transaction.
4. Store response in Redis: `SET idem:{idem_key} {response_json} EX 86400` (24h TTL, cache).
5. Return response.

Redis is fast-path cache. DB is durable store. On Redis miss: check DB table before executing.

### Serialization Conflicts — Transparent Retry

With `IsolationLevel: Serializable`, PostgreSQL may return `40001 serialization_failure`. The framework absorbs this with exponential backoff (3 retries: 10ms, 20ms, 40ms). After max retries: return HTTP 409 with `Retry-After: 1`. Clients never see a raw `40001` error.

### Isolation Levels Per Entity

```go
var LedgerDef = def.EntityDefinition{
    Name:           "finance_ledger_entry",
    IsolationLevel: def.IsolationSerializable,
    // ...
}
```

SERIALIZABLE for financial/inventory. READ COMMITTED for everything else. The pg driver sets the isolation level from `CompiledEntity.IsolationLevel` before beginning each transaction.

---

## 6. Caching Architecture

### Record Caching Is Wrong By Default

Do not cache individual entity records in Redis.

**Security**: cache key must include tenant_id + viewer policy hash. Key space explodes. Hit rate collapses. The cache becomes expensive infrastructure with no benefit.

**Correctness**: invalidation requires knowing all cache keys that contain a record. In distributed system with 50k tenants and multiple app servers, this is the distributed cache invalidation problem. Solving it correctly requires a cache invalidation bus more complex than a DB replica.

**Recommendation**: serve reads from DB primary or read replica. Use PgBouncer query-level statement caching for repeated identical queries. No application-level record caching.

**Exception**: global read-only reference tables (`currencies`, `countries`, `timezones`). Cache with 1-hour TTL and explicit admin-triggered invalidation.

### Permission Cache Architecture

Role-to-permission mapping cached with tag-based invalidation:

```
Redis key: perm:{tenant_id}:{role_name}   →  {entity: {op: bool}}
Tag:       perms:{tenant_id}              →  invalidate all role caches for this tenant
```

When admin changes any permission for tenant T: `DEL perms:{T}` tag invalidates all permission caches for that tenant.

Session validation hit path: Redis session lookup (1 round trip) → roles from session → permission cache lookup (1 round trip). Two Redis round trips total per request.

### Metadata Cache Hierarchy

**Tier 1 — In-process, immutable** (zero latency):
- `CompiledSchema` — permanent after startup Phase 4
- Route table — Fiber's internal trie

**Tier 2 — Redis, shared across instances** (1-2ms):
- SDUI base trees (keyed by `sdui:{entity}:{fingerprint}`)
- Feature flag evaluations
- Tenant slug → UUID mapping
- Role → permission mapping

**Tier 3 — Database, source of truth** (5-50ms):
- Everything else

### Stampede Prevention at 100k Concurrent

**Probabilistic Early Revalidation (PER)**:

```
P(early refresh) = (current_time - (expiry - beta × log(random())) > 0)
```

Where `beta` controls aggressiveness (typically 1-5). Stochastically distributes cache refresh across TTL period. At 100k concurrent reads, some fraction triggers refresh early. By the time key expires, the replacement is already in cache. No stampede.

This is the correct algorithm for high-concurrency caches. The framework's cache helper must implement PER as the default strategy for SDUI and feature flag caches.

---

## 7. Event Architecture

### CloudEvents Adoption — Non-Negotiable

The framework must adopt CloudEvents spec (v1.0) as its integration event envelope format. Do not invent a custom format.

Every integration event emitted by the outbox must be a valid CloudEvent:

```json
{
  "specversion": "1.0",
  "id": "uuid-v4",
  "source": "/tenants/{tenant_id}/entities/{entity_type}",
  "type": "awo.{entity_type}.{event_name}",
  "datacontenttype": "application/json",
  "dataschema": "https://schema.{your-domain}/events/{entity_type}.{event_name}/v{N}",
  "time": "2027-03-15T10:30:00Z",
  "awoschemaversion": "2",
  "aworequestedid": "{originating_request_id}",
  "data": { ... }
}
```

Reasons:
- CloudEvents is the CNCF standard. Every message queue, every event processing framework, every monitoring tool built in the next 20 years will speak CloudEvents.
- Schema Registry integrations (Confluent, AWS Glue) understand CloudEvents.
- The spec is designed for 20-year stability.

### Event Versioning Model

Versioning strategy: **event type URI versioning**:

```
type: "awo.contact.created.v1"  →  payload schema v1
type: "awo.contact.created.v2"  →  payload schema v2 (breaking change from v1)
```

Both versions emitted simultaneously during migration window (dual-write). Consumers migrate to v2 during window. After window: v1 emission stops.

The framework's `EventDef` declares current version and automatically version-stamps emitted events:

```go
type EventDef struct {
    Name    string  // "contact.created"
    Version int     // bumped on breaking payload change
    // ...
}
```

### Event Sourcing — Explicitly Out of Scope

Awo is a state-store system with integration events. It is NOT an event-sourced system. This must be stated explicitly in v1.0 documentation and enforced architecturally:

- No `EventStore` type in the framework
- No entity state reconstruction from events
- No event replay mechanism for rebuilding entity state

Any move toward event sourcing requires a v2.

### Ordering Guarantee via Monotonic Sequence

```sql
CREATE TABLE event_log (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           UUID NOT NULL,
    entity_type         TEXT NOT NULL,
    entity_id           UUID NOT NULL,
    event_type          TEXT NOT NULL,
    sequence_in_entity  BIGINT NOT NULL,  -- per entity_id sequence
    payload             JSONB NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL,
    schema_version      INT NOT NULL
);
```

`sequence_in_entity`: derived from a per-entity PostgreSQL sequence. Ordered consumers: order by `sequence_in_entity` within `(tenant_id, entity_id)`.

---

## 8. Security Beyond RBAC

### Attribute-Based Policies — Machine-Auditable

Current `PolicyFunc` (Go function) is powerful but not auditable. For regulated industries, a regulator must be able to inspect access control rules without executing code.

The framework must support two policy modes:

**Mode 1 — Programmatic** (current): `PolicyFunc`. Full Go code. Not auditable externally. For complex logic that cannot be expressed declaratively.

**Mode 2 — Declarative** (new): `PolicyExpression` using a restricted DSL:

```go
type PolicyExpression struct {
    When   string   // "viewer.hasRole('cardiologist') AND record.department == viewer.department"
    Action PolicyVerdict
}
```

Expression language requirements: deterministic, side-effect free, bounded (no loops, no recursion), machine-analyzable.

The `expr-lang/expr` package (already in go.mod) satisfies these requirements. Compiled at startup Phase 7 into compiled evaluators.

**Which one to use**: `PolicyExpression` for access control rules that must be auditable. `PolicyFunc` for complex business logic.

### Field-Level Encryption

```go
type FieldDef struct {
    // ...
    Encrypted      bool
    EncryptionKeyID string
}
```

The pg driver handles encryption/decryption transparently:
- At write: fetch current key from `driver.KeyManagement.GetCurrentKey(keyID)`, encrypt, store ciphertext + key version ID
- At read: read ciphertext + key version ID, fetch key by version, decrypt

```go
type KeyManagement interface {
    GetCurrentKey(keyID string) (KeyMaterial, error)
    GetKey(keyID, version string) (KeyMaterial, error)
    RotateKey(ctx context.Context, keyID string) (newVersion string, error)
}
```

**Key rotation**: old key remains in KMS for decrypt-only. New records use new key. Background job re-encrypts old records in batches. CLI command: `awo key rotate --field=contact.ssn --from=v1 --to=v2 --batch-size=1000`.

### Audit Log Tamper Evidence

Hash-chained audit log:

```sql
CREATE TABLE audit_log (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   UUID NOT NULL,
    op          TEXT NOT NULL,
    actor_id    TEXT NOT NULL,
    before_hash BYTEA,
    after_hash  BYTEA,
    prev_hash   BYTEA,      -- hash of previous audit_log row
    row_hash    BYTEA NOT NULL,  -- SHA-256(all above fields)
    created_at  TIMESTAMPTZ NOT NULL
);
```

`row_hash` chains each record to the previous via `prev_hash`. The chain head's `row_hash` can be published externally (cold storage, signed periodically) for independent verification.

The framework's `driver.AuditSink` interface enables SIEM integration. Default implementation: hash-chained PostgreSQL. Production: external WORM store.

### Supply-Chain Security

Three attack vectors:

**1. Malicious module in go.mod**: mitigation: `go mod vendor` + `go.sum` verification + reproducible builds + SLSA Level 2 compliance.

**2. Malicious driver implementation**: interface design minimizes exposed data. Drivers receive only what they need for the query. Cannot prevent network calls from a compromised driver — trust model: drivers are trusted as much as the code that selects them.

**3. Policy injection via filter DSL**: the compiler validates all filter predicates at startup against compiled field schema. The pg driver uses parameterized queries exclusively. Filter predicates produce parameterized SQL, never raw string SQL.

---

## 9. Performance Engineering

### Startup Parallelization — Critical at 10k+ Entities

Current sequential compilation: O(N × compile_cost). At 100k entities with 100μs per entity: 10 seconds. Too slow.

**Parallel compilation via topological sort**:
1. Compute dependency graph (O(V+E))
2. Find entities with no dependencies (leaves)
3. Compile all leaves in parallel goroutines (GOMAXPROCS parallel)
4. As each finishes: check if dependents' deps all done → add to parallel queue
5. Repeat until all compiled

Reduces from O(N) to O(depth × compile_cost). Typical graph depth 4-6: 600ms for 100k entities. Acceptable.

**Phase 5 (schema verification) batching**:

```sql
SELECT table_name, column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_schema = 'public'
ORDER BY table_name, ordinal_position;
```

One query, one result set. Build column map in-memory. Cross-reference against compiled schema. O(1) lookup per field. One DB round trip regardless of entity count.

**SDUI warm-up parallelization**: batch Redis writes using pipelining. 1000 SDUI trees per pipeline batch. 100k entities / 1000 per pipeline = 100 batches × 1-2ms = 100-200ms total. Acceptable.

### Lock Contention Analysis

**Hot path**: `CompiledSchema` map lookup: read-only, no lock, zero contention. Session Redis lookup: network I/O, no in-process lock. Per-tenant semaphore: `sync.Mutex` held for microseconds, contention minimal at 2 concurrent requests per tenant average.

**The one real contention risk**: per-tenant feature flag cache invalidation. Bulk operations changing flags for many tenants simultaneously → 50k Redis DEL operations competing for network bandwidth. Solution: rate-limit flag invalidations via an invalidation queue (Redis stream), processed at bounded rate.

### Memory Fragmentation

Long-lived processes with heavy allocation accumulate fragmentation.

**Monitoring**: `go_memstats_heap_fragmentation_ratio = 1 - (heap_alloc / heap_sys)`. Alert at 40%. At 60%: schedule graceful restart during low-traffic period.

The framework must expose this as a Prometheus metric. Primary source of fragmentation: `map[string]any` for record field storage. After eliminating this in favor of typed record access (already recommended), fragmentation should be significantly lower.

---

## 10. Observability Architecture

### Metrics Taxonomy — Stable Label Contracts

**Never use tenant_id as a label**. At 50k tenants × 100k entities × 4 ops: 20B time series. Prometheus dies.

Use:
```
tenant_tier = "enterprise" | "growth" | "starter"  (cardinality: 3)
entity_module = "finance" | "crm" | "hr"            (cardinality: ~20)
entity_name = "invoice" | "contact" | ...           (cardinality: 100-500)
op = "create" | "read" | "update" | "delete"        (cardinality: 4-10)
status = "2xx" | "4xx" | "5xx"                     (cardinality: 3)
```

High-cardinality per-tenant metrics: write to a separate time-series store (ClickHouse, TimescaleDB) via sidecar. Not Prometheus. Prometheus is for operational (low-cardinality) metrics.

Framework metric namespace (stable, versioned):

```
awo_http_request_duration_seconds{entity, op, status, tenant_tier}
awo_http_request_total{entity, op, status}
awo_store_query_duration_seconds{entity, op, driver}
awo_store_transaction_duration_seconds{entity, driver}
awo_store_connection_pool_size{driver, state}
awo_hook_duration_seconds{entity, timing, hook_name, result}
awo_hook_total{entity, timing, hook_name, result}
awo_policy_evaluation_duration_seconds{entity, op, result}
awo_cache_operation_total{cache, key_type, result}
awo_outbox_events_total{entity, event_type}
awo_outbox_lag_seconds{entity, event_type}
awo_outbox_delivery_attempts_total{entity, event_type, result}
awo_startup_phase_duration_seconds{phase}
awo_compiled_entities_total
awo_compiled_routes_total
```

### Trace Sampling Strategy

At 100k requests/second: 100% sampling is impossible.

**Adaptive sampling**:
- 100% sample: errors (4xx, 5xx), slow requests (p99 > 500ms), financial operations, first 10 requests per new tenant
- 1% sample: successful reads
- 10% sample: successful writes

Decision at root span (HTTP entry). Child spans inherit sampling decision. Outbox events inherit from originating request.

Head-based sampling with 100% error override (retroactively mark trace as sampled if request ends with error — requires tracing backend to buffer spans briefly).

### Structured Logging — Stable Field Schema

Every framework log entry has these core fields (never change):

```json
{
  "ts":          "2027-03-15T10:30:00.123Z",
  "level":       "info|warn|error",
  "msg":         "...",
  "request_id":  "uuid",
  "tenant_id":   "uuid",
  "actor_id":    "uuid|anon",
  "entity":      "contact",
  "op":          "create",
  "duration_ms": 45,
  "error":       "...",
  "error_code":  "...",
  "trace_id":    "hex32",
  "span_id":     "hex16"
}
```

Log field names are as stable as API field names. Renaming is a breaking change for every log consumer (Splunk queries, CloudWatch Insights, Datadog dashboards).

---

## 11. Framework Governance

### Interface Extension Without Breaking Changes

Go's structural typing: adding a method to an interface breaks every implementation.

**The permanent solution**: optional interface segregation via type assertion:

```go
// def/viewer.go — frozen at v1.0, never changes
type ViewerContext interface {
    ActorID()   string
    TenantID()  string
    OrgUnitID() uuid.UUID
    Roles()     []string
    IsSystem()  bool
}

// New capabilities go into separate optional interfaces
type OrgPathViewer interface {
    OrgPath() []uuid.UUID
}

type AttributeViewer interface {
    Attributes() map[string]string
}
```

The runtime checks for optional interfaces:
```go
if v, ok := viewer.(def.OrgPathViewer); ok {
    path = v.OrgPath()
} else {
    path = []uuid.UUID{viewer.OrgUnitID()}  // fallback
}
```

**Governance rule**: no new method may ever be added to `ViewerContext`. All new viewer capabilities go into separate optional interfaces.

This pattern applies to ALL interfaces in `def/` and `driver/`.

### RFC Process

Every change to `def/`, `filter/`, `driver/`, `registry/` requires:

**RFC mandatory sections**:
1. Problem statement: what breaks or is missing today
2. Proposed change: exact Go API changes (diff format)
3. Backward compatibility analysis: what existing code breaks, why
4. Migration path: automated codemod available? Manual steps?
5. Alternatives considered: at least 2, with explicit rejection rationale
6. Open questions

**RFC lifecycle**:
- `draft` → `review` (30-day comment for breaking; 14 days for additive)
- `review` → `accepted` (3 architecture team approvals, no unresolved blocking objections)
- `accepted` → `implemented` (merged to `experimental/` first)
- `implemented` → `stable` (2 minor releases without issues)

### Experimental APIs

Package `awo/x/` contains experimental APIs. No stability guarantees. Subject to change in any minor release. Graduated to stable in next major or minor (if additive).

### LTS Release Policy

Cadence:
- Major (vN.0.0): every 3-4 years. Breaking changes allowed. Migration tooling required.
- Minor (vN.M.0): every 6 months. Additive only. 12-month support window.
- LTS (vN.M.0-lts): every 18 months. 4-year support window. Critical fixes only.

Security patches: backported to ALL active LTS versions. Critical bugs: backported to latest 2 LTS versions.

A bank on v1.0 LTS (2027) must receive security patches until 2031. This requires dedicated maintainer capacity. The framework must account for this cost before committing to the LTS policy.

---

## 12. Testing Architecture

### Framework Conformance Suite

Any driver implementation must pass the conformance suite. Required tests for `driver.EntityStore` certification:

1. `Create` returns persisted record with generated ID
2. `Get` by ID returns created record
3. `Get` non-existent ID returns `ErrNotFound`
4. `Update` changes fields and increments version
5. `Update` with stale version returns `ErrVersionConflict` (optimistic lock)
6. `Delete` removes record; subsequent `Get` returns `ErrNotFound`
7. `List` with filter returns matching records only
8. `List` with pagination returns correct page
9. Tenant isolation: records created by tenant A are not visible to tenant B
10. `WithTx` commit: changes visible after commit
11. `WithTx` rollback: changes not visible after rollback
12. Concurrent `Create` with unique field: one succeeds, one returns `ErrDuplicate`
13. Serializable isolation: concurrent conflicting updates detected
14. Filter DSL operators: `Eq`, `Neq`, `Lt`, `Lte`, `Gt`, `Gte`, `In`, `NotIn`, `Contains`, `IsNull`, `IsNotNull`
15. `BulkCreate` is atomic: all succeed or none
16. AfterCommit hooks execute after transaction commits
17. Sensitive fields not returned in normal read responses
18. Unknown field in incoming record is ignored, not rejected
19. RLS policy prevents cross-tenant data access
20. `set_tenant_context` called before every tenant-scoped operation

**Certification**: third-party driver submits conformance test results via `go test -run Conformance ./...` to framework registry. Results published. Operators choose certified drivers.

### Golden Metadata Tests

For each entity in `platform/`, expected `CompiledEntity` output stored as JSON fixtures. Any change to the compiler output that changes a golden file requires explicit approval. Prevents accidental route changes, SQL template changes, or hook ordering changes from shipping undetected.

### Upgrade Testing — Mandatory

Every minor release must include:
1. Upgrade test: start with previous release's data, upgrade to current, verify all entities readable/writable
2. Migration fingerprint test: current compiled schema vs previous release's DB schema = expected compatibility
3. API backward compatibility test: requests valid in previous release are still valid in current

Run in CI against real PostgreSQL with seed data from previous release. Not optional. Release that fails upgrade tests is not releasable.

### Chaos Testing Strategy

Required chaos tests (minimum set):

- Redis SIGKILL during request: sessions fail, feature flags use defaults, SDUI generates on-demand. No data corruption.
- PostgreSQL connection killed mid-transaction: rollback. No partial writes. AfterCommit hooks do not fire.
- Temporal unavailable during AfterCommit: outbox event written. No error to client. Workflow not started. Outbox processor delivers when recovered.
- Hook panics: request continues. Error logged. Other hooks still execute.
- Disk full on outbox: outbox write fails. Entity write must also fail (atomicity: both in same transaction).
- Clock skew between app server and Redis: session expiry checked against Redis server time (EXPIREAT). No premature expiry.
- Rolling deploy with schema mismatch: V1 ignores unknown fields from V2. V2 handles NULL for new required fields. No 500s during deploy window.

Chaos suite runs weekly in staging. Not in every CI run.

### Deterministic Replay Testing

Hook chains, policy chains, validators must produce identical output for identical input regardless of execution environment.

Determinism test runner: replay same inputs 100 times with different goroutine scheduling (using `runtime.Gosched()` insertions). Verify identical output. Non-deterministic behavior detected. This is equivalent to Temporal's determinism enforcement applied to the hook/policy layer.

---

## 13. Operational Excellence

### Zero-Downtime Upgrade — The Hard Cases

**Case 1: Adding a new entity** — safe. Old binary ignores unknown table.

**Case 2: Adding a new required field** — safe IF migration adds nullable column with default first. Old binary doesn't include field in SELECT/INSERT; DB default applies on insert.

**Case 3: Changing an edge's cardinality** (OneToMany → ManyToMany) — **UNSAFE**. Requires maintenance window. Data written by old binary is in FK column; new binary reads from join table. Cannot be done during rolling deploy.

The framework CLI must classify migrations:

```
awo migration classify --from=v1.2 --to=v1.3

20270315143022_add_contact_notes.up.sql              → ZDT (add nullable column with default)
20270316091524_change_opportunity_cardinality.up.sql → MDW (cardinality change: join table backfill)
```

MDW migrations blocked during business hours without `--force-maintenance-window` flag.

### Canary Release Infrastructure

Tenant-sticky routing: load balancer routes each tenant to specific backend group based on `tenant_id` hash. New binary deployed to "canary" group (10% of backends). All requests from a given tenant during canary window go to same binary version.

Metrics comparison: error rate, latency, business metrics between canary and stable groups. Healthy after 24h → promote to full deployment. Regression detected → immediate rollback.

### Emergency Rollback

Time budget: under 5 minutes.

1. Tag current deployment as "bad" (30s)
2. Load balancer routes to previous version (instant — blue/green)
3. OR: deploy previous container image (2-3 min)
4. Verify health checks pass (1 min)

**Prerequisite**: previous version container image must be retained. Minimum: last 3 deployed images. Never delete "current production - 1" until 2 subsequent deploys succeed.

**DB rollback**: do NOT automatically rollback migrations. Previous code must tolerate current DB schema (additive migrations are backward compatible). If migration is not backward compatible: deployment order was wrong.

### Health Scoring

Beyond binary live/ready: a health score (0–100):

```json
{
  "score": 87,
  "components": {
    "database": {"score": 100, "latency_p99_ms": 3},
    "redis":     {"score": 95, "latency_p99_ms": 1},
    "temporal":  {"score": 75, "worker_count": 2, "note": "below target"},
    "outbox":    {"score": 80, "lag_seconds": 45}
  }
}
```

Load balancer deprioritizes instances with score below threshold.

---

## 14. Long-Term Sustainability

### Technical Debt Forecast

**2030 (3 years post-v1.0)**:
- Fiber coupling visible. Auth providers wanting gRPC/WebSocket face friction. `driver.AuthProvider` taking `*http.Request` doesn't cover all transport types.
- Filter DSL at 30+ operators. First driver compatibility gaps appear (CockroachDB driver supports 80% of operators).
- `EntityDefinition` at ~40 fields. New contributors take 1-2 hours to understand. Documentation is primary mitigation.
- amis driver falls behind amis versions (contrib — framework team doesn't own it).

**2035 (8 years post-v1.0)**:
- Vector search is mandatory for competitive ERP. `driver.SearchDriver` interface does not support semantic search. Either add `driver.SemanticSearch` interface (clean, no break) or extend SearchDriver (breaking). Correct choice: separate interface. But apps now implement two driver integrations.
- Go has evolved. Some Go 1.26 patterns look dated.
- Three LTS versions active simultaneously (v1.0, v1.4, v1.8). LTS maintenance consumes 1-2 FTE.
- Audit log hash-chain verification takes 2 hours for 100M records. Needs Merkle tree instead of linear chain.

**2040 (13 years post-v1.0)**:
- `EntityDefinition` at 60+ fields. Third-generation ERP capabilities (AI-native workflow, real-time collaboration, edge computing) cannot be naturally expressed as EntityDefinition fields. v2 seriously discussed.
- Compilation-unit model is hard ceiling on ecosystem growth. 10,000 modules → 5-minute build time. Developer experience unacceptable. Hot-reload for development becomes inevitable.
- PostgreSQL 18+ has native features making several framework patterns redundant. pg driver needs major updates.
- Outbox table for large tenants: 500M+ events, 156 monthly partitions. Partition management is nightly maintenance job.

**2045 (18 years post-v1.0)**:
- If 8 breaking-point mitigations from Section 15 held: no v2 needed. If even one was skipped at v1.0: v2 shipped around 2038.
- Filter DSL is now a 60-operator query language. Would have been better to adopt GraphQL or SQL subset from the beginning. Not possible now — too many clients.
- `def/` kernel itself: stable. Because frozen correctly at v1.0 with optional interface extensions, zero breaking changes accumulated.
- CloudEvents adoption: event envelope from 2027 still valid in 2045. Integration with 2045 event streaming systems works without modification.

### Abstractions to Freeze Forever

These must be declared permanently frozen at v1.0:

1. `def.ViewerContext` interface methods — `ActorID`, `TenantID`, `OrgUnitID`, `Roles`, `IsSystem`
2. `def.Op` string values — once published, string value is permanent
3. `def.HookTiming` constants and execution order
4. `def.PolicyVerdict` semantics — DENY wins, ALLOW requires explicit grant
5. Entity record ID type — `uuid.UUID`
6. HTTP response envelope — `{"data":..., "meta":...}` / `{"error":{...}}`
7. Filter DSL operator names — once operators are named, names are permanent
8. CloudEvents extension attribute names — once published, permanent
9. Hook timing contract — BeforeValidate is pre-TX, AfterSave is in-TX, AfterCommit is post-TX

### Extension Points Likely to Fail

**`driver.WorkflowEngine`**: interface assumes Temporal-like concepts (start workflow, send signal, query workflow). AWS Step Functions, Apache Airflow, and Temporal have fundamentally different models. Interface will either be too Temporal-specific or too abstract to be useful.

**`driver.SearchDriver`**: designed for keyword/full-text search. Vector search requires fundamentally different interface: `Search(queryVector []float32, topK int) ([]Record, error)`. By 2030, every serious ERP needs vector search for AI features. Interface must be extended or replaced.

**`driver.UIRenderer`**: neutrality of `UISchema` cannot survive 20 years of UI framework evolution. AR/VR, voice UI, AI-generated UI will require capabilities the intermediate representation cannot express.

### Over-Engineered Interfaces

**`driver.AuditSink`**: premature abstraction before two implementations exist. Concrete PostgreSQL hash-chained audit log covers 95% of deployments. Add interface when the second implementation is needed.

**`driver.KeyManagement`**: same reasoning. Add when first non-Vault implementation is needed.

**`SemanticHandler` in `def/`**: exposes driver concepts in kernel. Delete.

---

## 15. Future-Proofing — Survival to 2045

### Can Awo Survive Without a v2?

**Conditional yes, with specific design changes.** The eight conditions (each is a potential breaking point if skipped):

**Breaking point 1: `def.ViewerContext` interface evolution**

Without optional interface segregation: the first new viewer capability after v1.0 requires a v2. With optional interfaces: new capabilities added as separate interfaces, existing implementations continue to work.

Status: **fix required before v1.0.**

---

**Breaking point 2: Filter DSL wire format not stabilized**

Clients encode filters in HTTP query params. Once clients exist, encoding format is public API. Change in v1.1 breaks all clients.

Status: **must stabilize before v1.0. Produce formal grammar specification. Publish it. Frozen.**

---

**Breaking point 3: `PolicyFunc` missing `old Record` parameter**

Adding `old Record` post-v1.0 breaks every policy.

Status: **must include `old Record` (nil on Create) in v1.0.**

---

**Breaking point 4: `WorkflowTriggers` vs `EventDef`**

If `WorkflowTriggers []WorkflowTrigger` ships in v1.0, it's permanent. Temporal-specific concepts (TaskQueue) will be in the frozen API.

Status: **must replace with `EventDef` before v1.0.**

---

**Breaking point 5: No outbox at v1.0**

Every application that needs guaranteed delivery builds their own. Custom implementations diverge. When framework adds outbox in v1.1, migrating from custom implementations creates compatibility problems.

Status: **first-class outbox must be v1.0.**

---

**Breaking point 6: Module import path without version component**

If framework imports as `awo.so/def` (no `/v1/`), applications simultaneously importing old and new major versions get two different registries, two different compiled schemas, two different runtime instances.

Correct: `awo.so/v1/def` from the start. When v2 ships: `awo.so/v2/def`. Applications can import both and migrate gradually.

Status: **must decide before v1.0. Changing import paths post-v1.0 is breaking.**

---

**Breaking point 7: Custom event envelope**

If custom envelope ships in v1.0: migrating to CloudEvents in v1.2 is breaking for all external event consumers.

Status: **must adopt CloudEvents before v1.0.**

---

**Breaking point 8: `EntityDefinition.Name` not structurally immutable**

Temporal workflow IDs embed entity names. A rename (even accidental) breaks all in-flight workflows. The registry must structurally prevent it: first registration wins; no rename operation exists in the API; re-registration of an existing name panics.

Status: **must be enforced before v1.0.**

---

**Conclusion**: Awo CAN survive to 2045 without a v2 if and only if all 8 breaking points are addressed before v1.0 ships. If any one is skipped: a v2 becomes necessary before 2035.

---

## Architecture Risk Register

| Rank | Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| 1 | `def.ViewerContext` frozen incorrectly — blocks all auth evolution | Certain | Catastrophic (v2 required) | Optional interface segregation before v1.0 |
| 2 | Filter DSL wire format not stabilized — mass client breakage | High | High | Formal grammar spec before v1.0 |
| 3 | No transactional outbox at v1.0 — ecosystem fragmentation | High | High | First-class outbox in v1.0 |
| 4 | `WorkflowTriggers` frozen with Temporal-specific concepts | High | High (v2 for multi-engine) | Replace with EventDef before v1.0 |
| 5 | Custom event envelope — integration ecosystem fragmentation | Medium | High | Adopt CloudEvents before v1.0 |
| 6 | Goroutine leaks from third-party hooks — process instability | Medium | High | Hook timeout enforcement + leak detection |
| 7 | Tenant-level DB pool exhaustion — single tenant starves others | Medium | High | Per-tenant semaphore + circuit breaker |
| 8 | Schema fingerprint mismatch during rolling deploy | Medium | High | Bidirectional migration verification |
| 9 | `EntityDefinition` size growth — contributor attrition | Certain | Medium | Capability sub-structs before ecosystem growth |
| 10 | `driver.SearchDriver` insufficient for vector search by 2031 | High | Medium | Design with extension interface pattern |
| 11 | Audit log not tamper-evident — compliance failure | Medium | High | Hash-chained audit log before regulated deployments |
| 12 | LTS branch maintenance underestimated | Certain | Medium | Commit to LTS policy with explicit resource allocation |
| 13 | SDUI UISchema coupled to amis semantics | Medium | Medium | Define amis-agnostic UISchema before contrib/amis uses it |
| 14 | Supply-chain attack via malicious module | Low | Critical | Module signing enforcement in awo CLI |
| 15 | Module compatibility matrix not enforced | Medium | Medium | Compatibility verification in CLI before build |
| 16 | Stampede after Redis recovery — cascading DB overload | Low | High | PER algorithm + DB circuit breaker |
| 17 | Feature flag cache invalidation storm | Low | Medium | Rate-limited invalidation queue |
| 18 | Multi-region event ordering — incorrect state | Medium | High | Monotonic sequence per entity |
| 19 | Encryption key rotation not operationalized | High | High | Framework-provided key rotation CLI |
| 20 | Hook mutability — concurrent access to same record | Low | High | Framework guarantees single-threaded hook execution |

---

## Technical Debt Forecast Summary

| Year | Severity | Primary Sources |
|---|---|---|
| 2030 | Manageable | Fiber coupling, filter DSL growth, EntityDefinition field count |
| 2035 | Significant | Vector search gap, three active LTS branches, UI renderer mismatch |
| 2040 | Critical (if unaddressed) | EntityDefinition model exhaustion, compilation-unit ceiling at 10k modules |
| 2045 | Stable or v2 shipped | Depends entirely on whether 8 breaking points were addressed at v1.0 |

---

## Architectural Principles That Must Never Be Violated

1. **The kernel has zero external dependencies.** `def/` and `filter/` import nothing beyond `stdlib + uuid + decimal`. Any new kernel type requiring an external import is rejected.

2. **Interfaces in `def/` never have methods added after v1.0.** New capabilities go into new optional interfaces. Existing implementations are never broken.

3. **`EntityDefinition.Name` is a permanent identifier.** Registry rejects re-registration of any name. No rename API exists.

4. **The compilation phase is synchronous and completes before the first request.** `MustCompile()` either completes successfully or the process exits. Partial compilation is not a valid state.

5. **Policies are fail-closed.** No policy explicitly ALLOWs → DENY. Policy evaluation panics or times out → DENY. The framework never grants access due to an error condition.

6. **AfterCommit hooks never block the response.** Response sent after DB commit. AfterCommit runs in goroutine. Hook failures logged and queued. Clients never wait for AfterCommit.

7. **Tenant isolation is structural, not optional.** `set_tenant_context()` called on every DB connection before any tenant-scoped query. No code path executes tenant-scoped SQL without it.

8. **All integration events use CloudEvents v1.0+ envelope.** No custom format. No exceptions for "internal" events that eventually become external.

9. **Drivers receive only what they need.** Interface design minimizes exposed surface. No driver receives another tenant's data, the full CompiledSchema, or another component's credentials.

10. **Security decisions are never reversed by configuration.** Hash chaining on audit log: always on. RLS enforcement: always on. Session validation: always on. Cannot be disabled via configuration, feature flags, or environment variables.

---

## Final Production Readiness Score

Evaluated against: regulated industry deployment, 50k tenants, 100k concurrent, 20-year stability requirement.

| Dimension | Score | Rationale |
|---|---|---|
| Kernel stability | 72/100 | Correctly designed. Missing `old Record` in PolicyFunc, missing `/v1/` import path versioning. |
| Metadata compiler | 65/100 | Design correct. Not yet implemented. Largest gap. |
| Transaction model | 58/100 | Ambient TX correct. Idempotency keys missing. Serialization retry missing. |
| Event architecture | 40/100 | Outbox not implemented. CloudEvents not adopted. Event versioning not defined. |
| Plugin ecosystem | 50/100 | Compilation-unit model correct but undocumented. Module manifest missing. Conformance suite missing. |
| Security | 55/100 | RBAC exists. Field-level encryption missing. Audit tamper-evidence missing. ABAC missing. |
| Observability | 60/100 | Metrics taxonomy partially defined. OTel present. Log field schema not formalized. |
| Distributed architecture | 30/100 | No region tagging. No data locality declarations. Single-region assumption baked in. |
| Performance | 65/100 | Fiber model correct. Sequential startup compilation too slow at scale. Stampede prevention missing. |
| Developer experience | 55/100 | Entity creation simple. Policy debugging impossible. No conformance tests. No `awotest` helpers. |
| Governance | 35/100 | No RFC process. No LTS policy. No compatibility matrix. No deprecation lifecycle. |
| Testing | 45/100 | Unit test helpers exist. Conformance suite missing. Chaos tests missing. Golden metadata tests missing. |
| Operational excellence | 50/100 | Rolling deploy partially safe. No migration classification (ZDT vs MDW). No health scoring. |
| Future-proofing | 60/100 | Optional interface pattern proposed but not implemented. Filter DSL not stabilized. |

**Overall: 52/100**

Not production-ready for regulated industries. Acceptable for internal tooling and non-regulated SaaS. All gaps are known and addressable.

---

## Everything That Must Change Before v1.0

### Cannot Skip — Causes Inevitable v2

1. Optional interface extension pattern on all `def/` interfaces (ViewerContext frozen — new capabilities go into separate interfaces)
2. `/v1/` in module import path — `awo.so/v1/def`, `awo.so/v1/filter`, etc.
3. CloudEvents as the integration event envelope — any custom format shipped in v1.0 becomes permanent public API
4. Replace `WorkflowTriggers` on `EntityDefinition` with `EventDef`
5. `old Record` (nil on Create) in `PolicyFunc` signature — post-v1.0 addition breaks every policy
6. Filter DSL formal grammar specification — published, frozen, versioned before v1.0 ships

### Cannot Skip — Causes Compliance Failure

7. Transactional outbox as first-class framework primitive
8. Hash-chained audit log — write-once semantics for audit records
9. `driver.KeyManagement` interface — field-level encryption for PII/PHI
10. `DataLocality` on `EntityDefinition` — required for data residency compliance

### Cannot Skip — Causes Security Regression

11. Module signing verification in `awo` CLI
12. Per-tenant semaphore for request quota
13. Circuit breaker per dependency

### Cannot Skip — Causes Developer Ecosystem Failure

14. `driver/conformance/` test suite — third-party drivers have no certification path without it
15. `awotest` package with test helpers (NewViewer, NewRecord, NewMutableRecord, RunHook, AssertPolicy)
16. `awo debug policy` diagnostic command — policy debugging is impossible without it
17. Module manifest format and `awo module verify` CLI

### Cannot Skip — Causes Startup Failure at Scale

18. Parallel metadata compilation — sequential compilation at 10k+ entities exceeds acceptable startup time
19. Batch DB introspection in Phase 5 — one query for all columns, not one per table
20. PER (Probabilistic Early Revalidation) for cache — correctness under 100k concurrent load

### Should Change Before v1.0 (Technical Debt Prevention)

21. `EntityDefinition` capability sub-structs — establish pattern before struct reaches 60 fields
22. `awo/compat/` package skeleton — establishes deprecated symbol migration pattern
23. Formal deprecation lifecycle documentation
24. LTS release policy committed and published
25. `awo migration classify` command — ZDT vs MDW classification before any migration runs in production
26. Remove all flagged dead code: `framework/auth/session.go`, global `def.Register()`, `WorkflowTriggers`, zerolog in `framework/`, sequential `BulkCreate`
