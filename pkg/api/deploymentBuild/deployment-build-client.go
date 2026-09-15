package deploymentbuild

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

type DeploymentBuildClient struct {
	DeclarativeApiUrlRoot string
	DeployProxyEp         string
	GQLEp                 string
	Client                *api.IPHTTPClient
}

func NewDeploymentBuildClient(declarativeApiUrlRoot, deployProxyEp, gqlEp string, tokenStore api.ReadOnlyTokenStore) *DeploymentBuildClient {
	return &DeploymentBuildClient{
		DeclarativeApiUrlRoot: declarativeApiUrlRoot,
		DeployProxyEp:         deployProxyEp,
		GQLEp:                 gqlEp,
		Client:                api.NewIPHTTPClient(tokenStore),
	}
}

func (c *DeploymentBuildClient) GetDeploymentBuilds(
	orgId string,
	componentHandle string,
	projectHandle string,
	deploymentTrackId string,
) ([]BuildKind, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/components/%s/deploymentTracks/%s/builds",
		c.DeclarativeApiUrlRoot,
		projectHandle,
		componentHandle,
		deploymentTrackId,
	)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return []BuildKind{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.Client.Do(req, orgId)

	if err != nil {
		return []BuildKind{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response []BuildKind
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return []BuildKind{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *DeploymentBuildClient) CreateDeploymentBuilds(
	orgId string,
	componentName string,
	projectHandle string,
	deploymentTrackId string,
	commitHash string,
) (BuildKind, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/components/%s/deploymentTracks/%s/builds",
		c.DeclarativeApiUrlRoot,
		projectHandle,
		componentName,
		deploymentTrackId,
	)

	name := uuid.New().String()

	reqBody := BuildKind{
		ApiVersion: "core.choreo.dev/v1alpha1",
		Kind:       "Build",
		Metadata: BuildMetadata{
			Name:          name,
			ComponentName: componentName,
			ProjectName:   projectHandle,
		},
		Spec: BuildSpec{
			Revision: commitHash,
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return BuildKind{}, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return BuildKind{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.Client.Do(req, orgId)
	if err != nil {
		return BuildKind{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response BuildKind
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return BuildKind{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *DeploymentBuildClient) GetDeployments(
	orgId string,
	componentHandle string,
	projectHandle string,
	deploymentTrackId string,
) ([]DeployKind, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/components/%s/deploymentTracks/%s/deployments",
		c.DeclarativeApiUrlRoot,
		projectHandle,
		componentHandle,
		deploymentTrackId,
	)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return []DeployKind{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.Client.Do(req, orgId)

	if err != nil {
		return []DeployKind{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response []DeployKind
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return []DeployKind{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *DeploymentBuildClient) CreateDeployment(
	orgId string,
	componentName string,
	projectHandle string,
	deploymentTrackId string,
	envName string,
	buildRef string,
	deployOpts DeploySpecOpts,
) (DeploymentTrackResponse, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/components/%s/deploymentTracks/%s/deployments",
		c.DeclarativeApiUrlRoot,
		projectHandle,
		componentName,
		deploymentTrackId,
	)

	name := uuid.New().String()

	reqBody := DeployKind{
		ApiVersion: "core.choreo.dev/v1alpha1",
		Kind:       "Deployment",
		Metadata: DeployMetadata{
			Name:              name,
			ComponentName:     componentName,
			ProjectName:       projectHandle,
			Environment:       envName,
			DeploymentTrackId: deploymentTrackId,
		},
		Spec: DeploySpec{
			BuildRef:       buildRef,
			DeploySpecOpts: deployOpts,
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return DeploymentTrackResponse{}, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return DeploymentTrackResponse{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.Client.Do(req, orgId)
	if err != nil {
		return DeploymentTrackResponse{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response DeploymentTrackResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return DeploymentTrackResponse{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *DeploymentBuildClient) CreateProxyComponentDeployment(orgId, cmpId, apiV, buildId, envId string) error {
	// initiate deployment
	// url := fmt.Sprintf(
	// 	"%s/components/%s/versions/%s/initiate-deployment?environmentId=%s&forceBuild=false&accessMode=%s",
	// 	c.DeployProxyEp,
	// 	cmpId,
	// 	apiV,
	// 	envId,
	// 	"external", // | INTERNAL
	// )
	//
	// reqBody := struct{}{}
	//
	// reqBodyBytes, err := json.Marshal(reqBody)
	// if err != nil {
	// 	return fmt.Errorf("error while encoding request body: %w", err)
	// }
	//
	// req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	// if err != nil {
	// 	return fmt.Errorf("error while creating request: %w", err)
	// }
	//
	// res, err := c.Client.Do(req, orgId)
	// if err != nil {
	// 	return fmt.Errorf("error while executing request: %w", err)
	// }
	//
	// defer res.Body.Close()
	//
	// var response struct {
	// 	// "message":"Deployment initiated ", "success":true
	// 	Message string `json:"message"`
	// 	Success bool   `json:"success"`
	// }
	//
	// if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
	// 	return fmt.Errorf("error while decoding response: %w", err)
	// }

	// Deploy service Call
	url := fmt.Sprintf(
		"%s/components/%s/versions/%s/deploy-service?buildId=%s&environmentId=%s&accessMode=%s",
		c.DeployProxyEp,
		cmpId,
		apiV,
		buildId,
		envId,
		"external", // | INTERNAL
	)

	reqBody := struct{}{}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error while encoding request body: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	dRes, err := c.Client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while executing request: %w", err)
	}

	defer dRes.Body.Close()

	var response struct {
		// "message":"Deployment initiated ", "success":true
		Message string `json:"message"`
		Success bool   `json:"success"`
	}
	if err := json.NewDecoder(dRes.Body).Decode(&response); err != nil {
		return fmt.Errorf("error while decoding response: %w", err)
	}

	return nil
}

func (c *DeploymentBuildClient) GetProxyDeploymentInfo(
	orgId, orgHandler, orgUuid, cmpId, versionId, envId string) (info *ProxyDeployment, err error) {

	q := getProxyDeploymentQuery(orgHandler, orgUuid, cmpId, versionId, envId)

	req, err := http.NewRequest("POST", c.GQLEp, bytes.NewBuffer([]byte(q)))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.Client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			DeploymentInfo ProxyDeployment `json:"proxyDeployment"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.Data.DeploymentInfo, nil
}

func (c *DeploymentBuildClient) SuspendDeployment(
	orgId,
	orgHandler,
	componentId,
	releaseId,
	componentType string) (string, error) {

	q := GetSuspendDeploymentQuery(orgHandler, componentId, releaseId, componentType)
	req, err := http.NewRequest("POST", c.GQLEp, bytes.NewBuffer([]byte(q)))
	if err != nil {
		return "", fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.Client.Do(req, orgId)
	if err != nil {
		return "", fmt.Errorf("error while fetching component endpoints: %w", err)
	}
	defer resp.Body.Close()
	type Response struct {
		Data struct {
			StopDeployment string `json:"stopDeployment"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data.StopDeployment, nil
}

func (c *DeploymentBuildClient) GetProxyInsights(fFrom, fTo, orgId, envs, apiId string) (*ProxyInsightsResponse, error) {
	q := getComponentInsightsQuery(fFrom, fTo, orgId, envs, apiId)
	req, err := http.NewRequest("POST", c.GQLEp, bytes.NewBuffer([]byte(q)))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.Client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data ProxyInsightsResponse `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.Data, nil
}

func (c *DeploymentBuildClient) PromoteProxyComponent(
	orgId,
	cmpId,
	versionId,
	fromEnvId,
	targetEnvid,
	buildId string,
) (result *ProxyPromoteResponse, err error) {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/components/%s/versions/%s/promote?fromEnv=%s&targetEnv=%s&buildId=%s&accessMode=external",
			c.DeployProxyEp,
			cmpId,
			versionId,
			fromEnvId,
			targetEnvid,
			buildId,
		),
		nil,
	)

	if err != nil {
		return
	}

	resp, err := c.Client.Do(req, orgId)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return
}

func (c *DeploymentBuildClient) ResumeDeployment(
	orgId,
	orgHandler,
	componentId,
	releaseId,
	cmpType string) (string, error) {

	q := GetResumeDeploymentQuery(orgHandler, componentId, releaseId, cmpType)
	req, err := http.NewRequest("POST", c.GQLEp, bytes.NewBuffer([]byte(q)))
	if err != nil {
		return "", fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.Client.Do(req, orgId)
	if err != nil {
		return "", fmt.Errorf("error while fetching component endpoints: %w", err)
	}
	defer resp.Body.Close()
	type Response struct {
		Data struct {
			RedeployDeployment string `json:"redeployDeployment"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Data.RedeployDeployment, nil
}
