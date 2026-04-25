ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS tenant_id TEXT NOT NULL DEFAULT 'legacy_tenant';

ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS workspace_id TEXT NOT NULL DEFAULT 'legacy_workspace';

ALTER TABLE profiles
    ALTER COLUMN tenant_id DROP DEFAULT;

ALTER TABLE profiles
    ALTER COLUMN workspace_id DROP DEFAULT;

ALTER TABLE profiles
    DROP CONSTRAINT IF EXISTS profiles_pkey;

ALTER TABLE profiles
    ADD PRIMARY KEY (tenant_id, workspace_id, entity_id, scope);

DROP INDEX IF EXISTS profiles_entity_updated_at_idx;

CREATE INDEX IF NOT EXISTS profiles_tenant_workspace_entity_updated_at_idx
    ON profiles (tenant_id, workspace_id, entity_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS session_summaries_tenant_workspace_session_updated_at_idx
    ON session_summaries (tenant_id, workspace_id, session_id, updated_at DESC);
