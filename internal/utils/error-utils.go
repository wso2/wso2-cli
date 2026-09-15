package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/util/repo"
)

func HandleErr(err error) {
	if err != nil {
		if errors.Is(err, api.RepoAccessNeeded) || errors.Is(err, api.ErrNoAccountFound) {
			// we do not print any messages
			os.Exit(1)
		} else if errors.Is(err, api.ErrMaxProjectCountReached) {
			fmt.Fprintln(IO.ErrOut, heredoc.Docf(`

				Failed to create Project.

				Maximum number of projects allowed within the free tier has been reached.

				Please upgrade your tier and try again.
			`))
			os.Exit(1)
		} else if errors.Is(err, api.ErrMaxComponentCountReached) {
			fmt.Fprintln(IO.ErrOut, heredoc.Docf(`

				Failed to create Integration.

				Maximum number of integrations allowed within the free tier has been reached.

				Please upgrade your tier and try again.
			`))
			os.Exit(1)
		} else if errors.Is(err, api.ErrNotLoggedIn) {
			fmt.Fprintln(IO.ErrOut, heredoc.Docf(`

			You are not logged in

				To log in :
					Run: %s
			`, CS.Bold("$ wso2-integration-platform login")))
			os.Exit(1)
		} else if errors.Is(err, api.ErrTokenNotValid) || errors.Is(err, api.ErrRefreshToken) {
			fmt.Fprintln(IO.ErrOut, heredoc.Docf(`

			The access token is not valid

				To log in again :
					Run: %s
			`, CS.Bold("$ wso2-integration-platform login")))
			os.Exit(1)
		} else if errors.Is(err, api.ErrForbidden) {
			fmt.Fprintln(IO.ErrOut, heredoc.Docf(`

			You don't have permission to perform this action.

				To change the active organization :
					Run: %s
				To sign in using a different account :
					Run: %s
			`, "$ wso2-integration-platform change-org", "$ wso2-integration-platform login"))
			os.Exit(1)
		} else if errors.Is(err, api.ErrNotWithinProject) {
			fmt.Fprintln(IO.ErrOut, heredoc.Docf(`

			You need to be within a project in order to run this command

				To list the projects of active organizations :
					Run: %s

				To clone a project :
					Run: %s
			`, CS.Bold("$ wso2-integration-platform list projects"), CS.Bold("$ wso2-integration-platform project clone")))
			os.Exit(1)
		} else if errors.Is(err, repo.ErrGitCredentialsNotFound) {
			fmt.Fprintln(IO.ErrOut, heredoc.Doc(`
				Cannot find any existing git credentials to clone the repository.

				Please make sure you have confgiured git cli with your credentials.
			`))
			os.Exit(1)
		} else if errors.Is(err, repo.ErrGitRepoNotInitialized) {
			fmt.Fprintln(IO.ErrOut, heredoc.Doc(`
				The repository is not initialized. Please initialize the repository and try again.
			`))
			os.Exit(1)
		} else if errors.Is(err, internal.ErrUserInterrupted) {
			fmt.Fprintln(IO.ErrOut, "^C")
			os.Exit(1)
		} else if errors.Is(err, internal.ErrNonInteractive) {
			// This should be handled by individual commands with specific parameter info
			fmt.Fprintln(IO.ErrOut, "Error: missing required parameter in non-interactive mode")
			os.Exit(1)
		} else if errors.Is(err, api.ErrNotWithinGitRepo) {
			fmt.Fprintln(IO.ErrOut, heredoc.Docf(`

				Invalid Git repository.

				Please either run this command from within a Git repository or pass the repository path as an argument.

			`))
		} else if errors.Is(err, api.ErrFailedToResolveComp) {
			fmt.Fprintln(IO.ErrOut, fmt.Sprintf("Error: %s", err.Error()))
			os.Exit(1)
		} else if errors.Is(err, api.ComponentYamlNotFound) {
			fmt.Fprintln(IO.Out,
				heredoc.Docf(`

					%s

					%s
					%s
				`,
					CS.Yellow(i18n.T("Endpoint configurations not found")),
					CS.Yellow(i18n.T("Please make sure that you have a valid component.yaml file within .wso2 folder of the component directory and that it is commited to the repo.")),
					CS.Yellowf(
						i18n.T("\nFor more info on configuring service endpoints:\n%s"),
						"https://wso2.com/choreo/docs/develop-components/configure-endpoints/",
					),
				),
			)
			os.Exit(1)
		} else if errors.Is(err, api.NoGitCredentialsFound) {
			fmt.Fprintln(
				IO.Out,
				heredoc.Docf(i18n.T(`
                    No Git credentials are configured for the Organization.
                    For information on configuring credentials: https://wso2.com/choreo/docs/develop-components/develop-components-with-git/#connect-a-git-repository-to-choreo
                `)),
			)
		} else {
			if strings.Trim(err.Error(), " ") == "" {
				fmt.Fprintln(IO.ErrOut, heredoc.Docf(`

				Error: %s

				`, i18n.T("unknown error")))
				os.Exit(1)

			} else {
				fmt.Fprintln(IO.ErrOut, heredoc.Docf(`

				Error: %s

				`, err.Error()))
			}
			os.Exit(1)
		}
	}
}

// CreateNonInteractiveError creates a descriptive error for missing parameters in non-interactive mode.
//
// context describes what the value is for (e.g. "component selection") and is
// NOT a command name. An earlier version interpolated it into
// "Use 'wso2-integration-platform <context> --help'", which pointed every
// caller at a command that does not exist ("component selection",
// "deployment track selection", ...).
func CreateNonInteractiveError(context string, flagName string) error {
	return fmt.Errorf("missing required flag --%s for %s; run the command with --help for usage information", flagName, context)
}
