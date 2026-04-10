-- migrations/004_create_user_salt_table.down.sql

-- Удаляем индексы
DROP INDEX IF EXISTS idx_user_salt_user_id;
DROP INDEX IF EXISTS idx_user_salt_created_at;

-- Удаляем таблицу
DROP TABLE IF EXISTS user_salt;