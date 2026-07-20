> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# User Analytics & Behavior API

The User Analytics & Behavior API provides  user behavior analysis, risk assessment, and personalized insights to enhance security and user experience.

##  Overview

The User Analytics & Behavior API enables:

- **Behavioral Pattern Analysis**: Login patterns, activity trends, access behaviors
- **Risk Assessment**: Real-time risk scoring and threat detection
- **Anomaly Detection**: Unusual behavior identification and alerting
- **Personalized Insights**: Productivity recommendations and security suggestions
- **Predictive Analytics**: Future behavior prediction and risk forecasting
- **Baseline Establishment**: Normal behavior pattern learning

##  Base URL

```
http://localhost:8080/api/v1/analytics
```

##  Analytics Types

| Type | Description | Use Case |
|------|-------------|----------|
| `Behavior Pattern` |  user behavior analysis | Security baselines, productivity insights |
| `Risk Assessment` | Real-time risk scoring | Access control decisions, security alerts |
| `Anomaly Detection` | Unusual behavior identification | Threat detection, compliance monitoring |
| `Personalized Insights` | User-specific recommendations | Productivity enhancement, security awareness |

##  Risk Levels

| Level | Score Range | Description | Actions |
|-------|-------------|-------------|---------|
| `LOW` | 0-25 | Normal behavior | Standard access |
| `MEDIUM` | 26-50 | Slightly unusual |  monitoring |
| `HIGH` | 51-75 | Concerning behavior | Additional verification |
| `CRITICAL` | 76-100 | High-risk behavior | Immediate action required |

##  Quick Start

### 1. Health Check
```bash
curl -X GET http://localhost:8080/health | jq .
```

### 2. Get User Behavior Analytics
```bash
curl -X GET http://localhost:8080/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/behavior \
  -H "Content-Type: application/json" | jq .
```

### 3. Get Risk Assessment
```bash
curl -X GET http://localhost:8080/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/risk \
  -H "Content-Type: application/json" | jq .
```

### 4. Detect Anomalies
```bash
curl -X POST http://localhost:8080/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/detect-anomalies \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_period": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-31T23:59:59Z"
    },
    "anomaly_types": ["login_time", "location", "device"],
    "sensitivity": "medium"
  }' | jq .
```

##  Documentation Files

- **curl-examples.md** -  curl command examples
- **API Reference** - Detailed API specification
- **Schema Reference** - Request/response schemas
- **Analytics Guide** - Understanding analytics data
- **Integration Guide** - Integration with security systems

##  Testing

### Automated Testing
```bash
# Run user analytics API tests
./utilities/scripts/test-user-analytics.sh

# Test specific analytics types
./utilities/scripts/test-user-analytics.sh --type behavior
./utilities/scripts/test-user-analytics.sh --type risk
./utilities/scripts/test-user-analytics.sh --type anomaly
```

### Manual Testing
```bash
# Follow step-by-step examples
cat curl-examples.md
```

##  Common Use Cases

### Security Monitoring
```bash
# Get  user behavior pattern
curl -X GET http://localhost:8080/api/v1/analytics/users/user-id/behavior | jq .

# Check current risk assessment
curl -X GET http://localhost:8080/api/v1/analytics/users/user-id/risk | jq .

# Detect recent anomalies
curl -X POST http://localhost:8080/api/v1/analytics/users/user-id/detect-anomalies \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_period": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-31T23:59:59Z"
    },
    "anomaly_types": ["login_time", "location", "device", "access_pattern"],
    "sensitivity": "high"
  }' | jq .
```

### Productivity Analysis
```bash
# Get personalized insights
curl -X GET http://localhost:8080/api/v1/analytics/users/user-id/insights | jq .

# Focus on productivity metrics
curl -X GET http://localhost:8080/api/v1/analytics/users/user-id/behavior | \
  jq '.behavior_pattern.productivity_metrics'
```

### Compliance Monitoring
```bash
# Get security profile
curl -X GET http://localhost:8080/api/v1/analytics/users/user-id/behavior | \
  jq '.behavior_pattern.security_profile'

# Check compliance score
curl -X GET http://localhost:8080/api/v1/analytics/users/user-id/risk | \
  jq '.risk_assessment.compliance_indicators'
```

##  Expected Responses

