package godiff

import "context"

// LinesContext computes a line-level diff, returning early if ctx is cancelled.
//
//	patch, err := godiff.LinesContext(ctx, oldText, newText)
func LinesContext(ctx context.Context, a, b string) (Patch, error) {
	done := make(chan Patch, 1)
	go func() {
		done <- Lines(a, b)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-done:
		return p, nil
	}
}

// WordsContext computes a word-level diff, returning early if ctx is cancelled.
func WordsContext(ctx context.Context, a, b string) (Patch, error) {
	done := make(chan Patch, 1)
	go func() {
		done <- Words(a, b)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-done:
		return p, nil
	}
}

// CharsContext computes a character-level diff, returning early if ctx is cancelled.
func CharsContext(ctx context.Context, a, b string) (Patch, error) {
	done := make(chan Patch, 1)
	go func() {
		done <- Chars(a, b)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-done:
		return p, nil
	}
}

// JSONContext computes a JSON structural diff, returning early if ctx is cancelled.
func JSONContext(ctx context.Context, a, b []byte) (JSONPatch, error) {
	type result struct {
		p   JSONPatch
		err error
	}
	done := make(chan result, 1)
	go func() {
		p, err := JSON(a, b)
		done <- result{p, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-done:
		return r.p, r.err
	}
}
