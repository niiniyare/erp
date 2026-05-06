# AWO ERP — Release & Feature Delivery Handbook

> **Version:** 2.0  
> **Applies To:** AWO ERP (`awo.so`) — all tenants and environments  
> **Audiences:** Product Management · Engineering · Tenant Admins & Staff  
> **Last Updated:** 2025

---

## How to Read This Document

This handbook follows a single linear thread — from the moment a change is conceived to the moment a tenant in Nairobi opens their dashboard and sees something different. You do not need to read it all in one sitting, but the sections build on each other. Each section explains *why* before it explains *what* and *how*.

Three symbols mark content by audience. They are guides, not walls — every section is worth understanding at least at a high level regardless of your role.

| Symbol | Audience | Examples |
|--------|----------|---------|
| 📋 | **Product Management** — decisions, priorities, communication | Feature triage, release scope, tenant comms |
| 🔧 | **Engineering** — procedures, code, infrastructure | Migrations, flags, deployment steps |
| 👤 | **Tenant Admins & Staff** — what this means for your daily use | Version names, what changed, what to expect |

---

<br>

# Part I — Foundations

---

## 1. Why This Document Exists

AWO ERP is a multi-tenant SaaS platform. Every business decision about *how* a feature is built, *when* it ships, and *who* sees it first is shaped by that one fact. A bad release does not affect one user — it can affect every fuel station, hotel, and retail shop running on the platform simultaneously.

This document defines the full lifecycle of a change in AWO ERP: from the first conversation about an idea, through engineering, testing, and deployment, to the moment a tenant experiences it in production. It covers how we version software, how we make decisions about what ships and when, how we control who sees what, and how we communicate throughout.

What this document does **not** cover: day-to-day sprint planning, module-level business logic, or the internal architecture of individual features. It is about the system of delivery, not the content of any specific delivery.

---

## 2. Three Principles That Drive Everything 📋

Every process in this document exists to honour one or more of these three principles. When you encounter a rule that seems bureaucratic, trace it back here — the reasoning will become clear.

---

> ### Stability Over Speed
>
> A release that is delayed by a week costs nothing permanent. A release that corrupts financial data for a live tenant — or breaks their end-of-month reconciliation the night before a KRA submission — costs trust that takes months to rebuild. Every process here exists to widen the gap between "code is written" and "code is live" enough to catch problems before they reach production.

---

> ### Tenant Trust Is the Product
>
> Tenants are not running hobby projects on AWO ERP. They are running real businesses — fuel stations reconciling daily wetstock, hotels closing nightly accounts, LPG distributors tracking cylinder inventory. Surprise changes, unexplained downtime, or numbers that look different after an update erode their trust faster than any missing feature builds it. Every release process includes a communication step not as an afterthought, but as a first-class deliverable equal in importance to the code itself.

---

> ### Decouple Deploy From Release
>
> Shipping code to the server and making that code visible to tenants are two separate events, and keeping them separate is the single most important practice in safe delivery. When deploy and release are the same moment, a problem is already live by the time it is detected. When they are separated — by feature flags, by controlled rollouts, by dark launching — problems can be caught and reversed before any tenant is affected.

---

## 3. Glossary

Before anything else, these are the terms used throughout this document. Read through them once — the distinctions matter and will save confusion later.

| Term | Definition |
|------|------------|
| **Release** | A versioned set of changes made available to tenants |
| **Deploy** | Putting new code on the server — may or may not activate new features for tenants |
| **Feature Flag** | A named switch that controls whether a feature is visible to tenants, independent of deployment |
| **Tenant** | A business organisation using AWO ERP on its own isolated subdomain (e.g., `shellmaanzoni.awo.so`) |
| **Rollout** | The controlled process of making a change available to some or all tenants |
| **Migration** | A database schema change — `.up.sql` to apply, `.down.sql` to reverse |
| **Hotfix** | An emergency patch applied outside the normal release cycle for Critical or High bugs |
| **SemVer** | Semantic Versioning — a version number in the format `MAJOR.MINOR.PATCH` |
| **Dark Launch** | Deploying code that is switched off by flags — the feature exists on the server but tenants cannot see it |
| **RLS** | Row-Level Security — the PostgreSQL mechanism that ensures each tenant's data is completely isolated |
| **Codegen** | Auto-generated code (SQLC queries, Wire DI) that must never be hand-edited |
| **Flag Debt** | Accumulated feature flags that were never retired, creating dead code branches that complicate the codebase |
| **Feature Freeze** | A point in the release timeline after which no new features may be added to the release scope |
| **Code Freeze** | A point after which no code changes at all are permitted, except confirmed Critical bug fixes |

---

<br>

# Part II — The Two Version Systems

---

## 4. Why AWO ERP Maintains Two Version Numbers 📋 👤

AWO ERP tracks two completely different version identifiers in parallel. This is not redundancy — they serve entirely different readers and answer entirely different questions.

The **technical version** speaks to engineers and automated systems. It encodes the precise nature of a change: whether upgrading is safe without migration steps, whether APIs have changed in a way that breaks existing integrations, whether Temporal workflows in-flight will be affected. It must be unambiguous, machine-parseable, and exact.

The **end-user version** speaks to tenant admins and business users. A fuel station manager does not care that SQLC queries were regenerated or that a repository interface was refactored. They care that the monthly reconciliation report is now available, or that "AWO ERP 2025 Q3 — Tija" arrived with improved financial reporting. A version name should communicate value and set expectations — not require a decoder ring.

Choosing only one creates problems in the opposite direction. A purely technical version alienates tenants. A purely marketing version leaves engineers unable to safely manage dependencies, migrations, and integrations. Both must exist.

---

## 5. Technical Versioning: SemVer 🔧

AWO ERP uses **Semantic Versioning 2.0.0** — a three-part number with a specific, agreed meaning for each part.

```
MAJOR . MINOR . PATCH
  2   .   3   .   1
```

### MAJOR — Something Breaks

Increment MAJOR when an existing integration, API consumer, or correctly-configured tenant system will behave incorrectly after the update without anyone taking action. The emphasis is on *without anyone taking action* — if the change is backward-compatible and things keep working, it is not a MAJOR change.

**AWO ERP examples that earn a MAJOR bump:**
- Removing or renaming an API endpoint that tenant integrations call
- Changing the structure of a request or response body in a way that breaks callers
- Dropping a database column that external reporting tools query
- Changing the authentication token format, expiry model, or JWT structure
- Restructuring the RLS policy model in a way that requires tenant data re-scoping
- Changing the AWO ERP configuration key names (Viper keys, env var names)
- Altering a Temporal workflow interface in a way that invalidates in-flight workflow executions

When MAJOR increments, MINOR and PATCH both reset to zero. `1.9.4 → 2.0.0`.

### MINOR — Something New, Nothing Broken

Increment MINOR when new functionality is added in a backward-compatible way. Everything that worked before still works. Tenants and integrations do not need to change anything — they simply gain new capabilities.

**AWO ERP examples:**
- Adding a new API endpoint (existing endpoints are untouched)
- Adding optional fields to an existing API response
- Introducing a new module — the Selling module, for example — alongside existing ones
- Adding a new feature flag category
- Improving query performance without changing the response

When MINOR increments, PATCH resets to zero. `1.4.2 → 1.5.0`.

### PATCH — Something Fixed, Nothing Added

Increment PATCH when a confirmed defect is corrected and nothing new is introduced. The system had an intention; a PATCH brings it back in line with that intention.

**AWO ERP examples:**
- A pump attendant's cash shortfall calculating incorrectly — fixed
- Session token refresh failing for long sessions — fixed
- VAT rounding error on invoices above KES 100,000 — fixed
- A filter on the AR aging report excluding accounts it should include — fixed

`1.5.0 → 1.5.1`

### Pre-Release Labels

Before a version reaches general availability, it passes through stages:

| Label | Meaning | Who Has Access |
|-------|---------|----------------|
| `2.0.0-alpha.1` | Early internal build — incomplete, may have breaking changes | Engineering only |
| `2.0.0-beta.1` | Feature-complete, under active testing | Engineering + selected tenants |
| `2.0.0-rc.1` | Release Candidate — believed production-ready, final verification only | Engineering + pilot tenants |
| `2.0.0` | General Availability | All tenants |

### Where the Technical Version Lives in AWO ERP

The version is not stored in one place — it is injected at build time and flows through the entire system:

- **Build:** Injected via `ldflags` when compiling `cmd/`
- **API headers:** Every response carries `X-AWO-Version: 1.5.2`
- **Health check:** `GET /health` returns `{"version":"1.5.2","status":"ok",...}`
- **Structured logs:** Every Zerolog line includes `"version":"1.5.2"` — this makes it possible to compare error rates between versions in a single log query
- **Git tags:** `git tag v1.5.2` — the canonical source of truth for what is deployed

---

## 6. End-User Versioning: Named Calendar Releases 👤 📋

The end-user version is what appears in the tenant dashboard, in release note emails, and in support conversations. AWO ERP uses a **calendar-based name combined with a Swahili release name**:

```
AWO ERP  YYYY . QN  —  [Release Name]
              ↑              ↑
         Quarter       Single Swahili word
                       reflecting the theme
```

Release names are single Swahili words chosen to reflect the theme and character of each release. They are memorable, pronounceable, and carry meaning without requiring tenants to decode numbers. A tenant saying "we upgraded to *Ukuaji* last month" is more natural and less error-prone in a support conversation than "we're on version 1.3.0."

| End-User Name | Technical Version | Theme |
|---------------|------------------|-------|
| AWO ERP 2025 Q1 — **Msingi** | `1.0.0` | Foundation — the first GA release |
| AWO ERP 2025 Q2 — **Ukuaji** | `1.3.0` | Growth — expanded module coverage |
| AWO ERP 2025 Q3 — **Tija** | `1.6.0` | Productivity — workflow improvements |
| AWO ERP 2025 Q4 — **Nguvu** | `2.0.0` | Power — a breaking-change major release |

### Rules for End-User Version Names

