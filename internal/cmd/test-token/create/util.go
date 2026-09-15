package create

import (
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	apimpublisher "github.com/wso2/integration-platform-tools/pkg/api/apimPublisher"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func getEndpointForEnv(orgId string, comp models.Component, selectedEnv project.ProjectEnvironment, deploymentTrack models.DeploymentTrack, endpointFlag string) (*project.Endpoint, error) {
	endpointsSpinner := utils.CreateSpinner(i18n.T(" Fetching endpoints..."), "")
	endpointsSpinner.Start()
	endpoints, err := auth.ProjectClient.GetComponentEndpoints(comp.Id, deploymentTrack.Id, orgId)
	endpointsSpinner.Stop()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to get component endpoints: %w"), err)
	}

	var selectedEndpoints []project.Endpoint
	for _, endpoint := range *endpoints {
		if endpoint.EnvironmentID == selectedEnv.ID {
			if endpoint.Visibility == component.ComponentServiceVisibilityPublic {
				selectedEndpoints = append(selectedEndpoints, endpoint)
			}
		}
	}

	if endpointFlag != "" {
		for _, item := range selectedEndpoints {
			if item.DisplayName == endpointFlag {
				return &item, nil
			}
		}
		return nil, fmt.Errorf("%s", i18n.T("Invalid flag for endpoint name"))
	}

	if len(selectedEndpoints) == 0 {
		return nil, fmt.Errorf("%s", i18n.T("No public endpoints available for selected environment"))
	} else if len(selectedEndpoints) == 1 {
		return &selectedEndpoints[0], nil
	} else {
		var selectedEndpointName string
		var endpointNames []string

		for _, item := range selectedEndpoints {
			endpointNames = append(endpointNames, item.DisplayName)
		}

		err := prompt.NewPromptSelectMessage[string](
			prompt.PromptSelectOpts[string]{Message: i18n.T("Endpoint: "), Values: endpointNames},
			&selectedEndpointName,
		).Prompt()

		if err != nil {
			return nil, err
		}

		for _, item := range selectedEndpoints {
			if item.DisplayName == selectedEndpointName {
				return &item, nil
			}
		}
	}
	return nil, fmt.Errorf("%s", i18n.T("Failed to select endpoint"))
}

func GetApiKeyForEndpoint(
	apiId string,
	selectedOrg *api.Organization,
	selectedEnv project.ProjectEnvironment) (*apimpublisher.ApiKeyResponse, error) {

	apiKeySpinner := utils.CreateSpinner(i18n.T(" Generating API key"), "\n")
	apiKeySpinner.Start()
	apiKey, err := auth.ApimPublisherClient.GetApiKey(
		apiId,
		selectedOrg.UUID,
		selectedOrg.ID,
		selectedEnv.Name,
	)
	apiKeySpinner.Stop()

	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to get api key: %w"), err)
	}

	return apiKey, nil
}

func getDeploymentInfoForProxy(orgID, orgHandle, orgUUID, cmpId, apiVId, envID string) (info deploymentbuild.ProxyDeployment, err error) {
	endpointsSpinner := utils.CreateSpinner(i18n.T(" Fetching endpoints..."), "")
	endpointsSpinner.Start()
	depInfo, err := auth.DeploymentBuildClient.GetProxyDeploymentInfo(
		orgID,
		orgHandle,
		orgUUID,
		cmpId,
		apiVId,
		envID,
	)
	endpointsSpinner.Stop()

	return *depInfo, err
}