### Behavior Pattern Response (200 OK)
```json
{
  "behavior_pattern": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440001",
    "analysis_period": {
      "start": "2024-06-01T00:00:00Z",
      "end": "2024-07-01T00:00:00Z"
    },
    "last_updated": "2024-07-01T10:00:00Z",
    "login_patterns": {
      "typical_login_hours": [9, 10, 13, 14],
      "typical_login_days": [1, 2, 3, 4, 5],
      "average_session_duration": 14400000000000,
      "login_frequency": {
        "daily_average": 2.5,
        "weekly_average": 12.5,
        "monthly_average": 50.0,
        "consistency_score": 0.85
      },
      "location_consistency": {
        "primary_locations": [
          {
            "country": "US",
            "region": "CA",
            "city": "San Francisco",
            "frequency": 0.85,
            "is_trusted": true
          }
        ],
        "location_variance": 0.15,
        "trusted_location_ratio": 0.95
      },
      "device_consistency": {
        "primary_devices": [
          {
            "device_type": "laptop",
            "os": "Windows 10",
            "browser": "Chrome",
            "frequency": 0.80,
            "is_trusted": true
          }
        ],
        "device_variance": 0.20,
        "trusted_device_ratio": 0.90
      }
    },
    "activity_patterns": {
      "peak_activity_hours": [10, 11, 14, 15],
      "activity_distribution": {
        "administration": 0.2,
        "data_entry": 0.3,
        "document_access": 0.4,
        "reporting": 0.1
      },
      "productivity_metrics": {
        "task_completion_rate": 0.85,
        "average_task_duration": 2700000000000,
        "efficiency_score": 0.78
      }
    },
    "security_profile": {
      "password_change_frequency": 0.25,
      "mfa_usage_consistency": 0.95,
      "compliance_score": 0.88,
      "security_training_score": 0.92
    },
    "baseline_established": true,
    "baseline_confidence": 0.85
  }
}
```

### Risk Assessment Response (200 OK)
```json
{
  "risk_assessment": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "overall_risk_score": 25,
    "risk_level": "LOW",
    "risk_categories": {
      "access_patterns": 15,
      "login_behavior": 20,
      "security_behavior": 30,
      "compliance_behavior": 25
    },
    "risk_factors": [
      {
        "factor": "unusual_login_time",
        "severity": "medium",
        "score": 15,
        "description": "Login detected outside normal hours",
        "detected_at": "2024-01-01T02:30:00Z"
      }
    ],
    "behavioral_anomalies": [
      {
        "type": "login_time",
        "severity": "low",
        "description": "Login 2 hours earlier than usual",
        "confidence": 0.75
      }
    ],
    "predictive_risk_factors": [
      {
        "factor": "project_deadline_stress",
        "probability": 0.65,
        "impact": "medium",
        "timeframe": "next_week"
      }
    ],
    "recommendations": [
      {
        "priority": "medium",
        "action": "Enable MFA for all accounts",
        "reason": "Additional security layer recommended",
        "expected_impact": "Reduce risk score by 10 points"
      }
    ],
    "compliance_indicators": {
      "gdpr_compliance": 0.95,
      "sox_compliance": 0.88,
      "iso27001_compliance": 0.92
    },
    "last_assessment": "2024-01-01T10:00:00Z",
    "next_assessment": "2024-01-02T10:00:00Z"
  }
}
```

### Anomaly Detection Response (200 OK)
```json
{
  "anomaly_detection": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "analysis_period": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-31T23:59:59Z"
    },
    "anomalies_detected": [
      {
        "type": "unusual_login_time",
        "severity": "medium",
        "description": "Login detected at 2:30 AM, outside normal hours (9 AM - 6 PM)",
        "detected_at": "2024-01-15T02:30:00Z",
        "deviation_score": 0.85,
        "confidence": 0.78,
        "baseline_pattern": "Normal login hours: 9 AM - 6 PM",
        "context": {
          "ip_address": "192.168.1.100",
          "location": "San Francisco, CA",
          "device": "laptop"
        },
        "recommendations": [
          "Verify if this was authorized access",
          "Consider implementing time-based access controls",
          "Enable additional monitoring for unusual hours"
        ]
      },
      {
        "type": "new_device_access",
        "severity": "high",
        "description": "Access from previously unseen device",
        "detected_at": "2024-01-20T14:30:00Z",
        "deviation_score": 0.92,
        "confidence": 0.95,
        "baseline_pattern": "Typically uses Windows laptop",
        "context": {
          "device_type": "mobile",
          "os": "iOS",
          "browser": "Safari",
          "location": "New York, NY"
        },
        "recommendations": [
          "Verify device ownership",
          "Consider device registration requirements",
          "Enable device-based conditional access"
        ]
      }
    ],
    "anomaly_summary": {
      "total_anomalies": 2,
      "high_severity": 1,
      "medium_severity": 1,
      "low_severity": 0,
      "false_positive_rate": 0.05
    },
    "baseline_confidence": 0.82,
    "model_version": "v2.1.0",
    "next_analysis": "2024-02-01T00:00:00Z"
  }
}
```

