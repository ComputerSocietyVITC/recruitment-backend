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
)
