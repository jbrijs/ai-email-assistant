-- Migration: 000001_init
-- Created: 2024-12-01
-- Description: Rollback initial database schema for InboxAI

-- Drop triggers first
DROP TRIGGER IF EXISTS tg_emails_updated_at ON emails;
DROP TRIGGER IF EXISTS tg_threads_updated_at ON email_threads;
DROP TRIGGER IF EXISTS tg_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS set_updated_at();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS kg_edges;
DROP TABLE IF EXISTS kg_nodes;
DROP TABLE IF EXISTS rag_chunks;
DROP TABLE IF EXISTS summaries;
DROP TABLE IF EXISTS emails;
DROP TABLE IF EXISTS email_threads;
DROP TABLE IF EXISTS users;

-- Drop custom types
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
    DROP TYPE job_status;
  END IF;
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_kind') THEN
    DROP TYPE job_kind;
  END IF;
END$$;

-- Note: We don't drop the extensions (pgcrypto, vector) as they might be used by other parts of the system
-- If you need to drop them, uncomment the lines below:
-- DROP EXTENSION IF EXISTS vector;
-- DROP EXTENSION IF EXISTS pgcrypto;
