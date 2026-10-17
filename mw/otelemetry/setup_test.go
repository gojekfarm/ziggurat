package otelemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
)

// TestSetup_Disabled_NoOp verifies Setup(enabled=false) does not touch the global
// TracerProvider/propagator at all, and hands back a shutdown func that's always safe
// to call - this is the behavior that makes tracing "off by default" work: a caller can
// always call Setup unconditionally and gate it purely by the enabled flag.
func TestSetup_Disabled_NoOp(t *testing.T) {
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})

	shutdown, err := Setup(context.Background(), "ziggurat-test", "localhost:4317", false)
	if err != nil {
		t.Fatalf("Setup(enabled=false) returned an error: %v", err)
	}
	if shutdown == nil {
		t.Fatal("Setup(enabled=false) returned a nil shutdown func")
	}

	if got := otel.GetTracerProvider(); got != prevTP {
		t.Error("Setup(enabled=false) must not replace the global TracerProvider")
	}
	if got := otel.GetTextMapPropagator(); got != prevProp {
		t.Error("Setup(enabled=false) must not replace the global TextMapPropagator")
	}

	if err := shutdown(context.Background()); err != nil {
		t.Errorf("no-op shutdown returned an error: %v", err)
	}
}

// TestSetup_Enabled_RegistersProvider verifies the enabled=true path still does what it
// always did: register a real TracerProvider and composite propagator, without
// requiring a reachable collector (otlptracegrpc.New dials lazily, not at construction
// time, so this doesn't need a live endpoint at localhost:4317).
func TestSetup_Enabled_RegistersProvider(t *testing.T) {
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})

	shutdown, err := Setup(context.Background(), "ziggurat-test", "localhost:4317", true)
	if err != nil {
		t.Fatalf("Setup(enabled=true) returned an error: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	if got := otel.GetTracerProvider(); got == prevTP {
		t.Error("Setup(enabled=true) must register a new global TracerProvider")
	}
	if got := otel.GetTextMapPropagator(); got == prevProp {
		t.Error("Setup(enabled=true) must register a new global TextMapPropagator")
	}
}
