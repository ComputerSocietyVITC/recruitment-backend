package queries

// Questions-related SQL queries

const (
	// GetQuestionsByDepartmentQuery fetches all questions for a department
	GetQuestionsByDepartmentQuery = `
		SELECT id, department, title, body, created_at
		FROM questions
		WHERE department = $1
		ORDER BY created_at ASC
	`

	// GetAllQuestionsQuery fetches all questions from all departments
	GetAllQuestionsQuery = `
		SELECT id, department, title, body, created_at
		FROM questions
		ORDER BY department ASC, created_at ASC
	`

	// GetQuestionByIDQuery fetches a specific question by ID
	GetQuestionByIDQuery = `
		SELECT id, department, title, body, created_at
		FROM questions
		WHERE id = $1
	`

	// GetQuestionByApplicationIDQuery fetches questions for a specific application by application ID
	GetQuestionByApplicationIDQuery = `
		SELECT q.id, q.department, q.title, q.body, q.created_at
		FROM questions q
		JOIN applications a ON q.department = a.department
		WHERE a.id = $1
		ORDER BY q.created_at ASC
	`

	// DeleteQuestionByIDQuery deletes a specific question by ID
	DeleteQuestionByIDQuery = `
		DELETE FROM questions
		WHERE id = $1
	`

	// CreateQuestionQuery inserts a new question
	CreateQuestionQuery = `
		INSERT INTO questions (id, department, title, body, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, department, title, body, created_at
	`
)
