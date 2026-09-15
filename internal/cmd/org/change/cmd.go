package change

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var changeOrgParam ChangeOrgParams

var ChangeOrgCommand = &cobra.Command{
	Use:    "change-org [flags]",
	Short:  i18n.T("switch the active organization"),
	Long:   i18n.T("Switch the active organization for your local context"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		err := handleOrgChange(&changeOrgParam)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	ChangeOrgCommand.Flags().StringVarP(&changeOrgParam.orgFlag, "org", "o", "", i18n.T("name or handle of the organization to select"))
	common.AddGenericHelper(ChangeOrgCommand)
}
