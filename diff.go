// Package godiff provides structured, human-readable diffing for Go.
//
// It supports line-level, word-level, character-level, and JSON structural
// diffs, unified patch format output, and patch application. Designed to be
// deterministic, zero-dependency, and safe for use in AI agent pipelines.
//
// # Quick Start
//
//	result := godiff.Lines("hello\nworld\n", "hello\nGo\n")
//	fmt.Println(godiff.Unified(result, "old.txt", "new.txt", 3))
//
//	patch := godiff.Words("the cat sat", "the dog sat")
//	for _, op := range patch {
//	    fmt.Printf("%s %q\n", op.Type, op.Text)
//	}
package godiff

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// OpType represents the type of a diff operation.
type OpType string

const (
	// OpEqual indicates text that is the same in both inputs.
	OpEqual OpType = "equal"
	// OpInsert indicates text added in the new version.
	OpInsert OpType = "insert"
	// OpDelete indicates text removed from the old version.
	OpDelete OpType = "delete"
)

// Op represents a single diff operation.
type Op struct {
	Type OpType `json:"type"`
	Text string `json:"text"`
}

// String returns a human-readable representation of the Op.
func (o Op) String() string {
	switch o.Type {
	case OpInsert:
		return fmt.Sprintf("+ %s", o.Text)
	case OpDelete:
		return fmt.Sprintf("- %s", o.Text)
	default:
		return fmt.Sprintf("  %s", o.Text)
	}
}

// Patch is a slice of diff operations.
type Patch []Op

// HasChanges returns true if the patch contains any insertions or deletions.
func (p Patch) HasChanges() bool {
	for _, op := range p {
		if op.Type != OpEqual {
			return true
		}
	}
	return false
}

// Insertions returns only the inserted ops.
func (p Patch) Insertions() Patch {
	return p.filter(OpInsert)
}

// Deletions returns only the deleted ops.
func (p Patch) Deletions() Patch {
	return p.filter(OpDelete)
}

func (p Patch) filter(t OpType) Patch {
	var out Patch
	for _, op := range p {
		if op.Type == t {
			out = append(out, op)
		}
	}
	return out
}

// Stats returns counts of equal, inserted, and deleted ops.
func (p Patch) Stats() (equal, inserted, deleted int) {
	for _, op := range p {
		switch op.Type {
		case OpEqual:
			equal++
		case OpInsert:
			inserted++
		case OpDelete:
			deleted++
		}
	}
	return
}

// Apply applies the patch to the original string, returning the result.
// Returns an error if the patch does not match the source.
func (p Patch) Apply(src string) (string, error) {
	var b strings.Builder
	pos := 0
	for _, op := range p {
		switch op.Type {
		case OpEqual:
			if pos+len(op.Text) > len(src) {
				return "", fmt.Errorf("godiff: patch mismatch at position %d", pos)
			}
			if src[pos:pos+len(op.Text)] != op.Text {
				return "", fmt.Errorf("godiff: source mismatch: expected %q got %q", op.Text, src[pos:pos+len(op.Text)])
			}
			b.WriteString(op.Text)
			pos += len(op.Text)
		case OpInsert:
			b.WriteString(op.Text)
		case OpDelete:
			if pos+len(op.Text) > len(src) {
				return "", fmt.Errorf("godiff: patch mismatch at position %d", pos)
			}
			pos += len(op.Text)
		}
	}
	return b.String(), nil
}

// --- Myers diff algorithm (character level) ---

// Chars computes a character-level diff between a and b.
//
//	patch := godiff.Chars("kitten", "sitting")
func Chars(a, b string) Patch {
	return myersDiff(splitChars(a), splitChars(b))
}

// Words computes a word-level diff between a and b.
//
//	patch := godiff.Words("the cat sat", "the dog sat")
func Words(a, b string) Patch {
	return myersDiff(splitWords(a), splitWords(b))
}

// Lines computes a line-level diff between a and b.
// Lines should be newline-terminated or separated.
//
//	patch := godiff.Lines("foo\nbar\n", "foo\nbaz\n")
func Lines(a, b string) Patch {
	return myersDiff(splitLines(a), splitLines(b))
}

// splitChars splits a string into individual character tokens.
func splitChars(s string) []string {
	runes := []rune(s)
	out := make([]string, len(runes))
	for i, r := range runes {
		out[i] = string(r)
	}
	return out
}

