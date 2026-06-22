### Chapter 4 — Quick Start — Your First EntityDefinition in 30 Minutes

This chapter walks through the complete creation of a working system entity from zero: tooling installation, project startup, entity definition, database migration, API route verification, a basic amis page, and a simple Temporal workflow trigger. By the end you will have seen the full framework loop in motion and have a reference project to build against.

The example entity is `ServiceRequest` — a simplified field service ticket suitable for the Shell Maanzoni service station context. It has a status field, a customer name, a vehicle registration, a description, and a total cost. This is deliberately simple so the plumbing is visible without domain noise.

---

#### 4.1. Prerequisites and Tooling

##### 4.1.1. Go 1.22 or later

Awo requires Go 1.22 or later. Go 1.22 introduced the finalised `range` over integer syntax and improved the `net/http` routing model that Fiber builds on. Verify your installation with `go version`; the output must show `go1.22` or higher. Install or upgrade from `https://go.dev/dl/`.

The module path in all examples is `github.com/awolabs/awo`. In a real project this would be the module path of your application. The framework packages imported as `github.com/awolabs/awo/pkg/entity` and related paths are stable public API; `internal/` packages are not importable by module code.

##### 4.1.2. Docker Compose for local dependencies

The local development stack requires PostgreSQL 16, Redis 7, and the Temporal development server. These are provided as a `docker-compose.yml` in the starter project. Docker Desktop or Docker Engine with the Compose plugin must be installed. Verify with `docker compose version`; the output must show Compose v2 or later (the `docker compose` subcommand, not the legacy `docker-compose` binary).

The Docker Compose stack exposes PostgreSQL on `localhost:5432`, Redis on `localhost:6379`, and the Temporal development server on `localhost:7233`. The Temporal Web UI is available at `http://localhost:8080` once the stack is running. Do not expose these ports to external networks in a development environment.

##### 4.1.3. Awo CLI installation

The Awo CLI provides the `awo` command for entity scaffolding, migration management, seed data, server startup, and tenant management. Install it with:

```bash
go install github.com/awolabs/awo/cmd/awo@latest
```

Verify with `awo version`. The CLI reads configuration from the `AWO_*` environment variables (Appendix F) or from a `.env` file in the project root. All CLI commands that require database access will fail if `AWO_DB_DSN` is not set.

##### 4.1.4. Atlas CLI installation

Atlas is the database migration tool used by Awo. The `awo entity migrate` command delegates to Atlas under the hood, but installing Atlas directly gives access to its linting and schema inspection commands used in CI. Install from `https://atlasgo.io/getting-started` or with:

```bash
curl -sSf https://atlasgo.sh | sh
```

Verify with `atlas version`. Atlas 0.14 or later is required for the `migrate lint` features used in CI (§44.2.5).

##### 4.1.5. Temporal CLI installation

The Temporal CLI (`temporal`) provides the development server, workflow management commands, and schedule management. It is the easiest way to run a local Temporal server without a full Kubernetes deployment. Install from `https://docs.temporal.io/cli` or through the starter project's Docker Compose stack, which runs the Temporal development server as a service.

Verify with `temporal --version`. The Temporal development server stores workflow history in memory and resets on restart, which is appropriate for local development. It does not require a separate database. For staging and production, a persistent Temporal deployment with its own PostgreSQL store is required.

---

#### 4.2. Clone and Run the Starter Project

##### 4.2.1. Repository layout of the starter

Clone the starter project:

```bash
git clone https://github.com/awolabs/awo-starter.git myproject
cd myproject
```

The starter project layout:

```
myproject/
├── cmd/
│   └── server/
│       └── main.go          # Awo bootstrap entry point
├── modules/
│   └── .gitkeep             # Your module code goes here
├── migrations/
│   └── .gitkeep             # Versioned Atlas migration files go here
├── fixtures/
│   └── .gitkeep             # Seed data JSON files
├── .env.example             # Copy to .env and fill in values
├── docker-compose.yml       # Local dev stack
└── go.mod
```

Copy the example environment file and review it:

```bash
cp .env.example .env
```

The `.env` file contains `AWO_DB_DSN`, `AWO_REDIS_ADDR`, `AWO_TEMPORAL_HOST`, and the auth secret variables. The defaults in `.env.example` are correct for the Docker Compose stack.

##### 4.2.2. `docker compose up` — PostgreSQL, Redis, Temporal

Start the local dependency stack:

