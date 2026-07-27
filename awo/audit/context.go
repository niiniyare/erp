package audit

import "context"

// writerKey is the unexported context key for the AuditWriter.
type writerKey struct{}

// WithWriter embeds an AuditWriter into ctx. Called by framework middleware
// or main.go wiring so that service and hook code can retrieve the writer
// without importing the concrete implementation.
func WithWriter(ctx context.Context, w AuditWriter) context.Context {
	return context.WithValue(ctx, writerKey{}, w)
}

// WriterFromContext returns the AuditWriter embedded in ctx, or
// (NoopAuditWriter{}, false) if none is present.
//
// Callers that require an AuditWriter should treat false as a programming
// error — the middleware stack must always inject a writer before reaching
// handler or service code.
func WriterFromContext(ctx context.Context) (AuditWriter, bool) {
	w, ok := ctx.Value(writerKey{}).(AuditWriter)
	if !ok || w == nil {
		return NoopAuditWriter{}, false
	}
	return w, true
}
