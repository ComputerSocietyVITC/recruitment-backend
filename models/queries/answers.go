package queries

const SaveAnswersQuery = `
INSERT INTO answers (id, application_id,user_id, question_id, body, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

const GetApplicationAnswersQuery = `
SELECT id, application_id, question_id, body, created_at, updated_at
FROM answers WHERE application_id = $1
`

const CheckApplicationOwnershipQuery = `
SELECT user_id FROM applications WHERE id = $1
`
