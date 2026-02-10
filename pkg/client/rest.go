package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// restBaseURL derives the REST API base URL from the GraphQL endpoint.
// For example, "https://datahub.example.com/api/graphql" -> "https://datahub.example.com".
func (c *Client) restBaseURL() string {
	return strings.TrimSuffix(c.endpoint, "/api/graphql")
}

// aspectResponse represents the response from GET /aspects endpoint.
type aspectResponse struct {
	Value json.RawMessage `json:"value"`
}

// ingestProposal represents a metadata change proposal for the REST API.
type ingestProposal struct {
	EntityType string `json:"entityType"`
	EntityURN  string `json:"entityUrn"`
	ChangeType string `json:"changeType"`
	AspectName string `json:"aspectName"`
	Aspect     any    `json:"aspect"`
}

// genericAspect wraps aspect JSON in the format required by DataHub v1.3.0+.
type genericAspect struct {
	Value       string `json:"value"`
	ContentType string `json:"contentType"`
}

// ingestRequest wraps the proposal for POST /aspects?action=ingestProposal.
type ingestRequest struct {
	Proposal ingestProposal `json:"proposal"`
}

// getAspect retrieves a raw aspect JSON from the DataHub REST API.
func (c *Client) getAspect(ctx context.Context, entityURN, aspectName string) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/aspects/%s?aspect=%s&version=0",
		c.restBaseURL(), entityURN, aspectName)

	c.logger.Debug("REST GET aspect",
		"urn", entityURN,
		"aspect", aspectName,
		"url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setRESTHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("REST GET failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Debug("REST GET response",
		"status", resp.StatusCode,
		"response_size", len(body))

	if err := c.checkRESTStatus(resp.StatusCode, body); err != nil {
		return nil, err
	}

	var aspectResp aspectResponse
	if err := json.Unmarshal(body, &aspectResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal aspect response: %w", err)
	}

	return aspectResp.Value, nil
}

// postIngestProposal posts a metadata change proposal to the DataHub REST API.
// DataHub v1.3.0+ requires changeType and GenericAspect wrapper format.
func (c *Client) postIngestProposal(ctx context.Context, proposal ingestProposal) error {
	url := fmt.Sprintf("%s/aspects?action=ingestProposal", c.restBaseURL())

	if proposal.ChangeType == "" {
		proposal.ChangeType = "UPSERT"
	}

	aspectJSON, err := json.Marshal(proposal.Aspect)
	if err != nil {
		return fmt.Errorf("failed to marshal aspect: %w", err)
	}
	proposal.Aspect = genericAspect{
		Value:       string(aspectJSON),
		ContentType: "application/json",
	}

	reqBody := ingestRequest{Proposal: proposal}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal proposal: %w", err)
	}

	c.logger.Debug("REST POST ingestProposal",
		"urn", proposal.EntityURN,
		"aspect", proposal.AspectName,
		"entity_type", proposal.EntityType,
		"url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setRESTHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("REST POST failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Debug("REST POST response",
		"status", resp.StatusCode,
		"response_size", len(body))

	return c.checkRESTStatus(resp.StatusCode, body)
}

// setRESTHeaders sets common headers for REST API requests.
func (c *Client) setRESTHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-RestLi-Protocol-Version", "2.0.0")
}

// checkRESTStatus validates REST API response status codes.
func (c *Client) checkRESTStatus(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return fmt.Errorf("REST API error (status %d): %s", statusCode, truncateString(string(body), 200))
	}
}
