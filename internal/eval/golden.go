// ============================================================
// FILE     : internal/eval/golden.go
// PURPOSE  : Runs golden recall scenarios that catch memory quality regressions.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : ScenarioFile, Scenario, Result, Summary, RunFile, RunScenarios
// DEPENDS  : encoding/json, os, strings, internal/core, internal/recall, internal/store
// USED_BY  : cmd/cli, internal/eval tests, Makefile eval target
// ------------------------------------------------------------
// AGENT_NOTE: Eval fixtures must stay deterministic and must not call real Codex or external stores.
// ============================================================

// Package eval runs deterministic, local-only quality checks for recall behavior.
package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
)

// ScenarioFile is the top-level JSON fixture for golden eval scenarios.
type ScenarioFile struct {
	Scenarios              []Scenario              `json:"scenarios"`
	GraphReplayScenarios   []GraphReplayScenario   `json:"graph_replay_scenarios"`
	WorkerBacklogScenarios []WorkerBacklogScenario `json:"worker_backlog_scenarios"`
}

// Scenario defines one deterministic recall quality check.
type Scenario struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Prefetch    core.PrefetchRequest `json:"prefetch"`
	Fixtures    Fixtures             `json:"fixtures"`
	Expect      Expectation          `json:"expect"`
}

// Fixtures contains the in-memory store rows visible to one scenario.
type Fixtures struct {
	Notes            []core.Note                `json:"notes"`
	Plans            []core.Plan                `json:"plans"`
	Profiles         []core.Profile             `json:"profiles"`
	SessionSummaries []core.SessionSummary      `json:"session_summaries"`
	Memories         []core.MemoryResult        `json:"memories"`
	Documents        []core.DocumentChunkResult `json:"documents"`
}

// Expectation describes what a scenario must preserve.
type Expectation struct {
	BlockKinds       []string                   `json:"block_kinds"`
	BlockMetadata    []BlockMetadataExpectation `json:"block_metadata"`
	Contains         []string                   `json:"contains"`
	NotContains      []string                   `json:"not_contains"`
	Sources          []string                   `json:"sources"`
	MaxTokens        int                        `json:"max_tokens"`
	AllowExtraBlocks bool                       `json:"allow_extra_blocks"`
}

// BlockMetadataExpectation asserts operator-visible trust metadata on recall blocks.
type BlockMetadataExpectation struct {
	Kind          string           `json:"kind"`
	Scope         core.MemoryScope `json:"scope,omitempty"`
	Source        string           `json:"source,omitempty"`
	SourceID      string           `json:"source_id,omitempty"`
	Status        string           `json:"status,omitempty"`
	Freshness     string           `json:"freshness,omitempty"`
	OwnerEntityID string           `json:"owner_entity_id,omitempty"`
}

// Result is the pass/fail record for one scenario.
type Result struct {
	Scenario string   `json:"scenario"`
	Passed   bool     `json:"passed"`
	Errors   []string `json:"errors,omitempty"`
	Observed Observed `json:"observed"`
}

// Observed captures the deterministic recall output used for comparison.
type Observed struct {
	BlockKinds []string        `json:"block_kinds"`
	Blocks     []ObservedBlock `json:"blocks"`
	Sources    []string        `json:"sources"`
	Tokens     int             `json:"tokens"`
	Text       string          `json:"text"`
}

// ObservedBlock captures trust metadata for eval failure reporting.
type ObservedBlock struct {
	Kind          string           `json:"kind"`
	Scope         core.MemoryScope `json:"scope,omitempty"`
	Source        string           `json:"source,omitempty"`
	SourceID      string           `json:"source_id,omitempty"`
	Status        string           `json:"status,omitempty"`
	Freshness     string           `json:"freshness,omitempty"`
	OwnerEntityID string           `json:"owner_entity_id,omitempty"`
}

// Summary reports aggregate eval status.
type Summary struct {
	Passed  bool     `json:"passed"`
	Results []Result `json:"results"`
}

