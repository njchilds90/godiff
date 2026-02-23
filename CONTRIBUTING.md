# Contributing to godiff

Thank you for your interest in contributing!

## Guidelines

- Keep the zero-dependency policy. No external packages.
- All public functions must have GoDoc comments with examples.
- All new features must include table-driven tests.
- Run `go vet ./...` before submitting.
- Follow standard Go formatting (`gofmt`).

## Submitting Changes

1. Fork the repository
2. Create a branch: `git checkout -b feature/your-feature`
3. Commit your changes with clear messages
4. Open a Pull Request against `main`

## Reporting Issues

Open a GitHub Issue with a minimal reproducible example.
```

---

## Release Instructions (GitHub UI Only)

Once all files are committed:

**Step 1 — Create a tag and release:**
1. In your repo, click **"Releases"** (right sidebar)
2. Click **"Create a new release"**
3. Click **"Choose a tag"** → type `v1.0.0` → click **"Create new tag: v1.0.0 on publish"**
4. Title: `v1.0.0 — Initial Release`
5. In the description box, paste:
```
## godiff v1.0.0

First stable release. Full feature set ships in v1.0.0:

- Myers diff algorithm (chars, words, lines)
- Unified patch format output
- Patch.Apply
- JSON structural diff
- Similarity ratio
- LCS and closest match
- context.Context support throughout
- Zero dependencies
```
6. Click **"Publish release"**

---

## pkg.go.dev Indexing

After publishing the release, pkg.go.dev will auto-index within ~10 minutes.

**To force immediate indexing**, visit this URL in your browser:
```
https://pkg.go.dev/github.com/njchilds90/godiff@v1.0.0
```

Then verify at:
```
https://pkg.go.dev/github.com/njchilds90/godiff
