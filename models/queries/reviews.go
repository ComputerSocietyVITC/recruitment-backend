package queries

// Review-related SQL queries

// CreateReviewQuery creates or updates a review (upsert)
const CreateReviewQuery = `
INSERT INTO reviews (application_id, reviewer_id, department, shortlisted, comments)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (application_id, reviewer_id) 
DO UPDATE SET 
    shortlisted = EXCLUDED.shortlisted,
    comments = EXCLUDED.comments,
    updated_at = NOW()
RETURNING id, application_id, reviewer_id, department, shortlisted, comments, created_at, updated_at`

// GetReviewByIDQuery gets a review by its ID
const GetReviewByIDQuery = `
SELECT id, application_id, reviewer_id, department, shortlisted, comments, created_at, updated_at
FROM reviews 
WHERE id = $1`

// GetReviewByApplicationAndReviewerQuery gets a review by application and reviewer
const GetReviewByApplicationAndReviewerQuery = `
SELECT id, application_id, reviewer_id, department, shortlisted, comments, created_at, updated_at
FROM reviews 
WHERE application_id = $1 AND reviewer_id = $2`

// GetApplicationsForReviewerQuery gets all submitted applications for a specific department
const GetApplicationsForReviewerQuery = `
SELECT 
    a.id, a.user_id, a.department, a.submitted, a.chickened_out, a.created_at, a.updated_at,
    u.full_name as user_name, u.email as user_email,
    r.id as review_id, r.shortlisted, r.comments as review_comments, 
    r.reviewer_id, ru.full_name as reviewer_name, r.updated_at as reviewed_at
FROM applications a
JOIN users u ON a.user_id = u.id
LEFT JOIN reviews r ON a.id = r.application_id
LEFT JOIN users ru ON r.reviewer_id = ru.id
WHERE a.department = $1 AND a.submitted = true AND a.chickened_out = false
ORDER BY a.created_at DESC`

// GetApplicationsForReviewerWithPaginationQuery gets applications with pagination
const GetApplicationsForReviewerWithPaginationQuery = `
SELECT 
    a.id, a.user_id, a.department, a.submitted, a.chickened_out, a.created_at, a.updated_at,
    u.full_name as user_name, u.email as user_email,
    r.id as review_id, r.shortlisted, r.comments as review_comments, 
    r.reviewer_id, ru.full_name as reviewer_name, r.updated_at as reviewed_at
FROM applications a
JOIN users u ON a.user_id = u.id
LEFT JOIN reviews r ON a.id = r.application_id
LEFT JOIN users ru ON r.reviewer_id = ru.id
WHERE a.department = $1 AND a.submitted = true AND a.chickened_out = false
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3`

// CountApplicationsForReviewerQuery counts total applications for a department
const CountApplicationsForReviewerQuery = `
SELECT COUNT(*)
FROM applications a
WHERE a.department = $1 AND a.submitted = true AND a.chickened_out = false`

// GetReviewStatsQuery gets review statistics for a department
const GetReviewStatsQuery = `
SELECT 
    $1::text as department,
    COUNT(*) as total_applications,
    COUNT(r.id) as reviewed_count,
    COUNT(CASE WHEN r.shortlisted = true THEN 1 END) as shortlisted_count,
    COUNT(CASE WHEN r.shortlisted = false THEN 1 END) as rejected_count,
    COUNT(*) - COUNT(r.id) as pending_count
FROM applications a
LEFT JOIN reviews r ON a.id = r.application_id
WHERE a.department = $1 AND a.submitted = true AND a.chickened_out = false`

// GetReviewsByReviewerQuery gets all reviews by a specific reviewer
const GetReviewsByReviewerQuery = `
SELECT 
    r.id, r.application_id, r.reviewer_id, r.department, r.shortlisted, r.comments, r.created_at, r.updated_at,
    u.full_name as application_user_name, u.email as application_user_email,
    ru.full_name as reviewer_name, ru.email as reviewer_email
FROM reviews r
JOIN applications a ON r.application_id = a.id
JOIN users u ON a.user_id = u.id
JOIN users ru ON r.reviewer_id = ru.id
WHERE r.reviewer_id = $1
ORDER BY r.updated_at DESC`

// DeleteReviewQuery deletes a review
const DeleteReviewQuery = `
DELETE FROM reviews 
WHERE id = $1 AND reviewer_id = $2
RETURNING id`

// GetApplicationWithReviewQuery gets an application with its review details
const GetApplicationWithReviewQuery = `
SELECT 
    a.id, a.user_id, a.department, a.submitted, a.chickened_out, a.created_at, a.updated_at,
    u.full_name as user_name, u.email as user_email,
    r.id as review_id, r.shortlisted, r.comments as review_comments, 
    r.reviewer_id, ru.full_name as reviewer_name, r.updated_at as reviewed_at
FROM applications a
JOIN users u ON a.user_id = u.id
LEFT JOIN reviews r ON a.id = r.application_id
LEFT JOIN users ru ON r.reviewer_id = ru.id
WHERE a.id = $1`
