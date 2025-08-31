# Mermaid.js Diagrams Guide

## Overview

AWO ERP Documentation supports Mermaid.js for creating diagrams, flowcharts, and visualizations directly in Markdown. This guide covers all supported diagram types and provides examples for common use cases.

## Basic Syntax

Mermaid diagrams are created using fenced code blocks with the `mermaid` language identifier:

````markdown
```mermaid
graph TD
    A[Start] --> B[Process]
    B --> C[End]
```
````

## Supported Diagram Types

### 1. Flowcharts

Perfect for showing process flows, decision trees, and system workflows.

#### Basic Flowchart
```mermaid
graph TD
    A[User Login] --> B{Valid Credentials?}
    B -->|Yes| C[Dashboard]
    B -->|No| D[Error Message]
    D --> A
    C --> E[User Actions]
```

#### ERP Process Flow
```mermaid
flowchart LR
    A[Order Created] --> B[Inventory Check]
    B --> C{Stock Available?}
    C -->|Yes| D[Reserve Items]
    C -->|No| E[Backorder]
    D --> F[Generate Invoice]
    E --> G[Notify Customer]
    F --> H[Payment Processing]
    H --> I[Fulfill Order]
```

### 2. Sequence Diagrams

Ideal for showing API interactions, user flows, and system communications.

#### API Authentication Flow
```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant A as Auth API
    participant D as Database
    
    U->>F: Login Request
    F->>A: POST /api/v1/auth/login
    A->>D: Validate Credentials
    D-->>A: User Data
    A-->>F: JWT Token
    F-->>U: Dashboard Access
```

#### ABAC Authorization Flow
```mermaid
sequenceDiagram
    participant App as Application
    participant ABAC as ABAC Engine
    participant DB as Policy DB
    participant Attr as Attribute Store
    
    App->>ABAC: Authorization Request
    ABAC->>Attr: Collect User Attributes
    ABAC->>Attr: Collect Resource Attributes
    ABAC->>DB: Retrieve Policies
    ABAC->>ABAC: Evaluate Policies
    ABAC-->>App: Decision (Allow/Deny)
```

### 3. Class Diagrams

Great for showing object relationships and system architecture.

#### Finance Module Classes
```mermaid
classDiagram
    class Account {
        +String accountCode
        +String accountName
        +AccountType type
        +RootType rootType
        +Decimal currentBalance
        +Boolean isActive
        +validateAccount()
        +updateBalance()
    }
    
    class Transaction {
        +UUID transactionId
        +Date postingDate
        +String description
        +TransactionStatus status
        +List~TransactionEntry~ entries
        +post()
        +reverse()
    }
    
    class TransactionEntry {
        +UUID entryId
        +Account account
        +Decimal debitAmount
        +Decimal creditAmount
        +String description
    }
    
    Account ||--o{ TransactionEntry : "has many"
    Transaction ||--o{ TransactionEntry : "contains"
```

### 4. State Diagrams

Perfect for showing entity states and transitions.

#### Transaction State Machine
```mermaid
stateDiagram-v2
    [*] --> Draft
    Draft --> PendingApproval : submit()
    PendingApproval --> Approved : approve()
    PendingApproval --> Draft : reject()
    Approved --> Posted : post()
    Posted --> Reversed : reverse()
    Draft --> Cancelled : cancel()
    Reversed --> [*]
    Cancelled --> [*]
```

#### User Account States
```mermaid
stateDiagram-v2
    [*] --> Created
    Created --> Active : activate()
    Active --> Suspended : suspend()
    Suspended --> Active : reactivate()
    Active --> Locked : lockAccount()
    Locked --> Active : unlock()
    Active --> Disabled : disable()
    Disabled --> [*]
```

### 5. Entity Relationship Diagrams

Excellent for database schema visualization.

