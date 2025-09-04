# API Integration Tutorial

## Complete Guide to AWO ERP API Integration

This tutorial walks through integrating with the AWO ERP API from initial setup to production deployment, with practical examples in multiple programming languages.

## Prerequisites

Before starting, ensure you have:
- Access credentials (username, password, tenant ID)  
- Development environment set up
- API base URL (e.g., `http://localhost:8080` for local development)
- Basic understanding of REST APIs and JWT authentication

## Step 1: Authentication & Token Management

### Initial Authentication
```bash
# Login and get JWT token
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "developer@company.com",
    "password": "your-secure-password",
    "tenant_id": "your-tenant-uuid"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer",
  "user": {
    "id": "user-uuid",
    "username": "developer@company.com",
    "roles": ["developer", "api_user"],
    "permissions": ["read:users", "write:entities"]
  }
}
```

### Token Refresh Implementation

**JavaScript/Node.js Example:**
```javascript
class AWOERPClient {
  constructor(baseUrl, credentials) {
    this.baseUrl = baseUrl;
    this.credentials = credentials;
    this.accessToken = null;
    this.refreshToken = null;
    this.tokenExpiry = null;
  }

  async authenticate() {
    const response = await fetch(`${this.baseUrl}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(this.credentials)
    });

    const data = await response.json();
    if (response.ok) {
      this.accessToken = data.access_token;
      this.refreshToken = data.refresh_token;
      this.tokenExpiry = Date.now() + (data.expires_in * 1000);
      return data;
    }
    throw new Error(`Authentication failed: ${data.error?.message}`);
  }

  async refreshAccessToken() {
    const response = await fetch(`${this.baseUrl}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: this.refreshToken })
    });

    const data = await response.json();
    if (response.ok) {
      this.accessToken = data.access_token;
      this.tokenExpiry = Date.now() + (data.expires_in * 1000);
      return data;
    }
    throw new Error(`Token refresh failed: ${data.error?.message}`);
  }

  async makeRequest(endpoint, options = {}) {
    // Check if token needs refresh
    if (Date.now() >= this.tokenExpiry - 60000) { // 1 minute buffer
      await this.refreshAccessToken();
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.accessToken}`,
        'Content-Type': 'application/json',
        ...options.headers
      }
    });

    return response.json();
  }
}

// Usage
const client = new AWOERPClient('http://localhost:8080', {
  username: 'developer@company.com',
  password: 'your-password',
  tenant_id: 'tenant-uuid'
});

await client.authenticate();
```

**Python Example:**
```python
import requests
import time
from datetime import datetime, timedelta

class AWOERPClient:
    def __init__(self, base_url, credentials):
        self.base_url = base_url
        self.credentials = credentials
        self.access_token = None
        self.refresh_token = None
        self.token_expiry = None
        self.session = requests.Session()
    
    def authenticate(self):
        response = self.session.post(
            f"{self.base_url}/api/v1/auth/login",
            json=self.credentials
        )
        response.raise_for_status()
        
        data = response.json()
        self.access_token = data['access_token']
        self.refresh_token = data['refresh_token']
        self.token_expiry = datetime.now() + timedelta(seconds=data['expires_in'])
        
        # Set default authorization header
        self.session.headers.update({
            'Authorization': f"Bearer {self.access_token}"
        })
        
        return data
    
    def refresh_access_token(self):
        response = self.session.post(
            f"{self.base_url}/api/v1/auth/refresh",
            json={'refresh_token': self.refresh_token}
        )
        response.raise_for_status()
        
        data = response.json()
        self.access_token = data['access_token'] 
        self.token_expiry = datetime.now() + timedelta(seconds=data['expires_in'])
        
        # Update session header
        self.session.headers.update({
            'Authorization': f"Bearer {self.access_token}"
        })
        
        return data
    
    def make_request(self, method, endpoint, **kwargs):
        # Check token expiry
        if datetime.now() >= self.token_expiry - timedelta(minutes=1):
            self.refresh_access_token()
        
        response = self.session.request(method, f"{self.base_url}{endpoint}", **kwargs)
        response.raise_for_status()
        return response.json()
    
    def get(self, endpoint, **kwargs):
        return self.make_request('GET', endpoint, **kwargs)
    
    def post(self, endpoint, **kwargs):
        return self.make_request('POST', endpoint, **kwargs)
    
    def put(self, endpoint, **kwargs):
        return self.make_request('PUT', endpoint, **kwargs)
    
    def delete(self, endpoint, **kwargs):
        return self.make_request('DELETE', endpoint, **kwargs)

