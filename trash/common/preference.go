package common

//
// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"path/filepath"
// )
//
// // UserPreferences represents user preference data
// type UserPreferences struct {
// 	Login        string `json:"login"`
// 	Name         string `json:"name"`
// 	Role         string `json:"role"`
// 	Tel          string `json:"tel"`
// 	OldPassword  string `json:"old_password"`
// 	NewPassword  string `json:"new_password"`
// 	Company      string `json:"company"`
// 	SessionCookie string `json:"sessioncookie"`
// 	// Add other form fields as needed
// 	ExtraFields map[string]any `json:"extra_fields"`
// }
//
// // User represents user configuration data
// type User struct {
// 	Login         string                 `json:"login"`
// 	Name          string                 `json:"name"`
// 	Role          string                 `json:"role"`
// 	WorkPhone     string                 `json:"workphone"`
// 	Company       string                 `json:"company"`
// 	Password      string                 `json:"password"`
// 	SessionCookie string                 `json:"sessioncookie"`
// 	Config        map[string]any `json:"config"`
// }
//
// // Config represents database and application configuration
// type Config struct {
// 	DBDriver     string
// 	DBHost       string
// 	DBPort       int
// 	DBName       string
// 	DBUser       string
// 	DBPassword   string
// 	DateFormat   string
// }
//
// // UserService handles user-related operations
// type UserService struct {
// 	db *sql.DB
// }
//
// // NewUserService creates a new UserService
// func NewUserService(db *sql.DB) *UserService {
// 	return &UserService{db: db}
// }
//
// // SavePreferences saves user preferences to database and member file
// func (us *UserService) SavePreferences(config *Config, prefs *UserPreferences, memberFile, usersPath string) error {
// 	// Update employee record in database
// 	err := us.updateEmployeeRecord(prefs)
// 	if err != nil {
// 		return fmt.Errorf("failed to update employee record: %w", err)
// 	}
//
// 	// Get company default
// 	company, err := us.getCompanyDefault()
// 	if err != nil {
// 		return fmt.Errorf("failed to get company default: %w", err)
// 	}
//
// 	// Load existing user configuration
// 	user, err := us.loadUserConfig(memberFile, prefs.Login)
// 	if err != nil {
// 		return fmt.Errorf("failed to load user config: %w", err)
// 	}
//
// 	// Update user configuration with form data
// 	us.updateUserConfig(user, prefs, company)
//
// 	// Save user configuration to member file
// 	err = us.saveUserConfig(user, memberFile, usersPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to save user config: %w", err)
// 	}
//
// 	// Update session cookie in preferences
// 	prefs.SessionCookie = user.SessionCookie
//
// 	return nil
// }
//
// // updateEmployeeRecord updates the employee record in the database
// func (us *UserService) updateEmployeeRecord(prefs *UserPreferences) error {
// 	query := `UPDATE employee SET
// 	          name = $1,
// 	          role = $2,
// 	          workphone = $3
// 	          WHERE login = $4`
//
// 	_, err := us.db.Exec(query, prefs.Name, prefs.Role, prefs.Tel, prefs.Login)
// 	if err != nil {
// 		return fmt.Errorf("failed to execute update query: %w", err)
// 	}
//
// 	return nil
// }
//
// // getCompanyDefault retrieves the default company setting
// func (us *UserService) getCompanyDefault() (string, error) {
// 	var company string
// 	query := `SELECT value FROM defaults WHERE setting_key = 'company'`
//
// 	err := us.db.QueryRow(query).Scan(&company)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			// Return empty string if no default company is set
// 			return "", nil
// 		}
// 		return "", fmt.Errorf("failed to get company default: %w", err)
// 	}
//
// 	return company, nil
// }
//
// // loadUserConfig loads user configuration from member file
// func (us *UserService) loadUserConfig(memberFile, login string) (*User, error) {
// 	user := &User{
// 		Login:  login,
// 		Config: make(map[string]any),
// 	}
//
// 	// In a real implementation, you would read from the member file
// 	// This is a simplified version that demonstrates the structure
// 	// The original Perl code uses a custom User class to read from files
//
// 	// For demonstration, we'll create a basic user structure
// 	// In practice, you'd implement file reading logic here
// 	user.Login = login
//
// 	return user, nil
// }
//
// // updateUserConfig updates user configuration with form preferences
// func (us *UserService) updateUserConfig(user *User, prefs *UserPreferences, company string) {
// 	// Update basic fields
// 	user.Name = prefs.Name
// 	user.Role = prefs.Role
// 	user.WorkPhone = prefs.Tel
// 	user.Company = company
//
// 	// Update password if it changed
// 	if prefs.OldPassword != prefs.NewPassword && prefs.NewPassword != "" {
// 		user.Password = prefs.NewPassword
// 	}
//
// 	// Copy all extra fields from preferences
// 	if prefs.ExtraFields != nil {
// 		for key, value := range prefs.ExtraFields {
// 			user.Config[key] = value
// 		}
// 	}
// }
//
// // saveUserConfig saves user configuration to member file
// func (us *UserService) saveUserConfig(user *User, memberFile, usersPath string) error {
// 	// In the original Perl code, this calls $myconfig->save_member()
// 	// This would typically write to a configuration file or database
//
// 	// For demonstration purposes, we'll show the structure
// 	// In a real implementation, you'd write to the actual member file
//
// 	memberPath := filepath.Join(usersPath, memberFile)
//
// 	// Here you would implement the actual file writing logic
// 	// This might involve writing to a JSON file, INI file, or database
// 	// depending on your application's configuration format
//
// 	log.Printf("Saving user config to %s for login %s", memberPath, user.Login)
//
// 	// Example of what you might do:
// 	// return us.writeUserConfigFile(memberPath, user)
//
// 	return nil
// }
//
// // writeUserConfigFile writes user configuration to a file
// func (us *UserService) writeUserConfigFile(filePath string, user *User) error {
// 	// This would implement the actual file writing logic
// 	// The format would depend on your application's needs
//
// 	// Example implementation might use JSON, YAML, or custom format
// 	// For now, this is just a placeholder
//
// 	log.Printf("Would write user config to %s", filePath)
// 	return nil
// }
//
// // Alternative approach using a more structured preference system
// type PreferenceManager struct {
// 	db          *sql.DB
// 	userService *UserService
// }
//
// // NewPreferenceManager creates a new preference manager
// func NewPreferenceManager(db *sql.DB) *PreferenceManager {
// 	return &PreferenceManager{
// 		db:          db,
// 		userService: NewUserService(db),
// 	}
// }
//
// // SaveUserPreferences is a higher-level method that handles the complete preference saving process
// func (pm *PreferenceManager) SaveUserPreferences(config *Config, prefs *UserPreferences, memberFile, usersPath string) error {
// 	// Begin transaction for atomic operations
// 	tx, err := pm.db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback()
//
// 	// Create a temporary user service with the transaction
// 	txUserService := &UserService{db: tx}
//
// 	// Update employee record
// 	err = txUserService.updateEmployeeRecord(prefs)
// 	if err != nil {
// 		return fmt.Errorf("failed to update employee: %w", err)
// 	}
//
// 	// Get company default
// 	company, err := txUserService.getCompanyDefault()
// 	if err != nil {
// 		return fmt.Errorf("failed to get company default: %w", err)
// 	}
//
// 	// Commit database changes
// 	err = tx.Commit()
// 	if err != nil {
// 		return fmt.Errorf("failed to commit transaction: %w", err)
// 	}
//
// 	// Handle file-based user configuration
// 	user, err := pm.userService.loadUserConfig(memberFile, prefs.Login)
// 	if err != nil {
// 		return fmt.Errorf("failed to load user config: %w", err)
// 	}
//
// 	pm.userService.updateUserConfig(user, prefs, company)
//
// 	err = pm.userService.saveUserConfig(user, memberFile, usersPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to save user config: %w", err)
// 	}
//
// 	// Update session cookie
// 	prefs.SessionCookie = user.SessionCookie
//
// 	return nil
// }
//
// // Validation functions
// func (prefs *UserPreferences) Validate() error {
// 	if prefs.Login == "" {
// 		return fmt.Errorf("login is required")
// 	}
// 	if prefs.Name == "" {
// 		return fmt.Errorf("name is required")
// 	}
// 	// Add other validation rules as needed
// 	return nil
// }
//
// // Example usage and integration
// func Example() {
// 	// Example of how to use the preference system
// 	// db, err := sql.Open("postgres", "your-connection-string")
// 	// if err != nil {
// 	//     log.Fatal(err)
// 	// }
// 	// defer db.Close()
//
// 	// config := &Config{
// 	//     DBDriver: "postgres",
// 	//     // ... other config
// 	// }
//
// 	// prefs := &UserPreferences{
// 	//     Login:       "john.doe",
// 	//     Name:        "John Doe",
// 	//     Role:        "admin",
// 	//     Tel:         "555-1234",
// 	//     OldPassword: "old_pass",
// 	//     NewPassword: "new_pass",
// 	//     ExtraFields: make(map[string]any),
// 	// }
//
// 	// pm := NewPreferenceManager(db)
// 	// err = pm.SaveUserPreferences(config, prefs, "members.conf", "/path/to/users")
// 	// if err != nil {
// 	//     log.Fatal(err)
// 	// }
//
// 	log.Println("User preferences service ready")
// }
//
// // Additional helper functions for file handling
// type FileUserConfig struct {
// 	Login         string                 `json:"login"`
// 	Name          string                 `json:"name"`
// 	Role          string                 `json:"role"`
// 	WorkPhone     string                 `json:"workphone"`
// 	Company       string                 `json:"company"`
// 	Password      string                 `json:"password,omitempty"`
// 	SessionCookie string                 `json:"sessioncookie"`
// 	Settings      map[string]any `json:"settings"`
// }
//
// // ConvertUserToFileConfig converts User struct to file configuration format
// func ConvertUserToFileConfig(user *User) *FileUserConfig {
// 	return &FileUserConfig{
// 		Login:         user.Login,
// 		Name:          user.Name,
// 		Role:          user.Role,
// 		WorkPhone:     user.WorkPhone,
// 		Company:       user.Company,
// 		Password:      user.Password,
// 		SessionCookie: user.SessionCookie,
// 		Settings:      user.Config,
// 	}
// }
