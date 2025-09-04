# Workflow Engine

## ⚙️ Overview

The Workflow Engine is a powerful business process automation system that enables organizations to define, execute, and monitor complex business workflows. It supports various workflow patterns, approval hierarchies, escalation procedures, and integration with external systems while providing real-time monitoring and analytics.

## 🏗️ Workflow Architecture

### Core Components

```typescript
interface WorkflowEngine {
  // Core workflow execution
  workflow_executor: WorkflowExecutor;
  
  // Workflow definition and management
  workflow_designer: WorkflowDesigner;
  workflow_repository: WorkflowRepository;
  
  // Task management
  task_manager: TaskManager;
  task_scheduler: TaskScheduler;
  
  // Integration and communication
  notification_service: NotificationService;
  external_integrations: ExternalIntegrationManager;
  
  // Monitoring and analytics
  workflow_monitor: WorkflowMonitor;
  analytics_engine: WorkflowAnalytics;
  
  // State management
  state_manager: WorkflowStateManager;
  persistence_layer: WorkflowPersistence;
}

interface WorkflowDefinition {
  id: string;
  name: string;
  version: string;
  description: string;
  
  // Workflow metadata
  category: string;
  tags: string[];
  created_by: string;
  created_at: Date;
  
  // Workflow structure
  start_event: StartEvent;
  end_events: EndEvent[];
  tasks: Task[];
  gateways: Gateway[];
  sequence_flows: SequenceFlow[];
  
  // Configuration
  configuration: WorkflowConfiguration;
  
  // Validation and deployment
  validation_rules: ValidationRule[];
  deployment_status: 'draft' | 'active' | 'deprecated' | 'retired';
}
```

### Database Schema

