# Password Component

**FILE PURPOSE**: Secure password input control with visibility toggle and strength validation  
**SCOPE**: Password input, strength checking, visibility controls, and security validation  
**TARGET AUDIENCE**: Developers implementing authentication, user management, and security forms

## 📋 Component Overview

Password Component provides secure password input functionality with visibility toggle, strength indicators, validation rules, and formatting options. Essential for authentication and user account management in ERP systems.

### Schema Reference
- **Primary Schema**: `PasswordSchema.json`
- **Related Schemas**: `TextControlSchema.json`, `ValidationSchema.json`
- **Base Interface**: Secure text input control for passwords

## Basic Usage

```json
{
    "type": "password",
    "name": "password",
    "label": "Password",
    "placeholder": "Enter your password"
}
```

## Go Type Definition

```go
type PasswordProps struct {
    Type            string              `json:"type"`
    Name            string              `json:"name"`
    Label           interface{}         `json:"label"`
    Placeholder     string              `json:"placeholder"`
    Value           string              `json:"value"`
    Required        bool                `json:"required"`
    Disabled        bool                `json:"disabled"`
    ReadOnly        bool                `json:"readOnly"`
    ShowToggle      bool                `json:"showToggle"`
    ShowStrength    bool                `json:"showStrength"`
    MinLength       int                 `json:"minLength"`
    MaxLength       int                 `json:"maxLength"`
    Pattern         string              `json:"pattern"`
    AutoComplete    string              `json:"autoComplete"`    // "current-password", "new-password"
    MosaicText      string              `json:"mosaicText"`      // Masked display text
}
```

## Essential Variants

### Basic Password Input
```json
{
    "type": "password",
    "name": "current_password",
    "label": "Current Password",
    "placeholder": "Enter current password",
    "required": true,
    "autoComplete": "current-password"
}
```

### New Password with Strength
```json
{
    "type": "password",
    "name": "new_password",
    "label": "New Password",
    "placeholder": "Create a strong password",
    "required": true,
    "showToggle": true,
    "showStrength": true,
    "autoComplete": "new-password",
    "validations": {
        "minLength": 8,
        "matchRegexp": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]"
    },
    "validationErrors": {
        "minLength": "Password must be at least 8 characters long",
        "matchRegexp": "Password must contain uppercase, lowercase, number and special character"
    }
}
```

### Password Confirmation
```json
{
    "type": "password",
    "name": "confirm_password",
    "label": "Confirm Password",
    "placeholder": "Confirm your password",
    "required": true,
    "showToggle": true,
    "autoComplete": "new-password"
}
```

### Admin Password Override
```json
{
    "type": "password",
    "name": "admin_password",
    "label": "Administrator Password",
    "placeholder": "Enter administrator password for verification",
    "required": true,
    "showToggle": false,
    "autoComplete": "current-password"
}
```

## Real-World Use Cases

### User Registration
```json
{
    "type": "password",
    "name": "registration_password",
    "label": "Password",
    "placeholder": "Create a secure password",
    "required": true,
    "showToggle": true,
    "showStrength": true,
    "autoComplete": "new-password",
    "validations": {
        "minLength": 8,
        "maxLength": 128,
        "matchRegexp": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]{8,}$"
    },
    "validationErrors": {
        "minLength": "Password must be at least 8 characters",
        "maxLength": "Password cannot exceed 128 characters",
        "matchRegexp": "Password must contain: uppercase letter, lowercase letter, number, and special character (@$!%*?&)"
    },
    "hint": "Use a mix of letters, numbers, and symbols for stronger security"
}
```

### Login Form
```json
{
    "type": "password",
    "name": "login_password",
    "label": "Password",
    "placeholder": "Enter your password",
    "required": true,
    "showToggle": true,
    "autoComplete": "current-password"
}
```

### Password Reset
```json
{
    "type": "password",
    "name": "reset_password",
    "label": "New Password",
    "placeholder": "Enter your new password",
    "required": true,
    "showToggle": true,
    "showStrength": true,
    "autoComplete": "new-password",
    "validations": {
        "minLength": 8,
        "matchRegexp": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)"
    },
    "validationErrors": {
        "minLength": "New password must be at least 8 characters",
        "matchRegexp": "Password must contain uppercase, lowercase, and number"
    }
}
```

### Employee Account Setup
```json
{
    "type": "password",
    "name": "employee_password",
    "label": "Temporary Password",
    "placeholder": "Generate or enter temporary password",
    "required": true,
    "showToggle": true,
    "autoComplete": "new-password",
    "validations": {
        "minLength": 12
    },
    "validationErrors": {
        "minLength": "Employee passwords must be at least 12 characters"
    },
    "hint": "Employee will be required to change this password on first login"
}
```

### API Key Password
```json
{
    "type": "password",
    "name": "api_password",
    "label": "API Access Password",
    "placeholder": "Enter password to access API keys",
    "required": true,
    "showToggle": false,
    "autoComplete": "current-password"
}
```

### Database Connection
```json
{
    "type": "password",
    "name": "db_password",
    "label": "Database Password",
    "placeholder": "Enter database connection password",
    "required": true,
    "showToggle": true,
    "autoComplete": "off",
    "mosaicText": "•••••••••••••••••"
}
```

This component provides essential secure password input functionality for ERP authentication and user management scenarios.