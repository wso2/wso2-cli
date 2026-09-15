package build

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterBuildTools(mcp *server.MCPServer) {
	mcp.AddTool(GetBuildsTool, getBuilds)
	mcp.AddTool(GetCommitHistoryTool, getCommitHistory)
	mcp.AddTool(CreateBuildTool, createBuild)
}
