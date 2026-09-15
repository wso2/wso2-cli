package describe

import (
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var params BuildDescribeParams

var BuildDDescribeCommand = &cobra.Command{
	Use:   "build [build-id] [flags]",
	Short: i18n.T("Describe build details"),
	Args:  cobra.MaximumNArgs(1),
	Long:  i18n.T("View details such as the build status"),
	Example: heredoc.Docf(i18n.T(`

		To describe a build:
			%s

		To get the output in JSON format:
			%s

		To save the output to a file:
			%s
	`),
		"$ wso2-integration-platform describe build <build-id>",
		"$ wso2-integration-platform describe build <build-id> --output=json",
		"$ wso2-integration-platform describe build <build-id> --output=json > build.json",
	),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			runId, err := strconv.Atoi(args[0])
			if err != nil {
				utils.HandleErr(err)
			}

			params.RunIdFlag = runId
		}

		err := HandleDescribeBuild(&params)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(BuildDDescribeCommand.Flags(), &params.OrgFlag)
	common.AddProjectFlag(BuildDDescribeCommand.Flags(), &params.ProjectFlag)
	common.AddComponentFlag(BuildDDescribeCommand.Flags(), &params.ComponentFlag)
	common.AddDeploymentTrackFlag(BuildDDescribeCommand.Flags(), &params.DeploymentTrackFlag)
	common.AddOutputFlag(BuildDDescribeCommand.Flags(), &params.OutputFlag)
	common.AddGenericHelper(BuildDDescribeCommand)
}
