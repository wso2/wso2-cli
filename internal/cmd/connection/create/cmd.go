package create

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var params ConnectionCreateOptions

var ConnectionCreateCommand = &cobra.Command{
	Use:    "connection [flags]",
	Short:  i18n.T("create connection"),
	Long:   heredoc.Doc(i18n.T(`Create a new connection for a service`)),
	PreRun: common.VerifyIsUserLoggedIn,
	Example: heredoc.Docf(i18n.T(`
	
		To create a connection for a service :
			%s
			
	`),
		`$ wso2-integration-platform create connection --project=<project-name> --service=<service-name> --name=<connection-name>`),
	Run: func(cmd *cobra.Command, args []string) {

		err := handleConfigCreateCommand()

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ConnectionCreateCommand.Flags(), &params.orgFlag)
	common.AddProjectFlag(ConnectionCreateCommand.Flags(), &params.projectFlag)
	common.AddComponentFlag(ConnectionCreateCommand.Flags(), &params.componentFlag)
	ConnectionCreateCommand.Flags().StringVarP(&params.nameFlag, "service", "s", "", "name of service for which the connection needs to be created for")
	ConnectionCreateCommand.Flags().StringVarP(&params.nameFlag, "name", "n", "", "name of the new connection")
	common.AddGenericHelper(ConnectionCreateCommand)
}
