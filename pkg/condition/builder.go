// Package condition provides a production-ready runtime condition evaluation engine
// for feature flags, ABAC policies, and customer workflows.
package condition

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/google/uuid"
)

// Common errors
var (
	ErrFieldNotFound         = errors.New("field not found")
	ErrFunctionNotFound      = errors.New("function not found")
	ErrInvalidOperator       = errors.New("invalid operator")
	ErrInvalidExpression     = errors.New("invalid expression")
	ErrMaxDepthExceeded      = errors.New("maximum nesting depth exceeded")
	ErrResourceLimitExceeded = errors.New("resource limit exceeded")
	ErrValidation            = errors.New("validation error")
	ErrRegexTimeout          = errors.New("regex timeout")
	ErrRegexComplexity       = errors.New("regex pattern too complex")
	ErrEvaluationTimeout     = errors.New("evaluation timeout")
	ErrFormulaComplexity     = errors.New("formula too complex")
)

// Constants for resource limits with documented rationale
const (
	// DefaultMaxDepth limits nesting to prevent stack overflow
	// 10 levels is sufficient for most business logic while preventing abuse
	DefaultMaxDepth = 10

	// DefaultMaxConditions prevents DoS via condition explosion
	// 1000 conditions covers complex workflows without excessive resource use
	DefaultMaxConditions = 1000

	// DefaultTimeout prevents long-running evaluations from blocking
	DefaultTimeout = 30 * time.Second

	// DefaultRegexTimeout prevents ReDoS attacks
	DefaultRegexTimeout = 100 * time.Millisecond

	// MaxRegexCacheSize limits memory usage from compiled regex patterns
	MaxRegexCacheSize = 1000

	// MaxRegexPatternLength prevents compilation of extremely complex patterns
	MaxRegexPatternLength = 1000

	// MaxFormulaLength prevents compilation of extremely complex formulas
	MaxFormulaLength = 10000

	// MaxLRUEvictionBatch controls how many entries to evict at once
	MaxLRUEvictionBatch = 100
)

// FieldType represents the data type of a field
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeDate     FieldType = "date"
	FieldTypeDateTime FieldType = "datetime"
	FieldTypeTime     FieldType = "time"
	FieldTypeSelect   FieldType = "select"
)

// ValueType represents how a value is provided
type ValueType string

const (
	ValueTypeValue ValueType = "value"
	ValueTypeField ValueType = "field"
	ValueTypeFunc  ValueType = "func"
)

// OperatorType represents comparison operators
type OperatorType string

const (
	OpEqual          OperatorType = "equal"
	OpNotEqual       OperatorType = "not_equal"
	OpLess           OperatorType = "less"
	OpLessOrEqual    OperatorType = "less_or_equal"
	OpGreater        OperatorType = "greater"
	OpGreaterOrEqual OperatorType = "greater_or_equal"
	OpBetween        OperatorType = "between"
	OpNotBetween     OperatorType = "not_between"
	OpIsEmpty        OperatorType = "is_empty"
	OpIsNotEmpty     OperatorType = "is_not_empty"
	OpContains       OperatorType = "contains"
	OpNotContains    OperatorType = "not_contains"
	OpStartsWith     OperatorType = "starts_with"
	OpEndsWith       OperatorType = "ends_with"
	OpIn             OperatorType = "select_any_in"
	OpNotIn          OperatorType = "select_not_any_in"
	OpMatchRegexp    OperatorType = "match_regexp"
)

// Conjunction represents logical operators
type Conjunction string

const (
	ConjunctionAnd Conjunction = "and"
	ConjunctionOr  Conjunction = "or"
)

// Field represents a field definition
type Field struct {
	Type         FieldType      `json:"type"`
	Label        string         `json:"label"`
	Name         string         `json:"name,omitempty"`
	ValueTypes   []ValueType    `json:"valueTypes,omitempty"`
	Operators    []OperatorType `json:"operators,omitempty"`
	Placeholder  string         `json:"placeholder,omitempty"`
	DefaultValue any            `json:"defaultValue,omitempty"`
	Values       []SelectOption `json:"values,omitempty"`
}

// Validate checks if the field definition is valid
func (f *Field) Validate() error {
	if f.Label == "" {
		return fmt.Errorf("%w: field label is required", ErrValidation)
	}
	if f.Type == "" {
		return fmt.Errorf("%w: field type is required", ErrValidation)
	}
	return nil
}

// SelectOption represents an option in a select field
type SelectOption struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

// FuncArg represents a function argument
type FuncArg struct {
	Type         FieldType   `json:"type"`
	Label        string      `json:"label"`
	ValueTypes   []ValueType `json:"valueTypes,omitempty"`
	DefaultValue any         `json:"defaultValue,omitempty"`
	IsOptional   bool        `json:"isOptional,omitempty"`
}

// Validate checks if the function argument is valid
func (f *FuncArg) Validate() error {
	if f.Label == "" {
		return fmt.Errorf("%w: function argument label is required", ErrValidation)
	}
	if f.Type == "" {
		return fmt.Errorf("%w: function argument type is required", ErrValidation)
	}
	return nil
}

// FieldFunc represents a function definition
type FieldFunc struct {
	Type       string      `json:"type"`
	ReturnType FieldType   `json:"returnType"`
	Args       []FuncArg   `json:"args"`
	Label      string      `json:"label"`
	Handler    FuncHandler `json:"-"`
}

// Validate checks if the function definition is valid
func (f *FieldFunc) Validate() error {
	if f.Type == "" {
		return fmt.Errorf("%w: function type is required", ErrValidation)
	}
	if f.Label == "" {
		return fmt.Errorf("%w: function label is required", ErrValidation)
	}
	if f.Handler == nil {
		return fmt.Errorf("%w: function handler is required", ErrValidation)
	}
	for i, arg := range f.Args {
		if err := arg.Validate(); err != nil {
			return fmt.Errorf("arg %d: %w", i, err)
		}
	}
	return nil
}

