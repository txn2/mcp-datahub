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

	// UpdateColumnDescription sets the description for a specific schema field.
	UpdateColumnDescription(ctx context.Context, urn, fieldPath, description string) error

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

	// Query CRUD operations.

	// CreateQuery creates a new Query entity.
	CreateQuery(ctx context.Context, input client.CreateQueryInput) (*types.Query, error)

	// UpdateQuery updates an existing Query entity.
	UpdateQuery(ctx context.Context, input client.UpdateQueryInput) (*types.Query, error)

	// DeleteQuery deletes a Query entity.
	DeleteQuery(ctx context.Context, urn string) error

	// Structured properties (DataHub 1.3.x+).

	// GetStructuredProperties retrieves structured property values assigned to an entity.
	GetStructuredProperties(ctx context.Context, urn string) ([]types.StructuredPropertyValue, error)

	// ListStructuredPropertyDefinitions retrieves all structured property definitions.
	ListStructuredPropertyDefinitions(ctx context.Context) ([]types.StructuredPropertyDefinition, error)

	// UpsertStructuredProperties sets or updates structured property values on an entity.
	UpsertStructuredProperties(ctx context.Context, urn string, properties []types.StructuredPropertyInput) error

	// RemoveStructuredProperties removes structured properties from an entity.
	RemoveStructuredProperties(ctx context.Context, urn string, propertyURNs []string) error

	// Incidents (DataHub 1.3.x+).

	// GetIncidents retrieves active incidents for an entity.
	GetIncidents(ctx context.Context, urn string) (*types.IncidentResult, error)

	// RaiseIncident creates a new incident on entities.
	RaiseIncident(ctx context.Context, input types.RaiseIncidentInput) (string, error)

	// UpdateIncidentStatus changes the state of an incident (ACTIVE or RESOLVED).
	UpdateIncidentStatus(ctx context.Context, incidentURN, state, message string) error

	// ResolveIncident marks an incident as resolved.
	ResolveIncident(ctx context.Context, incidentURN, message string) error

	// Data contracts (DataHub 1.3.x+).

	// GetDataContract retrieves the data contract status for a dataset.
	GetDataContract(ctx context.Context, datasetURN string) (*types.DataContract, error)

	// Advanced search (DataHub 1.3.x+).

	// SearchAcrossEntities searches across entity types with optional advanced filtering.
	SearchAcrossEntities(ctx context.Context, query string, opts ...client.SearchOption) (*types.SearchResult, error)

	// Semantic search (DataHub 1.4.x + OpenSearch 2.19.3+).

	// SemanticSearch performs natural language semantic search across entities.
	SemanticSearch(ctx context.Context, query string, opts ...client.SearchOption) (*types.SearchResult, error)

	// Context documents (DataHub 1.4.x+).

	// GetDocument retrieves a context document by URN.
	GetDocument(ctx context.Context, urn string) (*types.Document, error)

	// GetRelatedDocuments retrieves documents linked to an entity.
	GetRelatedDocuments(ctx context.Context, urn string) ([]types.Document, error)

	// GetContextDocuments retrieves context documents linked to an entity
	// as simplified ContextDocument values for downstream consumption.
	GetContextDocuments(ctx context.Context, urn string) ([]types.ContextDocument, error)

	// UpsertContextDocument creates or updates a context document on an entity.
	// If doc.ID is empty, creates a new document. If set, updates the existing one.
	UpsertContextDocument(ctx context.Context, entityURN string, doc types.ContextDocumentInput) (*types.ContextDocument, error)

	// DeleteContextDocument removes a context document by its ID.
	DeleteContextDocument(ctx context.Context, documentID string) error

	// Entity creation (GraphQL mutations).

	// CreateTag creates a new tag entity.
	CreateTag(ctx context.Context, name, description string) (string, error)

	// CreateDomain creates a new domain entity.
	CreateDomain(ctx context.Context, name, description string) (string, error)

	// CreateGlossaryTerm creates a new glossary term entity.
	CreateGlossaryTerm(ctx context.Context, name, description, parentNode string) (string, error)

	// CreateDataProduct creates a new data product entity.
	CreateDataProduct(ctx context.Context, name, description, domainURN string) (string, error)

	// CreateDocument creates a new context document entity.
	CreateDocument(ctx context.Context, input types.CreateDocumentInput) (string, error)

	// CreateApplication creates a new application entity.
	CreateApplication(ctx context.Context, name, description string) (string, error)

	// CreateStructuredProperty creates a new structured property definition.
	CreateStructuredProperty(ctx context.Context, input types.CreateStructuredPropertyInput) (string, error)

	// UpsertDataContract creates or updates a data contract for a dataset.
	UpsertDataContract(ctx context.Context, input types.UpsertDataContractInput) (string, error)

	// Entity update operations.

	// AddOwner adds an owner to an entity.
	AddOwner(ctx context.Context, urn, ownerURN, ownershipType string) error

	// RemoveOwner removes an owner from an entity.
	RemoveOwner(ctx context.Context, urn, ownerURN string) error

	// SetDomain assigns a domain to an entity.
	SetDomain(ctx context.Context, entityURN, domainURN string) error

	// UnsetDomain removes the domain from an entity.
	UnsetDomain(ctx context.Context, entityURN string) error

	// UpdateIncident updates an existing incident.
	UpdateIncident(ctx context.Context, urn string, input types.UpdateIncidentInput) error

	// UpdateStructuredProperty updates a structured property definition.
	UpdateStructuredProperty(ctx context.Context, urn string, input types.UpdateStructuredPropertyInput) error

	// UpdateDocumentContents updates the title and text of a document.
	UpdateDocumentContents(ctx context.Context, urn, title, text string) error

	// UpdateDocumentStatus updates the status of a document.
	UpdateDocumentStatus(ctx context.Context, urn, status string) error

	// UpdateDocumentRelatedEntities updates entities related to a document.
	UpdateDocumentRelatedEntities(ctx context.Context, urn string, entityURNs []string) error

	// UpdateDocumentSubType updates the sub-type of a document.
	UpdateDocumentSubType(ctx context.Context, urn, subType string) error

	// Entity deletion (GraphQL mutations).

	// DeleteTag deletes a tag entity.
	DeleteTag(ctx context.Context, urn string) error

	// DeleteDomain deletes a domain entity.
	DeleteDomain(ctx context.Context, urn string) error

	// DeleteGlossaryEntity deletes a glossary term or node.
	DeleteGlossaryEntity(ctx context.Context, urn string) error

	// DeleteDataProduct deletes a data product entity.
	DeleteDataProduct(ctx context.Context, urn string) error

	// DeleteApplication deletes an application entity.
	DeleteApplication(ctx context.Context, urn string) error

	// DeleteDocument deletes a document entity.
	DeleteDocument(ctx context.Context, urn string) error

	// DeleteStructuredProperty deletes a structured property definition.
	DeleteStructuredProperty(ctx context.Context, urn string) error
}
