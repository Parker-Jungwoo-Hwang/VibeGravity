// ============================================================
// FILE     : internal/reasoning/codex_bridge_test.go
// PURPOSE  : Guards the disabled Codex bridge boundary and strict JSON decoding.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : codex bridge tests
// DEPENDS  : context, encoding/json, errors, strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests use a fake client only; do not call real Codex here.
// ============================================================

package reasoning

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestCodexStage1ExtractorDisabledByDefault(t *testing.T) {
	t.Parallel()

	extractor, err := NewCodexStage1Extractor(CodexBridgeConfig{}, nil)
	if err != nil {
		t.Fatalf("NewCodexStage1Extractor returned error: %v", err)
	}

	output, err := extractor.Extract(context.Background(), testProcessTurnEnvelope().Stage1)
	if err == nil {
		t.Fatalf("expected disabled Codex Stage 1 to fail")
	}
	if !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
	if output.CandidateEntities != nil || output.CandidateMemories != nil {
		t.Fatalf("expected no Stage 1 output when disabled, got %#v", output)
	}
}

func TestCodexStage1ExtractorRejectsUnknownOutputFields(t *testing.T) {
	t.Parallel()

	extractor, err := NewCodexStage1Extractor(CodexBridgeConfig{Enabled: true}, fakeCodexJSONClient{
		resp: CodexResponse{OutputJSON: json.RawMessage(`{
			"candidate_entities": [],
			"candidate_memories": [],
			"summary_hint": "",
			"task_hint": "",
			"freeform": "not allowed"
		}`)},
	})
	if err != nil {
		t.Fatalf("NewCodexStage1Extractor returned error: %v", err)
	}

	_, err = extractor.Extract(context.Background(), testProcessTurnEnvelope().Stage1)
	if err == nil {
		t.Fatalf("expected unknown Stage 1 field to be rejected")
	}
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected strict JSON unknown-field error, got %v", err)
	}
}

func TestCodexStage1ExtractorSendsSchemaMarkedRequest(t *testing.T) {
	t.Parallel()

	client := &capturingCodexJSONClient{
		resp: CodexResponse{OutputJSON: json.RawMessage(`{
			"candidate_entities": [{
				"entity_kind": "agent",
				"display_name": "Hermes",
				"confidence": 0.91,
				"metadata_json": {},
				"source_event_id": "evt_contract_1"
			}],
			"candidate_memories": [{
				"kind": "constraint",
				"artifact_class": "knowledge",
				"scope": "workspace_shared",
				"text": "Reasoning output must stay structured.",
				"confidence": 0.92,
				"raw_event_ids": ["evt_contract_1"]
			}],
			"summary_hint": "structured",
			"task_hint": "resolve graph operations"
		}`)},
	}
	extractor, err := NewCodexStage1Extractor(CodexBridgeConfig{Enabled: true}, client)
	if err != nil {
		t.Fatalf("NewCodexStage1Extractor returned error: %v", err)
	}

	output, err := extractor.Extract(context.Background(), testProcessTurnEnvelope().Stage1)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if client.req.Stage != StageNameExtract {
		t.Fatalf("expected extract stage request, got %#v", client.req)
	}
	if client.req.RequiredOutputSchema != Stage1ExtractOutputSchemaV0 {
		t.Fatalf("expected Stage 1 schema marker, got %q", client.req.RequiredOutputSchema)
	}
	if len(output.CandidateMemories) != 1 {
		t.Fatalf("expected decoded candidate memory, got %#v", output)
	}
}

func TestCodexStage2ResolverPreservesRequiredOutputSchema(t *testing.T) {
	t.Parallel()

	client := &capturingCodexJSONClient{
		resp: CodexResponse{OutputJSON: json.RawMessage(`{
			"operations": [],
			"profile_delta": {},
			"session_summary": "",
			"plan_delta": {},
			"trace": {
				"schema_version": "v0",
				"stage": "resolve",
				"codes": ["test_codex_stage2"],
				"metadata_json": {}
			}
		}`)},
	}
	resolver, err := NewCodexStage2Resolver(CodexBridgeConfig{Enabled: true}, client)
	if err != nil {
		t.Fatalf("NewCodexStage2Resolver returned error: %v", err)
	}

	output, err := resolver.Resolve(context.Background(), testProcessTurnEnvelope().Stage2)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if client.req.Stage != StageNameResolve {
		t.Fatalf("expected resolve stage request, got %#v", client.req)
	}
	if client.req.RequiredOutputSchema != Stage2ResolveOutputSchemaV0 {
		t.Fatalf("expected Stage 2 schema marker to be preserved, got %q", client.req.RequiredOutputSchema)
	}
	if output.Trace.Stage != StageNameResolve {
		t.Fatalf("expected resolved trace output, got %#v", output.Trace)
	}
}

