# Troubleshooting

Common issues and solutions for mcp-datahub.

## Connection Issues

### "unauthorized: invalid or missing token"

**Cause**: Invalid or expired DataHub token.

**Solution**:
1. Verify `DATAHUB_TOKEN` is set correctly
2. Generate a new token from DataHub Settings > Access Tokens
3. Ensure the token hasn't expired

### "connection refused"

**Cause**: Cannot reach DataHub server.

**Solution**:
1. Verify `DATAHUB_URL` is correct
2. Check network connectivity to DataHub
3. Ensure DataHub is running and accessible

### "certificate verify failed"

**Cause**: TLS certificate issue.

**Solution**:
1. Ensure DataHub uses a valid SSL certificate
2. Check if corporate proxy is intercepting traffic
3. Verify the certificate chain is complete

## Search Issues

### "no results found"

**Cause**: Query doesn't match any entities.

**Solution**:
1. Try broader search terms
2. Remove entity type filter
3. Verify entities exist in DataHub

### "too many results"

**Cause**: Query is too broad.

**Solution**:
1. Add more specific search terms
2. Use entity type filter: `entity_type: "DATASET"`
3. Reduce limit parameter

## Lineage Issues

### "lineage depth exceeded"

**Cause**: Requested depth exceeds configured maximum.

**Solution**:
1. Reduce depth parameter
2. Increase `DATAHUB_MAX_LINEAGE_DEPTH` if needed

### "no lineage found"

**Cause**: Entity has no upstream/downstream dependencies.

**Solution**:
1. Verify lineage exists in DataHub UI
2. Check if lineage is ingested for this platform

## Performance Issues

### Slow responses

**Cause**: Large result sets or deep lineage traversal.

**Solution**:
1. Reduce search limits
2. Reduce lineage depth
3. Increase timeout: `DATAHUB_TIMEOUT=60`

### Timeouts

**Cause**: Request exceeds timeout limit.

**Solution**:
1. Increase timeout value
2. Reduce query complexity
3. Check DataHub server performance

## Configuration Issues

### "invalid configuration"

**Cause**: Missing or invalid configuration values.

**Solution**:
1. Verify all required env vars are set
2. Check for typos in variable names
3. Ensure numeric values are valid

### Environment variables not loaded

**Cause**: Variables not exported or shell not reloaded.

**Solution**:
```bash
# Export variables
export DATAHUB_URL=https://datahub.company.com
export DATAHUB_TOKEN=your_token

# Or use .env file with direnv
```

## Debug Logging

Enable debug logging to see detailed information about requests, responses, and errors.

### Enable Debug Mode

```bash
export DATAHUB_DEBUG=1
./mcp-datahub
```

### What Debug Logging Shows

- GraphQL operation names and request sizes
- HTTP response status codes and sizes
- Request duration and retry attempts
- Detailed error messages with context
- Connection selection in multi-server mode

### Example Debug Output

```
[datahub] DEBUG: executing GraphQL query [operation=GetEntity endpoint=https://datahub.example.com/api/graphql request_size=256]
[datahub] DEBUG: received response [status=200 response_size=1024]
[datahub] DEBUG: request completed [operation=GetEntity duration_ms=150 attempts=1]
```

### Common Debug Scenarios

**Silent failures**: If a query returns empty results unexpectedly, debug logging will show if GraphQL returned null data without errors.

**Retry behavior**: See when and how often requests are retried, including backoff timing.

**Auth issues**: Debug logs show when requests fail with 401/403 status codes before the error is returned.

## Getting Help

If you're still having issues:

1. Check [GitHub Issues](https://github.com/txn2/mcp-datahub/issues)
2. Search existing discussions
3. Open a new issue with:
   - mcp-datahub version
   - DataHub version
   - Error messages
   - Steps to reproduce
