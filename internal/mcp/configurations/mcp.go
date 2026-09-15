package configurations

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterConfigurationTools(s *server.MCPServer) {
	s.AddTool(CreateConfigurationsTool, createConfigurations)
}
