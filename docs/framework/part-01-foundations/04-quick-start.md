# Quick Start — Your First EntityDefinition in 30 Minutes

## 4.1 Prerequisites and Tooling

### 4.1.1 Go 1.22 or later

Awo requires Go 1.22 or later. Earlier versions are not supported because Awo's code generation tools use language features introduced in 1.22. Verify your Go version:

```shell
go version
# Expected: go version go1.22.x or later
```

If you need to install or upgrade Go, use the official installer at `go.dev/dl` or a version manager such as `goenv`.

### 4.1.2 Docker Compose for local dependencies

Awo's local development stack requires PostgreSQL, Redis, and Temporal. All three are provided via a Docker Compose file that ships with the starter project. Verify Docker Engine and Docker Compose are installed:

```shell
docker --version
# Docker version 25.0.0 or later

docker compose version
# Docker Compose version v2.24.0 or later
```

Docker Desktop includes both components on macOS and Windows. On Linux, install Docker Engine and the Compose plugin separately following the official Docker documentation for your distribution.

### 4.1.3 Awo CLI installation

The Awo CLI provides code generation, migration, and development server commands. Install it with:

```shell
go install awo.so/cmd/awo@latest
awo version
```

The CLI is a compiled Go binary with no runtime dependencies beyond the Go standard library. It does not require a running Awo process to execute code generation or migration commands.

### 4.1.4 Atlas CLI installation

Atlas CLI is the database migration tool Awo uses for schema diffing, migration generation, and migration application. Awo's `awo entity migrate` command delegates to Atlas under the hood, but you should also have Atlas installed directly for advanced migration operations:

```shell
# macOS / Linux
curl -sSf https://atlasgo.sh | sh

atlas version
# atlas version v0.25.0 or later
```

### 4.1.5 Temporal CLI installation

The Temporal CLI provides a development server for local workflow execution and a command-line interface for inspecting workflow history:

```shell
# macOS via Homebrew
brew install temporal

# Linux
curl -sSf https://temporal.download/cli.sh | sh

temporal --version
# temporal version 0.13.0 or later
```

In local development, the Temporal development server runs as part of the Docker Compose stack. The CLI is needed for querying workflow history, signaling workflows, and terminating stuck workflows during development.

---

## 4.2 Clone and Run the Starter Project

### 4.2.1 Repository layout of the starter

The Awo starter project provides the minimum viable structure for a new Awo application:

```
my-awo-app/
├── cmd/
│   ├── server/         # API + Temporal worker process entrypoint
│   └── migrate/        # Migration runner (separate process for CI safety)
├── internal/
│   ├── modules/        # Your module registrations go here
│   │   └── registry.go # Calls Register() for each module
│   └── config/         # Typed configuration struct + loader
├── db/
│   └── migrations/     # Versioned Atlas migration files
├── docker-compose.yml  # Local dev: PostgreSQL, Redis, Temporal
├── .env.example        # Environment variable template
└── go.mod              # Module: your-org/your-app, requires awo.so
```

The starter project does not contain any ERP-specific modules. It provides the bootstrap code (main function, dependency wiring, graceful shutdown) and the empty module registry where you register your modules.

### 4.2.2 docker compose up — PostgreSQL, Redis, Temporal

Start the local development infrastructure:

```shell
# Copy the environment template.
cp .env.example .env

# Start all infrastructure services in the background.
docker compose up -d

# Verify all services are healthy.
docker compose ps
# NAME         STATUS          PORTS
# postgres     healthy         0.0.0.0:5432->5432/tcp
# redis        healthy         0.0.0.0:6379->6379/tcp
# temporal     healthy         0.0.0.0:7233->7233/tcp
# temporal-ui  running         0.0.0.0:8080->8080/tcp

# Wait for PostgreSQL to finish initializing (first run takes ~10 seconds).
docker compose logs postgres | grep "ready to accept connections"
```

The Temporal Web UI is available at `http://localhost:8080` and provides a dashboard for inspecting workflow execution history, querying running workflows, and viewing worker status.

### 4.2.3 awo serve — first run

Run the development server:

```shell
# Apply any pending migrations to the local database.
awo entity migrate --apply

# Start the Awo development server with hot reload on Go file changes.
awo serve --env=development --port=8080
```

