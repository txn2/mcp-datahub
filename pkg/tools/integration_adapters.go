package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/integration"
)

// Context keys for inter-middleware communication.
const (
	ContextKeyResolvedURN = "resolved_urn"
	ContextKeyAccessOK    = "access_ok"
)

// ErrAccessDenied is returned when access to a URN is denied.
var ErrAccessDenied = errors.New("access denied")

// URNResolverMiddleware wraps a URNResolver to resolve external IDs to DataHub URNs.
type URNResolverMiddleware struct {
	resolver integration.URNResolver
}

// NewURNResolverMiddleware creates a middleware that resolves external IDs to DataHub URNs.
func NewURNResolverMiddleware(r integration.URNResolver) *URNResolverMiddleware {
	return &URNResolverMiddleware{resolver: r}
}

// Before implements ToolMiddleware.
func (m *URNResolverMiddleware) Before(ctx context.Context, tc *ToolContext) (context.Context, error) {
	urn := extractURNFromInput(tc.Input)
	if urn == "" {
		return ctx, nil
	}

	// Skip resolution if already a DataHub URN
	if isDataHubURN(urn) {
		tc.Set(ContextKeyResolvedURN, urn)
		return ctx, nil
	}

	// Resolve external ID to DataHub URN
	resolved, err := m.resolver.ResolveToDataHubURN(ctx, urn)
	if err != nil {
		return ctx, err
	}

	tc.Set(ContextKeyResolvedURN, resolved)
	return ctx, nil
}

// After implements ToolMiddleware (no-op).
func (m *URNResolverMiddleware) After(_ context.Context, _ *ToolContext, result *mcp.CallToolResult, _ error) (*mcp.CallToolResult, error) {
	return result, nil
}

// AccessFilterMiddleware wraps an AccessFilter to control access to entities.
type AccessFilterMiddleware struct {
	filter integration.AccessFilter
}

// NewAccessFilterMiddleware creates a middleware that controls access to entities.
func NewAccessFilterMiddleware(f integration.AccessFilter) *AccessFilterMiddleware {
	return &AccessFilterMiddleware{filter: f}
}

// Before implements ToolMiddleware.
func (m *AccessFilterMiddleware) Before(ctx context.Context, tc *ToolContext) (context.Context, error) {
	urn := getEffectiveURN(tc)
	if urn == "" {
		return ctx, nil
	}

	// Check access
	allowed, err := m.filter.CanAccess(ctx, urn)
	if err != nil {
		return ctx, err
	}

	if !allowed {
		return ctx, ErrAccessDenied
	}

	tc.Set(ContextKeyAccessOK, true)
	return ctx, nil
}

// After implements ToolMiddleware - filters URNs from list results.
func (m *AccessFilterMiddleware) After(
	ctx context.Context,
	tc *ToolContext,
	result *mcp.CallToolResult,
	_ error,
) (*mcp.CallToolResult, error) {
	// Only filter list/search results
	if !isListTool(tc.ToolName) {
		return result, nil
	}

	if result == nil || result.IsError || len(result.Content) == 0 {
		return result, nil
	}

	// Extract and filter URNs from result
	filtered, err := m.filterResultURNs(ctx, result)
	if err != nil {
		// Log error but don't fail - return unfiltered result
		return result, nil
	}

	return filtered, nil
}

// filterResultURNs filters URNs in the result based on access permissions.
func (m *AccessFilterMiddleware) filterResultURNs(ctx context.Context, result *mcp.CallToolResult) (*mcp.CallToolResult, error) {
	// Parse result to map
	data, err := parseResultToMap(result)
	if err != nil {
		return result, err
	}

	// Extract URNs
	urns := extractURNsFromData(data)
	if len(urns) == 0 {
		return result, nil
	}

	// Filter URNs
	allowed, err := m.filter.FilterURNs(ctx, urns)
	if err != nil {
		return result, err
	}

	// Create allowed set for fast lookup
	allowedSet := make(map[string]bool)
	for _, urn := range allowed {
		allowedSet[urn] = true
	}

	// Filter the data
	filtered := filterDataByURNs(data, allowedSet)

	// Re-encode
	jsonData, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return result, err
	}

	return TextResult(string(jsonData)), nil
}

// AuditLoggerMiddleware wraps an AuditLogger to log tool invocations.
type AuditLoggerMiddleware struct {
	logger    integration.AuditLogger
	getUserID func(context.Context) string
}

// NewAuditLoggerMiddleware creates a middleware that logs tool invocations.
func NewAuditLoggerMiddleware(l integration.AuditLogger, getUserID func(context.Context) string) *AuditLoggerMiddleware {
	return &AuditLoggerMiddleware{
		logger:    l,
		getUserID: getUserID,
	}
}

// Before implements ToolMiddleware (no-op).
func (m *AuditLoggerMiddleware) Before(ctx context.Context, _ *ToolContext) (context.Context, error) {
	return ctx, nil
}

// After implements ToolMiddleware - logs the tool call.
func (m *AuditLoggerMiddleware) After(
	ctx context.Context,
	tc *ToolContext,
	result *mcp.CallToolResult,
	_ error,
) (*mcp.CallToolResult, error) {
	var userID string
	if m.getUserID != nil {
		userID = m.getUserID(ctx)
	}

	params := inputToMap(tc.Input)

	// Log async - don't block on logging failures
	go func() {
		_ = m.logger.LogToolCall(ctx, string(tc.ToolName), params, userID)
	}()

	return result, nil
}

// MetadataEnricherMiddleware wraps a MetadataEnricher to add custom metadata.
type MetadataEnricherMiddleware struct {
	enricher integration.MetadataEnricher
}

