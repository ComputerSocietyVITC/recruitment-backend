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
INSERT INTO questions (id, department, body, created_at) VALUES
-- Technical questions
('11111111-1111-1111-1111-111111111111', 'technical', 'What is your full name?', NOW()),
('22222222-2222-2222-2222-222222222222', 'technical', 'Why do you want to join our team?', NOW()),
('33333333-3333-3333-3333-333333333333', 'technical', 'What programming languages do you know?', NOW()),
('44444444-4444-4444-4444-444444444444', 'technical', 'Describe a project you built recently', NOW()),

-- Marketing questions
('55555555-5555-5555-5555-555555555555', 'marketing', 'What marketing strategies are you familiar with?', NOW()),
('66666666-6666-6666-6666-666666666666', 'marketing', 'How would you promote our brand?', NOW()),

-- Management questions  
('77777777-7777-7777-7777-777777777777', 'management', 'How do you handle team conflicts?', NOW()),
('88888888-8888-8888-8888-888888888888', 'management', 'Describe your leadership style', NOW()),

-- Social media questions
('99999999-9999-9999-9999-999999999999', 'social_media', 'Which social media platforms are you experienced with?', NOW()),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'social_media', 'How would you grow our social media presence?', NOW());


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