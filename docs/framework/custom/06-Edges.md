### Chapter 6 — Edges — Relationships Between EntityDefinitions

Edges declare the relationships between `EntityDefinition` objects. Where fields describe the attributes of a single entity, edges describe how entities connect to one another. An edge declaration drives four concrete outputs: the foreign key column (or junction table) in the PostgreSQL schema, the JOIN strategy available through the `EntityRepository` interface, the cascade and orphan-handling behaviour on delete, and the component hint that tells the amis page builder how to render the relationship in a form or detail view. Every relationship in Awo's data model — from a `SalesOrder` owning its `SalesOrderLine` children to a `User` belonging to multiple `Role` records — is expressed as an edge.

---

#### 6.1. Edge Fundamentals

##### 6.1.1. What an edge declaration generates — FK column, index, join method

A one-to-many edge declaration generates a foreign key `uuid` column on the many side's table, a B-tree index on that column, and a foreign key constraint referencing the one side's primary key. A many-to-many edge declaration generates a junction table with two UUID columns, each referencing one of the participating entity tables, and indexes on both columns as well as a composite unique index on the column pair.

Beyond the schema artefacts, the edge declaration registers a join method on the `EntityRepository` interface. When a caller passes `entity.Include("line_items")` to a `Query` or `Get` call, the repository performs the appropriate JOIN and populates the nested `EntityRecord` slice in the result. This is the only supported eager-loading mechanism; the repository does not perform lazy loading. Accessing a related entity without declaring an `Include` requires a separate `Query` call against the related entity's repository.

The Atlas migration generated from the `EntityDefinition` includes the full DDL for every declared edge: `ALTER TABLE`, `CREATE INDEX`, and `ADD CONSTRAINT` statements. Edges added after initial deployment require a new migration file reviewed and applied through the standard migration pipeline (§11).

##### 6.1.2. Edge direction — owner side vs inverse side

Every edge has an owner side and an inverse side. The owner side is where the foreign key column lives. For a one-to-many edge between `SalesOrder` (one) and `SalesOrderLine` (many), the owner side is `SalesOrderLine` — the FK column `sales_order_id` lives on `sales_order_lines`. For a many-to-many edge, the framework nominates the owner side based on declaration order; the junction table is named `{owner}_{inverse}` in alphabetical order.

In the `EntityDefinition`, edge direction is declared using `entity.Edge("name").To("TargetEntity")` for outgoing edges (owner side) and `entity.Edge("name").From("SourceEntity")` for incoming edges (inverse side). Both sides of an edge must be declared for the edge to be traversable in both directions. Declaring only one side creates a unidirectional edge: the FK column is still generated, but the reverse traversal is not available through the repository's `Include` mechanism.

##### 6.1.3. Edge naming conventions

Edge names use camelCase in the `EntityDefinition` declaration and are converted to snake_case for the generated FK column name. An edge named `assignedTechnician` on `ServiceRequest` generates the column `assigned_technician_id`. The edge name also becomes the key used in `entity.Include("assignedTechnician")` calls and in the nested `EntityRecord` map returned by the repository.

Edge names should be meaningful from the perspective of the entity declaring them, not from a generic relational perspective. On `SalesOrder`, the edge to `Customer` should be named `customer` (not `customerId` or `belongsToCustomer`). On `Customer`, the reverse edge to `SalesOrder` records should be named `salesOrders` (plural, reflecting the one-to-many cardinality). Consistent naming makes `Include` clauses self-documenting.

---

#### 6.2. One-to-Many Edges

One-to-many edges are the most common relationship type in ERP systems. A `SalesInvoice` has many `SalesInvoiceLine` records. A `Department` has many `Employee` records. A `Tank` has many `DipReading` records. This section covers every configuration option for this edge type.

##### 6.2.1. Declaring the edge on both sides

Both sides of a one-to-many edge must be declared explicitly. The "one" side declares a `From` edge (incoming, receiving the relationship), and the "many" side declares a `To` edge (outgoing, owning the FK column).

