package testkey

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func generateTestKey(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
	}

	targetProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
	}

	selectedComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *targetProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
	}

	compInfoWithRepo, err := common.GetComponentWithRepoData(targetOrg.ID, selectedComponent.Handler, targetProject.ID)

	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
	}

	dTrack, err := utils.GetDeploymentTrack(*selectedComponent, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
	}

	selectedEnvironment, err := utils.GetTargetEnv(ctx, *targetOrg, *targetProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
	}

	if selectedEnvironment.Critical {
		return utils.NewMCPErrorResponse(fmt.Errorf("Cannot generate test keys for critical environments. Use the web console for generating test keys for critical environments."), "Failed to generate test key."), nil
	}

	if selectedComponent.DisplayType == component.DisplayTypeGitProxy {
		var apiV *models.ApiVersion

		for _, entry := range compInfoWithRepo.ApiVersions {
			if entry.Latest {
				apiV = &entry
				break
			}
		}

		deploymentBuildClient := utils.GetDeploymentBuildClient(ctx)
		selectedEndpoint, err := deploymentBuildClient.GetProxyDeploymentInfo(
			targetOrg.ID,
			targetOrg.Handle,
			targetOrg.UUID,
			selectedComponent.Id,
			apiV.Id,
			selectedEnvironment.ID,
		)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
		}

		apimPublisherClient := utils.GetApimPublisherClient(ctx)
		apiKey, err := apimPublisherClient.GetApiKey(
			selectedEndpoint.APIID,
			targetOrg.UUID,
			targetOrg.ID,
			selectedEnvironment.Name,
		)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
		}

		nextSteps := []string{"A test key has been generated. You can now invoke the service endpoint by including this key in the 'Test-Key' header of your API request."}
		return utils.NewMCPResponse(map[string]string{"test_key": apiKey.Apikey, "validity_time": fmt.Sprintf("%d", apiKey.ValidityTime)}, "Test key generated successfully.", nextSteps)
	} else {
		projectClient := utils.GetProjectClient(ctx)
		endpoints, err := projectClient.GetComponentEndpoints(selectedComponent.Id, dTrack.Id, targetOrg.ID)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
		}

		if len(*endpoints) == 0 {
			return utils.NewMCPErrorResponse(fmt.Errorf("No endpoints found for the selected component."), "Failed to generate test key."), nil
		}

		var selectedEndpoints []project.Endpoint
		for _, endpoint := range *endpoints {
			if endpoint.EnvironmentID == selectedEnvironment.ID {
				if endpoint.Visibility == component.ComponentServiceVisibilityPublic {
					selectedEndpoints = append(selectedEndpoints, endpoint)
				}
			}
		}

		if len(selectedEndpoints) == 0 {
			return utils.NewMCPErrorResponse(fmt.Errorf("No public endpoints available for the selected environment."), "Failed to generate test key."), nil
		}
		if len(selectedEndpoints) == 1 {
			selectedEndpoint := selectedEndpoints[0]
			apimPublisherClient := utils.GetApimPublisherClient(ctx)
			apiKey, err := apimPublisherClient.GetApiKey(
				selectedEndpoint.APIMID,
				targetOrg.UUID,
				targetOrg.ID,
				selectedEnvironment.Name,
			)
			if err != nil {
				return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
			}

			nextSteps := []string{"A test key has been generated. You can now invoke the service endpoint by including this key in the 'Test-Key' header of your API request."}
			return utils.NewMCPResponse(map[string]string{"test_key": apiKey.Apikey, "validity_time": fmt.Sprintf("%d", apiKey.ValidityTime)}, "Test key generated successfully.", nextSteps)

		}

		if len(selectedEndpoints) > 1 {
			selectedEndpointUUID := utils.GetOptionalStringArgument(request, "endpoint_uuid", "")
			if selectedEndpointUUID != "" {
				return utils.NewMCPErrorResponse(fmt.Errorf("Endpoint for given APIM ID not found. Following endpoints are available: %v", selectedEndpoints), "Failed to generate test key."), nil
			}
			var selectedEndpoint *project.Endpoint
			for _, endpoint := range selectedEndpoints {
				if endpoint.APIMID == selectedEndpointUUID {
					selectedEndpoint = &endpoint
					break
				}
			}
			if selectedEndpoint == nil {
				return utils.NewMCPErrorResponse(fmt.Errorf("Endpoint APIM ID needs to be specified when multiple endpoints are available. Following endpoints are available: %v", selectedEndpoints), "Failed to generate test key."), nil
			}
			apimPublisherClient := utils.GetApimPublisherClient(ctx)
			apiKey, err := apimPublisherClient.GetApiKey(
				selectedEndpoint.APIMID,
				targetOrg.UUID,
				targetOrg.ID,
				selectedEnvironment.Name,
			)
			if err != nil {
				return utils.NewMCPErrorResponse(err, "Failed to generate test key."), nil
			}
			nextSteps := []string{"A test key has been generated. You can now invoke the service endpoint by including this key in the 'Test-Key' header of your API request.", "Example curl command: curl -H 'Test-Key: " + apiKey.Apikey + "' <your_backend_public_endpoint_url>"}
			return utils.NewMCPResponse(map[string]string{"test_key": apiKey.Apikey, "validity_time": fmt.Sprintf("%d", apiKey.ValidityTime)}, "Test key generated successfully.", nextSteps)
		}

	}
	return utils.NewMCPErrorResponse(fmt.Errorf(
		"test keys apply only to integrations that expose an HTTP endpoint — the 'API', 'AI Agent' and 'MCP Server' subtypes. "+
			"This integration's type is %q. An 'Automation' is a scheduled task, so run it with execute_task instead; "+
			"'Event Integration' and 'File Integration' are event handlers with no endpoint to invoke",
		selectedComponent.DisplayType), "Failed to generate test key."), nil
}
