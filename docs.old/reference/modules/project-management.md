# Project Management

##  Overview

The Project Management module provides  project planning, execution, and monitoring capabilities. It supports multiple project methodologies (Waterfall, Agile, Hybrid), resource management, time tracking, budget control, and collaboration tools for successful project delivery.

##  Project Structure & Planning

### Project Master Data

```sql
-- Projects master table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Project identification
    project_number VARCHAR(50) UNIQUE NOT NULL,
    project_name VARCHAR(255) NOT NULL,
    project_code VARCHAR(20) UNIQUE,
    description TEXT,
    
    -- Project classification
    project_type VARCHAR(50) DEFAULT 'internal', -- internal, client, product, research, maintenance
    project_category VARCHAR(100),
    priority VARCHAR(20) DEFAULT 'medium', -- low, medium, high, critical
    complexity VARCHAR(20) DEFAULT 'medium', -- simple, medium, complex, enterprise
    
    -- Project scope and objectives
    objectives TEXT,
    scope_description TEXT,
    success_criteria TEXT,
    deliverables JSONB, -- Array of deliverable descriptions
    
    -- Project timeline
    planned_start_date DATE NOT NULL,
    planned_end_date DATE NOT NULL,
    actual_start_date DATE,
    actual_end_date DATE,
    
    -- Project hierarchy
    parent_project_id UUID REFERENCES projects(id),
    program_id UUID REFERENCES programs(id),
    portfolio_id UUID REFERENCES portfolios(id),
    
    -- Project management
    project_manager_id UUID NOT NULL REFERENCES employees(id),
    sponsor_id UUID REFERENCES employees(id),
    client_id UUID REFERENCES customers(id),
    
    -- Financial tracking
    approved_budget DECIMAL(15,2),
    baseline_budget DECIMAL(15,2),
    actual_cost DECIMAL(15,2) DEFAULT 0,
    committed_cost DECIMAL(15,2) DEFAULT 0,
    remaining_budget DECIMAL(15,2),
    currency_code CHAR(3) DEFAULT 'USD',
    
    -- Revenue (for billable projects)
    contract_value DECIMAL(15,2),
    billed_amount DECIMAL(15,2) DEFAULT 0,
    collected_amount DECIMAL(15,2) DEFAULT 0,
    billing_type VARCHAR(20) DEFAULT 'time_and_material', -- fixed_price, time_and_material, milestone
    
    -- Progress tracking
    progress_percentage DECIMAL(5,2) DEFAULT 0,
    health_status VARCHAR(20) DEFAULT 'green', -- green, amber, red
    overall_status VARCHAR(20) DEFAULT 'planning', -- planning, active, on_hold, completed, cancelled
    
    -- Project methodology
    methodology VARCHAR(50) DEFAULT 'waterfall', -- waterfall, agile, scrum, kanban, hybrid
    sprint_duration_weeks INTEGER DEFAULT 2,
    
    -- Quality and risk
    quality_gate_required BOOLEAN DEFAULT false,
    risk_level VARCHAR(20) DEFAULT 'medium', -- low, medium, high
    
    -- Location and team
    primary_location_id UUID REFERENCES asset_locations(id),
    department_id UUID REFERENCES organizations(id),
    cost_center_id UUID REFERENCES cost_centers(id),
    
    -- Project settings
    time_tracking_enabled BOOLEAN DEFAULT true,
    expense_tracking_enabled BOOLEAN DEFAULT true,
    requires_timesheet_approval BOOLEAN DEFAULT true,
    billable_by_default BOOLEAN DEFAULT false,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_project_type CHECK (project_type IN ('internal', 'client', 'product', 'research', 'maintenance')),
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT valid_complexity CHECK (complexity IN ('simple', 'medium', 'complex', 'enterprise')),
    CONSTRAINT valid_health_status CHECK (health_status IN ('green', 'amber', 'red')),
    CONSTRAINT valid_overall_status CHECK (overall_status IN ('planning', 'active', 'on_hold', 'completed', 'cancelled')),
    CONSTRAINT valid_methodology CHECK (methodology IN ('waterfall', 'agile', 'scrum', 'kanban', 'hybrid')),
    CONSTRAINT valid_billing_type CHECK (billing_type IN ('fixed_price', 'time_and_material', 'milestone')),
    CONSTRAINT valid_dates CHECK (planned_end_date >= planned_start_date)
);

-- Project phases for waterfall methodology
CREATE TABLE project_phases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Phase identification
    phase_name VARCHAR(255) NOT NULL,
    phase_number INTEGER NOT NULL,
    description TEXT,
    
    -- Phase timeline
    planned_start_date DATE NOT NULL,
    planned_end_date DATE NOT NULL,
    actual_start_date DATE,
    actual_end_date DATE,
    
    -- Phase dependencies
    predecessor_phase_id UUID REFERENCES project_phases(id),
    
    -- Phase budget and progress
    phase_budget DECIMAL(12,2),
    actual_cost DECIMAL(12,2) DEFAULT 0,
    progress_percentage DECIMAL(5,2) DEFAULT 0,
    
    -- Phase status
    status VARCHAR(20) DEFAULT 'not_started', -- not_started, in_progress, completed, on_hold, cancelled
    
    -- Quality gates
    quality_gate_required BOOLEAN DEFAULT false,
    quality_gate_passed BOOLEAN DEFAULT false,
    quality_gate_date DATE,
    
    -- Deliverables
    deliverables JSONB, -- Array of phase deliverables
    
    -- Approval
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_phase_status CHECK (status IN ('not_started', 'in_progress', 'completed', 'on_hold', 'cancelled')),
    UNIQUE(project_id, phase_number)
);

-- Agile sprints for agile methodology
CREATE TABLE sprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Sprint identification
    sprint_name VARCHAR(255) NOT NULL,
    sprint_number INTEGER NOT NULL,
    sprint_goal TEXT,
    
    -- Sprint timeline
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    
    -- Sprint capacity
    planned_capacity_hours DECIMAL(8,2),
    actual_capacity_hours DECIMAL(8,2),
    team_velocity DECIMAL(6,2), -- Story points or hours per sprint
    
    -- Sprint status
    status VARCHAR(20) DEFAULT 'planned', -- planned, active, completed, cancelled
    
    -- Sprint metrics
    story_points_committed INTEGER DEFAULT 0,
    story_points_completed INTEGER DEFAULT 0,
    tasks_total INTEGER DEFAULT 0,
    tasks_completed INTEGER DEFAULT 0,
    
    -- Sprint review
    demo_date TIMESTAMPTZ,
    retrospective_date TIMESTAMPTZ,
    retrospective_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_sprint_status CHECK (status IN ('planned', 'active', 'completed', 'cancelled')),
    CONSTRAINT valid_sprint_dates CHECK (end_date >= start_date),
    UNIQUE(project_id, sprint_number)
);
```

