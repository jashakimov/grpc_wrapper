package meta

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

const (
	traceID = "X-Trace-Id"
	spanID  = "X-Span-Id"
)

func BuildSpanContext(ctx context.Context) context.Context {
	rawTraceID, rawSpanID, err := ExtractMeta(ctx)
	if err != nil {
		return ctx
	}

	traceID, err := trace.TraceIDFromHex(rawTraceID)
	if err != nil {
		return ctx
	}

	spanID, err := trace.SpanIDFromHex(rawSpanID)
	if err != nil {
		return ctx
	}

	parentSpanCtx := trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID:    traceID,
			SpanID:     spanID,
			Remote:     false,
			TraceFlags: 1,
		},
	)

	return trace.ContextWithRemoteSpanContext(ctx, parentSpanCtx)
}

func ExtractMeta(ctx context.Context) (string, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", errors.New("context info is empty")
	}

	tr := md.Get(traceID)
	if len(tr) == 0 {
		return "", "", errors.New("empty traceId")
	}
	sp := md.Get(spanID)
	if len(sp) == 0 {
		return "", "", errors.New("empty spanId")
	}
	return tr[0], sp[0], nil
}

func NewGrpcContext(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, traceID, span.SpanContext().TraceID().String())
	return metadata.AppendToOutgoingContext(ctx, spanID, span.SpanContext().SpanID().String())
}
