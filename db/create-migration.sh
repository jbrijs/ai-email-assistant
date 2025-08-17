#!/bin/bash

# Script to create new migration files
# Usage: ./create-migration.sh <migration_name>
# Example: ./create-migration.sh add_user_preferences

if [ $# -eq 0 ]; then
    echo "Usage: $0 <migration_name>"
    echo "Example: $0 add_user_preferences"
    exit 1
fi

MIGRATION_NAME=$1
TIMESTAMP=$(date +%Y%m%d%H%M%S)
FILENAME="${TIMESTAMP}_${MIGRATION_NAME}.sql"
FILEPATH="migrations/${FILENAME}"

# Create migrations directory if it doesn't exist
mkdir -p migrations

# Create the migration file
cat > "$FILEPATH" << EOF
-- Migration: ${MIGRATION_NAME}
-- Created: $(date)
-- Description: 

-- Up migration
-- Add your SQL statements here

-- Down migration (rollback)
-- Add your rollback SQL statements here
EOF

echo "Created migration file: $FILEPATH"
echo "Edit the file to add your migration logic"
echo ""
echo "To run migrations: docker-compose up migrations"
echo "To rollback: docker-compose run --rm migrations -path /migrations -database postgres://postgres:postgres@db:5432/inboxai?sslmode=disable down 1"
