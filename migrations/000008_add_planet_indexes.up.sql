-- migrations/000008_add_planet_indexes.up.sql

-- Индекс для связи планет с мирами (уже есть, но добавим на всякий случай)
CREATE INDEX IF NOT EXISTS idx_planets_world_id ON planets(world_id);

-- GIN индекс для JSONB поля data (ускоряет поиск внутри data->'resources' и data->>'type')
CREATE INDEX IF NOT EXISTS idx_planets_data ON planets USING GIN (data);

