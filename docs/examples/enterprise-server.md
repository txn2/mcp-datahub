# Example: Enterprise Server

Full enterprise setup with authentication, multi-tenant isolation, audit logging, and query context.

## Architecture

```mermaid
flowchart TB
    subgraph Clients
        ai[AI Assistants]
    end

    subgraph MCP Server
        auth[JWT Auth]
        tenant[Tenant Filter]
        audit[Audit Logger]
        rate[Rate Limiter]
        datahub[DataHub Tools]
        trino[Trino Tools]
    end

    subgraph Backend
        dh_api[DataHub API]
        tr_cluster[Trino Cluster]
        audit_db[(Audit DB)]
    end

    ai --> auth --> tenant --> rate --> datahub
    ai --> auth --> tenant --> rate --> trino
    datahub --> dh_api
    trino --> tr_cluster
    audit --> audit_db
```

## Complete Code

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "log"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/golang-jwt/jwt/v5"
    _ "github.com/lib/pq"
    "github.com/modelcontextprotocol/go-sdk/mcp"

    datahubclient "github.com/txn2/mcp-datahub/pkg/client"
    "github.com/txn2/mcp-datahub/pkg/integration"
    datahubtools "github.com/txn2/mcp-datahub/pkg/tools"

    trinoclient "github.com/txn2/mcp-trino/pkg/client"
    trinotools "github.com/txn2/mcp-trino/pkg/tools"
)

func main() {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "enterprise-data-server",
        Version: "1.0.0",
    }, nil)

    // Initialize database for audit logging
    auditDB, err := sql.Open("postgres", os.Getenv("AUDIT_DATABASE_URL"))
    if err != nil {
        log.Fatalf("Failed to connect to audit database: %v", err)
    }
    defer auditDB.Close()

    // DataHub client
    dhClient, err := datahubclient.New(datahubclient.Config{
        URL:   os.Getenv("DATAHUB_URL"),
        Token: os.Getenv("DATAHUB_TOKEN"),
    })
    if err != nil {
        log.Fatalf("Failed to create DataHub client: %v", err)
    }
    defer dhClient.Close()

    // Trino client
    trClient, err := trinoclient.New(trinoclient.Config{
        Host:    os.Getenv("TRINO_HOST"),
        Port:    443,
        User:    os.Getenv("TRINO_USER"),
        Catalog: os.Getenv("TRINO_CATALOG"),
    })
    if err != nil {
        log.Fatalf("Failed to create Trino client: %v", err)
    }
    defer trClient.Close()

    // Create middleware and integration components
    jwtAuth := NewJWTAuthMiddleware(os.Getenv("JWT_SECRET"))
    tenantFilter := NewTenantAccessFilter(loadTenantConfig())
    auditLogger := NewDBAuditLogger(auditDB)
    rateLimiter := NewRateLimiter(10, 50) // 10/sec, burst 50
    queryProvider := NewTrinoQueryProvider(trClient)

    // User ID extractor for audit logging
    getUserID := func(ctx context.Context) string {
        if id, ok := ctx.Value("user_id").(string); ok {
            return id
        }
        return "anonymous"
    }

    // DataHub toolkit with all enterprise features
    datahubToolkit := datahubtools.NewToolkit(dhClient,
        datahubtools.Config{
            DefaultLimit:    20,
            MaxLimit:        100,
            MaxLineageDepth: 5,
        },
        // Authentication
        datahubtools.WithMiddleware(jwtAuth),
        // Rate limiting
        datahubtools.WithMiddleware(rateLimiter),
        // Access control
        datahubtools.WithAccessFilter(tenantFilter),
        // Audit logging
        datahubtools.WithAuditLogger(auditLogger, getUserID),
        // Query context injection
        datahubtools.WithQueryProvider(queryProvider),
    )
    datahubToolkit.RegisterAll(server)

    // Trino toolkit
    trinoToolkit := trinotools.NewToolkit(trClient,
        trinotools.Config{
            MaxRows:      1000,
            QueryTimeout: 60,
        },
        trinotools.WithMiddleware(jwtAuth),
        trinotools.WithMiddleware(rateLimiter),
    )
    trinoToolkit.RegisterAll(server)

    log.Println("Starting enterprise data MCP server...")

    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}

