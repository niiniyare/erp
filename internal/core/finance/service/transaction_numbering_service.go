package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TransactionNumberingService handles automatic generation of transaction numbers
type TransactionNumberingService interface {
	// GenerateTransactionNumber generates a unique transaction number
	GenerateTransactionNumber(ctx context.Context, req GenerateNumberRequest) (string, error)

	// ReserveTransactionNumber reserves a transaction number for later use
	ReserveTransactionNumber(ctx context.Context, req GenerateNumberRequest) (string, error)

	// ConfirmTransactionNumber confirms usage of a reserved number
	ConfirmTransactionNumber(ctx context.Context, number string) error

	// ReleaseTransactionNumber releases a reserved number back to the pool
	ReleaseTransactionNumber(ctx context.Context, number string) error

	// GetNextNumber gets the next number without generating (for preview)
	GetNextNumber(ctx context.Context, req GenerateNumberRequest) (string, error)

	// ValidateTransactionNumber validates a transaction number format
	ValidateTransactionNumber(ctx context.Context, number string, transactionType domain.TransactionType) error
}

// GenerateNumberRequest represents a request to generate a transaction number
type GenerateNumberRequest struct {
	TenantID        uuid.UUID              `json:"tenant_id"`
	EntityID        *uuid.UUID             `json:"entity_id,omitempty"`
	TransactionType domain.TransactionType `json:"transaction_type"`
	TransactionDate time.Time              `json:"transaction_date"`
	Series          *string                `json:"series,omitempty"` // Optional custom series
	CustomPrefix    *string                `json:"custom_prefix,omitempty"`
	CustomSuffix    *string                `json:"custom_suffix,omitempty"`
}

// NumberingRule defines how transaction numbers are generated
type NumberingRule struct {
	ID              uuid.UUID              `json:"id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	EntityID        *uuid.UUID             `json:"entity_id,omitempty"`
	TransactionType domain.TransactionType `json:"transaction_type"`
	Series          string                 `json:"series"`                // Series identifier (e.g., "MAIN", "ADJ")
	Prefix          string                 `json:"prefix"`                // Fixed prefix (e.g., "TXN", "JE")
	DateFormat      *string                `json:"date_format,omitempty"` // Date format in number (e.g., "YYYY", "YYYYMM")
	NumberLength    int                    `json:"number_length"`         // Length of numeric portion
	Suffix          *string                `json:"suffix,omitempty"`
	ResetFrequency  ResetFrequency         `json:"reset_frequency"` // When to reset sequence
	StartingNumber  int64                  `json:"starting_number"` // Starting sequence number
	CurrentNumber   int64                  `json:"current_number"`  // Current sequence number
	IsActive        bool                   `json:"is_active"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ResetFrequency defines when sequence numbers reset
type ResetFrequency string

const (
	ResetFrequencyNever     ResetFrequency = "NEVER"     // Never reset
	ResetFrequencyDaily     ResetFrequency = "DAILY"     // Reset daily
	ResetFrequencyMonthly   ResetFrequency = "MONTHLY"   // Reset monthly
	ResetFrequencyYearly    ResetFrequency = "YEARLY"    // Reset yearly
	ResetFrequencyQuarterly ResetFrequency = "QUARTERLY" // Reset quarterly
)

// NumberReservation tracks reserved transaction numbers
type NumberReservation struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	TransactionNumber string     `json:"transaction_number"`
	ReservedBy        uuid.UUID  `json:"reserved_by"`
	ReservedAt        time.Time  `json:"reserved_at"`
	ExpiresAt         time.Time  `json:"expires_at"`
	IsConfirmed       bool       `json:"is_confirmed"`
	ConfirmedAt       *time.Time `json:"confirmed_at,omitempty"`
}

// transactionNumberingService implements TransactionNumberingService
type transactionNumberingService struct {
	// Note: In a full implementation, this would use a repository for persistent storage
	rules        map[string]*NumberingRule // In-memory storage for demo
	reservations map[string]*NumberReservation
	mutex        sync.RWMutex
	tracing      tracing.TracingService
}

// Dependencies for TransactionNumberingService
type TransactionNumberingServiceDeps struct {
	Tracing tracing.TracingService
	// In production, add:
	// NumberingRuleRepository domain.NumberingRuleRepository
	// ReservationRepository domain.ReservationRepository
}

