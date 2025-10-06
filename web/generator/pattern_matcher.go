package generator

import (
	"fmt"
	"strings"
)

// PatternType defines the type of UI pattern to generate
type PatternType string

const (
	// Basic CRUD patterns
	PatternCRUDTable     PatternType = "crud_table"
	PatternCreateForm    PatternType = "create_form"
	PatternEditForm      PatternType = "edit_form"
	PatternDetailView    PatternType = "detail_view"
	PatternSearchFilter  PatternType = "search_filter"
	
	// Advanced patterns
	PatternHierarchyTree PatternType = "hierarchy_tree"
	PatternKanbanBoard   PatternType = "kanban_board"
	PatternWizardForm    PatternType = "wizard_form"
	PatternDashboard     PatternType = "dashboard"
	PatternReportView    PatternType = "report_view"
	PatternCalendarView  PatternType = "calendar_view"
	PatternCardGrid      PatternType = "card_grid"
	PatternMasterDetail  PatternType = "master_detail"
	
	// Specialized patterns
	PatternUserProfile   PatternType = "user_profile"
	PatternContactCard   PatternType = "contact_card"
	PatternInvoiceForm   PatternType = "invoice_form"
	PatternTimeTracking  PatternType = "time_tracking"
	PatternFileManager   PatternType = "file_manager"
	PatternTaskBoard     PatternType = "task_board"
	PatternChatInterface PatternType = "chat_interface"
	PatternNotifications PatternType = "notifications"
)

// PatternScore represents the fitness score for a pattern
type PatternScore struct {
	Pattern PatternType `json:"pattern"`
	Score   float64     `json:"score"`
	Reasons []string    `json:"reasons"`
}

// PatternMatcher analyzes struct information and recommends optimal UI patterns
type PatternMatcher struct {
	rules []PatternRule
}

// PatternRule defines criteria for pattern matching
type PatternRule struct {
	Pattern   PatternType
	Weight    float64
	Condition func(StructInfo) (bool, string)
}

// NewPatternMatcher creates a new pattern matcher with default rules
func NewPatternMatcher() *PatternMatcher {
	pm := &PatternMatcher{
		rules: []PatternRule{},
	}
	
	pm.initializeDefaultRules()
	return pm
}

// MatchPatterns analyzes a struct and returns ranked UI patterns
func (pm *PatternMatcher) MatchPatterns(structInfo StructInfo) []PatternScore {
	var scores []PatternScore
	
	for _, rule := range pm.rules {
		if matches, reason := rule.Condition(structInfo); matches {
			score := PatternScore{
				Pattern: rule.Pattern,
				Score:   rule.Weight,
				Reasons: []string{reason},
			}
			
			// Apply contextual scoring modifiers
			score.Score *= pm.calculateContextualModifier(structInfo, rule.Pattern)
			
			scores = append(scores, score)
		}
	}
	
	// Sort by score (highest first)
	pm.sortScoresByWeight(scores)
	
	// Merge duplicate patterns and combine reasons
	scores = pm.mergeDuplicatePatterns(scores)
	
	return scores
}

// GetBestPattern returns the highest-scoring pattern
func (pm *PatternMatcher) GetBestPattern(structInfo StructInfo) (PatternType, float64) {
	patterns := pm.MatchPatterns(structInfo)
	if len(patterns) == 0 {
		return PatternCRUDTable, 0.5 // Default fallback
	}
	
	return patterns[0].Pattern, patterns[0].Score
}

// GetRecommendedPatterns returns patterns above a score threshold
func (pm *PatternMatcher) GetRecommendedPatterns(structInfo StructInfo, threshold float64) []PatternScore {
	patterns := pm.MatchPatterns(structInfo)
	var recommended []PatternScore
	
	for _, pattern := range patterns {
		if pattern.Score >= threshold {
			recommended = append(recommended, pattern)
		}
	}
	
	return recommended
}

