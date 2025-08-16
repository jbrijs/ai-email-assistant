-- Initial migration for InboxAI
-- Extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto; -- for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS vector;

-- Users (store Gmail history cursor for incremental sync)
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT UNIQUE NOT NULL,
  name TEXT,
  google_id TEXT UNIQUE,
  gmail_history_id TEXT, -- last successful historyId
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Email threads (aggregate view for conversation-level summary/classification)
CREATE TABLE IF NOT EXISTS email_threads (
  id TEXT PRIMARY KEY,                -- Gmail threadId
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  subject TEXT,
  participants JSONB,                 -- [{name,email,role}]
  labels TEXT[],
  snippet TEXT,
  last_message_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_email_threads_user_last
  ON email_threads(user_id, last_message_at DESC);

-- Emails (individual messages)
CREATE TABLE IF NOT EXISTS emails (
  id TEXT PRIMARY KEY,                -- Gmail messageId
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  thread_id TEXT NOT NULL REFERENCES email_threads(id) ON DELETE CASCADE,
  subject TEXT,
  sender TEXT,
  recipients TEXT[],                  
  cc TEXT[],
  bcc TEXT[],
  body_text TEXT,
  body_html TEXT,
  received_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- uniqueness safeguard: a message belongs to exactly one user
CREATE UNIQUE INDEX IF NOT EXISTS ux_emails_user_msg ON emails(user_id, id);
CREATE INDEX IF NOT EXISTS idx_emails_user_thread_rec
  ON emails(user_id, thread_id, received_at DESC);

-- Thread-level LLM outputs (1 row per thread)
CREATE TABLE IF NOT EXISTS summaries (
  thread_id TEXT PRIMARY KEY REFERENCES email_threads(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  summary_md TEXT NOT NULL,           -- markdown bullets
  category TEXT,                      
  priority TEXT,
  tags TEXT[],
  action_items JSONB,                 -- [{task,due_date,owner,confidence}]
  entities JSONB,                     -- {people,orgs,money,dates,ids}
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_summaries_user_category
  ON summaries(user_id, category);

-- RAG chunks (multiple chunks per message/thread)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='rag_chunks' AND column_name='embedding') THEN
    CREATE TABLE rag_chunks (
      id BIGSERIAL PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      thread_id TEXT NOT NULL REFERENCES email_threads(id) ON DELETE CASCADE,
      message_id TEXT NOT NULL REFERENCES emails(id) ON DELETE CASCADE,
      chunk_index INT NOT NULL,       -- order of chunks within the message
      chunk_text TEXT NOT NULL,
      embedding VECTOR(768),         
      sent_at TIMESTAMPTZ,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
  END IF;
END$$;

CREATE UNIQUE INDEX IF NOT EXISTS ux_rag_chunks_msg_idx
  ON rag_chunks(message_id, chunk_index);
CREATE INDEX IF NOT EXISTS idx_rag_chunks_user_sent
  ON rag_chunks(user_id, sent_at DESC);
-- Vector index: create after data load for speed; run ANALYZE afterward.
-- CREATE INDEX IF NOT EXISTS idx_rag_chunks_vec
--   ON rag_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Lightweight graph
CREATE TABLE IF NOT EXISTS kg_nodes (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT,                          -- 'Person','Org','Invoice','Event','Thread'
  key TEXT,                           -- unique per user, e.g. 'email:jane@acme.com'
  props JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_kg_nodes_user_key
  ON kg_nodes(user_id, key);

CREATE TABLE IF NOT EXISTS kg_edges (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  src_id BIGINT NOT NULL REFERENCES kg_nodes(id) ON DELETE CASCADE,
  dst_id BIGINT NOT NULL REFERENCES kg_nodes(id) ON DELETE CASCADE,
  type TEXT,                          -- 'MENTIONS','ABOUT','SENT_BY','RELATED_TO','SCHEDULED_FOR', etc.
  props JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_kg_edges_user_src_dst
  ON kg_edges(user_id, src_id, dst_id);

-- Jobs table: queue with SKIP LOCKED
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_kind') THEN
    CREATE TYPE job_kind AS ENUM ('INGEST_THREAD','SUMMARIZE_THREAD','EMBED_MESSAGE','EXTRACT_ENTITIES');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
    CREATE TYPE job_status AS ENUM ('QUEUED','RUNNING','SUCCEEDED','FAILED');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS jobs (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kind job_kind NOT NULL,
  payload JSONB NOT NULL,             -- carries thread_id/message_id, etc.
  status job_status NOT NULL DEFAULT 'QUEUED',
  run_after TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  last_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_jobs_status_run_after ON jobs(status, run_after);
CREATE INDEX IF NOT EXISTS idx_jobs_user ON jobs(user_id);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname='tg_users_updated_at') THEN
    CREATE TRIGGER tg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname='tg_threads_updated_at') THEN
    CREATE TRIGGER tg_threads_updated_at BEFORE UPDATE ON email_threads
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname='tg_emails_updated_at') THEN
    CREATE TRIGGER tg_emails_updated_at BEFORE UPDATE ON emails
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;
