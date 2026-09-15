package integration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/cmd/component/create"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	pkgcomponent "github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/util/repo"
)

func getIntegrations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve integrations."), nil
	}

	targetProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve integrations."), nil
	}

	integrationUUID := utils.GetOptionalStringArgument(request, "integration_uuid", "")

	if integrationUUID != "" {
		integration, err := utils.GetTargetComponentByUUID(ctx, *targetOrg, *targetProject, integrationUUID)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to retrieve integration."), nil
		}
		nextSteps := []string{
			fmt.Sprintf("To get more details about a specific integration, call `get_integrations` again with `integration_uuid`: '%s'.", integration.Id),
			fmt.Sprintf("To check the build history for an integration, use `get_builds` with its `integration_uuid`: '%s'.", integration.Id),
			fmt.Sprintf("To check the current deployment status for an integration, use `get_deployment` with its `integration_uuid`: '%s'.", integration.Id),
		}
		return utils.NewMCPResponse([]*models.Component{integration}, "Integration retrieved successfully.", nextSteps)
	}

	componentClient := utils.GetComponentClient(ctx)
	allComponents, err := componentClient.GetAllComponents(targetOrg.Handle, targetOrg.ID, targetProject.ID, false)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve integrations."), nil
	}

	integrations := make([]models.Component, 0, len(allComponents))
	for _, c := range allComponents {
		if pkgcomponent.IsIntegrationDisplayType(c.DisplayType) {
			integrations = append(integrations, c)
		}
	}

	nextSteps := []string{
		"To get more details about a specific integration, call `get_integrations` again with `integration_uuid`: '<INTEGRATION_UUID>'.",
		"To check the build history for an integration, use `get_builds` with its `integration_uuid`.",
		"To check the current deployment status for an integration, use `get_deployment` with its `integration_uuid`.",
	}
	return utils.NewMCPResponse(integrations, "Integrations retrieved successfully.", nextSteps)
}

// Branch defaults used when the caller does not supply one. The cloud editor
// has no terminal, so the user cannot run `git branch --show-current`; these
// two cover essentially every freshly published GitHub repository.
const (
	DefaultBranch  = "main"
	FallbackBranch = "master"
)

// reloadCostNotice is appended to any suggestion that the user reload the cloud
// editor window. Reloading restarts the editor and therefore tears down this MCP
// server process and the agent conversation with it, so the agent must present
// reloading as the user's choice with a stated cost — not as a routine step.
const reloadCostNotice = "IMPORTANT: reloading the editor window restarts it and WILL END this agent conversation — " +
	"the chat history for this session will be lost. Do not reload on the user's behalf and do not present it as a " +
	"required step. Mention it only as an optional action they may take when they are finished with this conversation, " +
	"and let them decide."