// NewMetadataEnricherMiddleware creates a middleware that enriches entity responses.
func NewMetadataEnricherMiddleware(e integration.MetadataEnricher) *MetadataEnricherMiddleware {
	return &MetadataEnricherMiddleware{enricher: e}
}

// Before implements ToolMiddleware (no-op).
func (m *MetadataEnricherMiddleware) Before(ctx context.Context, _ *ToolContext) (context.Context, error) {
	return ctx, nil
}

// After implements ToolMiddleware - enriches entity results.
func (m *MetadataEnricherMiddleware) After(
	ctx context.Context,
	tc *ToolContext,
	result *mcp.CallToolResult,
	_ error,
) (*mcp.CallToolResult, error) {
	// Only enrich single-entity results
	if !isEntityTool(tc.ToolName) {
		return result, nil
	}

	if result == nil || result.IsError || len(result.Content) == 0 {
		return result, nil
	}

	urn := getEffectiveURN(tc)
	if urn == "" {
		return result, nil
	}

	// Parse result to map
	data, err := parseResultToMap(result)
	if err != nil {
		return result, nil
	}

	// Enrich the data
	enriched, err := m.enricher.EnrichEntity(ctx, urn, data)
	if err != nil {
		// Log error but don't fail - return original result
		return result, nil
	}

	// Re-encode
	jsonData, err := json.MarshalIndent(enriched, "", "  ")
	if err != nil {
		return result, nil
	}

	return TextResult(string(jsonData)), nil
}

// Helper functions

// extractURNFromInput extracts a URN from various input types.
func extractURNFromInput(input any) string {
	switch v := input.(type) {
	case GetEntityInput:
		return v.URN
	case GetSchemaInput:
		return v.URN
	case GetLineageInput:
		return v.URN
	case GetQueriesInput:
		return v.URN
	case GetGlossaryTermInput:
		return v.URN
	case GetDataProductInput:
		return v.URN
	default:
		return ""
	}
}

// isDataHubURN checks if a string is a DataHub URN.
func isDataHubURN(s string) bool {
	return strings.HasPrefix(s, "urn:li:")
}

// getEffectiveURN returns the resolved URN if available, otherwise the input URN.
func getEffectiveURN(tc *ToolContext) string {
	if resolved, ok := tc.Get(ContextKeyResolvedURN); ok {
		if s, ok := resolved.(string); ok {
			return s
		}
	}
	return extractURNFromInput(tc.Input)
}

// isListTool returns true for tools that return lists of entities.
func isListTool(name ToolName) bool {
	//nolint:exhaustive // Only list-type tools return true
	switch name {
	case ToolSearch, ToolListTags, ToolListDomains, ToolListDataProducts, ToolGetLineage:
		return true
	default:
		return false
	}
}

// isEntityTool returns true for tools that return a single entity.
func isEntityTool(name ToolName) bool {
	//nolint:exhaustive // Only entity-type tools return true
	switch name {
	case ToolGetEntity, ToolGetSchema, ToolGetGlossaryTerm, ToolGetDataProduct:
		return true
	default:
		return false
	}
}

// parseResultToMap parses a result's text content to a map.
func parseResultToMap(result *mcp.CallToolResult) (map[string]any, error) {
	if len(result.Content) == 0 {
		return nil, errors.New("no content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return nil, errors.New("not text content")
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &data); err != nil {
		return nil, err
	}

	return data, nil
}

// extractURNsFromData extracts all URNs from a result map.
func extractURNsFromData(data map[string]any) []string {
	var urns []string

	// Check for "entities" array (search results)
	if entities, ok := data["entities"].([]any); ok {
		for _, e := range entities {
			if entity, ok := e.(map[string]any); ok {
				if urn, ok := entity["urn"].(string); ok {
					urns = append(urns, urn)
				}
			}
		}
	}

	// Check for "nodes" array (lineage results)
	if nodes, ok := data["nodes"].([]any); ok {
		for _, n := range nodes {
			if node, ok := n.(map[string]any); ok {
				if urn, ok := node["urn"].(string); ok {
					urns = append(urns, urn)
				}
			}
		}
	}

	// Check for direct "urn" field
	if urn, ok := data["urn"].(string); ok {
		urns = append(urns, urn)
	}

	return urns
}

// filterDataByURNs filters entities/nodes by allowed URNs.
func filterDataByURNs(data map[string]any, allowed map[string]bool) map[string]any {
	result := make(map[string]any)
	for k, v := range data {
		result[k] = v
	}

	// Filter "entities" array
	if entities, ok := data["entities"].([]any); ok {
		filtered := make([]any, 0)
		for _, e := range entities {
			if entity, ok := e.(map[string]any); ok {
				if urn, ok := entity["urn"].(string); ok && allowed[urn] {
					filtered = append(filtered, entity)
				}
			}
		}
		result["entities"] = filtered
		// Update count if present
		if _, ok := result["total"]; ok {
			result["total"] = len(filtered)
		}
	}

	// Filter "nodes" array
	if nodes, ok := data["nodes"].([]any); ok {
		filtered := make([]any, 0)
		for _, n := range nodes {
			if node, ok := n.(map[string]any); ok {
				if urn, ok := node["urn"].(string); ok && allowed[urn] {
					filtered = append(filtered, node)
				}
			}
		}
		result["nodes"] = filtered
	}

	return result
}

// inputToMap converts an input struct to a map for logging.
func inputToMap(input any) map[string]any {
	data, err := json.Marshal(input)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}

// Verify interface implementations.
var (
	_ ToolMiddleware = (*URNResolverMiddleware)(nil)
	_ ToolMiddleware = (*AccessFilterMiddleware)(nil)
	_ ToolMiddleware = (*AuditLoggerMiddleware)(nil)
	_ ToolMiddleware = (*MetadataEnricherMiddleware)(nil)
)
