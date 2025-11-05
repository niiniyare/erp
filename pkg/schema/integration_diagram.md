# Theme & Token System Integration Architecture

This document provides a comprehensive visual representation of how the Theme and Token systems integrate with the ERP schema engine, showing data flow, dependencies, and runtime interactions.

## System Architecture Diagram

```mermaid
graph TB
    %% ========================================
    %% CORE THEME & TOKEN SUBSYSTEM
    %% ========================================
    subgraph "🎨 Theme & Token Core System"
        direction TB
        
        %% Theme Management Layer
        subgraph "Theme Management"
            TM[ThemeManager<br/>🔧 Registration & Retrieval]
            TC[ThemeCache<br/>⚡ LRU + TTL Cache]
            TR[ThemeRegistry<br/>💾 Storage & Lookup]
            TB[ThemeBuilder<br/>🏗️ Fluent Construction]
        end
        
        %% Token Processing Layer  
        subgraph "Token Processing"
            TKR[TokenRegistry<br/>🎯 Resolution Engine]
            TKS[TokenResolver<br/>🔍 Reference Resolution]
            TKC[CompiledTokenMap<br/>⚡ O(1) Lookups]
            TKV[TokenValidator<br/>✅ Circular Detection]
        end
        
        %% Theme Data Models
        subgraph "Theme Models"
            THM[Theme<br/>📋 Complete Definition]
            DTK[DesignTokens<br/>🎨 Token Hierarchy]
            TOV[ThemeOverrides<br/>🎛️ Customizations]
        end
    end

    %% ========================================
    %% SCHEMA CORE COMPONENTS
    %% ========================================
    subgraph "📋 Schema Core Components"
        direction TB
        
        SCH[Schema<br/>🏠 Core Data Model]
        BLD[Builder<br/>🏗️ Schema Construction]
        FLD[Field<br/>📝 Form Fields]
        ACT[Action<br/>⚡ User Actions]
        LAY[Layout<br/>📐 Visual Structure]
        
        %% Schema Methods
        SCH_APPLY[schema.ApplyTheme()<br/>🎨 Theme Application]
        SCH_BUILD[builder.Build()<br/>🔨 Schema Creation]
    end

    %% ========================================
    %% REGISTRY & STORAGE LAYER
    %% ========================================
    subgraph "🗃️ Registry & Storage Layer"
        direction TB
        
        GREG[GlobalRegistry<br/>🌍 Centralized Access]
        SREG[SchemaRegistry<br/>📚 Schema Storage]
        MREG[MixinRegistry<br/>🧩 Reusable Components]
        
        %% Storage backends
        FSYS[FileSystem<br/>💾 File Storage]
        REDIS[Redis<br/>⚡ Cache Layer] 
        S3[S3<br/>☁️ Cloud Storage]
    end

    %% ========================================
    %% VALIDATION & PROCESSING
    %% ========================================
    subgraph "✅ Validation & Processing"
        direction TB
        
        VAL[Validator<br/>🔍 Schema Validation]
        VTHM[ThemeValidator<br/>🎨 Theme Validation]
        ENR[Enricher<br/>✨ Data Enhancement]
        PRS[Parser<br/>📖 Config Processing]
        
        %% Processing steps
        VAL_FLOW[Validation Flow<br/>schema → theme → tokens]
        ENR_FLOW[Enrichment Flow<br/>context → theme → resolution]
    end

    %% ========================================
    %% RUNTIME & CONTEXT SYSTEM
    %% ========================================
    subgraph "🏃 Runtime & Context System"
        direction TB
        
        RT[Runtime<br/>🚀 Execution Engine]
        CTX[Context<br/>📡 Request Context]
        STATE[RuntimeState<br/>💾 Session State]
        
        %% Context types
        TCTX[ThemeContext<br/>🎨 Active Theme]
        TENCTX[TenantContext<br/>🏢 Multi-tenant]
        USRCTX[UserContext<br/>👤 User Preferences]
    end

    %% ========================================
    %% ENTERPRISE & MULTI-TENANT
    %% ========================================
    subgraph "🏢 Enterprise & Multi-Tenant"
        direction TB
        
        ENT[Enterprise<br/>💼 Advanced Features]
        TTM[TenantThemeManager<br/>🏢 Tenant Themes]
        TTC[TenantConfig<br/>⚙️ Tenant Settings]
        
        %% Multi-tenant flow
        TENANT_FLOW[Tenant Resolution<br/>context → tenant → theme]
        OVERRIDE_FLOW[Override Application<br/>base → tenant → custom]
    end

    %% ========================================
    %% DATA FLOW CONNECTIONS
    %% ========================================
    
    %% Schema Building Flow
    BLD -->|"creates"| SCH
    BLD -->|"applies theme"| TM
    BLD -->|"uses builder"| TB
    SCH_BUILD -->|"triggers"| VAL
    
    %% Theme Application Flow
    SCH -->|"calls ApplyTheme()"| SCH_APPLY
    SCH_APPLY -->|"fetches theme"| TM
    TM -->|"checks cache"| TC
    TC -->|"miss: loads from"| TR
    TM -->|"returns"| THM
    
    %% Token Resolution Flow
    FLD -->|"references tokens"| TKR
    ACT -->|"uses tokens"| TKR
    LAY -->|"applies tokens"| TKR
    TKR -->|"checks compiled"| TKC
    TKC -->|"miss: resolves via"| TKS
    TKS -->|"validates with"| TKV
    
    %% Registry Integration
    TM -->|"registers themes"| TR
    TR -->|"stores in"| GREG
    GREG -->|"persists to"| FSYS
    GREG -->|"caches in"| REDIS
    GREG -->|"backups to"| S3
    
    %% Validation Flow
    VAL -->|"validates schema"| SCH
    VAL -->|"validates theme"| VTHM
    VTHM -->|"checks theme"| THM
    VTHM -->|"validates tokens"| TKV
    TKV -->|"checks references"| DTK
    
    %% Enrichment Flow
    ENR -->|"enriches schema"| SCH
    ENR -->|"resolves theme"| TM
    ENR -->|"resolves tokens"| TKR
    ENR_FLOW -->|"applies to"| FLD
    
    %% Runtime Context Flow
    RT -->|"creates"| CTX
    CTX -->|"contains"| TCTX
    CTX -->|"contains"| TENCTX
    CTX -->|"maintains"| STATE
    STATE -->|"tracks active"| THM
    
    %% Multi-Tenant Flow
    TENCTX -->|"identifies tenant"| TTM
    TTM -->|"loads config"| TTC
    TTM -->|"applies overrides"| TOV
    TENANT_FLOW -->|"resolves to"| THM
    OVERRIDE_FLOW -->|"customizes"| DTK
    
    %% Enterprise Integration
    ENT -->|"enables"| TTM
    ENT -->|"provides"| TTC
    ENT -->|"manages"| TOV
    
    %% Parser Integration
    PRS -->|"parses theme config"| THM
    PRS -->|"extracts tokens"| DTK
    PRS -->|"builds overrides"| TOV

    %% ========================================
    %% STYLING & VISUAL GROUPING
    %% ========================================
    
    %% Core Theme System - Blue
    style TM fill:#3b82f6,stroke:#1e40af,stroke-width:3px,color:#fff
    style TKR fill:#3b82f6,stroke:#1e40af,stroke-width:3px,color:#fff
    style THM fill:#60a5fa,stroke:#2563eb,stroke-width:2px,color:#fff
    style DTK fill:#60a5fa,stroke:#2563eb,stroke-width:2px,color:#fff
    
    %% Schema Core - Green
    style SCH fill:#10b981,stroke:#059669,stroke-width:3px,color:#fff
    style BLD fill:#10b981,stroke:#059669,stroke-width:2px,color:#fff
    style SCH_APPLY fill:#34d399,stroke:#10b981,stroke-width:2px,color:#000
    
    %% Registry Layer - Purple
    style GREG fill:#8b5cf6,stroke:#7c3aed,stroke-width:2px,color:#fff
    style TR fill:#a78bfa,stroke:#8b5cf6,stroke-width:2px,color:#fff
    
    %% Runtime - Orange
    style RT fill:#f59e0b,stroke:#d97706,stroke-width:2px,color:#fff
    style CTX fill:#fbbf24,stroke:#f59e0b,stroke-width:2px,color:#000
    
    %% Enterprise - Red
    style ENT fill:#ef4444,stroke:#dc2626,stroke-width:2px,color:#fff
    style TTM fill:#f87171,stroke:#ef4444,stroke-width:2px,color:#fff
    
    %% Performance Components - Yellow
    style TC fill:#eab308,stroke:#ca8a04,stroke-width:2px,color:#000
    style TKC fill:#eab308,stroke:#ca8a04,stroke-width:2px,color:#000
```