### Personalized Insights Response (200 OK)
```json
{
  "insights": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "generated_at": "2024-01-01T10:00:00Z",
    "productivity_insights": [
      {
        "category": "time_management",
        "insight": "You're most productive between 10 AM and 2 PM",
        "recommendation": "Schedule important tasks during peak hours",
        "impact": "medium",
        "confidence": 0.85,
        "data_source": "activity_patterns"
      },
      {
        "category": "workflow_optimization",
        "insight": "Document access patterns suggest workflow optimization opportunities",
        "recommendation": "Consider organizing frequently accessed documents",
        "impact": "medium",
        "confidence": 0.78,
        "data_source": "access_patterns"
      }
    ],
    "security_insights": [
      {
        "category": "authentication",
        "insight": "MFA usage is inconsistent across applications",
        "recommendation": "Enable MFA for all critical applications",
        "impact": "high",
        "confidence": 0.90,
        "data_source": "security_profile"
      },
      {
        "category": "access_patterns",
        "insight": "Elevated access requests spike before project deadlines",
        "recommendation": "Plan access requests in advance",
        "impact": "medium",
        "confidence": 0.82,
        "data_source": "access_request_patterns"
      }
    ],
    "efficiency_insights": [
      {
        "category": "application_usage",
        "insight": "High application switching frequency detected",
        "recommendation": "Consider workflow consolidation tools",
        "impact": "low",
        "confidence": 0.70,
        "data_source": "application_usage"
      }
    ],
    "collaboration_insights": [
      {
        "category": "team_interaction",
        "insight": "Higher collaboration during morning hours",
        "recommendation": "Schedule team meetings in the morning",
        "impact": "low",
        "confidence": 0.75,
        "data_source": "collaboration_patterns"
      }
    ],
    "insights_summary": {
      "total_insights": 6,
      "high_impact": 1,
      "medium_impact": 3,
      "low_impact": 2,
      "actionable_recommendations": 5
    }
  }
}
```

##  Common Issues

1. **Insufficient data** - Need minimum activity for accurate analysis
2. **Baseline not established** - Requires learning period for patterns
3. **Privacy concerns** - Ensure compliance with privacy regulations
4. **Data quality issues** - Incomplete or inaccurate behavioral data
5. **False positives** - Anomaly detection may flag normal behavior

##  Performance Notes

- Analytics are computed asynchronously for better performance
- Cached results are refreshed based on data freshness requirements
- Machine learning models are updated periodically
- Large datasets may require paginated responses

##  Security and Privacy

- **Data Anonymization**: Personal data is anonymized for analysis
- **Encryption**: All analytics data is encrypted at rest
- **Access Controls**: Strict access controls for analytics data
- **Compliance**: GDPR, CCPA, and other privacy regulation compliance
- **Audit Trail**: All analytics access is logged and audited

##  Related APIs

- **User Management**: User data source for analytics
- **Access Request Workflow**: Behavioral data influences approval
- **Conditional Access**: Risk scores affect access decisions
- **Entity Management**: Analytics are scoped to entities

##  Analytics Configuration

### Data Collection
- **Login Events**: Track login times, locations, devices
- **Activity Events**: Monitor application usage, document access
- **Security Events**: Password changes, MFA usage, violations
- **Workflow Events**: Access requests, approvals, task completion

### Privacy Controls
- **Data Retention**: Configurable retention periods
- **Anonymization**: Automatic PII anonymization
- **Consent Management**: User consent tracking
- **Data Minimization**: Collect only necessary data

##  Monitoring and Metrics

### Analytics Quality Metrics
- **Baseline Confidence**: Accuracy of behavioral baselines
- **Anomaly Detection Accuracy**: True positive/false positive rates
- **Risk Assessment Precision**: Risk prediction accuracy
- **Insight Relevance**: User feedback on insights

### System Performance
- **Analytics Generation Time**: Time to compute analytics
- **Data Processing Latency**: Real-time vs batch processing
- **Storage Utilization**: Analytics data storage usage
- **API Response Times**: Analytics API performance

##  Advanced Features

### Machine Learning Models
- **Behavioral Clustering**: Group users by behavior patterns
- **Predictive Analytics**: Forecast future behavior and risks
- **Recommendation Engine**: Personalized security and productivity recommendations
- **Adaptive Thresholds**: Dynamic anomaly detection thresholds

### Integration Capabilities
- **SIEM Integration**: Export analytics to security information systems
- **Business Intelligence**: Connect with BI tools for reporting
- **Workflow Automation**: Trigger automated responses based on analytics
- **Third-party APIs**: Integrate with external analytics platforms