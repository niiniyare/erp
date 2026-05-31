# Human Resources Management

##  Overview

The Human Resources Management module provides  employee lifecycle management, from recruitment and onboarding to performance management and offboarding. It includes payroll processing, benefits administration, time tracking, and compliance with labor regulations.

##  Employee Management

### Employee Master Data

```sql
-- Employee core information
CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id), -- Link to system user account
    
    -- Employee identification
    employee_number VARCHAR(50) UNIQUE NOT NULL,
    employee_status VARCHAR(20) DEFAULT 'active', -- active, inactive, terminated, on_leave
    
    -- Personal information
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    preferred_name VARCHAR(100),
    date_of_birth DATE,
    gender VARCHAR(20),
    marital_status VARCHAR(20),
    nationality VARCHAR(100),
    
    -- Contact information
    personal_email VARCHAR(255),
    work_email VARCHAR(255),
    personal_phone VARCHAR(20),
    work_phone VARCHAR(20),
    mobile_phone VARCHAR(20),
    emergency_contact JSONB, -- {name, relationship, phone, email}
    
    -- Address information
    home_address JSONB,
    mailing_address JSONB,
    
    -- Employment details
    hire_date DATE NOT NULL,
    termination_date DATE,
    rehire_date DATE,
    original_hire_date DATE, -- For rehires
    
    -- Job information
    job_title VARCHAR(255),
    job_level VARCHAR(50),
    department_id UUID REFERENCES organizations(id),
    cost_center_id UUID REFERENCES cost_centers(id),
    location_id UUID REFERENCES asset_locations(id),
    
    -- Reporting structure
    manager_id UUID REFERENCES employees(id),
    reports_to_id UUID REFERENCES employees(id), -- Can be different from manager
    
    -- Employment classification
    employment_type VARCHAR(50) DEFAULT 'full_time', -- full_time, part_time, contract, intern, temporary
    employee_category VARCHAR(50) DEFAULT 'regular', -- regular, probationary, seasonal, consultant
    work_schedule VARCHAR(50) DEFAULT 'standard', -- standard, flexible, shift, remote
    
    -- Compensation
    pay_grade VARCHAR(20),
    pay_frequency VARCHAR(20) DEFAULT 'monthly', -- weekly, bi_weekly, semi_monthly, monthly
    base_salary DECIMAL(12,2),
    hourly_rate DECIMAL(8,2),
    overtime_rate DECIMAL(8,2),
    currency_code CHAR(3) DEFAULT 'USD',
    
    -- Benefits eligibility
    benefits_eligible BOOLEAN DEFAULT true,
    benefits_start_date DATE,
    
    -- Time tracking
    exempt_from_overtime BOOLEAN DEFAULT false,
    time_tracking_required BOOLEAN DEFAULT true,
    default_work_hours_per_day DECIMAL(4,2) DEFAULT 8,
    default_work_days_per_week DECIMAL(3,1) DEFAULT 5,
    
    -- Legal and compliance
    tax_id VARCHAR(50), -- SSN or equivalent
    work_authorization_status VARCHAR(50),
    work_authorization_expiry DATE,
    i9_verification_date DATE,
    
    -- Profile and preferences
    profile_picture_url VARCHAR(500),
    bio TEXT,
    skills JSONB, -- Array of skills
    certifications JSONB, -- Array of certifications
    languages JSONB, -- Array of {language, proficiency}
    
    -- System fields
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_employee_status CHECK (employee_status IN ('active', 'inactive', 'terminated', 'on_leave')),
    CONSTRAINT valid_employment_type CHECK (employment_type IN ('full_time', 'part_time', 'contract', 'intern', 'temporary')),
    CONSTRAINT valid_employee_category CHECK (employee_category IN ('regular', 'probationary', 'seasonal', 'consultant')),
    CONSTRAINT valid_pay_frequency CHECK (pay_frequency IN ('weekly', 'bi_weekly', 'semi_monthly', 'monthly'))
);

-- Employee job history for tracking promotions, transfers, etc.
CREATE TABLE employee_job_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Job change details
    change_type VARCHAR(50) NOT NULL, -- promotion, transfer, demotion, salary_change, title_change
    effective_date DATE NOT NULL,
    end_date DATE,
    
    -- Previous job information
    previous_job_title VARCHAR(255),
    previous_department_id UUID REFERENCES organizations(id),
    previous_manager_id UUID REFERENCES employees(id),
    previous_pay_grade VARCHAR(20),
    previous_base_salary DECIMAL(12,2),
    previous_location_id UUID REFERENCES asset_locations(id),
    
    -- New job information
    new_job_title VARCHAR(255),
    new_department_id UUID REFERENCES organizations(id),
    new_manager_id UUID REFERENCES employees(id),
    new_pay_grade VARCHAR(20),
    new_base_salary DECIMAL(12,2),
    new_location_id UUID REFERENCES asset_locations(id),
    
    -- Change details
    reason VARCHAR(255),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_change_type CHECK (change_type IN (
        'promotion', 'transfer', 'demotion', 'salary_change', 'title_change', 'department_change'
    ))
);

-- Employee documents and attachments
CREATE TABLE employee_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Document details
    document_type VARCHAR(50) NOT NULL, -- contract, resume, certificate, id_copy, etc.
    document_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size INTEGER,
    file_type VARCHAR(20),
    
    -- Document metadata
    upload_date TIMESTAMPTZ DEFAULT NOW(),
    expiry_date DATE,
    is_confidential BOOLEAN DEFAULT false,
    is_required BOOLEAN DEFAULT false,
    
    -- Verification
    verified BOOLEAN DEFAULT false,
    verified_by UUID REFERENCES users(id),
    verified_at TIMESTAMPTZ,
    
    -- Access control
    accessible_to JSONB, -- Array of role IDs or user IDs who can access
    
    notes TEXT,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    
    CONSTRAINT valid_document_type CHECK (document_type IN (
        'contract', 'resume', 'certificate', 'id_copy', 'visa', 'passport', 
        'tax_document', 'bank_details', 'emergency_contact', 'other'
    ))
);
```

