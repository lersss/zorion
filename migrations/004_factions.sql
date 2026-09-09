CREATE TABLE IF NOT EXISTS factions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    homeworld_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    strength INT NOT NULL DEFAULT 1,
    resources JSONB NOT NULL DEFAULT '{}',
    color TEXT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_factions_homeworld ON factions(homeworld_id);