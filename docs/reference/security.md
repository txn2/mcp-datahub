# Security Reference

Security considerations and best practices for mcp-datahub.

## Authentication

### Token Handling

- Tokens are passed via environment variables only
- Tokens are never logged or included in error messages
- Tokens are redacted in debug output

### Best Practices

1. **Use environment variables**: Never hardcode tokens
2. **Rotate regularly**: Change tokens periodically
3. **Minimal permissions**: Use tokens with least required access
4. **Secure storage**: Use secret management systems

## Network Security

### TLS/SSL

All connections to DataHub should use HTTPS:

```bash
export DATAHUB_URL=https://datahub.company.com  # Good
export DATAHUB_URL=http://datahub.company.com   # Bad
```

### Certificate Verification

TLS certificate verification is enabled by default. Do not disable in production.

## Access Control

### DataHub Permissions

mcp-datahub respects DataHub's built-in authorization:

- Users can only access entities they have permission to view
- Tokens inherit the permissions of the user who created them

### Custom Access Filtering

Implement the `AccessFilter` interface for additional controls:

```go
type AccessFilter interface {
    CanAccess(ctx context.Context, urn string) (bool, error)
    FilterURNs(ctx context.Context, urns []string) ([]string, error)
}
```

## Rate Limiting

### Built-in Limits

- `DATAHUB_MAX_LIMIT`: Caps search result size
- `DATAHUB_MAX_LINEAGE_DEPTH`: Limits lineage traversal

### Custom Rate Limiting

Add rate limiting via middleware:

```go
toolkit := tools.NewToolkit(client,
    tools.WithMiddleware(rateLimitMiddleware),
)
```

## Audit Logging

Implement the `AuditLogger` interface:

```go
type AuditLogger interface {
    LogToolCall(ctx context.Context, tool string, params map[string]any, userID string) error
}
```

## Supply Chain Security

### SLSA Level 3

All releases include SLSA provenance attestations.

### Cosign Signing

- Binary releases signed with Cosign (keyless OIDC)
- Docker images signed with Cosign
- Checksums signed with Cosign

### Verification

```bash
# Verify binary signature
cosign verify-blob --bundle mcp-datahub.sigstore.json mcp-datahub

# Verify Docker image
cosign verify ghcr.io/txn2/mcp-datahub:latest
```

## Vulnerability Reporting

Report security vulnerabilities via:

1. [GitHub Security Advisories](https://github.com/txn2/mcp-datahub/security/advisories/new)
2. Email: security@txn2.com

**Do NOT report via public GitHub issues.**
