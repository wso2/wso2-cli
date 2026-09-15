package deployment

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	workflowmgt "github.com/wso2/integration-platform-tools/pkg/api/workflow-mgt"
)

func getDeploymentDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve deployment details."), nil
	}

	selectedComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve deployment details."), nil
	}

	selectedEnvironment, err := utils.GetTargetEnv(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve deployment details."), nil
	}

	latestVersion, err := utils.GetLatestApiVersion(*selectedComponent)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve deployment details."), nil
	}

	componentClient := utils.GetComponentClient(ctx)
	componentDeployment, err := componentClient.GetComponentDeployment(
		targetOrg.Handle,
		targetOrg.UUID,
		targetOrg.ID,
		selectedComponent.Id,
		latestVersion.Id,
		selectedEnvironment.ID)

	if componentDeployment == nil {
		return utils.NewMCPErrorResponse(fmt.Errorf("component deployment not found"), "Failed to retrieve deployment details."), nil
	}

	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve deployment info."), nil
	}

	deploymentInfo := map[string]any{
		"build":            componentDeployment.Build,
		"deploymentStatus": componentDeployment.DeploymentStatusV2,
		"Build":            componentDeployment.Build,
	}

	if componentDeployment.InvokeUrl != "" {
		deploymentInfo["invokeUrl"] = componentDeployment.InvokeUrl
	}

	projectClient := utils.GetProjectClient(ctx)
	endpoints, _ := projectClient.GetComponentEndpoints(selectedComponent.Id, latestVersion.Id, targetOrg.ID)
	if endpoints != nil {
		deploymentInfo["endpoints"] = *endpoints
	}

	nextSteps := []string{
		fmt.Sprintf("If the deployment is 'ACTIVE', you can invoke the service at the provided URL: '%s'.", componentDeployment.InvokeUrl),
		"To view the application logs for this deployment, use `get_logs` with `log_type`: 'application'.",
	}
	return utils.NewMCPResponse(deploymentInfo, "Deployment details retrieved successfully.", nextSteps)
}

