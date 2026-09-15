package create

import (
	"github.com/spf13/cobra"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var opts = &CreateExecutionParams{}

var CreateExecCmd = &cobra.Command{
	Use:    "execution [flags]",
	Short:  "Trigger an execution",
	Long:   "Trigger an execution of a sheduled/manual task",
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleImpl(opts); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(CreateExecCmd.Flags(), &opts.Org)
	common.AddProjectFlag(CreateExecCmd.Flags(), &opts.Project)
	common.AddComponentFlag(CreateExecCmd.Flags(), &opts.Component)
	common.AddEnvFlag(CreateExecCmd.Flags(), &opts.Env)
	common.AddDeploymentTrackFlag(CreateExecCmd.Flags(), &opts.DeploymentTrack)
	common.AddGenericHelper(CreateExecCmd)
}
