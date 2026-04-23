// ============================================================
// FILE     : tools/headercheck/main.go
// PURPOSE  : Validates machine-readable headers on Go source files.
// LAYER    : util
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : docs/code-header-policy.md
// USED_BY  : Makefile, agent handoff checks
// ------------------------------------------------------------
// AGENT_NOTE: Keep this dependency-free so header checks run in a fresh clone.
// ============================================================

// Package main implements the VibeGravity source header checker.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	requiredFields = []string{
		"FILE",
		"PURPOSE",
		"LAYER",
		"STATUS",
		"EXPORTS",
		"DEPENDS",
		"USED_BY",
		"AGENT_NOTE",
	}
	validLayers = map[string]bool{
		"domain":      true,
		"application": true,
		"interface":   true,
		"infra":       true,
		"util":        true,
		"test":        true,
	}
	validStatuses = map[string]bool{
		"draft":        true,
		"active":       true,
		"experimental": true,
		"deprecated":   true,
	}
	fieldPattern = regexp.MustCompile(`^//\s*([A-Z_]+)\s*:\s*(.*)$`)
)

func main() {
	root := flag.String("root", ".", "repository root to scan")
	flag.Parse()

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		os.Exit(2)
	}

	var failures []string
	err = filepath.WalkDir(absRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, err))
			return nil
		}
		if entry.IsDir() {
			if shouldSkipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, err))
			return nil
		}
		rel = filepath.ToSlash(rel)
		failures = append(failures, validateFile(path, rel)...)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan files: %v\n", err)
		os.Exit(2)
	}

	sort.Strings(failures)
	if len(failures) > 0 {
		fmt.Println("code header check failed:")
		for _, failure := range failures {
			fmt.Printf("  - %s\n", failure)
		}
		os.Exit(1)
	}

	fmt.Println("code header check passed")
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "bin", "dist", "node_modules", "vendor":
		return true
	default:
		return false
	}
}

func validateFile(path, rel string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: read: %v", rel, err)}
	}
	text := string(data)
	if strings.HasPrefix(text, "// Code generated ") {
		return nil
	}

	packageOffset := packageOffset(text)
	if packageOffset < 0 {
		return []string{fmt.Sprintf("%s: missing package declaration", rel)}
	}

	fields := parseHeaderFields(text[:packageOffset])
	var failures []string
	for _, field := range requiredFields {
		value, ok := fields[field]
		if !ok {
			failures = append(failures, fmt.Sprintf("%s: missing %s header field", rel, field))
			continue
		}
		if strings.TrimSpace(value) == "" {
			failures = append(failures, fmt.Sprintf("%s: empty %s header field", rel, field))
		}
	}

	if got := fields["FILE"]; got != "" && got != rel {
		failures = append(failures, fmt.Sprintf("%s: FILE header is %q", rel, got))
	}
	if got := fields["LAYER"]; got != "" && !validLayers[got] {
		failures = append(failures, fmt.Sprintf("%s: invalid LAYER %q", rel, got))
	}
	if got := fields["STATUS"]; got != "" && !validStatuses[got] {
		failures = append(failures, fmt.Sprintf("%s: invalid STATUS %q", rel, got))
	}

	return failures
}

func parseHeaderFields(prefix string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(prefix, "\n") {
		match := fieldPattern.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		fields[match[1]] = strings.TrimSpace(match[2])
	}
	return fields
}

func packageOffset(text string) int {
	offset := 0
	for _, line := range strings.SplitAfter(text, "\n") {
		if strings.HasPrefix(line, "package ") {
			return offset
		}
		offset += len(line)
	}
	return -1
}
