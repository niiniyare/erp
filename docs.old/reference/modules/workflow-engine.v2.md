# Customer Workflow Engine - Complete Implementation Guide

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture & Design Principles](#architecture--design-principles)
3. [Database Schema](#database-schema)
4. [Workflow Definition Language](#workflow-definition-language)
5. [Backend Implementation](#backend-implementation)
6. [Temporal Integration](#temporal-integration)
7. [API Layer](#api-layer)
8. [AMIS Frontend Integration](#amis-frontend-integration)
9. [Pre-built Workflow Templates](#pre-built-workflow-templates)
10. [Service Integration Guide](#service-integration-guide)
11. [Security & Performance](#security--performance)
12. [Testing Strategy](#testing-strategy)
13. [Deployment & Operations](#deployment--operations)

---

## System Overview

### What This System Provides

The Customer Workflow Engine is a **declarative, visual workflow orchestration platform** that enables business users to design and execute complex workflows without writing code. It combines the power of Temporal's durable execution with a user-friendly AMIS-based visual designer.

**Core Capabilities:**

-  **Visual Workflow Designer** - Drag-and-drop interface for building workflows
-  **Durable Execution** - Workflows survive system failures and restarts
-  **Human-in-the-Loop** - Approval tasks with role-based assignment
-  **Entity-Aware Routing** - Leverages organizational hierarchy for approvals
-  **Real-time Monitoring** - Track workflow progress and step executions
-  **Extensible Actions** - Easy integration with internal services and external APIs
-  **Complete Audit Trail** - Every decision and action is recorded
-  **Multi-tenant Isolation** - Row-level security ensures data separation

### Key Design Principles

1. **Declarative Over Imperative**
   - Workflows are defined in JSON, not code
   - Business users can understand and modify workflows
   - Version control for workflow definitions

2. **Durability First**
   - Temporal ensures workflows complete even after failures
   - Automatic retries with exponential backoff
   - Long-running workflows (days/weeks) are first-class citizens

3. **Type Safety Throughout**
   - All workflow definitions are validated before execution
   - Strong typing in Go backend prevents runtime errors
   - JSON Schema validation for workflow structure

4. **Tenant Isolation by Default**
   - PostgreSQL Row-Level Security (RLS) on every table
   - Tenant context automatically set in transactions
   - No cross-tenant data leakage possible

5. **Observable & Debuggable**
   - Comprehensive logging at every step
   - OpenTelemetry tracing integration
   - Step-by-step execution history

6. **Extensible Architecture**
   - Step executor registry pattern
   - Easy to add new step types
   - Plugin-style action definitions

### Use Cases & Examples

**Approval Workflows**
```
Invoice > $10,000 → Manager Approval → Director Approval → CFO Approval → Payment
```

**Onboarding Processes**
```
New Employee → Create Accounts → Assign Equipment → Setup Training → Notify Manager
```

**Document Routing**
```
Contract Draft → Legal Review → Finance Review → Executive Approval → Archive
```

**Data Processing Pipelines**
```
Import CSV → Validate Data → Transform Records → Load to Database → Send Report
```

**Compliance Workflows**
```
Quarterly Audit → Collect Documents → Review Checklist → Approvals → Submit to Regulator
```

---

## Architecture & Design Principles

### System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend Layer                            │
│  ┌─────────────────────┐  ┌──────────────────────────────────┐ │
│  │  AMIS Workflow      │  │  Task Management                 │ │
│  │  Builder            │  │  Interface                       │ │
│  │  (Visual Designer)  │  │  (Approval Queue)               │ │
│  └─────────────────────┘  └──────────────────────────────────┘ │
└────────────────────┬────────────────────────────────────────────┘
                     │ HTTP/REST (JSON)
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                         API Layer (Fiber)                        │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────┐ │
│  │  Workflow API    │  │  Task API        │  │  Admin API   │ │
│  │  - Templates     │  │  - My Tasks      │  │  - Metrics   │ │
│  │  - Instances     │  │  - Complete      │  │  - Health    │ │
│  │  - Start         │  │  - Delegate      │  │  - Config    │ │
│  └──────────────────┘  └──────────────────┘  └──────────────┘ │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Workflow Engine Service                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Engine Core                                             │  │
│  │  - Definition Parser     - Validation Engine            │  │
│  │  - Step Registry         - Expression Evaluator         │  │
│  │  - Action Registry       - Assignment Resolver          │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Step Executors                                          │  │
│  │  ValidationExecutor | ConditionExecutor | UserTaskExecutor  │
│  │  ActionExecutor | NotificationExecutor | LoopExecutor    │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────┬──────────────────┬─────────────────────────┘
                     │                  │
          ┌──────────▼──────────┐      │
          │  Temporal Cluster   │      │
          │  - Workflows        │      │
          │  - Activities       │      │
          │  - Workers          │      │
          └─────────────────────┘      │
                                       │
                     ┌─────────────────▼─────────────────┐
                     │  Database Layer (PostgreSQL)      │
                     │  - Templates                      │
                     │  - Instances                      │
                     │  - Step Executions                │
                     │  - User Tasks                     │
                     │  - Variables                      │
                     └───────────────────────────────────┘
```

### Data Flow Diagrams

#### Workflow Creation Flow

```
User designs workflow in AMIS builder
           ↓
JSON definition generated
           ↓
POST /api/workflows/templates
           ↓
Validation Engine checks:
- Valid JSON structure
- All steps have valid types
- No circular dependencies
- Required fields present
           ↓
Store in workflow_templates table
           ↓
Register triggers (if event-based)
           ↓
Return template ID to user
```

#### Workflow Execution Flow

```
Trigger Event (invoice.created)
           ↓
Workflow Engine receives event
           ↓
Load template from database
           ↓
Create workflow_instances record
           ↓
Start Temporal workflow
    CustomWorkflowExecution(definition, instanceID, input)
           ↓
Temporal Worker picks up workflow
           ↓
Execute steps sequentially:
    For each step:
        ↓
    Create workflow_step_executions record
        ↓
    Execute step via appropriate executor
        ↓
    If UserTask: Create task, wait for signal
    If Action: Call service, record result
    If Condition: Evaluate, choose next step
        ↓
    Update step execution with result
        ↓
    Get next step ID from result
           ↓
All steps complete
           ↓
Update workflow_instances.status = 'completed'
           ↓
Execute completion hooks
```

#### User Task (Approval) Flow

```
Workflow reaches UserTask step
           ↓
Create workflow_user_tasks record
           ↓
Determine assignee:
- By user ID (direct assignment)
- By role (any user with role)
- By role in entity (role in specific org unit)
           ↓
Send notification to assignee(s)
           ↓
Temporal workflow PAUSES (waits for signal)
           ↓
User views task in "My Tasks" UI
           ↓
User clicks "Approve" or "Reject"
           ↓
POST /api/workflows/tasks/{id}/complete
           ↓
Update task status and decision
           ↓
Send Temporal signal: "task-{id}-completed"
           ↓
Temporal workflow RESUMES
           ↓
Decision determines next step:
- Approved → on_approved step
- Rejected → on_rejected step
           ↓
Continue workflow execution
```

### Component Responsibilities

**Frontend (AMIS)**
- Visual workflow designer
- Template management UI
- Task inbox for approvals
- Workflow monitoring dashboard
- Form rendering for data collection

**API Layer (Fiber)**
- Request validation and authentication
- Tenant context injection
- RESTful endpoints
- Rate limiting and throttling
- API documentation (Swagger)

**Workflow Engine Service**
- Template CRUD operations
- Workflow instance management
- Step execution orchestration
- Variable resolution
- Expression evaluation

**Temporal Workers**
- Durable workflow execution
- Activity retry logic
- Long-running workflow support
- Timer/signal management
- State persistence

**Database (PostgreSQL + RLS)**
- Data persistence
- Multi-tenant isolation
- Audit trail storage
- Complex queries for reporting
- Transaction management

---

## Database Schema

### Overview

The database schema is designed with the following principles:

1. **Multi-tenant by Default** - Every table has `tenant_id` with RLS policies
2. **Audit Trail** - Created/updated timestamps on all tables
3. **JSONB for Flexibility** - Dynamic data stored as JSONB
4. **Foreign Key Constraints** - Referential integrity enforced
5. **Performance Indexes** - Strategic indexes on query columns
6. **Soft Deletes** - `deleted_at` for data retention

### Core Tables

#### 1. Workflow Templates

Stores customer-defined workflow blueprints.

```sql
-- =====================================================
-- WORKFLOW TEMPLATES
-- =====================================================
CREATE TABLE workflow_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Basic Information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50),  -- approval, notification, data_processing, integration
    
    -- Trigger Configuration
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'manual',
    trigger_event VARCHAR(100),  -- e.g., invoice.created, user.updated
    trigger_conditions JSONB DEFAULT '{}'::jsonb,
    
    -- Workflow Definition (Core JSON)
    definition JSONB NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- AMIS UI State (for visual builder)
    amis_builder_state JSONB,
    
    -- Status Flags
    is_active BOOLEAN DEFAULT true,
    is_system BOOLEAN DEFAULT false,  -- Pre-built templates
    is_draft BOOLEAN DEFAULT false,
    
    -- Validation
    validation_errors JSONB DEFAULT '[]'::jsonb,
    last_validated_at TIMESTAMPTZ,
    
    -- Usage Statistics
    execution_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    avg_execution_time_ms INTEGER,
    
    -- Access Control
    owner_id UUID NOT NULL REFERENCES users(id),
    allowed_roles TEXT[],
    allowed_entity_ids UUID[],
    
    -- Versioning Support
    parent_template_id UUID REFERENCES workflow_templates(id),
    published_at TIMESTAMPTZ,
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT valid_trigger_type CHECK (
        trigger_type IN ('event', 'manual', 'scheduled', 'webhook', 'api')
    ),
    CONSTRAINT valid_category CHECK (
        category IS NULL OR category IN (
            'approval', 'notification', 'data_processing', 'integration', 'compliance'
        )
    )
);

-- Indexes for Performance
CREATE INDEX idx_workflow_templates_tenant 
    ON workflow_templates(tenant_id);
    
CREATE INDEX idx_workflow_templates_trigger 
    ON workflow_templates(trigger_event) 
    WHERE is_active = true AND deleted_at IS NULL;
    
CREATE INDEX idx_workflow_templates_category 
    ON workflow_templates(category);
    
CREATE INDEX idx_workflow_templates_owner 
    ON workflow_templates(owner_id);

-- Composite index for common query patterns
CREATE INDEX idx_workflow_templates_tenant_active 
    ON workflow_templates(tenant_id, is_active, deleted_at);

COMMENT ON TABLE workflow_templates IS 
    'Customer-defined workflow templates with visual builder state';
COMMENT ON COLUMN workflow_templates.definition IS 
    'Complete workflow definition in JSON format (steps, variables, transitions)';
COMMENT ON COLUMN workflow_templates.amis_builder_state IS 
    'AMIS visual builder state for re-editing the workflow';
COMMENT ON COLUMN workflow_templates.trigger_conditions IS 
    'Additional conditions that must be met for event-triggered workflows';
```

**Field Explanations:**

- `definition`: The core workflow JSON (steps, variables, transitions)
- `amis_builder_state`: Stores the visual designer state for re-editing
- `trigger_type`: How the workflow starts (manual, event, scheduled, etc.)
- `validation_errors`: Stores validation issues found during last check
- `parent_template_id`: Enables template versioning and inheritance
- `allowed_roles/entities`: Restricts who can start this workflow

#### 2. Workflow Instances

Tracks individual workflow executions.

```sql
-- =====================================================
-- WORKFLOW INSTANCES (Executions)
-- =====================================================
CREATE TABLE workflow_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES workflow_templates(id) ON DELETE RESTRICT,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Temporal Integration
    temporal_workflow_id VARCHAR(255) NOT NULL UNIQUE,
    temporal_run_id VARCHAR(255) NOT NULL,
    
    -- Execution Context
    trigger_type VARCHAR(50) NOT NULL,
    triggered_by UUID REFERENCES users(id),  -- User who started it
    trigger_data JSONB,  -- Event data that initiated workflow
    
    -- Current Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    current_step_id VARCHAR(100),
    current_step_name VARCHAR(255),
    
    -- Progress Tracking
    steps_completed INTEGER DEFAULT 0,
    total_steps INTEGER NOT NULL,  -- NOTE: approximate for parallel/loop workflows; actual executed steps may exceed this count
    completion_percentage INTEGER DEFAULT 0,
    
    -- Timing Information
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    execution_time_ms INTEGER,
    
    -- Workflow Data
    context_data JSONB DEFAULT '{}'::jsonb,  -- Accumulates as workflow progresses
    output_data JSONB,  -- Final output when completed
    
    -- Error Handling
    error_message TEXT,
    error_stack TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    
    -- Business Context
    entity_id UUID REFERENCES entities(uuid),  -- Associated org unit
    related_record_type VARCHAR(100),  -- invoice, purchase_order, etc.
    related_record_id UUID,
    
    -- Metadata
    tags TEXT[],  -- For filtering/searching
    priority VARCHAR(20) DEFAULT 'normal',  -- low, normal, high, urgent
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_instance_status CHECK (
        status IN ('pending', 'running', 'paused', 'completed', 'failed', 'cancelled', 'timeout')
    ),
    CONSTRAINT valid_priority CHECK (
        priority IN ('low', 'normal', 'high', 'urgent')
    ),
    CONSTRAINT completion_percentage_range CHECK (
        completion_percentage >= 0 AND completion_percentage <= 100
    )
);

-- Performance Indexes
CREATE INDEX idx_workflow_instances_template 
    ON workflow_instances(template_id);
    
CREATE INDEX idx_workflow_instances_tenant 
    ON workflow_instances(tenant_id);
    
CREATE INDEX idx_workflow_instances_status 
    ON workflow_instances(status) 
    WHERE status IN ('running', 'paused', 'failed');
    
CREATE INDEX idx_workflow_instances_temporal 
    ON workflow_instances(temporal_workflow_id);
    
CREATE INDEX idx_workflow_instances_started 
    ON workflow_instances(started_at DESC);
    
CREATE INDEX idx_workflow_instances_related 
    ON workflow_instances(related_record_type, related_record_id);

-- Composite index for dashboard queries
CREATE INDEX idx_workflow_instances_tenant_status_started 
    ON workflow_instances(tenant_id, status, started_at DESC);

COMMENT ON TABLE workflow_instances IS 
    'Individual workflow executions tracked with Temporal integration';
COMMENT ON COLUMN workflow_instances.context_data IS 
    'Mutable data that accumulates as workflow progresses through steps';
COMMENT ON COLUMN workflow_instances.trigger_data IS 
    'Immutable snapshot of data that started the workflow';
```

**Field Explanations:**

- `temporal_workflow_id`: Links to Temporal's workflow execution
- `context_data`: Mutable data passed between steps (e.g., variables)
- `trigger_data`: Immutable snapshot of original input
- `related_record_type/id`: Links workflow to business entity (invoice, PO, etc.)
- `completion_percentage`: For progress bars in UI

#### 3. Workflow Step Executions

Records each step execution within a workflow.

```sql
-- =====================================================
-- WORKFLOW STEP EXECUTIONS
-- =====================================================
CREATE TABLE workflow_step_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Step Identification
    step_id VARCHAR(100) NOT NULL,
    step_name VARCHAR(255) NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    step_order INTEGER NOT NULL,  -- For UI display ordering
    
    -- Execution Lifecycle
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    
    -- Data Flow
    input_data JSONB,
    output_data JSONB,
    
    -- Error Handling
    error_message TEXT,
    error_details JSONB,  -- Structured error information
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    
    -- User Task Specific Fields
    assigned_to_user_id UUID REFERENCES users(id),
    assigned_to_role VARCHAR(50),
    assigned_to_entity_id UUID REFERENCES entities(uuid),
    completed_by UUID REFERENCES users(id),
    completion_comment TEXT,
    
    -- Performance Metrics
    activity_task_queue VARCHAR(100),  -- Temporal task queue used
    worker_identity VARCHAR(255),  -- Which worker executed this
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_step_status CHECK (
        status IN ('pending', 'running', 'waiting', 'completed', 'failed', 'skipped', 'timeout')
    ),
    CONSTRAINT valid_step_type CHECK (
        step_type IN (
            'validation', 'condition', 'user_task', 'action', 
            'notification', 'parallel', 'loop', 'wait', 'webhook'
        )
    )
);

-- Performance Indexes
CREATE INDEX idx_workflow_step_executions_instance 
    ON workflow_step_executions(instance_id, step_order);
    
CREATE INDEX idx_workflow_step_executions_status 
    ON workflow_step_executions(status);
    
CREATE INDEX idx_workflow_step_executions_assigned 
    ON workflow_step_executions(assigned_to_user_id)
    WHERE assigned_to_user_id IS NOT NULL;
    
CREATE INDEX idx_workflow_step_executions_type 
    ON workflow_step_executions(step_type);

-- Index for analytics queries
CREATE INDEX idx_workflow_step_executions_completed 
    ON workflow_step_executions(completed_at DESC) 
    WHERE status = 'completed';

COMMENT ON TABLE workflow_step_executions IS 
    'Individual step executions within workflow instances with detailed tracking';
COMMENT ON COLUMN workflow_step_executions.step_order IS 
    'Sequence number for ordering steps chronologically in UI';
```

**Field Explanations:**

- `step_order`: Sequential number for displaying steps in chronological order
- `input_data/output_data`: Step's input and produced output
- `assigned_to_*`: For user tasks, tracks who it's assigned to
- `worker_identity`: Which Temporal worker executed this (for debugging)

#### 4. User Tasks (Approvals)

Human-in-the-loop tasks requiring manual action.

```sql
-- =====================================================
-- USER TASKS (Approvals/Reviews)
-- =====================================================
CREATE TABLE workflow_user_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    step_execution_id UUID NOT NULL REFERENCES workflow_step_executions(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Task Display Information
    title VARCHAR(500) NOT NULL,
    description TEXT,
    task_type VARCHAR(50) DEFAULT 'approval',  -- approval, review, input, decision
    priority VARCHAR(20) DEFAULT 'normal',  -- low, normal, high, urgent
    
    -- Assignment Strategy
    assigned_to_user_id UUID REFERENCES users(id),  -- Direct user assignment
    assigned_to_role VARCHAR(50),  -- Role-based assignment
    assigned_to_entity_id UUID REFERENCES entities(uuid),  -- Within specific org unit
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Deadline Management
    due_at TIMESTAMPTZ,
    reminder_sent_at TIMESTAMPTZ,
    escalated_at TIMESTAMPTZ,
    escalated_to UUID REFERENCES users(id),
    escalation_level INTEGER DEFAULT 0,
    
    -- Task Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    completed_at TIMESTAMPTZ,
    completed_by UUID REFERENCES users(id),
    
    -- User Response
    decision VARCHAR(20),  -- approved, rejected, delegated, cancelled
    comment TEXT,
    form_data JSONB,  -- Additional data collected from user
    
    -- UI Configuration
    form_schema JSONB,  -- AMIS form schema for data collection
    action_buttons JSONB,  -- Custom action button configurations
    
    -- Display Context
    context_data JSONB,  -- Data to show in task details
    attachments JSONB,  -- File references
    
    -- Access Control
    can_delegate BOOLEAN DEFAULT true,
    can_reassign BOOLEAN DEFAULT false,
    requires_comment BOOLEAN DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- Required: task status/assignment changes must be auditable

    CONSTRAINT valid_task_status CHECK (
        status IN ('pending', 'in_progress', 'completed', 'cancelled', 'timeout', 'escalated')
    ),
    CONSTRAINT valid_task_decision CHECK (
        decision IS NULL OR decision IN ('approved', 'rejected', 'delegated', 'cancelled', 'skipped')
    ),
    CONSTRAINT valid_task_priority CHECK (
        priority IN ('low', 'normal', 'high', 'urgent')
    ),
    CONSTRAINT valid_task_type CHECK (
        task_type IN ('approval', 'review', 'input', 'decision', 'acknowledgment')
    ),
    CONSTRAINT assignment_required CHECK (
        assigned_to_user_id IS NOT NULL OR 
        assigned_to_role IS NOT NULL
    )
);

-- Performance Indexes
CREATE INDEX idx_workflow_user_tasks_instance 
    ON workflow_user_tasks(instance_id);
    
CREATE INDEX idx_workflow_user_tasks_assigned_user 
    ON workflow_user_tasks(assigned_to_user_id, status)
    WHERE status IN ('pending', 'in_progress');
    
CREATE INDEX idx_workflow_user_tasks_assigned_role 
    ON workflow_user_tasks(assigned_to_role, status)
    WHERE status IN ('pending', 'in_progress');
    
CREATE INDEX idx_workflow_user_tasks_due 
    ON workflow_user_tasks(due_at)
    WHERE status IN ('pending', 'in_progress') AND due_at IS NOT NULL;
    
CREATE INDEX idx_workflow_user_tasks_status 
    ON workflow_user_tasks(status);

-- Composite index for "My Tasks" queries
CREATE INDEX idx_workflow_user_tasks_my_tasks 
    ON workflow_user_tasks(tenant_id, assigned_to_user_id, status, due_at)
    WHERE status IN ('pending', 'in_progress');

COMMENT ON TABLE workflow_user_tasks IS 
    'User tasks requiring manual action within workflows (approvals, reviews, decisions)';
COMMENT ON COLUMN workflow_user_tasks.form_schema IS 
    'AMIS form schema for collecting additional data from the user';
COMMENT ON COLUMN workflow_user_tasks.context_data IS 
    'Contextual data to display to help user make decision';
```

**Field Explanations:**

- `form_schema`: AMIS form definition for collecting extra data
- `action_buttons`: Custom button configs (beyond approve/reject)
- `context_data`: Information shown to user for decision-making
- `escalation_level`: Tracks how many times task was escalated
- `can_delegate/reassign`: Controls whether user can transfer task

#### 5. Supporting Tables

**Workflow Triggers** - Event and schedule registration

```sql
-- =====================================================
-- WORKFLOW TRIGGERS
-- =====================================================
CREATE TABLE workflow_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES workflow_templates(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Trigger Configuration
    trigger_type VARCHAR(50) NOT NULL,
    event_name VARCHAR(100),  -- For event triggers
    schedule_expression VARCHAR(100),  -- For scheduled triggers (cron)
    webhook_path VARCHAR(255),  -- For webhook triggers
    webhook_secret VARCHAR(255),  -- Store as HMAC-SHA256 hash (never plaintext); verify incoming requests with constant-time compare
    webhook_method VARCHAR(10) DEFAULT 'POST',  -- HTTP method
    
    -- Filtering Conditions
    conditions JSONB DEFAULT '[]'::jsonb,  -- JSONPath conditions
    
    -- Status & Metrics
    is_active BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMPTZ,
    trigger_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    
    -- Configuration
    max_concurrent_instances INTEGER,  -- Limit concurrent executions
    cooldown_period INTERVAL,  -- Minimum time between triggers
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_trigger_type CHECK (
        trigger_type IN ('event', 'scheduled', 'webhook', 'manual')
    ),
    CONSTRAINT event_trigger_has_name CHECK (
        trigger_type != 'event' OR event_name IS NOT NULL
    ),
    CONSTRAINT scheduled_trigger_has_expression CHECK (
        trigger_type != 'scheduled' OR schedule_expression IS NOT NULL
    ),
    CONSTRAINT webhook_trigger_has_path CHECK (
        trigger_type != 'webhook' OR webhook_path IS NOT NULL
    )
);

CREATE INDEX idx_workflow_triggers_event 
    ON workflow_triggers(event_name, is_active)
    WHERE trigger_type = 'event';
    
CREATE INDEX idx_workflow_triggers_template 
    ON workflow_triggers(template_id);

CREATE UNIQUE INDEX idx_workflow_triggers_webhook_path 
    ON workflow_triggers(tenant_id, webhook_path)
    WHERE trigger_type = 'webhook';

COMMENT ON TABLE workflow_triggers IS 
    'Defines when and how workflows should be automatically triggered';
```

**Workflow Variables** - Shared state across steps

```sql
-- =====================================================
-- WORKFLOW VARIABLES
-- =====================================================
CREATE TABLE workflow_variables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Variable Definition
    key VARCHAR(100) NOT NULL,
    value JSONB NOT NULL,
    data_type VARCHAR(20) NOT NULL,  -- string, number, boolean, object, array
    
    -- Metadata
    created_by_step VARCHAR(100),
    updated_by_step VARCHAR(100),
    is_output BOOLEAN DEFAULT false,  -- Include in workflow output
    
    -- History Tracking
    previous_value JSONB,
    updated_count INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(instance_id, key),
    
    CONSTRAINT valid_data_type CHECK (
        data_type IN ('string', 'number', 'boolean', 'object', 'array', 'null')
    )
);

CREATE INDEX idx_workflow_variables_instance 
    ON workflow_variables(instance_id);

CREATE INDEX idx_workflow_variables_instance_key 
    ON workflow_variables(instance_id, key);

COMMENT ON TABLE workflow_variables IS 
    'Named variables that persist across workflow steps, like a shared state store';
```

**Action Definitions** - Reusable actions

```sql
-- =====================================================
-- ACTION DEFINITIONS (Reusable Actions)
-- =====================================================
CREATE TABLE workflow_action_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,  -- NULL for system actions
    
    -- Action Information
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50),  -- data, notification, integration, transformation
    icon VARCHAR(50),  -- Icon for UI display
    
    -- Action Type & Configuration
    action_type VARCHAR(50) NOT NULL,  -- service_call, http_request, email, database, custom
    configuration_schema JSONB NOT NULL,  -- JSON schema for parameters
    
    -- Service Integration (for internal services)
    service_name VARCHAR(100),
    service_method VARCHAR(100),
    
    -- HTTP Integration (for external APIs)
    http_endpoint VARCHAR(500),
    http_method VARCHAR(10),  -- GET, POST, PUT, DELETE, PATCH
    http_headers JSONB,
    http_auth_type VARCHAR(50),  -- none, basic, bearer, api_key
    
    -- Response Handling
    success_criteria JSONB,  -- How to determine success
    response_mapping JSONB,  -- How to map response to variables
    
    -- Retry Configuration
    default_retry_policy JSONB,
    default_timeout_seconds INTEGER DEFAULT 30,
    
    -- Access Control
    is_system BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    minimum_role VARCHAR(50),
    allowed_entities UUID[],
    
    -- Usage Tracking
    execution_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    avg_duration_ms INTEGER,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_action_name UNIQUE (tenant_id, name),
    CONSTRAINT valid_action_type CHECK (
        action_type IN (
            'service_call', 'http_request', 'email', 'sms', 
            'database', 'custom', 'notification'
        )
    ),
    CONSTRAINT valid_http_method CHECK (
        http_method IS NULL OR http_method IN ('GET', 'POST', 'PUT', 'DELETE', 'PATCH')
    )
);

CREATE INDEX idx_workflow_action_definitions_category 
    ON workflow_action_definitions(category);

CREATE INDEX idx_workflow_action_definitions_tenant 
    ON workflow_action_definitions(tenant_id, is_active);

COMMENT ON TABLE workflow_action_definitions IS 
    'Reusable action definitions that can be used in workflow steps';
COMMENT ON COLUMN workflow_action_definitions.configuration_schema IS 
    'JSON Schema defining required and optional parameters for this action';
```

### Database Functions & Triggers

**Automatic Progress Tracking**

```sql
-- =====================================================
-- AUTOMATIC PROGRESS TRACKING
-- =====================================================

-- Function: Update workflow instance progress
CREATE OR REPLACE FUNCTION update_workflow_instance_progress()
RETURNS TRIGGER AS $$
DECLARE
    v_total_steps INTEGER;
    v_completed_steps INTEGER;
BEGIN
    -- Get total steps for this instance
    SELECT total_steps INTO v_total_steps
    FROM workflow_instances
    WHERE id = NEW.instance_id;
    
    -- Count completed steps
    SELECT COUNT(*) INTO v_completed_steps
    FROM workflow_step_executions
    WHERE instance_id = NEW.instance_id
      AND status = 'completed';
    
    -- Update instance progress
    UPDATE workflow_instances SET
        steps_completed = v_completed_steps,
        completion_percentage = CASE 
            WHEN v_total_steps > 0 
            THEN CAST(v_completed_steps * 100.0 / v_total_steps AS INTEGER)
            ELSE 0
        END,
        current_step_id = CASE 
            WHEN NEW.status = 'running' THEN NEW.step_id
            ELSE current_step_id
        END,
        current_step_name = CASE 
            WHEN NEW.status = 'running' THEN NEW.step_name
            ELSE current_step_name
        END,
        updated_at = NOW()
    WHERE id = NEW.instance_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_instance_progress
    AFTER INSERT OR UPDATE OF status ON workflow_step_executions
    FOR EACH ROW
    EXECUTE FUNCTION update_workflow_instance_progress();

COMMENT ON FUNCTION update_workflow_instance_progress() IS 
    'Automatically maintains workflow instance progress metrics when steps complete';
```

**Template Statistics**

```sql
-- Function: Update template statistics
CREATE OR REPLACE FUNCTION update_workflow_template_statistics()
RETURNS TRIGGER AS $$
BEGIN
    -- Only update on status change to completed or failed
    IF NEW.status IN ('completed', 'failed') AND 
       OLD.status = 'running' THEN
        
        UPDATE workflow_templates SET
            execution_count = execution_count + 1,
            success_count = CASE 
                WHEN NEW.status = 'completed' 
                THEN success_count + 1 
                ELSE success_count 
            END,
            failure_count = CASE 
                WHEN NEW.status = 'failed' 
                THEN failure_count + 1 
                ELSE failure_count 
            END,
            avg_execution_time_ms = CASE
                WHEN avg_execution_time_ms IS NULL 
                THEN NEW.execution_time_ms
                WHEN NEW.execution_time_ms IS NOT NULL
                THEN CAST(
                    (avg_execution_time_ms * execution_count + NEW.execution_time_ms) 
                    / (execution_count + 1) AS INTEGER
                )
                ELSE avg_execution_time_ms
            END
        WHERE id = NEW.template_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_template_statistics
    AFTER UPDATE OF status ON workflow_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_workflow_template_statistics();
```

**Automatic Timestamps**

```sql
-- Function: Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables with updated_at
CREATE TRIGGER trigger_workflow_templates_updated_at
    BEFORE UPDATE ON workflow_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_workflow_instances_updated_at
    BEFORE UPDATE ON workflow_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_workflow_triggers_updated_at
    BEFORE UPDATE ON workflow_triggers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_workflow_variables_updated_at
    BEFORE UPDATE ON workflow_variables
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_workflow_action_definitions_updated_at
    BEFORE UPDATE ON workflow_action_definitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_workflow_user_tasks_updated_at
    BEFORE UPDATE ON workflow_user_tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Row Level Security (RLS)

```sql
-- =====================================================
-- ROW LEVEL SECURITY
-- =====================================================

-- Enable RLS on all workflow tables
ALTER TABLE workflow_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_instances ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_step_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_user_tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_triggers ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_variables ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_action_definitions ENABLE ROW LEVEL SECURITY;

-- Helper function to get current tenant ID from session
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Tenant isolation policies for application role
CREATE POLICY tenant_isolation_policy ON workflow_templates
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_policy ON workflow_instances
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_policy ON workflow_step_executions
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_policy ON workflow_user_tasks
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_policy ON workflow_triggers
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_policy ON workflow_variables
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

-- Action definitions: tenant-specific OR system-wide
CREATE POLICY action_definition_access_policy ON workflow_action_definitions
    FOR SELECT TO application_role
    USING (
        tenant_id IS NULL  -- System actions (available to all)
        OR tenant_id = current_tenant_id()  -- Tenant-specific actions
    );

CREATE POLICY action_definition_modify_policy ON workflow_action_definitions
    FOR INSERT, UPDATE, DELETE TO application_role
    USING (tenant_id = current_tenant_id());  -- Can only modify own actions

-- Admin full access (for internal tooling)
CREATE POLICY admin_full_access_templates ON workflow_templates
    FOR ALL TO admin_role USING (TRUE);

CREATE POLICY admin_full_access_instances ON workflow_instances
    FOR ALL TO admin_role USING (TRUE);

CREATE POLICY admin_full_access_step_executions ON workflow_step_executions
    FOR ALL TO admin_role USING (TRUE);

CREATE POLICY admin_full_access_user_tasks ON workflow_user_tasks
    FOR ALL TO admin_role USING (TRUE);
```

### Permissions

```sql
-- =====================================================
-- GRANT PERMISSIONS
-- =====================================================

-- Application role (used by API layer)
GRANT SELECT, INSERT, UPDATE, DELETE ON workflow_templates TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON workflow_instances TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON workflow_step_executions TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON workflow_user_tasks TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON workflow_triggers TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON workflow_variables TO application_role;
GRANT SELECT ON workflow_action_definitions TO application_role;
GRANT INSERT, UPDATE, DELETE ON workflow_action_definitions TO application_role;

-- Admin role (for internal operations)
GRANT ALL ON ALL TABLES IN SCHEMA public TO admin_role;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO admin_role;

-- Temporal worker role (for workflow execution)
GRANT SELECT, INSERT, UPDATE ON workflow_instances TO temporal_worker_role;
GRANT SELECT, INSERT, UPDATE ON workflow_step_executions TO temporal_worker_role;
GRANT SELECT, INSERT, UPDATE ON workflow_user_tasks TO temporal_worker_role;
GRANT SELECT, INSERT, UPDATE ON workflow_variables TO temporal_worker_role;
GRANT SELECT ON workflow_templates TO temporal_worker_role;
GRANT SELECT ON workflow_action_definitions TO temporal_worker_role;
```

---

## Workflow Definition Language

### Overview

Workflows are defined using a **declarative JSON structure** that describes the business logic without requiring code. This JSON is stored in the `workflow_templates.definition` column and executed by the Temporal workflow engine.

### Core Structure

```json
{
  "name": "Workflow Name",
  "version": 1,
  "description": "What this workflow does",
  "variables": [...],  // Input and shared variables
  "steps": [...]       // Execution steps
}
```

### Complete Example: Invoice Approval

```json
{
  "name": "Invoice Approval Workflow",
  "version": 1,
  "description": "Multi-level approval for invoices based on amount thresholds",
  
  "variables": [
    {
      "name": "invoice",
      "type": "object",
      "required": true,
      "description": "The invoice being approved"
    },
    {
      "name": "approval_level",
      "type": "integer",
      "default": 0,
      "description": "Current approval level"
    },
    {
      "name": "approval_chain",
      "type": "array",
      "default": [],
      "description": "History of approvals"
    }
  ],
  
  "steps": [
    {
      "id": "validate_invoice",
      "type": "validation",
      "name": "Validate Invoice Data",
      "description": "Ensure invoice has all required fields",
      "config": {
        "rules": [
          {
            "field": "$.invoice.amount",
            "operator": ">",
            "value": 0,
            "error_message": "Invoice amount must be greater than zero"
          },
          {
            "field": "$.invoice.vendor_id",
            "operator": "exists",
            "error_message": "Vendor ID is required"
          },
          {
            "field": "$.invoice.entity_id",
            "operator": "exists",
            "error_message": "Entity ID is required"
          }
        ],
        "continue_on_error": false
      },
      "on_success": "check_amount_threshold",
      "on_failure": "notify_creator_error"
    },
    
    {
      "id": "check_amount_threshold",
      "type": "condition",
      "name": "Determine Approval Level",
      "description": "Route to appropriate approver based on amount",
      "config": {
        "condition": "$.invoice.amount > 50000",
        "expression_type": "simple"
      },
      "on_true": "require_cfo_approval",
      "on_false": "check_director_threshold"
    },
    
    {
      "id": "check_director_threshold",
      "type": "condition",
      "name": "Check Director Threshold",
      "config": {
        "condition": "$.invoice.amount > 10000"
      },
      "on_true": "require_director_approval",
      "on_false": "require_manager_approval"
    },
    
    {
      "id": "require_manager_approval",
      "type": "user_task",
      "name": "Manager Approval",
      "description": "Manager must approve this invoice",
      "config": {
        "title": "Approve Invoice ${invoice.invoice_number}",
        "description": "Invoice from ${invoice.vendor_name} for $${invoice.amount}",
        "priority": "normal",
        
        "assign_to_role": "manager",
        "entity_scope": "$.invoice.entity_id",
        
        "timeout": "48h",
        "timeout_action": "escalate",
        "escalate_to_role": "director",
        
        "form_fields": [
          {
            "name": "comment",
            "type": "textarea",
            "label": "Approval Comment",
            "placeholder": "Add your comments...",
            "required": false
          }
        ],
        
        "actions": [
          {
            "id": "approve",
            "label": "Approve",
            "style": "primary",
            "icon": "fa fa-check"
          },
          {
            "id": "reject",
            "label": "Reject",
            "style": "danger",
            "icon": "fa fa-times"
          },
          {
            "id": "delegate",
            "label": "Delegate",
            "style": "default",
            "icon": "fa fa-share"
          }
        ]
      },
      "on_approved": "create_payment",
      "on_rejected": "notify_rejection",
      "on_timeout": "escalate_to_director"
    },
    
    {
      "id": "require_director_approval",
      "type": "user_task",
      "name": "Director Approval",
      "config": {
        "title": "High-Value Invoice Approval",
        "description": "Invoice requires director approval due to amount",
        "priority": "high",
        "assign_to_role": "director",
        "entity_scope": "$.invoice.entity_id",
        "timeout": "72h",
        "timeout_action": "escalate",
        "escalate_to_role": "cfo"
      },
      "on_approved": "create_payment",
      "on_rejected": "notify_rejection"
    },
    
    {
      "id": "require_cfo_approval",
      "type": "user_task",
      "name": "CFO Approval",
      "config": {
        "title": "Critical Invoice Approval Required",
        "description": "High-value invoice requires CFO approval",
        "priority": "urgent",
        "assign_to_role": "cfo",
        "timeout": "96h",
        "timeout_action": "auto_reject"
      },
      "on_approved": "create_payment",
      "on_rejected": "notify_rejection"
    },
    
    {
      "id": "create_payment",
      "type": "action",
      "name": "Create Payment Record",
      "description": "Generate payment record in system",
      "config": {
        "action": "create_payment",
        "service": "payment_service",
        "method": "CreatePayment",
        "params": {
          "invoice_id": "$.invoice.id",
          "amount": "$.invoice.amount",
          "vendor_id": "$.invoice.vendor_id",
          "approved_by": "$.completed_by",
          "approval_date": "${NOW()}"
        },
        "retry": {
          "max_attempts": 3,
          "initial_interval": "1s",
          "backoff_coefficient": 2.0,
          "maximum_interval": "30s"
        }
      },
      "on_success": "notify_approval_complete",
      "on_failure": "notify_payment_error"
    },
    
    {
      "id": "notify_approval_complete",
      "type": "notification",
      "name": "Send Approval Notification",
      "config": {
        "template": "invoice_approved",
        "channel": "email",
        "recipients": [
          "$.invoice.created_by_email",
          "$.completed_by_email"
        ],
        "data": {
          "invoice_number": "$.invoice.invoice_number",
          "amount": "$.invoice.amount",
          "approved_by_name": "$.completed_by_name",
          "approval_comment": "$.comment"
        }
      },
      "on_success": "end"
    },
    
    {
      "id": "notify_rejection",
      "type": "notification",
      "name": "Send Rejection Notification",
      "config": {
        "template": "invoice_rejected",
        "channel": "email",
        "recipients": ["$.invoice.created_by_email"],
        "data": {
          "invoice_number": "$.invoice.invoice_number",
          "rejected_by_name": "$.completed_by_name",
          "rejection_reason": "$.comment"
        }
      },
      "on_success": "end"
    },
    
    {
      "id": "notify_creator_error",
      "type": "notification",
      "name": "Notify Validation Error",
      "config": {
        "template": "invoice_validation_error",
        "channel": "email",
        "recipients": ["$.invoice.created_by_email"],
        "data": {
          "invoice_number": "$.invoice.invoice_number",
          "errors": "$.validation_errors"
        }
      },
      "on_success": "end"
    },
    
    {
      "id": "notify_payment_error",
      "type": "notification",
      "name": "Notify Payment Creation Error",
      "config": {
        "template": "payment_creation_error",
        "channel": "email",
        "recipients": ["finance@company.com"],
        "data": {
          "invoice_number": "$.invoice.invoice_number",
          "error": "$.error_message"
        }
      },
      "on_success": "end"
    }
  ]
}
```

### Step Types Reference

#### 1. Validation Step

Validates data before allowing workflow to proceed.

```json
{
  "type": "validation",
  "config": {
    "rules": [
      {
        "field": "$.data.field_name",  // JSONPath to field
        "operator": ">",  // Comparison operator
        "value": 0,  // Expected value
        "error_message": "Field must be greater than 0"
      }
    ],
    "continue_on_error": false,  // Stop workflow on validation failure
    "collect_all_errors": true   // Collect all errors before failing
  }
}
```

**Available Operators:**
- `=` - Equals
- `!=` - Not equals
- `>` - Greater than
- `<` - Less than
- `>=` - Greater than or equal
- `<=` - Less than or equal
- `in` - Value in list
- `not_in` - Value not in list
- `exists` - Field exists and is not null
- `not_exists` - Field does not exist or is null
- `regex` - Matches regular expression
- `contains` - String contains substring
- `starts_with` - String starts with prefix
- `ends_with` - String ends with suffix

**Example with Multiple Rules:**

```json
{
  "id": "validate_user_registration",
  "type": "validation",
  "name": "Validate User Registration",
  "config": {
    "rules": [
      {
        "field": "$.user.email",
        "operator": "regex",
        "value": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
        "error_message": "Invalid email format"
      },
      {
        "field": "$.user.age",
        "operator": ">=",
        "value": 18,
        "error_message": "User must be at least 18 years old"
      },
      {
        "field": "$.user.country",
        "operator": "in",
        "value": ["US", "CA", "UK", "AU"],
        "error_message": "Service not available in this country"
      }
    ],
    "continue_on_error": false,
    "collect_all_errors": true
  },
  "on_success": "next_step",
  "on_failure": "send_validation_error_email"
}
```

#### 2. Condition Step

Branches workflow execution based on conditional logic.

**Implementation:** condition steps are evaluated using `pkg/condition/builder.go` which wraps `expr-lang` (compiled, type-checked, sandboxed — no arbitrary code execution). JSONPath `$.` field paths in workflow definitions are stripped of the leading `$.` by the expression adapter (`engine/expression.go`) before being passed to the condition builder's `GetValue`, which uses dot-notation internally.

**There is NO `javascript` expression type.** Use the structured `ConditionGroup`/`ConditionRule` format, or the `if` formula field (expr-lang) for power users.

**Standard binary condition (single rule):**

```json
{
  "type": "condition",
  "config": {
    "rule": {
      "id": "amount_check",
      "left": { "type": "field", "field": "invoice.amount" },
      "op": "greater",
      "right": 10000
    }
  },
  "on_true": "high_value_path",
  "on_false": "standard_path"
}
```

**Available operators** (from `pkg/condition/builder.go`):
- `equal`, `not_equal`
- `less`, `less_or_equal`, `greater`, `greater_or_equal`
- `between`, `not_between`
- `is_empty`, `is_not_empty`
- `contains`, `not_contains`, `starts_with`, `ends_with`
- `select_any_in`, `select_not_any_in`
- `match_regexp` (bounded timeout, ReDoS-safe via `regexp2`)

**Compound condition (ConditionGroup — AND/OR/NOT):**

```json
{
  "type": "condition",
  "config": {
    "group": {
      "id": "g1",
      "conjunction": "and",
      "children": [
        {
          "id": "r1",
          "left": { "type": "field", "field": "invoice.amount" },
          "op": "greater",
          "right": 10000
        },
        {
          "id": "r2",
          "left": { "type": "field", "field": "entity.type" },
          "op": "equal",
          "right": "COMPANY"
        }
      ]
    }
  },
  "on_true": "director_approval",
  "on_false": "manager_approval"
}
```

**Power-user formula** (expr-lang, compiled and type-checked at template validation time):

```json
{
  "type": "condition",
  "config": {
    "formula": "invoice.amount > 10000 && entity.type == 'COMPANY'"
  },
  "on_true": "director_approval",
  "on_false": "manager_approval"
}
```

> Note: field references in `formula` use dot-notation directly (no `$.` prefix). The formula is compiled once at validation and cached — runtime execution is fast.

**Multi-way branching (evaluated top-to-bottom, first match wins):**

```json
{
  "id": "determine_approval_path",
  "type": "condition",
  "name": "Determine Approval Path",
  "config": {
    "cases": [
      {
        "formula": "invoice.amount < 1000",
        "next_step": "auto_approve"
      },
      {
        "formula": "invoice.amount >= 1000 && invoice.amount < 10000",
        "next_step": "manager_approval"
      },
      {
        "formula": "invoice.amount >= 10000 && invoice.amount < 50000",
        "next_step": "director_approval"
      },
      {
        "formula": "invoice.amount >= 50000",
        "next_step": "cfo_approval"
      }
    ],
    "default": "error_handler"
  }
}
```

#### 3. User Task Step (Approvals)

Creates a task for human review/approval.

```json
{
  "type": "user_task",
  "config": {
    // Display Information
    "title": "Approve Invoice ${invoice_number}",
    "description": "Invoice from ${vendor_name} for $${amount}",
    "priority": "normal",  // low, normal, high, urgent
    
    // Assignment (choose one strategy)
    "assign_to_user_id": "uuid",  // Direct user assignment
    // OR
    "assign_to_role": "manager",  // Role-based
    // OR
    "assign_to_role_in_entity": {
      "role": "manager",
      "entity": "$.invoice.entity_id"  // Role within specific entity
    },
    
    // Deadline & Escalation
    "timeout": "48h",  // Duration before timeout
    "timeout_action": "escalate",  // escalate, auto_approve, auto_reject, cancel
    "escalate_to_role": "director",
    "send_reminder_at": "24h",  // Send reminder before due
    
    // Form for Data Collection
    "form_fields": [
      {
        "name": "comment",
        "type": "textarea",
        "label": "Comment",
        "placeholder": "Add your comments...",
        "required": false
      },
      {
        "name": "risk_level",
        "type": "select",
        "label": "Risk Assessment",
        "options": [
          {"label": "Low", "value": "low"},
          {"label": "Medium", "value": "medium"},
          {"label": "High", "value": "high"}
        ],
        "required": true
      },
      {
        "name": "follow_up_date",
        "type": "date",
        "label": "Follow-up Date",
        "required": false
      }
    ],
    
    // Custom Actions
    "actions": [
      {"id": "approve", "label": "Approve", "style": "primary"},
      {"id": "reject", "label": "Reject", "style": "danger"},
      {"id": "delegate", "label": "Delegate", "style": "default"},
      {"id": "request_info", "label": "Request More Info", "style": "info"}
    ],
    
    // Permissions
    "can_delegate": true,
    "can_reassign": false,
    "requires_comment": false
  },
  "on_approved": "next_step",
  "on_rejected": "rejection_handler",
  "on_delegated": "delegate_handler",
  "on_request_info": "info_request_handler"
}
```

**Assignment Strategies Explained:**

1. **Direct User Assignment** - Task assigned to specific user
```json
"assign_to_user_id": "123e4567-e89b-12d3-a456-426614174000"
```

2. **Role-Based** - Task goes to any user with the role (first available)
```json
"assign_to_role": "manager"
```

3. **Role in Entity** - Task assigned to user with role in specific org unit
```json
"assign_to_role_in_entity": {
  "role": "manager",
  "entity": "$.invoice.entity_id"  // Path to entity ID in context
}
```

#### 4. Action Step (Service Calls)

Executes a pre-defined action or service call.

```json
{
  "type": "action",
  "config": {
    "action": "create_payment",  // Action definition name
    "service": "payment_service",
    "method": "CreatePayment",
    
    // Parameters (JSONPath values resolved at runtime)
    "params": {
      "invoice_id": "$.invoice.id",
      "amount": "$.invoice.amount",
      "vendor_id": "$.invoice.vendor_id",
      "approved_by": "$.completed_by",
      "static_field": "some_value"  // Static values also supported
    },
    
    // Retry Configuration
    "retry": {
      "max_attempts": 3,
      "initial_interval": "1s",
      "backoff_coefficient": 2.0,
      "maximum_interval": "30s"
    },
    
    // Timeout
    "timeout": "30s",
    
    // Result Mapping
    "output_mapping": {
      "payment_id": "$.result.id",
      "payment_status": "$.result.status"
    }
  },
  "on_success": "next_step",
  "on_failure": "error_handler"
}
```

**HTTP Request Action:**

```json
{
  "type": "action",
  "config": {
    "action_type": "http_request",
    "http_method": "POST",
    "http_endpoint": "https://api.vendor.com/webhooks/invoice",
    "http_headers": {
      "Content-Type": "application/json",
      "Authorization": "Bearer ${secrets.vendor_api_key}"
    },
    "http_body": {
      "invoice_id": "$.invoice.id",
      "amount": "$.invoice.amount",
      "status": "approved"
    },
    "success_criteria": {
      "status_code": [200, 201],
      "body_contains": "success"
    }
  }
}
```

#### 5. Notification Step

Sends notifications via email, SMS, push, or webhook.

```json
{
  "type": "notification",
  "config": {
    "template": "invoice_approved",  // Template name
    "channel": "email",  // email, sms, push, webhook
    
    // Recipients (can be static or JSONPath)
    "recipients": [
      "$.invoice.created_by_email",
      "finance@company.com",
      "$.manager_email"
    ],
    
    // Template Data
    "data": {
      "invoice_number": "$.invoice.invoice_number",
      "amount": "$.invoice.amount",
      "approved_by": "$.completed_by_name",
      "comment": "$.comment"
    },
    
    // Optional Configuration
    "priority": "normal",
    "send_immediately": true,
    "attachments": ["$.invoice.pdf_url"]
  },
  "on_success": "next_step",
  "on_failure": "log_notification_failure"
}
```

**Multi-channel Notification:**

```json
{
  "id": "notify_urgent_approval",
  "type": "notification",
  "name": "Send Urgent Approval Notification",
  "config": {
    "channels": [
      {
        "type": "email",
        "template": "urgent_approval_email",
        "recipients": ["$.approver.email"]
      },
      {
        "type": "sms",
        "template": "urgent_approval_sms",
        "recipients": ["$.approver.phone"]
      },
      {
        "type": "push",
        "recipients": ["$.approver.user_id"],
        "title": "Urgent Approval Required",
        "body": "Invoice ${invoice_number} needs approval"
      }
    ],
    "data": {
      "invoice_number": "$.invoice.invoice_number",
      "amount": "$.invoice.amount"
    }
  }
}
```

#### 6. Parallel Step

Executes multiple branches concurrently.

```json
{
  "type": "parallel",
  "config": {
    "branches": [
      {
        "id": "branch_1",
        "name": "Credit Check",
        "steps": ["check_credit", "validate_score"]
      },
      {
        "id": "branch_2",
        "name": "Background Check",
        "steps": ["run_background_check", "verify_employment"]
      },
      {
        "id": "branch_3",
        "name": "Reference Check",
        "steps": ["contact_references", "collect_feedback"]
      }
    ],
    "wait_for": "all",  // all, any, none
    "timeout": "24h",
    "continue_on_branch_failure": false
  },
  "on_complete": "merge_results",
  "on_timeout": "timeout_handler"
}
```

**Wait Strategies:**

- `all` - Wait for all branches to complete (default)
- `any` - Continue when any branch completes (race condition)
- `none` - Start all branches and continue immediately (fire-and-forget)

#### 7. Loop Step

Iterates over a collection, executing steps for each item.

```json
{
  "type": "loop",
  "config": {
    "collection": "$.invoice.line_items",  // Array to iterate over
    "iterator_name": "item",  // Variable name for current item
    
    // Steps to execute for each item
    "steps": ["validate_item", "calculate_tax", "update_inventory"],
    
    // Loop Control
    "max_iterations": 1000,  // Safety limit
    "continue_on_error": false,  // Stop on first error
    "parallel": false,  // Execute serially (true for parallel)
    "batch_size": 10,  // Process in batches (if parallel)
    
    // Break Condition (optional)
    "break_condition": "$.total_processed > 100"
  },
  "on_complete": "next_step",
  "on_error": "error_handler"
}
```

**Parallel Loop Example:**

```json
{
  "id": "process_invoices",
  "type": "loop",
  "name": "Process Multiple Invoices",
  "config": {
    "collection": "$.invoices",
    "iterator_name": "invoice",
    "steps": ["validate_invoice", "create_payment"],
    "parallel": true,
    "batch_size": 5,  // Process 5 at a time
    "max_iterations": 100,
    "continue_on_error": true  // Don't stop if one fails
  }
}
```

#### 8. Wait Step

Pauses workflow execution.

```json
{
  "type": "wait",
  "config": {
    // Option 1: Wait for fixed duration
    "duration": "24h",
    
    // Option 2: Wait until specific time
    // "until": "$.scheduled_date",
    
    // Option 3: Wait for external signal
    // "for_signal": "payment_received",
    // "signal_timeout": "48h"
  },
  "on_timeout": "timeout_handler",
  "on_signal": "continue_workflow"
}
```

**Examples:**

```json
// Wait 24 hours before sending reminder
{
  "id": "wait_before_reminder",
  "type": "wait",
  "config": {
    "duration": "24h"
  },
  "on_complete": "send_reminder"
}

// Wait until scheduled date
{
  "id": "wait_until_scheduled",
  "type": "wait",
  "config": {
    "until": "$.contract.start_date"
  },
  "on_complete": "activate_contract"
}

// Wait for external event
{
  "id": "wait_for_payment",
  "type": "wait",
  "config": {
    "for_signal": "payment_confirmed",
    "signal_timeout": "7d"
  },
  "on_signal": "fulfill_order",
  "on_timeout": "cancel_order"
}
```

### Expression Language

The workflow engine uses two complementary expression mechanisms. Understanding where each applies prevents confusion:

---

#### 1. JSONPath — Data Access in Step Configs

**Scope:** `field`, `params`, `recipients`, `output_mapping`, `entity_scope` fields inside step `config` objects.

JSONPath `$.` prefixed paths select values from the live `WorkflowContext` at execution time.

```javascript
// Access trigger/input data
$.invoice.amount
$.invoice.vendor_id

// Access variables set by previous steps
$.variables.approval_level

// Access a specific step's output
$.steps.validate_invoice.result

// Access workflow execution context
$.context.user_id
$.context.tenant_id
$.context.entity_id
```

**How it works internally:** `engine/expression.go` strips the `$.` prefix, then delegates to `pkg/condition/builder.go`'s `EvalContext.GetValue()`, which traverses the context data map using dot-notation. Complex JSONPath selectors (filters, wildcards `[*]`) are handled in the expression adapter layer.

---

#### 2. Condition Builder — Logic Evaluation

**Scope:** `condition` step `config.rule`, `config.group`, `config.formula`, and `config.cases[*].formula`.

Uses `pkg/condition/builder.go` backed by `expr-lang` (compiled at template validation, cached, type-checked). **No arbitrary code execution — no JavaScript.**

Field references inside formulas use dot-notation **without** the `$.` prefix:

```go
// In formula strings — dot notation, no $. prefix
"invoice.amount > 10000"
"entity.type == 'COMPANY' && invoice.status != 'cancelled'"
"now.After(invoice.due_date)"   // 'now' is injected by EvalContext
```

**Available built-in functions** (registered on `EvalContext`):

| Category | Functions |
|---|---|
| Date/Time | `now` (variable), custom handlers via `RegisterFunction` |
| String | handled via expr-lang operators |
| Comparison | all `OperatorType` values from `pkg/condition/builder.go` |

Custom functions are registered via `EvalContext.RegisterFunction(name, handler)` and available in all formula expressions.

---

#### 3. Template Strings — Display Text Interpolation

**Scope:** `title`, `description`, `notification.data` fields in step configs.

Uses `${variable}` syntax resolved at runtime from `WorkflowContext.Variables`:

```javascript
"Invoice ${invoice.invoice_number} from ${invoice.vendor_name}"
"Amount: $${invoice.amount}"
"Assigned to ${assignee_name} — due ${due_date}"
```

Uses dot-notation path resolution (same as condition builder) for nested paths like `${invoice.vendor.name}`.

---

#### Summary: Which syntax where

| Location | Syntax | Resolver |
|---|---|---|
| `config.params.*`, `config.field`, `config.entity_scope` | `$.invoice.amount` | JSONPath adapter → `GetValue` |
| `config.rule`, `config.group` | ConditionRule/Group JSON | `pkg/condition/builder.go` Evaluator |
| `config.formula`, `config.cases[*].formula` | `invoice.amount > 10000` | expr-lang via `evaluateFormula` |
| `config.title`, `config.description`, notification `data` | `${invoice.amount}` | `resolveTemplate` string substitution |

### Validation Rules

Before execution, workflow definitions are validated:

1. **Structure Validation**
   - Valid JSON structure
   - Required fields present
   - Correct data types

2. **Step Validation**
   - All step IDs are unique
   - All referenced steps exist
   - Valid step types
   - Required config fields present

3. **Transition Validation**
   - No orphaned steps
   - No circular dependencies
   - All paths lead to 'end' or exit step
   - No unreachable steps

4. **Expression Validation**
   - Valid JSONPath syntax
   - Referenced fields exist in context
   - Type compatibility

5. **Resource Limits**
   - Maximum steps: 100
   - Maximum nesting depth: 10
   - Maximum loop iterations: 10,000
   - Maximum parallel branches: 20

---

## Backend Implementation

### Project Structure

```
internal/
├── workflow/
│   ├── domain/                       # ← Shared types imported by ALL sub-packages
│   │   ├── types.go                  # StepDefinition, WorkflowDefinition, VariableDefinition
│   │   ├── result.go                 # StepResult, WorkflowContext, TaskDecision
│   │   └── errors.go                 # Sentinel errors (ErrUnauthorized, ErrMaxDelegation…)
│   │   # IMPORTANT: engine/ and executors/ both import domain/ — never import each other.
│   │   # This breaks the circular dependency: engine creates executors via constructor
│   │   # injection using the StepExecutor interface defined in domain/.
│   ├── engine/
│   │   ├── engine.go                 # Core engine orchestration
│   │   ├── validator.go              # Definition validation
│   │   ├── expression.go             # JSONPath → dot-notation adapter (wraps pkg/condition)
│   │   ├── registry.go               # Step executor registry
│   │   └── assignment.go             # User task assignment
│   ├── executors/
│   │   ├── validation.go             # Validation step executor
│   │   ├── condition.go              # Condition step executor (uses pkg/condition/builder.go)
│   │   ├── user_task.go              # User task executor
│   │   ├── action.go                 # Action step executor
│   │   ├── notification.go           # Notification executor
│   │   ├── loop.go                   # Loop executor
│   │   ├── parallel.go               # Parallel executor
│   │   └── wait.go                   # Wait executor
│   ├── temporal/
│   │   ├── workflows.go              # Temporal workflow definitions
│   │   ├── activities.go             # Temporal activities
│   │   └── worker.go                 # Temporal worker setup
│   ├── handlers/
│   │   ├── template_handler.go       # Template CRUD API
│   │   ├── instance_handler.go       # Instance management API
│   │   ├── task_handler.go           # User task API
│   │   └── admin_handler.go          # Admin/metrics API
│   ├── service.go                    # Workflow service interface
│   └── models.go                     # Domain models
├── shared/
│   ├── logger/                       # Logging utilities
│   └── context.go                    # Tenant context helpers
└── db/
    └── sqlc/                         # Generated database code
        ├── store.go                  # Database store interface
        └── queries.sql.go            # Generated queries
```

### Core Engine Implementation

```go
// internal/workflow/engine/engine.go
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "awo.so/db/sqlc"
	"awo.so/internal/shared/logger"
	"go.temporal.io/sdk/client"
)

// Engine orchestrates workflow execution
type Engine struct {
	store          db.Store
	temporal       client.Client
	stepRegistry   *StepRegistry
	actionRegistry *ActionRegistry
	validator      *Validator
	expressionEval *ExpressionEvaluator
	logger         logger.Logger
}

// Config holds engine configuration
type Config struct {
	Store              db.Store
	TemporalClient     client.Client
	Logger             logger.Logger
	MaxConcurrentTasks int
	DefaultTimeout     time.Duration
}

// NewEngine creates a new workflow engine
func NewEngine(cfg Config) (*Engine, error) {
	if cfg.Store == nil {
		return nil, fmt.Errorf("store is required")
	}
	if cfg.TemporalClient == nil {
		return nil, fmt.Errorf("temporal client is required")
	}
	if cfg.Logger == nil {
		cfg.Logger = logger.WithFields(logger.Fields{"component": "workflow-engine"})
	}

	engine := &Engine{
		store:          cfg.Store,
		temporal:       cfg.TemporalClient,
		stepRegistry:   NewStepRegistry(),
		actionRegistry: NewActionRegistry(),
		validator:      NewValidator(),
		expressionEval: NewExpressionEvaluator(),
		logger:         cfg.Logger,
	}

	// Register step executors
	engine.registerStepExecutors()

	return engine, nil
}

// registerStepExecutors registers all step types
func (e *Engine) registerStepExecutors() {
	e.stepRegistry.Register("validation", NewValidationExecutor(e))
	e.stepRegistry.Register("condition", NewConditionExecutor(e))
	e.stepRegistry.Register("user_task", NewUserTaskExecutor(e))
	e.stepRegistry.Register("action", NewActionExecutor(e))
	e.stepRegistry.Register("notification", NewNotificationExecutor(e))
	e.stepRegistry.Register("parallel", NewParallelExecutor(e))
	e.stepRegistry.Register("loop", NewLoopExecutor(e))
	e.stepRegistry.Register("wait", NewWaitExecutor(e))
	e.stepRegistry.Register("webhook", NewWebhookExecutor(e))
}

// WorkflowDefinition represents a complete workflow
type WorkflowDefinition struct {
	Name        string                `json:"name"`
	Version     int                   `json:"version"`
	Description string                `json:"description"`
	Variables   []VariableDefinition  `json:"variables"`
	Steps       []StepDefinition      `json:"steps"`
}

// VariableDefinition defines a workflow variable
type VariableDefinition struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, integer, decimal, boolean, object, array
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
}

// StepDefinition defines a workflow step
type StepDefinition struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config"`

	// Transitions
	OnSuccess  string `json:"on_success,omitempty"`
	OnFailure  string `json:"on_failure,omitempty"`
	OnTimeout  string `json:"on_timeout,omitempty"`
	OnApproved string `json:"on_approved,omitempty"`
	OnRejected string `json:"on_rejected,omitempty"`
	OnTrue     string `json:"on_true,omitempty"`
	OnFalse    string `json:"on_false,omitempty"`
}

// CreateTemplateRequest represents a request to create a workflow template
type CreateTemplateRequest struct {
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Category           string                 `json:"category"`
	TriggerType        string                 `json:"trigger_type"`
	TriggerEvent       string                 `json:"trigger_event,omitempty"`
	TriggerConditions  map[string]interface{} `json:"trigger_conditions,omitempty"`
	DefinitionJSON     json.RawMessage        `json:"definition"`
	AMISBuilderState   json.RawMessage        `json:"amis_builder_state,omitempty"`
	AllowedRoles       []string               `json:"allowed_roles,omitempty"`
	AllowedEntityIDs   []uuid.UUID            `json:"allowed_entity_ids,omitempty"`
	IsDraft            bool                   `json:"is_draft"`
}

// CreateTemplate creates a new workflow template
func (e *Engine) CreateTemplate(ctx context.Context, req CreateTemplateRequest) (*db.WorkflowTemplate, error) {
	e.logger.InfoContext(ctx, "Creating workflow template",
		logger.Fields{"name": req.Name, "trigger_type": req.TriggerType},
	)

	// Parse and validate definition
	var definition WorkflowDefinition
	if err := json.Unmarshal(req.DefinitionJSON, &definition); err != nil {
		e.logger.ErrorContext(ctx, "Failed to parse workflow definition",
			logger.Fields{"error": err.Error()},
		)
		return nil, fmt.Errorf("invalid workflow definition: %w", err)
	}

	// Validate definition
	if err := e.validator.Validate(&definition); err != nil {
		e.logger.ErrorContext(ctx, "Workflow validation failed",
			logger.Fields{"error": err.Error(), "definition": definition.Name},
		)
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Use tenant-aware transaction
	var template *db.WorkflowTemplate
	err := e.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
		// Create template record
		params := db.CreateWorkflowTemplateParams{
			Name:              req.Name,
			Description:       sql.NullString{String: req.Description, Valid: req.Description != ""},
			Category:          sql.NullString{String: req.Category, Valid: req.Category != ""},
			TriggerType:       req.TriggerType,
			TriggerEvent:      sql.NullString{String: req.TriggerEvent, Valid: req.TriggerEvent != ""},
			TriggerConditions: req.TriggerConditions,
			Definition:        req.DefinitionJSON,
			AmisBuilderState:  req.AMISBuilderState,
			IsActive:          false, // Start inactive
			IsDraft:           req.IsDraft,
			OwnerID:           getUserIDFromContext(ctx),
			AllowedRoles:      req.AllowedRoles,
			AllowedEntityIds:  req.AllowedEntityIDs,
		}

		created, err := store.CreateWorkflowTemplate(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create template: %w", err)
		}

		template = &created

		// Register trigger if event-based
		if req.TriggerType == "event" && req.TriggerEvent != "" {
			if err := e.registerEventTrigger(ctx, store, template); err != nil {
				return fmt.Errorf("failed to register trigger: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	e.logger.InfoContext(ctx, "Workflow template created successfully",
		logger.Fields{"template_id": template.ID, "name": template.Name},
	)

	return template, nil
}

// StartWorkflow initiates a workflow instance
func (e *Engine) StartWorkflow(ctx context.Context, templateID uuid.UUID, input map[string]interface{}) (*db.WorkflowInstance, error) {
	e.logger.InfoContext(ctx, "Starting workflow",
		logger.Fields{"template_id": templateID},
	)

	// Load template
	template, err := e.store.GetWorkflowTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("workflow template is not active")
	}

	// Parse definition
	var definition WorkflowDefinition
	if err := json.Unmarshal(template.Definition, &definition); err != nil {
		return nil, fmt.Errorf("invalid workflow definition: %w", err)
	}

	// Validate input variables
	if err := e.validateInput(&definition, input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	var instance *db.WorkflowInstance

	// Create instance and start Temporal workflow in transaction
	err = e.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
		// Create instance record
		instanceID := uuid.New()
		temporalWorkflowID := fmt.Sprintf("workflow-%s", instanceID)

		params := db.CreateWorkflowInstanceParams{
			ID:                 instanceID,
			TemplateID:         templateID,
			TemporalWorkflowID: temporalWorkflowID,
			// RACE CONDITION RISK: TemporalRunID is empty at create time and updated
		// after Temporal confirms the start. If the process crashes between
		// CreateWorkflowInstance and UpdateWorkflowInstance, the DB record has
		// Status="pending" and no TemporalRunID. A reconciliation job must
		// periodically query instances WHERE status='pending' AND temporal_run_id=''
		// AND started_at < NOW()-interval '5 minutes', then either restart or mark failed.
		TemporalRunID: "", // Will be updated after Temporal start
			TriggerType:        "manual",
			TriggeredBy:        sql.NullUUID{UUID: getUserIDFromContext(ctx), Valid: true},
			TriggerData:        input,
			Status:             "pending",
			TotalSteps:         int32(len(definition.Steps)),
			ContextData:        input,
		}

		created, err := store.CreateWorkflowInstance(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create instance: %w", err)
		}

		instance = &created

		// Start Temporal workflow
		workflowOptions := client.StartWorkflowOptions{
			ID:        temporalWorkflowID,
			TaskQueue: "custom-workflows",
		}

		we, err := e.temporal.ExecuteWorkflow(
			ctx,
			workflowOptions,
			"CustomWorkflowExecution",
			definition,
			instanceID,
			input,
		)

		if err != nil {
			return fmt.Errorf("failed to start temporal workflow: %w", err)
		}

		// Update instance with Temporal run ID
		updateParams := db.UpdateWorkflowInstanceParams{
			ID:            instance.ID,
			TemporalRunID: we.GetRunID(),
			Status:        "running",
		}

		if err := store.UpdateWorkflowInstance(ctx, updateParams); err != nil {
			// Try to cancel Temporal workflow
			_ = e.temporal.CancelWorkflow(ctx, temporalWorkflowID, we.GetRunID())
			return fmt.Errorf("failed to update instance: %w", err)
		}

		instance.TemporalRunID = we.GetRunID()
		instance.Status = "running"

		return nil
	})

	if err != nil {
		return nil, err
	}

	e.logger.InfoContext(ctx, "Workflow started successfully",
		logger.Fields{
			"instance_id":          instance.ID,
			"temporal_workflow_id": instance.TemporalWorkflowID,
		},
	)

	return instance, nil
}

// GetUserTasks retrieves pending tasks for a user
func (e *Engine) GetUserTasks(ctx context.Context, userID uuid.UUID, filters TaskFilters) ([]db.WorkflowUserTask, error) {
	e.logger.DebugContext(ctx, "Fetching user tasks",
		logger.Fields{"user_id": userID, "status": filters.Status},
	)

	params := db.GetUserTasksParams{
		UserID: userID,
		Status: filters.Status,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}

	tasks, err := e.store.GetUserTasks(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}

	return tasks, nil
}

// CompleteUserTask completes a user task (approval/rejection)
func (e *Engine) CompleteUserTask(ctx context.Context, taskID uuid.UUID, decision TaskDecision) error {
	e.logger.InfoContext(ctx, "Completing user task",
		logger.Fields{"task_id": taskID, "decision": decision.Decision},
	)

	userID := getUserIDFromContext(ctx)

	// Get task
	task, err := e.store.GetUserTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Verify user can complete this task
	if !e.canCompleteTask(ctx, userID, task) {
		return fmt.Errorf("user not authorized to complete this task")
	}

	// Complete in transaction
	err = e.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
		// Update task
		updateParams := db.CompleteUserTaskParams{
			ID:          taskID,
			Status:      "completed",
			CompletedBy: sql.NullUUID{UUID: userID, Valid: true},
			Decision:    sql.NullString{String: decision.Decision, Valid: true},
			Comment:     sql.NullString{String: decision.Comment, Valid: decision.Comment != ""},
			FormData:    decision.FormData,
		}

		if err := store.CompleteUserTask(ctx, updateParams); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		// Update step execution
		stepParams := db.CompleteStepExecutionParams{
			ID:                task.StepExecutionID,
			Status:            "completed",
			CompletedBy:       sql.NullUUID{UUID: userID, Valid: true},
			CompletionComment: sql.NullString{String: decision.Comment, Valid: decision.Comment != ""},
			OutputData:        decision.FormData,
		}

		if err := store.CompleteStepExecution(ctx, stepParams); err != nil {
			return fmt.Errorf("failed to update step execution: %w", err)
		}

		// Get instance for Temporal workflow ID
		instance, err := store.GetWorkflowInstance(ctx, task.InstanceID)
		if err != nil {
			return fmt.Errorf("failed to get instance: %w", err)
		}

		// Send signal to Temporal workflow to resume
		signalName := fmt.Sprintf("task-%s-completed", taskID)
		err = e.temporal.SignalWorkflow(
			ctx,
			instance.TemporalWorkflowID,
			instance.TemporalRunID,
			signalName,
			decision,
		)

		if err != nil {
			return fmt.Errorf("failed to signal temporal workflow: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	e.logger.InfoContext(ctx, "User task completed successfully",
		logger.Fields{"task_id": taskID, "decision": decision.Decision},
	)

	return nil
}

// canCompleteTask checks if user is authorized to complete task
func (e *Engine) canCompleteTask(ctx context.Context, userID uuid.UUID, task db.WorkflowUserTask) bool {
	// Check direct assignment
	if task.AssignedToUserID.Valid && task.AssignedToUserID.UUID == userID {
		return true
	}

	// Check role assignment
	if task.AssignedToRole.Valid {
		userRoles, err := e.store.GetUserRoles(ctx, userID)
		if err != nil {
			return false
		}

		for _, role := range userRoles {
			if role == task.AssignedToRole.String {
				// If entity-scoped, verify user has role in that entity
				if task.AssignedToEntityID.Valid {
					hasRoleInEntity, err := e.store.UserHasRoleInEntity(
						ctx,
						userID,
						task.AssignedToRole.String,
						task.AssignedToEntityID.UUID,
					)
					return err == nil && hasRoleInEntity
				}
				return true
			}
		}
	}

	return false
}

// validateInput validates workflow input against variable definitions
func (e *Engine) validateInput(definition *WorkflowDefinition, input map[string]interface{}) error {
	for _, varDef := range definition.Variables {
		if varDef.Required {
			if _, exists := input[varDef.Name]; !exists {
				return fmt.Errorf("required variable %s is missing", varDef.Name)
			}
		}
	}
	return nil
}

// Helper to register event trigger
func (e *Engine) registerEventTrigger(ctx context.Context, store db.Store, template *db.WorkflowTemplate) error {
	params := db.CreateWorkflowTriggerParams{
		TemplateID:  template.ID,
		TriggerType: "event",
		EventName:   template.TriggerEvent,
		IsActive:    template.IsActive,
	}

	_, err := store.CreateWorkflowTrigger(ctx, params)
	return err
}

// TaskFilters for filtering user tasks
type TaskFilters struct {
	Status string
	Limit  int32
	Offset int32
}

// TaskDecision represents a user's decision on a task
type TaskDecision struct {
	Decision    string                 `json:"decision"` // approved, rejected, delegated, etc.
	Comment     string                 `json:"comment"`
	FormData    map[string]interface{} `json:"form_data"`
	CompletedBy uuid.UUID              `json:"-"`
}

// ctxKey is a typed key to avoid context value shadowing across packages.
// Use this instead of raw strings as context keys.
type ctxKey string

const (
	ctxKeyUserID          ctxKey = "user_id"
	ctxKeyStepExecutionID ctxKey = "step_execution_id"
)

// Helper to get user ID from context
func getUserIDFromContext(ctx context.Context) uuid.UUID {
	// Extract from your auth middleware using typed key (prevents shadowing)
	userID, _ := ctx.Value(ctxKeyUserID).(uuid.UUID)
	return userID
}
```

### Step Executors

#### Base Executor Interface

```go
// internal/workflow/executors/executor.go
package executors

import (
	"context"

	"github.com/google/uuid"
)

// StepExecutor interface that all executors must implement
type StepExecutor interface {
	Execute(ctx context.Context, step StepDefinition, wfCtx WorkflowContext) (*StepResult, error)
	Type() string
}

// StepResult contains the execution result
type StepResult struct {
	Success    bool
	Output     map[string]interface{}
	NextStepID string
	Error      error
}

// WorkflowContext holds workflow execution context
type WorkflowContext struct {
	InstanceID  uuid.UUID
	Variables   map[string]interface{}
	StepOutputs map[string]interface{}
	Input       map[string]interface{}
}

// StepDefinition is imported from engine package
type StepDefinition struct {
	ID          string
	Type        string
	Name        string
	Description string
	Config      map[string]interface{}

	OnSuccess  string
	OnFailure  string
	OnTimeout  string
	OnApproved string
	OnRejected string
	OnTrue     string
	OnFalse    string
}
```

#### Validation Executor

```go
// internal/workflow/executors/validation.go
package executors

import (
	"context"
	"fmt"

	"awo.so/internal/shared/logger"
	"awo.so/internal/workflow/engine"
)

// ValidationExecutor handles validation steps
type ValidationExecutor struct {
	engine *engine.Engine
	logger logger.Logger
}

// NewValidationExecutor creates a new validation executor
func NewValidationExecutor(eng *engine.Engine) *ValidationExecutor {
	return &ValidationExecutor{
		engine: eng,
		logger: eng.Logger().WithFields(logger.Fields{"executor": "validation"}),
	}
}

// Type returns the executor type
func (e *ValidationExecutor) Type() string {
	return "validation"
}

// Execute executes a validation step
func (e *ValidationExecutor) Execute(ctx context.Context, step StepDefinition, wfCtx WorkflowContext) (*StepResult, error) {
	e.logger.DebugContext(ctx, "Executing validation step",
		logger.Fields{"step_id": step.ID, "step_name": step.Name},
	)

	rules, ok := step.Config["rules"].([]interface{})
	if !ok {
		return &StepResult{
			Success:    false,
			NextStepID: step.OnFailure,
			Error:      fmt.Errorf("invalid rules configuration"),
		}, nil
	}

	var errors []string
	collectAllErrors := getBool(step.Config, "collect_all_errors", true)

	for i, r := range rules {
		rule, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		field := getString(rule, "field", "")
		operator := getString(rule, "operator", "")
		expectedValue := rule["value"]
		errorMessage := getString(rule, "error_message", fmt.Sprintf("Validation rule %d failed", i))

		// Evaluate JSONPath expression
		actualValue, err := e.engine.ExpressionEval().Evaluate(field, wfCtx)
		if err != nil {
			if collectAllErrors {
				errors = append(errors, fmt.Sprintf("Field %s evaluation failed: %v", field, err))
				continue
			}
			return &StepResult{
				Success:    false,
				NextStepID: step.OnFailure,
				Error:      fmt.Errorf("field evaluation failed: %w", err),
			}, nil
		}

		// Perform validation
		valid := e.validateValue(actualValue, operator, expectedValue)
		if !valid {
			if collectAllErrors {
				errors = append(errors, errorMessage)
				continue
			}
			return &StepResult{
				Success:    false,
				NextStepID: step.OnFailure,
				Error:      fmt.Errorf(errorMessage),
				Output: map[string]interface{}{
					"validation_error": errorMessage,
					"field":            field,
				},
			}, nil
		}
	}

	if len(errors) > 0 {
		return &StepResult{
			Success:    false,
			NextStepID: step.OnFailure,
			Error:      fmt.Errorf("validation failed: %d errors", len(errors)),
			Output: map[string]interface{}{
				"validation_errors": errors,
			},
		}, nil
	}

	return &StepResult{
		Success:    true,
		NextStepID: step.OnSuccess,
		Output: map[string]interface{}{
			"validation_passed": true,
		},
	}, nil
}

// validateValue performs the actual validation based on operator
func (e *ValidationExecutor) validateValue(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "=":
		return actual == expected
	case "!=":
		return actual != expected
	case ">":
		return compareNumbers(actual, expected) > 0
	case ">=":
		return compareNumbers(actual, expected) >= 0
	case "<":
		return compareNumbers(actual, expected) < 0
	case "<=":
		return compareNumbers(actual, expected) <= 0
	case "exists":
		return actual != nil
	case "not_exists":
		return actual == nil
	case "in":
		values, ok := expected.([]interface{})
		if !ok {
			return false
		}
		for _, v := range values {
			if actual == v {
				return true
			}
		}
		return false
	case "regex":
		pattern, ok := expected.(string)
		if !ok {
			return false
		}
		str, ok := actual.(string)
		if !ok {
			return false
		}
		matched, _ := regexp.MatchString(pattern, str)
		return matched
	case "contains":
		str, ok := actual.(string)
		if !ok {
			return false
		}
		substring, ok := expected.(string)
		if !ok {
			return false
		}
		return strings.Contains(str, substring)
	}
	return false
}

// Helper functions
func compareNumbers(a, b interface{}) int {
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat < bFloat {
		return -1
	}
	if aFloat > bFloat {
		return 1
	}
	return 0
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func getBool(m map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}

func getString(m map[string]interface{}, key string, defaultVal string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}
```

#### User Task Executor

```go
// internal/workflow/executors/user_task.go
package executors

import (
	"context"
	"fmt"
	"strings"
	"time"

	"database/sql"
	"github.com/google/uuid"
	db "awo.so/db/sqlc"
	"awo.so/internal/shared/logger"
	"awo.so/internal/workflow/engine"
)

// UserTaskExecutor handles user task (approval) steps
type UserTaskExecutor struct {
	engine *engine.Engine
	store  db.Store
	logger logger.Logger
}

// NewUserTaskExecutor creates a new user task executor
func NewUserTaskExecutor(eng *engine.Engine) *UserTaskExecutor {
	return &UserTaskExecutor{
		engine: eng,
		store:  eng.Store(),
		logger: eng.Logger().WithFields(logger.Fields{"executor": "user_task"}),
	}
}

// Type returns the executor type
func (e *UserTaskExecutor) Type() string {
	return "user_task"
}

// Execute executes a user task step
func (e *UserTaskExecutor) Execute(ctx context.Context, step StepDefinition, wfCtx WorkflowContext) (*StepResult, error) {
	e.logger.InfoContext(ctx, "Creating user task",
		logger.Fields{"step_id": step.ID, "instance_id": wfCtx.InstanceID},
	)

	config := step.Config

	// Resolve template variables in title and description
	title := e.resolveTemplate(getString(config, "title", ""), wfCtx)
	description := e.resolveTemplate(getString(config, "description", ""), wfCtx)
	priority := getString(config, "priority", "normal")
	taskType := getString(config, "task_type", "approval")

	// Determine assignee
	assignedToUserID := e.resolveUserAssignment(ctx, config, wfCtx)
	assignedToRole := e.resolveRoleAssignment(ctx, config, wfCtx)
	assignedToEntityID := e.resolveEntityAssignment(ctx, config, wfCtx)

	// Parse timeout and calculate due date
	var dueAt sql.NullTime
	if timeoutStr := getString(config, "timeout", ""); timeoutStr != "" {
		duration, err := parseDuration(timeoutStr)
		if err == nil {
			due := time.Now().Add(duration)
			dueAt = sql.NullTime{Time: due, Valid: true}
		}
	}

	// Generate form schema (AMIS format)
	formSchema := e.generateFormSchema(config["form_fields"])

	// Get action buttons
	actionButtons := config["actions"]

	// Create user task in database
	taskID := uuid.New()
	
	createParams := db.CreateUserTaskParams{
		ID:                 taskID,
		StepExecutionID:    getStepExecutionIDFromContext(ctx),
		InstanceID:         wfCtx.InstanceID,
		Title:              title,
		Description:        sql.NullString{String: description, Valid: description != ""},
		TaskType:           taskType,
		Priority:           priority,
		AssignedToUserID:   assignedToUserID,
		AssignedToRole:     assignedToRole,
		AssignedToEntityID: assignedToEntityID,
		DueAt:              dueAt,
		Status:             "pending",
		FormSchema:         formSchema,
		ActionButtons:      actionButtons,
		ContextData:        wfCtx.Variables,
		CanDelegate:        getBool(config, "can_delegate", true),
		CanReassign:        getBool(config, "can_reassign", false),
		RequiresComment:    getBool(config, "requires_comment", false),
	}

	task, err := e.store.CreateUserTask(ctx, createParams)
	if err != nil {
		e.logger.ErrorContext(ctx, "Failed to create user task",
			logger.Fields{"error": err.Error()},
		)
		return &StepResult{
			Success:    false,
			NextStepID: step.OnFailure,
			Error:      fmt.Errorf("failed to create user task: %w", err),
		}, nil
	}

	// Send notification to assignee
	if err := e.notifyTaskAssignment(ctx, &task); err != nil {
		e.logger.ErrorContext(ctx, "Failed to send task notification",
			logger.Fields{"task_id": taskID, "error": err.Error()},
		)
		// Non-fatal error, continue
	}

	e.logger.InfoContext(ctx, "User task created successfully",
		logger.Fields{"task_id": taskID, "title": title},
	)

	// Return result with task ID
	// The Temporal workflow will wait for signal to resume
	return &StepResult{
		Success: true,
		Output: map[string]interface{}{
			"task_id":     taskID.String(),
			"task_status": "waiting",
			"assigned_to": e.getAssignmentDescription(assignedToUserID, assignedToRole, assignedToEntityID),
		},
		// NextStepID will be determined after task completion
	}, nil
}

// resolveTemplate replaces ${variable} placeholders
func (e *UserTaskExecutor) resolveTemplate(template string, wfCtx WorkflowContext) string {
	result := template
	for key, value := range wfCtx.Variables {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprint(value))
	}
	// Also support nested paths like ${invoice.amount}
	for key, value := range flattenMap("", wfCtx.Variables) {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprint(value))
	}
	return result
}

// resolveUserAssignment determines user assignment
func (e *UserTaskExecutor) resolveUserAssignment(ctx context.Context, config map[string]interface{}, wfCtx WorkflowContext) sql.NullUUID {
	if userIDStr, ok := config["assign_to_user_id"].(string); ok {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			return sql.NullUUID{UUID: userID, Valid: true}
		}
	}
	return sql.NullUUID{Valid: false}
}

// resolveRoleAssignment determines role assignment
func (e *UserTaskExecutor) resolveRoleAssignment(ctx context.Context, config map[string]interface{}, wfCtx WorkflowContext) sql.NullString {
	if role, ok := config["assign_to_role"].(string); ok {
		return sql.NullString{String: role, Valid: true}
	}
	if roleInEntity, ok := config["assign_to_role_in_entity"].(map[string]interface{}); ok {
		if role, ok := roleInEntity["role"].(string); ok {
			return sql.NullString{String: role, Valid: true}
		}
	}
	return sql.NullString{Valid: false}
}

// resolveEntityAssignment determines entity assignment
func (e *UserTaskExecutor) resolveEntityAssignment(ctx context.Context, config map[string]interface{}, wfCtx WorkflowContext) sql.NullUUID {
	var entityPath string

	// Check for entity_scope (simple path)
	if scope, ok := config["entity_scope"].(string); ok {
		entityPath = scope
	}

	// Check for assign_to_role_in_entity (nested config)
	if roleInEntity, ok := config["assign_to_role_in_entity"].(map[string]interface{}); ok {
		if entity, ok := roleInEntity["entity"].(string); ok {
			entityPath = entity
		}
	}

	if entityPath != "" {
		// Evaluate JSONPath to get entity ID
		if entityIDVal, err := e.engine.ExpressionEval().Evaluate(entityPath, wfCtx); err == nil {
			if entityIDStr, ok := entityIDVal.(string); ok {
				if entityID, err := uuid.Parse(entityIDStr); err == nil {
					return sql.NullUUID{UUID: entityID, Valid: true}
				}
			}
		}
	}

	return sql.NullUUID{Valid: false}
}

// generateFormSchema creates AMIS form schema
func (e *UserTaskExecutor) generateFormSchema(fields interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}

	formFields, ok := fields.([]interface{})
	if !ok {
		return nil
	}

	amisFields := make([]map[string]interface{}, len(formFields))

	for i, f := range formFields {
		field, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		amisFields[i] = map[string]interface{}{
			"type":        getString(field, "type", "input-text"),
			"name":        getString(field, "name", ""),
			"label":       getString(field, "label", ""),
			"placeholder": getString(field, "placeholder", ""),
			"required":    getBool(field, "required", false),
		}

		// Add options for select fields
		if options, ok := field["options"]; ok {
			amisFields[i]["options"] = options
		}
	}

	return map[string]interface{}{
		"type": "form",
		"body": amisFields,
	}
}

// notifyTaskAssignment sends notification to assigned user
func (e *UserTaskExecutor) notifyTaskAssignment(ctx context.Context, task *db.WorkflowUserTask) error {
	// TODO: Integrate with notification service
	e.logger.InfoContext(ctx, "Sending task assignment notification",
		logger.Fields{"task_id": task.ID, "title": task.Title},
	)
	return nil
}

// getAssignmentDescription returns human-readable assignment description
func (e *UserTaskExecutor) getAssignmentDescription(userID sql.NullUUID, role sql.NullString, entityID sql.NullUUID) string {
	if userID.Valid {
		return fmt.Sprintf("User: %s", userID.UUID)
	}
	if role.Valid {
		if entityID.Valid {
			return fmt.Sprintf("Role: %s in Entity: %s", role.String, entityID.UUID)
		}
		return fmt.Sprintf("Role: %s", role.String)
	}
	return "Unassigned"
}

// Helper to get step execution ID from context.
// Returns uuid.Nil and an error if not set — callers must handle this to avoid
// orphaned step execution records with no parent instance.
func getStepExecutionIDFromContext(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(ctxKeyStepExecutionID).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("step_execution_id not found in context")
	}
	return id, nil
}

// Helper to flatten nested map for template resolution
func flattenMap(prefix string, m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if nested, ok := v.(map[string]interface{}); ok {
			for nk, nv := range flattenMap(key, nested) {
				result[nk] = nv
			}
		} else {
			result[key] = v
		}
	}
	return result
}

