package common

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"os"
// 	"regexp"
// 	"strconv"
// 	"strings"
// 	"time"
// )
//
// // Config represents database configuration
// type Config struct {
// 	DSN         string `json:"dsn"`
// 	Host        string `json:"host"`
// 	Port        int    `json:"port"`
// 	Database    string `json:"database"`
// 	Username    string `json:"username"`
// 	Password    string `json:"password"`
// 	CountryCode string `json:"country_code"`
// }
//
// // Form represents form data structure
// type Form struct {
// 	ID                 int                      `json:"id"`
// 	Direction          string                   `json:"direction"`
// 	ALL                []map[string]interface{} `json:"all"`
// 	Login              string                   `json:"login"`
// 	Reference          string                   `json:"reference"`
// 	Description        string                   `json:"description"`
// 	Notes              string                   `json:"notes"`
// 	TransDate          string                   `json:"transdate"`
// 	Department         string                   `json:"department"`
// 	RowCount           int                      `json:"rowcount"`
// 	ClosedTo           string                   `json:"closedto"`
// 	DB                 string                   `json:"db"`
// 	Fld                string                   `json:"fld"`
// 	Move               string                   `json:"move"`
// 	RemoveAuditTrail   string                   `json:"removeaudittrail"`
// 	RevTrans           bool                     `json:"revtrans"`
// 	AuditTrail         bool                     `json:"audittrail"`
// 	ExtendedLog        bool                     `json:"extendedlog"`
// 	Method             string                   `json:"method"`
// 	Precision          int                      `json:"precision"`
// 	Chart              []map[string]interface{} `json:"chart"`
//
// 	// Dynamic fields for account entries and other form data
// 	Fields map[string]interface{} `json:"fields"`
// }
//
// // AccountingService handles accounting operations
// type AccountingService struct {
// 	db     *sql.DB
// 	logger *log.Logger
// }
//
// // NewAccountingService creates a new accounting service instance
// func NewAccountingService(db *sql.DB) *AccountingService {
// 	return &AccountingService{
// 		db:     db,
// 		logger: log.New(os.Stdout, "[ACCOUNTING] ", log.LstdFlags),
// 	}
// }
//
// // GetDB returns the database connection for external use
// func (s *AccountingService) GetDB() *sql.DB {
// 	return s.db
// }
//
// // ConnectDB establishes database connection based on config
// func ConnectDB(config *Config) (*sql.DB, error) {
// 	// This would typically use a specific driver like postgres, mysql, etc.
// 	// For this example, we'll use a generic approach
// 	db, err := sql.Open("postgres", config.DSN) // or config.BuildDSN()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to database: %w", err)
// 	}
//
// 	if err := db.Ping(); err != nil {
// 		return nil, fmt.Errorf("failed to ping database: %w", err)
// 	}
//
// 	return db, nil
// }
//
// // BuildDSN constructs database connection string from config
// func (c *Config) BuildDSN() string {
// 	if c.DSN != "" {
// 		return c.DSN
// 	}
// 	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
// 		c.Host, c.Port, c.Username, c.Password, c.Database)
// }
//
// // Dispatch retrieves dispatch records ordered by ID
// func (s *AccountingService) Dispatch(config *Config, form *Form) error {
// 	// Validate direction parameter
// 	direction := "ASC"
// 	if strings.ToUpper(form.Direction) == "DESC" {
// 		direction = "DESC"
// 	}
//
// 	// Sort order is handled by form.Direction (ASC/DESC)
// 	query := fmt.Sprintf(`
// 		SELECT id, description
// 		FROM dispatch
// 		ORDER BY 1 %s`, direction)
//
// 	s.logger.Printf("Executing dispatch query with direction: %s", direction)
//
// 	rows, err := s.db.Query(query)
// 	if err != nil {
// 		s.logger.Printf("Dispatch query failed: %v", err)
// 		return fmt.Errorf("dispatch query failed: %w", err)
// 	}
// 	defer rows.Close()
//
// 	// Clear existing results
// 	form.ALL = make([]map[string]interface{}, 0)
//
// 	// Fetch all rows and store in form.ALL
// 	for rows.Next() {
// 		var id int
// 		var description string
//
// 		if err := rows.Scan(&id, &description); err != nil {
// 			s.logger.Printf("Failed to scan dispatch row: %v", err)
// 			return fmt.Errorf("failed to scan dispatch row: %w", err)
// 		}
//
// 		record := map[string]interface{}{
// 			"id":          id,
// 			"description": description,
// 		}
// 		form.ALL = append(form.ALL, record)
// 	}
//
// 	s.logger.Printf("Retrieved %d dispatch records", len(form.ALL))
// 	return rows.Err()
// }
//
// // ClosedTo retrieves system defaults for closed period settings
// func (s *AccountingService) ClosedTo(config *Config, form *Form) error {
// 	// Get default values for closedto, revtrans, audittrail, extendedlog
// 	defaults := []string{"closedto", "revtrans", "audittrail", "extendedlog"}
//
// 	query := `
// 		SELECT fldname, fldvalue
// 		FROM defaults
// 		WHERE fldname IN (?, ?, ?, ?)`
//
// 	s.logger.Printf("Retrieving closed period defaults")
//
// 	rows, err := s.db.Query(query, defaults[0], defaults[1], defaults[2], defaults[3])
// 	if err != nil {
// 		s.logger.Printf("ClosedTo query failed: %v", err)
// 		return fmt.Errorf("closedto query failed: %w", err)
// 	}
// 	defer rows.Close()
//
// 	// Initialize fields map if needed
// 	if form.Fields == nil {
// 		form.Fields = make(map[string]interface{})
// 	}
//
// 	// Store retrieved defaults in form
// 	for rows.Next() {
// 		var fieldName, fieldValue string
// 		if err := rows.Scan(&fieldName, &fieldValue); err != nil {
// 			s.logger.Printf("Failed to scan defaults: %v", err)
// 			return fmt.Errorf("failed to scan defaults: %w", err)
// 		}
// 		form.Fields[fieldName] = fieldValue
//
// 		// Also set specific form fields for convenience
// 		switch fieldName {
// 		case "closedto":
// 			form.ClosedTo = fieldValue
// 		case "revtrans":
// 			form.RevTrans = fieldValue == "1" || strings.ToLower(fieldValue) == "true"
// 		case "audittrail":
// 			form.AuditTrail = fieldValue == "1" || strings.ToLower(fieldValue) == "true"
// 		case "extendedlog":
// 			form.ExtendedLog = fieldValue == "1" || strings.ToLower(fieldValue) == "true"
// 		}
// 	}
//
// 	s.logger.Printf("Retrieved defaults for: %v", defaults)
// 	return rows.Err()
// }
//
// // CloseBooks closes accounting books for a period
// func (s *AccountingService) CloseBooks(config *Config, form *Form) error {
// 	s.logger.Printf("Starting close books process for period: %s", form.ClosedTo)
//
// 	// Begin transaction
// 	tx, err := s.db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer func() {
// 		if err != nil {
// 			s.logger.Printf("Rolling back close books transaction: %v", err)
// 			tx.Rollback()
// 		}
// 	}()
//
// 	// Prepare delete statement for existing defaults
// 	deleteQuery := `DELETE FROM defaults WHERE fldname = ?`
// 	deleteStmt, err := tx.Prepare(deleteQuery)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare delete statement: %w", err)
// 	}
// 	defer deleteStmt.Close()
//
// 	// Prepare insert statement for new defaults
// 	insertQuery := `INSERT INTO defaults (fldname, fldvalue) VALUES (?, ?)`
// 	insertStmt, err := tx.Prepare(insertQuery)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare insert statement: %w", err)
// 	}
// 	defer insertStmt.Close()
//
// 	// Convert and validate closedto date using datetonum equivalent
// 	closedToDate := s.dateToNum(config, form.ClosedTo)
// 	if err := s.validateDate(closedToDate); err != nil {
// 		return fmt.Errorf("invalid closedto date: %w", err)
// 	}
//
// 	// Process each default field
// 	fields := []string{"revtrans", "closedto", "audittrail", "extendedlog"}
// 	for _, field := range fields {
// 		// Delete existing entry
// 		if _, err := deleteStmt.Exec(field); err != nil {
// 			return fmt.Errorf("failed to delete default %s: %w", field, err)
// 		}
//
// 		// Insert new value if it exists in form
// 		var value interface{}
// 		switch field {
// 		case "closedto":
// 			value = closedToDate
// 		case "revtrans":
// 			if form.RevTrans {
// 				value = "1"
// 			}
// 		case "audittrail":
// 			if form.AuditTrail {
// 				value = "1"
// 			}
// 		case "extendedlog":
// 			if form.ExtendedLog {
// 				value = "1"
// 			}
// 		}
//
// 		if value != nil {
// 			if _, err := insertStmt.Exec(field, value); err != nil {
// 				return fmt.Errorf("failed to insert default %s: %w", field, err)
// 			}
// 			s.logger.Printf("Updated default %s = %v", field, value)
// 		}
// 	}
//
// 	// Remove audit trail records if requested
// 	if form.RemoveAuditTrail != "" {
// 		removeQuery := `DELETE FROM audittrail WHERE transdate < ?`
// 		result, err := tx.Exec(removeQuery, form.RemoveAuditTrail)
// 		if err != nil {
// 			return fmt.Errorf("failed to remove audit trail: %w", err)
// 		}
// 		if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
// 			s.logger.Printf("Removed %d audit trail records", rowsAffected)
// 		}
// 	}
//
// 	// Commit transaction
// 	if err := tx.Commit(); err != nil {
// 		return fmt.Errorf("failed to commit close books transaction: %w", err)
// 	}
//
// 	s.logger.Printf("Successfully closed books for period: %s", form.ClosedTo)
// 	return nil
// }
//
// // EarningsAccounts retrieves chart of accounts for earnings
// func (s *AccountingService) EarningsAccounts(config *Config, form *Form, db *sql.DB) error {
// 	// Get chart of accounts with translations
// 	// Note: This assumes a countrycode field exists in config
// 	query := `
// 		SELECT c.accno, c.description, l.description AS translation
// 		FROM chart c
// 		LEFT JOIN translation l ON (l.trans_id = c.id AND l.language_code = ?)
// 		WHERE c.charttype = 'A' AND c.category = 'Q'
// 		ORDER BY c.accno`
//
// 	rows, err := db.Query(query, config.GetCountryCode()) // Assuming this method exists
// 	if err != nil {
// 		return fmt.Errorf("earnings accounts query failed: %w", err)
// 	}
// 	defer rows.Close()
//
// 	// Initialize chart array
// 	var chart []map[string]interface{}
//
// 	// Process each account
// 	for rows.Next() {
// 		var accno, description string
// 		var translation sql.NullString
//
// 		if err := rows.Scan(&accno, &description, &translation); err != nil {
// 			return fmt.Errorf("failed to scan chart row: %w", err)
// 		}
//
// 		// Use translation if available, otherwise use original description
// 		finalDescription := description
// 		if translation.Valid {
// 			finalDescription = translation.String
// 		}
//
// 		account := map[string]interface{}{
// 			"accno":       accno,
// 			"description": finalDescription,
// 		}
// 		chart = append(chart, account)
// 	}
//
// 	// Store chart in form
// 	if form.Fields == nil {
// 		form.Fields = make(map[string]interface{})
// 	}
// 	form.Fields["chart"] = chart
//
// 	// Get method and precision defaults
// 	if err := s.getDefaults(db, form, []string{"method", "precision"}); err != nil {
// 		return fmt.Errorf("failed to get defaults: %w", err)
// 	}
//
// 	// Set default method if not specified
// 	if _, exists := form.Fields["method"]; !exists {
// 		form.Fields["method"] = "accrual"
// 	}
//
// 	return rows.Err()
// }
//
// // PostYearend posts year-end entries
// func (s *AccountingService) PostYearend(config *Config, form *Form, db *sql.DB) error {
// 	// Begin transaction
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()
//
// 	// Generate unique reference
// 	uid := fmt.Sprintf("%d%d", time.Now().Unix(), os.Getpid())
//
// 	// Get currency (assuming first 3 characters)
// 	curr, err := s.getCurrency(db, config)
// 	if err != nil {
// 		return fmt.Errorf("failed to get currency: %w", err)
// 	}
// 	if len(curr) > 3 {
// 		curr = curr[:3]
// 	}
//
// 	// Insert GL header record
// 	insertGLQuery := `
// 		INSERT INTO gl (reference, employee_id, curr)
// 		VALUES (?, (SELECT id FROM employee WHERE login = ?), ?)`
//
// 	if _, err := tx.Exec(insertGLQuery, uid, form.Login, curr); err != nil {
// 		return fmt.Errorf("failed to insert GL header: %w", err)
// 	}
//
// 	// Get the inserted GL ID
// 	var glID int
// 	if err := tx.QueryRow("SELECT id FROM gl WHERE reference = ?", uid).Scan(&glID); err != nil {
// 		return fmt.Errorf("failed to get GL ID: %w", err)
// 	}
// 	form.ID = glID
//
// 	// Update reference if not provided
// 	if form.Reference == "" {
// 		// This would typically call a function to generate next GL number
// 		form.Reference = s.updateDefaults(config, "glnumber", tx)
// 	}
//
// 	// Parse department
// 	var departmentID int
// 	if form.Department != "" {
// 		parts := strings.Split(form.Department, "--")
// 		if len(parts) > 1 {
// 			if id, err := strconv.Atoi(parts[1]); err == nil {
// 				departmentID = id
// 			}
// 		}
// 	}
//
// 	// Insert department transaction if department specified
// 	if departmentID > 0 {
// 		deptQuery := `INSERT INTO dpt_trans (trans_id, department_id) VALUES (?, ?)`
// 		if _, err := tx.Exec(deptQuery, form.ID, departmentID); err != nil {
// 			return fmt.Errorf("failed to insert department transaction: %w", err)
// 		}
// 	}
//
// 	// Update GL record with details
// 	updateGLQuery := `
// 		UPDATE gl SET
// 			reference = ?,
// 			description = ?,
// 			notes = ?,
// 			transdate = ?,
// 			department_id = ?
// 		WHERE id = ?`
//
// 	if _, err := tx.Exec(updateGLQuery, form.Reference, form.Description,
// 		form.Notes, form.TransDate, departmentID, form.ID); err != nil {
// 		return fmt.Errorf("failed to update GL record: %w", err)
// 	}
//
// 	// Insert account transactions
// 	if err := s.insertAccountTransactions(tx, form); err != nil {
// 		return fmt.Errorf("failed to insert account transactions: %w", err)
// 	}
//
// 	// Insert yearend record
// 	yearendQuery := `INSERT INTO yearend (trans_id, transdate) VALUES (?, ?)`
// 	if _, err := tx.Exec(yearendQuery, form.ID, form.TransDate); err != nil {
// 		return fmt.Errorf("failed to insert yearend record: %w", err)
// 	}
//
// 	// Add audit trail entry
// 	if err := s.addAuditTrail(tx, form); err != nil {
// 		return fmt.Errorf("failed to add audit trail: %w", err)
// 	}
//
// 	// Commit transaction
// 	return tx.Commit()
// }
//
// // CompanyDefaults retrieves company default settings
// func (s *AccountingService) CompanyDefaults(config *Config, form *Form, db *sql.DB) error {
// 	return s.getDefaults(db, form, []string{"company", "address"})
// }
//
// // RemoveLocks removes all semaphore locks
// func (s *AccountingService) RemoveLocks(config *Config, form *Form, db *sql.DB) error {
// 	query := `DELETE FROM semaphore`
// 	if _, err := db.Exec(query); err != nil {
// 		return fmt.Errorf("failed to remove locks: %w", err)
// 	}
// 	return nil
// }
//
// // Move changes the order of records in specified table
// func (s *AccountingService) Move(config *Config, form *Form, db *sql.DB) error {
// 	// Security validation - only allow specific tables and fields
// 	allowedTables := []string{"paymentmethod", "curr"}
// 	allowedFields := []string{"id", "curr"}
//
// 	if !s.contains(allowedTables, form.DB) {
// 		return fmt.Errorf("invalid table name")
// 	}
// 	if !s.contains(allowedFields, form.Fld) {
// 		return fmt.Errorf("invalid column name")
// 	}
//
// 	// Begin transaction
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()
//
// 	// Get current row number
// 	var currentRN int
// 	getRNQuery := fmt.Sprintf("SELECT rn FROM %s WHERE %s = ?", form.DB, form.Fld)
// 	if err := tx.QueryRow(getRNQuery, form.ID).Scan(&currentRN); err != nil {
// 		return fmt.Errorf("failed to get current row number: %w", err)
// 	}
//
// 	// Get maximum row number
// 	var maxRN int
// 	getMaxQuery := fmt.Sprintf("SELECT MAX(rn) FROM %s", form.DB)
// 	if err := tx.QueryRow(getMaxQuery).Scan(&maxRN); err != nil {
// 		return fmt.Errorf("failed to get max row number: %w", err)
// 	}
//
// 	var targetID string
// 	var newRN int
//
// 	switch form.Move {
// 	case "down":
// 		if currentRN < maxRN {
// 			newRN = currentRN + 1
// 			// Get ID of record to swap with
// 			getIDQuery := fmt.Sprintf("SELECT %s FROM %s WHERE rn = ?", form.Fld, form.DB)
// 			if err := tx.QueryRow(getIDQuery, newRN).Scan(&targetID); err != nil {
// 				return fmt.Errorf("failed to get target ID: %w", err)
// 			}
// 		} else {
// 			return nil // Already at bottom
// 		}
//
// 	case "up":
// 		if currentRN > 1 {
// 			newRN = currentRN - 1
// 			// Get ID of record to swap with
// 			getIDQuery := fmt.Sprintf("SELECT %s FROM %s WHERE rn = ?", form.Fld, form.DB)
// 			if err := tx.QueryRow(getIDQuery, newRN).Scan(&targetID); err != nil {
// 				return fmt.Errorf("failed to get target ID: %w", err)
// 			}
// 		} else {
// 			return nil // Already at top
// 		}
//
// 	default:
// 		return fmt.Errorf("invalid move direction: %s", form.Move)
// 	}
//
// 	// Perform the swap
// 	if targetID != "" {
// 		// Update current record's row number
// 		updateQuery1 := fmt.Sprintf("UPDATE %s SET rn = ? WHERE %s = ?", form.DB, form.Fld)
// 		if _, err := tx.Exec(updateQuery1, newRN, form.ID); err != nil {
// 			return fmt.Errorf("failed to update current record: %w", err)
// 		}
//
// 		// Update target record's row number
// 		updateQuery2 := fmt.Sprintf("UPDATE %s SET rn = ? WHERE %s = ?", form.DB, form.Fld)
// 		if _, err := tx.Exec(updateQuery2, currentRN, targetID); err != nil {
// 			return fmt.Errorf("failed to update target record: %w", err)
// 		}
// 	}
//
// 	// Commit transaction
// 	return tx.Commit()
// }
//
// // Helper methods
//
// // dateToNum converts date to numeric format (equivalent to Perl's datetonum)
// func (s *AccountingService) dateToNum(config *Config, dateStr string) string {
// 	// Remove leading/trailing whitespace
// 	dateStr = strings.TrimSpace(dateStr)
// 	// TODO: Implement date conversion logic
// 	return dateStr
// }
//
// // Config represents database configuration
// type Config struct {
// 	DSN         string `json:"dsn"`
// 	Host        string `json:"host"`
// 	Port        int    `json:"port"`
// 	Database    string `json:"database"`
// 	Username    string `json:"username"`
// 	Password    string `json:"password"`
// 	CountryCode string `json:"country_code"`
// }
//
// // Form represents form data structure
// type Form struct {
// 	ID                 int                      `json:"id"`
// 	Direction          string                   `json:"direction"`
// 	ALL                []map[string]interface{} `json:"all"`
// 	Login              string                   `json:"login"`
// 	Reference          string                   `json:"reference"`
// 	Description        string                   `json:"description"`
// 	Notes              string                   `json:"notes"`
// 	TransDate          string                   `json:"transdate"`
// 	Department         string                   `json:"department"`
// 	RowCount           int                      `json:"rowcount"`
// 	ClosedTo           string                   `json:"closedto"`
// 	DB                 string                   `json:"db"`
// 	Fld                string                   `json:"fld"`
// 	Move               string                   `json:"move"`
// 	RemoveAuditTrail   string                   `json:"removeaudittrail"`
// 	RevTrans           bool                     `json:"revtrans"`
// 	AuditTrail         bool                     `json:"audittrail"`
// 	ExtendedLog        bool                     `json:"extendedlog"`
// 	Method             string                   `json:"method"`
// 	Precision          int                      `json:"precision"`
// 	Chart              []map[string]interface{} `json:"chart"`
//
// 	// Dynamic fields for account entries and other form data
// 	Fields map[string]interface{} `json:"fields"`
// }
//
// // AccountingService handles accounting operations
// type AccountingService struct {
// 	db     *sql.DB
// 	logger *log.Logger
// }
//
// // NewAccountingService creates a new accounting service instance
// func NewAccountingService(db *sql.DB) *AccountingService {
// 	return &AccountingService{
// 		db:     db,
// 		logger: log.New(os.Stdout, "[ACCOUNTING] ", log.LstdFlags),
// 	}
// }
//
// // GetDB returns the database connection for external use
// func (s *AccountingService) GetDB() *sql.DB {
// 	return s.db
// }
//
// // ConnectDB establishes database connection based on config
// func ConnectDB(config *Config) (*sql.DB, error) {
// 	// This would typically use a specific driver like postgres, mysql, etc.
// 	// For this example, we'll use a generic approach
// 	db, err := sql.Open("postgres", config.DSN) // or config.BuildDSN()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to database: %w", err)
// 	}
//
// 	if err := db.Ping(); err != nil {
// 		return nil, fmt.Errorf("failed to ping database: %w", err)
// 	}
//
// 	return db, nil
// }
//
// // BuildDSN constructs database connection string from config
// func (c *Config) BuildDSN() string {
// 	if c.DSN != "" {
// 		return c.DSN
// 	}
// 	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
// 		c.Host, c.Port, c.Username, c.Password, c.Database)
// }
//
// // Dispatch retrieves dispatch records ordered by ID
// func (s *AccountingService) Dispatch(config *Config, form *Form) error {
// 	// Validate direction parameter
// 	direction := "ASC"
// 	if strings.ToUpper(form.Direction) == "DESC" {
// 		direction = "DESC"
// 	}
//
// 	// Sort order is handled by form.Direction (ASC/DESC)
// 	query := fmt.Sprintf(`
// 		SELECT id, description
// 		FROM dispatch
// 		ORDER BY 1 %s`, direction)
//
// 	s.logger.Printf("Executing dispatch query with direction: %s", direction)
//
// 	rows, err := s.db.Query(query)
// 	if err != nil {
// 		s.logger.Printf("Dispatch query failed: %v", err)
// 		return fmt.Errorf("dispatch query failed: %w", err)
// 	}
// 	defer rows.Close()
//
// 	// Clear existing results
// 	form.ALL = make([]map[string]interface{}, 0)
//
// 	// Fetch all rows and store in form.ALL
// 	for rows.Next() {
// 		var id int
// 		var description string
//
// 		if err := rows.Scan(&id, &description); err != nil {
// 			s.logger.Printf("Failed to scan dispatch row: %v", err)
// 			return fmt.Errorf("failed to scan dispatch row: %w", err)
// 		}
//
// 		record := map[string]interface{}{
// 			"id":          id,
// 			"description": description,
// 		}
// 		form.ALL = append(form.ALL, record)
// 	}
//
// 	s.logger.Printf("Retrieved %d dispatch records", len(form.ALL))
// 	return rows.Err()
// }
//
// // ClosedTo retrieves system defaults for closed period settings
// func (s *AccountingService) ClosedTo(config *Config, form *Form) error {
// 	// Get default values for closedto, revtrans, audittrail, extendedlog
// 	defaults := []string{"closedto", "revtrans", "audittrail", "extendedlog"}
//
// 	query := `
// 		SELECT fldname, fldvalue
// 		FROM defaults
// 		WHERE fldname IN (?, ?, ?, ?)`
//
// 	s.logger.Printf("Retrieving closed period defaults")
//
// 	rows, err := s.db.Query(query, defaults[0], defaults[1], defaults[2], defaults[3])
// 	if err != nil {
// 		s.logger.Printf("ClosedTo query failed: %v", err)
// 		return fmt.Errorf("closedto query failed: %w", err)
// 	}
// 	defer rows.Close()
//
// 	// Initialize fields map if needed
// 	if form.Fields == nil {
// 		form.Fields = make(map[string]interface{})
// 	}
//
// 	// Store retrieved defaults in form
// 	for rows.Next() {
// 		var fieldName, fieldValue string
// 		if err := rows.Scan(&fieldName, &fieldValue); err != nil {
// 			s.logger.Printf("Failed to scan defaults: %v", err)
// 			return fmt.Errorf("failed to scan defaults: %w", err)
// 		}
// 		form.Fields[fieldName] = fieldValue
//
// 		// Also set specific form fields for convenience
// 		switch fieldName {
// 		case "closedto":
// 			form.ClosedTo = fieldValue
// 		case "revtrans":
// 			form.RevTrans = fieldValue == "1" || strings.ToLower(fieldValue) == "true"
// 		case "audittrail":
// 			form.AuditTrail = fieldValue == "1" || strings.ToLower(fieldValue) == "true"
// 		case "extendedlog":
// 			form.ExtendedLog = fieldValue == "1" || strings.ToLower(fieldValue) == "true"
// 		}
// 	}
//
// 	s.logger.Printf("Retrieved defaults for: %v", defaults)
// 	return rows.Err()
// }
//
// // CloseBooks closes accounting books for a period
// func (s *AccountingService) CloseBooks(config *Config, form *Form) error {
// 	s.logger.Printf("Starting close books process for period: %s", form.ClosedTo)
//
// 	// Begin transaction
// 	tx, err := s.db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer func() {
// 		if err != nil {
// 			s.logger.Printf("Rolling back close books transaction: %v", err)
// 			tx.Rollback()
// 		}
// 	}()
//
// 	// Prepare delete statement for existing defaults
// 	deleteQuery := `DELETE FROM defaults WHERE fldname = ?`
// 	deleteStmt, err := tx.Prepare(deleteQuery)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare delete statement: %w", err)
// 	}
// 	defer deleteStmt.Close()
//
// 	// Prepare insert statement for new defaults
// 	insertQuery := `INSERT INTO defaults (fldname, fldvalue) VALUES (?, ?)`
// 	insertStmt, err := tx.Prepare(insertQuery)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare insert statement: %w", err)
// 	}
// 	defer insertStmt.Close()
//
// 	// Convert and validate closedto date using datetonum equivalent
// 	closedToDate := s.dateToNum(config, form.ClosedTo)
// 	if err := s.validateDate(closedToDate); err != nil {
// 		return fmt.Errorf("invalid closedto date: %w", err)
// 	}
//
// 	// Process each default field
// 	fields := []string{"revtrans", "closedto", "audittrail", "extendedlog"}
// 	for _, field := range fields {
// 		// Delete existing entry
// 		if _, err := deleteStmt.Exec(field); err != nil {
// 			return fmt.Errorf("failed to delete default %s: %w", field, err)
// 		}
//
// 		// Insert new value if it exists in form
// 		var value interface{}
// 		switch field {
// 		case "closedto":
// 			value = closedToDate
// 		case "revtrans":
// 			if form.RevTrans {
// 				value = "1"
// 			}
// 		case "audittrail":
// 			if form.AuditTrail {
// 				value = "1"
// 			}
// 		case "extendedlog":
// 			if form.ExtendedLog {
// 				value = "1"
// 			}
// 		}
//
// 		if value != nil {
// 			if _, err := insertStmt.Exec(field, value); err != nil {
// 				return fmt.Errorf("failed to insert default %s: %w", field, err)
// 			}
// 			s.logger.Printf("Updated default %s = %v", field, value)
// 		}
// 	}
//
// 	// Remove audit trail records if requested
// 	if form.RemoveAuditTrail != "" {
// 		removeQuery := `DELETE FROM audittrail WHERE transdate < ?`
// 		result, err := tx.Exec(removeQuery, form.RemoveAuditTrail)
// 		if err != nil {
// 			return fmt.Errorf("failed to remove audit trail: %w", err)
// 		}
// 		if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
// 			s.logger.Printf("Removed %d audit trail records", rowsAffected)
// 		}
// 	}
//
// 	// Commit transaction
// 	if err := tx.Commit(); err != nil {
// 		return fmt.Errorf("failed to commit close books transaction: %w", err)
// 	}
//
// 	s.logger.Printf("Successfully closed books for period: %s", form.ClosedTo)
// 	return nil
// }
//
// // EarningsAccounts retrieves chart of accounts for earnings
// func (s *AccountingService) EarningsAccounts(config *Config, form *Form, db *sql.DB) error {
// 	// Get chart of accounts with translations
// 	// Note: This assumes a countrycode field exists in config
// 	query := `
// 		SELECT c.accno, c.description, l.description AS translation
// 		FROM chart c
// 		LEFT JOIN translation l ON (l.trans_id = c.id AND l.language_code = ?)
// 		WHERE c.charttype = 'A' AND c.category = 'Q'
// 		ORDER BY c.accno`
//
// 	rows, err := db.Query(query, config.GetCountryCode()) // Assuming this method exists
// 	if err != nil {
// 		return fmt.Errorf("earnings accounts query failed: %w", err)
// 	}
// 	defer rows.Close()
//
// 	// Initialize chart array
// 	var chart []map[string]interface{}
//
// 	// Process each account
// 	for rows.Next() {
// 		var accno, description string
// 		var translation sql.NullString
//
// 		if err := rows.Scan(&accno, &description, &translation); err != nil {
// 			return fmt.Errorf("failed to scan chart row: %w", err)
// 		}
//
// 		// Use translation if available, otherwise use original description
// 		finalDescription := description
// 		if translation.Valid {
// 			finalDescription = translation.String
// 		}
//
// 		account := map[string]interface{}{
// 			"accno":       accno,
// 			"description": finalDescription,
// 		}
// 		chart = append(chart, account)
// 	}
//
// 	// Store chart in form
// 	if form.Fields == nil {
// 		form.Fields = make(map[string]interface{})
// 	}
// 	form.Fields["chart"] = chart
//
// 	// Get method and precision defaults
// 	if err := s.getDefaults(db, form, []string{"method", "precision"}); err != nil {
// 		return fmt.Errorf("failed to get defaults: %w", err)
// 	}
//
// 	// Set default method if not specified
// 	if _, exists := form.Fields["method"]; !exists {
// 		form.Fields["method"] = "accrual"
// 	}
//
// 	return rows.Err()
// }
//
// // PostYearend posts year-end entries
// func (s *AccountingService) PostYearend(config *Config, form *Form, db *sql.DB) error {
// 	// Begin transaction
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()
//
// 	// Generate unique reference
// 	uid := fmt.Sprintf("%d%d", time.Now().Unix(), os.Getpid())
//
// 	// Get currency (assuming first 3 characters)
// 	curr, err := s.getCurrency(db, config)
// 	if err != nil {
// 		return fmt.Errorf("failed to get currency: %w", err)
// 	}
// 	if len(curr) > 3 {
// 		curr = curr[:3]
// 	}
//
// 	// Insert GL header record
// 	insertGLQuery := `
// 		INSERT INTO gl (reference, employee_id, curr)
// 		VALUES (?, (SELECT id FROM employee WHERE login = ?), ?)`
//
// 	if _, err := tx.Exec(insertGLQuery, uid, form.Login, curr); err != nil {
// 		return fmt.Errorf("failed to insert GL header: %w", err)
// 	}
//
// 	// Get the inserted GL ID
// 	var glID int
// 	if err := tx.QueryRow("SELECT id FROM gl WHERE reference = ?", uid).Scan(&glID); err != nil {
// 		return fmt.Errorf("failed to get GL ID: %w", err)
// 	}
// 	form.ID = glID
//
// 	// Update reference if not provided
// 	if form.Reference == "" {
// 		// This would typically call a function to generate next GL number
// 		form.Reference = s.updateDefaults(config, "glnumber", tx)
// 	}
//
// 	// Parse department
// 	var departmentID int
// 	if form.Department != "" {
// 		parts := strings.Split(form.Department, "--")
// 		if len(parts) > 1 {
// 			if id, err := strconv.Atoi(parts[1]); err == nil {
// 				departmentID = id
// 			}
// 		}
// 	}
//
// 	// Insert department transaction if department specified
// 	if departmentID > 0 {
// 		deptQuery := `INSERT INTO dpt_trans (trans_id, department_id) VALUES (?, ?)`
// 		if _, err := tx.Exec(deptQuery, form.ID, departmentID); err != nil {
// 			return fmt.Errorf("failed to insert department transaction: %w", err)
// 		}
// 	}
//
// 	// Update GL record with details
// 	updateGLQuery := `
// 		UPDATE gl SET
// 			reference = ?,
// 			description = ?,
// 			notes = ?,
// 			transdate = ?,
// 			department_id = ?
// 		WHERE id = ?`
//
// 	if _, err := tx.Exec(updateGLQuery, form.Reference, form.Description,
// 		form.Notes, form.TransDate, departmentID, form.ID); err != nil {
// 		return fmt.Errorf("failed to update GL record: %w", err)
// 	}
//
// 	// Insert account transactions
// 	if err := s.insertAccountTransactions(tx, form); err != nil {
// 		return fmt.Errorf("failed to insert account transactions: %w", err)
// 	}
//
// 	// Insert yearend record
// 	yearendQuery := `INSERT INTO yearend (trans_id, transdate) VALUES (?, ?)`
// 	if _, err := tx.Exec(yearendQuery, form.ID, form.TransDate); err != nil {
// 		return fmt.Errorf("failed to insert yearend record: %w", err)
// 	}
//
// 	// Add audit trail entry
// 	if err := s.addAuditTrail(tx, form); err != nil {
// 		return fmt.Errorf("failed to add audit trail: %w", err)
// 	}
//
// 	// Commit transaction
// 	return tx.Commit()
// }
//
// // CompanyDefaults retrieves company default settings
// func (s *AccountingService) CompanyDefaults(config *Config, form *Form, db *sql.DB) error {
// 	return s.getDefaults(db, form, []string{"company", "address"})
// }
//
// // RemoveLocks removes all semaphore locks
// func (s *AccountingService) RemoveLocks(config *Config, form *Form, db *sql.DB) error {
// 	query := `DELETE FROM semaphore`
// 	if _, err := db.Exec(query); err != nil {
// 		return fmt.Errorf("failed to remove locks: %w", err)
// 	}
// 	return nil
// }
//
// // Move changes the order of records in specified table
// func (s *AccountingService) Move(config *Config, form *Form, db *sql.DB) error {
// 	// Security validation - only allow specific tables and fields
// 	allowedTables := []string{"paymentmethod", "curr"}
// 	allowedFields := []string{"id", "curr"}
//
// 	if !s.contains(allowedTables, form.DB) {
// 		return fmt.Errorf("invalid table name")
// 	}
// 	if !s.contains(allowedFields, form.Fld) {
// 		return fmt.Errorf("invalid column name")
// 	}
//
// 	// Begin transaction
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()
//
// 	// Get current row number
// 	var currentRN int
// 	getRNQuery := fmt.Sprintf("SELECT rn FROM %s WHERE %s = ?", form.DB, form.Fld)
// 	if err := tx.QueryRow(getRNQuery, form.ID).Scan(&currentRN); err != nil {
// 		return fmt.Errorf("failed to get current row number: %w", err)
// 	}
//
// 	// Get maximum row number
// 	var maxRN int
// 	getMaxQuery := fmt.Sprintf("SELECT MAX(rn) FROM %s", form.DB)
// 	if err := tx.QueryRow(getMaxQuery).Scan(&maxRN); err != nil {
// 		return fmt.Errorf("failed to get max row number: %w", err)
// 	}
//
// 	var targetID string
// 	var newRN int
//
// 	switch form.Move {
// 	case "down":
// 		if currentRN < maxRN {
// 			newRN = currentRN + 1
// 			// Get ID of record to swap with
// 			getIDQuery := fmt.Sprintf("SELECT %s FROM %s WHERE rn = ?", form.Fld, form.DB)
// 			if err := tx.QueryRow(getIDQuery, newRN).Scan(&targetID); err != nil {
// 				return fmt.Errorf("failed to get target ID: %w", err)
// 			}
// 		} else {
// 			return nil // Already at bottom
// 		}
//
// 	case "up":
// 		if currentRN > 1 {
// 			newRN = currentRN - 1
// 			// Get ID of record to swap with
// 			getIDQuery := fmt.Sprintf("SELECT %s FROM %s WHERE rn = ?", form.Fld, form.DB)
// 			if err := tx.QueryRow(getIDQuery, newRN).Scan(&targetID); err != nil {
// 				return fmt.Errorf("failed to get target ID: %w", err)
// 			}
// 		} else {
// 			return nil // Already at top
// 		}
//
// 	default:
// 		return fmt.Errorf("invalid move direction: %s", form.Move)
// 	}
//
// 	// Perform the swap
// 	if targetID != "" {
// 		// Update current record's row number
// 		updateQuery1 := fmt.Sprintf("UPDATE %s SET rn = ? WHERE %s = ?", form.DB, form.Fld)
// 		if _, err := tx.Exec(updateQuery1, newRN, form.ID); err != nil {
// 			return fmt.Errorf("failed to update current record: %w", err)
// 		}
//
// 		// Update target record's row number
// 		updateQuery2 := fmt.Sprintf("UPDATE %s SET rn = ? WHERE %s = ?", form.DB, form.Fld)
// 		if _, err := tx.Exec(updateQuery2, currentRN, targetID); err != nil {
// 			return fmt.Errorf("failed to update target record: %w", err)
// 		}
// 	}
//
// 	// Commit transaction
// 	return tx.Commit()
// }
//
// // validateDate validates the date format (YYYYMMDD)
// func (s *AccountingService) validateDate(dateStr string) error {
// 	dateStr = strings.TrimSpace(dateStr)
//
// 	if len(dateStr) != 8 {
// 		return fmt.Errorf("date must be 8 characters (YYYYMMDD), got: %s", dateStr)
// 	}
//
// 	year, err := strconv.Atoi(dateStr[:4])
// 	if err != nil {
// 		return fmt.Errorf("invalid year in date %s: %w", dateStr, err)
// 	}
//
// 	month, err := strconv.Atoi(dateStr[4:6])
// 	if err != nil {
// 		return fmt.Errorf("invalid month in date %s: %w", dateStr, err)
// 	}
//
// 	day, err := strconv.Atoi(dateStr[6:8])
// 	if err != nil {
// 		return fmt.Errorf("invalid day in date %s: %w", dateStr, err)
// 	}
//
// 	// Validate the date is actually valid
// 	if month < 1 || month > 12 {
// 		return fmt.Errorf("invalid month %d in date %s", month, dateStr)
// 	}
//
// 	if day < 1 || day > 31 {
// 		return fmt.Errorf("invalid day %d in date %s", day, dateStr)
// 	}
//
// 	// Use Go's time package to validate the actual date
// 	_, err = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).MarshalText()
// 	if err != nil {
// 		return fmt.Errorf("invalid date %s: %w", dateStr, err)
// 	}
//
// 	return nil
// }
//
// // getDefaults retrieves default values from database
// func (s *AccountingService) getDefaults(form *Form, fields []string) error {
// 	if len(fields) == 0 {
// 		return nil
// 	}
//
// 	// Build query with placeholders
// 	placeholders := strings.Repeat("?,", len(fields))
// 	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
//
// 	query := fmt.Sprintf("SELECT fldname, fldvalue FROM defaults WHERE fldname IN (%s)", placeholders)
//
// 	// Convert fields to interface{} slice for variadic function
// 	args := make([]interface{}, len(fields))
// 	for i, field := range fields {
// 		args[i] = field
// 	}
//
// 	s.logger.Printf("Getting defaults for fields: %v", fields)
//
// 	rows, err := s.db.Query(query, args...)
// 	if err != nil {
// 		s.logger.Printf("Defaults query failed: %v", err)
// 		return fmt.Errorf("defaults query failed: %w", err)
// 	}
// 	defer rows.Close()
//
// 	// Initialize fields map if needed
// 	if form.Fields == nil {
// 		form.Fields = make(map[string]interface{})
// 	}
//
// 	// Store retrieved defaults
// 	for rows.Next() {
// 		var fieldName, fieldValue string
// 		if err := rows.Scan(&fieldName, &fieldValue); err != nil {
// 			s.logger.Printf("Failed to scan defaults: %v", err)
// 			return fmt.Errorf("failed to scan defaults: %w", err)
// 		}
// 		form.Fields[fieldName] = fieldValue
//
// 		// Set specific form fields
// 		switch fieldName {
// 		case "precision":
// 			if precision, err := strconv.Atoi(fieldValue); err == nil {
// 				form.Precision = precision
// 			}
// 		case "method":
// 			form.Method = fieldValue
// 		}
// 	}
//
// 	return rows.Err()
// }
//
// // contains checks if a slice contains a string
// func (s *AccountingService) contains(slice []string, item string) bool {
// 	for _, s := range slice {
// 		if s == item {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// // insertAccountTransactions inserts account transaction entries
// func (s *AccountingService) insertAccountTransactions(tx *sql.Tx, form *Form) error {
// 	insertQuery := `
// 		INSERT INTO acc_trans (trans_id, chart_id, amount, transdate, source)
// 		VALUES (?, (SELECT id FROM chart WHERE accno = ?), ?, ?, ?)`
//
// 	stmt, err := tx.Prepare(insertQuery)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare account transaction statement: %w", err)
// 	}
// 	defer stmt.Close()
//
// 	s.logger.Printf("Processing %d account transaction rows", form.RowCount)
//
// 	// Process each row
// 	for i := 1; i <= form.RowCount; i++ {
// 		accnoKey := fmt.Sprintf("accno_%d", i)
// 		creditKey := fmt.Sprintf("credit_%d", i)
// 		debitKey := fmt.Sprintf("debit_%d", i)
//
// 		// Extract account number
// 		var accno string
// 		if accnoField, exists := form.Fields[accnoKey]; exists {
// 			if accnoStr, ok := accnoField.(string); ok {
// 				parts := strings.Split(accnoStr, "--")
// 				if len(parts) > 0 {
// 					accno = parts[0]
// 				}
// 			}
// 		}
//
// 		// Calculate amount
// 		var amount float64
// 		if creditField, exists := form.Fields[creditKey]; exists {
// 			switch v := creditField.(type) {
// 			case float64:
// 				amount = v
// 			case string:
// 				if f, err := strconv.ParseFloat(v, 64); err == nil {
// 					amount = f
// 				}
// 			}
// 		}
// 		if debitField, exists := form.Fields[debitKey]; exists {
// 			switch v := debitField.(type) {
// 			case float64:
// 				amount = -v // Debit is negative
// 			case string:
// 				if f, err := strconv.ParseFloat(v, 64); err == nil {
// 					amount = -f // Debit is negative
// 				}
// 			}
// 		}
//
// 		// Insert transaction if amount is non-zero
// 		if amount != 0 && accno != "" {
// 			if _, err := stmt.Exec(form.ID, accno, amount, form.TransDate, form.Reference); err != nil {
// 				return fmt.Errorf("failed to insert account transaction for row %d: %w", i, err)
// 			}
// 			s.logger.Printf("Inserted transaction: account=%s, amount=%.2f", accno, amount)
// 		}
// 	}
//
// 	return nil
// }
//
// // addAuditTrail adds an audit trail entry
// func (s *AccountingService) addAuditTrail(tx *sql.Tx, form *Form) error {
// 	auditQuery := `
// 		INSERT INTO audittrail (tablename, reference, formname, action, trans_id, transdate)
// 		VALUES (?, ?, ?, ?, ?, ?)`
//
// 	_, err := tx.Exec(auditQuery, "gl", form.Reference, "yearend", "posted", form.ID, form.TransDate)
// 	if err != nil {
// 		return fmt.Errorf("failed to insert audit trail: %w", err)
// 	}
//
// 	s.logger.Printf("Added audit trail for GL transaction %d", form.ID)
// 	return nil
// }
//
// // getCurrency gets the default currency from system
// func (s *AccountingService) getCurrency(config *Config) (string, error) {
// 	query := `SELECT fldvalue FROM defaults WHERE fldname = 'curr' LIMIT 1`
//
// 	var currency string
// 	err := s.db.QueryRow(query).Scan(&currency)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			// Return default currency if none found
// 			return "USD", nil
// 		}
// 		return "", fmt.Errorf("failed to get currency: %w", err)
// 	}
//
// 	return currency, nil
// }
//
// // updateDefaults generates next sequence number for specified counter
// func (s *AccountingService) updateDefaults(config *Config, counter string, tx *sql.Tx) string {
// 	// Get current value
// 	var currentVal int
// 	query := `SELECT COALESCE(fldvalue::integer, 0) FROM defaults WHERE fldname = ?`
// 	err := tx.QueryRow(query, counter).Scan(&currentVal)
//
// 	// Increment the value
// 	nextVal := currentVal + 1
//
// 	// Update or insert the new value
// 	updateQuery := `
// 		INSERT INTO defaults (fldname, fldvalue)
// 		VALUES (?, ?)
// 		ON CONFLICT (fldname)
// 		DO UPDATE SET fldvalue = EXCLUDED.fldvalue`
//
// 	_, err = tx.Exec(updateQuery, counter, strconv.Itoa(nextVal))
// 	if err != nil {
// 		s.logger.Printf("Failed to update defaults for %s: %v", counter, err)
// 		// Fallback to timestamp-based reference
// 		return fmt.Sprintf("GL%d", time.Now().Unix())
// 	}
//
// 	return fmt.Sprintf("GL%06d", nextVal)
// }
//
// // GetCountryCode returns the country code from config
// func (c *Config) GetCountryCode() string {
// 	if c.CountryCode != "" {
// 		return c.CountryCode
// 	}
// 	return "US" // Default fallback
// }
//
// // contains checks if a slice contains a string
// func (s *AccountingService) contains(slice []string, item string) bool {
// 	for _, s := range slice {
// 		if s == item {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// // insertAccountTransactions inserts account transaction entries
// func (s *AccountingService) insertAccountTransactions(tx *sql.Tx, form *Form) error {
// 	insertQuery := `
// 		INSERT INTO acc_trans (trans_id, chart_id, amount, transdate, source)
// 		VALUES (?, (SELECT id FROM chart WHERE accno = ?), ?, ?, ?)`
//
// 	stmt, err := tx.Prepare(insertQuery)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare account transaction statement: %w", err)
// 	}
// 	defer stmt.Close()
//
// 	// Process each row
// 	for i := 1; i <= form.RowCount; i++ {
// 		accnoKey := fmt.Sprintf("accno_%d", i)
// 		creditKey := fmt.Sprintf("credit_%d", i)
// 		debitKey := fmt.Sprintf("debit_%d", i)
//
// 		// Extract account number
// 		var accno string
// 		if accnoField, exists := form.Fields[accnoKey]; exists {
// 			if accnoStr, ok := accnoField.(string); ok {
// 				parts := strings.Split(accnoStr, "--")
// 				if len(parts) > 0 {
// 					accno = parts[0]
// 				}
// 			}
// 		}
//
// 		// Calculate amount
// 		var amount float64
// 		if creditField, exists := form.Fields[creditKey]; exists {
// 			if credit, ok := creditField.(float64); ok {
// 				amount = credit
// 			}
// 		}
// 		if debitField, exists := form.Fields[debitKey]; exists {
// 			if debit, ok := debitField.(float64); ok {
// 				amount = -debit // Debit is negative
// 			}
// 		}
//
// 		// Insert transaction if amount is non-zero
// 		if amount != 0 && accno != "" {
// 			if _, err := stmt.Exec(form.ID, accno, amount, form.TransDate, form.Reference); err != nil {
// 				return fmt.Errorf("failed to insert account transaction: %w", err)
// 			}
// 		}
// 	}
//
// 	return nil
// }
//
// // addAuditTrail adds an audit trail entry
// func (s *AccountingService) addAuditTrail(tx *sql.Tx, form *Form) error {
// 	auditQuery := `
// 		INSERT INTO audittrail (tablename, reference, formname, action, trans_id, transdate)
// 		VALUES (?, ?, ?, ?, ?, ?)`
//
// 	_, err := tx.Exec(auditQuery, "gl", form.Reference, "yearend", "posted", form.ID, form.TransDate)
// 	if err != nil {
// 		return fmt.Errorf("failed to insert audit trail: %w", err)
// 	}
//
// 	return nil
// }
//
// // getCurrency gets the default currency
// func (s *AccountingService) getCurrency(db *sql.DB, config *Config) (string, error) {
// 	// This is a placeholder - implementation would depend on your system
// 	// For now, return a default currency
// 	return "USD", nil
// }
//
// // updateDefaults generates next sequence number for specified counter
// func (s *AccountingService) updateDefaults(config *Config, counter string, tx *sql.Tx) string {
// 	// This is a placeholder - implementation would depend on your system
// 	// Typically this would increment a counter in the defaults table
// 	return fmt.Sprintf("GL%d", time.Now().Unix())
// }
//
// // GetCountryCode returns the country code from config
// func (c *Config) GetCountryCode() string {
// 	// Placeholder - this would return the actual country code from config
// 	return "US"
// }