# Usage
client = AWOERPClient('http://localhost:8080', {
    'username': 'developer@company.com',
    'password': 'your-password',
    'tenant_id': 'tenant-uuid'
})

client.authenticate()
```

## Step 2: Basic API Operations

### User Management
```python
# Get current user profile
profile = client.get('/api/v1/users/profile')
print(f"User: {profile['data']['username']}")

# List users (with pagination)
users = client.get('/api/v1/users', params={
    'page': 1,
    'limit': 20,
    'status': 'active'
})

for user in users['data']:
    print(f"- {user['username']} ({user['email']})")

# Create new user
new_user = client.post('/api/v1/users', json={
    'username': 'newuser@company.com',
    'email': 'newuser@company.com',
    'first_name': 'New',
    'last_name': 'User',
    'roles': ['user']
})
```

### Organization Management  
```python
# Get organization hierarchy
orgs = client.get('/api/v1/organizations')

# Create department
department = client.post('/api/v1/organizations', json={
    'name': 'Engineering Department',
    'type': 'department',
    'parent_id': 'parent-org-uuid',
    'settings': {
        'budget_code': 'ENG-001',
        'cost_center': 'CC-100'
    }
})

# Update organization
client.put(f"/api/v1/organizations/{department['data']['id']}", json={
    'name': 'Software Engineering Department',
    'settings': {
        'budget_code': 'ENG-001',
        'cost_center': 'CC-100',
        'head_count': 50
    }
})
```

### Access Request Workflow
```python
# Submit access request
request = client.post('/api/v1/access-requests', json={
    'resource_type': 'entity',
    'resource_id': 'sensitive-project-uuid',
    'access_level': 'write',
    'justification': 'Need write access for project delivery milestone',
    'duration': 'temporary',
    'expires_at': '2025-12-31T23:59:59Z'
})

request_id = request['data']['id']
print(f"Access request submitted: {request_id}")

# Check request status
status = client.get(f'/api/v1/access-requests/{request_id}')
print(f"Status: {status['data']['status']}")

# List pending requests (if user is an approver)
pending = client.get('/api/v1/access-requests', params={
    'status': 'pending',
    'approver_id': 'my-user-id'
})
```

## Step 3: Advanced Authorization (ABAC)

### Simple Authorization Check
```python
def check_permission(client, user_id, resource_id, resource_type, action):
    result = client.post('/abac/authorize', json={
        'user_id': user_id,
        'resource_id': resource_id,
        'resource_type': resource_type,
        'action': action
    })
    return result['allowed']

# Usage
can_edit = check_permission(client, 'user-123', 'document-456', 'document', 'edit')
if can_edit:
    print("User can edit document")
else:
    print("Access denied")
```

###  Authorization with Context
```python
def check_with_context(client, user_id, resource_id, resource_type, action, context=None):
    payload = {
        'user_id': user_id,
        'resource_id': resource_id,
        'resource_type': resource_type,
        'action': action,
        'explain_decision': True,
        'include_advice': True
    }
    
    if context:
        payload['context'] = context
    
    result = client.post('/abac/evaluate', json=payload)
    
    return {
        'allowed': result['allowed'],
        'decision': result['decision'],
        'reasoning': result.get('explanation', {}).get('reasoning_summary'),
        'advice': result.get('advice', []),
        'evaluation_time': result.get('evaluation_time_ms', 0)
    }