// parseDuration parses duration strings like "48h", "2d", "30m"
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(days*24) * time.Hour, nil
	}
	return time.ParseDuration(s)
}
```



---

## Temporal Integration

### Temporal Workflow Implementation

```go
// internal/workflow/temporal/workflows.go
package temporal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// CustomWorkflowExecution is the main Temporal workflow
func CustomWorkflowExecution(
	ctx workflow.Context,
	definition WorkflowDefinition,
	instanceID uuid.UUID,
	input map[string]interface{},
) error {

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting custom workflow",
		"instanceID", instanceID,
		"name", definition.Name,
		"version", definition.Version,
	)

	// Setup activity options with retries
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Initialize workflow context
	wfContext := WorkflowContext{
		InstanceID:  instanceID,
		Variables:   initializeVariables(definition.Variables, input),
		StepOutputs: make(map[string]interface{}),
		Input:       input,
	}

	// Execute workflow steps
	currentStepID := definition.Steps[0].ID
	stepCount := 0
	maxSteps := len(definition.Steps) * 10 // Prevent infinite loops

	for currentStepID != "" && currentStepID != "end" {
		stepCount++
		if stepCount > maxSteps {
			return fmt.Errorf("workflow exceeded maximum step count (%d)", maxSteps)
		}

		// Find step definition
		step := findStep(definition.Steps, currentStepID)
		if step == nil {
			return fmt.Errorf("step not found: %s", currentStepID)
		}

		logger.Info("Executing step",
			"stepID", step.ID,
			"stepType", step.Type,
			"stepName", step.Name,
			"stepOrder", stepCount,
		)

		// Record step start
		var stepExecID uuid.UUID
		err := workflow.ExecuteActivity(ctx, RecordStepStartActivity, instanceID, *step, stepCount).Get(ctx, &stepExecID)
		if err != nil {
			logger.Error("Failed to record step start", "error", err)
			return err
		}

		// Add step execution ID to context for activity access
		ctx = workflow.WithValue(ctx, "step_execution_id", stepExecID)

		// Execute step based on type
		var result StepResult
		var stepErr error

		switch step.Type {
		case "validation":
			stepErr = workflow.ExecuteActivity(ctx, ExecuteValidationActivity, instanceID, *step, wfContext).Get(ctx, &result)

		case "condition":
			stepErr = workflow.ExecuteActivity(ctx, ExecuteConditionActivity, instanceID, *step, wfContext).Get(ctx, &result)

		case "user_task":
			result, stepErr = executeUserTaskStep(ctx, instanceID, *step, wfContext)

		case "action":
			stepErr = workflow.ExecuteActivity(ctx, ExecuteActionActivity, instanceID, *step, wfContext).Get(ctx, &result)

		case "notification":
			stepErr = workflow.ExecuteActivity(ctx, ExecuteNotificationActivity, instanceID, *step, wfContext).Get(ctx, &result)

		case "parallel":
			result, stepErr = executeParallelStep(ctx, instanceID, *step, wfContext, definition.Steps)

		case "loop":
			stepErr = workflow.ExecuteActivity(ctx, ExecuteLoopActivity, instanceID, *step, wfContext).Get(ctx, &result)

		case "wait":
			result, stepErr = executeWaitStep(ctx, instanceID, *step, wfContext)

		case "webhook":
			stepErr = workflow.ExecuteActivity(ctx, ExecuteWebhookActivity, instanceID, *step, wfContext).Get(ctx, &result)

		default:
			stepErr = fmt.Errorf("unknown step type: %s", step.Type)
		}

		// Handle step errors
		if stepErr != nil {
			logger.Error("Step execution failed",
				"stepID", step.ID,
				"stepType", step.Type,
				"error", stepErr,
			)

			// Record step failure
			_ = workflow.ExecuteActivity(ctx, RecordStepFailureActivity, instanceID, stepExecID, step.ID, stepErr.Error()).Get(ctx, nil)

			// Determine next step based on error handling
			if step.OnFailure != "" {
				currentStepID = step.OnFailure
				continue
			}

			// Mark instance as failed
			_ = workflow.ExecuteActivity(ctx, MarkInstanceFailedActivity, instanceID, stepErr.Error()).Get(ctx, nil)
			return stepErr
		}

		// Update context with step output
		wfContext.StepOutputs[step.ID] = result.Output
		for key, value := range result.Output {
			wfContext.Variables[key] = value
		}

		// Record step completion
		_ = workflow.ExecuteActivity(ctx, RecordStepCompletionActivity, instanceID, stepExecID, step.ID, result.Output).Get(ctx, nil)

		// Determine next step
		currentStepID = result.NextStepID

		logger.Info("Step completed successfully",
			"stepID", step.ID,
			"nextStep", currentStepID,
		)
	}

	logger.Info("Workflow completed successfully", "instanceID", instanceID)

	// Mark workflow instance as completed
	_ = workflow.ExecuteActivity(ctx, CompleteWorkflowInstanceActivity, instanceID, wfContext.Variables).Get(ctx, nil)

	return nil
}

