package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/marketplace"
)

func listMarketplaceResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, err := utils.GetRequiredStringArgument(request, "resource_type")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to list marketplace resources."), nil
	}

	switch resourceType {
	case "platform-service":
		return listMarketplaceServices(ctx, request)
	case "database":
		return getDatabases(ctx, request)
	case "third-party":
		return listThirdPartyServices(ctx, request)
	default:
		return utils.NewMCPErrorResponse(fmt.Errorf("invalid resource_type. Possible values: platform-service, database, third-party"), "Failed to list marketplace resources."), nil
	}
}

func getDatabases(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 1. Check if the user is logged in
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	// 2. Determine the target organization from the request
	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve databases."), nil
	}

	// 3. Call the updated client method to get the full response
	marketplaceClient := utils.GetMarketplaceClient(ctx)
	response, err := marketplaceClient.GetDatabases(targetOrg.ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve databases."), nil
	}

	for i := range response.Data {
		for j := range response.Data[i].ConnectionSchemas {
			response.Data[i].ConnectionSchemas[j].Entries = []marketplace.ConnectionSchemaEntry{}
		}
	}

	// 4. Create a JSON response from the entire response object
	nextSteps := []string{
		"To connect an integration to one of these databases, use `create_database_connection` with the `service_id` of the desired database.",
		"Note - the integration's region and database region does not need to be the same. You can connect a database in a different region to an integration in a different region.",
	}
	return utils.NewMCPResponse(response, "Databases retrieved successfully.", nextSteps)
}

func listMarketplaceServices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to list marketplace services."), nil
	}

	query, err := utils.GetRequiredStringArgument(request, "query")
	if err != nil {
		return utils.NewMCPErrorResponse(fmt.Errorf("Please provide a name or keyword to search for services in the marketplace."), "Failed to list marketplace services."), nil
	}

	limitStr := utils.GetOptionalStringArgument(request, "limit", "8")
	limit := int64(8)
	if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
		limit = l
	}
	offsetStr := utils.GetOptionalStringArgument(request, "offset", "0")
	offset := int64(0)
	if o, err := strconv.ParseInt(offsetStr, 10, 64); err == nil {
		offset = o
	}
	networkVisibilityFilter := utils.GetOptionalStringArgument(request, "networkVisibilityFilter", "public,org,project")

	reqBody := map[string]any{
		"limit":                      limit,
		"offset":                     offset,
		"query":                      query,
		"networkVisibilityFilter":    networkVisibilityFilter,
		"networkVisibilityprojectId": "all",
		"sortBy":                     "createdTime",
		"sortAscending":              false,
		"searchContent":              false,
		"aggregateByMajorVersion":    true,
		"includeCreated":             false,
		"isThirdParty":               false,
	}

	// Convert map to MarketplaceGetServicesReq
	var servicesReq marketplace.MarketplaceGetServicesReq
	b, _ := json.Marshal(reqBody)
	_ = json.Unmarshal(b, &servicesReq)

	marketplaceClient := utils.GetMarketplaceClient(ctx)
	response, err := marketplaceClient.GetMarketplaceServices(targetOrg.ID, servicesReq)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to list marketplace services."), nil
	}

	nextSteps := []string{"To connect an integration to one of these services, use `create_connection` with the `marketplace_service_id` of the desired service."}
	return utils.NewMCPResponse(response, "Marketplace services listed successfully.", nextSteps)
}

func listThirdPartyServices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to list marketplace services."), nil
	}

	query := utils.GetOptionalStringArgument(request, "query", "")

	limitStr := utils.GetOptionalStringArgument(request, "limit", "8")
	limit := int64(8)
	if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
		limit = l
	}
	offsetStr := utils.GetOptionalStringArgument(request, "offset", "0")
	offset := int64(0)
	if o, err := strconv.ParseInt(offsetStr, 10, 64); err == nil {
		offset = o
	}
	networkVisibilityFilter := utils.GetOptionalStringArgument(request, "networkVisibilityFilter", "public,org,project")

	marketplaceClient := utils.GetMarketplaceClient(ctx)

	if resourceId := utils.GetOptionalStringArgument(request, "resource_id", ""); resourceId != "" {
		response, err := marketplaceClient.GetThirdPartyServiceOpenApi(targetOrg.ID, resourceId)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to list marketplace services."), nil
		}
		nextSteps := []string{
			"Study the OpenAPI spec to understand the endpoints and parameters.",
			"If you want the current application to integrate with this service, use environment variables to load the base URL and credentials, and implement the integration as needed.",
			"It is *discouraged* to create types for the entire OpenAPI specification, you should only create types for the specific endpoints you want to use.",
		}
		return utils.NewMCPResponse(response, "Third party service openapi retrieved successfully.", nextSteps)
	}

	reqBody := map[string]any{
		"limit":                      limit,
		"offset":                     offset,
		"query":                      query,
		"networkVisibilityFilter":    networkVisibilityFilter,
		"networkVisibilityprojectId": "all",
		"sortBy":                     "createdTime",
		"sortAscending":              false,
		"searchContent":              false,
		"aggregateByMajorVersion":    true,
		"includeCreated":             false,
		"isThirdParty":               true,
	}

	// Convert map to MarketplaceGetServicesReq
	var servicesReq marketplace.MarketplaceGetServicesReq
	b, _ := json.Marshal(reqBody)
	_ = json.Unmarshal(b, &servicesReq)

	response, err := marketplaceClient.GetMarketplaceServices(targetOrg.ID, servicesReq)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to list marketplace services."), nil
	}

	nextSteps := []string{
		"To get the implementation details of a third party service, use `list_marketplace_resources{resource_type: 'third-party', resource_id: 'RESOURCE_ID_OF_THE_THIRD_PARTY_SERVICE'}`",
		"To connect a component to one of these services, use `create_connection` with the `marketplace_service_id` of the desired service.",
	}
	return utils.NewMCPResponse(response, "Marketplace services listed successfully.", nextSteps)
}
