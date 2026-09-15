package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/wso2/integration-platform-tools/internal/auth"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// AuthTokenKey is the context key for the authentication token
	AuthTokenKey ContextKey = "auth_token"
	// ClientFactoryKey is the context key for the client factory
	ClientFactoryKey ContextKey = "client_factory"
	// HttpModeKey is the context key for HTTP mode
	HttpModeKey ContextKey = "http_mode"
)

// ContextTokenStore is a token store that retrieves tokens from context
// This is used for HTTP mode where tokens are provided per request
type ContextTokenStore struct {
	ctx context.Context
}

// NewContextTokenStore creates a new context-aware token store
func NewContextTokenStore(ctx context.Context) *ContextTokenStore {
	return &ContextTokenStore{
		ctx: ctx,
	}
}

// GetTokenForOrg retrieves the token for a specific organization from context
func (c *ContextTokenStore) GetTokenForOrg(orgID string, forceSignIn bool) (string, error) {
	// Extract token from context
	token, ok := c.ctx.Value(AuthTokenKey).(string)
	if !ok || token == "" {
		return "", fmt.Errorf("no authentication token found in context")
	}

	// Remove "Bearer " prefix if present
	if after, ok0 := strings.CutPrefix(token, "Bearer "); ok0 {
		token = after
	}

	// In HTTP mode, we don't validate the orgID against the token
	// as the token should be valid for the requested org
	return token, nil
}

// GetTokenForActiveOrg retrieves the token for the active organization from context
func (c *ContextTokenStore) GetTokenForActiveOrg() (string, error) {
	return c.GetTokenForOrg("", false)
}

// SetTokenInContext creates a new context with the provided token
func SetTokenInContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, AuthTokenKey, token)
}

// GetTokenFromContext extracts the token from context
func GetTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(AuthTokenKey).(string)
	return token, ok
}

// SetHttpModeInContext creates a new context with http_mode set to true
func SetHttpModeInContext(ctx context.Context, httpMode bool) context.Context {
	return context.WithValue(ctx, HttpModeKey, httpMode)
}

// IsHttpMode checks if the context is in HTTP mode
func IsHttpMode(ctx context.Context) bool {
	httpMode, ok := ctx.Value(HttpModeKey).(bool)
	return ok && httpMode
}

// SetClientFactoryInContext creates a new context with the provided client factory
func SetClientFactoryInContext(ctx context.Context, clientFactory *auth.ClientFactory) context.Context {
	return context.WithValue(ctx, ClientFactoryKey, clientFactory)
}

// GetClientFactoryFromContext extracts the client factory from context
// Returns nil if no client factory is found (e.g., in stdio mode)
func GetClientFactoryFromContext(ctx context.Context) *auth.ClientFactory {
	clientFactory, ok := ctx.Value(ClientFactoryKey).(*auth.ClientFactory)
	if !ok {
		return nil
	}
	return clientFactory
}

// NewContextClientFactory creates a client factory from context
// This is MCP-specific and used only in HTTP mode
func NewContextClientFactory(ctx context.Context) *auth.ClientFactory {
	contextTokenStore := NewContextTokenStore(ctx)
	return auth.NewClientFactory(contextTokenStore)
}