// FuncHandler is the function signature for custom functions
// Context is passed for cancellation and timeout support
type FuncHandler func(ctx context.Context, args []any, evalCtx *EvalContext) (any, error)

// Config represents the condition builder configuration
// WARNING: Config should not be modified after creating an Evaluator
type Config struct {
	ValueTypes []ValueType           `json:"valueTypes,omitempty"`
	Fields     []any                 `json:"fields"`
	Funcs      []any                 `json:"funcs,omitempty"`
	MaxLevel   int                   `json:"maxLevel,omitempty"`
	Types      map[string]TypeConfig `json:"types"`
}

// TypeConfig represents configuration for a specific field type
type TypeConfig struct {
	DefaultOp   OperatorType   `json:"defaultOp,omitempty"`
	Operators   []OperatorType `json:"operators"`
	Placeholder string         `json:"placeholder,omitempty"`
	ValueTypes  []ValueType    `json:"valueTypes,omitempty"`
}

// Expression represents a value or field reference
type Expression struct {
	Type  ValueType `json:"type"`
	Value any       `json:"value,omitempty"`
	Field string    `json:"field,omitempty"`
	Func  *FuncCall `json:"func,omitempty"`
}

// Validate checks if the expression is valid
func (e *Expression) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("%w: expression type is required", ErrValidation)
	}

	switch e.Type {
	case ValueTypeValue:
		// Value can be anything including nil for null checks
	case ValueTypeField:
		if e.Field == "" {
			return fmt.Errorf("%w: field name is required for field expression", ErrValidation)
		}
	case ValueTypeFunc:
		if e.Func == nil {
			return fmt.Errorf("%w: function call is required for func expression", ErrValidation)
		}
		if err := e.Func.Validate(); err != nil {
			return fmt.Errorf("function call: %w", err)
		}
	default:
		return fmt.Errorf("%w: unknown expression type %s", ErrValidation, e.Type)
	}

	return nil
}

// FuncCall represents a function call
type FuncCall struct {
	Type string `json:"type"`
	Args []any  `json:"args"`
}

// Validate checks if the function call is valid
func (f *FuncCall) Validate() error {
	if f.Type == "" {
		return fmt.Errorf("%w: function type is required", ErrValidation)
	}
	// Note: Args validation happens during evaluation when we know the function signature
	return nil
}

// ConditionRule represents a single condition
type ConditionRule struct {
	ID    string       `json:"id"`
	Left  Expression   `json:"left"`
	Op    OperatorType `json:"op"`
	Right any          `json:"right,omitempty"`
	If    string       `json:"if,omitempty"` // Formula expression using expr-lang
}

// Validate checks if the rule is valid
func (r *ConditionRule) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("%w: rule ID is required", ErrValidation)
	}

	// If this is a formula rule, don't validate traditional rule fields
	if r.If != "" {
		if len(r.If) > MaxFormulaLength {
			return fmt.Errorf("%w: formula length %d exceeds limit %d",
				ErrFormulaComplexity, len(r.If), MaxFormulaLength)
		}
		return nil
	}

	if err := r.Left.Validate(); err != nil {
		return fmt.Errorf("left expression: %w", err)
	}

	if r.Op == "" {
		return fmt.Errorf("%w: operator is required", ErrValidation)
	}

	if !isValidOperator(r.Op) {
		return fmt.Errorf("%w: unknown operator %s", ErrInvalidOperator, r.Op)
	}

	// Unary operators don't need a right value
	if !isUnaryOperator(r.Op) {
		if r.Right == nil {
			return fmt.Errorf("%w: right value is required for operator %s", ErrValidation, r.Op)
		}
	}

	return nil
}

// isUnaryOperator checks if an operator doesn't require a right operand
func isUnaryOperator(op OperatorType) bool {
	return op == OpIsEmpty || op == OpIsNotEmpty
}

// isValidOperator checks if an operator is recognized
func isValidOperator(op OperatorType) bool {
	switch op {
	case OpEqual, OpNotEqual, OpLess, OpLessOrEqual, OpGreater, OpGreaterOrEqual,
		OpBetween, OpNotBetween, OpIsEmpty, OpIsNotEmpty, OpContains, OpNotContains,
		OpStartsWith, OpEndsWith, OpIn, OpNotIn, OpMatchRegexp:
		return true
	default:
		return false
	}
}

// ConditionGroup represents a group of conditions
type ConditionGroup struct {
	ID          string      `json:"id"`
	Conjunction Conjunction `json:"conjunction"`
	Not         bool        `json:"not,omitempty"`
	Children    []any       `json:"children"`
	If          string      `json:"if,omitempty"` // Formula expression using expr-lang
}

// Validate checks if the group is valid
func (g *ConditionGroup) Validate() error {
	if g.ID == "" {
		return fmt.Errorf("%w: group ID is required", ErrValidation)
	}

	if g.Conjunction != ConjunctionAnd && g.Conjunction != ConjunctionOr {
		return fmt.Errorf("%w: invalid conjunction %s", ErrValidation, g.Conjunction)
	}

	if len(g.Children) == 0 && g.If == "" {
		return fmt.Errorf("%w: group must have at least one child or a formula", ErrValidation)
	}

	if g.If != "" && len(g.If) > MaxFormulaLength {
		return fmt.Errorf("%w: formula length %d exceeds limit %d",
			ErrFormulaComplexity, len(g.If), MaxFormulaLength)
	}

	return nil
}

