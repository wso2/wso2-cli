package describe

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var DescribeParams DescribeConfigParams

var ConfigDescribeCommand = &cobra.Command{
	Use:   "config [flags]",
	Short: i18n.T("describe a config"),
	Long:  i18n.T("Get information about a specific config-map or secret of a component"),
	Example: heredoc.Docf(i18n.T(`

		To view details of a particular config :
			%s

		To get the output in JSON format:
			%s

		To save the output to a file:
			%s
	`),
		"$ wso2-integration-platform describe config --project=<project-name> --component=<component-name> --env=<env-name> --deployment-track=<branch> --name=<config-name>",
		"$ wso2-integration-platform describe config --output=json --project=<project-name> --component=<component-name> --env=<env-name> --deployment-track=<branch> --name=<config-name>",
		"$ wso2-integration-platform describe config --output=json --project=<project-name> --component=<component-name> --env=<env-name> --deployment-track=<branch> --name=<config-name> > config.json",
	),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		err := handleDescribeConfig(&DescribeParams)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ConfigDescribeCommand.Flags(), &DescribeParams.orgFlag)
	common.AddProjectFlag(ConfigDescribeCommand.Flags(), &DescribeParams.projectFlag)
	common.AddComponentFlag(ConfigDescribeCommand.Flags(), &DescribeParams.componentFlag)
	common.AddEnvFlag(ConfigDescribeCommand.Flags(), &DescribeParams.envFlag)
	common.AddDeploymentTrackFlag(ConfigDescribeCommand.Flags(), &DescribeParams.deploymentTrackFlag)
	ConfigDescribeCommand.Flags().StringVarP(&DescribeParams.nameFlag, "name", "n", "", "name of the config")
	common.AddOutputFlag(ConfigDescribeCommand.Flags(), &DescribeParams.outputFlag)
	common.AddGenericHelper(ConfigDescribeCommand)
}
