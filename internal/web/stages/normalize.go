package stages

import (
	"fmt"
	"net/http"
	"strings"

	"awo.so/internal/pipeline"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/web/ui"
)

// ─── INVARIANT: Normalize vs Validate ────────────────────────────────────────
//
// NormalizeStage and ValidateStage have strictly separate contracts:
//
//   NormalizeStage (priority 60):
//     - Pure canonicalization. Mutates the schema to a standard form.
//     - MUST NEVER return an error. If something cannot be silently corrected,
//       it is NOT canonicalization — it belongs in ValidateStage.
//     - Runs unconditionally on both legacy PageFn and AST-compiled schemas.
//       Canonicalization is idempotent; the typed AST already emits canonical
//       output, so the pass is a no-op on clean schemas.
//
//   ValidateStage (priority 70):
//     - Enforcement only. Reads the canonicalized schema and rejects violations.
//     - All errors are *sharedErrors.BusinessError with code prefix VALIDATE_*.
//     - Errors wrap ui.ErrSchemaInvalid so errors.Is(err, ui.ErrSchemaInvalid)
//       continues to work in SchemaHandler.
//     - When DataKeyASTCompiled is true, structural invariants (syncLocation,
//       transparent background) are guaranteed by the typed AST — those rules
//       are skipped. Security rules always run regardless of path.
//
// Rule of thumb: "Can I fix this silently without guessing intent?" → Normalize.
//                "Is this a programmer error that must be rejected?"  → Validate.

// ─── VALIDATE ERROR CODES ────────────────────────────────────────────────────

const (
	// CodeValidateCRUDSyncLocation is returned when a CRUD component is missing
	// syncLocation:true. This prevents URL state loss on pagination.
	CodeValidateCRUDSyncLocation = "VALIDATE_CRUD_SYNC_LOCATION"

	// CodeValidateChartTransparentBg is returned when a chart component is
	// missing style.background:"transparent". Required for dark-mode compatibility.
	CodeValidateChartTransparentBg = "VALIDATE_CHART_TRANSPARENT_BG"

	// CodeValidateAPIMethodPrefix is returned when an api string lacks an HTTP
	// method prefix (get:/post:/put:/delete:/patch:). AMIS silently uses GET for
	// unprefixed strings — explicit prefix prevents invisible misrouting.
	CodeValidateAPIMethodPrefix = "VALIDATE_API_METHOD_PREFIX"

	// CodeValidateIAMInExpression is returned when a visibleOn/disabledOn/hiddenOn/
	// requiredOn expression contains an IAM keyword. IAM decisions belong in Go;
	// AMIS expressions execute in the browser without access to the authoritative
	// permission store.
	CodeValidateIAMInExpression = "VALIDATE_IAM_IN_EXPRESSION"
)

// ─── NORMALIZE STAGE ─────────────────────────────────────────────────────────

// NormalizeStage canonicalizes the compiled schema to a standard form.
// It MUST NEVER return an error — see the Invariant block above.
//
// Canonicalization applied:
//   - Trims leading/trailing whitespace from "api" string values.
//   - Lowercases "type" field values (AMIS type identifiers are case-sensitive
//     lower-case; a PageFn that emits "CRUD" instead of "crud" would break).
//
// NormalizeStage is Priority 60, Required true.
type NormalizeStage struct {
	pipeline.BaseStage
}

// NewNormalizeStage constructs a NormalizeStage.
func NewNormalizeStage() *NormalizeStage {
	return &NormalizeStage{
		BaseStage: pipeline.BaseStage{
			StageName:       "ui.normalize",
			StageOperations: []string{ui.OperationKey},
			StagePriority:   ui.PriorityNormalize,
			StageRequired:   true,
			StageDependsOn:  []string{"ui.compile"},
		},
	}
}

// Execute canonicalizes the schema. Always succeeds.
func (s *NormalizeStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	schema, ok := opCtx.Data[ui.DataKeySchema].(ui.Schema)
	if !ok {
		// Schema absent — nothing to canonicalize. ValidateStage will catch the missing key.
		return pipeline.StageResult{Status: "skipped", Message: "DataKeySchema missing — nothing to canonicalize"}, nil
	}

	canonicalizeSchema("$", schema)

	return pipeline.StageResult{
		Status:  "completed",
		Message: "schema canonicalized",
	}, nil
}

