package otelpyroscope

import (
	"context"
	"runtime/pprof"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func Test_tracerProvider(t *testing.T) {
	otel.SetTracerProvider(NewTracerProvider(trace.NewTracerProvider()))

	tracer := otel.Tracer("")
	labels := make(map[string]string)

	ctx, spanR := tracer.Start(context.Background(), "RootSpan")
	pprof.ForLabels(ctx, func(key, value string) bool {
		labels[key] = value
		return true
	})
	spanID, ok := labels[spanIDLabelName]
	if !ok {
		t.Fatal("span ID label not found")
	}
	if len(spanID) != 16 {
		t.Fatalf("invalid span ID: %q", spanID)
	}
	name, ok := labels[spanNameLabelName]
	if !ok {
		t.Fatal("span name label not found")
	}
	if name != "RootSpan" {
		t.Fatalf("invalid span name: %q", name)
	}

	// Nested child span has the same labels.
	ctx, spanA := tracer.Start(ctx, "SpanA")
	pprof.ForLabels(ctx, func(key, value string) bool {
		if v, ok := labels[key]; !ok || v != value {
			t.Fatalf("nested span labels mismatch: %q=%q", key, value)
		}
		return true
	})

	spanA.End()
	spanR.End()

	// Child span created after the root span end using its context.
	ctx, spanB := tracer.Start(ctx, "SpanB")
	pprof.ForLabels(ctx, func(key, value string) bool {
		if v, ok := labels[key]; !ok || v != value {
			t.Fatalf("nested span labels mismatch: %q=%q", key, value)
		}
		return true
	})
	spanB.End()

	// A new root span.
	ctx, spanC := tracer.Start(context.Background(), "SpanC")
	pprof.ForLabels(ctx, func(key, value string) bool {
		if v, ok := labels[key]; !ok || v == value {
			t.Fatalf("unexpected match: %q=%q", key, value)
		}
		return true
	})
	spanC.End()
}
