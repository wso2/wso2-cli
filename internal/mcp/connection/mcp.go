package connection

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/wso2/integration-platform-tools/internal/mcp/connection/database"
	"github.com/wso2/integration-platform-tools/internal/mcp/connection/service"
)

func RegisterConnectionTools(s *server.MCPServer) {
	database.RegisterDatabaseConnectionTools(s)
	service.RegisterConnectionTools(s)
}
