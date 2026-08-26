# ABAC Policy Evaluation Engine - Decision Document

## 1. Reuse vs Build: Expression Evaluator / DSL

### Executive Summary
Recommend Google CEL (cel-go) for expression evaluation due to proven security, performance, and extensive Go ecosystem support.

### Recommended Option
**Reuse: Google CEL with github.com/google/cel-go**

### Rationale and Trade-offs

**Why CEL:**
- Non-Turing complete language designed for simplicity, speed, safety, and portability with C-like syntax
- Battle-tested in Google Cloud IAM, Kubernetes admission controllers
- Stable v1+ library with active maintenance
- Built-in sandboxing and resource limits
- Excellent debugging and type checking
- Extensible with custom functions

**Pros:**
- Zero security vulnerabilities from custom parser
- Rich standard library and type system
- Production-ready performance
- Strong Go integration
- Built-in cost estimation for DoS protection

**Cons:**
- External dependency
- Learning curve for policy authors
- Less control over syntax

**Alternatives rejected:**
- expr-lang: Smaller ecosystem, fewer security features
- OPA/Rego: Too heavyweight, different paradigm
- OpenFGA: Relationship-focused, not general ABAC
- Custom DSL: High security/maintenance risk

### Implementation Sketch

```go
package policy

import (
    "github.com/google/cel-go/cel"
    "github.com/google/cel-go/checker/decls"
)

type PolicyEngine struct {
    env      *cel.Env
    policies map[string]*CompiledPolicy
}

type CompiledPolicy struct {
    ID          string
    Version     string
    TenantID    string
    Expression  string
    Program     cel.Program
    CostEst     checker.CostEstimate
}

func NewPolicyEngine() (*PolicyEngine, error) {
    env, err := cel.NewEnv(
        cel.Declarations(
            decls.NewVar("subject", decls.NewMapType(decls.String, decls.Dyn)),
            decls.NewVar("resource", decls.NewMapType(decls.String, decls.Dyn)),
            decls.NewVar("environment", decls.NewMapType(decls.String, decls.Dyn)),
        ),
        cel.OptionalTypes(),
        // Add custom functions
        cel.Function("hasRole",
            cel.MemberOverload("hasRole_string", []*cel.Type{cel.StringType}, cel.BoolType,
                cel.FunctionBinding(func(args ...ref.Val) ref.Val {
                    // Implementation
                }))),
    )
    if err != nil {
        return nil, err
    }
    
    return &PolicyEngine{
        env:      env,
        policies: make(map[string]*CompiledPolicy),
    }, nil
}

func (pe *PolicyEngine) CompilePolicy(id, tenantID, expression string) error {
    ast, issues := pe.env.Compile(expression)
    if issues != nil && issues.Err() != nil {
        return issues.Err()
    }
    
    program, err := pe.env.Program(ast)
    if err != nil {
        return err
    }
    
    cost, err := pe.env.EstimateCost(ast)
    if err != nil {
        return err
    }
    
    pe.policies[fmt.Sprintf("%s:%s", tenantID, id)] = &CompiledPolicy{
        ID:         id,
        TenantID:   tenantID,
        Expression: expression,
        Program:    program,
        CostEst:    cost,
    }
    
    return nil
}
```

**Policy Lifecycle:**
1. **Author**: Web UI with CEL syntax validation
2. **Version**: Semantic versioning (v1.2.3) stored in DB
3. **Compile**: Pre-compile to CEL Program, validate cost limits
4. **Rollout**: Blue-green deployment with health checks
5. **Rollback**: Instant version switch, cached programs retained

**Example Policies:**
```cel
// Manager approval within limit
subject.role == "manager" && 
resource.amount <= subject.approval_limit &&
resource.department == subject.department

// Time-based access
environment.current_time.getHours() >= 9 && 
environment.current_time.getHours() <= 17

// Multi-factor authentication required
resource.sensitivity == "high" ? 
  has(subject.mfa_verified) && subject.mfa_verified : true
```

### Testing & Validation Plan

**Unit Tests:**
- CEL expression compilation success/failure
- Custom function registration and execution
- Cost estimation accuracy
- Type checking validation

**Integration Tests:**
- Policy compilation pipeline end-to-end
- Error handling for malformed policies
- Performance regression detection

