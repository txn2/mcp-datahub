package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// CreateInput is the input for the datahub_create tool.
type CreateInput struct {
	//nolint:lll // struct tag cannot be split
	What string `json:"what" jsonschema_description:"Entity type to create: tag, domain, glossary_term, data_product, document, application, query, incident, structured_property, or data_contract" jsonschema_enum:"tag,domain,glossary_term,data_product,document,application,query,incident,structured_property,data_contract"`

	// Common fields
	Name        string `json:"name,omitempty" jsonschema_description:"Name or title of the entity"`
	Description string `json:"description,omitempty" jsonschema_description:"Description or content of the entity"`

	// Glossary term fields
	ParentNode string `json:"parent_node,omitempty" jsonschema_description:"Parent glossary node URN (glossary_term only)"`

	// Data product fields
	DomainURN string `json:"domain_urn,omitempty" jsonschema_description:"Domain URN (data_product only, required)"`

	// Query fields
	Value       string   `json:"value,omitempty" jsonschema_description:"SQL statement (query only)"`
	Language    string   `json:"language,omitempty" jsonschema_description:"Query language, default SQL (query only)"`
	DatasetURNs []string `json:"dataset_urns,omitempty" jsonschema_description:"Associated dataset URNs (query, data_contract)"`

	// Incident fields
	EntityURNs   []string `json:"entity_urns,omitempty" jsonschema_description:"Affected entity URNs (incident only)"`
	IncidentType string   `json:"incident_type,omitempty" jsonschema_description:"Incident type: OPERATIONAL, CUSTOM, etc. (incident only)"`
	Priority     string   `json:"priority,omitempty" jsonschema_description:"Priority: LOW, MEDIUM, HIGH, CRITICAL (incident only)"`

	// Document fields
	Status        string   `json:"status,omitempty" jsonschema_description:"Publication status: PUBLISHED or UNPUBLISHED (document only)"`
	SubType       string   `json:"sub_type,omitempty" jsonschema_description:"Document sub-type (document only)"`
	RelatedAssets []string `json:"related_assets,omitempty" jsonschema_description:"Related asset URNs (document only)"`
	GlobalContext bool     `json:"global_context,omitempty" jsonschema_description:"Show in global search (document only)"`

	// Structured property fields
	QualifiedName string   `json:"qualified_name,omitempty" jsonschema_description:"Fully qualified name (structured_property)"`
	ValueType     string   `json:"value_type,omitempty" jsonschema_description:"Value type: string, number, date, urn (structured_property)"`
	Cardinality   string   `json:"cardinality,omitempty" jsonschema_description:"SINGLE or MULTIPLE (structured_property only)"`
	EntityTypes   []string `json:"entity_types,omitempty" jsonschema_description:"Applicable entity types (structured_property only)"`

	// Data contract fields
	SchemaAssertions      []string `json:"schema_assertions,omitempty" jsonschema_description:"Schema assertion URNs"`
	FreshnessAssertions   []string `json:"freshness_assertions,omitempty" jsonschema_description:"Freshness assertion URNs"`
	DataQualityAssertions []string `json:"data_quality_assertions,omitempty" jsonschema_description:"Data quality assertion URNs"`

	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerCreateTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		createInput, ok := input.(CreateInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleCreate(ctx, req, createInput)
	}

	wrappedHandler := t.wrapHandler(ToolCreate, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolCreate),
		Description:  t.getDescription(ToolCreate, cfg),
		Annotations:  t.getAnnotations(ToolCreate, cfg),
		Icons:        t.getIcons(ToolCreate, cfg),
		Title:        t.getTitle(ToolCreate, cfg),
		OutputSchema: t.getOutputSchema(ToolCreate, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateInput) (*mcp.CallToolResult, *CreateOutput, error) {
		result, out, err := wrappedHandler(ctx, req, input)
		if typed, ok := out.(*CreateOutput); ok {
			return result, typed, err
		}
		return result, nil, err
	})
}

