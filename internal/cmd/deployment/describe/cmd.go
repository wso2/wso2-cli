package describe

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var describeParams DescribeDeploymentParams

var DeploymentDescribeCommand = &cobra.Command{
	Use:   "deployment [flags]",
	Short: i18n.T("describe a deployment"),
	Long:  i18n.T("Get detailed information, component deployment status and deployed endpoints/URLs."),
	Example: heredoc.Docf(i18n.T(`

		To view details of a particular deployment :
			%s

		To get the output in JSON format:
			%s

		To save the output to a file:
			%s
	`),
		"$ wso2-integration-platform describe deployment --project=<project-name> --component=<component-name> --env=<env-name>",
		"$ wso2-integration-platform describe deployment --project=<project-name> --component=<component-name> --env=<env-name> --output=json",
		"$ wso2-integration-platform describe deployment --project=<project-name> --component=<component-name> --env=<env-name> --output=json > deployment.json",
	),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		err := HandleDescribeDeployment(&describeParams)
		if err != nil {
			utils.HandleErr(err)
		}
	},
	Args: cobra.MaximumNArgs(0),
}

func init() {
	common.AddOrgFlag(DeploymentDescribeCommand.Flags(), &describeParams.orgFlag)
	common.AddProjectFlag(DeploymentDescribeCommand.Flags(), &describeParams.projectFlag)
	common.AddComponentFlag(DeploymentDescribeCommand.Flags(), &describeParams.componentFlag)
	common.AddEnvFlag(DeploymentDescribeCommand.Flags(), &describeParams.envFlag)
	common.AddDeploymentTrackFlag(DeploymentDescribeCommand.Flags(), &describeParams.deploymentTrackFlag)
	common.AddOutputFlag(DeploymentDescribeCommand.Flags(), &describeParams.outputFlag)
	common.AddGenericHelper(DeploymentDescribeCommand)
}