// EvalContext holds runtime evaluation context
// Thread-safe for concurrent reads, writes must be externally synchronized
type EvalContext struct {
	mu        sync.RWMutex
	Data      map[string]any
	Functions map[string]FuncHandler
	Fields    map[string]Field
	Now       time.Time

	// Resource limits
	maxConditions  int
	conditionCount int32 // Use atomic for thread-safety

	// Metrics for observability
	metrics   EvaluationMetrics
	metricsMu sync.Mutex // Protects Duration field
}

// EvaluationMetrics tracks evaluation statistics
type EvaluationMetrics struct {
	RulesEvaluated  int32
	GroupsEvaluated int32
	Duration        time.Duration
	CacheHits       int32
	CacheMisses     int32
	Errors          int32
}

// GetMetrics returns a copy of the current metrics (thread-safe)
func (ctx *EvalContext) GetMetrics() EvaluationMetrics {
	ctx.metricsMu.Lock()
	defer ctx.metricsMu.Unlock()

	return EvaluationMetrics{
		RulesEvaluated:  atomic.LoadInt32(&ctx.metrics.RulesEvaluated),
		GroupsEvaluated: atomic.LoadInt32(&ctx.metrics.GroupsEvaluated),
		Duration:        ctx.metrics.Duration,
		CacheHits:       atomic.LoadInt32(&ctx.metrics.CacheHits),
		CacheMisses:     atomic.LoadInt32(&ctx.metrics.CacheMisses),
		Errors:          atomic.LoadInt32(&ctx.metrics.Errors),
	}
}

// EvalOptions configures evaluation behavior
type EvalOptions struct {
	MaxDepth      int
	MaxConditions int
	Timeout       time.Duration
	CacheResults  bool
	RegexTimeout  time.Duration
}

// DefaultEvalOptions returns sensible defaults
func DefaultEvalOptions() EvalOptions {
	return EvalOptions{
		MaxDepth:      DefaultMaxDepth,
		MaxConditions: DefaultMaxConditions,
		Timeout:       DefaultTimeout,
		CacheResults:  true,
		RegexTimeout:  DefaultRegexTimeout,
	}
}

// NewEvalContext creates a new evaluation context
func NewEvalContext(data map[string]any, opts EvalOptions) *EvalContext {
	if data == nil {
		data = make(map[string]any)
	}

	return &EvalContext{
		Data:          data,
		Functions:     make(map[string]FuncHandler),
		Fields:        make(map[string]Field),
		Now:           time.Now(),
		maxConditions: opts.MaxConditions,
	}
}

// RegisterFunction registers a custom function (thread-safe)
func (ctx *EvalContext) RegisterFunction(name string, handler FuncHandler) error {
	if name == "" {
		return fmt.Errorf("%w: function name is required", ErrValidation)
	}
	if handler == nil {
		return fmt.Errorf("%w: function handler is required", ErrValidation)
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.Functions[name] = handler
	return nil
}

// RegisterField registers a field definition (thread-safe)
func (ctx *EvalContext) RegisterField(field Field) error {
	if err := field.Validate(); err != nil {
		return err
	}
	if field.Name == "" {
		return fmt.Errorf("%w: field name is required", ErrValidation)
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.Fields[field.Name] = field
	return nil
}

// GetValue retrieves a value from context data (thread-safe)
// Supports nested field access via dot notation (e.g., "user.profile.email")
func (ctx *EvalContext) GetValue(path string) (any, error) {
	if path == "" {
		return nil, fmt.Errorf("%w: empty path", ErrFieldNotFound)
	}

	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	parts := strings.Split(path, ".")
	var current any = ctx.Data

	for i, part := range parts {
		if current == nil {
			return nil, fmt.Errorf("%w: nil value at %s", ErrFieldNotFound, strings.Join(parts[:i], "."))
		}

		switch v := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrFieldNotFound, path)
			}
		default:
			// Use reflection as fallback for struct fields
			rv := reflect.ValueOf(current)
			if rv.Kind() == reflect.Pointer {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Struct {
				field := rv.FieldByName(part)
				if !field.IsValid() {
					return nil, fmt.Errorf("%w: %s", ErrFieldNotFound, path)
				}
				current = field.Interface()
			} else {
				return nil, fmt.Errorf("cannot access field %s on non-map/struct type", part)
			}
		}
	}

	return current, nil
}

// lruEntry tracks access time for LRU eviction
type lruEntry struct {
	pattern    string
	regex      *regexp2.Regexp
	lastAccess int64 // Unix timestamp
}

// regexCache implements a bounded LRU cache for compiled regex patterns
type regexCache struct {
	entries sync.Map // map[string]*lruEntry
	size    atomic.Int32
	max     int32
}

// newRegexCache creates a new bounded regex cache
func newRegexCache(maxSize int32) *regexCache {
	return &regexCache{
		max: maxSize,
	}
}

// Get retrieves a compiled regex from cache and updates access time
func (rc *regexCache) Get(pattern string) (*regexp2.Regexp, bool) {
	if val, ok := rc.entries.Load(pattern); ok {
		entry := val.(*lruEntry)
		atomic.StoreInt64(&entry.lastAccess, time.Now().Unix())
		return entry.regex, true
	}
	return nil, false
}

// Set stores a compiled regex in cache with LRU eviction
func (rc *regexCache) Set(pattern string, re *regexp2.Regexp) {
	entry := &lruEntry{
		pattern:    pattern,
		regex:      re,
		lastAccess: time.Now().Unix(),
	}

	// Try to store
	if _, loaded := rc.entries.LoadOrStore(pattern, entry); loaded {
		return // Already exists
	}

	// Check if we need to evict
	newSize := rc.size.Add(1)
	if newSize > rc.max {
		rc.evictLRU()
	}
}

