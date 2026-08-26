# Policy Evaluation API

## Overview

The Policy Evaluation API is the core of the ABAC system, providing real-time authorization decisions with sub-millisecond response times. This service evaluates policies against request contexts and returns  authorization decisions with obligations and advice.

**Goa Service**: `policy-evaluation`  
**Base Path**: `/api/v1/evaluate`

## Business Value

- **Real-time Authorization**: Sub-50ms policy evaluation for user experience
- ** Context**: Rich attribute-based decision making
- **Obligation Enforcement**: Automated compliance and security actions
- **Audit Trail**: Complete decision logging for compliance and security analysis

## Performance Targets

- **Response Time**: <50ms for 99% of evaluations
- **Throughput**: 10,000+ concurrent evaluations
- **Cache Hit Rate**: >90% for repeated evaluations
- **Availability**: 99.99% uptime with automatic failover

## API Endpoints

### Single Policy Evaluation

**Action**: `EvaluatePolicy`

```yaml
POST /
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>
X-Request-ID: <correlation_id>

Request Body:
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "subject": {
    "id": "user_123",
    "type": "user",
    "attributes": {
      "user.id": "user_123",
      "user.department": "finance",
      "user.security_level": 7,
      "user.employment_status": "active",
      "user.roles": ["financial_analyst", "report_viewer"],
      "user.manager_id": "mgr_456",
      "user.clearance": "SECRET",
      "user.last_training": "2024-12-01T00:00:00Z"
    }
  },
  "resource": {
    "id": "resource_789",
    "type": "financial_report",
    "attributes": {
      "resource.id": "resource_789",
      "resource.classification": "financial",
      "resource.type": "report",
      "resource.sensitivity": "high",
      "resource.owner": "finance_team",
      "resource.created_date": "2025-01-10T00:00:00Z",
      "resource.data_classification": "confidential",
      "resource.geographic_restriction": ["US", "CA"]
    }
  },
  "action": {
    "name": "read",
    "attributes": {
      "action.risk_level": "medium",
      "action.compliance_required": true
    }
  },
  "environment": {
    "time": {
      "current_time": "2025-01-15T14:30:00Z",
      "timezone": "UTC",
      "business_hours": true,
      "current_hour": 14,
      "day_of_week": 3,
      "is_weekend": false
    },
    "session": {
      "id": "session_abc123",
      "mfa_verified": true,
      "mfa_verified_at": "2025-01-15T09:00:00Z",
      "login_time": "2025-01-15T09:00:00Z",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "device_type": "desktop",
      "security_score": 0.95
    },
    "location": {
      "network": "corporate",
      "country": "US",
      "region": "California",
      "city": "San Francisco",
      "building": "headquarters",
      "floor": "5"
    },
    "system": {
      "threat_level": "low",
      "maintenance_mode": false,
      "peak_usage": false
    }
  },
  "context": {
    "request_source": "web_app",
    "correlation_id": "corr_xyz789",
    "trace_enabled": true,
    "cache_preference": "prefer_cache"
  }
}

Response: 200 OK
{
  "decision": "ALLOW",
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "evaluation_id": "eval_550e8400-e29b-41d4-a716-446655440000",
  "evaluated_at": "2025-01-15T14:30:00.123Z",
  "applicable_policies": [
    {
      "policy_id": "policy_550e8400-e29b-41d4-a716-446655440001",
      "policy_name": "financial_data_access_policy",
      "policy_version": 2,
      "effect": "ALLOW",
      "priority": 100,
      "matched": true,
      "evaluation_details": {
        "target_match": true,
        "rule_evaluation": "SATISFIED",
        "satisfied_conditions": [
          "user.security_level >= 6 (7 >= 6)",
          "time.business_hours == true",
          "session.mfa_verified == true",
          "user.department in ['finance', 'accounting', 'audit']",
          "resource.classification == 'financial'"
        ],
        "failing_conditions": [],
        "condition_trace": [
          {
            "condition": "user.security_level >= 6",
            "result": true,
            "actual_value": 7,
            "expected_value": ">=6"
          },
          {
            "condition": "time.business_hours == true",
            "result": true,
            "actual_value": true,
            "expected_value": true
          }
        ]
      }
    }
  ],
  "obligations": [
    {
      "type": "LOG_ACCESS",
      "id": "obligation_001",
      "parameters": {
        "level": "HIGH",
        "include_context": true,
        "retention_days": 2555,
        "compliance_flags": ["SOX", "PCI"]
      },
      "execution_required": true,
      "execution_deadline": "2025-01-15T14:30:05.123Z"
    },
    {
      "type": "MASK_SENSITIVE_FIELDS",
      "id": "obligation_002",
      "parameters": {
        "fields": ["ssn", "account_number", "routing_number"],
        "mask_type": "partial",
        "preserve_length": true
      },
      "execution_required": true,
      "execution_context": "response_transformation"
    },
    {
      "type": "RATE_LIMIT_CHECK",
      "id": "obligation_003",
      "parameters": {
        "max_requests_per_hour": 100,
        "current_count": 23,
        "reset_time": "2025-01-15T15:00:00Z"
      },
      "execution_required": true
    }
  ],
  "advice": [
    {
      "type": "NOTIFY_MANAGER",
      "id": "advice_001",
      "parameters": {
        "delay_minutes": 5,
        "include_resource_info": false,
        "notification_template": "sensitive_access"
      },
      "recommended": true,
      "priority": "low"
    },
    {
      "type": "SECURITY_ALERT",
      "id": "advice_002",
      "parameters": {
        "alert_level": "info",
        "reason": "high_value_resource_access",
        "include_user_context": true
      },
      "recommended": true,
      "priority": "medium"
    }
  ],
  "evaluation_metadata": {
    "evaluation_time_ms": 15.2,
    "cache_hit": false,
    "cache_key": "eval_cache_abc123",
    "attribute_retrieval_time_ms": 8.1,
    "policy_evaluation_time_ms": 7.1,
    "pdp_instance": "pdp-cluster-01",
    "policies_considered": 12,
    "policies_matched": 1,
    "attribute_sources": {
      "user_attributes": "ldap_service",
      "resource_attributes": "resource_service",
      "environment_attributes": "context_service"
    }
  },
  "risk_assessment": {
    "overall_risk_score": 25,
    "risk_factors": [
      {
        "factor": "high_value_resource",
        "score": 15,
        "weight": 0.3
      },
      {
        "factor": "sensitive_action",
        "score": 10,
        "weight": 0.2
      }
    ],
    "risk_level": "medium",
    "additional_monitoring": false
  },
  "context": {
    "tenant_id": "tenant_123",
    "correlation_id": "corr_xyz789",
    "evaluation_context": "standard",
    "cache_ttl_seconds": 300
  }
}
```

