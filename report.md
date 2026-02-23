## Report Builder Architecture

A report builder for a finance system has several distinct layers that need to be designed independently before they can work together. Let me walk through each one.

---

### The Core Problem

A report builder is fundamentally a **query generator**. The user defines what they want through a UI, that definition gets stored as configuration, and at runtime an engine translates that configuration into a SQL query, executes it, and formats the result. Every design decision flows from this.

```
User UI
  → Report Definition (stored JSON)
    → Query Engine (config → SQL)
      → Postgres
        → Result Transformer (raw rows → formatted output)
          → Renderer (table / chart / PDF / Excel)
```

---

### Layer 1 — Report Definition Storage

This is the schema that stores everything about a saved report. It needs to capture what data to fetch, how to filter it, how to group it, and how to display it — all without hardcoding any of it.

```sql
-- =====================================================================
-- REPORT DEFINITIONS
-- Stores the complete configuration of a saved report
-- =====================================================================
CREATE TABLE report_definitions (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  -- Identity
  report_code       VARCHAR(50) NOT NULL,
  report_name       VARCHAR(255) NOT NULL,
  description       TEXT,
  report_type       VARCHAR(30) NOT NULL CHECK (
    report_type IN (
      'TABULAR',        -- rows and columns, like a ledger
      'SUMMARY',        -- grouped with subtotals, like trial balance
      'FINANCIAL',      -- structured P&L / Balance Sheet / Cash Flow
      'PIVOT',          -- cross-tab with dynamic columns
      'COMPARISON',     -- side-by-side periods
      'TREND',          -- time series
      'CUSTOM'
    )
  ),
  -- Data source
  base_entity       VARCHAR(50) NOT NULL CHECK (
    base_entity IN (
      'TRANSACTIONS',
      'ENTRIES',
      'ACCOUNTS',
      'INVOICES',
      'PAYMENTS',
      'COUNTERPARTIES'
    )
  ),
  -- Report configuration (the heart of the definition)
  columns           JSONB NOT NULL DEFAULT '[]',   -- what columns to show
  filters           JSONB NOT NULL DEFAULT '[]',   -- filter conditions
  groupings         JSONB NOT NULL DEFAULT '[]',   -- GROUP BY dimensions
  sort_order        JSONB NOT NULL DEFAULT '[]',   -- ORDER BY
  metrics           JSONB NOT NULL DEFAULT '[]',   -- aggregated measures
  date_config       JSONB NOT NULL DEFAULT '{}',   -- date range settings
  comparison_config JSONB DEFAULT NULL,            -- period comparison settings
  -- Display config
  display_config    JSONB NOT NULL DEFAULT '{}',   -- formatting, totals, etc.
  chart_config      JSONB DEFAULT NULL,            -- chart type and mapping
  -- Behaviour
  is_system_report  BOOLEAN NOT NULL DEFAULT FALSE,
  is_public         BOOLEAN NOT NULL DEFAULT FALSE, -- visible to all tenant users
  is_template       BOOLEAN NOT NULL DEFAULT FALSE, -- can be copied
  allow_drill_down  BOOLEAN NOT NULL DEFAULT TRUE,
  -- Scheduling
  is_scheduled      BOOLEAN NOT NULL DEFAULT FALSE,
  schedule_config   JSONB DEFAULT NULL,
  -- Metadata
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at        TIMESTAMPTZ,
  created_by        UUID REFERENCES users(id),
  updated_by        UUID REFERENCES users(id),
  UNIQUE (tenant_id, report_code)
);

-- =====================================================================
-- REPORT PARAMETERS
-- Runtime parameters a report accepts (date range, department, etc.)
-- Users fill these in each time they run the report
-- =====================================================================
CREATE TABLE report_parameters (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  report_id       UUID NOT NULL REFERENCES report_definitions(id) ON DELETE CASCADE,
  param_key       VARCHAR(50) NOT NULL,     -- 'date_from', 'department', 'account_type'
  param_label     VARCHAR(100) NOT NULL,    -- 'Start Date', 'Department'
  param_type      VARCHAR(20) NOT NULL CHECK (
    param_type IN (
      'DATE', 'DATERANGE', 'TEXT', 'NUMBER',
      'SELECT', 'MULTISELECT', 'BOOLEAN', 'ACCOUNT', 'ENTITY'
    )
  ),
  is_required     BOOLEAN NOT NULL DEFAULT TRUE,
  default_value   TEXT,
  options_query   TEXT,     -- SQL to populate SELECT options dynamically
  display_order   INTEGER DEFAULT 999,
  UNIQUE (report_id, param_key)
);

-- =====================================================================
-- REPORT PERMISSIONS
-- Who can view, edit, run, or share a report
-- =====================================================================
CREATE TABLE report_permissions (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  report_id   UUID NOT NULL REFERENCES report_definitions(id) ON DELETE CASCADE,
  -- Grantee (one of these will be set)
  user_id     UUID REFERENCES users(id),
  role_id     UUID,                         -- reference to your roles table
  -- Permission level
  permission  VARCHAR(20) NOT NULL CHECK (
    permission IN ('VIEW', 'RUN', 'EDIT', 'SHARE', 'ADMIN')
  ),
  granted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  granted_by  UUID REFERENCES users(id)
);

-- =====================================================================
-- REPORT EXECUTION LOG
-- Every time a report is run, record it
-- Useful for performance monitoring and audit
-- =====================================================================
CREATE TABLE report_executions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  report_id       UUID NOT NULL REFERENCES report_definitions(id),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  -- Runtime context
  parameters      JSONB NOT NULL DEFAULT '{}',  -- the actual values used
  -- Performance
  started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at    TIMESTAMPTZ,
  duration_ms     INTEGER,
  row_count       INTEGER,
  -- Result
  status          VARCHAR(20) NOT NULL DEFAULT 'RUNNING' CHECK (
    status IN ('RUNNING', 'COMPLETED', 'FAILED', 'CANCELLED', 'TIMEOUT')
  ),
  error_message   TEXT,
  -- Output
  output_format   VARCHAR(20),   -- 'JSON', 'CSV', 'PDF', 'XLSX'
  output_location TEXT,          -- S3 path for large exports
  -- Audit
  executed_by     UUID REFERENCES users(id)
);

-- =====================================================================
-- SAVED REPORT SNAPSHOTS
-- Cached results for scheduled reports or slow reports
-- =====================================================================
CREATE TABLE report_snapshots (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  report_id       UUID NOT NULL REFERENCES report_definitions(id),
  execution_id    UUID REFERENCES report_executions(id),
  tenant_id       UUID NOT NULL,
  parameters      JSONB NOT NULL,
  snapshot_data   JSONB NOT NULL,           -- the actual result data
  row_count       INTEGER,
  valid_until     TIMESTAMPTZ,              -- cache expiry
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

### Layer 2 — Configuration Schema (the JSON structures)

The JSON stored in `columns`, `filters`, `groupings`, and `metrics` needs a strict schema so the engine knows how to interpret it. Define this as Go structs that serialize to JSON.

```go
// report/definition.go

