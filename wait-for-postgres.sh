#!/bin/bash

set -e

host="$1"
port="$2"
user="$3"
password="$4"
database="$5"
shift 5
cmd="$@"

# Function to check if PostgreSQL is ready
check_postgres() {
    PGPASSWORD="$password" psql -h "$host" -p "$port" -U "$user" -d "$database" -c '\q' 2>/dev/null
}

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL at $host:$port..."
until check_postgres; do
    echo "PostgreSQL is unavailable - sleeping"
    sleep 2
done

echo "PostgreSQL is up - checking if migrations need to be run..."

# Check if tables exist (specifically the users table which should be created first)
TABLES_EXIST=$(PGPASSWORD="$password" psql -h "$host" -p "$port" -U "$user" -d "$database" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users';" 2>/dev/null | tr -d ' ')

if [ "$TABLES_EXIST" = "0" ]; then
    echo "Running database migrations..."
    
    # Run the migration script
    PGPASSWORD="$password" psql -h "$host" -p "$port" -U "$user" -d "$database" -f /root/models/migrations/001_initial_up.sql
    
    echo "Database migrations completed successfully!"
else
    echo "Database tables already exist, skipping migrations..."
fi

echo "Starting the application..."
exec $cmd
