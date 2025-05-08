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
