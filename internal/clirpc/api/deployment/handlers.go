package deployment

import (
	"encoding/json"
	"fmt"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/clirpc/server"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	workflowmgt "github.com/wso2/integration-platform-tools/pkg/api/workflow-mgt"
)

func CreateDeployment(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId                string `json:"orgId"`
		OrgHandler           string `json:"orgHandler"`
		ComponentName        string `json:"componentName"`
		ComponentId          string `json:"componentId"`
		ComponentHandle      string `json:"componentHandle"`
		ComponentDisplayType string `json:"componentDisplayType"`
		ProjectHandle        string `json:"projectHandle"`
		ProjectId            string `json:"projectId"`
		VersionId            string `json:"versionId"`
		CommitHash           string `json:"commitHash"`
		EnvName              string `json:"envName"`
		EnvId                string `json:"envId"`
		BuildRef             string `json:"buildRef"`
		CronExpression       string `json:"cronExpression"`
		CronTimezone         string `json:"cronTimezone"`
		ProxyTargetUrl       string `json:"proxyTargetUrl"`
		ProxySandboxUrl      string `json:"proxySandboxUrl"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	deployOpts := deploymentbuild.DeploySpecOpts{}

	componentType := component.GetTypeForDisplayType(request.ComponentDisplayType)
	if componentType == "service" {
		remoteComWithRepoData, err := auth.ComponentClient.GetComponentInfo(
			request.OrgId,
			request.ComponentHandle,
			request.ProjectId)
		if err != nil {
			return server.Result{}, err
		}

		matchingAppEnv, err := common.GetReleaseEnvForDeploymentTrack(*remoteComWithRepoData, request.VersionId, request.EnvId)
		if err != nil {
			return server.Result{}, err
		}

		_, err = auth.ComponentClient.GenerateEndPoints(
			request.ComponentId,
			request.VersionId,
			matchingAppEnv.ReleaseId,
			request.CommitHash,
			request.OrgId,
		)

		if err != nil {
			return server.Result{}, err
		}
	} else if componentType == "scheduled-task" {
		if request.CronExpression != "" {
			deployOpts.ScheduleExp = request.CronExpression
		}
		if request.CronTimezone != "" {
			deployOpts.ScheduleTZ = request.CronTimezone
		}
	} else if request.ComponentDisplayType == component.DisplayTypeGitProxy {
		if request.ProxyTargetUrl != "" || request.ProxySandboxUrl != "" {
			deployOpts.ProxyConf = &deploymentbuild.ProxyDeploymentConfig{}
			if request.ProxyTargetUrl != "" {
				deployOpts.ProxyConf.Keys.ProductionEndpoint = request.ProxyTargetUrl
			}
			if request.ProxySandboxUrl != "" {
				deployOpts.ProxyConf.Keys.SandboxEndpoint = &request.ProxySandboxUrl
			}
		}
	}

	_, err := auth.DeploymentBuildClient.CreateDeployment(request.OrgId, request.ComponentName, request.ProjectHandle, request.VersionId, request.EnvName, request.BuildRef, deployOpts)
	if err != nil {
		return server.Result{}, err
	}

	return server.Result{}, nil
}

func GetProxyDeploymentInfo(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId       string `json:"orgId"`
		OrgHandler  string `json:"orgHandler"`
		OrgUuid     string `json:"orgUuid"`
		ComponentId string `json:"componentId"`
		VersionId   string `json:"versionId"`
		EnvId       string `json:"envId"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	deploymentInfo, err := auth.DeploymentBuildClient.GetProxyDeploymentInfo(
		request.OrgId,
		request.OrgHandler,
		request.OrgUuid,
		request.ComponentId,
		request.VersionId,
		request.EnvId,
	)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		DeploymentInfo deploymentbuild.ProxyDeployment `json:"deploymentInfo"`
	}{
		DeploymentInfo: *deploymentInfo,
	}), nil
}

func CheckWorkflowStatus(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId   string `json:"orgId"`
		BuildId string `json:"buildId"`
		EnvId   string `json:"envId"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	wfStatus, err := auth.WorkflowMgtClient.CheckWorkflowStatus(
		request.OrgId,
		request.BuildId,
		request.EnvId,
	)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Data workflowmgt.PromotionRequestStatus `json:"data"`
	}{
		Data: *wfStatus,
	}), nil
}

func PromoteProxy(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId            string `json:"orgId"`
		ComponentId      string `json:"componentId"`
		ApiId            string `json:"apiId"`
		PromoteFromEnvId string `json:"promoteFromEnvId"`
		EnvId            string `json:"envId"`
		BuildId          string `json:"buildId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	_, err := auth.DeploymentBuildClient.PromoteProxyComponent(
		request.OrgId,
		request.ComponentId,
		request.ApiId,
		request.PromoteFromEnvId,
		request.EnvId,
		request.BuildId,
	)
	if err != nil {
		return server.Result{}, err
	}

	return server.Result{}, nil
}

func RequestPromoteApproval(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId          string `json:"orgId"`
		OrgHandler     string `json:"orgHandler"`
		EnvId          string `json:"envId"`
		EnvName        string `json:"envName"`
		BuildId        string `json:"buildId"`
		ProjectId      string `json:"projectId"`
		ProjectName    string `json:"projectName"`
		RequestComment string `json:"requestComment"`
		ComponentName  string `json:"componentName"`
		EnvFromId      string `json:"envFromId"`
		EnvFromName    string `json:"envFromName"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	resource := fmt.Sprintf("%s_%s", request.BuildId, request.EnvId)

	err := auth.WorkflowMgtClient.RequestWorkflowApproval(
		request.OrgId,
		workflowmgt.RequestWorkflowApproval{
			ProjectId:      request.ProjectId,
			RequestComment: request.RequestComment,
			Context: workflowmgt.RequestWorkflowApprovalContext{
				WorkflowDefinitionIdentifier: "ENV_PROMOTION",
				Resource:                     resource,
			},
			Data: workflowmgt.RequestWorkflowApprovalData{
				Payload: workflowmgt.RequestWorkflowApprovalDataPayload{
					OrgName:       request.OrgHandler,
					ProjectName:   request.ProjectName,
					ComponentName: request.ComponentName,
					BuildId:       request.BuildId,
					EnvFrom: workflowmgt.RequestWorkflowApprovalEnv{
						Id:   request.EnvFromId,
						Name: request.EnvFromName,
					},
					EnvTo: workflowmgt.RequestWorkflowApprovalEnv{
						Id:   request.EnvId,
						Name: request.EnvName,
					},
				},
				Metadata: make(map[string]string),
			},
		},
	)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct{}{}), nil
}

func CancelApprovalRequest(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId         string `json:"orgId"`
		WkfInstanceId string `json:"wkfInstanceId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	err := auth.WorkflowMgtClient.CancelApprovalRequest(
		request.OrgId,
		request.WkfInstanceId,
	)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct{}{}), nil
}