# Usage with environmental context
context = {
    'ip_address': '192.168.1.100',
    'user_agent': 'MyApp/1.0',
    'time_of_day': 'business_hours',
    'location': 'corporate_office'
}

auth_result = check_with_context(
    client, 'user-123', 'financial-report-789', 
    'financial_report', 'export', context
)

print(f"Allowed: {auth_result['allowed']}")
print(f"Reasoning: {auth_result['reasoning']}")
```

### Bulk Authorization for Performance
```python
def check_bulk_permissions(client, user_id, resources):
    requests = []
    for resource in resources:
        requests.append({
            'resource_id': resource['id'],
            'resource_type': resource['type'], 
            'action': resource['action']
        })
    
    result = client.post('/abac/evaluate-bulk', json={
        'user_id': user_id,
        'requests': requests,
        'use_cache': True
    })
    
    # Return dictionary mapping resource_id to permission
    permissions = {}
    for i, response in enumerate(result['responses']):
        resource_id = resources[i]['id']
        permissions[resource_id] = response['allowed']
    
    return permissions

# Usage
resources = [
    {'id': 'doc-1', 'type': 'document', 'action': 'read'},
    {'id': 'doc-2', 'type': 'document', 'action': 'write'},
    {'id': 'report-3', 'type': 'report', 'action': 'export'}
]

permissions = check_bulk_permissions(client, 'user-123', resources)
for resource_id, allowed in permissions.items():
    print(f"{resource_id}: {'✓' if allowed else '✗'}")
```

## Step 4: Error Handling & Resilience

### Robust Error Handling
```python
import logging
from requests.exceptions import RequestException, Timeout, ConnectionError

class APIError(Exception):
    def __init__(self, message, status_code=None, error_code=None, details=None):
        super().__init__(message)
        self.status_code = status_code
        self.error_code = error_code
        self.details = details

class AWOERPClientWithRetry(AWOERPClient):
    def __init__(self, base_url, credentials, max_retries=3, backoff_factor=1.0):
        super().__init__(base_url, credentials)
        self.max_retries = max_retries
        self.backoff_factor = backoff_factor
        self.logger = logging.getLogger(__name__)
    
    def make_request_with_retry(self, method, endpoint, **kwargs):
        last_exception = None
        
        for attempt in range(self.max_retries + 1):
            try:
                # Check token expiry
                if datetime.now() >= self.token_expiry - timedelta(minutes=1):
                    self.refresh_access_token()
                
                response = self.session.request(method, f"{self.base_url}{endpoint}", **kwargs)
                
                # Handle specific HTTP status codes
                if response.status_code == 401:
                    # Token might be expired, try refresh once
                    if attempt == 0:
                        self.refresh_access_token()
                        continue
                    raise APIError("Authentication failed", 401)
                
                elif response.status_code == 403:
                    error_data = response.json()
                    raise APIError(
                        error_data.get('error', {}).get('message', 'Access denied'),
                        403,
                        error_data.get('error', {}).get('code')
                    )
                
                elif response.status_code == 429:
                    # Rate limited - wait and retry
                    retry_after = int(response.headers.get('Retry-After', 60))
                    self.logger.warning(f"Rate limited, waiting {retry_after}s")
                    time.sleep(retry_after)
                    continue
                
                elif response.status_code >= 500:
                    # Server error - retry with backoff
                    if attempt < self.max_retries:
                        wait_time = self.backoff_factor * (2 ** attempt)
                        self.logger.warning(f"Server error {response.status_code}, retrying in {wait_time}s")
                        time.sleep(wait_time)
                        continue
                
                response.raise_for_status()
                return response.json()
                
            except (ConnectionError, Timeout) as e:
                last_exception = e
                if attempt < self.max_retries:
                    wait_time = self.backoff_factor * (2 ** attempt)
                    self.logger.warning(f"Connection error, retrying in {wait_time}s")
                    time.sleep(wait_time)
                    continue
                
            except RequestException as e:
                raise APIError(f"Request failed: {str(e)}")
        
        # All retries exhausted
        raise APIError(f"Request failed after {self.max_retries} retries") from last_exception
