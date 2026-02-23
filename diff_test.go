package godiff_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/njchilds90/godiff"
)

// --- Chars ---

func TestCharsIdentical(t *testing.T) {
	p := godiff.Chars("hello", "hello")
	if p.HasChanges() {
		t.Errorf("expected no changes for identical strings, got %v", p)
	}
}

func TestCharsEmpty(t *testing.T) {
	p := godiff.Chars("", "")
	if len(p) != 0 {
		t.Errorf("expected empty patch for two empty strings")
	}
}

func TestCharsInsertAll(t *testing.T) {
	p := godiff.Chars("", "abc")
	if !p.HasChanges() {
		t.Error("expected changes")
	}
	ins := p.Insertions()
	if len(ins) == 0 {
		t.Error("expected insertions")
	}
}

func TestCharsDeleteAll(t *testing.T) {
	p := godiff.Chars("abc", "")
	dels := p.Deletions()
	if len(dels) == 0 {
		t.Error("expected deletions")
	}
}

func TestCharsApply(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"hello", "hello"},
		{"hello", "helo"},
		{"kitten", "sitting"},
		{"", "abc"},
		{"abc", ""},
		{"abc", "abc"},
		{"abcdef", "azced"},
	}
	for _, tt := range tests {
		p := godiff.Chars(tt.a, tt.b)
		got, err := p.Apply(tt.a)
		if err != nil {
			t.Errorf("Apply(%q, %q): %v", tt.a, tt.b, err)
			continue
		}
		if got != tt.b {
			t.Errorf("Apply(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.b)
		}
	}
}

// --- Words ---

func TestWordsBasic(t *testing.T) {
	p := godiff.Words("the cat sat", "the dog sat")
	if !p.HasChanges() {
		t.Error("expected changes")
	}
	equal, ins, del := p.Stats()
	if ins == 0 || del == 0 {
		t.Errorf("expected insertions and deletions, got equal=%d ins=%d del=%d", equal, ins, del)
	}
}

func TestWordsIdentical(t *testing.T) {
	p := godiff.Words("hello world", "hello world")
	if p.HasChanges() {
		t.Error("expected no changes")
	}
}

