## Section 13: Workflow System (Representation Only)

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Workflow struct definition
- Workflow types (linear, branching)
- Step and transition concepts
- **Critical**: Schema represents workflows, doesn't execute them
- Integration with workflow engines (Temporal, Cadence, etc.)

### 13.1 Important Disclaimer

**This schema system REPRESENTS workflows, it does NOT execute them.**

The Workflow struct in the schema is for:
- ✅ **Documenting** workflow structure
- ✅ **Visualizing** workflow steps
- ✅ **UI guidance** for multi-step forms
- ✅ **Integration metadata** for external workflow engines

The schema does NOT:
- ❌ Execute workflow logic
- ❌ Handle retries, compensation, sagas
- ❌ Manage long-running processes
- ❌ Provide durable execution

**For actual workflow execution, integrate with:**
- Temporal
- Cadence
- Conductor
- Camunda
- Custom workflow engine

### 13.2 Workflow Struct

```go
type Workflow struct {
    Enabled     bool            `json:"enabled"`
    Type        string          `json:"type"` // "linear", "branching"
    Steps       []WorkflowStep  `json:"steps"`
    Transitions []Transition    `json:"transitions"`
}

type WorkflowStep struct {
    ID          string                   `json:"id"`
    Name        string                   `json:"name"`
    Description string                   `json:"description,omitempty"`
    Type        string                   `json:"type"` // "human", "system"
    Assignee    string                   `json:"assignee,omitempty"`
    Fields      []string                 `json:"fields,omitempty"`
    Actions     []string                 `json:"actions,omitempty"`
    Timeout     int                      `json:"timeout,omitempty"` // seconds
    Condition   *condition.ConditionGroup `json:"condition,omitempty"`
}

type Transition struct {
    From      string                   `json:"from"`
    To        string                   `json:"to"`
    Action    string                   `json:"action"`
    Condition *condition.ConditionGroup `json:"condition,omitempty"`
}
```

### 13.3 Linear Workflow Example

**Use case:** Simple approval process

```json
{
  "id": "expense-approval",
  "type": "workflow",
  "title": "Expense Approval",
  
  "workflow": {
    "enabled": true,
    "type": "linear",
    "steps": [
      {
        "id": "submit",
        "name": "Submit Request",
        "type": "human",
        "assignee": "employee",
        "fields": ["amount", "category", "description", "receipt"]
      },
      {
        "id": "manager_review",
        "name": "Manager Review",
        "type": "human",
        "assignee": "manager",
        "fields": ["amount", "category", "description"],
        "actions": ["approve", "reject"],
        "timeout": 86400
      },
      {
        "id": "finance_review",
        "name": "Finance Review",
        "type": "human",
        "assignee": "finance",
        "fields": ["amount", "category", "description", "receipt"],
        "actions": ["approve", "reject", "request_info"],
        "timeout": 86400
      },
      {
        "id": "payment",
        "name": "Process Payment",
        "type": "system",
        "description": "Automatic payment processing"
      },
      {
        "id": "complete",
        "name": "Complete",
        "type": "system"
      }
    ],
    "transitions": [
      {"from": "submit", "to": "manager_review", "action": "submit"},
      {"from": "manager_review", "to": "finance_review", "action": "approve"},
      {"from": "manager_review", "to": "complete", "action": "reject"},
      {"from": "finance_review", "to": "payment", "action": "approve"},
      {"from": "finance_review", "to": "complete", "action": "reject"},
      {"from": "payment", "to": "complete", "action": "success"}
    ]
  }
}
```

**Visual:**
```
Submit → Manager Review → Finance Review → Payment → Complete
           │                 │
           └─ Reject ────────┴─ Reject
```

### 13.4 Branching Workflow Example

**Use case:** Conditional approval based on amount

