package service

import (
	"context"
	"slices"
	"testing"

	"github.com/ChargePi/openev-data-mcp/internal/vectorstore"
	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
	"go.uber.org/zap"
)

type fakeRepo struct {
	vehicles []vehicle.Vehicle
}

func (r *fakeRepo) ListVehicles(context.Context) ([]vehicle.Vehicle, error) {
	return r.vehicles, nil
}

func (r *fakeRepo) GetVehiclesByUniqueCodes(_ context.Context, codes []string) ([]vehicle.Vehicle, error) {
	var out []vehicle.Vehicle
	for _, v := range r.vehicles {
		if slices.Contains(codes, v.UniqueCode) {
			out = append(out, v)
		}
	}
	return out, nil
}

type fakeEmbedder struct {
	embedded []string
}

func (e *fakeEmbedder) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	e.embedded = append(e.embedded, texts...)
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = []float32{1, 0}
	}
	return out, nil
}

func (e *fakeEmbedder) EmbedQuery(context.Context, string) ([]float32, error) {
	return []float32{1, 0}, nil
}

type fakeStore struct {
	points       map[string]vectorstore.Point
	hits         []vectorstore.Hit
	lastMinScore float32
}

func newFakeStore() *fakeStore {
	return &fakeStore{points: map[string]vectorstore.Point{}}
}

func (s *fakeStore) EnsureCollection(context.Context, uint64) (bool, error) { return false, nil }

func (s *fakeStore) ListIndexed(context.Context) (map[string]vectorstore.IndexedPoint, error) {
	out := make(map[string]vectorstore.IndexedPoint, len(s.points))
	for id, p := range s.points {
		out[id] = vectorstore.IndexedPoint{VehicleID: p.VehicleID, DocHash: p.DocHash}
	}
	return out, nil
}

func (s *fakeStore) Upsert(_ context.Context, points []vectorstore.Point) error {
	for _, p := range points {
		s.points[vectorstore.PointID(p.UniqueCode)] = p
	}
	return nil
}

func (s *fakeStore) Delete(_ context.Context, ids []string) error {
	for _, id := range ids {
		delete(s.points, id)
	}
	return nil
}

func (s *fakeStore) Search(_ context.Context, _ []float32, _ int, minScore float32) ([]vectorstore.Hit, error) {
	s.lastMinScore = minScore
	var out []vectorstore.Hit
	for _, h := range s.hits {
		if h.Score >= minScore {
			out = append(out, h)
		}
	}
	return out, nil
}

func testVehicle(id int, code, model string) vehicle.Vehicle {
	return vehicle.Vehicle{
		ID:          id,
		UniqueCode:  code,
		Make:        vehicle.NamedEntity{Name: "Make"},
		Model:       vehicle.NamedEntity{Name: model},
		Year:        2024,
		Trim:        vehicle.NamedEntity{Name: "Base"},
		VehicleType: "passenger_car",
	}
}

func newTestService(repo *fakeRepo, emb *fakeEmbedder, store *fakeStore) *SearchService {
	return NewSearchService(repo, emb, store, SearchConfig{Dimensions: 2, BatchSize: 2, MinScore: 0.6}, zap.NewNop())
}

func TestSyncIndex(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{vehicles: []vehicle.Vehicle{
		testVehicle(1, "a", "A"),
		testVehicle(2, "b", "B"),
		testVehicle(3, "c", "C"),
	}}
	emb := &fakeEmbedder{}
	store := newFakeStore()
	svc := newTestService(repo, emb, store)

	// Initial seed indexes everything, across multiple batches.
	if err := svc.SyncIndex(ctx); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if len(store.points) != 3 || len(emb.embedded) != 3 {
		t.Fatalf("seed: got %d points, %d embeddings; want 3, 3", len(store.points), len(emb.embedded))
	}

	// No changes: nothing is embedded.
	emb.embedded = nil
	if err := svc.SyncIndex(ctx); err != nil {
		t.Fatalf("resync: %v", err)
	}
	if len(emb.embedded) != 0 {
		t.Fatalf("resync: embedded %d documents, want 0", len(emb.embedded))
	}

	// New vehicle, changed vehicle and removed vehicle.
	repo.vehicles = []vehicle.Vehicle{
		testVehicle(1, "a", "A"),
		testVehicle(2, "b", "B changed"),
		testVehicle(4, "d", "D"),
	}
	emb.embedded = nil
	if err := svc.SyncIndex(ctx); err != nil {
		t.Fatalf("diff sync: %v", err)
	}
	if len(emb.embedded) != 2 {
		t.Fatalf("diff sync: embedded %d documents, want 2", len(emb.embedded))
	}
	if _, ok := store.points[vectorstore.PointID("c")]; ok {
		t.Errorf("removed vehicle c is still indexed")
	}
	if _, ok := store.points[vectorstore.PointID("d")]; !ok {
		t.Errorf("new vehicle d is not indexed")
	}
	if len(store.points) != 3 {
		t.Errorf("got %d points, want 3", len(store.points))
	}

	// A changed vehicle ID (e.g. re-created database) triggers a re-upsert.
	repo.vehicles[0].ID = 10
	emb.embedded = nil
	if err := svc.SyncIndex(ctx); err != nil {
		t.Fatalf("id change sync: %v", err)
	}
	if len(emb.embedded) != 1 || store.points[vectorstore.PointID("a")].VehicleID != 10 {
		t.Fatalf("id change sync: embedded %d documents, want 1 with updated vehicle ID", len(emb.embedded))
	}
}

func TestSearchVehicles_PreservesRankingAndFiltersLowScores(t *testing.T) {
	repo := &fakeRepo{vehicles: []vehicle.Vehicle{
		testVehicle(1, "a", "A"),
		testVehicle(2, "b", "B"),
		testVehicle(3, "c", "C"),
	}}
	store := newFakeStore()
	store.hits = []vectorstore.Hit{
		{UniqueCode: "c", Score: 0.9},
		{UniqueCode: "missing", Score: 0.8},
		{UniqueCode: "a", Score: 0.7},
		{UniqueCode: "b", Score: 0.59},
	}
	svc := newTestService(repo, &fakeEmbedder{}, store)

	results, err := svc.SearchVehicles(context.Background(), "query")
	if err != nil {
		t.Fatalf("SearchVehicles: %v", err)
	}
	if store.lastMinScore != 0.6 {
		t.Fatalf("min score passed to store = %v, want 0.6", store.lastMinScore)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].UniqueCode != "c" || results[1].UniqueCode != "a" {
		t.Fatalf("unexpected results order: %+v", results)
	}
}
