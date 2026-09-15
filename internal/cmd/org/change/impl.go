package change

import (
	"errors"
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

func handleOrgChange(changeOrgParam *ChangeOrgParams) error {
	cUser, err := auth.GetCurrentUser()

	if err != nil {
		return err
	}

	if cUser.IsPATLogin {
		return errors.New("Change org is not supported with PAT login")
	}

	orgs, _, err := common.FetchUserOrgs()

	var targetOrg *api.Organization

	if changeOrgParam.orgFlag != "" {
		for _, org := range orgs {
			if org.Handle == changeOrgParam.orgFlag || org.Name == changeOrgParam.orgFlag {
				targetOrg = &org
				break
			}
		}
		if targetOrg == nil {
			return fmt.Errorf(i18n.T("organization with handle %s not found"), changeOrgParam.orgFlag)
		}
	} else {
		targetOrg, err = promptToSelectOrg(orgs)
		if err != nil {
			return err
		}
	}

	setOrgSpinner := utils.CreateSpinner(i18n.T(" Setting active organization..."), "\n")
	setOrgSpinner.Start()
	err = auth.SetSelectedOrg(targetOrg, orgs)
	setOrgSpinner.Stop()

	if err != nil {
		return fmt.Errorf(i18n.T("failed to set active organization: %w"), err)
	}

	auth.SetInitialRegionOfOrg()

	fmt.Fprintf(utils.IO.Out, i18n.T("%s Successfully set organization '%s' as the active organization. \n"),
		utils.CS.Green("✓"),
		utils.CS.Bold(targetOrg.Name))

	time.Sleep(1 * time.Second)

	fmt.Fprint(utils.IO.Out, heredoc.Docf(i18n.T(`

		To list the projects of active organizations :
			%s
	`), "$ wso2-integration-platform list projects"))
	return nil

}
