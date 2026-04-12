package pipeline

import (
	"awo.so/internal/core/finance/domain"
	corePipeline "awo.so/internal/pipeline"
)

// NewPostTransactionRegistry builds a StageRegistry wired with the standard
// finance posting stages in priority order:
//
//	100 – LoadTransactionStage
//	200 – BalanceCheckStage
//	300 – PeriodCheckStage  (skipped if periodRepo is nil)
//	400 – AccountValidateStage
//	600 – GLPostStage
//
// Pass periodRepo = nil to skip period validation (e.g. in tests or when
// period management is not yet provisioned for a tenant).
func NewPostTransactionRegistry(
	txnRepo domain.TransactionRepository,
	accountRepo domain.AccountsRepository,
	periodRepo domain.PeriodRepository,
) *corePipeline.StageRegistry {
	r := corePipeline.NewStageRegistry()
	r.Register(
		NewLoadTransactionStage(txnRepo),
		NewBalanceCheckStage(),
		NewPeriodCheckStage(periodRepo),
		NewAccountValidateStage(accountRepo),
		NewGLPostStage(txnRepo),
	)
	return r
}

// NewPostTransactionPipeline is a convenience constructor that builds the
// registry and wraps it in a PipelineBuilder in one call.
//
// txRunner is the tenant-scoped TxRunner used for TxHooks. It may be nil when
// no TxHooks are registered (the default posting stages do not register hooks).
func NewPostTransactionPipeline(
	txnRepo domain.TransactionRepository,
	accountRepo domain.AccountsRepository,
	periodRepo domain.PeriodRepository,
	txRunner corePipeline.TxRunner,
) *corePipeline.PipelineBuilder {
	registry := NewPostTransactionRegistry(txnRepo, accountRepo, periodRepo)
	return corePipeline.NewPipelineBuilder(registry, txRunner)
}
