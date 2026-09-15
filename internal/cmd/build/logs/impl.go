package logs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/build/buildstep"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	comp "github.com/wso2/integration-platform-tools/pkg/api/component"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func printBuildLog(opts *BuildLogFlags) (err error) {

	orgId, projectId, err := common.ResolveContext(opts.org, opts.project)

	if err == nil {
		if opts.org == "" {
			opts.org = orgId
		}

		if opts.project == "" {
			opts.project = projectId
		}
	}

	org, err := common.ResolveTargetOrganization(opts.org)

	if err != nil {
		return
	}

	project, err := common.ResolveTargetProject(org, opts.project)

	if err != nil {
		return
	}

	cmp, err := common.ResolveTargetComponent(org, project.ID, opts.component)

	if err != nil {
		return
	}

	if strings.HasPrefix(strings.ToLower(cmp.DisplayType), "byoi") {
		utils.PrintInfo("%s", i18n.T("Cannot get build logs for pre-built image based components\n"))
		return
	}

	dtrack, err := common.ResolveDeploymentTrack(cmp.DeploymentTracks, opts.deploymentTrack)

	if err != nil {
		return
	}

	selectedBuild, err := common.ResolveBuild(org, project.Handler, dtrack.Id, cmp.Name, opts.runId)

	if err != nil {
		return
	}

	// Argo build status contains cluster id and GitHub runner builds does not.
	isArgo := selectedBuild.Status.ClusterId != ""

	bStep, err := resolveBuildStep(org, project, cmp, selectedBuild, dtrack.Id, isArgo, opts.step)

	if err != nil {
		return err
	}

	if bStep == -1 {
		return fmt.Errorf("error resolving buildstep")
	}

	buildLogsSpinner := utils.CreateSpinner("Fetching logs...", "")
	buildLogsSpinner.Start()
	logsContent, err := GetBuildLogsContent(org, project, cmp, dtrack, selectedBuild, bStep, isArgo)
	buildLogsSpinner.Stop()

	if err != nil {
		return err
	}

	utils.PrintInfo("%s", logsContent)

	return nil
}

func resolveBuildStep(
	org *api.Organization,
	project *models.Project,
	component *models.Component,
	selectedBuild *deploymentbuild.BuildKind,
	deploymentTrackId string,
	isArgo bool,
	targetStep string) (buildstep.BuildStep, error) {

	if len(targetStep) > 0 {
		return buildstep.ResolveBuildStepFromShortForm(targetStep), nil
	}

	bsSpin := utils.CreateSpinner("Fetching build steps...", "")
	bsSpin.Start()
	availableSteps, err := GetAvailableBuildSteps(org, project, component, selectedBuild, deploymentTrackId, isArgo)
	bsSpin.Stop()

	if err != nil {
		return -1, err
	}

	var selectedStep string

	err = prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{
			Message: "Please select the build step to view logs",
			Values:  availableSteps,
		},
		&selectedStep,
	).Prompt()

	if err != nil {
		return -1, err
	}

	return buildstep.ResolveBuildStep(selectedStep), nil
}