### Bulk Policy Evaluation

**Action**: `BulkEvaluate`

```yaml
POST /bulk
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "requests": [
    {
      "request_id": "req_001",
      "subject": {
        "id": "user_123",
        "attributes": {
          "user.department": "finance",
          "user.security_level": 7,
          "user.employment_status": "active"
        }
      },
      "resource": {
        "id": "resource_789",
        "attributes": {
          "resource.classification": "financial",
          "resource.type": "report"
        }
      },
      "action": {"name": "read"},
      "environment": {
        "time": {"business_hours": true},
        "session": {"mfa_verified": true}
      }
    },
    {
      "request_id": "req_002", 
      "subject": {
        "id": "user_456",
        "attributes": {
          "user.department": "hr",
          "user.security_level": 3,
          "user.employment_status": "active"
        }
      },
      "resource": {
        "id": "resource_789",
        "attributes": {
          "resource.classification": "financial",
          "resource.type": "report"
        }
      },
      "action": {"name": "read"},
      "environment": {
        "time": {"business_hours": true},
        "session": {"mfa_verified": false}
      }
    },
    {
      "request_id": "req_003",
      "subject": {
        "id": "user_789",
        "attributes": {
          "user.department": "finance",
          "user.security_level": 8,
          "user.employment_status": "active"
        }
      },
      "resource": {
        "id": "resource_456",
        "attributes": {
          "resource.classification": "public",
          "resource.type": "announcement"
        }
      },
      "action": {"name": "read"},
      "environment": {
        "time": {"business_hours": false}
      }
    }
  ],
  "options": {
    "fail_fast": false,
    "include_details": true,
    "max_evaluation_time_ms": 1000,
    "parallel_evaluation": true,
    "cache_results": true
  }
}

Response: 200 OK
{
  "bulk_evaluation_id": "bulk_eval_550e8400-e29b-41d4-a716-446655440000",
  "evaluated_at": "2025-01-15T14:30:00.123Z",
  "results": [
    {
      "request_id": "req_001",
      "decision": "ALLOW",
      "evaluation_time_ms": 12.3,
      "cache_hit": false,
      "policies_matched": 1,
      "obligations_count": 2,
      "risk_score": 25
    },
    {
      "request_id": "req_002",
      "decision": "DENY",
      "evaluation_time_ms": 8.7,
      "cache_hit": false,
      "policies_matched": 0,
      "denial_reason": "Insufficient security clearance and missing MFA",
      "failing_conditions": [
        "user.security_level >= 6 (3 < 6)",
        "session.mfa_verified == true (false != true)"
      ],
      "risk_score": 45
    },
    {
      "request_id": "req_003",
      "decision": "ALLOW",
      "evaluation_time_ms": 5.1,
      "cache_hit": true,
      "policies_matched": 1,
      "obligations_count": 0,
      "risk_score": 5
    }
  ],
  "summary": {
    "total_requests": 3,
    "allow_count": 2,
    "deny_count": 1,
    "error_count": 0,
    "total_evaluation_time_ms": 26.1,
    "avg_evaluation_time_ms": 8.7,
    "cache_hit_rate": 0.33,
    "parallel_efficiency": 0.89
  },
  "performance_metrics": {
    "peak_concurrent_evaluations": 3,
    "attribute_retrieval_time_ms": 15.4,
    "policy_matching_time_ms": 8.2,
    "rule_evaluation_time_ms": 2.5
  }
}
```

