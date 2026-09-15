package devops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

type DevopsClient struct {
	devopsApiBaseUrl string
	client           *api.IPHTTPClient
}

func NewDevopsClient(devopsApiBaseUrl string, tokenStore api.ReadOnlyTokenStore) *DevopsClient {
	return &DevopsClient{
		devopsApiBaseUrl: devopsApiBaseUrl,
		client:           api.NewIPHTTPClient(tokenStore),
	}
}

func (c *DevopsClient) GetAllDataPlaneClusters(orgId string, orgUuid string) ([]DataPlaneItem, []GatewayInfo, error) {
	cloudDataPlanes, err := c.GetCloudPlaneClusters(orgUuid, orgId)
	if err != nil {
		return nil, nil, err
	}

	dataPlanes, err := c.GetDataPlaneClusters(orgId)
	if err != nil {
		return nil, nil, err
	}

	return dataPlanes, cloudDataPlanes, nil
}

func (c *DevopsClient) GetCloudPlaneClusters(orgUuid string, orgId string) ([]GatewayInfo, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+"/api/v1/clusters/clouddataplanes?org_uuid="+orgUuid, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching cloud data planes: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response []GatewayInfo

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *DevopsClient) GetDataPlaneClusters(orgId string) ([]DataPlaneItem, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+"/api/v1/clusters/dataplanes?org_id="+orgId, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching data planes: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response []DataPlaneItem

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *DevopsClient) GetBuildPackOptions(orgUuid string, orgId string, componentType string) ([]BuildPack, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/buildpacks?orgUuid=%s&componentType=%s&originCloud=devant", c.devopsApiBaseUrl, orgUuid, componentType),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching build pack options: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response []BuildPack

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *DevopsClient) GetConfigMapList(orgId string, orgUuid string, envId string, projectId string) ([]ConfigItem, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/environments/%s/configmap?organization_id=%s&project_id=%s", envId, orgUuid, projectId), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching config-map list: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data []ConfigItem `json:"data"`
	}

	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *DevopsClient) GetSecretsList(orgId string, orgUuid string, envId string, projectId string) ([]ConfigItem, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/environments/%s/secret?organization_id=%s&project_id=%s", envId, orgUuid, projectId), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching secret list: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data []ConfigItem `json:"data"`
	}

	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *DevopsClient) GetConfigMapDetails(orgId string, orgUuid string, envId string, configId string, projectId string) (ConfigItem, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/environments/%s/configmap/%s?organization_id=%s&project_id=%s", envId, configId, orgUuid, projectId), nil)
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while fetching config-map data: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data ConfigItem `json:"data"`
	}

	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return ConfigItem{}, err
	}

	return response.Data, nil
}

func (c *DevopsClient) GetSecretDetails(orgId string, orgUuid string, envId string, configId string, projectId string) (ConfigItem, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/environments/%s/secret/%s?organization_id=%s&project_id=%s", envId, configId, orgUuid, projectId), nil)
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while fetching config secrets: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data ConfigItem `json:"data"`
	}

	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return ConfigItem{}, err
	}

	return response.Data, nil
}

func (c *DevopsClient) GetReleaseContainers(orgId string, orgUuid string, componentId string, releaseId string, projectId string) ([]ConfigReleaseContainer, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/components/%s/release/%s?organization_id=%s&project_id=%s", componentId, releaseId, orgUuid, projectId), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching release containers: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data struct {
			Containers []ConfigReleaseContainer `json:"containers"`
		} `json:"data"`
	}

	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data.Containers, nil
}

