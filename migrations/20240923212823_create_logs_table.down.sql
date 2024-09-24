-- migrations/xxxxxx_create_logs_table.down.sql
DROP INDEX IF EXISTS idx_logs_created_at;
DROP TABLE IF EXISTS logs;