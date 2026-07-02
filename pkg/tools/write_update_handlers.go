package tools

import (
	"context"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// Action constants for update operations.
const (
	actionAdd    = "add"
	actionRemove = "remove"
	actionSet    = "set"
)

func (t *Toolkit) handleUpdateDescription(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if err := c.UpdateDescription(ctx, input.URN, input.Value); err != nil {
		return UpdateOutput{}, err
	}
	aspectName := "unknown"
	if parsed, parseErr := client.ParseURN(input.URN); parseErr == nil {
		if info, lookupErr := client.LookupDescriptionAspect(parsed.EntityType); lookupErr == nil {
			aspectName = info.AspectName
		}
	}
	return UpdateOutput{URN: input.URN, What: "description (" + aspectName + ")", Action: "updated"}, nil
}

func (t *Toolkit) handleUpdateColumnDescription(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if input.FieldPath == "" {
		return UpdateOutput{}, errRequired("field_path")
	}
	if err := c.UpdateColumnDescription(ctx, input.URN, input.FieldPath, input.Value); err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: input.URN, What: "column_description", Action: "updated"}, nil
}

func (t *Toolkit) handleUpdateTag(
	ctx context.Context, c DataHubClient, input UpdateInput, action string,
) (UpdateOutput, error) {
	if action == "" {
		return UpdateOutput{}, errRequired("action (add or remove)")
	}
	if input.TargetURN == "" {
		return UpdateOutput{}, errRequired("target_urn")
	}
	switch action {
	case actionAdd:
		if err := c.AddTag(ctx, input.URN, input.TargetURN); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "tag", Action: "added", TargetURN: input.TargetURN}, nil
	case actionRemove:
		if err := c.RemoveTag(ctx, input.URN, input.TargetURN); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "tag", Action: "removed", TargetURN: input.TargetURN}, nil
	default:
		return UpdateOutput{}, errInvalidAction(action, "add", "remove")
	}
}

func (t *Toolkit) handleUpdateGlossaryTerm(
	ctx context.Context, c DataHubClient, input UpdateInput, action string,
) (UpdateOutput, error) {
	if action == "" {
		return UpdateOutput{}, errRequired("action (add or remove)")
	}
	if input.TargetURN == "" {
		return UpdateOutput{}, errRequired("target_urn")
	}
	switch action {
	case actionAdd:
		if err := c.AddGlossaryTerm(ctx, input.URN, input.TargetURN); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "glossary_term", Action: "added", TargetURN: input.TargetURN}, nil
	case actionRemove:
		if err := c.RemoveGlossaryTerm(ctx, input.URN, input.TargetURN); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "glossary_term", Action: "removed", TargetURN: input.TargetURN}, nil
	default:
		return UpdateOutput{}, errInvalidAction(action, "add", "remove")
	}
}

func (t *Toolkit) handleUpdateLink(
	ctx context.Context, c DataHubClient, input UpdateInput, action string,
) (UpdateOutput, error) {
	if action == "" {
		return UpdateOutput{}, errRequired("action (add or remove)")
	}
	if input.URL == "" {
		return UpdateOutput{}, errRequired("url")
	}
	switch action {
	case actionAdd:
		if err := c.AddLink(ctx, input.URN, input.URL, input.Value); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "link", Action: "added"}, nil
	case actionRemove:
		if err := c.RemoveLink(ctx, input.URN, input.URL); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "link", Action: "removed"}, nil
	default:
		return UpdateOutput{}, errInvalidAction(action, "add", "remove")
	}
}

func (t *Toolkit) handleUpdateOwner(
	ctx context.Context, c DataHubClient, input UpdateInput, action string,
) (UpdateOutput, error) {
	if action == "" {
		return UpdateOutput{}, errRequired("action (add or remove)")
	}
	if input.TargetURN == "" {
		return UpdateOutput{}, errRequired("target_urn (owner URN)")
	}
	switch action {
	case actionAdd:
		ownerType := input.OwnershipType
		if ownerType == "" {
			ownerType = "TECHNICAL_OWNER"
		}
		if err := c.AddOwner(ctx, input.URN, input.TargetURN, ownerType); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "owner", Action: "added", TargetURN: input.TargetURN}, nil
	case actionRemove:
		if err := c.RemoveOwner(ctx, input.URN, input.TargetURN); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "owner", Action: "removed", TargetURN: input.TargetURN}, nil
	default:
		return UpdateOutput{}, errInvalidAction(action, "add", "remove")
	}
}

func (t *Toolkit) handleUpdateDomain(
	ctx context.Context, c DataHubClient, input UpdateInput, action string,
) (UpdateOutput, error) {
	if action == "" {
		action = actionSet
	}
	switch action {
	case actionSet:
		if input.TargetURN == "" {
			return UpdateOutput{}, errRequired("target_urn (domain URN)")
		}
		if err := c.SetDomain(ctx, input.URN, input.TargetURN); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "domain", Action: "set", TargetURN: input.TargetURN}, nil
	case actionRemove:
		if err := c.UnsetDomain(ctx, input.URN); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "domain", Action: "removed"}, nil
	default:
		return UpdateOutput{}, errInvalidAction(action, "set", "remove")
	}
}

