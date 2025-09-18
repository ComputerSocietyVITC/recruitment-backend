-- Example script to create reviewer users
-- Run this after applying the 002_add_reviewer_system_up.sql migration

-- Example: Create a technical reviewer
INSERT INTO users (full_name, email, verified, reg_num, hashed_password, role, department)
VALUES (
    'Technical Reviewer',
    'tech.reviewer@example.com',
    true,
    'REV001',
    '$2a$10$Q8Ltxi7JDz.VJydOo1d73eorls8XOL1OihDfSMwiZo.mJ0fNip.1C', -- password: "password123"
    'reviewer',
    'technical'
);

-- Example: Create a design reviewer
INSERT INTO users (full_name, email, verified, reg_num, hashed_password, role, department)
VALUES (
    'Design Reviewer',
    'design.reviewer@example.com',
    true,
    'REV002',
    '$2a$10$Q8Ltxi7JDz.VJydOo1d73eorls8XOL1OihDfSMwiZo.mJ0fNip.1C', -- password: "password123"
    'reviewer',
    'design'
);

-- Example: Create a management reviewer
INSERT INTO users (full_name, email, verified, reg_num, hashed_password, role, department)
VALUES (
    'Management Reviewer',
    'mgmt.reviewer@example.com',
    true,
    'REV003',
    '$2a$10$Q8Ltxi7JDz.VJydOo1d73eorls8XOL1OihDfSMwiZo.mJ0fNip.1C', -- password: "password123"
    'reviewer',
    'management'
);

-- Example: Create a social reviewer
INSERT INTO users (full_name, email, verified, reg_num, hashed_password, role, department)
VALUES (
    'Social Reviewer',
    'social.reviewer@example.com',
    true,
    'REV004',
    '$2a$10$Q8Ltxi7JDz.VJydOo1d73eorls8XOL1OihDfSMwiZo.mJ0fNip.1C', -- password: "password123"
    'reviewer',
    'social'
);

-- Note: The hashed password above corresponds to "password123"
-- In production, use a strong password and hash it properly with bcrypt
