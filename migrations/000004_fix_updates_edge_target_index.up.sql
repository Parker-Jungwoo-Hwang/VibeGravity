DROP INDEX IF EXISTS memory_edges_single_updates_target_idx;

CREATE UNIQUE INDEX memory_edges_single_updates_target_idx
    ON memory_edges (to_memory_id)
    WHERE edge_kind = 'updates';
