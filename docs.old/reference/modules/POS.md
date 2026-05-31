# AwoERP POS Terminal
## Comprehensive Technical & Product Documentation
**Version:** 1.0.0-draft  
**Last Updated:** 2026-05-11  
**Status:** Pre-release  

---

> **How to read this document**
>
> This documentation serves two audiences simultaneously. Sections marked with a
> **👤 Non-Technical Reader** callout explain concepts without assuming engineering
> background. Sections without the callout assume familiarity with software
> development. You can read selectively — jump to any section independently.
>
> **Release tagging:** Features are tagged `[v1.0.0]` (must ship), `[v1.x]`
> (near-future minor release), or `[Future]` (roadmap consideration).

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Feature Overview & Release Prioritization](#2-feature-overview--release-prioritization)
3. [User Interface & Experience](#3-user-interface--experience)
4. [Offline-First Architecture](#4-offline-first-architecture)
5. [Local Network & Multi-Terminal Support](#5-local-network--multi-terminal-support)
6. [Proposed Tech Stack](#6-proposed-tech-stack)
7. [System Architecture](#7-system-architecture)
8. [Data Model](#8-data-model)
9. [Core Features — Deep Dive](#9-core-features--deep-dive)
10. [Deployment & Operations](#10-deployment--operations)
11. [Developer Guide](#11-developer-guide)
12. [Appendices](#12-appendices)

---

## 1. Introduction

### 1.1 What is AwoERP POS?

AwoERP POS is an industry-agnostic, offline-first Point of Sale terminal system
designed to serve businesses of all types — restaurants, retail shops, service
providers, pharmacies, markets, and beyond. It is the transactional front end of
the broader AwoERP ecosystem, meaning it handles the moment a sale happens: the
product selection, payment collection, receipt generation, and real-time inventory
adjustment.

Unlike cloud-only POS systems that become inoperable when the internet goes down,
AwoERP POS stores all operational data locally on each terminal and synchronises
seamlessly with the central database when connectivity is available. Multiple
terminals on the same local network share live inventory and configuration
without relying on an external server.

> **👤 Non-Technical Reader**
> Think of AwoERP POS as a smart cash register that keeps working even without
> the internet, automatically keeps itself in sync with your back office, and
> can talk to other registers in the same building — all without you having to
> think about any of it.

### 1.2 Design Philosophy

AwoERP POS is built around five principles:

**Resilience over convenience.** The system must continue operating under degraded
conditions — no internet, slow networks, partial hardware failures. Online
connectivity is treated as a bonus, never a requirement for basic operation.

**Industry neutrality.** The core engine is business-logic agnostic. Industry-specific
behaviour (e.g., table management for restaurants, weight-based pricing for
markets) is delivered as pluggable configuration profiles, not hard-coded
assumptions.

**Operator simplicity.** A cashier with 15 minutes of training should be fully
productive. Power features exist for managers but never clutter the default view.

**Local-first data ownership.** Business data is owned by the operator, not a vendor.
All data is stored locally and exportable at any time.

**Deployability.** Installation should be achievable by a non-engineer. The system
should run on modest, low-cost hardware without complex infrastructure.

### 1.3 Who This Documentation Is For

| Audience | Recommended Sections |
|---|---|
| Business Owner / Operator | §1, §2, §3, §9 (overview only), §10.1–10.2 |
| System Administrator / IT | §4, §5, §7, §10 (full), §11 |
| Backend / Go Developer | §6, §7, §8, §11 (full) |
| Frontend / UI Developer | §3, §6.4, §11 |
| Technical Evaluator | §2, §6, §7, Appendix B |

### 1.4 Versioning Strategy & Release Philosophy

AwoERP POS follows **Semantic Versioning (SemVer)**:

```
MAJOR.MINOR.PATCH
  │      │     └─ Bug fixes, no new features
  │      └─────── New features, backward-compatible
  └────────────── Breaking changes, architecture shifts
```

**v1.0.0** targets the minimum viable system that a real business can operate
profitably. It is deliberately scoped to avoid over-engineering. Every feature
not in v1.0.0 has been consciously deferred, not forgotten.

The guiding question for v1.0.0 inclusion is: *"Would a business refuse to open
without this?"* If yes, it ships in v1. If it makes the product better but the
business can still open, it is deferred.

---

## 2. Feature Overview & Release Prioritization

### 2.1 Feature Matrix

| Feature | Category | v1.0.0 | Rationale |
|---|---|:---:|---|
| Product catalogue & search | Core | ✅ | Cannot sell without it |
| Cart & order management | Core | ✅ | Cannot sell without it |
| Cash payment | Payment | ✅ | Most basic payment method |
| Card payment (integrated) | Payment | ✅ | Majority of transactions |
| Mobile money / QR payment | Payment | v1.x | Regional relevance; API-dependent |
| Split payment | Payment | ✅ | Common in hospitality |
| Offline operation | Architecture | ✅ | Core design pillar |
| Local network sync | Architecture | ✅ | Multi-terminal businesses |
| Cloud sync | Architecture | ✅ | Back-office visibility |
| Receipt printing | Output | ✅ | Legal/customer requirement |
| Digital receipt (email/SMS) | Output | v1.x | Nice-to-have; adds dependency |
| Barcode scanning | Input | ✅ | Retail essential |
| Manual product entry | Input | ✅ | Fallback, services |
| Discount & promotion engine | Sales | ✅ | Basic discounts only in v1 |
| Loyalty / points system | Sales | Future | Significant complexity |
| Tax configuration | Compliance | ✅ | Legal requirement |
| Multi-currency | Compliance | v1.x | Region-dependent |
| Role-based access (Cashier/Manager/Admin) | Security | ✅ | Operational requirement |
| Shift management & cash drawer | Operations | ✅ | Required for reconciliation |
| Real-time inventory deduction | Inventory | ✅ | Prevents overselling |
| Purchase order / receiving | Inventory | Future | Back-office, not POS scope |
| Daily sales reports | Reporting | ✅ | Manager daily need |
| Advanced analytics | Reporting | Future | Valuable but not day-one |
| Table/floor management | Industry | v1.x | Restaurant-specific |
| Kitchen display system (KDS) | Industry | v1.x | Restaurant-specific |
| Weight-based pricing | Industry | v1.x | Market/deli-specific |
| Plugin / extension API | Platform | Future | After core stabilises |
| Multi-language UI | Accessibility | v1.x | Internationalisation |
| Customer management (CRM-lite) | CRM | v1.x | Useful but not blocking |

### 2.2 Industry-Agnostic Design Rationale

Most POS systems are designed around one industry and awkwardly extended to others.
AwoERP POS inverts this — the core engine handles universal commerce primitives
(products, quantities, prices, payments), and an **Industry Profile** layer
adds terminology, UI layouts, and business rules on top.

```
┌─────────────────────────────────────────────────┐
│              Industry Profile Layer              │
│  [Restaurant]  [Retail]  [Services]  [Custom]   │
├─────────────────────────────────────────────────┤
│               Core POS Engine                    │
│  Orders · Products · Payments · Inventory       │
│  Sync · Offline · Roles · Reporting             │
└─────────────────────────────────────────────────┘
```

This means a single codebase deploys to a café, a pharmacy, and a clothing store
with only configuration changes.

### 2.3 Competitive Landscape Snapshot

> **👤 Non-Technical Reader**
> This section compares AwoERP POS to well-known systems you may have heard of.
> Understanding where they succeed and where they fall short helps explain why
> certain design decisions were made.

**Square POS**
- Strengths: Excellent free tier, beautiful hardware, seamless card processing,
  strong mobile app.
- Weaknesses: Offline mode is limited (read-only or card-not-present only),
  heavy cloud dependency, limited customisation without developer access,
  hardware ecosystem is proprietary.
- Lesson for AwoERP: Square proves that simplicity wins adoption. The UI bar
  is high.

**Toast POS** *(Restaurant-focused)*
- Strengths: Deep restaurant workflows (table management, KDS, modifiers),
  purpose-built Android hardware, solid offline support for restaurant operations.
- Weaknesses: Expensive ($110+/month), restaurant-only, difficult to customise
  for other industries, complex initial setup.
- Lesson for AwoERP: Restaurant-specific depth is valuable — but only if delivered
  through a modular profile system, not baked into the core.

**Lightspeed Retail / Restaurant**
- Strengths: Comprehensive inventory management, strong e-commerce integration,
  mature multi-location support.
- Weaknesses: Expensive, requires constant connectivity for most features,
  offline mode is best-effort, UI is complex for basic cashiers.
- Lesson for AwoERP: Inventory depth matters for retail. Offline cannot be
  an afterthought.

**Shopify POS**
- Strengths: Tight e-commerce integration, excellent for omnichannel retail,
  clean UI, good reporting.
- Weaknesses: Offline support is very limited (only processes previously loaded
  items), no meaningful multi-terminal local sync, requires Shopify subscription.
- Lesson for AwoERP: E-commerce integration is future scope; offline-first is
  non-negotiable from day one.

**Summary Table**

| Capability | Square | Toast | Lightspeed | Shopify POS | AwoERP (target) |
|---|:---:|:---:|:---:|:---:|:---:|
| True offline operation | ⚠️ | ✅ | ⚠️ | ❌ | ✅ |
| Local network multi-terminal | ❌ | ✅ | ⚠️ | ❌ | ✅ |
| Industry-agnostic | ✅ | ❌ | ⚠️ | ⚠️ | ✅ |
| Self-hostable | ❌ | ❌ | ❌ | ❌ | ✅ |
| Open data export | ⚠️ | ⚠️ | ✅ | ⚠️ | ✅ |
| Low-cost hardware support | ✅ | ❌ | ❌ | ⚠️ | ✅ |
| Free tier / open core | ✅ | ❌ | ❌ | ❌ | TBD |

*✅ = Yes  ⚠️ = Partial  ❌ = No*

---

## 3. User Interface & Experience

### 3.1 Design Principles

`[v1.0.0]`

**Speed above decoration.** Every tap or click must reach its destination in at
most two interactions. A cashier serving ten customers per hour cannot afford
deep menu hierarchies.

**Glanceability.** Cart totals, stock indicators, and payment status must be
readable at arm's length on a 10-inch screen.

**Error forgiveness.** Mistakes must be correctable without supervisor intervention
for 90% of cases. The system should never lock a cashier out of a correction.

**Touch-first, keyboard-optional.** The primary interaction model is a touchscreen.
Hardware keyboard and barcode scanner are first-class input methods, never
afterthoughts.

**Accessible contrast.** All UI elements meet WCAG AA contrast ratios. This aids
readability in bright retail environments and accommodates operators with varying
visual acuity.

> **👤 Non-Technical Reader**
> The screen design follows the same idea as a well-laid-out physical store:
> the things you use most often are closest at hand, and you never have to go
> hunting for the basics.

### 3.2 Terminal Layout & Navigation Flows

The terminal UI is divided into three persistent zones:

```
┌──────────────────────────────────────────────────────────────────┐
│  HEADER BAR                                                       │
│  [☰ Menu]  AwoERP POS          [Cashier: Amara] [12:34] [● Sync] │
├──────────────────────┬───────────────────────────────────────────┤
│                      │                                           │
│   PRODUCT PANEL      │           CART PANEL                      │
│                      │                                           │
│  [Search / Scan]     │  Item             Qty    Price            │
│                      │  ─────────────────────────────            │
│  ┌────┐ ┌────┐       │  Espresso          1     KES 250          │
│  │ 🥤 │ │ 🍕 │       │  Croissant         2     KES 300          │
│  └────┘ └────┘       │  ─────────────────────────────            │
│  ┌────┐ ┌────┐       │                                           │
│  │ 📦 │ │ 🧴 │       │  Subtotal               KES 550           │
│  └────┘ └────┘       │  Tax (16%)               KES 88           │
│                      │  ─────────────────────────────            │
│  [Category Filter]   │  TOTAL                  KES 638           │
│                      │                                           │
├──────────────────────┤  [Discount]  [Hold]  [Void]               │
│  NUMPAD / SHORTCUTS  ├───────────────────────────────────────────┤
│  [1][2][3]  [Clear]  │          [ CHARGE →  KES 638 ]            │
│  [4][5][6]  [×Qty ]  │                                           │
│  [7][8][9]  [Price]  │                                           │
│     [0]     [Enter]  │                                           │
└──────────────────────┴───────────────────────────────────────────┘
```

**Key navigation flows:**

*Standard sale:*
Scan/tap product → Adjust quantity if needed → Tap CHARGE → Select payment
method → Confirm → Print/send receipt → Next transaction.

*Discount application:*
Select line item or cart → Tap Discount → Enter percentage or fixed amount →
Confirm (requires manager PIN for discounts above configured threshold).

*Hold and recall:*
Tap Hold → Assign identifier → Serve next customer → Recall held order from
header menu → Continue or void.

*End of shift:*
Menu → Shift → Close Shift → Count cash drawer → Submit report → Confirm.

### 3.3 Role-Based Views

`[v1.0.0]`

| Role | Capabilities |
|---|---|
| **Cashier** | Process sales, apply pre-approved discounts, accept payments, print receipts, handle returns (within policy), view own shift summary |
| **Supervisor** | All cashier actions + approve large discounts, void transactions, access any terminal's current cart, view live floor summary |
| **Manager** | All supervisor actions + configure products/pricing, manage users, run reports, open/close day, adjust inventory |
| **Admin** | All manager actions + system configuration, sync settings, hardware setup, data export |

Role switching is done via PIN entry, not full logout. A cashier can request
supervisor approval without leaving the current sale.

```
[Cashier View]                         [Manager View — additional menu items]
─────────────────────────────────      ───────────────────────────────────────
 Products · Cart · Payments             + Product Management
 Hold Orders · Receipts                 + Pricing & Discounts
 My Shift Summary                       + User Management
                                        + Inventory Adjustments
                                        + Reports & Exports
                                        + System Configuration
```

### 3.4 Industry Configuration Modes

`[v1.x]` *(Profile system; core layout ships in v1.0.0)*

When setting up AwoERP POS, an operator selects an Industry Profile. This
changes default terminology, layout emphasis, and available features — but not
the underlying engine.

| Profile | Terminology changes | Layout emphasis | Extra features |
|---|---|---|---|
| **Retail** | Items, SKU, Stock | Barcode scan prominent, category grid | Stock level badge on products |
| **Restaurant (QSR)** | Menu items, Combos | Large product tiles, modifier flow | Order number display, kitchen printer routing |
| **Restaurant (Full Service)** | Menu, Table, Cover | Table map view, course management | Table assignment, split-by-seat |
| **Services** | Services, Bookings | Simple item list, time-based | Service duration, staff assignment |
| **Market / Wholesale** | Products, Weight, Bulk | Weight input prominent | Price-per-kg entry, bulk discount |
| **Custom** | Configurable | Configurable | Mix-and-match |

### 3.5 Competitive UI Analysis

| System | Strengths | Weaknesses | AwoERP Lesson |
|---|---|---|---|
| **Square** | Minimal cognitive load, fast checkout flow, excellent typography | Not customisable, icons can be ambiguous without labels | Copy the 2-tap checkout principle |
| **Toast** | Dense information without feeling crowded, modifiers are elegant | Onboarding complex, takes days to train | Offer a "simple mode" that hides advanced features by default |
| **Lightspeed** | Rich product data visible at a glance | Too many columns, overwhelming for basic cashiers | Progressive disclosure: show advanced data on demand |
| **Shopify POS** | Clean, consistent with brand | Too minimalist, hides important info like stock levels | Always surface stock count on product tiles |

---

## 4. Offline-First Architecture

### 4.1 Why Offline-First Matters

> **👤 Non-Technical Reader**
> Most modern software is "online-first" — it assumes your internet is always
> working, and stops functioning (or behaves unpredictably) when it isn't.
> AwoERP POS is built the other way around: it assumes your internet might be
> unreliable, and treats the online connection as a helpful bonus rather than
> a hard requirement. This means your business keeps running during an outage,
> and your data catches up automatically when the connection returns.

Internet outages are not edge cases in real business operations. Router reboots,
ISP instability, underground cable faults, power fluctuations, and mobile data
dead spots are daily realities — especially in markets outside North America and
Western Europe.

A POS system that goes offline during lunch rush is not just inconvenient; it
is a business liability. The offline-first design ensures:

- **Zero downtime sales**: Every sale can be completed without internet access.
- **No data loss**: All transactions are persisted locally before any
  acknowledgement is sent to the customer.
- **Predictable behaviour**: The system behaves identically online and offline.
  There is no "offline mode" the cashier switches into — it just works.
- **Graceful sync**: When connectivity returns, data synchronises automatically
  with no manual intervention.

**What "offline" means in AwoERP POS:**

| Condition | System behaviour |
|---|---|
| Internet down, LAN up | Full operation; syncs to other local terminals |
| Internet down, LAN down | Full operation on current terminal; syncs on reconnection |
| Internet up, LAN down | Full operation; syncs directly to cloud |
| Internet up, LAN up | Full operation; syncs to both local terminals and cloud |
| No connectivity at all | Full operation; queues all changes for later sync |

### 4.2 Local Data Store Design

`[v1.0.0]`

Each terminal maintains a complete local copy of all data it needs to operate.
This is not a cache — it is the authoritative operational store. The cloud is
a replica and reporting layer, not the source of truth for active sales.

**Local store contents:**

```
LOCAL TERMINAL DATABASE
├── products/           — Full product catalogue with prices
├── inventory/          — Current stock levels (terminal-local delta)
├── orders/             — All orders (open, held, completed, voided)
├── payments/           — Payment records with idempotency keys
├── customers/          — Customer profiles (if CRM-lite enabled)
├── config/             — Business rules, tax rates, discount policies
├── sync_log/           — Outbound change queue
├── users/              — Hashed credentials for offline auth
└── sessions/           — Shift records
```

**Technology choice:** SQLite is used as the embedded local store. It is a
battle-tested, single-file, zero-administration database that runs on every
platform. The file can be backed up with a simple file copy.

**Why not just use files or memory?**
SQLite provides ACID transactions — meaning if the terminal loses power mid-sale,
the transaction either completed fully or not at all. Partial writes are
impossible. This is the same guarantee a bank's ATM provides.

### 4.3 Conflict-Free Replicated Data Types (CRDTs) — Theory & Practice

`[v1.0.0]`

> **👤 Non-Technical Reader**
> Imagine two cashiers on different registers sell the last two bottles of
> juice at exactly the same moment, while offline. When they reconnect,
> how does the system know which sale was "right"? CRDTs are a mathematical
> approach to resolving these conflicts automatically — without losing either
> sale and without requiring human intervention.

**The core problem:**

When multiple terminals modify the same data independently (because they are
offline from each other), those modifications may conflict when synchronised.
Naive "last write wins" approaches lose data. Manual conflict resolution is
impractical at transaction scale.

CRDTs are data structures that are mathematically designed to merge their
state without conflicts. They achieve this by encoding not just the current
value but the history of operations.

**CRDTs used in AwoERP POS:**

*G-Counter (Grow-only Counter)*
Used for: total units sold, transaction count.
Properties: Only increments, never decrements. Merges by taking the maximum
of each terminal's count.

```
Terminal A count: 5  →  Merge result: 7
Terminal B count: 7  →  (take max per terminal)
```

*PN-Counter (Positive-Negative Counter)*
Used for: inventory levels (sales decrease, receiving increases).
Properties: Two G-counters internally (additions and subtractions).

```
Initial stock: 100 units
Terminal A sold 3  →  A's subtract counter: 3
Terminal B sold 2  →  B's subtract counter: 2
After merge:   100 - 3 - 2 = 95  ✅ (not 100 - 2 = 98, as last-write-wins would give)
```

*LWW-Register (Last-Write-Wins Register)*
Used for: product name, price (non-concurrent changes).
Properties: Uses timestamps + terminal ID as tiebreaker for deterministic
resolution. Acceptable for configuration data that changes infrequently.

*LWW-Element-Set*
Used for: order line items (add/remove modifiers, products).
Properties: Tracks both additions and deletions with timestamps. An item
deleted after being added remains deleted even if a stale add arrives later.

**When CRDTs are not sufficient:**

Some operations have business constraints that CRDTs cannot enforce purely
mathematically. For these, AwoERP uses **operational transforms** with
business-rule validation at sync time:

- *Negative inventory*: If merged inventory goes negative, a warning is
  generated and a supervisor notification is queued. The transaction is never
  rejected retroactively (the customer already left), but the discrepancy is
  surfaced.
- *Duplicate payment IDs*: Payments use UUIDs with terminal prefix, making
  cross-terminal duplicates impossible by construction.

### 4.4 Sync Engine: Strategies, Triggers & Failure Recovery

`[v1.0.0]`

**Sync is continuous, not scheduled.**

The sync engine runs in the background at all times. It does not wait for a
"sync now" button. Whenever connectivity exists (internet or LAN), changes are
streamed outward and inbound updates are received.

**Sync trigger hierarchy:**

```
1. Transaction completion  → Immediate push attempt
2. Connectivity restored   → Flush entire queue, pull latest state
3. Periodic heartbeat      → Every 30 seconds when idle (configurable)
4. Manual force-sync       → Manager-initiated, for troubleshooting
```

**Change queue structure:**

Each change is recorded as an immutable event with metadata before any attempt
to propagate it:

```json
{
  "event_id": "trm-A-1716480000-0042",
  "terminal_id": "TRM-A",
  "timestamp": "2026-05-11T10:00:00.000Z",
  "vector_clock": {"TRM-A": 42, "TRM-B": 38},
  "entity_type": "order",
  "entity_id": "ord-20260511-0088",
  "operation": "complete",
  "payload": { ... },
  "checksum": "sha256:abc123..."
}
```

The `vector_clock` records the logical time at each terminal, enabling the
sync engine to detect and correctly order concurrent events even when wall
clocks differ.

**Failure recovery:**

```
SYNC ATTEMPT
     │
     ├─ Success ──→ Mark event as synced, remove from queue
     │
     ├─ Transient failure (network) ──→ Exponential backoff: 1s, 2s, 4s, 8s...
     │                                   Max interval: 5 minutes
     │                                   Queue persists across restarts
     │
     ├─ Business rule conflict ──→ Log conflict, notify manager, mark as
     │                             "needs review", continue other events
     │
     └─ Permanent failure (data corruption) ──→ Quarantine event, alert admin,
                                                 never discard silently
```

**Queue persistence:**
The sync queue is written to SQLite with the same ACID guarantees as orders.
A terminal can be restarted, updated, or lose power and its queue will be
intact when it comes back up. Events are never discarded until acknowledged
by the receiving system.

### 4.5 What Works Offline vs. What Requires Connectivity

| Feature | Offline | Notes |
|---|:---:|---|
| Process sales (cash) | ✅ | Full functionality |
| Process sales (card — stored credentials) | ✅ | With pre-authorised payment provider SDK |
| Process sales (card — new card) | ⚠️ | Depends on payment terminal capability |
| Recall held orders | ✅ | Stored locally |
| Apply discounts | ✅ | Rules stored locally |
| Print receipts | ✅ | Local printer |
| Email/SMS receipts | ❌ | Queued, sent on reconnection |
| Check stock levels (own terminal) | ✅ | Local inventory |
| Check stock levels (other terminals) | ❌ | Requires LAN or internet |
| Manager reports | ✅ | Local data only; cloud data after sync |
| Product catalogue updates | ⚠️ | Apply on next sync |
| Price changes | ⚠️ | Apply on next sync |
| New user login (unknown user) | ❌ | Credentials must have been synced |
| Add new product | ✅ | Queued for sync |
| Refunds | ✅ | Logged; cloud reconciliation on sync |

### 4.6 Competitive Analysis: Offline Support

| System | Offline capability | Conflict strategy | Assessment |
|---|---|---|---|
| **Square** | Card-not-present only ($200 limit cap per transaction); product list read-only | Last-write-wins | Inadequate for true offline-first operation |
| **Toast** | Full restaurant operation offline; syncs automatically | Proprietary; operation-log based | Industry-leading for restaurants; approach inspires AwoERP |
| **Lightspeed** | "Offline mode" available but must be pre-enabled; limited to pre-loaded products | Last-write-wins with manual review | Better than Square but still requires preparation |
| **Shopify POS** | Read-only offline; cannot complete new sales with unloaded products | N/A (minimal offline) | Not suitable for offline-first requirements |

---

## 5. Local Network & Multi-Terminal Support

### 5.1 Network Topology

`[v1.0.0]`

> **👤 Non-Technical Reader**
> In a shop with multiple registers, all registers need to agree on how much
> stock is left. If register 1 sells the last item, register 2 should know
> immediately — even if the internet is down. AwoERP terminals form a small
> local team, talking directly to each other over your shop's Wi-Fi or
> ethernet, so they stay in sync without needing the outside world.

AwoERP POS supports two local network topologies:

**Primary: Hub-and-Spoke (recommended for v1.0.0)**

One terminal (or a dedicated lightweight device like a Raspberry Pi) acts as the
Local Coordinator — the hub. All other terminals (spokes) connect to it.

```
                    ┌─────────────────────┐
                    │    Cloud Server      │
                    │  (sync when online)  │
                    └──────────┬──────────┘
                               │ WAN
                    ┌──────────┴──────────┐
                    │  Local Coordinator   │
                    │  (SQLite + Sync Hub) │
                    └──┬──────┬──────┬───┘
                       │ LAN  │      │
              ┌────────┘      │      └────────┐
              │               │               │
       ┌──────┴─────┐  ┌──────┴─────┐  ┌─────┴──────┐
       │ Terminal 1  │  │ Terminal 2  │  │ Terminal 3  │
       │  (Cashier)  │  │  (Cashier)  │  │  (Manager)  │
       └────────────┘  └────────────┘  └────────────┘
```

Advantages: Simple to reason about, single point of truth on LAN, easy to
monitor, single sync agent handles cloud communication.

**Alternative: Peer-to-Peer (for resilience / Future)**

Each terminal maintains connections to all other terminals. No single point of
failure. Significantly more complex to implement correctly.

```
     T1 ─────── T2
      │  ╲   ╱  │
      │    ╳    │
      │  ╱   ╲  │
     T3 ─────── T4
          │
        Cloud
```

*Deferred to v1.x due to complexity of distributed coordination without a
coordinator.*

### 5.2 Terminal Discovery & Registration

`[v1.0.0]`

Terminals discover each other automatically using **mDNS (multicast DNS)** — the
same technology that makes printers appear automatically on home networks.

**Discovery flow:**

```
NEW TERMINAL BOOTS
        │
        ▼
  Broadcasts mDNS announcement:
  "_awoerp._tcp.local"
  { terminal_id, version, role }
        │
        ▼
  Local Coordinator receives announcement
        │
        ├─ Terminal not registered?
        │       └─ Prompt admin: "New terminal detected. Approve?"
        │
        └─ Terminal registered?
                └─ Exchange sync state, establish connection
```

**Security:**
All LAN communication is encrypted with TLS using certificates issued by the
Local Coordinator as the internal CA. No terminal can join the network without
a certificate signed by the coordinator. Certificates are provisioned during
the admin registration flow.

**Registration steps for a new terminal (admin perspective):**
1. Install AwoERP POS on new device.
2. Start the application — it announces itself on the LAN.
3. On any existing terminal, a notification appears: "New terminal [name]
   wants to join. Approve?"
4. Admin approves using their PIN.
5. New terminal downloads initial data from coordinator.
6. Ready to use within 2–5 minutes depending on catalogue size.

### 5.3 Shared State: Inventory, Pricing & Business Logic Sync

`[v1.0.0]`

**What is shared across all terminals in real time:**

| Data | Sync frequency | Direction |
|---|---|---|
| Inventory levels (delta changes) | Per transaction | All terminals ↔ Coordinator |
| Product catalogue | On change | Coordinator → All terminals |
| Pricing & discounts | On change | Coordinator → All terminals |
| Tax configuration | On change | Coordinator → All terminals |
| Active promotions | On change | Coordinator → All terminals |
| Shift open/close events | Per event | Terminal → Coordinator |
| Open/held orders | Per change | Terminal → Coordinator |

**What is terminal-local (not shared in real time):**

- Current cashier's in-progress cart (shared only on Hold)
- Local shift cash-drawer counts
- Terminal-specific configuration (printer settings, hardware)

**Inventory propagation example:**

```
T1 sells 3 units of SKU-001:
  T1 local stock: 20 → 17
  T1 sends delta: {sku: "SKU-001", delta: -3, at: vector_clock_T1}

Coordinator receives:
  Master stock: 20 → 17
  Broadcasts to T2, T3: {sku: "SKU-001", delta: -3}

T2, T3 receive:
  T2 local stock: 20 → 17
  T3 local stock: 20 → 17

T2 sells 2 units of SKU-001 simultaneously (before T1's delta arrived):
  T2 local stock (stale): 20 → 18
  After merging T1's delta: 18 - 3 = 15  ← PN-Counter merge (correct)
```

### 5.4 Conflict Resolution in Multi-Terminal Scenarios

`[v1.0.0]`

**Case 1: Concurrent stock deduction**

Both T1 and T2 sell the last item simultaneously (coordinator unreachable):

- Both sales complete locally (neither can be rejected retroactively).
- On sync, PN-Counter merge results in stock = -1.
- Coordinator raises a **stock discrepancy alert** to the manager.
- Manager resolves: adjust stock, trigger re-order, or accept the discrepancy.
- Neither customer's sale is reversed — business integrity preserved.

**Case 2: Price change during active sale**

Manager changes product price during T1's active cart:

- T1's in-progress cart retains the price at the time the item was added.
- New sales on T1 (after cart is charged) use the new price.
- No retroactive changes to committed transactions.

**Case 3: Order status collision**

T1 marks order #42 as "completed". T2 simultaneously marks it "voided"
(supervisor error):

- Both events carry timestamps and terminal IDs.
- Business rule: "voided" always wins over "completed" if void timestamp is
  within 5 minutes of completion (configurable grace period).
- Outside grace period: "completed" wins; void requires manual override.

### 5.5 Network Failure Handling

```
LAN FAILURE DETECTED
        │
        ├─ Terminal has cached state from < 30 seconds ago?
        │       └─ Continue operating; display "LAN: reconnecting" indicator
        │
        ├─ Terminal has cached state from > 30 seconds ago?
        │       └─ Display "Inventory data may be stale" warning
        │          Continue operating with local data
        │
        └─ Terminal has no coordinator connection?
                └─ Operate fully independently
                   Queue all changes for sync on reconnection
                   Display "Offline mode" indicator in header bar
```

**Reconnection sequence:**
1. Connection restored → Exchange vector clocks with coordinator.
2. Coordinator determines which events each side is missing.
3. Missing events replayed in causal order.
4. Conflicts resolved using CRDT merge rules.
5. Business-rule conflicts queued for manager review.
6. "Online" indicator restored. Normal operation resumes.

### 5.6 Competitive Analysis: Multi-Terminal Support

| System | Multi-terminal | Local sync | Offline multi-terminal |
|---|---|---|---|
| **Square** | Yes (cloud-managed) | No | No — each terminal is independent and blind to others |
| **Toast** | Yes (local + cloud) | Yes | Yes — strongest competitor here |
| **Lightspeed** | Yes (cloud-managed) | No | No |
| **Shopify POS** | Yes (cloud-managed) | No | No |
| **AwoERP target** | Yes (local + cloud) | Yes | Yes |

Toast is the benchmark for local multi-terminal sync. Its implementation is
proprietary but observable behaviour suggests an operation-log approach similar
to what AwoERP implements.

---

## 6. Proposed Tech Stack

> **👤 Non-Technical Reader**
> This section describes the programming languages, tools, and technologies
> used to build AwoERP POS. Think of it as listing the materials and tools
> a construction crew uses. You do not need to understand the tools to use
> the building — but if you are hiring developers or evaluating the project,
> this section is for you.

### 6.1 Go-First Stack (Primary Recommendation)

Go is recommended as the primary language for all backend and service components
for the following reasons:

- **Single binary deployment**: Go compiles to a self-contained executable.
  No runtime to install, no dependency conflicts, trivial to update.
- **Low resource footprint**: Critical for running on modest terminal hardware
  (fanless mini-PCs, tablets, Raspberry Pi 4+).
- **Excellent concurrency model**: Goroutines and channels are well-suited to
  the event-driven, concurrent nature of sync engines.
- **Strong standard library**: HTTP/2, TLS, SQLite bindings (via CGo or
  pure-Go driver), JSON handling — all production-ready in stdlib or first-party
  packages.
- **Cross-platform compilation**: A single Go build environment produces
  binaries for Windows, macOS, Linux, and ARM without modification.
- **Mature ecosystem for this problem domain**: Badger (embedded KV), bbolt,
  go-sqlite3, NATS (messaging), Fyne/Wails (UI) are all well-maintained.

**Recommended Go-first stack:**

| Layer | Technology | Rationale |
|---|---|---|
| **Backend service** | Go 1.22+ | Core language; sync engine, API, business logic |
| **Local database** | SQLite via `mattn/go-sqlite3` | Embedded, ACID, zero-admin |
| **LAN messaging** | NATS (embedded, `nats-server`) | Lightweight pub/sub; embeds in single binary |
| **Cloud sync API** | Go HTTP/2 + protobuf | Efficient binary serialisation for sync events |
| **Terminal UI** | Wails v2 (Go backend + web frontend) | Native window, Go business logic, web UI flexibility |
| **Frontend (UI layer)** | React + TypeScript (via Wails) | Component ecosystem; runs inside Wails WebView |
| **Styling** | Tailwind CSS | Utility-first; fast iteration; no runtime |
| **State management** | Zustand | Lightweight; sufficient for POS state complexity |
| **Receipt printing** | Go `escpos` library | Direct ESC/POS protocol to thermal printers |
| **Barcode scanning** | HID input capture (Go) | Scanners present as keyboards; captured at OS level |
| **Payment integration** | Vendor SDK (wrapped in Go) | Square/Stripe Terminal, Adyen etc. |
| **Cloud server** | Go + PostgreSQL | Central sync target; same language as terminal |
| **Infrastructure** | Docker + single docker-compose file | Easy self-hosted deployment |
| **Authentication** | JWT (terminal-to-cloud) + bcrypt PIN (local) | Minimal dependencies |
| **Service discovery** | `hashicorp/mdns` | mDNS in pure Go |

**Pros:**
- Entire stack is Go-centric; one language for all backend concerns.
- Minimal operational complexity; single binary per terminal.
- Excellent performance on resource-constrained hardware.
- Strong type safety catches integration errors at compile time.

**Cons:**
- Wails is maturing but not as battle-tested as Electron for complex UIs.
- Go UI ecosystem is less mature than mobile-native or web-native options.
- CGo (for SQLite) complicates cross-compilation slightly; use `modernc/sqlite`
  (pure Go) to eliminate this.

### 6.2 Alternative Stacks for Comparison

**Alternative A: Electron / Node.js Full-Stack**

| Layer | Technology |
|---|---|
| Backend service | Node.js (TypeScript) |
| Local database | SQLite via `better-sqlite3` |
| UI | Electron + React |
| LAN messaging | Socket.IO or NATS.js |
| Cloud server | Node.js + PostgreSQL |

Pros: Largest ecosystem; fastest UI development; easiest to hire for.
Cons: High memory footprint (Electron uses 200–500 MB baseline); slow startup;
V8 runtime required; single-binary deployment is complex.

**Alternative B: Flutter (Dart) + Go Backend**

| Layer | Technology |
|---|---|
| UI | Flutter (Dart) |
| Backend service | Go (unchanged from primary) |
| Local database | `sqflite` (Flutter SQLite) |
| Communication | gRPC (Flutter ↔ Go) |

Pros: Native-feeling UI on all platforms including iOS/Android (tablet POS);
hardware acceleration; excellent touch support.
Cons: Dart is a niche language; IPC between Flutter and Go adds complexity;
smaller talent pool than JS/React.

**Alternative C: Rust + Tauri**

| Layer | Technology |
|---|---|
| Backend service | Rust |
| UI | Tauri + React |
| Local database | SQLite via `rusqlite` |

Pros: Best-in-class performance and memory safety; smallest binary size.
Cons: Steep learning curve; slower development velocity; smaller ecosystem
for POS-specific libraries.

### 6.3 Technology Decision Matrix

| Criterion | Go + Wails | Electron/Node | Flutter + Go | Rust + Tauri |
|---|:---:|:---:|:---:|:---:|
| **Developer familiarity (Go primary)** | ✅ | ⚠️ | ⚠️ | ❌ |
| **UI richness** | ⚠️ | ✅ | ✅ | ⚠️ |
| **Resource footprint** | ✅ | ❌ | ✅ | ✅ |
| **Cross-platform** | ✅ | ✅ | ✅ | ✅ |
| **Single binary deploy** | ✅ | ❌ | ❌ | ✅ |
| **Offline capability** | ✅ | ✅ | ✅ | ✅ |
| **Ecosystem maturity** | ✅ | ✅ | ⚠️ | ⚠️ |
| **Hire-ability** | ⚠️ | ✅ | ⚠️ | ❌ |
| **Development velocity** | ✅ | ✅ | ✅ | ❌ |

**Recommendation:** Go + Wails for v1.0.0. If the team grows and Flutter
expertise is available, consider migrating the UI layer to Flutter while
keeping the Go backend unchanged. The architecture deliberately separates UI
from business logic to make this migration possible.

### 6.4 Key Libraries & Packages

```
go.mod dependencies (primary):
├── github.com/wailsapp/wails/v2            — Desktop application shell
├── modernc.org/sqlite                      — Pure-Go SQLite (no CGo)
├── github.com/nats-io/nats-server/v2       — Embedded NATS for LAN messaging
├── github.com/nats-io/nats.go              — NATS client
├── github.com/hashicorp/mdns              — mDNS service discovery
├── github.com/google/uuid                  — UUID generation
├── google.golang.org/protobuf             — Protobuf serialisation
├── github.com/golang-jwt/jwt/v5           — JWT authentication
├── golang.org/x/crypto                    — bcrypt for PIN hashing
├── github.com/mike42/escpos               — Thermal printer protocol
└── github.com/shopspring/decimal          — Precise decimal arithmetic (currency)
```

> **Important:** Never use float64 for currency calculations. The
> `shopspring/decimal` package provides arbitrary-precision decimal arithmetic,
> eliminating floating-point rounding errors that would cause reconciliation
> discrepancies.

### 6.5 Database Layer: Embedded vs. Server-Side

**Terminal (edge): SQLite**
- Embedded in the application process.
- Single file; trivially backed up.
- ACID compliant with WAL mode for concurrent readers.
- Suitable for workloads up to ~10,000 transactions/day per terminal.

**Local Coordinator: SQLite (small deployments) or PostgreSQL (large)**
- Small business (≤5 terminals, ≤500 transactions/day): SQLite is sufficient.
- Medium/large (>5 terminals or high volume): PostgreSQL with logical
  replication for cloud sync.

**Cloud server: PostgreSQL**
- Full-featured relational database.
- Supports the complex reporting queries that analytics requires.
- Logical replication used to stream changes to/from terminal sync agents.

```
Edge Terminal         Local Coordinator         Cloud Server
  (SQLite)    ←──→     (SQLite/PG)     ←──→    (PostgreSQL)
  WAL mode             Change capture           Logical replication
  ~10MB/day           ~50MB/day               Unlimited
```

---

## 7. System Architecture

### 7.1 High-Level Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                         BUSINESS PREMISES                           │
│                                                                      │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐                        │
│  │Terminal 1│   │Terminal 2│   │Terminal 3│                        │
│  │          │   │          │   │          │                        │
│  │  UI      │   │  UI      │   │  UI      │                        │
│  │  Engine  │   │  Engine  │   │  Engine  │                        │
│  │  SQLite  │   │  SQLite  │   │  SQLite  │                        │
│  │  NATS    │   │  NATS    │   │  NATS    │                        │
│  └────┬─────┘   └────┬─────┘   └────┬─────┘                        │
│       │              │              │                               │
│       └──────────────┴──────────────┘                               │
│                       LAN (mDNS + NATS)                            │
│                            │                                        │
│               ┌────────────┴──────────────┐                        │
│               │    Local Coordinator       │                        │
│               │                           │                        │
│               │  NATS Server (embedded)   │                        │
│               │  Sync Agent               │                        │
│               │  SQLite / PostgreSQL       │                        │
│               │  Hardware Manager         │                        │
│               └────────────┬──────────────┘                        │
│                             │                                        │
└─────────────────────────────│────────────────────────────────────────┘
                              │ WAN (HTTPS / gRPC-TLS)
                              │ (optional; offline-first)
                    ┌─────────┴──────────┐
                    │    Cloud Server     │
                    │                    │
                    │  Sync API (Go)     │
                    │  PostgreSQL        │
                    │  Reporting API     │
                    │  Admin Dashboard   │
                    └────────────────────┘
```

### 7.2 Terminal (Edge) Component

Each terminal runs a single Go binary containing:

```
TERMINAL BINARY
├── UI Shell (Wails WebView)
│   └── React frontend communicates via Wails bindings
│
├── POS Engine
│   ├── Order service       — Cart, order lifecycle
│   ├── Product service     — Catalogue queries, search
│   ├── Inventory service   — Local stock tracking
│   ├── Payment service     — Payment provider integration
│   ├── Discount engine     — Rules evaluation
│   ├── Tax engine          — Rate application
│   ├── Receipt service     — Print + digital queuing
│   └── Session service     — Shift management
│
├── Data Layer
│   ├── SQLite (modernc)    — Primary local store
│   └── Migration engine    — Schema versioning
│
├── Sync Client
│   ├── NATS subscriber     — LAN events inbound
│   ├── NATS publisher      — LAN events outbound
│   ├── Change queue        — Durable outbound queue
│   └── CRDT engine         — Merge logic
│
└── Hardware Abstraction
    ├── Printer driver      — ESC/POS
    ├── Scanner input       — HID capture
    └── Cash drawer         — Serial/USB trigger
```

### 7.3 Local Network Coordinator

The coordinator can run on a dedicated terminal, any terminal in "coordinator
mode", or a lightweight device such as a Raspberry Pi 4.

```
COORDINATOR BINARY
├── NATS Server (embedded)  — Message broker for all LAN terminals
├── Sync Aggregator         — Merges terminal change streams
├── Master Data Store       — SQLite/PostgreSQL for LAN-authoritative data
├── Cloud Sync Agent        — Batches and streams changes to cloud server
├── mDNS Registrar          — Advertises coordinator presence on LAN
├── Certificate Authority   — Issues TLS certs to new terminals
└── Admin API               — Local HTTP API for coordinator management
```

### 7.4 Central Sync Server

```
CLOUD SERVER (Docker Compose)
├── Sync API (Go)
│   ├── /sync/push          — Receive batches from coordinators
│   ├── /sync/pull          — Serve changes to coordinators
│   └── /sync/clock         — Vector clock exchange
│
├── Reporting API (Go)
│   ├── /reports/sales      — Aggregated sales data
│   ├── /reports/inventory  — Stock levels across locations
│   └── /reports/shifts     — Shift summaries
│
├── Admin API (Go)
│   ├── /admin/terminals    — Terminal registry
│   ├── /admin/users        — User management
│   └── /admin/config       — Business configuration push
│
├── PostgreSQL              — Primary data store
├── Redis (optional)        — API response caching for reporting
└── Nginx                   — TLS termination, rate limiting
```

### 7.5 Data Flow Diagrams

**Offline → Online sync:**

```
[Terminal offline, 3 transactions complete]

  T: ord-001 completed  →  Queue: [ord-001]
  T: ord-002 completed  →  Queue: [ord-001, ord-002]
  T: ord-003 completed  →  Queue: [ord-001, ord-002, ord-003]

[Connectivity restored]

  T → Coordinator: "I have events after clock {T:39}"
  Coordinator → T: "Send them"
  T → Coordinator: [ord-001, ord-002, ord-003] with vector clocks
  Coordinator: Merge using CRDTs, detect conflicts
  Coordinator → T: "Acknowledged. Your clock is now {T:42}"
  Coordinator → Cloud: [merged events batch]
  Cloud → Coordinator: "Acknowledged"
  T: Clear queue, update sync state
```

**Multi-terminal real-time:**

```
[T1 sells item; T2 concurrently sells same item]

  T1 → NATS topic "inventory.delta": {sku:"A", delta:-1, clock:{T1:10}}
  T2 → NATS topic "inventory.delta": {sku:"A", delta:-1, clock:{T2:7}}

  Coordinator receives both:
    Apply PN-Counter merge:
    Stock(A) = Initial - T1_subtractions - T2_subtractions
             = 10 - 1 - 1 = 8

  Coordinator → NATS broadcast: {sku:"A", stock:8}
  T1, T2, T3 update local stock to 8
```

### 7.6 Security Boundaries

```
┌──────────────────────────────────────────────────┐
│  TRUST ZONE 1: Terminal (physical access required) │
│                                                    │
│  Local PIN authentication (bcrypt)                │
│  Encrypted SQLite database                        │
│  No plain text secrets on disk                    │
└──────────────────────┬───────────────────────────┘
                       │ mTLS (mutual TLS)
┌──────────────────────┴───────────────────────────┐
│  TRUST ZONE 2: Local Network (LAN)                │
│                                                    │
│  TLS 1.3 for all NATS traffic                     │
│  Certificate-based terminal authentication        │
│  Coordinator is internal CA                       │
└──────────────────────┬───────────────────────────┘
                       │ HTTPS / gRPC-TLS
┌──────────────────────┴───────────────────────────┐
│  TRUST ZONE 3: Cloud (internet)                    │
│                                                    │
│  JWT bearer tokens (15-min expiry)                │
│  Refresh tokens (stored in OS keychain)           │
│  Rate limiting on all endpoints                   │
│  Data encrypted at rest (PostgreSQL TDE)          │
└──────────────────────────────────────────────────┘
```

---

## 8. Data Model

### 8.1 Core Entities

```sql
-- Products
CREATE TABLE products (
    id           TEXT PRIMARY KEY,  -- UUID
    sku          TEXT UNIQUE NOT NULL,
    name         TEXT NOT NULL,
    description  TEXT,
    category_id  TEXT,
    price        TEXT NOT NULL,     -- Decimal string; never float
    cost_price   TEXT,
    tax_class    TEXT NOT NULL DEFAULT 'standard',
    unit         TEXT NOT NULL DEFAULT 'item', -- item, kg, litre, etc.
    track_stock  BOOLEAN NOT NULL DEFAULT TRUE,
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL,
    -- CRDT metadata
    version      INTEGER NOT NULL DEFAULT 1,
    updated_by   TEXT NOT NULL      -- terminal_id of last update
);

-- Inventory
CREATE TABLE inventory (
    id              TEXT PRIMARY KEY,
    product_id      TEXT NOT NULL REFERENCES products(id),
    location_id     TEXT NOT NULL,  -- terminal or warehouse
    -- PN-Counter components
    units_added     INTEGER NOT NULL DEFAULT 0,
    units_subtracted INTEGER NOT NULL DEFAULT 0,
    -- Derived: current_stock = units_added - units_subtracted
    low_stock_threshold INTEGER,
    updated_at      TEXT NOT NULL,
    vector_clock    TEXT NOT NULL   -- JSON: {"TRM-A": 5, "TRM-B": 3}
);

-- Orders
CREATE TABLE orders (
    id              TEXT PRIMARY KEY,  -- UUID with terminal prefix
    order_number    TEXT NOT NULL,     -- Human-readable: ORD-20260511-0042
    terminal_id     TEXT NOT NULL,
    cashier_id      TEXT NOT NULL,
    status          TEXT NOT NULL,     -- open, held, completed, voided, refunded
    industry_data   TEXT,              -- JSON: table_id, room_number, etc.
    subtotal        TEXT NOT NULL,
    discount_total  TEXT NOT NULL DEFAULT '0',
    tax_total       TEXT NOT NULL,
    total           TEXT NOT NULL,
    notes           TEXT,
    created_at      TEXT NOT NULL,
    completed_at    TEXT,
    voided_at       TEXT,
    -- Sync metadata
    vector_clock    TEXT NOT NULL,
    synced_at       TEXT
);

-- Order Lines
CREATE TABLE order_lines (
    id              TEXT PRIMARY KEY,
    order_id        TEXT NOT NULL REFERENCES orders(id),
    product_id      TEXT NOT NULL,
    product_name    TEXT NOT NULL,  -- Snapshot; product may change later
    product_sku     TEXT NOT NULL,
    quantity        TEXT NOT NULL,  -- Decimal for weight-based
    unit_price      TEXT NOT NULL,  -- Price at time of sale
    discount        TEXT NOT NULL DEFAULT '0',
    tax_rate        TEXT NOT NULL,
    tax_amount      TEXT NOT NULL,
    line_total      TEXT NOT NULL,
    modifiers       TEXT,           -- JSON array: size, extras, etc.
    voided          BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      INTEGER NOT NULL DEFAULT 0
);

-- Payments
CREATE TABLE payments (
    id              TEXT PRIMARY KEY,  -- UUID with terminal prefix
    order_id        TEXT NOT NULL REFERENCES orders(id),
    method          TEXT NOT NULL,  -- cash, card, mobile_money, voucher
    amount          TEXT NOT NULL,
    tender          TEXT,           -- Amount given (cash)
    change_given    TEXT,           -- Change returned (cash)
    reference       TEXT,           -- Card auth code, mobile money ref
    status          TEXT NOT NULL,  -- pending, completed, failed, refunded
    provider_data   TEXT,           -- JSON: raw payment provider response
    created_at      TEXT NOT NULL,
    completed_at    TEXT
);

-- Users
CREATE TABLE users (
    id              TEXT PRIMARY KEY,
    username        TEXT UNIQUE NOT NULL,
    display_name    TEXT NOT NULL,
    role            TEXT NOT NULL,  -- cashier, supervisor, manager, admin
    pin_hash        TEXT NOT NULL,  -- bcrypt
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TEXT NOT NULL,
    last_login_at   TEXT
);

-- Sessions (Shifts)
CREATE TABLE sessions (
    id              TEXT PRIMARY KEY,
    terminal_id     TEXT NOT NULL,
    cashier_id      TEXT NOT NULL,
    opened_at       TEXT NOT NULL,
    closed_at       TEXT,
    opening_float   TEXT NOT NULL DEFAULT '0',
    closing_cash    TEXT,
    expected_cash   TEXT,
    variance        TEXT,
    status          TEXT NOT NULL   -- open, closed
);
```

### 8.2 Industry Extension Schema

Rather than altering core tables, industry data is stored in a `metadata` JSON
column (e.g., `industry_data` on orders) and in dedicated extension tables.

```sql
-- Restaurant extension
CREATE TABLE restaurant_tables (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,       -- "Table 4", "Bar Stool 2"
    capacity    INTEGER,
    section     TEXT,
    active      BOOLEAN DEFAULT TRUE
);

CREATE TABLE order_table_assignments (
    order_id    TEXT PRIMARY KEY REFERENCES orders(id),
    table_id    TEXT REFERENCES restaurant_tables(id),
    covers      INTEGER,             -- Number of diners
    assigned_at TEXT NOT NULL
);
```

### 8.3 Sync Metadata & Vector Clocks

```sql
-- Outbound sync queue (append-only)
CREATE TABLE sync_queue (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id        TEXT UNIQUE NOT NULL,
    terminal_id     TEXT NOT NULL,
    timestamp       TEXT NOT NULL,
    vector_clock    TEXT NOT NULL,   -- JSON
    entity_type     TEXT NOT NULL,
    entity_id       TEXT NOT NULL,
    operation       TEXT NOT NULL,   -- insert, update, delete
    payload         TEXT NOT NULL,   -- JSON
    checksum        TEXT NOT NULL,
    status          TEXT NOT NULL,   -- pending, syncing, synced, failed
    attempts        INTEGER NOT NULL DEFAULT 0,
    last_attempt    TEXT,
    error           TEXT
);

-- Inbound sync state
CREATE TABLE sync_state (
    peer_id         TEXT PRIMARY KEY,  -- terminal or cloud server ID
    last_clock      TEXT NOT NULL,     -- JSON vector clock
    last_sync_at    TEXT NOT NULL
);
```

### 8.4 Schema Versioning & Migrations

Migrations are embedded in the Go binary using `embed.FS`. They run automatically
on startup if the schema version is behind. Migrations are strictly append-only
— existing columns are never renamed or dropped in minor versions.

```
migrations/
├── 0001_initial_schema.sql
├── 0002_add_customer_table.sql
├── 0003_add_restaurant_extension.sql
└── 0004_add_loyalty_fields.sql
```

Migration tracking:

```sql
CREATE TABLE schema_migrations (
    version     INTEGER PRIMARY KEY,
    applied_at  TEXT NOT NULL,
    checksum    TEXT NOT NULL
);
```

---

## 9. Core Features — Deep Dive

### 9.1 Product & Inventory Management

`[v1.0.0]`

> **👤 Non-Technical Reader**
> This is your product catalogue — everything you sell, how much it costs,
> and how much of it you have. AwoERP lets you manage this from any terminal
> or from the cloud admin dashboard, and changes spread to all terminals
> automatically.

**Product management capabilities (v1.0.0):**
- Create, edit, deactivate products.
- Assign to categories for UI grouping.
- Set base price, cost price (for margin reporting).
- Assign tax class (VAT, exempt, zero-rated).
- Configure unit type (item, kg, litre, pair).
- Upload product image (displayed on product tile in UI).
- Set low-stock threshold for alerts.
- Barcode/SKU assignment.

**Inventory tracking:**
- Stock is deducted automatically on order completion.
- Stock is restored on void or refund.
- Manager can perform manual stock adjustments with reason code.
- Stock count report shows current levels, low-stock items, out-of-stock.

**What is deferred:**
- Purchase orders and goods receiving (`Future`).
- Supplier management (`Future`).
- Multi-location inventory transfers (`v1.x`).
- Composite products / Bill of Materials (`v1.x`).

### 9.2 Order Management & Cart Engine

`[v1.0.0]`

The cart engine handles all in-progress order state. Its design priorities are
speed (sub-10ms recalculation on any change) and correctness (totals always
match line items; tax always computed from configured rules, never from UI input).

**Cart operations:**
- Add product by barcode scan, SKU entry, or tile tap.
- Set quantity (integer or decimal for weight-based products).
- Override price (requires manager PIN above threshold).
- Apply line-item discount (percentage or fixed).
- Apply order-level discount.
- Add free-text note to line item or order.
- Remove line item.
- Hold order (assign identifier, visible to all terminals on LAN).
- Recall held order.
- Void order (requires reason code; manager PIN for completed orders).
- Split order (split lines across two or more payments).

**Calculation order:**

```
For each line:
  line_subtotal = unit_price × quantity
  line_discount = apply_discount_rules(line_subtotal)
  line_net      = line_subtotal - line_discount
  line_tax      = line_net × tax_rate(product.tax_class)
  line_total    = line_net + line_tax

Order totals:
  subtotal      = SUM(line_net)
  order_discount = apply_order_discount(subtotal)
  tax_total     = SUM(line_tax) adjusted for order_discount
  total         = subtotal - order_discount + tax_total
```

Tax is computed inclusive or exclusive of price depending on business
configuration (tax-inclusive pricing is common in many markets; tax-exclusive
is standard in others).

### 9.3 Payment Processing

`[v1.0.0 (cash + card); v1.x (mobile money, QR)`

**Supported methods in v1.0.0:**

*Cash:*
- Enter tender amount; system calculates change.
- Change display is large and prominent for cashier and customer.
- Cash drawer opens via ESC/POS command on payment completion.

*Card (integrated terminal):*
- Communicates with external payment terminal (Stripe Terminal, Square Reader,
  Adyen, or region-specific acquirer SDK).
- Payment terminal managed as a separate device; POS sends amount, terminal
  handles card interaction, result returned to POS.
- Offline card: Dependent on payment terminal capability. Some terminals
  support store-and-forward up to configured limits.

*Split payment:*
- Any combination of cash and card.
- Each tender recorded separately on the payment record.
- Correct change calculated based on cash portions only.

**Idempotency:**
Every payment attempt is assigned a unique idempotency key before the first
attempt. If the network drops during a card payment, the cashier can retry
safely — the payment provider will return the same result without double-charging.

### 9.4 Receipt Generation

`[v1.0.0 (print); v1.x (digital)`

**Thermal receipt (v1.0.0):**
- Communicates via ESC/POS protocol to any compatible thermal printer.
- Receipt template is configurable: business name, logo, footer message,
  tax breakdown, return policy.
- Duplicate receipt print available from order history.

**Digital receipt (v1.x):**
- Customer provides email or phone at checkout.
- Receipt queued as a background job.
- Sent via configured SMTP or SMS gateway on next connectivity.

**Receipt contents:**
```
[Business Logo]
AWOERP DEMO STORE
Westlands, Nairobi
VAT No: P051234567X

Receipt #: ORD-20260511-0042
Date: 11 May 2026  10:34 AM
Cashier: Amara K.
Terminal: Register 1

─────────────────────────────
Espresso              KES 250
Croissant x2          KES 300
─────────────────────────────
Subtotal              KES 550
VAT (16%)              KES 88
─────────────────────────────
TOTAL                 KES 638
─────────────────────────────
Cash (tendered)       KES 700
Change                 KES 62
─────────────────────────────
Thank you! Visit again.
WiFi: CafeGuest / pw: coffee
─────────────────────────────
[QR code: digital receipt URL]
```

### 9.5 Discounts, Promotions & Tax Engine

`[v1.0.0 (basic); Future (complex promotions)`

**v1.0.0 discount capabilities:**
- Percentage discount on line item.
- Fixed amount discount on line item.
- Percentage discount on order total.
- Fixed amount discount on order total.
- Cashier discount limit (e.g., max 10% without supervisor PIN).
- Manager override for above-threshold discounts.

**Tax engine:**
- Multiple tax classes configured per business.
- Products assigned to a tax class.
- Tax classes can be: standard rate, reduced rate, zero-rated, exempt.
- Inclusive and exclusive pricing modes.
- Tax number printed on receipt.
- Tax breakdown in daily report.

**Deferred (Future):**
- Time-based promotions (happy hour, lunch special).
- Bundle pricing (buy 2 get 1 free).
- Customer-specific pricing.
- Loyalty-based discounts.

### 9.6 Reporting & Analytics

`[v1.0.0 (basic); Future (advanced)`

**v1.0.0 reports (available offline from local data):**

*Daily Sales Summary (end-of-day):*
- Total sales, transactions, average transaction value.
- Sales by payment method.
- Tax collected.
- Void and refund totals.
- Top 10 products by revenue and units.

*Shift Report:*
- Opening and closing float.
- Cash sales total.
- Expected cash in drawer.
- Actual cash count.
- Variance with explanation field.

*Inventory Snapshot:*
- Current stock levels.
- Low-stock items (below threshold).
- Out-of-stock items.
- Stock movements during period.

**Future:**
- Sales trends (weekly, monthly, year-over-year).
- Hourly transaction heatmap.
- Cashier performance comparison.
- Customer purchase history (with CRM-lite).
- Exportable to CSV/Excel.

### 9.7 User & Role Management

`[v1.0.0]`

Users are managed from any Manager-level terminal or the cloud admin dashboard.
Changes propagate to all terminals via the standard sync mechanism, ensuring
a new user or a deactivated user is updated everywhere within seconds on a
connected network.

**PIN policy:**
- Minimum 4 digits; 6 digits recommended.
- PINs are hashed with bcrypt; never stored in plain text.
- PIN is required for every role switch.
- Manager can reset any user's PIN.
- Failed PIN attempt limit: 5 attempts → 30-second lockout (configurable).

**Audit trail:**
Every action that modifies data is recorded with the user ID, terminal ID,
and timestamp. The audit log is append-only and cannot be modified by any
in-system role. This provides accountability without surveillance.

---

## 10. Deployment & Operations

### 10.1 Hardware Requirements

> **👤 Non-Technical Reader**
> AwoERP POS is designed to run on affordable, widely available hardware.
> You do not need expensive proprietary equipment.

**Minimum terminal hardware:**

| Component | Minimum | Recommended |
|---|---|---|
| CPU | Intel Celeron / ARM Cortex-A72 | Intel Core i3 / AMD Ryzen 3 |
| RAM | 2 GB | 4 GB |
| Storage | 32 GB SSD | 64 GB SSD |
| Display | 10" 1280×800 touch | 15.6" 1920×1080 touch |
| OS | Ubuntu 22.04 LTS / Windows 10 | Ubuntu 22.04 LTS |
| Network | 100 Mbps Ethernet or Wi-Fi 5 | Ethernet preferred |

**Compatible peripheral hardware:**
- **Receipt printers:** Any ESC/POS compatible thermal printer (Epson TM series,
  Star Micronics, Bixolon, or generic Chinese equivalents).
- **Barcode scanners:** Any USB HID barcode scanner (plug-and-play, no drivers).
- **Cash drawer:** RJ11/RJ12 connection through receipt printer (standard).
- **Card payment terminal:** Stripe Terminal, Square Reader, or region-specific
  acquirer hardware (connected via USB or Bluetooth).
- **Customer display:** Optional second monitor or VFD display (v1.x).

**Budget hardware examples:**
- Mini PC: Beelink Mini S12 (~$150) — excellent for a permanent register.
- Tablet: Lenovo Tab P11 (~$200) — good for mobile or pop-up registers.
- Coordinator (small): Raspberry Pi 4 4GB (~$55) — lightweight, reliable.

### 10.2 Installation & Configuration Guide

**Step 1: Prepare the device**

```bash
# On Ubuntu 22.04 LTS
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl
```

**Step 2: Download and install AwoERP POS**

```bash
curl -fsSL https://releases.awoerp.io/pos/install.sh | bash
# The installer:
#   - Downloads the appropriate binary for your architecture
#   - Creates /opt/awoerp-pos/ directory
#   - Installs systemd service (auto-start on boot)
#   - Creates /var/awoerp-pos/ data directory
```

**Step 3: First-run setup**

```
AwoERP POS First Run Setup
──────────────────────────
1. Business Information
   Business name: _________________
   Industry profile: [Retail] [Restaurant] [Services] [Custom]
   Tax number (optional): _________________
   
2. Admin Account
   Admin PIN (6 digits): ______
   Confirm PIN: ______
   
3. Network Role
   [◉] Standalone / Primary terminal (no coordinator detected)
   [ ] Join existing network (coordinator detected at 192.168.1.x)
   
4. Hardware Setup
   Printer: [Auto-detect] [Manual setup] [Skip]
   Scanner: [Auto-detect — connected via USB]
   Card terminal: [None] [Stripe] [Square] [Adyen] [Other]
```

**Step 4: Adding products**

Products can be added through:
- The terminal UI (Manager → Products → Add Product).
- CSV import (Manager → Products → Import).
- Cloud admin dashboard (syncs to all terminals).

**Step 5: Opening the first shift**

```
Manager menu → Sessions → Open Shift
Enter opening cash float: KES 2,000
Confirm → Shift open. Ready to trade.
```

### 10.3 Network Setup Checklist

For multi-terminal deployments:

```
PRE-INSTALLATION
□ All terminals connected to same LAN segment (same router/switch)
□ mDNS/Bonjour not blocked by router (most home routers: fine; enterprise
  routers: may need IT to enable multicast forwarding)
□ Dedicated coordinator device prepared (or designate first terminal)
□ Static IP assigned to coordinator (recommended; prevents address changes)
□ Thermal printers connected and tested

COORDINATOR SETUP
□ Install AwoERP on coordinator device
□ Select "Coordinator" role during first-run setup
□ Note coordinator IP address displayed on screen
□ Complete business configuration on coordinator

TERMINAL SETUP (repeat per terminal)
□ Install AwoERP on terminal
□ On first-run, select "Join existing network"
□ Terminal auto-discovers coordinator via mDNS
□ Approve terminal on coordinator (admin PIN required)
□ Terminal downloads initial data (2–5 minutes)
□ Test a sample sale, confirm sync on coordinator

CLOUD SYNC (optional)
□ Create account at cloud.awoerp.io
□ Retrieve API key from dashboard
□ Enter API key on coordinator: Settings → Cloud Sync → Configure
□ Confirm sync indicator shows green
```

### 10.4 Upgrade Strategy

`[v1.0.0]`

Upgrades are zero-downtime and occur during a business's natural pause
(end-of-day close or shift change). The process is:

1. Coordinator downloads new version in the background.
2. At end of shift (or manager-initiated), coordinator notifies terminals:
   "Update available. Apply now or defer to [time]."
3. Each terminal, on acknowledgement, downloads and applies the update.
4. SQLite migrations run automatically on first start of new version.
5. If a migration fails, the terminal rolls back to the previous version
   and reports the error.

Coordinator and terminal versions are independently upgradeable, with
backward-compatibility guaranteed for one major version behind.

### 10.5 Backup & Disaster Recovery

**Local backup:**
- SQLite database file is a single file: `/var/awoerp-pos/data/pos.db`.
- Automatic daily snapshot with 30-day rolling retention.
- Manager can trigger manual backup from menu.
- Backup can be saved to USB, network share, or cloud storage.

**Coordinator backup:**
- Coordinator maintains a complete replica of all terminal data.
- Coordinator database is the primary backup for all terminals.

**Disaster recovery scenarios:**

| Scenario | Recovery |
|---|---|
| Terminal disk failure | Restore from coordinator. New terminal joins and syncs within 5 minutes. |
| Coordinator failure | Promote any terminal to coordinator. Others join new coordinator. Cloud sync catches up. |
| All local hardware destroyed | Restore from cloud backup. New hardware installs, syncs from cloud. Historical data intact. |
| Accidental order deletion | Restore from previous day's snapshot (orders are append-only in normal operation; deletion is a soft-delete flag only). |

---

## 11. Developer Guide

### 11.1 Repository Structure

```
awoerp-pos/
├── cmd/
│   ├── terminal/           — Terminal binary entry point
│   └── coordinator/        — Coordinator binary entry point
│
├── internal/
│   ├── engine/             — Core POS business logic
│   │   ├── order/
│   │   ├── product/
│   │   ├── inventory/
│   │   ├── payment/
│   │   ├── discount/
│   │   ├── tax/
│   │   └── session/
│   │
│   ├── sync/               — CRDT & sync engine
│   │   ├── crdt/           — CRDT data structures
│   │   ├── queue/          — Durable change queue
│   │   ├── client/         — Sync client (to coordinator)
│   │   └── server/         — Sync server (coordinator role)
│   │
│   ├── store/              — SQLite data access layer
│   │   ├── migrations/
│   │   └── queries/        — sqlc-generated query code
│   │
│   ├── hardware/           — Printer, scanner, cash drawer
│   ├── network/            — mDNS, NATS, TLS
│   └── config/             — Configuration loading & validation
│
├── frontend/               — React/TypeScript UI
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── store/          — Zustand state
│   │   └── wailsjs/        — Auto-generated Wails bindings
│   └── package.json
│
├── cloud/                  — Cloud server (separate deployable)
│   ├── cmd/server/
│   ├── internal/
│   │   ├── sync/
│   │   ├── reporting/
│   │   └── admin/
│   └── migrations/
│
├── scripts/                — Build, release, dev utilities
├── docs/                   — This documentation
├── Makefile
├── go.mod
└── docker-compose.yml      — Cloud server deployment
```

### 11.2 Local Development Setup

**Prerequisites:**
- Go 1.22+
- Node.js 20+
- Wails CLI v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- (Linux) GTK3 dev headers: `sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev`

```bash
# Clone
git clone https://github.com/awoerp/pos.git && cd pos

# Install frontend dependencies
cd frontend && npm install && cd ..

# Run terminal in dev mode (hot reload for frontend, Go restart on backend change)
wails dev

# Run cloud server
cd cloud && go run ./cmd/server

# Run all tests
make test

# Build production binary
make build        # Current platform
make build-all    # All platforms (Windows, Linux, Linux ARM)
```

**Recommended dev environment:**
- Run 2–3 terminal instances (different ports) to test multi-terminal sync locally.
- Use the `scripts/seed_demo.sh` script to populate demo products and orders.

```bash
# Start 3 terminals + 1 coordinator in dev mode
make dev-multi
```

### 11.3 API Reference Overview

The terminal exposes a local HTTP API (used internally by the Wails frontend
and accessible to integrations on the same machine):

```
Base URL: http://localhost:8742/api/v1

Authentication: X-Terminal-Token header (set during terminal init)

PRODUCTS
GET    /products              — List products (with search, category filter)
GET    /products/:id          — Get single product
POST   /products              — Create product [Manager]
PUT    /products/:id          — Update product [Manager]
DELETE /products/:id          — Deactivate product [Manager]

ORDERS
GET    /orders                — List orders (paginated)
GET    /orders/:id            — Get order with lines and payments
POST   /orders                — Create order
PUT    /orders/:id            — Update order (add lines, apply discount)
POST   /orders/:id/complete   — Complete order
POST   /orders/:id/hold       — Hold order
POST   /orders/:id/void       — Void order [Supervisor]
POST   /orders/:id/recall     — Recall held order

PAYMENTS
POST   /payments              — Record payment against order

INVENTORY
GET    /inventory             — Stock levels (local terminal)
POST   /inventory/adjust      — Manual adjustment [Manager]

SESSIONS
POST   /sessions/open         — Open shift
POST   /sessions/close        — Close shift
GET    /sessions/current      — Current shift summary

REPORTS
GET    /reports/daily         — Daily sales summary
GET    /reports/shift/:id     — Shift report
GET    /reports/inventory     — Stock snapshot
```

The cloud server exposes an additional reporting and admin API on port 8743
(proxied via Nginx on production).

### 11.4 Plugin / Extension System

`[Future]`

The v1.0.0 system does not expose a public plugin API. Extension points are
planned for v2.0.0:

- **Industry Profile plugins**: New profiles installable without rebuilding.
- **Payment provider plugins**: Standard interface for new payment methods.
- **Report plugins**: Custom report templates.
- **Webhook events**: POST to configured URLs on order complete, stock low, etc.

The internal service interfaces in the `engine/` package are designed with
these extension points in mind — services communicate through defined interfaces,
making future plugin injection straightforward.

### 11.5 Testing Strategy

```
TEST PYRAMID
              /\
             /  \
            / E2E \        — Wails UI tests (Playwright)
           /──────\
          / Integration \  — HTTP API + SQLite + NATS
         /──────────────\
        /   Unit Tests   \ — Engine logic, CRDT merge, tax calc
       /──────────────────\
```

**Critical test coverage targets:**

| Area | Coverage target | Why |
|---|---|---|
| CRDT merge logic | 100% | Mathematical correctness is essential |
| Tax & discount calculation | 100% | Financial accuracy is non-negotiable |
| Payment idempotency | 100% | Double-charge prevention |
| Sync queue durability | 95%+ | Data loss prevention |
| Order lifecycle | 90%+ | Core business flow |
| UI flows | 70%+ | Regression prevention |

**Running tests:**

```bash
make test               # All unit + integration tests
make test-coverage      # With HTML coverage report
make test-e2e           # End-to-end (requires Playwright)
make test-crdt          # CRDT-specific stress tests (concurrent operations)
```

---

## 12. Appendices

### Appendix A: Glossary

| Term | Definition |
|---|---|
| **ACID** | Atomicity, Consistency, Isolation, Durability — database transaction properties that guarantee data integrity |
| **CRDT** | Conflict-free Replicated Data Type — a data structure that can be merged across replicas without conflicts |
| **ESC/POS** | A printer control language used by most thermal receipt printers |
| **HID** | Human Interface Device — the USB device class that covers keyboards, mice, and barcode scanners |
| **idempotency** | A property of an operation where performing it multiple times has the same effect as performing it once |
| **LWW** | Last-Write-Wins — a conflict resolution strategy that keeps the most recently timestamped value |
| **mDNS** | Multicast DNS — a zero-configuration protocol for discovering devices on a local network |
| **NATS** | A lightweight messaging system used for communication between terminals |
| **PN-Counter** | Positive-Negative Counter — a CRDT that supports both increment and decrement |
| **protobuf** | Protocol Buffers — a compact binary serialisation format developed by Google |
| **SKU** | Stock Keeping Unit — a unique identifier for a product |
| **SQLite** | A lightweight, embedded relational database stored as a single file |
| **TLS** | Transport Layer Security — the encryption protocol used to secure network communication |
| **Vector clock** | A mechanism for tracking the causal ordering of events across multiple systems |
| **WAL** | Write-Ahead Log — a SQLite mode that allows concurrent reads during writes |

### Appendix B: Full Competitive Feature Comparison

| Feature | Square | Toast | Lightspeed | Shopify POS | AwoERP v1 | AwoERP future |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Offline sales (cash) | ⚠️ | ✅ | ✅ | ❌ | ✅ | ✅ |
| Offline sales (card) | ⚠️ | ✅ | ⚠️ | ❌ | ⚠️ | ✅ |
| Local multi-terminal sync | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ |
| CRDT conflict resolution | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| Self-hostable | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| Industry-agnostic | ✅ | ❌ | ⚠️ | ⚠️ | ✅ | ✅ |
| Restaurant profile | ⚠️ | ✅ | ✅ | ❌ | v1.x | ✅ |
| Retail profile | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ |
| Table management | ⚠️ | ✅ | ✅ | ❌ | v1.x | ✅ |
| KDS (Kitchen Display) | ❌ | ✅ | ✅ | ❌ | v1.x | ✅ |
| Weight-based pricing | ❌ | ❌ | ✅ | ❌ | v1.x | ✅ |
| Open data export | ⚠️ | ⚠️ | ✅ | ⚠️ | ✅ | ✅ |
| Barcode scanning | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Split payments | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Loyalty programme | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| E-commerce integration | ✅ | ❌ | ✅ | ✅ | ❌ | v2.x |
| Free / open core | ✅ | ❌ | ❌ | ❌ | TBD | TBD |
| Low-cost hardware | ✅ | ❌ | ❌ | ⚠️ | ✅ | ✅ |
| Monthly cost (entry) | $0 + % | $110+ | $69+ | $89+ | TBD | TBD |

*✅ = Yes  ⚠️ = Partial  ❌ = No*

### Appendix C: Recommended Hardware

| Product | Category | Price range | Notes |
|---|---|---|---|
| Beelink Mini S12 Pro | Mini PC / terminal | ~$150–180 | Intel N100, 16GB RAM option, fanless |
| Lenovo Tab P11 (2nd gen) | Tablet terminal | ~$180–220 | 11", good touch, Android — needs Wails mobile port |
| Raspberry Pi 4 (4GB) | Coordinator | ~$55–75 | Ideal for coordinator-only role |
| Epson TM-T88VI | Receipt printer | ~$200–300 | Industry standard, excellent reliability |
| Bixolon SRP-350V | Receipt printer | ~$120–160 | Good budget option |
| Honeywell Voyager 1200g | Barcode scanner | ~$80–120 | Durable, plug-and-play |
| APG Vasario 1616 | Cash drawer | ~$80–120 | Standard RJ11, works with any Epson printer |
| Stripe Terminal BBPOS WisePOS E | Card terminal | ~$249 | Integrated SDK, good offline support |

### Appendix D: CRDT Reference

**G-Counter merge operation:**

```go
// G-Counter: each terminal maintains its own counter segment
type GCounter struct {
    counts map[string]int64 // terminal_id → count
}

func (c *GCounter) Increment(terminalID string, by int64) {
    c.counts[terminalID] += by
}

func (c *GCounter) Value() int64 {
    var total int64
    for _, v := range c.counts {
        total += v
    }
    return total
}

// Merge: take max of each terminal's segment
func (c *GCounter) Merge(other *GCounter) {
    for id, val := range other.counts {
        if existing, ok := c.counts[id]; !ok || val > existing {
            c.counts[id] = val
        }
    }
}
```

**PN-Counter (inventory tracking):**

```go
type PNCounter struct {
    additions    GCounter // stock received
    subtractions GCounter // stock sold
}

func (c *PNCounter) Value() int64 {
    return c.additions.Value() - c.subtractions.Value()
}

func (c *PNCounter) Add(terminalID string, amount int64) {
    c.additions.Increment(terminalID, amount)
}

func (c *PNCounter) Subtract(terminalID string, amount int64) {
    c.subtractions.Increment(terminalID, amount)
}

func (c *PNCounter) Merge(other *PNCounter) {
    c.additions.Merge(&other.additions)
    c.subtractions.Merge(&other.subtractions)
}
```

**Vector clock comparison:**

```go
type VectorClock map[string]int64

// IsConcurrentWith returns true if neither clock dominates the other
// (i.e., the events are truly concurrent and need CRDT merge)
func (vc VectorClock) IsConcurrentWith(other VectorClock) bool {
    aAhead, bAhead := false, false
    for id, val := range vc {
        if val > other[id] { aAhead = true }
    }
    for id, val := range other {
        if val > vc[id] { bAhead = true }
    }
    return aAhead && bAhead
}
```

### Appendix E: Changelog & Roadmap

**v1.0.0 (target: Q3 2026)**
- Core POS engine (products, orders, payments, inventory)
- Offline-first operation with SQLite
- LAN multi-terminal sync via NATS + CRDT
- Cloud sync to central PostgreSQL
- Cashier, Supervisor, Manager, Admin roles
- Thermal printer support (ESC/POS)
- Barcode scanner support (HID)
- Cash and card payment
- Split payments
- Daily and shift reports
- Basic discount engine
- Tax engine (inclusive/exclusive)
- Terminal discovery via mDNS

**v1.1.0**
- Restaurant industry profile (table management, basic KDS)
- Customer-facing display support
- Digital receipts (email/SMS)
- Multi-currency
- Multi-language UI (i18n framework)

**v1.2.0**
- Mobile money / QR payment integration
- Weight-based pricing profile
- Customer profiles (CRM-lite)
- CSV/Excel data export

**v2.0.0**
- Plugin / extension API
- E-commerce integration (WooCommerce, Shopify as order source)
- Advanced analytics dashboard
- Loyalty programme
- Peer-to-peer (no coordinator) multi-terminal topology
- Mobile app (iOS/Android) cashier client

---

*AwoERP POS Terminal Documentation — v1.0.0-draft*  
*© 2026 AwoERP. All rights reserved.*  
*For corrections or contributions, open an issue at github.com/awoerp/pos*
