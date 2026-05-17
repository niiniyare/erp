package iam

// AUTHZ-3 — Regression tests: No Authorization Bypass Paths
//
// Static-analysis guards. Scan source files to ensure no enforcement bypass
// patterns have crept in. Pass when codebase is clean; catch regressions early.
//
// Rules (docs/reference/modules/iam/tasks.md AUTHZ-3):
//   1. No handler may check authorization via direct DB role query.
//   2. No handler may inspect session.UserType to bypass authorization.
//   3. No service may make authorization decisions based on role strings.
//   4. UserService must not make authorization decisions; identity only.

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// repoRoot walks up from cwd until it finds go.mod.
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
			t.Fatal("repo root not found (no go.mod)")
		}
		dir = parent
	}
}

// sourceFiles returns all non-test .go files under root, skipping dirs
// whose path contains any of the skip substrings.
func sourceFiles(t *testing.T, root string, skipDirSubstr ...string) []string {
	t.Helper()
	var out []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			for _, s := range skipDirSubstr {
				if strings.Contains(rel, s) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			out = append(out, path)
		}
		return nil
	})
	return out
}

// grepFiles returns "file:line: content" strings where line contains pattern.
func grepFiles(files []string, pattern string) []string {
	var hits []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, pattern) {
				hits = append(hits, fmt.Sprintf("%s:%d: %s", f, i+1, strings.TrimSpace(line)))
			}
		}
	}
	return hits
}

// TestAUTHZ3_NoUserTypeAuthzBypassInHandlers — Rule 2
//
// Handler files must not compare session.UserType to make access decisions.
// IsPlatform()/IsPortal() are allowed (context helpers, not authz gates).
// Raw UserType string comparisons are the bypass pattern.
func TestAUTHZ3_NoUserTypeAuthzBypassInHandlers(t *testing.T) {
	root := repoRoot(t)
	handlerDir := filepath.Join(root, "internal", "api", "handlers")
	if _, err := os.Stat(handlerDir); os.IsNotExist(err) {
		t.Skip("no handler directory found")
	}

	files := sourceFiles(t, handlerDir)
	bypassPatterns := []string{
		`UserType == "PLATFORM"`,
		`UserType == "TENANT"`,
		`UserType == "PORTAL"`,
		`UserType == "API"`,
		`UserType != "PLATFORM"`,
		`UserType != "TENANT"`,
	}

	for _, pat := range bypassPatterns {
		hits := grepFiles(files, pat)
		assert.Empty(t, hits,
			"AUTHZ-3: handler uses UserType for access control (bypass detected).\n"+
				"Pattern: %q\nAll authz must go through Authorize() middleware → Enforce().", pat)
	}
}

// TestAUTHZ3_NoDirectRoleAssignmentsQueryInHandlers — Rule 1
//
// Handlers must not query the role_assignments table directly.
// Only repository files may issue SQL against role_assignments.
func TestAUTHZ3_NoDirectRoleAssignmentsQueryInHandlers(t *testing.T) {
	root := repoRoot(t)
	dirs := []string{
		filepath.Join(root, "internal", "api", "handlers"),
		filepath.Join(root, "internal", "api", "middleware"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		files := sourceFiles(t, dir, "repository")
		hits := grepFiles(files, "role_assignments")
		assert.Empty(t, hits,
			"AUTHZ-3: handler/middleware references role_assignments directly.\n"+
				"Authz must use AuthzService.Enforce(), not raw DB role queries.")
	}
}

// TestAUTHZ3_NoHasRoleInHandlers — Rule 3
//
// Handlers must not call HasRole() to gate access. HasRole is audit/display
// only. Access gates belong in middleware via Enforce().
func TestAUTHZ3_NoHasRoleInHandlers(t *testing.T) {
	root := repoRoot(t)
	handlerDir := filepath.Join(root, "internal", "api", "handlers")
	if _, err := os.Stat(handlerDir); os.IsNotExist(err) {
		t.Skip("no handler directory found")
	}

	files := sourceFiles(t, handlerDir)
	hits := grepFiles(files, "HasRole(")
	assert.Empty(t, hits,
		"AUTHZ-3: handler calls HasRole() for access gating.\n"+
			"Use Authorize() middleware. HasRole is for audit/display only.")
}

// TestAUTHZ3_UserServiceHasNoEnforceMethod — Rule 4
//
// UserService interface must not carry authorization methods (Enforce,
// GetPolicies, AddPolicy, etc.). Authorization belongs exclusively to
// AuthzService. This test uses reflection to prove the separation.
func TestAUTHZ3_UserServiceHasNoEnforceMethod(t *testing.T) {
	t.Parallel()

	userSvcType := reflect.TypeOf((*UserService)(nil)).Elem()
	authzOnlyMethods := []string{"Enforce", "EnforceBatch", "AddPolicy", "RemovePolicy", "GetPolicies"}

	for _, method := range authzOnlyMethods {
		_, found := userSvcType.MethodByName(method)
		assert.False(t, found,
			"AUTHZ-3: UserService has method %q — authorization methods belong to AuthzService only.\n"+
				"UserService must be identity-only (CRUD, authn, password, MFA).", method)
	}
}
