package types

// SearchResult represents search results from DataHub.
type SearchResult struct {
	// Entities is the list of matching entities.
	Entities []SearchEntity `json:"entities"`

	// Total is the total number of matches.
	Total int `json:"total"`

	// Offset is the result offset.
	Offset int `json:"offset"`

	// Limit is the result limit.
	Limit int `json:"limit"`
}

// SearchEntity represents a single search result entity.
type SearchEntity struct {
	// URN is the unique identifier.
	URN string `json:"urn"`

	// Type is the entity type.
	Type string `json:"type"`

	// Name is the display name.
	Name string `json:"name"`

	// Description is the entity description.
	Description string `json:"description,omitempty"`

	// Platform is the data platform.
	Platform string `json:"platform,omitempty"`

	// Owners are the entity owners.
	Owners []Owner `json:"owners,omitempty"`

	// Tags are the entity tags.
	Tags []Tag `json:"tags,omitempty"`

	// Domain is the entity domain.
	Domain *Domain `json:"domain,omitempty"`

	// MatchedFields shows which fields matched the query.
	MatchedFields []MatchedField `json:"matched_fields,omitempty"`
}

// MatchedField indicates a field that matched the search query.
type MatchedField struct {
	// Name is the field name.
	Name string `json:"name"`

	// Value is the matched value.
	Value string `json:"value"`
}
