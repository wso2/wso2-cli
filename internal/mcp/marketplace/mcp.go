package marketplace

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMarketplaceTools(s *server.MCPServer) {
	s.AddTool(ListMarketplaceResourcesTool, listMarketplaceResources)
}
