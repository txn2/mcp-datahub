package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// glossaryChildRelationshipType is the graph relationship DataHub writes
// between a glossary entity and its parent node. A node's children are the
// INCOMING side of that edge.
//
// RelationshipsInput.types is [String!] in the schema, so the name is not
// discoverable from the schema files. Verified empirically against DataHub
// v1.6.0 (see TestGlossaryHierarchyIntegration) and against the upstream UI
// query datahub-web-react/src/graphql/glossaryNode.graphql, which fetches node
// children the same way.
const glossaryChildRelationshipType = "IsPartOf"

// Glossary hierarchy queries.
const (
	// GetRootGlossaryNodesQuery lists glossary nodes that have no parent node.
	GetRootGlossaryNodesQuery = `
query getRootGlossaryNodes($input: GetRootGlossaryEntitiesInput!) {
  getRootGlossaryNodes(input: $input) {
    start
    count
    total
    nodes {
      urn
      properties {
        name
        description
      }
      childrenCount {
        termsCount
        nodesCount
      }
    }
  }
}
`

	// GetRootGlossaryTermsQuery lists glossary terms that have no parent node.
	GetRootGlossaryTermsQuery = `
query getRootGlossaryTerms($input: GetRootGlossaryEntitiesInput!) {
  getRootGlossaryTerms(input: $input) {
    start
    count
    total
    terms {
      urn
      name
      properties {
        name
        description
      }
    }
  }
}
`

	// GetGlossaryNodeChildrenQuery lists the nodes and terms directly under a
	// glossary node. Children arrive as a single mixed relationship page; the
	// entity "type" field discriminates nodes from terms.
	GetGlossaryNodeChildrenQuery = `
query getGlossaryNodeChildren($urn: String!, $input: RelationshipsInput!) {
  glossaryNode(urn: $urn) {
    urn
    exists
    children: relationships(input: $input) {
      start
      count
      total
      relationships {
        entity {
          urn
          type
          ... on GlossaryNode {
            properties {
              name
              description
            }
            childrenCount {
              termsCount
              nodesCount
            }
          }
          ... on GlossaryTerm {
            name
            properties {
              name
              description
            }
          }
        }
      }
    }
  }
}
`

	// GetGlossaryParentChainQuery reads the ancestor nodes of a glossary term
	// or glossary node. The polymorphic entity(urn:) lookup lets one query
	// serve both, since parentNodes exists on each.
	GetGlossaryParentChainQuery = `
query getGlossaryParentChain($urn: String!) {
  entity(urn: $urn) {
    urn
    type
    ... on GlossaryTerm {
      parentNodes {
        count
        nodes {
          urn
          properties {
            name
            description
          }
          childrenCount {
            termsCount
            nodesCount
          }
        }
      }
    }
    ... on GlossaryNode {
      parentNodes {
        count
        nodes {
          urn
          properties {
            name
            description
          }
          childrenCount {
            termsCount
            nodesCount
          }
        }
      }
    }
  }
}
`
)

// GraphQL EntityType values for glossary entities, used to discriminate the
// mixed children of a glossary node.
const (
	graphQLTypeGlossaryNode = "GLOSSARY_NODE"
	graphQLTypeGlossaryTerm = "GLOSSARY_TERM"
)

// glossaryEntityFields decodes a glossary node or term. DataHub returns both
// through the same relationship edge, so the children selection merges the two
// inline fragments into one JSON object shape. Term-only and node-only fields
// are simply absent for the other kind.
type glossaryEntityFields struct {
	URN        string `json:"urn"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Properties struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"properties"`
	ChildrenCount struct {
		TermsCount int `json:"termsCount"`
		NodesCount int `json:"nodesCount"`
	} `json:"childrenCount"`
}

// displayName prefers the properties name, which carries the display name a
// user set, and falls back to the entity name, which is the URN-derived id.
func (g glossaryEntityFields) displayName() string {
	if g.Properties.Name != "" {
		return g.Properties.Name
	}
	return g.Name
}

// toNode converts the decoded fields to a glossary node with a known parent.
func (g glossaryEntityFields) toNode(parentNode string) types.GlossaryNode {
	return types.GlossaryNode{
		URN:         g.URN,
		Name:        g.displayName(),
		Description: g.Properties.Description,
		ParentNode:  parentNode,
		TermsCount:  g.ChildrenCount.TermsCount,
		NodesCount:  g.ChildrenCount.NodesCount,
	}
}

