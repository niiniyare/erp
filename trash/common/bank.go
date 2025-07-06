package common

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"strings"
// )
//
// // Config represents configuration settings.
// type Config struct {
// 	CountryCode string
// 	// Add other configuration fields as necessary
// }
//
// // Form represents the form data structure.
// type Form struct {
// 	DB *sql.DB
// 	// Add other form fields as necessary
// 	ALL []map[string]interface{}
// 	ID int
// 	// Add other fields from the form as necessary
// }
//
// // BankAccount represents a bank account record.
// type BankAccount struct {
// 	ID int
// 	AccNo string
// 	Description string
// 	Name string
// 	IBAN string
// 	BIC string
// 	MemberNumber string
// 	DCN string
// 	RVC string
// 	QRBAN string
// 	StrDbkgInf string
// 	InvDescriptionQR string
// 	Address1 string
// 	Address2 string
// 	City string
// 	State string
// 	ZipCode string
// 	Country string
// 	Translation string
// }
//
// // bankAccounts retrieves bank accounts from the database.
// func bankAccounts(config *Config, form *Form) error {
// 	// Connect to the database
// 	dbh, err := sql.Open("your-driver", "your-dsn")
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to database: %v", err)
// 	}
// 	defer dbh.Close()
//
// 	query := `
// 		SELECT c.id, c.accno, c.description,
// 			bk.name, bk.iban, bk.bic, bk.membernumber, bk.dcn, bk.rvc,
// 			bk.qriban, bk.strdbkginf, bk.invdescriptionqr,
// 			ad.address1, ad.address2, ad.city,
// 			ad.state, ad.zipcode, ad.country,
// 			l.description AS translation
// 		FROM chart c
// 		LEFT JOIN bank bk ON (bk.id = c.id)
// 		LEFT JOIN address ad ON (c.id = ad.trans_id)
// 		LEFT JOIN translation l ON (l.trans_id = c.id AND l.language_code = ?)
// 		WHERE c.link LIKE '%AR_paid%'
// 		ORDER BY c.accno`
//
// 	rows, err := dbh.Query(query, config.CountryCode)
// 	if err != nil {
// 		return fmt.Errorf("failed to execute query: %v", err)
// 	}
// 	defer rows.Close()
//
// 	for rows.Next() {
// 		var ba BankAccount
// 		var address1, address2, city, state, zipcode, country sql.NullString
// 		var translation sql.NullString
//
// 		err := rows.Scan(
// 			&ba.ID, &ba.AccNo, &ba.Description,
// 			&ba.Name, &ba.IBAN, &ba.BIC, &ba.MemberNumber, &ba.DCN, &ba.RVC,
// 			&ba.QRBAN, &ba.StrDbkgInf, &ba.InvDescriptionQR,
// 			&address1, &address2, &city, &state, &zipcode, &country,
// 			&translation,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("failed to scan row: %v", err)
// 		}
//
// 		// Build address string
// 		address := []string{}
// 		if address1.Valid {
// 			address = append(address, address1.String)
// 		}
// 		if address2.Valid {
// 			address = append(address, address2.String)
// 		}
// 		if city.Valid {
// 			address = append(address, city.String)
// 		}
// 		if state.Valid {
// 			address = append(address, state.String)
// 		}
// 		if zipcode.Valid {
// 			address = append(address, zipcode.String)
// 		}
// 		if country.Valid {
// 			address = append(address, country.String)
// 		}
//
// 		ba.Address1 = strings.Join(address, "\n")
//
// 		// Use translation if available
// 		if translation.Valid {
// 			ba.Description = translation.String
// 		}
//
// 		// Append to form's ALL slice
// 		form.ALL = append(form.ALL, map[string]interface{}{
// 			"id": ba.ID, "accno": ba.AccNo, "description": ba.Description,
// 			"name": ba.Name, "iban": ba.IBAN, "bic": ba.BIC,
// 			"membernumber": ba.MemberNumber, "dcn": ba.DCN, "rvc": ba.RVC,
// 			"qriban": ba.QRBAN, "strdbkginf": ba.StrDbkgInf,
// 			"invdescriptionqr": ba.InvDescriptionQR, "address": ba.Address1,
// 		})
// 	}
//
// 	return nil
// }
//
// // getBank retrieves a specific bank account from the database.
// func getBank(config *Config, form *Form) error {
// 	// Connect to the database
// 	dbh, err := sql.Open("your-driver", "your-dsn")
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to database: %v", err)
// 	}
// 	defer dbh.Close()
//
// 	query := `
// 		SELECT c.accno, c.description,
// 			bk.name, bk.iban, bk.bic, bk.membernumber, bk.dcn, bk.rvc,
// 			bk.qriban, bk.strdbkginf, bk.invdescriptionqr,
// 			ad.address1, ad.address2, ad.city,
// 			ad.state, ad.zipcode, ad.country,
// 			l.description AS translation
// 		FROM chart c
// 		LEFT JOIN bank bk ON (c.id = bk.id)
// 		LEFT JOIN address ad ON (c.id = ad.trans_id)
// 		LEFT JOIN translation l ON (l.trans_id = c.id AND l.language_code = ?)
// 		WHERE c.id = ?`
//
// 	row := dbh.QueryRow(query, config.CountryCode, form.ID)
//
// 	var ba BankAccount
// 	var translation sql.NullString
//
// 	err = row.Scan(
// 		&ba.AccNo, &ba.Description,
// 		&ba.Name, &ba.IBAN, &ba.BIC, &ba.MemberNumber, &ba.DCN, &ba.RVC,
// 		&ba.QRBAN, &ba.StrDbkgInf, &ba.InvDescriptionQR,
// 		&ba.Address1, &ba.Address2, &ba.City,
// 		&ba.State, &ba.ZipCode, &ba.Country,
// 		&translation,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to scan row: %v", err)
// 	}
//
// 	// Use translation if available
// 	if translation.Valid {
// 		ba.Description = translation.String
// 	}
//
// 	// Set form fields
// 	form.AccNo = ba.AccNo
// 	form.Description = ba.Description
// 	// Set other form fields as necessary
//
// 	return nil
// }
//
// // saveBank saves or updates a bank account in the database.
// func saveBank(config *Config, form *Form) (bool, error) {
// 	// Connect to the database
// 	dbh, err := sql.Open("your-driver", "your-dsn")
// 	if err != nil {
// 		return false, fmt.Errorf("failed to connect to database: %v", err)
// 	}
// 	defer dbh.Close()
//
// 	// Check if the bank record exists
// 	var id int
// 	err = dbh.QueryRow("SELECT id FROM bank WHERE id = ?", form.ID).Scan(&id)
// 	if err != nil && err != sql.ErrNoRows {
// 		return false, fmt.Errorf("failed to check bank existence: %v", err)
// 	}
//
// 	// Check if any of the fields are non-empty
// 	ok := false
// 	fields := []string{"name", "iban", "bic", "address1", "address2", "city", "state", "zipcode", "country", "membernumber", "rvc", "dcn", "qriban", "strdbkginf", "invdescriptionqr"}
// 	for _, field := range fields {
// 		// Use reflection or a map to check form fields dynamically
// 		// For simplicity, assuming non-empty check here
// 		if true { // Replace with actual check
// 			ok = true
// 			break
// 		}
// 	}
//
// 	if ok {
// 		if id != 0 {
// 			// Update existing bank record
// 			query := `
// 				UPDATE bank SET
// 					name = ?, iban = ?, bic = ?, membernumber = ?, rvc = ?,
// 					qriban = ?, strdbkginf = ?, invdescriptionqr = ?, dcn = ?
// 				WHERE id = ?`
// 			_, err = dbh.Exec(query, form.Name, form.IBAN, form.BIC, form.MemberNumber, form.RVC, form.QRBAN, form.StrDbkgInf, form.InvDescriptionQR, form.DCN, form.ID)
// 			if err != nil {
// 				return false, fmt.Errorf("failed to update bank: %v", err)
// 			}
// 		} else {
// 			// Insert new bank record
// 			query := `
// 				INSERT INTO bank (id, name, iban, bic, membernumber, rvc, dcn, qriban, strdbkginf, invdescriptionqr)
// 				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
// 			_, err = dbh.Exec(query, form.ID, form.Name, form.IBAN, form.BIC, form.MemberNumber, form.RVC, form.DCN, form.QRBAN, form.StrDbkgInf, form.InvDescriptionQR)
// 			if err != nil {
// 				return false, fmt.Errorf("failed to insert bank: %v", err)
// 			}
//
// 			// Insert new address record
// 			query = `INSERT INTO address (id, trans_id) VALUES (?, ?)`
// 			_, err = dbh.Exec(query, id, form.ID)
// 			if err != nil {
// 				return false, fmt.Errorf("failed to insert address: %v", err)
// 			}
// 		}
//
// 		// Update address record
// 		query := `
// 			UPDATE address SET
// 				address1 = ?, address2 = ?, city = ?, state = ?, zipcode = ?, country = ?
// 			WHERE trans_id = ?`
// 		_, err = dbh.Exec(query, form.Address1, form.Address2, form.City, form.State, form.ZipCode, form.Country, form.ID)
// 		if err != nil {
// 			return false, fmt.Errorf("failed to update address: %v", err)
// 		}
// 	} else {
// 		// Delete bank and address records
// 		query := `DELETE FROM bank WHERE id = ?`
// 		_, err = dbh.Exec(query, form.ID)
// 		if err != nil {
// 			return false, fmt.Errorf("failed to delete bank: %v", err)
// 		}
//
// 		query = `DELETE FROM address WHERE trans_id = ?`
// 		_, err = dbh.Exec(query, form.ID)
// 		if err != nil {
// 			return false, fmt.Errorf("failed to delete address: %v", err)
// 		}
// 	}
//
// 	// Commit the transaction
// 	err = dbh.(*sql.Tx).Commit()
// 	if err != nil {
// 		return false, fmt.Errorf("failed to commit transaction: %v", err)
// 	}
//
// 	return true, nil
// }
//
// func Example() {
// 	config := &Config{CountryCode: "US"} // Example configuration
// 	form := &Form{ID: 1} // Example form data
//
// 	// Example usage
// 	err := bankAccounts(config, form)
// 	if err != nil {
// 		log.Fatalf("Error retrieving bank accounts: %v", err)
// 	}
//
// 	err = getBank(config, form)
// 	if err != nil {
// 		log.Fatalf("Error retrieving bank: %v", err)
// 	}
//
// 	success, err := saveBank(config, form)
// 	if err != nil {
// 		log.Fatalf("Error saving bank: %v", err)
// 	}
//
// 	fmt.Println("Operation successful:", success)
// }
//
