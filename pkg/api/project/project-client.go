package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

type ProjectClient struct {
	projectApiUrl     string
	billingConsoleUrl string
	insightQueryEp    string
	client            *api.IPHTTPClient
}

func NewProjectClient(projectApiUrl, billingConsoleUrl, insightQueryEp string, tokenStore api.ReadOnlyTokenStore) *ProjectClient {
	return &ProjectClient{
		projectApiUrl:     projectApiUrl,
		billingConsoleUrl: billingConsoleUrl,
		client:            api.NewIPHTTPClient(tokenStore),
		insightQueryEp:    insightQueryEp,
	}
}

// Get Projects by Org ID
func (c *ProjectClient) GetProjectsByOrgID(orgID string) (projects []models.Project, err error) {
	query := GetProjectsByOrgIDQuery(orgID)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgID)
	if err != nil {
		return nil, fmt.Errorf("error while fetching projects: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	type Response struct {
		Data struct {
			Projects []models.Project `json:"projects"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	projects = response.Data.Projects
	return projects, nil
}

func (c *ProjectClient) CreateProject(createProjectReq GetProjectMutationRequest, orgId string) (*models.Project, error) {
	mutation := ""
	if createProjectReq.Repository != "" {
		mutation = GetCreateMonoRepoProjectMutation(createProjectReq)
	} else {
		mutation = GetCreateMultiRepoProjectMutation(createProjectReq)
	}

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(mutation)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, strconv.Itoa(createProjectReq.OrgID))
	if err != nil {
		var errorResponse api.ForbiddenApiResponse
		if resp != nil {
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
				return nil, fmt.Errorf("error while decoding error response: %w", err)
			}
		}

		if errorResponse.Metadata.AdditionalData == "MAX_PROJECT_LIMIT_REACHED" {
			return nil, api.ErrMaxProjectCountReached
		}

		return nil, fmt.Errorf("error: %s", errorResponse.Message)
	}
	defer resp.Body.Close()
	type Response struct {
		Data struct {
			CreateProject models.Project `json:"createProject"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	project := response.Data.CreateProject
	return &project, nil
}

func (c *ProjectClient) UpdateProject(updateProjectReq GetProjectMutationRequest) (*models.Project, error) {
	mutation := GetUpdateProjectMutation(updateProjectReq)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(mutation)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, strconv.Itoa(updateProjectReq.OrgID))
	if err != nil {
		return nil, fmt.Errorf("error while updating project: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	type Response struct {
		Data struct {
			UpdateProject *models.Project `json:"updateProject"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	project := response.Data.UpdateProject
	return project, nil
}

func (c *ProjectClient) GetProjectEnvironments(orgUuid string, orgId string, projectId string) (*[]ProjectEnvironment, error) {
	query := GetProjectEnvironmentsQuery(orgUuid, projectId)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching project envs: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	type Response struct {
		Data struct {
			Environments *[]ProjectEnvironment `json:"environments"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	environments := response.Data.Environments
	return environments, nil
}

func (c *ProjectClient) GetComponentEndpoints(componentId string, versionId string, orgId string) (*[]Endpoint, error) {
	query := GetComponentEndpointsQuery(componentId, versionId)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching component endpoints: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	type Response struct {
		Data struct {
			ComponentEndpoints *[]Endpoint `json:"componentEndpoints"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	environments := response.Data.ComponentEndpoints
	return environments, nil
}

func (c *ProjectClient) GetProjectInfo(orgId string, projectId string) (*models.Project, error) {
	query := GetProjectInfoQuery(orgId, projectId)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("error while fetching project info: %w", err)
	}

	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data struct {
			Project models.Project `json:"project"`
		} `json:"data"`
	}

	var r Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	project := r.Data.Project

	return &project, nil
}

func (c *ProjectClient) GetProjectComponents(
	orgId string,
	orgHandle string,
	projectId string) (*[]models.Component, error) {

	query := GetProjectComponentsQuery(orgId, orgHandle, projectId)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)

	if err != nil {
		return nil, fmt.Errorf("error while fetching project components: %w", err)
	}

	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Response struct {
		Data struct {
			Project struct {
				Components []models.Component `json:"components"`
			} `json:"project"`
		} `json:"data"`
	}

	var r Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	components := r.Data.Project.Components

	return &components, nil
}

func (c *ProjectClient) DeleteProject(orgId string, projectId string) (resp DeleteProjectResponse, err error) {
	mutation := GetDeleteProjectQuery(orgId, projectId)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(mutation)))
	if err != nil {
		err = fmt.Errorf("error while creating request: %w", err)
		return
	}

	httpResp, err := c.client.Do(req, orgId)
	if err != nil {
		err = fmt.Errorf("error while deleting project: %w", err)
		return
	}

	defer httpResp.Body.Close()

	type Response struct {
		Data struct {
			DeleteProject DeleteProjectResponse `json:"deleteProject"`
		} `json:"data"`
	}

	var r Response
	if err = json.NewDecoder(httpResp.Body).Decode(&r); err != nil {
		return
	}

	resp = r.Data.DeleteProject

	return
}

func (c *ProjectClient) GetEnvList(orgId, orgUUID, projectId string) ([]EnvMetadata, error) {
	q := FetchEnvironmentsQuery(orgUUID, projectId)

	req, err := http.NewRequest("POST", c.insightQueryEp, bytes.NewBuffer([]byte(q)))
	if err != nil {
		err = fmt.Errorf("error while creating request: %w", err)
		return nil, err
	}

	httpResp, err := c.client.Do(req, orgId)
	if err != nil {
		err = fmt.Errorf("error while fetching environments: %w", err)
		return nil, err
	}

	defer httpResp.Body.Close()

	type Response struct {
		Data EnviroumentListResponse `json:"data"`
	}

	var r Response
	if err = json.NewDecoder(httpResp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return r.Data.ListEnvironments, nil
}
