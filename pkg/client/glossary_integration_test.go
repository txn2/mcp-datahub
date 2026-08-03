//go:build integration

package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// TestIntegrationGlossaryHierarchy exercises glossary node creation and the
// full hierarchy read path against a live DataHub instance.
//
// It exists because the children query cannot be settled by mocks: the parent
// relationship name ("IsPartOf") is a free-form string in RelationshipsInput,
// so only a live instance proves the client asks for the right edge. Verified
// against DataHub v1.6.0.
//
//	export DATAHUB_URL=http://localhost:8080
//	export DATAHUB_TOKEN=<token>   # any non-empty value on an unauthenticated quickstart
//	make test-integration
//
// The test builds this tree, then deletes it:
//
//	root
//	├── child node
//	│   └── grandchild term
//	└── child term
func TestIntegrationGlossaryHierarchy(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	ctx := testCtx(t)

	suffix := nanos()
	rootURN := createGlossaryNodeForTest(ctx, t, c, glossaryEntitySpec{
		Name: fmt.Sprintf("inttest_root_%s", suffix), Definition: "root node",
	})
	childNodeURN := createGlossaryNodeForTest(ctx, t, c, glossaryEntitySpec{
		Name: fmt.Sprintf("inttest_child_%s", suffix), Definition: "child node", Parent: rootURN,
	})
	childTermURN := createGlossaryTermForTest(ctx, t, c, glossaryEntitySpec{
		Name: fmt.Sprintf("inttest_term_%s", suffix), Definition: "child term", Parent: rootURN,
	})
	grandchildTermURN := createGlossaryTermForTest(ctx, t, c, glossaryEntitySpec{
		Name: fmt.Sprintf("inttest_deep_%s", suffix), Definition: "grandchild term", Parent: childNodeURN,
	})

	// Children come from the graph index, which DataHub populates
	// asynchronously, so a freshly created child is not visible at once.
	t.Run("children of root", func(t *testing.T) {
		var children *types.GlossaryChildren
		converged := eventually(t, func() bool {
			var err error
			children, err = c.GetGlossaryNodeChildren(ctx, rootURN, 0, 100)
			if err != nil {
				t.Fatalf("GetGlossaryNodeChildren: %v", err)
			}
			return children.Total == 2 && len(children.Nodes) == 1 && len(children.Terms) == 1
		})
		if !converged {
			t.Fatalf("children of %s did not converge: %+v", rootURN, children)
		}
		if len(children.Nodes) != 1 || children.Nodes[0].URN != childNodeURN {
			t.Errorf("Nodes = %+v, want the single child node %s", children.Nodes, childNodeURN)
		}
		if len(children.Terms) != 1 || children.Terms[0].URN != childTermURN {
			t.Errorf("Terms = %+v, want the single child term %s", children.Terms, childTermURN)
		}
		// childrenCount is DataHub's own tally, so it cross-checks the edge
		// the children query walked.
		if len(children.Nodes) == 1 && children.Nodes[0].TermsCount != 1 {
			t.Errorf("child node TermsCount = %d, want 1", children.Nodes[0].TermsCount)
		}
	})

	t.Run("children paging", func(t *testing.T) {
		page, err := c.GetGlossaryNodeChildren(ctx, rootURN, 1, 1)
		if err != nil {
			t.Fatalf("GetGlossaryNodeChildren: %v", err)
		}
		if page.Total != 2 {
			t.Errorf("Total = %d, want 2 (the full child set, not the page)", page.Total)
		}
		if len(page.Nodes)+len(page.Terms) != 1 {
			t.Errorf("page held %d children, want 1", len(page.Nodes)+len(page.Terms))
		}
	})

	t.Run("children of an unknown node", func(t *testing.T) {
		_, err := c.GetGlossaryNodeChildren(ctx, "urn:li:glossaryNode:inttest_missing_"+suffix, 0, 10)
		if err == nil {
			t.Fatal("expected an error for a glossary node that does not exist")
		}
	})

	t.Run("parent chain of a nested term", func(t *testing.T) {
		chain, err := c.GetGlossaryParentChain(ctx, grandchildTermURN)
		if err != nil {
			t.Fatalf("GetGlossaryParentChain: %v", err)
		}
		if len(chain) != 2 {
			t.Fatalf("len(chain) = %d, want 2", len(chain))
		}
		if chain[0].URN != childNodeURN {
			t.Errorf("chain[0] = %q, want the direct parent %q", chain[0].URN, childNodeURN)
		}
		if chain[1].URN != rootURN {
			t.Errorf("chain[1] = %q, want the root %q", chain[1].URN, rootURN)
		}
		if chain[0].ParentNode != rootURN {
			t.Errorf("chain[0].ParentNode = %q, want %q", chain[0].ParentNode, rootURN)
		}
	})

	t.Run("parent chain of a root node", func(t *testing.T) {
		chain, err := c.GetGlossaryParentChain(ctx, rootURN)
		if err != nil {
			t.Fatalf("GetGlossaryParentChain: %v", err)
		}
		if len(chain) != 0 {
			t.Errorf("chain = %+v, want empty for a root node", chain)
		}
	})

	// Root listings are served from the search index, so they lag the write.
	t.Run("root nodes include the new root", func(t *testing.T) {
		found := eventually(t, func() bool {
			nodes, total, err := c.GetRootGlossaryNodes(ctx, 0, 100)
			if err != nil {
				t.Fatalf("GetRootGlossaryNodes: %v", err)
			}
			if total < len(nodes) {
				t.Errorf("total = %d, want at least the %d nodes returned", total, len(nodes))
			}
			return containsGlossaryNode(nodes, rootURN)
		})
		if !found {
			t.Errorf("root node %s not listed by GetRootGlossaryNodes", rootURN)
		}
	})

	t.Run("root terms exclude a parented term", func(t *testing.T) {
		terms, _, err := c.GetRootGlossaryTerms(ctx, 0, 100)
		if err != nil {
			t.Fatalf("GetRootGlossaryTerms: %v", err)
		}
		for _, term := range terms {
			if term.URN == childTermURN {
				t.Errorf("term %s has a parent node and must not be listed as a root term", childTermURN)
			}
		}
	})

}