```go
// Example: One-to-many edge declared on both sides
// On SalesInvoice (the "one" side):
var SalesInvoiceDefinition = entity.Define("SalesInvoice",
    entity.Fields( /* ... */ ),
    entity.Edges(
        entity.Edge("lineItems").
            From("SalesInvoiceLine").
            OneToMany(),
    ),
)

// On SalesInvoiceLine (the "many" side, FK owner):
var SalesInvoiceLineDefinition = entity.Define("SalesInvoiceLine",
    entity.Fields(
        entity.Field("sales_invoice_id").
            Type(entity.Link).
            LinkedEntity("SalesInvoice").
            Required().
            Immutable(),
        entity.Field("item_code").Type(entity.Data).MaxLen(50).Required(),
        entity.Field("quantity").Type(entity.Float).Required(),
        entity.Field("unit_price").Type(entity.Currency).Required(),
        entity.Field("line_total").Type(entity.Currency).Required(),
    ),
    entity.Edges(
        entity.Edge("salesInvoice").
            To("SalesInvoice").
            Required(),
    ),
)
```

The `Link` field `sales_invoice_id` on `SalesInvoiceLine` and the `To` edge declaration are redundant in terms of schema generation — only one FK column is created — but both must be present. The `Link` field makes the FK value accessible as a typed field value in the `EntityRecord`. The edge declaration makes the join traversal available through `Include`.

##### 6.2.2. FK column placement — always on the many side

The foreign key column is always on the many side of a one-to-many relationship. This is a SQL normalisation requirement, not a framework convention. On `SalesInvoiceLine`, the column `sales_invoice_id uuid NOT NULL REFERENCES sales_invoices(id)` is generated. There is no corresponding column on `sales_invoices` — the one-to-many relationship is navigated by querying `sales_invoice_lines WHERE sales_invoice_id = ?`, not by storing an array of IDs.

This placement has implications for cascade behaviour: the FK constraint lives on the child table, so cascade and restriction rules are configured as options on the child's FK constraint.

##### 6.2.3. Eager loading vs lazy loading — performance implications

Awo does not perform lazy loading. When you call `repo.Get(ctx, id)` without an `Include` option, you receive the parent record only — no child records. To load child records, you must either include them explicitly or query them separately.

```go
// Example: Eager loading child records with Include
tc, err := tenant.FromContext(ctx)
if err != nil {
    return err
}
repo, err := entity.Resolve(ctx, tc, "SalesInvoice")
if err != nil {
    return err
}
record, err := repo.Get(ctx, invoiceID,
    entity.Include("lineItems"),
    entity.Include("customer"),
)
if err != nil {
    return err
}
lineItems := record.GetRelated("lineItems") // []entity.EntityRecord
```

Each `Include` directive adds a JOIN to the generated SQL. Including deeply nested edges — `Include("lineItems.item")` to load the `Item` for each line — adds additional JOINs. The framework limits nesting depth to three levels by default to prevent accidental cartesian product queries. For reporting use cases that require deeper joins, use a dedicated report query via the `Aggregate` method or a raw SQL view registered as a read-only system entity.

> **Warning:** Avoid including large one-to-many edges on list queries. Including `lineItems` on a `Query` that returns 100 `SalesInvoice` records will execute 100 additional queries or a single JOIN that produces a large result set. For list views, display summary fields (line count, total amount) computed in the database rather than fetching all child records.

##### 6.2.4. Cascade delete — when to use, when to guard with a before_delete hook

`CascadeDelete()` on a one-to-many edge configures the PostgreSQL foreign key with `ON DELETE CASCADE`. When the parent record is deleted, all child records are automatically deleted by the database without invoking Awo hooks on each child. Use `CascadeDelete()` when the child records have no independent existence and their deletion requires no business logic — draft document lines, custom field value records, log entries.

Never use `CascadeDelete()` when child records have financial or inventory significance. A `SalesInvoiceLine` that has been posted to the GL must not be silently deleted by a cascade. For these cases, guard the parent deletion with a `before_delete` hook (§7.5) that rejects the deletion if any financially significant child records exist.