```bash
docker compose up -d
```

This starts four services: `postgres` (PostgreSQL 16 with the `awo_dev` database pre-created), `redis` (Redis 7 in standalone mode), `temporal` (Temporal development server), and `temporal-ui` (the Temporal Web UI on port 8080). Wait for the health checks to pass — about ten seconds on a typical machine — before proceeding. You can check readiness with `docker compose ps`; all four services should show status `healthy`.

The `postgres` service runs with `POSTGRES_DB=awo_dev`, `POSTGRES_USER=awo`, and `POSTGRES_PASSWORD=awo_dev_password`. These credentials match the `AWO_DB_DSN` in `.env.example`. Do not use these credentials outside of a local development environment.

##### 4.2.3. `awo serve` — first run

With the stack running and `.env` loaded, start the Awo server:

```bash
source .env && awo serve
```

On first run, `awo serve` runs pending migrations (including the framework's own schema bootstrap migrations that create the `public.tenants` table and the system schema), creates a default development tenant (`dev.awo.app`), seeds the system roles and default configuration, and starts the Fiber HTTP server on port 3000 and the Temporal worker.

The output should end with:

```
INFO  awo: EntityRegistry populated  system_entities=47 custom_entities=0
INFO  awo: Temporal worker started   task_queues=[awo.system]
INFO  awo: Fiber listening           addr=:3000
```

Verify the server is responding:

```bash
curl -s http://localhost:3000/health/live | jq .
# {"status":"ok"}
```

---

#### 4.3. Define a System Entity

##### 4.3.1. Run `awo entity create --type=system`

The `awo entity create` command scaffolds the files for a new system entity. Run it from the project root:

```bash
awo entity create --name=ServiceRequest --type=system --module=fieldservice
```

This creates the following files:

```
modules/fieldservice/
├── entity_service_request.go      # EntityDefinition declaration
├── hooks_service_request.go       # Hook function stubs
├── policy_service_request.go      # Privacy policy stub
├── page_service_request.go        # Page builder function stubs
└── fieldservice.go                # Module Register function (created once)
```

It also creates a stub migration file in `migrations/` with the naming convention `{timestamp}_create_service_requests.sql`. The migration file is empty at this point; it will be populated by `awo entity migrate --diff` after you add fields.

##### 4.3.2. Scaffold structure — schema file, repository interface, hook stubs

The generated `entity_service_request.go` contains a skeleton `EntityDefinition` with no fields or edges yet:

```go
// Example: Generated EntityDefinition scaffold
package fieldservice

import "github.com/awolabs/awo/pkg/entity"

var ServiceRequestDefinition = entity.Define("ServiceRequest")
```

The generated `hooks_service_request.go` contains stub implementations of the standard hook interfaces, all returning `nil`. The generated `policy_service_request.go` contains a `TenantIsolation` policy applied by default — the minimum safe policy for any new entity. The generated `page_service_request.go` contains empty builder functions for the list and form pages.

Register the module in `cmd/server/main.go`:

```go
// Example: Registering a module in the bootstrap entry point
package main

import (
    "github.com/awolabs/awo/pkg/entity"
    "myproject/modules/fieldservice"
)

func main() {
    app := entity.NewApp()
    app.RegisterModule(&fieldservice.Module{})
    app.Run()
}
```

##### 4.3.3. Add fields to the schema

Open `modules/fieldservice/entity_service_request.go` and add the fields:

```go
// Example: ServiceRequest EntityDefinition with fields
package fieldservice

import "github.com/awolabs/awo/pkg/entity"

var ServiceRequestDefinition = entity.Define("ServiceRequest",
    entity.Fields(
        entity.Field("request_number").
            Type(entity.Data).
            MaxLen(20).
            Immutable().
            Required(),
        entity.Field("customer_name").
            Type(entity.Data).
            MaxLen(100).
            Required(),
        entity.Field("vehicle_registration").
            Type(entity.Data).
            MaxLen(20).
            Required(),
        entity.Field("description").
            Type(entity.LongText),
        entity.Field("status").
            Type(entity.Select).
            Options("open", "in_progress", "completed", "cancelled").
            Default("open").
            Required(),
        entity.Field("total_cost").
            Type(entity.Currency),
        entity.Field("requested_at").
            Type(entity.DateTime).
            Immutable().
            Default(entity.Now),
        entity.Field("completed_at").
            Type(entity.DateTime),
    ),
    entity.NamingSeries(
        entity.Series("SR-{YYYY}-{MM}-{SEQ}").
            SequenceName("service_request_seq").
            ResetPolicy(entity.ResetMonthly).
            PaddedLength(4),
    ),
)
```

The `Immutable()` constraint on `request_number` and `requested_at` means these fields are set on create and rejected on update. The `Currency` type for `total_cost` maps to a `numeric(20,4)` column. The naming series will generate values like `SR-2025-06-0001`.

##### 4.3.4. Declare an edge to an existing entity

Add an edge to the built-in `User` entity to record which technician is assigned:

```go
// Example: Adding an edge to the ServiceRequest definition
var ServiceRequestDefinition = entity.Define("ServiceRequest",
    entity.Fields( /* as above */ ),
    entity.Edges(
        entity.Edge("assigned_technician").
            To("User").
            Optional(),
    ),
    entity.NamingSeries( /* as above */ ),
)
```

This edge generates an `assigned_technician_id` UUID foreign key column in the `service_requests` table, with an index and a foreign key constraint referencing the `users` table.

---

#### 4.4. Generate and Apply the Migration

##### 4.4.1. `awo entity migrate --dry-run` — preview the SQL

Before writing any migration file, preview what Atlas will generate:

```bash
awo entity migrate --dry-run
```

The output shows the SQL Atlas would generate to bring the tenant schema in line with the current `EntityDefinition`. For the `ServiceRequest` entity this will include a `CREATE TABLE service_requests (...)` statement with columns for each declared field, the appropriate PostgreSQL types (`text`, `numeric(20,4)`, `timestamptz`, `uuid`), and the foreign key constraint for `assigned_technician_id`.

Review the output carefully before proceeding. The `--dry-run` flag makes no changes to the filesystem or the database.

##### 4.4.2. Review the generated Atlas migration file

Generate the versioned migration file:

```bash
awo entity migrate --diff --name=create_service_requests
```

This creates a file in `migrations/` with a name like `20250601120000_create_service_requests.sql`. Open it and review the generated SQL:

```sql
-- Example: Generated Atlas migration for ServiceRequest
-- atlas:sum h1:...checksum...

-- Create sequence for naming series
CREATE SEQUENCE IF NOT EXISTS service_request_seq START 1;

-- Create table
CREATE TABLE "service_requests" (
    "id"                        uuid          NOT NULL DEFAULT gen_random_uuid(),
    "request_number"            text          NOT NULL,
    "customer_name"             varchar(100)  NOT NULL,
    "vehicle_registration"      varchar(20)   NOT NULL,
    "description"               text,
    "status"                    text          NOT NULL DEFAULT 'open',
    "total_cost"                numeric(20,4),
    "requested_at"              timestamptz   NOT NULL DEFAULT now(),
    "completed_at"              timestamptz,
    "assigned_technician_id"    uuid,
    "created_at"                timestamptz   NOT NULL DEFAULT now(),
    "updated_at"                timestamptz   NOT NULL DEFAULT now(),
    "created_by"                uuid,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_service_requests_assigned_technician"
        FOREIGN KEY ("assigned_technician_id")
        REFERENCES "users" ("id")
        ON DELETE SET NULL
);

CREATE INDEX "idx_service_requests_status" ON "service_requests" ("status");
CREATE INDEX "idx_service_requests_assigned_technician_id"
    ON "service_requests" ("assigned_technician_id");
CREATE UNIQUE INDEX "idx_service_requests_request_number"
    ON "service_requests" ("request_number");
```

The framework automatically adds `created_at`, `updated_at`, and `created_by` columns to every system entity. The `request_number` unique index enforces the naming series uniqueness. The `status` index supports the common filter-by-status query.

##### 4.4.3. `awo entity migrate --apply` — execute

Apply the migration to the development tenant schema:

```bash
awo entity migrate --apply
```

Atlas connects to the `awo_dev` database, sets the search path to `t_dev` (the development tenant schema), verifies that the migration file's checksum has not been tampered with, and executes the SQL. Successful output:

```
Migrating tenant "dev" (schema: t_dev)
  -- migrating version 20250601120000
     -> CREATE SEQUENCE IF NOT EXISTS service_request_seq START 1;
     -> CREATE TABLE "service_requests" ...
     -> CREATE INDEX ...
  -- ok (47ms)

Applied 1 migration to 1 tenant (0 errors).
```

---

#### 4.5. Wire an API Route

##### 4.5.1. Auto-generated CRUD routes from EntityDefinition

Registering `ServiceRequestDefinition` with the framework automatically creates the following routes on the Fiber server. No additional code is required:

```
GET    /api/v1/service-requests              list, with filter/sort/pagination
GET    /api/v1/service-requests/:id          single record
POST   /api/v1/service-requests              create
PATCH  /api/v1/service-requests/:id          update
DELETE /api/v1/service-requests/:id          delete
POST   /api/v1/service-requests/:id/submit   lifecycle: submit
POST   /api/v1/service-requests/:id/cancel   lifecycle: cancel
```

The entity name `ServiceRequest` maps to the URL slug `service-requests` by PascalCase-to-kebab-case conversion. The `submit` and `cancel` action routes are generated for all entities and trigger the `on_submit` and `on_cancel` hooks respectively.

##### 4.5.2. Add a custom action route

Custom action routes are declared in the `EntityDefinition` using `entity.Action`. Add a `complete` action that sets the status to `completed` and records the completion timestamp:

```go
// Example: Declaring a custom action on the EntityDefinition
var ServiceRequestDefinition = entity.Define("ServiceRequest",
    entity.Fields( /* ... */ ),
    entity.Edges( /* ... */ ),
    entity.Actions(
        entity.Action("complete").
            Method(entity.POST).
            Handler(CompleteServiceRequest),
    ),
)

func CompleteServiceRequest(ctx entity.ActionContext) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    repo, err := entity.Resolve(ctx, tc, "ServiceRequest")
    if err != nil {
        return err
    }
    record, err := repo.Get(ctx, ctx.RecordID())
    if err != nil {
        return err
    }
    if record.Get("status") == "completed" {
        return entity.NewBusinessRuleError("already_completed", "service request is already completed")
    }
    _, err = repo.Update(ctx, ctx.RecordID(), map[string]any{
        "status":       "completed",
        "completed_at": entity.Now(),
    })
    return err
}
```

This generates the route `POST /api/v1/service-requests/:id/complete`.

##### 4.5.3. Test the endpoint with curl

Create a service request. The `X-Awo-Tenant` header identifies the development tenant:

```bash
curl -s -X POST http://localhost:3000/api/v1/service-requests \
  -H "Content-Type: application/json" \
  -H "X-Awo-Tenant: dev" \
  -H "Cookie: awo_session=<your-dev-session-cookie>" \
  -d '{
    "customer_name": "John Kamau",
    "vehicle_registration": "KCC 123A",
    "description": "Oil change and tyre rotation",
    "total_cost": "4500.00"
  }' | jq .
```

A successful response:

```json
{
  "data": {
    "id": "01923f4a-...",
    "request_number": "SR-2025-06-0001",
    "customer_name": "John Kamau",
    "vehicle_registration": "KCC 123A",
    "description": "Oil change and tyre rotation",
    "status": "open",
    "total_cost": "4500.0000",
    "requested_at": "2025-06-01T09:14:33Z",
    "completed_at": null,
    "assigned_technician_id": null,
    "created_at": "2025-06-01T09:14:33Z",
    "updated_at": "2025-06-01T09:14:33Z"
  },
  "meta": {}
}
```

The `request_number` was generated automatically by the naming series. The `status` defaulted to `open`. The `total_cost` is returned as a string in `numeric(20,4)` precision.

---

#### 4.6. Emit an amis Page Definition

##### 4.6.1. Register a page builder function

Open the generated `modules/fieldservice/page_service_request.go` and implement the list page builder:

```go
// Example: amis page builder for ServiceRequest list
package fieldservice

import (
    "github.com/awolabs/awo/pkg/entity"
    "github.com/awolabs/awo/pkg/permission"
)

func BuildServiceRequestListPage(ctx entity.PageContext) (entity.PageDefinition, error) {
    perms := permission.FromContext(ctx)

    columns := []entity.AmisColumn{
        {Name: "request_number", Label: "Request No.", Sortable: true, Width: 140},
        {Name: "customer_name", Label: "Customer", Sortable: true},
        {Name: "vehicle_registration", Label: "Vehicle Reg.", Width: 120},
        {Name: "status", Label: "Status", Type: "mapping",
            Map: map[string]string{
                "open":        "<span class='label label-warning'>Open</span>",
                "in_progress": "<span class='label label-info'>In Progress</span>",
                "completed":   "<span class='label label-success'>Completed</span>",
                "cancelled":   "<span class='label label-default'>Cancelled</span>",
            },
        },
        {Name: "total_cost", Label: "Total (KES)", Type: "number",
            Precision: 2, Prefix: "KES "},
        {Name: "requested_at", Label: "Requested", Type: "datetime",
            Format: "DD/MM/YYYY HH:mm"},
    }

    page := entity.NewCRUDPage().
        Title("Service Requests").
        API("/api/v1/service-requests").
        Columns(columns).
        FilterFields(
            entity.FilterField("status", "Status", entity.FilterTypeSelect).
                Options("open", "in_progress", "completed", "cancelled"),
            entity.FilterField("customer_name", "Customer", entity.FilterTypeText),
        )

    if perms.CanCreate("ServiceRequest") {
        page.AddCreateButton("New Service Request", "/pages/service-request-form")
    }

    return page.Build(), nil
}
```

Register the page builder in the `EntityDefinition`:

```go
// Example: Adding the page builder binding to the EntityDefinition
var ServiceRequestDefinition = entity.Define("ServiceRequest",
    entity.Fields( /* ... */ ),
    entity.Edges( /* ... */ ),
    entity.Actions( /* ... */ ),
    entity.Pages(
        entity.Page("service-request-list").
            Type(entity.PageTypeList).
            Builder(BuildServiceRequestListPage),
    ),
)
```

##### 4.6.2. Visit the page in the browser

With the server running, the amis page definition is available at:

```
GET /api/v1/pages/service-request-list
X-Awo-Tenant: dev
Cookie: awo_session=<your-dev-session-cookie>
```

The response is a complete amis page JSON definition. The amis client, loaded at `http://localhost:3000/app`, renders this definition into a paginated table with filter controls and a create button. No frontend code changes were required.

To verify the page definition without the browser:

```bash
curl -s http://localhost:3000/api/v1/pages/service-request-list \
  -H "X-Awo-Tenant: dev" \
  -H "Cookie: awo_session=<your-dev-session-cookie>" | jq '.type'
# "crud"
```

---

#### 4.7. Trigger a Simple Workflow

##### 4.7.1. Define a one-activity workflow

Create a new file `modules/fieldservice/workflow_service_request.go`:

```go
// Example: Simple one-activity workflow for new service request notification
package fieldservice

import (
    "context"
    "fmt"
    "time"

    "github.com/awolabs/awo/pkg/entity"
    "github.com/awolabs/awo/pkg/tenant"
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/workflow"
)

// ServiceRequestCreatedInput is the input passed to the workflow at start.
type ServiceRequestCreatedInput struct {
    TenantSlug      string
    ServiceRequestID string
}

// ServiceRequestCreatedWorkflow is the durable workflow triggered on new service requests.
func ServiceRequestCreatedWorkflow(ctx workflow.Context, input ServiceRequestCreatedInput) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            MaxAttempts: 3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    return workflow.ExecuteActivity(ctx, NotifyManagerActivity, input).Get(ctx, nil)
}

// Activities holds the dependencies injected into activity functions.
type Activities struct {
    Repo entity.EntityRepository
}

// NotifyManagerActivity fetches the service request and logs a notification.
// In production this would send an SMS via Africa's Talking or push a notification.
func (a *Activities) NotifyManagerActivity(ctx context.Context, input ServiceRequestCreatedInput) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return fmt.Errorf("resolving tenant context: %w", err)
    }
    _ = tc

    record, err := a.Repo.Get(ctx, input.ServiceRequestID)
    if err != nil {
        return fmt.Errorf("fetching service request %s: %w", input.ServiceRequestID, err)
    }

    activity.GetLogger(ctx).Info("new service request received",
        "request_number", record.Get("request_number"),
        "customer_name", record.Get("customer_name"),
        "vehicle_registration", record.Get("vehicle_registration"),
    )

    return nil
}
```

##### 4.7.2. Bind it to the entity's on_submit hook

> **Note:** For this quick start, the workflow is triggered on `after_save` (i.e., on every new record creation) rather than on `on_submit`, since the entity does not have a formal submission flow yet. In a production entity, use `on_submit` to trigger workflows only when a document is formally finalised.

Add the workflow trigger binding to the `EntityDefinition`:

```go
// Example: Binding the workflow trigger to the EntityDefinition
var ServiceRequestDefinition = entity.Define("ServiceRequest",
    entity.Fields( /* ... */ ),
    entity.Edges( /* ... */ ),
    entity.Actions( /* ... */ ),
    entity.Pages( /* ... */ ),
    entity.WorkflowTriggers(
        entity.OnCreate(
            ServiceRequestCreatedWorkflow,
            entity.TaskQueue("fieldservice.service-request.created"),
            entity.WorkflowIDPattern("{tenant}.service-request.{id}.created"),
            entity.BuildInput(func(tc tenant.TenantContext, rec entity.EntityRecord) any {
                return ServiceRequestCreatedInput{
                    TenantSlug:       tc.Slug(),
                    ServiceRequestID: rec.ID(),
                }
            }),
        ),
    ),
)
```

Register the workflow and activities on the worker in `modules/fieldservice/fieldservice.go`:

```go
// Example: Registering workflows and activities in the module
package fieldservice

import (
    "github.com/awolabs/awo/pkg/entity"
    "go.temporal.io/sdk/worker"
)

type Module struct{}

func (m *Module) Register(app entity.App) error {
    if err := app.RegisterEntity(ServiceRequestDefinition); err != nil {
        return err
    }
    acts := &Activities{Repo: app.Repository("ServiceRequest")}
    app.RegisterWorkflow("fieldservice.service-request.created",
        ServiceRequestCreatedWorkflow,
        worker.RegisterOptions{},
    )
    app.RegisterActivity("fieldservice.service-request.created",
        acts,
        worker.RegisterOptions{},
    )
    return nil
}
```

##### 4.7.3. Watch it run in the Temporal Web UI

Restart the server to pick up the new workflow registration:

```bash
awo serve
```

Create a new service request via the API (§4.5.3). Then open the Temporal Web UI at `http://localhost:8080`. Navigate to the default namespace, and under **Workflows** you will see a new workflow with ID matching the pattern `dev.service-request.{uuid}.created` in **Completed** status.

Click the workflow to see the event history. You will see a `WorkflowExecutionStarted` event, a `ActivityTaskScheduled` event for `NotifyManagerActivity`, an `ActivityTaskStarted` event, an `ActivityTaskCompleted` event, and a `WorkflowExecutionCompleted` event. The activity log output is visible in the worker's stdout:

```
INFO  fieldservice: new service request received
      request_number=SR-2025-06-0001
      customer_name=John Kamau
      vehicle_registration=KCC 123A
```

The complete loop — HTTP request → field validation → database persist → hook execution → Temporal workflow trigger → activity execution — is now demonstrated end to end. Every subsequent entity you build in Awo follows this same structural pattern.

> **Note:** The Temporal development server stores workflow history in memory and resets on restart. Completed workflow histories will disappear after `awo serve` is restarted. This is expected behaviour in development; production deployments use a persistent Temporal server with its own PostgreSQL store.

---

#### Chapter summary

Chapter 4 walked through the complete framework loop: installing tooling (§4.1), starting the local stack (§4.2), defining a system entity with fields, edges, naming series, and a custom action (§4.3), generating and applying a reviewed Atlas migration (§4.4), verifying the auto-generated API routes and testing them with curl (§4.5), registering and visiting an amis page definition (§4.6), and binding a Temporal workflow trigger with a one-activity implementation (§4.7). The three most important concepts demonstrated are the EntityDefinition-to-migration pipeline (§4.3 and §4.4, which enforces reviewed schema changes), the automatic API route generation from EntityDefinition bindings (§4.5.1, which eliminates boilerplate route handler code), and the workflow trigger binding pattern (§4.7.2, which connects the persistence layer to durable asynchronous processing).

**Next chapters to read:**

- §5 — Field System (the complete field type and constraint reference; §4.3.3 used a subset of field types and you will need the full reference for production entity definitions)
- §7 — The EntityRecord Lifecycle (the complete hook execution model; §4.3.4 and §4.7.2 showed basic hook and trigger registration, but production hooks require understanding of transaction boundaries, hook chaining, and error propagation)
- §21 — Server-Driven UI Philosophy (the complete page builder pipeline; §4.6 showed a minimal list page, but production pages involve form layouts, conditional field visibility, permission-gated actions, and caching behaviour)
- §27 — Defining Workflows (the complete Temporal workflow reference; §4.7 showed a one-activity workflow, but production workflows require understanding of retry policies, versioning, signals, and saga compensation)
