package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetColumnLineageInput is the input for the deprecated datahub_get_column_lineage tool.
//
// Deprecated: use GetLineageInput with Level="column" instead.
type GetColumnLineageInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the dataset"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

// registerGetColumnLineageTool registers the deprecated datahub_get_column_lineage alias.
// It delegates to handleGetLineage with Level="column".
func (t *Toolkit) registerGetColumnLineageTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		colLineageInput, ok := input.(GetColumnLineageInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		lineageInput := GetLineageInput{
			URN:        colLineageInput.URN,
			Level:      "column",
			Connection: colLineageInput.Connection,
		}
		return t.handleGetLineage(ctx, req, lineageInput)
	}

	wrappedHandler := t.wrapHandler(ToolGetColumnLineage, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolGetColumnLineage),
		Description:  t.getDescription(ToolGetColumnLineage, cfg),
		Annotations:  t.getAnnotations(ToolGetColumnLineage, cfg),
		Icons:        t.getIcons(ToolGetColumnLineage, cfg),
		Title:        t.getTitle(ToolGetColumnLineage, cfg),
		OutputSchema: t.getOutputSchema(ToolGetColumnLineage, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetColumnLineageInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}
