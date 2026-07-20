> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Awo ERP: Identity & Access Management (IAM) System

Welcome to the documentation for the Awo ERP's  Identity and Access Management (IAM) system. This system is built on a sophisticated, ABAC-centric design that provides enterprise-grade security, flexibility, and performance.

This guide serves as the central starting point for developers, SREs, and technical writers to understand how identity is managed and how authorization decisions are made.

## Getting Started: A Recommended Learning Path

For newcomers to the system, we recommend reading the following documents in order to build a strong foundational understanding.

### Core Concepts & Architecture

1.  **[Identity model and RBAC](./01_identity_model_and_rbac.md)**
    *   **Summary**: This is the essential starting point. This document details the foundational **Person-Employee-User** identity model, which separates personal, employment, and system-access data. It also covers the baseline **Role-Based Access Control (RBAC)** schema, including roles and permissions tables.

2.  **[ABAC Evaluation lifecycle ](./02_abac_evaluation_lifecycle.md)**
    *   **Summary**: Dive deep into the core of our security model. This document explains the end-to-end lifecycle of a permission check, from the moment a user makes a request to the final `Allow` or `Deny` decision. It defines the key players: PDP, PEP, and PIPs.

3.  **[Hybrid access control strategy](./03_hybrid_access_control_strategy.md)**
    *   **Summary**: This document explains the advanced strategy of how we fuse RBAC and ABAC. It covers powerful, enterprise-grade concepts built into the database and services, such as **temporal (time-bound) access**, **dynamic role activation**, and fully-audited **role delegation**.

4.  **[Architecture service integration](./04_architecture_service_integration.md)**
    *   **Summary**: This document provides the high-level software architecture blueprint. It explains how the core Go services (Identity, Access, ABAC) interact as a cohesive system, with ABAC acting as the central decision engine that queries the other services for attributes.

## Detailed API Reference

The following documents provide a detailed, API-first reference for the various microservices that make up the ABAC system. They are invaluable for understanding the specific capabilities and data contracts of each component.

*   **[Policy evaluation](./abac/policy_evaluation.md)**
    *   **Summary**: The core of the system. This API is the Policy Decision Point (PDP) that makes the real-time `Allow`/`Deny` decisions.

*   **[Context attribute](./abac/context_attribute.md)**
    *   **Summary**: This service is a primary Policy Information Point (PIP). It is responsible for managing and providing the real-time subject, resource, and environment attributes needed for evaluation.

*   **[Policy](./abac/policy.md)**
    *   **Summary**: This API allows administrators to manage the lifecycle of security policies, including creation, testing, and impact simulation.

*   **[Attribute Definitions](./abac/attribute_definitions.md)**
    *   **Summary**: Provides the schema and validation rules for all attributes used in the system, ensuring data consistency.

*   **[Audit monitoring ](./abac/audit_monitoring.md)**
    *   **Summary**: The compliance and security hub. This API provides access to  audit logs, security event monitoring, and user behavior analytics.

*   **[Administrative performance](./abac/administrative_performance.md)**
    *   **Summary**: The command and control center for SREs and administrators. This API provides tools for monitoring system health, managing performance, and optimizing the ABAC engine.
