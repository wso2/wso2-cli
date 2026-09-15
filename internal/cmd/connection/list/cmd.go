package list

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var params ConfigListOptions

var ConnectionListCommand = &cobra.Command{
	Use:     "connections [flags]",
	Aliases: []string{"connection"},
	Short:   i18n.T("list connections"),
	Long:    heredoc.Doc(i18n.T(`List connections within a project`)),
	PreRun:  common.VerifyIsUserLoggedIn,
	Example: heredoc.Docf(i18n.T(`

		To list connections within a project :
			%s

		To get the output in JSON format :
			%s

		To save the output to a file :
			%s

	`),
		"$ wso2-integration-platform list connections --project=<project-name>",
		"$ wso2-integration-platform list connections --project=<project-name> --output=json",
		"$ wso2-integration-platform list connections --project=<project-name> --output=json > connections.json"),
	Run: func(cmd *cobra.Command, args []string) {
		err := handleConnectionListCommand()

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ConnectionListCommand.Flags(), &params.orgFlag)
	common.AddProjectFlag(ConnectionListCommand.Flags(), &params.projectFlag)
	common.AddComponentFlag(ConnectionListCommand.Flags(), &params.componentFlag)
	common.AddOutputFlag(ConnectionListCommand.Flags(), &params.outputFlag)
	common.AddGenericHelper(ConnectionListCommand)
}
