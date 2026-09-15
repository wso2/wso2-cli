//go:build !linux

// in gh actions, the keyring package fails to load on ubuntu runner as it requires an interactive terminal
// ignoring this test on linux for now
package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

func TestTokenStore(t *testing.T) {
	userStore := NewUserStore()
	orgStore := NewOrgStore()
	tokStore := NewOrgTokenStore(*orgStore)

	// Test StoreUserInfo and RetrieveUserInfo
	userInfo := &api.UserInfo{
		DisplayName:   "test-user",
		Organizations: []api.Organization{{ID: "test-org", Name: "Test Organization"}},
	}
	err := userStore.StoreUserInfo(*userInfo)
	assert.NoError(t, err, "StoreUserInfo failed")

	retrievedUserInfo, err := userStore.RetrieveUserInfo()
	assert.NoError(t, err, "RetrieveUserInfo failed")
	assert.Equal(t, userInfo, retrievedUserInfo, "Retrieved user info does not match stored user info")

	// Test StoreOrganizationInfo and RetrieveOrganizationInfo
	orgInfo := &api.Organization{ID: "test-org", Name: "Test Organization"}
	err = orgStore.SetDefaultOrg(orgInfo)
	assert.NoError(t, err, "StoreOrganizationInfo failed")

	retrievedOrgInfo, err := orgStore.GetDefaultOrg()
	assert.NoError(t, err, "RetrieveOrganizationInfo failed")
	assert.Equal(t, orgInfo, retrievedOrgInfo, "Retrieved organization info does not match stored organization info")

	// Test StoreToken and RetrieveToken
	token := &api.AccessToken{AccessToken: "test-token", RefreshToken: "Bearer", ExpirationTime: 3600}
	err = tokStore.StoreToken(region.GetCurrentRegion(), orgInfo.ID, token)
	assert.NoError(t, err, "StoreToken failed")

	retrievedToken, err := tokStore.RetrieveToken(region.GetCurrentRegion(), orgInfo.ID)
	assert.NoError(t, err, "RetrieveToken failed")
	assert.Equal(t, token, retrievedToken, "Retrieved token does not match stored token")

	// Test DeleteToken
	err = tokStore.DeleteToken(region.GetCurrentRegion(), orgInfo.ID)
	assert.NoError(t, err, "DeleteToken failed")

	_, err = tokStore.RetrieveToken(region.GetCurrentRegion(), orgInfo.ID)
	assert.Error(t, err, "RetrieveToken should fail after DeleteToken")

	// Test DeleteUserInfo
	err = userStore.ClearUserInfo()
	assert.NoError(t, err, "DeleteUserInfo failed")

	_, err = userStore.RetrieveUserInfo()
	assert.Error(t, err, "RetrieveUserInfo should fail after DeleteUserInfo")

	// Test DeleteOrganizationInfo
	err = orgStore.RemoveDefaultOrg()
	assert.NoError(t, err, "DeleteOrganizationInfo failed")

	retrievedOrgInfo, err = orgStore.GetDefaultOrg()
	assert.Error(t, err, "RetrieveOrganizationInfo failed")
	assert.Nil(t, retrievedOrgInfo, "Retrieved organization info should be nil after DeleteOrganizationInfo")
}