// initializeDefaultRules sets up the default pattern matching rules
func (pm *PatternMatcher) initializeDefaultRules() {
	// Basic CRUD patterns - these have the highest weights as defaults
	pm.addRule(PatternCRUDTable, 10.0, func(s StructInfo) (bool, string) {
		return s.IsCRUDEntity, "Entity supports CRUD operations"
	})
	
	pm.addRule(PatternCreateForm, 9.0, func(s StructInfo) (bool, string) {
		return s.IsCRUDEntity && !s.IsReadOnly, "Entity supports creation"
	})
	
	pm.addRule(PatternEditForm, 9.0, func(s StructInfo) (bool, string) {
		return s.IsCRUDEntity && !s.IsReadOnly, "Entity supports updates"
	})
	
	pm.addRule(PatternDetailView, 8.0, func(s StructInfo) (bool, string) {
		return s.IsCRUDEntity, "Entity has viewable details"
	})
	
	pm.addRule(PatternSearchFilter, 7.0, func(s StructInfo) (bool, string) {
		return s.IsCRUDEntity && pm.hasFilterableFields(s), "Entity has filterable fields"
	})
	
	// Hierarchy patterns
	pm.addRule(PatternHierarchyTree, 12.0, func(s StructInfo) (bool, string) {
		return s.HasHierarchy, "Entity has hierarchical structure"
	})
	
	// Workflow patterns
	pm.addRule(PatternKanbanBoard, 11.0, func(s StructInfo) (bool, string) {
		return s.HasWorkflow && pm.hasStatusField(s), "Entity has workflow states"
	})
	
	pm.addRule(PatternWizardForm, 10.0, func(s StructInfo) (bool, string) {
		return pm.hasComplexForm(s), "Entity has complex form requirements"
	})
	
	// Dashboard patterns
	pm.addRule(PatternDashboard, 8.0, func(s StructInfo) (bool, string) {
		return pm.hasAggregateFields(s), "Entity suitable for dashboard aggregation"
	})
	
	pm.addRule(PatternReportView, 7.0, func(s StructInfo) (bool, string) {
		return s.IsReadOnly && pm.hasDateFields(s), "Read-only entity with time-based data"
	})
	
	// Calendar patterns
	pm.addRule(PatternCalendarView, 11.0, func(s StructInfo) (bool, string) {
		return pm.hasCalendarFields(s), "Entity has date/time fields suitable for calendar"
	})
	
	// Card patterns
	pm.addRule(PatternCardGrid, 8.0, func(s StructInfo) (bool, string) {
		return pm.hasCardLikeFields(s), "Entity has fields suitable for card display"
	})
	
	// Master-detail patterns
	pm.addRule(PatternMasterDetail, 9.0, func(s StructInfo) (bool, string) {
		return len(s.Fields) > 10, "Entity has many fields suitable for master-detail layout"
	})
	
	// Specialized entity patterns
	pm.addUserProfileRules()
	pm.addContactRules()
	pm.addFinancialRules()
	pm.addProjectRules()
	pm.addCommunicationRules()
}

// addUserProfileRules adds rules for user profile patterns
func (pm *PatternMatcher) addUserProfileRules() {
	pm.addRule(PatternUserProfile, 12.0, func(s StructInfo) (bool, string) {
		return pm.isUserLikeEntity(s), "Entity appears to be user/person related"
	})
}

// addContactRules adds rules for contact card patterns
func (pm *PatternMatcher) addContactRules() {
	pm.addRule(PatternContactCard, 11.0, func(s StructInfo) (bool, string) {
		return pm.isContactLikeEntity(s), "Entity appears to be contact information"
	})
}

// addFinancialRules adds rules for financial patterns
func (pm *PatternMatcher) addFinancialRules() {
	pm.addRule(PatternInvoiceForm, 13.0, func(s StructInfo) (bool, string) {
		return pm.isInvoiceLikeEntity(s), "Entity appears to be invoice/financial document"
	})
}

// addProjectRules adds rules for project management patterns
func (pm *PatternMatcher) addProjectRules() {
	pm.addRule(PatternTaskBoard, 12.0, func(s StructInfo) (bool, string) {
		return pm.isTaskLikeEntity(s), "Entity appears to be task/project related"
	})
	
	pm.addRule(PatternTimeTracking, 11.0, func(s StructInfo) (bool, string) {
		return pm.isTimeTrackingEntity(s), "Entity tracks time or duration"
	})
}

