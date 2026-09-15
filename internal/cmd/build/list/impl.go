package list

import (
	"fmt"
	"strings"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

func handleBuildListCommand(flags *BuildListFlags) error {
	// An invalid --output falls back to a persisted OUTPUT_FORMAT (with a
	// warning); with no global set it is an error.
	outputFormat, err := common.ResolveOutputFormat(flags.OutputFlag)
	if err != nil {
		return err
	}

	orgId, projectId, err := common.ResolveContext(flags.Org, flags.Project)
	if err == nil {
		if flags.Org == "" {
			flags.Org = orgId
		}

		if flags.Project == "" {
			flags.Project = projectId
		}
	}

	org, err := common.ResolveTargetOrganization(flags.Org)
	if err != nil {
		return fmt.Errorf(i18n.T("Error resolving organization: %w"), err)
	}

	project, err := common.ResolveTargetProject(org, flags.Project)
	if err != nil {
		return fmt.Errorf(i18n.T("Error resolving project: %w"), err)
	}

	componentInfo, err := common.ResolveTargetComponent(org, project.ID, flags.Component)
	if err != nil {
		return fmt.Errorf(i18n.T("Error resolving component: %w"), err)
	}

	if strings.HasPrefix(strings.ToLower(componentInfo.DisplayType), "byoi") {
		utils.PrintInfo("%s", i18n.T("Cannot list builds for pre-built image based components\n"))
		return nil
	}

	compInfoWithRepo, err := common.GetComponentWithRepoData(org.ID, componentInfo.Handler, project.ID)
	if err != nil {
		return fmt.Errorf(i18n.T("Error resolving component: %w"), err)
	}
	deploymentTrack, err := common.ResolveDeploymentTrack(
		compInfoWithRepo.DeploymentTracks,
		flags.DeploymentTrack,
	)
	if err != nil {
		return fmt.Errorf(i18n.T("Error resolving deployment track: %w"), err)
	}

	versionId := deploymentTrack.Id

	// Current Git API proxy impl doesn't have deployment tracks
	if componentInfo.DisplayType == component.DisplayTypeGitProxy {
		for _, v := range compInfoWithRepo.ApiVersions {
			if v.Latest {
				versionId = v.VersionId
			}
		}
	}

	buildList, err := common.GetDeploymentBuild(org.ID, componentInfo.Name, project.Handler, versionId)
	if err != nil {
		return fmt.Errorf(i18n.T("Error getting deployment builds: %w"), err)
	}

	if outputFormat.IsStructured() {
		return printBuildListStructured(*buildList, outputFormat, *componentInfo, *project)
	}

	printBuildListTable(*buildList, *componentInfo, *project)
	return nil
}
