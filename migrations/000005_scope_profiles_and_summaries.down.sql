DROP INDEX IF EXISTS session_summaries_tenant_workspace_session_updated_at_idx;

DROP INDEX IF EXISTS profiles_tenant_workspace_entity_updated_at_idx;

ALTER TABLE profiles
    DROP CONSTRAINT IF EXISTS profiles_pkey;

ALTER TABLE profiles
    ADD PRIMARY KEY (entity_id, scope);

CREATE INDEX IF NOT EXISTS profiles_entity_updated_at_idx
    ON profiles (entity_id, updated_at DESC);

ALTER TABLE profiles
    DROP COLUMN IF EXISTS workspace_id;

ALTER TABLE profiles
    DROP COLUMN IF EXISTS tenant_id;