```sql
-- Workflow definitions
CREATE TABLE workflow_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Workflow identification
    workflow_key VARCHAR(100) NOT NULL, -- Unique key for versions
    name VARCHAR(255) NOT NULL,
    version VARCHAR(20) NOT NULL DEFAULT '1.0',
    description TEXT,
    
    -- Workflow metadata
    category VARCHAR(100),
    tags JSONB DEFAULT '[]',
    created_by UUID NOT NULL REFERENCES users(id),
    
    -- Workflow definition (BPMN-like structure)
    workflow_definition JSONB NOT NULL,
    
    -- Configuration
    configuration JSONB DEFAULT '{}',
    
    -- Validation and deployment
    validation_errors JSONB DEFAULT '[]',
    deployment_status VARCHAR(20) DEFAULT 'draft',
    deployed_at TIMESTAMPTZ,
    deployed_by UUID REFERENCES users(id),
    
    -- Versioning
    parent_version_id UUID REFERENCES workflow_definitions(id),
    is_latest_version BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_deployment_status CHECK (deployment_status IN ('draft', 'active', 'deprecated', 'retired')),
    UNIQUE(tenant_id, workflow_key, version)
);

-- Workflow instances (executions)
CREATE TABLE workflow_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Workflow reference
    workflow_definition_id UUID NOT NULL REFERENCES workflow_definitions(id),
    workflow_key VARCHAR(100) NOT NULL,
    workflow_version VARCHAR(20) NOT NULL,
    
    -- Instance identification
    business_key VARCHAR(255), -- Optional business identifier
    
    -- Execution context
    started_by UUID REFERENCES users(id),
    initiating_event VARCHAR(100),
    
    -- Instance state
    status VARCHAR(20) DEFAULT 'running',
    current_tasks JSONB DEFAULT '[]', -- Array of active task IDs
    
    -- Data and variables
    process_variables JSONB DEFAULT '{}',
    business_data JSONB DEFAULT '{}',
    
    -- Timing
    started_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    
    -- Error handling
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    
    -- Parent-child relationships (for sub-processes)
    parent_instance_id UUID REFERENCES workflow_instances(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_status CHECK (status IN ('running', 'completed', 'suspended', 'terminated', 'failed'))
);

-- Tasks within workflow instances
CREATE TABLE workflow_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    
    -- Task identification
    task_key VARCHAR(100) NOT NULL, -- Reference to task definition
    task_name VARCHAR(255) NOT NULL,
    task_type VARCHAR(50) NOT NULL, -- user_task, service_task, script_task, etc.
    
    -- Task assignment
    assignee_type VARCHAR(20) DEFAULT 'user', -- user, role, group, system
    assignee_id VARCHAR(255), -- User ID, role name, or group ID
    assigned_at TIMESTAMPTZ,
    
    -- Task state
    status VARCHAR(20) DEFAULT 'created',
    priority INTEGER DEFAULT 50, -- 1-100 scale
    
    -- Timing and SLA
    due_date TIMESTAMPTZ,
    sla_duration_minutes INTEGER,
    sla_violated BOOLEAN DEFAULT false,
    
    -- Task data
    task_data JSONB DEFAULT '{}',
    form_data JSONB DEFAULT '{}',
    
    -- Execution tracking
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    completed_by UUID REFERENCES users(id),
    
    -- Delegation and forwarding
    delegated_to UUID REFERENCES users(id),
    delegated_at TIMESTAMPTZ,
    delegation_reason TEXT,
    
    -- Comments and attachments
    comments JSONB DEFAULT '[]',
    attachments JSONB DEFAULT '[]',
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_task_type CHECK (task_type IN ('user_task', 'service_task', 'script_task', 'send_task', 'receive_task', 'manual_task')),
    CONSTRAINT valid_status CHECK (status IN ('created', 'assigned', 'started', 'completed', 'cancelled', 'failed')),
    CONSTRAINT valid_assignee_type CHECK (assignee_type IN ('user', 'role', 'group', 'system'))
);

-- Workflow execution history
CREATE TABLE workflow_execution_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    
    -- Event details
    event_type VARCHAR(50) NOT NULL, -- started, task_completed, gateway_evaluated, ended, error
    event_timestamp TIMESTAMPTZ DEFAULT NOW(),
    
    -- Context
    task_id UUID REFERENCES workflow_tasks(id),
    user_id UUID REFERENCES users(id),
    
    -- Event data
    event_data JSONB DEFAULT '{}',
    
    -- Performance metrics
    duration_ms INTEGER,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_event_type CHECK (event_type IN ('started', 'task_created', 'task_completed', 'gateway_evaluated', 'ended', 'error', 'suspended', 'resumed'))
);

-- Workflow escalations
CREATE TABLE workflow_escalations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Escalation trigger
    workflow_task_id UUID NOT NULL REFERENCES workflow_tasks(id) ON DELETE CASCADE,
    escalation_rule_id VARCHAR(100), -- Reference to escalation rule definition
    
    -- Escalation details
    escalation_type VARCHAR(30) NOT NULL, -- sla_violation, manual, automatic
    escalation_level INTEGER DEFAULT 1,
    
    -- Escalation actions
    escalated_to UUID REFERENCES users(id),
    escalation_action VARCHAR(50), -- reassign, notify, delegate, skip
    
    -- Timing
    triggered_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    
    -- Escalation data
    escalation_reason TEXT,
    escalation_data JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_escalation_type CHECK (escalation_type IN ('sla_violation', 'manual', 'automatic')),
    CONSTRAINT valid_escalation_action CHECK (escalation_action IN ('reassign', 'notify', 'delegate', 'skip', 'terminate'))
);
```

## 🎯 Workflow Types & Patterns

### Standard Workflow Patterns

#### 1. Sequential Approval Workflow

```yaml
sequential_approval_workflow:
  name: "Purchase Order Approval"
  type: "sequential_approval"
  
  start_event:
    type: "message_start"
    trigger: "purchase_order_submitted"
    
  tasks:
    - id: "supervisor_review"
      type: "user_task"
      name: "Supervisor Review"
      assignment:
        type: "expression"
        expression: "${initiator.supervisor_id}"
      
      conditions:
        - expression: "${purchase_order.amount <= 5000}"
      
      form:
        fields:
          - name: "approval_decision"
            type: "radio"
            options: ["approve", "reject", "request_changes"]
            required: true
          - name: "comments"
            type: "textarea"
            required: false
      
      sla:
        duration: "P1D" # 1 day
        escalation:
          - level: 1
            after: "PT4H" # 4 hours
            action: "notify_manager"
          - level: 2
            after: "P1D" # 1 day
            action: "reassign_to_manager"
    
    - id: "manager_review"
      type: "user_task"
      name: "Manager Review"
      assignment:
        type: "role"
        role: "department_manager"
      
      conditions:
        - expression: "${purchase_order.amount > 5000 && purchase_order.amount <= 25000}"
      
      sla:
        duration: "P2D"
        escalation:
          - level: 1
            after: "P1D"
            action: "notify_director"
    
    - id: "director_approval"
      type: "user_task"
      name: "Director Approval"
      assignment:
        type: "role"
        role: "director"
      
      conditions:
        - expression: "${purchase_order.amount > 25000}"
      
      sla:
        duration: "P3D"
  
  gateways:
    - id: "amount_gateway"
      type: "exclusive"
      conditions:
        - target: "supervisor_review"
          expression: "${purchase_order.amount <= 5000}"
        - target: "manager_review"
          expression: "${purchase_order.amount > 5000 && purchase_order.amount <= 25000}"
        - target: "director_approval"
          expression: "${purchase_order.amount > 25000}"
  
  end_events:
    - id: "approved"
      type: "message_end"
      message: "purchase_order_approved"
    - id: "rejected"
      type: "message_end"
      message: "purchase_order_rejected"
```

