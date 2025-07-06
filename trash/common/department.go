package common

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"regexp"
// 	"strings"
//
// 	_ "github.com/lib/pq" // Assuming PostgreSQL; adjust import for your DB
// )
//
// // Config holds database configuration details
// type departmentConfig struct {
// 	Driver   string
// 	Host     string
// 	Port     int
// 	User     string
// 	Password string
// 	DBName   string
// 	SSLMode  string
// }
//
// // Form holds form data and database connection info
// type departmentForm struct {
// 	DB         *sql.DB
// 	SortColumn string
// 	Direction  string
// 	All        []Department
// 	ID         int
// 	Description string
// 	Role        string
// 	Orphaned   bool
// }
//
// // Department represents a department record
// type Department struct {
// 	ID          int
// 	Description string
// 	Role        string
// }
//
// // dbConnect connects to the database using provided config
// func (f *Form) dbConnect(cfg Config) error {
// 	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
// 		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
// 	db, err := sql.Open(cfg.Driver, dsn)
// 	if err != nil {
// 		return err
// 	}
// 	// Verify connection
// 	if err := db.Ping(); err != nil {
// 		return err
// 	}
// 	f.DB = db
// 	return nil
// }
//
// // dberror handles database errors by logging and panicking
// func (f *Form) dberror(query string, err error) {
// 	log.Fatalf("Database error during query [%s]: %v", query, err)
// }
//
// // sortOrder sets default sort column and direction if not set
// func (f *Form) sortOrder() {
// 	if f.SortColumn == "" {
// 		f.SortColumn = "description"
// 	}
// 	if f.Direction != "ASC" && f.Direction != "DESC" {
// 		f.Direction = "ASC"
// 	}
// }
//
// // departments fetches all departments ordered by description and direction
// func (f *Form) departments(cfg Config) error {
// 	if err := f.dbConnect(cfg); err != nil {
// 		return err
// 	}
// 	defer f.DB.Close()
//
// 	f.sortOrder()
//
// 	query := fmt.Sprintf(`SELECT id, description, role FROM department ORDER BY %s %s`, f.SortColumn, f.Direction)
//
// 	rows, err := f.DB.Query(query)
// 	if err != nil {
// 		f.dberror(query, err)
// 		return err
// 	}
// 	defer rows.Close()
//
// 	f.All = []Department{}
// 	for rows.Next() {
// 		var d Department
// 		err := rows.Scan(&d.ID, &d.Description, &d.Role)
// 		if err != nil {
// 			return err
// 		}
// 		f.All = append(f.All, d)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// // getDepartment fetches a single department by id and checks if it is orphaned
// func (f *Form) getDepartment(cfg Config) error {
// 	if err := f.dbConnect(cfg); err != nil {
// 		return err
// 	}
// 	defer f.DB.Close()
//
// 	// Ensure ID is positive
// 	if f.ID <= 0 {
// 		return fmt.Errorf("invalid department id: %d", f.ID)
// 	}
//
// 	query := `SELECT description, role FROM department WHERE id = $1`
// 	err := f.DB.QueryRow(query, f.ID).Scan(&f.Description, &f.Role)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return fmt.Errorf("department with id %d not found", f.ID)
// 		}
// 		f.dberror(query, err)
// 		return err
// 	}
//
// 	// Check if department is orphaned (not used in dpt_trans)
// 	checkQuery := `SELECT EXISTS (SELECT 1 FROM dpt_trans WHERE department_id = $1)`
// 	var inUse bool
// 	err = f.DB.QueryRow(checkQuery, f.ID).Scan(&inUse)
// 	if err != nil {
// 		f.dberror(checkQuery, err)
// 		return err
// 	}
// 	f.Orphaned = !inUse
//
// 	return nil
// }
//
// // saveDepartment inserts or updates a department record
// func (f *Form) saveDepartment(cfg Config) error {
// 	if err := f.dbConnect(cfg); err != nil {
// 		return err
// 	}
// 	defer f.DB.Close()
//
// 	// Clean up description string
// 	f.Description = cleanDescription(f.Description)
//
// 	if f.ID > 0 {
// 		// Update existing
// 		query := `UPDATE department SET description = $1, role = $2 WHERE id = $3`
// 		_, err := f.DB.Exec(query, f.Description, f.Role, f.ID)
// 		if err != nil {
// 			f.dberror(query, err)
// 			return err
// 		}
// 	} else {
// 		// Insert new
// 		query := `INSERT INTO department (description, role) VALUES ($1, $2)`
// 		_, err := f.DB.Exec(query, f.Description, f.Role)
// 		if err != nil {
// 			f.dberror(query, err)
// 			return err
// 		}
// 	}
//
// 	return nil
// }
//
// // deleteDepartment deletes a department by id
// func (f *Form) deleteDepartment(cfg Config) error {
// 	if err := f.dbConnect(cfg); err != nil {
// 		return err
// 	}
// 	defer f.DB.Close()
//
// 	if f.ID <= 0 {
// 		return fmt.Errorf("invalid department id: %d", f.ID)
// 	}
//
// 	query := `DELETE FROM department WHERE id = $1`
// 	_, err := f.DB.Exec(query, f.ID)
// 	if err != nil {
// 		f.dberror(query, err)
// 		return err
// 	}
//
// 	return nil
// }
//
// // cleanDescription cleans the description string as per original regex replacements
// func cleanDescription(desc string) string {
// 	// Replace multiple dashes with a single dash
// 	reDash := regexp.MustCompile(`-(-)+`)
// 	desc = reDash.ReplaceAllString(desc, "-")
//
// 	// Replace multiple spaces with single space
// 	reSpace := regexp.MustCompile(` ( )+`)
// 	desc = reSpace.ReplaceAllString(desc, " ")
//
// 	// Trim leading and trailing spaces as good practice
// 	desc = strings.TrimSpace(desc)
//
// 	return desc
// }
//
// func example() {
// 	// Example usage:
//
// 	cfg := Config{
// 		Driver:   "postgres",
// 		Host:     "localhost",
// 		Port:     5432,
// 		User:     "youruser",
// 		Password: "yourpass",
// 		DBName:   "yourdb",
// 		SSLMode:  "disable",
// 	}
//
// 	form := Form{
// 		Direction: "ASC",
// 	}
//
// 	err := form.departments(cfg)
// 	if err != nil {
// 		log.Fatalf("Error fetching departments: %v", err)
// 	}
//
// 	for _, d := range form.All {
// 		fmt.Printf("ID: %d, Description: %s, Role: %s\n", d.ID, d.Description, d.Role)
// 	}
//
// 	// Fetch a single department example
// 	form.ID = 1
// 	err = form.getDepartment(cfg)
// 	if err != nil {
// 		log.Printf("Error fetching department: %v", err)
// 	} else {
// 		fmt.Printf("Department ID %d: Description=%s, Role=%s, Orphaned=%v\n",
// 			form.ID, form.Description, form.Role, form.Orphaned)
// 	}
//
// 	// Save a department example
// 	form.Description = "New Department--Name"
// 	form.Role = "Manager"
// 	form.ID = 0 // New department
// 	err = form.saveDepartment(cfg)
// 	if err != nil {
// 		log.Printf("Error saving department: %v", err)
// 	} else {
// 		fmt.Println("Department saved successfully")
// 	}
//
// 	// Delete a department example
// 	form.ID = 2
// 	err = form.deleteDepartment(cfg)
// 	if err != nil {
// 		log.Printf("Error deleting department: %v", err)
// 	} else {
// 		fmt.Println("Department deleted successfully")
// 	}
// }
//
