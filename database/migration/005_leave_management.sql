CREATE TABLE IF NOT EXISTS leave_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL, -- "Annual Leave", "Sick Leave"
    code VARCHAR(50),
    description TEXT,
    days_allowed DECIMAL(5,1) NOT NULL DEFAULT 0, 
    is_paid BOOLEAN DEFAULT true,
    requires_documentation BOOLEAN DEFAULT false, -- Medical cert for sick leave
    carry_forward_allowed BOOLEAN DEFAULT false,
    max_carry_forward_days DECIMAL(5,1) DEFAULT 0,
    color_code VARCHAR(10) DEFAULT '#3498db',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    UNIQUE(company_id, name)
);

CREATE INDEX idx_leave_types_company ON leave_types(company_id);

CREATE TABLE IF NOT EXISTS leave_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    leave_type_id UUID NOT NULL,
    year INTEGER NOT NULL,
    total_days DECIMAL(6,1) NOT NULL DEFAULT 0,
    used_days DECIMAL(6,1) NOT NULL DEFAULT 0,
    pending_days DECIMAL(6,1) NOT NULL DEFAULT 0, -- Awaiting approval
    carried_forward_days DECIMAL(6,1) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_id, leave_type_id, year),
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (leave_type_id) REFERENCES leave_types(id) ON DELETE CASCADE
);

CREATE INDEX idx_leave_balances_employee ON leave_balances(employee_id);

CREATE TABLE IF NOT EXISTS leave_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    leave_type_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    days_requested DECIMAL(5,1) NOT NULL,
    reason TEXT,
    attachment_url TEXT, -- Medical certificates
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled', 'withdrawn')),
    
    
    current_step INTEGER DEFAULT 1,
    approved_by UUID, -- Final approver
    approved_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (leave_type_id) REFERENCES leave_types(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES employees(id) ON DELETE SET NULL,
    CONSTRAINT valid_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_leave_requests_employee ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
