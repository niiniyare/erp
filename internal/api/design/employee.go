package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// EMPLOYEE & HUMAN RESOURCES TYPES
// ============================================================================

// Employee represents an organization employee with comprehensive employment information.
// Links to person record for personal details and provides employment-specific data.
var Employee = ResultType("application/vnd.erp.employee", func() {
	Description("Organization employee with employment details, compensation, performance tracking, and HR management information")
	
	Attributes(func() {
		Field(1, "id", String, "Unique employee identifier", func() {
			Format(FormatUUID)
			Example("emp-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for employee record")
		})
		
		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary")
		})
		
		Field(3, "person_id", String, "Associated person record", func() {
			Format(FormatUUID)
			Example("person-456e7890-e89b-12d3-a456-426614174000")
			Description("Links to personal information")
		})
		
		Field(4, "employee_number", String, "Human-readable employee identifier", func() {
			Pattern("^[A-Z0-9]{3,15}$")
			Example("EMP-12345")
			Description("Business-friendly unique employee ID")
		})
		
		Field(5, "badge_number", String, "Physical badge or card number", func() {
			Pattern("^[A-Z0-9]{3,20}$")
			Example("BADGE-67890")
			Description("ID card or security badge number")
		})
		
		Field(6, "entity_id", String, "Primary organizational entity", func() {
			Format(FormatUUID)
			Example("entity-789abc12-def3-4567-890a-bcdef1234567")
			Description("Department or organizational unit")
		})
		
		Field(7, "manager_id", String, "Direct manager employee ID", func() {
			Format(FormatUUID)
			Example("emp-manager-123e4567-e89b-12d3-a456-426614174000")
			Description("Immediate supervisor reference")
		})
		
		Field(8, "job_title", String, "Current job title", func() {
			MinLength(2)
			MaxLength(100)
			Example("Senior Financial Analyst")
			Description("Official position title")
		})
		
		Field(9, "job_level", String, "Employment level classification", func() {
			Enum("INTERN", "ENTRY", "JUNIOR", "INTERMEDIATE", "SENIOR", "LEAD", 
				"PRINCIPAL", "MANAGER", "DIRECTOR", "VP", "SVP", "EVP", "EXECUTIVE")
			Example("SENIOR")
			Description("Hierarchical level for compensation and reporting")
		})
		
		Field(10, "job_family", String, "Job function category", func() {
			Enum("ENGINEERING", "FINANCE", "SALES", "MARKETING", "HR", "OPERATIONS", 
				"LEGAL", "CUSTOMER_SUCCESS", "PRODUCT", "DESIGN", "EXECUTIVE", "ADMIN")
			Example("FINANCE")
			Description("Functional area for career development")
		})
		
		Field(11, "employment_type", String, "Type of employment", func() {
			Enum("FULL_TIME", "PART_TIME", "CONTRACT", "TEMPORARY", "INTERN", 
				"CONSULTANT", "SEASONAL", "ON_CALL")
			Default("FULL_TIME")
			Example("FULL_TIME")
			Description("Employment classification")
		})
		
		Field(12, "employment_status", String, "Current employment status", func() {
			Enum("ACTIVE", "ON_LEAVE", "SUSPENDED", "TERMINATED", "RETIRED", "DECEASED")
			Default("ACTIVE")
			Example("ACTIVE")
			Description("Current status in the organization")
		})
		
		Field(13, "hire_date", String, "Date of hire", func() {
			Format(FormatDate)
			Example("2023-01-15")
			Description("Start date with the organization")
		})
		
		Field(14, "start_date", String, "Current position start date", func() {
			Format(FormatDate)
			Example("2023-06-01")
			Description("Start date for current role")
		})
		
		Field(15, "end_date", String, "Employment end date", func() {
			Format(FormatDate)
			Example("2024-12-31")
			Description("End date if employment has terminated")
		})
		
		Field(16, "probation_end_date", String, "Probation period end date", func() {
			Format(FormatDate)
			Example("2023-04-15")
			Description("When probationary period ends")
		})
		
		Field(17, "work_location", String, "Primary work location", func() {
			Enum("OFFICE", "REMOTE", "HYBRID", "FIELD", "CUSTOMER_SITE", "TRAVELING")
			Default("OFFICE")
			Example("HYBRID")
			Description("Where employee primarily works")
		})
		
		Field(18, "office_location", String, "Office assignment", func() {
			MaxLength(100)
			Example("New York - 42nd Street Office")
			Description("Specific office or building location")
		})
		
		Field(19, "work_schedule", String, "Work schedule type", func() {
			Enum("STANDARD", "FLEXIBLE", "COMPRESSED", "SHIFT", "ROTATING", "ON_CALL")
			Default("STANDARD")
			Example("FLEXIBLE")
			Description("Type of work schedule arrangement")
		})
		
		Field(20, "weekly_hours", Float64, "Expected weekly work hours", func() {
			Minimum(1.0)
			Maximum(80.0)
			Default(40.0)
			Example(40.0)
			Description("Standard weekly hours commitment")
		})
		
		Field(21, "compensation", Type("EmployeeCompensation", func() {
			Field(1, "salary_amount", Float64, "Annual salary amount", func() {
				Minimum(0.0)
				Example(75000.00)
			})
			Field(2, "salary_currency", String, "Salary currency", func() {
				Pattern("^[A-Z]{3}$")
				Default("USD")
				Example("USD")
			})
			Field(3, "pay_frequency", String, "Payment frequency", func() {
				Enum("WEEKLY", "BI_WEEKLY", "SEMI_MONTHLY", "MONTHLY", "ANNUALLY")
				Default("BI_WEEKLY")
				Example("BI_WEEKLY")
			})
			Field(4, "hourly_rate", Float64, "Hourly rate for hourly employees", func() {
				Minimum(0.0)
				Example(35.50)
			})
			Field(5, "overtime_eligible", Boolean, "Eligible for overtime pay", func() {
				Default(false)
				Example(true)
			})
			Field(6, "commission_eligible", Boolean, "Eligible for commission", func() {
				Default(false)
				Example(false)
			})
			Field(7, "bonus_eligible", Boolean, "Eligible for bonuses", func() {
				Default(true)
				Example(true)
			})
			Field(8, "equity_eligible", Boolean, "Eligible for equity compensation", func() {
				Default(false)
				Example(true)
			})
			Field(9, "last_review_date", String, "Last compensation review", func() {
				Format(FormatDate)
				Example("2023-07-01")
			})
			Field(10, "next_review_date", String, "Next scheduled review", func() {
				Format(FormatDate)
				Example("2024-07-01")
			})
			Required("salary_currency", "pay_frequency")
		}), "Compensation details", func() {
			Description("Salary, benefits, and compensation structure")
		})
		
		Field(22, "benefits", MapOf(String, Any), "Employee benefits enrollment", func() {
			Example(map[string]any{
				"health_insurance":    true,
				"dental_insurance":    true,
				"vision_insurance":    false,
				"life_insurance":      true,
				"disability_insurance": true,
				"retirement_plan":     "401k",
				"pto_accrual_rate":    "20_days_annual",
				"sick_leave":          "10_days_annual",
			})
			Description("Benefits enrollment and entitlements")
		})
		
		Field(23, "skills", ArrayOf(Type("EmployeeSkill", func() {
			Field(1, "skill_name", String, "Name of skill", func() {
				MinLength(2)
				MaxLength(50)
				Example("Financial Modeling")
			})
			Field(2, "skill_category", String, "Skill category", func() {
				Enum("TECHNICAL", "SOFT_SKILL", "LANGUAGE", "CERTIFICATION", "DOMAIN_KNOWLEDGE")
				Example("TECHNICAL")
			})
			Field(3, "proficiency_level", String, "Skill proficiency", func() {
				Enum("BEGINNER", "INTERMEDIATE", "ADVANCED", "EXPERT")
				Example("ADVANCED")
			})
			Field(4, "years_experience", UInt, "Years of experience", func() {
				Maximum(50)
				Example(5)
			})
			Field(5, "last_used", String, "Last time skill was used", func() {
				Format(FormatDate)
				Example("2023-12-01")
			})
			Field(6, "certified", Boolean, "Has formal certification", func() {
				Default(false)
			})
			Required("skill_name", "skill_category", "proficiency_level")
		})), "Employee skills and competencies", func() {
			Description("Technical and professional skills inventory")
		})
		
		Field(24, "performance_reviews", ArrayOf(Type("PerformanceReview", func() {
			Field(1, "review_id", String, "Review identifier", func() {
				Format(FormatUUID)
			})
			Field(2, "review_period", String, "Review period", func() {
				Example("Q4 2023")
			})
			Field(3, "review_type", String, "Type of review", func() {
				Enum("ANNUAL", "QUARTERLY", "PROBATIONARY", "PROJECT", "360_DEGREE")
				Example("QUARTERLY")
			})
			Field(4, "overall_rating", String, "Overall performance rating", func() {
				Enum("EXCEEDS", "MEETS", "BELOW", "UNSATISFACTORY")
				Example("MEETS")
			})
			Field(5, "goals_met", UInt, "Number of goals achieved", func() {
				Example(4)
			})
			Field(6, "total_goals", UInt, "Total number of goals", func() {
				Example(5)
			})
			Field(7, "reviewer_id", String, "Manager who conducted review", func() {
				Format(FormatUUID)
			})
			Field(8, "review_date", String, "Date of review", func() {
				Format(FormatDate)
				Example("2023-12-15")
			})
			Field(9, "summary", String, "Review summary", func() {
				MaxLength(1000)
				Example("Strong performance across all objectives. Demonstrated leadership in key projects.")
			})
			Required("review_id", "review_period", "review_type", "overall_rating", "review_date")
		})), "Performance review history", func() {
			Description("Historical performance evaluations and ratings")
		})
		
		Field(25, "training_records", ArrayOf(Type("TrainingRecord", func() {
			Field(1, "training_id", String, "Training identifier", func() {
				Format(FormatUUID)
			})
			Field(2, "training_name", String, "Training program name", func() {
				MaxLength(100)
				Example("Advanced Excel for Financial Analysis")
			})
			Field(3, "training_type", String, "Type of training", func() {
				Enum("MANDATORY", "PROFESSIONAL_DEVELOPMENT", "COMPLIANCE", "SAFETY", "TECHNICAL")
				Example("PROFESSIONAL_DEVELOPMENT")
			})
			Field(4, "provider", String, "Training provider", func() {
				MaxLength(100)
				Example("Corporate University")
			})
			Field(5, "completion_date", String, "Date completed", func() {
				Format(FormatDate)
				Example("2023-08-15")
			})
			Field(6, "expiry_date", String, "Training expiration", func() {
				Format(FormatDate)
				Example("2025-08-15")
			})
			Field(7, "score", Float64, "Training score", func() {
				Minimum(0.0)
				Maximum(100.0)
				Example(92.5)
			})
			Field(8, "status", String, "Training status", func() {
				Enum("COMPLETED", "IN_PROGRESS", "SCHEDULED", "OVERDUE", "WAIVED")
				Example("COMPLETED")
			})
			Required("training_id", "training_name", "training_type", "status")
		})), "Training and development records", func() {
			Description("Professional development and mandatory training")
		})
		
		Field(26, "emergency_contacts", ArrayOf(Type("EmployeeEmergencyContact", func() {
			Field(1, "contact_id", String, "Contact identifier", func() {
				Format(FormatUUID)
			})
			Field(2, "name", String, "Contact name", func() {
				MinLength(2)
				MaxLength(100)
				Example("Jane Doe")
			})
			Field(3, "relationship", String, "Relationship to employee", func() {
				Example("Spouse")
			})
			Field(4, "primary_phone", String, "Primary phone", func() {
				Pattern("^\\+?[1-9]\\d{1,14}$")
				Example("+1-555-999-8888")
			})
			Field(5, "secondary_phone", String, "Secondary phone", func() {
				Pattern("^\\+?[1-9]\\d{1,14}$")
				Example("+1-555-888-7777")
			})
			Field(6, "email", String, "Email address", func() {
				Format(FormatEmail)
				Example("jane.doe@email.com")
			})
			Field(7, "address", Address, "Contact address", func() {
				Description("Emergency contact address")
			})
			Field(8, "is_primary", Boolean, "Primary emergency contact", func() {
				Default(false)
			})
			Required("contact_id", "name", "relationship", "primary_phone")
		})), "Emergency contact information", func() {
			Description("People to contact in case of emergency")
		})
		
		Field(27, "disciplinary_actions", ArrayOf(Type("DisciplinaryAction", func() {
			Field(1, "action_id", String, "Action identifier", func() {
				Format(FormatUUID)
			})
			Field(2, "action_type", String, "Type of disciplinary action", func() {
				Enum("VERBAL_WARNING", "WRITTEN_WARNING", "SUSPENSION", "TERMINATION", "PERFORMANCE_PLAN")
				Example("WRITTEN_WARNING")
			})
			Field(3, "reason", String, "Reason for action", func() {
				MaxLength(500)
				Example("Excessive tardiness - 5 instances in 30 days")
			})
			Field(4, "action_date", String, "Date of action", func() {
				Format(FormatDate)
				Example("2023-11-15")
			})
			Field(5, "issued_by", String, "Manager who issued action", func() {
				Format(FormatUUID)
			})
			Field(6, "resolution_date", String, "Expected resolution date", func() {
				Format(FormatDate)
				Example("2023-12-15")
			})
			Field(7, "status", String, "Action status", func() {
				Enum("ACTIVE", "RESOLVED", "EXPIRED", "APPEALED")
				Example("ACTIVE")
			})
			Required("action_id", "action_type", "reason", "action_date", "issued_by", "status")
		})), "Disciplinary action history", func() {
			Description("Record of disciplinary actions and performance issues")
		})
		
		Field(28, "time_off_balances", MapOf(String, Float64), "Time off balances", func() {
			Example(map[string]any{
				"vacation_days":    15.5,
				"sick_days":       8.0,
				"personal_days":   3.0,
				"comp_time_hours": 16.0,
			})
			Description("Current balances for various time off types")
		})
		
		Field(29, "certifications", ArrayOf(String), "Professional certifications", func() {
			Example([]string{"CPA", "PMP", "Six Sigma Black Belt"})
			Description("Current professional certifications held")
		})
		
		Field(30, "security_clearance", String, "Security clearance level", func() {
			Enum("NONE", "CONFIDENTIAL", "SECRET", "TOP_SECRET", "SCI")
			Default("NONE")
			Example("SECRET")
			Description("Government security clearance level")
		})
		
		Field(31, "background_check", Type("BackgroundCheck", func() {
			Field(1, "completed_date", String, "Background check completion", func() {
				Format(FormatDate)
				Example("2023-01-10")
			})
			Field(2, "status", String, "Check status", func() {
				Enum("PENDING", "CLEARED", "FLAGGED", "FAILED")
				Example("CLEARED")
			})
			Field(3, "vendor", String, "Background check vendor", func() {
				Example("SecureCheck Inc.")
			})
			Field(4, "expiry_date", String, "Check expiration", func() {
				Format(FormatDate)
				Example("2028-01-10")
			})
			Required("status")
		}), "Background check information", func() {
			Description("Security background verification details")
		})
		
		Field(32, "work_authorization", Type("WorkAuthorization", func() {
			Field(1, "status", String, "Authorization status", func() {
				Enum("CITIZEN", "PERMANENT_RESIDENT", "WORK_VISA", "PENDING", "EXPIRED")
				Example("CITIZEN")
			})
			Field(2, "document_type", String, "Authorization document", func() {
				Enum("PASSPORT", "BIRTH_CERTIFICATE", "GREEN_CARD", "H1B", "L1", "TN", "OTHER")
				Example("PASSPORT")
			})
			Field(3, "document_number", String, "Document number", func() {
				MaxLength(50)
				Example("P123456789")
			})
			Field(4, "expiry_date", String, "Authorization expiry", func() {
				Format(FormatDate)
				Example("2030-01-15")
			})
			Field(5, "verified_date", String, "Last verification", func() {
				Format(FormatDate)
				Example("2023-01-15")
			})
			Required("status")
		}), "Work authorization status", func() {
			Description("Legal authorization to work verification")
		})
		
		Field(33, "preferences", MapOf(String, Any), "Employee preferences", func() {
			Example(map[string]any{
				"communication_method": "email",
				"preferred_pronouns":  "he/him",
				"work_from_home_days": []string{"Tuesday", "Thursday"},
				"parking_spot":       "A-25",
			})
			Description("Personal work preferences and accommodations")
		})
		
		Field(34, "notes", String, "HR notes", func() {
			MaxLength(2000)
			Example("High performer, eligible for promotion consideration in Q2 2024")
			Description("Internal HR notes and observations")
		})
		
		Field(35, "tags", ArrayOf(String), "Employee classification tags", func() {
			Example([]string{"high_performer", "leadership_potential", "remote_eligible"})
			Description("Searchable tags for employee categorization")
		})
		
		Field(36, "is_active", Boolean, "Active employment status", func() {
			Default(true)
			Example(true)
			Description("Whether employee is currently active")
		})
		
		// Audit fields from common.go
		AuditFields()
		
		Required("id", "tenant_id", "person_id", "employee_number", "entity_id", 
			"job_title", "employment_type", "employment_status", "hire_date", 
			"is_active", "created_at")
	})
	
	View("default", func() {
		Description("Standard employee view for general listings and references")
		Attribute("id")
		Attribute("employee_number")
		Attribute("person_id")
		Attribute("job_title")
		Attribute("job_level")
		Attribute("entity_id")
		Attribute("manager_id")
		Attribute("employment_status")
		Attribute("hire_date")
		Attribute("is_active")
	})
	
	View("detailed", func() {
		Description("Complete employee profile with all employment information")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("person_id")
		Attribute("employee_number")
		Attribute("badge_number")
		Attribute("entity_id")
		Attribute("manager_id")
		Attribute("job_title")
		Attribute("job_level")
		Attribute("job_family")
		Attribute("employment_type")
		Attribute("employment_status")
		Attribute("hire_date")
		Attribute("start_date")
		Attribute("end_date")
		Attribute("probation_end_date")
		Attribute("work_location")
		Attribute("office_location")
		Attribute("work_schedule")
		Attribute("weekly_hours")
		Attribute("compensation")
		Attribute("benefits")
		Attribute("skills")
		Attribute("performance_reviews")
		Attribute("training_records")
		Attribute("emergency_contacts")
		Attribute("time_off_balances")
		Attribute("certifications")
		Attribute("security_clearance")
		Attribute("background_check")
		Attribute("work_authorization")
		Attribute("preferences")
		Attribute("notes")
		Attribute("tags")
		Attribute("is_active")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
	
	View("hr", func() {
		Description("HR management view with sensitive employment information")
		Attribute("id")
		Attribute("employee_number")
		Attribute("person_id")
		Attribute("job_title")
		Attribute("employment_status")
		Attribute("hire_date")
		Attribute("compensation")
		Attribute("benefits")
		Attribute("performance_reviews")
		Attribute("disciplinary_actions")
		Attribute("time_off_balances")
		Attribute("background_check")
		Attribute("work_authorization")
		Attribute("notes")
	})
	
	View("manager", func() {
		Description("Manager view for direct reports and team information")
		Attribute("id")
		Attribute("employee_number")
		Attribute("person_id")
		Attribute("job_title")
		Attribute("job_level")
		Attribute("employment_status")
		Attribute("skills")
		Attribute("performance_reviews")
		Attribute("training_records")
		Attribute("time_off_balances")
		Attribute("is_active")
	})
	
	View("directory", func() {
		Description("Employee directory view for organizational charts")
		Attribute("id")
		Attribute("employee_number")
		Attribute("person_id")
		Attribute("job_title")
		Attribute("entity_id")
		Attribute("manager_id")
		Attribute("work_location")
		Attribute("employment_status")
	})
	
	View("summary", func() {
		Description("Minimal view for quick references and lookups")
		Attribute("id")
		Attribute("employee_number")
		Attribute("job_title")
		Attribute("employment_status")
		Attribute("is_active")
	})
})