### Performance Management

```sql
-- Performance review cycles
CREATE TABLE performance_review_cycles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Cycle identification
    cycle_name VARCHAR(255) NOT NULL,
    cycle_year INTEGER NOT NULL,
    cycle_type VARCHAR(50) DEFAULT 'annual', -- annual, semi_annual, quarterly, monthly
    
    -- Cycle timeline
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    self_review_deadline DATE,
    manager_review_deadline DATE,
    final_review_deadline DATE,
    
    -- Review scope
    applies_to_all_employees BOOLEAN DEFAULT true,
    target_departments JSONB, -- Array of department IDs if not all employees
    target_job_levels JSONB, -- Array of job levels if specific levels
    
    -- Review configuration
    enable_self_review BOOLEAN DEFAULT true,
    enable_peer_review BOOLEAN DEFAULT false,
    enable_360_review BOOLEAN DEFAULT false,
    enable_goal_setting BOOLEAN DEFAULT true,
    
    -- Status
    status VARCHAR(20) DEFAULT 'draft', -- draft, active, completed, cancelled
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_cycle_type CHECK (cycle_type IN ('annual', 'semi_annual', 'quarterly', 'monthly')),
    CONSTRAINT valid_status CHECK (status IN ('draft', 'active', 'completed', 'cancelled'))
);

-- Individual performance reviews
CREATE TABLE performance_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Review identification
    employee_id UUID NOT NULL REFERENCES employees(id),
    review_cycle_id UUID NOT NULL REFERENCES performance_review_cycles(id),
    reviewer_id UUID NOT NULL REFERENCES employees(id), -- Usually the manager
    
    -- Review period
    review_period_start DATE NOT NULL,
    review_period_end DATE NOT NULL,
    
    -- Review status
    status VARCHAR(20) DEFAULT 'not_started', -- not_started, self_review, manager_review, completed
    
    -- Self-review
    self_review_submitted BOOLEAN DEFAULT false,
    self_review_submitted_at TIMESTAMPTZ,
    self_review_comments TEXT,
    
    -- Manager review
    manager_review_completed BOOLEAN DEFAULT false,
    manager_review_completed_at TIMESTAMPTZ,
    manager_comments TEXT,
    
    -- Overall ratings
    overall_rating VARCHAR(20), -- exceeds, meets, below, unsatisfactory
    overall_score DECIMAL(3,2), -- 1.00 to 5.00 scale
    
    -- Development planning
    strengths TEXT,
    areas_for_improvement TEXT,
    development_goals TEXT,
    training_recommendations JSONB,
    
    -- Career planning
    career_aspirations TEXT,
    promotion_readiness VARCHAR(20), -- ready, developing, not_ready
    succession_planning_notes TEXT,
    
    -- Compensation impact
    merit_increase_percentage DECIMAL(5,2),
    bonus_amount DECIMAL(10,2),
    promotion_recommended BOOLEAN DEFAULT false,
    
    -- Final review meeting
    review_meeting_date DATE,
    employee_acknowledgment BOOLEAN DEFAULT false,
    employee_acknowledgment_date TIMESTAMPTZ,
    employee_comments TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_status CHECK (status IN ('not_started', 'self_review', 'manager_review', 'completed')),
    CONSTRAINT valid_overall_rating CHECK (overall_rating IN ('exceeds', 'meets', 'below', 'unsatisfactory')),
    CONSTRAINT valid_promotion_readiness CHECK (promotion_readiness IN ('ready', 'developing', 'not_ready')),
    UNIQUE(employee_id, review_cycle_id)
);

-- Performance goals and objectives
CREATE TABLE performance_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    review_cycle_id UUID REFERENCES performance_review_cycles(id),
    
    -- Goal details
    goal_title VARCHAR(255) NOT NULL,
    goal_description TEXT NOT NULL,
    goal_category VARCHAR(50), -- performance, development, behavioral, project
    
    -- Goal measurement
    measurement_criteria TEXT,
    target_value VARCHAR(100),
    actual_value VARCHAR(100),
    
    -- Timeline
    start_date DATE NOT NULL,
    target_date DATE NOT NULL,
    completion_date DATE,
    
    -- Priority and weight
    priority VARCHAR(20) DEFAULT 'medium', -- high, medium, low
    weight_percentage DECIMAL(5,2) DEFAULT 0, -- Weight in overall performance
    
    -- Progress tracking
    status VARCHAR(20) DEFAULT 'not_started', -- not_started, in_progress, completed, cancelled
    progress_percentage DECIMAL(5,2) DEFAULT 0,
    
    -- Evaluation
    achievement_level VARCHAR(20), -- exceeded, achieved, partially_achieved, not_achieved
    achievement_score DECIMAL(3,2), -- 1.00 to 5.00 scale
    
    -- Comments
    manager_comments TEXT,
    employee_comments TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_goal_category CHECK (goal_category IN ('performance', 'development', 'behavioral', 'project')),
    CONSTRAINT valid_priority CHECK (priority IN ('high', 'medium', 'low')),
    CONSTRAINT valid_goal_status CHECK (status IN ('not_started', 'in_progress', 'completed', 'cancelled')),
    CONSTRAINT valid_achievement_level CHECK (achievement_level IN ('exceeded', 'achieved', 'partially_achieved', 'not_achieved'))
);
```

