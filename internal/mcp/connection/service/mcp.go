package service

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterConnectionTools(s *server.MCPServer) {
	s.AddTool(CreateConnectionTool, CreateConnection)
}
