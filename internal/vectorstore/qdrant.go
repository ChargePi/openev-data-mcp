package vectorstore

import (
	"context"
	"fmt"

	"github.com/ChargePi/openev-data-mcp/internal/config"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

const (
	payloadUniqueCode = "unique_code"
	payloadVehicleID  = "vehicle_id"
	payloadDocHash    = "doc_hash"

	scrollPageSize = 256
)

// pointNamespace seeds deterministic point IDs derived from vehicle unique codes.
var pointNamespace = uuid.MustParse("6f1d2c3a-8b4e-4f5a-9c7d-0e1f2a3b4c5d")

// PointID returns the stable Qdrant point ID for a vehicle unique code.
func PointID(uniqueCode string) string {
	return uuid.NewSHA1(pointNamespace, []byte(uniqueCode)).String()
}

// Point is a single vehicle document to be stored in the collection.
type Point struct {
	UniqueCode string
	VehicleID  int
	DocHash    string
	Vector     []float32
	Payload    map[string]any
}

// IndexedPoint is the sync-relevant state of a point already in the collection.
type IndexedPoint struct {
	VehicleID int
	DocHash   string
}

// Hit is a single search result.
type Hit struct {
	UniqueCode string
	Score      float32
}

type Store struct {
	client     *qdrant.Client
	collection string
}

func Connect(cfg config.QdrantConfig) (*Store, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   cfg.Host,
		Port:   cfg.Port,
		APIKey: cfg.APIKey,
		UseTLS: cfg.UseTLS,
		// The check logs via the stdlib logger and cannot read the client version from module info.
		SkipCompatibilityCheck: true,
	})
	if err != nil {
		return nil, fmt.Errorf("creating qdrant client: %w", err)
	}
	return &Store{client: client, collection: cfg.Collection}, nil
}

func (s *Store) Close() error {
	if err := s.client.Close(); err != nil {
		return fmt.Errorf("closing qdrant client: %w", err)
	}
	return nil
}

// EnsureCollection creates the collection if it does not exist and verifies
// that an existing collection uses the expected vector size.
func (s *Store) EnsureCollection(ctx context.Context, size uint64) (created bool, err error) {
	exists, err := s.client.CollectionExists(ctx, s.collection)
	if err != nil {
		return false, fmt.Errorf("checking collection %q: %w", s.collection, err)
	}

	if exists {
		info, err := s.client.GetCollectionInfo(ctx, s.collection)
		if err != nil {
			return false, fmt.Errorf("getting collection %q info: %w", s.collection, err)
		}
		existing := info.GetConfig().GetParams().GetVectorsConfig().GetParams().GetSize()
		if existing != size {
			return false, fmt.Errorf("collection %q has vector size %d, expected %d", s.collection, existing, size)
		}
		return false, nil
	}

	err = s.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: s.collection,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     size,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return false, fmt.Errorf("creating collection %q: %w", s.collection, err)
	}
	return true, nil
}

// ListIndexed returns every point in the collection keyed by point ID.
func (s *Store) ListIndexed(ctx context.Context) (map[string]IndexedPoint, error) {
	indexed := make(map[string]IndexedPoint)
	limit := uint32(scrollPageSize)

	var offset *qdrant.PointId
	for {
		points, next, err := s.client.ScrollAndOffset(ctx, &qdrant.ScrollPoints{
			CollectionName: s.collection,
			Offset:         offset,
			Limit:          &limit,
			WithPayload:    qdrant.NewWithPayloadInclude(payloadVehicleID, payloadDocHash),
		})
		if err != nil {
			return nil, fmt.Errorf("scrolling collection %q: %w", s.collection, err)
		}

		for _, p := range points {
			payload := p.GetPayload()
			indexed[p.GetId().GetUuid()] = IndexedPoint{
				VehicleID: int(payload[payloadVehicleID].GetIntegerValue()),
				DocHash:   payload[payloadDocHash].GetStringValue(),
			}
		}

		if next == nil {
			return indexed, nil
		}
		offset = next
	}
}

func (s *Store) Upsert(ctx context.Context, points []Point) error {
	if len(points) == 0 {
		return nil
	}

	structs := make([]*qdrant.PointStruct, len(points))
	for i, p := range points {
		payload := map[string]any{
			payloadUniqueCode: p.UniqueCode,
			payloadVehicleID:  p.VehicleID,
			payloadDocHash:    p.DocHash,
		}
		for k, v := range p.Payload {
			payload[k] = v
		}

		structs[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(PointID(p.UniqueCode)),
			Vectors: qdrant.NewVectorsDense(p.Vector),
			Payload: qdrant.NewValueMap(payload),
		}
	}

	wait := true
	_, err := s.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: s.collection,
		Wait:           &wait,
		Points:         structs,
	})
	if err != nil {
		return fmt.Errorf("upserting %d points: %w", len(points), err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, pointIDs []string) error {
	if len(pointIDs) == 0 {
		return nil
	}

	ids := make([]*qdrant.PointId, len(pointIDs))
	for i, id := range pointIDs {
		ids[i] = qdrant.NewIDUUID(id)
	}

	wait := true
	_, err := s.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: s.collection,
		Wait:           &wait,
		Points:         qdrant.NewPointsSelectorIDs(ids),
	})
	if err != nil {
		return fmt.Errorf("deleting %d points: %w", len(pointIDs), err)
	}
	return nil
}

// Search returns up to limit points whose similarity to vector is at least minScore.
func (s *Store) Search(ctx context.Context, vector []float32, limit int, minScore float32) ([]Hit, error) {
	l := uint64(limit)
	points, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: s.collection,
		Query:          qdrant.NewQueryDense(vector),
		Limit:          &l,
		ScoreThreshold: &minScore,
		WithPayload:    qdrant.NewWithPayloadInclude(payloadUniqueCode),
	})
	if err != nil {
		return nil, fmt.Errorf("querying collection %q: %w", s.collection, err)
	}

	hits := make([]Hit, len(points))
	for i, p := range points {
		hits[i] = Hit{
			UniqueCode: p.GetPayload()[payloadUniqueCode].GetStringValue(),
			Score:      p.GetScore(),
		}
	}
	return hits, nil
}