**Example Test Cases:**
```go
func TestPolicyCompilation(t *testing.T) {
    tests := []struct {
        name        string
        expression  string
        expectError bool
        expectedMsg string
    }{
        {
            name:        "valid_manager_check",
            expression:  "subject.role == 'manager'",
            expectError: false,
        },
        {
            name:        "invalid_syntax",
            expression:  "subject.role ==",
            expectError: true,
            expectedMsg: "syntax error",
        },
        {
            name:        "type_mismatch",
            expression:  "subject.role + 123",
            expectError: true,
            expectedMsg: "type mismatch",
        },
    }
    
    engine, _ := NewPolicyEngine()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := engine.CompilePolicy("test", "tenant1", tt.expression)
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Estimated Effort
**4-6 person-days** (1 senior Go developer)
- Days 1-2: CEL integration and basic compilation
- Days 3-4: Policy lifecycle management
- Days 5-6: Testing and documentation

### Risks & Mitigations
- **Risk**: CEL learning curve for policy authors
  - **Mitigation**:  templates, interactive tutorials, validation feedback
- **Risk**: Performance overhead from CEL interpretation
  - **Mitigation**: Policy pre-compilation, cost estimation, benchmarking
- **Risk**: CEL library bugs or security issues
  - **Mitigation**: Regular updates, sandboxing, resource limits

## 2. Policy Semantics & Combining Algorithms

### Executive Summary
Implement Deny-overrides as default with support for Permit-overrides and First-applicable, using fail-closed semantics for indeterminate cases.

### Recommended Option
**Primary: Deny-overrides (default), Secondary: Permit-overrides, First-applicable**

### Rationale and Trade-offs

**Deny-overrides as default:**
- Most secure for ERP financial/sensitive data
- Explicit security principle: any deny wins
- Easier audit compliance reasoning
- Consistent with enterprise security best practices

**Additional algorithms:**
- **Permit-overrides**: For performance-critical, low-risk operations
- **First-applicable**: For simple ordered rule evaluation

**Indeterminate handling:**
- **Default**: Fail-closed (DENY) for security
- **Configurable**: Per-tenant override for specific resource types
- **Explicit**: Policies can use `has()` to check attribute existence

### Implementation Sketch

```go
type Decision int

const (
    Permit Decision = iota
    Deny
    NotApplicable
    Indeterminate
)

type PolicyResult struct {
    Decision    Decision
    Obligations []Obligation
    Advice      []Advice
    Applicable  bool
    RuleID      string
    Error       error
}

type Obligation struct {
    Type   string                 `json:"type"`
    Data   map[string]any `json:"data"`
    RuleID string                 `json:"rule_id"`
}

func CombineResults(results []PolicyResult, algorithm CombiningAlgorithm) PolicyResult {
    switch algorithm {
    case DenyOverrides:
        return combineWithDenyOverrides(results)
    case PermitOverrides:
        return combineWithPermitOverrides(results)
    case FirstApplicable:
        return combineWithFirstApplicable(results)
    default:
        return PolicyResult{Decision: Indeterminate}
    }
}

func combineWithDenyOverrides(results []PolicyResult) PolicyResult {
    var allObligations []Obligation
    var allAdvice []Advice
    hasApplicable := false
    hasPermit := false
    
    // First pass: look for any DENY
    for _, result := range results {
        if !result.Applicable {
            continue
        }
        hasApplicable = true
        
        if result.Decision == Deny {
            // Collect obligations from deny rules
            allObligations = append(allObligations, result.Obligations...)
            allAdvice = append(allAdvice, result.Advice...)
            return PolicyResult{
                Decision:    Deny,
                Obligations: allObligations,
                Advice:      allAdvice,
                Applicable:  true,
            }
        }
    }
    
    // Second pass: collect permits and obligations
    for _, result := range results {
        if result.Applicable && result.Decision == Permit {
            hasPermit = true
            allObligations = append(allObligations, result.Obligations...)
            allAdvice = append(allAdvice, result.Advice...)
        }
    }
    
    // Determine final decision
    finalDecision := NotApplicable
    if hasApplicable {
        if hasPermit {
            finalDecision = Permit
        } else {
            finalDecision = Indeterminate // Had applicable rules but no permit/deny
        }
    }
    
    return PolicyResult{
        Decision:    finalDecision,
        Obligations: allObligations,
        Advice:      allAdvice,
        Applicable:  hasApplicable,
    }
}

