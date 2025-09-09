-- Rollback migration: 000001_initial
-- This script reverses all changes made in the up migration

-- Drop indexes
DROP INDEX IF EXISTS idx_questions_department;
DROP INDEX IF EXISTS idx_answers_question_id;
DROP INDEX IF EXISTS idx_answers_application_id;
DROP INDEX IF EXISTS idx_applications_department;
DROP INDEX IF EXISTS idx_applications_user_id;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;

-- Drop triggers (in reverse order of creation)
DROP TRIGGER IF EXISTS update_answers_updated_at ON answers;
DROP TRIGGER IF EXISTS update_applications_updated_at ON applications;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order of creation, respecting foreign key dependencies)
DROP TABLE IF EXISTS answers;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS users;

-- Drop custom types
DROP TYPE IF EXISTS department;
DROP TYPE IF EXISTS user_role;

-- Drop extension (be careful with this in production - other schemas might use it)
DROP EXTENSION IF EXISTS "uuid-ossp";
