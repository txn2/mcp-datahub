package types

// GlossaryTerm represents a business glossary term.
type GlossaryTerm struct {
	// URN is the unique identifier.
	URN string `json:"urn"`

	// Name is the term name.
	Name string `json:"name"`

	// Description is the term definition.
	Description string `json:"description,omitempty"`

	// ParentNode is the parent glossary node URN.
	ParentNode string `json:"parent_node,omitempty"`

	// Owners are the term owners.
	Owners []Owner `json:"owners,omitempty"`

	// RelatedTerms are related glossary terms.
	RelatedTerms []GlossaryTermRelation `json:"related_terms,omitempty"`

	// Properties contains custom properties.
	Properties map[string]string `json:"properties,omitempty"`
}

// GlossaryNode represents a business glossary node: a directory in the
// glossary that contains glossary terms and other glossary nodes.
type GlossaryNode struct {
	// URN is the unique identifier.
	URN string `json:"urn"`

	// Name is the node name.
	Name string `json:"name"`

	// Description is the node definition. DataHub stores it in the
	// glossaryNodeInfo aspect's "definition" field and exposes it as
	// "description" on the GraphQL API.
	Description string `json:"description,omitempty"`

	// ParentNode is the parent glossary node URN. Empty for a root node.
	ParentNode string `json:"parent_node,omitempty"`

	// TermsCount is the number of glossary terms directly under this node.
	TermsCount int `json:"terms_count"`

	// NodesCount is the number of glossary nodes directly under this node.
	NodesCount int `json:"nodes_count"`
}

// GlossaryChildren is a page of the entities directly under a glossary node.
// A node's children are a single mixed collection in DataHub, so Nodes and
// Terms are two views of the same page: Total, Start, and Count describe the
// combined result set, not either slice on its own.
type GlossaryChildren struct {
	// Nodes are the child glossary nodes in this page.
	Nodes []GlossaryNode `json:"nodes,omitempty"`

	// Terms are the child glossary terms in this page.
	Terms []GlossaryTerm `json:"terms,omitempty"`

	// Start is the offset of this page within the child set.
	Start int `json:"start"`

	// Count is the number of children returned in this page.
	Count int `json:"count"`

	// Total is the number of children under the node.
	Total int `json:"total"`
}

// GlossaryTermRelation represents a relationship between glossary terms.
type GlossaryTermRelation struct {
	// URN is the related term URN.
	URN string `json:"urn"`

	// Name is the related term name.
	Name string `json:"name"`

	// RelationType is the type of relationship.
	RelationType string `json:"relation_type"`
}
