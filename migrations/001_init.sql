-- ============================
-- Enable extensions
-- ============================

CREATE EXTENSION IF NOT EXISTS postgis;

-- ============================
-- Geo Content table
-- ============================

CREATE TABLE geo_content (
    id UUID PRIMARY KEY,

    external_id UUID UNIQUE NOT NULL,

    type TEXT NOT NULL,

    -- Geographic location (WGS84)
    location GEOGRAPHY(POINT, 4326) NOT NULL,

    published_at TIMESTAMPTZ NOT NULL,

    -- Arbitrary metadata (tags, source, severity, etc.)
    attributes JSONB NOT NULL DEFAULT '{}',

    -- Base relevance score (domain-defined)
    base_score DOUBLE PRECISION NOT NULL DEFAULT 1.0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================
-- Indexes
-- ============================

-- Spatial index
CREATE INDEX idx_geo_content_location
ON geo_content
USING GIST (location);

-- Common filters
CREATE INDEX idx_geo_content_type
ON geo_content (type);

-- JSONB metadata queries
CREATE INDEX idx_geo_content_attributes
ON geo_content
USING GIN (attributes);

-- Sorting / pagination helper
CREATE INDEX idx_geo_content_published_at
ON geo_content (published_at DESC);