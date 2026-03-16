package types

// Document represents a DataHub context document (DataHub 1.4.x+).
// Documents are first-class entities designed for AI consumption — tutorials,
// runbooks, FAQs, and reference guides linked to data assets.
type Document struct {
	// URN is the unique identifier (urn:li:document:{id}).
	URN string `json:"urn"`

	// Title is the document title.
	Title string `json:"title"`

	// Content is the document body text.
	Content string `json:"content,omitempty"`

	// Status is the publication state: PUBLISHED or UNPUBLISHED.
	Status string `json:"status,omitempty"`

	// SubType is the document sub-type classification.
	SubType string `json:"sub_type,omitempty"`

	// Source describes where the document originated.
	Source *DocumentSource `json:"source,omitempty"`

	// Settings contains visibility and behavior settings.
	Settings *DocumentSettings `json:"settings,omitempty"`

	// RelatedAssets lists the data entities linked to this document.
	RelatedAssets []DocumentRelatedAsset `json:"related_assets,omitempty"`

	// RelatedDocuments lists other documents linked to this document.
	RelatedDocuments []DocumentRelatedDocument `json:"related_documents,omitempty"`

	// ParentDocument is the parent document in a hierarchy.
	ParentDocument *DocumentParent `json:"parent_document,omitempty"`

	// Owners lists the document owners.
	Owners []Owner `json:"owners,omitempty"`

	// Tags lists tags applied to this document.
	Tags []Tag `json:"tags,omitempty"`

	// GlossaryTerms lists glossary terms associated with this document.
	GlossaryTerms []GlossaryTerm `json:"glossary_terms,omitempty"`

	// Domain is the data domain this document belongs to.
	Domain *Domain `json:"domain,omitempty"`

	// Created is the creation timestamp (epoch ms).
	Created int64 `json:"created,omitempty"`

	// LastModified is the last modification timestamp (epoch ms).
	LastModified int64 `json:"last_modified,omitempty"`
}

// DocumentSource describes the origin of a document.
type DocumentSource struct {
	// SourceType is NATIVE or EXTERNAL.
	SourceType string `json:"source_type"`

	// ExternalURL is the URL for external documents (Confluence, Notion, etc.).
	ExternalURL string `json:"external_url,omitempty"`
}

// DocumentSettings contains document visibility settings.
type DocumentSettings struct {
	// ShowInGlobalContext controls whether the document appears in global search.
	// When false, the document is only accessible through linked assets.
	ShowInGlobalContext bool `json:"show_in_global_context"`
}

// DocumentRelatedAsset links a document to a data entity.
type DocumentRelatedAsset struct {
	// URN is the data entity URN.
	URN string `json:"urn"`
}

// DocumentRelatedDocument links a document to another document.
type DocumentRelatedDocument struct {
	// URN is the related document URN.
	URN string `json:"urn"`
}

// DocumentParent is the parent document in a hierarchy.
type DocumentParent struct {
	// URN is the parent document URN.
	URN string `json:"urn"`
}
