package queries

// Questions-related SQL queries

const (
	// GetQuestionsByDepartmentQuery fetches all questions for a department
	GetQuestionsByDepartmentQuery = `
		SELECT id, department, body, created_at
		FROM questions
		WHERE department = $1
		ORDER BY created_at ASC
	`

	// GetAllQuestionsQuery fetches all questions from all departments
	GetAllQuestionsQuery = `
		SELECT id, department, body, created_at
		FROM questions
		ORDER BY department ASC, created_at ASC
	`

	// GetQuestionByIDQuery fetches a specific question by ID
	GetQuestionByIDQuery = `
		SELECT id, department, body, created_at
		FROM questions
		WHERE id = $1
	`

	// DeleteQuestionByIDQuery deletes a specific question by ID
	DeleteQuestionByIDQuery = `
		DELETE FROM questions
		WHERE id = $1
	`

	// CreateQuestionQuery inserts a new question
	CreateQuestionQuery = `
		INSERT INTO questions (id, department, body, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, department, body, created_at
	`
)