// splitWords splits a string into word tokens, preserving whitespace as tokens.
func splitWords(s string) []string {
	var tokens []string
	var cur strings.Builder
	inWord := false
	for _, r := range s {
		isWord := !unicode.IsSpace(r)
		if isWord != inWord {
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
			inWord = isWord
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

// splitLines splits a string into line tokens, each including its trailing newline.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i+1])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// myersDiff implements the Myers diff algorithm over a slice of string tokens.
func myersDiff(a, b []string) Patch {
	n, m := len(a), len(b)
	if n == 0 && m == 0 {
		return Patch{}
	}
	if n == 0 {
		p := make(Patch, m)
		for i, s := range b {
			p[i] = Op{Type: OpInsert, Text: s}
		}
		return p
	}
	if m == 0 {
		p := make(Patch, n)
		for i, s := range a {
			p[i] = Op{Type: OpDelete, Text: s}
		}
		return p
	}

	max := n + m
	v := make([]int, 2*max+1)
	var trace [][]int

	for d := 0; d <= max; d++ {
		snap := make([]int, len(v))
		copy(snap, v)
		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[k-1+max] < v[k+1+max]) {
				x = v[k+1+max]
			} else {
				x = v[k-1+max] + 1
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[k+max] = x
			if x >= n && y >= m {
				trace = append(trace, snap)
				return backtrack(a, b, trace, d)
			}
		}
		trace = append(trace, snap)
	}
	return backtrack(a, b, trace, max)
}

func backtrack(a, b []string, trace [][]int, d int) Patch {
	n, m := len(a), len(b)
	max := n + m
	x, y := n, m
	var ops []Op

	for d := len(trace) - 1; d >= 0; d-- {
		v := trace[d]
		k := x - y
		var prevK int
		if k == -d || (k != d && v[k-1+max] < v[k+1+max]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := v[prevK+max]
		prevY := prevX - prevK

		for x > prevX && y > prevY {
			x--
			y--
			ops = append(ops, Op{Type: OpEqual, Text: a[x]})
		}
		if d > 0 {
			if x == prevX {
				y--
				ops = append(ops, Op{Type: OpInsert, Text: b[y]})
			} else {
				x--
				ops = append(ops, Op{Type: OpDelete, Text: a[x]})
			}
		}
	}

	// Reverse ops
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}

	return mergePatch(ops)
}

// mergePatch merges consecutive ops of the same type.
func mergePatch(ops []Op) Patch {
	if len(ops) == 0 {
		return Patch{}
	}
	var merged Patch
	cur := ops[0]
	for _, op := range ops[1:] {
		if op.Type == cur.Type {
			cur.Text += op.Text
		} else {
			merged = append(merged, cur)
			cur = op
		}
	}
	merged = append(merged, cur)
	return merged
}

// --- Unified diff format ---

// UnifiedOptions controls unified diff output.
type UnifiedOptions struct {
	// Context is the number of unchanged lines shown around changes. Default 3.
	Context int
	// OldHeader is the label for the old file (e.g. "--- a/file.go").
	OldHeader string
	// NewHeader is the label for the new file (e.g. "+++ b/file.go").
	NewHeader string
	// Timestamp includes modification timestamps in headers (like GNU diff).
	Timestamp bool
}

// Unified formats a line-level Patch as a unified diff string.
// oldName and newName are used in the file headers.
// context is the number of unchanged context lines around each hunk (typically 3).
//
//	patch := godiff.Lines(oldText, newText)
//	fmt.Print(godiff.Unified(patch, "old.go", "new.go", 3))
func Unified(patch Patch, oldName, newName string, context int) string {
	if context < 0 {
		context = 3
	}
	lines := patchToLineOps(patch)
	hunks := buildHunks(lines, context)
	if len(hunks) == 0 {
		return ""
	}

	var b strings.Builder
	ts := ""
	now := time.Now().UTC().Format("2006-01-02 15:04:05 +0000")
	if true {
		ts = "\t" + now
	}
	_ = ts
	b.WriteString(fmt.Sprintf("--- %s\n", oldName))
	b.WriteString(fmt.Sprintf("+++ %s\n", newName))

	for _, h := range hunks {
		b.WriteString(h.header())
		for _, l := range h.lines {
			b.WriteString(l)
		}
	}
	return b.String()
}

type lineOp struct {
	t    OpType
	text string
}

func patchToLineOps(patch Patch) []lineOp {
	var ops []lineOp
	for _, op := range patch {
		lines := splitLines(op.Text)
		if len(lines) == 0 && op.Text != "" {
			lines = []string{op.Text}
		}
		for _, l := range lines {
			ops = append(ops, lineOp{t: op.Type, text: l})
		}
	}
	return ops
}

type hunk struct {
	oldStart, oldLen int
	newStart, newLen int
	lines            []string
}

func (h hunk) header() string {
	return fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", h.oldStart, h.oldLen, h.newStart, h.newLen)
}

func buildHunks(ops []lineOp, ctx int) []hunk {
	var hunks []hunk
	n := len(ops)
	i := 0
	oldLine, newLine := 1, 1

	type linePos struct {
		op      lineOp
		oldLine int
		newLine int
	}
	var positioned []linePos
	ol, nl := 1, 1
	for _, op := range ops {
		positioned = append(positioned, linePos{op, ol, nl})
		if op.t != OpInsert {
			ol++
		}
		if op.t != OpDelete {
			nl++
		}
	}
	_ = i
	_ = oldLine
	_ = newLine
	_ = n

	// Find change regions
	changed := make([]bool, len(positioned))
	for i, lp := range positioned {
		if lp.op.t != OpEqual {
			changed[i] = true
		}
	}

	// Build hunks
	in := 0
	for in < len(positioned) {
		if !changed[in] {
			in++
			continue
		}
		// Find hunk range
		start := in - ctx
		if start < 0 {
			start = 0
		}
		end := in + ctx + 1
		// extend
		for end < len(positioned) && (changed[end] || withinCtx(changed, end, ctx)) {
			end++
		}
		if end > len(positioned) {
			end = len(positioned)
		}

		h := hunk{
			oldStart: positioned[start].oldLine,
			newStart: positioned[start].newLine,
		}
		for _, lp := range positioned[start:end] {
			prefix := " "
			if lp.op.t == OpInsert {
				prefix = "+"
				h.newLen++
			} else if lp.op.t == OpDelete {
				prefix = "-"
				h.oldLen++
			} else {
				h.oldLen++
				h.newLen++
			}
			text := lp.op.text
			if !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			h.lines = append(h.lines, prefix+text)
		}
		hunks = append(hunks, h)
		in = end
	}
	return hunks
}

func withinCtx(changed []bool, i, ctx int) bool {
	for j := i - ctx; j <= i+ctx; j++ {
		if j >= 0 && j < len(changed) && changed[j] {
			return true
		}
	}
	return false
}

// --- Ratio / similarity ---

// Ratio returns a similarity ratio between 0.0 and 1.0 for two strings (line-level).
// 1.0 means identical, 0.0 means completely different.
//
//	r := godiff.Ratio("hello world", "hello Go")
func Ratio(a, b string) float64 {
	if a == b {
		return 1.0
	}
	patch := Lines(a, b)
	var matches int
	total := len(splitLines(a)) + len(splitLines(b))
	if total == 0 {
		return 1.0
	}
	for _, op := range patch {
		if op.Type == OpEqual {
			matches += len(splitLines(op.Text))
		}
	}
	return 2.0 * float64(matches) / float64(total)
}

// RatioChars returns a character-level similarity ratio between 0.0 and 1.0.
//
//	r := godiff.RatioChars("kitten", "sitting")
func RatioChars(a, b string) float64 {
	if a == b {
		return 1.0
	}
	total := len([]rune(a)) + len([]rune(b))
	if total == 0 {
		return 1.0
	}
	patch := Chars(a, b)
	var matches int
	for _, op := range patch {
		if op.Type == OpEqual {
			matches += len([]rune(op.Text))
		}
	}
	return 2.0 * float64(matches) / float64(total)
}

// --- LCS (Longest Common Subsequence) ---

// LCS returns the longest common subsequence of lines between a and b.
//
//	common := godiff.LCS("foo\nbar\nbaz\n", "foo\nqux\nbaz\n")
func LCS(a, b string) []string {
	aLines := splitLines(a)
	bLines := splitLines(b)
	return lcs(aLines, bLines)
}

func lcs(a, b []string) []string {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] > dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}
	var result []string
	i, j := n, m
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			result = append(result, a[i-1])
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	for l, r := 0, len(result)-1; l < r; l, r = l+1, r-1 {
		result[l], result[r] = result[r], result[l]
	}
	return result
}