// ReportDefinition is the complete in-memory representation
// of a report, deserialized from report_definitions
type ReportDefinition struct {
    ID          uuid.UUID       `json:"id"`
    TenantID    uuid.UUID       `json:"tenant_id"`
    ReportCode  string          `json:"report_code"`
    ReportName  string          `json:"report_name"`
    ReportType  ReportType      `json:"report_type"`
    BaseEntity  BaseEntity      `json:"base_entity"`
    Columns     []ColumnDef     `json:"columns"`
    Filters     []FilterDef     `json:"filters"`
    Groupings   []GroupingDef   `json:"groupings"`
    Metrics     []MetricDef     `json:"metrics"`
    SortOrder   []SortDef       `json:"sort_order"`
    DateConfig  DateConfig      `json:"date_config"`
    Comparison  *ComparisonConfig `json:"comparison_config,omitempty"`
    Display     DisplayConfig   `json:"display_config"`
    Chart       *ChartConfig    `json:"chart_config,omitempty"`
}

// ColumnDef defines a single column in the output
type ColumnDef struct {
    Key          string     `json:"key"`           // 'account_code'
    Label        string     `json:"label"`         // 'Account Code'
    SourceField  string     `json:"source_field"`  // actual DB column
    DataType     string     `json:"data_type"`     // 'text','number','date','currency'
    Format       string     `json:"format"`        // '#,##0.00', 'YYYY-MM-DD'
    Width        int        `json:"width"`
    Visible      bool       `json:"visible"`
    Pinned       bool       `json:"pinned"`
    Sortable     bool       `json:"sortable"`
    Aggregatable bool       `json:"aggregatable"`
}

// FilterDef defines one filter condition
type FilterDef struct {
    Field     string      `json:"field"`      // 'root_type', 'department'
    Operator  Operator    `json:"operator"`   // EQ, IN, BETWEEN, GT, LT, LIKE
    Value     interface{} `json:"value"`      // 'REVENUE' or ['KE','UG'] or {from,to}
    IsParam   bool        `json:"is_param"`   // TRUE = value comes from runtime params
    ParamKey  string      `json:"param_key"`  // which param to bind
    Connector string      `json:"connector"`  // 'AND' | 'OR'
}

