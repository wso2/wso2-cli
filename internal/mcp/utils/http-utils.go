package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/wso2/integration-platform-tools/internal/auth"
)

func AuthFromRequest(ctx context.Context, r *http.Request) context.Context {
	ctx = SetHttpModeInContext(ctx, true)

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ctx
	}
	token := extractTokenFromHeader(authHeader)

	ctx = SetTokenInContext(ctx, token)
	return SetClientFactoryInContext(ctx, NewContextClientFactory(ctx))
}

func TokenFromContext(ctx context.Context) (string, error) {
	auth, ok := GetTokenFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing auth")
	}
	if auth == "" {
		return "", fmt.Errorf("empty auth token")
	}
	return auth, nil
}

// extractTokenFromAuth extracts the token from the Authorization header
func extractTokenFromHeader(authHeader string) string {
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return authHeader
}

// TokenClaims represents the JWT payload structure
type TokenClaims struct {
	Organization struct {
		Handle string `json:"handle"`
		UUID   string `json:"uuid"`
	} `json:"organization"`
	Sub           string   `json:"sub"`
	Aud           []string `json:"aud"`
	Iss           string   `json:"iss"`
	Exp           int64    `json:"exp"`
	Iat           int64    `json:"iat"`
	Organizations []string `json:"organizations"`
}

func ExtractOrgFromToken(token string) (uuid string, err error) {
	// Retrieve Organization by calling the PAT Introspection endpoint for PAT tokens
	if strings.HasPrefix(token, "chp_") {
		patInfo, err := auth.GetPATInfo(token)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve org uuid from PAT token: %w", err)
		}

		return patInfo.OrganizationUUID, nil
	}

	// Retrieve Organization by decoding the token for bearer tokens
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	payload := parts[1]

	// Add padding if necessary for base64 decoding
	for len(payload)%4 != 0 {
		payload += "="
	}

	decodedPayload, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims TokenClaims
	if err := json.Unmarshal(decodedPayload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse claims: %w", err)
	}

	return claims.Organization.UUID, nil
}