// executeUserTaskStep handles user task execution with waiting
func executeUserTaskStep(
	ctx workflow.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
) (StepResult, error) {

	logger := workflow.GetLogger(ctx)

	// Create user task
	var taskResult StepResult
	err := workflow.ExecuteActivity(ctx, CreateUserTaskActivity, instanceID, step, wfContext).Get(ctx, &taskResult)
	if err != nil {
		return StepResult{}, err
	}

	taskID := taskResult.Output["task_id"].(string)

	// Setup timeout
	timeout := 48 * time.Hour // Default timeout
	if timeoutStr, ok := step.Config["timeout"].(string); ok {
		timeout, _ = parseDuration(timeoutStr)
	}

	// Wait for task completion or timeout
	selector := workflow.NewSelector(ctx)
	var taskDecision TaskDecision
	taskCompleted := false
	timedOut := false

	// Setup signal channel for task completion
	signalName := fmt.Sprintf("task-%s-completed", taskID)
	signalChan := workflow.GetSignalChannel(ctx, signalName)
	selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &taskDecision)
		taskCompleted = true
	})

	// Setup timer for timeout
	timer := workflow.NewTimer(ctx, timeout)
	selector.AddFuture(timer, func(f workflow.Future) {
		timedOut = true
	})

	// Wait for completion or timeout
	selector.Select(ctx)

	if timedOut {
		logger.Warn("Task timed out", "taskID", taskID, "timeout", timeout)

		// Handle timeout based on configuration
		// Use getString helper — direct cast panics if key is absent or wrong type
		timeoutAction := getString(step.Config, "timeout_action", "cancel")

		switch timeoutAction {
		case "escalate":
			// Create escalation task
			var escalationResult StepResult
			err := workflow.ExecuteActivity(ctx, EscalateTaskActivity, taskID, step.Config).Get(ctx, &escalationResult)
			if err != nil {
				return StepResult{}, err
			}

			// Wait for escalated task
			escalatedTaskID := escalationResult.Output["task_id"].(string)
			escalationSignal := workflow.GetSignalChannel(ctx, fmt.Sprintf("task-%s-completed", escalatedTaskID))
			escalationSignal.Receive(ctx, &taskDecision)

		case "auto_approve":
			taskDecision = TaskDecision{
				Decision: "approved",
				Comment:  "Auto-approved due to timeout",
			}

		case "auto_reject":
			taskDecision = TaskDecision{
				Decision: "rejected",
				Comment:  "Auto-rejected due to timeout",
			}

		case "cancel":
			return StepResult{
				Success:    false,
				NextStepID: step.OnTimeout,
				Error:      fmt.Errorf("task cancelled due to timeout"),
			}, nil
		}
	}

	// Determine next step based on decision
	nextStepID := ""
	switch taskDecision.Decision {
	case "approved":
		nextStepID = step.OnApproved
	case "rejected":
		nextStepID = step.OnRejected
	case "delegated":
		// Guard against delegation cycles or unbounded recursion.
		// Each delegation increments the escalation_level stored in the task;
		// the executor must check this before creating a new task and return
		// an error (routing to on_failure) if it exceeds MaxDelegationDepth (3).
		// Implementation note: pass current depth via wfContext.Variables["_delegation_depth"].
		delegationDepth, _ := wfCtx.Variables["_delegation_depth"].(int)
		const MaxDelegationDepth = 3
		if delegationDepth >= MaxDelegationDepth {
			return StepResult{
				Success:    false,
				NextStepID: step.OnFailure,
				Error:      fmt.Errorf("max delegation depth (%d) reached", MaxDelegationDepth),
			}, nil
		}
		wfCtx.Variables["_delegation_depth"] = delegationDepth + 1
		return executeUserTaskStep(ctx, instanceID, step, wfCtx)
	default:
		nextStepID = step.OnFailure
	}

	return StepResult{
		Success:    true,
		NextStepID: nextStepID,
		Output: map[string]interface{}{
			"decision":     taskDecision.Decision,
			"comment":      taskDecision.Comment,
			"completed_by": taskDecision.CompletedBy,
			"form_data":    taskDecision.FormData,
		},
	}, nil
}

