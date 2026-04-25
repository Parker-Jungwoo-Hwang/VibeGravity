// ============================================================
// FILE     : internal/recall/freshness.go
// PURPOSE  : Converts worker/Codex freshness signals into recall-visible metadata.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : FreshnessProvider, Freshness, BacklogFreshnessProvider
// DEPENDS  : context, time, internal/core, internal/store
// USED_BY  : internal/recall, cmd/server, cmd/cli, tests
// ------------------------------------------------------------
// AGENT_NOTE: Freshness is operator visibility only; it must not mutate graph state.
// ============================================================

package recall

import (
	"context"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const defaultBacklogStaleAfter = 30 * time.Second

// FreshnessProvider reports whether stored recall may lag behind raw turn input.
type FreshnessProvider interface {
	RecallFreshness(ctx context.Context, req *core.PrefetchRequest) (Freshness, error)
}

// Freshness is a narrow recall-only view of worker/Codex freshness.
type Freshness struct {
	Freshness       string
	LagSeconds      *int64
	Reasons         []string
	AffectedSources []string
}

// BacklogFreshnessProvider maps read-only worker backlog metrics to recall freshness.
type BacklogFreshnessProvider struct {
	Jobs       store.JobMetricsStore
	StaleAfter time.Duration
	Clock      func() time.Time
}

// RecallFreshness returns stale metadata when ready or retrying worker jobs imply
// derived memories may not include the latest raw events.
func (p BacklogFreshnessProvider) RecallFreshness(ctx context.Context, req *core.PrefetchRequest) (Freshness, error) {
	if p.Jobs == nil {
		return Freshness{Freshness: "stored"}, nil
	}
	now := time.Now().UTC()
	if p.Clock != nil {
		now = p.Clock().UTC()
	}
	metrics, err := p.Jobs.GetJobBacklogMetrics(ctx, &core.JobBacklogMetricsRequest{
		TenantID:     req.TenantID,
		WorkspaceID:  req.WorkspaceID,
		GeneratedNow: now,
	})
	if err != nil {
		return Freshness{}, err
	}
	staleAfter := p.StaleAfter
	if staleAfter == 0 {
		staleAfter = defaultBacklogStaleAfter
	}
	state := Freshness{Freshness: "stored"}
	if metrics == nil {
		return state, nil
	}
	if metrics.OldestQueuedAgeSeconds != nil {
		age := *metrics.OldestQueuedAgeSeconds
		if metrics.Counts.ReadyQueued > 0 && age >= int64(staleAfter.Seconds()) {
			state.Freshness = "stale"
			state.LagSeconds = maxLagSeconds(state.LagSeconds, age)
			state.Reasons = append(state.Reasons, "worker_backlog_stale")
			state.AffectedSources = derivedRecallSources()
		}
	}
	if metrics.OldestRunningAgeSeconds != nil {
		age := *metrics.OldestRunningAgeSeconds
		if metrics.Counts.Running > 0 && age >= int64(staleAfter.Seconds()) {
			state.Freshness = "stale"
			state.LagSeconds = maxLagSeconds(state.LagSeconds, age)
			state.Reasons = append(state.Reasons, "worker_running_stale")
			state.AffectedSources = derivedRecallSources()
		}
	}
	if metrics.RetryableQueuedAttempts > 0 {
		if state.Freshness == "stored" {
			state.Freshness = "stale"
			state.AffectedSources = derivedRecallSources()
		}
		state.Reasons = append(state.Reasons, "codex_or_worker_retry_backlog")
	}
	return state, nil
}

func maxLagSeconds(current *int64, candidate int64) *int64 {
	if current != nil && *current >= candidate {
		return current
	}
	value := candidate
	return &value
}

func applyRecallFreshness(blocks []core.RecallBlock, freshness Freshness) []core.RecallBlock {
	if freshness.Freshness == "" || freshness.Freshness == "stored" || len(freshness.AffectedSources) == 0 {
		return blocks
	}
	affected := make(map[string]struct{}, len(freshness.AffectedSources))
	for _, source := range freshness.AffectedSources {
		affected[source] = struct{}{}
	}
	for i := range blocks {
		if _, ok := affected[blocks[i].Source]; ok {
			blocks[i].Freshness = freshness.Freshness
		}
	}
	return blocks
}

func derivedRecallSources() []string {
	return []string{"memories", "profile", "session_summaries"}
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
