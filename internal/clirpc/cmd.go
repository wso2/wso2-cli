package clirpc

import (
	"github.com/spf13/cobra"
	// load the project handlers
	_ "github.com/wso2/integration-platform-tools/internal/clirpc/api/apim"
	_ "github.com/wso2/integration-platform-tools/internal/clirpc/api/auth"
	_ "github.com/wso2/integration-platform-tools/internal/clirpc/api/build"
	_ "github.com/wso2/integration-platform-tools/internal/clirpc/api/component"
	_ "github.com/wso2/integration-platform-tools/internal/clirpc/api/connections"
	_ "github.com/wso2/integration-platform-tools/internal/clirpc/api/deployment"
	_ "github.com/wso2/integration-platform-tools/internal/clirpc/api/project"
	_ "github.com/wso2/integration-platform-tools/internal/clirpc/api/repo"
	"github.com/wso2/integration-platform-tools/internal/clirpc/server"
)

var version string = "DEV"

var cmdOpts = struct {
	method string
	params string
}{}

var RPCCmd = &cobra.Command{
	Use: "start-rpc-server",
	Run: func(cmd *cobra.Command, args []string) {
		// Start the RPC server
		if v := cmd.Parent().Version; v != "" {
			version = v
		}

		if cmdOpts.method != "" {
			server.RunSingleCmd(cmdOpts.method, cmdOpts.params)
			return
		} else {
			server.StartRPCServer(version)
		}
	},
	Hidden: true,
}

func init() {
	RPCCmd.Flags().StringVar(&cmdOpts.method, "method", "", "Method to call")
	RPCCmd.Flags().StringVar(&cmdOpts.params, "params", "", "Parameters to pass to the method")
	RPCCmd.Flags().MarkHidden("method")
	RPCCmd.Flags().MarkHidden("params")
}
