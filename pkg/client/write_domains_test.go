package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetDomain(t *testing.T) {
	tests := []struct {
		name      string
		entityURN string
		domainURN string
		response  string
		wantErr   bool
	}{
		{
			name:      "success",
			entityURN: "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			domainURN: "urn:li:domain:engineering",
			response:  `{"data": {"setDomain": true}}`,
		},
		{
			name:      "graphql error",
			entityURN: "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			domainURN: "urn:li:domain:missing",
			response:  `{"errors": [{"message": "domain not found"}]}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["entityUrn"] != tt.entityURN {
					t.Errorf("entityUrn = %v, want %v", vars["entityUrn"], tt.entityURN)
				}
				if vars["domainUrn"] != tt.domainURN {
					t.Errorf("domainUrn = %v, want %v", vars["domainUrn"], tt.domainURN)
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

			err := c.SetDomain(context.Background(), tt.entityURN, tt.domainURN)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnsetDomain(t *testing.T) {
	tests := []struct {
		name      string
		entityURN string
		response  string
		wantErr   bool
	}{
		{
			name:      "success",
			entityURN: "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			response:  `{"data": {"unsetDomain": true}}`,
		},
		{
			name:      "graphql error",
			entityURN: "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			response:  `{"errors": [{"message": "entity not found"}]}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["entityUrn"] != tt.entityURN {
					t.Errorf("entityUrn = %v, want %v", vars["entityUrn"], tt.entityURN)
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

			err := c.UnsetDomain(context.Background(), tt.entityURN)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
