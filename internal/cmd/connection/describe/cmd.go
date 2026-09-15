package describe

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var DescribeParams DescribeConnectionsParams

var ConnectionDescribeCommand = &cobra.Command{
	Use:    "connection [flags]",
	Short:  i18n.T("describe a connection"),
	Long:   i18n.T("Get information about a specific connection within a project"),
	PreRun: common.VerifyIsUserLoggedIn,
	Example: heredoc.Docf(i18n.T(`

		To view details of a particular connection :
			%s

		To get the output in JSON format:
			%s

		To save the output to a file:
			%s
	`),
		"$ wso2-integration-platform describe connection --project=<project-name> --component=<component-name> --name=<connection-name>",
		"$ wso2-integration-platform describe connection --project=<project-name> --component=<component-name> --name=<connection-name> --output=json",
		"$ wso2-integration-platform describe connection --project=<project-name> --component=<component-name> --name=<connection-name> --output=json > connection.json",
	),
	Run: func(cmd *cobra.Command, args []string) {
		err := handleDescribeConnection(&DescribeParams)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ConnectionDescribeCommand.Flags(), &DescribeParams.orgFlag)
	common.AddProjectFlag(ConnectionDescribeCommand.Flags(), &DescribeParams.projectFlag)
	common.AddComponentFlag(ConnectionDescribeCommand.Flags(), &DescribeParams.componentFlag)
	ConnectionDescribeCommand.Flags().StringVarP(&DescribeParams.nameFlag, "name", "n", "", "name of the connection")
	common.AddOutputFlag(ConnectionDescribeCommand.Flags(), &DescribeParams.outputFlag)
	common.AddGenericHelper(ConnectionDescribeCommand)
}
