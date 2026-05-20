# AWO ERP — Solo Developer Operations Guide
### Go · PostgreSQL · Temporal · Redis · AMIS-UI | East Africa Edition — Kenya Market Focus

---

> **Who this is for:** A single developer building and operating a multi-tenant ERP (AWO ERP) on the Go + PostgreSQL + Temporal + Redis + AMIS-UI stack, targeting the East African market (Kenya primary), where infrastructure budget, network latency, and operational bandwidth are real constraints — not theoretical ones.

---

## Table of Contents

- [Part I — Philosophy & Constraints](#part-i--philosophy--constraints)
  - [1.1 The Solo Developer Doctrine](#11-the-solo-developer-doctrine)
  - [1.2 What You Are NOT Building (and Why That Is Correct)](#12-what-you-are-not-building-and-why-that-is-correct)
  - [1.3 The East Africa Constraint Model](#13-the-east-africa-constraint-model)
- [Part II — System Architecture](#part-ii--system-architecture)
  - [2.1 Component Map](#21-component-map)
  - [2.2 HTTP Request Lifecycle](#22-http-request-lifecycle)
  - [2.3 Async Lifecycle (Temporal)](#23-async-lifecycle-temporal)
  - [2.4 Multi-Tenancy Boundary Model](#24-multi-tenancy-boundary-model)
  - [2.5 Wire DI Graph Overview](#25-wire-di-graph-overview)
  - [2.6 AMIS-UI Integration Model](#26-amis-ui-integration-model)
- [Part III — Hosting & Infrastructure](#part-iii--hosting--infrastructure)
  - [3.1 Hosting Decision Matrix (with Kenya Latency Reality)](#31-hosting-decision-matrix-with-kenya-latency-reality)
  - [3.2 Recommended: Single VPS Strategy](#32-recommended-single-vps-strategy)
  - [3.3 Containerisation Model](#33-containerisation-model)
  - [3.4 Reverse Proxy & TLS (Caddy)](#34-reverse-proxy--tls-caddy)
  - [3.5 Subdomain Tenant Routing](#35-subdomain-tenant-routing)
  - [3.6 Cost Model & Budget Breakdown (KES)](#36-cost-model--budget-breakdown-kes)
  - [3.7 Latency Optimisation for East Africa](#37-latency-optimisation-for-east-africa)
- [Part IV — Database Operations](#part-iv--database-operations)
  - [4.1 PostgreSQL Setup & Hardening](#41-postgresql-setup--hardening)
  - [4.2 Migration Strategy (golang-migrate in Production)](#42-migration-strategy-golang-migrate-in-production)
  - [4.3 RLS Architecture — Tenant Isolation](#43-rls-architecture--tenant-isolation)
  - [4.4 Connection Pooling (PgBouncer)](#44-connection-pooling-pgbouncer)
  - [4.5 SQLC Workflow — The Full Cycle](#45-sqlc-workflow--the-full-cycle)
  - [4.6 Backup Strategy (Non-Negotiable)](#46-backup-strategy-non-negotiable)
  - [4.7 Data Integrity Controls](#47-data-integrity-controls)
  - [4.8 Disaster Recovery Playbook](#48-disaster-recovery-playbook)
- [Part V — Deployment](#part-v--deployment)
  - [5.1 The Golden Rule: One-Command Deploy](#51-the-golden-rule-one-command-deploy)
  - [5.2 CI/CD Pipeline Design (GitHub Actions)](#52-cicd-pipeline-design-github-actions)
  - [5.3 Zero-Downtime Strategy on a Budget](#53-zero-downtime-strategy-on-a-budget)
  - [5.4 Codegen in CI (Wire + SQLC)](#54-codegen-in-ci-wire--sqlc)
  - [5.5 Environment Promotion](#55-environment-promotion)
  - [5.6 Rollback Runbook](#56-rollback-runbook)
- [Part VI — Observability (Lightweight)](#part-vi--observability-lightweight)
  - [6.1 The Lightweight Stack](#61-the-lightweight-stack)
  - [6.2 Structured Logging with Zerolog](#62-structured-logging-with-zerolog)
  - [6.3 Distributed Tracing (OpenTelemetry, Solo-Scale)](#63-distributed-tracing-opentelemetry-solo-scale)
  - [6.4 Temporal Workflow Visibility](#64-temporal-workflow-visibility)
  - [6.5 The Five Alerts That Actually Matter](#65-the-five-alerts-that-actually-matter)
  - [6.6 Dashboard Design for a Solo Operator](#66-dashboard-design-for-a-solo-operator)
- [Part VII — Security & Configuration](#part-vii--security--configuration)
  - [7.1 Secrets Management (No .env in Production)](#71-secrets-management-no-env-in-production)
  - [7.2 Config Hierarchy (Viper)](#72-config-hierarchy-viper)
  - [7.3 TLS, CORS, CSRF & Security Headers](#73-tls-cors-csrf--security-headers)
  - [7.4 API Key Lifecycle Management](#74-api-key-lifecycle-management)
  - [7.5 RLS as a Security Control](#75-rls-as-a-security-control)
- [Part VIII — Maintainability & Operational Fitness](#part-viii--maintainability--operational-fitness)
  - [8.1 The Solo Developer's Weekly Ops Ritual](#81-the-solo-developers-weekly-ops-ritual)
  - [8.2 Codebase Health Gates](#82-codebase-health-gates)
  - [8.3 Technical Debt Register — AWO's Known Open Gaps](#83-technical-debt-register--awos-known-open-gaps)
  - [8.4 Module Boundary Rules](#84-module-boundary-rules)
  - [8.5 Dependency Management & Upgrade Policy](#85-dependency-management--upgrade-policy)
- [Part IX — Solo Developer Runbooks](#part-ix--solo-developer-runbooks)
  - [RB-01 Application Won't Start](#rb-01-application-wont-start)
  - [RB-02 Database Connection Exhausted](#rb-02-database-connection-exhausted)
  - [RB-03 Tenant Data Bleed (RLS Failure)](#rb-03-tenant-data-bleed-rls-failure)
  - [RB-04 Migration Failed Mid-Deployment](#rb-04-migration-failed-mid-deployment)
  - [RB-05 Full Database Restore from Backup](#rb-05-full-database-restore-from-backup)
  - [RB-06 Temporal Worker Down](#rb-06-temporal-worker-down)
  - [RB-07 Redis Cache Corruption / Eviction Storm](#rb-07-redis-cache-corruption--eviction-storm)
  - [RB-08 Server Disk Full](#rb-08-server-disk-full)
  - [RB-09 Emergency Rollback (Production)](#rb-09-emergency-rollback-production)
  - [RB-10 New Tenant Onboarding Checklist](#rb-10-new-tenant-onboarding-checklist)
- [Part X — Evolution Path](#part-x--evolution-path)
  - [10.1 Stage 0 → Stage 1: Local to Live (First Tenant)](#101-stage-0--stage-1-local-to-live-first-tenant)
  - [10.2 Stage 1 → Stage 2: Hardening (First Paying Customer)](#102-stage-1--stage-2-hardening-first-paying-customer)
  - [10.3 Stage 2 → Stage 3: Growth (Team + Multi-Region)](#103-stage-2--stage-3-growth-team--multi-region)
  - [10.4 Temporal Activation Roadmap](#104-temporal-activation-roadmap)
  - [10.5 When to Stop Being Solo](#105-when-to-stop-being-solo)
- [Appendices](#appendices)
  - [A — Environment Variable Reference](#a--environment-variable-reference)
  - [B — Makefile Target Reference](#b--makefile-target-reference)
  - [C — Recommended Tooling Stack](#c--recommended-tooling-stack)
  - [D — East Africa Provider Directory](#d--east-africa-provider-directory)

---

## Part I — Philosophy & Constraints

### 1.1 The Solo Developer Doctrine

Building an ERP as a solo developer in the East African market is not an infrastructure challenge — it is a **discipline challenge**. Every service you add is a service you must monitor, patch, debug, and recover. Every abstraction layer you introduce is a layer that can fail at 2 AM in Nairobi when your first enterprise client's accountant is locked out before month-end close.

The goal of this guide is not "best infrastructure" as defined by AWS Well-Architected Framework for a 50-engineer org. It is:

| Principle | What It Means in Practice |
|---|---|
| **Minimum operational surface area** | Fewest moving parts that still serve real enterprise tenants reliably |
| **Maximum recoverability** | When something breaks, you restore full service in under 30 minutes, alone |
| **Predictable deployments** | Deploy at any time — including from a phone — and know exactly what will happen |
| **Cost-conscious from day one** | Every shilling of compute must earn its keep against KES-denominated revenue |

---

### 1.2 What You Are NOT Building (and Why That Is Correct)

> ⚠️ **Rule:** Do not build for 10,000 tenants on day one. Build for 10 tenants done excellently.
> Do not add a service because it appears on a CNCF landscape diagram. Add it because a real problem demands it.

| Do NOT Build | DO Build Instead |
|---|---|
| Multi-region active-active cluster | Single-region VPS with WAL-based backup to a second location |
| Kubernetes orchestration | Docker Compose with named volumes and a one-command restart |
| Distributed microservices mesh | Monolith with clean internal domain boundaries (what AWO already is) |
| Complex multi-stage CI/CD pipelines | One deploy script with a tested rollback path |
| ELK stack + full SIEM | Prometheus + Loki + one Grafana instance on the same VPS |
| Vault + Consul secrets cluster | systemd environment file + `age` encryption at rest |
| Multi-CDN global distribution | Cloudflare free tier (Nairobi PoP covers East Africa well) |
| Managed Kubernetes (GKE/EKS) | Hetzner or DigitalOcean VPS — 10x cheaper, easier to reason about |

---

### 1.3 The East Africa Constraint Model

Operating from Kenya introduces constraints that no AWS Solutions Architect whitepaper addresses directly.

| Constraint | Reality | Design Response |
|---|---|---|
| **Internet latency** | Nairobi → eu-west-1: ~180ms round-trip. Nairobi → af-south-1 (Cape Town): ~60ms | Host in Cape Town or Johannesburg, not Europe/US |
| **Power reliability** | Outages are periodic — KPLC load shedding is real | Redis persistence to disk; Postgres WAL; graceful shutdown handlers in all services |
| **Mobile-first users** | Most SME accountants use Android on mobile data | AMIS-UI must stay lean; API responses paginated; no large JSON blobs |
| **Payment infrastructure** | M-Pesa is the primary rail, not cards | Design billing module around STK Push / C2B callbacks; Safaricom API latency is a factor |
| **Bandwidth cost** | Business broadband is expensive; mobile data is metered | Gzip all API responses; avoid unnecessary polling; use WebSockets sparingly |
| **Talent density** | Junior devs available in Nairobi, Mombasa, Kisumu — but Go expertise is rare | Keep codebase readable; document patterns; avoid exotic dependencies |
| **Regulatory** | KRA compliance (VAT, ETR), EPRA (fuel), Banking Act for financial modules | Audit trails are not optional; the AWO audit package is a compliance tool |
| **FX risk** | USD-denominated hosting billed in KES at CBK rate | Budget hosting in USD; set a monthly hard limit; prefer predictable flat-rate pricing |

---

## Part II — System Architecture

### 2.1 Component Map

```
┌─────────────────────────────────────────────────────────┐
│                     CLIENT LAYER                        │
│  Browser (AMIS-UI / JSON Schema pages)                  │
│  Mobile (Android — M-Pesa STK Push callbacks)           │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTPS
┌──────────────────────▼──────────────────────────────────┐
│               CADDY (Reverse Proxy + TLS)               │
│  Subdomain routing: {tenant}.awo.so → Fiber             │
│  Auto-HTTPS via Let's Encrypt / ZeroSSL                 │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│            AWO ERP — Fiber v2 HTTP Server               │
│                                                         │
│  Middleware Stack (in order):                           │
│    TenantMiddleware → AuthMiddleware → Authorize()      │
│    → RouteSecurityManager → ObservabilityMiddleware     │
│                                                         │
│  Handler Domains:                                       │
│    /auth/*   /iam/*   /tenant/*                         │
│    /finance/*   /entity/*   /audit/*                    │
└───────┬──────────────────────┬──────────────────────────┘
        │                      │
┌───────▼───────┐    ┌─────────▼──────────┐
│  PostgreSQL   │    │   Temporal Server   │
│  (RLS-based   │    │   (Workflow engine) │
│  multi-tenant)│    │                     │
│               │    │  Temporal Worker    │
│  PgBouncer    │    │  (same binary,      │
│  (pooler)     │    │   separate cmd)     │
└───────┬───────┘    └─────────────────────┘
        │
┌───────▼───────┐    ┌─────────────────────┐
│     Redis     │    │  Prometheus          │
│  (sessions,   │    │  Loki (logs)         │
│   cache,      │    │  Grafana (dashboards)│
│   rate limit) │    │  Temporal UI         │
└───────────────┘    └─────────────────────┘
```

---

### 2.2 HTTP Request Lifecycle

Every authenticated, tenant-scoped request flows through exactly this chain. Understanding this prevents 80% of production bugs.

```
1. Caddy receives HTTPS request
   └── Extracts subdomain: acme.awo.so → tenant_slug = "acme"

2. TenantMiddleware
   └── Resolves tenant_slug → tenant_id (UUID) via Redis cache
   └── Injects tenant_id into Fiber context
   └── Sets PostgreSQL SET app.tenant_id = '<uuid>' for RLS

3. AuthMiddleware
   └── Validates JWT / session token
   └── Loads session.Permissions []string (no per-request DB lookup)
   └── Injects user_id, session into context

4. Authorize("module.resource.action")
   └── Checks permission string against session.Permissions
   └── Returns 403 if not present

5. Handler
   └── Reads tenantID via shared.GetTenantID(ctx)
   └── Calls Service layer

6. Service
   └── Calls Repository

7. Repository
   └── Calls store.WithTenant(ctx, tenantID, func(...))
   └── PostgreSQL RLS policy filters rows automatically

8. Response
   └── Zerolog structured log written
   └── OpenTelemetry span closed
   └── Prometheus counter incremented
```

---

### 2.3 Async Lifecycle (Temporal)

```
HTTP Handler
  └── Calls temporal.Client.ExecuteWorkflow(ctx, options, workflowFn, args)
      └── Returns WorkflowRun (ID + RunID) — handler returns 202 Accepted immediately

Temporal Server
  └── Queues workflow task

Temporal Worker (AWO worker binary, same repo)
  └── Polls task queue
  └── Executes WorkflowFn
      └── Calls Activity functions (e.g., SendEmail, RecordJournalEntry)
          └── Activities interact with DB via store.WithTenant()
          └── Activities interact with Redis cache
  └── Workflow completes or enters retry/compensation logic

Client polls
  └── GET /api/v1/workflows/{id}/status → WorkflowRun.Get(ctx)
```

> **Important AWO note:** As of current state, zero concrete workflows exist. The SDK is wired. Temporal is the correct tool for long-running processes like month-end close, bulk invoice generation, and M-Pesa reconciliation. Activate as features demand it.

---

### 2.4 Multi-Tenancy Boundary Model

AWO uses **PostgreSQL Row-Level Security** as the tenancy enforcement layer — not application-level filtering. This is a strong guarantee: a bug in the application layer cannot accidentally return another tenant's data as long as `WithTenant()` is always used.

```
Tenant Lifecycle:
  PENDING → ACTIVE → SUSPENDED → ACTIVE
      │          │
      └──────────┴──► ARCHIVED

Subdomain pattern:  {slug}.awo.so
Database pattern:   Every table has tenant_id UUID column
RLS policy:         WHERE tenant_id = current_setting('app.tenant_id')::uuid
Code pattern:       store.WithTenant(ctx, tenantID, func(ctx, s db.Store) error { ... })
```

**Never bypass this.** A raw SQLC query executed outside `WithTenant()` will silently return all rows across all tenants on development (RLS off) and explode in production (RLS on). This is a critical rule.

---

### 2.5 Wire DI Graph Overview

AWO uses Google Wire for **compile-time** dependency injection. There is no runtime DI container — `wire_gen.go` is generated Go code.

```
main.go
  └── InitializeApp() [wire_gen.go]
       ├── config.Load()
       ├── db.NewPool(config) → *pgxpool.Pool
       ├── db.NewStore(pool) → db.Store
       ├── cache.NewService(config) → cache.Service
       ├── tracing.NewService(config) → tracing.Service
       ├── metrics.NewProvider(config) → metrics.MetricsProvider
       ├── logger.NewLogger(config) → logger.Logger
       ├── temporal.NewClient(config) → client.Client
       │
       ├── iam.NewFacade(store, cache, logger, ...) → iam.Facade
       ├── tenant.NewFacade(store, cache, logger) → tenant.Facade
       ├── finance.NewService(store, logger, ...) → finance.Service
       ├── audit.NewService(store, logger) → audit.Service
       │
       └── handlers.NewDependencies(...) → *handlers.Dependencies
            └── fiber.App → registered routes
```

**Rules:**
- Never hand-edit `wire_gen.go` — run `make wire`
- Adding a new domain: write provider, add to wire set, run `make wire`
- Build tag `//go:build wireinject` is required on `wire.go`

---

### 2.6 AMIS-UI Integration Model

AMIS renders admin UI from JSON schema definitions. The schema lives in `web/schemas/pages/`. The API serves data; AMIS renders it.

```
web/schemas/pages/finance/accounts.json
  └── AMIS renders a full CRUD page
  └── Calls GET /api/v1/finance/accounts → Fiber handler → Service → DB
  └── POST /api/v1/finance/accounts on form submit

Deployment: Static JSON files served by Caddy or embedded in binary via go:embed
Evolution: Replace JSON pages with a custom React frontend when revenue justifies it
```

**East Africa consideration:** JSON schema pages are light — a finance dashboard schema is typically under 50KB. This is important for users on mobile data.

---

## Part III — Hosting & Infrastructure

### 3.1 Hosting Decision Matrix (with Kenya Latency Reality)

Latency from Nairobi to major data center regions:

| Region | Provider | Approx RTT from Nairobi | Notes |
|---|---|---|---|
| **af-south-1** (Cape Town) | AWS | ~55–70ms | Best AWS option for EA |
| **Johannesburg** | Hetzner | ~50–65ms | Best price-to-latency ratio |
| **eu-west-1** (Ireland) | AWS | ~170–200ms | Too far for interactive ERP |
| **eu-central-1** (Frankfurt) | AWS / Hetzner | ~150–180ms | Common mistake choice |
| **us-east-1** (N. Virginia) | AWS | ~220–280ms | Never use for primary |
| **Nairobi** | Azure | ~5–10ms | Azure has a Nairobi PoP — good for future |
| **Mombasa/Nairobi** | Safaricom Cloud | ~5–20ms | Emerging; limited product range |

> **Recommendation for AWO:** Host primary server in **Hetzner Falkenstein or Hetzner Johannesburg** when available, or **DigitalOcean NYC1 with Cloudflare in front** as a fallback. Cloudflare's Nairobi PoP (~10ms) terminates TLS locally even if origin is in Europe, dramatically improving perceived performance for AMIS-UI loads.

---

### 3.2 Recommended: Single VPS Strategy

**For Stage 0–2 (0–50 tenants), run everything on one VPS.**

```
┌─────────────────────────────────────────────────┐
│         Hetzner CX31 or DigitalOcean 4GB        │
│         2 vCPU / 8GB RAM / 80GB SSD             │
│         ~$20–25/month = ~KES 2,600–3,200        │
├─────────────────────────────────────────────────┤
│  Docker Compose Services:                        │
│    awo-api        (Fiber HTTP server)            │
│    awo-worker     (Temporal worker)              │
│    postgres:16    (primary DB)                   │
│    redis:7        (cache + sessions)             │
│    temporal       (workflow server)              │
│    caddy          (reverse proxy + TLS)          │
│    prometheus     (metrics scraper)              │
│    grafana        (dashboards)                   │
│    loki           (log aggregation)              │
│    pgbouncer      (connection pooler)            │
└─────────────────────────────────────────────────┘
```

**Why this is correct for solo dev:**
- One SSH session to debug everything
- One `docker compose logs -f` to see the whole system
- One backup target (the VPS disk + off-site pg_dump)
- One server to patch, not a fleet
- Vertical scaling (upgrade to CX41 = 4 vCPU/16GB) covers you through Stage 2

---

### 3.3 Containerisation Model

**Use Docker Compose. Not Swarm. Not K8s. Not Nomad.**

`docker-compose.yml` (production-safe structure):

```yaml
version: '3.9'

services:
  caddy:
    image: caddy:2-alpine
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config

  postgres:
    image: postgres:16-alpine
    restart: always
    environment:
      POSTGRES_DB: awo_production
      POSTGRES_USER: awo
      POSTGRES_PASSWORD_FILE: /run/secrets/pg_password
    volumes:
      - pg_data:/var/lib/postgresql/data
      - ./scripts/init-rls.sql:/docker-entrypoint-initdb.d/00-rls.sql:ro
    secrets: [pg_password]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U awo"]
      interval: 10s
      timeout: 5s
      retries: 5

  pgbouncer:
    image: pgbouncer/pgbouncer:latest
    restart: always
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DATABASES_HOST: postgres
      PGBOUNCER_POOL_MODE: transaction
      PGBOUNCER_MAX_CLIENT_CONN: 200
      PGBOUNCER_DEFAULT_POOL_SIZE: 20

  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --appendonly yes --maxmemory 512mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data

  awo-api:
    image: ghcr.io/your-org/awo-erp:${IMAGE_TAG:-latest}
    restart: always
    depends_on:
      postgres:
        condition: service_healthy
    env_file: /etc/awo/production.env
    ports:
      - "127.0.0.1:3000:3000"

  awo-worker:
    image: ghcr.io/your-org/awo-erp:${IMAGE_TAG:-latest}
    restart: always
    command: ["./awo", "worker"]
    env_file: /etc/awo/production.env
    depends_on:
      postgres:
        condition: service_healthy

  temporal:
    image: temporalio/auto-setup:1.24
    restart: always
    environment:
      DB: postgresql
      DB_PORT: 5432
      POSTGRES_USER: awo
      POSTGRES_PWD_FILE: /run/secrets/pg_password
      POSTGRES_SEEDS: postgres
    depends_on:
      postgres:
        condition: service_healthy
    secrets: [pg_password]

  prometheus:
    image: prom/prometheus:latest
    restart: always
    volumes:
      - ./observability/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus

  loki:
    image: grafana/loki:latest
    restart: always
    volumes:
      - loki_data:/loki

  grafana:
    image: grafana/grafana:latest
    restart: always
    environment:
      GF_SERVER_ROOT_URL: https://ops.awo.so
      GF_SECURITY_ADMIN_PASSWORD_FILE: /run/secrets/grafana_password
    volumes:
      - grafana_data:/var/lib/grafana
    secrets: [grafana_password]

volumes:
  pg_data:
  redis_data:
  caddy_data:
  caddy_config:
  prometheus_data:
  loki_data:
  grafana_data:

secrets:
  pg_password:
    file: /etc/awo/secrets/pg_password
  grafana_password:
    file: /etc/awo/secrets/grafana_password
```

---

### 3.4 Reverse Proxy & TLS (Caddy)

Caddy handles TLS automatically — no certbot, no cron jobs.

```Caddyfile
# All tenant subdomains → AWO API
*.awo.so {
  reverse_proxy awo-api:3000
  tls {
    on_demand
  }
  encode gzip
  log {
    output file /var/log/caddy/access.log
    format json
  }
}

# Ops dashboard (restricted by IP)
ops.awo.so {
  reverse_proxy grafana:3000
  @allowed remote_ip 41.139.0.0/16  # Your IP range
  respond @allowed 403
}

# Temporal UI (internal only)
temporal.awo.so {
  reverse_proxy temporal:8080
  basicauth {
    admin $2a$14$...   # bcrypt hash
  }
}
```

**On-demand TLS** means new tenant subdomains get certificates automatically on first request — no manual cert provisioning per tenant.

---

### 3.5 Subdomain Tenant Routing

```
acme.awo.so
  └── Caddy: wildcard *.awo.so → awo-api:3000
       └── TenantMiddleware reads Host header
            └── Extracts "acme" from "acme.awo.so"
                 └── Redis GET tenant:slug:acme → tenant_id UUID
                      └── If miss: DB lookup → cache for 5 minutes
                           └── Inject into context
```

**Known AWO gap:** TLD-aware parsing for `.co.ke`, `.com` multi-part TLDs is TODO. Until fixed, use `*.awo.so` only — do not allow custom domains like `acme.co.ke` pointing to AWO without implementing proper subdomain extraction first.

---

### 3.6 Cost Model & Budget Breakdown (KES)

All costs in USD with KES equivalent at ~130 KES/USD (verify current CBK rate).

#### Minimal Production Stack (~$30/month = ~KES 3,900)

| Component | Provider | USD/month | KES/month | Notes |
|---|---|---|---|---|
| VPS (4GB/2vCPU) | Hetzner CX21 | $6 | ~780 | Primary server |
| Backups (VPS snapshot) | Hetzner | $1.20 | ~156 | Automated daily snapshots |
| Off-site backup storage | Backblaze B2 | $0.60 | ~78 | 10GB pg_dump storage |
| Domain (.so TLD) | Namecheap | $2.50 | ~325 | amortized monthly |
| Cloudflare | Cloudflare | $0 | $0 | Free tier covers you to Stage 2 |
| **Total** | | **~$10** | **~KES 1,340** | Lean MVP |

#### Comfortable Production Stack (~$35/month = ~KES 4,550)

| Component | Provider | USD/month | KES/month | Notes |
|---|---|---|---|---|
| VPS (8GB/4vCPU) | Hetzner CX31 | $15 | ~1,950 | Handles 20–50 tenants comfortably |
| Managed PostgreSQL backup | Hetzner | $5 | ~650 | Or DIY to Backblaze |
| Backblaze B2 | Backblaze | $1 | ~130 | 50GB backup retention |
| Cloudflare Pro | Cloudflare | $20 | ~2,600 | WAF + DDoS — worth it for fintech |
| Domain | Namecheap | $2.50 | ~325 | |
| Uptime monitoring | BetterUptime | $0 | $0 | Free tier |
| **Total** | | **~$43** | **~KES 5,600** | Serious production |

#### When to Scale

| Signal | Action |
|---|---|
| CPU > 70% sustained | Upgrade VPS vertically (Hetzner CX41 = $25/mo) |
| DB connections > 80 (with PgBouncer) | Tune PgBouncer pool size first; then consider read replica |
| RAM > 85% | Add swap (4GB); then upgrade VPS |
| Response p95 > 500ms | Profile with pprof before throwing hardware at it |
| 50+ active tenants | Consider separating DB to dedicated VPS ($15/mo) |
| 100+ active tenants | Re-evaluate architecture; hire a second developer |

---

### 3.7 Latency Optimisation for East Africa

Network latency for a Nairobi-based user hitting a server in Johannesburg is ~60ms. For an ERP with many API calls per page load, this compounds fast.

**Strategies implemented at the infrastructure level:**

```
1. Cloudflare in front of everything
   └── TLS terminates in Nairobi Cloudflare PoP (~5ms to client)
   └── Cloudflare connects to origin over optimised backbone (~30ms vs ~60ms raw)

2. HTTP/2 multiplexing via Caddy
   └── Multiple API calls over one connection
   └── Critical for AMIS-UI which makes several requests on page load

3. Gzip compression on all API responses (Caddy + Fiber)
   └── A 200KB JSON response compresses to ~15KB on mobile data

4. Redis session cache
   └── Auth check = 1ms Redis read instead of DB roundtrip
   └── Tenant resolution = 1ms Redis read

5. Permission pre-loading at login
   └── session.Permissions []string loaded once at JWT issue
   └── Authorize() middleware = in-memory slice check, zero DB/Redis

6. PgBouncer transaction pooling
   └── Avoids PostgreSQL connection overhead per request
   └── Each request gets a pool connection, not a dedicated backend process

7. SQLC prepared statements
   └── Query plans cached at PostgreSQL level
   └── No query planning overhead on repeated calls
```

**Target latencies (from Nairobi):**

| Operation | Target p50 | Target p95 | If Exceeded |
|---|---|---|---|
| Login (auth + session) | < 150ms | < 300ms | Check Redis |
| List endpoint (paginated) | < 200ms | < 400ms | Check query plan |
| Create/Update mutation | < 250ms | < 500ms | Check transaction scope |
| AMIS page initial load | < 800ms | < 1.5s | Check Cloudflare cache |

---

## Part IV — Database Operations

### 4.1 PostgreSQL Setup & Hardening

**Initial server configuration** (add to `postgresql.conf`):

```ini
# Memory — tune to 25% of available RAM for shared_buffers
shared_buffers = 2GB              # For 8GB VPS
effective_cache_size = 6GB        # ~75% of RAM
work_mem = 16MB                   # Per sort/hash operation
maintenance_work_mem = 256MB      # For VACUUM, CREATE INDEX

# WAL (critical for backup integrity)
wal_level = replica               # Required for WAL archiving
archive_mode = on
archive_command = 'test ! -f /var/wal-archive/%f && cp %p /var/wal-archive/%f'
max_wal_senders = 3

# Connections (PgBouncer sits in front — keep this low)
max_connections = 50              # PgBouncer manages the rest

# Logging (for audit and debugging)
log_min_duration_statement = 200  # Log queries > 200ms
log_line_prefix = '%t [%p] [%a] [%d] '
log_checkpoints = on
log_lock_waits = on

# Autovacuum tuning (ERP tables get heavy writes)
autovacuum_vacuum_scale_factor = 0.05   # More aggressive than default 0.2
autovacuum_analyze_scale_factor = 0.02
```

**Create the application user with minimum privileges:**

```sql
-- Create app user (not superuser)
CREATE USER awo_app WITH PASSWORD 'strong-password-here';
GRANT CONNECT ON DATABASE awo_production TO awo_app;
GRANT USAGE ON SCHEMA public TO awo_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO awo_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO awo_app;

-- Allow RLS to work (app user cannot bypass it)
ALTER ROLE awo_app SET row_security = on;
```

---

### 4.2 Migration Strategy (golang-migrate in Production)

AWO uses `golang-migrate` with the naming format `001001_description.up.sql`.

**Migration rules — never violate:**

```
✅ DO:
  - Write migrations as additive (add columns, add tables, add indexes)
  - Always write the corresponding .down.sql
  - Test .down.sql — it should fully reverse the .up.sql
  - Make migrations idempotent where possible (CREATE IF NOT EXISTS)
  - One concern per migration file

❌ DO NOT:
  - DROP a column that the current binary reads (deploy binary first, then drop)
  - Rename a column without a two-step migration
  - Add a NOT NULL column without a DEFAULT in the same migration
  - Run migrations manually on production without checking migration status first
```

**Pre-deployment migration check:**

```bash
# Check current migration state
docker exec awo-api ./awo migrate status

# Dry-run: show pending migrations
docker exec awo-api ./awo migrate list --pending

# Apply migrations (part of deploy script, runs before binary starts)
docker exec awo-api ./awo migrate up
```

**Two-phase column rename pattern (safe for production):**

```sql
-- Migration 001: Add new column
ALTER TABLE users ADD COLUMN full_name TEXT;
UPDATE users SET full_name = first_name || ' ' || last_name;

-- Deploy new binary that reads both columns

-- Migration 002 (next sprint): Drop old columns
ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
```

---

### 4.3 RLS Architecture — Tenant Isolation

Every table that stores tenant data must follow this pattern:

```sql
-- 1. Add tenant_id column
ALTER TABLE invoices ADD COLUMN tenant_id UUID NOT NULL;

-- 2. Create index (critical for performance)
CREATE INDEX idx_invoices_tenant_id ON invoices(tenant_id);

-- 3. Enable RLS
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices FORCE ROW LEVEL SECURITY;

-- 4. Create policy
CREATE POLICY tenant_isolation ON invoices
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

-- 5. App user cannot bypass RLS (superuser can — use for emergency only)
ALTER TABLE invoices FORCE ROW LEVEL SECURITY;
```

**How `WithTenant` sets the context:**

```go
// internal/platform/db/store.go (conceptual)
func (s *store) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error {
    tx, err := s.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    _, err = tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID.String())
    if err != nil {
        return err
    }

    if err := fn(ctx, &txStore{tx: tx}); err != nil {
        return err
    }
    return tx.Commit(ctx)
}
```

**Verifying RLS works (run this in production after deploy):**

```sql
-- Connect as app user
SET app.tenant_id = 'tenant-uuid-A';
SELECT COUNT(*) FROM invoices; -- Should only return tenant A rows

SET app.tenant_id = 'tenant-uuid-B';
SELECT COUNT(*) FROM invoices; -- Should only return tenant B rows

-- Reset and check total (superuser only)
RESET app.tenant_id;
SELECT COUNT(*) FROM invoices; -- Returns all rows (superuser bypasses RLS)
```

---

### 4.4 Connection Pooling (PgBouncer)

Without a connection pooler, each Fiber goroutine that hits the DB opens a PostgreSQL backend process. At 100 concurrent requests, that is 100 Postgres processes — a guaranteed OOM on a small VPS.

**PgBouncer configuration (`pgbouncer.ini`):**

```ini
[databases]
awo_production = host=postgres port=5432 dbname=awo_production

[pgbouncer]
listen_port = 6432
listen_addr = 0.0.0.0
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt

# Transaction pooling is correct for AWO
# (does not preserve SET LOCAL across statements — WithTenant handles this)
pool_mode = transaction

max_client_conn = 200         # Max connections FROM the app
default_pool_size = 20        # Max simultaneous backend connections TO postgres
min_pool_size = 5
reserve_pool_size = 5
reserve_pool_timeout = 3

# Timeouts
server_connect_timeout = 5
server_idle_timeout = 600
client_idle_timeout = 0       # Never disconnect idle clients (ERP users leave tabs open)

log_connections = 0           # Too noisy in production; enable when debugging
log_disconnections = 0
```

> ⚠️ **Important:** `SET LOCAL app.tenant_id` inside a transaction works correctly with PgBouncer in transaction mode because `WithTenant` wraps the entire operation in a transaction. The `SET LOCAL` is scoped to that transaction and cleared when it commits/rolls back. This is by design.

---

### 4.5 SQLC Workflow — The Full Cycle

```
Write SQL → Generate Go → Write tests → Commit

1. Write query in db/queries/<domain>.sql
2. Run: make sqlc    (generates db/sqlc/<domain>.sql.go)
3. Write repository adapter in internal/core/<domain>/repository/<domain>_sqlc.go
4. Run: make mock    (generates test mocks)
5. Write unit tests using mocks
6. Run: make test
7. Commit generated files (db/sqlc/ and mocks/)

NEVER:
  - Hand-edit db/sqlc/
  - Commit without running make sqlc after a query change
  - Write raw SQL in handler or service layer
```

**SQLC query template (with RLS-compatible pattern):**

```sql
-- db/queries/finance.sql

-- name: ListAccountsByTenant :many
SELECT * FROM accounts
WHERE tenant_id = @tenant_id
ORDER BY created_at DESC
LIMIT @limit OFFSET @offset;

-- Note: @tenant_id is redundant when using WithTenant()
-- but is good documentation and a safety double-lock
```

---

### 4.6 Backup Strategy (Non-Negotiable)

Data loss is the one event from which you cannot recover reputation with enterprise clients. Treat backups as a product feature.

**Three-layer backup approach:**

```
Layer 1: Continuous WAL Archiving (Point-in-Time Recovery)
  └── PostgreSQL streams WAL segments to /var/wal-archive/
  └── Script syncs to Backblaze B2 every 5 minutes
  └── Recovery granularity: any point in time
  └── Cost: ~$0.50/month for 50GB on Backblaze B2

Layer 2: Daily pg_dump (Logical Backup)
  └── Runs at 01:00 EAT (11 PM UTC previous day)
  └── pg_dump --format=custom → compressed .pgdump file
  └── Uploaded to Backblaze B2 with 30-day retention
  └── Tested weekly via automated restore to staging

Layer 3: VPS Snapshot (Infrastructure Backup)
  └── Hetzner daily automated snapshots (whole-disk)
  └── Retained for 7 days
  └── Use only if server is unrecoverable — prefer DB-level restore
```

**Backup automation script (`/etc/cron.d/awo-backup`):**

```bash
#!/bin/bash
# /usr/local/bin/awo-backup.sh

set -euo pipefail

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="/tmp/awo_${TIMESTAMP}.pgdump"
B2_BUCKET="awo-backups"
LOG="/var/log/awo/backup.log"

echo "[$(date)] Starting backup..." >> "$LOG"

# Dump
PGPASSWORD="$DB_PASSWORD" pg_dump \
  --host=localhost \
  --port=6432 \
  --username=awo_app \
  --format=custom \
  --compress=9 \
  --file="$BACKUP_FILE" \
  awo_production

# Upload to Backblaze B2
b2 upload-file "$B2_BUCKET" "$BACKUP_FILE" "daily/awo_${TIMESTAMP}.pgdump"

# Cleanup local
rm -f "$BACKUP_FILE"

# Verify upload
b2 ls "$B2_BUCKET" daily/ | grep "$TIMESTAMP" >> "$LOG"

echo "[$(date)] Backup complete: awo_${TIMESTAMP}.pgdump" >> "$LOG"

# Alert on failure (send via Slack webhook or email)
```

```
# Cron schedule (EAT = UTC+3)
0 22 * * * root /usr/local/bin/awo-backup.sh   # 01:00 EAT
```

**Weekly backup verification ritual (every Monday):**

```bash
# Download latest backup
b2 download-file-by-name awo-backups daily/awo_<latest>.pgdump /tmp/restore_test.pgdump

# Restore to staging database
PGPASSWORD="$STAGING_DB_PASSWORD" pg_restore \
  --host=staging-postgres \
  --username=awo_app \
  --dbname=awo_staging \
  --clean \
  /tmp/restore_test.pgdump

# Verify row counts match production
psql -h staging-postgres -U awo_app awo_staging -c "
  SELECT
    schemaname,
    tablename,
    n_live_tup
  FROM pg_stat_user_tables
  ORDER BY n_live_tup DESC
  LIMIT 20;
"
```

---

### 4.7 Data Integrity Controls

**Database-level constraints (never rely on application layer alone):**

```sql
-- Foreign key constraints (enforce referential integrity)
ALTER TABLE invoices
  ADD CONSTRAINT fk_invoices_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenants(id)
  ON DELETE RESTRICT;  -- Never silently cascade deletes for financial data

-- Check constraints
ALTER TABLE journal_entries
  ADD CONSTRAINT chk_journal_balanced
  CHECK (debit_amount >= 0 AND credit_amount >= 0);

-- Not null on critical fields
ALTER TABLE transactions
  ALTER COLUMN amount SET NOT NULL,
  ALTER COLUMN currency SET NOT NULL,
  ALTER COLUMN tenant_id SET NOT NULL;

-- Unique constraints
ALTER TABLE tenants
  ADD CONSTRAINT uq_tenants_slug UNIQUE (slug);
```

**Soft deletes for financial data (never hard delete):**

```sql
-- All financial tables use deleted_at instead of DELETE
ALTER TABLE invoices ADD COLUMN deleted_at TIMESTAMPTZ;

-- Update RLS policy to exclude soft-deleted rows
CREATE POLICY tenant_isolation ON invoices
  USING (
    tenant_id = current_setting('app.tenant_id', true)::uuid
    AND deleted_at IS NULL
  );
```

**Audit trail (non-optional for compliance):**

Every mutation in AWO must call `audit.Service.Record()` after success. This is a KRA compliance requirement — tax authorities must be able to audit every financial record change.

```go
// In every handler that mutates financial data
_ = deps.AuditService.Record(c.UserContext(), audit.Event{
    Action:     "invoice.created",
    ActorID:    userID,
    TenantID:   tenantID,
    Resource:   "invoice",
    ResourceID: invoice.ID.String(),
    Metadata:   map[string]any{"amount": invoice.Total, "currency": "KES"},
})
```

---

### 4.8 Disaster Recovery Playbook

**RTO (Recovery Time Objective): 30 minutes**
**RPO (Recovery Point Objective): 1 hour (WAL archiving) / 24 hours (daily dump fallback)**

| Scenario | Detection | Recovery Path | Estimated Time |
|---|---|---|---|
| Application crash | Grafana alert + Caddy 502 | `docker compose restart awo-api` | 2 minutes |
| Database corruption | pg_dump fails / query errors | Restore from latest pg_dump → replay WAL | 20–30 minutes |
| Server destroyed | All monitoring goes dark | Provision new VPS → restore from B2 | 45–60 minutes |
| Accidental bulk delete | Client reports missing data | PITR restore to pre-delete timestamp | 20 minutes |
| Redis total loss | Session errors across all tenants | `docker compose restart redis` (data from AOF) | 5 minutes |

Full step-by-step for each scenario is in [Part IX — Runbooks](#part-ix--solo-developer-runbooks).

---

## Part V — Deployment

### 5.1 The Golden Rule: One-Command Deploy

```bash
./scripts/deploy.sh v1.2.3
```

That is the entire deploy. It should:
1. Build and push the Docker image
2. Pull new image on server
3. Run migrations
4. Perform health check
5. Switch traffic
6. Roll back automatically if health check fails

---

**`scripts/deploy.sh`:**

```bash
#!/bin/bash
set -euo pipefail

VERSION="${1:-latest}"
SERVER="root@awo-production"
COMPOSE_DIR="/opt/awo"
IMAGE="ghcr.io/your-org/awo-erp"
HEALTH_URL="https://api.awo.so/health"
SLACK_WEBHOOK="${SLACK_WEBHOOK:-}"

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }
notify() { [[ -n "$SLACK_WEBHOOK" ]] && curl -s -X POST "$SLACK_WEBHOOK" -d "{\"text\":\"$*\"}" || true; }

log "Deploying AWO ERP $VERSION"
notify "🚀 Deploy started: $VERSION"

# 1. Build & push
docker build --build-arg VERSION="$VERSION" -t "$IMAGE:$VERSION" .
docker push "$IMAGE:$VERSION"

# 2. SSH to server, update and deploy
ssh "$SERVER" bash << REMOTE
  set -euo pipefail
  cd $COMPOSE_DIR

  # Save current version for rollback
  PREVIOUS=\$(grep IMAGE_TAG .env | cut -d= -f2)
  echo "\$PREVIOUS" > /tmp/awo_previous_version

  # Update image tag
  sed -i "s/IMAGE_TAG=.*/IMAGE_TAG=$VERSION/" .env

  # Pull new image
  docker compose pull awo-api awo-worker

  # Run migrations (safe — idempotent)
  docker compose run --rm awo-api ./awo migrate up

  # Rolling restart: worker first (no traffic), then api
  docker compose up -d --no-deps awo-worker
  sleep 5
  docker compose up -d --no-deps awo-api
REMOTE

# 3. Health check (retry for 60 seconds)
log "Health checking..."
for i in {1..12}; do
  if curl -sf "$HEALTH_URL" > /dev/null; then
    log "Health check passed"
    notify "✅ Deploy successful: $VERSION"
    exit 0
  fi
  log "Health check attempt $i failed, retrying..."
  sleep 5
done

# 4. Auto-rollback
log "Health check failed — rolling back!"
notify "❌ Deploy FAILED: $VERSION — rolling back"
PREVIOUS=$(ssh "$SERVER" cat /tmp/awo_previous_version)
"$0" "$PREVIOUS"
exit 1
```

---

### 5.2 CI/CD Pipeline Design (GitHub Actions)

```yaml
# .github/workflows/deploy.yml
name: Build, Test & Deploy

on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        options: [staging, production]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: awo_test
          POSTGRES_USER: awo
          POSTGRES_PASSWORD: testpassword
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
      redis:
        image: redis:7-alpine

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Install tools
        run: |
          go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
          go install github.com/google/wire/cmd/wire@latest

      - name: Verify generated files are up to date
        run: |
          make generate
          git diff --exit-code db/sqlc/ wire_gen.go

      - name: Run tests
        run: make test
        env:
          DATABASE_URL: postgres://awo:testpassword@localhost:5432/awo_test

      - name: Run linter
        uses: golangci/golangci-lint-action@v6

  build:
    needs: test
    runs-on: ubuntu-latest
    outputs:
      image-tag: ${{ steps.meta.outputs.version }}
    steps:
      - uses: actions/checkout@v4

      - name: Log in to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Docker metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ghcr.io/${{ github.repository }}
          tags: |
            type=sha,prefix=,format=short

      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - name: Deploy
        env:
          SSH_KEY: ${{ secrets.PRODUCTION_SSH_KEY }}
          SLACK_WEBHOOK: ${{ secrets.SLACK_WEBHOOK }}
        run: |
          echo "$SSH_KEY" > /tmp/deploy_key && chmod 600 /tmp/deploy_key
          GIT_SSH_COMMAND="ssh -i /tmp/deploy_key" ./scripts/deploy.sh ${{ needs.build.outputs.image-tag }}
```

---

### 5.3 Zero-Downtime Strategy on a Budget

True blue/green requires two servers. For solo dev, use this instead:

```
Strategy: Worker-first rolling restart

1. Restart awo-worker (no user traffic) → 5 second startup
2. Wait for worker healthy (Temporal reconnected)
3. Restart awo-api (brief connection reset, Caddy buffers)
   └── Caddy will retry failed connections for 10s
   └── In-flight requests complete (Fiber graceful shutdown = 30s)
4. Caddy routes to new awo-api

Downtime: ~0 seconds for stateless requests
         ~0 seconds for sessions (stored in Redis)
         Temporal workflows: safe (durable execution, auto-resume)
```

Configure graceful shutdown in AWO:

```go
// cmd/api/main.go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
<-quit

log.Info().Msg("Shutting down gracefully...")
if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
    log.Fatal().Err(err).Msg("Forced shutdown")
}
```

---

### 5.4 Codegen in CI (Wire + SQLC)

The CI pipeline verifies generated files are committed and up-to-date. This prevents a common solo dev failure: forgetting to run `make generate` before pushing.

```yaml
- name: Verify generated files are up to date
  run: |
    make generate
    # If this fails, a developer changed a query or provider without regenerating
    git diff --exit-code db/sqlc/ wire_gen.go internal/mocks/
```

**`Makefile` targets:**

```makefile
.PHONY: generate sqlc mock wire test lint build

generate: sqlc mock wire

sqlc:
	sqlc generate

mock:
	go generate ./...

wire:
	cd cmd/api && wire

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

build:
	go build -ldflags="-X main.Version=$(VERSION)" -o bin/awo ./cmd/api

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t awo-erp:$(VERSION) .
```

---

### 5.5 Environment Promotion

```
Development (local Termux / laptop)
  └── docker-compose.dev.yml
  └── .env.local (committed with safe defaults, no real secrets)
  └── make generate && make test before every push

Staging (same VPS, different Docker network)
  └── docker-compose.staging.yml
  └── /etc/awo/staging.env (server-only, not in git)
  └── Auto-deployed on every push to main

Production
  └── docker-compose.yml
  └── /etc/awo/production.env (server-only, never in git)
  └── Deployed via: ./scripts/deploy.sh <version>
  └── Requires explicit manual trigger or tagged release
```

---

### 5.6 Rollback Runbook

```bash
# Immediate rollback (< 2 minutes)
ssh root@awo-production
cd /opt/awo

# See previous version
cat /tmp/awo_previous_version

# Roll back
sed -i "s/IMAGE_TAG=.*/IMAGE_TAG=$(cat /tmp/awo_previous_version)/" .env
docker compose pull awo-api awo-worker
docker compose up -d --no-deps awo-api awo-worker

# Verify
curl -sf https://api.awo.so/health && echo "Rollback successful"
```

> **Database migrations:** If the new version ran a destructive migration (rare but possible), a code rollback is not enough. You must also run `./awo migrate down 1`. This is why additive-only migrations are enforced — most rollbacks require zero DB changes.

---

## Part VI — Observability (Lightweight)

### 6.1 The Lightweight Stack

```
┌─────────────────────────────────────────────────────────┐
│  What You Run                  Why Not The Alternative  │
├─────────────────────────────────────────────────────────┤
│  Prometheus (metrics)          Not Datadog ($$$)        │
│  Loki (log aggregation)        Not Elasticsearch (RAM)  │
│  Grafana (dashboards)          Not Kibana (complexity)  │
│  Temporal UI (workflow viz)    Built-in, free           │
│  BetterUptime (external ping)  Not PagerDuty ($$$)      │
│  Zerolog (structured logs)     Already in AWO           │
└─────────────────────────────────────────────────────────┘

Total additional RAM: ~400MB (Prometheus + Loki + Grafana on same VPS)
Total cost: $0
```

**Prometheus scrape config (`observability/prometheus.yml`):**

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'awo-api'
    static_configs:
      - targets: ['awo-api:9090']

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  - job_name: 'caddy'
    static_configs:
      - targets: ['caddy:2019']
```

---

### 6.2 Structured Logging with Zerolog

AWO uses Zerolog. Every log entry must be structured — no `fmt.Printf` in production code.

**What to log:**

```go
// ✅ Good — structured, searchable in Loki
log.Info().
    Str("tenant_id", tenantID.String()).
    Str("user_id", userID.String()).
    Str("action", "invoice.created").
    Str("invoice_id", invoice.ID.String()).
    Dur("duration_ms", time.Since(start)).
    Msg("Invoice created")

// ❌ Bad — not queryable
log.Info().Msg(fmt.Sprintf("Invoice %s created for tenant %s", invoiceID, tenantID))
```

**What NOT to log (especially in Kenya context — data privacy):**

```
❌ Never log:
  - Passwords or tokens (even hashed)
  - Full API keys
  - KRA PIN numbers
  - M-Pesa phone numbers in full (mask last 4: 0712***456)
  - Bank account numbers
  - National ID numbers
  - Full request/response bodies on financial endpoints
```

**Log levels in production:**

| Level | When to Use |
|---|---|
| `Error` | Unexpected failures that need attention |
| `Warn` | Expected failures (auth failures, not found) — low volume |
| `Info` | Business events (tenant created, invoice issued, user login) |
| `Debug` | Disabled in production; enabled per-request via header for debugging |

---

### 6.3 Distributed Tracing (OpenTelemetry, Solo-Scale)

AWO has `tracing.Service` wired. For solo scale, keep it simple:

```
Development: Console exporter (print spans to stdout)
Staging:     Jaeger (all-in-one Docker container)
Production:  Jaeger OR skip until you have a performance problem to diagnose
```

**Pragmatic tracing rule:** Instrument at the handler and DB layers only. Do not manually create spans in every function. The Fiber OpenTelemetry middleware handles HTTP spans automatically.

```go
// Add to middleware stack in routes.go — this covers all HTTP traces
app.Use(otelfiber.Middleware())
```

---

### 6.4 Temporal Workflow Visibility

Temporal ships with a UI at port 8080. Expose it internally:

```
temporal.awo.so → Caddy → temporal:8080 (with basicauth)
```

The Temporal UI shows:
- All running/completed/failed workflows
- Workflow execution history (full event log)
- Activity retry history
- Search by workflow type, tenant ID, status

This is your primary debugging tool for async operations. No additional instrumentation needed.

---

### 6.5 The Five Alerts That Actually Matter

Configure these in Grafana. Everything else is noise.

```yaml
# 1. Application Down
Alert: awo_api_up == 0
Severity: CRITICAL
Action: Page immediately — this is user-visible

# 2. Database Connection Exhausted
Alert: pg_stat_activity_count > 45  # Near max_connections of 50
Severity: HIGH
Action: Check PgBouncer pool; restart if needed

# 3. Disk Space Critical
Alert: node_filesystem_free_bytes < 5GB
Severity: HIGH
Action: Clean Docker images; check pg WAL archive size

# 4. Backup Failed
Alert: awo_backup_last_success_timestamp > 26h  # Miss 2 consecutive
Severity: HIGH
Action: Check backup script logs immediately

# 5. High Error Rate
Alert: rate(awo_http_errors_total[5m]) > 10
Severity: MEDIUM
Action: Check logs for new error pattern; may indicate bad deploy
```

**Notification channels (free tier friendly):**

```
Grafana → Slack webhook (free) — for INFO/MEDIUM
Grafana → Telegram bot (free) — for HIGH/CRITICAL
BetterUptime → SMS/call (free tier = 3 monitors) — for DOWN
```

---

### 6.6 Dashboard Design for a Solo Operator

One dashboard. Five rows. Open it every morning.

```
Row 1: Health
  └── API up/down | DB connections | Redis memory | Disk used

Row 2: Business Metrics
  └── Active tenants | Requests/min per tenant | Error rate | P95 latency

Row 3: Database
  └── QPS | Slow queries (>200ms) | Lock waits | Replication lag

Row 4: Infrastructure
  └── CPU % | RAM % | Network in/out | Disk I/O

Row 5: Backup Status
  └── Last backup time | Backup file size | Restore test result
```

---

## Part VII — Security & Configuration

### 7.1 Secrets Management (No .env in Production)

```
❌ Never:
  - Commit .env files with real credentials to git
  - Pass secrets as Docker environment variables in Compose file
  - Store secrets in plaintext on disk
  - Log secrets at any level

✅ Do:
  - Store secrets in /etc/awo/secrets/ (mode 0600, owned by awo user)
  - Reference secrets via Docker secrets (file-based, not env)
  - Encrypt secrets at rest with `age` if backups include the server disk
  - Rotate database passwords quarterly
```

**File structure on production server:**

```
/etc/awo/
├── production.env          # Non-secret config (feature flags, timeouts)
├── secrets/
│   ├── pg_password         # PostgreSQL password (mode 0400)
│   ├── redis_password      # Redis password (mode 0400)
│   ├── jwt_secret          # JWT signing key (mode 0400)
│   ├── grafana_password    # Grafana admin (mode 0400)
│   └── age.key             # age encryption key for backup (mode 0400)
```

**`/etc/awo/production.env` (non-secret, allowed in git as template):**

```env
# Application
APP_ENV=production
APP_PORT=3000
APP_BASE_DOMAIN=awo.so
LOG_LEVEL=info

# Database (password loaded via Docker secret)
DB_HOST=pgbouncer
DB_PORT=6432
DB_NAME=awo_production
DB_USER=awo_app
DB_MAX_CONNS=20
DB_MIN_CONNS=5

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_DB=0

# Temporal
TEMPORAL_HOST=temporal:7233
TEMPORAL_NAMESPACE=awo-production

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4317
PROMETHEUS_PORT=9090
```

---

### 7.2 Config Hierarchy (Viper)

AWO uses Viper. The load order (later overrides earlier):

```
1. Default values (in config.go)
2. Config file (/etc/awo/config.yaml)
3. Environment variables (production.env via Docker)
4. Command-line flags (for one-off overrides)
```

**Never use `os.Getenv()` directly in AWO code.** Always go through the `config.Config` struct loaded at startup via `config.Load()`. This ensures all config is validated at startup, not at first use.

---

### 7.3 TLS, CORS, CSRF & Security Headers

**CORS** — the open gap in AWO must be fixed before production:

```go
// internal/api/middleware/security.go
// Replace hardcoded dev origins with config-driven allowlist

func NewRouteSecurityManager(cfg *config.Config) *RouteSecurityManager {
    return &RouteSecurityManager{
        allowedOrigins: cfg.Security.AllowedOrigins, // ["https://*.awo.so"]
    }
}
```

**Required security headers (Caddy layer):**

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'
```

---

### 7.4 API Key Lifecycle Management

AWO's IAM module includes API key management. For the Kenya market, API keys are critical because:
- M-Pesa webhook callbacks use API keys, not sessions
- KRA ETR integrations use server-to-server API keys
- Third-party accounting software integrations need long-lived keys

**Key rotation policy:**

| Key Type | Rotation Frequency | Auto-Rotate? |
|---|---|---|
| JWT signing key | 90 days | No — requires redeploy |
| Session tokens | On logout / 24h idle | Yes — Redis TTL |
| API keys (user-issued) | User-controlled | No |
| API keys (integration) | 180 days | Send reminder email at 150 days |
| Database password | 90 days | No — manual |

---

### 7.5 RLS as a Security Control

PostgreSQL RLS is not just a tenancy convenience — it is a **defence-in-depth security control**. Even if:
- A bug in the application bypasses the authorization middleware
- An SQLC query is called without the service layer
- A developer accidentally writes a raw query

...RLS still enforces data isolation at the database level, as long as:

```
1. awo_app role cannot bypass RLS (ALTER ROLE awo_app NOINHERIT; / FORCE ROW LEVEL SECURITY)
2. Every table has RLS ENABLED and FORCED
3. WithTenant() is always used (enforced by code review and naming conventions)
4. The superuser (postgres) is NEVER used by the application — only for migrations
```

Periodically audit with:

```sql
-- Find any tables that DON'T have RLS enabled
SELECT tablename, rowsecurity
FROM pg_tables
WHERE schemaname = 'public'
  AND rowsecurity = false
  AND tablename NOT IN ('schema_migrations'); -- whitelist of global tables
```

---

## Part VIII — Maintainability & Operational Fitness

### 8.1 The Solo Developer's Weekly Ops Ritual

**Every Monday morning (~20 minutes):**

```
□ 1. Check Grafana dashboard — anything red from the weekend?
□ 2. Verify last backup succeeded (Row 5 of dashboard)
□ 3. Check disk usage — is WAL archive growing faster than expected?
□ 4. Review error logs: docker compose logs --since 7d awo-api | grep ERROR
□ 5. Check Temporal UI — any stuck or failed workflows?
□ 6. Review pending GitHub Dependabot PRs — any security patches?
□ 7. Check PgBouncer stats: SHOW POOLS; SHOW STATS; (via pgbouncer admin)
```

**Every month (~2 hours):**

```
□ 1. Run backup restore test (restore to staging, verify data integrity)
□ 2. Review slow query log (queries > 200ms) — add indexes if needed
□ 3. Run VACUUM ANALYZE on production database
□ 4. Rotate any secrets due for rotation this month
□ 5. Review open gaps in AWO (see 8.3) — close one
□ 6. Update Go dependencies: go get -u ./... → run tests → merge
□ 7. Review server costs — are you in budget?
```

---

### 8.2 Codebase Health Gates

These checks run in CI and block merges if they fail:

```makefile
# Every PR must pass:
ci-check:
  make generate          # Regenerate
  git diff --exit-code   # Nothing should change if you committed generated files
  make test              # All tests pass
  make lint              # golangci-lint clean

# Additional checks (weekly, not per PR):
health:
  go vet ./...
  govulncheck ./...      # Check for known CVEs in dependencies
  staticcheck ./...      # Additional static analysis
```

**Non-negotiable code rules for AWO:**

```
1. No raw queries outside db/sqlc/ — use SQLC
2. No store access outside repository layer — enforce via linting
3. No I-prefixed interfaces — role-suffixed names only
4. No hand-editing wire_gen.go or db/sqlc/ — regenerate
5. No fmt.Printf in non-cmd code — use logger.Logger
6. No panic() in non-startup code — return errors
7. No time.Sleep() in handler/service layer — use Temporal for waits
```

---

### 8.3 Technical Debt Register — AWO's Known Open Gaps

These are known issues to close in priority order:

| # | Gap | Risk | Fix |
|---|---|---|---|
| 1 | CORS origins hardcoded to dev | HIGH — blocks production | Move to config.Security.AllowedOrigins |
| 2 | Email service absent | HIGH — password reset broken | Integrate Resend or SendGrid (free tier) |
| 3 | Rate limiting TODO in routes.go | MEDIUM — DoS exposure | Add golang.org/x/time/rate middleware |
| 4 | TLD-aware subdomain parsing missing | MEDIUM — blocks .co.ke domains | Use golang.org/x/net/publicsuffix |
| 5 | Double-entry enforcement layer unconfirmed | HIGH — financial integrity | Add DB CHECK constraint on journal entries |
| 6 | Zero Temporal workflows | LOW — feature gap | Implement as features demand |
| 7 | Notification service absent | LOW | Queue-based email/SMS via Temporal activity |

---

### 8.4 Module Boundary Rules

AWO is a monolith with domain boundaries. These boundaries are intentional and must not be blurred:

```
❌ Never:
  - Import internal/core/finance from internal/core/iam (cross-domain direct coupling)
  - Import internal/api/handlers from internal/core/* (inward dependency violation)
  - Access db.Store directly from a handler (skip service+repository layers)
  - Call another domain's repository from a service (go through that domain's service)

✅ Allowed:
  - shared/* used by all layers
  - platform/* used by all layers
  - Handler → Service → Repository → Store
  - Service → another domain's Service (not its repository)
  - Temporal Activity → Service layer (activities are thin wrappers)
```

---

### 8.5 Dependency Management & Upgrade Policy

```bash
# Check for outdated dependencies
go list -m -u all | grep '\['

# Check for known CVEs
govulncheck ./...

# Upgrade policy:
# - Security patches: apply immediately, run tests, deploy
# - Minor versions: batch monthly, run full test suite
# - Major versions (e.g., Fiber v2 → v3): evaluate in a branch; never rush
# - Generated tool versions (sqlc, wire): pin in Makefile, upgrade carefully
```

**Pin tool versions in Makefile:**

```makefile
SQLC_VERSION := v1.27.0
WIRE_VERSION := v0.6.0
GOLANGCI_VERSION := v1.62.0

tools:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
	go install github.com/google/wire/cmd/wire@$(WIRE_VERSION)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_VERSION)
```

---

## Part IX — Solo Developer Runbooks

> These runbooks are designed to be executed alone, at any time of day, from a phone terminal if necessary. Each has a time estimate and escalation path.

---

### RB-01 Application Won't Start

**Symptoms:** 502 Bad Gateway; `awo-api` container exits immediately  
**Time estimate:** 5–10 minutes

```bash
# Step 1: Check container status
docker compose ps

# Step 2: Read exit logs
docker compose logs --tail=100 awo-api

# Step 3: Common causes and fixes

# Cause A: DB not ready yet
# Fix: Wait 30s and retry; check postgres health
docker compose ps postgres

# Cause B: Config env var missing
# Symptom: "required config key X not found"
# Fix: Check /etc/awo/production.env; compare against .env.example

# Cause C: Migration failed (table missing)
# Symptom: "relation X does not exist"
# Fix: Run migrations manually
docker compose run --rm awo-api ./awo migrate up

# Cause D: Port conflict
# Symptom: "bind: address already in use"
# Fix:
ss -tlnp | grep 3000
kill -9 <PID>

# Cause E: Binary panic at startup
# Symptom: goroutine stack trace
# Fix: Roll back to previous version (see RB-09)

# Step 4: Start and verify
docker compose up -d awo-api
sleep 5
curl -sf http://localhost:3000/health && echo "OK"
```

---

### RB-02 Database Connection Exhausted

**Symptoms:** Errors like "sorry, too many clients already"; p95 latency spike  
**Time estimate:** 5 minutes

```bash
# Step 1: Check current connections
docker exec -it postgres psql -U awo_app awo_production -c "
  SELECT count(*), state, wait_event_type, wait_event
  FROM pg_stat_activity
  GROUP BY state, wait_event_type, wait_event
  ORDER BY count DESC;
"

# Step 2: Check PgBouncer pool stats
docker exec -it pgbouncer psql -p 6432 -U pgbouncer pgbouncer -c "SHOW POOLS;"
docker exec -it pgbouncer psql -p 6432 -U pgbouncer pgbouncer -c "SHOW STATS;"

# Step 3: Find and kill long-running idle connections (if any)
docker exec -it postgres psql -U awo_app awo_production -c "
  SELECT pg_terminate_backend(pid)
  FROM pg_stat_activity
  WHERE state = 'idle'
    AND state_change < NOW() - INTERVAL '10 minutes'
    AND pid <> pg_backend_pid();
"

# Step 4: If PgBouncer itself is exhausted, restart it
docker compose restart pgbouncer

# Step 5: If problem recurs, reduce AWO's DB pool size in production.env
# DB_MAX_CONNS=10  (from 20)
# Then restart api
docker compose restart awo-api
```

---

### RB-03 Tenant Data Bleed (RLS Failure)

**Symptoms:** A tenant reports seeing another tenant's data — CRITICAL security incident  
**Time estimate:** 15–20 minutes  
**This is your highest severity incident.**

```bash
# Step 1: Immediately pull all clients' sessions to force re-auth
docker exec redis redis-cli FLUSHDB  # Nuclear option — all sessions invalidated
# OR targeted: find the affected tenant's sessions and delete

# Step 2: Check if RLS is enabled on all tables
docker exec -it postgres psql -U postgres awo_production -c "
  SELECT tablename, rowsecurity, forcerowsecurity
  FROM pg_tables
  WHERE schemaname = 'public'
  ORDER BY tablename;
"
# Every tenant data table must show: rowsecurity=on, forcerowsecurity=on

# Step 3: Check if a recent migration disabled RLS
git log --oneline db/migration/ -20
# Review each recent migration for: ALTER TABLE x DISABLE ROW LEVEL SECURITY

# Step 4: Check if awo_app role can bypass
docker exec -it postgres psql -U postgres awo_production -c "
  SELECT rolname, rolinherit, rowsecurity
  FROM pg_roles
  WHERE rolname = 'awo_app';
"
# rowsecurity must be 'on'

# Step 5: Re-enable RLS on any table where it was disabled
docker exec -it postgres psql -U postgres awo_production -c "
  ALTER TABLE <affected_table> ENABLE ROW LEVEL SECURITY;
  ALTER TABLE <affected_table> FORCE ROW LEVEL SECURITY;
"

# Step 6: Audit which data was accessed
# Query your audit log (audit.events table) for the affected tenant_id
# during the window of the bleed

# Step 7: Notify affected tenants — this is a legal obligation in Kenya
# Draft breach notification email (data protection obligations)
```

---

### RB-04 Migration Failed Mid-Deployment

**Symptoms:** Deploy script failed after `migrate up`; app may be on old binary with partial schema  
**Time estimate:** 10–15 minutes

```bash
# Step 1: Check migration state
docker compose run --rm awo-api ./awo migrate status

# Step 2: Check if migration is marked dirty
# If dirty=true, golang-migrate requires manual intervention:
docker exec -it postgres psql -U awo_app awo_production -c "
  SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;
"

# Step 3a: If migration was partially applied and caused damage, roll back:
docker compose run --rm awo-api ./awo migrate down 1

# Step 3b: If migration failed cleanly (before making changes):
# Fix the issue in the SQL file, push a corrected migration
# Do NOT re-run a partially-applied migration without first rolling it back

# Step 4: If dirty flag is set, force-clear it (DANGEROUS — only if you know state):
docker exec -it postgres psql -U postgres awo_production -c "
  UPDATE schema_migrations SET dirty = false WHERE version = <VERSION>;
"

# Step 5: Roll back application to previous version (see RB-09)
# Then fix migration, push, and redeploy cleanly

# Prevention: Always test migrations on staging first
# Prevention: Write reversible migrations (test the .down.sql)
```

---

### RB-05 Full Database Restore from Backup

**Symptoms:** Database corrupted or lost; need to restore from backup  
**Time estimate:** 20–45 minutes depending on DB size  
**RTO target: 30 minutes**

```bash
# Step 1: Stop the application (prevent writes during restore)
docker compose stop awo-api awo-worker

# Step 2: Download latest backup from Backblaze B2
b2 ls awo-backups daily/ | tail -5  # Find latest
b2 download-file-by-name awo-backups daily/awo_<timestamp>.pgdump /tmp/restore.pgdump

# Step 3: Drop and recreate database
docker exec -it postgres psql -U postgres -c "
  SELECT pg_terminate_backend(pid)
  FROM pg_stat_activity
  WHERE datname = 'awo_production';
"
docker exec -it postgres psql -U postgres -c "DROP DATABASE awo_production;"
docker exec -it postgres psql -U postgres -c "CREATE DATABASE awo_production OWNER awo_app;"

# Step 4: Restore
docker exec -it postgres pg_restore \
  --host=localhost \
  --username=awo_app \
  --dbname=awo_production \
  --no-owner \
  --no-privileges \
  /tmp/restore.pgdump

# Step 5: Re-apply RLS policies (if backup pre-dates RLS setup)
docker exec -it postgres psql -U postgres awo_production < scripts/init-rls.sql

# Step 6: Run any migrations that occurred after backup
docker compose run --rm awo-api ./awo migrate up

# Step 7: Verify data integrity
docker exec -it postgres psql -U awo_app awo_production -c "
  SELECT tablename, n_live_tup
  FROM pg_stat_user_tables
  ORDER BY n_live_tup DESC LIMIT 20;
"

# Step 8: Start application
docker compose up -d awo-api awo-worker

# Step 9: Verify health
curl -sf https://api.awo.so/health && echo "Restore complete"

# Step 10: Notify tenants of service restoration and data loss window
```

---

### RB-06 Temporal Worker Down

**Symptoms:** Async operations not processing; workflow statuses stuck in "Running"  
**Time estimate:** 5 minutes

```bash
# Step 1: Check worker status
docker compose ps awo-worker
docker compose logs --tail=50 awo-worker

# Step 2: Check Temporal server connectivity
docker compose logs --tail=20 temporal

# Step 3: Restart worker
docker compose restart awo-worker

# Step 4: Verify worker reconnected in Temporal UI
# Open https://temporal.awo.so
# Check Workers section — should show awo-worker connected

# Step 5: If Temporal server itself is down, restart it
docker compose restart temporal

# Note: All in-flight workflows are DURABLE in Temporal's DB
# When worker reconnects, they resume automatically — no data loss
# Temporal stores its state in PostgreSQL (same DB, different schema)

# Step 6: If workflows are stuck after worker restart, force-retry:
# In Temporal UI: select workflow → Signal → Reset to last activity
```

---

### RB-07 Redis Cache Corruption / Eviction Storm

**Symptoms:** All users getting logged out; high DB load spike; "session not found" errors  
**Time estimate:** 5 minutes

```bash
# Step 1: Check Redis memory
docker exec redis redis-cli INFO memory | grep used_memory_human

# Step 2: Check eviction stats
docker exec redis redis-cli INFO stats | grep evicted

# Step 3a: If corruption suspected, restart Redis (AOF ensures data survives)
docker compose restart redis

# Step 3b: If eviction storm (maxmemory hit), increase memory limit
# In docker-compose.yml:
# command: redis-server --appendonly yes --maxmemory 1gb --maxmemory-policy allkeys-lru
docker compose up -d --no-deps redis

# Step 4: If cache is intentionally being flushed (e.g., after security event)
docker exec redis redis-cli FLUSHDB
# All sessions cleared — users must re-login
# Tenant slug cache repopulates automatically on next request

# Step 5: Verify Redis recovery
docker exec redis redis-cli PING  # Should return PONG
docker exec redis redis-cli INFO keyspace  # Show databases and key counts
```

---

### RB-08 Server Disk Full

**Symptoms:** Writes failing; logs stopping; PostgreSQL errors about disk space  
**Time estimate:** 10 minutes

```bash
# Step 1: Find what is using disk
df -h
du -sh /var/lib/docker/volumes/*  # Docker volumes
du -sh /var/wal-archive/          # WAL archive (usually the culprit)
du -sh /var/log/                  # Logs

# Step 2: Clean Docker build cache and unused images
docker system prune -f
docker image prune -a -f          # Remove all untagged images

# Step 3: Rotate WAL archive (keep only last 3 days)
find /var/wal-archive -mtime +3 -delete

# Step 4: Rotate old log files
find /var/log/caddy -mtime +7 -delete
journalctl --vacuum-time=7d

# Step 5: Clean old pg_dump temp files
find /tmp -name "*.pgdump" -mtime +1 -delete

# Step 6: Verify postgres can write again
docker exec -it postgres psql -U postgres -c "SELECT 1;"

# Step 7: Set up automatic cleanup (add to cron)
# 0 3 * * * find /var/wal-archive -mtime +3 -delete
# 0 4 * * 0 docker system prune -f

# Prevention: Alert when disk < 5GB (see 6.5 — Alert #3)
```

---

### RB-09 Emergency Rollback (Production)

**Symptoms:** Bad deploy broke production; need immediate rollback  
**Time estimate:** 2–3 minutes

```bash
# Step 1: Connect to server
ssh root@awo-production

# Step 2: Check current and previous versions
cat /opt/awo/.env | grep IMAGE_TAG          # Current
cat /tmp/awo_previous_version               # Previous

# Step 3: Roll back
cd /opt/awo
PREV=$(cat /tmp/awo_previous_version)
sed -i "s/IMAGE_TAG=.*/IMAGE_TAG=${PREV}/" .env

docker compose pull awo-api awo-worker
docker compose up -d --no-deps awo-worker
sleep 5
docker compose up -d --no-deps awo-api

# Step 4: Verify
sleep 10
curl -sf https://api.awo.so/health && echo "Rollback complete: running ${PREV}"

# Step 5: Check if migrations need reverting
# Only needed if the new version added a breaking migration
docker compose run --rm awo-api ./awo migrate status
# If current version > what previous binary expects:
docker compose run --rm awo-api ./awo migrate down 1

# Step 6: Notify team/yourself via Slack
# "Rolled back from X to Y — investigating root cause"

# Step 7: Root cause analysis before redeploying
# Check CI test logs — what test did the bad version pass that it shouldn't have?
```

---

### RB-10 New Tenant Onboarding Checklist

**Use this every time a new business signs up for AWO ERP.**  
**Time estimate:** 15 minutes

```bash
# Step 1: Create tenant via AWO API (or admin script)
curl -X POST https://api.awo.so/api/v1/admin/tenants \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Ltd",
    "slug": "acme",
    "plan": "standard",
    "country": "KE",
    "currency": "KES"
  }'

# Step 2: Verify tenant was created and is in PENDING state
curl https://api.awo.so/api/v1/admin/tenants/acme

# Step 3: Create admin user for tenant
curl -X POST https://acme.awo.so/api/v1/iam/users \
  -d '{"email": "admin@acmeltd.co.ke", "role": "tenant_admin"}'

# Step 4: Verify subdomain resolves (Cloudflare wildcard should handle this)
curl -I https://acme.awo.so/health
# Expected: 200 OK (or redirect)

# Step 5: Activate tenant (PENDING → ACTIVE)
curl -X PATCH https://api.awo.so/api/v1/admin/tenants/acme \
  -d '{"status": "active"}'

# Step 6: Verify RLS isolation
# Run the RLS verification query in RB-03 to confirm tenant is isolated

# Step 7: Send welcome email with login credentials
# (Manual until email service is implemented — gap from 8.3)

# Step 8: Log in the Tenant Register spreadsheet:
#   Date | Company | Slug | Plan | Contact | KRA PIN | Notes

# Step 9: Verify billing record created (if billing module is active)
```

---

## Part X — Evolution Path

### 10.1 Stage 0 → Stage 1: Local to Live (First Tenant)

**Duration:** 1–2 weeks  
**Goal:** One paying tenant on production

```
Checklist:
□ VPS provisioned (Hetzner CX21 minimum)
□ Domain and wildcard DNS configured
□ Caddy running with HTTPS
□ PostgreSQL with RLS verified (run RB-03 test)
□ Backup script running (verify with manual trigger)
□ BetterUptime monitoring the health endpoint
□ CORS origins updated to production domains
□ Email service wired (Resend free tier — 100 emails/day)
□ Rate limiting implemented in routes.go
□ One tenant onboarded and verified end-to-end
□ First backup restore test completed
```

---

### 10.2 Stage 1 → Stage 2: Hardening (First Paying Customer)

**Duration:** 1 month  
**Goal:** Production-grade reliability for revenue-generating tenants

```
Close these AWO open gaps (from 8.3):
□ CORS hardened
□ Rate limiting live
□ Double-entry constraint enforced at DB level
□ Email notifications working (password reset minimum)
□ Audit log verified for KRA compliance

Infrastructure hardening:
□ Upgrade VPS to CX31 (8GB RAM)
□ Cloudflare Pro (WAF enabled)
□ Automated weekly backup restore test
□ PgBouncer metrics in Grafana
□ Five critical alerts all configured and tested
□ Security headers verified (use securityheaders.com)
□ Penetration test (basic — OWASP ZAP scan against staging)

Operational:
□ Weekly ops ritual established (see 8.1)
□ Incident response playbook printed (yes, physical copy)
□ Monthly billing report for tenant (even if manual)
```

---

### 10.3 Stage 2 → Stage 3: Growth (Team + Multi-Region)

**Trigger signals:** 50+ active tenants, consistent revenue, first team hire

```
Infrastructure evolution:
□ Separate PostgreSQL to dedicated VPS (Hetzner CX41)
□ Read replica for reporting queries
□ Staging environment upgraded to production-equivalent
□ Consider Hetzner Johannesburg (lower latency from Kenya than EU)
□ Evaluate Azure Nairobi PoP for primary hosting when mature

Architecture evolution:
□ Extract Temporal worker to separate service (if CPU-intensive)
□ Separate observability stack to its own VPS
□ Implement custom domain support (fix TLD parsing gap)
□ M-Pesa webhook processor as dedicated Temporal workflow

Team evolution:
□ Code review process established (no more solo merges to main)
□ On-call rotation documented (even if just you + one hire)
□ Runbooks reviewed and updated by new team member
□ Architecture Decision Records (ADRs) started
```

---

### 10.4 Temporal Activation Roadmap

Temporal is wired but dormant. Activate workflows in this order — highest business value first:

| Priority | Workflow | Why It Needs Temporal | Dependencies |
|---|---|---|---|
| 1 | **Bulk invoice generation** | Long-running, must survive server restart | Finance module complete |
| 2 | **M-Pesa STK Push + reconciliation** | Multi-step, compensatable, external API | M-Pesa SDK integrated |
| 3 | **Month-end close** | Multi-day process with human approval gates | Finance + Auth complete |
| 4 | **Tenant onboarding** | Multi-step: create → setup → notify → activate | Email service live |
| 5 | **KRA ETR submission** | Retry logic for government API unreliability | ETR integration built |
| 6 | **Payroll processing** | Time-bound, multi-approval, audit-critical | HR module built |

---

### 10.5 When to Stop Being Solo

These signals indicate you need a second developer — not more automation:

```
1. You are responding to production incidents more than once per week
2. A feature request requires a module that touches 3+ domains simultaneously
3. A customer is waiting more than 2 weeks for a critical bug fix
4. Backup restoration has never been tested (you keep meaning to)
5. You cannot take a vacation without laptop anxiety
6. Monthly AWS/Hetzner bills exceed your MRR growth rate
7. You have a customer asking for SLA guarantees you cannot honour alone
```

**When you hire, the first hire should be:**  
An SRE / DevOps engineer who can own the operational layer — not another feature developer. Operational debt is the most expensive kind for a fintech ERP.

---

## Appendices

### A — Environment Variable Reference

```env
# ── Application ──────────────────────────────────────
APP_ENV=production                    # development | staging | production
APP_PORT=3000                         # HTTP port
APP_BASE_DOMAIN=awo.so                # Used for subdomain extraction
APP_GRACEFUL_SHUTDOWN_TIMEOUT=30s

# ── Database ─────────────────────────────────────────
DB_HOST=pgbouncer                     # Always connect through PgBouncer
DB_PORT=6432
DB_NAME=awo_production
DB_USER=awo_app
DB_PASSWORD=                          # Loaded via Docker secret
DB_MAX_CONNS=20
DB_MIN_CONNS=5
DB_CONN_TIMEOUT=5s
DB_QUERY_TIMEOUT=30s

# ── Redis ────────────────────────────────────────────
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=                       # Loaded via Docker secret
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_DIAL_TIMEOUT=5s

# ── Auth / Security ──────────────────────────────────
JWT_SECRET=                           # Loaded via Docker secret — min 32 chars
JWT_EXPIRY=24h
SESSION_TTL=24h
BCRYPT_COST=12                        # 12 for production, 10 for dev

# ── Temporal ─────────────────────────────────────────
TEMPORAL_HOST=temporal:7233
TEMPORAL_NAMESPACE=awo-production
TEMPORAL_TASK_QUEUE=awo-main

# ── Observability ────────────────────────────────────
LOG_LEVEL=info                        # debug | info | warn | error
LOG_FORMAT=json                       # json | pretty (pretty for dev only)
OTEL_SERVICE_NAME=awo-erp
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4317
PROMETHEUS_PORT=9090

# ── Features ─────────────────────────────────────────
FEATURE_MFA_ENABLED=true
FEATURE_RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_SECOND=20
RATE_LIMIT_BURST=50
```

---

### B — Makefile Target Reference

```makefile
# ── Codegen ──────────────────────────────
make generate        # sqlc + mock + wire (run after ANY schema/query change)
make sqlc            # Regenerate db/sqlc/ from db/queries/
make mock            # Regenerate internal/mocks/ via mockgen
make wire            # Regenerate wire_gen.go

# ── Quality ──────────────────────────────
make test            # go test -race -count=1 ./...
make test-integration  # Run integration tests (requires Docker)
make lint            # golangci-lint run ./...
make vet             # go vet ./...
make vuln            # govulncheck ./...

# ── Build ────────────────────────────────
make build           # Build binary to bin/awo
make docker-build    # Build Docker image
make docker-push     # Push to GHCR

# ── Database ─────────────────────────────
make migrate-up      # Apply all pending migrations
make migrate-down    # Roll back last migration
make migrate-status  # Show migration state
make seed            # Run scripts/seed.sh (dev only)

# ── Local Dev ────────────────────────────
make dev             # Start dev stack (docker-compose.dev.yml)
make dev-down        # Stop dev stack
make logs            # Follow all service logs
make psql            # Connect to local PostgreSQL

# ── Ops ──────────────────────────────────
make backup          # Trigger manual backup
make health          # Check all service health endpoints
make deploy VERSION=v1.2.3  # Deploy specific version
```

---

### C — Recommended Tooling Stack

| Category | Tool | Free Tier | Notes |
|---|---|---|---|
| **Hosting** | Hetzner Cloud | No | CX21 at $6/mo is best value globally |
| **DNS / CDN** | Cloudflare | Yes | Nairobi PoP; WAF on Pro |
| **TLS / Proxy** | Caddy | Yes | Zero-config HTTPS |
| **Monitoring** | Grafana Cloud | Yes | 10k metrics free |
| **Uptime** | BetterUptime | Yes | 3 monitors free |
| **Log storage** | Grafana Loki | Yes | Self-hosted is free |
| **Backup storage** | Backblaze B2 | Yes | $0.006/GB — cheapest |
| **Secrets** | `age` encryption | Yes | Simple, auditable |
| **CI/CD** | GitHub Actions | Yes | 2000 min/month free |
| **Container registry** | GitHub GHCR | Yes | Free for public repos |
| **Email** | Resend | Yes | 100 emails/day free |
| **Error tracking** | Sentry | Yes | 5k errors/month free |
| **M-Pesa** | Safaricom Daraja | Usage-based | Register at developer.safaricom.co.ke |
| **DNS registrar** | Porkbun / Namecheap | No | .so TLD ~$25/year |

---

### D — East Africa Provider Directory

| Service | Provider | Contact | Notes |
|---|---|---|---|
| **M-Pesa API** | Safaricom Daraja | developer.safaricom.co.ke | C2B, B2C, STK Push |
| **KRA ETR** | KRA eTIMS | kra.go.ke/etims | VAT Electronic Tax Invoice |
| **SMS Gateway** | Africa's Talking | africastalking.com | Best coverage EA; Nairobi office |
| **Payment Gateway** | Pesapal | pesapal.com | Cards + M-Pesa; KE-based |
| **Cloud (Local)** | Safaricom Cloud | safaricomcloud.com | Nairobi latency; limited services |
| **Cloud (Regional)** | AWS af-south-1 | aws.amazon.com | Cape Town; best AWS for EA |
| **Internet Backup** | Zuku / Faiba | — | BGP failover if on-prem |
| **EPRA (fuel)** | EPRA | epra.go.ke | Petroleum pricing regulations |
| **Data Protection** | ODPC | odpc.go.ke | Kenya Data Protection Act compliance |

---

*AWO ERP Solo Developer Operations Guide — v1.0*  
*Built for the East African market. Optimised for a single developer who ships.*