// NewTransactionNumberingService creates a new transaction numbering service
func NewTransactionNumberingService(deps TransactionNumberingServiceDeps) TransactionNumberingService {
	service := &transactionNumberingService{
		rules:        make(map[string]*NumberingRule),
		reservations: make(map[string]*NumberReservation),
		tracing:      deps.Tracing,
	}

	// Initialize default rules (in production, these would be loaded from database)
	service.initializeDefaultRules()

	return service
}

// initializeDefaultRules sets up default numbering rules
func (s *transactionNumberingService) initializeDefaultRules() {
	// Default rules for common transaction types
	defaultRules := []*NumberingRule{
		{
			ID:              uuid.New(),
			TransactionType: domain.TransactionTypeManual,
			Series:          "MAIN",
			Prefix:          "TXN",
			DateFormat:      stringPtr("YYYY"),
			NumberLength:    6,
			ResetFrequency:  ResetFrequencyYearly,
			StartingNumber:  1,
			CurrentNumber:   0,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			TransactionType: domain.TransactionTypeJournalEntry,
			Series:          "MAIN",
			Prefix:          "JE",
			DateFormat:      stringPtr("YYYY"),
			NumberLength:    6,
			ResetFrequency:  ResetFrequencyYearly,
			StartingNumber:  1,
			CurrentNumber:   0,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			TransactionType: domain.TransactionTypeAdjustment,
			Series:          "MAIN",
			Prefix:          "ADJ",
			DateFormat:      stringPtr("YYYY"),
			NumberLength:    6,
			ResetFrequency:  ResetFrequencyYearly,
			StartingNumber:  1,
			CurrentNumber:   0,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			TransactionType: domain.TransactionTypeSystem,
			Series:          "MAIN",
			Prefix:          "SYS",
			DateFormat:      stringPtr("YYYYMMDD"),
			NumberLength:    4,
			ResetFrequency:  ResetFrequencyDaily,
			StartingNumber:  1,
			CurrentNumber:   0,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, rule := range defaultRules {
		key := s.getRuleKey(rule.TenantID, rule.EntityID, rule.TransactionType, rule.Series)
		s.rules[key] = rule
	}
}

// GenerateTransactionNumber generates a unique transaction number
func (s *transactionNumberingService) GenerateTransactionNumber(ctx context.Context, req GenerateNumberRequest) (string, error) {
	ctx, span := s.tracing.StartSpan(ctx, "TransactionNumberingService.GenerateTransactionNumber")
	defer span.End()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get or create numbering rule
	rule, err := s.getNumberingRule(req)
	if err != nil {
		return "", fmt.Errorf("failed to get numbering rule: %w", err)
	}

	// Check if sequence needs reset
	if s.shouldResetSequence(rule, req.TransactionDate) {
		rule.CurrentNumber = rule.StartingNumber - 1
		rule.UpdatedAt = time.Now()
	}

	// Increment sequence
	rule.CurrentNumber++
	rule.UpdatedAt = time.Now()

	// Generate the number
	number := s.formatTransactionNumber(rule, req.TransactionDate)

	// Validate uniqueness (in production, check against database)
	if s.isNumberExists(req.TenantID, number) {
		return "", fmt.Errorf("generated number %s already exists", number)
	}

	span.SetAttributes(
		attribute.String("transaction_type", string(req.TransactionType)),
		attribute.String("generated_number", number),
		attribute.Int64("sequence_number", rule.CurrentNumber),
	)

	return number, nil
}

// ReserveTransactionNumber reserves a transaction number for later use
func (s *transactionNumberingService) ReserveTransactionNumber(ctx context.Context, req GenerateNumberRequest) (string, error) {
	ctx, span := s.tracing.StartSpan(ctx, "TransactionNumberingService.ReserveTransactionNumber")
	defer span.End()

	// Generate the number
	number, err := s.GenerateTransactionNumber(ctx, req)
	if err != nil {
		return "", err
	}

	// Create reservation
	// TODO: Extract user ID from context
	userID := uuid.New() // placeholder
	reservation := &NumberReservation{
		ID:                uuid.New(),
		TenantID:          req.TenantID,
		TransactionNumber: number,
		ReservedBy:        userID,
		ReservedAt:        time.Now(),
		ExpiresAt:         time.Now().Add(30 * time.Minute), // 30 minute expiry
		IsConfirmed:       false,
	}

	s.reservations[number] = reservation

	return number, nil
}

// ConfirmTransactionNumber confirms usage of a reserved number
func (s *transactionNumberingService) ConfirmTransactionNumber(ctx context.Context, number string) error {
	ctx, span := s.tracing.StartSpan(ctx, "TransactionNumberingService.ConfirmTransactionNumber")
	defer span.End()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	reservation, exists := s.reservations[number]
	if !exists {
		return fmt.Errorf("reservation not found for number %s", number)
	}

	if reservation.IsConfirmed {
		return fmt.Errorf("number %s is already confirmed", number)
	}

	if time.Now().After(reservation.ExpiresAt) {
		delete(s.reservations, number)
		return fmt.Errorf("reservation for number %s has expired", number)
	}

	// Confirm the reservation
	reservation.IsConfirmed = true
	now := time.Now()
	reservation.ConfirmedAt = &now

	return nil
}

// ReleaseTransactionNumber releases a reserved number back to the pool
func (s *transactionNumberingService) ReleaseTransactionNumber(ctx context.Context, number string) error {
	ctx, span := s.tracing.StartSpan(ctx, "TransactionNumberingService.ReleaseTransactionNumber")
	defer span.End()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	reservation, exists := s.reservations[number]
	if !exists {
		return fmt.Errorf("reservation not found for number %s", number)
	}

	if reservation.IsConfirmed {
		return fmt.Errorf("cannot release confirmed number %s", number)
	}

	// Remove reservation
	delete(s.reservations, number)

	// Decrement the sequence number if this was the last generated number
	// Note: This is a simplified approach. In production, you'd want more sophisticated gap handling

	return nil
}

// GetNextNumber gets the next number without generating (for preview)
func (s *transactionNumberingService) GetNextNumber(ctx context.Context, req GenerateNumberRequest) (string, error) {
	ctx, span := s.tracing.StartSpan(ctx, "TransactionNumberingService.GetNextNumber")
	defer span.End()

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get or create numbering rule
	rule, err := s.getNumberingRule(req)
	if err != nil {
		return "", fmt.Errorf("failed to get numbering rule: %w", err)
	}

	// Calculate next number without incrementing
	nextNumber := rule.CurrentNumber + 1
	if s.shouldResetSequence(rule, req.TransactionDate) {
		nextNumber = rule.StartingNumber
	}

	// Create temporary rule for formatting
	tempRule := *rule
	tempRule.CurrentNumber = nextNumber

	return s.formatTransactionNumber(&tempRule, req.TransactionDate), nil
}

// ValidateTransactionNumber validates a transaction number format
func (s *transactionNumberingService) ValidateTransactionNumber(ctx context.Context, number string, transactionType domain.TransactionType) error {
	ctx, span := s.tracing.StartSpan(ctx, "TransactionNumberingService.ValidateTransactionNumber")
	defer span.End()

	if strings.TrimSpace(number) == "" {
		return fmt.Errorf("transaction number cannot be empty")
	}

	if len(number) > 50 {
		return fmt.Errorf("transaction number cannot exceed 50 characters")
	}

	// Add transaction-type specific validations
	switch transactionType {
	case domain.TransactionTypeManual:
		if !strings.HasPrefix(number, "TXN") && !strings.HasPrefix(number, "MAN") {
			return fmt.Errorf("manual transaction numbers should start with TXN or MAN")
		}
	case domain.TransactionTypeJournalEntry:
		if !strings.HasPrefix(number, "JE") {
			return fmt.Errorf("journal entry numbers should start with JE")
		}
	case domain.TransactionTypeAdjustment:
		if !strings.HasPrefix(number, "ADJ") {
			return fmt.Errorf("adjustment transaction numbers should start with ADJ")
		}
	}

	return nil
}

// Helper methods

func (s *transactionNumberingService) getNumberingRule(req GenerateNumberRequest) (*NumberingRule, error) {
	series := "MAIN"
	if req.Series != nil {
		series = *req.Series
	}

	key := s.getRuleKey(req.TenantID, req.EntityID, req.TransactionType, series)

	if rule, exists := s.rules[key]; exists {
		return rule, nil
	}

	// If no specific rule found, create a default one
	return s.createDefaultRule(req, series), nil
}

func (s *transactionNumberingService) getRuleKey(tenantID uuid.UUID, entityID *uuid.UUID, transactionType domain.TransactionType, series string) string {
	entityStr := "nil"
	if entityID != nil {
		entityStr = entityID.String()
	}
	return fmt.Sprintf("%s:%s:%s:%s", tenantID.String(), entityStr, string(transactionType), series)
}

func (s *transactionNumberingService) createDefaultRule(req GenerateNumberRequest, series string) *NumberingRule {
	prefix := s.getDefaultPrefix(req.TransactionType)

	rule := &NumberingRule{
		ID:              uuid.New(),
		TenantID:        req.TenantID,
		EntityID:        req.EntityID,
		TransactionType: req.TransactionType,
		Series:          series,
		Prefix:          prefix,
		DateFormat:      stringPtr("YYYY"),
		NumberLength:    6,
		ResetFrequency:  ResetFrequencyYearly,
		StartingNumber:  1,
		CurrentNumber:   0,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Store the new rule
	key := s.getRuleKey(req.TenantID, req.EntityID, req.TransactionType, series)
	s.rules[key] = rule

	return rule
}

func (s *transactionNumberingService) getDefaultPrefix(transactionType domain.TransactionType) string {
	switch transactionType {
	case domain.TransactionTypeManual:
		return "TXN"
	case domain.TransactionTypeJournalEntry:
		return "JE"
	case domain.TransactionTypeAdjustment:
		return "ADJ"
	case domain.TransactionTypeSystem:
		return "SYS"
	case domain.TransactionTypeRecurring:
		return "REC"
	case domain.TransactionTypeClosing:
		return "CLS"
	case domain.TransactionTypeOpening:
		return "OPN"
	default:
		return "TXN"
	}
}

func (s *transactionNumberingService) shouldResetSequence(rule *NumberingRule, transactionDate time.Time) bool {
	// This is simplified logic. In production, you'd track the last reset date
	switch rule.ResetFrequency {
	case ResetFrequencyDaily:
		// Check if it's a new day
		return !s.isSameDay(rule.UpdatedAt, transactionDate)
	case ResetFrequencyMonthly:
		// Check if it's a new month
		return !s.isSameMonth(rule.UpdatedAt, transactionDate)
	case ResetFrequencyYearly:
		// Check if it's a new year
		return rule.UpdatedAt.Year() != transactionDate.Year()
	case ResetFrequencyQuarterly:
		// Check if it's a new quarter
		return s.getQuarter(rule.UpdatedAt) != s.getQuarter(transactionDate) || rule.UpdatedAt.Year() != transactionDate.Year()
	default:
		return false
	}
}

func (s *transactionNumberingService) formatTransactionNumber(rule *NumberingRule, transactionDate time.Time) string {
	var parts []string

	// Add prefix
	if rule.Prefix != "" {
		parts = append(parts, rule.Prefix)
	}

	// Add date format
	if rule.DateFormat != nil {
		dateStr := s.formatDate(*rule.DateFormat, transactionDate)
		parts = append(parts, dateStr)
	}

	// Add sequence number with padding
	sequenceStr := fmt.Sprintf("%0*d", rule.NumberLength, rule.CurrentNumber)
	parts = append(parts, sequenceStr)

	// Add suffix
	if rule.Suffix != nil {
		parts = append(parts, *rule.Suffix)
	}

	return strings.Join(parts, "-")
}

func (s *transactionNumberingService) formatDate(format string, date time.Time) string {
	switch format {
	case "YYYY":
		return strconv.Itoa(date.Year())
	case "YY":
		return fmt.Sprintf("%02d", date.Year()%100)
	case "YYYYMM":
		return fmt.Sprintf("%04d%02d", date.Year(), date.Month())
	case "YYYYMMDD":
		return fmt.Sprintf("%04d%02d%02d", date.Year(), date.Month(), date.Day())
	case "MM":
		return fmt.Sprintf("%02d", date.Month())
	case "DD":
		return fmt.Sprintf("%02d", date.Day())
	case "QQ":
		return fmt.Sprintf("Q%d", s.getQuarter(date))
	default:
		return strconv.Itoa(date.Year())
	}
}

func (s *transactionNumberingService) isNumberExists(tenantID uuid.UUID, number string) bool {
	// In production, this would query the database
	// For now, just check reservations
	_, exists := s.reservations[number]
	return exists
}

func (s *transactionNumberingService) isSameDay(date1, date2 time.Time) bool {
	return date1.Year() == date2.Year() && date1.YearDay() == date2.YearDay()
}

func (s *transactionNumberingService) isSameMonth(date1, date2 time.Time) bool {
	return date1.Year() == date2.Year() && date1.Month() == date2.Month()
}

func (s *transactionNumberingService) getQuarter(date time.Time) int {
	month := int(date.Month())
	return ((month - 1) / 3) + 1
}

// Utility function
func stringPtr(s string) *string {
	return &s
}
