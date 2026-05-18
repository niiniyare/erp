# Contributing to the UI System

This guide covers how to add pages, blocks, and stages to the AWO ERP UI
pipeline. Read `architecture.md` first.

---

## Adding a New Page

### Step 1 — Register the page

In `internal/web/registry/registry.go` (or a package `init()` function):

```go
func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/my-module/my-page",
        Module: "my-module",
        Title:  "My Page",
        ASTFn: func(sess ui.UISessionContext) any {
            return screens.MyPageScreen(sess)
        },
    })
}
```

`ASTPageFn` is preferred for all new pages. Use `Fn` (legacy `PageFn`) only
when migrating an existing page that hasn't been converted yet.

`ValidateRegistry()` (called at startup in `NewUIPipeline`) panics if Module
or Title is missing, or if both `Fn` and `ASTFn` are nil.

### Step 2 — Write the screen

Create `internal/web/dsl/screens/my_page.go` (≤ 60 lines):

```go
package screens

import (
    "awo.so/internal/web/ast"
    "awo.so/internal/web/dsl/blocks"
    "awo.so/internal/web/ui"
)

func MyPageScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "My Page",
        Body: []ast.Node{
            blocks.FilterBarBlock(sess, blocks.FilterBarConfig{ShowSearch: true}),
            blocks.DataTableBlock(sess, blocks.DataTableConfig{
                APIURL:  "/api/v1/my-module/items",
                Columns: myPageColumns(),
            }),
        },
    }
}

func myPageColumns() []blocks.ColumnDef {
    return []blocks.ColumnDef{
        {Name: "name", Label: "Name", Sortable: true},
        {Name: "status", Label: "Status", Type: "status"},
        {Name: "created_at", Label: "Created", Type: "date", Sortable: true},
    }
}
```

### Step 3 — Run the architecture guard

```bash
bash scripts/check-arch.sh
```

Fix any violations before pushing.

---

## Adding a New Block

New blocks go in `internal/web/dsl/blocks/`. Choose the correct family file or
create a new one if the concern is genuinely new.

```go
// internal/web/dsl/blocks/my_block.go
package blocks

import (
    "awo.so/internal/web/ast"
    "awo.so/internal/web/ui"
)

// MyBlockConfig configures MyBlock.
type MyBlockConfig struct {
    Title    string
    ReadOnly bool
}

// MyBlock renders …
// Callers must not gate this call behind permission checks — the block
// handles visibility internally.
func MyBlock(sess ui.UISessionContext, cfg MyBlockConfig) ast.Node {
    // permission check here if needed:
    // if !sess.Can("read", "my-module.things") { return emptyNode() }
    return ast.CardNode{
        Title: cfg.Title,
        Body:  []ast.Node{ /* … */ },
    }
}
```

Rules:
- Return `ast.Node`, never `map[string]any`
- Accept `ui.UISessionContext` as the first argument
- Permission checks inside the block, not in callers
- Errors via `sharedErrors.BusinessError` with `DSL_*` code (panic for
  programmer errors like missing required config)

---

## Adding a New Stage

```go
// internal/web/stages/my_stage.go
package stages

import (
    "awo.so/internal/pipeline"
    "awo.so/internal/web/ui"
)

type MyStage struct {
    pipeline.BaseStage
    // injected dependencies here
}

func NewMyStage() *MyStage {
    return &MyStage{
        BaseStage: pipeline.BaseStage{
            StageName:       "ui.my_stage",
            StageOperations: []string{ui.OperationKey},
            StagePriority:   55, // between Compile(50) and Normalize(60)
            StageRequired:   true,
            StageDependsOn:  []string{"ui.compile"}, // required — never omit
        },
    }
}

func (s *MyStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    // read from opCtx.Data
    // write to StageResult.Outputs
    return pipeline.StageResult{Status: "completed"}, nil
}

var _ pipeline.Stage = (*MyStage)(nil)
```

Then register in `internal/web/wire.go`:

```go
reg.Register(
    // …existing stages…
    instrument(stages.NewMyStage()),
)
```

And update `DependsOn` of any stage that reads `MyStage`'s outputs.

---

## Architecture Guards Reference

Run `bash scripts/check-arch.sh` before every push. Eight guards:

| Guard | What it checks |
|-------|---------------|
| 1 | No `map[string]any` in `internal/web/` outside `ast/` and `ui/` |
| 2 | No IAM imports in `dsl/` |
| 3 | No permission checks in `VisibleOn` expressions |
| 4 | No `PageFn` / `ASTPageFn` called outside `stages/` and `registry/` |
| 5 | Every non-root stage declares `StageDependsOn` |
| 6 | No `sync.Map` or third-party in-process caches in `internal/web/` |
| 7 | No raw block-level AST nodes constructed directly in `screens/` |
| 8 | Every `screens/*.go` file is under 60 lines |

All guards run as the first CI job and block build + test if any fail.

---

## Page Migration (PageFn → ASTPageFn)

Existing pages registered with `Fn` (legacy `PageFn`) continue to work during
the migration window. To migrate:

1. Write an `ASTFn` that returns an `ast.Node` (usually a `screens/` function).
2. Update the `RegisterPage` call to set `ASTFn` instead of `Fn`.
3. Delete the old `PageFn` implementation.
4. Run the architecture guard to confirm no residual `map[string]any` usage.

The `CompileStage` dispatches to `ASTFn` first. When `ASTFn` is present, the
legacy `Fn` is never called.
