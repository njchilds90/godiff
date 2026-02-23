# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-23

### Added
- Character-level diff via Myers algorithm (`Chars`)
- Word-level diff with whitespace preservation (`Words`)
- Line-level diff (`Lines`)
- Unified diff format output (`Unified`)
- Patch application (`Patch.Apply`)
- JSON structural diff (`JSON`, `JSONStrings`)
- Similarity ratio (`Ratio`, `RatioChars`)
- Longest Common Subsequence (`LCS`)
- Closest match helpers (`ClosestMatch`, `ClosestMatches`)
- `context.Context` variants for all major functions
- `Patch.Stats()`, `Patch.HasChanges()`, `Patch.Insertions()`, `Patch.Deletions()`
- `JSONPatch.FilterByType()`, `JSONPatch.HasChanges()`
- Full table-driven test suite
- GitHub Actions CI (Go 1.21, 1.22, 1.23)
- Zero external dependencies