```go
// Example: Edge with cascade delete (safe: draft line items only)
entity.Edge("draftLineItems").
    From("DraftOrderLine").
    OneToMany().
    CascadeDelete()

// Example: Edge without cascade delete (guarded by before_delete hook)
entity.Edge("invoiceLines").
    From("SalesInvoiceLine").
    OneToMany()
    // No CascadeDelete — before_delete hook checks for posted lines
```

##### 6.2.5. Orphan handling — restrict, set null, cascade

Three orphan-handling policies are available when the parent record is deleted:

`entity.OnDeleteRestrict()` — generates `ON DELETE RESTRICT`. The database prevents deletion of the parent if any child records exist. This is the safest policy for financially significant relationships. A `JournalEntry` with posted `JournalEntryLine` records cannot be deleted; the deletion must be rejected with a user-facing error from a `before_delete` hook that explains why.

`entity.OnDeleteSetNull()` — generates `ON DELETE SET NULL`. The child's FK column is set to NULL when the parent is deleted. Use this for optional relationships where the child can exist independently: a `ServiceRequest` whose `assignedTechnician` (a `User`) is deleted should retain the service request with a null technician reference rather than being deleted itself.

`entity.OnDeleteCascade()` — equivalent to `CascadeDelete()` described above. Provided as a named option for symmetry.

The default when no policy is declared is `OnDeleteRestrict`. This is intentional: the safe default prevents accidental data loss.

---

#### 6.3. Many-to-Many Edges

Many-to-many edges are less common in ERP than one-to-many edges but appear in role assignment, product category membership, and any tagging system where an entity belongs to multiple groups and each group contains multiple entities.

##### 6.3.1. Junction table generation

A many-to-many edge declaration generates a junction table named `{entity_a}_{entity_b}` (alphabetical order of entity names, snake_case). The table has two UUID foreign key columns, each referencing one of the participant tables, and a composite unique constraint on the column pair to prevent duplicate associations.

```go
// Example: Many-to-many edge between User and Role
// On User:
var UserDefinition = entity.Define("User",
    entity.Fields( /* ... */ ),
    entity.Edges(
        entity.Edge("roles").
            To("Role").
            ManyToMany(),
    ),
)

// On Role:
var RoleDefinition = entity.Define("Role",
    entity.Fields( /* ... */ ),
    entity.Edges(
        entity.Edge("users").
            From("User").
            ManyToMany(),
    ),
)
```

This generates a junction table `roles_users` with columns `user_id uuid REFERENCES users(id)` and `role_id uuid REFERENCES roles(id)`, plus `PRIMARY KEY (user_id, role_id)`. Indexes are created on both FK columns.

##### 6.3.2. Junction table annotations — adding payload fields to the relationship

Some many-to-many relationships carry data about the association itself. A `User` assigned to a `Project` may have a role within that project (`lead`, `contributor`, `reviewer`) that is distinct from their system role. This payload belongs on the junction table, not on either participant.

```go
// Example: Many-to-many edge with junction table payload
entity.Edge("projectMembers").
    To("Project").
    ManyToMany().
    JunctionFields(
        entity.Field("project_role").
            Type(entity.Select).
            Options("lead", "contributor", "reviewer").
            Required(),
        entity.Field("joined_at").
            Type(entity.DateTime).
            Default(entity.Now),
    )
```

Junction table records are accessible through the `EntityRepository` via the `GetJunction(edgeName, parentID, relatedID)` method, which returns an `EntityRecord` representing the junction row including its payload fields. Creating or updating a many-to-many association with payload fields uses `SetJunction(edgeName, parentID, relatedID, payload)`.

##### 6.3.3. Querying through many-to-many edges

Many-to-many edges are traversed using the same `Include` mechanism as one-to-many edges:

```go
// Example: Loading a user's roles through a many-to-many edge
tc, err := tenant.FromContext(ctx)
if err != nil {
    return err
}
repo, err := entity.Resolve(ctx, tc, "User")
if err != nil {
    return err
}
record, err := repo.Get(ctx, userID, entity.Include("roles"))
if err != nil {
    return err
}
roles := record.GetRelated("roles") // []entity.EntityRecord, one per role
```

Filtering on many-to-many edges — "find all users who have the `finance_manager` role" — uses the `HasEdge` filter predicate:

```go
// Example: Filtering by many-to-many edge membership
filter := entity.NewFilter().
    HasEdgeWith("roles", entity.NewFilter().Eq("name", "finance_manager"))
users, _, err := repo.Query(ctx, filter)
```

The `HasEdgeWith` predicate generates a subquery against the junction table. It is less efficient than filtering on a direct field but is the correct pattern for many-to-many membership queries.

##### 6.3.4. Performance characteristics of deep many-to-many joins

Every `Include` on a many-to-many edge adds at least two JOINs to the generated SQL: one to the junction table and one to the related entity's table. Chaining multiple many-to-many `Include` directives can produce queries with many joins that PostgreSQL may execute inefficiently.

For read-heavy reporting queries that join across many-to-many boundaries, consider materialising the result using a PostgreSQL view registered as a read-only system entity, or using the `Aggregate` method of `EntityRepository` with a custom `GroupBy` spec that lets the database compute the join once rather than per-request. For UI list views, denormalise the most commonly displayed many-to-many data into a computed field on the primary entity updated by an `after_save` hook on the junction.

---

#### 6.4. Self-Referencing Edges

Self-referencing edges model hierarchical data: an `Account` tree in the chart of accounts, an `Employee` reporting structure, a `Category` hierarchy for inventory items. Awo supports two storage strategies for hierarchies: materialised paths for deep trees and adjacency lists for shallow ones.

##### 6.4.1. Tree structures — parent/children pattern

A self-referencing edge declares both the parent (FK to the same entity) and the children (inverse, one-to-many). The FK column `parent_id` is nullable — root nodes have `NULL` in `parent_id`.

```go
// Example: Self-referencing edge for account tree structure
var AccountDefinition = entity.Define("Account",
    entity.Fields(
        entity.Field("account_name").Type(entity.Data).MaxLen(100).Required(),
        entity.Field("account_type").Type(entity.Select).
            Options("asset", "liability", "equity", "income", "expense").Required(),
        entity.Field("parent_id").Type(entity.Link).LinkedEntity("Account"),
        entity.Field("path").Type(entity.Data).MaxLen(500).Immutable(),
    ),
    entity.Edges(
        entity.Edge("parent").To("Account").Optional(),
        entity.Edge("children").From("Account").OneToMany(),
    ),
)
```

##### 6.4.2. Materialised path for deep hierarchies (account trees, org charts)

The materialised path pattern stores the full ancestry of each node as a delimited string in a `path` column: `0001.0003.0012` for a node three levels deep. The framework maintains the path automatically via a `before_save` hook that computes the path from the parent's path plus the current node's sequence position.

Materialised paths enable highly efficient subtree queries using PostgreSQL's `LIKE` operator: `WHERE path LIKE '0001.%'` returns all descendants of node `0001` in a single index scan on a B-tree index over the `path` column, without recursion. This is far more efficient than recursive CTEs for deep trees with frequent reads, which is the typical ERP chart of accounts access pattern.

```go
// Example: before_save hook maintaining materialised path
func maintainAccountPath(ctx hook.Context) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    record := ctx.Record()
    parentID, _ := record.Get("parent_id").(string)
    if parentID == "" {
        record.Set("path", record.ID())
        return nil
    }
    repo, err := entity.Resolve(ctx, tc, "Account")
    if err != nil {
        return err
    }
    parent, err := repo.Get(ctx, parentID)
    if err != nil {
        return fmt.Errorf("loading parent account: %w", err)
    }
    parentPath, _ := parent.Get("path").(string)
    record.Set("path", parentPath+"."+record.ID())
    return nil
}
```

> **Warning:** Never allow a `parent_id` to point to a descendant of the current node. This creates a cycle in the tree that makes path computation infinite and subtree queries return incorrect results. Validate acyclicity in a `before_save` hook by checking that the proposed parent's `path` does not contain the current node's ID.

##### 6.4.3. Adjacency list for shallow hierarchies

For hierarchies that are known to be shallow (two or three levels maximum) — product category groups, department structures in small organisations — the adjacency list pattern without a materialised path is simpler. Each node stores only `parent_id`; ancestor traversal uses recursive CTEs when needed.

