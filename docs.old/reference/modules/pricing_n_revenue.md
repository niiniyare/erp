# Awo ERP — Pricing & Revenue Management Module Specification

**Version:** 1.0.0  
**Status:** Draft  
**Module Code:** `pricing`  
**Owner:** Platform Architecture Team  
**Last Updated:** 2024-11-15

---

## Table of Contents

1. [Module Philosophy](#1-module-philosophy)
2. [Domain Overview — What Is Pricing?](#2-domain-overview--what-is-pricing)
3. [Integration Architecture](#3-integration-architecture)
4. [Domain Glossary](#4-domain-glossary)
5. [Price List Architecture](#5-price-list-architecture)
6. [Regulated Pricing — EPRA & Government-Controlled Goods](#6-regulated-pricing--epra--government-controlled-goods)
7. [Multi-Tier & Customer Segment Pricing](#7-multi-tier--customer-segment-pricing)
8. [Fleet & Contract Pricing](#8-fleet--contract-pricing)
9. [Promotions & Discounts](#9-promotions--discounts)
10. [Price Change Workflow & Approvals](#10-price-change-workflow--approvals)
11. [Price Resolution Engine](#11-price-resolution-engine)
12. [Revenue Recognition — IFRS 15](#12-revenue-recognition--ifrs-15)
13. [Revenue Reporting & Analytics](#13-revenue-reporting--analytics)
14. [Data Model](#14-data-model)
15. [Event Catalogue](#15-event-catalogue)
16. [API Surface](#16-api-surface)
17. [Permissions & Roles](#17-permissions--roles)
18. [Feature Flags & Configuration](#18-feature-flags--configuration)
19. [Appendix A — Price Resolution Worked Examples](#appendix-a--price-resolution-worked-examples)
20. [Appendix B — IFRS 15 Revenue Recognition Scenarios](#appendix-b--ifrs-15-revenue-recognition-scenarios)

---

## 1. Module Philosophy

### 1.1 Why Pricing Is Its Own Module

Pricing is not a property of a product. Pricing is a business policy applied to a product, at a moment in time, for a specific customer, on a specific channel. A litre of AGO diesel does not have a price — it has many prices simultaneously:

- The walk-in cash customer pays the regulated EPRA pump price.
- The fleet customer on a monthly contract pays EPRA minus a negotiated discount.
- The credit account customer pays EPRA plus a small risk premium.
- The same litre transferred to the workshop for a company vehicle is valued at cost for internal accounting purposes.

None of these are contradictions. They are all correct prices for the same product, applied to different commercial relationships. A system that treats "price" as a single number on a product record cannot handle this reality. Awo ERP models pricing as a **first-class domain** with its own lifecycle, approval workflow, resolution engine, and audit trail.

### 1.2 What This Module Owns

- **Price lists** — named, versioned collections of prices.
- **Price list entries** — the effective price of a specific item on a specific price list for a defined period.
- **Price tiers** — the named customer segments that map to price lists (CASH, CREDIT, FLEET, INTERNAL, WHOLESALE).
- **Pricing rules** — quantity breaks, minimum order values, volume-based discounts.
- **Promotions** — time-bounded price reductions or value-adds.
- **Price change workflow** — the drafting, approval, and activation process for new prices.
- **Price resolution engine** — the runtime logic that answers "what is the price for this item, for this customer, right now?"
- **Revenue recognition schedules** — the rules that govern when revenue earned is recognised in the GL.

### 1.3 What This Module Does NOT Own

| Concern | Owned By |
|---|---|
| Product catalogue (name, SKU, unit) | Inventory |
| Customer master | Sales / AR |
| GL journal posting | Finance GL |
| Contract terms & SLAs | Sales (contracts) |
| Invoicing | Finance AR |
| Tax computation | Tax module |
| Forecourt pump price display | Forecourt |

---

## 2. Domain Overview — What Is Pricing?

### 2.1 The Business Reality

For a petroleum retailer the pricing domain has layers of complexity that are easy to underestimate until they cause a compliance problem or a customer dispute:

**Regulated prices.** Fuel prices in Kenya are set monthly by the Energy and Petroleum Regulatory Authority (EPRA). The authority publishes maximum retail prices for PMS, AGO, and DPK. A station cannot legally sell above the published maximum. Crucially, EPRA prices vary by pump station location — a station in Nairobi has a different maximum to one in Mombasa or Kisumu, because transport costs differ. Any price management system must understand that the *ceiling* on a fuel price is a location-specific, time-bound regulatory number that changes on the 14th or 15th of each month.

**Cost-based floors.** A station cannot sustainably sell below its landed cost. The landed cost is itself variable — it depends on the ex-depot price, transport costs, and levies (including the Petroleum Development Levy and the Road Maintenance Levy). There is therefore a dynamic floor and a regulatory ceiling, and the station's commercial decision is what margin to operate within that band.

**Customer differentiation.** Not all customers are equal in commercial terms. A fleet operator who guarantees 50,000 litres per month is worth a discount over the walk-in customer. A credit account customer costs more to serve (working capital tied up, collection risk) and may attract a premium. The pricing system must support these differentials without violating the regulatory ceiling.

**Multi-product complexity.** Beyond fuel, the same ERP instance manages lubricants, shop goods, tyres, and workshop labour. Each of these has different pricing dynamics: lubricants have manufacturer-suggested retail prices, shop goods have competitive market pricing, workshop labour is quoted per job. The pricing module must accommodate all of these without being built specifically for any one of them.

### 2.2 The Revenue Management Perspective

Pricing is the upstream of revenue. Getting the price wrong — whether by applying the wrong tier to a customer, missing a regulatory update, or failing to retire an expired promotion — directly distorts revenue. Revenue management is the discipline of:

1. Ensuring the right price is charged every time (pricing integrity).
2. Ensuring revenue earned is recognised in the right period (revenue integrity).
3. Understanding where revenue comes from and where it is trending (revenue intelligence).

These three concerns map to the three main sections of this module: the price resolution engine, the revenue recognition engine, and the reporting layer.

---

## 3. Integration Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                     PRICING MODULE                                   │
│                                                                      │
│  Price Lists & Entries                                               │
│  Price Tiers                                                         │
│  Promotions & Rules                                                  │
│  Price Change Workflow                                               │
│  Price Resolution Engine ──────────────────────► Sales (POS, order) │
│                          ──────────────────────► Forecourt (pump)   │
│                          ──────────────────────► Procurement (PO)   │
│  Revenue Recognition     ──────────────────────► Finance GL         │
│  Revenue Schedules       ──────────────────────► Finance AR         │
│                                                                      │
│         READS FROM                                                   │
│  Products          ◄───────────────────────────── Inventory         │
│  Customer segments ◄───────────────────────────── Sales / AR        │
│  Fleet contracts   ◄───────────────────────────── Sales (contracts) │
│  EPRA published    ◄───────────────────── External / manual import  │
│  GL accounts       ◄───────────────────────────── Finance GL        │
│  Tax rates         ◄───────────────────────────── Tax module        │
│  Site / location   ◄───────────────────────────── Core / Sites      │
└──────────────────────────────────────────────────────────────────────┘
```

### 3.1 Relationship with the Sales Module

The Sales module is the primary consumer of the price resolution engine. Every sales transaction — whether raised at the forecourt POS, on a sales order, or on a proforma invoice — calls the price resolver with a set of context parameters and receives back a resolved unit price and the audit trail of how it was derived. The Sales module never stores prices independently; it stores only the resolved price and the `price_resolution_id` that links back to the Pricing module's audit record.

### 3.2 Relationship with the Forecourt Module

The Forecourt module calls the price resolver at shift-open time to lock the pump price for each nozzle/product for the duration of the shift. This price is immutable for the shift once locked. If EPRA publishes a new price during a shift, the new price takes effect at the next shift open — the pump display may show the new price, but the system price remains locked to the shift-open resolution until handover.

### 3.3 Relationship with Finance

The revenue recognition engine writes recognition schedules to the Finance GL at the point of obligation satisfaction. For point-of-sale transactions (fuel, shop goods) this is immediate. For subscription services, retainers, or advance payments it is spread over the performance period. Finance owns the journal; Pricing owns the schedule that instructs Finance when and how much to post.

---

## 4. Domain Glossary

| Term | Definition |
|---|---|
| **Price List** | A named collection of prices for a defined set of items, applicable to a specific customer segment or channel. Price lists are date-effective and versioned. |
| **Price List Entry** | A single price for a specific item on a specific price list, effective for a defined period. The atomic unit of pricing. |
| **Price Tier** | A named customer or channel segment that maps to a price list (e.g. CASH, CREDIT, FLEET, INTERNAL, WHOLESALE). |
| **Price Resolution** | The runtime process of determining the final applicable price for a specific item, customer, quantity, channel, and date. |
| **Price Resolution Audit** | The immutable record of how a price was resolved for a specific transaction — which price list was used, which rules were applied, and what the final price was. |
| **Pricing Rule** | A conditional modifier on a base price — e.g. a quantity break (10% off for > 1,000 L), a minimum order surcharge, or a value-based discount. |
| **Promotion** | A time-bounded price reduction or value-add (buy X get Y, percentage off, fixed amount off) applicable to a defined set of products and customers. |
| **Regulated Price** | A price set or capped by a government authority (EPRA for fuel). The regulated price is a ceiling, not a target. |
| **Price Ceiling** | The maximum price above which a sale is non-compliant with regulation. |
| **Price Floor** | The minimum price below which a sale is commercially unsustainable (typically landed cost). Not a regulatory concept but a business control. |
| **Landed Cost** | The total cost of acquiring a product at the point of sale, including purchase price, transport, levies, and handling. |
| **Performance Obligation** | Under IFRS 15, a distinct promise to transfer a good or service to a customer. Revenue is recognised when this obligation is satisfied. |
| **Transaction Price** | The amount of consideration a seller expects to receive in exchange for satisfying its performance obligations. May differ from the list price after discounts and variable consideration. |
| **Revenue Recognition Schedule** | A timetable specifying when portions of a transaction price should be recognised as revenue in the GL. Relevant for multi-element arrangements and advance payments. |
| **Deferred Revenue** | Revenue received or invoiced before the performance obligation is satisfied. Held as a liability until recognition criteria are met. |
| **Variable Consideration** | A component of the transaction price that is uncertain at contract inception — rebates, volume discounts, performance bonuses, penalties. |
| **Price Waterfall** | The sequential set of deductions from the published list price that arrives at the net price: list → trade discount → volume discount → promotional discount → net price. |
| **EPRA** | Energy and Petroleum Regulatory Authority (Kenya). Publishes maximum pump prices monthly for gazetted petroleum products. |

---

## 5. Price List Architecture

### 5.1 Design Principles

**Price lists are the source of truth, not products.** A product has no inherent price in Awo ERP. A price only exists as an entry on a price list. This separation ensures that the same product can have different prices in different commercial contexts without any ambiguity.

**Price lists are date-effective.** Every price list entry has an `effective_from` and an `effective_to`. This means the system retains the complete history of what was charged when, which is essential for dispute resolution, regulatory audits, and period-end reconciliations. You can answer the question "what was the pump price for AGO on the 17th of last month?" without any approximation.

**Price lists are versioned.** When prices change, the old entries are closed (their `effective_to` is set) and new entries are created. The price list itself is never mutated — only new entries are added. This is the same append-only principle applied throughout Awo ERP.

**Price lists inherit.** A child price list can inherit all entries from a parent and override only specific items. This allows a site-specific price list to differ from the national default on only one or two products without duplicating every entry.

### 5.2 Price List Hierarchy

```
National Default Price List (base)
│
├── Site Price List — Shell Maanzoni (overrides fuel prices for Nairobi zone)
│     ├── CASH tier        ← EPRA max for Nairobi, updated monthly
│     ├── CREDIT tier      ← EPRA max + KES 2.00 risk premium
│     ├── FLEET tier       ← EPRA max − negotiated discount per contract
│     └── INTERNAL tier    ← Landed cost (for inter-department transfers)
│
└── Wholesale Price List (for bulk off-taker customers)
      └── WHOLESALE tier   ← Negotiated, typically ex-depot + margin
```

### 5.3 Price List Entry Lifecycle

```
DRAFT
  │  (created by pricing analyst or system import)
  ▼
PENDING_APPROVAL
  │  (submitted for manager / compliance review)
  ▼
APPROVED
  │  (approved; activation is scheduled for effective_from date)
  ▼
ACTIVE
  │  (effective_from ≤ today ≤ effective_to; used in price resolution)
  ▼
SUPERSEDED
     (effective_to passed, or a newer entry replaced it)
```

### 5.4 Price List Types

| Type | Description | Examples |
|---|---|---|
| `RETAIL` | Prices for direct sales to individual customers | Pump price, shop shelf price |
| `FLEET` | Prices for named fleet contract customers | Volume-discounted fuel price |
| `WHOLESALE` | Prices for bulk resellers or off-takers | Ex-depot bulk price |
| `INTERNAL` | Transfer prices for inter-department or inter-company transactions | Workshop fuel allocation at cost |
| `PROCUREMENT` | Purchase prices from suppliers (used for landed cost computation) | Depot ex-price, transport rate |
| `PROMOTIONAL` | Time-bounded override prices | EPRA regulatory special, seasonal discount |

---

## 6. Regulated Pricing — EPRA & Government-Controlled Goods

### 6.1 The EPRA Price Cycle

Kenya's Energy and Petroleum Regulatory Authority (EPRA) publishes revised maximum retail pump prices every month, typically effective on the 14th or 15th. The published prices are location-specific — stations are classified into pricing zones based on their distance from the inland depot network. A Nairobi station and a Kisumu station have different maximums for the same product because transport costs to each location differ.

The published EPRA price is a **maximum** — a station may sell below it but never above it. In practice, in a competitive market most stations sell at or very near the maximum because margins are already thin and there is little incentive to undercut.

The EPRA pricing cycle creates a specific operational challenge: prices change on a fixed day each month, but petrol stations run 24 hours. The price change must be activated at exactly the right moment — the first pump transaction after the effective date must use the new price. A failure to update in time is a commercial loss (if the price increased) or a regulatory violation (if the price decreased and the station continues charging the higher amount).

### 6.2 EPRA Price Import

EPRA publishes prices in PDF and, more recently, in machine-readable formats on its website. Awo supports:

- **Manual entry** — a pricing officer enters the new prices into the pending price list entry form.
- **File import** — a CSV/Excel upload parsed by the system, with field mapping for product, zone, and price.
- **API import** — (future / feature-flagged) a scheduled job that fetches and parses the EPRA publication automatically.

Regardless of import method, the imported price enters the workflow as a `DRAFT` entry and must pass through the approval workflow before activation. The approval workflow for EPRA-updated prices is streamlined (a single compliance-officer approval is sufficient) compared to the full commercial pricing change workflow.

### 6.3 Compliance Enforcement

The price resolution engine enforces the regulated price ceiling at the point of resolution. When a price is resolved for a regulated product:

1. The engine retrieves the applicable EPRA maximum for the product, site zone, and transaction date.
2. If the resolved price (after discounts, promotions, etc.) exceeds the ceiling, the engine **rejects the resolution** and returns an error — not a silently capped price. This forces the pricing team to correct the price list, not the system to quietly hide a violation.
3. If the resolved price is below the floor (landed cost), the engine logs a `pricing.margin_alert` event — a warning, not a block — because selling below cost is a commercial decision the business is permitted to make.

### 6.4 EPRA Price Entry Format

When a new EPRA price is imported:
- `regulation_ref` captures the EPRA gazette notice number.
- `effective_from` is the EPRA-gazetted effective date.
- `price_zone` links to the site's regulatory pricing zone.
- `regulated_maximum` stores the EPRA ceiling.
- `entry_status` is forced to `PENDING_APPROVAL` regardless of the import method.

---

## 7. Multi-Tier & Customer Segment Pricing

### 7.1 Why Customer Segments Matter

Not all customers cost the same to serve and not all customers represent the same commercial value. Pricing tiers allow the business to:

- **Reward volume** — fleet customers who guarantee monthly volumes receive a discount.
- **Price for risk** — credit customers who pay 30 days after purchase carry receivables risk; a small premium compensates.
- **Protect margin** — walk-in cash customers on a regulated product should pay the full regulated price; there is no reason to offer them a discount.
- **Enable internal accounting** — fuel drawn from the tank for the company's own vehicles should be valued at cost, not retail price, for management accounts.

### 7.2 Standard Tiers

| Tier Code | Applies To | Pricing Basis | Typical Delta |
|---|---|---|---|
| `CASH` | Walk-in cash / MPesa / card customers | EPRA maximum (fuel) or RRP (shop) | Baseline |
| `CREDIT` | Named account customers on deferred terms | EPRA + risk premium | +KES 1–5/L |
| `FLEET` | Contract fleet operators | EPRA − volume discount | −KES 2–10/L |
| `WHOLESALE` | Bulk off-takers, resellers | Ex-depot + negotiated margin | Varies widely |
| `INTERNAL` | Inter-department / inter-company | Landed cost (WAC) | Cost price only |
| `STAFF` | Employee purchases | RRP − staff discount % | Site policy |

### 7.3 Tier Assignment

A customer is assigned to a price tier in the Sales / AR module as part of their customer record. The Pricing module reads this assignment at resolution time. A single customer may have different tier assignments per product category — for example, a fleet customer might be on the `FLEET` tier for fuel but `CASH` tier for lubricants.

Tier assignment is stored on the customer record with an effective date. If a fleet contract expires and is not renewed, the customer automatically falls back to the `CASH` tier from the expiry date onwards — the price resolver checks the tier-assignment effective date at every resolution, it does not rely on the current assignment alone.

---

## 8. Fleet & Contract Pricing

### 8.1 Fleet Pricing in the Petrol Station Context

Fleet pricing is the most commercially sensitive part of the pricing domain for a petroleum retailer. A well-managed fleet account can represent 30–50% of a site's total fuel volume. Getting it wrong — either by offering too steep a discount that erodes margin to zero, or by quoting inconsistently and losing the account — has a material impact on site profitability.

Fleet pricing in Kenya typically works as follows:

- The station agrees a **per-litre discount below the prevailing EPRA price** with the fleet operator.
- The discount is expressed as a fixed KES amount (e.g. EPRA − KES 5.00) or as a percentage.
- Because EPRA changes monthly, the actual net price changes monthly even if the discount agreement is static.
- The fleet operator is given a credit account. Drivers fill up and sign a delivery note. An invoice is raised at the end of the week or month.
- Volume thresholds may be embedded in the agreement: e.g. the KES 5 discount applies only if the customer takes ≥ 20,000 L/month. Below that, a smaller discount (KES 2) applies.

### 8.2 Contract Price Structure

A fleet contract is owned by the Sales module (as a contract record) but its pricing terms are owned by the Pricing module. A contract pricing record specifies:

- The customer it applies to.
- The products it covers.
- The base reference (typically `EPRA_MAX` for the site's zone).
- The discount method (`FIXED_AMOUNT` or `PERCENTAGE`).
- The discount value at each volume threshold.
- The contract period (`effective_from` / `effective_to`).
- The review cadence (monthly auto-renewal, annual renegotiation, etc.).

### 8.3 Volume Threshold Discounts

```
Contract: ACME Logistics Fleet Account
Product:  AGO (Diesel)
Period:   2024-11-01 to 2025-10-31

Volume Tier    |  Discount Method  |  Discount Value
< 10,000 L/mo  |  FIXED_AMOUNT     |  KES 2.00/L
10,000–30,000  |  FIXED_AMOUNT     |  KES 4.00/L
> 30,000 L/mo  |  FIXED_AMOUNT     |  KES 6.00/L

Reference price: EPRA_MAX (Nairobi zone, current month)
Effective net price = EPRA_MAX − discount_for_volume_tier
```

Volume tier assessment can be done in two ways, configurable per contract:

- **Retrospective rebate** — the customer fills at the CASH price during the month, and a credit note is issued at month-end for the discount earned based on actual volume. This is simpler operationally but creates an accounts receivable adjustment process.
- **Prospective tier** — the previous month's volume determines the discount tier applied in the current month. The customer fills at the discounted price from day one of the month.

The Pricing module supports both models. Retrospective rebates generate a `pricing.rebate_schedule` record that instructs the Finance AR module to issue a credit note on a specified date.

### 8.4 Contract Compliance Monitoring

The Pricing module monitors fleet contract compliance:

- **Volume shortfall** — if a customer's actual monthly volume falls below a tier threshold, the module logs a `pricing.contract.volume_shortfall` event. The sales team is notified. If the contract has a minimum volume commitment clause, the shortfall may trigger a top-up charge.
- **Price ceiling breach** — if any transaction on a fleet account resolves to a price above the EPRA maximum (which should be impossible if the system is configured correctly, but may occur if the EPRA import is delayed), the module logs a `pricing.compliance.ceiling_breach` event immediately.
- **Contract expiry** — 30, 14, and 7 days before a contract expires, the module emits `pricing.contract.expiry_warning` events. The sales team must either renew or let the customer fall to the default tier.

---

## 9. Promotions & Discounts

### 9.1 What Qualifies as a Promotion

A promotion is a **time-bounded, conditionally applicable modification to the selling price**. It differs from a price list entry (which is the baseline price) in that it is:

- Temporary by design — it has a hard start and end date/time.
- Conditional — it may apply only to specific products, customer segments, quantities, or transaction channels.
- Stackable (with limits) — multiple promotions may apply to the same transaction, subject to stacking rules.
- Campaign-driven — it is created as part of a marketing or commercial decision, not as a routine pricing update.

### 9.2 Promotion Types

| Type | Description | Example |
|---|---|---|
| `PERCENTAGE_OFF` | Reduce the resolved base price by a percentage. | 5% off premium fuel on Sundays |
| `FIXED_AMOUNT_OFF` | Reduce by a fixed monetary amount. | KES 3 off per litre during launch week |
| `FIXED_PRICE` | Override the resolved price with a specific price. | PMS at KES 180.00 flat for loyalty members |
| `BUY_X_GET_Y` | Purchase X units of product A, receive Y units of product B free or discounted. | Buy 20L of engine oil, get 1L coolant free |
| `BUNDLE` | A set of products priced together below the sum of their individual prices. | Full service pack: oil + filter + labour at bundled rate |
| `VOLUME_BREAK` | Reduce unit price when quantity exceeds a threshold. | 3% off when filling ≥ 50L |
| `LOYALTY_REWARD` | Points earn or redemption event applied to a transaction. | Earn 1 point per litre; redeem 100 points for KES 50 off |
| `FREE_GIFT` | Physical product added to a qualifying order at zero price. | Free air freshener with full tank |

### 9.3 Promotion Conditions

A promotion can carry any combination of the following conditions, evaluated at resolution time:

| Condition | Logic |
|---|---|
| `product_ids` | Promotion applies only to these specific products or product categories. |
| `customer_tier` | Applies only to customers on a specific price tier (e.g. FLEET only). |
| `customer_ids` | Applies only to a named list of customers (targeted promotion). |
| `site_ids` | Applies only at specific sites. |
| `channel` | Applies only on a specific transaction channel (FORECOURT, ONLINE, WHOLESALE). |
| `min_quantity` | Minimum quantity to qualify. |
| `min_transaction_value` | Minimum transaction value to qualify. |
| `day_of_week` | Promotion is active only on specified days. |
| `time_of_day` | Promotion active only between specified hours (e.g. 06:00–08:00 morning rush). |
| `payment_method` | Applies only to specific payment methods (MPesa-only promotions). |

### 9.4 Stacking Rules

The default stacking rule is **best-price wins**: if multiple promotions are applicable to a transaction, the one producing the lowest net price is applied. No two promotions are combined additively by default because this can produce unintended margin outcomes.

When deliberately stackable promotions are needed (e.g. a loyalty earn that stacks with a volume break), the promotions are explicitly linked in a `promotion_stack_group`. Only promotions in the same stack group are combined. Promotions outside the group are evaluated against the stack's result and the better price wins.

Promotions are always subject to the regulatory price ceiling. If a promotion on a regulated product would produce a price above the EPRA ceiling, the promotion is invalid and should not have been created — the system rejects the promotion entry at creation time with a ceiling violation error.

### 9.5 Promotion Approval Workflow

Like price list entries, promotions pass through an approval workflow before activation:

```
DRAFT → PENDING_APPROVAL → APPROVED → SCHEDULED → ACTIVE → EXPIRED
```

A promotion in `APPROVED` state with a future `starts_at` is in `SCHEDULED` state logically — it will activate automatically at `starts_at` without further intervention. Once `ends_at` passes, it moves to `EXPIRED` and is excluded from resolution immediately.

---

## 10. Price Change Workflow & Approvals

### 10.1 Why Controlled Price Changes Matter

An uncontrolled price change in a live system is one of the most damaging events in retail operations. Consider:

- A junior clerk mis-types the new EPRA AGO price as KES 140.50 instead of KES 164.50. For the next three hours until someone notices, the station loses KES 24 on every litre dispensed.
- A fleet discount is entered with the wrong product ID, accidentally applying a KES 10/L discount to the premium petrol (PMS) instead of the diesel (AGO). The fleet customer fills 5,000L before the error is caught.
- A promotional price entered without an end date runs indefinitely, not for the intended 3-day period.

The price change workflow is the control that prevents these outcomes.

### 10.2 Workflow Tiers

Different price changes carry different levels of risk and therefore route through different approval tiers:

| Change Type | Risk Level | Approval Required |
|---|---|---|
| EPRA regulatory update (matching published figure exactly) | Low | Compliance Officer (1 approver) |
| Retail price change within ± 5% of current price | Medium | Pricing Manager (1 approver) |
| Retail price change > ± 5% of current price | High | Pricing Manager + General Manager (2 approvers) |
| New fleet/contract price | High | Sales Manager + Pricing Manager (2 approvers) |
| New promotion creation | Medium | Marketing Manager (1 approver) |
| Promotion with value > KES 50,000 estimated impact | High | Marketing Manager + General Manager (2 approvers) |
| Internal transfer price change | Low | Finance Manager (1 approver) |

The workflow tier is determined automatically at submission time based on the change parameters. The submitter cannot select their own approval tier.

### 10.3 Workflow States

```
                     ┌──────────────────────────────────┐
                     │   Pricing Analyst / Importer      │
                     └──────────────┬───────────────────┘
                                    │ creates
                                    ▼
                               [ DRAFT ]
                                    │ submits
                                    ▼
                         [ PENDING_APPROVAL ]
                          │               │
                     approved          rejected
                          │               │
                          ▼               ▼
                      [ APPROVED ]   [ REJECTED ]
                      (2-tier:           │
                    2nd approver         │
                    must also sign)  ────┘
                          │
                     scheduled for
                     effective_from
                          │
               ┌──────────┴──────────┐
               │  on effective_from  │
               ▼                     ▼
           [ ACTIVE ]         (old entry →)
                              [ SUPERSEDED ]
```

### 10.4 Dual Control for High-Risk Changes

For high-risk changes requiring two approvers, the second approver can only act after the first has approved — the system does not allow out-of-order approval. If the first approver rejects, the item returns to `DRAFT` and the second approver is never asked. This prevents a situation where one approver approves a change they know is wrong, expecting the second approver to catch it.

### 10.5 Effective Date Scheduling

An approved price change with a future `effective_from` is held in `APPROVED` state and activated by a scheduled job at the exact `effective_from` datetime. For EPRA changes this is typically midnight on the 14th or 15th of the month.

The activation process:
1. The job runs every minute, querying for `APPROVED` entries where `effective_from <= now()`.
2. For each entry found, it sets `status = ACTIVE` and simultaneously closes any currently active entry for the same item/tier/site by setting its `effective_to = effective_from` of the new entry.
3. Both operations happen in the same database transaction. The changeover is atomic — there is never a moment where either zero prices or two prices are active for the same item.
4. A `pricing.price_list_entry.activated` event is emitted, which the Forecourt module listens to for pump display updates.

### 10.6 Emergency Price Changes

An emergency override allows an authorised manager to activate a price change immediately, bypassing the scheduled effective date. Emergency overrides:

- Require a mandatory justification text field.
- Are automatically escalated to the next approval tier above normal.
- Generate a `pricing.emergency_override.activated` event that is immediately visible to all managers and the audit log.
- Cannot be initiated by the same person who created the price entry.

Emergency overrides are the right tool when, for example, the EPRA publishes an intra-month correction (rare but it happens) or when a data entry error in an active price is discovered and the business cannot wait for the next shift change.

---

## 11. Price Resolution Engine

### 11.1 What the Engine Does

The price resolution engine is a pure function: given a **resolution context**, it returns a **resolved price** and an **audit record**. It does not modify any data. It does not cache results in a way that changes behaviour. Given the same inputs at the same moment in time, it always returns the same output.

### 11.2 Resolution Context

```go
// PriceResolutionRequest is the input to the engine.
type PriceResolutionRequest struct {
    TenantID      uuid.UUID
    SiteID        uuid.UUID
    ProductID     uuid.UUID
    CustomerID    *uuid.UUID   // nil for anonymous/walk-in
    CustomerTier  string       // CASH, CREDIT, FLEET, INTERNAL, WHOLESALE, STAFF
    Quantity      decimal.Decimal
    Channel       string       // FORECOURT, POS, SALES_ORDER, INTERNAL
    PaymentMethod *string      // CASH, MPESA, CARD, CREDIT — nil if unknown at resolution time
    TransactionAt time.Time    // point in time to resolve for
    LockToShift   *uuid.UUID   // if non-nil, use the price locked to this shift
}
```

### 11.3 The Price Waterfall

The engine applies the following steps in order. Each step may modify the running price or add a line to the audit trail. The process short-circuits only in error conditions.

```
Step 1 — BASE PRICE
│ Look up the active price list entry for:
│   (product_id, customer_tier, site_id, transaction_date)
│ Fallback chain: site-specific list → national default list → error if none found
│ Result: base_price
│
Step 2 — SHIFT LOCK (if applicable)
│ If LockToShift is set: retrieve the price locked to that shift for this nozzle/product.
│ This bypasses Steps 3–6 — the shift price is already resolved and immutable.
│ Result: locked_price (and skip to Step 7)
│
Step 3 — CONTRACT PRICE
│ If customer has an active fleet/contract pricing entry for this product:
│   Apply contract discount (fixed amount or percentage) to base_price.
│   Evaluate the customer's current volume tier against their contract thresholds.
│ Result: contract_price (may equal base_price if no contract applies)
│
Step 4 — PRICING RULES
│ Evaluate all active pricing rules for the context:
│   Quantity breaks (based on Quantity in request)
│   Minimum order rules
│   Channel-specific rules
│ Apply the most beneficial rule (or all rules if stacking is enabled).
│ Result: rule_adjusted_price
│
Step 5 — PROMOTIONS
│ Retrieve all active promotions whose conditions match the resolution context.
│ Evaluate each promotion's discount type.
│ Apply per stacking rules (default: best promotion wins).
│ Result: promoted_price
│
Step 6 — REGULATORY CEILING CHECK
│ If the product has a regulated_maximum for this site's zone and date:
│   If promoted_price > regulated_maximum → REJECT (ceiling violation)
│   If promoted_price < 0 → REJECT (negative price)
│ Log a margin_alert if promoted_price < landed_cost
│
Step 7 — TAX CALCULATION
│ Call Tax module with (product_id, site_id, customer_id, net_price, quantity).
│ Tax module returns tax_lines[] (VAT, excise duty, levy breakdowns).
│ Result: gross_price = net_price + sum(tax_lines)
│
Step 8 — AUDIT RECORD
│ Write pricing_resolution_audit row capturing:
│   Every step's input and output
│   Which price list entry was used
│   Which contract, rules, promotions were applied
│   Final net and gross price
│   Regulatory ceiling and margin status
│ Return PriceResolutionResult
```

### 11.4 Resolution Result

```go
type PriceResolutionResult struct {
    ResolutionID      uuid.UUID
    ProductID         uuid.UUID
    CustomerTier      string
    // Price waterfall components
    BasePrice         decimal.Decimal
    ContractDiscount  decimal.Decimal
    RuleDiscount      decimal.Decimal
    PromoDiscount     decimal.Decimal
    NetPrice          decimal.Decimal  // price before tax
    TaxLines          []TaxLine
    GrossPrice        decimal.Decimal  // price customer pays
    Currency          string
    // Metadata
    PriceListEntryID  uuid.UUID
    ContractID        *uuid.UUID
    PromotionIDs      []uuid.UUID
    RegulatedMaximum  *decimal.Decimal
    MarginAlert       bool
    ResolvedAt        time.Time
    EffectiveUntil    *time.Time  // nil for spot resolution; set for shift-locked prices
}
```

### 11.5 Performance

The price resolver is called on every transaction. At a busy petrol station this may be hundreds of calls per hour. The engine is designed for low latency:

- Active price list entries are cached in-process (Ristretto) with a TTL of 60 seconds. The cache is invalidated immediately on `pricing.price_list_entry.activated` events.
- Contract prices are cached per customer with a TTL of 5 minutes.
- Promotions are cached as a set with a TTL of 30 seconds.
- The regulatory ceiling table is cached for the entire calendar month and refreshed on the 13th of each month.
- Tax rates are cached per (product, site) pair with a 24-hour TTL.

Cache invalidation is driven by domain events published through the outbox, not by time expiry alone. A price change activation immediately invalidates the relevant cache entry across all application nodes.

---

## 12. Revenue Recognition — IFRS 15

### 12.1 The Standard

IFRS 15 "Revenue from Contracts with Customers" requires revenue to be recognised when (or as) a performance obligation is satisfied — that is, when control of the promised good or service is transferred to the customer. For most petroleum retail transactions this is straightforward: the customer fills their tank, drives away, and revenue is recognised at the pump. But Awo ERP serves businesses where the picture is more complex.

### 12.2 Revenue Recognition Patterns by Transaction Type

#### 12.2.1 Point-of-Sale — Immediate Recognition

The simplest case: a walk-in customer buys 30 litres of PMS, pays by MPesa, and drives away. One performance obligation (deliver fuel), satisfied at a single point in time (the dispense), with certain consideration (KES 5,250).

```
At dispense:
DR  MPesa Clearing     5,250
    CR  Revenue — PMS          5,250
```

No revenue recognition schedule is needed. Revenue is posted immediately as part of the shift close GL journal.

#### 12.2.2 Credit Sales — Obligation Satisfied, Cash Not Yet Received

A fleet customer fills 500 litres on account. The performance obligation (deliver fuel) is satisfied at the pump. Revenue is recognised immediately even though cash will arrive 30 days later. The difference from the above case is the debit entry goes to AR, not cash.

```
At dispense:
DR  Accounts Receivable     75,000
    CR  Revenue — AGO               75,000
```

This is correct IFRS 15 treatment. Revenue is earned when the obligation is satisfied, not when cash is received.

#### 12.2.3 Advance Payments — Deferred Revenue

A corporate customer prepays KES 500,000 for fuel to be drawn down over 3 months. At receipt of payment, no performance obligation has been satisfied — no fuel has been delivered. The payment is a liability (deferred revenue) until each draw-down.

```
On receipt of prepayment:
DR  Bank                    500,000
    CR  Deferred Revenue (Contract Liability)    500,000

On each draw-down (e.g. 10,000L AGO at KES 150):
DR  Deferred Revenue        150,000
    CR  Revenue — AGO               150,000
```

Awo manages this through a `pricing_revenue_schedule` record that tracks the outstanding deferred revenue balance and releases it proportionally with each delivery against the prepayment.

#### 12.2.4 Bundled Arrangements — Allocating Transaction Price

A workshop offers a full service package: engine oil (4L), oil filter, and labour — bundled at KES 4,500. The stand-alone prices are: oil KES 2,000, filter KES 800, labour KES 2,500 (total KES 5,300). The bundle discount of KES 800 must be allocated proportionally across the three performance obligations:

```
Standalone price total:  5,300
Bundle transaction price: 4,500
Discount to allocate:     800

Allocation:
  Oil:    2,000 / 5,300 × 4,500 = 1,698
  Filter:   800 / 5,300 × 4,500 =   679
  Labour: 2,500 / 5,300 × 4,500 = 2,123
  Total:                           4,500  ✓
```

Revenue for oil and filter is recognised when the service is completed (goods transferred). Revenue for labour is recognised when the workshop completes the work.

The Pricing module stores the `standalone_selling_price` on each bundled item entry, enabling the allocation calculation. The revenue recognition engine computes the allocation at transaction time and instructs Finance to post each obligation's revenue separately.

#### 12.2.5 Variable Consideration — Retrospective Rebates

A fleet contract includes a clause: if the customer takes ≥ 50,000 L in a quarter, they receive a retrospective rebate of KES 3/L on all AGO purchased in that quarter. At the time of each individual transaction, it is uncertain whether the customer will reach the threshold.

IFRS 15 requires the entity to estimate variable consideration and constrain the amount to the extent it is "highly probable" that a significant reversal will not occur. Practically for a petrol station this means:

- If the customer has consistently exceeded the threshold for the past 4 quarters, include the full KES 3 discount in the expected transaction price from day one of the quarter.
- If it is genuinely uncertain, recognise revenue at the higher price and accrue a rebate liability as the customer's volume accumulates.

Awo stores a `variable_consideration_policy` per contract specifying the method (`CONSTRAINED` or `UNCONSTRAINED`) and the estimation basis. The system tracks the customer's cumulative volume against the threshold throughout the quarter and, at quarter end:

- If `CONSTRAINED`: emits the credit note instruction to Finance AR for the earned rebate.
- If `UNCONSTRAINED`: the revenue was already recognised at the discounted price; no further adjustment is needed unless the customer fell short of the threshold, in which case a revenue correction is posted.

### 12.3 Revenue Recognition Schedule

For any transaction that is not immediate-recognition, the Pricing module creates a `pricing_revenue_schedule` record:

```
recognition_schedule_id
transaction_id          → link to Sales or AR transaction
total_transaction_price
recognised_to_date
outstanding_balance
schedule_lines[]
  - recognition_date
  - amount_to_recognise
  - trigger (DATE | DELIVERY | MILESTONE | MANUAL)
  - status (PENDING | RECOGNISED | REVERSED)
```

A nightly job evaluates all `PENDING` schedule lines whose `recognition_date <= today` or whose `trigger` event has fired, and posts the recognition journal to Finance GL.

---

## 13. Revenue Reporting & Analytics

### 13.1 The Revenue Intelligence Layer

Pricing decisions without feedback are blind. The revenue reporting layer closes the loop between the pricing decisions made (what price was set) and their commercial outcomes (what volume was sold, what margin was earned, where did we leave money on the table). This is not a reporting module bolted on as an afterthought — the data structures are designed from the start to support these queries efficiently.

### 13.2 Standard Reports

#### Daily Revenue Summary

- Gross revenue by product, by site, by payment method.
- Volume sold by product (links to Forecourt module's shift summaries).
- Average net price achieved per litre vs price list vs EPRA maximum.
- Margin per litre (net price − WAC) by product.
- Cash variance (from Forecourt reconciliation).

#### Price Compliance Report

- All transactions in a period, flagged against the applicable regulatory ceiling.
- Any transactions where gross price ≠ resolved price (should be zero; non-zero indicates a bypass or system error).
- Price list entry effective dates vs the first transaction using each entry (confirms price activation was timely).

#### Fleet Account Performance

- Volume by fleet account vs contracted minimum.
- Actual net price achieved vs contract price (confirms discounts are being applied correctly — neither too much nor too little).
- Rebate liability accrual vs rebate payments issued.
- Accounts at risk of volume shortfall (< 50% of monthly minimum by mid-month).

#### Promotion Effectiveness

- Volume uplift during promotion period vs same period prior year / prior period.
- Revenue impact (gross revenue at promoted price vs what would have been earned at base price).
- Estimated margin cost of the promotion.
- Redemption rate (for targeted promotions: what % of eligible customers participated).

#### Revenue Recognition Status

- Deferred revenue balance by customer, by contract.
- Recognition schedule: amounts expected to be recognised in each future month.
- Overdue recognition events (schedule lines past their recognition date not yet posted).
- Rebate accrual balance.

### 13.3 Key Metrics

These metrics are computed and stored as daily aggregates on a `pricing_daily_metrics` table to enable fast dashboard queries without scanning transaction history:

| Metric | Computation |
|---|---|
| `revenue_gross` | Sum of gross_price × quantity for the day |
| `revenue_net` | Sum of net_price × quantity (ex-tax) |
| `volume_sold_l` | Sum of quantity for fuel products (litres) |
| `avg_net_price_per_litre` | revenue_net / volume_sold_l |
| `avg_margin_per_litre` | (avg_net_price − avg_wac) |
| `promo_discount_total` | Sum of promo_discount amounts applied |
| `fleet_discount_total` | Sum of fleet/contract discount amounts applied |
| `revenue_deferred_balance` | Outstanding deferred revenue at day end |
| `rebate_accrual_balance` | Accrued but unpaid rebate liability |

---

## 14. Data Model

> Standard multi-tenancy columns (`tenant_id UUID NOT NULL` with RLS) are present on all tables and omitted from individual listings for brevity.

### 14.1 Price Lists

```sql
CREATE TABLE pricing_price_lists (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    name            VARCHAR(200) NOT NULL,
    code            VARCHAR(50)  NOT NULL,         -- CASH, FLEET, WHOLESALE, etc.
    list_type       VARCHAR(30)  NOT NULL,         -- RETAIL | FLEET | WHOLESALE | INTERNAL | PROMOTIONAL | PROCUREMENT
    currency        VARCHAR(3)   NOT NULL DEFAULT 'KES',
    parent_id       UUID REFERENCES pricing_price_lists(id),  -- for inheritance
    site_id         UUID REFERENCES sites(id),     -- NULL = tenant-wide default
    description     TEXT,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_by      UUID         NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code, site_id)
);
```

### 14.2 Price List Entries

```sql
CREATE TYPE price_entry_status AS ENUM (
    'DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'ACTIVE', 'SUPERSEDED', 'REJECTED'
);

CREATE TABLE pricing_price_list_entries (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    price_list_id       UUID NOT NULL REFERENCES pricing_price_lists(id),
    -- The item being priced. Exactly one of product_id or service_id will be set.
    product_id          UUID REFERENCES inventory_products(id),
    service_id          UUID REFERENCES services(id),
    unit_price          NUMERIC(14, 4) NOT NULL,
    currency            VARCHAR(3)     NOT NULL DEFAULT 'KES',
    -- Date effectiveness
    effective_from      TIMESTAMPTZ    NOT NULL,
    effective_to        TIMESTAMPTZ,              -- NULL = open-ended
    status              price_entry_status NOT NULL DEFAULT 'DRAFT',
    -- Regulatory context (for fuel products)
    regulated_maximum   NUMERIC(14, 4),           -- EPRA ceiling if applicable
    price_zone_id       UUID REFERENCES pricing_zones(id),
    regulation_ref      VARCHAR(100),             -- EPRA gazette notice number
    -- Standalone selling price (for bundle allocation under IFRS 15)
    standalone_price    NUMERIC(14, 4),
    -- Workflow metadata
    submitted_by        UUID REFERENCES users(id),
    submitted_at        TIMESTAMPTZ,
    approved_by         UUID REFERENCES users(id),
    approved_at         TIMESTAMPTZ,
    second_approved_by  UUID REFERENCES users(id),
    second_approved_at  TIMESTAMPTZ,
    rejection_reason    TEXT,
    change_reason       TEXT NOT NULL DEFAULT '',
    -- Prevent overlapping active entries for same item+list
    CONSTRAINT no_overlapping_active_entries EXCLUDE USING gist (
        tenant_id WITH =,
        price_list_id WITH =,
        COALESCE(product_id::TEXT, service_id::TEXT) WITH =,
        tstzrange(effective_from, effective_to, '[)') WITH &&
    ) WHERE (status = 'ACTIVE')
);
```

### 14.3 Pricing Zones (Regulatory)

```sql
CREATE TABLE pricing_zones (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    code        VARCHAR(20) NOT NULL,    -- NAIROBI, MOMBASA, KISUMU, NAKURU, etc.
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

-- Sites belong to one pricing zone
-- This is a column on the sites table: pricing_zone_id UUID REFERENCES pricing_zones(id)
```

### 14.4 Pricing Rules

```sql
CREATE TYPE pricing_rule_type AS ENUM (
    'QUANTITY_BREAK',       -- unit price drops above a quantity threshold
    'VOLUME_DISCOUNT',      -- percentage discount based on total transaction value
    'MIN_ORDER_SURCHARGE',  -- surcharge applied if order is below minimum
    'CHANNEL_PREMIUM',      -- premium for certain channels (e.g. urgent order)
    'PAYMENT_METHOD_DISCOUNT' -- discount for specific payment methods
);

CREATE TABLE pricing_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    name            VARCHAR(200)       NOT NULL,
    rule_type       pricing_rule_type  NOT NULL,
    -- Scope: which price lists and products this rule applies to
    price_list_ids  UUID[]             NOT NULL DEFAULT '{}',  -- empty = all lists
    product_ids     UUID[]             NOT NULL DEFAULT '{}',  -- empty = all products
    site_ids        UUID[]             NOT NULL DEFAULT '{}',  -- empty = all sites
    channel         VARCHAR(30),                              -- NULL = all channels
    -- Rule parameters stored as JSONB for flexibility across rule types
    -- QUANTITY_BREAK:   {"breaks": [{"min_qty": 50, "discount_pct": 3}, ...]}
    -- VOLUME_DISCOUNT:  {"breaks": [{"min_value": 5000, "discount_pct": 2}, ...]}
    -- MIN_ORDER_SURCHARGE: {"min_qty": 5, "surcharge_amount": 50}
    -- CHANNEL_PREMIUM:  {"channel": "URGENT", "premium_pct": 5}
    -- PAYMENT_METHOD_DISCOUNT: {"method": "MPESA", "discount_pct": 1}
    parameters      JSONB              NOT NULL DEFAULT '{}',
    effective_from  TIMESTAMPTZ        NOT NULL,
    effective_to    TIMESTAMPTZ,
    is_active       BOOLEAN            NOT NULL DEFAULT TRUE,
    priority        INT                NOT NULL DEFAULT 100,  -- lower = evaluated first
    created_by      UUID               NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ        NOT NULL DEFAULT now()
);
```

### 14.5 Contract / Fleet Pricing

```sql
CREATE TABLE pricing_contracts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    -- The sales contract this pricing is attached to
    sales_contract_id   UUID REFERENCES sales_contracts(id),
    customer_id         UUID NOT NULL REFERENCES customers(id),
    name                VARCHAR(200) NOT NULL,
    description         TEXT,
    effective_from      DATE NOT NULL,
    effective_to        DATE,
    -- Volume commitment (for rebate / tier evaluation)
    commitment_period   VARCHAR(20) DEFAULT 'MONTHLY',   -- MONTHLY | QUARTERLY | ANNUAL
    min_volume_l        NUMERIC(14, 2),                  -- NULL = no commitment
    -- Shortfall handling: NONE | SURCHARGE | TIER_DEMOTION
    shortfall_action    VARCHAR(20) DEFAULT 'NONE',
    shortfall_surcharge_per_l NUMERIC(12, 4),            -- if SURCHARGE
    status              VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_by          UUID NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE pricing_contract_lines (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    contract_id         UUID NOT NULL REFERENCES pricing_contracts(id),
    product_id          UUID NOT NULL REFERENCES inventory_products(id),
    -- Base reference: EPRA_MAX | PRICE_LIST | FIXED
    base_reference      VARCHAR(20) NOT NULL DEFAULT 'EPRA_MAX',
    base_price_list_id  UUID REFERENCES pricing_price_lists(id),
    -- Discount applied to base
    discount_method     VARCHAR(20) NOT NULL,  -- FIXED_AMOUNT | PERCENTAGE
    -- Volume-tiered discounts stored as JSONB:
    -- [{"min_vol_l": 0, "discount": 2.00}, {"min_vol_l": 20000, "discount": 4.00}]
    discount_tiers      JSONB NOT NULL DEFAULT '[{"min_vol_l": 0, "discount": 0}]',
    -- Variable consideration
    has_retrospective_rebate BOOLEAN NOT NULL DEFAULT FALSE,
    rebate_threshold_l  NUMERIC(14, 2),
    rebate_per_l        NUMERIC(12, 4),
    rebate_period       VARCHAR(20),            -- MONTHLY | QUARTERLY
    -- Ceiling enforcement: can contract price exceed EPRA max?
    ceiling_enforced    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 14.6 Promotions

```sql
CREATE TYPE promotion_type AS ENUM (
    'PERCENTAGE_OFF', 'FIXED_AMOUNT_OFF', 'FIXED_PRICE',
    'BUY_X_GET_Y', 'BUNDLE', 'VOLUME_BREAK', 'LOYALTY_REWARD', 'FREE_GIFT'
);

CREATE TYPE promotion_status AS ENUM (
    'DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'ACTIVE', 'PAUSED', 'EXPIRED', 'CANCELLED'
);

CREATE TABLE pricing_promotions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    name                VARCHAR(200)    NOT NULL,
    code                VARCHAR(50),               -- promotional code if customer-entered
    promotion_type      promotion_type  NOT NULL,
    status              promotion_status NOT NULL DEFAULT 'DRAFT',
    -- Temporal bounds
    starts_at           TIMESTAMPTZ     NOT NULL,
    ends_at             TIMESTAMPTZ,               -- NULL = open-ended (requires approval)
    -- Discount parameters (type-specific, validated at save time)
    -- PERCENTAGE_OFF:     {"pct": 5.0}
    -- FIXED_AMOUNT_OFF:   {"amount": 3.00, "currency": "KES"}
    -- FIXED_PRICE:        {"price": 180.00, "currency": "KES"}
    -- BUY_X_GET_Y:        {"buy_product_id": "...", "buy_qty": 20, "get_product_id": "...", "get_qty": 1, "get_discount_pct": 100}
    -- VOLUME_BREAK:       {"breaks": [{"min_qty": 50, "discount_pct": 3}]}
    parameters          JSONB           NOT NULL DEFAULT '{}',
    -- Conditions
    applies_to_product_ids   UUID[]     DEFAULT '{}',    -- empty = all products
    applies_to_category_ids  UUID[]     DEFAULT '{}',
    applies_to_customer_tiers VARCHAR(30)[] DEFAULT '{}',-- empty = all tiers
    applies_to_customer_ids  UUID[]     DEFAULT '{}',    -- empty = all customers
    applies_to_site_ids      UUID[]     DEFAULT '{}',    -- empty = all sites
    applies_to_channels      VARCHAR(30)[] DEFAULT '{}', -- empty = all channels
    applies_to_payment_methods VARCHAR(30)[] DEFAULT '{}',
    min_quantity             NUMERIC(14, 3),
    min_transaction_value    NUMERIC(14, 2),
    days_of_week             INT[]      DEFAULT '{}',    -- 0=Sun..6=Sat, empty=all days
    time_from                TIME,                       -- NULL = no time restriction
    time_to                  TIME,
    -- Stacking
    stack_group_id      UUID,                            -- NULL = not stackable
    -- Ceiling: promoted price is still subject to regulatory ceiling
    ceiling_enforced    BOOLEAN NOT NULL DEFAULT TRUE,
    -- Estimated impact (for approval tier routing)
    estimated_volume_impact_l NUMERIC(14, 2),
    estimated_revenue_impact  NUMERIC(14, 2),
    -- Workflow
    submitted_by        UUID REFERENCES users(id),
    approved_by         UUID REFERENCES users(id),
    approved_at         TIMESTAMPTZ,
    created_by          UUID NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 14.7 Price Resolution Audit

```sql
-- Every price resolution call creates one immutable row here.
-- This is the "receipt" proving what price was applied and why.
CREATE TABLE pricing_resolution_audits (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    -- Context
    product_id          UUID NOT NULL REFERENCES inventory_products(id),
    site_id             UUID NOT NULL REFERENCES sites(id),
    customer_id         UUID REFERENCES customers(id),
    customer_tier       VARCHAR(30),
    channel             VARCHAR(30),
    quantity            NUMERIC(14, 3),
    payment_method      VARCHAR(30),
    resolved_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Waterfall components
    base_price          NUMERIC(14, 4) NOT NULL,
    contract_discount   NUMERIC(14, 4) NOT NULL DEFAULT 0,
    rule_discount       NUMERIC(14, 4) NOT NULL DEFAULT 0,
    promo_discount      NUMERIC(14, 4) NOT NULL DEFAULT 0,
    net_price           NUMERIC(14, 4) NOT NULL,
    gross_price         NUMERIC(14, 4) NOT NULL,
    currency            VARCHAR(3)    NOT NULL DEFAULT 'KES',
    -- Source references
    price_list_entry_id UUID REFERENCES pricing_price_list_entries(id),
    contract_id         UUID REFERENCES pricing_contracts(id),
    promotion_ids       UUID[]        DEFAULT '{}',
    -- Regulatory status
    regulated_maximum   NUMERIC(14, 4),
    ceiling_status      VARCHAR(20),   -- OK | MARGIN_ALERT | CEILING_BREACH
    -- Link to the transaction that used this resolution (set by the Sales module on commit)
    transaction_id      UUID,
    transaction_type    VARCHAR(30),   -- SALE | SHIFT_LOCK | SALES_ORDER | INTERNAL
    -- Shift lock
    shift_context_id    UUID REFERENCES forecourt_shift_contexts(id)
    -- NEVER UPDATE. This is a permanent audit record.
);
```

### 14.8 Revenue Recognition

```sql
CREATE TYPE recognition_trigger AS ENUM (
    'DATE',         -- recognised on a specific date
    'DELIVERY',     -- recognised when a linked delivery is confirmed
    'MILESTONE',    -- recognised when a milestone event fires
    'MANUAL'        -- requires explicit Finance action
);

CREATE TABLE pricing_revenue_schedules (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID NOT NULL,
    -- Source transaction (from Sales or AR module)
    transaction_ref         VARCHAR(100) NOT NULL,
    transaction_type        VARCHAR(30)  NOT NULL,  -- SALE | PREPAYMENT | CONTRACT | BUNDLE
    customer_id             UUID         NOT NULL REFERENCES customers(id),
    total_transaction_price NUMERIC(14, 2) NOT NULL,
    recognised_to_date      NUMERIC(14, 2) NOT NULL DEFAULT 0,
    deferred_balance        NUMERIC(14, 2) NOT NULL,
    currency                VARCHAR(3)    NOT NULL DEFAULT 'KES',
    created_at              TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE TABLE pricing_revenue_schedule_lines (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    schedule_id         UUID NOT NULL REFERENCES pricing_revenue_schedules(id),
    line_number         INT  NOT NULL,
    recognition_date    DATE,
    trigger_type        recognition_trigger NOT NULL DEFAULT 'DATE',
    trigger_event_ref   VARCHAR(100),        -- e.g. delivery_id for DELIVERY trigger
    amount_to_recognise NUMERIC(14, 2)   NOT NULL,
    product_id          UUID REFERENCES inventory_products(id),
    gl_revenue_account  UUID REFERENCES gl_accounts(id),
    status              VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    -- PENDING | RECOGNISED | REVERSED | SKIPPED
    recognised_at       TIMESTAMPTZ,
    gl_journal_line_id  UUID,                -- link to Finance GL line when recognised
    reversal_reason     TEXT,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now()
);
```

### 14.9 Daily Metrics Aggregate

```sql
CREATE TABLE pricing_daily_metrics (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID NOT NULL,
    site_id                 UUID NOT NULL REFERENCES sites(id),
    product_id              UUID REFERENCES inventory_products(id),  -- NULL = site total
    metric_date             DATE NOT NULL,
    -- Volume & Revenue
    transactions_count      INT          NOT NULL DEFAULT 0,
    volume_sold_l           NUMERIC(14, 3) NOT NULL DEFAULT 0,
    revenue_gross           NUMERIC(14, 2) NOT NULL DEFAULT 0,
    revenue_net             NUMERIC(14, 2) NOT NULL DEFAULT 0,
    -- Price performance
    avg_net_price           NUMERIC(14, 4),
    avg_wac                 NUMERIC(14, 4),
    avg_margin_per_unit     NUMERIC(14, 4),
    -- Discounts
    promo_discount_total    NUMERIC(14, 2) NOT NULL DEFAULT 0,
    fleet_discount_total    NUMERIC(14, 2) NOT NULL DEFAULT 0,
    -- Deferred revenue
    deferred_revenue_recognised NUMERIC(14, 2) NOT NULL DEFAULT 0,
    deferred_revenue_balance    NUMERIC(14, 2),
    -- Regulatory
    epra_max_price          NUMERIC(14, 4),
    pct_below_ceiling       NUMERIC(7, 4),  -- how much below EPRA max is avg_net_price
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, site_id, product_id, metric_date)
);
```

---

## 15. Event Catalogue

| Event Type | Trigger | Key Payload |
|---|---|---|
| `pricing.entry.created` | New price list entry drafted | `entry_id`, `product_id`, `price_list_id`, `unit_price`, `effective_from` |
| `pricing.entry.submitted` | Entry submitted for approval | `entry_id`, `submitted_by` |
| `pricing.entry.approved` | Entry approved (may need 2nd) | `entry_id`, `approved_by`, `requires_second_approval` |
| `pricing.entry.activated` | Entry goes ACTIVE (via schedule or immediately) | `entry_id`, `product_id`, `price_list_id`, `unit_price`, `effective_from` |
| `pricing.entry.rejected` | Entry rejected in workflow | `entry_id`, `rejected_by`, `rejection_reason` |
| `pricing.entry.superseded` | Entry closed by a newer one | `entry_id`, `superseded_by_entry_id` |
| `pricing.contract.created` | New fleet contract created | `contract_id`, `customer_id` |
| `pricing.contract.activated` | Contract becomes effective | `contract_id`, `customer_id`, `effective_from` |
| `pricing.contract.expiry_warning` | 30/14/7 days to expiry | `contract_id`, `customer_id`, `days_remaining` |
| `pricing.contract.expired` | Contract past effective_to | `contract_id`, `customer_id` |
| `pricing.contract.volume_shortfall` | Customer behind minimum | `contract_id`, `customer_id`, `committed_l`, `actual_l`, `shortfall_l` |
| `pricing.promotion.activated` | Promotion goes ACTIVE | `promotion_id`, `name`, `starts_at`, `ends_at` |
| `pricing.promotion.expired` | Promotion ends | `promotion_id`, `name` |
| `pricing.compliance.ceiling_breach` | Resolved price > regulated max | `entry_id`, `product_id`, `resolved_price`, `regulated_maximum` |
| `pricing.compliance.margin_alert` | Resolved price < landed cost | `product_id`, `resolved_price`, `landed_cost` |
| `pricing.resolution.completed` | Any successful price resolution | `resolution_id`, `product_id`, `customer_tier`, `net_price` |
| `pricing.recognition.posted` | Revenue recognition schedule line recognised | `schedule_line_id`, `amount`, `gl_journal_line_id` |
| `pricing.rebate.accrued` | Rebate milestone reached | `contract_id`, `customer_id`, `rebate_amount`, `period` |
| `pricing.rebate.credit_note_requested` | Rebate ready for credit note | `contract_id`, `customer_id`, `amount` |
| `pricing.emergency_override.activated` | Emergency price change activated | `entry_id`, `activated_by`, `justification` |

---

## 16. API Surface

Base path: `/api/v1/pricing`

### 16.1 Price Lists

```
GET    /price-lists
POST   /price-lists
GET    /price-lists/{price_list_id}
PATCH  /price-lists/{price_list_id}

GET    /price-lists/{price_list_id}/entries
POST   /price-lists/{price_list_id}/entries
GET    /price-lists/{price_list_id}/entries/{entry_id}
PATCH  /price-lists/{price_list_id}/entries/{entry_id}

POST   /price-lists/{price_list_id}/entries/{entry_id}/submit
POST   /price-lists/{price_list_id}/entries/{entry_id}/approve
POST   /price-lists/{price_list_id}/entries/{entry_id}/reject
POST   /price-lists/{price_list_id}/entries/{entry_id}/emergency-activate

GET    /price-lists/{price_list_id}/entries/history?product_id=&from=&to=
```

### 16.2 Pricing Zones

```
GET    /zones
POST   /zones
GET    /zones/{zone_id}
PATCH  /zones/{zone_id}
POST   /zones/import-epra          -- import EPRA published prices for all zones
```

### 16.3 Contracts

```
GET    /contracts
POST   /contracts
GET    /contracts/{contract_id}
PATCH  /contracts/{contract_id}
POST   /contracts/{contract_id}/lines
GET    /contracts/{contract_id}/lines
PATCH  /contracts/{contract_id}/lines/{line_id}
GET    /contracts/{contract_id}/volume-summary?period=
GET    /contracts/{contract_id}/rebate-status
```

### 16.4 Promotions

```
GET    /promotions
POST   /promotions
GET    /promotions/{promotion_id}
PATCH  /promotions/{promotion_id}
POST   /promotions/{promotion_id}/submit
POST   /promotions/{promotion_id}/approve
POST   /promotions/{promotion_id}/pause
POST   /promotions/{promotion_id}/cancel
```

### 16.5 Price Resolution

```
POST   /resolve                    -- resolve a price for a given context
POST   /resolve/batch              -- resolve prices for multiple items in one call
GET    /resolve/audit/{resolution_id}  -- retrieve a specific resolution audit record
GET    /resolve/audit?transaction_id=  -- audit records for a transaction
```

### 16.6 Revenue Recognition

```
GET    /recognition/schedules
GET    /recognition/schedules/{schedule_id}
GET    /recognition/schedules/{schedule_id}/lines
POST   /recognition/schedules/{schedule_id}/lines/{line_id}/recognise  -- manual trigger
POST   /recognition/run-nightly    -- trigger nightly recognition job (admin)
GET    /recognition/deferred-balance?customer_id=&as_of=
```

### 16.7 Reports

```
GET    /reports/daily-revenue?site_id=&date=
GET    /reports/price-compliance?site_id=&from=&to=
GET    /reports/fleet-performance?from=&to=
GET    /reports/promotion-effectiveness?promotion_id=
GET    /reports/margin-analysis?site_id=&product_id=&from=&to=
GET    /reports/recognition-forecast?from=&to=
GET    /reports/epra-adherence?from=&to=
```

---

## 17. Permissions & Roles

### 17.1 Pricing Roles

| Role | Description |
|---|---|
| **Pricing Analyst** | Creates and maintains price list entries; imports EPRA prices; creates promotions. Cannot approve their own entries. |
| **Pricing Manager** | Approves medium-risk price changes; first approver for high-risk changes. |
| **Compliance Officer** | Reviews and approves EPRA regulatory price entries. |
| **Sales Manager** | Approves fleet contracts and contract pricing. |
| **Marketing Manager** | Approves promotions below the high-risk threshold. |
| **General Manager** | Second approver for high-risk price changes; approves emergency overrides. |
| **Finance Manager** | Manages revenue recognition schedules; approves internal transfer prices. |
| **Revenue Analyst** | Read access to all reports; no write access. |

### 17.2 Permission Matrix

| Permission | Analyst | Pricing Mgr | Compliance | Sales Mgr | Marketing | GM | Finance |
|---|---|---|---|---|---|---|---|
| `pricing.entry.create` | ✓ | ✓ | ✓ | | | | |
| `pricing.entry.submit` | ✓ | ✓ | ✓ | | | | |
| `pricing.entry.approve.medium` | | ✓ | | | | ✓ | |
| `pricing.entry.approve.high` | | | | | | ✓ | |
| `pricing.entry.approve.regulatory` | | | ✓ | | | | |
| `pricing.entry.reject` | | ✓ | ✓ | | | ✓ | |
| `pricing.entry.emergency_activate` | | | | | | ✓ | |
| `pricing.contract.create` | | | | ✓ | | | |
| `pricing.contract.approve` | | ✓ | | ✓ | | ✓ | |
| `pricing.promotion.create` | ✓ | | | | ✓ | | |
| `pricing.promotion.approve.medium` | | | | | ✓ | ✓ | |
| `pricing.promotion.approve.high` | | | | | | ✓ | |
| `pricing.resolution.call` | (service) | (service) | (service) | (service) | (service) | (service) | (service) |
| `pricing.recognition.view` | | | | | | | ✓ |
| `pricing.recognition.manual_post` | | | | | | | ✓ |
| `pricing.report.daily_revenue` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `pricing.report.compliance` | | ✓ | ✓ | | | ✓ | ✓ |
| `pricing.report.margin` | | ✓ | | | | ✓ | ✓ |
| `pricing.zone.manage` | | | ✓ | | | | |

---

## 18. Feature Flags & Configuration

### 18.1 Feature Flags

| Flag | Default | Effect |
|---|---|---|
| `pricing.regulated_ceiling_enforcement` | `true` | Reject resolutions above EPRA/regulated ceiling. Disabling for non-regulated tenants. |
| `pricing.floor_enforcement` | `false` | Block (not just warn) on resolutions below landed cost. |
| `pricing.contract_pricing` | `true` | Enable fleet/contract pricing engine. |
| `pricing.promotions` | `true` | Enable promotions module. |
| `pricing.promotion_stacking` | `false` | Allow promotions to stack additively. |
| `pricing.variable_consideration` | `false` | Enable retrospective rebate and variable consideration handling. |
| `pricing.revenue_recognition_schedules` | `false` | Enable IFRS 15 recognition scheduling. Required for advance payments and bundled arrangements. |
| `pricing.multi_currency` | `false` | Enable price list entries in currencies other than tenant default. |
| `pricing.price_inheritance` | `true` | Enable parent price list inheritance for site-specific lists. |
| `pricing.epra_auto_import` | `false` | Scheduled job to auto-fetch and import EPRA published prices. |
| `pricing.daily_metrics_aggregation` | `true` | Nightly job to build pricing_daily_metrics rows. |
| `pricing.two_tier_approval` | `true` | Require dual approver for high-risk changes. |

### 18.2 Tenant / Site Configuration

| Key | Type | Default | Description |
|---|---|---|---|
| `pricing.default_currency` | string | `KES` | Primary pricing currency. |
| `pricing.epra_import_day` | int | `13` | Day of month to refresh EPRA ceiling cache. |
| `pricing.price_change_notification_email` | string | — | Email to notify on any price activation. |
| `pricing.approval_timeout_hours` | int | `48` | Hours before a pending-approval entry auto-expires. |
| `pricing.shift_lock_price` | bool | `true` | Lock prices to shift-open resolution in Forecourt module. |
| `pricing.margin_alert_threshold_pct` | decimal | `0.02` | Margin below this % of revenue triggers alert. |
| `pricing.rebate_credit_note_lead_days` | int | `5` | Days before period end to raise rebate credit note. |
| `pricing.resolution_cache_ttl_seconds` | int | `60` | Active price list cache TTL in seconds. |
| `pricing.volume_tier_assessment_method` | string | `RETROSPECTIVE` | `RETROSPECTIVE` or `PROSPECTIVE` for fleet contracts. |

---

## Appendix A — Price Resolution Worked Examples

### A.1 Walk-In Cash Customer, AGO, No Promotion

```
Resolution Context:
  Product:        AGO (Diesel)
  Site:           Shell Maanzoni (Nairobi zone)
  Customer:       Anonymous
  Tier:           CASH
  Quantity:       50 L
  Channel:        FORECOURT
  Date:           2024-11-15

Step 1 — Base Price:
  Price list:     Shell Maanzoni — CASH tier
  Active entry:   KES 164.50/L  (EPRA Nairobi November 2024)
  Base price:     164.50

Step 2 — Shift Lock:   not applicable (spot resolution)
Step 3 — Contract:     no contract (anonymous customer)
Step 4 — Rules:        QUANTITY_BREAK rule: 50L is below 100L minimum → no break
Step 5 — Promotions:   no active promotions match this context

Step 6 — Ceiling Check:
  Regulated maximum:   KES 164.50 (EPRA Nairobi, November 2024)
  Resolved price:      KES 164.50
  Status:              OK (at ceiling, not above)

Step 7 — Tax:
  VAT (16%):           KES 26.32
  Petroleum Levy:      KES 21.79  (per EPRA levy schedule, included in pump price)
  Net of levies:       KES 164.50 (Kenya pump price is inclusive of all levies)
  [Note: In Kenya, EPRA publishes the all-inclusive pump price. Tax is not added on top.]
  Gross price:         KES 164.50

Result:
  Net price:    KES 164.50/L
  Gross price:  KES 164.50/L
  50L total:    KES 8,225.00
```

### A.2 Fleet Customer, Mid-Month, Volume Tier Evaluation

```
Resolution Context:
  Product:        AGO (Diesel)
  Site:           Shell Maanzoni
  Customer:       ACME Logistics Ltd
  Tier:           FLEET
  Quantity:       500 L  (this transaction)
  Month-to-date volume on contract: 18,500 L before this transaction
  Contract:       ACME Fleet Contract 2024 (prospective tier method)
  Date:           2024-11-15

Step 1 — Base Price:
  Price list:   Shell Maanzoni — CASH tier (used as reference for fleet delta)
  Base price:   KES 164.50/L

Step 3 — Contract Price:
  Contract line: AGO, EPRA_MAX − FIXED_AMOUNT, tiered discounts:
    0–10,000L/mo:    KES 2.00 discount
    10,001–30,000L:  KES 4.00 discount
    > 30,000L:       KES 6.00 discount
  
  Previous month volume: 22,000L → qualifies for 10,001–30,000 tier this month
  Prospective discount:  KES 4.00/L
  Contract price:        KES 164.50 − 4.00 = KES 160.50/L

Step 4 — Rules:     QUANTITY_BREAK: 500L > 100L minimum → 3% break:
  3% of 160.50 = KES 4.82/L
  Rule-adjusted: KES 160.50 − 4.82 = KES 155.68/L

Step 5 — Promotions:   none active for FLEET tier

Step 6 — Ceiling:
  Regulated maximum:   KES 164.50
  Resolved price:      KES 155.68
  Status:              OK
  Margin check:        WAC = KES 148.20. Margin = KES 7.48/L (4.8%). OK.

Result:
  Net price:    KES 155.68/L
  500L total:   KES 77,840.00
  Discount vs CASH:  KES 4,410.00 (4.32% of CASH revenue — logged in fleet_discount_total)
```

### A.3 Emergency EPRA Price Change — Mid-Month Correction

```
Scenario:
EPRA publishes a correction on 2024-11-20 reducing the AGO maximum for Nairobi
from KES 164.50 to KES 161.30, effective immediately (intra-month correction — rare).

Action sequence:
1. Compliance officer is notified (via industry channel).
2. Opens the price list entry form, imports the correction.
3. System routes to EMERGENCY workflow (because effective date = today = immediate).
4. GM approves (mandatory for emergency activations).
5. Entry is activated: effective_from = 2024-11-20T14:30:00.
6. Previous entry (KES 164.50) is closed: effective_to = 2024-11-20T14:30:00.
7. Events emitted:
   - pricing.entry.activated → Forecourt module updates pump display at next transaction.
   - pricing.emergency_override.activated → All managers notified.
8. From 14:30:00 onwards, all resolutions for AGO return KES 161.30.

Transactions between 00:00 and 14:30 today were correctly priced at KES 164.50
(the regulation changed at publication time, not retroactively).
No adjustment required for earlier transactions.
```

---

## Appendix B — IFRS 15 Revenue Recognition Scenarios

### B.1 Prepayment Drawdown

```
Scenario:
A corporate customer pays KES 1,000,000 in advance on 2024-11-01.
They draw down fuel over November and December.

On 2024-11-01 (receipt):
  DR  Bank                          1,000,000
      CR  Deferred Revenue (Contract Liability)     1,000,000

Revenue schedule created:
  Total:    KES 1,000,000
  Trigger:  DELIVERY (recognised as each delivery is confirmed)

On 2024-11-15 (delivery: 3,000L AGO at KES 164.50 = KES 493,500):
  DR  Deferred Revenue             493,500
      CR  Revenue — AGO                            493,500
  Schedule: recognised KES 493,500, outstanding KES 506,500

On 2024-12-10 (delivery: 3,000L AGO at KES 161.30 = KES 483,900):
  DR  Deferred Revenue             483,900
      CR  Revenue — AGO                            483,900
  Schedule: recognised KES 977,400, outstanding KES 22,600

On 2024-12-31 (remaining balance returned to customer, contract expired):
  DR  Deferred Revenue              22,600
      CR  Bank (refund)                              22,600
  Schedule: CLOSED
```

### B.2 Bundled Workshop Service

```
Scenario:
Full service package sold on 2024-11-15 for KES 4,500.
Components:
  - Engine oil 5W-30 4L   (standalone KES 2,000 — delivered immediately)
  - Oil filter             (standalone KES 800 — delivered immediately)
  - Labour (oil change)    (standalone KES 2,500 — service performed same day)
Total standalone: KES 5,300

Allocation of KES 4,500 transaction price:
  Oil:    2,000 / 5,300 × 4,500 = KES 1,698.11  → recognised at delivery (goods)
  Filter:   800 / 5,300 × 4,500 = KES   679.25  → recognised at delivery (goods)
  Labour: 2,500 / 5,300 × 4,500 = KES 2,122.64  → recognised on workshop sign-off

All three obligations satisfied on 2024-11-15:

  DR  Cash                          4,500.00
      CR  Revenue — Lubricants               1,698.11
      CR  Revenue — Parts                      679.25
      CR  Revenue — Workshop Labour          2,122.64
```

---

*End of Pricing & Revenue Management Module Specification v1.0.0*
