# Docker Setup for Recruitment Backend

This directory contains Docker configuration files to containerize the recruitment backend application with PostgreSQL database.

## Files Overview

- `Dockerfile` - Multi-stage build for the Go application
- `docker-compose.yml` - Complete stack with PostgreSQL and application
- `wait-for-postgres.sh` - Script to handle database initialization and migrations
- `.env.docker` - Environment variables template for Docker
- `.dockerignore` - Files to exclude from Docker build context

## Quick Start

### 1. Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+

### 2. Configuration

Copy the environment template and update the values:

```bash
cp .env.docker .env
```

**Important**: Update the following in your `.env` file:
- `POSTGRES_PASSWORD` - Use a strong password
- `JWT_SECRET` - Generate using: `openssl rand -base64 32`
- SMTP settings for email functionality

### 3. Start the Application

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# View only app logs
docker-compose logs -f app
```

### 4. First Run Setup

On the first run, the application will:

1. **Wait for PostgreSQL** to be ready
2. **Check if tables exist** in the database
3. **Run migrations automatically** if this is a fresh database
4. **Start the application** once everything is ready

The migration script (`wait-for-postgres.sh`) ensures that:
- Database tables are created from `models/migrations/001_initial_up.sql`
- Application starts only after successful database setup
- Subsequent runs skip migration if tables already exist

## Available Services

### Application
- **URL**: http://localhost:8080
- **Health Check**: http://localhost:8080/ping
- **API Base**: http://localhost:8080/api/v1

### PostgreSQL Database
- **Host**: localhost
- **Port**: 5432
- **Database**: recruitment_db
- **Username**: postgres
- **Password**: (as set in .env file)

## Common Commands

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f [service_name]

# Rebuild application (after code changes)
docker-compose build app
docker-compose up -d app

# Access database directly
docker-compose exec postgres psql -U postgres -d recruitment_db

# Access application container
docker-compose exec app sh

# Remove everything (including volumes)
docker-compose down -v
```

## Development Workflow

### Making Code Changes

1. Make your code changes
2. Rebuild the application:
   ```bash
   docker-compose build app
   docker-compose up -d app
   ```

### Database Management

#### Reset Database
```bash
# Stop services and remove volumes
docker-compose down -v

# Start fresh (will run migrations again)
docker-compose up -d
```

#### Manual Migration
```bash
# Access the database container
docker-compose exec postgres psql -U postgres -d recruitment_db

# Run SQL commands manually
\i /path/to/migration.sql
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_DB` | Database name | recruitment_db |
| `POSTGRES_USER` | Database user | postgres |
| `POSTGRES_PASSWORD` | Database password | ⚠️ Set in .env |
| `JWT_SECRET` | JWT signing secret | ⚠️ Set in .env |
| `SMTP_HOST` | SMTP server host | smtp.gmail.com |
| `SMTP_PORT` | SMTP server port | 587 |
| `SMTP_USER` | SMTP username | ⚠️ Set in .env |
| `SMTP_PASSWORD` | SMTP password | ⚠️ Set in .env |

## Security Considerations

### Production Deployment

1. **Change default passwords**:
   - Database password
   - JWT secret
   - Default admin credentials

2. **Configure HTTPS**:
   - Use a reverse proxy (nginx/traefik)
   - Set up SSL certificates

3. **Network security**:
   - Remove port exposures if using reverse proxy
   - Configure proper firewall rules

4. **Environment variables**:
   - Use Docker secrets or external secret management
   - Never commit `.env` files with real credentials

### Example Production docker-compose.override.yml

```yaml
version: '3.8'
services:
  app:
    ports: []  # Remove direct port exposure
    environment:
      GIN_MODE: release
      
  postgres:
    ports: []  # Remove direct port exposure
```

## Troubleshooting

### Application won't start
1. Check if PostgreSQL is healthy: `docker-compose ps`
2. View application logs: `docker-compose logs app`
3. Verify environment variables in `.env`

### Database connection issues
1. Ensure PostgreSQL is running: `docker-compose ps postgres`
2. Check database logs: `docker-compose logs postgres`
3. Verify database credentials match between app and postgres services

### Migration issues
1. Check if migration files exist in container: `docker-compose exec app ls -la models/migrations/`
2. Manually run migration: `docker-compose exec postgres psql -U postgres -d recruitment_db -f /path/to/migration.sql`

### Port conflicts
If port 8080 or 5432 are in use:
```yaml
# In docker-compose.yml, change ports mapping
services:
  app:
    ports:
      - "3000:8080"  # Use port 3000 instead
  postgres:
    ports:
      - "5433:5432"  # Use port 5433 instead
```

## Monitoring

### Health Checks
- PostgreSQL: Built-in health check using `pg_isready`
- Application: Access `/ping` endpoint

### Logs
```bash
# Follow all logs
docker-compose logs -f

# Follow specific service
docker-compose logs -f app
docker-compose logs -f postgres

# Show last N lines
docker-compose logs --tail=50 app
```

## Cleanup

```bash
# Stop and remove containers
docker-compose down

# Remove containers, networks, and volumes
docker-compose down -v

# Remove images as well
docker-compose down -v --rmi all
```
