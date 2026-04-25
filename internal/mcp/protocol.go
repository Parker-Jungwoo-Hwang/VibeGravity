// ============================================================
// FILE     : internal/mcp/protocol.go
// PURPOSE  : Serves the VibeGravity MCP tool surface over JSON-RPC transports.
// LAYER    : interface
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : ProtocolVersion, Server, NewServer
// DEPENDS  : internal/mcp/surface.go, encoding/json, bufio, io
// USED_BY  : cmd/cli, MCP protocol roundtrip tests
// ------------------------------------------------------------
// AGENT_NOTE: Stdout must contain only newline-delimited JSON-RPC messages.
// ============================================================

package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// ProtocolVersion is the MCP version this server advertises.
const ProtocolVersion = "2025-11-25"

const (
	jsonRPCVersion = "2.0"

	errParseError     = -32700
	errInvalidRequest = -32600
	errMethodNotFound = -32601
	errInvalidParams  = -32602
	errInternalError  = -32603
)

// Server handles MCP JSON-RPC messages for a Surface.
type Server struct {
	surface *Surface
}

// NewServer creates an MCP protocol server over the shared tool surface.
func NewServer(surface *Surface) (*Server, error) {
	if surface == nil {
		return nil, fmt.Errorf("mcp surface is required")
	}
	return &Server{surface: surface}, nil
}

// ServeStdio serves newline-delimited MCP JSON-RPC over stdin/stdout style streams.
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	if s == nil || s.surface == nil {
		return fmt.Errorf("mcp server is not initialized")
	}
	scanner := bufio.NewScanner(in)
	const maxMessageBytes = 1024 * 1024
	scanner.Buffer(make([]byte, 0, 64*1024), maxMessageBytes)
	encoder := json.NewEncoder(out)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		resp, respond := s.HandleMessage(ctx, line)
		if !respond {
			continue
		}
		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("write mcp response: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read mcp input: %w", err)
	}
	return nil
}

// HandleMessage handles one MCP JSON-RPC message. Notifications return respond=false.
func (s *Server) HandleMessage(ctx context.Context, raw json.RawMessage) (json.RawMessage, bool) {
	var req rpcRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return marshalRPCError(nil, errParseError, "Parse error", nil), true
	}
	if req.JSONRPC != jsonRPCVersion || req.Method == "" {
		return marshalRPCError(req.ID, errInvalidRequest, "Invalid Request", nil), true
	}
	if len(req.ID) == 0 {
		if req.Method == "notifications/initialized" {
			return nil, false
		}
		return nil, false
	}

	switch req.Method {
	case "initialize":
		return marshalRPCResult(req.ID, initializeResult{
			ProtocolVersion: ProtocolVersion,
			Capabilities: serverCapabilities{
				Tools: toolsCapability{ListChanged: false},
			},
			ServerInfo: implementationInfo{
				Name:        "vibegravity",
				Title:       "VibeGravity MCP Server",
				Version:     "0.1.0",
				Description: "Hermes-first shared memory kernel tools.",
			},
			Instructions: "Use VibeGravity tools for memory recall, sync, notes, plans, corrections, and timeline visibility.",
		}), true
	case "ping":
		return marshalRPCResult(req.ID, map[string]any{}), true
	case "tools/list":
		return marshalRPCResult(req.ID, listToolsResult{Tools: s.protocolTools()}), true
	case "tools/call":
		result, err := s.callTool(ctx, req.Params)
		if err != nil {
			return marshalRPCError(req.ID, errInvalidParams, err.Error(), nil), true
		}
		return marshalRPCResult(req.ID, result), true
	default:
		return marshalRPCError(req.ID, errMethodNotFound, "Method not found", map[string]string{"method": req.Method}), true
	}
}

func (s *Server) protocolTools() []protocolTool {
	tools := s.surface.Tools()
	out := make([]protocolTool, 0, len(tools))
	for _, tool := range tools {
		out = append(out, protocolTool{
			Name:        tool.Name,
			Title:       tool.Name,
			Description: tool.Description,
			InputSchema: toolInputSchema(tool.Name),
		})
	}
	return out
}

