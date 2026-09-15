package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/pkg/api"
	pkgcomponent "github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/api/marketplace"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	platformservices "github.com/wso2/integration-platform-tools/pkg/api/platform-services"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func GetTargetOrgByUuid(ctx context.Context, request mcp.CallToolRequest) (*api.Organization, error) {
	orgUuid := GetArgument(request, "org_uuid")

	if EnsureAuthenticated(ctx) != nil {
		return nil, fmt.Errorf("cannot get target org as user is not authenticated")
	}

	orgClient := GetOrgClient(ctx)

	// No org UUID provided — derive from token (HTTP) or active session (STDIO).
	if orgUuid == nil {
		if IsHttpMode(ctx) {
			token, err := TokenFromContext(ctx)
			if err != nil || token == "" {
				return nil, fmt.Errorf("no valid token found in context")
			}
			orgUUIDInToken, err := ExtractOrgFromToken(token)
			if err != nil {
				return nil, fmt.Errorf("failed to get org uuid from token: %w", err)
			}
			return orgClient.GetOrganization(orgUUIDInToken)
		}
		// STDIO: GetOrganizations uses GetTokenForActiveOrg so no org ID needed.
		if initialOrgID := os.Getenv("CLOUD_INITIAL_ORG_ID"); initialOrgID != "" {
			return orgClient.GetOrganization(initialOrgID)
		}
		return auth.GetSelectedOrganization()
	}

	// Org UUID provided — look it up.
	if orgUuidStr, ok := orgUuid.(string); ok && orgUuidStr != "" {
		return orgClient.GetOrganization(orgUuidStr)
	}

	return nil, fmt.Errorf("invalid org uuid provided")
}

func GetTargetProject(ctx context.Context, targetOrg api.Organization, request mcp.CallToolRequest) (*models.Project, error) {
	projectUuid := GetArgument(request, "project_uuid")
	if projectUuid != nil {
		projectUuid = projectUuid.(string)
	} else if initial := os.Getenv("CLOUD_INITIAL_PROJECT_ID"); initial != "" {
		projectUuid = initial
	} else {
		return nil, fmt.Errorf("project_uuid is required")
	}
	if projectUuid == "" {
		return nil, fmt.Errorf("project_uuid cannot be empty")
	}

	projectClient := GetProjectClient(ctx)
	allProjects, err := projectClient.GetProjectsByOrgID(targetOrg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get retrieve projects: %w", err)
	}

	var selectedProject *models.Project
	for _, project := range allProjects {
		if project.ID == projectUuid {
			selectedProject = &project
			break
		}
	}
	if selectedProject == nil {
		return nil, fmt.Errorf("matching project not found")
	}

	return selectedProject, nil
}

func GetTargetProjectByUUID(ctx context.Context, targetOrg api.Organization, projectUUID string) (*models.Project, error) {
	if projectUUID == "" {
		return nil, fmt.Errorf("project_uuid cannot be empty")
	}
	projectClient := GetProjectClient(ctx)
	allProjects, err := projectClient.GetProjectsByOrgID(targetOrg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get retrieve projects: %w", err)
	}

	var selectedProject *models.Project
	for _, project := range allProjects {
		if project.ID == projectUUID {
			selectedProject = &project
			break
		}
	}
	if selectedProject == nil {
		return nil, fmt.Errorf("matching project not found")
	}

	return selectedProject, nil
}

func GetTargetComponent(ctx context.Context, targetOrg api.Organization, targetProject models.Project, request mcp.CallToolRequest) (*models.Component, error) {
	componentUuid := GetArgument(request, "integration_uuid")
	if componentUuid != nil {
		componentUuid = componentUuid.(string)
	} else {
		return nil, fmt.Errorf("integration_uuid is required")
	}
	if componentUuid == "" {
		return nil, fmt.Errorf("integration_uuid cannot be empty")
	}

	return GetTargetComponentByUUID(ctx, targetOrg, targetProject, componentUuid.(string))
}

