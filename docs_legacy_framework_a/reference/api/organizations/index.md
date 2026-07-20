> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Organizations API

## Overview

The Organizations API provides  organizational structure management for AWO ERP, supporting hierarchical organizations, department management, and organizational analytics. This API enables complex organizational structures with support for cost centers, budget management, and organizational reporting.

## Key Features

- **Hierarchical Structure**: Multi-level organizational hierarchies with parent-child relationships
- **Department Management**: Create and manage departments, teams, and business units
- **Cost Center Integration**: Link organizations to financial cost centers and budgets
- **Archive Support**: Soft-delete organizations while preserving historical data
- **Organizational Analytics**: Insights into organizational structure and performance
- **Bulk Operations**: Efficient bulk organizational operations

## Endpoints Overview

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/v1/organizations` | List organizations with hierarchy support |
| POST | `/api/v1/organizations` | Create new organization/department |
| GET | `/api/v1/organizations/{id}` | Get organization details |
| PUT | `/api/v1/organizations/{id}` | Update organization information |
| POST | `/api/v1/organizations/{id}/archive` | Archive organization |
| GET | `/api/v1/organizations/{id}/hierarchy` | Get organizational hierarchy |

## Organization Management

### List Organizations
```bash
curl -X GET "http://localhost:8080/api/v1/organizations" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "include_hierarchy=true" \
  -d "status=active" \
  -d "type=department"
```

**Response:**
```json
{
  "data": [
    {
      "id": "org-123",
      "name": "Engineering Department", 
      "type": "department",
      "status": "active",
      "parent_id": "org-root",
      "level": 2,
      "path": "/company/engineering",
      "settings": {
        "budget_code": "ENG-001",
        "cost_center": "CC-100", 
        "head_count": 50,
        "budget_limit": 1000000
      },
      "metadata": {
        "manager_id": "user-456",
        "description": "Software engineering and development",
        "location": "Building A, Floor 3"
      },
      "created_at": "2025-01-01T10:00:00Z",
      "updated_at": "2025-01-15T14:30:00Z",
      "children_count": 3
    }
  ],
  "hierarchy": {
    "total_levels": 4,
    "max_depth": 3,
    "total_nodes": 25
  }
}
```

### Create Organization
```bash
curl -X POST "http://localhost:8080/api/v1/organizations" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Platform Engineering Team",
    "type": "team",
    "parent_id": "org-123",
    "settings": {
      "budget_code": "ENG-PLT-001",
      "cost_center": "CC-101",
      "head_count": 12,
      "budget_limit": 200000
    },
    "metadata": {
      "manager_id": "user-789",
      "description": "Platform infrastructure and DevOps",
      "location": "Building A, Floor 3",
      "tags": ["platform", "infrastructure", "devops"]
    }
  }'
```

### Update Organization
```bash
curl -X PUT "http://localhost:8080/api/v1/organizations/org-123" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Software Engineering Department",
    "settings": {
      "budget_code": "ENG-001",
      "cost_center": "CC-100",
      "head_count": 55,
      "budget_limit": 1200000
    },
    "metadata": {
      "description": "Software engineering, platform, and mobile development",
      "location": "Building A, Floors 3-4"
    }
  }'
```

### Get Organization Hierarchy
```bash
curl -X GET "http://localhost:8080/api/v1/organizations/org-123/hierarchy" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "depth=3" \
  -d "include_inactive=false"
```

**Response:**
```json
{
  "organization": {
    "id": "org-123",
    "name": "Engineering Department",
    "type": "department",
    "level": 2
  },
  "hierarchy": {
    "children": [
      {
        "id": "org-124",
        "name": "Platform Engineering Team", 
        "type": "team",
        "level": 3,
        "settings": {
          "head_count": 12,
          "budget_limit": 200000
        },
        "children": []
      },
      {
        "id": "org-125",
        "name": "Mobile Engineering Team",
        "type": "team", 
        "level": 3,
        "settings": {
          "head_count": 15,
          "budget_limit": 250000
        },
        "children": []
      }
    ]
  },
  "statistics": {
    "total_descendants": 2,
    "total_employees": 27,
    "total_budget": 450000,
    "active_teams": 2
  }
}
```

## Organization Types and Structure

### Supported Organization Types
- **company**: Root-level company/enterprise
- **division**: Major business divisions
- **department**: Functional departments 
- **team**: Working teams within departments
- **project**: Project-based organizational units
- **cost_center**: Financial cost centers
- **location**: Geographic locations/offices

### Hierarchical Relationships
```bash
# Get full organizational chart
curl -X GET "http://localhost:8080/api/v1/organizations" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "view=tree" \
  -d "max_depth=5" \
  -d "include_stats=true"
