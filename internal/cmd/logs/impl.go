package logs

import (
	"errors"
	"strings"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

func HandleComponentLogs(params *LogsParams) error {
	selectedOrg, err := common.ResolveTargetOrganization(params.orgFlag)
	if err != nil {
		return err
	}

	err = resolveLogType(&params.typeFlag)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(selectedOrg, params.projectFlag)
	if err != nil {
		return err
	}

	if params.typeFlag == ApplicationLog || params.typeFlag == GatewayLog || params.typeFlag == ProjectLog {

		cloudDataPlanes, dataPlanes, err := getDataPlaneInfo(selectedOrg.ID, selectedOrg.UUID)
		if err != nil {
			return err
		}

		if params.typeFlag == ApplicationLog || params.typeFlag == GatewayLog {

			remoteComponent, err := common.ResolveTargetComponent(selectedOrg, project.ID, params.componentFlag)
			if err != nil {
				return err
			}

			compInfoWithRepo, err := common.GetComponentWithRepoData(selectedOrg.ID, remoteComponent.Handler, project.ID)

			if err != nil {
				return err
			}

			deploymentTrack, err := common.ResolveDeploymentTrack(compInfoWithRepo.DeploymentTracks, params.deploymentTrackFlag)
			if err != nil {
				return err
			}

			selectedEnv, err := common.GetProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID, params.envFlag)
			if err != nil {
				return err
			}

			selectedDataPlaneHost, isCilium := resolveDataPlane(*selectedEnv, cloudDataPlanes, dataPlanes)

			if params.typeFlag == GatewayLog &&
				!strings.HasSuffix(remoteComponent.DisplayType, "Service") &&
				!strings.HasSuffix(remoteComponent.DisplayType, "Webhook") &&
				remoteComponent.DisplayType != component.DisplayTypeGitProxy {
				return errors.New(i18n.T("gateway logs is only applicable for service, webhook and proxy type components"))
			} else if params.typeFlag == ApplicationLog && remoteComponent.DisplayType == component.DisplayTypeProxy {
				return errors.New(i18n.T("application logs is not applicable for proxy type components"))
			}

			err = printComponentLogs(logsParams, selectedOrg.ID, *remoteComponent, *selectedEnv, selectedDataPlaneHost, isCilium, []string{deploymentTrack.ApiVersion}, []string{deploymentTrack.Id})
			if err != nil {
				return err
			}

		} else if params.typeFlag == ProjectLog {

			selectedEnv, err := common.GetProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID, params.envFlag)
			if err != nil {
				return err
			}

			selectedDataPlaneHost, isCilium := resolveDataPlane(*selectedEnv, cloudDataPlanes, dataPlanes)

			err = printProjectLogs(logsParams, *project, *selectedEnv, selectedDataPlaneHost, isCilium)
			if err != nil {
				return err
			}

		}

	} else if params.typeFlag == BuildLog {

		remoteComponent, err := common.ResolveTargetComponent(selectedOrg, project.ID, params.componentFlag)
		if err != nil {
			return err
		}

		remoteComWithRepoData, err := common.GetComponentWithRepoData(selectedOrg.ID, remoteComponent.Handler, project.ID)
		if err != nil {
			return err
		}

		deploymentTrack, err := common.ResolveDeploymentTrack(remoteComWithRepoData.DeploymentTracks, params.deploymentTrackFlag)
		if err != nil {
			return err
		}

		versionId := deploymentTrack.Id

		// Current Git API proxy impl doesn't have deployment tracks
		if remoteComponent.DisplayType == component.DisplayTypeGitProxy {
			for _, v := range remoteComWithRepoData.ApiVersions {
				if v.Latest {
					versionId = v.VersionId
				}
			}
		}

		selectedBuild, err := common.ResolveBuild(selectedOrg, project.Handler, versionId, remoteComponent.Name, params.runIdFlag)
		if err != nil {
			return err
		}

		// Argo build status contains cluster id and GitHub runner builds does not.
		isArgo := selectedBuild.Status.ClusterId != ""

		if !isArgo {
			if strings.HasPrefix(strings.ToLower(remoteComponent.DisplayType), "mi") {
				err = printMiBuildLogs(selectedOrg.ID, remoteComponent.Id, selectedBuild.Status.RunID)
			} else {
				err = printBuildLogs(selectedOrg.Handle, selectedOrg.ID, project.ID, remoteComponent.DisplayType, remoteComponent.Id, selectedBuild.Status.RunID)
			}
		} else {
			dataPlanes, cloudDataPlanes, err := auth.DevopsClient.GetAllDataPlaneClusters(selectedOrg.ID, selectedOrg.UUID)
			if err != nil {
				return err
			}

			selectedDataPlaneHost, err := common.GetDataPlaneHost(cloudDataPlanes, dataPlanes, selectedBuild.Status.ClusterId)
			if err != nil {
				return err
			}

			err = printArgoBuildLogs(logsParams, selectedOrg.ID, *remoteComponent, selectedDataPlaneHost, deploymentTrack.Id, selectedBuild.Status.BuildRef)
			if err != nil {
				return err
			}
		}

		if err != nil {
			return err
		}

	}

	return nil
}
