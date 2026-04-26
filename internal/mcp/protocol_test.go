// ============================================================
// FILE     : internal/mcp/protocol_test.go
// PURPOSE  : Verifies MCP JSON-RPC lifecycle, tool listing, and tool call roundtrips.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MCP protocol server tests
// DEPENDS  : internal/mcp/protocol.go, internal/mcp/surface_test.go
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Protocol tests should prove real JSON-RPC shape, not only adapter delegation.
// ============================================================

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
)

func TestServerHandlesInitializeAndToolRoundtrip(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{
		prefetchResp: &core.PrefetchResponse{
			Blocks: []core.RecallBlock{{Kind: "pinned_note", Priority: 100, Text: "Keep Hermes first."}},
			Meta:   core.RecallMeta{EstimatedTokens: 4, Sources: []string{"notes"}},
		},
	})
	server := newProtocolServer(t, surface)

	initResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`))
	if !respond {
		t.Fatalf("initialize did not produce a response")
	}
	var initEnvelope map[string]any
	decodeJSONMessage(t, initResp, &initEnvelope)
	result := initEnvelope["result"].(map[string]any)
	if result["protocolVersion"] != ProtocolVersion {
		t.Fatalf("unexpected protocol version: %#v", result["protocolVersion"])
	}
	if _, ok := result["capabilities"].(map[string]any)["tools"]; !ok {
		t.Fatalf("initialize response did not advertise tools: %s", string(initResp))
	}

	listResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":"tools","method":"tools/list"}`))
	if !respond {
		t.Fatalf("tools/list did not produce a response")
	}
	if !strings.Contains(string(listResp), `"prefetch"`) || !strings.Contains(string(listResp), `"inputSchema"`) {
		t.Fatalf("tools/list missing expected tool schema: %s", string(listResp))
	}
	if !strings.Contains(string(listResp), `"required":["tenant_id","workspace_id","session_id","actor_id"]`) || !strings.Contains(string(listResp), `"recall_preview"`) {
		t.Fatalf("tools/list did not expose recall preview required inputs: %s", string(listResp))
	}

	callResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"prefetch","arguments":{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main"}}}`))
	if !respond {
		t.Fatalf("tools/call did not produce a response")
	}
	if !strings.Contains(string(callResp), `"structuredContent"`) || !strings.Contains(string(callResp), "Keep Hermes first.") {
		t.Fatalf("tools/call did not return structured prefetch output: %s", string(callResp))
	}
}

func TestServerToolSchemasExposeTrustLoopInputs(t *testing.T) {
	t.Parallel()

	server := newProtocolServer(t, newTestSurface(t, &fakeService{}))
	tools := server.protocolTools()

	byName := make(map[string]protocolTool, len(tools))
	for _, tool := range tools {
		byName[tool.Name] = tool
	}
	assertRequiredInputsExact(t, byName["prefetch"], "tenant_id", "workspace_id", "session_id", "actor_id")
	assertRequiredInputsExact(t, byName["recall_preview"], "tenant_id", "workspace_id", "session_id", "actor_id")
	assertRequiredInputsExact(t, byName["degraded_status"], "tenant_id", "workspace_id", "session_id", "actor_id")
	assertRequiredInputsExact(t, byName["sync_turn"], "tenant_id", "workspace_id", "session_id", "actor_id", "idempotency_key", "turn_events")
	assertRequiredInputsExact(t, byName["search_memory"], "tenant_id", "workspace_id")
	assertRequiredInputsExact(t, byName["search_documents"], "tenant_id", "workspace_id")
	assertRequiredInputsExact(t, byName["add_note"], "tenant_id", "workspace_id", "scope", "owner_entity_id", "text")
	assertRequiredInputsExact(t, byName["create_plan"], "tenant_id", "workspace_id", "title", "scope", "owner_entity_id")
	assertRequiredInputsExact(t, byName["update_plan"], "tenant_id", "workspace_id", "plan_id")
	assertRequiredInputsExact(t, byName["correct_memory"], "tenant_id", "workspace_id", "memory_id", "operator_id", "idempotency_key", "correction_text")
	assertRequiredInputsExact(t, byName["view_timeline"], "tenant_id", "workspace_id", "entity_id")
	assertRequiredInputsExact(t, byName["explain_memory"], "tenant_id", "workspace_id", "memory_id")

	correctionProps := byName["correct_memory"].InputSchema["properties"].(map[string]any)
	if _, ok := correctionProps["evidence_json"]; !ok {
		t.Fatalf("correct_memory schema should expose evidence_json for provenance")
	}
	if _, ok := correctionProps["entity_id"]; !ok {
		t.Fatalf("correct_memory schema should expose entity_id for private visibility")
	}
	if _, ok := correctionProps["visible_group_ids"]; !ok {
		t.Fatalf("correct_memory schema should expose visible_group_ids for group visibility")
	}
	timelineProps := byName["view_timeline"].InputSchema["properties"].(map[string]any)
	if _, ok := timelineProps["scopes"]; !ok {
		t.Fatalf("view_timeline schema should expose scopes for visibility review")
	}
	explainProps := byName["explain_memory"].InputSchema["properties"].(map[string]any)
	if _, ok := explainProps["entity_id"]; !ok {
		t.Fatalf("explain_memory schema should expose entity_id for private visibility")
	}
	if _, ok := explainProps["visible_group_ids"]; !ok {
		t.Fatalf("explain_memory schema should expose visible_group_ids for group visibility")
	}
}

