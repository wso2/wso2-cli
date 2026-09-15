package gqlbuild

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type GQLBuildClient struct {
	gqlUrl     string
	httpClient *api.IPHTTPClient
}

func (c *GQLBuildClient) GetBuildList(orgId string, cmpId string, versionId string) ([]BuildListResponse, error) {
	q := GetBuildListQuery(orgId, cmpId, versionId)

	req, err := http.NewRequest("POST", c.gqlUrl, bytes.NewBuffer([]byte(q)))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.httpClient.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			Builds []BuildListResponse `json:"buildsByVersion"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data.Builds, nil
}

func (c *GQLBuildClient) GetDeploymentStatusByVersion(orgId, cmpId, versionId string) ([]DeploymentStatusByVersionResponse, error) {

	q := GetDeploymentStatusByVersionQuery(versionId, cmpId)

	req, err := http.NewRequest("POST", c.gqlUrl, bytes.NewBuffer([]byte(q)))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.httpClient.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			Builds []DeploymentStatusByVersionResponse `json:"buildsByVersion"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data.Builds, nil
}

func (c *GQLBuildClient) TriggerProxyBuild(orgId, cmpId, commitHash, apiId string) (*string, error) {
	q := GetCreateProxyBuild(cmpId, commitHash, apiId)

	req, err := http.NewRequest("POST", c.gqlUrl, bytes.NewBuffer([]byte(q)))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.httpClient.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			Response string `json:"triggerProxyBuild"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.Data.Response, nil
}

func (c *GQLBuildClient) GetBuildLogsForStep(
	orgId string,
	componentId string,
	runId int,
	logType string,
) (*ProjectBuildLogsData, error) {

	q := GetBuildStepLogQuery(componentId, runId, logType)

	req, err := http.NewRequest("POST", c.gqlUrl, bytes.NewBuffer([]byte(q)))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.httpClient.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			BuildLogs ProjectBuildLogsData `json:"buildLogs"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.Data.BuildLogs, nil
}

func (c *GQLBuildClient) GetScanReport(orgId string, componentId, branch string) (*DeployScanResult, error) {
	q := GetScanResultQuery(componentId, branch)

	req, err := http.NewRequest("POST", c.gqlUrl, bytes.NewBuffer([]byte(q)))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.httpClient.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Data struct {
			ScanResult DeployScanResult `json:"scanResult"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.Data.ScanResult, nil
}

func NewGQLBuildClient(gqlUrl string, tokenStore api.ReadOnlyTokenStore) *GQLBuildClient {
	return &GQLBuildClient{
		gqlUrl:     gqlUrl,
		httpClient: api.NewIPHTTPClient(tokenStore),
	}
}
