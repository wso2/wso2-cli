package suspend

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var suspendOpts = &SuspendCmdOpts{}

var SuspendCmd = &cobra.Command{
	Use:    "deployment [component-name] [flags]",
	Short:  i18n.T("Suspend a deployment"),
	Long:   i18n.T("Suspend a deployment that is deployed"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleDeploymentSuspend(suspendOpts, args); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(SuspendCmd.Flags(), &suspendOpts.Org)
	common.AddProjectFlag(SuspendCmd.Flags(), &suspendOpts.Project)
	common.AddEnvFlag(SuspendCmd.Flags(), &suspendOpts.Env)
	common.AddDeploymentTrackFlag(SuspendCmd.Flags(), &suspendOpts.DeploymentTrack)
}
