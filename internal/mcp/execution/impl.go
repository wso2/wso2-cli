package execution

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

func isExecutionSupported(displayType string) bool {
	switch displayType {
	case component.DisplayTypeManualTrigger, component.DisplayTypeScheduledTask, component.DisplayTypeBuildpackJob,
		component.DisplayTypeMiJob, component.DisplayTypeMiCronjob, component.DisplayTypeBuildpackCronJob,
		component.DisplayTypeByocCronjob, component.DisplayTypeByoiCronjob, component.DisplayTypeByocJob,
		component.DisplayTypeByoiJob:
		return true
	default:
		return false
	}
}

func getExecutions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target project", err), nil
	}

	selectedComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target component", err), nil
	}

	if !isExecutionSupported(selectedComponent.DisplayType) {
		return mcp.NewToolResultError(fmt.Sprintf(
			"Executions apply only to the 'Automation' subtype (a scheduled task). This integration's type is %q, which has no executions. "+
				"Request-driven integrations ('API', 'AI Agent', 'MCP Server') are invoked via their HTTP endpoint; event handlers "+
				"('Event Integration', 'File Integration') run when their trigger fires and cannot be run on demand.",
			selectedComponent.DisplayType)), nil
	}

	selectedEnvironment, err := utils.GetTargetEnv(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target environment", err), nil
	}

	latestVersion, err := utils.GetLatestApiVersion(*selectedComponent)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get latest api version", err), nil
	}

	componentClient := utils.GetComponentClient(ctx)
	componentDeployment, err := componentClient.GetComponentDeployment(
		targetOrg.Handle,
		targetOrg.UUID,
		targetOrg.ID,
		selectedComponent.Id,
		latestVersion.Id,
		selectedEnvironment.ID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get component deployment", err), nil
	}

	cloudDplanes, dplanes, err := common.GetDataPlaneInfo(targetOrg.ID, targetOrg.UUID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve executions."), nil
	}

	dpHost, isCilium := common.ResolveDataPlane(*selectedEnvironment, cloudDplanes, dplanes)

	logsClient := utils.GetLogsClient(ctx)
	executions, err := logsClient.GetExecutionsListV2(
		targetOrg.ID,
		dpHost,
		componentDeployment.ReleaseId,
		isCilium,
		10,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve executions."), nil
	}

	nextSteps := []string{"To view the logs for a specific execution, use `get_logs` with `log_type`: 'execution' and the `execution_id` from the list."}
	return utils.NewMCPResponse(executions, "Executions retrieved successfully.", nextSteps)
}

func executeTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target project", err), nil
	}

	selectedComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target component", err), nil
	}

	if !isExecutionSupported(selectedComponent.DisplayType) {
		return mcp.NewToolResultError(fmt.Sprintf(
			"Executions apply only to the 'Automation' subtype (a scheduled task). This integration's type is %q, which has no executions. "+
				"Request-driven integrations ('API', 'AI Agent', 'MCP Server') are invoked via their HTTP endpoint; event handlers "+
				"('Event Integration', 'File Integration') run when their trigger fires and cannot be run on demand.",
			selectedComponent.DisplayType)), nil
	}

	selectedEnvironment, err := utils.GetTargetEnv(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target environment", err), nil
	}

	latestVersion, err := utils.GetLatestApiVersion(*selectedComponent)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get latest api version", err), nil
	}

	componentClient := utils.GetComponentClient(ctx)
	componentDeployment, err := componentClient.GetComponentDeployment(
		targetOrg.Handle,
		targetOrg.UUID,
		targetOrg.ID,
		selectedComponent.Id,
		latestVersion.Id,
		selectedEnvironment.ID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get component deployment", err), nil
	}

	if componentDeployment == nil {
		return utils.NewMCPErrorResponse(fmt.Errorf("component is not deployed in the selected environment"), "Failed to execute task."), nil
	}

	runtimeArguments := utils.GetOptionalStringArgument(request, "runtime_arguments", "")
	runtimeArgumentsList := []string{}
	if runtimeArguments != "" {
		runtimeArgumentsList = strings.Split(runtimeArguments, ",")
	}

	execution, err := componentClient.TriggerExecution(
		targetOrg.ID,
		targetOrg.Handle,
		selectedProject.ID,
		selectedComponent.Id,
		componentDeployment.ReleaseId,
		runtimeArgumentsList,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to execute task."), nil
	}

	nextSteps := []string{fmt.Sprintf("Task execution has been triggered. To view the logs from this run, use `get_logs` with `log_type`: 'execution', `integration_uuid`: '%s', and `execution_id`: '%s'.", selectedComponent.Id, execution.RunId)}
	return utils.NewMCPResponse(execution, "Task execution triggered successfully.", nextSteps)
}
