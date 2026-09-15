package logs

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var opts = &ExecutionLogsOpts{}

var ExecutionLogsCmd = &cobra.Command{
	Use:     "executions",
	Aliases: []string{"execution"},
	Short:   i18n.T("display execution logs"),
	Long:    i18n.T("Display exection logs of a scheduled/manual task"),
	PreRun:  common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleCmd(opts); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ExecutionLogsCmd.Flags(), &opts.Org)
	common.AddProjectFlag(ExecutionLogsCmd.Flags(), &opts.Project)
	common.AddComponentFlag(ExecutionLogsCmd.Flags(), &opts.Component)
	common.AddDeploymentTrackFlag(ExecutionLogsCmd.Flags(), &opts.DeploymentTrack)
	ExecutionLogsCmd.Flags().StringVar(&opts.ExecutionId, "id", "", "execution id of the target execution")
	ExecutionLogsCmd.Flags().IntVar(&opts.AttemptNumber, "attempt", 0, "attempt number of the target execution")
	common.AddEnvFlag(ExecutionLogsCmd.Flags(), &opts.Env)
}
