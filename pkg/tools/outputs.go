package tools

import "github.com/txn2/mcp-datahub/pkg/types"

// BrowseOutput is the structured output of the datahub_browse tool.
// Exactly one of Tags, Domains, or DataProducts is populated per call.
type BrowseOutput struct {
	Tags         []types.Tag         `json:"tags,omitempty"`
	Domains      []types.Domain      `json:"domains,omitempty"`
	DataProducts []types.DataProduct `json:"data_products,omitempty"`
}

// CreateOutput is the structured output of the datahub_create tool.
type CreateOutput struct {
	URN    string `json:"urn"`
	What   string `json:"what"`
	Action string `json:"action"`
}

// UpdateOutput is the structured output of the datahub_update tool.
type UpdateOutput struct {
	URN       string `json:"urn"`
	What      string `json:"what"`
	Action    string `json:"action"`
	TargetURN string `json:"target_urn,omitempty"`
}

// DeleteOutput is the structured output of the datahub_delete tool.
type DeleteOutput struct {
	URN    string `json:"urn"`
	What   string `json:"what"`
	Action string `json:"action"`
}
