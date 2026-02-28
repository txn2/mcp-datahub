package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/integration"
)

// ToolkitOption configures a Toolkit.
type ToolkitOption func(*Toolkit)

// WithMiddleware adds global middleware to all tools.
func WithMiddleware(mw ToolMiddleware) ToolkitOption {
	return func(t *Toolkit) {
		t.middlewares = append(t.middlewares, mw)
	}
}

// WithToolMiddleware adds middleware to a specific tool.
func WithToolMiddleware(name ToolName, mw ToolMiddleware) ToolkitOption {
	return func(t *Toolkit) {
		t.toolMiddlewares[name] = append(t.toolMiddlewares[name], mw)
	}
}

// WithDescriptions sets toolkit-level description overrides for multiple tools.
// These take priority over default descriptions but are overridden by
// per-registration WithDescription options.
func WithDescriptions(descs map[ToolName]string) ToolkitOption {
	return func(t *Toolkit) {
		for name, desc := range descs {
			t.descriptions[name] = desc
		}
	}
}

// toolConfig holds per-registration configuration.
type toolConfig struct {
	middlewares  []ToolMiddleware
	description  *string
	annotations  *mcp.ToolAnnotations
	icons        []mcp.Icon
	title        *string
	outputSchema any
}

// ToolOption configures a single tool registration.
type ToolOption func(*toolConfig)

// WithDescription overrides the description for a single tool registration.
// This is the highest priority override, taking precedence over both
// toolkit-level WithDescriptions and default descriptions.
func WithDescription(desc string) ToolOption {
	return func(cfg *toolConfig) {
		cfg.description = &desc
	}
}

// WithPerToolMiddleware adds middleware for a single tool registration.
func WithPerToolMiddleware(mw ToolMiddleware) ToolOption {
	return func(cfg *toolConfig) {
		cfg.middlewares = append(cfg.middlewares, mw)
	}
}

// WithAnnotations sets toolkit-level annotation overrides for multiple tools.
// These take priority over default annotations but are overridden by
// per-registration WithAnnotation options.
func WithAnnotations(anns map[ToolName]*mcp.ToolAnnotations) ToolkitOption {
	return func(t *Toolkit) {
		for name, ann := range anns {
			t.annotations[name] = ann
		}
	}
}

// WithAnnotation overrides the annotations for a single tool registration.
// This is the highest priority override, taking precedence over both
// toolkit-level WithAnnotations and default annotations.
func WithAnnotation(ann *mcp.ToolAnnotations) ToolOption {
	return func(cfg *toolConfig) {
		cfg.annotations = ann
	}
}

// WithIcons sets toolkit-level icon overrides for multiple tools.
// These take priority over default icons but are overridden by
// per-registration WithIcon options.
func WithIcons(icons map[ToolName][]mcp.Icon) ToolkitOption {
	return func(t *Toolkit) {
		for name, ic := range icons {
			t.icons[name] = ic
		}
	}
}

// WithIcon overrides the icons for a single tool registration.
// This is the highest priority override, taking precedence over both
// toolkit-level WithIcons and default icons.
func WithIcon(icons []mcp.Icon) ToolOption {
	return func(cfg *toolConfig) {
		cfg.icons = icons
	}
}

// WithTitles sets toolkit-level title overrides for multiple tools.
// These take priority over default titles but are overridden by
// per-registration WithTitle options.
func WithTitles(titles map[ToolName]string) ToolkitOption {
	return func(t *Toolkit) {
		for name, title := range titles {
			t.titles[name] = title
		}
	}
}

// WithTitle overrides the display title for a single tool registration.
// This is the highest priority override, taking precedence over both
// toolkit-level WithTitles and default titles.
func WithTitle(title string) ToolOption {
	return func(cfg *toolConfig) {
		cfg.title = &title
	}
}

// WithOutputSchemas sets toolkit-level output schema overrides for multiple tools.
// These take priority over default output schemas but are overridden by
// per-registration WithOutputSchema options.
func WithOutputSchemas(schemas map[ToolName]any) ToolkitOption {
	return func(t *Toolkit) {
		for name, schema := range schemas {
			t.outputSchemas[name] = schema
		}
	}
}

// WithOutputSchema overrides the output schema for a single tool registration.
// This is the highest priority override, taking precedence over both
// toolkit-level WithOutputSchemas and default output schemas.
func WithOutputSchema(schema any) ToolOption {
	return func(cfg *toolConfig) {
		cfg.outputSchema = schema
	}
}

// Integration Interface Options
// These options connect the integration interfaces to the middleware system.

// WithURNResolver adds URN resolution capability to the toolkit.
// When configured, tools will resolve external identifiers to DataHub URNs
// before execution.
func WithURNResolver(r integration.URNResolver) ToolkitOption {
	return func(t *Toolkit) {
		t.urnResolver = r
	}
}

// WithAccessFilter adds access control capability to the toolkit.
// When configured, tools will check access before execution and filter
// results to only include accessible entities.
func WithAccessFilter(f integration.AccessFilter) ToolkitOption {
	return func(t *Toolkit) {
		t.accessFilter = f
	}
}

// WithAuditLogger adds audit logging capability to the toolkit.
// When configured, all tool invocations will be logged with the tool name,
// parameters, and user ID.
//
// The getUserID function extracts the user ID from the context. If nil,
// an empty user ID will be logged.
func WithAuditLogger(l integration.AuditLogger, getUserID func(context.Context) string) ToolkitOption {
	return func(t *Toolkit) {
		t.auditLogger = l
		t.getUserID = getUserID
	}
}

// WithMetadataEnricher adds metadata enrichment capability to the toolkit.
// When configured, entity responses will be enriched with custom metadata
// before being returned.
func WithMetadataEnricher(e integration.MetadataEnricher) ToolkitOption {
	return func(t *Toolkit) {
		t.metadataEnricher = e
	}
}

// WithQueryProvider adds query execution context to the toolkit.
// When configured, tools will enrich their responses with query execution
// context from the provider (table resolution, query examples, availability).
//
// This enables bidirectional integration with query engines like Trino.
func WithQueryProvider(p integration.QueryProvider) ToolkitOption {
	return func(t *Toolkit) {
		t.queryProvider = p
	}
}
