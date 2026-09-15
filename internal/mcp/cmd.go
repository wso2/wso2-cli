package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/wso2/integration-platform-tools/internal/mcp/auth"
	"github.com/wso2/integration-platform-tools/internal/mcp/build"
	"github.com/wso2/integration-platform-tools/internal/mcp/buildpack"
	"github.com/wso2/integration-platform-tools/internal/mcp/configurations"
	"github.com/wso2/integration-platform-tools/internal/mcp/connection"
	"github.com/wso2/integration-platform-tools/internal/mcp/customDomain"
	"github.com/wso2/integration-platform-tools/internal/mcp/database"
	"github.com/wso2/integration-platform-tools/internal/mcp/deployment"
	"github.com/wso2/integration-platform-tools/internal/mcp/execution"
	"github.com/wso2/integration-platform-tools/internal/mcp/info"
	"github.com/wso2/integration-platform-tools/internal/mcp/integration"
	"github.com/wso2/integration-platform-tools/internal/mcp/logs"
	"github.com/wso2/integration-platform-tools/internal/mcp/marketplace"
	"github.com/wso2/integration-platform-tools/internal/mcp/org"
	"github.com/wso2/integration-platform-tools/internal/mcp/project"
	testkey "github.com/wso2/integration-platform-tools/internal/mcp/test-key"
	testuser "github.com/wso2/integration-platform-tools/internal/mcp/test-user"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	cliutils "github.com/wso2/integration-platform-tools/internal/utils"
)

var (
	httpMode bool
	port     int
)

var MCPCommand = &cobra.Command{
	Use:   "start-mcp-server",
	Short: "Start the MCP server in stdio or HTTP mode",
	Long:  "Start the Model Context Protocol (MCP) server for the Integration Platform CLI. Supports both stdio and HTTP modes.",
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewMCPServer(
			"Integration Platform MCP Server",
			"1.0.0",
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithToolCapabilities(true),
			server.WithLogging(),
			server.WithRecovery(),
		)
		org.RegisterOrgTools(s, httpMode)
		project.RegisterProjectTools(s)
		integration.RegisterIntegrationTools(s)
		build.RegisterBuildTools(s)
		execution.RegisterExecutionTools(s)
		deployment.RegisterDeploymentTools(s)
		logs.RegisterLogsTools(s)
		buildpack.RegisteBuildpackTools(s)
		configurations.RegisterConfigurationTools(s)
		testkey.RegisterOrgTools(s)
		database.RegisterDatabaseTools(s)
		connection.RegisterConnectionTools(s)
		testuser.RegisterTestUserTools(s)
		marketplace.RegisterMarketplaceTools(s)
		info.RegisterInfoTools(s)
		auth.RegisterAuthTools(s, httpMode)
		customDomain.RegisterCustomDomainTools(s)

		if httpMode {
			httpServer := server.NewStreamableHTTPServer(s, server.WithHTTPContextFunc(utils.AuthFromRequest))
			fmt.Printf("HTTP MCP server listening on :%d/mcp", port)
			err := httpServer.Start(fmt.Sprintf(":%d", port))
			if err != nil {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		} else {
			if err := server.ServeStdio(s); err != nil {
				fmt.Fprintf(cliutils.IO.ErrOut, "Stdio server error: %v\n", err)
			}
		}
	},
	Hidden: true,
}

func init() {
	MCPCommand.Flags().BoolVar(&httpMode, "http", false, "Start MCP server in HTTP mode")
	MCPCommand.Flags().IntVar(&port, "port", 8080, "Port to run HTTP MCP server on")
}
