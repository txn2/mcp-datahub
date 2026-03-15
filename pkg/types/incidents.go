package types

// Incident represents a DataHub incident on an entity (DataHub 1.4.x+).
type Incident struct {
	// URN is the unique identifier for this incident.
	URN string `json:"urn"`

	// Type is the incident type (e.g., "OPERATIONAL", "CUSTOM").
	Type string `json:"type"`

	// CustomType is the custom incident type when Type is "CUSTOM".
	CustomType string `json:"custom_type,omitempty"`

	// Title is the incident title.
	Title string `json:"title"`

	// Description explains the incident.
	Description string `json:"description,omitempty"`

	// State is the current status: ACTIVE or RESOLVED.
	State string `json:"state"`

	// Source describes where the incident originated (e.g., "MANUAL").
	Source string `json:"source,omitempty"`

	// Created is the creation timestamp (epoch ms).
	Created int64 `json:"created,omitempty"`

	// CreatedBy is the actor who created the incident.
	CreatedBy string `json:"created_by,omitempty"`

	// LastUpdated is the last update timestamp (epoch ms).
	LastUpdated int64 `json:"last_updated,omitempty"`

	// LastUpdatedBy is the actor who last updated the incident.
	LastUpdatedBy string `json:"last_updated_by,omitempty"`
}

// IncidentResult holds a list of incidents with a total count.
type IncidentResult struct {
	// Total is the total number of matching incidents.
	Total int `json:"total"`

	// Incidents is the list of incidents.
	Incidents []Incident `json:"incidents"`
}

// RaiseIncidentInput contains parameters for creating a new incident.
type RaiseIncidentInput struct {
	// Type is the incident type (e.g., "OPERATIONAL").
	Type string `json:"type"`

	// CustomType is the custom incident type when Type is "CUSTOM".
	CustomType string `json:"custom_type,omitempty"`

	// Title is the incident title.
	Title string `json:"title"`

	// Description explains the incident.
	Description string `json:"description,omitempty"`

	// ResourceURNs are the entity URNs affected by this incident.
	// Only the first element is sent to the DataHub API (resourceUrn field).
	ResourceURNs []string `json:"resource_urns"`
}
