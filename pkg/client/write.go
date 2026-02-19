package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// entityTypeFromURN derives the DataHub entity type string from a parsed URN.
// Maps URN entity types to the REST API entity type names.
func entityTypeFromURN(urn string) (string, error) {
	parsed, err := ParseURN(urn)
	if err != nil {
		return "", err
	}
	return parsed.EntityType, nil
}

// editableSchemaAspect is the REST API representation of editableSchemaMetadata.
type editableSchemaAspect struct {
	EditableSchemaFieldInfo []editableFieldInfo `json:"editableSchemaFieldInfo"`
}

// editableFieldInfo represents a field's editable metadata for REST API read-modify-write.
// Uses json.RawMessage for tags and glossaryTerms to preserve existing data.
type editableFieldInfo struct {
	FieldPath     string          `json:"fieldPath"`
	Description   string          `json:"description,omitempty"`
	GlobalTags    json.RawMessage `json:"globalTags,omitempty"`
	GlossaryTerms json.RawMessage `json:"glossaryTerms,omitempty"`
}

// editablePropertiesAspect represents the editableDatasetProperties aspect.
type editablePropertiesAspect struct {
	Description  string         `json:"description"`
	Created      *auditStampRaw `json:"created,omitempty"`
	LastModified *auditStampRaw `json:"lastModified,omitempty"`
}

// UpdateDescription sets the editable description for any entity using read-modify-write.
func (c *Client) UpdateDescription(ctx context.Context, urn, description string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return fmt.Errorf("UpdateDescription: %w", err)
	}

	props, err := c.readEditableProperties(ctx, urn)
	if err != nil {
		return fmt.Errorf("UpdateDescription: %w", err)
	}

	props.Description = description

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: "editableDatasetProperties",
		Aspect:     props,
	})
}

// readEditableProperties reads the current editableDatasetProperties aspect.
// Returns an empty aspect if none exists (not an error).
func (c *Client) readEditableProperties(ctx context.Context, urn string) (*editablePropertiesAspect, error) {
	raw, err := c.getAspect(ctx, urn, "editableDatasetProperties")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &editablePropertiesAspect{}, nil
		}
		return nil, fmt.Errorf("reading editableDatasetProperties: %w", err)
	}

	var props editablePropertiesAspect
	if err := json.Unmarshal(raw, &props); err != nil {
		return nil, fmt.Errorf("parsing editableDatasetProperties: %w", err)
	}
	return &props, nil
}

// globalTagsAspect represents the globalTags aspect structure.
type globalTagsAspect struct {
	Tags []tagAssociation `json:"tags"`
}

// tagAssociation represents a tag association in the globalTags aspect.
type tagAssociation struct {
	Tag string `json:"tag"`
}

// AddTag adds a tag to an entity using read-modify-write on the globalTags aspect.
func (c *Client) AddTag(ctx context.Context, urn, tagURN string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return fmt.Errorf("AddTag: %w", err)
	}

	// Read current tags
	tags, err := c.readGlobalTags(ctx, urn)
	if err != nil {
		return fmt.Errorf("AddTag: %w", err)
	}

	// Check for duplicate
	for _, t := range tags.Tags {
		if t.Tag == tagURN {
			return nil // Already present
		}
	}

	// Add and write
	tags.Tags = append(tags.Tags, tagAssociation{Tag: tagURN})

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: "globalTags",
		Aspect:     tags,
	})
}

// RemoveTag removes a tag from an entity using read-modify-write on the globalTags aspect.
func (c *Client) RemoveTag(ctx context.Context, urn, tagURN string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return fmt.Errorf("RemoveTag: %w", err)
	}

	// Read current tags
	tags, err := c.readGlobalTags(ctx, urn)
	if err != nil {
		return fmt.Errorf("RemoveTag: %w", err)
	}

	// Filter out the tag
	filtered := make([]tagAssociation, 0, len(tags.Tags))
	for _, t := range tags.Tags {
		if t.Tag != tagURN {
			filtered = append(filtered, t)
		}
	}
	tags.Tags = filtered

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: "globalTags",
		Aspect:     tags,
	})
}

// readGlobalTags reads the current globalTags aspect for an entity.
// Returns an empty aspect if none exists (not an error).
func (c *Client) readGlobalTags(ctx context.Context, urn string) (*globalTagsAspect, error) {
	raw, err := c.getAspect(ctx, urn, "globalTags")
	if err != nil {
		// Not found means no tags yet - return empty
		if errors.Is(err, ErrNotFound) {
			return &globalTagsAspect{Tags: []tagAssociation{}}, nil
		}
		return nil, fmt.Errorf("reading globalTags: %w", err)
	}

	var tags globalTagsAspect
	if err := json.Unmarshal(raw, &tags); err != nil {
		return nil, fmt.Errorf("parsing globalTags: %w", err)
	}
	return &tags, nil
}

// glossaryTermsAspect represents the glossaryTerms aspect structure.
type glossaryTermsAspect struct {
	Terms []termAssociation `json:"terms"`
}

// termAssociation represents a glossary term association.
type termAssociation struct {
	URN string `json:"urn"`
}

// AddGlossaryTerm adds a glossary term to an entity using read-modify-write.
func (c *Client) AddGlossaryTerm(ctx context.Context, urn, termURN string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return fmt.Errorf("AddGlossaryTerm: %w", err)
	}

	terms, err := c.readGlossaryTerms(ctx, urn)
	if err != nil {
		return fmt.Errorf("AddGlossaryTerm: %w", err)
	}

	// Check for duplicate
	for _, t := range terms.Terms {
		if t.URN == termURN {
			return nil
		}
	}

	terms.Terms = append(terms.Terms, termAssociation{URN: termURN})

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: "glossaryTerms",
		Aspect:     terms,
	})
}

