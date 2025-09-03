package queries

const GetAllApplicationsQuery = `
SELECT id, user_id, department, submitted, created_at, updated_at
FROM applications 
ORDER BY created_at DESC
`

const CreateApplicationQuery = `
INSERT INTO applications (id, user_id, department, submitted, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, department, submitted, created_at, updated_at
`

const GetUserApplicationsQuery = `
SELECT id, user_id, department, submitted, created_at, updated_at
FROM applications 
WHERE user_id = $1
ORDER BY created_at DESC
`