The development server starts the Fiber HTTP server and the Temporal worker in the same process. Hot reload is implemented via process restart on file changes; there is no incremental reload. On first start, you should see log output confirming that all system entity routes have been registered and the Temporal worker has connected.

```
INFO  server starting  {"port": 8080, "env": "development"}
INFO  entity registered  {"name": "tenant", "routes": ["GET /api/v1/entities/tenant/:id", ...]}
INFO  temporal worker connected  {"task_queue": "default", "namespace": "default"}
INFO  server ready  {"addr": "0.0.0.0:8080"}
```

---

## 4.3 Define a System Entity

### 4.3.1 Run awo entity create --type=system

The Awo CLI scaffolds the boilerplate for a new system entity:

```shell
awo entity create \
  --type=system \
  --module=crm \
  --name=contact \
  --label="Contact" \
  --label-plural="Contacts"
```

This command creates the following files:

```
internal/modules/crm/
├── entity_contact.go       # EntityDefinition declaration
├── entity_contact_hooks.go # Hook interface stubs
└── entity_contact_test.go  # Hook unit test stubs
```

And appends a migration stub to `db/migrations/` for the new table.

### 4.3.2 Scaffold structure — schema file, repository interface, hook stubs

The generated `entity_contact.go` contains the EntityDefinition with placeholder fields:

```go
package crm

import "awo.so/internal/entity"

// ContactDefinition declares the Contact system entity.
// Add fields, edges, hooks, and permission bindings here.
var ContactDefinition = entity.SystemDefinition{
    Name:        "contact",
    Module:      "crm",
    Label:       "Contact",
    LabelPlural: "Contacts",
    Fields:      []entity.FieldDef{
        // TODO: Add fields here.
    },
    Edges: []entity.EdgeDef{},
    Hooks: entity.HookSet{},
    Permissions: entity.PermissionSet{
        Create: []string{"role:tenant.admin"},
        Read:   []string{"role:tenant.admin"},
        Write:  []string{"role:tenant.admin"},
        Delete: []string{"role:tenant.admin"},
    },
}
```

### 4.3.3 Add fields to the schema

Replace the placeholder `Fields` slice with your actual field declarations:

```go
Fields: []entity.FieldDef{
    {Name: "first_name",   Type: entity.FieldData, Required: true, MaxLen: 100},
    {Name: "last_name",    Type: entity.FieldData, Required: true, MaxLen: 100},
    {Name: "email",        Type: entity.FieldData,
        Validators: []entity.FieldValidator{entity.ValidateEmail},
        Unique: true},
    {Name: "phone",        Type: entity.FieldData,
        Validators: []entity.FieldValidator{entity.ValidateE164Phone}},
    {Name: "company_name", Type: entity.FieldData, MaxLen: 200},
    {Name: "status",       Type: entity.FieldSelect,
        Options: []string{"Active", "Inactive", "Prospect"},
        Default: "Prospect"},
    {Name: "notes",        Type: entity.FieldLongText},
    {Name: "created_by",   Type: entity.FieldLink, LinkTarget: "user", Immutable: true},
},
```

### 4.3.4 Declare an edge to an existing entity

Contacts may belong to a customer (a many-to-one relationship). Declare the edge:

```go
Edges: []entity.EdgeDef{
    {
        Name:      "customer",
        Target:    "customer",
        Type:      entity.EdgeManyToOne,
        // The FK column on the contact table is named "customer_id".
        // This column is automatically added to the migration.
    },
},
```

---

## 4.4 Generate and Apply the Migration

### 4.4.1 awo entity migrate --dry-run — preview the SQL

Before applying any migration to a database, preview the SQL that will be generated:

```shell
awo entity migrate --dry-run --env=development
```

Output:

```sql
-- Dry run: no changes applied.

-- New migration: 20241215143022_create_contact.sql

CREATE TABLE "contact" (
    "id"           uuid         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    "first_name"   varchar(100) NOT NULL,
    "last_name"    varchar(100) NOT NULL,
    "email"        varchar(255) UNIQUE,
    "phone"        varchar(50),
    "company_name" varchar(200),
    "status"       varchar(20)  NOT NULL DEFAULT 'Prospect'
                                CHECK (status IN ('Active', 'Inactive', 'Prospect')),
    "notes"        text,
    "customer_id"  uuid REFERENCES "customer" ("id"),
    "created_by"   uuid NOT NULL REFERENCES "user" ("id"),
    "created_at"   timestamptz  NOT NULL DEFAULT now(),
    "updated_at"   timestamptz  NOT NULL DEFAULT now(),
    "deleted_at"   timestamptz
);

CREATE INDEX "contact_email_idx" ON "contact" ("email");
CREATE INDEX "contact_customer_id_idx" ON "contact" ("customer_id");
CREATE INDEX "contact_deleted_at_idx" ON "contact" ("deleted_at");
```

