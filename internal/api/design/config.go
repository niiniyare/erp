package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// SYSTEM CONFIGURATION & SETTINGS TYPES
// ============================================================================

// ConfigChangeRecord represents a single configuration change record
var ConfigChangeRecord = Type("ConfigChangeRecord", func() {
	Description("Single configuration change record")

	Field(1, "change_id", String, "Change identifier", func() {
		Format(FormatUUID)
	})
	Field(2, "previous_value", Any, "Previous configuration value", func() {
		Example("old-smtp.mailserver.com")
	})
	Field(3, "new_value", Any, "New configuration value", func() {
		Example("new-smtp.mailserver.com")
	})
	Field(4, "change_reason", String, "Reason for change", func() {
		MaxLength(300)
		Example("Migrating to new SMTP provider for better reliability")
	})
	Field(5, "changed_by", String, "User who made change", func() {
		Format(FormatUUID)
	})
	Field(6, "changed_at", String, "Change timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T14:30:00Z")
	})
	Field(7, "approval_required", Boolean, "Required approval", func() {
		Default(false)
	})
	Field(8, "approved_by", String, "Approver user ID", func() {
		Format(FormatUUID)
	})
	Field(9, "approved_at", String, "Approval timestamp", func() {
		Format(FormatDateTime)
	})
	Required("change_id", "new_value", "changed_by", "changed_at")
})

// TemplateConfig represents a single configuration in a template
var TemplateConfig = Type("TemplateConfig", func() {
	Description("Single configuration definition in a template")

	Field(1, "key", String, "Configuration key", func() {
		Pattern("^[a-z][a-z0-9_.]{2,199}$")
		Example("email.smtp.server_host")
	})
	Field(2, "name", String, "Configuration name", func() {
		Example("SMTP Server Host")
	})
	Field(3, "description", String, "Configuration description", func() {
		Example("SMTP server hostname for outbound email")
	})
	Field(4, "value_type", String, "Data type", func() {
		Enum("STRING", "INTEGER", "FLOAT", "BOOLEAN", "JSON", "ARRAY",
			"URL", "EMAIL", "PASSWORD", "FILE_PATH", "DATE", "DURATION")
		Example("STRING")
	})
	Field(5, "default_value", Any, "Default value", func() {
		Example("smtp.mailserver.com")
	})
	Field(6, "is_required", Boolean, "Required flag", func() {
		Default(false)
	})
	Field(7, "is_sensitive", Boolean, "Sensitive data flag", func() {
		Default(false)
	})
	Field(8, "validation_rules", MapOf(String, Any), "Validation constraints", func() {
		Example(map[string]any{
			"pattern":    "^[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			"min_length": 3,
			"max_length": 100,
		})
	})
	Field(9, "order", UInt, "Display order", func() {
		Example(1)
	})
	Required("key", "name", "value_type", "is_required", "is_sensitive", "order")
})

// TemplateVariable represents a template variable
var TemplateVariable = Type("TemplateVariable", func() {
	Description("Template variable definition")

	Field(1, "name", String, "Variable name", func() {
		Pattern("^[a-zA-Z][a-zA-Z0-9_]{1,49}$")
		Example("COMPANY_NAME")
	})
	Field(2, "description", String, "Variable description", func() {
		Example("Company display name")
	})
	Field(3, "data_type", String, "Variable data type", func() {
		Enum("STRING", "INTEGER", "FLOAT", "BOOLEAN", "DATE", "URL", "EMAIL")
		Example("STRING")
	})
	Field(4, "default_value", Any, "Default variable value", func() {
		Example("Awo Enterprise")
	})
	Field(5, "is_required", Boolean, "Required variable flag", func() {
		Default(true)
	})
	Field(6, "validation_pattern", String, "Validation regex pattern", func() {
		Example("^[A-Za-z0-9\\s]{2,100}$")
	})
	Required("name", "description", "data_type", "is_required")
})

