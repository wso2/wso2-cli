package testkey

import "github.com/mark3labs/mcp-go/server"

func RegisterOrgTools(s *server.MCPServer) {
	s.AddTool(GenerateTestKeyTool, generateTestKey)
}
