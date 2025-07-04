package accounting

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"regexp"
// 	"strings"
// )
//
// // Form represents the form data structure.
// type Form struct {
// 	// Fields from the Perl form hash, using appropriate types
// 	AR, ARAmount, ARTax, ARPaid, ARDiscount string
// 	AP, APAmount, APTax, APPaid, APDiscount string
// 	IC, ICIncome, ICSale, ICExpense, ICCogs  string
// 	ICTaxpart, ICTaxservice                  string
//
// 	Link         string
// 	ParentAccno  string
// 	ParentID     int64
// 	Accno        string
// 	GifiAccno    string
// 	Description  string
// 	Charttype    string
// 	Category     string
// 	Contra       int
// 	AllowGL      int
// 	ID           int64
// 	ICTaxpart    string
// 	ICTaxservice string
// 	ARTax        string
// 	APTax        string
//
// 	db *sql.DB
// }
//
// // DBConnectNoAuto connects to the database and disables autocommit (starts a transaction).
// func (f *Form) DBConnectNoAuto(dsn string) (*sql.Tx, error) {
// 	db, err := sql.Open("postgres", dsn) // assuming postgres, change as needed
// 	if err != nil {
// 		return nil, err
// 	}
// 	f.db = db
// 	// Begin a transaction (autoCommit off)
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return tx, nil
// }
//
// // Quote safely quotes a string for SQL. Using prepared statements is better, but for similarity, we keep this.
// func Quote(s string) string {
// 	// Replace single quote with two single quotes for SQL escaping
// 	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
// }
//
// // Dberror logs the error and query
// func (f *Form) Dberror(query string, err error) {
// 	log.Fatalf("Database error: %v\nQuery: %s", err, query)
// }
//
// // SaveAccount saves the account data into the database.
// func (f *Form) SaveAccount(dsn string) (bool, error) {
// 	// Connect to DB and start transaction
// 	tx, err := f.DBConnectNoAuto(dsn)
// 	if err != nil {
// 		return false, err
// 	}
// 	defer func() {
// 		if f.db != nil {
// 			f.db.Close()
// 		}
// 	}()
//
// 	// Reset link
// 	f.Link = ""
//
// 	// Append items to link string if present
// 	items := []string{
// 		f.AR,
// 		f.ARAmount,
// 		f.ARTax,
// 		f.ARPaid,
// 		f.ARDiscount,
// 		f.AP,
// 		f.APAmount,
// 		f.APTax,
// 		f.APPaid,
// 		f.APDiscount,
// 		f.IC,
// 		f.ICIncome,
// 		f.ICSale,
// 		f.ICExpense,
// 		f.ICCogs,
// 		f.ICTaxpart,
// 		f.ICTaxservice,
// 	}
//
// 	for _, item := range items {
// 		if item != "" {
// 			f.Link += item + ":"
// 		}
// 	}
// 	// Remove trailing colon if any
// 	f.Link = strings.TrimSuffix(f.Link, ":")
//
// 	// Strip blanks from parent_accno and split on --
// 	null := ""
// 	if strings.Contains(f.ParentAccno, "--") {
// 		parts := strings.SplitN(f.ParentAccno, "--", 2)
// 		f.ParentAccno = parts[0]
// 		if len(parts) > 1 {
// 			null = parts[1]
// 		}
// 	}
//
// 	// Remove spaces and single quotes from accno, gifi_accno, parent_accno
// 	for _, field := range []string{"Accno", "GifiAccno", "ParentAccno"} {
// 		switch field {
// 		case "Accno":
// 			f.Accno = removeSpacesAndSingleQuote(f.Accno)
// 		case "GifiAccno":
// 			f.GifiAccno = removeSpacesAndSingleQuote(f.GifiAccno)
// 		case "ParentAccno":
// 			f.ParentAccno = removeSpacesAndSingleQuote(f.ParentAccno)
// 		}
// 	}
//
// 	// Clean up accno, gifi_accno, description, parent_accno fields with regex replacements
// 	fieldsToClean := []*string{&f.Accno, &f.GifiAccno, &f.Description, &f.ParentAccno}
// 	for _, fieldPtr := range fieldsToClean {
// 		cleanField(fieldPtr)
// 	}
//
// 	// If parent_accno is set, get parent_id
// 	if f.ParentAccno != "" {
// 		err = tx.QueryRow("SELECT id FROM chart WHERE accno = $1", f.ParentAccno).Scan(&f.ParentID)
// 		if err != nil && err != sql.ErrNoRows {
// 			tx.Rollback()
// 			return false, err
// 		}
// 	}
//
// 	// Ensure parent_id, contra, allow_gl are integer values (already int types)
// 	// No action needed
//
// 	var query string
// 	var res sql.Result
//
// 	if f.ID != 0 {
// 		// Update existing record
// 		query = `
// 		UPDATE chart SET
// 			accno = $1,
// 			parent_id = $2,
// 			description = $3,
// 			charttype = $4,
// 			gifi_accno = $5,
// 			category = $6,
// 			link = $7,
// 			contra = $8,
// 			allow_gl = $9
// 		WHERE id = $10
// 		`
// 		res, err = tx.Exec(query, f.Accno, f.ParentID, f.Description, f.Charttype, f.GifiAccno, f.Category, f.Link, f.Contra, f.AllowGL, f.ID)
// 		if err != nil {
// 			tx.Rollback()
// 			f.Dberror(query, err)
// 			return false, err
// 		}
// 	} else {
// 		// Insert new record
// 		query = `
// 		INSERT INTO chart
// 			(accno, parent_id, description, charttype, gifi_accno, category, link, contra, allow_gl)
// 		VALUES
// 			($1, $2, $3, $4, $5, $6, $7, $8, $9)
// 		RETURNING id
// 		`
// 		err = tx.QueryRow(query, f.Accno, f.ParentID, f.Description, f.Charttype, f.GifiAccno, f.Category, f.Link, f.Contra, f.AllowGL).Scan(&f.ID)
// 		if err != nil {
// 			tx.Rollback()
// 			f.Dberror(query, err)
// 			return false, err
// 		}
// 	}
//
// 	chartID := f.ID
//
// 	if f.ID == 0 {
// 		// Get id from chart for accno
// 		err = tx.QueryRow("SELECT id FROM chart WHERE accno = $1", f.Accno).Scan(&chartID)
// 		if err != nil {
// 			tx.Rollback()
// 			return false, err
// 		}
// 	} else {
// 		// Check if tax accounts changed and update acc_trans.tax field accordingly
// 		if f.ICTaxpart != "" || f.ICTaxservice != "" || f.ARTax != "" || f.APTax != "" {
// 			var oldAccno, oldDescription string
// 			err = tx.QueryRow("SELECT accno, description FROM chart WHERE id = $1", f.ID).Scan(&oldAccno, &oldDescription)
// 			if err != nil {
// 				tx.Rollback()
// 				return false, err
// 			}
// 			// Compare and update if changed
// 			if f.Accno != oldAccno || f.Description != oldDescription {
// 				updateQuery := `
// 				UPDATE acc_trans
// 				SET tax = (SELECT accno || '--' || description FROM chart WHERE id = $1)
// 				WHERE tax_chart_id = $1
// 				`
// 				_, err = tx.Exec(updateQuery, f.ID)
// 				if err != nil {
// 					tx.Rollback()
// 					return false, err
// 				}
// 			}
// 		}
// 	}
//
// 	// Handle tax table insertion or deletion
// 	if f.ICTaxpart != "" || f.ICTaxservice != "" || f.ARTax != "" || f.APTax != "" {
// 		// Check if tax exists for chart_id
// 		var taxID int64
// 		err = tx.QueryRow("SELECT chart_id FROM tax WHERE chart_id = $1", chartID).Scan(&taxID)
// 		if err != nil && err != sql.ErrNoRows {
// 			tx.Rollback()
// 			return false, err
// 		}
//
// 		if taxID == 0 {
// 			// Insert tax record with rate 0
// 			insertTaxQuery := "INSERT INTO tax (chart_id, rate) VALUES ($1, 0)"
// 			_, err = tx.Exec(insertTaxQuery, chartID)
// 			if err != nil {
// 				tx.Rollback()
// 				f.Dberror(insertTaxQuery, err)
// 				return false, err
// 			}
// 		}
//
// 	} else {
// 		// Remove tax for this chart id if id exists
// 		if f.ID != 0 {
// 			deleteQuery := "DELETE FROM tax WHERE chart_id = $1"
// 			_, err = tx.Exec(deleteQuery, f.ID)
// 			if err != nil {
// 				tx.Rollback()
// 				f.Dberror(deleteQuery, err)
// 				return false, err
// 			}
// 		}
// 	}
//
// 	err = tx.Commit()
// 	if err != nil {
// 		return false, err
// 	}
//
// 	return true, nil
// }
//
// // removeSpacesAndSingleQuote removes spaces and single quotes from a string
// func removeSpacesAndSingleQuote(s string) string {
// 	return strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "'", "")
// }
//
// // cleanField applies regex replacements to clean a string field similar to Perl code
// func cleanField(field *string) {
// 	// Replace -(-+) with -
// 	reDash := regexp.MustCompile(`-(-+)`)
// 	*field = reDash.ReplaceAllString(*field, "-")
//
// 	// Replace multiple spaces with single space
// 	reSpace := regexp.MustCompile(` +`)
// 	*field = reSpace.ReplaceAllString(*field, " ")
//
// 	// Trim leading and trailing whitespace
// 	*field = strings.TrimSpace(*field)
// }
//
// func example() {
// 	// Example usage
// 	dsn := "user=youruser dbname=yourdb sslmode=disable" // change to your database DSN
// 	form := Form{
// 		// Assign example data to form fields
// 		Accno:       "1000",
// 		ParentAccno: "5000--something",
// 		Description: "Example Account",
// 		Charttype:   "A",
// 		GifiAccno:   "1234",
// 		Category:    "Category1",
// 		Contra:      0,
// 		AllowGL:     1,
// 		// Set other fields as needed
// 	}
//
// 	ok, err := form.SaveAccount(dsn)
// 	if err != nil {
// 		log.Fatalf("Failed to save account: %v", err)
// 	}
// 	if ok {
// 		fmt.Println("Account saved successfully.")
// 	}
// }
//
