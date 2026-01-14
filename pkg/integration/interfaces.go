package integration

import "context"

// URNResolver resolves external identifiers to DataHub URNs.
// Implement this interface to map your internal IDs to DataHub URNs.
type URNResolver interface {
	// ResolveToDataHubURN converts an external identifier to a DataHub URN.
	ResolveToDataHubURN(ctx context.Context, externalID string) (string, error)
}

// AccessFilter controls access to DataHub entities.
// Implement this interface to add custom authorization logic.
type AccessFilter interface {
	// CanAccess checks if the current user can access the given URN.
	CanAccess(ctx context.Context, urn string) (bool, error)

	// FilterURNs filters a list of URNs to only those accessible by the current user.
	FilterURNs(ctx context.Context, urns []string) ([]string, error)
}

// AuditLogger logs tool invocations for audit purposes.
// Implement this interface to add custom audit logging.
type AuditLogger interface {
	// LogToolCall logs a tool invocation.
	LogToolCall(ctx context.Context, tool string, params map[string]any, userID string) error
}

// MetadataEnricher adds additional metadata to tool responses.
// Implement this interface to enrich DataHub responses with custom data.
type MetadataEnricher interface {
	// EnrichEntity adds custom metadata to an entity response.
	EnrichEntity(ctx context.Context, urn string, data map[string]any) (map[string]any, error)
}
