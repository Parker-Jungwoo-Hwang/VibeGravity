CREATE TABLE raw_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    event_kind TEXT NOT NULL,
    source TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    payload_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX raw_events_tenant_source_idempotency_key_idx
    ON raw_events (tenant_id, source, idempotency_key);
CREATE INDEX raw_events_tenant_workspace_session_created_at_idx
    ON raw_events (tenant_id, workspace_id, session_id, created_at DESC);

CREATE TABLE ingest_jobs (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    job_kind TEXT NOT NULL CHECK (job_kind IN (
        'process_turn_event',
        'embed_document_chunks',
        'dream_session',
        'dream_workspace',
        'rebuild_profile',
        'correction_apply',
        'maintenance'
    )),
    status TEXT NOT NULL DEFAULT 'queued',
    raw_event_ids TEXT[] NOT NULL DEFAULT '{}',
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_by TEXT,
    locked_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ingest_jobs_kind_status_available_at_idx
    ON ingest_jobs (job_kind, status, available_at);

CREATE TABLE entities (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    entity_kind TEXT NOT NULL,
    display_name TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memory_groups (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN (
        'agent_private',
        'workspace_shared',
        'group_shared',
        'session_scratch'
    )),
    group_id TEXT,
    owner_entity_id TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN (
        'fact',
        'preference',
        'trait',
        'goal',
        'constraint',
        'relationship',
        'decision',
        'procedure',
        'task_state',
        'doc_fact',
        'summary',
        'hypothesis'
    )),
    artifact_class TEXT NOT NULL DEFAULT 'knowledge' CHECK (artifact_class IN (
        'context',
        'knowledge',
        'timeline',
        'plan'
    )),
    text TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN (
        'active',
        'superseded',
        'archived',
        'deleted'
    )),
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ,
    latest_flag BOOLEAN NOT NULL DEFAULT true,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX memories_tenant_workspace_scope_status_idx
    ON memories (tenant_id, workspace_id, scope, status);
CREATE INDEX memories_fingerprint_idx
    ON memories (fingerprint);

CREATE TABLE memory_edges (
    from_memory_id TEXT NOT NULL REFERENCES memories (id) ON DELETE CASCADE,
    to_memory_id TEXT NOT NULL REFERENCES memories (id) ON DELETE CASCADE,
    edge_kind TEXT NOT NULL CHECK (edge_kind IN (
        'updates',
        'extends',
        'supports',
        'contradicts',
        'derived_from',
        'references_doc',
        'belongs_to',
        'corrected_by'
    )),
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    created_by_job_id TEXT REFERENCES ingest_jobs (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (from_memory_id, to_memory_id, edge_kind)
);

CREATE INDEX memory_edges_from_memory_edge_kind_idx
    ON memory_edges (from_memory_id, edge_kind);
CREATE INDEX memory_edges_to_memory_edge_kind_idx
    ON memory_edges (to_memory_id, edge_kind);
CREATE UNIQUE INDEX memory_edges_single_updates_target_idx
    ON memory_edges (to_memory_id)
    WHERE edge_kind = 'updates';

CREATE TABLE memory_trace (
    memory_id TEXT PRIMARY KEY REFERENCES memories (id) ON DELETE CASCADE,
    raw_event_ids TEXT[] NOT NULL,
    reasoning_job_id TEXT REFERENCES ingest_jobs (id),
    reasoning_stage TEXT NOT NULL,
    candidate_snapshot_json JSONB NOT NULL,
    applied_operations_json JSONB NOT NULL,
    operator_correction_flag BOOLEAN NOT NULL DEFAULT false,
    related_document_ids TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memory_corrections (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    memory_id TEXT NOT NULL REFERENCES memories (id) ON DELETE CASCADE,
    operator_id TEXT NOT NULL,
    raw_event_id TEXT NOT NULL REFERENCES raw_events (id) ON DELETE CASCADE,
    idempotency_key TEXT NOT NULL,
    correction_text TEXT NOT NULL,
    evidence_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'recorded' CHECK (status IN (
        'recorded',
        'applied',
        'dismissed'
    )),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX memory_corrections_tenant_workspace_idempotency_key_idx
    ON memory_corrections (tenant_id, workspace_id, idempotency_key);
CREATE UNIQUE INDEX memory_corrections_raw_event_id_idx
    ON memory_corrections (raw_event_id);
CREATE INDEX memory_corrections_memory_created_at_idx
    ON memory_corrections (memory_id, created_at DESC);

CREATE TABLE profiles (
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN (
        'agent_private',
        'workspace_shared',
        'group_shared',
        'session_scratch'
    )),
    static_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    dynamic_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_memory_ids TEXT[] NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version BIGINT NOT NULL DEFAULT 1,
    PRIMARY KEY (tenant_id, workspace_id, entity_id, scope)
);

CREATE INDEX profiles_tenant_workspace_entity_updated_at_idx
    ON profiles (tenant_id, workspace_id, entity_id, updated_at DESC);

CREATE TABLE session_summaries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    summary_text TEXT NOT NULL,
    source_event_ids TEXT[] NOT NULL DEFAULT '{}',
    source_memory_ids TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX session_summaries_tenant_workspace_session_updated_at_idx
    ON session_summaries (tenant_id, workspace_id, session_id, updated_at DESC);

CREATE TABLE notes (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    note_kind TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN (
        'agent_private',
        'workspace_shared',
        'group_shared',
        'session_scratch'
    )),
    owner_entity_id TEXT NOT NULL,
    text TEXT NOT NULL,
    pinned BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX notes_workspace_pinned_expires_at_idx
    ON notes (workspace_id, pinned, expires_at);

CREATE TABLE plans (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN (
        'agent_private',
        'workspace_shared',
        'group_shared',
        'session_scratch'
    )),
    owner_entity_id TEXT NOT NULL,
    evidence_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX plans_workspace_status_idx
    ON plans (workspace_id, status);

CREATE TABLE plan_items (
    id TEXT PRIMARY KEY,
    plan_id TEXT NOT NULL REFERENCES plans (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    evidence_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    source TEXT NOT NULL,
    title TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    version_hint TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX documents_tenant_workspace_fingerprint_idx
    ON documents (tenant_id, workspace_id, fingerprint);

CREATE TABLE document_chunks (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL CHECK (chunk_index >= 0),
    text TEXT NOT NULL,
    heading_path TEXT NOT NULL DEFAULT '',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    neighbor_chunk_ids TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX document_chunks_document_chunk_index_idx
    ON document_chunks (document_id, chunk_index);

CREATE TABLE memory_group_memberships (
    group_id TEXT NOT NULL REFERENCES memory_groups (id) ON DELETE CASCADE,
    entity_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, entity_id)
);