// canonicalizeSchema walks the schema tree and mutates nodes to canonical form.
func canonicalizeSchema(path string, node ui.M) {
	// Lowercase the "type" field — AMIS type identifiers are always lower-case.
	if t, ok := node["type"].(string); ok && t != strings.ToLower(t) {
		node["type"] = strings.ToLower(t)
	}

	// Trim whitespace from "api" string values.
	if api, ok := node["api"].(string); ok {
		trimmed := strings.TrimSpace(api)
		if trimmed != api {
			node["api"] = trimmed
		}
	}

	for k, v := range node {
		childPath := path + "." + k
		switch val := v.(type) {
		case ui.M:
			canonicalizeSchema(childPath, val)
		case []any:
			for i, item := range val {
				if m, ok := item.(ui.M); ok {
					canonicalizeSchema(fmt.Sprintf("%s[%d]", childPath, i), m)
				}
			}
		}
	}
}

// ─── VALIDATE STAGE ──────────────────────────────────────────────────────────

// ValidateStage enforces structural and security rules on the canonicalized schema.
// All violations are returned as *sharedErrors.BusinessError with code VALIDATE_*.
// Errors wrap ui.ErrSchemaInvalid so SchemaHandler can map them to HTTP 500.
//
// Rule sets:
//   - Structural rules: CRUD syncLocation, chart transparent background, API method prefix.
//     Skipped when DataKeyASTCompiled is true — the typed AST guarantees these invariants.
//   - Security rules: no IAM keyword in AMIS conditional expressions.
//     Always runs regardless of compilation path.
//
// ValidateStage is Priority 70, Required true.
type ValidateStage struct {
	pipeline.BaseStage
}

// NewValidateStage constructs a ValidateStage.
func NewValidateStage() *ValidateStage {
	return &ValidateStage{
		BaseStage: pipeline.BaseStage{
			StageName:       "ui.validate",
			StageOperations: []string{ui.OperationKey},
			StagePriority:   ui.PriorityValidate,
			StageRequired:   true,
			StageDependsOn:  []string{"ui.normalize"},
		},
	}
}

// Execute validates the schema against structural and security rules.
func (s *ValidateStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	schema, ok := opCtx.Data[ui.DataKeySchema].(ui.Schema)
	if !ok {
		return pipeline.StageResult{}, fmt.Errorf("ui.validate: DataKeySchema missing — NormalizeStage must run first")
	}

	astCompiled, _ := opCtx.Data[ui.DataKeyASTCompiled].(bool)

	// Structural rules: skipped for AST-compiled schemas (invariants guaranteed by node.Compile()).
	if !astCompiled {
		if err := walkAndValidate("$", schema, structuralRules); err != nil {
			return pipeline.StageResult{}, err
		}
	}

	// Security rules: always run.
	if err := walkAndValidate("$", schema, securityRules); err != nil {
		return pipeline.StageResult{}, err
	}

	msg := "schema passed structural and security validation"
	if astCompiled {
		msg = "schema passed security validation (AST path: structural rules skipped)"
	}
	return pipeline.StageResult{
		Status:  "completed",
		Message: msg,
	}, nil
}

// ─── Schema Walker ────────────────────────────────────────────────────────────

type validateFunc func(path string, node ui.M) error

