package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// Client is a GraphQL client for DataHub.
type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
	config     Config
}

// New creates a new DataHub client with the given configuration.
func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Apply defaults for unset values
	defaults := DefaultConfig()
	if cfg.Timeout == 0 {
		cfg.Timeout = defaults.Timeout
	}
	if cfg.RetryMax == 0 {
		cfg.RetryMax = defaults.RetryMax
	}
	if cfg.DefaultLimit == 0 {
		cfg.DefaultLimit = defaults.DefaultLimit
	}
	if cfg.MaxLimit == 0 {
		cfg.MaxLimit = defaults.MaxLimit
	}
	if cfg.MaxLineageDepth == 0 {
		cfg.MaxLineageDepth = defaults.MaxLineageDepth
	}

	// Ensure URL ends with /api/graphql
	endpoint := cfg.URL
	if !strings.HasSuffix(endpoint, "/api/graphql") {
		endpoint = strings.TrimSuffix(endpoint, "/") + "/api/graphql"
	}

	return &Client{
		endpoint: endpoint,
		token:    cfg.Token,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
	}, nil
}

// NewFromEnv creates a new DataHub client from environment variables.
func NewFromEnv() (*Client, error) {
	cfg, err := FromEnv()
	if err != nil {
		return nil, err
	}
	return New(cfg)
}

// graphQLRequest represents a GraphQL request.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// graphQLResponse represents a GraphQL response.
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors,omitempty"`
}

// graphQLError represents a GraphQL error.
type graphQLError struct {
	Message string `json:"message"`
	Path    []any  `json:"path,omitempty"`
}

// Execute executes a GraphQL query and unmarshals the response into result.
func (c *Client) Execute(ctx context.Context, query string, variables map[string]any, result any) error {
	reqBody := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.RetryMax; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt*attempt) * 100 * time.Millisecond)
		}

		lastErr = c.doRequest(ctx, jsonBody, result)
		if lastErr == nil {
			return nil
		}

		// Don't retry on certain errors
		if errors.Is(lastErr, ErrUnauthorized) || errors.Is(lastErr, ErrForbidden) || errors.Is(lastErr, ErrNotFound) {
			return lastErr
		}
	}

	return lastErr
}

func (c *Client) doRequest(ctx context.Context, jsonBody []byte, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ErrTimeout
		}
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		// Error intentionally ignored; response body close errors are non-actionable
		_ = resp.Body.Close() //nolint:errcheck // closing response body, error not actionable
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusOK:
		// Continue processing
	default:
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		errMsg := gqlResp.Errors[0].Message
		if strings.Contains(strings.ToLower(errMsg), "not found") {
			return ErrNotFound
		}
		return fmt.Errorf("graphql error: %s", errMsg)
	}

	if result != nil && gqlResp.Data != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return nil
}

// Close closes the client.
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// Ping tests the connection to DataHub.
func (c *Client) Ping(ctx context.Context) error {
	var result map[string]any
	return c.Execute(ctx, PingQuery, nil, &result)
}

// Config returns the client configuration.
func (c *Client) Config() Config {
	return c.config
}