type Operator string
const (
    OpEq      Operator = "EQ"
    OpNeq     Operator = "NEQ"
    OpIn      Operator = "IN"
    OpNotIn   Operator = "NOT_IN"
    OpBetween Operator = "BETWEEN"
    OpGt      Operator = "GT"
    OpLt      Operator = "LT"
    OpGte     Operator = "GTE"
    OpLte     Operator = "LTE"
    OpLike    Operator = "LIKE"
    OpIsNull  Operator = "IS_NULL"
    OpNotNull Operator = "NOT_NULL"
)

// MetricDef defines an aggregated measure
type MetricDef struct {
    Key        string    `json:"key"`         // 'total_revenue'
    Label      string    `json:"label"`       // 'Total Revenue'
    Expression string    `json:"expression"`  // 'SUM(debit_amount)'
    Function   AggFunc   `json:"function"`    // SUM, AVG, COUNT, MIN, MAX
    Field      string    `json:"field"`       // 'debit_amount'
    Format     string    `json:"format"`      // 'currency', 'number', 'percent'
    Condition  string    `json:"condition"`   // FILTER clause e.g. WHERE root_type='REVENUE'
}

type AggFunc string
const (
    AggSum   AggFunc = "SUM"
    AggAvg   AggFunc = "AVG"
    AggCount AggFunc = "COUNT"
    AggMin   AggFunc = "MIN"
    AggMax   AggFunc = "MAX"
)

// GroupingDef defines how to group results
type GroupingDef struct {
    Field       string `json:"field"`
    Label       string `json:"label"`
    ShowSubtotal bool  `json:"show_subtotal"`
    Collapsed   bool  `json:"collapsed_by_default"`
}

// DateConfig controls how date ranges work
type DateConfig struct {
    Mode       string `json:"mode"`        // 'FIXED','RELATIVE','PARAM','FISCAL'
    Field      string `json:"field"`       // 'transaction_date','posting_date'
    Preset     string `json:"preset"`      // 'THIS_MONTH','LAST_QUARTER','YTD'
    FixedFrom  string `json:"fixed_from"`  // used when mode = FIXED
    FixedTo    string `json:"fixed_to"`
    ParamKey   string `json:"param_key"`   // used when mode = PARAM
}

// ComparisonConfig defines period comparison behaviour
type ComparisonConfig struct {
    Enabled      bool     `json:"enabled"`
    Periods      []Period `json:"periods"`
    ShowVariance bool     `json:"show_variance"`
    ShowPctChange bool    `json:"show_pct_change"`
}

type Period struct {
    Label    string `json:"label"`
    Offset   string `json:"offset"`    // '-1Y','-1Q','-1M' relative to primary
    DateFrom string `json:"date_from"` // or explicit dates
    DateTo   string `json:"date_to"`
}

// DisplayConfig controls report presentation
type DisplayConfig struct {
    ShowGrandTotal    bool   `json:"show_grand_total"`
    ShowSubtotals     bool   `json:"show_subtotals"`
    ShowZeroRows      bool   `json:"show_zero_rows"`
    PaginationSize    int    `json:"pagination_size"`
    DefaultBasis      string `json:"default_basis"`   // 'ACCRUAL' | 'CASH'
    CurrencyCode      string `json:"currency_code"`
    IndentHierarchy   bool   `json:"indent_hierarchy"`
    FreezeFirstColumn bool   `json:"freeze_first_column"`
}
```

---

### Layer 3 — The Query Engine

This is the most critical piece. It translates a `ReportDefinition` + runtime parameters into a safe, parameterized SQL query.

```go
// report/engine/query_builder.go

type QueryBuilder struct {
    def    *ReportDefinition
    params map[string]interface{}  // runtime parameter values
    args   []interface{}           // positional SQL args ($1, $2 ...)
    argIdx int
}

func NewQueryBuilder(def *ReportDefinition, params map[string]interface{}) *QueryBuilder {
    return &QueryBuilder{
        def:    def,
        params: params,
        argIdx: 1,
    }
}

// Build produces the final SQL and argument list
func (qb *QueryBuilder) Build() (string, []interface{}, error) {
    // Validate the definition before building
    if err := qb.validate(); err != nil {
        return "", nil, fmt.Errorf("invalid report definition: %w", err)
    }

    var q strings.Builder

    q.WriteString(qb.buildSelect())
    q.WriteString(qb.buildFrom())
    q.WriteString(qb.buildJoins())
    q.WriteString(qb.buildWhere())
    q.WriteString(qb.buildGroupBy())
    q.WriteString(qb.buildHaving())
    q.WriteString(qb.buildOrderBy())
    q.WriteString(qb.buildLimit())

    return q.String(), qb.args, nil
}

