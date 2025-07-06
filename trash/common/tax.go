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
// // TaxManager handles tax-related database operations
// type TaxManager struct {
// 	db *sql.DB
// }
//
// // TaxRate represents a tax rate record
// type TaxRate struct {
// 	ID              int       `json:"id" db:"id"`
// 	Accno           string    `json:"accno" db:"accno"`
// 	Description     string    `json:"description" db:"description"`
// 	Rate            float64   `json:"rate" db:"rate"`
// 	TaxNumber       string    `json:"taxnumber" db:"taxnumber"`
// 	ValidTo         *time.Time `json:"validto" db:"validto"`
// 	VATKey          string    `json:"vatkey" db:"vatkey"`
// 	FormDigit       int       `json:"formdigit" db:"formdigit"`
// 	ValidFrom       *time.Time `json:"validfrom" db:"validfrom"`
// 	Translation     string    `json:"translation" db:"translation"`
// 	ReverseChargeID *int      `json:"reversecharge_id" db:"reversecharge_id"`
// 	ReverseCharge   string    `json:"reversecharge" db:"reversecharge"`
// }
//
// // Form represents the form data structure
// type Form struct {
// 	TaxRates     []TaxRate         `json:"taxrates"`
// 	TaxAccounts  string            `json:"taxaccounts"`
// 	Fields       map[string]string `json:"fields"`
// }
//
// // MyConfig represents database configuration
// type MyConfig struct {
// 	CountryCode string `json:"countrycode"`
// }
//
// // NewTaxManager creates a new TaxManager instance
// func NewTaxManager(db *sql.DB) *TaxManager {
// 	return &TaxManager{db: db}
// }
//
// // Taxes retrieves all tax rates from the database
// func (tm *TaxManager) Taxes(myConfig *MyConfig, form *Form) error {
// 	query := `SELECT c.id, c.accno, c.description,
// 		t.rate * 100 AS rate, t.taxnumber, t.validto,
// 		t.vatkey, t.formdigit, t.validfrom,
// 		COALESCE(l.description, '') AS translation,
// 		t.reversecharge_id,
// 		COALESCE(c2.accno, '') AS reversecharge
// 		FROM chart c
// 		JOIN tax t ON (c.id = t.chart_id)
// 		LEFT JOIN chart c2 ON c2.id = t.reversecharge_id
// 		LEFT JOIN translation l ON (l.trans_id = c.id AND l.language_code = ?)
// 		ORDER BY c.accno, t.validto`
//
// 	rows, err := tm.db.Query(query, myConfig.CountryCode)
// 	if err != nil {
// 		return fmt.Errorf("failed to execute taxes query: %w", err)
// 	}
// 	defer rows.Close()
//
// 	var taxRates []TaxRate
// 	for rows.Next() {
// 		var tr TaxRate
// 		var validTo, validFrom sql.NullTime
// 		var reverseChargeID sql.NullInt64
//
// 		err := rows.Scan(
// 			&tr.ID,
// 			&tr.Accno,
// 			&tr.Description,
// 			&tr.Rate,
// 			&tr.TaxNumber,
// 			&validTo,
// 			&tr.VATKey,
// 			&tr.FormDigit,
// 			&validFrom,
// 			&tr.Translation,
// 			&reverseChargeID,
// 			&tr.ReverseCharge,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("failed to scan tax rate row: %w", err)
// 		}
//
// 		// Handle nullable timestamps
// 		if validTo.Valid {
// 			tr.ValidTo = &validTo.Time
// 		}
// 		if validFrom.Valid {
// 			tr.ValidFrom = &validFrom.Time
// 		}
// 		if reverseChargeID.Valid {
// 			id := int(reverseChargeID.Int64)
// 			tr.ReverseChargeID = &id
// 		}
//
// 		// Use translation if available, otherwise use description
// 		if tr.Translation != "" {
// 			tr.Description = tr.Translation
// 		}
//
// 		taxRates = append(taxRates, tr)
// 	}
//
// 	if err := rows.Err(); err != nil {
// 		return fmt.Errorf("error iterating tax rate rows: %w", err)
// 	}
//
// 	form.TaxRates = taxRates
// 	return nil
// }
//
// // SaveTaxes saves tax rates to the database
// func (tm *TaxManager) SaveTaxes(myConfig *MyConfig, form *Form) error {
// 	// Begin transaction
// 	tx, err := tm.db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback()
//
// 	// Delete all existing tax records
// 	if _, err := tx.Exec("DELETE FROM tax"); err != nil {
// 		return fmt.Errorf("failed to delete existing tax records: %w", err)
// 	}
//
// 	// Process each tax account
// 	if form.TaxAccounts != "" {
// 		taxAccounts := strings.Split(form.TaxAccounts, " ")
// 		for _, item := range taxAccounts {
// 			item = strings.TrimSpace(item)
// 			if item == "" {
// 				continue
// 			}
//
// 			if err := tm.processTaxAccount(tx, myConfig, form, item); err != nil {
// 				return fmt.Errorf("failed to process tax account %s: %w", item, err)
// 			}
// 		}
// 	}
//
// 	// Commit transaction
// 	if err := tx.Commit(); err != nil {
// 		return fmt.Errorf("failed to commit transaction: %w", err)
// 	}
//
// 	return nil
// }
//
// // processTaxAccount processes a single tax account entry
// func (tm *TaxManager) processTaxAccount(tx *sql.Tx, myConfig *MyConfig, form *Form, item string) error {
// 	parts := strings.Split(item, "_")
// 	if len(parts) != 2 {
// 		return fmt.Errorf("invalid tax account format: %s", item)
// 	}
//
// 	chartID, err := strconv.Atoi(parts[0])
// 	if err != nil {
// 		return fmt.Errorf("invalid chart_id in %s: %w", item, err)
// 	}
//
// 	index := parts[1]
//
// 	// Get tax rate for this index
// 	taxRateKey := fmt.Sprintf("taxrate_%s", index)
// 	taxRateStr, exists := form.Fields[taxRateKey]
// 	if !exists || taxRateStr == "" {
// 		// Skip if no tax rate provided
// 		log.Printf("Index %s: tax rate empty or missing", index)
// 		return nil
// 	}
//
// 	log.Printf("Index %s: tax rate %s", index, taxRateStr)
//
// 	// Parse tax rate
// 	rate, err := tm.parseAmount(myConfig, taxRateStr)
// 	if err != nil {
// 		return fmt.Errorf("failed to parse tax rate %s: %w", taxRateStr, err)
// 	}
// 	rate = rate / 100 // Convert percentage to decimal
//
// 	// Get other tax fields
// 	taxNumber := form.Fields[fmt.Sprintf("taxnumber_%s", index)]
// 	vatKey := form.Fields[fmt.Sprintf("vatkey_%s", index)]
// 	validFromStr := form.Fields[fmt.Sprintf("validfrom_%s", index)]
// 	validToStr := form.Fields[fmt.Sprintf("validto_%s", index)]
// 	reverseChargeStr := form.Fields[fmt.Sprintf("reversecharge_%s", index)]
// 	formDigitStr := form.Fields[fmt.Sprintf("formdigit_%s", index)]
//
// 	// Parse form digit
// 	var formDigit int
// 	if formDigitStr != "" {
// 		if fd, err := strconv.Atoi(formDigitStr); err == nil {
// 			formDigit = fd
// 		}
// 	}
//
// 	// Get reverse charge ID
// 	var reverseChargeID *int
// 	if reverseChargeStr != "" {
// 		var rcID sql.NullInt64
// 		query := "SELECT id FROM chart WHERE accno = ?"
// 		err := tx.QueryRow(query, reverseChargeStr).Scan(&rcID)
// 		if err != nil && err != sql.ErrNoRows {
// 			return fmt.Errorf("failed to get reverse charge ID for %s: %w", reverseChargeStr, err)
// 		}
// 		if rcID.Valid {
// 			id := int(rcID.Int64)
// 			reverseChargeID = &id
// 		}
// 	}
//
// 	// Parse dates
// 	var validFrom, validTo *time.Time
// 	if validFromStr != "" {
// 		if date, err := tm.parseDate(validFromStr); err == nil {
// 			validFrom = &date
// 		}
// 	}
// 	if validToStr != "" {
// 		if date, err := tm.parseDate(validToStr); err == nil {
// 			validTo = &date
// 		}
// 	}
//
// 	// Insert tax record
// 	insertQuery := `INSERT INTO tax (chart_id, rate, taxnumber, vatkey, validfrom, validto, reversecharge_id, formdigit)
// 		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
//
// 	_, err = tx.Exec(insertQuery, chartID, rate, taxNumber, vatKey, validFrom, validTo, reverseChargeID, formDigit)
// 	if err != nil {
// 		return fmt.Errorf("failed to insert tax record: %w", err)
// 	}
//
// 	return nil
// }
//
// // parseAmount parses a string amount considering locale settings
// func (tm *TaxManager) parseAmount(myConfig *MyConfig, amountStr string) (float64, error) {
// 	// Simple implementation - in a real system, this would handle locale-specific formatting
// 	// Remove common currency symbols and spaces
// 	cleaned := strings.ReplaceAll(amountStr, " ", "")
// 	cleaned = strings.ReplaceAll(cleaned, ",", ".")
//
// 	// Handle percentage symbol
// 	cleaned = strings.ReplaceAll(cleaned, "%", "")
//
// 	amount, err := strconv.ParseFloat(cleaned, 64)
// 	if err != nil {
// 		return 0, fmt.Errorf("invalid amount format: %s", amountStr)
// 	}
//
// 	return amount, nil
// }
//
// // parseDate parses a date string - simplified implementation
// func (tm *TaxManager) parseDate(dateStr string) (time.Time, error) {
// 	// Try common date formats
// 	formats := []string{
// 		"2006-01-02",
// 		"01/02/2006",
// 		"02.01.2006",
// 		"2006-01-02 15:04:05",
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
// // GetTaxRateByAccount retrieves a specific tax rate by account number
// func (tm *TaxManager) GetTaxRateByAccount(accno string) (*TaxRate, error) {
// 	query := `SELECT c.id, c.accno, c.description,
// 		t.rate * 100 AS rate, t.taxnumber, t.validto,
// 		t.vatkey, t.formdigit, t.validfrom,
// 		t.reversecharge_id,
// 		COALESCE(c2.accno, '') AS reversecharge
// 		FROM chart c
// 		JOIN tax t ON (c.id = t.chart_id)
// 		LEFT JOIN chart c2 ON c2.id = t.reversecharge_id
// 		WHERE c.accno = ?
// 		ORDER BY t.validto DESC
// 		LIMIT 1`
//
// 	var tr TaxRate
// 	var validTo, validFrom sql.NullTime
// 	var reverseChargeID sql.NullInt64
//
// 	err := tm.db.QueryRow(query, accno).Scan(
// 		&tr.ID,
// 		&tr.Accno,
// 		&tr.Description,
// 		&tr.Rate,
// 		&tr.TaxNumber,
// 		&validTo,
// 		&tr.VATKey,
// 		&tr.FormDigit,
// 		&validFrom,
// 		&reverseChargeID,
// 		&tr.ReverseCharge,
// 	)
//
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("tax rate not found for account: %s", accno)
// 		}
// 		return nil, fmt.Errorf("failed to get tax rate: %w", err)
// 	}
//
// 	// Handle nullable fields
// 	if validTo.Valid {
// 		tr.ValidTo = &validTo.Time
// 	}
// 	if validFrom.Valid {
// 		tr.ValidFrom = &validFrom.Time
// 	}
// 	if reverseChargeID.Valid {
// 		id := int(reverseChargeID.Int64)
// 		tr.ReverseChargeID = &id
// 	}
//
// 	return &tr, nil
// }
//
// // ValidateTaxRate validates a tax rate for a given date
// func (tm *TaxManager) ValidateTaxRate(accno string, date time.Time) (bool, error) {
// 	taxRate, err := tm.GetTaxRateByAccount(accno)
// 	if err != nil {
// 		return false, err
// 	}
//
// 	// Check if the date falls within the valid range
// 	if taxRate.ValidFrom != nil && date.Before(*taxRate.ValidFrom) {
// 		return false, nil
// 	}
// 	if taxRate.ValidTo != nil && date.After(*taxRate.ValidTo) {
// 		return false, nil
// 	}
//
// 	return true, nil
// }
//
// // Example usage and testing
