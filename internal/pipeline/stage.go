package pipeline

// Stage is the fundamental unit of work in a pipeline. Each stage is
// independently testable, declares its own feature-flag gate, priority, and
// optional compensation function.
//
// Naming convention: "{module}.{description}" — e.g. "budget.check",
// "gl.post_transaction", "tax.calculate".
type Stage interface {
	// Name returns the unique identifier for this stage.
	Name() string

	// Operations returns the operation keys this stage applies to.
	// Use []string{"*"} to apply to every operation.
	// Example: []string{"ap.invoice.process", "ap.invoice.approve"}
	Operations() []string

	// FeatureFlag is the tenant feature flag that gates this stage.
	// An empty string means the stage is always included.
	// The PipelineBuilder excludes the stage entirely if the flag is off.
	FeatureFlag() string

	// Priority controls execution order within a pipeline.
	// Standard bands:
	//   100–199  Validation
	//   200–299  Enrichment / data fetching
	//   300–399  Business rule checks
	//   400–499  Tax & regulatory
	//   500–599  Workflow & approval gates
	//   600–699  Core financial posting
	//   700–799  Secondary financial & storage
	//   800–899  Events & notifications
	//   900–999  Audit & compliance
	Priority() int

	// Required controls failure behaviour.
	// true  → pipeline aborts and compensates if Execute returns an error.
	// false → error is logged, execution continues.
	Required() bool

	// RunCondition is an optional expr-lang formula evaluated by pkg/condition
	// before Execute is called. An empty string means always run.
	// The formula has access to opCtx.Data and opCtx.Flags.
	RunCondition() string

	// DependsOn returns the names of stages whose output this stage reads.
	// The pipeline uses this to identify which stages in the same priority
	// band can run in parallel (those sharing no dependency).
	DependsOn() []string

	// Execute performs the stage's work.
	Execute(opCtx *OperationContext) (StageResult, error)
}

// Simulatable is an optional extension to Stage for stages that have side
// effects. When opCtx.DryRun is true the pipeline calls Simulate instead of
// Execute, allowing the stage to report its expected outcome without writing
// to the database.
type Simulatable interface {
	Simulate(opCtx *OperationContext) (StageResult, error)
}

// StageResult carries the outcome of a single stage execution.
type StageResult struct {
	// Status is one of: "completed" | "skipped" | "suspended" | "simulated" | "failed"
	Status string

	// Outputs are key/value pairs written to opCtx.Data after execution.
	// Keys should follow the "{stage_name}.{key}" convention.
	Outputs map[string]any

	// Message is a human-readable description of the outcome, used for UI
	// progress display and the stage log.
	Message string

	// NextStageID is populated by condition/script stages that branch
	// execution to a specific named stage rather than the next in priority order.
	NextStageID string
}

// stageCheckpoint records a stage that has completed so it can be compensated
// in LIFO order if the pipeline aborts later.
type stageCheckpoint struct {
	StageName string
	Output    map[string]any
}

// BaseStage provides a default implementation of all Stage interface methods
// except Execute. Embed it in a concrete stage struct and override only what
// differs.
//
// Example:
//
//	type BudgetCheckStage struct {
//	    pipeline.BaseStage
//	    svc BudgetService
//	}
//
//	func NewBudgetCheckStage(svc BudgetService) *BudgetCheckStage {
//	    return &BudgetCheckStage{
//	        BaseStage: pipeline.BaseStage{
//	            StageName:        "budget.check",
//	            StageOperations:  []string{"ap.invoice.process"},
//	            StageFeatureFlag: "budget",
//	            StagePriority:    310,
//	            StageRequired:    false,
//	        },
//	        svc: svc,
//	    }
//	}
type BaseStage struct {
	StageName        string
	StageOperations  []string
	StageFeatureFlag string
	StagePriority    int
	StageRequired    bool
	StageRunCond     string
	StageDependsOn   []string
}

func (b BaseStage) Name() string         { return b.StageName }
func (b BaseStage) Operations() []string { return b.StageOperations }
func (b BaseStage) FeatureFlag() string  { return b.StageFeatureFlag }
func (b BaseStage) Priority() int        { return b.StagePriority }
func (b BaseStage) Required() bool       { return b.StageRequired }
func (b BaseStage) RunCondition() string { return b.StageRunCond }
func (b BaseStage) DependsOn() []string  { return b.StageDependsOn }