#### 2. Parallel Approval Workflow

```yaml
parallel_approval_workflow:
  name: "Multi-Department Budget Approval"
  type: "parallel_approval"
  
  tasks:
    - id: "finance_review"
      type: "user_task"
      name: "Finance Review"
      assignment:
        type: "role"
        role: "finance_manager"
      
      parallel_group: "department_reviews"
      
    - id: "legal_review"
      type: "user_task"
      name: "Legal Review"
      assignment:
        type: "role"
        role: "legal_counsel"
      
      parallel_group: "department_reviews"
      conditions:
        - expression: "${budget_request.legal_review_required == true}"
    
    - id: "it_review"
      type: "user_task"
      name: "IT Review"
      assignment:
        type: "role"
        role: "it_manager"
      
      parallel_group: "department_reviews"
      conditions:
        - expression: "${budget_request.category == 'technology'}"
  
  gateways:
    - id: "parallel_gateway_start"
      type: "parallel"
      outgoing: ["finance_review", "legal_review", "it_review"]
    
    - id: "parallel_gateway_end"
      type: "parallel"
      incoming: ["finance_review", "legal_review", "it_review"]
      join_condition: "all_completed"
```

#### 3. Event-Driven Workflow

```yaml
event_driven_workflow:
  name: "Customer Onboarding"
  type: "event_driven"
  
  start_event:
    type: "message_start"
    trigger: "customer_registration_completed"
  
  tasks:
    - id: "send_welcome_email"
      type: "service_task"
      service: "email_service"
      configuration:
        template: "welcome_email"
        recipient: "${customer.email}"
      
    - id: "create_customer_account"
      type: "service_task"
      service: "customer_service"
      method: "create_account"
      
    - id: "schedule_onboarding_call"
      type: "user_task"
      assignment:
        type: "role"
        role: "customer_success"
      
      timer:
        type: "duration"
        duration: "P1D" # Schedule 1 day after registration
  
  events:
    - id: "payment_received"
      type: "message_intermediate"
      message: "payment_received"
      
    - id: "document_uploaded"
      type: "message_intermediate"
      message: "document_uploaded"
  
  event_handlers:
    - event: "payment_received"
      action: "activate_premium_features"
    - event: "document_uploaded"
      action: "start_document_review"
```

## 🛠️ Workflow Engine Implementation

### Workflow Executor

