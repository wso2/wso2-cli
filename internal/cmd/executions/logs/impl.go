package logs

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/huh"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/internal/utils/textstyle"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func handleCmd(params *ExecutionLogsOpts) (err error) {

	if params.ExecutionId == "" {
		utils.PrintError("%s", heredoc.Docf(`
                Execution Id argument is required.

                To obtain execution Id Use:
                %s 
                
            `,
			textstyle.CreateTextStyle().AddTextStyle(textstyle.BOLD).Text("wso2-integration-platform list executions")),
		)
		return fmt.Errorf("missing required argument")
	}

	orgId, projectId, err := common.ResolveContext(params.Org, params.Project)

	if err == nil {
		if params.Org == "" {
			params.Org = orgId
		}

		if params.Project == "" {
			params.Project = projectId
		}
	}

	sOrg, err := common.ResolveTargetOrganization(params.Org)
	if err != nil {
		return err
	}

	sProject, err := common.ResolveTargetProject(sOrg, params.Project)
	if err != nil {
		return err
	}

	sComponent, err := common.ResolveTargetComponent(sOrg, sProject.ID, params.Component)
	if err != nil {
		return err
	}

	switch sComponent.DisplayType {
	case component.DisplayTypeManualTrigger, component.DisplayTypeScheduledTask, component.DisplayTypeBuildpackJob,
		component.DisplayTypeMiJob, component.DisplayTypeMiCronjob, component.DisplayTypeBuildpackCronJob,
		component.DisplayTypeByocCronjob, component.DisplayTypeByoiCronjob, component.DisplayTypeByocJob,
		component.DisplayTypeByoiJob:
	default:
		return fmt.Errorf("executions apply only to Automation integrations (scheduled tasks)")
	}

	cInitStat, err := auth.ComponentClient.GetComponentInitStatus(sOrg.ID, sOrg.Handle, sProject.ID, sComponent.Id)
	if err != nil {
		return err
	}

	if cInitStat.Data.Status == "queued" || cInitStat.Data.Status == "in_progress" {
		utils.PrintInfo("%s", i18n.T("Component initialization is in progress. Please run the command in a while...\n"))
		return nil
	}

	dTrack, err := common.ResolveDeploymentTrack(sComponent.DeploymentTracks, params.DeploymentTrack)
	if err != nil {
		return err
	}

	sEnv, err := common.GetProjectEnv(sOrg.UUID, sOrg.ID, sProject.ID, params.Env)
	if err != nil {
		return err
	}

	var latestVersion models.ApiVersion
	for _, versionItem := range sComponent.ApiVersions {
		if versionItem.Latest {
			latestVersion = versionItem
			break
		}
	}

	deploymentStatusSpinner := utils.CreateSpinner(
		fmt.Sprintf(i18n.T(" Fetching deployment status for the %s environment"), sEnv.Name),
		"")
	deploymentStatusSpinner.Start()
	componentDeployment, err := auth.ComponentClient.GetComponentDeployment(
		sOrg.Handle,
		sOrg.UUID,
		sOrg.ID,
		sComponent.Id,
		latestVersion.Id,
		sEnv.ID)
	deploymentStatusSpinner.Stop()

	if err != nil {
		return err
	}

	cloudDplanes, dplanes, err := common.GetDataPlaneInfo(sOrg.ID, sOrg.UUID)
	if err != nil {
		return err
	}

	dpHost, isCilium := common.ResolveDataPlane(*sEnv, cloudDplanes, dplanes)

	execSpinner := utils.CreateSpinner("Fetching execution attempts...", "")
	execSpinner.Start()
	executionAttempts, err := auth.LogsClient.GetExecutionAttempts(
		sOrg.ID,
		dpHost,
		params.ExecutionId,
		componentDeployment.ReleaseId,
		isCilium,
	)
	execSpinner.Stop()
	if err != nil {
		return err
	}

	attempt, err := resolveAttempt(executionAttempts, params.AttemptNumber)

	if err != nil {
		return err
	}

	logSpinner := utils.CreateSpinner("Fetching logs...", "")
	logSpinner.Start()
	logs, err := auth.LogsClient.GetExecutionLogs(
		sOrg.ID,
		dpHost,
		isCilium,
		sComponent.Id,
		dTrack.Id,
		attempt.ID,
		sEnv.ID,
		sEnv.Name,
	)
	logSpinner.Stop()

	if err != nil {
		return err
	}

	printLogs(logs)

	return nil
}

func resolveAttempt(attempts []logs.ExecutionAttempt, attemptNumber int) (*logs.ExecutionAttempt, error) {
	if attemptNumber < 0 && attemptNumber > len(attempts) {
		return nil, fmt.Errorf("invalid attempt value")
	}

	if attemptNumber == 0 {
		var opts = make([]huh.Option[int], 0)

		for i, attempt := range attempts {
			attStTime, err := strconv.Atoi(attempt.StartTime)

			if err != nil {
				return nil, err
			}

			opts = append(opts, huh.NewOption(
				fmt.Sprintf("Attempt #%d - %s", i+1, utils.TimeAgo(time.Unix(int64(attStTime), 0))),
				i+1,
			))
		}

		err := prompt.NewPromptSelectMessage(
			prompt.PromptSelectOpts[int]{
				Options: opts,
				Message: "Select the attempt number",
			},
			&attemptNumber,
		).Prompt()

		if err != nil {
			return nil, err
		}

	}

	for i, attempt := range attempts {
		if i+1 == attemptNumber {
			return &attempt, nil
		}
	}

	return nil, fmt.Errorf("execution with attempt number %d not found", attemptNumber)
}

func printLogs(logItemsList []logs.ComponentLogItem) {
	if len(logItemsList) > 0 {
		fmt.Println()
	}
	for _, logItem := range logItemsList {
		var logEntry string
		if logItem.LogEntry != "" {
			logEntry = logItem.LogEntry
		}

		logMessage := fmt.Sprintf(
			"[%s] %s",
			utils.CS.Green(logItem.TimeGenerated),
			logEntry,
		)
		fmt.Println(logMessage)
	}
}
