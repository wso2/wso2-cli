package list

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var opts = &ListExecutionParams{}

var ListExecCmd = &cobra.Command{
	Use:     "execution [flags]",
	Aliases: []string{"executions"},
	Short:   "list executions",
	Long:    "List executions of a sheduled/manual task",
	Example: heredoc.Docf(i18n.T(`

		To list executions of a component :
			%s

		To get the output in JSON format :
			%s

		To save the output to a file :
			%s

	`),
		"$ wso2-integration-platform list executions --component=<component-name> --project=<project-name>",
		"$ wso2-integration-platform list executions --component=<component-name> --project=<project-name> --output=json",
		"$ wso2-integration-platform list executions --component=<component-name> --project=<project-name> --output=json > executions.json"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleImpl(opts); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ListExecCmd.Flags(), &opts.Org)
	common.AddProjectFlag(ListExecCmd.Flags(), &opts.Project)
	common.AddComponentFlag(ListExecCmd.Flags(), &opts.Component)
	common.AddEnvFlag(ListExecCmd.Flags(), &opts.Env)
	common.AddDeploymentTrackFlag(ListExecCmd.Flags(), &opts.DeploymentTrack)
	ListExecCmd.Flags().IntVar(&opts.Limit, "limit", 50, "limit the number of executions")
	common.AddOutputFlag(ListExecCmd.Flags(), &opts.OutputFlag)
	common.AddGenericHelper(ListExecCmd)
}
