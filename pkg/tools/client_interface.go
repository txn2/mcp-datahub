package tools

import (
	"context"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// DataHubClient defines the interface for DataHub operations.
// This allows for mocking in tests.
type DataHubClient interface {
	// Search searches for entities.
	Search(ctx context.Context, query string, opts ...client.SearchOption) (*types.SearchResult, error)

	// GetEntity retrieves a single entity by URN.
	GetEntity(ctx context.Context, urn string) (*types.Entity, error)

	// GetSchema retrieves schema for a dataset.
	GetSchema(ctx context.Context, urn string) (*types.SchemaMetadata, error)

	// GetSchemas retrieves schemas for multiple datasets by URN.
	GetSchemas(ctx context.Context, urns []string) (map[string]*types.SchemaMetadata, error)

	// GetLineage retrieves lineage for an entity.
	GetLineage(ctx context.Context, urn string, opts ...client.LineageOption) (*types.LineageResult, error)

	// GetColumnLineage retrieves fine-grained column-level lineage for a dataset.
	GetColumnLineage(ctx context.Context, urn string) (*types.ColumnLineage, error)

	// GetQueries retrieves queries for a dataset.
	GetQueries(ctx context.Context, urn string) (*types.QueryList, error)

	// GetGlossaryTerm retrieves a glossary term.
	GetGlossaryTerm(ctx context.Context, urn string) (*types.GlossaryTerm, error)

	// ListTags lists all tags.
	ListTags(ctx context.Context, filter string) ([]types.Tag, error)

	// ListDomains lists all domains.
	ListDomains(ctx context.Context) ([]types.Domain, error)

	// ListDataProducts lists all data products.
	ListDataProducts(ctx context.Context) ([]types.DataProduct, error)

	// GetDataProduct retrieves a data product by URN.
	GetDataProduct(ctx context.Context, urn string) (*types.DataProduct, error)

	// Ping tests the connection.
	Ping(ctx context.Context) error

	// Close closes the client.
	Close() error

	// Write operations (require WriteEnabled config).

	// UpdateDescription sets the editable description for an entity.
	UpdateDescription(ctx context.Context, urn, description string) error

	// AddTag adds a tag to an entity.
	AddTag(ctx context.Context, urn, tagURN string) error

	// RemoveTag removes a tag from an entity.
	RemoveTag(ctx context.Context, urn, tagURN string) error

	// AddGlossaryTerm adds a glossary term to an entity.
	AddGlossaryTerm(ctx context.Context, urn, termURN string) error

	// RemoveGlossaryTerm removes a glossary term from an entity.
	RemoveGlossaryTerm(ctx context.Context, urn, termURN string) error

	// AddLink adds a link to an entity.
	AddLink(ctx context.Context, urn, linkURL, description string) error

	// RemoveLink removes a link from an entity by URL.
	RemoveLink(ctx context.Context, urn, linkURL string) error

	// Structured properties (DataHub 1.4.x+).

	// GetStructuredProperties retrieves structured property values assigned to an entity.
	GetStructuredProperties(ctx context.Context, urn string) ([]types.StructuredPropertyValue, error)

	// ListStructuredPropertyDefinitions retrieves all structured property definitions.
	ListStructuredPropertyDefinitions(ctx context.Context) ([]types.StructuredPropertyDefinition, error)

	// UpsertStructuredProperties sets or updates structured property values on an entity.
	UpsertStructuredProperties(ctx context.Context, urn string, properties []types.StructuredPropertyInput) error

	// RemoveStructuredProperties removes structured properties from an entity.
	RemoveStructuredProperties(ctx context.Context, urn string, propertyURNs []string) error

	// Incidents (DataHub 1.4.x+).

	// GetIncidents retrieves active incidents for an entity.
	GetIncidents(ctx context.Context, urn string) (*types.IncidentResult, error)

	// RaiseIncident creates a new incident on entities.
	RaiseIncident(ctx context.Context, input types.RaiseIncidentInput) (string, error)

	// ResolveIncident marks an incident as resolved.
	ResolveIncident(ctx context.Context, incidentURN, message string) error

	// Data contracts (DataHub 1.4.x+).

	// GetDataContract retrieves the data contract status for a dataset.
	GetDataContract(ctx context.Context, datasetURN string) (*types.DataContract, error)

	// Semantic search (DataHub 1.4.x + OpenSearch 2.19.3+).

	// SemanticSearch performs natural language semantic search across entities.
	SemanticSearch(ctx context.Context, query string, opts ...client.SearchOption) (*types.SearchResult, error)

	// Context documents (DataHub 1.4.x+).

	// GetDocument retrieves a context document by URN.
	GetDocument(ctx context.Context, urn string) (*types.Document, error)

	// GetRelatedDocuments retrieves documents linked to an entity.
	GetRelatedDocuments(ctx context.Context, urn string) ([]types.Document, error)
}
