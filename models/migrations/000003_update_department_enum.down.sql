-- Rollback migration: 000003_update_department_enum
-- This script reverses the department enum changes, restoring the 'marketing' option and removing 'design'

-- Create a new enum type with the original values (including 'marketing', excluding 'design')
CREATE TYPE department_new AS ENUM ('technical', 'marketing', 'management', 'social_media');

-- Alter the tables to use the restored enum type
ALTER TABLE questions ALTER COLUMN department TYPE department_new USING department::text::department_new;
ALTER TABLE applications ALTER COLUMN department TYPE department_new USING department::text::department_new;

-- Drop the current enum type
DROP TYPE department;

-- Rename the new enum type to the original name
ALTER TYPE department_new RENAME TO department;