```go
type WorkflowExecutor struct {
    repository      WorkflowRepository
    taskManager     TaskManager
    eventBus        EventBus
    stateManager    StateManager
    expressionEngine ExpressionEngine
}

type WorkflowInstance struct {
    ID               string                 `json:"id"`
    WorkflowKey      string                 `json:"workflow_key"`
    BusinessKey      string                 `json:"business_key,omitempty"`
    Status           WorkflowStatus         `json:"status"`
    Variables        map[string]interface{} `json:"variables"`
    CurrentTasks     []string               `json:"current_tasks"`
    StartedAt        time.Time              `json:"started_at"`
    EndedAt          *time.Time             `json:"ended_at,omitempty"`
    StartedBy        string                 `json:"started_by"`
    TenantID         string                 `json:"tenant_id"`
}

type WorkflowStatus string

const (
    StatusRunning    WorkflowStatus = "running"
    StatusCompleted  WorkflowStatus = "completed"
    StatusSuspended  WorkflowStatus = "suspended"
    StatusTerminated WorkflowStatus = "terminated"
    StatusFailed     WorkflowStatus = "failed"
)

func (we *WorkflowExecutor) StartWorkflow(ctx context.Context, request StartWorkflowRequest) (*WorkflowInstance, error) {
    // Get workflow definition
    definition, err := we.repository.GetWorkflowDefinition(request.WorkflowKey, request.Version)
    if err != nil {
        return nil, fmt.Errorf("workflow definition not found: %w", err)
    }
    
    // Create workflow instance
    instance := &WorkflowInstance{
        ID:           uuid.New().String(),
        WorkflowKey:  request.WorkflowKey,
        BusinessKey:  request.BusinessKey,
        Status:       StatusRunning,
        Variables:    request.Variables,
        CurrentTasks: []string{},
        StartedAt:    time.Now(),
        StartedBy:    request.StartedBy,
        TenantID:     request.TenantID,
    }
    
    // Save instance
    if err := we.stateManager.SaveWorkflowInstance(ctx, instance); err != nil {
        return nil, fmt.Errorf("failed to save workflow instance: %w", err)
    }
    
    // Execute start event
    if err := we.executeStartEvent(ctx, instance, definition); err != nil {
        return nil, fmt.Errorf("failed to execute start event: %w", err)
    }
    
    // Emit workflow started event
    we.eventBus.Publish("workflow.started", WorkflowEvent{
        WorkflowInstanceID: instance.ID,
        EventType:         "started",
        Timestamp:         time.Now(),
    })
    
    return instance, nil
}

func (we *WorkflowExecutor) executeStartEvent(ctx context.Context, instance *WorkflowInstance, definition *WorkflowDefinition) error {
    startEvent := definition.StartEvent
    
    // Find next tasks from start event
    nextTasks := we.findNextTasks(definition, startEvent.ID)
    
    for _, taskDef := range nextTasks {
        if err := we.createTask(ctx, instance, taskDef); err != nil {
            return fmt.Errorf("failed to create task %s: %w", taskDef.ID, err)
        }
    }
    
    return nil
}

func (we *WorkflowExecutor) createTask(ctx context.Context, instance *WorkflowInstance, taskDef TaskDefinition) error {
    // Evaluate task conditions
    if !we.evaluateConditions(taskDef.Conditions, instance.Variables) {
        return nil // Skip task if conditions not met
    }
    
    // Resolve assignee
    assignee, err := we.resolveAssignee(ctx, taskDef.Assignment, instance)
    if err != nil {
        return fmt.Errorf("failed to resolve assignee: %w", err)
    }
    
    // Create task
    task := &WorkflowTask{
        ID:                 uuid.New().String(),
        WorkflowInstanceID: instance.ID,
        TaskKey:           taskDef.ID,
        TaskName:          taskDef.Name,
        TaskType:          taskDef.Type,
        AssigneeType:      assignee.Type,
        AssigneeID:        assignee.ID,
        Status:            TaskStatusCreated,
        Priority:          taskDef.Priority,
        TaskData:          taskDef.Data,
        DueDate:           we.calculateDueDate(taskDef.SLA),
        CreatedAt:         time.Now(),
    }
    
    // Save task
    if err := we.taskManager.SaveTask(ctx, task); err != nil {
        return fmt.Errorf("failed to save task: %w", err)
    }
    
    // Update instance current tasks
    instance.CurrentTasks = append(instance.CurrentTasks, task.ID)
    if err := we.stateManager.SaveWorkflowInstance(ctx, instance); err != nil {
        return fmt.Errorf("failed to update workflow instance: %w", err)
    }
    
    // Handle different task types
    switch taskDef.Type {
    case "user_task":
        return we.handleUserTask(ctx, task)
    case "service_task":
        return we.handleServiceTask(ctx, task, taskDef)
    case "script_task":
        return we.handleScriptTask(ctx, task, taskDef)
    default:
        return fmt.Errorf("unsupported task type: %s", taskDef.Type)
    }
}

func (we *WorkflowExecutor) CompleteTask(ctx context.Context, request CompleteTaskRequest) error {
    // Get task
    task, err := we.taskManager.GetTask(ctx, request.TaskID)
    if err != nil {
        return fmt.Errorf("task not found: %w", err)
    }
    
    // Validate completion
    if err := we.validateTaskCompletion(task, request); err != nil {
        return fmt.Errorf("invalid task completion: %w", err)
    }
    
    // Update task status
    task.Status = TaskStatusCompleted
    task.CompletedAt = &time.Time{}
    *task.CompletedAt = time.Now()
    task.CompletedBy = &request.CompletedBy
    task.FormData = request.FormData
    
    if err := we.taskManager.SaveTask(ctx, task); err != nil {
        return fmt.Errorf("failed to save task: %w", err)
    }
    
    // Get workflow instance
    instance, err := we.stateManager.GetWorkflowInstance(ctx, task.WorkflowInstanceID)
    if err != nil {
        return fmt.Errorf("workflow instance not found: %w", err)
    }
    
    // Update instance variables with form data
    for key, value := range request.FormData {
        instance.Variables[key] = value
    }
    
    // Remove task from current tasks
    instance.CurrentTasks = removeFromSlice(instance.CurrentTasks, task.ID)
    
    // Get workflow definition
    definition, err := we.repository.GetWorkflowDefinitionByInstance(instance.ID)
    if err != nil {
        return fmt.Errorf("workflow definition not found: %w", err)
    }
    
    // Continue workflow execution
    return we.continueWorkflow(ctx, instance, definition, task.TaskKey)
}

func (we *WorkflowExecutor) continueWorkflow(ctx context.Context, instance *WorkflowInstance, definition *WorkflowDefinition, completedTaskKey string) error {
    // Find next elements after completed task
    nextElements := we.findNextElements(definition, completedTaskKey)
    
    for _, element := range nextElements {
        switch element.Type {
        case "task":
            taskDef := we.getTaskDefinition(definition, element.ID)
            if err := we.createTask(ctx, instance, taskDef); err != nil {
                return fmt.Errorf("failed to create next task: %w", err)
            }
            
        case "gateway":
            if err := we.evaluateGateway(ctx, instance, definition, element.ID); err != nil {
                return fmt.Errorf("failed to evaluate gateway: %w", err)
            }
            
        case "end_event":
            return we.completeWorkflow(ctx, instance, element.ID)
        }
    }
    
    return nil
}

func (we *WorkflowExecutor) evaluateGateway(ctx context.Context, instance *WorkflowInstance, definition *WorkflowDefinition, gatewayID string) error {
    gateway := we.getGatewayDefinition(definition, gatewayID)
    
    switch gateway.Type {
    case "exclusive":
        return we.evaluateExclusiveGateway(ctx, instance, definition, gateway)
    case "parallel":
        return we.evaluateParallelGateway(ctx, instance, definition, gateway)
    case "inclusive":
        return we.evaluateInclusiveGateway(ctx, instance, definition, gateway)
    default:
        return fmt.Errorf("unsupported gateway type: %s", gateway.Type)
    }
}

func (we *WorkflowExecutor) evaluateExclusiveGateway(ctx context.Context, instance *WorkflowInstance, definition *WorkflowDefinition, gateway GatewayDefinition) error {
    // Evaluate conditions in order
    for _, condition := range gateway.Conditions {
        if we.expressionEngine.Evaluate(condition.Expression, instance.Variables) {
            // Create next task/element
            return we.executeNextElement(ctx, instance, definition, condition.Target)
        }
    }
    
    // No condition matched, check for default flow
    if gateway.DefaultFlow != "" {
        return we.executeNextElement(ctx, instance, definition, gateway.DefaultFlow)
    }
    
    return fmt.Errorf("no matching condition in exclusive gateway %s", gateway.ID)
}
```