func toolInputSchema(name string) map[string]any {
	base := func(required []string, properties map[string]any) map[string]any {
		schema := map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           properties,
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	}
	stringProp := func(description string) map[string]any {
		return map[string]any{"type": "string", "description": description}
	}
	intProp := func(description string) map[string]any {
		return map[string]any{"type": "integer", "description": description}
	}
	boolProp := func(description string) map[string]any {
		return map[string]any{"type": "boolean", "description": description}
	}
	jsonProp := func(description string) map[string]any {
		return map[string]any{"description": description}
	}
	stringArrayProp := func(description string) map[string]any {
		return map[string]any{
			"type":        "array",
			"description": description,
			"items":       map[string]any{"type": "string"},
		}
	}

	scopeArray := stringArrayProp("Visible memory scopes such as agent_private, workspace_shared, group_shared, or session_scratch.")
	switch name {
	case "prefetch", "recall_preview", "degraded_status":
		return base([]string{"tenant_id", "workspace_id", "session_id", "actor_id"}, map[string]any{
			"tenant_id":     stringProp("Tenant identifier."),
			"workspace_id":  stringProp("Workspace identifier."),
			"session_id":    stringProp("Session identifier."),
			"actor_id":      stringProp("Actor requesting recall."),
			"query":         stringProp("Question or task used to assemble recall."),
			"budget_tokens": intProp("Maximum approximate recall token budget."),
			"mode":          stringProp("Recall mode such as default, small, or rich."),
		})
	case "sync_turn":
		return base([]string{"tenant_id", "workspace_id", "session_id", "actor_id", "idempotency_key", "turn_events"}, map[string]any{
			"tenant_id":       stringProp("Tenant identifier."),
			"workspace_id":    stringProp("Workspace identifier."),
			"session_id":      stringProp("Session identifier."),
			"actor_id":        stringProp("Actor that produced the turn."),
			"idempotency_key": stringProp("Request-level idempotency key."),
			"turn_events": map[string]any{
				"type":        "array",
				"description": "Raw turn events to record.",
				"minItems":    1,
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"event_kind"},
					"properties": map[string]any{
						"event_kind":   stringProp("Raw event kind."),
						"source":       stringProp("Producer source."),
						"fingerprint":  stringProp("Optional event fingerprint."),
						"occurred_at":  stringProp("RFC3339 event timestamp."),
						"payload_json": jsonProp("Raw event payload object."),
					},
				},
			},
		})
	case "search_memory":
		return base([]string{"tenant_id", "workspace_id"}, map[string]any{
			"tenant_id":         stringProp("Tenant identifier."),
			"workspace_id":      stringProp("Workspace identifier."),
			"owner_entity_id":   stringProp("Actor used for agent_private filtering."),
			"visible_group_ids": stringArrayProp("Group identifiers visible to the actor."),
			"query":             stringProp("Memory search query."),
			"scopes":            scopeArray,
			"artifact_classes":  stringArrayProp("Memory artifact classes to include."),
		})
	case "search_documents":
		return base([]string{"tenant_id", "workspace_id"}, map[string]any{
			"tenant_id":    stringProp("Tenant identifier."),
			"workspace_id": stringProp("Workspace identifier."),
			"query":        stringProp("Document search query."),
		})
	case "add_note":
		return base([]string{"tenant_id", "workspace_id", "scope", "owner_entity_id", "text"}, map[string]any{
			"tenant_id":       stringProp("Tenant identifier."),
			"workspace_id":    stringProp("Workspace identifier."),
			"note_kind":       stringProp("Note kind."),
			"scope":           stringProp("Memory scope for the note."),
			"owner_entity_id": stringProp("Owner required by the note contract."),
			"text":            stringProp("Note text."),
			"pinned":          boolProp("Whether the note should receive recall priority."),
			"expires_at":      stringProp("Optional RFC3339 expiration timestamp."),
		})
	case "create_plan":
		return base([]string{"tenant_id", "workspace_id", "title", "scope", "owner_entity_id"}, map[string]any{
			"tenant_id":       stringProp("Tenant identifier."),
			"workspace_id":    stringProp("Workspace identifier."),
			"title":           stringProp("Plan title."),
			"status":          stringProp("Plan status."),
			"scope":           stringProp("Plan visibility scope."),
			"owner_entity_id": stringProp("Owner required by the plan contract."),
			"evidence_json":   jsonProp("Optional structured evidence."),
			"items":           planItemsSchema(),
		})
	case "update_plan":
		return base([]string{"tenant_id", "workspace_id", "plan_id"}, map[string]any{
			"tenant_id":     stringProp("Tenant identifier."),
			"workspace_id":  stringProp("Workspace identifier."),
			"plan_id":       stringProp("Plan identifier."),
			"title":         stringProp("Optional replacement title."),
			"status":        stringProp("Optional replacement status."),
			"evidence_json": jsonProp("Optional structured evidence."),
			"items":         planItemsSchema(),
		})
	case "correct_memory":
		return base([]string{"tenant_id", "workspace_id", "memory_id", "operator_id", "idempotency_key", "correction_text"}, map[string]any{
			"tenant_id":       stringProp("Tenant identifier."),
			"workspace_id":    stringProp("Workspace identifier."),
			"memory_id":       stringProp("Memory being corrected."),
			"operator_id":     stringProp("Human or operator actor applying the correction."),
			"idempotency_key": stringProp("Correction idempotency key."),
			"correction_text": stringProp("Replacement truth or correction instruction."),
			"evidence_json":   jsonProp("Optional correction evidence."),
		})
	case "view_timeline":
		return base([]string{"tenant_id", "workspace_id", "entity_id"}, map[string]any{
			"tenant_id":    stringProp("Tenant identifier."),
			"workspace_id": stringProp("Workspace identifier."),
			"scopes":       scopeArray,
			"entity_id":    stringProp("Actor used for private timeline filtering."),
			"from":         stringProp("Optional RFC3339 lower time bound."),
			"to":           stringProp("Optional RFC3339 upper time bound."),
			"limit":        intProp("Maximum number of timeline items."),
		})
	case "explain_memory":
		return base([]string{"tenant_id", "workspace_id", "memory_id"}, map[string]any{
			"tenant_id":         stringProp("Tenant identifier."),
			"workspace_id":      stringProp("Workspace identifier."),
			"memory_id":         stringProp("Memory identifier to explain."),
			"entity_id":         stringProp("Actor used for private memory visibility."),
			"visible_group_ids": stringArrayProp("Group identifiers visible to the actor."),
		})
	default:
		return map[string]any{"type": "object"}
	}
}

