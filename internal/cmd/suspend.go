package cmd

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	deployment "github.com/wso2/integration-platform-tools/internal/cmd/deployment/suspend"
)

var suspendCmd = &cobra.Command{
	Use:   "suspend",
	Short: i18n.T("Suspend a resource"),
	Long:  i18n.T("Suspend a resource"),
}

func init() {
	suspendCmd.AddCommand(deployment.SuspendCmd)
	common.AddGenericHelper(suspendCmd)
}