### Evaluation with Approval Workflow

**Action**: `EvaluateWithApproval`

```yaml
POST /with-approval
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "subject": {
    "id": "user_123",
    "attributes": {
      "user.department": "finance",
      "user.security_level": 4,
      "user.employment_status": "active",
      "user.manager_id": "mgr_456"
    }
  },
  "resource": {
    "id": "classified_doc_456",
    "attributes": {
      "resource.classification": "top_secret",
      "resource.type": "document",
      "resource.owner": "security_team",
      "resource.sensitivity": "critical"
    }
  },
  "action": {"name": "read"},
  "environment": {
    "time": {"business_hours": true},
    "session": {"mfa_verified": true},
    "location": {"network": "corporate"}
  },
  "approval_context": {
    "justification": "Need access for quarterly audit compliance review",
    "urgency": "medium",
    "business_reason": "SOX compliance requirement",
    "expected_duration_hours": 2,
    "additional_context": {
      "audit_case_id": "AUD-2025-001",
      "compliance_officer": "officer_789"
    }
  }
}

Response: 202 Accepted
{
  "decision": "PENDING_APPROVAL",
  "evaluation_id": "eval_pending_550e8400-e29b-41d4-a716-446655440001",
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "evaluated_at": "2025-01-15T14:30:00.123Z",
  "approval_required_reason": "Resource classification exceeds user clearance level",
  "approval_request": {
    "id": "approval_req_550e8400-e29b-41d4-a716-446655440001",
    "status": "pending",
    "created_at": "2025-01-15T14:30:00.123Z",
    "expires_at": "2025-01-16T14:30:00.123Z",
    "approvers": [
      {
        "id": "mgr_456",
        "name": "Jane Manager",
        "email": "jane.manager@company.com",
        "role": "security_manager",
        "approval_level": "level_2"
      },
      {
        "id": "security_officer_123",
        "name": "Security Officer",
        "email": "security@company.com",
        "role": "security_officer",
        "approval_level": "level_3"
      }
    ],
    "approval_workflow": {
      "type": "sequential",
      "steps": [
        {
          "step": 1,
          "approver_id": "mgr_456",
          "required": true,
          "estimated_time_hours": 2
        },
        {
          "step": 2,
          "approver_id": "security_officer_123",
          "required": true,
          "estimated_time_hours": 4
        }
      ]
    },
    "justification": "Need access for quarterly audit compliance review",
    "business_reason": "SOX compliance requirement",
    "urgency": "medium"
  },
  "estimated_approval_time": "4-6 hours",
  "notification_sent": true,
  "monitoring": {
    "escalation_time": "2025-01-15T18:30:00Z",
    "auto_deny_time": "2025-01-16T14:30:00Z"
  },
  "interim_access": {
    "granted": false,
    "reason": "Classification level too high for interim access"
  }
}
```

### Get Evaluation Status

**Action**: `GetEvaluationStatus`

```yaml
GET /status/{evaluation_id}
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Path Parameters:
- evaluation_id: UUID (required)

Response: 200 OK
{
  "evaluation_id": "eval_pending_550e8400-e29b-41d4-a716-446655440001",
  "status": "approved",
  "final_decision": "ALLOW",
  "original_request": {
    "subject_id": "user_123",
    "resource_id": "classified_doc_456",
    "action": "read",
    "requested_at": "2025-01-15T14:30:00.123Z"
  },
  "approval_details": {
    "approved_by": {
      "id": "mgr_456",
      "name": "Jane Manager",
      "role": "security_manager",
      "approved_at": "2025-01-15T16:45:00Z"
    },
    "secondary_approval": {
      "id": "security_officer_123",
      "name": "Security Officer", 
      "role": "security_officer",
      "approved_at": "2025-01-15T17:15:00Z"
    },
    "approval_comments": "Approved for audit purposes. Access expires in 48 hours per SOX requirements.",
    "approval_duration": "6 hours 45 minutes"
  },
  "conditional_access": {
    "expires_at": "2025-01-17T16:45:00Z",
    "max_duration_hours": 48,
    "additional_obligations": [
      {
        "type": "TIME_LIMITED_ACCESS",
        "parameters": {
          "max_duration_hours": 48,
          "auto_revoke": true
        }
      },
      {
        "type": "ENHANCED_MONITORING",
        "parameters": {
          "monitor_all_actions": true,
          "real_time_alerts": true
        }
      }
    ],
    "additional_restrictions": [
      {
        "type": "NO_EXPORT_ALLOWED",
        "parameters": {
          "blocked_actions": ["export", "download", "print"]
        }
      }
    ]
  },
  "compliance_trail": {
    "audit_case_id": "AUD-2025-001",
    "compliance_officer": "officer_789",
    "legal_basis": "SOX_compliance_audit",
    "retention_period": "7_years"
  }
}
```

