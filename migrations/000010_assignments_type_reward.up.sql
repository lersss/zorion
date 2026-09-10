-- migrations/000010_assignments_type_reward.up.sql
--
-- Добавляет недостающие колонки в assignments:
--   type   — тип контракта (используется в фильтрах и в Create)
--   reward — награда (целое число)
--
-- IF NOT EXISTS — защита на случай, если миграция уже применялась
-- вручную (колонки были добавлены через DBeaver до появления файла).

ALTER TABLE assignments
    ADD COLUMN IF NOT EXISTS type text NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS reward integer NOT NULL DEFAULT 0;