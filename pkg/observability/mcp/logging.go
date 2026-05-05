package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// LoggingHooks returns MCP server hooks that log every request and error
// through the provided logger.
func LoggingHooks(logger *zap.Logger, hooks *mcpserver.Hooks) {
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		logger.Info("mcp request", zap.String("method", string(method)))
	})
	hooks.AddAfterComplete(func(ctx context.Context, id any, message *mcp.CompleteRequest, result *mcp.CompleteResult) {
		logger.Debug("mcp response sent", zap.String("method", string(message.Method)))
	})
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		logger.Error("mcp request error",
			zap.String("method", string(method)),
			zap.Error(err),
		)
	})
}