func TestServerRequiredToolSchemaMatchesServiceValidation(t *testing.T) {
	t.Parallel()

	server := newProtocolServer(t, newTestSurface(t, requiredFieldValidationService{}))
	tools := server.protocolTools()
	for _, tool := range tools {
		required, ok := tool.InputSchema["required"].([]string)
		if !ok || len(required) == 0 {
			continue
		}
		validArgs := validToolArguments(tool.Name)
		completeResp, respond := server.HandleMessage(context.Background(), callToolMessage(t, tool.Name, validArgs))
		if !respond {
			t.Fatalf("%s complete call did not produce a response", tool.Name)
		}
		var complete rpcResponseForTest
		decodeJSONMessage(t, completeResp, &complete)
		if complete.Error != nil || complete.Result.IsError {
			t.Fatalf("%s complete call should pass service validation: %s", tool.Name, string(completeResp))
		}
		for _, field := range required {
			args := validToolArguments(tool.Name)
			delete(args, field)
			raw, respond := server.HandleMessage(context.Background(), callToolMessage(t, tool.Name, args))
			if !respond {
				t.Fatalf("%s missing %s did not produce a response", tool.Name, field)
			}
			var envelope rpcResponseForTest
			decodeJSONMessage(t, raw, &envelope)
			if envelope.Error != nil {
				t.Fatalf("%s missing %s should be a tool validation error, got protocol error: %s", tool.Name, field, string(raw))
			}
			if !envelope.Result.IsError {
				t.Fatalf("%s missing %s should fail service validation: %s", tool.Name, field, string(raw))
			}
			if !strings.Contains(envelope.Result.Content[0].Text, field+" is required") {
				t.Fatalf("%s missing %s returned wrong validation message: %s", tool.Name, field, string(raw))
			}
		}
	}
}