func (qb *QueryBuilder) buildSelect() string {
    var parts []string

    // Regular columns
    for _, col := range qb.def.Columns {
        if !col.Visible {
            continue
        }
        safe := qb.safeField(col.SourceField)
        parts = append(parts, fmt.Sprintf("%s AS %s", safe, col.Key))
    }

    // Aggregated metrics
    for _, m := range qb.def.Metrics {
        expr := qb.buildMetricExpression(m)
        parts = append(parts, fmt.Sprintf("%s AS %s", expr, m.Key))
    }

    if len(parts) == 0 {
        return "SELECT *\n"
    }
    return "SELECT\n  " + strings.Join(parts, ",\n  ") + "\n"
}

func (qb *QueryBuilder) buildMetricExpression(m MetricDef) string {
    field := qb.safeField(m.Field)
    base  := fmt.Sprintf("%s(%s)", m.Function, field)

    // FILTER clause for conditional aggregation
    if m.Condition != "" {
        return fmt.Sprintf("%s FILTER (WHERE %s)", base, m.Condition)
    }
    return base
}

func (qb *QueryBuilder) buildFrom() string {
    table := qb.baseEntityTable()
    return fmt.Sprintf("FROM %s\n", table)
}

func (qb *QueryBuilder) baseEntityTable() string {
    switch qb.def.BaseEntity {
    case "TRANSACTIONS":
        return "finance_transactions ft"
    case "ENTRIES":
        return "finance_transaction_entries fte"
    case "ACCOUNTS":
        return "finance_accounts fa"
    case "INVOICES":
        return "finance_invoices inv"
    case "PAYMENTS":
        return "finance_payments pmt"
    default:
        return "finance_transactions ft"
    }
}

func (qb *QueryBuilder) buildJoins() string {
    // Determine which joins are needed based on fields referenced
    needed := qb.referencedTables()
    var joins []string

    base := qb.def.BaseEntity
    for _, t := range needed {
        j := qb.joinClause(base, t)
        if j != "" {
            joins = append(joins, j)
        }
    }
    return strings.Join(joins, "\n") + "\n"
}

func (qb *QueryBuilder) joinClause(base, target string) string {
    type joinKey struct{ from, to string }
    joinMap := map[joinKey]string{
        {"ENTRIES", "TRANSACTIONS"}: "JOIN finance_transactions ft ON ft.id = fte.transaction_id AND ft.deleted_at IS NULL",
        {"ENTRIES", "ACCOUNTS"}:     "JOIN finance_accounts fa ON fa.id = fte.account_id",
        {"ENTRIES", "GROUPS"}:       "LEFT JOIN finance_account_groups fag ON fag.id = fa.account_group_id",
        {"TRANSACTIONS", "ENTRIES"}: "LEFT JOIN finance_transaction_entries fte ON fte.transaction_id = ft.id",
        {"INVOICES", "COUNTERPARTIES"}: "JOIN finance_counterparties cp ON cp.id = inv.counterparty_id",
    }
    clause, ok := joinMap[joinKey{base, target}]
    if !ok {
        return ""
    }
    return clause
}

func (qb *QueryBuilder) buildWhere() string {
    conditions := []string{}

    // Always enforce tenant isolation
    tenantArg := qb.nextArg(qb.def.TenantID)
    conditions = append(conditions,
        fmt.Sprintf("ft.tenant_id = %s", tenantArg),
        "ft.deleted_at IS NULL",
    )

    // Date filter
    dateConditions := qb.buildDateConditions()
    conditions = append(conditions, dateConditions...)

    // User-defined filters
    for _, f := range qb.def.Filters {
        cond, err := qb.buildFilterCondition(f)
        if err != nil {
            continue // log and skip invalid filters
        }
        conditions = append(conditions, cond)
    }

    if len(conditions) == 0 {
        return ""
    }
    return "WHERE\n  " + strings.Join(conditions, "\n  AND ") + "\n"
}

