[<-- Back to Index](README.md)

## Business Module Integration

Every business module in AWO ERP calls `authz.Service` to protect its operations. This chapter documents the integration patterns and resource naming conventions for the two primary business modules: Finance and Sales.

---

### Finance Module

The Finance Module handles the most sensitive data in the ERP. Authorization here directly protects revenue figures, payroll, and audit trails.

```markdown
FINANCE RESOURCE NAMING CONVENTION:

invoice/{id}                Sales invoices
invoice/{id}/approve        Approval action on invoice
payment/{id}                Payment records
payment/{id}/allocate       Payment allocation action
journal/{id}                Journal entries
journal/{id}/post           Post a journal entry
journal/closed/{id}         Read-only closed-period journals
period/{id}/close           Period closing action
budget/{id}                 Budget records
budget/{id}/approve         Budget approval
report/finance/{type}       Financial reports (pnl, balance-sheet, etc.)
report/finance/{type}/export Export action on reports
account/{id}                Chart of accounts
ledger/{id}                 Ledger entries
payroll/{id}                Payroll records (HR/Finance boundary)
```

**Route protection setup:**
```go
financeRoutes := app.Group("/api/v1/finance",
    authnMiddleware,
)

financeRoutes.Get("/invoices",             svc.Middleware("invoice", "read"),    listInvoices)
financeRoutes.Get("/invoices/:id",         svc.Middleware("invoice", "read"),    getInvoice)
financeRoutes.Post("/invoices",            svc.Middleware("invoice", "create"),  createInvoice)
financeRoutes.Put("/invoices/:id",         svc.Middleware("invoice", "update"),  updateInvoice)
financeRoutes.Delete("/invoices/:id",      svc.Middleware("invoice", "delete"),  deleteInvoice)
financeRoutes.Post("/invoices/:id/approve",svc.Middleware("invoice", "approve"), approveInvoice)

financeRoutes.Get("/reports/:type",        svc.Middleware("report/finance", "read"),   getReport)
financeRoutes.Post("/reports/:type/export",svc.Middleware("report/finance", "export"), exportReport)

financeRoutes.Post("/periods/:id/close",   svc.Middleware("period", "close"), closePeriod)
```

**Default Finance Policies by Role:**
```markdown
role:cfo — Chief Financial Officer
  invoice/*           * allow     (full invoice control)
  payment/*           * allow     (full payment control)
  journal/*           * allow     (full journal control)
  period/*/close      execute allow  (period closing)
  budget/*            * allow
  report/finance/*    * allow
  payroll/*           * allow     (if HR/Finance boundary permits)

role:finance-manager
  invoice/*           * allow
  payment/*           * allow
  journal/*           create, read, update allow
  journal/closed/*    read allow  (read-only closed period)
  journal/closed/*    * deny      (cannot modify)
  budget/*            read allow
  report/finance/*    read allow

role:finance-viewer
  invoice/*           read allow
  payment/*           read allow
  journal/*           read allow
  report/finance/*    read allow

role:auditor  (time-limited)
  invoice/*           read allow
  payment/*           read allow
  journal/*           read allow
  report/finance/*    read, export allow
  payroll/*           read allow
  → All writes: denied by absence of allow rules
```

**Finance-specific authorization scenarios:**
```markdown
SCENARIO: Year-end audit (external auditors from Deloitte)

Required: Read access to all finance data, no write access.
Duration: 30 days (Jan 15 – Feb 14, 2026).

Implementation:
  expiry := time.Date(2026, 2, 14, 23, 59, 59, 0, time.UTC)
  svc.AssignRole(ctx, tenantID,
      "tenant:usr_auditor_deloitte",
      "role:auditor",
      dom,
      authz.WithExpiry(expiry),
      authz.WithAssignedBy("platform:admin"),
  )

After Feb 14: next Enforce call for this user → expired → role revoked → DENY


SCENARIO: Approve invoice over KES 1,000,000 (dual control)

Policy: Invoices > 1M require CFO approval (not just finance-manager).

Implementation via resource path:
  invoice/standard/{id}    → invoices ≤ 1M
  invoice/high-value/{id}  → invoices > 1M

Policies:
  p | role:finance-manager | {dom} | invoice/standard/* | approve | allow
  p | role:finance-manager | {dom} | invoice/high-value/*| approve | deny
  p | role:cfo             | {dom} | invoice/*           | approve | allow

Route handler classifies invoice before building object path:
  obj := "invoice/standard/" + id   // if amount ≤ 1M
  obj := "invoice/high-value/" + id // if amount > 1M

Enforce(Request{Subject: mgr, Domain: dom, Object: obj, Action: "approve"})
  → finance-manager + standard invoice → ALLOW
  → finance-manager + high-value invoice → DENY (explicit deny wins)
  → cfo + any invoice → ALLOW
```

