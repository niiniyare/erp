---
title: "Documentation Reports — Overview"
id: report-readme
status: accepted
category: OVERVIEW
stability: STABLE
audience: [framework-authors]
since: "1.0"
normative-level: informative
related:
  - "[Documentation Architecture](../documentation-architecture.md)"
  - "[Quality Gates](../quality-gates.md)"
---

# Documentation Reports

**Section: 00-documentation/reports/**

This directory contains quality and coverage reports for the Awo documentation set. Reports are generated at each major version release and after significant documentation additions.

---

## Reports

| Report | Purpose |
|---|---|
| [Documentation Coverage](documentation-coverage.md) | Complete file index, section coverage |
| [Architecture Law Coverage](architecture-law-coverage.md) | Verification that all 20 Laws and 12 Invariants are documented in non-primary locations |
| [Mermaid Diagram Index](mermaid-diagram-index.md) | Index of all diagrams by document and type |

---

## Report Generation Schedule

- **Major release**: All reports regenerated
- **Minor release**: Coverage report updated
- **Patch release**: No report update required unless documentation was added

Reports are living documents — update them when the documentation set changes significantly.
