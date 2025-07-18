package common

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"strings"
// 	"time"
// )
//
// // RecurringTransaction represents a recurring transaction record
// type RecurringTransaction struct {
// 	Module           string  `json:"module"`
// 	Transaction      string  `json:"transaction"`
// 	Invoice          bool    `json:"invoice"`
// 	Description      string  `json:"description"`
// 	Name             string  `json:"name"`
// 	VCNumber         string  `json:"vcnumber"`
// 	NameID           int     `json:"name_id"`
// 	Amount           float64 `json:"amount"`
// 	ID               int     `json:"id"`
// 	NextDate         *time.Time `json:"nextdate"`
// 	EndDate          *time.Time `json:"enddate"`
// 	Repeat           int     `json:"repeat"`
// 	Unit             string  `json:"unit"`
// 	RecurringEmail   string  `json:"recurringemail"`
// 	RecurringPrint   string  `json:"recurringprint"`
// 	Overdue          int     `json:"overdue"`
// 	VC               string  `json:"vc"`
// 	ExchangeRate     float64 `json:"exchangerate"`
// 	Curr             string  `json:"curr"`
// 	Expired          bool    `json:"expired"`
// 	Department       string  `json:"department"`
// }
//
// // RecurringDetails represents detailed recurring transaction information
// type RecurringDetails struct {
// 	ID               int        `json:"id"`
// 	NextDate         *time.Time `json:"nextdate"`
// 	EndDate          *time.Time `json:"enddate"`
// 	Repeat           int        `json:"repeat"`
// 	Unit             string     `json:"unit"`
// 	ARID             *int       `json:"arid"`
// 	ARInvoice        *bool      `json:"arinvoice"`
// 	APID             *int       `json:"apid"`
// 	APInvoice        *bool      `json:"apinvoice"`
// 	Overdue          *int       `json:"overdue"`
// 	Paid             *int       `json:"paid"`
// 	Req              *int       `json:"req"`
// 	OEID             *int       `json:"oeid"`
// 	CustomerID       *int       `json:"customer_id"`
// 	VendorID         *int       `json:"vendor_id"`
// 	VC               string     `json:"vc"`
// 	Invoice          bool       `json:"invoice"`
// 	RecurringEmail   string     `json:"recurringemail"`
// 	RecurringPrint   string     `json:"recurringprint"`
// 	Message          string     `json:"message"`
// }
//
// // Config represents database and application configuration
// type Config struct {
// 	DBDriver   string
// 	DateFormat string
// 	Precision  int
// 	Company    string
// }
//
// // Form represents the form/request data structure
// type Form struct {
// 	Sort         string
// 	Transactions map[string][]RecurringTransaction
// 	Defaults     map[string]any
// }
//
// // RecurringService handles recurring transaction operations
// type RecurringService struct {
// 	db *sql.DB
// }
//
// // NewRecurringService creates a new RecurringService
// func NewRecurringService(db *sql.DB) *RecurringService {
// 	return &RecurringService{db: db}
// }
//
// // GetRecurringTransactions retrieves all recurring transactions
// func (rs *RecurringService) GetRecurringTransactions(config *Config, form *Form) error {
// 	// Get defaults (precision, company)
// 	defaults, err := rs.getDefaults()
// 	if err != nil {
// 		return fmt.Errorf("failed to get defaults: %w", err)
// 	}
// 	form.Defaults = defaults
//
// 	// Define sort order mapping
// 	ordinal := map[string]int{
// 		"reference":   10,
// 		"department":  26,
// 		"description": 4,
// 		"name":        5,
// 		"vcnumber":    6,
// 		"nextdate":    12,
// 		"enddate":     13,
// 	}
//
// 	// Set default sort
// 	if form.Sort == "" {
// 		form.Sort = "nextdate"
// 	}
//
// 	sortOrder := rs.buildSortOrder(form.Sort, ordinal)
//
// 	// Get default currency
// 	defaultCurrency, err := rs.getDefaultCurrency()
// 	if err != nil {
// 		return fmt.Errorf("failed to get default currency: %w", err)
// 	}
//
// 	// Build and execute the main query
// 	query := rs.buildRecurringQuery(defaultCurrency, sortOrder)
//
// 	rows, err := rs.db.Query(query)
// 	if err != nil {
// 		return fmt.Errorf("failed to execute query: %w", err)
// 	}
// 	defer rows.Close()
//
// 	// Process results
// 	form.Transactions = make(map[string][]RecurringTransaction)
// 	err = rs.processRecurringRows(rows, form)
// 	if err != nil {
// 		return fmt.Errorf("failed to process rows: %w", err)
// 	}
//
// 	return nil
// }
//
// // GetRecurringDetails retrieves details for a specific recurring transaction
// func (rs *RecurringService) GetRecurringDetails(config *Config, id int) (*RecurringDetails, error) {
// 	details := &RecurringDetails{}
//
// 	// Main details query
// 	query := `SELECT s.*, ar.id AS arid, ar.invoice AS arinvoice,
//                      ap.id AS apid, ap.invoice AS apinvoice,
//                      EXTRACT(DAY FROM (ar.duedate - ar.transdate)) AS overdue,
//                      EXTRACT(DAY FROM (ar.datepaid - ar.transdate)) AS paid,
//                      EXTRACT(DAY FROM (oe.reqdate - oe.transdate)) AS req,
//                      oe.id AS oeid, oe.customer_id, oe.vendor_id
//               FROM recurring s
//               LEFT JOIN ar ON (ar.id = s.id)
//               LEFT JOIN ap ON (ap.id = s.id)
//               LEFT JOIN oe ON (oe.id = s.id)
//               WHERE s.id = $1`
//
// 	row := rs.db.QueryRow(query, id)
// 	err := rs.scanRecurringDetails(row, details)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to scan recurring details: %w", err)
// 	}
//
// 	// Set VC based on customer or vendor
// 	if details.CustomerID != nil {
// 		details.VC = "customer"
// 	} else if details.VendorID != nil {
// 		details.VC = "vendor"
// 	}
//
// 	// Set invoice flag
// 	details.Invoice = (details.ARID != nil && details.ARInvoice != nil && *details.ARInvoice) ||
// 		(details.APID != nil && details.APInvoice != nil && *details.APInvoice)
//
// 	// Get recurring email details
// 	emailDetails, err := rs.getRecurringEmailDetails(id)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get email details: %w", err)
// 	}
// 	details.RecurringEmail = emailDetails.Email
// 	details.Message = emailDetails.Message
//
// 	// Get recurring print details
// 	details.RecurringPrint, err = rs.getRecurringPrintDetails(id)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get print details: %w", err)
// 	}
//
// 	return details, nil
// }
//
// // UpdateRecurring updates the next date for a recurring transaction
// func (rs *RecurringService) UpdateRecurring(config *Config, id int, nextDate time.Time) error {
// 	// Get current repeat and unit
// 	var repeat int
// 	var unit string
// 	query := `SELECT repeat, unit FROM recurring WHERE id = $1`
// 	err := rs.db.QueryRow(query, id).Scan(&repeat, &unit)
// 	if err != nil {
// 		return fmt.Errorf("failed to get repeat info: %w", err)
// 	}
//
// 	// Build interval calculation based on database driver
// 	var intervalExpr string
// 	switch config.DBDriver {
// 	case "postgres", "Pg", "PgPP":
// 		intervalExpr = fmt.Sprintf("(date '%s' + interval '%d %s')",
// 			nextDate.Format("2006-01-02"), repeat, unit)
// 	case "mysql":
// 		intervalExpr = fmt.Sprintf("DATE_ADD('%s', INTERVAL %d %s)",
// 			nextDate.Format("2006-01-02"), repeat, strings.ToUpper(unit))
// 	default:
// 		intervalExpr = fmt.Sprintf("(date '%s' + interval '%d %s')",
// 			nextDate.Format("2006-01-02"), repeat, unit)
// 	}
//
// 	// Check if this would be the last occurrence
// 	checkQuery := fmt.Sprintf(`SELECT %s > enddate FROM recurring WHERE id = $1`, intervalExpr)
// 	var isLast bool
// 	err = rs.db.QueryRow(checkQuery, id).Scan(&isLast)
// 	if err != nil {
// 		return fmt.Errorf("failed to check if last repeat: %w", err)
// 	}
//
// 	// Update next date
// 	var updateQuery string
// 	if isLast {
// 		updateQuery = `UPDATE recurring SET nextdate = NULL WHERE id = $1`
// 		_, err = rs.db.Exec(updateQuery, id)
// 	} else {
// 		updateQuery = fmt.Sprintf(`UPDATE recurring SET nextdate = %s WHERE id = $1`, intervalExpr)
// 		_, err = rs.db.Exec(updateQuery, id)
// 	}
//
// 	if err != nil {
// 		return fmt.Errorf("failed to update recurring: %w", err)
// 	}
//
// 	return nil
// }
//
// // Helper methods
//
// func (rs *RecurringService) getDefaults() (map[string]any, error) {
// 	defaults := make(map[string]any)
//
// 	// This would typically query a defaults/config table
// 	// For now, returning empty map as the original Perl code structure isn't clear
//
// 	return defaults, nil
// }
//
// func (rs *RecurringService) getDefaultCurrency() (string, error) {
// 	var currency string
// 	query := `SELECT curr FROM curr ORDER BY rn LIMIT 1`
// 	err := rs.db.QueryRow(query).Scan(&currency)
// 	if err != nil {
// 		return "", err
// 	}
// 	return currency, nil
// }
//
// func (rs *RecurringService) buildSortOrder(sort string, ordinal map[string]int) string {
// 	// In a real implementation, you'd build proper ORDER BY clause
// 	// This is simplified for the conversion
// 	return sort
// }
//
// func (rs *RecurringService) buildRecurringQuery(defaultCurrency, sortOrder string) string {
// 	return fmt.Sprintf(`
// 		SELECT 'ar' AS module, 'ar' AS transaction, a.invoice,
// 		       a.description, n.name, n.customernumber AS vcnumber,
// 		       n.id AS name_id, a.amount, s.*, se.formname AS recurringemail,
// 		       sp.formname AS recurringprint,
// 		       EXTRACT(DAY FROM (s.nextdate - CURRENT_DATE)) AS overdue, 'customer' AS vc,
// 		       COALESCE(ex.buy, 1) AS exchangerate, a.curr,
// 		       (s.nextdate IS NULL OR s.nextdate > s.enddate) AS expired,
// 		       d.description AS department
// 		FROM recurring s
// 		JOIN ar a ON (a.id = s.id)
// 		LEFT JOIN department d ON (d.id = a.department_id)
// 		JOIN customer n ON (n.id = a.customer_id)
// 		LEFT JOIN recurringemail se ON (se.id = s.id)
// 		LEFT JOIN recurringprint sp ON (sp.id = s.id)
// 		LEFT JOIN exchangerate ex ON (ex.curr = a.curr AND a.transdate = ex.transdate)
//
// 		UNION
//
// 		SELECT 'ap' AS module, 'ap' AS transaction, a.invoice,
// 		       a.description, n.name, n.vendornumber AS vcnumber,
// 		       n.id AS name_id, a.amount, s.*, se.formname AS recurringemail,
// 		       sp.formname AS recurringprint,
// 		       EXTRACT(DAY FROM (s.nextdate - CURRENT_DATE)) AS overdue, 'vendor' AS vc,
// 		       COALESCE(ex.sell, 1) AS exchangerate, a.curr,
// 		       (s.nextdate IS NULL OR s.nextdate > s.enddate) AS expired,
// 		       d.description AS department
// 		FROM recurring s
// 		JOIN ap a ON (a.id = s.id)
// 		LEFT JOIN department d ON (d.id = a.department_id)
// 		JOIN vendor n ON (n.id = a.vendor_id)
// 		LEFT JOIN recurringemail se ON (se.id = s.id)
// 		LEFT JOIN recurringprint sp ON (sp.id = s.id)
// 		LEFT JOIN exchangerate ex ON (ex.curr = a.curr AND a.transdate = ex.transdate)
//
// 		UNION
//
// 		SELECT 'gl' AS module, 'gl' AS transaction, false AS invoice,
// 		       a.description, '' AS name, '' AS vcnumber, 0 AS name_id,
// 		       (SELECT SUM(ac.amount) FROM acc_trans ac WHERE ac.trans_id = a.id AND ac.amount > 0) AS amount,
// 		       s.*, se.formname AS recurringemail,
// 		       sp.formname AS recurringprint,
// 		       EXTRACT(DAY FROM (s.nextdate - CURRENT_DATE)) AS overdue, '' AS vc,
// 		       1 AS exchangerate, '%s' AS curr,
// 		       (s.nextdate IS NULL OR s.nextdate > s.enddate) AS expired,
// 		       d.description AS department
// 		FROM recurring s
// 		JOIN gl a ON (a.id = s.id)
// 		LEFT JOIN department d ON (d.id = a.department_id)
// 		LEFT JOIN recurringemail se ON (se.id = s.id)
// 		LEFT JOIN recurringprint sp ON (sp.id = s.id)
//
// 		UNION
//
// 		SELECT 'oe' AS module, 'so' AS transaction, false AS invoice,
// 		       a.description, n.name, n.customernumber AS vcnumber,
// 		       n.id AS name_id, a.amount, s.*, se.formname AS recurringemail,
// 		       sp.formname AS recurringprint,
// 		       EXTRACT(DAY FROM (s.nextdate - CURRENT_DATE)) AS overdue, 'customer' AS vc,
// 		       COALESCE(ex.buy, 1) AS exchangerate, a.curr,
// 		       (s.nextdate IS NULL OR s.nextdate > s.enddate) AS expired,
// 		       d.description AS department
// 		FROM recurring s
// 		JOIN oe a ON (a.id = s.id)
// 		LEFT JOIN department d ON (d.id = a.department_id)
// 		JOIN customer n ON (n.id = a.customer_id)
// 		LEFT JOIN recurringemail se ON (se.id = s.id)
// 		LEFT JOIN recurringprint sp ON (sp.id = s.id)
// 		LEFT JOIN exchangerate ex ON (ex.curr = a.curr AND a.transdate = ex.transdate)
// 		WHERE a.quotation = false
//
// 		UNION
//
// 		SELECT 'oe' AS module, 'po' AS transaction, false AS invoice,
// 		       a.description, n.name, n.vendornumber AS vcnumber,
// 		       n.id AS name_id, a.amount, s.*, se.formname AS recurringemail,
// 		       sp.formname AS recurringprint,
// 		       EXTRACT(DAY FROM (s.nextdate - CURRENT_DATE)) AS overdue, 'vendor' AS vc,
// 		       COALESCE(ex.sell, 1) AS exchangerate, a.curr,
// 		       (s.nextdate IS NULL OR s.nextdate > s.enddate) AS expired,
// 		       d.description AS department
// 		FROM recurring s
// 		JOIN oe a ON (a.id = s.id)
// 		LEFT JOIN department d ON (d.id = a.department_id)
// 		JOIN vendor n ON (n.id = a.vendor_id)
// 		LEFT JOIN recurringemail se ON (se.id = s.id)
// 		LEFT JOIN recurringprint sp ON (sp.id = s.id)
// 		LEFT JOIN exchangerate ex ON (ex.curr = a.curr AND a.transdate = ex.transdate)
// 		WHERE a.quotation = false
//
// 		ORDER BY %s`, defaultCurrency, sortOrder)
// }
//
// func (rs *RecurringService) processRecurringRows(rows *sql.Rows, form *Form) error {
// 	var currentID int
// 	var currentTransaction string
// 	var currentIndex int
// 	emailMap := make(map[string]bool)
// 	printMap := make(map[string]bool)
//
// 	for rows.Next() {
// 		var rt RecurringTransaction
// 		var recurringEmail, recurringPrint sql.NullString
//
// 		err := rows.Scan(
// 			&rt.Module, &rt.Transaction, &rt.Invoice, &rt.Description,
// 			&rt.Name, &rt.VCNumber, &rt.NameID, &rt.Amount,
// 			&rt.ID, &rt.NextDate, &rt.EndDate, &rt.Repeat, &rt.Unit,
// 			&recurringEmail, &recurringPrint, &rt.Overdue, &rt.VC,
// 			&rt.ExchangeRate, &rt.Curr, &rt.Expired, &rt.Department,
// 		)
// 		if err != nil {
// 			return err
// 		}
//
// 		// Set default exchange rate
// 		if rt.ExchangeRate == 0 {
// 			rt.ExchangeRate = 1
// 		}
//
// 		// Handle grouping logic
// 		if rt.ID != currentID {
// 			// Process previous record's email/print data
// 			if currentID != 0 {
// 				rs.finalizeEmailPrint(form, currentTransaction, currentIndex, emailMap, printMap)
// 			}
//
// 			// Reset for new record
// 			emailMap = make(map[string]bool)
// 			printMap = make(map[string]bool)
//
// 			// Add new transaction
// 			if form.Transactions[rt.Transaction] == nil {
// 				form.Transactions[rt.Transaction] = []RecurringTransaction{}
// 			}
// 			form.Transactions[rt.Transaction] = append(form.Transactions[rt.Transaction], rt)
//
// 			currentID = rt.ID
// 			currentTransaction = rt.Transaction
// 			currentIndex = len(form.Transactions[rt.Transaction]) - 1
// 		}
//
// 		// Collect email and print form names
// 		if recurringEmail.Valid && recurringEmail.String != "" {
// 			emailMap[recurringEmail.String] = true
// 		}
// 		if recurringPrint.Valid && recurringPrint.String != "" {
// 			printMap[recurringPrint.String] = true
// 		}
// 	}
//
// 	// Handle last record
// 	if currentID != 0 {
// 		rs.finalizeEmailPrint(form, currentTransaction, currentIndex, emailMap, printMap)
// 	}
//
// 	return rows.Err()
// }
//
// func (rs *RecurringService) finalizeEmailPrint(form *Form, transaction string, index int, emailMap, printMap map[string]bool) {
// 	// Build email string
// 	var emailParts []string
// 	for email := range emailMap {
// 		emailParts = append(emailParts, email)
// 	}
// 	form.Transactions[transaction][index].RecurringEmail = strings.Join(emailParts, ":")
//
// 	// Build print string
// 	var printParts []string
// 	for print := range printMap {
// 		printParts = append(printParts, print)
// 	}
// 	form.Transactions[transaction][index].RecurringPrint = strings.Join(printParts, ":")
// }
//
// func (rs *RecurringService) scanRecurringDetails(row *sql.Row, details *RecurringDetails) error {
// 	return row.Scan(
// 		&details.ID, &details.NextDate, &details.EndDate, &details.Repeat, &details.Unit,
// 		&details.ARID, &details.ARInvoice, &details.APID, &details.APInvoice,
// 		&details.Overdue, &details.Paid, &details.Req, &details.OEID,
// 		&details.CustomerID, &details.VendorID,
// 	)
// }
//
// type EmailDetails struct {
// 	Email   string
// 	Message string
// }
//
// func (rs *RecurringService) getRecurringEmailDetails(id int) (*EmailDetails, error) {
// 	query := `SELECT formname, format FROM recurringemail WHERE id = $1`
// 	rows, err := rs.db.Query(query, id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var parts []string
// 	var message string
//
// 	for rows.Next() {
// 		var formname, format string
// 		if err := rows.Scan(&formname, &format); err != nil {
// 			return nil, err
// 		}
// 		parts = append(parts, fmt.Sprintf("%s:%s", formname, format))
// 		// In the original, message was set to the last message found
// 		// This is a simplification
// 	}
//
// 	return &EmailDetails{
// 		Email:   strings.Join(parts, ":"),
// 		Message: message,
// 	}, nil
// }
//
// func (rs *RecurringService) getRecurringPrintDetails(id int) (string, error) {
// 	query := `SELECT formname, format, printer FROM recurringprint WHERE id = $1`
// 	rows, err := rs.db.Query(query, id)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer rows.Close()
//
// 	var parts []string
// 	for rows.Next() {
// 		var formname, format, printer string
// 		if err := rows.Scan(&formname, &format, &printer); err != nil {
// 			return "", err
// 		}
// 		parts = append(parts, fmt.Sprintf("%s:%s:%s", formname, format, printer))
// 	}
//
// 	return strings.Join(parts, ":"), nil
// }
//
// // Example usage
// func Example() {
// 	// This would typically be set up with your actual database connection
// 	// db, err := sql.Open("postgres", "your-connection-string")
// 	// if err != nil {
// 	//     log.Fatal(err)
// 	// }
// 	// defer db.Close()
//
// 	// service := NewRecurringService(db)
// 	// config := &Config{
// 	//     DBDriver: "postgres",
// 	//     DateFormat: "YYYY-MM-DD",
// 	// }
// 	// form := &Form{}
//
// 	// err = service.GetRecurringTransactions(config, form)
// 	// if err != nil {
// 	//     log.Fatal(err)
// 	// }
//
// 	log.Println("Recurring transactions service ready")
// }
