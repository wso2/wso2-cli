package cmd

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	build "github.com/wso2/integration-platform-tools/internal/cmd/build/describe"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	cmpDescribe "github.com/wso2/integration-platform-tools/internal/cmd/component/describe"
	configDescribe "github.com/wso2/integration-platform-tools/internal/cmd/config/describe"
	connectionDescribe "github.com/wso2/integration-platform-tools/internal/cmd/connection/describe"
	deploymentDescribe "github.com/wso2/integration-platform-tools/internal/cmd/deployment/describe"
	execution "github.com/wso2/integration-platform-tools/internal/cmd/executions/describe"
	project "github.com/wso2/integration-platform-tools/internal/cmd/project/describe"
)

var describeCmd = &cobra.Command{
	Use:     "describe",
	Aliases: []string{"desc"},
	Short:   i18n.T("get detailed information about a resource"),
	Long:    i18n.T(`Get detailed information about a resource in the Integration Platform.`),
}

func init() {
	describeCmd.AddCommand(cmpDescribe.ComponentDescribeCommand)
	describeCmd.AddCommand(project.ProjectDescribeCmd)
	describeCmd.AddCommand(configDescribe.ConfigDescribeCommand)
	describeCmd.AddCommand(connectionDescribe.ConnectionDescribeCommand)
	describeCmd.AddCommand(build.BuildDDescribeCommand)
	describeCmd.AddCommand(execution.DescribeExecCmd)
	describeCmd.AddCommand(deploymentDescribe.DeploymentDescribeCommand)
	common.AddGenericHelper(describeCmd)
}
