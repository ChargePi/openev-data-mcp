package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// loggingHooks returns MCP server hooks that log every request and error
// through the provided logger.
func loggingHooks(logger *zap.Logger) *mcpserver.Hooks {
	return &mcpserver.Hooks{
		OnBeforeAny: []mcpserver.BeforeAnyHookFunc{
			func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
				logger.Info("mcp request", zap.String("method", string(method)))
			},
		},
		OnAfterComplete: []mcpserver.OnAfterCompleteFunc{
			func(ctx context.Context, id any, message *mcp.CompleteRequest, result *mcp.CompleteResult) {
				logger.Debug("mcp response sent", zap.String("method", string(message.Method)))
			},
		},
		OnError: []mcpserver.OnErrorHookFunc{
			func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
				logger.Error("mcp request error",
					zap.String("method", string(method)),
					zap.Error(err),
				)
			},
		},
	}
}
