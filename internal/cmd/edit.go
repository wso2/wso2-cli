package cmd

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: i18n.T("edit a resource declaration"),
	Long:  i18n.T("This command allows you to edit a resource definition in the Integration Platform"),
}

func init() {
	common.AddGenericHelper(editCmd)

}