// ValidationMessage represents a validation error message
var ValidationMessage = Type("ValidationMessage", func() {
	Description("Configuration validation error message")

	Field(1, "field_name", String, "Field with validation error", func() {
		Example("email.smtp.port")
	})
	Field(2, "error_code", String, "Error classification code", func() {
		Enum("REQUIRED", "FORMAT", "RANGE", "DEPENDENCY", "SECURITY", "CUSTOM")
		Example("FORMAT")
	})
	Field(3, "error_message", String, "Human-readable error message", func() {
		Example("Port must be between 1 and 65535")
	})
	Field(4, "current_value", Any, "Current invalid value", func() {
		Example("99999")
	})
	Field(5, "expected_format", String, "Expected value format", func() {
		Example("Integer between 1-65535")
	})
	Field(6, "severity", String, "Error severity level", func() {
		Enum("ERROR", "WARNING", "INFO")
		Example("ERROR")
	})
	Required("field_name", "error_code", "error_message", "severity")
})

// ConfigDependency represents a configuration dependency
var ConfigDependency = Type("ConfigDependency", func() {
	Description("Configuration dependency relationship")

	Field(1, "dependency_type", String, "Type of dependency", func() {
		Enum("REQUIRES", "CONFLICTS", "IMPLIES", "EXCLUDES")
		Example("REQUIRES")
	})
	Field(2, "source_key", String, "Source configuration key", func() {
		Example("email.enabled")
	})
	Field(3, "target_key", String, "Target configuration key", func() {
		Example("email.smtp.server_host")
	})
	Field(4, "condition", String, "Dependency condition", func() {
		Example("email.enabled == true")
	})
	Field(5, "error_message", String, "Error message when dependency violated", func() {
		Example("SMTP server host is required when email is enabled")
	})
	Required("dependency_type", "source_key", "target_key", "condition")
})

