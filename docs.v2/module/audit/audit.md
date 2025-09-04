# Audit Log Module Documentation

## Overview

The Audit Log module provides  event tracking, security monitoring, and compliance reporting capabilities for multi-tenant applications. It captures detailed records of user actions, system events, and security-related activities with advanced analytics and forensic investigation features.

## Table of Contents

1. [Architecture](#architecture)
2. [Core Concepts](#core-concepts)
3. [Data Model](#data-model)
4. [Event Types and Categories](#event-types-and-categories)
5. [Risk Scoring System](#risk-scoring-system)
6. [Query Operations](#query-operations)
7. [Analytics and Reporting](#analytics-and-reporting)
8. [Security Features](#security-features)
9. [Compliance and Governance](#compliance-and-governance)
10. [Performance Considerations](#performance-considerations)
11. [Best Practices](#best-practices)
12. [API Reference](#api-reference)

## Architecture

The Audit Log module is built using:
- **Go** for high-performance service implementation
- **SQLC** for type-safe SQL query generation
- **PGX** for PostgreSQL database connectivity

### System Design

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Application   │───▶│   Audit Service  │───▶│   PostgreSQL    │
│   Components    │    │     (Go)         │    │   Database      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                               │
                               ▼
                       ┌──────────────────┐
                       │   SQLC Generated │
                       │     Queries      │
                       └──────────────────┘
```

### Multi-Tenant Isolation

- **Tenant-scoped operations**: All queries include tenant isolation
- **Row-level security**: Database-level tenant separation
- **Zero cross-tenant leakage**: Guaranteed data isolation

## Core Concepts

### Audit Event Lifecycle

1. **Event Creation**: System or user action triggers event logging
2. **Risk Assessment**: Automatic risk scoring based on context
3. **Categorization**: Events classified by type and severity
4. **Storage**: Persistent storage with metadata enrichment
5. **Analysis**: Real-time and batch analytics processing
6. **Retention**: Automated cleanup based on retention policies

### Key Entities

- **Audit Events**: Core log entries with  metadata
- **Users**: Actors performing actions (human users, service accounts)
- **Entities**: Business objects being accessed or modified
- **Resources**: System resources (files, endpoints, services)
- **Sessions**: User session contexts for event correlation

## Data Model

### Primary Audit Event Structure

```go
type AuditEvent struct {
    ID               UUID           // Unique event identifier
    TenantID         UUID           // Tenant isolation key
    UserID           *UUID          // Acting user (nullable for system events)
    EventType        string         // Specific action performed
    EventCategory    string         // High-level categorization
    Severity         string         // Event importance level
    TargetUserID     *UUID          // Target of admin actions
    EntityID         *UUID          // Business entity reference
    ResourceID       *UUID          // System resource reference
    ActionID         *UUID          // Specific action reference
    RoleID           *UUID          // User role at event time
    PermissionID     *UUID          // Permission being exercised
    Decision         *string        // Access control decision (ALLOW/DENY)
    Reason           *string        // Decision rationale
    RiskScore        *int           // Calculated risk score (0-100)
    Context          JSONB          // Additional event metadata
    IPAddress        *inet          // Source IP address
    UserAgent        *string        // Client user agent
    SessionID        *UUID          // User session identifier
    ComplianceFlags  JSONB          // Compliance-related tags
    CreatedAt        timestamp      // Event timestamp
}
```

### Context Data Structure

The `Context` field supports structured metadata:

```json
{
    "request_id": "req_123456789",
    "api_endpoint": "/api/v1/users",
    "http_method": "POST",
    "response_status": 201,
    "data_classification": "PII",
    "geographic_location": "US-WEST",
    "device_fingerprint": "fp_abcdef123",
    "custom_attributes": {
        "department": "engineering",
        "project": "audit-system"
    }
}
```

## Event Types and Categories

### Event Categories

| Category | Description | Examples |
|----------|-------------|----------|
| `ACCESS` | Authentication and authorization | Login, logout, permission checks |
| `DATA` | Data access and manipulation | Read, create, update, delete operations |
| `ADMIN` | Administrative actions | User management, role assignments |
| `SYSTEM` | System-level events | Service starts, configuration changes |
| `SECURITY` | Security-related events | Failed logins, suspicious activity |
| `COMPLIANCE` | Compliance-specific events | Data export, audit trail access |

### Event Types by Category

#### ACCESS Events
- `LOGIN_SUCCESS` / `LOGIN_FAILURE`
- `LOGOUT`
- `PASSWORD_CHANGE`
- `MFA_CHALLENGE` / `MFA_SUCCESS` / `MFA_FAILURE`
- `PERMISSION_GRANT` / `PERMISSION_DENY`

#### DATA Events
- `DATA_READ` / `DATA_CREATE` / `DATA_UPDATE` / `DATA_DELETE`
- `FILE_UPLOAD` / `FILE_DOWNLOAD`
- `DATA_EXPORT` / `DATA_IMPORT`
- `BULK_OPERATION`

#### ADMIN Events
- `USER_CREATE` / `USER_UPDATE` / `USER_DELETE`
- `ROLE_ASSIGN` / `ROLE_REVOKE`
- `PERMISSION_MODIFY`
- `TENANT_CONFIGURATION`

#### SECURITY Events
- `SUSPICIOUS_LOGIN`
- `RATE_LIMIT_EXCEEDED`
- `SECURITY_POLICY_VIOLATION`
- `ANOMALOUS_BEHAVIOR`

### Severity Levels

| Severity | Description | Use Cases |
|----------|-------------|-----------|
| `LOW` | Routine operations | Normal data access, successful logins |
| `INFO` | Informational events | System notifications, routine admin actions |
| `WARN` | Potentially concerning | Failed permission checks, rate limiting |
| `HIGH` | Significant security events | Multiple failed logins, privilege escalation |
| `CRITICAL` | Severe security incidents | Data breaches, system compromises |

## Risk Scoring System

### Risk Score Calculation

Risk scores range from 0-100, calculated based on:

- **Base Event Risk**: Inherent risk of the event type
- **User Context**: User role, history, and current behavior
- **Environmental Factors**: Time, location, device characteristics
- **Historical Patterns**: Deviation from normal behavior

### Risk Factors

| Factor | Weight | Examples |
|--------|--------|----------|
| Event Type | 30% | Administrative actions score higher |
| User Behavior | 25% | Unusual access patterns, new locations |
| Time Context | 15% | After-hours access, weekend activity |
| Resource Sensitivity | 20% | PII access, financial data |
| Geographic Context | 10% | Foreign IP addresses, VPN usage |

### Risk Thresholds

- **Low Risk (0-30)**: Normal operations, standard business activities
- **Medium Risk (31-60)**: Elevated attention, potential monitoring
- **High Risk (61-80)**: Investigation recommended, alert generation
- **Critical Risk (81-100)**: Immediate action required, security response

## Query Operations

### Basic Event Retrieval

#### Get Audit Events
```sql
-- Retrieve events with  filtering
SELECT * FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND created_at >= $start_time 
  AND created_at <= $end_time
  AND ($event_category IS NULL OR event_category = $event_category)
ORDER BY created_at DESC;
```

#### Get User History
```sql
-- Retrieve complete user audit trail
SELECT event_type, event_category, severity, decision, 
       risk_score, context, ip_address, created_at
FROM audit_log 
WHERE tenant_id = current_tenant_id() AND user_id = $user_id
ORDER BY created_at DESC;
```

### Security-Focused Queries

#### High-Risk Events
```sql
-- Identify high-risk security events
SELECT id, user_id, event_type, risk_score, reason, created_at
FROM audit_log 
WHERE tenant_id = current_tenant_id() 
  AND risk_score >= $min_risk_score
ORDER BY risk_score DESC, created_at DESC;
```

#### Failed Access Attempts
```sql
-- Track access control failures
SELECT user_id, event_type, ip_address, reason, created_at
FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND decision = 'DENY' 
  AND event_category = 'ACCESS'
  AND created_at >= $start_time;
```

### Entity and Resource Tracking

#### Entity-Specific Events
```sql
-- Track all events for specific business entities
SELECT user_id, event_type, decision, context, created_at
FROM audit_log 
WHERE tenant_id = current_tenant_id() 
  AND entity_id = $entity_id
ORDER BY created_at DESC;
```

#### Resource Access Patterns
```sql
-- Monitor resource access patterns
SELECT user_id, event_type, decision, created_at
FROM audit_log 
WHERE tenant_id = current_tenant_id() 
  AND resource_id = $resource_id
ORDER BY created_at DESC;
```

## Analytics and Reporting

### Statistical Analysis

#### Event Distribution by Category
```sql
-- Analyze event patterns by category
SELECT event_category,
       COUNT(*) AS event_count,
       COUNT(*) FILTER (WHERE decision = 'DENY') AS denied_count,
       AVG(risk_score) AS avg_risk_score,
       COUNT(DISTINCT user_id) AS unique_users
FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND created_at >= $start_time 
  AND created_at <= $end_time
GROUP BY event_category
ORDER BY event_count DESC;
```

#### Severity Analysis
```sql
-- Examine event severity distribution
SELECT severity,
       COUNT(*) AS event_count,
       COUNT(DISTINCT user_id) AS unique_users,
       AVG(risk_score) AS avg_risk_score
FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND created_at >= $start_time
GROUP BY severity
ORDER BY CASE severity
    WHEN 'CRITICAL' THEN 5
    WHEN 'HIGH' THEN 4
    WHEN 'WARN' THEN 3
    WHEN 'INFO' THEN 2
    WHEN 'LOW' THEN 1
END DESC;
```

### User Risk Profiling

####  Risk Assessment
```sql
-- Generate detailed user risk profile
SELECT user_id,
       COUNT(*) AS total_events,
       AVG(risk_score) AS avg_risk_score,
       MAX(risk_score) AS max_risk_score,
       COUNT(*) FILTER (WHERE decision = 'DENY') AS failed_attempts,
       COUNT(*) FILTER (WHERE severity IN ('HIGH', 'CRITICAL')) AS high_severity_events,
       COUNT(DISTINCT ip_address) AS unique_ips,
       COUNT(DISTINCT event_category) AS unique_categories,
       MIN(created_at) AS first_event,
       MAX(created_at) AS last_event,
       COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '24 hours') AS events_last_24h
FROM audit_log 
WHERE tenant_id = current_tenant_id() AND user_id = $user_id
GROUP BY user_id;
```

### Time-Series Analysis

#### Hourly Event Rates
```sql
-- Analyze event volume patterns
SELECT DATE_TRUNC('hour', created_at) AS hour_bucket,
       COUNT(*) AS event_count,
       COUNT(DISTINCT user_id) AS unique_users,
       AVG(risk_score) AS avg_risk_score,
       COUNT(*) FILTER (WHERE decision = 'DENY') AS denied_count
FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND created_at >= $start_time
GROUP BY DATE_TRUNC('hour', created_at)
ORDER BY hour_bucket DESC;
```

## Security Features

### Suspicious Activity Detection

#### IP-Based Analysis
```sql
-- Identify suspicious IP activity patterns
SELECT ip_address,
       COUNT(*) AS event_count,
       COUNT(DISTINCT user_id) AS unique_users,
       COUNT(*) FILTER (WHERE decision = 'DENY') AS denied_count,
       AVG(risk_score) AS avg_risk_score
FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND (decision = 'DENY' OR risk_score >= $min_risk_score)
GROUP BY ip_address
HAVING COUNT(*) >= $min_event_count
ORDER BY avg_risk_score DESC;
```

#### Anomaly Detection
```sql
-- Detect unusual user behavior patterns
WITH user_baseline AS (
    SELECT user_id, 
           AVG(risk_score) AS baseline_risk,
           STDDEV(risk_score) AS risk_deviation
    FROM audit_log 
    WHERE tenant_id = current_tenant_id()
      AND created_at >= $baseline_start
    GROUP BY user_id
)
SELECT al.user_id, al.event_type, al.risk_score, al.created_at,
       ABS(al.risk_score - ub.baseline_risk) AS deviation
FROM audit_log al
JOIN user_baseline ub ON al.user_id = ub.user_id
WHERE al.tenant_id = current_tenant_id()
  AND ABS(al.risk_score - ub.baseline_risk) > (2 * ub.risk_deviation)
ORDER BY deviation DESC;
```

### Access Control Effectiveness

#### Permission Analysis
```sql
-- Evaluate access control decision patterns
SELECT user_id, resource_id, permission_id,
       COUNT(*) AS total_attempts,
       COUNT(*) FILTER (WHERE decision = 'ALLOW') AS allowed_attempts,
       COUNT(*) FILTER (WHERE decision = 'DENY') AS denied_attempts,
       ROUND(COUNT(*) FILTER (WHERE decision = 'DENY')::NUMERIC / COUNT(*) * 100, 2) AS denial_rate_pct
FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND event_category = 'ACCESS'
  AND decision IS NOT NULL
GROUP BY user_id, resource_id, permission_id
ORDER BY denial_rate_pct DESC;
```

## Compliance and Governance

### Regulatory Compliance Support

#### Data Access Tracking
```sql
-- Track personal data access for GDPR compliance
SELECT user_id, event_type, context, created_at
FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND event_category = 'DATA'
  AND (context ? 'personal_data' OR context ? 'pii')
ORDER BY created_at DESC;
```

#### Compliance Flag Management
```sql
-- Identify events requiring compliance attention
SELECT id, user_id, event_type, compliance_flags, created_at
FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND compliance_flags ? $compliance_flag
ORDER BY created_at DESC;
```

### Administrative Oversight

#### Admin Action Monitoring
```sql
-- Track administrative actions on users
SELECT user_id AS admin_user, target_user_id, 
       event_type, decision, reason, created_at
FROM audit_log 
WHERE tenant_id = current_tenant_id()
  AND target_user_id IS NOT NULL
  AND event_category = 'ADMIN'
ORDER BY created_at DESC;
```

## Performance Considerations

### Indexing Strategy

```sql
-- Recommended indexes for optimal performance
CREATE INDEX CONCURRENTLY idx_audit_log_tenant_created 
    ON audit_log (tenant_id, created_at DESC);

CREATE INDEX CONCURRENTLY idx_audit_log_tenant_user_created 
    ON audit_log (tenant_id, user_id, created_at DESC);

CREATE INDEX CONCURRENTLY idx_audit_log_tenant_category_created 
    ON audit_log (tenant_id, event_category, created_at DESC);

CREATE INDEX CONCURRENTLY idx_audit_log_tenant_risk_score 
    ON audit_log (tenant_id, risk_score DESC) 
    WHERE risk_score >= 70;

-- JSON index for context queries
CREATE INDEX CONCURRENTLY idx_audit_log_context_gin 
    ON audit_log USING GIN (context);
```

### Query Optimization

- **Time-range queries**: Always include time bounds to leverage time-based partitioning
- **Tenant isolation**: Ensure all queries include tenant_id for optimal performance
- **Pagination**: Use LIMIT and OFFSET for large result sets
- **Selective filtering**: Apply most restrictive filters first

### Data Retention and Archival

```sql
-- Archive old events before deletion
WITH old_events AS (
    SELECT id, event_type, user_id, created_at
    FROM audit_log 
    WHERE tenant_id = current_tenant_id()
      AND created_at < $cutoff_date
)
-- Move to archive table then delete
DELETE FROM audit_log 
WHERE id IN (SELECT id FROM old_events);
```

## Best Practices

### Event Logging Guidelines

1. ** Context**: Include relevant metadata in context fields
2. **Consistent Categorization**: Use standardized event types and categories
3. **Appropriate Severity**: Match severity to business impact
4. **Rich IP Context**: Capture network information for security analysis
5. **Session Correlation**: Link related events through session IDs

### Security Considerations

1. **Sensitive Data Protection**: Never log passwords, tokens, or PII in plain text
2. **Risk Score Calibration**: Regularly review and adjust risk scoring algorithms
3. **Real-time Monitoring**: Implement alerts for high-risk events
4. **Access Control**: Restrict audit log access to authorized personnel only
5. **Tamper Evidence**: Implement integrity checks for audit data

### Performance Best Practices

1. **Batch Processing**: Use bulk operations for high-volume scenarios
2. **Asynchronous Logging**: Implement async event logging to avoid blocking operations
3. **Data Lifecycle Management**: Implement automated archival and cleanup
4. **Query Optimization**: Use appropriate indexes and query patterns
5. **Resource Monitoring**: Monitor database performance and storage usage

### Compliance Best Practices

1. **Retention Policies**: Implement appropriate data retention based on regulations
2. **Access Logging**: Log all access to audit data itself
3. **Regular Auditing**: Perform periodic reviews of audit log integrity
4. **Documentation**: Maintain clear documentation of event meanings and processes
5. **Change Management**: Track all changes to audit configuration and policies

## API Reference

### Core Operations

#### Create Audit Event
```go
func (s *AuditService) CreateAuditEvent(ctx context.Context, 
    tenantID uuid.UUID, req CreateAuditEventRequest) (*AuditEvent, error)
```

#### Query Events
```go
func (s *AuditService) GetAuditEvents(ctx context.Context, 
    tenantID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)
```

#### User-Specific Queries
```go
func (s *AuditService) GetUserAuditHistory(ctx context.Context, 
    tenantID, userID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)

func (s *AuditService) GetUserRiskProfile(ctx context.Context, 
    tenantID, userID uuid.UUID, startTime *time.Time) (*UserRiskProfile, error)
```

#### Security Operations
```go
func (s *AuditService) GetHighRiskEvents(ctx context.Context, 
    tenantID uuid.UUID, minRiskScore int, filters AuditEventFilters) ([]AuditEvent, error)

func (s *AuditService) GetFailedAccessAttempts(ctx context.Context, 
    tenantID uuid.UUID, startTime, endTime time.Time, 
    userID *uuid.UUID, ipAddress *string, limit, offset int) ([]AuditEvent, error)
```

#### Analytics Operations
```go
func (s *AuditService) GetAuditStatsByCategory(ctx context.Context, 
    tenantID uuid.UUID, startTime, endTime time.Time, 
    severity *string) ([]AuditStatsByCategory, error)

func (s *AuditService) GetAuditLogHealth(ctx context.Context, 
    tenantID uuid.UUID) (*AuditLogHealth, error)
```

#### Maintenance Operations
```go
func (s *AuditService) UpdateEventRiskScore(ctx context.Context, 
    tenantID, eventID uuid.UUID, riskScore int) error

func (s *AuditService) DeleteOldAuditEvents(ctx context.Context, 
    tenantID uuid.UUID, cutoffDate time.Time) error
```

### Error Handling

All operations return structured errors with appropriate context:

- **Validation Errors**: Invalid input parameters or data
- **Authorization Errors**: Insufficient permissions for operations
- **Not Found Errors**: Requested resources don't exist
- **Conflict Errors**: Resource conflicts or constraint violations
- **Internal Errors**: Database or system-level failures

### Response Formats

All responses follow consistent patterns with appropriate HTTP status codes and structured JSON responses containing relevant data and metadata.

---

This documentation provides a guide to the Audit Log module, covering all aspects from basic usage to advanced security analytics and compliance features. The module is designed to provide enterprise-grade audit capabilities while maintaining high performance and strict tenant isolation.