// toTerm converts the decoded fields to a glossary term with a known parent.
func (g glossaryEntityFields) toTerm(parentNode string) types.GlossaryTerm {
	return types.GlossaryTerm{
		URN:         g.URN,
		Name:        g.displayName(),
		Description: g.Properties.Description,
		ParentNode:  parentNode,
	}
}

// requireGlossaryURN checks that urn names one of the allowed glossary entity
// types. The hierarchy queries resolve whatever URN they are handed — glossary
// or not — so rejecting the wrong kind here turns a silently empty result into
// a clear error.
func requireGlossaryURN(urn string, allowed ...string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return err
	}
	for _, a := range allowed {
		if entityType == a {
			return nil
		}
	}
	return fmt.Errorf("%w: %s is not a %s URN", ErrInvalidURN, urn, strings.Join(allowed, " or "))
}

// glossaryPaging normalizes start/count for the glossary hierarchy queries.
// DataHub requires both, so a non-positive count falls back to DefaultLimit and
// anything above MaxLimit is clamped rather than passed through.
//
// The limits fall back to DefaultConfig when unset, because sending count: 0
// would ask DataHub for an empty page rather than a default-sized one.
func (c *Client) glossaryPaging(start, count int) (int, int) {
	defaults := DefaultConfig()

	defaultLimit := c.config.DefaultLimit
	if defaultLimit <= 0 {
		defaultLimit = defaults.DefaultLimit
	}
	maxLimit := c.config.MaxLimit
	if maxLimit <= 0 {
		maxLimit = defaults.MaxLimit
	}

	if start < 0 {
		start = 0
	}
	switch {
	case count <= 0:
		count = defaultLimit
	case count > maxLimit:
		count = maxLimit
	}
	return start, count
}

// rootGlossaryPage decodes either root query. The two share a shape because
// DataHub returns the same entity projection under a different list name, so
// only the field matching the query is populated.
type rootGlossaryPage struct {
	Total int                    `json:"total"`
	Nodes []glossaryEntityFields `json:"nodes"`
	Terms []glossaryEntityFields `json:"terms"`
}

// fetchRootGlossaryPage runs a root glossary query and returns the page found
// under respField, the name of the query's field in the GraphQL response.
func (c *Client) fetchRootGlossaryPage(ctx context.Context, query, respField string, start, count int) (rootGlossaryPage, error) {
	start, count = c.glossaryPaging(start, count)

	variables := map[string]any{
		"input": map[string]any{"start": start, "count": count},
	}
	var response map[string]rootGlossaryPage
	if err := c.Execute(ctx, query, variables, &response); err != nil {
		return rootGlossaryPage{}, err
	}
	return response[respField], nil
}

// GetRootGlossaryNodes lists glossary nodes that have no parent node, i.e. the
// top level of the glossary tree. It returns the requested page and the total
// number of root nodes so the caller can page through them.
func (c *Client) GetRootGlossaryNodes(ctx context.Context, start, count int) ([]types.GlossaryNode, int, error) {
	page, err := c.fetchRootGlossaryPage(ctx, GetRootGlossaryNodesQuery, "getRootGlossaryNodes", start, count)
	if err != nil {
		return nil, 0, fmt.Errorf("GetRootGlossaryNodes: %w", err)
	}

	nodes := make([]types.GlossaryNode, 0, len(page.Nodes))
	for _, n := range page.Nodes {
		// Root nodes have no parent by definition.
		nodes = append(nodes, n.toNode(""))
	}
	return nodes, page.Total, nil
}

// GetRootGlossaryTerms lists glossary terms that have no parent node. It
// returns the requested page and the total number of root terms so the caller
// can page through them.
func (c *Client) GetRootGlossaryTerms(ctx context.Context, start, count int) ([]types.GlossaryTerm, int, error) {
	page, err := c.fetchRootGlossaryPage(ctx, GetRootGlossaryTermsQuery, "getRootGlossaryTerms", start, count)
	if err != nil {
		return nil, 0, fmt.Errorf("GetRootGlossaryTerms: %w", err)
	}

	terms := make([]types.GlossaryTerm, 0, len(page.Terms))
	for _, t := range page.Terms {
		// Root terms have no parent by definition.
		terms = append(terms, t.toTerm(""))
	}
	return terms, page.Total, nil
}

