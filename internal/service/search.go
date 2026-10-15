package service

import (
	"context"
	"fmt"

	"github.com/ChargePi/openev-data-mcp/internal/embedding"
	"github.com/ChargePi/openev-data-mcp/internal/vectorstore"
	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// SearchRepository is the data-access interface the search service depends on.
type SearchRepository interface {
	ListVehicles(ctx context.Context) ([]vehicle.Vehicle, error)
	GetVehiclesByUniqueCodes(ctx context.Context, codes []string) ([]vehicle.Vehicle, error)
}

// VectorStore is the vector database interface the search service depends on.
type VectorStore interface {
	EnsureCollection(ctx context.Context, size uint64) (bool, error)
	ListIndexed(ctx context.Context) (map[string]vectorstore.IndexedPoint, error)
	Upsert(ctx context.Context, points []vectorstore.Point) error
	Delete(ctx context.Context, pointIDs []string) error
	Search(ctx context.Context, vector []float32, limit int, minScore float32) ([]vectorstore.Hit, error)
}

type SearchConfig struct {
	Dimensions  int
	BatchSize   int
	SearchLimit int
	MinScore    float32
}

type SearchService struct {
	repo     SearchRepository
	embedder embedding.Embedder
	store    VectorStore
	cfg      SearchConfig
	logger   *zap.Logger
}

func NewSearchService(repo SearchRepository, embedder embedding.Embedder, store VectorStore, cfg SearchConfig, logger *zap.Logger) *SearchService {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	if cfg.SearchLimit <= 0 {
		cfg.SearchLimit = 10
	}
	return &SearchService{repo: repo, embedder: embedder, store: store, cfg: cfg, logger: logger}
}

type pendingPoint struct {
	vehicle vehicle.Vehicle
	doc     string
	hash    string
}

// SyncIndex provisions the vector collection and brings it in line with the
// vehicles in the database: new or changed vehicles are embedded and upserted,
// vehicles that no longer exist are removed, and unchanged ones are skipped.
func (s *SearchService) SyncIndex(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "SyncIndex")
	defer span.End()

	err := s.syncIndex(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.logger.Error("failed to sync vector index", zap.Error(err))
	}
	return err
}

func (s *SearchService) syncIndex(ctx context.Context) error {
	created, err := s.store.EnsureCollection(ctx, uint64(s.cfg.Dimensions))
	if err != nil {
		return fmt.Errorf("ensuring collection: %w", err)
	}
	if created {
		s.logger.Info("created vector collection", zap.Int("dimensions", s.cfg.Dimensions))
	}

	vehicles, err := s.repo.ListVehicles(ctx)
	if err != nil {
		return fmt.Errorf("listing vehicles: %w", err)
	}

	indexed, err := s.store.ListIndexed(ctx)
	if err != nil {
		return fmt.Errorf("listing indexed vehicles: %w", err)
	}

	diff, err := diffIndex(vehicles, indexed)
	if err != nil {
		return err
	}
	// Embed and upsert batch by batch so progress persists if a later batch fails.
	for start := 0; start < len(diff.pending); start += s.cfg.BatchSize {
		end := min(start+s.cfg.BatchSize, len(diff.pending))
		if err := s.indexBatch(ctx, diff.pending[start:end]); err != nil {
			return err
		}
		s.logger.Debug("indexed vehicle batch", zap.Int("done", end), zap.Int("total", len(diff.pending)))
	}

	if err := s.store.Delete(ctx, diff.stale); err != nil {
		return fmt.Errorf("deleting stale vehicles: %w", err)
	}

	s.logger.Info("vector index synced",
		zap.Int("added", diff.added),
		zap.Int("updated", diff.updated),
		zap.Int("deleted", len(diff.stale)),
		zap.Int("unchanged", diff.unchanged),
	)
	return nil
}

type indexDiff struct {
	pending                   []pendingPoint
	stale                     []string
	added, updated, unchanged int
}

