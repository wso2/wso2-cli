package list

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var params CmpListOptions

var ComponentListCommand = &cobra.Command{
	Use:     "integrations [flags]",
	Aliases: []string{"integration"},
	Short:   i18n.T("list integrations"),
	Long:    i18n.T("List integrations in your project"),
	Example: heredoc.Docf(i18n.T(`

		To list the integrations within a project :
			%s

		To get the output in JSON format :
			%s

		To save the output to a file :
			%s
	`),
		"$ wso2-integration-platform list integrations --project=<project-name>",
		"$ wso2-integration-platform list integrations --project=<project-name> --output=json",
		"$ wso2-integration-platform list integrations --project=<project-name> --output=json > integrations.json"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		err := HandleComponentListCommand(&params)

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ComponentListCommand.Flags(), &params.OrgFlag)
	common.AddProjectFlag(ComponentListCommand.Flags(), &params.ProjectFlag)
	common.AddOutputFlag(ComponentListCommand.Flags(), &params.OutputFlag)
	common.AddGenericHelper(ComponentListCommand)
}
