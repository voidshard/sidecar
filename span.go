package main

import (
	"context"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Span is a wrapper around an OpenTelemetry span
type Span struct {
	root    trace.Span
	Context context.Context
}

// End ends the span
func (s *Span) End() {
	s.root.End()
}

// SetAttributes sets the attributes on the span
func (s *Span) SetAttributes(attrs ...map[string]interface{}) {
	s.root.SetAttributes(toOtelAttrs(attrs...)...)
}

// NewSpan creates a new span with the given name and attributes from the default tracer
func NewSpan(ctx context.Context, name string, attrs ...map[string]interface{}) *Span {
	// https://pkg.go.dev/go.opentelemetry.io/otel/trace#Tracer
	// Start creates a span and a context.Context containing the newly-created span.
	//
	// If the context.Context provided in `ctx` contains a Span then the newly-created
	// Span will be a child of that span, otherwise it will be a root span. This behavior
	// can be overridden by providing `WithNewRoot()` as a SpanOption, causing the
	// newly-created Span to be a root span even if `ctx` contains a Span.
	//
	// When creating a Span it is recommended to provide all known span attributes using
	// the `WithAttributes()` SpanOption as samplers will only have access to the
	// attributes provided when a Span is created.
	//
	// Any Span that is created MUST also be ended. This is the responsibility of the user.
	// Implementations of this API may leak memory or other resources if Spans are not ended.
	ctx, span := tracer.Start(ctx, name, trace.WithAttributes(toOtelAttrs(attrs...)...))
	return &Span{root: span, Context: ctx}
}

// Err records an error on the span, if the error is not nil
func (s *Span) Err(err error) error {
	if err != nil {
		s.root.RecordError(err)
	}
	return err
}

// toOtelAttrs converts a map of string to interface{} to a slice of OpenTelemetry attributes
func toOtelAttrs(attrs ...map[string]interface{}) []attribute.KeyValue {
	otelAttrs := []attribute.KeyValue{}
	for _, attrset := range attrs {
		for k, v := range attrset {
			var val reflect.Value
			if reflect.TypeOf(v).Kind() == reflect.Ptr {
				// if the value is a pointer, dereference it (otherwise we get the pointer address)
				val = reflect.ValueOf(v).Elem()
			} else {
				val = reflect.ValueOf(v)
			}

			err, ok := v.(error)
			if ok {
				// the only "struct" we will accept, should be set with Err() (above)
				otelAttrs = append(otelAttrs, attribute.String(k, err.Error()))
				continue
			}

			// do our best to convert the value to an OTel attribute
			switch val.Kind() {
			case reflect.String:
				otelAttrs = append(otelAttrs, attribute.String(k, val.String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				otelAttrs = append(otelAttrs, attribute.Int64(k, val.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				otelAttrs = append(otelAttrs, attribute.Int64(k, int64(val.Uint())))
			case reflect.Float32, reflect.Float64:
				otelAttrs = append(otelAttrs, attribute.Float64(k, val.Float()))
			case reflect.Bool:
				otelAttrs = append(otelAttrs, attribute.Bool(k, val.Bool()))
			}
		}
	}
	return otelAttrs
}