// executeParallelStep executes multiple branches in parallel
func executeParallelStep(
	ctx workflow.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
	allSteps []StepDefinition,
) (StepResult, error) {

	logger := workflow.GetLogger(ctx)
	branches := step.Config["branches"].([]interface{})
	waitFor := step.Config["wait_for"].(string) // all, any, none

	// Create child workflows for each branch
	childCtx, cancelHandler := workflow.WithCancel(ctx)
	defer cancelHandler()

	var futures []workflow.ChildWorkflowFuture
	results := make([]interface{}, len(branches))

	for i, b := range branches {
		branch := b.(map[string]interface{})
		branchID := branch["id"].(string)
		branchSteps := branch["steps"].([]interface{})

		logger.Info("Starting parallel branch", "branchID", branchID)

		// Execute branch as child workflow
		childWorkflowOptions := workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("branch-%s-%s", instanceID, branchID),
		}
		childCtx = workflow.WithChildOptions(childCtx, childWorkflowOptions)

		childFuture := workflow.ExecuteChildWorkflow(
			childCtx,
			ExecuteBranchWorkflow,
			instanceID,
			branchSteps,
			wfContext,
			allSteps,
		)

		futures = append(futures, childFuture)
	}

	// Wait based on configuration
	switch waitFor {
	case "all":
		// Wait for all branches to complete
		for i, future := range futures {
			var branchResult interface{}
			if err := future.Get(ctx, &branchResult); err != nil {
				logger.Error("Branch failed", "branchIndex", i, "error", err)
				return StepResult{Success: false, Error: err}, nil
			}
			results[i] = branchResult
		}

	case "any":
		// Wait for first branch to complete
		selector := workflow.NewSelector(ctx)
		completed := false

		for i, future := range futures {
			idx := i
			fut := future
			selector.AddFuture(fut, func(f workflow.Future) {
				if !completed {
					var branchResult interface{}
					f.Get(ctx, &branchResult)
					results[idx] = branchResult
					completed = true
					cancelHandler() // Cancel other branches
				}
			})
		}

		selector.Select(ctx)

	case "none":
		// Start all branches but don't wait (fire-and-forget)
		logger.Info("Started all branches in fire-and-forget mode")
	}

	return StepResult{
		Success:    true,
		NextStepID: step.OnSuccess,
		Output: map[string]interface{}{
			"branch_results": results,
			"branches_count": len(branches),
		},
	}, nil
}

