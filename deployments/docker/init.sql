-- Initial database setup for DMN Engine
-- This file is executed when PostgreSQL container starts for the first time

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Main definitions table
CREATE TABLE IF NOT EXISTS dmn_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         VARCHAR(255) NOT NULL,
    version     INT NOT NULL DEFAULT 1,
    name        VARCHAR(255),
    source      TEXT NOT NULL,
    parsed_model JSONB,
    checksum    VARCHAR(64),
    tenant_id   VARCHAR(64),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_definition UNIQUE(key, version, tenant_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_dmn_def_key ON dmn_definitions(key);
CREATE INDEX IF NOT EXISTS idx_dmn_def_tenant ON dmn_definitions(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_dmn_def_key_version ON dmn_definitions(key, version DESC);
CREATE INDEX IF NOT EXISTS idx_dmn_def_created ON dmn_definitions(created_at DESC);

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO dmn;
