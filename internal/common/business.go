package common

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"regexp"
// 	"strings"
//
// 	_ "github.com/lib/pq" // or other DB driver as needed
// )
//
// // Config holds database config info
// type Config struct {
// 	DSN string // Data Source Name for sql.Open
// }
//
// // Form holds form data and other info
// type Form struct {
// 	ID          int
// 	Description string
// 	Discount    float64
// 	Direction   string // "ASC" or "DESC"
// 	ALL         []Business
// 	db          *sql.DB
// }
//
// // Business holds a business record
// type Business struct {
// 	ID          int
// 	Description string
// 	Discount    float64
// }
//
// // dbconnect connects to the database and stores the handle in the form
// func (f *Form) dbconnect(cfg Config) error {
// 	db, err := sql.Open("postgres", cfg.DSN)
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to database: %w", err)
// 	}
// 	// Test the connection
// 	err = db.Ping()
// 	if err != nil {
// 		return fmt.Errorf("failed to ping database: %w", err)
// 	}
// 	f.db = db
// 	return nil
// }
//
// // dberror prints the error and query and exits (simulate form->dberror)
// func (f *Form) dberror(query string, err error) {
// 	log.Fatalf("Database error: %v\nQuery: %s", err, query)
// }
//
// // sortOrder sets the default sort order if not set
// func (f *Form) sortOrder() {
// 	// Default to ASC if direction is not set or invalid
// 	dir := strings.ToUpper(f.Direction)
// 	if dir != "ASC" && dir != "DESC" {
// 		f.Direction = "ASC"
// 	} else {
// 		f.Direction = dir
// 	}
// }
//
// // business fetches all businesses ordered by description
// func (f *Form) business(cfg Config) {
// 	if err := f.dbconnect(cfg); err != nil {
// 		log.Fatalf("Error in dbconnect: %v", err)
// 	}
// 	defer f.db.Close()
//
// 	f.sortOrder()
// 	query := fmt.Sprintf(`SELECT id, description, discount FROM business ORDER BY description %s`, f.Direction)
//
// 	rows, err := f.db.Query(query)
// 	if err != nil {
// 		f.dberror(query, err)
// 	}
// 	defer rows.Close()
//
// 	var all []Business
// 	for rows.Next() {
// 		var b Business
// 		err = rows.Scan(&b.ID, &b.Description, &b.Discount)
// 		if err != nil {
// 			f.dberror(query, err)
// 		}
// 		all = append(all, b)
// 	}
// 	if err = rows.Err(); err != nil {
// 		f.dberror(query, err)
// 	}
//
// 	f.ALL = all
// }
//
// // getBusiness fetches a single business by ID and populates the form fields
// func (f *Form) getBusiness(cfg Config) {
// 	if err := f.dbconnect(cfg); err != nil {
// 		log.Fatalf("Error in dbconnect: %v", err)
// 	}
// 	defer f.db.Close()
//
// 	// Ensure ID is valid (>0)
// 	if f.ID <= 0 {
// 		log.Fatalf("Invalid business ID: %d", f.ID)
// 	}
//
// 	query := `SELECT description, discount FROM business WHERE id = $1`
// 	row := f.db.QueryRow(query, f.ID)
//
// 	err := row.Scan(&f.Description, &f.Discount)
// 	if err == sql.ErrNoRows {
// 		log.Fatalf("No business found with id %d", f.ID)
// 	} else if err != nil {
// 		f.dberror(query, err)
// 	}
// }
//
// // saveBusiness inserts or updates a business record based on form.ID
// func (f *Form) saveBusiness(cfg Config) {
// 	if err := f.dbconnect(cfg); err != nil {
// 		log.Fatalf("Error in dbconnect: %v", err)
// 	}
// 	defer f.db.Close()
//
// 	// Clean up description
// 	// s/-(-)+/-/g in Perl means replace multiple consecutive '-' with single '-'
// 	reDash := regexp.MustCompile(`-+`)
// 	f.Description = reDash.ReplaceAllString(f.Description, "-")
//
// 	// s/ ( )+/ /g means replace multiple spaces with single space
// 	reSpace := regexp.MustCompile(` +`)
// 	f.Description = reSpace.ReplaceAllString(f.Description, " ")
//
// 	// Convert discount percentage to decimal
// 	f.Discount = f.Discount / 100.0
//
// 	var query string
// 	var err error
// 	if f.ID > 0 {
// 		// UPDATE
// 		query = `UPDATE business SET description = $1, discount = $2 WHERE id = $3`
// 		_, err = f.db.Exec(query, f.Description, f.Discount, f.ID)
// 	} else {
// 		// INSERT
// 		query = `INSERT INTO business (description, discount) VALUES ($1, $2)`
// 		_, err = f.db.Exec(query, f.Description, f.Discount)
// 	}
//
// 	if err != nil {
// 		f.dberror(query, err)
// 	}
// }
//
// // deleteBusiness deletes a business record by ID
// func (f *Form) deleteBusiness(cfg Config) {
// 	if err := f.dbconnect(cfg); err != nil {
// 		log.Fatalf("Error in dbconnect: %v", err)
// 	}
// 	defer f.db.Close()
//
// 	if f.ID <= 0 {
// 		log.Fatalf("Invalid business ID for delete: %d", f.ID)
// 	}
//
// 	query := `DELETE FROM business WHERE id = $1`
// 	_, err := f.db.Exec(query, f.ID)
// 	if err != nil {
// 		f.dberror(query, err)
// 	}
// }
//
//
