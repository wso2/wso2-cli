package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
	"github.com/wso2/integration-platform-tools/pkg/api/marketplace"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func CreateConnection(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create connection."), nil
	}

	sourceComponentType, err := utils.GetRequiredStringArgument(request, "source_integration_type")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create connection."), nil
	}

	// Try to get project_uuid from request, otherwise infer from integration_uuid
	projectUuid := ""
	if utils.GetArgument(request, "project_uuid") == nil {
		componentUuidVal := utils.GetArgument(request, "integration_uuid")
		if componentUuidVal == nil || componentUuidVal == "" {
			return utils.NewMCPErrorResponse(fmt.Errorf("integration_uuid is required to infer project_uuid"), "Failed to create connection."), nil
		}
		componentUuid := componentUuidVal.(string)
		// Fetch all projects for the org
		allProjects, err := utils.GetProjectClient(ctx).GetProjectsByOrgID(targetOrg.ID)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to create connection."), nil
		}
		found := false
		for _, project := range allProjects {
			components, err := utils.GetProjectClient(ctx).GetProjectComponents(targetOrg.ID, targetOrg.Handle, project.ID)
			if err != nil {
				continue
			}
			for _, cmp := range *components {
				if cmp.Id == componentUuid {
					projectUuid = project.ID
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return utils.NewMCPErrorResponse(fmt.Errorf("component not found in any project"), "Failed to create connection."), nil
		}
		// Set it in the request for downstream utils
		request.Params.Arguments.(map[string]interface{})["project_uuid"] = projectUuid
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create connection."), nil
	}

	selectedComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create connection."), nil
	}

	serviceId, err := utils.GetRequiredStringArgument(request, "marketplace_service_id")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create connection."), nil
	}

	name := utils.GetOptionalStringArgument(request, "name", "")
	description := utils.GetOptionalStringArgument(request, "description", "")
	schemaNameArg := utils.GetOptionalStringArgument(request, "schema_name", "")

	// Get service details (including schemas) from the marketplace
	serviceDetails, err := utils.GetMarketplaceClient(ctx).GetMarketplaceServiceDetails(targetOrg.ID, serviceId, "public")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create connection."), nil
	}

	if len(serviceDetails.ConnectionSchemas) == 0 {
		return utils.NewMCPErrorResponse(fmt.Errorf("no connection schemas available for the selected service"), "Failed to create connection."), nil
	}

	schemaReference := ""
	selectedSchema := serviceDetails.ConnectionSchemas[0]
	if schemaNameArg != "" {
		found := false
		for _, schema := range serviceDetails.ConnectionSchemas {
			if schema.Name == schemaNameArg {
				schemaReference = schema.ID
				selectedSchema = schema
				found = true
				break
			}
		}
		if !found {
			availableSchemas := []string{}
			for _, schema := range serviceDetails.ConnectionSchemas {
				availableSchemas = append(availableSchemas, schema.Name)
			}
			return utils.NewMCPErrorResponse(fmt.Errorf("Schema '%s' not found. Available schemas: %v", schemaNameArg, availableSchemas), "Failed to create connection."), nil
		}
	} else {
		schemaReference = serviceDetails.ConnectionSchemas[0].ID
	}

	// Get all environments for the project
	envsPtr, err := utils.GetProjectClient(ctx).GetProjectEnvironments(targetOrg.UUID, targetOrg.ID, selectedProject.ID)
	if err != nil || envsPtr == nil || len(*envsPtr) == 0 {
		return utils.NewMCPErrorResponse(fmt.Errorf("No environments found for the project."), "Failed to create connection."), nil
	}
	envs := *envsPtr

	var conn *connections.Connection

	// Check if it's a third-party service and use the appropriate method
	if serviceDetails.IsThirdParty {
		// Create third-party connection with provided configuration values
		thirdPartyConn, err := createThirdPartyConnection(ctx, targetOrg, selectedProject, selectedComponent, &serviceDetails, selectedSchema, envs, name, description, schemaReference, sourceComponentType)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to create third-party connection."), nil
		}
		conn = thirdPartyConn
	} else {
		// Create regular platform connection
		regularConn, err := utils.GetConnectionsClient(ctx).CreateNewConnection(
			targetOrg.ID,
			targetOrg.UUID,
			selectedProject.ID,
			name,
			serviceId,
			schemaReference,
			"PUBLIC",
			envs,
			selectedComponent.Id,
			false, // generateCreds=false
		)
		if err != nil {
			return utils.NewMCPErrorResponse(fmt.Errorf("Failed to create connection: %v", err), "Failed to create connection."), nil
		}
		conn = &regularConn
	}

	// Extract ServiceURL for all environments
	serviceURLs := map[string]string{}
	for envId, config := range conn.Configurations {
		if entry, ok := config.Entries["ServiceURL"]; ok {
			urlVal := entry.Value
			if sourceComponentType == "web_app" {
				// Replace the host with /choreo-apis and keep the path
				// Example: https://host/bookmanager/book-manager-service/v1 -> /choreo-apis/bookmanager/book-manager-service/v1
				if strings.HasPrefix(urlVal, "http://") || strings.HasPrefix(urlVal, "https://") {
					u, err := url.Parse(urlVal)
					if err == nil {
						urlVal = "/choreo-apis" + u.Path
					}
				}
			}
			serviceURLs[envId] = urlVal
		}
	}

	// Get buildpack language for the component
	buildPack, _ := utils.GetTargetBuildPack(ctx, *targetOrg, "service", request)
	buildPackLanguage := ""
	if buildPack != nil {
		buildPackLanguage = buildPack.Language
	}

	componentTypeForGuide := "SERVICE"
	if sourceComponentType == "web_app" {
		componentTypeForGuide = "WEB_APP_SPA"
	}

	audience := "console"
	if sourceComponentType == "web_app" {
		audience = "mcp"
	}

	// Fetch the service connection guide using the new method
	guide, err := utils.GetMarketplaceClient(ctx).GetServiceConnectionHowToUse(
		targetOrg.ID,
		targetOrg.UUID,
		conn.ServiceId,
		conn.SchemaReference,
		conn.GroupUuid,
		componentTypeForGuide,
		buildPackLanguage,
		"component_v11",
		conn.Name,
		audience,
		false,
	)
	if err != nil {
		guide = "Failed to fetch guide: " + err.Error()
	}

	result := map[string]interface{}{
		"serviceURLs": serviceURLs,
		"connection":  conn,
		"guide":       guide,
	}
	connectionName := conn.Name
	connectionNameEnvVar := strings.ToUpper(strings.ReplaceAll(connectionName, "-", "_")) + "_SERVICE_URL"
	connectionURL := ""
	if len(conn.Configurations) > 0 {
		// Assuming there's at least one environment and ServiceURL is present
		for _, config := range conn.Configurations {
			if entry, ok := config.Entries["ServiceURL"]; ok {
				connectionURL = entry.Value
				break
			}
		}
	}

	var nextSteps []string
	if sourceComponentType == "web_app" {
		nextSteps = append(nextSteps, "Update the frontend code to read in the connection details from /app/public/config.js `<script src=\"/public/config.js\"></script>` as per the guide.")
		nextSteps = append(nextSteps, "IMPORTANT: In your code, implement sign-in and sign-out buttons that redirect to the relative paths '/auth/login' and `/auth/logout?session_hint=${Cookies.get('session_hint')}`. From those endpoints, the platform's managed service will handle the authentication flow automatically. Tip: If the frontend does not have authentication, make a GET request to /auth/userinfo. If it returns 401, direct the user to /auth/login; if it returns 200, show the logout button.")
		nextSteps = append(nextSteps, "IMPORTANT: Update your API calls to use the format `window.configs.apiUrl + \"/your-backend-path\"`. You must configure the request to include credentials so the secure session cookie is sent automatically.")
		nextSteps = append(nextSteps, "IMPORTANT: Code changes to the frontend should be EXTREMELY minimal but above mentioned changes are required. So plan the least disruptive way to implement these changes.")
		nextSteps = append(nextSteps, "Commit and push the changes to the repository.")
		nextSteps = append(nextSteps, "Call create_build and create_deployment to create a new build and deployment for the component.")
		nextSteps = append(nextSteps, fmt.Sprintf("Create new configmap with `create_configurations` with window.config.apiUrl as the value of the %s. This config needs to be mounted to the webapp at '/app/public/config.js'. Refer the guide for the more details.", serviceURLs))
	} else {
		nextSteps = append(nextSteps, "Connection created. A new build is required to inject the connection details into the component.")
		nextSteps = append(nextSteps, fmt.Sprintf("Use `create_build` for integration_uuid: '%s' and then deploy the new build.", selectedComponent.Id))
		nextSteps = append(nextSteps, "To connect to this service, add the following to your component.yaml file under the `connections` section:")
		nextSteps = append(nextSteps, fmt.Sprintf(
			"```yaml\n- type: service\n  id: %s\n  name: %s\n  parameters:\n    - name: %s\n      value: %s\n```",
			connectionName, connectionName, connectionNameEnvVar, connectionURL,
		))
		nextSteps = append(nextSteps, "Then, in your code, you can access the connection parameters as environment variables. For example, in Go:")
		nextSteps = append(nextSteps, fmt.Sprintf(
			"```go\nserviceURL := os.Getenv(\"%s\")\n// Use serviceURL to make API calls to the connected service\n```",
			connectionNameEnvVar,
		))
		nextSteps = append(nextSteps, "For more details, refer to the platform documentation on service connections.")
	}

	return utils.NewMCPResponse(result, "Service connection created successfully.", nextSteps)
}