1. **One name per quarter** — regardless of how many technical patch releases occur within it
2. **PATCH releases within a quarter do not earn a new name** — they are called "*Ukuaji* — Update 1", "*Ukuaji* — Update 2", and so on
3. **A new name is only issued for a MINOR or MAJOR technical release**
4. **MAJOR versions receive their own announcement and are communicated weeks in advance**

### The Complete Translation Table

| Technical Version | End-User Version | Type | Notes |
|------------------|-----------------|------|-------|
| `1.0.0` | AWO ERP 2025 Q1 — Msingi | MAJOR (initial GA) | First release |
| `1.0.1` | Msingi — Update 1 | PATCH | Bug fix, no new name |
| `1.0.2` | Msingi — Update 2 | PATCH | Bug fix |
| `1.1.0` | Msingi — Feature Update | MINOR | New features within same quarter |
| `1.3.0` | AWO ERP 2025 Q2 — Ukuaji | MINOR (new quarter) | New named release |
| `2.0.0` | AWO ERP 2026 — Nguvu | MAJOR | Breaking changes, major announcement |

### Where Each Version Appears

| Location | Technical Version | End-User Version |
|----------|-----------------|-----------------|
| API response headers | ✅ `X-AWO-Version` | ❌ |
| Health check endpoint | ✅ | ❌ |
| Application logs | ✅ | ❌ |
| Tenant dashboard footer | ❌ | ✅ |
| Release notes email | Brief mention | ✅ Primary |
| Support ticket system | ✅ (engineer reference) | ✅ (tenant reference) |
| Git repository tags | ✅ | ❌ |
| Changelog (internal) | ✅ | ✅ (translated section) |

### Keeping Both in Sync: RELEASES.md

The single source of truth for the mapping between technical and end-user versions is a `RELEASES.md` file at the root of the AWO ERP repository. Every time a new technical tag is created, the release manager adds a row to this file with the corresponding end-user label and a one-line summary. Both the engineering changelog and the tenant-facing release notes are generated from this file — ensuring they never fall out of step with each other.

---

<br>

# Part III — What Goes Into a Release

---

## 7. Three Inputs That Shape Every Release 📋

No release is assembled arbitrarily. Every change that ships comes from one of three sources, and each source is handled differently.

**Roadmap items** are changes that have been deliberately planned as part of AWO ERP's product strategy. They represent the direction the platform is growing — new modules, expanded capabilities, architectural improvements. For AWO ERP right now, this includes completing the Selling module, integrating Temporal workflows for asynchronous processing, and building the tenant self-service onboarding flow. Roadmap items are the backbone of MINOR and MAJOR releases.

**Bug reports** arrive through support tickets, internal monitoring alerts, and direct tenant communication. A bug is not a request for something new — it is a report that something that should work does not. Bugs do not wait for quarterly roadmap cycles. A critical calculation error in fuel sales reporting cannot sit in a backlog for six weeks.

**Tenant feedback** is the most ambiguous of the three. When a tenant admin says "it would be really helpful if…", that is not automatically a bug or a roadmap item. It first goes through triage to determine whether it identifies a defect (then it is a bug), aligns with the product direction (then it enters the roadmap), or is outside scope (then it is declined with a clear explanation). Feedback without triage becomes a pile of unmaintainable promises.

---

## 8. Feature Triage: From Idea to Decision 📋

When a feature request enters the pipeline — from the roadmap or from tenant feedback — it is evaluated against four questions in sequence. This is not a committee — it is a structured thought process that the Product Manager works through, with input from the Engineering Lead where cost estimation is required.

**1. Does it serve AWO ERP's core use case?**
AWO ERP is an ERP built for businesses operating in the Kenyan market — fuel stations, hotels, retail, LPG supply chains. A feature that directly serves these operations earns higher priority than a generic ERP feature that could belong to any platform. An automated fuel wetstock reconciliation tool is core; a generic Gantt chart project manager is not.

**2. How many tenants need it?**
A feature that five tenants have independently requested is different from a feature that one tenant asked for once. The more widely a need is felt, the more it justifies engineering investment. Track requests — patterns matter.

**3. What is the implementation cost?**
A new filter on an existing report costs one or two days. A full LPG cylinder tracking module costs weeks. Cost is not a reason to reject a feature, but it is a reason to schedule it correctly and not squeeze it into a release where it will be rushed.

**4. Does it introduce risk?**
Features that touch authentication, financial calculations, or RLS policies require significantly more testing and a more conservative rollout plan than features that add a new report filter. High-risk features must not be fast-tracked.

### Triage Outcomes

| Outcome | Meaning | Next Step |
|---------|---------|-----------|
| **Must-have** | Blocks a tenant workflow or a commercial deal | Prioritise immediately; enter current sprint if possible |
| **Nice-to-have** | Adds genuine value but nothing breaks without it | Schedule for next MINOR release |
| **Deferred** | Good idea, wrong time — capacity, dependencies, or timing | Add to backlog with a documented reason and a revisit date |
| **Rejected** | Outside AWO ERP's scope or product direction | Close with a clear written explanation to the requester |

Every rejected request deserves an explanation. "This is outside our current scope because AWO ERP focuses on…" is respectful. Silence or vague dismissal is not.

---

## 9. Bug Triage: Severity Classification 🔧 📋

Not all bugs are equal, and treating them as if they are — queuing a data loss bug alongside a cosmetic misalignment — is how critical problems sit unfixed for weeks. AWO ERP uses four severity levels with specific, unambiguous definitions.

| Severity | Definition | Mandatory Response | Release Path |
|----------|-----------|-------------------|-------------|
| **Critical** | Data loss, security breach, RLS isolation failure, complete feature unavailability for all tenants | Immediate — drop everything | Emergency hotfix within hours |
| **High** | Major feature broken for multiple tenants, no viable workaround | Within 24 hours | Hotfix or expedited patch release |
| **Medium** | Feature degraded but a workaround exists; affects some tenants | Within the current sprint | Next scheduled patch release |
| **Low** | Minor visual issue, edge case with a clear workaround, cosmetic defect | Best effort | Next MINOR release |

> ⚠️ **Security vulnerabilities are always Critical**, regardless of how they present. A security bug that appears minor in isolation can be the entry point for something far more damaging.

The Engineering Lead owns severity classification. When a bug arrives, the first question is always: *can data be lost, can the wrong tenant see the wrong data, or can all tenants not access a critical feature?* If the answer is yes to any of these, it is Critical.

---

## 10. Who Decides, When, and With What 📋

Clarity about decision ownership prevents the two most common release problems: decisions made too late (because nobody knew they were responsible) and decisions made by the wrong person (because authority was unclear).

| Decision | Owner | When It Happens | What They Need to Decide |
|----------|-------|----------------|--------------------------|
| Feature enters the roadmap | Product Manager | Quarterly planning | Tenant demand data, strategic fit, rough cost estimate |
| Feature moves to active development | PM + Engineering Lead | Sprint planning | Available capacity, open dependencies, risk assessment |
| Bug severity classification | Engineering Lead | On receipt of report | Reproduction steps, affected tenants, scope of impact |
| Hotfix vs. queue for next release | Engineering Lead + PM | On Critical or High bug | Severity, tenant impact, complexity of fix |
| Release scope is locked | Product Manager | Feature freeze date | Status of all in-progress items, test coverage level |
| Go / No-go for production deploy | Engineering Lead | After staging validation | Test results, migration dry-run outcome, monitoring baseline |
| Release rollback | Engineering Lead | Post-release incident | Error rate data, P0 bug reports, tenant impact scope |

### The Release Scope Document

Before any release enters active development, the Product Manager creates a **Release Scope Document**. It is a short, living document — not a specification — that answers:

- What is the technical version number?
- What is the end-user version name?
- What features are included?
- What bugs are included?
- What is explicitly excluded, and why?
- What is the target release date?
- What is the rollout strategy?
- Which tenant contacts need to be notified?

This document is the single reference point for any disagreement about whether something is "in" or "out" of a release. If it is not in the document, it is not in the release.

---

## 11. Features vs. Bug Fixes: Why the Distinction Matters 📋 🔧

The distinction between a feature and a bug fix is not administrative — it determines the version bump, the testing scope, the rollout strategy, the communication approach, and whether a feature flag is required. Getting this wrong causes either under-testing (treating a feature as a patch) or over-process (treating a patch as a feature).

### Defining a Feature

A feature introduces behaviour that did not previously exist. A user can now do something they could not do before, or the system can now do something it could not before.

**AWO ERP feature examples:**
- A daily fuel reconciliation report in the Finance module
- M-Pesa payment recording in the finance transaction flow
- Tenant admins inviting new users from the dashboard without contacting support
- LPG cylinder tracking as a new inventory category

Features always require a design step before coding begins, a feature flag to control rollout, end-user release notes, and a MINOR or MAJOR version bump.

### Defining a Bug Fix

A bug fix corrects behaviour that deviates from how the system was designed or documented to work. The system had a clear intention; the bug means it is not meeting that intention.

**AWO ERP bug fix examples:**
- A pump attendant's cash shortfall calculation producing the wrong amount
- Session tokens failing to refresh for sessions longer than twelve hours
- VAT rounding incorrectly on invoices above KES 100,000
- A filter on the AR aging report not correctly excluding zeroed-out accounts

Bug fixes always require a reproduction case, a failing test written before the fix, and a PATCH version bump.

### The Grey Zone: Improvements, Refactors, and Performance Work

Some changes fit neither definition cleanly, and forcing them into one box or the other leads to either over-communication or under-communication.

**Improvements** are changes where the system works correctly but the experience is poor — a slow API response, a confusing UI flow, an unhelpful error message. The system is doing what it was built to do, but not doing it well. Treat as PATCH if the change is invisible to users; treat as MINOR if it meaningfully changes what a user sees or how they work.

*Example:* The tenant dashboard takes 4 seconds to load because of an unoptimised SQL join. Fixing the query with a targeted index — the dashboard loads in under 500ms — is a PATCH. No new capability exists; something slow is now fast. It does not appear in end-user release notes, but it does appear in the engineering changelog.

**Refactors** restructure code without changing any visible behaviour. Migrating from manual dependency wiring to Google Wire DI, consolidating duplicated repository logic, extracting a domain service from a handler — these are PATCH level changes and appear only in the engineering changelog. Tenants will never read about them, and that is correct.