// SystemConfig represents system-wide configuration settings.
// Provides centralized configuration management with validation and audit trails.
var SystemConfig = ResultType("application/vnd.erp.config", func() {
	Description("System-wide configuration settings with validation, encryption, and audit trail support")

	Attributes(func() {
		Field(1, "id", String, "Unique configuration identifier", func() {
			Format(FormatUUID)
			Example("config-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for configuration record")
		})

		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary (null for global configs)")
		})

		Field(3, "key", String, "Configuration key", func() {
			Pattern("^[a-z][a-z0-9_.]{2,199}$")
			MinLength(3)
			MaxLength(200)
			Example("email.smtp.server_host")
			Description("Hierarchical configuration key using dot notation")
		})

		Field(4, "name", String, "Human-readable configuration name", func() {
			MinLength(3)
			MaxLength(150)
			Example("SMTP Server Host")
			Description("User-friendly name for UI display")
		})

		Field(5, "description", String, "Configuration description", func() {
			MaxLength(500)
			Example("SMTP server hostname for outbound email delivery")
			Description("Detailed explanation of configuration purpose")
		})

		Field(6, "category", String, "Configuration category", func() {
			Enum("SYSTEM", "EMAIL", "DATABASE", "SECURITY", "INTEGRATION", "UI",
				"REPORTING", "WORKFLOW", "COMPLIANCE", "PERFORMANCE", "MONITORING")
			Example("EMAIL")
			Description("Functional grouping for organization")
		})

		Field(7, "value_type", String, "Data type of configuration value", func() {
			Enum("STRING", "INTEGER", "FLOAT", "BOOLEAN", "JSON", "ARRAY", "URL",
				"EMAIL", "PASSWORD", "FILE_PATH", "DATE", "DURATION")
			Example("STRING")
			Description("Expected data type for validation")
		})

		Field(8, "value", Any, "Configuration value", func() {
			Example("smtp.mailserver.com")
			Description("Actual configuration value (encrypted if sensitive)")
		})

		Field(9, "default_value", Any, "Default configuration value", func() {
			Example("localhost")
			Description("Default value when not explicitly set")
		})

		Field(10, "is_encrypted", Boolean, "Encryption status", func() {
			Default(false)
			Example(true)
			Description("Whether value is stored encrypted")
		})

		Field(11, "is_sensitive", Boolean, "Sensitive data flag", func() {
			Default(false)
			Example(true)
			Description("Whether value contains sensitive information")
		})

		Field(12, "is_required", Boolean, "Required configuration flag", func() {
			Default(false)
			Example(true)
			Description("Whether configuration must have a value")
		})

		Field(13, "is_system_managed", Boolean, "System-managed flag", func() {
			Default(false)
			Example(false)
			Description("Whether value is managed by system automation")
		})

		Field(14, "is_user_editable", Boolean, "User-editable flag", func() {
			Default(true)
			Example(false)
			Description("Whether users can modify this configuration")
		})

		Field(15, "validation_rules", MapOf(String, Any), "Validation constraints", func() {
			Example(map[string]any{
				"pattern":        "^[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
				"min_length":     3,
				"max_length":     100,
				"allowed_values": []string{"smtp.gmail.com", "smtp.outlook.com", "smtp.mailserver.com"},
			})
			Description("Rules for validating configuration values")
		})

		Field(16, "environment_specific", Boolean, "Environment-specific flag", func() {
			Default(false)
			Example(true)
			Description("Whether value differs across environments")
		})

		Field(17, "environment_values", MapOf(String, Any), "Per-environment values", func() {
			Example(map[string]any{
				"development": "smtp-dev.mailserver.com",
				"staging":     "smtp-staging.mailserver.com",
				"production":  "smtp.mailserver.com",
			})
			Description("Different values for different environments")
		})

		Field(18, "scope", String, "Configuration scope", func() {
			Enum("GLOBAL", "TENANT", "ENTITY", "USER", "APPLICATION", "MODULE")
			Default("GLOBAL")
			Example("TENANT")
			Description("Level at which configuration applies")
		})

		Field(19, "restart_required", Boolean, "Restart requirement flag", func() {
			Default(false)
			Example(true)
			Description("Whether application restart is needed for changes")
		})

		Field(20, "feature_flag_key", String, "Associated feature flag", func() {
			Pattern("^[a-z][a-z0-9_]{2,99}$")
			Example("enable_email_notifications")
			Description("Feature flag that controls this configuration")
		})

		Field(21, "dependencies", ArrayOf(String), "Dependent configuration keys", func() {
			Example([]string{"email.smtp.port", "email.smtp.username", "email.smtp.password"})
			Description("Other configurations that depend on this one")
		})

		Field(22, "tags", ArrayOf(String), "Configuration classification tags", func() {
			Example([]string{"email", "smtp", "communication", "critical"})
			Description("Searchable tags for organization")
		})

		Field(23, "version", String, "Configuration version", func() {
			Pattern("^\\d+\\.\\d+\\.\\d+$")
			Default("1.0.0")
			Example("2.1.0")
			Description("Version for tracking configuration changes")
		})

		Field(24, "change_history", ArrayOf(ConfigChangeRecord), "Configuration change history", func() {
			Description("Audit trail of configuration modifications")
		})

		Field(25, "effective_from", String, "Effective start date", func() {
			Format(FormatDateTime)
			Example("2023-12-01T00:00:00Z")
			Description("When configuration becomes active")
		})

		Field(26, "expires_at", String, "Configuration expiration", func() {
			Format(FormatDateTime)
			Example("2024-12-01T00:00:00Z")
			Description("When configuration expires (if applicable)")
		})

		Field(27, "last_validated_at", String, "Last validation timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T10:00:00Z")
			Description("When configuration was last validated")
		})

		Field(28, "validation_status", String, "Current validation status", func() {
			Enum("VALID", "INVALID", "WARNING", "UNKNOWN", "PENDING")
			Default("UNKNOWN")
			Example("VALID")
			Description("Result of last validation check")
		})

		Field(29, "validation_errors", ArrayOf(String), "Validation error messages", func() {
			Example([]string{"SMTP server unreachable", "Invalid port number"})
			Description("Current validation errors (if any)")
		})

		Field(30, "usage_count", UInt, "Configuration usage count", func() {
			Example(1247)
			Description("Number of times configuration has been accessed")
		})

		Field(31, "last_accessed_at", String, "Last access timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T14:25:00Z")
			Description("When configuration was last read")
		})

		Field(32, "documentation_url", String, "Configuration documentation", func() {
			Format(FormatURI)
			Example("https://docs.company.com/config/email-smtp")
			Description("Link to configuration documentation")
		})

		Field(33, "owner_team", String, "Team responsible for configuration", func() {
			Example("platform-team")
			Description("Team that owns this configuration")
		})

		Field(34, "contact_email", String, "Contact for configuration issues", func() {
			Format(FormatEmail)
			Example("platform-team@company.com")
			Description("Email for configuration-related questions")
		})

		Field(35, "is_active", Boolean, "Active configuration status", func() {
			Default(true)
			Example(true)
			Description("Whether configuration is currently active")
		})

		// Audit fields from common.go
		AuditFields()

		Required("id", "key", "name", "category", "value_type", "scope",
			"is_encrypted", "is_sensitive", "is_required", "is_active", "created_at")
	})

	View("default", func() {
		Description("Standard configuration view for management interfaces")
		Attribute("id")
		Attribute("key")
		Attribute("name")
		Attribute("category")
		Attribute("value_type")
		Attribute("scope")
		Attribute("is_required")
		Attribute("is_sensitive")
		Attribute("validation_status")
		Attribute("is_active")
		Attribute("created_at")
	})

	View("detailed", func() {
		Description("Complete configuration view with all settings and history")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("key")
		Attribute("name")
		Attribute("description")
		Attribute("category")
		Attribute("value_type")
		Attribute("value")
		Attribute("default_value")
		Attribute("is_encrypted")
		Attribute("is_sensitive")
		Attribute("is_required")
		Attribute("is_system_managed")
		Attribute("is_user_editable")
		Attribute("validation_rules")
		Attribute("environment_specific")
		Attribute("environment_values")
		Attribute("scope")
		Attribute("restart_required")
		Attribute("feature_flag_key")
		Attribute("dependencies")
		Attribute("tags")
		Attribute("version")
		Attribute("change_history")
		Attribute("effective_from")
		Attribute("expires_at")
		Attribute("validation_status")
		Attribute("validation_errors")
		Attribute("usage_count")
		Attribute("last_accessed_at")
		Attribute("documentation_url")
		Attribute("owner_team")
		Attribute("contact_email")
		Attribute("is_active")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})

	View("public", func() {
		Description("Public configuration view without sensitive data")
		Attribute("id")
		Attribute("key")
		Attribute("name")
		Attribute("description")
		Attribute("category")
		Attribute("value_type")
		Attribute("default_value")
		Attribute("validation_rules")
		Attribute("scope")
		Attribute("documentation_url")
		Attribute("is_active")
	})

	View("runtime", func() {
		Description("Runtime configuration view for application consumption")
		Attribute("key")
		Attribute("value_type")
		Attribute("value")
		Attribute("environment_specific")
		Attribute("environment_values")
		Attribute("effective_from")
		Attribute("expires_at")
		Attribute("is_active")
	})

	View("audit", func() {
		Description("Audit view for compliance and change tracking")
		Attribute("id")
		Attribute("key")
		Attribute("name")
		Attribute("category")
		Attribute("is_sensitive")
		Attribute("change_history")
		Attribute("validation_status")
		Attribute("last_validated_at")
		Attribute("owner_team")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
})