func GetAvailableBuildSteps(
	org *api.Organization,
	project *models.Project,
	component *models.Component,
	selectedBuild *deploymentbuild.BuildKind,
	deploymentTrackId string,
	isArgo bool) ([]string, error) {

	var status *comp.DeploymentBuildStatusResponse
	var err error

	if isArgo {
		dataPlanes, cloudDataPlanes, err := auth.DevopsClient.GetAllDataPlaneClusters(org.ID, org.UUID)
		if err != nil {
			return nil, err
		}

		selectedDataPlaneHost, err := common.GetDataPlaneHost(cloudDataPlanes, dataPlanes, selectedBuild.Status.ClusterId)
		if err != nil {
			return nil, err
		}

		status, err = getArgoBuildLog(org.ID, selectedDataPlaneHost, *component, deploymentTrackId, selectedBuild.Status.BuildRef)
		if err != nil {
			return nil, err
		}
	} else {
		status, err = auth.ComponentClient.GetComponentBuildStatus(
			org.Handle,
			project.ID,
			component.Id,
			selectedBuild.Status.RunID,
			org.ID,
		)
	}

	if err != nil {
		return nil, err
	}

	availableSteps := make([]string, 0)

	for _, initStep := range status.Data.Init.Steps {
		if buildstep.HasLogs(buildstep.ResolveBuildStep(initStep.Name), initStep.Conclusion) {
			availableSteps = append(availableSteps, initStep.Name)
		}
	}

	for _, bStep := range status.Data.Build.Steps {
		if buildstep.HasLogs(buildstep.ResolveBuildStep(bStep.Name), bStep.Conclusion) {
			availableSteps = append(availableSteps, bStep.Name)
		}
	}

	for _, finalStep := range status.Data.Deploy.Steps {
		if buildstep.HasLogs(buildstep.ResolveBuildStep(finalStep.Name), finalStep.Conclusion) {
			availableSteps = append(availableSteps, finalStep.Name)
		}
	}

	return availableSteps, nil
}

func GetBuildLogs(org *api.Organization, project *models.Project, component *models.Component, selectedBuild *deploymentbuild.BuildKind, deploymentTrackId string, isArgo bool) (*comp.DeploymentBuildStatusResponse, error) {
	var status *comp.DeploymentBuildStatusResponse
	var err error

	if isArgo {
		dataPlanes, cloudDataPlanes, err := auth.DevopsClient.GetAllDataPlaneClusters(org.ID, org.UUID)
		if err != nil {
			return nil, err
		}

		selectedDataPlaneHost, err := common.GetDataPlaneHost(cloudDataPlanes, dataPlanes, selectedBuild.Status.ClusterId)
		if err != nil {
			return nil, err
		}

		status, err = getArgoBuildLog(org.ID, selectedDataPlaneHost, *component, deploymentTrackId, selectedBuild.Status.BuildRef)
		if err != nil {
			return nil, err
		}
	} else {
		status, err = auth.ComponentClient.GetComponentBuildStatus(
			org.Handle,
			project.ID,
			component.Id,
			selectedBuild.Status.RunID,
			org.ID,
		)
	}

	if err != nil {
		return nil, err
	}

	// Decode all base64 encoded logs
	if status.Data.Init.Log != "" {
		decodedLog, decodeErr := decodeAndPrintLog(status.Data.Init.Log)
		if decodeErr == nil {
			status.Data.Init.Log = decodedLog
		}
	}

	if status.Data.Build.Log != "" {
		decodedLog, decodeErr := decodeAndPrintLog(status.Data.Build.Log)
		if decodeErr == nil {
			status.Data.Build.Log = decodedLog
		}
	}

	if status.Data.OasValidation.Log != "" {
		decodedLog, decodeErr := decodeAndPrintLog(status.Data.OasValidation.Log)
		if decodeErr == nil {
			status.Data.OasValidation.Log = decodedLog
		}
	}

	if status.Data.GovernanceCheck.Log != "" {
		decodedLog, decodeErr := decodeAndPrintLog(status.Data.GovernanceCheck.Log)
		if decodeErr == nil {
			status.Data.GovernanceCheck.Log = decodedLog
		}
	}

	if status.Data.UpdateApi.Log != "" {
		decodedLog, decodeErr := decodeAndPrintLog(status.Data.UpdateApi.Log)
		if decodeErr == nil {
			status.Data.UpdateApi.Log = decodedLog
		}
	}

	return status, nil
}

