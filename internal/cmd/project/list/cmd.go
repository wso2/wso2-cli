package list

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var ListParams ListProjectsParams

var ProjectListCommand = &cobra.Command{
	Use:     "projects [flags]",
	Aliases: []string{"project"},
	Short:   i18n.T("list projects"),
	Long:    i18n.T("List all projects within an organization."),
	Example: heredoc.Docf(i18n.T(`

		To list projects of an organization :
			%s

		To get the output in JSON format :
			%s

		To save the output to a file :
			%s
	`),
		"$ wso2-integration-platform list projects --org=<org-name>",
		"$ wso2-integration-platform list projects --output=json",
		"$ wso2-integration-platform list projects --output=json > projects.json",
	),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {

		err := handleListProjects(&ListParams)

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ProjectListCommand.Flags(), &ListParams.orgFlag)
	common.AddOutputFlag(ProjectListCommand.Flags(), &ListParams.outputFlag)
	common.AddGenericHelper(ProjectListCommand)
}
