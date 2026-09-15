package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	platformservices "github.com/wso2/integration-platform-tools/pkg/api/platform-services"
)

const (
	DatabaseServerStatusActive = "ACTIVE"
)

func getDatabaseServer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve database server."), nil
	}

	databaseServerId := utils.GetOptionalStringArgument(request, "database_server_uuid", "")
	databaseServerName := utils.GetOptionalStringArgument(request, "database_server_name", "")
	if databaseServerId == "" {
		// Get database servers if 'database_uuid' argument is not provided
		databaseServers, err := utils.GetPlatformServicesClient(ctx).GetDatabaseServers(
			targetOrg.ID,
			targetOrg.UUID,
		)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to retrieve database servers."), nil
		}

		// Return the list of database servers in the error response prompting the user to select one
		if len(databaseServers) == 0 {
			return utils.NewMCPErrorResponse(fmt.Errorf("no database servers found for the organization"), "Failed to retrieve database server."), nil
		}

		databaseServerList := make(map[string]string, len(databaseServers))
		for _, dbServer := range databaseServers {
			databaseServerList[dbServer.Name] = dbServer.ID
		}
		if databaseServerName != "" && databaseServerList[databaseServerName] != "" {
			databaseServerId = databaseServerList[databaseServerName]
		} else {
			// Build message with database names and IDs
			var sb strings.Builder
			sb.WriteString("No database UUID provided. Available database servers:\n")
			for name, id := range databaseServerList {
				sb.WriteString(name + " (ID: " + id + ")\n")
			}
			return utils.NewMCPErrorResponse(fmt.Errorf("%s", sb.String()), "Failed to retrieve database server."), nil
		}
	}

	databaseServer, err := utils.GetPlatformServicesClient(ctx).GetDatabaseServer(
		targetOrg.ID,
		databaseServerId,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve database server."), nil
	}

	nextSteps := []string{fmt.Sprintf("If the database status is 'ACTIVE' and it is not yet published, use `publish_default_database` with `database_server_uuid`: '%s'.", databaseServer.ID)}
	return utils.NewMCPResponse(databaseServer, "Database server retrieved successfully.", nextSteps)
}

func createDatabaseServer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	databaseType, err := utils.GetRequiredStringArgument(request, "database_type")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database server."), nil
	}

	databaseName, err := utils.GetRequiredStringArgument(request, "name")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database server."), nil
	}

	cloudProvider, err := utils.GetRequiredStringArgument(request, "cloud_provider")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database server."), nil
	}

	region, err := utils.GetRequiredStringArgument(request, "region")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database server."), nil
	}

	servicePlans, err := utils.GetPlatformServicesClient(ctx).GetDatabaseServicePlans(
		targetOrg.ID,
		databaseType,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database server."), nil
	}

	servicePlan, err := utils.GetRequiredStringArgument(request, "service_plan")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database server."), nil
	}

	// Get the service plan ID based on the provided service plan name. Ignore case sensitivity when comparing names.
	var servicePlanId string
	for _, plan := range servicePlans {
		if strings.EqualFold(plan.Name, servicePlan) {
			servicePlanId = plan.ID
			break
		}
	}

	createDatabaseServerReq := &platformservices.CreateDatabaseServerRequest{
		Name:            databaseName,
		Region:          region,
		CloudProvider:   cloudProvider,
		ServicePlanId:   servicePlanId,
		IsVectorEnabled: false,
	}

	databaseServer, err := utils.GetPlatformServicesClient(ctx).CreateDatabaseServer(
		targetOrg.ID,
		createDatabaseServerReq,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create database server."), nil
	}

	nextSteps := []string{
		fmt.Sprintf("Database server provisioning has started. To check its status, periodically call `get_database_server` with `database_server_uuid`: '%s' until the status is 'ACTIVE'.", databaseServer.ID),
		fmt.Sprintf("Once the database server is 'ACTIVE', you must publish it to the marketplace using `publish_default_database` with `database_server_uuid`: '%s' to make it connectable.", databaseServer.ID),
		"Key information: the database requires the connection to be secure, insecure connections are not allowed.",
	}
	return utils.NewMCPResponse(databaseServer, "Database server creation initiated successfully.", nextSteps)
}

func publishDefaultDatabase(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	databaseServerId, err := utils.GetRequiredStringArgument(request, "database_server_uuid")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to publish database."), nil
	}

	// Get the database server to ensure it exists and is in ACTIVE state.
	databaseServer, err := utils.GetPlatformServicesClient(ctx).GetDatabaseServer(
		targetOrg.ID,
		databaseServerId,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to publish database."), nil
	}
	if databaseServer.Status != DatabaseServerStatusActive {
		return utils.NewMCPErrorResponse(fmt.Errorf("database server is not in ACTIVE state. Please ensure the database server is active before publishing."), "Failed to publish database."), nil
	}

	// TODO: Create databases per environment and publish them to marketplace.
	databases, err := utils.GetPlatformServicesClient(ctx).GetDBInstances(
		targetOrg.ID,
		databaseServerId,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to publish database."), nil
	}
	if len(databases) == 0 {
		return utils.NewMCPErrorResponse(fmt.Errorf("no databases found for the given database server"), "Failed to publish database."), nil
	}

	database := databases[0] // Assuming we want to publish the first database

	// Check if credentials are already created for the database.
	credentials, err := utils.GetPlatformServicesClient(ctx).GetDBCredentials(
		targetOrg.ID,
		databaseServerId,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to publish database."), nil
	}

	if len(credentials) == 0 {
		// Get environment templates of the organization.
		environmentTemplates, err := utils.GetDevopsClient(ctx).GetEnvironmentTemplates(
			targetOrg.ID,
		)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to publish database."), nil
		}
		applicableEnvironments := make([]string, 0, len(*environmentTemplates))
		for _, template := range *environmentTemplates {
			applicableEnvironments = append(applicableEnvironments, template.ID)
		}

		// Create credentials for the database
		createDBCredRequest := &platformservices.CreateDBCredRequest{
			Database:               database.Name,
			DisplayName:            database.Name + " Credentials",
			IsSuperAdmin:           true,
			ApplicableEnvironments: applicableEnvironments,
		}

		_, err = utils.GetPlatformServicesClient(ctx).CreateDBCredentials(
			targetOrg.ID,
			databaseServerId,
			createDBCredRequest,
		)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to publish database."), nil
		}
	}

	if !database.DisplayOnMarketplace {
		publishDatabaseRequest := &platformservices.PublishDatabaseRequest{
			DisplayOnMarketplace: true,
			Name:                 database.Name,
			Status:               database.Status,
		}

		_, err := utils.GetPlatformServicesClient(ctx).PublishDatabase(
			targetOrg.ID,
			databaseServerId,
			database.Name,
			publishDatabaseRequest,
		)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to publish database."), nil
		}
	}

	response := map[string]any{
		"database": database,
	}

	nextSteps := []string{"The database has been published to the marketplace. It is now discoverable and can be connected to a component using `create_database_connection`."}
	return utils.NewMCPResponse(response, "Database published successfully.", nextSteps)
}
