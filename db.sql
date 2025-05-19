-- Permissions Table
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL UNIQUE
);

-- Roles Table
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL UNIQUE,
    permissions JSONB DEFAULT '[]' -- Optional: store permission IDs as JSONB array
);

-- Roles-Permissions Junction Table (for many-to-many relationship)
CREATE TABLE role_permissions (
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Users Table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    role_id INTEGER REFERENCES roles(id) ON DELETE SET NULL
);

-- PTO Types Table
CREATE TABLE pto_types (
    id SERIAL PRIMARY KEY,
    title VARCHAR(50) NOT NULL UNIQUE, -- e.g., 'sick_leave', 'vacation_leave', 'undertime'
    is_counted BOOLEAN NOT NULL DEFAULT TRUE -- TRUE for types with balance tracking (sick/vacation), FALSE for undertime
);

-- PTO Balances Table
CREATE TABLE pto_balances (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pto_type_id INTEGER NOT NULL REFERENCES pto_types(id) ON DELETE RESTRICT,
    balance DECIMAL(10, 2) NOT NULL DEFAULT 0.0, -- Balance in days (or hours for undertime)
    CONSTRAINT unique_user_pto_type UNIQUE (user_id, pto_type_id),
    CONSTRAINT positive_balance CHECK (balance >= 0)
);

-- PTO Requests Table
CREATE TABLE pto_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pto_type_id INTEGER NOT NULL REFERENCES pto_types(id) ON DELETE RESTRICT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    hours DECIMAL(10, 2), -- For undertime or partial days
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- e.g., 'pending', 'approved', 'denied'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_date_range CHECK (end_date >= start_date),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'approved', 'denied'))
);

-- Add indexes for better performance
CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_pto_balances_user_id ON pto_balances(user_id);
CREATE INDEX idx_pto_balances_pto_type_id ON pto_balances(pto_type_id);
CREATE INDEX idx_pto_requests_user_id ON pto_requests(user_id);
CREATE INDEX idx_pto_requests_pto_type_id ON pto_requests(pto_type_id);
CREATE INDEX idx_pto_requests_status ON pto_requests(status);


INSERT INTO role_permissions (role_id, permission_id) 
VALUES 
    (1, 3), 
    (2, 3),
    (3, 2),
    (4, 3),
    (5, 3),
    (6, 2),
    (7, 2),
    (8, 3),
    (9, 3),
    (10, 2);


-- Insert users
INSERT INTO users (name, email, role_id) 
VALUES 
    ('Avegail Serote-Dayuja', 'avegail@goodlygrowth.com', 2), -- admin
    ('Brian Espina', 'brian@goodlygrowth.com', 4),
    ('John Reynald Tubay', 'john@goodlygrowth.com', 5),
    ('Kenneth Romero', 'keneth@goodlygrowth.com', 3);

INSERT INTO pto_balances (user_id, pto_type_id, balance) 
VALUES 
    (1, 2, 13.0), -- 10 days vacation leave
    (1, 1, 4.0), -- 15 days sick leave
    (2, 2, 18.0), -- 10 days vacation leave
    (2, 1, 4.0), -- 15 days sick leave
    (3, 2, 20.0), -- 10 days vacation leave
    (3, 1, 5.0), -- 15 days sick leave
    (4, 2, 15.0), -- 10 days vacation leave
    (4, 1, 4.0);


SELECT 
    pr.id AS request_id,
    u.id AS user_id,
    u.name AS user_name,
    u.email AS user_email,
    r_managed.title AS user_role,
    pt.title AS pto_type,
    pr.start_date,
    pr.end_date,
    pr.days AS days,
                                      pr.status,
                                        pr.created_at
                                          FROM users u_manager
                                          JOIN roles r_manager ON u_manager.role_id = r_manager.id
                                          JOIN role_management rm ON r_manager.id = rm.manager_role_id
                                          JOIN roles r_managed ON rm.managed_role_id = r_managed.id
                                          JOIN users u ON u.role_id = r_managed.id
                                          JOIN pto_requests pr ON u.id = pr.user_id
                                          JOIN pto_types pt ON pr.pto_type_id = pt.id
                                          WHERE u_manager.id = 2
                                          ORDER BY pr.created_at DESC;
