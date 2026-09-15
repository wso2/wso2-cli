package describe

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	componentDescribe "github.com/wso2/integration-platform-tools/internal/cmd/component/describe"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

func HandleDescribeDeployment(params *DescribeDeploymentParams) error {
	outputFormat, err := common.ParseOutputFormat(params.outputFlag)
	if err != nil {
		return err
	}

	// When --output json is piped or redirected (stdout is not a TTY), any
	// interactive picker would contaminate the JSON stream. A deployment is
	// scoped by component and environment, each of which falls back to an
	// interactive prompt when its flag is omitted. Require them up front in
	// structured non-TTY mode rather than letting a downstream resolver fail
	// with a low-level "could not open a new TTY" error.
	//
	// --project is intentionally not required here: it can be resolved from the
	// local context file (see ResolveContext below), so demanding the flag
	// would reject a valid context-based setup.
	if outputFormat.IsStructured() && !utils.IO.IsStdoutTTY() &&
		(params.componentFlag == "" || params.envFlag == "") {
		return fmt.Errorf("%s", heredoc.Doc(i18n.T(`
			--component and --env are required with --output=json when output is piped or redirected.

			Provide the deployment (and the flags that identify it) explicitly:
			  wso2-integration-platform describe deployment --project=<project> --component=<component> --env=<env>

			Or run the command on a terminal without a pipe to choose them interactively:
			  wso2-integration-platform describe deployment --output=json`)))
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

	selectedEnv, err := common.GetProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID, params.envFlag)
	if err != nil {
		return err
	}

	isProxy := remoteComponent.DisplayType == component.DisplayTypeGitProxy

	if outputFormat.IsStructured() {
		// Reuse the component describe builders: a deployment describe is the
		// per-environment deployment shape for a single env.
		var envOut componentDescribe.EnvironmentDeploymentOutput
		if isProxy {
			envOut, err = componentDescribe.BuildProxyEnvOutput(selectedOrg, &componentWithRepoData, selectedEnv)
		} else {
			envOut, err = componentDescribe.BuildDeploymentEnvOutput(&componentWithRepoData, selectedOrg, selectedEnv)
		}
		if err != nil {
			return err
		}

		rendered, err := common.RenderStructured(outputFormat, envOut)
		if err != nil {
			return err
		}
		fmt.Fprintln(utils.IO.Out, rendered)
		return nil
	}

	if isProxy {
		err = componentDescribe.DisplayProxyDeploymentInfo(selectedOrg, &componentWithRepoData, selectedEnv)
		if err != nil {
			return err
		}
	} else {
		err = componentDescribe.PrintDeploymentInfoForEnv(&componentWithRepoData, selectedOrg, selectedEnv)
		if err != nil {
			return err
		}
	}

	return nil
}
