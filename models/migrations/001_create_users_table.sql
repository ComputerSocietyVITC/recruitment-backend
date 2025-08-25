-- 001_create_users_table.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Define user roles enum
CREATE TYPE user_role AS ENUM ('applicant', 'evaluator', 'admin', 'super_admin');

-- Will be extended as needed in the future
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone_number TEXT,
    hashed_password TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'applicant',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Optional: Create a super admin user (update the email/password as needed)
INSERT INTO users (id, full_name, email, phone_number, hashed_password, role, created_at, updated_at)
VALUES (
    uuid_generate_v4(),
    'Root',
    'admin@comp.socks',
    '+91 9898888110',
    '$2a$10$WDlCHHvtqDcP9IbauQ0XPelvwU4t7Qaf9Gm9eb2cgXtX4M.oDUVJi',
    'super_admin',
    NOW(),
    NOW()
)