### Task Management System

```go
type TaskManager struct {
    repository    TaskRepository
    assignmentEngine AssignmentEngine
    notification  NotificationService
    escalation    EscalationManager
}

type WorkflowTask struct {
    ID                 string                 `json:"id"`
    WorkflowInstanceID string                 `json:"workflow_instance_id"`
    TaskKey           string                 `json:"task_key"`
    TaskName          string                 `json:"task_name"`
    TaskType          string                 `json:"task_type"`
    AssigneeType      string                 `json:"assignee_type"`
    AssigneeID        string                 `json:"assignee_id"`
    Status            TaskStatus             `json:"status"`
    Priority          int                    `json:"priority"`
    DueDate           *time.Time             `json:"due_date,omitempty"`
    TaskData          map[string]interface{} `json:"task_data"`
    FormData          map[string]interface{} `json:"form_data"`
    CreatedAt         time.Time              `json:"created_at"`
    AssignedAt        *time.Time             `json:"assigned_at,omitempty"`
    StartedAt         *time.Time             `json:"started_at,omitempty"`
    CompletedAt       *time.Time             `json:"completed_at,omitempty"`
    CompletedBy       *string                `json:"completed_by,omitempty"`
}

type TaskStatus string

const (
    TaskStatusCreated   TaskStatus = "created"
    TaskStatusAssigned  TaskStatus = "assigned"
    TaskStatusStarted   TaskStatus = "started"
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusCancelled TaskStatus = "cancelled"
    TaskStatusFailed    TaskStatus = "failed"
)

func (tm *TaskManager) AssignTask(ctx context.Context, taskID string, assigneeID string) error {
    task, err := tm.repository.GetTask(ctx, taskID)
    if err != nil {
        return err
    }
    
    if task.Status != TaskStatusCreated {
        return fmt.Errorf("task cannot be assigned in status: %s", task.Status)
    }
    
    // Update assignment
    task.AssigneeID = assigneeID
    task.AssigneeType = "user"
    task.Status = TaskStatusAssigned
    now := time.Now()
    task.AssignedAt = &now
    
    if err := tm.repository.SaveTask(ctx, task); err != nil {
        return err
    }
    
    // Send notification
    return tm.notification.SendTaskAssignmentNotification(ctx, task)
}

func (tm *TaskManager) DelegateTask(ctx context.Context, request DelegateTaskRequest) error {
    task, err := tm.repository.GetTask(ctx, request.TaskID)
    if err != nil {
        return err
    }
    
    // Validate delegation permissions
    if task.AssigneeID != request.DelegatedBy {
        return fmt.Errorf("only assigned user can delegate task")
    }
    
    // Create delegation record
    delegation := &TaskDelegation{
        TaskID:          request.TaskID,
        DelegatedFrom:   request.DelegatedBy,
        DelegatedTo:     request.DelegatedTo,
        DelegationReason: request.Reason,
        DelegatedAt:     time.Now(),
    }
    
    if err := tm.repository.SaveDelegation(ctx, delegation); err != nil {
        return err
    }
    
    // Update task assignee
    task.AssigneeID = request.DelegatedTo
    if err := tm.repository.SaveTask(ctx, task); err != nil {
        return err
    }
    
    // Send notifications
    if err := tm.notification.SendDelegationNotification(ctx, task, delegation); err != nil {
        return err
    }
    
    return nil
}

func (tm *TaskManager) GetUserTasks(ctx context.Context, userID string, filters TaskFilters) ([]*WorkflowTask, error) {
    // Get directly assigned tasks
    directTasks, err := tm.repository.GetTasksByAssignee(ctx, userID, filters)
    if err != nil {
        return nil, err
    }
    
    // Get role-based assigned tasks
    userRoles, err := tm.assignmentEngine.GetUserRoles(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    var roleTasks []*WorkflowTask
    for _, role := range userRoles {
        tasks, err := tm.repository.GetTasksByRole(ctx, role, filters)
        if err != nil {
            return nil, err
        }
        roleTasks = append(roleTasks, tasks...)
    }
    
    // Get group-assigned tasks
    userGroups, err := tm.assignmentEngine.GetUserGroups(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    var groupTasks []*WorkflowTask
    for _, group := range userGroups {
        tasks, err := tm.repository.GetTasksByGroup(ctx, group, filters)
        if err != nil {
            return nil, err
        }
        groupTasks = append(groupTasks, tasks...)
    }
    
    // Combine and deduplicate tasks
    allTasks := append(directTasks, roleTasks...)
    allTasks = append(allTasks, groupTasks...)
    
    return tm.deduplicateTasks(allTasks), nil
}
```

