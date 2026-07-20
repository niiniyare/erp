> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Context & Attribute Management API

## Overview

The Context & Attribute Management API provides real-time attribute retrieval and context management for accurate policy evaluation. This service manages subject, resource, and environment attributes with high-performance caching and validation.

**Goa Service**: `context-attribute-management`  
**Base Path**: `/api/v1/attributes`

## Business Value

- **Real-time Context**: Dynamic attribute updates for immediate policy reflection
- **High Performance**: Sub-10ms attribute retrieval with intelligent caching
- **Data Integrity**:  validation and freshness tracking
- **Flexible Sources**: Integration with multiple attribute providers

## Performance Targets

- **Attribute Retrieval**: <10ms for 95% of requests
- **Cache Hit Rate**: >90% for frequently accessed attributes
- **Validation Time**: <5ms for attribute validation
- **Update Propagation**: <1s for critical attribute changes

## API Endpoints

### Subject Attribute Management

#### Set/Update Subject Attributes

**Action**: `SetSubjectAttributes`

```yaml
PUT /subjects/{subject_id}
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Path Parameters:
- subject_id: string (required) - Subject identifier

Request Body:
{
  "attributes": {
    "user.department": "finance",
    "user.security_level": 7,
    "user.employment_status": "active",
    "user.roles": ["financial_analyst", "report_viewer"],
    "user.manager_id": "mgr_456",
    "user.hire_date": "2020-03-15",
    "user.last_training_date": "2024-12-01",
    "user.clearance": "SECRET",
    "user.location": "headquarters",
    "user.cost_center": "FIN001"
  },
  "metadata": {
    "source": "hr_system",
    "updated_by": "hr_sync_service",
    "confidence_score": 1.0,
    "last_verified": "2025-01-15T14:30:00Z",
    "expires_at": "2025-01-16T14:30:00Z",
    "data_classification": "internal"
  },
  "options": {
    "merge_strategy": "overwrite",
    "validate_attributes": true,
    "invalidate_cache": true,
    "notify_subscribers": true,
    "create_audit_trail": true
  }
}

Response: 200 OK
{
  "subject_id": "user_123",
  "operation": "update",
  "completed_at": "2025-01-15T14:30:00Z",
  "attributes_summary": {
    "total_attributes": 10,
    "attributes_updated": 7,
    "attributes_added": 3,
    "attributes_removed": 0,
    "unchanged_attributes": 0
  },
  "validation_results": {
    "valid": true,
    "warnings": [
      {
        "attribute": "user.last_training_date",
        "message": "Training date is over 30 days old",
        "severity": "warning"
      }
    ],
    "errors": []
  },
  "cache_impact": {
    "entries_invalidated": 45,
    "policies_affected": 12,
    "dependent_evaluations_cleared": 156
  },
  "propagation": {
    "notifications_sent": 3,
    "downstream_services": ["policy_engine", "audit_service"],
    "propagation_time_ms": 234
  },
  "performance": {
    "validation_time_ms": 4.2,
    "storage_time_ms": 8.7,
    "cache_update_time_ms": 2.1
  }
}
```

#### Get Subject Attributes

**Action**: `GetSubjectAttributes`

```yaml
GET /subjects/{subject_id}?include_metadata=true&include_derived=true&fresh_only=false
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Path Parameters:
- subject_id: string (required)

Query Parameters:
- include_metadata: boolean (default: false) - Include attribute metadata
- include_derived: boolean (default: false) - Include computed attributes
- fresh_only: boolean (default: false) - Only return non-expired attributes
- attribute_filter: string - Comma-separated attribute names
- source_filter: string - Filter by attribute source

Response: 200 OK
{
  "subject_id": "user_123",
  "subject_type": "user",
  "retrieved_at": "2025-01-15T14:30:00Z",
  "attributes": {
    "user.department": "finance",
    "user.security_level": 7,
    "user.employment_status": "active",
    "user.roles": ["financial_analyst", "report_viewer"],
    "user.manager_id": "mgr_456",
    "user.hire_date": "2020-03-15",
    "user.last_training_date": "2024-12-01",
    "user.clearance": "SECRET",
    "user.years_of_service": 4.8,
    "user.is_manager": false,
    "user.location": "headquarters",
    "user.cost_center": "FIN001"
  },
  "derived_attributes": {
    "user.experience_level": "senior",
    "user.risk_profile": "low",
    "user.clearance_expired": false,
    "user.training_current": false,
    "user.access_level": "standard",
    "user.emergency_contact_verified": true
  },
  "attribute_metadata": {
    "user.department": {
      "source": "hr_system",
      "last_updated": "2025-01-15T09:00:00Z",
      "confidence": 1.0,
      "expires_at": "2025-01-16T09:00:00Z",
      "verification_method": "system_sync",
      "data_classification": "internal"
    },
    "user.security_level": {
      "source": "security_system",
      "last_updated": "2025-01-10T15:30:00Z",
      "confidence": 1.0,
      "expires_at": null,
      "verification_method": "manual_assignment",
      "data_classification": "confidential"
    },
    "user.years_of_service": {
      "source": "computed",
      "last_computed": "2025-01-15T14:30:00Z",
      "computation_method": "hire_date_calculation",
      "accuracy": 0.99
    }
  },
  "freshness_status": {
    "fresh_attributes": 10,
    "stale_attributes": 1,
    "expired_attributes": 0,
    "unknown_freshness": 0
  },
  "cache_info": {
    "cache_hit": true,
    "cache_age_seconds": 145,
    "cache_ttl_seconds": 1655
  }
}
```

