package comman

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/lib/pq" // Assuming PostgreSQL; adjust import as needed
)

type Form struct {
	ID                         int
	Accno                      string
	Description                string
	Charttype                  string
	GifiAccno                  string
	Category                   string
	Link                       string
	Contra                     int
	AllowGL                    int
	ParentID                   int
	ParentAccno                string
	ParentAccnoDescription     string
	Orphaned                   bool
	AR                         string
	ARAmount                   string
	ARTax                      string
	ARPaid                     string
	ARDiscount                 string
	AP                         string
	APAmount                   string
	APTax                      string
	APPaid                     string
	APDiscount                 string
	IC                         string
	ICIncome                   string
	ICSale                     string
	ICExpense                  string
	ICCogs                     string
	ICTaxpart                  string
	ICTaxservice               string
	CategoryOrig               string // To hold original category if needed
	ParentAccnoOrig            string // To hold original parent accno if needed
	ParentAccnoDescriptionOrig string // To hold original parent accno description if needed
	// Add other fields as needed
}

type MyConfig struct {
	// Add fields as needed for DB config
	DSN string
}

func (f *Form) dbconnect(myconfig *MyConfig) (*sql.DB, error) {
	// Connect to DB with AutoCommit on (default)
	db, err := sql.Open("postgres", myconfig.DSN)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (f *Form) dbconnectNoAuto(myconfig *MyConfig) (*sql.DB, error) {
	// Connect to DB with AutoCommit off (transaction)
	db, err := sql.Open("postgres", myconfig.DSN)
	if err != nil {
		return nil, err
	}
	// We will begin a transaction later in save_account
	return db, nil
}

func (f *Form) dberror(query string, err error) {
	log.Fatalf("Database error on query [%s]: %v", query, err)
}

func (f *Form) get_defaults(db *sql.DB, keys []string) (map[string]string, error) {
	// This function is assumed to fetch default values for keys like '%accno_id'
	// Since original code calls: $form->get_defaults($dbh, \@{['%accno_id']});
	// We'll simulate fetching defaults from some defaults table or config

	// For demonstration, return empty map
	// Implement as needed
	return map[string]string{}, nil
}

func (f *Form) GetAccount(myconfig *MyConfig) error {
	db, err := f.dbconnect(myconfig)
	if err != nil {
		return err
	}
	defer db.Close()

	// Ensure ID is numeric (already int)
	// Prepare query
	query := `
		SELECT accno, description, charttype, gifi_accno,
		       category, link, contra, allow_gl, parent_id
		FROM chart
		WHERE id = $1
	`
	row := db.QueryRow(query, f.ID)

	var contra, allowGL, parentID sql.NullInt64
	var accno, description, charttype, gifiAccno, category, link sql.NullString

	err = row.Scan(&accno, &description, &charttype, &gifiAccno,
		&category, &link, &contra, &allowGL, &parentID)
	if err != nil {
		return fmt.Errorf("error fetching account: %w", err)
	}

	// Assign values to form, checking for NULLs
	if accno.Valid {
		f.Accno = accno.String
	}
	if description.Valid {
		f.Description = description.String
	}
	if charttype.Valid {
		f.Charttype = charttype.String
	}
	if gifiAccno.Valid {
		f.GifiAccno = gifiAccno.String
	}
	if category.Valid {
		f.Category = category.String
	}
	if link.Valid {
		f.Link = link.String
	}
	if contra.Valid {
		f.Contra = int(contra.Int64)
	}
	if allowGL.Valid {
		f.AllowGL = int(allowGL.Int64)
	}
	if parentID.Valid {
		f.ParentID = int(parentID.Int64)
	} else {
		f.ParentID = 0
	}

	// If parent_id exists, get parent_accno and description
	if f.ParentID != 0 {
		parentQuery := `SELECT accno, description FROM chart WHERE id = $1`
		row := db.QueryRow(parentQuery, f.ParentID)
		var pAccno, pDescription sql.NullString
		err = row.Scan(&pAccno, &pDescription)
		if err != nil {
			return fmt.Errorf("error fetching parent account: %w", err)
		}
		if pAccno.Valid && pDescription.Valid {
			f.ParentAccno = fmt.Sprintf("%s--%s", pAccno.String, pDescription.String)
		}
	}

	// Get default accounts
	defaults, err := f.get_defaults(db, []string{"%accno_id"})
	if err != nil {
		return fmt.Errorf("error getting defaults: %w", err)
	}
	for k, v := range defaults {
		// Assuming keys correspond to struct fields, set accordingly
		switch k {
		case "accno_id":
			// No direct mapping in form, skip or implement if needed
		default:
			// No action
		}
	}

	// Check if we have any transactions
	transQuery := `SELECT trans_id FROM acc_trans WHERE chart_id = $1 LIMIT 1`
	row = db.QueryRow(transQuery, f.ID)
	var transID sql.NullInt64
	err = row.Scan(&transID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking transactions: %w", err)
	}
	f.Orphaned = !transID.Valid

	return nil
}

func (f *Form) SaveAccount(myconfig *MyConfig) error {
	db, err := f.dbconnectNoAuto(myconfig)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Build link string
	var linkParts []string
	items := []string{
		f.AR, f.ARAmount, f.ARTax, f.ARPaid, f.ARDiscount,
		f.AP, f.APAmount, f.APTax, f.APPaid, f.APDiscount,
		f.IC, f.ICIncome, f.ICSale, f.ICExpense, f.ICCogs, f.ICTaxpart, f.ICTaxservice,
	}
	for _, item := range items {
		if item != "" {
			linkParts = append(linkParts, item)
		}
	}
	f.Link = strings.Join(linkParts, ":")

	// Strip blanks from parent_accno (split by "--")
	parts := strings.SplitN(f.ParentAccno, "--", 2)
	if len(parts) > 0 {
		f.ParentAccno = parts[0]
	}

	// Remove spaces and single quotes from accno, gifi_accno, parent_accno
	reSpaceQuote := regexp.MustCompile(`[ ']+`)
	f.Accno = reSpaceQuote.ReplaceAllString(f.Accno, "")
	f.GifiAccno = reSpaceQuote.ReplaceAllString(f.GifiAccno, "")
	f.ParentAccno = reSpaceQuote.ReplaceAllString(f.ParentAccno, "")

	// Clean up accno, gifi_accno, description, parent_accno: replace multiple dashes with single dash,
	// multiple spaces with single space, trim leading/trailing spaces
	reMultiDash := regexp.MustCompile(`-+`)
	reMultiSpace := regexp.MustCompile(`\s+`)

	cleanField := func(s string) string {
		s = reMultiDash.ReplaceAllString(s, "-")
		s = reMultiSpace.ReplaceAllString(s, " ")
		s = strings.TrimSpace(s)
		return s
	}

	f.Accno = cleanField(f.Accno)
	f.GifiAccno = cleanField(f.GifiAccno)
	f.Description = cleanField(f.Description)
	f.ParentAccno = cleanField(f.ParentAccno)

	// Get parent_id if parent_accno exists
	if f.ParentAccno != "" {
		query := `SELECT id FROM chart WHERE accno = $1`
		row := tx.QueryRow(query, f.ParentAccno)
		var parentID sql.NullInt64
		err = row.Scan(&parentID)
		if err != nil && err != sql.ErrNoRows {
			tx.Rollback()
			return fmt.Errorf("error fetching parent_id: %w", err)
		}
		if parentID.Valid {
			f.ParentID = int(parentID.Int64)
		} else {
			f.ParentID = 0
		}
	} else {
		f.ParentID = 0
	}

	// Ensure contra and allow_gl are ints (already int in struct)

	// Insert or update chart record
	if f.ID != 0 {
		// Update
		query := `
			UPDATE chart SET
				accno = $1,
				parent_id = $2,
				description = $3,
				charttype = $4,
				gifi_accno = $5,
				category = $6,
				link = $7,
				contra = $8,
				allow_gl = $9
			WHERE id = $10
		`
		_, err = tx.Exec(query,
			f.Accno,
			f.ParentID,
			f.Description,
			f.Charttype,
			f.GifiAccno,
			f.Category,
			f.Link,
			f.Contra,
			f.AllowGL,
			f.ID,
		)
		if err != nil {
			tx.Rollback()
			f.dberror(query, err)
			return err
		}
	} else {
		// Insert
		query := `
			INSERT INTO chart
				(accno, parent_id, description, charttype, gifi_accno, category, link, contra, allow_gl)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`
		err = tx.QueryRow(query,
			f.Accno,
			f.ParentID,
			f.Description,
			f.Charttype,
			f.GifiAccno,
			f.Category,
			f.Link,
			f.Contra,
			f.AllowGL,
		).Scan(&f.ID)
		if err != nil {
			tx.Rollback()
			f.dberror(query, err)
			return err
		}
	}

	chartID := f.ID

	// If updating and tax-related fields exist, update acc_trans.tax if accno or description changed
	if f.ID != 0 && (f.ICTaxpart != "" || f.ICTaxservice != "" || f.ARTax != "" || f.APTax != "") {
		var oldAccno, oldDescription string
		row := tx.QueryRow("SELECT accno, description FROM chart WHERE id = $1", f.ID)
		err = row.Scan(&oldAccno, &oldDescription)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error fetching old accno/description: %w", err)
		}
		if f.Accno != oldAccno || f.Description != oldDescription {
			updateQuery := `
				UPDATE acc_trans
				SET tax = (SELECT accno || '--' || description FROM chart WHERE id = $1)
				WHERE tax_chart_id = $1
			`
			_, err = tx.Exec(updateQuery, f.ID)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error updating acc_trans tax: %w", err)
			}
		}
	}

	// Handle tax table insert/delete
	if f.ICTaxpart != "" || f.ICTaxservice != "" || f.ARTax != "" || f.APTax != "" {
		// Check if tax record exists
		var taxID int
		err = tx.QueryRow("SELECT chart_id FROM tax WHERE chart_id = $1", chartID).Scan(&taxID)
		if err != nil && err != sql.ErrNoRows {
			tx.Rollback()
			return fmt.Errorf("error checking tax record: %w", err)
		}
		if err == sql.ErrNoRows {
			// Insert tax record with rate 0
			insertTax := `INSERT INTO tax (chart_id, rate) VALUES ($1, 0)`
			_, err = tx.Exec(insertTax, chartID)
			if err != nil {
				tx.Rollback()
				f.dberror(insertTax, err)
				return err
			}
		}
	} else {
		// Remove tax record if exists and form.ID != 0
		if f.ID != 0 {
			deleteTax := `DELETE FROM tax WHERE chart_id = $1`
			_, err = tx.Exec(deleteTax, f.ID)
			if err != nil {
				tx.Rollback()
				f.dberror(deleteTax, err)
				return err
			}
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func main() {
	// Example usage:

	myconfig := &MyConfig{
		DSN: "user=youruser dbname=yourdb sslmode=disable", // Adjust DSN accordingly
	}

	form := &Form{
		ID: 1, // Example ID
	}

	err := form.GetAccount(myconfig)
	if err != nil {
		log.Fatalf("GetAccount error: %v", err)
	}

	// Modify form fields as needed, then save
	form.Description = "Updated Description"
	err = form.SaveAccount(myconfig)
	if err != nil {
		log.Fatalf("SaveAccount error: %v", err)
	}

	fmt.Println("Account saved successfully")
}

// Assume Form struct encapsulates form data and database connection methods
type Form struct {
	ID int64
	DB *sql.DB
}

// Simulate method to connect to database with no auto-commit (transactions)
func (f *Form) dbconnectNoAuto() (*sql.DB, *sql.Tx, error) {
	// In real code, you would open the DB connection here if not already open
	// For demonstration, assume f.DB is already connected
	if f.DB == nil {
		return nil, nil, fmt.Errorf("database not connected")
	}
	tx, err := f.DB.Begin()
	if err != nil {
		return nil, nil, err
	}
	return f.DB, tx, nil
}

// Simulate method to get default account IDs from the database
func (f *Form) getDefaults(tx *sql.Tx, keys []string) (map[string]int64, error) {
	// This method queries the database for defaults for keys like "%_accno_id"
	// Here we simulate this by querying a fictional defaults table or similar
	// For simplicity, assume defaults are fetched from a table "defaults" with columns key and value

	defaults := make(map[string]int64)

	for _, key := range keys {
		var val int64
		err := tx.QueryRow("SELECT value FROM defaults WHERE key = ?", key).Scan(&val)
		if err != nil {
			if err == sql.ErrNoRows {
				// Default to 0 if not found
				val = 0
			} else {
				return nil, err
			}
		}
		defaults[key] = val
	}

	return defaults, nil
}

// Simulate error handling method for DB errors
func (f *Form) dberror(query string, err error) {
	log.Fatalf("Database error on query %q: %v", query, err)
}

// deleteAccount deletes the account with all related records, similar to the original Perl code
func (f *Form) deleteAccount() error {
	// Connect to DB with transaction
	db, tx, err := f.dbconnectNoAuto()
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %v", err)
	}

	// Ensure tx rollback on error or panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Get defaults for keys "%_accno_id"
	defaults, err := f.getDefaults(tx, []string{"inventory_accno_id", "income_accno_id", "expense_accno_id"})
	if err != nil {
		return fmt.Errorf("failed to get defaults: %v", err)
	}

	id := f.ID

	// For each key, check if any parts use the default account, then update parts referencing this id
	for _, key := range []string{"inventory_accno_id", "income_accno_id", "expense_accno_id"} {
		defaultVal := defaults[key]

		// Count parts where the key column equals defaultVal
		countQuery := fmt.Sprintf("SELECT count(*) FROM parts WHERE %s = ?", key)
		var count int
		err = tx.QueryRow(countQuery, defaultVal).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to count parts for %s: %v", key, err)
		}

		if count > 0 {
			if defaultVal != 0 {
				// Update parts where key column equals current id to defaultVal
				updateQuery := fmt.Sprintf("UPDATE parts SET %s = ? WHERE %s = ?", key, key)
				_, err = tx.Exec(updateQuery, defaultVal, id)
				if err != nil {
					f.dberror(updateQuery, err)
					return err
				}
			} else {
				// If defaultVal is zero, rollback and return early
				_ = tx.Rollback()
				return nil
			}
		}
	}

	// Delete from chart table
	_, err = tx.Exec("DELETE FROM chart WHERE id = ?", id)
	if err != nil {
		f.dberror("DELETE FROM chart WHERE id = ?", err)
		return err
	}

	// Delete from bank table
	_, err = tx.Exec("DELETE FROM bank WHERE id = ?", id)
	if err != nil {
		f.dberror("DELETE FROM bank WHERE id = ?", err)
		return err
	}

	// Delete from address table (trans_id = id)
	_, err = tx.Exec("DELETE FROM address WHERE trans_id = ?", id)
	if err != nil {
		f.dberror("DELETE FROM address WHERE trans_id = ?", err)
		return err
	}

	// Delete from translation table (trans_id = id)
	_, err = tx.Exec("DELETE FROM translation WHERE trans_id = ?", id)
	if err != nil {
		f.dberror("DELETE FROM translation WHERE trans_id = ?", err)
		return err
	}

	// Delete from partstax, customertax, vendortax, tax where chart_id = id
	tables := []string{"partstax", "customertax", "vendortax", "tax"}
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE chart_id = ?", table)
		_, err = tx.Exec(query, id)
		if err != nil {
			f.dberror(query, err)
			return err
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Disconnect DB in Perl code; here we do not close DB as it's typically reused
	_ = db // to avoid unused variable warning if DB close not needed here

	return nil
}

func main() {
	// Example usage:
	// Open DB connection, initialize Form struct, call deleteAccount

	// For demonstration, open a DB connection (adjust to your DSN)
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	form := Form{
		ID: 123, // example id to delete
		DB: db,
	}

	err = form.deleteAccount()
	if err != nil {
		log.Printf("Error deleting account: %v", err)
	} else {
		log.Printf("Account deleted successfully")
	}
}