// ExecuteBranchWorkflow executes a branch of parallel steps
func ExecuteBranchWorkflow(
	ctx workflow.Context,
	instanceID uuid.UUID,
	stepIDs []interface{},
	wfContext WorkflowContext,
	allSteps []StepDefinition,
) (interface{}, error) {

	logger := workflow.GetLogger(ctx)
	results := make([]interface{}, 0)

	for _, stepID := range stepIDs {
		step := findStep(allSteps, stepID.(string))
		if step == nil {
			return nil, fmt.Errorf("step not found: %s", stepID)
		}

		logger.Info("Executing branch step", "stepID", step.ID)

		// Execute step based on type
		var result StepResult
		var err error

		switch step.Type {
		case "action":
			err = workflow.ExecuteActivity(ctx, ExecuteActionActivity, instanceID, *step, wfContext).Get(ctx, &result)
		case "notification":
			err = workflow.ExecuteActivity(ctx, ExecuteNotificationActivity, instanceID, *step, wfContext).Get(ctx, &result)
		case "validation":
			err = workflow.ExecuteActivity(ctx, ExecuteValidationActivity, instanceID, *step, wfContext).Get(ctx, &result)
		default:
			return nil, fmt.Errorf("unsupported step type in parallel branch: %s", step.Type)
		}

		if err != nil {
			return nil, err
		}

		results = append(results, result.Output)
	}

	return results, nil
}

// executeWaitStep handles wait/delay steps
func executeWaitStep(
	ctx workflow.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
) (StepResult, error) {

	config := step.Config

	if duration, ok := config["duration"].(string); ok {
		// Wait for fixed duration
		d, _ := parseDuration(duration)
		workflow.Sleep(ctx, d)

	} else if until, ok := config["until"].(string); ok {
		// Wait until specific time
		targetTime, _ := parseTime(until)
		now := workflow.Now(ctx)
		if targetTime.After(now) {
			workflow.Sleep(ctx, targetTime.Sub(now))
		}

	} else if signalName, ok := config["for_signal"].(string); ok {
		// Wait for external signal with optional timeout
		signalChan := workflow.GetSignalChannel(ctx, signalName)
		var signalData interface{}

		if timeoutStr, ok := config["signal_timeout"].(string); ok {
			timeout, _ := parseDuration(timeoutStr)
			selector := workflow.NewSelector(ctx)
			timedOut := false

			selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
				c.Receive(ctx, &signalData)
			})

			timer := workflow.NewTimer(ctx, timeout)
			selector.AddFuture(timer, func(f workflow.Future) {
				timedOut = true
			})

			selector.Select(ctx)

			if timedOut {
				return StepResult{
					Success:    false,
					NextStepID: step.OnTimeout,
					Error:      fmt.Errorf("signal timeout"),
				}, nil
			}
		} else {
			signalChan.Receive(ctx, &signalData)
		}

		return StepResult{
			Success:    true,
			NextStepID: step.OnSuccess,
			Output: map[string]interface{}{
				"signal_data": signalData,
			},
		}, nil
	}

	return StepResult{
		Success:    true,
		NextStepID: step.OnSuccess,
	}, nil
}

// Helper functions
func initializeVariables(varDefs []VariableDefinition, input map[string]interface{}) map[string]interface{} {
	variables := make(map[string]interface{})

	for _, varDef := range varDefs {
		if value, ok := input[varDef.Name]; ok {
			variables[varDef.Name] = value
		} else if varDef.Default != nil {
			variables[varDef.Name] = varDef.Default
		}
	}

	// Also include all input data
	for k, v := range input {
		if _, exists := variables[k]; !exists {
			variables[k] = v
		}
	}

	return variables
}

func findStep(steps []StepDefinition, stepID string) *StepDefinition {
	for _, step := range steps {
		if step.ID == stepID {
			return &step
		}
	}
	return nil
}

// parseDuration is defined once in internal/workflow/domain/duration.go
// and imported by both executors/ and temporal/ packages.
// Do NOT redefine it here — this copy is shown for documentation only.
// Canonical signature: func ParseDuration(s string) (time.Duration, error)
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(days*24) * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func parseTime(s string) (time.Time, error) {
	// Try multiple formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// Types used in workflow
type StepResult struct {
	Success    bool
	Output     map[string]interface{}
	NextStepID string
	Error      error
}

type WorkflowContext struct {
	InstanceID  uuid.UUID
	Variables   map[string]interface{}
	StepOutputs map[string]interface{}
	Input       map[string]interface{}
}

type TaskDecision struct {
	Decision    string
	Comment     string
	FormData    map[string]interface{}
	CompletedBy uuid.UUID
}
```

### Temporal Activities

```go
// internal/workflow/temporal/activities.go
package temporal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "awo.so/db/sqlc"
	"awo.so/internal/shared/logger"
	"awo.so/internal/workflow/engine"
)

// Activities struct holds dependencies for Temporal activities
type Activities struct {
	engine *engine.Engine
	store  db.Store
	logger logger.Logger
}

// NewActivities creates a new activities instance
func NewActivities(eng *engine.Engine, store db.Store) *Activities {
	return &Activities{
		engine: eng,
		store:  store,
		logger: eng.Logger().WithFields(logger.Fields{"component": "temporal-activities"}),
	}
}

// RecordStepStartActivity records the start of a step execution
func (a *Activities) RecordStepStartActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	stepOrder int,
) (uuid.UUID, error) {

	stepExecID := uuid.New()

	params := db.CreateStepExecutionParams{
		ID:         stepExecID,
		InstanceID: instanceID,
		StepID:     step.ID,
		StepName:   step.Name,
		StepType:   step.Type,
		StepOrder:  int32(stepOrder),
		Status:     "running",
		StartedAt:  time.Now(),
	}

	err := a.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
		_, err := store.CreateStepExecution(ctx, params)
		return err
	})

	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to record step start",
			logger.Fields{"step_id": step.ID, "error": err.Error()},
		)
		return uuid.Nil, fmt.Errorf("failed to record step start: %w", err)
	}

	a.logger.DebugContext(ctx, "Step execution started",
		logger.Fields{"step_execution_id": stepExecID, "step_id": step.ID},
	)

	return stepExecID, nil
}

// RecordStepCompletionActivity records step completion
func (a *Activities) RecordStepCompletionActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	stepExecID uuid.UUID,
	stepID string,
	output map[string]interface{},
) error {

	outputJSON, _ := json.Marshal(output)

	err := a.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
		params := db.CompleteStepExecutionParams{
			ID:          stepExecID,
			Status:      "completed",
			CompletedAt: sql.NullTime{Time: time.Now(), Valid: true},
			OutputData:  outputJSON,
		}

		return store.CompleteStepExecution(ctx, params)
	})

	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to record step completion",
			logger.Fields{"step_execution_id": stepExecID, "error": err.Error()},
		)
		return fmt.Errorf("failed to record step completion: %w", err)
	}

	a.logger.DebugContext(ctx, "Step execution completed",
		logger.Fields{"step_execution_id": stepExecID, "step_id": stepID},
	)

	return nil
}

// RecordStepFailureActivity records step failure
func (a *Activities) RecordStepFailureActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	stepExecID uuid.UUID,
	stepID string,
	errorMessage string,
) error {

	err := a.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
		params := db.FailStepExecutionParams{
			ID:           stepExecID,
			Status:       "failed",
			CompletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
			ErrorMessage: sql.NullString{String: errorMessage, Valid: true},
		}

		return store.FailStepExecution(ctx, params)
	})

	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to record step failure",
			logger.Fields{"step_execution_id": stepExecID, "error": err.Error()},
		)
		return fmt.Errorf("failed to record step failure: %w", err)
	}

	return nil
}

// CompleteWorkflowInstanceActivity marks workflow as completed
func (a *Activities) CompleteWorkflowInstanceActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	output map[string]interface{},
) error {

	outputJSON, _ := json.Marshal(output)

	err := a.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
		params := db.CompleteWorkflowInstanceParams{
			ID:                  instanceID,
			Status:              "completed",
			CompletedAt:         sql.NullTime{Time: time.Now(), Valid: true},
			OutputData:          outputJSON,
			CompletionPercentage: 100,
		}

		return store.CompleteWorkflowInstance(ctx, params)
	})

	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to complete workflow instance",
			logger.Fields{"instance_id": instanceID, "error": err.Error()},
		)
		return fmt.Errorf("failed to complete workflow instance: %w", err)
	}

	a.logger.InfoContext(ctx, "Workflow instance completed",
		logger.Fields{"instance_id": instanceID},
	)

	return nil
}

// MarkInstanceFailedActivity marks workflow as failed
func (a *Activities) MarkInstanceFailedActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	errorMessage string,
) error {

	err := a.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
		params := db.FailWorkflowInstanceParams{
			ID:           instanceID,
			Status:       "failed",
			CompletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
			ErrorMessage: sql.NullString{String: errorMessage, Valid: true},
		}

		return store.FailWorkflowInstance(ctx, params)
	})

	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to mark instance as failed",
			logger.Fields{"instance_id": instanceID, "error": err.Error()},
		)
		return fmt.Errorf("failed to mark instance as failed: %w", err)
	}

	return nil
}

// ExecuteValidationActivity executes validation step
func (a *Activities) ExecuteValidationActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
) (StepResult, error) {

	executor := a.engine.StepRegistry().Get("validation")
	result, err := executor.Execute(ctx, step, wfContext)
	if err != nil {
		return StepResult{}, err
	}

	return *result, nil
}

// ExecuteConditionActivity executes condition step
func (a *Activities) ExecuteConditionActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
) (StepResult, error) {

	executor := a.engine.StepRegistry().Get("condition")
	result, err := executor.Execute(ctx, step, wfContext)
	if err != nil {
		return StepResult{}, err
	}

	return *result, nil
}

// CreateUserTaskActivity creates a user task
func (a *Activities) CreateUserTaskActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
) (StepResult, error) {

	executor := a.engine.StepRegistry().Get("user_task")
	result, err := executor.Execute(ctx, step, wfContext)
	if err != nil {
		return StepResult{}, err
	}

	return *result, nil
}

// ExecuteActionActivity executes action step
func (a *Activities) ExecuteActionActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
) (StepResult, error) {

	executor := a.engine.StepRegistry().Get("action")
	result, err := executor.Execute(ctx, step, wfContext)
	if err != nil {
		return StepResult{}, err
	}

	return *result, nil
}

// ExecuteNotificationActivity executes notification step
func (a *Activities) ExecuteNotificationActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
) (StepResult, error) {

	executor := a.engine.StepRegistry().Get("notification")
	result, err := executor.Execute(ctx, step, wfContext)
	if err != nil {
		return StepResult{}, err
	}

	return *result, nil
}

// ExecuteLoopActivity executes loop step
func (a *Activities) ExecuteLoopActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
) (StepResult, error) {

	executor := a.engine.StepRegistry().Get("loop")
	result, err := executor.Execute(ctx, step, wfContext)
	if err != nil {
		return StepResult{}, err
	}

	return *result, nil
}

// ExecuteWebhookActivity executes webhook step
func (a *Activities) ExecuteWebhookActivity(
	ctx context.Context,
	instanceID uuid.UUID,
	step StepDefinition,
	wfContext WorkflowContext,
) (StepResult, error) {

	executor := a.engine.StepRegistry().Get("webhook")
	result, err := executor.Execute(ctx, step, wfContext)
	if err != nil {
		return StepResult{}, err
	}

	return *result, nil
}

// EscalateTaskActivity escalates a timed-out task
func (a *Activities) EscalateTaskActivity(
	ctx context.Context,
	taskID string,
	config map[string]interface{},
) (StepResult, error) {

	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		return StepResult{}, fmt.Errorf("invalid task ID: %w", err)
	}

	var newTask *db.WorkflowUserTask

	err = a.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
		// Get original task
		originalTask, err := store.GetUserTask(ctx, taskUUID)
		if err != nil {
			return fmt.Errorf("failed to get original task: %w", err)
		}

		// Mark original as escalated
		escalateParams := db.EscalateUserTaskParams{
			ID:          taskUUID,
			Status:      "escalated",
			EscalatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}

		if err := store.EscalateUserTask(ctx, escalateParams); err != nil {
			return fmt.Errorf("failed to escalate task: %w", err)
		}

		// Create escalated task
		escalateToRole := config["escalate_to_role"].(string)

		newTaskParams := db.CreateUserTaskParams{
			ID:                 uuid.New(),
			StepExecutionID:    originalTask.StepExecutionID,
			InstanceID:         originalTask.InstanceID,
			Title:              "[ESCALATED] " + originalTask.Title,
			Description:        sql.NullString{String: fmt.Sprintf("Escalated from %s", originalTask.AssignedToRole.String), Valid: true},
			TaskType:           originalTask.TaskType,
			Priority:           "urgent",
			AssignedToRole:     sql.NullString{String: escalateToRole, Valid: true},
			AssignedToEntityID: originalTask.AssignedToEntityID,
			DueAt:              sql.NullTime{Time: time.Now().Add(48 * time.Hour), Valid: true},
			Status:             "pending",
			FormSchema:         originalTask.FormSchema,
			ActionButtons:      originalTask.ActionButtons,
			ContextData:        originalTask.ContextData,
			EscalationLevel:    originalTask.EscalationLevel + 1,
		}

		created, err := store.CreateUserTask(ctx, newTaskParams)
		if err != nil {
			return fmt.Errorf("failed to create escalated task: %w", err)
		}

		newTask = &created
		return nil
	})

	if err != nil {
		return StepResult{}, err
	}

	a.logger.InfoContext(ctx, "Task escalated successfully",
		logger.Fields{
			"original_task_id": taskID,
			"new_task_id":      newTask.ID,
			"escalated_to":     config["escalate_to_role"],
		},
	)

	return StepResult{
		Success: true,
		Output: map[string]interface{}{
			"task_id":        newTask.ID.String(),
			"escalated_from": taskID,
		},
	}, nil
}
```

### Temporal Worker Setup

```go
// internal/workflow/temporal/worker.go
package temporal

import (
	"context"
	"fmt"

	db "awo.so/db/sqlc"
	"awo.so/internal/shared/logger"
	"awo.so/internal/workflow/engine"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	TaskQueue          string
	MaxConcurrentTasks int
	MaxConcurrentWFs   int
}

// DefaultWorkerConfig returns default worker configuration
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		TaskQueue:          "custom-workflows",
		MaxConcurrentTasks: 100,
		MaxConcurrentWFs:   50,
	}
}

// StartWorker starts a Temporal worker
func StartWorker(
	ctx context.Context,
	temporalClient client.Client,
	store db.Store,
	eng *engine.Engine,
	config WorkerConfig,
	log logger.Logger,
) (worker.Worker, error) {

	if config.TaskQueue == "" {
		config.TaskQueue = "custom-workflows"
	}

	log.InfoContext(ctx, "Starting Temporal worker",
		logger.Fields{
			"task_queue":          config.TaskQueue,
			"max_concurrent_wfs":  config.MaxConcurrentWFs,
			"max_concurrent_tasks": config.MaxConcurrentTasks,
		},
	)

	// Create worker with options
	w := worker.New(temporalClient, config.TaskQueue, worker.Options{
		MaxConcurrentWorkflowTaskPollers:       10,
		MaxConcurrentActivityTaskPollers:       10,
		MaxConcurrentWorkflowTaskExecutionSize: config.MaxConcurrentWFs,
		MaxConcurrentActivityExecutionSize:     config.MaxConcurrentTasks,
	})

	// Register workflows
	w.RegisterWorkflow(CustomWorkflowExecution)
	w.RegisterWorkflow(ExecuteBranchWorkflow)

	// Register activities
	activities := NewActivities(eng, store)
	w.RegisterActivity(activities.RecordStepStartActivity)
	w.RegisterActivity(activities.RecordStepCompletionActivity)
	w.RegisterActivity(activities.RecordStepFailureActivity)
	w.RegisterActivity(activities.CompleteWorkflowInstanceActivity)
	w.RegisterActivity(activities.MarkInstanceFailedActivity)
	w.RegisterActivity(activities.ExecuteValidationActivity)
	w.RegisterActivity(activities.ExecuteConditionActivity)
	w.RegisterActivity(activities.CreateUserTaskActivity)
	w.RegisterActivity(activities.ExecuteActionActivity)
	w.RegisterActivity(activities.ExecuteNotificationActivity)
	w.RegisterActivity(activities.ExecuteLoopActivity)
	w.RegisterActivity(activities.ExecuteWebhookActivity)
	w.RegisterActivity(activities.EscalateTaskActivity)

	// Start worker
	if err := w.Start(); err != nil {
		log.ErrorContext(ctx, "Failed to start Temporal worker",
			logger.Fields{"error": err.Error()},
		)
		return nil, fmt.Errorf("failed to start worker: %w", err)
	}

	log.InfoContext(ctx, "Temporal worker started successfully",
		logger.Fields{"task_queue": config.TaskQueue},
	)

	return w, nil
}

// StopWorker gracefully stops a Temporal worker
func StopWorker(ctx context.Context, w worker.Worker, log logger.Logger) {
	log.InfoContext(ctx, "Stopping Temporal worker")
	w.Stop()
	log.InfoContext(ctx, "Temporal worker stopped")
}
```

---

## API Layer

### Workflow Template Handler

```go
// internal/workflow/handlers/template_handler.go
package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	db "awo.so/db/sqlc"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/workflow/engine"
)

// TemplateHandler handles workflow template operations
type TemplateHandler struct {
	engine *engine.Engine
	store  db.Store
	logger logger.Logger
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(eng *engine.Engine, store db.Store) *TemplateHandler {
	return &TemplateHandler{
		engine: eng,
		store:  store,
		logger: eng.Logger().WithFields(logger.Fields{"handler": "template"}),
	}
}

// RegisterRoutes registers all template routes
func (h *TemplateHandler) RegisterRoutes(api fiber.Router) {
	templates := api.Group("/workflows/templates")

	templates.Get("/", h.ListTemplates)
	templates.Post("/", h.CreateTemplate)
	templates.Get("/:id", h.GetTemplate)
	templates.Put("/:id", h.UpdateTemplate)
	templates.Delete("/:id", h.DeleteTemplate)
	templates.Post("/:id/activate", h.ActivateTemplate)
	templates.Post("/:id/deactivate", h.DeactivateTemplate)
	templates.Post("/:id/duplicate", h.DuplicateTemplate)
	templates.Post("/validate", h.ValidateTemplate)
}

// CreateTemplateRequest represents create template request
type CreateTemplateRequest struct {
	Name               string                 `json:"name" validate:"required"`
	Description        string                 `json:"description"`
	Category           string                 `json:"category"`
	TriggerType        string                 `json:"trigger_type" validate:"required,oneof=event manual scheduled webhook"`
	TriggerEvent       string                 `json:"trigger_event"`
	TriggerConditions  map[string]interface{} `json:"trigger_conditions"`
	Definition         json.RawMessage        `json:"definition" validate:"required"`
	AMISBuilderState   json.RawMessage        `json:"amis_builder_state"`
	AllowedRoles       []string               `json:"allowed_roles"`
	AllowedEntityIDs   []uuid.UUID            `json:"allowed_entity_ids"`
	IsDraft            bool                   `json:"is_draft"`
}

// ListTemplates returns list of workflow templates
// GET /api/workflows/templates
func (h *TemplateHandler) ListTemplates(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get query parameters
	category := c.Query("category")
	triggerType := c.Query("trigger_type")
	isActive := c.QueryBool("is_active", true)
	isDraft := c.QueryBool("is_draft", false)

	params := db.ListWorkflowTemplatesParams{
		Category:    category,
		TriggerType: triggerType,
		IsActive:    isActive,
		IsDraft:     isDraft,
		Limit:       int32(c.QueryInt("limit", 20)),
		Offset:      int32(c.QueryInt("offset", 0)),
	}

	templates, err := h.store.ListWorkflowTemplates(ctx, params)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to list templates",
			logger.Fields{"error": err.Error()},
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list templates")
	}

	return c.JSON(fiber.Map{
		"data":  templates,
		"count": len(templates),
	})
}