func createIntegration(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create integration."), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create integration."), nil
	}

	name, err := utils.GetRequiredStringArgument(request, "name")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create integration."), nil
	}

	description := utils.GetOptionalStringArgument(request, "description", "")

	subtype, err := utils.GetRequiredStringArgument(request, "subtype")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create integration."), nil
	}

	componentType, componentSubType, err := ResolveSubtype(subtype)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create integration."), nil
	}

	srcGitRepoUrl, err := utils.GetRequiredStringArgument(request, "repository_url")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create integration."), nil
	}
	// Normalize: strip trailing slash so ParseGitURL doesn't produce an empty repo name segment.
	srcGitRepoUrl = strings.TrimSuffix(srcGitRepoUrl, "/")

	// branch is optional: the cloud editor has no terminal, so the caller often
	// cannot run `git branch --show-current` to discover it. When omitted, try
	// DefaultBranch and then FallbackBranch, which covers essentially every
	// freshly published GitHub repository.
	srcGitRepoBranch := utils.GetOptionalStringArgument(request, "branch", "")
	branchWasSpecified := srcGitRepoBranch != ""
	if !branchWasSpecified {
		srcGitRepoBranch = DefaultBranch
	}

	srcGitRepoSubpath := utils.GetOptionalStringArgument(request, "subpath", "")

	buildpack := utils.GetOptionalStringArgument(request, "buildpack", "ballerina")

	repoOrg, repoName, err := repo.ParseGitURL(srcGitRepoUrl)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create integration."), nil
	}

	isPublicRepo, _ := utils.IsGHRepoPublic(ctx, srcGitRepoUrl)
	utils.MCPDebugf("create_integration: repo visibility check: isPublicRepo=%t", isPublicRepo)

	// Only guess when the caller did not tell us the branch. If they did, use it
	// as given rather than silently validating against a different one.
	candidateBranches := []string{srcGitRepoBranch}
	if !branchWasSpecified {
		candidateBranches = []string{DefaultBranch, FallbackBranch}
	}

	var isRepoValid bool
	for i, branch := range candidateBranches {
		srcGitRepoBranch = branch
		utils.MCPDebugf("create_integration: validating repo %s/%s branch=%s subpath=%q", repoOrg, repoName, branch, srcGitRepoSubpath)
		isRepoValid, err = utils.IsGHRepoValid(ctx, targetOrg, srcGitRepoUrl, branch, srcGitRepoSubpath, isPublicRepo)
		if err == nil && isRepoValid {
			break
		}
		// A missing GitHub App installation is not branch-specific, so trying
		// another branch cannot help — surface it immediately.
		if errors.Is(err, api.RepoAccessNeeded) {
			break
		}
		if i < len(candidateBranches)-1 {
			utils.MCPDebugf("create_integration: branch %s did not validate, trying %s", branch, candidateBranches[i+1])
		}
	}

	if err != nil {
		if errors.Is(err, api.RepoAccessNeeded) {
			return utils.NewMCPErrorResponse(err,
				fmt.Sprintf(
					"The WSO2 Cloud GitHub App does not have permission to access this repository. "+
						"The platform needs this access for ALL repositories (public and private) to set up "+
						"CI/CD webhooks and workflow files. "+
						"To fix this, complete both steps below and then retry create_integration:\n\n"+
						"Step 1 — Grant GitHub App access:\n"+
						"  • Open %s in your browser\n"+
						"  • Select the GitHub account or organisation that owns %s/%s\n"+
						"  • Choose 'Only select repositories', select %s, then click Install\n\n"+
						"Step 2 — Push your WSO2 Integrator source to the repository:\n"+
						"  • Your repository must contain a valid WSO2 Integrator project at the specified branch/subpath\n"+
						"  • At minimum the subpath must contain a Ballerina.toml and at least one .bal source file\n"+
						"  • Commit and push the code to branch '%s' before retrying\n\n"+
						"Once both steps are done, call create_integration again with the same parameters.",
					utils.GetGhAppInstallURL(ctx),
					repoOrg, repoName,
					repoName,
					srcGitRepoBranch,
				),
			), nil
		}
		return utils.NewMCPErrorResponse(err, "Failed to create integration."), nil
	}
	if !isRepoValid {
		return utils.NewMCPErrorResponse(fmt.Errorf("invalid repo"), "Failed to create integration."), nil
	}

	autoBuild := utils.GetOptionalBoolArgument(request, "auto_build", true)
	autoDeploy := utils.GetOptionalBoolArgument(request, "auto_deploy", true)

	// The platform uses the .git suffix to identify the repository in its internal URL parsing.
	// Always ensure it is present; normalise trailing slashes too.
	cleanRepoUrl := strings.TrimSuffix(srcGitRepoUrl, ".git") + ".git"

	params := &create.CreateComponentParams{
		ComponentName:        name,
		Description:          description,
		DisplayName:          name,
		ComponentType:        componentType,
		ComponentSubType:     componentSubType,
		Repo:                 repoName,
		RepoOrg:              repoOrg,
		RepoBranch:           srcGitRepoBranch,
		Subpath:              srcGitRepoSubpath,
		BuildPack:            buildpack,
		OriginCloud:          "devant",
		IsPublicRepo:         isPublicRepo,
		PullLatestSubmodules: true,
		AutoBuild:            &autoBuild,
		AutoDeploy:           &autoDeploy,
	}

	componentReqData, err := create.GetComponentKindForCreate(
		params,
		selectedProject.Handler,
		targetOrg.ID,
		targetOrg.UUID,
		cleanRepoUrl,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create integration."), nil
	}

	componentClient := utils.GetComponentClient(ctx)

	// Refuse to create a second copy of something already in this project.
	// Without this, asking to "deploy" twice silently produces a duplicate
	// integration building from the same source.
	if existing, listErr := componentClient.GetAllComponents(targetOrg.Handle, targetOrg.ID, selectedProject.ID, false); listErr == nil {
		if dup := findExistingIntegration(existing, name, repoOrg, repoName, srcGitRepoBranch, srcGitRepoSubpath); dup != nil {
			overview := fmt.Sprintf("%s/organizations/%s/projects/%s/components/%s/overview",
				utils.GetChoreConsoleBaseUrl(), targetOrg.Handle, selectedProject.Handler, dup.component.Handler)
			utils.MCPDebugf("create_integration: refusing duplicate (%s) of %q in project %q",
				dup.reason, dup.component.Name, selectedProject.Name)
			return utils.NewMCPErrorResponse(
				fmt.Errorf("%s", dup.detail(selectedProject.Name, overview)),
				"Integration already exists in this project — not creating a duplicate."), nil
		}
	} else {
		// Non-fatal: if the listing fails we simply lose the duplicate check
		// rather than blocking a legitimate create.
		utils.MCPDebugf("create_integration: duplicate pre-check skipped, listing failed: %v", listErr)
	}

	utils.MCPDebugf("create_integration: sending create request for %q in project %q", name, selectedProject.Handler)
	created, err := componentClient.CreateNewComponent(targetOrg.ID, selectedProject.Handler, *componentReqData)
	if err != nil {
		utils.MCPDebugf("create_integration: CreateNewComponent error: %v", err)
		return utils.NewMCPErrorResponse(fmt.Errorf("failed to create integration: %w", err), "Failed to create integration."), nil
	}
	utils.MCPDebugf("create_integration: create request succeeded; correlation-id=%s", created.CorrelationID)

	// Confirm via the declarative GET immediately — no lag on this path.
	// If the backend actually created the component it will be visible here right away.
	componentHandle := componentReqData.Metadata.Name
	utils.MCPDebugf("create_integration: confirming via declarative GET for handle=%q", componentHandle)
	declarativeKind, declErr := componentClient.GetComponentDeclarative(selectedProject.Handler, componentHandle, targetOrg.ID)
	if declErr != nil || declarativeKind.Metadata.Name == "" {
		utils.MCPDebugf("create_integration: declarative confirmation failed: err=%v, name=%q", declErr, declarativeKind.Metadata.Name)
		return utils.NewMCPErrorResponse(
			fmt.Errorf("component creation returned HTTP 201 but the declarative API found no record for %q — the backend silently rejected it (correlation-id: %s); provide this ID to the backend team", componentHandle, created.CorrelationID),
			"Integration creation could not be confirmed.",
		), nil
	}

	utils.MCPDebugf("create_integration: declarative confirmed; polling GraphQL for component handle=%q", componentHandle)
	// GraphQL may lag, so poll until the full component record (with UUID and
	// deployment tracks) appears. We already know creation succeeded above.
	var createdComponent *models.Component
	for attempt, delay := range []time.Duration{0, 3 * time.Second, 6 * time.Second, 9 * time.Second} {
		if attempt > 0 {
			time.Sleep(delay)
		}
		allComponents, listErr := componentClient.GetAllComponents(targetOrg.Handle, targetOrg.ID, selectedProject.ID, false)
		utils.MCPDebugf("create_integration: GraphQL poll attempt %d", attempt+1)
		if listErr != nil {
			utils.MCPDebugf("create_integration: poll error: %v", listErr)
			continue
		}
		for i := range allComponents {
			if allComponents[i].Handler == componentHandle {
				createdComponent = &allComponents[i]
				break
			}
		}
		if createdComponent != nil {
			utils.MCPDebugf("create_integration: component found in GraphQL on attempt %d; uuid=%s", attempt+1, createdComponent.Id)
			break
		}
	}
	if createdComponent == nil {
		return utils.NewMCPErrorResponse(
			fmt.Errorf("integration was created (confirmed via declarative API) but has not yet propagated to the project index; check the Devant console or retry shortly"),
			"Integration created but not yet indexed.",
		), nil
	}

	var codeServerBindWarning string
	// In the cloud editor, bind this editor instance to the integration that was
	// just created, so the editor knows which component it is working on. This is
	// the same call the VS Code extension makes via clirpc component/updateCodeServer.
	//
	// Deliberately non-fatal: the integration exists and is building by this
	// point, so failing the whole tool call over a binding failure would report a
	// successful creation as an error and invite a duplicate retry.
	//
	// We do NOT reload the editor ourselves, and could not if we wanted to — the
	// MCP protocol has no host-command primitive, and this server runs as a stdio
	// child of the editor, so a reload would kill the process before it could
	// return this result. Reloading is the user's call, and it ends the agent
	// conversation, so the notice below states that cost explicitly rather than
	// letting the agent suggest a reload as if it were free.
	if utils.IsCloudEditor() {
		utils.MCPDebugf("create_integration: cloud editor detected; binding code server to component uuid=%s", createdComponent.Id)
		if bindErr := utils.BindCodeServer(ctx, targetOrg, selectedProject.ID, createdComponent.Id, ""); bindErr != nil {
			utils.MCPDebugf("create_integration: code server binding failed: %v", bindErr)
			codeServerBindWarning = fmt.Sprintf(
				"NOTE: the integration was created successfully, but linking this cloud editor session to it failed (%v). "+
					"The integration itself is unaffected and is building normally. Tell the user they can open the "+
					"integration from the console, or reload the editor window to re-link it. %s", bindErr, reloadCostNotice)
		} else {
			utils.MCPDebugf("create_integration: code server bound to component uuid=%s", createdComponent.Id)
			codeServerBindWarning = "NOTE: this cloud editor session is now linked to the new integration. " +
				"The editor may not reflect the new integration in its UI until the window is reloaded. " +
				"This is cosmetic — the integration is already building and deploying regardless. " + reloadCostNotice
		}
	}

	overviewUrl := fmt.Sprintf("%s/organizations/%s/projects/%s/components/%s/overview",
		utils.GetChoreConsoleBaseUrl(),
		targetOrg.Handle,
		selectedProject.Handler,
		createdComponent.Handler,
	)

	nextSteps := []string{fmt.Sprintf("Integration created with uuid=%q (auto_build=%t, auto_deploy=%t).", createdComponent.Id, autoBuild, autoDeploy)}
	nextSteps = append(nextSteps, fmt.Sprintf("View the integration overview at: %s", overviewUrl))
	if autoBuild {
		nextSteps = append(nextSteps, fmt.Sprintf("A build was triggered automatically; use `get_builds` with integration_uuid=%q to track progress.", createdComponent.Id))
	} else {
		nextSteps = append(nextSteps, fmt.Sprintf("Auto-build was disabled; use `create_build` with integration_uuid=%q to trigger a build manually.", createdComponent.Id))
	}
	if autoDeploy {
		nextSteps = append(nextSteps, "It will deploy automatically once the build succeeds; use `get_deployment` to check status.")
		nextSteps = append(nextSteps, fmt.Sprintf("If the integration has mandatory configurables (e.g. API keys, secrets), open %s → find the environment card (e.g. Development) → click 'Configure' to set them.", overviewUrl))
	} else {
		nextSteps = append(nextSteps, "Auto-deploy was disabled; use `create_deployment` to deploy manually once it's built.")
	}
	if codeServerBindWarning != "" {
		nextSteps = append(nextSteps, codeServerBindWarning)
	}

	return utils.NewMCPResponse(createdComponent, "Integration created successfully.", nextSteps)
}