func getArgoBuildLog(orgId string, selectedDataPlaneHost string, component models.Component, deploymentTrackId string, workflowName string) (*comp.DeploymentBuildStatusResponse, error) {
	reqBody := logs.GetBuildLogsReqBody{
		ComponentId:       component.Id,
		DeploymentTrackId: deploymentTrackId,
		WorkflowName:      workflowName,
	}

	return auth.LogsClient.GetBuildLogs(reqBody, orgId, selectedDataPlaneHost)
}

func decodeAndPrintLog(log string) (string, error) {
	decodedLog, err := base64.StdEncoding.DecodeString(log)
	if err != nil {
		return "", fmt.Errorf("failed to decode build log data: %v", err)
	}

	return string(decodedLog), nil
}

func GetBuildLogsContent(targetOrg *api.Organization, selectedProject *models.Project, selectedComponent *models.Component, dpTrack *models.DeploymentTrack, selectedBuild *deploymentbuild.BuildKind, selectedBuildStep buildstep.BuildStep, isArgo bool) (string, error) {
	switch selectedBuildStep {
	case buildstep.DOCKER_BUILD, buildstep.PACK_BUILD, buildstep.BALLERINA_BUILD:
		if isArgo {
			dataPlanes, cloudDataPlanes, err := auth.DevopsClient.GetAllDataPlaneClusters(targetOrg.ID, targetOrg.UUID)
			if err != nil {
				return "", fmt.Errorf("failed to get data plane cluster: %v", err)
			}

			selectedDataPlaneHost, err := common.GetDataPlaneHost(cloudDataPlanes, dataPlanes, selectedBuild.Status.ClusterId)
			if err != nil {
				return "", fmt.Errorf("failed to get data plane host: %v", err)
			}

			logs, err := auth.LogsClient.GetBuildLogs(logs.GetBuildLogsReqBody{
				ComponentId:       selectedComponent.Id,
				DeploymentTrackId: dpTrack.Id,
				WorkflowName:      selectedBuild.Status.BuildRef,
			}, targetOrg.ID, selectedDataPlaneHost)
			if err != nil {
				return "", fmt.Errorf("failed to get build logs: %v", err)
			}

			return decodeAndPrintLog(logs.Data.Build.Log)
		} else {
			buildLogsDataRes, err := auth.ComponentClient.GetComponentBuildStatus(targetOrg.Handle, selectedProject.ID, selectedComponent.Id, selectedBuild.Status.RunID, targetOrg.ID)
			if err != nil {
				return "", fmt.Errorf("failed to get build log data: %v", err)
			}

			return decodeAndPrintLog(buildLogsDataRes.Data.Build.Log)
		}

	case buildstep.SOURCE_CONFIGURATION_FILE_VALIDATION, buildstep.GIT_CHECKOUT:
		if isArgo {
			dataPlanes, cloudDataPlanes, err := auth.DevopsClient.GetAllDataPlaneClusters(targetOrg.ID, targetOrg.UUID)
			if err != nil {
				return "", fmt.Errorf("failed to get data plane cluster: %v", err)
			}

			selectedDataPlaneHost, err := common.GetDataPlaneHost(cloudDataPlanes, dataPlanes, selectedBuild.Status.ClusterId)
			if err != nil {
				return "", fmt.Errorf("failed to get data plane host: %v", err)
			}

			buildLogData, err := auth.LogsClient.GetBuildLogs(logs.GetBuildLogsReqBody{
				ComponentId:       selectedComponent.Id,
				DeploymentTrackId: dpTrack.Id,
				WorkflowName:      selectedBuild.Status.BuildRef,
			}, targetOrg.ID, selectedDataPlaneHost)
			if err != nil {
				return "", fmt.Errorf("failed to get build logs: %v", err)
			}

			return decodeAndPrintLog(buildLogData.Data.Init.Log)
		} else {
			buildLogData, err := auth.ComponentClient.GetComponentBuildStatus(targetOrg.Handle, selectedProject.ID, selectedComponent.Id, selectedBuild.Status.RunID, targetOrg.ID)
			if err != nil {
				return "", fmt.Errorf("failed to get build logs: %v", err)
			}

			return decodeAndPrintLog(buildLogData.Data.Init.Log)
		}
	case buildstep.TRIVY, buildstep.CHECKOV:
		builds, err := auth.ComponentClient.GetDeploymentStatusByVersion(selectedComponent.Id, dpTrack.Id, targetOrg.ID)
		if err != nil {
			return "", fmt.Errorf("failed to get deployment status version: %v", err)
		}

		var buildRef *component.ComponentDeploymentStatusForVersion

		for _, cb := range builds {
			if cb.Id == selectedBuild.Status.RunID {
				buildRef = &cb
				break
			}
		}

		if buildRef == nil {
			return "", fmt.Errorf("failed to get resolve build ref: %v", err)
		}

		result, err := auth.GQLBuildClient.GetScanReport(targetOrg.ID, selectedComponent.Id, buildRef.Sha)
		if err != nil {
			return "", fmt.Errorf("failed to get scan report: %v", err)
		}

		switch selectedBuildStep {
		case buildstep.TRIVY:
			return decodeAndPrintLog(result.TrivyScan)
		case buildstep.CHECKOV:
			decodedLog, err := base64.StdEncoding.DecodeString(result.CheckovScan)
			if err != nil {
				return "", fmt.Errorf("failed to decode build log data: %v", err)
			}

			var data CheckCovResponse
			data.Results.FailedChecks = make([]CheckCovFailedCheck, 0)

			if err := json.Unmarshal(decodedLog, &data); err != nil {
				return "", fmt.Errorf("failed to unmarshal build log data: %v", err)
			}

			tdata := make([][]string, 0)

			for _, checks := range data.Results.FailedChecks {
				tdata = append(tdata, []string{checks.CheckID, checks.CheckName})
			}

			table := utils.GenTable(
				tdata,
				[]string{"Checkov Code", "Description"},
				&utils.TableConfig{
					ColSeparator: "|",
					HeaderLine:   true,
					RowSeparator: "-",
				})
			return fmt.Sprintf("\nPassed : %d Failed : %d\n\n%s\n", data.Summary.Passed, data.Summary.Failed, table), nil
		}

	default:
		// fetch the logs from the buildlogs graphql Query
		logFetchKey := buildstep.ResolveLogFetchKey(selectedBuildStep)
		if logFetchKey == "unknown" {
			return "", fmt.Errorf("failed to resolve buildstep")
		}

		result, err := auth.GQLBuildClient.GetBuildLogsForStep(targetOrg.ID, selectedComponent.Id, selectedBuild.Status.RunID, logFetchKey)
		if err != nil {
			return "", fmt.Errorf("failed to get build logs")

		}

		switch selectedBuildStep {
		case buildstep.TRIVY_LIBRARY:
			return decodeAndPrintLog(result.LibraryTrivyReport)
		case buildstep.TRIVY_DROPINS:
			return decodeAndPrintLog(result.DropinsTrivyReport)
		case buildstep.INTEGRATION_PROJECT_BUILD:
			return decodeAndPrintLog(result.IntegrationProjectBuild)
		case buildstep.POST_BUILD_CHECK:
			return decodeAndPrintLog(result.PostBuildCheckLogs)
		case buildstep.MAIN_SEQUENCE_VALIDATION:
			return decodeAndPrintLog(result.MainSequenceValidation)
		case buildstep.MI_VERSION_VALIDATION:
			return decodeAndPrintLog(result.MIVersionValidation)
		case buildstep.OAS_VALIDATION:
			return decodeAndPrintLog(result.ProxyBuildLogs)
		case buildstep.API_GOVERNANCE:
			return decodeAndPrintLog(result.GovernanceLogs)
		case buildstep.READ_COMOPNENT_CONFIG:
			return decodeAndPrintLog(result.ConfigValidationLogs)
		}
	}
	return "", nil
}
