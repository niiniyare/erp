> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "41 – Testing Strategy"
volume: "vol-09-engineering"
chapter: 41
section: "Engineering"
status: "implemented"
---

# Chapter 41 – Testing Strategy

## Table of Contents
- [41.1 Testing Philosophy](#411-testing-philosophy)
- [41.2 Unit Testing a PageFn](#412-unit-testing-a-pagefn)
- [41.3 Unit Testing an ASTPageFn](#413-unit-testing-an-astpagefn)
- [41.4 Testing Permission Gating](#414-testing-permission-gating)
- [41.5 Testing Feature Flag Gating](#415-testing-feature-flag-gating)
- [41.6 ValidateStage Testing](#416-validatestage-testing)
- [41.7 Registry Testing with ValidateRegistry](#417-registry-testing-with-validateregistry)
- [41.8 Termux Sandbox Note](#418-termux-sandbox-note)

---

## 41.1 Testing Philosophy

The AwoERP UI pipeline is **read-only and non-transactional**.  No page function
writes to the database, mutates shared state, or makes network calls.  This
makes the unit test story extremely simple:

- **`PageFn` is a pure function**: given the same `UISessionContext`, it always
  returns the same schema.  Construct a context, call the function, assert the
  output.
- **`ASTPageFn` + `CompileTree`**: typed, deterministic.  All validation errors
  are collected before any output is emitted — there are no partial results to
  reason about.
- **`ValidateRegistry()`** at startup catches registration errors before any
  request is served.  Calling it in `TestMain` makes it a fast compile-time-like
  guard.

No mocking framework is needed.  No database is needed.  No HTTP server is
needed.

---

## 41.2 Unit Testing a PageFn

A `PageFn` takes a `UISessionContext` and returns `any` (an AMIS-compatible map
or struct).  Test it by constructing a minimal context and asserting the
returned structure.

```go
func TestFinanceDashboardScreen_HasInitAPI(t *testing.T) {
    sess := ui.UISessionContext{
        TenantID:    "t_test",
        UserID:      "u_test",
        Permissions: []string{"finance.invoices.read"},
        Locale:      "en",
        Currency:    "KES",
    }

    schema := screens.FinanceDashboardScreen(sess)

    page, ok := schema.(ast.PageNode)
    if !ok {
        t.Fatalf("expected ast.PageNode, got %T", schema)
    }
    if page.InitAPI == nil {
        t.Error("expected InitAPI to be set on finance dashboard")
    }
    if page.InitAPI.URL != "/api/v1/finance/dashboard/summary" {
        t.Errorf("unexpected InitAPI URL: %s", page.InitAPI.URL)
    }
}
```

**Pattern rules:**
1. Build `UISessionContext` inline — no helper needed for simple cases.
2. Type-assert the return value to the expected node type.
3. Assert specific fields that are load-bearing (InitAPI URL, title, body
   length).  Do not assert entire JSON output — it is brittle to layout changes.

---

## 41.3 Unit Testing an ASTPageFn

`ASTPageFn` returns an `ast.Node` directly.  Use `ast.CompileTree` to validate
the node tree and collect any compilation errors before asserting schema fields.

```go
func TestInvoiceListScreen_CompilesClean(t *testing.T) {
    sess := ui.UISessionContext{
        TenantID:    "t_test",
        Permissions: []string{"finance.invoices.read"},
    }

    node := screens.InvoiceListScreen(sess)

    errs := ast.CompileTree(node)
    if len(errs) > 0 {
        t.Fatalf("CompileTree returned errors: %v", errs)
    }
}

func TestInvoiceListScreen_HasCRUDNode(t *testing.T) {
    sess := ui.UISessionContext{
        TenantID:    "t_test",
        Permissions: []string{"finance.invoices.read"},
    }

    node := screens.InvoiceListScreen(sess)

    page, ok := node.(ast.PageNode)
    if !ok {
        t.Fatalf("expected PageNode")
    }
    if len(page.Body) == 0 {
        t.Fatal("expected at least one body node")
    }
    if _, ok := page.Body[0].(ast.CRUDNode); !ok {
        t.Errorf("expected first body node to be CRUDNode, got %T", page.Body[0])
    }
}
```

`CompileTree` returns all errors collected during the AST walk — if it returns
a non-empty slice, the page would fail `ValidateStage` in production.  A clean
`CompileTree` result is a strong guarantee of schema correctness.

---

## 41.4 Testing Permission Gating

To test that a block correctly hides elements when the user lacks a permission,
construct two `UISessionContext` values — one with the permission and one without
— and assert the difference in the returned schema.

```go
func TestQuickActionsBlock_HidesCreateWhenUnauthorised(t *testing.T) {
    withPerm := ui.UISessionContext{
        TenantID:    "t_test",
        Permissions: []string{"finance.invoices.create"},
    }
    withoutPerm := ui.UISessionContext{
        TenantID:    "t_test",
        Permissions: []string{},
    }

    actionsWithPerm    := blocks.QuickActionsBlock(withPerm, financeActions)
    actionsWithoutPerm := blocks.QuickActionsBlock(withoutPerm, financeActions)

    toolbarWith    := extractToolbar(actionsWithPerm)
    toolbarWithout := extractToolbar(actionsWithoutPerm)

    if len(toolbarWith.Buttons) == len(toolbarWithout.Buttons) {
        t.Error("expected fewer buttons when permission is absent")
    }
}
```

**Key insight**: Because `UISessionContext.Permissions` is a plain `[]string`,
there is no mock needed — just set the slice to the desired permission set.  The
pre-resolved permission model makes permission testing trivially simple.

Error codes to assert for gated elements: blocks that gate on permissions should
return an empty node (`ast.NullNode{}`) or omit the restricted element entirely
rather than returning a partial node.

---

## 41.5 Testing Feature Flag Gating

Feature flags are carried in `UISessionContext.Flags` as a `map[string]bool`.
Test flag-gated blocks the same way as permission-gated blocks:

```go
func TestReportBlock_ShowsNewReportWhenFlagEnabled(t *testing.T) {
    withFlag := ui.UISessionContext{
        TenantID: "t_test",
        Flags:    map[string]bool{"feature.new_report_engine": true},
    }
    withoutFlag := ui.UISessionContext{
        TenantID: "t_test",
        Flags:    map[string]bool{},
    }

    nodeWith    := blocks.ReportBlock(withFlag)
    nodeWithout := blocks.ReportBlock(withoutFlag)

    tabsWith    := extractTabs(nodeWith)
    tabsWithout := extractTabs(nodeWithout)

    if len(tabsWith) <= len(tabsWithout) {
        t.Error("expected extra tab when new report engine flag is on")
    }
}
```

The `sess.Flag("feature.new_report_engine")` helper on `UISessionContext` reads
from `Flags` with a false default — so in tests, simply populate the `Flags` map
directly.

---

## 41.6 ValidateStage Testing

`ValidateStage` can be exercised directly without running the full pipeline.
Construct a schema and call `ValidateStage.Execute` with a minimal
`PipelineContext`:

```go
func TestValidateStage_RejectsChartWithoutStyle(t *testing.T) {
    badChart := ast.ChartNode{
        // Style is intentionally omitted — should fail validation
        ChartType: "bar",
        Data:      "${chart_data}",
    }

    page := ast.PageNode{Body: []ast.Node{badChart}}

    pCtx := &pipeline.PipelineContext{Schema: page}
    stage := stages.NewValidateStage()

    err := stage.Execute(context.Background(), pCtx)
    if err == nil {
        t.Fatal("expected ValidateStage to reject chart without style")
    }
    if !strings.Contains(err.Error(), "VALIDATE_CHART_MISSING_STYLE") {
        t.Errorf("expected VALIDATE_CHART_MISSING_STYLE error code, got: %v", err)
    }
}
```

This pattern lets you write regression tests for every validation rule without
spinning up an HTTP server.

---

## 41.7 Registry Testing with ValidateRegistry

`ValidateRegistry()` walks the entire page registry and checks every registered
`PageRegistration` for structural correctness:
- `Route` is non-empty
- Exactly one of `Fn` or `ASTFn` is set
- `Module` is non-empty
- `Title` is non-empty

Call it in `TestMain` so that a bad registration fails the test binary before
any individual test runs:

```go
// internal/web/dsl/register/register_test.go
func TestMain(m *testing.M) {
    if err := registry.ValidateRegistry(); err != nil {
        fmt.Fprintf(os.Stderr, "registry validation failed: %v\n", err)
        os.Exit(1)
    }
    os.Exit(m.Run())
}
```

This is the cheapest possible guard against typos in `RegisterPage` calls —
it runs in milliseconds and catches errors that would otherwise surface only
when a specific route is requested.

Error code prefix for registry failures: `REGISTRY_*`.

---

## 41.8 Termux Sandbox Note

> **Important**: Go test execution (`go test ./...`) cannot be run inside the
> Termux environment due to sandbox restrictions that block `/tmp` creation.
>
> All test patterns documented in this chapter are correct Go test code.  To
> run them, execute them on a non-sandboxed host (CI, a Linux dev machine, or
> Docker) using:
>
> ```sh
> go test ./internal/web/...
> ```
>
> Do not attempt to run `go test`, `go build`, or `go vet` inside Termux.
> Tell the user to run these commands in their normal development environment.
