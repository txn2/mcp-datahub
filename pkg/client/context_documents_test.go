package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestGetContextDocuments(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns context documents",
			response: `{"data": {"entity": {"relatedDocuments": {
				"documents": [
					{
						"urn": "urn:li:document:ctx-1",
						"subType": "RUNBOOK",
						"info": {
							"title": "Runbook A",
							"contents": {"text": "Step 1"},
							"created": {"time": 1700000000000},
							"lastModified": {"time": 1700001000000}
						},
						"ownership": {"owners": [{"owner": {"urn": "urn:li:corpuser:alice", "username": "alice"}}]}
					},
					{
						"urn": "urn:li:document:ctx-2",
						"subType": "FAQ",
						"info": {
							"title": "FAQ B",
							"contents": {"text": "Q: Why?"},
							"created": {"time": 1700002000000},
							"lastModified": {"time": 1700003000000}
						},
						"ownership": {"owners": []}
					}
				]
			}}}}`,
			wantLen: 2,
		},
		{
			name:     "no related documents",
			response: `{"data": {"entity": {}}}`,
			wantLen:  0,
		},
		{
			name:     "graphql error",
			response: `{"errors": [{"message": "entity not found"}]}`,
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
				config:     DefaultConfig(),
			}

			docs, err := c.GetContextDocuments(context.Background(), "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(docs) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(docs), tt.wantLen)
			}

			if tt.wantLen == 0 {
				return
			}

			// Verify first document fields
			d := docs[0]
			if d.ID != "ctx-1" {
				t.Errorf("ID = %q, want ctx-1", d.ID)
			}
			if d.Title != "Runbook A" {
				t.Errorf("Title = %q, want Runbook A", d.Title)
			}
			if d.Content != "Step 1" {
				t.Errorf("Content = %q, want Step 1", d.Content)
			}
			if d.ContentType != "text/markdown" {
				t.Errorf("ContentType = %q, want text/markdown", d.ContentType)
			}
			if d.Category != "RUNBOOK" {
				t.Errorf("Category = %q, want RUNBOOK", d.Category)
			}
			if d.CreatedAt != 1700000000000 {
				t.Errorf("CreatedAt = %d, want 1700000000000", d.CreatedAt)
			}
			if d.UpdatedAt != 1700001000000 {
				t.Errorf("UpdatedAt = %d, want 1700001000000", d.UpdatedAt)
			}
			if d.Author == nil || d.Author.Username != "alice" {
				t.Errorf("Author = %+v, want alice", d.Author)
			}

			// Verify second document has no author
			if docs[1].Author != nil {
				t.Errorf("docs[1].Author = %+v, want nil", docs[1].Author)
			}
		})
	}
}

