-- Создаем таблицу токенов сброса пароля
CREATE TABLE IF NOT EXISTS reset_password_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем индексы для улучшения производительности
CREATE INDEX IF NOT EXISTS idx_reset_password_tokens_token ON reset_password_tokens(token);
CREATE INDEX IF NOT EXISTS idx_reset_password_tokens_expires_at ON reset_password_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_reset_password_tokens_user_id ON reset_password_tokens(user_id);