#### Bulk Subject Attribute Update

**Action**: `BulkUpdateSubjectAttributes`

```yaml
POST /subjects/bulk-update
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "updates": [
    {
      "subject_id": "user_123",
      "attributes": {
        "user.security_level": 8,
        "user.last_training_date": "2025-01-10"
      },
      "metadata": {
        "source": "security_review",
        "updated_by": "security_admin"
      }
    },
    {
      "subject_id": "user_456",
      "attributes": {
        "user.department": "hr",
        "user.security_level": 4,
        "user.location": "remote"
      },
      "metadata": {
        "source": "hr_system",
        "updated_by": "hr_sync"
      }
    },
    {
      "subject_id": "user_789",
      "attributes": {
        "user.employment_status": "terminated",
        "user.termination_date": "2025-01-15"
      },
      "metadata": {
        "source": "hr_system",
        "updated_by": "hr_admin"
      }
    }
  ],
  "options": {
    "validate_all": true,
    "fail_on_error": false,
    "batch_size": 100,
    "parallel_processing": true,
    "invalidate_cache": true
  }
}

Response: 200 OK
{
  "bulk_update_id": "bulk_550e8400-e29b-41d4-a716-446655440000",
  "started_at": "2025-01-15T14:30:00Z",
  "completed_at": "2025-01-15T14:30:01.234Z",
  "summary": {
    "total_subjects": 3,
    "successful_updates": 3,
    "failed_updates": 0,
    "warnings": 1,
    "total_attributes_updated": 8
  },
  "results": [
    {
      "subject_id": "user_123",
      "status": "success",
      "attributes_updated": 2,
      "processing_time_ms": 156.7
    },
    {
      "subject_id": "user_456",
      "status": "success",
      "attributes_updated": 3,
      "processing_time_ms": 142.3
    },
    {
      "subject_id": "user_789",
      "status": "success",
      "attributes_updated": 2,
      "processing_time_ms": 189.2,
      "warnings": [
        {
          "message": "Termination triggers automatic access revocation",
          "severity": "info"
        }
      ]
    }
  ],
  "performance_metrics": {
    "total_processing_time_ms": 488.2,
    "avg_processing_time_ms": 162.7,
    "parallel_efficiency": 0.89,
    "cache_invalidations": 156
  },
  "downstream_impact": {
    "policies_affected": 23,
    "evaluations_invalidated": 567,
    "notifications_triggered": 8
  }
}
```

### Resource Attribute Management

#### Set Resource Attributes

**Action**: `SetResourceAttributes`

```yaml
PUT /resources/{resource_id}
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "attributes": {
    "resource.classification": "financial",
    "resource.type": "report",
    "resource.sensitivity": "high",
    "resource.owner": "finance_team",
    "resource.created_date": "2025-01-10T00:00:00Z",
    "resource.data_retention_years": 7,
    "resource.geographic_restriction": ["US", "CA"],
    "resource.compliance_flags": ["SOX", "PCI"],
    "resource.size_mb": 2.5,
    "resource.format": "pdf",
    "resource.encryption_status": "encrypted",
    "resource.backup_location": "secure_vault"
  },
  "metadata": {
    "source": "document_management_system",
    "auto_derived": false,
    "classification_confidence": 0.95,
    "last_scanned": "2025-01-15T12:00:00Z"
  },
  "options": {
    "validate_compliance": true,
    "auto_classify": false,
    "update_related_resources": true
  }
}

Response: 200 OK
{
  "resource_id": "resource_789",
  "operation": "update",
  "completed_at": "2025-01-15T14:30:00Z",
  "attributes_updated": 12,
  "compliance_check": {
    "passed": true,
    "flags_verified": ["SOX", "PCI"],
    "recommendations": [
      "Consider adding GDPR classification for EU access"
    ]
  },
  "auto_classification": {
    "enabled": false,
    "manual_override": true
  },
  "related_updates": {
    "parent_resources": 1,
    "child_resources": 3,
    "linked_resources": 0
  }
}
```

#### Get Resource Attributes

**Action**: `GetResourceAttributes`

