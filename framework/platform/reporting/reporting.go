// Package reporting provides saved-report management and async run tracking.
package reporting

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RunStatus is the state of a report run.
type RunStatus string

const (
	RunStatusRunning RunStatus = "running"
	RunStatusDone    RunStatus = "done"
	RunStatusFailed  RunStatus = "failed"
)

// Report is a saved report def.
type Report struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	Name      string          `json:"name"`
	Entity    string          `json:"entity"`
	Filters   json.RawMessage `json:"filters"`
	Columns   json.RawMessage `json:"columns"`
	CreatedBy string          `json:"created_by"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Run tracks one execution of a Report.
type Run struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	ReportID   uuid.UUID  `json:"report_id"`
	Status     RunStatus  `json:"status"`
	StartedBy  string     `json:"started_by"`
	ResultURL  string     `json:"result_url,omitempty"`
	ErrorMsg   string     `json:"error_msg,omitempty"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// Store is the persistence interface for reports and runs.
type Store interface {
	CreateReport(ctx context.Context, r *Report) error
	GetReport(ctx context.Context, id uuid.UUID) (*Report, error)
	ListReports(ctx context.Context) ([]*Report, error)
	UpdateReport(ctx context.Context, r *Report) error
	DeleteReport(ctx context.Context, id uuid.UUID) error

	CreateRun(ctx context.Context, run *Run) error
	CompleteRun(ctx context.Context, runID uuid.UUID, resultURL string) error
	FailRun(ctx context.Context, runID uuid.UUID, errMsg string) error
}

// Runner executes a report and returns a URL to the result artifact.
type Runner interface {
	Run(ctx context.Context, r *Report) (resultURL string, err error)
}

// Service manages reports and dispatches runs.
type Service struct {
	store  Store
	runner Runner
}

func NewService(store Store, runner Runner) *Service {
	return &Service{store: store, runner: runner}
}

// Execute runs a report asynchronously (records a Run, then calls runner).
// For async execution, wrap this in a Temporal activity.
func (s *Service) Execute(ctx context.Context, reportID uuid.UUID, startedBy string) error {
	r, err := s.store.GetReport(ctx, reportID)
	if err != nil {
		return err
	}
	run := &Run{ReportID: r.ID, TenantID: r.TenantID, Status: RunStatusRunning, StartedBy: startedBy}
	if err := s.store.CreateRun(ctx, run); err != nil {
		return err
	}
	url, runErr := s.runner.Run(ctx, r)
	if runErr != nil {
		return s.store.FailRun(ctx, run.ID, runErr.Error())
	}
	return s.store.CompleteRun(ctx, run.ID, url)
}
