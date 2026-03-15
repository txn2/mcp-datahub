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
							"result": {
								"type": "PASSING",
								"assertionResults": [
									{
										"assertion": {"urn": "urn:li:assertion:freshness"},
										"type": "FRESHNESS",
										"result": {
											"type": "SUCCESS",
											"nativeResults": [
												{"key": "last_updated", "value": "2024-01-01T00:00:00Z"}
											]
										}
									},
									{
										"assertion": {"urn": "urn:li:assertion:schema"},
										"type": "SCHEMA",
										"result": {
											"type": "SUCCESS",
											"nativeResults": []
										}
									}
								]
							}
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
							"result": {
								"type": "FAILING",
								"assertionResults": [
									{
										"assertion": {"urn": "urn:li:assertion:quality"},
										"type": "DATA_QUALITY",
										"result": {
											"type": "FAILURE",
											"nativeResults": [
												{"key": "null_count", "value": "150"},
												{"key": "threshold", "value": "100"}
											]
										}
									}
								]
							}
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
		{
			name: "null result",
			responseJSON: `{
				"data": {
					"dataset": {
						"contract": {
							"result": null
						}
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
					"result": {
						"type": "FAILING",
						"assertionResults": [{
							"assertion": {"urn": "urn:li:assertion:quality-1"},
							"type": "DATA_QUALITY",
							"result": {
								"type": "FAILURE",
								"nativeResults": [
									{"key": "null_count", "value": "150"},
									{"key": "threshold", "value": "100"}
								]
							}
						}]
					}
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

	if len(result.AssertionResults) != 1 {
		t.Fatalf("AssertionResults count = %d, want 1", len(result.AssertionResults))
	}

	ar := result.AssertionResults[0]
	if ar.AssertionURN != "urn:li:assertion:quality-1" {
		t.Errorf("AssertionURN = %q", ar.AssertionURN)
	}
	if ar.Type != "DATA_QUALITY" {
		t.Errorf("Type = %q", ar.Type)
	}
	if ar.ResultType != "FAILURE" {
		t.Errorf("ResultType = %q", ar.ResultType)
	}
	if ar.NativeResults["null_count"] != "150" {
		t.Errorf("NativeResults[null_count] = %q", ar.NativeResults["null_count"])
	}
	if ar.NativeResults["threshold"] != "100" {
		t.Errorf("NativeResults[threshold] = %q", ar.NativeResults["threshold"])
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