// GetGlossaryNodeChildren lists the glossary nodes and terms directly under
// nodeURN. Nodes and terms share one paged result set in DataHub, so Total,
// Start, and Count describe the combined page rather than either slice.
//
// Returns ErrInvalidURN unless nodeURN is a glossary node URN, and ErrNotFound
// if that node does not exist; DataHub answers a lookup of an unknown glossary
// node with an empty stub rather than an error, which is otherwise
// indistinguishable from a node with no children.
//
// Children are read from DataHub's graph index, which is populated
// asynchronously: a term or node created moments earlier may not appear yet.
// GetGlossaryParentChain reads the entity itself and is immediately
// consistent, so prefer it when confirming a just-written parent.
func (c *Client) GetGlossaryNodeChildren(ctx context.Context, nodeURN string, start, count int) (*types.GlossaryChildren, error) {
	if err := requireGlossaryURN(nodeURN, entityTypeGlossaryNode); err != nil {
		return nil, fmt.Errorf("GetGlossaryNodeChildren(%s): %w", nodeURN, err)
	}
	start, count = c.glossaryPaging(start, count)

	var response struct {
		GlossaryNode struct {
			URN    string `json:"urn"`
			Exists *bool  `json:"exists"`
			// The query aliases the relationships page to "children".
			Children struct {
				Start         int `json:"start"`
				Count         int `json:"count"`
				Total         int `json:"total"`
				Relationships []struct {
					Entity glossaryEntityFields `json:"entity"`
				} `json:"relationships"`
			} `json:"children"`
		} `json:"glossaryNode"`
	}

	variables := map[string]any{
		"urn": nodeURN,
		"input": map[string]any{
			"types":     []string{glossaryChildRelationshipType},
			"direction": "INCOMING",
			"start":     start,
			"count":     count,
		},
	}
	if err := c.Execute(ctx, GetGlossaryNodeChildrenQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetGlossaryNodeChildren(%s): %w", nodeURN, err)
	}

	// Only a definitive false means absent. Older DataHub versions that omit
	// "exists" leave the pointer nil, and the children page still stands.
	if response.GlossaryNode.Exists != nil && !*response.GlossaryNode.Exists {
		return nil, fmt.Errorf("GetGlossaryNodeChildren(%s): %w", nodeURN, ErrNotFound)
	}

	page := response.GlossaryNode.Children
	children := &types.GlossaryChildren{
		Start: page.Start,
		Count: page.Count,
		Total: page.Total,
	}
	for _, rel := range page.Relationships {
		// Only glossary entities can be children of a node; anything else on
		// the edge is left out of the split but still counted in Total.
		switch rel.Entity.Type {
		case graphQLTypeGlossaryNode:
			children.Nodes = append(children.Nodes, rel.Entity.toNode(nodeURN))
		case graphQLTypeGlossaryTerm:
			children.Terms = append(children.Terms, rel.Entity.toTerm(nodeURN))
		}
	}
	return children, nil
}

// GetGlossaryParentChain returns the ancestor nodes of a glossary term or
// glossary node, ordered from the direct parent up to the root. A root entity
// has an empty chain, and a URN that is neither a term nor a node is rejected
// with ErrInvalidURN.
//
// Each returned node carries the URN of its own parent — the next element in
// the chain — so the caller can rebuild the branch without another lookup.
func (c *Client) GetGlossaryParentChain(ctx context.Context, urn string) ([]types.GlossaryNode, error) {
	// entity(urn:) resolves any entity type, so reject non-glossary URNs here
	// rather than returning a silently empty chain for, say, a dataset.
	if err := requireGlossaryURN(urn, entityTypeGlossaryTerm, entityTypeGlossaryNode); err != nil {
		return nil, fmt.Errorf("GetGlossaryParentChain(%s): %w", urn, err)
	}

	var response struct {
		Entity struct {
			URN         string `json:"urn"`
			ParentNodes struct {
				Count int                    `json:"count"`
				Nodes []glossaryEntityFields `json:"nodes"`
			} `json:"parentNodes"`
		} `json:"entity"`
	}

	if err := c.Execute(ctx, GetGlossaryParentChainQuery, map[string]any{"urn": urn}, &response); err != nil {
		return nil, fmt.Errorf("GetGlossaryParentChain(%s): %w", urn, err)
	}

	raw := response.Entity.ParentNodes.Nodes
	chain := make([]types.GlossaryNode, 0, len(raw))
	for i, n := range raw {
		// The chain runs child-to-root, so the next element is this node's parent.
		parent := ""
		if i+1 < len(raw) {
			parent = raw[i+1].URN
		}
		chain = append(chain, n.toNode(parent))
	}
	return chain, nil
}
