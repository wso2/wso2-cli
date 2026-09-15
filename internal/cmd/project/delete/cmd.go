package delete

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var projectDeleteOpts = &ProjectDeleteOpts{}

var DeleteCmd = &cobra.Command{
	Use:     "project [project-name] [flags]",
	Short:   i18n.T("delete a project"),
	Long:    i18n.T(`This command allows you to delete a project.`),
	Args:    cobra.MaximumNArgs(1),
	Example: `wso2-integration-platform delete project <project-name>`,
	PreRun:  common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) > 0 {
			projectDeleteOpts.Project = args[0]
		}

		err := HandleDeleteProject(projectDeleteOpts)

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(DeleteCmd.Flags(), &projectDeleteOpts.Org)
	DeleteCmd.Flags().BoolVarP(&projectDeleteOpts.Force, "force", "f", false, "Force delete the project")
}