## ⏰ Time & Attendance

### Time Tracking System

```sql
-- Employee time entries
CREATE TABLE time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Employee and date
    employee_id UUID NOT NULL REFERENCES employees(id),
    entry_date DATE NOT NULL,
    
    -- Time tracking
    clock_in_time TIMESTAMPTZ,
    clock_out_time TIMESTAMPTZ,
    break_start_time TIMESTAMPTZ,
    break_end_time TIMESTAMPTZ,
    
    -- Calculated hours
    regular_hours DECIMAL(6,2) DEFAULT 0,
    overtime_hours DECIMAL(6,2) DEFAULT 0,
    break_hours DECIMAL(6,2) DEFAULT 0,
    total_hours DECIMAL(6,2) DEFAULT 0,
    
    -- Location and method
    clock_in_location JSONB, -- {latitude, longitude, address}
    clock_out_location JSONB,
    entry_method VARCHAR(20) DEFAULT 'manual', -- manual, biometric, mobile, web, kiosk
    
    -- Project and cost allocation
    project_id UUID REFERENCES projects(id),
    cost_center_id UUID REFERENCES cost_centers(id),
    task_description TEXT,
    
    -- Status and approval
    status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, approved, rejected
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    rejection_reason TEXT,
    
    -- System tracking
    ip_address INET,
    device_info JSONB,
    
    -- Comments
    employee_notes TEXT,
    manager_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_entry_method CHECK (entry_method IN ('manual', 'biometric', 'mobile', 'web', 'kiosk')),
    CONSTRAINT valid_status CHECK (status IN ('draft', 'submitted', 'approved', 'rejected'))
);

-- Work schedules for employees
CREATE TABLE work_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Schedule identification
    schedule_name VARCHAR(255) NOT NULL,
    schedule_type VARCHAR(20) DEFAULT 'fixed', -- fixed, flexible, shift, custom
    
    -- Default schedule (for fixed schedules)
    default_start_time TIME,
    default_end_time TIME,
    default_break_duration_minutes INTEGER DEFAULT 60,
    default_work_days JSONB, -- Array of day numbers (0=Sunday, 1=Monday, etc.)
    
    -- Flexible schedule parameters
    core_hours_start TIME,
    core_hours_end TIME,
    minimum_work_hours_per_day DECIMAL(4,2),
    maximum_work_hours_per_day DECIMAL(4,2),
    
    -- Overtime rules
    overtime_threshold_daily DECIMAL(4,2) DEFAULT 8,
    overtime_threshold_weekly DECIMAL(5,2) DEFAULT 40,
    overtime_calculation_method VARCHAR(20) DEFAULT 'daily_and_weekly',
    
    -- Approval requirements
    requires_approval BOOLEAN DEFAULT false,
    auto_approve_within_schedule BOOLEAN DEFAULT true,
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_schedule_type CHECK (schedule_type IN ('fixed', 'flexible', 'shift', 'custom')),
    CONSTRAINT valid_overtime_calculation CHECK (overtime_calculation_method IN ('daily_only', 'weekly_only', 'daily_and_weekly'))
);

-- Employee schedule assignments
CREATE TABLE employee_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    work_schedule_id UUID NOT NULL REFERENCES work_schedules(id),
    
    -- Assignment period
    effective_start_date DATE NOT NULL,
    effective_end_date DATE,
    
    -- Schedule overrides for this employee
    custom_start_time TIME,
    custom_end_time TIME,
    custom_work_days JSONB,
    custom_break_duration_minutes INTEGER,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);
```

