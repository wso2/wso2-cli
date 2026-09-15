package link

import (
	"path/filepath"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
)

func HandleLinkComponent(params *LinkComponentParams) error {

	selectedOrg, err := common.ResolveTargetOrganization(params.orgFlag)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(selectedOrg, params.projectFlag)
	if err != nil {
		return err
	}

	remoteComponent, err := common.ResolveTargetComponent(selectedOrg, project.ID, params.componentFlag)
	if err != nil {
		return err
	}

	declarativeComponent, err := auth.ComponentClient.GetComponentDeclarative(project.Handler, remoteComponent.Name, selectedOrg.ID)
	if err != nil {
		return err
	}

	repoRootPath, relativePath, err := common.GetRepoRootAndSubPath()
	if err != nil {
		return err
	}

	if relativePath != declarativeComponent.Spec.Source.Github.Path {
		utils.PrintError("%s", heredoc.Docf(`

		Invalid component directory.
		Please navigate into the component directory and re-run this command.

		`))
		return nil
	}

	err = CreateComponentLink(filepath.Join(repoRootPath, relativePath), project.Handler, selectedOrg.Handle, remoteComponent.Name)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	utils.PrintInfo("%s", heredoc.Docf(`

		%s file successfully created within component directory.

		`,
		constants.COMPONENT_LINK_FILE,
	))

	return nil
}