### Escalation Management

```go
type EscalationManager struct {
    repository     TaskRepository
    escalationRepo EscalationRepository
    notification   NotificationService
    scheduler      TaskScheduler
}

type EscalationRule struct {
    ID                  string        `json:"id"`
    TaskType           string        `json:"task_type"`
    EscalationLevels   []EscalationLevel `json:"escalation_levels"`
    Enabled            bool          `json:"enabled"`
}

type EscalationLevel struct {
    Level           int           `json:"level"`
    TriggerAfter    time.Duration `json:"trigger_after"`
    EscalationAction string       `json:"escalation_action"`
    EscalateTo      string        `json:"escalate_to"`
    NotificationTemplate string   `json:"notification_template"`
}

func (em *EscalationManager) ScheduleEscalations(ctx context.Context, task *WorkflowTask) error {
    // Get escalation rules for task type
    rules, err := em.escalationRepo.GetEscalationRules(ctx, task.TaskType)
    if err != nil {
        return err
    }
    
    for _, rule := range rules {
        if !rule.Enabled {
            continue
        }
        
        for _, level := range rule.EscalationLevels {
            escalationTime := task.CreatedAt.Add(level.TriggerAfter)
            
            // Schedule escalation
            escalation := &ScheduledEscalation{
                TaskID:           task.ID,
                EscalationLevel:  level.Level,
                ScheduledTime:    escalationTime,
                EscalationAction: level.EscalationAction,
                EscalateTo:       level.EscalateTo,
                Status:          "scheduled",
            }
            
            if err := em.scheduler.ScheduleEscalation(ctx, escalation); err != nil {
                return fmt.Errorf("failed to schedule escalation: %w", err)
            }
        }
    }
    
    return nil
}

func (em *EscalationManager) ProcessEscalation(ctx context.Context, escalation *ScheduledEscalation) error {
    // Get task
    task, err := em.repository.GetTask(ctx, escalation.TaskID)
    if err != nil {
        return err
    }
    
    // Check if task is still active
    if task.Status == TaskStatusCompleted || task.Status == TaskStatusCancelled {
        return nil // Task already completed, no escalation needed
    }
    
    // Execute escalation action
    switch escalation.EscalationAction {
    case "notify":
        return em.executeNotificationEscalation(ctx, task, escalation)
    case "reassign":
        return em.executeReassignmentEscalation(ctx, task, escalation)
    case "delegate":
        return em.executeDelegationEscalation(ctx, task, escalation)
    case "skip":
        return em.executeSkipEscalation(ctx, task, escalation)
    case "terminate":
        return em.executeTerminationEscalation(ctx, task, escalation)
    default:
        return fmt.Errorf("unknown escalation action: %s", escalation.EscalationAction)
    }
}

func (em *EscalationManager) executeNotificationEscalation(ctx context.Context, task *WorkflowTask, escalation *ScheduledEscalation) error {
    // Create escalation record
    escalationRecord := &WorkflowEscalation{
        TaskID:           task.ID,
        EscalationType:   "sla_violation",
        EscalationLevel:  escalation.EscalationLevel,
        EscalatedTo:      escalation.EscalateTo,
        EscalationAction: "notify",
        TriggeredAt:      time.Now(),
        EscalationReason: fmt.Sprintf("Task overdue by %v", time.Since(*task.DueDate)),
    }
    
    if err := em.escalationRepo.SaveEscalation(ctx, escalationRecord); err != nil {
        return err
    }
    
    // Send notification
    return em.notification.SendEscalationNotification(ctx, task, escalationRecord)
}

func (em *EscalationManager) executeReassignmentEscalation(ctx context.Context, task *WorkflowTask, escalation *ScheduledEscalation) error {
    // Create escalation record
    escalationRecord := &WorkflowEscalation{
        TaskID:           task.ID,
        EscalationType:   "sla_violation",
        EscalationLevel:  escalation.EscalationLevel,
        EscalatedTo:      escalation.EscalateTo,
        EscalationAction: "reassign",
        TriggeredAt:      time.Now(),
        EscalationReason: "Task escalated due to SLA violation",
    }
    
    if err := em.escalationRepo.SaveEscalation(ctx, escalationRecord); err != nil {
        return err
    }
    
    // Reassign task
    originalAssignee := task.AssigneeID
    task.AssigneeID = escalation.EscalateTo
    
    if err := em.repository.SaveTask(ctx, task); err != nil {
        return err
    }
    
    // Send notifications
    if err := em.notification.SendReassignmentNotification(ctx, task, originalAssignee, escalation.EscalateTo); err != nil {
        return err
    }
    
    return nil
}
```

