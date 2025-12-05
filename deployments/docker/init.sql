-- DMN Engine Database Schema

-- Definitions table
CREATE TABLE IF NOT EXISTS dmn_definitions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key             VARCHAR(255) NOT NULL,
    version         INT NOT NULL DEFAULT 1,
    name            VARCHAR(255),
    source          TEXT NOT NULL,
    parsed_model    JSONB,
    checksum        VARCHAR(64),
    tenant_id       VARCHAR(64),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT unique_definition UNIQUE(key, version, tenant_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_dmn_def_key ON dmn_definitions(key);
CREATE INDEX IF NOT EXISTS idx_dmn_def_tenant ON dmn_definitions(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_dmn_def_key_version ON dmn_definitions(key, version DESC);
CREATE INDEX IF NOT EXISTS idx_dmn_def_created ON dmn_definitions(created_at DESC);

-- Evaluation history (for audit and analytics)
CREATE TABLE IF NOT EXISTS dmn_evaluation_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    definition_id   UUID REFERENCES dmn_definitions(id),
    definition_key  VARCHAR(255) NOT NULL,
    inputs          JSONB NOT NULL,
    outputs         JSONB,
    matched_rules   TEXT[],
    duration_ns     BIGINT,
    evaluated_at    TIMESTAMPTZ DEFAULT NOW(),
    tenant_id       VARCHAR(64),
    correlation_id  VARCHAR(255),
    error_message   TEXT
);

-- Indexes for history
CREATE INDEX IF NOT EXISTS idx_eval_hist_def_key ON dmn_evaluation_history(definition_key);
CREATE INDEX IF NOT EXISTS idx_eval_hist_tenant ON dmn_evaluation_history(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_eval_hist_evaluated ON dmn_evaluation_history(evaluated_at DESC);
CREATE INDEX IF NOT EXISTS idx_eval_hist_correlation ON dmn_evaluation_history(correlation_id) WHERE correlation_id IS NOT NULL;

-- Optional: Partitioning for high-volume history (uncomment if needed)
-- CREATE TABLE dmn_evaluation_history_2024_01 PARTITION OF dmn_evaluation_history
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