// addCommunicationRules adds rules for communication patterns
func (pm *PatternMatcher) addCommunicationRules() {
	pm.addRule(PatternChatInterface, 10.0, func(s StructInfo) (bool, string) {
		return pm.isMessageLikeEntity(s), "Entity appears to be message/communication related"
	})
	
	pm.addRule(PatternNotifications, 9.0, func(s StructInfo) (bool, string) {
		return pm.isNotificationEntity(s), "Entity appears to be notification related"
	})
}

// Helper methods for pattern detection

// hasFilterableFields checks if entity has fields suitable for filtering
func (pm *PatternMatcher) hasFilterableFields(s StructInfo) bool {
	filterableCount := 0
	for _, field := range s.Fields {
		if pm.isFieldFilterable(field) {
			filterableCount++
		}
	}
	return filterableCount >= 3
}

// hasStatusField checks for workflow status fields
func (pm *PatternMatcher) hasStatusField(s StructInfo) bool {
	return s.HasStatus || pm.hasFieldWithName(s, "status", "state", "stage", "phase")
}

// hasComplexForm determines if entity needs a wizard form
func (pm *PatternMatcher) hasComplexForm(s StructInfo) bool {
	// Complex forms typically have:
	// - Many fields (>8)
	// - Multiple relationships
	// - Validation requirements
	// - File uploads
	
	fieldCount := len(s.Fields)
	relationshipCount := 0
	hasFileFields := false
	
	for _, field := range s.Fields {
		if field.IsRelationship {
			relationshipCount++
		}
		if strings.Contains(strings.ToLower(field.Type), "file") ||
		   strings.Contains(strings.ToLower(field.Name), "file") ||
		   strings.Contains(strings.ToLower(field.Name), "document") ||
		   strings.Contains(strings.ToLower(field.Name), "attachment") {
			hasFileFields = true
		}
	}
	
	return fieldCount > 8 || relationshipCount > 3 || hasFileFields
}

// hasAggregateFields checks for fields suitable for aggregation
func (pm *PatternMatcher) hasAggregateFields(s StructInfo) bool {
	for _, field := range s.Fields {
		if field.IsNumberType {
			return true
		}
	}
	return false
}

// hasDateFields checks for date/time fields
func (pm *PatternMatcher) hasDateFields(s StructInfo) bool {
	for _, field := range s.Fields {
		if field.IsTimeType {
			return true
		}
	}
	return false
}

// hasCalendarFields checks for fields suitable for calendar display
func (pm *PatternMatcher) hasCalendarFields(s StructInfo) bool {
	hasStartDate := false
	hasEndDate := false
	hasTitle := false
	
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		
		if field.IsTimeType {
			if strings.Contains(fieldName, "start") || strings.Contains(fieldName, "begin") {
				hasStartDate = true
			} else if strings.Contains(fieldName, "end") || strings.Contains(fieldName, "finish") {
				hasEndDate = true
			}
		}
		
		if field.IsStringType && (strings.Contains(fieldName, "title") || 
		   strings.Contains(fieldName, "name") || strings.Contains(fieldName, "subject")) {
			hasTitle = true
		}
	}
	
	// End date is optional but adds to calendar completeness
	_ = hasEndDate // Mark as intentionally unused for future enhancements
	return hasStartDate && hasTitle
}

// hasCardLikeFields checks for fields suitable for card display
func (pm *PatternMatcher) hasCardLikeFields(s StructInfo) bool {
	hasTitle := false
	hasDescription := false
	hasImage := false
	
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		
		if field.IsStringType {
			if strings.Contains(fieldName, "title") || strings.Contains(fieldName, "name") {
				hasTitle = true
			} else if strings.Contains(fieldName, "description") || strings.Contains(fieldName, "summary") {
				hasDescription = true
			} else if strings.Contains(fieldName, "image") || strings.Contains(fieldName, "photo") ||
			         strings.Contains(fieldName, "avatar") || strings.Contains(fieldName, "picture") {
				hasImage = true
			}
		}
	}
	
	return hasTitle && (hasDescription || hasImage)
}

