package mcptracing

import (
	"context"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("mcp.server")

// AddHooks registers OTel tracing hooks onto h following the MCP semantic
// conventions (https://opentelemetry.io/docs/specs/semconv/gen-ai/mcp/).
//
// One span is created per MCP request. The span is opened in OnBeforeAny,
// enriched with method-specific attributes by the typed before-hooks, and
// closed in OnSuccess or OnError. Spans are correlated across the three hook
// sites via the request id stored in a sync.Map.
func AddHooks(h *mcpserver.Hooks) {
	var spans sync.Map

	h.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, _ any) {
		_, span := tracer.Start(ctx, string(method), trace.WithSpanKind(trace.SpanKindServer))
		span.SetAttributes(attribute.String("mcp.method.name", string(method)))
		spans.Store(id, span)
	})

	// Enrich resource-read spans with the resource URI and operation name.
	h.AddBeforeReadResource(func(_ context.Context, id any, message *mcp.ReadResourceRequest) {
		if v, ok := spans.Load(id); ok {
			v.(trace.Span).SetAttributes(
				attribute.String("mcp.resource.uri", message.Params.URI),
				attribute.String("gen_ai.operation.name", "read_resource"),
			)
		}
	})

	// Enrich tool-call spans with the tool name and update the span name to
	// follow the "{method} {target}" convention from the spec.
	h.AddBeforeCallTool(func(_ context.Context, id any, message *mcp.CallToolRequest) {
		if v, ok := spans.Load(id); ok {
			span := v.(trace.Span)
			span.SetName("tools/call " + message.Params.Name)
			span.SetAttributes(
				attribute.String("gen_ai.tool.name", message.Params.Name),
				attribute.String("gen_ai.operation.name", "execute_tool"),
			)
		}
	})

	h.AddOnSuccess(func(_ context.Context, id any, _ mcp.MCPMethod, _ any, _ any) {
		if v, ok := spans.LoadAndDelete(id); ok {
			v.(trace.Span).End()
		}
	})

	h.AddOnError(func(_ context.Context, id any, _ mcp.MCPMethod, _ any, err error) {
		if v, ok := spans.LoadAndDelete(id); ok {
			span := v.(trace.Span)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
			span.End()
		}
	})
}