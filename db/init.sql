-- Initialize the inboxai database
-- This runs when the PostgreSQL container first starts

-- Create the database if it doesn't exist
SELECT 'CREATE DATABASE inboxai'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'inboxai')\gexec

-- Connect to the inboxai database
\c inboxai;
