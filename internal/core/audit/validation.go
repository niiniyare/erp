package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Validation constants
const (
	MaxEventTypeLength     = 100
	MaxEventCategoryLength = 100
	MaxSeverityLength      = 20
	MaxDecisionLength      = 50
	MaxReasonLength        = 500
	MaxUserAgentLength     = 1000
	MaxContextSize         = 10000 // bytes
	MaxComplianceFlagsSize = 5000  // bytes
	MinRiskScore           = 0
	MaxRiskScore           = 100
)

// Valid enum values — must match DB CHECK constraints in 000450_audit_log.up.sql exactly.
var (
	ValidSeverities = map[string]bool{
		"LOW":      true,
		"INFO":     true,
		"WARN":     true,
		"HIGH":     true,
		"CRITICAL": true,
	}

	ValidDecisions = map[string]bool{
		"ALLOW": true,
		"DENY":  true,
		"WARN":  true,
	}

	// DB CHECK: event_category IN ('ACCESS','ADMIN','DATA','AUTH','SYSTEM','COMPLIANCE')
	ValidEventCategories = map[string]bool{
		"ACCESS":     true, // authorization / access control events
		"ADMIN":      true, // administrative / configuration changes
		"DATA":       true, // data access events
		"AUTH":       true, // authentication events (login, logout, MFA)
		"SYSTEM":     true, // system-level events
		"COMPLIANCE": true, // regulatory compliance events
	}
)

