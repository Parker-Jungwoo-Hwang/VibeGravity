// ============================================================
// FILE     : internal/mcp/stdio_smoke_test.go
// PURPOSE  : Smoke-tests MCP stdio trust-loop tools against core DTO semantics.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MCP stdio smoke tests
// DEPENDS  : internal/mcp/protocol.go, internal/mcp/surface.go, internal/core
// USED_BY  : go test ./internal/mcp
// ------------------------------------------------------------
// AGENT_NOTE: Keep this as protocol smoke coverage; do not add Hermes packaging behavior here.
// ============================================================

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestStdioTrustLoopToolsPassCoreDTOsThrough(t *testing.T) {
	t.Parallel()

	lag := int64(42)
	service := &stdioSmokeService{
		prefetchResp: &core.PrefetchResponse{
			Blocks: []core.RecallBlock{{
				ID:        "mem_1",
				Kind:      "memory",
				Priority:  90,
				Text:      "Keep Hermes first.",
				Scope:     core.MemoryScopeWorkspaceShared,
				Source:    "memories",
				SourceID:  "mem_1",
				Status:    "active",
				Freshness: "stale",
			}},
			Meta: core.RecallMeta{
				EstimatedTokens:     8,
				Sources:             []string{"memories"},
				Freshness:           "stale",
				FreshnessLagSeconds: &lag,
				Degraded:            true,
				DegradedReasons:     []string{"worker backlog"},
			},
		},
		correctResp:  &core.CorrectMemoryResponse{MemoryID: "mem_1", RawEventID: "evt_correction", CorrectionID: "corr_1", CorrectionRecorded: true, TraceWritten: true, Status: "recorded"},
		explainResp:  &core.ExplainMemoryResponse{MemoryID: "mem_1", Trace: core.MemoryTraceResult{ReasoningJobID: "job_1", OperatorCorrectionFlag: true}},
		timelineResp: &core.GetTimelineResponse{Items: []core.TimelineItem{{ID: "item_1", MemoryID: "mem_1", Text: "Correction recorded"}}},
	}
	server := newProtocolServer(t, newTestSurface(t, service))

	responses := serveStdioMessages(t, server, []string{
		`{"jsonrpc":"2.0","id":"init","method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}`,
		`{"jsonrpc":"2.0","id":"list","method":"tools/list"}`,
		callMessage("recall", "recall_preview", `{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main","query":"next task","budget_tokens":512,"mode":"default"}`),
		callMessage("correct", "correct_memory", `{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1","operator_id":"operator_1","idempotency_key":"idem_1","correction_text":"Use the corrected rule.","evidence_json":{"source":"smoke"}}`),
		callMessage("explain", "explain_memory", `{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1","entity_id":"agent:hermes-main","visible_group_ids":["group_design"]}`),
		callMessage("timeline", "view_timeline", `{"tenant_id":"tenant_1","workspace_id":"workspace_1","scopes":["workspace_shared","agent_private"],"entity_id":"agent:hermes-main","limit":7}`),
		callMessage("degraded", "degraded_status", `{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main","query":"status","budget_tokens":128,"mode":"default"}`),
	})

	assertRPCResultContains(t, responses["init"], `"protocolVersion":"2025-11-25"`)
	assertToolListed(t, responses["list"], "recall_preview")
	assertToolListed(t, responses["list"], "correct_memory")
	assertToolListed(t, responses["list"], "explain_memory")
	assertToolListed(t, responses["list"], "view_timeline")
	assertToolListed(t, responses["list"], "degraded_status")
	assertRPCResultContains(t, responses["recall"], "Keep Hermes first.")
	assertRPCResultContains(t, responses["correct"], `"correction_recorded":true`)
	assertRPCResultContains(t, responses["explain"], `"reasoning_job_id":"job_1"`)
	assertRPCResultContains(t, responses["timeline"], "Correction recorded")
	assertRPCResultContains(t, responses["degraded"], `"freshness":"stale"`)

	if service.prefetchCalls != 2 {
		t.Fatalf("expected recall_preview and degraded_status to call Prefetch, got %d calls", service.prefetchCalls)
	}
	if got := service.prefetchReqs[0]; got.TenantID != "tenant_1" || got.WorkspaceID != "workspace_1" || got.SessionID != "session_1" || got.ActorID != "agent:hermes-main" || got.Query != "next task" || got.BudgetTokens != 512 || got.Mode != "default" {
		t.Fatalf("recall_preview request was not passed through unchanged: %#v", got)
	}
	if got := service.correctReq; got.TenantID != "tenant_1" || got.WorkspaceID != "workspace_1" || got.MemoryID != "mem_1" || got.OperatorID != "operator_1" || got.IdempotencyKey != "idem_1" || got.CorrectionText != "Use the corrected rule." || string(got.EvidenceJSON) != `{"source":"smoke"}` {
		t.Fatalf("correct_memory request was not passed through unchanged: %#v", got)
	}
	if got := service.explainReq; got.TenantID != "tenant_1" || got.WorkspaceID != "workspace_1" || got.MemoryID != "mem_1" || got.EntityID != "agent:hermes-main" || len(got.VisibleGroupIDs) != 1 || got.VisibleGroupIDs[0] != "group_design" {
		t.Fatalf("explain_memory request was not passed through unchanged: %#v", got)
	}
	if got := service.timelineReq; got.TenantID != "tenant_1" || got.WorkspaceID != "workspace_1" || got.EntityID != "agent:hermes-main" || got.Limit != 7 || len(got.Scopes) != 2 || got.Scopes[0] != core.MemoryScopeWorkspaceShared || got.Scopes[1] != core.MemoryScopeAgentPrivate {
		t.Fatalf("view_timeline request was not passed through unchanged: %#v", got)
	}
	if got := service.prefetchReqs[1]; got.Query != "status" || got.BudgetTokens != 128 {
		t.Fatalf("degraded_status request was not passed through unchanged: %#v", got)
	}
}