// CreateTemplate creates a new workflow template
// POST /api/workflows/templates
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	ctx := c.Context()

	var req CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := shared.ValidateStruct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Create template using engine
	engineReq := engine.CreateTemplateRequest{
		Name:              req.Name,
		Description:       req.Description,
		Category:          req.Category,
		TriggerType:       req.TriggerType,
		TriggerEvent:      req.TriggerEvent,
		TriggerConditions: req.TriggerConditions,
		DefinitionJSON:    req.Definition,
		AMISBuilderState:  req.AMISBuilderState,
		AllowedRoles:      req.AllowedRoles,
		AllowedEntityIDs:  req.AllowedEntityIDs,
		IsDraft:           req.IsDraft,
	}

	template, err := h.engine.CreateTemplate(ctx, engineReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to create template",
			logger.Fields{"error": err.Error()},
		)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	h.logger.InfoContext(ctx, "Template created",
		logger.Fields{"template_id": template.ID},
	)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": template,
	})
}

// GetTemplate returns a single workflow template
// GET /api/workflows/templates/:id
func (h *TemplateHandler) GetTemplate(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	template, err := h.store.GetWorkflowTemplate(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Template not found")
	}

	return c.JSON(fiber.Map{
		"data": template,
	})
}

// UpdateTemplate updates a workflow template
// PUT /api/workflows/templates/:id
func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	var req CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Get existing template
	existing, err := h.store.GetWorkflowTemplate(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Template not found")
	}

	// Update template
	params := db.UpdateWorkflowTemplateParams{
		ID:                id,
		Name:              req.Name,
		Description:       sql.NullString{String: req.Description, Valid: req.Description != ""},
		Category:          sql.NullString{String: req.Category, Valid: req.Category != ""},
		TriggerType:       req.TriggerType,
		TriggerEvent:      sql.NullString{String: req.TriggerEvent, Valid: req.TriggerEvent != ""},
		TriggerConditions: req.TriggerConditions,
		Definition:        req.Definition,
		AMISBuilderState:  req.AMISBuilderState,
		AllowedRoles:      req.AllowedRoles,
		AllowedEntityIds:  req.AllowedEntityIDs,
		IsDraft:           req.IsDraft,
		Version:           existing.Version + 1,
	}

	updated, err := h.store.UpdateWorkflowTemplate(ctx, params)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to update template",
			logger.Fields{"template_id": id, "error": err.Error()},
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update template")
	}

	return c.JSON(fiber.Map{
		"data": updated,
	})
}

// DeleteTemplate soft-deletes a workflow template
// DELETE /api/workflows/templates/:id
func (h *TemplateHandler) DeleteTemplate(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	// Check for running instances
	count, err := h.store.CountRunningInstances(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check running instances")
	}

	if count > 0 {
		return fiber.NewError(fiber.StatusConflict, 
			fmt.Sprintf("Cannot delete template with %d running instances", count))
	}

	// Soft delete
	if err := h.store.DeleteWorkflowTemplate(ctx, id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete template")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ActivateTemplate activates a workflow template
// POST /api/workflows/templates/:id/activate
func (h *TemplateHandler) ActivateTemplate(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	if err := h.store.ActivateTemplate(ctx, id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to activate template")
	}

	return c.JSON(fiber.Map{
		"message": "Template activated successfully",
	})
}

// DeactivateTemplate deactivates a workflow template
// POST /api/workflows/templates/:id/deactivate
func (h *TemplateHandler) DeactivateTemplate(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	if err := h.store.DeactivateTemplate(ctx, id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to deactivate template")
	}

	return c.JSON(fiber.Map{
		"message": "Template deactivated successfully",
	})
}

// DuplicateTemplate creates a copy of a template
// POST /api/workflows/templates/:id/duplicate
func (h *TemplateHandler) DuplicateTemplate(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	// Get original template
	original, err := h.store.GetWorkflowTemplate(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Template not found")
	}

	// Create copy
	params := db.CreateWorkflowTemplateParams{
		Name:              original.Name + " (Copy)",
		Description:       original.Description,
		Category:          original.Category,
		TriggerType:       original.TriggerType,
		TriggerEvent:      original.TriggerEvent,
		TriggerConditions: original.TriggerConditions,
		Definition:        original.Definition,
		AMISBuilderState:  original.AmisBuilderState,
		IsActive:          false,
		IsDraft:           true,
		OwnerID:           getUserIDFromContext(ctx),
		AllowedRoles:      original.AllowedRoles,
		AllowedEntityIds:  original.AllowedEntityIds,
		ParentTemplateID:  sql.NullUUID{UUID: id, Valid: true},
	}

	duplicate, err := h.store.CreateWorkflowTemplate(ctx, params)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to duplicate template")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": duplicate,
	})
}

// ValidateTemplate validates a workflow definition without saving
// POST /api/workflows/templates/validate
func (h *TemplateHandler) ValidateTemplate(c *fiber.Ctx) error {
	ctx := c.Context()

	var req struct {
		Definition json.RawMessage `json:"definition" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Parse definition
	var definition engine.WorkflowDefinition
	if err := json.Unmarshal(req.Definition, &definition); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"valid": false,
			"errors": []string{"Invalid workflow definition JSON"},
		})
	}

	// Validate
	validator := engine.NewValidator()
	if err := validator.Validate(&definition); err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"valid":  false,
			"errors": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"valid":   true,
		"message": "Workflow definition is valid",
	})
}
```

### Workflow Instance Handler

```go
// internal/workflow/handlers/instance_handler.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	db "awo.so/db/sqlc"
	"awo.so/internal/shared/logger"
	"awo.so/internal/workflow/engine"
)

// InstanceHandler handles workflow instance operations
type InstanceHandler struct {
	engine *engine.Engine
	store  db.Store
	logger logger.Logger
}

// NewInstanceHandler creates a new instance handler
func NewInstanceHandler(eng *engine.Engine, store db.Store) *InstanceHandler {
	return &InstanceHandler{
		engine: eng,
		store:  store,
		logger: eng.Logger().WithFields(logger.Fields{"handler": "instance"}),
	}
}

// RegisterRoutes registers all instance routes
func (h *InstanceHandler) RegisterRoutes(api fiber.Router) {
	instances := api.Group("/workflows/instances")

	instances.Get("/", h.ListInstances)
	instances.Post("/start", h.StartWorkflow)
	instances.Get("/:id", h.GetInstance)
	instances.Post("/:id/cancel", h.CancelWorkflow)
	instances.Post("/:id/retry", h.RetryWorkflow)
	instances.Get("/:id/steps", h.GetStepExecutions)
	instances.Get("/:id/variables", h.GetVariables)
}

// StartWorkflowRequest represents start workflow request
type StartWorkflowRequest struct {
	TemplateID uuid.UUID              `json:"template_id" validate:"required"`
	Input      map[string]interface{} `json:"input" validate:"required"`
	Priority   string                 `json:"priority"`
	Tags       []string               `json:"tags"`
}

// ListInstances returns list of workflow instances
// GET /api/workflows/instances
func (h *InstanceHandler) ListInstances(c *fiber.Ctx) error {
	ctx := c.Context()

	params := db.ListWorkflowInstancesParams{
		Status:     c.Query("status"),
		TemplateID: parseUUIDQuery(c.Query("template_id")),
		Limit:      int32(c.QueryInt("limit", 20)),
		Offset:     int32(c.QueryInt("offset", 0)),
	}

	instances, err := h.store.ListWorkflowInstances(ctx, params)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list instances")
	}

	return c.JSON(fiber.Map{
		"data":  instances,
		"count": len(instances),
	})
}

// StartWorkflow starts a new workflow instance
// POST /api/workflows/instances/start
func (h *InstanceHandler) StartWorkflow(c *fiber.Ctx) error {
	ctx := c.Context()

	var req StartWorkflowRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := shared.ValidateStruct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Start workflow
	instance, err := h.engine.StartWorkflow(ctx, req.TemplateID, req.Input)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to start workflow",
			logger.Fields{"template_id": req.TemplateID, "error": err.Error()},
		)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	h.logger.InfoContext(ctx, "Workflow started",
		logger.Fields{"instance_id": instance.ID, "template_id": req.TemplateID},
	)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": instance,
	})
}

// GetInstance returns a single workflow instance
// GET /api/workflows/instances/:id
func (h *InstanceHandler) GetInstance(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid instance ID")
	}

	instance, err := h.store.GetWorkflowInstance(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Instance not found")
	}

	return c.JSON(fiber.Map{
		"data": instance,
	})
}

// CancelWorkflow cancels a running workflow
// POST /api/workflows/instances/:id/cancel
func (h *InstanceHandler) CancelWorkflow(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid instance ID")
	}

	// Get instance
	instance, err := h.store.GetWorkflowInstance(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Instance not found")
	}

	// Cancel in Temporal
	err = h.engine.TemporalClient().CancelWorkflow(ctx, instance.TemporalWorkflowID, instance.TemporalRunID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to cancel workflow")
	}

	// Update status
	err = h.store.CancelWorkflowInstance(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update status")
	}

	return c.JSON(fiber.Map{
		"message": "Workflow cancelled successfully",
	})
}

// RetryWorkflow retries a failed workflow
// POST /api/workflows/instances/:id/retry
func (h *InstanceHandler) RetryWorkflow(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid instance ID")
	}

	instance, err := h.store.GetWorkflowInstance(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Instance not found")
	}

	if instance.Status != "failed" {
		return fiber.NewError(fiber.StatusBadRequest, "Only failed workflows can be retried")
	}

	// Start new instance with same input
	var input map[string]interface{}
	if err := json.Unmarshal(instance.TriggerData, &input); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse trigger data")
	}

	newInstance, err := h.engine.StartWorkflow(ctx, instance.TemplateID, input)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": newInstance,
	})
}

// GetStepExecutions returns step executions for an instance
// GET /api/workflows/instances/:id/steps
func (h *InstanceHandler) GetStepExecutions(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid instance ID")
	}

	steps, err := h.store.GetStepExecutions(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get step executions")
	}

	return c.JSON(fiber.Map{
		"data":  steps,
		"count": len(steps),
	})
}

// GetVariables returns workflow variables
// GET /api/workflows/instances/:id/variables
func (h *InstanceHandler) GetVariables(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid instance ID")
	}

	variables, err := h.store.GetWorkflowVariables(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get variables")
	}

	return c.JSON(fiber.Map{
		"data":  variables,
		"count": len(variables),
	})
}

// Helper to parse UUID from query string
func parseUUIDQuery(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, _ := uuid.Parse(s)
	return id
}
```

### User Task Handler

```go
// internal/workflow/handlers/task_handler.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	db "awo.so/db/sqlc"
	"awo.so/internal/shared/logger"
	"awo.so/internal/workflow/engine"
)

// TaskHandler handles user task operations
type TaskHandler struct {
	engine *engine.Engine
	store  db.Store
	logger logger.Logger
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(eng *engine.Engine, store db.Store) *TaskHandler {
	return &TaskHandler{
		engine: eng,
		store:  store,
		logger: eng.Logger().WithFields(logger.Fields{"handler": "task"}),
	}
}

// RegisterRoutes registers all task routes
func (h *TaskHandler) RegisterRoutes(api fiber.Router) {
	tasks := api.Group("/workflows/tasks")

	tasks.Get("/my-tasks", h.GetMyTasks)
	tasks.Get("/:id", h.GetTask)
	tasks.Post("/:id/complete", h.CompleteTask)
	tasks.Post("/:id/delegate", h.DelegateTask)
	tasks.Post("/:id/reassign", h.ReassignTask)
}

// CompleteTaskRequest represents complete task request
type CompleteTaskRequest struct {
	Decision string                 `json:"decision" validate:"required,oneof=approved rejected delegated cancelled"`
	Comment  string                 `json:"comment"`
	FormData map[string]interface{} `json:"form_data"`
}

// GetMyTasks returns tasks assigned to current user
// GET /api/workflows/tasks/my-tasks
func (h *TaskHandler) GetMyTasks(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := getUserIDFromContext(ctx)

	filters := engine.TaskFilters{
		Status: c.Query("status", "pending"),
		Limit:  int32(c.QueryInt("limit", 20)),
		Offset: int32(c.QueryInt("offset", 0)),
	}

	tasks, err := h.engine.GetUserTasks(ctx, userID, filters)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tasks")
	}

	return c.JSON(fiber.Map{
		"data":  tasks,
		"count": len(tasks),
	})
}

// GetTask returns a single task
// GET /api/workflows/tasks/:id
func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	task, err := h.store.GetUserTask(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Task not found")
	}

	return c.JSON(fiber.Map{
		"data": task,
	})
}

// CompleteTask completes a user task
// POST /api/workflows/tasks/:id/complete
func (h *TaskHandler) CompleteTask(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := getUserIDFromContext(ctx)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	var req CompleteTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := shared.ValidateStruct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Complete task
	decision := engine.TaskDecision{
		Decision:    req.Decision,
		Comment:     req.Comment,
		FormData:    req.FormData,
		CompletedBy: userID,
	}

	err = h.engine.CompleteUserTask(ctx, id, decision)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to complete task",
			logger.Fields{"task_id": id, "error": err.Error()},
		)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	h.logger.InfoContext(ctx, "Task completed",
		logger.Fields{"task_id": id, "decision": req.Decision},
	)

	return c.JSON(fiber.Map{
		"message": "Task completed successfully",
	})
}

// DelegateTask delegates a task to another user
// POST /api/workflows/tasks/:id/delegate
func (h *TaskHandler) DelegateTask(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	var req struct {
		DelegateToUserID uuid.UUID `json:"delegate_to_user_id" validate:"required"`
		Comment          string    `json:"comment"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Delegate task
	err = h.store.DelegateUserTask(ctx, db.DelegateUserTaskParams{
		ID:               id,
		DelegateToUserID: req.DelegateToUserID,
		Comment:          sql.NullString{String: req.Comment, Valid: req.Comment != ""},
	})

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delegate task")
	}

	return c.JSON(fiber.Map{
		"message": "Task delegated successfully",
	})
}