### Work Breakdown Structure (WBS)

```sql
-- Project tasks and work items
CREATE TABLE project_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Task identification
    task_number VARCHAR(50) NOT NULL,
    task_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Work breakdown structure
    parent_task_id UUID REFERENCES project_tasks(id),
    wbs_code VARCHAR(50), -- Hierarchical code like 1.2.3.1
    level_number INTEGER DEFAULT 1,
    sort_order INTEGER DEFAULT 0,
    
    -- Task classification
    task_type VARCHAR(50) DEFAULT 'task', -- milestone, summary, task, deliverable
    work_type VARCHAR(50) DEFAULT 'development', -- planning, development, testing, documentation, review
    
    -- Task assignment
    assigned_to_id UUID REFERENCES employees(id),
    assigned_team_id UUID REFERENCES teams(id),
    
    -- Task timeline
    planned_start_date DATE,
    planned_end_date DATE,
    planned_duration_hours DECIMAL(8,2),
    
    actual_start_date DATE,
    actual_end_date DATE,
    actual_duration_hours DECIMAL(8,2) DEFAULT 0,
    
    -- Progress tracking
    estimated_effort_hours DECIMAL(8,2),
    actual_effort_hours DECIMAL(8,2) DEFAULT 0,
    remaining_effort_hours DECIMAL(8,2),
    progress_percentage DECIMAL(5,2) DEFAULT 0,
    
    -- Agile/Scrum fields
    story_points INTEGER,
    sprint_id UUID REFERENCES sprints(id),
    epic_id UUID REFERENCES epics(id),
    
    -- Task priority and status
    priority VARCHAR(20) DEFAULT 'medium', -- low, medium, high, critical
    status VARCHAR(20) DEFAULT 'not_started', -- not_started, in_progress, completed, on_hold, cancelled
    
    -- Dependencies
    constraint_type VARCHAR(20) DEFAULT 'finish_to_start', -- finish_to_start, start_to_start, finish_to_finish, start_to_finish
    
    -- Financial tracking
    budgeted_cost DECIMAL(12,2),
    actual_cost DECIMAL(12,2) DEFAULT 0,
    
    -- Quality and acceptance
    acceptance_criteria TEXT,
    definition_of_done TEXT,
    quality_reviewed BOOLEAN DEFAULT false,
    quality_approved BOOLEAN DEFAULT false,
    
    -- Task metadata
    is_critical_path BOOLEAN DEFAULT false,
    is_milestone BOOLEAN DEFAULT false,
    is_billable BOOLEAN DEFAULT false,
    
    -- Comments and notes
    notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_task_type CHECK (task_type IN ('milestone', 'summary', 'task', 'deliverable')),
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT valid_status CHECK (status IN ('not_started', 'in_progress', 'completed', 'on_hold', 'cancelled')),
    CONSTRAINT valid_constraint_type CHECK (constraint_type IN ('finish_to_start', 'start_to_start', 'finish_to_finish', 'start_to_finish')),
    UNIQUE(project_id, task_number)
);

-- Task dependencies for scheduling
CREATE TABLE task_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    predecessor_task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    successor_task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    
    -- Dependency type
    dependency_type VARCHAR(20) DEFAULT 'finish_to_start',
    lag_days INTEGER DEFAULT 0, -- Positive for delay, negative for overlap
    
    -- Dependency metadata
    is_critical BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_dependency_type CHECK (dependency_type IN ('finish_to_start', 'start_to_start', 'finish_to_finish', 'start_to_finish')),
    CONSTRAINT no_self_dependency CHECK (predecessor_task_id != successor_task_id),
    UNIQUE(predecessor_task_id, successor_task_id)
);

-- Task assignments for team collaboration
CREATE TABLE task_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    assignee_id UUID NOT NULL REFERENCES employees(id),
    
    -- Assignment details
    assignment_type VARCHAR(20) DEFAULT 'responsible', -- responsible, accountable, consulted, informed
    allocation_percentage DECIMAL(5,2) DEFAULT 100, -- Percentage of person's time allocated
    
    -- Assignment timeline
    assigned_date DATE DEFAULT CURRENT_DATE,
    start_date DATE,
    end_date DATE,
    
    -- Assignment status
    status VARCHAR(20) DEFAULT 'assigned', -- assigned, accepted, in_progress, completed, declined
    accepted_at TIMESTAMPTZ,
    
    -- Work allocation
    estimated_hours DECIMAL(8,2),
    actual_hours DECIMAL(8,2) DEFAULT 0,
    
    -- Assignment notes
    assignment_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_assignment_type CHECK (assignment_type IN ('responsible', 'accountable', 'consulted', 'informed')),
    CONSTRAINT valid_assignment_status CHECK (status IN ('assigned', 'accepted', 'in_progress', 'completed', 'declined')),
    UNIQUE(task_id, assignee_id)
);
```