func TestStdioToolCallReportsMissingRequiredCoreField(t *testing.T) {
	t.Parallel()

	service := &stdioSmokeService{correctResp: &core.CorrectMemoryResponse{}}
	server := newProtocolServer(t, newTestSurface(t, service))
	responses := serveStdioMessages(t, server, []string{
		callMessage("missing", "correct_memory", `{"tenant_id":"tenant_1","workspace_id":"workspace_1","operator_id":"operator_1","correction_text":"missing target"}`),
	})

	var envelope struct {
		Result callToolResult `json:"result"`
	}
	decodeJSONMessage(t, responses["missing"], &envelope)
	if !envelope.Result.IsError {
		t.Fatalf("expected missing memory_id to return a tool error, got %s", string(responses["missing"]))
	}
	if !strings.Contains(envelope.Result.Content[0].Text, "memory_id is required") {
		t.Fatalf("expected required-field error, got %s", envelope.Result.Content[0].Text)
	}
}

func callMessage(id string, name string, arguments string) string {
	return fmt.Sprintf(`{"jsonrpc":"2.0","id":%q,"method":"tools/call","params":{"name":%q,"arguments":%s}}`, id, name, arguments)
}

func serveStdioMessages(t *testing.T, server *Server, messages []string) map[string]json.RawMessage {
	t.Helper()

	input := strings.Join(append(messages, ""), "\n")
	var out strings.Builder
	if err := server.ServeStdio(context.Background(), strings.NewReader(input), &out); err != nil {
		t.Fatalf("ServeStdio returned error: %v", err)
	}
	responses := make(map[string]json.RawMessage)
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var envelope struct {
			ID json.RawMessage `json:"id"`
		}
		decodeJSONMessage(t, json.RawMessage(line), &envelope)
		var id string
		decodeJSONMessage(t, envelope.ID, &id)
		responses[id] = json.RawMessage(line)
	}
	return responses
}

func assertRPCResultContains(t *testing.T, raw json.RawMessage, want string) {
	t.Helper()

	if len(raw) == 0 {
		t.Fatalf("missing JSON-RPC response")
	}
	if !strings.Contains(string(raw), want) {
		t.Fatalf("expected response to contain %q, got %s", want, string(raw))
	}
}

func assertToolListed(t *testing.T, raw json.RawMessage, name string) {
	t.Helper()

	assertRPCResultContains(t, raw, fmt.Sprintf(`"name":%q`, name))
}

type stdioSmokeService struct {
	prefetchCalls int
	prefetchReqs  []core.PrefetchRequest
	correctReq    *core.CorrectMemoryRequest
	explainReq    *core.ExplainMemoryRequest
	timelineReq   *core.GetTimelineRequest
	prefetchResp  *core.PrefetchResponse
	correctResp   *core.CorrectMemoryResponse
	explainResp   *core.ExplainMemoryResponse
	timelineResp  *core.GetTimelineResponse
}

func (s *stdioSmokeService) Prefetch(_ context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchCalls++
	s.prefetchReqs = append(s.prefetchReqs, *req)
	return s.prefetchResp, nil
}

func (s *stdioSmokeService) SyncTurn(context.Context, *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *stdioSmokeService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *stdioSmokeService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *stdioSmokeService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *stdioSmokeService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *stdioSmokeService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *stdioSmokeService) UpdatePlan(context.Context, *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *stdioSmokeService) CorrectMemory(_ context.Context, req *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	if req.MemoryID == "" {
		return nil, errors.Join(core.ErrInvalidArgument, errors.New("memory_id is required"))
	}
	s.correctReq = req
	return s.correctResp, nil
}

func (s *stdioSmokeService) GetTimeline(_ context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	s.timelineReq = req
	return s.timelineResp, nil
}

func (s *stdioSmokeService) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainReq = req
	return s.explainResp, nil
}