### Leave Management

```sql
-- Leave types configuration
CREATE TABLE leave_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Leave type details
    leave_type_code VARCHAR(20) UNIQUE NOT NULL,
    leave_type_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Leave properties
    is_paid BOOLEAN DEFAULT true,
    affects_payroll BOOLEAN DEFAULT true,
    requires_approval BOOLEAN DEFAULT true,
    requires_documentation BOOLEAN DEFAULT false,
    
    -- Accrual settings
    accrual_method VARCHAR(20) DEFAULT 'monthly', -- monthly, annual, per_pay_period, manual
    accrual_rate DECIMAL(6,4), -- Days accrued per period
    accrual_frequency VARCHAR(20) DEFAULT 'monthly',
    
    -- Limits and restrictions
    maximum_balance DECIMAL(6,2),
    maximum_carry_forward DECIMAL(6,2),
    minimum_balance DECIMAL(6,2) DEFAULT 0,
    minimum_notice_days INTEGER DEFAULT 1,
    maximum_consecutive_days INTEGER,
    
    -- Year-end handling
    carry_forward_allowed BOOLEAN DEFAULT true,
    payout_on_termination BOOLEAN DEFAULT false,
    
    -- Approval workflow
    approval_levels INTEGER DEFAULT 1,
    auto_approve_threshold_days DECIMAL(4,2) DEFAULT 0,
    
    -- Calendar settings
    exclude_weekends BOOLEAN DEFAULT true,
    exclude_holidays BOOLEAN DEFAULT true,
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_accrual_method CHECK (accrual_method IN ('monthly', 'annual', 'per_pay_period', 'manual')),
    CONSTRAINT valid_accrual_frequency CHECK (accrual_frequency IN ('weekly', 'bi_weekly', 'monthly', 'quarterly', 'annual'))
);

-- Employee leave balances
CREATE TABLE leave_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    leave_type_id UUID NOT NULL REFERENCES leave_types(id),
    
    -- Balance information
    balance_year INTEGER NOT NULL,
    opening_balance DECIMAL(6,2) DEFAULT 0,
    accrued_balance DECIMAL(6,2) DEFAULT 0,
    used_balance DECIMAL(6,2) DEFAULT 0,
    carried_forward_balance DECIMAL(6,2) DEFAULT 0,
    current_balance DECIMAL(6,2) DEFAULT 0,
    
    -- Pending transactions
    pending_requests DECIMAL(6,2) DEFAULT 0,
    available_balance DECIMAL(6,2) DEFAULT 0,
    
    -- Last accrual processing
    last_accrual_date DATE,
    next_accrual_date DATE,
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(employee_id, leave_type_id, balance_year)
);

-- Leave requests and applications
CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Request identification
    request_number VARCHAR(50) UNIQUE NOT NULL,
    employee_id UUID NOT NULL REFERENCES employees(id),
    leave_type_id UUID NOT NULL REFERENCES leave_types(id),
    
    -- Leave period
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    start_time TIME, -- For partial day leaves
    end_time TIME,
    total_days DECIMAL(5,2) NOT NULL,
    
    -- Request details
    reason TEXT,
    emergency_request BOOLEAN DEFAULT false,
    documentation_provided BOOLEAN DEFAULT false,
    documentation_urls JSONB,
    
    -- Coverage arrangements
    coverage_arranged BOOLEAN DEFAULT false,
    coverage_employee_id UUID REFERENCES employees(id),
    coverage_notes TEXT,
    
    -- Status and workflow
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected, cancelled, completed
    submitted_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Approval chain
    current_approver_id UUID REFERENCES employees(id),
    approval_level INTEGER DEFAULT 1,
    final_approver_id UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    rejection_reason TEXT,
    
    -- Comments
    employee_comments TEXT,
    manager_comments TEXT,
    hr_comments TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_status CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled', 'completed')),
    CONSTRAINT valid_dates CHECK (end_date >= start_date)
);

-- Leave request approvals (for multi-level approval)
CREATE TABLE leave_request_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    leave_request_id UUID NOT NULL REFERENCES leave_requests(id) ON DELETE CASCADE,
    
    -- Approval details
    approval_level INTEGER NOT NULL,
    approver_id UUID NOT NULL REFERENCES employees(id),
    
    -- Approval decision
    decision VARCHAR(20) NOT NULL, -- approved, rejected, pending
    decision_date TIMESTAMPTZ,
    comments TEXT,
    
    -- Delegation
    delegated_from_id UUID REFERENCES employees(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_decision CHECK (decision IN ('approved', 'rejected', 'pending')),
    UNIQUE(leave_request_id, approval_level)
);
```