## ‍ Resource Management

### Team and Resource Planning

```sql
-- Project teams
CREATE TABLE project_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Team identification
    team_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Team lead
    team_lead_id UUID REFERENCES employees(id),
    
    -- Team settings
    max_team_size INTEGER DEFAULT 10,
    is_cross_functional BOOLEAN DEFAULT false,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- Project team members
CREATE TABLE project_team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_team_id UUID NOT NULL REFERENCES project_teams(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id),
    
    -- Member role and allocation
    role VARCHAR(100), -- Developer, Designer, QA, Business Analyst, etc.
    allocation_percentage DECIMAL(5,2) DEFAULT 100,
    hourly_rate DECIMAL(8,2),
    
    -- Membership timeline
    start_date DATE NOT NULL,
    end_date DATE,
    
    -- Member status
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, temporary
    
    -- Skills and capabilities
    skills JSONB, -- Array of skills relevant to the project
    certification_level VARCHAR(20), -- junior, intermediate, senior, expert
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_member_status CHECK (status IN ('active', 'inactive', 'temporary')),
    UNIQUE(project_team_id, employee_id)
);

-- Resource planning and capacity
CREATE TABLE resource_capacity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Resource identification
    employee_id UUID NOT NULL REFERENCES employees(id),
    capacity_date DATE NOT NULL,
    
    -- Capacity allocation
    available_hours DECIMAL(6,2) DEFAULT 8, -- Standard work day
    allocated_hours DECIMAL(6,2) DEFAULT 0, -- Hours allocated to projects
    overtime_hours DECIMAL(6,2) DEFAULT 0,
    leave_hours DECIMAL(6,2) DEFAULT 0,
    
    -- Capacity utilization
    utilization_percentage DECIMAL(5,2) DEFAULT 0,
    
    -- Capacity notes
    notes TEXT,
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(employee_id, capacity_date)
);

-- Resource allocation to projects
CREATE TABLE resource_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id),
    
    -- Allocation period
    allocation_start_date DATE NOT NULL,
    allocation_end_date DATE,
    
    -- Allocation details
    role VARCHAR(100),
    allocation_percentage DECIMAL(5,2) NOT NULL,
    daily_hours DECIMAL(6,2),
    
    -- Billing information
    billing_rate DECIMAL(8,2),
    is_billable BOOLEAN DEFAULT true,
    
    -- Allocation status
    status VARCHAR(20) DEFAULT 'planned', -- planned, confirmed, active, completed
    
    -- Approval
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_allocation_status CHECK (status IN ('planned', 'confirmed', 'active', 'completed'))
);
```

### Resource Optimization