// duplicateMatch records why an integration is considered already present.
type duplicateMatch struct {
	component *models.Component
	reason    string // "name" or "source"
}

// detail renders the guidance shown to the caller. Both cases end the same way:
// deploying to a different project is the supported path, and it is done by
// passing a different project_uuid — not by inventing a new name here.
func (d duplicateMatch) detail(projectName, overviewURL string) string {
	common := fmt.Sprintf(
		"\n\nIt is already there as %q (integration_uuid: %s)\n  %s\n\n"+
			"Do NOT create a duplicate. What to do instead:\n"+
			"  • To ship new code: the integration is already wired to this repository — use `create_build` "+
			"with integration_uuid %q, then `create_deployment`. Pushing to the branch may also build it automatically.\n"+
			"  • To check what is running: use `get_deployment` with that integration_uuid.\n"+
			"  • To deploy into a DIFFERENT project: call create_integration again with that project's "+
			"`project_uuid` (use `get_projects` to find it). The same repository can back integrations in "+
			"several projects; what is not allowed is two copies in the same one.\n"+
			"Ask the user which of these they meant before doing anything.",
		d.component.Name, d.component.Id, overviewURL, d.component.Id)

	switch d.reason {
	case "name":
		return fmt.Sprintf("an integration named %q already exists in project %q.%s",
			d.component.Name, projectName, common)
	default:
		return fmt.Sprintf("this repository, branch and path are already deployed in project %q.%s",
			projectName, common)
	}
}

// findExistingIntegration reports an integration in the project that the request
// would duplicate — either by name, or by pointing at the same source.
//
// Source matching deliberately includes the subpath: a monorepo legitimately
// backs several integrations from one repository and branch, distinguished only
// by path, and those must not be treated as duplicates of each other.
func findExistingIntegration(existing []models.Component, name, repoOrg, repoName, branch, subpath string) *duplicateMatch {
	norm := func(s string) string { return strings.ToLower(strings.Trim(strings.TrimSpace(s), "/")) }

	for i := range existing {
		c := &existing[i]
		if !pkgcomponent.IsIntegrationDisplayType(c.DisplayType) {
			continue
		}

		if strings.EqualFold(c.Name, name) || strings.EqualFold(c.Handler, name) {
			return &duplicateMatch{component: c, reason: "name"}
		}

		r := c.Repository
		if norm(r.OrganizationApp) == norm(repoOrg) &&
			norm(r.NameApp) == norm(repoName) &&
			norm(r.BranchApp) == norm(branch) &&
			norm(r.AppSubPath) == norm(subpath) &&
			norm(repoName) != "" {
			return &duplicateMatch{component: c, reason: "source"}
		}
	}
	return nil
}
