> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Authentication API Reference

## Overview
Complete reference for authentication endpoints including login, token refresh, and session management.

## Endpoints

### POST /auth/login
Authenticate user and return JWT token.

**Request Body:**
```json
{
  "username": "string",
  "password": "string",
  "tenant_id": "string"
}
```

**Response:**
```json
{
  "access_token": "string",
  "refresh_token": "string", 
  "expires_in": 3600,
  "user": {
    "id": "string",
    "username": "string",
    "roles": ["string"]
  }
}
```
