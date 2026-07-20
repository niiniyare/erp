> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Access Request Schema Reference

## Core Schemas

### AccessRequest
Complete schema for access request objects.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "AccessRequest",
  "required": ["resource_type", "resource_id", "access_level", "justification"],
  "properties": {
    "id": {
      "type": "string",
      "format": "uuid",
      "readOnly": true
    },
    "user_id": {
      "type": "string", 
      "format": "uuid",
      "readOnly": true
    },
    "resource_type": {
      "type": "string",
      "enum": ["entity", "role", "permission", "module"]
    },
    "resource_id": {
      "type": "string",
      "format": "uuid"
    },
    "access_level": {
      "type": "string", 
      "enum": ["read", "write", "admin", "execute"]
    },
    "status": {
      "type": "string",
      "enum": ["pending", "approved", "denied", "expired"],
      "readOnly": true
    },
    "justification": {
      "type": "string",
      "minLength": 10,
      "maxLength": 1000
    },
    "duration": {
      "type": "string",
      "enum": ["temporary", "permanent"],
      "default": "temporary"
    },
    "expires_at": {
      "type": "string",
      "format": "date-time"
    },
    "created_at": {
      "type": "string",
      "format": "date-time",
      "readOnly": true
    },
    "approved_at": {
      "type": "string",
      "format": "date-time",
      "readOnly": true
    },
    "approved_by": {
      "type": "string",
      "format": "uuid",
      "readOnly": true
    }
  }
}
```

### ApprovalAction
Schema for approval/denial actions.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object", 
  "title": "ApprovalAction",
  "properties": {
    "comment": {
      "type": "string",
      "maxLength": 500
    },
    "conditions": {
      "type": "object",
      "properties": {
        "time_restrictions": {
          "type": "array",
          "items": {
            "type": "string",
            "pattern": "^[0-2][0-9]:[0-5][0-9]-[0-2][0-9]:[0-5][0-9]$"
          }
        },
        "ip_whitelist": {
          "type": "array", 
          "items": {
            "type": "string",
            "format": "ipv4"
          }
        }
      }
    }
  }
}
```