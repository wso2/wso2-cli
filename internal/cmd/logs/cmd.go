package logs

import (
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	buildLog "github.com/wso2/integration-platform-tools/internal/cmd/build/logs"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	execution "github.com/wso2/integration-platform-tools/internal/cmd/executions/logs"
	"github.com/wso2/integration-platform-tools/internal/cmd/logs/application"
	"github.com/wso2/integration-platform-tools/internal/cmd/logs/gateway"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

const (
	ProjectLog     = "project"
	BuildLog       = "build"
	ApplicationLog = "application"
	GatewayLog     = "gateway"
)

var LogTypes = []string{ApplicationLog, GatewayLog, BuildLog}

var logsParams LogsParams

var LogsCommand = &cobra.Command{
	Use:     "logs [command]",
	Aliases: []string{"log"},
	Short:   i18n.T("display logs"),
	Long:    i18n.T("Display logs for different resources in the Integration Platform."),
	Run: func(cmd *cobra.Command, args []string) {
		err := HandleComponentLogs(&logsParams)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	LogsCommand.AddCommand(buildLog.BuildLogsCmd)
	LogsCommand.AddCommand(application.ApplicationLogsCmd)
	LogsCommand.AddCommand(gateway.GatewayLogsCmd)
	LogsCommand.AddCommand(execution.ExecutionLogsCmd)
	common.AddGenericHelper(LogsCommand)
}
