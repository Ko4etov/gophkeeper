-- migrations/001_create_records_table.down.sql

DROP INDEX IF EXISTS idx_records_user_client;
DROP INDEX IF EXISTS idx_records_user_updated;
DROP INDEX IF EXISTS idx_records_user_deleted;
DROP INDEX IF EXISTS idx_records_user_id;
DROP INDEX IF EXISTS idx_records_client_id;
DROP INDEX IF EXISTS idx_records_data_type;
DROP INDEX IF EXISTS idx_records_tags;

DROP TABLE IF EXISTS records;