// ConfigTemplate represents reusable configuration templates.
var ConfigTemplate = Type("ConfigTemplate", func() {
	Description("Reusable configuration template for standardizing common settings")

	Field(1, "id", String, "Unique template identifier", func() {
		Format(FormatUUID)
		Example("template-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for template")
	})

	Field(2, "name", String, "Template name", func() {
		Pattern("^[a-z][a-z0-9_]{2,99}$")
		MinLength(3)
		MaxLength(100)
		Example("email_smtp_template")
		Description("Machine-readable template identifier")
	})

	Field(3, "display_name", String, "Human-readable template name", func() {
		MinLength(3)
		MaxLength(150)
		Example("Email SMTP Configuration Template")
		Description("User-friendly name for UI display")
	})

	Field(4, "description", String, "Template description", func() {
		MaxLength(500)
		Example("Standard SMTP configuration for email delivery services")
		Description("Explanation of template purpose and usage")
	})

	Field(5, "category", String, "Template category", func() {
		Enum("SYSTEM", "EMAIL", "DATABASE", "SECURITY", "INTEGRATION", "UI",
			"REPORTING", "WORKFLOW", "COMPLIANCE", "PERFORMANCE", "MONITORING")
		Example("EMAIL")
		Description("Functional classification")
	})

	Field(6, "configurations", ArrayOf(TemplateConfig), "Template configuration definitions", func() {
		Description("Set of related configurations in this template")
	})

	Field(7, "variables", ArrayOf(TemplateVariable), "Template variables for customization", func() {
		Description("Variables that can be substituted when applying template")
	})

	Field(8, "version", String, "Template version", func() {
		Pattern("^\\d+\\.\\d+\\.\\d+$")
		Default("1.0.0")
		Example("1.2.0")
		Description("Version for template changes")
	})

	Field(9, "is_active", Boolean, "Template active status", func() {
		Default(true)
		Example(true)
		Description("Whether template is available for use")
	})

	Field(10, "usage_count", UInt, "Template usage count", func() {
		Example(45)
		Description("Number of times template has been applied")
	})

	// Audit fields
	AuditFields()

	Required("id", "name", "display_name", "category", "configurations",
		"version", "is_active", "created_at")
})

