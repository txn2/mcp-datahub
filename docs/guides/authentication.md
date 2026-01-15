# How to Add Authentication

Add authentication to your custom MCP server.

## Goal

Secure your MCP server so only authenticated users can access DataHub tools.

## Prerequisites

- A working custom MCP server ([Building a Custom Server](../tutorials/building-custom-server.md))
- An authentication system (JWT, OAuth, or API keys)

## Option 1: JWT Authentication

### Step 1: Create JWT Middleware

```go
package auth

import (
    "context"
    "errors"
    "strings"

    "github.com/golang-jwt/jwt/v5"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/txn2/mcp-datahub/pkg/tools"
)

type JWTMiddleware struct {
    SecretKey []byte
    Issuer    string
}

func (m *JWTMiddleware) Before(ctx context.Context, tc *tools.ToolContext) (context.Context, error) {
    token := extractToken(ctx)
    if token == "" {
        return ctx, errors.New("unauthorized: missing token")
    }

    claims, err := m.validateToken(token)
    if err != nil {
        return ctx, err
    }

    // Add claims to context
    ctx = context.WithValue(ctx, "user_id", claims["sub"])
    ctx = context.WithValue(ctx, "user_email", claims["email"])
    ctx = context.WithValue(ctx, "user_roles", claims["roles"])

    return ctx, nil
}

func (m *JWTMiddleware) After(ctx context.Context, tc *tools.ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
    return result, err
}

func (m *JWTMiddleware) validateToken(tokenString string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return m.SecretKey, nil
    })

    if err != nil || !token.Valid {
        return nil, errors.New("unauthorized: invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("unauthorized: invalid claims")
    }

    if m.Issuer != "" && claims["iss"] != m.Issuer {
        return nil, errors.New("unauthorized: invalid issuer")
    }

    return claims, nil
}

func extractToken(ctx context.Context) string {
    if token, ok := ctx.Value("auth_token").(string); ok {
        return strings.TrimPrefix(token, "Bearer ")
    }
    return ""
}
```

### Step 2: Wire the Middleware

```go
jwtMiddleware := &auth.JWTMiddleware{
    SecretKey: []byte(os.Getenv("JWT_SECRET")),
    Issuer:    "your-auth-server",
}

toolkit := tools.NewToolkit(datahubClient,
    tools.WithMiddleware(jwtMiddleware),
)
```

## Option 2: OAuth 2.0 / OIDC

### Step 1: Create OAuth Middleware

```go
package auth

import (
    "context"
    "errors"

    "github.com/coreos/go-oidc/v3/oidc"
    "github.com/txn2/mcp-datahub/pkg/tools"
)

type OAuthMiddleware struct {
    verifier *oidc.IDTokenVerifier
}

func NewOAuthMiddleware(issuerURL, clientID string) (*OAuthMiddleware, error) {
    provider, err := oidc.NewProvider(context.Background(), issuerURL)
    if err != nil {
        return nil, err
    }

    verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
    return &OAuthMiddleware{verifier: verifier}, nil
}

func (m *OAuthMiddleware) Before(ctx context.Context, tc *tools.ToolContext) (context.Context, error) {
    token := extractToken(ctx)
    if token == "" {
        return ctx, errors.New("unauthorized: missing token")
    }

    idToken, err := m.verifier.Verify(ctx, token)
    if err != nil {
        return ctx, errors.New("unauthorized: invalid token")
    }

    var claims struct {
        Email  string   `json:"email"`
        Groups []string `json:"groups"`
    }
    if err := idToken.Claims(&claims); err != nil {
        return ctx, err
    }

    ctx = context.WithValue(ctx, "user_id", idToken.Subject)
    ctx = context.WithValue(ctx, "user_email", claims.Email)
    ctx = context.WithValue(ctx, "user_groups", claims.Groups)

    return ctx, nil
}
```

### Step 2: Configure for Your Provider

**Okta:**
```go
middleware, _ := auth.NewOAuthMiddleware(
    "https://your-org.okta.com",
    "your-client-id",
)
```

**Auth0:**
```go
middleware, _ := auth.NewOAuthMiddleware(
    "https://your-tenant.auth0.com/",
    "your-client-id",
)
```

**Google:**
```go
middleware, _ := auth.NewOAuthMiddleware(
    "https://accounts.google.com",
    "your-client-id.apps.googleusercontent.com",
)
```

## Option 3: API Key Authentication

### Step 1: Create API Key Middleware

```go
package auth

import (
    "context"
    "crypto/subtle"
    "errors"

    "github.com/txn2/mcp-datahub/pkg/tools"
)

type APIKeyMiddleware struct {
    keys map[string]APIKeyInfo
}

type APIKeyInfo struct {
    UserID string
    Roles  []string
}

func NewAPIKeyMiddleware(keys map[string]APIKeyInfo) *APIKeyMiddleware {
    return &APIKeyMiddleware{keys: keys}
}

func (m *APIKeyMiddleware) Before(ctx context.Context, tc *tools.ToolContext) (context.Context, error) {
    apiKey := extractAPIKey(ctx)
    if apiKey == "" {
        return ctx, errors.New("unauthorized: missing API key")
    }

    info, valid := m.validateKey(apiKey)
    if !valid {
        return ctx, errors.New("unauthorized: invalid API key")
    }

    ctx = context.WithValue(ctx, "user_id", info.UserID)
    ctx = context.WithValue(ctx, "user_roles", info.Roles)

    return ctx, nil
}

func (m *APIKeyMiddleware) validateKey(key string) (APIKeyInfo, bool) {
    for storedKey, info := range m.keys {
        if subtle.ConstantTimeCompare([]byte(key), []byte(storedKey)) == 1 {
            return info, true
        }
    }
    return APIKeyInfo{}, false
}

func extractAPIKey(ctx context.Context) string {
    if key, ok := ctx.Value("api_key").(string); ok {
        return key
    }
    return ""
}
```

### Step 2: Load Keys from Configuration

```go
// Load from environment or config file
keys := map[string]auth.APIKeyInfo{
    os.Getenv("API_KEY_USER1"): {UserID: "user1", Roles: []string{"read"}},
    os.Getenv("API_KEY_ADMIN"): {UserID: "admin", Roles: []string{"read", "admin"}},
}

middleware := auth.NewAPIKeyMiddleware(keys)
```

## Verification

Test that authentication is working:

1. **Without token**: Should receive "unauthorized" error
2. **With invalid token**: Should receive "invalid token" error
3. **With valid token**: Should access tools normally

```go
// Test helper
func TestAuthentication(t *testing.T) {
    // Create context without token
    ctx := context.Background()
    _, err := middleware.Before(ctx, &tools.ToolContext{})
    if err == nil {
        t.Error("Expected error for missing token")
    }

    // Create context with valid token
    ctx = context.WithValue(ctx, "auth_token", validToken)
    ctx, err = middleware.Before(ctx, &tools.ToolContext{})
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    userID := ctx.Value("user_id")
    if userID == nil {
        t.Error("Expected user_id in context")
    }
}
```

## Troubleshooting

**"unauthorized: missing token"**

- Verify the token is being passed in the context
- Check the context key name matches your transport layer

**"unauthorized: invalid token"**

- Verify the secret key or OIDC configuration
- Check token expiration
- Verify the issuer matches

**Token validation is slow**

- Cache OIDC provider configuration
- Use JWT with local validation instead of introspection

## Next Steps

- [Multi-Tenant Setup](multi-tenant.md): Add tenant isolation
- [Audit Logging](audit-logging.md): Log authenticated actions
