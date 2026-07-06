package workflow

import (
	"testing"

	"go.temporal.io/sdk/testsuite"
)

// WorkflowTestEnv wraps Temporal's test environment with convenience methods.
// Use it in unit tests to execute workflows without a running Temporal server.
type WorkflowTestEnv struct {
	*testsuite.TestWorkflowEnvironment
	suite *testsuite.WorkflowTestSuite
}

// NewTestEnv creates a WorkflowTestEnv for t. The test environment is
// automatically completed when the workflow function under test returns.
func NewTestEnv(t testing.TB) *WorkflowTestEnv {
	t.Helper()
	s := &testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()
	return &WorkflowTestEnv{
		TestWorkflowEnvironment: env,
		suite:                   s,
	}
}

// MockActivity registers a mock result for the given activity function.
// The mock is used instead of the real activity during test execution.
// result may be a value (returned as the activity result) or an error.
func (e *WorkflowTestEnv) MockActivity(fn any, result any) {
	e.OnActivity(fn).Return(result)
}
