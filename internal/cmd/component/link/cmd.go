package link

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var linkParams LinkComponentParams

var ComponentLinkCommand = &cobra.Command{
	Use:   "link [integration-name] [flags]",
	Short: i18n.T("link directory to an integration"),
	Long:  i18n.T("Associate a directory with an integration by creating a .wso2/link.yaml file within the integration directory"),
	Example: heredoc.Docf(i18n.T(`
	
		To link an integration directory :
			%s

	`),
		"$ wso2-integration-platform link <integration-name> --project=<project-name>"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			linkParams.componentFlag = args[0]
		}

		err := HandleLinkComponent(&linkParams)
		if err != nil {
			utils.HandleErr(err)
		}
	},
	Args:   cobra.MaximumNArgs(1),
	Hidden: true,
}

func init() {
	common.AddOrgFlag(ComponentLinkCommand.Flags(), &linkParams.orgFlag)
	common.AddProjectFlag(ComponentLinkCommand.Flags(), &linkParams.projectFlag)
	common.AddGenericHelper(ComponentLinkCommand)
}