// RunFile loads and runs golden scenarios from a JSON fixture file.
func RunFile(ctx context.Context, path string) (*Summary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read golden scenarios: %w", err)
	}
	var file ScenarioFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse golden scenarios: %w", err)
	}
	summary := RunScenarios(ctx, file.Scenarios)
	graphSummary := RunGraphReplayScenarios(ctx, file.GraphReplayScenarios)
	summary.Results = append(summary.Results, graphSummary.Results...)
	summary.Passed = summary.Passed && graphSummary.Passed
	workerSummary := RunWorkerBacklogScenarios(ctx, file.WorkerBacklogScenarios)
	summary.Results = append(summary.Results, workerSummary.Results...)
	summary.Passed = summary.Passed && workerSummary.Passed
	return summary, nil
}

// RunScenarios executes scenarios against the recall assembler with in-memory stores.
func RunScenarios(ctx context.Context, scenarios []Scenario) *Summary {
	results := make([]Result, 0, len(scenarios))
	passed := true
	for _, scenario := range scenarios {
		result := runScenario(ctx, scenario)
		if !result.Passed {
			passed = false
		}
		results = append(results, result)
	}
	return &Summary{Passed: passed, Results: results}
}

func runScenario(ctx context.Context, scenario Scenario) Result {
	errs := validateScenario(scenario)
	assembler := recall.NewAssembler(recall.Dependencies{
		Notes:     noteFixtureStore(scenario.Fixtures.Notes),
		Plans:     planFixtureStore(scenario.Fixtures.Plans),
		Profiles:  profileFixtureStore(scenario.Fixtures.Profiles),
		Summaries: summaryFixtureStore(scenario.Fixtures.SessionSummaries),
		Memories:  memoryFixtureStore(scenario.Fixtures.Memories),
		Documents: documentFixtureStore(scenario.Fixtures.Documents),
	})
	resp, err := assembler.Prefetch(ctx, &scenario.Prefetch)
	if err != nil {
		errs = append(errs, "prefetch failed: "+err.Error())
		return Result{Scenario: scenario.Name, Passed: false, Errors: errs}
	}

	observed := observe(resp)
	errs = append(errs, compareExpectation(scenario.Expect, observed)...)
	return Result{
		Scenario: scenario.Name,
		Passed:   len(errs) == 0,
		Errors:   errs,
		Observed: observed,
	}
}

func validateScenario(scenario Scenario) []string {
	var errs []string
	if strings.TrimSpace(scenario.Name) == "" {
		errs = append(errs, "scenario name is required")
	}
	return errs
}

func observe(resp *core.PrefetchResponse) Observed {
	if resp == nil {
		return Observed{}
	}
	kinds := make([]string, 0, len(resp.Blocks))
	blocks := make([]ObservedBlock, 0, len(resp.Blocks))
	texts := make([]string, 0, len(resp.Blocks))
	for _, block := range resp.Blocks {
		kinds = append(kinds, block.Kind)
		blocks = append(blocks, ObservedBlock{
			Kind:          block.Kind,
			Scope:         block.Scope,
			Source:        block.Source,
			SourceID:      block.SourceID,
			Status:        block.Status,
			Freshness:     block.Freshness,
			OwnerEntityID: block.OwnerEntityID,
		})
		texts = append(texts, block.Text)
	}
	return Observed{
		BlockKinds: kinds,
		Blocks:     blocks,
		Sources:    append([]string(nil), resp.Meta.Sources...),
		Tokens:     resp.Meta.EstimatedTokens,
		Text:       strings.Join(texts, "\n"),
	}
}

