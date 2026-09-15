package project

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterProjectTools(s *server.MCPServer) {
	s.AddTool(GetProjectsTool, getProjects)
	s.AddTool(GetProjectEnvironmentsTool, getProjectEnvironments)
	s.AddTool(CreateProjectTool, createProject)
}