// Search searches for entities in DataHub.
func (c *Client) Search(ctx context.Context, query string, opts ...SearchOption) (*types.SearchResult, error) {
	options := &searchOptions{
		limit:  c.config.DefaultLimit,
		offset: 0,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Clamp limit
	if options.limit > c.config.MaxLimit {
		options.limit = c.config.MaxLimit
	}

	// Default to DATASET if no entity type specified
	entityType := options.entityType
	if entityType == "" {
		entityType = "DATASET"
	}

	input := map[string]any{
		"type":  entityType,
		"query": query,
		"start": options.offset,
		"count": options.limit,
	}

	variables := map[string]any{
		"input": input,
	}

	var response struct {
		Search struct {
			Start         int `json:"start"`
			Count         int `json:"count"`
			Total         int `json:"total"`
			SearchResults []struct {
				Entity struct {
					URN         string `json:"urn"`
					Type        string `json:"type"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Platform    struct {
						Name string `json:"name"`
					} `json:"platform"`
					// For DataProduct, GlossaryTerm, Tag - name/description in properties
					Properties struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"properties"`
					Ownership struct {
						Owners []struct {
							Owner struct {
								URN      string `json:"urn"`
								Username string `json:"username"`
								Name     string `json:"name"`
							} `json:"owner"`
							Type string `json:"type"`
						} `json:"owners"`
					} `json:"ownership"`
					Tags struct {
						Tags []struct {
							Tag struct {
								URN         string `json:"urn"`
								Name        string `json:"name"`
								Description string `json:"description"`
							} `json:"tag"`
						} `json:"tags"`
					} `json:"tags"`
					Domain struct {
						Domain struct {
							URN        string `json:"urn"`
							Properties struct {
								Name        string `json:"name"`
								Description string `json:"description"`
							} `json:"properties"`
						} `json:"domain"`
					} `json:"domain"`
				} `json:"entity"`
				MatchedFields []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"matchedFields"`
			} `json:"searchResults"`
		} `json:"search"`
	}

	if err := c.Execute(ctx, SearchQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("search(%q): %w", query, err)
	}

	result := &types.SearchResult{
		Total:  response.Search.Total,
		Offset: response.Search.Start,
		Limit:  response.Search.Count,
	}

	for _, sr := range response.Search.SearchResults {
		name := sr.Entity.Name
		description := sr.Entity.Description
		// For DataProduct, GlossaryTerm, Tag the name/description come from properties
		if sr.Entity.Properties.Name != "" {
			name = sr.Entity.Properties.Name
		}
		if sr.Entity.Properties.Description != "" {
			description = sr.Entity.Properties.Description
		}

		entity := types.SearchEntity{
			URN:         sr.Entity.URN,
			Type:        sr.Entity.Type,
			Name:        name,
			Description: description,
			Platform:    sr.Entity.Platform.Name,
		}

		// Parse ownership
		for _, o := range sr.Entity.Ownership.Owners {
			ownerName := o.Owner.Username
			if o.Owner.Name != "" {
				ownerName = o.Owner.Name
			}
			entity.Owners = append(entity.Owners, types.Owner{
				URN:  o.Owner.URN,
				Name: ownerName,
				Type: types.OwnershipType(o.Type),
			})
		}

		// Parse tags
		for _, t := range sr.Entity.Tags.Tags {
			entity.Tags = append(entity.Tags, types.Tag{
				URN:         t.Tag.URN,
				Name:        t.Tag.Name,
				Description: t.Tag.Description,
			})
		}

		// Parse domain
		if sr.Entity.Domain.Domain.URN != "" {
			entity.Domain = &types.Domain{
				URN:         sr.Entity.Domain.Domain.URN,
				Name:        sr.Entity.Domain.Domain.Properties.Name,
				Description: sr.Entity.Domain.Domain.Properties.Description,
			}
		}

		for _, mf := range sr.MatchedFields {
			entity.MatchedFields = append(entity.MatchedFields, types.MatchedField{
				Name:  mf.Name,
				Value: mf.Value,
			})
		}

		result.Entities = append(result.Entities, entity)
	}

	return result, nil
}

// GetEntity retrieves a single entity by URN.
func (c *Client) GetEntity(ctx context.Context, urn string) (*types.Entity, error) {
	variables := map[string]any{
		"urn": urn,
	}

	var response struct {
		Entity struct {
			URN         string `json:"urn"`
			Type        string `json:"type"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Platform    struct {
				Name string `json:"name"`
			} `json:"platform"`
			Properties struct {
				Name             string `json:"name"`
				Description      string `json:"description"`
				CustomProperties []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"customProperties"`
			} `json:"properties"`
			SubTypes struct {
				TypeNames []string `json:"typeNames"`
			} `json:"subTypes"`
			Ownership struct {
				Owners []struct {
					Owner struct {
						URN      string `json:"urn"`
						Username string `json:"username"`
						Name     string `json:"name"`
						Info     struct {
							DisplayName string `json:"displayName"`
							Email       string `json:"email"`
						} `json:"info"`
					} `json:"owner"`
					Type string `json:"type"`
				} `json:"owners"`
			} `json:"ownership"`
			Tags struct {
				Tags []struct {
					Tag struct {
						URN         string `json:"urn"`
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"tag"`
				} `json:"tags"`
			} `json:"tags"`
			GlossaryTerms struct {
				Terms []struct {
					Term struct {
						URN        string `json:"urn"`
						Properties struct {
							Name        string `json:"name"`
							Description string `json:"description"`
						} `json:"properties"`
					} `json:"term"`
				} `json:"terms"`
			} `json:"glossaryTerms"`
			Domain struct {
				Domain struct {
					URN        string `json:"urn"`
					Properties struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"properties"`
				} `json:"domain"`
			} `json:"domain"`
			Deprecation struct {
				Deprecated       bool   `json:"deprecated"`
				Note             string `json:"note"`
				Actor            string `json:"actor"`
				DecommissionTime int64  `json:"decommissionTime"`
			} `json:"deprecation"`
		} `json:"entity"`
	}

	if err := c.Execute(ctx, GetEntityQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetEntity(%s): %w", urn, err)
	}

	if response.Entity.URN == "" {
		return nil, fmt.Errorf("GetEntity(%s): %w", urn, ErrNotFound)
	}

	entity := &types.Entity{
		URN:         response.Entity.URN,
		Type:        response.Entity.Type,
		Name:        response.Entity.Name,
		Platform:    response.Entity.Platform.Name,
		Description: response.Entity.Description,
	}

	if response.Entity.Properties.Name != "" {
		entity.Name = response.Entity.Properties.Name
	}
	if response.Entity.Properties.Description != "" {
		entity.Description = response.Entity.Properties.Description
	}

	// Parse ownership
	for _, o := range response.Entity.Ownership.Owners {
		name := o.Owner.Username
		if o.Owner.Info.DisplayName != "" {
			name = o.Owner.Info.DisplayName
		} else if o.Owner.Name != "" {
			name = o.Owner.Name
		}
		entity.Owners = append(entity.Owners, types.Owner{
			URN:   o.Owner.URN,
			Name:  name,
			Email: o.Owner.Info.Email,
			Type:  types.OwnershipType(o.Type),
		})
	}

	// Parse tags
	for _, t := range response.Entity.Tags.Tags {
		entity.Tags = append(entity.Tags, types.Tag{
			URN:         t.Tag.URN,
			Name:        t.Tag.Name,
			Description: t.Tag.Description,
		})
	}

	// Parse glossary terms
	for _, gt := range response.Entity.GlossaryTerms.Terms {
		entity.GlossaryTerms = append(entity.GlossaryTerms, types.GlossaryTerm{
			URN:         gt.Term.URN,
			Name:        gt.Term.Properties.Name,
			Description: gt.Term.Properties.Description,
		})
	}

	// Parse domain
	if response.Entity.Domain.Domain.URN != "" {
		entity.Domain = &types.Domain{
			URN:         response.Entity.Domain.Domain.URN,
			Name:        response.Entity.Domain.Domain.Properties.Name,
			Description: response.Entity.Domain.Domain.Properties.Description,
		}
	}

	// Parse subTypes
	if len(response.Entity.SubTypes.TypeNames) > 0 {
		entity.SubTypes = response.Entity.SubTypes.TypeNames
	}

	if response.Entity.Deprecation.Deprecated {
		entity.Deprecation = &types.Deprecation{
			Deprecated:       response.Entity.Deprecation.Deprecated,
			Note:             response.Entity.Deprecation.Note,
			Actor:            response.Entity.Deprecation.Actor,
			DecommissionTime: response.Entity.Deprecation.DecommissionTime,
		}
	}

	// Parse custom properties
	if len(response.Entity.Properties.CustomProperties) > 0 {
		entity.Properties = make(map[string]any)
		for _, cp := range response.Entity.Properties.CustomProperties {
			entity.Properties[cp.Key] = cp.Value
		}
	}

	return entity, nil
}

// GetSchema retrieves schema metadata for a dataset.
func (c *Client) GetSchema(ctx context.Context, urn string) (*types.SchemaMetadata, error) {
	variables := map[string]any{
		"urn": urn,
	}

	var response struct {
		Dataset struct {
			SchemaMetadata rawSchemaMetadata `json:"schemaMetadata"`
		} `json:"dataset"`
	}

	if err := c.Execute(ctx, GetSchemaQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetSchema(%s): %w", urn, err)
	}

	return parseSchemaMetadata(response.Dataset.SchemaMetadata), nil
}

// GetLineage retrieves lineage for an entity.
func (c *Client) GetLineage(ctx context.Context, urn string, opts ...LineageOption) (*types.LineageResult, error) {
	options := &lineageOptions{
		direction: LineageDirectionDownstream,
		depth:     1,
	}
	for _, opt := range opts {
		opt(options)
	}

	if options.depth > c.config.MaxLineageDepth {
		options.depth = c.config.MaxLineageDepth
	}

	variables := map[string]any{
		"urn":       urn,
		"direction": options.direction,
	}

	var response struct {
		SearchAcrossLineage struct {
			SearchResults []struct {
				Entity struct {
					URN         string `json:"urn"`
					Type        string `json:"type"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Platform    struct {
						Name string `json:"name"`
					} `json:"platform"`
				} `json:"entity"`
				Degree int `json:"degree"`
				Paths  []struct {
					Path []struct {
						URN string `json:"urn"`
					} `json:"path"`
				} `json:"paths"`
			} `json:"searchResults"`
		} `json:"searchAcrossLineage"`
	}

	if err := c.Execute(ctx, GetLineageQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetLineage(%s): %w", urn, err)
	}

	result := &types.LineageResult{
		Start:     urn,
		Direction: options.direction,
		Depth:     options.depth,
	}

	// Build nodes and edges (filter by depth client-side since maxHops is not supported)
	nodesByDegree := make(map[int][]string)
	edgeSet := make(map[string]bool) // Track unique edges

	for _, sr := range response.SearchAcrossLineage.SearchResults {
		// Filter by depth client-side
		if sr.Degree > options.depth {
			continue
		}
		result.Nodes = append(result.Nodes, types.LineageNode{
			URN:         sr.Entity.URN,
			Type:        sr.Entity.Type,
			Name:        sr.Entity.Name,
			Description: sr.Entity.Description,
			Platform:    sr.Entity.Platform.Name,
			Level:       sr.Degree,
		})
		nodesByDegree[sr.Degree] = append(nodesByDegree[sr.Degree], sr.Entity.URN)

		// Build edges from paths if available (only within depth limit)
		for _, pathGroup := range sr.Paths {
			if len(pathGroup.Path) > 1 {
				// Only include edges where both nodes are within depth
				maxPathIdx := options.depth
				if maxPathIdx > len(pathGroup.Path)-1 {
					maxPathIdx = len(pathGroup.Path) - 1
				}
				for i := 0; i < maxPathIdx; i++ {
					edgeKey := pathGroup.Path[i].URN + "->" + pathGroup.Path[i+1].URN
					if !edgeSet[edgeKey] {
						edgeSet[edgeKey] = true
						result.Edges = append(result.Edges, types.LineageEdge{
							Source: pathGroup.Path[i].URN,
							Target: pathGroup.Path[i+1].URN,
						})
					}
				}
			}
		}
	}

	// If no paths were provided, infer edges from degree
	// For upstream: nodes at degree N connect to nodes at degree N-1
	// For downstream: nodes at degree N connect to nodes at degree N+1
	if len(result.Edges) == 0 && len(result.Nodes) > 0 {
		// Add edge from start node to degree 1 nodes
		for _, nodeURN := range nodesByDegree[1] {
			if options.direction == LineageDirectionUpstream {
				result.Edges = append(result.Edges, types.LineageEdge{
					Source: nodeURN,
					Target: urn,
				})
			} else {
				result.Edges = append(result.Edges, types.LineageEdge{
					Source: urn,
					Target: nodeURN,
				})
			}
		}
	}

	return result, nil
}

// GetQueries retrieves queries associated with a dataset.
// Returns empty result if usage stats are not configured for the dataset.
func (c *Client) GetQueries(ctx context.Context, urn string) (*types.QueryList, error) {
	variables := map[string]any{
		"urn": urn,
	}

	var response struct {
		Dataset struct {
			UsageStats struct {
				Buckets []struct {
					Metrics struct {
						TopSQLQueries []string `json:"topSqlQueries"`
					} `json:"metrics"`
				} `json:"buckets"`
			} `json:"usageStats"`
		} `json:"dataset"`
	}

	// Execute query - may return error if usage stats not configured
	err := c.Execute(ctx, GetQueriesQuery, variables, &response)
	if err != nil {
		// Return empty result if usage stats are not available
		// This is common when usage tracking isn't configured
		return &types.QueryList{Total: 0}, nil
	}

	result := &types.QueryList{}

	for _, bucket := range response.Dataset.UsageStats.Buckets {
		for _, q := range bucket.Metrics.TopSQLQueries {
			result.Queries = append(result.Queries, types.Query{
				Statement: q,
			})
		}
	}
	result.Total = len(result.Queries)

	return result, nil
}

// GetGlossaryTerm retrieves a glossary term by URN.
func (c *Client) GetGlossaryTerm(ctx context.Context, urn string) (*types.GlossaryTerm, error) {
	variables := map[string]any{
		"urn": urn,
	}

	var response struct {
		GlossaryTerm struct {
			URN              string `json:"urn"`
			Name             string `json:"name"`
			HierarchicalName string `json:"hierarchicalName"`
			Properties       struct {
				Name             string `json:"name"`
				Description      string `json:"description"`
				CustomProperties []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"customProperties"`
			} `json:"properties"`
			ParentNodes struct {
				Nodes []struct {
					URN        string `json:"urn"`
					Properties struct {
						Name string `json:"name"`
					} `json:"properties"`
				} `json:"nodes"`
			} `json:"parentNodes"`
			Ownership struct {
				Owners []struct {
					Owner struct {
						URN      string `json:"urn"`
						Username string `json:"username"`
					} `json:"owner"`
					Type string `json:"type"`
				} `json:"owners"`
			} `json:"ownership"`
		} `json:"glossaryTerm"`
	}

	if err := c.Execute(ctx, GetGlossaryTermQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetGlossaryTerm(%s): %w", urn, err)
	}

	if response.GlossaryTerm.URN == "" {
		return nil, fmt.Errorf("GetGlossaryTerm(%s): %w", urn, ErrNotFound)
	}

	term := &types.GlossaryTerm{
		URN:         response.GlossaryTerm.URN,
		Name:        response.GlossaryTerm.Name,
		Description: response.GlossaryTerm.Properties.Description,
	}

	if response.GlossaryTerm.Properties.Name != "" {
		term.Name = response.GlossaryTerm.Properties.Name
	}

	// Parse parent node (take the first one if multiple)
	if len(response.GlossaryTerm.ParentNodes.Nodes) > 0 {
		term.ParentNode = response.GlossaryTerm.ParentNodes.Nodes[0].URN
	}

	// Parse ownership
	for _, o := range response.GlossaryTerm.Ownership.Owners {
		term.Owners = append(term.Owners, types.Owner{
			URN:  o.Owner.URN,
			Name: o.Owner.Username,
			Type: types.OwnershipType(o.Type),
		})
	}

	// Parse custom properties
	if len(response.GlossaryTerm.Properties.CustomProperties) > 0 {
		term.Properties = make(map[string]string)
		for _, cp := range response.GlossaryTerm.Properties.CustomProperties {
			term.Properties[cp.Key] = cp.Value
		}
	}

	return term, nil
}

// ListTags lists all tags, optionally filtered.
func (c *Client) ListTags(ctx context.Context, filter string) ([]types.Tag, error) {
	query := "*"
	if filter != "" {
		query = filter
	}

	input := map[string]any{
		"type":  "TAG",
		"query": query,
		"start": 0,
		"count": c.config.MaxLimit,
	}

	variables := map[string]any{
		"input": input,
	}

	var response struct {
		Search struct {
			SearchResults []struct {
				Entity struct {
					URN         string `json:"urn"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Properties  struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"properties"`
				} `json:"entity"`
			} `json:"searchResults"`
		} `json:"search"`
	}

	if err := c.Execute(ctx, ListTagsQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("ListTags: %w", err)
	}

	tags := make([]types.Tag, 0, len(response.Search.SearchResults))
	for _, sr := range response.Search.SearchResults {
		tag := types.Tag{
			URN:         sr.Entity.URN,
			Name:        sr.Entity.Name,
			Description: sr.Entity.Description,
		}
		if sr.Entity.Properties.Name != "" {
			tag.Name = sr.Entity.Properties.Name
		}
		if sr.Entity.Properties.Description != "" {
			tag.Description = sr.Entity.Properties.Description
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// ListDomains lists all domains.
func (c *Client) ListDomains(ctx context.Context) ([]types.Domain, error) {
	var response struct {
		ListDomains struct {
			Total   int `json:"total"`
			Domains []struct {
				URN        string `json:"urn"`
				Properties struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"properties"`
				Ownership struct {
					Owners []struct {
						Owner struct {
							URN      string `json:"urn"`
							Username string `json:"username"`
						} `json:"owner"`
						Type string `json:"type"`
					} `json:"owners"`
				} `json:"ownership"`
				Entities struct {
					Total int `json:"total"`
				} `json:"entities"`
			} `json:"domains"`
		} `json:"listDomains"`
	}

	if err := c.Execute(ctx, ListDomainsQuery, nil, &response); err != nil {
		return nil, fmt.Errorf("ListDomains: %w", err)
	}

	domains := make([]types.Domain, 0, len(response.ListDomains.Domains))
	for _, d := range response.ListDomains.Domains {
		domain := types.Domain{
			URN:         d.URN,
			Name:        d.Properties.Name,
			Description: d.Properties.Description,
			EntityCount: d.Entities.Total,
		}

		// Parse ownership
		for _, o := range d.Ownership.Owners {
			domain.Owners = append(domain.Owners, types.Owner{
				URN:  o.Owner.URN,
				Name: o.Owner.Username,
				Type: types.OwnershipType(o.Type),
			})
		}

		domains = append(domains, domain)
	}

	return domains, nil
}

// ListDataProducts lists all data products.
// Uses search API as fallback if listDataProducts query is not available.
func (c *Client) ListDataProducts(ctx context.Context) ([]types.DataProduct, error) {
	// Try the dedicated listDataProducts query first
	var response struct {
		ListDataProducts struct {
			Total        int `json:"total"`
			DataProducts []struct {
				URN        string `json:"urn"`
				Properties struct {
					Name             string `json:"name"`
					Description      string `json:"description"`
					CustomProperties []struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					} `json:"customProperties"`
				} `json:"properties"`
				Domain struct {
					Domain struct {
						URN        string `json:"urn"`
						Properties struct {
							Name string `json:"name"`
						} `json:"properties"`
					} `json:"domain"`
				} `json:"domain"`
				Ownership struct {
					Owners []struct {
						Owner struct {
							URN      string `json:"urn"`
							Username string `json:"username"`
							Name     string `json:"name"`
						} `json:"owner"`
						Type string `json:"type"`
					} `json:"owners"`
				} `json:"ownership"`
			} `json:"dataProducts"`
		} `json:"listDataProducts"`
	}

	err := c.Execute(ctx, ListDataProductsQuery, nil, &response)
	if err == nil {
		var products []types.DataProduct
		for _, dp := range response.ListDataProducts.DataProducts {
			product := types.DataProduct{
				URN:         dp.URN,
				Name:        dp.Properties.Name,
				Description: dp.Properties.Description,
			}

			if dp.Domain.Domain.URN != "" {
				product.Domain = &types.Domain{
					URN:  dp.Domain.Domain.URN,
					Name: dp.Domain.Domain.Properties.Name,
				}
			}

			for _, o := range dp.Ownership.Owners {
				name := o.Owner.Username
				if o.Owner.Name != "" {
					name = o.Owner.Name
				}
				product.Owners = append(product.Owners, types.Owner{
					URN:  o.Owner.URN,
					Name: name,
					Type: types.OwnershipType(o.Type),
				})
			}

			if len(dp.Properties.CustomProperties) > 0 {
				product.Properties = make(map[string]string)
				for _, cp := range dp.Properties.CustomProperties {
					product.Properties[cp.Key] = cp.Value
				}
			}

			products = append(products, product)
		}
		return products, nil
	}

	// Fall back to search API for older DataHub versions
	searchResults, searchErr := c.Search(ctx, "*", WithEntityType("DATA_PRODUCT"), WithLimit(c.config.MaxLimit))
	if searchErr != nil {
		// Return original error if search also fails
		return nil, fmt.Errorf("ListDataProducts: %w (search fallback also failed: %w)", err, searchErr)
	}

	products := make([]types.DataProduct, 0, len(searchResults.Entities))
	for _, e := range searchResults.Entities {
		products = append(products, types.DataProduct{
			URN:         e.URN,
			Name:        e.Name,
			Description: e.Description,
		})
	}

	return products, nil
}