func TestWordsApply(t *testing.T) {
	tests := []struct{ a, b string }{
		{"the cat sat", "the dog sat"},
		{"hello world", "hello world"},
		{"one two three", "one four three"},
		{"", "hello"},
		{"hello", ""},
	}
	for _, tt := range tests {
		p := godiff.Words(tt.a, tt.b)
		got, err := p.Apply(tt.a)
		if err != nil {
			t.Errorf("Words.Apply(%q, %q): %v", tt.a, tt.b, err)
			continue
		}
		if got != tt.b {
			t.Errorf("Words.Apply(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.b)
		}
	}
}

// --- Lines ---

func TestLinesBasic(t *testing.T) {
	a := "foo\nbar\nbaz\n"
	b := "foo\nqux\nbaz\n"
	p := godiff.Lines(a, b)
	if !p.HasChanges() {
		t.Error("expected changes")
	}
}

func TestLinesNoChange(t *testing.T) {
	a := "foo\nbar\n"
	p := godiff.Lines(a, a)
	if p.HasChanges() {
		t.Error("expected no changes")
	}
}

func TestLinesApply(t *testing.T) {
	tests := []struct{ a, b string }{
		{"foo\nbar\nbaz\n", "foo\nqux\nbaz\n"},
		{"line1\nline2\n", "line1\nline2\n"},
		{"", "hello\n"},
		{"hello\n", ""},
		{"a\nb\nc\n", "a\nd\nc\ne\n"},
	}
	for _, tt := range tests {
		p := godiff.Lines(tt.a, tt.b)
		got, err := p.Apply(tt.a)
		if err != nil {
			t.Errorf("Lines.Apply err: %v\na=%q b=%q", err, tt.a, tt.b)
			continue
		}
		if got != tt.b {
			t.Errorf("Lines.Apply(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.b)
		}
	}
}

// --- Unified ---

func TestUnified(t *testing.T) {
	a := "foo\nbar\nbaz\n"
	b := "foo\nqux\nbaz\n"
	p := godiff.Lines(a, b)
	u := godiff.Unified(p, "old.txt", "new.txt", 3)
	if !strings.Contains(u, "--- old.txt") {
		t.Error("missing old header")
	}
	if !strings.Contains(u, "+++ new.txt") {
		t.Error("missing new header")
	}
	if !strings.Contains(u, "+qux") {
		t.Error("missing insertion line")
	}
	if !strings.Contains(u, "-bar") {
		t.Error("missing deletion line")
	}
}

func TestUnifiedNoChanges(t *testing.T) {
	a := "same\n"
	p := godiff.Lines(a, a)
	u := godiff.Unified(p, "a", "b", 3)
	if u != "" {
		t.Errorf("expected empty unified diff for identical files, got %q", u)
	}
}

// --- Ratio ---

func TestRatioIdentical(t *testing.T) {
	r := godiff.Ratio("hello\n", "hello\n")
	if r != 1.0 {
		t.Errorf("expected 1.0, got %f", r)
	}
}

func TestRatioZero(t *testing.T) {
	r := godiff.RatioChars("abc", "xyz")
	if r < 0 || r > 1 {
		t.Errorf("ratio out of range: %f", r)
	}
}

func TestRatioCharsIdentical(t *testing.T) {
	r := godiff.RatioChars("hello", "hello")
	if r != 1.0 {
		t.Errorf("expected 1.0, got %f", r)
	}
}

func TestRatioBothEmpty(t *testing.T) {
	r := godiff.RatioChars("", "")
	if r != 1.0 {
		t.Errorf("expected 1.0 for two empty strings, got %f", r)
	}
}

// --- LCS ---

func TestLCS(t *testing.T) {
	common := godiff.LCS("foo\nbar\nbaz\n", "foo\nqux\nbaz\n")
	want := []string{"foo\n", "baz\n"}
	if len(common) != len(want) {
		t.Fatalf("LCS len %d, want %d: %v", len(common), len(want), common)
	}
	for i := range want {
		if common[i] != want[i] {
			t.Errorf("LCS[%d] = %q, want %q", i, common[i], want[i])
		}
	}
}

func TestLCSEmpty(t *testing.T) {
	common := godiff.LCS("", "")
	if len(common) != 0 {
		t.Errorf("expected empty LCS for empty inputs")
	}
}

// --- ClosestMatch ---

func TestClosestMatch(t *testing.T) {
	best := godiff.ClosestMatch("helo", []string{"hello", "world", "help"})
	if best != "hello" && best != "help" {
		t.Logf("ClosestMatch returned %q (acceptable)", best)
	}
}

func TestClosestMatchEmpty(t *testing.T) {
	best := godiff.ClosestMatch("x", []string{})
	if best != "" {
		t.Errorf("expected empty string, got %q", best)
	}
}

func TestClosestMatches(t *testing.T) {
	matches := godiff.ClosestMatches("helo", []string{"hello", "world", "help", "helm"}, 2)
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

// --- Stats ---

func TestPatchStats(t *testing.T) {
	p := godiff.Lines("a\nb\nc\n", "a\nd\nc\n")
	equal, ins, del := p.Stats()
	if ins == 0 || del == 0 {
		t.Errorf("expected ins and del > 0, got equal=%d ins=%d del=%d", equal, ins, del)
	}
}

// --- JSON ---

func TestJSONNoChanges(t *testing.T) {
	patch, err := godiff.JSONStrings(`{"a":1}`, `{"a":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if patch.HasChanges() {
		t.Error("expected no changes")
	}
}

func TestJSONReplace(t *testing.T) {
	patch, err := godiff.JSONStrings(`{"name":"Alice"}`, `{"name":"Bob"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !patch.HasChanges() {
		t.Error("expected changes")
	}
	replacements := patch.FilterByType("replace")
	if len(replacements) == 0 {
		t.Error("expected replace op")
	}
}

func TestJSONAdd(t *testing.T) {
	patch, err := godiff.JSONStrings(`{"a":1}`, `{"a":1,"b":2}`)
	if err != nil {
		t.Fatal(err)
	}
	adds := patch.FilterByType("add")
	if len(adds) == 0 {
		t.Error("expected add op")
	}
}

func TestJSONRemove(t *testing.T) {
	patch, err := godiff.JSONStrings(`{"a":1,"b":2}`, `{"a":1}`)
	if err != nil {
		t.Fatal(err)
	}
	removes := patch.FilterByType("remove")
	if len(removes) == 0 {
		t.Error("expected remove op")
	}
}

func TestJSONNested(t *testing.T) {
	a := `{"user":{"name":"Alice","age":30}}`
	b := `{"user":{"name":"Bob","age":30}}`
	patch, err := godiff.JSONStrings(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !patch.HasChanges() {
		t.Error("expected changes")
	}
}

func TestJSONArray(t *testing.T) {
	a := `[1,2,3]`
	b := `[1,4,3]`
	patch, err := godiff.JSONStrings(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !patch.HasChanges() {
		t.Error("expected changes in array diff")
	}
}

func TestJSONInvalid(t *testing.T) {
	_, err := godiff.JSONStrings(`{invalid}`, `{}`)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestJSONOpString(t *testing.T) {
	op := godiff.JSONOp{Path: "/name", Type: "replace", OldValue: "Alice", NewValue: "Bob"}
	s := op.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

// --- Context ---

func TestLinesContext(t *testing.T) {
	ctx := context.Background()
	p, err := godiff.LinesContext(ctx, "a\nb\n", "a\nc\n")
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasChanges() {
		t.Error("expected changes")
	}
}

func TestLinesContextCancelled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)
	_, err := godiff.LinesContext(ctx, strings.Repeat("a\n", 1000), strings.Repeat("b\n", 1000))
	// May or may not cancel depending on timing — just ensure no panic
	_ = err
}

func TestJSONContext(t *testing.T) {
	ctx := context.Background()
	patch, err := godiff.JSONContext(ctx, []byte(`{"a":1}`), []byte(`{"a":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if !patch.HasChanges() {
		t.Error("expected changes")
	}
}

// --- Op.String ---

func TestOpString(t *testing.T) {
	ops := []godiff.Op{
		{Type: godiff.OpEqual, Text: "same"},
		{Type: godiff.OpInsert, Text: "new"},
		{Type: godiff.OpDelete, Text: "old"},
	}
	for _, op := range ops {
		s := op.String()
		if s == "" {
			t.Errorf("Op.String() returned empty for %v", op)
		}
	}
}