**Performance work** is PATCH level when it requires no schema changes, and MINOR level when it introduces new database indexes, caching layers, or infrastructure changes that affect how data is stored or retrieved.

### The Distinction in Practice

| Dimension | Feature | Bug Fix |
|-----------|---------|---------|
| Version bump | MINOR or MAJOR | PATCH |
| End-user release notes | Required | Only for High or Critical severity |
| Feature flag required | Always | Rarely |
| Testing scope | Full: unit, integration, manual | Targeted: regression test for the specific defect |
| Rollout strategy | Gradual or segmented | Usually all-at-once — all tenants need the fix |
| Communication | Announcement with "what you can now do" | Brief notification for High/Critical only |
| PM involvement | Full design review | Awareness only for Low/Medium severity |

---

## 12. The Feature Delivery Procedure 📋 🔧

```
STEP 1 — INTAKE
  Feature request received: roadmap item, tenant feedback, or PM initiative.
  Logged with source, date, and requester.

STEP 2 — TRIAGE
  PM evaluates against four questions (Section 8).
  Outcome: Must-have / Nice-to-have / Deferred / Rejected.
  Requester informed of outcome with reasoning.

STEP 3 — DESIGN
  PM writes a brief feature spec: what it does for the user, not how it is built.
  Engineering Lead reviews for technical feasibility and produces an effort estimate.
  Output: agreed specification + estimate. No coding begins without both.

STEP 4 — FLAG CREATION
  Engineer creates the feature flag in AWO ERP's config layer.
  Default: OFF in all environments. This is non-negotiable.
  Flag naming convention: feature.module.capability
    e.g., feature.finance.monthly_reconciliation_report

STEP 5 — DEVELOPMENT
  Feature is built entirely behind the flag.
  Code is committed and deployed to development and staging while the flag is OFF.
  This includes: new handlers, services, repository methods, migrations, tests.

STEP 6 — TESTING
  Flag turned ON in staging only.
  QA testing, edge case validation, performance testing.
  PM signs off on behaviour against the agreed specification.
  Flag remains OFF in production throughout this step.

STEP 7 — RELEASE SCOPE LOCK
  Feature enters the Release Scope Document.
  Technical version number confirmed (MINOR or MAJOR).
  Tenant communication drafted.

STEP 8 — PRODUCTION DEPLOY (Dark)
  Code deployed to production. Flag remains OFF for all tenants.
  Migrations run. Feature exists on the server but is invisible to tenants.
  Minimum dark launch period: 24 hours.

STEP 9 — CONTROLLED ROLLOUT
  Flag turned ON for pilot tenant(s).
  Engineering monitors logs and metrics for 24–48 hours.
  If stable: flag enabled for the next rollout cohort.
  If problems: flag turned OFF for pilot tenant, investigation begins.

STEP 10 — FULL ROLLOUT
  Flag turned ON for all tenants after stable pilot period.
  End-user release notes published.
  Tenant communication sent.

STEP 11 — FLAG RETIREMENT
  Once confirmed stable and universal, the flag is removed from code.
  No more branching — the feature always runs.
  Flag retirement is logged in the engineering changelog.
```

---

## 13. The Bug Fix Procedure 🔧

```
STEP 1 — REPORT
  Bug received via support ticket, log alert, or internal discovery.
  Assigned to an engineer immediately.

STEP 2 — REPRODUCE
  Engineer confirms the bug is real and writes exact reproduction steps.
  If not reproducible after reasonable effort: close with a documented explanation.
  Do not fix what cannot be reliably reproduced.

STEP 3 — CLASSIFY
  Severity assigned: Critical / High / Medium / Low.
  Critical or High: escalate to Engineering Lead immediately.
  The severity classification triggers the rest of the procedure.

STEP 4 — WRITE A FAILING TEST
  Before writing a single line of fix code: write a test that proves the bug exists.
  This test must fail before the fix and pass after.
  This is the guard against the bug returning silently in a future release.

STEP 5 — PATCH
  Implement the fix. The goal is the smallest change that fully resolves the issue.
  Do not refactor while fixing. Do not improve adjacent code while fixing.
  Scope creep in a bug fix is how new bugs are introduced.

STEP 6 — VERIFY
  The failing test now passes.
  All existing tests still pass (no regression).
  Fix verified manually in staging environment.

STEP 7 — DEPLOY
  PATCH version bump.
  Critical / High: deploy via emergency hotfix path (Section 14).
  Medium / Low: queue for next scheduled patch release.

STEP 8 — CONFIRM
  Monitor logs and metrics for 30 minutes post-deploy.
  If the bug was reported by a tenant: confirm with them directly.

STEP 9 — DOCUMENT
  Entry added to engineering changelog.
  High or Critical: add plain-language summary to tenant communication.
```

---

## 14. Emergency Hotfix Procedure 🔧 📋

A hotfix is reserved for Critical or High severity bugs that cannot wait for the next scheduled release. The pressure of an emergency is precisely when discipline matters most — shortcuts taken here become the next emergency.

```
TRIGGER: Critical or High severity bug confirmed in production.

STEP 1 — DECLARE
  Engineering Lead formally declares a hotfix.
  PM is notified immediately.
  Affected tenants receive an initial message within 15 minutes:
  "We are aware of an issue affecting [feature]. Our team is actively working on it.
  We will update you within 30 minutes."

STEP 2 — BRANCH
  Branch from the current production tag — not from dev or staging.
  The production tag is the known good state.

  git checkout -b hotfix/1.5.1 v1.5.0

STEP 3 — FIX
  Minimal fix only. No other changes — no improvements, no refactors.
  Write the failing test. Implement the fix. Verify locally.

STEP 4 — REVIEW
  At least one other engineer reviews the change before it is merged.
  No exceptions. The urgency of an emergency does not justify skipping review.
  A second pair of eyes has caught more than one would-be disaster.

STEP 5 — STAGING FAST-TRACK
  Deploy to staging.
  Verify the specific bug is resolved.
  Run critical path smoke tests — auth, tenant middleware, the affected feature.

STEP 6 — PRODUCTION DEPLOY
  Tag the release: git tag v1.5.1
  Deploy to production.
  Run migration if required.

STEP 7 — MONITOR
  Dedicated monitoring for 60 minutes post-deploy.
  Engineering Lead remains available and reachable.

STEP 8 — COMMUNICATE
  Tenant update as soon as the fix is confirmed:
  "The issue with [feature] has been resolved as of [time]. We apologise for the disruption."
  Internal post-mortem scheduled within 48 hours.

STEP 9 — MERGE BACK
  Merge the hotfix branch back into the development branch.
  This is not optional — a fix that exists only in the hotfix branch will be
  overwritten by the next regular release.
```

---

<br>

# Part IV — When Releases Happen

---

## 15. Release Cadence: How Often AWO ERP Ships 📋

Three models exist for scheduling software releases. Understanding all three explains why AWO ERP's chosen approach is right for this context.

**Continuous delivery** releases changes as soon as they are ready — sometimes many times per day. This works for consumer applications like social media where individual changes are small and the cost of a bad change is low. It requires extremely mature automated testing and feature flag infrastructure. For an ERP managing live financial data for businesses, continuous delivery without that infrastructure is irresponsible.

**Scheduled releases** ship on a fixed calendar — weekly, bi-weekly, monthly. Teams batch changes together. This is predictable for planning but creates deadline pressure that is the single most common cause of corners being cut.

**Milestone-based releases** ship when a defined set of features is complete, regardless of the calendar. Common in traditional enterprise software. Quality is high but timing is unpredictable, and "it's done when it's done" is not a communication strategy.

### AWO ERP's Hybrid Approach

AWO ERP uses a **hybrid of scheduled and milestone-based** delivery:

- **PATCH releases:** As needed, no fixed schedule. When a patch is ready, tested, and verified, it ships. Critical patches ship within hours via the hotfix process.
- **MINOR releases:** Quarterly cadence — one named end-user release per quarter. Tenants can expect a meaningful update every three months, which gives them time to absorb changes, retrain staff where needed, and plan around the update.
- **MAJOR releases:** Milestone-based — a MAJOR version ships when it is genuinely ready, not when a calendar says it should. Forcing a major version onto a schedule creates the conditions for dangerous shortcuts.

Quarterly for MINOR releases is the right rhythm for AWO ERP's tenants specifically. These are businesses — fuel stations closing monthly accounts, hotels managing room occupancy — not individual consumers using a hobby app. A monthly release cycle creates change fatigue: staff must constantly re-learn workflows. An annual release cycle means bug fixes and improvements sit undeployed for months. Quarterly strikes the balance.

---

## 16. Factors That Shape the Exact Release Date 📋

The quarterly cadence sets the window; these factors set the exact date within it.

**Tenant activity patterns.** Every tenant has peak periods. For a fuel station, month-end is intense — shift reconciliations, wetstock reports, cash balancing. For a hotel, public holidays drive high transaction volume. Never release during a known peak period. The ideal release day is a Tuesday or Wednesday morning — lower traffic than Monday, full working week ahead to respond to issues, staff available to act quickly.

**Kenyan regulatory calendar.** KRA VAT filing deadlines, EPRA fuel pricing cycle announcements, NSSF/SHIF submission dates — releases that touch financial calculations or statutory reporting must not ship within 72 hours of a compliance deadline. A broken calculation the night before a KRA submission is a relationship-ending incident.

**Engineering team availability.** A release on a Friday afternoon means weekend on-call support if something goes wrong. A release with the Engineering Lead travelling means degraded rollback capability. Timing must account for who is actually available to respond in the hours immediately following deployment.

**Staging soak time.** A new version must run on staging for a minimum of 48 hours before production. MAJOR releases require a minimum of one week on staging. Time on staging is not waiting — it is actively monitoring for issues that only appear under sustained load or over time.

**Dependency readiness.** If a release depends on an infrastructure change — a new Redis configuration, a Temporal worker update, a database upgrade — the release cannot ship until that dependency is confirmed stable.

---

## 17. Freeze Periods: Non-Negotiable Boundaries 🔧 📋

