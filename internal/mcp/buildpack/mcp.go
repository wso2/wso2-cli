package buildpack

import "github.com/mark3labs/mcp-go/server"

func RegisteBuildpackTools(s *server.MCPServer) {
	s.AddTool(GetBuildPacksTool, getBuildPacks)
}
