// ============================================================
// FILE     : internal/hermes/provider.go
// PURPOSE  : Maps Hermes memory-provider lifecycle hooks to VibeGravity core service calls.
// LAYER    : interface
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Provider, ProviderTool, NewProvider
// DEPENDS  : context, fmt, strings, internal/core
// USED_BY  : Hermes integration tests, future plugin runtime
// ------------------------------------------------------------
// AGENT_NOTE: Keep this adapter thin so Hermes, HTTP, and MCP share the same core semantics.
// ============================================================

package hermes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// Provider maps Hermes memory-provider lifecycle hooks to the core service.
type Provider struct {
	service core.VibeGravityService
}

// ProviderTool describes one operator tool exposed through the Hermes provider.
type ProviderTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewProvider creates a Hermes-facing adapter over the shared VibeGravity service.
func NewProvider(service core.VibeGravityService) (*Provider, error) {
	if service == nil {
		return nil, fmt.Errorf("%w: hermes provider service is required", core.ErrInvalidArgument)
	}
	return &Provider{service: service}, nil
}

// IsAvailable checks whether the backing VibeGravity service can answer a minimal prefetch.
func (p *Provider) IsAvailable(ctx context.Context, req *core.PrefetchRequest) bool {
	if p == nil || p.service == nil || req == nil {
		return false
	}
	_, err := p.service.Prefetch(ctx, req)
	return err == nil
}

// Prefetch calls the core recall assembler before a Hermes turn.
func (p *Provider) Prefetch(ctx context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	if p == nil || p.service == nil {
		return nil, fmt.Errorf("%w: hermes provider is not initialized", core.ErrInvalidArgument)
	}
	return p.service.Prefetch(ctx, req)
}

// SyncTurn records a completed Hermes turn through the hot ingest path.
func (p *Provider) SyncTurn(ctx context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	if p == nil || p.service == nil {
		return nil, fmt.Errorf("%w: hermes provider is not initialized", core.ErrInvalidArgument)
	}
	return p.service.SyncTurn(ctx, req)
}

// RenderContext turns typed recall blocks into compact Hermes text context.
func (p *Provider) RenderContext(resp *core.PrefetchResponse) string {
	if resp == nil || len(resp.Blocks) == 0 {
		return ""
	}
	lines := make([]string, 0, len(resp.Blocks))
	for _, block := range resp.Blocks {
		text := strings.TrimSpace(block.Text)
		if text == "" {
			continue
		}
		kind := strings.TrimSpace(block.Kind)
		if kind == "" {
			kind = "memory"
		}
		labels := []string{kind, fmt.Sprintf("%d", block.Priority)}
		if block.Scope != "" {
			labels = append(labels, string(block.Scope))
		}
		if source := strings.TrimSpace(block.Source); source != "" {
			labels = append(labels, source)
		}
		if freshness := strings.TrimSpace(block.Freshness); freshness != "" {
			labels = append(labels, freshness)
		}
		lines = append(lines, fmt.Sprintf("[%s] %s", strings.Join(labels, ":"), text))
	}
	return strings.Join(lines, "\n")
}

// GetTools lists the minimum Hermes provider tools backed by the v1 core API.
func (p *Provider) GetTools() []ProviderTool {
	return []ProviderTool{
		{Name: "recall_preview", Description: "Preview the scoped recall Hermes will receive."},
		{Name: "search_memory", Description: "Search visible VibeGravity memories."},
		{Name: "add_note", Description: "Create a human-authored recall control note."},
		{Name: "show_plan", Description: "Show active structured plans."},
		{Name: "explain_memory", Description: "Show provenance for a remembered item."},
		{Name: "correct_memory", Description: "Record a human correction for a memory."},
		{Name: "view_timeline", Description: "View memory and correction activity."},
		{Name: "degraded_status", Description: "Show whether recall is fresh or degraded."},
	}
}

// CallTool dispatches a Hermes provider tool to the shared core service.
func (p *Provider) CallTool(ctx context.Context, name string, input json.RawMessage) (json.RawMessage, error) {
	if p == nil || p.service == nil {
		return nil, fmt.Errorf("%w: hermes provider is not initialized", core.ErrInvalidArgument)
	}
	switch name {
	case "recall_preview":
		var req core.PrefetchRequest
		return callJSON(ctx, input, &req, p.service.Prefetch)
	case "search_memory":
		var req core.SearchMemoriesRequest
		return callJSON(ctx, input, &req, p.service.SearchMemories)
	case "add_note":
		var req core.AddNoteRequest
		return callJSON(ctx, input, &req, p.service.AddNote)
	case "explain_memory":
		var req core.ExplainMemoryRequest
		return callJSON(ctx, input, &req, p.service.ExplainMemory)
	case "correct_memory":
		var req core.CorrectMemoryRequest
		return callJSON(ctx, input, &req, p.service.CorrectMemory)
	case "view_timeline":
		var req core.GetTimelineRequest
		return callJSON(ctx, input, &req, p.service.GetTimeline)
	case "degraded_status":
		var req core.PrefetchRequest
		return callJSON(ctx, input, &req, func(ctx context.Context, req *core.PrefetchRequest) (*core.RecallMeta, error) {
			resp, err := p.service.Prefetch(ctx, req)
			if err != nil {
				return nil, err
			}
			if resp == nil {
				return &core.RecallMeta{}, nil
			}
			return &resp.Meta, nil
		})
	case "show_plan":
		return nil, fmt.Errorf("%w: hermes provider tool %q needs a read-only plan API", core.ErrNotImplemented, name)
	default:
		return nil, fmt.Errorf("%w: unknown hermes provider tool %q", core.ErrInvalidArgument, name)
	}
}

// OnSessionEnd records a session-end hint for future dreaming integration.
func (p *Provider) OnSessionEnd(context.Context, string) error {
	if p == nil || p.service == nil {
		return fmt.Errorf("%w: hermes provider is not initialized", core.ErrInvalidArgument)
	}
	return nil
}

func callJSON[Req any, Resp any](ctx context.Context, input json.RawMessage, req *Req, call func(context.Context, *Req) (*Resp, error)) (json.RawMessage, error) {
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	if err := json.Unmarshal(input, req); err != nil {
		return nil, fmt.Errorf("%w: decode hermes tool input: %v", core.ErrInvalidArgument, err)
	}
	resp, err := call(ctx, req)
	if err != nil {
		return nil, err
	}
	output, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("encode hermes tool output: %w", err)
	}
	return output, nil
}
