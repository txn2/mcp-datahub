package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestUpdateStructuredProperty(t *testing.T) {
	tests := []struct {
		name       string
		urn        string
		input      types.UpdateStructuredPropertyInput
		response   string
		wantErr    bool
		checkInput func(t *testing.T, input map[string]any)
	}{
		{
			name: "success with all fields",
			urn:  "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
			input: types.UpdateStructuredPropertyInput{
				DisplayName: "Updated Retention",
				Description: "Updated description",
				NewAllowedValues: []types.AllowedValue{
					{Value: "180d", Description: "180 days"},
				},
			},
			response: `{"data": {"updateStructuredProperty": {"urn": "urn:li:structuredProperty:io.acryl.privacy.retentionTime"}}}`,
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["urn"] != "urn:li:structuredProperty:io.acryl.privacy.retentionTime" {
					t.Errorf("urn = %v", input["urn"])
				}
				if input["displayName"] != "Updated Retention" {
					t.Errorf("displayName = %v", input["displayName"])
				}
				if input["description"] != "Updated description" {
					t.Errorf("description = %v", input["description"])
				}
				avList, _ := input["newAllowedValues"].([]any)
				if len(avList) != 1 {
					t.Fatalf("newAllowedValues length = %d, want 1", len(avList))
				}
				av0, _ := avList[0].(map[string]any)
				valObj, _ := av0["value"].(map[string]any)
				if valObj["stringValue"] != "180d" {
					t.Errorf("newAllowedValues[0].value.stringValue = %v, want 180d", valObj["stringValue"])
				}
				if av0["description"] != "180 days" {
					t.Errorf("newAllowedValues[0].description = %v", av0["description"])
				}
			},
		},
		{
			name: "success with only display name",
			urn:  "urn:li:structuredProperty:prop1",
			input: types.UpdateStructuredPropertyInput{
				DisplayName: "New Name",
			},
			response: `{"data": {"updateStructuredProperty": {"urn": "urn:li:structuredProperty:prop1"}}}`,
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["displayName"] != "New Name" {
					t.Errorf("displayName = %v", input["displayName"])
				}
				if _, has := input["description"]; has {
					t.Error("expected no description when empty")
				}
				if _, has := input["newAllowedValues"]; has {
					t.Error("expected no newAllowedValues when empty")
				}
			},
		},
		{
			name: "graphql error",
			urn:  "urn:li:structuredProperty:prop1",
			input: types.UpdateStructuredPropertyInput{
				DisplayName: "Fails",
			},
			response: `{"errors": [{"message": "property not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if tt.checkInput != nil {
					tt.checkInput(t, input)
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

			err := c.UpdateStructuredProperty(context.Background(), tt.urn, tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