func TestCodexStage2ResolverRejectsInvalidJSONObjects(t *testing.T) {
	t.Parallel()

	resolver, err := NewCodexStage2Resolver(CodexBridgeConfig{Enabled: true}, fakeCodexJSONClient{
		resp: CodexResponse{OutputJSON: json.RawMessage(`{
			"operations": [],
			"profile_delta": [],
			"session_summary": "",
			"plan_delta": {},
			"trace": {
				"schema_version": "v0",
				"stage": "resolve",
				"codes": [],
				"metadata_json": {}
			}
		}`)},
	})
	if err != nil {
		t.Fatalf("NewCodexStage2Resolver returned error: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), testProcessTurnEnvelope().Stage2)
	if err == nil {
		t.Fatalf("expected non-object profile_delta to be rejected")
	}
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if !strings.Contains(err.Error(), "profile_delta") {
		t.Fatalf("expected profile_delta error, got %v", err)
	}
}

func TestCodexStage2ResolverRejectsMismatchedRequiredOutputSchema(t *testing.T) {
	t.Parallel()

	resolver, err := NewCodexStage2Resolver(CodexBridgeConfig{Enabled: true}, fakeCodexJSONClient{})
	if err != nil {
		t.Fatalf("NewCodexStage2Resolver returned error: %v", err)
	}
	input := testProcessTurnEnvelope().Stage2
	input.RequiredOutputSchema = "wrong.schema"

	_, err = resolver.Resolve(context.Background(), input)
	if err == nil {
		t.Fatalf("expected mismatched Stage 2 schema marker to be rejected")
	}
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestMockCodexJSONClientRunsThroughBridgeRunners(t *testing.T) {
	t.Parallel()

	mockClient := NewMockCodexJSONClient()
	cfg := CodexBridgeConfig{Enabled: true}
	stage1, err := NewCodexStage1Extractor(cfg, mockClient)
	if err != nil {
		t.Fatalf("NewCodexStage1Extractor returned error: %v", err)
	}
	stage2, err := NewCodexStage2Resolver(cfg, mockClient)
	if err != nil {
		t.Fatalf("NewCodexStage2Resolver returned error: %v", err)
	}
	orchestrator, err := NewPipelineOrchestrator(stage1, stage2, nil)
	if err != nil {
		t.Fatalf("NewPipelineOrchestrator returned error: %v", err)
	}

	result, err := orchestrator.ProcessTurn(context.Background(), testProcessTurnEnvelope())
	if err != nil {
		t.Fatalf("ProcessTurn returned error: %v", err)
	}
	if result.Stage1.SummaryHint != "mock_codex_stage1_no_candidates" {
		t.Fatalf("expected mocked Stage 1 output, got %#v", result.Stage1)
	}
	if len(result.Stage2.Trace.Codes) != 1 || result.Stage2.Trace.Codes[0] != "mock_codex_bridge_no_operations" {
		t.Fatalf("expected mocked Stage 2 bridge trace, got %#v", result.Stage2.Trace)
	}
}

type fakeCodexJSONClient struct {
	resp CodexResponse
	err  error
}

func (c fakeCodexJSONClient) CompleteJSON(context.Context, CodexRequest) (CodexResponse, error) {
	if c.err != nil {
		return CodexResponse{}, c.err
	}
	return c.resp, nil
}

type capturingCodexJSONClient struct {
	req  CodexRequest
	resp CodexResponse
	err  error
}

func (c *capturingCodexJSONClient) CompleteJSON(_ context.Context, req CodexRequest) (CodexResponse, error) {
	c.req = req
	if c.err != nil {
		return CodexResponse{}, c.err
	}
	return c.resp, nil
}
