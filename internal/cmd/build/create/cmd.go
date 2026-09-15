package create

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var buildParams BuildComponentParams

var ComponentBuildCommand = &cobra.Command{
	Use:   "build [component-name] [flags]",
	Short: i18n.T("Trigger a build for a component"),
	Args:  cobra.MaximumNArgs(1),
	Long:  i18n.T("Trigger a build for a component in a selected deployment track within your project."),
	Example: heredoc.Docf(`
	
		To build a component from your default project :
			%s
	`,
		"$ wso2-integration-platform create build <component-name> --project='Default Project' --deployment-track=main"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			buildParams.ComponentFlag = args[0]
		}

		err := HandleBuildComponent(&buildParams)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ComponentBuildCommand.Flags(), &buildParams.OrgFlag)
	common.AddProjectFlag(ComponentBuildCommand.Flags(), &buildParams.ProjectFlag)
	common.AddDeploymentTrackFlag(ComponentBuildCommand.Flags(), &buildParams.DeploymentTrackFlag)
	common.AddGenericHelper(ComponentBuildCommand)
}
