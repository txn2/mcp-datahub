package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"
)

// v3AspectRequest wraps an aspect value for the OpenAPI v3 POST body.
type v3AspectRequest struct {
	Value any `json:"value"`
}

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
// entityType is required for v3 URL construction; ignored in v1.
func (c *Client) getAspect(ctx context.Context, entityType, entityURN, aspectName string) (json.RawMessage, error) {
	var reqURL string
	if c.config.isV3() {
		reqURL = fmt.Sprintf("%s/openapi/v3/entity/%s/%s/%s",
			c.restBaseURL(), entityType, url.PathEscape(entityURN), aspectName)
	} else {
		reqURL = fmt.Sprintf("%s/aspects/%s?aspect=%s&version=0",
			c.restBaseURL(), entityURN, aspectName)
	}

	c.logger.Debug("REST GET aspect",
		"urn", entityURN,
		"aspect", aspectName,
		"url", reqURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setRESTHeaders(req)

	resp, err := c.httpClient.Do(req) //#nosec G704 -- URL is constructed from configured endpoint, not arbitrary user input
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

	// DataHub may return 200 OK with a null or empty value when the entity
	// exists but the requested aspect has never been written. Treat this the
	// same as 404 so callers initialize a default struct.
	if isNullOrEmptyJSON(aspectResp.Value) {
		return nil, ErrNotFound
	}

	return aspectResp.Value, nil
}

// postIngestProposal posts a metadata change proposal to the DataHub REST API.
// DataHub v1.3.0+ requires changeType and GenericAspect wrapper format.
// OpenAPI v3 (DataHub >= 1.4.0) uses a simpler body and URL structure.
func (c *Client) postIngestProposal(ctx context.Context, proposal ingestProposal) error {
	if c.config.isV3() {
		return c.postAspectV3(ctx, proposal)
	}
	return c.postAspectV1(ctx, proposal)
}

// postAspectV1 sends a metadata change proposal via the legacy Rest.li endpoint.
func (c *Client) postAspectV1(ctx context.Context, proposal ingestProposal) error {
	reqURL := fmt.Sprintf("%s/aspects?action=ingestProposal", c.restBaseURL())

	if proposal.ChangeType == "" {
		proposal.ChangeType = "UPSERT"
	}

	aspectJSON, err := json.Marshal(proposal.Aspect)
	if err != nil {
		return fmt.Errorf("failed to marshal aspect: %w", err)
	}
	proposal.Aspect = genericAspect{
		Value:       escapeNonASCII(aspectJSON),
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
		"url", reqURL)

	return c.doPost(ctx, reqURL, jsonBody)
}

// postAspectV3 sends an aspect update via the OpenAPI v3 endpoint.
func (c *Client) postAspectV3(ctx context.Context, proposal ingestProposal) error {
	reqURL := fmt.Sprintf("%s/openapi/v3/entity/%s/%s/%s",
		c.restBaseURL(), proposal.EntityType,
		url.PathEscape(proposal.EntityURN), proposal.AspectName)

	jsonBody, err := json.Marshal(v3AspectRequest{Value: proposal.Aspect})
	if err != nil {
		return fmt.Errorf("failed to marshal aspect: %w", err)
	}

	c.logger.Debug("REST POST v3 aspect",
		"urn", proposal.EntityURN,
		"aspect", proposal.AspectName,
		"entity_type", proposal.EntityType,
		"url", reqURL)

	return c.doPost(ctx, reqURL, jsonBody)
}

// doPost executes an HTTP POST with the given URL and JSON body.
func (c *Client) doPost(ctx context.Context, reqURL string, jsonBody []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setRESTHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req) //#nosec G704 -- URL is constructed from configured endpoint, not arbitrary user input
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
// The X-RestLi-Protocol-Version header is only sent for v1 (Rest.li) endpoints.
func (c *Client) setRESTHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	if !c.config.isV3() {
		req.Header.Set("X-RestLi-Protocol-Version", "2.0.0")
	}
}

// isNullOrEmptyJSON returns true if the raw JSON message is nil, empty,
// or the JSON literal "null". This covers cases where DataHub returns
// 200 OK but the aspect value is absent.
func isNullOrEmptyJSON(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return true
	}
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || string(trimmed) == "null"
}

// escapeNonASCII converts non-ASCII characters in JSON bytes to \uXXXX escape
// sequences. DataHub's GenericAspect.value is typed as "bytes" in the PDL schema,
// and RestLi uses Avro-style encoding where only characters U+0000–U+00FF are
// allowed. Characters like em dash (U+2014) cause ingestProposal validation
// failures unless escaped. Since non-ASCII characters in JSON can only appear
// inside string values, escaping all non-ASCII runes is safe.
func escapeNonASCII(data []byte) string {
	// Fast path: if all bytes are ASCII, skip allocation.
	allASCII := true
	for _, b := range data {
		if b > 0x7F {
			allASCII = false
			break
		}
	}
	if allASCII {
		return string(data)
	}

	var buf strings.Builder
	buf.Grow(len(data))
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r > 0x7F {
			if r <= 0xFFFF {
				fmt.Fprintf(&buf, "\\u%04x", r)
			} else {
				// Supplementary character: encode as surrogate pair.
				r -= 0x10000
				fmt.Fprintf(&buf, "\\u%04x\\u%04x", 0xD800+(r>>10), 0xDC00+(r&0x3FF))
			}
		} else {
			buf.WriteByte(data[i])
		}
		i += size
	}
	return buf.String()
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
