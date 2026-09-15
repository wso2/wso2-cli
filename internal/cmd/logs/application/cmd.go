package application

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var opts = &ApplicationLogsOpts{}

var ApplicationLogsCmd = &cobra.Command{
	Use:    "application [flags]",
	Short:  i18n.T("display application logs"),
	Long:   i18n.T("Display application logs of a deployed component"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if err := printApplicationLogs(opts); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ApplicationLogsCmd.Flags(), &opts.org)
	common.AddProjectFlag(ApplicationLogsCmd.Flags(), &opts.project)
	common.AddComponentFlag(ApplicationLogsCmd.Flags(), &opts.component)
	common.AddDeploymentTrackFlag(ApplicationLogsCmd.Flags(), &opts.deploymentTrack)
	common.AddEnvFlag(ApplicationLogsCmd.Flags(), &opts.env)

	ApplicationLogsCmd.Flags().BoolVarP(&opts.followFlag, "follow", "f", false, i18n.T("watch for additional logs output."))
	ApplicationLogsCmd.Flags().StringVarP(&opts.queryFlag, "query", "q", "", i18n.T("filter logs against a search query"))
	ApplicationLogsCmd.Flags().UintVarP(&opts.limitFlag, "limit", "l", 100, i18n.T("limit the number of logs"))
}