// --- Close matches ---

// ClosestMatch returns the element from candidates most similar to target (line-level ratio).
// Returns empty string if candidates is empty.
//
//	best := godiff.ClosestMatch("helo", []string{"hello", "world", "help"})
func ClosestMatch(target string, candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	best := candidates[0]
	bestScore := RatioChars(target, candidates[0])
	for _, c := range candidates[1:] {
		score := RatioChars(target, c)
		if score > bestScore {
			bestScore = score
			best = c
		}
	}
	return best
}

// ClosestMatches returns up to n candidates sorted by similarity to target (descending).
//
//	matches := godiff.ClosestMatches("helo", []string{"hello", "world", "help"}, 2)
func ClosestMatches(target string, candidates []string, n int) []string {
	type scored struct {
		s     string
		score float64
	}
	sc := make([]scored, len(candidates))
	for i, c := range candidates {
		sc[i] = scored{c, RatioChars(target, c)}
	}
	// Simple insertion sort (n is typically small)
	for i := 1; i < len(sc); i++ {
		for j := i; j > 0 && sc[j].score > sc[j-1].score; j-- {
			sc[j], sc[j-1] = sc[j-1], sc[j]
		}
	}
	if n > len(sc) {
		n = len(sc)
	}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = sc[i].s
	}
	return out
}
