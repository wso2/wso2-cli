package database

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
)

// createDatabaseConnection creates a database connection for a given component.
// It uses the serviceId to get the database from the marketplace.
// It then creates a database connection with the given component UUID.
// It returns the database connection details.
func createDatabaseConnection(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}

	cmpName, err := utils.GetRequiredStringArgument(request, "component_name")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}

	serviceId, err := utils.GetRequiredStringArgument(request, "service_id")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}

	componentUuid, err := utils.GetRequiredStringArgument(request, "integration_uuid")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}

	component, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}

	description := utils.GetOptionalStringArgument(request, "description", "")

	// Get the databases from the marketplace and filter the database by the serviceId.
	databases, err := utils.GetMarketplaceClient(ctx).GetMarketplaceDatabases(targetOrg.ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}
	database := utils.GetDatabaseByResourceId(*databases, serviceId)
	if database == nil {
		return utils.NewMCPErrorResponse(fmt.Errorf("failed to get database by resource ID"), "Failed to create database connection."), nil
	}

	// Check if there are any connection schemas available
	if len(database.ConnectionSchemas) == 0 {
		return utils.NewMCPErrorResponse(fmt.Errorf("no connection schemas available for the selected database"), "Failed to create database connection."), nil
	}

	databaseCredentials, err := utils.GetPlatformServicesClient(ctx).GetDatabaseCredentials(targetOrg.ID, database.ResourceDetails.DatabaseServerId, database.Name)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}

	// Get the environment templates for the target organization.
	// This is used to map the environment variables to the database credentials.
	envTemplates, err := utils.GetDevopsClient(ctx).GetEnvironmentTemplates(targetOrg.ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}
	envToDatabaseMapping := utils.GetEnvironmentToDatabaseMapping(*envTemplates, database.ResourceId, database.Name, *databaseCredentials)
	connectionReq := connections.CreateDatabaseConnectionReq{
		Name:            fmt.Sprintf("connection-%s-%s", cmpName, database.Name),
		Description:     description,
		ServiceId:       serviceId,
		SchemaReference: database.ConnectionSchemas[0].Id,
		Visibilities: []connections.DatabaseConnectionVisibility{
			{
				ComponentUuid:    componentUuid,
				OrganizationUuid: targetOrg.UUID,
				ProjectUuid:      selectedProject.ID,
			},
		},
		EnvMapping: envToDatabaseMapping,
	}

	// Create the database connection.
	connection, err := utils.GetConnectionsClient(ctx).CreateDatabaseConnection(targetOrg.ID, connectionReq)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database connection."), nil
	}

	// Get the database connection guide.
	buildPackLanguage := utils.GetBuildPackLanguage(*component)
	connectionGuide, err := utils.GetMarketplaceClient(ctx).GetDatabaseConnectionGuide(targetOrg.ID, targetOrg.UUID, serviceId, connection.SchemaReference, connection.Name, buildPackLanguage)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to get database connection guide."), nil
	}

	result := map[string]interface{}{
		"connection": connection,
		"guide":      connectionGuide,
	}
	nextSteps := []string{fmt.Sprintf("Database connection created. To apply the settings, you must trigger a new build with `create_build` for integration_uuid: '%s' and then deploy it.", component.Id)}
	return utils.NewMCPResponse(result, "Database connection created successfully.", nextSteps)
}