// evictLRU removes least recently used entries
func (rc *regexCache) evictLRU() {
	var entries []*lruEntry

	// Collect all entries
	rc.entries.Range(func(key, value any) bool {
		entries = append(entries, value.(*lruEntry))
		return true
	})

	if len(entries) == 0 {
		return
	}

	// Sort by access time (oldest first)
	// Simple selection of oldest entries without full sort for performance
	toEvict := min(MaxLRUEvictionBatch, len(entries)/4) // Evict 25% or batch size
	if toEvict == 0 {
		toEvict = 1
	}

	// Find oldest entries using partial selection
	for i := 0; i < toEvict; i++ {
		oldestIdx := i
		oldestTime := atomic.LoadInt64(&entries[i].lastAccess)

		for j := i + 1; j < len(entries); j++ {
			accessTime := atomic.LoadInt64(&entries[j].lastAccess)
			if accessTime < oldestTime {
				oldestIdx = j
				oldestTime = accessTime
			}
		}

		// Swap to position i
		if oldestIdx != i {
			entries[i], entries[oldestIdx] = entries[oldestIdx], entries[i]
		}

		// Delete the oldest entry
		rc.entries.Delete(entries[i].pattern)
		rc.size.Add(-1)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Evaluator evaluates conditions with resource limits
// Thread-safe for concurrent evaluations
type Evaluator struct {
	config       *Config
	opts         EvalOptions
	regexCache   *regexCache
	programCache sync.Map // Cache compiled expr programs for formulas
}

// NewEvaluator creates a new condition evaluator
// The config should not be modified after creating the evaluator
func NewEvaluator(config *Config, opts EvalOptions) *Evaluator {
	// Apply defaults if not set
	if opts.MaxDepth == 0 {
		opts.MaxDepth = DefaultMaxDepth
	}
	if opts.MaxConditions == 0 {
		opts.MaxConditions = DefaultMaxConditions
	}
	if opts.Timeout == 0 {
		opts.Timeout = DefaultTimeout
	}
	if opts.RegexTimeout == 0 {
		opts.RegexTimeout = DefaultRegexTimeout
	}

	// Config MaxLevel takes precedence if it's more restrictive
	if config != nil && config.MaxLevel > 0 && config.MaxLevel < opts.MaxDepth {
		opts.MaxDepth = config.MaxLevel
	}

	return &Evaluator{
		config:     config,
		opts:       opts,
		regexCache: newRegexCache(MaxRegexCacheSize),
	}
}

// Evaluate evaluates a condition with context and timeout
// This method is thread-safe and can be called concurrently
func (e *Evaluator) Evaluate(ctx context.Context, root any, evalCtx *EvalContext) (bool, error) {
	if root == nil {
		return false, fmt.Errorf("%w: nil root condition", ErrInvalidExpression)
	}
	if evalCtx == nil {
		return false, fmt.Errorf("%w: nil evaluation context", ErrInvalidExpression)
	}

	startTime := time.Now()
	defer func() {
		evalCtx.metricsMu.Lock()
		evalCtx.metrics.Duration = time.Since(startTime)
		evalCtx.metricsMu.Unlock()
	}()

	// Apply timeout if configured
	if e.opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.opts.Timeout)
		defer cancel()
	}

	result, err := e.evaluateWithDepth(ctx, root, evalCtx, 0)
	if err != nil {
		atomic.AddInt32(&evalCtx.metrics.Errors, 1)

		// Wrap timeout errors for clarity
		if errors.Is(err, context.DeadlineExceeded) {
			return false, fmt.Errorf("%w: %v", ErrEvaluationTimeout, err)
		}
	}
	return result, err
}

// evaluateWithDepth handles depth tracking and delegates to specific evaluators
func (e *Evaluator) evaluateWithDepth(ctx context.Context, node any, evalCtx *EvalContext, depth int) (bool, error) {
	// Check context cancellation early
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	// Check depth limit to prevent stack overflow
	if depth > e.opts.MaxDepth {
		return false, fmt.Errorf("%w: depth %d exceeds limit %d", ErrMaxDepthExceeded, depth, e.opts.MaxDepth)
	}

	// Check condition count to prevent resource exhaustion
	count := atomic.AddInt32(&evalCtx.conditionCount, 1)
	if int(count) > evalCtx.maxConditions {
		return false, fmt.Errorf("%w: %d conditions", ErrResourceLimitExceeded, count)
	}

	// Type switch for efficient type handling (avoid JSON marshaling)
	switch v := node.(type) {
	case *ConditionGroup:
		if err := v.Validate(); err != nil {
			return false, fmt.Errorf("group %s: %w", v.ID, err)
		}
		result, err := e.evaluateGroup(ctx, v, evalCtx, depth)
		if err != nil {
			return false, fmt.Errorf("group %s: %w", v.ID, err)
		}
		return result, nil

	case ConditionGroup:
		if err := v.Validate(); err != nil {
			return false, fmt.Errorf("group %s: %w", v.ID, err)
		}
		result, err := e.evaluateGroup(ctx, &v, evalCtx, depth)
		if err != nil {
			return false, fmt.Errorf("group %s: %w", v.ID, err)
		}
		return result, nil

	case *ConditionRule:
		if err := v.Validate(); err != nil {
			return false, fmt.Errorf("rule %s: %w", v.ID, err)
		}
		result, err := e.evaluateRule(ctx, v, evalCtx)
		if err != nil {
			return false, fmt.Errorf("rule %s: %w", v.ID, err)
		}
		return result, nil

	case ConditionRule:
		if err := v.Validate(); err != nil {
			return false, fmt.Errorf("rule %s: %w", v.ID, err)
		}
		result, err := e.evaluateRule(ctx, &v, evalCtx)
		if err != nil {
			return false, fmt.Errorf("rule %s: %w", v.ID, err)
		}
		return result, nil

	case map[string]any:
		// Fallback for dynamic JSON-like structures
		return e.evaluateMapNode(ctx, v, evalCtx, depth)

	default:
		return false, fmt.Errorf("%w: unknown node type %T", ErrInvalidExpression, node)
	}
}