## Integration Flow Explanation

### 1. **Schema Creation & Theme Application**

```
Developer Request → Builder → Schema → ApplyTheme() → ThemeManager → Theme
```

**Process:**
1. Developer uses `Builder` to create schema structure
2. Theme can be applied during building or after creation
3. `schema.ApplyTheme()` triggers theme resolution
4. `ThemeManager` fetches theme from cache or registry
5. Theme tokens are compiled and applied to components

### 2. **Token Resolution Pipeline**

```
Component Token Reference → TokenRegistry → CompiledMap → TokenResolver → Actual Value
```

**Process:**
1. Fields/Actions reference tokens (e.g., `{semantic.colors.primary}`)
2. `TokenRegistry` checks precompiled map for O(1) lookup
3. If not cached, `TokenResolver` resolves reference chain
4. `TokenValidator` ensures no circular references
5. Final value returned and cached for future use

### 3. **Multi-Tenant Theme Resolution**

```
HTTP Request → TenantContext → TenantThemeManager → TenantConfig → Theme + Overrides
```

**Process:**
1. Request middleware extracts tenant ID from context
2. `TenantThemeManager` loads tenant-specific configuration
3. Base theme loaded from registry
4. Tenant overrides applied to create customized theme
5. Resulting theme cached per tenant

