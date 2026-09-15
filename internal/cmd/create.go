package cmd

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	componentBuild "github.com/wso2/integration-platform-tools/internal/cmd/build/create"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	componentCreate "github.com/wso2/integration-platform-tools/internal/cmd/component/create"
	configCreate "github.com/wso2/integration-platform-tools/internal/cmd/config/create"
	connectionCreate "github.com/wso2/integration-platform-tools/internal/cmd/connection/create"
	componentDeploy "github.com/wso2/integration-platform-tools/internal/cmd/deployment/create"
	execution "github.com/wso2/integration-platform-tools/internal/cmd/executions/create"
	projectCreate "github.com/wso2/integration-platform-tools/internal/cmd/project/create"
	getToken "github.com/wso2/integration-platform-tools/internal/cmd/test-token/create"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: i18n.T("create a new resource"),
	Long:  i18n.T(`This command allows you to create a resource in the Integration Platform.`),
}

func init() {
	createCmd.AddCommand(componentCreate.ComponentCreateCommand)
	createCmd.AddCommand(projectCreate.ProjectCreateCommand)
	createCmd.AddCommand(configCreate.ConfigCreateCommand)
	createCmd.AddCommand(connectionCreate.ConnectionCreateCommand)
	createCmd.AddCommand(componentBuild.ComponentBuildCommand)
	createCmd.AddCommand(componentDeploy.ComponentDeployCommand)
	createCmd.AddCommand(getToken.ComponentTokenCommand)
	createCmd.AddCommand(execution.CreateExecCmd)
	common.AddGenericHelper(createCmd)
}