func combineWithPermitOverrides(results []PolicyResult) PolicyResult {
    var allObligations []Obligation
    var allAdvice []Advice
    hasApplicable := false
    
    // Look for any PERMIT first
    for _, result := range results {
        if !result.Applicable {
            continue
        }
        hasApplicable = true
        
        if result.Decision == Permit {
            allObligations = append(allObligations, result.Obligations...)
            allAdvice = append(allAdvice, result.Advice...)
            return PolicyResult{
                Decision:    Permit,
                Obligations: allObligations,
                Advice:      allAdvice,
                Applicable:  true,
            }
        }
    }
    
    // Collect denies if no permits found
    hasDeny := false
    for _, result := range results {
        if result.Applicable && result.Decision == Deny {
            hasDeny = true
            allObligations = append(allObligations, result.Obligations...)
            allAdvice = append(allAdvice, result.Advice...)
        }
    }
    
    finalDecision := NotApplicable
    if hasApplicable {
        finalDecision = Deny
        if !hasDeny {
            finalDecision = Indeterminate
        }
    }
    
    return PolicyResult{
        Decision:    finalDecision,
        Obligations: allObligations,
        Advice:      allAdvice,
        Applicable:  hasApplicable,
    }
}

func combineWithFirstApplicable(results []PolicyResult) PolicyResult {
    for _, result := range results {
        if result.Applicable {
            return result
        }
    }
    return PolicyResult{Decision: NotApplicable}
}
```

### Testing & Validation Plan

**Unit Tests:**
```go
func TestDenyOverridesCombining(t *testing.T) {
    tests := []struct {
        name     string
        results  []PolicyResult
        expected Decision
        numOblg  int
    }{
        {
            name: "deny_wins_over_permit",
            results: []PolicyResult{
                {Decision: Permit, Applicable: true, Obligations: []Obligation{{Type: "log"}}},
                {Decision: Deny, Applicable: true, Obligations: []Obligation{{Type: "alert"}}},
            },
            expected: Deny,
            numOblg:  2,
        },
        {
            name: "multiple_permits_combined",
            results: []PolicyResult{
                {Decision: Permit, Applicable: true, Obligations: []Obligation{{Type: "log1"}}},
                {Decision: Permit, Applicable: true, Obligations: []Obligation{{Type: "log2"}}},
            },
            expected: Permit,
            numOblg:  2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := combineWithDenyOverrides(tt.results)
            assert.Equal(t, tt.expected, result.Decision)
            assert.Len(t, result.Obligations, tt.numOblg)
        })
    }
}
```

### Estimated Effort
**3-4 person-days**

### Risks & Mitigations
- **Risk**: Complex obligation collection logic leading to bugs
  - **Mitigation**: Extensive unit testing, formal specification document
- **Risk**: Performance impact from multiple rule evaluation
  - **Mitigation**: Short-circuit evaluation, parallel processing

## 3. Attribute Resolution (PIP) and Caching

### Executive Summary
Implement async attribute resolution with two-tier caching (local + Redis) and intelligent invalidation strategies.

### Recommended Option
**Async PIP interface with hybrid caching (Ristretto local + Redis distributed)**

### Rationale and Trade-offs

**Async design:**
- Enables parallel attribute fetching
- Better resource utilization
- Supports batching for efficiency

**Hybrid caching:**
- Local cache (Ristretto): Ultra-low latency for hot data
- Distributed cache (Redis): Consistency across instances
- TTL-based expiration with smart invalidation

### Implementation Sketch

```go
package pip

import (
    "context"
    "fmt"
    "time"
    "github.com/dgraph-io/ristretto"
    "github.com/go-redis/redis/v8"
)

// Attribute resolver interface
type AttributeResolver interface {
    GetAttribute(ctx context.Context, req AttributeRequest) (*AttributeValue, error)
    GetAttributes(ctx context.Context, reqs []AttributeRequest) (map[string]*AttributeValue, error)
    InvalidateAttribute(ctx context.Context, path AttributePath) error
}

type AttributeRequest struct {
    TenantID   string
    Path       AttributePath
    Required   bool // Fail if not found vs return nil
}

type AttributePath struct {
    Category string // subject, resource, environment
    EntityID string
    Property string
}

type AttributeValue struct {
    Value      any
    Type       string
    Source     string
    Timestamp  time.Time
    TTL        time.Duration
    Cacheable  bool
}

// Cached attribute resolver
type CachedAttributeResolver struct {
    localCache  *ristretto.Cache
    redisClient *redis.Client
    pipServices map[string]PIPService
}