```typescript
interface ResourceOptimizer {
  optimizeResourceAllocation(projectId: string, constraints: ResourceConstraints): Promise<OptimizationResult>;
  detectResourceConflicts(employeeId: string, dateRange: DateRange): Promise<ResourceConflict[]>;
  suggestResourceReallocation(projectId: string): Promise<ReallocationSuggestion[]>;
  calculateResourceUtilization(employeeId: string, period: DateRange): Promise<UtilizationMetrics>;
}

interface ResourceConstraints {
  max_overtime_percentage: number;
  required_skills: string[];
  budget_constraints: BudgetConstraint[];
  timeline_constraints: TimelineConstraint[];
  availability_constraints: AvailabilityConstraint[];
}

interface OptimizationResult {
  feasible: boolean;
  recommended_allocations: ResourceAllocation[];
  conflicts: ResourceConflict[];
  suggestions: OptimizationSuggestion[];
  cost_impact: number;
  timeline_impact: number;
}

class ResourcePlanningService implements ResourceOptimizer {
  async optimizeResourceAllocation(projectId: string, constraints: ResourceConstraints): Promise<OptimizationResult> {
    const project = await this.getProject(projectId);
    const tasks = await this.getProjectTasks(projectId);
    const availableResources = await this.getAvailableResources(project.tenant_id, project.planned_start_date, project.planned_end_date);
    
    // Create optimization model
    const model = this.createOptimizationModel(tasks, availableResources, constraints);
    
    // Solve using linear programming or heuristic algorithms
    const solution = await this.solveOptimization(model);
    
    // Analyze results and generate recommendations
    const conflicts = await this.detectResourceConflicts(solution);
    const suggestions = this.generateOptimizationSuggestions(solution, constraints);
    
    return {
      feasible: solution.feasible,
      recommended_allocations: solution.allocations,
      conflicts,
      suggestions,
      cost_impact: solution.totalCost - project.baseline_budget,
      timeline_impact: this.calculateTimelineImpact(solution, project)
    };
  }
  
  async detectResourceConflicts(employeeId: string, dateRange: DateRange): Promise<ResourceConflict[]> {
    const allocations = await this.getResourceAllocations(employeeId, dateRange);
    const capacity = await this.getResourceCapacity(employeeId, dateRange);
    const conflicts: ResourceConflict[] = [];
    
    // Group allocations by date
    const allocationsByDate = this.groupAllocationsByDate(allocations);
    
    for (const [date, dayAllocations] of allocationsByDate.entries()) {
      const dayCapacity = capacity.find(c => c.capacity_date.toDateString() === date);
      const totalAllocated = dayAllocations.reduce((sum, alloc) => sum + alloc.daily_hours, 0);
      const availableHours = (dayCapacity?.available_hours || 8) - (dayCapacity?.leave_hours || 0);
      
      if (totalAllocated > availableHours) {
        conflicts.push({
          employee_id: employeeId,
          conflict_date: new Date(date),
          available_hours: availableHours,
          allocated_hours: totalAllocated,
          overallocation_hours: totalAllocated - availableHours,
          conflicting_projects: dayAllocations.map(a => a.project_id),
          severity: this.calculateConflictSeverity(totalAllocated, availableHours)
        });
      }
    }
    
    return conflicts;
  }
  
  async calculateResourceUtilization(employeeId: string, period: DateRange): Promise<UtilizationMetrics> {
    const allocations = await this.getResourceAllocations(employeeId, period);
    const capacity = await this.getResourceCapacity(employeeId, period);
    const timeEntries = await this.getTimeEntries(employeeId, period);
    
    const totalAvailableHours = capacity.reduce((sum, c) => sum + c.available_hours - c.leave_hours, 0);
    const totalAllocatedHours = allocations.reduce((sum, a) => sum + a.daily_hours, 0);
    const totalActualHours = timeEntries.reduce((sum, t) => sum + t.total_hours, 0);
    const totalBillableHours = timeEntries.filter(t => t.is_billable).reduce((sum, t) => sum + t.total_hours, 0);
    
    return {
      employee_id: employeeId,
      period,
      total_available_hours: totalAvailableHours,
      total_allocated_hours: totalAllocatedHours,
      total_actual_hours: totalActualHours,
      total_billable_hours: totalBillableHours,
      utilization_percentage: (totalActualHours / totalAvailableHours) * 100,
      allocation_percentage: (totalAllocatedHours / totalAvailableHours) * 100,
      billable_percentage: (totalBillableHours / totalActualHours) * 100,
      efficiency_score: (totalActualHours / totalAllocatedHours) * 100
    };
  }
}
```

## ⏱️ Time Tracking & Timesheets

### Time Entry Management

