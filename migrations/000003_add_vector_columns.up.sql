-- ADR-002 forbids hardcoding embedding dimensions before model selection.
-- The pgvector type is intentionally dimensionless here; when ADR-002b fixes
-- the model, regenerate this migration or add a follow-up migration that uses
-- vector(<embedding_dims>) from config and backfills rows safely.
ALTER TABLE memories
    ADD COLUMN embedding vector,
    ADD COLUMN embedding_model TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN embedding_dims INTEGER NOT NULL DEFAULT 0 CHECK (embedding_dims >= 0),
    ADD COLUMN embedding_updated_at TIMESTAMPTZ;

ALTER TABLE document_chunks
    ADD COLUMN embedding vector,
    ADD COLUMN embedding_model TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN embedding_dims INTEGER NOT NULL DEFAULT 0 CHECK (embedding_dims >= 0),
    ADD COLUMN embedding_updated_at TIMESTAMPTZ;
