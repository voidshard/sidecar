package main

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var propgator propagation.TextMapPropagator
var tracer trace.Tracer

func init() {
	setupOTelSDK(context.Background())
}

// traceIDGenerator is a custom ID generator so we can set trace IDs depending on the HTTP context.
//
// We use an IP address set in the context to generate the trace ID.
type traceIDGenerator struct{}

// NewIDs generates a new trace and span ID.
func (traceIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	// https://github.com/open-telemetry/opentelemetry-go/blob/main/trace/internal/telemetry/id.go

	data := newID(ctx.Value(ctxIPv4))
	tr := trace.TraceID{}
	copy(tr[:], data[:len(tr)])

	data = newID() // random id
	sp := trace.SpanID{}
	copy(sp[:], data[:len(sp)])

	return tr, sp
}

// NewSpanID generates a new span ID.
func (traceIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	data := newID() // random id
	sp := trace.SpanID{}
	copy(sp[:], data[:len(sp)])
	return sp
}

// discoverServiceName attempts to find the service name from OTel environment variables
// so we can set our tracer service name appropriately.
func discoverServiceName() string {
	svcName := os.Getenv("OTEL_SERVICE_NAME")
	if svcName == "" {
		svcAttrs := os.Getenv("OTEL_RESOURCE_ATTRIBUTES")
		for _, b := range strings.Split(svcAttrs, ",") {
			kv := strings.Split(b, "=")
			if len(kv) != 2 {
				continue
			}
			if strings.ToLower(kv[0]) == "service.name" {
				svcName = kv[1]
			}
		}
	}
	if svcName == "" {
		// opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
		svcName = "unknown"
	}

	return svcName
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (func(), error) {
	// Configure a new OTLP exporter
	client := otlptracegrpc.NewClient()
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	// Create a new tracer provider with a batch span processor and the otlp exporter
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp), sdktrace.WithIDGenerator(traceIDGenerator{}))

	// Register the global Tracer provider
	otel.SetTracerProvider(tp)

	tracer = otel.Tracer(discoverServiceName())
	propgator = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})

	// Register the W3C trace context and baggage propagators so data is propagated across services/processes
	otel.SetTextMapPropagator(propgator)

	return func() {
		_ = exp.Shutdown(ctx)
		_ = tp.Shutdown(ctx)
	}, err
}
