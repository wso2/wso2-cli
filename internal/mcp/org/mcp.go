package org

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterOrgTools(s *server.MCPServer, httpMode bool) {
	s.AddTool(GetActiveOrgTool, getActiveOrg)
	s.AddTool(GetOrganizationsTool, getOrganizations)
	// In HTTP mode, we don't need to change the organization as it is already set in the token
	if !httpMode {
		s.AddTool(ChangeOrgTool, changeOrg)
	}
}