func (qb *QueryBuilder) buildFilterCondition(f FilterDef) (string, error) {
    field := qb.safeField(f.Field)

    // Bind runtime param if needed
    value := f.Value
    if f.IsParam {
        v, ok := qb.params[f.ParamKey]
        if !ok {
            return "", fmt.Errorf("missing required parameter: %s", f.ParamKey)
        }
        value = v
    }

    switch f.Operator {
    case OpEq:
        return fmt.Sprintf("%s = %s", field, qb.nextArg(value)), nil

    case OpNeq:
        return fmt.Sprintf("%s != %s", field, qb.nextArg(value)), nil

    case OpIn:
        vals, ok := value.([]interface{})
        if !ok {
            return "", fmt.Errorf("IN operator requires array value")
        }
        placeholders := make([]string, len(vals))
        for i, v := range vals {
            placeholders[i] = qb.nextArg(v)
        }
        return fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ",")), nil

    case OpBetween:
        m, ok := value.(map[string]interface{})
        if !ok {
            return "", fmt.Errorf("BETWEEN operator requires {from, to}")
        }
        return fmt.Sprintf("%s BETWEEN %s AND %s",
            field, qb.nextArg(m["from"]), qb.nextArg(m["to"])), nil

    case OpLike:
        return fmt.Sprintf("%s ILIKE %s", field, qb.nextArg(fmt.Sprintf("%%%v%%", value))), nil

    case OpIsNull:
        return fmt.Sprintf("%s IS NULL", field), nil

    case OpNotNull:
        return fmt.Sprintf("%s IS NOT NULL", field), nil

    case OpGt:
        return fmt.Sprintf("%s > %s", field, qb.nextArg(value)), nil

    case OpLt:
        return fmt.Sprintf("%s < %s", field, qb.nextArg(value)), nil

    default:
        return "", fmt.Errorf("unsupported operator: %s", f.Operator)
    }
}

func (qb *QueryBuilder) buildGroupBy() string {
    if len(qb.def.Groupings) == 0 {
        return ""
    }
    var fields []string
    for _, g := range qb.def.Groupings {
        fields = append(fields, qb.safeField(g.Field))
    }
    return "GROUP BY " + strings.Join(fields, ", ") + "\n"
}

func (qb *QueryBuilder) buildOrderBy() string {
    if len(qb.def.SortOrder) == 0 {
        return ""
    }
    var parts []string
    for _, s := range qb.def.SortOrder {
        dir := "ASC"
        if s.Descending {
            dir = "DESC"
        }
        parts = append(parts, fmt.Sprintf("%s %s", qb.safeField(s.Field), dir))
    }
    return "ORDER BY " + strings.Join(parts, ", ") + "\n"
}

// safeField prevents SQL injection by allowlisting valid field names
func (qb *QueryBuilder) safeField(field string) string {
    allowed := map[string]string{
        // Transactions
        "transaction_date":    "ft.transaction_date",
        "posting_date":        "ft.posting_date",
        "transaction_number":  "ft.transaction_number",
        "transaction_status":  "ft.transaction_status",
        "total_debit_amount":  "ft.total_debit_amount",
        "total_credit_amount": "ft.total_credit_amount",
        "description":         "ft.description",
        "source_module":       "ft.source_module",
        // Entries
        "debit_amount":   "fte.debit_amount",
        "credit_amount":  "fte.credit_amount",
        "department":     "fte.department",
        "cost_center":    "fte.cost_center",
        // Accounts
        "account_code":    "fa.account_code",
        "account_name":    "fa.account_name",
        "root_type":       "fa.root_type",
        "account_type":    "fa.account_type",
        "normal_balance":  "fa.normal_balance",
        "account_category":"fa.account_category",
        // Groups
        "account_group":   "fag.group_name",
        "statement_order": "fag.statement_order",
        // Invoices
        "invoice_date":         "inv.invoice_date",
        "due_date":             "inv.due_date",
        "outstanding_amount":   "inv.outstanding_amount",
        "invoice_status":       "inv.invoice_status",
        // Counterparty
        "counterparty_name": "cp.counterparty_name",
        "counterparty_code": "cp.counterparty_code",
    }
    safe, ok := allowed[field]
    if !ok {
        // Never interpolate unknown fields — fail loudly
        panic(fmt.Sprintf("report engine: unknown field reference '%s'", field))
    }
    return safe
}

// nextArg adds a value to the args list and returns its placeholder
func (qb *QueryBuilder) nextArg(v interface{}) string {
    qb.args = append(qb.args, v)
    placeholder := fmt.Sprintf("$%d", qb.argIdx)
    qb.argIdx++
    return placeholder
}

// validate checks the definition for obvious problems before building
func (qb *QueryBuilder) validate() error {
    if qb.def.TenantID == uuid.Nil {
        return fmt.Errorf("tenant_id is required")
    }
    if qb.def.BaseEntity == "" {
        return fmt.Errorf("base_entity is required")
    }
    if len(qb.def.Columns) == 0 && len(qb.def.Metrics) == 0 {
        return fmt.Errorf("at least one column or metric is required")
    }
    // Validate all filter fields are in the allowlist
    for _, f := range qb.def.Filters {
        if f.IsParam && qb.params[f.ParamKey] == nil && f.Operator != OpIsNull {
            return fmt.Errorf("required parameter '%s' is missing", f.ParamKey)
        }
    }
    return nil
}
```

---

### Layer 4 — Report Executor

Sits between the API handler and the query builder. Handles caching, execution logging, timeout, and result transformation.

```go
// report/executor.go

