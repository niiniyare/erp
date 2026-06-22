---
title: "Chapter 37: CRM Module"
part: "Part VI — Built-In Modules"
chapter: 37
section: "37-crm-module"
related:
  - "[Chapter 33: Finance Module](33-finance-module.md)"
  - "[Chapter 27: Defining Workflows](../part-05-workflow/27-defining-workflows.md)"
---

# Chapter 37: CRM Module

The CRM module manages the customer-facing side of the business: customer and contact records, address management, segmentation, and the lead-to-opportunity sales pipeline. It integrates with Finance (customer as AR debtor), Inventory (customer delivery addresses), and HR (salesperson assignment).

---

## 37.1. Customer and Contact

### 37.1.1. Customer Entity — Individual vs Organisation

Customers in Awo are either individuals (persons) or organisations (companies). The distinction affects which fields are required, how the customer is displayed, and how invoicing works (individual invoices use full legal name; company invoices include trading name and PIN).

```go
type Customer struct {
    // Identity
    CustomerNumber  string     // CUST-0001 — auto-generated naming series
    CustomerType    string     // "individual" | "organisation"
    // For organisations
    CompanyName     *string
    TradingName     *string    // DBA name
    KraPIN          *string    // Kenya Revenue Authority PIN (required for B2B)
    VATNumber       *string    // VAT registration if VAT-registered
    // For individuals
    FirstName       *string
    LastName        *string
    NationalID      *string
    // Common
    PrimaryEmail    string
    PrimaryPhone    string
    Website         *string
    // Classification
    CustomerGroupID *uuid.UUID
    TerritoryID     *uuid.UUID
    SalespersonID   *uuid.UUID // assigned account manager
    // Financial
    CreditLimit     decimal.Decimal
    PaymentTermsDays int        // e.g. 30 (net 30)
    Currency        string      // default invoice currency
    // Status
    Status          string     // "prospect" | "active" | "inactive" | "blacklisted"
    // GL link
    ReceivableAccountID uuid.UUID // AR sub-ledger account
}
```

**Why separate individual vs organisation?** Kenya's tax requirements differ:
- B2B transactions above KES 1M require the buyer's KRA PIN on the invoice
- Individual customers are subject to withholding tax on services in some cases
- Organisation invoices may need VAT reverse-charge handling

### 37.1.2. Contact Entity — Linked to Customer

An organisation customer has multiple contacts (procurement officer, finance director, CEO). A contact can also be linked to multiple organisations (a consultant who buys on behalf of several clients).

```go
type Contact struct {
    CustomerID    uuid.UUID   // primary customer link
    FirstName     string
    LastName      string
    JobTitle      *string
    Department    *string
    Email         *string
    Phone         *string
    IsPrimary     bool        // primary contact for invoices/notifications
    // Portal access
    HasPortalAccess bool
    PortalUserID    *uuid.UUID
}
```

**Portal access**: Contacts with `has_portal_access = true` get a portal user account. They can log in to view their invoices, make payments, and track order status. Portal users are separate from internal ERP users — they authenticate through the same session system but have a restricted permission set (`portal_customer` role).

### 37.1.3. Address Entity — Multiple Addresses Per Customer

```go
type Address struct {
    CustomerID    uuid.UUID
    Label         string     // "Head Office" | "Warehouse" | "Billing"
    AddressLine1  string
    AddressLine2  *string
    City          string
    County        string     // Nairobi, Mombasa, Kisumu, etc. (Kenya counties)
    PostalCode    *string
    Country       string     // default "KE"
    IsBilling     bool       // used for invoice address
    IsShipping    bool       // used for delivery address
    IsDefault     bool       // pre-filled when creating documents
    GeoLat        *float64   // for delivery routing
    GeoLng        *float64
}
```

Customers can have many addresses. An invoice picks its billing address; a delivery order picks its shipping address. Both default to `is_default = true` addresses but can be overridden per-document.

**County field** uses Kenya's 47 counties. The `select` field definition includes all counties so the UI shows a dropdown rather than free text, enabling clean reporting by county.

