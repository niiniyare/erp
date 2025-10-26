# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

**Awo Enterprise Resource Planning System** is a sophisticated, multi-tenant ERP platform built with military-grade security and intelligent automation.

### Key Characteristics
- **Purpose**: Next-generation ERP with military-grade security
- **Architecture**: Clean Architecture with Domain-Driven Design (DDD)
- **Status**: 80% complete financial module, active development
- **Target**: Multi-tenant SaaS with white-label capabilities

### Important Schema References
- Always reference schemas at:
  - `@web/types/components`
  - `@web/components/utils/ui`
  - `@web/styles/css`
- Do not reinvent the wheel - use existing schema implementations

## API Architecture

**Current Implementation**: Hybrid architecture combining **Fiber v2** (HTTP routing/middleware) with **Goa v3** (type generation/OpenAPI).

### Development Notes
- UI implementation is at @web/ for now, so we will use that package

[... rest of the existing content remains unchanged ...]