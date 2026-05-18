package guard_test

// runtime_invariants_test.go — AUTHZ-RUNTIME-1/2/3
//
// Regression guards: session types must NEVER become authorization authorities.
//
// These tests scan source text so they run without any database or runtime
// dependency — same pattern as contract_boundary_test.go.
//
// AUTHZ-RUNTIME-1: SessionContext and ResolvedSession have no authorization
//                  methods (Can, CanDo, HasPermission, CanAccess, Permit, …).
// AUTHZ-RUNTIME-2: Session serialization structs have no JSON "permissions" key.
// AUTHZ-RUNTIME-3: No exported session type anywhere in iam/ or contract/
//                  carries a struct field whose name or JSON tag implies
//                  authorization data (Permissions, Grants, Can, Allow, Deny).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// AUTHZ-RUNTIME-1: no authorization method declarations on session types
// ---------------------------------------------------------------------------

func TestAUTHZ_RUNTIME_1_NoAuthzMethodsOnSessionTypes(t *testing.T) {
	root := repoRoot(t)

	// Methods that indicate a session type is making authorization decisions.
	// Any method with these names on a session type is a violation.
	forbiddenMethodNames := []string{
		"Can",
		"CanDo",
		"CanAccess",
		"HasPermission",
		"IsAllowed",
		"IsPermitted",
		"Permit",
		"Allow",
		"Deny",
		"Authorize",
		"Check",
	}

	// Receiver type names considered "session types".
	sessionReceiverTypes := []string{
		"SessionContext",
		"ResolvedSession",
		"Session",
	}

	var violations []string

	// Scan iam/ and contract/ for method declarations.
	scanDirs := []string{
		filepath.Join(root, "internal", "core", "iam"),
		filepath.Join(root, "internal", "core", "iam", "contract"),
		filepath.Join(root, "internal", "core", "iam", "domain"),
	}

	for _, dir := range scanDirs {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if err != nil {
			// Directory may not exist — skip gracefully.
			continue
		}
		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				ast.Inspect(file, func(n ast.Node) bool {
					fn, ok := n.(*ast.FuncDecl)
					if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
						return true
					}

					// Determine receiver type name.
					recv := fn.Recv.List[0].Type
					var recvName string
					switch r := recv.(type) {
					case *ast.Ident:
						recvName = r.Name
					case *ast.StarExpr:
						if id, ok := r.X.(*ast.Ident); ok {
							recvName = id.Name
						}
					}

					// Check if this is a session receiver with a forbidden method.
					for _, st := range sessionReceiverTypes {
						if recvName != st {
							continue
						}
						for _, fm := range forbiddenMethodNames {
							if fn.Name.Name == fm {
								pos := fset.Position(fn.Pos())
								violations = append(violations,
									rel(root, pos.Filename)+":"+string(rune('0'+pos.Line/100%10))+
										" method "+recvName+"."+fn.Name.Name)
							}
						}
					}
					return true
				})
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("AUTHZ-RUNTIME-1 FAIL: authorization methods found on session types (%d):\n  %s\n\n"+
			"Fix: session types are identity carriers, not authorization authorities.\n"+
			"Authorization must go through middleware.Authorize → authzService.Enforce → Casbin.",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// ---------------------------------------------------------------------------
// AUTHZ-RUNTIME-2: session serialization has no "permissions" JSON key
// ---------------------------------------------------------------------------

func TestAUTHZ_RUNTIME_2_SessionSerializationHasNoPermissionsKey(t *testing.T) {
	root := repoRoot(t)

	// JSON tag values that imply a cached permissions payload.
	// We look for struct tags with these exact JSON key names.
	forbiddenJSONKeys := []string{
		`"permissions"`,
		`"grants"`,
		`"allow_list"`,
		`"deny_list"`,
		`"acl"`,
		`"capabilities"`,
		`"access_rights"`,
		`"can_"`,   // prefix: can_read, can_write, etc.
		`"perm_"`,  // prefix: perm_finance, etc.
	}

	// Files that define the session serialization structs.
	// We scan the entire iam domain and session files.
	scanDirs := []string{
		filepath.Join(root, "internal", "core", "iam"),
		filepath.Join(root, "internal", "core", "iam", "domain"),
		filepath.Join(root, "internal", "core", "iam", "contract"),
	}

	var violations []string

	for _, dir := range scanDirs {
		walkGoFiles(t, dir, nil, func(path string) {
			if strings.HasSuffix(path, "_test.go") {
				return
			}
			content := readFile(t, path)
			for _, key := range forbiddenJSONKeys {
				if strings.Contains(content, `json:`+key) || strings.Contains(content, `json: `+key) {
					violations = append(violations,
						rel(root, path)+": contains json tag "+key)
				}
			}
		})
	}

	if len(violations) > 0 {
		t.Errorf("AUTHZ-RUNTIME-2 FAIL: session serialization contains permissions JSON keys (%d):\n  %s\n\n"+
			"Fix: session JSON must carry identity and context only.\n"+
			"Authorization state must never be cached in the session payload.",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// ---------------------------------------------------------------------------
// AUTHZ-RUNTIME-3: no exported session type has a permissions-implying field
// ---------------------------------------------------------------------------

func TestAUTHZ_RUNTIME_3_NoPermissionsFieldInSessionTypes(t *testing.T) {
	root := repoRoot(t)

	// Field name prefixes/patterns that imply cached authorization state.
	// Checked case-insensitively on the field name.
	forbiddenFieldPatterns := []string{
		"permission",  // Permissions, PermissionMap, PermissionSet
		"grant",       // Grants, GrantedRoles
		"acl",         // ACL, ACLMap
		"allowlist",   // AllowList
		"denylist",    // DenyList
		"canread",     // CanRead
		"canwrite",    // CanWrite
		"candelete",   // CanDelete
		"canadmin",    // CanAdmin
		"capability",  // Capabilities
		"accessrights", // AccessRights
	}

	scanDirs := []string{
		filepath.Join(root, "internal", "core", "iam"),
		filepath.Join(root, "internal", "core", "iam", "domain"),
		filepath.Join(root, "internal", "core", "iam", "contract"),
	}

	var violations []string

	for _, dir := range scanDirs {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if err != nil {
			continue
		}
		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				ast.Inspect(file, func(n ast.Node) bool {
					st, ok := n.(*ast.StructType)
					if !ok {
						return true
					}
					for _, field := range st.Fields.List {
						for _, name := range field.Names {
							if !name.IsExported() {
								continue
							}
							lower := strings.ToLower(name.Name)
							for _, pat := range forbiddenFieldPatterns {
								if strings.Contains(lower, pat) {
									pos := fset.Position(field.Pos())
									violations = append(violations,
										rel(root, pos.Filename)+": exported field "+name.Name+
											" matches pattern \""+pat+"\"")
								}
							}
						}
					}
					return true
				})
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("AUTHZ-RUNTIME-3 FAIL: permissions-implying field found in IAM/contract structs (%d):\n  %s\n\n"+
			"Fix: session types must carry identity and context only.\n"+
			"Remove field and route all authorization through middleware.Authorize → Casbin.",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// ---------------------------------------------------------------------------
// AUTHZ-RUNTIME-4: no session type caches a Roles slice for authorization
// ---------------------------------------------------------------------------
//
// Carrying a Roles []string in the session is a common anti-pattern that leads
// to handler-level role checks (if contains(sess.Roles, "admin")) bypassing
// Casbin. The session already embeds EntityScope and Configuration for the
// identity / context questions — explicit roles are never needed by consumers.

func TestAUTHZ_RUNTIME_4_NoRolesFieldInSessionTypes(t *testing.T) {
	root := repoRoot(t)

	// Only flag fields whose name is literally "Roles" or "Role" as exported
	// fields on the known session structs. We use AST so we only flag struct
	// fields, not local variables or function parameters.
	sessionStructNames := map[string]bool{
		"Session":         true,
		"ResolvedSession": true,
		"SessionContext":  true,
	}

	var violations []string

	scanDirs := []string{
		filepath.Join(root, "internal", "core", "iam"),
		filepath.Join(root, "internal", "core", "iam", "domain"),
		filepath.Join(root, "internal", "core", "iam", "contract"),
	}

	for _, dir := range scanDirs {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if err != nil {
			continue
		}

		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				// Walk top-level type declarations to identify session struct names.
				for _, decl := range file.Decls {
					genDecl, ok := decl.(*ast.GenDecl)
					if !ok {
						continue
					}
					for _, spec := range genDecl.Specs {
						typeSpec, ok := spec.(*ast.TypeSpec)
						if !ok || !sessionStructNames[typeSpec.Name.Name] {
							continue
						}
						structType, ok := typeSpec.Type.(*ast.StructType)
						if !ok {
							continue
						}
						for _, field := range structType.Fields.List {
							for _, name := range field.Names {
								lower := strings.ToLower(name.Name)
								if lower == "roles" || lower == "role" {
									pos := fset.Position(field.Pos())
									violations = append(violations,
										rel(root, pos.Filename)+": "+typeSpec.Name.Name+
											"."+name.Name+" caches roles in session")
								}
							}
						}
					}
				}
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("AUTHZ-RUNTIME-4 FAIL: Roles field found on session type (%d):\n  %s\n\n"+
			"Fix: remove Roles from session. Role lookups must happen through\n"+
			"authzService.GetRoles() or via Casbin g-rules — never from a cached slice.",
			len(violations), strings.Join(violations, "\n  "))
	}
}
