-- migrations/000008_add_planet_indexes.down.sql

DROP INDEX IF EXISTS idx_planets_world_id;
DROP INDEX IF EXISTS idx_planets_data;