package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	cmpDel "github.com/wso2/integration-platform-tools/internal/cmd/component/delete"
	configDel "github.com/wso2/integration-platform-tools/internal/cmd/config/delete"
	projDel "github.com/wso2/integration-platform-tools/internal/cmd/project/delete"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a resource",
	Long:  `This command allows you to delete a resource in the Integration Platform`,
}

func init() {
	deleteCmd.AddCommand(cmpDel.ComponentDeleteCmd)
	deleteCmd.AddCommand(projDel.DeleteCmd)
	deleteCmd.AddCommand(configDel.ConfigDeleteCommand)
	common.AddGenericHelper(deleteCmd)
}
