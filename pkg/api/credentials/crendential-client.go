package credentials

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type CredentialClient struct {
	gqlUrl string
	client *api.IPHTTPClient
}

func NewCredentialClient(gqlEp string, tokenStore api.ReadOnlyTokenStore) *CredentialClient {
	return &CredentialClient{
		gqlUrl: gqlEp,
		client: api.NewIPHTTPClient(tokenStore),
	}
}

func (cc *CredentialClient) GetCommonCredentials(orgId, orgUUID string) ([]CredentialEntry, error) {
	q := getCommonCredentialQuery(orgUUID)

	req, err := http.NewRequest("POST", cc.gqlUrl, bytes.NewBuffer([]byte(q)))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := cc.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			CommonCredentials []CredentialEntry `json:"commonCredentials"`
		} `json:"data"`
	}

	var resp Response
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return resp.Data.CommonCredentials, nil
}

func (c *CredentialClient) GetCommonCredentialsDetails(orgId, orgUuid, credentialId string) (resp CredentialEntry, err error) {
	query := getCommonCredentialDetailsQuery(orgUuid, credentialId)

	req, err := http.NewRequest("POST", c.gqlUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		err = fmt.Errorf("error while creating request: %w", err)
		return
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		err = fmt.Errorf("error while executing request: %w", err)
		return
	}

	defer res.Body.Close()

	var response struct {
		Data struct {
			CommonCredential CredentialEntry `json:"commonCredential"`
		} `json:"data"`
	}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		err = fmt.Errorf("error while decoding response: %w", err)
		return
	}

	resp = response.Data.CommonCredential

	return
}