type Executor struct {
    db        *sql.DB
    reportRepo ReportRepository
    snapRepo   SnapshotRepository
    cache      Cache
}

type RunOptions struct {
    TenantID     uuid.UUID
    ReportID     uuid.UUID
    Parameters   map[string]interface{}
    OutputFormat string    // JSON, CSV, XLSX, PDF
    UseCache     bool
    CacheTTL     time.Duration
    ExecutedBy   uuid.UUID
    Timeout      time.Duration
}

func (e *Executor) Run(ctx context.Context, opts RunOptions) (*ReportResult, error) {
    // 1. Load report definition
    def, err := e.reportRepo.GetByID(ctx, opts.TenantID, opts.ReportID)
    if err != nil {
        return nil, fmt.Errorf("load report: %w", err)
    }

    // 2. Enforce tenant isolation — definition must belong to this tenant
    if def.TenantID != opts.TenantID {
        return nil, ErrUnauthorized
    }

    // 3. Check cache if enabled
    if opts.UseCache {
        if snapshot := e.checkCache(ctx, def, opts.Parameters); snapshot != nil {
            return snapshot, nil
        }
    }

    // 4. Build query
    def.TenantID = opts.TenantID
    qb  := NewQueryBuilder(def, opts.Parameters)
    sql, args, err := qb.Build()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    // 5. Log execution start
    execID := e.logExecutionStart(ctx, def, opts)

    // 6. Execute with timeout
    queryCtx, cancel := context.WithTimeout(ctx,
        coalesce(opts.Timeout, 30*time.Second))
    defer cancel()

    start := time.Now()
    rows, err := e.db.QueryContext(queryCtx, sql, args...)
    if err != nil {
        e.logExecutionEnd(ctx, execID, "FAILED", 0, err.Error())
        return nil, fmt.Errorf("execute report: %w", err)
    }
    defer rows.Close()

    // 7. Scan results
    result, err := e.scanRows(rows, def)
    if err != nil {
        e.logExecutionEnd(ctx, execID, "FAILED", 0, err.Error())
        return nil, fmt.Errorf("scan results: %w", err)
    }

    duration := time.Since(start)
    e.logExecutionEnd(ctx, execID, "COMPLETED", len(result.Rows), "")

    // 8. Post-process (subtotals, percentages, comparison periods)
    result = e.postProcess(result, def, opts.Parameters)

    // 9. Cache if enabled and report is slow enough to warrant it
    if opts.UseCache && duration > 2*time.Second {
        e.cacheResult(ctx, def, opts.Parameters, result, opts.CacheTTL)
    }

    return result, nil
}

func (e *Executor) scanRows(rows *sql.Rows, def *ReportDefinition) (*ReportResult, error) {
    cols, err := rows.Columns()
    if err != nil {
        return nil, err
    }

    var result ReportResult
    result.Columns = cols

    for rows.Next() {
        // Scan into generic interface slice
        values := make([]interface{}, len(cols))
        pointers := make([]interface{}, len(cols))
        for i := range values {
            pointers[i] = &values[i]
        }

        if err := rows.Scan(pointers...); err != nil {
            return nil, err
        }

        row := make(map[string]interface{}, len(cols))
        for i, col := range cols {
            row[col] = values[i]
        }
        result.Rows = append(result.Rows, row)
    }
    return &result, rows.Err()
}

// postProcess adds subtotals, percentages, hierarchy indentation
func (e *Executor) postProcess(
    result *ReportResult,
    def *ReportDefinition,
    params map[string]interface{},
) *ReportResult {
    if def.Display.ShowSubtotals && len(def.Groupings) > 0 {
        result = e.injectSubtotals(result, def)
    }
    if def.Display.ShowGrandTotal {
        result = e.injectGrandTotal(result, def)
    }
    return result
}
```

---

### Layer 5 — Report Service API

The HTTP handler layer, clean and thin.

```go
// api/report_handler.go

type ReportHandler struct {
    executor *report.Executor
    defSvc   *report.DefinitionService
}

