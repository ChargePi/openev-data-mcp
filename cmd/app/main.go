package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChargePi/openev-data-mcp/internal/app"
	"github.com/ChargePi/openev-data-mcp/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "openev-data-mcp",
	Short: "MCP server serving the open-ev-data dataset",
	RunE:  run,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "path to config file")
}

func run(cmd *cobra.Command, _ []string) error {
	logger, err := buildLogger()
	if err != nil {
		return err
	}
	defer logger.Sync() //nolint:errcheck

	cfg, err := config.Load(configFile)
	if err != nil {
		logger.Fatal("loading config", zap.Error(err))
	}

	a := app.New(logger)
	a.Run(cmd.Context(), cfg)
	a.Shutdown(context.Background())
	return nil
}

// buildLogger writes to stderr so stdout remains clean for the MCP stdio protocol.
func buildLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	return cfg.Build()
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}