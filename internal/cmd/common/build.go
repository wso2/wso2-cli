package common

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func AddRunIdFlag(cmdFlags *pflag.FlagSet, bindTo *int) {
	cmdFlags.IntVar(bindTo, "build-id", 0, i18n.T("Build ID of the build"))
}

func GetDeploymentBuild(orgId string, componentName string, projectHandle string, deploymentTrackId string) (deploymentBuild *[]deploymentbuild.BuildKind, err error) {
	deploymentBuildSpinner := utils.CreateSpinner(" Fetching component build...", "")
	deploymentBuildSpinner.Start()
	deploymentBuildRes, err := auth.DeploymentBuildClient.GetDeploymentBuilds(orgId, componentName, projectHandle, deploymentTrackId)
	deploymentBuildSpinner.Stop()
	if err != nil {
		return nil, err
	}
	return &deploymentBuildRes, nil
}

func ResolveBuild(
	selectedOrg *api.Organization,
	projectHandle string,
	deploymentTrackId string,
	componentName string,
	runIdFlag int,
) (*deploymentbuild.BuildKind, error) {

	var selectedBuild *deploymentbuild.BuildKind

	buildList, err := GetDeploymentBuild(selectedOrg.ID, componentName, projectHandle, deploymentTrackId)
	if err != nil {
		return nil, err
	}

	if len(*buildList) == 0 {
		return nil, fmt.Errorf("no builds found for the deployment track of the component")
	}

	if runIdFlag != 0 {
		for _, item := range *buildList {
			if item.Status.RunID == runIdFlag {
				selectedBuild = &item
				break
			}
		}
	} else {
		selectedBuild, err = promptToSelectBuild(*buildList)
		if err != nil {
			if errors.Is(err, internal.ErrNonInteractive) {
				return nil, utils.CreateNonInteractiveError("build selection", "build-id")
			}
			return nil, fmt.Errorf("%s", i18n.T(" failed to select the build"))
		}
	}

	if selectedBuild == nil {
		return nil, fmt.Errorf("%s", i18n.T("invalid build-id selection"))
	}

	return selectedBuild, nil
}

func ResolveSucceededBuild(
	selectedOrg *api.Organization,
	projectHandle string,
	deploymentTrackId string,
	componentName string,
	runIdFlag int,
) (*deploymentbuild.BuildKind, error) {

	var selectedBuild *deploymentbuild.BuildKind

	buildList, err := GetDeploymentBuild(selectedOrg.ID, componentName, projectHandle, deploymentTrackId)
	if err != nil {
		return nil, err
	}

	var succeededBuilds []deploymentbuild.BuildKind

	for _, item := range *buildList {
		if item.Status.Conclusion == "success" {
			succeededBuilds = append(succeededBuilds, item)
		}
	}

	if len(succeededBuilds) == 0 {
		return nil, fmt.Errorf("No succeeded builds found for the selected component")
	}

	if runIdFlag != 0 {
		for _, item := range succeededBuilds {
			if item.Status.RunID == runIdFlag {
				selectedBuild = &item
				break
			}
		}
	} else {
		selectedBuild, err = promptToSelectBuild(succeededBuilds)
		if err != nil {
			if errors.Is(err, internal.ErrNonInteractive) {
				return nil, utils.CreateNonInteractiveError("build selection", "build-id")
			}
			return nil, fmt.Errorf("%s", i18n.T(" failed to select the build"))
		}
	}

	if selectedBuild == nil {
		return nil, fmt.Errorf("%s", i18n.T("invalid build-id selection"))
	}

	return selectedBuild, nil
}

func promptToSelectBuild(builds []deploymentbuild.BuildKind) (*deploymentbuild.BuildKind, error) {
	var buildPromptSelection string
	var buildPromptList []string

	for _, item := range builds {
		if item.Status.RunID > 0 {
			buildPromptList = append(
				buildPromptList,
				fmt.Sprintf(
					"%s - %s",
					strconv.Itoa(item.Status.RunID),
					fmt.Sprintf("%s(Commit ID: %s)", item.Status.GitCommit.Message, ShortCommitId(item.Spec.Revision)),
				),
			)
		}
	}

	err := prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{
			Message: "Build ID:",
			Values:  buildPromptList,
		},
		&buildPromptSelection,
	).Prompt()

	if err != nil {
		return nil, err
	}

	buildId, err := strconv.Atoi(strings.Split(buildPromptSelection, " - ")[0])
	if err != nil {
		return nil, err
	}

	for _, item := range builds {
		if item.Status.RunID == buildId {
			return &item, nil
		}
	}

	return nil, nil
}

func GetBuildStatusStr(selectedBuild deploymentbuild.BuildKind) string {
	conclusion := selectedBuild.Status.Conclusion

	if selectedBuild.Status.Conclusion == "" {
		conclusion = utils.CS.Blue(selectedBuild.Status.Status)
	} else {
		if selectedBuild.Status.Conclusion == "success" {
			conclusion = utils.CS.Green(conclusion)
		} else if selectedBuild.Status.Conclusion == "failure" {
			conclusion = utils.CS.Red(conclusion)
		}
	}
	return conclusion
}

func GetArgoBuildLog(orgId string, selectedDataPlaneHost string, component models.Component, deploymentTrackId string, workflowName string) (*component.DeploymentBuildStatusResponse, error) {
	reqBody := logs.GetBuildLogsReqBody{
		ComponentId:       component.Id,
		DeploymentTrackId: deploymentTrackId,
		WorkflowName:      workflowName,
	}

	logsSpinner := utils.CreateSpinner(fmt.Sprintf(i18n.T(" Fetching build logs for component %s "), component.Name), "\n")
	logsSpinner.Start()
	logs, err := auth.LogsClient.GetBuildLogs(reqBody, orgId, selectedDataPlaneHost)
	logsSpinner.Stop()

	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to get build logs: %w"), err)
	}

	return logs, nil

}
