package database

import "github.com/mark3labs/mcp-go/server"

func RegisterDatabaseTools(mcp *server.MCPServer) {
	mcp.AddTool(GetDatabaseServerTool, getDatabaseServer)
	mcp.AddTool(CreateDatabaseServerTool, createDatabaseServer)
	mcp.AddTool(PublishDefaultDatabaseTool, publishDefaultDatabase)
}
