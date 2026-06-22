---
title: "Chapter 6: Edges — Relationships Between EntityDefinitions"
part: "Part II — The EntityDefinition System"
chapter: 6
section: "06-edges"
related:
  - "[Chapter 5: Field System](05-field-system.md)"
  - "[Chapter 8: The Persistence Interface](08-persistence-interface.md)"
  - "[Chapter 9: Privacy Policies](09-privacy-policies.md)"
---

# Chapter 6: Edges — Relationships Between EntityDefinitions

Every meaningful ERP system deals with relationships: customers have invoices, invoices have line items, line items reference products. In Awo, relationships between EntityDefinitions are declared as **edges** — typed, directional links that generate foreign key columns, indexes, join methods, and cascade rules. Understanding edge semantics deeply is essential to designing schemas that are both correct and performant.

---

## 6.1. Edge Fundamentals

### 6.1.1. What an Edge Declaration Generates

An edge declaration is not merely documentation. At schema load time the framework derives concrete artifacts from it:

**Database artifacts:**
- A `uuid` column on the "many" side (e.g. `customer_id uuid NOT NULL REFERENCES customers(id)`)
- A B-tree index on that column (e.g. `CREATE INDEX ON invoices(customer_id)`)
- Optionally, a cascade or restrict rule on delete/update

**Code artifacts:**
- A typed accessor method on the EntityRecord: `invoice.Customer()` returns a `*Customer` (lazy-loaded by default)
- A query method on the EntityRepository: `repo.QueryWith(filter, edge.Customer)` performs a join
- An amis form `select` widget pre-configured with the target entity's label and data source

**API artifacts:**
- Automatically populated `_links` object in JSON responses (HATEOAS-style)
- `?expand=customer` query parameter support to inline the related entity in the response

### 6.1.2. Edge Direction — Owner Side vs Inverse Side

Edges in Awo are directional. The **owner** side carries the foreign key column. The **inverse** side provides the back-reference.

```
Invoice (owner) ──── customer_id ──→ Customer (inverse)
```

The `Invoice` entity owns the edge: it holds `customer_id`. The `Customer` entity has an inverse edge (`.invoices`) that provides the list of invoices for a customer, backed by a query rather than a column.

```go
// In Invoice schema — OWNER SIDE
edge.To("customer", Customer.Type).
    Field("customer_id").
    Required()

// In Customer schema — INVERSE SIDE
edge.From("invoices", Invoice.Type).
    Ref("customer")
```

The framework enforces that exactly one side declares the column via `Field()`. Declaring it on both sides is a compile-time schema validation error.

### 6.1.3. Edge Naming Conventions

| Edge type | Naming convention | Example |
|---|---|---|
| Owner → single target | Singular noun | `customer`, `branch`, `created_by` |
| Inverse → many records | Plural noun | `invoices`, `line_items`, `attachments` |
| Self-referencing parent | `parent` | `parent` |
| Self-referencing children | `children` | `children` |
| Many-to-many | Singular describes one side | `roles`, `tags`, `members` |

Use descriptive names when multiple edges connect the same pair of entities:

```go
// Invoice has two User edges — billed_by and approved_by
edge.To("billed_by", User.Type).Field("billed_by_id").Required()
edge.To("approved_by", User.Type).Field("approved_by_id").Optional()
```

---

## 6.2. One-to-Many Edges

One-to-many is the most common relationship in ERP. A Customer has many Invoices. An Invoice has many InvoiceItems. A PurchaseOrder has many PurchaseOrderLines.

### 6.2.1. Declaring the Edge on Both Sides

```go
// --- invoice_item.go ---
func (InvoiceItem) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("invoice", Invoice.Type).
            Field("invoice_id").
            Required(),
    }
}

// --- invoice.go ---
func (Invoice) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("items", InvoiceItem.Type).
            Ref("invoice"),
    }
}
```

The `InvoiceItem.invoice_id` column is the only physical artifact. The `Invoice.items` inverse edge is a logical view — it runs `SELECT * FROM invoice_items WHERE invoice_id = $1` when accessed.

### 6.2.2. FK Column Placement — Always on the Many Side

This rule is absolute: the foreign key lives on the **many** side. Awo schema validation rejects any attempt to place it on the **one** side (which would require an array column, creating a denormalised mess).

Wrong:
```go
// Customer holds invoice_ids[] — WRONG, Awo rejects this
field.Strings("invoice_ids")
```

Right:
```go
// InvoiceItem holds invoice_id — CORRECT
edge.To("invoice", Invoice.Type).Field("invoice_id").Required()
```

