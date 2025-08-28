-- 001_create_users_table.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Define user roles enum
CREATE TYPE user_role AS ENUM ('applicant', 'evaluator', 'admin', 'super_admin');

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    phone_number TEXT NOT NULL,
    hashed_password TEXT NOT NULL,
    reset_token TEXT,
    reset_token_expires_at TIMESTAMP WITH TIME ZONE,
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

-- Define department enum for questions
CREATE TYPE department AS ENUM ('technical', 'marketing', 'management', 'social_media');

-- Create questions table
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    department department NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create applications table
CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    department department NOT NULL,
    submitted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for applications updated_at
CREATE TRIGGER update_applications_updated_at
BEFORE UPDATE ON applications
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create answers table
CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for answers updated_at
CREATE TRIGGER update_answers_updated_at
BEFORE UPDATE ON answers
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Optional: Create a super admin user (update the email/password as needed)
INSERT INTO users (id, full_name, email, phone_number, hashed_password, role, created_at, updated_at)
VALUES (
    uuid_generate_v4(),
    'Root',
    'admin@comp.socks',
    '+91 9898888110',
    '$2a$10$Q8Ltxi7JDz.VJydOo1d73eorls8XOL1OihDfSMwiZo.mJ0fNip.1C',
    'super_admin',
    NOW(),
    NOW()
)