func compareExpectation(expect Expectation, observed Observed) []string {
	var errs []string
	if len(expect.BlockKinds) > 0 {
		if expect.AllowExtraBlocks {
			if !hasKindPrefix(observed.BlockKinds, expect.BlockKinds) {
				errs = append(errs, fmt.Sprintf("block kinds prefix got %v want %v", observed.BlockKinds, expect.BlockKinds))
			}
		} else if !reflect.DeepEqual(observed.BlockKinds, expect.BlockKinds) {
			errs = append(errs, fmt.Sprintf("block kinds got %v want %v", observed.BlockKinds, expect.BlockKinds))
		}
	}
	if len(expect.BlockMetadata) > 0 {
		errs = append(errs, compareBlockMetadata(expect.BlockMetadata, observed.Blocks)...)
	}
	for _, needle := range expect.Contains {
		if !strings.Contains(observed.Text, needle) {
			errs = append(errs, fmt.Sprintf("missing expected text %q", needle))
		}
	}
	for _, needle := range expect.NotContains {
		if strings.Contains(observed.Text, needle) {
			errs = append(errs, fmt.Sprintf("unexpected text %q", needle))
		}
	}
	if len(expect.Sources) > 0 && !reflect.DeepEqual(observed.Sources, expect.Sources) {
		errs = append(errs, fmt.Sprintf("sources got %v want %v", observed.Sources, expect.Sources))
	}
	if expect.MaxTokens > 0 && observed.Tokens > expect.MaxTokens {
		errs = append(errs, fmt.Sprintf("tokens got %d max %d", observed.Tokens, expect.MaxTokens))
	}
	return errs
}

func compareBlockMetadata(expect []BlockMetadataExpectation, observed []ObservedBlock) []string {
	if len(observed) < len(expect) {
		return []string{fmt.Sprintf("block metadata count got %d want at least %d", len(observed), len(expect))}
	}
	var errs []string
	for i, want := range expect {
		got := observed[i]
		if want.Kind != "" && got.Kind != want.Kind {
			errs = append(errs, fmt.Sprintf("block[%d] kind got %q want %q", i, got.Kind, want.Kind))
		}
		if want.Scope != "" && got.Scope != want.Scope {
			errs = append(errs, fmt.Sprintf("block[%d] scope got %q want %q", i, got.Scope, want.Scope))
		}
		if want.Source != "" && got.Source != want.Source {
			errs = append(errs, fmt.Sprintf("block[%d] source got %q want %q", i, got.Source, want.Source))
		}
		if want.SourceID != "" && got.SourceID != want.SourceID {
			errs = append(errs, fmt.Sprintf("block[%d] source_id got %q want %q", i, got.SourceID, want.SourceID))
		}
		if want.Status != "" && got.Status != want.Status {
			errs = append(errs, fmt.Sprintf("block[%d] status got %q want %q", i, got.Status, want.Status))
		}
		if want.Freshness != "" && got.Freshness != want.Freshness {
			errs = append(errs, fmt.Sprintf("block[%d] freshness got %q want %q", i, got.Freshness, want.Freshness))
		}
		if want.OwnerEntityID != "" && got.OwnerEntityID != want.OwnerEntityID {
			errs = append(errs, fmt.Sprintf("block[%d] owner_entity_id got %q want %q", i, got.OwnerEntityID, want.OwnerEntityID))
		}
	}
	return errs
}

func hasKindPrefix(got []string, want []string) bool {
	if len(got) < len(want) {
		return false
	}
	return reflect.DeepEqual(got[:len(want)], want)
}

type noteFixtureStore []core.Note

func (s noteFixtureStore) AddNote(context.Context, *core.Note) error {
	return core.ErrNotImplemented
}

func (s noteFixtureStore) ListPinnedNotes(_ context.Context, req *core.ListPinnedNotesRequest) ([]*core.Note, error) {
	notes := make([]*core.Note, 0, len(s))
	for i := range s {
		note := s[i]
		if note.TenantID != req.TenantID || note.WorkspaceID != req.WorkspaceID || !note.Pinned {
			continue
		}
		if !scopeVisible(note.Scope, note.OwnerEntityID, req.OwnerEntityID, req.Scopes) {
			continue
		}
		notes = append(notes, &note)
	}
	return notes, nil
}

type planFixtureStore []core.Plan

func (s planFixtureStore) CreatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s planFixtureStore) UpdatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s planFixtureStore) GetActivePlans(_ context.Context, req *core.GetActivePlansRequest) ([]*core.Plan, error) {
	plans := make([]*core.Plan, 0, len(s))
	for i := range s {
		plan := s[i]
		if plan.TenantID != req.TenantID || plan.WorkspaceID != req.WorkspaceID {
			continue
		}
		if !scopeVisible(plan.Scope, plan.OwnerEntityID, req.OwnerEntityID, req.Scopes) {
			continue
		}
		plans = append(plans, &plan)
	}
	return plans, nil
}