// RemoveGlossaryTerm removes a glossary term from an entity using read-modify-write.
func (c *Client) RemoveGlossaryTerm(ctx context.Context, urn, termURN string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return fmt.Errorf("RemoveGlossaryTerm: %w", err)
	}

	terms, err := c.readGlossaryTerms(ctx, urn)
	if err != nil {
		return fmt.Errorf("RemoveGlossaryTerm: %w", err)
	}

	filtered := make([]termAssociation, 0, len(terms.Terms))
	for _, t := range terms.Terms {
		if t.URN != termURN {
			filtered = append(filtered, t)
		}
	}
	terms.Terms = filtered

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: "glossaryTerms",
		Aspect:     terms,
	})
}

// readGlossaryTerms reads the current glossaryTerms aspect for an entity.
func (c *Client) readGlossaryTerms(ctx context.Context, urn string) (*glossaryTermsAspect, error) {
	raw, err := c.getAspect(ctx, urn, "glossaryTerms")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &glossaryTermsAspect{Terms: []termAssociation{}}, nil
		}
		return nil, fmt.Errorf("reading glossaryTerms: %w", err)
	}

	var terms glossaryTermsAspect
	if err := json.Unmarshal(raw, &terms); err != nil {
		return nil, fmt.Errorf("parsing glossaryTerms: %w", err)
	}
	return &terms, nil
}

// institutionalMemoryAspect represents the institutionalMemory aspect.
type institutionalMemoryAspect struct {
	Elements []linkElement `json:"elements"`
}

// linkElement represents a link in the institutionalMemory aspect.
type linkElement struct {
	URL         string        `json:"url"`
	Description string        `json:"description"`
	Created     auditStampRaw `json:"created"`
}

// auditStampRaw represents an audit stamp with millisecond timestamp.
type auditStampRaw struct {
	Time  int64  `json:"time"`
	Actor string `json:"actor"`
}

// AddLink adds a link to an entity using read-modify-write on institutionalMemory.
func (c *Client) AddLink(ctx context.Context, urn, linkURL, description string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return fmt.Errorf("AddLink: %w", err)
	}

	memory, err := c.readInstitutionalMemory(ctx, urn)
	if err != nil {
		return fmt.Errorf("AddLink: %w", err)
	}

	// Check for duplicate URL
	for _, e := range memory.Elements {
		if e.URL == linkURL {
			return nil
		}
	}

	memory.Elements = append(memory.Elements, linkElement{
		URL:         linkURL,
		Description: description,
		Created: auditStampRaw{
			Time:  0, // DataHub will fill this in
			Actor: "urn:li:corpuser:datahub",
		},
	})

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: "institutionalMemory",
		Aspect:     memory,
	})
}

// RemoveLink removes a link from an entity by URL using read-modify-write.
func (c *Client) RemoveLink(ctx context.Context, urn, linkURL string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return fmt.Errorf("RemoveLink: %w", err)
	}

	memory, err := c.readInstitutionalMemory(ctx, urn)
	if err != nil {
		return fmt.Errorf("RemoveLink: %w", err)
	}

	filtered := make([]linkElement, 0, len(memory.Elements))
	for _, e := range memory.Elements {
		if e.URL != linkURL {
			filtered = append(filtered, e)
		}
	}
	memory.Elements = filtered

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: "institutionalMemory",
		Aspect:     memory,
	})
}

// readInstitutionalMemory reads the current institutionalMemory aspect.
func (c *Client) readInstitutionalMemory(ctx context.Context, urn string) (*institutionalMemoryAspect, error) {
	raw, err := c.getAspect(ctx, urn, "institutionalMemory")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &institutionalMemoryAspect{Elements: []linkElement{}}, nil
		}
		return nil, fmt.Errorf("reading institutionalMemory: %w", err)
	}

	var memory institutionalMemoryAspect
	if err := json.Unmarshal(raw, &memory); err != nil {
		return nil, fmt.Errorf("parsing institutionalMemory: %w", err)
	}
	return &memory, nil
}

// UpdateColumnDescription sets the editable description for a specific column
// using read-modify-write on the editableSchemaMetadata aspect.
func (c *Client) UpdateColumnDescription(ctx context.Context, urn, fieldPath, description string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return fmt.Errorf("UpdateColumnDescription: %w", err)
	}

	schema, err := c.readEditableSchema(ctx, urn)
	if err != nil {
		return fmt.Errorf("UpdateColumnDescription: %w", err)
	}

	// Find or create the field entry
	found := false
	for i := range schema.EditableSchemaFieldInfo {
		if schema.EditableSchemaFieldInfo[i].FieldPath == fieldPath {
			schema.EditableSchemaFieldInfo[i].Description = description
			found = true
			break
		}
	}
	if !found {
		schema.EditableSchemaFieldInfo = append(schema.EditableSchemaFieldInfo, editableFieldInfo{
			FieldPath:   fieldPath,
			Description: description,
		})
	}

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: "editableSchemaMetadata",
		Aspect:     schema,
	})
}

// readEditableSchema reads the current editableSchemaMetadata aspect.
// Returns an empty aspect if none exists (not an error).
func (c *Client) readEditableSchema(ctx context.Context, urn string) (*editableSchemaAspect, error) {
	raw, err := c.getAspect(ctx, urn, "editableSchemaMetadata")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &editableSchemaAspect{}, nil
		}
		return nil, fmt.Errorf("reading editableSchemaMetadata: %w", err)
	}

	var schema editableSchemaAspect
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, fmt.Errorf("parsing editableSchemaMetadata: %w", err)
	}
	return &schema, nil
}
