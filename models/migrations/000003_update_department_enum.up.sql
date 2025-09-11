-- Migration: 000003_update_department_enum
-- This script updates the department enum to remove 'marketing' and add 'design', keeping technical, management, social_media, design

-- First, we need to create a new enum type with the desired values
CREATE TYPE department_new AS ENUM ('technical', 'management', 'social_media', 'design');

-- Update any existing records that have 'marketing' department to 'management' (or another appropriate department)
-- You may want to adjust this based on your business logic
UPDATE questions SET department = 'management'::department WHERE department = 'marketing';
UPDATE applications SET department = 'management'::department WHERE department = 'marketing';

-- Alter the tables to use the new enum type
ALTER TABLE questions ALTER COLUMN department TYPE department_new USING department::text::department_new;
ALTER TABLE applications ALTER COLUMN department TYPE department_new USING department::text::department_new;

-- Drop the old enum type
DROP TYPE department;

-- Rename the new enum type to the original name
ALTER TYPE department_new RENAME TO department;