func createDeployment(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	targetProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create deployment."), nil
	}

	targetComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *targetProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create deployment."), nil
	}

	dpTrack, err := utils.GetDeploymentTrack(*targetComponent, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create deployment."), nil
	}

	selectedEnvironment, err := utils.GetTargetEnv(ctx, *targetOrg, *targetProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create deployment."), nil
	}

	buildRefOrImageId, err := utils.GetRequiredStringArgument(request, "build_ref_or_image_id")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create deployment."), nil
	}

	if selectedEnvironment.Critical && len(selectedEnvironment.PromoteFrom) > 0 {
		workflowMgtClient := utils.GetWorkflowMgtClient(ctx)
		workflowMgtEnabled, err := workflowMgtClient.IsWorkflowEnabled(targetOrg.ID, workflowmgt.ENV_PROMOTION_WORKFLOW_ID)
		if err != nil {
			return utils.NewMCPErrorResponse(fmt.Errorf("failed to check workflow management status: %w", err), "Failed to create deployment."), nil
		}

		if workflowMgtEnabled {
			lowerEnv, err := utils.GetEnvironmentById(ctx, targetOrg.UUID, targetOrg.ID, targetProject.ID, selectedEnvironment.PromoteFrom[0])
			if err != nil {
				return mcp.NewToolResultErrorFromErr("failed to get lower environment", err), nil
			}

			workflowStatus, err := workflowMgtClient.GetEnvPromotionWfStatus(targetOrg.ID, buildRefOrImageId, selectedEnvironment.ID)
			if err != nil {
				return utils.NewMCPErrorResponse(fmt.Errorf("failed to get workflow management status: %w", err), "Failed to create deployment."), nil
			}

			switch workflowStatus.Status {
			case "NOT_FOUND", "ENABLED", "REJECTED", "CANCELLED", "TIMEOUT":
				buildId, err := utils.GetRequiredStringArgument(request, "build_run_id")
				if err != nil {
					return utils.NewMCPErrorResponse(err, "Failed to create deployment."), nil
				}

				requestComment := "Please approve the environment promotion request."
				wfInstanceId, err := workflowMgtClient.CreateEnvPromotionRequest(targetOrg.ID, targetOrg.Handle, targetProject.ID, targetProject.Name, requestComment, targetComponent.Name, buildId, buildRefOrImageId, selectedEnvironment.ID, selectedEnvironment.Name, lowerEnv.ID, lowerEnv.Name)
				if err != nil {
					return utils.NewMCPErrorResponse(fmt.Errorf("failed to create env promotion request: %w", err), "Failed to create deployment."), nil
				}

				wfRequestUrl := fmt.Sprintf("%s/organizations/%s/approvals?wfInstanceId=%s", utils.GetChoreConsoleBaseUrl(), targetOrg.Handle, wfInstanceId)
				return utils.NewMCPErrorResponse(fmt.Errorf(
					"your organization requires approvals to create deployments for %s environment. Deployment creation approval request was submitted successfully. A reviewer with workflow management permissions should approve the request through following link: %s. Try again to create the deployment after the request is approved",
					selectedEnvironment.Name, wfRequestUrl), "Failed to create deployment."), nil

			case "PENDING":
				wfRequestUrl := fmt.Sprintf("%s/organizations/%s/approvals?wfInstanceId=%s", utils.GetChoreConsoleBaseUrl(), targetOrg.Handle, workflowStatus.WorkflowInstanceId)
				return utils.NewMCPErrorResponse(fmt.Errorf(
					"your organization requires approvals to create deployments for %s environment. Deployment creation approval request has been submitted previously. A reviewer with workflow management permissions should approve the request through following link: %s. Try again to create the deployment after the request is approved",
					selectedEnvironment.Name, wfRequestUrl), "Failed to create deployment."), nil
			}
		}
	}

	if len(selectedEnvironment.PromoteFrom) > 0 {
		lowerEnv, err := utils.GetEnvironmentById(ctx, targetOrg.UUID, targetOrg.ID, targetProject.ID, selectedEnvironment.PromoteFrom[0])
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to get lower environment", err), nil
		}

		latestVersion, err := utils.GetLatestApiVersion(*targetComponent)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to get latest api version", err), nil
		}
		componentClient := utils.GetComponentClient(ctx)
		componentDeployment, err := componentClient.GetComponentDeployment(
			targetOrg.Handle,
			targetOrg.UUID,
			targetOrg.ID,
			targetComponent.Id,
			latestVersion.Id,
			selectedEnvironment.ID)
		if componentDeployment == nil {
			// No deployment in the upper environment. Fetch lower environment configuration if exists.
			configMappingApiClient := utils.GetConfigMappingApiClient(ctx)
			lowerEnvConfigs, err := configMappingApiClient.GetConfigMapping(targetOrg.ID, targetProject.ID,
				targetComponent.Id, lowerEnv.TemplateId, dpTrack.Id)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("failed to get config mappings", err), nil
			}
			upperEnvConfigs, err := configMappingApiClient.GetConfigMapping(targetOrg.ID, targetProject.ID,
				targetComponent.Id, selectedEnvironment.TemplateId, dpTrack.Id)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("failed to get config mappings", err), nil
			}
			if len(lowerEnvConfigs.Configurations) > 0 && len(upperEnvConfigs.Configurations) == 0 {
				return mcp.NewToolResultError("Create configurations for environment '" + selectedEnvironment.Name +
					"' before creating a deployment. No configurations found in the environment '" + selectedEnvironment.Name +
					"'." + "Resuse the" + lowerEnv.Name + "configuration or create a new configuration."), nil
			}

		}
	}

	deployOpts := deploymentbuild.DeploySpecOpts{}

	componentType := component.GetTypeForDisplayType(targetComponent.DisplayType)
	if componentType == "scheduled-task" {
		cronExpression, err := utils.GetRequiredStringArgument(request, "cron_expression")
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to create deployment."), nil
		}
		if cronExpression != "" {
			deployOpts.ScheduleExp = cronExpression
		}
		cronTimezone, err := utils.GetRequiredStringArgument(request, "cron_timezone")
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to create deployment."), nil
		}
		if cronTimezone != "" {
			deployOpts.ScheduleTZ = cronTimezone
		}
	}

	deploymentBuildClient := utils.GetDeploymentBuildClient(ctx)
	deployment, err := deploymentBuildClient.CreateDeployment(
		targetOrg.ID,
		targetComponent.Name,
		targetProject.Handler,
		dpTrack.Id, // TODO: for API-Proxy, this is the component?.apiVersions?.find((item) => item.latest)?.versionId
		selectedEnvironment.Name,
		buildRefOrImageId,
		deployOpts)

	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create deployment."), nil
	}

	nextSteps := []string{
		fmt.Sprintf("Deployment initiated for component '%s' in environment '%s'. To check the deployment status, use `get_deployment` with `integration_uuid`: '%s' and `environment_uuid`: '%s'.", targetComponent.Id, selectedEnvironment.ID, targetComponent.Id, selectedEnvironment.ID),
		"Make sure deployment gets active within 5 minutes by checking the deployment status using `get_deployment`.",
		"To view the application logs for this deployment, use `get_logs` with `log_type`: 'application'.",
	}
	return utils.NewMCPResponse(deployment, "Deployment initiated successfully.", nextSteps)
}
