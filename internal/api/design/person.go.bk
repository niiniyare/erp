package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// PERSON & INDIVIDUAL TYPES
// ============================================================================

// Person represents an individual person with complete personal information.
// Serves as the foundation for employees, customers, vendors, and other human entities.
var Person = ResultType("application/vnd.erp.person", func() {
	Description("Individual person with comprehensive personal information, contact details, and identification for human entity management")
	
	Attributes(func() {
		Field(1, "id", String, "Unique person identifier", func() {
			Format(FormatUUID)
			Example("person-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for person record")
		})
		
		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary")
		})
		
		Field(3, "person_number", String, "Human-readable person identifier", func() {
			Pattern("^[A-Z]{2,4}-[0-9]{4,8}$")
			Example("PER-12345678")
			Description("Business-friendly unique identifier")
		})
		
		Field(4, "title", String, "Personal title or honorific", func() {
			Enum("MR", "MS", "MRS", "DR", "PROF", "REV", "HON", "SIR", "LADY", "LORD")
			Example("DR")
			Description("Formal title for addressing and correspondence")
		})
		
		Field(5, "first_name", String, "Given name", func() {
			MinLength(1)
			MaxLength(50)
			Pattern("^[a-zA-Z\\s\\-']{1,50}$")
			Example("John")
			Description("Primary given name")
		})
		
		Field(6, "middle_name", String, "Middle name or initial", func() {
			MaxLength(50)
			Pattern("^[a-zA-Z\\s\\-']{0,50}$")
			Example("Michael")
			Description("Optional middle name")
		})
		
		Field(7, "last_name", String, "Family surname", func() {
			MinLength(1)
			MaxLength(50)
			Pattern("^[a-zA-Z\\s\\-']{1,50}$")
			Example("Doe")
			Description("Family or surname")
		})
		
		Field(8, "suffix", String, "Name suffix", func() {
			Enum("JR", "SR", "II", "III", "IV", "V", "ESQ", "MD", "PHD", "CPA")
			Example("JR")
			Description("Professional or generational suffix")
		})
		
		Field(9, "preferred_name", String, "Preferred or nickname", func() {
			MaxLength(30)
			Pattern("^[a-zA-Z\\s\\-']{0,30}$")
			Example("Johnny")
			Description("Name person prefers to be called")
		})
		
		Field(10, "full_name", String, "Complete formatted name", func() {
			MaxLength(200)
			Example("Dr. John Michael Doe Jr.")
			Description("Full name in display format")
		})
		
		Field(11, "gender", String, "Gender identification", func() {
			Enum("MALE", "FEMALE", "NON_BINARY", "PREFER_NOT_TO_SAY", "OTHER")
			Example("MALE")
			Description("Gender for demographic and legal purposes")
		})
		
		Field(12, "date_of_birth", String, "Birth date", func() {
			Format(FormatDate)
			Example("1985-03-15")
			Description("Date of birth for age verification and demographics")
		})
		
		Field(13, "place_of_birth", String, "Birth location", func() {
			MaxLength(100)
			Example("New York, NY, USA")
			Description("City, state/province, country of birth")
		})
		
		Field(14, "nationality", String, "Nationality or citizenship", func() {
			Pattern("^[A-Z]{2,3}$")
			Example("USA")
			Description("ISO country code for nationality")
		})
		
		Field(15, "marital_status", String, "Marital status", func() {
			Enum("SINGLE", "MARRIED", "DIVORCED", "WIDOWED", "SEPARATED", "DOMESTIC_PARTNERSHIP", "OTHER")
			Example("MARRIED")
			Description("Legal marital status")
		})
		
		Field(16, "primary_language", String, "Primary spoken language", func() {
			Pattern("^[a-z]{2}(-[A-Z]{2})?$")
			Default("en")
			Example("en-US")
			Description("ISO language code for primary language")
		})
		
		Field(17, "languages_spoken", ArrayOf(String), "Additional languages", func() {
			Example([]string{"es-ES", "fr-FR", "de-DE"})
			Description("Other languages person speaks fluently")
		})
		
		Field(18, "primary_email", String, "Primary email address", func() {
			Format(FormatEmail)
			Example("john.doe@email.com")
			Description("Main email for communication")
		})
		
		Field(19, "secondary_email", String, "Secondary email address", func() {
			Format(FormatEmail)
			Example("j.doe.personal@email.com")
			Description("Alternative email contact")
		})
		
		Field(20, "primary_phone", String, "Primary phone number", func() {
			Pattern("^\\+?[1-9]\\d{1,14}$")
			Example("+1-555-123-4567")
			Description("Main phone number with country code")
		})
		
		Field(21, "secondary_phone", String, "Secondary phone number", func() {
			Pattern("^\\+?[1-9]\\d{1,14}$")
			Example("+1-555-987-6543")
			Description("Alternative phone contact")
		})
		
		Field(22, "mobile_phone", String, "Mobile phone number", func() {
			Pattern("^\\+?[1-9]\\d{1,14}$")
			Example("+1-555-456-7890")
			Description("Mobile/cellular phone number")
		})
		
		Field(23, "home_address", Address, "Residential address", func() {
			Description("Primary home address")
		})
		
		Field(24, "mailing_address", Address, "Mailing address", func() {
			Description("Address for postal correspondence")
		})
		
		Field(25, "emergency_contact", Type("EmergencyContact", func() {
			Field(1, "name", String, "Emergency contact name", func() {
				MinLength(2)
				MaxLength(100)
				Example("Jane Doe")
			})
			Field(2, "relationship", String, "Relationship to person", func() {
				Example("Spouse")
			})
			Field(3, "phone", String, "Emergency contact phone", func() {
				Pattern("^\\+?[1-9]\\d{1,14}$")
				Example("+1-555-999-8888")
			})
			Field(4, "email", String, "Emergency contact email", func() {
				Format(FormatEmail)
				Example("jane.doe@email.com")
			})
			Required("name", "relationship", "phone")
		}), "Emergency contact information", func() {
			Description("Primary emergency contact person")
		})
		
		Field(26, "identification", ArrayOf(Type("PersonIdentification", func() {
			Field(1, "type", String, "Identification type", func() {
				Enum("PASSPORT", "DRIVERS_LICENSE", "NATIONAL_ID", "SSN", "TIN", "VISA", "WORK_PERMIT")
				Example("PASSPORT")
			})
			Field(2, "number", String, "Identification number", func() {
				MinLength(3)
				MaxLength(50)
				Example("P123456789")
			})
			Field(3, "issuing_authority", String, "Authority that issued ID", func() {
				MaxLength(100)
				Example("U.S. Department of State")
			})
			Field(4, "issued_date", String, "Issue date", func() {
				Format(FormatDate)
				Example("2020-01-15")
			})
			Field(5, "expiry_date", String, "Expiration date", func() {
				Format(FormatDate)
				Example("2030-01-15")
			})
			Field(6, "issuing_country", String, "Country of issue", func() {
				Pattern("^[A-Z]{2,3}$")
				Example("USA")
			})
			Required("type", "number", "issuing_authority")
		})), "Identification documents", func() {
			Description("Government-issued identification documents")
		})
		
		Field(27, "employment_eligibility", String, "Work authorization status", func() {
			Enum("CITIZEN", "PERMANENT_RESIDENT", "WORK_VISA", "STUDENT_VISA", "NO_AUTHORIZATION", "PENDING")
			Example("CITIZEN")
			Description("Legal authorization to work")
		})
		
		Field(28, "veteran_status", String, "Military veteran status", func() {
			Enum("VETERAN", "ACTIVE_DUTY", "RESERVE", "GUARD", "NOT_VETERAN", "PREFER_NOT_TO_SAY")
			Example("VETERAN")
			Description("Military service status for reporting")
		})
		
		Field(29, "disability_status", String, "Disability status", func() {
			Enum("NO_DISABILITY", "HAS_DISABILITY", "PREFER_NOT_TO_SAY")
			Example("PREFER_NOT_TO_SAY")
			Description("Disability status for compliance reporting")
		})
		
		Field(30, "ethnicity", String, "Ethnic background", func() {
			Enum("HISPANIC_LATINO", "NOT_HISPANIC_LATINO", "PREFER_NOT_TO_SAY")
			Example("NOT_HISPANIC_LATINO")
			Description("Ethnicity for demographic reporting")
		})
		
		Field(31, "race", ArrayOf(String), "Racial identification", func() {
			Example([]string{"WHITE", "BLACK_AFRICAN_AMERICAN"})
			Description("Race categories for demographic reporting")
		})
		
		Field(32, "professional_credentials", ArrayOf(Type("ProfessionalCredential", func() {
			Field(1, "type", String, "Credential type", func() {
				Example("CERTIFICATION")
			})
			Field(2, "name", String, "Credential name", func() {
				Example("Certified Public Accountant")
			})
			Field(3, "issuing_organization", String, "Issuing body", func() {
				Example("American Institute of CPAs")
			})
			Field(4, "credential_number", String, "Certificate/license number", func() {
				Example("CPA-123456")
			})
			Field(5, "issued_date", String, "Issue date", func() {
				Format(FormatDate)
				Example("2020-06-15")
			})
			Field(6, "expiry_date", String, "Expiration date", func() {
				Format(FormatDate)
				Example("2023-06-15")
			})
			Field(7, "is_active", Boolean, "Active status", func() {
				Default(true)
			})
			Required("type", "name", "issuing_organization")
		})), "Professional credentials", func() {
			Description("Licenses, certifications, and professional qualifications")
		})
		
		Field(33, "education", ArrayOf(Type("Education", func() {
			Field(1, "institution_name", String, "Educational institution", func() {
				MaxLength(100)
				Example("University of California, Berkeley")
			})
			Field(2, "degree_type", String, "Type of degree", func() {
				Enum("HIGH_SCHOOL", "ASSOCIATE", "BACHELOR", "MASTER", "DOCTORATE", "CERTIFICATE", "DIPLOMA")
				Example("BACHELOR")
			})
			Field(3, "degree_name", String, "Degree or field of study", func() {
				MaxLength(100)
				Example("Bachelor of Science in Computer Science")
			})
			Field(4, "major", String, "Major field of study", func() {
				MaxLength(50)
				Example("Computer Science")
			})
			Field(5, "minor", String, "Minor field of study", func() {
				MaxLength(50)
				Example("Mathematics")
			})
			Field(6, "graduation_date", String, "Graduation date", func() {
				Format(FormatDate)
				Example("2007-05-15")
			})
			Field(7, "gpa", Float64, "Grade point average", func() {
				Minimum(0.0)
				Maximum(4.0)
				Example(3.75)
			})
			Field(8, "honors", String, "Academic honors", func() {
				Example("Magna Cum Laude")
			})
			Required("institution_name", "degree_type", "degree_name")
		})), "Educational background", func() {
			Description("Academic qualifications and educational history")
		})
		
		Field(34, "social_media", MapOf(String, String), "Social media profiles", func() {
			Example(map[string]any{
				"linkedin": "https://linkedin.com/in/johndoe",
				"twitter":  "@johndoe",
				"github":   "johndoe",
			})
			Description("Professional and personal social media accounts")
		})
		
		Field(35, "preferences", MapOf(String, Any), "Personal preferences", func() {
			Example(map[string]any{
				"communication_method": "email",
				"best_contact_time":   "business_hours",
				"language_preference": "en-US",
				"accessibility_needs": []string{"large_print"},
			})
			Description("Communication and accessibility preferences")
		})
		
		Field(36, "notes", String, "Additional notes", func() {
			MaxLength(2000)
			Example("Prefers to be contacted via email. Available for travel assignments.")
			Description("Miscellaneous notes and observations")
		})
		
		Field(37, "tags", ArrayOf(String), "Classification tags", func() {
			Example([]string{"executive", "remote_worker", "high_performer"})
			Description("Searchable tags for categorization")
		})
		
		Field(38, "privacy_settings", MapOf(String, Boolean), "Privacy preferences", func() {
			Example(map[string]any{
				"allow_directory_listing": true,
				"share_photo":            false,
				"allow_external_contact": true,
				"marketing_emails":       false,
			})
			Description("Privacy and data sharing preferences")
		})
		
		Field(39, "is_active", Boolean, "Active status", func() {
			Default(true)
			Example(true)
			Description("Whether person record is currently active")
		})
		
		Field(40, "status", String, "Person record status", func() {
			Enum("ACTIVE", "INACTIVE", "ARCHIVED", "DECEASED")
			Default("ACTIVE")
			Example("ACTIVE")
			Description("Current status of person record")
		})
		
		// Audit fields from common.go
		AuditFields()
		
		Required("id", "tenant_id", "person_number", "first_name", "last_name", 
			"full_name", "primary_email", "is_active", "status", "created_at")
	})
	
	View("default", func() {
		Description("Standard person view for general listings and references")
		Attribute("id")
		Attribute("person_number")
		Attribute("full_name")
		Attribute("primary_email")
		Attribute("primary_phone")
		Attribute("status")
		Attribute("created_at")
	})
	
	View("detailed", func() {
		Description("Complete person profile with all personal information")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("person_number")
		Attribute("title")
		Attribute("first_name")
		Attribute("middle_name")
		Attribute("last_name")
		Attribute("suffix")
		Attribute("preferred_name")
		Attribute("full_name")
		Attribute("gender")
		Attribute("date_of_birth")
		Attribute("place_of_birth")
		Attribute("nationality")
		Attribute("marital_status")
		Attribute("primary_language")
		Attribute("languages_spoken")
		Attribute("primary_email")
		Attribute("secondary_email")
		Attribute("primary_phone")
		Attribute("secondary_phone")
		Attribute("mobile_phone")
		Attribute("home_address")
		Attribute("mailing_address")
		Attribute("emergency_contact")
		Attribute("identification")
		Attribute("employment_eligibility")
		Attribute("professional_credentials")
		Attribute("education")
		Attribute("social_media")
		Attribute("preferences")
		Attribute("notes")
		Attribute("tags")
		Attribute("privacy_settings")
		Attribute("is_active")
		Attribute("status")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
	
	View("contact", func() {
		Description("Contact information view for communication purposes")
		Attribute("id")
		Attribute("full_name")
		Attribute("preferred_name")
		Attribute("primary_email")
		Attribute("secondary_email")
		Attribute("primary_phone")
		Attribute("mobile_phone")
		Attribute("home_address")
		Attribute("mailing_address")
		Attribute("emergency_contact")
		Attribute("primary_language")
		Attribute("preferences")
	})
	
	View("public", func() {
		Description("Public directory view with limited information")
		Attribute("id")
		Attribute("full_name")
		Attribute("preferred_name")
		Attribute("primary_email")
		Attribute("professional_credentials")
	})
	
	View("summary", func() {
		Description("Minimal view for quick references and lookups")
		Attribute("id")
		Attribute("person_number")
		Attribute("full_name")
		Attribute("primary_email")
		Attribute("status")
	})
})