func (t *Toolkit) handleUpdateStructuredProperties(
	ctx context.Context, c DataHubClient, input UpdateInput, action string,
) (UpdateOutput, error) {
	if action == "" {
		action = actionSet
	}
	switch action {
	case actionSet:
		if len(input.Properties) == 0 {
			return UpdateOutput{}, errRequired("properties")
		}
		if err := c.UpsertStructuredProperties(ctx, input.URN, input.Properties); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "structured_properties", Action: "updated"}, nil
	case actionRemove:
		if len(input.PropertyURNs) == 0 {
			return UpdateOutput{}, errRequired("property_urns")
		}
		if err := c.RemoveStructuredProperties(ctx, input.URN, input.PropertyURNs); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "structured_properties", Action: "removed"}, nil
	default:
		return UpdateOutput{}, errInvalidAction(action, "set", "remove")
	}
}

func (t *Toolkit) handleUpdateCustomProperties(
	ctx context.Context, c DataHubClient, input UpdateInput, action string,
) (UpdateOutput, error) {
	if action == "" {
		action = actionSet
	}
	switch action {
	case actionSet:
		if len(input.CustomProperties) == 0 {
			return UpdateOutput{}, errRequired("custom_properties")
		}
		if err := c.SetCustomProperties(ctx, input.URN, input.CustomProperties); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "custom_properties", Action: "updated"}, nil
	case actionRemove:
		if len(input.PropertyKeys) == 0 {
			return UpdateOutput{}, errRequired("property_keys")
		}
		if err := c.RemoveCustomProperties(ctx, input.URN, input.PropertyKeys); err != nil {
			return UpdateOutput{}, err
		}
		return UpdateOutput{URN: input.URN, What: "custom_properties", Action: "removed"}, nil
	default:
		return UpdateOutput{}, errInvalidAction(action, "set", "remove")
	}
}

func (t *Toolkit) handleUpdateStructuredProperty(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if err := c.UpdateStructuredProperty(ctx, input.URN, types.UpdateStructuredPropertyInput{
		DisplayName: input.Name,
		Description: input.Description,
	}); err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: input.URN, What: "structured_property", Action: "updated"}, nil
}

func (t *Toolkit) handleUpdateIncidentStatus(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if input.State == "" {
		return UpdateOutput{}, errRequired("state (ACTIVE or RESOLVED)")
	}
	if err := c.UpdateIncidentStatus(ctx, input.URN, input.State, input.Value); err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: input.URN, What: "incident_status", Action: "updated to " + input.State}, nil
}

func (t *Toolkit) handleUpdateIncident(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if err := c.UpdateIncident(ctx, input.URN, types.UpdateIncidentInput{
		Title:       input.Name,
		Description: input.Description,
		Priority:    input.Priority,
	}); err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: input.URN, What: "incident", Action: "updated"}, nil
}

func (t *Toolkit) handleUpdateQuery(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if _, err := c.UpdateQuery(ctx, client.UpdateQueryInput{
		URN:         input.URN,
		Name:        input.Name,
		Description: input.Description,
		Statement:   input.Value,
		Language:    input.Language,
		DatasetURNs: input.DatasetURNs,
	}); err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: input.URN, What: "query", Action: "updated"}, nil
}

func (t *Toolkit) handleUpdateDocumentContents(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if err := c.UpdateDocumentContents(ctx, input.URN, input.Title, input.Text); err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: input.URN, What: "document_contents", Action: "updated"}, nil
}

func (t *Toolkit) handleUpdateDocumentStatus(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if input.Value == "" {
		return UpdateOutput{}, errRequired("value (status)")
	}
	if err := c.UpdateDocumentStatus(ctx, input.URN, input.Value); err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: input.URN, What: "document_status", Action: "updated"}, nil
}

func (t *Toolkit) handleUpdateDocumentRelatedEntities(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if err := c.UpdateDocumentRelatedEntities(ctx, input.URN, input.EntityURNs); err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: input.URN, What: "document_related_entities", Action: "updated"}, nil
}

func (t *Toolkit) handleUpdateDocumentSubType(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	if input.Value == "" {
		return UpdateOutput{}, errRequired("value (sub_type)")
	}
	if err := c.UpdateDocumentSubType(ctx, input.URN, input.Value); err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: input.URN, What: "document_sub_type", Action: "updated"}, nil
}

func (t *Toolkit) handleUpdateDataContract(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	urn, err := c.UpsertDataContract(ctx, types.UpsertDataContractInput{
		DatasetURN:               input.URN,
		SchemaAssertionURNs:      input.SchemaAssertions,
		FreshnessAssertionURNs:   input.FreshnessAssertions,
		DataQualityAssertionURNs: input.DataQualityAssertions,
	})
	if err != nil {
		return UpdateOutput{}, err
	}
	return UpdateOutput{URN: urn, What: "data_contract", Action: "updated"}, nil
}