### 6.2.3. Eager Loading vs Lazy Loading — Performance Implications

By default, Awo uses **lazy loading**: accessing `invoice.Items()` triggers a database query if the items have not already been fetched. This is convenient for single-record views but catastrophic in list contexts (the N+1 query problem).

For list queries where related data is needed, always eager-load with `With()`:

```go
// BAD — N+1: fetches invoices, then N queries for items
invoices, _ := repo.Query(ctx, filter.Eq("status", "approved"))
for _, inv := range invoices {
    items := inv.Items()  // triggers a query per invoice
}

// GOOD — single query with LEFT JOIN
invoices, _ := repo.QueryWith(ctx,
    filter.Eq("status", "approved"),
    query.With(edge.Items),
)
```

The `QueryWith` method generates a single SQL statement with a `LEFT JOIN` and populates the related records in memory.

Eager-loading is not always better. For large parent result sets (>500 rows) where only a few records need their children, selective lazy loading may produce fewer total rows transferred. Profile first, optimise second.

### 6.2.4. Cascade Delete — When to Use, When to Guard with a `before_delete` Hook

Cascade delete at the database level (`ON DELETE CASCADE`) is fast but silent. It will delete child records without running any Awo hooks. This means:
- No `before_delete` hooks run on child records
- No audit log entries for deleted children
- No Temporal workflow compensation triggers

**Use `ON DELETE CASCADE` only for:**
- Truly owned sub-records with no independent lifecycle (e.g. line items, address components)
- Records where the parent deletion is itself audited and the children deletion is implicit

```go
edge.To("invoice", Invoice.Type).
    Field("invoice_id").
    Required().
    Annotations(entsql.OnDelete(entsql.Cascade))
```

**Use a `before_delete` hook instead when:**
- Child records have their own audit requirement
- Child records trigger external side effects (inventory adjustments, GL postings)
- You need to validate that the parent can be safely deleted

```go
func (Invoice) Hooks() []ent.Hook {
    return []ent.Hook{
        hook.On(func(next ent.Mutator) ent.Mutator {
            return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
                if m.Op() == ent.OpDelete {
                    id, _ := m.(*ent.InvoiceMutation).ID()
                    // check no GL postings exist
                    count, err := glRepo.Count(ctx, filter.Eq("invoice_id", id))
                    if err != nil {
                        return nil, err
                    }
                    if count > 0 {
                        return nil, errs.NewBusinessError("POSTED_INVOICE",
                            "cannot delete an invoice with GL postings")
                    }
                }
                return next.Mutate(ctx, m)
            })
        }, ent.OpDelete),
    }
}
```

### 6.2.5. Orphan Handling — Restrict, Set Null, Cascade

Three strategies for what happens to children when the parent is deleted:

| Strategy | SQL | Use case |
|---|---|---|
| `RESTRICT` | `ON DELETE RESTRICT` | Default. Prevents parent deletion if children exist. Safe, explicit. |
| `SET NULL` | `ON DELETE SET NULL` | Children become parentless (nullable FK). Use for optional associations. |
| `CASCADE` | `ON DELETE CASCADE` | Children are deleted with the parent. Use for owned sub-records only. |

```go
// RESTRICT (default — no annotation needed)
edge.To("department", Department.Type).
    Field("department_id").
    Optional()

// SET NULL
edge.To("manager", User.Type).
    Field("manager_id").
    Optional().
    Annotations(entsql.OnDelete(entsql.SetNull))

// CASCADE — line items owned by the order
edge.To("purchase_order", PurchaseOrder.Type).
    Field("purchase_order_id").
    Required().
    Annotations(entsql.OnDelete(entsql.Cascade))
```

---

## 6.3. Many-to-Many Edges

Many-to-many relationships arise frequently in ERP: a Product can belong to many Categories, a User can hold many Roles, an Invoice can be tagged with many Cost Centres.

### 6.3.1. Junction Table Generation

Awo generates the junction table automatically from a many-to-many edge declaration:

```go
// --- product.go ---
func (Product) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("categories", Category.Type),
    }
}

// --- category.go ---
func (Category) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("products", Product.Type).
            Ref("categories"),
    }
}
```

Generated SQL:
```sql
CREATE TABLE product_categories (
    product_id  uuid NOT NULL REFERENCES products(id)  ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);
CREATE INDEX ON product_categories(category_id);
```

The junction table name follows the convention `{owner}_{inverse}` in alphabetical order of the two entity names.

### 6.3.2. Junction Table Annotations — Adding Payload Fields to the Relationship