```sql
-- Project time entries
CREATE TABLE project_time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Time entry identification
    employee_id UUID NOT NULL REFERENCES employees(id),
    project_id UUID NOT NULL REFERENCES projects(id),
    task_id UUID REFERENCES project_tasks(id),
    
    -- Time tracking
    entry_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    duration_hours DECIMAL(6,2) NOT NULL,
    break_duration_minutes INTEGER DEFAULT 0,
    
    -- Work classification
    work_type VARCHAR(50) DEFAULT 'regular', -- regular, overtime, weekend, holiday
    activity_type VARCHAR(50), -- development, testing, documentation, meeting, etc.
    
    -- Billing information
    is_billable BOOLEAN DEFAULT false,
    billing_rate DECIMAL(8,2),
    billing_amount DECIMAL(10,2),
    
    -- Entry details
    description TEXT NOT NULL,
    location VARCHAR(100),
    
    -- Status and approval
    status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, approved, rejected, invoiced
    submitted_at TIMESTAMPTZ,
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    rejection_reason TEXT,
    
    -- Invoice tracking
    invoice_id UUID REFERENCES sales_invoices(id),
    invoiced_at TIMESTAMPTZ,
    
    -- Time entry metadata
    entry_method VARCHAR(20) DEFAULT 'manual', -- manual, timer, mobile, import
    timer_started_at TIMESTAMPTZ,
    timer_stopped_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_work_type CHECK (work_type IN ('regular', 'overtime', 'weekend', 'holiday')),
    CONSTRAINT valid_status CHECK (status IN ('draft', 'submitted', 'approved', 'rejected', 'invoiced')),
    CONSTRAINT valid_entry_method CHECK (entry_method IN ('manual', 'timer', 'mobile', 'import')),
    CONSTRAINT positive_duration CHECK (duration_hours > 0)
);

-- Timesheet management
CREATE TABLE timesheets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Timesheet identification
    timesheet_number VARCHAR(50) UNIQUE NOT NULL,
    employee_id UUID NOT NULL REFERENCES employees(id),
    
    -- Timesheet period
    week_start_date DATE NOT NULL,
    week_end_date DATE NOT NULL,
    
    -- Timesheet totals
    total_hours DECIMAL(8,2) DEFAULT 0,
    regular_hours DECIMAL(8,2) DEFAULT 0,
    overtime_hours DECIMAL(8,2) DEFAULT 0,
    billable_hours DECIMAL(8,2) DEFAULT 0,
    
    -- Status and workflow
    status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, approved, rejected, locked
    submitted_at TIMESTAMPTZ,
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    rejection_reason TEXT,
    
    -- Comments
    employee_comments TEXT,
    manager_comments TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_timesheet_status CHECK (status IN ('draft', 'submitted', 'approved', 'rejected', 'locked')),
    CONSTRAINT valid_week_period CHECK (week_end_date >= week_start_date),
    UNIQUE(employee_id, week_start_date)
);

-- Timesheet line items (summary of time entries)
CREATE TABLE timesheet_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timesheet_id UUID NOT NULL REFERENCES timesheets(id) ON DELETE CASCADE,
    
    -- Project and task reference
    project_id UUID NOT NULL REFERENCES projects(id),
    task_id UUID REFERENCES project_tasks(id),
    
    -- Daily time breakdown
    monday_hours DECIMAL(5,2) DEFAULT 0,
    tuesday_hours DECIMAL(5,2) DEFAULT 0,
    wednesday_hours DECIMAL(5,2) DEFAULT 0,
    thursday_hours DECIMAL(5,2) DEFAULT 0,
    friday_hours DECIMAL(5,2) DEFAULT 0,
    saturday_hours DECIMAL(5,2) DEFAULT 0,
    sunday_hours DECIMAL(5,2) DEFAULT 0,
    
    -- Line totals
    total_hours DECIMAL(6,2) DEFAULT 0,
    billable_hours DECIMAL(6,2) DEFAULT 0,
    
    -- Line description
    description TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Time Tracking Analytics

```typescript
interface TimeTrackingAnalytics {
  generateProductivityReport(employeeId: string, period: DateRange): Promise<ProductivityReport>;
  analyzeProjectTimeSpent(projectId: string): Promise<ProjectTimeAnalysis>;
  calculateBillableUtilization(employeeId: string, period: DateRange): Promise<BillableUtilizationReport>;
  identifyTimeTrackingPatterns(tenantId: string): Promise<TimeTrackingPattern[]>;
}

interface ProductivityReport {
  employee_id: string;
  reporting_period: DateRange;
  
  time_distribution: {
    total_hours: number;
    billable_hours: number;
    non_billable_hours: number;
    project_hours: number;
    administrative_hours: number;
    meeting_hours: number;
  };
  
  productivity_metrics: {
    hours_per_day_average: number;
    billable_percentage: number;
    project_efficiency: number;
    multitasking_index: number; // Number of different projects/tasks per day
  };
  
  project_breakdown: {
    project_id: string;
    project_name: string;
    hours_spent: number;
    percentage_of_total: number;
    efficiency_rating: number;
  }[];
  
  trends: {
    daily_patterns: DailyPattern[];
    weekly_patterns: WeeklyPattern[];
    productivity_trend: 'increasing' | 'stable' | 'decreasing';
  };
}

