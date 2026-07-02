package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// customPropsFromProposal decodes the customProperties map written by an ingest proposal.
func customPropsFromProposal(t *testing.T, aspectJSON string) map[string]string {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(aspectJSON), &fields); err != nil {
		t.Fatalf("failed to unmarshal inner aspect: %v", err)
	}
	raw, ok := fields["customProperties"]
	if !ok {
		t.Fatal("customProperties field missing from written aspect")
	}
	var props map[string]string
	if err := json.Unmarshal(raw, &props); err != nil {
		t.Fatalf("failed to unmarshal customProperties: %v", err)
	}
	return props
}

func TestSetCustomProperties_MergesAndPreserves(t *testing.T) {
	var gotProps map[string]string
	var gotAspect, gotEntityType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Existing aspect carries a required field plus an existing custom property.
			propsJSON := `{"definition":"a term","termSource":"INTERNAL",` +
				`"customProperties":{"existing":"keep","source_system":"legacy"}}`
			resp := aspectResponse{Value: json.RawMessage(propsJSON)}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		proposal, aspectJSON := extractProposalWireFormat(t, r.Body)
		gotEntityType, _ = proposal["entityType"].(string)
		gotAspect, _ = proposal["aspectName"].(string)
		gotProps = customPropsFromProposal(t, aspectJSON)

		// Required field must survive the read-modify-write cycle.
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(aspectJSON), &fields); err != nil {
			t.Fatalf("failed to unmarshal aspect: %v", err)
		}
		if _, ok := fields["termSource"]; !ok {
			t.Error("termSource required field was not preserved")
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

	err := c.SetCustomProperties(context.Background(), "urn:li:glossaryTerm:revenue",
		map[string]string{"source_system": "warehouse", "owner": "data-team"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotEntityType != "glossaryTerm" {
		t.Errorf("entityType = %q, want glossaryTerm", gotEntityType)
	}
	if gotAspect != "glossaryTermInfo" {
		t.Errorf("aspectName = %q, want glossaryTermInfo", gotAspect)
	}
	want := map[string]string{"existing": "keep", "source_system": "warehouse", "owner": "data-team"}
	if len(gotProps) != len(want) {
		t.Fatalf("customProperties = %v, want %v", gotProps, want)
	}
	for k, v := range want {
		if gotProps[k] != v {
			t.Errorf("customProperties[%q] = %q, want %q", k, gotProps[k], v)
		}
	}
}

func TestRemoveCustomProperties_DeletesKeys(t *testing.T) {
	var gotProps map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			propsJSON := `{"name":"orders","customProperties":{"source_system":"legacy","keep":"me"}}`
			resp := aspectResponse{Value: json.RawMessage(propsJSON)}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		gotProps = customPropsFromProposal(t, aspectJSON)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.RemoveCustomProperties(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,db.orders,PROD)",
		[]string{"source_system", "absent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := gotProps["source_system"]; ok {
		t.Error("source_system should have been removed")
	}
	if gotProps["keep"] != "me" {
		t.Errorf("keep = %q, want me", gotProps["keep"])
	}
}

func TestSetCustomProperties_NoExistingAspect(t *testing.T) {
	var gotProps map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		gotProps = customPropsFromProposal(t, aspectJSON)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.SetCustomProperties(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,db.t,PROD)",
		map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotProps["a"] != "b" {
		t.Errorf("customProperties[a] = %q, want b", gotProps["a"])
	}
}

func TestCustomProperties_UnsupportedEntityType(t *testing.T) {
	c := &Client{logger: NopLogger{}}

	// tag has no customProperties field on tagProperties.
	if err := c.SetCustomProperties(context.Background(), "urn:li:tag:PII",
		map[string]string{"a": "b"}); err == nil {
		t.Fatal("expected error for tag entity")
	} else if !errors.Is(err, ErrUnsupportedCustomPropertiesEntity) {
		t.Errorf("expected ErrUnsupportedCustomPropertiesEntity, got: %v", err)
	}

	if err := c.RemoveCustomProperties(context.Background(), "urn:li:corpuser:jdoe",
		[]string{"a"}); err == nil {
		t.Fatal("expected error for corpuser entity")
	} else if !errors.Is(err, ErrUnsupportedCustomPropertiesEntity) {
		t.Errorf("expected ErrUnsupportedCustomPropertiesEntity, got: %v", err)
	}
}

func TestCustomProperties_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	if err := c.SetCustomProperties(context.Background(), "not-a-urn",
		map[string]string{"a": "b"}); err == nil {
		t.Fatal("expected error for invalid URN")
	}
}