#### Finance Schema
```mermaid
erDiagram
    TENANT ||--o{ ENTITY : "has many"
    ENTITY ||--o{ ACCOUNT : "owns"
    ACCOUNT ||--o{ TRANSACTION_ENTRY : "participates in"
    TRANSACTION ||--o{ TRANSACTION_ENTRY : "contains"
    ACCOUNT_GROUP ||--o{ ACCOUNT : "categorizes"
    
    TENANT {
        uuid id PK
        string name
        string domain
        timestamp created_at
    }
    
    ENTITY {
        uuid id PK
        uuid tenant_id FK
        string entity_name
        string entity_code
    }
    
    ACCOUNT {
        uuid id PK
        uuid tenant_id FK
        uuid entity_id FK
        string account_code
        string account_name
        decimal current_balance
    }
    
    TRANSACTION {
        uuid id PK
        uuid tenant_id FK
        date posting_date
        string description
        string status
    }
    
    TRANSACTION_ENTRY {
        uuid id PK
        uuid transaction_id FK
        uuid account_id FK
        decimal debit_amount
        decimal credit_amount
    }
```

### 6. Gantt Charts

Great for project timelines and implementation phases.

#### ERP Implementation Timeline
```mermaid
gantt
    title AWO ERP Implementation Phases
    dateFormat YYYY-MM-DD
    section Phase 1
    System Setup           :done, setup, 2024-01-01, 2024-02-01
    User Management        :done, users, 2024-01-15, 2024-03-01
    Authentication         :done, auth, 2024-02-01, 2024-03-15
    
    section Phase 2
    Finance Module         :active, finance, 2024-03-01, 2024-05-01
    ABAC Implementation    :abac, 2024-03-15, 2024-04-30
    API Documentation      :done, docs, 2024-04-01, 2024-04-15
    
    section Phase 3
    Reporting Engine       :reports, 2024-05-01, 2024-06-15
    Analytics Dashboard    :analytics, 2024-05-15, 2024-07-01
    Mobile Application     :mobile, 2024-06-01, 2024-08-01
```

### 7. Git Graphs

Useful for showing repository workflows and branching strategies.

#### Git Flow Strategy
```mermaid
gitgraph
    commit id: "Initial"
    branch develop
    checkout develop
    commit id: "Feature A"
    branch feature/user-auth
    checkout feature/user-auth
    commit id: "Auth implementation"
    commit id: "Tests added"
    checkout develop
    merge feature/user-auth
    commit id: "Integration tests"
    checkout main
    merge develop
    commit id: "Release v1.0"
```

### 8. User Journey Maps

Perfect for documenting user experiences and workflows.

#### User Onboarding Journey
```mermaid
journey
    title User Onboarding Experience
    section Registration
      Register Account     : 5: User
      Email Verification   : 3: User
      Profile Setup        : 4: User
    section First Login
      System Login         : 4: User
      Dashboard Tour       : 5: User
      Initial Configuration: 3: User, Admin
    section First Transaction
      Create Account       : 2: User
      Record Transaction   : 3: User
      Generate Report      : 4: User
```

## Advanced Features

### 1. Subgraphs and Clusters

Group related nodes for better organization:

```mermaid
flowchart TB
    subgraph Frontend
        UI[User Interface]
        API[API Client]
    end
    
    subgraph Backend
        AUTH[Auth Service]
        ABAC[ABAC Engine]
        DB[(Database)]
    end
    
    UI --> API
    API --> AUTH
    AUTH --> ABAC
    ABAC --> DB
```

### 2. Styling and Themes

Custom styling for better visual appeal:

```mermaid
flowchart LR
    A[Start]:::startClass --> B[Process]:::processClass
    B --> C[End]:::endClass
    
    classDef startClass fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef processClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef endClass fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
```

### 3. Click Events and Links

Add interactivity to diagrams:

```mermaid
flowchart TD
    A[API Documentation] --> B[Swagger UI]
    A --> C[Postman Collection]
    
    click A "/reference/api/" "Go to API docs"
    click B "/reference/api/swagger-ui/" "Open Swagger UI"
```

## Best Practices

