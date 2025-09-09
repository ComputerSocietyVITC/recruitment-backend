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

## Database Setup

This application requires PostgreSQL with the `uuid-ossp` extension. You can install PostgreSQL either natively on your system or using Docker (recommended for development).

### Option 1: PostgreSQL with Docker (Recommended)

1. **Start PostgreSQL with Docker**:
   ```bash
   # Quick setup using make command (recommended)
   make db-docker
   
   # Or manually create and start a PostgreSQL container:
   docker run -d \
     --name recruitment_postgres \
     -e POSTGRES_DB=recruitment_db \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=password \
     -p 5432:5432 \
     -v recruitment_postgres_data:/var/lib/postgresql/data \
     postgres:17
   
   # Then run migrations
   make db-up
   ```

2. **Verify connection**:
   ```bash
   docker exec -it recruitment_postgres psql -U postgres -d recruitment_db
   ```

3. **Stop the database when done**:
   ```bash
   make db-docker-stop
   # Or manually: docker stop recruitment_postgres
   ```

### Option 2: Native PostgreSQL Installation

#### Ubuntu/Debian:
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### macOS (using Homebrew):
```bash
brew install postgresql
brew services start postgresql
```

#### Windows:
Download and install from [PostgreSQL official website](https://www.postgresql.org/download/windows/)

### Create Database and User

After installing PostgreSQL, create the database:

```bash
# Connect to PostgreSQL
sudo -u postgres psql

# Create database and user
CREATE DATABASE recruitment_db;
CREATE USER recruitment_user WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE recruitment_db TO recruitment_user;
\q
```

### Run Database Migrations

The project includes make commands to simplify migration management:

```bash
# Run migrations to create tables
make db-up

# Or manually with psql:
psql -h localhost -U postgres -d recruitment_db -f models/migrations/001_initial_up.sql
```

To rollback migrations (if needed):
```bash
# Rollback migrations
make db-down

# Or manually:
psql -h localhost -U postgres -d recruitment_db -f models/migrations/001_initial_down.sql
```

**Note**: The migration script will create:
- Users table with roles (applicant, evaluator, admin, super_admin)
- Questions table for different departments
- Applications table for user applications
- Answers table for question responses
- A default super admin user with email: `admin@comp.socks` and password: `password123`

To rollback migrations (if needed):
```bash
psql -h localhost -U postgres -d recruitment_db -f models/migrations/001_initial_down.sql
```

## Environment Configuration

This application follows best practices for environment variable management across different deployment scenarios.

### Quick Setup

1. **For Development:**
   ```bash
   cp .env.example .env
   # Edit .env with your development values
   ```

2. **For Testing:**
   ```bash
   cp .env.testing.example .env.testing
   # The app will automatically load this when ENV=testing
   ```

3. **For Production:**
   ```bash
   cp .env.production.example .env.production
   # Configure with secure production values
   # Consider using your platform's secrets management instead
   ```

4. **For Docker:**
   ```bash
   cp .env.docker.example .env.docker
   # Use with docker-compose
   ```

### Environment Loading Priority

The application loads environment files in this order (first found wins):

1. `.env.{ENV}.local` (e.g., `.env.development.local`)
2. `.env.{ENV}` (e.g., `.env.development`)
3. `.env.local`
4. `.env`
5. System environment variables (always available)

### Environment Variables

#### Core Configuration
```bash
# Environment: development, testing, production
ENV=development

# Server
PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password    # REQUIRED in production
DB_NAME=recruitment_db
DB_MAX_CONNS=10                     # Connection pool size
DB_MIN_CONNS=2                      # Minimum connections

# Security
JWT_SECRET=your-32+-char-secret     # REQUIRED - Generate with: make jwt-secret
JWT_EXPIRY_DURATION=24h             # Examples: 1h, 30m, 7d

# OTP Configuration
EMAIL_VERIFICATION_OTP_DURATION=10m # Email verification OTP validity (examples: 5m, 10m, 15m)
PASSWORD_RESET_OTP_DURATION=30m     # Password reset OTP validity (examples: 15m, 30m, 1h)

# Network Security
TRUSTED_PROXIES=127.0.0.1,10.0.0.0/8    # Comma-separated trusted proxy IPs/ranges
CORS_ALLOWED_ORIGINS=*                   # Comma-separated allowed origins (* for dev only!)

# Email
SMTP_HOST=smtp.example.com          # REQUIRED in production
SMTP_PORT=587
SMTP_USER=your_smtp_user            # REQUIRED in production
SMTP_PASSWORD=your_smtp_password    # REQUIRED in production
EMAIL_FROM=recruitment@yourcompany.com

# Email Templates (customize with your organization's branding)
# Use {{.OTP}}, {{.TOKEN}}, {{.DURATION}} as placeholders
EMAIL_VERIFICATION_SUBJECT=Verify your email address
EMAIL_VERIFICATION_BODY=Your OTP is: <strong>{{.OTP}}</strong>. Valid for {{.DURATION}}.
EMAIL_RESEND_VERIFICATION_SUBJECT=New verification code
EMAIL_RESEND_VERIFICATION_BODY=Your new OTP: <strong>{{.OTP}}</strong>. Valid for {{.DURATION}}.
EMAIL_PASSWORD_RESET_SUBJECT=Password reset request
EMAIL_PASSWORD_RESET_BODY=Reset token: <strong>{{.TOKEN}}</strong>. Valid for {{.DURATION}}.
EMAIL_PASSWORD_RESET_SUCCESS_SUBJECT=Password reset successful
EMAIL_PASSWORD_RESET_SUCCESS_BODY=Your password was successfully reset.

# Business Logic
ALLOWED_EMAIL_DOMAINS=company.com,university.edu
MAXIMUM_APPLICATIONS_PER_USER=2
```

#### Environment-Specific Behavior

- **Development**: Loads `.env` files, relaxed validation, additional logging
- **Testing**: Uses test database, shorter JWT expiry, debug features
- **Production**: Strict validation, requires all security variables, no `.env` loading by default

#### Security Notes

🔒 **Critical for Production:**
- Generate JWT_SECRET with: `make jwt-secret` (minimum 32 characters)
- Use strong, unique database passwords
- Configure proper SMTP credentials
- Review ALLOWED_EMAIL_DOMAINS carefully
- Set CORS_ALLOWED_ORIGINS to specific domains (never use * in production)
- Configure TRUSTED_PROXIES for your infrastructure
- Never commit actual `.env` files to version control
- Use your deployment platform's secrets management

🚨 **Default Credentials to Change:**
- Default super admin password: `password123`
- Default JWT_SECRET in examples

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

1. **Set up PostgreSQL**: Follow the [Database Setup](#database-setup) section above
2. **Run database migrations**: Execute the migration scripts as described in the database setup
3. **Configure environment**: Set environment variables or create a `.env` file (see [Environment Variables](#environment-variables))
4. **Install Go dependencies**:
   ```bash
   go mod tidy
   ```
5. **Run the application**:
   ```bash
   go run main.go
   # or using make
   make run
   ```

The server will start on the port specified in your environment variables (default: 8080).

### Default Admin Access

After running migrations, you can access the application with the default super admin account:
- **Email**: `admin@comp.socks`
- **Password**: `password123`

**Important**: Change these credentials immediately in production!

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

# Database commands
make db-docker      # Start PostgreSQL in Docker container and run migrations
make db-up          # Run database migrations (up)
make db-down        # Rollback database migrations (down)
make db-docker-stop # Stop PostgreSQL Docker container
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