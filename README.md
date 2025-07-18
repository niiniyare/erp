# 🌀 Awo ERP System

**Awo** is a modern, modular, and scalable ERP system built to support multi-organization environments with clean architecture and a forward-thinking foundation. Designed to start simple and grow effortlessly, Awo adapts to the evolving needs of businesses — from early-stage teams to enterprise-scale operations.

---

## ✨ Key Highlights

- **Multi-Tenant by Design**  
  Serve multiple organizations securely with built-in tenant isolation and contextual access control.

- **Flexible Organizational Modeling**  
  Each tenant can define its own structure — including departments, projects, regions, and other custom business entities.

- **Modular Feature Architecture**  
  Activate only what each tenant needs. Awo is composed of loosely coupled modules that can be turned on/off per customer.

- **Reliable Business Workflows**  
  Workflow orchestration powered by Temporal ensures stateful, auditable, and resilient automation across tasks like approvals, onboarding, and finance operations.

- **Low-Code Administration UI**  
  Powered by Baidu AMIS, Awo provides a dynamic, schema-driven admin interface for faster configuration and data interaction.

- **Modern API Interfaces**  
  REST APIs for browser apps and gRPC services for mobile clients and integrations, all designed using Goa for consistency and extensibility.

- **Feature Flag Support**  
  Granular control over features per tenant or user segment enables progressive rollout and customer-specific configurations.

- **Built for Growth**  
  Awo’s architecture supports easy extensibility, clean separation of concerns, and maintainable modularity — ideal for solo developers and small teams managing complex systems.

---

## 🧰 Technology Stack

- **Backend**: Go  
- **API Design**: Goa (REST + gRPC)  
- **Workflows**: Temporal.io  
- **Database**: PostgreSQL (with Row-Level Security)  
- **Query Layer**: SQLC  
- **Authentication/Authorization**: Casbin (tentative)  
- **Cache**: Redis  
- **UI Framework**: Baidu AMIS (Low-Code)  
- **Feature Flags**: Built-in (with planned integrations)

---

## 🚀 Project Philosophy

> **“Start humble. Scale securely.”**

Awo is crafted for clarity, developer happiness, and long-term adaptability. It balances minimalism with flexibility, enabling early-stage deployment with the confidence to grow into an enterprise-grade system — all with a single developer or a lean team.

---

## 📦 Use Cases

- Multi-branch retail or franchise management  
- NGOs and project-based operations  
- Corporate holding structures with distinct entities  
- SaaS platforms targeting business verticals  
- Internal business tooling for multi-department workflows

---

## 🔒 Status

This is a **private repository** under active development. Awo is not yet production-ready. Deployment in critical environments should follow a thorough review and testing process.