// PersonRelationship represents relationships between people.
var PersonRelationship = Type("PersonRelationship", func() {
	Description("Relationship connection between two people with context and metadata")
	
	Field(1, "id", String, "Unique relationship identifier", func() {
		Format(FormatUUID)
		Example("rel-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for relationship record")
	})
	
	Field(2, "person_a_id", String, "First person in relationship", func() {
		Format(FormatUUID)
		Example("person-456e7890-e89b-12d3-a456-426614174000")
		Description("Primary person reference")
	})
	
	Field(3, "person_b_id", String, "Second person in relationship", func() {
		Format(FormatUUID)
		Example("person-789abc12-def3-4567-890a-bcdef1234567")
		Description("Related person reference")
	})
	
	Field(4, "relationship_type", String, "Type of relationship", func() {
		Enum("SPOUSE", "PARENT", "CHILD", "SIBLING", "RELATIVE", "FRIEND", 
			"COLLEAGUE", "SUPERVISOR", "SUBORDINATE", "BUSINESS_PARTNER", 
			"EMERGENCY_CONTACT", "REFERENCE", "OTHER")
		Example("SPOUSE")
		Description("Nature of the relationship")
	})
	
	Field(5, "relationship_description", String, "Detailed relationship description", func() {
		MaxLength(200)
		Example("Married since 2010, primary emergency contact")
		Description("Additional context about the relationship")
	})
	
	Field(6, "start_date", String, "Relationship start date", func() {
		Format(FormatDate)
		Example("2010-06-15")
		Description("When relationship began")
	})
	
	Field(7, "end_date", String, "Relationship end date", func() {
		Format(FormatDate)
		Example("2023-12-31")
		Description("When relationship ended (if applicable)")
	})
	
	Field(8, "is_active", Boolean, "Active relationship status", func() {
		Default(true)
		Example(true)
		Description("Whether relationship is currently active")
	})
	
	Field(9, "is_reciprocal", Boolean, "Reciprocal relationship flag", func() {
		Default(true)
		Example(true)
		Description("Whether relationship applies both ways")
	})
	
	Field(10, "privacy_level", String, "Privacy level for relationship", func() {
		Enum("PUBLIC", "INTERNAL", "RESTRICTED", "CONFIDENTIAL")
		Default("INTERNAL")
		Example("INTERNAL")
		Description("Who can view this relationship information")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "person_a_id", "person_b_id", "relationship_type", "is_active", "created_at")
})

// PersonNote represents notes and observations about a person.
var PersonNote = Type("PersonNote", func() {
	Description("Notes, observations, and comments about a person for record keeping")
	
	Field(1, "id", String, "Unique note identifier", func() {
		Format(FormatUUID)
		Example("note-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for note record")
	})
	
	Field(2, "person_id", String, "Associated person identifier", func() {
		Format(FormatUUID)
		Example("person-456e7890-e89b-12d3-a456-426614174000")
		Description("Person this note refers to")
	})
	
	Field(3, "note_type", String, "Type of note", func() {
		Enum("GENERAL", "PERSONAL", "PROFESSIONAL", "MEDICAL", "LEGAL", 
			"PERFORMANCE", "BEHAVIOR", "ACHIEVEMENT", "CONCERN", "FOLLOW_UP")
		Example("PERFORMANCE")
		Description("Category of note for organization")
	})
	
	Field(4, "title", String, "Note title or subject", func() {
		MinLength(1)
		MaxLength(100)
		Example("Quarterly Performance Review")
		Description("Brief title for the note")
	})
	
	Field(5, "content", String, "Note content", func() {
		MinLength(1)
		MaxLength(5000)
		Example("Excellent performance this quarter. Exceeded all targets and demonstrated strong leadership skills.")
		Description("Detailed note content")
	})
	
	Field(6, "priority", String, "Note priority level", func() {
		Enum("LOW", "NORMAL", "HIGH", "URGENT")
		Default("NORMAL")
		Example("HIGH")
		Description("Priority for follow-up or attention")
	})
	
	Field(7, "is_confidential", Boolean, "Confidential note flag", func() {
		Default(false)
		Example(true)
		Description("Whether note contains sensitive information")
	})
	
	Field(8, "visibility", String, "Note visibility scope", func() {
		Enum("PUBLIC", "TEAM", "MANAGEMENT", "HR_ONLY", "ADMIN_ONLY")
		Default("TEAM")
		Example("MANAGEMENT")
		Description("Who can view this note")
	})
	
	Field(9, "tags", ArrayOf(String), "Note classification tags", func() {
		Example([]string{"performance", "leadership", "quarterly_review"})
		Description("Searchable tags for note categorization")
	})
	
	Field(10, "follow_up_required", Boolean, "Follow-up action needed", func() {
		Default(false)
		Example(true)
		Description("Whether note requires follow-up action")
	})
	
	Field(11, "follow_up_date", String, "Follow-up due date", func() {
		Format(FormatDate)
		Example("2024-01-15")
		Description("When follow-up action is due")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "person_id", "note_type", "title", "content", "created_at", "created_by")
})