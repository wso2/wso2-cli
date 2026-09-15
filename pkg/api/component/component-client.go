package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

type ComponentClient struct {
	componentApiUrl    string
	componentManageUrl string
	configMangeUrl     string
	declarativeApiUrl  string
	client             *api.IPHTTPClient
}

func NewComponentClient(
	configManageUrl,
	componentApiUrl string,
	componentManageUrl string,
	declarativeApiUrl string,
	tokenStore api.ReadOnlyTokenStore) *ComponentClient {

	return &ComponentClient{
		configMangeUrl:     configManageUrl,
		componentApiUrl:    componentApiUrl,
		componentManageUrl: componentManageUrl,
		declarativeApiUrl:  declarativeApiUrl,
		client:             api.NewIPHTTPClient(tokenStore),
	}
}

func (c *ComponentClient) GetComponents(
	orgHandler string,
	orgId string,
	projectId string) ([]models.Component, error) {

	return c.GetAllComponents(orgHandler, orgId, projectId, false)
}

func (c *ComponentClient) GetAllComponents(
	orgHandler string,
	orgId string,
	projectId string,
	includeSystemComps bool) ([]models.Component, error) {

	query := GetProjectComponentsQuery(orgHandler, projectId)

	if includeSystemComps {
		query = GetProjectComponentsWithSystemQuery(orgHandler, projectId)
	}

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			Components []models.Component `json:"components"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	components := response.Data.Components
	return components, nil
}

func (c *ComponentClient) GetComponentsDeclarative(projectHandle string, orgId string) ([]ComponentKind, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/components",
		c.declarativeApiUrl,
		projectHandle,
	)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return []ComponentKind{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return []ComponentKind{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response []ComponentKind
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return []ComponentKind{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *ComponentClient) GetComponentDeclarative(
	projectHandle string,
	componentName string,
	orgId string,
) (ComponentKind, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/components/%s",
		c.declarativeApiUrl,
		projectHandle,
		componentName,
	)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return ComponentKind{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return ComponentKind{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response ComponentKind
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return ComponentKind{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *ComponentClient) GetComponentDeployment(
	orgHandler string,
	orgUuid string,
	orgId string,
	componentId string,
	versionId string,
	envId string,
) (*ComponentDeployment, error) {
	query := GetComponentDeploymentQuery(orgHandler, orgUuid, componentId, versionId, envId)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			ComponentDeployment ComponentDeployment `json:"componentDeployment"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	componentDeployment := response.Data.ComponentDeployment
	return &componentDeployment, nil
}

func (c *ComponentClient) CreateNewComponent(
	orgId string,
	projectHandle string,
	reqBody ComponentKind,
) (component ComponentKind, err error) {
	url := fmt.Sprintf(
		"%s/projects/%s/components",
		c.declarativeApiUrl,
		projectHandle,
	)

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return ComponentKind{}, fmt.Errorf("error while encoding request body: %w", err)
	}

	debugf("CreateNewComponent: POST %s (component=%q)", url, reqBody.Metadata.Name)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return ComponentKind{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		var errorResponse api.ForbiddenDeclarativeResponse
		if res != nil {
			rawErr, _ := io.ReadAll(res.Body)
			debugf("CreateNewComponent: HTTP %d error body: %s", res.StatusCode, string(rawErr))
			if err := json.Unmarshal(rawErr, &errorResponse); err != nil {
				return ComponentKind{}, err
			}
		}
		if strings.Contains(errorResponse.Message, "free tier") ||
			strings.Contains(errorResponse.Message, "maximum number of components reached") {
			return ComponentKind{}, api.ErrMaxComponentCountReached
		}
		return ComponentKind{}, err
	}
	defer res.Body.Close()

	rawBody, _ := io.ReadAll(res.Body)
	correlationID := res.Header.Get("X-Correlation-Id")
	debugf("CreateNewComponent: HTTP %d correlation-id=%s body-len=%d", res.StatusCode, correlationID, len(rawBody))

	// Surface any platform-level error even when HTTP 200.
	var errEnvelope struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	_ = json.Unmarshal(rawBody, &errEnvelope)
	if errEnvelope.Error != "" || (errEnvelope.Message != "" && !strings.Contains(errEnvelope.Message, reqBody.Metadata.Name)) {
		return ComponentKind{}, fmt.Errorf("platform error (HTTP %d): %s %s [body: %s]",
			res.StatusCode, errEnvelope.Error, errEnvelope.Message, string(rawBody))
	}

	var response ComponentKind
	if err := json.Unmarshal(rawBody, &response); err != nil {
		return ComponentKind{}, fmt.Errorf("error decoding response (HTTP %d): %w [body: %s]",
			res.StatusCode, err, string(rawBody))
	}

	response.CorrelationID = correlationID
	return response, nil
}

