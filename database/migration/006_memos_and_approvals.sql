CREATE TABLE IF NOT EXISTS memos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    employee_id UUID NOT NULL, -- Sender/Requester
    memo_type VARCHAR(50) NOT NULL CHECK (memo_type IN ('request', 'disciplinary', 'announcement', 'general')),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    reference_number VARCHAR(100),
    
    
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('draft', 'pending', 'approved', 'rejected', 'archived')),
    current_step INTEGER DEFAULT 1,
    priority VARCHAR(20) DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE
);

CREATE INDEX idx_memos_company ON memos(company_id);
CREATE INDEX idx_memos_status ON memos(status);


CREATE TABLE IF NOT EXISTS memo_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memo_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (memo_id) REFERENCES memos(id) ON DELETE CASCADE,
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    UNIQUE(memo_id, employee_id)
);

CREATE TABLE IF NOT EXISTS approval_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    workflow_type VARCHAR(50) NOT NULL CHECK (workflow_type IN ('leave', 'memo', 'expense')), 
    department_id UUID NULL, 
    steps JSONB NOT NULL, -- [{"step": 1, "role_id": "uuid", "approver_id": null}, {"step": 2, "role_id": "uuid"}]
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS approval_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('leave_request', 'memo')),
    entity_id UUID NOT NULL,
    step_number INTEGER NOT NULL,
    approver_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL CHECK (action IN ('approved', 'rejected', 'requested_changes')),
    comments TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (approver_id) REFERENCES employees(id) ON DELETE CASCADE
);

CREATE INDEX idx_approval_history_entity ON approval_history(entity_type, entity_id);
