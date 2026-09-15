package list

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var params ConfigListOptions

var ConfigListCommand = &cobra.Command{
	Use:     "configs [flags]",
	Aliases: []string{"config"},
	Short:   i18n.T("list configs"),
	Long:    heredoc.Doc(i18n.T(`List config-maps and secrets within a component`)),
	Example: heredoc.Docf(i18n.T(`

		To list the configs within a component :
			%s

		To get the output in JSON format :
			%s

		To save the output to a file :
			%s

	`),
		"$ wso2-integration-platform list configs --project=<project-name> --component=<component-name> --env=<env-name> --deployment-track=<branch>",
		"$ wso2-integration-platform list configs --project=<project-name> --component=<component-name> --env=<env-name> --deployment-track=<branch> --output=json",
		"$ wso2-integration-platform list configs --project=<project-name> --component=<component-name> --env=<env-name> --deployment-track=<branch> --output=json > configs.json"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		err := handleConfigListCommand()

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ConfigListCommand.Flags(), &params.orgFlag)
	common.AddProjectFlag(ConfigListCommand.Flags(), &params.projectFlag)
	common.AddComponentFlag(ConfigListCommand.Flags(), &params.componentFlag)
	common.AddEnvFlag(ConfigListCommand.Flags(), &params.envFlag)
	common.AddDeploymentTrackFlag(ConfigListCommand.Flags(), &params.deploymentTrackFlag)
	common.AddOutputFlag(ConfigListCommand.Flags(), &params.outputFlag)
	common.AddGenericHelper(ConfigListCommand)
}
