package describe

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var projDescFlags ProjectDescribeFlags

var ProjectDescribeCmd = &cobra.Command{
	Use:   "project [project-name] [flags]",
	Short: i18n.T("Describe a project"),
	Long:  i18n.T(`This command allows you to describe a project.`),
	Example: heredoc.Docf(i18n.T(`

		To describe a project in the current organization:
			%s

		To get the output in JSON format:
			%s

		To save the output to a file:
			%s
	`),
		"$ wso2-integration-platform describe project <project-name>",
		"$ wso2-integration-platform describe project <project-name> --output=json",
		"$ wso2-integration-platform describe project <project-name> --output=json > project.json",
	),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			projDescFlags.Project = args[0]
		}

		if err := handleProjectDescribe(&projDescFlags); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ProjectDescribeCmd.Flags(), &projDescFlags.Org)
	common.AddOutputFlag(ProjectDescribeCmd.Flags(), &projDescFlags.outputFlag)
}