func GetTargetComponentByUUID(ctx context.Context, targetOrg api.Organization, targetProject models.Project, componentUUID string) (*models.Component, error) {
	if componentUUID == "" {
		return nil, fmt.Errorf("integration_uuid cannot be empty")
	}

	componentClient := GetComponentClient(ctx)
	components, err := componentClient.GetAllComponents(targetOrg.Handle, targetOrg.ID, targetProject.ID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get retrieve components: %w", err)
	}

	var selectedComponent *models.Component
	for _, component := range components {
		if component.Id == componentUUID && pkgcomponent.IsIntegrationDisplayType(component.DisplayType) {
			selectedComponent = &component
			break
		}
	}
	if selectedComponent == nil {
		return nil, fmt.Errorf("matching component not found")
	}

	component, err := componentClient.GetComponentInfo(targetOrg.ID, selectedComponent.Handler, targetProject.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get retrieve component: %w", err)
	}

	return component, nil
}

func GetTargetEnv(ctx context.Context, targetOrg api.Organization, targetProject models.Project, request mcp.CallToolRequest) (*project.ProjectEnvironment, error) {
	selectedEnvironmentUuid := GetArgument(request, "environment_uuid")
	if selectedEnvironmentUuid != nil {
		selectedEnvironmentUuid = selectedEnvironmentUuid.(string)
	} else {
		return nil, fmt.Errorf("environment_uuid is required")
	}
	if selectedEnvironmentUuid == "" {
		return nil, fmt.Errorf("environment_uuid cannot be empty")
	}

	projectClient := GetProjectClient(ctx)
	environments, err := projectClient.GetProjectEnvironments(targetOrg.UUID, targetOrg.ID, targetProject.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve environments: %w", err)
	}
	var selectedEnvironment *project.ProjectEnvironment
	for _, environment := range *environments {
		if environment.ID == selectedEnvironmentUuid || environment.TemplateId == selectedEnvironmentUuid {
			selectedEnvironment = &environment
			break
		}
	}
	if selectedEnvironment == nil {
		return nil, fmt.Errorf("matching environment not found")
	}

	return selectedEnvironment, nil
}

func GetDeploymentTrack(selectedComponent models.Component, request mcp.CallToolRequest) (*models.DeploymentTrack, error) {
	deploymentTrackId := ""
	if GetArgument(request, "component_deployment_track_id") != nil {
		deploymentTrackId = GetArgument(request, "component_deployment_track_id").(string)
	}

	if deploymentTrackId != "" {
		for _, dtrack := range selectedComponent.DeploymentTracks {
			if dtrack.Id == deploymentTrackId {
				return &dtrack, nil
			}
		}
	}

	for _, dtrack := range selectedComponent.DeploymentTracks {
		if dtrack.Latest {
			return &dtrack, nil
		}
	}
	return nil, fmt.Errorf("failed to get latest deployment track")

}

func GetLatestApiVersion(selectedComponent models.Component) (*models.ApiVersion, error) {
	var latestVersion *models.ApiVersion
	for _, versionItem := range selectedComponent.ApiVersions {
		if versionItem.Latest {
			latestVersion = &versionItem
			break
		}
	}
	if latestVersion == nil {
		return nil, fmt.Errorf("failed to get latest api version")
	}

	return latestVersion, nil
}

