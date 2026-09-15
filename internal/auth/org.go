package auth

import (
	"fmt"

	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

func SetSelectedOrg(org *api.Organization, orgs []api.Organization) (err error) {
	_, err = rOTokenStore.GetTokenForOrg2(org.ID, true, orgs)
	if err != nil {
		return fmt.Errorf("failed to get token for new org: %w", err)
	}
	err = orgStore.SetDefaultOrg(org)
	if err != nil {
		return fmt.Errorf("failed to set selected org: %w", err)
	}
	return nil
}

func SetInitToken(env string, orgId string) error {
	token := &api.AccessToken{AccessToken: env}
	claims := token.GetDecodedToken()
	if claims == nil {
		return fmt.Errorf("failed to decode token")
	}

	if token.IsExpired() {
		newToken, err := authClient.ExchangeSTSToken(token.AccessToken, claims.Organization.Handle)
		if err != nil {
			return err
		}
		token = newToken
	}

	defaultOrg := &api.Organization{ID: orgId}

	err := orgStore.SetDefaultOrg(defaultOrg)
	if err != nil {
		return err
	}

	err = tokenStore.StoreToken(region.GetCurrentRegion(), defaultOrg.ID, token)
	if err != nil {
		return err
	}

	orgs, err := orgStore.GetOrgs()
	if err != nil {
		return err
	}
	for _, item := range orgs {
		if item.Handle == claims.Organization.Handle {
			defaultOrg = &item
			err := orgStore.SetDefaultOrg(defaultOrg)
			if err != nil {
				return err
			}
			break
		}
	}

	err = tokenStore.StoreToken(region.GetCurrentRegion(), defaultOrg.ID, token)
	if err != nil {
		return err
	}

	_, err = usrStore.GenerateUserInfoFromToken(env, orgId)
	if err != nil {
		return err
	}
	return nil
}

// get last selected organization or the default organization if no organization is selected previously
// default org can be the first org in the list of orgs
func GetSelectedOrganization() (org *api.Organization, err error) {
	var lastOrg *api.Organization
	var userInfo *api.UserInfo
	var orgs []api.Organization

	if lastOrg, err = orgStore.GetDefaultOrg(); err != nil {
		return
	}

	if userInfo, err = usrStore.RetrieveUserInfo(); err != nil {
		return
	}

	if lastOrg != nil {
		// return the last selected org if the current user is a member of the org
		orgs, err = OrgClient.GetOrganizations()
		if err != nil {
			return
		}

		for _, currentOrg := range orgs {
			if currentOrg.ID == lastOrg.ID {
				org = &currentOrg
				return
			}
		}
	}

	// if not return the first org in the list of orgs
	if len(userInfo.Organizations) == 0 {
		err = fmt.Errorf("no organizations found for the user")
		return
	}
	// if not return the first org in the list of orgs
	// and store it as the last selected org
	if err = orgStore.SetDefaultOrg(&userInfo.Organizations[0]); err != nil {
		return
	}
	org = &userInfo.Organizations[0]
	return
}

// get a org by orgID
func GetOrg(orgID string) (org *api.Organization, err error) {
	var userInfo *api.UserInfo

	if userInfo, err = usrStore.RetrieveUserInfo(); err != nil {
		return
	}
	for _, currentOrg := range userInfo.Organizations {
		if currentOrg.ID == orgID {
			org = &currentOrg
			return
		}
	}
	err = fmt.Errorf("no organization found with ID %s", orgID)
	return
}

// Does the user have access to the org
func HasAccessToOrg(orgID string) (hasAccess bool, err error) {
	orgs, err := OrgClient.GetOrganizations()

	if err != nil {
		return
	}

	for _, currentOrg := range orgs {
		if currentOrg.ID == orgID {
			hasAccess = true
			return
		}
	}

	return
}
