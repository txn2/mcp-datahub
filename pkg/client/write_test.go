package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			propsJSON := `{"description":"old desc",` +
				`"created":{"actor":"urn:li:corpuser:admin","time":1000},` +
				`"lastModified":{"actor":"urn:li:corpuser:admin","time":2000}}`
			resp := aspectResponse{Value: json.RawMessage(propsJSON)}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		proposal, aspectJSON := extractProposalWireFormat(t, r.Body)

		if proposal["entityType"] != "dataset" {
			t.Errorf("expected entity type 'dataset', got %v", proposal["entityType"])
		}
		if proposal["aspectName"] != "editableDatasetProperties" {
			t.Errorf("expected aspect 'editableDatasetProperties', got %v", proposal["aspectName"])
		}

		var props editablePropertiesAspect
		if err := json.Unmarshal([]byte(aspectJSON), &props); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if props.Description != "new description" {
			t.Errorf("expected 'new description', got %q", props.Description)
		}
		if props.Created == nil || props.Created.Actor != "urn:li:corpuser:admin" {
			t.Error("expected created audit stamp to be preserved")
		}
		if props.LastModified == nil || props.LastModified.Time != 2000 {
			t.Error("expected lastModified audit stamp to be preserved")
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

func TestUpdateDescription_NoExistingProperties(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		var props editablePropertiesAspect
		if err := json.Unmarshal([]byte(aspectJSON), &props); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if props.Description != "first description" {
			t.Errorf("expected 'first description', got %q", props.Description)
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
		"first description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDescription_MarkdownContent(t *testing.T) {
	mdDesc := "## Overview\n\n**bold** and _italic_\n\n- item 1\n- item 2\n\n`code`"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			propsJSON := `{"description":"old",` +
				`"created":{"actor":"urn:li:corpuser:datahub","time":0},` +
				`"lastModified":{"actor":"urn:li:corpuser:datahub","time":0}}`
			resp := aspectResponse{Value: json.RawMessage(propsJSON)}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		var props editablePropertiesAspect
		if err := json.Unmarshal([]byte(aspectJSON), &props); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if props.Description != mdDesc {
			t.Errorf("markdown not preserved:\ngot:  %q\nwant: %q",
				props.Description, mdDesc)
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
		mdDesc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadEditableProperties_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := aspectResponse{Value: json.RawMessage(`not valid json`)}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.readEditableProperties(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestUpdateDescription_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.UpdateDescription(context.Background(), "not-a-urn", "desc")
	if err == nil {
		t.Fatal("expected error for invalid URN")
	}
}

func TestAddTag(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch {
		case r.Method == http.MethodGet:
			// Return existing tags
			resp := aspectResponse{
				Value: json.RawMessage(`{"tags":[{"tag":"urn:li:tag:existing"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)

		case r.Method == http.MethodPost:
			proposal, aspectJSON := extractProposalWireFormat(t, r.Body)

			if proposal["aspectName"] != "globalTags" {
				t.Errorf("expected aspect 'globalTags', got %v", proposal["aspectName"])
			}

			// Verify the tag was added to existing tags
			var tags globalTagsAspect
			if err := json.Unmarshal([]byte(aspectJSON), &tags); err != nil {
				t.Fatalf("failed to unmarshal inner aspect: %v", err)
			}
			if len(tags.Tags) != 2 {
				t.Errorf("expected 2 tags, got %d", len(tags.Tags))
			}

			w.WriteHeader(http.StatusOK)
		}
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
		"urn:li:tag:newtag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddTag_Duplicate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"tags":[{"tag":"urn:li:tag:existing"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// POST should not be called for duplicates
		t.Error("POST should not be called for duplicate tag")
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
		"urn:li:tag:existing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddTag_NoExistingTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// No existing tags
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// POST with the new tag
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		var tags globalTagsAspect
		if err := json.Unmarshal([]byte(aspectJSON), &tags); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if len(tags.Tags) != 1 {
			t.Errorf("expected 1 tag, got %d", len(tags.Tags))
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
		"urn:li:tag:newtag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"tags":[{"tag":"urn:li:tag:keep"},{"tag":"urn:li:tag:remove"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		var tags globalTagsAspect
		if err := json.Unmarshal([]byte(aspectJSON), &tags); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if len(tags.Tags) != 1 {
			t.Errorf("expected 1 tag after removal, got %d", len(tags.Tags))
		}
		if tags.Tags[0].Tag != "urn:li:tag:keep" {
			t.Errorf("expected remaining tag 'urn:li:tag:keep', got %q", tags.Tags[0].Tag)
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

	err := c.RemoveTag(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:tag:remove")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddGlossaryTerm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"terms":[{"urn":"urn:li:glossaryTerm:existing"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		proposal, aspectJSON := extractProposalWireFormat(t, r.Body)
		var terms glossaryTermsAspect
		if err := json.Unmarshal([]byte(aspectJSON), &terms); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if len(terms.Terms) != 2 {
			t.Errorf("expected 2 terms, got %d", len(terms.Terms))
		}
		if proposal["aspectName"] != "glossaryTerms" {
			t.Errorf("expected aspect 'glossaryTerms', got %v", proposal["aspectName"])
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

	err := c.AddGlossaryTerm(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:glossaryTerm:newterm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddGlossaryTerm_Duplicate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"terms":[{"urn":"urn:li:glossaryTerm:existing"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		t.Error("POST should not be called for duplicate term")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddGlossaryTerm(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:glossaryTerm:existing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveGlossaryTerm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"terms":[{"urn":"urn:li:glossaryTerm:keep"},{"urn":"urn:li:glossaryTerm:remove"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		var terms glossaryTermsAspect
		if err := json.Unmarshal([]byte(aspectJSON), &terms); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if len(terms.Terms) != 1 {
			t.Errorf("expected 1 term after removal, got %d", len(terms.Terms))
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

	err := c.RemoveGlossaryTerm(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:glossaryTerm:remove")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			linkJSON := `{"elements":[{"url":"https://existing.com",` +
				`"description":"existing",` +
				`"created":{"time":0,"actor":"urn:li:corpuser:datahub"}}]}`
			resp := aspectResponse{
				Value: json.RawMessage(linkJSON),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		proposal, aspectJSON := extractProposalWireFormat(t, r.Body)
		if proposal["aspectName"] != "institutionalMemory" {
			t.Errorf("expected aspect 'institutionalMemory', got %v", proposal["aspectName"])
		}
		var memory institutionalMemoryAspect
		if err := json.Unmarshal([]byte(aspectJSON), &memory); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if len(memory.Elements) != 2 {
			t.Errorf("expected 2 links, got %d", len(memory.Elements))
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

	err := c.AddLink(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"https://newlink.com", "New Link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddLink_Duplicate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			linkJSON := `{"elements":[{"url":"https://existing.com",` +
				`"description":"existing",` +
				`"created":{"time":0,"actor":"urn:li:corpuser:datahub"}}]}`
			resp := aspectResponse{
				Value: json.RawMessage(linkJSON),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		t.Error("POST should not be called for duplicate link")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddLink(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"https://existing.com", "Existing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			linkJSON := `{"elements":[` +
				`{"url":"https://keep.com","description":"keep",` +
				`"created":{"time":0,"actor":"urn:li:corpuser:datahub"}},` +
				`{"url":"https://remove.com","description":"remove",` +
				`"created":{"time":0,"actor":"urn:li:corpuser:datahub"}}]}`
			resp := aspectResponse{
				Value: json.RawMessage(linkJSON),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		var memory institutionalMemoryAspect
		if err := json.Unmarshal([]byte(aspectJSON), &memory); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if len(memory.Elements) != 1 {
			t.Errorf("expected 1 link after removal, got %d", len(memory.Elements))
		}
		if memory.Elements[0].URL != "https://keep.com" {
			t.Errorf("expected remaining link 'https://keep.com', got %q", memory.Elements[0].URL)
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

	err := c.RemoveLink(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"https://remove.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddTag_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.AddTag(context.Background(), "not-a-urn", "urn:li:tag:PII")
	if err == nil {
		t.Fatal("expected error for invalid URN")
	}
}

func TestRemoveTag_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.RemoveTag(context.Background(), "not-a-urn", "urn:li:tag:PII")
	if err == nil {
		t.Fatal("expected error for invalid URN")
	}
}

func TestAddTag_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
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
		"urn:li:tag:newtag")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestRemoveTag_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.RemoveTag(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:tag:remove")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestReadGlobalTags_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := aspectResponse{
			Value: json.RawMessage(`not valid json`),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.readGlobalTags(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAddGlossaryTerm_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.AddGlossaryTerm(context.Background(), "not-a-urn", "urn:li:glossaryTerm:Term")
	if err == nil {
		t.Fatal("expected error for invalid URN")
	}
}

func TestRemoveGlossaryTerm_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.RemoveGlossaryTerm(context.Background(), "not-a-urn", "urn:li:glossaryTerm:Term")
	if err == nil {
		t.Fatal("expected error for invalid URN")
	}
}

func TestAddGlossaryTerm_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddGlossaryTerm(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:glossaryTerm:newterm")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestRemoveGlossaryTerm_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.RemoveGlossaryTerm(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:glossaryTerm:remove")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestReadGlossaryTerms_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := aspectResponse{
			Value: json.RawMessage(`not valid json`),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.readGlossaryTerms(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAddLink_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.AddLink(context.Background(), "not-a-urn", "https://test.com", "desc")
	if err == nil {
		t.Fatal("expected error for invalid URN")
	}
}

func TestRemoveLink_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.RemoveLink(context.Background(), "not-a-urn", "https://test.com")
	if err == nil {
		t.Fatal("expected error for invalid URN")
	}
}

func TestAddLink_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddLink(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"https://newlink.com", "desc")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestRemoveLink_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.RemoveLink(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"https://remove.com")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestReadInstitutionalMemory_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := aspectResponse{
			Value: json.RawMessage(`not valid json`),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.readInstitutionalMemory(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAddGlossaryTerm_NoExistingTerms(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		var terms glossaryTermsAspect
		if err := json.Unmarshal([]byte(aspectJSON), &terms); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if len(terms.Terms) != 1 {
			t.Errorf("expected 1 term, got %d", len(terms.Terms))
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

	err := c.AddGlossaryTerm(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:glossaryTerm:newterm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddLink_NoExistingLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, aspectJSON := extractProposalWireFormat(t, r.Body)
		var memory institutionalMemoryAspect
		if err := json.Unmarshal([]byte(aspectJSON), &memory); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if len(memory.Elements) != 1 {
			t.Errorf("expected 1 link, got %d", len(memory.Elements))
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

	err := c.AddLink(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"https://newlink.com", "New Link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDescription_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`server error`))
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
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestUpdateColumnDescription(t *testing.T) {
	existingAspect := `{"editableSchemaFieldInfo":[` +
		`{"fieldPath":"id","description":"existing id desc",` +
		`"globalTags":{"tags":[{"tag":"urn:li:tag:PII"}]}},` +
		`{"fieldPath":"name","description":"existing name desc"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{Value: json.RawMessage(existingAspect)}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		proposal, aspectJSON := extractProposalWireFormat(t, r.Body)

		if proposal["aspectName"] != "editableSchemaMetadata" {
			t.Errorf("expected aspect 'editableSchemaMetadata', got %v", proposal["aspectName"])
		}

		var schema editableSchemaAspect
		if err := json.Unmarshal([]byte(aspectJSON), &schema); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}

		// Should still have 2 fields
		if len(schema.EditableSchemaFieldInfo) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(schema.EditableSchemaFieldInfo))
		}

		// "id" field should have updated description and preserved tags
		idField := schema.EditableSchemaFieldInfo[0]
		if idField.FieldPath != "id" {
			t.Errorf("expected fieldPath 'id', got %q", idField.FieldPath)
		}
		if idField.Description != "updated id desc" {
			t.Errorf("expected 'updated id desc', got %q", idField.Description)
		}
		if idField.GlobalTags == nil {
			t.Error("expected globalTags to be preserved")
		}

		// "name" field should be untouched
		nameField := schema.EditableSchemaFieldInfo[1]
		if nameField.Description != "existing name desc" {
			t.Errorf("expected 'existing name desc', got %q", nameField.Description)
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

	err := c.UpdateColumnDescription(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"id", "updated id desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateColumnDescription_NewField(t *testing.T) {
	existingAspect := `{"editableSchemaFieldInfo":[` +
		`{"fieldPath":"id","description":"id desc"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{Value: json.RawMessage(existingAspect)}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		_, aspectJSON := extractProposalWireFormat(t, r.Body)

		var schema editableSchemaAspect
		if err := json.Unmarshal([]byte(aspectJSON), &schema); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}

		// Should now have 2 fields (existing + new)
		if len(schema.EditableSchemaFieldInfo) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(schema.EditableSchemaFieldInfo))
		}

		newField := schema.EditableSchemaFieldInfo[1]
		if newField.FieldPath != "new_column" {
			t.Errorf("expected fieldPath 'new_column', got %q", newField.FieldPath)
		}
		if newField.Description != "new column desc" {
			t.Errorf("expected 'new column desc', got %q", newField.Description)
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

	err := c.UpdateColumnDescription(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"new_column", "new column desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateColumnDescription_NoExistingAspect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, aspectJSON := extractProposalWireFormat(t, r.Body)

		var schema editableSchemaAspect
		if err := json.Unmarshal([]byte(aspectJSON), &schema); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}

		if len(schema.EditableSchemaFieldInfo) != 1 {
			t.Fatalf("expected 1 field, got %d", len(schema.EditableSchemaFieldInfo))
		}
		if schema.EditableSchemaFieldInfo[0].FieldPath != "email" {
			t.Errorf("expected fieldPath 'email', got %q", schema.EditableSchemaFieldInfo[0].FieldPath)
		}
		if schema.EditableSchemaFieldInfo[0].Description != "Email address" {
			t.Errorf("expected 'Email address', got %q", schema.EditableSchemaFieldInfo[0].Description)
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

	err := c.UpdateColumnDescription(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"email", "Email address")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateColumnDescription_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.UpdateColumnDescription(context.Background(), "not-a-urn", "col", "desc")
	if err == nil {
		t.Fatal("expected error for invalid URN")
	}
}

func TestUpdateColumnDescription_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`server error`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.UpdateColumnDescription(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"col", "desc")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestReadEditableSchema_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := aspectResponse{Value: json.RawMessage(`not valid json`)}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.readEditableSchema(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestUpdateColumnDescription_NullAspectValue(t *testing.T) {
	// Tests the P0 fix: DataHub returns 200 with null value
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"value":null}`))
			return
		}

		_, aspectJSON := extractProposalWireFormat(t, r.Body)

		var schema editableSchemaAspect
		if err := json.Unmarshal([]byte(aspectJSON), &schema); err != nil {
			t.Fatalf("failed to unmarshal inner aspect: %v", err)
		}
		if len(schema.EditableSchemaFieldInfo) != 1 {
			t.Fatalf("expected 1 field, got %d", len(schema.EditableSchemaFieldInfo))
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

	err := c.UpdateColumnDescription(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"col", "new desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEntityTypeFromURN(t *testing.T) {
	tests := []struct {
		name    string
		urn     string
		want    string
		wantErr bool
	}{
		{"dataset", "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)", "dataset", false},
		{"glossaryTerm", "urn:li:glossaryTerm:Classification", "glossaryTerm", false},
		{"tag", "urn:li:tag:PII", "tag", false},
		{"dashboard", "urn:li:dashboard:(looker,dashboards.123)", "dashboard", false},
		{"invalid", "not-a-urn", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := entityTypeFromURN(tt.urn)
			if (err != nil) != tt.wantErr {
				t.Errorf("entityTypeFromURN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("entityTypeFromURN() = %q, want %q", got, tt.want)
			}
		})
	}
}