```go
// Example: Shallow adjacency list for department hierarchy
var DepartmentDefinition = entity.Define("Department",
    entity.Fields(
        entity.Field("name").Type(entity.Data).MaxLen(100).Required(),
        entity.Field("parent_id").Type(entity.Link).LinkedEntity("Department"),
    ),
    entity.Edges(
        entity.Edge("parent").To("Department").Optional(),
        entity.Edge("subDepartments").From("Department").OneToMany(),
    ),
)
```

For a three-level hierarchy (Division → Department → Team), two `Include` directives load the full structure: `Include("subDepartments")` and `Include("subDepartments.subDepartments")`. Beyond three levels, the recursive CTE approach or materialised paths should be preferred.

##### 6.4.4. Querying ancestors and descendants efficiently

With materialised paths, ancestor and descendant queries are single SQL statements:

```go
// Example: Querying all descendants of an account using materialised path
tc, err := tenant.FromContext(ctx)
if err != nil {
    return err
}
repo, err := entity.Resolve(ctx, tc, "Account")
if err != nil {
    return err
}
// Load the target account to get its path
target, err := repo.Get(ctx, accountID)
if err != nil {
    return err
}
targetPath, _ := target.Get("path").(string)
// Query all descendants: path starts with target's path followed by "."
filter := entity.NewFilter().StartsWith("path", targetPath+".")
descendants, _, err := repo.Query(ctx, filter)
if err != nil {
    return err
}
```

For adjacency list hierarchies, the `EntityRepository` provides a `QueryDescendants(ctx, id, maxDepth)` helper that generates a recursive CTE. At depth 1 it is equivalent to `Query` with `HasEdge("parent", ...)`. At depth N it generates an `WITH RECURSIVE` query. The maximum supported recursive depth is 20; beyond this, materialised paths must be used.

---

#### 6.5. Polymorphic Relationships

Polymorphic relationships allow a single edge to point to records of different entity types. An `Attachment` that can be linked to a `PurchaseOrder`, a `SalesInvoice`, or a `ServiceRequest` is a canonical example. Rather than creating three separate FK columns (`purchase_order_id`, `sales_invoice_id`, `service_request_id`) with complex nullable logic, a `DynamicLink` field stores the entity name and the record ID as a pair.

##### 6.5.1. When to use DynamicLink vs a union of concrete Links

Use `DynamicLink` when: the set of entity types that can be linked is open-ended or expected to grow over time; the logic that operates on the linked record needs to dispatch generically (look up the entity definition, call the resolver); or the linking entity is a cross-cutting concern like attachments, comments, notifications, or audit trail entries that applies to many entity types.

Use a union of concrete `Link` fields when: the set of entity types is small and closed; each linked type has distinct business semantics that warrant separate fields; and the linking logic needs to enforce that exactly one of the FK columns is populated (mutual exclusion is easier to validate on named fields than on a type discriminator string).

##### 6.5.2. DynamicLink storage — `{field}_type` + `{field}_id` column pair

