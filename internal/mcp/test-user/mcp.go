package testuser

import "github.com/mark3labs/mcp-go/server"

func RegisterTestUserTools(s *server.MCPServer) {
	s.AddTool(ManageTestUsersTool, manageTestUsers)
}