##  Payroll Management

### Payroll Structure

```sql
-- Payroll periods
CREATE TABLE payroll_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Period identification
    period_name VARCHAR(100) NOT NULL,
    pay_frequency VARCHAR(20) NOT NULL, -- weekly, bi_weekly, semi_monthly, monthly
    
    -- Period dates
    period_start_date DATE NOT NULL,
    period_end_date DATE NOT NULL,
    pay_date DATE NOT NULL,
    
    -- Period status
    status VARCHAR(20) DEFAULT 'open', -- open, processing, approved, paid, closed
    
    -- Processing timestamps
    calculated_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    approved_by UUID REFERENCES users(id),
    paid_at TIMESTAMPTZ,
    
    -- Period statistics
    total_employees INTEGER DEFAULT 0,
    total_gross_pay DECIMAL(15,2) DEFAULT 0,
    total_deductions DECIMAL(15,2) DEFAULT 0,
    total_net_pay DECIMAL(15,2) DEFAULT 0,
    total_employer_taxes DECIMAL(15,2) DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_pay_frequency CHECK (pay_frequency IN ('weekly', 'bi_weekly', 'semi_monthly', 'monthly')),
    CONSTRAINT valid_status CHECK (status IN ('open', 'processing', 'approved', 'paid', 'closed')),
    CONSTRAINT valid_period_dates CHECK (period_end_date >= period_start_date)
);

-- Salary components and pay elements
CREATE TABLE salary_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Component identification
    component_code VARCHAR(20) UNIQUE NOT NULL,
    component_name VARCHAR(255) NOT NULL,
    component_type VARCHAR(20) NOT NULL, -- earning, deduction, benefit, tax
    
    -- Component properties
    is_taxable BOOLEAN DEFAULT true,
    is_fixed BOOLEAN DEFAULT false,
    is_mandatory BOOLEAN DEFAULT false,
    affects_overtime BOOLEAN DEFAULT true,
    
    -- Calculation method
    calculation_method VARCHAR(30) DEFAULT 'fixed', -- fixed, percentage, formula, hourly
    calculation_formula TEXT, -- For complex calculations
    
    -- Tax and compliance
    tax_category VARCHAR(50),
    reporting_category VARCHAR(50),
    
    -- GL account mapping
    expense_account_id UUID REFERENCES accounts(id),
    liability_account_id UUID REFERENCES accounts(id),
    
    -- Component limits
    minimum_amount DECIMAL(10,2),
    maximum_amount DECIMAL(10,2),
    annual_limit DECIMAL(12,2),
    
    -- Display and reporting
    display_order INTEGER DEFAULT 0,
    show_on_payslip BOOLEAN DEFAULT true,
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_component_type CHECK (component_type IN ('earning', 'deduction', 'benefit', 'tax')),
    CONSTRAINT valid_calculation_method CHECK (calculation_method IN ('fixed', 'percentage', 'formula', 'hourly'))
);

-- Employee salary assignments
CREATE TABLE employee_salary_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    salary_component_id UUID NOT NULL REFERENCES salary_components(id),
    
    -- Assignment details
    effective_start_date DATE NOT NULL,
    effective_end_date DATE,
    
    -- Component values
    fixed_amount DECIMAL(10,2),
    percentage_value DECIMAL(5,4),
    calculation_base VARCHAR(50), -- base_salary, gross_pay, specific_component
    
    -- Frequency override
    pay_frequency_override VARCHAR(20),
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    UNIQUE(employee_id, salary_component_id, effective_start_date)
);
```

