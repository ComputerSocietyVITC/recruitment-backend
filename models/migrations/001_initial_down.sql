-- 001_drop_users_table.sql

DROP TRIGGER IF EXISTS update_answers_updated_at ON answers;

DROP TRIGGER IF EXISTS update_applications_updated_at ON applications;

DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS answers;

DROP TABLE IF EXISTS applications;

DROP TABLE IF EXISTS questions;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS department;

DROP TYPE IF EXISTS user_role;

DROP EXTENSION IF EXISTS "uuid-ossp";
