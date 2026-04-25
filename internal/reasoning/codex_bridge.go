// ============================================================
// FILE     : internal/reasoning/codex_bridge.go
// PURPOSE  : Defines the disabled-by-default Codex JSON bridge boundary.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : CodexBridgeConfig, CodexJSONClient, CodexRequest, CodexResponse, CodexStage1Extractor, CodexStage2Resolver
// DEPENDS  : bytes, context, encoding/json, fmt, internal/core
// USED_BY  : internal/reasoning tests, future worker wiring
// ------------------------------------------------------------
// AGENT_NOTE: Keep this bridge schema-first; do not add local extraction fallback here.
// ============================================================

package reasoning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// Stage1ExtractOutputSchemaV0 marks the required structured Stage 1 output contract.
const Stage1ExtractOutputSchemaV0 = "stage1.extract.output.v0"

// CodexBridgeConfig controls whether the Codex bridge runners are allowed to call a client.
type CodexBridgeConfig struct {
	Enabled bool
}

// CodexRequest is the schema-marked request sent to a Codex JSON client.
type CodexRequest struct {
	Stage                StageName       `json:"stage"`
	RequiredOutputName   StageName       `json:"required_output_name"`
	RequiredOutputSchema string          `json:"required_output_schema"`
	InputJSON            json.RawMessage `json:"input_json"`
}

// CodexResponse is the raw structured JSON returned by a Codex JSON client.
type CodexResponse struct {
	OutputJSON json.RawMessage `json:"output_json"`
}

// CodexJSONClient is the narrow client boundary for schema-first Codex calls.
type CodexJSONClient interface {
	CompleteJSON(ctx context.Context, req CodexRequest) (CodexResponse, error)
}

// CodexStage1Extractor runs Stage 1 through the Codex JSON client when explicitly enabled.
type CodexStage1Extractor struct {
	cfg    CodexBridgeConfig
	client CodexJSONClient
}

// NewCodexStage1Extractor creates a disabled-by-default Stage 1 Codex runner.
func NewCodexStage1Extractor(cfg CodexBridgeConfig, client CodexJSONClient) (*CodexStage1Extractor, error) {
	if cfg.Enabled && client == nil {
		return nil, fmt.Errorf("%w: enabled codex stage1 extractor requires a client", core.ErrInvalidArgument)
	}
	return &CodexStage1Extractor{cfg: cfg, client: client}, nil
}

// Extract sends Stage 1 input to Codex only when the bridge is explicitly enabled.
func (e *CodexStage1Extractor) Extract(ctx context.Context, input Stage1Input) (Stage1Output, error) {
	if err := validateStage1Input(input); err != nil {
		return Stage1Output{}, err
	}
	if e == nil || !e.cfg.Enabled {
		return Stage1Output{}, fmt.Errorf("%w: codex stage1 extractor is disabled", core.ErrNotImplemented)
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return Stage1Output{}, fmt.Errorf("marshal stage1 input: %w", err)
	}
	resp, err := e.client.CompleteJSON(ctx, CodexRequest{
		Stage:                StageNameExtract,
		RequiredOutputName:   StageNameExtract,
		RequiredOutputSchema: Stage1ExtractOutputSchemaV0,
		InputJSON:            inputJSON,
	})
	if err != nil {
		return Stage1Output{}, fmt.Errorf("codex stage1 complete: %w", err)
	}
	output, err := decodeStrictJSON[Stage1Output]("stage1 output", resp.OutputJSON)
	if err != nil {
		return Stage1Output{}, err
	}
	if err := validateStage1Output(output); err != nil {
		return Stage1Output{}, err
	}
	return output, nil
}

// CodexStage2Resolver runs Stage 2 through the Codex JSON client when explicitly enabled.
type CodexStage2Resolver struct {
	cfg    CodexBridgeConfig
	client CodexJSONClient
}

// NewCodexStage2Resolver creates a disabled-by-default Stage 2 Codex runner.
func NewCodexStage2Resolver(cfg CodexBridgeConfig, client CodexJSONClient) (*CodexStage2Resolver, error) {
	if cfg.Enabled && client == nil {
		return nil, fmt.Errorf("%w: enabled codex stage2 resolver requires a client", core.ErrInvalidArgument)
	}
	return &CodexStage2Resolver{cfg: cfg, client: client}, nil
}