func (c *DevopsClient) GetConfigMounts(orgId string, orgUuid string, componentId string, releaseId string, containerId string, projectId string) ([]ConfigMountData, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/components/%s/release/%s/container/%s/config-mount?organization_id=%s&project_id=%s", componentId, releaseId, containerId, orgUuid, projectId), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching config mounts: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data []ConfigMountData `json:"data"`
	}

	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *DevopsClient) CreateConfigMap(
	orgId string,
	orgUuid string,
	projectId string,
	envId string,
	configType string,
	name string,
	appEnvId string,
	data interface{},
) (ConfigItem, error) {
	type NewConfigRequest struct {
		ConfigType       string      `json:"config_type"`
		Data             interface{} `json:"data"`
		EnvironmentID    string      `json:"environment_id"`
		Metadata         struct{}    `json:"metadata"`
		Name             string      `json:"name"`
		OrganizationID   string      `json:"organization_id"`
		ProjectID        string      `json:"project_id"`
		AppEnvironmentID string      `json:"app_environment_id"`
		IsBase64         bool        `json:"isBase64"`
	}

	reqBody := NewConfigRequest{
		IsBase64:         false,
		OrganizationID:   orgUuid,
		ProjectID:        projectId,
		ConfigType:       configType,
		EnvironmentID:    envId,
		Metadata:         struct{}{},
		Name:             name,
		Data:             data,
		AppEnvironmentID: appEnvId,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/environments/%s/configmap?organization_id=%s&project_id=%s", envId, orgUuid, projectId), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data ConfigItem `json:"data"`
	}

	var response Response

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return ConfigItem{}, err
	}

	return response.Data, nil
}

func (c *DevopsClient) CreateConfigMount(
	orgId string,
	orgUuid string,
	projectId string,
	componentId string,
	containerId string,
	appEnvId string,
	configMapId *string,
	mountPath string,
	mountPermission string,
	mountType string,
	configKey string,
	secretId *string,

) error {
	type ConfigMountData struct {
		AppEnvironmentID string  `json:"app_environment_id"`
		ConfigKey        string  `json:"config_key"`
		ConfigMapID      *string `json:"configmap_id"`
		ContainerID      string  `json:"container_id"`
		DeployChanges    bool    `json:"deploy_changes"`
		MountPath        string  `json:"mount_path"`
		MountPermissions string  `json:"mount_permissions"`
		MountType        string  `json:"mount_type"`
		SecretID         *string `json:"secret_id"`
	}

	reqBody := ConfigMountData{
		AppEnvironmentID: appEnvId,
		DeployChanges:    true,
		ConfigKey:        configKey,
		ConfigMapID:      configMapId,
		SecretID:         secretId,
		ContainerID:      containerId,
		MountPath:        mountPath,
		MountPermissions: mountPermission,
		MountType:        mountType,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/components/%s/release/%s/container/%s/config-mount?organization_id=%s&project_id=%s", componentId, appEnvId, containerId, orgUuid, projectId), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	return nil
}

func (c *DevopsClient) CreateConfigSecret(
	orgId string,
	orgUuid string,
	projectId string,
	envId string,
	configType string,
	name string,
	appEnvId string,
	data interface{},
) (ConfigItem, error) {
	type NewConfigRequest struct {
		ConfigType       string      `json:"config_type"`
		Data             interface{} `json:"data"`
		EnvironmentID    string      `json:"environment_id"`
		Metadata         struct{}    `json:"metadata"`
		Name             string      `json:"name"`
		OrganizationID   string      `json:"organization_id"`
		ProjectID        string      `json:"project_id"`
		AppEnvironmentID string      `json:"app_environment_id"`
		IsBase64         bool        `json:"isBase64"`
		SaveType         string      `json:"save_type"`
		SecretType       string      `json:"secret_type"`
	}

	reqBody := NewConfigRequest{
		IsBase64:         false,
		OrganizationID:   orgUuid,
		ProjectID:        projectId,
		ConfigType:       configType,
		EnvironmentID:    envId,
		Metadata:         struct{}{},
		Name:             name,
		Data:             data,
		AppEnvironmentID: appEnvId,
		SaveType:         "Save",
		SecretType:       "Opaque",
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/environments/%s/secret?organization_id=%s&project_id=%s", envId, orgUuid, projectId), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return ConfigItem{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data ConfigItem `json:"data"`
	}

	var response Response

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return ConfigItem{}, err
	}

	return response.Data, nil
}

func (c *DevopsClient) DeleteConfigMount(orgId string, orgUuid string, componentId string, releaseId string, containerId string, configMountId string, projectId string) error {
	req, err := http.NewRequest("DELETE", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/components/%s/release/%s/container/%s/config-mount/%s?organization_id=%s&project_id=%s", componentId, releaseId, containerId, configMountId, orgUuid, projectId), nil)

	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while fetching config mounts: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error while reading response: %w", err)
	}

	return nil
}

