package create

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var createParams CreateProjectParams

var ProjectCreateCommand = &cobra.Command{
	Use:   "project <project-name> [flags]",
	Args:  cobra.MaximumNArgs(1),
	Short: i18n.T("create a project"),
	Long:  i18n.T("Create a new project in the Integration Platform."),
	Example: heredoc.Docf(i18n.T(`
	
		To create a project :
			%s
	`),
		"$ wso2-integration-platform create project <project-name> --description <project-description>"),

	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			createParams.Name = args[0]
		}

		err := HandleProjectCreate(&createParams)

		if err != nil {
			utils.HandleErr(err)
		} else {
			utils.PrintInfo("%s", heredoc.Docf(i18n.T(`
			
			To create a new component within your newly created project :
				%s	
		`),
				fmt.Sprintf("$ wso2-integration-platform create component <component-name> --project=%s", createParams.Name),
			))
		}
	},
}

func init() {
	common.AddOrgFlag(ProjectCreateCommand.Flags(), &createParams.OrgFlag)
	common.AddDescriptionFlag(ProjectCreateCommand.Flags(), &createParams.Description)
	common.AddGenericHelper(ProjectCreateCommand)
}
