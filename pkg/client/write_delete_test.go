package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// extractGraphQLVariables decodes a GraphQL request body and returns all variables.
func extractGraphQLVariables(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var req graphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		t.Fatalf("failed to decode GraphQL request: %v", err)
	}
	return req.Variables
}

func TestDeleteTag(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:tag:Deprecated",
			response: `{"data": {"deleteTag": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:tag:NotFound",
			response: `{"errors": [{"message": "tag not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["urn"] != tt.urn {
					t.Errorf("urn = %v, want %v", vars["urn"], tt.urn)
				}
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

			err := c.DeleteTag(context.Background(), tt.urn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteDomain(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:domain:engineering",
			response: `{"data": {"deleteDomain": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:domain:missing",
			response: `{"errors": [{"message": "domain not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["urn"] != tt.urn {
					t.Errorf("urn = %v, want %v", vars["urn"], tt.urn)
				}
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

			err := c.DeleteDomain(context.Background(), tt.urn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteGlossaryEntity(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:glossaryTerm:revenue",
			response: `{"data": {"deleteGlossaryEntity": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:glossaryTerm:unknown",
			response: `{"errors": [{"message": "entity not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["urn"] != tt.urn {
					t.Errorf("urn = %v, want %v", vars["urn"], tt.urn)
				}
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

			err := c.DeleteGlossaryEntity(context.Background(), tt.urn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteDataProduct(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:dataProduct:analytics",
			response: `{"data": {"deleteDataProduct": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:dataProduct:missing",
			response: `{"errors": [{"message": "not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["urn"] != tt.urn {
					t.Errorf("urn = %v, want %v", vars["urn"], tt.urn)
				}
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

			err := c.DeleteDataProduct(context.Background(), tt.urn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteApplication(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:application:myapp",
			response: `{"data": {"deleteApplication": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:application:missing",
			response: `{"errors": [{"message": "application not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["urn"] != tt.urn {
					t.Errorf("urn = %v, want %v", vars["urn"], tt.urn)
				}
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

			err := c.DeleteApplication(context.Background(), tt.urn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteStructuredProperty(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:structuredProperty:prop1",
			response: `{"data": {"deleteStructuredProperty": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:structuredProperty:missing",
			response: `{"errors": [{"message": "property not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["urn"] != tt.urn {
					t.Errorf("input.urn = %v, want %v", input["urn"], tt.urn)
				}
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

			err := c.DeleteStructuredProperty(context.Background(), tt.urn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