class TimeAnalyticsService implements TimeTrackingAnalytics {
  async generateProductivityReport(employeeId: string, period: DateRange): Promise<ProductivityReport> {
    const timeEntries = await this.getTimeEntries(employeeId, period);
    const projectAllocations = await this.getProjectAllocations(employeeId, period);
    
    // Calculate time distribution
    const timeDistribution = this.calculateTimeDistribution(timeEntries);
    
    // Calculate productivity metrics
    const productivityMetrics = this.calculateProductivityMetrics(timeEntries, period);
    
    // Analyze project breakdown
    const projectBreakdown = this.analyzeProjectBreakdown(timeEntries, projectAllocations);
    
    // Identify trends and patterns
    const trends = this.analyzeTrends(timeEntries, period);
    
    return {
      employee_id: employeeId,
      reporting_period: period,
      time_distribution: timeDistribution,
      productivity_metrics: productivityMetrics,
      project_breakdown: projectBreakdown,
      trends
    };
  }
  
  async analyzeProjectTimeSpent(projectId: string): Promise<ProjectTimeAnalysis> {
    const project = await this.getProject(projectId);
    const tasks = await this.getProjectTasks(projectId);
    const timeEntries = await this.getProjectTimeEntries(projectId);
    const budget = await this.getProjectBudget(projectId);
    
    // Analyze time vs. estimates
    const taskAnalysis = tasks.map(task => {
      const taskTimeEntries = timeEntries.filter(entry => entry.task_id === task.id);
      const actualHours = taskTimeEntries.reduce((sum, entry) => sum + entry.duration_hours, 0);
      const estimatedHours = task.estimated_effort_hours || 0;
      
      return {
        task_id: task.id,
        task_name: task.task_name,
        estimated_hours: estimatedHours,
        actual_hours: actualHours,
        variance_hours: actualHours - estimatedHours,
        variance_percentage: estimatedHours > 0 ? ((actualHours - estimatedHours) / estimatedHours) * 100 : 0,
        completion_status: task.status
      };
    });
    
    // Calculate project-level metrics
    const totalEstimatedHours = tasks.reduce((sum, task) => sum + (task.estimated_effort_hours || 0), 0);
    const totalActualHours = timeEntries.reduce((sum, entry) => sum + entry.duration_hours, 0);
    const totalBillableHours = timeEntries.filter(entry => entry.is_billable).reduce((sum, entry) => sum + entry.duration_hours, 0);
    
    return {
      project_id: projectId,
      project_name: project.project_name,
      analysis_date: new Date(),
      
      time_summary: {
        total_estimated_hours: totalEstimatedHours,
        total_actual_hours: totalActualHours,
        total_billable_hours: totalBillableHours,
        variance_hours: totalActualHours - totalEstimatedHours,
        variance_percentage: totalEstimatedHours > 0 ? ((totalActualHours - totalEstimatedHours) / totalEstimatedHours) * 100 : 0
      },
      
      task_analysis: taskAnalysis,
      
      team_performance: await this.analyzeTeamPerformance(projectId, timeEntries),
      
      efficiency_metrics: {
        estimation_accuracy: this.calculateEstimationAccuracy(taskAnalysis),
        productivity_index: this.calculateProductivityIndex(timeEntries, tasks),
        billable_ratio: totalActualHours > 0 ? (totalBillableHours / totalActualHours) * 100 : 0
      }
    };
  }
}
```

##  Project Financial Management

### Budget Tracking and Cost Control

```sql
-- Project budgets
CREATE TABLE project_budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Budget identification
    budget_name VARCHAR(255) NOT NULL,
    budget_type VARCHAR(20) DEFAULT 'baseline', -- baseline, current, forecast, approved
    budget_version INTEGER DEFAULT 1,
    
    -- Budget period
    budget_start_date DATE NOT NULL,
    budget_end_date DATE NOT NULL,
    
    -- Budget categories
    labor_budget DECIMAL(15,2) DEFAULT 0,
    material_budget DECIMAL(15,2) DEFAULT 0,
    equipment_budget DECIMAL(15,2) DEFAULT 0,
    travel_budget DECIMAL(15,2) DEFAULT 0,
    other_budget DECIMAL(15,2) DEFAULT 0,
    contingency_budget DECIMAL(15,2) DEFAULT 0,
    total_budget DECIMAL(15,2) DEFAULT 0,
    
    -- Budget status
    status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, approved, active, closed
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    
    -- Budget notes
    assumptions TEXT,
    notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_budget_type CHECK (budget_type IN ('baseline', 'current', 'forecast', 'approved')),
    CONSTRAINT valid_budget_status CHECK (status IN ('draft', 'submitted', 'approved', 'active', 'closed'))
);