// JWTAuthMiddleware validates JWT tokens and extracts user context
type JWTAuthMiddleware struct {
    secretKey []byte
}

func NewJWTAuthMiddleware(secret string) *JWTAuthMiddleware {
    return &JWTAuthMiddleware{secretKey: []byte(secret)}
}

func (m *JWTAuthMiddleware) Before(ctx context.Context, tc *datahubtools.ToolContext) (context.Context, error) {
    tokenString, ok := ctx.Value("auth_token").(string)
    if !ok || tokenString == "" {
        return ctx, errors.New("unauthorized: missing token")
    }

    tokenString = strings.TrimPrefix(tokenString, "Bearer ")

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
        return m.secretKey, nil
    })
    if err != nil || !token.Valid {
        return ctx, errors.New("unauthorized: invalid token")
    }

    claims := token.Claims.(jwt.MapClaims)
    ctx = context.WithValue(ctx, "user_id", claims["sub"])
    ctx = context.WithValue(ctx, "tenant_id", claims["tenant_id"])
    ctx = context.WithValue(ctx, "user_roles", claims["roles"])

    return ctx, nil
}

func (m *JWTAuthMiddleware) After(ctx context.Context, tc *datahubtools.ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
    return result, err
}

// TenantAccessFilter isolates data by tenant
type TenantAccessFilter struct {
    config map[string][]string // tenant -> allowed domains
}

func NewTenantAccessFilter(config map[string][]string) *TenantAccessFilter {
    return &TenantAccessFilter{config: config}
}

func (f *TenantAccessFilter) CanAccess(ctx context.Context, urn string) (bool, error) {
    tenantID, ok := ctx.Value("tenant_id").(string)
    if !ok {
        return false, errors.New("tenant context required")
    }

    allowedDomains := f.config[tenantID]
    for _, domain := range allowedDomains {
        if strings.Contains(urn, domain) {
            return true, nil
        }
    }
    return false, nil
}

func (f *TenantAccessFilter) FilterURNs(ctx context.Context, urns []string) ([]string, error) {
    var allowed []string
    for _, urn := range urns {
        ok, _ := f.CanAccess(ctx, urn)
        if ok {
            allowed = append(allowed, urn)
        }
    }
    return allowed, nil
}

func loadTenantConfig() map[string][]string {
    configJSON := os.Getenv("TENANT_CONFIG")
    if configJSON == "" {
        return map[string][]string{}
    }
    var config map[string][]string
    json.Unmarshal([]byte(configJSON), &config)
    return config
}

// DBAuditLogger logs to PostgreSQL
type DBAuditLogger struct {
    db *sql.DB
}

func NewDBAuditLogger(db *sql.DB) *DBAuditLogger {
    return &DBAuditLogger{db: db}
}

func (l *DBAuditLogger) LogToolCall(ctx context.Context, tool string, params map[string]any, userID string) error {
    paramsJSON, _ := json.Marshal(params)
    tenantID, _ := ctx.Value("tenant_id").(string)

    _, err := l.db.ExecContext(ctx, `
        INSERT INTO audit_log (timestamp, tool, user_id, tenant_id, params)
        VALUES ($1, $2, $3, $4, $5)
    `, time.Now().UTC(), tool, userID, tenantID, paramsJSON)

    return err
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
    mu       sync.Mutex
    buckets  map[string]*bucket
    rate     float64
    capacity int
}

type bucket struct {
    tokens    float64
    lastCheck time.Time
}

func NewRateLimiter(ratePerSecond float64, capacity int) *RateLimiter {
    return &RateLimiter{
        buckets:  make(map[string]*bucket),
        rate:     ratePerSecond,
        capacity: capacity,
    }
}

func (r *RateLimiter) Before(ctx context.Context, tc *datahubtools.ToolContext) (context.Context, error) {
    key := "global"
    if userID, ok := ctx.Value("user_id").(string); ok {
        key = userID
    }

    r.mu.Lock()
    defer r.mu.Unlock()

    b, exists := r.buckets[key]
    if !exists {
        r.buckets[key] = &bucket{tokens: float64(r.capacity) - 1, lastCheck: time.Now()}
        return ctx, nil
    }

    elapsed := time.Since(b.lastCheck).Seconds()
    b.tokens += elapsed * r.rate
    if b.tokens > float64(r.capacity) {
        b.tokens = float64(r.capacity)
    }
    b.lastCheck = time.Now()

    if b.tokens < 1 {
        return ctx, errors.New("rate limit exceeded")
    }
    b.tokens--

    return ctx, nil
}

