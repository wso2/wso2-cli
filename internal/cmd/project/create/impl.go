package create

import (
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

func HandleProjectCreate(params *CreateProjectParams) error {
	selectedOrg, err := common.ResolveTargetOrganization(params.OrgFlag)
	if err != nil {
		return err
	}

	existingProjects, err := common.FetchProjectsOfOrg(*selectedOrg)
	if err != nil {
		return err
	}

	err = resolveProjectName(params, existingProjects)
	if err != nil {
		return err
	}

	_, err = handleCreateProject(params, *selectedOrg)
	if err != nil {
		return err
	}

	utils.PrintInfo(i18n.T("\nProject '%s' has been successfully created!\n"), params.Name)
	return nil
}
