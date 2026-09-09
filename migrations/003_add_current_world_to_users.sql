-- Добавляем поле current_world_id в таблицу users
ALTER TABLE users ADD COLUMN current_world_id UUID REFERENCES worlds(id);

-- Для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_users_current_world ON users(current_world_id);