```yaml
GET /resources/{resource_id}?include_computed=true&include_relationships=true
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Response: 200 OK
{
  "resource_id": "resource_789",
  "resource_type": "document",
  "retrieved_at": "2025-01-15T14:30:00Z",
  "attributes": {
    "resource.classification": "financial",
    "resource.type": "report",
    "resource.sensitivity": "high",
    "resource.owner": "finance_team",
    "resource.created_date": "2025-01-10T00:00:00Z",
    "resource.size_mb": 2.5,
    "resource.last_modified": "2025-01-12T10:15:00Z",
    "resource.format": "pdf",
    "resource.encryption_status": "encrypted",
    "resource.compliance_flags": ["SOX", "PCI"],
    "resource.geographic_restriction": ["US", "CA"]
  },
  "computed_attributes": {
    "resource.age_days": 5,
    "resource.requires_approval": true,
    "resource.access_frequency": "high",
    "resource.risk_score": 78,
    "resource.compliance_score": 0.94,
    "resource.popularity_index": 0.87
  },
  "relationships": {
    "parent_resources": [
      {
        "id": "resource_parent_001",
        "type": "folder",
        "relationship": "contains"
      }
    ],
    "child_resources": [
      {
        "id": "resource_child_001",
        "type": "attachment",
        "relationship": "attached_to"
      }
    ],
    "related_resources": [
      {
        "id": "resource_related_001",
        "type": "report",
        "relationship": "version_of",
        "similarity_score": 0.89
      }
    ]
  },
  "access_statistics": {
    "total_accesses": 1547,
    "unique_users": 89,
    "last_accessed": "2025-01-15T13:45:00Z",
    "peak_access_hour": 14,
    "avg_session_duration_minutes": 23
  }
}
```

### Environment Context Management

#### Get Environment Context

**Action**: `GetEnvironmentContext`

```yaml
GET /context/environment?include_dynamic=true&include_predictions=false
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Query Parameters:
- include_dynamic: boolean (default: true) - Include real-time dynamic attributes
- include_predictions: boolean (default: false) - Include predictive attributes
- location_context: string - Specific location context
- time_context: string - Specific time context (ISO 8601)

Response: 200 OK
{
  "retrieved_at": "2025-01-15T14:30:00.123Z",
  "context_version": "2.1.5",
  "environment": {
    "time": {
      "current_time": "2025-01-15T14:30:00Z",
      "timezone": "UTC",
      "business_hours": true,
      "current_hour": 14,
      "current_minute": 30,
      "day_of_week": 3,
      "day_of_month": 15,
      "day_of_year": 15,
      "week_of_year": 3,
      "month": 1,
      "year": 2025,
      "is_weekend": false,
      "is_holiday": false,
      "holiday_name": null,
      "fiscal_quarter": "Q1",
      "fiscal_year": 2025
    },
    "system": {
      "load_average": 0.65,
      "active_users": 1247,
      "active_sessions": 2156,
      "maintenance_mode": false,
      "backup_running": false,
      "version": "2.1.5",
      "uptime_hours": 720.5,
      "cpu_utilization": 0.67,
      "memory_utilization": 0.72,
      "disk_utilization": 0.45
    },
    "security": {
      "threat_level": "low",
      "active_incidents": 0,
      "incident_history_24h": 2,
      "security_alert_level": "green",
      "last_security_scan": "2025-01-15T06:00:00Z",
      "firewall_status": "active",
      "ids_status": "active",
      "compliance_mode": "strict"
    },
    "network": {
      "corporate_network_status": "operational",
      "vpn_status": "operational",
      "internet_connectivity": "excellent",
      "bandwidth_utilization": 0.34,
      "external_threats_detected": 0,
      "dns_resolution_time_ms": 12.3,
      "network_latency_ms": 8.7
    },
    "location": {
      "primary_datacenter": "us-east-1",
      "backup_datacenter": "us-west-2",
      "active_offices": ["headquarters", "branch_1", "branch_2"],
      "emergency_mode": false,
      "business_continuity_status": "normal"
    }
  },
  "dynamic_attributes": {
    "env.peak_usage_period": false,
    "env.backup_window": false,
    "env.security_drill_active": false,
    "env.high_load_detected": false,
    "env.maintenance_scheduled": false,
    "env.incident_response_active": false,
    "env.compliance_audit_active": true,
    "env.change_freeze_active": false
  },
  "predictive_attributes": {
    "env.expected_load_increase": false,
    "env.maintenance_probability_6h": 0.05,
    "env.security_event_likelihood": 0.02,
    "env.peak_usage_prediction_2h": false
  },
  "cache_info": {
    "ttl_seconds": 60,
    "cache_key": "env_context_20250115_1430",
    "last_refresh": "2025-01-15T14:30:00Z"
  }
}
```

### Session Context Management

#### Set Session Context

**Action**: `SetSessionContext`

