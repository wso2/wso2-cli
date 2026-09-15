package auth

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterAuthTools(s *server.MCPServer, httpMode bool) {
	// Login tools are only needed in stdio mode — HTTP mode uses Bearer tokens per request
	if !httpMode {
		s.AddTool(LoginTool, loginToolHandler)
		s.AddTool(CheckLoginStatusTool, checkLoginStatusHandler)
	}
}