// GetDataProduct retrieves a data product by URN.
func (c *Client) GetDataProduct(ctx context.Context, urn string) (*types.DataProduct, error) {
	variables := map[string]any{
		"urn": urn,
	}

	var response struct {
		DataProduct struct {
			URN        string `json:"urn"`
			Properties struct {
				Name             string `json:"name"`
				Description      string `json:"description"`
				CustomProperties []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"customProperties"`
			} `json:"properties"`
			Domain struct {
				Domain struct {
					URN        string `json:"urn"`
					Properties struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"properties"`
				} `json:"domain"`
			} `json:"domain"`
			Ownership struct {
				Owners []struct {
					Owner struct {
						URN      string `json:"urn"`
						Username string `json:"username"`
						Name     string `json:"name"`
						Info     struct {
							DisplayName string `json:"displayName"`
							Email       string `json:"email"`
						} `json:"info"`
					} `json:"owner"`
					Type string `json:"type"`
				} `json:"owners"`
			} `json:"ownership"`
		} `json:"dataProduct"`
	}

	if err := c.Execute(ctx, GetDataProductQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetDataProduct(%s): %w", urn, err)
	}

	if response.DataProduct.URN == "" {
		return nil, fmt.Errorf("GetDataProduct(%s): %w", urn, ErrNotFound)
	}

	product := &types.DataProduct{
		URN:         response.DataProduct.URN,
		Name:        response.DataProduct.Properties.Name,
		Description: response.DataProduct.Properties.Description,
	}

	if response.DataProduct.Domain.Domain.URN != "" {
		product.Domain = &types.Domain{
			URN:         response.DataProduct.Domain.Domain.URN,
			Name:        response.DataProduct.Domain.Domain.Properties.Name,
			Description: response.DataProduct.Domain.Domain.Properties.Description,
		}
	}

	for _, o := range response.DataProduct.Ownership.Owners {
		name := o.Owner.Username
		if o.Owner.Info.DisplayName != "" {
			name = o.Owner.Info.DisplayName
		} else if o.Owner.Name != "" {
			name = o.Owner.Name
		}
		product.Owners = append(product.Owners, types.Owner{
			URN:  o.Owner.URN,
			Name: name,
			Type: types.OwnershipType(o.Type),
		})
	}

	// Note: Assets field not available in all DataHub versions
	// Data product assets can be queried via search with the data product URN filter

	if len(response.DataProduct.Properties.CustomProperties) > 0 {
		product.Properties = make(map[string]string)
		for _, cp := range response.DataProduct.Properties.CustomProperties {
			product.Properties[cp.Key] = cp.Value
		}
	}

	return product, nil
}

