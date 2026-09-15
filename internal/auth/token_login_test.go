package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTestJWT builds a minimal unsigned JWT with the given JSON payload map.
func makeTestJWT(t *testing.T, payload map[string]any) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	encoded := base64.RawURLEncoding.EncodeToString(body)
	return fmt.Sprintf("%s.%s.fakesig", header, encoded)
}

func TestParseSTSTokenClaims_Valid(t *testing.T) {
	token := makeTestJWT(t, map[string]any{
		"sub": "user-123",
		"organization": map[string]any{
			"handle": "my-org",
			"uuid":   "aaaabbbb-cccc-dddd-eeee-ffffffffffff",
		},
	})

	claims, err := parseSTSTokenClaims(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.Sub)
	assert.Equal(t, "my-org", claims.Organization.Handle)
	assert.Equal(t, "aaaabbbb-cccc-dddd-eeee-ffffffffffff", claims.Organization.UUID)
}

func TestParseSTSTokenClaims_EmptyOrgHandle(t *testing.T) {
	token := makeTestJWT(t, map[string]any{
		"sub":          "user-456",
		"organization": map[string]any{"handle": "", "uuid": "some-uuid"},
	})

	claims, err := parseSTSTokenClaims(token)
	require.NoError(t, err)
	assert.Equal(t, "user-456", claims.Sub)
	assert.Equal(t, "", claims.Organization.Handle)
}

func TestParseSTSTokenClaims_MissingOrganizationField(t *testing.T) {
	token := makeTestJWT(t, map[string]any{"sub": "user-789"})

	claims, err := parseSTSTokenClaims(token)
	require.NoError(t, err)
	assert.Equal(t, "", claims.Organization.Handle)
	assert.Equal(t, "", claims.Organization.UUID)
}

func TestParseSTSTokenClaims_NotAJWT(t *testing.T) {
	_, err := parseSTSTokenClaims("not-a-jwt")
	assert.ErrorContains(t, err, "invalid JWT format")
}

func TestParseSTSTokenClaims_TwoPartJWT(t *testing.T) {
	_, err := parseSTSTokenClaims("header.payload")
	assert.ErrorContains(t, err, "invalid JWT format")
}

func TestParseSTSTokenClaims_InvalidBase64Payload(t *testing.T) {
	_, err := parseSTSTokenClaims("header.!!!invalid!!!.sig")
	assert.Error(t, err)
}

func TestParseSTSTokenClaims_NonJSONPayload(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
	token := fmt.Sprintf("header.%s.sig", payload)
	_, err := parseSTSTokenClaims(token)
	assert.Error(t, err)
}
