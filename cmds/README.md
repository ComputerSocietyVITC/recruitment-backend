# Command Line Utilities

This directory contains command line utilities for the recruitment backend project.

## Available Commands

### 1. Generate JWT Secret

Generates a cryptographically secure random JWT secret that can be used for signing JWT tokens.

```bash
go run cmds/main.go jwt-secret
```

**Output:**
- Base64 encoded secret (recommended for environment variables)
- Hex encoded secret
- Example .env file entry

### 2. Hash Password

Generates a bcrypt hash for a given password. Useful for creating admin accounts or testing.

```bash
go run cmds/main.go hash-password <password>
```

**Example:**
```bash
go run cmds/main.go hash-password mypassword123
```

**Output:**
- Original password (for reference)
- Bcrypt hash (safe to store in database)
- Cost factor used

## Usage Examples

### Generate a new JWT secret for production:
```bash
go run cmds/main.go jwt-secret
```

### Create a hashed password for an admin user:
```bash
go run cmds/main.go hash-password admin123
```

## Security Notes

1. **JWT Secret**: Keep your JWT secret secure and never commit it to version control. Store it in environment variables.

2. **Password Hashing**: The bcrypt hashes generated are safe to store in your database. Never store plain text passwords.

3. **Production Use**: When using these utilities in production, ensure you're running them in a secure environment where the output won't be logged or exposed.

## Building Standalone Executables

You can also build standalone executables for these utilities:

```bash
# Build for current platform
go build -o bin/recruitment-cli cmds/*.go

# Use the executable
./bin/recruitment-cli jwt-secret
./bin/recruitment-cli hash-password mypassword123
```

### Cross-platform builds:
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/recruitment-cli-linux cmds/*.go

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/recruitment-cli.exe cmds/*.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/recruitment-cli-mac cmds/*.go
```
