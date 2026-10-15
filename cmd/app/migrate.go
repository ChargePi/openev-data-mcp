package main

import (
	"context"

	"github.com/ChargePi/openev-data-mcp/internal/app"
	"github.com/ChargePi/openev-data-mcp/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Provision the vector collection and sync it with the vehicles in the database",
	// Errors are logged; a usage dump would only add noise in job logs.
	SilenceUsage: true,
	RunE:         migrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func migrate(cmd *cobra.Command, _ []string) error {
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
	defer a.Shutdown(context.Background())

	if err := a.Migrate(cmd.Context(), cfg); err != nil {
		logger.Error("migration failed", zap.Error(err))
		return err
	}
	return nil
}
