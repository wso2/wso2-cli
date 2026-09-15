package cmd

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	deployment "github.com/wso2/integration-platform-tools/internal/cmd/deployment/resume"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: i18n.T("Resume a resource"),
	Long:  i18n.T("Resume a suspended resource"),
}

func init() {
	resumeCmd.AddCommand(deployment.RedeployCmd)
	common.AddGenericHelper(resumeCmd)
}