// Entity type detection methods

// isUserLikeEntity detects user/person entities
func (pm *PatternMatcher) isUserLikeEntity(s StructInfo) bool {
	entityName := strings.ToLower(s.Name)
	userKeywords := []string{"user", "person", "employee", "member", "account", "profile"}
	
	for _, keyword := range userKeywords {
		if strings.Contains(entityName, keyword) {
			return true
		}
	}
	
	// Check for typical user fields
	userFields := []string{"email", "password", "username", "firstname", "lastname", "name"}
	matchCount := 0
	
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		for _, userField := range userFields {
			if strings.Contains(fieldName, userField) {
				matchCount++
				break
			}
		}
	}
	
	return matchCount >= 2
}

// isContactLikeEntity detects contact information entities
func (pm *PatternMatcher) isContactLikeEntity(s StructInfo) bool {
	entityName := strings.ToLower(s.Name)
	contactKeywords := []string{"contact", "customer", "client", "vendor", "supplier", "lead"}
	
	for _, keyword := range contactKeywords {
		if strings.Contains(entityName, keyword) {
			return true
		}
	}
	
	// Check for contact fields
	contactFields := []string{"email", "phone", "address", "company", "organization"}
	matchCount := 0
	
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		for _, contactField := range contactFields {
			if strings.Contains(fieldName, contactField) {
				matchCount++
				break
			}
		}
	}
	
	return matchCount >= 2
}

// isInvoiceLikeEntity detects financial document entities
func (pm *PatternMatcher) isInvoiceLikeEntity(s StructInfo) bool {
	entityName := strings.ToLower(s.Name)
	financialKeywords := []string{"invoice", "bill", "payment", "order", "transaction", "receipt"}
	
	for _, keyword := range financialKeywords {
		if strings.Contains(entityName, keyword) {
			return true
		}
	}
	
	// Check for financial fields
	financialFields := []string{"amount", "total", "subtotal", "tax", "discount", "price", "cost"}
	matchCount := 0
	
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		for _, finField := range financialFields {
			if strings.Contains(fieldName, finField) {
				matchCount++
				break
			}
		}
	}
	
	return matchCount >= 2
}

// isTaskLikeEntity detects task/project entities
func (pm *PatternMatcher) isTaskLikeEntity(s StructInfo) bool {
	entityName := strings.ToLower(s.Name)
	taskKeywords := []string{"task", "project", "issue", "ticket", "todo", "work", "job"}
	
	for _, keyword := range taskKeywords {
		if strings.Contains(entityName, keyword) {
			return true
		}
	}
	
	// Check for task fields
	taskFields := []string{"status", "priority", "assignee", "deadline", "duedate", "progress"}
	matchCount := 0
	
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		for _, taskField := range taskFields {
			if strings.Contains(fieldName, taskField) {
				matchCount++
				break
			}
		}
	}
	
	return matchCount >= 2
}

// isTimeTrackingEntity detects time tracking entities
func (pm *PatternMatcher) isTimeTrackingEntity(s StructInfo) bool {
	entityName := strings.ToLower(s.Name)
	timeKeywords := []string{"timesheet", "timeentry", "log", "tracking", "hours", "duration"}
	
	for _, keyword := range timeKeywords {
		if strings.Contains(entityName, keyword) {
			return true
		}
	}
	
	// Check for time fields
	timeFields := []string{"starttime", "endtime", "duration", "hours", "minutes", "logged"}
	matchCount := 0
	
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		for _, timeField := range timeFields {
			if strings.Contains(fieldName, timeField) {
				matchCount++
				break
			}
		}
	}
	
	return matchCount >= 2
}

