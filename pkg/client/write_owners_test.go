package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddOwner(t *testing.T) {
	tests := []struct {
		name          string
		urn           string
		ownerURN      string
		ownershipType string
		response      string
		wantErr       bool
		checkInput    func(t *testing.T, input map[string]any)
	}{
		{
			name:          "corp user with explicit ownership type",
			urn:           "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			ownerURN:      "urn:li:corpuser:johndoe",
			ownershipType: "DATA_STEWARD",
			response:      `{"data": {"addOwner": true}}`,
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["ownerUrn"] != "urn:li:corpuser:johndoe" {
					t.Errorf("ownerUrn = %v", input["ownerUrn"])
				}
				if input["ownerEntityType"] != "CORP_USER" {
					t.Errorf("ownerEntityType = %v, want CORP_USER", input["ownerEntityType"])
				}
				if input["ownershipTypeUrn"] != "urn:li:ownershipType:DATA_STEWARD" {
					t.Errorf("ownershipTypeUrn = %v", input["ownershipTypeUrn"])
				}
				if input["resourceUrn"] != "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)" {
					t.Errorf("resourceUrn = %v", input["resourceUrn"])
				}
			},
		},
		{
			name:          "corp group detected from URN prefix",
			urn:           "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			ownerURN:      "urn:li:corpGroup:data-team",
			ownershipType: "TECHNICAL_OWNER",
			response:      `{"data": {"addOwner": true}}`,
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["ownerEntityType"] != "CORP_GROUP" {
					t.Errorf("ownerEntityType = %v, want CORP_GROUP", input["ownerEntityType"])
				}
			},
		},
		{
			name:          "empty ownership type defaults to TECHNICAL_OWNER",
			urn:           "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			ownerURN:      "urn:li:corpuser:johndoe",
			ownershipType: "",
			response:      `{"data": {"addOwner": true}}`,
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["ownershipTypeUrn"] != "urn:li:ownershipType:TECHNICAL_OWNER" {
					t.Errorf("ownershipTypeUrn = %v, want urn:li:ownershipType:TECHNICAL_OWNER", input["ownershipTypeUrn"])
				}
			},
		},
		{
			name:          "double-prefix guard",
			urn:           "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			ownerURN:      "urn:li:corpuser:johndoe",
			ownershipType: "urn:li:ownershipType:BUSINESS_OWNER",
			response:      `{"data": {"addOwner": true}}`,
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				// Should not double-prefix
				if input["ownershipTypeUrn"] != "urn:li:ownershipType:BUSINESS_OWNER" {
					t.Errorf("ownershipTypeUrn = %v, want urn:li:ownershipType:BUSINESS_OWNER", input["ownershipTypeUrn"])
				}
			},
		},
		{
			name:          "graphql error",
			urn:           "urn:li:dataset:test",
			ownerURN:      "urn:li:corpuser:johndoe",
			ownershipType: "TECHNICAL_OWNER",
			response:      `{"errors": [{"message": "entity not found"}]}`,
			wantErr:       true,
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

			err := c.AddOwner(context.Background(), tt.urn, tt.ownerURN, tt.ownershipType)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveOwner(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		ownerURN string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			ownerURN: "urn:li:corpuser:johndoe",
			response: `{"data": {"removeOwner": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:dataset:test",
			ownerURN: "urn:li:corpuser:johndoe",
			response: `{"errors": [{"message": "entity not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["ownerUrn"] != tt.ownerURN {
					t.Errorf("ownerUrn = %v, want %v", input["ownerUrn"], tt.ownerURN)
				}
				if input["resourceUrn"] != tt.urn {
					t.Errorf("resourceUrn = %v, want %v", input["resourceUrn"], tt.urn)
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

			err := c.RemoveOwner(context.Background(), tt.urn, tt.ownerURN)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
