package queries

const GetAllApplicationsQuery = `
SELECT id, user_id, department, submitted, chickened_out, created_at, updated_at
FROM applications 
ORDER BY created_at DESC
`

const CreateApplicationQuery = `
INSERT INTO applications (id, user_id, department, submitted, chickened_out, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, department, submitted, chickened_out, created_at, updated_at
`

const GetUserApplicationsQuery = `
SELECT id, user_id, department, submitted, chickened_out, created_at, updated_at
FROM applications 
WHERE user_id = $1
ORDER BY created_at DESC
`

const GetApplicationByIDQuery = `
SELECT id, user_id, department, submitted, chickened_out, created_at, updated_at
FROM applications 
WHERE id = $1
`

const SubmitApplicationQuery = `
UPDATE applications 
SET submitted = true, updated_at = $2
WHERE id = $1 AND user_id = $3
RETURNING id, user_id, department, submitted, chickened_out, created_at, updated_at
`

const DeleteApplicationQuery = `
DELETE FROM applications 
WHERE id = $1 AND user_id = $2 AND submitted = false
`

const CountUserApplicationsQuery = `
SELECT COUNT(*) FROM applications WHERE user_id = $1
`

const UpdateApplicationChickenedOutStatusQuery = `
UPDATE applications
SET chickened_out = $3, updated_at = NOW()
WHERE id = $1 AND user_id = $2
`
