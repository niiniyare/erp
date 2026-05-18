package stages

import (
	"fmt"
	"strings"

	"awo.so/internal/pipeline"
	"awo.so/internal/web/ui"
)

// ─── TASK 6 — NORMALIZE STAGE ────────────────────────────────────────────────
//
// NormalizeStage enforces AMIS structural compliance rules on the compiled schema.
// It runs after CompileStage and before ValidateStage.
//
// DESCRIPTION:
// Reads DataKeySchema. Walks the schema tree. Enforces:
//   1. CRUD components must have syncLocation: true
//   2. Chart components must have style.background: "transparent"
//   3. All API strings must carry an HTTP method prefix (get:/post:/put:/delete:/patch:)
//
// Why these rules:
//   - syncLocation: prevents URL state loss on CRUD pagination.
//   - transparent chart bg: required for dark-mode compatibility.
//   - method prefix: AMIS sends GET for unprefixed APIs; explicit prefix prevents silent mistakes.
//
// RISKS:
// Required: a non-compliant schema aborts the pipeline. This is intentional —
// catching structural bugs at request time is better than shipping broken UI.
// In CI, NormalizeStage can be called directly against test schemas to catch
// issues before deployment.

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
		},
	}
}

// Execute validates AMIS compliance. Returns SchemaValidationError on violation.
//
// When DataKeyASTCompiled is true, structural invariants (syncLocation,
// transparent background) are guaranteed by the typed AST — those rules are
// skipped to avoid redundant checks. Security rules (no IAM in expressions)
// always run regardless of compilation path.
func (s *NormalizeStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	schema, ok := opCtx.Data[ui.DataKeySchema].(ui.Schema)
	if !ok {
		return pipeline.StageResult{}, fmt.Errorf("ui.normalize: DataKeySchema missing — CompileStage must run first")
	}

	astCompiled, _ := opCtx.Data[ui.DataKeyASTCompiled].(bool)

	rules := normalizeRules
	if astCompiled {
		// Structural invariants guaranteed by typed AST nodes — skip redundant checks.
		rules = normalizeRulesLegacyOnly
	}

	if err := walkSchema("$", schema, rules); err != nil {
		return pipeline.StageResult{}, err
	}

	msg := "schema passed AMIS compliance checks"
	if astCompiled {
		msg = "schema passed AMIS compliance checks (AST path: structural rules skipped)"
	}
	return pipeline.StageResult{
		Status:  "completed",
		Message: msg,
	}, nil
}

// ─── TASK 7 — VALIDATE STAGE ─────────────────────────────────────────────────
//
// ValidateStage enforces security rules: no IAM expressions in AMIS conditionals.
//
// DESCRIPTION:
// Reads DataKeySchema. Walks schema looking for visibleOn/disabledOn/hiddenOn
// expressions containing IAM keywords (role:, permission:, .roles, .permissions).
//
// WHY:
// AMIS expressions execute in the browser. If a PageFn embeds raw permission
// data in expressions (e.g. "${user.role === 'ADMIN'}"), it leaks IAM structure
// to the client and creates a second unauthorized decision point.
// The contract: Go decides what to show. AMIS renders it. Never the reverse.
//
// ALLOWED:   "${can_approve_invoice}"  — boolean pre-set by Go
// FORBIDDEN: "${user.role === 'ADMIN'}" — role check in browser
//
// RISKS:
// Required: a forbidden expression aborts the pipeline.

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
		},
	}
}

// Execute scans for forbidden IAM expressions in schema conditionals.
func (s *ValidateStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	schema, ok := opCtx.Data[ui.DataKeySchema].(ui.Schema)
	if !ok {
		return pipeline.StageResult{}, fmt.Errorf("ui.validate: DataKeySchema missing")
	}

	if err := walkSchema("$", schema, securityRules); err != nil {
		return pipeline.StageResult{}, err
	}

	return pipeline.StageResult{
		Status:  "completed",
		Message: "schema passed security expression checks",
	}, nil
}

// ─── Schema Walker ────────────────────────────────────────────────────────────

type ruleFunc func(path string, node ui.M) error

// walkSchema recursively visits every M node in the schema tree and applies rules.
func walkSchema(path string, node ui.M, rules []ruleFunc) error {
	for _, rule := range rules {
		if err := rule(path, node); err != nil {
			return err
		}
	}
	for k, v := range node {
		childPath := path + "." + k
		switch val := v.(type) {
		case ui.M:
			if err := walkSchema(childPath, val, rules); err != nil {
				return err
			}
		case []any:
			for i, item := range val {
				if m, ok := item.(ui.M); ok {
					if err := walkSchema(fmt.Sprintf("%s[%d]", childPath, i), m, rules); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// ─── Normalize Rules ──────────────────────────────────────────────────────────

// normalizeRules is the full rule set applied to legacy PageFn-compiled schemas.
var normalizeRules = []ruleFunc{
	ruleCRUDSyncLocation,
	ruleChartTransparentBg,
	ruleAPIMethodPrefix,
}

// normalizeRulesLegacyOnly is the reduced rule set for AST-compiled schemas.
// Structural invariants (syncLocation, transparent bg) are guaranteed by node
// Compile() — only the API method prefix check still applies because raw API
// strings could theoretically appear inside custom map[string]any values passed
// through legacy blocks embedded in an otherwise AST-compiled page.
var normalizeRulesLegacyOnly = []ruleFunc{
	ruleAPIMethodPrefix,
}

func ruleCRUDSyncLocation(path string, node ui.M) error {
	if node["type"] != "crud" {
		return nil
	}
	sync, ok := node["syncLocation"].(bool)
	if !ok || !sync {
		return &ui.SchemaValidationError{
			Rule:    "crud-sync-location",
			Path:    path,
			Message: "CRUD component missing syncLocation:true — add it via amis.CRUD() builder",
		}
	}
	return nil
}

func ruleChartTransparentBg(path string, node ui.M) error {
	if node["type"] != "chart" {
		return nil
	}
	style, _ := node["style"].(ui.M)
	if style == nil || style["background"] != "transparent" {
		return &ui.SchemaValidationError{
			Rule:    "chart-transparent-bg",
			Path:    path,
			Message: "chart missing style.background:\"transparent\" — required for dark mode compatibility",
		}
	}
	return nil
}

var methodPrefixes = []string{"get:", "post:", "put:", "delete:", "patch:"}

func ruleAPIMethodPrefix(path string, node ui.M) error {
	apiStr, ok := node["api"].(string)
	if !ok || apiStr == "" {
		return nil
	}
	for _, p := range methodPrefixes {
		if strings.HasPrefix(apiStr, p) {
			return nil
		}
	}
	return &ui.SchemaValidationError{
		Rule:    "api-method-prefix",
		Path:    path,
		Message: fmt.Sprintf("api %q missing method prefix — use get:/post:/put:/delete:/patch:", apiStr),
	}
}

// ─── Security Rules ───────────────────────────────────────────────────────────

var securityRules = []ruleFunc{
	ruleNoIAMExpressions,
}

// iamKeywords are banned inside AMIS conditional expressions.
// They indicate that IAM data is being embedded in client-side logic.
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
				return &ui.SchemaValidationError{
					Rule: "no-iam-in-expressions",
					Path: path + "." + key,
					Message: fmt.Sprintf(
						"expression %q contains IAM keyword %q — "+
							"use a boolean data variable set by Go (e.g. ${can_approve_invoice}) instead",
						expr, kw,
					),
				}
			}
		}
	}
	return nil
}

var _ pipeline.Stage = (*NormalizeStage)(nil)
var _ pipeline.Stage = (*ValidateStage)(nil)
