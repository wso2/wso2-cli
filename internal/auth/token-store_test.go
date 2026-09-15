package auth

import (
	"testing"

	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

func TestOrgTokenStore(t *testing.T) {
	orgStore := NewOrgStore()
	orgInfo := &api.Organization{ID: "test-org", Name: "Test Organization"}
	orgStore.SetDefaultOrg(orgInfo)
	tokenStore := NewOrgTokenStore(*orgStore)

	// Test StoreToken and RetrieveToken
	token := &api.AccessToken{
		AccessToken:    "test-token",
		RefreshToken:   "Bearer",
		ExpirationTime: 3600,
	}

	err := tokenStore.StoreToken(region.GetCurrentRegion(), orgInfo.ID, token)
	if err != nil {
		t.Error("StoreToken failed")
	}

	retrievedToken, err := tokenStore.RetrieveToken(region.GetCurrentRegion(), orgInfo.ID)
	if err != nil {
		t.Error("RetrieveToken failed")
	}

	if token.AccessToken != retrievedToken.AccessToken {
		t.Error("Retrieved AccessToken does not match stored AccessToken")
	}

	if token.RefreshToken != retrievedToken.RefreshToken {
		t.Error("Retrieved RefreshToken does not match stored RefreshToken")
	}

	if token.ExpirationTime != retrievedToken.ExpirationTime {
		t.Error("Retrieved ExpirationTime does not match stored ExpirationTime")
	}

	// Test DeleteToken
	err = tokenStore.DeleteToken(region.GetCurrentRegion(), orgInfo.ID)
	if err != nil {
		t.Error("DeleteToken failed")
	}

	tok, err := tokenStore.RetrieveToken(region.GetCurrentRegion(), orgInfo.ID)

	if err == nil {
		t.Error("RetrieveToken should fail after DeleteToken")
	}

	if tok != nil {
		t.Error("RetrieveToken should return nil after DeleteToken")
	}

}