-- Project cost tracking
CREATE TABLE project_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Cost identification
    cost_date DATE NOT NULL,
    cost_type VARCHAR(50) NOT NULL, -- labor, material, equipment, travel, subcontractor, other
    cost_category VARCHAR(100),
    
    -- Cost details
    description TEXT NOT NULL,
    quantity DECIMAL(12,4) DEFAULT 1,
    unit_cost DECIMAL(12,4),
    total_cost DECIMAL(15,2) NOT NULL,
    
    -- Cost classification
    is_billable BOOLEAN DEFAULT false,
    is_reimbursable BOOLEAN DEFAULT false,
    markup_percentage DECIMAL(5,2) DEFAULT 0,
    billable_amount DECIMAL(15,2),
    
    -- Source references
    employee_id UUID REFERENCES employees(id),
    vendor_id UUID REFERENCES vendors(id),
    purchase_order_id UUID REFERENCES purchase_orders(id),
    expense_report_id UUID REFERENCES expense_reports(id),
    time_entry_id UUID REFERENCES project_time_entries(id),
    
    -- Approval and status
    status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, approved, rejected, invoiced
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    
    -- GL integration
    posted_to_gl BOOLEAN DEFAULT false,
    journal_entry_id UUID REFERENCES journal_entries(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_cost_type CHECK (cost_type IN ('labor', 'material', 'equipment', 'travel', 'subcontractor', 'other')),
    CONSTRAINT valid_cost_status CHECK (status IN ('draft', 'submitted', 'approved', 'rejected', 'invoiced')),
    CONSTRAINT positive_cost CHECK (total_cost >= 0)
);

-- Project invoicing and billing
CREATE TABLE project_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Invoice identification
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    
    -- Invoice period
    billing_period_start DATE NOT NULL,
    billing_period_end DATE NOT NULL,
    invoice_date DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date DATE NOT NULL,
    
    -- Invoice amounts
    labor_amount DECIMAL(15,2) DEFAULT 0,
    expense_amount DECIMAL(15,2) DEFAULT 0,
    material_amount DECIMAL(15,2) DEFAULT 0,
    subtotal DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    
    -- Billing details
    billing_type VARCHAR(20) NOT NULL, -- time_and_material, fixed_price, milestone, progress
    milestone_percentage DECIMAL(5,2), -- For milestone billing
    progress_percentage DECIMAL(5,2), -- For progress billing
    
    -- Invoice status
    status VARCHAR(20) DEFAULT 'draft', -- draft, sent, paid, cancelled
    sent_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    
    -- Payment information
    payment_terms VARCHAR(100),
    payment_method VARCHAR(50),
    payment_reference VARCHAR(100),
    
    -- Invoice content
    description TEXT,
    notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_billing_type CHECK (billing_type IN ('time_and_material', 'fixed_price', 'milestone', 'progress')),
    CONSTRAINT valid_invoice_status CHECK (status IN ('draft', 'sent', 'paid', 'cancelled'))
);
```

### Earned Value Management (EVM)

```typescript
interface EarnedValueMetrics {
  project_id: string;
  measurement_date: Date;
  
  // Basic EVM values
  planned_value: number;        // PV - Budgeted cost of work scheduled
  earned_value: number;         // EV - Budgeted cost of work performed
  actual_cost: number;          // AC - Actual cost of work performed
  budget_at_completion: number; // BAC - Total project budget
  
  // Variance analysis
  schedule_variance: number;    // SV = EV - PV
  cost_variance: number;        // CV = EV - AC
  variance_at_completion: number; // VAC = BAC - EAC
  
  // Performance indices
  schedule_performance_index: number; // SPI = EV / PV
  cost_performance_index: number;     // CPI = EV / AC
  
  // Forecasting
  estimate_at_completion: number;     // EAC - Forecasted total cost
  estimate_to_complete: number;       // ETC - Remaining cost forecast
  to_complete_performance_index: number; // TCPI = (BAC - EV) / (BAC - AC)
  
  // Time forecasting
  estimated_completion_date: Date;
  schedule_variance_days: number;
}

