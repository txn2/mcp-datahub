package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// UpdateInput is the input for the datahub_update tool.
type UpdateInput struct {
	//nolint:lll // struct tag cannot be split
	What string `json:"what" jsonschema_description:"What to update: description, column_description, tag, glossary_term, link, owner, domain, structured_properties, structured_property, incident_status, incident, query, document_contents, document_status, document_related_entities, document_sub_type, or data_contract" jsonschema_enum:"description,column_description,tag,glossary_term,link,owner,domain,structured_properties,structured_property,incident_status,incident,query,document_contents,document_status,document_related_entities,document_sub_type,data_contract"`

	//nolint:lll // struct tag cannot be split
	Action string `json:"action,omitempty" jsonschema_description:"Action: add/remove (tag, glossary_term, link, owner), set/remove (domain, structured_properties), not used for other what values" jsonschema_enum:"add,remove,set"`

	// Entity identification
	URN string `json:"urn" jsonschema_description:"URN of the entity to update"`

	// Common value field
	Value string `json:"value,omitempty" jsonschema_description:"New value (description text, status, sub_type, label, message, etc.)"`

	// Target URN for add/remove operations (tag, glossary_term, owner, domain)
	TargetURN string `json:"target_urn,omitempty" jsonschema_description:"Target URN (tag, glossary term, owner, or domain URN)"`

	// Link-specific fields
	URL string `json:"url,omitempty" jsonschema_description:"URL for link operations"`

	// Column description
	FieldPath string `json:"field_path,omitempty" jsonschema_description:"Schema field path (column_description only)"`

	// Owner-specific
	OwnershipType string `json:"ownership_type,omitempty" jsonschema_description:"Ownership type (owner add only, e.g. TECHNICAL_OWNER)"`

	// Structured properties on assets
	Properties   []types.StructuredPropertyInput `json:"properties,omitempty" jsonschema_description:"Structured property values to set"`
	PropertyURNs []string                        `json:"property_urns,omitempty" jsonschema_description:"Property URNs to remove"`

	// Query-specific
	Name        string   `json:"name,omitempty" jsonschema_description:"Updated name (query, incident, structured_property)"`
	Description string   `json:"description,omitempty" jsonschema_description:"Updated description"`
	Language    string   `json:"language,omitempty" jsonschema_description:"Query language (query only)"`
	DatasetURNs []string `json:"dataset_urns,omitempty" jsonschema_description:"Dataset URNs (query, data_contract)"`

	// Incident-specific
	IncidentType string `json:"incident_type,omitempty" jsonschema_description:"Incident type (incident only)"`
	Priority     string `json:"priority,omitempty" jsonschema_description:"Priority: LOW, MEDIUM, HIGH, CRITICAL (incident only)"`
	State        string `json:"state,omitempty" jsonschema_description:"Incident state: ACTIVE, RESOLVED (incident_status only)"`

	// Document-specific
	Title      string   `json:"title,omitempty" jsonschema_description:"Document title (document_contents only)"`
	Text       string   `json:"text,omitempty" jsonschema_description:"Document text (document_contents only)"`
	EntityURNs []string `json:"entity_urns,omitempty" jsonschema_description:"Related entity URNs (document_related_entities only)"`

	// Data contract fields
	SchemaAssertions []string `json:"schema_assertions,omitempty" jsonschema_description:"Schema assertion URNs"`
	//nolint:lll // struct tag
	FreshnessAssertions   []string `json:"freshness_assertions,omitempty" jsonschema_description:"Freshness assertion URNs (data_contract only)"`
	DataQualityAssertions []string `json:"data_quality_assertions,omitempty" jsonschema_description:"Data quality assertion URNs"`

	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerUpdateTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		updateInput, ok := input.(UpdateInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleUpdate(ctx, req, updateInput)
	}

	wrappedHandler := t.wrapHandler(ToolUpdate, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolUpdate),
		Description:  t.getDescription(ToolUpdate, cfg),
		Annotations:  t.getAnnotations(ToolUpdate, cfg),
		Icons:        t.getIcons(ToolUpdate, cfg),
		Title:        t.getTitle(ToolUpdate, cfg),
		OutputSchema: t.getOutputSchema(ToolUpdate, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateInput) (*mcp.CallToolResult, *UpdateOutput, error) {
		result, out, err := wrappedHandler(ctx, req, input)
		if typed, ok := out.(*UpdateOutput); ok {
			return result, typed, err
		}
		return result, nil, err
	})
}

