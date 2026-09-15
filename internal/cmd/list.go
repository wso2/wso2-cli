package cmd

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	buildList "github.com/wso2/integration-platform-tools/internal/cmd/build/list"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	componentList "github.com/wso2/integration-platform-tools/internal/cmd/component/list"
	configList "github.com/wso2/integration-platform-tools/internal/cmd/config/list"
	connectionList "github.com/wso2/integration-platform-tools/internal/cmd/connection/list"
	execution "github.com/wso2/integration-platform-tools/internal/cmd/executions/list"
	orgList "github.com/wso2/integration-platform-tools/internal/cmd/org/list"
	projectList "github.com/wso2/integration-platform-tools/internal/cmd/project/list"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   i18n.T("list available resources"),
	Long:    i18n.T(`Display a list of resources (projects, components, configs, etc.) in the Integration Platform`),
}

func init() {
	listCmd.AddCommand(orgList.OrgListCommand)
	listCmd.AddCommand(projectList.ProjectListCommand)
	listCmd.AddCommand(componentList.ComponentListCommand)
	listCmd.AddCommand(configList.ConfigListCommand)
	listCmd.AddCommand(connectionList.ConnectionListCommand)
	listCmd.AddCommand(buildList.BuildListCommand)
	listCmd.AddCommand(execution.ListExecCmd)
	common.AddGenericHelper(listCmd)
}
