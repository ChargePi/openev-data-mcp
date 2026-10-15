package app

import (
	"context"
	"fmt"

	"github.com/ChargePi/openev-data-mcp/internal/config"
)

// Migrate provisions the vector collection and syncs it with the vehicles in
// the database, then returns.
func (a *App) Migrate(ctx context.Context, cfg *config.Config) error {
	a.logger.Info("starting vector index migration")

	a.setupObservability(ctx, cfg)
	repo := a.connectDatabase(cfg)
	searchSvc := a.newSearchService(ctx, cfg, repo)

	if err := searchSvc.SyncIndex(ctx); err != nil {
		return fmt.Errorf("syncing vector index: %w", err)
	}

	a.logger.Info("vector index migration finished")
	return nil
}
