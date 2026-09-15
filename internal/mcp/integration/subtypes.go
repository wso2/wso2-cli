package integration

import (
	"fmt"
	"strings"

	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

// Integration subtypes exposed by the MCP server. Each maps to an underlying
// Platform component type (and, where the same component type is shared by
// multiple subtypes, a component subtype) so that callers only ever see
// integration vocabulary.
const (
	SubtypeAutomation       = "Automation"
	SubtypeAPI              = "API"
	SubtypeAIAgent          = "AI Agent"
	SubtypeMCPServer        = "MCP Server"
	SubtypeEventIntegration = "Event Integration"
	SubtypeFileIntegration  = "File Integration"
)

// Subtypes is the exhaustive, ordered list of integration subtypes the MCP
// server can create. No other platform component type is reachable through it.
var Subtypes = []string{
	SubtypeAutomation,
	SubtypeAPI,
	SubtypeAIAgent,
	SubtypeMCPServer,
	SubtypeEventIntegration,
	SubtypeFileIntegration,
}

type subtypeMapping struct {
	componentType    string
	componentSubType string
}

var subtypeMappings = map[string]subtypeMapping{
	SubtypeAutomation:       {component.ComponentTypeScheduledTask, ""},
	SubtypeAPI:              {component.ComponentTypeService, ""},
	SubtypeAIAgent:          {component.ComponentTypeService, "aiAgent"},
	SubtypeMCPServer:        {component.ComponentTypeService, "MCP"},
	SubtypeEventIntegration: {component.ComponentTypeEventHandler, ""},
	SubtypeFileIntegration:  {component.ComponentTypeEventHandler, "fileIntegration"},
}

// ResolveSubtype maps an integration subtype to the underlying platform
// component type and component subtype to send on create.
func ResolveSubtype(subtype string) (componentType string, componentSubType string, err error) {
	mapping, ok := subtypeMappings[subtype]
	if !ok {
		return "", "", fmt.Errorf("unsupported integration subtype %q: must be one of %s", subtype, strings.Join(Subtypes, ", "))
	}
	return mapping.componentType, mapping.componentSubType, nil
}