func (t *Toolkit) handleUpdate(
	ctx context.Context, _ *mcp.CallToolRequest, input UpdateInput,
) (*mcp.CallToolResult, any, error) {
	if input.What == "" {
		return ErrorResult("what parameter is required"), nil, nil
	}
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	action := input.Action

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	output, updateErr := t.dispatchUpdate(ctx, datahubClient, input, action)
	if updateErr != nil {
		return ErrorResult("Update " + input.What + " failed: " + updateErr.Error()), nil, nil
	}

	jsonResult, jsonErr := JSONResult(output)
	if jsonErr != nil {
		return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
	}
	return jsonResult, &output, nil
}

func (t *Toolkit) dispatchUpdate(
	ctx context.Context, c DataHubClient, input UpdateInput, action string,
) (UpdateOutput, error) {
	// Metadata operations (add/remove/set on tags, terms, links, owners, domains, structured props)
	if out, ok := t.dispatchUpdateMetadata(ctx, c, input, action); ok {
		return out.output, out.err
	}

	// Entity and resource operations
	return t.dispatchUpdateEntity(ctx, c, input)
}

// updateResult bundles the output and error from an update operation.
type updateResult struct {
	output UpdateOutput
	err    error
}

// dispatchUpdateMetadata handles add/remove/set operations on entity metadata.
// Returns the result and true if handled, or zero value and false to continue dispatch.
func (t *Toolkit) dispatchUpdateMetadata(
	ctx context.Context, c DataHubClient, input UpdateInput, action string,
) (updateResult, bool) {
	switch input.What {
	case "tag":
		out, err := t.handleUpdateTag(ctx, c, input, action)
		return updateResult{out, err}, true
	case "glossary_term":
		out, err := t.handleUpdateGlossaryTerm(ctx, c, input, action)
		return updateResult{out, err}, true
	case "link":
		out, err := t.handleUpdateLink(ctx, c, input, action)
		return updateResult{out, err}, true
	case "owner":
		out, err := t.handleUpdateOwner(ctx, c, input, action)
		return updateResult{out, err}, true
	case "domain":
		out, err := t.handleUpdateDomain(ctx, c, input, action)
		return updateResult{out, err}, true
	case "structured_properties":
		out, err := t.handleUpdateStructuredProperties(ctx, c, input, action)
		return updateResult{out, err}, true
	default:
		return updateResult{}, false
	}
}

func (t *Toolkit) dispatchUpdateEntity(
	ctx context.Context, c DataHubClient, input UpdateInput,
) (UpdateOutput, error) {
	switch input.What {
	case "description":
		return t.handleUpdateDescription(ctx, c, input)
	case "column_description":
		return t.handleUpdateColumnDescription(ctx, c, input)
	case "structured_property":
		return t.handleUpdateStructuredProperty(ctx, c, input)
	case "incident_status":
		return t.handleUpdateIncidentStatus(ctx, c, input)
	case "incident":
		return t.handleUpdateIncident(ctx, c, input)
	case "query":
		return t.handleUpdateQuery(ctx, c, input)
	case "document_contents":
		return t.handleUpdateDocumentContents(ctx, c, input)
	case "document_status":
		return t.handleUpdateDocumentStatus(ctx, c, input)
	case "document_related_entities":
		return t.handleUpdateDocumentRelatedEntities(ctx, c, input)
	case "document_sub_type":
		return t.handleUpdateDocumentSubType(ctx, c, input)
	case "data_contract":
		return t.handleUpdateDataContract(ctx, c, input)
	default:
		return UpdateOutput{}, fmt.Errorf("unsupported what value: %s", input.What)
	}
}
