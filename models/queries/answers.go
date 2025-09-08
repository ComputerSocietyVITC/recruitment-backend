package queries

const GetApplicationAnswersQuery = `
SELECT id, application_id, question_id, body, created_at, updated_at
FROM answers WHERE application_id = $1
`

const CheckApplicationOwnershipQuery = `
SELECT user_id FROM applications WHERE id = $1 AND submitted = false
`

const UpsertAnswerQuery = `
INSERT INTO answers (id, application_id, user_id, question_id, body, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (application_id, question_id)
DO UPDATE SET 
    body = EXCLUDED.body,
    updated_at = EXCLUDED.updated_at
RETURNING id, application_id, user_id, question_id, body, created_at, updated_at
`

const DeleteAnswerQuery = `
DELETE FROM answers 
WHERE id = $1 AND application_id = $2 AND user_id = $3
`

const GetAnswerByIDQuery = `
SELECT id, application_id, user_id, question_id, body, created_at, updated_at
FROM answers 
WHERE id = $1
`

const GetUserAnswersForApplicationQuery = `
SELECT a.id, a.application_id, a.user_id, a.question_id, a.body, a.created_at, a.updated_at
FROM answers a
INNER JOIN applications app ON a.application_id = app.id
WHERE a.application_id = $1 AND app.user_id = $2
ORDER BY a.created_at ASC
`

const GetAnswersByUserQuery = `
SELECT id, application_id, user_id, question_id, body, created_at, updated_at
FROM answers 
WHERE user_id = $1
ORDER BY created_at DESC
`

const ValidateQuestionApplicationDepartmentQuery = `
SELECT app.department as app_department, q.department as question_department
FROM applications app, questions q
WHERE app.id = $1 AND q.id = $2
`
