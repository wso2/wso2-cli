package create

import (
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

func handleConfigCreateCommand() error {
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

	err = resolveConfigName(&params.nameFlag, existingConfigs)
	if err != nil {
		return err
	}

	err = resolveConfigType(&params.typeFlag)
	if err != nil {
		return err
	}

	err = resolveMountType(&params.mountTypeFlag)
	if err != nil {
		return err
	}

	if params.mountTypeFlag == common.MountTypeFile {
		err = resolveConfigMountPath(&params.fileMountPathFlag)
		if err != nil {
			return err
		}

		err = resolveConfigMountContent(&params.fileMountContentFlag)
		if err != nil {
			return err
		}
	} else {
		envVars, err := resolveEnvValues(params.envVars)
		if err != nil {
			return err
		}
		params.envVars = envVars
	}

	err = common.CreateConfig(common.CreateConfigParams{
		ConfigName:    params.nameFlag,
		ConfigType:    params.typeFlag,
		MountType:     params.mountTypeFlag,
		EnvVars:       params.envVars,
		FileMountPath: params.fileMountPathFlag,
		OrgId:         selectedOrg.ID,
		OrgUuid:       selectedOrg.UUID,
		ProjectId:     project.ID,
		EnvId:         selectedEnv.ID,
		AppEnvId:      matchingAppEnv.ReleaseId,
		ComponentId:   remoteComponent.Id,
	})
	if err != nil {
		return err
	}

	utils.PrintInfo(i18n.T("\nConfig '%s' has been successfully created!\n"), params.nameFlag)

	time.Sleep(1 * time.Second)

	utils.PrintInfo("%s", heredoc.Docf(i18n.T(`
			
		To view details of the created config :
			%s
		
		To list all the configs available within your component :
			%s
	
		To deploy this component :
			%s

	`),
		fmt.Sprintf(`$ wso2-integration-platform describe config --project="%s" --component="%s" --env="%s" --deployment-track="%s" --name="%s"`, project.Name, remoteComponent.Name, selectedEnv.Name, deploymentTrack.Branch, params.nameFlag),
		fmt.Sprintf(`$ wso2-integration-platform list configs --project="%s" --component="%s" --env="%s" --deployment-track="%s"`, project.Name, remoteComponent.Name, selectedEnv.Name, deploymentTrack.Branch),
		fmt.Sprintf(`$ wso2-integration-platform create deployment --project="%s" --component="%s" --deployment-track="%s" --env="%s" `, project.Name, remoteComponent.Name, deploymentTrack.Branch, selectedEnv.Name)))

	return nil
}