Freeze periods are the most commonly violated release discipline, and their violation is the most common cause of release-day incidents. The pressure to "just get this one thing in" before a freeze is a pressure that must be resisted every single time.

### Feature Freeze — One Week Before Release

After feature freeze, no new features may be added to the release scope. Features that are not complete by this date are deferred to the next release. Only bug fixes and test improvements may be merged.

**Why this matters:** A feature added in the final week of a release has received no soak time on staging. It has not been tested under real conditions. The engineer who built it is the only person who has seen it run. This is how a last-minute addition becomes a production incident.

### Code Freeze — 48 Hours Before Release

After code freeze, no code changes of any kind may be merged. The only exception is a confirmed Critical severity bug in the upcoming release itself. All remaining time goes into verification: migration dry-runs, staging smoke tests, monitoring baseline capture, release checklist completion.

**Why this matters:** Every code change, no matter how small, restarts the verification clock. A "tiny fix" committed 12 hours before a release has not been tested alongside everything else. The 48-hour window is the buffer that turns "we think it is ready" into "we know it is ready."

---

## 18. Time-Sensitive Releases Outside the Quarterly Cadence 🔧 📋

Some releases cannot wait for the quarterly window. These are not exceptions to the release process — they follow the same discipline, compressed into a shorter time frame.

**Security patches** follow the hotfix path regardless of quarterly timing. A confirmed Critical security vulnerability ships within 24 hours of confirmation. There is no discussion about whether it can wait for the next quarter.

**Regulatory changes** require AWO ERP to ship a compliant update before the effective date of the change. If KRA announces a VAT rate change effective on a specific date, AWO ERP must be updated and deployed before that date — not after. These are treated as PATCH releases if the change is purely numerical, or MINOR releases if new logic is required.

**Broken integrations** caused by third-party API changes — M-Pesa, Pesapal, Shell card systems — are treated as High severity bugs and patched outside the regular schedule. AWO ERP cannot control when these partners change their APIs; it can control how fast it responds.

---

## 19. What Earns a Major Version 📋 🔧

A MAJOR version bump carries significant weight. It tells tenants: something fundamental has changed. Used too liberally, it creates alarm where none is warranted. Used too sparingly, it allows breaking changes to slip through without the preparation tenants deserve.

### The Technical Threshold

A release earns MAJOR status when it meets **at least one** of these conditions:

1. An existing API endpoint changes its request or response structure in a way that breaks current callers
2. A database column, table, or field is removed that existing integrations or reports depend on
3. The authentication or authorisation model changes in a way that invalidates existing sessions or tokens
4. The RLS policy model changes in a way that requires tenant data to be re-scoped
5. A module that tenants actively rely on is retired or fundamentally redesigned in a breaking way
6. The AWO ERP configuration format changes incompatibly — env var names, Viper keys
7. A Temporal workflow interface changes in a way that invalidates currently in-flight workflow executions

### Breaking Changes: Precise Examples

| Change | Breaking? | Why |
|--------|-----------|-----|
| Adding an optional field to an API response | ❌ No | Existing consumers ignore unknown fields |
| Removing a field from an API response | ✅ Yes | Consumers expecting the field will fail |
| Adding a new table to the database | ❌ No | Does not affect existing queries |
| Renaming an existing column | ✅ Yes | All queries referencing the old name fail |
| Adding a new API endpoint at a new route | ❌ No | Does not affect existing routes |
| Changing an existing route's HTTP method (GET → POST) | ✅ Yes | All callers using the old method fail |
| Adding a new permission string | ❌ No | Existing sessions are unaffected |
| Restructuring the permission model (flat → hierarchical) | ✅ Yes | All existing permission checks may fail |

### The Strategic Threshold

Some changes are technically backward-compatible but are architecturally significant enough to warrant a MAJOR version for clarity. Examples in AWO ERP:

- Introducing Temporal as the workflow engine, replacing synchronous long-running handlers
- Migrating from Casbin file-based policies to a dynamic CEL-based ABAC engine
- Changing the multi-tenancy model (e.g., from shared RLS to schema-per-tenant)
- Adding a category of functionality so significant it redefines what the product is — a full payroll engine, or native mobile tenant access

In these cases, a MAJOR version signals to tenants and integrators: *the mental model has changed, even if your integration still works.*

### Choosing Between a Heavily Loaded MINOR and a MAJOR

It is tempting to accumulate features into a large release and call it MAJOR for the ceremony. Resist this. MAJOR is about the *nature* of change, not the *quantity*.

| Scenario | Correct Version |
|----------|----------------|
| 10 new features, all backward-compatible | `MINOR` — e.g., `1.10.0` |
| 1 breaking API change + 5 new features | `MAJOR` — e.g., `2.0.0` |
| Massive internal refactor, no user-visible change | `MINOR` with engineering note |
| Full HR module launch as a new product pillar | `MAJOR` — strategic threshold |
| Security patch removes a deprecated endpoint tenants use | `MAJOR` — if tenants depend on it |

### The Obligation That Comes With a MAJOR Release

Shipping a MAJOR version is not just tagging a git commit. It creates obligations:

1. **Advance notice** — Minimum four weeks before a MAJOR release that requires tenant or integration changes
2. **Migration guide** — A written, step-by-step document for any tenant or integration that must change
3. **Compatibility window** — Where possible, maintain the old behaviour behind a flag for one additional MINOR release cycle to give tenants time to migrate without being forced
4. **Elevated support availability** — Engineering and support team on high availability for 48 hours after a MAJOR release

### Deprecation Policy

When a feature, API endpoint, or behaviour is planned for removal in a future MAJOR release:

1. Mark it as **deprecated** in the current MINOR release
2. Include a `Deprecation: true` header and `Sunset: [date]` in API responses for deprecated endpoints
3. Include a clear "will be removed in vX.0.0" statement in release notes
4. Allow a minimum of **one full quarterly release cycle** to pass before removing the item
5. Remove it in the MAJOR release exactly as announced — no surprises

> The lesson from watching larger platforms: tenants forgive breaking changes they were warned about. They do not forgive surprises.

---

<br>

# Part V — Controlling What Is Live

---

## 20. Feature Flags: The Light Switches of AWO ERP 🔧 📋

A feature flag is a named switch that controls whether a deployed feature is visible and usable by tenants. The code for the feature already lives on the server — the flag determines whether tenants encounter it. Flipping a flag does not require a new deployment, a new build, or any downtime. It is a configuration change that takes effect on the next request.

This is why the principle of *decoupling deploy from release* is achievable in practice. Without feature flags, deploy and release are the same moment. With feature flags, they can be separated by hours, days, or weeks.

### Four Types of Flags

| Type | Purpose | Who Controls It | Lifetime |
|------|---------|----------------|---------|
| **Release flag** | Controls rollout of a new feature to tenants | Engineering / PM | Temporary — retired once fully rolled out |
| **Ops flag** | Emergency kill-switch for a feature causing problems in production | Engineering Lead | Permanent — kept as a standing safety valve |
| **Experiment flag** | A/B test — different tenants see different experiences for comparison | PM | Temporary — retired after a decision is made |
| **Permission flag** | Restricts a feature to specific tenant tiers or subscription levels | PM / Business | Permanent — tied to billing model |

### How Flags Are Stored and Evaluated in AWO ERP 🔧

AWO ERP uses two mechanisms, appropriate to the scope of the flag:

**Platform-level flags** apply globally across the entire platform. They are stored in environment variables or the Viper configuration file and evaluated at startup or per-request via `config.Config`.

```go
// internal/platform/handler/finance_handler.go
// Checking a platform-level flag before serving a handler

if deps.Config.Features.NewFinanceDashboard {
    // The new dashboard is active — route to the new implementation
    return deps.FinanceHandler.NewDashboardHandler(c)
}

// Flag is OFF — serve the stable legacy implementation
return deps.FinanceHandler.LegacyDashboardHandler(c)
```

**Tenant-level flags** apply to individual tenants. They are stored in the `tenant_settings` table in PostgreSQL under RLS, ensuring each tenant's flag state is isolated from every other tenant's. These allow one tenant to be on a feature while another is not.

```go
// internal/service/tenant_service.go
// Checking a tenant-level flag within a request context

tenantSettings, err := deps.TenantService.GetSettings(ctx, tenantID)
if err != nil {
    return err
}

if tenantSettings.Features["new_selling_module"] {
    // This tenant has opted into the new Selling module
    return deps.SellingHandler.NewModuleHandler(c)
}

// Feature not yet enabled for this tenant
return c.JSON(fiber.StatusForbidden, fiber.Map{
    "error": "This feature is not yet available for your account.",
})
```

### The Flag Lifecycle 🔧

Every flag is born off and dies in the code. Between creation and retirement, it passes through clearly defined stages:

```
CREATE
  Flag added to config struct. Default: OFF in all environments.
  Code behind the flag is written and committed.
  ↓
INTERNAL TEST
  Flag turned ON in development environment only.
  Engineer validates the implementation.
  ↓
STAGING
  Flag turned ON in staging. Full QA cycle runs.
  PM signs off against specification.
  ↓
DARK LAUNCH (Production)
  Code deployed to production. Flag remains OFF.
  Minimum 24 hours in this state.
  ↓
PILOT
  Flag turned ON for one or two selected tenants in production.
  Engineering monitors closely for 24–48 hours.
  ↓
GRADUAL ROLLOUT
  Flag turned ON for expanding cohorts of tenants.
  Each cohort observed before the next is enabled.
  ↓
FULL ROLLOUT
  Flag ON for all tenants.
  ↓
RETIRE
  Flag removed from code entirely.
  Feature always runs — no more branching.
  Retirement recorded in engineering changelog.
```

### Flag Debt: The Hidden Cost of Flags That Never Die

Every flag that is created but never retired accumulates as **flag debt**. The cost is not obvious at first, but it compounds:

- Code that branches on multiple flags becomes exponentially complex to reason about. With three flags each having two states, there are eight possible code paths. With five flags, there are thirty-two.
- Engineers stop knowing which flags are active in production, leading to bugs in the "off" path that are never tested.
- Support cannot explain why Tenant A sees a feature that Tenant B does not.
- The next engineer to touch the code wastes time deciphering old branching logic that should have been removed months ago.

