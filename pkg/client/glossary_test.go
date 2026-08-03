package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// glossaryTestServer returns a client wired to a server that captures the
// GraphQL variables of the last request and replies with a fixed body.
func glossaryTestServer(t *testing.T, response string, capture *map[string]any) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode GraphQL request: %v", err)
		}
		if capture != nil {
			*capture = req.Variables
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	return &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		config:     Config{DefaultLimit: 10, MaxLimit: 100},
		logger:     NopLogger{},
	}
}

// pagingInput pulls the start/count the client sent inside the "input" variable.
func pagingInput(t *testing.T, vars map[string]any) (start, count float64) {
	t.Helper()
	input, ok := vars["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'input' variable to be a map, got %T", vars["input"])
	}
	start, ok = input["start"].(float64)
	if !ok {
		t.Fatalf("expected 'start' to be a number, got %T", input["start"])
	}
	count, ok = input["count"].(float64)
	if !ok {
		t.Fatalf("expected 'count' to be a number, got %T", input["count"])
	}
	return start, count
}

func TestGetRootGlossaryNodes(t *testing.T) {
	const twoNodes = `{"data": {"getRootGlossaryNodes": {
		"start": 0, "count": 2, "total": 5,
		"nodes": [
			{"urn": "urn:li:glossaryNode:finance",
			 "properties": {"name": "Finance", "description": "Finance terms"},
			 "childrenCount": {"termsCount": 3, "nodesCount": 1}},
			{"urn": "urn:li:glossaryNode:legal",
			 "properties": {"name": "Legal", "description": ""},
			 "childrenCount": {"termsCount": 0, "nodesCount": 0}}
		]}}}`

	tests := []struct {
		name      string
		start     int
		count     int
		response  string
		wantStart float64
		wantCount float64
		wantLen   int
		wantTotal int
		wantErr   bool
	}{
		{
			name:      "returns page and total",
			start:     0,
			count:     2,
			response:  twoNodes,
			wantStart: 0,
			wantCount: 2,
			wantLen:   2,
			wantTotal: 5,
		},
		{
			name:      "non-positive count falls back to default limit",
			start:     -5,
			count:     0,
			response:  twoNodes,
			wantStart: 0,
			wantCount: 10,
			wantLen:   2,
			wantTotal: 5,
		},
		{
			name:      "count above max is clamped",
			start:     20,
			count:     5000,
			response:  twoNodes,
			wantStart: 20,
			wantCount: 100,
			wantLen:   2,
			wantTotal: 5,
		},
		{
			name:     "empty glossary",
			count:    10,
			response: `{"data": {"getRootGlossaryNodes": {"start": 0, "count": 0, "total": 0, "nodes": []}}}`,
		},
		{
			name:     "graphql error",
			count:    10,
			response: `{"errors": [{"message": "boom"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var vars map[string]any
			c := glossaryTestServer(t, tt.response, &vars)

			nodes, total, err := c.GetRootGlossaryNodes(context.Background(), tt.start, tt.count)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(nodes) != tt.wantLen {
				t.Fatalf("len(nodes) = %d, want %d", len(nodes), tt.wantLen)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if tt.wantCount != 0 {
				gotStart, gotCount := pagingInput(t, vars)
				if gotStart != tt.wantStart || gotCount != tt.wantCount {
					t.Errorf("paging = (%v, %v), want (%v, %v)", gotStart, gotCount, tt.wantStart, tt.wantCount)
				}
			}
		})
	}
}

func TestGetRootGlossaryNodesFields(t *testing.T) {
	response := `{"data": {"getRootGlossaryNodes": {
		"start": 0, "count": 1, "total": 1,
		"nodes": [
			{"urn": "urn:li:glossaryNode:finance",
			 "properties": {"name": "Finance", "description": "Finance terms"},
			 "childrenCount": {"termsCount": 3, "nodesCount": 1}}
		]}}}`

	c := glossaryTestServer(t, response, nil)

	nodes, _, err := c.GetRootGlossaryNodes(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("GetRootGlossaryNodes: %v", err)
	}

	got := nodes[0]
	if got.URN != "urn:li:glossaryNode:finance" {
		t.Errorf("URN = %q, want urn:li:glossaryNode:finance", got.URN)
	}
	if got.Name != "Finance" {
		t.Errorf("Name = %q, want Finance", got.Name)
	}
	if got.Description != "Finance terms" {
		t.Errorf("Description = %q, want %q", got.Description, "Finance terms")
	}
	if got.ParentNode != "" {
		t.Errorf("ParentNode = %q, want empty for a root node", got.ParentNode)
	}
	if got.TermsCount != 3 || got.NodesCount != 1 {
		t.Errorf("child counts = (%d, %d), want (3, 1)", got.TermsCount, got.NodesCount)
	}
}

func TestGetRootGlossaryTerms(t *testing.T) {
	tests := []struct {
		name      string
		response  string
		wantLen   int
		wantTotal int
		wantName  string
		wantErr   bool
	}{
		{
			name: "prefers properties name over urn-derived name",
			response: `{"data": {"getRootGlossaryTerms": {
				"start": 0, "count": 1, "total": 4,
				"terms": [
					{"urn": "urn:li:glossaryTerm:revenue", "name": "revenue",
					 "properties": {"name": "Revenue", "description": "Total revenue"}}
				]}}}`,
			wantLen:   1,
			wantTotal: 4,
			wantName:  "Revenue",
		},
		{
			name: "falls back to entity name when properties are absent",
			response: `{"data": {"getRootGlossaryTerms": {
				"start": 0, "count": 1, "total": 1,
				"terms": [{"urn": "urn:li:glossaryTerm:revenue", "name": "revenue"}]}}}`,
			wantLen:   1,
			wantTotal: 1,
			wantName:  "revenue",
		},
		{
			name:     "empty glossary",
			response: `{"data": {"getRootGlossaryTerms": {"start": 0, "count": 0, "total": 0, "terms": []}}}`,
		},
		{
			name:     "graphql error",
			response: `{"errors": [{"message": "boom"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := glossaryTestServer(t, tt.response, nil)

			terms, total, err := c.GetRootGlossaryTerms(context.Background(), 0, 10)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(terms) != tt.wantLen {
				t.Fatalf("len(terms) = %d, want %d", len(terms), tt.wantLen)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if tt.wantLen == 0 {
				return
			}
			if terms[0].Name != tt.wantName {
				t.Errorf("Name = %q, want %q", terms[0].Name, tt.wantName)
			}
			if terms[0].ParentNode != "" {
				t.Errorf("ParentNode = %q, want empty for a root term", terms[0].ParentNode)
			}
		})
	}
}

// mixedChildren is a children page holding one node and one term, the shape
// DataHub returns for the IsPartOf relationship (verified against v1.6.0).
const mixedChildren = `{"data": {"glossaryNode": {
	"urn": "urn:li:glossaryNode:finance",
	"exists": true,
	"children": {
		"start": 0, "count": 2, "total": 7,
		"relationships": [
			{"entity": {"urn": "urn:li:glossaryNode:revenue", "type": "GLOSSARY_NODE",
			 "properties": {"name": "Revenue", "description": "Revenue sub-glossary"},
			 "childrenCount": {"termsCount": 2, "nodesCount": 0}}},
			{"entity": {"urn": "urn:li:glossaryTerm:arr", "type": "GLOSSARY_TERM", "name": "arr",
			 "properties": {"name": "ARR", "description": "Annual recurring revenue"}}}
		]}}}}`

func TestGetGlossaryNodeChildren(t *testing.T) {
	var vars map[string]any
	c := glossaryTestServer(t, mixedChildren, &vars)

	children, err := c.GetGlossaryNodeChildren(context.Background(), "urn:li:glossaryNode:finance", 0, 2)
	if err != nil {
		t.Fatalf("GetGlossaryNodeChildren: %v", err)
	}

	if children.Start != 0 || children.Count != 2 || children.Total != 7 {
		t.Errorf("paging = (%d, %d, %d), want (0, 2, 7)", children.Start, children.Count, children.Total)
	}
	if len(children.Nodes) != 1 || len(children.Terms) != 1 {
		t.Fatalf("split = (%d nodes, %d terms), want (1, 1)", len(children.Nodes), len(children.Terms))
	}

	node := children.Nodes[0]
	if node.URN != "urn:li:glossaryNode:revenue" || node.Name != "Revenue" {
		t.Errorf("node = (%q, %q), want (urn:li:glossaryNode:revenue, Revenue)", node.URN, node.Name)
	}
	if node.ParentNode != "urn:li:glossaryNode:finance" {
		t.Errorf("node ParentNode = %q, want urn:li:glossaryNode:finance", node.ParentNode)
	}
	if node.TermsCount != 2 || node.NodesCount != 0 {
		t.Errorf("node child counts = (%d, %d), want (2, 0)", node.TermsCount, node.NodesCount)
	}

	term := children.Terms[0]
	if term.URN != "urn:li:glossaryTerm:arr" || term.Name != "ARR" {
		t.Errorf("term = (%q, %q), want (urn:li:glossaryTerm:arr, ARR)", term.URN, term.Name)
	}
	if term.Description != "Annual recurring revenue" {
		t.Errorf("term Description = %q, want %q", term.Description, "Annual recurring revenue")
	}
	if term.ParentNode != "urn:li:glossaryNode:finance" {
		t.Errorf("term ParentNode = %q, want urn:li:glossaryNode:finance", term.ParentNode)
	}
}

func TestGetGlossaryNodeChildrenRelationshipInput(t *testing.T) {
	var vars map[string]any
	c := glossaryTestServer(t, mixedChildren, &vars)

	if _, err := c.GetGlossaryNodeChildren(context.Background(), "urn:li:glossaryNode:finance", 5, 3); err != nil {
		t.Fatalf("GetGlossaryNodeChildren: %v", err)
	}

	if vars["urn"] != "urn:li:glossaryNode:finance" {
		t.Errorf("urn variable = %v, want urn:li:glossaryNode:finance", vars["urn"])
	}

	input, ok := vars["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'input' variable to be a map, got %T", vars["input"])
	}
	if input["direction"] != "INCOMING" {
		t.Errorf("direction = %v, want INCOMING", input["direction"])
	}
	relTypes, ok := input["types"].([]any)
	if !ok || len(relTypes) != 1 || relTypes[0] != "IsPartOf" {
		t.Errorf("types = %v, want [IsPartOf]", input["types"])
	}
	start, count := pagingInput(t, vars)
	if start != 5 || count != 3 {
		t.Errorf("paging = (%v, %v), want (5, 3)", start, count)
	}
}

func TestGetGlossaryNodeChildrenErrors(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		response string
		wantErr  error
	}{
		{
			name:    "empty urn",
			urn:     "",
			wantErr: ErrInvalidURN,
		},
		{
			name:    "malformed urn",
			urn:     "not-a-urn",
			wantErr: ErrInvalidURN,
		},
		{
			// A term has no children, so the caller passed the wrong entity.
			name:    "glossary term urn",
			urn:     "urn:li:glossaryTerm:arr",
			wantErr: ErrInvalidURN,
		},
		{
			name:    "dataset urn",
			urn:     "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
			wantErr: ErrInvalidURN,
		},
		{
			name: "node does not exist",
			urn:  "urn:li:glossaryNode:missing",
			response: `{"data": {"glossaryNode": {"urn": "urn:li:glossaryNode:missing", "exists": false,
				"children": {"start": 0, "count": 0, "total": 0, "relationships": []}}}}`,
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := glossaryTestServer(t, tt.response, nil)

			_, err := c.GetGlossaryNodeChildren(context.Background(), tt.urn, 0, 10)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetGlossaryNodeChildrenWithoutExistsField(t *testing.T) {
	// DataHub versions that do not return "exists" must not be read as absent.
	response := `{"data": {"glossaryNode": {"urn": "urn:li:glossaryNode:finance",
		"children": {"start": 0, "count": 0, "total": 0, "relationships": []}}}}`
	c := glossaryTestServer(t, response, nil)

	children, err := c.GetGlossaryNodeChildren(context.Background(), "urn:li:glossaryNode:finance", 0, 10)
	if err != nil {
		t.Fatalf("GetGlossaryNodeChildren: %v", err)
	}
	if children.Total != 0 || len(children.Nodes) != 0 || len(children.Terms) != 0 {
		t.Errorf("children = %+v, want an empty page", children)
	}
}

func TestGetGlossaryNodeChildrenGraphQLError(t *testing.T) {
	c := glossaryTestServer(t, `{"errors": [{"message": "boom"}]}`, nil)

	if _, err := c.GetGlossaryNodeChildren(context.Background(), "urn:li:glossaryNode:finance", 0, 10); err == nil {
		t.Fatal("expected an error for a GraphQL error response")
	}
}

func TestGetGlossaryParentChain(t *testing.T) {
	// DataHub returns the chain direct-parent first (verified against v1.6.0).
	response := `{"data": {"entity": {
		"urn": "urn:li:glossaryTerm:arr",
		"type": "GLOSSARY_TERM",
		"parentNodes": {"count": 2, "nodes": [
			{"urn": "urn:li:glossaryNode:revenue",
			 "properties": {"name": "Revenue", "description": "Revenue sub-glossary"},
			 "childrenCount": {"termsCount": 2, "nodesCount": 0}},
			{"urn": "urn:li:glossaryNode:finance",
			 "properties": {"name": "Finance", "description": "Finance terms"},
			 "childrenCount": {"termsCount": 0, "nodesCount": 1}}
		]}}}}`

	c := glossaryTestServer(t, response, nil)

	chain, err := c.GetGlossaryParentChain(context.Background(), "urn:li:glossaryTerm:arr")
	if err != nil {
		t.Fatalf("GetGlossaryParentChain: %v", err)
	}

	if len(chain) != 2 {
		t.Fatalf("len(chain) = %d, want 2", len(chain))
	}
	if chain[0].URN != "urn:li:glossaryNode:revenue" {
		t.Errorf("chain[0] = %q, want the direct parent urn:li:glossaryNode:revenue", chain[0].URN)
	}
	if chain[1].URN != "urn:li:glossaryNode:finance" {
		t.Errorf("chain[1] = %q, want the root urn:li:glossaryNode:finance", chain[1].URN)
	}
	// Each entry points at its own parent, the next link up the chain.
	if chain[0].ParentNode != "urn:li:glossaryNode:finance" {
		t.Errorf("chain[0].ParentNode = %q, want urn:li:glossaryNode:finance", chain[0].ParentNode)
	}
	if chain[1].ParentNode != "" {
		t.Errorf("chain[1].ParentNode = %q, want empty for the root", chain[1].ParentNode)
	}
	if chain[0].Name != "Revenue" || chain[0].TermsCount != 2 {
		t.Errorf("chain[0] = (%q, %d terms), want (Revenue, 2 terms)", chain[0].Name, chain[0].TermsCount)
	}
}

func TestGetGlossaryParentChainNodeURN(t *testing.T) {
	response := `{"data": {"entity": {
		"urn": "urn:li:glossaryNode:revenue",
		"type": "GLOSSARY_NODE",
		"parentNodes": {"count": 1, "nodes": [
			{"urn": "urn:li:glossaryNode:finance", "properties": {"name": "Finance"}}
		]}}}}`

	c := glossaryTestServer(t, response, nil)

	chain, err := c.GetGlossaryParentChain(context.Background(), "urn:li:glossaryNode:revenue")
	if err != nil {
		t.Fatalf("GetGlossaryParentChain: %v", err)
	}
	if len(chain) != 1 || chain[0].URN != "urn:li:glossaryNode:finance" {
		t.Fatalf("chain = %+v, want the single parent urn:li:glossaryNode:finance", chain)
	}
}

func TestGetGlossaryParentChainRoot(t *testing.T) {
	response := `{"data": {"entity": {"urn": "urn:li:glossaryNode:finance", "type": "GLOSSARY_NODE",
		"parentNodes": {"count": 0, "nodes": []}}}}`

	c := glossaryTestServer(t, response, nil)

	chain, err := c.GetGlossaryParentChain(context.Background(), "urn:li:glossaryNode:finance")
	if err != nil {
		t.Fatalf("GetGlossaryParentChain: %v", err)
	}
	if len(chain) != 0 {
		t.Errorf("chain = %+v, want empty for a root node", chain)
	}
}

func TestGetGlossaryParentChainRejectsNonGlossaryURN(t *testing.T) {
	tests := []struct {
		name string
		urn  string
	}{
		{name: "empty urn", urn: ""},
		{name: "malformed urn", urn: "not-a-urn"},
		{name: "tag urn", urn: "urn:li:tag:PII"},
		{name: "dataset urn", urn: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A reply is configured so a failure to reject would surface as a
			// passing call rather than a transport error.
			c := glossaryTestServer(t, `{"data": {"entity": {"urn": "x"}}}`, nil)

			_, err := c.GetGlossaryParentChain(context.Background(), tt.urn)
			if !errors.Is(err, ErrInvalidURN) {
				t.Fatalf("error = %v, want ErrInvalidURN", err)
			}
		})
	}
}

func TestGetGlossaryParentChainGraphQLError(t *testing.T) {
	c := glossaryTestServer(t, `{"errors": [{"message": "boom"}]}`, nil)

	if _, err := c.GetGlossaryParentChain(context.Background(), "urn:li:glossaryTerm:arr"); err == nil {
		t.Fatal("expected an error for a GraphQL error response")
	}
}

func TestGlossaryPagingUsesDefaultsWhenConfigUnset(t *testing.T) {
	// A zero-value config must not degenerate into asking for count: 0.
	c := &Client{config: Config{}}
	defaults := DefaultConfig()

	tests := []struct {
		name      string
		start     int
		count     int
		wantStart int
		wantCount int
	}{
		{name: "unset count", count: 0, wantCount: defaults.DefaultLimit},
		{name: "negative count", count: -1, wantCount: defaults.DefaultLimit},
		{name: "negative start", start: -3, count: 5, wantCount: 5},
		{name: "above max", count: defaults.MaxLimit + 1, wantCount: defaults.MaxLimit},
		{name: "within range", start: 7, count: 25, wantStart: 7, wantCount: 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, count := c.glossaryPaging(tt.start, tt.count)
			if start != tt.wantStart || count != tt.wantCount {
				t.Errorf("glossaryPaging(%d, %d) = (%d, %d), want (%d, %d)",
					tt.start, tt.count, start, count, tt.wantStart, tt.wantCount)
			}
		})
	}
}