### Policy Impact Simulation

**Action**: `SimulateEvaluation`

```yaml
POST /simulate
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "baseline_request": {
    "subject": {
      "id": "user_123",
      "attributes": {
        "user.security_level": 5,
        "user.department": "finance"
      }
    },
    "resource": {
      "id": "resource_789",
      "attributes": {
        "resource.classification": "financial"
      }
    },
    "action": {"name": "read"},
    "environment": {
      "time": {"current_hour": 14}
    }
  },
  "policy_changes": [
    {
      "action": "modify",
      "policy_id": "policy_550e8400-e29b-41d4-a716-446655440001",
      "changes": {
        "rule": {
          "and": [
            {"user.security_level": {"gte": 6}},
            {"time.current_hour": {"between": [9, 17]}}
          ]
        }
      }
    }
  ],
  "variations": [
    {
      "name": "higher_security_level",
      "attribute_changes": {
        "user.security_level": 6
      }
    },
    {
      "name": "after_hours",
      "attribute_changes": {
        "time.current_hour": 22
      }
    }
  ]
}

Response: 200 OK
{
  "simulation_id": "sim_550e8400-e29b-41d4-a716-446655440000",
  "executed_at": "2025-01-15T14:45:00Z",
  "baseline_result": {
    "decision": "ALLOW",
    "policies_matched": 1,
    "evaluation_time_ms": 15.2
  },
  "simulated_results": {
    "baseline_with_changes": {
      "decision": "DENY",
      "reason": "user.security_level 5 < 6 (new requirement)",
      "policies_matched": 0,
      "evaluation_time_ms": 12.8
    },
    "variations": [
      {
        "name": "higher_security_level",
        "baseline_decision": "ALLOW",
        "simulated_decision": "ALLOW",
        "impact": "no_change",
        "evaluation_time_ms": 13.1
      },
      {
        "name": "after_hours",
        "baseline_decision": "ALLOW",
        "simulated_decision": "DENY",
        "impact": "access_denied",
        "reason": "time.current_hour 22 not between [9,17]",
        "evaluation_time_ms": 11.7
      }
    ]
  },
  "impact_summary": {
    "baseline_access_preserved": false,
    "variations_affected": 1,
    "performance_impact": {
      "avg_time_change_ms": -1.4,
      "performance_improvement": true
    }
  },
  "recommendations": [
    "Consider granular access controls for users with security_level 5",
    "Implement emergency access procedures for after-hours scenarios"
  ]
}
```

### Clear Evaluation Cache

**Action**: `ClearCache`

```yaml
DELETE /cache
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "cache_scope": {
    "cache_types": ["evaluation", "policy", "attribute"],
    "selective_clear": {
      "user_ids": ["user_123", "user_456"],
      "resource_ids": ["resource_789"],
      "policy_ids": ["policy_550e8400-e29b-41d4-a716-446655440001"],
      "cache_patterns": ["eval_*_finance_*"]
    }
  },
  "options": {
    "clear_all": false,
    "warm_cache": true,
    "async_clear": true
  },
  "reason": "Policy update deployment"
}

Response: 200 OK
{
  "cache_clear_id": "clear_550e8400-e29b-41d4-a716-446655440000",
  "started_at": "2025-01-15T15:00:00Z",
  "status": "in_progress",
  "cleared_caches": {
    "evaluation": {
      "entries_cleared": 1547,
      "cache_size_before_mb": 45.7,
      "cache_size_after_mb": 32.1
    },
    "policy": {
      "entries_cleared": 89,
      "policies_recompiled": 12
    },
    "attribute": {
      "entries_cleared": 892,
      "attributes_refreshed": 234
    }
  },
  "selective_clear_results": {
    "users_affected": 2,
    "resources_affected": 1,
    "policies_affected": 1,
    "pattern_matches": 234
  },
  "cache_warming": {
    "enabled": true,
    "estimated_completion": "2025-01-15T15:05:00Z",
    "priority_policies": 12,
    "critical_evaluations": 45
  },
  "performance_impact": {
    "expected_cache_miss_rate": 0.15,
    "temporary_latency_increase_ms": 5.2,
    "recovery_time_minutes": 3
  }
}
```