func (r *RateLimiter) After(ctx context.Context, tc *datahubtools.ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
    return result, err
}

// TrinoQueryProvider provides query context from Trino
type TrinoQueryProvider struct {
    client *trinoclient.Client
}

func NewTrinoQueryProvider(client *trinoclient.Client) *TrinoQueryProvider {
    return &TrinoQueryProvider{client: client}
}

func (p *TrinoQueryProvider) Name() string {
    return "trino"
}

func (p *TrinoQueryProvider) ResolveTable(ctx context.Context, urn string) (*integration.TableIdentifier, error) {
    // Parse URN to extract table info
    // urn:li:dataset:(urn:li:dataPlatform:trino,catalog.schema.table,PROD)
    parts := parseDatasetURN(urn)
    if parts == nil {
        return nil, nil
    }
    return &integration.TableIdentifier{
        Catalog: parts.Catalog,
        Schema:  parts.Schema,
        Table:   parts.Table,
    }, nil
}

func (p *TrinoQueryProvider) GetTableAvailability(ctx context.Context, urn string) (*integration.TableAvailability, error) {
    table, err := p.ResolveTable(ctx, urn)
    if err != nil || table == nil {
        return nil, err
    }
    return &integration.TableAvailability{
        Available: true,
        Table:     table,
    }, nil
}

func (p *TrinoQueryProvider) GetQueryExamples(ctx context.Context, urn string) ([]integration.QueryExample, error) {
    table, _ := p.ResolveTable(ctx, urn)
    if table == nil {
        return nil, nil
    }

    tableName := table.String()
    return []integration.QueryExample{
        {Name: "Sample", SQL: "SELECT * FROM " + tableName + " LIMIT 10", Category: "sample"},
        {Name: "Count", SQL: "SELECT COUNT(*) FROM " + tableName, Category: "aggregation"},
    }, nil
}

func (p *TrinoQueryProvider) GetExecutionContext(ctx context.Context, urns []string) (*integration.ExecutionContext, error) {
    tables := make(map[string]*integration.TableIdentifier)
    for _, urn := range urns {
        if table, _ := p.ResolveTable(ctx, urn); table != nil {
            tables[urn] = table
        }
    }
    return &integration.ExecutionContext{
        Tables: tables,
        Source: "trino",
    }, nil
}

func (p *TrinoQueryProvider) Close() error {
    return nil
}

type urnParts struct {
    Catalog string
    Schema  string
    Table   string
}

func parseDatasetURN(urn string) *urnParts {
    // Simplified parsing
    // Real implementation would handle all URN formats
    return nil
}
```

## Configuration

```bash
# DataHub
export DATAHUB_URL=https://datahub.example.com
export DATAHUB_TOKEN=token

# Trino
export TRINO_HOST=trino.example.com
export TRINO_USER=service_account
export TRINO_CATALOG=hive

# Authentication
export JWT_SECRET=your-secret-key-at-least-32-bytes

# Audit Database
export AUDIT_DATABASE_URL=postgres://user:pass@host:5432/audit

# Tenant Configuration
export TENANT_CONFIG='{"acme":["acme","shared"],"globex":["globex","shared"]}'
```

## Database Schema

```sql
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    tool VARCHAR(100) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255),
    params JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_tenant ON audit_log(tenant_id);
```

## Features

| Feature | Component |
|---------|-----------|
| Authentication | JWTAuthMiddleware |
| Multi-tenancy | TenantAccessFilter |
| Audit logging | DBAuditLogger |
| Rate limiting | RateLimiter |
| Query context | TrinoQueryProvider |
| Metadata | DataHub tools |
| Query execution | Trino tools |

## Dependencies

```bash
go get github.com/txn2/mcp-datahub
go get github.com/txn2/mcp-trino
go get github.com/golang-jwt/jwt/v5
go get github.com/lib/pq
```

## Next Steps

- [Authentication Guide](../guides/authentication.md): Custom auth implementations
- [Multi-Tenant Guide](../guides/multi-tenant.md): Tenant isolation patterns
- [Audit Logging Guide](../guides/audit-logging.md): Compliance logging
