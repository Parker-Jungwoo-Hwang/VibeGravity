// ============================================================
// FILE     : cmd/worker/main_test.go
// PURPOSE  : Verifies worker composition preserves reasoning stop-lines.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : worker main composition tests
// DEPENDS  : context, encoding/json, testing, time, internal/core, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Worker defaults may use mocked Codex bridge runners, not local text extraction or real Codex calls.
// ============================================================

package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

func TestNewReasonerUsesMockedCodexBridgeWithoutLocalExtraction(t *testing.T) {
	t.Parallel()

	reasoner, err := newReasoner(nil)
	if err != nil {
		t.Fatalf("newReasoner returned error: %v", err)
	}

	result, err := reasoner.ProcessTurn(context.Background(), workerStopLineEnvelope())
	if err != nil {
		t.Fatalf("ProcessTurn returned error: %v", err)
	}
	if len(result.Stage1.CandidateMemories) != 0 {
		t.Fatalf("default worker reasoning must not extract local memories from raw text, got %#v", result.Stage1.CandidateMemories)
	}
	if result.Stage1.SummaryHint != "mock_codex_stage1_no_candidates" {
		t.Fatalf("default worker reasoning should use the mocked Codex bridge, got %#v", result.Stage1)
	}
	if len(result.Stage2.Operations) != 0 {
		t.Fatalf("default mocked worker reasoning must not produce graph writes, got %#v", result.Stage2.Operations)
	}
	if len(result.Stage2.Trace.Codes) != 1 || result.Stage2.Trace.Codes[0] != "mock_codex_bridge_no_operations" {
		t.Fatalf("expected mocked Codex bridge trace, got %#v", result.Stage2.Trace)
	}
}

func TestNewReasonerIgnoresRealCodexEnvAndStaysMocked(t *testing.T) {
	t.Setenv("VIBEGRAVITY_CODEX_ENABLED", "true")
	t.Setenv("VIBEGRAVITY_CODEX_ENDPOINT", "https://codex.invalid/should-not-be-called")
	t.Setenv("VIBEGRAVITY_CODEX_MODEL", "real-codex-disabled-in-worker-defaults")

	reasoner, err := newReasoner(nil)
	if err != nil {
		t.Fatalf("newReasoner returned error: %v", err)
	}

	result, err := reasoner.ProcessTurn(context.Background(), workerStopLineEnvelope())
	if err != nil {
		t.Fatalf("ProcessTurn returned error: %v", err)
	}
	if result.Stage1.SummaryHint != "mock_codex_stage1_no_candidates" {
		t.Fatalf("worker default should remain mocked even with Codex env set, got %#v", result.Stage1)
	}
	if len(result.Stage1.CandidateEntities) != 0 || len(result.Stage1.CandidateMemories) != 0 || len(result.Stage2.Operations) != 0 {
		t.Fatalf("worker default must not locally extract or write graph operations, got stage1=%#v stage2=%#v", result.Stage1, result.Stage2)
	}
	if string(result.Stage2.Trace.MetadataJSON) != `{"client":"mock_codex_json_client"}` {
		t.Fatalf("expected mocked bridge metadata, got %s", string(result.Stage2.Trace.MetadataJSON))
	}
}

func workerStopLineEnvelope() *reasoning.ProcessTurnEnvelope {
	rawEvents := []*core.RawEvent{{
		ID:          "evt_stop_line_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		SessionID:   "session_1",
		ActorID:     "agent:hermes-main",
		EventKind:   "message",
		Source:      "hermes",
		Fingerprint: "fp_evt_stop_line_1",
		OccurredAt:  time.Date(2026, time.April, 25, 8, 0, 0, 0, time.UTC),
		PayloadJSON: json.RawMessage(`{"text":"Remember this as a local extractor regression if parsed locally."}`),
		CreatedAt:   time.Date(2026, time.April, 25, 8, 0, 0, 0, time.UTC),
	}}
	return &reasoning.ProcessTurnEnvelope{
		JobID:       "job_stop_line",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEventIDs: []string{"evt_stop_line_1"},
		RawEvents:   rawEvents,
		Stage1: reasoning.Stage1Input{
			JobID:       "job_stop_line",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			RawEvents:   rawEvents,
		},
		Stage2: reasoning.Stage2Input{
			JobID:                "job_stop_line",
			TenantID:             "tenant_1",
			WorkspaceID:          "workspace_1",
			RawEvents:            rawEvents,
			RequiredOutputName:   reasoning.StageNameResolve,
			RequiredOutputSchema: reasoning.Stage2ResolveOutputSchemaV0,
		},
	}
}
