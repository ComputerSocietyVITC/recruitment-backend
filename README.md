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

## Environment Variables

Copy the example environment file and update the values:

```bash
cp .env.example .env
# Then edit .env with your preferred values
```

Required environment variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres                    # or recruitment_user if you created a specific user
DB_PASSWORD=password                # change to your secure password
DB_NAME=recruitment_db
DB_MAX_CONNS=10
DB_MIN_CONNS=2

# Server Configuration
PORT=8080
GIN_MODE=debug                      # use 'release' for production

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production  # Generate with: make jwt-secret
JWT_EXPIRY_DURATION=24h            # Optional: JWT token expiry duration (default: 24h)
                                   # Examples: 1h, 30m, 2h30m, 7d, 168h
```

**Security Notes:**
- Change the default `DB_PASSWORD` in production
- Generate a secure JWT secret using `make jwt-secret`
- Use environment-specific configurations
- The default super admin password is `password123` - change this immediately in production

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