func createThirdPartyConnection(
	ctx context.Context,
	targetOrg *api.Organization,
	selectedProject *models.Project,
	selectedComponent *models.Component,
	serviceDetails *marketplace.MarketplaceService,
	selectedSchema marketplace.MarketplaceServiceScheme,
	envs []project.ProjectEnvironment,
	name string,
	description string,
	schemaReference string,
	sourceComponentType string,
) (*connections.Connection, error) {

	// Construct configurations for each environment
	configurations := make(map[string]connections.ThirdPartyConnectionConfig)
	envMapping := make(map[string]connections.ThirdPartyConnectionEnvMapping)

	// Extract the parameter reference from the endpointRefs map.
	var parameterReference string
	if serviceDetails.EndpointRefs != nil && len(serviceDetails.EndpointRefs) > 0 {
		for key := range serviceDetails.EndpointRefs {
			parameterReference = key
			break // We only need the first one.
		}
	} else {
		return nil, fmt.Errorf("marketplace service response for third-party service '%s' is missing the required 'endpointRefs' field", serviceDetails.Name)
	}

	for _, env := range envs {
		// Build entries from schema entries
		entries := make(map[string]connections.ThirdPartyConnectionEntry)
		for _, schemaEntry := range selectedSchema.Entries {
			entries[schemaEntry.Name] = connections.ThirdPartyConnectionEntry{
				Key:         schemaEntry.Name,
				Value:       "",
				IsFile:      false,
				IsSensitive: schemaEntry.IsSensitive,
			}
		}

		configurations[env.TemplateId] = connections.ThirdPartyConnectionConfig{
			EnvironmentUuid: env.TemplateId,
			IsCritical:      env.Critical,
			Entries:         entries,
		}

		envMapping[env.TemplateId] = connections.ThirdPartyConnectionEnvMapping{
			ParameterReference: parameterReference,
			ResourceId:         env.TemplateId,
		}
	}

	// Determine component type and visibility type
	componentType := "service"
	visibilityComponentType := "buildpackService"

	// Create the request payload
	thirdPartyReq := connections.CreateThirdPartyConnectionReq{
		Configurations:  configurations,
		Name:            name,
		SchemaReference: schemaReference,
		Visibilities: []connections.ConnectionVisibility{
			{
				ComponentUuid:    selectedComponent.Id,
				OrganizationUUID: targetOrg.UUID,
				ProjectUUID:      selectedProject.ID,
				ComponentType:    visibilityComponentType,
			},
		},
		ServiceId:     serviceDetails.ServiceID,
		Description:   description,
		EnvMapping:    envMapping,
		ComponentType: componentType,
	}

	// Call the third-party connection API
	conn, err := utils.GetConnectionsClient(ctx).CreateThirdPartyConnection(targetOrg.ID, thirdPartyReq, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create third-party connection: %w", err)
	}

	return conn, nil
}
