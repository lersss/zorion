CREATE TABLE IF NOT EXISTS planet_resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT NOT NULL, -- 'mineral', 'organic', 'energy', 'rare'
    hardness FLOAT DEFAULT 0,
    elasticity FLOAT DEFAULT 0,
    conductivity FLOAT DEFAULT 0,
    heat_resistance FLOAT DEFAULT 0,
    chemical_activity FLOAT DEFAULT 0,
    density FLOAT DEFAULT 0,
    biocompatibility FLOAT DEFAULT 0,
    energy_density FLOAT DEFAULT 0,
    volatility FLOAT DEFAULT 0,
    quantity INT NOT NULL DEFAULT 100,
    is_known BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_planet_resources_planet ON planet_resources(planet_id);