```

### Organization Path Management
Organizations have path-based identifiers for easy navigation:
- `/company` (Root organization)
- `/company/engineering` (Engineering department) 
- `/company/engineering/platform` (Platform team)
- `/company/finance/accounting` (Accounting team)

## Financial Integration

### Budget and Cost Center Management
```bash
curl -X PUT "http://localhost:8080/api/v1/organizations/org-123" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "settings": {
      "budget_code": "ENG-001",
      "cost_center": "CC-100",
      "fiscal_year": "2025",
      "budget_limit": 1500000,
      "budget_spent": 800000,
      "budget_allocated": {
        "salaries": 800000,
        "equipment": 200000, 
        "training": 100000,
        "travel": 50000,
        "other": 350000
      }
    }
  }'
```

### Organization Financial Reporting
```bash
curl -X GET "http://localhost:8080/api/v1/organizations/org-123/financial-summary" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "period=2025-Q1" \
  -d "include_children=true"
```

## Organization Members and Roles

### Get Organization Members
```bash
curl -X GET "http://localhost:8080/api/v1/organizations/org-123/members" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "include_roles=true" \
  -d "status=active"
```

### Assign Users to Organizations
```bash
curl -X POST "http://localhost:8080/api/v1/organizations/org-123/members" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-789",
    "role": "team_lead",
    "start_date": "2025-02-01",
    "permissions": ["manage_team", "approve_budget", "hire"]
  }'
```

## Archive and Restoration

### Archive Organization
```bash
curl -X POST "http://localhost:8080/api/v1/organizations/org-123/archive" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "archive_reason": "Department restructuring",
    "effective_date": "2025-03-01",
    "reassign_members": true,
    "target_organization_id": "org-456"
  }'
```

**Response:**
```json
{
  "archived": true,
  "archived_at": "2025-02-01T15:30:00Z",
  "archive_id": "arch-123",
  "affected_users": 15,
  "reassignments": {
    "completed": 15,
    "failed": 0
  },
  "financial_impact": {
    "budget_transferred": 500000,
    "target_organization": "org-456"
  }
}
```

### Restore Archived Organization
```bash
curl -X POST "http://localhost:8080/api/v1/organizations/org-123/restore" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "restore_members": true,
    "restore_budget": true
  }'
```

## Analytics and Reporting

### Organization Analytics
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/organizations/org-123" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "period=30d" \
  -d "include_productivity=true"
```

**Response:**
```json
{
  "organization_id": "org-123",
  "period": "30d", 
  "summary": {
    "total_employees": 55,
    "budget_utilization": 0.67,
    "productivity_score": 85,
    "employee_satisfaction": 4.2,
    "retention_rate": 0.95
  },
  "trends": {
    "headcount_change": "+5",
    "budget_trend": "under_budget",
    "productivity_trend": "increasing",
    "satisfaction_trend": "stable"
  },
  "breakdown": {
    "by_role": {
      "senior_engineer": 15,
      "engineer": 25,
      "manager": 5,
      "intern": 10
    },
    "by_team": {
      "platform": 12,
      "mobile": 15,
      "web": 18,
      "infrastructure": 10
    }
  }
}
```

### Cross-Organization Reporting
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/organizations/compare" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "org_ids=org-123,org-124,org-125" \
  -d "metrics=headcount,budget,productivity" \
  -d "period=90d"
```

## Advanced Operations

### Bulk Organization Updates
```bash
curl -X POST "http://localhost:8080/api/v1/organizations/bulk-update" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filter": {
      "type": "team",
      "parent_id": "org-123"
    },
    "updates": {
      "settings": {
        "fiscal_year": "2025",
        "budget_review_date": "2025-03-01"
      }
    }
  }'