**Prevention:**
- Every flag is created with a `// TODO(retire): target v1.X.0` comment in the code
- The Release Scope Document for each version explicitly lists flags to retire in that version
- A monthly engineering review scans for any flag older than 90 days and schedules its retirement

---

## 21. Settings and Configuration: Three Layers 🔧 📋

AWO ERP has a strict three-layer hierarchy for configuration. Every setting belongs to exactly one layer, which determines who can change it, when the change takes effect, and how broadly it applies.

```
╔══════════════════════════════════════════════════════╗
║               PLATFORM LAYER                         ║
║  Affects: every tenant, every user on the platform   ║
║  Owner: Engineering / DevOps                         ║
║  Changed: via deployment (requires restart)          ║
╠══════════════════════════════════════════════════════╣
║               TENANT LAYER                           ║
║  Affects: one tenant and all users within it         ║
║  Owner: AWO ERP team / Tenant Admin                  ║
║  Changed: via admin UI or database update            ║
║           — takes effect immediately, no restart     ║
╠══════════════════════════════════════════════════════╣
║               USER LAYER                             ║
║  Affects: one individual user within a tenant        ║
║  Owner: The individual user                          ║
║  Changed: via user preferences UI                    ║
║           — takes effect immediately                 ║
╚══════════════════════════════════════════════════════╝
```

### Platform-Level Configuration 🔧

Managed via Viper, loaded at startup from environment variables with a fallback to `.env` or `config.yaml` for local development. Production always uses environment variables — never config files committed to version control.

| Setting | Example Value | Why It Matters |
|---------|--------------|---------------|
| `DATABASE_URL` | `postgres://...` | Contains credentials — never in code or git |
| `REDIS_URL` | `redis://...` | Session and cache storage |
| `TEMPORAL_HOST` | `temporal:7233` | Async workflow engine address |
| `JWT_SECRET` | (rotated regularly) | Changing this invalidates all active sessions |
| `CORS_ORIGINS` | `https://awo.so` | Must include all production tenant subdomains |
| `FEATURE_NEW_DASHBOARD` | `true` / `false` | Global platform-level feature flag |
| `BASE_DOMAIN` | `awo.so` | Changing this is a MAJOR version event |

> ⚠️ Platform configuration changes require a deployment restart to take effect. This means they must be coordinated — changing `JWT_SECRET` in production while tenants are logged in will invalidate every active session simultaneously.

### Tenant-Level Settings 📋 🔧

Tenant settings are stored in the PostgreSQL `tenant_settings` table, scoped by tenant ID under RLS. They can be changed without deployment and take effect on the next request.

| Setting | Purpose |
|---------|---------|
| `currency` | `KES`, `USD` — how amounts display throughout the tenant's UI |
| `fiscal_year_start` | Month number — drives all financial reporting periods |
| `vat_rate` | Default VAT rate for this tenant's invoices |
| `timezone` | `Africa/Nairobi` — all date and time display |
| `require_mfa` | Whether multi-factor authentication is enforced for this tenant |
| `allowed_payment_methods` | `["cash", "mpesa", "shell_card"]` — restricts available payment options |
| `features.new_selling_module` | Tenant-level feature flag |

### User-Level Preferences 👤

User preferences are the narrowest layer. They affect only the individual user and are always subordinate to tenant settings — if a tenant setting disables a feature, no user preference can re-enable it.

| Preference | Purpose |
|------------|---------|
| `notification_email` | Whether to receive system notifications by email |
| `dashboard_layout` | Which widgets appear on the home screen |
| `date_format` | `DD/MM/YYYY` vs `MM/DD/YYYY` display format |
| `language` | `en` (English) or `sw` (Swahili) |

### When Configuration Changes Require a Release

| Change Type | Requires a Release? | Requires Restart? | Immediate? |
|-------------|--------------------|--------------------|-----------|
| New platform env var (code must read it) | ✅ Yes | ✅ Yes | After restart |
| Changing an existing env var value | ❌ No | ✅ Yes | After restart |
| New tenant setting column (migration needed) | ✅ Yes | ❌ No | After migration |
| Changing a tenant setting value | ❌ No | ❌ No | ✅ Immediately |
| New user preference (UI + storage) | ✅ Yes | ❌ No | After release |
| Changing a user preference value | ❌ No | ❌ No | ✅ Immediately |

---

<br>

# Part VI — Reaching Tenants

---

## 22. Rollout Models: Choosing How Changes Spread 📋

The code is deployed. The feature flag exists. Now the question is: which tenants get access, in what order, and how fast? The right answer depends on the nature of the change.

### All-at-Once (Big Bang)

The flag moves from OFF to ON for every tenant simultaneously. Simple, immediate, no segmentation needed.

**When this is the right choice:**
- Bug fixes — all tenants are affected by the defect, all should receive the fix together
- Security patches — delay for any tenant creates continued risk
- Regulatory compliance changes — all tenants must comply by the same deadline (KRA, EPRA)
- Small, low-risk UI improvements with no workflow impact

**When this is dangerous:**
- Any feature with a non-trivial migration path
- Features that change established daily workflows (staff will need time to adapt)
- Anything touching financial calculations or RLS policies
- Features not thoroughly tested at scale

### Per-Tenant Pilot

A single tenant receives access before anyone else. This is the "early access" model and the starting point for most meaningful feature rollouts.

The ideal pilot tenant is:
- **Technically comfortable** — their admin can clearly articulate what is working and what is not
- **Representative** — their usage patterns reflect the broader tenant population, not an unusual edge case
- **Trusting and communicative** — they have an established relationship with the AWO ERP team and will provide honest feedback
- **Recoverable** — if the feature causes them problems, those problems can be resolved without data loss

For AWO ERP, a standing cohort of pilot tenants should be maintained — businesses who have agreed to receive early access in exchange for structured feedback. This relationship is valuable and should be cultivated.

### Segmented Rollout by Module Usage

The update reaches only tenants who actively use the relevant module. A Finance Dashboard improvement should first reach tenants who use the Finance module daily — not tenants who primarily use the Selling module and have never opened Finance.

This requires features to be independently flaggable per module:

```sql
-- Enable the Finance Dashboard update only for Finance-active tenants
UPDATE tenant_settings
SET features = features || '{"new_finance_dashboard": true}'::jsonb
WHERE tenant_id IN (
    SELECT DISTINCT tenant_id
    FROM finance_transactions
    WHERE created_at > NOW() - INTERVAL '30 days'
);
```

Tenants who haven't used the Finance module in 30 days will receive the update in the next wave, after the feature is confirmed stable.

### Segmented Rollout by Tenant Tier

The tenant population is segmented by usage intensity and subscription level:

| Segment | Description | Receives Update |
|---------|------------|----------------|
| **Power users** | High daily transaction volume, frequent logins, established workflows | First — best positioned for feedback |
| **Active tenants** | Regular use, stable patterns | Second — after power user validation |
| **Trial tenants** | Recently onboarded, still evaluating | Third — lower stakes if there are issues |
| **Dormant tenants** | Low or no recent activity | Last — or excluded from some updates entirely |

### Time-Based Waves

The tenant population is divided into cohorts, each receiving access on a different day. This is the most controlled rollout model and is appropriate for any significant feature or MAJOR version.

```
Day  0  →  Deploy to production (dark — flag OFF for all tenants)
Day  1  →  Flag ON for Cohort A: pilot tenants (~5% of tenants)
            Monitor logs, error rates, and pilot tenant feedback
Day  3  →  If Day 1–3 clean: Flag ON for Cohort B (~25% of tenants)
Day  7  →  If Day 3–7 clean: Flag ON for Cohort C (~70% of tenants)
Day 10  →  Flag ON for all remaining tenants
Day 14  →  Flag retired — feature always active, branching code removed
```

Each wave is a checkpoint. If a problem appears at Day 3, only 25% of tenants are affected. The flag is turned OFF for Cohort B while the problem is investigated — Cohort A may or may not have the flag turned OFF depending on whether the issue is widespread.

### The Rollout Decision Matrix

| Release Type | Recommended Rollout |
|-------------|---------------------|
| Critical bug fix | All-at-once — immediately |
| Security patch | All-at-once — immediately |
| Regulatory change (KRA, EPRA deadline) | All-at-once — by the deadline |
| Minor UI improvement, low risk | All-at-once |
| New minor feature, low complexity | Pilot → all-at-once within 48 hours |
| New minor feature, moderate complexity | Pilot → tier-segmented → all over 1 week |
| New module or major feature | Pilot → module-segmented → time-based waves over 2 weeks |
| MAJOR version | 4-week advance notice + time-based waves over 2 weeks |

---

## 23. Rollout Communication: What Tenants Know and When 👤 📋

Tenants should never be surprised by a change in their software. Surprise — even positive surprise — erodes trust because it signals that the team does not respect the tenant's need to plan and manage their own operation. Communication follows a three-stage model.

### Before the Rollout

- **MAJOR releases and significant MINOR features:** Notify tenants 2–4 weeks in advance with a plain-language "what's changing" summary. Give them time to prepare staff, update any manual processes, or raise concerns.
- **Smaller MINOR features:** Notify one week in advance.
- **Bug fixes:** No advance notice needed unless the fix visibly changes behaviour — in which case a brief note ("we corrected the cash shortfall calculation; here is what changed") is appropriate.

### At the Rollout

A worked example — rolling out the Monthly Reconciliation Report to a fuel station tenant:

**Email to tenant admin on rollout day:**

> **AWO ERP — Ukuaji Update: New Monthly Reconciliation Report**
>
> Starting today, you can generate a complete monthly reconciliation report directly from your Finance section.
>
> The report includes all daily sales by pump, dip readings, and a payment method breakdown (cash, M-Pesa, Shell card), ready to export to Excel.
>
> To access it: Finance → Reports → Monthly Reconciliation.
>
> Questions? Reply to this email or reach us at support@awo.so.

Notice: no technical jargon, no version numbers in the body, no mention of API endpoints or database migrations. The only question the tenant cares about is "what can I now do that I couldn't do before?"