// ConfigValidation represents configuration validation results.
var ConfigValidation = Type("ConfigValidation", func() {
	Description("Result of configuration validation with detailed feedback")

	Field(1, "config_id", String, "Configuration identifier", func() {
		Format(FormatUUID)
		Example("config-123e4567-e89b-12d3-a456-426614174000")
		Description("Configuration that was validated")
	})

	Field(2, "config_key", String, "Configuration key", func() {
		Pattern("^[a-z][a-z0-9_.]{2,199}$")
		Example("email.smtp.server_host")
		Description("Key of validated configuration")
	})

	Field(3, "validation_type", String, "Type of validation performed", func() {
		Enum("SYNTAX", "CONNECTIVITY", "PERMISSIONS", "BUSINESS_RULE", "DEPENDENCY")
		Example("CONNECTIVITY")
		Description("Category of validation check")
	})

	Field(4, "status", String, "Validation result status", func() {
		Enum("VALID", "INVALID", "WARNING", "ERROR", "TIMEOUT")
		Example("VALID")
		Description("Overall validation result")
	})

	Field(5, "messages", ArrayOf(ValidationMessage), "Validation messages and errors", func() {
		Description("Detailed validation feedback")
	})

	Field(6, "validated_at", String, "Validation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T14:30:00Z")
		Description("When validation was performed")
	})

	Field(7, "validation_duration_ms", UInt, "Validation processing time", func() {
		Example(1250)
		Description("Time taken to complete validation")
	})

	Field(8, "validator_version", String, "Validator version", func() {
		Pattern("^v\\d+\\.\\d+\\.\\d+$")
		Example("v1.2.3")
		Description("Version of validation engine used")
	})

	Required("config_id", "config_key", "validation_type", "status",
		"messages", "validated_at")
})