// evaluateMapNode handles map[string]any nodes by converting to typed structures
func (e *Evaluator) evaluateMapNode(ctx context.Context, m map[string]any, evalCtx *EvalContext, depth int) (bool, error) {
	// Extract ID for better error messages
	id, _ := m["id"].(string)
	if id == "" {
		id = "unknown"
	}

	if _, hasConjunction := m["conjunction"]; hasConjunction {
		var group ConditionGroup
		if err := mapToStruct(m, &group); err != nil {
			return false, fmt.Errorf("group %s: invalid structure: %w", id, err)
		}
		if err := group.Validate(); err != nil {
			return false, fmt.Errorf("group %s: %w", id, err)
		}
		result, err := e.evaluateGroup(ctx, &group, evalCtx, depth)
		if err != nil {
			return false, fmt.Errorf("group %s: %w", id, err)
		}
		return result, nil
	}

	var rule ConditionRule
	if err := mapToStruct(m, &rule); err != nil {
		return false, fmt.Errorf("rule %s: invalid structure: %w", id, err)
	}
	if err := rule.Validate(); err != nil {
		return false, fmt.Errorf("rule %s: %w", id, err)
	}
	result, err := e.evaluateRule(ctx, &rule, evalCtx)
	if err != nil {
		return false, fmt.Errorf("rule %s: %w", id, err)
	}
	return result, nil
}

// mapToStruct converts a map to a struct using JSON marshaling
// This is a performance compromise for dynamic data
func mapToStruct(m map[string]any, target any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}
	return nil
}

// evaluateGroup evaluates a condition group with AND/OR/NOT logic
func (e *Evaluator) evaluateGroup(ctx context.Context, group *ConditionGroup, evalCtx *EvalContext, depth int) (bool, error) {
	atomic.AddInt32(&evalCtx.metrics.GroupsEvaluated, 1)

	// Evaluate formula if present using expr-lang
	if group.If != "" {
		result, err := e.evaluateFormula(ctx, group.If, evalCtx)
		if err != nil {
			return false, fmt.Errorf("formula evaluation failed: %w", err)
		}
		// Formula takes precedence over children evaluation
		boolResult, ok := result.(bool)
		if !ok {
			return false, fmt.Errorf("formula must return boolean, got %T", result)
		}
		if group.Not {
			boolResult = !boolResult
		}
		return boolResult, nil
	}

	// Evaluate all children with short-circuit optimization
	var finalResult bool
	if group.Conjunction == ConjunctionAnd {
		finalResult = true
		for i, child := range group.Children {
			result, err := e.evaluateWithDepth(ctx, child, evalCtx, depth+1)
			if err != nil {
				return false, fmt.Errorf("child %d: %w", i, err)
			}
			finalResult = finalResult && result
			// Short-circuit on first false for AND
			if !finalResult {
				break
			}
		}
	} else { // ConjunctionOr
		finalResult = false
		for i, child := range group.Children {
			result, err := e.evaluateWithDepth(ctx, child, evalCtx, depth+1)
			if err != nil {
				return false, fmt.Errorf("child %d: %w", i, err)
			}
			finalResult = finalResult || result
			// Short-circuit on first true for OR
			if finalResult {
				break
			}
		}
	}

	// Apply NOT negation
	if group.Not {
		finalResult = !finalResult
	}

	return finalResult, nil
}

// evaluateFormula evaluates an expr-lang formula with complexity limits
func (e *Evaluator) evaluateFormula(ctx context.Context, formula string, evalCtx *EvalContext) (any, error) {
	if formula == "" {
		return nil, errors.New("empty formula")
	}

	// Check formula length to prevent compilation of overly complex expressions
	if len(formula) > MaxFormulaLength {
		return nil, fmt.Errorf("%w: formula length %d exceeds limit %d",
			ErrFormulaComplexity, len(formula), MaxFormulaLength)
	}

	// Check cache for compiled program
	var program *vm.Program
	if cached, ok := e.programCache.Load(formula); ok {
		program = cached.(*vm.Program)
		atomic.AddInt32(&evalCtx.metrics.CacheHits, 1)
	} else {
		atomic.AddInt32(&evalCtx.metrics.CacheMisses, 1)

		// Compile the expression
		// Create environment with available functions and data
		env := make(map[string]any)

		// Add context data
		evalCtx.mu.RLock()
		for k, v := range evalCtx.Data {
			env[k] = v
		}
		evalCtx.mu.RUnlock()

		// Compile with type checking
		compiled, err := expr.Compile(formula, expr.Env(env), expr.AsBool())
		if err != nil {
			return nil, fmt.Errorf("formula compilation failed: %w", err)
		}

		program = compiled
		e.programCache.Store(formula, program)
	}

	// Prepare runtime environment
	env := make(map[string]any)
	evalCtx.mu.RLock()
	for k, v := range evalCtx.Data {
		env[k] = v
	}
	evalCtx.mu.RUnlock()

	// Add special variables
	env["now"] = evalCtx.Now

	// Execute the program with context awareness
	// Note: expr-lang doesn't natively support context cancellation
	// We rely on the parent evaluation timeout
	result, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("formula execution failed: %w", err)
	}

	return result, nil
}