// Resolve sends prepared Stage 2 input to Codex only when the bridge is explicitly enabled.
func (r *CodexStage2Resolver) Resolve(ctx context.Context, input Stage2Input) (Stage2Output, error) {
	if err := validatePreparedStage2Input(input); err != nil {
		return Stage2Output{}, err
	}
	if input.RequiredOutputSchema != Stage2ResolveOutputSchemaV0 {
		return Stage2Output{}, fmt.Errorf("%w: stage2 required output schema must be %q", core.ErrInvalidArgument, Stage2ResolveOutputSchemaV0)
	}
	if r == nil || !r.cfg.Enabled {
		return Stage2Output{}, fmt.Errorf("%w: codex stage2 resolver is disabled", core.ErrNotImplemented)
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return Stage2Output{}, fmt.Errorf("marshal stage2 input: %w", err)
	}
	resp, err := r.client.CompleteJSON(ctx, CodexRequest{
		Stage:                StageNameResolve,
		RequiredOutputName:   input.RequiredOutputName,
		RequiredOutputSchema: input.RequiredOutputSchema,
		InputJSON:            inputJSON,
	})
	if err != nil {
		return Stage2Output{}, fmt.Errorf("codex stage2 complete: %w", err)
	}
	output, err := decodeStrictJSON[Stage2Output]("stage2 output", resp.OutputJSON)
	if err != nil {
		return Stage2Output{}, err
	}
	if err := validateStage2Output(output); err != nil {
		return Stage2Output{}, err
	}
	if err := validateStage2OutputJSONFields(output); err != nil {
		return Stage2Output{}, err
	}
	return output, nil
}

func validateStage1Input(input Stage1Input) error {
	if input.JobID == "" {
		return fmt.Errorf("%w: stage1 job_id is required", core.ErrInvalidArgument)
	}
	if input.TenantID == "" {
		return fmt.Errorf("%w: stage1 tenant_id is required", core.ErrInvalidArgument)
	}
	if input.WorkspaceID == "" {
		return fmt.Errorf("%w: stage1 workspace_id is required", core.ErrInvalidArgument)
	}
	if len(input.RawEvents) == 0 {
		return fmt.Errorf("%w: stage1 raw event bundle is required", core.ErrInvalidArgument)
	}
	return nil
}

func validateStage1Output(output Stage1Output) error {
	for i, entity := range output.CandidateEntities {
		if entity.EntityKind == "" {
			return fmt.Errorf("%w: stage1 candidate_entities[%d].entity_kind is required", core.ErrInvalidArgument, i)
		}
		if entity.DisplayName == "" {
			return fmt.Errorf("%w: stage1 candidate_entities[%d].display_name is required", core.ErrInvalidArgument, i)
		}
		if !validConfidence(entity.Confidence) {
			return fmt.Errorf("%w: stage1 candidate_entities[%d].confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, i)
		}
		if err := validateJSONObject(fmt.Sprintf("stage1 candidate_entities[%d].metadata_json", i), entity.MetadataJSON); err != nil {
			return err
		}
		if entity.SourceEventID == "" {
			return fmt.Errorf("%w: stage1 candidate_entities[%d].source_event_id is required", core.ErrInvalidArgument, i)
		}
	}
	for i, memory := range output.CandidateMemories {
		if memory.Kind == "" {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].kind is required", core.ErrInvalidArgument, i)
		}
		if memory.ArtifactClass == "" {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].artifact_class is required", core.ErrInvalidArgument, i)
		}
		if memory.Scope == "" {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].scope is required", core.ErrInvalidArgument, i)
		}
		if memory.Text == "" {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].text is required", core.ErrInvalidArgument, i)
		}
		if !validConfidence(memory.Confidence) {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, i)
		}
		if len(memory.RawEventIDs) == 0 {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].raw_event_ids are required", core.ErrInvalidArgument, i)
		}
		for _, rawEventID := range memory.RawEventIDs {
			if rawEventID == "" {
				return fmt.Errorf("%w: stage1 candidate_memories[%d].raw_event_ids cannot contain empty ids", core.ErrInvalidArgument, i)
			}
		}
	}
	return nil
}

func validateStage2OutputJSONFields(output Stage2Output) error {
	if err := validateJSONObject("stage2 profile_delta", output.ProfileDelta); err != nil {
		return err
	}
	if err := validateJSONObject("stage2 plan_delta", output.PlanDelta); err != nil {
		return err
	}
	if err := validateJSONObject("stage2 trace.metadata_json", output.Trace.MetadataJSON); err != nil {
		return err
	}
	for i, operation := range output.Operations {
		if err := validateJSONObject(fmt.Sprintf("stage2 operations[%d].metadata", i), operation.Metadata); err != nil {
			return err
		}
		if operation.Memory != nil {
			if err := validateJSONObject(fmt.Sprintf("stage2 operations[%d].memory.metadata_json", i), operation.Memory.MetadataJSON); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeStrictJSON[T any](label string, raw json.RawMessage) (T, error) {
	var value T
	if len(bytes.TrimSpace(raw)) == 0 {
		return value, fmt.Errorf("%w: %s JSON is required", core.ErrInvalidArgument, label)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("%w: decode %s: %v", core.ErrInvalidArgument, label, err)
	}
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return value, fmt.Errorf("%w: decode %s: trailing JSON is not allowed", core.ErrInvalidArgument, label)
	}
	return value, nil
}

func validateJSONObject(label string, raw json.RawMessage) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		return fmt.Errorf("%w: %s must be a JSON object", core.ErrInvalidArgument, label)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return fmt.Errorf("%w: %s must be a JSON object", core.ErrInvalidArgument, label)
	}
	return nil
}

func validConfidence(confidence float64) bool {
	return confidence > 0 && confidence <= 1
}