func NewCachedAttributeResolver(redisClient *redis.Client) (*CachedAttributeResolver, error) {
    localCache, err := ristretto.NewCache(&ristretto.Config{
        NumCounters: 1e4,     // 10k counters
        MaxCost:     1 << 20, // 1MB max cost
        BufferItems: 64,
    })
    if err != nil {
        return nil, err
    }
    
    return &CachedAttributeResolver{
        localCache:  localCache,
        redisClient: redisClient,
        pipServices: make(map[string]PIPService),
    }, nil
}

func (car *CachedAttributeResolver) GetAttribute(ctx context.Context, req AttributeRequest) (*AttributeValue, error) {
    cacheKey := generateAttributeCacheKey(req)
    
    // 1. Check local cache first (fastest)
    if val, found := car.localCache.Get(cacheKey); found {
        if attr := val.(*AttributeValue); !isExpired(attr) {
            return attr, nil
        }
    }
    
    // 2. Check distributed cache
    if cached, err := car.redisClient.Get(ctx, cacheKey).Result(); err == nil {
        attr := &AttributeValue{}
        if err := json.Unmarshal([]byte(cached), attr); err == nil && !isExpired(attr) {
            // Store in local cache for next time
            car.localCache.Set(cacheKey, attr, 1)
            return attr, nil
        }
    }
    
    // 3. Fetch from PIP service
    pipService, exists := car.pipServices[req.Path.Category]
    if !exists {
        if req.Required {
            return nil, fmt.Errorf("no PIP service for category: %s", req.Path.Category)
        }
        return nil, nil // Optional attribute not found
    }
    
    attr, err := pipService.GetAttribute(ctx, req)
    if err != nil {
        if req.Required {
            return nil, err
        }
        return nil, nil // Optional attribute error
    }
    
    // 4. Store in caches if cacheable
    if attr.Cacheable {
        car.storeInCaches(ctx, cacheKey, attr)
    }
    
    return attr, nil
}

func generateAttributeCacheKey(req AttributeRequest) string {
    return fmt.Sprintf("attr:%s:%s:%s:%s", 
        req.TenantID, req.Path.Category, req.Path.EntityID, req.Path.Property)
}

func (car *CachedAttributeResolver) storeInCaches(ctx context.Context, key string, attr *AttributeValue) {
    // Local cache
    car.localCache.SetWithTTL(key, attr, 1, attr.TTL)
    
    // Distributed cache
    data, _ := json.Marshal(attr)
    car.redisClient.Set(ctx, key, data, attr.TTL)
}

// Batch attribute resolution
func (car *CachedAttributeResolver) GetAttributes(ctx context.Context, reqs []AttributeRequest) (map[string]*AttributeValue, error) {
    results := make(map[string]*AttributeValue)
    var uncachedReqs []AttributeRequest
    
    // Check caches for all requests
    for _, req := range reqs {
        if attr, err := car.GetAttribute(ctx, req); err == nil && attr != nil {
            results[req.Path.String()] = attr
        } else {
            uncachedReqs = append(uncachedReqs, req)
        }
    }
    
    // Batch fetch uncached attributes
    if len(uncachedReqs) > 0 {
        batchResults, err := car.batchFetchFromPIPs(ctx, uncachedReqs)
        if err != nil {
            return results, err
        }
        
        for path, attr := range batchResults {
            results[path] = attr
        }
    }
    
    return results, nil
}
```

**Cache Strategy:**
- **Hot attributes** (user roles): 30s local, 2min distributed
- **Warm attributes** (department info): 5min local, 15min distributed  
- **Cold attributes** (real-time data): No caching
- **Static attributes** (resource metadata): 1hour local, 6hour distributed

**Invalidation Strategies:**
```go
func (car *CachedAttributeResolver) InvalidateUserAttributes(tenantID, userID string) error {
    pattern := fmt.Sprintf("attr:%s:subject:%s:*", tenantID, userID)
    return car.invalidateByPattern(pattern)
}

func (car *CachedAttributeResolver) InvalidateResourceAttributes(tenantID, resourceID string) error {
    pattern := fmt.Sprintf("attr:%s:resource:%s:*", tenantID, resourceID)
    return car.invalidateByPattern(pattern)
}

