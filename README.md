# Recruitment Backend

A Go-based backend service for recruitment management built with Gin and PostgreSQL.

## Project Structure

```
├── main.go                 # Application entry point
├── cmds/                   # Command line utilities
│   ├── main.go            # CLI entry point
│   ├── jwt_secret.go      # JWT secret generator
│   ├── hash_password.go   # Password hasher
│   └── README.md          # CLI documentation
├── database/
│   └── connection.go       # Database connection and configuration
├── models/
│   ├── user.go            # User model and DTOs
│   └── migrations/        # Database migration files
├── routes/
│   └── users.go           # User-related API routes
├── middleware/
│   ├── middleware.go      # Custom middleware function
└── utils/
    ├── env.go             # Environment variable utilities
    └── utils_test.go      # Comprehensive tests
```

## Environment Variables

Set the following environment variables (defaults provided for development):

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=recruitment_db
DB_MAX_CONNS=10
DB_MIN_CONNS=2

# Server Configuration
PORT=8080
GIN_MODE=debug

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY_DURATION=24h  # Optional: JWT token expiry duration (default: 24h)
                         # Examples: 1h, 30m, 2h30m, 7d, 168h
```

## Configuration

## Command Line Utilities

The project includes helpful command line utilities in the `cmds/` directory:

### Generate JWT Secret
```bash
go run cmds/*.go jwt-secret
# or using make
make jwt-secret
```

### Hash Password
```bash
go run cmds/*.go hash-password <password>
# or using make
make hash-password PWD=yourpassword
```

### Build Standalone CLI
```bash
make build-cli
./bin/recruitment-cli jwt-secret
./bin/recruitment-cli hash-password mypassword
```

For detailed documentation, see [cmds/README.md](cmds/README.md).

## Development Setup

1. Install PostgreSQL and create a database
2. Run the migration to create the users table:
   ```sql
   -- Execute the SQL in models/migrations/001_create_users_table.sql
   ```
3. Set environment variables or use defaults
4. Run the application:
   ```bash
   go run main.go
   # or using make
   make run
   ```

## Available Make Commands

```bash
make help           # Show all available commands
make run            # Start the server
make build          # Build the application
make jwt-secret     # Generate a random JWT secret
make hash-password  # Generate a password hash (usage: make hash-password PWD=yourpassword)
make build-cli      # Build standalone CLI utility
make test           # Run tests
make clean          # Clean build artifacts
```

## Contributing

Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:

- How to fork the repository and create pull requests
- Semantic commit message format
- Code formatting and quality standards
- Development setup

Before contributing, make sure to:
- Format your code with `go fmt ./...`
- Follow semantic commit message conventions
- Create PRs from your own fork