package comon

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// Account represents a chart of accounts entry
type Account struct {
	ID             int
	Accno          string
	Description    string
	ChartType      string
	GifiAccno      string
	Category       string
	Link           string
	AllowGL        bool
	ParentAccno    sql.NullString
	Translation    sql.NullString
	Amount         sql.NullFloat64
	GifiDescription string
	Debit          float64
	Credit         float64
}

// Transaction represents a financial transaction
type Transaction struct {
	ID          int
	Reference   string
	Description string
	Name        sql.NullString
	TransDate   time.Time
	Invoice     bool
	Curr        string
	Amount      float64
	Module      string
	Cleared     bool
	Source      sql.NullString
	Till        sql.NullString
	ChartID     int
	VcID        int
	Debit       float64
	Credit      float64
	Accno       []string
}

// Form holds application state and data
type Form struct {
	CA           interface{} // Slice of Account or Transaction
	Precision    int
	Company      string
	FromDate     string
	ToDate       string
	Year         int
	Month        int
	Interval     string
	Method       string
	Accno        string
	GifiAccno    string
	AccountType  string
	Balance      float64
	Department   string
	ProjectNumber string
	CountryCode  string
}

// DBConfig holds database configuration
type DBConfig struct {
	Driver string
}

// AllAccounts retrieves all accounts with balances and GIFI descriptions
func AllAccounts(db *sql.DB, form *Form, config *DBConfig) error {
	// Get defaults (precision, company)
	if err := form.getDefaults(db); err != nil {
		return err
	}

	// Step 1: Get account balances
	balanceMap := make(map[string]float64)
	query := `
		SELECT c.accno, SUM(ac.amount) AS amount
		FROM chart c
		JOIN acc_trans ac ON ac.chart_id = c.id
		WHERE ac.approved = '1'
		GROUP BY c.accno
	`
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("account balance query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var accno string
		var amount sql.NullFloat64
		if err := rows.Scan(&accno, &amount); err != nil {
			return err
		}
		if amount.Valid {
			balanceMap[accno] = amount.Float64
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Step 2: Get GIFI descriptions
	gifiMap := make(map[string]string)
	query = "SELECT accno, description FROM gifi"
	rows, err = db.Query(query)
	if err != nil {
		return fmt.Errorf("GIFI query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var accno, desc string
		if err := rows.Scan(&accno, &desc); err != nil {
			return err
		}
		gifiMap[accno] = desc
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Step 3: Get full account details
	query = `
		SELECT c.id, c.accno, c.description, c.charttype, c.gifi_accno,
			c.category, c.link, c.allow_gl, c2.accno AS parent_accno,
			l.description AS translation
		FROM chart c
		LEFT JOIN chart c2 ON c2.id = c.parent_id
		LEFT JOIN translation l ON l.trans_id = c.id AND l.language_code = $1
		ORDER BY c.accno
	`
	rows, err = db.Query(query, form.CountryCode)
	if err != nil {
		return fmt.Errorf("account details query failed: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var a Account
		var parentAccno, translation sql.NullString
		err := rows.Scan(
			&a.ID, &a.Accno, &a.Description, &a.ChartType, &a.GifiAccno,
			&a.Category, &a.Link, &a.AllowGL, &parentAccno, &translation,
		)
		if err != nil {
			return err
		}

		// Process retrieved data
		a.ParentAccno = parentAccno
		a.Translation = translation
		if amount, exists := balanceMap[a.Accno]; exists {
			a.Amount = sql.NullFloat64{Float64: amount, Valid: true}
		}
		if desc, exists := gifiMap[a.GifiAccno]; exists {
			a.GifiDescription = desc
		}
		if a.Amount.Valid {
			if a.Amount.Float64 < 0 {
				a.Debit = -a.Amount.Float64
			} else {
				a.Credit = a.Amount.Float64
			}
		}
		if translation.Valid {
			a.Description = translation.String
		}

		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	form.CA = accounts
	return nil
}

// AllTransactions retrieves transactions for specified accounts
func AllTransactions(db *sql.DB, form *Form, config *DBConfig) error {
	// Get defaults (precision, company)
	if err := form.getDefaults(db); err != nil {
		return err
	}

	// Calculate date range if needed
	if form.Year != 0 && form.Month != 0 {
		form.FromDate, form.ToDate = form.calculateDateRange()
	}

	// Step 1: Get chart IDs
	var chartIDs []int
	query := "SELECT id FROM chart WHERE accno = $1"
	if form.AccountType == "gifi" {
		query = "SELECT id FROM chart WHERE gifi_accno = $1 AND charttype = 'A'"
		param := form.GifiAccno
		rows, err := db.Query(query, param)
		if err != nil {
			return fmt.Errorf("chart ID query failed: %w", err)
		}
		defer rows.Close()
		
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				return err
			}
			chartIDs = append(chartIDs, id)
		}
		if err := rows.Err(); err != nil {
			return err
		}
	} else {
		param := form.Accno
		rows, err := db.Query(query, param)
		if err != nil {
			return fmt.Errorf("chart ID query failed: %w", err)
		}
		defer rows.Close()
		
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				return err
			}
			chartIDs = append(chartIDs, id)
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}

	if len(chartIDs) == 0 {
		return errors.New("no matching chart accounts found")
	}

	// Step 2: Build query with dynamic parameters
	query, args := buildTransactionQuery(form, chartIDs, config.Driver)
	rows, err := db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("transaction query failed: %w", err)
	}
	defer rows.Close()

	// Step 3: Process transactions
	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var name, source sql.NullString
		var till sql.NullString
		err := rows.Scan(
			&t.ID, &t.Reference, &t.Description, &name, &t.TransDate,
			&t.Invoice, &t.Curr, &t.Amount, &t.Module, &t.Cleared,
			&source, &till, &t.ChartID, &t.VcID,
		)
		if err != nil {
			return err
		}

		t.Name = name
		t.Source = source
		t.Till = till

		// Additional processing
		processTransactionModule(&t)
		processTransactionAmounts(db, &t, form) // Handles contra accounts

		transactions = append(transactions, t)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	form.CA = transactions
	return nil
}

// Helper functions would be implemented below:
// - form.getDefaults()
// - form.calculateDateRange()
// - buildTransactionQuery()
// - processTransactionModule()
// - processTransactionAmounts()

func main() {
	// Example usage
	db, err := sql.Open("postgres", "connection_string")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	form := &Form{
		Accno:       "1200",
		CountryCode: "US",
	}
	config := &DBConfig{Driver: "pg"}

	if err := AllAccounts(db, form, config); err != nil {
		log.Fatal(err)
	}

	// Use form.CA which now contains []Account
}
