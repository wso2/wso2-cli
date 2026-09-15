package execution

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterExecutionTools(s *server.MCPServer) {
	s.AddTool(GetExecutionsTool, getExecutions)
	s.AddTool(ExecuteTaskTool, executeTask)
}
