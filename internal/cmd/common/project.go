package common

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

const ProjectFlagName = "project"

func FetchProjectsOfOrg(currentOrg api.Organization) (projects []models.Project, err error) {
	spinner := utils.CreateSpinner(i18n.T(" Fetching projects..."), "")
	spinner.Start()
	projects, err = auth.ProjectClient.GetProjectsByOrgID(currentOrg.ID)
	spinner.Stop()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to get projects: %w"), err)
	}
	return
}

func ResolveTargetProject(selectedOrg *api.Organization, projectFlagVal string) (*models.Project, error) {
	var selectedProject *models.Project
	projects, err := FetchProjectsOfOrg(*selectedOrg)
	if err != nil {
		return nil, fmt.Errorf("%s", i18n.T(" failed to get projects"))
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("%w - no projects found within organization", api.ErrFailedToResolveProj)
	}

	if projectFlagVal != "" {
		for _, project := range projects {
			if project.Name == projectFlagVal || project.Handler == projectFlagVal || project.ID == projectFlagVal {
				selectedProject = &project
				break
			}
		}
		// An explicit --project value that does not match should be a hard
		// error, not a silent fallback to the interactive picker — a typo'd
		// flag value should never be ignored.
		if selectedProject == nil {
			return nil, fmt.Errorf(i18n.T("project %q not found in organization"), projectFlagVal)
		}
	}

	if selectedProject == nil {
		selectedProjectName, err := promptToSelectProject(projects)
		if err != nil {
			if errors.Is(err, internal.ErrNonInteractive) {
				return nil, utils.CreateNonInteractiveError("project selection", "project")
			}
			return nil, fmt.Errorf("%s", i18n.T(" failed to select the project"))
		}
		for _, project := range projects {
			if project.Name == selectedProjectName {
				selectedProject = &project
				break
			}
		}
	}

	if selectedProject == nil {
		return nil, fmt.Errorf("%s", i18n.T(" invalid project selection"))
	}

	return selectedProject, nil
}

func promptToSelectProject(projects []models.Project) (string, error) {
	var selectedProjectName string
	var projectNames []string

	for _, project := range projects {
		projectNames = append(projectNames, project.Name)
	}

	err := prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{Message: "Project:", Values: projectNames},
		&selectedProjectName,
	).Prompt()

	if err != nil {
		return "", err
	}

	return selectedProjectName, nil
}

func GetProjectEnv(
	orgUuid string,
	orgId string,
	projectId string,
	selectedEnvName string) (env *project.ProjectEnvironment, err error) {

	var selectedEnv *project.ProjectEnvironment
	var envNames []string
	envs, err := GetAllProjectEnv(orgUuid, orgId, projectId)
	if err != nil {
		return nil, err
	}

	for _, item := range *envs {
		envNames = append(envNames, item.Name)
		if item.Name == selectedEnvName || item.Env == selectedEnvName {
			selectedEnv = &item
			break
		}
	}

	if selectedEnv == nil {
		err = prompt.NewPromptSelectMessage[string](
			prompt.PromptSelectOpts[string]{Message: i18n.T("Environment: "), Values: envNames},
			&selectedEnvName,
		).Prompt()

		if err != nil {
			if errors.Is(err, internal.ErrNonInteractive) {
				return nil, utils.CreateNonInteractiveError("environment selection", "env")
			}
			utils.PrintInfo(i18n.T("%s Failed to select environment\n"), utils.CS.Red("!"))
			return nil, err
		}

		for _, item := range *envs {
			if item.Name == selectedEnvName {
				selectedEnv = &item
				break
			}
		}
	}

	if selectedEnv == nil {
		return nil, errors.New(i18n.T("failed to get environment"))
	}

	return selectedEnv, nil
}

func GetAllProjectEnv(orgUuid string, orgId string, projectId string) (envs *[]project.ProjectEnvironment, err error) {
	getEnvSpinner := utils.CreateSpinner(i18n.T(" Fetching environments..."), "")
	getEnvSpinner.Start()
	envs, err = auth.ProjectClient.GetProjectEnvironments(orgUuid, orgId, projectId)
	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to get project env details: %w"), err)
	}

	getEnvSpinner.Stop()
	return envs, nil
}

func GetDevProjectEnv(envs *[]project.ProjectEnvironment) (env *project.ProjectEnvironment, err error) {
	for _, env := range *envs {
		if env.Env == "dev" {
			return &env, nil
		}
	}
	return nil, errors.New(i18n.T("failed to find dev env"))
}

func AddProjectFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, ProjectFlagName, "p", "", i18n.T("project ID, name, or handle"))
}

func AddEnvFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, "env", "e", "", i18n.T("name of the environment (development, production, etc.)"))
}

func GenRepoUrl(gitProvider string, gitOrg string, gitRepo string, serverUrl string) string {
	providerUrl := serverUrl

	switch gitProvider {
	case "github":
		providerUrl = "https://github.com"
	case "bitbucket":
		providerUrl = "https://bitbucket.org"
	}

	return fmt.Sprintf("%s/%s/%s", providerUrl, gitOrg, gitRepo)
}
