package embedding

import (
	"context"
	"math"
	"time"
)

const (
	taskRetrievalDocument = "RETRIEVAL_DOCUMENT"
	taskRetrievalQuery    = "RETRIEVAL_QUERY"

	// maxRetries bounds retries of rate-limited or transiently failing requests.
	maxRetries = 10
	// defaultRetryDelay is used when the API does not say how long to wait.
	defaultRetryDelay = 30 * time.Second
	// retryDelaySlack is added to the server-provided delay so the quota window has fully reset.
	retryDelaySlack = time.Second
	// maxRetryDelay caps how long to wait before a retry. Longer delays mean a
	// daily quota is exhausted, so failing is better than blocking for hours;
	// the sync is incremental and resumes on the next run.
	maxRetryDelay = 2 * time.Minute
)

// Embedder turns text into dense vectors. Documents and queries are embedded
// separately because retrieval models optimise them for different roles.
type Embedder interface {
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}

// normalize scales v to unit length in place.
func normalize(v []float32) []float32 {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if sum == 0 {
		return v
	}
	norm := float32(math.Sqrt(sum))
	for i := range v {
		v[i] /= norm
	}
	return v
}
