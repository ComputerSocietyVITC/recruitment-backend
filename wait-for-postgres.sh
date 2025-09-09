#!/bin/bash

set -e

host="$1"
port="$2"
user="$3"
password="$4"
database="$5"
shift 5
cmd="$@"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to log with timestamp
log() {
    echo -e "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

# Function to check if PostgreSQL is ready
check_postgres() {
    PGPASSWORD="$password" psql -h "$host" -p "$port" -U "$user" -d "$database" -c '\q' 2>/dev/null
}

# Function to check if a table exists
table_exists() {
    local table_name=$1
    PGPASSWORD="$password" psql -h "$host" -p "$port" -U "$user" -d "$database" -t -c \
        "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table_name';" \
        2>/dev/null | tr -d ' '
}

# Function to run migration
run_migration() {
    local migration_file=$1
    log "${GREEN}Running migration: $migration_file${NC}"
    
    if PGPASSWORD="$password" psql -h "$host" -p "$port" -U "$user" -d "$database" -f "$migration_file"; then
        log "${GREEN}Migration completed successfully: $migration_file${NC}"
        return 0
    else
        log "${RED}Migration failed: $migration_file${NC}"
        return 1
    fi
}

# Wait for PostgreSQL to be ready
log "${YELLOW}Waiting for PostgreSQL at $host:$port...${NC}"
retry_count=0
max_retries=30

until check_postgres; do
    retry_count=$((retry_count + 1))
    if [ $retry_count -ge $max_retries ]; then
        log "${RED}PostgreSQL is still unavailable after $max_retries attempts - giving up${NC}"
        exit 1
    fi
    log "${YELLOW}PostgreSQL is unavailable - sleeping (attempt $retry_count/$max_retries)${NC}"
    sleep 2
done

log "${GREEN}PostgreSQL is up and running!${NC}"

# Check if this is a migration-only run
if [ "$cmd" = "echo Migration completed" ]; then
    log "${YELLOW}Running database migrations...${NC}"
    
    # Check if tables exist (specifically the users table which should be created first)
    TABLES_EXIST=$(table_exists "users")
    
    if [ "$TABLES_EXIST" = "0" ]; then
        log "${YELLOW}Database tables not found, running initial migration...${NC}"
        
        # Run all migration files in order
        migration_dir="/root/models/migrations"
        if [ -d "$migration_dir" ]; then
            for migration_file in "$migration_dir"/*_up.sql; do
                if [ -f "$migration_file" ]; then
                    if ! run_migration "$migration_file"; then
                        log "${RED}Migration failed, exiting...${NC}"
                        exit 1
                    fi
                else
                    log "${YELLOW}No migration files found in $migration_dir${NC}"
                fi
            done
        else
            log "${RED}Migration directory not found: $migration_dir${NC}"
            exit 1
        fi
        
        log "${GREEN}All database migrations completed successfully!${NC}"
    else
        log "${GREEN}Database tables already exist, skipping migrations...${NC}"
    fi
    
    # Verify migration success
    USERS_TABLE_EXISTS=$(table_exists "users")
    APPLICATIONS_TABLE_EXISTS=$(table_exists "applications")
    QUESTIONS_TABLE_EXISTS=$(table_exists "questions")
    
    if [ "$USERS_TABLE_EXISTS" = "1" ] && [ "$APPLICATIONS_TABLE_EXISTS" = "1" ] && [ "$QUESTIONS_TABLE_EXISTS" = "1" ]; then
        log "${GREEN}Database schema verification successful!${NC}"
    else
        log "${RED}Database schema verification failed!${NC}"
        exit 1
    fi
    
    log "${GREEN}Migration service completed successfully!${NC}"
    exit 0
fi

# For application startup
log "${GREEN}Starting the application...${NC}"
exec $cmd