```yaml
PUT /context/sessions/{session_id}
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "session": {
    "id": "session_abc123",
    "user_id": "user_123",
    "created_at": "2025-01-15T09:00:00Z",
    "last_activity": "2025-01-15T14:25:00Z",
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "device_fingerprint": "fp_abc123xyz",
    "device_type": "desktop",
    "operating_system": "Windows 11",
    "browser": "Chrome 120.0",
    "mfa_verified": true,
    "mfa_verified_at": "2025-01-15T09:00:00Z",
    "mfa_method": "totp",
    "location": {
      "country": "US",
      "region": "California",
      "city": "San Francisco",
      "network": "corporate",
      "building": "headquarters",
      "floor": "5",
      "coordinates": {
        "latitude": 37.7749,
        "longitude": -122.4194
      }
    },
    "security_score": 0.95,
    "risk_indicators": [],
    "session_type": "interactive",
    "source_application": "web_portal"
  },
  "computed_attributes": {
    "session.duration_minutes": 325,
    "session.activity_level": "high",
    "session.location_consistent": true,
    "session.device_trusted": true,
    "session.behavior_normal": true
  }
}

Response: 200 OK
{
  "session_id": "session_abc123",
  "operation": "update",
  "completed_at": "2025-01-15T14:30:00Z",
  "context_updated": true,
  "security_assessment": {
    "risk_level": "low",
    "trust_score": 0.95,
    "anomalies_detected": 0,
    "location_verified": true,
    "device_verified": true,
    "behavior_score": 0.92
  },
  "recommendations": [],
  "monitoring": {
    "enhanced_monitoring": false,
    "alert_thresholds": "standard",
    "session_timeout_minutes": 480
  }
}
```

#### Get Session Context

**Action**: `GetSessionContext`

```yaml
GET /context/sessions/{session_id}?include_analytics=true&include_risk_assessment=true
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Response: 200 OK
{
  "session_id": "session_abc123",
  "user_id": "user_123",
  "status": "active",
  "created_at": "2025-01-15T09:00:00Z",
  "last_activity": "2025-01-15T14:25:00Z",
  "duration_minutes": 325,
  "attributes": {
    "session.mfa_verified": true,
    "session.mfa_method": "totp",
    "session.ip_address": "192.168.1.100",
    "session.device_type": "desktop",
    "session.operating_system": "Windows 11",
    "session.browser": "Chrome 120.0",
    "session.location.network": "corporate",
    "session.location.country": "US",
    "session.location.city": "San Francisco",
    "session.security_score": 0.95,
    "session.risk_level": "low",
    "session.device_trusted": true,
    "session.location_verified": true
  },
  "activity_summary": {
    "actions_performed": 47,
    "resources_accessed": 12,
    "policy_evaluations": 89,
    "last_action": "document_read",
    "last_action_time": "2025-01-15T14:25:00Z",
    "unusual_activity": false,
    "activity_pattern": "normal"
  },
  "risk_assessment": {
    "overall_risk_score": 15,
    "risk_factors": [
      {
        "factor": "location_change",
        "score": 5,
        "weight": 0.2,
        "description": "Minor location variance within expected range"
      },
      {
        "factor": "device_consistency",
        "score": 0,
        "weight": 0.3,
        "description": "Same device as previous sessions"
      }
    ],
    "behavioral_analysis": {
      "typing_pattern_match": 0.94,
      "click_pattern_match": 0.87,
      "navigation_pattern_match": 0.91,
      "time_pattern_match": 0.89
    },
    "anomalies": [],
    "recommendations": [
      "Continue standard monitoring",
      "No additional security measures required"
    ]
  },
  "performance_metrics": {
    "avg_response_time_ms": 234.5,
    "total_data_transferred_mb": 45.7,
    "error_count": 0,
    "timeout_count": 0
  }
}
```

### Attribute Query and Analytics

#### Query Attributes by Pattern

**Action**: `QueryAttributes`

```yaml
GET /query?pattern=user.security_*&subject_ids=user_123,user_456&include_metadata=true
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Query Parameters:
- pattern: string (required) - Attribute name pattern (supports wildcards)
- subject_ids: string - Comma-separated subject IDs
- resource_ids: string - Comma-separated resource IDs
- include_metadata: boolean (default: false)
- include_derived: boolean (default: false)
- fresh_only: boolean (default: false)
- limit: integer (default: 100, max: 1000)

Response: 200 OK
{
  "query": {
    "pattern": "user.security_*",
    "executed_at": "2025-01-15T14:30:00Z",
    "execution_time_ms": 8.3
  },
  "results": [
    {
      "subject_id": "user_123",
      "subject_type": "user",
      "attributes": {
        "user.security_level": 7,
        "user.security_clearance": "SECRET",
        "user.security_training_date": "2024-12-01",
        "user.security_incidents": 0,
        "user.security_score": 0.92
      },
      "metadata": {
        "user.security_level": {
          "source": "security_system",
          "last_updated": "2025-01-10T15:30:00Z",
          "confidence": 1.0
        },
        "user.security_clearance": {
          "source": "hr_system", 
          "last_updated": "2025-01-05T12:00:00Z",
          "expires_at": "2026-01-05T12:00:00Z"
        }
      }
    },
    {
      "subject_id": "user_456",
      "subject_type": "user", 
      "attributes": {
        "user.security_level": 4,
        "user.security_clearance": "CONFIDENTIAL",
        "user.security_training_date": "2024-11-15",
        "user.security_incidents": 1,
        "user.security_score": 0.78
      }
    }
  ],
  "summary": {
    "total_results": 2,
    "total_attributes": 10,
    "unique_attribute_names": 5,
    "cache_hit_rate": 0.67
  }
}
```

