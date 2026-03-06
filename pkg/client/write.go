package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
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

// descriptionAspectInfo holds the aspect name and field name for updating
// an entity's description. The field is "description" for most entity types,
// but "definition" for glossary entities.
type descriptionAspectInfo struct {
	AspectName string
	FieldName  string
}

// descriptionAspectMap maps DataHub entity types to their description aspect.
// glossaryTerm and glossaryNode use "definition" instead of "description".
// dataProduct and domain use non-editable property aspects.
var descriptionAspectMap = map[string]descriptionAspectInfo{
	"dataset":      {AspectName: "editableDatasetProperties", FieldName: "description"},
	"dashboard":    {AspectName: "editableDashboardProperties", FieldName: "description"},
	"chart":        {AspectName: "editableChartProperties", FieldName: "description"},
	"dataFlow":     {AspectName: "editableDataFlowProperties", FieldName: "description"},
	"dataJob":      {AspectName: "editableDataJobProperties", FieldName: "description"},
	"container":    {AspectName: "editableContainerProperties", FieldName: "description"},
	"dataProduct":  {AspectName: "dataProductProperties", FieldName: "description"},
	"domain":       {AspectName: "domainProperties", FieldName: "description"},
	"glossaryTerm": {AspectName: "glossaryTermInfo", FieldName: "definition"},
	"glossaryNode": {AspectName: "glossaryNodeInfo", FieldName: "definition"},
}

// ErrUnsupportedEntityType is returned when an entity type does not support description updates.
var ErrUnsupportedEntityType = fmt.Errorf("unsupported entity type for description update")

// lookupDescriptionAspect returns the aspect info for updating the description of the given entity type.
func lookupDescriptionAspect(entityType string) (descriptionAspectInfo, error) {
	info, ok := descriptionAspectMap[entityType]
	if !ok {
		return descriptionAspectInfo{}, fmt.Errorf("%w: %s", ErrUnsupportedEntityType, entityType)
	}
	return info, nil
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

// editablePropertiesAspect represents a generic properties aspect for description updates.
// Uses a map to handle different field names ("description" vs "definition") across entity types.
type editablePropertiesAspect struct {
	fields map[string]json.RawMessage
}

// MarshalJSON serializes the aspect as a flat JSON object.
func (a *editablePropertiesAspect) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.fields)
}

// setDescription sets the description value in the aspect under the given field name.
func (a *editablePropertiesAspect) setDescription(fieldName, value string) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encoding description: %w", err)
	}
	a.fields[fieldName] = encoded
	return nil
}

// UpdateDescription sets the editable description for any entity using read-modify-write.
// Resolves the correct aspect name and field name based on the entity type in the URN.
func (c *Client) UpdateDescription(ctx context.Context, urn, description string) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return fmt.Errorf("UpdateDescription: %w", err)
	}

	aspectInfo, err := lookupDescriptionAspect(entityType)
	if err != nil {
		return fmt.Errorf("UpdateDescription: %w", err)
	}

	props, err := c.readEditableProperties(ctx, urn, aspectInfo.AspectName)
	if err != nil {
		return fmt.Errorf("UpdateDescription: %w", err)
	}

	if err := props.setDescription(aspectInfo.FieldName, description); err != nil {
		return fmt.Errorf("UpdateDescription: %w", err)
	}

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: aspectInfo.AspectName,
		Aspect:     props,
	})
}

// readEditableProperties reads the current properties aspect for an entity.
// Returns an empty aspect if none exists (not an error).
func (c *Client) readEditableProperties(ctx context.Context, urn, aspectName string) (*editablePropertiesAspect, error) {
	raw, err := c.getAspect(ctx, urn, aspectName)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &editablePropertiesAspect{fields: map[string]json.RawMessage{}}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", aspectName, err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", aspectName, err)
	}
	return &editablePropertiesAspect{fields: fields}, nil
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
// Per GlossaryTerms.pdl, auditStamp is a required field.
type glossaryTermsAspect struct {
	Terms      []termAssociation `json:"terms"`
	AuditStamp auditStampRaw     `json:"auditStamp"`
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
	terms.AuditStamp = newAuditStamp()

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
	terms.AuditStamp = newAuditStamp()

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
// Per InstitutionalMemoryMetadata.pdl, the audit stamp field is "createStamp".
type linkElement struct {
	URL         string        `json:"url"`
	Description string        `json:"description"`
	CreateStamp auditStampRaw `json:"createStamp"`
}

// auditStampRaw represents an audit stamp with millisecond timestamp.
// Per AuditStamp.pdl: time (epoch ms) and actor (entity URN) are required.
type auditStampRaw struct {
	Time  int64  `json:"time"`
	Actor string `json:"actor"`
}

// newAuditStamp creates an audit stamp with the current time.
func newAuditStamp() auditStampRaw {
	return auditStampRaw{
		Time:  time.Now().UnixMilli(),
		Actor: "urn:li:corpuser:datahub",
	}
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
		CreateStamp: newAuditStamp(),
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
