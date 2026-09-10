CREATE TABLE compatibility_matrix (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category VARCHAR(50) NOT NULL,   -- 'subterrain' или 'surface'
    type_a VARCHAR(100) NOT NULL,
    type_b VARCHAR(100) NOT NULL,
    compatible BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(category, type_a, type_b)
);