### Payroll Processing

```typescript
interface PayrollCalculationEngine {
  calculatePayroll(periodId: string): Promise<PayrollResult[]>;
  calculateEmployeePayroll(employeeId: string, periodId: string): Promise<EmployeePayrollResult>;
  processPayrollAdjustments(periodId: string, adjustments: PayrollAdjustment[]): Promise<void>;
  generatePayslips(periodId: string): Promise<Payslip[]>;
}

interface EmployeePayrollResult {
  employee_id: string;
  payroll_period_id: string;
  
  // Earnings breakdown
  earnings: {
    base_salary: number;
    overtime_pay: number;
    commissions: number;
    bonuses: number;
    allowances: number;
    other_earnings: number;
    gross_pay: number;
  };
  
  // Deductions breakdown
  deductions: {
    federal_tax: number;
    state_tax: number;
    social_security: number;
    medicare: number;
    health_insurance: number;
    retirement_401k: number;
    other_deductions: number;
    total_deductions: number;
  };
  
  // Net pay calculation
  net_pay: number;
  
  // Time and attendance
  time_data: {
    regular_hours: number;
    overtime_hours: number;
    vacation_hours: number;
    sick_hours: number;
    holiday_hours: number;
  };
  
  // Year-to-date totals
  ytd_totals: {
    ytd_gross_pay: number;
    ytd_deductions: number;
    ytd_net_pay: number;
    ytd_taxes: number;
  };
}

class PayrollCalculationService implements PayrollCalculationEngine {
  async calculatePayroll(periodId: string): Promise<PayrollResult[]> {
    const period = await this.getPayrollPeriod(periodId);
    const employees = await this.getEligibleEmployees(period.tenant_id, period);
    const results: PayrollResult[] = [];
    
    for (const employee of employees) {
      try {
        const result = await this.calculateEmployeePayroll(employee.id, periodId);
        results.push(result);
      } catch (error) {
        console.error(`Failed to calculate payroll for employee ${employee.id}:`, error);
        // Log error but continue with other employees
      }
    }
    
    return results;
  }
  
  async calculateEmployeePayroll(employeeId: string, periodId: string): Promise<EmployeePayrollResult> {
    const period = await this.getPayrollPeriod(periodId);
    const employee = await this.getEmployee(employeeId);
    const salaryComponents = await this.getEmployeeSalaryComponents(employeeId, period.period_end_date);
    const timeData = await this.getTimeData(employeeId, period);
    
    // Calculate earnings
    const earnings = await this.calculateEarnings(employee, salaryComponents, timeData, period);
    
    // Calculate deductions
    const deductions = await this.calculateDeductions(employee, earnings, salaryComponents, period);
    
    // Calculate net pay
    const netPay = earnings.gross_pay - deductions.total_deductions;
    
    // Get YTD totals
    const ytdTotals = await this.calculateYTDTotals(employeeId, period);
    
    return {
      employee_id: employeeId,
      payroll_period_id: periodId,
      earnings,
      deductions,
      net_pay: netPay,
      time_data: timeData,
      ytd_totals: ytdTotals
    };
  }
  
  private async calculateEarnings(
    employee: Employee, 
    components: SalaryComponent[], 
    timeData: TimeData, 
    period: PayrollPeriod
  ) {
    const earningComponents = components.filter(c => c.component_type === 'earning');
    let earnings = {
      base_salary: 0,
      overtime_pay: 0,
      commissions: 0,
      bonuses: 0,
      allowances: 0,
      other_earnings: 0,
      gross_pay: 0
    };
    
    for (const component of earningComponents) {
      let amount = 0;
      
      switch (component.calculation_method) {
        case 'fixed':
          amount = component.fixed_amount || 0;
          break;
          
        case 'hourly':
          if (component.component_code === 'BASE_SALARY') {
            amount = (employee.hourly_rate || 0) * timeData.regular_hours;
            earnings.base_salary = amount;
          } else if (component.component_code === 'OVERTIME') {
            amount = (employee.overtime_rate || employee.hourly_rate * 1.5 || 0) * timeData.overtime_hours;
            earnings.overtime_pay = amount;
          }
          break;
          
        case 'percentage':
          const baseAmount = this.getCalculationBase(component.calculation_base, earnings);
          amount = baseAmount * (component.percentage_value || 0);
          break;
          
        case 'formula':
          amount = await this.evaluateFormula(component.calculation_formula, {
            employee,
            timeData,
            period,
            earnings
          });
          break;
      }
      
      // Categorize earnings
      switch (component.component_code) {
        case 'COMMISSION':
          earnings.commissions += amount;
          break;
        case 'BONUS':
          earnings.bonuses += amount;
          break;
        case 'ALLOWANCE':
          earnings.allowances += amount;
          break;
        default:
          earnings.other_earnings += amount;
      }
      
      earnings.gross_pay += amount;
    }
    
    return earnings;
  }
  
  private async calculateDeductions(
    employee: Employee, 
    earnings: any, 
    components: SalaryComponent[], 
    period: PayrollPeriod
  ) {
    const deductionComponents = components.filter(c => c.component_type === 'deduction' || c.component_type === 'tax');
    let deductions = {
      federal_tax: 0,
      state_tax: 0,
      social_security: 0,
      medicare: 0,
      health_insurance: 0,
      retirement_401k: 0,
      other_deductions: 0,
      total_deductions: 0
    };
    
    for (const component of deductionComponents) {
      let amount = 0;
      
      switch (component.calculation_method) {
        case 'fixed':
          amount = component.fixed_amount || 0;
          break;
          
        case 'percentage':
          const baseAmount = this.getCalculationBase(component.calculation_base, earnings);
          amount = baseAmount * (component.percentage_value || 0);
          break;
          
        case 'formula':
          amount = await this.evaluateFormula(component.calculation_formula, {
            employee,
            earnings,
            period
          });
          break;
      }
      
      // Apply limits
      if (component.maximum_amount && amount > component.maximum_amount) {
        amount = component.maximum_amount;
      }
      if (component.minimum_amount && amount < component.minimum_amount) {
        amount = component.minimum_amount;
      }
      
      // Categorize deductions
      switch (component.tax_category) {
        case 'FEDERAL_TAX':
          deductions.federal_tax += amount;
          break;
        case 'STATE_TAX':
          deductions.state_tax += amount;
          break;
        case 'SOCIAL_SECURITY':
          deductions.social_security += amount;
          break;
        case 'MEDICARE':
          deductions.medicare += amount;
          break;
        case 'HEALTH_INSURANCE':
          deductions.health_insurance += amount;
          break;
        case 'RETIREMENT_401K':
          deductions.retirement_401k += amount;
          break;
        default:
          deductions.other_deductions += amount;
      }
      
      deductions.total_deductions += amount;
    }
    
    return deductions;
  }
}
```

