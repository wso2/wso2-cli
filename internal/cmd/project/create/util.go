package create

import (
	"fmt"
	"strconv"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"

	commonCmd "github.com/wso2/integration-platform-tools/internal/cmd/common"
)

const (
	MonoRepoProject   = "mono-repository"
	MultiRepoProject  = "multi-repository"
	PROJECT_LINK_FILE = "project.yaml" // todo: delete
)

type ProjectLinkData struct {
	ProjectId string `yaml:"projectId"`
	OrgId     int    `yaml:"orgId"`
}

var projectTypes = []string{MonoRepoProject, MultiRepoProject}

func resolveProjectName(params *CreateProjectParams, existingProjects []models.Project) error {
	if params.Name == "" {
		err := prompt.NewPromptInputMessage(
			prompt.PromptInputOpts{
				Message: i18n.T("Project name:"),
				Validate: func(s string) error {
					return validateProjectName(s, existingProjects)
				},
			},
			&params.Name,
		).Prompt()

		if err != nil {
			return fmt.Errorf("%s", i18n.T(" failed to get a valid project name"))
		}
	}
	return validateProjectName(params.Name, existingProjects)

}

func handleCreateProject(params *CreateProjectParams, org api.Organization) (projInfo *models.Project, err error) {
	orgIdInt, err := strconv.Atoi(org.ID)
	if err != nil {
		return
	}

	projectReq := project.GetProjectMutationRequest{
		OrgID:        orgIdInt,
		OrgHandler:   org.Handle,
		Name:         params.Name,
		Description:  params.Description,
		Version:      "1.0.0",
		Region:       region.GetCurrentRegion(),
		Repository:   "",
		Branch:       "",
		CredentialId: "",
	}

	createProjectSpinner := utils.CreateSpinner(i18n.T(" Creating new project..."), "")
	createProjectSpinner.Start()
	projInfo, err = auth.ProjectClient.CreateProject(projectReq, org.ID)
	createProjectSpinner.Stop()

	if err != nil {
		return
	}

	return
}

func validateProjectName(projectName string, existingProjects []models.Project) error {
	err := prompt.ValidateNotEmpty(projectName)
	if err != nil {
		return err
	}

	err = commonCmd.ValidateNameText(projectName)
	if err != nil {
		return err
	}

	for _, item := range existingProjects {
		if item.Name == projectName {
			return fmt.Errorf("project name already exists")
		}
	}

	return nil
}