#### Attribute Health Check

**Action**: `GetAttributeHealth`

```yaml
GET /health?include_stats=true&include_performance=true
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Response: 200 OK
{
  "status": "healthy",
  "checked_at": "2025-01-15T14:30:00Z",
  "overall_health_score": 0.94,
  "statistics": {
    "total_subjects": 15420,
    "total_resources": 89234,
    "total_sessions": 2156,
    "total_attributes": 487231,
    "attributes_by_category": {
      "USER": 156891,
      "RESOURCE": 234567,
      "ENVIRONMENT": 12890,
      "SESSION": 82883
    },
    "attribute_freshness": {
      "fresh_attributes": 456789,
      "stale_attributes": 234,
      "expired_attributes": 12,
      "unknown_freshness": 567
    },
    "data_quality": {
      "valid_attributes": 486234,
      "invalid_attributes": 45,
      "validation_errors": 23,
      "consistency_score": 0.97
    }
  },
  "performance_metrics": {
    "avg_retrieval_time_ms": 8.1,
    "p95_retrieval_time_ms": 23.4,
    "p99_retrieval_time_ms": 45.7,
    "cache_hit_rate": 0.94,
    "cache_efficiency": 0.89,
    "throughput_per_second": 2456,
    "error_rate": 0.001
  },
  "health_indicators": {
    "attribute_freshness": {
      "status": "good",
      "score": 0.95,
      "details": "2% stale attributes within acceptable limits"
    },
    "cache_performance": {
      "status": "excellent",
      "score": 0.94,
      "details": "Cache hit rate exceeds target"
    },
    "response_time": {
      "status": "good", 
      "score": 0.92,
      "details": "Average response time within SLA"
    },
    "error_rate": {
      "status": "excellent",
      "score": 0.99,
      "details": "Error rate well below threshold"
    }
  },
  "recommendations": [
    "Consider cleaning up 234 stale attributes",
    "12 expired attributes need refresh",
    "Investigate 45 invalid attributes",
    "Monitor cache performance during peak hours"
  ],
  "alerts": [
    {
      "level": "warning",
      "message": "Stale attribute count increasing",
      "threshold": 200,
      "current_value": 234,
      "recommended_action": "Review attribute refresh schedules"
    }
  ]
}
```

## Goa DSL Implementation Notes

### Service Definition

```go
var _ = Service("context-attribute-management", func() {
    Description("Context and Attribute Management Service")
    
    HTTP(func() {
        Path("/api/v1/attributes")
        Header("X-Tenant-ID", String, "Tenant identifier", func() {
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
    })
    
    JWT(func() {
        Description("JWT authentication")
        Scope("attribute:read", "Read attributes")
        Scope("attribute:write", "Write attributes")
        Scope("context:read", "Read context")
        Scope("context:write", "Write context")
    })
})
```

### Type Definitions

```go
var AttributeSet = Type("AttributeSet", func() {
    Attribute("subject_id", String, "Subject identifier")
    Attribute("subject_type", String, "Subject type", func() {
        Enum("user", "service", "device", "application")
        Default("user")
    })
    Attribute("attributes", MapOf(String, Any), "Attribute key-value pairs")
    Attribute("metadata", AttributeMetadata, "Attribute metadata")
    Attribute("retrieved_at", String, "Retrieval timestamp", func() {
        Format(FormatDateTime)
    })
    
    Required("subject_id", "attributes")
})

var AttributeMetadata = Type("AttributeMetadata", func() {
    Attribute("source", String, "Attribute source system")
    Attribute("last_updated", String, "Last update timestamp", func() {
        Format(FormatDateTime)
    })
    Attribute("confidence", Float64, "Confidence score (0-1)")
    Attribute("expires_at", String, "Expiration timestamp", func() {
        Format(FormatDateTime)
    })
    Attribute("verification_method", String, "Verification method")
    Attribute("data_classification", String, "Data classification level")
})

var SessionContext = Type("SessionContext", func() {
    Attribute("id", String, "Session identifier")
    Attribute("user_id", String, "User identifier")
    Attribute("created_at", String, "Session creation time", func() {
        Format(FormatDateTime)
    })
    Attribute("last_activity", String, "Last activity time", func() {
        Format(FormatDateTime)
    })
    Attribute("ip_address", String, "IP address")
    Attribute("user_agent", String, "User agent string")
    Attribute("device_fingerprint", String, "Device fingerprint")
    Attribute("mfa_verified", Boolean, "MFA verification status")
    Attribute("location", LocationInfo, "Location information")
    Attribute("security_score", Float64, "Security score (0-1)")
    
    Required("id", "user_id", "created_at")
})

var EnvironmentContext = Type("EnvironmentContext", func() {
    Attribute("retrieved_at", String, "Context retrieval time", func() {
        Format(FormatDateTime)
    })
    Attribute("time", TimeContext, "Time-related attributes")
    Attribute("system", SystemContext, "System-related attributes")
    Attribute("security", SecurityContext, "Security-related attributes")
    Attribute("network", NetworkContext, "Network-related attributes")
    
    Required("retrieved_at", "time", "system")
})
```

