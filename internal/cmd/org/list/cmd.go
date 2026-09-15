package list

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var orgListParams OrgListOptions

var OrgListCommand = &cobra.Command{
	Use:     "organizations [flags]",
	Aliases: []string{"org", "organization", "orgs"},
	Short:   i18n.T("list organizations"),
	PreRun:  common.VerifyIsUserLoggedIn,
	Long: heredoc.Docf(i18n.T(`
		List all organizations you are a member of.

		To select an organization:
			%s
	`), utils.CS.Bold("$ wso2-integration-platform change-org")),
	Example: heredoc.Docf(i18n.T(`

		To get the output in JSON format :
			%s

		To save the output to a file :
			%s
	`),
		"$ wso2-integration-platform list organizations --output=json",
		"$ wso2-integration-platform list organizations --output=json > orgs.json",
	),
	Run: func(cmd *cobra.Command, args []string) {
		err := handleOrgList(&orgListParams)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOutputFlag(OrgListCommand.Flags(), &orgListParams.outputFlag)
	common.AddGenericHelper(OrgListCommand)
}
