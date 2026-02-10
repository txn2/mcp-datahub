package client

import "errors"

// Sentinel errors for DataHub operations.
var (
	// ErrUnauthorized indicates invalid or missing authentication token.
	ErrUnauthorized = errors.New("unauthorized: invalid or missing token")

	// ErrForbidden indicates insufficient permissions.
	ErrForbidden = errors.New("forbidden: insufficient permissions")

	// ErrNotFound indicates the requested entity was not found.
	ErrNotFound = errors.New("entity not found")

	// ErrInvalidURN indicates the URN format is invalid.
	ErrInvalidURN = errors.New("invalid DataHub URN format")

	// ErrTimeout indicates the request timed out.
	ErrTimeout = errors.New("request timed out")

	// ErrRateLimited indicates rate limiting by DataHub.
	ErrRateLimited = errors.New("rate limited by DataHub")

	// ErrNotConfigured indicates the client is not properly configured.
	ErrNotConfigured = errors.New("datahub client not configured")

	// ErrWriteDisabled indicates write operations are not enabled.
	ErrWriteDisabled = errors.New("write operations are disabled: set WriteEnabled to true in config")
)
