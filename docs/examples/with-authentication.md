# Example: With Authentication

MCP server with JWT authentication and access control.

## Complete Code

```go
package main

import (
    "context"
    "errors"
    "log"
    "os"
    "strings"

    "github.com/golang-jwt/jwt/v5"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/txn2/mcp-datahub/pkg/client"
    "github.com/txn2/mcp-datahub/pkg/tools"
)

func main() {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "authenticated-datahub-server",
        Version: "1.0.0",
    }, nil)

    datahubClient, err := client.New(client.Config{
        URL:   os.Getenv("DATAHUB_URL"),
        Token: os.Getenv("DATAHUB_TOKEN"),
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer datahubClient.Close()

    // Create authentication middleware
    jwtSecret := os.Getenv("JWT_SECRET")
    authMiddleware := &JWTAuthMiddleware{secretKey: []byte(jwtSecret)}

    // Create access filter
    accessFilter := &RoleBasedAccessFilter{
        adminRoles: []string{"admin", "data-steward"},
    }

    // Create toolkit with auth and access control
    toolkit := tools.NewToolkit(datahubClient,
        tools.WithMiddleware(authMiddleware),
        tools.WithAccessFilter(accessFilter),
    )
    toolkit.RegisterAll(server)

    log.Println("Starting authenticated DataHub MCP server...")

    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}

// JWTAuthMiddleware validates JWT tokens
type JWTAuthMiddleware struct {
    secretKey []byte
}

func (m *JWTAuthMiddleware) Before(ctx context.Context, tc *tools.ToolContext) (context.Context, error) {
    tokenString, ok := ctx.Value("auth_token").(string)
    if !ok || tokenString == "" {
        return ctx, errors.New("unauthorized: missing token")
    }

    tokenString = strings.TrimPrefix(tokenString, "Bearer ")

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return m.secretKey, nil
    })

    if err != nil || !token.Valid {
        return ctx, errors.New("unauthorized: invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return ctx, errors.New("unauthorized: invalid claims")
    }

    // Extract user info
    userID := claims["sub"].(string)
    ctx = context.WithValue(ctx, "user_id", userID)

    if email, ok := claims["email"].(string); ok {
        ctx = context.WithValue(ctx, "user_email", email)
    }

    if roles, ok := claims["roles"].([]any); ok {
        var roleStrings []string
        for _, r := range roles {
            roleStrings = append(roleStrings, r.(string))
        }
        ctx = context.WithValue(ctx, "user_roles", roleStrings)
    }

    return ctx, nil
}

func (m *JWTAuthMiddleware) After(ctx context.Context, tc *tools.ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
    return result, err
}

// RoleBasedAccessFilter controls access based on user roles
type RoleBasedAccessFilter struct {
    adminRoles []string
}

func (f *RoleBasedAccessFilter) CanAccess(ctx context.Context, urn string) (bool, error) {
    roles, ok := ctx.Value("user_roles").([]string)
    if !ok {
        return false, nil
    }

    // Admins can access everything
    for _, role := range roles {
        for _, adminRole := range f.adminRoles {
            if role == adminRole {
                return true, nil
            }
        }
    }

    // Check domain-based access
    // In a real implementation, query DataHub for entity domain
    // and check against user's allowed domains
    return true, nil
}

func (f *RoleBasedAccessFilter) FilterURNs(ctx context.Context, urns []string) ([]string, error) {
    var allowed []string
    for _, urn := range urns {
        ok, err := f.CanAccess(ctx, urn)
        if err != nil {
            return nil, err
        }
        if ok {
            allowed = append(allowed, urn)
        }
    }
    return allowed, nil
}
```

## Configuration

```bash
export DATAHUB_URL=https://your-datahub.example.com
export DATAHUB_TOKEN=your_token
export JWT_SECRET=your-secret-key-at-least-32-bytes
```

## Generating Test Tokens

```go
package main

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

func main() {
    secret := "your-secret-key-at-least-32-bytes"

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":   "user123",
        "email": "user@example.com",
        "roles": []string{"data-analyst", "sales-team"},
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, _ := token.SignedString([]byte(secret))
    fmt.Println(tokenString)
}
```

## Token Claims

Expected JWT claims:

| Claim | Type | Description |
|-------|------|-------------|
| `sub` | string | User ID |
| `email` | string | User email (optional) |
| `roles` | []string | User roles |
| `exp` | int | Expiration timestamp |

## Dependencies

```bash
go get github.com/golang-jwt/jwt/v5
```

## Next Steps

- [Combined Trino](combined-trino.md): Add Trino tools
- [Enterprise Server](enterprise-server.md): Full enterprise setup