// isMessageLikeEntity detects message/communication entities
func (pm *PatternMatcher) isMessageLikeEntity(s StructInfo) bool {
	entityName := strings.ToLower(s.Name)
	messageKeywords := []string{"message", "chat", "conversation", "comment", "post", "reply"}
	
	for _, keyword := range messageKeywords {
		if strings.Contains(entityName, keyword) {
			return true
		}
	}
	
	// Check for message fields
	messageFields := []string{"content", "body", "text", "sender", "recipient", "timestamp"}
	matchCount := 0
	
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		for _, msgField := range messageFields {
			if strings.Contains(fieldName, msgField) {
				matchCount++
				break
			}
		}
	}
	
	return matchCount >= 2
}

// isNotificationEntity detects notification entities
func (pm *PatternMatcher) isNotificationEntity(s StructInfo) bool {
	entityName := strings.ToLower(s.Name)
	notificationKeywords := []string{"notification", "alert", "reminder", "notice", "announcement"}
	
	for _, keyword := range notificationKeywords {
		if strings.Contains(entityName, keyword) {
			return true
		}
	}
	
	// Check for notification fields
	notificationFields := []string{"read", "seen", "dismissed", "priority", "type", "recipient"}
	matchCount := 0
	
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		for _, notifField := range notificationFields {
			if strings.Contains(fieldName, notifField) {
				matchCount++
				break
			}
		}
	}
	
	return matchCount >= 2
}

// Utility methods

// isFieldFilterable determines if a field can be used for filtering
func (pm *PatternMatcher) isFieldFilterable(field FieldInfo) bool {
	return field.IsStringType || field.IsBoolType || field.IsTimeType || 
	       field.IsRelationship || len(field.Options) > 0 || field.IsNumberType
}

// hasFieldWithName checks if entity has fields with specific names
func (pm *PatternMatcher) hasFieldWithName(s StructInfo, names ...string) bool {
	for _, field := range s.Fields {
		fieldName := strings.ToLower(field.Name)
		for _, name := range names {
			if strings.Contains(fieldName, strings.ToLower(name)) {
				return true
			}
		}
	}
	return false
}

// calculateContextualModifier applies contextual scoring based on entity characteristics
func (pm *PatternMatcher) calculateContextualModifier(s StructInfo, pattern PatternType) float64 {
	modifier := 1.0
	
	// Boost patterns for entities with many relationships
	relationshipCount := 0
	for _, field := range s.Fields {
		if field.IsRelationship {
			relationshipCount++
		}
	}
	
	if relationshipCount > 3 {
		switch pattern {
		case PatternMasterDetail, PatternDetailView:
			modifier += 0.2
		case PatternCRUDTable:
			modifier += 0.1
		}
	}
	
	// Boost workflow patterns for entities with audit trails
	if s.HasAuditTrail {
		switch pattern {
		case PatternKanbanBoard, PatternTaskBoard:
			modifier += 0.15
		}
	}
	
	// Reduce form patterns for read-only entities
	if s.IsReadOnly {
		switch pattern {
		case PatternCreateForm, PatternEditForm, PatternWizardForm:
			modifier -= 0.5
		case PatternReportView, PatternDetailView:
			modifier += 0.2
		}
	}
	
	// Boost hierarchy patterns for tenant-scoped entities
	if s.HasTenantScope {
		switch pattern {
		case PatternHierarchyTree:
			modifier += 0.1
		}
	}
	
	return modifier
}

// Helper methods for rule management

// addRule adds a new pattern matching rule
func (pm *PatternMatcher) addRule(pattern PatternType, weight float64, condition func(StructInfo) (bool, string)) {
	pm.rules = append(pm.rules, PatternRule{
		Pattern:   pattern,
		Weight:    weight,
		Condition: condition,
	})
}

// sortScoresByWeight sorts pattern scores by weight (descending)
func (pm *PatternMatcher) sortScoresByWeight(scores []PatternScore) {
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].Score < scores[j].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
}

// mergeDuplicatePatterns combines scores for the same pattern
func (pm *PatternMatcher) mergeDuplicatePatterns(scores []PatternScore) []PatternScore {
	patternMap := make(map[PatternType]*PatternScore)
	
	for _, score := range scores {
		if existing, exists := patternMap[score.Pattern]; exists {
			existing.Score += score.Score * 0.1 // Additional matches add 10% bonus
			existing.Reasons = append(existing.Reasons, score.Reasons...)
		} else {
			scoreCopy := score
			patternMap[score.Pattern] = &scoreCopy
		}
	}
	
	var merged []PatternScore
	for _, score := range patternMap {
		merged = append(merged, *score)
	}
	
	pm.sortScoresByWeight(merged)
	return merged
}

