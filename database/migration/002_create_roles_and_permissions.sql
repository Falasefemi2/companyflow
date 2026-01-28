CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NULL,    name VARCHAR(50) NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT false,
    permissions_cache JSONB DEFAULT '[]', 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (company_id, name),
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
);

CREATE INDEX idx_roles_company ON roles(company_id);

INSERT INTO roles (name, description, is_system_role, permissions_cache) VALUES 
    ('Super Admin', 'Full company access', true, '["all"]'),
    ('HR Manager', 'HR administration', true, '["hr_full"]'),
    ('Manager', 'Team management', true, '["team_management"]'),
    ('Employee', 'Standard access', true, '["self_service"]');

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL CHECK (action IN ('create', 'read', 'update', 'delete', 'approve', 'reject', 'manage')),
    resource VARCHAR(100) NOT NULL, -- 'employees', 'leaves', 'company_settings'
    conditions JSONB DEFAULT '{}', -- {"department": "own", "level": "subordinate"}
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE INDEX idx_permissions_role ON permissions(role_id);
