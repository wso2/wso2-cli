package database

import "github.com/mark3labs/mcp-go/server"

func RegisterDatabaseConnectionTools(s *server.MCPServer) {
	s.AddTool(CreateDatabaseConnectionTool, createDatabaseConnection)
}
