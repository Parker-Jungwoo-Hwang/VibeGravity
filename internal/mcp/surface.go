// ============================================================
// FILE     : internal/mcp/surface.go
// PURPOSE  : Exposes VibeGravity core operations as a small MCP-style tool surface.
// LAYER    : interface
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Surface, Tool, NewSurface
// DEPENDS  : context, encoding/json, fmt, internal/core
// USED_BY  : MCP integration tests, future protocol server
// ------------------------------------------------------------
// AGENT_NOTE: This surface must stay a thin adapter over core service semantics.
// ============================================================

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// Tool describes one MCP-visible VibeGravity operation.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Surface exposes core VibeGravity operations through tool names.
type Surface struct {
	service core.VibeGravityService
}

// NewSurface creates an MCP-style adapter over the shared core service.
func NewSurface(service core.VibeGravityService) (*Surface, error) {
	if service == nil {
		return nil, fmt.Errorf("%w: mcp service is required", core.ErrInvalidArgument)
	}
	return &Surface{service: service}, nil
}

// Tools lists the current MCP tool surface.
func (s *Surface) Tools() []Tool {
	return []Tool{
		{Name: "prefetch", Description: "Assemble a typed recall pack."},
		{Name: "recall_preview", Description: "Preview scoped recall with source and freshness metadata."},
		{Name: "sync_turn", Description: "Record raw turn events and enqueue memory processing."},
		{Name: "search_memory", Description: "Search visible memories."},
		{Name: "search_documents", Description: "Search document chunks."},
		{Name: "add_note", Description: "Create a human-authored recall note."},
		{Name: "create_plan", Description: "Create a structured plan."},
		{Name: "update_plan", Description: "Update a structured plan."},
		{Name: "correct_memory", Description: "Record a correction intent."},
		{Name: "view_timeline", Description: "Read memory and correction activity."},
		{Name: "explain_memory", Description: "Read provenance for one memory."},
		{Name: "degraded_status", Description: "Show whether recall is fresh or degraded."},
	}
}

// Call decodes a JSON tool request, delegates to the core service, and returns JSON.
func (s *Surface) Call(ctx context.Context, name string, input json.RawMessage) (json.RawMessage, error) {
	if s == nil || s.service == nil {
		return nil, fmt.Errorf("%w: mcp surface is not initialized", core.ErrInvalidArgument)
	}
	switch name {
	case "prefetch", "recall_preview":
		var req core.PrefetchRequest
		return callJSON(ctx, input, &req, s.service.Prefetch)
	case "sync_turn":
		var req core.SyncTurnRequest
		return callJSON(ctx, input, &req, s.service.SyncTurn)
	case "search_memory":
		var req core.SearchMemoriesRequest
		return callJSON(ctx, input, &req, s.service.SearchMemories)
	case "search_documents":
		var req core.SearchDocumentsRequest
		return callJSON(ctx, input, &req, s.service.SearchDocuments)
	case "add_note":
		var req core.AddNoteRequest
		return callJSON(ctx, input, &req, s.service.AddNote)
	case "create_plan":
		var req core.CreatePlanRequest
		return callJSON(ctx, input, &req, s.service.CreatePlan)
	case "update_plan":
		var req core.UpdatePlanRequest
		return callJSON(ctx, input, &req, s.service.UpdatePlan)
	case "correct_memory":
		var req core.CorrectMemoryRequest
		return callJSON(ctx, input, &req, s.service.CorrectMemory)
	case "view_timeline":
		var req core.GetTimelineRequest
		return callJSON(ctx, input, &req, s.service.GetTimeline)
	case "explain_memory":
		var req core.ExplainMemoryRequest
		return callJSON(ctx, input, &req, s.service.ExplainMemory)
	case "degraded_status":
		var req core.PrefetchRequest
		return callJSON(ctx, input, &req, func(ctx context.Context, req *core.PrefetchRequest) (*core.RecallMeta, error) {
			resp, err := s.service.Prefetch(ctx, req)
			if err != nil {
				return nil, err
			}
			if resp == nil {
				return &core.RecallMeta{}, nil
			}
			return &resp.Meta, nil
		})
	default:
		return nil, fmt.Errorf("%w: unknown mcp tool %q", core.ErrInvalidArgument, name)
	}
}

func callJSON[Req any, Resp any](ctx context.Context, input json.RawMessage, req *Req, call func(context.Context, *Req) (*Resp, error)) (json.RawMessage, error) {
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	if err := json.Unmarshal(input, req); err != nil {
		return nil, fmt.Errorf("%w: decode mcp tool input: %v", core.ErrInvalidArgument, err)
	}
	resp, err := call(ctx, req)
	if err != nil {
		return nil, err
	}
	output, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("encode mcp tool output: %w", err)
	}
	return output, nil
}