// ReassignTask reassigns a task to another user
// POST /api/workflows/tasks/:id/reassign
func (h *TaskHandler) ReassignTask(c *fiber.Ctx) error {
	ctx := c.Context()

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	var req struct {
		ReassignToUserID uuid.UUID `json:"reassign_to_user_id" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err = h.store.ReassignUserTask(ctx, db.ReassignUserTaskParams{
		ID:               id,
		AssignedToUserID: sql.NullUUID{UUID: req.ReassignToUserID, Valid: true},
	})

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to reassign task")
	}

	return c.JSON(fiber.Map{
		"message": "Task reassigned successfully",
	})
}
```

### Main Router Setup

```go
// internal/workflow/service.go
package workflow

import (
	"context"

	"github.com/gofiber/fiber/v2"
	db "awo.so/db/sqlc"
	"awo.so/internal/shared/logger"
	"awo.so/internal/workflow/engine"
	"awo.so/internal/workflow/handlers"
	"awo.so/internal/workflow/temporal"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Service represents the workflow service
type Service struct {
	engine         *engine.Engine
	store          db.Store
	temporalClient temporalclient.Client
	temporalWorker worker.Worker
	logger         logger.Logger

	// Handlers
	templateHandler *handlers.TemplateHandler
	instanceHandler *handlers.InstanceHandler
	taskHandler     *handlers.TaskHandler
}

// Config holds service configuration
type Config struct {
	Store              db.Store
	TemporalClient     temporalclient.Client
	Logger             logger.Logger
	WorkerConfig       temporal.WorkerConfig
	MaxConcurrentTasks int
	DefaultTimeout     time.Duration
}

// NewService creates a new workflow service
func NewService(ctx context.Context, cfg Config) (*Service, error) {
	if cfg.Logger == nil {
		cfg.Logger = logger.WithFields(logger.Fields{"service": "workflow"})
	}

	// Create engine
	engineConfig := engine.Config{
		Store:              cfg.Store,
		TemporalClient:     cfg.TemporalClient,
		Logger:             cfg.Logger,
		MaxConcurrentTasks: cfg.MaxConcurrentTasks,
		DefaultTimeout:     cfg.DefaultTimeout,
	}

	eng, err := engine.NewEngine(engineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}

	// Start Temporal worker
	worker, err := temporal.StartWorker(
		ctx,
		cfg.TemporalClient,
		cfg.Store,
		eng,
		cfg.WorkerConfig,
		cfg.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start temporal worker: %w", err)
	}

	service := &Service{
		engine:          eng,
		store:           cfg.Store,
		temporalClient:  cfg.TemporalClient,
		temporalWorker:  worker,
		logger:          cfg.Logger,
		templateHandler: handlers.NewTemplateHandler(eng, cfg.Store),
		instanceHandler: handlers.NewInstanceHandler(eng, cfg.Store),
		taskHandler:     handlers.NewTaskHandler(eng, cfg.Store),
	}

	cfg.Logger.InfoContext(ctx, "Workflow service initialized successfully")

	return service, nil
}

// RegisterRoutes registers all workflow routes
func (s *Service) RegisterRoutes(api fiber.Router) {
	s.logger.Info("Registering workflow routes")

	// Register handler routes
	s.templateHandler.RegisterRoutes(api)
	s.instanceHandler.RegisterRoutes(api)
	s.taskHandler.RegisterRoutes(api)
}

// Shutdown gracefully shuts down the service
func (s *Service) Shutdown(ctx context.Context) error {
	s.logger.InfoContext(ctx, "Shutting down workflow service")

	// Stop Temporal worker
	temporal.StopWorker(ctx, s.temporalWorker, s.logger)

	s.logger.InfoContext(ctx, "Workflow service shutdown complete")
	return nil
}

// Engine returns the workflow engine
func (s *Service) Engine() *engine.Engine {
	return s.engine
}
```

---

## AMIS Frontend Integration

### Workflow Builder Page

```json
{
  "type": "page",
  "title": "Workflow Builder",
  "toolbar": [
    {
      "type": "button",
      "label": "Back to Templates",
      "actionType": "url",
      "url": "/workflows/templates",
      "icon": "fa fa-arrow-left"
    }
  ],
  "body": {
    "type": "wizard",
    "mode": "horizontal",
    "steps": [
      {
        "title": "Basic Information",
        "body": [
          {
            "type": "input-text",
            "name": "name",
            "label": "Workflow Name",
            "placeholder": "e.g., Invoice Approval",
            "required": true,
            "validations": {
              "maxLength": 255
            }
          },
          {
            "type": "textarea",
            "name": "description",
            "label": "Description",
            "placeholder": "What does this workflow do?",
            "maxRows": 3
          },
          {
            "type": "select",
            "name": "category",
            "label": "Category",
            "options": [
              {"label": "Approval", "value": "approval"},
              {"label": "Notification", "value": "notification"},
              {"label": "Data Processing", "value": "data_processing"},
              {"label": "Integration", "value": "integration"},
              {"label": "Compliance", "value": "compliance"}
            ],
            "value": "approval"
          },
          {
            "type": "divider"
          },
          {
            "type": "radios",
            "name": "trigger_type",
            "label": "How should this workflow start?",
            "options": [
              {
                "label": "Manual",
                "value": "manual",
                "description": "User initiates the workflow"
              },
              {
                "label": "Event",
                "value": "event",
                "description": "Automatic on system event"
              },
              {
                "label": "Schedule",
                "value": "scheduled",
                "description": "Time-based trigger"
              },
              {
                "label": "Webhook",
                "value": "webhook",
                "description": "External API trigger"
              }
            ],
            "value": "manual"
          },
          {
            "type": "select",
            "name": "trigger_event",
            "label": "Trigger Event",
            "source": "/api/workflows/events",
            "visibleOn": "${trigger_type === 'event'}",
            "required": true,
            "searchable": true
          },
          {
            "type": "input-text",
            "name": "schedule_expression",
            "label": "Schedule (Cron Expression)",
            "visibleOn": "${trigger_type === 'scheduled'}",
            "placeholder": "0 9 * * *",
            "description": "Example: 0 9 * * * (Daily at 9 AM)"
          }
        ]
      },
      {
        "title": "Variables",
        "body": [
          {
            "type": "alert",
            "level": "info",
            "body": "Variables are data that can be passed into and used throughout the workflow."
          },
          {
            "type": "combo",
            "name": "definition.variables",
            "label": "Workflow Variables",
            "multiple": true,
            "multiLine": true,
            "addable": true,
            "removable": true,
            "draggable": true,
            "items": [
              {
                "type": "input-text",
                "name": "name",
                "label": "Variable Name",
                "placeholder": "e.g., invoice_amount",
                "required": true,
                "validations": {
                  "matchRegexp": "/^[a-z][a-z0-9_]*$/"
                },
                "validationErrors": {
                  "matchRegexp": "Must start with lowercase letter and contain only letters, numbers, and underscores"
                }
              },
              {
                "type": "select",
                "name": "type",
                "label": "Data Type",
                "options": [
                  {"label": "Text (string)", "value": "string"},
                  {"label": "Number", "value": "number"},
                  {"label": "True/False (boolean)", "value": "boolean"},
                  {"label": "Object", "value": "object"},
                  {"label": "Array", "value": "array"}
                ],
                "value": "string",
                "required": true
              },
              {
                "type": "switch",
                "name": "required",
                "label": "Required",
                "description": "Must be provided when starting workflow"
              },
              {
                "type": "input-text",
                "name": "default",
                "label": "Default Value",
                "visibleOn": "${!required}",
                "placeholder": "Optional default value"
              },
              {
                "type": "textarea",
                "name": "description",
                "label": "Description",
                "placeholder": "What is this variable used for?",
                "maxRows": 2
              }
            ]
          }
        ]
      },
      {
        "title": "Design Workflow",
        "body": [
          {
            "type": "alert",
            "level": "warning",
            "body": "Drag and drop steps from the left panel to build your workflow. Connect steps to define the flow.",
            "showIcon": true
          },
          {
            "type": "flex",
            "direction": "row",
            "items": [
              {
                "type": "panel",
                "title": "Available Steps",
                "className": "w-64 m-r-md",
                "body": {
                  "type": "list",
                  "source": "${step_types}",
                  "listItem": {
                    "type": "card",
                    "className": "m-b-sm cursor-pointer hover:bg-gray-100",
                    "draggable": true,
                    "body": [
                      {
                        "type": "flex",
                        "items": [
                          {
                            "type": "icon",
                            "icon": "${icon}",
                            "className": "text-2xl m-r-sm",
                            "style": {"color": "${color}"}
                          },
                          {
                            "type": "container",
                            "body": [
                              {
                                "type": "tpl",
                                "tpl": "<strong>${label}</strong>",
                                "className": "block"
                              },
                              {
                                "type": "tpl",
                                "tpl": "${description}",
                                "className": "text-sm text-gray-600"
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  }
                }
              },
              {
                "type": "container",
                "className": "flex-1 border-l p-l-md",
                "body": {
                  "type": "workflow-canvas",
                  "name": "definition.steps",
                  "height": "600px"
                }
              }
            ]
          }
        ]
      },
      {
        "title": "Review & Publish",
        "body": [
          {
            "type": "service",
            "schemaApi": {
              "method": "post",
              "url": "/api/workflows/templates/validate",
              "data": {
                "definition": "${definition}"
              }
            },
            "body": [
              {
                "type": "alert",
                "level": "${valid ? 'success' : 'danger'}",
                "body": "${valid ? 'Workflow is valid and ready to publish!' : errors}",
                "showIcon": true
              },
              {
                "type": "divider"
              },
              {
                "type": "property",
                "title": "Workflow Summary",
                "column": 2,
                "items": [
                  {"label": "Name", "content": "${name}"},
                  {"label": "Category", "content": "${category}"},
                  {"label": "Trigger Type", "content": "${trigger_type}"},
                  {"label": "Total Steps", "content": "${definition.steps.length}"},
                  {"label": "Variables", "content": "${definition.variables.length}"}
                ]
              },
              {
                "type": "divider"
              },
              {
                "type": "switch",
                "name": "is_draft",
                "label": "Save as Draft",
                "description": "Draft workflows can be edited but not executed",
                "value": false
              }
            ]
          }
        ]
      }
    ],
    "api": {
      "method": "post",
      "url": "/api/workflows/templates",
      "data": {
        "&": "$$"
      }
    },
    "redirect": "/workflows/templates"
  }
}
```

### My Tasks Page

```json
{
  "type": "page",
  "title": "My Tasks",
  "subTitle": "Pending approvals and reviews",
  "body": {
    "type": "crud",
    "syncLocation": false,
    "api": {
      "method": "get",
      "url": "/api/workflows/tasks/my-tasks",
      "adaptor": "return { ...payload, data: payload.data || [] };"
    },
    "interval": 30000,
    "headerToolbar": [
      {
        "type": "tpl",
        "tpl": "<div class='flex items-center'><i class='fa fa-tasks text-xl m-r-sm'></i><span class='text-lg font-bold'>Total: ${count || 0} tasks</span></div>",
        "className": "v-middle"
      },
      "reload"
    ],
    "filter": {
      "title": "Filter Tasks",
      "submitText": "Search",
      "body": [
        {
          "type": "select",
          "name": "status",
          "label": "Status",
          "options": [
            {"label": "All Pending", "value": "pending"},
            {"label": "In Progress", "value": "in_progress"}
          ],
          "value": "pending"
        },
        {
          "type": "select",
          "name": "priority",
          "label": "Priority",
          "options": [
            {"label": "All", "value": ""},
            {"label": "Urgent", "value": "urgent"},
            {"label": "High", "value": "high"},
            {"label": "Normal", "value": "normal"},
            {"label": "Low", "value": "low"}
          ],
          "clearable": true
        },
        {
          "type": "input-text",
          "name": "search",
          "label": "Search",
          "placeholder": "Search by title or description",
          "clearable": true
        }
      ]
    },
    "footerToolbar": ["pagination", "statistics"],
    "perPage": 20,
    "columns": [
      {
        "name": "priority",
        "label": "Priority",
        "type": "mapping",
        "width": 100,
        "map": {
          "urgent": "<span class='label label-danger'>URGENT</span>",
          "high": "<span class='label label-warning'>HIGH</span>",
          "normal": "<span class='label label-info'>NORMAL</span>",
          "low": "<span class='label label-default'>LOW</span>"
        }
      },
      {
        "name": "title",
        "label": "Task",
        "type": "text",
        "searchable": true
      },
      {
        "name": "workflow_name",
        "label": "Workflow",
        "type": "text"
      },
      {
        "name": "description",
        "label": "Description",
        "type": "text",
        "toggled": false
      },
      {
        "name": "assigned_at",
        "label": "Assigned",
        "type": "datetime",
        "format": "fromNow",
        "width": 120
      },
      {
        "name": "due_at",
        "label": "Due",
        "type": "datetime",
        "format": "fromNow",
        "width": 120,
        "classNameExpr": "${DATETOTIME(due_at) < NOW() ? 'text-danger font-bold' : ''}"
      },
      {
        "type": "operation",
        "label": "Actions",
        "width": 150,
        "buttons": [
          {
            "type": "button",
            "label": "Review",
            "level": "primary",
            "size": "sm",
            "actionType": "dialog",
            "dialog": {
              "title": "${title}",
              "size": "lg",
              "closeOnEsc": false,
              "body": {
                "type": "service",
                "api": "/api/workflows/tasks/${id}",
                "body": [
                  {
                    "type": "alert",
                    "level": "info",
                    "body": "${data.description}",
                    "showIcon": true,
                    "className": "m-b-md"
                  },
                  {
                    "type": "divider"
                  },
                  {
                    "type": "grid",
                    "columns": [
                      {
                        "type": "panel",
                        "title": "Task Details",
                        "body": {
                          "type": "property",
                          "column": 1,
                          "items": [
                            {"label": "Workflow", "content": "${data.workflow_name}"},
                            {"label": "Priority", "content": "${data.priority}"},
                            {"label": "Assigned", "content": "${data.assigned_at|date:YYYY-MM-DD HH:mm}"},
                            {"label": "Due", "content": "${data.due_at|date:YYYY-MM-DD HH:mm}"}
                          ]
                        }
                      }
                    ]
                  },
                  {
                    "type": "divider"
                  },
                  {
                    "type": "panel",
                    "title": "Context Information",
                    "body": {
                      "type": "json",
                      "source": "${data.context_data}",
                      "levelExpand": 2
                    }
                  },
                  {
                    "type": "divider"
                  },
                  {
                    "type": "form",
                    "api": {
                      "method": "post",
                      "url": "/api/workflows/tasks/${id}/complete"
                    },
                    "body": [
                      "${data.form_schema}",
                      {
                        "type": "textarea",
                        "name": "comment",
                        "label": "Comment",
                        "placeholder": "Add your comments here...",
                        "minRows": 3,
                        "maxRows": 6,
                        "required": "${data.requires_comment}"
                      },
                      {
                        "type": "hidden",
                        "name": "decision",
                        "value": ""
                      }
                    ],
                    "actions": [
                      {
                        "type": "button",
                        "label": "Approve",
                        "level": "primary",
                        "actionType": "submit",
                        "icon": "fa fa-check",
                        "onClick": "this.props.formStore.setValues({decision: 'approved'})"
                      },
                      {
                        "type": "button",
                        "label": "Reject",
                        "level": "danger",
                        "actionType": "submit",
                        "icon": "fa fa-times",
                        "onClick": "this.props.formStore.setValues({decision: 'rejected'})"
                      },
                      {
                        "type": "button",
                        "label": "Delegate",
                        "level": "default",
                        "actionType": "dialog",
                        "visibleOn": "${data.can_delegate}",
                        "dialog": {
                          "title": "Delegate Task",
                          "body": {
                            "type": "form",
                            "api": {
                              "method": "post",
                              "url": "/api/workflows/tasks/${id}/delegate"
                            },
                            "body": [
                              {
                                "type": "select",
                                "name": "delegate_to_user_id",
                                "label": "Delegate To",
                                "source": "/api/users?role=${data.assigned_to_role}",
                                "required": true
                              },
                              {
                                "type": "textarea",
                                "name": "comment",
                                "label": "Reason",
                                "placeholder": "Why are you delegating this task?"
                              }
                            ]
                          }
                        }
                      }
                    ]
                  }
                ]
              }
            }
          }
        ]
      }
    ]
  }
}
```

### Workflow Instance Tracker

```json
{
  "type": "page",
  "title": "Workflow Execution Details",
  "body": {
    "type": "service",
    "api": "/api/workflows/instances/${instanceId}",
    "body": [
      {
        "type": "panel",
        "title": "${data.template_name}",
        "body": [
          {
            "type": "grid",
            "columns": [
              {
                "type": "panel",
                "title": "Status",
                "body": {
                  "type": "mapping",
                  "map": {
                    "running": {
                      "type": "status",
                      "value": "Running",
                      "icon": "fa fa-spinner fa-spin",
                      "className": "text-info"
                    },
                    "completed": {
                      "type": "status",
                      "value": "Completed",
                      "icon": "fa fa-check-circle",
                      "className": "text-success"
                    },
                    "failed": {
                      "type": "status",
                      "value": "Failed",
                      "icon": "fa fa-times-circle",
                      "className": "text-danger"
                    },
                    "paused": {
                      "type": "status",
                      "value": "Paused",
                      "icon": "fa fa-pause-circle",
                      "className": "text-warning"
                    },
                    "cancelled": {
                      "type": "status",
                      "value": "Cancelled",
                      "icon": "fa fa-ban",
                      "className": "text-muted"
                    }
                  },
                  "source": "${data.status}"
                }
              },
              {
                "type": "panel",
                "title": "Progress",
                "body": {
                  "type": "progress",
                  "value": "${data.completion_percentage}",
                  "showLabel": true,
                  "map": {
                    "*": {"className": "bg-info"},
                    "100": {"className": "bg-success"}
                  }
                }
              },
              {
                "type": "panel",
                "title": "Timing",
                "body": {
                  "type": "property",
                  "column": 1,
                  "items": [
                    {
                      "label": "Started",
                      "content": "${data.started_at|date:YYYY-MM-DD HH:mm:ss}"
                    },
                    {
                      "label": "Duration",
                      "content": "${data.execution_time_ms ? (data.execution_time_ms / 1000) + 's' : 'In progress'}"
                    },
                    {
                      "label": "Current Step",
                      "content": "${data.current_step_name || 'N/A'}"
                    }
                  ]
                }
              }
            ]
          }
        ]
      },
      {
        "type": "divider"
      },
      {
        "type": "tabs",
        "tabs": [
          {
            "title": "Timeline",
            "icon": "fa fa-stream",
            "body": {
              "type": "service",
              "api": "/api/workflows/instances/${instanceId}/steps",
              "body": {
                "type": "timeline",
                "source": "${data}",
                "items": [
                  {
                    "time": "${started_at|date:HH:mm:ss}",
                    "title": "${step_name}",
                    "detail": [
                      {
                        "type": "tpl",
                        "tpl": "<span class='label ${status === \"completed\" ? \"label-success\" : status === \"failed\" ? \"label-danger\" : status === \"running\" ? \"label-info\" : \"label-default\"}'>${status}</span>"
                      },
                      {
                        "type": "tpl",
                        "tpl": "<span class='text-muted m-l-sm'>${duration_ms ? (duration_ms + 'ms') : ''}</span>"
                      }
                    ],
                    "color": "${status === 'completed' ? '#52c41a' : status === 'failed' ? '#f5222d' : status === 'running' ? '#1890ff' : '#d9d9d9'}"
                  }
                ]
              }
            }
          },
          {
            "title": "Step Details",
            "icon": "fa fa-list",
            "body": {
              "type": "crud",
              "api": "/api/workflows/instances/${instanceId}/steps",
              "syncLocation": false,
              "columns": [
                {
                  "name": "step_order",
                  "label": "#",
                  "width": 60
                },
                {
                  "name": "step_name",
                  "label": "Step",
                  "type": "text"
                },
                {
                  "name": "step_type",
                  "label": "Type",
                  "type": "tag",
                  "width": 120
                },
                {
                  "name": "status",
                  "label": "Status",
                  "type": "mapping",
                  "width": 120,
                  "map": {
                    "completed": "<span class='label label-success'>Completed</span>",
                    "running": "<span class='label label-info'>Running</span>",
                    "failed": "<span class='label label-danger'>Failed</span>",
                    "waiting": "<span class='label label-warning'>Waiting</span>",
                    "pending": "<span class='label label-default'>Pending</span>"
                  }
                },
                {
                  "name": "duration_ms",
                  "label": "Duration",
                  "type": "text",
                  "width": 100,
                  "tpl": "${duration_ms ? (duration_ms + 'ms') : '-'}"
                },
                {
                  "name": "started_at",
                  "label": "Started",
                  "type": "datetime",
                  "format": "HH:mm:ss",
                  "width": 120
                },
                {
                  "type": "operation",
                  "label": "Actions",
                  "width": 100,
                  "buttons": [
                    {
                      "type": "button",
                      "label": "Details",
                      "level": "link",
                      "actionType": "dialog",
                      "dialog": {
                        "title": "Step Execution: ${step_name}",
                        "size": "lg",
                        "body": {
                          "type": "tabs",
                          "tabs": [
                            {
                              "title": "Input",
                              "body": {
                                "type": "json",
                                "source": "${input_data}",
                                "levelExpand": 2
                              }
                            },
                            {
                              "title": "Output",
                              "body": {
                                "type": "json",
                                "source": "${output_data}",
                                "levelExpand": 2
                              }
                            },
                            {
                              "title": "Error",
                              "visibleOn": "${error_message}",
                              "body": {
                                "type": "alert",
                                "level": "danger",
                                "body": "${error_message}",
                                "showIcon": true
                              }
                            }
                          ]
                        }
                      }
                    }
                  ]
                }
              ]
            }
          },
          {
            "title": "Variables",
            "icon": "fa fa-code",
            "body": {
              "type": "service",
              "api": "/api/workflows/instances/${instanceId}/variables",
              "body": {
                "type": "json",
                "source": "${data}",
                "levelExpand": 2
              }
            }
          },
          {
            "title": "Actions",
            "icon": "fa fa-cog",
            "body": {
              "type": "panel",
              "body": [
                {
                  "type": "button-toolbar",
                  "buttons": [
                    {
                      "type": "button",
                      "label": "Cancel Workflow",
                      "level": "danger",
                      "icon": "fa fa-ban",
                      "actionType": "ajax",
                      "api": {
                        "method": "post",
                        "url": "/api/workflows/instances/${instanceId}/cancel"
                      },
                      "confirmText": "Are you sure you want to cancel this workflow? This action cannot be undone.",
                      "visibleOn": "${data.status === 'running' || data.status === 'paused'}"
                    },
                    {
                      "type": "button",
                      "label": "Retry Workflow",
                      "level": "warning",
                      "icon": "fa fa-redo",
                      "actionType": "ajax",
                      "api": {
                        "method": "post",
                        "url": "/api/workflows/instances/${instanceId}/retry"
                      },
                      "confirmText": "This will start a new instance with the same input. Continue?",
                      "visibleOn": "${data.status === 'failed'}"
                    },
                    {
                      "type": "button",
                      "label": "View in Temporal",
                      "level": "info",
                      "icon": "fa fa-external-link-alt",
                      "actionType": "url",
                      "url": "http://temporal-ui:8080/namespaces/default/workflows/${data.temporal_workflow_id}",
                      "blank": true
                    }
                  ]
                }
              ]
            }
          }
        ]
      }
    ]
  }
}
```

---

## Pre-built Workflow Templates

### 1. Invoice Approval Template

```json
{
  "name": "Standard Invoice Approval",
  "description": "Multi-level invoice approval based on amount thresholds",
  "category": "approval",
  "trigger_type": "event",
  "trigger_event": "invoice.created",
  "is_system": true,
  "definition": {
    "name": "Invoice Approval Workflow",
    "version": 1,
    "variables": [
      {"name": "invoice", "type": "object", "required": true},
      {"name": "approval_chain", "type": "array", "default": []}
    ],
    "steps": [
      {
        "id": "validate_invoice",
        "type": "validation",
        "name": "Validate Invoice",
        "config": {
          "rules": [
            {
              "field": "$.invoice.amount",
              "operator": ">",
              "value": 0,
              "error_message": "Invoice amount must be positive"
            },
            {
              "field": "$.invoice.vendor_id",
              "operator": "exists",
              "error_message": "Vendor is required"
            }
          ]
        },
        "on_success": "check_amount_tier",
        "on_failure": "notify_creator_error"
      },
      {
        "id": "check_amount_tier",
        "type": "condition",
        "name": "Determine Approval Level",
        "config": {
          "condition": "$.invoice.amount > 50000"
        },
        "on_true": "require_cfo_approval",
        "on_false": "check_director_threshold"
      },
      {
        "id": "check_director_threshold",
        "type": "condition",
        "name": "Check Director Threshold",
        "config": {
          "condition": "$.invoice.amount > 10000"
        },
        "on_true": "require_director_approval",
        "on_false": "require_manager_approval"
      },
      {
        "id": "require_manager_approval",
        "type": "user_task",
        "name": "Manager Approval",
        "config": {
          "title": "Approve Invoice ${invoice.invoice_number}",
          "description": "Invoice from ${invoice.vendor_name} for $${invoice.amount}",
          "priority": "normal",
          "assign_to_role": "manager",
          "entity_scope": "$.invoice.entity_id",
          "timeout": "48h",
          "timeout_action": "escalate",
          "escalate_to_role": "director"
        },
        "on_approved": "notify_approval",
        "on_rejected": "notify_rejection"
      }
    ]
  }
}
```

### 2. Employee Onboarding Template

```json
{
  "name": "Employee Onboarding",
  "description": "Automated new employee setup process",
  "category": "data_processing",
  "trigger_type": "manual",
  "definition": {
    "name": "Employee Onboarding Workflow",
    "version": 1,
    "variables": [
      {"name": "employee", "type": "object", "required": true},
      {"name": "start_date", "type": "string", "required": true}
    ],
    "steps": [
      {
        "id": "validate_employee",
        "type": "validation",
        "name": "Validate Employee Data",
        "config": {
          "rules": [
            {"field": "$.employee.email", "operator": "regex", "value": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"},
            {"field": "$.employee.department", "operator": "exists"}
          ]
        },
        "on_success": "parallel_setup",
        "on_failure": "notify_hr_error"
      },
      {
        "id": "parallel_setup",
        "type": "parallel",
        "name": "Setup Accounts & Equipment",
        "config": {
          "branches": [
            {
              "id": "it_setup",
              "name": "IT Setup",
              "steps": ["create_email_account", "create_slack_account", "assign_laptop"]
            },
            {
              "id": "hr_setup",
              "name": "HR Setup",
              "steps": ["create_hr_record", "setup_payroll", "assign_benefits"]
            },
            {
              "id": "facilities_setup",
              "name": "Facilities Setup",
              "steps": ["assign_desk", "issue_access_badge"]
            }
          ],
          "wait_for": "all"
        },
        "on_complete": "schedule_orientation"
      },
      {
        "id": "schedule_orientation",
        "type": "action",
        "name": "Schedule Orientation",
        "config": {
          "action": "create_calendar_event",
          "params": {
            "title": "New Employee Orientation",
            "date": "$.start_date",
            "attendees": ["$.employee.email", "hr@company.com"]
          }
        },
        "on_success": "notify_manager"
      }
    ]
  }
}
```

### 3. Document Review Template

```json
{
  "name": "Document Review & Approval",
  "description": "Multi-stage document review process",
  "category": "approval",
  "trigger_type": "manual",
  "definition": {
    "name": "Document Review Workflow",
    "version": 1,
    "variables": [
      {"name": "document", "type": "object", "required": true},
      {"name": "reviewers", "type": "array", "required": true}
    ],
    "steps": [
      {
        "id": "loop_reviewers",
        "type": "loop",
        "name": "Sequential Review",
        "config": {
          "collection": "$.reviewers",
          "iterator_name": "reviewer",
          "steps": ["request_review"],
          "continue_on_error": false
        },
        "on_complete": "final_approval"
      },
      {
        "id": "request_review",
        "type": "user_task",
        "name": "Document Review",
        "config": {
          "title": "Review ${document.title}",
          "assign_to_user_id": "$.reviewer.user_id",
          "timeout": "72h",
          "form_fields": [
            {
              "name": "feedback",
              "type": "textarea",
              "label": "Review Comments"
            }
          ]
        },
        "on_approved": "next",
        "on_rejected": "notify_rejection"
      }
    ]
  }
}
```

---

## Service Integration Guide

### How External Services Interact with Workflow Engine

The workflow engine is designed to be consumed by other services in your ERP system. Here's how different parts of your application interact with it:

### 1. Import Structure

```go
// In any service that needs workflow functionality
package invoiceservice

import (
    "context"
    "github.com/google/uuid"
    
    db "awo.so/db/sqlc"
    "awo.so/internal/workflow"
    "awo.so/internal/workflow/engine"
)

type InvoiceService struct {
    store           db.Store
    workflowService *workflow.Service
}

func NewInvoiceService(store db.Store, wfService *workflow.Service) *InvoiceService {
    return &InvoiceService{
        store:           store,
        workflowService: wfService,
    }
}
```

### 2. Starting Workflows from External Services

```go
// Example: Invoice service starting approval workflow
func (s *InvoiceService) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
    // Create invoice in database
    invoice, err := s.store.CreateInvoice(ctx, db.CreateInvoiceParams{
        VendorID:      req.VendorID,
        Amount:        req.Amount,
        InvoiceNumber: req.InvoiceNumber,
        EntityID:      req.EntityID,
    })
    if err != nil {
        return nil, err
    }

    // Start workflow if amount exceeds threshold
    if invoice.Amount > 1000 {
        // Find the invoice approval template
        template, err := s.findApprovalTemplate(ctx, "invoice_approval")
        if err != nil {
            // Log error but don't fail invoice creation
            s.logger.ErrorContext(ctx, "Failed to find approval template", 
                logger.Fields{"error": err.Error()})
            return invoice, nil
        }

        // Prepare workflow input
        input := map[string]interface{}{
            "invoice": map[string]interface{}{
                "id":             invoice.ID.String(),
                "invoice_number": invoice.InvoiceNumber,
                "amount":         invoice.Amount,
                "vendor_id":      invoice.VendorID.String(),
                "vendor_name":    invoice.VendorName,
                "entity_id":      invoice.EntityID.String(),
                "created_by":     getUserIDFromContext(ctx).String(),
                "created_by_email": getUserEmailFromContext(ctx),
            },
        }

        // Start workflow
        instance, err := s.workflowService.Engine().StartWorkflow(ctx, template.ID, input)
        if err != nil {
            s.logger.ErrorContext(ctx, "Failed to start workflow", 
                logger.Fields{"invoice_id": invoice.ID, "error": err.Error()})
            // Don't fail - workflow can be started manually later
        } else {
            // Update invoice with workflow instance ID
            _ = s.store.UpdateInvoiceWorkflowInstance(ctx, db.UpdateInvoiceWorkflowInstanceParams{
                ID:                 invoice.ID,
                WorkflowInstanceID: sql.NullUUID{UUID: instance.ID, Valid: true},
            })

            s.logger.InfoContext(ctx, "Workflow started for invoice",
                logger.Fields{
                    "invoice_id":  invoice.ID,
                    "workflow_id": instance.ID,
                })
        }
    }

    return invoice, nil
}

func (s *InvoiceService) findApprovalTemplate(ctx context.Context, name string) (*db.WorkflowTemplate, error) {
    // Query for active template by name or category
    templates, err := s.store.ListWorkflowTemplates(ctx, db.ListWorkflowTemplatesParams{
        Category: "approval",
        IsActive: true,
        Limit:    1,
    })
    if err != nil {
        return nil, err
    }
    if len(templates) == 0 {
        return nil, fmt.Errorf("no approval template found")
    }
    return &templates[0], nil
}
```

### 3. Listening for Workflow Events

```go
// Example: Payment service listening for workflow completion
package paymentservice

import (
    "context"
    "encoding/json"
    
    db "awo.so/db/sqlc"
)

type PaymentService struct {
    store db.Store
}

// This would be called by an event handler/webhook
func (s *PaymentService) HandleWorkflowCompletion(ctx context.Context, instanceID uuid.UUID) error {
    // Get workflow instance
    instance, err := s.store.GetWorkflowInstance(ctx, instanceID)
    if err != nil {
        return err
    }

    // Parse output data
    var output map[string]interface{}
    if err := json.Unmarshal(instance.OutputData, &output); err != nil {
        return err
    }

    // Extract invoice ID from workflow output
    invoiceID, ok := output["invoice_id"].(string)
    if !ok {
        return fmt.Errorf("invoice_id not found in workflow output")
    }

    decision, _ := output["decision"].(string)

    if decision == "approved" {
        // Create payment record
        _, err := s.store.CreatePayment(ctx, db.CreatePaymentParams{
            InvoiceID:  uuid.MustParse(invoiceID),
            Amount:     getAmount(output),
            ApprovedBy: getApprovedBy(output),
            Status:     "pending",
        })
        if err != nil {
            return fmt.Errorf("failed to create payment: %w", err)
        }

        s.logger.InfoContext(ctx, "Payment created from workflow",
            logger.Fields{"invoice_id": invoiceID, "workflow_id": instanceID})
    }

    return nil
}
```

### 4. Main Application Setup

```go
// cmd/server/main.go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gofiber/fiber/v2"
    db "awo.so/db/sqlc"
    "awo.so/internal/workflow"
    "awo.so/internal/shared/logger"
    temporalclient "go.temporal.io/sdk/client"
)

func main() {
    ctx := context.Background()
    log := logger.New()

    // Initialize database
    store, err := db.NewDB(os.Getenv("DATABASE_URL"))
    if err != nil {
        log.FatalContext(ctx, "Failed to connect to database", 
            logger.Fields{"error": err.Error()})
    }
    defer store.Close()

    // Initialize Temporal client
    temporalClient, err := temporalclient.Dial(temporalclient.Options{
        HostPort: os.Getenv("TEMPORAL_HOST"),
    })
    if err != nil {
        log.FatalContext(ctx, "Failed to create Temporal client",
            logger.Fields{"error": err.Error()})
    }
    defer temporalClient.Close()

    // Initialize workflow service
    workflowConfig := workflow.Config{
        Store:          store,
        TemporalClient: temporalClient,
        Logger:         log,
        WorkerConfig: workflow.DefaultWorkerConfig(),
    }

    workflowService, err := workflow.NewService(ctx, workflowConfig)
    if err != nil {
        log.FatalContext(ctx, "Failed to create workflow service",
            logger.Fields{"error": err.Error()})
    }

    // Initialize Fiber app
    app := fiber.New(fiber.Config{
        ErrorHandler: customErrorHandler,
    })

    // Setup middleware
    app.Use(tenantMiddleware(store))
    app.Use(authMiddleware())

    // Register routes
    api := app.Group("/api")
    
    // Register workflow routes
    workflowService.RegisterRoutes(api)

    // Register other service routes
    // invoiceService.RegisterRoutes(api)
    // paymentService.RegisterRoutes(api)
    // etc.

    // Start server
    go func() {
        if err := app.Listen(":8080"); err != nil {
            log.FatalContext(ctx, "Failed to start server",
                logger.Fields{"error": err.Error()})
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.InfoContext(ctx, "Shutting down server...")

    // Shutdown workflow service
    if err := workflowService.Shutdown(ctx); err != nil {
        log.ErrorContext(ctx, "Error shutting down workflow service",
            logger.Fields{"error": err.Error()})
    }

    // Shutdown Fiber
    if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
        log.ErrorContext(ctx, "Error shutting down server",
            logger.Fields{"error": err.Error()})
    }

    log.InfoContext(ctx, "Server stopped")
}
```

### 5. Creating Custom Actions

Services can register custom actions that workflows can execute:

```go
// internal/actions/payment_actions.go
package actions

import (
    "context"
    
    db "awo.so/db/sqlc"
    "awo.so/internal/workflow/engine"
)

// RegisterPaymentActions registers payment-related actions
func RegisterPaymentActions(eng *engine.Engine, store db.Store) {
    // Register "create_payment" action
    eng.ActionRegistry().Register("create_payment", &CreatePaymentAction{
        store: store,
    })

    // Register "cancel_payment" action
    eng.ActionRegistry().Register("cancel_payment", &CancelPaymentAction{
        store: store,
    })
}

type CreatePaymentAction struct {
    store db.Store
}

func (a *CreatePaymentAction) Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    // Extract parameters
    invoiceID := uuid.MustParse(params["invoice_id"].(string))
    amount := params["amount"].(float64)
    approvedBy := uuid.MustParse(params["approved_by"].(string))

    // Create payment
    payment, err := a.store.CreatePayment(ctx, db.CreatePaymentParams{
        InvoiceID:  invoiceID,
        Amount:     amount,
        ApprovedBy: sql.NullUUID{UUID: approvedBy, Valid: true},
        Status:     "pending",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create payment: %w", err)
    }

    // Return result
    return map[string]interface{}{
        "payment_id":     payment.ID.String(),
        "payment_status": payment.Status,
        "created_at":     payment.CreatedAt,
    }, nil
}
```

### 6. Event-Driven Workflow Triggers

```go
// internal/events/workflow_triggers.go
package events

import (
    "context"
    
    db "awo.so/db/sqlc"
    "awo.so/internal/workflow"
)

type EventHandler struct {
    workflowService *workflow.Service
    store           db.Store
}

// HandleInvoiceCreated triggers workflows on invoice creation
func (h *EventHandler) HandleInvoiceCreated(ctx context.Context, invoice *Invoice) error {
    // Find active triggers for this event
    triggers, err := h.store.GetActiveTriggersForEvent(ctx, "invoice.created")
    if err != nil {
        return err
    }

    for _, trigger := range triggers {
        // Check trigger conditions
        if h.evaluateTriggerConditions(trigger, invoice) {
            // Prepare input
            input := map[string]interface{}{
                "invoice": invoice,
            }

            // Start workflow
            _, err := h.workflowService.Engine().StartWorkflow(ctx, trigger.TemplateID, input)
            if err != nil {
                // Log error but continue with other triggers
                log.ErrorContext(ctx, "Failed to start workflow",
                    logger.Fields{
                        "trigger_id": trigger.ID,
                        "error":      err.Error(),
                    })
            }
        }
    }

    return nil
}

func (h *EventHandler) evaluateTriggerConditions(trigger db.WorkflowTrigger, data interface{}) bool {
    // Evaluate trigger conditions using JSONPath
    // Return true if all conditions match
    return true // Simplified
}
```

### Directory Structure

```
your-erp-project/
├── cmd/
│   └── server/
│       └── main.go                    # Application entry point
├── internal/
│   ├── workflow/                      # Workflow engine (this package)
│   │   ├── engine/
│   │   ├── executors/
│   │   ├── temporal/
│   │   ├── handlers/
│   │   └── service.go
│   ├── invoice/                       # Invoice service
│   │   ├── service.go
│   │   └── handler.go
│   ├── payment/                       # Payment service
│   │   ├── service.go
│   │   └── handler.go
│   ├── actions/                       # Custom workflow actions
│   │   ├── payment_actions.go
│   │   └── notification_actions.go
│   └── events/                        # Event handlers
│       └── workflow_triggers.go
├── db/
│   └── sqlc/                          # Database layer
│       ├── store.go                   # Your existing store
│       └── queries.sql.go
└── migrations/
    └── 001_workflow_tables.sql        # Workflow schema
```

---

## Security & Performance

### Security Considerations

#### 1. Row-Level Security (RLS)

All workflow tables have RLS enabled. The tenant context **must** be set for every database operation:

```go
// Always use tenant-aware methods
err := store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
    // All queries here automatically filtered by tenant
    template, err := store.GetWorkflowTemplate(ctx, templateID)
    return err
})
```

#### 2. Input Validation

**Always validate workflow definitions before execution:**

```go
// Validate workflow definition
validator := engine.NewValidator()
if err := validator.Validate(&definition); err != nil {
    return fmt.Errorf("invalid workflow: %w", err)
}
```

**Validation checks:**
- No circular dependencies
- All referenced steps exist
- Expression syntax is valid
- Resource limits not exceeded (max steps: 100)

#### 3. Expression Injection Prevention

**There is no JavaScript evaluation path.** The engine uses two safe mechanisms only:

- `pkg/condition/builder.go` with `expr-lang` for condition/formula evaluation — compiled at validation time, type-checked, no I/O or side effects permitted
- JSONPath adapter for field access — read-only, traverses in-memory map only

```go
// ❌ NEVER add a JavaScript / Lua / arbitrary eval path
result := vm.RunScript(expression)

// ✅ Condition evaluation — sandboxed, compiled, no arbitrary code
ok, err := evaluator.Evaluate(ctx, conditionGroup, evalCtx)

// ✅ Field access — read-only JSONPath over in-memory context
val, err := expressionEval.Evaluate("$.invoice.amount", wfCtx)
```

The `ConditionRule.If` and `ConditionGroup.If` formula fields use `expr-lang` which is compiled at template-save time. Formulas that fail type-checking are rejected before storage.

#### 4. SSRF Protection for HTTP Action Steps

HTTP action steps (`action_type: "http_request"`) can be configured to call external endpoints. Without controls, this allows internal service probing.

**Required mitigations:**

```go
// action executor must validate endpoint before dispatch
func validateHTTPEndpoint(endpoint string) error {
    u, err := url.Parse(endpoint)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }
    // Block private/loopback ranges
    ip := net.ParseIP(u.Hostname())
    if ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()) {
        return fmt.Errorf("SSRF: private/loopback addresses not allowed")
    }
    // Require allowlist — only pre-approved external domains
    if !isAllowlisted(u.Hostname()) {
        return fmt.Errorf("SSRF: host %s not in allowlist", u.Hostname())
    }
    return nil
}
```

The `workflow_action_definitions.http_endpoint` field should be validated at **action definition creation time**, not only at execution time.

#### 5. Webhook Secret Verification

Incoming webhook requests must be verified using HMAC-SHA256:

```go
// Verify webhook signature (constant-time compare prevents timing attacks)
func verifyWebhookSignature(payload []byte, signature string, storedHash string) bool {
    // storedHash is HMAC-SHA256(secret, stored as hex) — never the raw secret
    mac := hmac.New(sha256.New, []byte(storedHash))
    mac.Write(payload)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

**Storage rule:** `workflow_triggers.webhook_secret` stores the HMAC key **hashed** (bcrypt or SHA-256 with salt). The raw secret is shown to the user once at creation and never stored recoverable.

#### 4. Action Authorization

**Check permissions before executing actions:**

```go
type Action struct {
    MinimumRole     string
    AllowedEntities []uuid.UUID
}

func (a *ActionExecutor) Execute(ctx context.Context, step StepDefinition, wfCtx WorkflowContext) (*StepResult, error) {
    // Check if user has required role
    if !hasRole(ctx, action.MinimumRole) {
        return nil, ErrUnauthorized
    }
    
    // Execute action
    // ...
}
```

#### 5. Temporal Namespace Isolation

**Use separate namespaces per tenant for enterprise plans:**

```go
temporalClient, err := temporalclient.Dial(temporalclient.Options{
    HostPort:  "temporal:7233",
    Namespace: fmt.Sprintf("tenant-%s", tenantID),
})
```

### Performance Optimization

#### 1. Database Indexing

**Critical indexes are already defined in schema:**
- `idx_workflow_instances_tenant_status_started` - Dashboard queries
- `idx_workflow_user_tasks_my_tasks` - Task queue queries
- `idx_workflow_step_executions_instance` - Step lookup

**Monitor slow queries:**

```sql
-- Enable slow query logging
ALTER SYSTEM SET log_min_duration_statement = 1000; -- Log queries > 1s

-- Identify missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE schemaname = 'public'
  AND tablename LIKE 'workflow_%'
ORDER BY correlation DESC;
```

#### 2. Connection Pooling

**Configure pgx pool appropriately:**

```go
config := db.DefaultDBConfig()
config.MaxConns = 25        // Max connections
config.MinConns = 5         // Min idle connections
config.ConnectTimeout = 10 * time.Second
config.HealthTimeout = 5 * time.Second

store, err := db.NewDBWithConfig(databaseURL, config)
```

#### 3. Temporal Worker Scaling

**Scale workers based on load:**

```go
workerConfig := temporal.WorkerConfig{
    TaskQueue:          "custom-workflows",
    MaxConcurrentWFs:   50,  // Concurrent workflows
    MaxConcurrentTasks: 100, // Concurrent activities
}

// Run multiple workers for high load
for i := 0; i < 3; i++ {
    worker, _ := temporal.StartWorker(ctx, temporalClient, store, engine, workerConfig, log)
    workers = append(workers, worker)
}
```

#### 4. Caching Strategy

**Cache frequently accessed data:**

```go
type CachedEngine struct {
    *engine.Engine
    templateCache *cache.Cache
}

func (e *CachedEngine) GetTemplate(ctx context.Context, id uuid.UUID) (*WorkflowTemplate, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("template:%s", id)
    if cached, found := e.templateCache.Get(cacheKey); found {
        return cached.(*WorkflowTemplate), nil
    }

    // Cache miss - fetch from database
    template, err := e.Engine.GetTemplate(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache for 5 minutes
    e.templateCache.Set(cacheKey, template, 5*time.Minute)
    return template, nil
}
```

**What to cache:**
- Workflow templates (TTL: 5 minutes)
- Action definitions (TTL: 10 minutes)
- User role mappings (TTL: 1 minute)

#### 5. Batch Operations

**Process multiple items efficiently:**

```go
// Instead of individual queries
for _, item := range items {
    store.CreateStepExecution(ctx, params)  // ❌ N queries
}

// Use batch insert
err := store.BatchCreateStepExecutions(ctx, paramsSlice)  // ✅ 1 query
```

#### 6. Async Processing

**Don't block API responses:**

```go
func (h *InstanceHandler) StartWorkflow(c *fiber.Ctx) error {
    // Validate request
    // ...

    // Start workflow asynchronously
    go func() {
        // Use background context with timeout
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        _, err := h.engine.StartWorkflow(ctx, req.TemplateID, req.Input)
        if err != nil {
            h.logger.ErrorContext(ctx, "Failed to start workflow", 
                logger.Fields{"error": err.Error()})
        }
    }()

    // Return immediately
    return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
        "message": "Workflow starting...",
    })
}
```

### Monitoring & Observability

#### 1. OpenTelemetry Tracing

**Already configured via otelpgx:**

```go
// Database tracing is automatic with otelpgx
config.ConnConfig.Tracer = otelpgx.NewTracer()

// Add custom spans in workflow logic
ctx, span := tracer.Start(ctx, "workflow.execute_step")
defer span.End()

span.SetAttributes(
    attribute.String("step.id", step.ID),
    attribute.String("step.type", step.Type),
)
```

#### 2. Metrics Collection

```go
// Define metrics
var (
    workflowStarted = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "workflows_started_total",
            Help: "Total workflows started",
        },
        []string{"template_id", "trigger_type"},
    )

    workflowDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "workflow_duration_seconds",
            Help:    "Workflow execution duration",
            Buckets: prometheus.ExponentialBuckets(1, 2, 10),
        },
        []string{"template_id", "status"},
    )

    stepExecutionDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "workflow_step_duration_seconds",
            Help:    "Step execution duration",
            Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
        },
        []string{"step_type"},
    )
)

