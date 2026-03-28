package types

// ContextDocumentAuthor identifies the author of a context document.
type ContextDocumentAuthor struct {
	// URN is the author entity URN (e.g., urn:li:corpuser:alice).
	URN string `json:"urn"`

	// Username is the author's login name.
	Username string `json:"username,omitempty"`
}

// ContextDocument is a simplified view of a Document for entity-attached
// context documentation. Downstream systems like mcp-data-platform consume
// this type for knowledge pipeline operations (capture, enrich, apply).
type ContextDocument struct {
	// ID is the document identifier (the portion after urn:li:document:).
	ID string `json:"id"`

	// Title is the document title.
	Title string `json:"title"`

	// Content is the document body text.
	Content string `json:"content,omitempty"`

	// ContentType describes the content format (e.g., "text/markdown").
	ContentType string `json:"content_type"`

	// Category classifies the document (maps to DataHub subType).
	Category string `json:"category,omitempty"`

	// CreatedAt is the creation timestamp (epoch ms).
	CreatedAt int64 `json:"created_at,omitempty"`

	// UpdatedAt is the last modification timestamp (epoch ms).
	UpdatedAt int64 `json:"updated_at,omitempty"`

	// Author is the primary author derived from ownership.
	Author *ContextDocumentAuthor `json:"author,omitempty"`
}

// ContextDocumentInput contains parameters for creating or updating a context document.
// When ID is empty, a new document is created. When set, the existing document is updated.
type ContextDocumentInput struct {
	// ID is the document identifier. Empty means create; populated means update.
	ID string `json:"id,omitempty"`

	// Title is the document title (required).
	Title string `json:"title"`

	// Content is the document body text.
	Content string `json:"content"`

	// Category classifies the document (maps to DataHub subType).
	Category string `json:"category,omitempty"`
}
