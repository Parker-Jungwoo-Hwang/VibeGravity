// ============================================================
// FILE     : internal/store/postgres/helpers.go
// PURPOSE  : Provides shared PostgreSQL store helpers for IDs, JSON, and row scanning.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : none
// DEPENDS  : crypto/rand, encoding/hex, encoding/json, time
// USED_BY  : internal/store/postgres implementations
// ------------------------------------------------------------
// AGENT_NOTE: Keep helper behavior deterministic where callers already provide stable IDs.
// ============================================================

package postgres

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

func newID(prefix string) (string, error) {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(b[:]), nil
}

func rawJSONOrEmpty(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func rawJSONOrNil(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}

func timeOrNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}