---

### Selling Module

The Selling Module manages the order-to-cash cycle. Authorization here protects pricing, discount authority, and customer data.

```markdown
SALES RESOURCE NAMING CONVENTION:

order/{id}              Sales orders
order/{id}/confirm      Confirm action
order/{id}/cancel       Cancel action
quotation/{id}          Sales quotations
quotation/{id}/approve  Approval (for large discounts)
discount/standard/{id}  Discounts ≤ defined threshold
discount/high/{id}      Discounts above threshold
customer/{id}           Customer records
invoice/{id}            Sales invoices (shared with Finance module scope)
report/sales/{type}     Sales reports
commission/{id}         Sales commission records
territory/{id}          Territory assignments
lead/{id}               Lead records
opportunity/{id}        Opportunity records
```

**Default Sales Policies by Role:**
```markdown
role:sales-manager
  order/*              * allow
  quotation/*          * allow
  discount/*           * allow     (can approve any discount)
  customer/*           * allow
  lead/*               * allow
  opportunity/*        * allow
  report/sales/*       read, export allow
  commission/*         read allow  (view, not modify)
  territory/*          read, update allow

role:sales-rep
  order/*              create, read, update allow
  order/*/cancel       execute deny    (cannot cancel — needs manager)
  quotation/*          create, read allow
  quotation/*/approve  execute deny    (cannot approve their own quotes)
  discount/standard/*  create allow    (within threshold)
  discount/high/*      create deny     (exceeds authority)
  customer/*           read, update allow
  customer/*           delete deny     (cannot delete customers)
  lead/*               * allow
  opportunity/*        * allow
  report/sales/own/*   read allow      (own performance only)
  report/sales/team/*  read deny       (cannot see team data)

role:sales-viewer
  order/*              read allow
  quotation/*          read allow
  customer/*           read allow
  report/sales/*       read allow
```

**Sales-specific scenarios:**
```markdown
SCENARIO: Discount approval authority levels

company policy:
  0-10%:    sales rep can apply without approval
  11-25%:   sales manager approval required
  26-50%:   director approval required
  >50%:     CEO only

Resource paths:
  discount/level-1/{quoteID}   (0-10%)
  discount/level-2/{quoteID}   (11-25%)
  discount/level-3/{quoteID}   (26-50%)
  discount/level-4/{quoteID}   (>50%)

Policies:
  p | role:sales-rep     | {dom} | discount/level-1/* | approve | allow
  p | role:sales-rep     | {dom} | discount/level-2/* | approve | deny
  p | role:sales-manager | {dom} | discount/level-2/* | approve | allow
  p | role:director      | {dom} | discount/level-3/* | approve | allow
  p | role:ceo           | {dom} | discount/level-4/* | approve | allow

Quote handler determines level before building object path:
  level := classifyDiscountLevel(discountPct)
  obj := fmt.Sprintf("discount/%s/%s", level, quoteID)

SCENARIO: Territory-based data access

Sales reps only see customers in their territory.
Territories: east-africa, west-africa, middle-east, europe

Resource path: customer/territory/{region}/{customerID}

Policies:
  p | role:sales-rep-east-africa | {dom} | customer/territory/east-africa/* | * | allow
  p | role:sales-rep-east-africa | {dom} | customer/territory/west-africa/* | * | deny

User assignment:
  AssignRole(ctx, tenantID, "tenant:usr_james",
      "role:sales-rep-east-africa", dom)

Enforce(Request{Subject: "tenant:usr_james", Domain: dom,
    Object: "customer/territory/west-africa/cust_123", Action: "read"})
→ DENY (deny rule wins, even though sales-rep role might allow *)
```

---

### Cross-Module Resource Access (Finance × Sales)

Invoices exist in both Finance and Sales contexts. The resource path convention disambiguates:

```markdown
SHARED RESOURCES:

Finance module owns:    invoice/finance/*   (posted, accounting)
Sales module owns:      invoice/sales/*     (drafts, proforma)
Shared read:            invoice/*/read available to both

Example: Sales manager creates invoice draft
  Object: "invoice/sales/draft/inv_2026_001"
  Action: "create"
  → Covered by role:sales-manager policy: invoice/* create allow

Example: Finance manager posts the invoice (GL entry)
  Object: "invoice/finance/posted/inv_2026_001"
  Action: "post"
  → Covered by role:finance-manager policy: invoice/* * allow

Example: Sales manager tries to delete a POSTED invoice
  Object: "invoice/finance/posted/inv_2026_001"
  Action: "delete"
  → role:sales-manager does NOT have finance/* delete allow
  → DENY
```

---

Next: [How Other Packages Use authz](./14-how-other-packages-use-authz.md)