// ValidationError represents a validation error with field context
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// Validate validates an AuditEvent
func (ae *AuditEvent) Validate() error {
	var errors ValidationErrors

	// Required fields
	if ae.ID == uuid.Nil {
		errors = append(errors, ValidationError{"ID", "cannot be nil"})
	}
	if ae.UserID == uuid.Nil {
		errors = append(errors, ValidationError{"UserID", "cannot be nil"})
	}
	if ae.CreatedAt.IsZero() {
		errors = append(errors, ValidationError{"CreatedAt", "cannot be zero"})
	}

	// Validate event type
	if err := validateEventType(ae.EventType); err != nil {
		errors = append(errors, ValidationError{"EventType", err.Error()})
	}

	// Validate event category
	if err := validateEventCategory(ae.EventCategory); err != nil {
		errors = append(errors, ValidationError{"EventCategory", err.Error()})
	}

	// Validate severity
	if err := validateSeverity(ae.Severity); err != nil {
		errors = append(errors, ValidationError{"Severity", err.Error()})
	}

	// Validate optional fields
	if ae.Decision != nil {
		if err := validateDecision(*ae.Decision); err != nil {
			errors = append(errors, ValidationError{"Decision", err.Error()})
		}
	}

	if ae.Reason != nil {
		if err := validateReason(*ae.Reason); err != nil {
			errors = append(errors, ValidationError{"Reason", err.Error()})
		}
	}

	if ae.RiskScore != nil {
		if err := validateRiskScore(*ae.RiskScore); err != nil {
			errors = append(errors, ValidationError{"RiskScore", err.Error()})
		}
	}

	if ae.IPAddress != nil {
		if err := validateIPAddress(*ae.IPAddress); err != nil {
			errors = append(errors, ValidationError{"IPAddress", err.Error()})
		}
	}

	if ae.UserAgent != nil {
		if err := validateUserAgent(*ae.UserAgent); err != nil {
			errors = append(errors, ValidationError{"UserAgent", err.Error()})
		}
	}

	if err := validateJSONSize(ae.Context, "Context", MaxContextSize); err != nil {
		errors = append(errors, ValidationError{"Context", err.Error()})
	}

	if err := validateJSONSize(ae.ComplianceFlags, "ComplianceFlags", MaxComplianceFlagsSize); err != nil {
		errors = append(errors, ValidationError{"ComplianceFlags", err.Error()})
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// Validate validates a CreateAuditEventRequest
func (req *CreateAuditEventRequest) Validate() error {
	var errors ValidationErrors

	// Required fields
	if err := validateEventType(req.EventType); err != nil {
		errors = append(errors, ValidationError{"EventType", err.Error()})
	}

	if err := validateEventCategory(req.EventCategory); err != nil {
		errors = append(errors, ValidationError{"EventCategory", err.Error()})
	}

	if err := validateSeverity(req.Severity); err != nil {
		errors = append(errors, ValidationError{"Severity", err.Error()})
	}

	// Validate optional fields
	if req.Decision != nil {
		if err := validateDecision(*req.Decision); err != nil {
			errors = append(errors, ValidationError{"Decision", err.Error()})
		}
	}

	if req.Reason != nil {
		if err := validateReason(*req.Reason); err != nil {
			errors = append(errors, ValidationError{"Reason", err.Error()})
		}
	}

	if req.RiskScore != nil {
		if err := validateRiskScore(*req.RiskScore); err != nil {
			errors = append(errors, ValidationError{"RiskScore", err.Error()})
		}
	}

	if req.IPAddress != nil {
		if err := validateIPAddress(*req.IPAddress); err != nil {
			errors = append(errors, ValidationError{"IPAddress", err.Error()})
		}
	}

	if req.UserAgent != nil {
		if err := validateUserAgent(*req.UserAgent); err != nil {
			errors = append(errors, ValidationError{"UserAgent", err.Error()})
		}
	}

	if err := validateJSONSize(req.Context, "Context", MaxContextSize); err != nil {
		errors = append(errors, ValidationError{"Context", err.Error()})
	}

	if err := validateJSONSize(req.ComplianceFlags, "ComplianceFlags", MaxComplianceFlagsSize); err != nil {
		errors = append(errors, ValidationError{"ComplianceFlags", err.Error()})
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// Validate validates AuditEventFilters
func (f *AuditEventFilters) Validate() error {
	var errors ValidationErrors

	if f.Limit < 0 {
		errors = append(errors, ValidationError{"Limit", "cannot be negative"})
	}
	if f.Limit > 1000 {
		errors = append(errors, ValidationError{"Limit", "cannot exceed 1000"})
	}
	if f.Offset < 0 {
		errors = append(errors, ValidationError{"Offset", "cannot be negative"})
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// Validate validates SimilarIncidentPatternsParams
func (p *SimilarIncidentPatternsParams) Validate() error {
	var errors ValidationErrors

	if p.SearchStart.IsZero() {
		errors = append(errors, ValidationError{"SearchStart", "cannot be zero"})
	}
	if p.SearchEnd.IsZero() {
		errors = append(errors, ValidationError{"SearchEnd", "cannot be zero"})
	}
	if p.SearchEnd.Before(p.SearchStart) {
		errors = append(errors, ValidationError{"SearchEnd", "cannot be before SearchStart"})
	}

	if p.PatternStart.IsZero() {
		errors = append(errors, ValidationError{"PatternStart", "cannot be zero"})
	}
	if p.PatternEnd.IsZero() {
		errors = append(errors, ValidationError{"PatternEnd", "cannot be zero"})
	}
	if p.PatternEnd.Before(p.PatternStart) {
		errors = append(errors, ValidationError{"PatternEnd", "cannot be before PatternStart"})
	}

	if p.MaxRiskDeviation < 0 {
		errors = append(errors, ValidationError{"MaxRiskDeviation", "cannot be negative"})
	}
	if p.Limit <= 0 {
		errors = append(errors, ValidationError{"Limit", "must be positive"})
	}
	if p.Limit > 1000 {
		errors = append(errors, ValidationError{"Limit", "cannot exceed 1000"})
	}

	if err := validateEventType(p.IncidentEventType); err != nil {
		errors = append(errors, ValidationError{"IncidentEventType", err.Error()})
	}

	if p.IncidentEventCategory != nil {
		if err := validateEventCategory(*p.IncidentEventCategory); err != nil {
			errors = append(errors, ValidationError{"IncidentEventCategory", err.Error()})
		}
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// Validate validates AnomalousUserBehaviorParams
func (p *AnomalousUserBehaviorParams) Validate() error {
	var errors ValidationErrors

	if p.StartTime.IsZero() {
		errors = append(errors, ValidationError{"StartTime", "cannot be zero"})
	}
	if p.EndTime.IsZero() {
		errors = append(errors, ValidationError{"EndTime", "cannot be zero"})
	}
	if p.EndTime.Before(p.StartTime) {
		errors = append(errors, ValidationError{"EndTime", "cannot be before StartTime"})
	}

	if p.BaselineStart.IsZero() {
		errors = append(errors, ValidationError{"BaselineStart", "cannot be zero"})
	}
	if p.BaselineEnd.IsZero() {
		errors = append(errors, ValidationError{"BaselineEnd", "cannot be zero"})
	}
	if p.BaselineEnd.Before(p.BaselineStart) {
		errors = append(errors, ValidationError{"BaselineEnd", "cannot be before BaselineStart"})
	}

	if p.Limit <= 0 {
		errors = append(errors, ValidationError{"Limit", "must be positive"})
	}
	if p.Limit > 1000 {
		errors = append(errors, ValidationError{"Limit", "cannot exceed 1000"})
	}

	if p.MinBaselineEvents < 0 {
		errors = append(errors, ValidationError{"MinBaselineEvents", "cannot be negative"})
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// Validate validates BulkUpdateRiskScoresParams
func (p *BulkUpdateRiskScoresParams) Validate() error {
	var errors ValidationErrors

	if err := validateRiskScore(p.NewRiskScore); err != nil {
		errors = append(errors, ValidationError{"NewRiskScore", err.Error()})
	}

	if err := validateEventType(p.EventType); err != nil {
		errors = append(errors, ValidationError{"EventType", err.Error()})
	}

	if p.StartTime.IsZero() {
		errors = append(errors, ValidationError{"StartTime", "cannot be zero"})
	}
	if p.EndTime.IsZero() {
		errors = append(errors, ValidationError{"EndTime", "cannot be zero"})
	}
	if p.EndTime.Before(p.StartTime) {
		errors = append(errors, ValidationError{"EndTime", "cannot be before StartTime"})
	}

	if p.CurrentRiskScore != nil {
		if err := validateRiskScore(*p.CurrentRiskScore); err != nil {
			errors = append(errors, ValidationError{"CurrentRiskScore", err.Error()})
		}
	}

	if p.EventCategory != nil {
		if err := validateEventCategory(*p.EventCategory); err != nil {
			errors = append(errors, ValidationError{"EventCategory", err.Error()})
		}
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// Validate validates BulkAddComplianceFlagsParams
func (p *BulkAddComplianceFlagsParams) Validate() error {
	var errors ValidationErrors

	if err := validateJSONSize(p.NewComplianceFlags, "NewComplianceFlags", MaxComplianceFlagsSize); err != nil {
		errors = append(errors, ValidationError{"NewComplianceFlags", err.Error()})
	}

	if err := validateEventType(p.EventType); err != nil {
		errors = append(errors, ValidationError{"EventType", err.Error()})
	}

	if p.StartTime.IsZero() {
		errors = append(errors, ValidationError{"StartTime", "cannot be zero"})
	}
	if p.EndTime.IsZero() {
		errors = append(errors, ValidationError{"EndTime", "cannot be zero"})
	}
	if p.EndTime.Before(p.StartTime) {
		errors = append(errors, ValidationError{"EndTime", "cannot be before StartTime"})
	}

	if p.EventCategory != nil {
		if err := validateEventCategory(*p.EventCategory); err != nil {
			errors = append(errors, ValidationError{"EventCategory", err.Error()})
		}
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// Helper validation functions

func validateEventType(eventType string) error {
	if eventType == "" {
		return errors.New("cannot be empty")
	}
	if len(eventType) > MaxEventTypeLength {
		return fmt.Errorf("cannot exceed %d characters", MaxEventTypeLength)
	}
	// Event type should contain only alphanumeric characters, underscores, and dots
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.]+$`, eventType)
	if !matched {
		return errors.New("can only contain alphanumeric characters, underscores, and dots")
	}
	return nil
}

func validateEventCategory(category string) error {
	if category == "" {
		return errors.New("cannot be empty")
	}
	if len(category) > MaxEventCategoryLength {
		return fmt.Errorf("cannot exceed %d characters", MaxEventCategoryLength)
	}
	if !ValidEventCategories[category] {
		return fmt.Errorf("must be one of: %v", getKeys(ValidEventCategories))
	}
	return nil
}

func validateSeverity(severity string) error {
	if severity == "" {
		return errors.New("cannot be empty")
	}
	if len(severity) > MaxSeverityLength {
		return fmt.Errorf("cannot exceed %d characters", MaxSeverityLength)
	}
	if !ValidSeverities[severity] {
		return fmt.Errorf("must be one of: %v", getKeys(ValidSeverities))
	}
	return nil
}

func validateDecision(decision string) error {
	if decision == "" {
		return errors.New("cannot be empty")
	}
	if len(decision) > MaxDecisionLength {
		return fmt.Errorf("cannot exceed %d characters", MaxDecisionLength)
	}
	if !ValidDecisions[decision] {
		return fmt.Errorf("must be one of: %v", getKeys(ValidDecisions))
	}
	return nil
}

func validateReason(reason string) error {
	if len(reason) > MaxReasonLength {
		return fmt.Errorf("cannot exceed %d characters", MaxReasonLength)
	}
	return nil
}

func validateRiskScore(score int) error {
	if score < MinRiskScore || score > MaxRiskScore {
		return fmt.Errorf("must be between %d and %d", MinRiskScore, MaxRiskScore)
	}
	return nil
}

func validateIPAddress(ip string) error {
	if ip == "" {
		return errors.New("cannot be empty")
	}
	_, err := netip.ParseAddr(ip)
	if err != nil {
		return fmt.Errorf("invalid IP address format: %v", err)
	}
	return nil
}

func validateUserAgent(userAgent string) error {
	if len(userAgent) > MaxUserAgentLength {
		return fmt.Errorf("cannot exceed %d characters", MaxUserAgentLength)
	}
	return nil
}

func validateJSONSize(data json.RawMessage, fieldName string, maxSize int) error {
	if data == nil {
		return nil
	}
	if len(data) > maxSize {
		return fmt.Errorf("%s JSON data cannot exceed %d bytes", fieldName, maxSize)
	}
	// Validate that it's valid JSON
	if !json.Valid(data) {
		return fmt.Errorf("%s contains invalid JSON", fieldName)
	}
	return nil
}

// Helper function to get keys from a map
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ValidateTimeRange validates that start time is before end time and both are not zero
func ValidateTimeRange(start, end time.Time, fieldPrefix string) error {
	if start.IsZero() {
		return fmt.Errorf("%sStart cannot be zero", fieldPrefix)
	}
	if end.IsZero() {
		return fmt.Errorf("%sEnd cannot be zero", fieldPrefix)
	}
	if end.Before(start) {
		return fmt.Errorf("%sEnd cannot be before %sStart", fieldPrefix, fieldPrefix)
	}
	return nil
}

// ValidateUUID validates that a UUID pointer is not nil and not empty
func ValidateUUID(id *uuid.UUID, fieldName string) error {
	if id == nil {
		return fmt.Errorf("%s cannot be nil", fieldName)
	}
	if *id == uuid.Nil {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}

// ValidateUUIDRequired validates that a UUID is not empty
func ValidateUUIDRequired(id uuid.UUID, fieldName string) error {
	if id == uuid.Nil {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}

// ValidateLimit validates pagination limits
func ValidateLimit(limit int, maxLimit int) error {
	if limit < 0 {
		return errors.New("limit cannot be negative")
	}
	if limit > maxLimit {
		return fmt.Errorf("limit cannot exceed %d", maxLimit)
	}
	return nil
}

// ValidateOffset validates pagination offset
func ValidateOffset(offset int) error {
	if offset < 0 {
		return errors.New("offset cannot be negative")
	}
	return nil
}
