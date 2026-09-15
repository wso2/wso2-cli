package build

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
)

func getBuilds(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	waitForCompletion := utils.GetOptionalBooleanArgument(request, "wait_for_completion", false)

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve builds."), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve builds."), nil
	}

	component, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve builds."), nil
	}

	dTrack, err := utils.GetDeploymentTrack(*component, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve builds."), nil
	}

	builds, err := utils.GetDeploymentBuildClient(ctx).GetDeploymentBuilds(targetOrg.ID, component.Name, selectedProject.Handler, dTrack.Id)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve builds."), nil
	}

	if waitForCompletion {
		// Wait up to 10 minutes (600 * 10s) for the latest build to complete
		for range 60 {
			// Re-fetch builds
			builds, err = utils.GetDeploymentBuildClient(ctx).GetDeploymentBuilds(targetOrg.ID, component.Name, selectedProject.Handler, dTrack.Id)
			if err != nil {
				return utils.NewMCPErrorResponse(err, "Failed to retrieve builds during wait."), nil
			}
			if len(builds) == 0 {
				return utils.NewMCPErrorResponse(fmt.Errorf("no builds found to wait for completion"), "Failed to retrieve builds."), nil
			}
			// If the latest build is still in progress, wait and try again
			if builds[0].Status.Status == "in_progress" {
				time.Sleep(10 * time.Second)
			} else {
				// Build is no longer in progress, return the builds
				return utils.NewMCPResponse(builds, "Builds retrieved successfully.", nil)
			}
		}
		// Timed out after 10 minutes
		return utils.NewMCPErrorResponse(fmt.Errorf("waiting for build to complete timed out after 10 minutes"), "Build completion timed out."), nil
	}

	nextSteps := []string{
		"To deploy a completed build, use `create_deployment` with the `build_ref_or_image_id` and the `runId` from the desired build entry.",
		"To view the logs for a specific build, use `get_logs` with `log_type`: 'build' and the `build_status_run_id` from the build entry.",
	}
	return utils.NewMCPResponse(builds, "Builds retrieved successfully.", nextSteps)
}

func getCommitHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create build."), nil
	}

	componentUuid, err := utils.GetRequiredStringArgument(request, "integration_uuid")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to get commit history."), nil
	}

	branch, err := utils.GetRequiredStringArgument(request, "branch")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to get commit history."), nil
	}

	commitHistory, err := utils.GetGitClient(ctx).GetCommitHistory(componentUuid, branch, targetOrg.ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to get commit history."), nil
	}

	nextSteps := []string{"To start a new build using one of these commits, use `create_build` with the desired `commit_hash`."}
	return utils.NewMCPResponse(commitHistory, "Commit history retrieved successfully.", nextSteps)
}

func createBuild(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to get commit history."), nil
	}

	targetProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create build."), nil
	}

	targetComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *targetProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create build."), nil
	}

	dpTrack, err := utils.GetDeploymentTrack(*targetComponent, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create build."), nil
	}

	branch, err := utils.GetRequiredStringArgument(request, "branch")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create build."), nil
	}

	commitHash := utils.GetOptionalStringArgument(request, "commit_hash", "")

	if commitHash == "" {
		gitClient := utils.GetGitClient(ctx)
		commitHistory, err := gitClient.GetCommitHistory(targetComponent.Id, branch, targetOrg.ID)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to get commit history."), nil
		}
		for _, commit := range commitHistory {
			if commit.IsLatest {
				commitHash = commit.Sha
				break
			}
		}
	}

	if commitHash == "" {
		return utils.NewMCPErrorResponse(fmt.Errorf("unable to find the latest commit hash in the branch, use `git rev-parse origin/main` with current origin and branch name"), "Failed to create build."), nil
	}

	buildResponse, err := utils.GetDeploymentBuildClient(ctx).CreateDeploymentBuilds(targetOrg.ID, targetComponent.Name, targetProject.Handler, dpTrack.Id, commitHash)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create build."), nil
	}

	nextSteps := []string{
		fmt.Sprintf("A new build has been initiated. To monitor its progress until completion, use `get_builds` with `integration_uuid`: '%s', `project_uuid`: '%s', `component_deployment_track_id`: '%s', and `wait_for_completion`: true.", targetComponent.Id, targetProject.ID, dpTrack.Id),
		"Once the build is complete, you can deploy it using `create_deployment` with the `build_ref_or_image_id` and the `runId`from the completed build info.",
	}
	return utils.NewMCPResponse(buildResponse, "Build created successfully.", nextSteps)
}
