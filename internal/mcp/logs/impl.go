package logs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	buildLogs "github.com/wso2/integration-platform-tools/internal/cmd/build/logs"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
)

func getLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logType, err := utils.GetRequiredStringArgument(request, "log_type")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve logs."), nil
	}

	switch logType {
	case "application", "gateway":
		return getRuntimeLogs(logType, ctx, request)
	case "execution":
		return getExecutionLogs(ctx, request)
	case "build":
		return getBuildLogs(ctx, request)
	default:
		return utils.NewMCPErrorResponse(fmt.Errorf("invalid log_type. Possible values: application, gateway, execution, build"), "Failed to retrieve logs."), nil
	}
}

func getRuntimeLogs(logType string, ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedEnv, err := utils.GetTargetEnv(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target environment", err), nil
	}

	if selectedComponent.DisplayType == component.DisplayTypeProxy {
		return utils.NewMCPErrorResponse(fmt.Errorf("application logs is not applicable for proxy type components"), "Failed to retrieve runtime logs."), nil
	}

	dpTrack, err := utils.GetDeploymentTrack(*selectedComponent, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get deployment track", err), nil
	}

	cloudDataPlanes, dataPlanes, err := common.GetDataPlaneInfo(targetOrg.ID, targetOrg.UUID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get data plane information", err), nil
	}

	selectedDataPlaneHost, isCilium := common.ResolveDataPlane(*selectedEnv, cloudDataPlanes, dataPlanes)

	logsClient := utils.GetLogsClient(ctx)
	logs, err := logsClient.GetComponentLogs(
		logs.GetComponentLogsReqBody{
			ComponentID:  selectedComponent.Id,
			EndTime:      time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
			Limit:        100,
			SearchPhrase: "",
			Sort:         "asc",
			SortingOrder: "asc",
			// Get logs for the last 10 minutes by default. ToDo: Make this configurable.
			StartTime:     time.Now().UTC().Add(-10 * time.Minute).Format("2006-01-02T15:04:05.999Z"),
			EnvironmentID: selectedEnv.ID,
			VersionList:   []string{dpTrack.ApiVersion},
			VersionIDList: []string{dpTrack.Id},
		},
		targetOrg.ID,
		selectedDataPlaneHost,
		isCilium,
		selectedEnv.Name,
		logType,
		false,
	)

	if err != nil {
		return utils.NewMCPErrorResponse(err, fmt.Sprintf("Failed to get %s logs.", logType)), nil
	}

	nextSteps := []string{"The logs have been retrieved. You can now analyze them for debugging or information."}
	return utils.NewMCPResponse(logs, "Logs retrieved successfully.", nextSteps)
}

func getExecutionLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	switch selectedComponent.DisplayType {
	case component.DisplayTypeManualTrigger, component.DisplayTypeScheduledTask, component.DisplayTypeBuildpackJob,
		component.DisplayTypeMiJob, component.DisplayTypeMiCronjob, component.DisplayTypeBuildpackCronJob,
		component.DisplayTypeByocCronjob, component.DisplayTypeByoiCronjob, component.DisplayTypeByocJob,
		component.DisplayTypeByoiJob:
	default:
		// NOTE: err is nil on this path — the lookup above succeeded. Build a
		// real error rather than passing the nil one.
		return utils.NewMCPErrorResponse(fmt.Errorf(
			"execution logs apply only to the 'Automation' subtype (a scheduled task). This integration's type is %q. "+
				"For request-driven integrations ('API', 'AI Agent', 'MCP Server') and event handlers "+
				"('Event Integration', 'File Integration'), use log_type 'application' instead",
			selectedComponent.DisplayType), "Failed to get execution logs."), nil
	}

	selectedEnv, err := utils.GetTargetEnv(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target environment", err), nil
	}

	dpTrack, err := utils.GetDeploymentTrack(*selectedComponent, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get deployment track", err), nil
	}

	cloudDataPlanes, dataPlanes, err := common.GetDataPlaneInfo(targetOrg.ID, targetOrg.UUID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get data plane information", err), nil
	}

	selectedDataPlaneHost, isCilium := common.ResolveDataPlane(*selectedEnv, cloudDataPlanes, dataPlanes)

	latestVersion, err := utils.GetLatestApiVersion(*selectedComponent)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve execution logs."), nil
	}

	componentClient := utils.GetComponentClient(ctx)
	componentDeployment, err := componentClient.GetComponentDeployment(
		targetOrg.Handle,
		targetOrg.UUID,
		targetOrg.ID,
		selectedComponent.Id,
		latestVersion.Id,
		selectedEnv.ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve execution logs."), nil
	}

	executionId := ""
	if utils.GetArgument(request, "execution_id") != nil {
		executionId = utils.GetArgument(request, "execution_id").(string)
	} else {
		return utils.NewMCPErrorResponse(fmt.Errorf("execution_id is required"), "Failed to retrieve execution logs."), nil
	}

	logsClient := utils.GetLogsClient(ctx)
	executionAttempts, err := logsClient.GetExecutionAttempts(
		targetOrg.ID,
		selectedDataPlaneHost,
		executionId,
		componentDeployment.ReleaseId,
		isCilium,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve execution logs."), nil
	}

	type AttemptLogs struct {
		Data logs.ExecutionAttempt   `json:"data"`
		Logs []logs.ComponentLogItem `json:"logs"`
	}

	attemptLogs := []AttemptLogs{}

	for _, attempt := range executionAttempts {
		logs, err := logsClient.GetExecutionLogs(
			targetOrg.ID,
			selectedDataPlaneHost,
			isCilium,
			selectedComponent.Id,
			dpTrack.Id,
			attempt.ID,
			selectedEnv.ID,
			selectedEnv.Name,
		)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to retrieve execution logs."), nil
		}
		attemptLogs = append(attemptLogs, AttemptLogs{
			Data: attempt,
			Logs: logs,
		})
	}

	nextSteps := []string{"The logs have been retrieved. You can now analyze them for debugging or information."}
	return utils.NewMCPResponse(attemptLogs, "Execution logs retrieved successfully.", nextSteps)
}

func getBuildLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	selectedComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	dpTrack, err := utils.GetDeploymentTrack(*selectedComponent, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get deployment track", err), nil
	}

	runId, err := utils.GetRequiredStringArgument(request, "build_status_run_id")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve build logs."), nil
	}

	runIdInt, err := strconv.Atoi(runId)
	if err != nil {
		return utils.NewMCPErrorResponse(fmt.Errorf("build_status_run_id must be an integer"), "Failed to retrieve build logs."), nil
	}

	var selectedBuild *deploymentbuild.BuildKind
	deploymentBuildClient := utils.GetDeploymentBuildClient(ctx)
	deploymentBuildRes, err := deploymentBuildClient.GetDeploymentBuilds(targetOrg.ID, selectedComponent.Handler, selectedProject.Handler, dpTrack.Id)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve build logs."), nil
	}
	for i := range deploymentBuildRes {
		item := &deploymentBuildRes[i]
		if item.Status != nil && item.Status.RunID == runIdInt {
			selectedBuild = item
			break
		}
	}
	if selectedBuild == nil {
		return utils.NewMCPErrorResponse(fmt.Errorf("failed to find build for the given build_status_run_id"), "Failed to retrieve build logs."), nil
	}

	isArgo := selectedBuild.Status.ClusterId != ""

	buildLogs, err := buildLogs.GetBuildLogs(targetOrg, selectedProject, selectedComponent, selectedBuild, dpTrack.Id, isArgo)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve build logs."), nil
	}

	nextSteps := []string{"The logs have been retrieved. You can now analyze them for debugging or information."}
	return utils.NewMCPResponse(buildLogs, "Build logs retrieved successfully.", nextSteps)
}
