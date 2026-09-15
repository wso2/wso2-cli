package customDomain

import "github.com/mark3labs/mcp-go/server"

func RegisterCustomDomainTools(mcp *server.MCPServer) {
	mcp.AddTool(GetCustomDomainsTool, getCustomDomains)
	mcp.AddTool(RegisterCustomDomainTool, registerCustomDomain)
	mcp.AddTool(CreateUrlMappingTool, createUrlMapping)
}
