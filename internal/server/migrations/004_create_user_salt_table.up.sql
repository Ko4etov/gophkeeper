-- migrations/004_create_user_salt_table.up.sql

-- Создаем таблицу для хранения соли пользователей
CREATE TABLE IF NOT EXISTS user_salt (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    salt TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создаем индексы
CREATE INDEX idx_user_salt_user_id ON user_salt(user_id);
CREATE INDEX idx_user_salt_created_at ON user_salt(created_at);