### 4. **Runtime Context Propagation**

```
Runtime.Initialize() → Context → ThemeContext + TenantContext → State Management
```

**Process:**
1. Runtime initializes with request context
2. Context contains theme ID and tenant information
3. `RuntimeState` tracks active theme for session
4. Components query state for current theme/tokens

### 5. **Validation & Processing Flow**

```
Schema → Validator → ThemeValidator → TokenValidator → Enricher → Final Schema
```

**Process:**
1. Schema structure validated first
2. Theme existence and validity checked
3. All token references validated (no circular deps)
4. Enricher resolves all tokens and adds metadata
5. Final schema ready for runtime execution

## Key Integration Benefits

### 🔄 **Seamless Integration**
- No breaking changes to existing schema APIs
- Theme application is optional and backward compatible
- Progressive enhancement of existing schemas

### ⚡ **Performance Optimized**
- O(1) token lookups via precompiled maps
- Multi-level caching (theme, token, resolution)
- Lazy loading with intelligent invalidation

### 🏢 **Enterprise Ready**
- Multi-tenant theme isolation
- Tenant-specific overrides and customizations
- Centralized theme management across deployments

### 🛡️ **Robust Validation**
- Circular reference detection
- Type-safe token resolution
- Runtime validation with detailed error messages

### 🎨 **Design System Foundation**
- Consistent styling across all components
- Three-tier token architecture (Primitives → Semantic → Components)
- Runtime theme switching capabilities

This architecture ensures the theme and token systems integrate seamlessly with the existing ERP schema engine while providing enterprise-grade performance, validation, and multi-tenancy support.