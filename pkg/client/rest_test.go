package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "standard endpoint",
			endpoint: "https://datahub.example.com/api/graphql",
			want:     "https://datahub.example.com",
		},
		{
			name:     "localhost endpoint",
			endpoint: "http://localhost:8080/api/graphql",
			want:     "http://localhost:8080",
		},
		{
			name:     "endpoint with path prefix",
			endpoint: "https://datahub.example.com/prefix/api/graphql",
			want:     "https://datahub.example.com/prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{endpoint: tt.endpoint}
			got := c.restBaseURL()
			if got != tt.want {
				t.Errorf("restBaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetAspect(t *testing.T) {
	expectedValue := json.RawMessage(`{"tags":["test"]}`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and path
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/aspects/urn:li:dataset:test" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("aspect") != "globalTags" {
			t.Errorf("unexpected aspect param: %s", r.URL.Query().Get("aspect"))
		}
		if r.URL.Query().Get("version") != "0" {
			t.Errorf("unexpected version param: %s", r.URL.Query().Get("version"))
		}

		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-RestLi-Protocol-Version") != "2.0.0" {
			t.Errorf("unexpected protocol version header: %s", r.Header.Get("X-RestLi-Protocol-Version"))
		}

		resp := aspectResponse{Value: expectedValue}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	result, err := c.getAspect(context.Background(), "urn:li:dataset:test", "globalTags")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(result) != string(expectedValue) {
		t.Errorf("got %s, want %s", string(result), string(expectedValue))
	}
}

func TestGetAspect_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.getAspect(context.Background(), "urn:li:dataset:missing", "globalTags")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestGetAspect_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "bad-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.getAspect(context.Background(), "urn:li:dataset:test", "globalTags")
	if err != ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestPostIngestProposal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and path
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/aspects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("action") != "ingestProposal" {
			t.Errorf("unexpected action param: %s", r.URL.Query().Get("action"))
		}

		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-RestLi-Protocol-Version") != "2.0.0" {
			t.Errorf("unexpected protocol version: %s", r.Header.Get("X-RestLi-Protocol-Version"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}

		// Verify request body
		var req ingestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Proposal.EntityURN != "urn:li:dataset:test" {
			t.Errorf("unexpected URN: %s", req.Proposal.EntityURN)
		}
		if req.Proposal.AspectName != "globalTags" {
			t.Errorf("unexpected aspect: %s", req.Proposal.AspectName)
		}
		if req.Proposal.EntityType != "dataset" {
			t.Errorf("unexpected entity type: %s", req.Proposal.EntityType)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":""}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.postIngestProposal(context.Background(), ingestProposal{
		EntityType: "dataset",
		EntityURN:  "urn:li:dataset:test",
		AspectName: "globalTags",
		Aspect:     map[string]any{"tags": []any{}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostIngestProposal_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.postIngestProposal(context.Background(), ingestProposal{
		EntityType: "dataset",
		EntityURN:  "urn:li:dataset:test",
		AspectName: "globalTags",
		Aspect:     map[string]any{},
	})
	if err != ErrForbidden {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestGetAspect_InvalidResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.getAspect(context.Background(), "urn:li:dataset:test", "globalTags")
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestPostIngestProposal_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`internal server error`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.postIngestProposal(context.Background(), ingestProposal{
		EntityType: "dataset",
		EntityURN:  "urn:li:dataset:test",
		AspectName: "globalTags",
		Aspect:     map[string]any{},
	})
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestPostIngestProposal_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.postIngestProposal(context.Background(), ingestProposal{
		EntityType: "dataset",
		EntityURN:  "urn:li:dataset:test",
		AspectName: "globalTags",
		Aspect:     map[string]any{},
	})
	if err != ErrRateLimited {
		t.Errorf("expected ErrRateLimited, got: %v", err)
	}
}

func TestCheckRESTStatus(t *testing.T) {
	c := &Client{logger: NopLogger{}}

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
		wantNil    bool
	}{
		{"ok", http.StatusOK, "", nil, true},
		{"unauthorized", http.StatusUnauthorized, "", ErrUnauthorized, false},
		{"forbidden", http.StatusForbidden, "", ErrForbidden, false},
		{"not found", http.StatusNotFound, "", ErrNotFound, false},
		{"rate limited", http.StatusTooManyRequests, "", ErrRateLimited, false},
		{"server error", http.StatusInternalServerError, "internal error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.checkRESTStatus(tt.statusCode, []byte(tt.body))
			if tt.wantNil {
				if err != nil {
					t.Errorf("expected nil error, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tt.wantErr != nil && err != tt.wantErr {
				t.Errorf("expected %v, got: %v", tt.wantErr, err)
			}
		})
	}
}
