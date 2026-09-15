package auth

import (
	"encoding/json"
	"fmt"

	"github.com/wso2/integration-platform-tools/pkg/api"
	clikeyring "github.com/wso2/integration-platform-tools/pkg/util/keyring"
)

const (
	ORG_ENTRY_KEY = "organization_info"
)

var (
	KeyringEntryNotFound = fmt.Errorf("keyring entry not found")
	OrgResolvingError    = fmt.Errorf("error resolving organization")
)

type OrgStore struct{}

func NewOrgStore() *OrgStore {
	return &OrgStore{}
}

func (o *OrgStore) GetOrgs() ([]api.Organization, error) {
	cOrg, err := o.GetDefaultOrg()

	if err != nil {
		return nil, err
	}

	orgs, err := OrgClient.GetOrganizations()

	if err != nil {
		return []api.Organization{*cOrg}, err
	}

	return orgs, nil
}

func (o *OrgStore) GetOrgById(id string) (*api.Organization, error) {
	orgs, err := o.GetOrgs()
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
		if org.ID == id {
			return &org, nil
		}
	}

	return nil, OrgResolvingError
}

func (o *OrgStore) SetDefaultOrg(org *api.Organization) error {
	orgInfoBytes, err := json.Marshal(org)

	if err != nil {
		return err
	}

	err = clikeyring.Set(serviceName, ORG_ENTRY_KEY, string(orgInfoBytes))

	if err != nil {
		return err
	}

	return nil
}

func (o *OrgStore) GetDefaultOrg() (*api.Organization, error) {
	orgInfo, err := clikeyring.Get(serviceName, ORG_ENTRY_KEY)
	if err != nil {
		return nil, KeyringEntryNotFound
	}

	var org api.Organization

	if err := json.Unmarshal([]byte(orgInfo), &org); err != nil {
		return nil, err
	}

	return &org, nil
}

func (o *OrgStore) RemoveDefaultOrg() error {
	return clikeyring.Delete(serviceName, ORG_ENTRY_KEY)
}
