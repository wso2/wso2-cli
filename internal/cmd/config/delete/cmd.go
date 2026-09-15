package delete

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var params ConfigDeleteOptions

var ConfigDeleteCommand = &cobra.Command{
	Use:   "config [config-name] [flags]",
	Short: i18n.T("delete config or secret"),
	Long:  heredoc.Doc(i18n.T(`Delete a config or secret within your component`)),
	Example: heredoc.Docf(i18n.T(`

		To delete a config :
			%s
			
	`),
		`$ wso2-integration-platform delete config <config-name> --project=<project-name> --component=<component-name> --org=<org-name> --env=<env-name>`),
	Args:   cobra.MaximumNArgs(1),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			params.nameFlag = args[0]
		}

		err := handleConfigDeleteCommand()

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ConfigDeleteCommand.Flags(), &params.orgFlag)
	common.AddProjectFlag(ConfigDeleteCommand.Flags(), &params.projectFlag)
	common.AddComponentFlag(ConfigDeleteCommand.Flags(), &params.componentFlag)
	common.AddEnvFlag(ConfigDeleteCommand.Flags(), &params.envFlag)
	common.AddDeploymentTrackFlag(ConfigDeleteCommand.Flags(), &params.deploymentTrackFlag)
	ConfigDeleteCommand.Flags().BoolVarP(&params.skipConfirm, "skip-confirm", "", false, "Skip confirmation prompt before deleting the config")
}