// ConfigDefinition represents the schema and metadata for configuration keys.
var ConfigDefinition = Type("ConfigDefinition", func() {
	Description("Schema definition and metadata for configuration keys to ensure consistency and validation")

	Field(1, "id", String, "Unique definition identifier", func() {
		Format(FormatUUID)
		Example("def-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for configuration definition")
	})

	Field(2, "key", String, "Configuration key pattern", func() {
		Pattern("^[a-z][a-z0-9_.]{2,199}$")
		MinLength(3)
		MaxLength(200)
		Example("email.smtp.server_host")
		Description("Configuration key this definition describes")
	})

	Field(3, "name", String, "Human-readable name", func() {
		MinLength(3)
		MaxLength(150)
		Example("SMTP Server Host")
		Description("User-friendly name for UI display")
	})

	Field(4, "description", String, "Detailed description", func() {
		MinLength(10)
		MaxLength(1000)
		Example("SMTP server hostname for outbound email delivery. Must be a valid FQDN that the application can reach on the configured port.")
		Description("Comprehensive explanation of configuration purpose and usage")
	})

	Field(5, "category", String, "Configuration category", func() {
		Enum("SYSTEM", "EMAIL", "DATABASE", "SECURITY", "INTEGRATION", "UI",
			"REPORTING", "WORKFLOW", "COMPLIANCE", "PERFORMANCE", "MONITORING")
		Example("EMAIL")
		Description("Functional grouping for organization")
	})

	Field(6, "subcategory", String, "Configuration subcategory", func() {
		MaxLength(50)
		Example("SMTP")
		Description("More specific categorization within category")
	})

	Field(7, "value_type", String, "Expected data type", func() {
		Enum("STRING", "INTEGER", "FLOAT", "BOOLEAN", "JSON", "ARRAY", "URL",
			"EMAIL", "PASSWORD", "FILE_PATH", "DATE", "DURATION", "REGEX", "BASE64")
		Example("STRING")
		Description("Data type for validation and UI rendering")
	})

	Field(8, "default_value", Any, "Default value", func() {
		Example("localhost")
		Description("Default value when configuration is not set")
	})

	Field(9, "example_values", ArrayOf(Any), "Example valid values", func() {
		Example([]any{"smtp.gmail.com", "smtp.outlook.com", "mail.company.com"})
		Description("Sample values to help users understand expected format")
	})

	Field(10, "validation_schema", MapOf(String, Any), "Validation rules", func() {
		Example(map[string]any{
			"pattern":           "^[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			"min_length":        3,
			"max_length":        100,
			"allowed_values":    []string{"smtp.gmail.com", "smtp.outlook.com"},
			"forbidden_values":  []string{"localhost", "127.0.0.1"},
			"custom_validation": "hostname_reachable",
		})
		Description("JSON schema-style validation rules")
	})

	Field(11, "is_required", Boolean, "Required configuration flag", func() {
		Default(false)
		Example(true)
		Description("Whether this configuration must have a value")
	})

	Field(12, "is_sensitive", Boolean, "Sensitive data indicator", func() {
		Default(false)
		Example(true)
		Description("Whether value contains sensitive information")
	})

	Field(13, "requires_encryption", Boolean, "Encryption requirement", func() {
		Default(false)
		Example(true)
		Description("Whether value must be stored encrypted")
	})

	Field(14, "scope", String, "Configuration scope", func() {
		Enum("GLOBAL", "TENANT", "ENTITY", "USER", "APPLICATION", "MODULE")
		Default("GLOBAL")
		Example("TENANT")
		Description("Level at which configuration can be set")
	})

	Field(15, "environment_specific", Boolean, "Environment-specific flag", func() {
		Default(false)
		Example(true)
		Description("Whether value typically differs across environments")
	})

	Field(16, "restart_required", Boolean, "Restart requirement", func() {
		Default(false)
		Example(true)
		Description("Whether application restart is needed for changes")
	})

	Field(17, "hot_reloadable", Boolean, "Hot reload capability", func() {
		Default(false)
		Example(true)
		Description("Whether changes can be applied without restart")
	})

	Field(18, "ui_widget", String, "Recommended UI widget", func() {
		Enum("TEXT_INPUT", "PASSWORD_INPUT", "TEXTAREA", "SELECT", "MULTISELECT",
			"CHECKBOX", "RADIO", "FILE_UPLOAD", "DATE_PICKER", "COLOR_PICKER",
			"SLIDER", "JSON_EDITOR", "CODE_EDITOR")
		Default("TEXT_INPUT")
		Example("SELECT")
		Description("Suggested UI component for editing this configuration")
	})

	Field(19, "ui_options", MapOf(String, Any), "UI widget options", func() {
		Example(map[string]any{
			"placeholder": "Enter SMTP server hostname",
			"help_text":   "This should be the FQDN of your SMTP server",
			"options": []map[string]any{
				{"label": "Gmail SMTP", "value": "smtp.gmail.com"},
				{"label": "Outlook SMTP", "value": "smtp.outlook.com"},
			},
		})
		Description("Configuration options for UI widget")
	})

	Field(20, "dependencies", ArrayOf(ConfigDependency), "Configuration dependencies", func() {
		Description("Other configurations this one depends on or affects")
	})

	Field(21, "feature_flags", ArrayOf(String), "Related feature flags", func() {
		Example([]string{"enable_email_notifications", "smtp_authentication"})
		Description("Feature flags that control availability of this configuration")
	})

	Field(22, "compliance_requirements", ArrayOf(String), "Compliance frameworks", func() {
		Example([]string{"SOX", "GDPR", "HIPAA"})
		Description("Regulatory compliance requirements affecting this configuration")
	})

	Field(23, "security_impact", String, "Security impact level", func() {
		Enum("NONE", "LOW", "MEDIUM", "HIGH", "CRITICAL")
		Default("NONE")
		Example("MEDIUM")
		Description("Security impact of changing this configuration")
	})

	Field(24, "change_risk", String, "Change risk assessment", func() {
		Enum("NONE", "LOW", "MEDIUM", "HIGH", "CRITICAL")
		Default("LOW")
		Example("HIGH")
		Description("Risk level associated with modifying this configuration")
	})

	Field(25, "testing_notes", String, "Testing guidance", func() {
		MaxLength(500)
		Example("Test email delivery after changing SMTP settings. Verify both internal and external email delivery.")
		Description("Guidance for testing configuration changes")
	})

	Field(26, "rollback_strategy", String, "Rollback approach", func() {
		MaxLength(300)
		Example("Revert to previous SMTP server. Keep backup configuration values for quick restoration.")
		Description("How to rollback if configuration change causes issues")
	})

	Field(27, "monitoring_metrics", ArrayOf(String), "Related metrics to monitor", func() {
		Example([]string{"email_delivery_rate", "smtp_connection_errors", "email_queue_length"})
		Description("Metrics to watch when this configuration changes")
	})

	Field(28, "documentation_url", String, "External documentation", func() {
		Format(FormatURI)
		Example("https://docs.company.com/config/email-smtp-host")
		Description("Link to detailed configuration documentation")
	})

	Field(29, "owner_team", String, "Responsible team", func() {
		Pattern("^[a-z][a-z0-9_-]{2,49}$")
		Example("platform-team")
		Description("Team that owns and maintains this configuration")
	})

	Field(30, "contact_email", String, "Support contact", func() {
		Format(FormatEmail)
		Example("platform-team@company.com")
		Description("Email for configuration-related questions")
	})

	Field(31, "version", String, "Definition version", func() {
		Pattern("^\\d+\\.\\d+\\.\\d+$")
		Default("1.0.0")
		Example("2.1.0")
		Description("Version for tracking definition changes")
	})

	Field(32, "tags", ArrayOf(String), "Classification tags", func() {
		Example([]string{"email", "smtp", "communication", "external_service"})
		Description("Searchable tags for organization and discovery")
	})

	Field(33, "deprecated", Boolean, "Deprecated status", func() {
		Default(false)
		Example(false)
		Description("Whether this configuration is deprecated")
	})

	Field(34, "deprecation_message", String, "Deprecation notice", func() {
		MaxLength(300)
		Example("This configuration is deprecated. Use email.smtp.server_hostname instead.")
		Description("Message explaining deprecation and migration path")
	})

	Field(35, "removal_version", String, "Planned removal version", func() {
		Pattern("^\\d+\\.\\d+\\.\\d+$")
		Example("3.0.0")
		Description("Version when deprecated configuration will be removed")
	})

	Field(36, "is_active", Boolean, "Active definition status", func() {
		Default(true)
		Example(true)
		Description("Whether this definition is currently active")
	})

	// Audit fields
	AuditFields()

	Required("id", "key", "name", "description", "category", "value_type",
		"is_required", "is_sensitive", "scope", "is_active", "created_at")
})
