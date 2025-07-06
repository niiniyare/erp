package common

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"strings"
// 	"time"
//
// 	_ "github.com/lib/pq" // Assuming PostgreSQL, adjust import if needed
// )
//
// // Config holds database configuration details
// type Config struct {
// 	DSN string // Data Source Name for database connection
// }
//
// // Form holds form data and methods to interact with the DB
// type Form struct {
// 	// Input/output fields
// 	ID          int64
// 	Description string
// 	Address1    string
// 	Address2    string
// 	City        string
// 	State       string
// 	Zipcode     string
// 	Country     string
// 	Direction   string // "ASC" or "DESC" for sorting
// 	All         []WarehouseRecord
// 	Orphaned    bool
//
// 	db *sql.DB
// }
//
// // WarehouseRecord holds a warehouse with address info
// type WarehouseRecord struct {
// 	ID          int64
// 	Description string
// 	Address1    string
// 	Address2    string
// 	City        string
// 	State       string
// 	Zipcode     string
// 	Country     string
// }
//
// // dbConnect connects to the database using the given config
// func dbConnect(cfg Config) (*sql.DB, error) {
// 	db, err := sql.Open("postgres", cfg.DSN)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Verify connection
// 	if err = db.Ping(); err != nil {
// 		db.Close()
// 		return nil, err
// 	}
// 	return db, nil
// }
//
// // dbConnectNoAuto connects to DB and disables auto-commit by beginning a transaction
// func dbConnectNoAuto(cfg Config) (*sql.DB, *sql.Tx, error) {
// 	db, err := dbConnect(cfg)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	// We will manage transactions manually
// 	tx, err := db.Begin()
// 	if err != nil {
// 		db.Close()
// 		return nil, nil, err
// 	}
// 	return db, tx, nil
// }
//
// // dberror handles database errors (for now, just returns the error with query info)
// func dberror(query string, err error) error {
// 	return fmt.Errorf("database error executing query %q: %w", query, err)
// }
//
// // sanitizeDirection validates the sorting direction to prevent SQL injection
// func sanitizeDirection(dir string) string {
// 	dir = strings.ToUpper(dir)
// 	if dir != "ASC" && dir != "DESC" {
// 		return "ASC"
// 	}
// 	return dir
// }
//
// // Warehouses fetches all warehouses with address info, sorted by description and direction
// func (f *Form) Warehouses(cfg Config) error {
// 	db, err := dbConnect(cfg)
// 	if err != nil {
// 		return err
// 	}
// 	defer db.Close()
//
// 	// Validate direction
// 	dir := sanitizeDirection(f.Direction)
//
// 	query := fmt.Sprintf(`SELECT w.id, w.description,
//                  a.address1, a.address2, a.city, a.state, a.zipcode, a.country
//                  FROM warehouse w
//                  JOIN address a ON a.trans_id = w.id
//                  ORDER BY w.description %s`, dir)
//
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return dberror(query, err)
// 	}
// 	defer rows.Close()
//
// 	var results []WarehouseRecord
// 	for rows.Next() {
// 		var rec WarehouseRecord
// 		err := rows.Scan(&rec.ID, &rec.Description, &rec.Address1, &rec.Address2, &rec.City, &rec.State, &rec.Zipcode, &rec.Country)
// 		if err != nil {
// 			return err
// 		}
// 		results = append(results, rec)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return err
// 	}
//
// 	f.All = results
// 	return nil
// }
//
// // GetWarehouse fetches warehouse and address info for the given ID and checks if it is orphaned
// func (f *Form) GetWarehouse(cfg Config) error {
// 	db, err := dbConnect(cfg)
// 	if err != nil {
// 		return err
// 	}
// 	defer db.Close()
//
// 	if f.ID <= 0 {
// 		return fmt.Errorf("invalid warehouse ID: %d", f.ID)
// 	}
//
// 	query := `SELECT w.description, a.address1, a.address2, a.city,
//                  a.state, a.zipcode, a.country
//                  FROM warehouse w
//                  JOIN address a ON a.trans_id = w.id
//                  WHERE w.id = $1`
//
// 	row := db.QueryRow(query, f.ID)
//
// 	var desc, a1, a2, city, state, zip, country sql.NullString
// 	err = row.Scan(&desc, &a1, &a2, &city, &state, &zip, &country)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return fmt.Errorf("warehouse with id %d not found", f.ID)
// 		}
// 		return dberror(query, err)
// 	}
//
// 	if desc.Valid {
// 		f.Description = desc.String
// 	}
// 	if a1.Valid {
// 		f.Address1 = a1.String
// 	}
// 	if a2.Valid {
// 		f.Address2 = a2.String
// 	}
// 	if city.Valid {
// 		f.City = city.String
// 	}
// 	if state.Valid {
// 		f.State = state.String
// 	}
// 	if zip.Valid {
// 		f.Zipcode = zip.String
// 	}
// 	if country.Valid {
// 		f.Country = country.String
// 	}
//
// 	// Check if warehouse is orphaned (no inventory referencing it)
// 	query = `SELECT 1 FROM inventory WHERE warehouse_id = $1 LIMIT 1`
// 	row = db.QueryRow(query, f.ID)
// 	var exists int
// 	err = row.Scan(&exists)
// 	if err == sql.ErrNoRows {
// 		// No inventory found, orphaned = true
// 		f.Orphaned = true
// 	} else if err != nil {
// 		return dberror(query, err)
// 	} else {
// 		// Inventory exists, orphaned = false
// 		f.Orphaned = false
// 	}
//
// 	return nil
// }
//
// // SaveWarehouse inserts or updates a warehouse and its address
// func (f *Form) SaveWarehouse(cfg Config) error {
// 	db, tx, err := dbConnectNoAuto(cfg)
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		// If tx is still valid, rollback on exit to avoid hanging transactions
// 		if tx != nil {
// 			tx.Rollback()
// 		}
// 		db.Close()
// 	}()
//
// 	// Clean description: replace multiple dashes with a single dash, multiple spaces with one space
// 	f.Description = strings.ReplaceAll(f.Description, "--", "-")
// 	for strings.Contains(f.Description, "--") {
// 		f.Description = strings.ReplaceAll(f.Description, "--", "-")
// 	}
// 	f.Description = strings.Join(strings.Fields(f.Description), " ")
//
// 	// Check if ID exists in DB
// 	if f.ID > 0 {
// 		query := `SELECT id FROM warehouse WHERE id = $1`
// 		var existingID int64
// 		err = tx.QueryRow(query, f.ID).Scan(&existingID)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				// ID does not exist, reset to 0 to insert new
// 				f.ID = 0
// 			} else {
// 				return dberror(query, err)
// 			}
// 		} else {
// 			f.ID = existingID
// 		}
// 	}
//
// 	if f.ID == 0 {
// 		// Insert new warehouse with temporary unique description (localtime + pid)
// 		uid := fmt.Sprintf("%d%d", time.Now().UnixNano(), getPID())
//
// 		query := `INSERT INTO warehouse (description) VALUES ($1) RETURNING id`
// 		err = tx.QueryRow(query, uid).Scan(&f.ID)
// 		if err != nil {
// 			return dberror(query, err)
// 		}
//
// 		// Insert address with trans_id = new warehouse id
// 		query = `INSERT INTO address (trans_id) VALUES ($1)`
// 		_, err = tx.Exec(query, f.ID)
// 		if err != nil {
// 			return dberror(query, err)
// 		}
// 	}
//
// 	// Update warehouse description
// 	query := `UPDATE warehouse SET description = $1 WHERE id = $2`
// 	_, err = tx.Exec(query, f.Description, f.ID)
// 	if err != nil {
// 		return dberror(query, err)
// 	}
//
// 	// Update address fields
// 	query = `UPDATE address SET
//               address1 = $1,
//               address2 = $2,
//               city = $3,
//               state = $4,
//               zipcode = $5,
//               country = $6
//               WHERE trans_id = $7`
// 	_, err = tx.Exec(query,
// 		nullString(f.Address1),
// 		nullString(f.Address2),
// 		nullString(f.City),
// 		nullString(f.State),
// 		nullString(f.Zipcode),
// 		nullString(f.Country),
// 		f.ID)
// 	if err != nil {
// 		return dberror(query, err)
// 	}
//
// 	// Commit transaction
// 	err = tx.Commit()
// 	if err != nil {
// 		return err
// 	}
// 	// Prevent deferred rollback
// 	tx = nil
//
// 	return nil
// }
//
// // DeleteWarehouse deletes warehouse and its address by ID
// func (f *Form) DeleteWarehouse(cfg Config) error {
// 	db, tx, err := dbConnectNoAuto(cfg)
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		if tx != nil {
// 			tx.Rollback()
// 		}
// 		db.Close()
// 	}()
//
// 	if f.ID <= 0 {
// 		return fmt.Errorf("invalid warehouse ID: %d", f.ID)
// 	}
//
// 	query := `DELETE FROM warehouse WHERE id = $1`
// 	_, err = tx.Exec(query, f.ID)
// 	if err != nil {
// 		return dberror(query, err)
// 	}
//
// 	query = `DELETE FROM address WHERE trans_id = $1`
// 	_, err = tx.Exec(query, f.ID)
// 	if err != nil {
// 		return dberror(query, err)
// 	}
//
// 	err = tx.Commit()
// 	if err != nil {
// 		return err
// 	}
// 	tx = nil
//
// 	return nil
// }
//
// // Helper to get PID (process id)
// func getPID() int {
// 	return 0 // Placeholder, Go doesn't provide direct PID easily, can use os.Getpid()
// }
//
// // nullString returns sql.NullString for a string, treating empty string as NULL
// func nullString(s string) sql.NullString {
// 	if strings.TrimSpace(s) == "" {
// 		return sql.NullString{Valid: false}
// 	}
// 	return sql.NullString{
// 		String: s,
// 		Valid:  true,
// 	}
// }
