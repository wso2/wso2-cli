package project

import (
	"encoding/json"
	"strconv"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/clirpc/server"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

type GetProjectsByOrgIDRequest struct {
	OrgID string `json:"orgID"`
}

type GetProjectsByOrgIDResponse struct {
	Projects []models.Project `json:"projects"`
}

func GetProjectsByOrgID(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request GetProjectsByOrgIDRequest
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	projects, err := auth.ProjectClient.GetProjectsByOrgID(request.OrgID)

	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(GetProjectsByOrgIDResponse{
		Projects: projects,
	}), nil
}

func CreateProject(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId       string `json:"orgId"`
		OrgHandler  string `json:"orgHandler"`
		ProjectName string `json:"projectName"`
		Region      string `json:"region"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	orgIdInt, err := strconv.Atoi(request.OrgId)
	if err != nil {
		return server.Result{}, err
	}

	createProjectParam := project.GetProjectMutationRequest{
		Name:       request.ProjectName,
		Region:     request.Region,
		Version:    "1.0.0",
		OrgID:      int(orgIdInt),
		OrgHandler: request.OrgHandler,
	}

	createdProject, err := auth.ProjectClient.CreateProject(createProjectParam, request.OrgId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Project models.Project `json:"project"`
	}{
		Project: *createdProject,
	}), nil
}

func GetEnvs(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId     string `json:"orgId"`
		OrgUuid   string `json:"orgUuid"`
		ProjectId string `json:"projectId"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	envs, err := auth.ProjectClient.GetProjectEnvironments(request.OrgUuid, request.OrgId, request.ProjectId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Envs []project.ProjectEnvironment `json:"envs"`
	}{
		Envs: *envs,
	}), nil
}