func (c *ComponentClient) GetComponentInfo(
	orgId string,
	componentHandle string,
	projectId string) (*models.Component, error) {

	query := GetComponentQuery(componentHandle, projectId)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return &models.Component{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return &models.Component{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			Component models.Component `json:"component"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return &models.Component{}, fmt.Errorf("error while decoding response: %w", err)
	}

	componentInfo := response.Data.Component
	return &componentInfo, nil
}

func (c *ComponentClient) GetComponentRepoInfo(
	orgId string,
	componentHandle string,
	projectId string) (*models.Component, error) {

	query := GetComponentRepoQuery(componentHandle, projectId)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return &models.Component{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return &models.Component{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			Component models.Component `json:"component"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return &models.Component{}, fmt.Errorf("error while decoding response: %w", err)
	}

	componentInfo := response.Data.Component
	return &componentInfo, nil
}

func (c *ComponentClient) GenerateEndPoints(
	componentId string,
	versionId string,
	releaseId string,
	commitHash string,
	orgId string,
) (endpoints []ComponentEndpointsData, err error) {
	query := GetGenerateEndpointsQuery(componentId, versionId, releaseId, commitHash)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			GenerateComponentEndpoints []ComponentEndpointsData `json:"generateComponentEndpoints"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	endpoints = response.Data.GenerateComponentEndpoints
	return
}

func (c *ComponentClient) GetAutoBuildStatus(
	componentId string, versionId string, orgId string,
) (AutoBuildStatusResponse, error) {
	query := GetAutoBuildTriggerStatusQuery(componentId, versionId)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return AutoBuildStatusResponse{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return AutoBuildStatusResponse{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			AutoBuildTrigger AutoBuildStatusResponse `json:"autoBuildTrigger"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return AutoBuildStatusResponse{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response.Data.AutoBuildTrigger, nil
}

func (c *ComponentClient) HandleEnableAutoBuild(
	componentId string, versionId string, envId string, orgId string,
) (AutoBuildToggleResponse, error) {
	query := GetEnableAutoBuildQuery(componentId, versionId, envId)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return AutoBuildToggleResponse{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return AutoBuildToggleResponse{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			HandleEnableAutoBuild AutoBuildToggleResponse `json:"handleEnableAutoBuild"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return AutoBuildToggleResponse{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response.Data.HandleEnableAutoBuild, nil
}

func (c *ComponentClient) HandleDisableAutoBuild(
	componentId string, versionId string, envId string, orgId string,
) (AutoBuildToggleResponse, error) {
	query := GetDisableAutoBuildQuery(componentId, versionId, envId)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return AutoBuildToggleResponse{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return AutoBuildToggleResponse{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			HandleDisableAutoBuild AutoBuildToggleResponse `json:"handleDisableAutoBuild"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return AutoBuildToggleResponse{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response.Data.HandleDisableAutoBuild, nil
}

func (c *ComponentClient) GetComponentInitStatus(
	orgId string,
	orgHandler string,
	projectId string,
	componentId string) (ComponentInitStatusResponse, error) {
	url := fmt.Sprintf("%s/orgs/%s/projects/%s/components/%s/init/status", c.componentManageUrl, orgHandler, projectId, componentId)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return ComponentInitStatusResponse{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return ComponentInitStatusResponse{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response ComponentInitStatusResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return ComponentInitStatusResponse{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *ComponentClient) GetComponentLimits(orgId string, orgUuid string) (ComponentLimits, error) {
	url := fmt.Sprintf("%s/orgs/%s/component-limits?originCloud=devant", c.componentManageUrl, orgUuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ComponentLimits{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return ComponentLimits{}, fmt.Errorf("error while fetching component limits: %w", err)
	}
	defer res.Body.Close()

	var response ComponentLimitsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return ComponentLimits{}, fmt.Errorf("error while decoding component limits response: %w", err)
	}

	return response.Data, nil
}

func (c *ComponentClient) MakeConfigurationsCall(
	orgId string,
	orgHandler string,
	projectId string,
	componentId string,
	envId string,
	versionId string,
	commitHash string,
	moduleName string) error {
	url := fmt.Sprintf(
		"%s/orgs/%s/projects/%s/components/%s/envs/%s/%s/configurations",
		c.configMangeUrl,
		orgHandler,
		projectId,
		componentId,
		envId,
		versionId,
	)

	query := fmt.Sprintf(`
	{
		"moduleName":"%s",
		"commitHash":"%s",
		"applyNow":false,
		"operation":0,
		"sourceUuid":"",
		"configs":[]
	}
	`,
		moduleName,
		commitHash,
	)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(query)))

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

func (c *ComponentClient) DeployComponent(
	orgId string,
	componentId *string,
	versionId *string,
	envId *string,
	branch *string,
	sha *string,
	shaDate *string) (ComponentDeploymentResponse, error) {

	query := GetComponentDeployQuery(*componentId, *versionId, *envId, *branch, *sha, *shaDate)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return ComponentDeploymentResponse{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return ComponentDeploymentResponse{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			DeployComponent ComponentDeploymentResponse `json:"deployComponent"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return ComponentDeploymentResponse{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response.Data.DeployComponent, nil
}

func (c *ComponentClient) GetDeploymentStatusByVersion(
	componentId string,
	versionId string,
	orgId string,
) ([]ComponentDeploymentStatusForVersion, error) {
	query := GetDeploymentStatusByVersionQuery(componentId, versionId)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			DeploymentStatusByVersion []ComponentDeploymentStatusForVersion `json:"deploymentStatusByVersion"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	return response.Data.DeploymentStatusByVersion, nil
}

func (c *ComponentClient) GetDeploymentTrackImages(
	componentId string,
	versionId string,
	orgId string,
) ([]ComponentDeploymentTrackImages, error) {
	query := GetDeploymentTrackImagesQuery(componentId, versionId)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			DeploymentTrackImages []ComponentDeploymentTrackImages `json:"deploymentTrackImages"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	return response.Data.DeploymentTrackImages, nil
}

func (c *ComponentClient) GetComponentBuildStatus(
	orgHandler string,
	projectId string,
	componentId string,
	runId int,
	orgId string) (*DeploymentBuildStatusResponse, error) {
	url := fmt.Sprintf(
		"%s/orgs/%s/projects/%s/components/%s/runs/%d/logs",
		c.componentManageUrl,
		orgHandler,
		projectId,
		componentId,
		runId,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response DeploymentBuildStatusResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	return &response, nil
}

func (c *ComponentClient) DeleteComponent(orgId, orgHandler, cmpId, projectId string) (resp ComponentDeleteResponse, err error) {
	query := GetDeleteComponentQuery(orgHandler, cmpId, projectId)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

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
			ComponentDeleteResponse ComponentDeleteResponse `json:"deleteComponentV2"`
		} `json:"data"`
	}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		err = fmt.Errorf("error while decoding response: %w", err)
		return
	}

	resp = response.Data.ComponentDeleteResponse

	return
}

func (c *ComponentClient) GetRepoDirStructure(
	gitOrg string,
	repoName string,
	branch string,
	isPublicRepo bool,
	orgID string,
) (resp []PathEntry, err error) {
	url := fmt.Sprintf(
		"%s/repositories/%s/%s/branches/%s/contents?isPublicRepo=%s",
		c.componentManageUrl,
		gitOrg,
		repoName,
		branch,
		strconv.FormatBool(isPublicRepo),
	)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		err = fmt.Errorf("error while creating request: %w", err)
		return
	}

	res, err := c.client.Do(req, orgID)

	if err != nil {
		err = fmt.Errorf("error while executing request: %w", err)
		return
	}

	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		err = fmt.Errorf("error while decoding response: %w", err)
		return
	}

	return
}

func (c *ComponentClient) GetRepoMetadata(orgId, gitOrgName, gitRepoName, branchName, subpath, secretRef string) (resp RepoMetadataResponse, err error) {
	query := GetRepoMetadataQuery(gitOrgName, gitRepoName, branchName, subpath, secretRef)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

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
			RepoMetadata RepoMetadataResponse `json:"repoMetadata"`
		} `json:"data"`
	}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		err = fmt.Errorf("error while decoding response: %w", err)
		return
	}

	resp = response.Data.RepoMetadata

	return
}

func (c *ComponentClient) GitTokenForRepository(gitOrg, gitRepo, orgId, secretRef string) (resp GitTokenForRepositoryResponse, err error) {
	query := GetGitTokenForRepositoryQuery(gitOrg, gitRepo, secretRef)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

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
			GitTokenForRepository GitTokenForRepositoryResponse `json:"gitTokenForRepository"`
		} `json:"data"`
	}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		err = fmt.Errorf("error while decoding response: %w", err)
		return
	}

	resp = response.Data.GitTokenForRepository

	return
}

func (c *ComponentClient) GetBuildLogsForType(orgId string, componentId string, runId int, logType string) (resp ProjectBuildLogsData, err error) {
	query := GetBuildLogsForType(componentId, runId, logType)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

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
			BuildLogs ProjectBuildLogsData `json:"buildLogs"`
		} `json:"data"`
	}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		err = fmt.Errorf("error while decoding response: %w", err)
		return
	}

	resp = response.Data.BuildLogs

	return
}

func (c *ComponentClient) TriggerExecution(
	orgId, orgHandle, projectId, componentId, releaseId string,
	args []string) (*RunPodResponse, error) {
	reqUrl := fmt.Sprintf(
		"%s/orgs/%s/projects/%s/components/%s/releases/%s/run-pod",
		c.componentManageUrl,
		orgHandle,
		projectId,
		componentId,
		releaseId,
	)

	var payload = struct {
		Args []string `json:"args"`
	}{Args: args}

	encPayload, err := json.Marshal(payload)

	if err != nil {
		return nil, fmt.Errorf("error while creating json payload: %w", err)
	}

	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(encPayload))

	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("error while making request: %w", err)
	}

	defer res.Body.Close()

	var response struct {
		Data RunPodResponse `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	return &response.Data, nil
}

func (c *ComponentClient) CreateByoiComponent(
	orgId string, name string, displayName string, description string, projectId string, componentType string, version string, imageUrl string, registryId string, isSystemComponent bool,
) (err error) {
	query := GetCreateByoiComponentQuery(name, displayName, description, projectId, componentType, imageUrl, registryId, version, "true")

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	return
}

func (c *ComponentClient) DeployByoiComponent(
	orgId string, componentId string, releaseId string, imageUrl string,
) (success bool, err error) {
	query := GetDeployImageQuery(componentId, releaseId, imageUrl)

	req, err := http.NewRequest("POST", c.componentApiUrl, bytes.NewBuffer([]byte(query)))

	if err != nil {
		return false, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return false, fmt.Errorf("error while executing request: %w", err)
	}

	type Response struct {
		Data struct {
			DeployImage struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			} `json:"deployImage"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return false, err
	}

	defer res.Body.Close()

	return response.Data.DeployImage.Success, nil
}