```

### Circuit Breaker Pattern
```python
from enum import Enum
import time

class CircuitState(Enum):
    CLOSED = "closed"
    OPEN = "open" 
    HALF_OPEN = "half_open"

class CircuitBreaker:
    def __init__(self, failure_threshold=5, timeout=60, expected_exception=APIError):
        self.failure_threshold = failure_threshold
        self.timeout = timeout
        self.expected_exception = expected_exception
        self.failure_count = 0
        self.last_failure_time = None
        self.state = CircuitState.CLOSED
    
    def call(self, func, *args, **kwargs):
        if self.state == CircuitState.OPEN:
            if time.time() - self.last_failure_time >= self.timeout:
                self.state = CircuitState.HALF_OPEN
            else:
                raise APIError("Circuit breaker is OPEN")
        
        try:
            result = func(*args, **kwargs)
            self.on_success()
            return result
        except self.expected_exception as e:
            self.on_failure()
            raise
    
    def on_success(self):
        self.failure_count = 0
        self.state = CircuitState.CLOSED
    
    def on_failure(self):
        self.failure_count += 1
        self.last_failure_time = time.time()
        
        if self.failure_count >= self.failure_threshold:
            self.state = CircuitState.OPEN

# Usage
circuit_breaker = CircuitBreaker()
robust_client = AWOERPClientWithRetry('http://localhost:8080', credentials)

def safe_api_call(endpoint, **kwargs):
    return circuit_breaker.call(robust_client.get, endpoint, **kwargs)
```

## Step 5: Production Considerations

### Configuration Management
```python
import os
from dataclasses import dataclass
from typing import Optional

@dataclass
class AWOERPConfig:
    base_url: str
    username: str
    password: str
    tenant_id: str
    timeout: int = 30
    max_retries: int = 3
    backoff_factor: float = 1.0
    rate_limit_per_hour: int = 1000
    enable_metrics: bool = False
    metrics_endpoint: Optional[str] = None
    
    @classmethod
    def from_env(cls):
        return cls(
            base_url=os.getenv('AWO_ERP_BASE_URL', 'http://localhost:8080'),
            username=os.getenv('AWO_ERP_USERNAME'),
            password=os.getenv('AWO_ERP_PASSWORD'), 
            tenant_id=os.getenv('AWO_ERP_TENANT_ID'),
            timeout=int(os.getenv('AWO_ERP_TIMEOUT', 30)),
            max_retries=int(os.getenv('AWO_ERP_MAX_RETRIES', 3)),
            backoff_factor=float(os.getenv('AWO_ERP_BACKOFF_FACTOR', 1.0)),
            rate_limit_per_hour=int(os.getenv('AWO_ERP_RATE_LIMIT', 1000))
        )
```

### Metrics & Monitoring
```python
import time
from collections import defaultdict, deque
import threading

class APIMetrics:
    def __init__(self):
        self.request_counts = defaultdict(int)
        self.response_times = defaultdict(deque)
        self.error_counts = defaultdict(int)
        self.lock = threading.Lock()
    
    def record_request(self, endpoint, response_time, status_code):
        with self.lock:
            self.request_counts[endpoint] += 1
            self.response_times[endpoint].append(response_time)
            
            # Keep only last 1000 response times
            if len(self.response_times[endpoint]) > 1000:
                self.response_times[endpoint].popleft()
            
            if status_code >= 400:
                self.error_counts[endpoint] += 1
    
    def get_stats(self):
        with self.lock:
            stats = {}
            for endpoint in self.request_counts:
                response_times = list(self.response_times[endpoint])
                stats[endpoint] = {
                    'total_requests': self.request_counts[endpoint],
                    'error_count': self.error_counts[endpoint],
                    'error_rate': self.error_counts[endpoint] / self.request_counts[endpoint] if self.request_counts[endpoint] > 0 else 0,
                    'avg_response_time': sum(response_times) / len(response_times) if response_times else 0,
                    'min_response_time': min(response_times) if response_times else 0,
                    'max_response_time': max(response_times) if response_times else 0
                }
            return stats