// POST /api/reports/:id/run
func (h *ReportHandler) RunReport(w http.ResponseWriter, r *http.Request) {
    tenantID := middleware.TenantFromContext(r.Context())
    reportID := chi.URLParam(r, "id")
    userID   := middleware.UserFromContext(r.Context())

    var body struct {
        Parameters   map[string]interface{} `json:"parameters"`
        OutputFormat string                 `json:"output_format"`
        UseCache     bool                   `json:"use_cache"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    result, err := h.executor.Run(r.Context(), report.RunOptions{
        TenantID:     tenantID,
        ReportID:     uuid.MustParse(reportID),
        Parameters:   body.Parameters,
        OutputFormat: body.OutputFormat,
        UseCache:     body.UseCache,
        ExecutedBy:   userID,
    })
    if err != nil {
        switch {
        case errors.Is(err, report.ErrUnauthorized):
            http.Error(w, "unauthorized", http.StatusForbidden)
        case errors.Is(err, context.DeadlineExceeded):
            http.Error(w, "report timed out", http.StatusRequestTimeout)
        default:
            http.Error(w, "report failed", http.StatusInternalServerError)
        }
        return
    }

    // Format output
    switch body.OutputFormat {
    case "XLSX":
        h.writeExcel(w, result)
    case "CSV":
        h.writeCSV(w, result)
    case "PDF":
        h.writePDF(w, result)
    default:
        json.NewEncoder(w).Encode(result)
    }
}

// POST /api/reports
func (h *ReportHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
    tenantID := middleware.TenantFromContext(r.Context())
    userID   := middleware.UserFromContext(r.Context())

    var def report.ReportDefinition
    if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
        http.Error(w, "invalid definition", http.StatusBadRequest)
        return
    }
    def.TenantID  = tenantID
    def.CreatedBy = userID

    saved, err := h.defSvc.Create(r.Context(), &def)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    json.NewEncoder(w).Encode(saved)
}

// GET /api/reports/:id/preview
// Runs a report with a LIMIT 100 for UI preview purposes
func (h *ReportHandler) PreviewReport(w http.ResponseWriter, r *http.Request) {
    // Same as RunReport but injects LIMIT 100 into options
}
```

---

### Layer 6 — Built-In Report Templates

Ship these out of the box so users can use the system immediately and optionally customize them.

```go
// report/builtins/builtins.go

var BuiltinReports = []ReportDefinition{
    {
        ReportCode: "TRIAL_BALANCE",
        ReportName: "Trial Balance",
        ReportType: "SUMMARY",
        BaseEntity: "ENTRIES",
        Columns: []ColumnDef{
            {Key: "account_code", Label: "Code",         SourceField: "account_code"},
            {Key: "account_name", Label: "Account",      SourceField: "account_name"},
            {Key: "root_type",    Label: "Type",         SourceField: "root_type"},
            {Key: "account_group",Label: "Group",        SourceField: "account_group"},
        },
        Metrics: []MetricDef{
            {Key: "total_debits",  Label: "Debits",  Function: AggSum, Field: "debit_amount"},
            {Key: "total_credits", Label: "Credits", Function: AggSum, Field: "credit_amount"},
        },
        Filters: []FilterDef{
            {Field: "transaction_status", Operator: OpEq, Value: "POSTED"},
            {Field: "transaction_date",   Operator: OpBetween, IsParam: true, ParamKey: "date_range"},
        },
        Groupings: []GroupingDef{
            {Field: "root_type",    ShowSubtotal: true},
            {Field: "account_code", ShowSubtotal: false},
        },
        SortOrder: []SortDef{
            {Field: "root_type"},
            {Field: "account_code"},
        },
        IsSystemReport: true,
    },
    {
        ReportCode: "AR_AGING",
        ReportName: "AR Aging Report",
        ReportType: "SUMMARY",
        BaseEntity: "INVOICES",
        Columns: []ColumnDef{
            {Key: "counterparty_name", Label: "Customer",    SourceField: "counterparty_name"},
            {Key: "invoice_number",    Label: "Invoice",     SourceField: "invoice_number"},
            {Key: "invoice_date",      Label: "Date",        SourceField: "invoice_date"},
            {Key: "due_date",          Label: "Due Date",    SourceField: "due_date"},
            {Key: "invoice_status",    Label: "Status",      SourceField: "invoice_status"},
        },
        Metrics: []MetricDef{
            {Key: "outstanding", Label: "Outstanding", Function: AggSum, Field: "outstanding_amount"},
        },
        Filters: []FilterDef{
            {Field: "invoice_type", Operator: OpEq, Value: "AR"},
            {Field: "invoice_status", Operator: OpNotIn,
             Value: []interface{}{"CANCELLED", "DRAFT"}},
        },
        IsSystemReport: true,
    },
}

// SeedBuiltinReports stamps the builtins into the database on tenant provisioning
func SeedBuiltinReports(ctx context.Context, repo ReportRepository, tenantID uuid.UUID) error {
    for _, def := range BuiltinReports {
        def.TenantID       = tenantID
        def.IsSystemReport = true
        def.IsPublic       = true
        if err := repo.Upsert(ctx, &def); err != nil {
            return fmt.Errorf("seed report %s: %w", def.ReportCode, err)
        }
    }
    return nil
}
```

---

### Layer 7 — Output Formatters

```go
// report/formatters/excel.go
func (h *ReportHandler) writeExcel(w http.ResponseWriter, result *ReportResult) {
    f := excelize.NewFile()
    sheet := "Report"
    f.NewSheet(sheet)

    // Header row
    for i, col := range result.Columns {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue(sheet, cell, col)
    }

    // Data rows
    for rowIdx, row := range result.Rows {
        for colIdx, col := range result.Columns {
            cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
            f.SetCellValue(sheet, cell, row[col])
        }
    }

    // Style currency columns
    currencyStyle, _ := f.NewStyle(&excelize.Style{
        NumFmt: 4,  // #,##0.00
    })
    for i, col := range result.ColumnDefs {
        if col.Format == "currency" {
            colName, _ := excelize.ColumnNumberToName(i + 1)
            f.SetColStyle(sheet, colName, currencyStyle)
        }
    }

    w.Header().Set("Content-Type",
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    w.Header().Set("Content-Disposition",
        fmt.Sprintf(`attachment; filename="%s.xlsx"`, result.ReportName))
    f.Write(w)
}
```

---

### Complete Architecture Diagram

```
┌──────────────── Frontend ────────────────────────────────┐
│                                                           │
│  Report Builder UI         Report Viewer                  │
│  ┌───────────────────┐     ┌──────────────────────────┐  │
│  │ Column picker      │     │ Table / Pivot / Chart    │  │
│  │ Filter builder     │     │ Drill-down               │  │
│  │ Group by selector  │     │ Export buttons           │  │
│  │ Date range picker  │     │ Parameter inputs         │  │
│  │ Preview panel      │     └──────────────────────────┘  │
│  └────────┬──────────┘                   ↑               │
└───────────┼──────────────────────────────┼───────────────┘
            │ POST /reports                │ GET /reports/:id/run
            ▼                             │
┌──────────────── Go API ──────────────────────────────────┐
│                                                           │
│  ReportHandler (HTTP)                                     │
│       │                                                   │
│  DefinitionService ──── ReportRepository                  │
│       │                    (CRUD on report_definitions)   │
│  Executor                                                 │
│   ├── QueryBuilder     ← safeField() allowlist           │
│   │    (config → SQL)    validates all field refs         │
│   ├── Result Scanner                                      │
│   ├── PostProcessor                                       │
│   │    (subtotals, %, hierarchy)                         │
│   └── Cache Layer      ← report_snapshots table          │
│                                                           │
│  Formatters                                               │
│   ├── JSON (default)                                      │
│   ├── Excel (excelize)                                    │
│   ├── CSV  (encoding/csv)                                 │
│   └── PDF  (chromedp / gotenberg)                         │
│                                                           │
│  Scheduler                                                │
│   └── reads report_definitions WHERE is_scheduled=TRUE   │
│       runs on cron, writes to report_snapshots           │
└──────────────────────────────────────────────────────────┘
            │
            ▼ parameterized SQL only, no dynamic field names
┌──────────────── Postgres ────────────────────────────────┐
│                                                           │
│  report_definitions                                       │
│  report_parameters                                        │
│  report_permissions                                       │
│  report_executions     (audit log)                       │
│  report_snapshots      (cache)                           │
│                                                           │
│  finance_* tables      (source data)                     │
└──────────────────────────────────────────────────────────┘
```

---

### The Most Important Design Decisions

**SQL injection is the primary risk.** The `safeField()` allowlist is not optional. The query builder must never interpolate user-controlled strings into SQL directly. Every field reference must go through the allowlist. Arguments must always be parameterized with `$1`, `$2`.

**Keep the query builder thin.** It should produce a single clean SQL query with joins, filters, grouping and ordering. All the post-processing — subtotals, percentages, comparisons, hierarchy indentation — happens in Go after the rows come back. This keeps the SQL debuggable and keeps the business logic testable.

**Report definitions are tenant data.** They must be validated for tenant ownership before execution. A user from tenant A must never be able to run or modify a report belonging to tenant B, even if they guess the UUID.

**Execution logging is not optional in a finance system.** Every report run needs to be recorded with who ran it, what parameters they used, and how long it took. This is both an audit requirement and a performance monitoring tool.

**Start with 6-8 built-in reports.** Users will customise those rather than building from scratch. The blank canvas of a report builder is intimidating. Pre-built reports for trial balance, AR aging, AP aging, P&L, and general ledger give immediate value and teach users what the system is capable of.