Sometimes the relationship itself carries data: the date a user was assigned a role, the quantity of a product in a bundle, the cost centre split percentage. This requires an explicit junction entity rather than an auto-generated junction table.

```go
// Explicit junction entity: UserRole
type UserRole struct {
    ent.Schema
}

func (UserRole) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("user_id", uuid.UUID{}).Immutable(),
        field.UUID("role_id", uuid.UUID{}).Immutable(),
        field.Time("assigned_at").Default(time.Now).Immutable(),
        field.UUID("assigned_by_id", uuid.UUID{}).Optional(),
        field.Time("expires_at").Optional(),
    }
}

func (UserRole) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("user", User.Type).Field("user_id").Required().Unique().Immutable(),
        edge.To("role", Role.Type).Field("role_id").Required().Unique().Immutable(),
        edge.To("assigned_by", User.Type).Field("assigned_by_id").Optional(),
    }
}
```

The explicit junction entity participates fully in the Awo lifecycle: it has hooks, can be queried directly, and appears in the audit log.

### 6.3.3. Querying Through Many-to-Many Edges

```go
// Get all products in a specific category
products, _, err := productRepo.Query(ctx,
    filter.HasEdge("categories", filter.Eq("slug", "electronics")),
)

// Get all categories for a product
categories, _, err := categoryRepo.Query(ctx,
    filter.HasEdge("products", filter.Eq("id", productID)),
)
```

Under the hood, these become EXISTS subqueries on the junction table:
```sql
SELECT * FROM products p
WHERE EXISTS (
    SELECT 1 FROM product_categories pc
    JOIN categories c ON c.id = pc.category_id
    WHERE pc.product_id = p.id AND c.slug = 'electronics'
);
```

### 6.3.4. Performance Characteristics of Deep Many-to-Many Joins

Many-to-many joins through two levels of junction tables are expensive. For read-heavy reporting queries, consider:

1. **Materialised paths** (for hierarchical many-to-many, e.g. cost-centre hierarchies)
2. **Denormalised counters** (e.g. `product.category_count int` updated by a hook)
3. **JSONB array columns** for small, stable sets that are always queried together
4. **Separate read models** updated by Temporal workflows for complex aggregations

Never join through more than two levels of many-to-many in a single query. If your query requires it, redesign the schema.

---

## 6.4. Self-Referencing Edges

Self-referencing edges are used for hierarchical data: account charts of accounts (parent account → child accounts), organisation structures (manager → reports), product categories (parent category → subcategories).

### 6.4.1. Tree Structures — Parent/Children Pattern

```go
func (Account) Edges() []ent.Edge {
    return []ent.Edge{
        // edge to parent account
        edge.To("parent", Account.Type).
            Field("parent_id").
            Optional().
            From("children"),
    }
}
```

This generates:
- `parent_id uuid REFERENCES accounts(id) ON DELETE RESTRICT` — a nullable self-FK
- `children` — the inverse, returning all accounts where `parent_id = this.id`

The `ON DELETE RESTRICT` default is critical: you cannot delete an account that has child accounts, preventing orphaned subtrees.

### 6.4.2. Materialised Path for Deep Hierarchies

Adjacency lists are efficient for shallow trees (2-3 levels) but require recursive CTEs for deep trees. For account charts of accounts, org charts with many levels, or category trees, use the materialised path pattern:

```go
func (Account) Fields() []ent.Field {
    return []ent.Field{
        field.String("path").
            Comment("Materialised path: /root-id/parent-id/this-id/").
            MaxLen(4096).
            Default("/"),
        field.Int("depth").
            Default(0),
    }
}
```

The `path` column is updated by a `before_save` hook whenever `parent_id` changes:

```go
// In before_save hook:
if parentID != nil {
    parent, _ := repo.Get(ctx, *parentID)
    record.Set("path", parent.GetString("path") + record.ID().String() + "/")
    record.Set("depth", parent.GetInt("depth") + 1)
}
```

With materialised paths, querying all descendants is a single index scan:
```sql
SELECT * FROM accounts WHERE path LIKE '/root-uuid/%';
```

Querying all ancestors is also efficient:
```sql
SELECT * FROM accounts WHERE $1 LIKE path || '%';
-- where $1 is the descendant's path
```

### 6.4.3. Adjacency List for Shallow Hierarchies

When your tree has at most 3 levels (e.g. Region → Branch → Department), the simple adjacency list is sufficient. Use recursive CTEs in PostgreSQL for breadth-first traversal:

```sql
WITH RECURSIVE org_tree AS (
    SELECT id, name, parent_id, 0 AS depth
    FROM departments
    WHERE parent_id IS NULL   -- root nodes

    UNION ALL

    SELECT d.id, d.name, d.parent_id, t.depth + 1
    FROM departments d
    JOIN org_tree t ON t.id = d.parent_id
)
SELECT * FROM org_tree ORDER BY depth, name;
```

### 6.4.4. Querying Ancestors and Descendants Efficiently

For materialised path hierarchies:

```go
// Get all descendants of an account
descendants, _, err := accountRepo.Query(ctx,
    filter.StartsWith("path", account.GetString("path")),
)

// Get all ancestors of an account (walk the path)
ancestorIDs := extractUUIDsFromPath(account.GetString("path"))
ancestors, _, err := accountRepo.Query(ctx,
    filter.In("id", ancestorIDs),
)
```

For adjacency lists on trees shallower than 5 levels, use the `WithChildren` recursive eager-loader built into the framework's tree utilities:

```go
tree, err := accountRepo.QueryTree(ctx, filter.IsNull("parent_id"))
// Returns a nested tree structure populated in memory with a single CTE query
```

---

## 6.5. Polymorphic Relationships

Polymorphic relationships allow one record to reference records from multiple different entity types. The canonical ERP examples: attachments (a file can be attached to an invoice, a purchase order, or a customer), comments (can appear on any document), activity log entries.

### 6.5.1. When to Use DynamicLink vs a Union of Concrete Links

**Use DynamicLink when:**
- The set of target entity types is open-ended or grows over time
- The entity (e.g. Attachment, Comment) is a cross-cutting concern used by many modules
- You do not need foreign-key-level referential integrity

**Use a union of concrete Links when:**
- The set of target types is small and fixed (2-3 types)
- You need foreign-key constraints
- You need to query by the specific target type frequently

```go
// Union of concrete links — for a Notification that targets either a Customer or a Vendor
edge.To("customer", Customer.Type).Field("customer_id").Optional()
edge.To("vendor", Vendor.Type).Field("vendor_id").Optional()
```

Add a check constraint: `CHECK (num_nonnulls(customer_id, vendor_id) = 1)` to enforce exactly one.

### 6.5.2. DynamicLink Storage — `{field}_type` + `{field}_id` Column Pair

```go
func (Attachment) Fields() []ent.Field {
    return []ent.Field{
        field.String("linked_type").
            MaxLen(100).
            Comment("Entity name: Invoice, PurchaseOrder, Customer, etc.").
            NotEmpty(),
        field.UUID("linked_id", uuid.UUID{}).
            Comment("UUID of the linked record in the named entity's table"),
        field.String("file_path").MaxLen(1024),
        field.String("file_name").MaxLen(255),
        field.Int("file_size_bytes"),
        field.String("mime_type").MaxLen(100),
    }
}
```

There is intentionally no foreign key on `linked_id`. Referential integrity is an application-layer responsibility.

Create a composite index for fast lookup:
```sql
CREATE INDEX ON attachments(linked_type, linked_id);
```

### 6.5.3. Querying Polymorphic Edges

```go
// Get all attachments for a specific invoice
attachments, _, err := attachmentRepo.Query(ctx,
    filter.Eq("linked_type", "Invoice").
    And(filter.Eq("linked_id", invoiceID)),
)

// Get all attachments for any of a set of invoices
attachments, _, err := attachmentRepo.Query(ctx,
    filter.Eq("linked_type", "Invoice").
    And(filter.In("linked_id", invoiceIDs)),
)
```

To load attachments for every invoice in a list response, use a batch loader to avoid N+1:

```go
// BatchLoader fetches all attachments for a list of (type, id) pairs in one query
loader := attachment.NewBatchLoader(attachmentRepo)
for _, invoice := range invoices {
    loader.Add("Invoice", invoice.ID)
}
attachmentsByID, err := loader.Load(ctx)
```

### 6.5.4. Limitations — No FK Constraint, Application-Layer Integrity Only

Without a foreign key constraint, polymorphic links can become stale:

1. The referenced record is deleted without notifying the Attachment entity
2. The `linked_type` is misspelled, pointing to a non-existent entity
3. Records from a decommissioned module accumulate as orphans

Mitigations:

**Soft-delete instead of hard-delete** for entities that are polymorphically referenced. Set `deleted_at` and filter by `IS NULL` in queries.

**Periodic orphan cleanup job** via a Temporal workflow:
```go
// Runs weekly: verifies that linked_id records still exist
workflow.ExecuteActivity(ctx, activities.CleanOrphanedAttachments)
```

**Registry validation at boot**: the framework validates that all registered `linked_type` values correspond to known EntityDefinitions during `EntityRegistry.Validate()`.
