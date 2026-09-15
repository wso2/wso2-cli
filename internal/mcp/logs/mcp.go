package logs

import "github.com/mark3labs/mcp-go/server"

func RegisterLogsTools(mcp *server.MCPServer) {
	mcp.AddTool(GetLogsTool, getLogs)
}
