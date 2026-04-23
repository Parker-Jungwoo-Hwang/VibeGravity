# Go Code Header: Development Log

Use this only for files that need explicit agent-audit context across repeated
edits. Prefer git history and ADRs for normal changes.

```go
// +--------------------------------------------------------------+
// | FILE    : path/to/file.go                                    |
// | MODULE  : module/package name                                |
// | PURPOSE : One-line summary                                   |
// +--------------------------------------------------------------+
//
// INTERFACE
//   Public  : ExportedName(arg) -> Result
//   Private : helperName()
//   Emits   : EventName
//
// CHANGE LOG
//   [YYYY-MM-DD] @handle  | INIT     | Initial implementation
//   [YYYY-MM-DD] agent    | PATCH    | Short change summary
//
// AGENT LOG
//   TASK       : Last agent task summary
//   CONFIDENCE : high | medium | low
//   NEXT       : Remaining follow-up
```