// GetPatternDescription returns a human-readable description of a pattern
func (pm *PatternMatcher) GetPatternDescription(pattern PatternType) string {
	descriptions := map[PatternType]string{
		PatternCRUDTable:     "Standard data table with create, read, update, delete operations",
		PatternCreateForm:    "Form interface for creating new records",
		PatternEditForm:     "Form interface for editing existing records",
		PatternDetailView:   "Detailed view of a single record with related information",
		PatternSearchFilter: "Advanced filtering and search interface",
		PatternHierarchyTree: "Tree structure for hierarchical data navigation",
		PatternKanbanBoard:   "Kanban-style board for workflow management",
		PatternWizardForm:    "Multi-step form for complex data entry",
		PatternDashboard:     "Overview dashboard with key metrics and summaries",
		PatternReportView:    "Read-only report interface with data visualization",
		PatternCalendarView:  "Calendar interface for date-based data",
		PatternCardGrid:      "Grid of cards showing summary information",
		PatternMasterDetail:  "Split view with list and detailed information",
		PatternUserProfile:   "User profile management interface",
		PatternContactCard:   "Contact information display and editing",
		PatternInvoiceForm:   "Financial document creation and management",
		PatternTimeTracking:  "Time logging and tracking interface",
		PatternFileManager:   "File upload and management interface",
		PatternTaskBoard:     "Task and project management board",
		PatternChatInterface: "Real-time communication interface",
		PatternNotifications: "Notification center and management",
	}
	
	if desc, exists := descriptions[pattern]; exists {
		return desc
	}
	
	return fmt.Sprintf("Custom pattern: %s", string(pattern))
}

// ValidatePattern checks if a pattern is valid for the given struct
func (pm *PatternMatcher) ValidatePattern(structInfo StructInfo, pattern PatternType) (bool, []string) {
	var issues []string
	
	switch pattern {
	case PatternCRUDTable:
		if !structInfo.IsCRUDEntity {
			issues = append(issues, "Entity does not support CRUD operations")
		}
		if len(structInfo.Fields) == 0 {
			issues = append(issues, "Entity has no displayable fields")
		}
		
	case PatternCreateForm, PatternEditForm:
		if structInfo.IsReadOnly {
			issues = append(issues, "Entity is read-only and cannot be modified")
		}
		if !structInfo.IsCRUDEntity {
			issues = append(issues, "Entity does not support CRUD operations")
		}
		
	case PatternHierarchyTree:
		if !structInfo.HasHierarchy {
			issues = append(issues, "Entity does not have hierarchical structure")
		}
		
	case PatternKanbanBoard:
		if !structInfo.HasWorkflow {
			issues = append(issues, "Entity does not have workflow states")
		}
		if !pm.hasStatusField(structInfo) {
			issues = append(issues, "Entity does not have status field for workflow")
		}
		
	case PatternCalendarView:
		if !pm.hasCalendarFields(structInfo) {
			issues = append(issues, "Entity does not have date/time fields suitable for calendar")
		}
	}
	
	return len(issues) == 0, issues
}

// TODO: Future enhancements for pattern matching:
// 1. Machine learning-based pattern recommendation
// 2. User preference learning and adaptation
// 3. Context-aware pattern suggestions based on user role
// 4. A/B testing framework for pattern effectiveness
// 5. Custom pattern definition and registration system
// 6. Performance-based pattern optimization
// 7. Integration with analytics for pattern usage tracking

// NOTE: Design considerations:
// 1. Pattern matching uses a scoring system for flexibility
// 2. Rules are weighted to prioritize more specific patterns
// 3. Contextual modifiers adapt scoring based on entity characteristics
// 4. Entity type detection uses both naming conventions and field analysis
// 5. The system is extensible for custom pattern registration