// evaluateRule evaluates a single condition rule
func (e *Evaluator) evaluateRule(ctx context.Context, rule *ConditionRule, evalCtx *EvalContext) (bool, error) {
	atomic.AddInt32(&evalCtx.metrics.RulesEvaluated, 1)

	// Evaluate formula if present (takes precedence)
	if rule.If != "" {
		result, err := e.evaluateFormula(ctx, rule.If, evalCtx)
		if err != nil {
			return false, fmt.Errorf("rule formula evaluation failed: %w", err)
		}
		boolResult, ok := result.(bool)
		if !ok {
			return false, fmt.Errorf("rule formula must return boolean, got %T", result)
		}
		return boolResult, nil
	}

	// Evaluate left expression
	leftVal, err := e.evaluateExpression(ctx, rule.Left, evalCtx)
	if err != nil {
		return false, fmt.Errorf("left expression: %w", err)
	}

	// Handle unary operators (don't need right value)
	if isUnaryOperator(rule.Op) {
		return e.applyUnaryOperator(rule.Op, leftVal)
	}

	// Evaluate right expression(s)
	rightVals, err := e.evaluateRightExpression(ctx, rule.Right, evalCtx)
	if err != nil {
		return false, fmt.Errorf("right expression: %w", err)
	}

	return e.applyOperator(ctx, rule.Op, leftVal, rightVals)
}

// evaluateRightExpression handles the various forms of right expressions
func (e *Evaluator) evaluateRightExpression(ctx context.Context, right any, evalCtx *EvalContext) ([]any, error) {
	if right == nil {
		return nil, errors.New("right expression is nil")
	}

	var rightVals []any

	switch v := right.(type) {
	case []any:
		// Array of expressions (for IN, BETWEEN, etc.)
		// Special handling: check if this looks like a literal array value
		// vs an array of Expression objects
		if len(v) > 0 {
			// Peek at first item to determine intent
			switch v[0].(type) {
			case Expression, map[string]any:
				// Treat as array of expressions
				for i, item := range v {
					expr, err := e.toExpression(item)
					if err != nil {
						return nil, fmt.Errorf("item %d: %w", i, err)
					}
					val, err := e.evaluateExpression(ctx, expr, evalCtx)
					if err != nil {
						return nil, fmt.Errorf("item %d: %w", i, err)
					}
					rightVals = append(rightVals, val)
				}
			default:
				// Treat as literal array values for IN operator
				rightVals = v
			}
		}

	case Expression:
		val, err := e.evaluateExpression(ctx, v, evalCtx)
		if err != nil {
			return nil, err
		}
		rightVals = []any{val}

	case map[string]any:
		// Try to convert to Expression
		expr, err := e.toExpression(v)
		if err != nil {
			return nil, err
		}
		val, err := e.evaluateExpression(ctx, expr, evalCtx)
		if err != nil {
			return nil, err
		}
		rightVals = []any{val}

	default:
		// Treat as literal value
		rightVals = []any{v}
	}

	return rightVals, nil
}

// toExpression converts various types to Expression
func (e *Evaluator) toExpression(v any) (Expression, error) {
	if v == nil {
		return Expression{Type: ValueTypeValue, Value: nil}, nil
	}

	if expr, ok := v.(Expression); ok {
		return expr, nil
	}

	// Try map conversion
	if m, ok := v.(map[string]any); ok {
		var expr Expression
		if err := mapToStruct(m, &expr); err != nil {
			return Expression{}, fmt.Errorf("invalid expression: %w", err)
		}
		return expr, nil
	}

	// Treat as literal value
	return Expression{Type: ValueTypeValue, Value: v}, nil
}

// evaluateExpression evaluates an expression to a concrete value
func (e *Evaluator) evaluateExpression(ctx context.Context, expr Expression, evalCtx *EvalContext) (any, error) {
	if err := expr.Validate(); err != nil {
		return nil, err
	}

	switch expr.Type {
	case ValueTypeValue:
		return expr.Value, nil

	case ValueTypeField:
		return evalCtx.GetValue(expr.Field)

	case ValueTypeFunc:
		if expr.Func == nil {
			return nil, errors.New("function call is nil")
		}

		evalCtx.mu.RLock()
		handler, ok := evalCtx.Functions[expr.Func.Type]
		evalCtx.mu.RUnlock()

		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrFunctionNotFound, expr.Func.Type)
		}

		// Pass context to function handler for cancellation support
		return handler(ctx, expr.Func.Args, evalCtx)

	default:
		return nil, fmt.Errorf("%w: unknown type %s", ErrInvalidExpression, expr.Type)
	}
}

// applyUnaryOperator applies operators that only need a left value
func (e *Evaluator) applyUnaryOperator(op OperatorType, left any) (bool, error) {
	switch op {
	case OpIsEmpty:
		return e.isEmpty(left), nil
	case OpIsNotEmpty:
		return !e.isEmpty(left), nil
	default:
		return false, fmt.Errorf("%w: %s is not a unary operator", ErrInvalidOperator, op)
	}
}