// walkAndValidate recursively visits every M node and applies rules.
// Returns on first violation.
func walkAndValidate(path string, node ui.M, rules []validateFunc) error {
	for _, rule := range rules {
		if err := rule(path, node); err != nil {
			return err
		}
	}
	for k, v := range node {
		childPath := path + "." + k
		switch val := v.(type) {
		case ui.M:
			if err := walkAndValidate(childPath, val, rules); err != nil {
				return err
			}
		case []any:
			for i, item := range val {
				if m, ok := item.(ui.M); ok {
					if err := walkAndValidate(fmt.Sprintf("%s[%d]", childPath, i), m, rules); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// ─── Structural Rules ─────────────────────────────────────────────────────────

// structuralRules are applied to legacy PageFn-compiled schemas.
// Skipped on AST-compiled schemas — typed nodes enforce these at compile time.
var structuralRules = []validateFunc{
	ruleValidateCRUDSyncLocation,
	ruleValidateChartTransparentBg,
	ruleValidateAPIMethodPrefix,
}

func ruleValidateCRUDSyncLocation(path string, node ui.M) error {
	if node["type"] != "crud" {
		return nil
	}
	sync, ok := node["syncLocation"].(bool)
	if !ok || !sync {
		return sharedErrors.NewBusinessError(
			CodeValidateCRUDSyncLocation,
			"CRUD component missing syncLocation:true — prevents URL state loss on pagination",
		).
			WithHTTPStatus(http.StatusInternalServerError).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("path", path).
			WithDetail("fix", "use CRUDNode from the typed AST, or set syncLocation:true in the PageFn").
			WithCause(ui.ErrSchemaInvalid)
	}
	return nil
}

func ruleValidateChartTransparentBg(path string, node ui.M) error {
	if node["type"] != "chart" {
		return nil
	}
	style, _ := node["style"].(ui.M)
	if style == nil || style["background"] != "transparent" {
		return sharedErrors.NewBusinessError(
			CodeValidateChartTransparentBg,
			`chart missing style.background:"transparent" — required for dark-mode compatibility`,
		).
			WithHTTPStatus(http.StatusInternalServerError).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("path", path).
			WithDetail("fix", `set style:{"background":"transparent"} on every chart node`).
			WithCause(ui.ErrSchemaInvalid)
	}
	return nil
}

var methodPrefixes = []string{"get:", "post:", "put:", "delete:", "patch:"}

func ruleValidateAPIMethodPrefix(path string, node ui.M) error {
	apiStr, ok := node["api"].(string)
	if !ok || apiStr == "" {
		return nil
	}
	for _, p := range methodPrefixes {
		if strings.HasPrefix(apiStr, p) {
			return nil
		}
	}
	return sharedErrors.NewBusinessError(
		CodeValidateAPIMethodPrefix,
		fmt.Sprintf("api %q missing HTTP method prefix — AMIS silently uses GET for unprefixed strings", apiStr),
	).
		WithHTTPStatus(http.StatusInternalServerError).
		WithCategory(sharedErrors.CategoryValidation).
		WithDetail("path", path).
		WithDetail("api", apiStr).
		WithDetail("fix", "prefix with get:/post:/put:/delete:/patch:").
		WithCause(ui.ErrSchemaInvalid)
}

// ─── Security Rules ───────────────────────────────────────────────────────────

// securityRules always run regardless of compilation path.
var securityRules = []validateFunc{
	ruleNoIAMExpressions,
}

// iamKeywords are banned inside AMIS conditional expressions.
// They indicate IAM data being embedded in client-side logic.
var iamKeywords = []string{
	"role:", "role ==", "role===", ".roles", ".role ",
	"permission:", ".permissions", "Can(", "CanDo(",
	"user.type", "UserType", "ADMIN", "SYSADMIN",
}

func ruleNoIAMExpressions(path string, node ui.M) error {
	for _, key := range []string{"visibleOn", "disabledOn", "hiddenOn", "requiredOn"} {
		expr, ok := node[key].(string)
		if !ok || expr == "" {
			continue
		}
		for _, kw := range iamKeywords {
			if strings.Contains(expr, kw) {
				return sharedErrors.NewBusinessError(
					CodeValidateIAMInExpression,
					fmt.Sprintf("expression in %q contains IAM keyword %q", key, kw),
				).
					WithHTTPStatus(http.StatusInternalServerError).
					WithCategory(sharedErrors.CategorySecurity).
					WithDetail("path", path+"."+key).
					WithDetail("expression", expr).
					WithDetail("keyword", kw).
					WithDetail("fix", "use a boolean data variable set by Go — e.g. ${can_approve_invoice}").
					WithCause(ui.ErrSchemaInvalid)
			}
		}
	}
	return nil
}

var _ pipeline.Stage = (*NormalizeStage)(nil)
var _ pipeline.Stage = (*ValidateStage)(nil)
