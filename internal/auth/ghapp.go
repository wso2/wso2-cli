package auth

import (
	"fmt"
	"time"

	"github.com/cli/cli/v2/pkg/iostreams"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/browser"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
)

func getGithubAuthCode() (code string, err error) {

	port, err := getRandomPort()
	if err != nil {
		return
	}
	timeout := 120 * time.Second

	callbackURL := getCallbackURL(port)

	link, err := GitClient.GetOauthURL(callbackURL)
	if err != nil {
		return
	}

	const (
		OPEN_BROWSER = "Open link using default browser"
		DEFAULT      = "Get a link to open in browser manually"
	)

	selectOpts := ""
	err = prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{
			Message: i18n.T("How would you like to open the GitHub auth link?"),
			Values:  []string{OPEN_BROWSER, DEFAULT},
		},
		&selectOpts,
	).Prompt()

	if err != nil {
		return
	}

	switch selectOpts {
	case OPEN_BROWSER:
		brwsr := browser.New("", iostreams.System().Out, iostreams.System().ErrOut)
		brwsr.Browse(link)
	default:
		fmt.Fprintln(
			utils.IO.Out,
			fmt.Sprintf(i18n.T("\nCopy the following link and open it in your browser:\n\n%s\n"), link),
		)
	}

	linkSpinner := utils.CreateSpinner(i18n.T(" Waiting for you to complete the GitHub auth process..."), "")
	linkSpinner.Start()

	code, err = ListenForAuthCode(port, timeout)
	linkSpinner.Stop()

	return
}

func getGithubAccessToken(code string, orgId string) (err error) {
	return GitClient.ObtainAccessToken(code, orgId)
}

func DoGithubAuth(orgId string) (err error) {
	code, err := getGithubAuthCode()
	if err != nil {
		return
	}

	err = getGithubAccessToken(code, orgId)
	if err != nil {
		return
	}

	return
}

func IsGitRepoAuthorized(gitOrgName, gitRepoName, orgId, credRef string) (isAuthorized bool, err error) {

	checkRepoAuthSpinner := utils.CreateSpinner(i18n.T(" Checking if repository is accessible..."), "")
	checkRepoAuthSpinner.Start()

	// TODO: Implement this after declarative API is ready
	// isPublicRepoResp, _ := GitClient.IsPublicRepo(orgName, repoName, orgId)
	// if isPublicRepoResp.IsPublicRepo {
	// 	checkRepoAuthSpinner.Stop()
	// 	return true, nil
	// }

	authorizedGitOrgs, err := GitClient.GetAuthorizedGithubRepos(orgId, credRef)
	checkRepoAuthSpinner.Stop()
	// If the retriving failed, try to authorize the user
	if err != nil {
		fmt.Println(i18n.T("Please authorize the Integration Platform to access your github repositories"))
		err = DoGithubAuth(orgId)
		if err != nil {
			return
		}
		checkRepoAuthSpinner.Start()
		authorizedGitOrgs, err = GitClient.GetAuthorizedGithubRepos(orgId, credRef)
		checkRepoAuthSpinner.Stop()
		if err != nil {
			return
		}
	}

	for _, authorizedGitOrg := range authorizedGitOrgs {
		if gitOrgName != "" && (authorizedGitOrg.OrgName == gitOrgName || authorizedGitOrg.OrgHandler == gitOrgName) {
			for _, authorizedGitRepo := range authorizedGitOrg.Repositories {
				if authorizedGitRepo.Name == gitRepoName {
					isAuthorized = true
					return
				}
			}
		}
	}

	return
}