### Caching Strategy

```go
type AttributeCacheConfig struct {
    SubjectAttributes struct {
        TTL          time.Duration `default:"30m"`
        MaxSize      int          `default:"50000"`
        EvictionPolicy string     `default:"LRU"`
    }
    ResourceAttributes struct {
        TTL          time.Duration `default:"1h"`
        MaxSize      int          `default:"100000"`
        EvictionPolicy string     `default:"LRU"`
    }
    EnvironmentContext struct {
        TTL          time.Duration `default:"1m"`
        MaxSize      int          `default:"1000"`
        EvictionPolicy string     `default:"TTL"`
    }
    SessionContext struct {
        TTL          time.Duration `default:"5m"`
        MaxSize      int          `default:"10000"`
        EvictionPolicy string     `default:"LRU"`
    }
}
```

### Performance Optimizations

- **Batch Operations**: Efficient bulk attribute updates and retrievals
- **Intelligent Caching**: Multi-level caching with invalidation strategies
- **Attribute Derivation**: Real-time computation of derived attributes
- **Data Compression**: Efficient storage and transmission of attribute data
- **Parallel Processing**: Concurrent attribute validation and updates


<!-- ##  4. Context & Attribute Management API ->
<!-- *Business Value*: Real-time attribute retrieval and context management for accurate policy evaluation. -->
<!---->
<!-- ```json  -->
<!-- # Context & Attribute Management API ->
<!---->
<!-- ## Set/Update Subject Attributes ->
<!-- PUT /attributes/subjects/{subject_id} -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "attributes": { -->
<!--     "user.department": "finance", -->
<!--     "user.security_level": 7, -->
<!--     "user.employment_status": "active", -->
<!--     "user.roles": ["financial_analyst", "report_viewer"], -->
<!--     "user.manager_id": "mgr_456", -->
<!--     "user.hire_date": "2020-03-15", -->
<!--     "user.last_training_date": "2024-12-01" -->
<!--   }, -->
<!--   "metadata": { -->
<!--     "source": "hr_system", -->
<!--     "updated_by": "hr_sync_service", -->
<!--     "confidence_score": 1.0, -->
<!--     "expires_at": "2025-01-16T14:30:00Z" -->
<!--   }, -->
<!--   "options": { -->
<!--     "merge_strategy": "overwrite", -->
<!--     "validate_attributes": true, -->
<!--     "invalidate_cache": true -->
<!--   } -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "subject_id": "user_123", -->
<!--   "attributes_updated": 7, -->
<!--   "attributes_added": 2, -->
<!--   "attributes_removed": 0, -->
<!--   "validation_results": { -->
<!--     "valid": true, -->
<!--     "warnings": [], -->
<!--     "errors": [] -->
<!--   }, -->
<!--   "cache_invalidation": { -->
<!--     "entries_cleared": 45, -->
<!--     "policies_affected": 12 -->
<!--   }, -->
<!--   "updated_at": "2025-01-15T14:30:00Z" -->
<!-- } -->
<!---->
<!-- ## Get Subject Attributes ->
<!-- GET /attributes/subjects/{subject_id}?include_metadata=true&include_derived=true -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "subject_id": "user_123", -->
<!--   "subject_type": "user", -->
<!--   "attributes": { -->
<!--     "user.department": "finance", -->
<!--     "user.security_level": 7, -->
<!--     "user.employment_status": "active", -->
<!--     "user.roles": ["financial_analyst", "report_viewer"], -->
<!--     "user.manager_id": "mgr_456", -->
<!--     "user.hire_date": "2020-03-15", -->
<!--     "user.years_of_service": 4.8, -->
<!--     "user.is_manager": false -->
<!--   }, -->
<!--   "derived_attributes": { -->
<!--     "user.experience_level": "senior", -->
<!--     "user.risk_profile": "low", -->
<!--     "user.clearance_expired": false -->
<!--   }, -->
<!--   "attribute_metadata": { -->
<!--     "user.department": { -->
<!--       "source": "hr_system", -->
<!--       "last_updated": "2025-01-15T09:00:00Z", -->
<!--       "confidence": 1.0, -->
<!--       "expires_at": "2025-01-16T09:00:00Z" -->
<!--     }, -->
<!--     "user.security_level": { -->
<!--       "source": "security_system", -->
<!--       "last_updated": "2025-01-10T15:30:00Z", -->
<!--       "confidence": 1.0, -->
<!--       "expires_at": null -->
<!--     } -->
<!--   }, -->
<!--   "last_updated": "2025-01-15T14:30:00Z" -->
<!-- } -->
<!---->
<!-- ## Bulk Attribute Update ->
<!-- POST /attributes/subjects/bulk -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "updates": [ -->
<!--     { -->
<!--       "subject_id": "user_123", -->
<!--       "attributes": { -->
<!--         "user.security_level": 8, -->
<!--         "user.last_training_date": "2025-01-10" -->
<!--       } -->
<!--     }, -->
<!--     { -->
<!--       "subject_id": "user_456",  -->
<!--       "attributes": { -->
<!--         "user.department": "hr", -->
<!--         "user.security_level": 4 -->
<!--       } -->
<!--     } -->
<!--   ], -->
<!--   "options": { -->
<!--     "validate_all": true, -->
<!--     "fail_on_error": false, -->
<!--     "batch_size": 100 -->
<!--   } -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "total_subjects": 2, -->
<!--   "successful_updates": 2, -->
<!--   "failed_updates": 0, -->
<!--   "results": [ -->
<!--     { -->
<!--       "subject_id": "user_123", -->
<!--       "status": "success", -->
<!--       "attributes_updated": 2 -->
<!--     }, -->
<!--     { -->
<!--       "subject_id": "user_456", -->
<!--       "status": "success",  -->
<!--       "attributes_updated": 2 -->
<!--     } -->
<!--   ], -->
<!--   "processing_time_ms": 125.5 -->
<!-- } -->
<!---->
<!-- ## Set Resource Attributes ->
<!-- PUT /attributes/resources/{resource_id} -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "attributes": { -->
<!--     "resource.classification": "financial", -->
<!--     "resource.type": "report", -->
<!--     "resource.sensitivity": "high", -->
<!--     "resource.owner": "finance_team", -->
<!--     "resource.created_date": "2025-01-10T00:00:00Z", -->
<!--     "resource.data_retention_years": 7, -->
<!--     "resource.geographic_restriction": ["US", "CA"] -->
<!--   }, -->
<!--   "metadata": { -->
<!--     "source": "document_management_system", -->
<!--     "auto_derived": false -->
<!--   } -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "resource_id": "resource_789", -->
<!--   "attributes_updated": 7, -->
<!--   "updated_at": "2025-01-15T14:30:00Z" -->
<!-- } -->
<!---->
<!-- ## Get Resource Attributes ->
<!-- GET /attributes/resources/{resource_id} -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "resource_id": "resource_789", -->
<!--   "resource_type": "document", -->
<!--   "attributes": { -->
<!--     "resource.classification": "financial", -->
<!--     "resource.type": "report", -->
<!--     "resource.sensitivity": "high", -->
<!--     "resource.owner": "finance_team", -->
<!--     "resource.created_date": "2025-01-10T00:00:00Z", -->
<!--     "resource.size_mb": 2.5, -->
<!--     "resource.last_modified": "2025-01-12T10:15:00Z" -->
<!--   }, -->
<!--   "computed_attributes": { -->
<!--     "resource.age_days": 5, -->
<!--     "resource.requires_approval": true, -->
<!--     "resource.compliance_flags": ["SOX", "PCI"] -->
<!--   }, -->
<!--   "last_updated": "2025-01-15T14:30:00Z" -->
<!-- } -->
<!---->
<!-- ## Get Environment Context ->
<!-- GET /context/environment?include_dynamic=true -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "environment": { -->
<!--     "time": { -->
<!--       "current_time": "2025-01-15T14:30:00Z", -->
<!--       "timezone": "UTC", -->
<!--       "business_hours": true, -->
<!--       "current_hour": 14, -->
<!--       "day_of_week": 3, -->
<!--       "is_weekend": false, -->
<!--       "is_holiday": false -->
<!--     }, -->
<!--     "system": { -->
<!--       "load_average": 0.65, -->
<!--       "active_users": 1247, -->
<!--       "maintenance_mode": false, -->
<!--       "version": "2.1.5" -->
<!--     }, -->
<!--     "security": { -->
<!--       "threat_level": "low", -->
<!--       "incident_active": false, -->
<!--       "security_alert_level": "green" -->
<!--     }, -->
<!--     "network": { -->
<!--       "corporate_network_status": "operational", -->
<!--       "vpn_status": "operational", -->
<!--       "external_threats_detected": 0 -->
<!--     } -->
<!--   }, -->
<!--   "dynamic_attributes": { -->
<!--     "env.peak_usage_period": false, -->
<!--     "env.backup_window": false, -->
<!--     "env.security_drill_active": false -->
<!--   }, -->
<!--   "retrieved_at": "2025-01-15T14:30:00.123Z", -->
<!--   "cache_ttl_seconds": 60 -->
<!-- } -->
<!---->
<!-- ## Set Session Context ->
<!-- PUT /context/sessions/{session_id} -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "session": { -->
<!--     "id": "session_abc123", -->
<!--     "user_id": "user_123", -->
<!--     "created_at": "2025-01-15T09:00:00Z", -->
<!--     "last_activity": "2025-01-15T14:25:00Z", -->
<!--     "ip_address": "192.168.1.100", -->
<!--     "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", -->
<!--     "device_fingerprint": "fp_abc123xyz", -->
<!--     "mfa_verified": true, -->
<!--     "mfa_verified_at": "2025-01-15T09:00:00Z", -->
<!--     "location": { -->
<!--       "country": "US", -->
<!--       "region": "California", -->
<!--       "city": "San Francisco", -->
<!--       "network": "corporate", -->
<!--       "building": "headquarters", -->
<!--       "floor": "5" -->
<!--     }, -->
<!--     "security_score": 0.95, -->
<!--     "risk_indicators": [] -->
<!--   } -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "session_id": "session_abc123", -->
<!--   "context_updated": true, -->
<!--   "security_assessment": { -->
<!--     "risk_level": "low", -->
<!--     "trust_score": 0.95, -->
<!--     "anomalies_detected": 0 -->
<!--   }, -->
<!--   "updated_at": "2025-01-15T14:30:00Z" -->
<!-- } -->
<!---->
<!-- ## Get Session Context ->
<!-- GET /context/sessions/{session_id} -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "session_id": "session_abc123", -->
<!--   "user_id": "user_123", -->
<!--   "active": true, -->
<!--   "created_at": "2025-01-15T09:00:00Z", -->
<!--   "last_activity": "2025-01-15T14:25:00Z", -->
<!--   "duration_minutes": 325, -->
<!--   "attributes": { -->
<!--     "session.mfa_verified": true, -->
<!--     "session.ip_address": "192.168.1.100", -->
<!--     "session.device_type": "desktop", -->
<!--     "session.location.network": "corporate", -->
<!--     "session.location.country": "US", -->
<!--     "session.security_score": 0.95, -->
<!--     "session.risk_level": "low" -->
<!--   }, -->
<!--   "activity_summary": { -->
<!--     "actions_performed": 47, -->
<!--     "resources_accessed": 12, -->
<!--     "last_action": "document_read", -->
<!--     "unusual_activity": false -->
<!--   } -->
<!-- } -->
<!---->
<!-- ## Query Attributes by Pattern ->
<!-- GET /attributes/query?pattern=user.security_*&subject_ids=user_123,user_456 -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "results": [ -->
<!--     { -->
<!--       "subject_id": "user_123", -->
<!--       "attributes": { -->
<!--         "user.security_level": 7, -->
<!--         "user.security_clearance": "SECRET", -->
<!--         "user.security_training_date": "2024-12-01" -->
<!--       } -->
<!--     }, -->
<!--     { -->
<!--       "subject_id": "user_456", -->
<!--       "attributes": { -->
<!--         "user.security_level": 4, -->
<!--         "user.security_clearance": "CONFIDENTIAL" -->
<!--       } -->
<!--     } -->
<!--   ], -->
<!--   "total_results": 2, -->
<!--   "query_time_ms": 8.3 -->
<!-- } -->
<!---->
<!-- ## Attribute Health Check ->
<!-- GET /attributes/health?include_stats=true -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "status": "healthy", -->
<!--   "statistics": { -->
<!--     "total_subjects": 15420, -->
<!--     "total_resources": 89234, -->
<!--     "total_attributes": 487231, -->
<!--     "attributes_by_category": { -->
<!--       "USER": 156891, -->
<!--       "RESOURCE": 234567, -->
<!--       "ENVIRONMENT": 12890, -->
<!--       "SESSION": 82883 -->
<!--     }, -->
<!--     "stale_attributes": 234, -->
<!--     "expired_attributes": 12, -->
<!--     "cache_hit_rate": 0.94 -->
<!--   }, -->
<!--   "health_indicators": { -->
<!--     "attribute_freshness": "good", -->
<!--     "cache_performance": "excellent",  -->
<!--     "response_time": "good", -->
<!--     "error_rate": "low" -->
<!--   }, -->
<!--   "recommendations": [ -->
<!--     "Consider cleaning up 234 stale attributes", -->
<!--     "12 expired attributes need refresh" -->
<!--   ], -->
<!--   "checked_at": "2025-01-15T14:30:00Z" -->
<!-- } -->
<!-- ``` -->
