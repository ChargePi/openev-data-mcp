package vectorstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ChargePi/openev-data-mcp/internal/config"
)

// Runs only when QDRANT_TEST_HOST is set, against a throwaway collection.
func TestStoreIntegration(t *testing.T) {
	host := os.Getenv("QDRANT_TEST_HOST")
	if host == "" {
		t.Skip("QDRANT_TEST_HOST not set")
	}

	ctx := context.Background()
	collection := fmt.Sprintf("test_ev_vehicles_%d", time.Now().UnixNano())
	store, err := Connect(config.QdrantConfig{Host: host, Port: 6334, Collection: collection})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() {
		_ = store.client.DeleteCollection(context.Background(), collection)
		_ = store.Close()
	})

	created, err := store.EnsureCollection(ctx, 2)
	if err != nil || !created {
		t.Fatalf("EnsureCollection: created=%v err=%v", created, err)
	}
	if created, err = store.EnsureCollection(ctx, 2); err != nil || created {
		t.Fatalf("EnsureCollection (existing): created=%v err=%v", created, err)
	}
	if _, err = store.EnsureCollection(ctx, 3); err == nil {
		t.Fatalf("EnsureCollection with mismatched size should fail")
	}

	err = store.Upsert(ctx, []Point{
		{UniqueCode: "a", VehicleID: 1, DocHash: "ha", Vector: []float32{1, 0}, Payload: map[string]any{"make": "A"}},
		{UniqueCode: "b", VehicleID: 2, DocHash: "hb", Vector: []float32{0, 1}},
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	indexed, err := store.ListIndexed(ctx)
	if err != nil {
		t.Fatalf("ListIndexed: %v", err)
	}
	if got := indexed[PointID("a")]; len(indexed) != 2 || got.VehicleID != 1 || got.DocHash != "ha" {
		t.Fatalf("ListIndexed: unexpected %+v", indexed)
	}

	hits, err := store.Search(ctx, []float32{0.1, 0.9}, 10, 0)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 2 || hits[0].UniqueCode != "b" {
		t.Fatalf("Search: unexpected %+v", hits)
	}

	// Similarity to "a" is ~0.11, so a 0.6 threshold keeps only "b".
	hits, err = store.Search(ctx, []float32{0.1, 0.9}, 10, 0.6)
	if err != nil {
		t.Fatalf("Search with min score: %v", err)
	}
	if len(hits) != 1 || hits[0].UniqueCode != "b" {
		t.Fatalf("Search with min score: unexpected %+v", hits)
	}

	if err := store.Delete(ctx, []string{PointID("b")}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if indexed, _ = store.ListIndexed(ctx); len(indexed) != 1 {
		t.Fatalf("after Delete: got %d points, want 1", len(indexed))
	}
}
