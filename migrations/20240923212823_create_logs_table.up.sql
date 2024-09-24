-- migrations/xxxxxx_create_logs_table.up.sql
CREATE TABLE logs (
                      id SERIAL PRIMARY KEY,
                      log_message TEXT NOT NULL,
                      log_level VARCHAR(20) NOT NULL,
                      created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_logs_created_at ON logs (created_at);