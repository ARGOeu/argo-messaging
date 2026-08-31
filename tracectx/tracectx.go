// Package tracectx carries the per-request trace ID through context.Context.
//
// The value is populated by the HTTP handler wrappers in the handlers package
// and consumed by downstream packages (stores, auth, brokers, subscriptions,
// topics, metrics) when they emit structured logs.
//
// A dedicated unexported key type is used so the context value cannot
// collide with any other package's context keys (see staticcheck SA1029).
package tracectx

import "context"

type traceIDKey struct{}

// TraceIDKey is the context.Context key under which the request trace ID is
// stored. Producers should call context.WithValue(ctx, tracectx.TraceIDKey,
// traceID); consumers should call tracectx.FromContext(ctx).
var TraceIDKey = traceIDKey{}

// WithTraceID returns a copy of ctx with the given trace ID attached.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// FromContext returns the trace ID previously attached via WithTraceID or
// the equivalent context.WithValue call. It returns the empty string if no
// trace ID is present.
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(TraceIDKey).(string); ok {
		return v
	}
	return ""
}