class ProductionAWOERPClient(AWOERPClientWithRetry):
    def __init__(self, config: AWOERPConfig):
        super().__init__(
            config.base_url,
            {
                'username': config.username,
                'password': config.password,
                'tenant_id': config.tenant_id
            },
            max_retries=config.max_retries,
            backoff_factor=config.backoff_factor
        )
        self.config = config
        self.metrics = APIMetrics() if config.enable_metrics else None
        self.session.timeout = config.timeout
    
    def make_request_with_retry(self, method, endpoint, **kwargs):
        start_time = time.time()
        status_code = 200
        
        try:
            result = super().make_request_with_retry(method, endpoint, **kwargs)
            return result
        except APIError as e:
            status_code = e.status_code or 500
            raise
        finally:
            if self.metrics:
                response_time = time.time() - start_time
                self.metrics.record_request(endpoint, response_time, status_code)
```

### Health Checks & Monitoring
```python
def health_check(client):
    """ health check for API connectivity."""
    health_status = {
        'api_reachable': False,
        'authentication_working': False,
        'authorization_working': False,
        'response_time_ms': None,
        'timestamp': datetime.now().isoformat()
    }
    
    try:
        start_time = time.time()
        
        # Test basic connectivity
        health_response = client.get('/health')
        health_status['api_reachable'] = True
        
        # Test authentication
        profile = client.get('/api/v1/users/profile')
        health_status['authentication_working'] = True
        
        # Test authorization
        auth_test = client.post('/abac/authorize', json={
            'user_id': profile['data']['id'],
            'resource_id': 'health-check-resource',
            'resource_type': 'system',
            'action': 'health_check'
        })
        health_status['authorization_working'] = True
        
        health_status['response_time_ms'] = (time.time() - start_time) * 1000
        
    except Exception as e:
        health_status['error'] = str(e)
    
    return health_status

# Periodic health monitoring
def monitor_api_health(client, interval_seconds=300):
    """Run periodic health checks."""
    while True:
        health = health_check(client)
        
        # Log health status
        if health.get('error'):
            logging.error(f"API health check failed: {health['error']}")
        else:
            logging.info(f"API healthy - response time: {health['response_time_ms']:.1f}ms")
        
        # Send to monitoring system
        if client.config.metrics_endpoint:
            try:
                requests.post(client.config.metrics_endpoint, json=health)
            except Exception as e:
                logging.warning(f"Failed to send health metrics: {e}")
        
        time.sleep(interval_seconds)

# Usage
config = AWOERPConfig.from_env()
client = ProductionAWOERPClient(config)
client.authenticate()

# Start background health monitoring
import threading
health_thread = threading.Thread(target=monitor_api_health, args=(client,), daemon=True)
health_thread.start()
```

## Next Steps

1. **Explore Interactive Documentation**: Use the [Swagger UI](../swagger-ui.md) to test endpoints
2. **Review Security Best Practices**: Read the [Security Guide](../../../contributing/securityHandbook.md)
3. **Implement Error Monitoring**: Set up logging and alerting for production use
4. **Optimize for Scale**: Implement connection pooling and async patterns for high-volume usage

## Additional Resources

- **[ABAC Integration Guide](../../../modules/user/abac/policy_evaluation.md)**: Deep dive into authorization patterns
- **[API Testing Scripts](../../api-source/utilities/scripts/)**: Ready-to-use testing utilities
- **[Postman Collections](../../api-source/utilities/postman/)**: Interactive API exploration
- **[Contributing Guide](../../../contributing/01-best-practices.md)**: Best practices for API integration