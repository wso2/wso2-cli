package configmapping

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type ConfigMappingSvcClient struct {
	configMapBaseApiBaseUrl string
	configSvcBaseApiUrl     string
	client                  *api.IPHTTPClient
}

func NewConfigMappingSvcClient(configMapBaseApiBaseUrl string, configSvcBaseApiUrl string, tokenStore api.ReadOnlyTokenStore) *ConfigMappingSvcClient {
	return &ConfigMappingSvcClient{
		configMapBaseApiBaseUrl: configMapBaseApiBaseUrl,
		configSvcBaseApiUrl:     configSvcBaseApiUrl,
		client:                  api.NewIPHTTPClient(tokenStore),
	}
}

func (c *ConfigMappingSvcClient) ResolveSecrets(
	orgId string,
	groupId string,
	data ResolveSecretsReq,
) (*ResolveSecretsResponse, error) {
	reqBodyBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/configs/groups/%s/resolve-secrets", c.configSvcBaseApiUrl, groupId), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response ResolveSecretsResponse

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *ConfigMappingSvcClient) GetDeployedConfigMappings(
	orgId string,
	projectId string,
	componentId string,
	envTemplateId string,
	deploymentTrackId string,
) (*ConfigMappingResponse, error) {
	values := url.Values{}
	values.Add("projectId", projectId)
	values.Add("componentId", componentId)
	values.Add("envTemplateId", envTemplateId)
	values.Add("deploymentTrackId", deploymentTrackId)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/configs/mappings/deploy?%s", c.configMapBaseApiBaseUrl, values.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching config mappings services: %w", err)
	}

	defer resp.Body.Close()

	var response ConfigMappingResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *ConfigMappingSvcClient) CreateConfigMapping(
	orgId string,
	projectId string,
	componentId string,
	envTemplateId string,
	deploymentTrackId string,
	configs []ConfigMapping,
) error {
	params := CreateConfigMappingReq{
		ProjectID:         projectId,
		ComponentID:       componentId,
		EnvTemplateID:     envTemplateId,
		DeploymentTrackID: deploymentTrackId,
		Configurations:    configs,
	}
	reqBodyBytes, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/configs/mappings", c.configMapBaseApiBaseUrl), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("error while creating config mapping, status code: %d", res.StatusCode)
}

func (c *ConfigMappingSvcClient) GetConfigMapping(
	orgId string,
	projectId string,
	componentId string,
	envTemplateId string,
	deploymentTrackId string,
) (*ConfigMappingResponse, error) {
	values := url.Values{}
	values.Add("projectId", projectId)
	values.Add("componentId", componentId)
	values.Add("envTemplateId", envTemplateId)
	values.Add("deploymentTrackId", deploymentTrackId)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/configs/mappings?%s", c.configMapBaseApiBaseUrl, values.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching config mapping: %w", err)
	}

	defer resp.Body.Close()

	var response ConfigMappingResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