## Goa DSL Implementation Notes

### Service Definition

```go
var _ = Service("policy-evaluation", func() {
    Description("Real-time ABAC Policy Evaluation Service")
    
    HTTP(func() {
        Path("/api/v1/evaluate")
        Header("X-Tenant-ID", String, "Tenant identifier", func() {
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
        Header("X-Request-ID", String, "Request correlation ID", func() {
            Example("req_550e8400-e29b-41d4-a716-446655440000")
        })
    })
    
    JWT(func() {
        Description("JWT authentication")
        Scope("evaluate:read", "Perform policy evaluation")
        Scope("evaluate:bulk", "Perform bulk evaluations")
        Scope("evaluate:admin", "Administrative evaluation functions")
    })
})
```

### Type Definitions

```go
var EvaluationRequest = Type("EvaluationRequest", func() {
    Attribute("request_id", String, "Unique request identifier", func() {
        Format(FormatUUID)
    })
    Attribute("subject", Subject, "Subject information")
    Attribute("resource", Resource, "Resource information")  
    Attribute("action", Action, "Action information")
    Attribute("environment", Environment, "Environment context")
    Attribute("context", EvaluationContext, "Additional context")
    
    Required("subject", "resource", "action")
})

var EvaluationResponse = Type("EvaluationResponse", func() {
    Attribute("decision", String, "Evaluation decision", func() {
        Enum("ALLOW", "DENY", "PENDING_APPROVAL", "ERROR")
    })
    Attribute("request_id", String, "Original request ID")
    Attribute("evaluation_id", String, "Evaluation identifier")
    Attribute("evaluated_at", String, "Evaluation timestamp", func() {
        Format(FormatDateTime)
    })
    Attribute("applicable_policies", ArrayOf(PolicyEvaluation), "Policies evaluated")
    Attribute("obligations", ArrayOf(Obligation), "Required obligations")
    Attribute("advice", ArrayOf(Advice), "Optional advice")
    Attribute("evaluation_metadata", EvaluationMetadata, "Performance and debug info")
    Attribute("risk_assessment", RiskAssessment, "Risk analysis")
    
    Required("decision", "evaluation_id", "evaluated_at")
})

var Subject = Type("Subject", func() {
    Attribute("id", String, "Subject identifier")
    Attribute("type", String, "Subject type", func() {
        Enum("user", "service", "device", "application")
        Default("user")
    })
    Attribute("attributes", MapOf(String, Any), "Subject attributes")
    
    Required("id", "attributes")
})

var Obligation = Type("Obligation", func() {
    Attribute("type", String, "Obligation type")
    Attribute("id", String, "Obligation identifier")
    Attribute("parameters", MapOf(String, Any), "Obligation parameters")
    Attribute("execution_required", Boolean, "Whether execution is required")
    Attribute("execution_deadline", String, "Execution deadline", func() {
        Format(FormatDateTime)
    })
    
    Required("type", "parameters", "execution_required")
})
```

### Performance Middleware

```go
// Performance monitoring middleware
func PerformanceMiddleware() func(endpoint.Endpoint) endpoint.Endpoint {
    return func(next endpoint.Endpoint) endpoint.Endpoint {
        return func(ctx context.Context, request any) (any, error) {
            start := time.Now()
            
            // Add performance context
            ctx = context.WithValue(ctx, "start_time", start)
            ctx = context.WithValue(ctx, "performance_tracking", true)
            
            response, err := next(ctx, request)
            
            duration := time.Since(start)
            
            // Log performance metrics
            if duration > 50*time.Millisecond {
                log.Warn("Slow evaluation detected", 
                    "duration_ms", duration.Milliseconds(),
                    "request", request)
            }
            
            return response, err
        }
    }
}
```

### Caching Strategy

```go
// Cache configuration
type CacheConfig struct {
    EvaluationCache struct {
        TTL     time.Duration `default:"5m"`
        Size    int          `default:"10000"`
        Enabled bool         `default:"true"`
    }
    PolicyCache struct {
        TTL     time.Duration `default:"1h"`
        Size    int          `default:"1000"`
        Enabled bool         `default:"true"`
    }
    AttributeCache struct {
        TTL     time.Duration `default:"30m"`
        Size    int          `default:"50000"`
        Enabled bool         `default:"true"`
    }
}
```

### Error Handling

```go
var EvaluationTimeout = ErrorResult("evaluation_timeout", func() {
    Description("Policy evaluation exceeded maximum time limit")
    HTTP(func() {
        Status(StatusRequestTimeout)
    })
})

var InsufficientAttributes = ErrorResult("insufficient_attributes", func() {
    Description("Required attributes missing for evaluation")
    HTTP(func() {
        Status(StatusBadRequest)
    })
})

var PolicyCompilationError = ErrorResult("policy_compilation_error", func() {
    Description("Policy compilation or execution error")
    HTTP(func() {
        Status(StatusInternalServerError)
    })
})
```

