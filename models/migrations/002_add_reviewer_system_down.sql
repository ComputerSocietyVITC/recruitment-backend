-- 002_add_reviewer_system_down.sql
-- Rollback migration for reviewer functionality

-- Drop indexes
DROP INDEX IF EXISTS idx_users_department;
DROP INDEX IF EXISTS idx_reviews_department;
DROP INDEX IF EXISTS idx_reviews_reviewer_id;
DROP INDEX IF EXISTS idx_reviews_application_id;

-- Drop constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_reviewer_department;

-- Drop reviews table
DROP TABLE IF EXISTS reviews;

-- Remove department column from users table
ALTER TABLE users DROP COLUMN IF EXISTS department;

-- Note: Cannot remove enum value 'reviewer' from user_role as PostgreSQL doesn't support this operation
-- If rollback is needed, you would need to:
-- 1. Update all reviewer users to 'applicant' role
-- 2. Create a new enum without 'reviewer'
-- 3. Alter the column to use the new enum
-- 4. Drop the old enum
