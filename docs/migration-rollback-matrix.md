# Migration Rollback Matrix

This matrix records rollback status for each migration version. It is the
operator-facing source of truth; do not edit already-released migration files
just to update rollback labels.

`Rollback Status` meanings:

- `Reversible`: down migration should restore the previous schema shape without
  known data loss beyond data written only to the new schema.
- `Scratch only`: down migration exists but can drop data or semantics; use only
  on empty, scratch, or backed-up databases.
- `Restore preferred`: prefer database restore or roll-forward fix for live data.

| Version | Purpose | Rollback Status | Down Migration | Live Data Risk | Release Note |
|---|---|---|---|---|---|
| `000001_create_pgvector_extension` | Enables pgvector extension. | Reversible on scratch DB; restore preferred on shared DB. | Drops `vector` extension if present. | Dropping the extension can break vector columns or other databases sharing the extension. | Confirm no dependent vector columns before down. |
| `000002_create_core_tables` | Creates core raw event, job, entity, memory, trace, correction, profile, note, plan, document, and group tables. | Scratch only. | Drops all core VibeGravity tables. | Destructive for all stored memory, raw events, corrections, notes, plans, documents, profiles, and jobs. | Use backup restore for live data. |
| `000003_add_vector_columns` | Adds embedding columns to memories and document chunks. | Reversible with data loss in embedding columns. | Drops embedding columns. | Loses stored embeddings and model/dims metadata on affected rows. | Safe only if embeddings can be recomputed and the release notes say so. |
| `000004_fix_updates_edge_target_index` | Corrects `updates` uniqueness guard to the target memory side. | Restore preferred for live trust-loop data. | Recreates the older source-side unique index. | Rolling back weakens the active correction lineage guard and may allow invalid latest-memory semantics. | Prefer roll-forward unless validating an old scratch schema. |
| `000005_scope_profiles_and_summaries` | Adds tenant/workspace scoping to profiles and session summary lookup index. | Scratch only unless backup exists. | Removes tenant/workspace columns from profiles and restores older primary key. | Can collapse tenant/workspace identity boundaries and lose scoped profile semantics. | Do not run down on live scoped memory without a tested restore plan. |

## Release Owner Checklist

- [ ] New migration added to this matrix.
- [ ] Rollback status reviewed against live data, not only syntax.
- [ ] `down.sql` exists or the migration is explicitly marked non-reversible.
- [ ] Release notes include migration order and rollback decision.
- [ ] Live PostgreSQL gate passed after applying migrations.
- [ ] Backup or snapshot id recorded before applying migrations.

## Source Review

Estimated source: current `migrations/` files and VibeGravity storage
invariants.

Suspected license: none.

Similarity risk: low.

Review required: yes before first public tag.
