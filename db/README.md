# Database Migrations

This directory contains database migrations for the InboxAI project. Migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate) and run automatically as part of the Docker Compose setup.

## How It Works

1. **Automatic Execution**: Migrations run automatically when you start the stack with `docker-compose up`
2. **Order Guaranteed**: The API service won't start until all migrations complete successfully
3. **One-shot Service**: The migration service runs once and exits after completion

## File Naming Convention

Migrations follow the format: `YYYYMMDDHHMMSS_description.sql`

Example: `20241201143000_add_user_preferences.sql`

## Creating New Migrations

Use the helper script to create new migration files:

```bash
cd db
./create-migration.sh <migration_name>
```

Example:

```bash
./create-migration.sh add_user_preferences
```

## Migration File Structure

Each migration file should contain:

```sql
-- Migration: add_user_preferences
-- Created: 2024-12-01 14:30:00
-- Description: Add user preferences table

-- Up migration (applied when running 'up')
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theme TEXT DEFAULT 'light',
    notifications_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Down migration (applied when running 'down')
DROP TABLE IF EXISTS user_preferences;
```

## Running Migrations Manually

### Run all pending migrations:

```bash
docker-compose run --rm migrations -path /migrations -database postgres://postgres:postgres@db:5432/inboxai?sslmode=disable up
```

### Rollback last migration:

```bash
docker-compose run --rm migrations -path /migrations -database postgres://postgres:postgres@db:5432/inboxai?sslmode=disable down 1
```

### Check migration status:

```bash
docker-compose run --rm migrations -path /migrations -database postgres://postgres:postgres@db:5432/inboxai?sslmode=disable version
```

## Troubleshooting

### Migration Fails

If migrations fail, check the logs:

```bash
docker-compose logs migrations
```

### Database Connection Issues

Ensure the database is healthy before running migrations:

```bash
docker-compose ps db
```

### Reset All Migrations

To reset all migrations (⚠️ **DESTRUCTIVE** - will drop all data):

```bash
docker-compose run --rm migrations -path /migrations -database postgres://postgres:postgres@db:5432/inboxai?sslmode=disable down
```

## Best Practices

1. **Always test migrations** in development before applying to production
2. **Keep migrations small** and focused on a single change
3. **Include rollback logic** in your down migrations
4. **Never modify existing migration files** that have been applied
5. **Use transactions** for complex migrations when possible

## Current Migrations

- `0001_init.sql` - Initial database schema with users, emails, threads, and RAG support
