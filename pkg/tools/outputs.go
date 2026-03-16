package tools

import "github.com/txn2/mcp-datahub/pkg/types"

// BrowseOutput is the structured output of the datahub_browse tool.
// Exactly one of Tags, Domains, or DataProducts is populated per call.
type BrowseOutput struct {
	Tags         []types.Tag         `json:"tags,omitempty"`
	Domains      []types.Domain      `json:"domains,omitempty"`
	DataProducts []types.DataProduct `json:"data_products,omitempty"`
}

// UpdateDescriptionOutput is the structured output of the datahub_update_description tool.
type UpdateDescriptionOutput struct {
	URN    string `json:"urn"`
	Aspect string `json:"aspect"`
	Action string `json:"action"`
}

// AddTagOutput is the structured output of the datahub_add_tag tool.
type AddTagOutput struct {
	URN    string `json:"urn"`
	Tag    string `json:"tag"`
	Aspect string `json:"aspect"`
	Action string `json:"action"`
}

// RemoveTagOutput is the structured output of the datahub_remove_tag tool.
type RemoveTagOutput struct {
	URN    string `json:"urn"`
	Tag    string `json:"tag"`
	Aspect string `json:"aspect"`
	Action string `json:"action"`
}

// AddGlossaryTermOutput is the structured output of the datahub_add_glossary_term tool.
type AddGlossaryTermOutput struct {
	URN    string `json:"urn"`
	Term   string `json:"term"`
	Aspect string `json:"aspect"`
	Action string `json:"action"`
}

// RemoveGlossaryTermOutput is the structured output of the datahub_remove_glossary_term tool.
type RemoveGlossaryTermOutput struct {
	URN    string `json:"urn"`
	Term   string `json:"term"`
	Aspect string `json:"aspect"`
	Action string `json:"action"`
}

// AddLinkOutput is the structured output of the datahub_add_link tool.
type AddLinkOutput struct {
	URN    string `json:"urn"`
	URL    string `json:"url"`
	Aspect string `json:"aspect"`
	Action string `json:"action"`
}

// RemoveLinkOutput is the structured output of the datahub_remove_link tool.
type RemoveLinkOutput struct {
	URN    string `json:"urn"`
	URL    string `json:"url"`
	Aspect string `json:"aspect"`
	Action string `json:"action"`
}

// CreateDocumentOutput is the structured output of the datahub_create_document tool.
type CreateDocumentOutput struct {
	URN    string `json:"urn"`
	Action string `json:"action"`
}

// UpdateDocumentOutput is the structured output of the datahub_update_document tool.
type UpdateDocumentOutput struct {
	URN    string `json:"urn"`
	Action string `json:"action"`
}
