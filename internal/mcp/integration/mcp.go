package integration

import "github.com/mark3labs/mcp-go/server"

func RegisterIntegrationTools(s *server.MCPServer) {
	s.AddTool(GetIntegrationsTool, getIntegrations)
	s.AddTool(CreateIntegrationTool, createIntegration)
}
