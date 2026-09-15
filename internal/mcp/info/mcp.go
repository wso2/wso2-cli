package info

import "github.com/mark3labs/mcp-go/server"

func RegisterInfoTools(s *server.MCPServer) {
	s.AddTool(GetStartedTool, getStarted)
}
