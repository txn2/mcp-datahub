package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// customPropertiesAspectMap maps DataHub entity types to the aspect that carries
// their legacy customProperties map. These are the non-editable "properties"/"info"
// aspects that mix in CustomProperties (verified against the upstream DataHub PDL
// models). tag is intentionally absent: tagProperties has no customProperties field.
//
// Writes use REST read-modify-write (getAspect + postIngestProposal) because DataHub
// exposes no GraphQL mutation for arbitrary customProperties. All existing aspect
// fields are preserved via descriptionAspect's map representation, so required fields
// on these aspects (for example glossaryTermInfo.termSource) survive the round-trip.
var customPropertiesAspectMap = map[string]string{
	entityTypeDataset:      "datasetProperties",
	entityTypeDashboard:    "dashboardInfo",
	entityTypeChart:        "chartInfo",
	entityTypeDataFlow:     "dataFlowInfo",
	entityTypeDataJob:      "dataJobInfo",
	entityTypeContainer:    "containerProperties",
	entityTypeDataProduct:  "dataProductProperties",
	entityTypeGlossaryNode: "glossaryNodeInfo",
	entityTypeGlossaryTerm: "glossaryTermInfo",
	entityTypeDomain:       "domainProperties",
}

// customProperties extracts the customProperties map from the aspect.
// Returns an empty map when the field is absent or null.
func (a *descriptionAspect) customProperties() (map[string]string, error) {
	raw, ok := a.fields["customProperties"]
	if !ok || isNullOrEmptyJSON(raw) {
		return map[string]string{}, nil
	}

	var props map[string]string
	if err := json.Unmarshal(raw, &props); err != nil {
		return nil, fmt.Errorf("parsing customProperties: %w", err)
	}
	if props == nil {
		props = map[string]string{}
	}
	return props, nil
}

// setCustomProperties writes the customProperties map back into the aspect,
// preserving all other fields of the read-modify-write cycle.
func (a *descriptionAspect) setCustomProperties(props map[string]string) error {
	encoded, err := json.Marshal(props)
	if err != nil {
		return fmt.Errorf("encoding customProperties: %w", err)
	}
	a.fields["customProperties"] = encoded
	return nil
}

// SetCustomProperties sets or overwrites the given legacy customProperties key/values
// on an entity. Keys not listed in properties are left untouched. This is the legacy
// customProperties map, distinct from structured properties (see UpsertStructuredProperties).
func (c *Client) SetCustomProperties(ctx context.Context, urn string, properties map[string]string) error {
	if err := c.mutateCustomProperties(ctx, urn, func(current map[string]string) {
		for key, value := range properties {
			current[key] = value
		}
	}); err != nil {
		return fmt.Errorf("SetCustomProperties: %w", err)
	}
	return nil
}

// RemoveCustomProperties removes the given keys from an entity's legacy customProperties
// map. Keys that are not present are ignored.
func (c *Client) RemoveCustomProperties(ctx context.Context, urn string, keys []string) error {
	if err := c.mutateCustomProperties(ctx, urn, func(current map[string]string) {
		for _, key := range keys {
			delete(current, key)
		}
	}); err != nil {
		return fmt.Errorf("RemoveCustomProperties: %w", err)
	}
	return nil
}

// mutateCustomProperties applies mutate to an entity's current customProperties map
// using REST read-modify-write, then writes the properties aspect back.
func (c *Client) mutateCustomProperties(
	ctx context.Context, urn string, mutate func(map[string]string),
) error {
	entityType, err := entityTypeFromURN(urn)
	if err != nil {
		return err
	}

	aspectName, ok := customPropertiesAspectMap[entityType]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnsupportedCustomPropertiesEntity, entityType)
	}

	props, err := c.readEditableProperties(ctx, entityType, urn, aspectName)
	if err != nil {
		return err
	}

	// The REST GET of dataProductProperties omits the display name; preserve it so
	// the write does not reset the name to the URN slug (mirrors UpdateDescription).
	if entityType == entityTypeDataProduct {
		if err = c.preserveDataProductName(ctx, urn, props); err != nil {
			return err
		}
	}

	current, err := props.customProperties()
	if err != nil {
		return err
	}

	mutate(current)

	if err := props.setCustomProperties(current); err != nil {
		return err
	}

	return c.postIngestProposal(ctx, ingestProposal{
		EntityType: entityType,
		EntityURN:  urn,
		AspectName: aspectName,
		Aspect:     props,
	})
}
