package queries

// User-related SQL queries

const (
	// CreateUserQuery inserts a new user into the database
	CreateUserQuery = `
		INSERT INTO users (id, full_name, email, phone_number, hashed_password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, full_name, email, phone_number, role, created_at, updated_at
	`

	// CreateUserWithVerificationQuery inserts a new user into the database with email verification
	CreateUserWithVerificationQuery = `
		INSERT INTO users (id, full_name, email, phone_number, verified, reset_token, reset_token_expires_at, hashed_password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, full_name, email, phone_number, role, created_at, updated_at
	`

	// GetAllUsersQuery retrieves all users from the database
	GetAllUsersQuery = `
		SELECT id, full_name, email, phone_number, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	// GetUserByIDQuery retrieves a single user by their ID
	GetUserByIDQuery = `
		SELECT id, full_name, email, phone_number, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	// GetUserByEmailQuery retrieves a user by their email (for authentication)
	GetUserByEmailQuery = `
		SELECT id, full_name, email, phone_number, hashed_password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	// UpdateUserRoleQuery updates a user's role
	UpdateUserRoleQuery = `
		UPDATE users 
		SET role = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, full_name, email, phone_number, role, created_at, updated_at
	`
)
