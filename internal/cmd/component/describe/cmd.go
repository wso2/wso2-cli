package describe

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var describeParams DescribeComponentParams

var ComponentDescribeCommand = &cobra.Command{
	Use:   "component [component-name] [flags]",
	Short: i18n.T("describe a component"),
	Long:  i18n.T("Get detailed information, including build, deployment, and endpoint details for a component in your project."),
	Example: heredoc.Docf(i18n.T(`

		To view details of a particular component :
			%s

		To get the output in JSON format:
			%s

		To save the output to a file:
			%s
	`),
		"$ wso2-integration-platform describe component <component-name> --project=<project-name>",
		"$ wso2-integration-platform describe component <component-name> --project=<project-name> --output=json",
		"$ wso2-integration-platform describe component <component-name> --project=<project-name> --output=json > component.json",
	),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			describeParams.componentFlag = args[0]
		}

		err := HandleDescribeComponent(&describeParams)
		if err != nil {
			utils.HandleErr(err)
		}
	},
	Args: cobra.MaximumNArgs(1),
}

func init() {
	common.AddOrgFlag(ComponentDescribeCommand.Flags(), &describeParams.orgFlag)
	common.AddProjectFlag(ComponentDescribeCommand.Flags(), &describeParams.projectFlag)
	common.AddOutputFlag(ComponentDescribeCommand.Flags(), &describeParams.outputFlag)
	common.AddGenericHelper(ComponentDescribeCommand)
}
