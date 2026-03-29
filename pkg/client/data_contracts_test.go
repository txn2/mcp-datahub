package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDataContract(t *testing.T) {
	tests := []struct {
		name         string
		responseJSON string
		wantStatus   string
		wantNil      bool
	}{
		{
			name: "passing contract",
			responseJSON: `{
				"data": {
					"dataset": {
						"contract": {
							"urn": "urn:li:dataContract:test",
							"properties": {
								"entityUrn": "urn:li:dataset:test",
								"freshness": [
									{"assertion": {"urn": "urn:li:assertion:freshness-1"}}
								],
								"schema": [
									{"assertion": {"urn": "urn:li:assertion:schema-1"}}
								],
								"dataQuality": []
							},
							"status": {"state": "PASSING"}
						}
					}
				}
			}`,
			wantStatus: "PASSING",
		},
		{
			name: "failing contract",
			responseJSON: `{
				"data": {
					"dataset": {
						"contract": {
							"urn": "urn:li:dataContract:test",
							"properties": {
								"entityUrn": "urn:li:dataset:test",
								"freshness": [],
								"schema": [],
								"dataQuality": [
									{"assertion": {"urn": "urn:li:assertion:quality-1"}}
								]
							},
							"status": {"state": "FAILING"}
						}
					}
				}
			}`,
			wantStatus: "FAILING",
		},
		{
			name: "no contract",
			responseJSON: `{
				"data": {
					"dataset": {
						"contract": null
					}
				}
			}`,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.responseJSON))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			result, err := c.GetDataContract(context.Background(), "urn:li:dataset:test")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestGetDataContract_Details(t *testing.T) {
	responseJSON := `{
		"data": {
			"dataset": {
				"contract": {
					"urn": "urn:li:dataContract:test",
					"properties": {
						"entityUrn": "urn:li:dataset:test",
						"freshness": [
							{"assertion": {"urn": "urn:li:assertion:freshness-1"}}
						],
						"schema": [
							{"assertion": {"urn": "urn:li:assertion:schema-1"}}
						],
						"dataQuality": [
							{"assertion": {"urn": "urn:li:assertion:quality-1"}}
						]
					},
					"status": {"state": "FAILING"}
				}
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	result, err := c.GetDataContract(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.AssertionResults) != 3 {
		t.Fatalf("AssertionResults count = %d, want 3", len(result.AssertionResults))
	}

	// Freshness assertion
	ar := result.AssertionResults[0]
	if ar.AssertionURN != "urn:li:assertion:freshness-1" {
		t.Errorf("AssertionURN = %q", ar.AssertionURN)
	}
	if ar.Type != "FRESHNESS" {
		t.Errorf("Type = %q", ar.Type)
	}

	// Schema assertion
	ar = result.AssertionResults[1]
	if ar.AssertionURN != "urn:li:assertion:schema-1" {
		t.Errorf("AssertionURN = %q", ar.AssertionURN)
	}
	if ar.Type != "SCHEMA" {
		t.Errorf("Type = %q", ar.Type)
	}

	// Data quality assertion
	ar = result.AssertionResults[2]
	if ar.AssertionURN != "urn:li:assertion:quality-1" {
		t.Errorf("AssertionURN = %q", ar.AssertionURN)
	}
	if ar.Type != "DATA_QUALITY" {
		t.Errorf("Type = %q", ar.Type)
	}
}

func TestGetDataContract_GraphQLError_ReturnsNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors": [{"message": "contract field not found"}]}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	result, err := c.GetDataContract(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Fatalf("expected nil error for graceful degradation, got: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}
