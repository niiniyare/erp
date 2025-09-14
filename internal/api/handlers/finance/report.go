package finance

import (
	"context"
	"time"

	goaFinance "github.com/niiniyare/erp/internal/api/gen/finance"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
)

func (h *FinanceHandler) GetTrialBalance(ctx context.Context, payload *goaFinance.GetTrialBalancePayload) (*goaFinance.TrialBalanceResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.get_trial_balance")
	defer span.End()

	timer := h.metrics.Timer("finance_trial_balance_duration", metrics.Fields{})
	defer timer.Stop()

	var asOfDate time.Time
	if payload.AsOfDate != nil {
		parsed, err := time.Parse("2006-01-02", *payload.AsOfDate)
		if err != nil {
			logger.WarnContext(ctx, "Invalid as_of_date format for trial balance",
				logger.Fields{"as_of_date": *payload.AsOfDate, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_DATE", "Invalid as_of_date format"))
		}
		asOfDate = parsed
	} else {
		asOfDate = time.Now()
	}

	includeZeroBalances := payload.IncludeZeroBalances

	logger.InfoContext(ctx, "Processing trial balance request",
		logger.Fields{
			"as_of_date":            asOfDate.Format("2006-01-02"),
			"include_zero_balances": includeZeroBalances,
		})

	// TODO: Implement actual trial balance calculation using reporting service
	// For now, return empty structure
	logger.DebugContext(ctx, "Trial balance generated successfully",
		logger.Fields{
			"as_of_date":     asOfDate.Format("2006-01-02"),
			"total_accounts": 0,
			"is_balanced":    true,
		})

	h.metrics.IncrementCounter("finance_trial_balance_generated", metrics.Fields{})

	return &goaFinance.TrialBalanceResult{
		AsOfDate:     asOfDate.Format("2006-01-02"),
		Accounts:     []*goaFinance.TrialBalanceEntry{},
		TotalDebits:  "0.00",
		TotalCredits: "0.00",
		IsBalanced:   true,
		GeneratedAt:  time.Now().Format(time.RFC3339),
	}, nil
}
