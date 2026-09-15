package auth

import (
	"github.com/wso2/integration-platform-tools/pkg/api"
)

type ReadOnlyTokenStore struct {
}

func (r *ReadOnlyTokenStore) GetTokenForOrg(orgID string, forceSignIn bool) (string, error) {
	// synchronize access to avoid concurrent refreshes of the same token
	tokenRefreshMutex.Lock()
	defer tokenRefreshMutex.Unlock()
	return r.getToken(orgID, forceSignIn, nil)
}

func (r *ReadOnlyTokenStore) GetTokenForOrg2(orgId string, forceSignIn bool, orgs []api.Organization) (string, error) {
	// synchronize access to the token retrival
	// to avoid concurrent refreshes of the same token
	tokenRefreshMutex.Lock()
	defer tokenRefreshMutex.Unlock()
	token, err := r.getToken(orgId, forceSignIn, orgs)
	return token, err
}

// GetToken - Get the access token
func (r *ReadOnlyTokenStore) GetTokenForActiveOrg() (string, error) {
	// Read the stored org directly from the keyring — avoids calling GetSelectedOrganization()
	// which calls OrgClient.GetOrganizations() → DoForActiveOrg() → GetTokenForActiveOrg() (infinite loop).
	org, err := orgStore.GetDefaultOrg()
	if err != nil {
		return "", err
	}
	return r.GetTokenForOrg(org.ID, false)
}

func (r *ReadOnlyTokenStore) GetActiveOrg() (*api.Organization, error) {
	return GetSelectedOrganization()
}
