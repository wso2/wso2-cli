package project

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func getProjects(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve projects."), nil
	}

	projectUUID := utils.GetOptionalStringArgument(request, "project_uuid", "")

	if projectUUID != "" {
		project, err := utils.GetTargetProjectByUUID(ctx, *targetOrg, projectUUID)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to retrieve project."), nil
		}
		nextSteps := []string{
			fmt.Sprintf("To see the components within a specific project, use `get_components` with the `project_uuid`: '%s'.", project.ID),
			fmt.Sprintf("To see the configured environments for a project, use `get_project_environments` with the `project_uuid`: '%s'.", project.ID),
		}
		return utils.NewMCPResponse([]*models.Project{project}, "Project retrieved successfully.", nextSteps)
	}

	projectClient := utils.GetProjectClient(ctx)
	projects, err := projectClient.GetProjectsByOrgID(targetOrg.ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve projects."), nil
	}

	nextSteps := []string{
		"To see the components within a specific project, use `get_components` with the `project_uuid`.",
		"To see the configured environments for a project, use `get_project_environments` with the `project_uuid`.",
	}
	return utils.NewMCPResponse(projects, "Projects retrieved successfully.", nextSteps)
}

func getProjectEnvironments(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get target organization: %v", err)), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve project environments."), nil
	}

	projectClient := utils.GetProjectClient(ctx)
	environments, err := projectClient.GetProjectEnvironments(targetOrg.UUID, targetOrg.ID, selectedProject.ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve project environments."), nil
	}

	nextSteps := []string{
		"To deploy a component, use the `environment_uuid` for your target environment (e.g., Development) in the `create_deployment` tool.",
		"To manage test users for an environment, use `manage_test_users` with the `environment_template_id`.",
	}
	return utils.NewMCPResponse(environments, "Project environments retrieved successfully.", nextSteps)
}

func createProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	projectName, err := utils.GetRequiredStringArgument(request, "project_name")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create project."), nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get target organization: %v", err)), nil
	}

	repoUrl := utils.GetOptionalStringArgument(request, "repo_url", "")
	branch := utils.GetOptionalStringArgument(request, "branch", "main")
	region := utils.GetOptionalStringArgument(request, "region", "US")
	version := utils.GetOptionalStringArgument(request, "version", "1.0.0")
	description := utils.GetOptionalStringArgument(request, "description", "")

	orgIdInt, err := strconv.Atoi(targetOrg.ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create project."), nil
	}

	projectReq := project.GetProjectMutationRequest{
		OrgID:        orgIdInt,
		OrgHandler:   targetOrg.Handle,
		Name:         projectName,
		Description:  description,
		Version:      version,
		Region:       region,
		Repository:   repoUrl,
		Branch:       branch,
		CredentialId: "",
	}

	projectClient := utils.GetProjectClient(ctx)
	createdProject, err := projectClient.CreateProject(projectReq, targetOrg.ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create project."), nil
	}

	nextSteps := []string{fmt.Sprintf("Project created successfully. To deploy components, you first need the environment details. Use `get_project_environments` with `project_uuid`: '%s'.", createdProject.ID)}
	return utils.NewMCPResponse(createdProject, "Project created successfully.", nextSteps)
}
