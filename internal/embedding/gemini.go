package embedding

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ChargePi/openev-data-mcp/internal/config"
	"go.uber.org/zap"
	"google.golang.org/genai"
)

// GeminiEmbedder embeds text using the Gemini API.
type GeminiEmbedder struct {
	client     *genai.Client
	model      string
	dimensions int32
	batchSize  int
	logger     *zap.Logger
}

func NewGeminiEmbedder(ctx context.Context, cfg config.EmbeddingConfig, logger *zap.Logger) (*GeminiEmbedder, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("creating Gemini client: %w", err)
	}

	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	return &GeminiEmbedder{
		client:     client,
		model:      cfg.Model,
		dimensions: int32(cfg.Dimensions),
		batchSize:  batchSize,
		logger:     logger,
	}, nil
}

func (g *GeminiEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, 0, len(texts))
	for start := 0; start < len(texts); start += g.batchSize {
		end := min(start+g.batchSize, len(texts))
		batch, err := g.embed(ctx, texts[start:end], taskRetrievalDocument)
		if err != nil {
			return nil, fmt.Errorf("embedding documents %d-%d: %w", start, end, err)
		}
		vectors = append(vectors, batch...)
	}
	return vectors, nil
}

func (g *GeminiEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	vectors, err := g.embed(ctx, []string{text}, taskRetrievalQuery)
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

func (g *GeminiEmbedder) embed(ctx context.Context, texts []string, taskType string) ([][]float32, error) {
	contents := make([]*genai.Content, len(texts))
	for i, t := range texts {
		contents[i] = genai.NewContentFromText(t, genai.RoleUser)
	}

	resp, err := g.embedWithRetry(ctx, contents, taskType)
	if err != nil {
		return nil, err
	}
	if len(resp.Embeddings) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(resp.Embeddings))
	}

	vectors := make([][]float32, len(resp.Embeddings))
	for i, e := range resp.Embeddings {
		if len(e.Values) != int(g.dimensions) {
			return nil, fmt.Errorf("embedding %d has %d dimensions, expected %d", i, len(e.Values), g.dimensions)
		}
		// Gemini only normalises full-size (3072) embeddings; truncated ones must be normalised by the caller.
		vectors[i] = normalize(e.Values)
	}
	return vectors, nil
}

// embedWithRetry calls the embed API, waiting and retrying when the request is
// rate limited or fails transiently.
func (g *GeminiEmbedder) embedWithRetry(ctx context.Context, contents []*genai.Content, taskType string) (*genai.EmbedContentResponse, error) {
	cfg := &genai.EmbedContentConfig{
		TaskType:             taskType,
		OutputDimensionality: &g.dimensions,
	}

	for attempt := 0; ; attempt++ {
		resp, err := g.client.Models.EmbedContent(ctx, g.model, contents, cfg)
		if err == nil {
			return resp, nil
		}

		delay, retryable := retryDelay(err)
		if !retryable || attempt >= maxRetries || delay > maxRetryDelay {
			return nil, fmt.Errorf("calling Gemini embed API: %w", err)
		}

		g.logger.Warn("Gemini embed request failed, retrying",
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", maxRetries),
			zap.Duration("delay", delay),
			zap.Error(err),
		)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("waiting to retry Gemini embed API: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

// retryDelay reports whether err is worth retrying and how long to wait first,
// preferring the delay the API asks for in its RetryInfo details.
func retryDelay(err error) (time.Duration, bool) {
	var apiErr genai.APIError
	if !errors.As(err, &apiErr) {
		return 0, false
	}

	switch apiErr.Code {
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusServiceUnavailable:
	default:
		return 0, false
	}

	for _, detail := range apiErr.Details {
		typ, _ := detail["@type"].(string)
		if !strings.HasSuffix(typ, "RetryInfo") {
			continue
		}
		raw, _ := detail["retryDelay"].(string)
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			return d + retryDelaySlack, true
		}
	}
	return defaultRetryDelay, true
}
