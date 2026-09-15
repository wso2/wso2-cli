package describe

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

func HandleDescribeComponent(params *DescribeComponentParams) error {
	outputFormat, err := common.ParseOutputFormat(params.outputFlag)
	if err != nil {
		return err
	}

	// When --output json is piped or redirected, an interactive component-name
	// prompt would contaminate the JSON stream.
	// Require the component name in that case; on a real terminal the prompt is
	// still fine.
	//
	// Use a component-describe-specific message rather than the generic
	// CreateNonInteractiveError: in structured mode the user typically also
	// needs the scoping flags (no picker is available to resolve them), and the
	// quickest fix is often to drop the pipe and let the picker run.
	if outputFormat.IsStructured() && params.componentFlag == "" && !utils.IO.IsStdoutTTY() {
		return fmt.Errorf("%s", heredoc.Doc(i18n.T(`
			A component name is required with --output=json when output is piped or redirected.

			Provide the component (and the flags that identify it) explicitly:
			  wso2-integration-platform describe component <component-name> --project=<project> --org=<org>

			Or run the command on a terminal without a pipe to choose a component interactively:
			  wso2-integration-platform describe component --output=json`)))
	}

	orgId, projectId, err := common.ResolveContext(params.orgFlag, params.projectFlag)

	if err == nil {
		if params.orgFlag == "" {
			params.orgFlag = orgId
		}

		if params.projectFlag == "" {
			params.projectFlag = projectId
		}
	}

	selectedOrg, err := common.ResolveTargetOrganization(params.orgFlag)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(selectedOrg, params.projectFlag)
	if err != nil {
		return err
	}

	remoteComponent, err := common.ResolveTargetComponent(selectedOrg, project.ID, params.componentFlag)
	if err != nil {
		return err
	}

	componentWithRepoData, err := common.GetComponentWithRepoData(selectedOrg.ID, remoteComponent.Handler, project.ID)
	if err != nil {
		return err
	}

	envs, err := common.GetAllProjectEnv(selectedOrg.UUID, selectedOrg.ID, componentWithRepoData.ProjectId)
	if err != nil {
		return err
	}

	isProxy := remoteComponent.DisplayType == component.DisplayTypeGitProxy

	if outputFormat.IsStructured() {
		return emitComponentJSON(&componentWithRepoData, selectedOrg, *project, envs, isProxy, outputFormat)
	}

	PrintGeneralComponentInfo(&componentWithRepoData, selectedOrg, *project)

	PrintRepoInfo(&componentWithRepoData)

	for _, env := range *envs {
		if isProxy {
			err = DisplayProxyDeploymentInfo(selectedOrg, &componentWithRepoData, &env)
			if err != nil {
				return err
			}
		} else {
			err = PrintDeploymentInfoForEnv(&componentWithRepoData, selectedOrg, &env)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
