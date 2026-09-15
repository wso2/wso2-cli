package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/pflag"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/util/common"
	"github.com/wso2/integration-platform-tools/pkg/util/repo"
)

func ShortCommitId(commitId string) string {
	return commitId[:8]
}

func AddGitOrgFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, "git-org", "g", "", i18n.T("Organizations of the Git repository"))
}

const GitRepoFlag = "git-repo"
const GitBranchFlag = "git-branch"

func AddGitRepoFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, GitRepoFlag, "r", "", "URL of the Git repository")
}

func AddGitBranchFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, GitBranchFlag, "b", "", i18n.T("git repository branch"))
}

func HandleGitRemoteURLs(repoRoot *string, gitRepo *string) (remoteUrl string, detected bool, err error) {
	if *gitRepo != "" {
		return *gitRepo, false, nil
	}
	var selectedRemote string

	if *repoRoot != "" {
		remoteUrls, err := repo.GetGitRemotesNames(*repoRoot)
		if err != nil || len(remoteUrls) == 0 {
			remoteUrl, err = promptForRepoUrl()
			return remoteUrl, false, err
		}

		if len(remoteUrls) > 1 {
			// if multiple remotes found, let user select one
			const (
				DETECTED_GIT_REMOTES = "Detect remote repository url from the current Git directory"
				ENTER_MANUALLY       = "Enter remote repository URL manually"
			)

			selection := ""

			err = prompt.NewPromptSelectMessage[string](
				prompt.PromptSelectOpts[string]{
					Message: i18n.T("Configure source repository:"),
					Values:  []string{DETECTED_GIT_REMOTES, ENTER_MANUALLY},
				},
				&selection,
			).Prompt()

			if err != nil {
				return "", false, err
			}

			switch selection {
			case DETECTED_GIT_REMOTES:
				err = prompt.NewPromptSelectMessage[string](
					prompt.PromptSelectOpts[string]{Message: i18n.T("Select remote repository URL:"), Values: remoteUrls},
					&selectedRemote,
				).Prompt()

				if err != nil {
					return "", false, err
				}

				return selectedRemote, true, nil
			default:
				remoteUrl, err = promptForRepoUrl()
				return remoteUrl, false, err
			}
		} else {
			// if only one remote found, ask user if it's okay to proceed with component creation
			proceedWithCurrentRemote := false
			prompt := &survey.Confirm{
				Message: fmt.Sprintf(
					i18n.T("Detected the following Git remote URL. Do you want to proceed with it?\n%s"),
					utils.CS.Blue(remoteUrls[0]),
				),
				Default: true,
			}
			err = survey.AskOne(prompt, &proceedWithCurrentRemote)
			if err != nil {
				return "", false, err
			}
			if proceedWithCurrentRemote {
				return remoteUrls[0], true, nil
			} else {
				remoteUrl, err = promptForRepoUrl()
				return remoteUrl, false, err
			}
		}
	} else {
		remoteUrl, err = promptForRepoUrl()
		return
	}
}

func promptForRepoUrl() (url string, err error) {

	err = prompt.NewPromptInputMessage(
		prompt.PromptInputOpts{
			Message: i18n.T("Remote repository URL:"),
			Validate: func(repoUrl string) error {
				_, _, err = repo.ParseGitURL(repoUrl)
				return err
			},
		},
		&url,
	).Prompt()

	if err != nil {
		return
	}
	return
}

func HandleComponentBranch(gitOrgName *string, gitRepoName *string, selectedBranch *string, orgId string, credRef string) error {
	branchFetchSpinner := utils.CreateSpinner(i18n.T(" Fetching branches..."), "")
	branchFetchSpinner.Start()
	branches, err := auth.GitClient.GetRepoBranches(*gitOrgName, *gitRepoName, orgId, credRef)
	branchFetchSpinner.Stop()
	if err != nil {
		return err
	}

	if *selectedBranch == "" {
		err = prompt.NewPromptSelectMessage[string](
			prompt.PromptSelectOpts[string]{
				Message: i18n.T("Branch: "),
				Values:  branches,
				Default: repo.GetDefaultBranch(branches),
			},
			selectedBranch,
		).Prompt()

		if err != nil {
			return err
		}
	} else if !common.StringExistsInSlice(*selectedBranch, branches) {
		return errors.New("invalid branch")
	}
	return nil
}

func HandleGitRepoAuthorization(repoUrl, orgId, orgUUID, credRef string) (repoOrg string, repoName string, err error) {
	gitOrgName, gitRepoName, err := repo.ParseGitURL(repoUrl)
	if err != nil {
		return "", "", err
	}

	isAuthorized, err := auth.IsGitRepoAuthorized(gitOrgName, gitRepoName, orgId, credRef)
	if err != nil {
		return "", "", err
	}

	if !isAuthorized {
		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			%s WSO2 Integration Platform is unable to access the given repository

			Please make sure that you have installed the WSO2 Integration Platform GitHub App to your repository.
			If you are a free-tier user, only public repositories are supported.

			To proceed, please install the WSO2 Integration Platform GitHub App using the following link:
				%s
		
			Then try again.

		`), utils.CS.Red("!"), auth.GetEnvConfig().GhApp.InstallUrl))

		return "", "", api.RepoAccessNeeded
	}

	return gitOrgName, gitRepoName, nil
}

func GetRepoRootAndSubPath() (repoRootPath string, repoRelativePath string, err error) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	repoRootPath, err = repo.GetGitRepoRoot(currentDirectory)
	if err != nil {
		return "", "", fmt.Errorf("%s", i18n.T("invalid Git repository"))
	}

	repoRelativePath, err = filepath.Rel(repoRootPath, currentDirectory)
	if err != nil {
		return repoRootPath, "", fmt.Errorf("%s", i18n.T("failed to find sub directory path"))
	}

	if repoRelativePath == "." {
		return repoRootPath, "", nil
	}
	return repoRootPath, repoRelativePath, nil
}

func GetComponentSubPath(componentPath string) (repoRootPath string, repoRelativePath string, err error) {
	repoRootPath, err = repo.GetGitRepoRoot(componentPath)
	if err != nil {
		return "", "", fmt.Errorf("%s", i18n.T("invalid Git repository"))
	}

	repoRelativePath, err = filepath.Rel(repoRootPath, componentPath)
	if err != nil {
		return repoRootPath, "", fmt.Errorf("%s", i18n.T("failed to find sub directory path"))
	}

	if repoRelativePath == "." {
		return repoRootPath, "", nil
	}
	return repoRootPath, repoRelativePath, nil
}

func GetRepoStructure(
	gitOrgName string,
	gitRepoName string,
	gitBranch string,
	isPublicRepo bool,
	orgID string,
) ([]component.PathEntry, error) {
	fetchDirTreeSpinner := utils.CreateSpinner(i18n.T("Fetching directory structure..."), "")
	fetchDirTreeSpinner.Start()
	resp, err := auth.ComponentClient.GetRepoDirStructure(
		gitOrgName,
		gitRepoName,
		gitBranch,
		isPublicRepo,
		orgID,
	)
	fetchDirTreeSpinner.Stop()

	return resp, err
}