// glossaryEntitySpec describes a glossary entity to create for a test.
type glossaryEntitySpec struct {
	Name       string
	Definition string
	Parent     string
}

// createGlossaryNodeForTest creates a glossary node and registers its deletion.
func createGlossaryNodeForTest(ctx context.Context, t *testing.T, c *Client, spec glossaryEntitySpec) string {
	t.Helper()
	urn, err := c.CreateGlossaryNode(ctx, spec.Name, spec.Definition, spec.Parent)
	if err != nil {
		t.Fatalf("CreateGlossaryNode(%s): %v", spec.Name, err)
	}
	if urn == "" {
		t.Fatalf("CreateGlossaryNode(%s) returned an empty URN", spec.Name)
	}
	registerGlossaryCleanup(t, c, urn)
	return urn
}

// createGlossaryTermForTest creates a glossary term and registers its deletion.
func createGlossaryTermForTest(ctx context.Context, t *testing.T, c *Client, spec glossaryEntitySpec) string {
	t.Helper()
	urn, err := c.CreateGlossaryTerm(ctx, spec.Name, spec.Definition, spec.Parent)
	if err != nil {
		t.Fatalf("CreateGlossaryTerm(%s): %v", spec.Name, err)
	}
	registerGlossaryCleanup(t, c, urn)
	return urn
}

// registerGlossaryCleanup deletes a glossary entity when the test finishes.
func registerGlossaryCleanup(t *testing.T, c *Client, urn string) {
	t.Helper()
	t.Cleanup(func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := c.DeleteGlossaryEntity(cleanCtx, urn); err != nil {
			t.Logf("cleanup: failed to delete %s: %v", urn, err)
		}
	})
}

// containsGlossaryNode reports whether the URN appears in the nodes.
func containsGlossaryNode(nodes []types.GlossaryNode, urn string) bool {
	for _, n := range nodes {
		if n.URN == urn {
			return true
		}
	}
	return false
}

// eventually polls cond until it holds or the search index lag budget expires.
func eventually(t *testing.T, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for {
		if cond() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(time.Second)
	}
}
