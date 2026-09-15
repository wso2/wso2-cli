package delete

import (
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/commons"
)

func handleConfigDeleteCommand() error {

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

	existingConfigs, err := common.GetConfigsOfComponent(selectedOrg.ID, selectedOrg.UUID, selectedEnv.ID, project.ID, matchingAppEnv.ReleaseId, remoteComponent.Id, configsMounts)
	if err != nil {
		return err
	}

	selectedConfig, err := common.ResolveTargetConfig(existingConfigs, params.nameFlag)
	if err != nil {
		return err
	}

	mountItem, err := common.GetMountForConfig(selectedConfig, configsMounts)
	if err != nil {
		return err
	}

	if !params.skipConfirm {
		okToDelete := commons.ConfirmDeleteWithName(selectedConfig.Name)

		if !okToDelete {
			fmt.Fprintln(utils.IO.Out, "Config delete cancelled")
			return nil
		}
	}

	err = common.DeleteConfigMount(selectedOrg.ID, selectedOrg.UUID, remoteComponent.Id, matchingAppEnv.ReleaseId, mountItem.ContainerID, mountItem.ID, project.ID)
	if err != nil {
		return err
	}

	utils.PrintInfo(i18n.T("\nConfig '%s' has been successfully deleted!\n"), selectedConfig.Name)

	return nil
}