func TestUpsertContextDocument_Create(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		callCount.Add(1)

		switch {
		case strings.Contains(req.Query, "createDocument"):
			// Verify input
			input, _ := req.Variables["input"].(map[string]any)
			if input["title"] != "My Runbook" {
				t.Errorf("title = %v, want My Runbook", input["title"])
			}
			assets, _ := input["relatedAssets"].([]any)
			if len(assets) != 1 {
				t.Errorf("relatedAssets len = %d, want 1", len(assets))
			}
			_, _ = w.Write([]byte(`{"data": {"createDocument": "urn:li:document:new-1"}}`))

		case strings.Contains(req.Query, "getDocument"):
			if req.Variables["urn"] != "urn:li:document:new-1" {
				t.Errorf("getDocument urn = %v, want urn:li:document:new-1", req.Variables["urn"])
			}
			_, _ = w.Write([]byte(`{"data": {"document": {
				"urn": "urn:li:document:new-1",
				"subType": "RUNBOOK",
				"info": {
					"title": "My Runbook",
					"contents": {"text": "Step 1: do things"},
					"status": {"state": "PUBLISHED"},
					"created": {"time": 1700000000000},
					"lastModified": {"time": 1700000000000}
				},
				"ownership": {"owners": [{"owner": {"urn": "urn:li:corpuser:bot", "username": "bot"}, "type": "DATAOWNER"}]}
			}}}`))

		default:
			t.Fatalf("unexpected query: %s", req.Query)
		}
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	doc, err := c.UpsertContextDocument(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)",
		types.ContextDocumentInput{
			Title:    "My Runbook",
			Content:  "Step 1: do things",
			Category: "RUNBOOK",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.ID != "new-1" {
		t.Errorf("ID = %q, want new-1", doc.ID)
	}
	if doc.Title != "My Runbook" {
		t.Errorf("Title = %q, want My Runbook", doc.Title)
	}
	if doc.Author == nil || doc.Author.Username != "bot" {
		t.Errorf("Author = %+v, want bot", doc.Author)
	}
	if callCount.Load() != 2 {
		t.Errorf("callCount = %d, want 2 (create + get)", callCount.Load())
	}
}

func TestUpsertContextDocument_Update(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		callCount.Add(1)

		switch {
		case strings.Contains(req.Query, "updateDocumentContents"):
			input, _ := req.Variables["input"].(map[string]any)
			if input["urn"] != "urn:li:document:existing-1" {
				t.Errorf("urn = %v, want urn:li:document:existing-1", input["urn"])
			}
			_, _ = w.Write([]byte(`{"data": {"updateDocumentContents": true}}`))

		case strings.Contains(req.Query, "updateDocumentSubType"):
			input, _ := req.Variables["input"].(map[string]any)
			if input["subType"] != "FAQ" {
				t.Errorf("subType = %v, want FAQ", input["subType"])
			}
			_, _ = w.Write([]byte(`{"data": {"updateDocumentSubType": true}}`))

		case strings.Contains(req.Query, "getDocument"):
			_, _ = w.Write([]byte(`{"data": {"document": {
				"urn": "urn:li:document:existing-1",
				"subType": "FAQ",
				"info": {
					"title": "Updated FAQ",
					"contents": {"text": "New content"},
					"status": {"state": "PUBLISHED"},
					"created": {"time": 1700000000000},
					"lastModified": {"time": 1700005000000}
				}
			}}}`))

		default:
			t.Fatalf("unexpected query: %s", req.Query)
		}
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	doc, err := c.UpsertContextDocument(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)",
		types.ContextDocumentInput{
			ID:       "existing-1",
			Title:    "Updated FAQ",
			Content:  "New content",
			Category: "FAQ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.ID != "existing-1" {
		t.Errorf("ID = %q, want existing-1", doc.ID)
	}
	if doc.Category != "FAQ" {
		t.Errorf("Category = %q, want FAQ", doc.Category)
	}
	if callCount.Load() != 3 {
		t.Errorf("callCount = %d, want 3 (update contents + update subType + get)", callCount.Load())
	}
}

func TestUpsertContextDocument_UpdateWithoutCategory(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		callCount.Add(1)

		switch {
		case strings.Contains(req.Query, "updateDocumentContents"):
			_, _ = w.Write([]byte(`{"data": {"updateDocumentContents": true}}`))

		case strings.Contains(req.Query, "getDocument"):
			_, _ = w.Write([]byte(`{"data": {"document": {
				"urn": "urn:li:document:doc-1",
				"info": {
					"title": "Title",
					"contents": {"text": "Body"},
					"status": {"state": "PUBLISHED"},
					"created": {"time": 1700000000000},
					"lastModified": {"time": 1700000000000}
				}
			}}}`))

		default:
			t.Fatalf("unexpected query (should not call updateDocumentSubType): %s", req.Query)
		}
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	_, err := c.UpsertContextDocument(context.Background(), "urn:li:dataset:x",
		types.ContextDocumentInput{ID: "doc-1", Title: "Title", Content: "Body"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount.Load() != 2 {
		t.Errorf("callCount = %d, want 2 (update contents + get, no subType update)", callCount.Load())
	}
}

func TestUpsertContextDocument_MissingTitle(t *testing.T) {
	c := &Client{logger: NopLogger{}, config: DefaultConfig()}
	_, err := c.UpsertContextDocument(context.Background(), "urn:li:dataset:x",
		types.ContextDocumentInput{Content: "body only"},
	)
	if err == nil {
		t.Fatal("expected error for missing title")
	}
	if !strings.Contains(err.Error(), "title is required") {
		t.Errorf("error = %v, want 'title is required'", err)
	}
}

func TestUpsertContextDocument_CreateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors": [{"message": "permission denied"}]}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	_, err := c.UpsertContextDocument(context.Background(), "urn:li:dataset:x",
		types.ContextDocumentInput{Title: "Doc", Content: "content"},
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertContextDocument_UpdateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors": [{"message": "not found"}]}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	_, err := c.UpsertContextDocument(context.Background(), "urn:li:dataset:x",
		types.ContextDocumentInput{ID: "doc-1", Title: "Doc", Content: "content"},
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertContextDocument_UpdateSubTypeError(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		callCount.Add(1)

		switch {
		case strings.Contains(req.Query, "updateDocumentContents"):
			_, _ = w.Write([]byte(`{"data": {"updateDocumentContents": true}}`))
		case strings.Contains(req.Query, "updateDocumentSubType"):
			_, _ = w.Write([]byte(`{"errors": [{"message": "subType update failed"}]}`))
		default:
			t.Fatalf("unexpected query: %s", req.Query)
		}
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	_, err := c.UpsertContextDocument(context.Background(), "urn:li:dataset:x",
		types.ContextDocumentInput{ID: "doc-1", Title: "Title", Content: "Body", Category: "FAQ"},
	)
	if err == nil {
		t.Fatal("expected error from subType update failure")
	}
	if !strings.Contains(err.Error(), "update category") {
		t.Errorf("error = %v, want 'update category'", err)
	}
}

func TestUpsertContextDocument_FetchError(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		callCount.Add(1)

		switch {
		case strings.Contains(req.Query, "createDocument"):
			_, _ = w.Write([]byte(`{"data": {"createDocument": "urn:li:document:new-1"}}`))
		case strings.Contains(req.Query, "getDocument"):
			_, _ = w.Write([]byte(`{"errors": [{"message": "fetch failed"}]}`))
		default:
			t.Fatalf("unexpected query: %s", req.Query)
		}
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	_, err := c.UpsertContextDocument(context.Background(), "urn:li:dataset:x",
		types.ContextDocumentInput{Title: "Doc", Content: "content"},
	)
	if err == nil {
		t.Fatal("expected error from fetch failure")
	}
	if !strings.Contains(err.Error(), "fetchContextDocument") {
		t.Errorf("error = %v, want 'fetchContextDocument'", err)
	}
}

func TestDeleteContextDocument(t *testing.T) {
	tests := []struct {
		name       string
		documentID string
		response   string
		wantErr    bool
	}{
		{
			name:       "success",
			documentID: "doc-1",
			response:   `{"data": {"deleteDocument": true}}`,
		},
		{
			name:       "not found",
			documentID: "missing",
			response:   `{"errors": [{"message": "document not found"}]}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				expectedURN := "urn:li:document:" + tt.documentID
				if vars["urn"] != expectedURN {
					t.Errorf("urn = %v, want %v", vars["urn"], expectedURN)
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

			err := c.DeleteContextDocument(context.Background(), tt.documentID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToContextDocument(t *testing.T) {
	d := &contextDocResponse{
		URN:     "urn:li:document:test-id",
		SubType: "TUTORIAL",
	}
	d.Info.Title = "Test Doc"
	d.Info.Contents.Text = "Content here"
	d.Info.Created.Time = 1700000000000
	d.Info.LastModified.Time = 1700001000000

	result := toContextDocument(d)

	if result.ID != "test-id" {
		t.Errorf("ID = %q, want test-id", result.ID)
	}
	if result.ContentType != "text/markdown" {
		t.Errorf("ContentType = %q, want text/markdown", result.ContentType)
	}
	if result.Category != "TUTORIAL" {
		t.Errorf("Category = %q, want TUTORIAL", result.Category)
	}
	if result.Author != nil {
		t.Errorf("Author = %+v, want nil (no owners)", result.Author)
	}
}

func TestDocumentToContextDocument(t *testing.T) {
	d := &types.Document{
		URN:          "urn:li:document:full-test",
		Title:        "Full Doc",
		Content:      "Full content",
		SubType:      "REFERENCE",
		Created:      1700000000000,
		LastModified: 1700002000000,
		Owners: []types.Owner{
			{URN: "urn:li:corpuser:alice", Name: "Alice Smith"},
			{URN: "urn:li:corpuser:bob", Name: "Bob Jones"},
		},
	}

	result := documentToContextDocument(d)

	if result.ID != "full-test" {
		t.Errorf("ID = %q, want full-test", result.ID)
	}
	if result.Title != "Full Doc" {
		t.Errorf("Title = %q, want Full Doc", result.Title)
	}
	if result.Category != "REFERENCE" {
		t.Errorf("Category = %q, want REFERENCE", result.Category)
	}
	// Username is extracted from URN, not from Owner.Name (which may be a display name).
	if result.Author == nil || result.Author.Username != "alice" {
		t.Errorf("Author = %+v, want username alice (from URN)", result.Author)
	}
	if result.CreatedAt != 1700000000000 {
		t.Errorf("CreatedAt = %d, want 1700000000000", result.CreatedAt)
	}
}

func TestDocumentIDFromURN(t *testing.T) {
	tests := []struct {
		urn  string
		want string
	}{
		{"urn:li:document:abc-123", "abc-123"},
		{"urn:li:document:", ""},
		{"not-a-urn", "not-a-urn"},
	}

	for _, tt := range tests {
		got := documentIDFromURN(tt.urn)
		if got != tt.want {
			t.Errorf("documentIDFromURN(%q) = %q, want %q", tt.urn, got, tt.want)
		}
	}
}

func TestBuildDocumentURN(t *testing.T) {
	got := BuildDocumentURN("my-doc-id")
	want := "urn:li:document:my-doc-id"
	if got != want {
		t.Errorf("BuildDocumentURN = %q, want %q", got, want)
	}
}

func TestUsernameFromOwnerURN(t *testing.T) {
	tests := []struct {
		urn  string
		want string
	}{
		{"urn:li:corpuser:alice", "alice"},
		{"urn:li:corpuser:", ""},
		{"urn:li:corpGroup:engineering", "urn:li:corpGroup:engineering"},
	}

	for _, tt := range tests {
		got := usernameFromOwnerURN(tt.urn)
		if got != tt.want {
			t.Errorf("usernameFromOwnerURN(%q) = %q, want %q", tt.urn, got, tt.want)
		}
	}
}
