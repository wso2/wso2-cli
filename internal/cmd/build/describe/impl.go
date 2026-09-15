package describe

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/build/buildstep"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/textstyle"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
)

var successStyle = textstyle.CreateTextStyle().SetTextColor(textstyle.GREEN)
var failureStyle = textstyle.CreateTextStyle().SetTextColor(textstyle.RED)
var grayStyle = textstyle.CreateTextStyle().SetTextColor(textstyle.BLACK).AddTextStyle(textstyle.BRIGHT)
var boldText = textstyle.CreateTextStyle().AddTextStyle(textstyle.BOLD)

func HandleDescribeBuild(params *BuildDescribeParams) error {
	outputFormat, err := common.ParseOutputFormat(params.OutputFlag)
	if err != nil {
		return err
	}

	// When --output json is piped or redirected (stdout is not a TTY), an
	// interactive build-id prompt would contaminate the JSON stream. Require
	// --build-id in that case; on a real terminal the prompt is still fine.
	//
	// Use a build-describe-specific message rather than the generic
	// CreateNonInteractiveError: in structured mode the user typically also
	// needs the scoping flags (no picker is available to resolve them), and the
	// quickest fix is often to drop the pipe and let the picker run.
	if outputFormat.IsStructured() && params.RunIdFlag == 0 && !utils.IO.IsStdoutTTY() {
		return fmt.Errorf("%s", heredoc.Doc(i18n.T(`
			A build id is required with --output=json when output is piped or redirected.

			Provide the build id (and the flags that identify it) explicitly:
			  wso2-integration-platform describe build <build-id> --project=<project> --component=<component> --deployment-track=<track>

			Or run the command on a terminal without a pipe to choose a build interactively:
			  wso2-integration-platform describe build --output=json`)))
	}

	selectedOrg, err := common.ResolveTargetOrganization(params.OrgFlag)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(selectedOrg, params.ProjectFlag)
	if err != nil {
		return err
	}

	remoteComponent, err := common.ResolveTargetComponent(selectedOrg, project.ID, params.ComponentFlag)
	if err != nil {
		return err
	}

	if strings.HasPrefix(strings.ToLower(remoteComponent.DisplayType), "byoi") {
		// In JSON mode, keep stdout empty so machine consumers see a clean
		// (empty) payload; the explanation still goes to the user via stderr.
		if outputFormat.IsStructured() {
			fmt.Fprint(utils.IO.ErrOut, i18n.T("Cannot describe a build for pre-built image based components\n"))
		} else {
			utils.PrintInfo("%s", i18n.T("Cannot describe a build for pre-built image based components\n"))
		}
		return nil
	}

	remoteComWithRepoData, err := common.GetComponentWithRepoData(selectedOrg.ID, remoteComponent.Handler, project.ID)
	if err != nil {
		return err
	}

	deploymentTrack, err := common.ResolveDeploymentTrack(remoteComWithRepoData.DeploymentTracks, params.DeploymentTrackFlag)
	if err != nil {
		return err
	}

	dtString := deploymentTrack.Id

	if remoteComWithRepoData.DisplayType == component.DisplayTypeGitProxy {
		for _, v := range remoteComWithRepoData.ApiVersions {
			if v.Latest {
				dtString = v.VersionId
				break
			}
		}
	}

	selectedBuild, err := common.ResolveBuild(selectedOrg, project.Handler, dtString, remoteComponent.Name, params.RunIdFlag)
	if err != nil {
		return err
	}

	completedAt := "-"
	if selectedBuild.Status.CompletedAt != "" && selectedBuild.Status.Conclusion != "" {
		completedAt = fmt.Sprintf("%s ago", utils.GetRelativeTime(selectedBuild.Status.CompletedAt))
	}

	if !outputFormat.IsStructured() {
		infoArr := make([]string, 0)

		infoArr = append(infoArr, fmt.Sprintf("Build ID: %s", fmt.Sprintf("%d", selectedBuild.Status.RunID)))
		infoArr = append(infoArr, fmt.Sprintf("Started: %s", fmt.Sprintf("%s ago", utils.GetRelativeTime(selectedBuild.Status.StartedAt))))
		if completedAt != "-" {
			infoArr = append(infoArr, fmt.Sprintf("Completed: %s", completedAt))
		}
		infoArr = append(infoArr, fmt.Sprintf("Status: %s", common.GetBuildStatusStr(*selectedBuild)))

		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

				Build details:

				%s

				Commit details:
				Commit hash: %s
				Commit message: %s

			`),
			strings.Join(infoArr, "\n"),
			selectedBuild.Spec.Revision,
			selectedBuild.Status.GitCommit.Message,
		))
	}

	var status *component.DeploymentBuildStatusResponse
	if selectedBuild.Status.Conclusion == "failure" {
		bsSpin := utils.CreateSpinner("Fetching build steps...", "")
		bsSpin.Start()

		// Argo build status contains cluster id and GitHub runner builds does not.
		isArgo := selectedBuild.Status.ClusterId != ""

		if isArgo {
			dataPlanes, cloudDataPlanes, err := auth.DevopsClient.GetAllDataPlaneClusters(selectedOrg.ID, selectedOrg.UUID)
			if err != nil {
				return err
			}

			dataPlaneHost, err := common.GetDataPlaneHost(cloudDataPlanes, dataPlanes, selectedBuild.Status.ClusterId)
			if err != nil {
				return err
			}

			status, err = common.GetArgoBuildLog(selectedOrg.ID, dataPlaneHost, *remoteComponent, deploymentTrack.Id, selectedBuild.Status.BuildRef)
			if err != nil {
				return err
			}
		} else {
			status, err = auth.ComponentClient.GetComponentBuildStatus(
				selectedOrg.Handle,
				project.ID,
				remoteComponent.Id,
				selectedBuild.Status.RunID,
				selectedOrg.ID,
			)
		}
		bsSpin.Stop()

		if err != nil {
			return err
		}

		if !outputFormat.IsStructured() {
			initSuccess := true
			buildSuccess := true

			fmt.Fprintln(utils.IO.Out, "Build steps:")
			initSuccess, tableData := populateBuildDataTable(status.Data.Init.Steps)
			fmt.Fprintln(utils.IO.Out, boldText.Text(" Initialization"))
			fmt.Fprintln(utils.IO.Out, utils.CreateTable(tableData, []string{}, ""))

			if initSuccess {
				buildSuccess, tableData = populateBuildDataTable(status.Data.Build.Steps)
				fmt.Fprintln(utils.IO.Out, boldText.Text(" Build"))
				fmt.Fprintln(utils.IO.Out, utils.CreateTable(tableData, []string{}, ""))
			} else {
				buildSuccess = false
				fmt.Fprintln(utils.IO.Out, fmt.Sprintf("%s %s", boldText.Text(" Build"), grayStyle.Text("(Skipped)")))
			}

			if buildSuccess {
				_, tableData = populateBuildDataTable(status.Data.Deploy.Steps)
				fmt.Fprintln(utils.IO.Out, boldText.Text(" Finalization"))
				fmt.Fprintln(utils.IO.Out, utils.CreateTable(tableData, []string{}, ""))
			} else {
				fmt.Fprintln(utils.IO.Out, fmt.Sprintf("%s %s", boldText.Text(" Finalization"), grayStyle.Text("(Skipped)")))
			}
		}
	}

	if outputFormat.IsStructured() {
		return emitBuildJSON(selectedBuild, status, outputFormat)
	}

	if selectedBuild.Status.Status == "completed" {
		time.Sleep(1 * time.Second)

		utils.PrintInfo("%s", heredoc.Docf(`

		To view the logs for this build :
			%s

		`,
			fmt.Sprintf(`$ wso2-integration-platform logs build --project="%s" --component="%s" --deployment-track="%s" --build-id=%s`,
				project.Name,
				remoteComponent.Name,
				deploymentTrack.Branch,
				strconv.Itoa(int(selectedBuild.Status.RunID)),
			),
		))

		if selectedBuild.Status.Conclusion == "success" {
			utils.PrintInfo("%s", heredoc.Docf(`
				To deploy this component to the Development environment :
					%s

			`,
				fmt.Sprintf(
					`$ wso2-integration-platform create deployment "%s" --env=Development --project="%s" --deployment-track="%s" --build-id=%s`,
					remoteComponent.Name,
					project.Name,
					deploymentTrack.Branch,
					strconv.Itoa(int(selectedBuild.Status.RunID)),
				),
			))
		}
	}

	return nil
}

// emitBuildJSON renders the resolved build (and step data when available) as a
// JSON object. It is only called when --output=json. UX affordances (the
// "to view logs" / "to deploy" hints, the 1s sleep) are intentionally skipped:
// the JSON payload describes the resource, not the CLI's suggestions.
func emitBuildJSON(b *deploymentbuild.BuildKind, status *component.DeploymentBuildStatusResponse, format common.OutputFormat) error {
	out := BuildDescribeOutput{
		ID:          b.Status.RunID,
		Status:      b.Status.Status,
		Conclusion:  b.Status.Conclusion,
		StartedAt:   b.Status.StartedAt,
		CompletedAt: b.Status.CompletedAt,
		Commit: BuildCommitOutput{
			Hash:    b.Spec.Revision,
			Message: b.Status.GitCommit.Message,
		},
	}

	if status != nil {
		out.Steps = &BuildStepsOutput{
			Initialization: toStepOutputs(status.Data.Init.Steps),
			Build:          toStepOutputs(status.Data.Build.Steps),
			Finalization:   toStepOutputs(status.Data.Deploy.Steps),
		}
	}

	rendered, err := common.RenderStructured(format, out)
	if err != nil {
		return err
	}
	fmt.Fprintln(utils.IO.Out, rendered)
	return nil
}

func toStepOutputs(steps []component.BuildStep) []BuildStepOutput {
	out := make([]BuildStepOutput, 0, len(steps))
	for _, s := range steps {
		out = append(out, BuildStepOutput{
			Name:        s.Name,
			Status:      s.Status,
			Conclusion:  s.Conclusion,
			StartedAt:   s.StartedAt,
			CompletedAt: s.CompletedAt,
		})
	}
	return out
}

func populateBuildDataTable(steps []component.BuildStep) (bool, [][]string) {
	overAllSuccess := true
	data := [][]string{}

	for _, step := range steps {
		var symbol string

		if step.Conclusion == "success" {
			symbol = successStyle.Text(constants.SUCCESS_SYMBOL)
		} else if step.Conclusion == "failure" {
			symbol = failureStyle.Text(constants.FAILURE_SYMBOL)
			overAllSuccess = false
		} else {
			symbol = constants.PENDING_SYMBOL
		}

		name := step.Name

		if step.Conclusion == "skipped" {
			name = fmt.Sprintf("%s %s", name, grayStyle.Text("(skipped)"))
		}

		if buildstep.HasLogs(buildstep.ResolveBuildStep(step.Name), step.Conclusion) {

			data = append(data, []string{
				symbol,
				fmt.Sprintf(
					"%s (id = %s)",
					name,
					buildstep.GetShortFormStepName(buildstep.ResolveBuildStep(step.Name)),
				),
			})
		} else {
			data = append(data, []string{
				symbol,
				name,
			})
		}
	}

	return overAllSuccess, data
}
