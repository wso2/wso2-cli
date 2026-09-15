package describe

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var opts = &DescribeExecutionParams{}

var DescribeExecCmd = &cobra.Command{
	Use:     "execution [flags]",
	Aliases: []string{"executions"},
	Short:   "describe execution",
	Long:    "Describe an execution of a sheduled/manual task",
	PreRun:  common.VerifyIsUserLoggedIn,
	Example: heredoc.Docf(i18n.T(`

		To describe an execution :
			%s

		To get the output in JSON format:
			%s

		To save the output to a file:
			%s
	`),
		"$ wso2-integration-platform describe execution --id=<execution-id> --project=<project-name> --component=<component-name> --env=<env-name>",
		"$ wso2-integration-platform describe execution --id=<execution-id> --project=<project-name> --component=<component-name> --env=<env-name> --output=json",
		"$ wso2-integration-platform describe execution --id=<execution-id> --project=<project-name> --component=<component-name> --env=<env-name> --output=json > execution.json",
	),
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleDescribeImpl(opts); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(DescribeExecCmd.Flags(), &opts.Org)
	common.AddProjectFlag(DescribeExecCmd.Flags(), &opts.Project)
	common.AddComponentFlag(DescribeExecCmd.Flags(), &opts.Component)
	common.AddEnvFlag(DescribeExecCmd.Flags(), &opts.Env)
	DescribeExecCmd.Flags().StringVar(&opts.ExecutionId, "id", "", "execution id of the target execution")
	common.AddOutputFlag(DescribeExecCmd.Flags(), &opts.Output)
	common.AddGenericHelper(DescribeExecCmd)
}
