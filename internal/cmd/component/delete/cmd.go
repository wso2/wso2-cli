package delete

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var cmpDeleteOpts = &ComponentDeleteOpts{}

var ComponentDeleteCmd = &cobra.Command{
	Use:     "component [component-name] [flags]",
	Short:   i18n.T("delete a component"),
	Long:    i18n.T(`This command allows you to delete a component in the Project`),
	Example: `wso2-integration-platform delete component <component-name>`,
	Args:    cobra.MaximumNArgs(1),
	PreRun:  common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			cmpDeleteOpts.ComponentName = args[0]
		}

		err := HandleDeleteComponent(cmpDeleteOpts)

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddProjectFlag(ComponentDeleteCmd.Flags(), &cmpDeleteOpts.Project)
	common.AddOrgFlag(ComponentDeleteCmd.Flags(), &cmpDeleteOpts.Org)
	ComponentDeleteCmd.Flags().BoolVarP(&cmpDeleteOpts.SkipConfirm, "skip-confirm", "", false, "Skip confirmation prompt before deleting the component")
}