func CreateJSONResponse(data any) (*mcp.CallToolResult, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to marshal data to JSON", err), nil
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

func GetRequiredStringArgument(request mcp.CallToolRequest, argumentName string) (string, error) {
	argumentValue := GetArgument(request, argumentName)
	if argumentValue == nil {
		return "", fmt.Errorf("argument %s is required", argumentName)
	}
	if argumentValue == "" {
		return "", fmt.Errorf("argument %s cannot be empty", argumentName)
	}
	return argumentValue.(string), nil
}

func GetOptionalStringArgument(request mcp.CallToolRequest, argumentName string, defaultValue string) string {
	argumentValue := GetArgument(request, argumentName)
	if argumentValue == nil {
		return defaultValue
	}
	if argumentValue == "" {
		return defaultValue
	}
	return argumentValue.(string)
}

func GetOptionalBoolArgument(request mcp.CallToolRequest, argumentName string, defaultValue bool) bool {
	argumentValue := GetArgument(request, argumentName)
	if argumentValue == nil {
		return defaultValue
	}
	value, ok := argumentValue.(bool)
	if !ok {
		return defaultValue
	}
	return value
}

func GetOptionalNumberArgument(request mcp.CallToolRequest, argumentName string, defaultValue int) int {
	argumentValue := GetArgument(request, argumentName)
	if argumentValue == nil {
		return defaultValue
	}
	value, ok := argumentValue.(int)
	if !ok {
		return defaultValue
	}
	return value
}

// GetDatabaseByResourceId returns the database resource with the given resourceId from the provided slice.
func GetDatabaseByResourceId(databases []marketplace.MarketplaceDatabaseResource, resourceId string) *marketplace.MarketplaceDatabaseResource {
	for i := range databases {
		if databases[i].ResourceId == resourceId {
			return &databases[i]
		}
	}
	return nil
}

func GetEnvironmentToDatabaseMapping(envTemplates []devops.EnvironmentTemplate,
	databaseResourceId string, databaseName string, databaseCredentials []platformservices.DatabaseCredential) map[string]connections.DatabaseConnectionEnvDetail {
	envToDatabaseMapping := make(map[string]connections.DatabaseConnectionEnvDetail)
	for _, credential := range databaseCredentials {
		if credential.DatabaseName == databaseName {
			for _, envTemplate := range envTemplates {
				if slices.Contains(credential.ApplicableEnvironments, envTemplate.ID) {
					envToDatabaseMapping[envTemplate.ID] = connections.DatabaseConnectionEnvDetail{
						ResourceId:         databaseResourceId,
						ParameterReference: credential.ID,
					}
				}
			}
		}
	}
	return envToDatabaseMapping
}

func GetBuildPackLanguage(component models.Component) string {
	buildPackConfig := component.Repository.BuildPackConfig
	if len(buildPackConfig) == 0 {
		return ""
	}
	return buildPackConfig[0].Buildpack.Language
}

func GetOptionalBooleanArgument(request mcp.CallToolRequest, argumentName string, defaultValue bool) bool {
	argumentValue := GetArgument(request, argumentName)
	if argumentValue == nil {
		return defaultValue
	}
	return argumentValue.(bool)
}

func GetArgument(request mcp.CallToolRequest, key string) interface{} {
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		return args[key]
	}
	return nil
}

// EnsureAuthenticated checks authentication for stdio mode and HTTP mode
// Returns nil if authenticated
// If not authenticated, returns error response in HTTP mode or request to use login tool in stdio mode.
func EnsureAuthenticated(ctx context.Context) *mcp.CallToolResult {
	if IsHttpMode(ctx) {
		// In HTTP mode, check if token is present in context
		token, err := TokenFromContext(ctx)
		if err == nil && token != "" {
			return nil
		}
		return mcp.NewToolResultError("Error: You are not logged in. Please send a valid token to access the mcp server.") // if not logged in, return error
	}
	// In stdio mode, check global authentication
	if auth.IsLoggedIn() {
		return nil
	}
	return mcp.NewToolResultError("Error: You are not logged in. Please use the 'login' tool to authenticate.") // if not logged in, return request to use login tool
}

func GetEnvironmentById(ctx context.Context, organizationUuid string, organizationId string,
	projectId string, environmentId string) (*project.ProjectEnvironment, error) {
	projectClient := GetProjectClient(ctx)
	environments, err := projectClient.GetProjectEnvironments(organizationUuid, organizationId, projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve environments: %w", err)
	}
	for _, env := range *environments {
		if env.ID == environmentId {
			return &env, nil
		}
	}
	return nil, fmt.Errorf("environment with ID %s not found in project %s", environmentId, projectId)
}

func GetPlatformHostname(componentType string) (string, error) {
	switch componentType {
	case "api":
		return auth.GetEnvConfig().PlatformHostnames.API, nil
	case "webapp":
		return auth.GetEnvConfig().PlatformHostnames.WebApp, nil
	default:
		return "", fmt.Errorf("unknown component type: %s", componentType)
	}
}

func GetChoreConsoleBaseUrl() string {
	return auth.GetEnvConfig().DevantConfig.ConsoleUrls.BaseUrl
}

func GetGhAppInstallURL(ctx context.Context) string {
	return auth.GetEnvConfig().GhApp.InstallUrl
}
