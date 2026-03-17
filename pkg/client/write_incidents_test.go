package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestUpdateIncident(t *testing.T) {
	tests := []struct {
		name       string
		urn        string
		input      types.UpdateIncidentInput
		response   string
		wantErr    bool
		checkInput func(t *testing.T, vars map[string]any)
	}{
		{
			name: "success with all fields",
			urn:  "urn:li:incident:inc1",
			input: types.UpdateIncidentInput{
				Title:       "Updated title",
				Description: "Updated desc",
				Priority:    "HIGH",
			},
			response: `{"data": {"updateIncident": true}}`,
			checkInput: func(t *testing.T, vars map[string]any) {
				t.Helper()
				if vars["urn"] != "urn:li:incident:inc1" {
					t.Errorf("urn = %v", vars["urn"])
				}
				input, _ := vars["input"].(map[string]any)
				if input["title"] != "Updated title" {
					t.Errorf("title = %v", input["title"])
				}
				if input["description"] != "Updated desc" {
					t.Errorf("description = %v", input["description"])
				}
				if input["priority"] != "HIGH" {
					t.Errorf("priority = %v", input["priority"])
				}
				// type and customType are not supported in UpdateIncidentInput
				if _, has := input["type"]; has {
					t.Error("type should not be sent in UpdateIncidentInput")
				}
				if _, has := input["customType"]; has {
					t.Error("customType should not be sent in UpdateIncidentInput")
				}
			},
		},
		{
			name: "success with partial fields omits empty",
			urn:  "urn:li:incident:inc2",
			input: types.UpdateIncidentInput{
				Title: "Title only",
			},
			response: `{"data": {"updateIncident": true}}`,
			checkInput: func(t *testing.T, vars map[string]any) {
				t.Helper()
				input, _ := vars["input"].(map[string]any)
				if input["title"] != "Title only" {
					t.Errorf("title = %v", input["title"])
				}
				if _, has := input["description"]; has {
					t.Error("expected no description when empty")
				}
				if _, has := input["priority"]; has {
					t.Error("expected no priority when empty")
				}
			},
		},
		{
			name: "graphql error",
			urn:  "urn:li:incident:inc1",
			input: types.UpdateIncidentInput{
				Title: "Fails",
			},
			response: `{"errors": [{"message": "incident not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if tt.checkInput != nil {
					tt.checkInput(t, vars)
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

			err := c.UpdateIncident(context.Background(), tt.urn, tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateIncidentStatus(t *testing.T) {
	tests := []struct {
		name       string
		urn        string
		state      string
		message    string
		response   string
		wantErr    bool
		checkInput func(t *testing.T, vars map[string]any)
	}{
		{
			name:     "resolve with message",
			urn:      "urn:li:incident:inc1",
			state:    "RESOLVED",
			message:  "Issue fixed",
			response: `{"data": {"updateIncidentStatus": true}}`,
			checkInput: func(t *testing.T, vars map[string]any) {
				t.Helper()
				if vars["urn"] != "urn:li:incident:inc1" {
					t.Errorf("urn = %v", vars["urn"])
				}
				input, _ := vars["input"].(map[string]any)
				if input["state"] != "RESOLVED" {
					t.Errorf("state = %v, want RESOLVED", input["state"])
				}
				if input["message"] != "Issue fixed" {
					t.Errorf("message = %v, want Issue fixed", input["message"])
				}
			},
		},
		{
			name:     "active state without message",
			urn:      "urn:li:incident:inc2",
			state:    "ACTIVE",
			message:  "",
			response: `{"data": {"updateIncidentStatus": true}}`,
			checkInput: func(t *testing.T, vars map[string]any) {
				t.Helper()
				input, _ := vars["input"].(map[string]any)
				if input["state"] != "ACTIVE" {
					t.Errorf("state = %v, want ACTIVE", input["state"])
				}
				if _, has := input["message"]; has {
					t.Error("expected no message when empty")
				}
			},
		},
		{
			name:     "empty state defaults to RESOLVED",
			urn:      "urn:li:incident:inc3",
			state:    "",
			message:  "",
			response: `{"data": {"updateIncidentStatus": true}}`,
			checkInput: func(t *testing.T, vars map[string]any) {
				t.Helper()
				input, _ := vars["input"].(map[string]any)
				if input["state"] != "RESOLVED" {
					t.Errorf("state = %v, want RESOLVED (default)", input["state"])
				}
			},
		},
		{
			name:     "graphql error",
			urn:      "urn:li:incident:inc1",
			state:    "RESOLVED",
			message:  "done",
			response: `{"errors": [{"message": "not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if tt.checkInput != nil {
					tt.checkInput(t, vars)
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

			err := c.UpdateIncidentStatus(context.Background(), tt.urn, tt.state, tt.message)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