// Smart invalidation on role changes
func (car *CachedAttributeResolver) OnUserRoleChange(tenantID, userID string) error {
    // Invalidate user's attributes
    car.InvalidateUserAttributes(tenantID, userID)
    
    // Invalidate evaluation results involving this user
    evalPattern := fmt.Sprintf("eval:%s:*:subject:%s:*", tenantID, userID)
    return car.invalidateEvaluationCache(evalPattern)
}
```

### Testing & Validation Plan

**Unit Tests:**
- Cache hit/miss scenarios for each tier
- TTL expiration behavior
- Batch resolution correctness
- Invalidation pattern matching

**Integration Tests:**
- End-to-end attribute resolution with mock PIPs
- Cache consistency during concurrent access
- Invalidation propagation across instances

**Performance Tests:**
```go
func BenchmarkAttributeResolution(b *testing.B) {
    resolver := setupTestResolver()
    req := AttributeRequest{
        TenantID: "test",
        Path:     AttributePath{Category: "subject", EntityID: "user1", Property: "role"},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := resolver.GetAttribute(context.Background(), req)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Estimated Effort
**6-8 person-days**
- Days 1-2: Basic async interface
- Days 3-4: Two-tier caching implementation
- Days 5-6: Invalidation strategies
- Days 7-8: Batch operations and testing

### Risks & Mitigations
- **Risk**: Cache consistency issues in distributed setup
  - **Mitigation**: Conservative TTLs, monitoring cache hit ratios, invalidation testing
- **Risk**: PIP service failures affecting availability
  - **Mitigation**: Circuit breakers, fallback values, graceful degradation
- **Risk**: Memory usage from large local caches
  - **Mitigation**: LRU eviction, memory monitoring, configurable limits

## 4. Evaluation Caching (Result Cache)

### Executive Summary
Implement selective result caching with content-based cache keys and conservative TTLs to balance performance with correctness.

### Recommended Option
**Selective result caching with attribute fingerprinting and short TTLs**

### Rationale and Trade-offs

**When to cache:**
- Complex policies with expensive PIP calls
- Stable subject/resource combinations
- Read-heavy access patterns

**When NOT to cache:**
- Time-sensitive policies (current_time dependencies)
- Policies with random/dynamic elements
- High-frequency attribute changes

**Cache invalidation complexity vs performance tradeoff:**
- Shorter TTLs (1-5 minutes) reduce invalidation complexity
- Fingerprinting ensures correctness despite longer TTLs

### Implementation Sketch

```go
type EvaluationCacheKey struct {
    TenantID           string
    PolicyID           string  
    PolicyVersion      string
    SubjectID          string
    SubjectFingerprint string // Hash of relevant subject attributes
    ResourceID         string
    ResourceFingerprint string // Hash of relevant resource attributes
    Action             string
    EnvironmentHash    string // Hash of time-insensitive env attributes
}

type CachedEvaluationResult struct {
    Result    PolicyResult
    Timestamp time.Time
    TTL       time.Duration
}

func GenerateEvaluationCacheKey(req EvaluationRequest, relevantAttrs AttributeSet) string {
    key := EvaluationCacheKey{
        TenantID:      req.TenantID,
        PolicyID:      req.PolicyID,
        PolicyVersion: req.PolicyVersion,
        SubjectID:     req.Subject.ID,
        ResourceID:    req.Resource.ID,
        Action:        req.Action,
    }
    
    // Generate content-based fingerprints
    key.SubjectFingerprint = generateAttributeFingerprint(relevantAttrs.Subject)
    key.ResourceFingerprint = generateAttributeFingerprint(relevantAttrs.Resource)
    key.EnvironmentHash = generateEnvironmentFingerprint(relevantAttrs.Environment)
    
    return fmt.Sprintf("eval:%s:%s:%s:%s:%s:%s:%s:%s:%s",
        key.TenantID, key.PolicyID, key.PolicyVersion,
        key.SubjectID, key.SubjectFingerprint,
        key.ResourceID, key.ResourceFingerprint,
        key.Action, key.EnvironmentHash)
}

func generateAttributeFingerprint(attrs map[string]*AttributeValue) string {
    if len(attrs) == 0 {
        return "empty"
    }
    
    // Sort keys for consistent hashing
    var keys []string
    for k := range attrs {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    hasher := sha256.New()
    for _, k := range keys {
        hasher.Write([]byte(fmt.Sprintf("%s:%v", k, attrs[k].Value)))
    }
    
    return hex.EncodeToString(hasher.Sum(nil))[:16] // 16 char hash
}

type EvaluationCache struct {
    redis    *redis.Client
    enabled  map[string]bool // Policy ID -> cacheable
    ttlByPolicy map[string]time.Duration
}

func (ec *EvaluationCache) Get(ctx context.Context, key string) (*CachedEvaluationResult, error) {
    data, err := ec.redis.Get(ctx, key).Result()
    if err != nil {
        return nil, err
    }
    
    cached := &CachedEvaluationResult{}
    if err := json.Unmarshal([]byte(data), cached); err != nil {
        return nil, err
    }
    
    // Check if expired
    if time.Since(cached.Timestamp) > cached.TTL {
        ec.redis.Del(ctx, key) // Clean up
        return nil, ErrCacheExpired
    }
    
    return cached, nil
}

func (ec *EvaluationCache) Set(ctx context.Context, key string, result PolicyResult, ttl time.Duration) error {
    cached := &CachedEvaluationResult{
        Result:    result,
        Timestamp: time.Now(),
        TTL:       ttl,
    }
    
    data, err := json.Marshal(cached)
    if err != nil {
        return err
    }
    
    return ec.redis.Set(ctx, key, data, ttl).Err()
}
```

**TTL Strategy:**
- **Stable policies**: 5 minutes
- **Dynamic policies**: 1 minute  
- **High-security policies**: 30 seconds
- **Test/dev environments**: 10 seconds

**Invalidation Rules:**
```go
func (ec *EvaluationCache) InvalidateOnPolicyChange(tenantID, policyID string) error {
    pattern := fmt.Sprintf("eval:%s:%s:*", tenantID, policyID)
    return ec.deleteByPattern(pattern)
}

func (ec *EvaluationCache) InvalidateOnUserChange(tenantID, userID string) error {
    pattern := fmt.Sprintf("eval:%s:*:*:%s:*", tenantID, userID)
    return ec.deleteByPattern(pattern)
}
```

### Testing & Validation Plan

**Unit Tests:**
- Cache key generation consistency
- TTL expiration handling
- Invalidation pattern matching
- Fingerprint collision resistance

**Integration Tests:**
- Cache performance under concurrent load
- Invalidation effectiveness across instances
- Memory usage patterns

**Performance Tests:**
```go
func TestCachePerformanceImprovement(t *testing.T) {
    // Measure evaluation time with/without cache
    // Assert significant improvement for cached results
    // Verify cache hit ratio targets
}
```

### Estimated Effort
**4-5 person-days**

### Risks & Mitigations
- **Risk**: Stale cached results leading to incorrect decisions
  - **Mitigation**: Conservative TTLs,  invalidation testing
- **Risk**: Cache key collisions
  - **Mitigation**: Strong hashing algorithms, collision monitoring
- **Risk**: Redis availability affecting performance
  - **Mitigation**: Fallback to evaluation without caching, Redis clustering

## 5. Explainability & Auditing

### Executive Summary
Provide detailed evaluation explanations with rule traces and attribute access logs, stored in ClickHouse for compliance and searchability.

### Recommended Option
**Structured explanations with ClickHouse audit storage and multi-channel exposure**

### Implementation Sketch

**EvaluationExplanation JSON Schema:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "EvaluationExplanation",
  "required": ["requestId", "tenantId", "timestamp", "decision"],
  "properties": {
    "requestId": {
      "type": "string",
      "description": "Unique identifier for this evaluation"
    },
    "tenantId": {
      "type": "string",
      "description": "Tenant performing the evaluation"
    },
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "When evaluation occurred"
    },
    "decision": {
      "enum": ["PERMIT", "DENY", "NOT_APPLICABLE", "INDETERMINATE"],
      "description": "Final authorization decision"
    },
    "policySet": {
      "type": "object",
      "properties": {
        "id": {"type": "string"},
        "version": {"type": "string"},
        "combiningAlgorithm": {"type": "string"}
      }
    },
    "request": {
      "type": "object",
      "properties": {
        "subject": {"type": "object"},
        "resource": {"type": "object"},
        "action": {"type": "string"},
        "environment": {"type": "object"}
      }
    },
    "ruleTraces": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "ruleId": {"type": "string"},
          "expression": {"type": "string"},
          "result": {"enum": ["PERMIT", "DENY", "NOT_APPLICABLE", "ERROR"]},
          "evaluationTimeMs": {"type": "number"},
          "attributeAccesses": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "path": {"type": "string"},
                "value": {},
                "valueType": {"type": "string"},
                "source": {"type": "string"},
                "cacheHit": {"type": "boolean"},
                "retrievalTimeMs": {"type": "number"}
              },
              "required": ["path", "source"]
            }
          },
          "obligations": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
