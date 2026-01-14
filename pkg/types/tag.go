package types

// Tag represents a DataHub tag.
type Tag struct {
	// URN is the unique identifier.
	URN string `json:"urn"`

	// Name is the tag name.
	Name string `json:"name"`

	// Description is the tag description.
	Description string `json:"description,omitempty"`

	// Properties contains custom properties.
	Properties map[string]string `json:"properties,omitempty"`
}
