package queries

// User-related SQL queries

const (
	// CreateUserQuery inserts a new user into the database
	CreateUserQuery = `
		INSERT INTO users (full_name, email, reg_num, verified, reset_token, reset_token_expires_at, hashed_password, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, full_name, email, reg_num, verified, role, created_at, updated_at
	`

	// GetAllUsersQuery retrieves all users from the database
	GetAllUsersQuery = `
		SELECT id, full_name, email, reg_num, verified, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	// GetUserByIDQuery retrieves a single user by their ID
	GetUserByIDQuery = `
		SELECT id, full_name, email, reg_num, verified, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	// GetUserByEmailQuery retrieves a user by their email (for authentication)
	GetUserByEmailQuery = `
		SELECT id, full_name, email, reg_num, verified, reset_token, reset_token_expires_at, hashed_password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	// GetUserByEmailPublicQuery retrieves a user by their email (public fields only)
	GetUserByEmailPublicQuery = `
		SELECT id, full_name, email, reg_num, verified, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	// UpdateUserRoleQuery updates a user's role
	UpdateUserRoleQuery = `
		UPDATE users 
		SET role = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, full_name, email, reg_num, role, created_at, updated_at
	`

	// UpdateUserRoleAndDepartmentQuery updates a user's role and department
	UpdateUserRoleAndDepartmentQuery = `
		UPDATE users 
		SET role = $2, department = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, full_name, email, reg_num, role, department, created_at, updated_at
	`

	// UpdateUserVerificationStatusQuery updates a user's email verification status
	UpdateUserVerificationStatusQuery = `
		UPDATE users
		SET verified = $2, reset_token = NULL, reset_token_expires_at = NULL, updated_at = NOW()
		WHERE id = $1
		RETURNING id, full_name, email, reg_num, verified, role, created_at, updated_at
	`

	// UpdateUserResetTokenQuery updates a user's reset token and expiration time
	UpdateUserResetTokenQuery = `
		UPDATE users
		SET reset_token = $2, reset_token_expires_at = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, full_name, email, reg_num, verified, reset_token, reset_token_expires_at, hashed_password, role, created_at, updated_at
	`

	// UpdateUserPasswordQuery updates a user's password and clears reset token
	UpdateUserPasswordQuery = `
		UPDATE users
		SET hashed_password = $2, reset_token = NULL, reset_token_expires_at = NULL, updated_at = NOW()
		WHERE id = $1
		RETURNING id, full_name, email, reg_num, verified, role, created_at, updated_at
	`

	// DeleteUserQuery deletes a user by their ID
	DeleteUserQuery = `
		DELETE FROM users
		WHERE id = $1
	`
)
