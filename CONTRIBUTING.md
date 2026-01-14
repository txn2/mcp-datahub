# Contributing to mcp-datahub

Thank you for your interest in contributing to mcp-datahub!

## Development Setup

### Prerequisites

- Go 1.24 or later
- golangci-lint v2.7+
- gosec
- govulncheck

### Getting Started

```bash
git clone https://github.com/txn2/mcp-datahub.git
cd mcp-datahub
make tidy
make verify
```

## Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `make verify`
5. Commit with a descriptive message
6. Push and create a Pull Request

## Commit Messages

Follow conventional commits:
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

## Code Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` and `goimports`
- Keep functions focused and small
- Use meaningful variable names

### Error Handling

- Always wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Use sentinel errors for known error conditions
- Never ignore errors silently

### Testing

- Write table-driven tests
- Aim for >80% coverage
- Test both success and failure paths
- Use mocks for external dependencies

### Documentation

- Add godoc comments to all exported types and functions
- Keep comments up to date with code changes
- Include examples where helpful

## Pull Request Checklist

- [ ] Tests pass: `make test`
- [ ] Linting passes: `make lint`
- [ ] Security scans pass: `make security`
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventions

## Questions?

Open an issue or start a discussion on GitHub.