func (c *DevopsClient) GetContainerRegistries(orgId string, orgUuid string) ([]ContainerRegistry, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+"/api/v1/container-registries?organization_id="+orgUuid, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching container registry: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data []ContainerRegistry `json:"data"`
	}
	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *DevopsClient) GetImageHistory(orgId string, orgUuid string, projectId string, componentId string, envVersionId string) ([]ImageHistoryResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/byoi/components/%s/versions/%s/images?organization_id=%s&project_id=%s", c.devopsApiBaseUrl, componentId, envVersionId, orgUuid, projectId), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching container history: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data []ImageHistoryResponse `json:"data"`
	}
	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *DevopsClient) GetSamples(orgId string, orgUuid string, projectId string) ([]ContainerImage, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/byoi/components/choreo-sample-images?organization_id=%s&project_id=%s", c.devopsApiBaseUrl, orgUuid, projectId), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching container registry: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data struct {
			Images []ContainerImage `json:"images"`
		} `json:"data"`
	}
	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data.Images, nil
}

func (c *DevopsClient) RegisterNewContainerRegistry(
	orgId string,
	orgUuid string,
) error {
	type NewConfigRequest struct {
		Name       string `json:"name"`
		Type       string `json:"type"`
		Provider   string `json:"provider"`
		Credential struct {
			Host string `json:"host"`
		} `json:"credential"`
	}

	reqBody := NewConfigRequest{
		Name:     "Samples Registry",
		Type:     "vendor-specific",
		Provider: "Azure",
		Credential: struct {
			Host string "json:\"host\""
		}{Host: "choreoanonymouspullable.azurecr.io"},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.devopsApiBaseUrl+"/api/v1/container-registries?organization_id="+orgUuid, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	return nil
}

func (c *DevopsClient) CreateByoiEndpoints(
	orgId string,
	orgUuid string,
	componentId string,
	releaseId string,
	projectId string,
	reqBody CreateByoiEndpointsRequest,
) error {
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/byoi/components/%s/releases/%s/endpoints?organization_id=%s&project_id=%s", c.devopsApiBaseUrl, componentId, releaseId, orgUuid, projectId), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	return nil
}

func (c *DevopsClient) GetDeploymentHistory(orgId, orgUuid, componentId, releaseId, projectId string) (history []DeploymentHistoryEntry, err error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/api/v1/components/%s/release/%s/deploy-history?organization_id=%s&project_id=%s",
			c.devopsApiBaseUrl,
			componentId,
			releaseId,
			orgUuid,
			projectId,
		),
		nil,
	)

	if err != nil {
		err = fmt.Errorf("error while creating request: %w", err)
		return
	}

	resp, err := c.client.Do(req, orgId)
	defer resp.Body.Close()

	if err != nil {
		err = fmt.Errorf("error while fetching deployment history: %w", err)
		return
	}

	type Response struct {
		Data []DeploymentHistoryEntry `json:"data"`
	}

	var result Response

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	history = result.Data

	return
}

func (c *DevopsClient) GetDeploymentPipeline(orgUuid string, orgId string) ([]DeploymentPipelineResponse, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/organizations/%s/deployment-pipelines", c.devopsApiBaseUrl, orgUuid),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching deployment pipeline options: %w", err)
	}
	defer resp.Body.Close()

	type Response struct {
		Data []DeploymentPipelineResponse `json:"data"`
	}

	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *DevopsClient) InitOrgRegion(
	orgId string,
	orgUuid string,
) error {
	type Request struct {
		Region string `json:"region"`
	}
	reqBody := Request{
		Region: region.GetCurrentRegion(),
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/organizations/%s/projects/init", orgUuid), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	return nil
}

func (c *DevopsClient) GetEnvironmentTemplates(orgId string) (*[]EnvironmentTemplate, error) {
	req, err := http.NewRequest("GET", c.devopsApiBaseUrl+fmt.Sprintf("/api/v1/organizations/%s/environment-templates", orgId), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching environment templates: %w", err)
	}
	defer resp.Body.Close()

	var response GetEnvironmentTemplatesResp

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding environment templates response: %w", err)
	}

	return &response.Data, nil
}
