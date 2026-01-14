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

// GlossaryTermRelation represents a relationship between glossary terms.
type GlossaryTermRelation struct {
	// URN is the related term URN.
	URN string `json:"urn"`

	// Name is the related term name.
	Name string `json:"name"`

	// RelationType is the type of relationship.
	RelationType string `json:"relation_type"`
}