### Payroll Records

```sql
-- Employee payroll records
CREATE TABLE employee_payroll_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Record identification
    payroll_period_id UUID NOT NULL REFERENCES payroll_periods(id),
    employee_id UUID NOT NULL REFERENCES employees(id),
    
    -- Time and attendance
    regular_hours DECIMAL(6,2) DEFAULT 0,
    overtime_hours DECIMAL(6,2) DEFAULT 0,
    vacation_hours DECIMAL(6,2) DEFAULT 0,
    sick_hours DECIMAL(6,2) DEFAULT 0,
    holiday_hours DECIMAL(6,2) DEFAULT 0,
    
    -- Earnings
    base_salary DECIMAL(12,2) DEFAULT 0,
    overtime_pay DECIMAL(12,2) DEFAULT 0,
    commissions DECIMAL(12,2) DEFAULT 0,
    bonuses DECIMAL(12,2) DEFAULT 0,
    allowances DECIMAL(12,2) DEFAULT 0,
    other_earnings DECIMAL(12,2) DEFAULT 0,
    gross_pay DECIMAL(12,2) DEFAULT 0,
    
    -- Pre-tax deductions
    pretax_health_insurance DECIMAL(10,2) DEFAULT 0,
    pretax_retirement_401k DECIMAL(10,2) DEFAULT 0,
    pretax_other DECIMAL(10,2) DEFAULT 0,
    
    -- Taxable income
    taxable_income DECIMAL(12,2) DEFAULT 0,
    
    -- Tax deductions
    federal_income_tax DECIMAL(10,2) DEFAULT 0,
    state_income_tax DECIMAL(10,2) DEFAULT 0,
    social_security_tax DECIMAL(10,2) DEFAULT 0,
    medicare_tax DECIMAL(10,2) DEFAULT 0,
    state_disability_tax DECIMAL(10,2) DEFAULT 0,
    unemployment_tax DECIMAL(10,2) DEFAULT 0,
    
    -- Post-tax deductions
    posttax_health_insurance DECIMAL(10,2) DEFAULT 0,
    posttax_life_insurance DECIMAL(10,2) DEFAULT 0,
    posttax_other DECIMAL(10,2) DEFAULT 0,
    
    -- Total deductions and net pay
    total_deductions DECIMAL(12,2) DEFAULT 0,
    net_pay DECIMAL(12,2) DEFAULT 0,
    
    -- Employer taxes and contributions
    employer_social_security DECIMAL(10,2) DEFAULT 0,
    employer_medicare DECIMAL(10,2) DEFAULT 0,
    employer_unemployment DECIMAL(10,2) DEFAULT 0,
    employer_retirement_match DECIMAL(10,2) DEFAULT 0,
    
    -- Year-to-date totals
    ytd_gross_pay DECIMAL(15,2) DEFAULT 0,
    ytd_deductions DECIMAL(15,2) DEFAULT 0,
    ytd_net_pay DECIMAL(15,2) DEFAULT 0,
    ytd_federal_tax DECIMAL(15,2) DEFAULT 0,
    ytd_state_tax DECIMAL(15,2) DEFAULT 0,
    ytd_social_security DECIMAL(15,2) DEFAULT 0,
    ytd_medicare DECIMAL(15,2) DEFAULT 0,
    
    -- Record status
    status VARCHAR(20) DEFAULT 'calculated', -- calculated, approved, paid, voided
    approved_at TIMESTAMPTZ,
    approved_by UUID REFERENCES users(id),
    
    -- Payment information
    payment_method VARCHAR(20) DEFAULT 'direct_deposit', -- direct_deposit, check, cash
    payment_reference VARCHAR(100),
    payment_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_status CHECK (status IN ('calculated', 'approved', 'paid', 'voided')),
    CONSTRAINT valid_payment_method CHECK (payment_method IN ('direct_deposit', 'check', 'cash')),
    UNIQUE(payroll_period_id, employee_id)
);

-- Payroll adjustments for corrections and one-time payments
CREATE TABLE payroll_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Adjustment identification
    adjustment_number VARCHAR(50) UNIQUE NOT NULL,
    employee_id UUID NOT NULL REFERENCES employees(id),
    payroll_period_id UUID REFERENCES payroll_periods(id),
    
    -- Adjustment details
    adjustment_type VARCHAR(30) NOT NULL, -- correction, bonus, reimbursement, deduction
    adjustment_reason TEXT NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    
    -- Tax implications
    affects_taxes BOOLEAN DEFAULT true,
    is_taxable BOOLEAN DEFAULT true,
    
    -- GL account mapping
    debit_account_id UUID REFERENCES accounts(id),
    credit_account_id UUID REFERENCES accounts(id),
    
    -- Approval
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    
    -- Processing
    processed BOOLEAN DEFAULT false,
    processed_at TIMESTAMPTZ,
    processed_in_period_id UUID REFERENCES payroll_periods(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_adjustment_type CHECK (adjustment_type IN ('correction', 'bonus', 'reimbursement', 'deduction'))
);
```

This  HR management system provides complete employee lifecycle management, sophisticated time tracking, flexible leave management, and robust payroll processing capabilities while ensuring compliance with labor regulations and tax requirements.