### After the Rollout

- A support channel remains open for 72 hours specifically for questions about the new feature
- A feedback link is included in the rollout communication — structured feedback from tenants at this moment is the most valuable product input available
- If the rollout caused any disruption, a follow-up communication within 48 hours explains what happened and what has been done to prevent recurrence

---

## 24. Multi-Tenant Specifics in AWO ERP 🔧 📋

### RLS Implications of Schema Changes

AWO ERP uses PostgreSQL Row-Level Security to ensure every query runs within a tenant's isolated data context. Schema changes that affect tables governed by RLS policies require extra care.

**Safe changes (additive):**
- Adding a new table — apply the RLS policy immediately, before any data is inserted
- Adding a nullable column to an existing table
- Adding a new index

**Unsafe changes (require a multi-step migration):**
- Renaming a column that RLS policies reference directly
- Changing the data type of a tenant-scoped column
- Altering or dropping an existing RLS policy

For unsafe changes, spread the migration across multiple releases:

```sql
-- Release 1: Add the new column alongside the old one
ALTER TABLE transactions ADD COLUMN tenant_ref UUID;

-- Application code (same release): Write to BOTH columns simultaneously
-- This ensures no data is lost regardless of which column is queried

-- Release 2: Backfill the new column from the old one for all existing rows
UPDATE transactions SET tenant_ref = legacy_tenant_id WHERE tenant_ref IS NULL;

-- Release 3: Apply NOT NULL constraint and drop the old column
ALTER TABLE transactions ALTER COLUMN tenant_ref SET NOT NULL;
ALTER TABLE transactions DROP COLUMN legacy_tenant_id;
```

This three-release pattern ensures no single migration breaks in-flight requests during a deployment.

### What Happens When a Pilot Tenant Reports a Bug

Because AWO ERP runs one codebase for all tenants, a bug on a feature-flagged feature for the pilot tenant may not affect tenants who have not yet received the flag. This is exactly the scenario flags are designed to handle gracefully.

```
1. Bug confirmed on pilot tenant.

2. Immediately turn the feature flag OFF for the pilot tenant.
   No code deployment needed — the flag change takes effect on the next request.

3. Confirm whether other tenants with the flag ON are also affected.
   If YES: turn the flag OFF for all tenants with the flag currently ON.
           The feature is effectively rolled back without touching production code.
   If NO (isolated to pilot tenant): investigate tenant-specific data or configuration.

4. Fix the bug in the development environment.
   Follow the full bug fix procedure: failing test first, then fix, then verify.

5. Re-enable the flag for the pilot tenant after confirming the fix.
   Proceed with the rollout from the pilot stage.
```

This scenario illustrates concretely why feature flags matter. Without them, rolling back a feature means redeploying a previous binary — a slow, high-risk operation. With flags, the rollback is a database update that takes effect in milliseconds.

### Tenant Admin Visibility Into Their Version State 👤

Currently, tenant admins have limited visibility into what version of AWO ERP they are running. The near-term target for the admin dashboard includes:

- The current end-user version name prominently displayed: "You are on AWO ERP 2025 Q2 — *Ukuaji*"
- A "What's new" section listing features that have been enabled for their specific tenant
- A "Coming soon" section for tenants who opt into early notifications
- A link to the full release notes for the current version

This transparency is not cosmetic. Tenants who understand what version they are on and what changed can self-diagnose many support issues before contacting the team. Every hour of tenant self-service is an hour the support team can spend on genuinely complex problems.

---

<br>

# Part VII — Engineering Procedures

---

## 25. Database Migrations 🔧

### File Naming Convention

AWO ERP uses `golang-migrate`. Migration files follow a strict naming format:

```
{MODULE_ID}{SEQUENCE}_description.up.sql
{MODULE_ID}{SEQUENCE}_description.down.sql
```

`MODULE_ID` is a 3-digit module prefix. `SEQUENCE` is a 3-digit sequential number within that module.

```
001001_create_tenants_table.up.sql      # Core module (001), first migration
001001_create_tenants_table.down.sql
001002_add_tenant_status_column.up.sql
001002_add_tenant_status_column.down.sql
002001_create_users_table.up.sql        # IAM module (002), first migration
003001_create_chart_of_accounts.up.sql  # Finance module (003), first migration
```

**Absolute rules:**
- Descriptions use underscores, lowercase, and past-tense verbs: `create`, `add`, `drop`, `rename`, `alter`
- Never reuse a sequence number — even if a migration is deleted
- Never modify a migration that has already been applied in any environment
- Both `.up.sql` and `.down.sql` must be written and reviewed before the migration is committed

### Running Migrations

```bash
# Apply all pending migrations
make migrate-up

# Roll back the last migration
make migrate-down

# Roll back a specific number of migrations
make migrate-down-n N=3

# Check the currently applied migration version
make migrate-version

# Force a specific migration version — only for correcting failed states, never routine use
make migrate-force VERSION=001005
```

Migrations run automatically as part of every production deployment. They are never skipped, and they are never run manually on production outside of the deployment process.

### Zero-Downtime Migration Patterns

For tables with significant row counts, standard `ALTER TABLE` commands acquire locks that prevent reads and writes for the duration of the operation. At scale, this means downtime. These patterns avoid that.

**Pattern 1: Add columns as nullable first**

```sql
-- .up.sql
-- Phase 1: Add column without a NOT NULL constraint — no lock held for extended period
ALTER TABLE fuel_sales ADD COLUMN reconciled_by UUID;

-- The application writes reconciled_by when it is available.
-- Existing rows have NULL, which is valid.

-- Phase 2 (a future release, after all rows are populated):
ALTER TABLE fuel_sales ALTER COLUMN reconciled_by SET NOT NULL;
```

**Pattern 2: Create indexes concurrently**

```sql
-- This acquires a full table lock — never use on large tables in production:
CREATE INDEX idx_fuel_sales_tenant ON fuel_sales(tenant_id);

-- This builds the index without locking reads or writes:
CREATE INDEX CONCURRENTLY idx_fuel_sales_tenant ON fuel_sales(tenant_id);
```

**Pattern 3: Rename via dual-write**

Renaming a column is instant in SQL but immediately breaks all live queries using the old name. The safe approach spans multiple releases:

```sql
-- Release 1: Add the new column
ALTER TABLE accounts ADD COLUMN account_code VARCHAR(20);

-- Application: write to both old_code and account_code simultaneously

-- Release 2: Backfill account_code from old_code for all existing rows
UPDATE accounts SET account_code = old_code WHERE account_code IS NULL;

-- Release 3: Add NOT NULL constraint, then drop the old column
ALTER TABLE accounts ALTER COLUMN account_code SET NOT NULL;
ALTER TABLE accounts DROP COLUMN old_code;
```

### Migration Safety Checklist

Before any migration runs in production:

- [ ] Tested in development with a realistic data volume
- [ ] Run successfully in staging without errors
- [ ] `.down.sql` tested and confirmed to restore the previous state exactly
- [ ] Tables with > 10,000 rows use a zero-downtime pattern
- [ ] RLS policies applied to any new tables before data is inserted
- [ ] No references to extensions, functions, or types not present in production
- [ ] Timing estimate documented for migrations expected to run longer than 60 seconds

---

## 26. Code Generation and Build 🔧

The rule in AWO ERP is absolute: **generated files are never hand-edited**. They are regenerated from their sources whenever the source changes. Editing a generated file creates a false version of truth — the next `make generate` will silently overwrite it.

### Pre-Release Generation Checklist

```bash
# Run the full generation sequence — must complete with no errors
make generate
# This executes: sqlc → mockgen → wire (in that order — order matters)

# Verify no files changed as a result
git status
# Expected: clean working tree
# If files changed: a source file was modified without regenerating
# The release must not proceed until the source of the diff is identified and resolved
```

### SQLC Query Regeneration

```bash
# Reads:  db/queries/*.sql  and  sqlc.yaml
# Writes: db/sqlc/*.go      — do not hand-edit these files

make sqlc
```

After adding, modifying, or removing any file in `db/queries/`, run `make sqlc`. The generated code in `db/sqlc/` must always reflect the current state of the SQL query files. Committing query changes without regenerating causes compilation failures on the next build.

### Wire Dependency Injection Regeneration

```bash
# Reads:  wire.go  (files carrying the //go:build wireinject tag)
# Writes: wire_gen.go  (carries the //go:build !wireinject tag — do not hand-edit)

make wire
```

Run `make wire` after adding a new provider to the dependency graph, modifying a `Dependencies` struct, or introducing a new domain handler. The two files are always a pair — `wire.go` is the declaration, `wire_gen.go` is the result.

### Mock Regeneration

```bash
# Reads:  repository interface definitions
# Writes: test/mocks/*.go  — do not hand-edit these files

make mock
```

After modifying any repository interface — adding a method, changing a signature, removing a method — regenerate mocks. Tests using the old mock signature will fail to compile. This is intentional and acts as a contract-change detector: if the interface changed but the tests were not updated, the compilation failure catches it before the tests even run.

---

## 27. Release Workflow 🔧

### Branching Strategy

```
feature/description     →  dev     (merged via reviewed PR)
bugfix/issue-id         →  dev     (merged via reviewed PR)
dev                     →  staging (automated or manually promoted)
staging                 →  main    (release PR — requires Engineering Lead + PM sign-off)
hotfix/description      →  main    (emergency — branched from the current production tag)
```

The `main` branch always represents the exact current state of production. No direct commits to `main` under any circumstances.

Branch names follow a consistent format:
- `feature/selling-module-order-entry`
- `bugfix/123-vat-rounding-on-large-invoices`
- `hotfix/rls-bypass-in-session-middleware`

### Pull Request Requirements

**Every PR into `dev` requires:**
- [ ] At least one reviewer who is not the author — no exceptions
- [ ] All existing tests passing: `make test`
- [ ] No new linting errors: `make lint`
- [ ] For features: at least one unit test covering the new logic
- [ ] For bug fixes: a regression test that fails without the fix and passes with it
- [ ] Codegen verified clean: `make generate` produces no diffs
- [ ] Migration files included if any schema change was made