A field declared `Type(entity.DynamicLink)` with name `subject` generates two PostgreSQL columns: `subject_type text` (storing the entity name as a string, e.g. `"SalesInvoice"`) and `subject_id uuid` (storing the linked record's primary key). There is no database-level FK constraint because PostgreSQL cannot enforce a FK across dynamically determined tables.

```go
// Example: Attachment entity with a DynamicLink to any attachable entity
var AttachmentDefinition = entity.Define("Attachment",
    entity.Fields(
        entity.Field("subject").
            Type(entity.DynamicLink).
            AllowedTypes("SalesInvoice", "PurchaseOrder", "ServiceRequest", "JournalEntry").
            Required(),
        entity.Field("file_name").Type(entity.Data).MaxLen(255).Required(),
        entity.Field("file_key").Type(entity.Attach).Required(),
        entity.Field("uploaded_by").Type(entity.Link).LinkedEntity("User").Required(),
        entity.Field("uploaded_at").Type(entity.DateTime).Default(entity.Now).Immutable(),
    ),
)
```

`AllowedTypes(...)` declares which entity names are valid values for the `subject_type` column. The framework validates the `subject_type` value against this list at field validator time, and validates that a record with `subject_id` exists in the named entity's table via an async validator.

##### 6.5.3. Querying polymorphic edges

Querying all attachments for a specific record requires filtering on both columns of the `DynamicLink` pair:

```go
// Example: Querying all attachments for a SalesInvoice record
tc, err := tenant.FromContext(ctx)
if err != nil {
    return err
}
repo, err := entity.Resolve(ctx, tc, "Attachment")
if err != nil {
    return err
}
filter := entity.NewFilter().
    Eq("subject_type", "SalesInvoice").
    Eq("subject_id", invoiceID)
attachments, _, err := repo.Query(ctx, filter)
if err != nil {
    return err
}
```

A composite index on `(subject_type, subject_id)` is generated automatically for `DynamicLink` fields. This index makes the two-column filter efficient.

The inverse direction — finding the linked subject record from an `Attachment` — uses the `EntityResolver` to dispatch dynamically:

```go
// Example: Resolving the subject of a DynamicLink record
attachment, err := attachmentRepo.Get(ctx, attachmentID)
if err != nil {
    return err
}
subjectType, _ := attachment.Get("subject_type").(string)
subjectID, _ := attachment.Get("subject_id").(string)
subjectRepo, err := entity.Resolve(ctx, tc, subjectType)
if err != nil {
    return err
}
subject, err := subjectRepo.Get(ctx, subjectID)
if err != nil {
    return err
}
```

##### 6.5.4. Limitations — no FK constraint, application-layer integrity only

Because `DynamicLink` has no database-level FK constraint, three integrity risks must be managed at the application layer:

First, dangling references: if a `SalesInvoice` is deleted without cleaning up its `Attachment` records, the attachment's `subject_id` points to a non-existent record. Guard against this by registering a `before_delete` hook on every entity listed in `AllowedTypes` that queries the `Attachment` repository and either deletes the attachments or rejects the deletion.

Second, type spoofing: a malicious or buggy client could submit `subject_type = "User"` for an attachment that should only be linked to financial documents. The `AllowedTypes` constraint in the field declaration prevents this at the field validator layer.

Third, cross-tenant references: a `subject_id` value could reference a record in a different tenant's schema if tenant isolation is not enforced. The async validator for `DynamicLink` fields always resolves the subject repository through `entity.Resolve(ctx, tc, subjectType)` — which carries the current tenant context — ensuring the existence check is performed within the correct tenant schema.

> **Danger:** Never skip the `AllowedTypes` declaration on a `DynamicLink` field. Without it, any entity name — including `User`, `Tenant`, and other sensitive system entities — is a valid `subject_type` value. An attacker who can write attachment records could use this to enumerate the IDs of sensitive records by submitting UUIDs and observing whether the async validator accepts or rejects them.

---

#### Chapter summary

Chapter 6 covers the full edge system: the fundamentals of FK generation, direction, and naming (§6.1); one-to-many edge configuration including cascade policies and orphan handling (§6.2); many-to-many edges with junction table payload fields and filtering patterns (§6.3); self-referencing edges with materialised path and adjacency list strategies (§6.4); and polymorphic `DynamicLink` edges with their application-layer integrity requirements (§6.5). The three most important concepts are the eager-loading `Include` mechanism and its N+1 implications (§6.2.3), the materialised path pattern for efficient subtree queries on account and organisational hierarchies (§6.4.2), and the `DynamicLink` integrity risks that require application-layer guards (§6.5.4).

**Next chapters to read:**

- §7 — The EntityRecord Lifecycle (the hook system that fires on edge operations — cascade delete guards, path maintenance hooks, and DynamicLink cleanup all live in `before_delete` and `before_save` hooks)
- §8 — The Persistence Interface (the `EntityRepository` methods used throughout this chapter — `Include`, `HasEdgeWith`, `QueryDescendants`, `GetJunction`, `SetJunction` — are fully specified in the interface reference)
- §11 — Database Migrations (every edge declaration generates schema artefacts that must be reviewed in Atlas migration files before deployment; understanding the migration lifecycle is essential before adding edges to production entities)