### Monitoring & Observability

- **Real-time Metrics**: Evaluation latency, throughput, error rates
- **Performance Tracking**: P50, P95, P99 response times
- **Cache Monitoring**: Hit rates, eviction rates, memory usage
- **Error Analysis**: Error categorization and trending
- **Distributed Tracing**: Request flow across services
- **Health Checks**: Service health and dependency status


<!-- ## ⚡ 3. Policy Evaluation API (Core ABAC Engine) -->
<!-- *Business Value*: Real-time authorization decisions with  context evaluation and audit trails. -->
<!---->
<!---->
<!-- ```json  -->
<!-- # Policy Evaluation API (Core ABAC Engine) -->
<!---->
<!-- ## Single Policy Evaluation ->
<!-- POST /evaluate -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "request_id": "req_550e8400-e29b-41d4-a716-446655440000", -->
<!--   "subject": { -->
<!--     "id": "user_123", -->
<!--     "type": "user", -->
<!--     "attributes": { -->
<!--       "user.id": "user_123", -->
<!--       "user.department": "finance", -->
<!--       "user.security_level": 7, -->
<!--       "user.employment_status": "active", -->
<!--       "user.roles": ["financial_analyst", "report_viewer"], -->
<!--       "user.manager_id": "mgr_456" -->
<!--     } -->
<!--   }, -->
<!--   "resource": { -->
<!--     "id": "resource_789", -->
<!--     "type": "financial_report", -->
<!--     "attributes": { -->
<!--       "resource.classification": "financial", -->
<!--       "resource.type": "report", -->
<!--       "resource.sensitivity": "high", -->
<!--       "resource.owner": "finance_team", -->
<!--       "resource.created_date": "2025-01-10T00:00:00Z" -->
<!--     } -->
<!--   }, -->
<!--   "action": { -->
<!--     "name": "read", -->
<!--     "attributes": { -->
<!--       "action.risk_level": "medium" -->
<!--     } -->
<!--   }, -->
<!--   "environment": { -->
<!--     "time": { -->
<!--       "current_time": "2025-01-15T14:30:00Z", -->
<!--       "timezone": "UTC", -->
<!--       "business_hours": true -->
<!--     }, -->
<!--     "session": { -->
<!--       "id": "session_abc123", -->
<!--       "mfa_verified": true, -->
<!--       "login_time": "2025-01-15T09:00:00Z", -->
<!--       "ip_address": "192.168.1.100", -->
<!--       "user_agent": "Mozilla/5.0..." -->
<!--     }, -->
<!--     "location": { -->
<!--       "network": "corporate", -->
<!--       "country": "US", -->
<!--       "building": "headquarters" -->
<!--     } -->
<!--   }, -->
<!--   "context": { -->
<!--     "request_source": "web_app", -->
<!--     "correlation_id": "corr_xyz789" -->
<!--   } -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "decision": "ALLOW", -->
<!--   "request_id": "req_550e8400-e29b-41d4-a716-446655440000", -->
<!--   "evaluation_id": "eval_550e8400-e29b-41d4-a716-446655440000", -->
<!--   "applicable_policies": [ -->
<!--     { -->
<!--       "policy_id": "550e8400-e29b-41d4-a716-446655440001", -->
<!--       "policy_name": "financial_data_access_policy", -->
<!--       "effect": "ALLOW", -->
<!--       "priority": 100, -->
<!--       "matched": true, -->
<!--       "evaluation_details": { -->
<!--         "target_match": true, -->
<!--         "rule_evaluation": "SATISFIED", -->
<!--         "failing_conditions": [], -->
<!--         "satisfied_conditions": [ -->
<!--           "user.security_level >= 5", -->
<!--           "time.business_hours == true", -->
<!--           "session.mfa_verified == true" -->
<!--         ] -->
<!--       } -->
<!--     } -->
<!--   ], -->
<!--   "obligations": [ -->
<!--     { -->
<!--       "type": "LOG_ACCESS", -->
<!--       "parameters": { -->
<!--         "level": "HIGH", -->
<!--         "include_context": true, -->
<!--         "retention_days": 2555 -->
<!--       }, -->
<!--       "execution_required": true -->
<!--     }, -->
<!--     { -->
<!--       "type": "MASK_SENSITIVE_FIELDS", -->
<!--       "parameters": { -->
<!--         "fields": ["ssn", "account_number"], -->
<!--         "mask_type": "partial" -->
<!--       }, -->
<!--       "execution_required": true -->
<!--     } -->
<!--   ], -->
<!--   "advice": [ -->
<!--     { -->
<!--       "type": "NOTIFY_MANAGER", -->
<!--       "parameters": { -->
<!--         "delay_minutes": 5, -->
<!--         "include_resource_info": false -->
<!--       }, -->
<!--       "recommended": true -->
<!--     } -->
<!--   ], -->
<!--   "evaluation_metadata": { -->
<!--     "evaluation_time_ms": 15.2, -->
<!--     "cache_hit": false, -->
<!--     "attribute_retrieval_time_ms": 8.1, -->
<!--     "policy_evaluation_time_ms": 7.1, -->
<!--     "evaluated_at": "2025-01-15T14:30:00.123Z", -->
<!--     "pdp_instance": "pdp-01" -->
<!--   }, -->
<!--   "context": { -->
<!--     "tenant_id": "tenant_123", -->
<!--     "correlation_id": "corr_xyz789" -->
<!--   } -->
<!-- } -->
<!---->
<!-- ## Bulk Policy Evaluation ->
<!-- POST /evaluate/bulk -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "requests": [ -->
<!--     { -->
<!--       "request_id": "req_001", -->
<!--       "subject": { -->
<!--         "id": "user_123", -->
<!--         "attributes": { -->
<!--           "user.department": "finance", -->
<!--           "user.security_level": 7 -->
<!--         } -->
<!--       }, -->
<!--       "resource": { -->
<!--         "id": "resource_789", -->
<!--         "attributes": { -->
<!--           "resource.classification": "financial" -->
<!--         } -->
<!--       }, -->
<!--       "action": {"name": "read"}, -->
<!--       "environment": { -->
<!--         "time": {"business_hours": true}, -->
<!--         "session": {"mfa_verified": true} -->
<!--       } -->
<!--     }, -->
<!--     { -->
<!--       "request_id": "req_002",  -->
<!--       "subject": { -->
<!--         "id": "user_456", -->
<!--         "attributes": { -->
<!--           "user.department": "hr", -->
<!--           "user.security_level": 3 -->
<!--         } -->
<!--       }, -->
<!--       "resource": { -->
<!--         "id": "resource_789", -->
<!--         "attributes": { -->
<!--           "resource.classification": "financial" -->
<!--         } -->
<!--       }, -->
<!--       "action": {"name": "read"}, -->
<!--       "environment": { -->
<!--         "time": {"business_hours": true} -->
<!--       } -->
<!--     } -->
<!--   ], -->
<!--   "options": { -->
<!--     "fail_fast": false, -->
<!--     "include_details": true, -->
<!--     "max_evaluation_time_ms": 1000 -->
<!--   } -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "results": [ -->
<!--     { -->
<!--       "request_id": "req_001", -->
<!--       "decision": "ALLOW", -->
<!--       "evaluation_time_ms": 12.3 -->
<!--     }, -->
<!--     { -->
<!--       "request_id": "req_002",  -->
<!--       "decision": "DENY", -->
<!--       "evaluation_time_ms": 8.7, -->
<!--       "denial_reason": "Insufficient security clearance" -->
<!--     } -->
<!--   ], -->
<!--   "summary": { -->
<!--     "total_requests": 2, -->
<!--     "allow_count": 1, -->
<!--     "deny_count": 1, -->
<!--     "error_count": 0, -->
<!--     "total_evaluation_time_ms": 21.0, -->
<!--     "avg_evaluation_time_ms": 10.5 -->
<!--   } -->
<!-- } -->
<!---->
<!-- ## Evaluate with Approval Workflow ->
<!-- POST /evaluate/with-approval -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "subject": { -->
<!--     "id": "user_123", -->
<!--     "attributes": { -->
<!--       "user.department": "finance", -->
<!--       "user.security_level": 4 -->
<!--     } -->
<!--   }, -->
<!--   "resource": { -->
<!--     "id": "classified_doc_456", -->
<!--     "attributes": { -->
<!--       "resource.classification": "top_secret", -->
<!--       "resource.owner": "security_team" -->
<!--     } -->
<!--   }, -->
<!--   "action": {"name": "read"}, -->
<!--   "environment": { -->
<!--     "time": {"business_hours": true} -->
<!--   }, -->
<!--   "approval_context": { -->
<!--     "justification": "Need access for quarterly audit", -->
<!--     "urgency": "medium", -->
<!--     "business_reason": "Compliance requirement" -->
<!--   } -->
<!-- } -->
<!---->
<!-- Response: 202 Accepted -->
<!-- { -->
<!--   "decision": "PENDING_APPROVAL", -->
<!--   "evaluation_id": "eval_pending_001", -->
<!--   "approval_request": { -->
<!--     "id": "approval_req_001", -->
<!--     "status": "pending", -->
<!--     "approvers": [ -->
<!--       { -->
<!--         "id": "mgr_789", -->
<!--         "name": "Jane Manager", -->
<!--         "email": "jane.manager@company.com", -->
<!--         "role": "security_manager" -->
<!--       } -->
<!--     ], -->
<!--     "justification": "Need access for quarterly audit", -->
<!--     "expires_at": "2025-01-16T14:30:00Z", -->
<!--     "approval_url": "https://portal.company.com/approvals/approval_req_001" -->
<!--   }, -->
<!--   "estimated_approval_time": "2-4 hours", -->
<!--   "notification_sent": true -->
<!-- } -->
<!---->
<!-- ## Get Evaluation Status ->
<!-- GET /evaluate/status/{evaluation_id} -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "evaluation_id": "eval_pending_001", -->
<!--   "status": "approved", -->
<!--   "final_decision": "ALLOW", -->
<!--   "approved_by": { -->
<!--     "id": "mgr_789", -->
<!--     "name": "Jane Manager", -->
<!--     "approved_at": "2025-01-15T16:45:00Z" -->
<!--   }, -->
<!--   "approval_comments": "Approved for audit purposes. Access expires in 48 hours.", -->
<!--   "conditional_access": { -->
<!--     "expires_at": "2025-01-17T16:45:00Z", -->
<!--     "additional_obligations": [ -->
<!--       { -->
<!--         "type": "TIME_LIMITED_ACCESS", -->
<!--         "parameters": {"max_duration_hours": 48} -->
<!--       } -->
<!--     ] -->
<!--   } -->
<!-- } -->
<!---->
<!-- ## Policy Impact Simulation ->
<!-- POST /evaluate/simulate -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "policy_changes": [ -->
<!--     { -->
<!--       "action": "modify", -->
<!--       "policy_id": "550e8400-e29b-41d4-a716-446655440001", -->
<!--       "changes": { -->
<!--         "rule": { -->
<!--           "and": [ -->
<!--             {"user.security_level": {"gte": 6}}, -->
<!--             {"time.current_hour": {"between": [9, 17]}} -->
<!--           ] -->
<!--         } -->
<!--       } -->
<!--     } -->
<!--   ], -->
<!--   "test_scenarios": [ -->
<!--     { -->
<!--       "subject_attributes": { -->
<!--         "user.security_level": 5, -->
<!--         "user.department": "finance" -->
<!--       }, -->
<!--       "resource_attributes": { -->
<!--         "resource.classification": "financial" -->
<!--       }, -->
<!--       "action": "read", -->
<!--       "environment_attributes": { -->
<!--         "time.current_hour": 14 -->
<!--       } -->
<!--     } -->
<!--   ] -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "simulation_id": "sim_001", -->
<!--   "results": [ -->
<!--     { -->
<!--       "scenario": 1, -->
<!--       "current_decision": "ALLOW", -->
<!--       "simulated_decision": "DENY",  -->
<!--       "impact": "Access would be revoked", -->
<!--       "affected_policies": [ -->
<!--         { -->
<!--           "policy_id": "550e8400-e29b-41d4-a716-446655440001", -->
<!--           "current_match": true, -->
<!--           "simulated_match": false, -->
<!--           "reason": "user.security_level 5 < 6 (new requirement)" -->
<!--         } -->
<!--       ] -->
<!--     } -->
<!--   ], -->
<!--   "impact_summary": { -->
<!--     "total_scenarios": 1, -->
<!--     "no_change": 0, -->
<!--     "new_allows": 0, -->
<!--     "new_denies": 1, -->
<!--     "estimated_user_impact": 156 -->
<!--   } -->
<!-- } -->
<!---->
<!-- ## Get Cached Evaluation ->
<!-- GET /evaluate/cache/{cache_key} -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "cache_hit": true, -->
<!--   "decision": "ALLOW", -->
<!--   "cached_at": "2025-01-15T14:25:00Z", -->
<!--   "expires_at": "2025-01-15T15:25:00Z", -->
<!--   "evaluation_metadata": { -->
<!--     "original_evaluation_time_ms": 15.2, -->
<!--     "cache_retrieval_time_ms": 0.8 -->
<!--   } -->
<!-- } -->
<!---->
<!-- ## Clear Evaluation Cache ->
<!-- DELETE /evaluate/cache -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "cache_keys": ["user_123:resource_789:read"], -->
<!--   "clear_all": false, -->
<!--   "reasons": ["policy_updated", "user_attributes_changed"] -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "cleared_entries": 1, -->
<!--   "cache_stats": { -->
<!--     "remaining_entries": 15420, -->
<!--     "hit_rate": 0.87, -->
<!--     "avg_retrieval_time_ms": 1.2 -->
<!--   } -->
<!-- } -->
<!-- ``` -->
