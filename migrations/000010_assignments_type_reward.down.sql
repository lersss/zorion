-- migrations/000010_assignments_type_reward.down.sql
--
-- Откат миграции 000010: убирает колонки type и reward из assignments.
--
-- ВНИМАНИЕ: удалит данные в этих колонках. Использовать только
-- при осознанном откате схемы (например, в тестовой среде).

ALTER TABLE assignments
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS reward;