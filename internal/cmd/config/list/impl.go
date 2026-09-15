package list

import (
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	projectapi "github.com/wso2/integration-platform-tools/pkg/api/project"
)

func handleConfigListCommand() error {
	// An invalid --output falls back to a persisted OUTPUT_FORMAT (with a
	// warning); with no global set it is an error.
	outputFormat, err := common.ResolveOutputFormat(params.outputFlag)
	if err != nil {
		return err
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

	remoteComWithRepoData, err := common.GetComponentWithRepoData(selectedOrg.ID, remoteComponent.Handler, project.ID)
	if err != nil {
		return err
	}

	// Determine which environments to list configs for. When --env is omitted
	// we list configs across every environment; when it is set we scope to that
	// single environment.
	var targetEnvs []projectapi.ProjectEnvironment
	if params.envFlag != "" {
		selectedEnv, err := common.GetProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID, params.envFlag)
		if err != nil {
			return err
		}
		targetEnvs = []projectapi.ProjectEnvironment{*selectedEnv}
	} else {
		envs, err := common.GetAllProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID)
		if err != nil {
			return err
		}
		targetEnvs = *envs
	}

	// Collect configs per environment. A single explicit env keeps strict
	// error semantics; when listing all envs we skip environments that have no
	// release for this deployment track (i.e. the component is not deployed
	// there) instead of failing the whole command.
	scopedToSingleEnv := params.envFlag != ""
	var configs []configWithEnv
	for _, env := range targetEnvs {
		matchingAppEnv, err := common.GetReleaseEnvForDeploymentTrack(remoteComWithRepoData, deploymentTrack.Id, env.ID)
		if err != nil {
			if scopedToSingleEnv {
				return err
			}
			// No release for this env/track combination — nothing deployed
			// here, so there are no configs to list. Skip it.
			continue
		}

		configsMounts, err := common.GetConfigMounts(selectedOrg.ID, selectedOrg.UUID, remoteComponent.Id, matchingAppEnv.ReleaseId, project.ID)
		if err != nil {
			return err
		}

		envConfigs, err := common.GetConfigsOfComponent(selectedOrg.ID, selectedOrg.UUID, env.ID, project.ID, matchingAppEnv.ReleaseId, remoteComponent.Id, configsMounts)
		if err != nil {
			return err
		}

		for _, c := range envConfigs {
			configs = append(configs, configWithEnv{ConfigItem: c, EnvName: env.Name})
		}
	}

	if outputFormat.IsStructured() {
		return printConfigListStructured(configs, outputFormat, *remoteComponent, *project)
	}

	err = printConfigList(configs)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	utils.PrintInfo("%s", heredoc.Docf(i18n.T(`
			
		To view details of a particular config :
			%s
		
		To add a new config or secret :
			%s

	`),
		fmt.Sprintf(`$ wso2-integration-platform describe config --project="%s" --component="%s" --env=<env-name> --deployment-track="%s" --name=<config-name>`, project.Name, remoteComponent.Name, deploymentTrack.Branch),
		fmt.Sprintf(`$ wso2-integration-platform create config --project="%s" --component="%s" --deployment-track="%s" --env=<env-name>`, project.Name, remoteComponent.Name, deploymentTrack.Branch)))

	return nil
}