// applyOperator applies binary operators with proper type handling and validation
func (e *Evaluator) applyOperator(ctx context.Context, op OperatorType, left any, right []any) (bool, error) {
	// Validate right operands based on operator requirements
	switch op {
	case OpBetween, OpNotBetween:
		if len(right) != 2 {
			return false, fmt.Errorf("%s requires exactly 2 values, got %d", op, len(right))
		}
	case OpIn, OpNotIn:
		// Can have any number of values
		if len(right) == 0 {
			return false, fmt.Errorf("%s requires at least one value", op)
		}
	default:
		// Most operators need exactly 1 right value
		if len(right) == 0 {
			return false, fmt.Errorf("%s requires a right value", op)
		}
	}

	switch op {
	case OpEqual:
		return e.compare(left, right[0]) == 0, nil

	case OpNotEqual:
		return e.compare(left, right[0]) != 0, nil

	case OpLess:
		return e.compare(left, right[0]) < 0, nil

	case OpLessOrEqual:
		return e.compare(left, right[0]) <= 0, nil

	case OpGreater:
		return e.compare(left, right[0]) > 0, nil

	case OpGreaterOrEqual:
		return e.compare(left, right[0]) >= 0, nil

	case OpBetween:
		// Check if left is between right[0] and right[1] (inclusive)
		return e.compare(left, right[0]) >= 0 && e.compare(left, right[1]) <= 0, nil

	case OpNotBetween:
		// Check if left is outside the range
		return e.compare(left, right[0]) < 0 || e.compare(left, right[1]) > 0, nil

	case OpContains:
		return e.contains(left, right[0]), nil

	case OpNotContains:
		return !e.contains(left, right[0]), nil

	case OpStartsWith:
		return e.startsWith(left, right[0]), nil

	case OpEndsWith:
		return e.endsWith(left, right[0]), nil

	case OpIn:
		return e.in(left, right), nil

	case OpNotIn:
		return !e.in(left, right), nil

	case OpMatchRegexp:
		// Calculate remaining context timeout for regex
		deadline, hasDeadline := ctx.Deadline()
		regexTimeout := e.opts.RegexTimeout
		if hasDeadline {
			remaining := time.Until(deadline)
			if remaining < regexTimeout {
				regexTimeout = remaining
			}
			if regexTimeout <= 0 {
				return false, context.DeadlineExceeded
			}
		}
		return e.matchRegexp(left, right[0], regexTimeout)

	default:
		return false, fmt.Errorf("%w: %s", ErrInvalidOperator, op)
	}
}

// Helper functions with improved null handling and type safety

// isEmpty checks if a value is considered empty
// Handles nil, empty strings, empty collections, and zero values appropriately
func (e *Evaluator) isEmpty(val any) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Invalid:
		return true
	default:
		return false
	}
}

// compare performs type-aware comparison of two values
// Returns: -1 if left < right, 0 if equal, 1 if left > right
func (e *Evaluator) compare(left, right any) int {
	// Handle nil values consistently
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return -1
	}
	if right == nil {
		return 1
	}

	// Try numeric comparison first (most common in business logic)
	leftNum, leftOk := e.toFloat64(left)
	rightNum, rightOk := e.toFloat64(right)

	if leftOk && rightOk {
		if leftNum < rightNum {
			return -1
		} else if leftNum > rightNum {
			return 1
		}
		return 0
	}

	// Try time comparison for date/datetime fields
	leftTime, leftOk := e.toTime(left)
	rightTime, rightOk := e.toTime(right)

	if leftOk && rightOk {
		if leftTime.Before(rightTime) {
			return -1
		} else if leftTime.After(rightTime) {
			return 1
		}
		return 0
	}

	// Try boolean comparison
	leftBool, leftOk := left.(bool)
	rightBool, rightOk := right.(bool)
	if leftOk && rightOk {
		if leftBool == rightBool {
			return 0
		}
		if !leftBool && rightBool {
			return -1
		}
		return 1
	}

	// Fallback to string comparison (case-sensitive)
	// Note: This may give unexpected results for numeric strings
	// but is necessary for text field comparisons
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	return strings.Compare(leftStr, rightStr)
}

// toFloat64 attempts to convert a value to float64
// Handles all numeric types including unsigned integers
func (e *Evaluator) toFloat64(val any) (float64, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case int16:
		return float64(v), true
	case int8:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint8:
		return float64(v), true
	default:
		return 0, false
	}
}