func planItemsSchema() map[string]any {
	return map[string]any{
		"type":        "array",
		"description": "Structured plan items.",
		"items": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"id":            map[string]any{"type": "string", "description": "Existing item identifier."},
				"title":         map[string]any{"type": "string", "description": "Item title."},
				"status":        map[string]any{"type": "string", "description": "Item status."},
				"evidence_json": map[string]any{"description": "Optional structured evidence."},
			},
		},
	}
}

func (s *Server) callTool(ctx context.Context, params json.RawMessage) (callToolResult, error) {
	var req callToolRequest
	if len(params) == 0 {
		return callToolResult{}, errors.New("tools/call params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return callToolResult{}, fmt.Errorf("decode tools/call params: %w", err)
	}
	if req.Name == "" {
		return callToolResult{}, errors.New("tools/call name is required")
	}
	if len(req.Arguments) == 0 {
		req.Arguments = json.RawMessage(`{}`)
	}
	raw, err := s.surface.Call(ctx, req.Name, req.Arguments)
	if err != nil {
		return callToolResult{
			Content: []textContent{{Type: "text", Text: err.Error()}},
			IsError: true,
		}, nil
	}
	var structured map[string]any
	if err := json.Unmarshal(raw, &structured); err != nil {
		return callToolResult{}, fmt.Errorf("decode tool output: %w", err)
	}
	return callToolResult{
		Content:           []textContent{{Type: "text", Text: string(raw)}},
		StructuredContent: structured,
		IsError:           false,
	}, nil
}

func marshalRPCResult(id json.RawMessage, result any) json.RawMessage {
	resp := rpcResponse{JSONRPC: jsonRPCVersion, ID: id, Result: result}
	raw, err := json.Marshal(resp)
	if err != nil {
		return marshalRPCError(id, errInternalError, "Internal error", nil)
	}
	return raw
}

func marshalRPCError(id json.RawMessage, code int, message string, data any) json.RawMessage {
	resp := rpcResponse{JSONRPC: jsonRPCVersion, ID: id, Error: &rpcError{Code: code, Message: message, Data: data}}
	raw, _ := json.Marshal(resp)
	return raw
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type implementationInfo struct {
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type initializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    serverCapabilities `json:"capabilities"`
	ServerInfo      implementationInfo `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

type serverCapabilities struct {
	Tools toolsCapability `json:"tools"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type listToolsResult struct {
	Tools []protocolTool `json:"tools"`
}

type protocolTool struct {
	Name        string         `json:"name"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

type callToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type callToolResult struct {
	Content           []textContent  `json:"content"`
	StructuredContent map[string]any `json:"structuredContent,omitempty"`
	IsError           bool           `json:"isError"`
}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
