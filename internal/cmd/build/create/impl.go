package create

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

func HandleBuildComponent(params *BuildComponentParams) error {
	orgId, projectId, err := common.ResolveContext(params.OrgFlag, params.ProjectFlag)

	if err == nil {
		if params.OrgFlag == "" {
			params.OrgFlag = orgId
		}

		if params.ProjectFlag == "" {
			params.ProjectFlag = projectId
		}
	}

	selectedOrg, err := common.ResolveTargetOrganization(params.OrgFlag)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(selectedOrg, params.ProjectFlag)
	if err != nil {
		return err
	}

	remoteComponent, err := common.ResolveTargetComponent(selectedOrg, project.ID, params.ComponentFlag)
	if err != nil {
		return err
	}

	if strings.HasPrefix(strings.ToLower(remoteComponent.DisplayType), "byoi") {
		utils.PrintInfo("%s", i18n.T("Cannot trigger a build for pre-built image based components\n"))
		return nil
	}

	remoteComWithRepoData, err := common.GetComponentWithRepoData(selectedOrg.ID, remoteComponent.Handler, project.ID)
	if err != nil {
		return err
	}

	resp, err := auth.ComponentClient.GetComponentInitStatus(selectedOrg.ID, selectedOrg.Handle, project.ID, remoteComponent.Id)
	if err != nil {
		return err
	}

	if resp.Data.Status == "queued" || resp.Data.Status == "in_progress" {
		utils.PrintInfo("%s", i18n.T("Component initialization is in progress. Please run the command in a while...\n"))
		return nil
	}

	deploymentTrack, err := common.ResolveDeploymentTrack(
		remoteComWithRepoData.DeploymentTracks,
		params.DeploymentTrackFlag,
	)
	if err != nil {
		return err
	}

	if strings.HasSuffix(strings.ToLower(remoteComponent.DisplayType), "service") &&
		!strings.HasPrefix(strings.ToLower(remoteComponent.DisplayType), "ballerina") &&
		!strings.HasPrefix(strings.ToLower(remoteComponent.DisplayType), "mi") &&
		!strings.HasPrefix(strings.ToLower(remoteComponent.DisplayType), "proxy") {
		// Check for deployment.yaml in component directory
		safeToBuild, err := common.IsSafeToBuildAndDeploy(project, remoteComponent, selectedOrg, deploymentTrack)
		if err != nil && !errors.Is(err, api.FuncionalityNotSupported) {
			return err
		}

		if !safeToBuild {
			return api.ComponentYamlNotFound
		}
	}

	versionId := deploymentTrack.Id

	// Current Git API proxy impl doesn't have deployment tracks
	if remoteComponent.DisplayType == component.DisplayTypeGitProxy {
		for _, v := range remoteComWithRepoData.ApiVersions {
			if v.Latest {
				versionId = v.VersionId
			}
		}
	}

	buildList, err := common.GetDeploymentBuild(
		selectedOrg.ID,
		remoteComponent.Name,
		project.Handler,
		versionId,
	)
	if err != nil {
		return fmt.Errorf(i18n.T("Error getting deployment builds: %w"), err)
	}

	for _, build := range *buildList {
		if build.Status.Status == "in_progress" || build.Status.Status == "queued" {
			fmt.Fprintln(utils.IO.ErrOut, i18n.T("A build is already in progress."))
			return nil
		}
	}

	latestCommitHash, err := common.GetLatestCommit(remoteComponent.Id, deploymentTrack.Branch, selectedOrg.ID)
	if err != nil {
		return err
	}

	deploymentBuildRes, err := triggerComponentBuild(
		selectedOrg.ID,
		remoteComponent.Name,
		project.Handler,
		versionId,
		latestCommitHash,
	)
	if err != nil {
		return err
	}

	utils.PrintInfo("%s", heredoc.Docf(`

		Build for component '%s' has been successfully triggered (Build ID: %s).

		To view the status of this build :
			%s


	`,
		remoteComponent.Name,
		strconv.Itoa(int(deploymentBuildRes.Status.RunID)),
		fmt.Sprintf(`$ wso2-integration-platform describe build %s --project="%s" --component="%s" --deployment-track="%s"`, strconv.Itoa(int(deploymentBuildRes.Status.RunID)), project.Name, remoteComponent.Name, deploymentTrack.Branch),
	))

	return nil
}
