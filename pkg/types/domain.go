package types

// Domain represents a DataHub data domain.
type Domain struct {
	// URN is the unique identifier.
	URN string `json:"urn"`

	// Name is the domain name.
	Name string `json:"name"`

	// Description is the domain description.
	Description string `json:"description,omitempty"`

	// ParentDomain is the parent domain URN.
	ParentDomain string `json:"parent_domain,omitempty"`

	// Owners are the domain owners.
	Owners []Owner `json:"owners,omitempty"`

	// EntityCount is the number of entities in this domain.
	EntityCount int `json:"entity_count,omitempty"`
}
