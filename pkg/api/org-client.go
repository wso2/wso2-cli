package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type OrgClient struct {
	orgApiUrl string
	client    *IPHTTPClient
}

func NewOrgClient(orgApiURL string, tokenStore ReadOnlyTokenStore) *OrgClient {
	return &OrgClient{orgApiUrl: orgApiURL, client: NewIPHTTPClient(tokenStore)}
}

// GetOrganizations lists all organizations for the authenticated user.
// Authentication uses the active session token — no org ID needed.
func (c *OrgClient) GetOrganizations() ([]Organization, error) {
	req, _ := http.NewRequest("GET", c.orgApiUrl, nil)
	resp, err := c.client.DoForActiveOrg(req)
	if err != nil {
		return nil, fmt.Errorf("error while fetching user organizations: %w", err)
	}
	defer resp.Body.Close()
	var orgs []Organization
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetOrganization returns the organization matching the given UUID.
func (c *OrgClient) GetOrganization(orgUUID string) (*Organization, error) {
	orgs, err := c.GetOrganizations()
	if err != nil {
		return nil, err
	}
	for _, org := range orgs {
		if org.UUID == orgUUID {
			return &org, nil
		}
	}
	return nil, fmt.Errorf("target organization not found")
}

func (c *OrgClient) GetOrgApiUrl() string {
	return c.orgApiUrl
}
