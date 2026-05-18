package guard_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks upward from this file's location until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod — is this a Go module?")
		}
		dir = parent
	}
}

// walkGoFiles calls fn for every non-test .go file under root that is not
// inside any of the excluded subtrees.
func walkGoFiles(t *testing.T, root string, excluded []string, fn func(path string)) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			for _, ex := range excluded {
				if strings.HasPrefix(path, ex) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fn(path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

// readFile returns file content as a string; fails the test on error.
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// rel returns path relative to root for readable test output.
func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return r
}

// ----------------------------------------------------------------------------
// GUARD-1: no direct "awo.so/internal/core/iam" import outside iam/ itself
// ----------------------------------------------------------------------------

func TestGuard_NoDirectIAMImportOutsideIAM(t *testing.T) {
	root := repoRoot(t)
	iamDir := filepath.Join(root, "internal", "core", "iam")
	guardDir := filepath.Join(root, "internal", "core", "iam", "guard")
	excluded := []string{iamDir}

	var violations []string

	walkGoFiles(t, filepath.Join(root, "internal"), excluded, func(path string) {
		content := readFile(t, path)
		// Direct iam package import (not sub-packages like iam/contract)
		if strings.Contains(content, `"awo.so/internal/core/iam"`) {
			violations = append(violations, rel(root, path))
		}
	})

	// guard/ itself is inside iam/ so excluded — but double-check by scanning
	// guard test files too (should never import iam directly)
	walkGoFiles(t, guardDir, nil, func(path string) {
		content := readFile(t, path)
		if strings.Contains(content, `"awo.so/internal/core/iam"`) {
			violations = append(violations, rel(root, path)+" [guard itself]")
		}
	})

	if len(violations) > 0 {
		t.Errorf("GUARD-1 FAIL: direct import of awo.so/internal/core/iam is forbidden outside internal/core/iam/\n"+
			"Offending files (%d):\n  %s\n\n"+
			"Fix: import awo.so/internal/core/iam/contract instead.",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// ----------------------------------------------------------------------------
// GUARD-2: no import of iam/service or iam/repository outside iam/ itself
// ----------------------------------------------------------------------------

func TestGuard_NoIAMInternalPackageImports(t *testing.T) {
	root := repoRoot(t)
	iamDir := filepath.Join(root, "internal", "core", "iam")
	excluded := []string{iamDir}

	forbidden := []string{
		`"awo.so/internal/core/iam/service"`,
		`"awo.so/internal/core/iam/repository"`,
	}

	var violations []string

	walkGoFiles(t, filepath.Join(root, "internal"), excluded, func(path string) {
		content := readFile(t, path)
		for _, pkg := range forbidden {
			if strings.Contains(content, pkg) {
				violations = append(violations, rel(root, path)+": imports "+pkg)
			}
		}
	})

	if len(violations) > 0 {
		t.Errorf("GUARD-2 FAIL: forbidden IAM-internal package imports found (%d):\n  %s",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// ----------------------------------------------------------------------------
// GUARD-3: no iam.Service field declaration outside iam/
// ----------------------------------------------------------------------------

func TestGuard_NoIAMServiceFieldOutsideIAM(t *testing.T) {
	root := repoRoot(t)
	iamDir := filepath.Join(root, "internal", "core", "iam")
	excluded := []string{iamDir}

	forbidden := []string{
		"iam.Service",
		"iam.SessionService",
		"iam.AuthzService",
	}

	var violations []string

	walkGoFiles(t, filepath.Join(root, "internal"), excluded, func(path string) {
		content := readFile(t, path)
		for _, pattern := range forbidden {
			if strings.Contains(content, pattern) {
				violations = append(violations, rel(root, path)+": contains "+pattern)
			}
		}
	})

	if len(violations) > 0 {
		t.Errorf("GUARD-3 FAIL: iam.Service / iam.SessionService field types found outside iam/ (%d):\n  %s\n\n"+
			"Fix: remove the field; inject via contract.AuthService or read from contract.FromContext.",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// ----------------------------------------------------------------------------
// GUARD-4: no live .Enforce( calls outside iam/ and middleware/
// (comment-only occurrences are acceptable as documentation)
// ----------------------------------------------------------------------------

func TestGuard_NoDirectEnforceCallsOutsideIAMAndMiddleware(t *testing.T) {
	root := repoRoot(t)
	iamDir := filepath.Join(root, "internal", "core", "iam")
	middlewareDir := filepath.Join(root, "internal", "api", "middleware")
	excluded := []string{iamDir, middlewareDir}

	var violations []string

	walkGoFiles(t, filepath.Join(root, "internal"), excluded, func(path string) {
		fset := token.NewFileSet()
		src, err := os.ReadFile(path)
		if err != nil {
			return
		}
		f, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			return // skip unparseable files
		}

		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if sel.Sel.Name == "Enforce" {
				pos := fset.Position(call.Pos())
				violations = append(violations,
					rel(root, path)+":"+pos.String())
			}
			return true
		})
	})

	if len(violations) > 0 {
		t.Errorf("GUARD-4 FAIL: live .Enforce( calls found outside iam/ and middleware/ (%d):\n  %s\n\n"+
			"Fix: remove service-layer Enforce calls; authorization is enforced by middleware.Authorize at the route.",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// ----------------------------------------------------------------------------
// GUARD-5: no planned Enforce bypass TODO comments in non-IAM source
// ----------------------------------------------------------------------------

func TestGuard_NoPlannedEnforceTODOComments(t *testing.T) {
	root := repoRoot(t)
	iamDir := filepath.Join(root, "internal", "core", "iam")
	excluded := []string{iamDir}

	// Patterns that indicate a planned direct-Enforce bypass documented as TODO
	dangerPatterns := []string{
		"TODO(authz)",
		"TODO: enforce",
		"TODO: call Enforce",
		"iamService.Enforce",
		"s.iamService.Enforce",
	}

	var violations []string

	walkGoFiles(t, filepath.Join(root, "internal"), excluded, func(path string) {
		content := readFile(t, path)
		for _, pattern := range dangerPatterns {
			if strings.Contains(content, pattern) {
				violations = append(violations, rel(root, path)+": contains \""+pattern+"\"")
			}
		}
	})

	if len(violations) > 0 {
		t.Errorf("GUARD-5 FAIL: planned Enforce bypass TODO found outside iam/ (%d):\n  %s\n\n"+
			"Fix: delete the TODO comment. Authorization belongs in middleware.Authorize, not service methods.",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// ----------------------------------------------------------------------------
// GUARD-6: CapabilityContext must not have Permissions / Features / Modules
// ----------------------------------------------------------------------------

func TestGuard_CapabilityContextHasNoPermissionFields(t *testing.T) {
	root := repoRoot(t)
	sharedCtx := filepath.Join(root, "internal", "shared", "context.go")

	content := readFile(t, sharedCtx)

	forbidden := []string{
		"Permissions map[string]bool",
		"Features    map[string]bool",
		"Features map[string]bool",
		"Modules     map[string]bool",
		"Modules map[string]bool",
	}

	var violations []string
	for _, f := range forbidden {
		if strings.Contains(content, f) {
			violations = append(violations, f)
		}
	}

	if len(violations) > 0 {
		t.Errorf("GUARD-6 FAIL: shadow permission fields found in shared/context.go CapabilityContext:\n  %s\n\n"+
			"Fix: remove these fields; feature flags must come from contract.SessionContext.FeatureEnabled, "+
			"authorization from middleware.Authorize.",
			strings.Join(violations, "\n  "))
	}
}

// ----------------------------------------------------------------------------
// GUARD-7: featureflag.Service must not appear as a struct field outside iam/
// (feature flags are read from SessionContext, not injected as a service)
// ----------------------------------------------------------------------------

func TestGuard_NoFeatureFlagServiceFieldInDomainServices(t *testing.T) {
	root := repoRoot(t)
	iamDir := filepath.Join(root, "internal", "core", "iam")
	// featureflag package itself and its tests are excluded
	featureflagDir := filepath.Join(root, "internal", "core", "featureflag")
	excluded := []string{iamDir, featureflagDir}

	var violations []string

	walkGoFiles(t, filepath.Join(root, "internal", "core"), excluded, func(path string) {
		content := readFile(t, path)
		if strings.Contains(content, "featureflag.Service") {
			violations = append(violations, rel(root, path))
		}
	})

	if len(violations) > 0 {
		t.Errorf("GUARD-7 FAIL: featureflag.Service injected into domain service (%d files):\n  %s\n\n"+
			"Fix: read feature flags via contract.FromContext(ctx).FeatureEnabled(key) instead.",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// ----------------------------------------------------------------------------
// GUARD-8: raw string context key reads for IAM data are forbidden
// (must use shared typed keys or contract.FromContext — not ctx.Value("tenant_id"))
// ----------------------------------------------------------------------------

func TestGuard_NoRawStringContextKeyReads(t *testing.T) {
	root := repoRoot(t)
	iamDir := filepath.Join(root, "internal", "core", "iam")
	sharedDir := filepath.Join(root, "internal", "shared") // shared defines the keys — allowed
	excluded := []string{iamDir, sharedDir}

	forbidden := []string{
		`ctx.Value("tenant_id")`,
		`ctx.Value("user_id")`,
		`ctx.Value("entity_id")`,
		`ctx.Value("session_id")`,
	}

	var violations []string

	walkGoFiles(t, filepath.Join(root, "internal"), excluded, func(path string) {
		content := readFile(t, path)
		for _, pattern := range forbidden {
			if strings.Contains(content, pattern) {
				violations = append(violations, rel(root, path)+": "+pattern)
			}
		}
	})

	if len(violations) > 0 {
		t.Errorf("GUARD-8 FAIL: raw string context key reads found (%d):\n  %s\n\n"+
			"Fix: use shared.GetTenantID(ctx) / contract.FromContext(ctx).TenantID() instead.",
			len(violations), strings.Join(violations, "\n  "))
	}
}
