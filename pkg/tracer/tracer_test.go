package tracer

import (
	"context"
	"testing"
)

func TestInitTracerLocalNoop(t *testing.T) {
	ctx := context.Background()
	tp, shutdown, err := InitTracer(ctx, Config{
		ServiceName:   "test-xlink",
		Endpoint:      "",
		SamplingRatio: 1.0,
	})
	if err != nil {
		t.Fatalf("expected nil error initializing local tracer, got %v", err)
	}
	if tp == nil {
		t.Fatal("expected non-nil TracerProvider")
	}

	if err := shutdown(ctx); err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}
