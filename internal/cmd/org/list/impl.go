package list

import (
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

func handleOrgList(params *OrgListOptions) error {
	// An invalid --output falls back to a persisted OUTPUT_FORMAT (with a
	// warning); with no global set it is an error.
	outputFormat, err := common.ResolveOutputFormat(params.outputFlag)
	if err != nil {
		return err
	}

	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return api.ErrNotLoggedIn
	}

	orgs, selectedOrg, err := common.FetchUserOrgs()
	if err != nil {
		return err
	}

	currentUser, err := auth.GetCurrentUser()
	if err != nil {
		return fmt.Errorf(i18n.T("failed to fetch user: %w"), err)
	}

	if outputFormat.IsStructured() {
		return printOrgListStructured(orgs, outputFormat)
	}

	fmt.Fprintf(utils.IO.Out, i18n.T("\nShowing organizations of user '%s'\n\n"), currentUser.DisplayName)

	printOrgList(*selectedOrg, orgs, *currentUser)

	// sleep for 1 second to allow the user to read the list of orgs
	time.Sleep(1 * time.Second)

	fmt.Fprintln(utils.IO.Out, heredoc.Docf(i18n.T(`

		To change the active organization :
			%s
	`), "$ wso2-integration-platform change-org"))
	return nil
}