### 37.1.4. Customer Segmentation — Tags and Tiers

Two orthogonal segmentation axes:

**Tags** — free-form labels attached to a customer. Many-to-many relationship:

```go
type CustomerTag struct {
    CustomerID uuid.UUID
    Tag        string    // "vip", "distributor", "government", "ngo", "sme"
}
```

Tags drive targeted communications (bulk SMS to all `distributor` customers), reporting filters, and pricing rules (VIP customers get a different price list).

**Tiers** — structured classification with business rules:

```go
type CustomerTier struct {
    Name              string
    DiscountPct       decimal.Decimal  // automatic discount on sales orders
    CreditLimitMin    decimal.Decimal
    CreditLimitMax    decimal.Decimal
    PaymentTermsDays  int
    PriceListID       *uuid.UUID
}
```

Tier assignment can be manual or automatic (a nightly workflow checks last 12 months of purchases and upgrades/downgrades tier based on spend thresholds).

---

## 37.2. Lead and Opportunity Pipeline

### 37.2.1. Lead Entity — Source, Status, Owner

A lead is an unqualified sales prospect. It may eventually convert to a customer and opportunity, or be discarded.

```go
type Lead struct {
    // Identity
    LeadNumber    string     // LEAD-0001
    // Contact info (duplicated from Customer — not yet a customer)
    Name          string
    Company       *string
    Email         *string
    Phone         *string
    // Qualification
    Source        string     // "website" | "referral" | "cold_call" | "exhibition" | "social_media"
    Status        string     // "new" | "contacted" | "qualified" | "converted" | "disqualified"
    Rating        string     // "hot" | "warm" | "cold"
    // Assignment
    OwnerID       uuid.UUID  // salesperson
    // Tracking
    ConvertedAt   *time.Time
    CustomerID    *uuid.UUID // set when converted
    DisqualReason *string
}
```

Leads are the entry point for new business. The `source` field is critical for marketing ROI analysis — knowing that 40% of converted customers came via referrals informs where to invest sales effort.

### 37.2.2. Lead to Customer Conversion Workflow

Converting a lead creates a customer record and optionally an opportunity. This must be atomic — if customer creation fails, the lead should not be marked converted.

```go
func LeadConversionWorkflow(ctx workflow.Context, params LeadConversionParams) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Step 1: Create customer from lead data
    var customerID uuid.UUID
    err := workflow.ExecuteActivity(ctx, activities.CreateCustomerFromLead,
        LeadToCustomerInput{
            LeadID:       params.LeadID,
            CustomerType: params.CustomerType,
            KraPIN:       params.KraPIN,
        }).Get(ctx, &customerID)
    if err != nil {
        return fmt.Errorf("create customer: %w", err)
    }

    // Step 2: Mark lead as converted
    workflow.ExecuteActivity(ctx, activities.MarkLeadConverted,
        MarkConvertedInput{
            LeadID:      params.LeadID,
            CustomerID:  customerID,
            ConvertedAt: workflow.Now(ctx),
        })

    // Step 3: Optionally create opening opportunity
    if params.CreateOpportunity {
        workflow.ExecuteActivity(ctx, activities.CreateOpportunityFromLead,
            OpportunityFromLeadInput{
                LeadID:          params.LeadID,
                CustomerID:      customerID,
                ExpectedValue:   params.ExpectedValue,
                ExpectedCloseBy: params.ExpectedCloseBy,
            })
    }

    // Step 4: Notify salesperson
    workflow.ExecuteActivity(ctx, activities.NotifyLeadConverted,
        NotifyConvertedInput{
            OwnerID:    params.OwnerID,
            CustomerID: customerID,
        })

    return nil
}
```

**Idempotency**: The workflow ID is `{tenant}.lead.{lead_id}.convert`. If the API is called twice (double-click), the second call hits `REJECT_DUPLICATE` and returns the result from the first run. The lead is not converted twice.

### 37.2.3. Opportunity Entity — Stage, Expected Value, Close Date

An opportunity represents a specific deal being pursued with a qualified customer:

```go
type Opportunity struct {
    OpportunityNumber string
    CustomerID        uuid.UUID
    ContactID         *uuid.UUID  // which contact we're engaging
    Title             string      // e.g. "Annual fuel supply contract 2026"
    // Pipeline
    Stage             string      // "prospect" | "proposal" | "negotiation" | "won" | "lost"
    Probability       int         // 0-100 — likelihood of closing
    // Value
    ExpectedValue     decimal.Decimal
    Currency          string
    // Timeline
    ExpectedCloseBy   time.Time
    ActualCloseDate   *time.Time
    // Assignment
    OwnerID           uuid.UUID
    // Outcome
    LostReason        *string    // "price" | "competition" | "no_budget" | "no_decision"
    // Links
    LeadID            *uuid.UUID  // if converted from lead
    QuoteID           *uuid.UUID  // linked sales quotation
}
```

**Probability** is set by stage convention:
- prospect: 10%
- proposal sent: 30%
- negotiation: 60%
- verbal agreement: 80%
- won: 100%
- lost: 0%

The probability times expected value gives weighted pipeline value — the key metric for sales forecasting.

**Stage transitions** are validated by a `before_save` hook that prevents illegal moves (e.g., jumping from `prospect` directly to `won` without going through `proposal`). The stage machine is configurable per-tenant.

### 37.2.4. Opportunity Pipeline Report

The pipeline report is a weighted forecast of expected revenue:

```go
type PipelineReport struct {
    AsOfDate         time.Time
    TotalPipelineValue  decimal.Decimal  // sum of expected_value for open opps
    WeightedPipeline    decimal.Decimal  // sum of expected_value * probability
    ByStage          []StageBreakdown
    ByOwner          []OwnerBreakdown
    ByCloseMonth     []MonthBreakdown
    AtRisk           []Opportunity       // past expected_close_by, still open
}

type StageBreakdown struct {
    Stage           string
    Count           int
    TotalValue      decimal.Decimal
    WeightedValue   decimal.Decimal
}
```

The AMIS schema for the pipeline renders as both a table (filterable by owner, stage, close month) and a Kanban board where sales reps drag cards between stages. A drag triggers a PATCH request updating the opportunity stage, which fires the stage-transition validation hook.

```go
// AmisKanban configuration for opportunity pipeline
func OpportunityKanbanSchema(ctx PageBuilderContext) amis.Schema {
    return amis.Kanban{
        Source:     "/api/v1/opportunities?status=open",
        GroupField: "stage",
        Card: amis.Card{
            Title:    "${title}",
            Subtitle: "${customer.name}",
            Body:     "KES ${expected_value|number} · ${probability}% · Due ${expected_close_by|date}",
            Footer: []amis.Action{
                {Label: "View", Type: "button", ActionType: "link", Link: "/crm/opportunities/${id}"},
            },
        },
        OnDrop: amis.Action{
            ActionType: "ajax",
            API: amis.API{
                Method: "PATCH",
                URL:    "/api/v1/opportunities/${id}",
                Data:   map[string]interface{}{"stage": "${targetGroup}"},
            },
        },
    }
}
```

**At-risk opportunities** — any opportunity where `expected_close_by < today` and status is still open — appear highlighted in the report. A nightly workflow scans for at-risk opportunities and sends a digest to the sales manager.

```go
func NightlyPipelineHealthWorkflow(ctx workflow.Context, params PipelineHealthParams) error {
    var atRiskOpps []Opportunity
    workflow.ExecuteActivity(ctx, activities.FindAtRiskOpportunities,
        AtRiskInput{TenantID: params.TenantID, AsOfDate: workflow.Now(ctx)}).
        Get(ctx, &atRiskOpps)

    if len(atRiskOpps) > 0 {
        workflow.ExecuteActivity(ctx, activities.SendPipelineDigest,
            DigestInput{
                TenantID:    params.TenantID,
                AtRiskCount: len(atRiskOpps),
                AtRiskOpps:  atRiskOpps,
            })
    }
    return nil
}
```
