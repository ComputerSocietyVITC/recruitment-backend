-- Rollback migration: 000002_add_regnum_to_users
-- This script reverses the addition of reg_num column to users table

-- Drop the unique constraint first
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_reg_num_unique;

-- Drop the reg_num column
ALTER TABLE users DROP COLUMN IF EXISTS reg_num;