func init() {
    prometheus.MustRegister(workflowStarted)
    prometheus.MustRegister(workflowDuration)
    prometheus.MustRegister(stepExecutionDuration)
}

// Record metrics
func (e *Engine) StartWorkflow(ctx context.Context, templateID uuid.UUID, input map[string]interface{}) (*WorkflowInstance, error) {
    start := time.Now()

    instance, err := e.startWorkflowInternal(ctx, templateID, input)

    // Record metric
    workflowStarted.WithLabelValues(
        templateID.String(),
        instance.TriggerType,
    ).Inc()

    if err != nil {
        return nil, err
    }

    return instance, nil
}
```

#### 3. Health Checks

```go
// Add health check endpoint
func (s *Service) HealthCheck(c *fiber.Ctx) error {
    ctx := c.Context()

    checks := map[string]string{
        "database": "ok",
        "temporal": "ok",
        "worker":   "ok",
    }

    // Check database
    if err := s.store.HealthCheck(ctx); err != nil {
        checks["database"] = err.Error()
    }

    // Check Temporal
    if _, err := s.temporalClient.CheckHealth(ctx, &healthpb.HealthCheckRequest{}); err != nil {
        checks["temporal"] = err.Error()
    }

    // Check worker
    // (Worker health check implementation)

    status := fiber.StatusOK
    for _, check := range checks {
        if check != "ok" {
            status = fiber.StatusServiceUnavailable
            break
        }
    }

    return c.Status(status).JSON(fiber.Map{
        "status": checks,
    })
}
```

---

## Testing Strategy

### Unit Tests

```go
// internal/workflow/engine/engine_test.go
package engine_test

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    db "awo.so/db/sqlc"
    "awo.so/internal/workflow/engine"
)

func TestCreateTemplate(t *testing.T) {
    // Setup
    mockStore := new(db.MockStore)
    mockTemporal := new(MockTemporalClient)
    
    eng, err := engine.NewEngine(engine.Config{
        Store:          mockStore,
        TemporalClient: mockTemporal,
    })
    assert.NoError(t, err)

    // Test data
    req := engine.CreateTemplateRequest{
        Name:           "Test Workflow",
        TriggerType:    "manual",
        DefinitionJSON: []byte(`{"name":"Test","version":1,"steps":[]}`),
    }

    // Mock expectations
    mockStore.On("CreateWorkflowTemplate", mock.Anything, mock.Anything).
        Return(db.WorkflowTemplate{ID: uuid.New()}, nil)

    // Execute
    template, err := eng.CreateTemplate(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, template)
    mockStore.AssertExpectations(t)
}

func TestValidator_CircularDependency(t *testing.T) {
    validator := engine.NewValidator()

    definition := &engine.WorkflowDefinition{
        Name: "Circular Test",
        Steps: []engine.StepDefinition{
            {ID: "step1", Type: "condition", OnTrue: "step2"},
            {ID: "step2", Type: "action", OnSuccess: "step1"}, // Circular!
        },
    }

    err := validator.Validate(definition)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circular dependency")
}
```

### Integration Tests

```go
// internal/workflow/integration_test.go
package workflow_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/suite"
    db "awo.so/db/sqlc"
    "awo.so/internal/workflow"
)

type WorkflowIntegrationTestSuite struct {
    suite.Suite
    store   db.Store
    service *workflow.Service
}

func (s *WorkflowIntegrationTestSuite) SetupSuite() {
    // Setup test database
    s.store = setupTestDB()
    
    // Setup test Temporal
    temporalClient := setupTestTemporal()
    
    // Create service
    service, err := workflow.NewService(context.Background(), workflow.Config{
        Store:          s.store,
        TemporalClient: temporalClient,
    })
    s.Require().NoError(err)
    s.service = service
}

func (s *WorkflowIntegrationTestSuite) TestCompleteWorkflowExecution() {
    ctx := context.Background()

    // Create template
    template := s.createTestTemplate()

    // Start workflow
    instance, err := s.service.Engine().StartWorkflow(ctx, template.ID, map[string]interface{}{
        "test_value": 100,
    })
    s.Require().NoError(err)

    // Wait for completion (with timeout)
    timeout := time.After(30 * time.Second)
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-timeout:
            s.Fail("Workflow did not complete in time")
            return
        case <-ticker.C:
            updated, err := s.store.GetWorkflowInstance(ctx, instance.ID)
            s.Require().NoError(err)

            if updated.Status == "completed" {
                s.Equal(100, updated.CompletionPercentage)
                return
            }
        }
    }
}

func TestWorkflowIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(WorkflowIntegrationTestSuite))
}
```

### E2E Tests

```go
// tests/e2e/workflow_test.go
package e2e_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestInvoiceApprovalWorkflow(t *testing.T) {
    // Setup test server
    server := setupTestServer()
    defer server.Close()

    // 1. Create workflow template
    template := createTemplate(t, server, map[string]interface{}{
        "name":         "Invoice Approval",
        "trigger_type": "manual",
        "definition": map[string]interface{}{
            "name":    "Invoice Approval",
            "version": 1,
            "steps": []map[string]interface{}{
                {
                    "id":   "validate",
                    "type": "validation",
                    "config": map[string]interface{}{
                        "rules": []map[string]interface{}{
                            {
                                "field":    "$.amount",
                                "operator": ">",
                                "value":    0,
                            },
                        },
                    },
                },
            },
        },
    })

    // 2. Start workflow
    instance := startWorkflow(t, server, template.ID, map[string]interface{}{
        "amount": 1000,
    })

    // 3. Verify workflow started
    assert.Equal(t, "running", instance.Status)

    // 4. Get pending tasks
    tasks := getMyTasks(t, server)
    assert.NotEmpty(t, tasks)

    // 5. Complete task
    completeTask(t, server, tasks[0].ID, map[string]interface{}{
        "decision": "approved",
        "comment":  "LGTM",
    })

    // 6. Wait for workflow completion
    waitForCompletion(t, server, instance.ID, 30*time.Second)

    // 7. Verify final status
    final := getWorkflowInstance(t, server, instance.ID)
    assert.Equal(t, "completed", final.Status)
}
```

---

## Deployment & Operations

### Docker Compose Setup

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: erp_db
      POSTGRES_USER: erp_user
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d

  temporal:
    image: temporalio/auto-setup:1.22.4
    ports:
      - "7233:7233"
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=temporal
      - POSTGRES_PWD=temporal
      - POSTGRES_SEEDS=temporal-postgres
    depends_on:
      - temporal-postgres

  temporal-postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: temporal
      POSTGRES_USER: temporal
      POSTGRES_PASSWORD: temporal
    volumes:
      - temporal_data:/var/lib/postgresql/data

  temporal-ui:
    image: temporalio/ui:2.21.3
    ports:
      - "8080:8080"
    environment:
      - TEMPORAL_ADDRESS=temporal:7233
      - TEMPORAL_CORS_ORIGINS=http://localhost:3000
    depends_on:
      - temporal

  erp-api:
    build: .
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgres://erp_user:secret@postgres:5432/erp_db?sslmode=disable
      - TEMPORAL_HOST=temporal:7233
      - LOG_LEVEL=info
    depends_on:
      - postgres
      - temporal
    volumes:
      - ./migrations:/app/migrations

volumes:
  postgres_data:
  temporal_data:
```

### Kubernetes Deployment

```yaml
# k8s/workflow-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workflow-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: workflow-service
  template:
    metadata:
      labels:
        app: workflow-service
    spec:
      containers:
      - name: workflow-service
        image: your-registry/erp-api:latest
        ports:
        - containerPort: 8000
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: connection-string
        - name: TEMPORAL_HOST
          value: temporal-frontend:7233
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Monitoring Setup

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'workflow-service'
    static_configs:
      - targets: ['workflow-service:8000']
    metrics_path: /metrics

  - job_name: 'temporal'
    static_configs:
      - targets: ['temporal:7233']
```

### Production Checklist

- [ ] Database migrations run successfully
- [ ] RLS policies tested with multiple tenants
- [ ] Temporal namespace per tenant configured
- [ ] Connection pooling configured (25 max, 5 min)
- [ ] Worker auto-scaling configured
- [ ] Metrics collection enabled (Prometheus)
- [ ] Distributed tracing enabled (OpenTelemetry)
- [ ] Log aggregation configured (ELK/Datadog)
- [ ] Backup strategy for workflow data
- [ ] Disaster recovery plan tested
- [ ] Security audit completed
- [ ] Load testing completed (1000+ concurrent workflows)
- [ ] Documentation updated
- [ ] Runbook created for on-call

---

## Summary

This workflow engine provides:

✅ **Complete PostgreSQL integration** using pgx  
✅ **Row-level security** for multi-tenancy  
✅ **Durable execution** via Temporal  
✅ **Visual workflow designer** with AMIS  
✅ **8+ step types** for complex workflows  
✅ **Clean API** for external service integration  
✅ **Production-ready** monitoring and deployment

### Next Steps

1. **Run migrations** to create workflow tables
2. **Start Temporal** cluster (Docker Compose provided)
3. **Initialize workflow service** in your main.go
4. **Register custom actions** for your domain
5. **Create pre-built templates** for common workflows
6. **Integrate with existing services** (invoices, payments, etc.)
7. **Deploy** using provided Docker/K8s configs

The engine is designed to be **plug-and-play** with your existing pgx-based infrastructure while providing enterprise-grade workflow orchestration capabilities.
