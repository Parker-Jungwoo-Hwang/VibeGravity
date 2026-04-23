# Go Code Header: Contextual Narrative

Use this for core domain, runtime contract, graph/apply, or reasoning modules
where the "why" matters more than a compact export list.

```go
// ================================================================
// FILE    : path/to/file.go
// CREATED : YYYY-MM-DD | AUTHOR: @github_handle
// ----------------------------------------------------------------
//
// [OVERVIEW]
// Two or three sentences describing the responsibility boundary.
// Focus on why this file is separated from adjacent modules.
//
// [ARCHITECTURE ROLE]
// Describe where this module sits in the VibeGravity runtime.
//
// [KEY DECISIONS]
// - Decision: reason.
//
// [KNOWN LIMITATIONS]
// - Current constraint or debt.
//
// [AGENT CONTEXT]
// The rule an agent must preserve while editing this file.
//
// ================================================================
```
