package gateway

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var opts = &GatewayLogOpts{}

var GatewayLogsCmd = &cobra.Command{
	Use:    "gateway [flags]",
	Short:  i18n.T("display gateway logs"),
	Long:   i18n.T("Display gateway logs of a deployed component"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if err := printGatewayLogs(opts); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(GatewayLogsCmd.Flags(), &opts.org)
	common.AddProjectFlag(GatewayLogsCmd.Flags(), &opts.project)
	common.AddComponentFlag(GatewayLogsCmd.Flags(), &opts.component)
	common.AddDeploymentTrackFlag(GatewayLogsCmd.Flags(), &opts.deploymentTrack)
	common.AddEnvFlag(GatewayLogsCmd.Flags(), &opts.env)

	GatewayLogsCmd.Flags().BoolVarP(&opts.followFlag, "follow", "f", false, i18n.T("watch for additional logs output."))
	GatewayLogsCmd.Flags().StringVarP(&opts.queryFlag, "query", "q", "", i18n.T("filter logs against a search query"))
	GatewayLogsCmd.Flags().UintVarP(&opts.limitFlag, "limit", "l", 100, i18n.T("limit the number of logs"))
}
