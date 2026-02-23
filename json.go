package godiff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// JSONOp represents a single JSON structural diff operation.
type JSONOp struct {
	// Path is the JSON pointer path to the changed value (e.g. "/user/name").
	Path string `json:"path"`
	// Type is "add", "remove", or "replace".
	Type string `json:"type"`
	// OldValue is the previous value (nil for add).
	OldValue interface{} `json:"old_value,omitempty"`
	// NewValue is the new value (nil for remove).
	NewValue interface{} `json:"new_value,omitempty"`
}

// String returns a human-readable description of the JSONOp.
func (op JSONOp) String() string {
	switch op.Type {
	case "add":
		return fmt.Sprintf("+ %s: %v", op.Path, op.NewValue)
	case "remove":
		return fmt.Sprintf("- %s: %v", op.Path, op.OldValue)
	case "replace":
		return fmt.Sprintf("~ %s: %v → %v", op.Path, op.OldValue, op.NewValue)
	default:
		return fmt.Sprintf("? %s", op.Path)
	}
}

// JSONPatch is a list of JSON diff operations.
type JSONPatch []JSONOp

// HasChanges returns true if there are any differences.
func (p JSONPatch) HasChanges() bool {
	return len(p) > 0
}

// FilterByType returns ops of the given type ("add", "remove", "replace").
func (p JSONPatch) FilterByType(t string) JSONPatch {
	var out JSONPatch
	for _, op := range p {
		if op.Type == t {
			out = append(out, op)
		}
	}
	return out
}

// JSON computes a structural diff between two JSON byte slices.
// Returns a JSONPatch describing adds, removes, and replacements.
//
//	a := []byte(`{"name":"Alice","age":30}`)
//	b := []byte(`{"name":"Bob","age":30,"city":"NY"}`)
//	patch, err := godiff.JSON(a, b)
func JSON(a, b []byte) (JSONPatch, error) {
	var av, bv interface{}
	if err := json.Unmarshal(a, &av); err != nil {
		return nil, fmt.Errorf("godiff: invalid JSON (a): %w", err)
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		return nil, fmt.Errorf("godiff: invalid JSON (b): %w", err)
	}
	var ops JSONPatch
	diffValue("", av, bv, &ops)
	sort.Slice(ops, func(i, j int) bool {
		return ops[i].Path < ops[j].Path
	})
	return ops, nil
}

// JSONStrings computes a structural diff between two JSON strings.
//
//	patch, err := godiff.JSONStrings(`{"x":1}`, `{"x":2}`)
func JSONStrings(a, b string) (JSONPatch, error) {
	return JSON([]byte(a), []byte(b))
}

func diffValue(path string, a, b interface{}, ops *JSONPatch) {
	switch av := a.(type) {
	case map[string]interface{}:
		bm, ok := b.(map[string]interface{})
		if !ok {
			*ops = append(*ops, JSONOp{Path: path, Type: "replace", OldValue: a, NewValue: b})
			return
		}
		diffMap(path, av, bm, ops)
	case []interface{}:
		bs, ok := b.([]interface{})
		if !ok {
			*ops = append(*ops, JSONOp{Path: path, Type: "replace", OldValue: a, NewValue: b})
			return
		}
		diffSlice(path, av, bs, ops)
	default:
		if !jsonEqual(a, b) {
			*ops = append(*ops, JSONOp{Path: path, Type: "replace", OldValue: a, NewValue: b})
		}
	}
}

func diffMap(path string, a, b map[string]interface{}, ops *JSONPatch) {
	for k, av := range a {
		childPath := jsonPath(path, k)
		if bv, exists := b[k]; exists {
			diffValue(childPath, av, bv, ops)
		} else {
			*ops = append(*ops, JSONOp{Path: childPath, Type: "remove", OldValue: av})
		}
	}
	for k, bv := range b {
		if _, exists := a[k]; !exists {
			*ops = append(*ops, JSONOp{Path: jsonPath(path, k), Type: "add", NewValue: bv})
		}
	}
}

func diffSlice(path string, a, b []interface{}, ops *JSONPatch) {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	for i := 0; i < maxLen; i++ {
		childPath := fmt.Sprintf("%s/%d", path, i)
		if i >= len(a) {
			*ops = append(*ops, JSONOp{Path: childPath, Type: "add", NewValue: b[i]})
		} else if i >= len(b) {
			*ops = append(*ops, JSONOp{Path: childPath, Type: "remove", OldValue: a[i]})
		} else {
			diffValue(childPath, a[i], b[i], ops)
		}
	}
}

func jsonPath(base, key string) string {
	key = strings.ReplaceAll(key, "~", "~0")
	key = strings.ReplaceAll(key, "/", "~1")
	return base + "/" + key
}

func jsonEqual(a, b interface{}) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