Review this SQL before proceeding. Check that: column types match your intent, the CHECK constraint on the `status` column lists all declared options, foreign keys reference the correct tables, and indexes are created for all link fields.

### 4.4.2 Review the generated Atlas migration file

After the dry run, generate the actual migration file:

```shell
awo entity migrate --diff --name="create_contact"
```

This creates `db/migrations/20241215143022_create_contact.sql` with the up-migration SQL and a corresponding `--down-` section for rollback. Review the generated file before committing it to version control. The migration checksum is computed and stored in the Atlas `atlas_schema_revisions` table when applied — any modification to the migration file after it has been applied will be detected as drift.

### 4.4.3 awo entity migrate --apply — execute

Apply pending migrations to the development database:

```shell
awo entity migrate --apply --env=development
```

Output:

```
Applying migration: 20241215143022_create_contact.sql
  ✓ CREATE TABLE "contact"
  ✓ CREATE INDEX "contact_email_idx"
  ✓ CREATE INDEX "contact_customer_id_idx"
Migration applied successfully.
```

After applying the migration, restart the development server. The `contact` entity routes will be registered and available.

---

## 4.5 Wire an API Route

### 4.5.1 Auto-generated CRUD routes from EntityDefinition

No additional code is required to make the `contact` entity accessible via the API. Once the EntityDefinition is registered and the migration is applied, the following routes are automatically available:

```
GET    /api/v1/entities/contact           List contacts (with filter, sort, pagination)
GET    /api/v1/entities/contact/:id       Get a single contact by ID
POST   /api/v1/entities/contact           Create a contact
PATCH  /api/v1/entities/contact/:id       Partial update a contact
DELETE /api/v1/entities/contact/:id       Delete a contact
```

All routes are guarded by the permission policy declared in the EntityDefinition. An unauthenticated request returns `401 Unauthorized`. A request from a user without `role:tenant.admin` returns `403 Forbidden`.

### 4.5.2 Add a custom action route

Custom actions are entity-specific operations beyond CRUD. Declare a custom action on the EntityDefinition to add a route:

```go
// In entity_contact.go, add to ContactDefinition:
Actions: []entity.ActionDef{
    {
        Name:        "convert_to_customer",
        Method:      entity.ActionMethodPost,
        Label:       "Convert to Customer",
        Permission:  "role:tenant.admin",
        HandlerFunc: ConvertContactToCustomer,
    },
},
```

Implement the handler function:

```go
// entity_contact_actions.go
package crm

import (
    "context"
    "fmt"
    "awo.so/internal/entity"
)

// ConvertContactToCustomer creates a Customer record from a Contact record.
// Called by the auto-generated route: POST /api/v1/entities/contact/:id/convert_to_customer
func ConvertContactToCustomer(ctx context.Context, action entity.ActionContext) (*entity.ActionResult, error) {
    contact, err := action.Repo.Get(ctx, "contact", action.RecordID)
    if err != nil {
        return nil, fmt.Errorf("convert_to_customer: get contact: %w", err)
    }

    var customerID string
    if err := action.Repo.WithTx(ctx, func(tx entity.EntityRepository) error {
        customer, err := tx.Create(ctx, "customer", map[string]any{
            "name":    fmt.Sprintf("%s %s", contact.Fields["first_name"], contact.Fields["last_name"]),
            "email":   contact.Fields["email"],
            "phone":   contact.Fields["phone"],
            "status":  "Active",
        })
        if err != nil {
            return fmt.Errorf("create customer: %w", err)
        }
        customerID = customer.ID

        if _, err := tx.Update(ctx, "contact", action.RecordID, map[string]any{
            "status":      "Active",
            "customer_id": customerID,
        }); err != nil {
            return fmt.Errorf("link contact to customer: %w", err)
        }

        return nil
    }); err != nil {
        return nil, err
    }

    return &entity.ActionResult{
        Message: fmt.Sprintf("Contact converted to Customer %s", customerID),
        Data:    map[string]any{"customer_id": customerID},
    }, nil
}
```