```json
{
  "workflow": {
    "enabled": true,
    "type": "branching",
    "steps": [
      {
        "id": "submit",
        "name": "Submit Request",
        "type": "human"
      },
      {
        "id": "auto_approve",
        "name": "Auto Approve",
        "type": "system",
        "condition": {
          "conjunction": "and",
          "rules": [
            {"left": {"type": "field", "field": "amount"}, "op": "less_or_equal", "right": 100}
          ]
        }
      },
      {
        "id": "manager_approve",
        "name": "Manager Approval",
        "type": "human",
        "assignee": "manager",
        "condition": {
          "conjunction": "and",
          "rules": [
            {"left": {"type": "field", "field": "amount"}, "op": "greater", "right": 100},
            {"left": {"type": "field", "field": "amount"}, "op": "less_or_equal", "right": 1000}
          ]
        }
      },
      {
        "id": "director_approve",
        "name": "Director Approval",
        "type": "human",
        "assignee": "director",
        "condition": {
          "conjunction": "and",
          "rules": [
            {"left": {"type": "field", "field": "amount"}, "op": "greater", "right": 1000}
          ]
        }
      },
      {
        "id": "complete",
        "name": "Complete",
        "type": "system"
      }
    ],
    "transitions": [
      {"from": "submit", "to": "auto_approve", "action": "submit", "condition": {"rules": [{"left": {"type": "field", "field": "amount"}, "op": "less_or_equal", "right": 100}]}},
      {"from": "submit", "to": "manager_approve", "action": "submit", "condition": {"rules": [{"left": {"type": "field", "field": "amount"}, "op": "greater", "right": 100}, {"left": {"type": "field", "field": "amount"}, "op": "less_or_equal", "right": 1000}]}},
      {"from": "submit", "to": "director_approve", "action": "submit", "condition": {"rules": [{"left": {"type": "field", "field": "amount"}, "op": "greater", "right": 1000}]}},
      {"from": "auto_approve", "to": "complete", "action": "approve"},
      {"from": "manager_approve", "to": "complete", "action": "approve"},
      {"from": "manager_approve", "to": "complete", "action": "reject"},
      {"from": "director_approve", "to": "complete", "action": "approve"},
      {"from": "director_approve", "to": "complete", "action": "reject"}
    ]
  }
}
```

**Visual:**
```
                    ┌─ Auto Approve ──┐
                    │                 │
Submit ─────────────┼─ Manager ───────┼─→ Complete
                    │                 │
                    └─ Director ──────┘

(Route depends on amount)
```

### 13.5 Integration with Temporal

**Schema defines structure:**
```json
{
  "workflow": {
    "enabled": true,
    "type": "linear",
    "steps": [...]
  }
}
```

**Temporal executes logic:**
```go
// Temporal workflow definition
func ExpenseApprovalWorkflow(ctx workflow.Context, request ExpenseRequest) error {
    // Load schema to get workflow structure
    schema := loadSchema("expense-approval")
    
    // Execute each step
    for _, step := range schema.Workflow.Steps {
        switch step.Type {
        case "human":
            // Wait for human action
            var action string
            err := workflow.ExecuteActivity(ctx, WaitForHumanAction, step.ID).Get(ctx, &action)
            if err != nil {
                return err
            }
            
        case "system":
            // Execute system action
            err := workflow.ExecuteActivity(ctx, ExecuteSystemAction, step.ID).Get(ctx, nil)
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}

// Activities
func WaitForHumanAction(ctx context.Context, stepID string) (string, error) {
    // Wait for user to take action via UI
    // This is a long-running activity
    return action, nil
}

func ExecuteSystemAction(ctx context.Context, stepID string) error {
    // Execute automated action
    return nil
}
```

### 13.6 UI Visualization

**The schema can be used to generate workflow diagrams:**

```go
func RenderWorkflowDiagram(schema *Schema) {
    // Generate SVG diagram from workflow structure
    svg := `<svg>...`
    
    for _, step := range schema.Workflow.Steps {
        // Draw step box
        addBox(svg, step.ID, step.Name)
    }
    
    for _, transition := range schema.Workflow.Transitions {
        // Draw arrow
        addArrow(svg, transition.From, transition.To)
    }
    
    return svg
}
```

### 13.7 Best Practices

#### Keep Workflow Simple in Schema

```json
// ✅ GOOD: Simple structure
{
  "workflow": {
    "steps": [
      {"id": "submit", "name": "Submit"},
      {"id": "review", "name": "Review"},
      {"id": "approve", "name": "Approve"}
    ]
  }
}

// ❌ BAD: Complex execution logic
{
  "workflow": {
    "steps": [
      {
        "id": "submit",
        "retry": {"max": 3, "backoff": "exponential"},
        "saga": {"compensation": "revert_submission"},
        "sideEffects": ["send_email", "update_cache"]
      }
    ]
  }
}
```

**Put complex logic in Temporal/Cadence, not schema.**

#### Use Schema for UI, Engine for Execution

**Schema:** What steps exist, who can see them
**Engine:** How steps execute, retry logic, compensation

#### Document Integration Points

```json
{
  "workflow": {
    "enabled": true,
    "engine": "temporal",
    "engineWorkflowID": "expense-approval-v1",
    "steps": [...]
  }
}
```

---

This completes **Part 3 (Sections 9-13)**.

**File 3 is complete!** We now have covered:
- Layout System
- Conditional Visibility
- Security & Permissions  
- Events & Actions (HTMX)
- Workflow System

# 📘 Schema Engine Specification v1.0 - Part 4

**Go Implementation**

---

**Table of Contents - Part 4**
- Part III: Go Implementation (Sections 14-20)

---

# PART III: GO IMPLEMENTATION

---

