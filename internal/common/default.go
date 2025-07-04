package common

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"strings"
// )
//
// // AccountingDefaults handles default account operations
// type AccountingDefaults struct {
// 	db *sql.DB
// }
//
// // Form represents the form data structure
// type Form struct {
// 	IC               string
// 	ICIncome         string
// 	ICExpense        string
// 	FXGain           string
// 	FXLoss           string
// 	InventoryAccno   string
// 	IncomeAccno      string
// 	ExpenseAccno     string
// 	FXGainAccno      string
// 	FXLossAccno      string
// 	DBVersion        string
// 	Optional         string
// 	Defaults         map[string]interface{}
// 	Accno            map[string]map[string]AccountInfo
//
// 	// Additional fields that appear in the required fields loop
// 	TransitionAccount  string
// 	SelectedAccount    string
// 	GLNumber          string
// 	SINumber          string
// 	VINumber          string
// 	BatchNumber       string
// 	VoucherNumber     string
// 	SONumber          string
// 	PONumber          string
// 	SQNumber          string
// 	RFQNumber         string
// 	PartNumber        string
// 	EmployeeNumber    string
// 	CustomerNumber    string
// 	VendorNumber      string
// 	ProjectNumber     string
// 	Precision         string
//
// 	// Dynamic fields map for optional fields
// 	Fields map[string]string
// }
//
// // MyConfig represents database configuration
// type MyConfig struct {
// 	CountryCode string
// 	// Add other config fields as needed
// }
//
// // AccountInfo holds account information
// type AccountInfo struct {
// 	ID          int    `json:"id"`
// 	Description string `json:"description"`
// }
//
// // ChartRow represents a row from the chart table
// type ChartRow struct {
// 	ID          int    `db:"id"`
// 	Accno       string `db:"accno"`
// 	Description string `db:"description"`
// 	Link        string `db:"link"`
// 	Translation string `db:"translation"`
// }
//
// // NewAccountingDefaults creates a new AccountingDefaults instance
// func NewAccountingDefaults(db *sql.DB) *AccountingDefaults {
// 	return &AccountingDefaults{db: db}
// }
//
// // SaveDefaults saves default account settings to the database
// func (ad *AccountingDefaults) SaveDefaults(myConfig *MyConfig, form *Form) error {
// 	// Parse account fields by splitting on "--"
// 	accountFields := []string{"IC", "IC_income", "IC_expense", "FX_gain", "FX_loss"}
// 	for _, field := range accountFields {
// 		switch field {
// 		case "IC":
// 			if parts := strings.Split(form.IC, "--"); len(parts) > 0 {
// 				form.IC = parts[0]
// 			}
// 		case "IC_income":
// 			if parts := strings.Split(form.ICIncome, "--"); len(parts) > 0 {
// 				form.ICIncome = parts[0]
// 			}
// 		case "IC_expense":
// 			if parts := strings.Split(form.ICExpense, "--"); len(parts) > 0 {
// 				form.ICExpense = parts[0]
// 			}
// 		case "FX_gain":
// 			if parts := strings.Split(form.FXGain, "--"); len(parts) > 0 {
// 				form.FXGain = parts[0]
// 			}
// 		case "FX_loss":
// 			if parts := strings.Split(form.FXLoss, "--"); len(parts) > 0 {
// 				form.FXLoss = parts[0]
// 			}
// 		}
// 	}
//
// 	// Set account number fields
// 	form.InventoryAccno = form.IC
// 	form.IncomeAccno = form.ICIncome
// 	form.ExpenseAccno = form.ICExpense
// 	form.FXGainAccno = form.FXGain
// 	form.FXLossAccno = form.FXLoss
//
// 	// Begin transaction
// 	tx, err := ad.db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback()
//
// 	// Prepare statements
// 	insertStmt, err := tx.Prepare("INSERT INTO defaults (fldname, fldvalue) VALUES (?, ?)")
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare insert statement: %w", err)
// 	}
// 	defer insertStmt.Close()
//
// 	deleteStmt, err := tx.Prepare("DELETE FROM defaults WHERE fldname = ?")
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare delete statement: %w", err)
// 	}
// 	defer deleteStmt.Close()
//
// 	// Handle version (must be present)
// 	if _, err := deleteStmt.Exec("version"); err != nil {
// 		return fmt.Errorf("failed to delete version: %w", err)
// 	}
// 	if _, err := insertStmt.Exec("version", form.DBVersion); err != nil {
// 		return fmt.Errorf("failed to insert version: %w", err)
// 	}
//
// 	// Handle account mappings
// 	accountMappings := map[string]string{
// 		"inventory": form.InventoryAccno,
// 		"income":    form.IncomeAccno,
// 		"expense":   form.ExpenseAccno,
// 		"fxgain":    form.FXGainAccno,
// 		"fxloss":    form.FXLossAccno,
// 	}
//
// 	for accType, accno := range accountMappings {
// 		fieldName := accType + "_accno_id"
// 		if _, err := deleteStmt.Exec(fieldName); err != nil {
// 			return fmt.Errorf("failed to delete %s: %w", fieldName, err)
// 		}
//
// 		query := fmt.Sprintf(`INSERT INTO defaults (fldname, fldvalue)
// 			VALUES ('%s', (SELECT id FROM chart WHERE accno = '%s'))`, fieldName, accno)
//
// 		if _, err := tx.Exec(query); err != nil {
// 			return fmt.Errorf("failed to insert %s: %w", fieldName, err)
// 		}
// 	}
//
// 	// Handle required fields
// 	requiredFields := map[string]string{
// 		"transitionaccount": form.TransitionAccount,
// 		"selectedaccount":   form.SelectedAccount,
// 		"glnumber":         form.GLNumber,
// 		"sinumber":         form.SINumber,
// 		"vinumber":         form.VINumber,
// 		"batchnumber":      form.BatchNumber,
// 		"vouchernumber":    form.VoucherNumber,
// 		"sonumber":         form.SONumber,
// 		"ponumber":         form.PONumber,
// 		"sqnumber":         form.SQNumber,
// 		"rfqnumber":        form.RFQNumber,
// 		"partnumber":       form.PartNumber,
// 		"employeenumber":   form.EmployeeNumber,
// 		"customernumber":   form.CustomerNumber,
// 		"vendornumber":     form.VendorNumber,
// 		"projectnumber":    form.ProjectNumber,
// 		"precision":        form.Precision,
// 	}
//
// 	for fieldName, fieldValue := range requiredFields {
// 		if _, err := deleteStmt.Exec(fieldName); err != nil {
// 			return fmt.Errorf("failed to delete %s: %w", fieldName, err)
// 		}
// 		if _, err := insertStmt.Exec(fieldName, fieldValue); err != nil {
// 			return fmt.Errorf("failed to insert %s: %w", fieldName, err)
// 		}
// 	}
//
// 	// Handle optional fields
// 	if form.Optional != "" {
// 		optionalFields := strings.Split(form.Optional, " ")
// 		for _, fieldName := range optionalFields {
// 			fieldName = strings.TrimSpace(fieldName)
// 			if fieldName == "" {
// 				continue
// 			}
//
// 			if _, err := deleteStmt.Exec(fieldName); err != nil {
// 				return fmt.Errorf("failed to delete optional field %s: %w", fieldName, err)
// 			}
//
// 			if fieldValue, exists := form.Fields[fieldName]; exists && fieldValue != "" {
// 				if _, err := insertStmt.Exec(fieldName, fieldValue); err != nil {
// 					return fmt.Errorf("failed to insert optional field %s: %w", fieldName, err)
// 				}
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
// // DefaultAccounts retrieves default account information
// func (ad *AccountingDefaults) DefaultAccounts(myConfig *MyConfig, form *Form) error {
// 	// Initialize maps
// 	if form.Defaults == nil {
// 		form.Defaults = make(map[string]interface{})
// 	}
// 	if form.Accno == nil {
// 		form.Accno = make(map[string]map[string]AccountInfo)
// 	}
//
// 	// Get defaults from defaults table
// 	defaults, err := ad.getDefaults()
// 	if err != nil {
// 		return fmt.Errorf("failed to get defaults: %w", err)
// 	}
//
// 	// Copy defaults to form
// 	for key, value := range defaults {
// 		form.Fields[key] = value
// 	}
//
// 	// Set specific default mappings
// 	if inventoryID, ok := defaults["inventory_accno_id"]; ok {
// 		form.Defaults["IC"] = inventoryID
// 	}
// 	if incomeID, ok := defaults["income_accno_id"]; ok {
// 		form.Defaults["IC_income"] = incomeID
// 		form.Defaults["IC_sale"] = incomeID
// 	}
// 	if expenseID, ok := defaults["expense_accno_id"]; ok {
// 		form.Defaults["IC_expense"] = expenseID
// 		form.Defaults["IC_cogs"] = expenseID
// 	}
// 	if fxGainID, ok := defaults["fxgain_accno_id"]; ok {
// 		form.Defaults["FX_gain"] = fxGainID
// 	}
// 	if fxLossID, ok := defaults["fxloss_accno_id"]; ok {
// 		form.Defaults["FX_loss"] = fxLossID
// 	}
//
// 	// Get IC accounts
// 	if err := ad.getICAccounts(myConfig, form); err != nil {
// 		return fmt.Errorf("failed to get IC accounts: %w", err)
// 	}
//
// 	// Get FX accounts
// 	if err := ad.getFXAccounts(myConfig, form); err != nil {
// 		return fmt.Errorf("failed to get FX accounts: %w", err)
// 	}
//
// 	return nil
// }
//
// // getDefaults retrieves all defaults from the defaults table
// func (ad *AccountingDefaults) getDefaults() (map[string]string, error) {
// 	defaults := make(map[string]string)
//
// 	rows, err := ad.db.Query("SELECT fldname, fldvalue FROM defaults")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	for rows.Next() {
// 		var name, value string
// 		if err := rows.Scan(&name, &value); err != nil {
// 			return nil, err
// 		}
// 		defaults[name] = value
// 	}
//
// 	return defaults, rows.Err()
// }
//
// // getICAccounts retrieves IC-related accounts
// func (ad *AccountingDefaults) getICAccounts(myConfig *MyConfig, form *Form) error {
// 	query := `SELECT c.id, c.accno, c.description, c.link,
// 		COALESCE(l.description, c.description) AS translation
// 		FROM chart c
// 		LEFT JOIN translation l ON (l.trans_id = c.id AND l.language_code = ?)
// 		WHERE c.link LIKE '%IC%'
// 		ORDER BY c.accno`
//
// 	rows, err := ad.db.Query(query, myConfig.CountryCode)
// 	if err != nil {
// 		return err
// 	}
// 	defer rows.Close()
//
// 	for rows.Next() {
// 		var row ChartRow
// 		if err := rows.Scan(&row.ID, &row.Accno, &row.Description, &row.Link, &row.Translation); err != nil {
// 			return err
// 		}
//
// 		// Use translation if available, otherwise use description
// 		description := row.Description
// 		if row.Translation != "" {
// 			description = row.Translation
// 		}
//
// 		// Process link field
// 		linkParts := strings.Split(row.Link, ":")
// 		for _, key := range linkParts {
// 			if strings.Contains(key, "IC") {
// 				nkey := key
// 				if strings.Contains(key, "cogs") {
// 					nkey = "IC_expense"
// 				}
// 				if strings.Contains(key, "sale") {
// 					nkey = "IC_income"
// 				}
//
// 				// Initialize nested map if needed
// 				if form.Accno[nkey] == nil {
// 					form.Accno[nkey] = make(map[string]AccountInfo)
// 				}
//
// 				form.Accno[nkey][row.Accno] = AccountInfo{
// 					ID:          row.ID,
// 					Description: description,
// 				}
// 			}
// 		}
// 	}
//
// 	return rows.Err()
// }
//
// // getFXAccounts retrieves FX-related accounts
// func (ad *AccountingDefaults) getFXAccounts(myConfig *MyConfig, form *Form) error {
// 	query := `SELECT c.id, c.accno, c.description,
// 		COALESCE(l.description, c.description) AS translation
// 		FROM chart c
// 		LEFT JOIN translation l ON (l.trans_id = c.id AND l.language_code = ?)
// 		WHERE (c.category = 'I' OR c.category = 'E')
// 		AND c.charttype = 'A'
// 		ORDER BY c.accno`
//
// 	rows, err := ad.db.Query(query, myConfig.CountryCode)
// 	if err != nil {
// 		return err
// 	}
// 	defer rows.Close()
//
// 	// Initialize FX account maps
// 	if form.Accno["FX_gain"] == nil {
// 		form.Accno["FX_gain"] = make(map[string]AccountInfo)
// 	}
// 	if form.Accno["FX_loss"] == nil {
// 		form.Accno["FX_loss"] = make(map[string]AccountInfo)
// 	}
//
// 	for rows.Next() {
// 		var row ChartRow
// 		if err := rows.Scan(&row.ID, &row.Accno, &row.Description, &row.Translation); err != nil {
// 			return err
// 		}
//
// 		// Use translation if available, otherwise use description
// 		description := row.Description
// 		if row.Translation != "" {
// 			description = row.Translation
// 		}
//
// 		accountInfo := AccountInfo{
// 			ID:          row.ID,
// 			Description: description,
// 		}
//
// 		form.Accno["FX_gain"][row.Accno] = accountInfo
// 		form.Accno["FX_loss"][row.Accno] = accountInfo
// 	}
//
// 	return rows.Err()
// }
//
// // Example usage
// func Example() {
// 	// This would typically be initialized with your actual database connection
// 	// db, err := sql.Open("your-driver", "your-connection-string")
// 	// if err != nil {
// 	//     log.Fatal(err)
// 	// }
// 	// defer db.Close()
//
// 	// ad := NewAccountingDefaults(db)
//
// 	// Example form and config
// 	form := &Form{
// 		Fields:   make(map[string]string),
// 		Defaults: make(map[string]interface{}),
// 		Accno:    make(map[string]map[string]AccountInfo),
// 	}
//
// 	myConfig := &MyConfig{
// 		CountryCode: "en",
// 	}
//
// 	// Example usage:
// 	// err = ad.SaveDefaults(myConfig, form)
// 	// err = ad.DefaultAccounts(myConfig, form)
//
// 	log.Println("Accounting defaults implementation ready")
// }