### 4.5.3 Test the endpoint with curl

Create a contact and test the custom action:

```shell
# Create a contact.
curl -X POST http://localhost:8080/api/v1/entities/contact \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: your-tenant-uuid" \
  -H "Cookie: awo_session=your-session-token" \
  -d '{
    "first_name": "Amina",
    "last_name": "Ochieng",
    "email": "amina@example.co.ke",
    "phone": "+254712345678",
    "status": "Prospect"
  }'

# Response:
# {"data": {"id": "...", "first_name": "Amina", "last_name": "Ochieng", ...}, "meta": {...}}

# Convert the contact to a customer.
curl -X POST http://localhost:8080/api/v1/entities/contact/CONTACT-ID/convert_to_customer \
  -H "X-Tenant-ID: your-tenant-uuid" \
  -H "Cookie: awo_session=your-session-token"

# Response:
# {"data": {"message": "Contact converted to Customer ...", "customer_id": "..."}, "meta": {...}}
```

---

## 4.6 Emit an amis Page Definition

### 4.6.1 Register a page builder function

The framework generates a default amis list + form page for every EntityDefinition. To register a custom page builder that overrides the default detail view:

```go
// entity_contact.go — update ContactDefinition:
PageBuilders: entity.PageBuilderSet{
    Detail: BuildContactDetailPage,
},
```

Implement the page builder function:

```go
// entity_contact_pages.go
package crm

import (
    "awo.so/internal/entity"
    "awo.so/internal/sdui"
)

// BuildContactDetailPage builds the amis JSON schema for the contact detail view.
// Receives the resolved permissions and feature flags for the current user.
func BuildContactDetailPage(ctx sdui.PageContext) (sdui.PageSchema, error) {
    return sdui.PageSchema{
        Type:  "page",
        Title: "Contact",
        Body: sdui.Container{
            Type: "grid",
            Columns: []sdui.Column{
                {
                    MD: 8,
                    Body: sdui.DetailForm{
                        Type: "detail",
                        API:  "/api/v1/entities/contact/${id}",
                        Body: []sdui.Field{
                            {Name: "first_name", Label: "First Name"},
                            {Name: "last_name",  Label: "Last Name"},
                            {Name: "email",      Label: "Email"},
                            {Name: "phone",      Label: "Phone"},
                            {Name: "status",     Label: "Status"},
                        },
                    },
                },
                {
                    MD: 4,
                    Body: sdui.ActionPanel{
                        Type: "panel",
                        Title: "Actions",
                        Body: []sdui.Action{
                            {
                                Type:    "button",
                                Label:   "Convert to Customer",
                                API:     "POST /api/v1/entities/contact/${id}/convert_to_customer",
                                Visible: ctx.Permissions.CanEdit("contact"),
                            },
                        },
                    },
                },
            },
        },
    }, nil
}
```

### 4.6.2 Visit the page in the browser

The framework serves the page schema at `/api/v1/pages/contact-detail`. The amis browser client fetches this URL when navigating to the contact detail page and renders the returned JSON schema. No browser reload is needed after changing the page builder function — the next page navigation fetches the updated schema from the API.

During development, the page schema cache is bypassed by appending `?cache=false` to the page URL. In production, the cache TTL is 5 minutes and is invalidated on permission or feature flag changes.

---

## 4.7 Trigger a Simple Workflow

### 4.7.1 Define a one-activity workflow

Create a Temporal workflow that sends a welcome notification when a contact is created:

```go
// internal/modules/crm/workflows/contact_welcome.go
package workflows

import (
    "fmt"
    "time"

    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"
    "awo.so/internal/modules/crm/notifications"
)

type ContactWelcomeInput struct {
    TenantID  string
    ContactID string
    Email     string
    FirstName string
}

// ContactWelcomeWorkflow sends a welcome notification after contact creation.
func ContactWelcomeWorkflow(ctx workflow.Context, input ContactWelcomeInput) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
            InitialInterval: 2 * time.Second,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    if err := workflow.ExecuteActivity(ctx, notifications.SendWelcomeEmailActivity, input).Get(ctx, nil); err != nil {
        return fmt.Errorf("send welcome email: %w", err)
    }
    return nil
}

// SendWelcomeEmailActivity is the Temporal activity that sends the email.
// The Activities struct pattern allows dependency injection for testability.
type Activities struct {
    EmailClient notifications.EmailClient
}

func (a *Activities) SendWelcomeEmailActivity(ctx context.Context, input ContactWelcomeInput) error {
    return a.EmailClient.Send(ctx, notifications.Email{
        To:      input.Email,
        Subject: fmt.Sprintf("Welcome, %s!", input.FirstName),
        Body:    "Thank you for joining. Your contact has been created.",
    })
}
```

### 4.7.2 Bind it to the entity's on_submit hook

Add a workflow trigger binding to the ContactDefinition:

```go
// entity_contact.go — update ContactDefinition:
WorkflowTriggers: []entity.WorkflowTrigger{
    {
        On:        entity.EventAfterCreate,
        WorkflowFn: "ContactWelcomeWorkflow",
        TaskQueue:  "crm.contact.notifications",
        InputBuilder: func(rec *entity.EntityRecord, tc entity.TriggerContext) (any, error) {
            return workflows.ContactWelcomeInput{
                TenantID:  rec.TenantID,
                ContactID: rec.ID,
                Email:     rec.Fields["email"].(string),
                FirstName: rec.Fields["first_name"].(string),
            }, nil
        },
    },
},
```

Register the workflow and activities on the Temporal worker in your module's `Register` function:

```go
// internal/modules/crm/module.go
func RegisterWorker(w worker.Worker, deps ModuleDeps) {
    w.RegisterWorkflow(workflows.ContactWelcomeWorkflow)
    acts := &workflows.Activities{EmailClient: deps.EmailClient}
    w.RegisterActivity(acts.SendWelcomeEmailActivity)
}
```

### 4.7.3 Watch it run in the Temporal Web UI

After creating a contact via the API, open the Temporal Web UI at `http://localhost:8080`. Navigate to the "Workflows" tab and filter by namespace `default`. You should see a new workflow execution with ID `{tenant}.contact.{id}.after_create`. Click the workflow to see its event history: `WorkflowExecutionStarted`, `ActivityTaskScheduled`, `ActivityTaskStarted`, `ActivityTaskCompleted`, `WorkflowExecutionCompleted`.

If the activity fails (e.g., the email service is unreachable in local development), the event history will show `ActivityTaskFailed` followed by a retry attempt. Temporal will retry automatically according to the configured retry policy.

To terminate a stuck workflow during development:

```shell
temporal workflow terminate \
  --workflow-id="tenant-uuid.contact.contact-uuid.after_create" \
  --reason="Development cleanup"
```

---

## Chapter Summary

Chapter 4 walked through the complete framework loop: installing tooling (§4.1), starting the local dependency stack (§4.2), defining a system entity with fields, edges, naming series, and a custom action (§4.3), generating and applying a reviewed Atlas migration (§4.4), verifying the auto-generated API routes (§4.5), registering an amis page builder (§4.6), and binding a Temporal workflow trigger (§4.7).

The three most important patterns demonstrated:

- **EntityDefinition → migration pipeline** (§4.3–4.4): every schema change flows through a reviewed Atlas migration file — no ad-hoc `ALTER TABLE`.
- **Auto-generated CRUD routes** (§4.5): registering an `EntityDefinition` is sufficient to generate list, get, create, update, delete, and lifecycle action routes. No boilerplate route handler code.
- **Workflow trigger binding** (§4.7.2): the `after_save` / `on_submit` trigger pattern connects synchronous persistence to durable asynchronous processing without coupling the HTTP handler to the workflow logic.

**Next chapters to read:**

- [§5 — Field System](../part-02-entity-system/05-field-system.md) — the complete field type and constraint reference; §4.3 used a subset of field types
- [§7 — Entity Record Lifecycle](../part-02-entity-system/07-entity-record-lifecycle.md) — the complete hook execution model and transaction boundaries
- [§21 — SDUI Philosophy](../part-04-sdui/21-sdui-philosophy.md) — the full page builder pipeline; §4.6 showed a minimal registration
- [§27 — Defining Workflows](../part-05-workflow/27-defining-workflows.md) — retry policies, versioning, signals, and saga compensation
