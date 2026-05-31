# UI Module — Complete Implementation Guide

> **Authoritative reference for how Awo ERP builds its frontend.**
> Covers the custom shell in `web/`, AMIS integration, UX principles, Go schema builders, and implementation patterns for every surface type.

---

## Table of Contents

1. [Executive Summary](./01-executive-summary.md)
2. [Why AMIS](./02-why-amis.md)
3. [Architecture Overview](./03-architecture-overview.md)
4. [Shell Implementation (`web/index.html`)](./04-shell-implementation.md)
5. [Dark Mode](./05-dark-mode.md)
6. [AMIS Core Concepts](./06-amis-core-concepts.md)
7. [API Contract](./07-api-contract.md)
8. [Go Schema Builder](./08-go-schema-builder.md)
9. [Component Reference](./09-component-reference.md)
10. [Advanced Components](./10-advanced-components.md)
11. [Page Patterns](./11-page-patterns.md)
12. [Routes & Auth](./12-routes-and-auth.md)
13. [UX Personas](./13-ux-personas.md)
14. [Layout Principles](./14-layout-principles.md)
15. [Forms](./15-forms.md)
16. [Tables & Lists](./16-tables-and-lists.md)
17. [Navigation](./17-navigation.md)
18. [Theming & Tenant Branding](./18-theming-and-branding.md)
19. [Notifications & Alerts](./19-notifications.md)
20. [Personalisation](./20-personalisation.md)
21. [Accessibility](./21-accessibility.md)
22. [Adoption & Performance](./22-adoption-and-performance.md)
23. [Design Tokens](./23-design-tokens.md)
24. [Known Issues & Workarounds](./24-known-issues.md)
25. [Implementation Checklist](./25-implementation-checklist.md)
26. [The AMIS Way](./26-the-amis-way.md) ⭐ Read this first if you're new to the project
27. [Decisions & Opinions](./27-decisions-and-opinions.md) ⭐ Architecture decisions with reasoning

---

## Start Here

If you are new: read [§26 The AMIS Way](./26-the-amis-way.md) before anything else. It explains what AMIS is for, where the current implementation drifts from it, and the decisions that will shape the next phase.

If you are building a new page: [§25 Implementation Checklist](./25-implementation-checklist.md) before marking it done.

If you are making an architecture decision: [§27 Decisions & Opinions](./27-decisions-and-opinions.md) first — the decision may already be made.

---

## Quick Reference

| Question | Answer |
|---|---|
| Where is the shell? | `web/index.html` — custom HTML/CSS/JS |
| How does routing work? | Hash-based (`#dashboard`, `#users`) |
| Where do schemas live? | `web/schemas/pages/{route}.json` (static files for now) |
| How does dark mode work? | `html.dark` class + CSS variable overrides (see [§05](./05-dark-mode.md)) |
| What envelope does the backend use? | `{success, data, meta}` — the fetcher bridges this to AMIS |
| What envelope does AMIS expect? | `{status: 0, data: ...}` |
| Where is the menu defined? | JS `menuConfig` array in `web/index.html` (plan: Go-driven) |
| Go schema package path | `awo/web/schema` |

---
