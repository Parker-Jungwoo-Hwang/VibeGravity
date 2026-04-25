// ============================================================
// FILE     : internal/reasoning/mock_codex_client.go
// PURPOSE  : Provides a deterministic mocked Codex JSON client for the worker reasoning bridge.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : MockCodexJSONClient, NewMockCodexJSONClient
// DEPENDS  : context, encoding/json, fmt, internal/core
// USED_BY  : cmd/worker, internal/reasoning tests
// ------------------------------------------------------------
// AGENT_NOTE: This mock must exercise the Codex bridge interface without becoming a local extractor.
// ============================================================

package reasoning

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// MockCodexJSONClient is a deterministic local client for exercising the Codex bridge boundary.
type MockCodexJSONClient struct{}

// NewMockCodexJSONClient creates a mocked Codex JSON client for non-networked worker wiring.
func NewMockCodexJSONClient() *MockCodexJSONClient {
	return &MockCodexJSONClient{}
}

// CompleteJSON returns schema-marked structured JSON without calling a real Codex API.
func (c *MockCodexJSONClient) CompleteJSON(_ context.Context, req CodexRequest) (CodexResponse, error) {
	if c == nil {
		return CodexResponse{}, fmt.Errorf("%w: mock codex client is required", core.ErrInvalidArgument)
	}
	switch req.Stage {
	case StageNameExtract:
		return c.completeStage1(req)
	case StageNameResolve:
		return c.completeStage2(req)
	default:
		return CodexResponse{}, fmt.Errorf("%w: unsupported mock codex stage %q", core.ErrInvalidArgument, req.Stage)
	}
}

func (c *MockCodexJSONClient) completeStage1(req CodexRequest) (CodexResponse, error) {
	if req.RequiredOutputName != StageNameExtract {
		return CodexResponse{}, fmt.Errorf("%w: mock stage1 required output name must be extract", core.ErrInvalidArgument)
	}
	if req.RequiredOutputSchema != Stage1ExtractOutputSchemaV0 {
		return CodexResponse{}, fmt.Errorf("%w: mock stage1 required output schema must be %q", core.ErrInvalidArgument, Stage1ExtractOutputSchemaV0)
	}
	input, err := decodeStrictJSON[Stage1Input]("mock stage1 input", req.InputJSON)
	if err != nil {
		return CodexResponse{}, err
	}
	if err := validateStage1Input(input); err != nil {
		return CodexResponse{}, err
	}
	output := Stage1Output{
		CandidateEntities: []CandidateEntity{},
		CandidateMemories: []CandidateMemory{},
		SummaryHint:       "mock_codex_stage1_no_candidates",
		TaskHint:          "mock_codex_stage2_no_operations",
	}
	return marshalMockCodexOutput("mock stage1 output", output)
}

func (c *MockCodexJSONClient) completeStage2(req CodexRequest) (CodexResponse, error) {
	if req.RequiredOutputName != StageNameResolve {
		return CodexResponse{}, fmt.Errorf("%w: mock stage2 required output name must be resolve", core.ErrInvalidArgument)
	}
	if req.RequiredOutputSchema != Stage2ResolveOutputSchemaV0 {
		return CodexResponse{}, fmt.Errorf("%w: mock stage2 required output schema must be %q", core.ErrInvalidArgument, Stage2ResolveOutputSchemaV0)
	}
	input, err := decodeStrictJSON[Stage2Input]("mock stage2 input", req.InputJSON)
	if err != nil {
		return CodexResponse{}, err
	}
	if err := validatePreparedStage2Input(input); err != nil {
		return CodexResponse{}, err
	}
	output := emptyStage2Output()
	output.Trace.Codes = []string{"mock_codex_bridge_no_operations"}
	output.Trace.MetadataJSON = json.RawMessage(`{"client":"mock_codex_json_client"}`)
	return marshalMockCodexOutput("mock stage2 output", output)
}

func marshalMockCodexOutput(label string, output any) (CodexResponse, error) {
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return CodexResponse{}, fmt.Errorf("marshal %s: %w", label, err)
	}
	return CodexResponse{OutputJSON: outputJSON}, nil
}