func TestServerToolCallUsesServiceValidationForPrefetchSchema(t *testing.T) {
	t.Parallel()

	service := &prefetchValidationService{assembler: recall.NewAssembler(recall.Dependencies{})}
	server := newProtocolServer(t, newTestSurface(t, service))

	completeResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"recall_preview","arguments":{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main"}}}`))
	if !respond {
		t.Fatalf("complete recall_preview did not produce a response")
	}
	var completeEnvelope rpcResponseForTest
	decodeJSONMessage(t, completeResp, &completeEnvelope)
	if completeEnvelope.Error != nil {
		t.Fatalf("complete recall_preview returned protocol error: %s", string(completeResp))
	}
	if completeEnvelope.Result.IsError {
		t.Fatalf("complete recall_preview returned tool error: %s", string(completeResp))
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected complete call to reach service once, got %d", service.prefetchCalls)
	}

	incompleteResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"recall_preview","arguments":{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1"}}}`))
	if !respond {
		t.Fatalf("incomplete recall_preview did not produce a response")
	}
	var incompleteEnvelope rpcResponseForTest
	decodeJSONMessage(t, incompleteResp, &incompleteEnvelope)
	if incompleteEnvelope.Error != nil {
		t.Fatalf("incomplete recall_preview should return a tool error, not protocol error: %s", string(incompleteResp))
	}
	if !incompleteEnvelope.Result.IsError {
		t.Fatalf("incomplete recall_preview should fail service validation: %s", string(incompleteResp))
	}
	if !strings.Contains(incompleteEnvelope.Result.Content[0].Text, "actor_id is required") {
		t.Fatalf("expected actor_id validation error, got %s", string(incompleteResp))
	}
}

func TestServerServeStdioRoundtrip(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "note", Text: "stdio ok"}}}})
	server := newProtocolServer(t, surface)
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"prefetch","arguments":{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main"}}}`,
		"",
	}, "\n")
	var out bytes.Buffer

	if err := server.ServeStdio(context.Background(), strings.NewReader(input), &out); err != nil {
		t.Fatalf("ServeStdio returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two JSON-RPC responses, got %d: %q", len(lines), out.String())
	}
	if !strings.Contains(lines[0], `"protocolVersion":"2025-11-25"`) {
		t.Fatalf("first response was not initialize: %s", lines[0])
	}
	if !strings.Contains(lines[1], "stdio ok") {
		t.Fatalf("second response was not tool output: %s", lines[1])
	}
}

func TestServerReturnsProtocolErrorForUnknownMethod(t *testing.T) {
	t.Parallel()

	server := newProtocolServer(t, newTestSurface(t, &fakeService{}))
	raw, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"missing"}`))
	if !respond {
		t.Fatalf("unknown method did not produce a response")
	}
	if !strings.Contains(string(raw), `"code":-32601`) {
		t.Fatalf("expected method-not-found error, got %s", string(raw))
	}
}

func newProtocolServer(t *testing.T, surface *Surface) *Server {
	t.Helper()

	server, err := NewServer(surface)
	if err != nil {
		t.Fatalf("NewServer returned error: %v", err)
	}
	return server
}

func decodeJSONMessage(t *testing.T, raw json.RawMessage, out any) {
	t.Helper()

	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("decode JSON message: %v; raw=%s", err, string(raw))
	}
}

func assertRequiredInputsExact(t *testing.T, tool protocolTool, want ...string) {
	t.Helper()

	required, ok := tool.InputSchema["required"].([]string)
	if !ok {
		t.Fatalf("%s schema missing required input list: %#v", tool.Name, tool.InputSchema)
	}
	if len(required) != len(want) {
		t.Fatalf("%s schema required fields = %#v, want exactly %#v", tool.Name, required, want)
	}
	got := make(map[string]bool, len(required))
	for _, field := range required {
		got[field] = true
	}
	for _, field := range want {
		if !got[field] {
			t.Fatalf("%s schema missing required field %q in %#v", tool.Name, field, required)
		}
	}
}

func callToolMessage(t *testing.T, name string, arguments map[string]any) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      "contract-" + name,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": arguments,
		},
	})
	if err != nil {
		t.Fatalf("marshal call tool message: %v", err)
	}
	return raw
}

func validToolArguments(name string) map[string]any {
	base := map[string]any{
		"tenant_id":    "tenant_1",
		"workspace_id": "workspace_1",
	}
	switch name {
	case "prefetch", "recall_preview", "degraded_status":
		base["session_id"] = "session_1"
		base["actor_id"] = "agent:hermes-main"
	case "sync_turn":
		base["session_id"] = "session_1"
		base["actor_id"] = "agent:hermes-main"
		base["idempotency_key"] = "turn_1"
		base["turn_events"] = []map[string]any{{"event_kind": "message", "payload_json": map[string]any{}}}
	case "add_note":
		base["scope"] = string(core.MemoryScopeWorkspaceShared)
		base["owner_entity_id"] = "agent:hermes-main"
		base["text"] = "Keep Hermes first."
	case "create_plan":
		base["title"] = "Ship trust loop"
		base["scope"] = string(core.MemoryScopeWorkspaceShared)
		base["owner_entity_id"] = "agent:hermes-main"
	case "update_plan":
		base["plan_id"] = "plan_1"
	case "correct_memory":
		base["memory_id"] = "mem_1"
		base["operator_id"] = "operator_1"
		base["idempotency_key"] = "corr_1"
		base["correction_text"] = "Use the corrected fact."
	case "view_timeline":
		base["entity_id"] = "agent:hermes-main"
	case "explain_memory":
		base["memory_id"] = "mem_1"
	}
	return base
}

type rpcResponseForTest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  callToolResult  `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type prefetchValidationService struct {
	fakeService
	assembler     *recall.Assembler
	prefetchCalls int
}

