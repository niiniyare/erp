// Package finance service layer.
//
// PostingService orchestrates journal entry posting: balance validation,
// period lock checking, ledger entry creation, and status transitions.
//
// The service is wired at startup via dependency injection. Action handlers
// in definition_journal.go and definition_payment.go call into this service.
// Currently returns "not implemented" until EntityRepository is wired.
package finance

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"awo.so/awo/def"
)

// PostingService orchestrates the journal entry posting lifecycle.
// Wire EntityRepository implementations via NewPostingService at startup.
type PostingService struct {
	// TODO: inject EntityRepository[JournalEntry] and EntityRepository[LedgerEntry]
	// via constructor when the concrete repository types are wired.
	// Example:
	//   journalRepo  EntityRepository
	//   ledgerRepo   EntityRepository
}

// NewPostingService constructs a PostingService with injected repositories.
func NewPostingService() *PostingService {
	return &PostingService{}
}

// PostJournalEntry validates balance and creates immutable LedgerEntry records.
// Transitions journal entry from "submitted" → "posted".
//
// Invariants enforced:
//   - Accounting period must be "open"
//   - sum(debit_amount) == sum(credit_amount) across all lines
//   - At least two lines required (minimum valid double-entry)
//   - Entry must be in "submitted" status
func (s *PostingService) PostJournalEntry(ctx context.Context, entryID uuid.UUID) error {
	// TODO: load JournalEntry by entryID
	// TODO: assert entry.status == "submitted"; return BusinessError if not
	// TODO: load linked AccountingPeriod; assert status == "open"
	// TODO: load all JournalEntryLines for entry
	// TODO: assert len(lines) >= 2
	// TODO: compute totalDebit = sum(line.debit_amount), totalCredit = sum(line.credit_amount)
	// TODO: if totalDebit != totalCredit → return BusinessError "unbalanced entry"
	// TODO: for each line → create LedgerEntry record
	// TODO: update JournalEntry: status="posted", total_debit=totalDebit, total_credit=totalCredit
	return fmt.Errorf("PostJournalEntry: not yet implemented — wire EntityRepository first")
}

// ReverseJournalEntry creates a mirror entry with swapped debits/credits.
// The original entry is marked "reversed"; the new entry links back via reversal_of_id.
func (s *PostingService) ReverseJournalEntry(ctx context.Context, entryID uuid.UUID, actor *def.Actor) error {
	// TODO: load original JournalEntry; assert status == "posted"
	// TODO: create new JournalEntry: is_reversal=true, reversal_of_id=entryID, status="draft"
	// TODO: load original lines; create mirrored lines with swapped debit/credit
	// TODO: submit and post the reversal entry via PostJournalEntry
	// TODO: update original entry: status="reversed"
	return fmt.Errorf("ReverseJournalEntry: not yet implemented — wire EntityRepository first")
}

// ValidateBalance checks that the given lines sum to zero net.
// Called before posting to provide early validation feedback.
func ValidateBalance(lines []*def.EntityRecord) error {
	if len(lines) < 2 {
		return fmt.Errorf("journal entry requires at least two lines; got %d", len(lines))
	}

	var totalDebit, totalCredit float64
	for _, line := range lines {
		totalDebit += toFloat64(line.Data["debit_amount"])
		totalCredit += toFloat64(line.Data["credit_amount"])
	}

	// Use epsilon comparison for floating-point safety.
	const epsilon = 0.00005
	diff := totalDebit - totalCredit
	if diff < -epsilon || diff > epsilon {
		return fmt.Errorf("unbalanced journal entry: debit %.4f ≠ credit %.4f (diff %.4f)",
			totalDebit, totalCredit, diff)
	}
	return nil
}
