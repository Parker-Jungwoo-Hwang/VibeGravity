package core

import (
	"encoding/json"
	"time"
)

// Profile is a rebuildable static and dynamic snapshot for an entity.
type Profile struct {
	EntityID        string          `json:"entity_id"`
	Scope           MemoryScope     `json:"scope"`
	StaticJSON      json.RawMessage `json:"static_json"`
	DynamicJSON     json.RawMessage `json:"dynamic_json"`
	SourceMemoryIDs []string        `json:"source_memory_ids"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Version         int64           `json:"version"`
}