**Release PRs from `staging` to `main` additionally require:**
- [ ] Engineering Lead review and explicit approval
- [ ] Product Manager sign-off on release scope
- [ ] Staging smoke test results documented and attached
- [ ] Link to the Release Scope Document in the PR description
- [ ] Migration dry-run completed and verified in staging

### Tagging a Release and Maintaining the Changelog

```bash
# After the release PR is merged to main:
git checkout main
git pull origin main

# Create an annotated tag with a meaningful message
git tag -a v1.5.0 -m "Release v1.5.0 — AWO ERP 2025 Q2 Ukuaji"
git push origin v1.5.0
```

The `CHANGELOG.md` at the repository root is updated before every release. It follows a consistent structure:

```markdown
## [1.5.0] — 2025-06-15 (AWO ERP 2025 Q2 — Ukuaji)

### Added
- Finance: Monthly reconciliation report with pump-level breakdown [#45]
- IAM: Tenant admin user invitation flow [#52]

### Fixed
- Fuel sales: VAT calculation rounding error on amounts > KES 100,000 [#61]
- Sessions: Token refresh failing for sessions older than 12 hours [#63]

### Changed
- Audit log entries now include the user agent string for better traceability [#58]

### Deprecated
- GET /api/v1/finance/summary — will be removed in v2.0.0.
  Replacement: GET /api/v1/finance/reports/summary

### Engineering (internal — not in tenant release notes)
- Wire DI: Added SellingHandler to Dependencies struct
- SQLC: Regenerated after adding reconciliation report queries
- Refactored TenantMiddleware to reduce per-request allocations
```

---

## 28. End-to-End Release Checklist 🔧

```
ONE WEEK BEFORE RELEASE
  [ ] Release Scope Document finalised and shared with PM and Engineering Lead
  [ ] Feature freeze in effect — no new features accepted
  [ ] All feature flags created and confirmed OFF in production
  [ ] Tenant communication drafted for MINOR / MAJOR releases
  [ ] All migration files written and peer-reviewed

48 HOURS BEFORE RELEASE
  [ ] Code freeze in effect
  [ ] Staging environment fully updated to release candidate
  [ ] make generate run — git status shows clean working tree
  [ ] make test — all tests passing
  [ ] make lint — zero errors
  [ ] Config validation passing in staging: make config-validate
  [ ] Migration dry-run completed in staging — timing documented
  [ ] Monitoring baseline captured from staging (latency, error rate, query times)

RELEASE DAY
  [ ] Engineering Lead and at least one other engineer confirmed available
  [ ] No blackout period active: no KRA deadlines, no tenant peak periods
  [ ] Final staging smoke test passed — all critical paths verified
  [ ] Database snapshot confirmed (pre-migration backup)
  [ ] Deploy to production
  [ ] Migrations applied automatically — verify with make migrate-version
  [ ] GET /health returns new version number with "status":"ok"
  [ ] First log lines show new version string
  [ ] Git tag created and pushed to origin
  [ ] CHANGELOG.md updated and committed to main

FIRST TWO HOURS POST-RELEASE
  [ ] Error rates in Zerolog: no spike vs pre-release baseline
  [ ] Prometheus metrics: latency and 5xx rates within baseline
  [ ] OpenTelemetry traces: key journeys completing without errors
  [ ] Feature flag enabled for pilot tenant(s)
  [ ] Pilot tenant confirmed able to access new feature
  [ ] Zero support tickets in first hour

24 HOURS POST-RELEASE
  [ ] Expand flag rollout per the rollout plan
  [ ] Tenant communication sent (MINOR / MAJOR releases)
  [ ] Engineering retrospective notes captured

14 DAYS POST-RELEASE
  [ ] All fully-rolled-out feature flags retired from code
  [ ] Release retrospective completed and documented
```

---

## 29. Deployment 🔧

### Environment Matrix

| Environment | Purpose | Data | Migrations | Feature Flags |
|-------------|---------|------|------------|---------------|
| **Development** | Daily engineering work | Test data only | Run freely | All flags ON by default |
| **Staging** | Pre-release validation | Synthetic / anonymised tenant data | Run before each test cycle | Match production intent |
| **Production** | Live tenant operations | All real tenant data | Run as part of each deploy | Controlled per rollout plan |

Staging must closely mirror production in data volume and configuration. A bug that only manifests at production scale is often a staging environment gap — not a code problem.

### Deployment Steps

```bash
# 1. Pull the latest main branch
git checkout main && git pull origin main

# 2. Build the production binary
make build
# Output: bin/awo-erp

# 3. Validate configuration against production environment
make config-validate
# This confirms: DATABASE_URL reachable, REDIS_URL reachable, JWT_SECRET set,
# TEMPORAL_HOST reachable, CORS_ORIGINS set for production domain,
# no development-only flags (DEBUG=true) active

# 4. Apply pending database migrations
make migrate-up

# 5. Replace the running binary (zero-downtime with process manager)
systemctl restart awo-erp
# or: pm2 reload awo-erp

# 6. Verify the health check returns the new version
curl https://app.awo.so/health
# Expected: {"status":"ok","version":"1.5.0","db":"connected","redis":"connected"}

# 7. Confirm version in the first log lines
# Zerolog output: {"level":"info","version":"1.5.0","event":"server_starting",...}

# 8. Confirm migration version matches expected
make migrate-version
```

### Rollback

**When to rollback:** Error rate spikes more than 2x baseline within 30 minutes of deploy; a Critical bug is confirmed in production; a migration caused unexpected data state.

```bash
# Application rollback — deploy the previous binary
git checkout v1.4.2
make build
systemctl restart awo-erp

# Database rollback — reverse the migrations applied in this release
make migrate-down-n N=2  # Replace 2 with the number of migrations in this release
```

**Before rolling back:** Turn the feature flag OFF for any tenants in the pilot cohort. Inform them that the feature has been temporarily withdrawn. Data written during the new version window must be reviewed if a data-transforming migration is rolled back — additive migrations (new nullable columns) are generally safe to reverse; migrations that transformed existing data are not.

---

<br>

# Part VIII — Post-Release: Monitoring and Communication

---

## 30. Monitoring a Release 🔧

### Structured Log Search with Zerolog

AWO ERP uses Zerolog for structured logging. Every log line carries the application version, making it possible to compare error rates between two versions in a single log query — invaluable when a canary release is running alongside the previous version.

At startup, the first log entry should always be:

```json
{
  "level": "info",
  "version": "1.5.0",
  "environment": "production",
  "event": "server_starting",
  "time": "2025-06-15T08:00:00Z"
}
```

During and after a release, watch for elevated error activity:

```bash
# Count ERROR-level log entries per minute after deployment
# In your log aggregation tool:
level=error | count() by bin(1m)

# Identify any new error messages not seen before the release
level=error | where time > [deploy_timestamp] | group by message | sort count desc
```

### OpenTelemetry Trace Verification

After every release, spot-check traces for key user journeys. A healthy trace for a standard AWO ERP request should look like this:

```
HTTP Request to POST /api/v1/finance/transactions          (38ms total)
  └── TenantMiddleware — resolves tenant from subdomain    (2ms)
  └── AuthMiddleware — validates JWT, loads user context   (4ms)
  └── FinanceHandler.RecordTransaction                     (30ms)
      └── FinanceService.CreateTransaction                 (27ms)
          └── TransactionRepository.WithTenant             (25ms)
              └── SQLC: insert_transaction query           (18ms)
```

Check specifically:
1. **Login flow** — `POST /api/v1/auth/login` completes in under 200ms with no errors
2. **Tenant middleware** — every request shows a `TenantMiddleware` span with `tenant_id` populated
3. **New endpoints** — any endpoint added in this release should appear in the trace explorer with expected latency
4. **N+1 detection** — a new feature that issues database queries in a loop will appear as repeated query spans

### Prometheus Metric Baselines

Capture these baseline values before every release and compare post-deploy:

| Metric | Healthy Baseline | Alert Threshold |
|--------|-----------------|----------------|
| HTTP request latency P99 | < 500ms | > 1,000ms |
| HTTP error rate (5xx responses) | < 0.1% | > 1% |
| Database query latency P95 | < 100ms | > 300ms |
| Active Temporal workflow count | Stable | > 2× normal |
| Redis cache hit rate | > 80% | < 50% |

A regression in any of these metrics within the first 60 minutes of a release is grounds for immediate investigation and potentially triggering a rollback.

---

## 31. Release Communication 📋 👤

### Internal Handoff: Engineering → Product → Support

Before any tenant-facing communication goes out, the Engineering Lead produces a brief internal release note for the Product Manager and support team. This note covers:

- What was released in technical terms
- Which feature flags are active and for which tenants
- Any known edge cases or limitations in the current version
- What support should watch for in incoming tickets
- How to reproduce the fix for each resolved bug — for support use when closing related tickets

Support cannot answer tenant questions well if they learn about a release at the same time tenants do.

### External: Tenant-Facing Release Notes 👤

Tenant release notes follow a fixed structure and use plain business language throughout. The rule: if a sentence could appear in a technical specification, it should not appear in a tenant release note.

---

**AWO ERP — Ukuaji Update (June 2025)**

**What's new for you:**

**Finance: Monthly Reconciliation Report**
You can now generate a complete monthly reconciliation for your fuel station directly from the Finance section. The report covers all daily sales by pump, dip readings, and a payment method breakdown — cash, M-Pesa, and Shell card — ready to export to Excel.
*To access: Finance → Reports → Monthly Reconciliation*

**User Management: Invite Team Members**
Tenant admins can now invite new staff members without contacting support. Enter their email address, assign their role, and they'll receive an invitation link.
*To access: Settings → Users → Invite User*

**Fixes in this update:**
- VAT calculations on invoices above KES 100,000 are now correctly rounded
- Login sessions no longer time out unexpectedly during long working sessions

**Questions?** Contact us at support@awo.so or reply to this email.

---

### Translating the Engineering Changelog into Tenant Language

The engineering changelog and the tenant announcement are written for entirely different readers. This is the translation process:

| Engineering Changelog Entry | Tenant Release Note |
|----------------------------|---------------------|
| `Added GET /api/v1/finance/reports/monthly-reconciliation` | "You can now generate a monthly reconciliation report from Finance" |
| `Fixed rounding error in PATCH /api/v1/finance/invoices for amount > 100000` | "VAT calculations on large invoices are now correctly rounded" |
| `Refactored repository layer to use interface-based SQLC adapters` | *(Not mentioned — internal change invisible to tenants)* |
| `Added Wire provider for SellingHandler to Dependencies struct` | *(Not mentioned — internal change invisible to tenants)* |
| `Deprecated GET /api/v1/finance/summary — sunset in v2.0.0` | "Note: The Finance Summary screen will be replaced with a new Reports section in our next major update. No action is needed now." |

The guiding rule: if the change is invisible to the tenant in their daily use of AWO ERP, it does not appear in tenant communications. If it changes what they see, what they can do, or what they need to do differently — it must be communicated, in language they use.

### Incident Communication 📋 👤

If a release causes a production issue, communication must be honest, immediate, and in plain language. Three stages:

**Within 15 minutes of confirming the incident:**
> "We are aware of an issue affecting [feature] following today's update. Our team is actively investigating. We will send an update within 30 minutes."

**When the cause is identified:**
> "We have identified the cause of the issue with [feature] and are implementing a fix. We expect to have this resolved by [time estimate]."

**When resolved:**
> "The issue with [feature] has been resolved as of [time]. [One sentence: what was happening and that it is fixed.] We apologise for the disruption to your operations."

**48 hours after resolution:**
> A brief post-mortem sent to affected tenants: what happened, why it happened, and what specific steps have been taken to prevent it happening again.

Never say "our systems experienced an unexpected technical anomaly." Say "a bug introduced in today's update was causing the cash shortfall report to display incorrect totals for some transactions."

---

<br>

# Part IX — Industry Reference

---

## 32. How Others Handle This 📋

### LaunchDarkly: Enterprise Flag Management

LaunchDarkly is the market leader in feature flag platforms. Its targeting model — flags ON for "tenants on the Pro plan" or "users in a specific region" or "5% of all traffic" — is exactly the capability AWO ERP needs for tenant segmentation. Flag changes propagate in milliseconds to all running instances without a deployment. A full audit log records every flag change with who made it and when.

**Lesson for AWO ERP:** The targeting model is the goal. LaunchDarkly's licensing cost makes it prohibitive at AWO ERP's current scale — Unleash (below) is the right near-term alternative.

### Unleash: Open-Source Self-Hosted Flags

Unleash provides the core feature flag capability of LaunchDarkly at zero licensing cost, self-hosted. It has a production-ready Go SDK and covers the capabilities AWO ERP needs: tenant-level targeting, percentage-based rollout, environment-specific flag states (dev/staging/production), and an admin UI that allows the PM to manage flags without writing SQL.

**Recommendation:** Deploy Unleash as a self-hosted Docker container. Integrate the Go SDK. Migrate existing flags from environment variables to Unleash incrementally — this does not require a MAJOR version bump and can be done module by module.

### GitLab: Flags Integrated With CI/CD

GitLab integrates feature flag configuration directly into its deployment pipeline. A merge request can include flag state changes that are applied atomically with the code that the flags govern — eliminating the risk of flags and code being out of sync.

**Lesson for AWO ERP:** Once a CI/CD pipeline is in place (see Section 33), flag changes should be versioned alongside code changes in the same pull request.

### Stripe: Config-as-API, Tenant-Scoped Feature Access

Every Stripe account has a feature set determined server-side by their relationship with Stripe. Merchants do not install a different version of Stripe — they use the same API, but certain capabilities are accessible or not based on their account configuration. The feature gating is invisible to the integration.

**Lesson for AWO ERP:** Feature access by tenant tier — trial, standard, enterprise — should be managed in the tenant settings layer, not by deploying different binaries or by branching on client-side configuration.

### PostgreSQL: Predictable Major Version Cadence

PostgreSQL releases one major version per year with a five-year support window. This predictability allows their users — including AWO ERP — to plan major upgrades years in advance. The deprecation timeline is communicated clearly and adhered to without exceptions.

**Lesson for AWO ERP:** Predictability in the release calendar is itself a feature. Tenants who can plan around a known quarterly cadence are tenants who trust the platform.

---

<br>

# Part X — Gaps and Roadmap

---

## 33. Known Gaps and Future Work 📋 🔧

### CI/CD Pipeline

AWO ERP does not currently have an automated CI/CD pipeline. All builds, tests, and deployments are performed manually. Every manual step in a process is an opportunity for human error — and the release checklist has many steps.

**Target state:** A GitHub Actions pipeline that automatically:
- Runs `make test` and `make lint` on every pull request — no PR is mergeable with failing tests
- Runs `make generate` and verifies the working tree is clean
- Builds the binary and runs smoke tests on merge to `dev`
- Automatically promotes to staging on merge to `staging`
- Requires a manual approval gate before any production deployment

### Integration Test Coverage

The `scripts/seed.sh` script provides basic smoke testing but does not cover the critical user journeys that tenants depend on daily. Before CI/CD can be trusted to gate releases, the integration test suite in `test/integration/` should cover:

- Authentication: login, token refresh, MFA enforcement
- Tenant isolation: RLS isolation confirmed between two tenants in the same test run
- Finance module: account creation, transaction recording, report generation
- User management: invite, role assignment, deactivation
- Every Critical and High severity bug ever fixed — regression tests must prevent their return

### Feature Flag Service: Current vs. Target State

| Dimension | Current | Target |
|-----------|---------|--------|
| Flag storage | Env vars + tenant settings JSONB | Unleash (self-hosted Docker) |
| Targeting | Per-tenant manual SQL update | Rule-based: tier, usage, geography |
| PM visibility | Code only | Admin dashboard in Unleash UI |
| Percentage rollout | Manual cohort SQL | Automated in Unleash |
| Flag retirement tracking | TODO comments in code | Automated flag age alerts |
| Audit trail | Audit service | Searchable flag history in Unleash |

### Blue/Green and Canary Deployment

AWO ERP currently uses a **rolling replace** strategy — the running binary is replaced directly. This means a brief restart window during each deployment. Two more sophisticated strategies are the right long-term targets:

**Blue/Green:** Two identical production environments exist simultaneously. The new version deploys to the inactive environment, is tested, and then all traffic is switched at the load balancer — instant cutover, instant rollback capability. Requires infrastructure investment in a second environment.

**Canary:** A small percentage of production traffic (5–10%) routes to the new version while the remainder stays on the current version. Errors in the canary are caught before they affect the full tenant population. Requires a load balancer with weighted routing capability.

Both require containerised (Docker) deployment and a CI/CD pipeline as prerequisites.

### Tenant Self-Service Beta Opt-In

Currently, tenants are assigned to beta cohorts by the AWO ERP team. The target is to allow tenant admins to manage their own beta participation:

- View available beta features from their admin dashboard
- Opt in to specific betas independently
- Opt out if a beta feature disrupts their workflow
- Submit structured feedback on beta features from within the product

This capability requires the in-app admin dashboard feature, Unleash integration, and a clear policy defining what AWO ERP commits to when a tenant opts into a beta.

---

<br>

# Appendices

---

## Appendix A — Version Decision Tree

```
A change is ready to version. What does it get?
│
├── Does it break any existing integration, API consumer,
│   or correctly-configured tenant system?
│   └── YES ──→ MAJOR  (e.g., 1.9.4 → 2.0.0)
│
├── Does it add new functionality without breaking anything?
│   └── YES ──→ MINOR  (e.g., 1.4.2 → 1.5.0)
│
└── Does it fix something broken without adding anything new?
    └── YES ──→ PATCH  (e.g., 1.5.0 → 1.5.1)
```

---

## Appendix B — Rollout Decision Tree

```
A change is ready to reach tenants. Which rollout model?
│
├── Security patch or regulatory requirement (KRA, EPRA)?
│   └── YES ──→ All-at-once — immediately
│
├── Bug fix affecting all tenants?
│   └── YES ──→ All-at-once — next maintenance window
│
├── New feature, low complexity and low risk?
│   └── YES ──→ Pilot tenant → All-at-once within 48 hours
│
├── New feature, moderate complexity?
│   └── YES ──→ Pilot → Tier-segmented → All over 1 week
│
└── Major feature or MAJOR version?
    └── YES ──→ 4-week advance notice → Pilot → Time-based
                waves over 2 weeks → Full rollout → Flag retired
```

---

## Appendix C — Release Calendar Template

```
AWO ERP Q[N] [YEAR] — [Release Name]
Technical version target: [X.Y.0]

[DATE − 35 days]   Release planning meeting — scope finalised
[DATE − 21 days]   Tenant communication sent (MINOR / MAJOR releases)
[DATE − 14 days]   All features under active development
[DATE −  7 days]   ★ FEATURE FREEZE
[DATE −  5 days]   Staging deployment complete — QA begins
[DATE −  2 days]   ★ CODE FREEZE
[DATE −  1 day]    Migration dry-run + final staging smoke test
[DATE]             ★ RELEASE DAY — production deploy
[DATE +  1 day]    Pilot tenant flag enabled
[DATE +  3 days]   Cohort A rollout (if Day 1–3 clean)
[DATE +  7 days]   Cohort B rollout (if Day 3–7 clean)
[DATE + 10 days]   Full rollout — all tenants
[DATE + 14 days]   Feature flags retired — release closed
[DATE + 21 days]   Retrospective documented
```

---

## Appendix D — Severity-to-Action Quick Reference

| Severity | Confirmed In Production → | Target Resolution | Communication |
|----------|--------------------------|-------------------|---------------|
| **Critical** | Declare hotfix immediately | Hours | Tenant notification within 15 min |
| **High** | Hotfix or expedited patch | Within 24 hours | Tenant notification within 1 hour |
| **Medium** | Queue for next patch release | Within the sprint | No proactive communication |
| **Low** | Queue for next minor release | Best effort | No proactive communication |

---

*AWO ERP — Built for businesses that move fast. Updated to keep them moving.*

*Document maintained in the AWO ERP repository at `docs/release-handbook.md`*