type profileFixtureStore []core.Profile

func (s profileFixtureStore) GetProfile(_ context.Context, entityID string, scope core.MemoryScope) (*core.Profile, error) {
	for i := range s {
		profile := s[i]
		if profile.EntityID == entityID && profile.Scope == scope {
			return &profile, nil
		}
	}
	return nil, core.ErrNotFound
}

func (s profileFixtureStore) UpsertProfile(context.Context, *core.Profile) error {
	return core.ErrNotImplemented
}

type summaryFixtureStore []core.SessionSummary

func (s summaryFixtureStore) UpsertSessionSummary(context.Context, *core.SessionSummary) error {
	return core.ErrNotImplemented
}

func (s summaryFixtureStore) GetSessionSummary(_ context.Context, sessionID string) (*core.SessionSummary, error) {
	for i := range s {
		summary := s[i]
		if summary.SessionID == sessionID {
			return &summary, nil
		}
	}
	return nil, core.ErrNotFound
}

type memoryFixtureStore []core.MemoryResult

func (s memoryFixtureStore) UpsertMemory(context.Context, *core.Memory) error {
	return core.ErrNotImplemented
}

func (s memoryFixtureStore) GetMemory(context.Context, string) (*core.Memory, error) {
	return nil, core.ErrNotImplemented
}

func (s memoryFixtureStore) SearchMemories(_ context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	results := make([]core.MemoryResult, 0, len(s))
	for _, memory := range s {
		if !scopeVisible(memory.Scope, memory.OwnerEntityID, req.OwnerEntityID, req.Scopes) {
			continue
		}
		if !artifactClassVisible(memory.ArtifactClass, req.ArtifactClasses) {
			continue
		}
		if query := strings.TrimSpace(strings.ToLower(req.Query)); query != "" {
			if !strings.Contains(strings.ToLower(memory.Text), query) && !wordOverlap(strings.ToLower(memory.Text), query) {
				continue
			}
		}
		results = append(results, memory)
	}
	return &core.SearchMemoriesResponse{Memories: results}, nil
}

func (s memoryFixtureStore) UpsertMemoryEdge(context.Context, *core.MemoryEdge) error {
	return core.ErrNotImplemented
}

func (s memoryFixtureStore) WriteMemoryTrace(context.Context, *core.MemoryTrace) error {
	return core.ErrNotImplemented
}

func (s memoryFixtureStore) ExplainMemory(context.Context, *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	return nil, core.ErrNotImplemented
}

type documentFixtureStore []core.DocumentChunkResult

func (s documentFixtureStore) AddDocumentWithChunks(context.Context, *core.Document, []*core.DocumentChunk) error {
	return core.ErrNotImplemented
}

func (s documentFixtureStore) AddDocument(context.Context, *core.Document) error {
	return core.ErrNotImplemented
}

func (s documentFixtureStore) AddDocumentChunks(context.Context, []*core.DocumentChunk) error {
	return core.ErrNotImplemented
}

func (s documentFixtureStore) SearchDocuments(_ context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	chunks := make([]core.DocumentChunkResult, 0, len(s))
	query := strings.TrimSpace(strings.ToLower(req.Query))
	for _, chunk := range s {
		if query == "" || strings.Contains(strings.ToLower(chunk.Text), query) || wordOverlap(strings.ToLower(chunk.Text), query) {
			chunks = append(chunks, chunk)
		}
	}
	return &core.SearchDocumentsResponse{Chunks: chunks}, nil
}

func scopeVisible(scope core.MemoryScope, owner string, actor string, scopes []core.MemoryScope) bool {
	if !slices.Contains(scopes, scope) {
		return false
	}
	if scope == core.MemoryScopeAgentPrivate {
		return owner == actor
	}
	return scope != core.MemoryScopeGroupShared
}

func artifactClassVisible(class core.ArtifactClass, classes []core.ArtifactClass) bool {
	return len(classes) == 0 || slices.Contains(classes, class)
}

func wordOverlap(text string, query string) bool {
	for _, word := range strings.Fields(query) {
		word = strings.Trim(word, ".,:;!?()[]{}\"'")
		if len(word) < 4 {
			continue
		}
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}