// GetColumnLineage retrieves fine-grained column-level lineage for a dataset.
// Returns empty result if fine-grained lineage is not available for the dataset.
func (c *Client) GetColumnLineage(ctx context.Context, urn string) (*types.ColumnLineage, error) {
	variables := map[string]any{
		"urn": urn,
	}

	var response struct {
		Dataset struct {
			FineGrainedLineages []struct {
				Upstreams []struct {
					Path    string `json:"path"`
					Dataset string `json:"dataset"`
				} `json:"upstreams"`
				Downstreams []struct {
					Path string `json:"path"`
				} `json:"downstreams"`
				TransformOperation string  `json:"transformOperation"`
				ConfidenceScore    float64 `json:"confidenceScore"`
				Query              string  `json:"query"`
			} `json:"fineGrainedLineages"`
		} `json:"dataset"`
	}

	if err := c.Execute(ctx, GetColumnLineageQuery, variables, &response); err != nil {
		// Return empty result if fine-grained lineage is not available
		return &types.ColumnLineage{DatasetURN: urn}, nil
	}

	result := &types.ColumnLineage{
		DatasetURN: urn,
	}

	// Build column lineage mappings from fine-grained lineages
	for _, fgl := range response.Dataset.FineGrainedLineages {
		// Each fine-grained lineage entry maps downstream columns to upstream columns
		for _, downstream := range fgl.Downstreams {
			for _, upstream := range fgl.Upstreams {
				mapping := types.ColumnLineageMapping{
					DownstreamColumn: downstream.Path,
					UpstreamDataset:  upstream.Dataset,
					UpstreamColumn:   upstream.Path,
					Transform:        fgl.TransformOperation,
					Query:            fgl.Query,
					ConfidenceScore:  fgl.ConfidenceScore,
				}
				result.Mappings = append(result.Mappings, mapping)
			}
		}
	}

	return result, nil
}

// GetSchemas retrieves schema metadata for multiple datasets by URN.
// Returns a map of URN to schema metadata. Datasets without schemas are omitted.
func (c *Client) GetSchemas(ctx context.Context, urns []string) (map[string]*types.SchemaMetadata, error) {
	if len(urns) == 0 {
		return make(map[string]*types.SchemaMetadata), nil
	}

	variables := map[string]any{
		"urns": urns,
	}

	var response struct {
		Entities []struct {
			URN            string            `json:"urn"`
			SchemaMetadata rawSchemaMetadata `json:"schemaMetadata"`
		} `json:"entities"`
	}

	if err := c.Execute(ctx, BatchGetSchemasQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetSchemas: %w", err)
	}

	result := make(map[string]*types.SchemaMetadata)

	for _, entity := range response.Entities {
		if entity.URN == "" {
			continue
		}

		result[entity.URN] = parseSchemaMetadata(entity.SchemaMetadata)
	}

	return result, nil
}
