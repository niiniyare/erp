ERP DSL Architecture Audit — Implementation Review

  Date: 2026-05-24 | Reviewer: Principal Architect | Verdict at end

  ---
  RUNTIME EXECUTION NOTE

  Cannot execute server (Termux sandbox — CLAUDE.md constraint). This review is static +
   structural. All findings are verifiable from code reading. No browser output
  available.

  ---
  1. IMPLEMENTATION AUDIT REPORT

  1.1 What Actually Exists

  ┌────────────────────┬────────────────────┬───────────────────────────────────────┐
  │       Layer        │       Status       │               Location                │
  ├────────────────────┼────────────────────┼───────────────────────────────────────┤
  │ AST node types     │ Partial (~60% of   │ internal/web/ast/                     │
  │                    │ documented set)    │                                       │
  ├────────────────────┼────────────────────┼───────────────────────────────────────┤
  │ UISessionContext   │ Implemented        │ internal/web/ui/types.go              │
  ├────────────────────┼────────────────────┼───────────────────────────────────────┤
  │ Pipeline           │ Implemented        │ internal/web/ui/pipeline.go           │
  │ constants/keys     │                    │                                       │
  ├────────────────────┼────────────────────┼───────────────────────────────────────┤
  │ SchemaHandler      │ Implemented        │ internal/web/handler/schema.go        │
  ├────────────────────┼────────────────────┼───────────────────────────────────────┤
  │ Pipeline stages    │ MISSING            │ Not found                             │
  ├────────────────────┼────────────────────┼───────────────────────────────────────┤
  │ DSL blocks         │ MISSING            │ internal/web/dsl/blocks/ — directory  │
  │                    │                    │ doesn't exist                         │
  ├────────────────────┼────────────────────┼───────────────────────────────────────┤
  │ DSL builders       │ MISSING            │ internal/web/dsl/builders/ —          │
  │                    │                    │ directory doesn't exist               │
  ├────────────────────┼────────────────────┼───────────────────────────────────────┤
  │ DSL screens        │ MISSING            │ internal/web/dsl/screens/ — directory │
  │                    │                    │  doesn't exist                        │
  └────────────────────┴────────────────────┴───────────────────────────────────────┘

  ---
  2. PHASE 1 — DSL ARCHITECTURE AUDIT

  Verdict: DSL layer does not exist.

  internal/web/dsl/ is absent from the codebase. Every block documented in Part 4 of the
   spec (PageHeaderBlock, FilterBarBlock, DataTableBlock, StatusBadgeColumn,
  EmptyStateBlock, KPIGridBlock, DetailCardBlock, LineItemsBlock, etc.) is
  unimplemented. The documentation describes a complete production system. The code
  contains none of it.

  This means every page currently served either:
  - Uses raw map[string]any via the legacy PageFn path (AMIS escape hatch)
  - Uses static JSON files from web/schemas/
  - Doesn't exist yet

  None of the architectural guarantees (permission-gated nodes, typed composition,
  validate-before-emit, ERP semantics) apply to anything in production.

  ---
  3. PHASE 2 — AST ARCHITECTURE REVIEW

  3.1 Strengths (genuine, not flattery)

  Node interface contract is correct:
  type Node interface {
      NodeType() string
      Validate() error
      Compile() map[string]any
  }
  Value receivers enforce immutability. ContainerNode separation is clean. CompileTree
  collects all errors before emitting any JSON — correct ordering. Error types
  (ErrRequiredField, ErrInvalidField) produce meaningful messages. Compile-time
  assertions (var _ Node = PageNode{}) catch interface breaks at build time.

  UISessionContext design is correct:
  Concrete struct with pre-resolved permissions/flags. NewUISessionContext is the sole
  constructor. No Casbin calls inside Can(). Permissions fingerprinted for cache
  invalidation. The pattern is right.

  SchemaHandler contract boundary is clean:
  Handler never reads Fiber Locals. Identity flows through Go context via
  contract.FromContext. Pipeline runner interface keeps the handler decoupled from
  pipeline implementation.

  ---
  3.2 Critical Structural Flaws

  FLAW 1 — PageNode missing InitAPI and Data (CRITICAL)

  Every real ERP page requires these. Without them, the page cannot load record data,
  inject permissions into scope, or populate breadcrumbs.

  layout.go PageNode:
  type PageNode struct {
      Title     string
      Body      []Node
      AsideBody []Node
      Toolbar   []Node
      CSSClass  string
      SubTitle  string
      Remark    string
      // MISSING: InitAPI APISpec
      // MISSING: Data    map[string]any
  }

  Doc §2.5 shows:
  ast.PageNode{
      Title:   "Invoice #${ref_number}",
      InitAPI: ast.APISpec{...},          // loads record + breadcrumbs
      Data:    ui.M{"can_approve": true},  // static scope pre-auth values
      Body:    []ast.Node{...},
  }

  Without InitAPI, no page can fetch its record. Without Data, no permission values
  reach the AMIS expression scope (${can_approve}, ${current_user_id},
  ${tenant_currency}). Every document detail page, every dashboard, every report is
  structurally impossible to build with this PageNode. Any actual page either uses the
  legacy PageFn raw path or the static JSON files — neither goes through AST
  compilation.

  FLAW 2 — TabsNode missing MountOnEnter/UnmountOnExit

  Doc §3A explicitly marks these as critical performance config — mandatory on every
  tabs node:

  // Documented requirement:
  MountOnEnter:  true,   // lazy-mount tab content
  UnmountOnExit: false,  // keep mounted after first visit

  layout.go TabsNode:
  type TabsNode struct {
      Tabs      []Tab
      Mode      string
      Mountable bool    // wrong field — maps to neither property
  }

  mountOnEnter and unmountOnExit are separate AMIS properties. Mountable bool maps to
  mountable in AMIS, which is a different property. The required mountOnEnter: true,
  unmountOnExit: false combination cannot be expressed. Result: every tabs node will
  re-mount content on every tab switch, firing redundant API calls. Doc estimates 22
  extra API calls per user per day.

  FLAW 3 — ActionNode has no Dialog/Drawer field

  Doc §3D shows:
  ast.ActionNode{
      ActionType: "dialog",
      Dialog: ast.DialogNode{   // THIS FIELD DOESN'T EXIST
          Title: "Approve Invoice",
          ...
      },
  }

  display.go ActionNode has no Dialog or Drawer field. AMIS requires the dialog schema
  to be nested inside the button config when actionType: "dialog". Without this field,
  dialog-type actions cannot be built from AST — callers must fall back to raw
  map[string]any to attach dialogs to buttons, breaking the typed contract.

  FLAW 4 — SplitPaneNode NodeType phantom

  func (s SplitPaneNode) NodeType() string { return "split_pane" }  // not an AMIS type

  func (s SplitPaneNode) Compile() ui.M {
      // ...emits "type": "grid"  // the actual AMIS type
  }

  "split_pane" is not an AMIS component. If NodeType() is used for logging, debugging,
  or any future routing logic, it returns a phantom type. AMIS receives "grid" but the
  system calls it "split_pane". This is a maintenance trap — when someone adds a switch
  n.NodeType() dispatch, SplitPaneNode silently escapes.

  FLAW 5 — StatNode emits raw HTML with custom CSS (CSS policy violation)

  display.go:
  func (s StatNode) Compile() ui.M {
      tpl := `<div class="erp-stat-card">`
      // ...erp-stat-label, erp-stat-value, erp-stat-trend--{mode}

  Doc §1.7 CSS policy: 4 rules total in awo.css. erp-stat-* classes are not in that
  list. These classes don't exist anywhere in the CSS. StatNode emits HTML that renders
  as unstyled divs. Every KPI card is broken — not at compile time (AST validation
  passes), not at server time (pipeline succeeds), only at render time in the browser as
   blank boxes.

  Additionally: StatNode.NodeType() returns "tpl" but the struct is called StatNode. The
   type signal is wrong. This is not a tpl — it's a KPI card that happens to compile to
  a tpl. Any system that iterates node types to determine component class sees "tpl" and
   cannot distinguish KPI cards from text templates.

  FLAW 6 — CRUDNode.Children() excludes RowActions

  func (c CRUDNode) Children() []Node {
      var all []Node
      all = append(all, c.Toolbar...)
      all = append(all, c.BulkActions...)
      if c.Filter != nil && c.Filter != Node(nil) {
          all = append(all, c.Filter)
      }
      return all  // RowActions []ActionNode skipped
  }

  RowActions []ActionNode are silently excluded from Children(). CompileTree will never
  call Validate() on them. A row action with empty Label or missing ActionType (invalid
  per ActionNode.Validate()) compiles to malformed AMIS schema without any error. The
  validate-before-emit guarantee is broken for all row-level actions.

  FLAW 7 — UISessionContext: interface → struct mismatch

  Doc §2.5 defines:
  type UISessionContext interface {
      Can(action, resource string) bool
      UserID() string
      TenantID() string
      Currency() string
      FeatureEnabled(flag string) bool
      Locale() string
  }

  Implementation is a concrete struct. This is actually the better design (value
  semantics = immutable, snapshottable, safe to cache), but:
  - Any external doc or test that imports UISessionContext as an interface breaks
  - Mock implementations are impossible — there's no interface to mock against
  - The Currency() method doesn't exist on the struct (Currency is a public field, not a
   method)

  FLAW 8 — unauthenticatedEnvelope in handler.go is raw map

  schema.go:198:
  func unauthenticatedEnvelope() fiber.Map {
      return fiber.Map{
          "status": 401,
          "data": fiber.Map{
              "type": "page",
              "body": fiber.Map{
                  "type": "alert",

  This is raw map[string]any producing AMIS schema — the pattern the entire architecture
   was designed to prevent. It's in the handler layer, not behind a PageFn, and won't be
   caught by NormalizeStage or ValidateStage. If AMIS changes the alert schema, this
  breaks silently at render time.

  ---
  3.3 Missing AST Nodes

  The following AMIS types are documented and required but have no ast.* implementation:

  ┌──────────────────────┬─────────────────┬──────────┬─────────────────────────────┐
  │     Missing Node     │    AMIS type    │ Severity │           Used in           │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ MappingNode          │ mapping         │ CRITICAL │ Every status column in      │
  │                      │                 │          │ every table                 │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ PropertyNode         │ property        │ CRITICAL │ Detail card display         │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ TplNode              │ tpl             │ HIGH     │ Currency display, computed  │
  │                      │                 │          │ text                        │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ AlertNode            │ alert           │ HIGH     │ Report banners, form        │
  │                      │                 │          │ warnings                    │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ NavNode / NavLink    │ nav             │ CRITICAL │ App sidebar                 │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ BreadcrumbNode       │ breadcrumb      │ HIGH     │ Every page header           │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ WizardNode /         │ wizard          │ HIGH     │ Multi-step creation flows   │
  │ WizardStep           │                 │          │                             │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ ButtonGroupNode      │ button-group    │ MEDIUM   │ Approve/reject pairs        │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ ButtonToolbarNode    │ button-toolbar  │ HIGH     │ Page toolbars               │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ DropdownButtonNode   │ dropdown-button │ MEDIUM   │ Overflow actions            │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ LinkNode             │ link            │ MEDIUM   │ Inline text links           │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ PickerNode           │ picker          │ HIGH     │ Customer/supplier entity    │
  │                      │                 │          │ selection                   │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ InputTagNode         │ input-tag       │ HIGH     │ Chip quick-filter strips    │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ InputTableNode       │ input-table     │ HIGH     │ Invoice/PO line items       │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ FormulaNode          │ formula         │ HIGH     │ Journal balance, subtotal   │
  │                      │                 │          │ calc                        │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ HiddenNode           │ hidden          │ MEDIUM   │ UUID fields in forms        │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ ImageNode            │ image           │ LOW      │ Product thumbnails          │
  ├──────────────────────┼─────────────────┼──────────┼─────────────────────────────┤
  │ IconNode             │ icon            │ MEDIUM   │ All icons throughout        │
  └──────────────────────┴─────────────────┴──────────┴─────────────────────────────┘

  MappingNode is the worst gap. Every ERP listing page has status badges. Without
  MappingNode, status columns either use the legacy PageFn raw path or aren't
  implemented. Status rendering is the most visible element of an ERP list page.

  ---
  4. PHASE 3 — RUNTIME EXECUTION REVIEW

  Server cannot be started (Termux). Inferred runtime state:

  Current schema serving path (what is actually running):
  - Static JSON files from web/schemas/ served directly
  - Legacy PageFn returning raw ui.M (the map escape hatch)
  - Pipeline stages: unknown — likely not wired up since no stage files were found

  AMIS rendering: Schema files in web/schemas/pages/ are hand-authored JSON. They bypass
   the entire compilation pipeline. There is no AST compilation, no permission gating in
   Go, no validate-before-emit in production currently.

  Pipeline stages not found: SessionStage, AuthzStage, CacheStage, RegistryStage,
  CompileStage, NormalizeStage, ValidateStage, CacheStoreStage, ResponseStage — none
  located in the codebase. The pipeline constants exist, the handler wires up a
  PipelineRunner interface, but no concrete pipeline stages found. Either they exist in
  an unscanned location or they aren't implemented.

  ---
  5. PHASE 4 — ERP DSL QUALITY AUDIT

  Can journal entry debit/credit balancing be represented?
  Partially. ComboNode can hold repeating line rows. FormulaNode (missing) would do the
  balance calculation. Without InputTableNode and FormulaNode, journal entry lines
  revert to raw map[string]any.

  Can tax aggregation / approval chains / aging reports be represented?
  No. Zero ERP-domain blocks exist. TaxSummaryBlock, ApprovalWorkflowBlock,
  AgingReportBlock are all documented and all unimplemented.

  Can inventory batch tracking / warehouse transfers be represented?
  No ERP-domain inventory blocks. Stock movement grid, lot tracking, serial number
  tables — none implemented.

  ---
  6. DSL DESIGN VERDICT

  This is not an ERP DSL. It is a well-designed foundation for an ERP DSL that has not
  been built yet.

  The AST layer has a genuinely correct architectural contract — typed nodes, immutable
  value semantics, validate-before-emit, tree traversal for validation. This is real
  compiler infrastructure thinking, not an AMIS JSON helper.

  But a compiler foundation without the language on top is infrastructure without
  product. The documented block library is the actual DSL — the vocabulary that makes
  ERP semantics expressible in Go without knowing AMIS internals. That vocabulary is
  entirely absent.

  The documentation describes a production-grade ERP UI compiler. The code implements
  the compiler backend but none of the language frontend. Everything between ast.Node
  and a working invoice screen is missing.

  ---
  7. REQUIRED REFACTOR PLAN (prioritized)

  P0 — Structural fixes (AST broken today)

  P0.1 — Add InitAPI and Data to PageNode
  Every ERP page needs these. Without them nothing can be built on the typed path.

  P0.2 — Fix TabsNode: add MountOnEnter + UnmountOnExit fields
  Replace Mountable bool with MountOnEnter bool and UnmountOnExit bool. Add default
  handling: when either is zero-value, emit the performance-correct default
  (mountOnEnter: true, unmountOnExit: false).

  P0.3 — Add Dialog and Drawer fields to ActionNode
  AMIS requires the dialog schema nested in the button. Without this every confirm/form
  dialog requires raw map escape.

  P0.4 — Add RowActions to CRUDNode.Children()
  Include RowActions in tree traversal so they receive validation.

  P0.5 — Implement MappingNode
  Status display is used on every listing page. This is the most used missing node.

  P1 — Compiler completeness

  P1.1 — Implement missing display nodes: TplNode, AlertNode, PropertyNode,
  BreadcrumbNode

  P1.2 — Implement missing form nodes: InputTableNode, InputTagNode, FormulaNode,
  HiddenNode, PickerNode

  P1.3 — Implement missing nav/action nodes: NavNode, ButtonToolbarNode,
  ButtonGroupNode, WizardNode

  P1.4 — Fix SplitPaneNode.NodeType(): Return "grid" to match compile output, or
  introduce a semantic wrapper that doesn't expose phantom types to external code.

  P1.5 — Fix StatNode: Either use AMIS statistic or card component (if exists in AMIS
  v3) instead of raw HTML injection. If raw HTML is unavoidable, move CSS classes to the
   permitted awo.css override file and document them.

  P2 — DSL block implementation

  Implement in this order (each depends on P0+P1 completion):

  1. PageHeaderBlock — needed on every page
  2. DataTableBlock — needed on every listing page
  3. StatusBadgeColumn + StatusBadgeBlock
  4. FilterBarBlock + QuickFilterChipsBlock
  5. EmptyStateBlock
  6. DetailCardBlock
  7. LineItemsBlock
  8. KPIGridBlock
  9. ApprovalWorkflowBlock
  10. ActivityFeedBlock

  P3 — Pipeline stage implementation

  Implement and wire up: SessionStage, AuthzStage, RegistryStage, CompileStage,
  NormalizeStage, ValidateStage, ResponseStage. Connect to SchemaHandler. Only then does
   the typed AST path serve actual requests.

  P4 — Architectural hardening

  - Replace unauthenticatedEnvelope raw map in handler.go with a typed AST-compiled
  schema
  - Add UISessionContext mock interface for testing
  - Add loadDataOnce: false enforcement in CRUDNode.Validate()

  ---
  8. IMPLEMENTATION TASKS

  ---
  TASK — Fix PageNode: Add InitAPI and Data fields

  PROBLEM
  PageNode has no InitAPI APISpec or Data map[string]any. Every real page needs initApi
  to fetch record data and data to inject permission scope values.

  WHY IT MATTERS
  Without these, AST compilation is structurally unusable for any production ERP page.
  All pages revert to legacy PageFn with raw maps.

  FILES
  internal/web/ast/layout.go

  IMPLEMENTATION
  Add fields to PageNode, update Compile() to emit them, update Validate() to call
  InitAPI.Validate() when set.

  TESTS
  - PageNode with InitAPI compiles to {"type":"page","initApi":"get:/api/v1/..."}
  - PageNode with Data compiles to {"type":"page","data":{...}}
  - PageNode without InitAPI compiles without initApi key

  COMMIT
  fix(ast): add InitAPI and Data to PageNode

  ---
  TASK — Fix TabsNode: separate MountOnEnter + UnmountOnExit

  PROBLEM
  Current Mountable bool doesn't map to required AMIS performance config. mountOnEnter
  and unmountOnExit are separate properties.

  WHY IT MATTERS
  Without unmountOnExit: false, every tab switch destroys and remounts content — all
  tabs re-fetch their API endpoints. In a 4-tab document form this is 3× extra API calls
   per navigation.

  FILES
  internal/web/ast/layout.go

  IMPLEMENTATION
  type TabsNode struct {
      Tabs          []Tab
      Mode          string
      MountOnEnter  bool  // default: true (performance default)
      UnmountOnExit bool  // default: false (keep mounted after first visit)
  }
  Compile defaults: if both are zero-value, emit the safe defaults (mountOnEnter: true,
  unmountOnExit: false).

  TESTS
  - Default TabsNode emits mountOnEnter: true, unmountOnExit: false
  - Explicit overrides respected

  COMMIT
  fix(ast): replace Mountable with MountOnEnter+UnmountOnExit on TabsNode

  ---
  TASK — Implement MappingNode

  PROBLEM
  MappingNode (AMIS mapping type) doesn't exist. Status columns in every table, every
  status badge outside tables — all use it. It's the most-used missing node.

  WHY IT MATTERS
  Every listing page has a status column. Without MappingNode, status display either
  uses raw maps or isn't implemented.

  FILES
  New implementation in internal/web/ast/display.go

  IMPLEMENTATION
  type MappingItem struct {
      Label string
      Level string // "success"|"warning"|"danger"|"info"|"default"
      Icon  string // optional fa class
  }

  type MappingNode struct {
      Value string           // AMIS expression, e.g. "${status}"
      Name  string           // field name when used inside table column
      Map   map[string]MappingItem
  }

  TESTS
  - Compiles to {"type":"mapping","value":"${status}","map":{...}}
  - Validate fails when Map is empty
  - Validate fails when Value is empty

  COMMIT
  feat(ast): implement MappingNode for status badge rendering

  ---
  TASK — Add Dialog/Drawer inline fields to ActionNode

  PROBLEM
  Dialog-type actions require the dialog schema nested in the button config. ActionNode
  has no Dialog or Drawer field.

  WHY IT MATTERS
  Every confirm action, every short-form modal, every approval dialog needs this.
  Without it, all dialog actions must use raw map[string]any.

  FILES
  internal/web/ast/display.go

  IMPLEMENTATION
  Add optional fields:
  type ActionNode struct {
      // ... existing fields ...
      Dialog  *DialogNode   // for ActionType "dialog"
      Drawer  *DrawerNode   // for ActionType "drawer"
  }
  Update Validate(): when ActionType == "dialog", require Dialog != nil. Update
  Compile(): emit "dialog" key from Dialog.Compile().

  COMMIT
  feat(ast): add Dialog and Drawer fields to ActionNode

  ---
  9. FINAL VERDICT

  Dimension: Production-grade
  Grade: No
  Reason: Pipeline stages missing. DSL layer doesn't exist. Pages served as raw JSON.
  ────────────────────────────────────────
  Dimension: ERP-grade
  Grade: No
  Reason: No ERP domain abstractions. No journal blocks, no approval chains, no
    tax/aging.
  ────────────────────────────────────────
  Dimension: Compiler-grade
  Grade: Partial
  Reason: AST contract is correct. CompileTree is correctly ordered. But critical nodes
    missing (PageNode.InitAPI, MappingNode, TabsNode config).
  ────────────────────────────────────────
  Dimension: Scalable to 500+ modules
  Grade: No
  Reason: No DSL block library means every module requires per-module raw AMIS JSON
    knowledge. The scale problem is not performance — it's authoring cost. Without
    blocks, each new page requires an AMIS expert.
  ────────────────────────────────────────
  Dimension: Still fundamentally CRUD scaffolding
  Grade: Yes
  Reason: Currently serving static JSON schemas. The typed compilation pipeline is wired

    up but not connected to any actual page implementations.

  The architecture is sound. The foundation is correct. The product layer is missing.

  The gap is not a design problem — it's an implementation backlog. The decisions (typed
   AST, validate-before-emit, UISessionContext pre-resolution, ContainerNode traversal)
  are the right ones. But the documentation has significantly outrun the implementation.
   A developer reading the docs would conclude this is a working system. The code shows
  it's 30% infrastructure, 0% DSL product.

  Priority: P0 structural fixes (PageNode, TabsNode, ActionNode, MappingNode) → pipeline
   stage wiring → DSL block implementation in order. Until at least one real screen
  (e.g. invoice listing) is built through the full typed path (ASTPageFn → CompileTree →
   AMIS render), the system has not been validated end-to-end.
