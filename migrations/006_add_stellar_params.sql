ALTER TABLE worlds ADD COLUMN spectral_class TEXT DEFAULT 'G';
ALTER TABLE worlds ADD COLUMN temperature INT DEFAULT 5778;
COMMENT ON COLUMN worlds.spectral_class IS 'Спектральный класс звезды (O, B, A, F, G, K, M, L, T, Y)';
COMMENT ON COLUMN worlds.temperature IS 'Температура звезды в Кельвинах';