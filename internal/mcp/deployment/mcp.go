package deployment

import "github.com/mark3labs/mcp-go/server"

func RegisterDeploymentTools(mcp *server.MCPServer) {
	mcp.AddTool(GetDeploymentTool, getDeploymentDetails)
	mcp.AddTool(CreateDeploymentTool, createDeployment)
}