class EarnedValueService {
  async calculateEarnedValue(projectId: string, measurementDate: Date = new Date()): Promise<EarnedValueMetrics> {
    const project = await this.getProject(projectId);
    const tasks = await this.getProjectTasks(projectId);
    const actualCosts = await this.getActualCosts(projectId, measurementDate);
    
    // Calculate Planned Value (PV)
    const plannedValue = this.calculatePlannedValue(tasks, measurementDate);
    
    // Calculate Earned Value (EV)
    const earnedValue = this.calculateEarnedValue(tasks, measurementDate);
    
    // Calculate Actual Cost (AC)
    const actualCost = actualCosts.reduce((sum, cost) => sum + cost.total_cost, 0);
    
    // Budget at Completion (BAC)
    const budgetAtCompletion = project.approved_budget || 0;
    
    // Calculate variances
    const scheduleVariance = earnedValue - plannedValue;
    const costVariance = earnedValue - actualCost;
    
    // Calculate performance indices
    const schedulePerformanceIndex = plannedValue > 0 ? earnedValue / plannedValue : 0;
    const costPerformanceIndex = actualCost > 0 ? earnedValue / actualCost : 0;
    
    // Forecast Estimate at Completion (EAC)
    let estimateAtCompletion: number;
    if (costPerformanceIndex > 0) {
      // EAC = BAC / CPI (assumes current performance continues)
      estimateAtCompletion = budgetAtCompletion / costPerformanceIndex;
    } else {
      estimateAtCompletion = budgetAtCompletion;
    }
    
    // Calculate other forecasting metrics
    const estimateToComplete = Math.max(0, estimateAtCompletion - actualCost);
    const varianceAtCompletion = budgetAtCompletion - estimateAtCompletion;
    
    // To Complete Performance Index
    const remainingWork = budgetAtCompletion - earnedValue;
    const remainingBudget = budgetAtCompletion - actualCost;
    const toCompletePerformanceIndex = remainingBudget > 0 ? remainingWork / remainingBudget : 0;
    
    // Time forecasting
    const { estimatedCompletionDate, scheduleVarianceDays } = this.forecastSchedule(
      project,
      schedulePerformanceIndex,
      measurementDate
    );
    
    return {
      project_id: projectId,
      measurement_date: measurementDate,
      planned_value: plannedValue,
      earned_value: earnedValue,
      actual_cost: actualCost,
      budget_at_completion: budgetAtCompletion,
      schedule_variance: scheduleVariance,
      cost_variance: costVariance,
      variance_at_completion: varianceAtCompletion,
      schedule_performance_index: schedulePerformanceIndex,
      cost_performance_index: costPerformanceIndex,
      estimate_at_completion: estimateAtCompletion,
      estimate_to_complete: estimateToComplete,
      to_complete_performance_index: toCompletePerformanceIndex,
      estimated_completion_date: estimatedCompletionDate,
      schedule_variance_days: scheduleVarianceDays
    };
  }
  
  private calculatePlannedValue(tasks: ProjectTask[], measurementDate: Date): number {
    let plannedValue = 0;
    
    for (const task of tasks) {
      if (task.planned_start_date && task.planned_end_date && task.budgeted_cost) {
        const taskStart = new Date(task.planned_start_date);
        const taskEnd = new Date(task.planned_end_date);
        
        if (measurementDate >= taskEnd) {
          // Task should be complete - include full budgeted cost
          plannedValue += task.budgeted_cost;
        } else if (measurementDate >= taskStart) {
          // Task is in progress - calculate proportional value
          const totalDays = (taskEnd.getTime() - taskStart.getTime()) / (1000 * 60 * 60 * 24);
          const elapsedDays = (measurementDate.getTime() - taskStart.getTime()) / (1000 * 60 * 60 * 24);
          const progressPercentage = Math.min(100, (elapsedDays / totalDays) * 100);
          
          plannedValue += (task.budgeted_cost * progressPercentage) / 100;
        }
        // If measurementDate < taskStart, no planned value earned yet
      }
    }
    
    return plannedValue;
  }
  
  private calculateEarnedValue(tasks: ProjectTask[], measurementDate: Date): number {
    let earnedValue = 0;
    
    for (const task of tasks) {
      if (task.budgeted_cost && task.progress_percentage) {
        // Only count progress up to the measurement date
        if (task.actual_start_date && new Date(task.actual_start_date) <= measurementDate) {
          earnedValue += (task.budgeted_cost * task.progress_percentage) / 100;
        }
      }
    }
    
    return earnedValue;
  }
  
  private forecastSchedule(
    project: Project, 
    spi: number, 
    measurementDate: Date
  ): { estimatedCompletionDate: Date; scheduleVarianceDays: number } {
    const originalDuration = new Date(project.planned_end_date).getTime() - new Date(project.planned_start_date).getTime();
    const remainingDuration = new Date(project.planned_end_date).getTime() - measurementDate.getTime();
    
    let estimatedCompletionDate: Date;
    if (spi > 0) {
      // Adjust remaining duration based on schedule performance
      const adjustedRemainingDuration = remainingDuration / spi;
      estimatedCompletionDate = new Date(measurementDate.getTime() + adjustedRemainingDuration);
    } else {
      estimatedCompletionDate = new Date(project.planned_end_date);
    }
    
    const scheduleVarianceDays = (estimatedCompletionDate.getTime() - new Date(project.planned_end_date).getTime()) / (1000 * 60 * 60 * 24);
    
    return { estimatedCompletionDate, scheduleVarianceDays };
  }
}
```

This  project management system provides robust planning, execution, and monitoring capabilities with advanced resource optimization, time tracking, and financial management features suitable for various project methodologies and organizational needs.