// diffIndex compares the vehicles in the database with the points already in
// the collection and works out which ones need (re)indexing or removal.
func diffIndex(vehicles []vehicle.Vehicle, indexed map[string]vectorstore.IndexedPoint) (indexDiff, error) {
	var diff indexDiff
	seen := make(map[string]bool, len(vehicles))

	for _, v := range vehicles {
		doc, err := buildDocument(v)
		if err != nil {
			return indexDiff{}, err
		}
		hash := documentHash(doc)
		id := vectorstore.PointID(v.UniqueCode)
		seen[id] = true

		existing, ok := indexed[id]
		switch {
		case !ok:
			diff.added++
		case existing.DocHash != hash || existing.VehicleID != v.ID:
			diff.updated++
		default:
			diff.unchanged++
			continue
		}
		diff.pending = append(diff.pending, pendingPoint{vehicle: v, doc: doc, hash: hash})
	}

	for id := range indexed {
		if !seen[id] {
			diff.stale = append(diff.stale, id)
		}
	}
	return diff, nil
}

func (s *SearchService) indexBatch(ctx context.Context, batch []pendingPoint) error {
	docs := make([]string, len(batch))
	for i, p := range batch {
		docs[i] = p.doc
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, docs)
	if err != nil {
		return fmt.Errorf("embedding vehicles: %w", err)
	}

	points := make([]vectorstore.Point, len(batch))
	for i, p := range batch {
		points[i] = vectorstore.Point{
			UniqueCode: p.vehicle.UniqueCode,
			VehicleID:  p.vehicle.ID,
			DocHash:    p.hash,
			Vector:     vectors[i],
			Payload: map[string]any{
				"make":  p.vehicle.Make.Name,
				"model": p.vehicle.Model.Name,
				"year":  p.vehicle.Year,
				"trim":  p.vehicle.Trim.Name,
			},
		}
	}
	if err := s.store.Upsert(ctx, points); err != nil {
		return fmt.Errorf("upserting vehicles: %w", err)
	}
	return nil
}

// SearchVehicles returns the vehicles that best match a natural language query,
// ordered by descending similarity.
func (s *SearchService) SearchVehicles(ctx context.Context, query string) ([]vehicle.Vehicle, error) {
	ctx, span := tracer.Start(ctx, "SearchVehicles")
	defer span.End()

	span.SetAttributes(attribute.String("search.query", query))
	s.logger.Debug("searching vehicles", zap.String("query", query))

	results, err := s.searchVehicles(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.logger.Error("failed to search vehicles", zap.String("query", query), zap.Error(err))
		return nil, err
	}

	span.SetAttributes(attribute.Int("vehicles.count", len(results)))
	s.logger.Debug("searched vehicles", zap.String("query", query), zap.Int("count", len(results)))
	return results, nil
}

func (s *SearchService) searchVehicles(ctx context.Context, query string) ([]vehicle.Vehicle, error) {
	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embedding query: %w", err)
	}

	hits, err := s.store.Search(ctx, vector, s.cfg.SearchLimit, s.cfg.MinScore)
	if err != nil {
		return nil, fmt.Errorf("searching vector index: %w", err)
	}
	if len(hits) == 0 {
		return []vehicle.Vehicle{}, nil
	}

	uniqueCodes := make([]string, len(hits))
	for i, h := range hits {
		uniqueCodes[i] = h.UniqueCode
	}

	vehicles, err := s.repo.GetVehiclesByUniqueCodes(ctx, uniqueCodes)
	if err != nil {
		return nil, fmt.Errorf("loading matched vehicles: %w", err)
	}
	byCode := make(map[string]vehicle.Vehicle, len(vehicles))
	for _, v := range vehicles {
		byCode[v.UniqueCode] = v
	}

	results := make([]vehicle.Vehicle, 0, len(hits))
	for _, h := range hits {
		if v, ok := byCode[h.UniqueCode]; ok {
			results = append(results, v)
		}
	}
	return results, nil
}
