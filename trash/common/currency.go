package common

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"strconv"
// 	"strings"
// 	"time"
// )
//
// // ExchangeRateManager handles exchange rate database operations
// type ExchangeRateManager struct {
// 	db *sql.DB
// }
//
// // ExchangeRate represents an exchange rate record
// type ExchangeRate struct {
// 	ID        int       `json:"id" db:"id"`
// 	TransDate time.Time `json:"transdate" db:"transdate"`
// 	Buy       float64   `json:"buy" db:"buy"`
// 	Sell      float64   `json:"sell" db:"sell"`
// 	Currency  string    `json:"curr" db:"curr"`
// }
//
// // Form represents the form data structure for exchange rate operations
// type Form struct {
// 	// Filter parameters
// 	Year            int    `json:"year"`            // Year for filtering exchange rates
// 	Month           int    `json:"month"`           // Month for filtering exchange rates
// 	Interval        string `json:"interval"`        // Interval type (monthly, quarterly, etc.)
// 	Currency        string `json:"currency"`        // Currency code for filtering
// 	TransDateFrom   string `json:"transdatefrom"`   // Start date for filtering
// 	TransDateTo     string `json:"transdateto"`     // End date for filtering
// 	TransDate       string `json:"transdate"`       // Transaction date for saving
//
// 	// Results
// 	Currencies   []string       `json:"currencies"`   // List of available currencies
// 	Years        []int          `json:"years"`        // List of available years
// 	Transactions []ExchangeRate `json:"transactions"` // Exchange rate records
//
// 	// Dynamic fields for currency rates
// 	Fields map[string]string `json:"fields"` // Dynamic form fields (e.g., "USDrate", "USDbuy", "USDsell")
// }
//
// // MyConfig represents database and system configuration
// type MyConfig struct {
// 	CountryCode   string `json:"countrycode"`   // Country code for localization
// 	DateFormat    string `json:"dateformat"`    // Date format preference
// 	NumberFormat  string `json:"numberformat"`  // Number format preference
// 	DecimalPlaces int    `json:"decimalplaces"` // Decimal places for amounts
// }
//
// // NewExchangeRateManager creates a new ExchangeRateManager instance
// func NewExchangeRateManager(db *sql.DB) *ExchangeRateManager {
// 	return &ExchangeRateManager{db: db}
// }
//
// // ExchangeRates retrieves basic exchange rate setup data
// // This method loads available currencies and years for the exchange rate interface
// func (erm *ExchangeRateManager) ExchangeRates(myConfig *MyConfig, form *Form) error {
// 	// Get list of available currencies from the database
// 	currencies, err := erm.getCurrencies(myConfig)
// 	if err != nil {
// 		return fmt.Errorf("failed to get currencies: %w", err)
// 	}
// 	form.Currencies = currencies
//
// 	// Get all available years from exchange rate records
// 	years, err := erm.getAllYears()
// 	if err != nil {
// 		return fmt.Errorf("failed to get available years: %w", err)
// 	}
// 	form.Years = years
//
// 	return nil
// }
//
// // GetExchangeRates retrieves exchange rates based on filter criteria
// // This method implements filtering by date range and currency
// func (erm *ExchangeRateManager) GetExchangeRates(myConfig *MyConfig, form *Form) error {
// 	// Initialize the WHERE clause - always true condition as base
// 	where := "1 = 1"
// 	var args []interface{}
// 	argIndex := 0
//
// 	// Get available currencies for the form
// 	currencies, err := erm.getCurrencies(myConfig)
// 	if err != nil {
// 		return fmt.Errorf("failed to get currencies: %w", err)
// 	}
// 	form.Currencies = currencies
//
// 	// Calculate date range if year and month are provided
// 	if form.Year > 0 && form.Month > 0 {
// 		dateFrom, dateTo, err := erm.calculateDateRange(form.Year, form.Month, form.Interval)
// 		if err != nil {
// 			return fmt.Errorf("failed to calculate date range: %w", err)
// 		}
// 		form.TransDateFrom = dateFrom
// 		form.TransDateTo = dateTo
// 	}
//
// 	// Add date range filters to WHERE clause
// 	if form.TransDateFrom != "" {
// 		where += " AND transdate >= ?"
// 		args = append(args, form.TransDateFrom)
// 		argIndex++
// 	}
// 	if form.TransDateTo != "" {
// 		where += " AND transdate <= ?"
// 		args = append(args, form.TransDateTo)
// 		argIndex++
// 	}
//
// 	// Add currency filter if specified
// 	if form.Currency != "" {
// 		where += " AND curr = ?"
// 		args = append(args, form.Currency)
// 		argIndex++
// 	}
//
// 	// Build the complete query with sorting
// 	// Default sort order is by transaction date (ascending)
// 	query := fmt.Sprintf(`SELECT id, transdate, buy, sell, curr
// 		FROM exchangerate
// 		WHERE %s
// 		ORDER BY transdate ASC`, where)
//
// 	// Execute the query
// 	rows, err := erm.db.Query(query, args...)
// 	if err != nil {
// 		return fmt.Errorf("failed to execute exchange rates query: %w", err)
// 	}
// 	defer rows.Close()
//
// 	// Process the results
// 	var exchangeRates []ExchangeRate
// 	for rows.Next() {
// 		var er ExchangeRate
// 		err := rows.Scan(&er.ID, &er.TransDate, &er.Buy, &er.Sell, &er.Currency)
// 		if err != nil {
// 			return fmt.Errorf("failed to scan exchange rate row: %w", err)
// 		}
// 		exchangeRates = append(exchangeRates, er)
// 	}
//
// 	// Check for any iteration errors
// 	if err := rows.Err(); err != nil {
// 		return fmt.Errorf("error iterating exchange rate rows: %w", err)
// 	}
//
// 	// Store results in form
// 	form.Transactions = exchangeRates
// 	return nil
// }
//
// // SaveExchangeRate saves exchange rates for a specific date
// // This method handles multiple currencies and uses transactions for data consistency
// func (erm *ExchangeRateManager) SaveExchangeRate(myConfig *MyConfig, form *Form) error {
// 	// Begin database transaction for atomicity
// 	tx, err := erm.db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	// Ensure rollback on error
// 	defer tx.Rollback()
//
// 	// Parse the transaction date
// 	transDate, err := erm.parseDate(form.TransDate)
// 	if err != nil {
// 		return fmt.Errorf("failed to parse transaction date %s: %w", form.TransDate, err)
// 	}
//
// 	// Prepare the DELETE statement to remove existing rates for the date
// 	deleteQuery := "DELETE FROM exchangerate WHERE transdate = ? AND curr = ?"
// 	deleteStmt, err := tx.Prepare(deleteQuery)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare delete statement: %w", err)
// 	}
// 	defer deleteStmt.Close()
//
// 	// Prepare the INSERT statement for new exchange rates
// 	insertQuery := "INSERT INTO exchangerate (transdate, buy, sell, curr) VALUES (?, ?, ?, ?)"
// 	insertStmt, err := tx.Prepare(insertQuery)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare insert statement: %w", err)
// 	}
// 	defer insertStmt.Close()
//
// 	// Process each currency from the currencies list
// 	// The currencies are stored as a colon-separated string in the original Perl
// 	if len(form.Currencies) > 0 {
// 		for _, currency := range form.Currencies {
// 			currency = strings.TrimSpace(currency)
// 			if currency == "" {
// 				continue
// 			}
//
// 			// Check if this currency has data in the form
// 			// Look for currency-specific fields like "USDrate", "USDbuy", "USDsell"
// 			rateKey := currency // Base currency key
// 			if _, exists := form.Fields[rateKey]; !exists {
// 				// Skip if no rate data for this currency
// 				continue
// 			}
//
// 			// Delete existing exchange rate for this date and currency
// 			_, err := deleteStmt.Exec(transDate, currency)
// 			if err != nil {
// 				return fmt.Errorf("failed to delete existing rate for %s: %w", currency, err)
// 			}
//
// 			// Parse buy and sell rates for this currency
// 			buyKey := fmt.Sprintf("%sbuy", currency)
// 			sellKey := fmt.Sprintf("%ssell", currency)
//
// 			buyRate, err := erm.parseAmount(myConfig, form.Fields[buyKey])
// 			if err != nil {
// 				log.Printf("Warning: failed to parse buy rate for %s: %v", currency, err)
// 				buyRate = 0
// 			}
//
// 			sellRate, err := erm.parseAmount(myConfig, form.Fields[sellKey])
// 			if err != nil {
// 				log.Printf("Warning: failed to parse sell rate for %s: %v", currency, err)
// 				sellRate = 0
// 			}
//
// 			// Only insert if we have at least one non-zero rate
// 			if buyRate != 0 || sellRate != 0 {
// 				_, err := insertStmt.Exec(transDate, buyRate, sellRate, currency)
// 				if err != nil {
// 					return fmt.Errorf("failed to insert exchange rate for %s: %w", currency, err)
// 				}
// 				log.Printf("Saved exchange rate for %s: buy=%.6f, sell=%.6f", currency, buyRate, sellRate)
// 			}
// 		}
// 	}
//
// 	// Commit the transaction
// 	if err := tx.Commit(); err != nil {
// 		return fmt.Errorf("failed to commit exchange rate transaction: %w", err)
// 	}
//
// 	return nil
// }
//
// // getCurrencies retrieves the list of available currencies from the database
// func (erm *ExchangeRateManager) getCurrencies(myConfig *MyConfig) ([]string, error) {
// 	// This is a simplified implementation - in a real system, this might query
// 	// a currencies table, chart of accounts, or configuration table
// 	query := "SELECT DISTINCT curr FROM exchangerate WHERE curr IS NOT NULL ORDER BY curr"
//
// 	rows, err := erm.db.Query(query)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query currencies: %w", err)
// 	}
// 	defer rows.Close()
//
// 	var currencies []string
// 	for rows.Next() {
// 		var currency string
// 		if err := rows.Scan(&currency); err != nil {
// 			return nil, fmt.Errorf("failed to scan currency: %w", err)
// 		}
// 		currencies = append(currencies, currency)
// 	}
//
// 	return currencies, rows.Err()
// }
//
// // getAllYears retrieves all years that have exchange rate data
// func (erm *ExchangeRateManager) getAllYears() ([]int, error) {
// 	query := `SELECT DISTINCT EXTRACT(YEAR FROM transdate) AS year
// 		FROM exchangerate
// 		ORDER BY year DESC`
//
// 	rows, err := erm.db.Query(query)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query years: %w", err)
// 	}
// 	defer rows.Close()
//
// 	var years []int
// 	for rows.Next() {
// 		var year int
// 		if err := rows.Scan(&year); err != nil {
// 			return nil, fmt.Errorf("failed to scan year: %w", err)
// 		}
// 		years = append(years, year)
// 	}
//
// 	return years, rows.Err()
// }
//
// // calculateDateRange calculates the date range based on year, month, and interval
// func (erm *ExchangeRateManager) calculateDateRange(year, month int, interval string) (string, string, error) {
// 	// Create the start date (first day of the month)
// 	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
//
// 	var endDate time.Time
//
// 	// Calculate end date based on interval
// 	switch strings.ToLower(interval) {
// 	case "monthly", "month", "":
// 		// End of the month
// 		endDate = startDate.AddDate(0, 1, -1)
// 	case "quarterly", "quarter":
// 		// End of the quarter
// 		endDate = startDate.AddDate(0, 3, -1)
// 	case "yearly", "year":
// 		// End of the year
// 		endDate = startDate.AddDate(1, 0, -1)
// 	default:
// 		// Default to monthly
// 		endDate = startDate.AddDate(0, 1, -1)
// 	}
//
// 	// Format dates as strings (YYYY-MM-DD format)
// 	dateFrom := startDate.Format("2006-01-02")
// 	dateTo := endDate.Format("2006-01-02")
//
// 	return dateFrom, dateTo, nil
// }
//
// // parseDate parses a date string using multiple common formats
// func (erm *ExchangeRateManager) parseDate(dateStr string) (time.Time, error) {
// 	// Try various date formats
// 	formats := []string{
// 		"2006-01-02",           // ISO format
// 		"01/02/2006",           // US format
// 		"02/01/2006",           // European format
// 		"02.01.2006",           // German format
// 		"2006-01-02 15:04:05",  // ISO with time
// 	}
//
// 	for _, format := range formats {
// 		if date, err := time.Parse(format, dateStr); err == nil {
// 			return date, nil
// 		}
// 	}
//
// 	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
// }
//
// // parseAmount parses a string amount considering locale settings
// func (erm *ExchangeRateManager) parseAmount(myConfig *MyConfig, amountStr string) (float64, error) {
// 	if amountStr == "" {
// 		return 0, nil
// 	}
//
// 	// Clean the amount string
// 	cleaned := strings.TrimSpace(amountStr)
//
// 	// Handle different decimal separators based on locale
// 	if myConfig.NumberFormat == "european" {
// 		// European format: 1.234,56
// 		// Replace comma with dot for parsing
// 		if strings.Contains(cleaned, ",") && strings.Contains(cleaned, ".") {
// 			// Remove thousand separators (dots) and replace decimal comma
// 			cleaned = strings.ReplaceAll(cleaned, ".", "")
// 			cleaned = strings.ReplaceAll(cleaned, ",", ".")
// 		} else if strings.Contains(cleaned, ",") {
// 			// Only comma present, assume it's decimal separator
// 			cleaned = strings.ReplaceAll(cleaned, ",", ".")
// 		}
// 	} else {
// 		// US format: 1,234.56
// 		// Remove thousand separators (commas) but keep decimal dot
// 		parts := strings.Split(cleaned, ".")
// 		if len(parts) == 2 {
// 			// Has decimal part
// 			integerPart := strings.ReplaceAll(parts[0], ",", "")
// 			cleaned = integerPart + "." + parts[1]
// 		} else {
// 			// No decimal part, remove all commas
// 			cleaned = strings.ReplaceAll(cleaned, ",", "")
// 		}
// 	}
//
// 	// Parse the cleaned amount
// 	amount, err := strconv.ParseFloat(cleaned, 64)
// 	if err != nil {
// 		return 0, fmt.Errorf("invalid amount format: %s", amountStr)
// 	}
//
// 	return amount, nil
// }
//
// // GetExchangeRateByDate retrieves the exchange rate for a specific currency and date
// func (erm *ExchangeRateManager) GetExchangeRateByDate(currency string, date time.Time) (*ExchangeRate, error) {
// 	query := `SELECT id, transdate, buy, sell, curr
// 		FROM exchangerate
// 		WHERE curr = ? AND transdate = ?`
//
// 	var er ExchangeRate
// 	err := erm.db.QueryRow(query, currency, date.Format("2006-01-02")).Scan(
// 		&er.ID, &er.TransDate, &er.Buy, &er.Sell, &er.Currency)
//
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("exchange rate not found for %s on %s", currency, date.Format("2006-01-02"))
// 		}
// 		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
// 	}
//
// 	return &er, nil
// }
//
// // GetLatestExchangeRate retrieves the most recent exchange rate for a currency
// func (erm *ExchangeRateManager) GetLatestExchangeRate(currency string) (*ExchangeRate, error) {
// 	query := `SELECT id, transdate, buy, sell, curr
// 		FROM exchangerate
// 		WHERE curr = ?
// 		ORDER BY transdate DESC
// 		LIMIT 1`
//
// 	var er ExchangeRate
// 	err := erm.db.QueryRow(query, currency).Scan(
// 		&er.ID, &er.TransDate, &er.Buy, &er.Sell, &er.Currency)
//
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("no exchange rate found for currency: %s", currency)
// 		}
// 		return nil, fmt.Errorf("failed to get latest exchange rate: %w", err)
// 	}
//
// 	return &er, nil
// }
//
// // Example usage and testing
// func Example() {
// 	// This would typically be initialized with your actual database connection
// 	// db, err := sql.Open("your-driver", "your-connection-string")
// 	// if err != nil {
// 	//     log.Fatal(err)
// 	// }
// 	// defer db.Close()
//
// 	// erm := NewExchangeRateManager(db)
//
// 	// Example form and config
// 	form := &Form{
// 		Fields: make(map[string]string),
// 	}
//
// 	myConfig := &MyConfig{
// 		CountryCode:   "US",
// 		DateFormat:    "MM/DD/YYYY",
// 		NumberFormat:  "us",
// 		DecimalPlaces: 6,
// 	}
//
// 	// Example usage:
// 	// err = erm.ExchangeRates(myConfig, form)
// 	// err = erm.GetExchangeRates(myConfig, form)
// 	// err = erm.SaveExchangeRate(myConfig, form)
//
// 	log.Println("Exchange rate management implementation ready")
// }