func (t *Toolkit) handleCreate(
	ctx context.Context, _ *mcp.CallToolRequest, input CreateInput,
) (*mcp.CallToolResult, any, error) {
	if input.What == "" {
		return ErrorResult("what parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	var urn string

	switch input.What {
	case "tag":
		urn, err = t.handleCreateTag(ctx, datahubClient, input)
	case "domain":
		urn, err = t.handleCreateDomain(ctx, datahubClient, input)
	case "glossary_term":
		urn, err = t.handleCreateGlossaryTerm(ctx, datahubClient, input)
	case "data_product":
		urn, err = t.handleCreateDataProduct(ctx, datahubClient, input)
	case "document":
		urn, err = t.handleCreateDocument(ctx, datahubClient, input)
	case "application":
		urn, err = t.handleCreateApplication(ctx, datahubClient, input)
	case "query":
		urn, err = t.handleCreateQuery(ctx, datahubClient, input)
	case "incident":
		urn, err = t.handleCreateIncident(ctx, datahubClient, input)
	case "structured_property":
		urn, err = t.handleCreateStructuredProperty(ctx, datahubClient, input)
	case "data_contract":
		urn, err = t.handleCreateDataContract(ctx, datahubClient, input)
	default:
		return ErrorResult("invalid what value: " + input.What), nil, nil
	}

	if err != nil {
		return ErrorResult("Create " + input.What + " failed: " + err.Error()), nil, nil
	}

	output := CreateOutput{URN: urn, What: input.What, Action: "created"}
	jsonResult, jsonErr := JSONResult(output)
	if jsonErr != nil {
		return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
	}
	return jsonResult, &output, nil
}

func (t *Toolkit) handleCreateTag(ctx context.Context, c DataHubClient, input CreateInput) (string, error) {
	if input.Name == "" {
		return "", errRequired("name")
	}
	return c.CreateTag(ctx, input.Name, input.Description)
}

func (t *Toolkit) handleCreateDomain(ctx context.Context, c DataHubClient, input CreateInput) (string, error) {
	if input.Name == "" {
		return "", errRequired("name")
	}
	return c.CreateDomain(ctx, input.Name, input.Description)
}

func (t *Toolkit) handleCreateGlossaryTerm(ctx context.Context, c DataHubClient, input CreateInput) (string, error) {
	if input.Name == "" {
		return "", errRequired("name")
	}
	return c.CreateGlossaryTerm(ctx, input.Name, input.Description, input.ParentNode)
}

func (t *Toolkit) handleCreateDataProduct(ctx context.Context, c DataHubClient, input CreateInput) (string, error) {
	if input.Name == "" {
		return "", errRequired("name")
	}
	if input.DomainURN == "" {
		return "", errRequired("domain_urn")
	}
	return c.CreateDataProduct(ctx, input.Name, input.Description, input.DomainURN)
}

func (t *Toolkit) handleCreateDocument(ctx context.Context, c DataHubClient, input CreateInput) (string, error) {
	if input.Name == "" {
		return "", errRequired("name")
	}
	return c.CreateDocument(ctx, types.CreateDocumentInput{
		Title:            input.Name,
		Content:          input.Description,
		Status:           input.Status,
		SubType:          input.SubType,
		RelatedAssetURNs: input.RelatedAssets,
		GlobalContext:    input.GlobalContext,
	})
}

func (t *Toolkit) handleCreateApplication(ctx context.Context, c DataHubClient, input CreateInput) (string, error) {
	if input.Name == "" {
		return "", errRequired("name")
	}
	return c.CreateApplication(ctx, input.Name, input.Description)
}

func (t *Toolkit) handleCreateQuery(ctx context.Context, c DataHubClient, input CreateInput) (string, error) {
	if input.Value == "" {
		return "", errRequired("value")
	}
	q, qErr := c.CreateQuery(ctx, client.CreateQueryInput{
		Name:        input.Name,
		Description: input.Description,
		Statement:   input.Value,
		Language:    input.Language,
		DatasetURNs: input.DatasetURNs,
	})
	if qErr != nil {
		return "", qErr
	}
	return q.URN, nil
}

func (t *Toolkit) handleCreateIncident(ctx context.Context, c DataHubClient, input CreateInput) (string, error) {
	if len(input.EntityURNs) == 0 {
		return "", errRequired("entity_urns")
	}
	if input.Name == "" {
		return "", errRequired("name (title)")
	}
	if input.IncidentType == "" {
		return "", errRequired("incident_type")
	}
	return c.RaiseIncident(ctx, types.RaiseIncidentInput{
		Type:         input.IncidentType,
		Title:        input.Name,
		Description:  input.Description,
		Priority:     input.Priority,
		ResourceURNs: input.EntityURNs,
	})
}

func (t *Toolkit) handleCreateStructuredProperty(
	ctx context.Context, c DataHubClient, input CreateInput,
) (string, error) {
	if input.QualifiedName == "" {
		return "", errRequired("qualified_name")
	}
	if input.ValueType == "" {
		return "", errRequired("value_type")
	}
	if len(input.EntityTypes) == 0 {
		return "", errRequired("entity_types")
	}
	return c.CreateStructuredProperty(ctx, types.CreateStructuredPropertyInput{
		QualifiedName: input.QualifiedName,
		DisplayName:   input.Name,
		Description:   input.Description,
		ValueType:     input.ValueType,
		Cardinality:   input.Cardinality,
		EntityTypes:   input.EntityTypes,
	})
}

func (t *Toolkit) handleCreateDataContract(
	ctx context.Context, c DataHubClient, input CreateInput,
) (string, error) {
	if len(input.DatasetURNs) == 0 {
		return "", errRequired("dataset_urns")
	}
	return c.UpsertDataContract(ctx, types.UpsertDataContractInput{
		DatasetURN:               input.DatasetURNs[0],
		SchemaAssertionURNs:      input.SchemaAssertions,
		FreshnessAssertionURNs:   input.FreshnessAssertions,
		DataQualityAssertionURNs: input.DataQualityAssertions,
	})
}
