# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

**Do NOT report security vulnerabilities through public GitHub issues.**

Report via:
1. [GitHub Security Advisories](https://github.com/txn2/mcp-datahub/security/advisories/new)
2. Email: cj@imti.co

### What to Expect

- Acknowledgment within 48 hours
- Fix within 90 days
- Credit in release notes

### Security Best Practices

1. **Credentials**: Use environment variables, never commit secrets
2. **Network**: Always use HTTPS for DataHub connections
3. **Access Control**: Use DataHub's built-in authorization
4. **Logging**: Monitor for unusual query patterns

## Security Features

- Query limits prevent excessive data transfer
- Timeouts prevent long-running operations
- SSL/TLS support with certificate verification
- Token redaction in logs
