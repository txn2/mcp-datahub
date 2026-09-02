package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newNotFoundTestClient serves a fixed GraphQL data payload for every request.
func newNotFoundTestClient(t *testing.T, data map[string]any) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"data": data})
	}))
	t.Cleanup(server.Close)

	client, err := New(Config{URL: server.URL, Token: "test-token", RetryMax: 0})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return client
}

const missingDatasetURN = "urn:li:dataset:(urn:li:dataPlatform:trino,warehouse.public.nope,PROD)"

// TestGetEntityKeyAspectStub covers the key-aspect stub DataHub returns for a
// dataset URN that was never ingested: the requested URN comes back with a name
// derived from it and every real aspect null, so the URN alone cannot tell it
// from a real entity. Only "exists": false marks it absent.
func TestGetEntityKeyAspectStub(t *testing.T) {
	tests := []struct {
		name     string
		entity   map[string]any
		wantErr  bool
		wantName string
	}{
		{
			name: "stub for missing dataset is not found",
			entity: map[string]any{
				"urn":    missingDatasetURN,
				"type":   "DATASET",
				"exists": false,
				"name":   "warehouse.public.nope",
			},
			wantErr: true,
		},
		{
			name: "existing dataset is returned",
			entity: map[string]any{
				"urn":    missingDatasetURN,
				"type":   "DATASET",
				"exists": true,
				"name":   "warehouse.public.real",
			},
			wantName: "warehouse.public.real",
		},
		{
			name: "omitted exists leaves the entity standing",
			entity: map[string]any{
				"urn":  missingDatasetURN,
				"type": "DATASET",
				"name": "warehouse.public.legacy",
			},
			wantName: "warehouse.public.legacy",
		},
		{
			name:    "empty urn is still not found",
			entity:  map[string]any{"urn": "", "type": "DATASET"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newNotFoundTestClient(t, map[string]any{"entity": tt.entity})

			entity, err := client.GetEntity(context.Background(), missingDatasetURN)
			if tt.wantErr {
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("GetEntity() error = %v, want ErrNotFound", err)
				}
				if entity != nil {
					t.Errorf("GetEntity() entity = %+v, want nil", entity)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetEntity() unexpected error: %v", err)
			}
			if entity.Name != tt.wantName {
				t.Errorf("GetEntity() Name = %q, want %q", entity.Name, tt.wantName)
			}
		})
	}
}

// TestGetGlossaryTermKeyAspectStub covers the stub for a glossary term URN that
// was never ingested. Its name is the URN's id segment, which is exactly what a
// real term's name usually is, so "exists" is the only reliable signal.
func TestGetGlossaryTermKeyAspectStub(t *testing.T) {
	const urn = "urn:li:glossaryTerm:Nonexistent"

	tests := []struct {
		name     string
		term     map[string]any
		wantErr  bool
		wantName string
	}{
		{
			name: "stub for missing term is not found",
			term: map[string]any{
				"urn":    urn,
				"exists": false,
				"name":   "Nonexistent",
			},
			wantErr: true,
		},
		{
			name: "existing term is returned",
			term: map[string]any{
				"urn":        urn,
				"exists":     true,
				"name":       "Nonexistent",
				"properties": map[string]any{"name": "PII", "description": "Personal data"},
			},
			wantName: "PII",
		},
		{
			name:     "omitted exists leaves the term standing",
			term:     map[string]any{"urn": urn, "name": "Legacy"},
			wantName: "Legacy",
		},
		{
			name:    "empty urn is still not found",
			term:    map[string]any{"urn": ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newNotFoundTestClient(t, map[string]any{"glossaryTerm": tt.term})

			term, err := client.GetGlossaryTerm(context.Background(), urn)
			if tt.wantErr {
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("GetGlossaryTerm() error = %v, want ErrNotFound", err)
				}
				if term != nil {
					t.Errorf("GetGlossaryTerm() term = %+v, want nil", term)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetGlossaryTerm() unexpected error: %v", err)
			}
			if term.Name != tt.wantName {
				t.Errorf("GetGlossaryTerm() Name = %q, want %q", term.Name, tt.wantName)
			}
		})
	}
}

// TestGetDataProductKeyAspectStub covers the stub for a data product URN that
// was never ingested. DataProduct exposes no "exists" field, so a null
// properties aspect is the absence signal: a product cannot be created without
// dataProductProperties.
func TestGetDataProductKeyAspectStub(t *testing.T) {
	const urn = "urn:li:dataProduct:nonexistent"

	tests := []struct {
		name     string
		product  map[string]any
		wantErr  bool
		wantName string
	}{
		{
			name:    "null properties is not found",
			product: map[string]any{"urn": urn, "properties": nil},
			wantErr: true,
		},
		{
			name:    "absent properties is not found",
			product: map[string]any{"urn": urn},
			wantErr: true,
		},
		{
			name: "existing product is returned",
			product: map[string]any{
				"urn":        urn,
				"properties": map[string]any{"name": "Customer 360", "description": "Curated"},
			},
			wantName: "Customer 360",
		},
		{
			name:    "empty urn is still not found",
			product: map[string]any{"urn": ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newNotFoundTestClient(t, map[string]any{"dataProduct": tt.product})

			product, err := client.GetDataProduct(context.Background(), urn)
			if tt.wantErr {
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("GetDataProduct() error = %v, want ErrNotFound", err)
				}
				if product != nil {
					t.Errorf("GetDataProduct() product = %+v, want nil", product)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetDataProduct() unexpected error: %v", err)
			}
			if product.Name != tt.wantName {
				t.Errorf("GetDataProduct() Name = %q, want %q", product.Name, tt.wantName)
			}
		})
	}
}
