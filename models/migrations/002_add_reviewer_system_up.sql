-- 002_add_reviewer_system_up.sql
-- Migration to add reviewer functionality to the recruitment system

-- Add reviewer role to the existing user_role enum
ALTER TYPE user_role ADD VALUE 'reviewer';

-- Add department column to users table for reviewers
ALTER TABLE users ADD COLUMN department department;

-- Create reviews table to store application reviews
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    reviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    department department NOT NULL,
    shortlisted BOOLEAN NOT NULL DEFAULT FALSE,
    comments TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(application_id, reviewer_id)
);

-- Create trigger for reviews updated_at
CREATE TRIGGER update_reviews_updated_at
BEFORE UPDATE ON reviews
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Add constraint to ensure reviewers have a department assigned
ALTER TABLE users ADD CONSTRAINT check_reviewer_department 
CHECK (
    (role = 'reviewer' AND department IS NOT NULL) OR 
    (role = 'applicant' AND department IS NULL)
);

-- Create index on reviews for better query performance
CREATE INDEX idx_reviews_application_id ON reviews(application_id);
CREATE INDEX idx_reviews_reviewer_id ON reviews(reviewer_id);
CREATE INDEX idx_reviews_department ON reviews(department);
CREATE INDEX idx_users_department ON users(department) WHERE department IS NOT NULL;