```

### Organization Restructuring
```bash
curl -X POST "http://localhost:8080/api/v1/organizations/restructure" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {
        "type": "move",
        "organization_id": "org-124", 
        "new_parent_id": "org-456"
      },
      {
        "type": "merge",
        "source_ids": ["org-125", "org-126"],
        "target_id": "org-127",
        "preserve_history": true
      }
    ],
    "effective_date": "2025-04-01",
    "notify_affected_users": true
  }'
```

## Error Handling

### Common Organization API Errors

#### Organization Not Found
```json
{
  "error": {
    "code": "ORGANIZATION_NOT_FOUND", 
    "message": "Organization with ID 'org-123' not found",
    "details": {"organization_id": "org-123"}
  }
}
```

#### Circular Hierarchy
```json
{
  "error": {
    "code": "CIRCULAR_HIERARCHY",
    "message": "Cannot set parent: would create circular reference",
    "details": {
      "organization_id": "org-123",
      "attempted_parent_id": "org-124",
      "circular_path": ["org-123", "org-124", "org-125", "org-123"]
    }
  }
}
```

#### Budget Limit Exceeded
```json
{
  "error": {
    "code": "BUDGET_LIMIT_EXCEEDED",
    "message": "Budget allocation exceeds parent organization limit",
    "details": {
      "requested_budget": 1500000,
      "available_budget": 1200000,
      "parent_organization": "org-root"
    }
  }
}
```

#### Active Members Exist
```json
{
  "error": {
    "code": "ACTIVE_MEMBERS_EXIST",
    "message": "Cannot archive organization with active members",
    "details": {
      "active_member_count": 15,
      "suggestion": "Reassign members before archiving"
    }
  }
}
```

## Best Practices

### Organizational Design
- **Logical Hierarchy**: Design hierarchy to match business structure
- **Balanced Depth**: Avoid excessive hierarchy levels (max 5-6 levels)
- **Clear Naming**: Use consistent, descriptive naming conventions
- **Budget Alignment**: Align organizational structure with financial reporting needs

### Performance Optimization
- **Lazy Loading**: Use pagination and filters for large hierarchies
- **Caching**: Cache frequently accessed organizational data
- **Bulk Operations**: Use bulk endpoints for efficiency
- **Selective Inclusion**: Only include necessary data in responses

### Data Integrity
- **Validation**: Validate hierarchy changes before committing
- **Audit Trail**: Track all organizational changes for compliance
- **Backup Planning**: Plan for organizational data recovery
- **Member Management**: Handle member reassignments during restructuring

## Integration Examples

### Organization Sync with HR System
```python
def sync_organizations_from_hr(api_client, hr_org_data):
    """Sync organizational structure from HR system."""
    results = {'created': 0, 'updated': 0, 'errors': []}
    
    # Sort by hierarchy level to ensure parents exist first
    hr_orgs = sorted(hr_org_data, key=lambda x: x.get('level', 0))
    
    for hr_org in hr_orgs:
        try:
            org_data = {
                'name': hr_org['department_name'],
                'type': hr_org['type'],
                'parent_id': hr_org.get('parent_department_id'),
                'settings': {
                    'budget_code': hr_org['cost_center'],
                    'cost_center': hr_org['cost_center'],
                    'head_count': hr_org['employee_count']
                },
                'metadata': {
                    'manager_id': hr_org.get('manager_employee_id'),
                    'description': hr_org.get('description'),
                    'location': hr_org.get('office_location')
                }
            }
            
            # Check if organization exists
            existing_org = api_client.find_organization_by_hr_id(hr_org['hr_dept_id'])
            
            if existing_org:
                api_client.update_organization(existing_org['id'], org_data)
                results['updated'] += 1
            else:
                api_client.create_organization(org_data)
                results['created'] += 1
                
        except Exception as e:
            results['errors'].append({
                'hr_org': hr_org['department_name'],
                'error': str(e)
            })
    
    return results
```

---

**Next**: [Tenants API](../../api-source/core-apis/tenants/README.md) | **Up**: [API Reference](../index.md)