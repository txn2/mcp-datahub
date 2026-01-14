package types

// Owner represents an owner of a DataHub entity.
type Owner struct {
	// URN is the owner's URN (corpuser or corpGroup).
	URN string `json:"urn"`

	// Type is the ownership type.
	Type OwnershipType `json:"type"`

	// Name is the owner's display name.
	Name string `json:"name,omitempty"`

	// Email is the owner's email address.
	Email string `json:"email,omitempty"`
}

// OwnershipType represents the type of ownership.
type OwnershipType string

// Ownership type constants.
const (
	OwnershipTypeTechnicalOwner OwnershipType = "TECHNICAL_OWNER"
	OwnershipTypeBusinessOwner  OwnershipType = "BUSINESS_OWNER"
	OwnershipTypeDataSteward    OwnershipType = "DATA_STEWARD"
	OwnershipTypeNone           OwnershipType = "NONE"
)
