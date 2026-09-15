package logs

import (
	"github.com/spf13/cobra"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var bLogOpts = &BuildLogFlags{}

var BuildLogsCmd = &cobra.Command{
	Use:    "build [flags]",
	Short:  "display build logs",
	Long:   "Display build logs for specific build run of a component",
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if err := printBuildLog(bLogOpts); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(BuildLogsCmd.Flags(), &bLogOpts.org)
	common.AddProjectFlag(BuildLogsCmd.Flags(), &bLogOpts.project)
	common.AddComponentFlag(BuildLogsCmd.Flags(), &bLogOpts.component)
	common.AddDeploymentTrackFlag(BuildLogsCmd.Flags(), &bLogOpts.deploymentTrack)
	common.AddRunIdFlag(BuildLogsCmd.Flags(), &bLogOpts.runId)
	BuildLogsCmd.Flags().StringVar(&bLogOpts.step, "step", "", "Build step for the logs")
}
