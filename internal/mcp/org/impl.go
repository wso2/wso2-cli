package org

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/internal/region"
)

func changeOrg(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to change organization."), nil
	}

	userInfo, err := auth.GetCurrentUser()
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to change organization."), nil
	}

	err = auth.SetSelectedOrg(targetOrg, userInfo.Organizations)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to change organization."), nil
	}
	nextSteps := []string{fmt.Sprintf("To see the projects in the new organization, use `get_projects`.")}
	return utils.NewMCPResponse(map[string]string{"org_uuid": targetOrg.UUID, "org_name": targetOrg.Name}, fmt.Sprintf("Active organization changed to %s", targetOrg.Name), nextSteps)
}

func getActiveOrg(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to get active organization."), nil
	}

	region := region.GetCurrentRegion()
	targetOrg.Region = &region
	nextSteps := []string{fmt.Sprintf("The active organization is '%s'. To see all projects within this organization, use the `get_projects` tool.", targetOrg.Name)}
	return utils.NewMCPResponse(targetOrg, "Active organization details retrieved.", nextSteps)
}

func getOrganizations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	orgs, err := utils.GetOrgClient(ctx).GetOrganizations()
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to get organizations."), nil
	}
	nextSteps := []string{"To switch to a different organization from this list, use `change_org` with the desired `org_uuid`."}
	return utils.NewMCPResponse(orgs, "Organizations retrieved successfully.", nextSteps)
}