func (s *prefetchValidationService) Prefetch(ctx context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchCalls++
	return s.assembler.Prefetch(ctx, req)
}

type requiredFieldValidationService struct{}

func (requiredFieldValidationService) Prefetch(_ context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"session_id":   req.SessionID,
		"actor_id":     req.ActorID,
	}); err != nil {
		return nil, err
	}
	return &core.PrefetchResponse{Meta: core.RecallMeta{Freshness: "stored"}}, nil
}

func (requiredFieldValidationService) SyncTurn(_ context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"session_id":      req.SessionID,
		"actor_id":        req.ActorID,
		"idempotency_key": req.IdempotencyKey,
	}); err != nil {
		return nil, err
	}
	if len(req.TurnEvents) == 0 {
		return nil, fmt.Errorf("%w: turn_events is required", core.ErrInvalidArgument)
	}
	return &core.SyncTurnResponse{Status: "accepted"}, nil
}

func (requiredFieldValidationService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return &core.AddDocumentResponse{Status: "created"}, nil
}

func (requiredFieldValidationService) SearchMemories(_ context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
	}); err != nil {
		return nil, err
	}
	return &core.SearchMemoriesResponse{}, nil
}

func (requiredFieldValidationService) SearchDocuments(_ context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
	}); err != nil {
		return nil, err
	}
	return &core.SearchDocumentsResponse{}, nil
}

func (requiredFieldValidationService) AddNote(_ context.Context, req *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"scope":           string(req.Scope),
		"owner_entity_id": req.OwnerEntityID,
		"text":            req.Text,
	}); err != nil {
		return nil, err
	}
	return &core.AddNoteResponse{Status: "created"}, nil
}

func (requiredFieldValidationService) CreatePlan(_ context.Context, req *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"title":           req.Title,
		"scope":           string(req.Scope),
		"owner_entity_id": req.OwnerEntityID,
	}); err != nil {
		return nil, err
	}
	return &core.CreatePlanResponse{Status: "created"}, nil
}

func (requiredFieldValidationService) UpdatePlan(_ context.Context, req *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"plan_id":      req.PlanID,
	}); err != nil {
		return nil, err
	}
	return &core.UpdatePlanResponse{Status: "updated"}, nil
}

func (requiredFieldValidationService) CorrectMemory(_ context.Context, req *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"memory_id":       req.MemoryID,
		"operator_id":     req.OperatorID,
		"idempotency_key": req.IdempotencyKey,
		"correction_text": req.CorrectionText,
	}); err != nil {
		return nil, err
	}
	return &core.CorrectMemoryResponse{Status: "applied"}, nil
}

func (requiredFieldValidationService) GetTimeline(_ context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"entity_id":    req.EntityID,
	}); err != nil {
		return nil, err
	}
	return &core.GetTimelineResponse{}, nil
}

func (requiredFieldValidationService) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	if err := requireToolFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"memory_id":    req.MemoryID,
	}); err != nil {
		return nil, err
	}
	return &core.ExplainMemoryResponse{MemoryID: req.MemoryID}, nil
}

func requireToolFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	return nil
}
