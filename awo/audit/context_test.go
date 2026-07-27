package audit

import (
	"context"
	"testing"
)

func TestWriterFromContext_Missing(t *testing.T) {
	t.Parallel()

	w, ok := WriterFromContext(context.Background())
	if ok {
		t.Error("WriterFromContext on empty context: expected ok=false")
	}
	// Returned writer must be a no-op (not nil).
	if w == nil {
		t.Error("WriterFromContext: returned writer must not be nil")
	}
}

func TestWithWriter_RoundTrip(t *testing.T) {
	t.Parallel()

	rw := &RecordingWriter{}
	ctx := WithWriter(context.Background(), rw)

	got, ok := WriterFromContext(ctx)
	if !ok {
		t.Error("WriterFromContext after WithWriter: expected ok=true")
	}
	if got != rw {
		t.Error("WriterFromContext: returned wrong writer")
	}
}
