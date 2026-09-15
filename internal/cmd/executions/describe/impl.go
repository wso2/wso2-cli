package describe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/textstyle"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func handleDescribeImpl(params *DescribeExecutionParams) error {
	outputFormat, err := common.ParseOutputFormat(params.Output)
	if err != nil {
		return err
	}

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

	// When --output json is piped or redirected (stdout is not a TTY), any
	// interactive picker would contaminate the JSON stream. The execution is
	// already pinned by the required --id, but component and env still fall back
	// to interactive pickers when omitted. Require them up front in structured
	// non-TTY mode rather than letting a downstream resolver fail with a
	// low-level "could not open a new TTY" error.
	//
	// --project is intentionally not required here: it can be resolved from the
	// local context file (see ResolveContext below), so demanding the flag
	// would reject a valid context-based setup.
	if outputFormat.IsStructured() && !utils.IO.IsStdoutTTY() &&
		(params.Component == "" || params.Env == "") {
		return fmt.Errorf("%s", heredoc.Doc(i18n.T(`
			--component and --env are required with --output=json when output is piped or redirected.

			Provide the execution (and the flags that identify it) explicitly:
			  wso2-integration-platform describe execution --id=<execution-id> --project=<project> --component=<component> --env=<env>

			Or run the command on a terminal without a pipe to choose them interactively:
			  wso2-integration-platform describe execution --id=<execution-id> --output=json`)))
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
		msg := i18n.T("Component initialization is in progress. Please run the command in a while...\n")
		// In JSON mode keep stdout empty so machine consumers see a clean
		// (empty) stream; the explanation still reaches the user via stderr.
		if outputFormat.IsStructured() {
			fmt.Fprint(utils.IO.ErrOut, msg)
		} else {
			utils.PrintInfo("%s", msg)
		}
		return nil
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

	execSpinner := utils.CreateSpinner("Fetching execution details...", "")
	execSpinner.Start()
	execution, err := auth.LogsClient.GetExecutionById(
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

	execSpinner = utils.CreateSpinner("Fetching execution attempts...", "")
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

	if outputFormat.IsStructured() {
		return emitExecutionJSON(execution, executionAttempts, outputFormat)
	}

	displayExecutionInformation(execution, executionAttempts)

	return nil
}

// emitExecutionJSON renders the resolved execution (and its attempts) as a JSON
// object. It is only called when --output=json. Decorative text and section
// headers are dropped because the payload describes the resource, not the CLI's
// layout. The status mapping and duration computation mirror
// displayExecutionInformation exactly.
func emitExecutionJSON(execution *logs.ExecutionListItemV2, attempts []logs.ExecutionAttempt, format common.OutputFormat) error {
	stVal := parseEpoch(execution.StartTime)

	out := ExecutionDescribeOutput{
		Name:      execution.ID,
		Status:    mapExecutionStatus(execution.Status),
		Revision:  shortRevision(execution.RevisionID),
		Duration:  executionDuration(execution.StartTime, execution.CompletionTime),
		StartTime: utils.TimeAgo(time.Unix(int64(stVal), 0)),
		Attempts:  make([]ExecutionAttemptOutput, 0, len(attempts)),
	}

	for num, attempt := range attempts {
		// The text path computes attempt duration from the execution's own
		// start/completion times (not the attempt's); keep that behavior.
		out.Attempts = append(out.Attempts, ExecutionAttemptOutput{
			Number:   num + 1,
			ID:       attempt.ID,
			Status:   mapExecutionStatus(execution.Status),
			Duration: executionDuration(execution.StartTime, execution.CompletionTime),
			Started:  fmt.Sprintf("%s ago", utils.TimeAgo(time.Unix(int64(stVal), 0))),
		})
	}

	rendered, err := common.RenderStructured(format, out)
	if err != nil {
		return err
	}
	fmt.Fprintln(utils.IO.Out, rendered)
	return nil
}

// mapExecutionStatus normalizes the API status into the CLI's vocabulary,
// matching the switch used in the text path.
func mapExecutionStatus(apiStatus string) string {
	switch apiStatus {
	case "Failed":
		return "failure"
	case "Succeeded":
		return "success"
	default:
		return "in-progress"
	}
}

// shortRevision returns the first 8 characters of a revision id, or the whole
// string when it is shorter, avoiding a slice-out-of-range panic.
func shortRevision(revisionID string) string {
	if len(revisionID) < 8 {
		return revisionID
	}
	return revisionID[:8]
}

// parseEpoch parses an epoch-seconds string, returning 0 on a parse error to
// match the text path's lenient handling.
func parseEpoch(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

// executionDuration computes the wall-clock duration between two epoch-seconds
// strings, returning "-" when the completion time is missing or unparseable.
func executionDuration(startTime, completionTime string) string {
	if completionTime == "" {
		return "-"
	}
	ctVal, err := strconv.Atoi(completionTime)
	if err != nil {
		return "-"
	}
	stVal := parseEpoch(startTime)
	return time.Unix(int64(ctVal), 0).Sub(time.Unix(int64(stVal), 0)).String()
}

func displayExecutionInformation(execution *logs.ExecutionListItemV2, attempts []logs.ExecutionAttempt) {
	status := "in-progress"
	duration := "-"

	switch execution.Status {
	case "Failed":
		status = "failure"
	case "Succeeded":
		status = "success"
	default:
		status = "in-progress"
	}

	stVal, err := strconv.Atoi(execution.StartTime)

	if err != nil {
		fmt.Printf("Error parsing StartTime for execution %s: %v\n", execution.ID, err)
		stVal = 0
	}

	if execution.CompletionTime != "" {
		ctVal, err := strconv.Atoi(execution.CompletionTime)

		if err != nil {
			fmt.Printf("Error parsing StartTime for execution %s: %v\n", execution.ID, err)
		} else {
			duration = time.Unix(int64(ctVal), 0).Sub(time.Unix(int64(stVal), 0)).String()
		}
	}

	utils.PrintInfo("\nName: %s\n", execution.ID)
	utils.PrintInfo("Status: %s\n", status)
	utils.PrintInfo("Revision: %s\n", shortRevision(execution.RevisionID))
	utils.PrintInfo("Duration: %s\n", duration)
	utils.PrintInfo("Start Time: %s\n", utils.TimeAgo(time.Unix(int64(stVal), 0)))

	if len(attempts) > 0 {
		utils.PrintInfo("\nAttempts:\n")
	}

	for num, attempt := range attempts {
		attemptStatus := "in-progress"
		attemptDuration := "-"

		switch execution.Status {
		case "Failed":
			attemptStatus = "failure"
		case "Succeeded":
			attemptStatus = "success"
		default:
			attemptStatus = "in-progress"
		}

		attemptStVal, err := strconv.Atoi(execution.StartTime)

		if err != nil {
			fmt.Printf("Error parsing StartTime for execution %s: %v\n", execution.ID, err)
			attemptStVal = 0
		}

		if execution.CompletionTime != "" {
			ctVal, err := strconv.Atoi(execution.CompletionTime)

			if err != nil {
				fmt.Printf("Error parsing StartTime for execution %s: %v\n", execution.ID, err)
			} else {
				attemptDuration = time.Unix(int64(ctVal), 0).Sub(time.Unix(int64(attemptStVal), 0)).String()
			}
		}

		utils.PrintInfo("  - Number: %d\n", num+1)
		utils.PrintInfo("    ID: %s\n", attempt.ID)
		utils.PrintInfo("    Status: %s\n", attemptStatus)
		utils.PrintInfo("    Duration: %s\n", attemptDuration)
		utils.PrintInfo("    Started: %s ago\n\n", utils.TimeAgo(time.Unix(int64(attemptStVal), 0)))
	}

}

func getRunStatus(status string) string {
	successT := textstyle.CreateTextStyle().SetTextColor(textstyle.GREEN)
	failureT := textstyle.CreateTextStyle().SetTextColor(textstyle.RED)
	midT := textstyle.CreateTextStyle().SetTextColor(textstyle.BLUE)

	switch status {
	case "success":
		return successT.Text(status)
	case "failure":
		return failureT.Text(status)
	default:
		return midT.Text(status)
	}
}
