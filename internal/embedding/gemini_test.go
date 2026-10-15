package embedding

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"google.golang.org/genai"
)

type retryDelayCase struct {
	name      string
	err       error
	wantDelay time.Duration
	wantRetry bool
}

func TestRetryDelay(t *testing.T) {
	tests := []retryDelayCase{
		{
			name: "rate limited with retry info",
			err: genai.APIError{Code: 429, Details: []map[string]any{
				{"@type": "type.googleapis.com/google.rpc.QuotaFailure"},
				{"@type": "type.googleapis.com/google.rpc.RetryInfo", "retryDelay": "13s"},
			}},
			wantDelay: 13*time.Second + retryDelaySlack,
			wantRetry: true,
		},
		{
			name:      "rate limited without retry info",
			err:       genai.APIError{Code: 429},
			wantDelay: defaultRetryDelay,
			wantRetry: true,
		},
		{
			name:      "wrapped unavailable",
			err:       fmt.Errorf("wrapped: %w", genai.APIError{Code: 503}),
			wantDelay: defaultRetryDelay,
			wantRetry: true,
		},
		{
			name:      "bad request",
			err:       genai.APIError{Code: 400},
			wantRetry: false,
		},
		{
			name:      "non-API error",
			err:       errors.New("boom"),
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay, retry := retryDelay(tt.err)
			if retry != tt.wantRetry || delay != tt.wantDelay {
				t.Fatalf("retryDelay() = (%v, %v), want (%v, %v)", delay, retry, tt.wantDelay, tt.wantRetry)
			}
		})
	}
}