### 1. Diagram Organization
- **Keep it Simple**: Focus on key elements, avoid clutter
- **Consistent Naming**: Use clear, descriptive labels
- **Logical Flow**: Organize from left-to-right or top-to-bottom
- **Color Coding**: Use colors to group related elements

### 2. Code Structure
```markdown
<!-- title: System Architecture Overview -->
```mermaid
flowchart TB
    %% This diagram shows the high-level architecture
    
    subgraph "Client Layer"
        WEB[Web App]
        MOBILE[Mobile App]
    end
    
    subgraph "API Layer"
        REST[REST API]
        AUTH[Authentication]
    end
    
    WEB --> REST
    MOBILE --> REST
    REST --> AUTH
```
```

### 3. Documentation Integration
- **Context**: Provide context before the diagram
- **Explanation**: Explain key elements after the diagram  
- **Cross-references**: Link to related documentation
- **Updates**: Keep diagrams in sync with code changes

## Common Use Cases

### 1. System Architecture
```mermaid
graph TB
    subgraph "Presentation Layer"
        A[Web Frontend]
        B[Mobile App]
        C[Admin Panel]
    end
    
    subgraph "API Gateway"
        D[Load Balancer]
        E[Rate Limiter]
        F[Authentication]
    end
    
    subgraph "Services"
        G[User Service]
        H[Finance Service]
        I[ABAC Service]
    end
    
    subgraph "Data Layer"
        J[(PostgreSQL)]
        K[(Redis Cache)]
        L[File Storage]
    end
    
    A --> D
    B --> D
    C --> D
    D --> E --> F
    F --> G
    F --> H
    F --> I
    G --> J
    H --> J
    I --> J
    G --> K
    H --> K
    I --> L
```

### 2. API Request Flow
```mermaid
sequenceDiagram
    participant C as Client
    participant G as API Gateway
    participant A as Auth Service
    participant S as Business Service
    participant D as Database
    
    C->>G: HTTP Request
    G->>A: Validate JWT
    A-->>G: Token Valid
    G->>S: Forward Request
    S->>D: Query Data
    D-->>S: Result Set
    S-->>G: Response
    G-->>C: HTTP Response
```

### 3. Database Relationships
```mermaid
erDiagram
    USER ||--o{ USER_ROLE : "has"
    ROLE ||--o{ USER_ROLE : "assigned to"
    USER ||--o{ TRANSACTION : "creates"
    TRANSACTION ||--o{ TRANSACTION_ENTRY : "contains"
    ACCOUNT ||--o{ TRANSACTION_ENTRY : "involved in"
    ACCOUNT_GROUP ||--o{ ACCOUNT : "contains"
```

## Troubleshooting

### Common Issues

#### 1. Diagram Not Rendering
- Check syntax errors in the Mermaid code
- Ensure proper fenced code block format
- Verify JavaScript is enabled

#### 2. Layout Problems
- Use subgraphs to organize complex diagrams
- Adjust node spacing with graph configuration
- Consider splitting large diagrams into smaller ones

#### 3. Mobile Display Issues
- Test diagrams on mobile devices
- Use shorter labels for better mobile display
- Consider horizontal layouts for wide screens

### Debug Tips

1. **Syntax Validation**: Use [Mermaid Live Editor](https://mermaid.live/) to test syntax
2. **Browser Console**: Check for JavaScript errors
3. **Progressive Enhancement**: Start simple, add complexity gradually

## Examples Repository

For more examples and templates, check:
- `docs/examples/mermaid-samples.md` (when available)
- [Mermaid.js Official Documentation](https://mermaid.js.org/)
- Architecture diagrams in existing documentation

## Theme Support

Diagrams automatically adapt to the documentation theme:
- **Light Mode**: Clean, professional styling
- **Dark Mode**: Dark background with contrasting colors
- **Print Mode**: Optimized for printing and PDFs

---

**Related**: [Documentation Guide](documentation-guide.md) | **Up**: [Contributing](../contributing/)