// toTime attempts to convert a value to time.Time
// Supports multiple common date/time formats
func (e *Evaluator) toTime(val any) (time.Time, bool) {
	if val == nil {
		return time.Time{}, false
	}

	switch v := val.(type) {
	case time.Time:
		return v, true
	case string:
		// Try common ISO formats first (most likely in APIs)
		formats := []string{
			time.RFC3339,          // "2006-01-02T15:04:05Z07:00"
			time.RFC3339Nano,      // "2006-01-02T15:04:05.999999999Z07:00"
			"2006-01-02T15:04:05", // ISO without timezone
			"2006-01-02",          // Date only
			"15:04:05",            // Time only
			"2006-01-02 15:04:05", // Common database format
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

// contains checks if haystack contains needle (case-sensitive substring match)
func (e *Evaluator) contains(haystack, needle any) bool {
	if haystack == nil || needle == nil {
		return false
	}

	// Handle slice/array types
	haystackVal := reflect.ValueOf(haystack)
	if haystackVal.Kind() == reflect.Slice || haystackVal.Kind() == reflect.Array {
		needleStr := fmt.Sprintf("%v", needle)
		for i := 0; i < haystackVal.Len(); i++ {
			itemStr := fmt.Sprintf("%v", haystackVal.Index(i).Interface())
			if strings.Contains(itemStr, needleStr) {
				return true
			}
		}
		return false
	}

	haystackStr := fmt.Sprintf("%v", haystack)
	needleStr := fmt.Sprintf("%v", needle)
	return strings.Contains(haystackStr, needleStr)
}

// startsWith checks if str starts with prefix (case-sensitive)
func (e *Evaluator) startsWith(str, prefix any) bool {
	if str == nil || prefix == nil {
		return false
	}
	strVal := fmt.Sprintf("%v", str)
	prefixVal := fmt.Sprintf("%v", prefix)
	return strings.HasPrefix(strVal, prefixVal)
}

// endsWith checks if str ends with suffix (case-sensitive)
func (e *Evaluator) endsWith(str, suffix any) bool {
	if str == nil || suffix == nil {
		return false
	}
	strVal := fmt.Sprintf("%v", str)
	suffixVal := fmt.Sprintf("%v", suffix)
	return strings.HasSuffix(strVal, suffixVal)
}

// in checks if needle exists in haystack array using compare for type-aware matching
func (e *Evaluator) in(needle any, haystack []any) bool {
	for _, item := range haystack {
		if e.compare(needle, item) == 0 {
			return true
		}
	}
	return false
}

// matchRegexp performs regex matching with timeout and complexity limits
// Uses regexp2 for ReDoS protection
func (e *Evaluator) matchRegexp(str, pattern any, timeout time.Duration) (bool, error) {
	if str == nil || pattern == nil {
		return false, nil
	}

	strVal := fmt.Sprintf("%v", str)
	patternVal := fmt.Sprintf("%v", pattern)

	// Validate pattern length to prevent DoS
	if len(patternVal) > MaxRegexPatternLength {
		return false, fmt.Errorf("%w: pattern length %d exceeds limit %d",
			ErrRegexComplexity, len(patternVal), MaxRegexPatternLength)
	}

	// Check cache first
	var re *regexp2.Regexp
	if cached, ok := e.regexCache.Get(patternVal); ok {
		re = cached
	} else {
		// Compile new regex
		var err error
		re, err = regexp2.Compile(patternVal, regexp2.None)
		if err != nil {
			return false, fmt.Errorf("invalid regexp: %w", err)
		}
		e.regexCache.Set(patternVal, re)
	}

	// Set match timeout to prevent ReDoS
	re.MatchTimeout = timeout

	// Perform match
	matched, err := re.MatchString(strVal)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			return false, fmt.Errorf("%w: pattern took too long to evaluate", ErrRegexTimeout)
		}
		return false, fmt.Errorf("regexp match error: %w", err)
	}

	return matched, nil
}

// Builder provides a fluent interface for building conditions programmatically
type Builder struct {
	group *ConditionGroup
}

func (b *Builder) AddRuleWithID(ruleID string, field string, op OperatorType, value any) {
	rule := &ConditionRule{
		ID: ruleID,
		Left: Expression{
			Type:  ValueTypeField,
			Field: field,
		},
		Op:    op,
		Right: value,
	}

	b.group.Children = append(b.group.Children, rule)
}

// NewBuilder creates a new condition builder with the specified conjunction
func NewBuilder(conjunction Conjunction) *Builder {
	return &Builder{
		group: &ConditionGroup{
			ID:          uuid.New().String(),
			Conjunction: conjunction,
			Children:    make([]any, 0),
		},
	}
}

// AddRule adds a simple field comparison rule
func (b *Builder) AddRule(fieldName string, op OperatorType, value any) *Builder {
	if fieldName == "" {
		return b // Skip invalid rules silently in builder pattern
	}
	if !isValidOperator(op) {
		return b
	}

	rule := ConditionRule{
		ID: uuid.New().String(),
		Left: Expression{
			Type:  ValueTypeField,
			Field: fieldName,
		},
		Op: op,
		Right: Expression{
			Type:  ValueTypeValue,
			Value: value,
		},
	}
	b.group.Children = append(b.group.Children, rule)
	return b
}

// AddFieldComparison adds a rule comparing two fields
func (b *Builder) AddFieldComparison(leftField, rightField string, op OperatorType) *Builder {
	if leftField == "" || rightField == "" {
		return b
	}
	if !isValidOperator(op) {
		return b
	}

	rule := ConditionRule{
		ID: uuid.New().String(),
		Left: Expression{
			Type:  ValueTypeField,
			Field: leftField,
		},
		Op: op,
		Right: Expression{
			Type:  ValueTypeField,
			Field: rightField,
		},
	}
	b.group.Children = append(b.group.Children, rule)
	return b
}

// AddBetweenRule adds a BETWEEN comparison rule
func (b *Builder) AddBetweenRule(fieldName string, min, max any) *Builder {
	if fieldName == "" {
		return b
	}

	rule := ConditionRule{
		ID: uuid.New().String(),
		Left: Expression{
			Type:  ValueTypeField,
			Field: fieldName,
		},
		Op: OpBetween,
		Right: []any{
			Expression{Type: ValueTypeValue, Value: min},
			Expression{Type: ValueTypeValue, Value: max},
		},
	}
	b.group.Children = append(b.group.Children, rule)
	return b
}

// AddInRule adds an IN comparison rule
func (b *Builder) AddInRule(fieldName string, values ...any) *Builder {
	if fieldName == "" || len(values) == 0 {
		return b
	}

	rightVals := make([]any, len(values))
	for i, v := range values {
		rightVals[i] = Expression{Type: ValueTypeValue, Value: v}
	}

	rule := ConditionRule{
		ID: uuid.New().String(),
		Left: Expression{
			Type:  ValueTypeField,
			Field: fieldName,
		},
		Op:    OpIn,
		Right: rightVals,
	}
	b.group.Children = append(b.group.Children, rule)
	return b
}

// AddFormula adds a rule with an expr-lang formula
func (b *Builder) AddFormula(formula string) *Builder {
	if formula == "" {
		return b
	}
	if len(formula) > MaxFormulaLength {
		return b
	}

	rule := ConditionRule{
		ID: uuid.New().String(),
		If: formula,
	}
	b.group.Children = append(b.group.Children, rule)
	return b
}

// AddGroup adds a nested condition group
func (b *Builder) AddGroup(group *ConditionGroup) *Builder {
	if group != nil {
		b.group.Children = append(b.group.Children, group)
	}
	return b
}

// Not negates the entire group
func (b *Builder) Not() *Builder {
	b.group.Not = true
	return b
}

// SetFormula sets an expr-lang formula for the entire group
// When set, this overrides child evaluation
func (b *Builder) SetFormula(formula string) *Builder {
	if formula == "" || len(formula) > MaxFormulaLength {
		return b
	}
	b.group.If = formula
	return b
}

// Build returns the constructed condition group
func (b *Builder) Build() *ConditionGroup {
	return b.group
}

// Validate checks if the built condition is valid
func (b *Builder) Validate() error {
	if b.group == nil {
		return fmt.Errorf("%w: builder has no group", ErrValidation)
	}
	return b.group.Validate()
}
