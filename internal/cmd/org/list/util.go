package list

import (
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

// Renders the organization list in JSON format
func printOrgListStructured(orgs []api.Organization, format common.OutputFormat) error {
	// Normalize a nil slice to an empty slice so JSON renders `[]` not `null`.
	if orgs == nil {
		orgs = []api.Organization{}
	}

	rendered, err := common.RenderStructured(format, orgs)
	if err != nil {
		return err
	}

	// The JSON payload always goes to stdout — even when it is just `[]`.
	fmt.Fprintln(utils.IO.Out, rendered)

	// On an empty result, guide the user via stderr
	if len(orgs) == 0 {
		fmt.Fprintln(utils.IO.ErrOut, i18n.T("No organizations found."))
	}
	return nil
}

func printOrgList(selectedOrg api.Organization, orgs []api.Organization, currentUser api.UserInfo) {
	data := [][]string{}
	for _, org := range orgs {
		selected := ""
		if selectedOrg.ID == org.ID {
			selected = " (active)"
		}
		invited := ""
		if org.Owner.IDPId != currentUser.IDPId {
			invited = utils.CS.Gray(" (invited)")
		}
		data = append(data, []string{org.Name + selected,
			org.Handle + invited})
	}

	tableStr := utils.CreateTable(data, []string{"NAME", "HANDLE"}, "")
	fmt.Println(tableStr)
}
