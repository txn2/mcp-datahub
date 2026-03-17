package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestGetIncidents(t *testing.T) {
	tests := []struct {
		name         string
		responseJSON string
		wantTotal    int
		wantCount    int
		wantErr      bool
	}{
		{
			name: "active incidents",
			responseJSON: `{
				"data": {
					"entity": {
						"incidents": {
							"total": 2,
							"incidents": [
								{
									"urn": "urn:li:incident:1",
									"type": "OPERATIONAL",
									"customType": "",
									"title": "Pipeline failure",
									"description": "ETL failed",
									"status": {
										"state": "ACTIVE",
										"lastUpdated": {"time": 1700001000000, "actor": "urn:li:corpuser:admin"}
									},
									"source": {"type": "MANUAL"},
									"created": {"time": 1700000000000, "actor": "urn:li:corpuser:admin"}
								},
								{
									"urn": "urn:li:incident:2",
									"type": "CUSTOM",
									"customType": "SLA_BREACH",
									"title": "SLA breach",
									"description": "",
									"status": {
										"state": "ACTIVE",
										"lastUpdated": {"time": 1700002000000, "actor": "urn:li:corpuser:bot"}
									},
									"source": {"type": "AUTOMATED"},
									"created": {"time": 1700001500000, "actor": "urn:li:corpuser:bot"}
								}
							]
						}
					}
				}
			}`,
			wantTotal: 2,
			wantCount: 2,
		},
		{
			name: "no incidents",
			responseJSON: `{
				"data": {
					"entity": {
						"incidents": {
							"total": 0,
							"incidents": []
						}
					}
				}
			}`,
			wantTotal: 0,
			wantCount: 0,
		},
		{
			name: "null incidents aspect",
			responseJSON: `{
				"data": {
					"entity": {
						"incidents": null
					}
				}
			}`,
			wantTotal: 0,
			wantCount: 0,
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
				config:     DefaultConfig(),
				logger:     NopLogger{},
			}

			result, err := c.GetIncidents(context.Background(), "urn:li:dataset:test")
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if result.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", result.Total, tt.wantTotal)
			}
			if len(result.Incidents) != tt.wantCount {
				t.Errorf("Incidents count = %d, want %d", len(result.Incidents), tt.wantCount)
			}
		})
	}
}

func TestGetIncidents_Details(t *testing.T) {
	responseJSON := `{
		"data": {
			"entity": {
				"incidents": {
					"total": 1,
					"incidents": [{
						"urn": "urn:li:incident:1",
						"type": "OPERATIONAL",
						"customType": "OUTAGE",
						"title": "DB outage",
						"description": "Primary replica down",
						"status": {
							"state": "ACTIVE",
							"lastUpdated": {"time": 1700001000000, "actor": "urn:li:corpuser:oncall"}
						},
						"source": {"type": "MANUAL"},
						"created": {"time": 1700000000000, "actor": "urn:li:corpuser:admin"}
					}]
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
		config:     DefaultConfig(),
		logger:     NopLogger{},
	}

	result, err := c.GetIncidents(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inc := result.Incidents[0]
	if inc.URN != "urn:li:incident:1" {
		t.Errorf("URN = %q", inc.URN)
	}
	if inc.Type != "OPERATIONAL" {
		t.Errorf("Type = %q", inc.Type)
	}
	if inc.CustomType != "OUTAGE" {
		t.Errorf("CustomType = %q", inc.CustomType)
	}
	if inc.Title != "DB outage" {
		t.Errorf("Title = %q", inc.Title)
	}
	if inc.State != "ACTIVE" {
		t.Errorf("State = %q", inc.State)
	}
	if inc.Source != "MANUAL" {
		t.Errorf("Source = %q", inc.Source)
	}
	if inc.Created != 1700000000000 {
		t.Errorf("Created = %d", inc.Created)
	}
	if inc.CreatedBy != "urn:li:corpuser:admin" {
		t.Errorf("CreatedBy = %q", inc.CreatedBy)
	}
	if inc.LastUpdated != 1700001000000 {
		t.Errorf("LastUpdated = %d", inc.LastUpdated)
	}
	if inc.LastUpdatedBy != "urn:li:corpuser:oncall" {
		t.Errorf("LastUpdatedBy = %q", inc.LastUpdatedBy)
	}
}

func TestGetIncidents_GraphQLError_ReturnsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors": [{"message": "incidents field not found"}]}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		config:     DefaultConfig(),
		logger:     NopLogger{},
	}

	result, err := c.GetIncidents(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Fatalf("expected nil error for graceful degradation, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected empty result, got total %d", result.Total)
	}
}

func TestRaiseIncident(t *testing.T) {
	tests := []struct {
		name     string
		input    types.RaiseIncidentInput
		response string
		wantURN  string
		wantErr  bool
	}{
		{
			name: "success",
			input: types.RaiseIncidentInput{
				Type:         "OPERATIONAL",
				Title:        "Pipeline down",
				Description:  "ETL failed",
				ResourceURNs: []string{"urn:li:dataset:test"},
			},
			response: `{"data": {"raiseIncident": "urn:li:incident:new-1"}}`,
			wantURN:  "urn:li:incident:new-1",
		},
		{
			name: "custom type",
			input: types.RaiseIncidentInput{
				Type:         "CUSTOM",
				CustomType:   "SLA_BREACH",
				Title:        "SLA breach",
				ResourceURNs: []string{"urn:li:dataset:test"},
			},
			response: `{"data": {"raiseIncident": "urn:li:incident:new-2"}}`,
			wantURN:  "urn:li:incident:new-2",
		},
		{
			name: "graphql error",
			input: types.RaiseIncidentInput{
				Type:         "OPERATIONAL",
				Title:        "Test",
				ResourceURNs: []string{"urn:li:dataset:test"},
			},
			response: `{"errors": [{"message": "mutation not supported"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&receivedBody)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			urn, err := c.RaiseIncident(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}

			// Verify resourceUrns is sent as an array
			if !tt.wantErr && receivedBody != nil {
				vars, _ := receivedBody["variables"].(map[string]any)
				input, _ := vars["input"].(map[string]any)
				resourceUrns, ok := input["resourceUrns"].([]any)
				if !ok {
					t.Fatalf("expected resourceUrns array, got %T", input["resourceUrns"])
				}
				if len(resourceUrns) != len(tt.input.ResourceURNs) {
					t.Fatalf("resourceUrns length = %d, want %d", len(resourceUrns), len(tt.input.ResourceURNs))
				}
				if resourceUrns[0] != tt.input.ResourceURNs[0] {
					t.Errorf("resourceUrns[0] = %q, want %q", resourceUrns[0], tt.input.ResourceURNs[0])
				}
			}
		})
	}
}

func TestRaiseIncident_EmptyResourceURNs(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	_, err := c.RaiseIncident(context.Background(), types.RaiseIncidentInput{
		Type:         "OPERATIONAL",
		Title:        "Test",
		ResourceURNs: []string{},
	})
	if err == nil {
		t.Fatal("expected error for empty resource URNs")
	}
}

func TestResolveIncident(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			response: `{"data": {"updateIncidentStatus": true}}`,
		},
		{
			name:     "graphql error",
			response: `{"errors": [{"message": "incident not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			err := c.ResolveIncident(context.Background(), "urn:li:incident:1", "Issue resolved")
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
