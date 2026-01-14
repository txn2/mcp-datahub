package types

// LineageResult represents the lineage graph for an entity.
type LineageResult struct {
	// Start is the URN of the entity we queried lineage for.
	Start string `json:"start"`

	// Nodes are the entities in the lineage graph.
	Nodes []LineageNode `json:"nodes"`

	// Edges are the relationships between nodes.
	Edges []LineageEdge `json:"edges"`

	// Direction is the lineage direction (UPSTREAM or DOWNSTREAM).
	Direction string `json:"direction"`

	// Depth is the depth of the lineage traversal.
	Depth int `json:"depth"`
}

// LineageNode represents an entity in the lineage graph.
type LineageNode struct {
	// URN is the unique identifier.
	URN string `json:"urn"`

	// Type is the entity type.
	Type string `json:"type"`

	// Name is the display name.
	Name string `json:"name"`

	// Platform is the data platform.
	Platform string `json:"platform,omitempty"`

	// Description is the entity description.
	Description string `json:"description,omitempty"`

	// Level is the distance from the start node.
	Level int `json:"level"`
}

// LineageEdge represents a lineage relationship between two entities.
type LineageEdge struct {
	// Source is the source entity URN.
	Source string `json:"source"`

	// Target is the target entity URN.
	Target string `json:"target"`

	// Type is the relationship type.
	Type string `json:"type,omitempty"`

	// Created is when the relationship was created.
	Created int64 `json:"created,omitempty"`

	// UpdatedBy is who created/updated the relationship.
	UpdatedBy string `json:"updated_by,omitempty"`

	// Properties contains additional edge properties.
	Properties map[string]any `json:"properties,omitempty"`
}
