package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// extractGraphQLInput decodes a GraphQL request body and returns the "input" variable.
func extractGraphQLInput(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var req graphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		t.Fatalf("failed to decode GraphQL request: %v", err)
	}
	input, ok := req.Variables["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'input' variable to be a map, got %T", req.Variables["input"])
	}
	return input
}

func TestAddTagGraphQL(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		tagURN   string
		response string
		wantErr  bool
	}{
		{
			name:     "domain entity",
			urn:      "urn:li:domain:engineering",
			tagURN:   "urn:li:tag:PII",
			response: `{"data": {"addTag": true}}`,
		},
		{
			name:     "glossaryTerm entity",
			urn:      "urn:li:glossaryTerm:revenue",
			tagURN:   "urn:li:tag:Sensitive",
			response: `{"data": {"addTag": true}}`,
		},
		{
			name:     "glossaryNode entity",
			urn:      "urn:li:glossaryNode:finance",
			tagURN:   "urn:li:tag:PII",
			response: `{"data": {"addTag": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:domain:engineering",
			tagURN:   "urn:li:tag:PII",
			response: `{"errors": [{"message": "tag not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["resourceUrn"] != tt.urn {
					t.Errorf("resourceUrn = %v, want %v", input["resourceUrn"], tt.urn)
				}
				if input["tagUrn"] != tt.tagURN {
					t.Errorf("tagUrn = %v, want %v", input["tagUrn"], tt.tagURN)
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

			err := c.AddTag(context.Background(), tt.urn, tt.tagURN)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveTagGraphQL(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		tagURN   string
		response string
		wantErr  bool
	}{
		{
			name:     "domain entity",
			urn:      "urn:li:domain:engineering",
			tagURN:   "urn:li:tag:PII",
			response: `{"data": {"removeTag": true}}`,
		},
		{
			name:     "glossaryTerm entity",
			urn:      "urn:li:glossaryTerm:revenue",
			tagURN:   "urn:li:tag:Sensitive",
			response: `{"data": {"removeTag": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:domain:engineering",
			tagURN:   "urn:li:tag:PII",
			response: `{"errors": [{"message": "tag not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["resourceUrn"] != tt.urn {
					t.Errorf("resourceUrn = %v, want %v", input["resourceUrn"], tt.urn)
				}
				if input["tagUrn"] != tt.tagURN {
					t.Errorf("tagUrn = %v, want %v", input["tagUrn"], tt.tagURN)
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

			err := c.RemoveTag(context.Background(), tt.urn, tt.tagURN)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddTermGraphQL(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		termURN  string
		response string
		wantErr  bool
	}{
		{
			name:     "domain entity",
			urn:      "urn:li:domain:engineering",
			termURN:  "urn:li:glossaryTerm:revenue",
			response: `{"data": {"addTerm": true}}`,
		},
		{
			name:     "glossaryTerm entity",
			urn:      "urn:li:glossaryTerm:revenue",
			termURN:  "urn:li:glossaryTerm:finance",
			response: `{"data": {"addTerm": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:domain:engineering",
			termURN:  "urn:li:glossaryTerm:revenue",
			response: `{"errors": [{"message": "term not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["resourceUrn"] != tt.urn {
					t.Errorf("resourceUrn = %v, want %v", input["resourceUrn"], tt.urn)
				}
				if input["termUrn"] != tt.termURN {
					t.Errorf("termUrn = %v, want %v", input["termUrn"], tt.termURN)
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

			err := c.AddGlossaryTerm(context.Background(), tt.urn, tt.termURN)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveTermGraphQL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		input := extractGraphQLInput(t, r)
		if input["resourceUrn"] != "urn:li:domain:engineering" {
			t.Errorf("resourceUrn = %v, want urn:li:domain:engineering", input["resourceUrn"])
		}
		if input["termUrn"] != "urn:li:glossaryTerm:revenue" {
			t.Errorf("termUrn = %v, want urn:li:glossaryTerm:revenue", input["termUrn"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": {"removeTerm": true}}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.RemoveGlossaryTerm(context.Background(), "urn:li:domain:engineering", "urn:li:glossaryTerm:revenue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDescriptionGraphQL(t *testing.T) {
	tests := []struct {
		name        string
		urn         string
		description string
		response    string
		wantErr     bool
	}{
		{
			name:        "domain entity",
			urn:         "urn:li:domain:engineering",
			description: "Engineering domain",
			response:    `{"data": {"updateDescription": true}}`,
		},
		{
			name:        "glossaryTerm entity",
			urn:         "urn:li:glossaryTerm:revenue",
			description: "Revenue metric definition",
			response:    `{"data": {"updateDescription": true}}`,
		},
		{
			name:        "glossaryNode entity",
			urn:         "urn:li:glossaryNode:finance",
			description: "Finance glossary node",
			response:    `{"data": {"updateDescription": true}}`,
		},
		{
			name:        "graphql error",
			urn:         "urn:li:domain:engineering",
			description: "Updated",
			response:    `{"errors": [{"message": "entity not found"}]}`,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["resourceUrn"] != tt.urn {
					t.Errorf("resourceUrn = %v, want %v", input["resourceUrn"], tt.urn)
				}
				if input["description"] != tt.description {
					t.Errorf("description = %v, want %v", input["description"], tt.description)
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

			err := c.UpdateDescription(context.Background(), tt.urn, tt.description)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateDescription_DatasetStillUsesREST(t *testing.T) {
	// Verify that dataset entities still use the REST path, not GraphQL.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// REST GET for reading current aspect
			resp := aspectResponse{Value: json.RawMessage(`{"description":"old"}`)}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// REST POST — verify it's an ingestProposal, not a GraphQL mutation
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		// REST ingestProposal has "proposal" key; GraphQL has "query" key
		if _, hasQuery := body["query"]; hasQuery {
			t.Error("dataset description should use REST, not GraphQL")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.UpdateDescription(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"new description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddTag_DatasetStillUsesREST(t *testing.T) {
	// Verify that dataset entities still use the REST path for tags.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{Value: json.RawMessage(`{"tags":[]}`)}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if _, hasQuery := body["query"]; hasQuery {
			t.Error("dataset tag should use REST, not GraphQL")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddTag(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:tag:PII")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGraphQLWriteTypes(t *testing.T) {
	// Verify the graphQLWriteTypes map has the correct entries.
	expected := []string{"domain", "glossaryTerm", "glossaryNode"}
	for _, et := range expected {
		if !graphQLWriteTypes[et] {
			t.Errorf("expected %q in graphQLWriteTypes", et)
		}
	}
	// Verify dataset is NOT in the map.
	notExpected := []string{"dataset", "dashboard", "chart", "dataFlow", "dataJob", "container", "dataProduct"}
	for _, et := range notExpected {
		if graphQLWriteTypes[et] {
			t.Errorf("did not expect %q in graphQLWriteTypes", et)
		}
	}
}

func TestMutationStrings(t *testing.T) {
	// Verify mutation strings contain expected GraphQL patterns.
	tests := []struct {
		name     string
		mutation string
		contains string
	}{
		{"AddTag", AddTagMutation, "addTag(input: $input)"},
		{"RemoveTag", RemoveTagMutation, "removeTag(input: $input)"},
		{"AddTerm", AddTermMutation, "addTerm(input: $input)"},
		{"RemoveTerm", RemoveTermMutation, "removeTerm(input: $input)"},
		{"UpdateDescription", UpdateDescriptionMutation, "updateDescription(input: $input)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.mutation, tt.contains) {
				t.Errorf("mutation does not contain %q", tt.contains)
			}
		})
	}
}
