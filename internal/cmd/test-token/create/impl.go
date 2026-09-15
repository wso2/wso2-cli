package create

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func handleGetApiKey(params *GetTokenParams) error {

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

	compInfoWithRepo, err := common.GetComponentWithRepoData(selectedOrg.ID, remoteComponent.Handler, project.ID)

	if err != nil {
		utils.HandleErr(fmt.Errorf(i18n.T("Error resolving component: %w"), err))
	}

	deploymentTrack, err := common.ResolveDeploymentTrack(compInfoWithRepo.DeploymentTracks, params.deploymentTrackFlag)
	if err != nil {
		return err
	}

	selectedEnv, err := common.GetProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID, params.envFlag)
	if err != nil {
		return err
	}

	if selectedEnv.Critical {
		return fmt.Errorf("%s", i18n.T("Cannot generate API-keys for critical environments"))
	}

	if remoteComponent.DisplayType == component.DisplayTypeGitProxy {
		var apiV *models.ApiVersion

		for _, entry := range compInfoWithRepo.ApiVersions {
			if entry.Latest {
				apiV = &entry
				break
			}
		}
		selectedEndpoint, err := getDeploymentInfoForProxy(
			selectedOrg.ID,
			selectedOrg.Handle,
			selectedOrg.UUID,
			remoteComponent.Id,
			apiV.Id,
			selectedEnv.ID,
		)
		if err != nil {
			return err
		}

		apiKetResponse, err := GetApiKeyForEndpoint(selectedEndpoint.APIID, selectedOrg, *selectedEnv)
		if err != nil {
			return err
		}

		fmt.Fprint(utils.IO.Out, heredoc.Docf(i18n.T(`

            Include the following header when calling the deployed endpoint :
            
                Api-Key: %s
            
            Api key is valid for %d seconds
        `), utils.CS.Bold(apiKetResponse.Apikey), apiKetResponse.ValidityTime))
	} else {
		selectedEndpoint, err := getEndpointForEnv(selectedOrg.ID, *remoteComponent, *selectedEnv, *deploymentTrack, params.endpointFlag)
		if err != nil {
			return err
		}

		apiKetResponse, err := GetApiKeyForEndpoint(selectedEndpoint.APIMID, selectedOrg, *selectedEnv)
		if err != nil {
			return err
		}

		fmt.Fprint(utils.IO.Out, heredoc.Docf(i18n.T(`

            Include the following header when calling the deployed endpoint :
            
                Api-Key: %s
            
            Api key is valid for %d seconds
        `), utils.CS.Bold(apiKetResponse.Apikey), apiKetResponse.ValidityTime))
	}

	return nil
}
