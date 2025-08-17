# InboxAI

A smart email management system that uses AI to summarize, categorize, and extract insights from your emails.

## Features

- **Email Integration**: Connect to Gmail and sync email history
- **AI Summarization**: Automatically summarize email threads using LLMs
- **Smart Categorization**: Categorize emails by priority, type, and content
- **RAG Search**: Vector-based search through email content
- **Knowledge Graph**: Build relationships between people, organizations, and topics

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Make (optional, for convenience commands)

### 1. Start the Services

```bash
# Start all services (database, API, Ollama)
make up

# Or use docker compose directly
docker compose up -d
```

### 2. Check Service Status

```bash
# View all running services
docker compose ps

# Check logs
make logs
```

### 3. Run Migrations

Migrations run automatically when you start the stack. The API service won't start until all migrations complete successfully.

To run migrations manually:

```bash
make migrate
```

## Project Structure

```
inbox-ai/
├── apps/
│   ├── api/          # Go API server
│   └── web/          # React frontend
├── db/
│   ├── migrations/   # Database migrations
│   ├── init.sql      # Initial database setup
│   └── create-migration.sh  # Helper script
├── docker-compose.yml
└── Makefile          # Convenience commands
```

## Database Migrations

The project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management.

### Creating New Migrations

```bash
# Create a new migration
make create-migration NAME=add_user_preferences

# Or use the script directly
cd db && ./create-migration.sh add_user_preferences
```

### Migration Commands

```bash
# Check migration status
make migrate-status

# Run pending migrations
make migrate

# Rollback last migration
make migrate-down
```

### Migration Files

Migrations are stored in `db/migrations/` with the format:

- `000001_description.up.sql` - Forward migration
- `000001_description.down.sql` - Rollback migration

## Development

### Useful Commands

```bash
# Build and start services
make build

# Stop all services
make down

# Clean everything (including volumes)
make clean

# Connect to database shell
make db-shell

# Connect to API container
make api-shell
```

### Adding New Features

1. **Database Changes**: Create a new migration using `make create-migration`
2. **API Changes**: Modify files in `apps/api/`
3. **Frontend Changes**: Modify files in `apps/web/`

## Services

- **Database**: PostgreSQL with pgvector extension for vector operations
- **API**: Go server handling business logic and AI integration
- **Ollama**: Local LLM service for AI operations
- **Web**: React frontend (coming soon)

## Environment Variables

Create a `.env.local` file with your configuration. Your Go API will automatically load this file:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=inboxai
DB_SSLMODE=disable

# API Configuration
PORT=3001
```

**Note**: The API will use these values, but if not provided, it will fall back to sensible defaults for local development.

### How Environment Variables Work

1. **Go API**: Uses `godotenv.Load()` to automatically load `.env.local` file
2. **Docker Compose**: References `.env.local` via `env_file: ./.env.local`
3. **Database Connection**: Built dynamically from individual DB\_\* variables
4. **Fallbacks**: If variables aren't set, the API uses hardcoded defaults

### When You Need .env.local

- **Custom database credentials** (different from defaults)
- **Custom API port** (different from 3001)
- **Production settings** (different database host, SSL settings, etc.)

For local development with the default Docker setup, you can actually run **without** `.env.local` since the defaults work perfectly.