## 📊 Workflow Analytics & Monitoring

### Performance Metrics

```typescript
interface WorkflowAnalytics {
  workflow_performance: WorkflowPerformanceMetrics;
  task_performance: TaskPerformanceMetrics;
  user_productivity: UserProductivityMetrics;
  bottleneck_analysis: BottleneckAnalysis;
  sla_compliance: SLAComplianceMetrics;
}

interface WorkflowPerformanceMetrics {
  workflow_id: string;
  workflow_name: string;
  period: DateRange;
  
  execution_metrics: {
    total_instances: number;
    completed_instances: number;
    failed_instances: number;
    average_duration_minutes: number;
    median_duration_minutes: number;
    completion_rate_percentage: number;
  };
  
  timing_metrics: {
    fastest_completion_minutes: number;
    slowest_completion_minutes: number;
    p95_duration_minutes: number;
    p99_duration_minutes: number;
  };
  
  cost_metrics: {
    average_processing_cost: number;
    total_labor_hours: number;
    automation_savings: number;
  };
  
  quality_metrics: {
    error_rate_percentage: number;
    rework_rate_percentage: number;
    first_pass_yield_percentage: number;
  };
}

interface TaskPerformanceMetrics {
  task_type: string;
  task_name: string;
  
  completion_metrics: {
    total_tasks: number;
    completed_tasks: number;
    average_completion_time_minutes: number;
    overdue_tasks: number;
    overdue_percentage: number;
  };
  
  assignment_metrics: {
    average_assignment_time_minutes: number;
    reassignment_rate_percentage: number;
    delegation_rate_percentage: number;
  };
  
  sla_metrics: {
    sla_met_percentage: number;
    average_sla_violation_hours: number;
    escalation_rate_percentage: number;
  };
}

// Analytics Service Implementation
class WorkflowAnalyticsService {
  async generateWorkflowPerformanceReport(
    workflowId: string, 
    period: DateRange
  ): Promise<WorkflowPerformanceMetrics> {
    
    const instances = await this.getWorkflowInstances(workflowId, period);
    const completedInstances = instances.filter(i => i.status === 'completed');
    const failedInstances = instances.filter(i => i.status === 'failed');
    
    // Calculate duration metrics
    const durations = completedInstances
      .map(i => (i.ended_at.getTime() - i.started_at.getTime()) / (1000 * 60))
      .sort((a, b) => a - b);
    
    const averageDuration = durations.reduce((sum, d) => sum + d, 0) / durations.length;
    const medianDuration = durations[Math.floor(durations.length / 2)];
    const p95Duration = durations[Math.floor(durations.length * 0.95)];
    const p99Duration = durations[Math.floor(durations.length * 0.99)];
    
    // Calculate cost metrics
    const laborHours = await this.calculateLaborHours(instances);
    const averageCost = await this.calculateAverageProcessingCost(instances);
    const automationSavings = await this.calculateAutomationSavings(workflowId, period);
    
    // Calculate quality metrics
    const errorRate = (failedInstances.length / instances.length) * 100;
    const reworkRate = await this.calculateReworkRate(instances);
    const firstPassYield = await this.calculateFirstPassYield(instances);
    
    return {
      workflow_id: workflowId,
      workflow_name: await this.getWorkflowName(workflowId),
      period,
      execution_metrics: {
        total_instances: instances.length,
        completed_instances: completedInstances.length,
        failed_instances: failedInstances.length,
        average_duration_minutes: averageDuration,
        median_duration_minutes: medianDuration,
        completion_rate_percentage: (completedInstances.length / instances.length) * 100
      },
      timing_metrics: {
        fastest_completion_minutes: Math.min(...durations),
        slowest_completion_minutes: Math.max(...durations),
        p95_duration_minutes: p95Duration,
        p99_duration_minutes: p99Duration
      },
      cost_metrics: {
        average_processing_cost: averageCost,
        total_labor_hours: laborHours,
        automation_savings: automationSavings
      },
      quality_metrics: {
        error_rate_percentage: errorRate,
        rework_rate_percentage: reworkRate,
        first_pass_yield_percentage: firstPassYield
      }
    };
  }
  
  async identifyBottlenecks(workflowId: string, period: DateRange): Promise<BottleneckAnalysis> {
    const tasks = await this.getTaskPerformanceData(workflowId, period);
    
    // Identify tasks with longest average duration
    const tasksByDuration = tasks.sort((a, b) => b.average_duration - a.average_duration);
    
    // Identify tasks with highest failure rate
    const tasksByFailureRate = tasks.sort((a, b) => b.failure_rate - a.failure_rate);
    
    // Identify tasks with most escalations
    const tasksByEscalations = tasks.sort((a, b) => b.escalation_count - a.escalation_count);
    
    return {
      workflow_id: workflowId,
      analysis_period: period,
      duration_bottlenecks: tasksByDuration.slice(0, 5),
      quality_bottlenecks: tasksByFailureRate.slice(0, 5),
      sla_bottlenecks: tasksByEscalations.slice(0, 5),
      recommendations: this.generateBottleneckRecommendations(tasks)
    };
  }
}
```

This  workflow engine provides powerful business process automation capabilities with sophisticated task management, escalation handling, and analytics to optimize organizational efficiency and ensure consistent process execution across the ERP system.
