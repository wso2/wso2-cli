package describe

import (
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
)

func handleDescribeConfig(params *DescribeConfigParams) error {
	outputFormat, err := common.ParseOutputFormat(params.outputFlag)
	if err != nil {
		return err
	}

	// When --output json is piped or redirected, an interactive config-name
	// prompt would contaminate the JSON stream.
	// Require --name in that case; on a real terminal the prompt is still fine.
	//
	// Use a config-describe-specific message rather than the generic
	// CreateNonInteractiveError: in structured mode the user typically also
	// needs the scoping flags (no picker is available to resolve them), and the
	// quickest fix is often to drop the pipe and let the picker run.
	if outputFormat.IsStructured() && params.nameFlag == "" && !utils.IO.IsStdoutTTY() {
		return fmt.Errorf("%s", heredoc.Doc(i18n.T(`
			--name is required with --output=json when output is piped or redirected.

			Provide the config (and the flags that identify it) explicitly:
			  wso2-integration-platform describe config --name=<config-name> --project=<project> --component=<component> --env=<env> --deployment-track=<track>

			Or run the command on a terminal without a pipe to choose a config interactively:
			  wso2-integration-platform describe config --output=json`)))
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

	deploymentTrack, err := common.ResolveDeploymentTrack(remoteComponent.DeploymentTracks, params.deploymentTrackFlag)
	if err != nil {
		return err
	}

	selectedEnv, err := common.GetProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID, params.envFlag)
	if err != nil {
		return err
	}

	remoteComWithRepoData, err := common.GetComponentWithRepoData(selectedOrg.ID, remoteComponent.Handler, project.ID)
	if err != nil {
		return err
	}

	matchingAppEnv, err := common.GetReleaseEnvForDeploymentTrack(remoteComWithRepoData, deploymentTrack.Id, selectedEnv.ID)
	if err != nil {
		return err
	}

	configsMounts, err := common.GetConfigMounts(selectedOrg.ID, selectedOrg.UUID, remoteComponent.Id, matchingAppEnv.ReleaseId, project.ID)
	if err != nil {
		return err
	}

	configs, err := common.GetConfigsOfComponent(selectedOrg.ID, selectedOrg.UUID, selectedEnv.ID, project.ID, matchingAppEnv.ReleaseId, remoteComponent.Id, configsMounts)
	if err != nil {
		return err
	}

	selectedConfig, err := common.ResolveTargetConfig(configs, params.nameFlag)
	if err != nil {
		return err
	}

	var configMapData devops.ConfigItem

	if selectedConfig.SecretType == "" {
		configMapData, err = getConfigMapData(selectedOrg.ID, selectedOrg.UUID, selectedEnv.ID, project.ID, selectedConfig.ID)
		if err != nil {
			return err
		}

	} else {
		configMapData, err = getSecretData(selectedOrg.ID, selectedOrg.UUID, selectedEnv.ID, project.ID, selectedConfig.ID)
		if err != nil {
			return err
		}
	}

	if outputFormat.IsStructured() {
		return emitConfigJSON(configMapData, configsMounts, outputFormat)
	}

	err = printConfigDetails(configMapData, configsMounts)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

		To add a new config or secret:
			%s

		To list all the configs available within your component :
			%s

	`),
		fmt.Sprintf(`$ wso2-integration-platform create config --project="%s" --component="%s" --env="%s" --deployment-track="%s"`, project.Name, remoteComponent.Name, selectedEnv.Name, deploymentTrack.Branch),
		fmt.Sprintf(`$ wso2-integration-platform list configs --project="%s" --component="%s" --env="%s" --deployment-track="%s"`, project.Name, remoteComponent.Name, selectedEnv.Name, deploymentTrack.Branch)))

	return nil
}
