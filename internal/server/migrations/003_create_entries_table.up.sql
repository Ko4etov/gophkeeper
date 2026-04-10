-- migrations/001_create_records_table.up.sql

CREATE TABLE IF NOT EXISTS records (
    id BIGSERIAL PRIMARY KEY,
    client_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    tags TEXT[],
    data JSONB NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создаем индексы
CREATE INDEX idx_records_user_id ON records(user_id);
CREATE INDEX idx_records_client_id ON records(client_id);
CREATE INDEX idx_records_user_updated ON records(user_id, updated_at);
CREATE UNIQUE INDEX idx_records_user_client ON records(user_id, client_id);
CREATE INDEX idx_records_data_type ON records(data_type);
CREATE INDEX idx_records_tags ON records USING GIN(tags);

-- Комментарии
COMMENT ON TABLE records IS 'User data entries (passwords, texts, cards, etc.)';
COMMENT ON COLUMN records.client_id IS 'Client-generated UUID for offline sync';
COMMENT ON COLUMN records.data_type IS 'Type: login_password, text, bank_card, binary, otp';
COMMENT ON COLUMN records.data IS 'Encrypted data in JSON format';
COMMENT ON COLUMN records.version IS 'Version number for conflict resolution';