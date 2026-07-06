---
title: "Observability — Section Overview"
id: obs-000-readme
status: accepted
category: GUIDE
stability: STABLE
audience: [module-authors, operators]
since: "1.0"
normative-level: informative
related:
  - "[Observability](observability.md)"
  - "[Configuration](../12-configuration/configuration.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# Observability

**Section 13 | Observability**

Observability covers structured logging, Prometheus metrics, health endpoints, and distributed tracing. All three pillars are instrumented by the framework; modules extend them using the provided APIs.

---

## Contents

| Document | ID | Purpose | Stability |
|---|---|---|---|
| [Observability](observability.md) | OBS-001 | Logging standards, metrics catalog, health endpoints, tracing | STABLE |

---

## Prerequisites

- [Configuration](../12-configuration/configuration.md) — log level, metrics, tracing config
- [Startup Sequence](../03-kernel/startup-sequence.md) — health endpoint behavior at startup
- [Glossary](../GLOSSARY.md) — Structured Logging, Prometheus, Health Check
