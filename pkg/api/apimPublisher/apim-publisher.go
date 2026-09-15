package apimpublisher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type ApimPublisherClient struct {
	apimPublisherBaseUrl string
	client               *api.IPHTTPClient
}

type ApiKeyResponse struct {
	Apikey       string `json:"apikey"`
	ValidityTime int    `json:"validityTime"`
}

func NewApimPublisherClient(apimPublisherBaseUrl string, tokenStore api.ReadOnlyTokenStore) *ApimPublisherClient {
	return &ApimPublisherClient{
		apimPublisherBaseUrl: apimPublisherBaseUrl,
		client:               api.NewIPHTTPClient(tokenStore),
	}
}

func (c *ApimPublisherClient) GetApiKey(apimId string, orgUuid string, orgId string, envKey string) (*ApiKeyResponse, error) {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/apis/%s/generate-key?organizationId=%s&keyType=%s",
			c.apimPublisherBaseUrl,
			apimId,
			orgUuid,
			envKey,
		),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching api key: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response *ApiKeyResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *ApimPublisherClient) GetSwaggerSpec(apimRevisionId string, orgUuid string, orgId string) (any, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/apis/%s/swagger?organizationId=%s",
			c.apimPublisherBaseUrl,
			apimRevisionId,
			orgUuid,
		),
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching swagger spec: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response any

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *ApimPublisherClient) UpdateApiProxyKeys(orgId, apiId, envId, orgUuid, targetUrl, sandboxUrl string) (err error) {
	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf(
			"%s/apis/%s/environments/%s/keys?organizationId=%s",
			c.apimPublisherBaseUrl,
			apiId,
			envId,
			orgUuid,
		),
		bytes.NewBuffer([]byte(
			fmt.Sprintf(`{"productionEndpoint":"%s","sandboxEndpointChoreo":"%s"}`, targetUrl, sandboxUrl),
		)),
	)

	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while fetching swagger spec: %w", err)
	}

	resp.Body.Close()

	return
}
