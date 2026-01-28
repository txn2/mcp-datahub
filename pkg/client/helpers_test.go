package client

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestExtractOperationName(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "named query",
			query:    "query GetEntity($urn: String!) { entity(urn: $urn) { urn } }",
			expected: "GetEntity",
		},
		{
			name:     "named mutation",
			query:    "mutation UpdateEntity($input: EntityInput!) { update(input: $input) { urn } }",
			expected: "UpdateEntity",
		},
		{
			name:     "query with spaces",
			query:    "  query   SearchEntities($query: String) { search(query: $query) { results } }",
			expected: "SearchEntities",
		},
		{
			name:     "anonymous query with brace",
			query:    "{ entity(urn: \"urn:li:dataset:1\") { urn } }",
			expected: opNameAnonymous,
		},
		{
			name:     "query without name",
			query:    "query { entity { urn } }",
			expected: opNameAnonymous,
		},
		{
			name:     "query with only parens (no space after query)",
			query:    "query($urn: String!) { entity(urn: $urn) { urn } }",
			expected: opNameUnknown, // "query(" doesn't match "query " pattern
		},
		{
			name:     "unknown format",
			query:    "something weird",
			expected: opNameUnknown,
		},
		{
			name:     "empty string",
			query:    "",
			expected: opNameUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractOperationName(tt.query)
			if result != tt.expected {
				t.Errorf("extractOperationName(%q) = %q, want %q", tt.query, result, tt.expected)
			}
		})
	}
}

func TestExtractName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "name with parens",
			input:    "GetEntity($urn: String!)",
			expected: "GetEntity",
		},
		{
			name:     "name with brace",
			input:    "GetEntity { urn }",
			expected: "GetEntity",
		},
		{
			name:     "name with space",
			input:    "GetEntity ",
			expected: "GetEntity",
		},
		{
			name:     "starts with paren",
			input:    "($urn: String!)",
			expected: opNameAnonymous,
		},
		{
			name:     "starts with brace",
			input:    "{ urn }",
			expected: opNameAnonymous,
		},
		{
			name:     "empty string",
			input:    "",
			expected: opNameAnonymous,
		},
		{
			name:     "just spaces",
			input:    "   ",
			expected: opNameAnonymous,
		},
		{
			name:     "name only",
			input:    "GetEntity",
			expected: "GetEntity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractName(tt.input)
			if result != tt.expected {
				t.Errorf("extractName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "truncated",
			input:    "hello world",
			maxLen:   5,
			expected: "hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "zero max length",
			input:    "hello",
			maxLen:   0,
			expected: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestHandleRequestError(t *testing.T) {
	c := &Client{logger: NopLogger{}}

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := c.handleRequestError(ctx, errors.New("connection refused"))
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("expected ErrTimeout, got: %v", err)
		}
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()
		// Force deadline to expire
		<-ctx.Done()
		err := c.handleRequestError(ctx, errors.New("connection refused"))
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("expected ErrTimeout, got: %v", err)
		}
	})

	t.Run("regular error", func(t *testing.T) {
		ctx := context.Background()
		origErr := errors.New("connection refused")
		err := c.handleRequestError(ctx, origErr)
		if errors.Is(err, ErrTimeout) {
			t.Error("should not be ErrTimeout for regular error")
		}
		if !strings.Contains(err.Error(), "connection refused") {
			t.Errorf("expected error to contain original message, got: %v", err)
		}
	})
}

func TestHandleGraphQLErrors(t *testing.T) {
	c := &Client{logger: NopLogger{}}

	t.Run("not found error", func(t *testing.T) {
		errs := []graphQLError{{Message: "Entity not found"}}
		err := c.handleGraphQLErrors(errs)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("generic error", func(t *testing.T) {
		errs := []graphQLError{{Message: "Something went wrong"}}
		err := c.handleGraphQLErrors(errs)
		if errors.Is(err, ErrNotFound) {
			t.Error("should not be ErrNotFound")
		}
		if !strings.Contains(err.Error(), "Something went wrong") {
			t.Errorf("expected error message in error, got: %v", err)
		}
	})
}

func TestParseGraphQLResponse(t *testing.T) {
	c := &Client{logger: NopLogger{}}

	t.Run("invalid JSON", func(t *testing.T) {
		err := c.parseGraphQLResponse([]byte("not json"), nil)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal response") {
			t.Errorf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("null data without errors", func(t *testing.T) {
		// Valid JSON but data is null
		body := []byte(`{"data": null}`)
		err := c.parseGraphQLResponse(body, nil)
		if err == nil {
			t.Error("expected error for null data")
		}
		if !strings.Contains(err.Error(), "null data without errors") {
			t.Errorf("expected null data error, got: %v", err)
		}
	})

	t.Run("valid response with nil result", func(t *testing.T) {
		body := []byte(`{"data": {"entity": {"urn": "test"}}}`)
		err := c.parseGraphQLResponse(body, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid response with result", func(t *testing.T) {
		body := []byte(`{"data": {"name": "test"}}`)
		var result struct {
			Name string `json:"name"`
		}
		err := c.parseGraphQLResponse(body, &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.Name != "test" {
			t.Errorf("expected name=test, got: %s", result.Name)
		}
	})

	t.Run("unmarshal data fails", func(t *testing.T) {
		// Data is a string but we expect an object
		body := []byte(`{"data": "not an object"}`)
		var result struct {
			Name string `json:"name"`
		}
		err := c.parseGraphQLResponse(body, &result)
		if err == nil {
			t.Error("expected error for data unmarshal failure")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal data") {
			t.Errorf("expected unmarshal data error, got: %v", err)
		}
	})

	t.Run("graphql errors in response", func(t *testing.T) {
		body := []byte(`{"data": null, "errors": [{"message": "Entity not found"}]}`)
		err := c.parseGraphQLResponse(body, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})
}
